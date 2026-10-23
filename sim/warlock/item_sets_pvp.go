package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

var bluePvPBonuses = map[int32]core.ApplyEffect{
	// Increases damage and healing done by magical spells and effects by up to 23.
	2: func(agent core.Agent) {
		c := agent.GetCharacter()
		c.AddStat(stats.SpellPower, 23)
	},
	// Reduces the casting time of your Immolate spell by 0.2 sec.
	4: func(agent core.Agent) {
		warlock := agent.(WarlockAgent).GetWarlock()
		warlock.GetOrRegisterAura(core.Aura{
			Label: "Immolate Cast Time Reduction",
			OnInit: func(aura *core.Aura, sim *core.Simulation) {
				for _, spell := range warlock.Immolate {
					spell.DefaultCast.CastTime -= time.Millisecond * 200
					spell.DefaultCast.GCD -= time.Millisecond * 200
				}
			},
		})
	},
	// +20 Stamina.
	6: func(agent core.Agent) {
		c := agent.GetCharacter()
		c.AddStat(stats.Stamina, 20)
	},
}

// Forever's warlock PvP sets keep the Classic rows (client ItemSetSpell, 1.60.1.70532). Warlord's and
// Field Marshal's Threads are new pieces that carry only the set ID; Champion's and Lieutenant
// Commander's Dreadgear keep their Classic pieces.
var ItemSetChampionsThreads = core.NewItemSet(core.ItemSet{
	ID:      541,
	Name:    "Champion's Dreadgear",
	Bonuses: bluePvPBonuses,
})

var ItemSetLieutenantCommandersThreads = core.NewItemSet(core.ItemSet{
	ID:      547,
	Name:    "Lieutenant Commander's Dreadgear",
	Bonuses: bluePvPBonuses,
})

var epicPvpBonuses = map[int32]core.ApplyEffect{
	// +20 Stamina.
	2: func(agent core.Agent) {
		c := agent.GetCharacter()
		c.AddStat(stats.Stamina, 20)
	},
	// Reduces the casting time of your Immolate spell by 0.2 sec.
	3: func(agent core.Agent) {
		warlock := agent.(WarlockAgent).GetWarlock()
		warlock.GetOrRegisterAura(core.Aura{
			Label: "Immolate Cast Time Reduction",
			OnInit: func(aura *core.Aura, sim *core.Simulation) {
				for _, spell := range warlock.Immolate {
					spell.DefaultCast.CastTime -= time.Millisecond * 200
					spell.DefaultCast.GCD -= time.Millisecond * 200
				}
			},
		})
	},
	// Increases damage and healing done by magical spells and effects by up to 23.
	6: func(agent core.Agent) {
		c := agent.GetCharacter()
		c.AddStat(stats.SpellPower, 23)
	},
}

var ItemSetWarlordsThreads = core.NewItemSet(core.ItemSet{
	ID:      391,
	Name:    "Warlord's Threads",
	Bonuses: epicPvpBonuses,
})

var ItemSetFieldMarshalsThreads = core.NewItemSet(core.ItemSet{
	ID:      392,
	Name:    "Field Marshal's Threads",
	Bonuses: epicPvpBonuses,
})

// Forever's Champion's Threads (345) and Lieutenant Commander's Threads (346), on new pieces, with
// bonuses under Forever only since the sim had no Classic version of them.
var ItemSetForeverChampionsThreads = core.NewItemSet(core.ItemSet{
	ID:      345,
	Name:    "Champion's Threads",
	Bonuses: core.RulesetSetBonuses(nil, bluePvPBonuses),
})

var ItemSetForeverLieutenantCommandersThreads = core.NewItemSet(core.ItemSet{
	ID:      346,
	Name:    "Lieutenant Commander's Threads",
	Bonuses: core.RulesetSetBonuses(nil, bluePvPBonuses),
})
