// levelingplan runs the Leveling Plan tab's talent search natively, on all the cores
// (docs/leveling_plan.md, Talent search). The browser's sims run in wasm workers and were too
// slow for a plan from 10 to 60.
//
// The UI's Search button sends the plan to a local server:
//
//	go run --tags=with_db ./tools/levelingplan -serve localhost:3001
//
// It answers with a line of progress now and then and the result at the end. The same search
// runs from a file of the request, for a run without the UI:
//
//	go run --tags=with_db ./tools/levelingplan -in request.json -out result.json
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/wowsims/classic/assets/database"
	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	serve := flag.String("serve", "", "address to serve the UI's searches on, like localhost:3001")
	in := flag.String("in", "", "a LevelingPlanSearchRequest JSON to search, without the UI")
	out := flag.String("out", "", "with -in, where to write the LevelingPlanSearchResult JSON (stdout when empty)")
	keep := flag.String("keep", "", "with -serve, a file to write each request to, to run it again with -in")
	levelRequests := flag.String("level-requests", "", "with -in, write each level's sim request (the plan's gear, talents and rotation there) into this directory as level_NN.json and don't search, for rotopt -settings")
	training := flag.Bool("training", false, "with -in, measure what each spell rank the plan casts is worth and make the training schedule under the gold budget (training.go), writing LevelingTrainingSettings to -out")
	trainingSessions := flag.Int("training-sessions", 400, "with -training, sessions per sim of a level")
	goldAt40 := flag.Float64("gold-at-40", 115, "with -training, the gold earned by level 40")
	ridingLevel := flag.Int("riding-level", 43, "with -training, the level we buy riding at, 0 for none")
	ridingGold := flag.Float64("riding-gold", 90, "with -training, what riding costs us, in gold")
	alwaysSpells := flag.String("always", "", "with -training, the spells the plan doesn't cast that we train anyway, comma separated, like \"Ghost Wolf,Astral Recall\" (the request's plan's when empty)")
	flag.Parse()

	if !core.WITH_DB {
		fmt.Fprintln(os.Stderr, "build with --tags=with_db, the search needs the item database")
		os.Exit(2)
	}
	log.SetFlags(log.Ltime)
	sim.RegisterAll()
	db := database.Load()

	switch {
	case *serve != "":
		log.Printf("serving leveling plan searches on http://%s/search", *serve)
		http.HandleFunc("/search", searchHandler(db, *keep))
		if err := http.ListenAndServe(*serve, nil); err != nil {
			log.Fatal(err)
		}
	case *in != "" && *training:
		var always []string
		if *alwaysSpells != "" {
			always = strings.Split(*alwaysSpells, ",")
		}
		if err := trainingFile(db, *in, *out, int32(*trainingSessions), *goldAt40, int32(*ridingLevel), *ridingGold, always); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *in != "" && *levelRequests != "":
		if err := writeLevelRequests(db, *in, *levelRequests); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *in != "":
		if err := searchFile(db, *in, *out); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		flag.Usage()
		os.Exit(2)
	}
}

func readRequest(data []byte) (*proto.LevelingPlanSearchRequest, error) {
	req := &proto.LevelingPlanSearchRequest{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, req); err != nil {
		return nil, fmt.Errorf("reading the request: %w", err)
	}
	return req, nil
}

