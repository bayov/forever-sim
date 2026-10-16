package core

import (
	"fmt"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// NormalMobStats is the average normal melee mob at one level: what we grind while leveling
// (docs/leveling_sim.md). The table comes from the 1.12 world database
// (mob_stats_auto_gen.go), until our own logs check it.
type NormalMobStats struct {
	Level  int32
	Health float64
	Armor  float64

	// A white swing's damage before armor, spread evenly between the two.
	MinDamage, MaxDamage float64
	// Seconds between swings.
	SwingSpeed float64
	// Yards a second, for the time a pulled mob takes to reach us.
	RunSpeed float64

	// The number of mobs this level's average is taken over, counting a mob that spans n
	// levels as 1/n at each.
	Spawns float64
	// The share of each creature type at this level, by spawns.
	MobTypes map[proto.MobType]float64
}

// NormalMobStatsAt is the average normal mob at the given level, from 1 to 60.
func NormalMobStatsAt(level int32) NormalMobStats {
	if level < 1 || level > int32(len(normalMobStatsByLevel)) {
		panic(fmt.Sprintf("no normal mob stats for level %d", level))
	}
	return normalMobStatsByLevel[level-1]
}

// NormalMobTarget is a normal mob at the given level, tanked by the first tank (us when we
// solo), of the given creature type. armorMultiplier scales the level's average armor (1 for
// the average mob).
func NormalMobTarget(level int32, mobType proto.MobType, armorMultiplier float64) *proto.Target {
	mob := NormalMobStatsAt(level)
	return &proto.Target{
		Name:    fmt.Sprintf("Level %d mob", level),
		Level:   level,
		MobType: mobType,
		Stats: stats.Stats{
			stats.Health: mob.Health,
			stats.Armor:  mob.Armor * armorMultiplier,
		}.ToFloatArray(),
		MinBaseDamage: mob.MinDamage,
		DamageSpread:  mob.MaxDamage/mob.MinDamage - 1,
		SwingSpeed:    mob.SwingSpeed,
		ParryHaste:    true,
		TankIndex:     0,
	}
}
