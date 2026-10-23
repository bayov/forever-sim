package warlock

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

func (warlock *Warlock) makeVoidwalker() *WarlockPet {
	cfg := PetConfig{
		Name:                 "Voidwalker",
		ManaPerIntellect:     11.5,
		MP5PerOwnerIntellect: 0.10,
		// Its melee is halved (wowsims/forever 43fce5a834: at level 30, +20 attack power moved its sheet's
		// damage by half of 20 x 2 / 14).
		DamageMultiplier: 0.5,
	}

	switch warlock.Level {
	case 25:
		cfg.Stats = stats.Stats{
			stats.Strength:  50,
			stats.Agility:   40,
			stats.Stamina:   87,
			stats.Intellect: 35,
			stats.Spirit:    61,
			stats.Mana:      60,
			stats.MP5:       0,
			stats.MeleeCrit: 3.2685 * core.CritRatingPerCritChance,
			stats.SpellCrit: 3.3355 * core.CritRatingPerCritChance,
		}
		cfg.AutoAttacks = core.AutoAttackOptions{
			MainHand: core.Weapon{
				BaseDamageMin: 2,
				BaseDamageMax: 7,
				SwingSpeed:    2,
			},
			AutoSwingMelee: true,
		}
	case 40:
		cfg.Stats = stats.Stats{
			stats.Strength:  74,
			stats.Agility:   58,
			stats.Stamina:   148,
			stats.Intellect: 49,
			stats.Spirit:    97,
			stats.Mana:      637,
			stats.MP5:       0,
			stats.MeleeCrit: 3.2685 * core.CritRatingPerCritChance,
			stats.SpellCrit: 3.3355 * core.CritRatingPerCritChance,
		}
		cfg.AutoAttacks = core.AutoAttackOptions{
			MainHand: core.Weapon{
				BaseDamageMin: 5,
				BaseDamageMax: 15,
				SwingSpeed:    2,
			},
			AutoSwingMelee: true,
		}
	case 50:
		cfg.Stats = stats.Stats{
			stats.Strength:  107,
			stats.Agility:   71,
			stats.Stamina:   190,
			stats.Intellect: 59,
			stats.Spirit:    123,
			stats.Mana:      1028,
			stats.MP5:       0,
			stats.MeleeCrit: 3.2685 * core.CritRatingPerCritChance,
			stats.SpellCrit: 3.3355 * core.CritRatingPerCritChance,
		}
		cfg.AutoAttacks = core.AutoAttackOptions{
			MainHand: core.Weapon{
				// Not updated
				BaseDamageMin: 5,
				BaseDamageMax: 15,
				SwingSpeed:    2,
			},
			AutoSwingMelee: true,
		}
	case 60:
		cfg.Stats = stats.Stats{
			stats.Strength:  129,
			stats.Agility:   85,
			stats.Stamina:   234,
			stats.Intellect: 70,
			stats.Spirit:    150,
			stats.Mana:      1066,
			stats.MP5:       0,
			stats.MeleeCrit: 3.2685 * core.CritRatingPerCritChance,
			stats.SpellCrit: 3.3355 * core.CritRatingPerCritChance,
		}
		cfg.AutoAttacks = core.AutoAttackOptions{
			MainHand: core.Weapon{
				BaseDamageMin: voidwalkerSwingAverage * (1 - 0.264),
				BaseDamageMax: voidwalkerSwingAverage * (1 + 0.264),
				SwingSpeed:    2,
			},
			AutoSwingMelee: true,
		}
	}

	cfg.Stats[stats.AttackPower] = demonAttackPower(warlock.Level)
	return warlock.makePet(cfg, warlock.Options.Summon == proto.WarlockOptions_Voidwalker)
}
