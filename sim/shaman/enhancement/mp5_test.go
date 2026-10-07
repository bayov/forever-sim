package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanMP5Buffs checks Blessing of Wisdom and another shaman's Mana Spring Totem by
// level.
//
// Under Forever both reach a Horde shaman and they stack (the user, 2026-10-08). Each takes
// the top rank the level knows: Blessing of Wisdom 12 MP5 at 14, 18 at 24 and 40 at 60, and
// Mana Spring 10 MP5 at 26 and 25 at 56. Under Classic the Horde shaman gets no Blessing of
// Wisdom.
func TestOrcShamanMP5Buffs(t *testing.T) {
	regular, improved := proto.TristateEffect_TristateEffectRegular, proto.TristateEffect_TristateEffectImproved
	cases := []struct {
		ruleset    proto.Ruleset
		level      int32
		manaSpring proto.TristateEffect
		mp5        float64
	}{
		{proto.Ruleset_RulesetForever, 20, regular, 12},
		{proto.Ruleset_RulesetForever, 30, regular, 18 + 10},
		{proto.Ruleset_RulesetForever, 30, improved, 18 + 12.5},
		{proto.Ruleset_RulesetForever, 60, regular, 40 + 25},
		{proto.Ruleset_RulesetForever, 60, improved, 40 + 31.25},
		{proto.Ruleset_RulesetClassic, 60, improved, 31.25},
	}
	for _, c := range cases {
		player := &proto.Player{
			Race:      proto.Race_RaceOrc,
			Class:     proto.Class_ClassShaman,
			Level:     c.level,
			Equipment: &proto.EquipmentSpec{},
			Buffs:     &proto.IndividualBuffs{BlessingOfWisdom: regular},
			Spec: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
				Options: &proto.EnhancementShaman_Options{},
			}},
		}
		raidBuffs := &proto.RaidBuffs{ManaSpringTotem: c.manaSpring}
		raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, raidBuffs, &proto.Debuffs{})
		env, _, _ := core.NewEnvironment(raid, &proto.Encounter{}, c.ruleset, false)
		character := env.Raid.Parties[0].Players[0].GetCharacter()
		if got := character.GetStat(stats.MP5); got != c.mp5 {
			t.Errorf("%s level %d, Mana Spring %s: got %.2f MP5, want %.2f", c.ruleset, c.level, c.manaSpring, got, c.mp5)
		}
		// MP5 keeps going in full while casting.
		if got := character.ManaRegenPerSecondWhileCasting(); math.Abs(got-c.mp5/5) > 1e-9 {
			t.Errorf("%s level %d, Mana Spring %s: got %.3f mana a second while casting, want %.3f", c.ruleset, c.level, c.manaSpring, got, c.mp5/5)
		}
	}
}
