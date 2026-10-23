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

	// Summoning Disorientation: the demon out while the warlock summons another (registerSummonDemon)
	SummonStunAura *core.Aura

	LifeTapManaMetrics *core.ResourceMetrics

	manaPooling bool
}

type PetConfig struct {
	Name             string
	ManaPerIntellect float64 // mana a point of the demon's own Intellect, on top of its base Mana
	Stats            stats.Stats
	AutoAttacks      core.AutoAttackOptions

	MP5PerOwnerIntellect float64 // MP5 a point of the warlock's Intellect (makeStatInheritance)
	APPerStrength        float64 // attack power a point of the demon's Strength, 0 for a fixed attack power in Stats
	DamageMultiplier     float64 // on everything the demon deals, 0 for none
}

// A demon's own weapon is a fixed share of the creature auto attack dps for its level (ExpectedStat,
// ExpansionID -2: 60.4529 at level 60), read off naked-warlock pet sheets at levels 20 and 30 by
// wowsims/forever (pets.go). The level 60 weapons follow from those shares; they are not yet read at 60.
// The other levels keep 1.12's weapons.
const creatureAutoAttackDps60 = 60.4529

// The Sayaad's average swing before attack power: 0.698 of the creature dps at 2.0 sec, spanning 21.5%
// either side (b207517048: 22.8% at level 20, 20.7% at 30), 66 to 103 at 60 (1.12 95 to 131).
const sayaadSwingAverage = 0.698 * creatureAutoAttackDps60 * 2

// The Voidwalker's average swing before its melee is halved: 0.494 of the creature dps at 2.0 sec,
// spanning 26.4% either side (43fce5a834, the best fit to three level 30 sheets at attack power 96, 108
// and 116), 44 to 76 at 60 (1.12 31 to 46).
const voidwalkerSwingAverage = 0.494 * creatureAutoAttackDps60 * 2

