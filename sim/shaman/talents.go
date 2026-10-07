package shaman

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

func (shaman *Shaman) ApplyTalents() {
	// Elemental Talents
	shaman.applyConcussion()
	shaman.applyElementalWarding()
	shaman.applyCallOfFlame()
	shaman.applyElementalDevastation()
	shaman.applyElementalFocus()
	shaman.applyElementalFury()

	// Enhancement Talents
	shaman.applyFlurry()
	shaman.applyImprovedStormstrike()
	shaman.applyMaelstromWeapon()
	shaman.registerRageOfTheFarseerCD()

	if shaman.Talents.AncestralKnowledge > 0 {
		shaman.MultiplyStat(stats.Intellect, 1+.02*float64(shaman.Talents.AncestralKnowledge))
	}

	if shaman.Talents.MentalDexterity > 0 {
		shaman.AddStatDependency(stats.Intellect, stats.AttackPower, []float64{0, .33, .67, 1}[shaman.Talents.MentalDexterity])
	}

	if shaman.Talents.MentalQuickness > 0 {
		shaman.AddStatDependency(stats.Intellect, stats.SpellPower, .15*float64(shaman.Talents.MentalQuickness))
	}

	shaman.AddStat(stats.MeleeCrit, core.CritRatingPerCritChance*float64(shaman.Talents.ThunderingStrikes))
	shaman.AddStat(stats.SpellCrit, core.SpellCritRatingPerCritChance*float64(shaman.Talents.ThunderingStrikes))

	shaman.AddStat(stats.Dodge, core.DodgeRatingPerDodgeChance*2*float64(shaman.Talents.Anticipation))

	if shaman.Talents.Toughness > 0 {
		shaman.MultiplyStat(stats.Stamina, 1+.02*float64(shaman.Talents.Toughness))
	}

	if shaman.Talents.SpiritWeapons {
		shaman.PseudoStats.CanParry = true

		// Rockbiter is the tanking imbue, so the threat modifier flips depending on which imbue is up.
		if shaman.Consumes.MainHandImbue == proto.WeaponImbue_RockbiterWeapon {
			shaman.PseudoStats.ThreatMultiplier *= 1.30
		} else {
			shaman.PseudoStats.ThreatMultiplier *= 0.70
		}
	}

	// Restoration Talents
	// TODO: Healing Way
	// TODO: Ancestral Healing
	shaman.registerNaturesSwiftnessCD()

	shaman.PseudoStats.SpiritRegenRateCasting += []float64{0, .17, .33, .50}[shaman.Talents.Mindfulness]

	if shaman.Talents.ImprovedReincarnation > 0 {
		shaman.MultiplyStat(stats.Health, 1+.02*float64(shaman.Talents.ImprovedReincarnation))
	}

	if shaman.Talents.TidalFocus > 0 {
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagShaman) && spell.ProcMask.Matches(core.ProcMaskSpellHealing) && spell.Cost != nil {
				spell.Cost.Multiplier -= shaman.Talents.TidalFocus
			}
		})

		shaman.AddStat(stats.MeleeHit, core.MeleeHitRatingPerHitChance*float64(shaman.Talents.TidalFocus))
		shaman.AddStat(stats.SpellHit, core.SpellHitRatingPerHitChance*float64(shaman.Talents.TidalFocus))
	}

	if shaman.Talents.NaturalGrace > 0 {
		threatMultiplier := 1 - .05*float64(shaman.Talents.NaturalGrace)
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagShaman) {
				spell.ThreatMultiplier *= threatMultiplier
			}
		})
	}

	if shaman.Talents.TidalMastery > 0 {
		critBonus := float64(shaman.Talents.TidalMastery) * core.SpellCritRatingPerCritChance
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagShaman) && spell.ProcMask.Matches(core.ProcMaskSpellHealing) {
				spell.BonusCritRating += critBonus
			}
		})
	}
}

