package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Wrack is the Affliction capstone (1316697, Forever's own spell): a 6 sec channel of 36
// Shadow damage a second with a 14.3% coefficient per tick, 200 mana, and while it is on the
// target the warlock's Corruption and Bane of Agony deal 10% more (its second effect, aura
// 271 on class mask 0x402). The client has no ranks above the one the talent teaches.
func (warlock *Warlock) registerWrackSpell() {
	if !warlock.Talents.Wrack || warlock.Level < 40 {
		return
	}

	actionID := core.ActionID{SpellID: 1316697}

	warlock.Wrack = warlock.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		SpellCode:   SpellCode_WarlockWrack,
		SpellSchool: core.SpellSchoolShadow,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       core.SpellFlagAPL | core.SpellFlagChanneled | core.SpellFlagResetAttackSwing | WarlockFlagAffliction,

		RequiredLevel: 40,

		ManaCost: core.ManaCostOptions{
			FlatCost: 200,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Wrack-" + warlock.Label,
			},
			NumberOfTicks:    6,
			TickLength:       time.Second,
			BonusCoefficient: 0.143,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, 36, isRollover)
				dot.SnapshotAttackerMultiplier *= warlock.soulSiphonMultiplier(target)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
			spell.DealOutcome(sim, result)
		},
		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			if useSnapshot {
				return spell.Dot(target).CalcSnapshotDamage(sim, target, spell.OutcomeExpectedMagicAlwaysHit)
			}
			return spell.CalcPeriodicDamage(sim, target, 36, spell.OutcomeExpectedMagicAlwaysHit)
		},
	})
}

// wrackActive is whether our Wrack is being channelled on the target.
func (warlock *Warlock) wrackActive(target *core.Unit) bool {
	return warlock.Wrack != nil && warlock.Wrack.Dot(target).IsActive()
}
