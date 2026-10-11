package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanWindfuryTotem checks Windfury Totem's attack power by rank, how often it procs
// and its wait between procs (shaman_audit.md 7.19).
//
// The extra swing gets 95, 179 or 246 attack power at levels 32, 42 and 52, as in the Forever
// client. Each landed white hit has a 20% chance to proc it, so 500 hits 2 sec apart should
// proc it about 100 times. The sim waits 1.5 sec between procs, so hits 0.5 sec apart never
// proc it twice within 1.5 sec. The client has 100 ms there, which notes.md asks to check at
// level 32. The totem is another shaman's, and we have Rockbiter on, so Windfury Weapon doesn't
// turn it off. A proc moves our next swing to right now.
func TestOrcShamanWindfuryTotem(t *testing.T) {
	start := func(level int32, seconds float64) (*core.Simulation, *EnhancementShaman) {
		sim, enh := weaponSim{level: level, weapon: &proto.ItemSpec{Id: darkEdge}, imbue: proto.WeaponImbue_RockbiterWeapon,
			totem: proto.TotemWeaponBuff_TotemWeaponBuffWindfury, seconds: seconds}.start()
		sim.PrePull()
		at(sim, 0.5, enh.AutoAttacks.CancelAutoSwing)
		return sim, enh
	}

	for _, c := range []struct {
		level int32
		ap    float64
	}{{32, 95}, {42, 179}, {52, 246}} {
		sim, enh := start(c.level, 2)
		at(sim, 1, func(sim *core.Simulation) {
			before := enh.GetStat(stats.AttackPower)
			enh.GetAura("Windfury Totem Raid Buff").Activate(sim)
			if got := enh.GetStat(stats.AttackPower) - before; got != c.ap {
				t.Errorf("level %d: Windfury Totem gave %.0f attack power, want %.0f", c.level, got, c.ap)
			}
		})
		runSim(sim)
	}

	// procTimes lands n white hits every apart, and returns when they procced it.
	procTimes := func(n int, every float64) []time.Duration {
		sim, enh := start(60, 1+every*float64(n))
		var procs []time.Duration
		for i := range n {
			at(sim, 1+every*float64(i), func(sim *core.Simulation) {
				enh.AutoAttacks.MHAuto().CalcAndDealDamage(sim, enh.CurrentTarget, 100, forcedOutcome(core.OutcomeHit))
				if enh.AutoAttacks.MainhandSwingAt() == sim.CurrentTime {
					procs = append(procs, sim.CurrentTime)
				}
			})
		}
		runSim(sim)
		return procs
	}

	// 100 procs give or take 9, and we allow 4 times that.
	if procs := len(procTimes(500, 2)); math.Abs(float64(procs)-100) > 36 {
		t.Errorf("500 white hits 2 sec apart procced Windfury Totem %d times, want about 100", procs)
	}

	procs := procTimes(400, 0.5)
	if len(procs) < 20 {
		t.Errorf("400 white hits 0.5 sec apart procced Windfury Totem %d times, want about 50", len(procs))
	}
	for i := 1; i < len(procs); i++ {
		if gap := procs[i] - procs[i-1]; gap < 1500*time.Millisecond {
			t.Errorf("Windfury Totem procced at %s and again %s later, want 1.5 sec at least", procs[i-1], gap)
		}
	}
}
