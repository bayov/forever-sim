package compsearch

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// Fight is the encounter a run's single-player sims use.
type Fight struct {
	Duration          float64
	DurationVariation float64
	MobType           proto.MobType
	FlaskZone         bool
	RaidHits          float64 // raid damage hits a minute, when SetRaidHits
	SetRaidHits       bool
	NoFlask           bool
}

// apply sets the fight on a request's encounter, and strips the flask for a no-flask scenario.
func (f Fight) apply(req *proto.RaidSimRequest) {
	enc := req.Encounter
	enc.Duration = f.Duration
	enc.DurationVariation = f.DurationVariation
	enc.NaturalFlaskZone = f.FlaskZone
	if f.SetRaidHits {
		enc.RaidDamageHitsPerMinute = f.RaidHits
	}
	if f.MobType != proto.MobType_MobTypeUnknown {
		for _, t := range enc.Targets {
			t.MobType = f.MobType
		}
	}
	if f.NoFlask {
		for _, party := range req.Raid.Parties {
			for _, p := range party.Players {
				if p.Consumes != nil {
					p.Consumes.Flask = proto.Flask_FlaskUnknown
				}
			}
		}
	}
}

// Value is one single-player sim's result (docs/raid_sim.md, The cache: the value).
type Value struct {
	DPS     float64 `json:"dps"`
	SE      float64 `json:"se"`
	Stdev   float64 `json:"stdev"`
	TPS     float64 `json:"tps"`
	OOMSecs float64 `json:"oom"`
	Iters   int32   `json:"iters"`
}

// Job is one single-player sim: a build at a faction's race, in a seat, in a fight.
type Job struct {
	Build      *Build
	Faction    proto.Faction
	Seat       Seat
	Fight      Fight
	Iterations int32
	Seed       int64

	Value *Value
	key   string
	req   *proto.RaidSimRequest
}

// Simmer runs single-player sims on every core and caches them by their normalized request
// (docs/raid_sim.md, The cache).
type Simmer struct {
	Pool    *Pool
	Dir     string // the disk cache, "" for memory only
	Workers int
	Labeled bool // SimOptions.use_labeled_rands

	version string
	mu      sync.Mutex
	mem     map[string]*Value
	Run     atomic.Int64 // sims run
	Hits    atomic.Int64 // cache hits
	OnSim   func(done, total int)
	touched atomic.Int64 // when the cache directory was last marked used, Unix seconds
}

// NewSimmer makes a sim runner. dir "" keeps the cache in memory only; "default" uses
// forever-sim/compsearch under the user's cache directory.
func NewSimmer(pool *Pool, dir string) (*Simmer, error) {
	version, err := Version()
	if err != nil {
		return nil, err
	}
	return newSimmerAt(pool, dir, version)
}

// newSimmerAt is NewSimmer for a given version (a test's stand-in for another binary).
func newSimmerAt(pool *Pool, dir, version string) (*Simmer, error) {
	if dir == "default" {
		base, err := os.UserCacheDir()
		if err != nil {
			return nil, err
		}
		dir = filepath.Join(base, "forever-sim", "compsearch")
	}
	s := &Simmer{Pool: pool, Workers: runtime.GOMAXPROCS(0), Labeled: true, version: version, mem: map[string]*Value{}}
	if dir != "" {
		s.Dir = filepath.Join(dir, version[:16])
		if err := os.MkdirAll(s.Dir, 0o755); err != nil {
			return nil, err
		}
		pruneOldVersions(dir, version[:16])
	}
	return s, nil
}

var versionOnce struct {
	sync.Once
	v   string
	err error
}

// Version is the SHA-256 of the running binary, which holds the sim and the item database: a code or
// database change starts a fresh cache. A git hash would miss an uncommitted edit. It reads the binary
// through /proc/self/exe where there is one, which stays the running binary when the file is rebuilt in
// place, and a long-running program calls it at startup (the server, the command line), so a binary
// replaced since never lends its hash to old code's results.
func Version() (string, error) {
	versionOnce.Do(func() {
		exe := "/proc/self/exe"
		if _, err := os.Stat(exe); err != nil {
			if exe, err = os.Executable(); err != nil {
				versionOnce.err = err
				return
			}
		}
		f, err := os.Open(exe)
		if err != nil {
			versionOnce.err = err
			return
		}
		defer f.Close()
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			versionOnce.err = err
			return
		}
		versionOnce.v = hex.EncodeToString(h.Sum(nil))
	})
	return versionOnce.v, versionOnce.err
}

// pruneOldVersions removes the caches of other binaries unused for a week. Two binaries in use at once
// (the CLI and the native server, or two checkouts) share the directory, so a cache that's in use is
// kept: each run marks its own as used.
func pruneOldVersions(dir, keep string) {
	now := time.Now()
	_ = os.Chtimes(filepath.Join(dir, keep), now, now)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() || e.Name() == keep || len(e.Name()) != 16 {
			continue
		}
		if info, err := e.Info(); err == nil && now.Sub(info.ModTime()) > 7*24*time.Hour {
			os.RemoveAll(filepath.Join(dir, e.Name()))
		}
	}
}

