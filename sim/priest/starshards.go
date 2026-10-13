package priest

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

const StarshardsRanks = 7
const StarshardsTicks = 6

// Night Elf only (SkillLineAbility race mask 8). Forever roughly doubled its damage and
// gave it a 30 sec cooldown (foreverchanges.pro): rank 7 deals 1800 Arcane over a 6 sec
// channel, Classic 936 with no cooldown. Damage per tick and the 16.7% a tick
// coefficient (100% over the channel) are from the client (1.60.1.70338, wago.tools).
var StarshardsSpellId = [StarshardsRanks + 1]int32{0, 10797, 19296, 19299, 19302, 19303, 19304, 19305}
var StarshardsTickDamage = [StarshardsRanks + 1]float64{0, 27, 50, 88, 127, 178, 240, 300}
var StarshardsManaCost = [StarshardsRanks + 1]float64{0, 50, 85, 140, 190, 245, 300, 350}
var StarshardsLevel = [StarshardsRanks + 1]int{0, 10, 18, 26, 34, 42, 50, 58}

const StarshardsTickCoef = 0.167

func (priest *Priest) registerStarshardsSpell() {
	if priest.Race != proto.Race_RaceNightElf {
		return
	}

	priest.Starshards = make([]*core.Spell, StarshardsRanks+1)
	cdTimer := priest.NewTimer()

	for rank := 1; rank <= StarshardsRanks; rank++ {
		if StarshardsLevel[rank] <= int(priest.Level) {
			priest.Starshards[rank] = priest.RegisterSpell(priest.newStarshardsSpellConfig(rank, cdTimer))
		}
	}
}

func (priest *Priest) newStarshardsSpellConfig(rank int, cdTimer *core.Timer) core.SpellConfig {
	baseDamage := StarshardsTickDamage[rank]

	return core.SpellConfig{
		ActionID:    core.ActionID{SpellID: StarshardsSpellId[rank]},
		SpellCode:   SpellCode_PriestStarshards,
		SpellSchool: core.SpellSchoolArcane,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagPriest | core.SpellFlagChanneled | core.SpellFlagAPL,

		RequiredLevel: StarshardsLevel[rank],
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: StarshardsManaCost[rank],
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: time.Second * 30,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: fmt.Sprintf("Starshards-%d", rank),
			},
			NumberOfTicks:    StarshardsTicks,
			TickLength:       time.Second,
			BonusCoefficient: StarshardsTickCoef,
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
				spell.Dot(target).Apply(sim)
			}
			spell.DealOutcome(sim, result)
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
			return spell.CalcPeriodicDamage(sim, target, baseDamage, spell.OutcomeExpectedMagicAlwaysHit)
		},
	}
}
