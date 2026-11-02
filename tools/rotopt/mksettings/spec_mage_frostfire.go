package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	// 20 yd for Fire Blast, cast on cooldown above 80% mana (no Improved Fire Blast or Flame
	// Throwing in the build; its other spells reach 30 or more).
	specs["mage_frostfire"] = mageSpec(proto.Mage_Options_ImbueGreaterFlame, 20)
}
