package priest

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const MindFlayRanks = 6
const MindFlayTicks = 3

// Damage per tick and coefficient from the Forever client (1.60.1.70338, wago.tools
// SpellEffect): rank 6 deals 390 over 3 sec (Classic 426), and every rank has 16.7% a
// tick (Classic 15%).
var MindFlaySpellId = [MindFlayRanks + 1]int32{0, 15407, 17311, 17312, 17313, 17314, 18807}
var MindFlayTickDamage = [MindFlayRanks + 1]float64{0, 21, 34, 51, 75, 98, 130}
var MindFlayManaCost = [MindFlayRanks + 1]float64{0, 45, 70, 100, 135, 165, 205}
var MindFlayLevel = [MindFlayRanks + 1]int{0, 20, 28, 36, 44, 52, 60}

const MindFlayTickCoef = 0.167

func (priest *Priest) registerMindFlay() {
	if !priest.Talents.MindFlay {
		return
	}

	priest.MindFlay = make([]*core.Spell, MindFlayRanks+1)

	for rank := 1; rank <= MindFlayRanks; rank++ {
		if MindFlayLevel[rank] <= int(priest.Level) {
			priest.MindFlay[rank] = priest.RegisterSpell(priest.newMindFlaySpellConfig(rank))
		}
	}
}

// Mind Flay is one channel with a single hit roll when it starts (it slows, so it is
// binary). The APL's channel action can clip it after any tick.
func (priest *Priest) newMindFlaySpellConfig(rank int) core.SpellConfig {
	baseDamage := MindFlayTickDamage[rank]

	return core.SpellConfig{
		SpellCode:   SpellCode_PriestMindFlay,
		ActionID:    core.ActionID{SpellID: MindFlaySpellId[rank]},
		SpellSchool: core.SpellSchoolShadow,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagPriest | core.SpellFlagChanneled | core.SpellFlagBinary | core.SpellFlagAPL,

		RequiredLevel: MindFlayLevel[rank],
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: MindFlayManaCost[rank],
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
				Label: fmt.Sprintf("MindFlay-%d", rank),
			},
			NumberOfTicks:       MindFlayTicks,
			TickLength:          time.Second,
			AffectedByCastSpeed: false,
			BonusCoefficient:    MindFlayTickCoef,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				priest.AddShadowWeavingStack(sim, target)
				spell.Dot(target).Apply(sim)
			}
			spell.DealOutcome(sim, result)
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
			return spell.CalcPeriodicDamage(sim, target, baseDamage, spell.OutcomeExpectedMagicAlwaysHit)
		},
	}
}
