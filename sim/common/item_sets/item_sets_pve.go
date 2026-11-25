package item_sets

import (
	"slices"
	"time"

	"github.com/wowsims/classic/sim/common/guardians"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

///////////////////////////////////////////////////////////////////////////
//                                 Cloth
///////////////////////////////////////////////////////////////////////////

var ItemSetNecropileRaiment = core.NewItemSet(core.ItemSet{
	Name: "Necropile Raiment",
	Bonuses: map[int32]core.ApplyEffect{
		// +3 Defense.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 5)
		},
		// +5 Intellect.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Intellect, 5)
			if character.Env.IsForever() {
				foreverNecropileDrain(character)
			}
		},
		// +15 All Resistances.
		4: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddResistances(15)
		},
		// Increases damage and healing done by magical spells and effects by up to 23.
		5: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.SpellPower, 23)
		},
	},
})

var ItemSetIronweaveBattlesuit = core.NewItemSet(core.ItemSet{
	Name: "Ironweave Battlesuit",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases your chance to resist Silence and Interrupt effects by 10%.
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// +200 Armor.
		8: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Armor, 200)
		},
	},
})

var ItemSetThePostmaster = core.NewItemSet(core.ItemSet{
	Name: "The Postmaster",
	Bonuses: map[int32]core.ApplyEffect{
		// +50 Armor.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Armor, 50)
		},
		// +10 Fire Resistance.
		// +10 Arcane Resistance.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.ArcaneResistance, 10)
			character.AddStat(stats.FireResistance, 10)
		},
		// Increases damage and healing done by magical spells and effects by up to 12.
		4: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.SpellPower, 12)
		},
		// Increases run speed by 5%.
		// +10 Intellect.
		5: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Intellect, 10)
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                                 Leather
///////////////////////////////////////////////////////////////////////////

// Embrace of the Viper uses the Forever set bonuses (Wailing Caverns leather set).
//
// Classic gave it Defense, resistances and a small Intellect bonus. The Forever
// beta reworked it into a leveling DPS set. The 4 piece heal when below 25%
// health and the 5 piece Dream Venom stun (1 sec, bosses are immune) do not
// change damage, so we only model the stat bonuses.
var ItemSetEmbraceOfTheViper = core.NewItemSet(core.ItemSet{
	Name: "Embrace of the Viper",
	Bonuses: map[int32]core.ApplyEffect{
		// +10 Intellect.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Intellect, 10)
		},
		// +10 Attack Power.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.AttackPower, 10)
		},
	},
})

// Defias Leather uses the Forever set bonuses (the Deadmines leather set).
//
// Classic gave it Stamina, Defense and a dodge bonus. Forever made it a rogue set, with
// Attack Power against Humanoids, a bleed when striking from behind and a point of
// Dagger skill.
var ItemSetDefiasLeather = core.NewItemSet(core.ItemSet{
	Name: "Defias Leather",
	Bonuses: map[int32]core.ApplyEffect{
		// +5 Arcane Resistance.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.ArcaneResistance, 5)
		},
		// +15 Attack Power against Humanoids.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			humanoids := core.FilterSlice(character.Env.Encounter.TargetUnits, func(unit *core.Unit) bool {
				return unit.MobType == proto.MobType_MobTypeHumanoid
			})
			core.MakePermanent(character.RegisterAura(core.Aura{
				ActionID: core.ActionID{SpellID: 1292025},
				Label:    "Humanoid Slaying 15",
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					for _, target := range humanoids {
						for _, at := range character.AttackTables[target.UnitIndex] {
							at.BonusAttackPowerTaken += 15
						}
					}
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					for _, target := range humanoids {
						for _, at := range character.AttackTables[target.UnitIndex] {
							at.BonusAttackPowerTaken -= 15
						}
					}
				},
			}))
		},
		// Grants your melee attacks a 5% chance when striking from behind to hit a vital
		// point, causing the target to Bleed for 68 to 83 Physical damage over 5 sec.
		//
		// The tooltip does not say how the bleed ticks, we assume once a second. A new
		// proc starts the bleed over with a fresh roll.
		4: func(agent core.Agent) {
			character := agent.GetCharacter()
			bleed := character.RegisterSpell(core.SpellConfig{
				ActionID:         core.ActionID{SpellID: 1292028},
				SpellSchool:      core.SpellSchoolPhysical,
				DefenseType:      core.DefenseTypeMelee,
				ProcMask:         core.ProcMaskEmpty,
				Flags:            core.SpellFlagPureDot | core.SpellFlagNoOnCastComplete,
				DamageMultiplier: 1,
				ThreatMultiplier: 1,
				Dot: core.DotConfig{
					Aura: core.Aura{
						Label: "Devious Strike",
					},
					NumberOfTicks: 5,
					TickLength:    time.Second,
					OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, _ bool) {
						dot.Snapshot(target, sim.Roll(67.5, 82.5)/5, false)
					},
					OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
						dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
					},
				},
				ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
					spell.Dot(target).Apply(sim)
				},
			})
			core.MakeProcTriggerAura(&character.Unit, core.ProcTrigger{
				ActionID:   core.ActionID{SpellID: 1292028},
				Name:       "Devious Strike Trigger",
				Callback:   core.CallbackOnSpellHitDealt,
				Outcome:    core.OutcomeLanded,
				ProcMask:   core.ProcMaskMelee,
				ProcChance: 0.05,
				Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
					if !character.PseudoStats.InFrontOfTarget {
						bleed.Cast(sim, result.Target)
					}
				},
			})
		},
		// Increased Daggers +1.
		5: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.PseudoStats.DaggersSkill += 1
		},
	},
})

