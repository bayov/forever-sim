package warrior

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

func (warrior *Warrior) ApplyTalents() {
	warrior.AddStat(stats.MeleeCrit, core.CritRatingPerCritChance*1*float64(warrior.Talents.Cruelty))
	warrior.AddStat(stats.Defense, 4*float64(warrior.Talents.Anticipation))
	warrior.AddStat(stats.Parry, 1*float64(warrior.Talents.Deflection))

	warrior.applyBastion()
	warrior.applyFocusedRage()
	warrior.applyAngerManagement()
	warrior.applyDeepWounds()
	warrior.applyTwoHandedWeaponSpecialization()
	warrior.applyWeaponmaster()
	warrior.applyBloodthrill()
	warrior.applyUnbridledWrath()
	warrior.applyDualWieldSpecialization()
	warrior.applyFuriousPrecision()
	warrior.applyEnrage()
	warrior.applyFlurry()
	warrior.applyShieldSpecialization()
	warrior.applyMasterOfDefense()
	warrior.registerDeathWishCD()
	warrior.registerSweepingStrikesCD()
	warrior.registerLastStandCD()
}

func (warrior *Warrior) applyAngerManagement() {
	if !warrior.Talents.AngerManagement {
		return
	}

	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: 12296})

	warrior.RegisterResetEffect(func(sim *core.Simulation) {
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period: time.Second * 3,
			OnAction: func(sim *core.Simulation) {
				warrior.AddRage(sim, 1, rageMetrics)
				warrior.LastAMTick = sim.CurrentTime
			},
		})
	})
}

func (warrior *Warrior) applyTwoHandedWeaponSpecialization() {
	if warrior.Talents.TwoHandedWeaponSpecialization == 0 || warrior.MainHand().HandType != proto.HandType_HandTypeTwoHand {
		return
	}

	multiplier := 1 + 0.01*float64(warrior.Talents.TwoHandedWeaponSpecialization)
	warrior.OnSpellRegistered(func(spell *core.Spell) {
		if spell.BonusCoefficient > 0 {
			spell.DamageMultiplier *= multiplier
		}
	})
}

// Weaponmaster folds the four Classic weapon specialization talents into one talent that pays
// out differently depending on what is equipped: 1% crit, 3% armor ignored or a 1% extra attack
// chance per point (client 1.60.1.70338, TraitDefinitionEffectPoints curves of spell 1290261).
func (warrior *Warrior) applyWeaponmaster() {
	points := warrior.Talents.Weaponmaster
	if points == 0 {
		return
	}

	if mask := warrior.GetProcMaskForTypes(proto.WeaponType_WeaponTypeSword); mask != core.ProcMaskUnknown {
		warrior.registerWeaponmasterExtraAttack(mask, 0.01*float64(points))
	}

	if mask := warrior.GetProcMaskForTypes(proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeStaff); mask != core.ProcMaskUnknown {
		// Armor ignore is a character wide modifier here, so an off-hand mace also discounts
		// armor for main hand attacks.
		warrior.PseudoStats.ArmorIgnorePercent += 0.03 * float64(points)
	}

	// the default character panel displays critical strike chance for main hand only
	switch warrior.GetProcMaskForTypes(proto.WeaponType_WeaponTypeAxe, proto.WeaponType_WeaponTypePolearm) {
	case core.ProcMaskMelee:
		warrior.AddStat(stats.MeleeCrit, 1*core.CritRatingPerCritChance*float64(points))
	case core.ProcMaskMeleeMH:
		warrior.AddStat(stats.MeleeCrit, 1*core.CritRatingPerCritChance*float64(points))
		warrior.OnSpellRegistered(func(spell *core.Spell) {
			if spell.ProcMask.Matches(core.ProcMaskMeleeOH) {
				spell.BonusCritRating -= 1 * core.CritRatingPerCritChance * float64(points)
			}
		})
	case core.ProcMaskMeleeOH:
		warrior.OnSpellRegistered(func(spell *core.Spell) {
			if spell.ProcMask.Matches(core.ProcMaskMeleeOH) {
				spell.BonusCritRating += 1 * core.CritRatingPerCritChance * float64(points)
			}
		})
	}
}

