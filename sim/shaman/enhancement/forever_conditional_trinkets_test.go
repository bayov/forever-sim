package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// withTargets sets the family and level of every target. A level 60 target has no resistance
// from its level, so magic hits on it deal their exact damage.
func withTargets(family proto.MobFamily, level int32) func(*proto.RaidSimRequest) {
	return func(request *proto.RaidSimRequest) {
		for _, target := range request.Encounter.Targets {
			target.MobFamily = family
			target.Level = level
		}
	}
}

// autocastRotation makes the shaman's rotation use every major cooldown when it's ready.
func autocastRotation(request *proto.RaidSimRequest) {
	request.Raid.Parties[0].Players[0].Rotation = &proto.APLRotation{
		PriorityList: []*proto.APLListItem{{Action: &proto.APLAction{
			Action: &proto.APLAction_AutocastOtherCooldowns{AutocastOtherCooldowns: &proto.APLActionAutocastOtherCooldowns{}},
		}}},
	}
}

// casts is how often spell was cast on any unit.
func casts(spell *core.Spell) int32 {
	total := int32(0)
	for _, m := range spell.SpellMetrics() {
		total += m.Casts
	}
	return total
}

// Blackhand's Breadth's mark raises our crit chance on the marked target by 5% for melee and
// spells alike, and by 10% on Orcs and Ogres. It rolls spell hit, so some uses miss.
func TestForeverBlackhandsBreadth(t *testing.T) {
	for _, tc := range []struct {
		family proto.MobFamily
		bonus  float64
	}{
		{proto.MobFamily_MobFamilyUnknown, 0.05},
		{proto.MobFamily_MobFamilyTroll, 0.05},
		{proto.MobFamily_MobFamilyOrc, 0.10},
		{proto.MobFamily_MobFamilyOgre, 0.10},
	} {
		sim, enh := newTrinketSimWith(t, withTargets(tc.family, 63), 1, 1000000, common.BlackhandsBreadth, 274760)
		use := enh.GetSpell(core.ActionID{ItemID: common.BlackhandsBreadth})
		if use == nil {
			t.Fatal("no Blackhand's Breadth use")
		}
		if enh.GetMajorCooldown(use.ActionID) == nil {
			t.Fatal("Blackhand's Breadth is not a major cooldown, so the rotation won't use it")
		}
		target := enh.CurrentTarget
		mark := target.GetAuraByID(core.ActionID{SpellID: 1318944})
		if mark == nil {
			t.Fatal("no Blackhand's Breadth mark on the target")
		}
		attackTable := enh.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand]
		lightningBolt := topRank(t, "Lightning Bolt", enh.LightningBolt)
		meleeBefore := enh.Stormstrike.PhysicalCritChance(attackTable)
		spellBefore := lightningBolt.SpellCritChance(target)

		if !use.Cast(sim, target) {
			t.Fatal("Blackhand's Breadth didn't cast")
		}
		if got := use.CD.TimeToReady(sim); got != 5*time.Minute {
			t.Errorf("cooldown %s after the use, want 5m", got)
		}
		deliveryBot := enh.GetSpell(core.ActionID{ItemID: 274760})
		if deliveryBot.CanCast(sim, target) {
			t.Error("Everlook Delivery Bot is ready right after Blackhand's Breadth")
		}

		// Use it until the mark lands, and count the misses on the way. About 1000 uses
		// give a few dozen misses at a boss's spell hit table.
		uses, misses := 1, 0
		if !mark.IsActive() {
			misses++
		}
		for !mark.IsActive() || uses < 1000 {
			sim.CurrentTime += 5 * time.Minute
			mark.Deactivate(sim)
			use.Cast(sim, target)
			uses++
			if !mark.IsActive() {
				misses++
			}
		}
		if misses == 0 || misses > uses/4 {
			t.Errorf("%s: %d of %d marks missed, want some but not many", tc.family, misses, uses)
		}

		if got := enh.Stormstrike.PhysicalCritChance(attackTable) - meleeBefore; math.Abs(got-tc.bonus) > 1e-9 {
			t.Errorf("%s: marked target raises Stormstrike's crit chance by %.4f, want %.2f", tc.family, got, tc.bonus)
		}
		if got := lightningBolt.SpellCritChance(target) - spellBefore; math.Abs(got-tc.bonus) > 1e-9 {
			t.Errorf("%s: marked target raises Lightning Bolt's crit chance by %.4f, want %.2f", tc.family, got, tc.bonus)
		}
		if mark.Duration != 20*time.Second {
			t.Errorf("mark lasts %s, want 20s", mark.Duration)
		}

		mark.Deactivate(sim)
		if got := enh.Stormstrike.PhysicalCritChance(attackTable) - meleeBefore; math.Abs(got) > 1e-9 {
			t.Errorf("%s: crit chance still %.4f higher after the mark is gone", tc.family, got)
		}
		if got := lightningBolt.SpellCritChance(target) - spellBefore; math.Abs(got) > 1e-9 {
			t.Errorf("%s: spell crit chance still %.4f higher after the mark is gone", tc.family, got)
		}
	}

	// Classic's Blackhand's Breadth has no use.
	_, enh := newTrinketSimWith(t, func(request *proto.RaidSimRequest) {
		request.SimOptions.Ruleset = proto.Ruleset_RulesetClassic
	}, 1, 10, common.BlackhandsBreadth)
	if enh.GetSpell(core.ActionID{ItemID: common.BlackhandsBreadth}) != nil {
		t.Error("Blackhand's Breadth has a use in Classic")
	}
}

