package priest

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

const DarkSacrificeRanks = 5

// Dark Sacrifice is a new Undead priest spell in Forever (SkillLineAbility race mask 16):
// instant, 10 min cooldown, the priest gives up health over 15 sec for mana. The client
// (1.60.1.70338, wago.tools) has 5 ticks of 3 sec of a health drain (aura 226) and a mana
// gain (aura 24) of the same size, 320 a tick at 60 for rank 5. The tooltip adds the
// priest's Spirit to the mana ("${$o2+$SPI}"), which the server must do, so the sim gives
// 1600 plus Spirit, the Spirit part with the first tick. The health loss is left out.
var DarkSacrificeSpellId = [DarkSacrificeRanks + 1]int32{0, 1277324, 1277325, 1277326, 1277327, 1277328}
var DarkSacrificeTickMana = [DarkSacrificeRanks + 1]float64{0, 88, 144, 204, 266, 320}
var DarkSacrificeLevel = [DarkSacrificeRanks + 1]int{0, 20, 30, 40, 50, 60}

func (priest *Priest) registerDarkSacrificeSpell() {
	if priest.Race != proto.Race_RaceUndead {
		return
	}

	rank := core.HighestRankAt(priest.Level, DarkSacrificeLevel[:])
	if rank == 0 {
		return
	}

	actionID := core.ActionID{SpellID: DarkSacrificeSpellId[rank]}
	manaMetrics := priest.NewManaMetrics(actionID)
	tickMana := DarkSacrificeTickMana[rank]

	priest.DarkSacrifice = priest.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolShadow,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagHelpful | core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    priest.NewTimer(),
				Duration: time.Minute * 10,
			},
		},

		Hot: core.DotConfig{
			Aura: core.Aura{
				Label: "Dark Sacrifice",
			},
			SelfOnly:      true,
			NumberOfTicks: 5,
			TickLength:    time.Second * 3,
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				mana := tickMana
				if dot.TickCount == 1 {
					mana += priest.GetStat(stats.Spirit)
				}
				priest.AddMana(sim, mana, manaMetrics)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.SelfHot().Apply(sim)
		},
	})
}