func (warrior *Warrior) registerWeaponmasterExtraAttack(procMask core.ProcMask, procChance float64) {
	icd := core.Cooldown{
		Timer:    warrior.NewTimer(),
		Duration: time.Millisecond * 200,
	}

	warrior.RegisterAura(core.Aura{
		Label:    "Weaponmaster",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() {
				return
			}
			if !spell.ProcMask.Matches(procMask) {
				return
			}
			if !icd.IsReady(sim) {
				return
			}
			if sim.RandomFloat("Weaponmaster") < procChance {
				icd.Use(sim)
				warrior.AutoAttacks.ExtraMHAttack(sim, 1, core.ActionID{SpellID: 12815}, spell.ActionID)
			}
		},
	})
}

// Bloodthrill hands out Overpower charges off the back of Rend instead of waiting for a dodge.
// 4% a point from main hand white hits and abilities (client 70338: curve 4 to 20, proc types
// melee auto and melee ability, and the tooltip says main hand).
func (warrior *Warrior) applyBloodthrill() {
	if warrior.Talents.Bloodthrill == 0 {
		return
	}

	procChance := 0.04 * float64(warrior.Talents.Bloodthrill)

	warrior.BloodthrillAura = warrior.RegisterAura(core.Aura{
		Label:    "Bloodthrill",
		ActionID: core.ActionID{SpellID: 56638},
		Duration: time.Second * 6,
	})

	core.MakePermanent(warrior.RegisterAura(core.Aura{
		Label: "Bloodthrill Trigger",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMeleeMH) {
				return
			}

			if !warrior.Rend.Dot(result.Target).IsActive() {
				return
			}

			if sim.Proc(procChance, "Bloodthrill") {
				warrior.BloodthrillAura.Activate(sim)
			}
		},
	}))
}

// 12% a point for 1 Rage on white hits, whatever the weapon. The 1 October build took away the
// 2 Rage two-handers got (client 70338: curve 12 to 60, spell 12964 gives 1 Rage, proc type
// melee auto attack).
func (warrior *Warrior) applyUnbridledWrath() {
	if warrior.Talents.UnbridledWrath == 0 {
		return
	}

	procChance := 0.12 * float64(warrior.Talents.UnbridledWrath)
	rageGain := 1.0

	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: 12964})

	warrior.RegisterAura(core.Aura{
		Label:    "Unbridled Wrath",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() {
				return
			}

			if spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) && sim.RandomFloat("Unbrided Wrath") < procChance {
				warrior.AddRage(sim, rageGain, rageMetrics)
			}
		},
	})
}

// 5% off-hand damage and 10% off-hand Rage a point. The hit chance it had went to Furious
// Precision on 1 October (client 70338 curves, and Aidan Moon's 50% at 5 points).
func (warrior *Warrior) applyDualWieldSpecialization() {
	points := warrior.Talents.DualWieldSpecialization
	if points == 0 {
		return
	}

	multiplier := 1 + 0.05*float64(points)

	warrior.AddOffHandDealtRageMultiplier(1 + 0.1*float64(points))

	warrior.OnSpellRegistered(func(spell *core.Spell) {
		if spell.ProcMask.Matches(core.ProcMaskMeleeOH) && spell.BonusCoefficient > 0 {
			spell.DamageMultiplier *= multiplier
		}
	})
}

// Furious Precision gives the off hand 4, 7 or 10% hit (client 70338, curve of spell 1323963).
func (warrior *Warrior) applyFuriousPrecision() {
	if warrior.Talents.FuriousPrecision == 0 {
		return
	}

	bonusHit := []float64{0, 4, 7, 10}[warrior.Talents.FuriousPrecision] * core.MeleeHitRatingPerHitChance
	warrior.OnSpellRegistered(func(spell *core.Spell) {
		if spell.ProcMask.Matches(core.ProcMaskMeleeOH) {
			spell.BonusHitRating += bonusHit
		}
	})
}

