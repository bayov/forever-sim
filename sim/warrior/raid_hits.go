package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// registerRaidDamageHits lands the raid damage we take on us.
//
// We don't sim the encounter's damage on the raid, so the options set how many times a minute an
// enemy lands a direct hit on us and how hard (Raid hits per minute and Raid damage per hit).
// Each hit is a Fire spell from our target that always lands for that damage, before our
// damage taken modifiers (Recklessness's 20% more, for one). It gives Rage like any damage we
// take (core.EnableRageBar: 2.5 times the damage over the attacker's Rage conversion, 15.4 Rage
// for 1500 from a level 63 boss), and anything else that reacts to a hit taken sees it too.
//
// The hits land at random times, with the set rate as their average: each wait for the next hit
// is drawn on its own, like the shaman's raid hits (shaman/shields.go), the way hits from many
// enemies that don't know about each other add up.
func (warrior *Warrior) registerRaidDamageHits() {
	hitsPerMinute, damage := warrior.RaidDamageHitsPerMinute, warrior.RaidDamagePerHit
	if hitsPerMinute <= 0 || damage <= 0 || warrior.CurrentTarget == nil {
		return
	}

	boss := warrior.CurrentTarget
	raidHit := boss.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{OtherID: proto.OtherAction_OtherActionDamageTaken},
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       core.SpellFlagIgnoreResists | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		DamageMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeAlwaysHit)
		},
	})

	averageWait := float64(time.Minute) / hitsPerMinute
	warrior.RegisterResetEffect(func(sim *core.Simulation) {
		hit := &core.PendingAction{}
		hit.OnAction = func(sim *core.Simulation) {
			raidHit.Cast(sim, &warrior.Unit)
			hit.NextActionAt = sim.CurrentTime + time.Duration(sim.RandomExpFloat("Raid Damage Hit Wait")*averageWait)
			sim.AddPendingAction(hit)
		}
		hit.NextActionAt = sim.CurrentTime + time.Duration(sim.RandomExpFloat("Raid Damage Hit Wait")*averageWait)
		sim.AddPendingAction(hit)
	})
}
