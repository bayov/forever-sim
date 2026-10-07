package database

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The dungeon loot tables of https://foreverchanges.pro, scraped by
// tools/database/foreverchanges/scrape.py.
//
// wowhead's Forever database has the stats of the items Forever added to the low level
// dungeons, but no drop data for most of them. Without a source the gear search treats
// an item as unconfirmed and we leave it out of the shipped sets, which is how First
// Mate Band (Mr. Smite) and every Ruins of Lordaeron drop stayed out of the level 20
// presets. ForeverChanges lists each boss with its loot, so the items get their drop
// source from here.
type ForeverChangesLoot struct {
	Items []ForeverChangesDrop `json:"items"`
}

type ForeverChangesDrop struct {
	// The dungeon's page slug, e.g. "ruins-of-lordaeron".
	Dungeon string `json:"dungeon"`
	// The boss name, "Trash mobs" for trash drops.
	Boss string `json:"boss"`
	// boss, trash, rare or object.
	Kind     string `json:"kind"`
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	ReqLevel int32  `json:"reqlevel"`
	// new, changed or same, compared with Classic.
	Status  string   `json:"status"`
	Tooltip []string `json:"tooltip"`
}

func ParseForeverChangesLoot(contents string) ForeverChangesLoot {
	var loot ForeverChangesLoot
	if err := json.Unmarshal([]byte(contents), &loot); err != nil {
		log.Fatalf("failed to parse foreverchanges loot: %s", err)
	}
	fmt.Printf("\n--\nForeverChanges dungeon drops loaded: %d\n--\n", len(loot.Items))
	return loot
}

// The wowhead zone ID of each dungeon page. The two Forever dungeons use the IDs of
// wowhead's Forever zone listing.
var foreverChangesDungeonZones = map[string]int32{
	"ragefire-chasm":     2437,
	"the-deadmines":      1581,
	"wailing-caverns":    718,
	"hall-of-thanes":     16919,
	"ruins-of-lordaeron": 16611,
	"shadowfang-keep":    209,
	"blackfathom-deeps":  719,
	"the-stockade":       717,
	"gnomeregan":         721,
	"razorfen-kraul":     491,
	"razorfen-downs":     722,
	// The four Scarlet Monastery wings share one zone.
	"scarlet-monastery-graveyard": 796,
	"scarlet-monastery-library":   796,
	"scarlet-monastery-armory":    796,
	"scarlet-monastery-cathedral": 796,
}

// ForeverZones are zones the Classic inputs do not carry: the two Forever dungeons, and
// the zones of the Forever rares whose drops we enter by hand.
var ForeverZones = []*proto.UIZone{
	{Id: 16611, Name: "Ruins of Lordaeron", Expansion: proto.Expansion_ExpansionVanilla},
	{Id: 16919, Name: "The Hall of Thanes", Expansion: proto.Expansion_ExpansionVanilla},
	{Id: 17, Name: "The Barrens", Expansion: proto.Expansion_ExpansionVanilla},
	{Id: 11, Name: "Wetlands", Expansion: proto.Expansion_ExpansionVanilla},
}

// Source is the drop source for the item. We have no NPC IDs for the Forever bosses,
// so the boss goes in by name next to the dungeon.
func (d ForeverChangesDrop) Source() *proto.UIItemSource {
	zone, ok := foreverChangesDungeonZones[d.Dungeon]
	if !ok {
		log.Fatalf("foreverchanges: no zone for dungeon %q", d.Dungeon)
	}
	return &proto.UIItemSource{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{ZoneId: zone, OtherName: d.Boss}}}
}

// The random suffixes the beta client gives each item that Classic wowhead never saw on
// it, scraped by tools/database/foreverchanges/suffixes.py. Suffixes that match no wowhead
// suffix come with their stats (IDs from 990000 up), the rest are wowhead IDs.
type ForeverChangesSuffixes struct {
	Suffixes []struct {
		ID    int32           `json:"id"`
		Name  string          `json:"name"`
		Stats map[int]float64 `json:"stats"`
	} `json:"suffixes"`
	Items map[int32][]int32 `json:"items"`
}

func ParseForeverChangesSuffixes(contents string) ForeverChangesSuffixes {
	var suffixes ForeverChangesSuffixes
	if err := json.Unmarshal([]byte(contents), &suffixes); err != nil {
		log.Fatalf("failed to parse foreverchanges suffixes: %s", err)
	}
	return suffixes
}

// AttachForeverChangesSuffixes adds the suffixes to the items' options and puts the new
// suffixes in the database. The wowhead ones are filled in with the rest of the item
// options later.
func (db *WowDatabase) AttachForeverChangesSuffixes(suffixes ForeverChangesSuffixes) {
	for _, suffix := range suffixes.Suffixes {
		var s stats.Stats
		for stat, value := range suffix.Stats {
			s[stat] = value
		}
		db.RandomSuffixes[suffix.ID] = &proto.ItemRandomSuffix{Id: suffix.ID, Name: suffix.Name, Stats: s.ToFloatArray()}
	}
	added := 0
	for itemID, options := range suffixes.Items {
		item, ok := db.Items[itemID]
		if !ok {
			continue
		}
		for _, option := range options {
			if !slices.Contains(item.RandomSuffixOptions, option) {
				item.RandomSuffixOptions = append(item.RandomSuffixOptions, option)
				added++
			}
		}
	}
	fmt.Printf("ForeverChanges random suffixes added: %d\n", added)
}
