package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Templar's Bulwark, the 11 point Protection talent (client 1.60.1.70753, spell 1311015, the
// October 30 hotfix's row): "When activated, this ability grants you an absorb shield equal to
// 100% of your maximum health for 8 sec. Applies Forbearance for 1 min. Cannot be cast while
// Forbearance is active." 110 mana, a 5 min cooldown (RecoveryTime 300000) that Sacred Duty cuts by
// 30 sec a rank, and no global cooldown (the row has no start recovery).
//
// It's an emergency button: a tank presses it when its health drops, which a scoring run can't
// follow (the replay needs nothing to read the tank's health, docs/tank_sim.md). So it's not a
// major cooldown the rotation fires by itself; a rotation casts it by name, on a health condition
// in the UI or on a timer that reads no health in a scoring run.
const TemplarsBulwarkSpellId = 1311015

func (paladin *Paladin) registerTemplarsBulwark() {
	if !paladin.Talents.TemplarsBulwark {
		return
	}
	actionID := core.ActionID{SpellID: TemplarsBulwarkSpellId}
	// The core absorb pipeline (core/absorb.go) runs it with any other shield on the paladin, oldest
	// first, and books what it absorbs as the Bulwark's healing.
	shield := paladin.NewAbsorbArray(core.Aura{
		Label:    "Templar's Bulwark" + paladin.Label,
		ActionID: actionID,
		Duration: time.Second * 8,
	}).Get(&paladin.Unit)
	paladin.templarsBulwarkShield = shield

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolHoly,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagAPL | SpellFlag_Forbearance | core.SpellFlagCastTimeNoGCD,

		ManaCost: core.ManaCostOptions{
			FlatCost:   110,
			Multiplier: paladin.benediction(),
		},
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Minute*5 - paladin.sacredDutyCooldownCut(),
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			shield.Apply(sim, spell, paladin.MaxHealth())
		},
	})
}

// sacredDutyCooldownCut is Sacred Duty's 30 sec a rank off Divine Shield, Divine Protection and
// Templar's Bulwark (1224697, effect 1: -60000 ms at 2 ranks on their class masks).
func (paladin *Paladin) sacredDutyCooldownCut() time.Duration {
	return time.Second * 30 * time.Duration(paladin.Talents.SacredDuty)
}
