package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
)

// writePlayer writes the sim player built from a character's newest export to `out`, and
// exits without reading a log. It is for tools that sim our own gear in a fight of their own.
// bayovui's /update-stat-weights skill takes the gear, talents, race, level and professions from
// it into a mksettings request, and measures stat weights on that.
//
// It is the same player logcheck replays: the export's weapon skill unless fullSkill, and no
// buffs, consumes, rotation or imbue. The JSON is
//
//	{ "player": <proto.Player>, "notes": [what we had to leave out], "export": <where it came from>,
//	  "date": <when it was taken>, "char": "Name-Realm", "guid": "Player-...", "class": "SHAMAN", "level": 30 }
//
// char names the character for bayovui, which matches it against the logged in one. The guid is
// what knows it for sure.
func writePlayer(archive, exportPath, char string, fullSkill bool, out string) error {
	var ex *export
	var err error
	switch {
	case exportPath != "":
		ex, err = loadExport(exportPath)
	case char != "":
		var all []*export
		var guid string
		all, err = loadExports(archive)
		if err == nil {
			guid, err = findChar(all, char)
		}
		if err == nil {
			// The newest of all. Export times are the game's local time, read as UTC, so
			// time.Now() would leave out the last few hours.
			ex, err = pickExport(all, time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC), map[string]bool{guid: true})
		}
	default:
		return fmt.Errorf("-player-out needs -char (a GUID or a name) or -export")
	}
	if err != nil {
		return err
	}
	build := fmt.Sprintf("%s.%d", ex.When.Version, ex.When.Build)
	tables, err := loadClientTables(wagoCache(), build)
	if err != nil {
		return err
	}
	c, err := buildPlayer(ex, tables)
	if err != nil {
		return err
	}
	if !fullSkill {
		setSkillGaps(c.player, ex)
	}
	player, err := protojson.Marshal(c.player)
	if err != nil {
		return err
	}
	notes := c.notes
	if notes == nil {
		notes = []string{}
	}
	doc, err := json.MarshalIndent(map[string]any{
		"player": json.RawMessage(player),
		"notes":  notes,
		"export": ex.path,
		"date":   ex.When.Date,
		"char":   ex.Who.Name + "-" + ex.Who.Realm,
		"guid":   ex.guid,
		"class":  ex.Who.ClassFile,
		"level":  ex.Who.Level,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(out, append(doc, '\n'), 0644)
}
