package paladin

import (
	"github.com/wowsims/classic/sim/core"
)

func (paladin *Paladin) registerRighteousFury() {
	if !paladin.Options.RighteousFury {
		return
	}
	actionID := core.ActionID{SpellID: 25780}

	// Forever's Righteous Fury is 90% Holy threat (Classic 60%), and Improved Righteous
	// Fury now takes 2% per point off all damage taken instead of adding threat.
	rfThreatMultiplier := 1.9
	damageTaken := 1 - 0.02*float64(paladin.Talents.ImprovedRighteousFury)

	paladin.OnSpellRegistered(func(spell *core.Spell) {
		if spell.SpellSchool.Matches(core.SpellSchoolHoly) {
			spell.ThreatMultiplier *= rfThreatMultiplier
		}
	})

	rfAura := core.MakePermanent(&core.Aura{Label: "Righteous Fury", ActionID: actionID})
	paladin.RegisterAura(*rfAura)
	paladin.PseudoStats.DamageTakenMultiplier *= damageTaken
}
