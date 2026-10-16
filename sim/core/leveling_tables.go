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
// Spirit: Wrath's game tables (healthRegenPerSpirit, tools/gen_regen_tables.py), a 2 sec tick
// of the class and level's ratio times the first 50 Spirit plus the lower ratio times the
// rest. On the Forever beta a player measured that shape: no constant, a lower ratio past 50
// (eu.forums.blizzard.com/en/wow/t/635182, 2026-10-12). Their ticks were 0.65 to 1.35 times
// these tables' (a level 8 warrior close, a level 14 paladin higher, a level 5 warlock
// lower), and the 1.12 server's formulas were further off (notes.md, Need to Verify).
//
// Warriors and rogues regen the most: at level 30 a warrior gets 0.56 a Spirit a tick and a
// rogue 0.33, a shaman 0.17 and a mage 0.10.
func outOfCombatHealthRegen(class proto.Class, level int32, spirit float64) float64 {
	ratios, ok := healthRegenPerSpirit[class]
	if !ok || spirit <= 0 {
		return 0
	}
	r := ratios[min(max(level, 1), 60)-1]
	return (min(spirit, 50)*r[0] + max(spirit-50, 0)*r[1]) / 2
}

// Sitting (to eat, drink or just rest) adds a third to Spirit's health regen. The Forever beta
// measurement above saw about 30% more; Blizzard's later games document 33%. (The 1.12 and
// 3.3.5 emulators use 50%.) Food and water come on top.
const sittingHealthRegenMultiplier = 4.0 / 3

// Troll Regeneration (racials.md, unchanged under Forever): 10% more health regen, and 10% of
// it goes on in combat. The beta measurement saw the combat part (+1 a tick against 10.5).
const (
	trollHealthRegenMultiplier = 1.1
	trollCombatHealthRegen     = 0.1
)

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
