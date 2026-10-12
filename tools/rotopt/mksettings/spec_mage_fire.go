package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["mage_fire"] = mageSpec(proto.FirePowerBuff_ElixirOfGreaterFirepower, proto.FrostPowerBuff_FrostPowerBuffUnknown)
}
