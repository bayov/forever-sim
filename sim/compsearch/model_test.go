package compsearch

import (
	"fmt"
	"math"
	"os"
	"sort"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

// TestModelMisses prints the model's worst misses per member (set COMPSEARCH_DEBUG=1, and
// COMPSEARCH_CACHE to a cache directory to keep the sims).
func TestModelMisses(t *testing.T) {
	if os.Getenv("COMPSEARCH_DEBUG") == "" {
		t.Skip()
	}
	p := loadTestPool(t)
	sim, _ := NewSimmer(p, os.Getenv("COMPSEARCH_CACHE"))
	sc := &Scenario{Pool: p, Sim: sim, Faction: proto.Faction_Horde, Iterations: 200, Seed: 101,
		Fight: Fight{Duration: 180, DurationVariation: 20, FlaskZone: true}}
	var builds []*Build
	for _, b := range p.Bases {
		builds = append(builds, b)
		for _, d := range b.Duties {
			builds = append(builds, d)
		}
	}
	m, err := sc.BuildModel(builds)
	if err != nil {
		t.Fatal(err)
	}
	type miss struct {
		b    *Build
		seat Seat
		mv   float64
		ev   float64
	}
	var ms []miss
	var jobs []*Job
	var seats []Seat
	for _, up := range []Gift{0, GiftAB, GiftSunderTank, GiftMinorArmor | GiftCoE, GiftBattleShout | GiftGoA | GiftWF, GiftSoE | GiftGoA | GiftAB | GiftCoE | GiftJoW} {
		for _, bl := range [][]proto.Blessing{nil, {proto.Blessing_BlessingKings}} {
			seats = append(seats, Seat{Up: up, Blessings: bl})
		}
	}
	for _, b := range p.Bases {
		for _, s := range seats {
			s.Own = b.Casts
			jobs = append(jobs, sc.job(b, s, 200))
		}
	}
	if err := sim.Sim(jobs); err != nil {
		t.Fatal(err)
	}
	for _, j := range jobs {
		ms = append(ms, miss{j.Build, j.Seat, m.V(j.Build, j.Seat), j.Value.DPS})
	}
	sort.Slice(ms, func(i, j int) bool { return math.Abs(ms[i].mv-ms[i].ev) > math.Abs(ms[j].mv-ms[j].ev) })
	for _, x := range ms[:40] {
		fmt.Printf("%-36s %-70s model %7.1f exact %7.1f miss %+6.1f\n", x.b.ID, x.seat.Up.String()+fmt.Sprint(x.seat.Blessings), x.mv, x.ev, x.mv-x.ev)
	}
}
