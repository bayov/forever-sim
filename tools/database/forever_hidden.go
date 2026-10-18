package database

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/wowsims/classic/sim/core/proto"
)

// MarkHiddenStats marks the items whose stats Forever still hides, from
// assets/database/hidden_items.json, so the UI can say the sim's stats for them
// are a guess.
//
// tools/sources/hidden.py writes that file after gen_db, because it reads db.json to
// tell which hidden items the sim has. So here we read the one the last run wrote.
// It's a run behind only for an item the sim gained since, and an item hidden.py
// didn't list as in the sim has no simStats, so we work its kind out the same way:
// Classic's stats below Forever's own IDs, part of them above.
//
// A missing file marks nothing, as on a fresh checkout.
func MarkHiddenStats(db *WowDatabase, path string) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		log.Fatalf("Cannot read %s: %v", path, err)
	}
	var file struct {
		Items map[string]struct {
			State    string `json:"state"`
			SimStats string `json:"simStats"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		log.Fatalf("Cannot parse %s: %v", path, err)
	}
	kinds := map[string]proto.UIItem_HiddenStats{
		"classic": proto.UIItem_HIDDEN_STATS_CLASSIC,
		"partial": proto.UIItem_HIDDEN_STATS_PARTIAL,
		"wowhead": proto.UIItem_HIDDEN_STATS_WOWHEAD,
	}
	marked := 0
	for key, e := range file.Items {
		if e.State != "hidden" {
			continue
		}
		id, err := strconv.Atoi(key)
		if err != nil {
			log.Fatalf("%s: item key %q isn't a number", path, key)
		}
		item, ok := db.Items[int32(id)]
		if !ok {
			continue
		}
		kind, ok := kinds[e.SimStats]
		if !ok {
			kind = proto.UIItem_HIDDEN_STATS_PARTIAL
			if id < foreverFirstID {
				kind = proto.UIItem_HIDDEN_STATS_CLASSIC
			}
		}
		item.HiddenStats = kind
		marked++
	}
	log.Printf("Marked %d items whose stats Forever still hides (%s)", marked, path)
}

// Forever's own items have IDs from here up, as in tools/sources/hidden.py.
const foreverFirstID = 250000
