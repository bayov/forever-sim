package core

import (
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Registers all consume-related effects to the Agent.
// TODO: Classic Consumes
func applyConsumeEffects(agent Agent, raidBuffs *proto.RaidBuffs) {
	character := agent.GetCharacter()
	consumes := character.Consumes

	if consumes == nil {
		return
	}

	applyFlaskConsumes(character, consumes)
	applyWeaponImbueConsumes(character, consumes)
	applyFoodConsumes(character, consumes)
	applyDefensiveBuffConsumes(character, consumes)
	applyPhysicalBuffConsumes(character, consumes)
	applySpellBuffConsumes(character, consumes, raidBuffs)
	applyZanzaBuffConsumes(character, consumes)
	applyForeverElixirCrit(character, consumes)
	applyHitConsumableConsumes(character, consumes)
	applyMiscConsumes(character, consumes.MiscConsumes)

	registerPotionCD(agent, consumes)
	registerConjuredCD(agent, consumes)
	registerExplosivesCD(agent, consumes)
}

// Under Forever pets take no consumables: the pet scrolls' spells (Scroll of Agility IV's 12174,
// Scroll of Strength IV's 12179) carry SPELL_ATTR3_ONLY_TARGET_PLAYERS (Attributes[3] 0x100),
// and the client has no pet food item (wowsims/forever a1d94ad2f0, abbf9e3a6f).
func ApplyPetConsumeEffects(pet *Character, ownerConsumes *proto.Consumes) {
	if pet.Env.IsForever() {
		return
	}
	pet.AddStat(stats.AttackPower, []float64{0, 40}[ownerConsumes.PetAttackPowerConsumable])
	pet.AddStat(stats.Agility, []float64{0, 17, 13, 9, 5}[ownerConsumes.PetAgilityConsumable])
	pet.AddStat(stats.Strength, []float64{0, 30, 17, 13, 9, 5}[ownerConsumes.PetStrengthConsumable])
}

///////////////////////////////////////////////////////////////////////////
//                             Flasks
///////////////////////////////////////////////////////////////////////////

func applyFlaskConsumes(character *Character, consumes *proto.Consumes) {
	if consumes.Flask == proto.Flask_FlaskUnknown {
		return
	}

	switch consumes.Flask {
	case proto.Flask_FlaskOfDistilledWisdom:
		character.AddStats(stats.Stats{
			stats.Mana: 2000,
		})
	case proto.Flask_FlaskOfSupremePower:
		character.AddStats(stats.Stats{
			stats.SpellPower: 150,
		})
	case proto.Flask_FlaskOfTheTitans:
		character.AddStats(stats.Stats{
			stats.Health: 1200,
		})
	case proto.Flask_FlaskOfChromaticResistance:
		character.AddResistances(25)
	case proto.Flask_FlaskOfNaturalAccuracy, proto.Flask_FlaskOfNaturalAggression,
		proto.Flask_FlaskOfNaturalPrecision, proto.Flask_FlaskOfNaturalSwiftness:
		applyNaturalFlask(character, consumes.Flask)
	}
}

// Forever's Flasks of Natural Accuracy, Aggression, Precision and Swiftness (items 274273 to
// 274276, spells 1293740 to 1293743) give 60 Stamina anywhere. Their second effect, an A_DUMMY,
// is the bonus the tooltip grants in Mount Hyjal, Hyjal Summit and the Barrow Deeps, and the
// auras after it name what that bonus is (wowsims/forever's consumables_natural_flasks.go):
//   - Accuracy: 5% melee and ranged hit (A_MOD_HIT_CHANCE) and 5% spell hit.
//   - Aggression: 4% crit (A_MOD_CRIT_PCT), melee, ranged and spell.
//   - Precision: 5% expertise (A_MOD_EXPERTISE), 5% off the target's dodge and parry.
//   - Swiftness: 5% melee and ranged attack speed (A_MOD_MELEE_RANGED_HASTE_2) and 5% casting
//     speed (A_MOD_CASTING_SPEED_NOT_STACK).
//
// The zone bonus applies only when the encounter says the fight is in one of those zones.
func applyNaturalFlask(character *Character, flask proto.Flask) {
	character.AddStat(stats.Stamina, 60)
	if !character.Env.Encounter.NaturalFlaskZone {
		return
	}
	switch flask {
	case proto.Flask_FlaskOfNaturalAccuracy:
		character.AddStats(stats.Stats{
			stats.MeleeHit: 5 * MeleeHitRatingPerHitChance,
			stats.SpellHit: 5 * SpellHitRatingPerHitChance,
		})
	case proto.Flask_FlaskOfNaturalAggression:
		character.AddStats(stats.Stats{
			stats.MeleeCrit: 4 * CritRatingPerCritChance,
			stats.SpellCrit: 4 * SpellCritRatingPerCritChance,
		})
	case proto.Flask_FlaskOfNaturalPrecision:
		character.AddStat(stats.Expertise, 5*ExpertiseRatingPerExpertiseChance)
	case proto.Flask_FlaskOfNaturalSwiftness:
		character.PseudoStats.MeleeSpeedMultiplier *= 1.05
		character.PseudoStats.RangedSpeedMultiplier *= 1.05
		character.PseudoStats.CastSpeedMultiplier *= 1.05
	}
}

///////////////////////////////////////////////////////////////////////////
//                             Weapon Imbues
///////////////////////////////////////////////////////////////////////////

func applyWeaponImbueConsumes(character *Character, consumes *proto.Consumes) {
	// There must be a nicer way to do this...
	shadowOilIcd := Cooldown{
		Timer:    character.NewTimer(),
		Duration: time.Second * 10,
	}

	if character.HasMHWeapon() {
		addImbueStats(character, consumes.MainHandImbue, true, shadowOilIcd)
	}
	if character.OffHand() != nil {
		addImbueStats(character, consumes.OffHandImbue, false, shadowOilIcd)
	}
}

// weaponEffectsWork says whether weapon imbues apply. A feral druid fights with its paws, so
// outside Forever its weapon's stones, oils and totem imbues do nothing. Forever makes them
// work in form: "Weapon enchantments and procs now all work while shapeshifted, including
// effects like weightstones, Crusader enchant, and Windfury Totem" (Blizzard's World of
// Warcraft: Forever class deep dive, Druid).
func (character *Character) weaponEffectsWork() bool {
	return !character.PseudoStats.FeralCombatEnabled || character.Env.IsForever()
}

// addImbueWeaponDamage adds a sharpening stone's or weightstone's flat damage to a weapon. A
// Forever cat takes it on its paw whole, like a weapon enchant's (sim/druid/forms.go,
// GetCatWeapon).
func addImbueWeaponDamage(character *Character, isMh bool, amount float64) {
	if character.PseudoStats.FeralCombatEnabled && !(character.Env.IsForever() && isMh) {
		return
	}
	weapon := character.AutoAttacks.MH()
	if !isMh {
		weapon = character.AutoAttacks.OH()
	}
	weapon.BaseDamageMin += amount
	weapon.BaseDamageMax += amount
}

// addForeverWeaponCrit is Forever's Elemental Sharpening Stone, Critical Weapon +2 (22755): its
// crit reaches only the attacks made with the weapon the stone is on, not ranged attacks or spells
// (wowsims/forever 90a276c26a, 002fddc3f2). The main hand's goes on the sheet's melee crit and is
// taken back off the off hand's attacks and the ranged ones; the off hand's goes on its attacks.
func addForeverWeaponCrit(character *Character, isMh bool, crit float64) {
	bonus := crit * CritRatingPerCritChance
	if isMh {
		character.AddStat(stats.MeleeCrit, bonus)
		character.OnSpellRegistered(func(spell *Spell) {
			offHandOnly := spell.ProcMask.Matches(ProcMaskMeleeOH) && !spell.ProcMask.Matches(ProcMaskMeleeMH)
			ranged := spell.CastType == proto.CastType_CastTypeRanged || spell.ProcMask.Matches(ProcMaskRanged)
			if offHandOnly || ranged {
				spell.BonusCritRating -= bonus
			}
		})
		return
	}
	character.OnSpellRegistered(func(spell *Spell) {
		if spell.ProcMask.Matches(ProcMaskMeleeOH) {
			spell.BonusCritRating += bonus
		}
	})
}

// Forever's wizard oils add as much healing as spell damage.
func addForeverOilHealing(character *Character, amount float64) {
	if character.Env.IsForever() {
		character.AddStat(stats.HealingPower, amount)
	}
}

func addImbueStats(character *Character, imbue proto.WeaponImbue, isMh bool, shadowOilIcd Cooldown) {
	if imbue != proto.WeaponImbue_WeaponImbueUnknown {
		switch imbue {
		// Wizard Oils. Forever's add the same amount of healing.
		case proto.WeaponImbue_MinorWizardOil:
			character.AddStats(stats.Stats{
				stats.SpellPower: 8,
			})
			addForeverOilHealing(character, 8)
		case proto.WeaponImbue_LesserWizardOil:
			character.AddStats(stats.Stats{
				stats.SpellPower: 16,
			})
			addForeverOilHealing(character, 16)
		case proto.WeaponImbue_WizardOil:
			character.AddStats(stats.Stats{
				stats.SpellPower: 24,
			})
			addForeverOilHealing(character, 24)
		case proto.WeaponImbue_BrilliantWizardOil:
			character.AddStats(stats.Stats{
				stats.SpellPower: 36,
				stats.SpellCrit:  1 * SpellCritRatingPerCritChance,
			})
			addForeverOilHealing(character, 36)
		// Forever's Blessed Wizard Oil is 58 spell damage against Undead (wowsims/forever's client
		// store 28890, "Increase Spell Dam Undead 58").
		case proto.WeaponImbue_BlessedWizardOil:
			bonus := TernaryFloat64(character.Env.IsForever(), 58, 60)
			character.Env.RegisterPostFinalizeEffect(func() {
				for _, target := range character.Env.Encounter.TargetUnits {
					if target.MobType != proto.MobType_MobTypeUndead {
						continue
					}

					for _, at := range character.AttackTables[target.UnitIndex] {
						at.BonusSpellDamageTaken += bonus
					}
				}
			})

		// Mana Oils. Forever's give more mana and all of them add healing.
		case proto.WeaponImbue_MinorManaOil:
			if character.Env.IsForever() {
				character.AddStats(stats.Stats{stats.MP5: 5, stats.HealingPower: 10})
			} else {
				character.AddStat(stats.MP5, 4)
			}
		case proto.WeaponImbue_LesserManaOil:
			if character.Env.IsForever() {
				character.AddStats(stats.Stats{stats.MP5: 10, stats.HealingPower: 20})
			} else {
				character.AddStat(stats.MP5, 8)
			}
		case proto.WeaponImbue_BrilliantManaOil:
			if character.Env.IsForever() {
				character.AddStats(stats.Stats{stats.MP5: 15, stats.HealingPower: 30})
			} else {
				character.AddStats(stats.Stats{stats.MP5: 12, stats.HealingPower: 25})
			}

		// Sharpening Stones
		case proto.WeaponImbue_SolidSharpeningStone:
			addImbueWeaponDamage(character, isMh, 6)
		case proto.WeaponImbue_DenseSharpeningStone:
			addImbueWeaponDamage(character, isMh, 8)
		case proto.WeaponImbue_ElementalSharpeningStone:
			if character.Env.IsForever() {
				addForeverWeaponCrit(character, isMh, 2)
			} else if character.weaponEffectsWork() {
				character.AddStats(stats.Stats{
					stats.MeleeCrit: 2 * CritRatingPerCritChance,
				})
				character.AddBonusRangedCritRating(-2.0)
			}
		// Forever's Consecrated Sharpening Stone is 99 attack power against Undead, once per hand
		// it is on (wowsims/forever's client store 28893, "Undead Slayer 99", and fcf802739d).
		case proto.WeaponImbue_ConsecratedSharpeningStone:
			bonus := TernaryFloat64(character.Env.IsForever(), 99, 100)
			character.Env.RegisterPostFinalizeEffect(func() {
				for _, target := range character.Env.Encounter.TargetUnits {
					if target.MobType != proto.MobType_MobTypeUndead {
						continue
					}

					for _, at := range character.AttackTables[target.UnitIndex] {
						at.BonusAttackPowerTaken += bonus
					}
				}
			})

		// Weightstones
		case proto.WeaponImbue_SolidWeightstone:
			addImbueWeaponDamage(character, isMh, 6)
		case proto.WeaponImbue_DenseWeightstone:
			addImbueWeaponDamage(character, isMh, 8)
		// Windfury Totem from another shaman. It used to take the main hand imbue slot, and
		// old settings still carry it there. It is the raid buffs' totem weapon buff now.
		case proto.WeaponImbue_Windfury:
			if character.weaponEffectsWork() {
				applyRaidTotemWeaponBuff(character, proto.TotemWeaponBuff_TotemWeaponBuffWindfury)
			}
		case proto.WeaponImbue_ShadowOil:
			if character.weaponEffectsWork() {
				registerShadowOil(character, isMh, shadowOilIcd)
			}
		case proto.WeaponImbue_FrostOil:
			if character.weaponEffectsWork() {
				registerFrostOil(character, isMh)
			}
		}
	}
}

// Forever's oil procs, Shadow Bolt (350026) and Frostbolt (350025) in client 1.60.1.70338,
// scale with the wielder's level up to their SpellLevels MaxLevel, 17 and 12. Every Forever
// sim is above that, so the range is the same at level 20 and 60. The proc chances in
// SpellItemEnchantment (15 and 10) are unchanged from Classic.
func shadowOilDamageRange(env *Environment) (float64, float64) {
	if env.IsForever() {
		return 63, 71
	}
	return 52, 61
}

func frostOilDamageRange(env *Environment) (float64, float64) {
	if env.IsForever() {
		return 39, 43
	}
	return 33, 38
}

func registerShadowOil(character *Character, isMh bool, icd Cooldown) {
	procChance := 0.15
	minDamage, maxDamage := shadowOilDamageRange(character.Env)

	procSpell := character.GetOrRegisterSpell(SpellConfig{
		ActionID:    ActionID{SpellID: 1382},
		SpellSchool: SpellSchoolShadow,
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskEmpty,
		Flags:       SpellFlagNoOnCastComplete | SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 0.56,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			damage := sim.Roll(minDamage, maxDamage)
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHitAndCrit)
		},
	})

	label := " MH"
	procMask := ProcMaskMeleeMH
	if !isMh {
		label = " OH"
		procMask = ProcMaskMeleeOH
	}

	MakePermanent(character.GetOrRegisterAura(Aura{
		Label: "Shadow Oil" + label,
		OnSpellHitDealt: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if !result.Landed() {
				return
			}

			if !spell.ProcMask.Matches(procMask) {
				return
			}

			if !icd.IsReady(sim) {
				return
			}

			if sim.RandomFloat("Shadow Oil") < procChance {
				icd.Use(sim)
				procSpell.Cast(sim, result.Target)
			}
		},
	}))
}

