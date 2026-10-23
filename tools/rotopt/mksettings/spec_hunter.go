package main

import "github.com/wowsims/classic/sim/core/proto"

// The three hunter specs share ui/hunter. Beast Mastery and Marksmanship shoot from range (25
// yards), Survival is the melee spec under Forever (Strider Kick, Mongoose Bite through Expose
// Prey, Lacerating Strikes, Predator's Edge) and stands at the boss.
func init() {
	specs["hunter_bm"] = hunterSpec(25, false, hunterRangedRaid)
	specs["hunter_mm"] = hunterSpec(25, false, hunterRangedRaid)
	// Lone Wolf is Marksmanship with the pet dismissed (Lone Wolf's 20% needs no pet out).
	lw := hunterSpec(25, false, hunterRangedRaid)
	withCat := lw.withSpec
	lw.withSpec = func(player *proto.Player) {
		withCat(player)
		player.GetHunter().Options.PetType = proto.Hunter_Options_PetNone
	}
	specs["hunter_lw"] = lw
	specs["hunter_sv"] = hunterSpec(5, true, hunterMeleeRaid)

	// Hunter's Mark and Scorpid Sting duties (standards 23): the same raid with that debuff off, so
	// it comes only from the hunter's own casts (the rotation's mark and scorpid knobs). Scorpid
	// Sting is a duty of the ranged hunters only (user ruling 2026-10-15), not of melee Survival.
	for _, name := range []string{"hunter_bm", "hunter_mm", "hunter_lw", "hunter_sv"} {
		base := *specs[name]
		raid := base.raid
		base.raid = func(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, debuffs *proto.Debuffs) {
			raid(raidBuffs, buffs, debuffs)
			debuffs.HuntersMark = proto.TristateEffect_TristateEffectMissing
		}
		specs[name+"_mark"] = &base
		if name == "hunter_sv" {
			continue
		}

		// Scorpid Sting duty: the raid's Scorpid Sting off (the rotation's scorpid knob).
		scorpid := *specs[name]
		scorpidRaid := scorpid.raid
		scorpid.raid = func(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, debuffs *proto.Debuffs) {
			scorpidRaid(raidBuffs, buffs, debuffs)
			debuffs.ScorpidSting = false
		}
		specs[name+"_scorpid"] = &scorpid
	}
}

// A ranged hunter's raid: Hunter's Mark from the raid (another hunter keeps it up, standards 23; the
// _mark specs switch it off for the hunter on duty), Blessing of Wisdom next to Might (a mana user),
// and a shaman in its group for Grace of Air (Agility is two ranged attack power and crit) and
// Mana Spring. Trueshot Aura from the raid's Marksmanship hunter, as for every other class: client
// 1.60.1.70425 made it a raid aura (it was a party aura, so hunters went without it). The melee
// debuffs and Faerie Fire stay, the physical shots go through armor.
func hunterRangedRaid(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, debuffs *proto.Debuffs) {
	raidBuffs.GraceOfAirTotem = proto.TristateEffect_TristateEffectRegular
	raidBuffs.ManaSpringTotem = proto.TristateEffect_TristateEffectRegular
	debuffs.HuntersMark = proto.TristateEffect_TristateEffectImproved
}

// Survival melees, so it joins a melee group: Windfury Totem and Strength of Earth, plus Mana Spring,
// the raid's Trueshot Aura, and Grace of Air from the raid's second shaman (the user's ruling 2026-10-16, "both totems": client 1.60.1.70425 lets Grace of Air stack with Windfury Totem and makes both raid-wide).
func hunterMeleeRaid(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, debuffs *proto.Debuffs) {
	raidBuffs.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffWindfury
	raidBuffs.GraceOfAirTotem = proto.TristateEffect_TristateEffectRegular
	raidBuffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectRegular
	raidBuffs.ManaSpringTotem = proto.TristateEffect_TristateEffectRegular
	debuffs.HuntersMark = proto.TristateEffect_TristateEffectImproved
}

