package main

import (
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// report prints what logcheck found. Every line that the sim can't explain also goes in
// flagged, which the report repeats at the end.
type report struct {
	cl          *combatLog
	ex          *export
	char        *character
	tables      *clientTables
	logPath     string
	minAttempts int
	flagged     []string
	// missing holds the spells the sim doesn't have, printed once for the whole report.
	missing map[string]bool
	// aliased holds the sim spells we matched to a log spell of another ID, printed once.
	aliased  map[int32]bool
	logNames map[int32]string
	// rounding is how the server turns damage into whole points: "floor" or "round".
	rounding string
}

// toWhole turns a sim hit into the whole points the log would show.
//
// The beta truncates, as far as we can tell (2026-10-13): with the sheet's spell power the
// sim gives Flame Shock ticks 21.7, Flametongue Totem's attack 8.8 and Flame Shock's direct
// hit 64.3, and the log has 21, 8 and 64 on every hit.
func (r *report) toWhole(v float64) float64 {
	if r.rounding == "round" {
		return math.Round(v)
	}
	// A hair over, so 8.9999999 from float sums still counts as 9.
	return math.Floor(v + 1e-6)
}

func (r *report) header() {
	fmt.Printf("Log     %s\n", r.logPath)
	fmt.Printf("Window  %s to %s (%s)\n", r.cl.from.Format(stampFormat), r.cl.to.Format(stampFormat), r.cl.to.Sub(r.cl.from).Round(time.Second))
	gap := r.cl.to.Sub(r.ex.time)
	when := fmt.Sprintf("%s before the window ends", gap.Round(time.Minute))
	if gap < 0 {
		when = fmt.Sprintf("%s after the window ends", (-gap).Round(time.Minute))
	}
	fmt.Printf("Export  %s (%s, %s)\n", filepath.Base(r.ex.path), r.ex.When.Date, when)
	p := r.char.player
	fmt.Printf("Sim     level %d %s %s, talents %s, %s and %s, client tables %s\n",
		p.Level, strings.TrimPrefix(p.Race.String(), "Race"), strings.TrimPrefix(p.Class.String(), "Class"),
		p.TalentsString, p.Profession1, p.Profession2, r.tables.build)
	for _, note := range r.char.notes {
		fmt.Printf("  note: %s\n", note)
	}
	fmt.Println()
}

// sheet compares the export's character sheet with the sim's stats for the same character.
// The sheet has whatever buffs were up when the user exported, and the sim has none.
func (r *report) sheet() {
	result := core.ComputeStats(&proto.ComputeStatsRequest{
		Raid:    core.SinglePlayerRaidProto(r.char.player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Ruleset: proto.Ruleset_RulesetForever,
	})
	if result.ErrorResult != "" {
		fmt.Println("Sheet check failed:", result.ErrorResult)
		return
	}
	s := stats.FromFloatArray(result.RaidStats.Parties[0].Players[0].FinalStats.Stats)

	sheet := func(path ...string) (float64, bool) {
		var value any
		raw, ok := r.ex.Stats[path[0]]
		if !ok || json.Unmarshal(raw, &value) != nil {
			return 0, false
		}
		for _, key := range path[1:] {
			m, ok := value.(map[string]any)
			if !ok {
				return 0, false
			}
			value = m[key]
		}
		n, ok := value.(float64)
		return n, ok
	}
	attackPower := func() (float64, bool) {
		base, ok := sheet("attackPower", "base")
		pos, _ := sheet("attackPower", "pos")
		neg, _ := sheet("attackPower", "neg")
		return base + pos + neg, ok
	}

	fmt.Println("Sheet check (the export's sheet has the buffs that were up, the sim has none)")
	fmt.Printf("  %-14s %8s %8s %8s\n", "", "sheet", "sim", "diff")
	var off []string
	rows := []struct {
		name  string
		sim   float64
		sheet func() (float64, bool)
	}{
		{"Strength", s[stats.Strength], func() (float64, bool) { return sheet("strength", "effective") }},
		{"Agility", s[stats.Agility], func() (float64, bool) { return sheet("agility", "effective") }},
		{"Stamina", s[stats.Stamina], func() (float64, bool) { return sheet("stamina", "effective") }},
		{"Intellect", s[stats.Intellect], func() (float64, bool) { return sheet("intellect", "effective") }},
		{"Spirit", s[stats.Spirit], func() (float64, bool) { return sheet("spirit", "effective") }},
		{"Health", s[stats.Health], func() (float64, bool) { return sheet("health") }},
		{"Mana", s[stats.Mana], func() (float64, bool) { return sheet("mana") }},
		{"Armor", s[stats.Armor], func() (float64, bool) { return sheet("armor", "effective") }},
		{"Attack power", s[stats.AttackPower], attackPower},
		{"Melee crit %", s[stats.MeleeCrit] - 0.04*r.mainHandSkillGap(), func() (float64, bool) { return sheet("crit") }},
		{"Spell crit %", s[stats.SpellCrit], func() (float64, bool) { return sheet("spellCrit", "nature") }},
		{"Dodge %", s[stats.Dodge], func() (float64, bool) { return sheet("dodge") }},
		{"Nature power", s[stats.SpellPower] + s[stats.SpellDamage] + s[stats.NaturePower], func() (float64, bool) { return sheet("spellDamage", "nature") }},
	}
	for _, row := range rows {
		want, ok := row.sheet()
		if !ok {
			continue
		}
		diff := row.sim - want
		mark := ""
		if math.Abs(diff) > 0.05 {
			mark = "  <-"
			off = append(off, fmt.Sprintf("%s %+.2f", row.name, diff))
		}
		fmt.Printf("  %-14s %8.2f %8.2f %+8.2f%s\n", row.name, want, row.sim, diff, mark)
	}
	if len(off) > 0 {
		// Every damage row depends on these stats, so a gap here explains damage flags
		// below before any sim formula does.
		what := fmt.Sprintf("the sim's character differs from the sheet (%s). A buff up at export time explains a sheet that's higher, anything else is a gap in the sim's character that moves the damage rows", strings.Join(off, ", "))
		fmt.Printf("  WARNING: %s.\n", what)
		r.flagged = append(r.flagged, "sheet check: "+what)
	}
	fmt.Println()
}

// mainHandSkillGap is how far the main hand's weapon skill is below the most for our level.
// The client's sheet takes 0.04% crit off for each point, and the sim's crit stat doesn't.
func (r *report) mainHandSkillGap() float64 {
	item, ok := core.ItemsByID[r.char.player.Equipment.Items[proto.ItemSlot_ItemSlotMainHand].Id]
	if !ok {
		return 0
	}
	return skillGaps[weaponSkillName(item)]
}

// logNotes lists what in the log the sim doesn't have: gear swaps or stat buffs, and buffs
// from other players.
func (r *report) logNotes() {
	// Each of our casts carries our max health, attack power, spell power and armor. A change
	// shows a gear swap or a stat buff (tools/combat_log/gear.py).
	var changes []string
	last := ""
	for _, e := range r.cl.events {
		if e.kind != "SPELL_CAST_SUCCESS" || e.srcGUID != r.cl.me || len(e.f) <= spellInfo+infoLevel {
			continue
		}
		f := e.f[spellInfo:]
		now := fmt.Sprintf("max health %s, AP %s, SP %s, armor %s", f[3], f[4], f[5], f[6])
		if last != "" && now != last {
			changes = append(changes, fmt.Sprintf("%s %s (%s)", e.t.Format("15:04:05"), now, e.spell))
		}
		last = now
	}
	if len(changes) > 0 {
		fmt.Printf("Our logged stats changed %d times in the window (gear swaps or stat buffs):\n", len(changes))
		for i, change := range changes {
			if i == 8 {
				fmt.Printf("  and %d more\n", len(changes)-i)
				break
			}
			fmt.Printf("  %s\n", change)
		}
	}

	others := map[string]int{}
	for _, e := range r.cl.events {
		if e.kind == "SPELL_AURA_APPLIED" && e.dstGUID == r.cl.me && !r.cl.ours[e.srcGUID] && !isCreature(e.srcGUID) {
			others[e.spell]++
		}
	}
	if len(others) > 0 {
		var names []string
		for name, n := range others {
			names = append(names, fmt.Sprintf("%s x%d", name, n))
		}
		slices.Sort(names)
		fmt.Printf("Buffs from other players, which the sim doesn't have: %s\n", strings.Join(names, ", "))
	}
	fmt.Println()
}

func (r *report) name(key abilityKey) string {
	if key.spell == 0 {
		return "Melee"
	}
	name := r.logName(key.spell)
	if name == "" {
		name = r.tables.spellNames[key.spell]
	}
	if name == "" {
		name = "?"
	}
	name = fmt.Sprintf("%s (%d)", name, key.spell)
	if key.periodic {
		name += " ticks"
	}
	return name
}

// logName is the spell's name in the log, or "" when the log never has it.
func (r *report) logName(id int32) string {
	if r.logNames == nil {
		r.logNames = map[int32]string{}
		for _, e := range r.cl.events {
			if e.spellID != 0 {
				r.logNames[e.spellID] = e.spell
			}
		}
	}
	return r.logNames[id]
}

// nameAliases pairs the client's name for a sim spell with the name the log gives the same
// thing under another ID.
var nameAliases = map[string]string{
	"Flametongue Totem Proc": "Flametongue Attack",
}

// matchByName files the sim's results under the log's spell when the two use different IDs
// for the same thing.
//
// The sim often keeps one ID for a spell where the client logs another. Windfury Weapon's
// attacks are 8232 in the sim and 439440 in the log, for example. When a spell is only in the
// sim and a spell of the same client name is only in the log, we take them as one.
func (r *report) matchByName(logKeys, simKeys map[abilityKey]*tally) map[abilityKey]*tally {
	logOnly := map[string]abilityKey{}
	for key := range logKeys {
		if _, ok := simKeys[key]; !ok && key.spell != 0 {
			logOnly[fmt.Sprintf("%s/%t", r.logName(key.spell), key.periodic)] = key
		}
	}
	matched := map[abilityKey]*tally{}
	for key, t := range simKeys {
		if _, ok := logKeys[key]; ok || key.spell == 0 {
			matched[key] = t
			continue
		}
		name := r.tables.spellNames[key.spell]
		if alias, ok := nameAliases[name]; ok {
			name = alias
		}
		logKey, ok := logOnly[fmt.Sprintf("%s/%t", name, key.periodic)]
		if !ok {
			matched[key] = t
			continue
		}
		if !r.aliased[key.spell] {
			r.aliased[key.spell] = true
			fmt.Printf("  the sim's %s (%d) is the log's %s\n", r.tables.spellNames[key.spell], key.spell, r.name(logKey))
		}
		if m := matched[logKey]; m != nil {
			merged := &tallies{byLevel: map[int]map[abilityKey]*tally{0: {logKey: m}}, whiteLanded: map[int]int{}}
			merged.merge(&tallies{byLevel: map[int]map[abilityKey]*tally{0: {logKey: t}}, whiteLanded: map[int]int{}})
			matched[logKey] = merged.byLevel[0][logKey]
		} else {
			matched[logKey] = t
		}
	}
	return matched
}

// level prints the comparison for the mobs of one level.
func (r *report) level(p *plan, logT, simT *tallies, rs *replayStats) {
	var inCombat time.Duration
	names := map[string]int{}
	for _, f := range p.fights {
		inCombat += f.end.Sub(f.start)
		names[f.name]++
	}
	var mobs []string
	for name, n := range names {
		mobs = append(mobs, fmt.Sprintf("%s x%d", name, n))
	}
	slices.Sort(mobs)
	fmt.Printf("=== Level %d mobs: %d fights, %s in combat (%s)\n", p.level, len(p.fights), inCombat.Round(time.Second), strings.Join(mobs, ", "))
	runs := float64(simT.iterations)
	fmt.Printf("Replayed %d casts a run in %d runs, %.0f%% of them late (cast after a retry), mob hits on us %.1f a minute\n",
		len(p.casts), simT.iterations, 100*float64(rs.late)/math.Max(1, float64(rs.cast)), p.hitsPerMinute)
	var missing []string
	for name, id := range rs.missing {
		if !r.missing[name] {
			missing = append(missing, fmt.Sprintf("%s (%d)", name, id))
			r.missing[name] = true
		}
	}
	if len(missing) > 0 {
		slices.Sort(missing)
		fmt.Printf("  not replayed, the sim doesn't have: %s\n", strings.Join(missing, ", "))
	}
	for name, n := range rs.failed {
		fmt.Printf("  %s failed to cast %.1f times a run\n", name, float64(n)/runs)
	}

	logKeys := logT.byLevel[p.level]
	simKeys := r.matchByName(logKeys, simT.byLevel[p.level])
	var keys []abilityKey
	for key, t := range logKeys {
		if t.attempts >= r.minAttempts {
			keys = append(keys, key)
		}
	}
	slices.SortFunc(keys, func(a, b abilityKey) int { return logKeys[b].attempts - logKeys[a].attempts })

	// A spell we cast can log its damage under another ID (Fire Nova casts 408342 and hits
	// with 408424), so we go by name.
	casts := map[string]bool{}
	for _, c := range p.casts {
		casts[c.name] = true
	}

	for _, key := range keys {
		lt, st := logKeys[key], simKeys[key]
		if st == nil {
			st = newTally()
		}
		name := r.name(key)
		fmt.Printf("\n%s: log %d, sim %.1f a run\n", name, lt.attempts, float64(st.attempts)/runs)
		if st.attempts == 0 {
			r.flag(p.level, name, "the sim never did it")
			continue
		}
		fmt.Printf("  %-8s %5s %7s %13s %8s %7s\n", "", "log", "", "log 95% CI", "sim", "p")
		for _, outcome := range outcomeOrder {
			k := lt.outcomes[outcome]
			simRate := float64(st.outcomes[outcome]) / float64(st.attempts)
			if k == 0 && simRate < 0.005 {
				continue
			}
			lo, hi := wilson(k, lt.attempts)
			pValue := binomialP(k, lt.attempts, simRate)
			mark := pMark(pValue)
			// A hit is what's left after the other outcomes, so its row only repeats theirs.
			if outcome == "hit" {
				mark = ""
			}
			fmt.Printf("  %-8s %5d %6.1f%% %5.1f-%5.1f%% %7.1f%% %7s%s\n", outcome, k, 100*float64(k)/float64(lt.attempts), 100*lo, 100*hi, 100*simRate, pString(pValue), mark)
			if mark == " !!" {
				r.flag(p.level, name, fmt.Sprintf("%s %d of %d (%.1f%%), the sim gives %.1f%%", outcome, k, lt.attempts, 100*float64(k)/float64(lt.attempts), 100*simRate))
			}
		}

		if ls, exact := spreadOf(lt.raw), spreadOf(st.raw); ls.n > 0 && exact.n > 0 {
			// The log has whole points, so we turn the sim's hits into whole points the way
			// the server does (-rounding) and compare those.
			whole := make([]float64, len(st.raw))
			for i, v := range st.raw {
				whole[i] = r.toWhole(v)
			}
			ss := spreadOf(whole)
			outside := 0
			for _, v := range lt.raw {
				if v < ss.min || v > ss.max {
					outside++
				}
			}
			// The averages tell more than the range.
			pValue := 1.0
			if gap, se := math.Abs(ls.mean()-ss.mean()), math.Sqrt(ss.variance()/float64(ls.n)); se > 0 {
				pValue = 2 * (1 - normalCDF(gap/se))
			} else if gap > 1e-9 {
				pValue = 0
			}
			mark := pMark(pValue)
			if outside > 0 {
				mark = " !!"
			}
			fmt.Printf("  raw      log %.0f-%.0f (avg %.1f)   sim %.0f-%.0f (avg %.1f, %.1f-%.1f before %s)   %d outside, avg p %s%s\n",
				ls.min, ls.max, ls.mean(), ss.min, ss.max, ss.mean(), exact.min, exact.max, r.rounding, outside, pString(pValue), mark)
			if mark == " !!" {
				r.flag(p.level, name, fmt.Sprintf("raw damage avg %.1f in %d hits (%.0f-%.0f), the sim's is %.1f (%.0f-%.0f, %.1f-%.1f before %s)",
					ls.mean(), ls.n, ls.min, ls.max, ss.mean(), ss.min, ss.max, exact.min, exact.max, r.rounding))
			}
		}

		if lt.spellLanded > 0 && st.spellLanded > 0 {
			lp, sp := 0, 0
			for _, n := range lt.partial {
				lp += n
			}
			for _, n := range st.partial {
				sp += n
			}
			simRate := float64(sp) / float64(st.spellLanded)
			pValue := binomialP(lp, lt.spellLanded, simRate)
			mark := pMark(pValue)
			fmt.Printf("  partial  log %d of %d (%.1f%%) %s   sim %.1f%% %s   p %s%s\n", lp, lt.spellLanded, 100*float64(lp)/float64(lt.spellLanded),
				shares(lt.partial, 1), 100*simRate, shares(st.partial, float64(st.spellLanded)/100), pString(pValue), mark)
			if mark == " !!" {
				r.flag(p.level, name, fmt.Sprintf("partial resists on %d of %d landed, the sim gives %.1f%%", lp, lt.spellLanded, 100*simRate))
			}
		}
	}

	// What the sim did that the log never had. Often the same thing under another spell ID.
	var simOnly []string
	for key, st := range simKeys {
		if _, ok := logKeys[key]; !ok && float64(st.attempts)/runs >= 0.5 {
			simOnly = append(simOnly, fmt.Sprintf("%s %.1f a run", r.name(key), float64(st.attempts)/runs))
		}
	}
	if len(simOnly) > 0 {
		slices.Sort(simOnly)
		fmt.Printf("\nOnly in the sim: %s\n", strings.Join(simOnly, ", "))
	}

	// Procs and other damage we didn't cast, per landed white swing.
	logWhite, simWhite := logT.whiteLanded[p.level], simT.whiteLanded[p.level]
	if logWhite == 0 || simWhite == 0 {
		fmt.Println()
		return
	}
	fmt.Printf("\nPer 100 landed white swings (log %d, sim %.1f a run), for what we didn't cast:\n", logWhite, float64(simWhite)/runs)
	for _, key := range keys {
		if key.spell == 0 || casts[r.logName(key.spell)] {
			continue
		}
		lt, st := logKeys[key], simKeys[key]
		simAttempts := 0
		if st != nil {
			simAttempts = st.attempts
		}
		rate := float64(simAttempts) / float64(simWhite)
		pValue := poissonP(lt.attempts, rate*float64(logWhite))
		mark := pMark(pValue)
		name := r.name(key)
		fmt.Printf("  %-36s log %5.1f (%d)   sim %5.1f   p %s%s\n", name, 100*float64(lt.attempts)/float64(logWhite), lt.attempts, 100*rate, pString(pValue), mark)
		if mark == " !!" {
			r.flag(p.level, name, fmt.Sprintf("%.1f per 100 landed white swings, the sim gives %.1f", 100*float64(lt.attempts)/float64(logWhite), 100*rate))
		}
	}
	fmt.Println()
}

// shares lists the partial resist steps, as counts (scale 1) or percents of the landed hits.
func shares(partial map[int]int, scale float64) string {
	var steps []int
	for step := range partial {
		steps = append(steps, step)
	}
	slices.Sort(steps)
	var parts []string
	for _, step := range steps {
		if scale == 1 {
			parts = append(parts, fmt.Sprintf("%d%% x%d", step, partial[step]))
		} else {
			parts = append(parts, fmt.Sprintf("%d%% %.1f%%", step, float64(partial[step])/scale))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func pString(p float64) string {
	switch {
	case p == 0:
		return "never"
	case p < 0.001:
		return "<0.001"
	}
	return fmt.Sprintf("%.3f", p)
}

// pMark flags a row whose chance under the sim is below 1 in 1000, since a report prints a
// hundred rows or so, and hints at 1 in 100.
func pMark(p float64) string {
	switch {
	case p < 0.001:
		return " !!"
	case p < 0.01:
		return " !"
	}
	return ""
}

func (r *report) flag(level int, name, what string) {
	r.flagged = append(r.flagged, fmt.Sprintf("level %d, %s: %s", level, name, what))
}

func (r *report) flags() {
	if len(r.flagged) == 0 {
		fmt.Println("Nothing flagged: every row has at least a 1 in 1000 chance under the sim.")
		return
	}
	fmt.Println("Flagged (less than a 1 in 1000 chance under the sim, or impossible in it):")
	for _, f := range r.flagged {
		fmt.Println("  " + f)
	}
}