func registerFrostOil(character *Character, isMh bool) {
	procChance := 0.10
	minDamage, maxDamage := frostOilDamageRange(character.Env)

	procSpell := character.GetOrRegisterSpell(SpellConfig{
		ActionID:    ActionID{SpellID: 1191},
		SpellSchool: SpellSchoolFrost,
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskEmpty,
		Flags:       SpellFlagNoOnCastComplete | SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 0.269,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			damage := sim.Roll(minDamage, maxDamage)
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHitAndCrit)
		},
	})

	label := " MH"
	procMask := ProcMaskMeleeMHAuto
	if !isMh {
		label = " OH"
		procMask = ProcMaskMeleeOHAuto
	}

	MakePermanent(character.GetOrRegisterAura(Aura{
		Label: "Frost Oil" + label,
		OnSpellHitDealt: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if !result.Landed() {
				return
			}

			if !spell.ProcMask.Matches(procMask) {
				return
			}

			if sim.RandomFloat("Frost Oil") < procChance {
				procSpell.Cast(sim, result.Target)
			}
		},
	}))
}

///////////////////////////////////////////////////////////////////////////
//                             Food
///////////////////////////////////////////////////////////////////////////

func applyFoodConsumes(character *Character, consumes *proto.Consumes) {
	if character.Env.IsForever() && applyForeverFood(character, consumes.Food) {
		// Forever changed this food, so we skip its Classic stats below.
	} else if consumes.Food != proto.Food_FoodUnknown {
		switch consumes.Food {
		case proto.Food_FoodHotWolfRibs:
			character.AddStats(stats.Stats{
				stats.Stamina: 8,
				stats.Spirit:  8,
			})
		case proto.Food_FoodSmokedSagefish:
			character.AddStats(stats.Stats{
				stats.MP5: 3,
			})
		case proto.Food_FoodSagefishDelight:
			character.AddStats(stats.Stats{
				stats.MP5: 6,
			})
		case proto.Food_FoodTenderWolfSteak:
			character.AddStats(stats.Stats{
				stats.Stamina: 12,
				stats.Spirit:  12,
			})
		case proto.Food_FoodGrilledSquid:
			character.AddStats(stats.Stats{
				stats.Agility: 10,
			})
		case proto.Food_FoodSmokedDesertDumpling:
			character.AddStats(stats.Stats{
				stats.Strength: 20,
			})
		case proto.Food_FoodNightfinSoup:
			character.AddStats(stats.Stats{
				stats.MP5: 8,
			})
		case proto.Food_FoodRunnTumTuberSurprise:
			character.AddStats(stats.Stats{
				stats.Intellect: 10,
			})
		case proto.Food_FoodDirgesKickChimaerokChops:
			character.AddStats(stats.Stats{
				stats.Stamina: 25,
			})
		case proto.Food_FoodBlessedSunfruitJuice:
			character.AddStats(stats.Stats{
				stats.Spirit: 10,
			})
		case proto.Food_FoodBlessSunfruit:
			character.AddStats(stats.Stats{
				stats.Strength: 10,
			})
		case proto.Food_FoodBarbecuedBuzzardWing:
			character.AddStats(stats.Stats{
				stats.Intellect: 10,
			})
		case proto.Food_FoodMithrilHeadTrout:
			character.AddStats(stats.Stats{
				stats.AttackPower:       20,
				stats.RangedAttackPower: 20,
			})
		case proto.Food_FoodBearBrisket:
			character.AddStats(stats.Stats{
				stats.Strength: 10,
			})
		case proto.Food_FoodTastyLionSteak:
			character.AddStats(stats.Stats{
				stats.Agility: 10,
			})
		case proto.Food_FoodBrinySeafoodStew:
			character.AddStats(stats.Stats{
				stats.SpellDamage: 14,
			})
		}
	}

	if consumes.Alcohol != proto.Alcohol_AlcoholUnknown {
		switch consumes.Alcohol {
		case proto.Alcohol_AlcoholRumseyRumBlackLabel:
			character.AddStats(stats.Stats{
				stats.Stamina: 15,
			})
		case proto.Alcohol_AlcoholGordokGreenGrog:
			character.AddStats(stats.Stats{
				stats.Stamina: 10,
			})
		case proto.Alcohol_AlcoholRumseyRumDark:
			character.AddStats(stats.Stats{
				stats.Stamina: 10,
			})
		case proto.Alcohol_AlcoholRumseyRumLight:
			character.AddStats(stats.Stats{
				stats.Stamina: 5,
			})
		case proto.Alcohol_AlcoholKreegsStoutBeatdown:
			character.AddStats(stats.Stats{
				stats.Spirit:    25,
				stats.Intellect: -5,
			})
		}
	}

	if consumes.DragonBreathChili {
		MakePermanent(DragonBreathChiliAura(character))
	}
}

