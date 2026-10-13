package hunter

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (hunter *Hunter) registerVolleySpell() {
	for rank := 3; rank >= 1; rank-- {
		config := hunter.getVolleyConfig(rank)

		if config.RequiredLevel <= int(hunter.Level) {
			hunter.Volley = hunter.GetOrRegisterSpell(config)
			return
		}
	}
}

// Forever's Volley (beta client, build 1.60.1.70338) has no cooldown (Classic 1 min) and hits
// harder: 70/91/112 a second by rank (Classic 50/65/80), from the spells it fires each tick
// (1279721, 1279719 and 1279715). Each tick is a ranged attack that can crit, so Mortal Shots and
// Barrage count. The spell power coefficient is Classic's, a hunter has no spell power.
func (hunter *Hunter) getVolleyConfig(rank int) core.SpellConfig {
	spellId := [4]int32{0, 1510, 14294, 14295}[rank]
	baseDamage := [4]float64{0, 70, 91, 112}[rank]
	manaCost := [4]float64{0, 350, 420, 490}[rank]
	level := [4]int{0, 40, 50, 58}[rank]

	return core.SpellConfig{
		SpellCode:   SpellCode_HunterVolley,
		ActionID:    core.ActionID{SpellID: spellId},
		SpellSchool: core.SpellSchoolArcane,
		DefenseType: core.DefenseTypeRanged,
		ProcMask:    core.ProcMaskRangedSpecial,
		Flags:       core.SpellFlagChanneled | core.SpellFlagAPL,

		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.inRangedRange()
		},

		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				Label: fmt.Sprintf("Volley (Rank %d)", rank),
			},
			NumberOfTicks:    6,
			TickLength:       time.Second * 1,
			BonusCoefficient: .056,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				for _, aoeTarget := range sim.Encounter.TargetUnits {
					dot.CalcAndDealPeriodicSnapshotDamage(sim, aoeTarget, dot.OutcomeTickPhysicalCrit)
				}
			},
		},

		CritDamageBonus:  hunter.mortalShots(),
		DamageMultiplier: hunter.barrageMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			hunter.Unit.AutoAttacks.DelayRangedUntil(sim, sim.CurrentTime+(time.Second*6))
			spell.AOEDot().Apply(sim)
		},
	}
}
