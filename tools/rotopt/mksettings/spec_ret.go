package main

import (
	"fmt"

	"github.com/wowsims/classic/sim/core/proto"
)

func init() {
	specs["ret"] = &specSetup{
		uiDir: "ui/retribution_paladin",
		// The paladin is in a melee group with a shaman. Enhancing Totems is not in Forever's tree,
		// so the totems are the regular ones.
		raid: func(raidBuffs *proto.RaidBuffs, _ *proto.IndividualBuffs, _ *proto.Debuffs) {
			raidBuffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectRegular
			raidBuffs.GraceOfAirTotem = proto.TristateEffect_TristateEffectRegular
		},
		consumes: func() *proto.Consumes {
			return &proto.Consumes{
				AgilityElixir:     proto.AgilityElixir_ElixirOfTheMongoose,
				AttackPowerBuff:   proto.AttackPowerBuff_JujuMight,
				DefaultPotion:     proto.Potions_MajorManaPotion,
				DefaultConjured:   proto.Conjured_ConjuredDemonicRune,
				DragonBreathChili: true,
				Flask:             proto.Flask_FlaskOfSupremePower,
				Food:              proto.Food_FoodBlessSunfruit,
				SapperExplosive:   proto.SapperExplosive_SapperGoblinSapper,
				SpellPowerBuff:    proto.SpellPowerBuff_GreaterArcaneElixir,
				StrengthBuff:      proto.StrengthBuff_JujuPower,
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
		withSpec: func(player *proto.Player) {
			player.Class = proto.Class_ClassPaladin
			player.Spec = &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
				Options: &proto.PaladinOptions{PrimarySeal: proto.PaladinSeal_Righteousness, Aura: proto.PaladinAura_RetributionAura},
			}}
		},
	}
}
