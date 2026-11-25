package warrior

import (
	"github.com/wowsims/classic/sim/core"
)

// Reaches in yards (client 71091, SpellRange and SpellRadius). Only a leveling session has
// distances.
const (
	meleeReach             = 5.0
	whirlwindRadius        = 8.0
	thunderClapRadius      = 8.0
	demoralizingShoutReach = 10.0
)

// knownRank is the highest rank of an ability the warrior knows, 0 when it knows none. The tables
// are by rank, with rank 0 unused, from the lowest to the highest. An APL that names any of the
// ranks casts the one the warrior knows.
func (warrior *Warrior) knownRank(levels []int32, spellIDs []int32) int {
	core.RegisterSpellRanks(spellIDs[1:]...)
	return core.HighestKnownRank(&warrior.Character, levels, spellIDs)
}

// areaTargets are the enemies an area ability hits, see core's AreaTargets.
func (warrior *Warrior) areaTargets(sim *core.Simulation, target *core.Unit, radius float64, limit int) []*core.Unit {
	warrior.areaBuffer = sim.Encounter.AreaTargets(sim, target, radius, limit, warrior.areaBuffer)
	return warrior.areaBuffer
}
