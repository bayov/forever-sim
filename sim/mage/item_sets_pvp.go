package mage

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's mage PvP sets, by the client's set IDs (ItemSetSpell, 1.60.1.70532). Champion's Arcanum
// (542) keeps its Classic pieces and rows, and Warlord's and Field Marshal's Regalia keep the Classic
// rows on new pieces that carry only the set ID. The 3- or 4-piece bonus shortens Blink's cooldown.
var (
	pvpMageSpellPower = core.SetBonusStats(stats.Stats{stats.SpellPower: 23}) // 14047
	pvpMageStamina    = core.SetBonusStats(stats.Stats{stats.Stamina: 20})    // 14467
)

// The Forever sets with no Classic counterpart in the sim, so with bonuses under Forever only.
// Champion's Regalia is two sets: 341 on the Classic pieces (15 Stamina, 14462) and 546 on new ones.
func init() {
	for _, set := range []struct {
		id      int32
		name    string
		bonuses map[int32]core.ApplyEffect
	}{
		{341, "Champion's Regalia", map[int32]core.ApplyEffect{2: pvpMageSpellPower, 6: core.SetBonusStats(stats.Stats{stats.Stamina: 15})}},
		{546, "Champion's Regalia", map[int32]core.ApplyEffect{2: pvpMageSpellPower, 6: pvpMageStamina}},
		{343, "Lieutenant Commander's Regalia", map[int32]core.ApplyEffect{2: pvpMageSpellPower, 6: pvpMageStamina}},
	} {
		core.NewItemSet(core.ItemSet{ID: set.id, Name: set.name, Bonuses: core.RulesetSetBonuses(nil, set.bonuses)})
	}
}

///////////////////////////////////////////////////////////////////////////
//                            Clasic Phase 2
///////////////////////////////////////////////////////////////////////////

var ItemSetChampionsArcanum = core.NewItemSet(core.ItemSet{
	ID:   542,
	Name: "Champion's Arcanum",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases damage and healing done by magical spells and effects by up to 23.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 23)
		},
		// Reduces the cooldown of your Blink spell by 1.5 sec.
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

var ItemSetLieutenantCommandersArcanum = core.NewItemSet(core.ItemSet{
	Name: "Lieutenant Commander's Arcanum",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases damage and healing done by magical spells and effects by up to 23.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 23)
		},
		// Reduces the cooldown of your Blink spell by 1.5 sec.
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

///////////////////////////////////////////////////////////////////////////
//                            Phase 3 Items
///////////////////////////////////////////////////////////////////////////

var ItemSetWarlordsRegalia = core.NewItemSet(core.ItemSet{
	ID:   387,
	Name: "Warlord's Regalia",
	Bonuses: map[int32]core.ApplyEffect{
		// +20 Stamina.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
		// Reduces the cooldown of your Blink spell by 1.5 sec.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// Increases damage and healing done by magical spells and effects by up to 23.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 23)
		},
	},
})

var ItemSetFieldMarshalsRegalia = core.NewItemSet(core.ItemSet{
	ID:   388,
	Name: "Field Marshal's Regalia",
	Bonuses: map[int32]core.ApplyEffect{
		// +20 Stamina.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
		// Reduces the cooldown of your Blink spell by 1.5 sec.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// Increases damage and healing done by magical spells and effects by up to 23.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 23)
		},
	},
})
