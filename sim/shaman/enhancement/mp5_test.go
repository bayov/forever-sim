package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanMP5Buffs checks Blessing of Wisdom, Blessing of Might and another shaman's Mana
// Spring Totem by level.
//
// Under Forever all three reach a Horde shaman, and Wisdom and Mana Spring stack (the user,
// 2026-10-08). Each takes the top rank the level knows: Blessing of Wisdom 12 MP5 at 14, 18 at
// 24 and 40 at 60, and Mana Spring 10 MP5 at 26 and 25 at 56. Blessing of Might gives 25 melee
// attack power at 20, 40 at 30 (rank 4 needs level 32) and 133 at 60, and no ranged attack
// power (client 70338, aura 99, shaman_audit.md 10.1). Under Classic the Horde shaman gets no
// blessing.
func TestOrcShamanMP5Buffs(t *testing.T) {
	regular, improved := proto.TristateEffect_TristateEffectRegular, proto.TristateEffect_TristateEffectImproved
	cases := []struct {
		ruleset    proto.Ruleset
		level      int32
		manaSpring proto.TristateEffect
		mp5        float64
		might      float64
	}{
		{proto.Ruleset_RulesetForever, 20, regular, 12, 25},
		{proto.Ruleset_RulesetForever, 30, regular, 18 + 10, 40},
		{proto.Ruleset_RulesetForever, 30, improved, 18 + 12.5, 40},
		{proto.Ruleset_RulesetForever, 60, regular, 40 + 25, 133},
		{proto.Ruleset_RulesetForever, 60, improved, 40 + 31.25, 133},
		{proto.Ruleset_RulesetClassic, 60, improved, 31.25, 0},
	}
	for _, c := range cases {
		newCharacter := func(might proto.TristateEffect) *core.Character {
			player := &proto.Player{
				Race:      proto.Race_RaceOrc,
				Class:     proto.Class_ClassShaman,
				Level:     c.level,
				Equipment: &proto.EquipmentSpec{},
				Buffs:     &proto.IndividualBuffs{BlessingOfWisdom: regular, BlessingOfMight: might},
				Spec: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
					Options: &proto.EnhancementShaman_Options{},
				}},
			}
			raidBuffs := &proto.RaidBuffs{ManaSpringTotem: c.manaSpring}
			raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, raidBuffs, &proto.Debuffs{})
			sim := core.NewSim(&proto.RaidSimRequest{
				Raid:       raid,
				Encounter:  &proto.Encounter{},
				SimOptions: &proto.SimOptions{Ruleset: c.ruleset, RandomSeed: 1},
			}, simsignals.CreateSignals())
			// The blessing's attack power goes on when the fight starts.
			sim.Reset()
			return sim.Raid.Parties[0].Players[0].GetCharacter()
		}
		plain, character := newCharacter(proto.TristateEffect_TristateEffectMissing), newCharacter(regular)
		if got := character.GetStat(stats.AttackPower) - plain.GetStat(stats.AttackPower); got != c.might {
			t.Errorf("%s level %d: Blessing of Might gave %.2f attack power, want %.2f", c.ruleset, c.level, got, c.might)
		}
		if got := character.GetStat(stats.RangedAttackPower) - plain.GetStat(stats.RangedAttackPower); got != 0 {
			t.Errorf("%s level %d: Blessing of Might gave %.2f ranged attack power, want 0", c.ruleset, c.level, got)
		}
		if got := character.GetStat(stats.MP5); got != c.mp5 {
			t.Errorf("%s level %d, Mana Spring %s: got %.2f MP5, want %.2f", c.ruleset, c.level, c.manaSpring, got, c.mp5)
		}
		// MP5 keeps going in full while casting.
		if got := character.ManaRegenPerSecondWhileCasting(); math.Abs(got-c.mp5/5) > 1e-9 {
			t.Errorf("%s level %d, Mana Spring %s: got %.3f mana a second while casting, want %.3f", c.ruleset, c.level, c.manaSpring, got, c.mp5/5)
		}
	}
}
