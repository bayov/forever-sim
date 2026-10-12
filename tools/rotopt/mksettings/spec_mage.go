package main

import "github.com/wowsims/classic/sim/core/proto"

// The three mage specs share the ui folder, the raid and most consumes. A mage stands in a
// caster group: Blessing of Wisdom in place of Might, a shaman's Mana Spring Totem, Moonkin
// Aura, and Curse of the Elements, which under Forever covers Arcane, Fire and Frost.
// Improved Scorch and Winter's Chill are the mage's own under Forever, so the raid gives none.
//
// The consumes are the same for every spec. On 2026-10-12 we tried every consumable the sim applies
// in each slot on the four searched builds (180 sec, 20000 iterations). The presets model
// Forever's launch, with no raid open, so nothing that only a raid sells: Wizard Oil in place of
// Brilliant Wizard Oil (Zandalar Tribe). Cerebral Cortex Compound is a Blasted Lands quest.
// One +2% crit elixir a character: Elixir of Sages with Elixir of Greater Intellect is 0.3 to 1
// ahead of the Owl with Greater Spirit (the Owl on top would be 10 to 21 more). Greater Mageblood
// Elixir is 0 to 6 and Lobster
// Stew 3 to 4 over Nightfin Soup. Major Spellblasting Potion is +7 for Frost and -15 for
// Frostfire, so the Major Mana Potion stays. Elixir of Greater Firepower is the Elixir of Holy
// Power under Forever (+40 Holy), so the Fire elixir is the plain Elixir of Firepower. Flask of
// Distilled Wisdom is 75 to 90 behind Supreme Power. A Goblin Land Mine costs 2 to 11 (its throw
// takes a GCD), and Mage Armor is 8 to 17 ahead of Ice Armor.
func mageSpec() *specSetup {
	return &specSetup{
		uiDir: "ui/mage",
		raid: func(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, debuffs *proto.Debuffs) {
			buffs.BlessingOfMight = proto.TristateEffect_TristateEffectMissing
			raidBuffs.ManaSpringTotem = proto.TristateEffect_TristateEffectRegular
		},
		consumes: func() *proto.Consumes {
			return &proto.Consumes{
				DefaultPotion:   proto.Potions_MajorManaPotion,
				DefaultConjured: proto.Conjured_ConjuredDemonicRune,
				FirePowerBuff:   proto.FirePowerBuff_ElixirOfFirepower,
				FrostPowerBuff:  proto.FrostPowerBuff_ElixirOfFrostPower,
				Flask:           proto.Flask_FlaskOfSupremePower,
				Food:            proto.Food_FoodLobsterStew,
				IntellectElixir: proto.IntellectElixir_ElixirOfGreaterIntellect,
				MainHandImbue:   proto.WeaponImbue_WizardOil,
				ManaRegenElixir: proto.ManaRegenElixir_GreaterMagebloodElixir,
				SpiritElixir:    proto.SpiritElixir_ElixirOfSages,
				SapperExplosive: proto.SapperExplosive_SapperGoblinSapper,
				SpellPowerBuff:  proto.SpellPowerBuff_GreaterArcaneElixir,
				ZanzaBuff:       proto.ZanzaBuff_CerebralCortexCompound,
			}
		},
		professions: [2]proto.Profession{proto.Profession_Engineering, proto.Profession_Tailoring},
		withSpec: func(player *proto.Player) {
			player.Class = proto.Class_ClassMage
			player.Spec = &proto.Player_Mage{Mage: &proto.Mage{
				Options: &proto.Mage_Options{Armor: proto.Mage_Options_MageArmor},
			}}
			// Casters stand at range, which only delays the bolts a little.
			player.DistanceFromTarget = 25
		},
	}
}
