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
	warrior.applyBloodCraze()
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
	// Under Forever it raises all Physical damage while a two-hander is equipped, Rend, Deep
	// Wounds and Thunder Clap included (client 70425, spell 12163: A_MOD_DAMAGE_PERCENT_DONE on
	// school 1, equip class two-handers), by 2/4/6% (development notes post 9, 2026-11-02: "was
	// 1/2/3%"; hotfix push 113539, curve 83308 2/4/6; client 70755 still has the 1% curve).
	if warrior.Env.IsForever() {
		multiplier = 1 + 0.02*float64(min(warrior.Talents.TwoHandedWeaponSpecialization, 3))
		warrior.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= multiplier
		warrior.twoHandSpecPhysical = multiplier
		return
	}
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
			if isOffHandWeaponAttack(spell) {
				spell.BonusCritRating -= 1 * core.CritRatingPerCritChance * float64(points)
			}
		})
	case core.ProcMaskMeleeOH:
		warrior.OnSpellRegistered(func(spell *core.Spell) {
			if isOffHandWeaponAttack(spell) {
				spell.BonusCritRating += 1 * core.CritRatingPerCritChance * float64(points)
			}
		})
	}
}

// Shield Bash ranks 1 to 3.
func isShieldBash(spellID int32) bool {
	switch spellID {
	case 72, 1671, 1672:
		return true
	}
	return false
}

// isShieldStrike is Shield Slam or Shield Bash. They strike with the shield, so the off-hand weapon
// modifiers (Dual Wield Specialization, Furious Precision, Weaponmaster's off-hand crit) pass them
// by; under Forever Shield Slam is a main-hand special no weapon proc hears (R-0010, R-0041).
func isShieldStrike(spell *core.Spell) bool {
	return spell.SpellCode == SpellCode_WarriorShieldSlam || isShieldBash(spell.SpellID)
}

