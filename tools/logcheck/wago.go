package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// clientTables holds what we read from the Forever client's DB2 tables on wago.tools.
//
// We download each table once per client build into the shared cache (wagoCache), the same
// folder tools/sources/wago.py and bayovui's build scripts use. When wago.tools doesn't have
// the export's build (or we're offline), we fall back to the newest build we already have.
type clientTables struct {
	build string
	// spellNames names every spell, for the sim's spells that never show up in the log.
	spellNames map[int32]string
	// enchantSpells maps an enchant ID (what an item link and the export carry) to the spells
	// that put it on an item. The sim's enchants are keyed by one or the other, depending on
	// where the entry came from.
	enchantSpells map[int32][]int32
	// enchants holds each enchant's effects, for the enchants the sim's database lacks.
	enchants map[int32]clientEnchant
	// equipStats holds the stats of each spell that only applies stat auras, by spell ID.
	// Some Forever enchants give their stats through such a spell (enchant 41, "Stamina +2",
	// casts Inferior Stamina, 7419).
	equipStats map[int32]stats.Stats
	// buffEffects holds the stat auras of each spell, for the buffs on us (buffs.go).
	buffEffects map[int32][]buffEffect
	// buffDurations holds how long each of those buffs lasts, 0 for one that lasts until it's
	// removed (a paladin aura).
	buffDurations map[int32]time.Duration
	// bonusShares holds the stat shares of each item bonus list, by its ID. Forever gives an
	// item its random suffix ("of the Falcon") as a bonus list in the item link.
	bonusShares map[int32][]statShare
	// randPropPoints gives an item's stat budget by item level, quality and budgetSlot,
	// which a bonus list's shares split.
	randPropPoints map[int32][3][5]float64
}

// statShare is one stat of an item bonus list: the client's ItemModType and its share of the
// item's budget, in 1/10000ths.
type statShare struct {
	modType, share int32
}

// clientEnchant is a row of SpellItemEnchantment. Each effect has a type (4 adds a
// resistance or armor, 5 adds a stat), an amount, and an argument that says which resistance
// or stat.
type clientEnchant struct {
	name    string
	effects []enchantEffect
}

type enchantEffect struct {
	kind, amount, arg int32
}

// wagoCache is the folder of client tables every tool shares, <version>/<Table>.csv under it.
// It's ~/.cache/wow-data/wago, or wago/ under $WOW_DATA_CACHE when that's set (see
// tools/sources/wago.py).
func wagoCache() string {
	root := os.Getenv("WOW_DATA_CACHE")
	if root == "" {
		home, _ := os.UserHomeDir()
		root = filepath.Join(home, ".cache", "wow-data")
	}
	return filepath.Join(root, "wago")
}

