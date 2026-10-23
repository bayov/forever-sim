package paladin

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (paladin *Paladin) registerSealOfRighteousness() {
	type proc struct {
		spellID int32
		value   float64
		scale   float64
		coeff   float64
	}

	type judge struct {
		spellID   int32
		minDamage float64
		maxDamage float64
		scale     float64
		coeff     float64
	}

	var ranks = []struct {
		level      int32
		spellID    int32
		manaCost   float64
		scaleLevel int32
		proc       proc
		judge      judge
	}{
		{level: 1, spellID: 20154, manaCost: 20, scaleLevel: 7, proc: proc{spellID: 25742, value: 108, scale: 18, coeff: 0.029}, judge: judge{spellID: 20187, minDamage: 15, maxDamage: 15, scale: 1.8, coeff: 0.144}},
		{level: 10, spellID: 20287, manaCost: 40, scaleLevel: 16, proc: proc{spellID: 25740, value: 216, scale: 17, coeff: 0.063}, judge: judge{spellID: 20280, minDamage: 25, maxDamage: 27, scale: 1.9, coeff: 0.312}},
		{level: 18, spellID: 20288, manaCost: 60, scaleLevel: 24, proc: proc{spellID: 25739, value: 352, scale: 23, coeff: 0.093}, judge: judge{spellID: 20281, minDamage: 39, maxDamage: 43, scale: 2.4, coeff: 0.462}},
		{level: 26, spellID: 20289, manaCost: 90, scaleLevel: 32, proc: proc{spellID: 25738, value: 541, scale: 31, coeff: 0.1}, judge: judge{spellID: 20282, minDamage: 57, maxDamage: 63, scale: 2.8, coeff: 0.5}},
		{level: 34, spellID: 20290, manaCost: 120, scaleLevel: 40, proc: proc{spellID: 25737, value: 785, scale: 37, coeff: 0.1}, judge: judge{spellID: 20283, minDamage: 78, maxDamage: 86, scale: 3.1, coeff: 0.5}},
		{level: 42, spellID: 20291, manaCost: 140, scaleLevel: 48, proc: proc{spellID: 25736, value: 1082, scale: 41, coeff: 0.1}, judge: judge{spellID: 20284, minDamage: 102, maxDamage: 112, scale: 3.8, coeff: 0.5}},
		{level: 50, spellID: 20292, manaCost: 170, scaleLevel: 56, proc: proc{spellID: 25735, value: 1407, scale: 47, coeff: 0.1}, judge: judge{spellID: 20285, minDamage: 131, maxDamage: 143, scale: 4.1, coeff: 0.5}},
		{level: 58, spellID: 20293, manaCost: 200, scaleLevel: 60, proc: proc{spellID: 25713, value: 1786, scale: 47, coeff: 0.1}, judge: judge{spellID: 20286, minDamage: 162, maxDamage: 178, scale: 4.1, coeff: 0.5}},
	}

	// Improved Seals is Forever's version of Improved Seal of Righteousness, and it scales
	// the judgement too. The judgement's spell power coefficient is 42.9% on every rank
	// under Forever (Classic ramped it up over the first three ranks), its base damage
	// is Classic's because wowhead hides it behind the coefficient. The seal's proc takes
	// 12% of spell power with a two-hander and 10% with a one-hander on every rank, counted
	// once (Classic started at 2.9%).
	//
	// Classic's talent raised the seal's base points, so it left the spell power part
	// alone. Forever's Improved Seals is a percent damage modifier on the proc and the
	// judgement spells themselves (client 1.60.1.70338, spell 20224, aura 108 with misc 0),
	// so it scales the whole hit, spell power included.
	improvedSoR := paladin.improvedSeals()

	for i, rank := range ranks {
		rank := rank
		if paladin.Level < rank.level {
			break
		}

		/*
		 * Seal of Righteousness is a Spell/Aura that when active makes the paladin capable of procing
		 * two different SpellIDs depending on a paladin's casted spell or melee swing.
		 *
		 * (Judgement of Righteousness):
		 *   - Deals flat damage that is affected by Improved SoR talent, and
		 *     has a spellpower scaling that is unaffected by that talent.
		 *   - Targets magic defense and rolls to hit and crit (melee under Forever, see below).
		 *
		 * (Seal of Righteousness):
		 *   - Procs from white hits.
		 *   - Cannot miss or be dodged/parried/blocked if the underlying white hit lands.
		 *   - Deals damage that is a function of weapon speed, and spellpower.
		 *   - Has 0.85 scale factor on base damage if using 1h, 1.2 if using 2h (Forever lowers
		 *     the 1h factor below 2.0 speed, see below).
		 *   - Calculates damage including spellpower scaling but ignoring damage multipliers,
		 *      then feeds that value as base damage into the proc spell.
		 */

		minDamage := rank.judge.minDamage + rank.judge.scale*float64(min(paladin.Level, rank.scaleLevel)-rank.level)
		maxDamage := rank.judge.maxDamage + rank.judge.scale*float64(min(paladin.Level, rank.scaleLevel)-rank.level)

		judgeConfig := core.SpellConfig{
			SpellCode:   SpellCode_PaladinJudgementOfRighteousness,
			ActionID:    core.ActionID{SpellID: rank.judge.spellID},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellDamage,
			Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagSuppressWeaponProcs | core.SpellFlagSuppressEquipProcs | core.SpellFlagBinary,

			DamageMultiplier: improvedSoR * core.TernaryFloat64(paladin.Ranged().ID == LibramOfLaw, 1.04, 1),
			ThreatMultiplier: 1,

			BonusCoefficient: core.TernaryFloat64(paladin.Env.IsForever(), 0.429, rank.judge.coeff), // client 70425 (0.5 in 70338)
			BonusCritRating:  paladin.holyPowerCrit(),

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				baseDamage := sim.Roll(minDamage, maxDamage)
				spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			},
		}
		if paladin.Env.IsForever() {
			// Forever's judgement is a melee attack (client 20286 and the lower ranks: DefenseType 2,
			// No Active Defense), so it rolls melee hit and crit, crits for double and can't be
			// dodged, parried or blocked, and the weapon specializations reach it (wowsims/forever
			// e10ca32c3d, d35afca4da: in game One- and Two-Handed Weapon Specialization raise Seal and
			// Judgement of Righteousness). It is a main hand ability hit to the listeners, weapon procs
			// included, as Judgement of Command is: the rows carry no Suppress Weapon Procs.
			judgeConfig.DefenseType = core.DefenseTypeMelee
			judgeConfig.ProcMask = core.ProcMaskMeleeMHSpecial
			judgeConfig.Flags = core.SpellFlagMeleeMetrics | core.SpellFlagBinary
			judgeConfig.DamageMultiplier *= paladin.getWeaponSpecializationModifier()
			judgeConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				baseDamage := sim.Roll(minDamage, maxDamage)
				spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialNoBlockDodgeParry)
			}
		}
		judgeSpell := paladin.RegisterSpell(judgeConfig)

		value := 0.01 * (rank.proc.value + rank.proc.scale*float64(min(paladin.Level, rank.scaleLevel)-rank.level))

		coeff := rank.proc.coeff
		damage := value * 0.85 * paladin.MainHand().SwingSpeed
		if paladin.has2hEquipped() {
			coeff *= 1.1 // from testing in SoD
			damage = value * 1.2 * paladin.MainHand().SwingSpeed
		}
		// Season of Discovery's tests found the proc's spell power and the target's damage taken
		// bonus counted twice.
		countTwice := true
		if paladin.Env.IsForever() {
			// Forever's seal tooltip (client 1.60.1.70425, description variables 1023) adds
			// spell power once, flat per swing, while the proc spells' own coefficient is 0 since
			// 70425 (0.1 before). Blizzard's 15 October beta notes: "Fixed an issue causing Seal of
			// Righteousness to gain almost twice as much benefit from Spell Damage than intended."
			// The tooltip's 12% with a two-hander fits the game, but its 7.65% with a one-hander
			// doesn't: one-handers at 1.8, 2.2 and 2.8 speed fit 0.096 to 0.101 at 0 and 50 spell
			// power (wowsims/forever 49f94c9326, 1.60.1.70425 beta, level 30, rank 4).
			//
			// The same measurements put a one-hander's weapon factor at 0.85 from 2.0 speed up and
			// 0.02 lower for each tenth of a second below (0.79 at 1.7, as fast as the test went;
			// eighteen weapons from 1.7 to 3.7 at level 20, ranks 1-3). Two-handers are 1.2 at every
			// speed. The fast weapon penalty leaves the spell power part alone.
			countTwice = false
			if paladin.has2hEquipped() {
				coeff = 0.12
			} else {
				coeff = 0.1
				damage = value * (0.85 - 0.2*max(0, 2.0-paladin.MainHand().SwingSpeed)) * paladin.MainHand().SwingSpeed
			}
		}

		procSpell := paladin.RegisterSpell(core.SpellConfig{
			SpellCode:   SpellCode_PaladinSealOfRighteousnessProc,
			ActionID:    core.ActionID{SpellID: rank.proc.spellID},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMelee,
			ProcMask:    core.ProcMaskMeleeMHSpecial,                                   //changed to ProcMaskMeleeMHSpecial, to allow procs from weapons/oils which do proc from SoR,
			Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagSuppressEquipProcs, // but Wild Strikes does not proc, nor equip procs

			// Holy Power's list (client 5923) leaves the procs out under Forever.
			BonusCritRating: core.TernaryFloat64(paladin.Env.IsForever(), 0, paladin.holyPowerCrit()),

			DamageMultiplier: paladin.getWeaponSpecializationModifier() * improvedSoR,
			ThreatMultiplier: 1,

			BonusCoefficient: coeff,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				baseDamage := damage
				if countTwice {
					// effectively scales with coeff x 2, and damage dealt multipliers affect half the damage taken bonus
					baseDamage += spell.BonusCoefficient * (spell.GetBonusDamage(target) + target.GetSchoolBonusDamageTaken(spell))
				}
				spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialCritOnly)
			},
		})

		aura := paladin.RegisterAura(core.Aura{
			Label:    "Seal of Righteousness" + paladin.Label + strconv.Itoa(i+1),
			ActionID: core.ActionID{SpellID: rank.spellID},
			Duration: time.Second * 30,

			OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !result.Landed() {
					return
				}
				if spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
					procSpell.Cast(sim, result.Target)
				}
			},
		})

		paladin.aurasSoR = append(paladin.aurasSoR, aura)
		paladin.registerSealEcho(aura, paladin.echoOfRighteousness, func(sim *core.Simulation, target *core.Unit) {
			procSpell.Cast(sim, target)
		})

		paladin.sealOfRighteousness = paladin.RegisterSpell(core.SpellConfig{
			ActionID:    aura.ActionID,
			SpellSchool: core.SpellSchoolHoly,
			Flags:       core.SpellFlagAPL,

			RequiredLevel: int(rank.level),
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{
				FlatCost:   rank.manaCost - paladin.getLibramSealCostReduction(),
				Multiplier: paladin.sealCostMultiplier(),
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				paladin.applySeal(aura, judgeSpell, spell, sim)
			},
		})

		paladin.spellsJoR = append(paladin.spellsJoR, judgeSpell)
	}
}
