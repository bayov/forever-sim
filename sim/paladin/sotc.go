package paladin

import (
	"slices"
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Seal of the Crusader swings the weapon 1.4 times as fast for 1/1.4 of the damage a swing.
const sealOfTheCrusaderSpeed = 1.4

func (paladin *Paladin) sealOfTheCrusaderActive() bool {
	return paladin.currentSeal != nil && paladin.currentSeal.IsActive() && slices.Contains(paladin.aurasSotC, paladin.currentSeal)
}

func (paladin *Paladin) registerSealOfTheCrusader() {
	type judge struct {
		spellID int32
		bonus   float64
	}

	var ranks = []struct {
		level      int32
		spellID    int32
		manaCost   float64
		scaleLevel int32
		ap         float64
		scale      float64
		judge      judge
	}{
		{level: 6, spellID: 21082, manaCost: 25, scaleLevel: 12, ap: 31, scale: 0.7, judge: judge{spellID: 21183, bonus: 20}},
		{level: 12, spellID: 20162, manaCost: 40, scaleLevel: 20, ap: 51, scale: 1.1, judge: judge{spellID: 20188, bonus: 30}},
		{level: 22, spellID: 20305, manaCost: 65, scaleLevel: 30, ap: 94, scale: 1.7, judge: judge{spellID: 20300, bonus: 50}},
		{level: 32, spellID: 20306, manaCost: 90, scaleLevel: 40, ap: 145, scale: 2, judge: judge{spellID: 20301, bonus: 80}},
		{level: 42, spellID: 20307, manaCost: 125, scaleLevel: 50, ap: 221, scale: 2.2, judge: judge{spellID: 20302, bonus: 110}},
		{level: 52, spellID: 20308, manaCost: 160, scaleLevel: 60, ap: 306, scale: 2.4, judge: judge{spellID: 20303, bonus: 140}},
	}

	// Forever folded the Classic Improved Seal of the Crusader talent into the judgement
	// (23 / 35 / 58 / 92 / 127 / 161 Holy damage taken, 15% over Classic), and the
	// judgement lasts 40 sec instead of 10. The seal's attack power is unchanged.
	improvedSotC := core.TernaryFloat64(paladin.Env.IsForever(), 1.15, 1)
	judgementDuration := core.TernaryDuration(paladin.Env.IsForever(), time.Second*40, time.Second*10)

	var libramAp, libramBonus float64
	if paladin.Ranged().ID == LibramOfFervor {
		libramAp = 48
		libramBonus = 33
	}

	// The judgement debuff is one aura per target (core.JudgementOfTheCrusaderAura registers it by
	// label), so it is built once with the highest rank the paladin knows. Built per rank, the
	// first registration (rank 1, 23 Holy) won and every rank's judgement applied it.
	top := 0
	for i, rank := range ranks {
		if paladin.Level >= rank.level {
			top = i
		}
	}
	judgementBonus := ranks[top].judge.bonus*improvedSotC + libramBonus + paladin.judgementOfTheCrusaderBonus
	debuffs := paladin.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.JudgementOfTheCrusaderAura(&paladin.Unit, target, judgementBonus, judgementDuration)
	})

	for i, rank := range ranks {
		rank := rank
		if paladin.Level < rank.level {
			break
		}

		judgeSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.judge.spellID},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagMeleeMetrics,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
				// One judgement per paladin (sow.go).
				if paladin.judgementOfWisdomAuras != nil {
					core.DropJudgementOfWisdom(sim, target, &paladin.Unit)
				}
				core.JudgeTheCrusader(sim, debuffs.Get(target), &paladin.Unit, judgementBonus)
			},
		})

		ap := rank.ap + rank.scale*float64(min(paladin.Level, rank.scaleLevel)-rank.level)

		aura := paladin.RegisterAura(core.Aura{
			Label:    "Seal of the Crusader" + paladin.Label + strconv.Itoa(i+1),
			ActionID: core.ActionID{SpellID: rank.spellID},
			Duration: time.Second * 30,
			OnGain: func(_ *core.Aura, sim *core.Simulation) {
				paladin.MultiplyMeleeSpeed(sim, sealOfTheCrusaderSpeed)
				paladin.AutoAttacks.MHAuto().DamageMultiplier /= sealOfTheCrusaderSpeed
				paladin.AddStatDynamic(sim, stats.AttackPower, ap+libramAp)
			},
			OnExpire: func(_ *core.Aura, sim *core.Simulation) {
				paladin.MultiplyMeleeSpeed(sim, 1/sealOfTheCrusaderSpeed)
				paladin.AutoAttacks.MHAuto().DamageMultiplier *= sealOfTheCrusaderSpeed
				paladin.AddStatDynamic(sim, stats.AttackPower, -(ap + libramAp))
			},
		})

		paladin.aurasSotC = append(paladin.aurasSotC, aura)

		paladin.RegisterSpell(core.SpellConfig{
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

		paladin.spellsJotC = append(paladin.spellsJotC, judgeSpell)
	}
}
