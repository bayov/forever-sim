//go:build practice

package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Files from the Forever client's storage, by FileDataID (the client's file number): models,
// skeletons, animations, textures and sounds.
//
// We extract them with TACTTool (from wowdev's TACTSharp) into ~/forever-practice/assets/ and serve
// them from there. TACTTool loads the build's tables on every run (about 2s), so we gather the
// files asked for in the same moment into one run. When a model comes out, we also queue the files
// it names (its skin, textures and skeleton), so the page doesn't wait a whole run for each.
//
// TACTTool reads the decryption keys from WoW.txt in its own folder. A file whose key is missing
// comes out zero-filled with no error, so we check every file before we keep it.

type assetStore struct {
	dir     string // ~/forever-practice/assets
	tool    string // the TACTTool binary
	base    string // a folder with the install's .build.info and a link to its Data
	product string
	region  string

	mu      sync.Mutex
	waiting map[int][]chan error
	queued  []int
	failed  map[int]error
	timer   *time.Timer
}

// batchDelay is how long we wait for more files before a run. The page asks for a model's files
// one after another as it reads them, so a short wait gathers them into one run.
const batchDelay = 40 * time.Millisecond

func newAssetStore(home, tool string, c *wowClient) (*assetStore, error) {
	info, err := c.buildInfo()
	if err != nil {
		return nil, err
	}
	a := &assetStore{
		dir:     filepath.Join(home, "assets"),
		tool:    tool,
		base:    filepath.Join(home, "wow"),
		product: info["Product"],
		region:  info["Branch"],
		waiting: map[int][]chan error{},
		failed:  map[int]error{},
	}
	if _, err := os.Stat(tool); err != nil {
		return nil, fmt.Errorf("no TACTTool at %s", tool)
	}
	for _, dir := range []string{a.dir, a.base} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	// TACTTool looks through every folder under the install it's given, and the WoW folder holds
	// links it can't open. So we give it a folder of its own with the install's .build.info and a
	// link to the Data folder.
	buildInfo, err := os.ReadFile(filepath.Join(c.dataRoot, ".build.info"))
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(a.base, ".build.info"), buildInfo, 0o644); err != nil {
		return nil, err
	}
	link := filepath.Join(a.base, "Data")
	os.Remove(link)
	if err := os.Symlink(filepath.Join(c.dataRoot, "Data"), link); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *assetStore) path(fdid int) string {
	return filepath.Join(a.dir, strconv.Itoa(fdid))
}

func (a *assetStore) has(fdid int) bool {
	_, err := os.Stat(a.path(fdid))
	return err == nil
}

// get makes sure a file is extracted, waiting for its run when it isn't.
func (a *assetStore) get(fdid int) error {
	if a.has(fdid) {
		return nil
	}
	return <-a.queue(fdid)
}

// queue asks for a file and gives a channel that says when it's ready.
func (a *assetStore) queue(fdid int) chan error {
	done := make(chan error, 1)
	a.mu.Lock()
	defer a.mu.Unlock()
	if err, ok := a.failed[fdid]; ok {
		done <- err
		return done
	}
	if a.has(fdid) {
		done <- nil
		return done
	}
	if _, ok := a.waiting[fdid]; !ok {
		a.queued = append(a.queued, fdid)
	}
	a.waiting[fdid] = append(a.waiting[fdid], done)
	if a.timer == nil {
		a.timer = time.AfterFunc(batchDelay, a.run)
	}
	return done
}

// run extracts everything queued in one TACTTool run. Runs happen one at a time, and files queued
// during a run go in the next.
func (a *assetStore) run() {
	a.mu.Lock()
	ids := a.queued
	a.queued = nil
	a.mu.Unlock()

	results := a.extract(ids)

	a.mu.Lock()
	var follow []int
	for _, id := range ids {
		err := results[id]
		if err != nil {
			a.failed[id] = err
		} else {
			follow = append(follow, a.namedFiles(id)...)
		}
		for _, done := range a.waiting[id] {
			done <- err
		}
		delete(a.waiting, id)
	}
	a.timer = nil
	pending := len(a.queued) > 0
	a.mu.Unlock()

	for _, id := range follow {
		a.queue(id)
	}
	a.mu.Lock()
	if pending && a.timer == nil {
		a.timer = time.AfterFunc(0, a.run)
	}
	a.mu.Unlock()
}

