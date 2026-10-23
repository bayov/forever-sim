package shaman

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's shaman PvP sets, by the client's set IDs (ItemSetSpell, 1.60.1.70532). Most pieces are
// new items that carry only the set ID. Champion's Earthshaker gives 20 Stamina at 6 pieces, not
// 15, and Warlord's Earthshaker its Shock crit at 3 pieces, not 4. Champion's Stormcaller (538)
// keeps its Classic pieces and rows. The Wartide sets give +44 healing and +15 spell damage (467550).
var (
	pvpShamanAttackPower = core.SetBonusStats(stats.Stats{stats.AttackPower: 40, stats.RangedAttackPower: 40}) // 14049
	pvpShamanSpellPower  = core.SetBonusStats(stats.Stats{stats.SpellPower: 23})                               // 14047
	pvpShamanStamina     = core.SetBonusStats(stats.Stats{stats.Stamina: 20})                                  // 14467
	pvpShamanHybrid      = core.SetBonusStats(stats.Stats{stats.HealingPower: 44, stats.SpellDamage: 15})      // 467550
)

// 22804: Improves your chance to get a critical strike with all Shock spells by 2%.
func pvpShamanShockCrit(agent core.Agent) {
	shaman := agent.(ShamanAgent).GetShaman()
	shaman.GetOrRegisterAura(core.Aura{
		Label:    "Shaman Shock Crit Bonus",
		ActionID: core.ActionID{SpellID: 22804},
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range core.Flatten([][]*core.Spell{shaman.EarthShock, shaman.FlameShock, shaman.FrostShock}) {
				if spell != nil {
					spell.BonusCritRating += 2 * core.CritRatingPerCritChance
				}
			}
		},
	})
}

// The Forever sets with no Classic counterpart in the sim, so with bonuses under Forever only.
func init() {
	lowRank := func(stat core.ApplyEffect) map[int32]core.ApplyEffect {
		return map[int32]core.ApplyEffect{2: stat, 4: pvpShamanShockCrit, 6: pvpShamanStamina}
	}
	highRank := func(stat core.ApplyEffect) map[int32]core.ApplyEffect {
		return map[int32]core.ApplyEffect{2: pvpShamanStamina, 3: pvpShamanShockCrit, 6: stat}
	}
	for _, set := range []struct {
		id      int32
		name    string
		bonuses map[int32]core.ApplyEffect
	}{
		{2088, "Lieutenant Commander's Earthshaker", lowRank(pvpShamanAttackPower)},
		{2089, "Field Marshal's Earthshaker", highRank(pvpShamanAttackPower)},
		{2094, "Champion's Thunderfist", lowRank(pvpShamanSpellPower)},
		{2096, "Lieutenant Commander's Thunderfist", lowRank(pvpShamanSpellPower)},
		{2095, "Warlord's Thunderfist", highRank(pvpShamanSpellPower)},
		{2097, "Field Marshal's Thunderfist", highRank(pvpShamanSpellPower)},
		{2090, "Champion's Wartide", lowRank(pvpShamanHybrid)},
		{2092, "Lieutenant Commander's Wartide", lowRank(pvpShamanHybrid)},
		{2091, "Warlord's Wartide", highRank(pvpShamanHybrid)},
		{2093, "Field Marshal's Wartide", highRank(pvpShamanHybrid)},
	} {
		core.NewItemSet(core.ItemSet{ID: set.id, Name: set.name, Bonuses: core.RulesetSetBonuses(nil, set.bonuses)})
	}
}

var ItemSetChampionsEarthshaker = core.NewItemSet(core.ItemSet{
	ID:   301,
	Name: "Champion's Earthshaker",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// +40 Attack Power.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.AttackPower:       40,
				stats.RangedAttackPower: 40,
			})
		},
		// Improves your chance to get a critical strike with all Shock spells by 2%.
		4: func(agent core.Agent) {
			shaman := agent.(ShamanAgent).GetShaman()
			shaman.GetOrRegisterAura(core.Aura{
				Label:    "Shaman Shock Crit Bonus",
				ActionID: core.ActionID{SpellID: 22804},
				OnInit: func(aura *core.Aura, sim *core.Simulation) {
					for _, spell := range core.Flatten([][]*core.Spell{shaman.EarthShock, shaman.FlameShock, shaman.FrostShock}) {
						if spell != nil {
							spell.BonusCritRating += 2 * core.CritRatingPerCritChance
						}
					}
				},
			})
		},
		// +15 Stamina.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 15)
		},
	}, map[int32]core.ApplyEffect{2: pvpShamanAttackPower, 4: pvpShamanShockCrit, 6: pvpShamanStamina}),
})

var ItemSetChampionsStormcaller = core.NewItemSet(core.ItemSet{
	ID:   538,
	Name: "Champion's Stormcaller",
	Bonuses: map[int32]core.ApplyEffect{
		// +40 Attack Power.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.AttackPower:       40,
				stats.RangedAttackPower: 40,
			})
		},
		// Improves your chance to get a critical strike with all Shock spells by 2%.
		4: func(agent core.Agent) {
			shaman := agent.(ShamanAgent).GetShaman()
			shaman.GetOrRegisterAura(core.Aura{
				Label:    "Shaman Shock Crit Bonus",
				ActionID: core.ActionID{SpellID: 22804},
				OnInit: func(aura *core.Aura, sim *core.Simulation) {
					for _, spell := range core.Flatten([][]*core.Spell{shaman.EarthShock, shaman.FlameShock, shaman.FrostShock}) {
						if spell != nil {
							spell.BonusCritRating += 2 * core.CritRatingPerCritChance
						}
					}
				},
			})
		},
		// +20 Stamina.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
	},
})

var ItemSetWarlordsEarthshaker = core.NewItemSet(core.ItemSet{
	ID:   386,
	Name: "Warlord's Earthshaker",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// +20 Stamina.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
		// Improves your chance to get a critical strike with all Shock spells by 2%.
		4: func(agent core.Agent) {
			shaman := agent.(ShamanAgent).GetShaman()
			shaman.GetOrRegisterAura(core.Aura{
				Label:    "Shaman Shock Crit Bonus",
				ActionID: core.ActionID{SpellID: 22804},
				OnInit: func(aura *core.Aura, sim *core.Simulation) {
					for _, spell := range core.Flatten([][]*core.Spell{shaman.EarthShock, shaman.FlameShock, shaman.FrostShock}) {
						if spell != nil {
							spell.BonusCritRating += 2 * core.CritRatingPerCritChance
						}
					}
				},
			})
		},
		// +40 Attack Power.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.AttackPower:       40,
				stats.RangedAttackPower: 40,
			})
		},
	}, map[int32]core.ApplyEffect{2: pvpShamanStamina, 3: pvpShamanShockCrit, 6: pvpShamanAttackPower}),
})
