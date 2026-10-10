package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// registerRaidDamageHits feeds the raid damage we take to our shield.
//
// We don't sim the encounter's damage on the raid, so the user sets how many times a minute
// an enemy lands a direct hit on us (Raid hits per minute). Each hit fires a Lightning Shield
// orb at our target or spends a Water Shield globe, the same as a real hit, and both shields
// still wait 3.5 sec between procs. The hits only reach our shield. Nothing else we have that
// reacts to taking a hit sees them.
//
// The hits land at random times, with the set rate as their average. Each wait for the next
// hit is drawn on its own, the way hits from many enemies that don't know about each other
// add up. An even pace would overstate the procs because of the 3.5 sec wait. At 15 hits a
// minute, a hit every 4 sec fires the shield every time (15 procs a minute), but random hits
// at the same rate give about 8 (shaman_audit.md 6.5).
//
// With a variation, each iteration picks its own rate between the rate minus the variation
// and the rate plus it, the way Duration +/- picks each fight's length.
func (shaman *Shaman) registerRaidDamageHits() {
	if shaman.RaidDamageHitsPerMinute <= 0 {
		return
	}
	shaman.RegisterResetEffect(func(sim *core.Simulation) {
		hitsPerMinute := shaman.RaidDamageHitsPerMinute
		if variation := shaman.RaidDamageHitsPerMinuteVariation; variation > 0 {
			hitsPerMinute += (sim.RandomFloat("Raid Damage Hits")*2 - 1) * variation
		}
		if hitsPerMinute <= 0 {
			return
		}
		averageWait := float64(time.Minute) / hitsPerMinute
		hit := &core.PendingAction{}
		hit.OnAction = func(sim *core.Simulation) {
			if aura := shaman.ActiveShieldAura; aura != nil && aura.IsActive() {
				shaman.shieldHitTaken[aura](sim, shaman.CurrentTarget)
			}
			hit.NextActionAt = sim.CurrentTime + time.Duration(sim.RandomExpFloat("Raid Damage Hit Wait")*averageWait)
			sim.AddPendingAction(hit)
		}
		hit.NextActionAt = sim.CurrentTime + time.Duration(sim.RandomExpFloat("Raid Damage Hit Wait")*averageWait)
		sim.AddPendingAction(hit)
	})
}
