package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic on-use trinket of the rogue's Level 60 presets (tools/database/forever_synthetic_items_rogue.go,
// standards 1b and 8a on the coordination branch): +335 attack power for 20 sec every 2 min on the offensive
// trinket cooldown, Earthstrike's size, since Forever's real Phase 1 trinkets are removed or not simulated.
const RogueSyntheticOnUseTrinket = 994202

func init() {
	core.NewSimpleStatOffensiveTrinketEffect(RogueSyntheticOnUseTrinket, stats.Stats{stats.AttackPower: 335, stats.RangedAttackPower: 335}, time.Second*20, time.Minute*2)
}
