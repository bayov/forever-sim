package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (warrior *Warrior) registerSlamSpell() {
	requiredLevel := 54
	spellID := int32(11605)
	flatDamageBonus := 87.0

	// Improved Slam now takes the same amount off the global cooldown as it does the cast time,
	// and stops Slam from resetting the swing timer entirely.
	castReduction := time.Millisecond * 250 * time.Duration(warrior.Talents.ImprovedSlam)

	// Forever gives Slam an 18 sec cooldown, and Improved Slam takes 3 sec off it (client
	// 70338: category cooldown 18000 on every rank, and effect 3 of spell 12862 is -3000 with no
	// rank curve, so 3 sec at either rank. ForeverChanges reads it as 3 sec a point, but the trait
	// curves override only the effects they name).
	var cooldown core.Cooldown
	if warrior.Env.IsForever() {
		cooldown = core.Cooldown{
			Timer:    warrior.NewTimer(),
			Duration: time.Second*18 - core.TernaryDuration(warrior.Talents.ImprovedSlam > 0, time.Second*3, 0),
		}
	}

	// Under Forever the core resets the swing when Slam's cast completes. Improved Slam's Slam
	// (1310196-1310200) is marked Do Not Reset Combat Timers (Attributes[2] 0x20000), the
	// base ranks are not (wowsims/forever 4fb2849945).
	flags := core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagOffensive
	if warrior.Talents.ImprovedSlam > 0 {
		flags |= core.SpellFlagKeepsSwingTimer
	}

	warrior.Slam = warrior.RegisterSpell(AnyStance, core.SpellConfig{
		SpellCode:   SpellCode_WarriorSlam,
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       flags,

		RequiredLevel: requiredLevel,

		RageCost: core.RageCostOptions{
			Cost:   15,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      max(core.GCDDefault-castReduction, core.GCDMin),
				CastTime: time.Millisecond*1500 - castReduction,
			},
			CD: cooldown,
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				if !warrior.Env.IsForever() && warrior.Talents.ImprovedSlam == 0 && spell.CastTime() > 0 {
					warrior.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+cast.CastTime, true)
				}
			},
		},

		// "Slam is no longer subject to spell pushback" (development notes, 2026-10-15, build
		// 1.60.1.70425). The raid hits (sim/core/encounters) would otherwise push it back.
		PushbackReduction: core.TernaryFloat64(warrior.Env.IsForever(), 1, 0),

		CritDamageBonus: warrior.impale(),

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		FlatThreatBonus:  140, // Should this be 54 or the old 140 value from before SoD?
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := flatDamageBonus + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))

			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			if !result.Landed() {
				spell.IssueRefund(sim)
			}
		},
	})
}
