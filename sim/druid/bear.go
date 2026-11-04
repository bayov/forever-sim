package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The Forever bear (docs/tank_sim.md, Rules audit: Feral tank). Values from the client, build
// 1.60.1.70707 with the launch realms' hotfixes (wago.tools), and where the client is silent from
// wowsims/forever's bear (its sim/druid, fbcd9a3), whose measurements this repo already cites.

// bearSwingSpeed is Bear and Dire Bear Form's attack speed: SpellShapeshiftForm 5 and 8 have a
// 2500 ms CombatRoundTime and a DamageVariance of 0.4, like Cat Form's 1000 ms.
const bearSwingSpeed = 2.5

// GetBearWeapon is what the bear swings: the equipped weapon's DPS at 2.5 sec, 20% either side,
// with the flat damage of the weapon's enchant and stone on top (as GetCatWeapon). It makes Rage at
// the two-hander's rate whatever the weapon: wowsims/forever reads 11.25 Rage a swing at 2.5 sec in
// a combat log, 4.5 a second of swing (their forms.go, BearRageIsTwoHanded).
func (druid *Druid) GetBearWeapon() core.Weapon {
	if druid.Env != nil && druid.Env.IsForever() && druid.HasMHWeapon() {
		weapon := druid.WeaponFromMainHand()
		if weapon.SwingSpeed > 0 {
			average := weapon.DPS() * bearSwingSpeed
			return core.Weapon{
				BaseDamageMin:        average*(1-catDamageVariance/2) + druid.catFlatWeaponDamage,
				BaseDamageMax:        average*(1+catDamageVariance/2) + druid.catFlatWeaponDamage,
				SwingSpeed:           bearSwingSpeed,
				NormalizedSwingSpeed: bearSwingSpeed,
				AttackPowerPerDPS:    core.DefaultAttackPowerPerDPS,
				RageAsTwoHand:        true,
			}
		}
	}
	return classicBearPaw
}

// classicBearPaw is 1.12's Bear Form weapon.
var classicBearPaw = core.Weapon{
	BaseDamageMin:        109,
	BaseDamageMax:        165,
	SwingSpeed:           bearSwingSpeed,
	NormalizedSwingSpeed: bearSwingSpeed,
	AttackPowerPerDPS:    core.DefaultAttackPowerPerDPS,
	RageAsTwoHand:        true,
}

// bearFormThreatMultiplier is Bear Form (Passive2) 21178: A_MOD_THREAT on every school, 50 in the
// hotfixed client. It is the whole form bonus: the sim's Feral Instinct (Swipe damage) adds no threat.
const bearFormThreatMultiplier = 1.5

// bearFormThreatMultiplierNov2 is the November 2 development notes' Bear and Dire Bear Form threat
// bonus, 115% (was 50%), "to compensate for Blessing of Salvation being removed"
// (orchestration/audit/ref/client_update/sources/devnotes_post9_2026-11-02.txt). Clients 70753 and
// 70755 unchanged (21178 bp 50); standards 37 models the notes' value under Forever.
const bearFormThreatMultiplierNov2 = 2.15

func (druid *Druid) formThreatMultiplier() float64 {
	return core.TernaryFloat64(druid.Env.IsForever(), bearFormThreatMultiplierNov2, bearFormThreatMultiplier)
}

// bearFormValues are a bear form's passive (Bear Form (Passive) 1178 from level 10, Dire Bear Form
// (Passive) 9635 from 40): armor from items up by the percent, health and attack power, each a base
// plus a per level growth from the form's level.
type bearFormValues struct {
	spellID     int32
	level       int32
	armorPct    float64
	health      float64
	healthLevel float64
	ap          float64
	apLevel     float64
}

var bearForm = bearFormValues{spellID: 5487, level: 10, armorPct: 180, health: 20, healthLevel: 18, ap: 30, apLevel: 3}
var direBearForm = bearFormValues{spellID: 9634, level: 40, armorPct: 360, health: 600, healthLevel: 32, ap: 120, apLevel: 3}

