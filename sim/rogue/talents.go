package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

func (rogue *Rogue) ApplyTalents() {
	rogue.applyRuthlessness()
	rogue.applyRelentlessStrikes()
	rogue.applySealFate()
	rogue.applyHackAndSlash()
	rogue.applyWeaponExpertise()
	rogue.applyInitiative()
	rogue.applySerratedBlades()
	rogue.registerDeathMark()

	rogue.AddStat(stats.Dodge, 1*float64(rogue.Talents.LightningReflexes))
	rogue.AddStat(stats.Parry, 2*float64(rogue.Talents.Deflection))
	rogue.AddStat(stats.MeleeCrit, 1*float64(rogue.Talents.Malice))
	rogue.AddStat(stats.MeleeHit, 1*float64(rogue.Talents.Precision))
	// Malice and Precision cover Poisons in Forever, and those roll against the spell stats.
	rogue.AddStat(stats.SpellCrit, 1*float64(rogue.Talents.Malice))
	rogue.AddStat(stats.SpellHit, 1*float64(rogue.Talents.Precision))
	rogue.AutoAttacks.OHConfig().DamageMultiplier *= rogue.dwsMultiplier()

	rogue.registerColdBloodCD()
	rogue.registerBladeFlurryCD()
	rogue.registerAdrenalineRushCD()
	rogue.registerPreparationCD()
	rogue.registerPremeditation()
	rogue.registerGhostlyStrikeSpell()
	rogue.applyRiposte()
}

// dwsMultiplier returns the offhand damage multiplier: 25/50% under the November 2 development
// notes (was 5 ranks at 5%; client 13715 unchanged in 1.60.1.70753).
func (rogue *Rogue) dwsMultiplier() float64 {
	return 1 + 0.25*float64(rogue.Talents.DualWieldSpecialization)
}

// opportunityMultiplier is Backstab, Garrote, Ambush and Mutilate's: 4% a rank, 5 ranks under the
// November 2 development notes (was 2 ranks at 5%; client 14057 unchanged in 1.60.1.70753). One
// percent modifier, so it multiplies the spell's others.
func (rogue *Rogue) opportunityMultiplier() float64 {
	return 1 + 0.04*float64(rogue.Talents.Opportunity)
}

func (rogue *Rogue) applyRuthlessness() {
	if rogue.Talents.Ruthlessness == 0 {
		return
	}

	procChance := 0.2 * float64(rogue.Talents.Ruthlessness)
	cpMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: 14161})
	rogue.OnComboPointsSpent(func(sim *core.Simulation, spell *core.Spell, comboPoints int32) {
		if sim.Proc(procChance, "Ruthlessness") {
			rogue.AddComboPointsIgnoreTarget(sim, 1, cpMetrics)
		}
	})
}

// Relentless Strikes is 5 ranks of 4% a combo point under the November 2 development notes (was one
// rank of 20%; client 14179 unchanged in 1.60.1.70753).
func (rogue *Rogue) applyRelentlessStrikes() {
	if rogue.Talents.RelentlessStrikes == 0 {
		return
	}
	chancePerPoint := 0.04 * float64(rogue.Talents.RelentlessStrikes)

	cpMetrics := rogue.NewEnergyMetrics(core.ActionID{SpellID: 14179})
	rogue.OnComboPointsSpent(func(sim *core.Simulation, spell *core.Spell, comboPoints int32) {
		if sim.Proc(chancePerPoint*float64(comboPoints), "RelentlessStrikes") {
			rogue.AddEnergy(sim, 25, cpMetrics)
		}
	})
}

// Cold Blood talent
func (rogue *Rogue) registerColdBloodCD() {
	if !rogue.Talents.ColdBlood {
		return
	}

	actionID := core.ActionID{SpellID: 14177}

	coldBloodAura := rogue.RegisterAura(core.Aura{
		Label:    "Cold Blood",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range rogue.Spellbook {
				if spell.Flags.Matches(SpellFlagColdBlooded) {
					spell.BonusCritRating += 100 * core.CritRatingPerCritChance
				}
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range rogue.Spellbook {
				if spell.Flags.Matches(SpellFlagColdBlooded) {
					spell.BonusCritRating -= 100 * core.CritRatingPerCritChance
				}
			}
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			// Cold Blood covers both halves of a Mutilate (class deep dive: "Now also applies to
			// Mutilate"; the client masks both). The off hand half lands last and spends it.
			if spell == rogue.Mutilate {
				return
			}
			if spell.Flags.Matches(SpellFlagColdBlooded) {
				aura.Deactivate(sim)
			}
		},
	})

	rogue.ColdBlood = rogue.RegisterSpell(core.SpellConfig{
		ActionID: actionID,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: time.Minute * 3,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			coldBloodAura.Activate(sim)
		},
	})

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell: rogue.ColdBlood,
		Type:  core.CooldownTypeDPS,
		// The next Cold Blooded hit spends the buff, and the builders carry that flag too.
		// Used on cooldown it lands on a Sinister Strike, so we hold it for Eviscerate.
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			return rogue.ReadyForEviscerate(sim, time.Second*6)
		},
	})
}