var ItemSetCadaverousGarb = core.NewItemSet(core.ItemSet{
	Name: "Cadaverous Garb",
	Bonuses: map[int32]core.ApplyEffect{
		// Increased Defense +3.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Defense, 3)
		},
		// +10 Attack Power.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.AttackPower, 10)
			character.AddStat(stats.RangedAttackPower, 10)
		},
		// +15 All Resistances.
		4: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddResistances(15)
		},
		// Improves your chance to hit by 2%.
		5: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.MeleeHit, 2)
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                                 Mail
///////////////////////////////////////////////////////////////////////////

var ItemSetBloodmailRegalia = core.NewItemSet(core.ItemSet{
	Name: "Bloodmail Regalia",
	Bonuses: map[int32]core.ApplyEffect{
		// Increased Defense +3.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Defense, 3)
		},
		// +10 Attack Power.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.AttackPower, 10)
			character.AddStat(stats.RangedAttackPower, 10)
			if character.Env.IsForever() {
				foreverBloodmailBleed(character)
			}
		},
		// +15 All Resistances.
		4: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddResistances(15)
		},
		// Increases your chance to parry an attack by 1%.
		5: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Parry, 1)
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                                 Plate
///////////////////////////////////////////////////////////////////////////

var ItemSetDeathboneGuardian = core.NewItemSet(core.ItemSet{
	Name: "Deathbone Guardian",
	Bonuses: map[int32]core.ApplyEffect{
		// Increased Defense +3.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Defense, 3)
		},
		// +50 Armor.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Armor, 50)
		},
		// +15 All Resistances.
		4: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddResistances(15)
		},
		// Increases your chance to parry an attack by 1%
		5: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Parry, 1)
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                                 Other
///////////////////////////////////////////////////////////////////////////

var ItemSetSpidersKiss = core.NewItemSet(core.ItemSet{
	Name: "Spider's Kiss",
	Bonuses: map[int32]core.ApplyEffect{
		// Chance on Hit: Immobilizes the target and lowers their armor by 100 for 10 sec.
		// Unsure about exlusive effects with this aura also looks like it might be lowering the characters armor instead of the enemy?
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			procAura := character.NewTemporaryStatsAura("Spider's Kiss", core.ActionID{SpellID: 17333}, stats.Stats{stats.Armor: -100}, time.Second*10)
			core.MakeProcTriggerAura(&character.Unit, core.ProcTrigger{
				ActionID:   core.ActionID{SpellID: 17333},
				Name:       "Spider's Kiss",
				Callback:   core.CallbackOnSpellHitDealt,
				Outcome:    core.OutcomeLanded,
				ProcMask:   core.ProcMaskMelee,
				ProcChance: 0.05,
				Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
					procAura.Activate(sim)
				},
			})
		},
	},
})

var ItemSetDalRendsArms = core.NewItemSet(core.ItemSet{
	Name: "Dal'Rend's Arms",
	Bonuses: map[int32]core.ApplyEffect{
		// +50 Attack Power.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.AttackPower, 50)
			character.AddStat(stats.RangedAttackPower, 50)
		},
	},
})

var ItemSetShardOfTheGods = core.NewItemSet(core.ItemSet{
	Name: "Shard of the Gods",
	Bonuses: map[int32]core.ApplyEffect{
		// +10 All Resistances.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddResistances(15)
		},
	},
})

