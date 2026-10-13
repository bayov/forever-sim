package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// SwiftJudgementSpellId is Swift Judgement, the 15 point Protection talent (client
// 1.60.1.70338, spell 1310994): instant and off the global cooldown, 1 min cooldown. It
// finishes Judgement's cooldown and leaves a one-charge aura (aura 108 / misc 14 at -100% on
// Judgement's class mask) that makes the next Judgement free.
const SwiftJudgementSpellId = 1310994

func (paladin *Paladin) registerSwiftJudgement() {
	if !paladin.Talents.SwiftJudgement {
		return
	}

	freeJudgement := paladin.RegisterAura(core.Aura{
		Label:    "Swift Judgement",
		ActionID: core.ActionID{SpellID: SwiftJudgementSpellId},
		Duration: core.NeverExpires,
		OnGain: func(_ *core.Aura, _ *core.Simulation) {
			paladin.judgement.Cost.Multiplier -= 100
		},
		OnExpire: func(_ *core.Aura, _ *core.Simulation) {
			paladin.judgement.Cost.Multiplier += 100
		},
	})
	paladin.swiftJudgementAura = freeJudgement

	paladin.RegisterSpell(core.SpellConfig{
		ActionID: freeJudgement.ActionID,
		Flags:    core.SpellFlagAPL | core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Minute,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			paladin.judgement.CD.Reset()
			freeJudgement.Activate(sim)
		},
	})
}
