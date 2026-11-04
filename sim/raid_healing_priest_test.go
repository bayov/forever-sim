package sim

// The Holy and Discipline priest's heals and healing talents (docs/healing_priest.md), and the absorb
// pipeline (sim/core/absorb.go), on small healing fights like raid_healing_test.go's: a Protection
// warrior the boss hits, the priest, and Frost mages. The priest wears nothing and gets its healing
// power as bonus stats, so a heal's size is known. Needs the item database: go test -tags with_db.

import (
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/priest"
	"github.com/wowsims/classic/sim/priest/healing"
)

// The talents in the client tree's order (proto/priest.proto).
var priestTalentOrder = [3][]string{
	{"power_in_light", "wand_specialization", "twin_disciplines", "silent_resolve", "holy_precision", "mental_strength", "martyrdom",
		"mental_agility", "inner_focus", "improved_power_word_shield", "meditation", "improved_inner_fire", "improved_mana_burn",
		"soul_warding", "penance", "renewed_hope", "divine_aegis", "power_infusion"},
	{"twilight_focus", "improved_renew", "holy_specialization", "spell_warding", "divine_fury", "holy_nova", "blessed_recovery",
		"inspiration", "holy_reach", "improved_healing", "searing_light", "holy_binding", "litany_of_light", "spirit_of_redemption",
		"spiritual_guidance", "spiritual_healing", "prayer_of_mending"},
	{},
}

// priestTalents is a talent string with these talents at these ranks and nothing else.
func priestTalents(given map[string]int) string {
	ranks := map[string]int{}
	for name, rank := range given {
		ranks[name] = rank
	}
	var trees []string
	for _, tree := range priestTalentOrder {
		s := ""
		for _, name := range tree {
			s += strconv.Itoa(ranks[name])
			delete(ranks, name)
		}
		trees = append(trees, strings.TrimRight(s, "0"))
	}
	if len(ranks) > 0 {
		panic("unknown priest talents")
	}
	return strings.TrimRight(strings.Join(trees, "-"), "-")
}

const testHealingPower = 800

// testPriest is a level 60 Human priest with nothing on, 800 healing power, 20,000 more mana and the
// talents given.
func testPriest(talents string) *proto.Player {
	bonus := make([]float64, stats.UnitStatsLen)
	bonus[stats.HealingPower] = testHealingPower
	bonus[stats.Intellect] = 300
	bonus[stats.Mana] = 20000
	return &proto.Player{
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassPriest,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Spec:          &proto.Player_HealingPriest{HealingPriest: &proto.HealingPriest{Options: &proto.HealingPriest_Options{}}},
		BonusStats:    &proto.UnitStats{Stats: bonus},
		Rotation:      &proto.APLRotation{},
	}
}

// priestRaid is the tank (player 0), the priest (1) and mages, in a 60 sec healing fight where the
// boss hits the tank for 600 every 2 sec.
func priestRaid(t *testing.T, talents string, mages int, events ...*proto.RaidDamageEvent) *proto.RaidSimRequest {
	rsr := healingRaid(t, 1, 0, mages, events...)
	players := rsr.Raid.Parties[0].Players
	priestPlayer := testPriest(talents)
	priestPlayer.Name = "priest"
	rsr = raidOf(t, "mage__frost", append([]*proto.Player{players[0], priestPlayer}, players[1:]...), 20)
	rsr.Raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	rsr.Encounter.TrackHealth = true
	rsr.Encounter.RaidDamageHitsPerMinute = 0
	rsr.Encounter.DurationVariation = 0
	rsr.Encounter.Duration = 60
	rsr.Encounter.RaidDamage = events
	rsr.Encounter.Targets[0].TankIndex = 0
	rsr.Encounter.Targets[0].SwingSpeed = 2
	rsr.Encounter.Targets[0].MinBaseDamage = 600
	return rsr
}

// castLine is a priority list line casting the spell at the target (Tank, Self, MostHurt...), with
// an optional condition.
func castLine(spellID int32, target string, condition string) string {
	line := `{"action":{"castSpell":{"spellId":{"spellId":` + strconv.Itoa(int(spellID)) + `}`
	if target != "" {
		line += `,"target":{"type":"` + target + `"}`
	}
	line += `}`
	if condition != "" {
		line += `,"condition":` + condition
	}
	return line + `}}`
}

