package main

import "github.com/wowsims/classic/sim/core/proto"

// The three hunter specs share ui/hunter. Beast Mastery and Marksmanship shoot from range (25
// yards), Survival is the melee spec under Forever (Strider Kick, Mongoose Bite through Expose
// Prey, Lacerating Strikes, Predator's Edge) and stands at the boss.
func init() {
	specs["hunter_bm"] = hunterSpec(25, false, hunterRangedRaid)
	specs["hunter_mm"] = hunterSpec(25, false, hunterRangedRaid)
	specs["hunter_sv"] = hunterSpec(5, true, hunterMeleeRaid)
}

// A ranged hunter's raid: its own Hunter's Mark, Blessing of Wisdom next to Might (a mana user),
// and a shaman in its group for Grace of Air (Agility is two ranged attack power and crit) and
// Mana Spring. Trueshot Aura only from a Marksmanship hunter's own talent (Forever makes it a party
// aura of ranged attack power), not from the raid. The melee debuffs and Faerie Fire stay, the
// physical shots go through armor.
func hunterRangedRaid(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, debuffs *proto.Debuffs) {
	raidBuffs.TrueshotAura = false
	raidBuffs.GraceOfAirTotem = proto.TristateEffect_TristateEffectRegular
	raidBuffs.ManaSpringTotem = proto.TristateEffect_TristateEffectRegular
	debuffs.HuntersMark = proto.TristateEffect_TristateEffectImproved
}

// Survival melees, so it joins a melee group: Windfury Totem (its own one air totem, Forever stops
// two from stacking) and Strength of Earth, plus Mana Spring.
func hunterMeleeRaid(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, debuffs *proto.Debuffs) {
	raidBuffs.TrueshotAura = false
	raidBuffs.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffWindfury
	raidBuffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectRegular
	raidBuffs.ManaSpringTotem = proto.TristateEffect_TristateEffectRegular
	debuffs.HuntersMark = proto.TristateEffect_TristateEffectImproved
}

func hunterSpec(distance float64, melee bool, raid func(*proto.RaidBuffs, *proto.IndividualBuffs, *proto.Debuffs)) *specSetup {
	return &specSetup{
		uiDir: "ui/hunter",
		raid:  raid,
		consumes: func() *proto.Consumes {
			return hunterConsumes(melee)
		},
		professions: [2]proto.Profession{proto.Profession_Engineering, proto.Profession_Alchemy},
		withSpec: func(player *proto.Player) {
			player.Class = proto.Class_ClassHunter
			player.DistanceFromTarget = distance
			player.Spec = &proto.Player_Hunter{Hunter: &proto.Hunter{
				Options: &proto.Hunter_Options{
					Ammo:           proto.Hunter_Options_ThoriumHeadedArrow,
					QuiverBonus:    proto.Hunter_Options_Speed15,
					PetType:        proto.Hunter_Options_Cat,
					PetAttackSpeed: proto.Hunter_Options_One,
					PetUptime:      1,
				},
			}}
		},
	}
}

// The level 60 consumes. On 2026-10-12 we tried every consumable the sim applies in each slot,
// one at a time, on the Beast Mastery and Survival builds at 180 sec. Both keep Elixir of the
// Mongoose (+27 over none, +13 over Greater Agility), Ground Scorpok Assay (+14, +11 for
// Survival), Juju Might (+8), Grilled Squid (+6 over Smoked Desert Dumpling), Elixir of Wisdom
// (+1.4) and Engineering's Goblin Land Mine (+7.6) and Goblin Sapper (+4.8). The ranged specs
// also need Major Mana Potion (+31), Demonic Rune (+28) and Mageblood (+2.4), and Flask of
// Distilled Wisdom is +1.1. Survival never runs low on mana and gets nothing from those, but
// Juju Power's Strength (+6.7) and Dragonbreath Chili (+3.2). Fire-toasted Bun's 2% hit
// (+9) is left out: Forever's is a seasonal item that expires 14 days after looting (item 23327). The pet gets the pet scrolls and Juju.
func hunterConsumes(melee bool) *proto.Consumes {
	consumes := &proto.Consumes{
		AgilityElixir:            proto.AgilityElixir_ElixirOfTheMongoose,
		AttackPowerBuff:          proto.AttackPowerBuff_JujuMight,
		FillerExplosive:          proto.Explosive_ExplosiveGoblinLandMine,
		Food:                     proto.Food_FoodGrilledSquid,
		IntellectElixir:          proto.IntellectElixir_ElixirOfWisdom,
		SapperExplosive:          proto.SapperExplosive_SapperGoblinSapper,
		ZanzaBuff:                proto.ZanzaBuff_GroundScorpokAssay,
		PetAttackPowerConsumable: 1,
		PetAgilityConsumable:     1,
		PetStrengthConsumable:    1,
	}
	if melee {
		consumes.StrengthBuff = proto.StrengthBuff_JujuPower
		consumes.DragonBreathChili = true
	} else {
		consumes.DefaultPotion = proto.Potions_MajorManaPotion
		consumes.DefaultConjured = proto.Conjured_ConjuredDemonicRune
		consumes.ManaRegenElixir = proto.ManaRegenElixir_MagebloodPotion
		consumes.Flask = proto.Flask_FlaskOfDistilledWisdom
	}
	return consumes
}
