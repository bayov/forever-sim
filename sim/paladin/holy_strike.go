package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Holy Strike is Forever's baseline Retribution strike: a weapon swing that lands as
// Holy damage (a percentage of weapon damage plus a flat amount) on a 10 sec cooldown.
//
// Eight ranks from level 6 to 60. The weapon part is 25/29/32/36/39/43/46/50% of weapon
// damage (the 2026-09-24 beta build raised it from 25/25/30/30/35/35/40/40 and made the
// 10 sec cooldown baseline, with Improved Holy Strike gone from the Holy tree). The flat
// Holy part grows a little per level after the rank is learned (the `ppl` markers on
// wowhead's Forever tooltip, base at the learn level plus scale per level up to five levels
// later). Being Holy it ignores armor, and being a yellow hit it gives no seal procs.
// Sacred Arbiter adds 20% damage and refreshes the judgement on the target (which a melee
// hit does anyway), Holy Power adds 3% crit per point, Iron Creed adds 5% threat per point.
// The flat Holy part takes 42.9% of spell power (the "SP mod" on wowhead's Forever spell
// pages).
var HolyStrikeLevel = [...]int{0, 6, 12, 20, 28, 36, 44, 52, 60}
var HolyStrikeSpellId = [...]int32{0, 679, 678, 1866, 680, 2495, 5569, 10332, 10333}

func (paladin *Paladin) registerHolyStrike() {
	ranks := []struct {
		level         int32
		spellID       int32
		manaCost      float64
		weaponPercent float64
		scaleLevel    int32
		holy          float64
		scale         float64
	}{
		{level: 6, spellID: 679, manaCost: 5, weaponPercent: 0.25, scaleLevel: 11, holy: 12, scale: 0.1},
		{level: 12, spellID: 678, manaCost: 9, weaponPercent: 0.29, scaleLevel: 17, holy: 17, scale: 0.1},
		{level: 20, spellID: 1866, manaCost: 12, weaponPercent: 0.32, scaleLevel: 25, holy: 19, scale: 0.2},
		{level: 28, spellID: 680, manaCost: 14, weaponPercent: 0.36, scaleLevel: 33, holy: 24, scale: 0.3},
		{level: 36, spellID: 2495, manaCost: 16, weaponPercent: 0.39, scaleLevel: 41, holy: 31, scale: 1.0},
		{level: 44, spellID: 5569, manaCost: 17, weaponPercent: 0.43, scaleLevel: 49, holy: 53, scale: 1.5},
		{level: 52, spellID: 10332, manaCost: 19, weaponPercent: 0.46, scaleLevel: 57, holy: 68, scale: 2.8},
		{level: 60, spellID: 10333, manaCost: 20, weaponPercent: 0.50, scaleLevel: 65, holy: 93, scale: 3.2},
	}

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * 10,
	}
	damageMultiplier := paladin.getWeaponSpecializationModifier()
	if paladin.Talents.SacredArbiter {
		damageMultiplier *= 1.2
	}
	threatMultiplier := 1 + 0.05*float64(paladin.Talents.IronCreed)

	for i, rank := range ranks {
		rank := rank
		if paladin.Level < rank.level {
			break
		}
		holy := rank.holy + rank.scale*float64(min(paladin.Level, rank.scaleLevel)-rank.level)

		paladin.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellCode:   SpellCode_PaladinHolyStrike,
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMelee,
			ProcMask:    core.ProcMaskMeleeMHSpecial,
			Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

			RequiredLevel: int(rank.level),
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{
				FlatCost:   rank.manaCost,
				Multiplier: paladin.benediction(),
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
				IgnoreHaste: true,
				CD:          cd,
			},

			BonusCritRating:  3 * float64(paladin.Talents.HolyPower) * core.CritRatingPerCritChance,
			DamageMultiplier: damageMultiplier,
			ThreatMultiplier: threatMultiplier,
			BonusCoefficient: 0.429,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				baseDamage := rank.weaponPercent*spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target)) + holy
				spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
			},
		})
	}
}
