package item_sets

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's layouts of the Dungeon Set 1 sets, from the client's ItemSetSpell rows (1.60.1.70532),
// the same rows wowsims/forever reads (sim/common/forever/items_sets.go). Every set now grants all
// resistances at 2 pieces, its stat row at 3, a defensive proc when struck at 4 (which a boss fight
// does not test, so it is left out), its offensive proc at 5 and 8 mana per 5 sec or a regeneration
// row at 6. The Classic layouts (2/4/6/8) stay in dungeon_set_1.go for the other rulesets.

// 18679: +8 all resistances.
func foreverDungeonResistances(agent core.Agent) {
	agent.GetCharacter().AddResistances(8)
}

// 9346: increases damage and healing done by magical spells and effects by up to 18.
func foreverDungeonSpellPower(agent core.Agent) {
	agent.GetCharacter().AddStat(stats.SpellPower, 18)
}

// 9336: +30 attack power.
func foreverDungeonAttackPower(agent core.Agent) {
	agent.GetCharacter().AddStats(stats.Stats{stats.AttackPower: 30, stats.RangedAttackPower: 30})
}

// 18378: restores 8 mana per 5 sec.
func foreverDungeonManaRegen(agent core.Agent) {
	agent.GetCharacter().AddStat(stats.MP5, 8)
}

// A 5-piece bonus whose row procs on 5% of magic spell casts, harmful or helpful (proc flags
// 0x14000), and restores mana: Sudden Insight (450554) for Magister's and Devout, Nature's Bounty
// (27782) for Wildheart, 200 mana each.
func foreverSpellCastMana(procID int32, manaID int32, name string) core.ApplyEffect {
	return func(agent core.Agent) {
		c := agent.GetCharacter()
		if !c.HasManaBar() {
			return
		}
		manaMetrics := c.NewManaMetrics(core.ActionID{SpellID: manaID})
		core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
			ActionID:   core.ActionID{SpellID: procID},
			Name:       name,
			Callback:   core.CallbackOnCastComplete,
			ProcMask:   core.ProcMaskSpellDamage | core.ProcMaskSpellHealing,
			ProcChance: 0.05,
			Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				c.AddMana(sim, 200, manaMetrics)
			},
		})
	}
}

// The 5-piece procs of Lightforge (450625, Crusader's Wrath 27499) and The Elements (450626, The
// Furious Storm 27775): +65 spell damage and healing for 10 sec, on melee swings and magic spell
// casts (proc flags 0x14004). The rows state 100% and no rate, so the server sets one. We keep the
// Classic rates of the same buffs as literals: 3 per minute on swings (our Classic Lightforge) and
// 4% of spell casts (Classic's The Furious Storm row, 27774).
func foreverSpellPowerProc(procID int32, buffID int32, buffName string) core.ApplyEffect {
	return func(agent core.Agent) {
		c := agent.GetCharacter()
		buff := c.NewTemporaryStatsAura(buffName, core.ActionID{SpellID: buffID}, stats.Stats{stats.SpellPower: 65}, time.Second*10)
		handler := func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			buff.Activate(sim)
		}
		core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
			ActionID: core.ActionID{SpellID: procID},
			Name:     buffName + " (Melee)",
			Callback: core.CallbackOnSpellHitDealt,
			Outcome:  core.OutcomeLanded,
			ProcMask: core.ProcMaskMeleeWhiteHit,
			PPM:      3,
			Handler:  handler,
		})
		core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
			ActionID:   core.ActionID{SpellID: procID},
			Name:       buffName + " (Spell)",
			Callback:   core.CallbackOnCastComplete,
			ProcMask:   core.ProcMaskSpellDamage | core.ProcMaskSpellHealing,
			ProcChance: 0.04,
			Handler:    handler,
		})
	}
}

// Rogue Armor Energize (27787 → 27788), Shadowcraft's 5-piece bonus under Forever (and Darkmantle's
// 4-piece, in sim/rogue): a melee swing restores 20 Energy (Classic: 35). The row states 100% on
// every swing (proc flags 0x4) and no rate, so we keep 1 per minute as a literal, as
// wowsims/forever does.
func foreverRogueArmorEnergize(agent core.Agent) {
	c := agent.GetCharacter()
	if !c.HasEnergyBar() {
		return
	}
	energyMetrics := c.NewEnergyMetrics(core.ActionID{SpellID: 27788})
	core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
		ActionID: core.ActionID{SpellID: 27787},
		Name:     "Rogue Armor Energize",
		Callback: core.CallbackOnSpellHitDealt,
		Outcome:  core.OutcomeLanded,
		ProcMask: core.ProcMaskMeleeWhiteHit,
		PPM:      1,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			c.AddEnergy(sim, 20, energyMetrics)
		},
	})
}

// https://www.wowhead.com/forever/item-set=185/wildheart-raiment
//
// 4 pieces (1302708): movement speed when struck. The 5-piece row (450608) also answers a melee
// swing with mana over time (27784) and a melee hit taken with rage (27783); like wowsims/forever we
// model only the spell cast's mana.
var foreverWildheartRaiment = map[int32]core.ApplyEffect{
	2: foreverDungeonResistances,
	3: func(agent core.Agent) {
		foreverDungeonSpellPower(agent)
		foreverDungeonAttackPower(agent)
	},
	5: foreverSpellCastMana(450608, 27782, "Wildheart Raiment - 5pc"),
	6: foreverDungeonManaRegen,
}

