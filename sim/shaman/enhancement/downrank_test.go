package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanDownrankCoefficient checks Forever's downranking penalty on Lightning Bolt and
// Chain Lightning at level 60: a rank keeps 1 - 5% for each level past its learn level + 17
// of its spell power coefficient (Blizzard's 2026-10-01 development notes, the formula from
// https://foreverchanges.pro/downranking). It fails when a downranked rank keeps its full
// coefficient, or when a rank close to the level loses some.
func TestOrcShamanDownrankCoefficient(t *testing.T) {
	_, enh := newShamanSim(newOrcShaman(60, "", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 10)

	tests := []struct {
		name string
		got  float64
		want float64
	}{
		{"Lightning Bolt rank 10 (56)", enh.LightningBolt[10].BonusCoefficient, 0.714},
		{"Lightning Bolt rank 9 (50)", enh.LightningBolt[9].BonusCoefficient, 0.714},
		{"Lightning Bolt rank 8 (44)", enh.LightningBolt[8].BonusCoefficient, 0.714},
		{"Lightning Bolt rank 7 (38)", enh.LightningBolt[7].BonusCoefficient, 0.714 * 0.75},
		{"Lightning Bolt rank 6 (32)", enh.LightningBolt[6].BonusCoefficient, 0.714 * 0.45},
		{"Lightning Bolt rank 1 (1)", enh.LightningBolt[1].BonusCoefficient, 0},
		{"Chain Lightning rank 4 (56)", enh.ChainLightning[4].BonusCoefficient, shaman.ChainLightningSpellCoef[4]},
		{"Chain Lightning rank 1 (32)", enh.ChainLightning[1].BonusCoefficient, shaman.ChainLightningSpellCoef[1] * 0.45},
	}
	for _, tc := range tests {
		if math.Abs(tc.got-tc.want) > 1e-9 {
			t.Errorf("%s: coefficient %.4f, want %.4f", tc.name, tc.got, tc.want)
		}
	}
}
