//go:build practice

// The practice server: serves the practice room's page from dist/practice, and reads the user's
// WoW Forever client for keybinds, action bars and CVars (docs/practice_room.md). Everything it
// writes goes to ~/forever-practice, never to the WoW folder or the repo.
//
//	go run -tags with_db,practice ./cmd/practice -wow "/mnt/d/Gaming/World of Warcraft"
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/practice"
	"google.golang.org/protobuf/encoding/protojson"
)

type server struct {
	client  *wowClient // nil when no WoW folder was found
	home    string     // ~/forever-practice
	repo    string
	dist    string
	spells  *spellNames
	icons   *iconCache
	records []*record
}

func main() {
	home, _ := os.UserHomeDir()
	addr := flag.String("addr", "127.0.0.1:3434", "address to serve on")
	wow := flag.String("wow", "/mnt/d/Gaming/World of Warcraft", "the WoW install folder (with .build.info)")
	repo := flag.String("repo", ".", "the forever-sim repo")
	data := flag.String("data", filepath.Join(home, "forever-practice"), "where the room keeps its files")
	wago := flag.String("wago", filepath.Join(home, "forever-logs", "wago"), "cached wago.tools tables, by build")
	flag.Parse()

	s := &server{home: *data, repo: *repo, dist: filepath.Join(*repo, "dist", "practice")}
	for _, dir := range []string{"attempts", "settings"} {
		if err := os.MkdirAll(filepath.Join(s.home, dir), 0o755); err != nil {
			log.Fatal(err)
		}
	}
	if client, err := findClient(*wow); err != nil {
		log.Printf("no WoW client: %v (the room uses the default bindings)", err)
	} else {
		s.client = client
		log.Printf("WoW Forever client: %s", client.forever)
	}
	s.spells = loadSpellNames(*wago)
	s.icons = newIconCache(filepath.Join(s.home, "icons.json"))
	log.Printf("%d spell names from %s", len(s.spells.names), *wago)

	log.Printf("practice room at http://%s/", *addr)
	log.Fatal(http.ListenAndServe(*addr, s.routes()))
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/request", s.handleRequest)
	mux.HandleFunc("GET /api/presets", s.handlePresets)
	mux.HandleFunc("GET /api/presets/{name}", s.handlePreset)
	mux.HandleFunc("GET /api/records", s.handleRecords)
	mux.HandleFunc("GET /api/character/{guid}", s.handleCharacter)
	mux.HandleFunc("PUT /api/settings/{guid}", s.handleSaveSettings)
	mux.HandleFunc("GET /api/spells", s.handleSpells)
	mux.HandleFunc("GET /api/icon/{kind}/{id}", s.handleIcon)
	mux.HandleFunc("GET /api/tooltip/{kind}/{id}", s.handleTooltip)
	mux.HandleFunc("GET /api/font/eui", s.handleFont)
	mux.HandleFunc("GET /api/attempts", s.handleListAttempts)
	mux.HandleFunc("POST /api/attempts", s.handleSaveAttempt)
	mux.HandleFunc("GET /api/attempts/{id}", s.handleGetAttempt)
	mux.HandleFunc("POST /api/replay", s.handleReplay)
	mux.Handle("GET /forever/assets/", http.StripPrefix("/forever/assets/", http.FileServer(http.Dir(filepath.Join(s.repo, "assets")))))
	mux.Handle("GET /", http.FileServer(http.Dir(s.dist)))
	return logged(mux)
}

func logged(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/practice.html", http.StatusFound)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write: %v", err)
	}
}

func writeError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

// handleRequest turns the class sim's export (a link or JSON) into the request the session
// runs: normalized, with the item database the WASM build doesn't have.
func (s *server) handleRequest(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 16<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	original, err := practice.ParseRequest(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	request, err := practice.Normalize(original)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	fillDatabase(request)
	data, err := protojson.Marshal(request)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	player := practice.Player(original)
	reply := map[string]any{
		"request":        json.RawMessage(data),
		"rotationSpells": rotationSpells(player.GetRotation()),
		"class":          player.GetClass().String(),
		"race":           player.GetRace().String(),
		"level":          player.GetLevel(),
		"name":           player.GetName(),
		"trinkets":       trinkets(player),
	}
	writeJSON(w, reply)
}

// trinkets are the two trinket slots' item ids, for the "Trinket 1" and "Trinket 2" of the
// keybind notes.
func trinkets(player *proto.Player) []int32 {
	out := []int32{0, 0}
	items := player.GetEquipment().GetItems()
	for i, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2} {
		if int(slot) < len(items) && items[slot] != nil {
			out[i] = items[slot].Id
		}
	}
	return out
}

