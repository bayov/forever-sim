//go:build practice

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// Spell and item icons and names, looked up on Wowhead's tooltip service once and kept in
// ~/forever-practice/icons.json. An id Wowhead doesn't know gets the question mark, without an
// error in the page.

type iconInfo struct {
	Icon string `json:"icon"`
	Name string `json:"name,omitempty"`
}

type iconCache struct {
	mu      sync.Mutex
	path    string
	entries map[string]iconInfo
	client  *http.Client
}

func newIconCache(path string) *iconCache {
	c := &iconCache{path: path, entries: map[string]iconInfo{}, client: &http.Client{Timeout: 8 * time.Second}}
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &c.entries)
	}
	return c
}

func (c *iconCache) get(kind string, id int) iconInfo {
	key := fmt.Sprintf("%s:%d", kind, id)
	c.mu.Lock()
	info, ok := c.entries[key]
	c.mu.Unlock()
	if ok {
		return info
	}
	info = iconInfo{Icon: "inv_misc_questionmark"}
	found := false
	for _, domain := range []string{"classic", "wotlk"} {
		req, _ := http.NewRequest("GET", fmt.Sprintf("https://nether.wowhead.com/%s/tooltip/%s/%d", domain, kind, id), nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36")
		resp, err := c.client.Do(req)
		if err != nil {
			// Offline: don't cache, try again next time.
			return info
		}
		var body struct {
			Icon string `json:"icon"`
			Name string `json:"name"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if err == nil && body.Icon != "" {
			info = iconInfo{Icon: body.Icon, Name: body.Name}
			found = true
			break
		}
	}
	_ = found
	c.mu.Lock()
	c.entries[key] = info
	data, _ := json.Marshal(c.entries)
	c.mu.Unlock()
	os.WriteFile(c.path, data, 0o644)
	return info
}

func (s *server) handleIcon(w http.ResponseWriter, r *http.Request) {
	kind := r.PathValue("kind")
	id, err := strconv.Atoi(r.PathValue("id"))
	if (kind != "spell" && kind != "item") || err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("bad icon request"))
		return
	}
	info := s.icons.get(kind, id)
	w.Header().Set("Cache-Control", "max-age=86400")
	writeJSON(w, info)
}

var _ = filepath.Join