func (a *assetStore) extract(ids []int) map[int]error {
	results := map[int]error{}
	if len(ids) == 0 {
		return results
	}
	out, err := os.MkdirTemp(a.dir, "run-")
	if err != nil {
		for _, id := range ids {
			results[id] = err
		}
		return results
	}
	defer os.RemoveAll(out)
	var list strings.Builder
	for _, id := range ids {
		fmt.Fprintf(&list, "%d;%d\n", id, id)
	}
	listPath := filepath.Join(out, "list.txt")
	if err := os.WriteFile(listPath, []byte(list.String()), 0o644); err != nil {
		for _, id := range ids {
			results[id] = err
		}
		return results
	}
	start := time.Now()
	cmd := exec.Command(a.tool, "-p", a.product, "-r", a.region, "-d", a.base, "-m", "list", "-i", listPath, "-o", out)
	// TACTTool reads its keys (WoW.txt) from the folder it runs in.
	cmd.Dir = filepath.Dir(a.tool)
	output, runErr := cmd.CombinedOutput()
	log.Printf("assets: extracted %d files in %v", len(ids), time.Since(start).Round(time.Millisecond))
	for _, id := range ids {
		src := filepath.Join(out, strconv.Itoa(id))
		data, err := os.ReadFile(src)
		switch {
		case err != nil && runErr != nil:
			results[id] = fmt.Errorf("TACTTool: %v: %s", runErr, lastLine(output))
		case err != nil:
			results[id] = fmt.Errorf("file %d is not in the client's storage", id)
		case allZero(data):
			results[id] = fmt.Errorf("file %d is encrypted with a key we don't have", id)
		default:
			results[id] = os.Rename(src, a.path(id))
		}
	}
	return results
}

func lastLine(b []byte) string {
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	return lines[len(lines)-1]
}

func allZero(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// namedFiles reads the files a model or skeleton names that the page loads right after it: a
// model's first skin, its textures and its skeleton, and a skeleton's parent.
func (a *assetStore) namedFiles(fdid int) []int {
	f, err := os.Open(a.path(fdid))
	if err != nil {
		return nil
	}
	defer f.Close()
	head := make([]byte, 4)
	if _, err := io.ReadFull(f, head); err != nil {
		return nil
	}
	magic := string(head)
	if magic != "MD21" && magic != "SKL1" {
		return nil
	}
	data, err := os.ReadFile(a.path(fdid))
	if err != nil {
		return nil
	}
	var out []int
	for pos := 0; pos+8 <= len(data); {
		id := string(data[pos : pos+4])
		size := int(binary.LittleEndian.Uint32(data[pos+4:]))
		body := data[pos+8 : min(len(data), pos+8+size)]
		switch id {
		case "SFID":
			if len(body) >= 4 {
				out = append(out, int(binary.LittleEndian.Uint32(body)))
			}
		case "TXID", "SKID":
			for i := 0; i+4 <= len(body); i += 4 {
				if v := int(binary.LittleEndian.Uint32(body[i:])); v != 0 {
					out = append(out, v)
				}
			}
		case "SKPD":
			if len(body) >= 12 {
				if v := int(binary.LittleEndian.Uint32(body[8:])); v != 0 {
					out = append(out, v)
				}
			}
		}
		pos += 8 + size
	}
	return out
}

// handleAsset serves one file, extracting it first when we don't have it yet.
func (s *server) handleAsset(w http.ResponseWriter, r *http.Request) {
	if s.assets == nil {
		writeError(w, http.StatusNotFound, errors.New("no WoW client or TACTTool"))
		return
	}
	fdid, err := strconv.Atoi(r.PathValue("fdid"))
	if err != nil || fdid <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("bad file id"))
		return
	}
	if err := s.assets.get(fdid); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Cache-Control", "max-age=604800, immutable")
	http.ServeFile(w, r, s.assets.path(fdid))
}

// handleAssetBatch extracts a list of files in one run, so the page can ask for everything it
// knows it needs (a character's whole list of files) before it reads them one by one. The reply
// names the files that failed.
func (s *server) handleAssetBatch(w http.ResponseWriter, r *http.Request) {
	if s.assets == nil {
		writeError(w, http.StatusNotFound, errors.New("no WoW client or TACTTool"))
		return
	}
	var ids []int
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&ids); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	chans := map[int]chan error{}
	for _, id := range ids {
		if id > 0 && !s.assets.has(id) {
			chans[id] = s.assets.queue(id)
		}
	}
	failed := map[string]string{}
	for id, done := range chans {
		if err := <-done; err != nil {
			failed[strconv.Itoa(id)] = err.Error()
		}
	}
	writeJSON(w, map[string]any{"failed": failed})
}
