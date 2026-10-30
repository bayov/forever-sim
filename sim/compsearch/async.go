package compsearch

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
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
	failed := func(err error) *proto.ProgressMetrics {
		return &proto.ProgressMetrics{FinalCompSearchResult: &proto.CompSearchResult{Error: &proto.ErrorOutcome{Message: err.Error()}}}
	}
	signals, err := simsignals.RegisterWithId(requestId)
	if err != nil {
		progress <- failed(err)
		return
	}
	go func() {
		// The deferred calls run last first: a panic becomes the final frame, the ticker stops, then the
		// final frame goes out, so nothing is left running whatever happened.
		var final *proto.ProgressMetrics
		defer func() { progress <- final }()
		defer simsignals.UnregisterId(requestId)
		done := make(chan struct{})
		var ticker sync.WaitGroup
		defer func() {
			close(done)
			ticker.Wait()
		}()
		defer func() {
			if r := recover(); r != nil {
				final = failed(fmt.Errorf("the search failed: %v\n%s", r, debug.Stack()))
			}
		}()
		pool, sim, err := serverSearch()
		if err != nil {
			final = failed(err)
			return
		}
		var mu sync.Mutex
		latest := &proto.ProgressMetrics{CompSearchPhase: "starting"}
		var simsDone func() int64
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
					if simsDone != nil {
						p.CompletedSims = int32(simsDone())
					}
					mu.Unlock()
					// A reader that stopped reading misses ticks rather than holding the search up.
					select {
					case progress <- p:
					default:
					}
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
			simsDone = sc.Counter.Done
			mu.Unlock()
		}, signals.Abort.IsTriggered)
		if err != nil {
			final = failed(err)
			return
		}
		final = &proto.ProgressMetrics{FinalCompSearchResult: res}
	}()
}
