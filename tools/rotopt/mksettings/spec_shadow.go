package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["shadow"] = &specSetup{
		uiDir: "ui/shadow_priest",
		// A caster takes Blessing of Wisdom where the melee take Might, and the shaman in a
		// caster group drops Mana Spring (Dwarf shamans give the Alliance totems too). The
		// melee debuffs in the shared raid do nothing for a priest, Curse of Shadow is the
		// one that counts (Curse of the Elements helps Starshards). Shadow Weaving as a raid
		// debuff is off under Forever, the priest's own talent gives it.
		raid: func(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, debuffs *proto.Debuffs) {
			buffs.BlessingOfMight = proto.TristateEffect_TristateEffectMissing
			raidBuffs.ManaSpringTotem = proto.TristateEffect_TristateEffectRegular
		},
		consumes: func() *proto.Consumes {
			return &proto.Consumes{
				DefaultPotion:   proto.Potions_MajorManaPotion,
				DefaultConjured: proto.Conjured_ConjuredDemonicRune,
				Flask:           proto.Flask_FlaskOfSupremePower,
				Food:            proto.Food_FoodRunnTumTuberSurprise,
				MainHandImbue:   proto.WeaponImbue_WizardOil,
				ManaRegenElixir: proto.ManaRegenElixir_MagebloodPotion,
				ShadowPowerBuff: proto.ShadowPowerBuff_ElixirOfShadowPower,
				SpellPowerBuff:  proto.SpellPowerBuff_GreaterArcaneElixir,
			}
		},
		withSpec: func(player *proto.Player) {
			player.Class = proto.Class_ClassPriest
			player.Spec = &proto.Player_ShadowPriest{ShadowPriest: &proto.ShadowPriest{
				Options: &proto.ShadowPriest_Options{Armor: proto.ShadowPriest_Options_InnerFire},
			}}
		},
	}
}
