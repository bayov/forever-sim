package shadow

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
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
	sim, sp := newShadowPriestSim(t, "00503523103--50052101231302051")
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

// TestShadowformCritBonus checks Shadowform's +100% crit bonus: Mind Blast and, since 70425,
// Shadow Word: Death crit for 200%, and Mind Blast goes back after Shadowform.
func TestShadowformCritBonus(t *testing.T) {
	sim, sp := newShadowPriestSim(t, "00503523103--50052101231302051")
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
	check("Shadowform", death, 2.0)
	sp.ShadowformAura.Deactivate(sim)
	check("caster form again", mindBlast, 1.5)
}

// TestDevouringPlagueCooldown checks the 24 sec cooldown every race has under Forever (70425).
func TestDevouringPlagueCooldown(t *testing.T) {
	_, sp := newShadowPriestSim(t, "00503523103--50052101231302051")
	if got := topRank(t, "Devouring Plague", sp.DevouringPlague).CD.Duration; got != priest.DevouringPlagueCooldown {
		t.Errorf("Devouring Plague cooldown %s, want %s", got, priest.DevouringPlagueCooldown)
	}
}

// TestInnerFocusCrit checks that a waiting Inner Focus adds 25% crit to the direct spells
// it can be spent on but not to Shadow Word: Death or to spells that only tick (Shadow Word:
// Pain, Devouring Plague, Mind Flay), whose ticks roll crit live under Forever.
func TestInnerFocusCrit(t *testing.T) {
	sim, sp := newShadowPriestSim(t, "00503521103--50052101231302051")
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

// TestInnerFocusCost checks that Inner Focus makes Mind Blast and Devouring Plague free.
// Devouring Plague is in the free cast's class mask through the fourth mask word.
func TestInnerFocusCost(t *testing.T) {
	sim, sp := newShadowPriestSim(t, "00503521103--50052101231302051")
	mindBlast := topRank(t, "Mind Blast", sp.MindBlast)
	plague := topRank(t, "Devouring Plague", sp.DevouringPlague)
	sp.InnerFocusAura.Activate(sim)
	if got := mindBlast.Cost.GetCurrentCost(); got != 0 {
		t.Errorf("Mind Blast costs %.2f under Inner Focus, want 0", got)
	}
	if got := plague.Cost.GetCurrentCost(); got != 0 {
		t.Errorf("Devouring Plague costs %.2f under Inner Focus, want 0", got)
	}
}

// TestDownrankCoefficient checks Forever's downranking penalty at level 60: a rank keeps its
// full spell power coefficient up to 17 levels below the caster and loses 5% of it for each
// level past that (Smite rank 7 at 46 keeps it, rank 6 at 38 keeps 75%, rank 4 at 22 nothing).
func TestDownrankCoefficient(t *testing.T) {
	_, sp := newShadowPriestSim(t, "00503523103--50052101231302051")
	checks := []struct {
		rank int
		want float64
	}{
		{8, priest.SmiteSpellCoef[8]},
		{7, priest.SmiteSpellCoef[7]},
		{6, priest.SmiteSpellCoef[6] * 0.75},
		{5, priest.SmiteSpellCoef[5] * 0.35},
		{4, 0},
	}
	for _, c := range checks {
		got := sp.Smite[c.rank].BonusCoefficient
		if math.Abs(got-c.want) > 1e-9 {
			t.Errorf("Smite rank %d coefficient %.4f, want %.4f", c.rank, got, c.want)
		}
	}
}

// TestSmiteRankTuning checks the report assumption the user ruled on 2026-10-15: Smite rank 8
// deals 9% more than the client's 166.62 to 186.38, and rank 7 is unchanged.
func TestSmiteRankTuning(t *testing.T) {
	for _, c := range []struct {
		rank      int
		low, high float64
	}{
		{8, 166.62 * 1.09, 186.38 * 1.09},
		{7, 123.74, 139.26},
	} {
		low, high := priest.SmiteBaseDamageAt(c.rank, 60)
		if math.Abs(low-c.low) > 1e-9 || math.Abs(high-c.high) > 1e-9 {
			t.Errorf("Smite rank %d at 60: %.2f to %.2f, want %.2f to %.2f", c.rank, low, high, c.low, c.high)
		}
	}
}

// TestCooldowns70425 checks the cooldowns client 1.60.1.70425 changed: Devouring Plague 24
// sec, Shadow Word: Death 18 sec, Vampiric Embrace 30 sec.
func TestCooldowns70425(t *testing.T) {
	_, sp := newShadowPriestSim(t, "00503523103--50052101231302051")
	for _, c := range []struct {
		name  string
		spell *core.Spell
		want  time.Duration
	}{
		{"Devouring Plague", topRank(t, "Devouring Plague", sp.DevouringPlague), 24 * time.Second},
		{"Shadow Word: Death", topRank(t, "Shadow Word: Death", sp.ShadowWordDeath), 18 * time.Second},
		{"Vampiric Embrace", sp.VampiricEmbrace, 30 * time.Second},
	} {
		if got := c.spell.CD.Duration; got != c.want {
			t.Errorf("%s cooldown %s, want %s", c.name, got, c.want)
		}
	}
}

// TestImprovedMindFlay checks 70425's Improved Mind Flay: each tick deals 10% a rank more
// than the one before it, compounding (1, 1.2, 1.44 at 2 points).
func TestImprovedMindFlay(t *testing.T) {
	_, sp := newShadowPriestSim(t, "00503523103--50052101231302051")
	for tick, want := range []float64{1, 1, 1.2, 1.44} {
		if tick == 0 {
			continue
		}
		if got := sp.ImprovedMindFlayTickMultiplier(int32(tick)); math.Abs(got-want) > 1e-9 {
			t.Errorf("Improved Mind Flay tick %d multiplier %.4f, want %.4f", tick, got, want)
		}
	}
}

// TestDarkInspiration checks the 70425 talent's Spirit buff: 3% a rank for 10 sec.
func TestDarkInspiration(t *testing.T) {
	sim, sp := newShadowPriestSim(t, "00503523103--50052101231302051")
	spirit := sp.GetStat(stats.Spirit)
	sp.DarkInspirationAura.Activate(sim)
	// Primary stats are floored, so within a point.
	if got, want := sp.GetStat(stats.Spirit), spirit*1.09; math.Abs(got-want) > 1 {
		t.Errorf("Spirit %.2f under Dark Inspiration 3/5, want %.2f", got, want)
	}
	if got := sp.DarkInspirationAura.Duration; got != 10*time.Second {
		t.Errorf("Dark Inspiration lasts %s, want 10s", got)
	}
}

// TestTwilightFocus checks the pushback protection: 70% at 3/3 on the cast and channelled
// spells in Twilight Focus's mask, none on the instants.
func TestTwilightFocus(t *testing.T) {
	_, sp := newShadowPriestSim(t, "-3025011323-50052101231302051")
	for _, c := range []struct {
		name  string
		spell *core.Spell
		want  float64
	}{
		{"Mind Blast", topRank(t, "Mind Blast", sp.MindBlast), 0.70},
		{"Mind Flay", topRank(t, "Mind Flay", sp.MindFlay), 0.70},
		{"Shadow Word: Pain", topRank(t, "Shadow Word: Pain", sp.ShadowWordPain), 0},
	} {
		if got := c.spell.PushbackReduction; math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%s ignores pushback %.2f of the time, want %.2f", c.name, got, c.want)
		}
	}
}
