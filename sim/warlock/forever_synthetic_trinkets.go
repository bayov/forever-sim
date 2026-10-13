package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic Phase 1 on-use trinket of the Level 60 presets (tools/database/forever_synthetic_items_warlock.go,
// standards 1b): 140 spell power for 20 sec every 2 min on the shared offensive trinket cooldown, 26.8 weighted
// points on average over a 180 sec fight, the size of Talisman of Ephemeral Power.
const WarlockSyntheticTalisman = 997022

func init() {
	core.NewSimpleStatOffensiveTrinketEffect(WarlockSyntheticTalisman, stats.Stats{stats.SpellPower: 140}, time.Second*20, time.Minute*2)
}
