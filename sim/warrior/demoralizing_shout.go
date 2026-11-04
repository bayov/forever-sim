package warrior

import (
	"math"

	"github.com/wowsims/classic/sim/core"
)

func (warrior *Warrior) registerDemoralizingShoutSpell() {
	rank := int32(5)
	actionId := core.DemoralizingShoutSpellId[rank]

	warrior.DemoralizingShoutAuras = warrior.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		// Improved Demoralizing Shout is gone from the Forever tree, and Forever's rank 5 takes
		// 196 + 1.4 a level from 54, 204 at 60, for 45 sec (client 70425, spell 11556), which is
		// Classic's 146 with the old talent at 5/5.
		if warrior.Env.IsForever() {
			return core.ForeverDemoralizingShoutAura(target, warrior.Level)
		}
		return core.DemoralizingShoutAura(target, 0, 5)
	})

	// Under Forever the shout's threat is its E_THREAT effect, 43 + 0.8 a level from 54 (capped at
	// 64), on each enemy it lands on: 47 at 60, the game's whole points (client 70753, spell 11556;
	// added in 70291, "Demoralizing Shout now generates threat"). The sim gave Classic's 43.2 until
	// 2026-11-03.
	flatThreat := 0.4 * 2 * float64(core.DemoralizingShoutLevel[rank])
	if warrior.Env.IsForever() {
		flatThreat = math.Floor(43 + 0.8*float64(min(warrior.Level, 64)-54))
	}

	warrior.DemoralizingShout = warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: actionId},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagAPL | SpellFlagOffensive,

		RageCost: core.RageCostOptions{
			Cost: warrior.shoutRageCost(),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		ThreatMultiplier: 0.4,
		FlatThreatBonus:  flatThreat,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
				result := spell.CalcAndDealOutcome(sim, aoeTarget, spell.OutcomeMagicHit)
				if result.Landed() {
					warrior.DemoralizingShoutAuras.Get(aoeTarget).Activate(sim)
				}
			}
		},

		RelatedAuras: []core.AuraArray{warrior.DemoralizingShoutAuras},
	})
}
