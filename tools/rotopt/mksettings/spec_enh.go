package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["enh"] = &specSetup{
		uiDir: "ui/enhancement_shaman",
		consumes: func() *proto.Consumes {
			// The Level 60 preset's consumes (ui/enhancement_shaman/presets.ts). Under Forever the
			// shaman imbue is a class option, so the main hand holds Shadow Oil beside it.
			return &proto.Consumes{
				AgilityElixir:     proto.AgilityElixir_ElixirOfTheMongoose,
				AttackPowerBuff:   proto.AttackPowerBuff_JujuMight,
				DefaultPotion:     proto.Potions_MajorManaPotion,
				DefaultConjured:   proto.Conjured_ConjuredDemonicRune,
				DragonBreathChili: true,
				FillerExplosive:   proto.Explosive_ExplosiveGoblinLandMine,
				FirePowerBuff:     proto.FirePowerBuff_ElixirOfFirepower,
				Flask:             proto.Flask_FlaskOfSupremePower,
				Food:              proto.Food_FoodGrilledSquid,
				IntellectElixir:   proto.IntellectElixir_ElixirOfTheOwl,
				MainHandImbue:     proto.WeaponImbue_ShadowOil,
				ManaRegenElixir:   proto.ManaRegenElixir_GreaterMagebloodElixir,
				NaturePowerBuff:   proto.NaturePowerBuff_ElixirOfNaturePower,
				SapperExplosive:   proto.SapperExplosive_SapperGoblinSapper,
				ShadowPowerBuff:   proto.ShadowPowerBuff_ElixirOfShadowPower,
				SpellPowerBuff:    proto.SpellPowerBuff_GreaterArcaneElixir,
				SpiritElixir:      proto.SpiritElixir_ElixirOfSages,
				StrengthBuff:      proto.StrengthBuff_ElixirOfTheGrizzly,
				ZanzaBuff:         proto.ZanzaBuff_ROIDS,
			}
		},
		// The Level 60 preset's professions.
		professions: [2]proto.Profession{proto.Profession_Engineering, proto.Profession_Leatherworking},
		lowLevel: func(string) (lowLevelSetup, error) {
			return lowLevelSetup{
				consumes: &proto.Consumes{MainHandImbue: proto.WeaponImbue_RockbiterWeapon},
				tankSelf: true,
			}, nil
		},
		withSpec: func(player *proto.Player) {
			player.Class = proto.Class_ClassShaman
			player.Spec = &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
				Options: &proto.EnhancementShaman_Options{
					SyncType:    proto.ShamanSyncType_Auto,
					ShamanImbue: proto.WeaponImbue_WindfuryWeapon,
					// The Level 60 preset's starting totems. The level 30 rotations put theirs down
					// in the prepull actions, which take the place of these.
					StartingTotems: &proto.StartingTotems{
						Earth: proto.EarthTotem_StrengthOfEarthTotem, EarthSecondsBeforePull: 28,
						Air: proto.AirTotem_WindfuryTotem, AirSecondsBeforePull: 27,
						Fire: proto.FireTotem_FlametongueTotem, FireSecondsBeforePull: 29,
						Water: proto.WaterTotem_ManaSpringTotem, WaterSecondsBeforePull: 30,
					},
				},
			}}
		},
	}
}
