package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Wrack is the Affliction capstone (1316697, Forever's own spell): a 6 sec channel of Shadow
// damage every second, 200 mana, and while it is on the target the warlock's Corruption and Bane
// of Agony deal 10% more (its second effect, aura 271 on class mask 0x402). The client has no
// ranks above the one the talent teaches.
//
// The client (1.60.1.70753, the same as 70707) still has 36 a second and a 14.3% coefficient a
// tick. Blizzard's development notes of 2026-11-02 double the coefficient to 28.6% a tick (0.286 a
// second) and make the base grow 2.5 a level, 36 at 40 to 86 at 60 (foreverchanges.pro reads it
// the same way). The sim takes the notes: the change is server side, like most of that round's.
const (
	wrackBaseDamage     = 36.0
	wrackDamagePerLevel = 2.5
	wrackCoefficient    = 0.286
)

func wrackTickDamage(level int32) float64 {
	return wrackBaseDamage + wrackDamagePerLevel*float64(level-40)
}

func (warlock *Warlock) registerWrackSpell() {
	if !warlock.Talents.Wrack || warlock.Level < 40 {
		return
	}

	baseDamage := wrackTickDamage(warlock.Level)

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
			BonusCoefficient: wrackCoefficient,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDamage, isRollover)
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
			return spell.CalcPeriodicDamage(sim, target, baseDamage, spell.OutcomeExpectedMagicAlwaysHit)
		},
	})
}

// wrackActive is whether our Wrack is being channelled on the target.
func (warlock *Warlock) wrackActive(target *core.Unit) bool {
	return warlock.Wrack != nil && warlock.Wrack.Dot(target).IsActive()
}
