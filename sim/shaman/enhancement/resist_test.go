package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	googleProto "google.golang.org/protobuf/proto"
)

// TestOrcShamanEarthShockResists checks what a target's Nature resistance does to our Earth
// Shock (shaman_audit.md 7.8).
//
// Under Forever it takes part of the damage, as it does for Lightning Bolt. On the beta the
// Elder Cloud Serpents took 30% off every Earth Shock that landed on them. Under Classic
// Earth Shock is binary. It never loses part of its damage, but it misses in full more often.
//
// A level 30 target with 75 Nature resistance has half the resistance cap at our level. Then
// 88% of the spells that can be partly resisted lose part of their damage. A binary spell
// lands on 96% times 1 - 0.75 * 0.5 of the rolls, which is 60%.
func TestOrcShamanEarthShockResists(t *testing.T) {
	const rolls = 2000
	for _, c := range []struct {
		ruleset proto.Ruleset
		landed  float64
		partial float64
	}{
		{proto.Ruleset_RulesetForever, 0.96, 0.88},
		{proto.Ruleset_RulesetClassic, 0.6, 0},
	} {
		target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
		target.Level = 30
		target.Stats[proto.Stat_StatNatureResistance] = 75
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(newOrcShaman(30, "", &proto.EnhancementShaman_Options{}), &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{target}},
			SimOptions: &proto.SimOptions{Ruleset: c.ruleset, RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()
		enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)

		earthShock := topRank(t, "Earth Shock", enh.EarthShock)
		landed, partial := 0, 0
		for range rolls {
			result := earthShock.CalcDamage(sim, enh.CurrentTarget, 100, earthShock.OutcomeMagicHit)
			if result.Landed() {
				landed++
				if result.ResistanceMultiplier < 1 {
					partial++
				}
			}
		}

		// Each count is within 4 standard errors of its chance.
		check := func(what string, got, n int, want float64) {
			if se := math.Sqrt(want * (1 - want) / float64(n)); math.Abs(float64(got)/float64(n)-want) > 4*se+1e-9 {
				t.Errorf("%s: %s on %d of %d, want %.2f", c.ruleset, what, got, n, want)
			}
		}
		check("Earth Shock landed", landed, rolls, c.landed)
		check("Earth Shock lost part of its damage", partial, landed, c.partial)
	}
}