func (druid *Druid) bearFormValues() bearFormValues {
	if druid.Level >= direBearForm.level {
		return direBearForm
	}
	return bearForm
}

// BearArmorMultiplier is the form's armor from items.
func (druid *Druid) BearArmorMultiplier() float64 {
	return 1 + druid.bearFormValues().armorPct/100
}

// registerBearFormSpell is Bear Form, or Dire Bear Form from level 40, the tank's starting form.
func (druid *Druid) registerBearFormSpell() {
	form := druid.bearFormValues()
	levels := float64(druid.Level - form.level)
	actionID := core.ActionID{SpellID: form.spellID}

	statBonus := druid.GetFormShiftStats().Add(stats.Stats{
		stats.AttackPower: form.ap + form.apLevel*levels,
		stats.Health:      form.health + form.healthLevel*levels,
		// Feral Swiftness and Natural Reaction's dodge are the druid's (talents.go): Feral
		// Swiftness only in Cat Form there, here in Bear Form too, as its text has it.
		stats.Dodge: 2 * float64(druid.Talents.FeralSwiftness) * core.DodgeRatingPerDodgeChance,
	})
	armorMultiplier := druid.BearArmorMultiplier()
	// Thick Hide: base armor the form scales, 1 a level a point (a level 30 character measured
	// 30 / 60 / 90 in Cat Form and 84 / 168 / 252 in Bear Form, wowsims/forever), and another 0.67 a
	// point for each point of Defense past five times the level, which the forms scale too.
	thickHide := float64(druid.Talents.ThickHide)
	if thickHide > 0 {
		statBonus[stats.Armor] += thickHide * float64(druid.Level) * armorMultiplier
	}
	// The Defense part is read when the form comes up: the engine has no Defense to Armor
	// dependency, and the tank stays in its form.
	thickHidePerDefense := 0.67 * thickHide * armorMultiplier
	var thickHideDefenseArmor float64

	feralApDep := druid.NewDynamicStatDependency(stats.FeralAttackPower, stats.AttackPower, 1)
	var hotwDep *stats.StatDependency
	if druid.Talents.HeartOfTheWild > 0 {
		// Forever's Heart of the Wild: 4% Stamina a point in the bear forms.
		hotwDep = druid.NewDynamicMultiplyStat(stats.Stamina, 1+0.04*float64(druid.Talents.HeartOfTheWild))
	}

	threatMultiplier := druid.formThreatMultiplier()
	druid.BearFormAura = druid.RegisterAura(core.Aura{
		Label:      "Bear Form",
		ActionID:   actionID,
		Duration:   core.NeverExpires,
		BuildPhase: core.Ternary(druid.StartingForm.Matches(Bear), core.CharacterBuildPhaseBase, core.CharacterBuildPhaseNone),
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			if !druid.Env.MeasuringStats && druid.form != Humanoid {
				druid.CancelShapeshift(sim)
			}
			druid.form = Bear
			druid.SetCurrentPowerBar(core.RageBar)
			druid.AutoAttacks.SetMH(druid.GetBearWeapon())
			druid.PseudoStats.ThreatMultiplier *= threatMultiplier
			druid.SetShapeshift(aura)

			druid.ApplyDynamicEquipScaling(sim, stats.Armor, armorMultiplier)
			druid.AddStatsDynamic(sim, statBonus)
			druid.EnableDynamicStatDep(sim, feralApDep)
			if hotwDep != nil {
				druid.EnableDynamicStatDep(sim, hotwDep)
			}
			if thickHidePerDefense > 0 {
				thickHideDefenseArmor = thickHidePerDefense * max(0, druid.GetStat(stats.Defense))
				druid.AddStatDynamic(sim, stats.Armor, thickHideDefenseArmor)
			}

			if !druid.Env.MeasuringStats {
				druid.AutoAttacks.SetReplaceMHSwing(druid.TryMaul)
				druid.AutoAttacks.EnableAutoSwing(sim)
				druid.manageCooldownsEnabled()
				druid.UpdateManaRegenRates()
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.form = Humanoid
			druid.SetCurrentPowerBar(core.ManaBar)
			druid.AutoAttacks.SetMH(druid.humanoidWeapon())
			druid.PseudoStats.ThreatMultiplier /= threatMultiplier
			druid.SetShapeshift(nil)

			druid.RemoveDynamicEquipScaling(sim, stats.Armor, armorMultiplier)
			druid.AddStatsDynamic(sim, statBonus.Invert())
			druid.DisableDynamicStatDep(sim, feralApDep)
			if hotwDep != nil {
				druid.DisableDynamicStatDep(sim, hotwDep)
			}
			if thickHideDefenseArmor > 0 {
				druid.AddStatDynamic(sim, stats.Armor, -thickHideDefenseArmor)
				thickHideDefenseArmor = 0
			}

			if !druid.Env.MeasuringStats {
				druid.AutoAttacks.SetReplaceMHSwing(nil)
				druid.AutoAttacks.EnableAutoSwing(sim)
				druid.manageCooldownsEnabled()
				druid.UpdateManaRegenRates()
				if druid.EnrageAura != nil {
					druid.EnrageAura.Deactivate(sim)
				}
				if druid.MaulQueueAura != nil {
					druid.MaulQueueAura.Deactivate(sim)
				}
			}
		},
	})

	// The shift: a bear's Rage starts at 0 in its form, and Furor gives 10 on top 20% of the time a
	// point (17056 effect 0, 17057's 10 Rage). It can be cast from Bear Form too, which leaves the form
	// and takes it again (Enrage and a queued Maul go with the old one), so a bear can shift before the
	// pull for Furor's Rage. The form the tank starts the fight in isn't a shift and keeps its starting
	// Rage (as wowsims/forever, 5e8c7f4efc).
	furorChance := 0.2 * float64(druid.Talents.Furor)
	furorMetrics := druid.NewRageMetrics(core.ActionID{SpellID: 17057})
	shiftMetrics := druid.NewRageMetrics(actionID)
	druid.BearForm = druid.RegisterSpell(Any, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,
		ManaCost: core.ManaCostOptions{
			BaseCost: 0.55,
			Factor:   druid.naturalShapeshifterFactor(),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.GCDDefault},
			IgnoreHaste: true,
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			if druid.BearFormAura.IsActive() {
				if druid.EnrageAura != nil {
					druid.EnrageAura.Deactivate(sim)
				}
				if druid.MaulQueueAura != nil {
					druid.MaulQueueAura.Deactivate(sim)
				}
			} else {
				druid.BearFormAura.Activate(sim)
			}
			if rage := druid.CurrentRage(); rage > 0 {
				druid.SpendRage(sim, rage, shiftMetrics)
			}
			if furorChance > 0 && sim.Proc(furorChance, "Furor") {
				druid.AddRage(sim, 10, furorMetrics)
			}
		},
	})
}

