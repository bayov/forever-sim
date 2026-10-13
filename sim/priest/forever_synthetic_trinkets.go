package priest

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic on-use trinket of the Level 60 presets (tools/database/forever_synthetic_items_priest.go,
// standards 1b): 140 spell power for 20 sec every 2 min on the shared offensive trinket cooldown.
const PriestSyntheticOnUseTrinket = 995034

func init() {
	core.NewSimpleStatOffensiveTrinketEffect(PriestSyntheticOnUseTrinket, stats.Stats{stats.SpellPower: 140}, time.Second*20, time.Minute*2)
}
