package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

type PetAbilityType int

// Pet AI doesn't use abilities immediately, so model this with a 1.6s GCD.
const PetGCD = time.Millisecond * 1600

// Under Forever the pet's global cooldown is the client's 1.5 sec (GCDMs on every pet ability row),
// as wowsims/forever has it.
const ForeverPetGCD = time.Millisecond * 1500

func (hp *HunterPet) gcd() time.Duration {
	if hp.Env.IsForever() {
		return ForeverPetGCD
	}
	return PetGCD
}

const (
	Unknown PetAbilityType = iota
	Bite
	Claw
	Screech
	FuriousHowl
	LightningBreath
	ScorpidPoison
)

func (hp *HunterPet) NewPetAbility(abilityType PetAbilityType, isPrimary bool) *core.Spell {
	switch abilityType {
	case Bite:
		return hp.newBite()
	case Claw:
		return hp.newClaw()
	case Screech:
		return hp.newScreech()
	// case FuriousHowl:
	// 	return hp.newFuriousHowl()
	case LightningBreath:
		return hp.newLightningBreath()
	case ScorpidPoison:
		return hp.newScorpidPoison()
	// case Swipe:
	// 	return hp.newSwipe()
	case Unknown:
		return nil
	default:
		panic("Invalid pet ability type")
	}
}

// Claw and Bite by the owner's level: the damage range is the client's base points times 1 -/+ half
// its variance (SpellEffect, build 1.60.1.70338, unchanged from Classic). TestPetAbilityDamage
// checks them.
var (
	petClawDamage = map[int32][2]float64{25: {16, 22}, 40: {26, 36}, 50: {35, 49}, 60: {43, 59}}
	petBiteDamage = map[int32][2]float64{25: {31, 37}, 40: {49, 59}, 50: {66, 80}, 60: {81, 99}}
)

// The level 60 ranks under Forever, from the client rows the same way: Lightning Breath rank 6
// (25012: 92, variance 0.132; Classic 99 to 113), Demoralizing Screech rank 4 (24579: 33, variance
// 0.556; Classic 26 to 46) and Scorpid Poison rank 4's tick (24587: 5; Classic 8).
var (
	foreverLightningBreath60 = [2]float64{86, 98}
	foreverScreech60         = [2]float64{24, 42}
)

const foreverScorpidPoisonTick60 = 5

func (hp *HunterPet) newClaw() *core.Spell {
	baseDamageMin, baseDamageMax := petClawDamage[hp.Owner.Level][0], petClawDamage[hp.Owner.Level][1]

	spellID := map[int32]int32{
		25: 16830,
		40: 16832,
		50: 3010,
		60: 3009,
	}[hp.Owner.Level]

	return hp.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellID},
		SpellCode:   SpellCode_HunterPetClaw,
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics,

		FocusCost: core.FocusCostOptions{
			Cost: 25,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: hp.gcd(),
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageMin, baseDamageMax)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		},
	})
}

func (hp *HunterPet) newBite() *core.Spell {
	baseDamageMin, baseDamageMax := petBiteDamage[hp.Owner.Level][0], petBiteDamage[hp.Owner.Level][1]

	spellID := map[int32]int32{
		25: 17257,
		40: 17259,
		50: 17260,
		60: 17261,
	}[hp.Owner.Level]

	return hp.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellID},
		SpellCode:   SpellCode_HunterPetBite,
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics,

		FocusCost: core.FocusCostOptions{
			Cost: 35,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: hp.gcd(),
			},
			CD: core.Cooldown{
				Timer:    hp.NewTimer(),
				Duration: 10 * time.Second,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageMin, baseDamageMax)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		},
	})
}

func (hp *HunterPet) newLightningBreath() *core.Spell {
	baseDamageMin := map[int32]float64{
		25: 36,
		40: 36,
		50: 78,
		60: 99,
	}[hp.Owner.Level]

	baseDamageMax := map[int32]float64{
		25: 41,
		40: 41,
		50: 91,
		60: 113,
	}[hp.Owner.Level]

	spellID := map[int32]int32{
		25: 25009,
		40: 25009, // rank 4 not available in SoD Phase 2
		50: 25011,
		60: 25012,
	}[hp.Owner.Level]

	if hp.Env.IsForever() && hp.Owner.Level == 60 {
		baseDamageMin, baseDamageMax = foreverLightningBreath60[0], foreverLightningBreath60[1]
	}

	return hp.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellID},
		SpellCode:   SpellCode_HunterPetLightningBreath,
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,

		FocusCost: core.FocusCostOptions{
			Cost: 50,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: hp.gcd(),
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageMin, baseDamageMax)

			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})
}

