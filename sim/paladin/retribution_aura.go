package paladin

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// The paladin's own Retribution Aura (the "Aura" option). Sanctity Aura is gone under
// Forever, so this is the damage aura a ret paladin runs, and it only does anything
// while something hits the paladin. It is the same aura as the raid buff (another
// paladin's), so the two do not stack.
func (paladin *Paladin) registerRetributionAura() {
	if paladin.Options.Aura != proto.PaladinAura_RetributionAura {
		return
	}
	core.ForeverRetributionAura(paladin.GetCharacter())
}
