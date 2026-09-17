package core

import (
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever racials, as read from the beta client tooltips (build 1.60.1.69876, see
// forever-wiki/racials.md). Every race has two actives and two passives, the resistance
// racials are gone and the weapon skill racials pay crit instead.
//
// Cooldowns per the client: Blood Fury, Eureka! and the Undead and Gnome utilities 2 min,
// Berserking, Elune's Light and the Human, Dwarf and Orc utilities 3 min.

func applyForeverRaceEffects(agent Agent) {
	character := agent.GetCharacter()

	switch character.Race {
	case proto.Race_RaceHuman:
		// The Human Spirit. Sword Specialization is crit now, and Mace Specialization
		// moved to the Dwarves.
		character.MultiplyStat(stats.Spirit, 1.05)
		character.AddWeaponSpecializationCrit(2, proto.WeaponType_WeaponTypeSword)

	case proto.Race_RaceDwarf:
		// Mace Specialization and Big Game Hunter. Stoneform is a defensive cooldown the
		// sim does not measure.
		character.AddWeaponSpecializationCrit(1, proto.WeaponType_WeaponTypeMace)
		character.mobTypeDamageAura(proto.MobType_MobTypeBeast, 1.05)

	case proto.Race_RaceNightElf:
		// Quickness keeps its dodge. Elune's Light is the new offensive cooldown.
		character.AddStat(stats.Dodge, 1)
		character.registerElunesLight()

	case proto.Race_RaceGnome:
		// Expansive Mind raises the resource pool itself now rather than Intellect.
		if character.HasManaBar() {
			character.MultiplyStat(stats.Mana, 1.05)
		}
		if character.HasRageBar() {
			character.MultiplyMaxRage(1.05)
		}
		if character.HasEnergyBar() {
			character.MultiplyMaxEnergy(1.05)
		}
		character.registerEureka()

	case proto.Race_RaceOrc:
		// Axe Specialization is crit now. Command (pet damage) is gone.
		character.AddWeaponSpecializationCrit(1, proto.WeaponType_WeaponTypeAxe)
		character.registerForeverBloodFury()

	case proto.Race_RaceUndead:
		character.registerTouchOfTheGrave()

	case proto.Race_RaceTauren:
		// Endurance carries a point of hit alongside the health.
		character.MultiplyStat(stats.Health, 1.05)
		character.AddStat(stats.MeleeHit, 1*MeleeHitRatingPerHitChance)
		character.AddStat(stats.SpellHit, 1*SpellHitRatingPerHitChance)

	case proto.Race_RaceTroll:
		// Beast Slaying stays. Bow and Throwing Specialization are gone, and Berserking
		// is a flat 10% instead of scaling with missing health.
		character.mobTypeDamageAura(proto.MobType_MobTypeBeast, 1.05)
		makeBerserkingCooldown(character, .1, character.NewTimer())

	case proto.Race_RaceSkyborneHighOrder, proto.Race_RaceSkyborneWindshaper:
		// Wind Blessed and Elemental Insight. The actives (Walk on Air, Read Ley Line,
		// Skysight) are movement and regeneration, so the two halves sim the same.
		character.PseudoStats.MeleeSpeedMultiplier *= 1.01
		character.PseudoStats.RangedSpeedMultiplier *= 1.01
		character.PseudoStats.CastSpeedMultiplier *= 1.01
		character.mobTypeDamageAura(proto.MobType_MobTypeElemental, 1.05)
	}
}

// Damage against one creature type, like Troll Beast Slaying. Applied to the attack
// tables after finalize because the targets are not known before then.
func (character *Character) mobTypeDamageAura(mobType proto.MobType, multiplier float64) {
	character.Env.RegisterPostFinalizeEffect(func() {
		for _, t := range character.Env.Encounter.Targets {
			if t.MobType == mobType {
				for _, at := range character.AttackTables[t.UnitIndex] {
					at.DamageDealtMultiplier *= multiplier
					at.CritMultiplier *= multiplier
				}
			}
		}
	})
}

// Blood Fury: Attack Power and Spell Power increased by 10% for 15 sec. The Classic
// healing penalty is gone. The 10% is taken from the character's total AP and SP at
// activation, which includes gear and buffs.
func (character *Character) registerForeverBloodFury() {
	actionID := ActionID{SpellID: 20572}

	var bonusAP, bonusSP float64
	aura := character.RegisterAura(Aura{
		Label:    "Blood Fury",
		ActionID: actionID,
		Duration: time.Second * 15,
		OnGain: func(aura *Aura, sim *Simulation) {
			bonusAP = character.GetStat(stats.AttackPower) * 0.1
			bonusSP = character.GetStat(stats.SpellPower) * 0.1
			character.AddStatDynamic(sim, stats.AttackPower, bonusAP)
			character.AddStatDynamic(sim, stats.SpellPower, bonusSP)
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			character.AddStatDynamic(sim, stats.AttackPower, -bonusAP)
			character.AddStatDynamic(sim, stats.SpellPower, -bonusSP)
		},
	})

	spell := character.RegisterSpell(SpellConfig{
		ActionID: actionID,
		Flags:    SpellFlagNoOnCastComplete,
		Cast: CastConfig{
			DefaultCast: Cast{
				GCD: GCDDefault,
			},
			CD: Cooldown{
				Timer:    character.NewTimer(),
				Duration: time.Minute * 2,
			},
		},
		ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
			aura.Activate(sim)
		},
	})

	character.AddMajorCooldown(MajorCooldown{
		Spell: spell,
		Type:  CooldownTypeDPS,
	})
}

// Elune's Light: critical strike chance increased by 10% for 15 sec, 3 min cooldown.
func (character *Character) registerElunesLight() {
	actionID := ActionID{SpellID: 1259799}

	aura := character.NewTemporaryStatsAura("Elune's Light", actionID, stats.Stats{
		stats.MeleeCrit: 10 * CritRatingPerCritChance,
		stats.SpellCrit: 10 * SpellCritRatingPerCritChance,
	}, time.Second*15)

	spell := character.RegisterSpell(SpellConfig{
		ActionID: actionID,
		Flags:    SpellFlagNoOnCastComplete,
		Cast: CastConfig{
			CD: Cooldown{
				Timer:    character.NewTimer(),
				Duration: time.Minute * 3,
			},
		},
		ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
			aura.Activate(sim)
		},
	})

	character.AddMajorCooldown(MajorCooldown{
		Spell: spell,
		Type:  CooldownTypeDPS,
	})
}

