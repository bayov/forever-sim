package main

// The buffs on our character: the ones up when BayovCore read the sheet, and the ones on us
// during the window.
//
// The sheet has the buffs that were up at export time, and the fight has the buffs that were
// up during the window. Those aren't the same: on Nov 5 another player's Blessing of Might came
// and went in the middle of a window. So the sheet check gives the sim the history record's
// buffs, and the replay gives it the buffs the log shows going on and off us, at those times.
//
// We read each buff's stats from the client's SpellEffect rows for its spell ID. The ID is the
// rank, so we get the rank the caster really used, where the sim's own buffs (sim/core/buffs.go)
// pick the highest rank our level can have. The client rows hold the same values as the sim's
// tables (Mark of the Wild, Power Word: Fortitude, Divine Spirit, Arcane Intellect, Blessing of
// Might, Devotion Aura, Strength of Earth), so this is only a different way to pick the rank.
// buffOverrides holds the few buffs whose client rows don't hold their stats.

import (
	"encoding/json"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// buffEffect is one stat aura of a buff in the client's SpellEffect table.
type buffEffect struct {
	aura, misc int32
	amount     float64
}

// buffAuras are the aura types we read from the client. The misc value says which stat.
var buffAuras = map[int32]bool{
	22:  true, // resistances, misc is a school mask where 1 is armor
	29:  true, // a stat (0 to 4), or all five (-1)
	34:  true, // max health
	36:  true, // a form, like Ghost Wolf
	80:  true, // a stat by a percent (Scroll of Endless Fortitude)
	99:  true, // melee attack power
	137: true, // stats by a percent (Blessing of Kings)
	290: true, // crit chance with everything (Leader of the Pack, the Camp Chair)
	674: true, // armor (misc 1)
}

var primaryStats = [5]stats.Stat{stats.Strength, stats.Agility, stats.Stamina, stats.Intellect, stats.Spirit}

// resistanceBits are aura 22's school mask bits after armor's.
var resistanceBits = map[int32]stats.Stat{
	4:  stats.FireResistance,
	8:  stats.NatureResistance,
	16: stats.FrostResistance,
	32: stats.ShadowResistance,
	64: stats.ArcaneResistance,
}

// buffOverrides are the buffs whose client rows don't hold their stats.
//
// Well Fed's amount comes from the food we ate (the second effect of the Nutritious Food spell
// that gave it), and the Well Fed spell itself says 1. The log never shows the food or the Well
// Fed, so we use the food we eat at these levels. That's the tier that eats with "Food" (434),
// which gives 3 Strength, 3 Intellect, 6 attack power or 4 spell damage. The Nov 8 and 9 sheets
// agree: Well Fed was 3 Strength or 3 Intellect every time.
//
// Veteran Adventurer's Loyal Companion is a pet that follows us, and the sheets with it up are
// the same as without it.
var buffOverrides = map[int32]stats.Stats{
	1248421: {stats.Intellect: 3},
	1248422: {stats.Strength: 3},
	1249519: {stats.AttackPower: 6, stats.RangedAttackPower: 6},
	1249520: {stats.SpellDamage: 4},
	1309405: {},
}

// Camp Benefits says we're at a camp, but not which camp items it holds. Its parts are hidden
// auras that neither the history nor the log shows. The one that matters most is the Camp
// Chair (2% crit with everything). But the chair isn't at every camp: on Nov 8 the sheets with
// Camp Benefits had 2% more crit than the sim, and on Nov 9 most had none. So we count the chair
// only when the export's sheet shows its crit, and then for every Camp Benefits in the window.
const (
	campBenefits  = 1229741
	campChairCrit = 2.0
)

// isForm says the buff is a form like Ghost Wolf. A sheet read in a form says nothing about
// the fight: on Nov 8 at 21:00 Ghost Wolf took 424 attack power off the sheet.
func (t *clientTables) isForm(id int32) bool {
	return slices.ContainsFunc(t.buffEffects[id], func(e buffEffect) bool { return e.aura == 36 })
}

// hasBuffStats says we know the buff's stats.
func (t *clientTables) hasBuffStats(id int32) bool {
	if s, ok := buffOverrides[id]; ok {
		return s != stats.Stats{}
	}
	if id == campBenefits {
		return true
	}
	return slices.ContainsFunc(t.buffEffects[id], func(e buffEffect) bool { return e.aura != 36 })
}

// buffTotal is what the buffs add to base, the sim's stats without buffs. The crit auras (290)
// don't add up: the game keeps the biggest one. A percent goes on the stat with the other buffs
// in it, and the game shows whole stats, so we round it.
func (t *clientTables) buffTotal(ids []int32, base stats.Stats, chair bool) stats.Stats {
	var total stats.Stats
	multiplier := [5]float64{1, 1, 1, 1, 1}
	crit := 0.0
	primary := func(misc int32, fn func(i int)) {
		for i := range primaryStats {
			if misc == -1 || misc == int32(i) {
				fn(i)
			}
		}
	}
	for _, id := range ids {
		if s, ok := buffOverrides[id]; ok {
			total = total.Add(s)
			continue
		}
		if id == campBenefits && chair {
			crit = max(crit, campChairCrit)
		}
		for _, e := range t.buffEffects[id] {
			switch e.aura {
			case 29:
				primary(e.misc, func(i int) { total[primaryStats[i]] += e.amount })
			case 80, 137:
				primary(e.misc, func(i int) { multiplier[i] *= 1 + e.amount/100 })
			case 99:
				total[stats.AttackPower] += e.amount
			case 34:
				total[stats.Health] += e.amount
			case 674:
				if e.misc&1 != 0 {
					total[stats.BonusArmor] += e.amount
				}
			case 22:
				if e.misc&1 != 0 {
					total[stats.BonusArmor] += e.amount
				}
				for bit, stat := range resistanceBits {
					if e.misc&bit != 0 {
						total[stat] += e.amount
					}
				}
			case 290:
				crit = max(crit, e.amount)
			}
		}
	}
	for i, stat := range primaryStats {
		if multiplier[i] != 1 {
			have := base[stat] + total[stat]
			total[stat] += math.Round(have*multiplier[i]) - have
		}
	}
	total[stats.MeleeCrit] += crit * core.CritRatingPerCritChance
	total[stats.SpellCrit] += crit * core.SpellCritRatingPerCritChance
	return total
}

// buffIDs reads an export's buffs as spell IDs. A buff without an ID comes as its name, and we
// skip it.
func buffIDs(raw json.RawMessage) []int32 {
	var list []any
	if json.Unmarshal(raw, &list) != nil {
		return nil
	}
	var ids []int32
	for _, v := range list {
		if n, ok := v.(float64); ok {
			ids = append(ids, int32(n))
		}
	}
	return ids
}

// buffStep is the buffs with stats on us from a time in the window on.
type buffStep struct {
	at  time.Duration // from the start of the window
	ids []int32
}

// windowBuffs follows the buffs with stats that others put on us, from the log's start to the
// window's end. It returns the set at the window's start and the set after each change in it.
//
// Our own buffs and totems are left out, because the replay casts them. The log misses some
// removals: on Nov 8 Blessing of Kings went on at 09:48 and again at 10:06 with no removal in
// between. So a buff also ends when its duration from the last apply runs out. A buff whose
// first line in the log is its removal was up before the log started. A death takes them all off.
//
// Some buffs never show in the log (Well Fed, a totem's aura). Those come from the export: when
// the export has one, we take it as up for the whole window. That's a guess, because the export
// was read at one moment, but it's the only thing that knows about them.
func windowBuffs(all []*event, cl *combatLog, ex *export, tables *clientTables) []buffStep {
	ours := map[string]bool{cl.me: true}
	var ourCasts []string
	logged := map[int32]bool{}
	for _, e := range all {
		if e.kind == "SPELL_SUMMON" && e.srcGUID == cl.me {
			ours[e.dstGUID] = true
		}
		if e.kind == "SPELL_CAST_SUCCESS" && e.srcGUID == cl.me && !slices.Contains(ourCasts, e.spell) {
			ourCasts = append(ourCasts, e.spell)
		}
		if e.dstGUID == cl.me && strings.HasPrefix(e.kind, "SPELL_AURA_") {
			logged[e.spellID] = true
		}
	}

	type span struct{ from, to time.Time }
	spans := map[int32][]span{}
	forever := cl.to.Add(time.Hour)
	for _, id := range buffIDs(ex.Buffs) {
		if logged[id] || !tables.hasBuffStats(id) {
			continue
		}
		// Our own totem's aura (Strength of Earth from Strength of Earth Totem) doesn't show
		// either, and the replay has it already.
		name := tables.spellNames[id]
		if slices.ContainsFunc(ourCasts, func(cast string) bool { return name != "" && strings.HasPrefix(cast, name) }) {
			continue
		}
		spans[id] = []span{{all[0].t, forever}}
	}

	type state struct {
		on, seen    bool
		since, last time.Time
	}
	states := map[int32]*state{}
	ends := func(id int32, st *state) time.Time {
		if d := tables.buffDurations[id]; d > 0 {
			return st.last.Add(d)
		}
		return forever
	}
	stop := func(id int32, st *state, t time.Time) {
		if end := ends(id, st); end.Before(t) {
			t = end
		}
		spans[id] = append(spans[id], span{st.since, t})
		st.on = false
	}
	for _, e := range all {
		if e.t.After(cl.to) {
			break
		}
		if e.kind == "UNIT_DIED" && e.dstGUID == cl.me {
			for id, st := range states {
				if st.on {
					stop(id, st, e.t)
				}
			}
			continue
		}
		if e.dstGUID != cl.me || !strings.HasPrefix(e.kind, "SPELL_AURA_") || ours[e.srcGUID] || !tables.hasBuffStats(e.spellID) {
			continue
		}
		id := e.spellID
		st := states[id]
		if st == nil {
			st = &state{}
			states[id] = st
		}
		switch {
		case e.kind == "SPELL_AURA_REMOVED" && st.on:
			stop(id, st, e.t)
		case e.kind == "SPELL_AURA_REMOVED" && !st.seen:
			from := all[0].t
			if d := tables.buffDurations[id]; d > 0 && e.t.Add(-d).After(from) {
				from = e.t.Add(-d)
			}
			spans[id] = append(spans[id], span{from, e.t})
		case e.kind == "SPELL_AURA_REMOVED":
		default:
			if st.on && ends(id, st).Before(e.t) {
				stop(id, st, e.t)
			}
			if !st.on {
				st.on, st.since = true, e.t
			}
			st.last = e.t
		}
		st.seen = true
	}
	for id, st := range states {
		if st.on {
			stop(id, st, forever)
		}
	}

	// The set changes only where a span starts or ends.
	at := func(t time.Time) []int32 {
		var ids []int32
		for id, list := range spans {
			if slices.ContainsFunc(list, func(s span) bool { return !t.Before(s.from) && t.Before(s.to) }) {
				ids = append(ids, id)
			}
		}
		slices.Sort(ids)
		return ids
	}
	times := []time.Time{cl.from}
	for _, list := range spans {
		for _, s := range list {
			for _, t := range []time.Time{s.from, s.to} {
				if t.After(cl.from) && !t.After(cl.to) {
					times = append(times, t)
				}
			}
		}
	}
	slices.SortFunc(times, time.Time.Compare)
	var steps []buffStep
	for _, t := range times {
		ids := at(t)
		if len(steps) > 0 && slices.Equal(ids, steps[len(steps)-1].ids) {
			continue
		}
		steps = append(steps, buffStep{at: t.Sub(cl.from), ids: ids})
	}
	return steps
}