// Forever turns Enrage into a 30% chance to gain 2% Physical damage a point for 12 sec from any
// damage taken, where Classic only fired it on crits (client 70338: 30% proc chance on spell
// 12317, curve 2 to 10 for the damage). We read the BlizzCon tooltip the other way round, a
// chance growing by points for a flat 2%, until the client tree came in.
func (warrior *Warrior) applyEnrage() {
	if warrior.Talents.Enrage == 0 {
		return
	}

	procChance := 0.3
	multiplier := 1 + 0.02*float64(warrior.Talents.Enrage)

	warrior.EnrageAura = warrior.GetOrRegisterAura(core.Aura{
		Label:    "Enrage",
		ActionID: core.ActionID{SpellID: 12880},
		Duration: time.Second * 12,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			warrior.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= multiplier
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			warrior.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] /= multiplier
		},
	})

	warrior.EnrageAura.NewExclusiveEffect("Enrage", true, core.ExclusiveEffect{Priority: 2})

	core.MakePermanent(warrior.RegisterAura(core.Aura{
		Label: "Enrage Trigger",
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Damage <= 0 {
				return
			}

			if sim.Proc(procChance, "Enrage") {
				warrior.EnrageAura.Activate(sim)
			}
		},
	}))
}

func (warrior *Warrior) applyFlurry() {
	if warrior.Talents.Flurry == 0 {
		return
	}

	talentAura := warrior.makeFlurryAura(warrior.Talents.Flurry)

	// This must be registered before the below trigger because in-game a crit weapon swing consumes a stack before the refresh, so you end up with:
	// 3 => 2
	// refresh
	// 2 => 3
	warrior.makeFlurryConsumptionTrigger(talentAura)

	core.MakePermanent(warrior.RegisterAura(core.Aura{
		Label: "Flurry Proc Trigger",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ProcMask.Matches(core.ProcMaskMelee) && result.Outcome.Matches(core.OutcomeCrit) {
				talentAura.Activate(sim)
				if talentAura.IsActive() {
					talentAura.SetStacks(sim, 3)
				}
				return
			}
		},
	}))
}

// These are separated out because of the T1 Shaman Tank 2P that can proc Flurry separately from the talent.
// It triggers the max-rank Flurry aura but with dodge, parry, or block.
func (warrior *Warrior) makeFlurryAura(points int32) *core.Aura {
	if points == 0 {
		return nil
	}

	spellID := []int32{12319, 12971, 12972, 12973, 12974}[points-1]
	attackSpeed := []float64{1.05, 1.1, 1.15, 1.2, 1.25}[points-1]

	aura := warrior.GetOrRegisterAura(core.Aura{
		Label:     fmt.Sprintf("Flurry Proc (%d)", spellID),
		ActionID:  core.ActionID{SpellID: spellID},
		Duration:  core.NeverExpires,
		MaxStacks: 3,
	})

	aura.NewExclusiveEffect("Flurry", true, core.ExclusiveEffect{
		Priority: attackSpeed,
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			warrior.MultiplyMeleeSpeed(sim, attackSpeed)
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			warrior.MultiplyMeleeSpeed(sim, 1/attackSpeed)
		},
	})

	return aura
}

