package main

import (
	"slices"

	"github.com/wowsims/classic/sim/core/proto"
)

// specSetup is what one spec adds to the shared level 60 raid in main.go: its ui folder, its
// consumes and its class options. Each spec registers its own in a spec_<name>.go file with
// an init function, so adding a spec never touches another spec's file.
type specSetup struct {
	// uiDir is the spec's ui folder, which holds its gear_sets and apls.
	uiDir string
	// raid changes the shared level 60 raid buffs, blessings and debuffs for the spec, a
	// shaman's totems for a melee spec say. Optional.
	raid func(raidBuffs *proto.RaidBuffs, buffs *proto.IndividualBuffs, debuffs *proto.Debuffs)
	// partyBuffs are the spec's party buffs at level 60. Optional, none when left out.
	partyBuffs func() *proto.PartyBuffs
	// consumes are the level 60 consumes.
	consumes func() *proto.Consumes
	// lowLevel sets up the solo build below level 60, given the -blessing flag. Optional,
	// the spec then runs with no consumes and Blessing of Kings.
	lowLevel func(blessing string) (lowLevelSetup, error)
	// professions replace Engineering and Alchemy when set.
	professions [2]proto.Profession
	// withSpec sets the player's class and spec options.
	withSpec func(player *proto.Player)
}

// lowLevelSetup is a spec's build below level 60. Nil fields keep the defaults.
type lowLevelSetup struct {
	consumes *proto.Consumes
	buffs    *proto.IndividualBuffs
	// tankSelf puts the boss on the player, so procs on being hit fire (Lightning Shield,
	// Retribution Aura).
	tankSelf bool
}

var specs = map[string]*specSetup{}

func specNames() []string {
	names := make([]string, 0, len(specs))
	for name := range specs {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
