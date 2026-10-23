package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	// Frost has the mana to spare for a Major Spellblasting Potion in place of the Major Mana
	// Potion (+7.9 at 180 sec, +4.6 at 300, the consumables auditor). The other specs run out.
	spec := mageSpec(proto.Mage_Options_ImbueGreaterFrost)
	consumes := spec.consumes
	spec.consumes = func() *proto.Consumes {
		c := consumes()
		c.DefaultPotion = proto.Potions_MajorSpellblastingPotion
		// Swiftness is 4.8 ahead of Aggression for Frost (ui/mage/presets.ts).
		c.Flask = proto.Flask_FlaskOfNaturalSwiftness
		return c
	}
	specs["mage_frost"] = spec
}