// Forever gives each well fed food one stat. Returns whether Forever changed the food, so the caller
// skips its Classic stats. The other foods are the same in Forever. The values are the Well Fed
// aura's amount in client 1.60.1.70338 (SpellEffect, aura 227 on the food's spell).
func applyForeverFood(character *Character, food proto.Food) bool {
	if bonus, ok := foreverFoodStats[food]; ok {
		character.AddStats(bonus)
		return true
	}
	switch food {
	case proto.Food_FoodHotWolfRibs:
		character.AddStat(stats.Agility, 10)
	case proto.Food_FoodTenderWolfSteak:
		character.AddStat(stats.Agility, 15)
	case proto.Food_FoodRunnTumTuberSurprise:
		character.AddStat(stats.Intellect, 15)
	case proto.Food_FoodSmokedSagefish:
		character.AddStat(stats.SpellDamage, 4)
	case proto.Food_FoodSagefishDelight:
		character.AddStat(stats.SpellDamage, 7)
	case proto.Food_FoodNightfinSoup:
		character.AddStat(stats.SpellDamage, 22)
	case proto.Food_FoodGrilledSquid:
		character.AddStats(stats.Stats{
			stats.MeleeCrit: 1 * CritRatingPerCritChance,
			stats.SpellCrit: 1 * SpellCritRatingPerCritChance,
		})
	default:
		return false
	}
	return true
}

func meleeAndRangedAP(ap float64) stats.Stats {
	return stats.Stats{stats.AttackPower: ap, stats.RangedAttackPower: ap}
}

// The Forever foods that Classic doesn't have, or that Classic gives something else.
var foreverFoodStats = map[proto.Food]stats.Stats{
	// Strength
	proto.Food_FoodSavoryStagSliders: {stats.Strength: 15},
	proto.Food_FoodSucculentPorkRibs: {stats.Strength: 5},
	proto.Food_FoodGoretuskLiverPie:  {stats.Strength: 3},
	proto.Food_FoodRoastedBoarMeat:   {stats.Strength: 1},
	// Agility
	proto.Food_FoodFlankAuPoivre:     {stats.Agility: 20},
	proto.Food_FoodSeasonedWolfKabob: {stats.Agility: 5},
	proto.Food_FoodCoyoteSteak:       {stats.Agility: 3},
	proto.Food_FoodSpicedWolfMeat:    {stats.Agility: 1},
	// Attack power, melee and ranged
	proto.Food_FoodMightfishSteak:        meleeAndRangedAP(40),
	proto.Food_FoodCookedGlossyMightfish: meleeAndRangedAP(30),
	proto.Food_FoodLongjawMudSnapper:     meleeAndRangedAP(10),
	proto.Food_FoodSlitherskinMackerel:   meleeAndRangedAP(6),
	proto.Food_FoodBrilliantSmallfish:    meleeAndRangedAP(2),
	// Spell damage, all schools
	proto.Food_FoodLobsterStew:        {stats.SpellDamage: 28},
	proto.Food_FoodRainbowFinAlbacore: {stats.SpellDamage: 1},
	// Intellect
	proto.Food_FoodPrehistoricPulledRaptor: {stats.Intellect: 20},
	proto.Food_FoodRoastRaptor:             {stats.Intellect: 5},
	proto.Food_FoodTwiceSpicedRaptorSlice:  {stats.Intellect: 3},
	proto.Food_FoodTastyRaptorBites:        {stats.Intellect: 1},
	// Spirit
	proto.Food_FoodWickedSmoothie:    {stats.Spirit: 20},
	proto.Food_FoodSpicySmoothie:     {stats.Spirit: 15},
	proto.Food_FoodCalcifiedSmoothie: {stats.Spirit: 10},
	proto.Food_FoodMrrgglSmrrthle:    {stats.Spirit: 5},
	proto.Food_FoodSlimySmoothie:     {stats.Spirit: 3},
	proto.Food_FoodVenomousSmoothie:  {stats.Spirit: 1},
	// Stamina
	proto.Food_FoodSavoryTurtleStew:     {stats.Stamina: 20},
	proto.Food_FoodSpiderSausage:        {stats.Stamina: 15},
	proto.Food_FoodGiantScrambledEggs:   {stats.Stamina: 10},
	proto.Food_FoodCuriouslyTastyOmelet: {stats.Stamina: 5},
	proto.Food_FoodBreakfastOmelette:    {stats.Stamina: 3},
	proto.Food_FoodSkywallSouffle:       {stats.Stamina: 1},
	// Healing
	proto.Food_FoodSagesTea:  {stats.HealingPower: 44},
	proto.Food_FoodSunnyTea:  {stats.HealingPower: 33},
	proto.Food_FoodTriageTea: {stats.HealingPower: 22},
	proto.Food_FoodRootTea:   {stats.HealingPower: 11},
	proto.Food_FoodRoyalTea:  {stats.HealingPower: 7},
	proto.Food_FoodPeaceTea:  {stats.HealingPower: 2},
	// Armor
	proto.Food_FoodPlatedArmorfish: {stats.BonusArmor: 150},
}

// Classic's Dragonbreath Chili (15851) does a flat 60. Forever's does 65 with a variance of
// 0.123 (client 1.60.1.70338), so it rolls 61 to 69.
func dragonbreathChiliDamage(sim *Simulation) float64 {
	if sim.IsForever() {
		return sim.Roll(61, 69)
	}
	return 60
}

func DragonBreathChiliAura(character *Character) *Aura {
	procChance := .05
	icd := Cooldown{
		Timer:    character.NewTimer(),
		Duration: time.Second * 10,
	}

	procSpell := character.RegisterSpell(SpellConfig{
		ActionID:    ActionID{SpellID: 15851},
		SpellSchool: SpellSchoolFire,
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskSpellDamageProc | ProcMaskSpellProc,
		Flags:       SpellFlagNoOnCastComplete | SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			for _, aoeTarget := range sim.Environment.Encounter.ActiveTargetUnits {
				spell.CalcAndDealDamage(sim, aoeTarget, dragonbreathChiliDamage(sim), spell.OutcomeMagicHitAndCrit)
			}
		},
	})

	aura := character.GetOrRegisterAura(Aura{
		Label:    "Dragonbreath Chili",
		ActionID: ActionID{SpellID: 15852},
		Duration: NeverExpires,
		OnSpellHitDealt: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(ProcMaskMelee) {
				return
			}
			if icd.IsReady(sim) && sim.RandomFloat("Dragonbreath Chili") < procChance {
				icd.Use(sim)
				procSpell.Cast(sim, result.Target)
			}
		},
	})
	return aura
}

///////////////////////////////////////////////////////////////////////////
//                             Defensive Buff Consumes
///////////////////////////////////////////////////////////////////////////

func applyDefensiveBuffConsumes(character *Character, consumes *proto.Consumes) {
	if consumes.ArmorElixir != proto.ArmorElixir_ArmorElixirUnknown {
		switch consumes.ArmorElixir {
		case proto.ArmorElixir_ElixirOfSuperiorDefense:
			character.AddStats(stats.Stats{
				stats.BonusArmor: 450,
			})
		case proto.ArmorElixir_ElixirOfGreaterDefense:
			character.AddStats(stats.Stats{
				stats.BonusArmor: 250,
			})
		case proto.ArmorElixir_ElixirOfDefense:
			character.AddStats(stats.Stats{
				stats.BonusArmor: 150,
			})
		case proto.ArmorElixir_ElixirOfMinorDefense:
			character.AddStats(stats.Stats{
				stats.BonusArmor: 50,
			})
		case proto.ArmorElixir_ScrollOfProtection:
			character.AddStat(stats.BonusArmor, scrollOfProtection[HighestRankAt(character.Level, scrollOfProtectionLevel[:])])
		case proto.ArmorElixir_ElixirOfThePhalanx:
			character.AddStats(stats.Stats{
				stats.Health:     400,
				stats.BonusArmor: 500,
			})
		}
	}

	if consumes.HealthElixir != proto.HealthElixir_HealthElixirUnknown {
		switch consumes.HealthElixir {
		case proto.HealthElixir_ElixirOfFortitude:
			character.AddStats(stats.Stats{
				stats.Health: 120,
			})
		case proto.HealthElixir_ElixirOfMinorFortitude:
			character.AddStats(stats.Stats{
				stats.Health: 27,
			})
		case proto.HealthElixir_ElixirOfFortitudeForever:
			character.AddStat(stats.Health, 200)
		case proto.HealthElixir_ElixirOfGreaterFortitude:
			character.AddStat(stats.Health, 400)
		}
	}
}

