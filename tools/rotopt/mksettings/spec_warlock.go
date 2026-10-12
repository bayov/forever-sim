package main

import "github.com/wowsims/classic/sim/core/proto"

// The three warlock specs share the ui folder, the raid and the consumes. They differ in the
// demon that is out and the one Demonic Sacrifice took before the pull.
func init() {
	warlockSpec := func(summon, sacrificed proto.WarlockOptions_Summon, imbue proto.WarlockOptions_WeaponImbue) *specSetup {
		return &specSetup{
			uiDir: "ui/warlock",
			raid: func(raidBuffs *proto.RaidBuffs, indBuffs *proto.IndividualBuffs, _ *proto.Debuffs) {
				// A caster group's shaman puts down Mana Spring.
				raidBuffs.ManaSpringTotem = proto.TristateEffect_TristateEffectRegular
				// Casters take no Blessing of Might, and the demon only gets the warlock's blessings.
				indBuffs.BlessingOfMight = proto.TristateEffect_TristateEffectMissing
			},
			// Every consumable the sim applies was tried in its slot on the Demonology build at
			// 180 sec (2026-10-12): Goblin Sapper +4.1 (Engineering), Cerebral Cortex Compound +10.7
			// (a Blasted Lands quest), Demonic Rune +15.8, Wizard Oil +15.7, Elixir of Shadow Power
			// +23.3. With Forever's elixirs and foods simulated: one of the five +2% crit elixirs (they
			// are role variants of one elixir), Elixir of Sages with Elixir of Greater Intellect, which
			// is 0.1 to 0.6 ahead of Elixir of the Owl with Elixir of Greater Spirit on every build
			// (Sages' Spirit feeds Life Tap), Greater
			// Mageblood Elixir +1.9 over Mageblood Potion, Lobster Stew +3.9 over Nightfin Soup, and
			// Major Spellblasting Potion 12.9 behind Major Mana Potion. Forever's Phase 1 has no
			// vanilla raid open, so Brilliant Wizard Oil (a Zandalar recipe) is out. Elixir of
			// Greater Firepower is Forever's Elixir of Holy Power, so the plain Elixir of Firepower
			// takes the slot. Goblin Land Mine (-7) and Flask of Distilled Wisdom (-88) lost.
			consumes: func() *proto.Consumes {
				return &proto.Consumes{
					DefaultPotion:   proto.Potions_MajorManaPotion,
					DefaultConjured: proto.Conjured_ConjuredDemonicRune,
					FirePowerBuff:   proto.FirePowerBuff_ElixirOfFirepower,
					Flask:           proto.Flask_FlaskOfSupremePower,
					Food:            proto.Food_FoodLobsterStew,
					IntellectElixir: proto.IntellectElixir_ElixirOfGreaterIntellect,
					MainHandImbue:   proto.WeaponImbue_WizardOil,
					ManaRegenElixir: proto.ManaRegenElixir_GreaterMagebloodElixir,
					SapperExplosive: proto.SapperExplosive_SapperGoblinSapper,
					ShadowPowerBuff: proto.ShadowPowerBuff_ElixirOfShadowPower,
					SpellPowerBuff:  proto.SpellPowerBuff_GreaterArcaneElixir,
					SpiritElixir:    proto.SpiritElixir_ElixirOfSages,
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
