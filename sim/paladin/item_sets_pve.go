package paladin

import (
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

///////////////////////////////////////////////////////////////////////////
//                            Classic Phase 1 Item Sets - Molten Core
///////////////////////////////////////////////////////////////////////////

var ItemSetLawbringerArmor = core.NewItemSet(core.ItemSet{
	Name: "Lawbringer Armor",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases the chance of triggering a Judgement of Light heal by 10%.
		3: func(agent core.Agent) {
			// Nothing to do
		},
		// Improves your chance to get a critical strike with spells by 1%.
		// Improves your chance to get a critical strike by 1%.
		5: func(agent core.Agent) {
			paladin := agent.(PaladinAgent).GetPaladin()
			paladin.AddStat(stats.MeleeCrit, 1)
			paladin.AddStat(stats.SpellCrit, 1)
		},
		// Gives the Paladin a chance on every melee hit to heal your party for 189 to 211.
		8: func(agent core.Agent) {
			// Nothing to do
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            Classic Phase 3 Item Sets - BWL
///////////////////////////////////////////////////////////////////////////

var ItemSetJudgementArmor = core.NewItemSet(core.ItemSet{
	Name: "Judgement Armor",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases the radius of a Paladin's auras by 10.
		3: func(agent core.Agent) {
			// Nothing to do
		},
		// Increases damage and healing done by magical spells and effects by up to 47.
		5: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 47)
		},
		// Inflicts 60 to 66 additional Holy damage on the target of a Paladin's Judgement.
		8: func(agent core.Agent) {
			paladin := agent.(PaladinAgent).GetPaladin()

			spellCodes := []int32{SpellCode_PaladinJudgementOfCommand, SpellCode_PaladinJudgementOfRighteousness}
			paladin.RegisterAura(core.Aura{
				Label: "Judgement - T2 - Paladin - 8P Bonus",
				OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if slices.Contains(spellCodes, spell.SpellCode) && result.Landed() {
						spell.CalcAndDealDamage(sim, result.Target, sim.Roll(60, 66), spell.OutcomeMagicCrit)
					}
				},
			})
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            Classic Phase 4 Item Sets - ZG and AB
///////////////////////////////////////////////////////////////////////////

var ItemSetFreethinkersArmor = core.NewItemSet(core.ItemSet{
	Name: "Freethinker's Armor",
	Bonuses: map[int32]core.ApplyEffect{
		// Restores 4 mana per 5 sec.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.MP5, 4)
		},
		// Reduces the casting time of your Holy Light spell by 0.1 sec.
		3: func(agent core.Agent) {
			// Nothing to do
		},
		// Increases the duration of all Blessings by 10%.
		5: func(agent core.Agent) {
			// Nothing to do
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            Classic Phase 5 Item Sets - AQ
///////////////////////////////////////////////////////////////////////////

var ItemSetAvengersBattlegear = core.NewItemSet(core.ItemSet{
	Name: "Avenger's Battlegear",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases the duration of your Judgements by 20%.
		3: func(agent core.Agent) {
			// Nothing to do
		},
		// Increases damage and healing done by magical spells and effects by up to 71.
		5: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellPower, 71)
		},
	},
})

var ItemSetBattlegearOfEternalJustice = core.NewItemSet(core.ItemSet{
	Name: "Battlegear of Eternal Justice",
	Bonuses: map[int32]core.ApplyEffect{
		// 20% chance to regain 100 mana when you cast a Judgement.
		3: func(agent core.Agent) {
			paladin := agent.(PaladinAgent).GetPaladin()

			manaMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: 0})

			core.MakeProcTriggerAura(&paladin.Unit, core.ProcTrigger{
				Name:       "Battlegear of Eternal Justice 3 pc",
				Callback:   core.CallbackOnCastComplete,
				ProcChance: 0.20,
				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					if spell.SpellCode == SpellCode_PaladinJudgementOfCommand || spell.SpellCode == SpellCode_PaladinJudgementOfRighteousness {
						paladin.AddMana(sim, 100, manaMetrics)
					}
				},
			})
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            Classic Phase 6 Item Sets - Naxx
///////////////////////////////////////////////////////////////////////////

var ItemSetRedemptionArmor = core.NewItemSet(core.ItemSet{
	Name: "Redemption Armor",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases the amount healed by your Judgement of Light by 20.
		2: func(agent core.Agent) {
			// Nothing to do
		},
		// Reduces cooldown on your Lay on Hands by 12 min.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// Your Flash of Light and Holy Light spells have a chance to imbue your target with Holy Power.
		6: func(agent core.Agent) {
			// Nothing to do
		},
		// Your Cleanse spell also heals the target for 200.
		8: func(agent core.Agent) {
			// Nothing to do
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            Forever Item Sets
///////////////////////////////////////////////////////////////////////////

// justiceBattlegearBonuses are the set bonuses of Justice Battlegear, Forever's Tier 1 retribution set
// (item set 2106: head 280930, shoulders 280929, chest 280928, legs 280927, hands 280925, feet 280926,
// from the 1.60.1.70338 client, spells 1300951, 1301036, 1301084 and 1301702).
//
// The client pieces carry no stats yet, so no real set is registered for them. We give the same bonuses
// to the synthetic Phase 1 retribution set below.
var justiceBattlegearBonuses = map[int32]core.ApplyEffect{
	// Increases your attack speed and casting speed by 1%.
	2: func(agent core.Agent) {
		c := agent.GetCharacter()
		c.PseudoStats.MeleeSpeedMultiplier *= 1.01
		c.PseudoStats.RangedSpeedMultiplier *= 1.01
		c.PseudoStats.CastSpeedMultiplier *= 1.01
	},
	// Reduces the cooldown of your Hammer of Justice spell by 5 sec.
	3: func(agent core.Agent) {
		// Nothing to do
	},
	// +36 Attack Power against Undead.
	4: func(agent core.Agent) {
		c := agent.GetCharacter()
		undead := core.FilterSlice(c.Env.Encounter.TargetUnits, func(unit *core.Unit) bool {
			return unit.MobType == proto.MobType_MobTypeUndead
		})
		core.MakePermanent(c.RegisterAura(core.Aura{
			Label: "Attack Power against Undead",
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				for _, target := range undead {
					for _, at := range c.AttackTables[target.UnitIndex] {
						at.BonusAttackPowerTaken += 36
					}
				}
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				for _, target := range undead {
					for _, at := range c.AttackTables[target.UnitIndex] {
						at.BonusAttackPowerTaken -= 36
					}
				}
			},
		}))
	},
	// Reduces the cooldown of your Judgement spell by 0.5 sec.
	5: func(agent core.Agent) {
		paladin := agent.(PaladinAgent).GetPaladin()
		paladin.OnSpellRegistered(func(spell *core.Spell) {
			if spell.ActionID.SpellID == JudgementSpellId {
				spell.CD.Duration -= time.Millisecond * 500
			}
		})
	},
}

// ItemSetRetributionSyntheticPhase1 is the set of the synthetic Phase 1 retribution items
// (tools/database/forever_synthetic_items_paladin.go), with the Justice Battlegear bonuses.
var ItemSetRetributionSyntheticPhase1 = core.NewItemSet(core.ItemSet{
	Name:    "Retribution Synthetic (Phase 1)",
	Bonuses: justiceBattlegearBonuses,
})
