package compsearch

import (
	"fmt"
	"io"
	"math"
	"runtime"
	"slices"
	"sync"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// dropContexts are the seats a drop must hold under: the full set and the empty set, with the camps
// on or off. Each camp setting has its own table, since with camps on Moonkin Aura takes the Camp Chair
// from a warrior.
func dropContexts(camps bool) []Seat {
	full := Seat{Up: ^Gift(0), Blessings: []proto.Blessing{proto.Blessing_BlessingKings, proto.Blessing_BlessingMight, proto.Blessing_BlessingWisdom},
		Package: proto.RaidPackage_RaidPackageRestoration, PI: true, Innervate: true, Camps: camps}
	return []Seat{full, {Camps: camps}}
}

// DropKey is the drop table's key for a build and a camp setting.
func DropKey(id string, camps bool) string {
	if camps {
		return id + "+camps"
	}
	return id
}

// dropVariant is one more way a drop must hold: a faction (its race's own rotation), a fight length (0
// for the request's own 180 ± 20 s) and a seed. The cache uses the table at every length, so it must hold
// at the short, long and very long ones too (an Enhancement shaman's own 5-minute totems run out).
type dropVariant struct {
	faction  proto.Faction
	duration float64
	seed     int64
}

func dropVariants(seed int64) []dropVariant {
	// Iteration i of a run uses seed + i, so a second sample needs a seed past the first run's iterations.
	return []dropVariant{{proto.Faction_Horde, 0, seed}, {proto.Faction_Horde, 0, seed + 100_000}, {proto.Faction_Alliance, 0, seed},
		{proto.Faction_Horde, 120, seed}, {proto.Faction_Horde, 300, seed}, {proto.Faction_Horde, 420, seed}}
}

// dropRequest is a build's request in a seat, before any normalization.
func dropRequest(p *Pool, b *Build, seat Seat, iterations int32, v dropVariant) (*proto.RaidSimRequest, error) {
	req := googleProto.Clone(b.Req).(*proto.RaidSimRequest)
	player, err := p.Player(b, v.faction)
	if err != nil {
		return nil, err
	}
	req.Raid.Parties[0].Players[0] = player
	seat.Own = b.Casts
	seat.Write(req, b)
	if v.duration > 0 {
		req.Encounter.Duration, req.Encounter.DurationVariation = v.duration, 0
	}
	req.SimOptions = &proto.SimOptions{Iterations: iterations, RandomSeed: v.seed, UseLabeledRands: true}
	return req, nil
}

// withFields is the request with the fields set on, or cleared.
func withFields(req *proto.RaidSimRequest, fields []string, on bool) *proto.RaidSimRequest {
	r := googleProto.Clone(req).(*proto.RaidSimRequest)
	for _, name := range fields {
		f := fieldByName[name]
		if on {
			f.Set(r)
		} else {
			f.Clear(r)
		}
	}
	return r
}

func dpsOf(req *proto.RaidSimRequest) (float64, error) {
	res := core.RunRaidSim(req)
	if res.Error != nil {
		return 0, fmt.Errorf("%s", res.Error.Message)
	}
	return res.RaidMetrics.Dps.Avg, nil
}

// unread is whether the build's result is bit-identical with the fields set and with them cleared, in
// every context.
func unread(p *Pool, b *Build, camps bool, fields []string, iterations int32, seed int64) (bool, error) {
	for _, v := range dropVariants(seed) {
		for _, seat := range dropContexts(camps) {
			ok, err := unreadIn(p, b, seat, fields, iterations, v)
			if err != nil || !ok {
				return false, err
			}
		}
	}
	return true, nil
}

func unreadIn(p *Pool, b *Build, seat Seat, fields []string, iterations int32, v dropVariant) (bool, error) {
	req, err := dropRequest(p, b, seat, iterations, v)
	if err != nil {
		return false, err
	}
	on, err := dpsOf(withFields(req, fields, true))
	if err != nil {
		return false, err
	}
	off, err := dpsOf(withFields(req, fields, false))
	if err != nil {
		return false, err
	}
	return math.Float64bits(on) == math.Float64bits(off), nil
}

// ComputeDrops finds, for every build, the fields it can't read: each one alone, then all together.
func ComputeDrops(p *Pool, iterations int32, seed int64, progress func(string)) (map[string][]string, error) {
	ids := make([]string, 0, len(p.Builds))
	for id := range p.Builds {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	out := map[string][]string{}
	var mu sync.Mutex
	var firstErr error
	sem := make(chan struct{}, runtime.GOMAXPROCS(0))
	var wg sync.WaitGroup
	for _, id := range ids {
		b := p.Builds[id]
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			var counts [2]int
			for ci, camps := range []bool{false, true} {
				var drops []string
				for _, f := range Fields {
					if !camps && len(f.Name) > 5 && f.Name[:5] == "camp." {
						continue
					}
					ok, err := unread(p, b, camps, []string{f.Name}, iterations, seed)
					if err != nil {
						mu.Lock()
						firstErr = fmt.Errorf("%s %s: %w", b.ID, f.Name, err)
						mu.Unlock()
						return
					}
					if ok {
						drops = append(drops, f.Name)
					}
				}
				// Together they must hold too; drop the last one until they do.
				for len(drops) > 0 {
					ok, err := unread(p, b, camps, drops, iterations, seed)
					if err != nil {
						mu.Lock()
						firstErr = err
						mu.Unlock()
						return
					}
					if ok {
						break
					}
					drops = drops[:len(drops)-1]
				}
				mu.Lock()
				out[DropKey(b.ID, camps)] = drops
				mu.Unlock()
				counts[ci] = len(drops)
			}
			if progress != nil {
				progress(fmt.Sprintf("%s: %d unread, %d with camps", b.ID, counts[0], counts[1]))
			}
		}()
	}
	wg.Wait()
	return out, firstErr
}

// MeasureLeftOut sims every base build's report request with and without the buffs the resolver
// leaves out (docs/raid_sim.md, Writing a role into a request): Thorns, Retribution Aura, Salvation and
// Judgement of Light.
func MeasureLeftOut(p *Pool, w io.Writer) error {
	fmt.Fprintln(w, "| Build | Report set | Without Thorns, Retribution Aura, Salvation, Judgement of Light | Difference | Bit-identical |")
	fmt.Fprintln(w, "|---|---|---|---|---|")
	for _, b := range p.Bases {
		req := googleProto.Clone(b.Req).(*proto.RaidSimRequest)
		req.SimOptions = &proto.SimOptions{Iterations: 2000, RandomSeed: 1}
		with, err := dpsOf(req)
		if err != nil {
			return err
		}
		req.Raid.Buffs.Thorns = missing
		req.Raid.Buffs.RetributionAura = missing
		req.Raid.Buffs.Salvation = false
		req.Raid.Debuffs.JudgementOfLight = false
		without, err := dpsOf(req)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "| %s | %.2f | %.2f | %+.2f | %v |\n", b.ID, with, without, without-with, math.Float64bits(with) == math.Float64bits(without))
	}
	return nil
}
