package database

import (
	"encoding/json"
	"log"
	"sort"
	"strconv"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/tools"
)

// The items whose stats we saw in game, from assets/db_inputs/forever_seen_in_game.json.
//
// The beta server sends some items to the game with stats that differ from Wowhead's
// Forever listing, and Forever hides others until someone finds them. BayovCore records
// the weapons and armor the user sees in game, and bayovui's build/seen_items.py (run by
// the sync skill) writes the ones whose stats differ from our database, or that it lacks,
// into the file. Entries with "by": "hand" are the ones we typed in before, from
// foreverchanges.pro and our own tooltips. gen_db merges them after the Wowhead listing,
// so they win.
//
// An item the database already has takes only the fields the entry gives, and its type
// too when the entry has an inventoryType. An item it doesn't have yet needs that slot and
// type, from the entry's inventoryType, classID and subclassID (the game's own numbers,
// which are also Wowhead's), or it's left out.

type foreverSeenItem struct {
	Name            string             `json:"name"`
	Ilvl            int32              `json:"ilvl"`
	RequiredLevel   int32              `json:"requiredLevel"`
	Quality         int32              `json:"quality"`
	Stats           map[string]float64 `json:"stats"`
	WeaponDamageMin float64            `json:"weaponDamageMin"`
	WeaponDamageMax float64            `json:"weaponDamageMax"`
	WeaponSpeed     float64            `json:"weaponSpeed"`
	HitRating       float64            `json:"hitRating"`
	CritRating      float64            `json:"critRating"`
	InventoryType   int32              `json:"inventoryType"`
	ClassID         int32              `json:"classID"`
	SubclassID      int32              `json:"subclassID"`
	By              string             `json:"by"`
}

// ReadForeverSeenInGame reads the file into items for MergeItems, with the database's
// items so it can tell a new item, which needs a slot, from one it overrides. It stops
// gen_db on a stat name the sim doesn't have, since that stat would be lost silently.
func ReadForeverSeenInGame(path string, have map[int32]*proto.UIItem) []*proto.UIItem {
	var file struct {
		Items map[string]foreverSeenItem `json:"items"`
	}
	if err := json.Unmarshal([]byte(tools.ReadFile(path)), &file); err != nil {
		log.Fatalf("Cannot parse %s: %v", path, err)
	}
	byName := map[string]stats.Stat{}
	for s := stats.Stat(0); s < stats.Len; s++ {
		byName[s.StatName()] = s
	}
	var out []*proto.UIItem
	for key, e := range file.Items {
		id, err := strconv.Atoi(key)
		if err != nil {
			log.Fatalf("%s: item key %q isn't a number", path, key)
		}
		var s stats.Stats
		for name, value := range e.Stats {
			stat, ok := byName[name]
			if !ok {
				log.Fatalf("%s: item %d (%s) has %s, which the sim has no stat for", path, id, e.Name, name)
			}
			s[stat] = value
		}
		item := &proto.UIItem{
			Id: int32(id), Name: e.Name, Stats: s.ToFloatArray(),
			WeaponDamageMin: e.WeaponDamageMin, WeaponDamageMax: e.WeaponDamageMax, WeaponSpeed: e.WeaponSpeed,
			HitRating: e.HitRating, CritRating: e.CritRating,
			Ilvl: e.Ilvl, RequiredLevel: e.RequiredLevel, Quality: proto.ItemQuality(e.Quality),
		}
		old, known := have[int32(id)]
		// MergeItem takes no faction restriction to mean none, and the game doesn't
		// tell us an item's faction. So an item keeps the one it has (Ring of Scorn
		// is Horde only).
		if known {
			item.FactionRestriction = old.FactionRestriction
		}
		if !known || e.InventoryType != 0 {
			wi := ForeverWowheadItem{ID: int32(id), Slot: e.InventoryType, Class: e.ClassID, Subclass: e.SubclassID}
			itemType, handType, ok := wi.itemType()
			if !ok {
				log.Printf("%s: item %d (%s) is new to the database and has no slot we know, so it's left out", path, id, e.Name)
				continue
			}
			item.Type, item.HandType = itemType, handType
			item.ArmorType, item.WeaponType, item.RangedWeaponType = wi.armorType(), wi.weaponType(), wi.rangedWeaponType()
			if !known {
				item.Phase, item.Expansion = 1, proto.Expansion_ExpansionVanilla
			}
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Id < out[j].Id })
	return out
}
