package main

import (
	"flag"

	"github.com/wowsims/classic/sim/core/proto"
)

func init() {
	specs["enh"] = &specSetup{
		uiDir: "ui/enhancement_shaman",
		consumes: func() *proto.Consumes {
			// The Level 60 preset's consumes (ui/enhancement_shaman/presets.ts). Under Forever the
			// shaman imbue is a class option, so the main hand holds Shadow Oil beside it.
			return &proto.Consumes{
				AgilityElixir:     proto.AgilityElixir_ElixirOfCunning,
				AttackPowerBuff:   proto.AttackPowerBuff_JujuMight,
				DefaultPotion:     proto.Potions_MajorManaPotion,
				DefaultConjured:   proto.Conjured_ConjuredDemonicRune,
				DragonBreathChili: true,
				FillerExplosive:   proto.Explosive_ExplosiveGoblinLandMine,
				FirePowerBuff:     proto.FirePowerBuff_ElixirOfFirepower,
				Flask:             proto.Flask_FlaskOfSupremePower,
				Food:              proto.Food_FoodGrilledSquid,
				IntellectElixir:   proto.IntellectElixir_ElixirOfGreaterIntellect,
				MainHandImbue:     proto.WeaponImbue_ShadowOil,
				ManaRegenElixir:   proto.ManaRegenElixir_GreaterMagebloodElixir,
				NaturePowerBuff:   proto.NaturePowerBuff_ElixirOfNaturePower,
				SapperExplosive:   proto.SapperExplosive_SapperGoblinSapper,
				ShadowPowerBuff:   proto.ShadowPowerBuff_ElixirOfShadowPower,
				SpellPowerBuff:    proto.SpellPowerBuff_GreaterArcaneElixir,
				SpiritElixir:      proto.SpiritElixir_ElixirOfSages,
				StrengthBuff:      proto.StrengthBuff_ElixirOfFerocity,
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
			// The shaman drops its melee party's totems itself (standards 23: no totem duty, every
			// shaman puts its own down, like a warrior's Sunder): Strength of Earth, Windfury Totem
			// (it can't stand with Grace of Air) and Mana Spring. The raid has none of its own.
			startingTotems := &proto.StartingTotems{
				Earth: proto.EarthTotem_StrengthOfEarthTotem, EarthSecondsBeforePull: 4,
				Air: proto.AirTotem_WindfuryTotem, AirSecondsBeforePull: 3,
				Water: proto.WaterTotem_ManaSpringTotem, WaterSecondsBeforePull: 2,
			}
			// The raid damage that hits a melee shaman in a level 60 raid fight and fires its
			// Lightning Shield (ui/enhancement_shaman/presets.ts RaidDamageHitsLevel60). The level
			// 30 builds' enemy hits the shaman itself.
			raidHits := 2.0
			if flag.Lookup("level").Value.String() != "60" {
				// The level 30 builds keep the old seconds. At level 60 they go down just before
				// the pull, so a 5 min totem lasts a 300 sec fight.
				startingTotems.EarthSecondsBeforePull, startingTotems.AirSecondsBeforePull, startingTotems.WaterSecondsBeforePull = 28, 27, 30
				raidHits = 0
			}
			player.Spec = &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
				Options: &proto.EnhancementShaman_Options{
					SyncType:    proto.ShamanSyncType_Auto,
					ShamanImbue: proto.WeaponImbue_WindfuryWeapon,
					// The Level 60 preset's starting totems. The level 30 rotations put theirs down
					// in the prepull actions, which take the place of these.
					StartingTotems:          startingTotems,
					RaidDamageHitsPerMinute: raidHits,
				},
			}}
		},
	}
}
