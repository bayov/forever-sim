package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Berserk, the 31 point Feral talent (1.60.1.70338, wago.tools 417141): for 15 sec on a 3
// min cooldown, Claw, Rake, Shred, Ravage and Pounce get 100% more crit chance (its class
// mask), and Primal Bite hits 3 targets with no cooldown, which a cat never uses. It is off
// the global cooldown and needs a feral form.
const BerserkCritBonus = 100.0

func (druid *Druid) registerBerserkCD() {
	if !druid.Talents.Berserk {
		return
	}

	actionID := core.ActionID{SpellID: 417141}
	var affectedSpells []*core.Spell

	druid.BerserkAura = druid.RegisterAura(core.Aura{
		Label:    "Berserk",
		ActionID: actionID,
		Duration: time.Second * 15,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells = core.FilterSlice(druid.Spellbook, func(spell *core.Spell) bool {
				return spell.Flags.Matches(SpellFlagBuilder)
			})
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedSpells {
				spell.BonusCritRating += BerserkCritBonus * core.CritRatingPerCritChance
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedSpells {
				spell.BonusCritRating -= BerserkCritBonus * core.CritRatingPerCritChance
			}
		},
	})

	druid.Berserk = druid.RegisterSpell(Cat|Bear, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Minute * 3,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.BerserkAura.Activate(sim)
		},
	})
}
