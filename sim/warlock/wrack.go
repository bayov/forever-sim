package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Wrack is the Affliction capstone (1316697, Forever's own spell): a 6 sec channel of Shadow
// damage every second, 200 mana. While it is channelled on the target, the warlock's Corruption,
// Siphon Life and Bane of Agony there tick twice as fast (its second effect, SpellEffect 1349557,
// aura 4 value 100 since hotfix 113766; see applyWrack). The client has no ranks above the one the
// talent teaches.
//
// Server hotfix 113766 (build 1.60.1.70888, 2026-11-10, "Wrack damage reduced by 30%") sets
// SpellEffect 1349556 to base 25.2 a second, 1.75 more a level from 40 (60.2 at 60), and a 0.2002
// coefficient a second: the 2026-11-02 notes' 36 + 2.5 a level and 0.286, times 0.7.
const (
	wrackBaseDamage     = 25.2
	wrackDamagePerLevel = 1.75
	wrackCoefficient    = 0.2002
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
