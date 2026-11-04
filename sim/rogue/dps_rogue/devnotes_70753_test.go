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

// Shank (server hotfix push 113483, 1329549 at level 60): 5 Energy, gives back 25, 10 sec cooldown,
// always hits (no miss, dodge or parry) and procs no poison or weapon effect.
func TestForeverShank(t *testing.T) {
	sim, rogue := newForeverRogueSim(t, "", 18832, 18805, 1)
	if rogue.Shank == nil {
		t.Fatal("no Shank at level 60")
	}
	if rogue.Shank.SpellID != 1329549 {
		t.Errorf("Shank's spell ID is %d, want 1329549", rogue.Shank.SpellID)
	}
	if !rogue.Shank.Flags.Matches(core.SpellFlagSuppressWeaponProcs) {
		t.Error("Shank doesn't suppress weapon procs")
	}
	target := rogue.CurrentTarget
	rogue.PoisedAura.Deactivate(sim)
	rogue.AddStatDynamic(sim, stats.MeleeHit, -20*core.MeleeHitRatingPerHitChance)
	for i := 0; i < 300; i++ {
		rogue.SpendEnergy(sim, rogue.CurrentEnergy(), rogue.EnergyRefundMetrics)
		rogue.AddEnergy(sim, 50, rogue.EnergyRefundMetrics)
		rogue.SetGCDTimer(sim, sim.CurrentTime)
		rogue.Shank.CD.Reset()
		if !rogue.Shank.Cast(sim, target) {
			t.Fatal("Shank didn't cast")
		}
		if math.Abs(rogue.CurrentEnergy()-70) > 1e-6 {
			t.Fatalf("Energy after a Shank is %.1f, want 70", rogue.CurrentEnergy())
		}
	}
	if m := rogue.Shank.SpellMetrics[target.UnitIndex]; m.Dodges+m.Parries != 0 {
		t.Errorf("Shank was dodged or parried %d times", m.Dodges+m.Parries)
	}
	if misses := rogue.Shank.SpellMetrics[target.UnitIndex].Misses; misses != 0 {
		t.Errorf("Shank missed %d times", misses)
	}
	if got := rogue.Shank.CD.Duration; got != 10*time.Second {
		t.Errorf("Shank's cooldown is %s, want 10s", got)
	}
}

