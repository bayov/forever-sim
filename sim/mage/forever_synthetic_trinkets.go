package mage

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic level 60 mage on-use trinket (tools/database/forever_synthetic_items_mage.go),
// standards 1b and 8a: +140 spell power for 20 sec on the shared offensive trinket cooldown, 2 min.
const MageSyntheticTalisman = 996032

func init() {
	core.NewSimpleStatOffensiveTrinketEffect(MageSyntheticTalisman, stats.Stats{stats.SpellPower: 140}, time.Second*20, time.Minute*2)
}
