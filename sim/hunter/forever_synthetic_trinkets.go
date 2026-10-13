package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The on-use halves of the hunter's synthetic trinkets (tools/database/forever_synthetic_items_hunter.go).
// Standards 1b prices an on-use trinket at 37 weighted points over a 180 sec fight, the size of
// Earthstrike: used at the pull and on cooldown, 20 sec every 2 min is 2 uses, 40 sec in 180.
// Ranged attack power costs 0.4, so 415 x 0.4 x 40 / 180 = 36.9; attack power with ranged attack
// power counts once at 0.5, so 335 x 0.5 x 40 / 180 = 37.2. Both share the offensive trinket
// cooldown, as Earthstrike does.
const (
	SyntheticRangedOnUseTrinket   = 993016
	SyntheticSurvivalOnUseTrinket = 993117
)

func init() {
	core.NewSimpleStatOffensiveTrinketEffect(SyntheticRangedOnUseTrinket, stats.Stats{stats.RangedAttackPower: 415}, time.Second*20, time.Minute*2)
	core.NewSimpleStatOffensiveTrinketEffect(SyntheticSurvivalOnUseTrinket, stats.Stats{stats.AttackPower: 335, stats.RangedAttackPower: 335}, time.Second*20, time.Minute*2)
}