// ReadyForEviscerate is true when the next rotation action is going to be a 5 point
// Eviscerate: 5 combo points, energy for it, and Slice and Dice not about to run out
// (a rotation refreshes Slice and Dice with those points instead if it is). Buffs that
// are spent by the next ability (Cold Blood, Eureka!) wait for this.
func (rogue *Rogue) ReadyForEviscerate(sim *core.Simulation, minSliceAndDice time.Duration) bool {
	if rogue.ComboPoints() < 5 || rogue.CurrentEnergy() < rogue.Eviscerate.Cost.GetCurrentCost() {
		return false
	}
	return rogue.SliceAndDiceAura.RemainingDuration(sim) >= minSliceAndDice
}

// Seal Fate talent
func (rogue *Rogue) applySealFate() {
	if rogue.Talents.SealFate == 0 {
		return
	}

	// 33/66/100% under the November 2 development notes (was 5 ranks at 20%; client 14186
	// unchanged in 1.60.1.70753).
	procChance := []float64{0, 0.33, 0.66, 1}[rogue.Talents.SealFate]
	cpMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: 14195})

	icd := core.Cooldown{
		Timer:    rogue.NewTimer(),
		Duration: 500 * time.Millisecond,
	}

	// "Mutilate will trigger this exactly once if either or both hands generates a critical
	// strike" (World of Warcraft: Forever class deep dive, Rogue). Both halves land at the same
	// moment, so the first crit of a Mutilate rolls and the other half's is ignored.
	lastMutilateRoll := time.Duration(-1)

	rogue.RegisterAura(core.Aura{
		Label:    "Seal Fate",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
			lastMutilateRoll = -1
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.Flags.Matches(SpellFlagBuilder) {
				return
			}

			if !result.Outcome.Matches(core.OutcomeCrit) {
				return
			}

			if spell == rogue.Mutilate || spell == rogue.mutilateOH {
				if lastMutilateRoll == sim.CurrentTime {
					return
				}
				lastMutilateRoll = sim.CurrentTime
			}

			if icd.IsReady(sim) && sim.Proc(procChance, "Seal Fate") {
				rogue.AddComboPoints(sim, 1, result.Target, cpMetrics)
				icd.Use(sim)
			}
		},
	})
}

// Initiative talent
func (rogue *Rogue) applyInitiative() {
	if rogue.Talents.Initiative == 0 {
		return
	}

	// 33 / 67 / 100% per the beta client, so rank 3 is a sure proc.
	procChance := []float64{0, .33, .67, 1}[rogue.Talents.Initiative]
	cpMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: 13980})

	rogue.RegisterAura(core.Aura{
		Label:    "Initiative",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell == rogue.Garrote || spell == rogue.Ambush {
				if result.Landed() {
					if sim.Proc(procChance, "Initiative") {
						rogue.AddComboPoints(sim, 1, result.Target, cpMetrics)
					}
				}
			}
		},
	})
}

// Weapon Expertise no longer grants weapon skill, it takes the chance for the rogue's
// attacks to be dodged or parried off the attack table directly.
func (rogue *Rogue) applyWeaponExpertise() {
	if rogue.Talents.WeaponExpertise == 0 {
		return
	}

	rogue.AddStat(stats.Expertise, float64(rogue.Talents.WeaponExpertise)*core.ExpertiseRatingPerExpertiseChance)
}

// Serrated Blades ignores a share of the target's Armor rather than a flat amount, 3% per
// rank. The Rupture bonus lives on the spell itself.
func (rogue *Rogue) applySerratedBlades() {
	if rogue.Talents.SerratedBlades == 0 {
		return
	}

	rogue.PseudoStats.ArmorIgnorePercent += 0.03 * float64(rogue.Talents.SerratedBlades)
}

