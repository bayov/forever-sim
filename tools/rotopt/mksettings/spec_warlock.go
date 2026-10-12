package main

import "github.com/wowsims/classic/sim/core/proto"

// The three warlock specs share the ui folder, the raid and the consumes. They differ in the
// demon that is out and the one Demonic Sacrifice took before the pull.
func init() {
	warlockSpec := func(summon, sacrificed proto.WarlockOptions_Summon, imbue proto.WarlockOptions_WeaponImbue) *specSetup {
		return &specSetup{
			uiDir: "ui/warlock",
			raid: func(raidBuffs *proto.RaidBuffs, _ *proto.IndividualBuffs, _ *proto.Debuffs) {
				// A caster group's shaman puts down Mana Spring.
				raidBuffs.ManaSpringTotem = proto.TristateEffect_TristateEffectRegular
			},
			consumes: func() *proto.Consumes {
				return &proto.Consumes{
					DefaultPotion:   proto.Potions_MajorManaPotion,
					DefaultConjured: proto.Conjured_ConjuredDemonicRune,
					FirePowerBuff:   proto.FirePowerBuff_ElixirOfGreaterFirepower,
					Flask:           proto.Flask_FlaskOfSupremePower,
					Food:            proto.Food_FoodRunnTumTuberSurprise,
					MainHandImbue:   proto.WeaponImbue_BrilliantWizardOil,
					ManaRegenElixir: proto.ManaRegenElixir_MagebloodPotion,
					ShadowPowerBuff: proto.ShadowPowerBuff_ElixirOfShadowPower,
					SpellPowerBuff:  proto.SpellPowerBuff_GreaterArcaneElixir,
					ZanzaBuff:       proto.ZanzaBuff_CerebralCortexCompound,
				}
			},
			professions: [2]proto.Profession{proto.Profession_Tailoring, proto.Profession_Enchanting},
			withSpec: func(player *proto.Player) {
				player.Class = proto.Class_ClassWarlock
				player.Spec = &proto.Player_Warlock{Warlock: &proto.Warlock{
					Options: &proto.WarlockOptions{
						Armor:         proto.WarlockOptions_DemonArmor,
						Summon:        summon,
						SacrificedPet: sacrificed,
						WeaponImbue:   imbue,
					},
				}}
			},
		}
	}

	specs["affliction"] = warlockSpec(proto.WarlockOptions_Imp, proto.WarlockOptions_NoSummon, proto.WarlockOptions_Spellstone)
	specs["demonology"] = warlockSpec(proto.WarlockOptions_Succubus, proto.WarlockOptions_Imp, proto.WarlockOptions_Spellstone)
	specs["destruction"] = warlockSpec(proto.WarlockOptions_Imp, proto.WarlockOptions_NoSummon, proto.WarlockOptions_Spellstone)
}