// isOffHandWeaponAttack is an attack made with the off-hand weapon.
func isOffHandWeaponAttack(spell *core.Spell) bool {
	return spell.ProcMask.Matches(core.ProcMaskMeleeOH) && !isShieldStrike(spell)
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
			// "Fixed an issue where Weaponmaster's Sword Specialization could be activated by
			// Shield Bash and Shield Slam" (development notes, 2026-10-15, build 1.60.1.70425).
			if isShieldStrike(spell) {
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
// 10% a point, 2 points (development notes post 9, 2026-11-02: "a 10/20% chance (was a 5-point
// talent granting a 4/8/12/16/20% chance)"; hotfix push 113539, curve 109366 10/20; client 70755 still has 4% a point) from main hand white swings, extra attacks included,
// and the on-next-swing Heroic Strike and Cleave. The row's proc types also name melee abilities,
// but in game the instant specials never proc it (wowsims/forever 788f62ce0f, combat logs:
// Hamstring 0/99, Overpower 0/9, Execute 0/7, off hand 0/200, Thunder Clap 0/13).
//
// Under Forever Windfury Totem's attack doesn't proc it, though the sim swings it as a white
// extra attack (user ruling 2026-11-15). wowsims/forever's logs (their 56473177e7): EV-0056's
// bystander warriors' 24 gains came from white swings 19, Heroic Strike 4 and Cleave 1, none from
// a totem hit, and EV-0136's Warcraft Logs sweep (to 2026-11-16) counts 10 of 1,633 totem hits from
// 9 wearers followed by a Bloodthrill gain (0.61%, 0.33-1.12), no more than the 0.74% the ability
// before the totem hit gives alone, against 3.5% a white swing. So those 10 are the ability's or
// swing's before it, not the totem hit's, and it rolls none. Their sim (talents_arms.go
// registerBloodthrill) strikes the totem's attack as a special, which the condition leaves out.
func (warrior *Warrior) applyBloodthrill() {
	if warrior.Talents.Bloodthrill == 0 {
		return
	}

	procChance := 0.04 * float64(warrior.Talents.Bloodthrill)
	if warrior.Env.IsForever() {
		procChance = 0.1 * float64(min(warrior.Talents.Bloodthrill, 2))
	}

	// The Overpower window it opens is spell 1282733, 5 sec (duration index 0 of client
	// 70009 to 70425; until 69893 the talent pointed at 1289681, 6 sec).
	warrior.BloodthrillAura = warrior.RegisterAura(core.Aura{
		Label:    "Bloodthrill",
		ActionID: core.ActionID{SpellID: 56638},
		Duration: time.Second * 5,
	})

	forever := warrior.Env.IsForever()
	core.MakePermanent(warrior.RegisterAura(core.Aura{
		Label: "Bloodthrill Trigger",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			// Heroic Strike and Cleave carry the main hand auto mask as well.
			if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMeleeMHAuto) {
				return
			}
			// Windfury Totem's white swing (a queued Heroic Strike in its place still procs it: in
			// game the strike goes off on the next swing).
			if forever && spell.ProcMask == core.ProcMaskMeleeMHAuto && warrior.AutoAttacks.IsWindfuryTotemAttack() {
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
// melee auto attack). Server hotfix push 114097 (2026-11-17) sets its ProcTypeMask (SpellAuraOptions
// 124699) from 4 to 20, melee auto attacks and melee abilities, so under Forever a landed damaging
// melee ability procs it too.
func (warrior *Warrior) applyUnbridledWrath() {
	if warrior.Talents.UnbridledWrath == 0 {
		return
	}

	procChance := 0.12 * float64(warrior.Talents.UnbridledWrath)
	rageGain := 1.0

	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: 12964})
	// EV-0174 (the user's own log, levels 42 to 47, Defiance 3): 5 threat a Rage, no stance or Defiance.
	rageMetrics.UnscaledThreat = true
	forever := warrior.Env.IsForever()

	// Push 114097's melee abilities: the ones that deal damage (Heroic Strike and Cleave, whose
	// white mask bit already counted, and the rest), but not Whirlwind, Shield Slam, Revenge or
	// Shield Bash. The row gives them no exception and no client flag sets them apart (Whirlwind's
	// attributes match Bloodthirst's); the exceptions are wowsims/forever's logs:
	//   - EV-0148 (2026-11-17 to 18, build 70938, push 114122; a level 60 Protection warrior with
	//     Unbridled Wrath 5/5): isolated landed white swings 58.8%, Heroic Strikes 61.7%, totem
	//     hits 59.3%, each target a Cleave lands on 0.60; Shield Slam 0 of 977, Revenge 0 of 601,
	//     Thunder Clap 0 of 2,238, Retaliation 0 of 49, Victory Rush 0 of 24, Shield Bash 0 of 44.
	//   - EV-0138 (2026-11-16, build 70938, push 114115; a level 60 Fury warrior): Heroic Strike 70%,
	//     Cleave 67%, Windfury Totem hits 59%, Whirlwind 2 of 425 (0.5%).
	// Thunder Clap is a spell in the sim and procs nothing; Retaliation and Victory Rush aren't in
	// it. Bloodthirst, Mortal Strike, Execute, Slam, Overpower and Hamstring have no count in
	// either log, so they keep the row's reading and proc it as a white hit does (To Verify, notes.md).
	procsOnAbility := func(spell *core.Spell, result *core.SpellResult) bool {
		if !spell.ProcMask.Matches(core.ProcMaskMeleeSpecial) || result.Damage <= 0 {
			return false
		}
		switch {
		case spell.ActionID.SpellID == 1680, // Whirlwind, both strikes
			spell.SpellCode == SpellCode_WarriorShieldSlam,
			spell.SpellCode == SpellCode_WarriorRevenge,
			isShieldBash(spell.ActionID.SpellID):
			return false
		}
		return true
	}

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

			if !spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) && !(forever && procsOnAbility(spell, result)) {
				return
			}
			if sim.RandomFloat("Unbrided Wrath") < procChance {
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
		if isOffHandWeaponAttack(spell) && spell.BonusCoefficient > 0 {
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
		if isOffHandWeaponAttack(spell) {
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
	// Forever's Flurry has 3 points for 10/20/30% (development notes post 9, 2026-11-02: "was 5
	// points granting 5/10/15/20/25%"; hotfix push 113539, curve 83289 10/20/30; client 70755 still has the
	// 5 point curve).
	if warrior.Env.IsForever() {
		points = min(points, 3)
		spellID = []int32{12319, 12971, 12972}[points-1]
		attackSpeed = 1 + 0.1*float64(points)
	}

	aura := warrior.GetOrRegisterAura(core.Aura{
		Label:    fmt.Sprintf("Flurry Proc (%d)", spellID),
		ActionID: core.ActionID{SpellID: spellID},
		// Forever's Flurry buff (12966, client 70338) lasts 15 sec.
		Duration:  core.TernaryDuration(warrior.Env.IsForever(), time.Second*15, core.NeverExpires),
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
	forever := warrior.Env.IsForever()
	queuedSpentAt := time.Duration(-1)
	return core.MakePermanent(warrior.GetOrRegisterAura(core.Aura{
		Label: fmt.Sprintf("Flurry Consume Trigger - %d", flurryAura.ActionID.SpellID),
		OnReset: func(_ *core.Aura, _ *core.Simulation) {
			queuedSpentAt = -1
		},
		OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			// Remove a stack: a white swing, landed or not, two within 0.5 sec spending one (report
			// assumption flurry_batch). Server hotfix push 114097 (2026-11-17) sets the buff's
			// ProcTypeMask (SpellAuraOptions 126501) from 4 to 20, melee auto attacks and melee
			// abilities.
			//
			// Under Forever a Heroic Strike or Cleave spends one charge a cast, landed or not, on one
			// target or two, apart from the white swings' 0.5 sec window: wowsims/forever's EV-0138
			// (2026-11-16, build 70938 with hotfixes to push 114115, a level 60 Fury warrior in Upper
			// Blackrock Spire): 36 clean casts with Flurry up each spent 1 (Cleave on 2 targets 22, on
			// 1 of 2 landed 3, on 1 target 2, Heroic Strike 8 landed and 1 dodged), and a Cleave 26 ms
			// after a white swing spent one of its own. The charge goes before the crit's refresh, so
			// a crit ends the cast at 3. Blizzard's November 16 hotfix ("Fixed an issue where Flurry
			// stacks were consumed by Cleave or Heroic Strike", t/2403040 post 11) reads the other
			// way; the log came after it (user ruling 2026-11-22: a rule measured from a Forever log
			// wins). Bloodthirst, Whirlwind and the other abilities spend none.
			if forever && spell.ProcMask.Matches(core.ProcMaskMeleeSpecial) {
				if !warrior.isQueuedSwing(spell) || !flurryAura.IsActive() || queuedSpentAt == sim.CurrentTime {
					return
				}
				queuedSpentAt = sim.CurrentTime // a Cleave's hits all land at its cast
				flurryAura.RemoveStack(sim)
				return
			}
			// Nor does Windfury Totem's attack, though the sim swings it as a white extra attack
			// (user ruling 2026-11-15); its crit still starts or refreshes Flurry (applyFlurry).
			// wowsims/forever's EV-0138 and EV-0140 (their 56473177e7; a Fury and a Protection
			// warrior in Upper Blackrock Spire with rank 3 totem, 2026-11-16, build 70938 with
			// hotfixes to push 114115): of 74 totem hits with Flurry up, 68 remove no charge and the
			// other 6 have a white swing, Heroic Strike, Cleave or a crit's refresh beside them;
			// 22 of 22 totem crits apply or refresh Flurry. Their EV-0136 sweep of earlier reports
			// (to 2026-11-16) had the totem hit spend a charge 99% of the time; push 114115 ended that.
			if forever && warrior.AutoAttacks.IsWindfuryTotemAttack() {
				return
			}
			if flurryAura.IsActive() && spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) && icd.IsReady(sim) {
				icd.Use(sim)
				flurryAura.RemoveStack(sim)
			}
		},
	}))
}

// Blood Craze heals 1 / 2 / 3% of max health over 6 sec, a tick every 2 sec, after a crit taken.
// Under Forever also after Bloodthirst deals damage or a single hit takes more than 20% of max
// health (client 70707: talent 16487, whose second effect holds the 20% and triggers 16488, a
// periodic heal every 2000 ms; forever-wiki/classes/warrior.md). A new trigger restarts it. For
// a tank under Forever's boss crit it's the big hits: a crushing blow on a pre-raid tank is
// about 20% of its health.
func (warrior *Warrior) applyBloodCraze() {
	points := warrior.Talents.BloodCraze
	if points == 0 {
		return
	}
	forever := warrior.Env.IsForever()
	hot := warrior.Unit.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: []int32{0, 16488, 16490, 16491}[points]},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagHelpful,

		DamageMultiplier: 1,
		// The November 2 development notes: "The Blood Craze and Gore Drinker talents no longer
		// generate threat when restoring Health to the Warrior." (client 70753 unchanged, standards 37.)
		ThreatMultiplier: core.TernaryFloat64(forever, 0, 1),

		Hot: core.DotConfig{
			Aura:          core.Aura{Label: "Blood Craze"},
			NumberOfTicks: 3,
			TickLength:    time.Second * 2,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, _ bool) {
				dot.SnapshotBaseDamage = 0.01 * float64(points) * warrior.MaxHealth() / 3
				// The caster's healing modifiers, not its damage ones (Defensive Stance's 10% less
				// damage done doesn't cut it). Healing taken modifiers come at each tick.
				dot.SnapshotAttackerMultiplier = dot.Spell.CasterHealingMultiplier()
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotHealing(sim, target, dot.OutcomeTick)
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.Hot(&warrior.Unit).Apply(sim)
		},
	})

	warrior.RegisterAura(core.Aura{
		Label:    "Blood Craze Trigger",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() {
				return
			}
			if result.DidCrit() || (forever && result.Damage > 0.2*warrior.MaxHealth()) {
				hot.Cast(sim, &warrior.Unit)
			}
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if forever && spell.SpellCode == SpellCode_WarriorBloodthirst && result.Landed() && result.Damage > 0 {
				hot.Cast(sim, &warrior.Unit)
			}
		},
	})
}

func (warrior *Warrior) applyShieldSpecialization() {
	if warrior.Talents.ShieldSpecialization == 0 {
		return
	}

	warrior.AddStat(stats.Block, core.BlockRatingPerBlockChance*1*float64(warrior.Talents.ShieldSpecialization))

	procChance := 0.2 * float64(warrior.Talents.ShieldSpecialization)
	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: 12727})
	// EV-0043: 25 threat a 5 Rage proc in Defensive Stance, never 45 (5 x 1.8).
	rageMetrics.UnscaledThreat = true

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
	// EV-0045: 5 threat a Rage in either stance, never 45 a proc.
	rageMetrics.UnscaledThreat = true

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

// 2% a point of Physical damage with a shield (client 70338, curve of spell 16538; its
// A_MOD_DAMAGE_PERCENT_DONE names school 1 only).
func (warrior *Warrior) applyBastion() {
	if warrior.Talents.Bastion == 0 || !warrior.PseudoStats.CanBlock {
		return
	}

	warrior.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= 1 + 0.02*float64(warrior.Talents.Bastion)
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