func (rogue *Rogue) registerBladeFlurryCD() {
	if !rogue.Talents.BladeFlurry {
		return
	}

	var curDmg float64
	// Under Forever the second target takes the first hit's final damage as it is: the damage
	// multipliers and armor already went into it, so the copy skips both (wowsims/forever
	// talents_combat.go copies result.Damage with IgnoreResists and IgnoreModifiers). Classic
	// takes the armor back out and lets the copy through the second target's armor and the
	// damage multipliers again.
	forever := rogue.Env.IsForever()
	bfFlags := core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete
	if forever {
		bfFlags |= core.SpellFlagIgnoreResists | core.SpellFlagIgnoreModifiers
	}
	bfHit := rogue.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 22482},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskEmpty, // No proc mask, so it won't proc itself.
		Flags:       bfFlags,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, curDmg, spell.OutcomeAlwaysHit)
		},
	})

	rogue.BladeFlurryAura = rogue.RegisterAura(core.Aura{
		Label:    "Blade Flurry",
		ActionID: core.ActionID{SpellID: 13877},
		Duration: time.Second * 15,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			rogue.MultiplyMeleeSpeed(sim, 1.2)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			rogue.MultiplyMeleeSpeed(sim, 1/1.2)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if sim.Encounter.AoETargetCount(0) < 2 {
				return
			}

			if result.Damage == 0 || !spell.ProcMask.Matches(core.ProcMaskMelee) {
				return
			}

			curDmg = result.Damage
			if !forever {
				// Undo armor reduction to get the raw damage value.
				curDmg /= result.ResistanceMultiplier
			}

			bfHit.Cast(sim, rogue.Env.NextTargetUnit(result.Target))
			bfHit.SpellMetrics[result.Target.UnitIndex].Casts--
		},
	})

	cooldownDur := time.Minute * 2
	rogue.BladeFlurry = rogue.RegisterSpell(core.SpellConfig{
		SpellCode: SpellCode_RogueBladeFlurry,
		ActionID:  core.ActionID{SpellID: 13877},
		Flags:     core.SpellFlagAPL,

		EnergyCost: core.EnergyCostOptions{
			Cost: 25,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: cooldownDur,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			rogue.BladeFlurryAura.Activate(sim)
		},
	})

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell:    rogue.BladeFlurry,
		Type:     core.CooldownTypeDPS,
		Priority: core.CooldownPriorityDefault,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			if sim.GetRemainingDuration() > cooldownDur+time.Second*15 {
				// We'll have enough time to cast another BF, so use it immediately to make sure we get the 2nd one.
				return true
			}

			// Since this is our last BF, wait until we have SND / procs up.
			sndTimeRemaining := rogue.SliceAndDiceAura.RemainingDuration(sim)
			return sndTimeRemaining >= time.Second
		},
	})
}

var AdrenalineRushActionID = core.ActionID{SpellID: 13750}

func (rogue *Rogue) registerAdrenalineRushCD() {
	if !rogue.Talents.AdrenalineRush {
		return
	}

	rogue.AdrenalineRushAura = rogue.RegisterAura(core.Aura{
		Label:    "Adrenaline Rush",
		ActionID: AdrenalineRushActionID,
		Duration: time.Second * 15,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			rogue.ApplyEnergyTickMultiplier(sim, 1.0)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			rogue.ApplyEnergyTickMultiplier(sim, -1.0)
		},
	})

	rogue.AdrenalineRush = rogue.RegisterSpell(core.SpellConfig{
		SpellCode: SpellCode_RogueAdrenalineRush,
		ActionID:  AdrenalineRushActionID,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: time.Minute * 5,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			rogue.AdrenalineRushAura.Activate(sim)
		},
	})

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell:    rogue.AdrenalineRush,
		Type:     core.CooldownTypeDPS,
		Priority: core.CooldownPriorityBloodlust,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			return rogue.CurrentEnergy() <= 45.0
		},
	})
}

// Lethality is 2 ranks of 10% under the November 2 development notes (was 5 ranks of 4%; client
// 14128 unchanged in 1.60.1.70753).
func (rogue *Rogue) lethality() float64 {
	return 0.10 * float64(rogue.Talents.Lethality)
}

// Aggression raises Sinister Strike and Eviscerate by 5/10/15% under Forever (November 2
// development notes; was 2/4/6% and Backstab too, as client 18427 in 1.60.1.70753 still has it).
func (rogue *Rogue) aggressionMultiplier() float64 {
	return []float64{1, 1.05, 1.10, 1.15}[rogue.Talents.Aggression]
}

// Coup de Grace (Improved Eviscerate) raises Eviscerate, and Death Mark's damage and storage rate,
// by 7/13/20% (November 2 development notes).
func (rogue *Rogue) coupDeGraceMultiplier() float64 {
	return []float64{1, 1.07, 1.13, 1.20}[rogue.Talents.CoupDeGrace]
}

// Efficient Strikes (Improved Sinister Strike) takes 3/5 Energy off Sinister Strike and, under the
// November 2 development notes, Hemorrhage.
func (rogue *Rogue) efficientStrikes() float64 {
	return []float64{0, 3, 5}[rogue.Talents.EfficientStrikes]
}
