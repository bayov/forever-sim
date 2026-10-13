package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic on-use trinket of the Level 60 warrior presets (tools/database/
// forever_synthetic_items_warrior.go, standards 1b): +335 attack power for 20 sec every 2 min, on the
// shared offensive trinket cooldown. That averages 37 weighted points over a 180 sec fight, the size
// of Earthstrike.
const WarriorSyntheticBattleTrinket = 991018

func init() {
	core.NewSimpleStatOffensiveTrinketEffect(WarriorSyntheticBattleTrinket, stats.Stats{stats.AttackPower: 335, stats.RangedAttackPower: 335}, time.Second*20, time.Minute*2)
}
