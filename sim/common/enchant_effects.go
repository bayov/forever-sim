package common

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// RevelationEffectID is Forever's Enchant Weapon - Revelation, by its spell ID like the other
// new Forever enchants (tools/database/forever_enchants.go).
const RevelationEffectID = 1248805

// RevelationProcChance, when 0 or more, replaces Revelation's chance for every class (rotopt's
// -revelation-chance). Below 0 the sim uses RevelationClassChance.
var RevelationProcChance = -1.0

// RevelationClassChance is the chance of a landed direct spell hit that does not crit to give
// Revelation, by class, for a level 60 wearer on spells at most 12 levels under them: the rates
// wowsims/forever's EV-0149 counts on 1,356 Warcraft Logs reports (mage 27,896 procs of 312,128
// rolls, warlock 3,385 of 38,274, priest 441 of 5,101, druid 3,328 of 42,144, shaman 2,972 of
// 40,695, paladin 426 of 6,066). The hunter's rate is not counted and takes the enchant's 8%. No
// warrior or rogue gains Revelation in those logs, so they have none. The tooltip's fall with the
// wearer's crit (about 0.27 points a crit percent in the same logs) has no formula yet and is left
// out. Until 2026-11-25 the sim gave every class 7.2%, a guess.
var RevelationClassChance = map[proto.Class]float64{
	proto.Class_ClassMage:    0.0894,
	proto.Class_ClassWarlock: 0.0884,
	proto.Class_ClassPriest:  0.0865,
	proto.Class_ClassDruid:   0.0790,
	proto.Class_ClassShaman:  0.0730,
	proto.Class_ClassPaladin: 0.0702,
	proto.Class_ClassHunter:  0.08,
}

// revelationLevelShare is the share of Revelation's chance a spell of learnLevel keeps for a
// wearer of level: all of it down to 12 levels under the wearer, none from 32 under, and a straight
// line between (EV-0149, clamp((SpellLevel + 32 - level) / 20, 0, 1)). A spell with no level of its
// own keeps all of it.
func revelationLevelShare(level int32, learnLevel int) float64 {
	if learnLevel <= 0 {
		return 1
	}
	return max(0, min(1, float64(learnLevel+32-int(level))/20))
}