// Venom is a passive now: a Shank gives 30% more poison damage for 6 sec, and 20% more procs to Deadly
// and Wound Poison but not Instant Poison (server hotfix push 113483, 1310703 effect 2's class mask).
func TestForeverVenomFromShank(t *testing.T) {
	sim, rogue := newForeverRogueSim(t, "00000000000001", 18805, 18816, 1) // Venom only; Core Hound Tooth, Perdition's Blade
	if rogue.VenomAura == nil {
		t.Fatal("the talents give no Venom")
	}
	target := rogue.CurrentTarget
	base := rogue.GetDeadlyPoisonProcChance()
	baseInstant := rogue.GetInstantPoisonProcChance()
	for !rogue.VenomAura.IsActive() {
		rogue.AddEnergy(sim, 100, rogue.EnergyRefundMetrics)
		rogue.SetGCDTimer(sim, sim.CurrentTime)
		rogue.Shank.CD.Reset()
		rogue.Shank.Cast(sim, target)
	}
	if got := rogue.GetDeadlyPoisonProcChance() - base; math.Abs(got-0.2) > 1e-9 {
		t.Errorf("Venom adds %.2f to the poison chance, want 0.20", got)
	}
	if got := rogue.GetInstantPoisonProcChance() - baseInstant; got != 0 {
		t.Errorf("Venom adds %.2f to Instant Poison's chance, want 0", got)
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

// Mutilate's halves strike for 90% of (weapon damage + 88) at level 60 (server hotfix push 113483; was
// 75% of weapon damage + 67): against a
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
	if got < 0.9*(88+low)-1e-6 || got > 0.9*(88+high)+1e-6 {
		t.Errorf("a Mutilate main hand half hit for %.2f, want between %.2f and %.2f", got, 0.9*(88+low), 0.9*(88+high))
	}
}

// Flawless Execution takes 15 off Eviscerate; Aggression 3/3 raises Sinister Strike and Eviscerate
// by 15% and no longer Backstab.
func TestForeverFlawlessExecutionAndAggression(t *testing.T) {
	sim, rogue := newForeverRogueSim(t, "-00000000000103", 18805, 18805, 1)
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

// Death Mark (client 1325758 at 60): 12% of the rogue's damage to the target a combo point is
// stored for 6 sec, Rupture and Garrote twice, then dealt up to 156 + 245 a point + 4% attack power
// a point. Coup de Grace 3/3 raises the rate and the damage by 20%.
func TestForeverDeathMark(t *testing.T) {
	// Death Mark and Coup de Grace 3/3, nothing else.
	sim, rogue := newForeverRogueSim(t, "-3-000000000000000001", 18805, 18805, 1)
	if rogue.DeathMark == nil {
		t.Fatal("the talents give no Death Mark")
	}
	target := rogue.CurrentTarget
	for sim.CurrentTime <= 0 {
		sim.Step()
	}
	rogue.AutoAttacks.CancelAutoSwing(sim) // only the Garrote below is stored
	rogue.AddStatDynamic(sim, stats.MeleeHit, 20*core.MeleeHitRatingPerHitChance)
	detonation := rogue.GetSpell(core.ActionID{SpellID: 1325758, Tag: 1})
	detonation.BonusCritRating -= 1000 * core.CritRatingPerCritChance
	for _, points := range []int32{1, 5} {
		for !rogue.DeathMarkAuras.Get(target).IsActive() {
			rogue.AddComboPoints(sim, points-rogue.ComboPoints(), target, rogue.DeathMark.ComboPointMetrics())
			rogue.AddEnergy(sim, 100, rogue.EnergyRefundMetrics)
			rogue.SetGCDTimer(sim, sim.CurrentTime)
			rogue.DeathMark.CD.Reset()
			rogue.DeathMark.Cast(sim, target)
		}
		ap := rogue.DeathMark.MeleeAttackPower(target)
		damageCap := 156 + 245*float64(points) + 0.04*float64(points)*ap
		// 1000 damage of Garrote is 2 x 12% x points x 1.2 stored.
		rogue.Garrote.CalcAndDealPeriodicDamage(sim, target, 1000, rogue.Garrote.OutcomeAlwaysHitNoHitCounter)
		dealt := rogue.Garrote.SpellMetrics[target.UnitIndex].TotalDamage
		rogue.Garrote.SpellMetrics[target.UnitIndex].TotalDamage = 0
		before := detonation.SpellMetrics[target.UnitIndex].TotalDamage
		for rogue.DeathMarkAuras.Get(target).IsActive() {
			sim.Step()
		}
		got := detonation.SpellMetrics[target.UnitIndex].TotalDamage - before
		want := min(dealt*2*0.12*float64(points)*1.2, damageCap) * 1.2
		if math.Abs(got-want) > 1 { // damage is whole numbers
			t.Errorf("a %d point Death Mark dealt %.2f, want %.2f (stored from %.1f, cap %.1f)", points, got, want, dealt, damageCap)
		}
	}
}

// Boot Knife 5/5 makes every Shank add 2 combo points; Flash of Steel 5/5 adds 15 Energy and takes
// 2 sec off the cooldown; Composure 2/2 makes Poised take 20 off.
func TestForeverShankTalentsAndComposure(t *testing.T) {
	sim, rogue := newForeverRogueSim(t, "-000000000200000005-00000000000000005", 18805, 18805, 1)
	target := rogue.CurrentTarget
	if got := rogue.SinisterStrike.Cost.GetCurrentCost(); got != 25 {
		t.Errorf("Sinister Strike under Poised with Composure 2/2 costs %.0f, want 25", got)
	}
	rogue.PoisedAura.Deactivate(sim)
	if got := rogue.Shank.CD.Duration; got != 8*time.Second {
		t.Errorf("Shank's cooldown is %s, want 8s", got)
	}
	for rogue.ComboPoints() == 0 {
		rogue.SpendEnergy(sim, rogue.CurrentEnergy(), rogue.EnergyRefundMetrics)
		rogue.AddEnergy(sim, 50, rogue.EnergyRefundMetrics)
		rogue.SetGCDTimer(sim, sim.CurrentTime)
		rogue.Shank.CD.Reset()
		rogue.Shank.Cast(sim, target)
	}
	if rogue.ComboPoints() != 2 || rogue.CurrentEnergy() != 85 {
		t.Errorf("a landed Shank left %d combo points and %.0f Energy, want 2 and 85", rogue.ComboPoints(), rogue.CurrentEnergy())
	}
}

// Blood Splatter: a 5 point Rupture has 3 more ticks (16 + 6 sec), Garrote 3 more (18 + 9 sec).
func TestForeverBloodSplatter(t *testing.T) {
	sim, rogue := newForeverRogueSim(t, "00000001", 18805, 18805, 1)
	if got := rogue.RuptureTicks(5); got != 11 {
		t.Errorf("a 5 point Rupture has %d ticks, want 11", got)
	}
	dot := rogue.Garrote.Dot(rogue.CurrentTarget)
	dot.Apply(sim)
	if got := dot.RemainingDuration(sim); got != 27*time.Second {
		t.Errorf("Garrote lasts %s, want 27s", got)
	}
}

// The re-ranked talents: Lethality 10% a rank, Dual Wield Specialization 25%, Opportunity 4%,
// Relentless Strikes 4% a combo point a rank.
func TestForeverReRankedTalents(t *testing.T) {
	// Malice 5, Lethality 2; Precision 3, Dual Wield Specialization 2; Opportunity 5.
	_, rogue := newForeverRogueSim(t, "00500000200000000-00000300000000200-005", 18805, 18805, 1)
	if got := rogue.SinisterStrike.CritDamageBonus; math.Abs(got-1.2) > 1e-9 {
		t.Errorf("Lethality 2/2 gives a crit damage bonus of x%.2f, want x1.20", got)
	}
	if got := rogue.AutoAttacks.OHConfig().DamageMultiplier; math.Abs(got-1.5) > 1e-9 {
		t.Errorf("Dual Wield Specialization 2/2 makes the off hand x%.2f, want x1.5", got)
	}
	if got := rogue.Backstab.DamageMultiplier; math.Abs(got-1.5*1.2) > 1e-9 {
		t.Errorf("Backstab with Opportunity 5/5 is x%.3f, want x1.8", got)
	}
}
