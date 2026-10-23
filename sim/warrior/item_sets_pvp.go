package warrior

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's warrior PvP sets keep the Classic rows (client ItemSetSpell, 1.60.1.70532). Warlord's
// and Field Marshal's Battlegear are new pieces that carry only the set ID, and Champion's and
// Lieutenant Commander's Battlearmor keep their Classic pieces. Forever adds Champion's and
// Lieutenant Commander's Battlegear (281, 282) on new pieces, with bonuses under Forever only since
// the sim had no Classic version of them. The 3- or 4-piece bonus shortens Intercept's cooldown.
func init() {
	bonuses := map[int32]core.ApplyEffect{
		2: core.SetBonusStats(stats.Stats{stats.AttackPower: 40, stats.RangedAttackPower: 40}), // 14049
		6: core.SetBonusStats(stats.Stats{stats.Stamina: 20}),                                  // 14467
	}
	core.NewItemSet(core.ItemSet{ID: 281, Name: "Champion's Battlegear", Bonuses: core.RulesetSetBonuses(nil, bonuses)})
	core.NewItemSet(core.ItemSet{ID: 282, Name: "Lieutenant Commander's Battlegear", Bonuses: core.RulesetSetBonuses(nil, bonuses)})
}

///////////////////////////////////////////////////////////////////////////
//                           Classic Phase 2
///////////////////////////////////////////////////////////////////////////

var ItemSetChampionsBattlearmor = core.NewItemSet(core.ItemSet{
	ID:   537,
	Name: "Champion's Battlearmor",
	Bonuses: map[int32]core.ApplyEffect{
		// +40 Attack Power.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.AttackPower:       40,
				stats.RangedAttackPower: 40,
			})
		},
		// Reduces the cooldown of your Intercept ability by 5 sec.
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

var ItemSetLieutenantCommandersBattlearmor = core.NewItemSet(core.ItemSet{
	ID:   545,
	Name: "Lieutenant Commander's Battlearmor",
	Bonuses: map[int32]core.ApplyEffect{
		// +40 Attack Power.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.AttackPower:       40,
				stats.RangedAttackPower: 40,
			})
		},
		// Reduces the cooldown of your Intercept ability by 5 sec.
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

var ItemSetWarlordsBattlegear = core.NewItemSet(core.ItemSet{
	ID:   383,
	Name: "Warlord's Battlegear",
	Bonuses: map[int32]core.ApplyEffect{
		// +20 Stamina.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
		// Reduces the cooldown of your Intercept ability by 5 sec.
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

var ItemSetFieldMarshalsBattlegear = core.NewItemSet(core.ItemSet{
	ID:   384,
	Name: "Field Marshal's Battlegear",
	Bonuses: map[int32]core.ApplyEffect{
		// +20 Stamina.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
		// Reduces the cooldown of your Intercept ability by 5 sec.
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