// brutalityRage is Idol of Brutality's 2 Rage off Maul and Swipe (items.go).
func (druid *Druid) brutalityRage() float64 {
	if ranged := druid.Ranged(); ranged != nil && ranged.ID == IdolOfBrutality {
		return 2
	}
	return 0
}

// applyBearTalentProcs is Natural Reaction's and Blood Frenzy's Rage in the bear forms: 5 Rage on
// a dodge at a 20% chance a point (Natural Reaction, 417053 restores 5), and 5 Rage on a crit at
// 50% a point (Blood Frenzy, 16959). The cat's half of Blood Frenzy is in talents.go.
func (druid *Druid) applyBearTalentProcs() {
	reaction := 0.2 * float64(druid.Talents.NaturalReaction)
	frenzy := 0.5 * float64(druid.Talents.BloodFrenzy)
	if reaction == 0 && frenzy == 0 {
		return
	}
	reactionMetrics := druid.NewRageMetrics(core.ActionID{SpellID: 417053})
	frenzyMetrics := druid.NewRageMetrics(core.ActionID{SpellID: 16959})
	core.MakePermanent(druid.RegisterAura(core.Aura{
		Label: "Bear Talent Rage",
		OnSpellHitTaken: func(_ *core.Aura, sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			if reaction > 0 && druid.InForm(Bear) && result.Outcome.Matches(core.OutcomeDodge) && sim.Proc(reaction, "Natural Reaction") {
				druid.AddRage(sim, 5, reactionMetrics)
			}
		},
		OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if frenzy > 0 && druid.InForm(Bear) && result.Outcome.Matches(core.OutcomeCrit) &&
				spell.ProcMask.Matches(core.ProcMaskMelee) && sim.Proc(frenzy, "Blood Frenzy Rage") {
				druid.AddRage(sim, 5, frenzyMetrics)
			}
		},
	}))
}

