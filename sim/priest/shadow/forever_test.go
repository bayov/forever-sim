package shadow

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/priest"
)

func newShadowPriestSim(t *testing.T, talents string) (*core.Simulation, *ShadowPriest) {
	player := &proto.Player{
		Race:          proto.Race_RaceUndead,
		Class:         proto.Class_ClassPriest,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec:          PlayerOptionsBasic,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 10, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*ShadowPriest)
}

func topRank(t *testing.T, name string, ranks []*core.Spell) *core.Spell {
	for i := len(ranks) - 1; i >= 0; i-- {
		if ranks[i] != nil {
			return ranks[i]
		}
	}
	t.Fatalf("no %s rank", name)
	return nil
}

// TestShadowformCosts checks that Shadowform's half cost multiplies with the percent cost
// talents, which add up among themselves (Devouring Contagion 2/2 and Mental Agility 3/3 on
// Devouring Plague: 40% of the cost, then half of that), and which spells each one reaches.
// Mental Agility leaves out Mind Blast, Mind Flay and Shadow Word: Death.
func TestShadowformCosts(t *testing.T) {
	sim, sp := newShadowPriestSim(t, "005030231305--500321501201302051")
	sp.ShadowformAura.Activate(sim)

	checks := []struct {
		name  string
		spell *core.Spell
		want  float64
	}{
		{"Devouring Plague", topRank(t, "Devouring Plague", sp.DevouringPlague), 985 * 0.4 * 0.5},
		{"Shadow Word: Pain", topRank(t, "Shadow Word: Pain", sp.ShadowWordPain), 470 * 0.9 * 0.5},
		{"Mind Blast", topRank(t, "Mind Blast", sp.MindBlast), 350 * 0.5},
		{"Mind Flay", topRank(t, "Mind Flay", sp.MindFlay), 205 * 0.5},
		{"Shadow Word: Death", topRank(t, "Shadow Word: Death", sp.ShadowWordDeath), 340 * 0.5},
	}
	for _, c := range checks {
		if got := c.spell.Cost.GetCurrentCost(); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%s costs %.2f in Shadowform, want %.2f", c.name, got, c.want)
		}
	}
}

// TestShadowformCritBonus checks Shadowform's +100% crit bonus: Mind Blast crits for 200%,
// Shadow Word: Death (not in its class mask) for 150%, and both go back after Shadowform.
func TestShadowformCritBonus(t *testing.T) {
	sim, sp := newShadowPriestSim(t, "005030231305--500321501201302051")
	target := sp.CurrentTarget
	at := sp.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand]
	mindBlast := topRank(t, "Mind Blast", sp.MindBlast)
	death := topRank(t, "Shadow Word: Death", sp.ShadowWordDeath)

	check := func(label string, spell *core.Spell, want float64) {
		if got := spell.CritMultiplier(at); math.Abs(got-want) > 1e-9 {
			t.Errorf("%s: %s crits for %.3f, want %.3f", label, spell.ActionID, got, want)
		}
	}
	check("caster form", mindBlast, 1.5)
	sp.ShadowformAura.Activate(sim)
	check("Shadowform", mindBlast, 2.0)
	check("Shadowform", death, 1.5)
	sp.ShadowformAura.Deactivate(sim)
	check("caster form again", mindBlast, 1.5)
}

// TestDevouringPlagueCooldown checks the 1 min cooldown every race has under Forever.
func TestDevouringPlagueCooldown(t *testing.T) {
	_, sp := newShadowPriestSim(t, "005030231305--500321501201302051")
	if got := topRank(t, "Devouring Plague", sp.DevouringPlague).CD.Duration; got != priest.DevouringPlagueCooldown {
		t.Errorf("Devouring Plague cooldown %s, want %s", got, priest.DevouringPlagueCooldown)
	}
}

// TestInnerFocusCrit checks that a waiting Inner Focus adds 25% crit to the direct spells
// it can be spent on but not to Shadow Word: Death or to spells that only tick (Shadow Word:
// Pain, Devouring Plague, Mind Flay), whose ticks roll crit live under Forever.
func TestInnerFocusCrit(t *testing.T) {
	sim, sp := newShadowPriestSim(t, "005030211305--500321501201302051")
	checks := []struct {
		name  string
		spell *core.Spell
		bonus float64
	}{
		{"Mind Blast", topRank(t, "Mind Blast", sp.MindBlast), 25},
		{"Shadow Word: Death", topRank(t, "Shadow Word: Death", sp.ShadowWordDeath), 0},
		{"Shadow Word: Pain", topRank(t, "Shadow Word: Pain", sp.ShadowWordPain), 0},
		{"Devouring Plague", topRank(t, "Devouring Plague", sp.DevouringPlague), 0},
		{"Mind Flay", topRank(t, "Mind Flay", sp.MindFlay), 0},
	}
	before := make([]float64, len(checks))
	for i, c := range checks {
		before[i] = c.spell.BonusCritRating
	}
	sp.InnerFocusAura.Activate(sim)
	for i, c := range checks {
		want := before[i] + c.bonus*core.SpellCritRatingPerCritChance
		if got := c.spell.BonusCritRating; math.Abs(got-want) > 1e-9 {
			t.Errorf("%s has %.2f bonus crit rating under Inner Focus, want %.2f", c.name, got, want)
		}
	}
}
