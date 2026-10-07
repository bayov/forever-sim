package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (paladin *Paladin) registerHolyShock() {
	if !paladin.Talents.HolyShock {
		return
	}

	// Forever moves the talent to level 30 as rank 1 (129 to 139 on the client tooltip,
	// 160 mana) with the trainer ranks behind it, and cuts the cooldown from 30 sec to
	// 10. The trainer ranks do less damage than Classic's 204, 279 and 365 at the low
	// end (foreverchanges spellbook build 70124 and wowhead agree).
	ranks := []struct {
		level     int32
		spellID   int32
		manaCost  float64
		minDamage float64
		maxDamage float64
	}{
		{level: 30, spellID: 1311606, manaCost: 160, minDamage: 129, maxDamage: 139},
		{level: 40, spellID: 20473, manaCost: 225, minDamage: 175, maxDamage: 189},
		{level: 48, spellID: 20929, manaCost: 275, minDamage: 248, maxDamage: 268},
		{level: 56, spellID: 20930, manaCost: 325, minDamage: 334, maxDamage: 362},
	}
	// Holy Power gives Holy Shock 3% crit per point where every other spell gets 1%.
	extraCrit := 2 * float64(paladin.Talents.HolyPower) * core.SpellCritRatingPerCritChance

	for i, rank := range ranks {
		rank := rank
		if paladin.Level < rank.level {
			break
		}

		paladin.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellDamage,
			Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

			RequiredLevel: int(rank.level),
			Rank:          i + 1,

			SpellCode: SpellCode_PaladinHolyShock,

			ManaCost: core.ManaCostOptions{
				FlatCost:   rank.manaCost,
				Multiplier: paladin.benediction(),
			},

			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
				CD: core.Cooldown{
					Timer:    paladin.NewTimer(),
					Duration: time.Second * 10,
				},
			},

			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			BonusCoefficient: 0.429,
			BonusCritRating:  extraCrit,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				baseDamage := sim.Roll(rank.minDamage, rank.maxDamage)
				spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			},
		})
	}
}
