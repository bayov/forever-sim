// rotopt tunes a rotation template against the sim.
//
// A template (tools/rotopt/rogue_ss.go for example) fixes the shape of a rotation and
// leaves its thresholds as knobs. rotopt runs the sim over the knob values for each fight
// length and writes the best rotation out as a preset.
//
//	go run --tags=with_db ./tools/rotopt -template rogue_ss -settings ss.json -durs 120,300,480 -out ui/rogue/apls
//
// The settings file is a RaidSimRequest or the JSON the UI exports under Export > JSON.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wowsims/classic/sim"
	_ "github.com/wowsims/classic/sim/common"
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
	evalOnly := flag.Bool("eval", false, "only run the starting knobs, no search")
	outDir := flag.String("out", "", "directory to write <name>[_<dur>s].apl.json into")
	outName := flag.String("name", "", "file name base for -out, defaults to the template name")
	timeline := flag.Bool("timeline", false, "print one iteration's cooldown and finisher casts for the starting knobs at the first fight length, no search")
	flag.Parse()

	log.SetOutput(io.Discard)
	sim.RegisterAll()

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
		setup.player().Race = proto.Race(proto.Race_value["Race"+*race])
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

	if *timeline {
		fmt.Print(setup.timeline(template.Build(start), durations[0]))
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