func init() {
	core.AddEffectsToTest = false

	///////////////////////////////////////////////////////////////////////////
	//                        All effects ordered by ID
	///////////////////////////////////////////////////////////////////////////

	// Ranged Scopes
	core.AddWeaponEffect(32, func(agent core.Agent, _ proto.ItemSlot) {
		w := agent.GetCharacter().AutoAttacks.Ranged()
		w.BaseDamageMin += 2
		w.BaseDamageMax += 2
	})

	// Accurate Scope
	core.AddWeaponEffect(33, func(agent core.Agent, _ proto.ItemSlot) {
		w := agent.GetCharacter().AutoAttacks.Ranged()
		w.BaseDamageMin += 3
		w.BaseDamageMax += 3
	})

	// Weapon - Fiery Blaze
	core.NewEnchantEffect(36, func(agent core.Agent) {
		character := agent.GetCharacter()

		procMask := character.GetProcMaskForEnchant(36)
		procChance := 0.15

		procSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 6296},
			SpellSchool: core.SpellSchoolFire,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellDamage,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
					damage := sim.Roll(9, 13)
					spell.CalcAndDealDamage(sim, aoeTarget, damage, spell.OutcomeMagicHitAndCrit)
				}

			},
		})

		aura := character.GetOrRegisterAura(core.Aura{
			Label:    "Fiery Blaze",
			Duration: core.NeverExpires,
			OnReset: func(aura *core.Aura, sim *core.Simulation) {
				aura.Activate(sim)
			},
			OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !result.Landed() || !spell.ProcMask.Matches(procMask) || spell.Flags.Matches(core.SpellFlagSuppressWeaponProcs) {
					return
				}

				if sim.RandomFloat("Fiery Blaze") < procChance {
					procSpell.Cast(sim, result.Target)
				}
			},
		})

		character.ItemSwap.RegisterOnSwapItemForEffect(36, aura)
	})

	// Weapon - Lesser Striking
	core.AddWeaponEffect(241, func(agent core.Agent, slot proto.ItemSlot) {
		w := agent.GetCharacter().AutoAttacks.MH()
		if slot == proto.ItemSlot_ItemSlotOffHand {
			w = agent.GetCharacter().AutoAttacks.OH()
		}
		w.BaseDamageMin += 2
		w.BaseDamageMax += 2
	})

	// Weapon - Beast Slaying
	core.AddWeaponEffect(249, func(agent core.Agent, slot proto.ItemSlot) {
		character := agent.GetCharacter()

		if character.CurrentTarget.MobType == proto.MobType_MobTypeBeast {
			w := character.AutoAttacks.MH()
			if slot == proto.ItemSlot_ItemSlotOffHand {
				w = character.AutoAttacks.OH()
			}

			w.BaseDamageMin += 2
			w.BaseDamageMax += 2

			w = character.AutoAttacks.Ranged()
			w.BaseDamageMin += 2
			w.BaseDamageMax += 2
		}
	})

	// Weapon - Minor Striking
	core.AddWeaponEffect(250, func(agent core.Agent, slot proto.ItemSlot) {
		w := agent.GetCharacter().AutoAttacks.MH()
		if slot == proto.ItemSlot_ItemSlotOffHand {
			w = agent.GetCharacter().AutoAttacks.OH()
		}
		w.BaseDamageMin += 1
		w.BaseDamageMax += 1
	})

	// Deadly Scope
	core.AddWeaponEffect(663, func(agent core.Agent, _ proto.ItemSlot) {
		w := agent.GetCharacter().AutoAttacks.Ranged()
		w.BaseDamageMin += 5
		w.BaseDamageMax += 5
	})

	// Sniper Scope
	core.AddWeaponEffect(664, func(agent core.Agent, _ proto.ItemSlot) {
		w := agent.GetCharacter().AutoAttacks.Ranged()
		w.BaseDamageMin += 7
		w.BaseDamageMax += 7
	})

	// Weapon - Fiery Weapon
	core.AddWeaponEffect(803, func(agent core.Agent, _ proto.ItemSlot) {
		character := agent.GetCharacter()

		procMask := character.GetProcMaskForEnchant(803)
		ppmm := character.AutoAttacks.NewPPMManager(6.0, procMask)

		procMaskOnAuto := core.ProcMaskDamageProc     // Both spell and melee proc combo
		procMaskOnSpecial := core.ProcMaskSpellDamage // TODO: check if core.ProcMaskSpellDamage remains on special

		procSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 13897},
			SpellSchool: core.SpellSchoolFire,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    procMaskOnAuto,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, 40, spell.OutcomeMagicHitAndCrit)
			},
		})

		aura := core.MakePermanent(character.GetOrRegisterAura(core.Aura{
			Label: "Fiery Weapon",
			OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !result.Landed() || spell.Flags.Matches(core.SpellFlagSuppressWeaponProcs) {
					return
				}
				for i := core.WeaponProcRolls(spell); i > 0; i-- {
					if ppmm.Proc(sim, spell.ProcMask, "Fiery Weapon") {
						if spell.ProcMask.Matches(core.ProcMaskMeleeSpecial) {
							procSpell.ProcMask = procMaskOnSpecial
						} else {
							procSpell.ProcMask = procMaskOnAuto
						}
						procSpell.Cast(sim, result.Target)
					}
				}
			},
		}))

		character.ItemSwap.RegisterOnSwapItemForEffectWithPPMManager(803, 6.0, &ppmm, aura)
	})

	// Weapon - Greater Striking
	core.AddWeaponEffect(805, func(agent core.Agent, slot proto.ItemSlot) {
		w := agent.GetCharacter().AutoAttacks.MH()
		if slot == proto.ItemSlot_ItemSlotOffHand {
			w = agent.GetCharacter().AutoAttacks.OH()
		}
		w.BaseDamageMin += 4
		w.BaseDamageMax += 4
	})

	// Weapon - Lesser Beastslayer
	core.AddWeaponEffect(853, func(agent core.Agent, slot proto.ItemSlot) {
		character := agent.GetCharacter()

		if character.CurrentTarget.MobType == proto.MobType_MobTypeBeast {
			w := character.AutoAttacks.MH()
			if slot == proto.ItemSlot_ItemSlotOffHand {
				w = character.AutoAttacks.OH()
			}

			w.BaseDamageMin += 6
			w.BaseDamageMax += 6

			w = character.AutoAttacks.Ranged()
			w.BaseDamageMin += 6
			w.BaseDamageMax += 6
		}
	})

	// Weapon - Lesser Elemental Slayer
	core.AddWeaponEffect(854, func(agent core.Agent, slot proto.ItemSlot) {
		character := agent.GetCharacter()

		if character.CurrentTarget.MobType == proto.MobType_MobTypeElemental {
			w := character.AutoAttacks.MH()
			if slot == proto.ItemSlot_ItemSlotOffHand {
				w = character.AutoAttacks.OH()
			}

			w.BaseDamageMin += 6
			w.BaseDamageMax += 6

			w = character.AutoAttacks.Ranged()
			w.BaseDamageMin += 6
			w.BaseDamageMax += 6
		}
	})

	// Boots - Minor Speed
	core.NewEnchantEffect(911, func(agent core.Agent) {
		character := agent.GetCharacter()

		character.RegisterAura(core.Aura{
			Label: "Minor Speed",
			OnInit: func(aura *core.Aura, sim *core.Simulation) {
				character.AddMoveSpeedModifier(&core.ActionID{SpellID: 13889}, 1.08)
			},
		})
	})

	// Weapon - Striking
	core.AddWeaponEffect(943, func(agent core.Agent, slot proto.ItemSlot) {
		w := agent.GetCharacter().AutoAttacks.MH()
		if slot == proto.ItemSlot_ItemSlotOffHand {
			w = agent.GetCharacter().AutoAttacks.OH()
		}
		w.BaseDamageMin += 3
		w.BaseDamageMax += 3
	})

	// Forever's 2H Minor Impact, Lesser Impact and Impact: +4, +5 and +6 damage, up from 1.12's
	// +2, +3 and +5. They have effect IDs of their own, see tools/database/forever_enchants.go.
	//
	// Greater Impact and Superior Impact keep 1.12's +7 and +9 (hyjal.cc recipe pages) and its
	// effect IDs, 963 and 1896.
	for effectID, damage := range map[int32]float64{7745: 4, 13529: 5, 13695: 6, 963: 7, 1896: 9} {
		core.AddWeaponEffect(effectID, func(agent core.Agent, slot proto.ItemSlot) {
			w := agent.GetCharacter().AutoAttacks.MH()
			w.BaseDamageMin += damage
			w.BaseDamageMax += damage
		})
	}

	// Forever's Weapon - Revelation (Enchanting 140): a direct spell hit that lands but does not
	// crit has a chance to give Revelation, whose one charge gives the next spell the wearer casts
	// +100% crit chance for 15 sec.
	//
	// Which hits roll it (wowsims/forever EV-0149, R-0051, on 1,356 Warcraft Logs reports): every
	// landed direct hit of a magic spell that deals damage, the wearer's own or an item's, each
	// target of an area spell on its own, and each hit of a channel that casts a spell a tick
	// (Arcane Missiles' missiles, Blizzard, Rain of Fire, Hellfire, Hurricane: their area ticks
	// are hits in game). Periodic ticks, damage shields and melee or ranged attacks (the paladin's
	// judgements) roll nothing. The user saw an enhancement shaman's Windfury and Flametongue
	// Weapon procs, its totems and Fire Nova neither roll nor spend it (2026-10-05), so the sim
	// keeps the shaman's imbue procs and the spells that go off from a totem out, though EV-0149
	// counts Fire Nova at 7.6% (left to the user).
	//
	// What spends the charge (EV-0028, EV-0149, R-0048): the cast of a spell it reaches, as the
	// cast completes, so every target an area spell hits crits and a spell in flight keeps its
	// crit; a channel keeps the charge to its end and every hit of it crits meanwhile, and a charge
	// gained during a channel is kept to that channel's end too. Until 2026-11-25 the first target
	// an area spell hit spent the charge, and channels neither rolled nor spent it.
	core.AddWeaponEffect(RevelationEffectID, func(agent core.Agent, _ proto.ItemSlot) {
		character := agent.GetCharacter()
		if character.HasAura("Revelation Weapon") {
			return
		}
		chance, ok := RevelationClassChance[character.Class]
		if RevelationProcChance >= 0 {
			chance, ok = RevelationProcChance, true
		}
		if !ok || chance == 0 {
			return
		}

		// The spells the buff reaches: the wearer's own damage spells and a channel's missiles,
		// not item procs, imbue procs, totems or spells that only tick.
		buffed := func(spell *core.Spell) bool {
			if !spell.ProcMask.Matches(core.ProcMaskSpellDamage) || spell.ProcMask.Matches(core.ProcMaskMeleeOrRanged) {
				return false
			}
			if spell.Flags.Matches(core.SpellFlagPassiveSpell | core.SpellFlagCastFromTotem) {
				return false
			}
			return spell.Flags.Matches(core.SpellFlagChanneled) || !spell.Flags.Matches(core.SpellFlagPureDot)
		}
		// The hits that roll it.
		rolls := func(spell *core.Spell, result *core.SpellResult) bool {
			if !result.Landed() || result.DidCrit() || result.Damage <= 0 {
				return false
			}
			if spell.ProcMask.Matches(core.ProcMaskMeleeOrRanged | core.ProcMaskSpellHealing) {
				return false
			}
			// Flametongue Weapon's and Flametongue Totem's procs (the imbues), not the item procs
			// that carry the damage proc mask beside the spell proc one (Thunderfury, Dragonbreath
			// Chili).
			if spell.ProcMask.Matches(core.ProcMaskDamageProc) && !spell.ProcMask.Matches(core.ProcMaskSpellProc) {
				return false
			}
			if spell.Flags.Matches(core.SpellFlagCastFromTotem | core.SpellFlagNoPartialResist | core.SpellFlagHelpful) {
				return false
			}
			return spell.ProcMask.Matches(core.ProcMaskSpellDamage) || spell.DefenseType == core.DefenseTypeMagic
		}
		learnLevel := func(spell *core.Spell) int {
			return core.Ternary(spell.Rank > 0, spell.RequiredLevel, 0)
		}
		roll := func(sim *core.Simulation, spell *core.Spell) bool {
			return sim.Proc(chance*revelationLevelShare(character.Level, learnLevel(spell)), "Revelation")
		}

		var buff *core.Aura
		// The channel holding the charge, which goes as it ends, clipped or not.
		var heldBy *core.Dot
		character.OnSpellRegistered(func(spell *core.Spell) {
			if !spell.Flags.Matches(core.SpellFlagChanneled) || !buffed(spell) {
				return
			}
			for _, dot := range append([]*core.Dot{spell.AOEDot()}, spell.Dots()...) {
				if dot == nil {
					continue
				}
				dot.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
					if heldBy == dot {
						heldBy = nil
						buff.Deactivate(sim)
					}
				})
			}
		})
		runningChannel := func() *core.Dot {
			if dot := character.ChanneledDot; dot != nil && dot.IsActive() && buffed(dot.Spell) {
				return dot
			}
			return nil
		}

		buff = character.RegisterAura(core.Aura{
			Label:    "Revelation",
			ActionID: core.ActionID{SpellID: RevelationEffectID},
			Duration: time.Second * 15,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				for _, spell := range character.Spellbook {
					if buffed(spell) {
						spell.BonusCritRating += 100 * core.SpellCritRatingPerCritChance
					}
				}
				// R-0048: a charge gained during a channel the buff reaches goes at its end.
				heldBy = runningChannel()
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				heldBy = nil
				for _, spell := range character.Spellbook {
					if buffed(spell) {
						spell.BonusCritRating -= 100 * core.SpellCritRatingPerCritChance
					}
				}
			},
			OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
				// A hit of this cast that gave the charge doesn't spend it, nor does a cast while a
				// channel holds it.
				if !spell.Flags.Matches(core.SpellFlagAPL) || !buffed(spell) || aura.StartedAt() == sim.CurrentTime || heldBy != nil {
					return
				}
				if spell.Flags.Matches(core.SpellFlagChanneled) {
					if dot := runningChannel(); dot != nil && dot.Spell == spell {
						heldBy = dot
						return
					}
				}
				aura.Deactivate(sim)
			},
		})

		core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Revelation Weapon",
			OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if rolls(spell, result) && roll(sim, spell) {
					buff.Activate(sim)
				}
			},
			// An area channel's ticks are hits of the spell it casts each tick.
			OnPeriodicDamageDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if spell.Flags.Matches(core.SpellFlagChanneled) && spell.AOEDot() != nil && rolls(spell, result) && roll(sim, spell) {
					buff.Activate(sim)
				}
			},
		}))
	})

	// Weapon - Superior Striking
	core.AddWeaponEffect(1897, func(agent core.Agent, slot proto.ItemSlot) {
		w := agent.GetCharacter().AutoAttacks.MH()
		if slot == proto.ItemSlot_ItemSlotOffHand {
			w = agent.GetCharacter().AutoAttacks.OH()
		}
		w.BaseDamageMin += 5
		w.BaseDamageMax += 5
	})

	// Weapon - Lifestealing
	core.AddWeaponEffect(1898, func(agent core.Agent, slot proto.ItemSlot) {
		character := agent.GetCharacter()

		procMask := character.GetProcMaskForEnchant(1898)
		ppmm := character.AutoAttacks.NewPPMManager(6.66, procMask)

		procMaskOnAuto := core.ProcMaskDamageProc     // Both spell and melee proc combo
		procMaskOnSpecial := core.ProcMaskSpellDamage // TODO: check if core.ProcMaskSpellDamage remains on special

		procSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 20004},
			SpellSchool: core.SpellSchoolShadow,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    procMaskOnAuto,
			Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, 30, spell.OutcomeMagicHitAndCrit)
			},
		})

		aura := core.MakePermanent(character.GetOrRegisterAura(core.Aura{
			Label: "Lifestealing Weapon",
			OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if result.Landed() && !spell.Flags.Matches(core.SpellFlagSuppressWeaponProcs) && ppmm.Proc(sim, spell.ProcMask, "Lifestealing Weapon") {
					if spell.ProcMask.Matches(core.ProcMaskMeleeSpecial) {
						procSpell.ProcMask = procMaskOnSpecial
					} else {
						procSpell.ProcMask = procMaskOnAuto
					}
					procSpell.Cast(sim, result.Target)
				}
			},
		}))

		character.ItemSwap.RegisterOnSwapItemForEffectWithPPMManager(1898, 6.66, &ppmm, aura)
	})

	// TODO: Crusader, Mongoose, and Executioner could also be modelled as AddWeaponEffect instead
	// ApplyCrusaderEffect will be applied twice if there is two weapons with this enchant.
	//   However, it will automatically overwrite one of them, so it should be ok.
	//   A single application of the aura will handle both mh and oh procs.
	core.NewEnchantEffect(1900, func(agent core.Agent) {
		character := agent.GetCharacter()

		procMask := character.GetProcMaskForEnchant(1900)
		ppmm := character.AutoAttacks.NewPPMManager(1.0, procMask)

		// We give 100 Strength at level 60 and below, and 4 less a level over 60.
		//
		// Below 60 the old formula grew instead, so a level 30 proc gave 220 Strength. The
		// Forever client's Holy Strength (20007, build 70291) is 100 with nothing per level, and
		// level 19 twinks got the full 100 in 2019 Classic.
		strBonus := 100.0 - 4.0*float64(max(character.Level-60, 0))
		mhAura := character.NewTemporaryStatsAura("Crusader Enchant MH", core.ActionID{SpellID: 20007, Tag: 1}, stats.Stats{stats.Strength: strBonus}, time.Second*15)
		ohAura := character.NewTemporaryStatsAura("Crusader Enchant OH", core.ActionID{SpellID: 20007, Tag: 2}, stats.Stats{stats.Strength: strBonus}, time.Second*15)

		aura := character.GetOrRegisterAura(core.Aura{
			Label:    "Crusader Enchant",
			Duration: core.NeverExpires,
			OnReset: func(aura *core.Aura, sim *core.Simulation) {
				aura.Activate(sim)
			},
			OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !result.Landed() || spell.Flags.Matches(core.SpellFlagSuppressWeaponProcs) {
					return
				}
				for i := core.WeaponProcRolls(spell); i > 0; i-- {
					if ppmm.Proc(sim, spell.ProcMask, "Crusader") {
						if spell.IsMH() {
							mhAura.Activate(sim)
						} else {
							ohAura.Activate(sim)
						}
					}
				}
			},
		})

		character.ItemSwap.RegisterOnSwapItemForEffectWithPPMManager(1900, 1.0, &ppmm, aura)
	})

	// Biznicks 247x128 Accurascope
	core.AddWeaponEffect(2523, func(agent core.Agent, _ proto.ItemSlot) {
		character := agent.GetCharacter()
		character.AddBonusRangedHitRating(3)
	})

	// Gloves - Threat
	core.NewEnchantEffect(2613, func(agent core.Agent) {
		character := agent.GetCharacter()

		character.RegisterAura(core.Aura{
			Label: "Threat +2%",
			OnReset: func(aura *core.Aura, sim *core.Simulation) {
				character.PseudoStats.ThreatMultiplier *= 1.02
			},
		})
	})

	// Cloak - Subtlety
	core.NewEnchantEffect(2621, func(agent core.Agent) {
		character := agent.GetCharacter()

		character.RegisterAura(core.Aura{
			Label: "Subtlety",
			OnReset: func(aura *core.Aura, sim *core.Simulation) {
				character.PseudoStats.ThreatMultiplier /= 1.02
			},
		})
	})

	core.AddEffectsToTest = true
}