// newPriestSim builds the request's sim, for the priest's spells and stats.
func newPriestSim(t *testing.T, rsr *proto.RaidSimRequest) (*core.Simulation, *healing.HealingPriest) {
	t.Helper()
	sim := core.NewSim(rsr, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[1].(*healing.HealingPriest)
}

func top(spells []*core.Spell) *core.Spell {
	for i := len(spells) - 1; i >= 0; i-- {
		if spells[i] != nil {
			return spells[i]
		}
	}
	return nil
}

// spellMetrics is a spell's metrics summed over its targets, or on one target (unit >= 0).
type spellMetrics struct {
	casts, hits, crits, ticks, critTicks int32
	healing, critHealing, overhealing    float64
}

func actionMetrics(m *proto.UnitMetrics, spellID int32, unit int32) spellMetrics {
	var s spellMetrics
	for _, a := range m.Actions {
		if a.Id.GetSpellId() != spellID {
			continue
		}
		for _, tgt := range a.Targets {
			if unit >= 0 && tgt.UnitIndex != unit {
				continue
			}
			s.casts += tgt.Casts
			s.hits += tgt.Hits
			s.crits += tgt.Crits
			s.ticks += tgt.Ticks
			s.critTicks += tgt.CritTicks
			s.healing += tgt.Healing
			s.critHealing += tgt.CritHealing
			s.overhealing += tgt.Overhealing
		}
	}
	return s
}

func near(got, want, tolerance float64) bool {
	return math.Abs(got-want) <= tolerance*math.Abs(want)
}

// The client's values at 60 (healing.go), for each heal's top rank and a downranked one: average
// base heal, coefficient after Forever's downranking, mana, cast time.
func TestHealingPriestRanks(t *testing.T) {
	sim, hp := newPriestSim(t, priestRaid(t, priestTalents(map[string]int{"holy_nova": 1, "prayer_of_mending": 1, "penance": 1}), 1))
	tank := sim.Raid.Parties[0].Players[0].GetCharacter()
	checks := []struct {
		name        string
		spell       *core.Spell
		avg, coef   float64
		cost        float64
		castSeconds float64
	}{
		{"Lesser Heal 3", hp.LesserHeal[3], 163 + 1.2*5, 0, 75, 2.5},
		{"Heal 4", hp.Heal[4], 651 + 4.5*5, 0.857 * 0.55, 305, 3},
		{"Greater Heal 1", hp.GreaterHeal[1], 844 + 5.1*5, 0.857 * 0.85, 370, 3},
		{"Greater Heal 5", hp.GreaterHeal[5], 1960, 0.857, 710, 3},
		{"Flash Heal 7", hp.FlashHeal[7], 825 + 4.2*4, 0.429, 380, 1.5},
		{"Prayer of Healing 5", hp.PrayerOfHealing[5], 649 + 1.6*9, 0.286, 1070, 3},
		{"Holy Nova 6", hp.HolyNova[6], 311, 0.107, 750, 0},
	}
	for _, c := range checks {
		if c.spell == nil {
			t.Errorf("%s isn't registered", c.name)
			continue
		}
		if got := c.spell.Cost.GetCurrentCost(); !near(got, c.cost, 1e-9) {
			t.Errorf("%s costs %.1f, want %.1f", c.name, got, c.cost)
		}
		if got := c.spell.DefaultCast.CastTime.Seconds(); !near(got, c.castSeconds, 1e-9) {
			t.Errorf("%s casts in %.2f s, want %.2f", c.name, got, c.castSeconds)
		}
		if got := c.spell.BonusCoefficient; math.Abs(got-c.coef) > 1e-6 {
			t.Errorf("%s has a %.4f coefficient, want %.4f", c.name, got, c.coef)
		}
		want := c.avg + c.coef*c.spell.HealingPower(&tank.Unit)
		if got := c.spell.ExpectedHealing(&tank.Unit, c.avg); !near(got, want, 1e-6) {
			t.Errorf("%s heals %.1f on average, want %.1f", c.name, got, want)
		}
	}
	if got := top(hp.Renew).Cost.GetCurrentCost(); got != 410 {
		t.Errorf("Renew 10 costs %.0f, want 410", got)
	}
	if got := top(hp.PowerWordShield).Cost.GetCurrentCost(); got != 500 {
		t.Errorf("Power Word: Shield 10 costs %.0f, want 500", got)
	}
	if got := top(hp.PowerWordShield).CD.Duration; got != 4*time.Second {
		t.Errorf("Power Word: Shield's cooldown is %s, want 4s", got)
	}
	if got := top(hp.PrayerOfMending); got == nil || got.Cost.GetCurrentCost() != 390 || got.CD.Duration != 10*time.Second {
		t.Errorf("Prayer of Mending 3: %v", got)
	}
	if got := top(hp.Penance); got == nil || got.Cost.GetCurrentCost() != 385 || got.CD.Duration != 12*time.Second {
		t.Errorf("Penance 4: %v", got)
	}
	// Binding Heal is gone since the November 2 notes, every rank.
	for _, id := range []int32{401937, 1240770, 1240771, 1240772, 1240773, 1240774} {
		if hp.GetSpell(core.ActionID{SpellID: id}) != nil {
			t.Errorf("Binding Heal %d is registered", id)
		}
	}
	// Forever's downranking (sim/core/downrank.go) on the sub-spells that carry their rank's learn
	// level, and on the shield: a rank's coefficient keeps 1 - 5% a level past 17.
	for _, c := range []struct {
		name  string
		spell *core.Spell
		coef  float64
		learn int
	}{
		{"Penance 1's heal bolt", hp.GetSpell(core.ActionID{SpellID: priest.PenanceHealSpellId[1]}), 0.19, 30},
		{"Prayer of Mending 1's heal", hp.GetSpell(core.ActionID{SpellID: priest.PrayerOfMendingHealID, Tag: 1}), 0.429, 40},
		{"Power Word: Shield 5", hp.PowerWordShield[5], priest.PowerWordShieldRanks[5].Coefficient(), int(priest.PowerWordShieldRanks[5].LearnLevel())},
		{"Renew 5", hp.Renew[5], priest.RenewRanks[5].Coefficient(), int(priest.RenewRanks[5].LearnLevel())},
	} {
		if c.spell == nil {
			t.Errorf("%s isn't registered", c.name)
			continue
		}
		got := c.spell.BonusCoefficient
		if c.name == "Renew 5" {
			got = c.spell.Hot(&tank.Unit).BonusCoefficient
		}
		want := c.coef * core.DownrankFactor(60, c.learn)
		if want >= c.coef || !near(got, want, 1e-6) {
			t.Errorf("%s has a %.4f coefficient, want %.4f (%.3f downranked)", c.name, got, want, c.coef)
		}
	}
}

// fightPriest runs the raid with the priest casting the lines, and gives the priest's metrics and
// the tank's and the priest's unit indexes.
func fightPriest(t *testing.T, rsr *proto.RaidSimRequest, lines ...string) (res *proto.RaidSimResult, m *proto.UnitMetrics, tank, self int32) {
	t.Helper()
	healerRotation(rsr, 1, lines...)
	res = runRaid(t, rsr)
	return res, playerMetrics(res, 1), playerMetrics(res, 0).UnitIndex, playerMetrics(res, 1).UnitIndex
}

// Each single target heal lands on its target, its hits average what Expected Healing says and its
// crits half again as much, and it costs its mana a cast.
func TestHealingPriestDirectHeals(t *testing.T) {
	talents := priestTalents(map[string]int{})
	for _, c := range []struct {
		name string
		pick func(*healing.HealingPriest) *core.Spell
		avg  float64
	}{
		{"Greater Heal 5", func(hp *healing.HealingPriest) *core.Spell { return hp.GreaterHeal[5] }, 1960},
		{"Flash Heal 7", func(hp *healing.HealingPriest) *core.Spell { return hp.FlashHeal[7] }, 825 + 4.2*4},
		{"Heal 4", func(hp *healing.HealingPriest) *core.Spell { return hp.Heal[4] }, 651 + 4.5*5},
	} {
		rsr := priestRaid(t, talents, 0)
		rsr.SimOptions.Iterations = 40
		sim, hp := newPriestSim(t, rsr)
		spell := c.pick(hp)
		tankUnit := &sim.Raid.Parties[0].Players[0].GetCharacter().Unit
		want := spell.ExpectedHealing(tankUnit, c.avg)
		manaCost := spell.Cost.GetCurrentCost()
		_, m, tank, _ := fightPriest(t, rsr, castLine(spell.SpellID, "Tank", ""))
		s := actionMetrics(m, spell.SpellID, tank)
		if s.casts == 0 || s.hits+s.crits != s.casts {
			t.Errorf("%s: %d casts, %d landed on the tank", c.name, s.casts, s.hits+s.crits)
			continue
		}
		if got := (s.healing - s.critHealing) / float64(s.hits); !near(got, want, 0.02) {
			t.Errorf("%s heals %.0f a hit, want %.0f", c.name, got, want)
		}
		if s.crits > 0 {
			if got := s.critHealing / float64(s.crits); !near(got, 1.5*want, 0.05) {
				t.Errorf("%s crits for %.0f, want %.0f", c.name, got, 1.5*want)
			}
		}
		if got, want := m.ManaSpentAvg, float64(s.casts)/40*manaCost; !near(got, want, 0.01) {
			t.Errorf("%s: %.0f mana spent a fight, want %.0f", c.name, got, want)
		}
	}
}

// Prayer of Healing heals its target's party, Holy Nova the priest's party, each for the rank's
// heal.
func TestHealingPriestGroupHeals(t *testing.T) {
	talents := priestTalents(map[string]int{"holy_nova": 1})
	for _, c := range []struct {
		name        string
		spellID     int32
		target      string
		wantTargets int
	}{
		{"Prayer of Healing", priest.PrayerOfHealingRanks[5].SpellID(), "Tank", 5},
		{"Holy Nova", priest.HolyNovaRanks[6].SpellID(), "", 5},
	} {
		rsr := priestRaid(t, talents, 3)
		rsr.SimOptions.Iterations = 5
		_, m, _, _ := fightPriest(t, rsr, castLine(c.spellID, c.target, ""))
		targets := 0
		var casts, landed int32
		for _, a := range m.Actions {
			if a.Id.GetSpellId() != c.spellID {
				continue
			}
			for _, tgt := range a.Targets {
				casts += tgt.Casts
				if tgt.Hits+tgt.Crits > 0 {
					targets++
					landed += tgt.Hits + tgt.Crits
				}
			}
		}
		if targets != c.wantTargets || casts == 0 || landed != casts*int32(c.wantTargets) {
			t.Errorf("%s: %d casts landed %d heals on %d players, want each on %d", c.name, casts, landed, targets, c.wantTargets)
		}
	}
}

// Renew ticks 5 times over 15 sec, a tick its base plus 20% of healing power, and can crit.
func TestHealingPriestRenew(t *testing.T) {
	rsr := priestRaid(t, "", 0)
	rsr.SimOptions.Iterations = 20
	sim, hp := newPriestSim(t, rsr)
	renew := top(hp.Renew)
	tankUnit := &sim.Raid.Parties[0].Players[0].GetCharacter().Unit
	tick := renew.ExpectedHealing(tankUnit, 166+0.2*renew.HealingPower(tankUnit))
	// One Renew at the start of each fight only.
	_, m, tank, _ := fightPriest(t, rsr, castLine(renew.SpellID, "Tank", `{"cmp":{"op":"OpLt","lhs":{"currentTime":{}},"rhs":{"const":{"val":"1s"}}}}`))
	s := actionMetrics(m, renew.SpellID, tank)
	if s.casts != 20 || s.ticks+s.critTicks != 5*s.casts {
		t.Fatalf("%d Renews ticked %d times, want 5 each", s.casts, s.ticks+s.critTicks)
	}
	if got := (s.healing - s.critHealing) / float64(s.ticks); !near(got, tick, 1e-6) {
		t.Errorf("a Renew tick heals %.1f, want %.1f", got, tick)
	}
	if got := m.ManaSpentAvg; !near(got, 410, 1e-9) {
		t.Errorf("Renew cost %.0f, want 410", got)
	}
}

// Power Word: Shield absorbs the boss's hits on the tank before its health takes them, and its
// absorbing counts as healing; Weakened Soul keeps another off for 15 sec.
func TestHealingPriestPowerWordShield(t *testing.T) {
	request := func() *proto.RaidSimRequest {
		rsr := priestRaid(t, "", 0)
		rsr.SimOptions.Iterations = 20
		rsr.Encounter.Duration = 44
		rsr.Encounter.Targets[0].MinBaseDamage = 800
		rsr.Encounter.Targets[0].DamageSpread = 0
		return rsr
	}
	rsr := request()
	sim, hp := newPriestSim(t, rsr)
	shield := top(hp.PowerWordShield)
	tankUnit := &sim.Raid.Parties[0].Players[0].GetCharacter().Unit
	absorb := 928 + 0.1*shield.HealingPower(tankUnit)

	without := runRaid(t, request())
	res, m, tank, _ := fightPriest(t, rsr, castLine(shield.SpellID, "Tank", ""))
	s := actionMetrics(m, shield.SpellID, tank)
	// 44 sec: a shield at 0, 15 and 30.
	if s.casts != 3*20 {
		t.Errorf("%d shields in 20 fights, want 3 each (Weakened Soul is 15 sec)", s.casts)
	}
	// A shield books its whole size: what it absorbed, and the rest when the next one replaced it.
	// The last one's rest at the fight's end isn't booked.
	if got := s.healing / float64(s.casts); got > absorb*(1+1e-9) || got < 0.9*absorb {
		t.Errorf("a shield books %.1f, want up to %.1f", got, absorb)
	}
	// What they absorbed is the effective part, and exactly what the tank's health didn't take.
	taken, takenWithout := playerMetrics(res, 0).Dtps.Avg, playerMetrics(without, 0).Dtps.Avg
	absorbed := (s.healing - s.overhealing) / 20 / 44
	if want := takenWithout - absorbed; !near(taken, want, 1e-3) {
		t.Errorf("the tank took %.2f DTPS shielded, want %.2f (%.2f unshielded, %.2f absorbed)", taken, want, takenWithout, absorbed)
	}
	if died := playerMetrics(without, 0).ChanceOfDeath; died > 0 {
		t.Errorf("the tank died in %.0f%% of fights unshielded", 100*died)
	}
	if !near(m.Ehps.Avg, absorbed, 1e-6) || m.Hps.Avg < m.Ehps.Avg {
		t.Errorf("shield HPS %.1f, EHPS %.1f: want EHPS what they absorbed, %.1f", m.Hps.Avg, m.Ehps.Avg, absorbed)
	}
}

// A shield on a player nothing hits absorbs nothing: when it runs out its whole size is booked as
// healing and overhealing (the default for absorbs, docs/healing_sim.md).
func TestHealingPriestShieldUnused(t *testing.T) {
	rsr := priestRaid(t, "", 1)
	rsr.Encounter.Duration = 100
	rsr.SimOptions.Iterations = 5
	_, hp := newPriestSim(t, rsr)
	shield := top(hp.PowerWordShield)
	absorb := 928 + 0.1*shield.HealingPower(&hp.Unit)
	_, m, _, _ := fightPriest(t, rsr, castLine(shield.SpellID, "", `{"cmp":{"op":"OpLt","lhs":{"currentTime":{}},"rhs":{"const":{"val":"1s"}}}}`))
	// Cast at the priest (the default target is the boss, which lands on us).
	s := actionMetrics(m, shield.SpellID, -1)
	if s.casts != 5 || !near(s.healing, 5*absorb, 1e-6) || !near(s.overhealing, s.healing, 1e-9) {
		t.Errorf("%d unused shields booked %.0f healing, %.0f overhealing, want %.0f of each", s.casts, s.healing, s.overhealing, 5*absorb)
	}
	if m.Ehps.Avg != 0 {
		t.Errorf("unused shields gave %.2f EHPS", m.Ehps.Avg)
	}
}

// A shield still up can be replaced by an equal or higher rank only (the November 2 notes). Rank
// 10 at the start, then rank 9 whenever it can be cast: Weakened Soul lifts at 15 s, but rank 10's
// shield holds rank 9 back until it runs out at 30 s; rank 9 then replaces itself at 45 s (a 55 s
// fight, so no cast lands on its last moment).
func TestHealingPriestShieldReplaceRank(t *testing.T) {
	rsr := priestRaid(t, "", 1)
	rsr.SimOptions.Iterations = 1
	rsr.Encounter.Duration = 55
	_, hp := newPriestSim(t, rsr)
	high, low := hp.PowerWordShield[10], hp.PowerWordShield[9]
	_, m, _, _ := fightPriest(t, rsr,
		castLine(high.SpellID, "", `{"cmp":{"op":"OpLt","lhs":{"currentTime":{}},"rhs":{"const":{"val":"1s"}}}}`),
		castLine(low.SpellID, "", ""))
	if h, l := actionMetrics(m, high.SpellID, -1).casts, actionMetrics(m, low.SpellID, -1).casts; h != 1 || l != 2 {
		t.Errorf("rank 10 cast %d times and rank 9 %d, want 1 and 2 (at 30 s and 45 s)", h, l)
	}
}

// Prayer of Mending heals 5 times from one cast while the raid takes damage, each heal its base
// plus 42.9% of healing power, jumping from player to player.
func TestHealingPriestPrayerOfMending(t *testing.T) {
	pulse := &proto.RaidDamageEvent{School: proto.SpellSchool_SpellSchoolFire, MinDamage: 300, MaxDamage: 300, Who: proto.RaidDamageEvent_Everyone, EverySeconds: 2, FirstAt: 1}
	rsr := priestRaid(t, priestTalents(map[string]int{"prayer_of_mending": 1}), 3, pulse)
	rsr.SimOptions.Iterations = 10
	sim, hp := newPriestSim(t, rsr)
	pom := top(hp.PrayerOfMending)
	tankUnit := &sim.Raid.Parties[0].Players[0].GetCharacter().Unit
	heal := hp.GetSpell(core.ActionID{SpellID: priest.PrayerOfMendingHealID, Tag: 3})
	want := heal.ExpectedHealing(tankUnit, 413)
	_, m, _, _ := fightPriest(t, rsr, castLine(pom.SpellID, "Tank", ""))
	cast := actionMetrics(m, pom.SpellID, -1)
	heals := actionMetrics(m, priest.PrayerOfMendingHealID, -1)
	targets := 0
	for _, a := range m.Actions {
		if a.Id.GetSpellId() == priest.PrayerOfMendingHealID {
			targets = len(a.Targets)
		}
	}
	// Each cast heals 5 times, but the last cast of a fight may not finish.
	if n := heals.hits + heals.crits; cast.casts == 0 || n > 5*cast.casts || n < 5*(cast.casts-10) {
		t.Errorf("%d casts healed %d times, want 5 each", cast.casts, n)
	}
	if targets < 3 {
		t.Errorf("Prayer of Mending healed %d players, want it to jump around the raid", targets)
	}
	if got := (heals.healing - heals.critHealing) / float64(heals.hits); !near(got, want, 1e-6) {
		t.Errorf("a Prayer of Mending heal is %.1f, want %.1f", got, want)
	}
}

// Penance at a player heals with 3 bolts over 2 sec, each its base plus 19% of healing power, on a
// 12 sec cooldown.
func TestHealingPriestPenanceHeal(t *testing.T) {
	rsr := priestRaid(t, priestTalents(map[string]int{"penance": 1}), 0)
	rsr.SimOptions.Iterations = 10
	rsr.Encounter.Duration = 59
	sim, hp := newPriestSim(t, rsr)
	penance := top(hp.Penance)
	tankUnit := &sim.Raid.Parties[0].Players[0].GetCharacter().Unit
	bolt := hp.GetSpell(core.ActionID{SpellID: priest.PenanceHealSpellId[4]})
	want := bolt.ExpectedHealing(tankUnit, 425)
	_, m, tank, _ := fightPriest(t, rsr, castLine(penance.SpellID, "Tank", ""))
	cast := actionMetrics(m, penance.SpellID, -1)
	bolts := actionMetrics(m, priest.PenanceHealSpellId[4], tank)
	// 59 sec at a 12 sec cooldown: casts at 0, 12, 24, 36 and 48.
	if cast.casts != 5*10 || bolts.hits+bolts.crits != 3*cast.casts {
		t.Errorf("%d Penances fired %d heal bolts at the tank, want 5 a fight and 3 each", cast.casts, bolts.hits+bolts.crits)
	}
	if got := (bolts.healing - bolts.critHealing) / float64(bolts.hits); !near(got, want, 1e-6) {
		t.Errorf("a Penance bolt heals %.1f, want %.1f", got, want)
	}
	if damage := actionMetrics(m, priest.PenanceBoltSpellId[4], -1); damage.casts > 0 {
		t.Errorf("Penance at the tank fired %d damage bolts", damage.casts)
	}
}

// talentSim builds the sim for a priest with these talents, for the talents' static effects.
func talentSim(t *testing.T, ranks map[string]int) (*core.Simulation, *healing.HealingPriest, *core.Unit) {
	ranks["holy_nova"], ranks["prayer_of_mending"], ranks["penance"] = 1, 1, 1
	sim, hp := newPriestSim(t, priestRaid(t, priestTalents(ranks), 0))
	return sim, hp, &sim.Raid.Parties[0].Players[0].GetCharacter().Unit
}

// Each healing talent moves exactly what it should against the same priest without it.
func TestHealingPriestTalentValues(t *testing.T) {
	_, base, tank := talentSim(t, map[string]int{})
	heal := func(hp *healing.HealingPriest, spell func(*healing.HealingPriest) *core.Spell) float64 {
		return spell(hp).ExpectedHealing(tank, 1000)
	}
	gh := func(hp *healing.HealingPriest) *core.Spell { return hp.GreaterHeal[5] }
	fh := func(hp *healing.HealingPriest) *core.Spell { return hp.FlashHeal[7] }
	renew := func(hp *healing.HealingPriest) *core.Spell { return top(hp.Renew) }
	nova := func(hp *healing.HealingPriest) *core.Spell { return top(hp.HolyNova) }
	pom := func(hp *healing.HealingPriest) *core.Spell { return top(hp.PrayerOfMending) }
	pws := func(hp *healing.HealingPriest) *core.Spell { return top(hp.PowerWordShield) }
	absorb := func(hp *healing.HealingPriest) float64 { return hp.PowerWordShieldAbsorb(10, tank) }
	penance := func(hp *healing.HealingPriest) *core.Spell { return top(hp.Penance) }
	penanceBolt := func(hp *healing.HealingPriest) *core.Spell {
		return hp.GetSpell(core.ActionID{SpellID: priest.PenanceHealSpellId[4]})
	}
	cost := func(hp *healing.HealingPriest, spell func(*healing.HealingPriest) *core.Spell) float64 {
		return spell(hp).Cost.GetCurrentCost()
	}
	crit := func(hp *healing.HealingPriest, spell func(*healing.HealingPriest) *core.Spell) float64 {
		return spell(hp).BonusCritRating
	}

	type check struct {
		what      string
		got, want float64
	}
	for _, c := range []struct {
		talent string
		rank   int
		checks func(hp *healing.HealingPriest) []check
	}{
		{"spiritual_healing", 3, func(hp *healing.HealingPriest) []check {
			return []check{
				{"Greater Heal", heal(hp, gh), 1.10 * heal(base, gh)},
				{"Renew", heal(hp, renew), 1.10 * heal(base, renew)},
				{"Penance's heal", heal(hp, penanceBolt), 1.10 * heal(base, penanceBolt)},
				{"Power Word: Shield", absorb(hp), absorb(base)},
			}
		}},
		{"improved_renew", 3, func(hp *healing.HealingPriest) []check {
			return []check{{"Renew", heal(hp, renew), 1.15 * heal(base, renew)}, {"Flash Heal", heal(hp, fh), heal(base, fh)}}
		}},
		{"twin_disciplines", 5, func(hp *healing.HealingPriest) []check {
			return []check{
				{"Holy Nova", heal(hp, nova), 1.05 * heal(base, nova)},
				{"Prayer of Mending", heal(hp, pom), 1.05 * heal(base, pom)},
				{"Renew", heal(hp, renew), 1.05 * heal(base, renew)},
				{"Penance's heal", heal(hp, penanceBolt), 1.05 * heal(base, penanceBolt)},
				{"Greater Heal", heal(hp, gh), heal(base, gh)},
			}
		}},
		{"improved_healing", 3, func(hp *healing.HealingPriest) []check {
			return []check{
				{"Greater Heal's cost", cost(hp, gh), 0.85 * cost(base, gh)},
				{"Penance's cost", cost(hp, penance), 0.85 * cost(base, penance)},
				{"Prayer of Mending's cost", cost(hp, pom), 0.85 * cost(base, pom)},
				{"Flash Heal's cost", cost(hp, fh), cost(base, fh)},
			}
		}},
		{"mental_agility", 3, func(hp *healing.HealingPriest) []check {
			return []check{
				{"Renew's cost", cost(hp, renew), 0.9 * cost(base, renew)},
				{"Power Word: Shield's cost", cost(hp, pws), 0.9 * cost(base, pws)},
				{"Holy Nova's cost", cost(hp, nova), 0.9 * cost(base, nova)},
				{"Greater Heal's cost", cost(hp, gh), cost(base, gh)},
			}
		}},
		{"holy_specialization", 5, func(hp *healing.HealingPriest) []check {
			return []check{
				{"Greater Heal's crit", crit(hp, gh), crit(base, gh) + 5},
				{"Renew's crit", crit(hp, renew), crit(base, renew) + 5},
				{"Power Word: Shield's crit", crit(hp, pws), crit(base, pws)},
			}
		}},
		{"divine_fury", 5, func(hp *healing.HealingPriest) []check {
			return []check{
				{"Greater Heal's cast", gh(hp).DefaultCast.CastTime.Seconds(), 2.5},
				{"Heal's cast", hp.Heal[4].DefaultCast.CastTime.Seconds(), 2.5},
				{"Flash Heal's cast", fh(hp).DefaultCast.CastTime.Seconds(), 1.5},
			}
		}},
		{"improved_power_word_shield", 3, func(hp *healing.HealingPriest) []check {
			return []check{{"the shield", absorb(hp), 1.2 * absorb(base)}}
		}},
		{"soul_warding", 1, func(hp *healing.HealingPriest) []check {
			return []check{
				{"Power Word: Shield's cost", cost(hp, pws), 0.85 * cost(base, pws)},
				{"Power Word: Shield's cooldown", pws(hp).CD.Duration.Seconds(), 0},
			}
		}},
		{"meditation", 3, func(hp *healing.HealingPriest) []check {
			return []check{{"regen while casting", hp.PseudoStats.SpiritRegenRateCasting, 0.5}}
		}},
		{"twilight_focus", 3, func(hp *healing.HealingPriest) []check {
			return []check{
				{"Greater Heal's pushback chance", gh(hp).PushbackReduction, 0.70},
				{"Prayer of Healing's pushback chance", top(hp.PrayerOfHealing).PushbackReduction, 0.70},
			}
		}},
		{"spiritual_guidance", 5, func(hp *healing.HealingPriest) []check {
			return []check{{"healing power", hp.GetStat(stats.HealingPower), base.GetStat(stats.HealingPower) + 0.25*hp.GetStat(stats.Spirit)}}
		}},
	} {
		_, hp, _ := talentSim(t, map[string]int{c.talent: c.rank})
		for _, k := range c.checks(hp) {
			if math.Abs(k.got-k.want) > 1e-6*max(1, math.Abs(k.want)) {
				t.Errorf("%s %d: %s is %.4f, want %.4f", c.talent, c.rank, k.what, k.got, k.want)
			}
		}
	}
}

// fightTalent runs the raid on a priest with talents, casting lines, and returns the priest's
// metrics and the result.
func fightTalent(t *testing.T, ranks map[string]int, iterations int32, events []*proto.RaidDamageEvent, mages int, lines ...string) (*proto.RaidSimResult, *proto.UnitMetrics) {
	rsr := priestRaid(t, priestTalents(ranks), mages, events...)
	rsr.SimOptions.Iterations = iterations
	res, m, _, _ := fightPriest(t, rsr, lines...)
	return res, m
}

var greaterHealAtTank = castLine(priest.GreaterHealRanks[5].SpellID(), "Tank", "")

// Divine Aegis shields a heal's target for 15% of each crit (to 1200), and the shields absorb.
func TestHealingPriestDivineAegis(t *testing.T) {
	_, m := fightTalent(t, map[string]int{"divine_aegis": 3}, 20, nil, 0, greaterHealAtTank)
	gh := actionMetrics(m, priest.GreaterHealRanks[5].SpellID(), -1)
	aegis := actionMetrics(m, 431624, -1)
	if gh.crits == 0 || aegis.healing == 0 {
		t.Fatalf("%d crits, %.0f healing from Divine Aegis", gh.crits, aegis.healing)
	}
	// What the shields were worth, absorbed or not, is 15% of the crits, less what went over the
	// 1200 limit and the last shield's rest at the fight's end.
	if got, most := aegis.healing, 0.15*gh.critHealing; got > most*(1+1e-9) || got < 0.7*most || got-aegis.overhealing <= 0 {
		t.Errorf("Divine Aegis booked %.0f (%.0f absorbed), want up to %.0f (15%% of %.0f crit healing)", got, got-aegis.overhealing, most, gh.critHealing)
	}
	// A crit adds to a shield still up, to 20 a level.
	sim, hp, tank := talentSim(t, map[string]int{"divine_aegis": 3})
	sim.AddPendingAction(&core.PendingAction{NextActionAt: time.Second, OnAction: func(sim *core.Simulation) {
		shield := hp.DivineAegisAbsorbs.Get(tank)
		spell := hp.GreaterHeal[5]
		shield.Add(sim, spell, 500, 1200)
		shield.Add(sim, spell, 500, 1200)
		if got := shield.Remaining(); got != 1000 {
			t.Errorf("two shields of 500 left %.0f, want 1000", got)
		}
		shield.Add(sim, spell, 500, 1200)
		if got := shield.Remaining(); got != 1200 {
			t.Errorf("three shields of 500 left %.0f, want the 1200 limit", got)
		}
	}})
	for !sim.Step() {
	}
	_, off := fightTalent(t, map[string]int{}, 20, nil, 0, greaterHealAtTank)
	if a := actionMetrics(off, 431624, -1); a.healing != 0 {
		t.Errorf("Divine Aegis without the talent: %.0f", a.healing)
	}
}

// Inspiration puts 25% more armor on the tank for 15 sec after a crit heal, so it takes less of the
// boss's melee.
func TestHealingPriestInspiration(t *testing.T) {
	on, _ := fightTalent(t, map[string]int{"inspiration": 3}, 40, nil, 0, greaterHealAtTank)
	off, _ := fightTalent(t, map[string]int{}, 40, nil, 0, greaterHealAtTank)
	tankOn, tankOff := playerMetrics(on, 0), playerMetrics(off, 0)
	if tankOn.Dtps.Avg >= 0.99*tankOff.Dtps.Avg {
		t.Errorf("the tank took %.1f DTPS with Inspiration, %.1f without", tankOn.Dtps.Avg, tankOff.Dtps.Avg)
	}
	sim, hp, tank := talentSim(t, map[string]int{"inspiration": 3})
	before := tank.PseudoStats.ArmorMultiplier
	sim.AddPendingAction(&core.PendingAction{NextActionAt: time.Second, OnAction: func(sim *core.Simulation) {
		hp.InspirationAuras.Get(tank).Activate(sim)
		if got := tank.PseudoStats.ArmorMultiplier; !near(got, 1.25*before, 1e-9) {
			t.Errorf("Inspiration took the tank's armor multiplier from %.3f to %.3f, want 25%% more", before, got)
		}
	}})
	for !sim.Step() {
	}
}

// Renewed Hope: heals on a target with Weakened Soul crit 10% more and take 5 sec a heal off it, so
// shields go out more often between Greater Heals.
func TestHealingPriestRenewedHope(t *testing.T) {
	shield := castLine(priest.PowerWordShieldRanks[10].SpellID(), "Tank", "")
	_, on := fightTalent(t, map[string]int{"renewed_hope": 5}, 20, nil, 0, shield, greaterHealAtTank)
	_, off := fightTalent(t, map[string]int{}, 20, nil, 0, shield, greaterHealAtTank)
	pwsOn, pwsOff := actionMetrics(on, priest.PowerWordShieldRanks[10].SpellID(), -1), actionMetrics(off, priest.PowerWordShieldRanks[10].SpellID(), -1)
	if pwsOn.casts <= pwsOff.casts+20 {
		t.Errorf("%d shields with Renewed Hope, %d without", pwsOn.casts, pwsOff.casts)
	}
	ghOn, ghOff := actionMetrics(on, priest.GreaterHealRanks[5].SpellID(), -1), actionMetrics(off, priest.GreaterHealRanks[5].SpellID(), -1)
	rateOn, rateOff := float64(ghOn.crits)/float64(ghOn.casts), float64(ghOff.crits)/float64(ghOff.casts)
	if rateOn < rateOff+0.04 {
		t.Errorf("Greater Heal crits %.1f%% with Renewed Hope, %.1f%% without", 100*rateOn, 100*rateOff)
	}
}

// Litany of Light gives back 10% of a heal's base cost when it isn't the same heal as the last one
// cast: Greater Heal alone only on the fight's first cast, Renew between Greater Heals on every
// Renew and the Greater Heal after it.
func TestHealingPriestLitanyOfLight(t *testing.T) {
	gained := func(m *proto.UnitMetrics) float64 {
		for _, r := range m.Resources {
			if r.Id.GetSpellId() == 1317006 {
				return r.Gain
			}
		}
		return 0
	}
	litany := map[string]int{"litany_of_light": 2}
	_, alone := fightTalent(t, litany, 10, nil, 0, greaterHealAtTank)
	if got := gained(alone) / 10; !near(got, 71, 1e-6) {
		t.Errorf("Greater Heal alone gave back %.1f mana a fight, want 71 (the first cast)", got)
	}
	renewID := priest.RenewRanks[10].SpellID()
	renew := castLine(renewID, "Tank", `{"not":{"val":{"dotIsActive":{"spellId":{"spellId":`+strconv.Itoa(int(renewID))+`},"targetUnit":{"type":"Tank"}}}}}`)
	_, mixed := fightTalent(t, litany, 10, nil, 0, renew, greaterHealAtTank)
	renews := float64(actionMetrics(mixed, renewID, -1).casts)
	if got, most := gained(mixed), renews*(41+71); got > most+1e-6 || got < most-10*71-1e-6 {
		t.Errorf("%.0f Renews between Greater Heals gave back %.0f mana, want %.0f (less a Greater Heal a fight at most)", renews, got, most)
	}
	_, off := fightTalent(t, map[string]int{}, 10, nil, 0, renew, greaterHealAtTank)
	if got := gained(off); got != 0 {
		t.Errorf("no Litany of Light gave back %.0f mana", got)
	}
}

// Holy Binding heals the priest for half of each Flash Heal on another.
func TestHealingPriestHolyBinding(t *testing.T) {
	pulse := &proto.RaidDamageEvent{School: proto.SpellSchool_SpellSchoolFire, MinDamage: 2000, MaxDamage: 2000, Who: proto.RaidDamageEvent_Everyone, EverySeconds: 10, FirstAt: 1}
	fh := castLine(priest.FlashHealRanks[7].SpellID(), "Tank", "")
	_, m := fightTalent(t, map[string]int{"holy_binding": 1}, 10, []*proto.RaidDamageEvent{pulse}, 0, fh)
	flash := actionMetrics(m, priest.FlashHealRanks[7].SpellID(), -1)
	binding := actionMetrics(m, 1328570, -1)
	if !near(binding.healing, 0.5*flash.healing, 1e-6) || binding.healing-binding.overhealing <= 0 {
		t.Errorf("Holy Binding healed us for %.0f (%.0f effective) of %.0f Flash Heal healing, want half", binding.healing, binding.healing-binding.overhealing, flash.healing)
	}
	_, off := fightTalent(t, map[string]int{}, 10, []*proto.RaidDamageEvent{pulse}, 0, fh)
	if b := actionMetrics(off, 1328570, -1); b.healing != 0 {
		t.Errorf("Holy Binding without the talent healed %.0f", b.healing)
	}
}

// Inner Focus makes the next heal free and gives it 25% more crit; Power Infusion is 20% more
// healing for 15 sec.
func TestHealingPriestInnerFocusAndPowerInfusion(t *testing.T) {
	sim, hp, tank := talentSim(t, map[string]int{"inner_focus": 1, "power_infusion": 1})
	gh := hp.GreaterHeal[5]
	sim.AddPendingAction(&core.PendingAction{NextActionAt: time.Second, OnAction: func(sim *core.Simulation) {
		critBefore, healBefore := gh.BonusCritRating, gh.ExpectedHealing(tank, 1000)
		hp.InnerFocusAura.Activate(sim)
		if got := gh.Cost.GetCurrentCost(); got != 0 {
			t.Errorf("Greater Heal costs %.0f under Inner Focus", got)
		}
		if got := gh.BonusCritRating - critBefore; got != 25 {
			t.Errorf("Inner Focus gives Greater Heal %.0f crit, want 25", got)
		}
		hp.InnerFocusAura.Deactivate(sim)
		hp.PowerInfusion.Cast(sim, &hp.Unit)
		if got := gh.ExpectedHealing(tank, 1000); !near(got, 1.2*healBefore, 1e-9) {
			t.Errorf("Greater Heal heals %.0f under Power Infusion, %.0f without", got, healBefore)
		}
	}})
	for !sim.Step() {
	}
}

// Blessed Recovery heals 25% of a melee hit for over 30% of our health over 6 sec.
func TestHealingPriestBlessedRecovery(t *testing.T) {
	rsr := priestRaid(t, priestTalents(map[string]int{"blessed_recovery": 3}), 0)
	rsr.SimOptions.Iterations = 10
	// The boss hits the priest, for more than 30% of its health.
	rsr.Encounter.Targets[0].TankIndex = 0
	rsr.Raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 1}}
	rsr.Encounter.Targets[0].MinBaseDamage = 1500
	rsr.Encounter.Targets[0].SwingSpeed = 3
	res := runRaid(t, rsr)
	m := playerMetrics(res, 1)
	br := actionMetrics(m, 27813, -1)
	if br.healing <= 0 || br.hits == 0 {
		t.Errorf("Blessed Recovery healed %.0f in %d ticks", br.healing, br.hits)
	}
	_, off := fightTalent(t, map[string]int{}, 10, nil, 0)
	if b := actionMetrics(off, 27813, -1); b.healing != 0 {
		t.Errorf("Blessed Recovery without the talent healed %.0f", b.healing)
	}
}

