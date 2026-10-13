package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// Sniper Shot, the 31 point Marksmanship talent (beta client 1310687, 1310785 and 1310786, build
// 1.60.1.70338): normalized ranged weapon damage plus 160/225/295 by rank, a 4 sec cast, a 15 sec
// cooldown of its own and 365 mana at every rank. Its range bonus to the next 3 shots does nothing
// here. Mortal Shots, Barrage and Efficiency leave it out of their class masks.
func (hunter *Hunter) registerSniperShotSpell() {
	if !hunter.Talents.SniperShot {
		return
	}

	rank := 1
	if hunter.Level >= 58 {
		rank = 3
	} else if hunter.Level >= 48 {
		rank = 2
	}
	spellID := [4]int32{0, 1310687, 1310785, 1310786}[rank]
	flatBonus := [4]float64{0, 160, 225, 295}[rank]

	hunter.SniperShot = hunter.RegisterSpell(core.SpellConfig{
		SpellCode:     SpellCode_HunterSniperShot,
		ActionID:      core.ActionID{SpellID: spellID},
		SpellSchool:   core.SpellSchoolPhysical,
		DefenseType:   core.DefenseTypeRanged,
		ProcMask:      core.ProcMaskRangedSpecial,
		Flags:         core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagShot,
		CastType:      proto.CastType_CastTypeRanged,
		Rank:          rank,
		RequiredLevel: 40,
		MissileSpeed:  60,

		ManaCost: core.ManaCostOptions{
			FlatCost: 365,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Second * 4,
			},
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				cast.CastTime = spell.CastTime()
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Second * 15,
			},
			CastTime: func(spell *core.Spell) time.Duration {
				return time.Duration(float64(spell.DefaultCast.CastTime) / hunter.RangedSwingSpeed())
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.inRangedRange()
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := hunter.AutoAttacks.Ranged().CalculateNormalizedWeaponDamage(sim, spell.RangedAttackPower(target, false)) +
				hunter.NormalizedAmmoDamageBonus +
				flatBonus

			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndCrit)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})
}
