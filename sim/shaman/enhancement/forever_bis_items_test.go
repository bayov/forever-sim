package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/shaman"
)

// withItems puts items in the given slots, and then runs the other setups.
func withItems(items map[proto.ItemSlot]int32, setups ...func(*proto.RaidSimRequest)) func(*proto.RaidSimRequest) {
	return func(request *proto.RaidSimRequest) {
		for slot, id := range items {
			request.Raid.Parties[0].Players[0].Equipment.Items[slot].Id = id
		}
		for _, setup := range setups {
			setup(request)
		}
	}
}

func classicRuleset(request *proto.RaidSimRequest) {
	request.SimOptions.Ruleset = proto.Ruleset_RulesetClassic
}

func inTerrain(terrain proto.OutdoorTerrain) func(*proto.RaidSimRequest) {
	return func(request *proto.RaidSimRequest) {
		request.Encounter.OutdoorTerrain = terrain
	}
}

// Each weapon procs at its own rate under each ruleset: Gravestone War Axe 1.9 PPM and Seeping
// Willow 0.9 under Forever (0.5 in Classic), and Brain Hacker and Treant's Bane 1 under Forever
// only. We count the procs from white swings alone over 5 hours.
func TestForeverWeaponProcRates(t *testing.T) {
	for _, tc := range []struct {
		name    string
		weapon  int32
		procID  int32
		forever float64
		classic float64
	}{
		{"Gravestone War Axe", common.GravestoneWarAxe, 18289, 1.9, 0.5},
		{"Seeping Willow", common.SeepingWillow, 17196, 0.9, 0.5},
		{"Brain Hacker", common.BrainHacker, 17148, 1, 0},
		{"Treant's Bane", common.TreantsBane, 1302354, 1, 0},
	} {
		for _, ruleset := range []proto.Ruleset{proto.Ruleset_RulesetForever, proto.Ruleset_RulesetClassic} {
			ppm := tc.forever
			if ruleset == proto.Ruleset_RulesetClassic {
				ppm = tc.classic
			}
			sim, enh := newTrinketSimWith(t, withItems(map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: tc.weapon},
				withTargets(proto.MobFamily_MobFamilyUnknown, 60),
				func(request *proto.RaidSimRequest) { request.SimOptions.Ruleset = ruleset }), 1, 5*3600)
			proc := enh.GetSpell(core.ActionID{SpellID: tc.procID})
			if ppm == 0 {
				if proc != nil {
					t.Errorf("%s procs in %s", tc.name, ruleset)
				}
				continue
			}
			if proc == nil {
				t.Fatalf("%s has no proc in %s", tc.name, ruleset)
			}
			runSim(sim)
			// Against a level 60 target about 9 in 10 white swings land, and a proc that misses
			// still counts as a cast.
			want := ppm * 300 * 0.9
			if got := float64(casts(proc)); got < want*0.8 || got > want*1.2 {
				t.Errorf("%s in %s: %.0f procs in 5 hours, want about %.0f (%.1f PPM)", tc.name, ruleset, got, want, ppm)
			}
		}
	}
}

