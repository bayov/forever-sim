package dpsrogue

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The rogue lines of Blizzard's November 2 development notes and client 1.60.1.70753
// (orchestration/round_70753/rogue.md).

// Poised: the rogue opens the fight at full Energy, so the first ability costs 10 less, and the
// buff is gone after it.
func TestForeverPoisedFirstAbility(t *testing.T) {
	sim, rogue := newForeverRogueSim(t, "", 18832, 18805, 1) // Brutality Blade, Core Hound Tooth
	target := rogue.CurrentTarget
	if !rogue.PoisedAura.IsActive() {
		t.Fatal("Poised isn't up at the pull")
	}
	for i, want := range []float64{35, 45} {
		rogue.AddEnergy(sim, 100, rogue.EnergyRefundMetrics)
		rogue.SetGCDTimer(sim, sim.CurrentTime)
		if got := rogue.SinisterStrike.Cost.GetCurrentCost(); got != want {
			t.Errorf("Sinister Strike %d costs %.0f, want %.0f", i+1, got, want)
		}
		rogue.SinisterStrike.Cast(sim, target)
	}
	if rogue.PoisedAura.IsActive() {
		t.Error("Poised is still up after a Sinister Strike")
	}
}

// Poised comes back at the pull when a prepull ability spent it and the Energy filled up again,
// and stays off when the Energy didn't.
func TestForeverPoisedAtThePull(t *testing.T) {
	for _, refill := range []bool{true, false} {
		sim, rogue := newForeverRogueSim(t, "", 18832, 18805, 1)
		rogue.PoisedAura.Deactivate(sim)
		rogue.SpendEnergy(sim, 40, rogue.EnergyRefundMetrics)
		if refill {
			rogue.AddEnergy(sim, 40, rogue.EnergyRefundMetrics)
		}
		for sim.CurrentTime <= 0 {
			sim.Step()
		}
		if rogue.PoisedAura.IsActive() != refill {
			t.Errorf("full Energy at the pull %v: Poised up %v", refill, rogue.PoisedAura.IsActive())
		}
	}
}

// Shank: 5 Energy, gives back 25 at level 60, 10 sec cooldown, never misses.
func TestForeverShank(t *testing.T) {
	sim, rogue := newForeverRogueSim(t, "", 18832, 18805, 1)
	if rogue.Shank == nil {
		t.Fatal("no Shank at level 60")
	}
	target := rogue.CurrentTarget
	rogue.PoisedAura.Deactivate(sim)
	rogue.AddStatDynamic(sim, stats.MeleeHit, -20*core.MeleeHitRatingPerHitChance)
	for i := 0; i < 300; i++ {
		rogue.SpendEnergy(sim, rogue.CurrentEnergy(), rogue.EnergyRefundMetrics)
		rogue.AddEnergy(sim, 50, rogue.EnergyRefundMetrics)
		rogue.SetGCDTimer(sim, sim.CurrentTime)
		rogue.Shank.CD.Reset()
		dodges := rogue.Shank.SpellMetrics[target.UnitIndex].Dodges + rogue.Shank.SpellMetrics[target.UnitIndex].Parries
		if !rogue.Shank.Cast(sim, target) {
			t.Fatal("Shank didn't cast")
		}
		landed := rogue.Shank.SpellMetrics[target.UnitIndex].Dodges+rogue.Shank.SpellMetrics[target.UnitIndex].Parries == dodges
		if want := core.TernaryFloat64(landed, 70, 49); math.Abs(rogue.CurrentEnergy()-want) > 1e-6 {
			t.Fatalf("Energy after a Shank (landed %v) is %.1f, want %.0f", landed, rogue.CurrentEnergy(), want)
		}
	}
	if misses := rogue.Shank.SpellMetrics[target.UnitIndex].Misses; misses != 0 {
		t.Errorf("Shank missed %d times", misses)
	}
	if got := rogue.Shank.CD.Duration; got != 10*time.Second {
		t.Errorf("Shank's cooldown is %s, want 10s", got)
	}
}

// Venom is a passive now: a Shank gives 30% more poison damage and 20% more poison procs for 6 sec.
func TestForeverVenomFromShank(t *testing.T) {
	sim, rogue := newForeverRogueSim(t, "00000000000000001", 18805, 18816, 1) // Venom only; Core Hound Tooth, Perdition's Blade
	if rogue.VenomAura == nil {
		t.Fatal("the talents give no Venom")
	}
	target := rogue.CurrentTarget
	base := rogue.GetDeadlyPoisonProcChance()
	for !rogue.VenomAura.IsActive() {
		rogue.AddEnergy(sim, 100, rogue.EnergyRefundMetrics)
		rogue.SetGCDTimer(sim, sim.CurrentTime)
		rogue.Shank.CD.Reset()
		rogue.Shank.Cast(sim, target)
	}
	if got := rogue.GetDeadlyPoisonProcChance() - base; math.Abs(got-0.2) > 1e-9 {
		t.Errorf("Venom adds %.2f to the poison chance, want 0.20", got)
	}
	if got := rogue.VenomAura.RemainingDuration(sim); got != 6*time.Second {
		t.Errorf("Venom lasts %s, want 6s", got)
	}
}

