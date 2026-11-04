package paladin

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Seal of Fury is Forever's tanking seal, a trainer ability from level 10 (client 1.60.1.70753,
// the same rows as 70707 with its hotfix; rank 7 is 20423):
//
// "Fills the Paladin with divine fury for 30 sec, causing melee attacks to deal an additional 35
// Holy damage. While a shield is equipped, each attack also grants an absorb shield equal to 50%
// of the Holy damage dealt. Only one Seal can be active on the Paladin at any one time.
// Unleashing this Seal's energy inflicts 153 Holy damage to an enemy and taunts the target to
// attack you for 3 sec."
//
// What the sim does with the rows, as wowsims/forever's seal_of_fury.go reads them:
//   - Each landed white swing casts the rank's damage spell (20418 at rank 7: 35 Holy, 10% of
//     spell power; the seal aura's proc type 4, done melee auto attacks). It rolls only crit, as
//     the Seal of Righteousness hit does (DefenseType melee).
//   - With a shield on, the hit's damage times 50% (the seal's effect 1) becomes an absorb shield
//     on the paladin. The rows give the shield no spell or duration of its own, so a new hit's
//     shield replaces the last one and lasts the seal's 30 sec (report assumption
//     pal_seal_of_fury_shield).
//   - Judgement of Fury (20414 at rank 7): 153 Holy plus 3.69 a level from 58 to 64, a 8.76%
//     spread, 42.9% of spell power, on the melee table (DefenseType melee): it rolls hit and crit,
//     never dodged, parried or blocked, as Judgement of Righteousness does. Then the 3 sec taunt
//     (20232), which the sim doesn't model: a taunt only matters to a tank that isn't on top.
//   - Improved Seals reaches the hit and the judgement (its mask names both, as wowsims/forever's
//     talents_holy.go has it); the seal costs what the other seals cost (Benediction).
//
// Not read: the seal's effect 0 and the judgement's effect 2 (a dummy of 1607 plus 42 a level,
// 9% and 18% of spell power at rank 7), which look like Seal of Righteousness's weapon speed
// values and do nothing the tooltip names (Calls for the user in docs/tank_paladin.md). Unsealed
// Wrath's Echo of Fury (1311701) isn't modelled either; no tank takes the 31 point Retribution
// talent.
var SealOfFuryLevel = [...]int{0, 10, 18, 26, 34, 42, 50, 58}
var SealOfFuryRankSpellId = [...]int32{0, 1311649, 1311656, 20163, 20419, 20421, 20422, 20423}

func (paladin *Paladin) registerSealOfFury() {
	type judge struct {
		spellID   int32
		baseLevel int32
		maxLevel  int32
		base      float64
		perLevel  float64
		variance  float64
	}
	ranks := []struct {
		level    int32
		spellID  int32
		procID   int32
		manaCost float64
		damage   float64
		judge    judge
	}{
		{level: 10, spellID: 1311649, procID: 1311647, manaCost: 40, damage: 6, judge: judge{1311650, 10, 16, 23, 1.71, 0.1}},
		{level: 18, spellID: 1311656, procID: 1311654, manaCost: 60, damage: 9, judge: judge{1311655, 18, 24, 36.9, 2.16, 0.1}},
		{level: 26, spellID: 20163, procID: 20231, manaCost: 90, damage: 14, judge: judge{20183, 25, 31, 54, 2.52, 0.1}},
		{level: 34, spellID: 20419, procID: 20415, manaCost: 120, damage: 19, judge: judge{20411, 34, 40, 74, 2.79, 0.09756}},
		{level: 42, spellID: 20421, procID: 20416, manaCost: 140, damage: 25, judge: judge{20412, 42, 48, 96, 3.42, 0.09756}},
		{level: 50, spellID: 20422, procID: 20417, manaCost: 170, damage: 32, judge: judge{20413, 50, 56, 123, 3.69, 0.08759}},
		{level: 58, spellID: 20423, procID: 20418, manaCost: 200, damage: 35, judge: judge{20414, 58, 64, 153, 3.69, 0.08759}},
	}
	const shieldShare = 0.5
	const sealDuration = time.Second * 30

	shield := paladin.registerAbsorbShield(core.Aura{
		Label:    "Seal of Fury Shield" + paladin.Label,
		ActionID: core.ActionID{SpellID: 20423, Tag: 1},
		Duration: sealDuration,
	})
	paladin.sealOfFuryShield = shield

	for i, rank := range ranks {
		rank := rank
		if paladin.Level < rank.level {
			break
		}

		jl := float64(min(paladin.Level, rank.judge.maxLevel) - rank.judge.baseLevel)
		judgeBase := rank.judge.base + rank.judge.perLevel*max(0, jl)
		judgeMin := judgeBase * (1 - rank.judge.variance/2)
		judgeMax := judgeBase * (1 + rank.judge.variance/2)

		judgeSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.judge.spellID},
			SpellCode:   SpellCode_PaladinJudgementOfFury,
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMelee,
			ProcMask:    core.ProcMaskMeleeMHSpecial,
			Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagBinary,

			// Libram of Law's 4% names Judgement of Fury too (items.go).
			DamageMultiplier: paladin.improvedSeals() * paladin.getWeaponSpecializationModifier() * core.TernaryFloat64(paladin.Ranged().ID == LibramOfLaw, 1.04, 1),
			ThreatMultiplier: 1,
			BonusCoefficient: 0.429,
			// Forever's downranking penalty reads the seal rank's learn level (core/downrank.go).
			DownrankLevel: int(rank.level),

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, sim.Roll(judgeMin, judgeMax), spell.OutcomeMeleeSpecialNoBlockDodgeParry)
			},
		})

		procSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.procID},
			SpellCode:   SpellCode_PaladinSealOfFuryProc,
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMelee,
			ProcMask:    core.ProcMaskMeleeMHSpecial,
			Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagSuppressEquipProcs,

			DamageMultiplier: paladin.improvedSeals() * paladin.getWeaponSpecializationModifier(),
			ThreatMultiplier: 1,
			BonusCoefficient: 0.1,
			DownrankLevel:    int(rank.level),

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				result := spell.CalcAndDealDamage(sim, target, rank.damage, spell.OutcomeMeleeSpecialCritOnly)
				if paladin.PseudoStats.CanBlock {
					shield.set(sim, result.Damage*shieldShare)
				}
			},
		})

		aura := paladin.RegisterAura(core.Aura{
			Label:    "Seal of Fury" + paladin.Label + strconv.Itoa(i+1),
			ActionID: core.ActionID{SpellID: rank.spellID},
			Duration: sealDuration,
			OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if result.Landed() && spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
					procSpell.Cast(sim, result.Target)
				}
			},
		})
		paladin.aurasSoF = append(paladin.aurasSoF, aura)

		paladin.sealOfFury = paladin.RegisterSpell(core.SpellConfig{
			ActionID:    aura.ActionID,
			SpellSchool: core.SpellSchoolHoly,
			Flags:       core.SpellFlagAPL,

			RequiredLevel: int(rank.level),
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{
				FlatCost:   rank.manaCost - paladin.getLibramSealCostReduction(),
				Multiplier: paladin.sealCostMultiplier(),
				Factor:     paladin.sealCostFactor(),
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

		paladin.spellsJoF = append(paladin.spellsJoF, judgeSpell)
	}
}
