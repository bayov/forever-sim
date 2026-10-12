package main

import "github.com/wowsims/classic/sim/core/proto"

// The three mage specs share the ui folder, the raid and most consumes. A mage stands in a
// caster group: Blessing of Wisdom in place of Might, a shaman's Mana Spring Totem, Moonkin
// Aura, and Curse of the Elements, which under Forever covers Arcane, Fire and Frost.
// Improved Scorch and Winter's Chill are the mage's own under Forever, so the raid gives none.
func mageSpec(firePower proto.FirePowerBuff, frostPower proto.FrostPowerBuff) *specSetup {
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
				FirePowerBuff:   firePower,
				FrostPowerBuff:  frostPower,
				Flask:           proto.Flask_FlaskOfSupremePower,
				Food:            proto.Food_FoodRunnTumTuberSurprise,
				MainHandImbue:   proto.WeaponImbue_BrilliantWizardOil,
				ManaRegenElixir: proto.ManaRegenElixir_MagebloodPotion,
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
