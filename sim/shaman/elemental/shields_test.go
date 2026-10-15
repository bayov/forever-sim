package elemental

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

// TestOrcShamanRaidDamageHits checks that Elemental's raid hits option reaches the shield, like
// Enhancement's (enhancement TestOrcShamanRaidDamageHits has the numbers).
//
// We keep Lightning Shield at 3 orbs for 9 min. 15 random hits a minute fire about 8 orbs a
// minute, and no hits fire none.
func TestOrcShamanRaidDamageHits(t *testing.T) {
	const minutes = 9
	for _, hitsPerMinute := range []float64{0, 15} {
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid: core.SinglePlayerRaidProto(&proto.Player{
				Race:      proto.Race_RaceOrc,
				Class:     proto.Class_ClassShaman,
				Level:     60,
				Equipment: &proto.EquipmentSpec{},
				Rotation:  &proto.APLRotation{},
				Spec: &proto.Player_ElementalShaman{ElementalShaman: &proto.ElementalShaman{
					Options: &proto.ElementalShaman_Options{RaidDamageHitsPerMinute: hitsPerMinute},
				}},
			}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  &proto.Encounter{Duration: minutes * 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
			SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1, ExactDamage: true},
		}, simsignals.CreateSignals())
		sim.Reset()
		sim.PrePull()
		ele := sim.Raid.Parties[0].Players[0].(*ElementalShaman)
		shield := ele.LightningShield[len(ele.LightningShield)-1]
		orb := ele.LightningShieldProcs[len(ele.LightningShieldProcs)-1]
		sim.AddPendingAction(&core.PendingAction{NextActionAt: core.DurationFromSeconds(0.5), OnAction: func(sim *core.Simulation) {
			ele.SetGCDTimer(sim, sim.CurrentTime)
			if !shield.Cast(sim, ele.CurrentTarget) {
				t.Fatalf("Lightning Shield didn't cast")
			}
			core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				Period:   time.Second,
				OnAction: func(sim *core.Simulation) { ele.ActiveShieldAura.SetStacks(sim, 3) },
			})
		}})
		for !sim.Step() {
		}

		metrics := orb.SpellMetrics[ele.CurrentTarget.UnitIndex]
		fired := float64(metrics.Hits + metrics.Misses)
		want := 0.0
		if hitsPerMinute > 0 {
			want = minutes * 60 / (3.5 + 60/hitsPerMinute)
		}
		if math.Abs(fired-want) > 15 {
			t.Errorf("%.0f hits a minute: %.0f orbs fired at our target, want about %.0f", hitsPerMinute, fired, want)
		}
	}
}