// With the Protection T2 4pc it's possible to have 2 different Flurry auras if using less than 5/5 points in Flurry.
// The two different buffs don't stack whatsoever. Instead the stronger aura takes precedence and each one is only refreshed by the corresponding triggers.
func (warrior *Warrior) makeFlurryConsumptionTrigger(flurryAura *core.Aura) *core.Aura {
	icd := core.Cooldown{
		Timer:    warrior.NewTimer(),
		Duration: time.Millisecond * 500,
	}
	return core.MakePermanent(warrior.GetOrRegisterAura(core.Aura{
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

func (warrior *Warrior) applyShieldSpecialization() {
	if warrior.Talents.ShieldSpecialization == 0 {
		return
	}

	warrior.AddStat(stats.Block, core.BlockRatingPerBlockChance*1*float64(warrior.Talents.ShieldSpecialization))

	procChance := 0.2 * float64(warrior.Talents.ShieldSpecialization)
	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: 12727})

	warrior.RegisterAura(core.Aura{
		Label:    "Shield Specialization",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.DidBlock() {
				if sim.Proc(procChance, "Shield Specialization") {
					warrior.AddRage(sim, 5.0, rageMetrics)
				}
			}
		},
	})
}

// 50% a point for 5 Rage (client 70338: curve 50 to 100, spell 23602 gives 5 Rage).
func (warrior *Warrior) applyMasterOfDefense() {
	if warrior.Talents.MasterOfDefense == 0 || !warrior.PseudoStats.CanBlock {
		return
	}

	procChance := 0.5 * float64(warrior.Talents.MasterOfDefense)
	rageGain := 5.0
	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: 12727, Tag: 1})

	warrior.RegisterAura(core.Aura{
		Label:    "Master of Defense",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.DidDodge() || result.DidParry() {
				if sim.Proc(procChance, "Master of Defense") {
					warrior.AddRage(sim, rageGain, rageMetrics)
				}
			}
		},
	})
}

// 2% a point with a shield (client 70338, curve of spell 16538).
func (warrior *Warrior) applyBastion() {
	if warrior.Talents.Bastion == 0 || !warrior.PseudoStats.CanBlock {
		return
	}

	warrior.PseudoStats.DamageDealtMultiplier *= 1 + 0.02*float64(warrior.Talents.Bastion)
}

// 1 Rage a point off the offensive abilities (client 70338, curve of spell 29787).
func (warrior *Warrior) applyFocusedRage() {
	if warrior.Talents.FocusedRage == 0 {
		return
	}

	discount := warrior.Talents.FocusedRage
	warrior.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Flags.Matches(SpellFlagOffensive) && spell.Cost != nil {
			spell.Cost.FlatModifier -= discount
		}
	})
}

func (warrior *Warrior) registerDeathWishCD() {
	if !warrior.Talents.DeathWish {
		return
	}

	actionID := core.ActionID{SpellID: 12328}

	deathWishAura := warrior.RegisterAura(core.Aura{
		Label:    "Death Wish",
		ActionID: actionID,
		Duration: time.Second * 30,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			warrior.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= 1.2
			warrior.PseudoStats.DamageTakenMultiplier *= 1.05
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			warrior.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] /= 1.2
			warrior.PseudoStats.DamageTakenMultiplier /= 1.05
		},
	})
	core.RegisterPercentDamageModifierEffect(deathWishAura, 1.2)

	warrior.DeathWish = warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagHelpful,
		RageCost: core.RageCostOptions{
			Cost: 10,
		},
		Cast: core.CastConfig{
			IgnoreHaste: true,
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Minute * 3,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			deathWishAura.Activate(sim)
		},
	})

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: warrior.DeathWish.Spell,
		Type:  core.CooldownTypeDPS,
	})
}

func (warrior *Warrior) registerLastStandCD() {
	if !warrior.Talents.LastStand {
		return
	}

	actionID := core.ActionID{SpellID: 12975}
	healthMetrics := warrior.NewHealthMetrics(actionID)

	var bonusHealth float64
	lastStandAura := warrior.RegisterAura(core.Aura{
		Label:    "Last Stand",
		ActionID: actionID,
		Duration: time.Second * 20,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			bonusHealth = warrior.MaxHealth() * 0.3
			warrior.AddStatsDynamic(sim, stats.Stats{stats.Health: bonusHealth})
			warrior.GainHealth(sim, bonusHealth, healthMetrics)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			warrior.AddStatsDynamic(sim, stats.Stats{stats.Health: -bonusHealth})
		},
	})

	lastStandSpell := warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID: actionID,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: core.TernaryDuration(warrior.Env.IsForever(), time.Minute*3, time.Minute*10),
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			lastStandAura.Activate(sim)
		},
	})

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: lastStandSpell.Spell,
		Type:  core.CooldownTypeSurvival,
	})
}

func (warrior *Warrior) impale() float64 {
	return 0.1 * float64(warrior.Talents.Impale)
}
