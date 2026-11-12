package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Venom became a passive in Blizzard's November 2 development notes: "causes your Shank ability to
// increase your Poison damage by 30% and your chance to apply Poisons by 20% for 6 seconds" (it was
// a finisher for 30% and 10%). Server hotfix push 113483 (launch realms, 2026-11-03) carries it: the
// talent is 1329328 and 1310703 is now its 6 sec aura, +30% to Instant Poison's damage and Deadly
// Poison's ticks (effects 0 and 1), and +20 to the proc chance of Instant, Deadly and Wound Poison
// (effect 2, class mask 268558336 = 0x10000000 Wound + 0x8000 + 0x4000 + 0x10000 Deadly + 0x2000 Instant). Shank activates the aura (shank.go).
func (rogue *Rogue) registerVenom() {
	if !rogue.Talents.Venom {
		return
	}

	rogue.VenomAura = rogue.RegisterAura(core.Aura{
		Label:    "Venom",
		ActionID: core.ActionID{SpellID: 1310703},
		Duration: time.Second * 6,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			rogue.venomPoisonChance += 0.2
			rogue.multiplyPoisonDamage(1.3)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			rogue.venomPoisonChance -= 0.2
			rogue.multiplyPoisonDamage(1 / 1.3)
		},
	})
}
