package enhancement

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// The Restoration talents' places in the tree's talent string (healing_test.go lists them).
const (
	restoIdxImprovedHealingWave = 0
	restoIdxTidalFocus          = 1
	restoIdxAncestralHealing    = 6
	restoIdxHealingFocus        = 7
	restoIdxTidalMastery        = 9
	restoIdxHealingWay          = 12
	restoIdxPurification        = 14
)

// resto is a talent string with points in one Restoration talent, and the rest of the
// Restoration string after it (like "1" for Riptide after Purification).
func resto(idx, points int, after string) string {
	return "--" + strings.Repeat("0", idx) + fmt.Sprint(points) + after
}

// TestOrcShamanHealTalentPoints checks every point of each healing talent against its
// Forever client curve (70338, the same in 70425). The other healing tests only check the
// top rank, so a wrong value for a lower rank would get past them.
func TestOrcShamanHealTalentPoints(t *testing.T) {
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

	// Improved Healing Wave (16182): 0.1 sec off Healing Wave a point, nothing else.
	for points := 1; points <= 5; points++ {
		_, enh := healSim(t, 60, resto(restoIdxImprovedHealingWave, points, ""), 0, 0)
		if got, want := enh.HealingWave[10].DefaultCast.CastTime, 3*time.Second-time.Duration(points)*100*time.Millisecond; got != want {
			t.Errorf("Improved Healing Wave %d: Healing Wave casts in %s, want %s", points, got, want)
		}
		if got := enh.LesserHealingWave[6].DefaultCast.CastTime; got != 1500*time.Millisecond {
			t.Errorf("Improved Healing Wave %d: Lesser Healing Wave casts in %s, want 1.5s", points, got)
		}
	}

	// Healing Way (29206): Healing Wave heals 8, 17 or 25% more. Lesser Healing Wave and
	// Chain Heal aren't in its mask.
	for points, bonus := range []float64{0, .08, .17, .25} {
		_, enh := healSim(t, 60, resto(restoIdxHealingWay, points, ""), 0, 0)
		if got := enh.HealingWave[10].DamageMultiplier; !near(got, 1+bonus) {
			t.Errorf("Healing Way %d: Healing Wave heals x%.3f, want x%.3f", points, got, 1+bonus)
		}
		for _, spell := range []*core.Spell{enh.LesserHealingWave[6], enh.ChainHeal[3]} {
			if got := spell.DamageMultiplier; !near(got, 1) {
				t.Errorf("Healing Way %d: %s heals x%.3f, want x1", points, spell.ActionID, got)
			}
		}
	}

	// Purification (16178): 2% more a point on Healing Wave, Lesser Healing Wave, Chain Heal
	// and Riptide (its heal and its ticks), but not on Healing Stream Totem.
	_, plain := healSim(t, 60, "", 0, 0)
	for points := 1; points <= 5; points++ {
		_, enh := healSim(t, 60, resto(restoIdxPurification, points, "1"), 0, 0)
		want := 1 + .02*float64(points)
		for _, spell := range []*core.Spell{enh.HealingWave[10], enh.LesserHealingWave[6], enh.ChainHeal[3], enh.Riptide} {
			if got := spell.DamageMultiplier; !near(got, want) {
				t.Errorf("Purification %d: %s heals x%.3f, want x%.3f", points, spell.ActionID, got, want)
			}
		}
		if got := enh.Riptide.SelfHot().DamageMultiplier; !near(got, 1) {
			t.Errorf("Purification %d: Riptide's ticks take x%.3f on top of the spell's, want x1", points, got)
		}
		stream := topRank(t, "Healing Stream Totem", enh.HealingStreamTotem)
		if got, base := stream.DamageMultiplier, topRank(t, "Healing Stream Totem", plain.HealingStreamTotem).DamageMultiplier; !near(got, base) {
			t.Errorf("Purification %d: Healing Stream Totem heals x%.3f, want x%.3f as without it", points, got, base)
		}
	}

	// Healing Focus (16181): a 23, 47 or 70% chance for a hit not to push back a heal with a
	// cast time. Lightning Bolt isn't in its mask.
	for points, chance := range []float64{0, .23, .47, .70} {
		_, enh := healSim(t, 60, resto(restoIdxHealingFocus, points, ""), 0, 0)
		for _, spell := range []*core.Spell{enh.HealingWave[10], enh.LesserHealingWave[6], enh.ChainHeal[3]} {
			if got := spell.PushbackReduction; !near(got, chance) {
				t.Errorf("Healing Focus %d: %s ignores pushback %.0f%% of the time, want %.0f%%", points, spell.ActionID, 100*got, 100*chance)
			}
		}
		if got := enh.LightningBolt[10].PushbackReduction; got != 0 {
			t.Errorf("Healing Focus %d: Lightning Bolt ignores pushback %.0f%% of the time, want never", points, 100*got)
		}
	}

	// Tidal Focus (16179): 1% off every heal's cost a point.
	for points := 1; points <= 5; points++ {
		_, enh := healSim(t, 60, resto(restoIdxTidalFocus, points, ""), 0, 0)
		for _, c := range []struct {
			spell *core.Spell
			cost  float64
		}{{enh.HealingWave[10], 620}, {enh.LesserHealingWave[6], 380}, {enh.ChainHeal[3], 405}} {
			want := c.cost * (1 - .01*float64(points))
			if got := c.spell.Cost.GetCurrentCost(); math.Abs(got-want) > 0.01 {
				t.Errorf("Tidal Focus %d: %s costs %.2f, want %.2f", points, c.spell.ActionID, got, want)
			}
		}
	}

	// Tidal Mastery (16194): 1% more crit a point on every heal, Riptide too, and none on
	// Lightning Bolt (its Forever mask has only the heals).
	for points := 1; points <= 5; points++ {
		_, enh := healSim(t, 60, resto(restoIdxTidalMastery, points, "000001"), 0, 0)
		_, base := healSim(t, 60, resto(restoIdxTidalMastery, 0, "000001"), 0, 0)
		want := .01 * float64(points)
		for i, spell := range []*core.Spell{enh.HealingWave[10], enh.LesserHealingWave[6], enh.ChainHeal[3], enh.Riptide} {
			other := []*core.Spell{base.HealingWave[10], base.LesserHealingWave[6], base.ChainHeal[3], base.Riptide}[i]
			if got := spell.HealingCritChance() - other.HealingCritChance(); !near(got, want) {
				t.Errorf("Tidal Mastery %d: %s gets %.3f more crit, want %.3f", points, spell.ActionID, got, want)
			}
		}
		if enh.LightningBolt[10].SpellCritChance(enh.CurrentTarget) != base.LightningBolt[10].SpellCritChance(base.CurrentTarget) {
			t.Errorf("Tidal Mastery %d changes Lightning Bolt's crit", points)
		}
	}

	// Ancestral Healing (16176): 8, 17 or 25% more armor for 15 sec after a heal crit.
	for points, bonus := range []float64{0, .08, .17, .25} {
		if points == 0 {
			continue
		}
		sim, enh := healSim(t, 60, resto(restoIdxAncestralHealing, points, ""), 0, 200)
		var armor float64
		at(sim, 0.5, func(sim *core.Simulation) { armor = enh.Armor() })
		healOnce(t, sim, enh, enh.LesserHealingWave[6], 1, 1.5)
		at(sim, 3, func(sim *core.Simulation) {
			if got := enh.Armor(); math.Abs(got-armor*(1+bonus)) > 0.01 {
				t.Errorf("Ancestral Healing %d: armor %.1f after a crit heal, want %.1f", points, got, armor*(1+bonus))
			}
		})
		runSim(sim)
	}
}

