package priest

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's priest PvP sets, by the client's set IDs (ItemSetSpell, 1.60.1.70532; wowsims/forever's
// priest/item_sets.go registers the same rows). Most pieces are new items that carry only the set
// ID. Lieutenant Commander's Raiment gives 20 Stamina at 6 pieces, not 15, and Forever's own
// Investiture sets (2072 to 2075) give +44 healing and +15 spell damage (467550) in place of 23
// spell damage and healing. Champion's Raiment (342) and Champion's Investiture (540), whose Classic
// pieces Forever keeps, keep their Classic rows. The 3- or 4-piece bonus lengthens Psychic Scream,
// which the sim does not cast.
var (
	pvpPriestSpellPower = core.SetBonusStats(stats.Stats{stats.SpellPower: 23})                          // 14047
	pvpPriestStamina    = core.SetBonusStats(stats.Stats{stats.Stamina: 20})                             // 14467
	pvpPriestHybrid     = core.SetBonusStats(stats.Stats{stats.HealingPower: 44, stats.SpellDamage: 15}) // 467550
)

// The Forever sets with no Classic counterpart in the sim, so with bonuses under Forever only. 549 and
// 2072 share their names with the Classic sets 342 and 540, whose pieces Forever keeps.
func init() {
	for _, set := range []struct {
		id      int32
		name    string
		bonuses map[int32]core.ApplyEffect
	}{
		{549, "Champion's Raiment", map[int32]core.ApplyEffect{2: pvpPriestSpellPower, 6: pvpPriestStamina}},
		{2072, "Champion's Investiture", map[int32]core.ApplyEffect{2: pvpPriestHybrid, 6: pvpPriestStamina}},
		{2073, "Warlord's Investiture", map[int32]core.ApplyEffect{2: pvpPriestStamina, 6: pvpPriestHybrid}},
		{2074, "Field Marshal's Investiture", map[int32]core.ApplyEffect{2: pvpPriestStamina, 6: pvpPriestHybrid}},
	} {
		core.NewItemSet(core.ItemSet{ID: set.id, Name: set.name, Bonuses: core.RulesetSetBonuses(nil, set.bonuses)})
	}
}

var ItemSetChampionsRaiment = core.NewItemSet(core.ItemSet{
	ID:   342,
	Name: "Champion's Raiment",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases damage and healing done by magical spells and effects by up to 23.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 23)
		},
		// Increases the duration of your Psychic Scream spell by 1 sec.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// +15 Stamina.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 15)
		},
	},
})

var ItemSetLieutenantCommandersRaiment = core.NewItemSet(core.ItemSet{
	ID:   344,
	Name: "Lieutenant Commander's Raiment",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// Increases damage and healing done by magical spells and effects by up to 23.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 23)
		},
		// Increases the duration of your Psychic Scream spell by 1 sec.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// +15 Stamina.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 15)
		},
	}, map[int32]core.ApplyEffect{2: pvpPriestSpellPower, 6: pvpPriestStamina}),
})

var ItemSetChampionsInvestiture = core.NewItemSet(core.ItemSet{
	ID:   540,
	Name: "Champion's Investiture",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases damage and healing done by magical spells and effects by up to 23.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 23)
		},
		// Increases the duration of your Psychic Scream spell by 1 sec.
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

var ItemSetLieutenantCommandersInvestiture = core.NewItemSet(core.ItemSet{
	ID:   2075,
	Name: "Lieutenant Commander's Investiture",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// Increases damage and healing done by magical spells and effects by up to 23.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 23)
		},
		// Increases the duration of your Psychic Scream spell by 1 sec.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// +20 Stamina.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
	}, map[int32]core.ApplyEffect{2: pvpPriestHybrid, 6: pvpPriestStamina}),
})

var ItemSetWarlordsRaiment = core.NewItemSet(core.ItemSet{
	ID:   390,
	Name: "Warlord's Raiment",
	Bonuses: map[int32]core.ApplyEffect{
		// +20 Stamina.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
		// Increases the duration of your Psychic Scream spell by 1 sec.
		3: func(agent core.Agent) {
			// Nothing to do
		},
		// Increases damage and healing done by magical spells and effects by up to 23.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 23)
		},
	},
})

var ItemSetFieldMarshalsRaiment = core.NewItemSet(core.ItemSet{
	ID:   389,
	Name: "Field Marshal's Raiment",
	Bonuses: map[int32]core.ApplyEffect{
		// +20 Stamina.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
		// Increases the duration of your Psychic Scream spell by 1 sec.
		3: func(agent core.Agent) {
			// Nothing to do
		},
		// Increases damage and healing done by magical spells and effects by up to 23.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 23)
		},
	},
})
