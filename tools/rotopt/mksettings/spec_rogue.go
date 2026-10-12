package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["rogue"] = &specSetup{
		uiDir: "ui/rogue",
		// The rogue is in a melee group with a shaman. Enhancing Totems is not in Forever's tree,
		// so the totems are the regular ones.
		raid: func(raidBuffs *proto.RaidBuffs, _ *proto.IndividualBuffs, _ *proto.Debuffs) {
			raidBuffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectRegular
			raidBuffs.GraceOfAirTotem = proto.TristateEffect_TristateEffectRegular
		},
		consumes: func() *proto.Consumes {
			return &proto.Consumes{
				AgilityElixir:     proto.AgilityElixir_ElixirOfTheMongoose,
				AttackPowerBuff:   proto.AttackPowerBuff_JujuMight,
				DefaultConjured:   proto.Conjured_ConjuredRogueThistleTea,
				DragonBreathChili: true,
				Flask:             proto.Flask_FlaskOfSupremePower,
				Food:              proto.Food_FoodGrilledSquid,
				MainHandImbue:     proto.WeaponImbue_InstantPoison,
				OffHandImbue:      proto.WeaponImbue_DeadlyPoison,
				SpellPowerBuff:    proto.SpellPowerBuff_GreaterArcaneElixir,
				StrengthBuff:      proto.StrengthBuff_JujuPower,
				ZanzaBuff:         proto.ZanzaBuff_GroundScorpokAssay,
				SapperExplosive:   proto.SapperExplosive_SapperGoblinSapper,
			}
		},
		lowLevel: func(string) (lowLevelSetup, error) {
			return lowLevelSetup{consumes: &proto.Consumes{
				DefaultConjured: proto.Conjured_ConjuredRogueThistleTea,
				MainHandImbue:   proto.WeaponImbue_InstantPoison,
				OffHandImbue:    proto.WeaponImbue_InstantPoison,
			}}, nil
		},
		withSpec: func(player *proto.Player) {
			player.Class = proto.Class_ClassRogue
			player.Spec = &proto.Player_Rogue{Rogue: &proto.Rogue{Options: &proto.RogueOptions{}}}
		},
	}
}