var ItemSetSpiritOfEskhandar = core.NewItemSet(core.ItemSet{
	Name: "Spirit of Eskhandar",
	Bonuses: map[int32]core.ApplyEffect{
		// 1% chance on a melee hit to call forth the spirit of Eskhandar to protect you in battle for 2 min.
		4: func(agent core.Agent) {
			character := agent.GetCharacter()
			core.MakeProcTriggerAura(&character.Unit, core.ProcTrigger{
				Name:       "Call of Eskhandar Trigger",
				Callback:   core.CallbackOnSpellHitDealt,
				Outcome:    core.OutcomeLanded,
				ProcMask:   core.ProcMaskMelee,
				ProcChance: 1,
				ICD:        time.Minute * 1,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					for _, petAgent := range character.PetAgents {
						if eskhandar, ok := petAgent.(*guardians.Eskhandar); ok {
							eskhandar.EnableWithTimeout(sim, eskhandar, time.Minute*2)
							break
						}
					}
				},
			})
		},
	},
})

var ItemSetMajorMojoInfusion = core.NewItemSet(core.ItemSet{
	Name: "Major Mojo Infusion",
	Bonuses: map[int32]core.ApplyEffect{
		// +30 Attack Power.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStats(stats.Stats{
				stats.AttackPower:       30,
				stats.RangedAttackPower: 30,
			})
		},
	},
})

var ItemSetOverlordsResolution = core.NewItemSet(core.ItemSet{
	Name: "Overlord's Resolution",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases our chance to dodge an attack by 1%
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Dodge, 1)
		},
	},
})

var ItemSetPrayerOfThePrimal = core.NewItemSet(core.ItemSet{
	Name: "Prayer of the Primal",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases healing done by up to 33
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.HealingPower, 33)
		},
	},
})

var ItemSetPrimalBlessing = core.NewItemSet(core.ItemSet{
	Name: "Primal Blessing",
	Bonuses: map[int32]core.ApplyEffect{
		// Grants a small chance when ranged or melee damage is dealt to infuse the wielder with a blessing from the Primal Gods.
		// Ranged and melee attack power increased by 300 for 12 sec.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()

			aura := character.RegisterAura(core.Aura{
				ActionID: core.ActionID{SpellID: 467742},
				Label:    "Primal Blessing",
				Duration: time.Second * 12,
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					character.AddStatsDynamic(sim, stats.Stats{
						stats.AttackPower:       300,
						stats.RangedAttackPower: 300,
					})
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					character.AddStatsDynamic(sim, stats.Stats{
						stats.AttackPower:       -300,
						stats.RangedAttackPower: -300,
					})
				},
			})

			core.MakeProcTriggerAura(&character.Unit, core.ProcTrigger{
				Name:       "Primal Blessing Trigger",
				Callback:   core.CallbackOnSpellHitDealt,
				ProcMask:   core.ProcMaskMeleeOrRanged,
				Outcome:    core.OutcomeLanded,
				ProcChance: 0.05,
				ICD:        time.Second * 72,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					aura.Activate(sim)
				},
			})
		},
	},
})

var ItemSetTwinBladesofHakkari = core.NewItemSet(core.ItemSet{
	Name: "The Twin Blades of Hakkari",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases Swords +6
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.PseudoStats.SwordsSkill += 6
		},
	},
})

var ItemSetZanzilsConcentration = core.NewItemSet(core.ItemSet{
	Name: "Zanzil's Concentration",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases damage and healing done by magical spells and effects by up to 6.
		// Improves your chance to hit with all spells and attacks by 1%.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStats(stats.Stats{
				stats.SpellPower: 6,
				stats.SpellHit:   1 * core.SpellHitRatingPerHitChance,
				stats.MeleeHit:   1 * core.MeleeHitRatingPerHitChance,
			})
		},
	},
})

var ItemSetRegaliaOfUndeadCleansing = core.NewItemSet(core.ItemSet{
	Name: "Regalia of Undead Cleansing",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases your damage against undead by 2%.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			delta := 1.02

			character.Env.RegisterPostFinalizeEffect(func() {
				for _, target := range character.Env.Encounter.TargetUnits {
					if target.MobType != proto.MobType_MobTypeUndead {
						continue
					}

					for _, at := range character.AttackTables[target.UnitIndex] {
						at.DamageDealtMultiplier *= delta
						at.CritMultiplier *= delta
					}
				}
			})
		},
	},
})

var ItemSetUndeadSlayersArmor = core.NewItemSet(core.ItemSet{
	Name: "Undead Slayer's Armor",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases your damage against undead by 2%.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			delta := 1.02

			character.Env.RegisterPostFinalizeEffect(func() {
				for _, target := range character.Env.Encounter.TargetUnits {
					if target.MobType != proto.MobType_MobTypeUndead {
						continue
					}

					for _, at := range character.AttackTables[target.UnitIndex] {
						at.DamageDealtMultiplier *= delta
						at.CritMultiplier *= delta
					}
				}
			})
		},
	},
})