func (hp *HunterPet) newScreech() *core.Spell {
	baseDamageMin := map[int32]float64{
		25: 12,
		40: 12,
		50: 19,
		60: 26,
	}[hp.Owner.Level]

	baseDamageMax := map[int32]float64{
		25: 16,
		40: 16,
		50: 25,
		60: 46,
	}[hp.Owner.Level]

	spellID := map[int32]int32{
		15: 24580,
		40: 24580,
		50: 24581,
		60: 24582,
	}[hp.Owner.Level]

	// Forever's rank 4 (client 24579, which 24582 teaches) has a 10 sec cooldown and takes the
	// attack power of every enemy within 5 yd of the pet (effect 2, radius 5) down 198 at 56 and 1.6
	// a level more, 204 at 60, for 30 sec. The reduction is the one Demoralizing Shout's takes, so
	// the strongest counts.
	var cooldown core.Cooldown
	var apReduction core.AuraArray
	if hp.Env.IsForever() && hp.Owner.Level == 60 {
		spellID = 24579
		baseDamageMin, baseDamageMax = foreverScreech60[0], foreverScreech60[1]
		cooldown = core.Cooldown{
			Timer:    hp.NewTimer(),
			Duration: time.Second * 10,
		}
		apReduction = hp.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			return core.ForeverDemoralizingScreechAura(target, 204)
		})
	}

	return hp.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellID},
		SpellCode:   SpellCode_HunterPetScreech,
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeSpecial,
		Flags:       core.SpellFlagMeleeMetrics,

		FocusCost: core.FocusCostOptions{
			Cost: 20,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: hp.gcd(),
			},
			CD:          cooldown,
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageMin, baseDamageMax)
			// Outside Forever the attack power reduction, like Demoralizing Shout's, is left out. The
			// sim's enemies all stand within 5 yd of the pet's target, as wowsims/forever has them.
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
			if apReduction != nil && result.Landed() {
				for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
					apReduction.Get(aoeTarget).Activate(sim)
				}
			}
		},
	})
}

// func (hp *HunterPet) newFuriousHowl() *core.Spell {
// 	actionID := core.ActionID{SpellID: 64495}

// 	petAura := hp.NewTemporaryStatsAura("FuriousHowl", actionID, stats.Stats{stats.AttackPower: 320, stats.RangedAttackPower: 320}, time.Second*20)
// 	ownerAura := hp.hunterOwner.NewTemporaryStatsAura("FuriousHowl", actionID, stats.Stats{stats.AttackPower: 320, stats.RangedAttackPower: 320}, time.Second*20)

// 	howlSpell := hp.RegisterSpell(core.SpellConfig{
// 		ActionID: actionID,

// 		FocusCost: core.FocusCostOptions{
// 			Cost: 20,
// 		},
// 		Cast: core.CastConfig{
// 			CD: core.Cooldown{
// 				Timer:    hp.NewTimer(),
// 				Duration: time.Second * 40,
// 			},
// 		},
// 		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
// 			return hp.IsEnabled()
// 		},
// 		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
// 			petAura.Activate(sim)
// 			ownerAura.Activate(sim)
// 		},
// 	})

// 	hp.hunterOwner.RegisterSpell(core.SpellConfig{
// 		ActionID: actionID,
// 		Flags:    core.SpellFlagAPL | core.SpellFlagMCD,
// 		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
// 			return howlSpell.CanCast(sim, target)
// 		},
// 		ApplyEffects: func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
// 			howlSpell.Cast(sim, target)
// 		},
// 	})

// 	hp.hunterOwner.AddMajorCooldown(core.MajorCooldown{
// 		Spell: howlSpell,
// 		Type:  core.CooldownTypeDPS,
// 	})

// 	return nil
// }

func (hp *HunterPet) newScorpidPoison() *core.Spell {
	baseDamageTick := map[int32]float64{
		25: 3,
		40: 6,
		50: 6,
		60: 8,
	}[hp.Owner.Level]
	if hp.Env.IsForever() && hp.Owner.Level == 60 {
		baseDamageTick = foreverScorpidPoisonTick60
	}
	spellID := map[int32]int32{
		25: 24583,
		40: 24586,
		50: 24586,
		60: 24587,
	}[hp.Owner.Level]

	return hp.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellID},
		SpellCode:   SpellCode_HunterPetScorpidPoison,
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagPassiveSpell | core.SpellFlagPoison,

		FocusCost: core.FocusCostOptions{
			Cost: 30,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: hp.gcd(),
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    hp.NewTimer(),
				Duration: time.Second * 4,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:     "ScorpidPoison",
				MaxStacks: 5,
				Duration:  time.Second * 10,
			},
			NumberOfTicks: 5,
			TickLength:    time.Second * 2,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, applyStack bool) {
				if !applyStack {
					return
				}

				// only the first stack snapshots the multiplier
				if dot.GetStacks() == 1 {
					attackTable := dot.Spell.Unit.AttackTables[target.UnitIndex][dot.Spell.CastType]
					dot.SnapshotAttackerMultiplier = dot.Spell.AttackerDamageMultiplier(attackTable, true)
					dot.SnapshotBaseDamage = 0
				}

				dot.SnapshotBaseDamage += baseDamageTick
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
			if !result.Landed() {
				return
			}

			dot := spell.Dot(target)
			dot.ApplyOrRefresh(sim)
			if dot.GetStacks() < dot.MaxStacks {
				dot.AddStack(sim)
				dot.TakeSnapshot(sim, true)
			}
		},
	})
}
