package main

import (
	"testing"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/tools/tankscore"
)

func init() { sim.RegisterAll() }

// raidBench is the bench's -role raid on a boss for a healer from req, with the defaults.
func raidBench(t *testing.T, healer, boss string) *bench {
	t.Helper()
	path := "../encounters/" + boss + ".json"
	enc, err := tankscore.LoadEncounter(path)
	if err != nil {
		t.Fatal(err)
	}
	f := encounterFight(enc)
	f.killSeconds = killSeconds(path)
	o := options{stopCast: 0.5, reaction: 0.3, role: "raid", scale: 1, healers: 2, fillers: 7, duration: f.killSeconds}
	rsr := read("req/" + healer + ".json")
	return &bench{o: o, fight: f, tank: firstPlayer(read("../tankbench/req/warrior_protection.json")), healer: firstPlayer(rsr),
		filler: firstPlayer(read("../dps_report/req/mage__frost.json")), buffs: rsr.Raid.Buffs, debuffs: rsr.Raid.Debuffs}
}

// fightHps sims n fights from seed in one sim (one split of a concurrent run) and gives each
// fight's healing per second of the healers.
func fightHps(t *testing.T, b *bench, seed int64, n int) []float64 {
	t.Helper()
	b.o.seed, b.o.iterations, b.o.allValues = seed, n, true
	res := core.RunRaidSim(b.request(1, false))
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	hps := make([]float64, n)
	for pi, party := range res.RaidMetrics.Parties {
		for si, m := range party.Players {
			if b.slots[pi][si] == roleHealer {
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