var ItemSetGarbOfTheUndeadSlayer = core.NewItemSet(core.ItemSet{
	Name: "Garb of the Undead Slayer",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases your damage against undead by 2%.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			delta := 1.02

			character.Env.RegisterPostFinalizeEffect(func() {
				for _, target := range character.Env.Encounter.TargetUnits {
					if target.MobType != proto.MobType_MobTypeUndead {
						continue
					}

					for _, at := range character.AttackTables[target.UnitIndex] {
						at.DamageDealtMultiplier *= delta
						at.CritMultiplier *= delta
					}
				}
			})
		},
	},
})

var ItemSetBattlegearOfUndeadSlaying = core.NewItemSet(core.ItemSet{
	Name: "Battlegear of Undead Slaying",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases your damage against undead by 2%.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			delta := 1.02

			character.Env.RegisterPostFinalizeEffect(func() {
				for _, target := range character.Env.Encounter.TargetUnits {
					if target.MobType != proto.MobType_MobTypeUndead {
						continue
					}

					for _, at := range character.AttackTables[target.UnitIndex] {
						at.DamageDealtMultiplier *= delta
						at.CritMultiplier *= delta
					}
				}
			})
		},
	},
})

// The chance of a Forever Scholomance set's 3-piece proc, which its row does not state.
const (
	// EV-0167: Bloodmail Regalia's bleed follows 10.7% of landed white swings and Windfury hits.
	BloodmailBleedChance = 0.107
	// EV-0168: Necropile Raiment's drain follows 13.9% of landed harmful spells that are not ticks.
	NecropileDrainChance = 0.139
)

// Windfury Weapon's hits (shaman.WindfuryWeaponSpellId), which roll Bloodmail Regalia's bleed.
var windfuryWeaponHits = []int32{8232, 8235, 10486, 16362}

// EV-0167: under Forever Bloodmail Regalia's 3-piece (1299740) puts Bloodmail (1299739), 7 Physical a
// second for 10 sec, on the target of 10.7% of landed white swings and Windfury hits, with no hit roll
// of its own; melee abilities never roll it (wowsims/forever, Warcraft Logs). The +10 attack power
// our 3-piece gives is the Forever 2-piece's, so from 3 pieces the totals match.
func foreverBloodmailBleed(character *core.Character) {
	bleed := character.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 1299739},
		SpellSchool:      core.SpellSchoolPhysical,
		DefenseType:      core.DefenseTypeMelee,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagPureDot,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		Dot: core.DotConfig{
			Aura:          core.Aura{Label: "Bloodmail"},
			NumberOfTicks: 10,
			TickLength:    time.Second,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, 7, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.Dot(target).Apply(sim)
		},
	})
	core.MakePermanent(character.RegisterAura(core.Aura{
		Label: "Bloodmail Regalia - 3pc",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !(spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) || slices.Contains(windfuryWeaponHits, spell.ActionID.SpellID)) {
				return
			}
			if sim.Proc(BloodmailBleedChance, "Bloodmail") {
				bleed.Cast(sim, result.Target)
			}
		},
	}))
}

// EV-0168: under Forever Necropile Raiment's 3-piece (1299737) casts Necropile Drain (1299735), 28
// Shadow taken from the target and given to the wearer, on 13.9% of landed harmful spells that are
// not ticks, with no hit roll of its own (wowsims/forever, Warcraft Logs).
func foreverNecropileDrain(character *core.Character) {
	healthMetrics := character.NewHealthMetrics(core.ActionID{SpellID: 1299735})
	var drain *core.Spell
	drain = character.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 1299735},
		SpellSchool:      core.SpellSchoolShadow,
		DefenseType:      core.DefenseTypeMagic,
		ProcMask:         core.ProcMaskSpellDamageProc,
		Flags:            core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealDamage(sim, target, 28, spell.OutcomeAlwaysHit)
			character.GainHealth(sim, result.Damage, healthMetrics)
		},
	})
	core.MakePermanent(character.RegisterAura(core.Aura{
		Label: "Necropile Raiment - 3pc",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskSpellDamage) || spell == drain ||
				spell.Flags.Matches(core.SpellFlagPassiveSpell) {
				return
			}
			if sim.Proc(NecropileDrainChance, "Necropile Drain") {
				drain.Cast(sim, result.Target)
			}
		},
	}))
}
