package paladin

import (
	"github.com/wowsims/classic/sim/core"
)

func (paladin *Paladin) registerRighteousFury() {
	if !paladin.Options.RighteousFury {
		// Instrument of Law takes 10% a point off all our threat while Righteous Fury is off
		// (client 1.60.1.70630, spell 1311085). Righteous Fury is on or off for the whole fight.
		paladin.PseudoStats.ThreatMultiplier *= 1 - 0.1*float64(paladin.Talents.InstrumentOfLaw)
		return
	}
	actionID := core.ActionID{SpellID: 25780}

	// Forever's Righteous Fury is 90% Holy threat again in client 70425 (spell 25780), after 60%
	// in 70338. Improved Righteous Fury takes 2% per point off all damage taken instead of adding
	// threat.
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