// The bear's Demoralizing Roar rank 5 (9898): Forever's attack power cut (core, 204 at 60) for 30
// sec on every enemy, 10 Rage, and 42 + 0.8 a level from 52 flat threat a target (its effect 1).
// Clearcasting makes it free (16870's class mask takes in its 0x8).
func (druid *Druid) registerDemoralizingRoarSpell() {
	druid.DemoralizingRoarAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.ForeverDemoralizingRoarAura(target, druid.Level)
	})
	flatThreat := 42 + 0.8*float64(min(druid.Level, 62)-52)

	druid.DemoralizingRoar = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 9898},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagAPL | SpellFlagOmen,
		RageCost:    core.RageCostOptions{Cost: 10},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.GCDDefault},
			IgnoreHaste: true,
		},
		ThreatMultiplier: 1,
		FlatThreatBonus:  flatThreat,
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			for _, target := range sim.Encounter.ActiveTargetUnits {
				result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
				if result.Landed() {
					druid.DemoralizingRoarAuras.Get(target).Activate(sim)
				}
			}
		},
		RelatedAuras: []core.AuraArray{druid.DemoralizingRoarAuras},
	})
}

// Enrage (5229): 10 Rage at once and 2 a second for 10 sec (effect 0, 20 tenths a 1 sec
// tick; Enraged Idol 3), off the GCD on a 1 min cooldown, and armor from items down 16% in Dire Bear Form (27% in
// Bear Form) while it lasts, the tooltip's (the client's 25503 states 26, on no form).
func (druid *Druid) registerEnrageSpell() {
	actionID := core.ActionID{SpellID: 5229}
	rageMetrics := druid.NewRageMetrics(actionID)
	armorScale := core.TernaryFloat64(druid.Level >= direBearForm.level, 0.84, 0.73)
	var tick *core.PendingAction

	druid.EnrageAura = druid.RegisterAura(core.Aura{
		Label:    "Enrage",
		ActionID: actionID,
		Duration: time.Second * 10,
		OnGain: func(_ *core.Aura, sim *core.Simulation) {
			druid.ApplyDynamicEquipScaling(sim, stats.Armor, armorScale)
			tick = core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				Period:   time.Second,
				NumTicks: 10,
				OnAction: func(sim *core.Simulation) {
					druid.AddRage(sim, 2+druid.enrageTickBonus, rageMetrics)
				},
			})
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.RemoveDynamicEquipScaling(sim, stats.Armor, armorScale)
			// The tenth tick lands as the aura runs out, so it's only cut short by an early end.
			if tick != nil && sim.CurrentTime < aura.StartedAt()+aura.Duration {
				tick.Cancel(sim)
			}
			tick = nil
		},
	})

	druid.Enrage = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL | core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: core.Cooldown{Timer: druid.NewTimer(), Duration: time.Minute},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.AddRage(sim, 10, rageMetrics)
			druid.EnrageAura.Activate(sim)
		},
	})
}

