package enhancement

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanEyeOfTheStorm checks how often a hit pushes back our cast with Eye of the Storm
// (shaman_audit.md 8.10).
//
// The client gives a 23, 47 and 70% chance for 1, 2 and 3 points to ignore the pushback on
// Lightning Bolt, Chain Lightning and Lava Burst. Without the talent the first hit in a cast adds
// 1 sec and the second 0.8 sec, but never past the cast bar's start (pushback_test.go). We cast a
// top rank Lightning Bolt (2.5 sec) and an enemy spell hits us 1 sec into it, where the full 1 sec
// fits, 2000 times for each rank, and count the casts that still end on time.
func TestOrcShamanEyeOfTheStorm(t *testing.T) {
	const casts = 2000
	for points, chance := range []float64{0, .23, .47, .70} {
		// Eye of the Storm points and Lava Burst.
		sim, enh := newShamanSim(newOrcShaman(60, fmt.Sprintf("000000000%d001", points), &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, casts*4+10)
		if enh.Talents.EyeOfTheStorm != int32(points) {
			t.Fatalf("got Eye of the Storm %d, want %d", enh.Talents.EyeOfTheStorm, points)
		}
		for _, s := range []struct {
			name  string
			ranks []*core.Spell
		}{
			{"Lightning Bolt", enh.LightningBolt},
			{"Chain Lightning", enh.ChainLightning},
			{"Lava Burst", []*core.Spell{enh.LavaBurst}},
		} {
			for _, spell := range s.ranks {
				if spell != nil && spell.PushbackReduction != chance {
					t.Errorf("Eye of the Storm %d: %s ignores pushback %.0f%% of the time, want %.0f%%", points, spell.ActionID, spell.PushbackReduction*100, chance*100)
				}
			}
		}

		hit := enh.CurrentTarget.RegisterSpell(core.SpellConfig{
			ActionID:         core.ActionID{SpellID: 1},
			SpellSchool:      core.SpellSchoolFire,
			DefenseType:      core.DefenseTypeMagic,
			ProcMask:         core.ProcMaskSpellDamage,
			DamageMultiplier: 1,
		})
		hitUs := func(sim *core.Simulation) {
			hit.CalcAndDealDamage(sim, &enh.Unit, 1, forcedOutcome(core.OutcomeHit))
		}
		lb := topRank(t, "Lightning Bolt", enh.LightningBolt)
		onTime := 0
		for i := range casts {
			start := float64(4 * i)
			at(sim, start+1, func(sim *core.Simulation) {
				enh.AddMana(sim, enh.MaxMana(), enh.GetManaNotCastingMetrics())
				castNow(t, sim, enh, lb)
			})
			at(sim, start+2, func(sim *core.Simulation) {
				hitUs(sim)
				switch enh.Hardcast.Expires {
				case core.DurationFromSeconds(start + 3.5):
					onTime++
				case core.DurationFromSeconds(start + 4.5):
				default:
					t.Fatalf("Eye of the Storm %d: the cast now ends at %s, want %s or %s", points, enh.Hardcast.Expires, core.DurationFromSeconds(start+3.5), core.DurationFromSeconds(start+4.5))
				}
			})
		}
		runSim(sim)
		if got := float64(onTime) / casts; math.Abs(got-chance) > .05 {
			t.Errorf("Eye of the Storm %d: %.1f%% of %d casts ignored the hit, want %.0f%%", points, got*100, casts, chance*100)
		}
	}

	// Without the talent a second hit adds 0.8 sec.
	sim, enh := newShamanSim(newOrcShaman(60, "", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 10)
	hit := enh.CurrentTarget.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 1},
		SpellSchool:      core.SpellSchoolFire,
		DefenseType:      core.DefenseTypeMagic,
		ProcMask:         core.ProcMaskSpellDamage,
		DamageMultiplier: 1,
	})
	at(sim, 1, func(sim *core.Simulation) {
		castNow(t, sim, enh, topRank(t, "Lightning Bolt", enh.LightningBolt))
	})
	// 1 sec into the cast, 1.5 sec left plus 1 sec. Then 1.5 sec left again, plus 0.8.
	at(sim, 2, func(sim *core.Simulation) {
		hit.CalcAndDealDamage(sim, &enh.Unit, 1, forcedOutcome(core.OutcomeHit))
	})
	at(sim, 3, func(sim *core.Simulation) {
		hit.CalcAndDealDamage(sim, &enh.Unit, 1, forcedOutcome(core.OutcomeHit))
		if want := 5300 * time.Millisecond; enh.Hardcast.Expires != want {
			t.Errorf("two hits push the cast to end at %s, want %s", enh.Hardcast.Expires, want)
		}
	})
	runSim(sim)
}
