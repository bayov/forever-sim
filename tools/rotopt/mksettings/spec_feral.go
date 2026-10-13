package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["feral"] = &specSetup{
		uiDir: "ui/feral_druid",
		// The cat is in a melee group with a shaman, who drops Strength of Earth and Windfury Totem,
		// as for the warrior (Forever's air totems don't stack in a party, and Windfury works on a
		// cat under Forever). Enhancing Totems is not in Forever's tree, so the totems are the
		// regular ones.
		raid: func(raidBuffs *proto.RaidBuffs, _ *proto.IndividualBuffs, _ *proto.Debuffs) {
			raidBuffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectRegular
			raidBuffs.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffWindfury
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
				Flask:             proto.Flask_FlaskOfDistilledWisdom,
				Food:              proto.Food_FoodFlankAuPoivre,
				MainHandImbue:     proto.WeaponImbue_ElementalSharpeningStone,
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
