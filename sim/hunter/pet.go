package hunter

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

type HunterPet struct {
	core.Pet

	config PetConfig

	hunterOwner *Hunter

	specialAbility *core.Spell
	focusDump      *core.Spell

	uptimePercent    float64
	hasOwnerCooldown bool
}

func (hunter *Hunter) NewHunterPet() *HunterPet {
	if hunter.Options.PetType == proto.Hunter_Options_PetNone {
		return nil
	}
	if hunter.Options.PetUptime <= 0 {
		return nil
	}
	petConfig := PetConfigs[hunter.Options.PetType]

	hunterPetBaseStats := stats.Stats{}

	baseMinDamage := 0.0
	baseMaxDamage := 0.0
	attackSpeed := 2.0

	switch hunter.Options.PetAttackSpeed {
	case proto.Hunter_Options_One:
		attackSpeed = 1.0
	case proto.Hunter_Options_OneTwo:
		attackSpeed = 1.2
	case proto.Hunter_Options_OneThree:
		attackSpeed = 1.3
	case proto.Hunter_Options_OneFour:
		attackSpeed = 1.4
	case proto.Hunter_Options_OneFive:
		attackSpeed = 1.5
	case proto.Hunter_Options_OneSix:
		attackSpeed = 1.6
	case proto.Hunter_Options_OneSeven:
		attackSpeed = 1.7
	case proto.Hunter_Options_Two:
		attackSpeed = 2
	case proto.Hunter_Options_TwoFour:
		attackSpeed = 2.4
	case proto.Hunter_Options_TwoFive:
		attackSpeed = 2.5
	}

	// Classic's pet weapon and stats. Forever's weapon and attack power are set in applyForeverRules.
	baseMinDamage = 18.17 * attackSpeed
	baseMaxDamage = 27.66 * attackSpeed

	hunterPetBaseStats = stats.Stats{
		stats.Strength:  136,
		stats.Agility:   100,
		stats.Stamina:   274,
		stats.Intellect: 50,
		stats.Spirit:    80,

		stats.AttackPower: -20,
	}

	hp := &HunterPet{
		Pet:         core.NewPet(petConfig.Name, &hunter.Character, hunterPetBaseStats, hunter.makeStatInheritance(), true, false),
		config:      petConfig,
		hunterOwner: hunter,

		hasOwnerCooldown: petConfig.SpecialAbility == FuriousHowl,
	}

	hp.Pet.MobType = petConfig.MobType
	// Blizzard's hunter deep dive: under Forever pets no longer take player buffs that raise their
	// stats. Set here, the buffs go on before Initialize.
	hp.IgnoresPlayerBuffs = true

	hp.EnableAutoAttacks(hp, core.AutoAttackOptions{
		MainHand: core.Weapon{
			BaseDamageMin: baseMinDamage,
			BaseDamageMax: baseMaxDamage,
			SwingSpeed:    attackSpeed,
		},
		AutoSwingMelee: true,
	})

	// Happiness
	hp.PseudoStats.DamageDealtMultiplier *= 1.25

	// Family scalars
	hp.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= hp.config.Damage
	hp.PseudoStats.ArmorMultiplier *= hp.config.Armor
	hp.MultiplyStat(stats.Health, hp.config.Health)

	hunter.AddPet(hp)

	return hp
}

func (hp *HunterPet) GetPet() *core.Pet {
	return &hp.Pet
}