// rotationSpells are the spells the imported rotation casts, in its order, for the bars of a
// character with no BayovCore record. The rotation only picks them, it never casts them.
func rotationSpells(rotation *proto.APLRotation) []int32 {
	var out []int32
	seen := map[int32]bool{}
	var walk func(v any)
	walk = func(v any) {
		switch t := v.(type) {
		case map[string]any:
			if cast, ok := t["castSpell"].(map[string]any); ok {
				if id, ok := cast["spellId"].(map[string]any); ok {
					if f, ok := id["spellId"].(float64); ok && !seen[int32(f)] {
						seen[int32(f)] = true
						out = append(out, int32(f))
					}
				}
			}
			keys := make([]string, 0, len(t))
			for k := range t {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				walk(t[k])
			}
		case []any:
			for _, e := range t {
				walk(e)
			}
		}
	}
	if rotation == nil {
		return out
	}
	data, err := protojson.Marshal(rotation)
	if err != nil {
		return out
	}
	var tree map[string]any
	json.Unmarshal(data, &tree)
	// Prepull actions first, then the priority list in order.
	for _, key := range []string{"prepullActions", "priorityList"} {
		if list, ok := tree[key].([]any); ok {
			for _, e := range list {
				walk(e)
			}
		}
	}
	return out
}

var presetPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// handlePresets lists the DPS Report's builds, as a quick way into the room.
func (s *server) handlePresets(w http.ResponseWriter, r *http.Request) {
	matches, _ := filepath.Glob(filepath.Join(s.repo, "tools", "dps_report", "req", "*.json"))
	out := []string{}
	for _, m := range matches {
		name := strings.TrimSuffix(filepath.Base(m), ".json")
		if !strings.HasSuffix(name, ".meta") {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	writeJSON(w, out)
}

func (s *server) handlePreset(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !presetPattern.MatchString(name) {
		writeError(w, http.StatusBadRequest, errors.New("bad preset name"))
		return
	}
	data, err := os.ReadFile(filepath.Join(s.repo, "tools", "dps_report", "req", name+".json"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

type recordSummary struct {
	GUID  string `json:"guid"`
	Name  string `json:"name"`
	Realm string `json:"realm"`
	Class string `json:"class"`
	Race  string `json:"race"`
	Level int    `json:"level"`
}

func (s *server) loadRecords() ([]*record, error) {
	if s.client == nil {
		return nil, nil
	}
	return s.client.records()
}

func (s *server) handleRecords(w http.ResponseWriter, r *http.Request) {
	records, err := s.loadRecords()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := []recordSummary{}
	for _, rec := range records {
		out = append(out, recordSummary{GUID: rec.GUID, Name: rec.Name, Realm: rec.Realm, Class: rec.Class, Race: rec.Race, Level: rec.Level})
	}
	writeJSON(w, out)
}

type characterReply struct {
	Record   *record           `json:"record,omitempty"`
	Bindings map[string]string `json:"bindings"`
	CVars    map[string]string `json:"cvars"`
	// UI is the account's EllesmereUI profile, NoteBinds the class keybind notes (ui.go).
	UI        map[string]any    `json:"ui,omitempty"`
	NoteBinds map[string]string `json:"noteBinds,omitempty"`
	// Saved are the room's own edits (bars, binds, fight settings) on top of the client's.
	Saved json.RawMessage `json:"saved,omitempty"`
}

var guidPattern = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

// handleCharacter gives a character's bindings, bars and CVars. "default" is a character with no
// record: Blizzard's default bindings and no bars.
func (s *server) handleCharacter(w http.ResponseWriter, r *http.Request) {
	guid := r.PathValue("guid")
	if !guidPattern.MatchString(guid) {
		writeError(w, http.StatusBadRequest, errors.New("bad GUID"))
		return
	}
	reply := characterReply{}
	account, name := "", ""
	if guid != "default" {
		records, err := s.loadRecords()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		for _, rec := range records {
			if rec.GUID == guid {
				reply.Record = rec
				account, name = rec.Account, rec.Name
			}
		}
		if reply.Record == nil {
			writeError(w, http.StatusNotFound, errors.New("no record for "+guid))
			return
		}
	}
	if s.client != nil {
		reply.Bindings = s.client.bindings(account, name)
		reply.CVars = s.client.cvars(account, name)
		reply.UI = s.client.ellesmereUI(account)
		if reply.Record != nil {
			reply.NoteBinds = s.client.noteBinds(reply.Record.Class)
		}
	} else {
		reply.Bindings = map[string]string{}
		parseBindings(defaultBindings, reply.Bindings)
		reply.CVars = map[string]string{}
	}
	if data, err := os.ReadFile(filepath.Join(s.home, "settings", guid+".json")); err == nil {
		reply.Saved = data
	}
	writeJSON(w, reply)
}

func (s *server) handleSaveSettings(w http.ResponseWriter, r *http.Request) {
	guid := r.PathValue("guid")
	if !guidPattern.MatchString(guid) {
		writeError(w, http.StatusBadRequest, errors.New("bad GUID"))
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
	if err != nil || !json.Valid(body) {
		writeError(w, http.StatusBadRequest, errors.New("the settings must be JSON"))
		return
	}
	if err := os.WriteFile(filepath.Join(s.home, "settings", guid+".json"), body, 0o644); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, map[string]bool{"ok": true})
}

// handleSpells gives names and ranks for spell ids: ?ids=1,2,3, or every spell with ?all=1.
func (s *server) handleSpells(w http.ResponseWriter, r *http.Request) {
	out := map[string]spellName{}
	for _, part := range strings.Split(r.URL.Query().Get("ids"), ",") {
		var id int32
		if _, err := fmt.Sscan(part, &id); err == nil {
			if n, ok := s.spells.names[id]; ok {
				out[part] = n
			}
		}
	}
	writeJSON(w, out)
}

var attemptIDPattern = regexp.MustCompile(`^[0-9]{8}-[0-9]{6}-[0-9a-f]{6}$`)

type attemptSummary struct {
	ID        string  `json:"id"`
	Mode      string  `json:"mode"`
	DPS       float64 `json:"dps"`
	LengthMs  float64 `json:"lengthMs"`
	Character string  `json:"character,omitempty"`
}

func (s *server) handleListAttempts(w http.ResponseWriter, r *http.Request) {
	entries, _ := os.ReadDir(filepath.Join(s.home, "attempts"))
	out := []attemptSummary{}
	for _, e := range entries {
		id := strings.TrimSuffix(e.Name(), ".json")
		if !attemptIDPattern.MatchString(id) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.home, "attempts", e.Name()))
		if err != nil {
			continue
		}
		var a practice.Attempt
		if json.Unmarshal(data, &a) != nil {
			continue
		}
		sum := attemptSummary{ID: id, Mode: a.Mode, LengthMs: a.EndedAtMs}
		if a.Result != nil {
			sum.DPS = a.Result.DPS
		}
		out = append(out, sum)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	writeJSON(w, out)
}

func (s *server) handleSaveAttempt(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 64<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var a practice.Attempt
	if err := json.Unmarshal(body, &a); err != nil || a.Version != 1 {
		writeError(w, http.StatusBadRequest, errors.New("not an attempt"))
		return
	}
	suffix := make([]byte, 3)
	rand.Read(suffix)
	id := time.Now().Format("20060102-150405") + "-" + hex.EncodeToString(suffix)
	if err := os.WriteFile(filepath.Join(s.home, "attempts", id+".json"), body, 0o644); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, map[string]string{"id": id})
}

func (s *server) handleGetAttempt(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !attemptIDPattern.MatchString(id) {
		writeError(w, http.StatusBadRequest, errors.New("bad attempt id"))
		return
	}
	data, err := os.ReadFile(filepath.Join(s.home, "attempts", id+".json"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// handleReplay runs an attempt again natively and gives its summary and log.
func (s *server) handleReplay(w http.ResponseWriter, r *http.Request) {
	var a practice.Attempt
	if err := json.NewDecoder(io.LimitReader(r.Body, 64<<20)).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	replay, err := practice.Replay(&a)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, map[string]any{"summary": replay.Summary(), "log": replay.FullLog()})
}
