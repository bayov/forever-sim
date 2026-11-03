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
// hit does anyway). Holy Power's 3% crit a point is gone with the October 30 hotfix
// (talents.go, holyPowerCrit). Iron Creed and Shieldsworn are below.
//
// The client (1.60.1.70338, wago.tools SpellEffect) builds the spell like Backstab:
// effect 1 is normalized weapon damage (effect 121) carrying the flat Holy part with a
// 25% spread and a 42.9% spell power coefficient, and effect 2 is the weapon damage
// percent (effect 31). The server adds the flat part to the weapon damage before it takes
// the percent, so the flat part and its spell power are halved at rank 8 too, the same
// rule that gives Backstab its 150% of the flat bonus. The tooltip shows the flat part
// before the percent ("plus 93"), the way Forever's Backstab tooltip does. The beta has not
// shown a Holy Strike hit yet, see notes.md. Normalized damage uses Forever's speeds, which
// the beta's Stormstrike (the same effect) showed are 0.3 above the usual ones: 3.6 for a
// two-hander and 2.7 for a one-hander (see sim/shaman/stormstrike.go).
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

	// Shieldsworn takes 4 sec off the cooldown (spell 1328544 in 1.60.1.70707+hf: aura 107 on Holy
	// Strike's class mask, -4000 ms).
	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second*10 - core.TernaryDuration(paladin.Talents.Shieldsworn, time.Second*4, 0),
	}
	damageMultiplier := paladin.getWeaponSpecializationModifier()
	if paladin.Talents.SacredArbiter {
		damageMultiplier *= 1.2
	}
	// Iron Creed: 10% more threat a rank of 3 since the October 30 hotfix (spell 1311034 in
	// 1.60.1.70707+hf, effect 0: aura 108 misc 2 at 30 for the third rank; it was 5% a rank of 5).
	threatMultiplier := 1 + 0.1*float64(paladin.Talents.IronCreed)
	ironCreed := paladin.registerIronCreed()

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

			BonusCritRating:  paladin.holyPowerCrit(),
			DamageMultiplier: damageMultiplier,
			ThreatMultiplier: threatMultiplier,
			// The flat part's 42.9% goes through the weapon percent too.
			BonusHitRating:   paladin.divinePrecisionHitRating(core.MeleeHitRatingPerHitChance),
			BonusCoefficient: 0.429 * rank.weaponPercent,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				flat := holy * (0.875 + 0.25*sim.RandomFloat("Holy Strike"))
				baseDamage := rank.weaponPercent * (paladin.holyStrikeWeaponDamage(sim, spell.MeleeAttackPower(target)) + flat)
				result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
				// Any Holy Strike gives Iron Creed's cut, landed or not: wowsims/forever checked in game
				// that one that misses or is avoided still does (talents_protection.go, applyIronCreed).
				if ironCreed != nil {
					ironCreed.Activate(sim)
				}
				if result.Landed() && paladin.Talents.Shieldsworn {
					paladin.shieldswornStrike(sim, target)
				}
			},
		})
	}
}

// holyStrikeWeaponDamage is the main hand's normalized weapon damage at Forever's speeds.
func (paladin *Paladin) holyStrikeWeaponDamage(sim *core.Simulation, attackPower float64) float64 {
	weapon := paladin.AutoAttacks.MH()
	return weapon.BaseDamage(sim) + (weapon.NormalizedSwingSpeed+0.3)*attackPower/weapon.AttackPowerPerDPS
}

// registerIronCreed is Iron Creed's second half: with Righteous Fury on, a Holy Strike takes 2% a rank off all damage taken for 6 sec (the trigger spell 1311033, aura 87 on all schools;
// the trait ranks 2 / 4 / 6% since the October 30 hotfix, 2% a rank of 5 before). Nil without the
// talent or without Righteous Fury, which is on or off for the whole fight.
func (paladin *Paladin) registerIronCreed() *core.Aura {
	if paladin.Talents.IronCreed == 0 || !paladin.Options.RighteousFury {
		return nil
	}
	damageTaken := 1 - 0.02*float64(paladin.Talents.IronCreed)
	return paladin.RegisterAura(core.Aura{
		Label:    "Iron Creed",
		ActionID: core.ActionID{SpellID: 1311033},
		Duration: time.Second * 6,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.DamageTakenMultiplier *= damageTaken
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.DamageTakenMultiplier /= damageTaken
		},
	})
}

// shieldswornStrike is Shieldsworn's shield on a Holy Strike that lands (spell 1328544 in
// 1.60.1.70707+hf): "Holy Strike now strikes with your shield, triggering your active Holy Shield
// and Shield Specialization talents without consuming their charges or triggering their
// cooldowns". So an active Holy Shield deals the damage a block would and keeps its charges, and
// Shield Specialization rolls its mana with no 3 sec cooldown. Its spell block (1328551) is not
// modelled.
func (paladin *Paladin) shieldswornStrike(sim *core.Simulation, target *core.Unit) {
	for i, aura := range paladin.holyShieldAura {
		if aura != nil && aura.IsActive() {
			paladin.holyShieldProc[i].Cast(sim, target)
		}
	}
	if paladin.shieldSpecializationMana != nil {
		paladin.shieldSpecializationMana(sim)
	}
}
