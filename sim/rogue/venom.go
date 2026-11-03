package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Venom became a passive in Blizzard's November 2 development notes: "causes your Shank ability to
// increase your Poison damage by 30% and your chance to apply Poisons by 20% for 6 seconds" (it was
// a finisher for 30% and 10%, client 1310703, which 1.60.1.70753 still carries). Shank activates
// the aura (shank.go).
func (rogue *Rogue) registerVenom() {
	if !rogue.Talents.Venom {
		return
	}

	rogue.VenomAura = rogue.RegisterAura(core.Aura{
		Label:    "Venom",
		ActionID: core.ActionID{SpellID: 1310703},
		Duration: time.Second * 6,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			rogue.additivePoisonBonusChance += 0.2
			rogue.multiplyPoisonDamage(1.3)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			rogue.additivePoisonBonusChance -= 0.2
			rogue.multiplyPoisonDamage(1 / 1.3)
		},
	})
}