func loadClientTables(dir, build string) (*clientTables, error) {
	tables := &clientTables{build: build, spellNames: map[int32]string{}, enchantSpells: map[int32][]int32{}, enchants: map[int32]clientEnchant{}}

	names, err := wagoTable(dir, &tables.build, "SpellName")
	if err != nil {
		return nil, err
	}
	for _, row := range names.rows {
		tables.spellNames[row.int("ID")] = row.get("Name_lang")
	}

	effects, err := wagoTable(dir, &tables.build, "SpellEffect")
	if err != nil {
		return nil, err
	}
	for _, row := range effects.rows {
		// 53 puts a permanent enchant on an item, 54 a temporary one (oils, stones).
		if effect := row.get("Effect"); effect == "53" || effect == "54" {
			enchant := row.int("EffectMiscValue_0")
			tables.enchantSpells[enchant] = append(tables.enchantSpells[enchant], row.int("SpellID"))
		}
	}
	enchants, err := wagoTable(dir, &tables.build, "SpellItemEnchantment")
	if err != nil {
		return nil, err
	}
	for _, row := range enchants.rows {
		enchant := clientEnchant{name: row.get("Name_lang")}
		for i := range 3 {
			effect := enchantEffect{
				kind:   row.int(fmt.Sprintf("Effect_%d", i)),
				amount: row.int(fmt.Sprintf("EffectPointsMin_%d", i)),
				arg:    row.int(fmt.Sprintf("EffectArg_%d", i)),
			}
			if effect.kind != 0 {
				enchant.effects = append(enchant.effects, effect)
			}
		}
		tables.enchants[row.int("ID")] = enchant
	}

	// An aura that changes a stat (29, its misc value is the stat or -1 for all five) or
	// attack power (99). Any other effect makes the spell more than stats.
	tables.equipStats = map[int32]stats.Stats{}
	other := map[int32]bool{}
	primary := []stats.Stat{stats.Strength, stats.Agility, stats.Stamina, stats.Intellect, stats.Spirit}
	for _, row := range effects.rows {
		spell := row.int("SpellID")
		amount, _ := strconv.ParseFloat(row.get("EffectBasePointsF"), 64)
		s := tables.equipStats[spell]
		switch misc := row.int("EffectMiscValue_0"); {
		case row.get("Effect") != "6":
			other[spell] = true
		case row.get("EffectAura") == "29" && misc >= 0 && misc < 5:
			s[primary[misc]] += amount
		case row.get("EffectAura") == "29" && misc == -1:
			for _, stat := range primary {
				s[stat] += amount
			}
		case row.get("EffectAura") == "99":
			s[stats.AttackPower] += amount
		default:
			other[spell] = true
		}
		tables.equipStats[spell] = s
	}
	for spell := range other {
		delete(tables.equipStats, spell)
	}

	// The stat auras of every buff (buffs.go). An aura on ourselves (6) or a party or raid
	// aura (35 and 65, like a totem's or a paladin aura's).
	tables.buffEffects = map[int32][]buffEffect{}
	for _, row := range effects.rows {
		if effect := row.get("Effect"); effect != "6" && effect != "35" && effect != "65" {
			continue
		}
		aura := row.int("EffectAura")
		if !buffAuras[aura] {
			continue
		}
		amount, _ := strconv.ParseFloat(row.get("EffectBasePointsF"), 64)
		spell := row.int("SpellID")
		tables.buffEffects[spell] = append(tables.buffEffects[spell], buffEffect{aura: aura, misc: row.int("EffectMiscValue_0"), amount: amount})
	}
	if err := tables.loadBuffDurations(dir); err != nil {
		// Without them a buff whose removal the log missed stays on to the window's end.
		fmt.Fprintln(os.Stderr, "buff durations:", err)
	}

	if err := tables.loadItemBonuses(dir); err != nil {
		// Without them the random suffixes are left off, and the report says so.
		fmt.Fprintln(os.Stderr, "item bonuses:", err)
	}
	return tables, nil
}

// loadBuffDurations reads SpellMisc and SpellDuration for the buffs in buffEffects and
// buffOverrides.
func (t *clientTables) loadBuffDurations(dir string) error {
	build := t.build
	durations, err := wagoTable(dir, &build, "SpellDuration")
	if err != nil {
		return err
	}
	byIndex := map[int32]time.Duration{}
	for _, row := range durations.rows {
		if ms := row.int("Duration"); ms > 0 {
			byIndex[row.int("ID")] = time.Duration(ms) * time.Millisecond
		}
	}
	misc, err := wagoTable(dir, &build, "SpellMisc")
	if err != nil {
		return err
	}
	t.buffDurations = map[int32]time.Duration{}
	for _, row := range misc.rows {
		spell := row.int("SpellID")
		_, override := buffOverrides[spell]
		if _, ok := t.buffEffects[spell]; ok || override || spell == campBenefits {
			if d, ok := byIndex[row.int("DurationIndex")]; ok {
				t.buffDurations[spell] = d
			}
		}
	}
	return nil
}

// loadItemBonuses reads ItemBonus and RandPropPoints. They go on their own copy of
// the build, so a fallback to another build here doesn't change the spell tables' build.
func (t *clientTables) loadItemBonuses(dir string) error {
	build := t.build
	bonuses, err := wagoTable(dir, &build, "ItemBonus")
	if err != nil {
		return err
	}
	t.bonusShares = map[int32][]statShare{}
	for _, row := range bonuses.rows {
		// Type 2 is a stat: Value_0 is the ItemModType and Value_1 its share.
		if row.get("Type") == "2" {
			list := row.int("ParentItemBonusListID")
			t.bonusShares[list] = append(t.bonusShares[list], statShare{row.int("Value_0"), row.int("Value_1")})
		}
	}
	build = t.build
	points, err := wagoTable(dir, &build, "RandPropPoints")
	if err != nil {
		return err
	}
	t.randPropPoints = map[int32][3][5]float64{}
	for _, row := range points.rows {
		var budget [3][5]float64
		for q, column := range []string{"Good", "Superior", "Epic"} {
			for i := range 5 {
				budget[q][i], _ = strconv.ParseFloat(row.get(fmt.Sprintf("%s_%d", column, i)), 64)
			}
		}
		t.randPropPoints[row.int("ID")] = budget
	}
	return nil
}

