package enhancement

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanRaidBuffRanks checks that the raid buffs take Forever's rank for our level
// (shaman_audit.md 10.2).
//
// The values are from client 70338. Devotion Aura gives 390 armor at 30 and 735 at 60, with
// no Improved Devotion Aura. Blood Pact grows a little a level, so it gives 22 Stamina at 30
// (rank 3, 21 plus 0.4 a level from 26) and 54 at 60 (rank 5, 49 plus 0.5 a level from 50).
// Arcane Intellect gives 15 Intellect at 30 and 31 at 60. Battle Shout gives 35 attack power
// at 30 (rank 3, 33 plus 0.3 a level from 22) and 115 at 60.
func TestOrcShamanRaidBuffRanks(t *testing.T) {
	improved := proto.TristateEffect_TristateEffectImproved
	newCharacter := func(level int32, raidBuffs *proto.RaidBuffs) *core.Character {
		player := &proto.Player{
			Race:      proto.Race_RaceOrc,
			Class:     proto.Class_ClassShaman,
			Level:     level,
			Equipment: &proto.EquipmentSpec{},
			Spec: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
				Options: &proto.EnhancementShaman_Options{},
			}},
		}
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, raidBuffs, &proto.Debuffs{}),
			Encounter:  &proto.Encounter{},
			SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
		}, simsignals.CreateSignals())
		// Devotion Aura and Battle Shout go on when the fight starts.
		sim.Reset()
		return sim.Raid.Parties[0].Players[0].GetCharacter()
	}
	cases := []struct {
		level                                int32
		armor, stamina, intellect, attackPow float64
	}{
		{30, 390, 22, 15, 35},
		{60, 735, 54, 31, 115},
	}
	for _, c := range cases {
		plain := newCharacter(c.level, &proto.RaidBuffs{})
		buffed := newCharacter(c.level, &proto.RaidBuffs{
			DevotionAura:     improved,
			BloodPact:        improved,
			ArcaneBrilliance: true,
			BattleShout:      improved,
		})
		for _, check := range []struct {
			name string
			stat stats.Stat
			want float64
		}{
			{"Devotion Aura armor", stats.BonusArmor, c.armor},
			{"Blood Pact Stamina", stats.Stamina, c.stamina},
			{"Arcane Intellect", stats.Intellect, c.intellect},
			{"Battle Shout attack power", stats.AttackPower, c.attackPow},
		} {
			if got := buffed.GetStat(check.stat) - plain.GetStat(check.stat); got != check.want {
				t.Errorf("level %d: %s %.2f, want %.2f", c.level, check.name, got, check.want)
			}
		}
	}
}
