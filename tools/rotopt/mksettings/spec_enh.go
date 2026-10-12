package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["enh"] = &specSetup{
		uiDir: "ui/enhancement_shaman",
		consumes: func() *proto.Consumes {
			return &proto.Consumes{
				AgilityElixir:     proto.AgilityElixir_ElixirOfTheMongoose,
				AttackPowerBuff:   proto.AttackPowerBuff_JujuMight,
				DefaultPotion:     proto.Potions_MajorManaPotion,
				DefaultConjured:   proto.Conjured_ConjuredDemonicRune,
				DragonBreathChili: true,
				FirePowerBuff:     proto.FirePowerBuff_ElixirOfFirepower,
				Flask:             proto.Flask_FlaskOfSupremePower,
				Food:              proto.Food_FoodBlessSunfruit,
				MainHandImbue:     proto.WeaponImbue_WindfuryWeapon,
				ManaRegenElixir:   proto.ManaRegenElixir_MagebloodPotion,
				OffHandImbue:      proto.WeaponImbue_WindfuryWeapon,
				SapperExplosive:   proto.SapperExplosive_SapperGoblinSapper,
				SpellPowerBuff:    proto.SpellPowerBuff_GreaterArcaneElixir,
				StrengthBuff:      proto.StrengthBuff_JujuPower,
				ZanzaBuff:         proto.ZanzaBuff_ROIDS,
			}
		},
		lowLevel: func(string) (lowLevelSetup, error) {
			return lowLevelSetup{
				consumes: &proto.Consumes{MainHandImbue: proto.WeaponImbue_RockbiterWeapon},
				tankSelf: true,
			}, nil
		},
		withSpec: func(player *proto.Player) {
			player.Class = proto.Class_ClassShaman
			player.Spec = &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
				Options: &proto.EnhancementShaman_Options{SyncType: proto.ShamanSyncType_Auto},
			}}
		},
	}
}
