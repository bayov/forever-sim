package enhancement

import (
	"fmt"
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanMindfulness checks how much Spirit regen Mindfulness keeps while we cast
// (shaman_audit.md 8.37).
//
// 1, 2 and 3 points keep 17, 33 and 50%, as the client says (1223033). A naked level 60 shaman
// has no MP5, so the casting regen is that share of the full regen.
func TestOrcShamanMindfulness(t *testing.T) {
	for points, share := range []float64{0, .17, .33, .50} {
		// Mindfulness is the third Restoration talent.
		_, enh := newShamanSim(newOrcShaman(60, fmt.Sprintf("--00%d", points), &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 2)
		if enh.Talents.Mindfulness != int32(points) {
			t.Fatalf("got Mindfulness %d, want %d", enh.Talents.Mindfulness, points)
		}
		full := enh.ManaRegenPerSecondWhileNotCasting()
		if got := enh.ManaRegenPerSecondWhileCasting(); full <= 0 || math.Abs(got-share*full) > 1e-9 {
			t.Errorf("Mindfulness %d: %.4f mana a second while casting, want %.4f (%.0f%% of %.4f)", points, got, share*full, share*100, full)
		}
	}
}
