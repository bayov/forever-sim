package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// registerBarkskinSpell is Barkskin under Forever (22812, level 44, client 1.60.1.70753): Physical
// damage taken 20% less for 15 sec (effect 1, aura 87 on school mask 1), no cost, a 1 min cooldown
// and the global cooldown (StartRecoveryTime 1500). It has no shapeshift row, so it's cast in any
// form, the bear forms too (Blizzard's class deep dive: "Barkskin ... can be used in form"). It
// doesn't stop auto attack (no such attribute). 1.12's was 20% of all damage for 12 sec, caster
// form only.
func (druid *Druid) registerBarkskinSpell() {
	if druid.Level < 44 {
		return
	}
	actionID := core.ActionID{SpellID: 22812}

	druid.BarkskinAura = druid.RegisterAura(core.Aura{
		Label:    "Barkskin",
		ActionID: actionID,
		Duration: time.Second * 15,
		OnGain: func(aura *core.Aura, _ *core.Simulation) {
			aura.Unit.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical] *= 0.8
		},
		OnExpire: func(aura *core.Aura, _ *core.Simulation) {
			aura.Unit.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical] /= 0.8
		},
	})

	druid.Barkskin = druid.RegisterSpell(Any, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.GCDDefault},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Minute,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.BarkskinAura.Activate(sim)
		},
	})
}
