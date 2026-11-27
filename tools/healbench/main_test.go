package main

import (
	"testing"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/tools/healscore"
	googleProto "google.golang.org/protobuf/proto"
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

// tankBench is the bench's -role tank on a boss for a healer request, with the defaults.
func tankBench(t *testing.T, rsr *proto.RaidSimRequest, boss string, fights int) *healscore.Bench {
	t.Helper()
	f, err := healscore.LoadFight("../encounters/" + boss + ".json")
	if err != nil {
		t.Fatal(err)
	}
	o := healscore.DefaultOptions("tank")
	o.Duration, o.Iterations = f.Duration("tank"), fights
	return &healscore.Bench{O: o, Fight: f, Tank: healscore.FirstPlayer(read("../tankbench/req/warrior_protection.json")), Healer: healscore.FirstPlayer(rsr),
		Filler: healscore.FirstPlayer(read("../dps_report/req/mage__frost.json")), Buffs: rsr.Raid.Buffs, Debuffs: rsr.Raid.Debuffs}
}

// isPace is the pace check: mana % at least the fight's remaining %.
func isPace(v *proto.APLValue) bool {
	c := v.GetCmp()
	return c != nil && c.Lhs.GetCurrentManaPercent() != nil && c.Rhs.GetRemainingTimePercent() != nil
}

// unpace drops the pace check from every And in a condition, and counts the ones it dropped.
func unpace(v *proto.APLValue) int {
	if v == nil {
		return 0
	}
	n := 0
	if and := v.GetAnd(); and != nil {
		kept := and.Vals[:0]
		for _, x := range and.Vals {
			if isPace(x) {
				n++
				continue
			}
			n += unpace(x)
			kept = append(kept, x)
		}
		and.Vals = kept
	}
	if or := v.GetOr(); or != nil {
		for _, x := range or.Vals {
			n += unpace(x)
		}
	}
	return n
}

// The priest tank healers run dry on Chromaggus, the one long single-tank stint (115 s): at the
// report's damage both end with a few hundred mana (docs/healing_priest.md, Mana on Chromaggus). Their
// Flash Heal first lines (more healing a second for less a mana) only fire while the mana keeps pace
// with the fight, so the long fight spends less on them: more healing a mana than the same rotation
// without the pace check (Holy 4.08 to 4.17, Discipline 4.24 to 4.30 at 200 fights).
func TestPriestTankHealersPaceFlashHeal(t *testing.T) {
	for _, healer := range []string{"priest_holy", "priest_discipline"} {
		paced := read("req/" + healer + ".json")
		unpaced := googleProto.Clone(paced).(*proto.RaidSimRequest)
		dropped := 0
		for _, item := range healscore.FirstPlayer(unpaced).Rotation.PriorityList {
			dropped += unpace(item.Action.GetCondition())
		}
		if dropped == 0 {
			t.Errorf("%s: no Flash Heal first line keeps pace with the fight", healer)
			continue
		}
		var perMana, manaEnd [2]float64
		for i, rsr := range []*proto.RaidSimRequest{unpaced, paced} {
			r, err := tankBench(t, rsr, "chromaggus", 200).Run(1, false)
			if err != nil {
				t.Fatal(err)
			}
			perMana[i], manaEnd[i] = r.Ehps*r.Duration/r.ManaSpent, r.ManaEnd
		}
		t.Logf("%s on Chromaggus: %.3f -> %.3f healing a mana, %.0f -> %.0f mana left", healer, perMana[0], perMana[1], manaEnd[0], manaEnd[1])
		if perMana[1] <= perMana[0]+0.02 {
			t.Errorf("%s on Chromaggus: paced %.3f healing a mana, not 0.02 above the unpaced %.3f", healer, perMana[1], perMana[0])
		}
	}
}
