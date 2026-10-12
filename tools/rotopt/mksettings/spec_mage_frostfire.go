package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["mage_frostfire"] = mageSpec(proto.FirePowerBuff_ElixirOfGreaterFirepower, proto.FrostPowerBuff_ElixirOfFrostPower)
}
