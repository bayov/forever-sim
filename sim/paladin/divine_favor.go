package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Divine Favor (spell 20216 in 1.60.1.70753): "When activated, gives your next Flash of Light, Holy
// Light, or Holy Shock spell a 100% critical effect chance." Aura 107 (flat crit) on the class mask
// of Holy Light, Flash of Light and Holy Shock (its damage and its heal), one charge. 4% of base
// mana, off the global cooldown, 2 min cooldown. Benediction's "all instant cast spells" takes its
// share off the cost.
func (paladin *Paladin) registerDivineFavor() {
	if !paladin.Talents.DivineFavor {
		return
	}

	var affectedSpells []*core.Spell
	paladin.OnSpellRegistered(func(spell *core.Spell) {
		if spell.SpellCode == SpellCode_PaladinHolyShock || isPaladinHeal(spell) {
			affectedSpells = append(affectedSpells, spell)
		}
	})

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Minute * 2,
	}

	consume := func(aura *core.Aura, sim *core.Simulation) {
		// Remove the buff and put the skill on cooldown.
		aura.Deactivate(sim)
		cd.Set(sim.CurrentTime + cd.Duration)
		paladin.UpdateMajorCooldowns()
	}
	aura := paladin.RegisterAura(core.Aura{
		Label:    "Divine Favor",
		ActionID: core.ActionID{SpellID: 20216},
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) {
				spell.BonusCritRating += core.SpellCritRatingPerCritChance * 100
			})
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) {
				spell.BonusCritRating -= core.SpellCritRatingPerCritChance * 100
			})
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.SpellCode == SpellCode_PaladinHolyShock {
				consume(aura, sim)
			}
		},
		OnHealDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if isPaladinHeal(spell) {
				consume(aura, sim)
			}
		},
	})

	divineFavor := paladin.RegisterSpell(core.SpellConfig{
		ActionID: aura.ActionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,
		ManaCost: core.ManaCostOptions{
			// No Benediction: 20216's SpellClassOptions mask is empty, so Benediction's class mask
			// (20101 e0 aura 108 misc 14) can't reach it, whatever the tooltip says.
			BaseCost: 0.04,
		},
		Cast: core.CastConfig{
			CD: cd,
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			aura.Activate(sim)
		},
	})
	paladin.AddMajorCooldown(core.MajorCooldown{
		Spell: divineFavor,
		Type:  core.CooldownTypeDPS,
	})
}