// Bolt's Charged Eyeball's shock deals 60 Nature (90 on a crit), and 3 times as much to a
// Dinosaur.
func TestForeverBoltsChargedEyeballDinosaur(t *testing.T) {
	for _, tc := range []struct {
		family proto.MobFamily
		damage float64
	}{
		{proto.MobFamily_MobFamilyUnknown, 60},
		{proto.MobFamily_MobFamilyDinosaur, 180},
	} {
		sim, enh := newTrinketSimWith(t, withTargets(tc.family, 60), 1, 100000, common.BoltsChargedEyeball)
		shock := enh.GetSpell(core.ActionID{SpellID: 1315996})
		target := enh.CurrentTarget
		for range 1000 {
			shock.Cast(sim, target)
		}
		m := shock.SpellMetrics()[target.UnitIndex]
		if m.Hits == 0 || m.Crits == 0 || m.ResistedHits+m.ResistedCrits > 0 {
			t.Fatalf("%s: %d hits, %d crits and %d resisted, want hits and crits and no resists",
				tc.family, m.Hits, m.Crits, m.ResistedHits+m.ResistedCrits)
		}
		want := tc.damage*float64(m.Hits) + tc.damage*1.5*float64(m.Crits)
		if math.Abs(m.TotalDamage-want) > 1e-6 {
			t.Errorf("%s: %d hits and %d crits dealt %.1f, want %.1f (%.0f a hit)", tc.family, m.Hits, m.Crits, m.TotalDamage, want, tc.damage)
		}
	}
}

// Chained Essence of Eranikus' cloud ticks 140 Nature (210 on a crit), and twice as much on a
// Troll.
func TestForeverChainedEssenceOfEranikusTroll(t *testing.T) {
	for _, tc := range []struct {
		family proto.MobFamily
		damage float64
	}{
		{proto.MobFamily_MobFamilyOrc, 140},
		{proto.MobFamily_MobFamilyTroll, 280},
	} {
		totalTicks, totalCrits := int32(0), int32(0)
		for seed := range 20 {
			sim, enh := newTrinketSimWith(t, func(request *proto.RaidSimRequest) {
				withTargets(tc.family, 60)(request)
				request.SimOptions.RandomSeed = int64(seed + 1)
			}, 1, 30, common.ChainedEssenceOfEranikus)
			use := enh.GetSpell(core.ActionID{ItemID: common.ChainedEssenceOfEranikus})
			cloud := enh.GetSpell(core.ActionID{SpellID: 1300240})
			at(sim, 1, func(sim *core.Simulation) { use.Cast(sim, enh.CurrentTarget) })
			runSim(sim)

			m := cloud.SpellMetrics()[enh.CurrentTarget.UnitIndex]
			if m.ResistedTicks+m.ResistedCritTicks > 0 {
				t.Fatalf("%s: %d resisted ticks, want none", tc.family, m.ResistedTicks+m.ResistedCritTicks)
			}
			want := tc.damage*float64(m.Ticks) + tc.damage*1.5*float64(m.CritTicks)
			if math.Abs(m.TotalTickDamage-want) > 1e-6 {
				t.Errorf("%s: %d ticks and %d crits dealt %.1f, want %.1f (%.0f a tick)", tc.family, m.Ticks, m.CritTicks, m.TotalTickDamage, want, tc.damage)
			}
			totalTicks += m.Ticks + m.CritTicks
			totalCrits += m.CritTicks
		}
		if totalTicks == 0 || totalCrits == totalTicks {
			t.Errorf("%s: %d ticks with %d crits, want normal ticks", tc.family, totalTicks, totalCrits)
		}
	}
}

