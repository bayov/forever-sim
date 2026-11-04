//go:build practice

package main

import (
	"net/http"
	"strings"
	"sync"
)

// The client's texture atlases (UiTextureAtlas and UiTextureAtlasMember): the modern frames name
// their art by atlas ("UI-Frame-Metal-CornerTopLeft"), and an atlas is a region of a texture
// file. The room's Blizzard frames ask for the ones they draw.

// atlasInfo is one atlas: its file, its size as the frames lay it out, and its region of the
// file in pixels. Tile says whether it repeats across (1), down (2) or both (3).
type atlasInfo struct {
	File   int `json:"file"`
	Width  int `json:"width"`
	Height int `json:"height"`
	Left   int `json:"left"`
	Right  int `json:"right"`
	Top    int `json:"top"`
	Bottom int `json:"bottom"`
	FileW  int `json:"fileWidth"`
	FileH  int `json:"fileHeight"`
	Tile   int `json:"tile,omitempty"`
}

type atlasIndex struct {
	once   sync.Once
	byName map[string]atlasInfo
}

func (s *server) atlases() map[string]atlasInfo {
	s.atlasIndex.once.Do(func() {
		s.atlasIndex.byName = map[string]atlasInfo{}
		if s.db2 == nil {
			return
		}
		files, err := s.db2.Table("UiTextureAtlas")
		if err != nil {
			return
		}
		members, err := s.db2.Table("UiTextureAtlasMember")
		if err != nil {
			return
		}
		elements, err := s.db2.Table("UiTextureAtlasElement")
		if err != nil {
			return
		}
		info := func(m db2Row) (atlasInfo, bool) {
			f := files.Get(m.Int("UiTextureAtlasID"))
			if !f.Ok() {
				return atlasInfo{}, false
			}
			a := atlasInfo{
				File:   f.Int("FileDataID"),
				Width:  m.Int("Width"),
				Height: m.Int("Height"),
				Left:   m.Int("CommittedLeft"),
				Right:  m.Int("CommittedRight"),
				Top:    m.Int("CommittedTop"),
				Bottom: m.Int("CommittedBottom"),
				FileW:  f.Int("AtlasWidth"),
				FileH:  f.Int("AtlasHeight"),
			}
			// Art made at twice the density ("-2x") lays out at half its pixels, unless the
			// member gives its size.
			if strings.HasSuffix(strings.ToLower(m.Str("CommittedName")), "-2x") {
				a.Width, a.Height = a.Width/2, a.Height/2
			}
			if w := m.Int("OverrideWidth"); w > 0 {
				a.Width = w
			}
			if h := m.Int("OverrideHeight"); h > 0 {
				a.Height = h
			}
			return a, true
		}
		// The frames name the element ("UI-Frame-Metal-CornerTopLeft"). Its art may sit in a
		// member of another name ("ui-frame-metal-cornertopleft-2x"). A name that starts with
		// "_" tiles across and one with "!" tiles down, like the game's.
		for i := 0; i < elements.Len(); i++ {
			e := elements.Row(i)
			ms := members.Where("UiTextureAtlasElementID", e.Int("ID"))
			if len(ms) == 0 {
				continue
			}
			a, ok := info(ms[len(ms)-1])
			if !ok {
				continue
			}
			name := e.Str("Name")
			if strings.HasPrefix(name, "_") {
				a.Tile |= 1
			}
			if strings.HasPrefix(name, "!") {
				a.Tile |= 2
			}
			s.atlasIndex.byName[strings.ToLower(name)] = a
		}
		for i := 0; i < members.Len(); i++ {
			m := members.Row(i)
			name := strings.ToLower(m.Str("CommittedName"))
			if _, ok := s.atlasIndex.byName[name]; ok {
				continue
			}
			if a, ok := info(m); ok {
				s.atlasIndex.byName[name] = a
			}
		}
	})
	return s.atlasIndex.byName
}

// handleAtlas gives the atlases named in ?names= (comma separated, any case). Unknown names
// are left out.
func (s *server) handleAtlas(w http.ResponseWriter, r *http.Request) {
	all := s.atlases()
	out := map[string]atlasInfo{}
	for _, name := range strings.Split(r.URL.Query().Get("names"), ",") {
		if a, ok := all[strings.ToLower(strings.TrimSpace(name))]; ok {
			out[name] = a
		}
	}
	w.Header().Set("Cache-Control", "max-age=86400")
	writeJSON(w, out)
}
