package core

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestWeaponSkillGap checks the attack table of a level 30 with a fist weapon at full skill and
// at the beta's Unarmed 85 of 150 (10-10), against mobs from 2 levels below to 3 above.
func TestWeaponSkillGap(t *testing.T) {
	fist := &Item{WeaponType: proto.WeaponType_WeaponTypeFist, HandType: proto.HandType_HandTypeOneHand}
	cases := []struct {
		mobLevel int32
		gap      float64
		// miss, dodge, parry, glance and crit suppression, in percent
		miss, dodge, parry, glance, critSuppression float64
	}{
		// Full skill is the table we had before the option.
		{30, 0, 5, 5, 5, 10, 0},
		{33, 0, 8, 6.5, 14, 40, 4.8},
		{28, 0, 4, 4, 4, 0, -0.4},
		// 65 points short against a mob of our level: every chance goes by 85 skill.
		{30, 65, 18, 11.5, 44, 40, 2.6},
		// Against a mob 2 levels below, 85 skill is still 55 short of its 140 defense.
		{28, 65, 16, 10.5, 38, 40, 2.2},
		// 5 points short: 5% more glances, 0.5% more of each other.
		{30, 5, 5.5, 5.5, 5.5, 20, 0.2},
	}
	for _, c := range cases {
		attacker := &Unit{Type: PlayerUnit, Level: 30, stats: stats.Stats{}}
		defender := &Unit{Type: EnemyUnit, Level: c.mobLevel, stats: stats.Stats{}}
		table := newWeaponAttackTable(attacker, defender, fist, c.gap)
		got := []float64{table.BaseMissChance, table.BaseDodgeChance, table.BaseParryChance, table.BaseGlanceChance, table.MeleeCritSuppression}
		want := []float64{c.miss, c.dodge, c.parry, c.glance, c.critSuppression}
		names := []string{"miss", "dodge", "parry", "glance", "crit suppression"}
		for i := range got {
			if math.Abs(100*got[i]-want[i]) > 1e-9 {
				t.Errorf("level %d mob, %.0f short: %s %.2f%%, want %.2f%%", c.mobLevel, c.gap, names[i], 100*got[i], want[i])
			}
		}
	}
}
