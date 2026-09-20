package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (paladin *Paladin) registerJudgement() {
	// Sanctified Judgement gives back part of the judged seal's mana cost.
	refundChance := []float64{0, 0.33, 0.66, 1}[paladin.Talents.SanctifiedJudgement]
	refundPortion := []float64{0, 0.2, 0.4, 0.6}[paladin.Talents.SanctifiedJudgement]
	refundMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: 1311074})

	// Judgement functions as a dummy spell in vanilla.
	// It rolls on the spell hit table and can only miss or hit.
	// Individual seals have their own effects that this spell triggers,
	// that are handled in the implementations of the seal auras.
	paladin.judgement = paladin.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 20271},
		SpellSchool: core.SpellSchoolHoly,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagCastTimeNoGCD,

		ManaCost: core.ManaCostOptions{
			BaseCost:   0.06,
			Multiplier: paladin.benediction(),
		},

		Cast: core.CastConfig{
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Second * (10 - time.Duration(paladin.Talents.ImprovedJudgement)),
			},
		},
		ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
			return paladin.currentSeal.IsActive()
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
			paladin.castSpecificJudgement(sim, target, paladin.currentJudgement, paladin.currentSeal)
			if refundChance > 0 && sim.Proc(refundChance, "Sanctified Judgement") {
				paladin.AddMana(sim, refundPortion*paladin.currentSealSpell.DefaultCast.Cost, refundMetrics)
			}
		},
	})
}

// Helper Function For casting Judgement.
//
// Under Forever a Judgement no longer consumes the seal, so the paladin judges every
// cooldown and only reseals when the 30 sec run out. Classic burns the seal.
func (paladin *Paladin) castSpecificJudgement(sim *core.Simulation, target *core.Unit, judgementSpell *core.Spell, matchingSeal *core.Aura) {
	judgementSpell.Cast(sim, target)
	if !sim.IsForever() {
		matchingSeal.Deactivate(sim)
	}
}
