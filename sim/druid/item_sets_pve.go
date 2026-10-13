package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

var ItemSetFeralheartRaiment = core.NewItemSet(core.ItemSet{
	Name: "Feralheart Raiment",
	Bonuses: map[int32]core.ApplyEffect{
		// (2) Set : +8 All Resistances.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddResistances(8)
		},
		// (4) Set : When struck in combat has a chance of returning 300 mana, 10 rage, or 40 energy to the wearer. (Proc chance: 2%)
		4: func(agent core.Agent) {
			c := agent.GetCharacter()
			actionID := core.ActionID{SpellID: 27781}
			manaMetrics := c.NewManaMetrics(actionID)
			energyMetrics := c.NewEnergyMetrics(actionID)
			rageMetrics := c.NewRageMetrics(actionID)

			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				Name:       "Nature's Bounty (Mana)",
				Callback:   core.CallbackOnCastComplete,
				ProcMask:   core.ProcMaskSpellDamage | core.ProcMaskSpellHealing,
				ProcChance: 0.02,
				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					c.AddMana(sim, 300, manaMetrics)
				},
			})
			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				Name:       "Nature's Bounty (Energy)",
				Callback:   core.CallbackOnSpellHitDealt,
				Outcome:    core.OutcomeLanded,
				ProcMask:   core.ProcMaskMeleeWhiteHit,
				ProcChance: 0.02,
				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					if c.HasEnergyBar() {
						c.AddEnergy(sim, 40, energyMetrics)
					}
				},
			})
			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				Name:       "Nature's Bounty (Rage)",
				Callback:   core.CallbackOnSpellHitTaken,
				ProcMask:   core.ProcMaskMelee,
				ProcChance: 0.02,
				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					if c.HasRageBar() {
						c.AddRage(sim, 10, rageMetrics)
					}
				},
			})
		},
		// (6) Set : Increases damage and healing done by magical spells and effects by up to 15.
		// (6) Set : +26 Attack Power.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 15)
			c.AddStat(stats.AttackPower, 26)
		},
		// (8) Set : +200 Armor.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Armor, 200)
		},
	},
})

var ItemSetCenarionRaiment = core.NewItemSet(core.ItemSet{
	Name: "Cenarion Raiment",
	Bonuses: map[int32]core.ApplyEffect{
		// (3) Set : Damage dealt by Thorns increased by 4 and duration increased by 50%.
		3: func(agent core.Agent) {
			// Nothing to do
		},
		// (5) Set : Improves your chance to get a critical strike with spells by 2%.
		5: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellCrit, 2*core.SpellCritRatingPerCritChance)
		},
		// (8) Set : Reduces the cooldown of your Tranquility and Hurricane spells by 50%.
		8: func(agent core.Agent) {
			// Nothing to do
		},
	},
})

var ItemSetWildheartRaiment = core.NewItemSet(core.ItemSet{
	Name: "Wildheart Raiment",
	Bonuses: map[int32]core.ApplyEffect{
		// (2) Set : +200 Armor.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Armor, 200)
		},
		// (4) Set : +26 Attack Power.
		// (4) Set : Increases damage and healing done by magical spells and effects by up to 15.
		4: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.AttackPower, 26)
			c.AddStat(stats.SpellPower, 15)
		},
		// (6) Set : When struck in combat has a chance of returning 300 mana, 10 rage, or 40 energy to the wearer. (Proc chance: 2%)
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			actionID := core.ActionID{SpellID: 27781}
			manaMetrics := c.NewManaMetrics(actionID)
			energyMetrics := c.NewEnergyMetrics(actionID)
			rageMetrics := c.NewRageMetrics(actionID)

			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				Name:       "Nature's Bounty (Mana)",
				Callback:   core.CallbackOnCastComplete,
				ProcMask:   core.ProcMaskSpellDamage | core.ProcMaskSpellHealing,
				ProcChance: 0.02,
				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					c.AddMana(sim, 300, manaMetrics)
				},
			})
			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				Name:       "Nature's Bounty (Energy)",
				Callback:   core.CallbackOnSpellHitDealt,
				Outcome:    core.OutcomeLanded,
				ProcMask:   core.ProcMaskMeleeWhiteHit,
				ProcChance: 0.02,
				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					if c.HasEnergyBar() {
						c.AddEnergy(sim, 40, energyMetrics)
					}
				},
			})
			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				Name:       "Nature's Bounty (Rage)",
				Callback:   core.CallbackOnSpellHitTaken,
				ProcMask:   core.ProcMaskMelee,
				ProcChance: 0.02,
				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					if c.HasRageBar() {
						c.AddRage(sim, 10, rageMetrics)
					}
				},
			})
		},
		// (8) Set : +8 All Resistances.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddResistances(8)
		},
	},
})

var ItemSetStormrageRaiment = core.NewItemSet(core.ItemSet{
	Name: "Stormrage Raiment",
	Bonuses: map[int32]core.ApplyEffect{
		// (3) Set : Allows 15% of your Mana regeneration to continue while casting.
		3: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.PseudoStats.SpiritRegenRateCasting += .15
		},
		// (5) Set : Reduces the casting time of your Regrowth spell by 0.2 sec.
		5: func(agent core.Agent) {
			// Nothing to do.
		},
		// (8) Set : Increases the duration of your Rejuvenation spell by 3 sec.
		8: func(agent core.Agent) {
			// Nothing to do.
		},
	},
})