///////////////////////////////////////////////////////////////////////////
//                             Physical Buff Consumes
///////////////////////////////////////////////////////////////////////////

func applyPhysicalBuffConsumes(character *Character, consumes *proto.Consumes) {
	if consumes.AttackPowerBuff != proto.AttackPowerBuff_AttackPowerBuffUnknown {
		switch consumes.AttackPowerBuff {
		case proto.AttackPowerBuff_JujuMight:
			character.AddStats(stats.Stats{
				stats.AttackPower:       40,
				stats.RangedAttackPower: 40,
			})
		case proto.AttackPowerBuff_WinterfallFirewater, proto.AttackPowerBuff_DistilledFirewater:
			character.AddStats(stats.Stats{
				stats.AttackPower: 35,
			})
		}
	}

	if consumes.AgilityElixir != proto.AgilityElixir_AgilityElixirUnknown {
		switch consumes.AgilityElixir {
		case proto.AgilityElixir_ElixirOfTheMongoose:
			if character.Env.IsForever() {
				// Forever's Mongoose (17538) is all crit (aura 290), like the other +2% elixirs.
				// The crit is applyForeverElixirCrit's.
				character.AddStat(stats.Agility, 25)
			} else {
				character.AddStats(stats.Stats{
					stats.Agility:   25,
					stats.MeleeCrit: 2 * CritRatingPerCritChance,
				})
			}
		case proto.AgilityElixir_ElixirOfGreaterAgility:
			character.AddStats(stats.Stats{
				stats.Agility: 25,
			})
		case proto.AgilityElixir_ElixirOfAgility:
			character.AddStats(stats.Stats{
				stats.Agility: 15,
			})
		case proto.AgilityElixir_ElixirOfLesserAgility:
			character.AddStats(stats.Stats{
				stats.Agility: 8,
			})
		case proto.AgilityElixir_ScrollOfAgility:
			character.AddStat(stats.Agility, scrollValue(character.Level))
		case proto.AgilityElixir_ElixirOfCunning:
			character.AddStats(stats.Stats{
				stats.Agility:   25,
				stats.Intellect: 25,
			})
		}
	}

	if consumes.StrengthBuff != proto.StrengthBuff_StrengthBuffUnknown {
		switch consumes.StrengthBuff {
		case proto.StrengthBuff_JujuPower:
			character.AddStats(stats.Stats{
				stats.Strength: 30,
			})
		case proto.StrengthBuff_ElixirOfGiants:
			character.AddStats(stats.Stats{
				stats.Strength: 25,
			})
		case proto.StrengthBuff_ElixirOfOgresStrength:
			character.AddStats(stats.Stats{
				stats.Strength: 8,
			})
		case proto.StrengthBuff_ScrollOfStrength:
			character.AddStat(stats.Strength, scrollValue(character.Level))
		case proto.StrengthBuff_ElixirOfGiantGrowth:
			character.AddStat(stats.Strength, 8)
		case proto.StrengthBuff_ElixirOfStrength:
			character.AddStat(stats.Strength, 10)
		case proto.StrengthBuff_ElixirOfTheGrizzly:
			character.AddStat(stats.Strength, 25)
		case proto.StrengthBuff_ElixirOfFerocity:
			character.AddStats(stats.Stats{
				stats.Strength: 18,
				stats.Agility:  18,
			})
		}
	}
}

// Forever's best elixirs add 2% critical strike chance (client aura 290, all crit), which is both
// melee and spell crit under Forever's single crit chance.
var foreverElixirCrit = stats.Stats{
	stats.MeleeCrit: 2 * CritRatingPerCritChance,
	stats.SpellCrit: 2 * SpellCritRatingPerCritChance,
}

// Five elixirs state 2% crit beside their own stat: Forever's Elixir of the Mongoose (Agility),
// Grizzly (Strength), Owl (Intellect), and Sages and Whale (Spirit). Each sits in its own stat's
// slot, so the stats stack, but only one elixir's crit counts: the game files carry no stacking
// table that lets four of them add 8%, and that much free crit is not plausible in game
// (wowsims/forever 2c0fd57e68, one ElixirCrit category where only the strongest applies).
func applyForeverElixirCrit(character *Character, consumes *proto.Consumes) {
	mongoose := consumes.AgilityElixir == proto.AgilityElixir_ElixirOfTheMongoose && character.Env.IsForever()
	if mongoose ||
		consumes.StrengthBuff == proto.StrengthBuff_ElixirOfTheGrizzly ||
		consumes.IntellectElixir == proto.IntellectElixir_ElixirOfTheOwl ||
		consumes.SpiritElixir == proto.SpiritElixir_ElixirOfSages ||
		consumes.SpiritElixir == proto.SpiritElixir_ElixirOfTheWhale {
		character.AddStats(foreverElixirCrit)
	}
}

// Scrolls of Agility and Strength give 5 / 9 / 13 / 17 at ranks I to IV, which need
// level 10, 25, 40 and 55. We use the best rank the character can read, so a level 30
// character gets Scroll of Strength II's 9 rather than rank IV's 17.
var scrollLevel = [...]int{0, 10, 25, 40, 55}
var scrollStat = [...]float64{0, 5, 9, 13, 17}

func scrollValue(level int32) float64 {
	return scrollStat[HighestRankAt(level, scrollLevel[:])]
}

// Scrolls of Protection give 60 / 120 / 180 / 240 armor and need level 1, 15, 30 and 45.
var scrollOfProtectionLevel = [...]int{0, 1, 15, 30, 45}
var scrollOfProtection = [...]float64{0, 60, 120, 180, 240}

///////////////////////////////////////////////////////////////////////////
//                             Spell Buff Consumes
///////////////////////////////////////////////////////////////////////////

func applySpellBuffConsumes(character *Character, consumes *proto.Consumes, raidBuffs *proto.RaidBuffs) {
	if consumes.SpellPowerBuff != proto.SpellPowerBuff_SpellPowerBuffUnknown {
		switch consumes.SpellPowerBuff {
		case proto.SpellPowerBuff_ArcaneElixir:
			character.AddStats(stats.Stats{
				stats.SpellDamage: 20,
			})
		case proto.SpellPowerBuff_GreaterArcaneElixir:
			character.AddStats(stats.Stats{
				stats.SpellDamage: 35,
			})
		case proto.SpellPowerBuff_LesserArcaneElixir:
			character.AddStats(stats.Stats{
				stats.SpellDamage: 15,
			})
		case proto.SpellPowerBuff_MinorArcaneElixir:
			character.AddStat(stats.SpellDamage, 5)
		}
	}

	if consumes.FirePowerBuff != proto.FirePowerBuff_FirePowerBuffUnknown {
		switch consumes.FirePowerBuff {
		case proto.FirePowerBuff_ElixirOfFirepower:
			character.AddStats(stats.Stats{
				stats.FirePower: 10,
			})
		case proto.FirePowerBuff_ElixirOfGreaterFirepower:
			// Forever made it the Elixir of Holy Power, +40 Holy instead of Fire.
			if character.Env.IsForever() {
				character.AddStat(stats.HolyPower, 40)
			} else {
				character.AddStat(stats.FirePower, 40)
			}
		}
	}

	if consumes.NaturePowerBuff == proto.NaturePowerBuff_ElixirOfNaturePower {
		character.AddStat(stats.NaturePower, 40)
	}

	switch consumes.HealingPowerBuff {
	case proto.HealingPowerBuff_GreaterClericsElixir:
		character.AddStat(stats.HealingPower, 40)
	case proto.HealingPowerBuff_ClericsElixir:
		character.AddStat(stats.HealingPower, 30)
	case proto.HealingPowerBuff_LesserClericsElixir:
		character.AddStat(stats.HealingPower, 20)
	case proto.HealingPowerBuff_MinorClericsElixir:
		character.AddStat(stats.HealingPower, 10)
	}

	if consumes.ShadowPowerBuff != proto.ShadowPowerBuff_ShadowPowerBuffUnknown {
		switch consumes.ShadowPowerBuff {
		case proto.ShadowPowerBuff_ElixirOfShadowPower:
			character.AddStats(stats.Stats{
				stats.ShadowPower: 40,
			})
		}
	}

	if consumes.FrostPowerBuff != proto.FrostPowerBuff_FrostPowerBuffUnknown {
		switch consumes.FrostPowerBuff {
		case proto.FrostPowerBuff_ElixirOfFrostPower:
			character.AddStats(stats.Stats{
				stats.FrostPower: 15,
			})
		}
	}

	if consumes.ManaRegenElixir != proto.ManaRegenElixir_ManaRegenElixirUnknown {
		switch consumes.ManaRegenElixir {
		case proto.ManaRegenElixir_MagebloodPotion:
			character.AddStats(stats.Stats{
				stats.MP5: 12,
			})
		case proto.ManaRegenElixir_LesserMagebloodElixir:
			character.AddStats(stats.Stats{
				stats.MP5: 6,
			})
		case proto.ManaRegenElixir_GreaterMagebloodElixir:
			character.AddStat(stats.MP5, 20)
		case proto.ManaRegenElixir_MinorMagebloodElixir:
			character.AddStat(stats.MP5, 3)
		case proto.ManaRegenElixir_ElixirOfWickedRegeneration:
			// It also gives 9 health every 5 sec, which the sim doesn't model.
			character.AddStat(stats.MP5, 10)
		}
	}

	// A stat elixir does not stack with the Scroll of the same stat. Only the higher one
	// counts, so with a Scroll of Intellect II (8) already on, the elixir's 6 adds nothing.
	// The buffs went on before the consumes, so we add only what the elixir gives above
	// the scroll.
	if elixir, ok := intellectElixirs[consumes.IntellectElixir]; ok {
		if raidBuffs.GetScrollOfIntellect() && !raidBuffs.GetArcaneBrilliance() {
			elixir -= scrollOfIntellect[HighestRankAt(character.Level, scrollOfIntellectLevel[:])]
		}
		if elixir > 0 {
			character.AddStat(stats.Intellect, elixir)
		}
	}
	if elixir, ok := spiritElixirs[consumes.SpiritElixir]; ok {
		if raidBuffs.GetScrollOfSpirit() && !raidBuffs.GetDivineSpirit() {
			elixir -= scrollOfSpirit[HighestRankAt(character.Level, scrollOfSpiritLevel[:])]
		}
		if elixir > 0 {
			character.AddStat(stats.Spirit, elixir)
		}
	}
}

