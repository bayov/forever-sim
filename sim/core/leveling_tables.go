package core

import (
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

// Out of combat recovery for the leveling session (leveling.go, docs/leveling_sim.md).

// A food, water or bandage: what it restores over how long, from the level we can use it.
type restorer struct {
	minLevel int32
	itemID   int32
	amount   float64
	duration time.Duration
}

// The best vendor water and food for each level bracket: the item tooltips in
// assets/db_inputs/wowhead_item_tooltips.csv, and the Forever client's drink and food spells
// (430 to 1137 and 433 to 1131: the same amounts, as so much every 5 sec).
var levelingWater = []restorer{
	{1, 159, 151.2, 18 * time.Second},    // Refreshing Spring Water
	{5, 1179, 436.8, 21 * time.Second},   // Ice Cold Milk
	{15, 1205, 835.2, 24 * time.Second},  // Melon Juice
	{25, 1708, 1344.6, 27 * time.Second}, // Sweet Nectar
	{35, 1645, 1992, 30 * time.Second},   // Moonberry Juice
	{45, 8766, 2934, 30 * time.Second},   // Morning Glory Dew
}

var levelingFood = []restorer{
	{1, 117, 61.2, 18 * time.Second},    // Tough Jerky
	{5, 2287, 243.6, 21 * time.Second},  // Haunch of Meat
	{15, 3770, 552, 24 * time.Second},   // Mutton Chop
	{25, 3771, 874.8, 27 * time.Second}, // Wild Hog Shank
	{35, 4599, 1392, 30 * time.Second},  // Cured Ham Steak
	{45, 8952, 2148, 30 * time.Second},  // Roasted Quail
}

// Bandages by the level whose First Aid (5 a level, kept up) reaches each one's requirement
// in the Forever item data: Linen at 1, Heavy Linen at 20, Wool at 50 and so on.
var levelingBandages = []restorer{
	{1, 1251, 66, 6 * time.Second},     // Linen Bandage, First Aid 1
	{4, 2581, 114, 6 * time.Second},    // Heavy Linen Bandage, 20
	{10, 3530, 161, 7 * time.Second},   // Wool Bandage, 50
	{15, 3531, 301, 7 * time.Second},   // Heavy Wool Bandage, 75
	{20, 6450, 400, 8 * time.Second},   // Silk Bandage, 100
	{25, 6451, 640, 8 * time.Second},   // Heavy Silk Bandage, 125
	{30, 8544, 800, 8 * time.Second},   // Mageweave Bandage, 150
	{35, 8545, 1104, 8 * time.Second},  // Heavy Mageweave Bandage, 175
	{40, 14529, 1360, 8 * time.Second}, // Runecloth Bandage, 200
	{45, 14530, 2000, 8 * time.Second}, // Heavy Runecloth Bandage, 225
}

// Recently Bandaged keeps us from the next bandage for 60 sec.
const recentlyBandagedDuration = time.Minute

// bestRestorer is the best of the list a level can use.
func bestRestorer(list []restorer, level int32) restorer {
	best := list[0]
	for _, r := range list {
		if r.minLevel <= level {
			best = r
		}
	}
	return best
}

// outOfCombatHealthRegen is the health a second we regen out of combat, standing, from
// Spirit: the 1.12 server's formula by class (cmangos Unit::OCTRegenHPPerSpirit). Sitting
// to eat or drink makes it half again as much. Forever's own rule is still to be seen.
func outOfCombatHealthRegen(class proto.Class, spirit float64) float64 {
	var regen float64
	switch class {
	case proto.Class_ClassDruid, proto.Class_ClassMage:
		regen = spirit*0.11 + 1
	case proto.Class_ClassHunter:
		regen = spirit*0.43 - 5.5
	case proto.Class_ClassPaladin:
		regen = spirit * 0.25
	case proto.Class_ClassPriest:
		regen = spirit*0.15 + 1.4
	case proto.Class_ClassRogue:
		regen = spirit*0.84 - 13
	case proto.Class_ClassShaman:
		regen = spirit*0.28 - 3.6
	case proto.Class_ClassWarlock:
		regen = spirit*0.12 + 1.5
	case proto.Class_ClassWarrior:
		regen = spirit*1.26 - 22.6
	}
	return max(regen, 0)
}

// mobKillXP is the XP a normal mob of mobLevel gives at our level, by the 1.12 server's
// formula (cmangos MaNGOS::XP::BaseGain): our level * 5 + 45 for a mob of our level, 5% more
// a level above us up to 4, and less a level below us down to nothing once the mob is
// trivial (gray). Forever may have changed it.
func mobKillXP(level, mobLevel int32) float64 {
	base := float64(level*5 + 45)
	if mobLevel >= level {
		return base * (1 + 0.05*float64(min(mobLevel-level, 4)))
	}
	diff := level - mobLevel
	if diff > trivialLevelDifference(level) {
		return 0
	}
	return base * (1 - float64(diff)/float64(zeroDifference(level)))
}

func trivialLevelDifference(level int32) int32 {
	switch {
	case level < 10:
		return 4
	case level < 20:
		return 5
	case level < 30:
		return 6
	case level < 40:
		return 7
	case level < 45:
		return 8
	case level < 50:
		return 9
	case level < 55:
		return 10
	case level < 60:
		return 11
	default:
		return 12
	}
}

func zeroDifference(level int32) int32 {
	switch {
	case level < 8:
		return 5
	case level < 10:
		return 6
	case level < 12:
		return 7
	case level < 16:
		return 8
	case level < 20:
		return 9
	case level < 30:
		return 11
	case level < 40:
		return 12
	case level < 45:
		return 13
	case level < 50:
		return 14
	case level < 55:
		return 15
	case level < 60:
		return 16
	default:
		return 17
	}
}