// Creeping Mold ticks 42 Nature 4 times under Forever (Classic 55, 15 times). Seeping Willow ticks
// 51 every 2 sec 5 times under Forever and no longer lowers the target's stats.
func TestForeverWeaponProcDots(t *testing.T) {
	for _, tc := range []struct {
		name       string
		weapon     int32
		procID     int32
		ruleset    proto.Ruleset
		tickDamage float64
		ticks      int32
		tickLength time.Duration
	}{
		{"Gravestone War Axe", common.GravestoneWarAxe, 18289, proto.Ruleset_RulesetForever, 42, 4, 3 * time.Second},
		{"Gravestone War Axe", common.GravestoneWarAxe, 18289, proto.Ruleset_RulesetClassic, 55, 15, 3 * time.Second},
		{"Seeping Willow", common.SeepingWillow, 17196, proto.Ruleset_RulesetForever, 51, 5, 2 * time.Second},
		{"Seeping Willow", common.SeepingWillow, 17196, proto.Ruleset_RulesetClassic, 20, 10, 3 * time.Second},
	} {
		sim, enh := newTrinketSimWith(t, withItems(map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: tc.weapon},
			withTargets(proto.MobFamily_MobFamilyUnknown, 60),
			func(request *proto.RaidSimRequest) { request.SimOptions.Ruleset = tc.ruleset }), 1, 100)
		proc := enh.GetSpell(core.ActionID{SpellID: tc.procID})
		target := enh.CurrentTarget
		dot := proc.Dot(target)
		if dot.NumberOfTicks != tc.ticks || dot.TickLength != tc.tickLength {
			t.Errorf("%s in %s: %d ticks every %s, want %d every %s", tc.name, tc.ruleset, dot.NumberOfTicks, dot.TickLength, tc.ticks, tc.tickLength)
		}

		agility := target.GetStat(stats.Agility)
		proc.Cast(sim, target)
		for !dot.IsActive() {
			proc.Cast(sim, target)
		}
		if got := dot.SnapshotBaseDamage; math.Abs(got-tc.tickDamage) > 1e-9 {
			t.Errorf("%s in %s: ticks %.1f, want %.0f", tc.name, tc.ruleset, got, tc.tickDamage)
		}
		lowered := agility - target.GetStat(stats.Agility)
		if tc.procID == 17196 && tc.ruleset == proto.Ruleset_RulesetForever && lowered != 0 {
			t.Errorf("Forever's Seeping Willow lowers Agility by %.0f", lowered)
		}
		if tc.procID == 17196 && tc.ruleset == proto.Ruleset_RulesetClassic && lowered != 20 {
			t.Errorf("Classic's Seeping Willow lowers Agility by %.0f, want 20", lowered)
		}
	}
}

// Brain Hacker's Physical hit rolls 150 to 350. Treant's Bane's Fire hit rolls 138 to 212 and
// deals 3 times as much to a Plant.
func TestForeverWeaponProcHits(t *testing.T) {
	for _, tc := range []struct {
		name     string
		weapon   int32
		procID   int32
		family   proto.MobFamily
		min, max float64
	}{
		{"Brain Hacker", common.BrainHacker, 17148, proto.MobFamily_MobFamilyUnknown, 150, 350},
		{"Treant's Bane", common.TreantsBane, 1302354, proto.MobFamily_MobFamilyUnknown, 138, 212},
		{"Treant's Bane", common.TreantsBane, 1302354, proto.MobFamily_MobFamilyPlant, 3 * 138, 3 * 212},
	} {
		sim, enh := newTrinketSimWith(t, withItems(map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: tc.weapon},
			withTargets(tc.family, 60)), 1, 100)
		proc := enh.GetSpell(core.ActionID{SpellID: tc.procID})
		target := enh.CurrentTarget
		for range 2000 {
			proc.Cast(sim, target)
		}
		m := proc.SpellMetrics()[target.UnitIndex]
		if m.Hits == 0 || m.Crits == 0 {
			t.Fatalf("%s on %s: %d hits and %d crits, want both", tc.name, tc.family, m.Hits, m.Crits)
		}
		// The crits deal 1.5 (spell) or 2 (Physical) times a roll. Take them out at the average.
		critMultiplier := 1.5
		if tc.procID == 17148 {
			critMultiplier = 2
		}
		hits := float64(m.Hits) + critMultiplier*float64(m.Crits)
		mean := m.TotalDamage / hits
		if want := (tc.min + tc.max) / 2; math.Abs(mean-want) > want*0.03 {
			t.Errorf("%s on %s: %.1f a hit on average, want about %.1f (%.0f to %.0f)", tc.name, tc.family, mean, want, tc.min, tc.max)
		}
	}
}

