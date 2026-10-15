package main

import "github.com/wowsims/classic/sim/core/proto"

// The level 60 rogue setup the Level 60 presets were searched under, shared by the
// rogue_combat, rogue_assassination and rogue_subtlety specs. The rogue is in a melee
// group with a warrior and an enhancement shaman, so it gets Battle Shout, Strength of
// Earth and Windfury Totem. Under Forever only one air totem works in a party, and
// Windfury Totem is 24 DPS ahead of Grace of Air for a Sinister Strike rogue (180 sec).
// The rest of the shared raid stays: both blessings (Wisdom does nothing for a rogue),
// every debuff but Curse of Shadow, which Forever folds into Curse of the Elements (Sunder
// Armor, Faerie Fire and Curse of Recklessness are the ones that matter to the rogue,
// Expose Armor does not stack with Sunder).
func rogueLevel60Raid(raidBuffs *proto.RaidBuffs, _ *proto.IndividualBuffs, debuffs *proto.Debuffs) {
	raidBuffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectRegular
	raidBuffs.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffWindfury
}

// rogueLevel60Consumes are the level 60 consumes, each kept because it helps (Sinister
// Strike on the synthetic swords, 180 sec, 10000 iterations, against the set without it):
// Dragonbreath Chili +10.5, Thistle Tea +10.6, Goblin Land Mine +5.2 on top of the Goblin
// Sapper +3.8. Among Forever's consumables (2026-10-12) Elixir of Ferocity is 5.4 ahead of
// Juju Power (3.3 for Backstab), Major Frenzy Potion 4.7 ahead of the Frenzy Potion (which
// is 2.4 over none), and Flank au Poivre 3.5 ahead of Grilled Squid and 0.3 ahead of
// Mightfish Steak. Elixir of Cunning loses 12 to the Elixir of the Mongoose. Ground Scorpok
// Assay is 4.9 ahead of R.O.I.D.S. The caster elixirs are left out, as for the warrior and
// the feral druid: Flask of Supreme Power and Greater Arcane Elixir add 8 DPS, all of it
// through Dragonbreath Chili scaling with spell power, and no rogue drinks them. The Elixir
// of the Mongoose is the one +2% crit elixir. Both blades carry a Dense Sharpening Stone (+8
// damage): Forever's poisons are class imbues beside it (see rogueWithSpec). Elemental Sharpening
// Stone's plans are Molten Core only.
func rogueLevel60Consumes() *proto.Consumes {
	return &proto.Consumes{
		AgilityElixir:     proto.AgilityElixir_ElixirOfTheMongoose,
		AttackPowerBuff:   proto.AttackPowerBuff_JujuMight,
		DefaultConjured:   proto.Conjured_ConjuredRogueThistleTea,
		DefaultPotion:     proto.Potions_MajorFrenzyPotion,
		DragonBreathChili: true,
		FillerExplosive:   proto.Explosive_ExplosiveGoblinLandMine,
		Food:              proto.Food_FoodFlankAuPoivre,
		MainHandImbue:     proto.WeaponImbue_DenseSharpeningStone,
		OffHandImbue:      proto.WeaponImbue_DenseSharpeningStone,
		SapperExplosive:   proto.SapperExplosive_SapperGoblinSapper,
		StrengthBuff:      proto.StrengthBuff_ElixirOfFerocity,
		ZanzaBuff:         proto.ZanzaBuff_GroundScorpokAssay,
	}
}

// rogueWithSpec puts Deadly Poison on the main hand and Instant Poison on the off hand, the best
// pair for every spec. Forever's poisons are class imbues (client 70338, SpellEffect 360), so they
// go in RogueOptions and leave the consumable imbue slots to the sharpening stones.
func rogueWithSpec(player *proto.Player) {
	player.Class = proto.Class_ClassRogue
	player.Spec = &proto.Player_Rogue{Rogue: &proto.Rogue{Options: &proto.RogueOptions{
		MainHandPoison: proto.WeaponImbue_DeadlyPoison,
		OffHandPoison:  proto.WeaponImbue_InstantPoison,
	}}}
}