func (shaman *Shaman) applyConcussion() {
	if shaman.Talents.Concussion == 0 {
		return
	}

	additiveMultiplier := 0.01 * float64(shaman.Talents.Concussion)
	affectedSpellCodes := []int32{SpellCode_ShamanLightningBolt, SpellCode_ShamanChainLightning, SpellCode_ShamanEarthShock}

	shaman.OnSpellRegistered(func(spell *core.Spell) {
		if slices.Contains(affectedSpellCodes, spell.SpellCode) {
			spell.DamageMultiplierAdditive += additiveMultiplier
		}
	})
}

func (shaman *Shaman) applyElementalWarding() {
	if shaman.Talents.ElementalWarding == 0 {
		return
	}

	// 3 / 7 / 10% per the beta client.
	multiplier := 1 - []float64{0, .03, .07, .10}[shaman.Talents.ElementalWarding]
	for _, school := range []stats.SchoolIndex{stats.SchoolIndexFire, stats.SchoolIndexFrost, stats.SchoolIndexNature} {
		shaman.PseudoStats.SchoolDamageTakenMultiplier[school] *= multiplier
	}
}

// The fire totems apply this themselves so that it lands on their base damage.
func (shaman *Shaman) callOfFlameMultiplier() float64 {
	return 1 + .05*float64(shaman.Talents.CallOfFlame)
}

func (shaman *Shaman) applyCallOfFlame() {
	if shaman.Talents.CallOfFlame == 0 {
		return
	}

	affectedSpellCodes := []int32{SpellCode_ShamanFlameShock, SpellCode_ShamanLavaBurst}
	multiplier := shaman.callOfFlameMultiplier()

	shaman.OnSpellRegistered(func(spell *core.Spell) {
		if slices.Contains(affectedSpellCodes, spell.SpellCode) {
			spell.DamageMultiplier *= multiplier
		}
	})
}

func (shaman *Shaman) shamanisticFocusReduction() int32 {
	return core.TernaryInt32(shaman.Talents.ShamanisticFocus, 45, 0)
}

// 0.17 / 0.33 / 0.50 sec per the beta client.
func (shaman *Shaman) elementalAlacrityReduction() time.Duration {
	return []time.Duration{0, 170, 330, 500}[shaman.Talents.ElementalAlacrity] * time.Millisecond
}

func (shaman *Shaman) improvedFireNovaBonus() float64 {
	return .1 * float64(shaman.Talents.ImprovedFireNova)
}

func (shaman *Shaman) improvedFireNovaCooldownReduction() time.Duration {
	return time.Second * 2 * time.Duration(shaman.Talents.ImprovedFireNova)
}

func (shaman *Shaman) applyElementalFocus() {
	if !shaman.Talents.ElementalFocus {
		return
	}

	var affectedSpells []*core.Spell

	shaman.ClearcastingAura = shaman.RegisterAura(core.Aura{
		Label:     "Clearcasting",
		ActionID:  core.ActionID{SpellID: 16246},
		Duration:  time.Second * 15,
		MaxStacks: 1,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells = shaman.getClearcastingSpells()
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) {
				if spell.Cost != nil {
					spell.Cost.Multiplier -= 100
				}
			})
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) {
				if spell.Cost != nil {
					spell.Cost.Multiplier += 100
				}
			})
		},
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
			if newStacks == 0 {
				aura.Deactivate(sim)
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			// OnCastComplete is called after OnSpellHitDealt / etc, so don't deactivate if it was just activated.
			if aura.RemainingDuration(sim) == aura.Duration {
				return
			}

			if aura.GetStacks() > 0 && shaman.isShamanDamagingSpell(spell) {
				aura.RemoveStack(sim)
			}
		},
	})

	core.MakePermanent(shaman.RegisterAura(core.Aura{
		Label: "Elemental Focus Trigger",
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if shaman.isShamanDamagingSpell(spell) && sim.Proc(0.10, "Elemental Focus") {
				shaman.ClearcastingAura.Activate(sim)
				shaman.ClearcastingAura.SetStacks(sim, shaman.ClearcastingAura.MaxStacks)
			}
		},
	}))
}

