package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	// Frost has the mana to spare for a Major Spellblasting Potion in place of the Major Mana
	// Potion (+7.9 at 180 sec, +4.6 at 300, the consumables auditor). The other specs run out.
	// 30 yd for Arcane Missiles, Arcane Blast and the wand (Arctic Reach 1/2 takes Frostbolt and
	// Ice Lance to 33).
	spec := mageSpec(proto.Mage_Options_ImbueGreaterFrost, 30)
	consumes := spec.consumes
	spec.consumes = func() *proto.Consumes {
		c := consumes()
		c.DefaultPotion = proto.Potions_MajorSpellblastingPotion
		return c
	}
	specs["mage_frost"] = spec
}