func (hp *HunterPet) Initialize() {
	// Here rather than at construction, which doesn't know the ruleset yet.
	core.ApplyPetConsumeEffects(&hp.Character, hp.hunterOwner.Consumes)

	hp.specialAbility = hp.NewPetAbility(hp.config.SpecialAbility, true)
	hp.focusDump = hp.NewPetAbility(hp.config.FocusDump, false)

	// Bestial Discipline: 10% focus regen a rank.
	hp.EnableFocusBar(1+0.1*float64(hp.hunterOwner.Talents.BestialDiscipline), func(sim *core.Simulation) {
		if hp.GCD.IsReady(sim) {
			hp.OnGCDReady(sim)
		}
	})

	if hp.Env.IsForever() {
		hp.applyForeverRules()
	} else {
		// Classic's pet: warrior attack power and crit scaling.
		hp.AddStatDependency(stats.Strength, stats.AttackPower, 2)
		hp.AddStatDependency(stats.Agility, stats.MeleeCrit, core.CritPerAgiAtLevel[proto.Class_ClassWarrior]*core.CritRatingPerCritChance)
		hp.AddStatDependency(stats.Intellect, stats.SpellCrit, core.CritPerIntAtLevel[proto.Class_ClassWarrior]*core.SpellCritRatingPerCritChance)
	}
}

// ForeverPetHitScalesWithSpeed: under Forever a pet's white hit scales with its attack speed, as in
// Classic (report assumption hunter_pet_speed, user ruling 2026-10-15). False gives the beta's
// behaviour, where every pet hits like a 2.0 sec pet.
const ForeverPetHitScalesWithSpeed = true

// The creature auto attack dps for level 60 in the client's ExpectedStat table (ExpansionID -2).
const creatureAutoAttackDps60 = 60.4529

// A Forever pet's swing before attack power, at 2.0 sec: the creature dps for its level over the
// swing, 20% either side, 97 to 145 at 60. Its own attack power is 30% of the creature dps as attack
// power (14 a point of dps), 254 at 60, with nothing from Strength. wowsims/forever fitted both to
// beta pet sheets (643d636345): a level 11 Carrion Bird's 20-27 at 44 attack power and 2.0 sec, a
// level 17 Cat's 28-37 at 75 attack power and 1.5 sec, and attack power of 36, 61 and 103 at levels
// 11, 17 and 30, each the creature dps x 4.2 beside the hunter's 10%.
const (
	foreverPetSwingAverage = creatureAutoAttackDps60 * 2
	foreverPetAttackPower  = 0.3 * 14 * creatureAutoAttackDps60
)

// What Forever changed about pets.
func (hp *HunterPet) applyForeverRules() {
	// Every pet swings every 2.0 sec, and a passive the beast brings (Faster Attack I to VII, Slower
	// Attack II and III) moves that to the beast's Classic speed. On the beta its hits stay the size
	// of a 2.0 sec pet's, so a fast pet deals more damage (beta players, gathered on ForeverChanges'
	// hunter pets page and the forever-hunter wiki).
	//
	// Report assumption hunter_pet_speed, user ruling 2026-10-15: the beta's behaviour is treated as
	// a bug, so the hit scales with the swing as in Classic (a 1.0 sec pet's hit is half a 2.0 sec
	// pet's), and every pet's white damage a second is the same.
	hitSeconds := hp.AutoAttacks.MH().SwingSpeed
	if !ForeverPetHitScalesWithSpeed {
		hitSeconds = 2
	}
	mh := hp.AutoAttacks.MH()
	mh.BaseDamageMin = foreverPetSwingAverage * (1 - 0.2) * hitSeconds / 2
	mh.BaseDamageMax = foreverPetSwingAverage * (1 + 0.2) * hitSeconds / 2

	// The pet's attack power is its own fixed share and the hunter's 10% (makeStatInheritance), with
	// none from Strength; Classic's 2 a point less 20 is not added (Initialize). Its crit is all the
	// hunter's, with none from Agility: wowsims/forever's Ragefire Chasm logs of a level 30 pet
	// (e56aaeb1a9) read the owner's sheet crit plus the pet's skill over the mobs' defense and nothing
	// of the pet's own (a staff at 1 skill and a bow at 150: pet 3.5% in 964 swings, a knife and a bow
	// at 150 with 5.96% on the sheet: 9.6% in 2,516).
	hp.AddStat(stats.AttackPower, 20+foreverPetAttackPower)

	// Focus comes back 10 a second, continuously (client PowerType Focus RegenCombat 10, and beta
	// logs where busy pets spend 9.5+ a second). Classic's is 26.25 every 5.25 sec.
	hp.SetFocusRegen(2.5, time.Millisecond*250)
}

