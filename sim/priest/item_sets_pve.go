package priest

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

var ItemSetVestmentsOfProphecy = core.NewItemSet(core.ItemSet{
	Name: "Vestments of Prophecy",
	Bonuses: map[int32]core.ApplyEffect{
		// -0.1 sec to the casting time of your Flash Heal spell.
		3: func(agent core.Agent) {
			// Nothing to do
		},
		// Improves your chance to get a critical strike with Holy spells by 2%.
		5: func(agent core.Agent) {
			priest := agent.(PriestAgent).GetPriest()
			priest.PseudoStats.SchoolBonusCritChance[stats.SchoolIndexHoly] += 2 * core.SpellCritRatingPerCritChance
		},
		// Increases your chance of a critical hit with Prayer of Healing by 25%.
		8: func(agent core.Agent) {
			// Nothing to do
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            Classic Phase 3 Item Sets - BWL
///////////////////////////////////////////////////////////////////////////

var ItemSetVestmentsOfTranscendence = core.NewItemSet(core.ItemSet{
	Name: "Vestments of Transcendence",
	Bonuses: map[int32]core.ApplyEffect{
		// Allows 15% of your Mana regeneration to continue while casting.
		3: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.PseudoStats.SpiritRegenRateCasting += .15
		},
		// When struck in melee there is a 50% chance you will Fade for 4 sec.
		5: func(agent core.Agent) {
			// Nothing to do
		},
		// Your Greater Heals now have a heal over time component equivalent to a rank 5 Renew.
		8: func(agent core.Agent) {
			// Nothing to do
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            Classic Phase 4 Item Sets - ZG and AB
///////////////////////////////////////////////////////////////////////////

var ItemSetConfessorsRaiment = core.NewItemSet(core.ItemSet{
	Name: "Confessor's Raiment",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases healing done by spells and effects by up to 22.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.HealingPower, 22)
		},
		// Increase the range of your Smite and Holy Fire spells by 5 yds.
		3: func(agent core.Agent) {
			// Nothing to do
		},
		// Reduces the casting time of your Mind Control spell by 0.5 sec.
		5: func(agent core.Agent) {
			// Nothing to do
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            Classic Phase 5 Item Sets - AQ
///////////////////////////////////////////////////////////////////////////

var ItemSetGarmentsOfTheOracle = core.NewItemSet(core.ItemSet{
	Name: "Garments of the Oracle",
	Bonuses: map[int32]core.ApplyEffect{
		// 20% chance that your heals on others will also heal you 10% of the amount healed.
		3: func(agent core.Agent) {
			// Nothing to do
		},
		// Increases the duration of your Renew spell by 3 sec.
		5: func(agent core.Agent) {
			// Nothing to do
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            Classic Phase 6 Item Sets - Naxx
///////////////////////////////////////////////////////////////////////////

var ItemSetVestmentsOfFaith = core.NewItemSet(core.ItemSet{
	Name: "Vestments of Faith",
	Bonuses: map[int32]core.ApplyEffect{
		// Reduces the mana cost of your Renew spell by 12%.
		2: func(agent core.Agent) {
			// Nothing to do
		},
		// On Greater Heal critical hits, your target will gain Armor of Faith, absorbing up to 500 damage.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// Reduces the threat from your healing spells.
		6: func(agent core.Agent) {
			// Nothing to do
		},
		// Each spell you cast can trigger an Epiphany, increasing your mana regeneration by 24 for 30 sec.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()

			procAura := c.NewTemporaryStatsAura("Epiphany", core.ActionID{SpellID: 28802}, stats.Stats{stats.MP5: 24}, time.Second*30)
			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				Name:       "Item - Epiphany Proc (Spell Cast)",
				Callback:   core.CallbackOnCastComplete,
				ProcMask:   core.ProcMaskSpellDamage,
				SpellFlags: core.SpellFlagHelpful,
				ProcChance: 0.05,
				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					procAura.Activate(sim)
				},
			})
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            Forever Level 60 Item Sets
///////////////////////////////////////////////////////////////////////////

// raimentsOfConvictionBonuses are the set bonuses of Raiments of Conviction, Forever's
// Tier 1 shadow priest set (ItemSet 2105 and ItemSetSpell in the 1.60.1.70338 client,
// "1.60.0 - Item - Tier 1 - Priest - Shadow" spells 1300955, 1301045, 1301087, 1301706).
//
// The client pieces carry no stats yet, so no real set is registered for them. We give the
// same bonuses to the synthetic Phase 1 shadow set below, like the enhancement set carries
// The Spiritcaller's Rage.
var raimentsOfConvictionBonuses = map[int32]core.ApplyEffect{
	// Improves your chance to hit by 1%. Melee and spell hit (auras 54 and 55).
	2: func(agent core.Agent) {
		c := agent.GetCharacter()
		c.AddStats(stats.Stats{
			stats.MeleeHit: 1 * core.MeleeHitRatingPerHitChance,
			stats.SpellHit: 1 * core.SpellHitRatingPerHitChance,
		})
	},
	// Increases the duration of your Shackle Undead spell by 10 sec.
	3: func(agent core.Agent) {
		// Nothing to do
	},
	// Increases damage done by your spells and effects by up to 21 when fighting Undead.
	4: func(agent core.Agent) {
		c := agent.GetCharacter()
		undead := core.FilterSlice(c.Env.Encounter.TargetUnits, func(unit *core.Unit) bool {
			return unit.MobType == proto.MobType_MobTypeUndead
		})
		core.MakePermanent(c.RegisterAura(core.Aura{
			Label: "Spell Damage against Undead",
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				for _, target := range undead {
					for _, at := range c.AttackTables[target.UnitIndex] {
						at.BonusSpellDamageTaken += 21
					}
				}
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				for _, target := range undead {
					for _, at := range c.AttackTables[target.UnitIndex] {
						at.BonusSpellDamageTaken -= 21
					}
				}
			},
		}))
	},
	// Reduces the cooldown of your Devouring Plague spell by 60 sec, which is all of its
	// 1 min cooldown under Forever.
	5: func(agent core.Agent) {
		priest := agent.(PriestAgent).GetPriest()
		priest.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellCode == SpellCode_PriestDevouringPlague {
				spell.CD.Duration = max(0, spell.CD.Duration-time.Minute)
			}
		})
	},
}

// ItemSetShadowSyntheticPhase1 is the set of the synthetic Phase 1 shadow priest items
// (tools/database/forever_synthetic_items_priest.go), with the Raiments of Conviction
// bonuses.
var ItemSetShadowSyntheticPhase1 = core.NewItemSet(core.ItemSet{
	Name:    "Shadow Synthetic (Phase 1)",
	Bonuses: raimentsOfConvictionBonuses,
})

// vestmentsOfConvictionBonuses are the set bonuses of Vestments of Conviction, Forever's
// Tier 1 Discipline and Holy set (ItemSet 2104 in the 1.60.1.70338 client, spells 1300954,
// 1301041, 1301085, 1301703). Like Raiments of Conviction, its client pieces carry no stats.
var vestmentsOfConvictionBonuses = map[int32]core.ApplyEffect{
	// Increased Spirit +10.
	2: func(agent core.Agent) {
		agent.GetCharacter().AddStat(stats.Spirit, 10)
	},
	// Reduces the cooldown of your Fear Ward spell by 60 sec.
	3: func(agent core.Agent) {
		// Nothing to do
	},
	// Increases healing done by up to 26 and damage done by up to 9 for all magical spells
	// and effects.
	4: func(agent core.Agent) {
		agent.GetCharacter().AddStats(stats.Stats{stats.SpellDamage: 9, stats.HealingPower: 26})
	},
	// Reduces the cooldown of your Penance and Prayer of Mending spells by 1 sec.
	5: func(agent core.Agent) {
		priest := agent.(PriestAgent).GetPriest()
		priest.OnSpellRegistered(func(spell *core.Spell) {
			if spell.CD.Timer != nil && slices.Contains(PenanceSpellId[:], spell.SpellID) {
				spell.CD.Duration -= time.Second
			}
		})
	},
}

// ItemSetSmiteSyntheticPhase1 is the set of the synthetic Phase 1 Smite items, the shadow
// set's armor pieces with the Vestments of Conviction bonuses: 10 Spirit and 9 spell damage
// beat Raiments' 1% hit by 4.1 DPS at 180 sec and 7.8 at 300 for the Smite build.
var ItemSetSmiteSyntheticPhase1 = core.NewItemSet(core.ItemSet{
	Name:    "Smite Synthetic (Phase 1)",
	Bonuses: vestmentsOfConvictionBonuses,
})
