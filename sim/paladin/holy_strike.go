package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Holy Strike is Forever's baseline Retribution strike: a weapon swing that lands as
// Holy damage (a percentage of weapon damage plus a flat amount) on a 12 sec cooldown.
//
// Wowhead's Forever database lists eight ranks from level 6 to 60, the percentage
// climbing from 25% to 40% and the flat part growing a little per level after the rank
// is learned (the `ppl` markers on the tooltip, base at the learn level plus scale per
// level up to five levels later). Being Holy it ignores armor, and being a yellow hit it
// gives no seal procs. Improved Holy Strike takes 1 sec per point off the cooldown,
// Sacred Arbiter adds 10% damage and refreshes the judgement on the target (which a
// melee hit does anyway), Iron Creed adds 5% threat per point.
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
		{level: 12, spellID: 678, manaCost: 9, weaponPercent: 0.25, scaleLevel: 17, holy: 17, scale: 0.1},
		{level: 20, spellID: 1866, manaCost: 12, weaponPercent: 0.30, scaleLevel: 25, holy: 19, scale: 0.2},
		{level: 28, spellID: 680, manaCost: 14, weaponPercent: 0.30, scaleLevel: 33, holy: 24, scale: 0.3},
		{level: 36, spellID: 2495, manaCost: 16, weaponPercent: 0.35, scaleLevel: 41, holy: 31, scale: 1.0},
		{level: 44, spellID: 5569, manaCost: 17, weaponPercent: 0.35, scaleLevel: 49, holy: 53, scale: 1.5},
		{level: 52, spellID: 10332, manaCost: 19, weaponPercent: 0.40, scaleLevel: 57, holy: 68, scale: 2.8},
		{level: 60, spellID: 10333, manaCost: 20, weaponPercent: 0.40, scaleLevel: 65, holy: 93, scale: 3.2},
	}

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * time.Duration(12-paladin.Talents.ImprovedHolyStrike),
	}
	damageMultiplier := paladin.getWeaponSpecializationModifier()
	if paladin.Talents.SacredArbiter {
		damageMultiplier *= 1.1
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

			DamageMultiplier: damageMultiplier,
			ThreatMultiplier: threatMultiplier,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				baseDamage := rank.weaponPercent*spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target)) + holy
				spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
			},
		})
	}
}
