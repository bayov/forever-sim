package rogue

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's rogue PvP sets, by the client's set IDs (ItemSetSpell, 1.60.1.70532). Warlord's and Field
// Marshal's Vestments are new pieces that carry only the set ID and keep the Classic rows, and
// Lieutenant Commander's Guard keeps its Classic pieces. Champion's Guard is in no Forever set.
// The 3- or 4-piece bonus shortens Gouge's cooldown.
func init() {
	attackPower := core.SetBonusStats(stats.Stats{stats.AttackPower: 40, stats.RangedAttackPower: 40}) // 14049
	stamina := core.SetBonusStats(stats.Stats{stats.Stamina: 20})                                      // 14467
	// The Forever sets with no Classic counterpart in the sim, so with bonuses under Forever only.
	// Champion's Vestments is two sets: 347 on the Classic pieces, with 1% parry (13665) and 15
	// Stamina (14462), and 522 on new ones.
	for _, set := range []struct {
		id      int32
		name    string
		bonuses map[int32]core.ApplyEffect
	}{
		{347, "Champion's Vestments", map[int32]core.ApplyEffect{
			2: core.SetBonusStats(stats.Stats{stats.Parry: 1 * core.ParryRatingPerParryChance}),
			6: core.SetBonusStats(stats.Stats{stats.Stamina: 15}),
		}},
		{522, "Champion's Vestments", map[int32]core.ApplyEffect{2: attackPower, 6: stamina}},
		{348, "Lieutenant Commander's Vestments", map[int32]core.ApplyEffect{2: attackPower, 6: stamina}},
	} {
		core.NewItemSet(core.ItemSet{ID: set.id, Name: set.name, Bonuses: core.RulesetSetBonuses(nil, set.bonuses)})
	}
}

///////////////////////////////////////////////////////////////////////////
//                          Classic Phase 2
///////////////////////////////////////////////////////////////////////////

var ItemSetChampionsGuard = core.NewItemSet(core.ItemSet{
	Name: "Champion's Guard",
	Bonuses: map[int32]core.ApplyEffect{
		// +40 Attack Power.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.AttackPower:       40,
				stats.RangedAttackPower: 40,
			})
		},
		// Reduces the cooldown of your Gouge ability by 1 sec.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// +20 Stamina.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
	},
})

var ItemSetLieutenantCommandersGuard = core.NewItemSet(core.ItemSet{
	ID:   548,
	Name: "Lieutenant Commander's Guard",
	Bonuses: map[int32]core.ApplyEffect{
		// +40 Attack Power.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.AttackPower:       40,
				stats.RangedAttackPower: 40,
			})
		},
		// Reduces the cooldown of your Gouge ability by 1 sec.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// +20 Stamina.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
	},
})

var ItemSetWarlordsVestments = core.NewItemSet(core.ItemSet{
	ID:   393,
	Name: "Warlord's Vestments",
	Bonuses: map[int32]core.ApplyEffect{
		// +20 Stamina.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
		// Reduces the cooldown of your Gouge ability by 1 sec.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// +40 Attack Power.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.AttackPower:       40,
				stats.RangedAttackPower: 40,
			})
		},
	},
})

var ItemSetFieldMarshalsVestments = core.NewItemSet(core.ItemSet{
	ID:   394,
	Name: "Field Marshal's Vestments",
	Bonuses: map[int32]core.ApplyEffect{
		// +20 Stamina.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
		// Reduces the cooldown of your Gouge ability by 1 sec.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// +40 Attack Power.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.AttackPower:       40,
				stats.RangedAttackPower: 40,
			})
		},
	},
})