func (shaman *Shaman) isShamanDamagingSpell(spell *core.Spell) bool {
	return spell.Flags.Matches(SpellFlagShaman) && spell.ProcMask.Matches(core.ProcMaskSpellDamage)
}

func (shaman *Shaman) getClearcastingSpells() []*core.Spell {
	return core.FilterSlice(
		shaman.Spellbook,
		func(spell *core.Spell) bool {
			return spell != nil && shaman.isShamanDamagingSpell(spell)
		},
	)
}

func (shaman *Shaman) applyElementalDevastation() {
	if shaman.Talents.ElementalDevastation == 0 {
		return
	}

	spellID := []int32{0, 30165, 29177, 29178}[shaman.Talents.ElementalDevastation]
	critBonus := 3.0 * float64(shaman.Talents.ElementalDevastation) * core.CritRatingPerCritChance
	procAura := shaman.NewTemporaryStatsAura("Elemental Devastation Proc", core.ActionID{SpellID: spellID}, stats.Stats{stats.MeleeCrit: critBonus}, time.Second*10)

	shaman.RegisterAura(core.Aura{
		Label:    "Elemental Devastation",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ProcMask.Matches(core.ProcMaskSpellDamage) && result.Outcome.Matches(core.OutcomeCrit) {
				procAura.Activate(sim)
			}
		},
	})
}
func (shaman *Shaman) applyElementalFury() {
	if shaman.Talents.ElementalFury == 0 {
		return
	}

	// 20% per rank, per the beta client.
	critDamageBonus := .2 * float64(shaman.Talents.ElementalFury)
	affectedSpellCodes := []int32{SpellCode_ShamanSearingTotem, SpellCode_ShamanMagmaTotem}

	shaman.OnSpellRegistered(func(spell *core.Spell) {
		if spell.DefenseType != core.DefenseTypeMagic {
			return
		}

		isElementalSpell := spell.Flags.Matches(SpellFlagShaman) && spell.SpellSchool.Matches(core.SpellSchoolFire|core.SpellSchoolFrost|core.SpellSchoolNature)
		if isElementalSpell || slices.Contains(affectedSpellCodes, spell.SpellCode) {
			spell.CritDamageBonus += critDamageBonus
		}
	})
}

func (shaman *Shaman) registerNaturesSwiftnessCD() {
	if !shaman.Talents.NaturesSwiftness {
		return
	}
	actionID := core.ActionID{SpellID: 16188}
	cdTimer := shaman.NewTimer()
	cd := time.Minute * 3

	var affectedSpells []*core.Spell

	nsAura := shaman.RegisterAura(core.Aura{
		Label:    "Natures Swiftness",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells = core.FilterSlice(
				shaman.Spellbook,
				func(spell *core.Spell) bool {
					return spell != nil && spell.SpellSchool.Matches(core.SpellSchoolNature) && spell.DefaultCast.CastTime > 0
				},
			)
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) { spell.CastTimeMultiplier -= 1 })
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) { spell.CastTimeMultiplier += 1 })
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.SpellSchool.Matches(core.SpellSchoolNature) && spell.DefaultCast.CastTime > 0 {
				// Remove the buff and put skill on CD
				aura.Deactivate(sim)
				cdTimer.Set(sim.CurrentTime + cd)
				shaman.UpdateMajorCooldowns()
			}
		},
	})

	nsSpell := shaman.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: cd,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			// Don't use NS unless we're casting a full-length lightning bolt, which is
			// the only spell shamans have with a cast longer than GCD.
			return !shaman.HasTemporarySpellCastSpeedIncrease()
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			nsAura.Activate(sim)
		},
	})

	shaman.AddMajorCooldown(core.MajorCooldown{
		Spell: nsSpell,
		Type:  core.CooldownTypeDPS,
	})
}

