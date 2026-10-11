package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanStoneskinTotem checks how much Stoneskin Totem takes off the hits we take
// (shaman_audit.md 7.29).
//
// Under Forever each rank takes its own flat amount off each physical hit, before armor: 11
// for rank 3 at level 30 and 30 for rank 6 at 60. On the beta our rank 3 totem's 11 fit the
// hits we took in two logs. With 572 armor it fit only before armor. Guardian Totems adds
// 10% a point, so 2 points make it 13.2. Another shaman's totem in the raid buffs gives the top
// rank at our level too. A hit smaller than the reduction deals nothing, and the reduction
// goes when a Tremor Totem takes the earth slot. We wear 512 armor, which makes 572 at level 30
// with our Agility, as in the first log.
func TestOrcShamanStoneskinTotem(t *testing.T) {
	for _, c := range []struct {
		name      string
		level     int32
		talents   string
		raid      bool
		reduction float64
	}{
		{"our totem at 30", 30, "", false, 11},
		{"our totem at 60", 60, "", false, 30},
		// Guardian Totems 2.
		{"our totem at 30 with Guardian Totems", 30, "-0002", false, 13.2},
		{"another shaman's totem at 30", 30, "", true, 11},
	} {
		t.Run(c.name, func(t *testing.T) {
			raidBuffs := &proto.RaidBuffs{}
			if c.raid {
				raidBuffs.StoneskinTotem = proto.TristateEffect_TristateEffectRegular
			}
			player := newOrcShaman(c.level, c.talents, &proto.EnhancementShaman_Options{})
			player.BonusStats = &proto.UnitStats{Stats: stats.Stats{stats.Armor: 512}.ToFloatArray()}
			sim, enh := newTankSimWithBuffs(player, c.level, raidBuffs)
			if c.talents != "" && enh.Talents.GuardianTotems != 2 {
				t.Fatalf("got Guardian Totems %d, want 2", enh.Talents.GuardianTotems)
			}
			// hit lands an enemy's hit of raw damage on us and returns what it dealt.
			hit := func(sim *core.Simulation, raw float64) float64 {
				return enh.CurrentTarget.AutoAttacks.MHAuto().CalcDamage(sim, &enh.Unit, raw, forcedOutcome(core.OutcomeHit)).Damage
			}

			at(sim, 1, func(sim *core.Simulation) {
				armor := enh.GetStat(stats.Armor)
				throughArmor := 1 - armor/(armor+400+85*float64(c.level))
				if !c.raid {
					if got, want := hit(sim, 100), 100*throughArmor; math.Abs(got-want) > 1e-9 {
						t.Errorf("before the totem a 100 damage hit dealt %.3f, want %.3f", got, want)
					}
					castNow(t, sim, enh, topRank(t, "Stoneskin Totem", enh.StoneskinTotem))
				}
				if got, want := hit(sim, 100), (100-c.reduction)*throughArmor; math.Abs(got-want) > 1e-9 {
					t.Errorf("a 100 damage hit dealt %.3f through %.0f armor, want %.3f (%.1f off before armor)", got, armor, want, c.reduction)
				}
				if got := hit(sim, c.reduction-1); got != 0 {
					t.Errorf("a %.1f damage hit dealt %.3f, want 0", c.reduction-1, got)
				}
				if !c.raid {
					castNow(t, sim, enh, enh.TremorTotem)
					if got, want := hit(sim, 100), 100*throughArmor; math.Abs(got-want) > 1e-9 {
						t.Errorf("after Tremor Totem a 100 damage hit dealt %.3f, want %.3f", got, want)
					}
				}
			})
			runSim(sim)
		})
	}
}
