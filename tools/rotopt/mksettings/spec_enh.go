package main

import (
	"flag"

	"github.com/wowsims/classic/sim/core/proto"
)

func init() {
	specs["enh"] = &specSetup{
		uiDir: "ui/enhancement_shaman",
		// The shaman drops its own party's totems (withSpec), and the raid's second shaman, the one
		// whose Grace of Air every melee group gets, gives it Grace of Air too (the user's ruling 2026-10-16, "both totems": client 1.60.1.70425 lets Grace of Air stack with Windfury Totem and makes both raid-wide).
		raid: func(raidBuffs *proto.RaidBuffs, _ *proto.IndividualBuffs, _ *proto.Debuffs) {
			raidBuffs.GraceOfAirTotem = proto.TristateEffect_TristateEffectRegular
			// That second shaman's fire slot holds Flametongue Totem at level 60 (the user's ruling
			// 2026-10-16: "Let's assume it works, that it doesn't interfere with other melee getting
			// Windfury Totem, and that shamans with their WF imbue get FTT on top"). Our Windfury
			// Weapon turns our own Windfury Totem's buff off for us, so the raid's Flametongue Totem
			// holds the weapon's totem slot.
			if flag.Lookup("level").Value.String() == "60" {
				raidBuffs.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffFlametongue
			}
		},
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
				Flask:             proto.Flask_FlaskOfNaturalPrecision,
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
			// and Mana Spring, in the rotation's prepull actions. Its own Windfury Weapon turns
			// Windfury Totem's buff off for itself. The totem is for the party.
			player.Spec = &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
				Options: &proto.EnhancementShaman_Options{
					SyncType:    proto.ShamanSyncType_Auto,
					ShamanImbue: proto.WeaponImbue_WindfuryWeapon,
				},
			}}
		},
	}
}