func hunterSpec(distance float64, melee bool, raid func(*proto.RaidBuffs, *proto.IndividualBuffs, *proto.Debuffs)) *specSetup {
	return &specSetup{
		uiDir: "ui/hunter",
		raid:  raid,
		consumes: func() *proto.Consumes {
			return hunterConsumes(melee)
		},
		professions: [2]proto.Profession{proto.Profession_Engineering, proto.Profession_Alchemy},
		withSpec: func(player *proto.Player) {
			player.Class = proto.Class_ClassHunter
			player.DistanceFromTarget = distance
			player.Spec = &proto.Player_Hunter{Hunter: &proto.Hunter{
				Options: &proto.Hunter_Options{
					Ammo:           proto.Hunter_Options_SwiftfeatherArrow,
					QuiverBonus:    proto.Hunter_Options_Speed15,
					PetType:        proto.Hunter_Options_Cat,
					PetAttackSpeed: proto.Hunter_Options_One,
					PetUptime:      1,
					CobraReflexes:  true,
				},
			}}
		},
	}
}

// The level 60 consumes. On 2026-10-12 we tried every consumable the sim applies in each slot,
// one at a time, on the Beast Mastery, Lone Wolf and Survival builds at 180 sec, and again with
// Forever's consumables after the shared consumables merge. All keep Elixir of the Mongoose (their
// one +2% crit elixir; +27 over none), Ground Scorpok Assay (+14), Juju Might (+8) and
// Engineering's Goblin Land Mine (+7.6) and Goblin Sapper (+4.8), and take Forever's Elixir of
// Ferocity (18 Strength and Agility, no crit: +11 Beast Mastery, +13 Lone Wolf, +4.8 Survival over
// Juju Power), Flank au Poivre (20 Agility: +4 to +6 over Grilled Squid; Mightfish Steak +0.2 to
// +1.6) and Elixir of Greater Intellect (+4 to +5 over Elixir of Wisdom). The ranged specs need
// Major Mana Potion and Demonic Rune (+28): Major Frenzy Potion is +5 at 120 and 180 sec for Beast
// Mastery but 19 behind at 300 sec once the mana runs out, and behind for Lone Wolf at every
// length. They take Greater Mageblood Elixir (Mageblood Potion's recipe only comes from Zandalar)
// and Flask of Distilled Wisdom (+1.1). Survival never runs low on mana, takes Major Frenzy
// Potion (+5.9), Dragonbreath Chili (+3.2) and a Dense Sharpening Stone on each sword (+9.2; the
// sim keeps Windfury Totem in its own slot, and Elemental Sharpening Stone's plans only drop in
// Molten Core). The ranged specs have no imbue: the two-hander never swings and a bow takes none.
// Fire-toasted Bun (+9 for Survival) is left out: Forever's is a seasonal item that expires 14
// days after looting (item 23327). No pet consumables: Forever pets take no player buffs, and
// Juju and the scrolls are player-cast buffs on the pet (orchestrator ruling; they were +17 for
// Beast Mastery).
func hunterConsumes(melee bool) *proto.Consumes {
	consumes := &proto.Consumes{
		AgilityElixir:   proto.AgilityElixir_ElixirOfTheMongoose,
		AttackPowerBuff: proto.AttackPowerBuff_JujuMight,
		FillerExplosive: proto.Explosive_ExplosiveGoblinLandMine,
		Food:            proto.Food_FoodFlankAuPoivre,
		IntellectElixir: proto.IntellectElixir_ElixirOfGreaterIntellect,
		SapperExplosive: proto.SapperExplosive_SapperGoblinSapper,
		StrengthBuff:    proto.StrengthBuff_ElixirOfFerocity,
		ZanzaBuff:       proto.ZanzaBuff_GroundScorpokAssay,
	}
	if melee {
		// The natural flasks (standards 11, ui/hunter/presets.ts): Precision for Survival, Aggression
		// for the ranged specs.
		consumes.Flask = proto.Flask_FlaskOfNaturalPrecision
		consumes.DefaultPotion = proto.Potions_MajorFrenzyPotion
		consumes.DragonBreathChili = true
		consumes.MainHandImbue = proto.WeaponImbue_DenseSharpeningStone
		consumes.OffHandImbue = proto.WeaponImbue_DenseSharpeningStone
	} else {
		consumes.DefaultPotion = proto.Potions_MajorManaPotion
		consumes.DefaultConjured = proto.Conjured_ConjuredDemonicRune
		consumes.ManaRegenElixir = proto.ManaRegenElixir_GreaterMagebloodElixir
		consumes.Flask = proto.Flask_FlaskOfNaturalAggression
	}
	return consumes
}