// The Intellect and Spirit each elixir gives, from client 1.60.1.70338 (Elixir of Wisdom's spell 3166
// from 1.60.1.70707, which cut it from 6 to 3). The familiar scrolls summon a pet, which the sim
// doesn't model.
var intellectElixirs = map[proto.IntellectElixir]float64{
	proto.IntellectElixir_ElixirOfWisdom:           3,
	proto.IntellectElixir_ElixirOfLesserIntellect:  6,
	proto.IntellectElixir_ElixirOfIntellect:        10,
	proto.IntellectElixir_ElixirOfTheOwl:           25,
	proto.IntellectElixir_ElixirOfGreaterIntellect: 25,
}

var spiritElixirs = map[proto.SpiritElixir]float64{
	proto.SpiritElixir_ElixirOfMinorSpirit:   3,
	proto.SpiritElixir_ElixirOfLesserSpirit:  6,
	proto.SpiritElixir_ElixirOfSpirit:        10,
	proto.SpiritElixir_ElixirOfGreaterSpirit: 18,
	proto.SpiritElixir_ElixirOfSages:         25,
	proto.SpiritElixir_ElixirOfTheWhale:      25,
}

///////////////////////////////////////////////////////////////////////////
//                             Zanza-esque Consumes
///////////////////////////////////////////////////////////////////////////

func applyZanzaBuffConsumes(character *Character, consumes *proto.Consumes) {
	if consumes.ZanzaBuff == proto.ZanzaBuff_ZanzaBuffUnknown {
		return
	}

	switch consumes.ZanzaBuff {
	case proto.ZanzaBuff_SpiritOfZanza:
		character.AddStats(stats.Stats{
			stats.Stamina: 50,
			stats.Spirit:  50,
		})
	case proto.ZanzaBuff_ROIDS:
		character.AddStats(stats.Stats{
			stats.Strength: 25,
		})
	case proto.ZanzaBuff_GroundScorpokAssay:
		character.AddStats(stats.Stats{
			stats.Agility: 25,
		})
	case proto.ZanzaBuff_CerebralCortexCompound:
		character.AddStats(stats.Stats{
			stats.Intellect: 25,
		})
	case proto.ZanzaBuff_GizzardGum:
		character.AddStats(stats.Stats{
			stats.Spirit: 25,
		})
	case proto.ZanzaBuff_LungJuiceCocktail:
		character.AddStats(stats.Stats{
			stats.Stamina: 25,
		})
	case proto.ZanzaBuff_DarnassusGiftCollection:
		character.AddStats(stats.Stats{
			stats.Agility: 30,
		})
	}
}

///////////////////////////////////////////////////////////////////////////
//                             Hit Consumables
///////////////////////////////////////////////////////////////////////////

func applyHitConsumableConsumes(character *Character, consumes *proto.Consumes) {
	if consumes.HitConsumable == proto.HitConsumable_HitConsumableUnknown {
		return
	}

	switch consumes.HitConsumable {
	case proto.HitConsumable_FireToastedBun:
		character.AddStats(stats.Stats{
			stats.MeleeHit: 2 * MeleeHitRatingPerHitChance,
		})
	case proto.HitConsumable_DarkDesire:
		character.AddStats(stats.Stats{
			stats.MeleeHit: 2 * MeleeHitRatingPerHitChance,
		})
	}
}

///////////////////////////////////////////////////////////////////////////
//                             Misc Consumes
///////////////////////////////////////////////////////////////////////////

func applyMiscConsumes(character *Character, miscConsumes *proto.MiscConsumes) {
	if miscConsumes == nil {
		return
	}

	if miscConsumes.BoglingRoot {
		character.PseudoStats.BonusPhysicalDamage += 1
	}

	if miscConsumes.RaptorPunch {
		character.AddStats(stats.Stats{
			stats.Intellect: 4,
			stats.Stamina:   -5,
		})
	}

	if miscConsumes.JujuEmber {
		character.AddStat(stats.FireResistance, 15)
	}

	if miscConsumes.JujuChill {
		character.AddStat(stats.FrostResistance, 15)
	}

	if miscConsumes.JujuFlurry && character.AutoAttacks.enabled {
		actionID := ActionID{SpellID: 16322}
		// In Vanilla Juju Flurry was bugged to act like Seal of the Crusader where it gave attack speed but also reduced damage done.
		jujuFlurryAura := character.RegisterAura(Aura{
			Label:    "Juju Flurry",
			ActionID: actionID,
			Duration: time.Second * 20,
			OnGain: func(aura *Aura, sim *Simulation) {
				aura.Unit.MultiplyMeleeSpeed(sim, 1.03)
				aura.Unit.AutoAttacks.MHAuto().DamageMultiplier /= 1.03
				aura.Unit.AutoAttacks.OHAuto().DamageMultiplier /= 1.03
			},
			OnExpire: func(aura *Aura, sim *Simulation) {
				aura.Unit.MultiplyMeleeSpeed(sim, 1/1.03)
				aura.Unit.AutoAttacks.MHAuto().DamageMultiplier *= 1.03
				aura.Unit.AutoAttacks.OHAuto().DamageMultiplier *= 1.03
			},
		})
		jujuFlurrySpell := character.RegisterSpell(SpellConfig{
			ActionID: actionID,
			ProcMask: ProcMaskEmpty,
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute,
				},
				SharedCD: Cooldown{
					Timer:    character.GetAttackSpeedBuffCD(),
					Duration: time.Second * 10,
				},
			},
			ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
				jujuFlurryAura.Activate(sim)
			},
		})
		character.AddMajorCooldown(MajorCooldown{
			Spell:    jujuFlurrySpell,
			Type:     CooldownTypeDPS,
			Priority: CooldownPriorityDefault,
		})
	}

	if miscConsumes.JujuEscape {
		actionID := ActionID{SpellID: 16321}
		jujuEscapeAura := character.RegisterAura(Aura{
			Label:    "Juju Escape",
			ActionID: actionID,
			Duration: time.Second * 10,
			OnGain: func(aura *Aura, sim *Simulation) {
				aura.Unit.AddStatDynamic(sim, stats.Dodge, 5*DodgeRatingPerDodgeChance)
			},
			OnExpire: func(aura *Aura, sim *Simulation) {
				aura.Unit.AddStatDynamic(sim, stats.Dodge, -5*DodgeRatingPerDodgeChance)
			},
		})
		jujuEscapeSpell := character.RegisterSpell(SpellConfig{
			ActionID: actionID,
			ProcMask: ProcMaskEmpty,
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute,
				},
			},
			ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
				jujuEscapeAura.Activate(sim)
			},
		})
		character.AddMajorCooldown(MajorCooldown{
			Spell:    jujuEscapeSpell,
			Type:     CooldownTypeSurvival,
			Priority: CooldownPriorityDefault,
		})
	}
}

///////////////////////////////////////////////////////////////////////////
//                             Engineering Explosives
///////////////////////////////////////////////////////////////////////////

var SapperActionID = ActionID{ItemID: 10646}
var SolidDynamiteActionID = ActionID{ItemID: 10507}
var DenseDynamiteActionID = ActionID{ItemID: 18641}
var ThoriumGrenadeActionID = ActionID{ItemID: 15993}
var GoblinLandMineActionID = ActionID{ItemID: 4395}
var EZThroDarkBombActionID = ActionID{ItemID: 260817}
var EZThroThoriumGrenadeActionID = ActionID{ItemID: 260816}
var HolyHandGrenadeActionID = ActionID{ItemID: 284868}

