package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["ele"] = &specSetup{
		uiDir: "ui/elemental_shaman",
		// The elemental shaman is in a caster group and is its only shaman, so its totems are the
		// ones the rotation puts down. It takes Wisdom and Kings from the paladins. Might is no use
		// to it. The caster debuffs that touch its schools stay on: Curse of the Elements and
		// Improved Scorch for Fire (Flame Shock, Lava Burst, the fire totems) and Winter's Chill for
		// Frost Shock. Nature has no debuff.
		raid: func(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, debuffs *proto.Debuffs) {
			buffs.BlessingOfMight = proto.TristateEffect_TristateEffectMissing
			debuffs.ImprovedScorch = true
			debuffs.WintersChill = true
		},
		consumes: func() *proto.Consumes {
			return &proto.Consumes{
				DefaultPotion:   proto.Potions_MajorManaPotion,
				DefaultConjured: proto.Conjured_ConjuredDemonicRune,
				FirePowerBuff:   proto.FirePowerBuff_ElixirOfFirepower,
				Flask:           proto.Flask_FlaskOfSupremePower,
				Food:            proto.Food_FoodNightfinSoup,
				MainHandImbue:   proto.WeaponImbue_WizardOil,
				ManaRegenElixir: proto.ManaRegenElixir_MagebloodPotion,
				SpellPowerBuff:  proto.SpellPowerBuff_GreaterArcaneElixir,
				ZanzaBuff:       proto.ZanzaBuff_CerebralCortexCompound,
			}
		},
		withSpec: func(player *proto.Player) {
			player.Class = proto.Class_ClassShaman
			// A caster stands at range, so it never swings.
			player.DistanceFromTarget = 20
			player.Spec = &proto.Player_ElementalShaman{ElementalShaman: &proto.ElementalShaman{
				Options: &proto.ElementalShaman_Options{},
			}}
		},
	}
}
