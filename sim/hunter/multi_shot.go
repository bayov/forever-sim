package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// Forever's Multi-Shot (beta client, build 1.60.1.70338) has one rank, no flat bonus, 13.9% of base
// mana and a 6 sec cooldown shared with Aimed Shot (Classic 10 sec alone, up to +150 at rank 5). It
// is instant since client 1.60.1.70425 (SpellMisc casting time index 3, 500 ms, became 1, none;
// not in the patch notes). Its interrupt flags (SpellInterrupts, 70425) are 9, without the pushback
// bit (0x2) that Aimed Shot's and Sniper Shot's 7 carry, which no longer matters with no cast.
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
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true, // Hunter GCD is locked at 1.5s
			CD: core.Cooldown{
				Timer:    timer,
				Duration: time.Second * 6,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.inRangedRange()
		},

		CritDamageBonus: hunter.mortalShots(),

		DamageMultiplier:  hunter.barrageMultiplier(),
		ThreatMultiplier:  1,
		PushbackReduction: 1,
		BonusCoefficient:  1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			curTarget := target
			// Only living targets, each one once.
			hits := sim.Encounter.AoETargetCount(numHits)
			for hitIndex := int32(0); hitIndex < hits; hitIndex++ {
				baseDamage := hunter.AutoAttacks.Ranged().CalculateNormalizedWeaponDamage(sim, spell.RangedAttackPower(curTarget, false)) +
					hunter.NormalizedAmmoDamageBonus

				results[hitIndex] = spell.CalcDamage(sim, curTarget, baseDamage, spell.OutcomeRangedHitAndCrit)

				curTarget = sim.Environment.NextTargetUnit(curTarget)
			}

			spell.WaitTravelTime(sim, func(s *core.Simulation) {
				for hitIndex := int32(0); hitIndex < hits; hitIndex++ {
					spell.DealDamage(sim, results[hitIndex])
				}
			})
		},
	})
}