// Forever's explosives that need no profession, on the 60 sec category (24) every explosive
// shares. Each casts in 1 sec, like the Engineering ones, and needs its item's level:
//   - EZ-Thro Dark Bomb 260817 (1269334): 450 Fire ±50% (Variance 1), 225 to 675, level 60.
//   - EZ-Thro Thorium Grenade 260816 (1269330): 400 Fire ±25% (Variance 0.5), 300 to 500, level 57.
//   - Holy Hand Grenade of Alterac 284868 (1317268): 400 Holy ±25%, 300 to 500, level 52.
//
// The two EZ-Thro rows lack SPELL_ATTR0_NOT_SHAPESHIFTED (Attributes[0] 0x10000), which every
// Engineering explosive's row but the sapper's and the Holy Hand Grenade's carry, so a druid
// throws them without leaving its form (wowsims/forever's NewSpellDataExplosive).
type openExplosive struct {
	actionID      ActionID
	school        SpellSchool
	min, max      float64
	requiredLevel int32
	inForm        bool
}

var openExplosives = map[proto.Explosive]openExplosive{
	proto.Explosive_ExplosiveEZThroDarkBomb:       {EZThroDarkBombActionID, SpellSchoolFire, 225, 675, 60, true},
	proto.Explosive_ExplosiveEZThroThoriumGrenade: {EZThroThoriumGrenadeActionID, SpellSchoolFire, 300, 500, 57, true},
	proto.Explosive_ExplosiveHolyHandGrenade:      {HolyHandGrenadeActionID, SpellSchoolHoly, 300, 500, 52, false},
}

func registerExplosivesCD(agent Agent, consumes *proto.Consumes) {
	character := agent.GetCharacter()
	hasFiller := consumes.FillerExplosive != proto.Explosive_ExplosiveUnknown
	hasSapper := consumes.SapperExplosive != proto.SapperExplosive_SapperUnknown

	if !hasSapper && !hasFiller {
		return
	}
	sharedTimer := character.NewTimer()

	if hasSapper && character.HasProfession(proto.Profession_Engineering) {
		var filler *Spell
		switch consumes.SapperExplosive {
		case proto.SapperExplosive_SapperGoblinSapper:
			filler = character.newSapperSpell(sharedTimer)
		}
		character.AddMajorCooldown(MajorCooldown{
			Spell:    filler,
			Type:     CooldownTypeDPS | CooldownTypeExplosive,
			Priority: CooldownPriorityLow + 30,
			ShouldActivate: func(s *Simulation, c *Character) bool {
				return !character.IsShapeshifted()
			},
		})
	}

	if explosive, ok := openExplosives[consumes.FillerExplosive]; ok {
		if character.Level < explosive.requiredLevel {
			return
		}
		config := character.newBasicExplosiveSpellConfig(sharedTimer, explosive.actionID, explosive.school, explosive.min, explosive.max, Cooldown{}, 0, 0)
		if explosive.inForm {
			config.Cast.ModifyCast = nil
		}
		character.AddMajorCooldown(MajorCooldown{
			Spell:    character.GetOrRegisterSpell(config),
			Type:     CooldownTypeDPS | CooldownTypeExplosive,
			Priority: CooldownPriorityLow + 10,
			ShouldActivate: func(s *Simulation, c *Character) bool {
				return explosive.inForm || !character.IsShapeshifted()
			},
		})
		return
	}

	if hasFiller {
		if !character.HasProfession(proto.Profession_Engineering) {
			return
		}

		var filler *Spell
		switch consumes.FillerExplosive {
		case proto.Explosive_ExplosiveSolidDynamite:
			filler = character.newSolidDynamiteSpell(sharedTimer)
		case proto.Explosive_ExplosiveDenseDynamite:
			filler = character.newDenseDynamiteSpell(sharedTimer)
		case proto.Explosive_ExplosiveThoriumGrenade:
			filler = character.newThoriumGrenadeSpell(sharedTimer)
		case proto.Explosive_ExplosiveGoblinLandMine:
			filler = character.newGoblinLandMineSpell(sharedTimer)
		}

		character.AddMajorCooldown(MajorCooldown{
			Spell:    filler,
			Type:     CooldownTypeDPS | CooldownTypeExplosive,
			Priority: CooldownPriorityLow + 10,
			ShouldActivate: func(s *Simulation, c *Character) bool {
				return !character.IsShapeshifted()
			},
		})
	}
}

// Creates a spell object for the common explosive case.
// TODO: create 10s delay on Goblin Landmine cast to damage
func (character *Character) newBasicExplosiveSpellConfig(sharedTimer *Timer, actionID ActionID, school SpellSchool, minDamage float64, maxDamage float64, cooldown Cooldown, selfMinDamage float64, selfMaxDamage float64) SpellConfig {
	isSapper := actionID.SameAction(SapperActionID)

	var defaultCast Cast
	if !isSapper {
		defaultCast = Cast{
			CastTime: time.Second,
		}
	}

	return SpellConfig{
		ActionID:    actionID,
		SpellSchool: school,
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskEmpty,
		Flags:       SpellFlagCastTimeNoGCD,

		Cast: CastConfig{
			DefaultCast: defaultCast,
			CD:          cooldown,
			IgnoreHaste: true,
			SharedCD: Cooldown{
				Timer:    sharedTimer,
				Duration: time.Minute,
			},
			ModifyCast: func(sim *Simulation, _ *Spell, _ *Cast) {
				character.CancelShapeshift(sim)
			},
		},

		// Explosives always have 1% resist chance, so just give them hit cap.
		BonusHitRating: 100 * SpellHitRatingPerHitChance,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
				baseDamage := sim.Roll(minDamage, maxDamage) * sim.Encounter.AOECapMultiplier()
				spell.CalcAndDealDamage(sim, aoeTarget, baseDamage, spell.OutcomeMagicHitAndCrit)
			}

			if isSapper {
				baseDamage := sim.Roll(selfMinDamage, selfMaxDamage)
				spell.CalcAndDealDamage(sim, &character.Unit, baseDamage, spell.OutcomeMagicHitAndCrit)
			}
		},
	}
}
func (character *Character) newSapperSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, SapperActionID, SpellSchoolFire, 450, 750, Cooldown{Timer: character.NewTimer(), Duration: time.Minute * 5}, 375, 625))
}

func (character *Character) newSolidDynamiteSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, SolidDynamiteActionID, SpellSchoolFire, 213, 287, Cooldown{}, 0, 0))
}
func (character *Character) newDenseDynamiteSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, DenseDynamiteActionID, SpellSchoolFire, 340, 460, Cooldown{}, 0, 0))
}
func (character *Character) newThoriumGrenadeSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, ThoriumGrenadeActionID, SpellSchoolFire, 300, 500, Cooldown{}, 0, 0))
}
func (character *Character) newGoblinLandMineSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, GoblinLandMineActionID, SpellSchoolFire, 394, 506, Cooldown{}, 0, 0))
}

///////////////////////////////////////////////////////////////////////////
//                             Potions
///////////////////////////////////////////////////////////////////////////

func registerPotionCD(agent Agent, consumes *proto.Consumes) {
	character := agent.GetCharacter()
	defaultPotion := consumes.DefaultPotion

	potionCD := character.NewTimer()

	if defaultPotion == proto.Potions_UnknownPotion {
		return
	}

	defaultMCD := makePotionActivation(defaultPotion, character, potionCD)

	if defaultMCD.Spell != nil {
		character.AddMajorCooldown(defaultMCD)
	}
}

func makePotionActivation(potionType proto.Potions, character *Character, potionCD *Timer) MajorCooldown {
	mcd := makePotionActivationInternal(potionType, character, potionCD)
	if mcd.Spell != nil {
		// Mark as 'Encounter Only' so that users are forced to select the generic Potion
		// placeholder action instead of specific potion spells, in APL prepull. This
		// prevents a mismatch between Consumes and Rotation settings.
		mcd.Spell.Flags |= SpellFlagEncounterOnly | SpellFlagPotion | SpellFlagCastTimeNoGCD
		oldApplyEffects := mcd.Spell.ApplyEffects
		mcd.Spell.ApplyEffects = func(sim *Simulation, target *Unit, spell *Spell) {
			oldApplyEffects(sim, target, spell)
			if sim.CurrentTime < 0 {
				potionCD.Set(sim.CurrentTime + 2*time.Minute)
				character.UpdateMajorCooldowns()
			}
		}
	}
	return mcd
}

