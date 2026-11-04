package protection

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// A paladin tank never bubbles (user ruling 2026-11-04, standards 40: "make sure paladin doesn't
// bubble, it causes pally to lose aggro"). Divine Shield, Divine Protection and Blessing of
// Protection on itself drop the boss off the paladin, which the sim doesn't model, so no paladin
// tank rotation may cast them. This fails when one of the rotations or tankbench requests does.
var bubbleSpellIDs = map[string]string{
	"642": "Divine Shield", "1020": "Divine Shield",
	"498": "Divine Protection", "5573": "Divine Protection",
	"1022": "Blessing of Protection", "5599": "Blessing of Protection", "10278": "Blessing of Protection",
}

func TestForeverPaladinTankNeverBubbles(t *testing.T) {
	var files []string
	for _, glob := range []string{
		"../../../ui/protection_paladin/apls/*.json",
		"../../../tools/tankbench/req/paladin_*.json",
	} {
		matches, err := filepath.Glob(glob)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, matches...)
	}
	if len(files) < 4 {
		t.Fatalf("found only %d paladin tank rotations and requests: %v", len(files), files)
	}
	spellID := regexp.MustCompile(`"spellId"\s*:\s*(\d+)`)
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range spellID.FindAllSubmatch(data, -1) {
			if name, ok := bubbleSpellIDs[string(m[1])]; ok {
				t.Errorf("%s casts %s (%s): a paladin tank never bubbles", f, name, m[1])
			}
		}
	}
}
