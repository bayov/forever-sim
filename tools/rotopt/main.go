// rotopt tunes a rotation template against the sim.
//
// A template (tools/rotopt/rogue_ss.go for example) fixes the shape of a rotation and
// leaves its thresholds as knobs. rotopt runs the sim over the knob values for each fight
// length and writes the best rotation out as a preset.
//
//	go run --tags=with_db ./tools/rotopt -template rogue_ss -settings ss.json -durs 120,300,480 -out ui/rogue/apls
//
// The settings file is a RaidSimRequest or the JSON the UI exports under Export > JSON.
// tools/rotopt/mksettings builds one from the shipped gear sets, rotations and the UI's
// default raid, so no export is needed.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/wowsims/classic/sim"
	_ "github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

func main() {
	templateName := flag.String("template", "", "rotation template, one of: "+strings.Join(templateNames(), ", "))
	settingsPath := flag.String("settings", "", "RaidSimRequest or IndividualSimSettings JSON")
	aplPath := flag.String("apl", "", "a rotation JSON to sim as is instead of a template (no knobs, use with -eval or -timeline)")
	durs := flag.String("durs", "120,300,480", "fight lengths in seconds, one search per length")
	combined := flag.Bool("combined", false, "one search over the mean of all fight lengths instead of one per length")
	iterations := flag.Int("iters", 3000, "sim iterations per candidate")
	confirmIterations := flag.Int("confirm-iters", 10000, "sim iterations for the final numbers")
	confidence := flag.Float64("confidence", 2, "standard errors of gain a knob change must show to be kept")
	knobsFlag := flag.String("knobs", "", "starting knob values, name=value,...")
	race := flag.String("race", "", "override the player's race, e.g. Gnome")
	level := flag.Int("level", 0, "override the player's level, e.g. 20")
	targets := flag.Int("targets", 0, "override the number of enemies, copies of the encounter's first target, for AoE rotations")
	bonusTalents := flag.Int("bonus-talents", -1, "override the player's extra talent points beyond what the level grants, e.g. 5 for the Forever beta's 16 points at level 20")
	bonus := flag.String("bonus", "", "extra stats on top of the gear, to see how the answers move with better itemization, e.g. AttackPower=400,Stamina=100")
	printStats := flag.Bool("print-stats", false, "print the player's final stats and exit")
	weights := flag.Bool("weights", false, "print stat weights for the starting knobs at the first fight length and exit")
	touchICD := flag.Duration("touch-icd", core.TouchOfTheGraveICD, "internal cooldown for the Undead Touch of the Grave racial (the tooltip has none, 15s is assumed), 0 for none")
	talents := flag.String("talents", "", "override the player's talent string")
	talentSearch := flag.Bool("talent-search", false, "search the talent tree with the starting knobs instead of the knobs (prints the best talent string)")
	talentsFile := flag.String("talents-file", "", "score every talent string in this file (one per line) with the starting knobs, best first")
	requiredTalents := flag.String("require", "", "talents the talent search keeps at max rank, by field name, e.g. spiritWeapons for the threat cut the sim does not score")
	gearSearch := flag.Bool("gear-search", false, "search every slot over the items the player can equip at their level, with the starting knobs (prints the gear and writes it to -gear-out)")
	gearOut := flag.String("gear-out", "", "settings JSON to write with the gear the search found, for the next run")
	gearQuality := flag.Int("gear-quality", 2, "lowest item quality the gear search considers (2 uncommon, 3 rare)")
	gearExclude := flag.String("gear-exclude", "", "item IDs the gear search leaves out, comma separated (Manual Crowd Pummeler and its three charges, say)")
	gearPhase := flag.Int("gear-phase", 0, "highest raid phase the gear search considers (1 MC, 2 DM, 3 BWL, 4 ZG, 5 AQ, 6 Naxx), 0 for all")
	dbPath := flag.String("db", "assets/database/db.json", "UI item database for the gear search")
	itemLevelsPath := flag.String("item-levels", "assets/db_inputs/wago_db2_items.csv", "wago item export, for which items need a PvP rank")
	evalOnly := flag.Bool("eval", false, "only run the starting knobs, no search")
	outDir := flag.String("out", "", "directory to write <name>[_<dur>s].apl.json into")
	outName := flag.String("name", "", "file name base for -out, defaults to the template name")
	timeline := flag.Bool("timeline", false, "print one iteration's cooldown and finisher casts for the starting knobs at the first fight length, no search")
	flag.Parse()

	log.SetOutput(io.Discard)
	sim.RegisterAll()
	core.TouchOfTheGraveICD = *touchICD

	template, ok := templates[*templateName]
	if *aplPath != "" {
		rot, err := loadRotation(*aplPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "loading rotation: %v\n", err)
			os.Exit(2)
		}
		template, ok = fixedRotation{rot}, true
		*templateName = strings.TrimSuffix(filepath.Base(*aplPath), ".apl.json")
	}
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown template %q, want one of %s\n", *templateName, strings.Join(templateNames(), ", "))
		os.Exit(2)
	}
	setup, err := loadSetup(*settingsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loading settings: %v\n", err)
		os.Exit(2)
	}
	if *race != "" {
		id, ok := proto.Race_value["Race"+*race]
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown race %q (SkyborneHighOrder, SkyborneWindshaper, Orc, ...)\n", *race)
			os.Exit(2)
		}
		setup.player().Race = proto.Race(id)
	}
	if *level != 0 {
		setup.player().Level = int32(*level)
	}
	if *bonusTalents >= 0 {
		setup.player().BonusTalentPoints = int32(*bonusTalents)
	}
	if l := setup.player().Level; l > 0 && l < core.CharacterMaxLevel {
		maxTalentPoints = max(0, int(l)-9)
		playerLevel = l
	}
	// The bonus points never take a character past the 51 a level 60 has.
	maxTalentPoints = min(51, maxTalentPoints+int(setup.player().BonusTalentPoints))
	if *talents != "" {
		setup.player().TalentsString = *talents
	}
	if *targets > 0 {
		setup.setTargets(*targets)
	}
	if *bonus != "" {
		p := setup.player()
		if p.BonusStats == nil {
			p.BonusStats = &proto.UnitStats{}
		}
		for _, kv := range strings.Split(*bonus, ",") {
			name, val, _ := strings.Cut(kv, "=")
			stat, ok := proto.Stat_value["Stat"+name]
			if !ok {
				fmt.Fprintf(os.Stderr, "unknown stat %q\n", name)
				os.Exit(2)
			}
			for len(p.BonusStats.Stats) <= int(stat) {
				p.BonusStats.Stats = append(p.BonusStats.Stats, 0)
			}
			v, _ := strconv.ParseFloat(val, 64)
			p.BonusStats.Stats[stat] += v
		}
	}

	start := defaultKnobs(template)
	if *knobsFlag != "" {
		for _, kv := range strings.Split(*knobsFlag, ",") {
			name, val, _ := strings.Cut(kv, "=")
			if _, ok := start[name]; !ok {
				fmt.Fprintf(os.Stderr, "unknown knob %q\n", name)
				os.Exit(2)
			}
			start[name], _ = strconv.ParseFloat(val, 64)
		}
	}

	var durations []float64
	for _, d := range strings.Split(*durs, ",") {
		v, _ := strconv.ParseFloat(d, 64)
		durations = append(durations, v)
	}

	if *printStats {
		fmt.Print(setup.finalStats(template.Build(start)))
		return
	}

	if *weights {
		fmt.Print(setup.statWeights(template.Build(start), durations[0], int32(*confirmIterations)))
		return
	}

	if *timeline {
		fmt.Print(setup.timeline(template.Build(start), durations[0]))
		return
	}

	if *gearSearch {
		level := int32(core.CharacterMaxLevel)
		if l := setup.player().Level; l > 0 {
			level = l
		}
		var exclude []int32
		for _, id := range strings.Split(*gearExclude, ",") {
			if v, err := strconv.Atoi(id); err == nil {
				exclude = append(exclude, int32(v))
			}
		}
		// An excluded item the player is wearing comes off first, or the search would
		// keep it since nothing in the pool is compared against it.
		if equipment := setup.player().Equipment; equipment != nil {
			for _, item := range equipment.Items {
				if item != nil && slices.Contains(exclude, item.Id) {
					item.Reset()
				}
			}
		}
		pool, err := loadGearPool(*dbPath, *itemLevelsPath, setup.player(), level, proto.ItemQuality(*gearQuality), int32(*gearPhase), exclude)
		if err != nil {
			fmt.Fprintf(os.Stderr, "loading items: %v\n", err)
			os.Exit(2)
		}
		for slot, options := range pool {
			fmt.Printf("%s: %d options\n", gearSlotNames[slot], len(options))
		}
		s := &searcher{
			setup: setup, template: template, durations: durations,
			iterations: int32(*iterations), confidence: *confidence, cache: map[string]result{},
		}
		s.searchGear(pool, start)
		confirm := &searcher{
			setup: setup, template: template, durations: durations,
			iterations: int32(*confirmIterations), cache: map[string]result{},
		}
		fmt.Printf("%gs: %s\n  sim runs %d\n", durations, confirm.score(start), s.evals)
		if *gearOut != "" {
			if err := setup.save(*gearOut); err != nil {
				fmt.Fprintf(os.Stderr, "writing %s: %v\n", *gearOut, err)
				os.Exit(1)
			}
			fmt.Printf("  wrote %s\n", *gearOut)
		}
		return
	}

	if *talentsFile != "" {
		data, err := os.ReadFile(*talentsFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "reading talents: %v\n", err)
			os.Exit(2)
		}
		s := &searcher{
			setup: setup, template: template, durations: durations,
			iterations: int32(*iterations), cache: map[string]result{},
		}
		type scored struct {
			talents string
			result  result
		}
		class := strings.ToLower(strings.TrimPrefix(setup.player().Class.String(), "Class"))
		trees, err := loadTalentTrees(class)
		if err != nil {
			fmt.Fprintf(os.Stderr, "loading talent trees: %v\n", err)
			os.Exit(2)
		}
		var results []scored
		for _, line := range strings.Split(string(data), "\n") {
			if line = strings.TrimSpace(line); line == "" {
				continue
			}
			// A hand written build over the point cap or past a row gate would still sim,
			// and win, so it is refused here instead.
			if !parseBuild(trees, line).valid(trees) {
				fmt.Printf("invalid %s\n", line)
				continue
			}
			setup.talents = line
			results = append(results, scored{line, s.score(start)})
		}
		sort.Slice(results, func(i, j int) bool { return results[i].result.dps > results[j].result.dps })
		for _, r := range results {
			fmt.Printf("%s %s\n", r.result, r.talents)
		}
		return
	}

	if *talentSearch {
		class := strings.ToLower(strings.TrimPrefix(setup.player().Class.String(), "Class"))
		trees, err := loadTalentTrees(class)
		if err != nil {
			fmt.Fprintf(os.Stderr, "loading talent trees: %v\n", err)
			os.Exit(2)
		}
		if err := requireTalents(trees, *requiredTalents); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}
		s := &searcher{
			setup: setup, template: template, durations: durations,
			iterations: int32(*iterations), confidence: *confidence, cache: map[string]result{},
		}
		best, _ := s.searchTalents(trees, setup.player().TalentsString, start)
		confirm := &searcher{
			setup: setup, template: template, durations: durations,
			iterations: int32(*confirmIterations), cache: map[string]result{},
		}
		setup.talents = best
		fmt.Printf("%gs: %s\n  talents %s\n  sim runs %d\n", durations, confirm.score(start), best, s.evals)
		return
	}

	// Each search gets its own fight lengths and output name.
	type job struct {
		durations []float64
		suffix    string
	}
	var jobs []job
	if *combined {
		jobs = []job{{durations, ""}}
	} else {
		for _, d := range durations {
			jobs = append(jobs, job{[]float64{d}, fmt.Sprintf("_%gs", d)})
		}
	}

	for _, j := range jobs {
		s := &searcher{
			setup: setup, template: template, durations: j.durations,
			iterations: int32(*iterations), confidence: *confidence, cache: map[string]result{},
		}
		best := start
		if !*evalOnly {
			best, _ = s.search(start)
		}

		confirm := &searcher{
			setup: setup, template: template, durations: j.durations,
			iterations: int32(*confirmIterations), cache: map[string]result{},
		}
		fmt.Printf("%gs: %s\n", j.durations, confirm.score(best))
		fmt.Printf("  start %s\n  best  %s\n  sim runs %d\n", confirm.score(start), best, s.evals)

		if *outDir != "" {
			name := *outName
			if name == "" {
				name = *templateName
			}
			path := filepath.Join(*outDir, name+j.suffix+".apl.json")
			if err := os.WriteFile(path, []byte(rotationJSON(template.Build(best))), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "writing %s: %v\n", path, err)
				os.Exit(1)
			}
			fmt.Printf("  wrote %s\n", path)
		}
	}
}

func templateNames() []string {
	var names []string
	for name := range templates {
		names = append(names, name)
	}
	return names
}
