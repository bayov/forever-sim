package main

import (
	"fmt"

	"github.com/wowsims/classic/sim/core/proto"
)

func init() {
	specs["ret"] = &specSetup{
		uiDir: "ui/retribution_paladin",
		// The paladin is in a melee group with a shaman. Enhancing Totems is not in Forever's tree,
		// so the totems are the regular ones. A shaman puts down one air totem for the party
		// since the 2026-09-24 beta build, and for a melee group that is Windfury Totem, and
		// Mana Spring Totem is its water totem.
		raid: func(raidBuffs *proto.RaidBuffs, _ *proto.IndividualBuffs, _ *proto.Debuffs) {
			raidBuffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectRegular
			raidBuffs.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffWindfury
			raidBuffs.ManaSpringTotem = proto.TristateEffect_TristateEffectRegular
		},
		// The consumes the level 60 search kept, each tried slot by slot on the synthetic Phase 1
		// set at 180 and 300 sec (see the Level 60 consumes in ui/retribution_paladin/presets.ts).
		consumes: func() *proto.Consumes {
			return &proto.Consumes{
				AgilityElixir:     proto.AgilityElixir_ElixirOfGreaterAgility,
				AttackPowerBuff:   proto.AttackPowerBuff_JujuMight,
				DefaultPotion:     proto.Potions_MajorManaPotion,
				DefaultConjured:   proto.Conjured_ConjuredDemonicRune,
				DragonBreathChili: true,
				FillerExplosive:   proto.Explosive_ExplosiveGoblinLandMine,
				FirePowerBuff:     proto.FirePowerBuff_ElixirOfGreaterFirepower,
				Flask:             proto.Flask_FlaskOfSupremePower,
				Food:              proto.Food_FoodLobsterStew,
				IntellectElixir:   proto.IntellectElixir_ElixirOfTheOwl,
				MainHandImbue:     proto.WeaponImbue_ShadowOil,
				ManaRegenElixir:   proto.ManaRegenElixir_GreaterMagebloodElixir,
				SapperExplosive:   proto.SapperExplosive_SapperGoblinSapper,
				ShadowPowerBuff:   proto.ShadowPowerBuff_ElixirOfShadowPower,
				SpellPowerBuff:    proto.SpellPowerBuff_GreaterArcaneElixir,
				SpiritElixir:      proto.SpiritElixir_ElixirOfGreaterSpirit,
				StrengthBuff:      proto.StrengthBuff_ElixirOfFerocity,
				ZanzaBuff:         proto.ZanzaBuff_ROIDS,
			}
		},
		// The paladin's own blessing, one at a time, where everyone else assumes a paladin in
		// the group and takes Kings. Kings won the level 20 search (see the wiki), the flag is
		// there to check that again.
		lowLevel: func(blessing string) (lowLevelSetup, error) {
			low := lowLevelSetup{consumes: &proto.Consumes{}, tankSelf: true}
			switch blessing {
			case "kings":
				low.buffs = &proto.IndividualBuffs{BlessingOfKings: true}
			case "might":
				low.buffs = &proto.IndividualBuffs{BlessingOfMight: proto.TristateEffect_TristateEffectRegular}
			case "wisdom":
				low.buffs = &proto.IndividualBuffs{BlessingOfWisdom: proto.TristateEffect_TristateEffectRegular}
			default:
				return low, fmt.Errorf("unknown blessing %q", blessing)
			}
			return low, nil
		},
		// Engineering for the Goblin Sapper and the Land Mine, and Blacksmithing like the level 30
		// build.
		professions: [2]proto.Profession{proto.Profession_Engineering, proto.Profession_Blacksmithing},
		withSpec: func(player *proto.Player) {
			player.Class = proto.Class_ClassPaladin
			player.Spec = &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
				Options: &proto.PaladinOptions{PrimarySeal: proto.PaladinSeal_Righteousness, Aura: proto.PaladinAura_RetributionAura},
			}}
		},
	}
}