// Totem of Rage gives our shocks 2% more damage under Forever, and 30 flat damage in Classic.
func TestForeverTotemOfRage(t *testing.T) {
	for _, ruleset := range []proto.Ruleset{proto.Ruleset_RulesetForever, proto.Ruleset_RulesetClassic} {
		setRuleset := func(request *proto.RaidSimRequest) { request.SimOptions.Ruleset = ruleset }
		_, without := newTrinketSimWith(t, setRuleset, 1, 10)
		_, with := newTrinketSimWith(t, withItems(map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotRanged: shaman.TotemOfRage}, setRuleset), 1, 10)
		for _, spells := range []struct {
			name         string
			with, before []*core.Spell
		}{
			{"Earth Shock", with.EarthShock, without.EarthShock},
			{"Flame Shock", with.FlameShock, without.FlameShock},
			{"Frost Shock", with.FrostShock, without.FrostShock},
		} {
			after, before := topRank(t, spells.name, spells.with), topRank(t, spells.name, spells.before)
			multiplier := after.DamageMultiplierAdditive - before.DamageMultiplierAdditive
			bonus := after.BonusDamage - before.BonusDamage
			wantMultiplier, wantBonus := 0.02, 0.0
			if ruleset == proto.Ruleset_RulesetClassic {
				wantMultiplier, wantBonus = 0, 30
			}
			if math.Abs(multiplier-wantMultiplier) > 1e-9 || bonus != wantBonus {
				t.Errorf("%s: %s gets %.2f more multiplier and %.0f more damage, want %.2f and %.0f",
					ruleset, spells.name, multiplier, bonus, wantMultiplier, wantBonus)
			}
		}
		lightningBolt := topRank(t, "Lightning Bolt", with.LightningBolt).DamageMultiplierAdditive -
			topRank(t, "Lightning Bolt", without.LightningBolt).DamageMultiplierAdditive
		if lightningBolt != 0 {
			t.Errorf("%s: Totem of Rage changes Lightning Bolt", ruleset)
		}
	}
}

// Second Wind restores 63 mana a second for 10 sec under Forever, twice that in the mountains,
// and 30 in Classic.
func TestForeverSecondWind(t *testing.T) {
	for _, tc := range []struct {
		ruleset proto.Ruleset
		terrain proto.OutdoorTerrain
		perTick float64
	}{
		{proto.Ruleset_RulesetForever, proto.OutdoorTerrain_OutdoorTerrainNone, 63},
		{proto.Ruleset_RulesetForever, proto.OutdoorTerrain_OutdoorTerrainForest, 63},
		{proto.Ruleset_RulesetForever, proto.OutdoorTerrain_OutdoorTerrainMountains, 126},
		{proto.Ruleset_RulesetClassic, proto.OutdoorTerrain_OutdoorTerrainMountains, 30},
	} {
		sim, enh := newTrinketSimWith(t, func(request *proto.RaidSimRequest) {
			request.SimOptions.Ruleset = tc.ruleset
			request.Encounter.OutdoorTerrain = tc.terrain
		}, 1, 30, common.SecondWind)
		use := enh.GetSpell(core.ActionID{SpellID: 15604})
		at(sim, 1, func(sim *core.Simulation) { use.Cast(sim, enh.CurrentTarget) })
		runSim(sim)
		var events int32
		var gain float64
		for _, m := range enh.Metrics.ToProto().Resources {
			if m.Id.GetSpellId() == 15604 {
				events, gain = m.Events, m.Gain
			}
		}
		if events != 10 || math.Abs(gain-10*tc.perTick) > 1e-6 {
			t.Errorf("%s in %s: %d ticks restored %.0f mana, want 10 ticks of %.0f", tc.ruleset, tc.terrain, events, gain, tc.perTick)
		}
		if got := use.CD.Duration; got != 15*time.Minute {
			t.Errorf("%s: cooldown %s, want 15m", tc.ruleset, got)
		}
	}
}

// Smolderweb's Eye has no poison under Forever. Classic keeps it.
func TestForeverSmolderwebsEye(t *testing.T) {
	_, enh := newTrinketSim(t, 1, 10, common.SmolderwebsEye)
	if enh.GetSpell(core.ActionID{SpellID: 17330}) != nil {
		t.Error("Smolderweb's Eye poisons under Forever")
	}
	_, enh = newTrinketSimWith(t, classicRuleset, 1, 10, common.SmolderwebsEye)
	if enh.GetSpell(core.ActionID{SpellID: 17330}) == nil {
		t.Error("Smolderweb's Eye has no poison in Classic")
	}
}

