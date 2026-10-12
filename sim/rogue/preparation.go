package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (rogue *Rogue) registerPreparationCD() {
	if !rogue.Talents.Preparation {
		return
	}

	rogue.Preparation = rogue.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{SpellID: 14185},
		Flags:    core.SpellFlagAPL,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: time.Minute * 10,
			},
			IgnoreHaste: true,
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			for _, affectedSpell := range rogue.preparationSpells() {
				affectedSpell.CD.Reset()
			}
		},
	})

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell:    rogue.Preparation,
		Type:     core.CooldownTypeDPS,
		Priority: core.CooldownPriorityDefault,
		// Spend it on the biggest cooldown it can bring back: Adrenaline Rush once its buff is
		// over, else Cold Blood or Premeditation. Vanish is only worth it to a rotation that uses
		// it, and those can press Preparation themselves.
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			if rogue.AdrenalineRush != nil {
				return !rogue.AdrenalineRush.CD.IsReady(sim) && !rogue.AdrenalineRushAura.IsActive()
			}
			for _, cd := range []*core.Spell{rogue.ColdBlood, rogue.Premeditation} {
				if cd != nil && cd.CD.TimeToReady(sim) > time.Second*30 {
					return true
				}
			}
			return false
		},
	})
}

// Forever's Preparation "finishes the cooldown on your other Rogue abilities", like Classic's,
// which reset every rogue spell with a cooldown, Adrenaline Rush and Blade Flurry included. The
// client spell (14185) is a dummy effect, so the list is the tooltip's.
func (rogue *Rogue) preparationSpells() []*core.Spell {
	return core.FilterSlice([]*core.Spell{
		rogue.AdrenalineRush, rogue.BladeFlurry, rogue.ColdBlood, rogue.Evasion, rogue.GhostlyStrike,
		rogue.Premeditation, rogue.Shadowstep, rogue.Vanish,
	}, func(spell *core.Spell) bool {
		return spell != nil
	})
}
