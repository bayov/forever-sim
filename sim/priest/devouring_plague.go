package priest

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const DevouringPlagueRanks = 6

// Forever lets every race learn Devouring Plague and puts it on a 1 min cooldown (Classic
// 3 min, Undead only). Damage per tick and the 10% a tick coefficient (80% over the 8
// ticks) are from the client (1.60.1.70338, wago.tools): rank 6 deals 848 over 24 sec,
// Classic 904. The damage heals the priest, which the sim leaves out.
var DevouringPlagueSpellId = [DevouringPlagueRanks + 1]int32{0, 2944, 19276, 19277, 19278, 19279, 19280}
var DevouringPlagueTickDamage = [DevouringPlagueRanks + 1]float64{0, 16, 29, 43, 61, 82, 106}
var DevouringPlagueManaCost = [DevouringPlagueRanks + 1]float64{0, 215, 350, 495, 645, 810, 985}
var DevouringPlagueLevel = [DevouringPlagueRanks + 1]int{0, 20, 28, 36, 44, 52, 60}

const DevouringPlagueTickCoef = 0.1
const DevouringPlagueCooldown = time.Minute

func (priest *Priest) registerDevouringPlagueSpell() {
	priest.DevouringPlague = make([]*core.Spell, DevouringPlagueRanks+1)
	cdTimer := priest.NewTimer()

	for rank := 1; rank <= DevouringPlagueRanks; rank++ {
		config := priest.getDevouringPlagueConfig(rank, cdTimer)

		if config.RequiredLevel <= int(priest.Level) {
			priest.DevouringPlague[rank] = priest.GetOrRegisterSpell(config)
		}
	}
}

func (priest *Priest) getDevouringPlagueConfig(rank int, cdTimer *core.Timer) core.SpellConfig {
	var ticks int32 = 8

	spellId := DevouringPlagueSpellId[rank]
	baseDotDamage := DevouringPlagueTickDamage[rank]
	manaCost := DevouringPlagueManaCost[rank]
	level := DevouringPlagueLevel[rank]

	return core.SpellConfig{
		SpellCode:   SpellCode_PriestDevouringPlague,
		ActionID:    core.ActionID{SpellID: spellId},
		SpellSchool: core.SpellSchoolShadow,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagPriest | core.SpellFlagAPL | core.SpellFlagDisease | core.SpellFlagPureDot,

		Rank:          rank,
		RequiredLevel: level,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: DevouringPlagueCooldown,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: fmt.Sprintf("Devouring Plague (Rank %d)", rank),
			},

			NumberOfTicks:    ticks,
			TickLength:       time.Second * 3,
			BonusCoefficient: DevouringPlagueTickCoef,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDotDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
			if result.Landed() {
				priest.AddShadowWeavingStack(sim, target)
				spell.Dot(target).Apply(sim)
			}
			spell.DealOutcome(sim, result)
		},
	}
}
