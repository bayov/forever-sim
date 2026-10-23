package item_sets

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The Dungeon Set 1 sets. Each keeps its Classic layout here, and under the Forever ruleset takes
// the layout in forever_dungeon_sets.go.

var ItemSetWildheartRaiment = core.NewItemSet(core.ItemSet{
	ID:   185,
	Name: "Wildheart Raiment",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// +200 Armor.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Armor, 200)
		},
		// +26 Attack Power,increases damage and healing done by magical spells and effects by up to 15.
		4: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.AttackPower:       26,
				stats.RangedAttackPower: 26,
				stats.SpellDamage:       15,
				stats.HealingPower:      15,
			})
		},
		// When struck in combat has a chance of returning 300 mana, 10 rage, or 40 energy to the wearer.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			actionID := core.ActionID{SpellID: 27781}
			manaMetrics := c.NewManaMetrics(actionID)
			energyMetrics := c.NewEnergyMetrics(actionID)
			rageMetrics := c.NewRageMetrics(actionID)

			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				ActionID:   actionID,
				Name:       "Nature's Bounty",
				Callback:   core.CallbackOnSpellHitTaken,
				ProcMask:   core.ProcMaskMelee,
				ProcChance: 0.02,
				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					if c.HasManaBar() {
						c.AddMana(sim, 300, manaMetrics)
					}
					if c.HasEnergyBar() {
						c.AddEnergy(sim, 40, energyMetrics)
					}
					if c.HasRageBar() {
						c.AddRage(sim, 10, rageMetrics)
					}
				},
			})
		},
		// +8 All Resistances.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddResistances(8)
		},
	}, foreverWildheartRaiment),
})

var ItemSetBeaststalkerArmor = core.NewItemSet(core.ItemSet{
	ID:   186,
	Name: "Beaststalker Armor",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// +200 Armor.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Armor, 200)
		},
		// +40 Attack Power.
		4: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.AttackPower:       40,
				stats.RangedAttackPower: 40,
			})
		},
		// Your normal ranged attacks have a 4% chance of restoring 200 mana.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			actionID := core.ActionID{SpellID: 27785}
			manaMetrics := c.NewManaMetrics(actionID)

			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				ActionID:   actionID,
				Name:       "Hunter Armor Energize",
				Callback:   core.CallbackOnSpellHitDealt,
				Outcome:    core.OutcomeLanded,
				ProcMask:   core.ProcMaskWhiteHit,
				ProcChance: 0.04,
				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					if c.HasManaBar() {
						c.AddMana(sim, 200, manaMetrics)
					}
				},
			})
		},
		// +8 All Resistances.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddResistances(8)
		},
	}, foreverBeaststalkerArmor),
})

var ItemSetMagistersRegalia = core.NewItemSet(core.ItemSet{
	ID:   181,
	Name: "Magister's Regalia",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// +200 Armor.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Armor, 200)
		},
		// Increases damage and healing done by magical spells and effects by up to 23.
		4: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 23)
		},
		// When struck in combat has a chance of freezing the attacker in place for 3 sec.
		6: func(agent core.Agent) {
			// No implementation in sim
		},
		// +8 All Resistances.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddResistances(8)
		},
	}, foreverMagistersRegalia),
})

var ItemSetLightforgeArmor = core.NewItemSet(core.ItemSet{
	ID:   188,
	Name: "Lightforge Armor",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// +200 Armor.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Armor, 200)
		},
		// +40 Attack Power
		4: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.AttackPower:       40,
				stats.RangedAttackPower: 40,
			})
		},
		// Chance on melee attack to increase your damage and healing done by magical spells and effects by up to 95 for 10 sec.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			actionID := core.ActionID{SpellID: 27498}

			procAura := c.NewTemporaryStatsAura("Crusader's Wrath", actionID, stats.Stats{stats.SpellPower: 95}, time.Second*10)
			handler := func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				procAura.Activate(sim)
			}

			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				ActionID: actionID,
				Name:     "Item - Crusader's Wrath Proc - Lightforge Armor",
				Callback: core.CallbackOnSpellHitDealt,
				Outcome:  core.OutcomeLanded,
				ProcMask: core.ProcMaskMeleeWhiteHit,
				PPM:      3, //6% is the sod proc rate, best data I could find online says 3 ppm
				Handler:  handler,
			})
		},
		// +8 All Resistances.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddResistances(8)
		},
	}, foreverLightforgeArmor),
})