// Request is the job's single-player request, resolved and normalized.
func (s *Simmer) Request(j *Job) (*proto.RaidSimRequest, error) {
	if j.req != nil {
		return j.req, nil
	}
	b := j.Build
	req := googleProto.Clone(b.Req).(*proto.RaidSimRequest)
	player, err := s.Pool.Player(b, j.Faction)
	if err != nil {
		return nil, err
	}
	req.Raid.Parties[0].Players[0] = player
	j.Seat.Write(req, b)
	j.Fight.apply(req)
	req.SimOptions = &proto.SimOptions{Iterations: j.Iterations, RandomSeed: j.Seed, UseLabeledRands: s.Labeled}
	normalize(req, b)
	j.req = req
	return req, nil
}

func (s *Simmer) key(j *Job) (string, error) {
	if j.key != "" {
		return j.key, nil
	}
	req, err := s.Request(j)
	if err != nil {
		return "", err
	}
	data, err := googleProto.MarshalOptions{Deterministic: true}.Marshal(req)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(append(data, []byte(s.version)...))
	j.key = hex.EncodeToString(h[:])
	return j.key, nil
}

func (s *Simmer) lookup(key string) *Value {
	s.mu.Lock()
	v, ok := s.mem[key]
	s.mu.Unlock()
	if ok {
		return v
	}
	if s.Dir == "" {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(s.Dir, key[:2], key+".json"))
	if err != nil {
		return nil
	}
	v = &Value{}
	if json.Unmarshal(data, v) != nil {
		return nil
	}
	s.mu.Lock()
	s.mem[key] = v
	s.mu.Unlock()
	return v
}

func (s *Simmer) store(key string, v *Value) {
	s.mu.Lock()
	s.mem[key] = v
	s.mu.Unlock()
	if s.Dir == "" {
		return
	}
	data, _ := json.Marshal(v)
	dir := filepath.Join(s.Dir, key[:2])
	_ = os.MkdirAll(dir, 0o755)
	// A temp file of its own, renamed into place, so a reader never sees half a value. A failed write
	// only costs the cache entry.
	f, err := os.CreateTemp(dir, "."+key+".*")
	if err != nil {
		return
	}
	_, err = f.Write(data)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil || os.Rename(f.Name(), filepath.Join(dir, key+".json")) != nil {
		_ = os.Remove(f.Name())
	}
	// Mark the cache used, at most hourly, so a long-running server's isn't pruned as idle.
	if now := time.Now().Unix(); now-s.touched.Load() > 3600 {
		s.touched.Store(now)
		_ = os.Chtimes(s.Dir, time.Now(), time.Now())
	}
}

// Sim runs the jobs, each at most once per distinct request, on every core.
func (s *Simmer) Sim(jobs []*Job) error { return s.SimAbortable(jobs, nil) }

// SimAbortable is Sim that starts no more sims once abort says so.
func (s *Simmer) SimAbortable(jobs []*Job, abort func() bool) error {
	return s.simCounted(jobs, abort, nil)
}

// simCounted is SimAbortable that also counts into a search's own counter. It starts no more sims after
// one fails.
func (s *Simmer) simCounted(jobs []*Job, abort func() bool, c *Counter) error {
	todo := map[string][]*Job{}
	var order []string
	for _, j := range jobs {
		if j.Value != nil {
			continue
		}
		key, err := s.key(j)
		if err != nil {
			return err
		}
		if v := s.lookup(key); v != nil {
			j.Value, j.req = v, nil
			s.Hits.Add(1)
			if c != nil {
				c.Hits.Add(1)
			}
			continue
		}
		if _, ok := todo[key]; !ok {
			order = append(order, key)
		}
		todo[key] = append(todo[key], j)
	}
	if len(order) == 0 {
		return nil
	}
	var wg sync.WaitGroup
	var firstErr atomic.Value
	var done atomic.Int64
	sem := make(chan struct{}, max(1, s.Workers))
	aborted := false
	for _, key := range order {
		key := key
		js := todo[key]
		if abort != nil && abort() {
			aborted = true
			break
		}
		if firstErr.Load() != nil {
			break
		}
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			v, err := runOne(js[0].req)
			if err != nil {
				firstErr.CompareAndSwap(nil, fmt.Errorf("%s: %w", js[0].Build.ID, err))
				return
			}
			s.store(key, v)
			// The request is about 37 KB and a search keeps its jobs, so it goes once the key is known.
			for _, j := range js {
				j.Value, j.req = v, nil
			}
			s.Run.Add(1)
			if c != nil {
				c.Run.Add(1)
			}
			if s.OnSim != nil {
				s.OnSim(int(done.Add(1)), len(order))
			}
		}()
	}
	wg.Wait()
	if err, ok := firstErr.Load().(error); ok && err != nil {
		return err
	}
	if aborted {
		return ErrAborted
	}
	return nil
}

func runOne(req *proto.RaidSimRequest) (*Value, error) {
	res := core.RunRaidSim(req)
	if res.Error != nil {
		return nil, fmt.Errorf("%s", res.Error.Message)
	}
	m := res.RaidMetrics
	iters := req.SimOptions.Iterations
	v := &Value{DPS: m.Dps.Avg, Stdev: m.Dps.Stdev, SE: m.Dps.Stdev / math.Sqrt(float64(iters)), Iters: iters}
	if len(m.Parties) > 0 && len(m.Parties[0].Players) > 0 {
		pm := m.Parties[0].Players[0]
		if pm.Threat != nil {
			v.TPS = pm.Threat.Avg
		}
		v.OOMSecs = pm.SecondsOomAvg
	}
	return v, nil
}
