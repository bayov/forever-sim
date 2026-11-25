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

// The rates wowsims/forever measured for the dungeon set procs whose rows state 100% and no rate
// (EV-0144, Warcraft Logs 2026-11-11 to 17): procs per minute, a white swing priced on its weapon's
// base speed and a spell cast on its cast time, at least 1.5 sec. Crusader's Wrath reads 0.92 fitted
// to swings and casts together, The Furious Storm and Nature's Bounty about 1, Rogue Armor Energize
// 0.83 (sim rate 0.8) and Warrior's Resolve 0.94 (sim rate 1).
const (
	CrusadersWrathPPM     = 0.9
	FuriousStormPPM       = 1.0
	naturesBountyPPM      = 1.0
	rogueArmorEnergizePPM = 0.8
	warriorsResolvePPM    = 1.0
)

// foreverCastRolls says whether a cast rolls a dungeon set's spell row (EV-0144): a damaging or
// healing spell, or any other cast that costs its caster a resource (a seal, a totem, Judgement,
// Counterspell), but no melee or ranged ability (Holy Strike, Stormstrike). A cast that costs
// nothing rolls nothing, on the global cooldown or off it.
func foreverCastRolls(spell *core.Spell) bool {
	if spell.ProcMask.Matches(core.ProcMaskMeleeOrRanged) {
		return false
	}
	if spell.ProcMask.Matches(core.ProcMaskSpellDamage | core.ProcMaskSpellHealing) {
		return true
	}
	return spell.Cost != nil && spell.Cost.BaseCost > 0
}

// foreverCastTimePPMChance is a rate's chance on a cast that took castTime, priced on at least the
// global cooldown's 1.5 sec, so an instant rolls 2.5% at 1 a minute and a 3 sec cast 5% (EV-0144).
func foreverCastTimePPMChance(ppm float64, castTime time.Duration) float64 {
	return ppm * max(castTime, core.GCDDefault).Seconds() / 60
}

// foreverOnCast calls handler on each cast foreverCastRolls takes, at chance(cast time).
func foreverOnCast(c *core.Character, label string, chance func(castTime time.Duration) float64, handler func(sim *core.Simulation)) {
	core.MakePermanent(c.RegisterAura(core.Aura{
		Label: label,
		OnCastComplete: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.Flags.Matches(core.SpellFlagNoProcs) || !foreverCastRolls(spell) {
				return
			}
			if sim.Proc(chance(spell.CurCast.CastTime), label) {
				handler(sim)
			}
		},
	}))
}

// A 5-piece bonus whose row procs on 5% of magic spell casts (proc flags 0x14000) and restores
// 200 mana: Sudden Insight (450554) for Magister's and Devout. The rows that state a chance fire at
// it on the casts foreverCastRolls takes (EV-0144: Sudden Insight follows 6.99% of Counterspells and
// none of Arcane Power, Presence of Mind or Evocation; Devout's 4.91% of Renews).
func foreverSpellCastMana(procID int32, manaID int32, name string) core.ApplyEffect {
	return func(agent core.Agent) {
		c := agent.GetCharacter()
		if !c.HasManaBar() {
			return
		}
		manaMetrics := c.NewManaMetrics(core.ActionID{SpellID: manaID})
		foreverOnCast(c, name, func(time.Duration) float64 { return 0.05 }, func(sim *core.Simulation) {
			c.AddMana(sim, 200, manaMetrics)
		})
	}
}