func makeHealthConsumableMCD(itemId int32, character *Character, cdTimer *Timer) MajorCooldown {
	// Using min values for healthstones as locks generally don't spec into improved
	minRoll := map[int32]float64{
		858:   140,
		929:   280,
		1710:  455,
		3928:  700,
		5509:  500,
		5510:  800,
		9421:  1200,
		13446: 1050,
	}[itemId]

	maxRoll := map[int32]float64{
		858:   180,
		929:   360,
		1710:  585,
		3928:  900,
		5509:  500,
		5510:  800,
		9421:  1200,
		13446: 1750,
	}[itemId]

	// Forever's healthstones restore 20% more.
	if character.Env.IsForever() {
		if forever, ok := map[int32]float64{5509: 600, 5510: 960, 9421: 1440}[itemId]; ok {
			minRoll, maxRoll = forever, forever
		}
	}

	cdDuration := time.Minute * 2

	actionID := ActionID{ItemID: itemId}
	healthMetrics := character.NewHealthMetrics(actionID)

	return MajorCooldown{
		Type: CooldownTypeSurvival,
		ShouldActivate: func(sim *Simulation, character *Character) bool {
			// Only pop if we have less than the max mana provided by the potion minus 1mp5 tick.
			return (character.MaxHealth()-(character.CurrentHealth()) >= maxRoll) && character.canUseItemsInForm()
		},
		Spell: character.GetOrRegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    cdTimer,
					Duration: cdDuration,
				},
				ModifyCast: func(sim *Simulation, _ *Spell, _ *Cast) {
					if !character.Env.IsForever() {
						character.CancelShapeshift(sim)
					}
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				healthGain := sim.RollWithLabel(minRoll, maxRoll, "Health Consumable")
				character.GainHealth(sim, healthGain, healthMetrics)
			},
		}),
	}
}

func makeManaConsumableMCD(itemId int32, character *Character, cdTimer *Timer) MajorCooldown {
	minRoll := map[int32]float64{
		3385:  280.0,
		3827:  455.0,
		6149:  700.0,
		4381:  150.0,
		12662: 900.0,
		13443: 900.0,
		13444: 1350.0,
	}[itemId]

	maxRoll := map[int32]float64{
		3385:  360.0,
		3827:  585.0,
		6149:  900.0,
		4381:  250.0,
		12662: 1500.0,
		13443: 1500.0,
		13444: 2250.0,
	}[itemId]

	cdDuration := time.Minute * 2
	if itemId == 4381 {
		// Minor Recombobulator is a trinket with a 5 min cooldown. Forever cut its mana to
		// 96 to 160 (wowhead Forever).
		cdDuration = time.Minute * 5
		if character.Env.IsForever() {
			minRoll, maxRoll = 96, 160
		}
	}

	actionID := ActionID{ItemID: itemId}
	manaMetrics := character.NewManaMetrics(actionID)

	return MajorCooldown{
		Type: CooldownTypeMana,
		ShouldActivate: func(sim *Simulation, character *Character) bool {
			// Only pop if we have less than the max mana provided by the potion minus 1mp5 tick.
			totalRegen := character.ManaRegenPerSecondWhileCasting() * 2
			return (character.MaxMana()-(character.CurrentMana()+totalRegen) >= maxRoll) && character.canUseItemsInForm()
		},
		// Regen only adds mana, so we only come to miss enough of it when we spend some.
		ShouldActivateNeedsEvent: true,
		Spell: character.GetOrRegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    cdTimer,
					Duration: cdDuration,
				},
				ModifyCast: func(sim *Simulation, _ *Spell, _ *Cast) {
					if !character.Env.IsForever() {
						character.CancelShapeshift(sim)
					}
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				manaGain := sim.RollWithLabel(minRoll, maxRoll, "Mana Consumable")
				character.AddMana(sim, manaGain, manaMetrics)
			},
		}),
	}
}

func makeArmorConsumableMCD(itemId int32, character *Character, cdTimer *Timer) MajorCooldown {
	actionID := ActionID{ItemID: itemId}
	cdDuration := time.Minute * 2
	lesserStoneshieldAura := character.NewTemporaryStatsAura("Lesser Stoneshield Potion", actionID, stats.Stats{stats.BonusArmor: 1000}, time.Second*90)
	greaterStoneshieldAura := character.NewTemporaryStatsAura("Greater Stoneshield Potion", actionID, stats.Stats{stats.BonusArmor: 2000}, time.Second*120)

	return MajorCooldown{
		Type: CooldownTypeSurvival,
		Spell: character.GetOrRegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    cdTimer,
					Duration: cdDuration,
				},
				ModifyCast: func(sim *Simulation, _ *Spell, _ *Cast) {
					if !character.Env.IsForever() {
						character.CancelShapeshift(sim)
					}
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				switch itemId {
				case 4623:
					lesserStoneshieldAura.Activate(sim)
				case 13455:

					greaterStoneshieldAura.Activate(sim)
				}
			},
		}),
	}
}

func makeMagicResistancePotionMCD(character *Character, cdTimer *Timer) MajorCooldown {
	actionID := ActionID{ItemID: 4623}
	cdDuration := time.Minute * 2

	stats := stats.Stats{
		stats.ArcaneResistance: 50,
		stats.FireResistance:   50,
		stats.FrostResistance:  50,
		stats.NatureResistance: 50,
		stats.ShadowResistance: 50,
	}

	// Since many people will keep this rolling as a substitute for capping Fire Resistance, show the stats as a baseline
	aura := character.NewTemporaryStatsAura("Magic Resistance Potion", actionID, stats, time.Minute*3)
	aura.BuildPhase = CharacterBuildPhaseConsumes

	return MajorCooldown{
		Type: CooldownTypeSurvival,
		Spell: character.GetOrRegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    cdTimer,
					Duration: cdDuration,
				},
				ModifyCast: func(sim *Simulation, _ *Spell, _ *Cast) {
					if !character.Env.IsForever() {
						character.CancelShapeshift(sim)
					}
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				aura.Activate(sim)
			},
		}),
	}
}

func makeRageConsumableMCD(itemId int32, character *Character, cdTimer *Timer) MajorCooldown {
	minRoll := map[int32]float64{
		5631:  20.0,
		5633:  30.0,
		13442: 45.0,
	}[itemId]

	maxRoll := map[int32]float64{
		5631:  40.0,
		5633:  60.0,
		13442: 75.0,
	}[itemId]

	cdDuration := time.Minute * 2

	actionID := ActionID{ItemID: itemId}
	rageMetrics := character.NewRageMetrics(actionID)
	aura := character.NewTemporaryStatsAura("Mighty Rage Potion", actionID, stats.Stats{stats.Strength: 60}, time.Second*20)
	return MajorCooldown{
		Type: CooldownTypeDPS,
		ShouldActivate: func(sim *Simulation, character *Character) bool {
			return character.canUseItemsInForm()
		},
		Spell: character.GetOrRegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    cdTimer,
					Duration: cdDuration,
				},
				ModifyCast: func(sim *Simulation, _ *Spell, _ *Cast) {
					if !character.Env.IsForever() {
						character.CancelShapeshift(sim)
					}
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				rageGain := sim.RollWithLabel(minRoll, maxRoll, "Rage Consumable")
				character.AddRage(sim, rageGain, rageMetrics)
				if itemId == 13442 {
					aura.Activate(sim)
				}
			},
		}),
	}
}

// makeStatBuffPotionMCD is one of Forever's potions that raise a stat for 30 sec.
//
// Their tooltips list no cooldown, but we assume they share the 2 min potion cooldown
// with every other potion, the way Classic's potions do.
func makeStatBuffPotionMCD(itemId int32, label string, buff stats.Stats, character *Character, cdTimer *Timer) MajorCooldown {
	actionID := ActionID{ItemID: itemId}
	aura := character.NewTemporaryStatsAura(label, actionID, buff, time.Second*30)
	return MajorCooldown{
		Type: CooldownTypeDPS,
		ShouldActivate: func(sim *Simulation, character *Character) bool {
			return character.canUseItemsInForm()
		},
		Spell: character.GetOrRegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    cdTimer,
					Duration: time.Minute * 2,
				},
				ModifyCast: func(sim *Simulation, _ *Spell, _ *Cast) {
					if !character.Env.IsForever() {
						character.CancelShapeshift(sim)
					}
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				aura.Activate(sim)
			},
		}),
	}
}

