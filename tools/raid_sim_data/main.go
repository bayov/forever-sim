// raid_sim_data writes the raid sim page's member data (ui/raid_sim/data/builds.json): every base build of
// the DPS Report with its Player for each faction (the build's best race there, with that race's own
// rotation and gear tuning, as the comp search sims it), and its duty and encounter variants' Players.
//
//	go run -tags with_db ./tools/raid_sim_data
//
// It reads tools/dps_report/req through the comp search's pool, so the page's members are the search's.
// With them, the report's fight (the page's Generic boss). Run it again when the report's requests change.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/wowsims/classic/sim/compsearch"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// A build's request for one purpose: the base build, a duty variant or an encounter variant.
type variant struct {
	ID        string                     `json:"id"`
	Name      string                     `json:"name"`
	Duty      string                     `json:"duty,omitempty"`
	MobType   string                     `json:"mobType,omitempty"`
	Published float64                    `json:"published"`
	Players   map[string]json.RawMessage `json:"players"` // by faction, protojson
}

type build struct {
	variant
	Class      string            `json:"class"`
	ClassName  string            `json:"className"`
	Spec       string            `json:"spec"`
	Summary    string            `json:"summary,omitempty"`
	IsDefault  bool              `json:"isDefault,omitempty"`
	Gives      string            `json:"gives"`
	Package    string            `json:"package,omitempty"`
	Races      map[string]string `json:"races"`
	Duties     []*variant        `json:"duties,omitempty"`
	Encounters []*variant        `json:"encounters,omitempty"`
}

type meta struct {
	Summary   string `json:"summary"`
	IsDefault bool   `json:"isDefault"`
}

var factions = []proto.Faction{proto.Faction_Horde, proto.Faction_Alliance}

func main() {
	root := flag.String("root", ".", "the forever-sim checkout")
	out := flag.String("out", "ui/raid_sim/data/builds.json", "the data file, relative to the checkout")
	flag.Parse()

	data, builds := generate(*root)
	path := filepath.Join(*root, *out)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fail(err)
	}
	n := 0
	for _, b := range builds {
		n += 1 + len(b.Duties) + len(b.Encounters)
	}
	fmt.Printf("wrote %s: %d base builds, %d players a faction, %d KB\n", path, len(builds), n, len(data)/1024)
}

// generate is the data file's contents, from the search's own pool.
func generate(root string) ([]byte, []*build) {
	pool, err := compsearch.LoadPool(root)
	if err != nil {
		fail(err)
	}
	players := func(b *compsearch.Build) map[string]json.RawMessage {
		m := map[string]json.RawMessage{}
		for _, f := range factions {
			p, err := pool.Player(b, f)
			if err != nil {
				fail(err)
			}
			// The page adds each player's items from its own copy of the database.
			p.Database = nil
			data, err := protojson.Marshal(p)
			if err != nil {
				fail(err)
			}
			m[f.String()] = data
		}
		return m
	}
	readMeta := func(id string) meta {
		var m meta
		if data, err := os.ReadFile(filepath.Join(root, "tools", "dps_report", "req", id+".meta.json")); err == nil {
			_ = json.Unmarshal(data, &m)
		}
		return m
	}

	var builds []*build
	for _, b := range pool.Bases {
		m := readMeta(b.ID)
		bo := &build{
			variant:   variant{ID: b.ID, Name: b.Name, Published: b.Published, Players: players(b)},
			Class:     b.Class.String(),
			ClassName: b.ClassName,
			Spec:      b.Spec,
			Summary:   m.Summary,
			IsDefault: m.IsDefault,
			Gives:     b.Gives.String(),
			Races:     map[string]string{},
		}
		if b.Pkg != proto.RaidPackage_RaidPackageNone {
			bo.Package = b.Pkg.String()
		}
		for _, f := range factions {
			bo.Races[f.String()] = b.Race[f].String()
		}
		var duties []proto.Duty
		for d := range b.Duties {
			duties = append(duties, d)
		}
		sort.Slice(duties, func(i, j int) bool { return duties[i] < duties[j] })
		for _, d := range duties {
			v := b.Duties[d]
			bo.Duties = append(bo.Duties, &variant{ID: v.ID, Name: v.Name, Duty: d.String(), Published: v.Published, Players: players(v)})
		}
		var mobs []proto.MobType
		for t := range b.Vs {
			mobs = append(mobs, t)
		}
		sort.Slice(mobs, func(i, j int) bool { return mobs[i] < mobs[j] })
		for _, t := range mobs {
			v := b.Vs[t]
			bo.Encounters = append(bo.Encounters, &variant{ID: v.ID, Name: v.Name, MobType: t.String(), Published: v.Published, Players: players(v)})
		}
		builds = append(builds, bo)
		// A pet variant (a Tallstrider or wolf hunter) is a build of its own, as in the search's pool.
		for _, pv := range b.Pets {
			pm := readMeta(pv.ID)
			pb := &build{
				variant:   variant{ID: pv.ID, Name: pv.Name, Published: pv.Published, Players: players(pv)},
				Class:     pv.Class.String(),
				ClassName: pv.ClassName,
				Spec:      pv.Spec,
				Summary:   pm.Summary,
				Gives:     pv.Gives.String(),
				Races:     map[string]string{},
			}
			for _, f := range factions {
				pb.Races[f.String()] = pv.Race[f].String()
			}
			builds = append(builds, pb)
		}
	}
	sort.SliceStable(builds, func(i, j int) bool {
		if builds[i].ClassName != builds[j].ClassName {
			return builds[i].ClassName < builds[j].ClassName
		}
		return builds[i].ID < builds[j].ID
	})

	// The report's fight (EncounterLevel60), which every base build's request has: the page's Generic boss.
	encounter, err := protojson.Marshal(pool.Bases[0].Req.Encounter)
	if err != nil {
		fail(err)
	}
	data, err := json.Marshal(map[string]any{"encounter": json.RawMessage(encounter), "builds": builds})
	if err != nil {
		fail(err)
	}
	return append(data, '\n'), builds
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
