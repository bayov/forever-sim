package holy

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// A healing paladin never bubbles the tank (user ruling 2026-11-04, standards 40: "make sure paladin
// never bubbles tank, it causes tank to lose aggro"). Blessing of Protection and Divine Intervention
// on the tank drop the boss off it, which the sim doesn't model, so no Holy paladin rotation or
// healbench request may cast them. The sim registers neither today; this keeps it that way when one
// is added. The paladin's own Divine Shield and Divine Protection are listed too: a healer has no use
// for them in these fights, and the tank side's test (protection/no_bubble_test.go) bans them there.
var bubbleSpellIDs = map[string]string{
	"1022": "Blessing of Protection", "5599": "Blessing of Protection", "10278": "Blessing of Protection",
	"19752": "Divine Intervention",
	"642":   "Divine Shield", "1020": "Divine Shield",
	"498": "Divine Protection", "5573": "Divine Protection",
}

func TestForeverHolyPaladinNeverBubblesTank(t *testing.T) {
	var files []string
	for _, glob := range []string{
		"../../../ui/holy_paladin/apls/*.json",
		"../../../tools/healbench/req/paladin_*.json",
	} {
		matches, err := filepath.Glob(glob)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, matches...)
	}
	if len(files) < 4 {
		t.Fatalf("found only %d Holy paladin rotations and requests: %v", len(files), files)
	}
	spellID := regexp.MustCompile(`"spellId"\s*:\s*(\d+)`)
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range spellID.FindAllSubmatch(data, -1) {
			if name, ok := bubbleSpellIDs[string(m[1])]; ok {
				t.Errorf("%s casts %s (%s): a paladin never bubbles the tank", f, name, m[1])
			}
		}
	}
}
