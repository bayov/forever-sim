package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The on-use halves of the druid's synthetic Phase 1 trinket pairs (standards 1b,
// tools/database/forever_synthetic_items_druid.go). Each is worth what Talisman of Ephemeral
// Power (spell power) or Earthstrike (attack power, Strength, Agility) is worth over a 3 minute
// fight: 20 sec every 2 min, used at 0 and 120 sec, averaging 27 weighted points for spell power
// (140 * 0.86 * 40 / 180) and 37 for the physical stats (165 Strength * 40 / 180). They share
// the offensive trinket cooldown like any on-use trinket. For the cat, 165 Strength sims 2.9
// ahead of 165 Agility and 3.0 ahead of 335 attack power (688.2, 685.3, 685.2 at 180 sec).
const (
	BalanceSyntheticSpellPowerTrinket = 998015
	FeralSyntheticStrengthTrinket     = 998115
)

func init() {
	core.NewSimpleStatOffensiveTrinketEffect(BalanceSyntheticSpellPowerTrinket, stats.Stats{stats.SpellPower: 140}, time.Second*20, time.Minute*2)
	core.NewSimpleStatOffensiveTrinketEffect(FeralSyntheticStrengthTrinket, stats.Stats{stats.Strength: 165}, time.Second*20, time.Minute*2)
}