// Draconic Infused Emblem has no use under Forever. 8% of our landed harmful spells give 35
// spell power for 10 sec, and 35 more spell damage against Dragonkin. Classic keeps its use.
func TestForeverDraconicInfusedEmblem(t *testing.T) {
	sim, enh := newTrinketSimWith(t, func(request *proto.RaidSimRequest) {
		request.Encounter.Targets[0].MobType = proto.MobType_MobTypeDragonkin
		request.Encounter.Targets[1].MobType = proto.MobType_MobTypeHumanoid
	}, 2, 1000, common.DraconicInfusedEmblem)
	if enh.GetSpell(core.ActionID{ItemID: common.DraconicInfusedEmblem}) != nil {
		t.Error("Draconic Infused Emblem has a use under Forever")
	}
	buff := enh.GetAuraByID(core.ActionID{SpellID: 1318930})
	if buff == nil {
		t.Fatal("no Draconic Infused Emblem buff")
	}
	if buff.Duration != 10*time.Second {
		t.Errorf("buff lasts %s, want 10s", buff.Duration)
	}
	dragon, humanoid := sim.Encounter.TargetUnits[0], sim.Encounter.TargetUnits[1]
	earthShock := topRank(t, "Earth Shock", enh.EarthShock)
	flameShock := topRank(t, "Flame Shock", enh.FlameShock)

	// Flame Shock's ticks and our melee don't roll it.
	for range 1000 {
		landHit(sim, flameShock, humanoid, true)
		landHit(sim, enh.Stormstrike, humanoid, false)
	}
	if buff.IsActive() {
		t.Fatal("a tick or a melee hit gave the buff")
	}

	spellPower := enh.GetStat(stats.SpellPower)
	procs := 0
	for range 5000 {
		landHit(sim, earthShock, humanoid, false)
		if buff.IsActive() {
			procs++
			if got := enh.GetStat(stats.SpellPower) - spellPower; got != 35 {
				t.Fatalf("buff gives %.0f spell power, want 35", got)
			}
			bonus := func(target *core.Unit) float64 {
				return enh.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand].BonusSpellDamageTaken
			}
			if bonus(dragon) != 35 || bonus(humanoid) != 0 {
				t.Fatalf("buff gives %.0f more against Dragonkin and %.0f against a Humanoid, want 35 and 0", bonus(dragon), bonus(humanoid))
			}
			buff.Deactivate(sim)
		}
	}
	if rate := float64(procs) / 5000; math.Abs(rate-0.08) > 0.015 {
		t.Errorf("buff on %.1f%% of Earth Shocks, want 8%%", rate*100)
	}
	if got := enh.AttackTables[dragon.UnitIndex][proto.CastType_CastTypeMainHand].BonusSpellDamageTaken; got != 0 {
		t.Errorf("%.0f spell damage against Dragonkin left after the buff", got)
	}

	_, enh = newTrinketSimWith(t, classicRuleset, 1, 10, common.DraconicInfusedEmblem)
	if enh.GetSpell(core.ActionID{ItemID: common.DraconicInfusedEmblem}) == nil {
		t.Error("Draconic Infused Emblem has no use in Classic")
	}
}