func (hp *HunterPet) Reset(_ *core.Simulation) {
	hp.uptimePercent = min(1, max(0, hp.hunterOwner.Options.PetUptime))
}

func (hp *HunterPet) ExecuteCustomRotation(sim *core.Simulation) {
	percentRemaining := sim.GetRemainingDurationPercent()
	if percentRemaining < 1.0-hp.uptimePercent { // once fight is % completed, disable pet.
		hp.Disable(sim)
		return
	}

	if hp.hasOwnerCooldown && hp.CurrentFocus() < 50 {
		// When a major ability (Furious Howl or Savage Rend) is ready, pool enough
		// energy to use on-demand.
		return
	}

	target := hp.CurrentTarget

	// using Cast() directly is very expensive, since cast failures are logged, involving string operations
	tryCast := func(spell *core.Spell) bool {
		if !spell.CanCast(sim, target) {
			return false
		}
		if !spell.Cast(sim, target) {
			panic(fmt.Sprintf("Cast failed after CanCast() for spell %d", spell.SpellID))
		}
		return true
	}

	if hp.focusDump == nil {
		if !tryCast(hp.specialAbility) && hp.GCD.IsReady(sim) {
			hp.WaitUntil(sim, sim.CurrentTime+time.Millisecond*500)
		}
		return
	}
	if hp.specialAbility == nil {
		if !tryCast(hp.focusDump) && hp.GCD.IsReady(sim) {
			hp.WaitUntil(sim, sim.CurrentTime+time.Millisecond*500)
		}
		return
	}

	if hp.config.CustomRotation != nil {
		hp.config.CustomRotation(sim, hp, tryCast)
	} else {
		if hp.specialAbility.IsReady(sim) {
			if !tryCast(hp.specialAbility) && hp.GCD.IsReady(sim) {
				hp.WaitUntil(sim, sim.CurrentTime+time.Millisecond*500)
			}
		} else if hp.focusDump.IsReady(sim) {
			if !tryCast(hp.focusDump) && hp.GCD.IsReady(sim) {
				hp.WaitUntil(sim, sim.CurrentTime+time.Millisecond*500)
			}
		}
	}
}

// Forever's pets scale with the hunter's gear (Blizzard's hunter deep dive). The numbers are beta
// testing on the forever-hunter wiki: 10% of the hunter's highest attack power and all of the
// hunter's crit chance (and 2 health a Stamina point and 30% of armor, which change no damage). The
// hunter's ranged attack power is the highest at the pull in every build here (Survival's 1619
// against 1537 melee), so the pet takes 10% of that. The inheritance also runs on every stat change
// in the fight, so it has to stay linear: a max() of the two would keep each proc's gain and
// drop its loss.
// The crit is the higher of the hunter's melee and ranged crit (wowsims/forever 684209eb0a, after
// e56aaeb1a9's Ragefire Chasm logs). Here both are the hunter's MeleeCrit: the only ranged-only crit
// is a per-spell bonus, and the only one a hunter can have is Elemental Sharpening Stone's -2%, which
// leaves melee the higher. Their 0.04% off a point of weapon skill below the level's maximum is
// always 0 here, where every weapon skill is at the maximum.
// The pet takes the hunter's ranged hit as melee hit and 17/8 of it as spell hit, so Lightning Breath
// doesn't roll from 0% hit (forever-bugs #394, wowsims/forever 9ce48b6aba). The hunter's ranged hit
// is its MeleeHit; Biznick's Accurascope's ranged-only 3% doesn't reach the pet.
// Classic pets inherit nothing.
func (hunter *Hunter) makeStatInheritance() core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		if !hunter.Env.IsForever() {
			return stats.Stats{}
		}
		hitPercent := ownerStats[stats.MeleeHit] / core.MeleeHitRatingPerHitChance
		return stats.Stats{
			stats.AttackPower: 0.1 * ownerStats[stats.RangedAttackPower],
			stats.MeleeCrit:   ownerStats[stats.MeleeCrit],
			stats.MeleeHit:    ownerStats[stats.MeleeHit],
			stats.SpellHit:    hitPercent * 17 / 8 * core.SpellHitRatingPerHitChance,
			// The 15 October dev notes: "Pets now correctly benefit from Expertise." No hunter gear
			// or talent gives any, so this is worth nothing today.
			stats.Expertise: ownerStats[stats.Expertise],
		}
	}
}

