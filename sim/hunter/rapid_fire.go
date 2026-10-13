package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Forever's Rapid Fire speeds up melee attacks too (40% ranged and melee for 15 sec, Classic ranged
// only), 5 min cooldown. Rapid Killing takes 1 min off a rank. Its other half, 20% on the next shot
// after a kill, needs a kill, which a boss fight doesn't have.
func (hunter *Hunter) registerRapidFire() {
	if hunter.Level < 26 {
		return
	}

	actionID := core.ActionID{SpellID: 3045}
	cooldown := time.Minute*5 - time.Minute*time.Duration(hunter.Talents.RapidKilling)

	hunter.RapidFireAura = hunter.RegisterAura(core.Aura{
		Label:    "Rapid Fire",
		ActionID: actionID,
		Duration: time.Second * 15,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			if sim.IsForever() {
				aura.Unit.MultiplyAttackSpeed(sim, 1.4)
			} else {
				aura.Unit.MultiplyRangedSpeed(sim, 1.4)
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			if sim.IsForever() {
				aura.Unit.MultiplyAttackSpeed(sim, 1/1.4)
			} else {
				aura.Unit.MultiplyRangedSpeed(sim, 1/1.4)
			}
		},
	})

	hunter.RapidFire = hunter.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: 100,
		},
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: cooldown,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			hunter.RapidFireAura.Activate(sim)
		},
	})

	hunter.AddMajorCooldown(core.MajorCooldown{
		Spell: hunter.RapidFire,
		Type:  core.CooldownTypeDPS,
	})
}