func (shaman *Shaman) applyFlurry() {
	if shaman.Talents.Flurry == 0 {
		return
	}

	talentAura := shaman.makeFlurryAura(shaman.Talents.Flurry)

	// This must be registered before the below trigger because in-game a crit weapon swing consumes a stack before the refresh, so you end up with:
	// 3 => 2
	// refresh
	// 2 => 3
	shaman.makeFlurryConsumptionTrigger(talentAura)

	shaman.RegisterAura(core.Aura{
		Label:    "Flurry Proc Trigger",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ProcMask.Matches(core.ProcMaskMelee) && result.Outcome.Matches(core.OutcomeCrit) {
				talentAura.Activate(sim)
				if talentAura.IsActive() {
					talentAura.SetStacks(sim, 3)
				}
				return
			}
		},
	})
}

// These are separated out because of the T1 Shaman Tank 2P that can proc Flurry separately from the talent.
// It triggers the max-rank Flurry aura but with dodge, parry, or block.
func (shaman *Shaman) makeFlurryAura(points int32) *core.Aura {
	if points == 0 {
		return nil
	}

	spellID := []int32{16257, 16277, 16278, 16279, 16280}[points-1]
	attackSpeed := []float64{1.05, 1.1, 1.15, 1.2, 1.25}[points-1]

	aura := shaman.GetOrRegisterAura(core.Aura{
		Label:     fmt.Sprintf("Flurry Proc (%d)", spellID),
		ActionID:  core.ActionID{SpellID: spellID},
		Duration:  core.NeverExpires,
		MaxStacks: 3,
	})

	aura.NewExclusiveEffect("Flurry", true, core.ExclusiveEffect{
		Priority: attackSpeed,
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			shaman.MultiplyMeleeSpeed(sim, attackSpeed)
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			shaman.MultiplyMeleeSpeed(sim, 1/(attackSpeed))
		},
	})

	return aura
}

// With the Warden T1 2pc it's possible to have 2 different Flurry auras if using less than 5/5 points in Flurry.
// The two different buffs don't stack whatsoever. Instead the stronger aura takes precedence and each one is only refreshed by the corresponding triggers.
func (shaman *Shaman) makeFlurryConsumptionTrigger(flurryAura *core.Aura) *core.Aura {
	icd := core.Cooldown{
		Timer:    shaman.NewTimer(),
		Duration: time.Millisecond * 500,
	}
	return core.MakePermanent(shaman.GetOrRegisterAura(core.Aura{
		Label: fmt.Sprintf("Flurry Consume Trigger - %d", flurryAura.ActionID.SpellID),
		OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			// Remove a stack.
			if flurryAura.IsActive() && spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) && icd.IsReady(sim) {
				icd.Use(sim)
				flurryAura.RemoveStack(sim)
			}
		},
	}))
}

func (shaman *Shaman) applyImprovedStormstrike() {
	if !shaman.Talents.Stormstrike || shaman.Talents.ImprovedStormstrike == 0 {
		return
	}

	// The proc and reset chances grow with points but the regen does not. Both ranks
	// give 50% mana regeneration while casting.
	points := float64(shaman.Talents.ImprovedStormstrike)
	procChance := .5 * points
	resetChance := .5 * points
	regenRate := .5

	focusAura := shaman.RegisterAura(core.Aura{
		Label:    "Improved Stormstrike",
		ActionID: core.ActionID{SpellID: 51521},
		Duration: time.Second * 15,
		// The mana tick amount is cached, so the rate change has to be pushed into it
		// the way Innervate does. Without this the talent never regenerated anything.
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			shaman.PseudoStats.SpiritRegenRateCasting += regenRate
			shaman.UpdateManaRegenRates()
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			shaman.PseudoStats.SpiritRegenRateCasting -= regenRate
			shaman.UpdateManaRegenRates()
		},
	})

	core.MakePermanent(shaman.RegisterAura(core.Aura{
		Label: "Improved Stormstrike Trigger",
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.SpellCode == SpellCode_ShamanStormstrike && sim.Proc(procChance, "Improved Stormstrike") {
				focusAura.Activate(sim)
			}
		},
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Outcome.Matches(core.OutcomeDodge|core.OutcomeParry) && sim.Proc(resetChance, "Improved Stormstrike Reset") {
				shaman.Stormstrike.CD.Reset()
			}
		},
	}))
}