// The absorbed part of a hit never reaches the tank score's replay: what it records is what the
// shield let through, so the replay at the stand-in healer's HPS agrees with the fight on every
// death, and shields lower the deaths.
func TestHealingPriestShieldTankScore(t *testing.T) {
	const testStandInHps = 300
	request := func(shield bool) *proto.RaidSimRequest {
		rsr := priestRaid(t, priestTalents(map[string]int{"soul_warding": 1}), 0)
		rsr.Encounter.TrackHealth = false
		rsr.Encounter.Duration = 60
		rsr.Encounter.Targets[0].MinBaseDamage = 2600
		rsr.SimOptions.Iterations = 300
		tank := rsr.Raid.Parties[0].Players[0]
		tank.HealingModel = &proto.HealingModel{Hps: testStandInHps, CadenceSeconds: 2, CadenceVariation: 1.5}
		tank.TankScoring = &proto.TankScoring{}
		if shield {
			healerRotation(rsr, 1, castLine(priest.PowerWordShieldRanks[10].SpellID(), "Tank", ""))
		}
		return rsr
	}
	deaths := map[bool]float64{}
	for _, shield := range []bool{false, true} {
		res := runRaid(t, request(shield))
		tank := playerMetrics(res, 0)
		survive := tank.GetTankScore().GetSurviveHps()
		if len(survive) != 300 {
			t.Fatalf("%d tank score iterations", len(survive))
		}
		replayDied := 0
		for _, h := range survive {
			if h > testStandInHps*(1+1e-6) {
				replayDied++
			}
		}
		if got := float64(replayDied) / 300; math.Abs(got-tank.ChanceOfDeath) > 1.5/300 {
			t.Errorf("shield %v: the replay dies in %.1f%% of fights, the fight in %.1f%%", shield, 100*got, 100*tank.ChanceOfDeath)
		}
		deaths[shield] = tank.ChanceOfDeath
	}
	t.Logf("the tank died in %.1f%% of fights shielded, %.1f%% not", 100*deaths[true], 100*deaths[false])
	if deaths[true] >= deaths[false] || deaths[false] == 0 {
		t.Errorf("the tank died in %.1f%% of fights shielded, %.1f%% not", 100*deaths[true], 100*deaths[false])
	}
}
