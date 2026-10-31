package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["ele"] = &specSetup{
		uiDir: "ui/elemental_shaman",
		// The elemental shaman is in a caster group, whose only DPS totem is Mana Spring. The rotation
		// puts it down (standards 23: every shaman drops its own totems), with Searing Totem for its
		// own damage. It takes Wisdom and Kings from the paladins. Might is no use
		// to it. Curse of the Elements is the one raid debuff on its schools: under Forever Improved
		// Scorch and Winter's Chill only help the mage's own spells, and Curse of Shadow is gone.
		raid: func(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, debuffs *proto.Debuffs) {
			buffs.BlessingOfMight = proto.TristateEffect_TristateEffectMissing
		},
		// The Level 60 preset's consumes (ui/elemental_shaman/presets.ts has what each is worth).
		// No filler explosive: the Goblin Land Mine's cast loses (standards 11, 2026-10-23).
		consumes: func() *proto.Consumes {
			return &proto.Consumes{
				DefaultPotion:   proto.Potions_CausticSmogPotion,
				DefaultConjured: proto.Conjured_ConjuredDemonicRune,
				FirePowerBuff:   proto.FirePowerBuff_ElixirOfFirepower,
				Flask:           proto.Flask_FlaskOfSupremePower,
				Food:            proto.Food_FoodLobsterStew,
				IntellectElixir: proto.IntellectElixir_ElixirOfGreaterIntellect,
				MainHandImbue:   proto.WeaponImbue_WizardOil,
				ManaRegenElixir: proto.ManaRegenElixir_GreaterMagebloodElixir,
				NaturePowerBuff: proto.NaturePowerBuff_ElixirOfNaturePower,
				SapperExplosive: proto.SapperExplosive_SapperGoblinSapper,
				SpellPowerBuff:  proto.SpellPowerBuff_GreaterArcaneElixir,
				SpiritElixir:    proto.SpiritElixir_ElixirOfSages,
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
	// The Mana Tide build drinks the Major Mana Potion (ui/elemental_shaman/presets.ts
	// ManaTideConsumes): it runs shorter on mana, and the potion is -4.3 / +1.8 / +27.2 at 120 /
	// 180 / 300 sec over the Caustic Smog Potion.
	specs["ele_mana_tide"] = withConsumes(specs["ele"], func(consumes *proto.Consumes) {
		consumes.DefaultPotion = proto.Potions_MajorManaPotion
	})
}
