//go:build practice

package main

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/wowsims/classic/assets/database"
	"github.com/wowsims/classic/sim/core/proto"
)

// Spell names and ranks from the wago.tools tables the repo already caches by client build. The
// page needs them to read macros like "/cast Frost Shock(Rank 1)".

type spellName struct {
	Name string `json:"name"`
	Rank string `json:"rank,omitempty"`
	Icon int    `json:"icon,omitempty"`
}

type spellNames struct {
	names map[int32]spellName
}

func loadSpellNames(wagoDir string) *spellNames {
	s := &spellNames{names: map[int32]spellName{}}
	builds, _ := os.ReadDir(wagoDir)
	var dirs []string
	for _, b := range builds {
		if b.IsDir() {
			dirs = append(dirs, filepath.Join(wagoDir, b.Name()))
		}
	}
	// Newest build last, so its rows win.
	sort.Strings(dirs)
	for _, dir := range dirs {
		readCSV(filepath.Join(dir, "SpellName.csv"), func(row map[string]string) {
			id := atoi32(row["ID"])
			n := s.names[id]
			n.Name = row["Name_lang"]
			s.names[id] = n
		})
		readCSV(filepath.Join(dir, "Spell.csv"), func(row map[string]string) {
			id := atoi32(row["ID"])
			if n, ok := s.names[id]; ok {
				n.Rank = row["NameSubtext_lang"]
				s.names[id] = n
			}
		})
		readCSV(filepath.Join(dir, "SpellMisc.csv"), func(row map[string]string) {
			id := atoi32(row["SpellID"])
			if n, ok := s.names[id]; ok {
				n.Icon, _ = strconv.Atoi(row["SpellIconFileDataID"])
				s.names[id] = n
			}
		})
	}
	return s
}

func readCSV(path string, each func(row map[string]string)) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil {
		return
	}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			return
		}
		if err != nil {
			continue
		}
		row := make(map[string]string, len(header))
		for i, h := range header {
			if i < len(rec) {
				row[h] = rec[i]
			}
		}
		each(row)
	}
}

func atoi32(s string) int32 {
	v, _ := strconv.Atoi(s)
	return int32(v)
}

var uiDatabase *proto.UIDatabase

// fillDatabase puts the items and enchants the player wears into player.database, which is how
// a WASM build with no item database gets them (as getDatabase does in sim/lib).
func fillDatabase(request *proto.RaidSimRequest) {
	if uiDatabase == nil {
		uiDatabase = database.Load()
	}
	items := map[int32]*proto.UIItem{}
	for _, item := range uiDatabase.Items {
		items[item.Id] = item
	}
	enchants := map[int32]*proto.UIEnchant{}
	for _, enchant := range uiDatabase.Enchants {
		enchants[enchant.EffectId] = enchant
	}
	suffixes := map[int32]*proto.ItemRandomSuffix{}
	for _, suffix := range uiDatabase.RandomSuffixes {
		suffixes[suffix.Id] = suffix
	}
	for _, party := range request.GetRaid().GetParties() {
		for _, player := range party.GetPlayers() {
			if player == nil || player.Equipment == nil {
				continue
			}
			db := &proto.SimDatabase{}
			for _, spec := range player.Equipment.Items {
				if spec == nil {
					continue
				}
				if item := items[spec.Id]; item != nil {
					db.Items = append(db.Items, &proto.SimItem{
						Id:                  item.Id,
						ClassAllowlist:      item.ClassAllowlist,
						Name:                item.Name,
						Type:                item.Type,
						ArmorType:           item.ArmorType,
						WeaponType:          item.WeaponType,
						HandType:            item.HandType,
						RangedWeaponType:    item.RangedWeaponType,
						Stats:               item.Stats,
						BonusPhysicalDamage: item.BonusPhysicalDamage,
						WeaponDamageMin:     item.WeaponDamageMin,
						WeaponDamageMax:     item.WeaponDamageMax,
						WeaponSpeed:         item.WeaponSpeed,
						SpellSchool:         item.SpellSchool,
						SetName:             item.SetName,
						SetId:               item.SetId,
						WeaponSkills:        item.WeaponSkills,
						HitRating:           item.HitRating,
						CritRating:          item.CritRating,
					})
				}
				if enchant := enchants[spec.Enchant]; enchant != nil {
					db.Enchants = append(db.Enchants, &proto.SimEnchant{EffectId: enchant.EffectId, Stats: enchant.Stats})
				}
				if suffix := suffixes[spec.RandomSuffix]; suffix != nil {
					db.RandomSuffixes = append(db.RandomSuffixes, suffix)
				}
			}
			player.Database = db
		}
	}
}
