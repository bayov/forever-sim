package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic Phase 1 on-use trinkets of the shaman presets (standards 1b, tools/database/forever_synthetic_items*.go).
// Each averages the weighted value of Talisman of Ephemeral Power (spell power) or Earthstrike (attack power) over a
// 180 sec fight: 20 sec every 2 min, so two uses, 140 * 0.86 * 40 / 180 = 26.8 and 335 * 0.5 * 40 / 180 = 37.2.
const (
	ElementalSyntheticOnUseTrinket        = 990115
	EnhancementSyntheticOnUseTrinket      = 990015
	EnhancementSyntheticSpellOnUseTrinket = 990016
	// The Warden's defensive on-use trinket: 1,665 bonus armor for 20 sec every 2 min, the same 37 weighted
	// points over 180 sec at standards 36's 0.1 a point of bonus armor (1665 * 0.1 * 40 / 180 = 37.0). It's a
	// defensive cooldown (the shared defensive trinket timer), so the rotation uses it on a timer.
	WardenSyntheticOnUseTrinket = 999016
)

func init() {
	core.NewSimpleStatOffensiveTrinketEffect(ElementalSyntheticOnUseTrinket, stats.Stats{stats.SpellPower: 140}, time.Second*20, time.Minute*2)
	core.NewSimpleStatOffensiveTrinketEffect(EnhancementSyntheticOnUseTrinket, stats.Stats{stats.AttackPower: 335}, time.Second*20, time.Minute*2)
	core.NewSimpleStatOffensiveTrinketEffect(EnhancementSyntheticSpellOnUseTrinket, stats.Stats{stats.SpellPower: 140}, time.Second*20, time.Minute*2)
	core.NewSimpleStatDefensiveTrinketEffect(WardenSyntheticOnUseTrinket, stats.Stats{stats.BonusArmor: 1665}, time.Second*20, time.Minute*2)
}
