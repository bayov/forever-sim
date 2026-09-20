package paladin

import (
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core"
)

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
	layOnHandsHealthMetrics := paladin.NewHealthMetrics(actionID)
	// Not a major cooldown. It used to fire itself under 10% health, which the sim only
	// tracks when a healing model is set (the UI always sets one), and then a paladin
	// taking hits dumped its whole mana pool on a heal the sim does nothing with. A
	// rotation that wants it casts it by name.
	paladin.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       core.SpellFlagAPL | core.SpellFlagMCD,
		SpellSchool: core.SpellSchoolHoly,
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
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			paladin.SpendMana(sim, paladin.CurrentMana(), layOnHandsManaMetrics)
			paladin.GainHealth(sim, paladin.MaxHealth(), layOnHandsHealthMetrics)
			paladin.AddMana(sim, manaReturn, layOnHandsManaMetrics)
		},
	})

}