// The Sayaad's and the Voidwalker's own attack power is 3 x level + 6, with none from Strength
// (b207517048: "Attack Power 103 (96 + 7)" at level 30 and 66 at 20; 43fce5a834: 96 on a level 30
// Voidwalker). Not yet read at 60, where it gives 186.
func demonAttackPower(level int32) float64 {
	return 3*float64(level) + 6
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
		// Before the dismissal, so the stun ending does not set the dismissed demon going again
		warlock.ActivePet.SummonStunAura.Deactivate(sim)
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
		Pet:          core.NewPet(cfg.Name, &warlock.Character, cfg.Stats, warlock.makeStatInheritance(cfg.MP5PerOwnerIntellect), enabledOnStart, false),
		owner:        warlock,
		OnPetDisable: func(sim *core.Simulation, isSacrifice bool) {},
	}

	// Forever's buffs are players only (client 70338 gives Gift of the Wild, Fortitude, Arcane
	// Brilliance, the Greater Blessings, Battle Shout, Leader of the Pack, Moonkin Aura, the stat
	// totems and the scrolls the players-only target flag, SPELL_ATTR3 0x100, and forever-bugs #27,
	// "Invalid Target" on a warlock pet, was closed as not a bug), so demons get none. They scale from
	// the warlock's stats instead, see makeStatInheritance. Debuffs on the target still count.
	wp.Pet.IgnoresPlayerBuffs = true

	// A demon's mana is its base Mana plus a share of its own Intellect, with none of the player's base
	// mana curve: the Imp's is 576 + 5 a point at level 60 (wowsims/forever's pet model, pets.go).
	modifier := cfg.ManaPerIntellect / 15
	wp.EnableManaBarWithModifier(modifier)
	wp.AddStat(stats.Mana, 15*20*modifier-20)

	// A demon's Spirit gives no regen: its regen is all MP5 from the warlock's Intellect, earned while
	// casting too (wowsims/forever f5b2f7ef33: at level 30 with 110 Intellect the pet panes read 33 MP5
	// for the Imp and 11 for the Voidwalker and Sayaad, nothing from the demon's own stats).
	wp.PseudoStats.SpiritRegenMultiplier = 0

	if cfg.APPerStrength > 0 {
		wp.AddStatDependency(stats.Strength, stats.AttackPower, cfg.APPerStrength)
	}
	if cfg.DamageMultiplier > 0 {
		wp.PseudoStats.DamageDealtMultiplier *= cfg.DamageMultiplier
	}

	// Imps don't melee
	if cfg.Name != "Imp" {
		wp.EnableAutoAttacks(wp, cfg.AutoAttacks)
	}

	// Summoning Disorientation (1249755): the demon out is stunned for 6 sec from the start of a summon's
	// cast. The development notes of the 2026-10-08 build say "A Warlock's current pet will now be stunned
	// while summoning a new pet", but the row has only dummy effects, so the stun is read from its name and
	// the notes, and the 6 sec from its duration (wowsims/forever d7968fbd13 does the same). A summon cast
	// in less, under Fel Domination or Master Summoner, dismisses the demon before the stun runs out.
	wp.SummonStunAura = wp.RegisterAura(core.Aura{
		Label:    "Summoning Disorientation",
		ActionID: core.ActionID{SpellID: 1249755},
		Duration: time.Second * 6,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			wp.InterruptHardcast(sim)
			wp.CancelGCDTimer(sim)
			wp.AutoAttacks.CancelAutoSwing(sim)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			if wp.IsEnabled() {
				wp.AutoAttacks.EnableAutoSwing(sim)
				wp.SetGCDTimer(sim, sim.CurrentTime)
			}
		},
	})

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
	// Here rather than at construction, which doesn't know the ruleset yet.
	core.ApplyPetConsumeEffects(&wp.Character, wp.Owner.Consumes)
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
	if !wp.IsEnabled() || wp.primaryAbility == nil || wp.SummonStunAura.IsActive() {
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

// makeStatInheritance is the Warlock Pet Scaling aura (client 70338, spell 416189, which names health,
// attack power, damage done, resistances, armor, hit, crit, haste and mana regen; Blizzard's warlock deep
// dive of 2026-10-07: "pets now scale from their summoner's stats"). Every base point is 0 in the client,
// so the shares are wowsims/forever's, read off beta pet sheets and logs (pets.go):
//   - Attack power: 0.17 and spell damage 0.10 a point of the warlock's spell damage, its generic spell
//     damage plus the better of Shadow and Fire (c3d9ea9516: at 30 spell damage the warlock's tooltip reads
//     "Increases your pet's Attack Power by 5 and Spell Damage by 3", and a Sayaad logged with 3 Spell
//     Damage at the warlock's 30 and 9 at ~96; measured on the Sayaad only).
//   - Hit: the warlock's spell hit whole for the demon's spells and 8/17 of it as melee hit (9ce48b6aba,
//     forever-bugs #394; not yet read on a pet sheet).
//   - Health: 7.5 a point of the warlock's Stamina (19922b122e, the 2026-10-13 Sayaad sheets). Armor: 35%
//     of the warlock's, Demon Armor included (7097a63bc9, a level 30 Imp's sheet).
//   - MP5: a share of the warlock's Intellect, 0.30 for the Imp and 0.10 for the others (f5b2f7ef33, level
//     30 pet panes; the Felhunter's not read).
//   - Crit: the warlock's spell crit, as the demon's melee and its spell crit, and no crit of the demon's
//     own (no base crit, none from its Agility or Intellect). Blizzard's intent: AidanZMoon on forever-bugs
//     #331 (13 to 14 October), demons "should be getting their master's spell crit as their own melee AND
//     spell crit chance", the scaling held back on the beta "until the baseline critical strike chance is
//     removed" (user ruling 2026-10-16: model Blizzard's intent, one value).
//
// notes.md lists what is still unmeasured.
func (warlock *Warlock) makeStatInheritance(mp5PerOwnerIntellect float64) core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		ownerHitChance := ownerStats[stats.SpellHit] / core.SpellHitRatingPerHitChance
		ownerCritChance := ownerStats[stats.SpellCrit] / core.SpellCritRatingPerCritChance
		spellDamage := ownerStats[stats.SpellPower] + ownerStats[stats.SpellDamage] + max(ownerStats[stats.FirePower], ownerStats[stats.ShadowPower])

		return stats.Stats{
			stats.Health:      ownerStats[stats.Stamina] * 7.5,
			stats.Armor:       ownerStats[stats.Armor] * 0.35,
			stats.MP5:         ownerStats[stats.Intellect] * mp5PerOwnerIntellect,
			stats.AttackPower: spellDamage * 0.17,
			stats.SpellDamage: spellDamage * 0.10,
			stats.SpellHit:    ownerHitChance * core.SpellHitRatingPerHitChance,
			stats.MeleeHit:    ownerHitChance * 8 / 17 * core.MeleeHitRatingPerHitChance,
			stats.SpellCrit:   ownerCritChance * core.SpellCritRatingPerCritChance,
			stats.MeleeCrit:   ownerCritChance * core.CritRatingPerCritChance,
		}
	}
}
