package paladin

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's paladin PvP sets, by the client's set IDs (ItemSetSpell, 1.60.1.70532). Most pieces are
// new items that carry only the set ID. The Aegis sets give +44 healing and +15 spell damage
// (467550) where Classic gave 23 spell damage and healing; the Vindication sets keep the 23.
// Lieutenant Commander's Redoubt (544), whose Classic pieces Forever keeps, keeps its rows. The 3-
// or 4-piece bonus shortens Hammer of Justice's cooldown, which the sim does not cast.
var (
	pvpPaladinSpellPower = core.SetBonusStats(stats.Stats{stats.SpellPower: 23})                          // 14047
	pvpPaladinStamina    = core.SetBonusStats(stats.Stats{stats.Stamina: 20})                             // 14467
	pvpPaladinHybrid     = core.SetBonusStats(stats.Stats{stats.HealingPower: 44, stats.SpellDamage: 15}) // 467550
)

// The Forever sets with no Classic counterpart in the sim, so with bonuses under Forever only.
func init() {
	for _, set := range []struct {
		id      int32
		name    string
		bonuses map[int32]core.ApplyEffect
	}{
		{2082, "Champion's Aegis", map[int32]core.ApplyEffect{2: pvpPaladinHybrid, 6: pvpPaladinStamina}},
		{401, "Lieutenant Commander's Aegis", map[int32]core.ApplyEffect{2: pvpPaladinHybrid, 6: pvpPaladinStamina}},
		{2083, "Warlord's Aegis", map[int32]core.ApplyEffect{2: pvpPaladinStamina, 6: pvpPaladinHybrid}},
		{2084, "Champion's Vindication", map[int32]core.ApplyEffect{2: pvpPaladinSpellPower, 6: pvpPaladinStamina}},
		{2086, "Lieutenant Commander's Vindication", map[int32]core.ApplyEffect{2: pvpPaladinSpellPower, 6: pvpPaladinStamina}},
		{2085, "Warlord's Vindication", map[int32]core.ApplyEffect{2: pvpPaladinStamina, 6: pvpPaladinSpellPower}},
		{2087, "Field Marshal's Vindication", map[int32]core.ApplyEffect{2: pvpPaladinStamina, 6: pvpPaladinSpellPower}},
	} {
		core.NewItemSet(core.ItemSet{ID: set.id, Name: set.name, Bonuses: core.RulesetSetBonuses(nil, set.bonuses)})
	}
}

var ItemSetLieutenantCommandersRedoubt = core.NewItemSet(core.ItemSet{
	ID:   544,
	Name: "Lieutenant Commander's Redoubt",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases damage and healing done by magical spells and effects by up to 23.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 23)
		},
		// Reduces the cooldown of your Hammer of Justice by 10 sec.
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

var ItemSetFieldMarshalsAegis = core.NewItemSet(core.ItemSet{
	ID:   402,
	Name: "Field Marshal's Aegis",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// +20 Stamina.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
		// Reduces the cooldown of your Hammer of Justice by 10 sec.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// Increases damage and healing done by magical spells and effects by up to 23.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 23)
		},
	}, map[int32]core.ApplyEffect{2: pvpPaladinStamina, 6: pvpPaladinHybrid}),
})
