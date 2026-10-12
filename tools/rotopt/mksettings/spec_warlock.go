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
			// Every consumable the sim applies was tried in its slot on the Demonology build at
			// 180 sec (2026-10-12): Nightfin Soup +7.8 over Runn Tum Tuber Surprise, Goblin Sapper
			// +4.1 (Engineering), Mageblood Potion +4.4, Cerebral Cortex Compound +10.7, Demonic Rune
			// +15.8, Major Mana Potion +23.7, Brilliant Wizard Oil +32.5, Elixir of Shadow Power +23.3.
			// Elixir of Greater Firepower is Forever's Elixir of Holy Power, so the plain Elixir of
			// Firepower takes the slot. Goblin Land Mine (-7) and Flask of Distilled Wisdom (-88) lost.
			consumes: func() *proto.Consumes {
				return &proto.Consumes{
					DefaultPotion:   proto.Potions_MajorManaPotion,
					DefaultConjured: proto.Conjured_ConjuredDemonicRune,
					FirePowerBuff:   proto.FirePowerBuff_ElixirOfFirepower,
					Flask:           proto.Flask_FlaskOfSupremePower,
					Food:            proto.Food_FoodNightfinSoup,
					MainHandImbue:   proto.WeaponImbue_BrilliantWizardOil,
					ManaRegenElixir: proto.ManaRegenElixir_MagebloodPotion,
					SapperExplosive: proto.SapperExplosive_SapperGoblinSapper,
					ShadowPowerBuff: proto.ShadowPowerBuff_ElixirOfShadowPower,
					SpellPowerBuff:  proto.SpellPowerBuff_GreaterArcaneElixir,
					ZanzaBuff:       proto.ZanzaBuff_CerebralCortexCompound,
				}
			},
			professions: [2]proto.Profession{proto.Profession_Engineering, proto.Profession_Tailoring},
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

	// The Level 60 builds' demons and imbues (ui/warlock/presets.ts).
	specs["demonology"] = warlockSpec(proto.WarlockOptions_Succubus, proto.WarlockOptions_Imp, proto.WarlockOptions_Spellstone)
	specs["destruction"] = warlockSpec(proto.WarlockOptions_Succubus, proto.WarlockOptions_NoSummon, proto.WarlockOptions_Spellstone)
	specs["affliction"] = warlockSpec(proto.WarlockOptions_Succubus, proto.WarlockOptions_NoSummon, proto.WarlockOptions_Spellstone)
	specs["destruction_fire"] = warlockSpec(proto.WarlockOptions_NoSummon, proto.WarlockOptions_Succubus, proto.WarlockOptions_Firestone)
	specs["affliction_dot"] = warlockSpec(proto.WarlockOptions_Succubus, proto.WarlockOptions_NoSummon, proto.WarlockOptions_Spellstone)

	// Every demon and sacrifice pair for the talent searches, as warlock_<demon out>_<demon
	// sacrificed>_<imbue>, warlock_succubus_imp_spellstone say.
	pets := map[string]proto.WarlockOptions_Summon{
		"none": proto.WarlockOptions_NoSummon, "imp": proto.WarlockOptions_Imp, "voidwalker": proto.WarlockOptions_Voidwalker,
		"succubus": proto.WarlockOptions_Succubus, "felhunter": proto.WarlockOptions_Felhunter,
	}
	imbues := map[string]proto.WarlockOptions_WeaponImbue{
		"none": proto.WarlockOptions_NoWeaponImbue, "spellstone": proto.WarlockOptions_Spellstone, "firestone": proto.WarlockOptions_Firestone,
	}
	for outName, out := range pets {
		for sacName, sac := range pets {
			for imbueName, imbue := range imbues {
				specs["warlock_"+outName+"_"+sacName+"_"+imbueName] = warlockSpec(out, sac, imbue)
			}
		}
	}
}
