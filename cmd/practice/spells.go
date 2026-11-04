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
	// Tab is the spellbook tab the spell sits on (its class skill line), for the room's spellbook.
	Tab *spellTab `json:"tab,omitempty"`
}

// spellTab is a skill line of the spellbook: its name and the icon file of its side tab.
type spellTab struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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

// spellTab finds a spell's spellbook tab in the client's SkillLineAbility table: the class skill
// line (category 7) it belongs to, or General (skill line 0 in the game) for the class's spells
// on no tree line. It is nil without a client or when the spell is no class ability.
func (s *server) spellTab(id int32) *spellTab {
	if s.db2 == nil {
		return nil
	}
	abilities, err := s.db2.Table("SkillLineAbility")
	if err != nil {
		return nil
	}
	lines, err := s.db2.Table("SkillLine")
	if err != nil {
		return nil
	}
	for _, row := range abilities.Where("Spell", int(id)) {
		line := lines.Get(row.Int("SkillLine"))
		if !line.Ok() {
			continue
		}
		// Category 7 is a class's talent tree lines (Enhancement, Combat), 8 the armor and
		// weapon skills, 11 professions. Only the trees have their own tab.
		if line.Int("CategoryID") == 7 {
			return &spellTab{ID: line.Int("ID"), Name: line.Str("DisplayName_lang"), Icon: line.Int("SpellIconFileID")}
		}
	}
	return nil
}
