package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// The gear search's faction check: an item only the other faction can get stays out of the
// pool, and comes off the starting set. Night Elf druids picked Horde quest rewards (Eye of
// the Beast, Blackhand's Breadth) before there was one, because our data has no faction on
// them (2026-11-23).
//
// The build's faction comes from its race (core.RaceFaction). An item belongs to one faction
// when the first of these says so:
//   - The content check (-gear-content, tools/rotopt/phase0_content.json) gives it a
//     "faction", for what our data can't tell: a Classic quest reward with no quest in our
//     data, a reputation vendor we have no row for.
//   - Its own restriction (UIItem.FactionRestriction), from wowhead's side of the item, or of
//     the quests when quests are its only source, or AtlasLoot's PvP reputation lists.
//   - Its sources, when every one of them belongs to the same faction: a reputation for one
//     faction's players, a quest only that side can take (-quest-sides) or a faction vendor
//     (factionVendors). An item with one source open to both factions stays.

var questSidesPath = flag.String("quest-sides", "assets/db_inputs/forever_wowhead_new_quests.json", "wowhead's Forever quests with the side that can take each, for the gear search's faction check. Empty skips it")

// factionVendors are Forever's reputation vendors only one faction can buy from. Our data has
// them as the "drop" source of what they sell.
var factionVendors = map[int32]proto.Faction{
	263569: proto.Faction_Alliance, // Bryanna Embreeze, Theramore Expeditionary Force, the Barrens
	263570: proto.Faction_Horde,    // Creeg Bothunk, Darkspear Raiders, the Barrens
}

type factionCheck struct {
	// The build's faction. Unknown keeps every item.
	faction proto.Faction
	// The content check's "faction" by item ID, and the quest sides by quest ID.
	tags   map[int32]proto.Faction
	quests map[int32]proto.Faction
	// What loadGearPool left out for it, by item ID, with the reason, for the dry run.
	dropped map[int32]poolItem
}

func parseFaction(s string) (proto.Faction, error) {
	switch strings.ToLower(s) {
	case "alliance":
		return proto.Faction_Alliance, nil
	case "horde":
		return proto.Faction_Horde, nil
	case "", "both":
		return proto.Faction_Unknown, nil
	}
	return proto.Faction_Unknown, fmt.Errorf("faction %q is not Alliance, Horde or both", s)
}

// newFactionCheck reads the content check's faction tags and the quest sides. Either path may
// be empty.
func newFactionCheck(race proto.Race, contentPath, questsPath string) (*factionCheck, error) {
	c := &factionCheck{faction: core.RaceFaction(race), tags: map[int32]proto.Faction{}, quests: map[int32]proto.Faction{}, dropped: map[int32]poolItem{}}
	if contentPath != "" {
		data, err := os.ReadFile(contentPath)
		if err != nil {
			return nil, err
		}
		var content map[string]struct {
			Faction string `json:"faction"`
		}
		if err := json.Unmarshal(data, &content); err != nil {
			return nil, fmt.Errorf("%s: %w", contentPath, err)
		}
		for k, v := range content {
			f, err := parseFaction(v.Faction)
			if err != nil {
				return nil, fmt.Errorf("%s: item %s: %w", contentPath, k, err)
			}
			if id, err := strconv.Atoi(k); err == nil && f != proto.Faction_Unknown {
				c.tags[int32(id)] = f
			}
		}
	}
	if questsPath != "" {
		data, err := os.ReadFile(questsPath)
		if err != nil {
			return nil, err
		}
		var file struct {
			Quests map[string]struct {
				Side string `json:"side"`
			} `json:"quests"`
		}
		if err := json.Unmarshal(data, &file); err != nil {
			return nil, fmt.Errorf("%s: %w", questsPath, err)
		}
		for k, q := range file.Quests {
			f, err := parseFaction(q.Side)
			if id, convErr := strconv.Atoi(k); err == nil && convErr == nil && f != proto.Faction_Unknown {
				c.quests[int32(id)] = f
			}
		}
	}
	return c, nil
}

// sourceFaction is the faction of one source, Unknown when both can use it.
func (c *factionCheck) sourceFaction(src *proto.UIItemSource) proto.Faction {
	switch {
	case src.GetRep() != nil:
		return src.GetRep().PlayerFaction
	case src.GetQuest() != nil:
		return c.quests[src.GetQuest().Id]
	case src.GetDrop() != nil:
		return factionVendors[src.GetDrop().NpcId]
	case src.GetSoldBy() != nil:
		return factionVendors[src.GetSoldBy().NpcId]
	}
	return proto.Faction_Unknown
}

// itemFaction is the only faction that can get the item, Unknown when both can, and what says
// so ("content", "item" or "sources").
func (c *factionCheck) itemFaction(item *proto.UIItem) (proto.Faction, string) {
	if f, ok := c.tags[item.Id]; ok {
		return f, "content"
	}
	switch item.FactionRestriction {
	case proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY:
		return proto.Faction_Alliance, "item"
	case proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY:
		return proto.Faction_Horde, "item"
	}
	if len(item.Sources) == 0 {
		return proto.Faction_Unknown, ""
	}
	f := c.sourceFaction(item.Sources[0])
	for _, src := range item.Sources[1:] {
		if c.sourceFaction(src) != f {
			return proto.Faction_Unknown, ""
		}
	}
	if f == proto.Faction_Unknown {
		return f, ""
	}
	return f, "sources"
}

// otherFaction says why only the other faction can get the item, or "" when the build can.
func (c *factionCheck) otherFaction(item *proto.UIItem) string {
	if c == nil || item == nil || c.faction == proto.Faction_Unknown {
		return ""
	}
	if f, why := c.itemFaction(item); f != proto.Faction_Unknown && f != c.faction {
		return why
	}
	return ""
}
