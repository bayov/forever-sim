//go:build practice

package practice

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Every DPS Report build (every class at 60) builds a session and casts its whole toolkit, items
// included. The enh shaman builds must never fail. Other specs' toolkits hold spells the sim
// only models for another spec (a balance druid's Cat Form), and those end the attempt with an
// error, which the test lists.
func TestEveryClassCastsItsToolkit(t *testing.T) {
	paths, _ := filepath.Glob("../../tools/dps_report/req/*.json")
	if len(paths) == 0 {
		t.Skip("no DPS Report requests")
	}
	for _, path := range paths {
		if strings.HasSuffix(path, ".meta.json") {
			continue
		}
		name := strings.TrimSuffix(filepath.Base(path), ".json")
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			request, err := ParseRequest(data)
			if err != nil {
				t.Fatal(err)
			}
			request, err = Normalize(request)
			if err != nil {
				t.Fatal(err)
			}
			s, err := NewSession(request, Settings{}, 7)
			if err != nil {
				t.Fatal(err)
			}
			d := newDriver(t, s)
			d.standAt(0, -4)
			toolkit := s.Toolkit()
			if len(toolkit) < 3 {
				t.Fatalf("only %d spells in the toolkit", len(toolkit))
			}
			for _, spell := range toolkit {
				in := Input{Kind: "cast", SpellID: spell.SpellID, Target: &UnitRef{Kind: "enemy"}}
				if spell.ItemID != 0 {
					in = Input{Kind: "useItem", ItemID: spell.ItemID}
				}
				d.press(in)
				d.wait(1600 * time.Millisecond)
			}
			d.wait(5 * time.Second)
			if s.Failed != "" {
				if strings.Contains(name, "shaman__enhancement") {
					t.Fatalf("the sim failed: %s", s.Failed)
				}
				t.Logf("the sim failed (a gap outside the enh shaman): %s", s.Failed)
				return
			}
			if s.pulled {
				s.Stop(d.now)
				if s.Summary() == nil {
					t.Fatal("no summary")
				}
			}
		})
	}
}
