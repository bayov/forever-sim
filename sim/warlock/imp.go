package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

func (warlock *Warlock) makeImp() *WarlockPet {
	cfg := PetConfig{
		Name:          "Imp",
		PowerModifier: 0.33,
	}

	switch warlock.Level {
	case 25:
		cfg.Stats = stats.Stats{
			stats.Strength:  47,
			stats.Agility:   25,
			stats.Stamina:   49,
			stats.Intellect: 94,
			stats.Spirit:    95,
			stats.Mana:      149,
			stats.MP5:       0,
			stats.MeleeCrit: 3.454 * core.CritRatingPerCritChance,
			stats.SpellCrit: 0.9075 * core.CritRatingPerCritChance,
		}
	case 40:
		cfg.Stats = stats.Stats{
			stats.Strength:  70,
			stats.Agility:   29,
			stats.Stamina:   67,
			stats.Intellect: 163,
			stats.Spirit:    163,
			stats.Mana:      318,
			stats.MP5:       0,
			stats.MeleeCrit: 3.454 * core.CritRatingPerCritChance,
			stats.SpellCrit: 0.9075 * core.CritRatingPerCritChance,
		}
	case 50:
		cfg.Stats = stats.Stats{
			stats.Strength:  101,
			stats.Agility:   32,
			stats.Stamina:   71,
			stats.Intellect: 212,
			stats.Spirit:    211,
			stats.Mana:      476,
			stats.MP5:       0,
			stats.MeleeCrit: 3.454 * core.CritRatingPerCritChance,
			stats.SpellCrit: 0.9075 * core.CritRatingPerCritChance,
		}
	case 60:
		cfg.Stats = stats.Stats{
			stats.Strength:  122,
			stats.Agility:   35,
			stats.Stamina:   86,
			stats.Intellect: 264,
			stats.Spirit:    260,
			stats.Mana:      576,
			stats.MP5:       0,
			stats.MeleeCrit: 3.454 * core.CritRatingPerCritChance,
			stats.SpellCrit: 0.9075 * core.CritRatingPerCritChance,
		}
	}

	return warlock.makePet(cfg, warlock.Options.Summon == proto.WarlockOptions_Imp)
}

func (wp *WarlockPet) registerImpFireboltSpell() {
	warlockLevel := wp.owner.Level
	// assuming max rank available
	rank := map[int32]int{25: 3, 40: 5, 50: 6, 60: 7}[warlockLevel]

	if rank == 0 {
		rank = 1
	}

	if wp.owner.Options.MaxFireboltRank != proto.WarlockOptions_NoMaximum {
		rank = min(rank, int(wp.owner.Options.MaxFireboltRank))
	}

	// Forever halved Firebolt too (42.6 to 47.8 at rank 7 and level 60, 1.12 85 to 96,
	// client 11763). The coefficient is on the imp's own spell power, which it has none of.
	data := impFireboltRanks[rank]
	spellCoeff := data.Coef
	damageLow, damageHigh := data.rangeAt(warlockLevel)
	spellId := data.SpellID
	manaCost := data.Cost
	level := data.Level

	wp.primaryAbility = wp.RegisterSpell(core.SpellConfig{
		ActionID:      core.ActionID{SpellID: spellId},
		SpellSchool:   core.SpellSchoolFire,
		DefenseType:   core.DefenseTypeMagic,
		ProcMask:      core.ProcMaskSpellDamage,
		Rank:          rank,
		RequiredLevel: level,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      time.Millisecond * 1000,
				CastTime: data.castTime(),
			},
			// Adding an artificial CD to account for real delay in imp casts in-game
			CD: core.Cooldown{
				Timer:    wp.NewTimer(),
				Duration: time.Millisecond * 200,
			},
		},

		DamageMultiplier: 1,
		// Improved Imp: 10% a point on Firebolt.
		DamageMultiplierAdditive: 1 + 0.1*float64(wp.owner.Talents.ImprovedImp),
		ThreatMultiplier:         1,
		BonusCoefficient:         spellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(damageLow, damageHigh)

			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			spell.DealDamage(sim, result)
		},
	})
}
