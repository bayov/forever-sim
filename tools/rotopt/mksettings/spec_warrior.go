package main

import "github.com/wowsims/classic/sim/core/proto"

// The DPS warrior is in a melee group with a shaman, who drops Strength of Earth and Windfury Totem,
// and the raid's second shaman adds Grace of Air (the user's ruling 2026-10-16, "both totems": client 1.60.1.70425 lets Grace of Air stack with Windfury Totem and makes both raid-wide). The paladins bless it with Kings and Might. The shared raid already has Battle Shout from another warrior, Sunder Armor,
// Faerie Fire, Curse of Recklessness and Expose Armor.
func init() {
	warriorRaid := func(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, _ *proto.Debuffs) {
		raidBuffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectRegular
		raidBuffs.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffWindfury
		raidBuffs.GraceOfAirTotem = proto.TristateEffect_TristateEffectRegular
		// Kings and Might, no Wisdom for a Rage user.
		buffs.BlessingOfWisdom = proto.TristateEffect_TristateEffectMissing
	}
	warriorConsumes := func(twoHand bool) func() *proto.Consumes {
		return func() *proto.Consumes {
			consumes := &proto.Consumes{
				AgilityElixir:     proto.AgilityElixir_ElixirOfTheMongoose,
				AttackPowerBuff:   proto.AttackPowerBuff_JujuMight,
				DefaultPotion:     proto.Potions_MightyRagePotion,
				DragonBreathChili: true,
				Flask:             proto.Flask_FlaskOfNaturalPrecision,
				Food:              proto.Food_FoodSmokedDesertDumpling,
				MainHandImbue:     proto.WeaponImbue_DenseSharpeningStone,
				OffHandImbue:      proto.WeaponImbue_DenseSharpeningStone,
				SapperExplosive:   proto.SapperExplosive_SapperGoblinSapper,
				StrengthBuff:      proto.StrengthBuff_ElixirOfFerocity,
				ZanzaBuff:         proto.ZanzaBuff_ROIDS,
			}
			if twoHand {
				consumes.OffHandImbue = proto.WeaponImbue_WeaponImbueUnknown
				// Arms takes Aggression, +31.8 over the Flask of the Titans to Precision's +10.0
				// (ui/warrior/presets.ts ArmsConsumes).
				consumes.Flask = proto.Flask_FlaskOfNaturalAggression
			}
			return consumes
		}
	}
	withWarrior := func(stance proto.WarriorStance) func(player *proto.Player) {
		return func(player *proto.Player) {
			player.Class = proto.Class_ClassWarrior
			player.Spec = &proto.Player_Warrior{Warrior: &proto.Warrior{
				Options: &proto.Warrior_Options{
					QueueDelay:   250,
					Shout:        proto.WarriorShout_WarriorShoutBattle,
					Stance:       stance,
					StartingRage: 0,
				},
			}}
		}
	}

	specs["fury"] = &specSetup{
		uiDir:    "ui/warrior",
		raid:     warriorRaid,
		consumes: warriorConsumes(false),
		withSpec: withWarrior(proto.WarriorStance_WarriorStanceBerserker),
	}
	specs["arms"] = &specSetup{
		uiDir:    "ui/warrior",
		raid:     warriorRaid,
		consumes: warriorConsumes(true),
		withSpec: withWarrior(proto.WarriorStance_WarriorStanceBattle),
	}
}
