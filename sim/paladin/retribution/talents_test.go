package retribution

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestHolyPowerOnItsSpellsOnly checks that Holy Power (client 1.60.1.70338, spell 5923) is a
// crit modifier on the paladin's own spells and not a spell crit stat, which would also reach
// Dragonbreath Chili, Shadow Oil and the other consumables. At 5 points Holy Shock gets 15%
// and Judgement of Righteousness 5%.
func TestHolyPowerOnItsSpellsOnly(t *testing.T) {
	const withHolyPower = "55003003201121251--05225230001"
	const withoutHolyPower = "55003003201121201--05225230001"

	_, with := retSim(t, withHolyPower, arcaniteReaper)
	_, without := retSim(t, withoutHolyPower, arcaniteReaper)

	if got, want := with.GetStat(stats.SpellCrit), without.GetStat(stats.SpellCrit); got != want {
		t.Errorf("spell crit stat with Holy Power = %.2f, want %.2f (no Holy Power)", got, want)
	}

	checks := []struct {
		name    string
		spellID int32
		want    float64
	}{
		{"Holy Shock", 20930, 15},
		{"Judgement of Righteousness", 20286, 5},
	}
	for _, c := range checks {
		spell := with.GetSpell(core.ActionID{SpellID: c.spellID})
		if spell == nil {
			t.Fatalf("%s (%d) is not registered", c.name, c.spellID)
		}
		if spell.BonusCritRating != c.want {
			t.Errorf("%s bonus crit = %.1f, want %.1f", c.name, spell.BonusCritRating, c.want)
		}
	}
}

// TestConsecratedGroundOnConsecration checks that Consecrated Ground (client 1310905, 5% a
// point to Holy spells against targets in the Consecration) reaches Consecration's own ticks:
// at 2 points the ticks snapshot 10% more than without the talent.
func TestConsecratedGroundOnConsecration(t *testing.T) {
	snapshot := func(talents string) float64 {
		sim, ret := retSim(t, talents, arcaniteReaper)
		castByID(sim, ret, 20924)
		return ret.GetSpell(core.ActionID{SpellID: 20924}).AOEDot().SnapshotAttackerMultiplier
	}
	with := snapshot("55003003201121251--05225230001")
	without := snapshot("55003003201121051--05225230001")
	if ratio := with / without; ratio < 1.0999 || ratio > 1.1001 {
		t.Errorf("Consecration tick multiplier with Consecrated Ground 2 / without = %.4f, want 1.1000", ratio)
	}
}

// TestSwiftJudgement checks Swift Judgement (client 1310994): it finishes Judgement's cooldown
// and makes the next Judgement free, and the Judgement after that costs mana again.
func TestSwiftJudgement(t *testing.T) {
	sim, ret := retSim(t, "02-5032210021-05225331001330311", arcaniteReaper)
	judgement := ret.GetSpell(core.ActionID{SpellID: 20271})
	swift := ret.GetSpell(core.ActionID{SpellID: 1310994})
	seal := ret.GetSpell(core.ActionID{SpellID: 20293})
	if judgement == nil || swift == nil || seal == nil {
		t.Fatalf("missing Judgement, Swift Judgement or Seal of Righteousness")
	}
	target := ret.CurrentTarget
	seal.Cast(sim, target)
	judgement.CD.Use(sim)
	if judgement.CD.IsReady(sim) {
		t.Fatalf("Judgement ready right after use")
	}
	swift.Cast(sim, target)
	if !judgement.CD.IsReady(sim) {
		t.Errorf("Judgement still on cooldown after Swift Judgement")
	}
	if cost := judgement.Cost.GetCurrentCost(); cost != 0 {
		t.Errorf("Judgement after Swift Judgement costs %.1f, want 0", cost)
	}
	judgement.Cast(sim, target)
	if cost := judgement.Cost.GetCurrentCost(); cost <= 0 {
		t.Errorf("the Judgement after the free one costs %.1f, want more than 0", cost)
	}
}

// TestJudgementOfTheCrusaderRank checks that a level 60 paladin's own Judgement of the Crusader
// puts the rank 6 debuff on the target: 161 Holy damage taken under Forever (140 x 1.15). Every
// rank used to share one debuff aura, so rank 1's 23 won.
func TestJudgementOfTheCrusaderRank(t *testing.T) {
	sim, ret := retSim(t, "52003-503-05225331001330311", arcaniteReaper)
	castByID(sim, ret, 20308)
	castByID(sim, ret, 20271)
	if got := ret.CurrentTarget.PseudoStats.SchoolBonusDamageTaken[stats.SchoolIndexHoly]; got < 160.99 || got > 161.01 {
		t.Errorf("Holy damage taken after Judgement of the Crusader = %.2f, want 161", got)
	}
}

// TestSealOfRighteousnessSpellPowerOnce checks that a Seal of Righteousness proc under Forever
// takes the tooltip's 12% of spell power with a two-hander (client 70425, description variables
// 1023) once: 500 more spell power adds as much to a proc as to the spell power part alone.
func TestSealOfRighteousnessSpellPowerOnce(t *testing.T) {
	sim, ret := retSim(t, "", arcaniteReaper)
	target := ret.CurrentTarget
	proc := ret.GetSpell(core.ActionID{SpellID: 25713})
	if proc == nil {
		t.Fatalf("no Seal of Righteousness rank 9 proc")
	}
	if proc.BonusCoefficient != 0.12 {
		t.Errorf("proc spell power coefficient = %.4f, want 0.12 with a two-hander", proc.BonusCoefficient)
	}
	proc.BonusCritRating = -10000 // no crits, so every proc is a plain hit
	metrics := &proc.SpellMetrics[target.UnitIndex]
	hit := func() float64 {
		before := metrics.TotalDamage
		proc.ApplyEffects(sim, target, proc)
		return metrics.TotalDamage - before
	}
	spellPowerPart := func() float64 {
		return proc.CalcDamage(sim, target, 0, proc.OutcomeAlwaysHit).Damage
	}
	lowHit, lowPart := hit(), spellPowerPart()
	ret.AddStatDynamic(sim, stats.SpellPower, 500)
	highHit, highPart := hit(), spellPowerPart()
	if highPart-lowPart < 1 {
		t.Fatalf("500 spell power adds %.2f to the spell power part", highPart-lowPart)
	}
	if got, want := highHit-lowHit, highPart-lowPart; got < want-0.01 || got > want+0.01 {
		t.Errorf("500 spell power adds %.2f to a proc, want %.2f (counted once)", got, want)
	}
}