// Frenzied Regeneration (22842, Forever's only rank): for 10 sec, each second turns up to 10 Rage
// into health, 1% of maximum health a point of Rage (effect 0's 1, read as a percent, as
// wowsims/forever does), 3 min cooldown. It reads nothing of the bear's health, so a tank score's
// replay follows it; when to press it is the rotation's.
func (druid *Druid) registerFrenziedRegenerationSpell() {
	actionID := core.ActionID{SpellID: 22842}
	rageMetrics := druid.NewRageMetrics(actionID)
	healthMetrics := druid.NewHealthMetrics(actionID)
	var tick *core.PendingAction

	druid.FrenziedRegenerationAura = druid.RegisterAura(core.Aura{
		Label:    "Frenzied Regeneration",
		ActionID: actionID,
		Duration: time.Second * 10,
		OnGain: func(_ *core.Aura, sim *core.Simulation) {
			tick = core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				Period:   time.Second,
				NumTicks: 10,
				OnAction: func(sim *core.Simulation) {
					rage := min(druid.CurrentRage(), 10)
					if rage <= 0 {
						return
					}
					druid.SpendRage(sim, rage, rageMetrics)
					druid.GainHealth(sim, rage*0.01*druid.MaxHealth()*druid.PseudoStats.HealingTakenMultiplier, healthMetrics)
				},
			})
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			// As Enrage's: the tenth tick lands as it runs out.
			if tick != nil && sim.CurrentTime < aura.StartedAt()+aura.Duration {
				tick.Cancel(sim)
			}
			tick = nil
		},
	})

	druid.FrenziedRegeneration = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL | core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.GCDDefault},
			IgnoreHaste: true,
			CD:          core.Cooldown{Timer: druid.NewTimer(), Duration: time.Minute * 3},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.FrenziedRegenerationAura.Activate(sim)
		},
	})
}

// Bash (5211 at 14, 6798 at 30, 8983 at 46: a 2, 3 or 4 sec stun): 10 Rage, 1 min cooldown, the
// global cooldown, a melee attack that can miss, be dodged or parried, with 80% of its Rage back
// when it fails (Attributes_1 0x8000000). Brutal Impact: 0.5 sec more stun and 15 sec less cooldown
// a point (16940's curves: 500 and -15000 ms). It deals no damage and has no threat in the client,
// and raid bosses can't be stunned, so it does nothing to a tank score; it's here for the APL and
// the leveling sim. Clearcasting makes it free (16870's mask takes in its 0x2000).
func (druid *Druid) registerBashSpell() {
	ranks := []struct {
		level   int32
		spellID int32
		stun    time.Duration
	}{{14, 5211, 2 * time.Second}, {30, 6798, 3 * time.Second}, {46, 8983, 4 * time.Second}}
	if druid.Level < ranks[0].level {
		return
	}
	rank := ranks[0]
	for _, r := range ranks {
		if druid.Level >= r.level {
			rank = r
		}
	}
	actionID := core.ActionID{SpellID: rank.spellID}
	impact := time.Duration(druid.Talents.BrutalImpact)
	stuns := druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(core.Aura{
			Label:    "Bash",
			ActionID: actionID,
			Duration: rank.stun + impact*500*time.Millisecond,
		})
	})

	druid.Bash = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagOmen,
		RageCost: core.RageCostOptions{
			Cost:   10,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.GCDDefault},
			IgnoreHaste: true,
			CD:          core.Cooldown{Timer: druid.NewTimer(), Duration: time.Minute - impact*15*time.Second},
		},
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
			if result.Landed() {
				stuns.Get(target).Activate(sim)
			} else {
				spell.IssueRefund(sim)
			}
		},
		RelatedAuras: []core.AuraArray{stuns},
	})
}
