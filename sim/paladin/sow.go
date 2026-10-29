package paladin

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// sealOfWisdomProcChance is the chance a melee hit under Seal of Wisdom gives the paladin mana. The
// client's aura rows say 100 (20166, 20356, 20357: ProcChance 100, proc types 4 and 16, the
// paladin's melee swings and melee abilities) and its tooltip only "a chance", so the server decides,
// as it does Judgement of Wisdom's 50% (core.judgementOfWisdomAura). We assume the same 50%.
// Unverified; it matters little, since a rotation judges the seal and replaces it at once.
const sealOfWisdomProcChance = 0.5

// Seal of Wisdom (client 1.60.1.70685): 30 sec, each melee attack has a chance to give the
// paladin 50 / 71 / 90 mana (20168, 20350, 20351). Its judgement puts Judgement of Wisdom on the
// enemy for 40 sec, 33 / 46 / 59 mana to each attacker on a proc, and the paladin's melee hits
// refresh it. Ranks at 38, 48 and 58 for 135, 170 and 200 mana.
//
// Only a rotation that names it registers it (core.RotationReferencesSpell), so its seal auras
// don't join the results of the builds that never cast it.
func (paladin *Paladin) registerSealOfWisdom() {
	if !paladin.usesSealOfWisdom {
		return
	}

	var ranks = []struct {
		level    int32
		spellID  int32
		manaCost float64
		procID   int32
		procMana float64
		judgeID  int32
		jowMana  float64
	}{
		{level: 38, spellID: 20166, manaCost: 135, procID: 20168, procMana: 50, judgeID: 20186, jowMana: 33},
		{level: 48, spellID: 20356, manaCost: 170, procID: 20350, procMana: 71, judgeID: 20354, jowMana: 46},
		{level: 58, spellID: 20357, manaCost: 200, procID: 20351, procMana: 90, judgeID: 20355, jowMana: 59},
	}

	judgementDuration := core.TernaryDuration(paladin.Env.IsForever(), time.Second*40, time.Second*10)

	// One debuff per target, built with the highest rank the paladin knows, as Judgement of the
	// Crusader's is (sotc.go).
	top := -1
	for i, rank := range ranks {
		if paladin.Level >= rank.level {
			top = i
		}
	}
	if top < 0 {
		return
	}
	jowMana := ranks[top].jowMana
	paladin.judgementOfWisdomAuras = paladin.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.JudgementOfWisdomCastAura(&paladin.Unit, target, jowMana, judgementDuration)
	})

	for i, rank := range ranks {
		rank := rank
		if paladin.Level < rank.level {
			break
		}

		judgeSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.judgeID},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagMeleeMetrics,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
				// One judgement per paladin.
				core.DropJudgementOfTheCrusader(sim, target, &paladin.Unit)
				core.JudgeWisdom(sim, paladin.judgementOfWisdomAuras.Get(target), &paladin.Unit, jowMana)
			},
		})

		manaMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: rank.procID})
		aura := paladin.RegisterAura(core.Aura{
			Label:    "Seal of Wisdom" + paladin.Label + strconv.Itoa(i+1),
			ActionID: core.ActionID{SpellID: rank.spellID},
			Duration: time.Second * 30,
			OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMelee) {
					return
				}
				if sim.Proc(sealOfWisdomProcChance, "Seal of Wisdom") {
					paladin.AddMana(sim, rank.procMana, manaMetrics)
				}
			},
		})
		paladin.aurasSoW = append(paladin.aurasSoW, aura)

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

		paladin.spellsJoW = append(paladin.spellsJoW, judgeSpell)
	}
}
