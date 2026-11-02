//go:build practice

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// Spell and item icons, names and tooltips, looked up on Wowhead's tooltip service once and kept
// in ~/forever-practice/icons.json. Wowhead's Forever data comes first, then Classic's. An id
// Wowhead doesn't know gets the question mark, without an error in the page.

type iconInfo struct {
	Icon string `json:"icon"`
	Name string `json:"name,omitempty"`
	// Tooltip is Wowhead's tooltip HTML, Buff the aura's. Fetched says we asked Wowhead for them,
	// since entries from before tooltips have neither.
	Tooltip string `json:"tooltip,omitempty"`
	Buff    string `json:"buff,omitempty"`
	Fetched bool   `json:"fetched,omitempty"`
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
	if ok && info.Fetched {
		return info
	}
	fetched, online := c.fetch(kind, id)
	if !online {
		// Offline: don't cache, try again next time.
		if ok {
			return info
		}
		return iconInfo{Icon: "inv_misc_questionmark"}
	}
	c.mu.Lock()
	c.entries[key] = fetched
	data, _ := json.Marshal(c.entries)
	c.mu.Unlock()
	os.WriteFile(c.path, data, 0o644)
	return fetched
}

func (c *iconCache) fetch(kind string, id int) (iconInfo, bool) {
	info := iconInfo{Icon: "inv_misc_questionmark", Fetched: true}
	urls := []string{
		fmt.Sprintf("https://nether.wowhead.com/forever/tooltip/%s/%d?dataEnv=17&locale=0", kind, id),
		fmt.Sprintf("https://nether.wowhead.com/classic/tooltip/%s/%d", kind, id),
		fmt.Sprintf("https://nether.wowhead.com/wotlk/tooltip/%s/%d", kind, id),
	}
	for _, url := range urls {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36")
		resp, err := c.client.Do(req)
		if err != nil {
			return info, false
		}
		var body struct {
			Icon    string `json:"icon"`
			Name    string `json:"name"`
			Tooltip string `json:"tooltip"`
			Buff    string `json:"buff"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if err == nil && body.Icon != "" {
			return iconInfo{Icon: body.Icon, Name: body.Name, Tooltip: body.Tooltip, Buff: body.Buff, Fetched: true}, true
		}
	}
	return info, true
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
	writeJSON(w, iconInfo{Icon: info.Icon, Name: info.Name})
}

// handleTooltip gives Wowhead's tooltip HTML for a spell or item, and the aura's.
func (s *server) handleTooltip(w http.ResponseWriter, r *http.Request) {
	kind := r.PathValue("kind")
	id, err := strconv.Atoi(r.PathValue("id"))
	if (kind != "spell" && kind != "item") || err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("bad tooltip request"))
		return
	}
	w.Header().Set("Cache-Control", "max-age=86400")
	writeJSON(w, s.icons.get(kind, id))
}