func searchFile(db *proto.UIDatabase, in, out string) error {
	data, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	req, err := readRequest(data)
	if err != nil {
		return err
	}
	p, err := newPlanner(req, db)
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	var last time.Time
	result := runSearch(ctx, req, p, func(update *proto.LevelingPlanSearchProgress) {
		if time.Since(last) > 2*time.Second {
			last = time.Now()
			fmt.Fprintf(os.Stderr, "%s, %d sims, best %.2fh\n", update.Stage, update.Sims, update.BestHours)
		}
	})
	log.Printf("done in %.0fs, %d sims: %s", result.Seconds, result.Sims, summary(result))
	data, err = protojson.MarshalOptions{Multiline: true}.Marshal(result)
	if err != nil {
		return err
	}
	if out == "" {
		_, err = os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(out, data, 0o644)
}

// writeLevelRequests writes the request the plan sims at each level, so rotopt can read its
// sessions (-session) or try another rotation there (-apl).
func writeLevelRequests(db *proto.UIDatabase, in, dir string) error {
	data, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	req, err := readRequest(data)
	if err != nil {
		return err
	}
	p, err := newPlanner(req, db)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// The plan's plain items aren't in the item database, so each request carries them.
	madeUp := &proto.SimDatabase{}
	for _, item := range p.gear.madeUp {
		sim, err := simItem(item)
		if err != nil {
			return err
		}
		madeUp.Items = append(madeUp.Items, sim)
	}
	for _, level := range planLevels(p.plan) {
		job := levelJob{level: level, talents: planTalentsAt(p.plan, level, p.shape)}
		request := p.request(job, 1000, levelSeed(p.seed, level))
		request.Raid.Parties[0].Players[0].Database = madeUp
		data, err := protojson.MarshalOptions{Multiline: true}.Marshal(request)
		if err != nil {
			return err
		}
		if err := os.WriteFile(fmt.Sprintf("%s/level_%02d.json", dir, level), data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// One search at a time, because each one takes all the cores.
var busy sync.Mutex

// searchHandler answers a search with a LevelingPlanSearchUpdate JSON a line: progress every
// half second, then the result. The UI's Stop closes the request, and the search stops.
func searchHandler(db *proto.UIDatabase, keep string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// The UI comes from the dev server on another port.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		if r.Method == http.MethodOptions {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "POST a LevelingPlanSearchRequest", http.StatusMethodNotAllowed)
			return
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		req, err := readRequest(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if keep != "" {
			if err := os.WriteFile(keep, data, 0o644); err != nil {
				log.Printf("keeping the request: %v", err)
			}
		}
		if !busy.TryLock() {
			http.Error(w, "Another search is running.", http.StatusConflict)
			return
		}
		defer busy.Unlock()

		w.Header().Set("Content-Type", "application/x-ndjson")
		flusher, _ := w.(http.Flusher)
		writer := bufio.NewWriter(w)
		var mu sync.Mutex
		send := func(update *proto.LevelingPlanSearchUpdate) {
			line, err := protojson.Marshal(update)
			if err != nil {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			writer.Write(line)
			writer.WriteByte('\n')
			writer.Flush()
			if flusher != nil {
				flusher.Flush()
			}
		}

		p, err := newPlanner(req, db)
		if err != nil {
			send(&proto.LevelingPlanSearchUpdate{Update: &proto.LevelingPlanSearchUpdate_Result{Result: &proto.LevelingPlanSearchResult{Error: err.Error()}}})
			return
		}
		log.Printf("searching levels %d to %d, %d sessions a sim", req.Plan.FromLevel, req.Plan.ToLevel, req.Sessions)
		var last time.Time
		result := runSearch(r.Context(), req, p, func(update *proto.LevelingPlanSearchProgress) {
			if time.Since(last) > 500*time.Millisecond {
				last = time.Now()
				send(&proto.LevelingPlanSearchUpdate{Update: &proto.LevelingPlanSearchUpdate_Progress{Progress: update}})
			}
		})
		log.Printf("done in %.0fs, %d sims: %s", result.Seconds, result.Sims, summary(result))
		send(&proto.LevelingPlanSearchUpdate{Update: &proto.LevelingPlanSearchUpdate_Result{Result: result}})
	}
}

func summary(result *proto.LevelingPlanSearchResult) string {
	if result.Error != "" {
		return result.Error
	}
	text := fmt.Sprintf("route %.2fh, plan %.2fh (±%.2fh)", result.RouteHours, result.PlanHours, result.Noise)
	if result.Passes > 0 {
		text += fmt.Sprintf(", plan's talents with the rotations found %.2fh (±%.2fh), %d passes", result.RotationOnlyHours, result.RotationOnlyNoise, result.Passes)
	}
	return text
}
