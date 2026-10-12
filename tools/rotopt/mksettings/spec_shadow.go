package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["shadow"] = &specSetup{
		uiDir: "ui/shadow_priest",
		// A caster takes Blessing of Wisdom where the melee take Might, and the shaman in a
		// caster group drops Mana Spring (Dwarf shamans give the Alliance totems too). The
		// melee debuffs in the shared raid do nothing for a priest. Forever has no Curse of
		// Shadow, its Curse of the Elements covers Shadow, and Shadow Weaving as a raid debuff
		// is off under Forever, the priest's own talent gives it.
		raid: func(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, debuffs *proto.Debuffs) {
			buffs.BlessingOfMight = proto.TristateEffect_TristateEffectMissing
			raidBuffs.ManaSpringTotem = proto.TristateEffect_TristateEffectRegular
			debuffs.CurseOfShadow = false
		},
		// Each consume was tried in and out of the set on the Level 60 build (180 sec, 20000
		// iterations, 2026-10-12): Nightfin Soup (22 spell damage under Forever) is 7.9 ahead
		// of Runn Tum Tuber Surprise and 3.1 ahead of Grilled Squid, the Goblin Sapper Charge
		// adds 3.7 (a filler explosive costs a GCD and loses 4), Elixir of Wisdom 0.9.
		// Brilliant Wizard Oil would be 12.5 ahead of Wizard Oil but is a Zul'Gurub recipe,
		// Shadow Oil and Dragonbreath Chili add nothing, Cerebral Cortex Compound +3.2. Mana never runs out in a raid fight
		// (about 25 min to empty), so Mageblood is left out, the potion and rune are kept for
		// longer fights.
		consumes: func() *proto.Consumes {
			return &proto.Consumes{
				DefaultPotion:   proto.Potions_MajorManaPotion,
				DefaultConjured: proto.Conjured_ConjuredDemonicRune,
				Flask:           proto.Flask_FlaskOfSupremePower,
				Food:            proto.Food_FoodNightfinSoup,
				IntellectElixir: proto.IntellectElixir_ElixirOfWisdom,
				MainHandImbue:   proto.WeaponImbue_WizardOil,
				SapperExplosive: proto.SapperExplosive_SapperGoblinSapper,
				ShadowPowerBuff: proto.ShadowPowerBuff_ElixirOfShadowPower,
				SpellPowerBuff:  proto.SpellPowerBuff_GreaterArcaneElixir,
				ZanzaBuff:       proto.ZanzaBuff_CerebralCortexCompound,
			}
		},
		// Engineering for the Goblin Sapper Charge, Tailoring as the priest's crafting.
		professions: [2]proto.Profession{proto.Profession_Engineering, proto.Profession_Tailoring},
		withSpec: func(player *proto.Player) {
			player.Class = proto.Class_ClassPriest
			player.Spec = &proto.Player_ShadowPriest{ShadowPriest: &proto.ShadowPriest{
				Options: &proto.ShadowPriest_Options{Armor: proto.ShadowPriest_Options_InnerFire},
			}}
		},
	}
}
