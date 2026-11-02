package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	// 30 yd for Scorch, Arcane Missiles and the wand (Improved Fire Blast 2/2 takes Fire Blast to
	// 35, Fireball and Pyroblast reach 35).
	specs["mage_fire"] = mageSpec(proto.Mage_Options_ImbueGreaterFlame, 30)
}
