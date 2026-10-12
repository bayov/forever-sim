package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// Forever trains Aimed Shot at level 20 instead of making it a talent. The beta client (build
// 1.60.1.70338) gives every rank a 2 sec cast (Classic 3) and a much smaller bonus (rank 6: 166,
// Classic 600). The mana costs and the 6 sec cooldown are Classic's, and the cooldown is shared
// with Multi-Shot (client category 2 on both).
func (hunter *Hunter) getAimedShotConfig(rank int, timer *core.Timer) core.SpellConfig {
	spellId := [7]int32{0, 19434, 20900, 20901, 20902, 20903, 20904}[rank]
	baseDamage := [7]float64{0, 20, 34, 55, 89, 125, 166}[rank]
	manaCost := [7]float64{0, 75, 115, 160, 210, 260, 310}[rank]
	level := [7]int{0, 20, 28, 36, 44, 52, 60}[rank]

	return core.SpellConfig{
		SpellCode:     SpellCode_HunterAimedShot,
		ActionID:      core.ActionID{SpellID: spellId},
		SpellSchool:   core.SpellSchoolPhysical,
		DefenseType:   core.DefenseTypeRanged,
		ProcMask:      core.ProcMaskRangedSpecial,
		Flags:         core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagShot,
		CastType:      proto.CastType_CastTypeRanged,
		Rank:          rank,
		RequiredLevel: level,
		MissileSpeed:  40,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Second * 2,
			},
			CD: core.Cooldown{
				Timer:    timer,
				Duration: time.Second * 6,
			},
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				cast.CastTime = spell.CastTime()
			},
			IgnoreHaste: true, // Hunter GCD is locked at 1.5s
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
			baseDamage := hunter.AutoAttacks.Ranged().CalculateNormalizedWeaponDamage(sim, spell.RangedAttackPower(target, false)) +
				hunter.NormalizedAmmoDamageBonus +
				baseDamage

			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndCrit)
			spell.WaitTravelTime(sim, func(s *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	}
}

func (hunter *Hunter) registerAimedShotSpell(timer *core.Timer) {
	for i := 6; i >= 1; i-- {
		config := hunter.getAimedShotConfig(i, timer)

		if config.RequiredLevel <= int(hunter.Level) {
			hunter.AimedShot = hunter.GetOrRegisterSpell(config)
			return
		}
	}
}
