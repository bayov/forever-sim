package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
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

// TestOrcShamanLightningShieldOrbs checks when the orbs fire, whom they hit, and how long the
// shield lasts (shaman_audit.md 7.15).
//
// An enemy that isn't our target hits us every 0.5 sec. The orbs hit that enemy, at most one
// every 3.5 sec, so a fresh shield fires at 0, 3.5 and 7 sec and goes away with its third orb.
// A missed orb uses up its orb too. A shield we cast inside the wait still waits, so the
// recast at 8 sec fires its first orb at 10.5 sec. On the beta (all four logs) the orbs never
// came less than 3.5 sec apart, each shield fired 3, and about 50 enemy hits on a shield cast
// inside the wait fired nothing.
//
// The enemy is level 63, which our orbs miss 17% of the time. We repeat this 16 times so that
// some orbs miss. Then a shield that takes no hits lasts 10 min.
func TestOrcShamanLightningShieldOrbs(t *testing.T) {
	const cycles, cycle = 16, 15.0
	target := core.NewDefaultTarget()
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(newOrcShaman(60, "", &proto.EnhancementShaman_Options{}), &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 900, Targets: []*proto.Target{target, target}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
	shield := topRank(t, "Lightning Shield", enh.LightningShield)
	aura := enh.LightningShieldAuras[shield.Rank]
	orb := enh.LightningShieldProcs[shield.Rank]
	enemy := sim.Encounter.TargetUnits[1]
	melee := enemy.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 1},
		SpellSchool:      core.SpellSchoolPhysical,
		DefenseType:      core.DefenseTypeMelee,
		ProcMask:         core.ProcMaskMeleeMHAuto,
		DamageMultiplier: 1,
	})
	cast := func(sim *core.Simulation) {
		enh.AddMana(sim, enh.MaxMana()-enh.CurrentMana(), enh.GetManaNotCastingMetrics())
		castNow(t, sim, enh, shield)
	}
	fired := func() int32 {
		m := orb.SpellMetrics[enemy.UnitIndex]
		return m.Hits + m.Misses
	}

	var orbs []time.Duration
	for i := range cycles {
		start := 1 + cycle*float64(i)
		at(sim, start, cast)
		at(sim, start+8, func(sim *core.Simulation) {
			if aura.IsActive() {
				t.Errorf("at %s: the shield is still up after its third orb", sim.CurrentTime)
			}
			cast(sim)
		})
		for k := range 23 {
			at(sim, start+0.5*float64(k), func(sim *core.Simulation) {
				before := fired()
				melee.CalcAndDealDamage(sim, &enh.Unit, 1, forcedOutcome(core.OutcomeHit))
				if fired() > before {
					orbs = append(orbs, sim.CurrentTime-core.DurationFromSeconds(start))
				}
			})
		}
	}
	tenMinutes := 1 + cycle*cycles
	at(sim, tenMinutes, cast)
	at(sim, tenMinutes+599.9, func(sim *core.Simulation) {
		if !aura.IsActive() {
			t.Errorf("the shield is gone 599.9 sec after we cast it")
		}
	})
	at(sim, tenMinutes+600.1, func(sim *core.Simulation) {
		if aura.IsActive() {
			t.Errorf("the shield is still up 600.1 sec after we cast it")
		}
	})
	runSim(sim)

	want := []time.Duration{0, 3500 * time.Millisecond, 7 * time.Second, 10500 * time.Millisecond}
	if len(orbs) != cycles*len(want) {
		t.Fatalf("got %d orbs at %v, want %d", len(orbs), orbs, cycles*len(want))
	}
	for i, got := range orbs {
		if got != want[i%len(want)] {
			t.Errorf("cycle %d: orb %d at %s, want %s", i/len(want)+1, i%len(want)+1, got, want[i%len(want)])
		}
	}
	if m := orb.SpellMetrics[enemy.UnitIndex]; m.Misses == 0 {
		t.Errorf("none of %d orbs missed", len(orbs))
	}
	if m := orb.SpellMetrics[enh.CurrentTarget.UnitIndex]; m.Hits+m.Misses > 0 {
		t.Errorf("%d orbs hit our target, which never hit us", m.Hits+m.Misses)
	}
}
