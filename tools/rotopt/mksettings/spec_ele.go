package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["ele"] = &specSetup{
		uiDir: "ui/elemental_shaman",
		// The elemental shaman is in a caster group and is its only shaman, so its totems are the
		// ones the rotation puts down. It takes Wisdom and Kings from the paladins. Might is no use
		// to it. Curse of the Elements is the one raid debuff on its schools: under Forever Improved
		// Scorch and Winter's Chill only help the mage's own spells, and Curse of Shadow is gone.
		raid: func(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, debuffs *proto.Debuffs) {
			buffs.BlessingOfMight = proto.TristateEffect_TristateEffectMissing
			debuffs.CurseOfShadow = false
		},
		// The Level 60 preset's consumes (ui/elemental_shaman/presets.ts has what each is worth).
		consumes: func() *proto.Consumes {
			return &proto.Consumes{
				DefaultPotion:   proto.Potions_MajorManaPotion,
				DefaultConjured: proto.Conjured_ConjuredDemonicRune,
				FillerExplosive: proto.Explosive_ExplosiveGoblinLandMine,
				FirePowerBuff:   proto.FirePowerBuff_ElixirOfFirepower,
				Flask:           proto.Flask_FlaskOfSupremePower,
				Food:            proto.Food_FoodLobsterStew,
				IntellectElixir: proto.IntellectElixir_ElixirOfTheOwl,
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
}
