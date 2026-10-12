package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["balance"] = &specSetup{
		uiDir: "ui/balance_druid",
		// The moonkin is in a caster group with a shaman, which puts Mana Spring down for it.
		// It takes Blessing of Wisdom and Kings, not Might. Of the shared debuffs only Curse of the
		// Elements reaches it (every school at 10% under Forever), the rest do nothing to it.
		raid: func(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, _ *proto.Debuffs) {
			raidBuffs.ManaSpringTotem = proto.TristateEffect_TristateEffectRegular
			// The crit aura is the moonkin's own (Moonkin Form), and no cat is in its group.
			raidBuffs.MoonkinAura = false
			raidBuffs.LeaderOfThePack = false
			buffs.BlessingOfMight = proto.TristateEffect_TristateEffectMissing
		},
		// The consumes the Level 60 preset uses (ui/balance_druid/presets.ts says what each is
		// worth).
		consumes: func() *proto.Consumes {
			return &proto.Consumes{
				DefaultConjured: proto.Conjured_ConjuredDemonicRune,
				DefaultPotion:   proto.Potions_MajorSpellblastingPotion,
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
			player.Class = proto.Class_ClassDruid
			// A caster stands at range, so the moonkin does not melee between casts.
			player.DistanceFromTarget = 30
			player.Spec = &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{
				Options: &proto.BalanceDruid_Options{InnervateTarget: &proto.UnitReference{Type: proto.UnitReference_Self}},
			}}
		},
	}
}
