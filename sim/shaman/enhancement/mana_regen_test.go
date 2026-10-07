package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanSpiritRegen checks that a naked Orc shaman regens what the Forever beta's
// GetManaRegen showed (2026-10-07), and that the Classic ruleset keeps the 1.12 shaman
// formula (15 + Spirit/5 every 2s tick).
func TestOrcShamanSpiritRegen(t *testing.T) {
	cases := []struct {
		ruleset proto.Ruleset
		level   int32
		regen   float64
	}{
		{proto.Ruleset_RulesetForever, 1, 5.5},
		{proto.Ruleset_RulesetForever, 30, 12.875},
		{proto.Ruleset_RulesetForever, 60, 19.125},
		{proto.Ruleset_RulesetClassic, 30, 12.8},
	}
	for _, c := range cases {
		player := &proto.Player{
			Race:      proto.Race_RaceOrc,
			Class:     proto.Class_ClassShaman,
			Level:     c.level,
			Equipment: &proto.EquipmentSpec{},
			Spec: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
				Options: &proto.EnhancementShaman_Options{},
			}},
		}
		raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
		env, _, _ := core.NewEnvironment(raid, &proto.Encounter{}, c.ruleset, false)
		character := env.Raid.Parties[0].Players[0].GetCharacter()
		if got := character.ManaRegenPerSecondWhileNotCasting(); math.Abs(got-c.regen) > 1e-9 {
			t.Errorf("%s level %d: got %.3f mana a second, want %.3f", c.ruleset, c.level, got, c.regen)
		}
		if got := character.ManaRegenPerSecondWhileCasting(); got != 0 {
			t.Errorf("%s level %d: got %.3f mana a second while casting, want 0", c.ruleset, c.level, got)
		}
	}
}

// TestOrcShamanDriftwoodIcon checks Polished Driftwood Icon against the Forever beta
// (2026-10-08). Its text says 8% of mana regen continues while casting.
//
// A geared level 30 Orc shaman with 83 Spirit and the icon read 16.626 and 1.331 from
// GetManaRegen. Taking off the 0.001 the beta adds to both, that's 16.625 out of combat and
// 8% of it while casting. Without gear the casting value went back to 0.001.
func TestOrcShamanDriftwoodIcon(t *testing.T) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotRanged] = &proto.ItemSpec{Id: 249398} // Polished Driftwood Icon

	player := &proto.Player{
		Race:       proto.Race_RaceOrc,
		Class:      proto.Class_ClassShaman,
		Level:      30,
		Equipment:  &proto.EquipmentSpec{Items: items},
		BonusStats: &proto.UnitStats{Stats: stats.Stats{stats.Spirit: 30}.ToFloatArray()},
		Spec: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
			Options: &proto.EnhancementShaman_Options{},
		}},
	}
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	env, _, _ := core.NewEnvironment(raid, &proto.Encounter{}, proto.Ruleset_RulesetForever, false)
	character := env.Raid.Parties[0].Players[0].GetCharacter()
	if got := character.ManaRegenPerSecondWhileNotCasting(); math.Abs(got-16.625) > 1e-9 {
		t.Errorf("got %.4f mana a second, want 16.625", got)
	}
	if got := character.ManaRegenPerSecondWhileCasting(); math.Abs(got-1.33) > 1e-9 {
		t.Errorf("got %.4f mana a second while casting, want 1.33", got)
	}
}
