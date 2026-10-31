package core

import (
	"math"
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

// MeleePullRatio and RangedPullRatio are how far past the tank's threat a player has to go to pull
// aggro: 110% in melee range and 130% at range (Classic's rule).
const (
	MeleePullRatio  = 1.1
	RangedPullRatio = 1.3
)

// aggroWatch follows one player's or pet's threat on the boss in an iteration of a raid with a
// threat check (Raid.threat_check, docs/raid_sim.md, Threat).
//
// We don't sim the tank yet, so the tank's threat is a steady threat per second that started a few
// seconds before the pull. A pass is the first moment the member's threat goes over the pull ratio
// times that line. Threat only grows when it's booked, while the tank's line grows all the time, so
// checking at each booking finds every pass and the highest share of the line the member reaches.
type aggroWatch struct {
	tankTPS   float64
	headStart float64
	pullRatio float64

	// This iteration.
	threat   float64
	peak     float64
	passedAt time.Duration // -1 while it hasn't passed

	// Over the iterations.
	passes    int32
	firstPass []int32 // passes by the second of the fight they came in
	peaks     DistributionMetrics
}

func newAggroWatch(check *proto.ThreatCheck, pullRatio float64) *aggroWatch {
	return &aggroWatch{
		tankTPS:   check.TankTps,
		headStart: check.HeadStart,
		pullRatio: pullRatio,
		passedAt:  -1,
		peaks:     NewDistributionMetrics(),
	}
}

// setupThreatCheck gives every player and pet in the raid an aggroWatch, when the raid asks for a
// threat check.
func (raid *Raid) setupThreatCheck(check *proto.ThreatCheck) {
	if check == nil || check.TankTps <= 0 {
		return
	}
	for _, party := range raid.Parties {
		for _, agent := range party.Players {
			character := agent.GetCharacter()
			pullRatio := MeleePullRatio
			if slices.Contains(check.Ranged, int32(party.Index*5+character.PartyIndex)) {
				pullRatio = RangedPullRatio
			}
			character.aggro = newAggroWatch(check, pullRatio)
			for _, pet := range character.Pets {
				pet.aggro = newAggroWatch(check, MeleePullRatio)
			}
		}
	}
}

// watchThreat is called whenever threat is booked for the unit on an enemy.
func (unit *Unit) watchThreat(sim *Simulation, target *Unit, threat float64) {
	watch := unit.aggro
	// The boss is the first target. Before the pull nobody is in combat.
	if watch == nil || target.Type != EnemyUnit || target.UnitIndex != 0 || sim.CurrentTime < 0 {
		return
	}
	watch.threat += threat
	tank := watch.tankTPS * (sim.CurrentTime.Seconds() + watch.headStart)
	if tank <= 0 {
		return
	}
	share := watch.threat / (tank * watch.pullRatio)
	watch.peak = max(watch.peak, share)
	if share > 1 && watch.passedAt < 0 {
		watch.passedAt = sim.CurrentTime
		if sim.Log != nil {
			unit.Log(sim, "Pulled aggro: %0.0f threat, past the tank's %0.0f x %0.1f", watch.threat, tank, watch.pullRatio)
		}
	}
}

func (watch *aggroWatch) reset() {
	watch.threat = 0
	watch.peak = 0
	watch.passedAt = -1
}

func (watch *aggroWatch) doneIteration(sim *Simulation) {
	if watch.passedAt >= 0 {
		watch.passes++
		second := int(watch.passedAt.Seconds())
		for len(watch.firstPass) <= second {
			watch.firstPass = append(watch.firstPass, 0)
		}
		watch.firstPass[second]++
	}
	// DistributionMetrics divides by the fight length.
	watch.peaks.Total = watch.peak * sim.Duration.Seconds()
	watch.peaks.doneIteration(sim)
}

func (watch *aggroWatch) toProto() *proto.AggroMetrics {
	n := math.Max(float64(watch.peaks.n), 1)
	return &proto.AggroMetrics{
		PullRatio: watch.pullRatio,
		Chance:    float64(watch.passes) / n,
		FirstPass: watch.firstPass,
		Peak:      watch.peaks.ToProto(),
	}
}
