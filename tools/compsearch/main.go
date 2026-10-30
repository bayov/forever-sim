// compsearch is the raid sim's comp search on the command line (docs/raid_sim.md, Where it runs).
//
//	go run -tags with_db ./tools/compsearch -size 20 -out result      # result.json and result.md
//	go run -tags with_db ./tools/compsearch -drops                    # rewrite sim/compsearch/drops.json
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/wowsims/classic/sim/compsearch"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	root := flag.String("root", ".", "the forever-sim checkout")
	drops := flag.Bool("drops", false, "recompute the cache's drop table, sim/compsearch/drops.json")
	size := flag.Int("size", 20, "raid size: 10, 20 or 40")
	tanks := flag.Int("tanks", 0, "tank slots (default by size)")
	healers := flag.Int("healers", 0, "healer slots (default by size)")
	faction := flag.String("faction", "", "Horde or Alliance (default both)")
	camps := flag.String("camps", "", "on, off or both (default both)")
	iters := flag.Int("iterations", 1000, "the screening sims' iterations")
	contender := flag.Int("contender-iterations", 10000, "the contenders' re-sims")
	seed := flag.Int64("seed", 101, "random seed")
	duration := flag.Float64("duration", 180, "fight length, seconds")
	variation := flag.Float64("variation", 20, "fight length variation, seconds")
	boss := flag.String("boss", "Generic", "Generic or Onyxia (Dragonkin, outside the flask zone)")
	health := flag.Float64("health", 0, "a health fight: the boss's health (the length is then a first guess)")
	leftOut := flag.Bool("measure-left-out", false, "sim every base build with and without the buffs the resolver leaves out")
	fixedDPS := flag.String("fixed-dps", "", "comma-separated report build ids that must be in the comp")
	fixedTanks := flag.String("fixed-tanks", "", "comma-separated tank presets, e.g. ProtectionWarrior")
	fixedHealers := flag.String("fixed-healers", "", "comma-separated healer presets, e.g. RestorationShaman")
	maxCombos := flag.Int("max-combos", 0, "rank at most this many tank and healer combos (0: all)")
	noSens := flag.Bool("no-sensitivity", false, "leave out the sensitivity scenarios")
	confirm := flag.Bool("confirm", false, "confirm the top sets with raid sims (milestone M1b)")
	matrix := flag.Bool("matrix", false, "re-run the search under each sensitivity scenario (lengths, bosses, raid damage, no flask, the camp rules)")
	cache := flag.String("cache", "default", `the sim cache directory ("" for memory only)`)
	out := flag.String("out", "", "write <out>.json and <out>.md (default: markdown to stdout)")
	quiet := flag.Bool("quiet", false, "no progress lines")
	flag.Parse()

	pool, err := compsearch.LoadPool(*root)
	if err != nil {
		fail(err)
	}
	if *drops {
		table, err := compsearch.ComputeDrops(pool, 60, 7, func(s string) { fmt.Fprintln(os.Stderr, s) })
		if err != nil {
			fail(err)
		}
		data, err := compsearch.WriteDrops(table)
		if err != nil {
			fail(err)
		}
		path := filepath.Join(*root, "sim", "compsearch", "drops.json")
		if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
			fail(err)
		}
		fmt.Printf("wrote %s for %d builds on %d cores\n", path, len(table), runtime.GOMAXPROCS(0))
		return
	}

	if *leftOut {
		if err := compsearch.MeasureLeftOut(pool, os.Stdout); err != nil {
			fail(err)
		}
		return
	}

	req := &proto.CompSearchRequest{Size: int32(*size), Tanks: int32(*tanks), Healers: int32(*healers),
		Iterations: int32(*iters), ContenderIterations: int32(*contender), Seed: *seed, MaxCombos: int32(*maxCombos),
		SkipSensitivity: *noSens, Confirm: *confirm,
		Encounter: &proto.CompEncounter{Name: *boss, Duration: *duration, DurationVariation: *variation}}
	if *health > 0 {
		req.Encounter.UseHealth = true
		req.Encounter.BossHealth = *health
	}
	if *boss == "Onyxia" {
		req.Encounter.MobType = proto.MobType_MobTypeDragonkin
	} else {
		req.Encounter.NaturalFlaskZone = true
	}
	switch strings.ToLower(*faction) {
	case "horde":
		req.Factions = []proto.Faction{proto.Faction_Horde}
	case "alliance":
		req.Factions = []proto.Faction{proto.Faction_Alliance}
	}
	switch strings.ToLower(*camps) {
	case "on":
		req.CampBuffs = []bool{true}
	case "off":
		req.CampBuffs = []bool{false}
	}
	for _, id := range split(*fixedDPS) {
		req.FixedDps = append(req.FixedDps, id)
	}
	for _, s := range split(*fixedTanks) {
		req.FixedTanks = append(req.FixedTanks, preset(s))
	}
	for _, s := range split(*fixedHealers) {
		req.FixedHealers = append(req.FixedHealers, preset(s))
	}

	simmer, err := compsearch.NewSimmer(pool, *cache)
	if err != nil {
		fail(err)
	}
	start := time.Now()
	last := time.Time{}
	progress := func(phase string, best *compsearch.Comp, sc *compsearch.Scenario) {
		if *quiet || time.Since(last) < 2*time.Second && phase != "done" {
			return
		}
		last = time.Now()
		b := ""
		if best != nil {
			b = fmt.Sprintf(", best %.1f", best.Score)
		}
		fmt.Fprintf(os.Stderr, "[%5.0fs] %s %v: %s%s (%d sims, %d cached)\n", time.Since(start).Seconds(), sc.Faction,
			map[bool]string{true: "camps", false: "no camps"}[sc.Camps], phase, b, simmer.Run.Load(), simmer.Hits.Load())
	}
	res, err := compsearch.Search(pool, simmer, req, progress)
	if err != nil {
		fail(err)
	}
	if *matrix {
		if err := compsearch.RunMatrix(pool, simmer, req, res, progress); err != nil {
			fail(err)
		}
	}
	md := compsearch.Markdown(res)
	if *out == "" {
		fmt.Print(md)
		return
	}
	data, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(res)
	if err != nil {
		fail(err)
	}
	if err := os.WriteFile(*out+".json", data, 0o644); err != nil {
		fail(err)
	}
	if err := os.WriteFile(*out+".md", []byte(md), 0o644); err != nil {
		fail(err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s.json and %s.md\n", *out, *out)
}

func split(s string) []string {
	var out []string
	for _, x := range strings.Split(s, ",") {
		if x = strings.TrimSpace(x); x != "" {
			out = append(out, x)
		}
	}
	return out
}

func preset(s string) proto.RaidPreset {
	v, ok := proto.RaidPreset_value["RaidPreset"+s]
	if !ok {
		fail(fmt.Errorf("unknown preset %q", s))
	}
	return proto.RaidPreset(v)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "compsearch:", err)
	os.Exit(1)
}