// Golden Banana gives 16 to all stats for 15 sec, and 32 for 30 sec in a forest or grassland.
// Frozen Heart of the Mountain's 29 Frost and Shadow damage doubles in snow. Either way the
// other burst trinkets wait 15 sec.
func TestForeverTerrainTrinkets(t *testing.T) {
	for _, tc := range []struct {
		terrain  proto.OutdoorTerrain
		banana   float64
		duration time.Duration
		heart    float64
	}{
		{proto.OutdoorTerrain_OutdoorTerrainNone, 16, 15 * time.Second, 29},
		{proto.OutdoorTerrain_OutdoorTerrainForest, 32, 30 * time.Second, 29},
		{proto.OutdoorTerrain_OutdoorTerrainGrassland, 32, 30 * time.Second, 29},
		{proto.OutdoorTerrain_OutdoorTerrainSnow, 16, 15 * time.Second, 58},
	} {
		sim, enh := newTrinketSimWith(t, func(request *proto.RaidSimRequest) {
			request.Encounter.OutdoorTerrain = tc.terrain
		}, 1, 100, common.GoldenBanana, common.FrozenHeartOfTheMountain)
		banana := enh.GetSpell(core.ActionID{ItemID: common.GoldenBanana})
		heart := enh.GetSpell(core.ActionID{ItemID: common.FrozenHeartOfTheMountain})
		if enh.GetMajorCooldown(banana.ActionID) == nil || enh.GetMajorCooldown(heart.ActionID) == nil {
			t.Fatal("Golden Banana or Frozen Heart of the Mountain is not a major cooldown")
		}

		agility := enh.GetStat(stats.Agility)
		if !banana.Cast(sim, enh.CurrentTarget) {
			t.Fatal("Golden Banana didn't cast")
		}
		buff := enh.GetAura("ItemActive-270226")
		if buff.Duration != tc.duration {
			t.Errorf("%s: banana lasts %s, want %s", tc.terrain, buff.Duration, tc.duration)
		}
		// Kings or racials don't touch a naked shaman's Agility here, so it adds as is.
		if got := enh.GetStat(stats.Agility) - agility; math.Abs(got-tc.banana) > 1e-9 {
			t.Errorf("%s: banana gives %.1f Agility, want %.0f", tc.terrain, got, tc.banana)
		}
		if heart.CanCast(sim, enh.CurrentTarget) {
			t.Errorf("%s: Frozen Heart ready right after Golden Banana", tc.terrain)
		}
		sim.CurrentTime += 15 * time.Second
		if !heart.CanCast(sim, enh.CurrentTarget) {
			t.Errorf("%s: Frozen Heart still not ready 15 sec after Golden Banana", tc.terrain)
		}

		frost := enh.GetStat(stats.FrostPower)
		heart.Cast(sim, enh.CurrentTarget)
		if got := enh.GetStat(stats.FrostPower) - frost; math.Abs(got-tc.heart) > 1e-9 {
			t.Errorf("%s: Frozen Heart gives %.1f Frost damage, want %.0f", tc.terrain, got, tc.heart)
		}
	}
}

// Each Forever use trinket we model is a major cooldown, so autocastOtherCooldowns uses it in a
// fight. Adaptive Combat Assistant waits for a fight where something hits us.
func TestForeverUseTrinketsAutocast(t *testing.T) {
	for _, id := range []int32{common.BlackhandsBreadth, common.EyeOfTheBeast, common.WeaknessAnalyzer, common.GoldenBanana} {
		sim, enh := newTrinketSimWith(t, autocastRotation, 1, 30, id)
		use := enh.GetSpell(core.ActionID{ItemID: id})
		if use == nil || enh.GetMajorCooldown(use.ActionID) == nil {
			t.Fatalf("item %d has no use or it isn't a major cooldown", id)
		}
		runSim(sim)
		if casts(use) != 1 {
			t.Errorf("item %d used %d times in 30 sec, want 1", id, casts(use))
		}
	}

	sim, enh := newTrinketSimWith(t, autocastRotation, 1, 30, common.AdaptiveCombatAssistant)
	use := enh.GetSpell(core.ActionID{ItemID: common.AdaptiveCombatAssistant})
	runSim(sim)
	if casts(use) != 0 {
		t.Errorf("Adaptive Combat Assistant used %d times when nothing hits us, want 0", casts(use))
	}
}

