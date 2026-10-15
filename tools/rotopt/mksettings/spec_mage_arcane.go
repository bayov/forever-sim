package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["mage_arcane"] = mageSpec(proto.Mage_Options_ImbuePrecision)
}
