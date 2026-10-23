package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (warrior *Warrior) registerThunderClapSpell() {
	spellID := int32(11581)
	baseDamage := 103.0
	has5pcConq := warrior.HasSetBonus(ItemSetConquerorsBattleGear, 5)
	attackSpeedReduction := core.TernaryInt32(has5pcConq, 15, 10)
	stanceMask := BattleStance
	cooldown := time.Second * 4
	apCoef := 0.0
	// Forever's Thunder Clap slows by 20% on a 6 sec cooldown and works in Defensive Stance too
	// (client 70338, spell 11581). It adds 7% of attack power, which wowsims/forever fits from rank 1
	// logs (1172f1c745: floor(10 + 0.07 × attack power) before armor).
	if warrior.Env.IsForever() {
		apCoef = 0.07
		attackSpeedReduction += 10
		stanceMask |= DefensiveStance
		cooldown = time.Second * 6
	}

	warrior.ThunderClapAuras = warrior.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.ThunderClapAura(target, spellID, attackSpeedReduction)
	})

	results := make([]*core.SpellResult, min(4, warrior.Env.GetNumTargets()))

	warrior.ThunderClap = warrior.RegisterSpell(stanceMask, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       core.SpellFlagAPL | SpellFlagOffensive,

		RageCost: core.RageCostOptions{
			Cost: 20 - []float64{0, 2, 4, 6}[warrior.Talents.ImprovedThunderClap],
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: cooldown,
			},
		},

		CritDamageBonus: warrior.impale(),

		DamageMultiplier: core.TernaryFloat64(has5pcConq, 1.5, 1),
		ThreatMultiplier: 2.5,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Only living targets, each one once.
			hits := results[:sim.Encounter.AoETargetCount(int32(len(results)))]
			for idx := range hits {
				results[idx] = spell.CalcDamage(sim, target, baseDamage+apCoef*spell.MeleeAttackPower(target), spell.OutcomeMagicHitAndCrit)
				target = sim.Environment.NextTargetUnit(target)
			}

			for _, result := range hits {
				spell.DealDamage(sim, result)
				if result.Landed() {
					warrior.ThunderClapAuras.Get(result.Target).Activate(sim)
				}
			}
		},

		RelatedAuras: []core.AuraArray{warrior.ThunderClapAuras},
	})
}