// Eye of the Beast's use gives 7% spell hit for 20 sec. Weakness Analyzer's gives 5% crit with
// everything until our first crit that isn't a tick.
func TestForeverEyeOfTheBeastAndWeaknessAnalyzer(t *testing.T) {
	sim, enh := newTrinketSim(t, 1, 100, common.EyeOfTheBeast, common.WeaknessAnalyzer)
	eye := enh.GetSpell(core.ActionID{ItemID: common.EyeOfTheBeast})
	analyzer := enh.GetSpell(core.ActionID{ItemID: common.WeaknessAnalyzer})

	spellHit := enh.GetStat(stats.SpellHit)
	eye.Cast(sim, enh.CurrentTarget)
	if got := enh.GetStat(stats.SpellHit) - spellHit; math.Abs(got-7) > 1e-9 {
		t.Errorf("Eye of the Beast gives %.1f spell hit, want 7", got)
	}
	if got := enh.GetAura("ItemActive-13968").Duration; got != 20*time.Second {
		t.Errorf("Eye of the Beast lasts %s, want 20s", got)
	}
	if analyzer.CanCast(sim, enh.CurrentTarget) {
		t.Error("Weakness Analyzer ready right after Eye of the Beast")
	}

	sim.CurrentTime += 20 * time.Second
	meleeCrit, spellCrit := enh.GetStat(stats.MeleeCrit), enh.GetStat(stats.SpellCrit)
	if !analyzer.Cast(sim, enh.CurrentTarget) {
		t.Fatal("Weakness Analyzer didn't cast 20 sec after Eye of the Beast")
	}
	if got := analyzer.CD.TimeToReady(sim); got != 6*time.Minute {
		t.Errorf("Weakness Analyzer cooldown %s, want 6m", got)
	}
	buff := enh.GetAuraByID(core.ActionID{SpellID: 1291101})
	if math.Abs(enh.GetStat(stats.MeleeCrit)-meleeCrit-5) > 1e-9 || math.Abs(enh.GetStat(stats.SpellCrit)-spellCrit-5) > 1e-9 {
		t.Errorf("Weakness Analyzer gives %.1f melee and %.1f spell crit, want 5 each",
			enh.GetStat(stats.MeleeCrit)-meleeCrit, enh.GetStat(stats.SpellCrit)-spellCrit)
	}

	// A hit and a crit tick leave it up. The first direct crit takes it off.
	target := enh.CurrentTarget
	flameShock := topRank(t, "Flame Shock", enh.FlameShock)
	deal := func(spell *core.Spell, outcome core.HitOutcome, periodic bool) {
		result := spell.NewResult(target)
		result.Outcome = outcome
		result.Damage = 100
		if periodic {
			spell.DealPeriodicDamage(sim, result)
		} else {
			spell.DealDamage(sim, result)
		}
	}
	deal(enh.Stormstrike, core.OutcomeHit, false)
	deal(flameShock, core.OutcomeCrit, true)
	if !buff.IsActive() {
		t.Fatal("Analyzing Weaknesses went on a hit or a crit tick")
	}
	deal(enh.Stormstrike, core.OutcomeCrit, false)
	if buff.IsActive() {
		t.Error("Analyzing Weaknesses still up after a crit")
	}
	if math.Abs(enh.GetStat(stats.MeleeCrit)-meleeCrit) > 1e-9 {
		t.Error("crit not back after Analyzing Weaknesses went")
	}
}

// Adaptive Combat Assistant's shield absorbs 450. When hits break it before its 15 sec are
// up, it deals 100 Nature (150 on a crit) to every enemy. A shield that lasts its 15 sec does
// nothing.
func TestForeverAdaptiveCombatAssistant(t *testing.T) {
	for _, tc := range []struct {
		hitsPerMinute float64
		breaks        bool
	}{
		{60, true},
		{0, false},
	} {
		sim, enh := newTrinketSimWith(t, func(request *proto.RaidSimRequest) {
			autocastRotation(request)
			withTargets(proto.MobFamily_MobFamilyUnknown, 60)(request)
			request.Encounter.RaidDamageHitsPerMinute = tc.hitsPerMinute
			request.Encounter.RaidDamagePerHit = 200
		}, 3, 14, common.AdaptiveCombatAssistant)
		use := enh.GetSpell(core.ActionID{ItemID: common.AdaptiveCombatAssistant})
		revenge := enh.GetSpell(core.ActionID{SpellID: 1291099})
		if tc.hitsPerMinute == 0 {
			// Nothing hits us, so autocast leaves it alone, and we use it by hand.
			at(sim, 1, func(sim *core.Simulation) { use.Cast(sim, enh.CurrentTarget) })
		}
		runSim(sim)
		if casts(use) != 1 {
			t.Fatalf("%.0f hits a minute: used %d times, want 1", tc.hitsPerMinute, casts(use))
		}

		hitTargets := 0
		for _, target := range sim.Encounter.TargetUnits {
			m := revenge.SpellMetrics()[target.UnitIndex]
			if m.Hits+m.Crits+m.Misses == 0 {
				continue
			}
			hitTargets++
			if want := 100*float64(m.Hits) + 150*float64(m.Crits); math.Abs(m.TotalDamage-want) > 1e-6 {
				t.Errorf("%s took %.1f from Sigmoid Revenge, want %.1f", target.Label, m.TotalDamage, want)
			}
		}
		if tc.breaks && hitTargets != 3 {
			t.Errorf("the broken shield hit %d of 3 targets", hitTargets)
		}
		if !tc.breaks && hitTargets != 0 {
			t.Errorf("a shield nothing broke hit %d targets", hitTargets)
		}
	}
}
