package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	// 20 yd for Fire Blast, which the finisher casts (its other spells reach 30; the finisher no
	// longer casts Arcane Explosion, which reaches 10).
	specs["mage_arcane"] = mageSpec(proto.Mage_Options_ImbuePrecision, 20)
}
