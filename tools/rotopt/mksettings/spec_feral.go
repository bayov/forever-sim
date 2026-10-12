package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["feral"] = &specSetup{
		uiDir: "ui/feral_druid",
		// The cat is in a melee group with a shaman, like the rogue. Enhancing Totems is not in
		// Forever's tree, so the totems are the regular ones.
		raid: func(raidBuffs *proto.RaidBuffs, _ *proto.IndividualBuffs, _ *proto.Debuffs) {
			raidBuffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectRegular
			raidBuffs.GraceOfAirTotem = proto.TristateEffect_TristateEffectRegular
			// The crit aura is the cat's own (Leader of the Pack), and no moonkin is in its group.
			raidBuffs.LeaderOfThePack = false
			raidBuffs.MoonkinAura = false
		},
		// The consumes the Level 60 preset uses (ui/feral_druid/presets.ts says what each is
		// worth). Shapeshifted, the cat can use no explosives and no weapon stone or oil.
		consumes: func() *proto.Consumes {
			return &proto.Consumes{
				AgilityElixir:     proto.AgilityElixir_ElixirOfTheMongoose,
				AttackPowerBuff:   proto.AttackPowerBuff_JujuMight,
				DefaultConjured:   proto.Conjured_ConjuredDemonicRune,
				DefaultPotion:     proto.Potions_MajorManaPotion,
				DragonBreathChili: true,
				Flask:             proto.Flask_FlaskOfDistilledWisdom,
				Food:              proto.Food_FoodFlankAuPoivre,
				StrengthBuff:      proto.StrengthBuff_ElixirOfFerocity,
				ZanzaBuff:         proto.ZanzaBuff_GroundScorpokAssay,
			}
		},
		withSpec: func(player *proto.Player) {
			player.Class = proto.Class_ClassDruid
			player.Spec = &proto.Player_FeralDruid{FeralDruid: &proto.FeralDruid{
				Options: &proto.FeralDruid_Options{
					InnervateTarget: &proto.UnitReference{},
					LatencyMs:       100,
				},
			}}
		},
	}
}