// Nature's Bounty (450608): Wildheart's 5-piece and Feralheart's 4-piece. EV-0144 measured three
// parts at about 1 a minute each: a spell cast restores 200 mana (27782) priced on its cast time; a
// landed white swing 20 energy over 5 sec (27784, 4 a sec), priced on the form's paw speed and
// restored in any form; and a landed melee swing taken 10 rage (27783), priced on the attacker's
// swing interval. Each energy proc ticks on its own, so one during the ticks adds five more.
func foreverNaturesBounty(name string) core.ApplyEffect {
	return func(agent core.Agent) {
		c := agent.GetCharacter()
		if c.HasManaBar() {
			manaMetrics := c.NewManaMetrics(core.ActionID{SpellID: 27782})
			foreverOnCast(c, name+" (Mana)", func(castTime time.Duration) float64 {
				return foreverCastTimePPMChance(naturesBountyPPM, castTime)
			}, func(sim *core.Simulation) {
				c.AddMana(sim, 200, manaMetrics)
			})
		}
		if c.HasEnergyBar() {
			energyMetrics := c.NewEnergyMetrics(core.ActionID{SpellID: 27784})
			core.MakePermanent(c.RegisterAura(core.Aura{
				Label: name + " (Energy)",
				OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
						return
					}
					// The swing's own weapon, read at the roll: a paw changes with the form.
					weapon := core.Ternary(spell.ProcMask.Matches(core.ProcMaskMeleeOH), c.AutoAttacks.OH(), c.AutoAttacks.MH())
					if !sim.Proc(naturesBountyPPM*weapon.SwingSpeed/60, name+" (Energy)") {
						return
					}
					core.StartPeriodicAction(sim, core.PeriodicActionOptions{
						Period:   time.Second,
						NumTicks: 5,
						OnAction: func(sim *core.Simulation) {
							c.AddEnergy(sim, 4, energyMetrics)
						},
					})
				},
			}))
		}
		if c.HasRageBar() {
			rageMetrics := c.NewRageMetrics(core.ActionID{SpellID: 27783})
			core.MakePermanent(c.RegisterAura(core.Aura{
				Label: name + " (Rage)",
				OnSpellHitTaken: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
						return
					}
					speed := spell.Unit.AutoAttacks.MainhandSwingSpeed().Seconds()
					if sim.Proc(naturesBountyPPM*speed/60, name+" (Rage)") {
						c.AddRage(sim, 10, rageMetrics)
					}
				},
			}))
		}
	}
}

// The 5-piece procs of Lightforge (450625, Crusader's Wrath 27499) and The Elements (450626, The
// Furious Storm 27775): +65 spell damage and healing for 10 sec, on melee swings and magic spell
// casts (proc flags 0x14004). The rows state 100% and no rate, so the server sets one: EV-0144
// measured it as procs per minute on landed white swings and on the casts foreverCastRolls takes,
// priced on their cast time. Soulforge's 4-piece (set 1673) uses the Lightforge row too. Two sets
// that proc the same buff share it, since 4 + 4 pieces fit.
func ForeverSpellPowerProc(procID int32, buffID int32, buffName string, ppm float64) core.ApplyEffect {
	return func(agent core.Agent) {
		c := agent.GetCharacter()
		buff := c.GetAura(buffName)
		if buff == nil {
			buff = c.NewTemporaryStatsAura(buffName, core.ActionID{SpellID: buffID}, stats.Stats{stats.SpellPower: 65}, time.Second*10)
		}
		core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
			ActionID: core.ActionID{SpellID: procID},
			Name:     buffName + " (Melee)",
			Callback: core.CallbackOnSpellHitDealt,
			Outcome:  core.OutcomeLanded,
			ProcMask: core.ProcMaskMeleeWhiteHit,
			PPM:      ppm,
			Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				buff.Activate(sim)
			},
		})
		foreverOnCast(c, buffName+" (Spell)", func(castTime time.Duration) float64 {
			return foreverCastTimePPMChance(ppm, castTime)
		}, func(sim *core.Simulation) {
			buff.Activate(sim)
		})
	}
}

// Rogue Armor Energize (27787 → 27788), Shadowcraft's 5-piece bonus under Forever (and Darkmantle's
// 4-piece, in sim/rogue): a melee swing restores 20 Energy (Classic: 35). The row states 100% on
// every swing (proc flags 0x4) and no rate; EV-0144 measured 0.83 a minute on 70,840 swings, priced
// on the weapon's base speed (0.8 in their sim). Each set names its own proc aura, since
// Darkmantle's 4-piece (set 512) and Forever Darkmantle's 4-piece (set 1676) fit together.
func ForeverRogueArmorEnergize(name string) core.ApplyEffect {
	return func(agent core.Agent) {
		c := agent.GetCharacter()
		if !c.HasEnergyBar() {
			return
		}
		energyMetrics := c.NewEnergyMetrics(core.ActionID{SpellID: 27788})
		core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
			ActionID: core.ActionID{SpellID: 27787},
			Name:     name,
			Callback: core.CallbackOnSpellHitDealt,
			Outcome:  core.OutcomeLanded,
			ProcMask: core.ProcMaskMeleeWhiteHit,
			PPM:      rogueArmorEnergizePPM,
			Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				c.AddEnergy(sim, 20, energyMetrics)
			},
		})
	}
}

