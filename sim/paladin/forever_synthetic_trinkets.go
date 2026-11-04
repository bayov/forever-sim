package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic Phase 1 retribution on-use trinket (tools/database/forever_synthetic_items_paladin.go,
// standards 1b and 8a). It averages the standard's on-use value over a 180 sec fight: 140 spell
// power for 20 sec every 2 min, used at 0 and 120 sec, is 140 x 0.86 x 40 / 180 = 27. A hybrid may
// take the attack power kind instead (165 Strength or 335 attack power, 37); on the synthetic set
// those were 0.6 to 2.7 behind on the twist build and 9 to 10 behind on Shock.
const SyntheticSpellPowerOnUseTrinket = 992015

// The Protection paladin tank's on-use trinket (tools/database/forever_synthetic_items_paladin_tank.go,
// standards 36): 165 Strength for 20 sec every 2 min (37 weighted points over 180 sec, as the other
// physical on-use trinkets). It fires on a timer (the rotation's autocast line), reading no health.
// The defensive kind (1,665 bonus armor) scored lower at every preset weight.
const ProtectionSyntheticStrengthTrinket = 990716

func init() {
	core.NewSimpleStatOffensiveTrinketEffect(SyntheticSpellPowerOnUseTrinket, stats.Stats{stats.SpellPower: 140}, time.Second*20, time.Minute*2)
	core.NewSimpleStatOffensiveTrinketEffect(ProtectionSyntheticStrengthTrinket, stats.Stats{stats.Strength: 165}, time.Second*20, time.Minute*2)
}
