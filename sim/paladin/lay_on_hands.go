package paladin

import (
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Lay on Hands (633, 2800 and 10310 in 1.60.1.70753): "Heals a friendly target for an amount equal
// to the Paladin's maximum health and restores 550 of their mana. Drains all of the Paladin's
// remaining Mana when used, but does not interrupt Mana regeneration." A 20 min cooldown (Classic's
// hour). Cast at a player it heals that player; at a mob, the paladin. It can't crit and nothing
// raises it, but the target's healing taken (a Mortal Strike) still applies.
func (paladin *Paladin) registerLayOnHands() {

	minLevels := []int32{50, 30, 10}
	idx := slices.IndexFunc(minLevels, func(level int32) bool {
		return paladin.Level >= level
	})

	if idx == -1 {
		return
	}

	spellID := []int32{10310, 2800, 633}[idx]
	manaReturn := []float64{550, 250, 0}[idx]

	// Only register the highest available rank of LoH (no benefit to using lower ranks)
	actionID := core.ActionID{SpellID: spellID}
	layOnHandsManaMetrics := paladin.NewManaMetrics(actionID)
	// The mana it gives back, on each player it lands on.
	targetManaMetrics := map[*core.Unit]*core.ResourceMetrics{}
	// Not a major cooldown. It used to fire itself under 10% health, which the sim only
	// tracks when a healing model is set (the UI always sets one), and then a paladin
	// taking hits dumped its whole mana pool on a heal the sim does nothing with. A
	// rotation that wants it casts it by name.
	paladin.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       core.SpellFlagAPL | core.SpellFlagMCD | core.SpellFlagHelpful | core.SpellFlagIgnoreAttackerModifiers,
		SpellSchool: core.SpellSchoolHoly,
		DefenseType: core.DefenseTypeMagic,
		SpellCode:   SpellCode_PaladinLayOnHands,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer: paladin.NewTimer(),
				// Forever cut the cooldown from an hour to 20 min.
				Duration: time.Minute * 20,
			},
		},
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			target = paladin.healTarget(target)
			if target.IsDead() {
				return
			}
			paladin.SpendMana(sim, paladin.CurrentMana(), layOnHandsManaMetrics)
			spell.CalcAndDealHealing(sim, target, paladin.MaxHealth(), spell.OutcomeHealing)
			if target.HasManaBar() {
				metrics := targetManaMetrics[target]
				if metrics == nil {
					metrics = target.NewManaMetrics(actionID)
					targetManaMetrics[target] = metrics
				}
				target.AddMana(sim, manaReturn, metrics)
			}
		},
	})

}