// makeDragonfirePotionMCD is Forever's Dragonfire Potion 250953 (1252554, category 4 with the
// other potions, 2 min): 280 Fire to every enemy in a 12 yd cone, and 35 Fire every 2 sec for 8
// sec on each. The row states Fixed Travel Time at Speed 1, so the fire lands 1 sec after use
// from any distance, and Always Hit, so it rolls no miss; the hit can crit (wowsims/forever
// 5b8a1fbc24). It has no spell power coefficient. It reaches only as far as the cone's radius.
func makeDragonfirePotionMCD(character *Character, cdTimer *Timer) MajorCooldown {
	const coneRadius = 12
	actionID := ActionID{ItemID: 250953}
	var spell *Spell
	spell = character.GetOrRegisterSpell(SpellConfig{
		ActionID:    actionID,
		SpellSchool: SpellSchoolFire,
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskEmpty,
		Flags:       SpellFlagNoOnCastComplete,
		Cast: CastConfig{
			CD: Cooldown{
				Timer:    cdTimer,
				Duration: time.Minute * 2,
			},
		},
		ExtraCastCondition: func(_ *Simulation, _ *Unit) bool {
			return character.DistanceFromTarget <= coneRadius
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: DotConfig{
			Aura: Aura{
				Label: "Dragonfire Potion",
			},
			NumberOfTicks: 4,
			TickLength:    time.Second * 2,
			OnTick: func(sim *Simulation, target *Unit, dot *Dot) {
				dot.Spell.CalcAndDealPeriodicDamage(sim, target, 35, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
			StartDelayedAction(sim, DelayedActionOptions{
				DoAt: sim.CurrentTime + time.Second,
				OnAction: func(sim *Simulation) {
					for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
						spell.CalcAndDealDamage(sim, aoeTarget, 280*sim.Encounter.AOECapMultiplier(), spell.OutcomeMagicCrit)
						spell.Dot(aoeTarget).Apply(sim)
					}
				},
			})
		},
	})
	return MajorCooldown{
		Type: CooldownTypeDPS,
		ShouldActivate: func(sim *Simulation, character *Character) bool {
			return character.canUseItemsInForm()
		},
		Spell: spell,
	}
}

// makeCausticSmogPotionMCD is Forever's Caustic Smog Potion 250951 (category 4 with the other
// potions, 2 min). Its 1252525 is a missile, range 20, at 10 yd/sec for no less than 1.5 sec,
// that leaves the cloud 1252526 for 12 sec, which puts 1252543 on every enemy inside it: 33
// Nature a second, rolled each tick from 31 to 35 (Variance 0.1). 1252543 states Always Hit, so
// no tick misses (wowsims/forever 5b8a1fbc24). It has no spell power coefficient.
func makeCausticSmogPotionMCD(character *Character, cdTimer *Timer) MajorCooldown {
	const maxRange = 20
	actionID := ActionID{ItemID: 250951}
	var spell *Spell
	spell = character.GetOrRegisterSpell(SpellConfig{
		ActionID:     actionID,
		SpellSchool:  SpellSchoolNature,
		DefenseType:  DefenseTypeMagic,
		ProcMask:     ProcMaskEmpty,
		Flags:        SpellFlagNoOnCastComplete,
		MissileSpeed: 10,
		Cast: CastConfig{
			CD: Cooldown{
				Timer:    cdTimer,
				Duration: time.Minute * 2,
			},
		},
		ExtraCastCondition: func(_ *Simulation, _ *Unit) bool {
			return character.DistanceFromTarget <= maxRange
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: DotConfig{
			Aura: Aura{
				Label: "Caustic Smog Potion",
			},
			NumberOfTicks: 12,
			TickLength:    time.Second,
			OnTick: func(sim *Simulation, target *Unit, dot *Dot) {
				dot.Spell.CalcAndDealPeriodicDamage(sim, target, sim.Roll(31, 35), dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
			StartDelayedAction(sim, DelayedActionOptions{
				DoAt: sim.CurrentTime + max(spell.TravelTime(), time.Millisecond*1500),
				OnAction: func(sim *Simulation) {
					for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
						spell.Dot(aoeTarget).Apply(sim)
					}
				},
			})
		},
	})
	return MajorCooldown{
		Type: CooldownTypeDPS,
		ShouldActivate: func(sim *Simulation, character *Character) bool {
			return character.canUseItemsInForm()
		},
		Spell: spell,
	}
}

func makePotionActivationInternal(potionType proto.Potions, character *Character, potionCD *Timer) MajorCooldown {
	if potionType == proto.Potions_UnknownPotion {
		return MajorCooldown{}
	}

	switch potionType {
	case proto.Potions_LesserHealingPotion:
		return makeHealthConsumableMCD(858, character, potionCD)
	case proto.Potions_HealingPotion:
		return makeHealthConsumableMCD(929, character, potionCD)
	case proto.Potions_GreaterHealingPotion:
		return makeHealthConsumableMCD(1710, character, potionCD)
	case proto.Potions_SuperiorHealingPotion:
		return makeHealthConsumableMCD(3928, character, potionCD)
	case proto.Potions_MajorHealingPotion:
		return makeHealthConsumableMCD(13446, character, potionCD)

	case proto.Potions_LesserManaPotion:
		return makeManaConsumableMCD(3385, character, potionCD)
	case proto.Potions_ManaPotion:
		return makeManaConsumableMCD(3827, character, potionCD)
	case proto.Potions_GreaterManaPotion:
		return makeManaConsumableMCD(6149, character, potionCD)
	case proto.Potions_SuperiorManaPotion:
		return makeManaConsumableMCD(13443, character, potionCD)
	case proto.Potions_MajorManaPotion:
		return makeManaConsumableMCD(13444, character, potionCD)

	case proto.Potions_RagePotion:
		return makeRageConsumableMCD(5631, character, potionCD)
	case proto.Potions_GreatRagePotion:
		return makeRageConsumableMCD(5633, character, potionCD)
	case proto.Potions_MightyRagePotion:
		return makeRageConsumableMCD(13442, character, potionCD)

	case proto.Potions_LesserStoneshieldPotion:
		return makeArmorConsumableMCD(4623, character, potionCD)
	case proto.Potions_GreaterStoneshieldPotion:
		return makeArmorConsumableMCD(13455, character, potionCD)

	case proto.Potions_MagicResistancePotion:
		return makeMagicResistancePotionMCD(character, potionCD)

	case proto.Potions_MinorFrenzyPotion:
		return makeStatBuffPotionMCD(250938, "Minor Frenzy Potion", meleeAndRangedAP(12), character, potionCD)
	case proto.Potions_LesserFrenzyPotion:
		return makeStatBuffPotionMCD(250939, "Lesser Frenzy Potion", meleeAndRangedAP(18), character, potionCD)
	case proto.Potions_FrenzyPotion:
		return makeStatBuffPotionMCD(250940, "Frenzy Potion", meleeAndRangedAP(28), character, potionCD)
	case proto.Potions_GreaterFrenzyPotion:
		return makeStatBuffPotionMCD(250941, "Greater Frenzy Potion", meleeAndRangedAP(40), character, potionCD)
	case proto.Potions_SuperiorFrenzyPotion:
		return makeStatBuffPotionMCD(250942, "Superior Frenzy Potion", meleeAndRangedAP(56), character, potionCD)
	case proto.Potions_MajorFrenzyPotion:
		return makeStatBuffPotionMCD(250943, "Major Frenzy Potion", meleeAndRangedAP(80), character, potionCD)
	case proto.Potions_MinorSpellblastingPotion:
		return makeStatBuffPotionMCD(250932, "Minor Spellblasting Potion", stats.Stats{stats.SpellDamage: 7}, character, potionCD)
	case proto.Potions_LesserSpellblastingPotion:
		return makeStatBuffPotionMCD(250933, "Lesser Spellblasting Potion", stats.Stats{stats.SpellDamage: 11}, character, potionCD)
	case proto.Potions_SpellblastingPotion:
		return makeStatBuffPotionMCD(250934, "Spellblasting Potion", stats.Stats{stats.SpellDamage: 16}, character, potionCD)
	case proto.Potions_GreaterSpellblastingPotion:
		return makeStatBuffPotionMCD(250935, "Greater Spellblasting Potion", stats.Stats{stats.SpellDamage: 23}, character, potionCD)
	case proto.Potions_SuperiorSpellblastingPotion:
		return makeStatBuffPotionMCD(250936, "Superior Spellblasting Potion", stats.Stats{stats.SpellDamage: 33}, character, potionCD)
	case proto.Potions_MajorSpellblastingPotion:
		return makeStatBuffPotionMCD(250937, "Major Spellblasting Potion", stats.Stats{stats.SpellDamage: 47}, character, potionCD)
	case proto.Potions_DragonfirePotion:
		return makeDragonfirePotionMCD(character, potionCD)
	case proto.Potions_CausticSmogPotion:
		return makeCausticSmogPotionMCD(character, potionCD)
	// case proto.Potions_GreaterArcaneProtectionPotion:
	// 	return makeSchoolProtectionConsumableMCD(13461, character, potionCD)
	// case proto.Potions_GreaterFireProtectionPotion:
	// 	return makeSchoolProtectionConsumableMCD(13457, character, potionCD)
	// case proto.Potions_GreaterFrostProtectionPotion:
	// 	return makeSchoolProtectionConsumableMCD(13456, character, potionCD)
	// case proto.Potions_GreaterFrostProtectionPotion:
	// 	return makeSchoolProtectionConsumableMCD(13460, character, potionCD)
	// case proto.Potions_GreaterFrostProtectionPotion:
	// 	return makeSchoolProtectionConsumableMCD(13458, character, potionCD)
	// case proto.Potions_GreaterFrostProtectionPotion:
	// 	return makeSchoolProtectionConsumableMCD(13459, character, potionCD)
	default:
		return MajorCooldown{}
	}
}

func registerConjuredCD(agent Agent, consumes *proto.Consumes) {
	character := agent.GetCharacter()
	conjuredType := consumes.DefaultConjured

	if conjuredType == proto.Conjured_ConjuredUnknown {
		return
	}

	timer := character.GetConjuredCD()

	var mcd MajorCooldown
	switch conjuredType {
	case proto.Conjured_ConjuredHealthstone:
		mcd = makeHealthConsumableMCD(5509, character, timer)
	case proto.Conjured_ConjuredGreaterHealthstone:
		mcd = makeHealthConsumableMCD(5510, character, timer)
	case proto.Conjured_ConjuredMajorHealthstone:
		mcd = makeHealthConsumableMCD(9421, character, timer)
	case proto.Conjured_ConjuredDemonicRune:
		mcd = makeManaConsumableMCD(12662, character, timer)
	case proto.Conjured_ConjuredMinorRecombobulator:
		mcd = makeManaConsumableMCD(4381, character, timer)
	// Handled in the rogue package
	// case proto.Conjured_ConjuredRogueThistleTea:
	default:
		return
	}

	character.AddMajorCooldown(mcd)
}
