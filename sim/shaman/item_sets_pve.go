package shaman

import (
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

var ItemSetTheFiveThunders = core.NewItemSet(core.ItemSet{
	Name: "The Five Thunders",
	Bonuses: map[int32]core.ApplyEffect{
		// +8 All Resistances.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddResistances(8)
		},
		// Chance on spell cast to increase your damage and healing by up to 95 for 10 sec.
		// (Proc chance: 4%)
		4: func(agent core.Agent) {
			c := agent.GetCharacter()

			procAura := c.NewTemporaryStatsAura("The Furious Storm", core.ActionID{SpellID: 27775}, stats.Stats{stats.SpellPower: 95}, time.Second*10)
			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				Name:       "Item - The Furious Storm Proc (Spell Cast)",
				Callback:   core.CallbackOnCastComplete,
				ProcMask:   core.ProcMaskSpellDamage | core.ProcMaskSpellHealing,
				ProcChance: 0.04,
				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					procAura.Activate(sim)
				},
			})
		},
		// Increases damage and healing done by magical spells and effects by up to 23.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddResistances(8)
		},
		// +200 Armor.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Armor, 200)
		},
	},
})

var ItemSetTheEarthfury = core.NewItemSet(core.ItemSet{
	Name: "The Earthfury",
	Bonuses: map[int32]core.ApplyEffect{
		// The radius of your totems that affect friendly targets is increased to 30 yd.
		3: func(agent core.Agent) {
			// Nothing to do
		},
		// After casting your Healing Wave or Lesser Healing Wave spell, gives you a 25% chance to gain Mana equal to 35% of the base cost of the spell.
		5: func(agent core.Agent) {
			// Nothing to do
		},
		// Your Healing Wave will now jump to additional nearby targets. Each jump reduces the effectiveness of the heal by 80%, and the spell will jump to up to two additional targets.
		8: func(agent core.Agent) {
			// Nothing to do
		},
	},
})

var ItemSetTheTenStorms = core.NewItemSet(core.ItemSet{
	Name: "The Ten Storms",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases the amount healed by Chain Heal to targets beyond the first by 30%.
		3: func(agent core.Agent) {
			// Nothing to do
		},
		// Improves your chance to get a critical strike with Nature spells by 3%.
		5: func(agent core.Agent) {
			shaman := agent.(ShamanAgent).GetShaman()
			shaman.PseudoStats.SchoolBonusCritChance[stats.SchoolIndexNature] += 3 * core.SpellCritRatingPerCritChance
		},
		// When you cast a Healing Wave or Lesser Healing Wave, there is a 25% chance the target also receives a free Lightning Shield that causes 50 Nature damage to attacker on hit.
		8: func(agent core.Agent) {
			// Nothing to do
		},
	},
})

var ItemSetStormcallersGarb = core.NewItemSet(core.ItemSet{
	Name: "Stormcaller's Garb",
	Bonuses: map[int32]core.ApplyEffect{
		// Your Lightning Bolt, Chain Lightning, and Shock spells have a 20% chance to grant up to 50 Nature damage to spells for 8 sec. (Proc chance: 20%)
		3: func(agent core.Agent) {
			shaman := agent.(ShamanAgent).GetShaman()

			buffAura := shaman.RegisterAura(core.Aura{
				ActionID: core.ActionID{SpellID: 26121},
				Label:    "Stormcaller's Wrath",
				Duration: time.Second * 8,
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					shaman.AddStatDynamic(sim, stats.NaturePower, 50)
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					shaman.AddStatDynamic(sim, stats.NaturePower, -50)
				},
			})

			affectedSpellCodes := []int32{SpellCode_ShamanLightningBolt, SpellCode_ShamanChainLightning, SpellCode_ShamanEarthShock, SpellCode_ShamanFlameShock, SpellCode_ShamanFrostShock}

			core.MakeProcTriggerAura(&shaman.Unit, core.ProcTrigger{
				Name:     "Stormcaller Spelldamage Bonus",
				Callback: core.CallbackOnSpellHitDealt,
				Outcome:  core.OutcomeLanded,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if slices.Contains(affectedSpellCodes, spell.SpellCode) && sim.Proc(0.20, "Stormcaller's Wrath") {
						buffAura.Activate(sim)
					}
				},
			})
		},
		// -0.4 seconds on the casting time of your Chain Heal spell.
		5: func(agent core.Agent) {
		},
	},
})