// budgetSlot is RandPropPoints' column for an item: chest, legs, head and two-handers get the
// most, then shoulders, waist, feet, hands and trinkets, then neck, wrists, rings, cloaks,
// shields and held in off-hand items, then one-handers, then ranged weapons. wago.tools' ItemSparse
// for Forever only has the items Forever added or changed, so we go by the sim's item.
func budgetSlot(item core.Item) (int, bool) {
	switch item.Type {
	case proto.ItemType_ItemTypeHead, proto.ItemType_ItemTypeChest, proto.ItemType_ItemTypeLegs:
		return 0, true
	case proto.ItemType_ItemTypeShoulder, proto.ItemType_ItemTypeWaist, proto.ItemType_ItemTypeFeet,
		proto.ItemType_ItemTypeHands, proto.ItemType_ItemTypeTrinket:
		return 1, true
	case proto.ItemType_ItemTypeNeck, proto.ItemType_ItemTypeWrist, proto.ItemType_ItemTypeFinger,
		proto.ItemType_ItemTypeBack:
		return 2, true
	case proto.ItemType_ItemTypeRanged:
		return 4, true
	case proto.ItemType_ItemTypeWeapon:
		switch {
		case item.HandType == proto.HandType_HandTypeTwoHand:
			return 0, true
		case item.WeaponType == proto.WeaponType_WeaponTypeShield || item.WeaponType == proto.WeaponType_WeaponTypeOffHand:
			return 2, true
		}
		return 3, true
	}
	return 0, false
}

// bonusStats is what an item's bonus lists add: each stat share of the item's budget, rounded
// half up like tools/sources/crosscheck.py. ok is false when a list has stats we can't place.
func (t *clientTables) bonusStats(item core.Item, lists []int32) (s stats.Stats, ok bool) {
	for _, list := range lists {
		shares := t.bonusShares[list]
		if len(shares) == 0 {
			continue
		}
		slot, hasSlot := budgetSlot(item)
		// The sim's item table keeps neither the item level nor the quality.
		db := dbItemLevels()[item.ID]
		budgets, hasLevel := t.randPropPoints[db.Ilvl]
		quality := int(db.Quality) - int(proto.ItemQuality_ItemQualityUncommon)
		if !hasSlot || !hasLevel || quality < 0 || quality > 2 {
			return s, false
		}
		budget := budgets[quality][slot]
		for _, share := range shares {
			stat, known := enchantStats[share.modType]
			if !known {
				return s, false
			}
			s[stat] += math.Floor(budget*float64(share.share)/10000 + 0.5)
		}
	}
	return s, true
}

type wagoRows struct {
	rows []wagoRow
}

type wagoRow struct {
	header map[string]int
	values []string
}

func (r wagoRow) get(column string) string {
	if i, ok := r.header[column]; ok && i < len(r.values) {
		return r.values[i]
	}
	return ""
}

func (r wagoRow) int(column string) int32 {
	n, _ := strconv.Atoi(r.get(column))
	return int32(n)
}

// wagoTable reads a table for *build, downloading it when we don't have it. When the download
// fails it uses the newest build on disk and sets *build to it.
func wagoTable(dir string, build *string, table string) (*wagoRows, error) {
	path := filepath.Join(dir, *build, table+".csv")
	if _, err := os.Stat(path); err != nil {
		if err := download(fmt.Sprintf("https://wago.tools/db2/%s/csv?build=%s", table, *build), path); err != nil {
			fallback := newestBuild(dir, table)
			if fallback == "" {
				return nil, fmt.Errorf("%s for build %s: %w", table, *build, err)
			}
			fmt.Fprintf(os.Stderr, "wago.tools: no %s for build %s (%v), using build %s\n", table, *build, err, fallback)
			*build = fallback
			path = filepath.Join(dir, fallback, table+".csv")
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	headerRow, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	header := map[string]int{}
	for i, name := range headerRow {
		header[name] = i
	}
	rows := &wagoRows{}
	for {
		values, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		rows.rows = append(rows.rows, wagoRow{header: header, values: values})
	}
	return rows, nil
}

func download(url, path string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %s", response.Status)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if len(body) < 100 {
		return fmt.Errorf("only %d bytes", len(body))
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// The audit runs several windows at once, and each may download the same table. So we
	// write to a temporary file and rename it into place. Another window then sees either no
	// file (and downloads its own copy) or the whole file, never a half-written one.
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(body); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// newestBuild is the highest build number under dir that has the table.
func newestBuild(dir, table string) string {
	entries, _ := os.ReadDir(dir)
	best := ""
	for _, entry := range entries {
		if _, err := os.Stat(filepath.Join(dir, entry.Name(), table+".csv")); err == nil && entry.Name() > best {
			best = entry.Name()
		}
	}
	return best
}
