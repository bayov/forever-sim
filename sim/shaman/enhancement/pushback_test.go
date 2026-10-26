package enhancement

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

// TestOrcShamanCastPushback hits us with a mob's swings while we cast Lightning Bolt (2.5
// sec at level 30). Each hit takes the cast bar back 1 sec, then 0.8, 0.6, 0.4 and 0.2 for
// every hit after that, as on the cmangos Classic server (Spell::Delayed). The bar can't go
// back past its start, so the first hit, half a second in, takes it back only half a
// second.
func TestOrcShamanCastPushback(t *testing.T) {
	mob := core.NormalMobTarget(30, proto.MobType_MobTypeBeast, 1)
	// One swing at the pull and none after, so only our own hits below land.
	mob.SwingSpeed = 1000
	sim := core.NewSim(normalMobRequest([]*proto.Target{mob}, true, 1), simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
	swing := sim.Encounter.TargetUnits[0].AutoAttacks.MHAuto()
	bolt := topRank(t, "Lightning Bolt", enh.LightningBolt)

	hit := func(sim *core.Simulation) {
		swing.CalcAndDealDamage(sim, &enh.Unit, 10, swing.OutcomeAlwaysHit)
	}
	check := func(seconds float64, want float64) {
		at(sim, seconds+0.001, func(sim *core.Simulation) {
			if got := enh.Hardcast.Expires; got != core.DurationFromSeconds(want) {
				t.Errorf("at %.1f sec the bolt lands at %s, want %.1f sec", seconds, got, want)
			}
		})
	}

	at(sim, 0.1, func(sim *core.Simulation) {
		castNow(t, sim, enh, bolt)
		if got := enh.Hardcast.Expires - sim.CurrentTime; got != 2500*time.Millisecond {
			t.Fatalf("Lightning Bolt casts in %s, want 2.5 sec", got)
		}
	})
	// 0.5 sec in, the bar goes back to its start: 2.5 sec left, landing at 3.1.
	at(sim, 0.6, hit)
	check(0.6, 3.1)
	// 1.5 sec left, plus 0.8.
	at(sim, 1.6, hit)
	check(1.6, 3.9)
	// 1.3 sec left, plus 0.6.
	at(sim, 2.6, hit)
	check(2.6, 4.5)
	runSim(sim)
}

// TestLevelingHealPushback grinds a session where we cast Healing Wave (3 sec) below 40%
// health. Every mob hit that lands while we cast pushes the heal back, and a whole session
// has plenty of them: the heal outlasts a mob's swing. Lesser Healing Wave's 1.5 sec, cast
// just after the hit that took us under 40%, ends before the next one.
func TestLevelingHealPushback(t *testing.T) {
	healingWave := `{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"40%"}}}},"castSpell":{"spellId":{"spellId":939}}}}`
	req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{MaxPackSize: 3}, 1), rotationOf(healingWave))
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	mobHit := regexp.MustCompile(`^\[Target \d+\] \[ \(#1\)\] \{OtherID: 7, Tag: 1\} (Hit|Crit|Crush|Block) `)
	casting := false
	hits, pushbacks, casts := 0, 0, 0
	for _, line := range strings.Split(result.Logs, "\n") {
		m := logTime.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		event := line[len(m[0]):]
		switch {
		case strings.HasPrefix(event, "[ (#1)] Casting {SpellID: 939}"):
			casting = true
			casts++
		case strings.HasPrefix(event, "[ (#1)] Completed cast {SpellID: 939}"):
			casting = false
		case casting && mobHit.MatchString(event):
			hits++
		case casting && strings.HasPrefix(event, "[ (#1)] Cast {SpellID: 939} pushed back"):
			pushbacks++
		}
	}
	if hits < 10 {
		t.Fatalf("%d mob hits landed in %d heals, want more to check pushback", hits, casts)
	}
	if pushbacks != hits {
		t.Errorf("%d mob hits landed while we cast Healing Wave, but %d pushed it back", hits, pushbacks)
	}
}
