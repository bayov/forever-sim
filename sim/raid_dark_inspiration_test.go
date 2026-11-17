package sim

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/priest"
)

// darkInspirationRaid is a Shadow priest and a mage in party 0 and a mage alone in party 1,
// with Dark Inspiration's given form (PartyBuffs.dark_inspiration) set on both parties or on
// neither.
func darkInspirationRaid(t *testing.T, given bool) *core.Simulation {
	rsr := raidOfParties(t, "mage__fire", [][]*proto.Player{
		{idle(raidPlayer(t, "priest__shadow")), idle(raidPlayer(t, "mage__fire"))},
		{idle(raidPlayer(t, "mage__fire"))},
	}, 1)
	rsr.Raid.Buffs = &proto.RaidBuffs{}
	for _, party := range rsr.Raid.Parties {
		party.Buffs = &proto.PartyBuffs{DarkInspiration: given}
	}
	return stepRaid(rsr)
}

// Dark Inspiration's given form stands in for a Shadow priest who isn't in the party. A party
// that holds a priest with the talent gets the real aura from it instead, so the given form
// drops there, and the Spirit counts once. A party without one still gets the given form.
func TestRaidForeverDarkInspirationGivenOnce(t *testing.T) {
	base := darkInspirationRaid(t, false)
	shadowSpirit := member(base, 0, 0).GetStat(stats.Spirit)
	mageSpirit := member(base, 0, 1).GetStat(stats.Spirit)
	loneMageSpirit := member(base, 1, 0).GetStat(stats.Spirit)

	sim := darkInspirationRaid(t, true)
	shadow := sim.Raid.Parties[0].Players[0].(priest.PriestAgent).GetPriest()
	mage := member(sim, 0, 1)
	if shadow.Talents.DarkInspiration == 0 || shadow.DarkInspirationAura == nil {
		t.Fatalf("the Shadow priest has no Dark Inspiration")
	}
	near := func(what string, got, want float64) {
		t.Helper()
		if math.Abs(got-want) > 1e-6 {
			t.Errorf("%s has %.3f Spirit, want %.3f", what, got, want)
		}
	}
	// Spirit is floored after its multipliers, and the base it's floored from is floored too, so
	// a base B times m comes out between floor(B*m) and (B+1)*m. Counted twice it would be
	// some 15% over.
	multiplied := func(what string, got, base, m float64) {
		t.Helper()
		if lo, hi := math.Floor(base*m+1e-6), (base+1)*m; got < lo || got > hi {
			t.Errorf("%s has %.3f Spirit, want %.0f to %.3f (%.0f × %.4f once)", what, got, lo, hi, base, m)
		}
	}

	raidAt(sim, 1, func(sim *core.Simulation) {
		near("the priest, before Dark Inspiration", shadow.GetStat(stats.Spirit), shadowSpirit)
		near("the priest's party's mage, before Dark Inspiration", mage.GetStat(stats.Spirit), mageSpirit)
		multiplied("the other party's mage", member(sim, 1, 0).GetStat(stats.Spirit), loneMageSpirit, 1+0.15*core.DarkInspirationUptime)

		mb := topRank(t, shadow.MindBlast)
		mb.BonusHitRating = 1e6
		castFree(sim, mb)
		if !shadow.DarkInspirationAura.IsActive() {
			t.Fatalf("Mind Blast didn't bring up Dark Inspiration")
		}
		multiplier := 1 + 0.03*float64(shadow.Talents.DarkInspiration)
		multiplied("the priest, under Dark Inspiration", shadow.GetStat(stats.Spirit), shadowSpirit, multiplier)
		multiplied("the priest's party's mage, under Dark Inspiration", mage.GetStat(stats.Spirit), mageSpirit, multiplier)
	})
	runSteps(sim)
}