// https://www.wowhead.com/forever/item-set=186/beaststalker-armor
//
// 4 pieces (1302712): removes a root when struck. 5 pieces (450577): a 5% chance on a melee or
// ranged auto attack (proc flags 0x44) to restore 200 mana (Sudden Insight, 450554).
var foreverBeaststalkerArmor = map[int32]core.ApplyEffect{
	2: foreverDungeonResistances,
	3: foreverDungeonAttackPower,
	5: func(agent core.Agent) {
		c := agent.GetCharacter()
		if !c.HasManaBar() {
			return
		}
		manaMetrics := c.NewManaMetrics(core.ActionID{SpellID: 450554})
		core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
			ActionID:   core.ActionID{SpellID: 450577},
			Name:       "Beaststalker Armor - 5pc",
			Callback:   core.CallbackOnSpellHitDealt,
			Outcome:    core.OutcomeLanded,
			ProcMask:   core.ProcMaskWhiteHit,
			ProcChance: 0.05,
			Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				c.AddMana(sim, 200, manaMetrics)
			},
		})
	},
	6: foreverDungeonManaRegen,
}

// https://www.wowhead.com/forever/item-set=181/magisters-regalia
//
// 4 pieces (27867): a chance to freeze an attacker when struck.
var foreverMagistersRegalia = map[int32]core.ApplyEffect{
	2: foreverDungeonResistances,
	3: foreverDungeonSpellPower,
	5: foreverSpellCastMana(450527, 450554, "Magister's Regalia - 5pc"),
	6: foreverDungeonManaRegen,
}

// https://www.wowhead.com/forever/item-set=188/lightforge-armor
//
// 4 pieces (1302716): silences an attacking caster.
var foreverLightforgeArmor = map[int32]core.ApplyEffect{
	2: foreverDungeonResistances,
	3: foreverDungeonSpellPower,
	5: foreverSpellPowerProc(450625, 27499, "Crusader's Wrath"),
	6: foreverDungeonManaRegen,
}

// https://www.wowhead.com/forever/item-set=182/vestments-of-the-devout
//
// 4 pieces (27778): shields the wearer when struck.
var foreverVestmentsOfTheDevout = map[int32]core.ApplyEffect{
	2: foreverDungeonResistances,
	3: foreverDungeonSpellPower,
	5: foreverSpellCastMana(450576, 450554, "Vestments of the Devout - 5pc"),
	6: foreverDungeonManaRegen,
}

// https://www.wowhead.com/forever/item-set=184/shadowcraft-armor
//
// 4 pieces (1302698): removes a snare when struck. 6 pieces (21347): health regeneration in combat,
// which the sim does not track.
var foreverShadowcraftArmor = map[int32]core.ApplyEffect{
	2: foreverDungeonResistances,
	3: foreverDungeonAttackPower,
	5: foreverRogueArmorEnergize,
}

// https://www.wowhead.com/forever/item-set=187/the-elements
//
// 4 pieces (1302718): disarms an attacker.
var foreverTheElements = map[int32]core.ApplyEffect{
	2: foreverDungeonResistances,
	3: foreverDungeonSpellPower,
	5: foreverSpellPowerProc(450626, 27775, "The Furious Storm"),
	6: foreverDungeonManaRegen,
}

// https://www.wowhead.com/forever/item-set=183/dreadmist-raiment
//
// 4 pieces (27780): fears an attacker when struck. 5 pieces (450585): a 5% chance on a spell cast
// to heal the wearer for 300 (Dark Reward, 450583), which changes no damage, so it is left out.
var foreverDreadmistRaiment = map[int32]core.ApplyEffect{
	2: foreverDungeonResistances,
	3: foreverDungeonSpellPower,
	6: foreverDungeonManaRegen,
}

// https://www.wowhead.com/forever/item-set=189/battlegear-of-valor
//
// 4 pieces (1302721): removes a stun when struck. 6 pieces (21347): health regeneration in combat.
var foreverBattlegearOfValor = map[int32]core.ApplyEffect{
	2: foreverDungeonResistances,
	// 13198: +15 Strength.
	3: func(agent core.Agent) {
		agent.GetCharacter().AddStat(stats.Strength, 15)
	},
	// 450587: a melee swing (proc flags 0x4) heals for 88 to 132 and gives 10 Rage (Warrior's
	// Resolve, 450589). The row states 100% and no rate; we keep our Classic set's 1 per minute.
	5: func(agent core.Agent) {
		c := agent.GetCharacter()
		actionID := core.ActionID{SpellID: 450589}
		healthMetrics := c.NewHealthMetrics(actionID)
		rageMetrics := c.NewRageMetrics(actionID)
		core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
			ActionID: core.ActionID{SpellID: 450587},
			Name:     "Warrior's Resolve",
			Callback: core.CallbackOnSpellHitDealt,
			Outcome:  core.OutcomeLanded,
			ProcMask: core.ProcMaskMeleeWhiteHit,
			PPM:      1,
			Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				c.GainHealth(sim, sim.Roll(88, 132), healthMetrics)
				if c.HasRageBar() {
					c.AddRage(sim, 10, rageMetrics)
				}
			},
		})
	},
}
