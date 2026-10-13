package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Libram IDs
const (
	SanctifiedOrb  = 20512
	LibramOfHope   = 22401
	LibramOfFervor = 23203

	// Forever's level 60 librams (client 1.60.1.70338). Libram of Law (spell 1291086) is a 4%
	// damage modifier whose spell mask holds Judgement of Righteousness (and Judgement of Fury)
	// but not Judgement of Command, though its tooltip says "your Judgement ability". Libram of
	// Infusion (spell 1306429) is 6% crit on Holy Shock.
	LibramOfLaw      = 272435
	LibramOfInfusion = 279248
)

func init() {
	core.NewSimpleStatOffensiveTrinketEffect(SanctifiedOrb, stats.Stats{stats.MeleeCrit: 3 * core.CritRatingPerCritChance, stats.SpellCrit: 3 * core.CritRatingPerCritChance}, time.Second*25, time.Minute*3)
}
