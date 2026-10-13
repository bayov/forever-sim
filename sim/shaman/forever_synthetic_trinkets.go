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
)

func init() {
	core.NewSimpleStatOffensiveTrinketEffect(ElementalSyntheticOnUseTrinket, stats.Stats{stats.SpellPower: 140}, time.Second*20, time.Minute*2)
	core.NewSimpleStatOffensiveTrinketEffect(EnhancementSyntheticOnUseTrinket, stats.Stats{stats.AttackPower: 335}, time.Second*20, time.Minute*2)
	core.NewSimpleStatOffensiveTrinketEffect(EnhancementSyntheticSpellOnUseTrinket, stats.Stats{stats.SpellPower: 140}, time.Second*20, time.Minute*2)
}
