package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// clientTables holds what we read from the Forever client's DB2 tables on wago.tools.
//
// We download each table once per client build into the archive's wago/ folder. When
// wago.tools doesn't have the export's build (or we're offline), we fall back to the newest
// build we already have.
type clientTables struct {
	build string
	// spellNames names every spell, for the sim's spells that never show up in the log.
	spellNames map[int32]string
	// enchantSpells maps an enchant ID (what an item link and the export carry) to the spells
	// that put it on an item. The sim's enchants are keyed by one or the other, depending on
	// where the entry came from.
	enchantSpells map[int32][]int32
}

func loadClientTables(dir, build string) (*clientTables, error) {
	tables := &clientTables{build: build, spellNames: map[int32]string{}, enchantSpells: map[int32][]int32{}}

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
	return tables, nil
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
	return os.WriteFile(path, body, 0o644)
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
