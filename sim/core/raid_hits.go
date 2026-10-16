package core

import (
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

// Raid damage on the players the boss isn't tanking (Encounter.raid_damage_hits_per_minute and
// raid_damage_per_hit).
//
// We don't sim the encounter's damage on the raid. A DPS player standing behind the boss still
// takes some of it, from raid-wide spells, cleaves and the like, so the encounter can set how
// many direct hits a minute land on each such player and how hard. Every class takes the same
// hits: each is a Fire spell from the boss that always lands for the set damage, before the
// player's damage taken modifiers, and ignores resistances. It does whatever any hit taken does:
// a Lightning Shield orb or a Water Shield globe, Rage and Enrage for a warrior, and spell
// pushback on a cast. It takes no health, because the boss isn't hitting the player (health.go
// only tracks it for a tanked player), so nobody dies of it.
//
// A player the boss is tanking takes the boss's own attacks instead and gets no raid hits.
//
// The hits land at random times, with the set rate as their average: each wait for the next hit
// is drawn on its own, the way hits from many enemies that don't know about each other add up.
// An even pace would overstate the procs of anything with an internal cooldown (a shield waits
// 3.5 sec between procs).
func (env *Environment) registerRaidDamageHits() {
	hitsPerMinute, damage := env.Encounter.RaidDamageHitsPerMinute, env.Encounter.RaidDamagePerHit
	if hitsPerMinute <= 0 || damage <= 0 || len(env.Encounter.TargetUnits) == 0 {
		return
	}

	boss := env.Encounter.TargetUnits[0]
	raidHit := boss.RegisterSpell(SpellConfig{
		ActionID:    ActionID{OtherID: proto.OtherAction_OtherActionDamageTaken},
		SpellSchool: SpellSchoolFire,
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskSpellDamage,
		Flags:       SpellFlagIgnoreResists | SpellFlagNoOnCastComplete | SpellFlagPassiveSpell,

		DamageMultiplier: 1,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeAlwaysHit)
		},
	})

	averageWait := float64(time.Minute) / hitsPerMinute
	for _, player := range env.Raid.AllPlayerUnits {
		if env.isTanked(player) {
			continue
		}
		player.RegisterResetEffect(func(sim *Simulation) {
			hit := &PendingAction{}
			hit.OnAction = func(sim *Simulation) {
				raidHit.Cast(sim, player)
				hit.NextActionAt = sim.CurrentTime + time.Duration(sim.RandomExpFloat("Raid Damage Hit Wait")*averageWait)
				sim.AddPendingAction(hit)
			}
			hit.NextActionAt = sim.CurrentTime + time.Duration(sim.RandomExpFloat("Raid Damage Hit Wait")*averageWait)
			sim.AddPendingAction(hit)
		})
	}
}

// Whether any target is attacking the unit.
func (env *Environment) isTanked(unit *Unit) bool {
	for _, target := range env.Encounter.TargetUnits {
		if target.CurrentTarget == unit {
			return true
		}
	}
	return false
}
