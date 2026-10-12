package main

import "github.com/wowsims/classic/sim/core/proto"

// The DPS warrior is in a melee group with a shaman, who drops Strength of Earth and Windfury Totem
// (Forever's air totems no longer stack in a party, so it is Windfury or Grace of Air, and Windfury is
// worth more to a warrior). The paladins bless it with Kings and Might. The shared raid already has Battle Shout from another warrior, Sunder Armor,
// Faerie Fire, Curse of Recklessness and Expose Armor.
func init() {
	warriorRaid := func(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, _ *proto.Debuffs) {
		raidBuffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectRegular
		raidBuffs.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffWindfury
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
				Food:              proto.Food_FoodSmokedDesertDumpling,
				MainHandImbue:     proto.WeaponImbue_DenseSharpeningStone,
				OffHandImbue:      proto.WeaponImbue_DenseSharpeningStone,
				SapperExplosive:   proto.SapperExplosive_SapperGoblinSapper,
				StrengthBuff:      proto.StrengthBuff_ElixirOfTheGrizzly,
				ZanzaBuff:         proto.ZanzaBuff_ROIDS,
			}
			if twoHand {
				consumes.OffHandImbue = proto.WeaponImbue_WeaponImbueUnknown
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