// Eureka!: the next 3 damaging abilities cost less and deal 10% more damage. The beta client
// gives each class its own discount: Energy 20%, Rage 40%, Mana 50% (Priests 15%, and their
// healing counts too). 2 min cooldown.
//
// A damaging ability is one the rotation can cast, that has a cost and that is set up to
// deal damage. Slice and Dice, Expose Armor or a shout have no damage multiplier and are
// left alone. Auto attacks are not abilities. A stack is spent on each cast.
func (character *Character) eurekaCostReduction() int32 {
	switch {
	case character.HasEnergyBar():
		return 20
	case character.HasRageBar():
		return 40
	case character.Class == proto.Class_ClassPriest:
		return 15
	}
	return 50
}

func (character *Character) registerEureka() {
	actionID := ActionID{SpellID: 1259812}
	costReduction := character.eurekaCostReduction()
	healingCounts := character.Class == proto.Class_ClassPriest

	affects := func(spell *Spell) bool {
		if !spell.Flags.Matches(SpellFlagAPL) || spell.Cost == nil {
			return false
		}
		if spell.Flags.Matches(SpellFlagHelpful) {
			return healingCounts
		}
		return spell.DamageMultiplier != 0
	}

	var affected []*Spell
	aura := character.RegisterAura(Aura{
		Label:     "Eureka!",
		ActionID:  actionID,
		Duration:  time.Second * 30,
		MaxStacks: 3,
		OnInit: func(aura *Aura, sim *Simulation) {
			for _, spell := range character.Spellbook {
				if affects(spell) {
					affected = append(affected, spell)
				}
			}
		},
		OnGain: func(aura *Aura, sim *Simulation) {
			for _, spell := range affected {
				spell.DamageMultiplier *= 1.1
				spell.Cost.Multiplier -= costReduction
			}
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			for _, spell := range affected {
				spell.DamageMultiplier /= 1.1
				spell.Cost.Multiplier += costReduction
			}
		},
		OnStacksChange: func(aura *Aura, sim *Simulation, oldStacks, newStacks int32) {
			if newStacks == 0 {
				aura.Deactivate(sim)
			}
		},
		OnCastComplete: func(aura *Aura, sim *Simulation, spell *Spell) {
			if aura.GetStacks() > 0 && affects(spell) {
				aura.RemoveStack(sim)
			}
		},
	})

	spell := character.RegisterSpell(SpellConfig{
		ActionID: actionID,
		Flags:    SpellFlagNoOnCastComplete,
		Cast: CastConfig{
			CD: Cooldown{
				Timer:    character.NewTimer(),
				Duration: time.Minute * 2,
			},
		},
		ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
			aura.Activate(sim)
			aura.SetStacks(sim, aura.MaxStacks)
		},
	})

	character.AddMajorCooldown(MajorCooldown{
		Spell: spell,
		Type:  CooldownTypeDPS,
		ShouldActivate: func(sim *Simulation, character *Character) bool {
			if character.NextAbilityBuffCondition == nil {
				return true
			}
			return character.NextAbilityBuffCondition(sim, character)
		},
	})
}

