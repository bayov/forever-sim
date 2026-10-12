package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Deadly Aspects (Forever's Improved Aspect of the Hawk): 2% a rank (Classic 1%) for Auto Shot to give
// Quick Shots (6150, 30% ranged attack speed for 12 sec) while Aspect of the Hawk is up, and the same
// chance for a melee auto attack to give Quick Strikes (1299448, 30% melee attack speed for 12 sec)
// while Aspect of the Beast is up.
func (hunter *Hunter) deadlyAspectsChance() float64 {
	return 0.02 * float64(hunter.Talents.DeadlyAspects)
}

func (hunter *Hunter) getMaxHawkRank() int {
	levels := [8]int{0, 10, 18, 28, 38, 48, 58, 60}
	for i := 7; i > 0; i-- {
		if levels[i] <= int(hunter.Level) {
			return i
		}
	}
	return 0
}

func (hunter *Hunter) registerAspectOfTheHawkSpell() {
	rank := hunter.getMaxHawkRank()
	if rank == 0 {
		return
	}

	hunter.AspectOfTheHawkAPMultiplier = 1.0

	// The beta client gives rank 6 55 ranged attack power, half of Classic's 110 and a step down from
	// rank 5's 90. Rank 7 at level 60 is Classic's 120.
	spellID := [8]int32{0, 13165, 14318, 14319, 14320, 14321, 14322, 25296}[rank]
	rap := [8]float64{0, 20, 35, 50, 70, 90, 55, 120}[rank]
	manaCost := [8]float64{0, 20, 35, 50, 70, 90, 110, 120}[rank]

	var quickShots *core.Aura
	if hunter.Talents.DeadlyAspects > 0 {
		quickShots = hunter.RegisterAura(core.Aura{
			Label:    "Quick Shots",
			ActionID: core.ActionID{SpellID: 6150},
			Duration: time.Second * 12,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				aura.Unit.MultiplyRangedSpeed(sim, 1.3)
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				aura.Unit.MultiplyRangedSpeed(sim, 1/1.3)
			},
		})
	}
	procChance := hunter.deadlyAspectsChance()

	actionID := core.ActionID{SpellID: spellID}
	hunter.AspectOfTheHawkAura = hunter.RegisterAura(core.Aura{
		Label:    "Aspect of the Hawk",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatDynamic(sim, stats.RangedAttackPower, rap*hunter.AspectOfTheHawkAPMultiplier)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatDynamic(sim, stats.RangedAttackPower, -rap*hunter.AspectOfTheHawkAPMultiplier)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if quickShots == nil || !spell.ProcMask.Matches(core.ProcMaskRangedAuto) {
				return
			}
			if sim.Proc(procChance, "Deadly Aspects") {
				quickShots.Activate(sim)
			}
		},
	})
	hunter.AspectOfTheHawkAura.NewExclusiveEffect("Aspect", true, core.ExclusiveEffect{})

	hunter.RegisterSpell(core.SpellConfig{
		ActionID:      actionID,
		Flags:         core.SpellFlagAPL,
		Rank:          rank,
		RequiredLevel: 10,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !hunter.AspectOfTheHawkAura.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			hunter.AspectOfTheHawkAura.Activate(sim)
		},
	})
}

// Forever's Aspect of the Beast has four ranks and adds melee attack power: 50/70/90/110 at levels
// 30/40/50/60 (beta client 13161, 1299445 to 1299447, build 1.60.1.70338). Classic's only made the
// hunter untrackable.
func (hunter *Hunter) registerAspectOfTheBeastSpell() {
	rank := 0
	for i := 4; i >= 1; i-- {
		if []int{0, 30, 40, 50, 60}[i] <= int(hunter.Level) {
			rank = i
			break
		}
	}
	if rank == 0 {
		return
	}

	spellID := [5]int32{0, 13161, 1299445, 1299446, 1299447}[rank]
	ap := [5]float64{0, 50, 70, 90, 110}[rank]
	manaCost := [5]float64{0, 50, 70, 90, 110}[rank]

	var quickStrikes *core.Aura
	if hunter.Talents.DeadlyAspects > 0 {
		quickStrikes = hunter.RegisterAura(core.Aura{
			Label:    "Quick Strikes",
			ActionID: core.ActionID{SpellID: 1299448},
			Duration: time.Second * 12,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				aura.Unit.MultiplyMeleeSpeed(sim, 1.3)
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				aura.Unit.MultiplyMeleeSpeed(sim, 1/1.3)
			},
		})
	}
	procChance := hunter.deadlyAspectsChance()

	actionID := core.ActionID{SpellID: spellID}
	hunter.AspectOfTheBeastAura = hunter.RegisterAura(core.Aura{
		Label:    "Aspect of the Beast",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatDynamic(sim, stats.AttackPower, ap)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatDynamic(sim, stats.AttackPower, -ap)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if quickStrikes == nil || !spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
				return
			}
			if sim.Proc(procChance, "Deadly Aspects") {
				quickStrikes.Activate(sim)
			}
		},
	})
	hunter.AspectOfTheBeastAura.NewExclusiveEffect("Aspect", true, core.ExclusiveEffect{})

	hunter.RegisterSpell(core.SpellConfig{
		ActionID:      actionID,
		Flags:         core.SpellFlagAPL,
		Rank:          rank,
		RequiredLevel: 30,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !hunter.AspectOfTheBeastAura.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			hunter.AspectOfTheBeastAura.Activate(sim)
		},
	})
}
