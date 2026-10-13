package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// Forever's Multi-Shot (beta client, build 1.60.1.70338) has one rank, no flat bonus, a 0.5 sec
// cast and 13.9% of base mana, and a 6 sec cooldown shared with Aimed Shot (Classic 10 sec alone,
// up to +150 at rank 5).
func (hunter *Hunter) registerMultiShotSpell(timer *core.Timer) {
	if hunter.Level < 18 {
		return
	}

	numHits := min(3, hunter.Env.GetNumTargets())
	results := make([]*core.SpellResult, numHits)

	hunter.MultiShot = hunter.RegisterSpell(core.SpellConfig{
		SpellCode:     SpellCode_HunterMultiShot,
		ActionID:      core.ActionID{SpellID: 2643},
		SpellSchool:   core.SpellSchoolPhysical,
		DefenseType:   core.DefenseTypeRanged,
		ProcMask:      core.ProcMaskRangedSpecial,
		Flags:         core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagShot,
		CastType:      proto.CastType_CastTypeRanged,
		Rank:          1,
		RequiredLevel: 18,
		MissileSpeed:  30,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.139,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 500,
			},
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				cast.CastTime = spell.CastTime()
			},
			IgnoreHaste: true, // Hunter GCD is locked at 1.5s
			CD: core.Cooldown{
				Timer:    timer,
				Duration: time.Second * 6,
			},
			CastTime: func(spell *core.Spell) time.Duration {
				return time.Duration(float64(spell.DefaultCast.CastTime) / hunter.RangedSwingSpeed())
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.inRangedRange()
		},

		CritDamageBonus: hunter.mortalShots(),

		DamageMultiplier: hunter.barrageMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			curTarget := target
			for hitIndex := int32(0); hitIndex < numHits; hitIndex++ {
				baseDamage := hunter.AutoAttacks.Ranged().CalculateNormalizedWeaponDamage(sim, spell.RangedAttackPower(curTarget, false)) +
					hunter.NormalizedAmmoDamageBonus

				results[hitIndex] = spell.CalcDamage(sim, curTarget, baseDamage, spell.OutcomeRangedHitAndCrit)

				curTarget = sim.Environment.NextTargetUnit(curTarget)
			}

			spell.WaitTravelTime(sim, func(s *core.Simulation) {
				for hitIndex := int32(0); hitIndex < numHits; hitIndex++ {
					spell.DealDamage(sim, results[hitIndex])
				}
			})
		},
	})
}