// TestOrcShamanHealingFocusPushback casts Healing Wave 400 times at level 30, and a mob hits
// us half a second into each cast. Without Healing Focus every hit pushes the heal back. With
// 1 and 3 points 23% and 70% of the hits don't (the client's curve). A wrong chance, or one
// that reads another spell's, fails it.
func TestOrcShamanHealingFocusPushback(t *testing.T) {
	const casts = 400
	pushed := func(points int) int {
		mob := core.NormalMobTarget(30, proto.MobType_MobTypeBeast, 1)
		// One swing at the pull and none after, so only our own hits below land.
		mob.SwingSpeed = 100000
		req := normalMobRequest([]*proto.Target{mob}, true, 1)
		req.Raid.Parties[0].Players[0].TalentsString = resto(restoIdxHealingFocus, points, "")
		// A timed fight, so our swings never kill the mob.
		req.Encounter.UseHealth = false
		req.Encounter.Duration = casts*4 + 10
		sim := core.NewSim(req, simsignals.CreateSignals())
		sim.Reset()
		sim.PrePull()
		enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
		enh.AddStatDynamic(sim, stats.Health, 100000)
		swing := sim.Encounter.TargetUnits[0].AutoAttacks.MHAuto()
		heal := enh.HealingWave[5]

		n := 0
		for i := range casts {
			start := 2 + 4*float64(i)
			var expires time.Duration
			at(sim, start, func(sim *core.Simulation) {
				enh.AddMana(sim, enh.MaxMana(), enh.GetManaNotCastingMetrics())
				castNow(t, sim, enh, heal)
				expires = enh.Hardcast.Expires
			})
			at(sim, start+0.5, func(sim *core.Simulation) {
				swing.CalcAndDealDamage(sim, &enh.Unit, 10, swing.OutcomeAlwaysHit)
				if enh.Hardcast.Expires != expires {
					n++
				}
			})
		}
		runSim(sim)
		return n
	}

	for _, c := range []struct {
		points int
		ignore float64
	}{{0, 0}, {1, .23}, {3, .70}} {
		got := pushed(c.points)
		want := casts * (1 - c.ignore)
		// Four standard deviations of the binomial.
		margin := 4 * math.Sqrt(casts*c.ignore*(1-c.ignore))
		t.Logf("Healing Focus %d: %d of %d hits pushed Healing Wave back, want about %.0f", c.points, got, casts, want)
		if math.Abs(float64(got)-want) > margin {
			t.Errorf("Healing Focus %d: %d of %d hits pushed Healing Wave back, want %.0f +- %.0f", c.points, got, casts, want, margin)
		}
	}
}
