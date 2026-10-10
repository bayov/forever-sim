package enhancement

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanShieldTriggers checks what makes Lightning Shield fire an orb and Water Shield
// spend a globe.
//
// Both shields go off on any direct hit an enemy lands on us, whether it's a melee, a ranged
// attack or a spell. A miss, a damage over time tick or our own damage doesn't set them off.
// We roll each hit once on a freshly cast shield, 4 sec apart so that the 3.5 sec cooldown is
// always ready.
func TestOrcShamanShieldTriggers(t *testing.T) {
	// Water Shield.
	sim, enh := newShamanSim(newOrcShaman(60, "--000000001", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 60)
	boss := enh.CurrentTarget
	enemySpell := func(id int32, procMask core.ProcMask, defense core.DefenseType) *core.Spell {
		return boss.RegisterSpell(core.SpellConfig{
			ActionID:         core.ActionID{SpellID: id},
			SpellSchool:      core.SpellSchoolPhysical,
			DefenseType:      defense,
			ProcMask:         procMask,
			DamageMultiplier: 1,
		})
	}
	melee := enemySpell(1, core.ProcMaskMeleeMHAuto, core.DefenseTypeMelee)
	ranged := enemySpell(2, core.ProcMaskRangedAuto, core.DefenseTypeRanged)
	harmful := enemySpell(3, core.ProcMaskSpellDamage, core.DefenseTypeMagic)
	own := enh.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 4},
		SpellSchool:      core.SpellSchoolFire,
		DefenseType:      core.DefenseTypeMagic,
		ProcMask:         core.ProcMaskSpellDamage,
		DamageMultiplier: 1,
	})

	hits := []struct {
		name  string
		roll  func(sim *core.Simulation)
		fires bool
	}{
		{"melee hit", func(sim *core.Simulation) { melee.CalcAndDealDamage(sim, &enh.Unit, 1, forcedOutcome(core.OutcomeHit)) }, true},
		{"melee miss", func(sim *core.Simulation) {
			melee.CalcAndDealDamage(sim, &enh.Unit, 1, forcedOutcome(core.OutcomeMiss))
		}, false},
		{"ranged hit", func(sim *core.Simulation) {
			ranged.CalcAndDealDamage(sim, &enh.Unit, 1, forcedOutcome(core.OutcomeHit))
		}, true},
		{"spell hit", func(sim *core.Simulation) {
			harmful.CalcAndDealDamage(sim, &enh.Unit, 1, forcedOutcome(core.OutcomeHit))
		}, true},
		{"damage over time tick", func(sim *core.Simulation) {
			// We can't add a damage over time spell to a running sim, so we send a spell hit
			// down the tick path.
			harmful.DealPeriodicDamage(sim, harmful.CalcDamage(sim, &enh.Unit, 1, forcedOutcome(core.OutcomeHit)))
		}, false},
		{"our own spell", func(sim *core.Simulation) { own.CalcAndDealDamage(sim, &enh.Unit, 1, forcedOutcome(core.OutcomeHit)) }, false},
	}

	shields := []struct {
		spell *core.Spell
		aura  func() *core.Aura
	}{
		{topRank(t, "Lightning Shield", enh.LightningShield), func() *core.Aura { return enh.ActiveShieldAura }},
		{enh.WaterShield, func() *core.Aura { return enh.WaterShieldAura }},
	}
	when := 0.0
	for _, shield := range shields {
		for _, hit := range hits {
			when += 4
			at(sim, when, func(sim *core.Simulation) {
				castNow(t, sim, enh, shield.spell)
				hit.roll(sim)
				want := int32(3)
				if hit.fires {
					want = 2
				}
				if got := shield.aura().GetStacks(); got != want {
					t.Errorf("%s with %s: %d charges left, want %d", hit.name, shield.spell.ActionID, got, want)
				}
			})
		}
	}
	runSim(sim)
}

// TestOrcShamanRaidDamageLightningShield checks that the raid damage we take fires Lightning
// Shield's orbs at our target, the same way it spends Water Shield's globes.
//
// A hit every 5 sec (5, 10, 15) fires an orb each time, and the third orb uses up the shield.
func TestOrcShamanRaidDamageLightningShield(t *testing.T) {
	sim, enh := newShamanSim(newOrcShaman(60, "", &proto.EnhancementShaman_Options{RaidDamageHitsPerMinute: 12}), &proto.Debuffs{}, 30)
	shield := topRank(t, "Lightning Shield", enh.LightningShield)
	orb := topRank(t, "Lightning Shield", enh.LightningShieldProcs)
	at(sim, 0.5, func(sim *core.Simulation) { castNow(t, sim, enh, shield) })
	for _, check := range []struct {
		at      float64
		charges int32
	}{{4, 3}, {6, 2}, {11, 1}, {16, 0}} {
		at(sim, check.at, func(sim *core.Simulation) {
			charges := int32(0)
			if aura := enh.ActiveShieldAura; aura != nil && aura.IsActive() {
				charges = aura.GetStacks()
			}
			if charges != check.charges {
				t.Errorf("at %s: %d charges left, want %d", sim.CurrentTime, charges, check.charges)
			}
		})
	}
	runSim(sim)
	if hits := orb.SpellMetrics[enh.CurrentTarget.UnitIndex].Hits; hits != 3 {
		t.Errorf("%d orbs hit our target, want 3", hits)
	}
}
