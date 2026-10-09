package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanEnemyAvoidance checks how often a mob dodges, parries and blocks our melee
// attacks, by its level, our weapon skill and our expertise.
//
// Dodge is 5% plus 0.1% for each point the mob's defense (its level times 5) is above our
// weapon skill, and skill from gear counts. Parry is 5% plus 0.1% a point of defense over
// our level times 5, or 0.6% a point when that gap is more than 10, so skill from gear
// doesn't lower it. Block is 5%. Blizzard gave 6.5% dodge and 14% parry for a mob 3 levels
// up in 2019 Classic, and magey's tests there fit the rest. Parry and block only happen
// from the front (TestOrcShamanAttackTable).
//
// Forever's expertise-like stat takes 1% off dodge and parry for each 1% of it. Dwarven
// Tree Chopper carries 0.6% of it under Forever, in place of Classic's 2 skill.
//
// We roll the white and the yellow table from the front, where all three can happen.
func TestOrcShamanEnemyAvoidance(t *testing.T) {
	cases := []struct {
		name               string
		level, targetLevel int32
		weapon             int32
		skill, expertise   float64
		dodge, parry       float64
	}{
		{"level 60 against a level 63 boss", 60, 63, darkEdge, 0, 0, 0.065, 0.14},
		{"310 skill against the boss", 60, 63, darkEdge, 10, 0, 0.055, 0.14},
		{"2% expertise against the boss", 60, 63, darkEdge, 0, 2, 0.045, 0.12},
		{"level 30 against Vishas (level 32)", 30, 32, whirlwind, 0, 0, 0.06, 0.06},
		{"152 skill and 0.6% expertise against Vishas", 30, 32, whirlwind, 2, 0.6, 0.052, 0.054},
		{"Dwarven Tree Chopper's 0.6% expertise against Vishas", 30, 32, treeChop, 0, 0, 0.054, 0.054},
		{"a level 31", 30, 31, whirlwind, 0, 0, 0.055, 0.055},
		{"the same level", 30, 30, whirlwind, 0, 0, 0.05, 0.05},
		{"5 levels below us", 30, 25, whirlwind, 0, 0, 0.025, 0.025},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim, enh := newTargetLevelSim(c.level, c.targetLevel, c.weapon, c.skill, stormstrike)
			enh.AddStatDynamic(sim, stats.Expertise, c.expertise)
			enh.PseudoStats.InFrontOfTarget = true

			// We roll each table many times from the front and count what comes out.
			const rolls = 200000
			white := enh.AutoAttacks.MHAuto()
			for _, attack := range []struct {
				name    string
				spell   *core.Spell
				outcome core.OutcomeApplier
			}{
				{"white", white, white.OutcomeMeleeWhite},
				{"Stormstrike", enh.Stormstrike, enh.Stormstrike.OutcomeMeleeWeaponSpecialHitAndCrit},
			} {
				var counts outcomeCounts
				for range rolls {
					counts.add(attack.spell.CalcDamage(sim, enh.CurrentTarget, 1, attack.outcome).Outcome)
				}
				for _, check := range []struct {
					outcome string
					got     int
					want    float64
				}{{"dodge", counts.dodge, c.dodge}, {"parry", counts.parry, c.parry}, {"block", counts.block, 0.05}} {
					rate := float64(check.got) / rolls
					if se := math.Sqrt(check.want * (1 - check.want) / rolls); math.Abs(rate-check.want) > 4*se {
						t.Errorf("%s: %.4f %s, want %.4f +- %.4f", attack.name, rate, check.outcome, check.want, 4*se)
					}
				}
			}
		})
	}
}
