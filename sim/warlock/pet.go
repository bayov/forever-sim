package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

type OnPetDisable func(sim *core.Simulation, isSacrifice bool)

type WarlockPet struct {
	core.Pet

	OnPetDisable OnPetDisable

	owner *Warlock

	primaryAbility   *core.Spell
	secondaryAbility *core.Spell

	SoulLinkAura *core.Aura

	LifeTapManaMetrics *core.ResourceMetrics

	manaPooling bool
}

type PetConfig struct {
	Name          string
	PowerModifier float64 // GetUnitPowerModifier("pet")
	Stats         stats.Stats
	AutoAttacks   core.AutoAttackOptions
}

func (warlock *Warlock) setDefaultActivePet() {
	switch warlock.Options.Summon {
	case proto.WarlockOptions_Imp:
		warlock.ActivePet = warlock.Imp
	case proto.WarlockOptions_Felhunter:
		warlock.ActivePet = warlock.Felhunter
	case proto.WarlockOptions_Succubus:
		warlock.ActivePet = warlock.Succubus
	case proto.WarlockOptions_Voidwalker:
		warlock.ActivePet = warlock.Voidwalker
	}
}

func (warlock *Warlock) changeActivePet(sim *core.Simulation, newPet *WarlockPet, isSacrifice bool) {
	if warlock.ActivePet != nil {
		warlock.ActivePet.Disable(sim, isSacrifice)

		// Sacrificed pets lose all buffs
		if isSacrifice {
			for _, aura := range warlock.ActivePet.GetAuras() {
				aura.Deactivate(sim)
			}
		}

		if warlock.MasterDemonologistAura != nil {
			warlock.MasterDemonologistAura.Deactivate(sim)
		}
	}

	warlock.ActivePet = newPet

	if newPet != nil {
		newPet.Enable(sim, newPet)
	}
}

func (warlock *Warlock) registerPets() {
	warlock.Felhunter = warlock.makeFelhunter()
	warlock.Imp = warlock.makeImp()
	warlock.Succubus = warlock.makeSuccubus()
	warlock.Voidwalker = warlock.makeVoidwalker()

	warlock.BasePets = []*WarlockPet{warlock.Felhunter, warlock.Imp, warlock.Succubus, warlock.Voidwalker}
}

func (warlock *Warlock) makePet(cfg PetConfig, enabledOnStart bool) *WarlockPet {
	// No crit of the demon's own: its crit is its master's spell crit (makeStatInheritance).
	cfg.Stats[stats.MeleeCrit] = 0
	cfg.Stats[stats.SpellCrit] = 0

	wp := &WarlockPet{
		Pet:          core.NewPet(cfg.Name, &warlock.Character, cfg.Stats, warlock.makeStatInheritance(), enabledOnStart, false),
		owner:        warlock,
		OnPetDisable: func(sim *core.Simulation, isSacrifice bool) {},
	}

	// Forever's buffs are players only (client 70338 gives Gift of the Wild, Fortitude, Arcane
	// Brilliance, the Greater Blessings, Battle Shout, Leader of the Pack, Moonkin Aura, the stat
	// totems and the scrolls the players-only target flag, SPELL_ATTR3 0x100, and forever-bugs #27,
	// "Invalid Target" on a warlock pet, was closed as not a bug), so demons get none. They scale from
	// the warlock's stats instead, see makeStatInheritance. Debuffs on the target still count.
	wp.Pet.IgnoresPlayerBuffs = true

	wp.EnableManaBarWithModifier(cfg.PowerModifier)

	if cfg.Name == "Imp" {
		// Imp gets 1mp/5 non casting regen per spirit
		wp.PseudoStats.SpiritRegenMultiplier = 1
		wp.PseudoStats.SpiritRegenRateCasting = 0
		wp.SpiritManaRegenPerSecond = func() float64 {
			// 1mp5 per spirit
			return wp.GetStat(stats.Spirit) / 5
		}

	} else {
		// Warrior scaling for all other pets
		wp.AddStat(stats.AttackPower, -20)
		wp.AddStatDependency(stats.Strength, stats.AttackPower, 2)

		// Imps generally don't melee
		wp.EnableAutoAttacks(wp, cfg.AutoAttacks)
	}

	core.ApplyPetConsumeEffects(&wp.Character, warlock.Consumes)

	warlock.AddPet(wp)

	return wp
}

func (warlock *Warlock) registerPetAbilities() {
	warlock.Imp.registerImpFireboltSpell()
	warlock.Succubus.registerSuccubusLashOfPainSpell()
}

func (wp *WarlockPet) GetPet() *core.Pet {
	return &wp.Pet
}

func (wp *WarlockPet) Initialize() {
}