var ItemSetSymbolsOfUnendingLife = core.NewItemSet(core.ItemSet{
	Name: "Symbols of Unending Life",
	Bonuses: map[int32]core.ApplyEffect{
		// (3) Set : Your finishing moves now refund 30 energy on a Miss, Dodge, Block, or Parry.
		3: func(agent core.Agent) {
			c := agent.GetCharacter()
			actionID := core.ActionID{SpellID: 26107}
			energyMetrics := c.NewEnergyMetrics(actionID)
			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				Name:     "Symbols of Unending Life Finisher Bonus",
				Callback: core.CallbackOnSpellHitDealt,
				Outcome:  core.OutcomeMiss | core.OutcomeDodge | core.OutcomeBlock | core.OutcomeParry,
				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					if spell.SpellCode == SpellCode_DruidFerociousBite || spell.SpellCode == SpellCode_DruidRip {
						c.AddEnergy(sim, 30, energyMetrics)
					}
				},
			})
		},
	},
})

// mobTypeBonus adds attack power or spell damage against every target of one creature type,
// for the Grovekeeper 4 piece bonuses.
func mobTypeBonus(c *core.Character, label string, mobType proto.MobType, attackPower, spellDamage float64) {
	targets := core.FilterSlice(c.Env.Encounter.TargetUnits, func(unit *core.Unit) bool {
		return unit.MobType == mobType
	})
	core.MakePermanent(c.RegisterAura(core.Aura{
		Label: label,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, target := range targets {
				for _, at := range c.AttackTables[target.UnitIndex] {
					at.BonusAttackPowerTaken += attackPower
					at.BonusSpellDamageTaken += spellDamage
				}
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, target := range targets {
				for _, at := range c.AttackTables[target.UnitIndex] {
					at.BonusAttackPowerTaken -= attackPower
					at.BonusSpellDamageTaken -= spellDamage
				}
			}
		},
	}))
}

// grovekeeperEclipseBonuses are the set bonuses of Grovekeeper Eclipse, Forever's crafted level 60
// Balance set (item set 2114, leather head, shoulders, chest, hands, legs and feet 280998 to 281003,
// in the 1.60.1.70338 client).
//
// The client pieces carry no stats yet, so no real set is registered for them. We give the same bonuses to
// the synthetic Phase 1 Balance set below, the way the enhancement shaman's carries The Spiritcaller's Rage.
var grovekeeperEclipseBonuses = map[int32]core.ApplyEffect{
	// Improves your chance to hit by 1%.
	2: func(agent core.Agent) {
		c := agent.GetCharacter()
		c.AddStats(stats.Stats{
			stats.MeleeHit: core.MeleeHitRatingPerHitChance,
			stats.SpellHit: core.SpellHitRatingPerHitChance,
		})
	},
	// Reduces the cooldown of your Nature's Grasp spell by 15 sec.
	3: func(agent core.Agent) {
		// Nothing to do
	},
	// Increases damage done by your spells and effects by up to 21 when fighting Demons.
	4: func(agent core.Agent) {
		mobTypeBonus(agent.GetCharacter(), "Spell Damage against Demons", proto.MobType_MobTypeDemon, 0, 21)
	},
	// Increases the duration of your Insect Swarm spell by 3 sec.
	5: func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		druid.InsectSwarmBonusDuration += time.Second * 3
	},
}

// grovekeeperFerocityBonuses are the set bonuses of Grovekeeper Ferocity, Forever's crafted level 60
// Feral set (item set 2115, leather head, shoulders, chest, hands, legs and feet 280992 to 280997, in the
// 1.60.1.70338 client). The synthetic Phase 1 Feral set below carries them.
var grovekeeperFerocityBonuses = map[int32]core.ApplyEffect{
	// Increases your attack speed and casting speed by 1%.
	2: func(agent core.Agent) {
		c := agent.GetCharacter()
		c.PseudoStats.MeleeSpeedMultiplier *= 1.01
		c.PseudoStats.RangedSpeedMultiplier *= 1.01
		c.PseudoStats.CastSpeedMultiplier *= 1.01
	},
	// Hibernate can be cast in Cat Form, faster, with a 15 sec cooldown.
	3: func(agent core.Agent) {
		// Nothing to do
	},
	// +36 Attack Power against Demons.
	4: func(agent core.Agent) {
		mobTypeBonus(agent.GetCharacter(), "Attack Power against Demons", proto.MobType_MobTypeDemon, 36, 0)
	},
	// Reduces the cooldown of your Shifting Power ability by 1 sec.
	5: func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		druid.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellCode == SpellCode_DruidShiftingPower {
				spell.CD.Duration -= time.Second
			}
		})
	},
}

// The sets of the synthetic Phase 1 druid items (tools/database/forever_synthetic_items_druid.go).
var ItemSetBalanceSyntheticPhase1 = core.NewItemSet(core.ItemSet{
	Name:    "Balance Synthetic (Phase 1)",
	Bonuses: grovekeeperEclipseBonuses,
})

var ItemSetFeralSyntheticPhase1 = core.NewItemSet(core.ItemSet{
	Name:    "Feral Synthetic (Phase 1)",
	Bonuses: grovekeeperFerocityBonuses,
})
