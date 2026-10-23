package druid

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's druid PvP sets, by the client's set IDs (ItemSetSpell, 1.60.1.70532; wowsims/forever's
// druid/item_sets.go registers the same rows). Their pieces are new items that carry only the set
// ID. Champion's and Lieutenant Commander's Refuge give +44 healing and +15 spell damage (467550) at
// 2 pieces instead of Classic's 40 attack power; Warlord's and Field Marshal's Sanctuary keep the
// Classic layout. The 3- or 4-piece bonus is movement speed in forms, which has no place in the sim.
var (
	pvpDruidStamina       = core.SetBonusStats(stats.Stats{stats.Stamina: 20})                                  // 14467
	pvpDruidRefuge        = core.SetBonusStats(stats.Stats{stats.HealingPower: 44, stats.SpellDamage: 15})      // 467550
	pvpDruidWildhide      = core.SetBonusStats(stats.Stats{stats.SpellPower: 23})                               // 14047
	pvpDruidAttackPower   = core.SetBonusStats(stats.Stats{stats.AttackPower: 40, stats.RangedAttackPower: 40}) // 14049
	foreverRefuge         = map[int32]core.ApplyEffect{2: pvpDruidRefuge, 6: pvpDruidStamina}
	foreverHighRankRefuge = map[int32]core.ApplyEffect{2: pvpDruidStamina, 6: pvpDruidRefuge}
)

// The Forever sets with no Classic counterpart in the sim, so with bonuses under Forever only.
func init() {
	for _, set := range []struct {
		id      int32
		name    string
		bonuses map[int32]core.ApplyEffect
	}{
		{2076, "Warlord's Refuge", foreverHighRankRefuge},
		{2077, "Field Marshal's Refuge", foreverHighRankRefuge},
		{382, "Champion's Sanctuary", map[int32]core.ApplyEffect{2: pvpDruidAttackPower, 6: pvpDruidStamina}},
		{381, "Lieutenant Commander's Sanctuary", map[int32]core.ApplyEffect{2: pvpDruidAttackPower, 6: pvpDruidStamina}},
		{2080, "Champion's Wildhide", map[int32]core.ApplyEffect{2: pvpDruidWildhide, 6: pvpDruidStamina}},
		{2081, "Lieutenant Commander's Wildhide", map[int32]core.ApplyEffect{2: pvpDruidWildhide, 6: pvpDruidStamina}},
		{2079, "Warlord's Wildhide", map[int32]core.ApplyEffect{2: pvpDruidStamina, 6: pvpDruidWildhide}},
		{2078, "Field Marshal's Wildhide", map[int32]core.ApplyEffect{2: pvpDruidStamina, 6: pvpDruidWildhide}},
	} {
		core.NewItemSet(core.ItemSet{ID: set.id, Name: set.name, Bonuses: core.RulesetSetBonuses(nil, set.bonuses)})
	}
}

var ItemSetChampionsRefuge = core.NewItemSet(core.ItemSet{
	ID:   539,
	Name: "Champion's Refuge",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// +40 Attack Power.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.AttackPower, 40)
		},
		// Increases your movement speed by 15% while in Bear, Cat, or Travel Form. Only active outdoors.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// +20 Stamina.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
	}, foreverRefuge),
})

var ItemSetLieutenantCommandersRefuge = core.NewItemSet(core.ItemSet{
	ID:   551,
	Name: "Lieutenant Commander's Refuge",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// +40 Attack Power.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.AttackPower, 40)
		},
		// Increases your movement speed by 15% while in Bear, Cat, or Travel Form. Only active outdoors.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// +20 Stamina.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
	}, foreverRefuge),
})

var ItemSetFieldMarshalsSanctuary = core.NewItemSet(core.ItemSet{
	ID:   397,
	Name: "Field Marshal's Sanctuary",
	Bonuses: map[int32]core.ApplyEffect{
		// +20 Stamina.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
		// Increases your movement speed by 15% while in Bear, Cat, or Travel Form. Only active outdoors.
		3: func(agent core.Agent) {
			// Nothing to do
		},
		// +40 Attack Power.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.AttackPower, 40)
			c.AddStat(stats.RangedAttackPower, 40)
		},
	},
})

var ItemSetWarlordsSanctuary = core.NewItemSet(core.ItemSet{
	ID:   398,
	Name: "Warlord's Sanctuary",
	Bonuses: map[int32]core.ApplyEffect{
		// +20 Stamina.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
		// Increases your movement speed by 15% while in Bear, Cat, or Travel Form. Only active outdoors.
		3: func(agent core.Agent) {
			// Nothing to do
		},
		// +40 Attack Power.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.AttackPower, 40)
			c.AddStat(stats.RangedAttackPower, 40)
		},
	},
})
