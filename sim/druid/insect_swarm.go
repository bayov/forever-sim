package druid

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const InsectSwarmRanks = 5

var InsectSwarmSpellId = [InsectSwarmRanks + 1]int32{0, 5570, 24974, 24975, 24976, 24977}

// Damage per tick from the Forever client (1.60.1.70338, wago.tools), about 60% of 1.12
// (rank 5 does 186 over 12 sec, 1.12 has 324). Insect Swarm moved to the Balance tree.
var InsectSwarmTickDamage = [InsectSwarmRanks + 1]float64{0, 8, 15, 20, 25, 31}
var InsectSwarmManaCost = [InsectSwarmRanks + 1]float64{0, 45, 85, 100, 140, 160}
var InsectSwarmLevel = [InsectSwarmRanks + 1]int{0, 20, 30, 40, 50, 60}

func (druid *Druid) registerInsectSwarmSpell() {
	druid.InsectSwarm = make([]*DruidSpell, InsectSwarmRanks+1)
	if !druid.Talents.InsectSwarm {
		return
	}

	druid.InsectSwarmAuras = druid.NewEnemyAuraArray(core.InsectSwarmAura)

	for rank := 1; rank <= InsectSwarmRanks; rank++ {
		level := InsectSwarmLevel[rank]
		if int32(level) <= druid.Level {
			// 12 sec in 2 sec ticks. Nature's Splendor makes it 2 sec longer and the
			// Grovekeeper Eclipse 5 piece 3 sec, and a part of a tick left over does nothing.
			duration := time.Second*12 + druid.InsectSwarmBonusDuration
			if druid.Talents.NaturesSplendor {
				duration += time.Second * 2
			}
			numTicks := int32(duration / (time.Second * 2))
			tickLength := time.Second * 2

			spellID := InsectSwarmSpellId[rank]
			baseDamage := InsectSwarmTickDamage[rank]
			manaCost := InsectSwarmManaCost[rank]
			spellCoef := .158

			druid.InsectSwarm[rank] = druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
				SpellCode:   SpellCode_DruidInsectSwarm,
				ActionID:    core.ActionID{SpellID: spellID},
				SpellSchool: core.SpellSchoolNature,
				DefenseType: core.DefenseTypeMagic,
				ProcMask:    core.ProcMaskSpellDamage,
				Flags:       SpellFlagOmen | core.SpellFlagAPL | core.SpellFlagBinary,

				ManaCost: core.ManaCostOptions{
					FlatCost: manaCost,
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
						Label: fmt.Sprintf("Insect Swarm (Rank %d)", rank),
						OnGain: func(aura *core.Aura, sim *core.Simulation) {
							druid.InsectSwarmAuras.Get(aura.Unit).Activate(sim)
						},
						OnExpire: func(aura *core.Aura, sim *core.Simulation) {
							insectSwarmAura := druid.InsectSwarmAuras.Get(aura.Unit)
							if !insectSwarmAura.IsPermanent() {
								insectSwarmAura.Deactivate(sim)
							}
						},
					},

					NumberOfTicks:    numTicks,
					TickLength:       tickLength,
					BonusCoefficient: spellCoef,

					OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
						dot.Snapshot(target, baseDamage, isRollover)
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

				RelatedAuras: []core.AuraArray{druid.InsectSwarmAuras},
			})
		}
	}
}
