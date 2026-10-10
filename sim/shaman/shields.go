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
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period: time.Duration(float64(time.Minute) / hitsPerMinute),
			OnAction: func(sim *core.Simulation) {
				if aura := shaman.ActiveShieldAura; aura != nil && aura.IsActive() {
					shaman.shieldHitTaken[aura](sim, shaman.CurrentTarget)
				}
			},
		})
	})
}
