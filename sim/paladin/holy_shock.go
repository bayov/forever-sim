package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Holy Shock blasts an enemy for Holy damage or heals an ally (the cast spell's dummy effect picks
// one of its two spells by the target, 1311606 and the trainer ranks in 1.60.1.70753). Cast at a
// mob it deals the damage spell's; cast at a player it heals with the heal spell's
// (HolyShockHealRanks), which Healing Light, Holy Power, Boundless Light and the Oath of Light
// raise. One cost and one 10 sec cooldown for both.
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
	// Holy Power gave Holy Shock 3% crit a point until the October 30 hotfix made it crit healing
	// only (talents.go, holyPowerCrit).
	extraCrit := 0.0
	if paladin.Ranged().ID == LibramOfInfusion {
		extraCrit += 6 * core.SpellCritRatingPerCritChance
	}

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * 10,
	}

	for i, rank := range ranks {
		rank := rank
		if paladin.Level < rank.level {
			break
		}
		heal := paladin.registerHolyShockHeal(i+1, extraCrit)

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
				CD: cd,
			},

			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			BonusHitRating:   paladin.divinePrecisionHitRating(core.SpellHitRatingPerHitChance),
			BonusCoefficient: 0.429,
			BonusCritRating:  extraCrit,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				if target != nil && target.Type != core.EnemyUnit {
					paladin.holyShockAlly(sim, target, spell, heal)
					return
				}
				baseDamage := sim.Roll(rank.minDamage, rank.maxDamage)
				spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			},
			// At an ally, the heal's average (APL Expected Healing); the cast is instant, so other
			// healers never see it in flight.
			ExpectedHealing: func(target *core.Unit) float64 {
				return heal.expected(paladin.healTarget(target))
			},
		})
	}
}

// holyShockHeal is one rank of Holy Shock's heal spell and its average for Expected Healing.
type holyShockHeal struct {
	spell    *core.Spell
	expected func(target *core.Unit) float64
}

func (paladin *Paladin) registerHolyShockHeal(rank int, extraCrit float64) *holyShockHeal {
	r := HolyShockHealRanks[rank]
	low, high := r.At(paladin.Level)
	oath := paladin.oathOfLightBonus(SpellCode_PaladinHolyShockHeal, r.learnLevel)
	h := &holyShockHeal{}
	h.spell = paladin.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_PaladinHolyShockHeal,
		ActionID:    core.ActionID{SpellID: r.spellID},
		SpellSchool: core.SpellSchoolHoly,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       core.SpellFlagHelpful | core.SpellFlagNoOnCastComplete,

		RequiredLevel: int(r.learnLevel),
		Rank:          rank,

		DamageMultiplier: paladin.healingLightMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: paladin.downrankCoefficient(r.coefficient, r.learnLevel),
		BonusCritRating:  extraCrit,
		CritDamageBonus:  paladin.holyPowerCritBonus(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Boundless Light's crit on a target our Oath of Light isn't on.
			bonusCrit := paladin.boundlessLightCrit(target)
			spell.BonusCritRating += bonusCrit
			spell.CalcAndDealHealing(sim, target, sim.Roll(low, high)+paladin.oathOn(target, oath), spell.OutcomeHealingCrit)
			spell.BonusCritRating -= bonusCrit
		},
	})
	paladin.illuminationBaseCost[h.spell] = r.manaCost
	h.expected = func(target *core.Unit) float64 {
		return h.spell.ExpectedHealing(target, (low+high)/2+paladin.oathOn(target, oath))
	}
	return h
}

// holyShockAlly is a Holy Shock cast at a player: its heal, and Light's Vigil when ours is on the
// target (lights_vigil.go).
func (paladin *Paladin) holyShockAlly(sim *core.Simulation, target *core.Unit, spell *core.Spell, heal *holyShockHeal) {
	// A heal at a player who died as it landed does nothing.
	if target.IsDead() {
		return
	}
	heal.spell.Cast(sim, target)
	paladin.detonateLightsVigil(sim, target, spell)
}
