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
		},
		// Each consume was tried in and out of the set on the Level 60 build (180 sec, 20000
		// iterations, 2026-10-12): Lobster Stew (28 spell damage under Forever) is 2.8 ahead of
		// Nightfin Soup, the Goblin Sapper Charge added 3.7 at 5 yd (a filler explosive costs a GCD
		// and loses 4), the Elixir of the Owl 16.4 over the Elixir of Wisdom (one 2% crit elixir a
		// character, Sages would add 13.9), the Major Spellblasting Potion 8.2 over the Major
		// Mana Potion and Cerebral Cortex Compound 3.2. Brilliant Wizard Oil would be 12.5 ahead of Wizard Oil but is a
		// Zul'Gurub recipe, Shadow Oil and Dragonbreath Chili add nothing. Mana never runs out
		// in a raid fight (about 25 min to empty), so Mageblood is left out.
		consumes: func() *proto.Consumes {
			return &proto.Consumes{
				DefaultPotion:   proto.Potions_MajorSpellblastingPotion,
				DefaultConjured: proto.Conjured_ConjuredDemonicRune,
				Flask:           proto.Flask_FlaskOfSupremePower,
				Food:            proto.Food_FoodLobsterStew,
				IntellectElixir: proto.IntellectElixir_ElixirOfTheOwl,
				MainHandImbue:   proto.WeaponImbue_WizardOil,
				ShadowPowerBuff: proto.ShadowPowerBuff_ElixirOfShadowPower,
				SpellPowerBuff:  proto.SpellPowerBuff_GreaterArcaneElixir,
				ZanzaBuff:       proto.ZanzaBuff_CerebralCortexCompound,
			}
		},
		// Engineering was for the Goblin Sapper Charge, Tailoring is the priest's crafting.
		professions: [2]proto.Profession{proto.Profession_Engineering, proto.Profession_Tailoring},
		withSpec: func(player *proto.Player) {
			player.Class = proto.Class_ClassPriest
			player.Spec = &proto.Player_ShadowPriest{ShadowPriest: &proto.ShadowPriest{
				Options: &proto.ShadowPriest_Options{Armor: proto.ShadowPriest_Options_InnerFire},
			}}
			// A caster stands at the longest range every spell of its rotation reaches (standards
			// 33, orchestration/caster_distance.md): Mind Flay's 20 yd, 22 with Shadow Reach 1/2.
			// The Goblin Sapper Charge and the Goblin Land Mine only reach 10 yd, so both are out
			// (they were 4.3 to 6.2 at 180 sec at 5 yd). The EZ-Thro Dark Bomb lost 5.6 to 6.7 on
			// Shadow and tied on Smite (+0.4) at 25 yd, so no explosive.
			player.DistanceFromTarget = 22
		},
	}
	// Shadow + Holy adds Greater Mageblood (ui/shadow_priest/presets.ts ConsumesHoly): the Holy
	// sub tree build runs short of mana.
	specs["shadow_holy"] = withConsumes(specs["shadow"], func(consumes *proto.Consumes) {
		consumes.ManaRegenElixir = proto.ManaRegenElixir_GreaterMagebloodElixir
	})
	// Smite's own consumes (ui/shadow_priest/presets.ts ConsumesSmite): the Major Mana Potion,
	// Greater Mageblood and Greater Spirit for its mana, Greater Firepower for Smite and Holy
	// Fire. No Goblin Land Mine since the priests stand at range (it was +0.4 / +2.7 at 180 / 300
	// sec at 5 yd). Smite has no Mind Flay, and every spell it casts reaches 30 yd.
	specs["smite"] = withDistance(withConsumes(specs["shadow"], func(consumes *proto.Consumes) {
		consumes.DefaultPotion = proto.Potions_MajorManaPotion
		consumes.FirePowerBuff = proto.FirePowerBuff_ElixirOfGreaterFirepower
		consumes.ManaRegenElixir = proto.ManaRegenElixir_GreaterMagebloodElixir
		consumes.SpiritElixir = proto.SpiritElixir_ElixirOfGreaterSpirit
	}), 30)
}