func (shaman *Shaman) applyMaelstromWeapon() {
	if shaman.Talents.MaelstromWeapon == 0 {
		return
	}

	// TODO: The tooltip has no proc rate, beta testing will have to measure it.
	// 2 PPM per point puts 5/5 at a full stack roughly every 30 sec.
	ppmm := shaman.AutoAttacks.NewPPMManager(2*float64(shaman.Talents.MaelstromWeapon), core.ProcMaskMelee)

	castTimeReductionPerStack := .04 * float64(shaman.Talents.MaelstromWeapon)
	costReductionPerStack := 4 * shaman.Talents.MaelstromWeapon

	shaman.MaelstromWeaponAura = shaman.RegisterAura(core.Aura{
		Label:     "Maelstrom Weapon",
		ActionID:  core.ActionID{SpellID: 51530},
		Duration:  time.Second * 30,
		MaxStacks: 5,
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
			stacks := newStacks - oldStacks
			for _, spell := range shaman.LightningBolt {
				if spell == nil || spell.Cost == nil {
					continue
				}

				spell.CastTimeMultiplier -= castTimeReductionPerStack * float64(stacks)
				spell.Cost.Multiplier -= costReductionPerStack * stacks
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.SpellCode == SpellCode_ShamanLightningBolt {
				aura.Deactivate(sim)
			}
		},
	})

	core.MakePermanent(shaman.RegisterAura(core.Aura{
		Label: "Maelstrom Weapon Trigger",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() {
				return
			}
			procced := ppmm.Proc(sim, spell.ProcMask, "Maelstrom Weapon")
			if spell.SpellCode == SpellCode_ShamanLightningBolt && shaman.lightningBoltTriggersMaelstrom {
				// We take the main hand swing chance, so the bolt procs half as often as a white hit would.
				procced = sim.RandomFloat("Maelstrom Weapon") < 0.5*ppmm.Chance(core.ProcMaskMeleeMHAuto)
			}
			if procced {
				shaman.MaelstromWeaponAura.Activate(sim)
				shaman.MaelstromWeaponAura.AddStack(sim)
			}
		},
	}))
}

func (shaman *Shaman) registerRageOfTheFarseerCD() {
	if !shaman.Talents.RageOfTheFarseer {
		return
	}

	actionID := core.ActionID{SpellID: 2825}
	multiplier := 1.30
	// TODO: The tooltip showed no cooldown, beta will confirm it. 3 minutes matches the other class cooldowns of this size.
	cd := time.Minute * 3

	buffAura := shaman.RegisterAura(core.Aura{
		Label:    "Rage of the Farseer",
		ActionID: actionID,
		Duration: time.Second * 25,
		// Attack speed only. The 2026-09-24 beta build took the cast speed half away.
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			shaman.MultiplyMeleeSpeed(sim, multiplier)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			shaman.MultiplyMeleeSpeed(sim, 1/multiplier)
		},
	})

	rageSpell := shaman.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: cd,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			buffAura.Activate(sim)
		},
	})

	shaman.AddMajorCooldown(core.MajorCooldown{
		Spell: rageSpell,
		Type:  core.CooldownTypeDPS,
	})
}

func (shaman *Shaman) totemManaMultiplier() int32 {
	return 100 - 5*shaman.Talents.TotemicFocus
}

// Restorative Totems uses Mod Spell Effectiveness (Base Value). Only the Healing Stream half is
// modelled here, Mana Spring's value is hard coded in buffs.go.
func (shaman *Shaman) restorativeTotemsModifier() float64 {
	return .1 * float64(shaman.Talents.RestorativeTotems)
}

// Purification uses Mod Spell Effectiveness (Base Healing)
func (shaman *Shaman) purificationHealingModifier() float64 {
	return .02 * float64(shaman.Talents.Purification)
}