// Touch of the Grave: spells and attacks have a 5% chance (10% for Mages, Priests and
// Warlocks) to drain health from the target, up to 5% of the caster's maximum health.
//
// "Up to" reads as a cap, and the amount under it is not published. The drain is modelled
// at the cap, so this is the upper bound of what the racial can be worth. It is Shadow
// damage that cannot crit, cannot proc anything and heals the caster for what it deals.
//
// The tooltip has no internal cooldown either. A dual wielding rogue lands enough hits
// that the cooldown decides whether the racial is worth 1% or 2%, so it is a variable
// that a comparison harness can sweep. Zero means every hit rolls.
var TouchOfTheGraveICD time.Duration = 0

func (character *Character) registerTouchOfTheGrave() {
	actionID := ActionID{SpellID: 1260189}
	procChance := 0.05
	switch character.Class {
	case proto.Class_ClassMage, proto.Class_ClassPriest, proto.Class_ClassWarlock:
		actionID = ActionID{SpellID: 1260201}
		procChance = 0.10
	}
	healthMetrics := character.NewHealthMetrics(actionID)

	drain := character.RegisterSpell(SpellConfig{
		ActionID:    actionID,
		SpellSchool: SpellSchoolShadow,
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskEmpty,
		Flags:       SpellFlagPassiveSpell | SpellFlagNoOnCastComplete | SpellFlagBinary,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			amount := character.MaxHealth() * 0.05
			result := spell.CalcAndDealDamage(sim, target, amount, spell.OutcomeMagicHit)
			if result.Landed() {
				character.GainHealth(sim, result.Damage, healthMetrics)
			}
		},
	})

	MakeProcTriggerAura(&character.Unit, ProcTrigger{
		Name:       "Touch of the Grave",
		Callback:   CallbackOnSpellHitDealt,
		ProcMask:   ProcMaskDirect,
		Outcome:    OutcomeLanded,
		ProcChance: procChance,
		ICD:        TouchOfTheGraveICD,
		Handler: func(sim *Simulation, spell *Spell, result *SpellResult) {
			drain.Cast(sim, result.Target)
		},
	})
}
