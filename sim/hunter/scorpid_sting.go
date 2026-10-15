package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// Forever's Scorpid Sting as the hunter's own cast, for the presets on Scorpid Sting duty (standards
// 23): they switch the raid's Scorpid Sting debuff off and keep the sting up themselves, in the sting
// slot in place of Serpent Sting. Forever redesigned it: 2% off the target's chance to hit for 20 sec
// (core.ForeverScorpidStingAura), one rank learned at 22. The beta tooltip reads 94 mana at level 38,
// 9% of that level's base mana by Multi-Shot's 13.9% and 145, so it costs 9% of base mana. Improved
// Stings makes it 15 / 30 / 45 sec longer, and Efficiency's sting discount is taken to cover it as it
// covers Serpent Sting. A hunter has one sting on a target, so each sting removes the other. Classic's
// Scorpid Sting (Strength and Agility) is not simulated.
func (hunter *Hunter) registerScorpidStingSpell() {
	if !hunter.Env.IsForever() || hunter.Level < 22 {
		return
	}

	duration := time.Second * time.Duration(20+15*hunter.Talents.ImprovedStings)
	hunter.scorpidStingAuras = hunter.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.ForeverScorpidStingAura(target, duration)
	})

	hunter.ScorpidSting = hunter.RegisterSpell(core.SpellConfig{
		SpellCode:     SpellCode_HunterScorpidSting,
		ActionID:      core.ActionID{SpellID: 3043},
		SpellSchool:   core.SpellSchoolNature,
		DefenseType:   core.DefenseTypeRanged,
		ProcMask:      core.ProcMaskRangedSpecial,
		Flags:         core.SpellFlagAPL | core.SpellFlagPoison | SpellFlagSting,
		CastType:      proto.CastType_CastTypeRanged,
		Rank:          1,
		RequiredLevel: 22,
		MissileSpeed:  40,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.09,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true, // Hunter GCD is locked at 1.5s
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.inRangedRange()
		},

		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeRangedHitNoHitCounter)

			spell.WaitTravelTime(sim, func(s *core.Simulation) {
				spell.DealOutcome(sim, result)

				if result.Landed() {
					if hunter.SerpentSting != nil {
						hunter.SerpentSting.Dot(target).Deactivate(sim)
					}
					hunter.scorpidStingAuras.Get(target).Activate(sim)
				}
			})
		},
	})
}
