package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// registerStartingTotems puts down the totems we have standing when the fight starts.
//
// We take them as put down before the pull, so they cost no mana and no global cooldown. Each
// goes down at the pull at the highest rank we know, and then we cut it short to the time it
// has left. A Searing Totem with 20 sec left attacks every 2.5 sec from the pull on and goes
// away at 20 sec. A prepull cast in the rotation that puts a totem in the same element comes
// first, so the starting totem takes its place.
func (shaman *Shaman) registerStartingTotems() {
	totems := shaman.StartingTotems
	if totems == nil {
		return
	}

	type startingTotem struct {
		slot        int
		ranks       []*core.Spell
		secondsLeft float64
	}
	var starting []startingTotem
	add := func(slot int, ranks []*core.Spell, secondsLeft float64) {
		if ranks != nil {
			starting = append(starting, startingTotem{slot, ranks, secondsLeft})
		}
	}
	add(EarthTotem, shaman.earthTotemRanks(totems.Earth), totems.EarthSecondsLeft)
	add(AirTotem, shaman.airTotemRanks(totems.Air), totems.AirSecondsLeft)
	add(FireTotem, shaman.fireTotemRanks(totems.Fire), totems.FireSecondsLeft)
	add(WaterTotem, shaman.waterTotemRanks(totems.Water), totems.WaterSecondsLeft)
	if len(starting) == 0 {
		return
	}

	shaman.RegisterPrepullAction(0, func(sim *core.Simulation) {
		for _, totem := range starting {
			spell := highestRank(totem.ranks)
			if spell == nil {
				continue
			}
			expiresAt := time.Duration(totem.secondsLeft * float64(time.Second))
			shaman.putDownStartingTotem(sim, totem.slot, spell, expiresAt)
		}
	})
}

// putDownStartingTotem puts the totem down for free and makes it go away at expiresAt.
//
// Each totem keeps up its own things: a buff on us, a dot on the target or a heal over time
// on the party. So we look at every aura in the fight that the totem turned on just now and
// make it run out at expiresAt instead. When expiresAt is 0 or past the totem's duration, the
// totem has its full duration.
func (shaman *Shaman) putDownStartingTotem(sim *core.Simulation, slot int, spell *core.Spell, expiresAt time.Duration) {
	startedBefore := map[*core.Aura]bool{}
	for _, unit := range sim.Environment.AllUnits {
		for _, aura := range unit.GetAuras() {
			if aura.IsActive() && aura.TimeActive(sim) == 0 {
				startedBefore[aura] = true
			}
		}
	}

	spell.ApplyEffects(sim, shaman.CurrentTarget, spell)

	if expiresAt <= 0 || expiresAt >= shaman.TotemExpirations[slot] {
		return
	}
	shaman.TotemExpirations[slot] = expiresAt
	for _, unit := range sim.Environment.AllUnits {
		for _, aura := range unit.GetAuras() {
			if aura.IsActive() && aura.TimeActive(sim) == 0 && !startedBefore[aura] && aura.ExpiresAt() > expiresAt {
				aura.UpdateExpires(sim, expiresAt)
			}
		}
	}
}

func highestRank(ranks []*core.Spell) *core.Spell {
	for rank := len(ranks) - 1; rank >= 0; rank-- {
		if ranks[rank] != nil {
			return ranks[rank]
		}
	}
	return nil
}

func (shaman *Shaman) earthTotemRanks(totem proto.EarthTotem) []*core.Spell {
	switch totem {
	case proto.EarthTotem_StrengthOfEarthTotem:
		return shaman.StrengthOfEarthTotem
	case proto.EarthTotem_StoneskinTotem:
		return shaman.StoneskinTotem
	case proto.EarthTotem_TremorTotem:
		return []*core.Spell{shaman.TremorTotem}
	}
	return nil
}

func (shaman *Shaman) airTotemRanks(totem proto.AirTotem) []*core.Spell {
	switch totem {
	case proto.AirTotem_WindfuryTotem:
		return shaman.WindfuryTotem
	case proto.AirTotem_GraceOfAirTotem:
		return shaman.GraceOfAirTotem
	}
	return nil
}

func (shaman *Shaman) fireTotemRanks(totem proto.FireTotem) []*core.Spell {
	switch totem {
	case proto.FireTotem_SearingTotem:
		return shaman.SearingTotem
	case proto.FireTotem_MagmaTotem:
		return shaman.MagmaTotem
	case proto.FireTotem_FlametongueTotem:
		return shaman.FlametongueTotem
	}
	return nil
}

func (shaman *Shaman) waterTotemRanks(totem proto.WaterTotem) []*core.Spell {
	switch totem {
	case proto.WaterTotem_ManaSpringTotem:
		return shaman.ManaSpringTotem
	case proto.WaterTotem_HealingStreamTotem:
		return shaman.HealingStreamTotem
	}
	return nil
}