var ItemSetVestmentsOfTheDevout = core.NewItemSet(core.ItemSet{
	ID:   182,
	Name: "Vestments of the Devout",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// +200 Armor.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Armor, 200)
		},
		// Increases damage and healing done by magical spells and effects by up to 23.
		4: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 23)
		},
		// 6 pieces: When struck in combat has a chance of shielding the wearer in a protective shield which will absorb 350 damage.
		6: func(agent core.Agent) {
			//No use case in Classic Sim
		},
		// +8 All Resistances.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddResistances(8)
		},
	}, foreverVestmentsOfTheDevout),
})

var ItemSetShadowcraftArmor = core.NewItemSet(core.ItemSet{
	ID:   184,
	Name: "Shadowcraft Armor",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// +200 Armor.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Armor, 200)
		},
		// +40 Attack Power.
		4: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.AttackPower:       40,
				stats.RangedAttackPower: 40,
			})
		},
		// Chance on melee attack to restore 35 energy.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			actionID := core.ActionID{SpellID: 27787}
			energyMetrics := c.NewEnergyMetrics(actionID)

			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				ActionID: actionID,
				Name:     "Rogue Armor Energize",
				Callback: core.CallbackOnSpellHitDealt,
				Outcome:  core.OutcomeLanded,
				ProcMask: core.ProcMaskMeleeWhiteHit,
				PPM:      1,
				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					if c.HasEnergyBar() {
						c.AddEnergy(sim, 35, energyMetrics)
					}
				},
			})
		},
		// +8 All Resistances.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddResistances(8)
		},
	}, foreverShadowcraftArmor),
})

var ItemSetTheElements = core.NewItemSet(core.ItemSet{
	ID:   187,
	Name: "The Elements",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// +200 Armor.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Armor, 200)
		},
		// Increases damage and healing done by magical spells and effects by up to 23.
		4: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.SpellDamage:  23,
				stats.HealingPower: 23,
			})
		},
		// Chance on spell cast to increase your damage and healing by up to 95 for 10 sec.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			actionID := core.ActionID{SpellID: 27774}

			procAura := c.NewTemporaryStatsAura("The Furious Storm", core.ActionID{SpellID: 27774}, stats.Stats{stats.SpellPower: 95}, time.Second*10)
			handler := func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				procAura.Activate(sim)
			}

			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				ActionID:   actionID,
				Name:       "Item - The Furious Storm Proc",
				Callback:   core.CallbackOnCastComplete,
				ProcMask:   core.ProcMaskSpellDamage | core.ProcMaskSpellHealing,
				ProcChance: 0.04,
				Handler:    handler,
			})
		},
		// +8 All Resistances.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddResistances(8)
		},
	}, foreverTheElements),
})

var ItemSetDreadmistRaiment = core.NewItemSet(core.ItemSet{
	ID:   183,
	Name: "Dreadmist Raiment",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// +200 Armor.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Armor, 200)
		},
		// Increases damage and healing done by magical spells and effects by up to 23.
		4: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 23)
		},
		// When struck in combat has a chance of causing the attacker to flee in terror for 2 seconds.
		6: func(agent core.Agent) {
			//No use case in Classic Sim
		},
		// +8 All Resistances.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddResistances(8)
		},
	}, foreverDreadmistRaiment),
})

var ItemSetBattlegearOfValor = core.NewItemSet(core.ItemSet{
	ID:   189,
	Name: "Battlegear of Valor",
	Bonuses: core.RulesetSetBonuses(map[int32]core.ApplyEffect{
		// +200 Armor.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Armor, 200)
		},
		// +40 Attack Power.
		4: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.AttackPower:       40,
				stats.RangedAttackPower: 40,
			})
		},
		// Chance on melee attack to heal you for 88 to 132
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			actionID := core.ActionID{SpellID: 27419}
			healthMetrics := c.NewHealthMetrics(core.ActionID{SpellID: 27419})

			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				ActionID: actionID,
				Name:     "Warrior's Resolve",
				Callback: core.CallbackOnSpellHitDealt,
				Outcome:  core.OutcomeLanded,
				ProcMask: core.ProcMaskMelee,
				PPM:      1,
				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					c.GainHealth(sim, sim.Roll(88, 133), healthMetrics)
				},
			})
		},
		// +8 All Resistances.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddResistances(8)
		},
	}, foreverBattlegearOfValor),
})
