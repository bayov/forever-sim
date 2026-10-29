package compsearch

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	googleProto "google.golang.org/protobuf/proto"
)

// ErrAborted is a search stopped by Abort.
var ErrAborted = errors.New("aborted")

var server struct {
	sync.Mutex
	pool *Pool
	sim  *Simmer
}

// Root is the forever-sim checkout the native server searches from: $FOREVER_SIM_ROOT, else the working
// directory or one above it. The search reads the report's requests there (tools/dps_report/req).
func Root() (string, error) {
	cands := []string{os.Getenv("FOREVER_SIM_ROOT"), ".", ".."}
	for _, c := range cands {
		if c == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(c, "tools", "dps_report", "req")); err == nil {
			return filepath.Abs(c)
		}
	}
	return "", errors.New("the comp search needs a forever-sim checkout: run the server from one, or set FOREVER_SIM_ROOT")
}

func serverSearch() (*Pool, *Simmer, error) {
	server.Lock()
	defer server.Unlock()
	if server.pool == nil {
		root, err := Root()
		if err != nil {
			return nil, nil, err
		}
		pool, err := LoadPool(root)
		if err != nil {
			return nil, nil, err
		}
		sim, err := NewSimmer(pool, "default")
		if err != nil {
			return nil, nil, err
		}
		server.pool, server.sim = pool, sim
	}
	return server.pool, server.sim, nil
}

// SearchAsync runs a search for the native server's /compSearchAsync, streaming the phase and the best
// comp so far every few seconds (the server drops a call that goes 10 minutes without progress), and
// stops on Abort.
func SearchAsync(req *proto.CompSearchRequest, progress chan *proto.ProgressMetrics, requestId string) {
	fail := func(err error) {
		progress <- &proto.ProgressMetrics{FinalCompSearchResult: &proto.CompSearchResult{Error: &proto.ErrorOutcome{Message: err.Error()}}}
	}
	signals, err := simsignals.RegisterWithId(requestId)
	if err != nil {
		fail(err)
		return
	}
	go func() {
		defer simsignals.UnregisterId(requestId)
		pool, sim, err := serverSearch()
		if err != nil {
			fail(err)
			return
		}
		var mu sync.Mutex
		latest := &proto.ProgressMetrics{CompSearchPhase: "starting"}
		done := make(chan struct{})
		var ticker sync.WaitGroup
		ticker.Add(1)
		go func() {
			defer ticker.Done()
			t := time.NewTicker(3 * time.Second)
			defer t.Stop()
			for {
				select {
				case <-done:
					return
				case <-t.C:
					mu.Lock()
					p := googleProto.Clone(latest).(*proto.ProgressMetrics)
					mu.Unlock()
					p.CompletedSims = int32(sim.Run.Load())
					progress <- p
				}
			}
		}()
		res, err := SearchWithAbort(pool, sim, req, func(phase string, best *Comp, sc *Scenario) {
			p := &proto.ProgressMetrics{CompSearchPhase: phase}
			if best != nil {
				p.CompSearchBest = sc.compProto(best, best, nil)
			}
			mu.Lock()
			latest = p
			mu.Unlock()
		}, signals.Abort.IsTriggered)
		close(done)
		ticker.Wait()
		if err != nil {
			fail(err)
			return
		}
		progress <- &proto.ProgressMetrics{FinalCompSearchResult: res}
	}()
}
