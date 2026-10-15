package core

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

// TestMobKillXP checks Classic's XP for a normal mob (cmangos MaNGOS::XP::BaseGain).
func TestMobKillXP(t *testing.T) {
	for _, c := range []struct {
		level, mobLevel int32
		want            float64
	}{
		{30, 30, 195},       // 30 * 5 + 45
		{30, 32, 195 * 1.1}, // 5% a level above
		{30, 40, 195 * 1.2}, // at most 4 levels' worth
		{30, 28, 195 * (1 - 2.0/12)},
		{30, 23, 195 * (1 - 7.0/12)}, // 7 below is still green at 30
		{30, 22, 0},                  // 8 below is gray
		{10, 5, 95 * (1 - 5.0/7)},
		{60, 48, 345 * (1 - 12.0/17)},
		{60, 47, 0},
	} {
		if got := mobKillXP(c.level, c.mobLevel); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("level %d killing a level %d: %.2f XP, want %.2f", c.level, c.mobLevel, got, c.want)
		}
	}
}

// TestOutOfCombatHealthRegen checks the 1.12 server's health regen a second from Spirit.
func TestOutOfCombatHealthRegen(t *testing.T) {
	if got := outOfCombatHealthRegen(proto.Class_ClassShaman, 50); math.Abs(got-10.4) > 1e-9 {
		t.Errorf("a shaman with 50 Spirit regens %.2f a sec, want 10.4", got)
	}
	if got := outOfCombatHealthRegen(proto.Class_ClassWarrior, 10); got != 0 {
		t.Errorf("a warrior with 10 Spirit regens %.2f a sec, want 0 (the formula is below 0)", got)
	}
}

// TestLevelingRestorers checks the best food, water and bandage for a few levels.
func TestLevelingRestorers(t *testing.T) {
	for _, c := range []struct {
		level                int32
		food, water, bandage int32
	}{
		{1, 117, 159, 1251},
		{30, 3771, 1708, 8544},
		{60, 8952, 8766, 14530},
	} {
		if f, w, b := bestRestorer(levelingFood, c.level).itemID, bestRestorer(levelingWater, c.level).itemID, bestRestorer(levelingBandages, c.level).itemID; f != c.food || w != c.water || b != c.bandage {
			t.Errorf("level %d eats %d, drinks %d and bandages with %d, want %d, %d and %d", c.level, f, w, b, c.food, c.water, c.bandage)
		}
	}
}

// TestPrepareLevelingEncounter: the pool is max pack size mobs at each level of the spread,
// a level 30 shaman's session lasts the session length, and the options' defaults apply.
func TestPrepareLevelingEncounter(t *testing.T) {
	raid := &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{{Class: proto.Class_ClassShaman, Level: 30}}}}}
	spread := int32(2)
	enc := prepareLevelingEncounter(raid, &proto.Encounter{Leveling: &proto.LevelingOptions{MobLevelOffset: 1, MobLevelSpread: &spread, MaxPackSize: 3}})
	if enc.Duration != 1200 || enc.UseHealth {
		t.Errorf("duration %.0f and use health %v, want 1200 and false", enc.Duration, enc.UseHealth)
	}
	levels := map[int32]int{}
	for _, mob := range enc.Targets {
		levels[mob.Level]++
		if mob.MobType != proto.MobType_MobTypeBeast {
			t.Errorf("a mob of type %s, want the default beast", mob.MobType)
		}
	}
	for l := int32(29); l <= 33; l++ {
		if levels[l] != 3 {
			t.Errorf("%d level %d mobs in the pool, want 3", levels[l], l)
		}
	}
	if len(levels) != 5 {
		t.Errorf("mob levels %v, want 29 to 33", levels)
	}
}
