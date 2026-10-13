package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// Forever lets druids drink Thistle Tea, in Cat Form too: "Rage potions and Thistletea are also
// now usable by Druids" (Blizzard's World of Warcraft: Forever class deep dive, Druid). Like the
// rogue's (sim/rogue/thistle_tea.go): 100 Energy, a 5 min cooldown, and the conjured items' 2 min
// shared cooldown. Drunk in Cat Form at 25 Energy or less: 4.2 ahead of the rogue's 10 at
// 180 sec and 2.9 at 300, and 3.8 and 1.7 ahead of 40.
func (druid *Druid) registerThistleTeaCD() {
	if !druid.Env.IsForever() || druid.Consumes.DefaultConjured != proto.Conjured_ConjuredRogueThistleTea {
		return
	}

	actionID := core.ActionID{ItemID: 7676}
	energyMetrics := druid.NewEnergyMetrics(actionID)

	thistleTea := druid.RegisterSpell(Cat, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Minute * 5,
			},
			SharedCD: core.Cooldown{
				Timer:    druid.GetConjuredCD(),
				Duration: time.Minute * 2,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.AddEnergy(sim, 100, energyMetrics)
		},
	})

	druid.AddMajorCooldown(core.MajorCooldown{
		Spell: thistleTea.Spell,
		Type:  core.CooldownTypeDPS,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			return druid.InForm(Cat) && druid.CurrentEnergy() <= 25
		},
	})
}
