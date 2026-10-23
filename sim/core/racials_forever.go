package core

import (
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever racials, as read from the beta client tooltips (build 1.60.1.69876, see
// forever-wiki/racials.md) and checked against client 1.60.1.70425. Every race has two actives
// and two passives, and the resistance racials are gone. The weapon skill racials paid crit with
// their weapon type until 70425, which dropped every weapon requirement: Human Ruthless and Orc
// Savage give crit with any weapon, and Dwarf Earth Fury gives crit damage instead.
//
// Cooldowns per the client: Eureka! and the Undead and Gnome utilities 2 min, Blood Fury
// (since client 70425), Berserking, Elune's Light and the Human, Dwarf and Orc utilities 3 min.

func applyForeverRaceEffects(agent Agent) {
	character := agent.GetCharacter()

	switch character.Race {
	case proto.Race_RaceHuman:
		// The Human Spirit, and Ruthless (20597): 2% crit with all spells and attacks. Client 70338
		// called it Sword Specialization and gave the 2% only with a sword or two-handed sword
		// (SpellEquippedItems subclass mask 384); 70425 renamed it and dropped the weapon rule.
		character.MultiplyStat(stats.Spirit, 1.05)
		character.AddStat(stats.MeleeCrit, 2*CritRatingPerCritChance)
		character.AddStat(stats.SpellCrit, 2*SpellCritRatingPerCritChance)

	case proto.Race_RaceDwarf:
		// Earth Fury (1259719): "Increases your critical strike bonus to damage and healing by
		// 2%" (auras 163 and 50, 2 each). Client 70338 called it Mace Specialization, 1% crit with
		// all spells and attacks with a mace or two-handed mace; 70425 replaced it. And Big Game
		// Hunter. Stoneform is a defensive cooldown the sim does not measure.
		character.critDamageBonus(1.02)
		character.mobTypeDamageAura(proto.MobType_MobTypeBeast, 1.05)

	case proto.Race_RaceNightElf:
		// Quickness keeps its dodge. Elune's Light is the new offensive cooldown.
		character.AddStat(stats.Dodge, 1)
		character.registerElunesLight()
		character.registerShadowmeld()

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
		// Savage: 1% crit with all spells and abilities, whatever the weapon. In client 70338 it was
		// Axe Specialization, the same 1% only with an axe or a two-handed axe equipped
		// (SpellEquippedItems 255488, axes only); client 70425 renamed spell 20574 and dropped
		// the weapon requirement. Command (pet damage) is gone.
		character.AddStat(stats.MeleeCrit, 1*CritRatingPerCritChance)
		character.AddStat(stats.SpellCrit, 1*SpellCritRatingPerCritChance)
		character.registerForeverBloodFury()

	case proto.Race_RaceUndead:
		character.registerTouchOfTheGrave()

	case proto.Race_RaceTauren:
		// Endurance (20550): 5% health and, since client 70425, "critical strike bonus to damage
		// and healing increased by 2%" (auras 163 and 50) in place of 70338's 1% hit.
		character.MultiplyStat(stats.Health, 1.05)
		character.critDamageBonus(1.02)

	case proto.Race_RaceTroll:
		// Beast Slaying stays. Bow and Throwing Specialization are gone, and Berserking
		// is a flat 10% instead of scaling with missing health.
		character.mobTypeDamageAura(proto.MobType_MobTypeBeast, 1.05)
		makeBerserkingCooldown(character, .1, character.NewTimer())

	case proto.Race_RaceSkyborneHighOrder, proto.Race_RaceSkyborneWindshaper:
		// Wind Blessed (1259710: 3% spellcasting, melee and ranged haste since client 70425, 1% in
		// 70338) and Elemental Insight. The actives (Walk on Air, Read Leyline, Skysight) are
		// movement and regeneration, so the two halves sim the same.
		character.PseudoStats.MeleeSpeedMultiplier *= 1.03
		character.PseudoStats.RangedSpeedMultiplier *= 1.03
		character.PseudoStats.CastSpeedMultiplier *= 1.03
		character.mobTypeDamageAura(proto.MobType_MobTypeElemental, 1.05)
	}
}

// critDamageBonus multiplies the damage of every critical strike, melee, ranged and spell, like
// the 2% of Dwarf Earth Fury and Tauren Endurance (aura 163, "critical strike bonus to damage"):
// a 200% melee crit does 204% and a 150% spell crit 153%. Applied to the attack tables after
// finalize because the targets are not known before then.
func (character *Character) critDamageBonus(multiplier float64) {
	character.Env.RegisterPostFinalizeEffect(func() {
		for _, t := range character.Env.Encounter.Targets {
			for _, at := range character.AttackTables[t.UnitIndex] {
				at.CritMultiplier *= multiplier
			}
		}
	})
}

// Damage against one creature type, like Troll Beast Slaying. Applied to the attack
// tables after finalize because the targets are not known before then. The client gives Beast
// Slaying, Big Game Hunter and Elemental Insight aura 168 at 5% (damage done against the creature
// type), a plain damage aura: a crit doubles the bigger hit and gets no 5% of its own.
func (character *Character) mobTypeDamageAura(mobType proto.MobType, multiplier float64) {
	character.Env.RegisterPostFinalizeEffect(func() {
		for _, t := range character.Env.Encounter.Targets {
			if t.MobType == mobType {
				for _, at := range character.AttackTables[t.UnitIndex] {
					at.DamageDealtMultiplier *= multiplier
				}
			}
		}
	})
}

// Blood Fury: Attack Power and Spell Power increased by 10% for 15 sec. The Classic
// healing penalty is gone. The client gives it percent auras for attack power, ranged attack
// power and spell power, so the 10% follows the total while it's up. On the beta (2026-10-12)
// a Scroll of Strength used during Blood Fury added 20 attack power, where it adds 18 on its
// own.
//
// On the beta (2026-10-11, beta_results.txt) bonus healing also goes up 10% of its total,
// from 59 to 65 with 49 spell power and the +10 healing gloves. So the healing-only stat
// gets its 10% too. Blood Fury is also off the GCD: Lightning Shield went out 0.245 sec after
// it, and the client gives it no GCD category.
//
// Each weapon hit also gets 10% of the weapon's average damage, so the average hit is 1.1
// times bigger. The character sheet showed it first: the main hand range went up by 12.35 on
// Rage of the Storm (100 to 147) and 3.5 on Bloody Brass Knuckles (24 to 46), with the same
// spread, and the off-hand and ranged lines went up 10% too. The log agreed: 187 fist hits
// came out 1.105 times bigger and 34 Windfury attacks fit the weapon part too, where +10%
// attack power alone gives 1.056. Stormstrike rolls the same weapon, so it gets it too, and
// Stormstrikes with Blood Fury up fit that on a dagger and on a fist (2026-10-12). We add it
// to each weapon's damage range, after the weapon enchants and stones.
//
// Client 70425 lengthens its cooldown from 2 min to 3 min (SpellCooldowns 120000 to 180000);
// the effects are unchanged.
func (character *Character) registerForeverBloodFury() {
	actionID := ActionID{SpellID: 20572}

	statDeps := []*stats.StatDependency{
		character.NewDynamicMultiplyStat(stats.AttackPower, 1.1),
		character.NewDynamicMultiplyStat(stats.RangedAttackPower, 1.1),
		character.NewDynamicMultiplyStat(stats.SpellPower, 1.1),
		character.NewDynamicMultiplyStat(stats.HealingPower, 1.1),
	}
	var bonusWeaponDamage [3]float64
	weapons := func() [3]*Weapon {
		return [3]*Weapon{character.AutoAttacks.MH(), character.AutoAttacks.OH(), character.AutoAttacks.Ranged()}
	}
	aura := character.RegisterAura(Aura{
		Label:    "Blood Fury",
		ActionID: actionID,
		Duration: time.Second * 15,
		OnGain: func(aura *Aura, sim *Simulation) {
			for _, dep := range statDeps {
				character.EnableDynamicStatDep(sim, dep)
			}
			for i, weapon := range weapons() {
				bonusWeaponDamage[i] = weapon.AverageDamage() * 0.1
				weapon.BaseDamageMin += bonusWeaponDamage[i]
				weapon.BaseDamageMax += bonusWeaponDamage[i]
			}
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			for _, dep := range statDeps {
				character.DisableDynamicStatDep(sim, dep)
			}
			for i, weapon := range weapons() {
				weapon.BaseDamageMin -= bonusWeaponDamage[i]
				weapon.BaseDamageMax -= bonusWeaponDamage[i]
			}
		},
	})

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

// Shadowmeld (20580, client 1.60.1.70532): no global cooldown (no StartRecoveryCategory), a 10
// sec category cooldown that the tooltip raises to 2 min in combat, A_MOD_STEALTH, a threat drop
// (A_MOD_TOTAL_THREAT -999999) and the cancel auto attack bit (Attributes[0] 0x100000). Unlike
// Prowl and Stealth it has no not-in-combat bit, so it goes in combat. While it is up the
// character counts as stealthed for the abilities that need stealth (Ravage, Ambush, Garrote),
// melee is paused, and the next cast breaks it (AuraInterruptFlags 0x3C0F). As in
// wowsims/forever's shadowmeld.go the threat drop is left out: it does not change damage.
//
// The fight is in combat from the pull, so it always takes the 2 min cooldown. It cannot go
// before the pull: no class starts the fight in Stealth (orchestration/standards.md item 13).
// It is not a major cooldown, so only an APL line that follows it with a stealth ability casts it.
func (character *Character) registerShadowmeld() {
	actionID := ActionID{SpellID: 20580}

	var meleePaused bool
	character.ShadowmeldAura = character.RegisterAura(Aura{
		Label:    "Shadowmeld",
		ActionID: actionID,
		Duration: NeverExpires,
		OnGain: func(aura *Aura, sim *Simulation) {
			meleePaused = character.AutoAttacks.enabled
			character.AutoAttacks.CancelAutoSwing(sim)
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			if meleePaused {
				character.AutoAttacks.EnableAutoSwing(sim)
			}
		},
		OnCastComplete: func(aura *Aura, sim *Simulation, spell *Spell) {
			if spell.ActionID != actionID && spell.Flags.Matches(SpellFlagAPL|SpellFlagMCD) {
				aura.Deactivate(sim)
			}
		},
	})

	character.Shadowmeld = character.RegisterSpell(SpellConfig{
		ActionID: actionID,
		Flags:    SpellFlagAPL | SpellFlagNoOnCastComplete,
		Cast: CastConfig{
			CD: Cooldown{
				Timer:    character.NewTimer(),
				Duration: time.Minute * 2,
			},
		},
		ExtraCastCondition: func(sim *Simulation, _ *Unit) bool {
			return sim.CurrentTime >= 0 && !character.ShadowmeldAura.IsActive()
		},
		ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
			character.ShadowmeldAura.Activate(sim)
		},
	})
}

// IsShadowmelded is whether Shadowmeld is up, which counts as stealth for the abilities that
// need it.
func (character *Character) IsShadowmelded() bool {
	return character.ShadowmeldAura != nil && character.ShadowmeldAura.IsActive()
}

// BreakShadowmeld ends Shadowmeld, as the stealth abilities do when they land.
func (character *Character) BreakShadowmeld(sim *Simulation) {
	if character.IsShadowmelded() {
		character.ShadowmeldAura.Deactivate(sim)
	}
}

// Eureka!: damaging abilities cost 10% less and deal 10% more damage for 8 sec, for every
// class since the 2026-09-24 beta build (before it the discount was Energy 20%, Rage 40%,
// Mana 50% and Priests 15%). A Priest's healing counts too. 2 min cooldown.
//
// Client 70425 shortened it from 15 sec to 8 sec (duration index 8 to 31) and dropped its
// SpellAuraOptions row, so the 3-charge limit is gone: every ability in the 8 sec gets it.
// Under 70338 it was the next 3 damaging abilities within 15 sec.
//
// Which spells it covers comes from the client: each class's Eureka! names its spells with a
// class mask, one for the cost, one for direct damage and one for periodic damage
// (racials_forever_eureka.go). Only Drain Life, Drain Soul and Wrack (Warlock) and Mind
// Flay, Penance and Starshards (Priest) get it on their ticks, as the 2026-10-01 development
// notes say ("no longer benefits periodic effects at all. Channeled spells do not count as
// periodics"). So a Corruption, a curse, Siphon Life, Incinerate, Shadow Word: Pain,
// Rupture, Garrote or Rend gets nothing, and Immolate and Holy Fire get it on the hit only.
//
// A class without its own Eureka! keeps the general rule: a spell the rotation can cast,
// that has a cost and that is set up to deal damage (or heal, for a Priest).
func (character *Character) registerEureka() {
	actionID := ActionID{SpellID: 1259812}
	healingCounts := character.Class == proto.Class_ClassPriest
	scopes := eurekaScopes[character.Class]

	scopeOf := func(spell *Spell) uint8 {
		if scopes != nil {
			return scopes[spell.SpellID]
		}
		if !spell.Flags.Matches(SpellFlagAPL) || spell.Cost == nil {
			return 0
		}
		if spell.Flags.Matches(SpellFlagHelpful) {
			if healingCounts {
				return eurekaCost | eurekaDirect | eurekaPeriodic
			}
			return 0
		}
		if spell.DamageMultiplier == 0 {
			return 0
		}
		return eurekaCost | eurekaDirect | eurekaPeriodic
	}

	type affectedSpell struct {
		spell *Spell
		scope uint8
	}
	var affected []affectedSpell
	aura := character.RegisterAura(Aura{
		Label:    "Eureka!",
		ActionID: actionID,
		Duration: time.Second * 8, // duration index 31 in client 70425 (15 sec in 70338)
		OnInit: func(aura *Aura, sim *Simulation) {
			for _, spell := range character.Spellbook {
				if scope := scopeOf(spell); scope != 0 {
					affected = append(affected, affectedSpell{spell, scope})
				}
			}
		},
		OnGain: func(aura *Aura, sim *Simulation) {
			for _, a := range affected {
				applyEureka(a.spell, a.scope, 1)
			}
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			for _, a := range affected {
				applyEureka(a.spell, a.scope, -1)
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

// applyEureka adds (sign 1) or takes back (sign -1) Eureka!'s 10% on one spell. A spell whose
// damage all counts (direct damage with no ticks, a channel, a spell covered both ways) gets
// a 10% multiplier; Immolate or Holy Fire, covered on the hit only, get a 10% multiplier on the
// hit only.
// The cost cut is its own factor too: in game each percent cost row multiplies the others
// (wowsims/forever fef2bcaebd: Improved Wrath with Moonglow leaves a 120-mana Wrath at 45, not 30).
func applyEureka(spell *Spell, scope uint8, sign float64) {
	if scope&eurekaCost != 0 && spell.Cost != nil {
		if sign > 0 {
			spell.Cost.Factor *= 0.9
		} else {
			spell.Cost.Factor /= 0.9
		}
	}
	direct, periodic := scope&eurekaDirect != 0, scope&eurekaPeriodic != 0
	if !direct && !periodic {
		return
	}
	hasTicks := len(spell.Dots()) > 0 || spell.AOEDot() != nil
	if direct && hasTicks && !periodic {
		// A factor of its own on the hit too, beside the class's percent rows (Aftermath and
		// Agonizing Flames on Immolate multiply the impact multiplier).
		if sign > 0 {
			spell.ImpactDamageMultiplierAdditive *= 1.1
		} else {
			spell.ImpactDamageMultiplierAdditive /= 1.1
		}
		return
	}
	if sign > 0 {
		spell.DamageMultiplier *= 1.1
	} else {
		spell.DamageMultiplier /= 1.1
	}
}

// Touch of the Grave: spells and attacks have a 20% chance (30% for Mages, Priests and
// Warlocks) to drain health from the target, up to 5% of the caster's maximum health, at
// most once every 25 sec. Client 70425 raised the chances from 5% and 10% and the
// ProcCategoryRecovery from 1000 to 25000 ms on both procs (1260189 and 1260201); the
// drain itself (1260198) is unchanged.
//
// "Up to" reads as a cap, and the amount under it is not published. The drain is modelled
// at the cap, so this is the upper bound of what the racial can be worth. It is Shadow
// damage that cannot crit, cannot proc anything and heals the caster for what it deals.
//
// We assumed a 15 sec internal cooldown for a while and took it out on 2026-09-23 once the
// beta showed nothing like it. Client 70338 does give the proc aura a ProcCategoryRecovery of
// 1000 ms, a 1 sec internal cooldown too short to see in those logs, so the two triggers below
// share a 1 sec cooldown.
//
// The roll happens when a hostile ability is used, not when it lands, so an ability rolls
// whether or not it hits. Since the 2026-09-24 beta build the ability needs a damage
// component: an Expose Armor or a Kidney Shot no longer rolls, but the cast that puts a
// paladin's Consecration on the ground does (the way Shadow Word: Pain rolls on its cast
// in the development notes). Only the cast rolls there, the eight seconds of ticks that
// follow roll for nothing. We use the same test for a damage component as Eureka!: the
// spell has a damage multiplier.
//
// That is what the first trigger below is for. The second one is for everything that
// lands without announcing a cast: auto attacks, a paladin's judgement, a Windfury
// attack, anything the engine fires off the back of another spell. SpellFlagNoOnCastComplete
// is exactly that set, so between them the two triggers give every action one roll and
// no action two.
//
// The split shows up in what a miss does. An ability rolls when you press it, so it rolls
// whether or not it lands, but a swing only rolls when it connects. We have nothing from
// the beta on a missed swing and this is the quieter of the two guesses.
func (character *Character) registerTouchOfTheGrave() {
	actionID := ActionID{SpellID: 1260189}
	procChance := 0.20
	switch character.Class {
	case proto.Class_ClassMage, proto.Class_ClassPriest, proto.Class_ClassWarlock:
		actionID = ActionID{SpellID: 1260201}
		procChance = 0.30
	}
	healthMetrics := character.NewHealthMetrics(actionID)

	drain := character.RegisterSpell(SpellConfig{
		ActionID:    actionID,
		SpellSchool: SpellSchoolShadow,
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskEmpty,
		// "Touch of the Dead no longer scales with Shadow increases on the target or the
		// caster" (development notes, 2026-10-15, build 1.60.1.70425): no Shadowform, Shadow
		// Weaving, Curse of Shadow or Shadow Mastery on the drain.
		Flags: SpellFlagPassiveSpell | SpellFlagNoOnCastComplete | SpellFlagBinary | SpellFlagIgnoreModifiers,

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

	icd := Cooldown{
		Timer:    character.NewTimer(),
		Duration: time.Second * 25, // ProcCategoryRecovery 25000 in client 70425 (1000 in 70338)
	}

	proc := func(sim *Simulation, spell *Spell, result *SpellResult) {
		if !icd.IsReady(sim) {
			return
		}
		icd.Use(sim)
		// A cast gives the handler no result, so the drain follows the unit the player
		// is on rather than the one that was hit.
		target := character.CurrentTarget
		if result != nil {
			target = result.Target
		}
		if target != nil {
			drain.Cast(sim, target)
		}
	}

	MakeProcTriggerAura(&character.Unit, ProcTrigger{
		Name:     "Touch of the Grave",
		Callback: CallbackOnCastComplete,
		ProcMask: ProcMaskDirect,
		Handler: func(sim *Simulation, spell *Spell, result *SpellResult) {
			if spell.DamageMultiplier != 0 && sim.Proc(procChance, "Touch of the Grave") {
				proc(sim, spell, result)
			}
		},
	})
	MakeProcTriggerAura(&character.Unit, ProcTrigger{
		Name:       "Touch of the Grave (Attacks)",
		Callback:   CallbackOnSpellHitDealt,
		ProcMask:   ProcMaskDirect,
		SpellFlags: SpellFlagNoOnCastComplete,
		Outcome:    OutcomeLanded,
		ProcChance: procChance,
		Handler:    proc,
	})
}
