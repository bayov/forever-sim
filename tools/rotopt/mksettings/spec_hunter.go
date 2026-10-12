package main

import "github.com/wowsims/classic/sim/core/proto"

// The three hunter specs share ui/hunter. Beast Mastery and Marksmanship shoot from range (25
// yards), Survival is the melee spec under Forever (Strider Kick, Mongoose Bite through Expose
// Prey, Lacerating Strikes, Predator's Edge) and stands at the boss.
func init() {
	specs["hunter_bm"] = hunterSpec(25, hunterRangedRaid)
	specs["hunter_mm"] = hunterSpec(25, hunterRangedRaid)
	specs["hunter_sv"] = hunterSpec(5, hunterMeleeRaid)
}

// A ranged hunter's raid: its own Hunter's Mark, Blessing of Wisdom next to Might (a mana user),
// and a shaman in its group for Grace of Air (Agility is two ranged attack power and crit) and
// Mana Spring. Trueshot Aura only from a Marksmanship hunter's own talent (Forever makes it a party
// aura of ranged attack power), not from the raid. The melee debuffs and Faerie Fire stay, the
// physical shots go through armor.
func hunterRangedRaid(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, debuffs *proto.Debuffs) {
	raidBuffs.TrueshotAura = false
	raidBuffs.GraceOfAirTotem = proto.TristateEffect_TristateEffectImproved
	raidBuffs.ManaSpringTotem = proto.TristateEffect_TristateEffectRegular
	debuffs.HuntersMark = proto.TristateEffect_TristateEffectImproved
}

// Survival melees, so it joins a melee group: Windfury Totem (its own one air totem, Forever stops
// two from stacking) and Strength of Earth, plus Mana Spring.
func hunterMeleeRaid(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, debuffs *proto.Debuffs) {
	raidBuffs.TrueshotAura = false
	raidBuffs.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffWindfury
	raidBuffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectImproved
	raidBuffs.ManaSpringTotem = proto.TristateEffect_TristateEffectRegular
	debuffs.HuntersMark = proto.TristateEffect_TristateEffectImproved
}

func hunterSpec(distance float64, raid func(*proto.RaidBuffs, *proto.IndividualBuffs, *proto.Debuffs)) *specSetup {
	return &specSetup{
		uiDir: "ui/hunter",
		raid:  raid,
		consumes: func() *proto.Consumes {
			return &proto.Consumes{
				AgilityElixir:            proto.AgilityElixir_ElixirOfTheMongoose,
				AttackPowerBuff:          proto.AttackPowerBuff_JujuMight,
				DefaultPotion:            proto.Potions_MajorManaPotion,
				DefaultConjured:          proto.Conjured_ConjuredDemonicRune,
				Food:                     proto.Food_FoodGrilledSquid,
				ManaRegenElixir:          proto.ManaRegenElixir_MagebloodPotion,
				StrengthBuff:             proto.StrengthBuff_JujuPower,
				ZanzaBuff:                proto.ZanzaBuff_GroundScorpokAssay,
				PetAttackPowerConsumable: 1,
				PetAgilityConsumable:     1,
				PetStrengthConsumable:    1,
			}
		},
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
