package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["feral"] = &specSetup{
		uiDir: "ui/feral_druid",
		// The cat is in a melee group with a shaman, who drops Strength of Earth and Windfury Totem,
		// as for the warrior (Windfury works on a cat under Forever), and the raid's second shaman
		// adds Grace of Air (the user's ruling 2026-10-16, "both totems": client 1.60.1.70425 lets Grace of Air stack with Windfury Totem and makes both raid-wide). Enhancing Totems is not in Forever's tree, so the
		// totems are the regular ones.
		raid: func(raidBuffs *proto.RaidBuffs, _ *proto.IndividualBuffs, _ *proto.Debuffs) {
			raidBuffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectRegular
			raidBuffs.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffWindfury
			raidBuffs.GraceOfAirTotem = proto.TristateEffect_TristateEffectRegular
			// The crit aura is the cat's own (Leader of the Pack), and no moonkin is in its group.
			raidBuffs.LeaderOfThePack = false
			raidBuffs.MoonkinAura = false
		},
		// The consumes the Level 60 preset uses (ui/feral_druid/presets.ts says what each is
		// worth). Under Forever the cat drinks in form and its weapon takes a stone; it still
		// can use no explosives.
		consumes: func() *proto.Consumes {
			return &proto.Consumes{
				AgilityElixir:     proto.AgilityElixir_ElixirOfTheMongoose,
				AttackPowerBuff:   proto.AttackPowerBuff_JujuMight,
				DefaultConjured:   proto.Conjured_ConjuredRogueThistleTea,
				DefaultPotion:     proto.Potions_MightyRagePotion,
				DragonBreathChili: true,
				FillerExplosive:   proto.Explosive_ExplosiveEZThroDarkBomb,
				Flask:             proto.Flask_FlaskOfNaturalAggression,
				Food:              proto.Food_FoodFlankAuPoivre,
				MainHandImbue:     proto.WeaponImbue_DenseWeightstone,
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
