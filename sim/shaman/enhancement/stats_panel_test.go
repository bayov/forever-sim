package enhancement

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanStatsPanel checks that the stats panel leaves out the totems we put down
// ourselves.
//
// A level 60 shaman knows Strength of Earth (53 Strength) and Grace of Air (89 Agility),
// but the fight only gets them once the rotation casts them. When another shaman's
// Strength of Earth is picked in the raid buffs, the panel should still show it.
func TestOrcShamanStatsPanel(t *testing.T) {
	finalStats := func(raidBuffs *proto.RaidBuffs) stats.Stats {
		player := &proto.Player{
			Race:      proto.Race_RaceOrc,
			Class:     proto.Class_ClassShaman,
			Level:     60,
			Equipment: &proto.EquipmentSpec{},
			Spec: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
				Options: &proto.EnhancementShaman_Options{},
			}},
		}
		result := core.ComputeStats(&proto.ComputeStatsRequest{
			Raid:    core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, raidBuffs, &proto.Debuffs{}),
			Ruleset: proto.Ruleset_RulesetForever,
		})
		return stats.FromFloatArray(result.RaidStats.Parties[0].Players[0].FinalStats.Stats)
	}
	check := func(name string, have, want float64) {
		if have != want {
			t.Errorf("%s: got %.1f, want %.1f", name, have, want)
		}
	}

	own := finalStats(&proto.RaidBuffs{})
	check("Strength with our own totems", own[stats.Strength], 88)
	check("Agility with our own totems", own[stats.Agility], 52)

	raid := finalStats(&proto.RaidBuffs{StrengthOfEarthTotem: proto.TristateEffect_TristateEffectRegular})
	check("Strength with a raid Strength of Earth", raid[stats.Strength], 88+53)
	check("Agility with a raid Strength of Earth", raid[stats.Agility], 52)
}