var ItemSetGiftOfTheGatheringStorm = core.NewItemSet(core.ItemSet{
	Name: "Gift of the Gathering Storm",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases the chain target damage multiplier of your Chain Lightning spell by 5%.
		3: func(agent core.Agent) {
			shaman := agent.(ShamanAgent).GetShaman()
			shaman.RegisterAura(core.Aura{
				Label: "Gift of the Gathering Storm Chain Lightning Bonus",
				OnInit: func(aura *core.Aura, sim *core.Simulation) {
					shaman.ChainLightningBounceCoefficient += 0.05
				},
			})
		},
	},
})

var ItemSetTheEarthshatterer = core.NewItemSet(core.ItemSet{
	Name: "The Earthshatterer",
	Bonuses: map[int32]core.ApplyEffect{
		// Reduces the mana cost of your totem spells by 12%.
		2: func(agent core.Agent) {
			shaman := agent.(ShamanAgent).GetShaman()
			shaman.OnSpellRegistered(func(spell *core.Spell) {
				if spell.Flags.Matches(SpellFlagTotem) {
					spell.Cost.Multiplier -= 12
				}
			})
		},
		// Increases the mana gained from your Mana Spring totems by 25%.
		4: func(agent core.Agent) {
		},
		// Your Healing Wave and Lesser Healing Wave spells have a chance to imbue your target with Totemic Power.
		6: func(agent core.Agent) {
		},
		// Your Lightning Shield spell also grants you 15 mana per 5 sec. while active.
		8: func(agent core.Agent) {
			shaman := agent.(ShamanAgent).GetShaman()
			shaman.RegisterAura(core.Aura{
				Label: "Lightning Shield",
				OnInit: func(_ *core.Aura, _ *core.Simulation) {
					for _, lsAura := range shaman.LightningShieldAuras {
						if lsAura == nil {
							return
						}

						oldOnGain := lsAura.OnGain
						lsAura.OnGain = func(aura *core.Aura, sim *core.Simulation) {
							oldOnGain(aura, sim)
							shaman.AddStatDynamic(sim, stats.MP5, 15)
						}

						oldOnExpire := lsAura.OnExpire
						lsAura.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
							oldOnExpire(aura, sim)
							shaman.AddStatDynamic(sim, stats.MP5, -15)
						}
					}
				},
			})
		},
	},
})

// spiritcallersRageBonuses are the set bonuses of The Spiritcaller's Rage, Forever's crafted level 60
// enhancement set (Spiritcaller Leggings 273932, Grips 273933, Spaulders 273934, Treads 273935, from the
// 1.60.1.70205 client).
//
// The client pieces carry no stats yet, so no real set is registered for them. We give the same bonuses to
// the synthetic Phase 1 enhancement set below, because it is the only enhancement set Forever has at level 60.
var spiritcallersRageBonuses = map[int32]core.ApplyEffect{
	// Increases your attack speed and casting speed by 1%.
	2: func(agent core.Agent) {
		c := agent.GetCharacter()
		c.PseudoStats.MeleeSpeedMultiplier *= 1.01
		c.PseudoStats.RangedSpeedMultiplier *= 1.01
		c.PseudoStats.CastSpeedMultiplier *= 1.01
	},
	// Increases the radius of effect of your beneficial Air and Earth totems by 10 yds.
	3: func(agent core.Agent) {
		// Nothing to do
	},
	// +36 Attack Power against Elementals.
	4: func(agent core.Agent) {
		c := agent.GetCharacter()
		elementals := core.FilterSlice(c.Env.Encounter.TargetUnits, func(unit *core.Unit) bool {
			return unit.MobType == proto.MobType_MobTypeElemental
		})
		core.MakePermanent(c.RegisterAura(core.Aura{
			Label: "Attack Power against Elementals",
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				for _, target := range elementals {
					for _, at := range c.AttackTables[target.UnitIndex] {
						at.BonusAttackPowerTaken += 36
					}
				}
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				for _, target := range elementals {
					for _, at := range c.AttackTables[target.UnitIndex] {
						at.BonusAttackPowerTaken -= 36
					}
				}
			},
		}))
	},
	// Reduces the cooldown of your Stormstrike ability by 0.5 sec.
	5: func(agent core.Agent) {
		shaman := agent.(ShamanAgent).GetShaman()
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellCode == SpellCode_ShamanStormstrike {
				spell.CD.Duration -= time.Millisecond * 500
			}
		})
	},
}

// ItemSetEnhancementSyntheticPhase1 is the set of the synthetic Phase 1 enhancement items
// (tools/database/forever_synthetic_items.go), with The Spiritcaller's Rage bonuses.
var ItemSetEnhancementSyntheticPhase1 = core.NewItemSet(core.ItemSet{
	Name:    "Enhancement Synthetic (Phase 1)",
	Bonuses: spiritcallersRageBonuses,
})
