package compsearch

import (
	"fmt"
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

// dropRequest is a build's request in a seat, before any normalization.
func dropRequest(p *Pool, b *Build, seat Seat, iterations int32, seed int64) (*proto.RaidSimRequest, error) {
	req := googleProto.Clone(b.Req).(*proto.RaidSimRequest)
	player, err := p.Player(b, proto.Faction_Horde)
	if err != nil {
		return nil, err
	}
	req.Raid.Parties[0].Players[0] = player
	seat.Own = b.Casts
	seat.Write(req, b)
	req.SimOptions = &proto.SimOptions{Iterations: iterations, RandomSeed: seed, UseLabeledRands: true}
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
	for _, seat := range dropContexts(camps) {
		req, err := dropRequest(p, b, seat, iterations, seed)
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
		if math.Float64bits(on) != math.Float64bits(off) {
			return false, nil
		}
	}
	return true, nil
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
