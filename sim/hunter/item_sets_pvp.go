package hunter

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's hunter PvP sets, by the client's set IDs (ItemSetSpell, 1.60.1.70532). Champion's and
// Lieutenant Commander's Pursuit are new pieces that carry only the set ID, with +20 Agility (14384)
// and +20 Stamina (14467) where Classic gave 1% parry and 15 Stamina. The other four keep their
// Classic rows. The 3- or 4-piece bonus shortens Concussive Shot's cooldown.
var foreverLowRankPursuit = map[int32]core.ApplyEffect{
	2: core.SetBonusStats(stats.Stats{stats.Agility: 20}),
	6: core.SetBonusStats(stats.Stats{stats.Stamina: 20}),
}

///////////////////////////////////////////////////////////////////////////
//                            Classic Phase 2
///////////////////////////////////////////////////////////////////////////

// https://www.wowhead.com/classic/item-set=362/lieutenant-commanders-pursuit
var ItemSetLieutenantCommandersPursuit = core.NewItemSet(core.ItemSet{
	ID:   362,
	Name: "Lieutenant Commander's Pursuit",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// Increases your chance to parry an attack by 1%.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Parry, 1*core.ParryRatingPerParryChance)
		},
		// Reduces the cooldown of your Concussive Shot by 1 sec.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// +15 Stamina.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 15)
		},
	}, foreverLowRankPursuit),
})

// https://www.wowhead.com/classic/item-set=361/champions-pursuit
var ItemSetChampionsPursuit = core.NewItemSet(core.ItemSet{
	ID:   361,
	Name: "Champion's Pursuit",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// Increases your chance to parry an attack by 1%.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Parry, 1*core.ParryRatingPerParryChance)
		},
		// Reduces the cooldown of your Concussive Shot by 1 sec.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// +15 Stamina.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 15)
		},
	}, foreverLowRankPursuit),
})

// https://www.wowhead.com/classic/item-set=543/champions-pursuance
var ItemSetChampionsPursuance = core.NewItemSet(core.ItemSet{
	ID:   543,
	Name: "Champion's Pursuance",
	Bonuses: map[int32]core.ApplyEffect{
		// +20 Agility.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Agility, 20)
		},
		// Reduces the cooldown of your Concussive Shot by 1 sec.
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

// https://www.wowhead.com/classic/item-set=550/lieutenant-commanders-pursuance
var ItemSetLieutenantCommandersPursuance = core.NewItemSet(core.ItemSet{
	ID:   550,
	Name: "Lieutenant Commander's Pursuance",
	Bonuses: map[int32]core.ApplyEffect{
		// +20 Agility.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Agility, 20)
		},
		// Reduces the cooldown of your Concussive Shot by 1 sec.
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
//                            Classic Phase 3
///////////////////////////////////////////////////////////////////////////

var ItemSetWarlordsPursuit = core.NewItemSet(core.ItemSet{
	ID:   396,
	Name: "Warlord's Pursuit",
	Bonuses: map[int32]core.ApplyEffect{
		// 20 Stamina
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
		// Reduces the cooldown of your Concussive Shot by 1 sec.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// +20 Agi
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Agility, 20)
		},
	},
})

var ItemSetFieldMarshalsPursuit = core.NewItemSet(core.ItemSet{
	ID:   395,
	Name: "Field Marshal's Pursuit",
	Bonuses: map[int32]core.ApplyEffect{
		// 20 stamina
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
		// Reduces the cooldown of your Concussive Shot by 1 sec.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// +20 Agi
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Agility, 20)
		},
	},
})