// Rune and Jewel of the Guard Captain triple their equip in a forest or grassland, Briarwood
// Reed doubles its in a swamp. Classic has none of it.
func TestForeverTerrainEquips(t *testing.T) {
	for _, tc := range []struct {
		item   int32
		stat   stats.Stat
		extras map[proto.OutdoorTerrain]float64
	}{
		{common.RuneOfTheGuardCaptain, stats.AttackPower, map[proto.OutdoorTerrain]float64{
			proto.OutdoorTerrain_OutdoorTerrainForest: 28, proto.OutdoorTerrain_OutdoorTerrainGrassland: 28}},
		{common.RuneOfTheGuardCaptain, stats.RangedAttackPower, map[proto.OutdoorTerrain]float64{
			proto.OutdoorTerrain_OutdoorTerrainForest: 28, proto.OutdoorTerrain_OutdoorTerrainGrassland: 28}},
		{common.JewelOfTheGuardCaptain, stats.SpellPower, map[proto.OutdoorTerrain]float64{
			proto.OutdoorTerrain_OutdoorTerrainForest: 16, proto.OutdoorTerrain_OutdoorTerrainGrassland: 16}},
		{common.BriarwoodReed, stats.SpellPower, map[proto.OutdoorTerrain]float64{
			proto.OutdoorTerrain_OutdoorTerrainSwamp: 15}},
	} {
		_, base := newTrinketSim(t, 1, 10, tc.item)
		for _, terrain := range []proto.OutdoorTerrain{
			proto.OutdoorTerrain_OutdoorTerrainForest, proto.OutdoorTerrain_OutdoorTerrainGrassland,
			proto.OutdoorTerrain_OutdoorTerrainSnow, proto.OutdoorTerrain_OutdoorTerrainSwamp,
			proto.OutdoorTerrain_OutdoorTerrainMountains,
		} {
			_, enh := newTrinketSimWith(t, inTerrain(terrain), 1, 10, tc.item)
			if got := enh.GetStat(tc.stat) - base.GetStat(tc.stat); got != tc.extras[terrain] {
				t.Errorf("item %d in %s: %.0f more %s, want %.0f", tc.item, terrain, got, tc.stat.StatName(), tc.extras[terrain])
			}
			_, classic := newTrinketSimWith(t, func(request *proto.RaidSimRequest) {
				classicRuleset(request)
				inTerrain(terrain)(request)
			}, 1, 10, tc.item)
			_, classicBase := newTrinketSimWith(t, classicRuleset, 1, 10, tc.item)
			if got := classic.GetStat(tc.stat) - classicBase.GetStat(tc.stat); got != 0 {
				t.Errorf("item %d in %s gives %.0f more %s in Classic", tc.item, terrain, got, tc.stat.StatName())
			}
		}
	}
}

// Smokey's Lighter hits every enemy for 210 Fire (315 on a crit), twice that on a Plant. Its 90
// sec cooldown puts the other burst trinkets on 10 sec. Classic has no use.
func TestForeverSmokeysLighter(t *testing.T) {
	sim, enh := newTrinketSimWith(t, func(request *proto.RaidSimRequest) {
		withTargets(proto.MobFamily_MobFamilyUnknown, 60)(request)
		request.Encounter.Targets[2].MobFamily = proto.MobFamily_MobFamilyPlant
	}, 3, 100000, common.SmokeysLighter, common.GoldenBanana)
	lighter := enh.GetSpell(core.ActionID{ItemID: common.SmokeysLighter})
	if lighter == nil || enh.GetMajorCooldown(lighter.ActionID) == nil {
		t.Fatal("Smokey's Lighter has no use or it isn't a major cooldown")
	}
	banana := enh.GetSpell(core.ActionID{ItemID: common.GoldenBanana})
	if !lighter.Cast(sim, enh.CurrentTarget) {
		t.Fatal("Smokey's Lighter didn't cast")
	}
	if got := lighter.CD.TimeToReady(sim); got != 90*time.Second {
		t.Errorf("cooldown %s, want 90s", got)
	}
	if banana.CanCast(sim, enh.CurrentTarget) {
		t.Error("Golden Banana ready right after Smokey's Lighter")
	}
	sim.CurrentTime += 10 * time.Second
	if !banana.CanCast(sim, enh.CurrentTarget) {
		t.Error("Golden Banana still not ready 10 sec after Smokey's Lighter")
	}

	for range 500 {
		sim.CurrentTime += 90 * time.Second
		lighter.Cast(sim, enh.CurrentTarget)
	}
	for i, target := range sim.Encounter.TargetUnits {
		damage := 210.0
		if i == 2 {
			damage = 420
		}
		m := lighter.SpellMetrics()[target.UnitIndex]
		if m.Hits == 0 || m.Crits == 0 {
			t.Fatalf("%s: %d hits and %d crits, want both", target.Label, m.Hits, m.Crits)
		}
		if want := damage*float64(m.Hits) + damage*1.5*float64(m.Crits); math.Abs(m.TotalDamage-want) > 1e-6 {
			t.Errorf("%s took %.1f, want %.1f (%.0f a hit)", target.Label, m.TotalDamage, want, damage)
		}
	}

	_, enh = newTrinketSimWith(t, classicRuleset, 1, 10, common.SmokeysLighter)
	if enh.GetSpell(core.ActionID{ItemID: common.SmokeysLighter}) != nil {
		t.Error("Smokey's Lighter has a use in Classic")
	}
}