// Ghostly Strike lets the next Ambush or Garrote within 7 sec go without Stealth.
func TestForeverGhostlyStrikeOpener(t *testing.T) {
	sim, rogue := newForeverRogueSim(t, subtletyPresetTalents, 18805, 18805, 1)
	if rogue.GhostlyStrike == nil {
		t.Fatal("the talents give no Ghostly Strike")
	}
	target := rogue.CurrentTarget
	rogue.AddEnergy(sim, 100, rogue.EnergyRefundMetrics)
	rogue.SetGCDTimer(sim, sim.CurrentTime)
	if rogue.Ambush.CanCast(sim, target) {
		t.Fatal("Ambush is castable out of Stealth")
	}
	rogue.GhostlyStrike.Cast(sim, target)
	rogue.AddEnergy(sim, 100, rogue.EnergyRefundMetrics)
	rogue.SetGCDTimer(sim, sim.CurrentTime)
	if !rogue.Ambush.CanCast(sim, target) || !rogue.Garrote.CanCast(sim, target) {
		t.Fatal("Ambush or Garrote isn't castable after Ghostly Strike")
	}
	rogue.Ambush.Cast(sim, target)
	rogue.AddEnergy(sim, 100, rogue.EnergyRefundMetrics)
	rogue.SetGCDTimer(sim, sim.CurrentTime)
	if rogue.Garrote.CanCast(sim, target) {
		t.Error("Garrote is still castable after the Ambush spent Ghostly Strike's opening")
	}
}

// Hemorrhage raises the rogue's Garrote by 15% too.
func TestForeverHemorrhageGarrote(t *testing.T) {
	sim, rogue := newForeverRogueSim(t, subtletyPresetTalents, 18805, 18805, 1)
	target := rogue.CurrentTarget
	dot := rogue.Garrote.Dot(target)
	dot.Apply(sim)
	metrics := &rogue.Garrote.SpellMetrics[target.UnitIndex]
	rogue.HemorrhageAuras.Get(target).Activate(sim)
	ap := rogue.Garrote.MeleeAttackPower(target)
	multiplier := rogue.Garrote.AttackerDamageMultiplier(rogue.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand], true)
	want := (92 + 0.03*ap) * 1.15 * multiplier
	dot.TickOnce(sim)
	if got := metrics.TotalDamage; math.Abs(got-want) >= 1 {
		t.Errorf("Garrote under Hemorrhage ticks for %.3f, want %.3f", got, want)
	}
}

// Mutilate's halves strike for 90% of (weapon damage + 67) at level 60, was 75%: against a
// target with no armor, the same seed and no crits, the damage is 0.9 times the weapon part.
func TestForeverMutilateNinetyPercent(t *testing.T) {
	sim, rogue := newMutilateRogueSim(t, 600)
	target := rogue.CurrentTarget
	target.AddStatsDynamic(sim, stats.Stats{stats.Armor: -target.GetStat(stats.Armor)})
	rogue.Mutilate.BonusCritRating -= 1000 * core.CritRatingPerCritChance
	rogue.PseudoStats.InFrontOfTarget = false
	rogue.AddStatDynamic(sim, stats.MeleeHit, 20*core.MeleeHitRatingPerHitChance)
	rogue.AddStatDynamic(sim, stats.Expertise, 20*core.ExpertiseRatingPerExpertiseChance)
	rogue.AddEnergy(sim, 100, rogue.EnergyRefundMetrics)
	rogue.SetGCDTimer(sim, sim.CurrentTime)
	mh := rogue.AutoAttacks.MH()
	ap := rogue.Mutilate.MeleeAttackPower(target)
	low := mh.BaseDamageMin + mh.NormalizedSwingSpeed*ap/mh.AttackPowerPerDPS
	high := mh.BaseDamageMax + mh.NormalizedSwingSpeed*ap/mh.AttackPowerPerDPS
	rogue.Mutilate.Cast(sim, target)
	got := rogue.Mutilate.SpellMetrics[target.UnitIndex].TotalDamage / rogue.Mutilate.DamageMultiplier /
		rogue.Mutilate.AttackerDamageMultiplier(rogue.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand], false)
	if got < 0.9*(67+low)-1e-6 || got > 0.9*(67+high)+1e-6 {
		t.Errorf("a Mutilate main hand half hit for %.2f, want between %.2f and %.2f", got, 0.9*(67+low), 0.9*(67+high))
	}
}

// Flawless Execution takes 15 off Eviscerate; Aggression 3/3 raises Sinister Strike and Eviscerate
// by 15% and no longer Backstab.
func TestForeverFlawlessExecutionAndAggression(t *testing.T) {
	sim, rogue := newForeverRogueSim(t, "-0000000000100003", 18805, 18805, 1)
	if !rogue.Talents.FlawlessExecution || rogue.Talents.Aggression != 3 {
		t.Fatal("the talents give no Flawless Execution or no Aggression 3")
	}
	rogue.PoisedAura.Deactivate(sim)
	if got := rogue.Eviscerate.Cost.GetCurrentCost(); got != 20 {
		t.Errorf("Eviscerate costs %.0f, want 20", got)
	}
	if got := rogue.SinisterStrike.DamageMultiplier; math.Abs(got-1.15) > 1e-9 {
		t.Errorf("Sinister Strike's multiplier is %.3f, want 1.15", got)
	}
	if got := rogue.Backstab.DamageMultiplier; math.Abs(got-1.5) > 1e-9 {
		t.Errorf("Backstab's multiplier is %.3f, want 1.5", got)
	}
}
