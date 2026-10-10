package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
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

// TestOrcShamanOurAvoidance checks how often a mob's swing misses us, or we dodge, parry or
// block it.
//
// Miss is 5%, dodge is our sheet dodge and parry is 5% with Spirit Weapons. Each is 0.2%
// lower a level the mob is above us. We have no shield, so we never block. Anticipation adds
// 2% dodge a point, as its Forever text says. While we cast, only a miss stops a swing. The
// beta fits miss and dodge against level 25 to 35 mobs (shaman_audit.md 6.4). We roll a level
// 34 mob's swing 20,000 times with Anticipation 3/3 and Spirit Weapons, and again while we
// cast.
func TestOrcShamanOurAvoidance(t *testing.T) {
	_, plain := newTankSim(newOrcShaman(30, spiritWeapons, &proto.EnhancementShaman_Options{}), 34)
	// Anticipation 3/3 and Spirit Weapons.
	sim, enh := newTankSim(newOrcShaman(30, "-00000000030001", &proto.EnhancementShaman_Options{}), 34)
	if got := enh.GetStat(stats.Dodge) - plain.GetStat(stats.Dodge); math.Abs(got-6) > 1e-9 {
		t.Errorf("Anticipation 3/3 added %.2f%% dodge, want 6%%", got)
	}

	const rolls = 20000
	swing := enh.CurrentTarget.AutoAttacks.MHAuto()
	roll := func(sim *core.Simulation) map[core.HitOutcome]float64 {
		counts := map[core.HitOutcome]float64{}
		for range rolls {
			result := swing.CalcDamage(sim, &enh.Unit, 100, swing.OutcomeEnemyMeleeWhite)
			counts[result.Outcome&(core.OutcomeMiss|core.OutcomeDodge|core.OutcomeParry|core.OutcomeBlock)]++
		}
		return counts
	}
	check := func(name string, counts map[core.HitOutcome]float64, miss, dodge, parry float64) {
		for _, c := range []struct {
			outcome core.HitOutcome
			label   string
			want    float64
		}{{core.OutcomeMiss, "miss", miss}, {core.OutcomeDodge, "dodge", dodge}, {core.OutcomeParry, "parry", parry}, {core.OutcomeBlock, "block", 0}} {
			if got := counts[c.outcome] / rolls; math.Abs(got-c.want) > 0.006 {
				t.Errorf("%s: %s %.2f%%, want %.2f%%", name, c.label, 100*got, 100*c.want)
			}
		}
	}

	var free, casting map[core.HitOutcome]float64
	at(sim, 1, func(sim *core.Simulation) {
		free = roll(sim)
		enh.Hardcast.Expires = sim.CurrentTime + time.Second
		casting = roll(sim)
		enh.Hardcast.Expires = sim.CurrentTime
	})
	runSim(sim)

	dodge := enh.GetStat(stats.Dodge) / 100
	check("not casting", free, 0.042, dodge-0.008, 0.042)
	check("casting", casting, 0.042, 0, 0)
}