type PetConfig struct {
	Name    string
	MobType proto.MobType

	SpecialAbility PetAbilityType
	FocusDump      PetAbilityType

	Health float64
	Armor  float64
	Damage float64

	CustomRotation func(*core.Simulation, *HunterPet, func(*core.Spell) bool)
}

// Abilities reference: https://wotlk.wowhead.com/hunter-pets
// https://wotlk.wowhead.com/guides/hunter-dps-best-pets-taming-loyalty-burning-crusade-classic
var PetConfigs = map[proto.Hunter_Options_PetType]PetConfig{
	proto.Hunter_Options_Cat: {
		Name:    "Cat",
		MobType: proto.MobType_MobTypeBeast,

		SpecialAbility: Bite,
		FocusDump:      Claw,

		Health: 0.98,
		Armor:  1.00,
		Damage: 1.10,

		CustomRotation: func(sim *core.Simulation, hp *HunterPet, tryCast func(*core.Spell) bool) {
			// Forever's 10 focus a second always pays for Bite on cooldown.
			if hp.specialAbility.CD.IsReady(sim) && (sim.IsForever() || hp.CurrentFocusPerSecond() > hp.focusDump.Cost.BaseCost/1.6) {
				if !tryCast(hp.specialAbility) && hp.GCD.IsReady(sim) {
					hp.WaitUntil(sim, sim.CurrentTime+time.Millisecond*500)
				}
			} else {
				if !tryCast(hp.focusDump) && hp.GCD.IsReady(sim) {
					hp.WaitUntil(sim, sim.CurrentTime+time.Millisecond*500)
				}
			}
		},
	},
	proto.Hunter_Options_WindSerpent: {
		Name:    "Wind Serpent",
		MobType: proto.MobType_MobTypeBeast,

		SpecialAbility: Bite,
		FocusDump:      LightningBreath,

		Health: 1.00,
		Armor:  1.00,
		Damage: 1.07,
	},
	proto.Hunter_Options_Bat: {
		Name:    "Bat",
		MobType: proto.MobType_MobTypeBeast,

		SpecialAbility: Bite,
		FocusDump:      Screech,

		Health: 1.00,
		Armor:  1.00,
		Damage: 1.07,

		CustomRotation: func(sim *core.Simulation, hp *HunterPet, tryCast func(*core.Spell) bool) {
			if hp.specialAbility.CD.IsReady(sim) {
				if !tryCast(hp.specialAbility) && hp.GCD.IsReady(sim) {
					hp.WaitUntil(sim, sim.CurrentTime+time.Millisecond*500)
				}
			} else {
				if !tryCast(hp.focusDump) && hp.GCD.IsReady(sim) {
					hp.WaitUntil(sim, sim.CurrentTime+time.Millisecond*500)
				}
			}
		},
	},
	proto.Hunter_Options_Bear: {
		Name:    "Bear",
		MobType: proto.MobType_MobTypeBeast,

		SpecialAbility: Bite,
		FocusDump:      Claw,

		Health: 1.08,
		Armor:  1.05,
		Damage: 0.91,
	},
	proto.Hunter_Options_Boar: {
		Name:    "Boar",
		MobType: proto.MobType_MobTypeBeast,

		//SpecialAbility: Gore,
		FocusDump: Bite,

		Health: 1.04,
		Armor:  1.09,
		Damage: 0.90,
	},
	proto.Hunter_Options_CarrionBird: {
		Name:    "Carrion Bird",
		MobType: proto.MobType_MobTypeBeast,

		SpecialAbility: Bite, // Screech
		FocusDump:      Claw,

		Health: 1.00,
		Armor:  1.05,
		Damage: 1.00,
	},
	proto.Hunter_Options_Owl: {
		Name:    "Owl",
		MobType: proto.MobType_MobTypeBeast,

		//SpecialAbility: Screech,
		FocusDump: Claw,

		Health: 1.00,
		Armor:  1.00,
		Damage: 1.07,
	},
	proto.Hunter_Options_Crab: {
		Name:    "Crab",
		MobType: proto.MobType_MobTypeBeast,

		FocusDump: Claw,

		Health: 0.96,
		Armor:  1.13,
		Damage: 0.95,
	},
	proto.Hunter_Options_Crocolisk: {
		Name:    "Crocolisk",
		MobType: proto.MobType_MobTypeBeast,

		FocusDump: Bite,

		Health: 0.95,
		Armor:  1.10,
		Damage: 1.00,
	},
	proto.Hunter_Options_Gorilla: {
		Name:    "Gorilla",
		MobType: proto.MobType_MobTypeBeast,

		// SpecialAbility: Thunderstomp,
		FocusDump: Bite,

		Health: 1.04,
		Armor:  1.00,
		Damage: 1.02,
	},
	proto.Hunter_Options_Hyena: {
		Name:    "Hyena",
		MobType: proto.MobType_MobTypeBeast,

		FocusDump: Bite,

		Health: 1.00,
		Armor:  1.05,
		Damage: 1.00,
	},
	proto.Hunter_Options_Raptor: {
		Name:    "Raptor",
		MobType: proto.MobType_MobTypeBeast,

		SpecialAbility: Bite,
		FocusDump:      Claw,

		Health: 0.95,
		Armor:  1.03,
		Damage: 1.10,
	},
	proto.Hunter_Options_Scorpid: {
		Name:    "Scorpid",
		MobType: proto.MobType_MobTypeBeast,

		SpecialAbility: ScorpidPoison,
		FocusDump:      Claw,

		Health: 1.00,
		Armor:  1.10,
		Damage: 0.94,

		CustomRotation: func(sim *core.Simulation, hp *HunterPet, tryCast func(*core.Spell) bool) {
			target := hp.CurrentTarget

			if (hp.specialAbility.Dot(target).GetStacks() < hp.specialAbility.Dot(target).MaxStacks || hp.specialAbility.Dot(target).RemainingDuration(sim) < time.Second*3) && hp.CurrentFocus() < 90 {
				if !tryCast(hp.specialAbility) && hp.GCD.IsReady(sim) {
					hp.WaitUntil(sim, sim.CurrentTime+time.Millisecond*500)
				}
			} else {
				if !tryCast(hp.focusDump) && hp.GCD.IsReady(sim) {
					hp.WaitUntil(sim, sim.CurrentTime+time.Millisecond*500)
				}
			}
		},
	},
	proto.Hunter_Options_Spider: {
		Name:    "Spider",
		MobType: proto.MobType_MobTypeBeast,

		FocusDump: Bite,

		Health: 1.00,
		Armor:  1.00,
		Damage: 1.07,
	},
	proto.Hunter_Options_Tallstrider: {
		Name:    "Tallstrider",
		MobType: proto.MobType_MobTypeBeast,

		FocusDump: Bite,

		Health: 1.05,
		Armor:  1.00,
		Damage: 1.00,
	},
	proto.Hunter_Options_Turtle: {
		Name:    "Turtle",
		MobType: proto.MobType_MobTypeBeast,

		// SpecialAbility: ShellShield,
		FocusDump: Bite,

		Health: 1.00,
		Armor:  1.13,
		Damage: 0.90,
	},
	proto.Hunter_Options_Wolf: {
		Name:    "Wolf",
		MobType: proto.MobType_MobTypeBeast,

		// SpecialAbility: FuriousHowl,
		FocusDump: Bite,

		Health: 1.00,
		Armor:  1.05,
		Damage: 1.00,
	},
}
