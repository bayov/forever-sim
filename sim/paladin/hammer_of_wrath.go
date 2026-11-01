package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core/proto"

	"github.com/wowsims/classic/sim/core"
)

func (paladin *Paladin) registerHammerOfWrath() {
	// Forever's damage at level 60, about 5% below Classic. Ranks 1 and 2 are wowhead's
	// Forever tooltip, which shows them at their cap (level 49 and 57). Rank 3 keeps growing
	// 3.1 a level up to 65 and wowhead shows it there (489 to 538), so at 60 it is 15.5
	// lower. ForeverChanges' spellbook agrees (474 to 522).
	ranks := []struct {
		level     int32
		spellID   int32
		minDamage float64
		maxDamage float64
		manaCost  float64
	}{
		{level: 44, spellID: 24275, manaCost: 295, minDamage: 286, maxDamage: 314},
		{level: 52, spellID: 24274, manaCost: 360, minDamage: 382, maxDamage: 421},
		{level: 60, spellID: 24239, manaCost: 425, minDamage: 473.5, maxDamage: 522.5},
	}

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * 6,
	}

	for i, rank := range ranks {
		rank := rank
		if paladin.Level < rank.level {
			break
		}

		paladin.GetOrRegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeRanged,
			ProcMask:    core.ProcMaskRangedSpecial, // TODO to be tested
			Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
			CastType:    proto.CastType_CastTypeRanged,

			Rank:          i + 1,
			RequiredLevel: int(rank.level),
			SpellCode:     SpellCode_PaladinHammerOfWrath,

			ManaCost: core.ManaCostOptions{
				FlatCost: rank.manaCost,
				// Benediction's spell list in the client (1.60.1.70338, spell 20101) leaves
				// Hammer of Wrath out, even when Champion of the Light 5/5 makes it instant.
				Multiplier: paladin.holyConduit(),
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: time.Second,
					// Champion of the Light takes 0.2 sec a rank off the cast since the October 30
					// hotfix (spell 1311084 in 1.60.1.70707+hf, effect 1: aura 107 on Hammer of
					// Wrath's class mask, -1000 ms at 5 ranks), so 5/5 makes it instant. Instrument of
					// Law, which had the cut before, left the tree.
					CastTime: time.Second - time.Millisecond*200*time.Duration(paladin.Talents.ChampionOfTheLight),
				},
				IgnoreHaste: true,
				CD:          cd,
			},

			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			BonusCoefficient: 0.429,
			BonusHitRating:   paladin.hammerOfWrathHitRating(),
			BonusCritRating:  paladin.hammerOfWrathCritRating(),

			ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
				return sim.IsExecutePhase20()
			},
			ExtraCastConditionNeedsEvent: true,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				damage := sim.Roll(rank.minDamage, rank.maxDamage)
				spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeRangedHitAndCrit)
			},
		})
	}
}

// Classic's Precision is melee hit only, so the sim takes it back off this ranged attack. Forever's
// (client 20189) is an unscoped A_MOD_HIT_CHANCE, so Hammer of Wrath keeps it.
func (paladin *Paladin) hammerOfWrathHitRating() float64 {
	if paladin.Env.IsForever() {
		return 0
	}
	return -float64(paladin.Talents.Precision) * core.MeleeHitRatingPerHitChance
}

// Conviction reaches Hammer of Wrath since the October 30 hotfix: spell 20117 is A_MOD_CRIT_PCT
// (aura 290, all spells and attacks) in 1.60.1.70707+hf, where it was A_MOD_WEAPON_CRIT_PERCENT on
// melee weapons (aura 52), which left Hammer of Wrath out (wowsims/forever conviction_test.go).
func (paladin *Paladin) hammerOfWrathCritRating() float64 {
	return paladin.holyPowerCrit()
}
