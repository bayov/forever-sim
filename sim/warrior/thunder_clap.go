package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (warrior *Warrior) registerThunderClapSpell() {
	spellIDs := []int32{0, 6343, 8198, 8204, 8205, 11580, 11581}
	rank := warrior.knownRank([]int32{0, 6, 18, 28, 38, 48, 58}, spellIDs)
	if rank == 0 {
		return
	}
	spellID := spellIDs[rank]
	baseDamage := []float64{0, 10, 23, 37, 55, 82, 103}[rank]
	slowDuration := time.Second * []time.Duration{0, 10, 14, 18, 22, 26, 30}[rank]
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
		aura := core.ThunderClapAura(target, spellID, attackSpeedReduction)
		if aura.Duration != core.NeverExpires { // the raid's, which stays up
			aura.Duration = slowDuration
		}
		return aura
	})

	results := make([]*core.SpellResult, min(4, warrior.Env.GetNumTargets()))

	warrior.ThunderClap = warrior.RegisterSpell(stanceMask, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       core.SpellFlagAPL | SpellFlagOffensive,
		MaxRange:    thunderClapRadius,

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
		// EV-0048: under Forever its threat is 1 x its damage with no flat bonus (0.9965 to 0.9994 at 1
		// to 4 targets). Classic keeps 2.5.
		ThreatMultiplier: core.TernaryFloat64(warrior.Env.IsForever(), 1, 2.5),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Only living targets, each one once.
			targets := warrior.areaTargets(sim, target, thunderClapRadius, len(results))
			hits := results[:len(targets)]
			for idx, target := range targets {
				results[idx] = spell.CalcDamage(sim, target, baseDamage+apCoef*spell.MeleeAttackPower(target), spell.OutcomeMagicHitAndCrit)
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