func (wp *WarlockPet) Reset(_ *core.Simulation) {
	wp.manaPooling = false
}

func (wp *WarlockPet) Disable(sim *core.Simulation, isSacrifice bool) {
	wp.Pet.Disable(sim)

	if wp.OnPetDisable != nil {
		wp.OnPetDisable(sim, isSacrifice)
	}
}

func (wp *WarlockPet) ApplyOnPetDisable(newOnPetDisable OnPetDisable) {
	oldOnPetDisable := wp.OnPetDisable
	if oldOnPetDisable == nil {
		wp.OnPetDisable = oldOnPetDisable
	} else {
		wp.OnPetDisable = func(sim *core.Simulation, isSacrifice bool) {
			oldOnPetDisable(sim, isSacrifice)
			newOnPetDisable(sim, isSacrifice)
		}
	}
}

func (wp *WarlockPet) ExecuteCustomRotation(sim *core.Simulation) {
	if !wp.IsEnabled() || wp.primaryAbility == nil {
		return
	}

	if wp.manaPooling {
		maxPossibleCasts := sim.GetRemainingDuration().Seconds() / wp.primaryAbility.CurCast.CastTime.Seconds()

		if wp.CurrentMana() > (maxPossibleCasts*wp.primaryAbility.CurCast.Cost)*0.75 {
			wp.manaPooling = false
			wp.WaitUntil(sim, sim.CurrentTime+10*time.Millisecond)
			return
		}

		if wp.CurrentMana() >= wp.MaxMana()*0.94 {
			wp.manaPooling = false
			wp.WaitUntil(sim, sim.CurrentTime+10*time.Millisecond)
			return
		}

		if wp.manaPooling {
			return
		}
	}

	if !wp.primaryAbility.IsReady(sim) {
		wp.WaitUntil(sim, wp.primaryAbility.CD.ReadyAt())
		return
	}

	if wp.Unit.CurrentMana() >= wp.primaryAbility.CurCast.Cost {
		wp.primaryAbility.Cast(sim, wp.CurrentTarget)
	} else if !wp.owner.Options.PetPoolMana {
		wp.manaPooling = true
	}
}

// makeStatInheritance is the Warlock Pet Scaling aura (client 70338, spell 416189, which carries the
// attack power, damage done, hit, crit and resistance auras; Blizzard's warlock deep dive of 2026-10-07:
// "pets now scale from their summoner's stats"). Forever's amounts are server side (every base point is
// 0 in the client) and unmeasured, so these are Season of Discovery's as a placeholder (the user's backup
// source, orchestration/pet_buffs.md):
//   - Hit: 100% of the warlock's spell hit (the higher of Fire and Shadow, one number in the sim) for the
//     demon's spells and 80% of it as physical hit (Blizzard's Season of Discovery hotfix, 2024-12-04).
//   - Attack power: 0.5 per point of the warlock's highest school's spell power (a Season of Discovery
//     player test, 2024-01-06).
//   - Spell damage: about 25% of that spell power (the same test; the tester was unsure of it).
//   - Stamina, Intellect and armor: the test says they are inherited but gives no shares, so none here.
//   - Crit: the warlock's spell crit, as the demon's melee and its spell crit, and no crit of the demon's
//     own (no base crit, none from its Agility or Intellect). Blizzard's intent: AidanZMoon on forever-bugs
//     #331 (13 to 14 October), demons "should be getting their master's spell crit as their own melee AND
//     spell crit chance", the scaling held back on the beta "until the baseline critical strike chance is
//     removed" (user ruling 2026-10-16: model Blizzard's intent, one value).
//
// notes.md lists it under Need to Verify.
func (warlock *Warlock) makeStatInheritance() core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		ownerHitChance := ownerStats[stats.SpellHit] / core.SpellHitRatingPerHitChance
		ownerCritChance := ownerStats[stats.SpellCrit] / core.SpellCritRatingPerCritChance
		highestSchoolPower := ownerStats[stats.SpellPower] + ownerStats[stats.SpellDamage] + max(ownerStats[stats.FirePower], ownerStats[stats.ShadowPower])

		return stats.Stats{
			stats.AttackPower: highestSchoolPower * 0.5,
			stats.SpellDamage: highestSchoolPower * 0.25,
			stats.SpellHit:    ownerHitChance * core.SpellHitRatingPerHitChance,
			stats.MeleeHit:    ownerHitChance * 0.8 * core.MeleeHitRatingPerHitChance,
			stats.SpellCrit:   ownerCritChance * core.SpellCritRatingPerCritChance,
			stats.MeleeCrit:   ownerCritChance * core.CritRatingPerCritChance,
		}
	}
}
