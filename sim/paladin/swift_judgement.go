package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// SwiftJudgementSpellId is Swift Judgement, the 15 point Protection talent (needs Templar's
// Bulwark). The launch realms' October 30 hotfix made it a Judgement of its own (spell 1310994 in
// 1.60.1.70707+hf): "Unleash the energy of a Seal spell upon an enemy. Does not consume the Seal.
// Refer to individual Seals for Judgement effect. Does not share a cooldown with your Judgement
// ability." Its row is Judgement's (20271: effect 77, the same attributes and range), with a 1 min
// recovery of its own, no power cost and no global cooldown (no start recovery). Before the hotfix
// it finished Judgement's cooldown and made the next one free.
//
// We read it literally (report assumption pal_swift_judgement): it judges the active seal as
// Judgement does, free, off the global cooldown, on its own 1 min cooldown, which Improved
// Judgement doesn't shorten (25956's class mask names Judgement only). Sanctified Judgement
// ("Gives your Judgement ability ...") gives no mana back for it.
const SwiftJudgementSpellId = 1310994

func (paladin *Paladin) registerSwiftJudgement() {
	if !paladin.Talents.SwiftJudgement {
		return
	}

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: SwiftJudgementSpellId},
		SpellSchool: core.SpellSchoolHoly,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagCastTimeNoGCD,

		Cast: core.CastConfig{
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Minute,
			},
		},
		ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
			return paladin.currentSeal.IsActive()
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
			paladin.castSpecificJudgement(sim, target, paladin.currentJudgement, paladin.currentSeal)
		},
	})
}
