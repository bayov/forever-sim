package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

func (warlock *Warlock) makeImp() *WarlockPet {
	cfg := PetConfig{
		Name:                 "Imp",
		ManaPerIntellect:     5,
		MP5PerOwnerIntellect: 0.30,
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

	imp := warlock.makePet(cfg, warlock.Options.Summon == proto.WarlockOptions_Imp)
	// The Imp stands beside the warlock, closing in to Firebolt's 30 yd range when the warlock stands
	// farther (wowsims/forever pets.go), so its bolts travel the warlock's distance.
	imp.StartDistanceFromTarget = min(warlock.StartDistanceFromTarget, 30)
	return imp
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
	// client 11763). The coefficient is on the Imp's spell damage, a tenth of the warlock's
	// (makeStatInheritance) plus Demonic Knowledge's share.
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
		// Every rank's row has Speed 16 (11763 and the ranks before it), and the bolt lands after its
		// travel (wowsims/forever a21497c37b). The Imp casts back to back, with no artificial delay.
		MissileSpeed: 16,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      time.Millisecond * 1000,
				CastTime: data.castTime(),
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
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})
}
