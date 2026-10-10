package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
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

// TestOrcShamanRaidDamageHits checks that the raid damage we take fires Lightning Shield's
// orbs at our target, and that the hits land at random times.
//
// We take 15 hits a minute. A hit every 4 sec would fire an orb each time, 15 orbs a minute.
// Random hits at the same rate often come during the 3.5 sec wait after an orb. Then the next
// orb comes after the wait plus the wait for the next hit (4 sec on average), so 60 / 7.5 = 8
// orbs a minute. We keep the shield at 3 orbs and count them over 9 min (the shield lasts 10).
func TestOrcShamanRaidDamageHits(t *testing.T) {
	const minutes, hitsPerMinute = 9, 15
	player := newOrcShaman(60, "", &proto.EnhancementShaman_Options{RaidDamageHitsPerMinute: hitsPerMinute})
	sim, enh := newShamanSim(player, &proto.Debuffs{}, minutes*60)
	shield := topRank(t, "Lightning Shield", enh.LightningShield)
	orb := topRank(t, "Lightning Shield", enh.LightningShieldProcs)
	at(sim, 0.5, func(sim *core.Simulation) {
		castNow(t, sim, enh, shield)
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period:   time.Second,
			OnAction: func(sim *core.Simulation) { enh.ActiveShieldAura.SetStacks(sim, 3) },
		})
	})
	runSim(sim)
	// An orb can miss (TestOrcShamanLightningShieldOrbOutcome), so we count misses too.
	metrics := orb.SpellMetrics[enh.CurrentTarget.UnitIndex]
	fired := metrics.Hits + metrics.Misses
	if want := minutes * 60 / (3.5 + 60.0/hitsPerMinute); math.Abs(float64(fired)-want) > 15 {
		t.Errorf("%d orbs fired at our target, want about %.0f", fired, want)
	}
}

// TestOrcShamanLightningShieldOrbOutcome checks that under Forever an orb misses as often as
// our spells do and never crits.
//
// A level 30 against a level 33 misses spells 17% of the time (TestOrcShamanSpellMissChance).
// We fire 3000 orbs from a full shield with 100% spell crit, so a crit would show at once.
func TestOrcShamanLightningShieldOrbOutcome(t *testing.T) {
	sim, enh := newTargetLevelSim(30, 33, whirlwind, 0, "")
	shield := topRank(t, "Lightning Shield", enh.LightningShield)
	orb := topRank(t, "Lightning Shield", enh.LightningShieldProcs)
	const orbs = 3000
	at(sim, 1, func(sim *core.Simulation) {
		enh.AddStatDynamic(sim, stats.SpellCrit, 100*core.SpellCritRatingPerCritChance)
		castNow(t, sim, enh, shield)
		for range orbs {
			enh.ActiveShieldAura.SetStacks(sim, 3)
			orb.Cast(sim, enh.CurrentTarget)
		}
	})
	runSim(sim)

	metrics := orb.SpellMetrics[enh.CurrentTarget.UnitIndex]
	if missed := float64(metrics.Misses) / orbs; math.Abs(missed-0.17) > 0.025 {
		t.Errorf("%.1f%% of orbs missed, want 17%%", 100*missed)
	}
	if metrics.Crits != 0 {
		t.Errorf("%d orbs crit, want none", metrics.Crits)
	}
}
