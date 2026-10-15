package main

import (
	"flag"

	"github.com/wowsims/classic/sim/core/proto"
)

// shamanTotems says who puts down the shaman's party totems at level 60 (standards 23). "raid"
// (the base presets): the raid's totem buffs give them, as if another shaman in the party
// dropped them, and the shaman casts only the totems it puts down for its own damage. "self"
// (the totem duty presets): the shaman drops them itself, and the raid has none of them.
var shamanTotems = flag.String("totems", "raid", "who puts down a level 60 shaman's party totems: raid (base preset) or self (totem duty preset)")

// shamanTotemsFromRaid is true at level 60 when the raid gives the party totems.
func shamanTotemsFromRaid() bool {
	return *shamanTotems != "self" && flag.Lookup("level").Value.String() == "60"
}

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
		// A melee party gets Strength of Earth, Windfury Totem (which can't stand with Grace of Air)
		// and Mana Spring. Our own Windfury Weapon turns Windfury Totem's buff off for us, so it
		// is the party's. None of these is improved: Forever took Enhancing Totems out, and the
		// builds have no Restorative Totems.
		raid: func(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, debuffs *proto.Debuffs) {
			if shamanTotemsFromRaid() {
				raidBuffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectRegular
				raidBuffs.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffWindfury
				raidBuffs.ManaSpringTotem = proto.TristateEffect_TristateEffectRegular
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
			startingTotems := &proto.StartingTotems{
				Earth: proto.EarthTotem_StrengthOfEarthTotem, EarthSecondsBeforePull: 28,
				Air: proto.AirTotem_WindfuryTotem, AirSecondsBeforePull: 27,
				Water: proto.WaterTotem_ManaSpringTotem, WaterSecondsBeforePull: 30,
			}
			if shamanTotemsFromRaid() {
				// The raid's totems take their place.
				startingTotems = &proto.StartingTotems{}
			}
			player.Spec = &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
				Options: &proto.EnhancementShaman_Options{
					SyncType:    proto.ShamanSyncType_Auto,
					ShamanImbue: proto.WeaponImbue_WindfuryWeapon,
					// The Level 60 preset's starting totems. The level 30 rotations put theirs down
					// in the prepull actions, which take the place of these.
					StartingTotems: startingTotems,
				},
			}}
		},
	}
}
