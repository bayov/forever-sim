package main

import (
	"testing"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/tools/healscore"
)

func init() { sim.RegisterAll() }

// raidBench is the bench's -role raid on a boss for a healer from req, with the defaults.
func raidBench(t *testing.T, healer, boss string) *healscore.Bench {
	t.Helper()
	f, err := healscore.LoadFight("../encounters/" + boss + ".json")
	if err != nil {
		t.Fatal(err)
	}
	o := healscore.Options{StopCast: 0.5, Reaction: 0.3, Role: "raid", Scale: 1, Healers: 2, Fillers: 7, Duration: f.KillSeconds}
	rsr := read("req/" + healer + ".json")
	return &healscore.Bench{O: o, Fight: f, Tank: healscore.FirstPlayer(read("../tankbench/req/warrior_protection.json")), Healer: healscore.FirstPlayer(rsr),
		Filler: healscore.FirstPlayer(read("../dps_report/req/mage__frost.json")), Buffs: rsr.Raid.Buffs, Debuffs: rsr.Raid.Debuffs}
}

// fightHps sims n fights from seed in one sim (one split of a concurrent run) and gives each
// fight's healing per second of the healers.
func fightHps(t *testing.T, b *healscore.Bench, seed int64, n int) []float64 {
	t.Helper()
	b.O.Seed, b.O.Iterations, b.O.AllValues = seed, n, true
	req, err := b.Request(1, false)
	if err != nil {
		t.Fatal(err)
	}
	res := core.RunRaidSim(req)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	hps := make([]float64, n)
	for pi, party := range res.RaidMetrics.Parties {
		for si, m := range party.Players {
			if b.Slots[pi][si] == healscore.RoleHealer {
				for i := range hps {
					hps[i] += m.Hps.AllValues[i]
				}
			}
		}
	}
	return hps
}

// A fight comes out the same whichever fights the sim ran before it, so a split's results don't
// depend on how many fights it has (or the machine's cores).
//
// A healer's rotation that found nothing to do skips its next checks until an action could be
// ready (apl_idle.go). It used to miss that a hit shows to the healer only a reaction time after it
// lands, with no event then, so it saw the hit late. Whether it skipped at all depends on how much
// skipping paid in the fights before (aplIdleTries), so a fight's healing depended on them.
func TestRaidFightsArePrefixStable(t *testing.T) {
	for _, c := range []struct {
		healer, boss string
		fights       int
	}{
		{"priest_holy_raid", "magmadar", 40},
		{"shaman_restoration_raid", "onyxia", 20},
	} {
		b := raidBench(t, c.healer, c.boss)
		const seed = 101
		together := fightHps(t, b, seed, c.fights)
		for i := range together {
			if alone := fightHps(t, b, seed+int64(i), 1)[0]; alone != together[i] {
				t.Errorf("%s on %s, fight %d (seed %d): %.3f HPS alone, %.3f after %d other fights",
					c.healer, c.boss, i, seed+int64(i), alone, together[i], i)
			}
		}
	}
}