// 450577: a 5% chance on a melee or ranged auto attack (proc flags 0x44) to restore 200 mana
// (Sudden Insight, 450554). Beaststalker's 5-piece and Beastmaster's 4-piece.
func foreverAutoAttackMana(name string) core.ApplyEffect {
	return func(agent core.Agent) {
		c := agent.GetCharacter()
		if !c.HasManaBar() {
			return
		}
		manaMetrics := c.NewManaMetrics(core.ActionID{SpellID: 450554})
		core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
			ActionID:   core.ActionID{SpellID: 450577},
			Name:       name,
			Callback:   core.CallbackOnSpellHitDealt,
			Outcome:    core.OutcomeLanded,
			ProcMask:   core.ProcMaskWhiteHit,
			ProcChance: 0.05,
			Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				c.AddMana(sim, 200, manaMetrics)
			},
		})
	}
}

// 450587: a melee swing (proc flags 0x4) heals for 88 to 132 (110 with 0.4 variance) and gives 10
// Rage (Warrior's Resolve, 450589). Valor's 5-piece and Heroism's 4-piece. The row states 100% and
// no rate; EV-0144 measured 0.94 a minute on 21,667 swings, and we keep 1, as wowsims/forever does. Each set names its own proc aura, since
// Classic Heroism's 4-piece (set 511) and Forever Heroism's 4-piece (set 1778) fit together.
func foreverWarriorsResolve(name string) core.ApplyEffect {
	return func(agent core.Agent) {
		c := agent.GetCharacter()
		actionID := core.ActionID{SpellID: 450589}
		healthMetrics := c.NewHealthMetrics(actionID)
		rageMetrics := c.NewRageMetrics(actionID)
		core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
			ActionID: core.ActionID{SpellID: 450587},
			Name:     name,
			Callback: core.CallbackOnSpellHitDealt,
			Outcome:  core.OutcomeLanded,
			ProcMask: core.ProcMaskMeleeWhiteHit,
			PPM:      warriorsResolvePPM,
			Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				c.GainHealth(sim, sim.Roll(88, 132), healthMetrics)
				if c.HasRageBar() {
					c.AddRage(sim, 10, rageMetrics)
				}
			},
		})
	}
}

// https://www.wowhead.com/forever/item-set=185/wildheart-raiment
//
// 4 pieces (1302708): movement speed when struck. 5 pieces: Nature's Bounty (450608), see
// foreverNaturesBounty.
var foreverWildheartRaiment = map[int32]core.ApplyEffect{
	2: foreverDungeonResistances,
	3: func(agent core.Agent) {
		foreverDungeonSpellPower(agent)
		foreverDungeonAttackPower(agent)
	},
	5: foreverNaturesBounty("Wildheart Raiment - 5pc"),
	6: foreverDungeonManaRegen,
}

// https://www.wowhead.com/forever/item-set=186/beaststalker-armor
//
// 4 pieces (1302712): removes a root when struck. 5 pieces (450577): a 5% chance on a melee or
// ranged auto attack (proc flags 0x44) to restore 200 mana (Sudden Insight, 450554).
var foreverBeaststalkerArmor = map[int32]core.ApplyEffect{
	2: foreverDungeonResistances,
	3: foreverDungeonAttackPower,
	5: foreverAutoAttackMana("Beaststalker Armor - 5pc"),
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
	5: ForeverSpellPowerProc(450625, 27499, "Crusader's Wrath", CrusadersWrathPPM),
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
	5: ForeverRogueArmorEnergize("Rogue Armor Energize"),
}

// https://www.wowhead.com/forever/item-set=187/the-elements
//
// 4 pieces (1302718): disarms an attacker.
var foreverTheElements = map[int32]core.ApplyEffect{
	2: foreverDungeonResistances,
	3: foreverDungeonSpellPower,
	5: ForeverSpellPowerProc(450626, 27775, "The Furious Storm", FuriousStormPPM),
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
	5: foreverWarriorsResolve("Warrior's Resolve"),
}
