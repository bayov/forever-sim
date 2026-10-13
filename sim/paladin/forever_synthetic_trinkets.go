package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic Phase 1 retribution on-use trinkets (tools/database/forever_synthetic_items_paladin.go,
// standards 1b and 8a). Each averages the standard's on-use value over a 180 sec fight: 20 sec
// every 2 min, used at 0 and 120 sec, is 140 spell power (x 0.86 x 40 / 180 = 27), 165 Strength
// or 335 attack power (37).
const (
	SyntheticSpellPowerOnUseTrinket  = 992015
	SyntheticStrengthOnUseTrinket    = 992016
	SyntheticAttackPowerOnUseTrinket = 992017
)

func init() {
	core.NewSimpleStatOffensiveTrinketEffect(SyntheticSpellPowerOnUseTrinket, stats.Stats{stats.SpellPower: 140}, time.Second*20, time.Minute*2)
	core.NewSimpleStatOffensiveTrinketEffect(SyntheticStrengthOnUseTrinket, stats.Stats{stats.Strength: 165}, time.Second*20, time.Minute*2)
	core.NewSimpleStatOffensiveTrinketEffect(SyntheticAttackPowerOnUseTrinket, stats.Stats{stats.AttackPower: 335, stats.RangedAttackPower: 335}, time.Second*20, time.Minute*2)
}
