//go:build practice

package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The client's game tables (DB2s), as CSV from wago.tools for the exact build the client runs.
//
// We fetch each table the first time something needs it and keep it in
// ~/forever-practice/db2/<build>/. Reading the tables from the local client would need a DB2
// reader, the build's table layouts and the hotfix cache, which waits for Milestone 4.

type db2Tables struct {
	dir   string
	build string

	mu     sync.Mutex
	tables map[string]*db2Table
}

func newDB2Tables(home, build string) *db2Tables {
	return &db2Tables{dir: filepath.Join(home, "db2", build), build: build, tables: map[string]*db2Table{}}
}

// buildInfo reads the Forever line of the install's .build.info, with its Version
// ("1.60.1.70707") and Branch ("eu").
func (c *wowClient) buildInfo() (map[string]string, error) {
	f, err := os.Open(filepath.Join(c.dataRoot, ".build.info"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var cols []string
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "|")
		if cols == nil {
			for _, h := range fields {
				cols = append(cols, strings.SplitN(h, "!", 2)[0])
			}
			continue
		}
		row := map[string]string{}
		for i, v := range fields {
			if i < len(cols) {
				row[cols[i]] = v
			}
		}
		if row["Product"] == "wowf" {
			return row, nil
		}
	}
	return nil, fmt.Errorf("no wowf line in .build.info")
}

// Table loads a table, fetching it from wago.tools on the first use.
func (d *db2Tables) Table(name string) (*db2Table, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if t, ok := d.tables[name]; ok {
		return t, nil
	}
	path := filepath.Join(d.dir, name+".csv")
	if _, err := os.Stat(path); err != nil {
		if err := d.fetch(name, path); err != nil {
			return nil, err
		}
	}
	t, err := readDB2Table(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	d.tables[name] = t
	return t, nil
}

// MustTable is Table for code that would rather see an empty table than an error. A missing
// table then shows up as a missing model part, and the error goes to the log.
func (d *db2Tables) MustTable(name string) *db2Table {
	t, err := d.Table(name)
	if err != nil {
		log.Printf("db2: %v", err)
		return &db2Table{col: map[string]int{}, byID: map[int]int{}, indexes: map[string]map[int][]int{}}
	}
	return t
}

func (d *db2Tables) fetch(name, path string) error {
	if err := os.MkdirAll(d.dir, 0o755); err != nil {
		return err
	}
	url := fmt.Sprintf("https://wago.tools/db2/%s/csv?build=%s", name, d.build)
	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", url, resp.Status)
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// db2Table is one table's rows, with the columns by name.
type db2Table struct {
	header []string
	col    map[string]int
	rows   [][]string
	byID   map[int]int

	mu      sync.Mutex
	indexes map[string]map[int][]int
}

func readDB2Table(path string) (*db2Table, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	t := &db2Table{header: header, col: map[string]int{}, byID: map[int]int{}, indexes: map[string]map[int][]int{}}
	for i, h := range header {
		t.col[h] = i
	}
	idCol, hasID := t.col["ID"]
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		t.rows = append(t.rows, rec)
		if hasID && idCol < len(rec) {
			if id, err := strconv.Atoi(rec[idCol]); err == nil {
				t.byID[id] = len(t.rows) - 1
			}
		}
	}
	return t, nil
}

// db2Row is one row. A zero row (from a missing ID) reads as zeros and empty strings.
type db2Row struct {
	t   *db2Table
	rec []string
}

func (r db2Row) Ok() bool { return r.rec != nil }

func (r db2Row) Str(col string) string {
	if r.rec == nil {
		return ""
	}
	i, ok := r.t.col[col]
	if !ok || i >= len(r.rec) {
		return ""
	}
	return r.rec[i]
}

func (r db2Row) Int(col string) int {
	v, err := strconv.Atoi(r.Str(col))
	if err != nil {
		f, _ := strconv.ParseFloat(r.Str(col), 64)
		return int(f)
	}
	return v
}

func (r db2Row) Float(col string) float64 {
	v, _ := strconv.ParseFloat(r.Str(col), 64)
	return v
}

// Has says whether the table has a column, since columns change between builds.
func (t *db2Table) Has(col string) bool {
	_, ok := t.col[col]
	return ok
}

func (t *db2Table) Len() int { return len(t.rows) }

func (t *db2Table) Row(i int) db2Row { return db2Row{t, t.rows[i]} }

// Get finds a row by its ID.
func (t *db2Table) Get(id int) db2Row {
	if i, ok := t.byID[id]; ok {
		return db2Row{t, t.rows[i]}
	}
	return db2Row{}
}

// Where gives the rows whose column holds the value. The first call for a column builds its index.
func (t *db2Table) Where(col string, value int) []db2Row {
	t.mu.Lock()
	index, ok := t.indexes[col]
	if !ok {
		index = map[int][]int{}
		if c, has := t.col[col]; has {
			for i, rec := range t.rows {
				if c < len(rec) {
					v, err := strconv.Atoi(rec[c])
					if err != nil {
						f, _ := strconv.ParseFloat(rec[c], 64)
						v = int(f)
					}
					index[v] = append(index[v], i)
				}
			}
		}
		t.indexes[col] = index
	}
	t.mu.Unlock()
	var out []db2Row
	for _, i := range index[value] {
		out = append(out, db2Row{t, t.rows[i]})
	}
	return out
}

// First is Where's first row, or a zero row.
func (t *db2Table) First(col string, value int) db2Row {
	if rows := t.Where(col, value); len(rows) > 0 {
		return rows[0]
	}
	return db2Row{}
}
