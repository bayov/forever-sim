package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Per rank values are the client's rank curves (wago.tools TraitDefinitionEffectPoints, build
// 1.60.1.70338), and the spells a talent touches are the ones its class mask names.
func (hunter *Hunter) ApplyTalents() {
	hunter.applyBeastMasteryTalents()
	hunter.applyMarksmanshipTalents()
	hunter.applySurvivalTalents()
}

func (hunter *Hunter) applyBeastMasteryTalents() {
	// Deadly Aspects: aspects.go. Summon Hawk: summon_hawk.go.
	if hunter.pet != nil {
		// Endurance Training: 3% pet health a rank (and armor, which nothing hits here).
		if hunter.Talents.EnduranceTraining > 0 {
			hunter.pet.MultiplyStat(stats.Health, 1+0.03*float64(hunter.Talents.EnduranceTraining))
		}

		// Focused Fire: 1% a rank on everything the hunter and the pet deal while the pet is out.
		if hunter.Talents.FocusedFire > 0 {
			multiplier := 1 + 0.01*float64(hunter.Talents.FocusedFire)
			hunter.PseudoStats.DamageDealtMultiplier *= multiplier
			hunter.pet.PseudoStats.DamageDealtMultiplier *= multiplier
		}

		// Unleashed Fury: 3% pet damage a rank (Classic 4%). The hawks get it in summon_hawk.go.
		hunter.pet.PseudoStats.DamageDealtMultiplier *= 1 + 0.03*float64(hunter.Talents.UnleashedFury)

		// Ferocity: 2% pet crit a rank (Classic 3%).
		hunter.pet.AddStat(stats.MeleeCrit, core.CritRatingPerCritChance*2*float64(hunter.Talents.Ferocity))
		hunter.pet.AddStat(stats.SpellCrit, core.SpellCritRatingPerCritChance*2*float64(hunter.Talents.Ferocity))

		hunter.applyFrenzy()
		hunter.registerIntimidation()
		hunter.registerBestialWrathCD()
	}

	// Bestial Discipline: the pet's 10% focus regen a rank is in pet.go. Forever adds 25% of the
	// hunter's mana regen while casting a rank.
	hunter.PseudoStats.SpiritRegenRateCasting += 0.25 * float64(hunter.Talents.BestialDiscipline)
}

func (hunter *Hunter) applyMarksmanshipTalents() {
	// Lethal Attacks: 1% crit a rank with all attacks (Classic's Lethal Shots was ranged only).
	if hunter.Talents.LethalAttacks > 0 {
		hunter.AddStat(stats.MeleeCrit, float64(hunter.Talents.LethalAttacks)*core.CritRatingPerCritChance)
		hunter.AddStat(stats.SpellCrit, float64(hunter.Talents.LethalAttacks)*core.SpellCritRatingPerCritChance)
	}

	// Efficiency: 3% off a rank (Classic 2%) on the shots, stings and melee abilities in its
	// class mask. Sniper Shot and Strider Kick are not in it.
	if hunter.Talents.Efficiency > 0 {
		hunter.OnSpellRegistered(func(spell *core.Spell) {
			switch spell.SpellCode {
			case SpellCode_HunterAimedShot, SpellCode_HunterArcaneShot, SpellCode_HunterMultiShot,
				SpellCode_HunterSerpentSting, SpellCode_HunterVolley, SpellCode_HunterSummonHawk,
				SpellCode_HunterRaptorStrike, SpellCode_HunterMongooseBite, SpellCode_HunterWingClip,
				SpellCode_HunterLacerate:
				if spell.Cost != nil {
					spell.Cost.Multiplier -= 3 * hunter.Talents.Efficiency
				}
			}
		})
	}

	// Careful Aim: 20% of Intellect a rank as attack power, melee and ranged (1223984 has one
	// effect for each).
	if hunter.Talents.CarefulAim > 0 {
		apPerInt := 0.2 * float64(hunter.Talents.CarefulAim)
		hunter.AddStatDependency(stats.Intellect, stats.AttackPower, apPerInt)
		hunter.AddStatDependency(stats.Intellect, stats.RangedAttackPower, apPerInt)
	}

	// Lone Wolf: 20% damage done while no pet is out.
	if hunter.Talents.LoneWolf && hunter.pet == nil {
		hunter.PseudoStats.DamageDealtMultiplier *= 1.2
	}

	// Rapid Killing: rapid_fire.go. Improved Stings: serpent_sting.go. Improved Arcane Shot:
	// arcane_shot.go. Barrage: aimed_shot.go, multi_shot.go and volley.go. Trueshot Aura: hunter.go.
	// Sniper Shot: sniper_shot.go.

	if hunter.Talents.RangedWeaponSpecialization > 0 {
		mult := 1 + 0.01*float64(hunter.Talents.RangedWeaponSpecialization)
		hunter.OnSpellRegistered(func(spell *core.Spell) {
			if spell.ProcMask.Matches(core.ProcMaskRanged) && spell.SpellCode != SpellCode_HunterSerpentSting {
				spell.DamageMultiplier *= mult
			}
		})
	}

	hunter.applyRapidRecuperation()
}

func (hunter *Hunter) applySurvivalTalents() {
	// Improved Tracking: 1% damage a rank against the creature type the hunter tracks. A hunter
	// can track every type a raid boss is (Beasts, Demons, Dragonkin, Elementals, Giants,
	// Humanoids, Undead), so it counts against all of them. Damage only, like Classic's Slaying
	// talents' damage half, but the crit damage half is gone. The pet does not get it.
	if hunter.Talents.ImprovedTracking > 0 {
		multiplier := 1 + 0.01*float64(hunter.Talents.ImprovedTracking)
		hunter.Env.RegisterPostFinalizeEffect(func() {
			for _, t := range hunter.Env.Encounter.Targets {
				switch t.MobType {
				case proto.MobType_MobTypeBeast, proto.MobType_MobTypeDemon, proto.MobType_MobTypeDragonkin,
					proto.MobType_MobTypeElemental, proto.MobType_MobTypeGiant, proto.MobType_MobTypeHumanoid,
					proto.MobType_MobTypeUndead:
					for _, at := range hunter.AttackTables[t.UnitIndex] {
						at.DamageDealtMultiplier *= multiplier
					}
				}
			}
		})
	}

	// Savage Strikes: 2% crit a rank (Classic 10% on Raptor Strike and Mongoose Bite) on the melee
	// abilities in its mask: Raptor Strike, Mongoose Bite, Wing Clip, Strider Kick, Lacerate and the
	// Lacerating Strikes bleed. Predator's Edge: 6% crit damage a rank on the same ones, and 10% off
	// hand damage a rank.
	if hunter.Talents.SavageStrikes > 0 || hunter.Talents.PredatorsEdge > 0 {
		bonusCrit := 2 * float64(hunter.Talents.SavageStrikes) * core.CritRatingPerCritChance
		critDamage := 0.06 * float64(hunter.Talents.PredatorsEdge)
		hunter.OnSpellRegistered(func(spell *core.Spell) {
			if hunter.isMeleeAbility(spell) {
				spell.BonusCritRating += bonusCrit
				spell.CritDamageBonus += critDamage
			}
		})
	}
	if hunter.Talents.PredatorsEdge > 0 && hunter.AutoAttacks.IsDualWielding {
		hunter.AutoAttacks.OHConfig().DamageMultiplier *= 1 + 0.1*float64(hunter.Talents.PredatorsEdge)
	}

	if hunter.Talents.Survivalist > 0 {
		hunter.MultiplyStat(stats.Health, 1.0+0.02*float64(hunter.Talents.Survivalist))
	}

	// Surefooted: 1% hit a rank on every attack (Classic: melee only).
	hunter.AddStat(stats.MeleeHit, float64(hunter.Talents.Surefooted)*core.MeleeHitRatingPerHitChance)
	hunter.AddStat(stats.SpellHit, float64(hunter.Talents.Surefooted)*core.SpellHitRatingPerHitChance)

	// Clever Traps: 15% trap damage a rank. Survival Tactics: 5% trap hit a rank (Classic's Trap
	// Mastery). Survivalist's Discipline: 20% off the trap cooldowns a rank.
	hunter.OnSpellRegistered(func(spell *core.Spell) {
		if !spell.Flags.Matches(SpellFlagTrap) {
			return
		}
		spell.DamageMultiplier *= 1 + 0.15*float64(hunter.Talents.CleverTraps)
		spell.BonusHitRating += 5 * float64(hunter.Talents.SurvivalTactics) * core.SpellHitRatingPerHitChance
		if spell.CD.Timer != nil {
			spell.CD.Duration = time.Duration(float64(spell.CD.Duration) * (1 - 0.2*float64(hunter.Talents.SurvivalistsDiscipline)))
		}
	})

	// Lightning Reflexes: 2% Agility a rank (Classic 3%). The client parks a second Lightning
	// Reflexes at 3% a rank off the canvas, where no one can take it.
	if hunter.Talents.LightningReflexes > 0 {
		hunter.MultiplyStat(stats.Agility, 1.0+0.02*float64(hunter.Talents.LightningReflexes))
	}

	hunter.applyResourcefulness()
	hunter.applyExposePrey()
	// Strider Kick: strider_kick.go. Lacerating Strikes: mongoose_bite.go.
}

// The melee abilities in Savage Strikes' and Predator's Edge's class masks.
func (hunter *Hunter) isMeleeAbility(spell *core.Spell) bool {
	switch spell.SpellCode {
	case SpellCode_HunterRaptorStrikeHit, SpellCode_HunterMongooseBite, SpellCode_HunterWingClip,
		SpellCode_HunterStriderKick, SpellCode_HunterLacerate, SpellCode_HunterLaceratingStrikes:
		return true
	}
	return false
}

func (hunter *Hunter) applyFrenzy() {
	if hunter.Talents.Frenzy == 0 {
		return
	}

	procChance := 0.2 * float64(hunter.Talents.Frenzy)

	procAura := hunter.pet.RegisterAura(core.Aura{
		Label:    "Frenzy Proc",
		ActionID: core.ActionID{SpellID: 19625},
		Duration: time.Second * 8,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.MultiplyAttackSpeed(sim, 1.3)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.MultiplyAttackSpeed(sim, 1/1.3)
		},
	})

	hunter.pet.RegisterAura(core.Aura{
		Label:    "Frenzy",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, spellResult *core.SpellResult) {
			if !spellResult.Outcome.Matches(core.OutcomeCrit) {
				return
			}
			if procChance == 1 || sim.RandomFloat("Frenzy") < procChance {
				procAura.Activate(sim)
			}
		},
	})
}

// Intimidation gives the pet's next attack 100% more crit chance (Forever adds that to the 3 sec
// stun, which a boss ignores). 8% of base mana, 1 min cooldown. Not a major cooldown: the
// rotation casts it if it names it.
func (hunter *Hunter) registerIntimidation() {
	if !hunter.Talents.Intimidation {
		return
	}

	actionID := core.ActionID{SpellID: 19577}
	bonusCrit := 100 * float64(core.CritRatingPerCritChance)

	petAura := hunter.pet.RegisterAura(core.Aura{
		Label:    "Intimidation",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatDynamic(sim, stats.MeleeCrit, bonusCrit)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatDynamic(sim, stats.MeleeCrit, -bonusCrit)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Landed() && spell.ProcMask.Matches(core.ProcMaskMelee) {
				aura.Deactivate(sim)
			}
		},
	})

	hunter.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.08,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Minute,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.pet.IsEnabled()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			petAura.Activate(sim)
		},
	})
}

func (hunter *Hunter) registerBestialWrathCD() {
	if !hunter.Talents.BestialWrath {
		return
	}

	actionID := core.ActionID{SpellID: 19574}

	hunter.BestialWrathPetAura = hunter.pet.RegisterAura(core.Aura{
		Label:    "Bestial Wrath Pet",
		ActionID: actionID,
		Duration: time.Second * 18,
	}).AttachMultiplicativePseudoStatBuff(&hunter.pet.PseudoStats.DamageDealtMultiplier, 1.5)

	bwSpell := hunter.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.12,
		},

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Minute * 2,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.pet.IsEnabled()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			hunter.BestialWrathPetAura.Activate(sim)
		},
	})

	hunter.AddMajorCooldown(core.MajorCooldown{
		Spell: bwSpell,
		Type:  core.CooldownTypeDPS,
	})
}

// Mortal Shots: 6% crit damage a rank (Classic also 6%) on Auto Shot, Aimed, Arcane and Multi-Shot,
// Serpent Sting and Volley. Not Sniper Shot or the hawks.
func (hunter *Hunter) mortalShots() float64 {
	return 0.06 * float64(hunter.Talents.MortalShots)
}

// Barrage: 3/7/10% on Aimed Shot, Multi-Shot and Volley (Classic 5/10/15% on Multi-Shot and Volley).
func (hunter *Hunter) barrageMultiplier() float64 {
	return []float64{1, 1.03, 1.07, 1.10}[hunter.Talents.Barrage]
}

// Rapid Recuperation: a landed Serpent Sting gives 25% of mana regen while casting a rank for 15
// sec (1242512). The Rapid Killing half needs a kill, which a boss fight doesn't have.
func (hunter *Hunter) applyRapidRecuperation() {
	if hunter.Talents.RapidRecuperation == 0 {
		return
	}

	regen := 0.25 * float64(hunter.Talents.RapidRecuperation)
	procAura := hunter.RegisterAura(core.Aura{
		Label:    "Rapid Recuperation",
		ActionID: core.ActionID{SpellID: 1242512},
		Duration: time.Second * 15,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			hunter.PseudoStats.SpiritRegenRateCasting += regen
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			hunter.PseudoStats.SpiritRegenRateCasting -= regen
		},
	})

	core.MakePermanent(hunter.RegisterAura(core.Aura{
		Label: "Rapid Recuperation Trigger",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.SpellCode == SpellCode_HunterSerpentSting && result.Landed() {
				procAura.Activate(sim)
			}
		},
	}))
}

// Resourcefulness takes 30% a rank off the traps and the melee abilities in its mask (Raptor
// Strike, Mongoose Bite, Wing Clip and the parry strike; not Strider Kick or Lacerate), and a crit has a
// 50% chance a rank to give 50% of mana regen while casting for 30 sec (1242688). The tooltip
// leaves the chance as "?%", the 50 a rank is the client's rank curve.
func (hunter *Hunter) applyResourcefulness() {
	if hunter.Talents.Resourcefulness == 0 {
		return
	}

	hunter.OnSpellRegistered(func(spell *core.Spell) {
		switch spell.SpellCode {
		case SpellCode_HunterRaptorStrike, SpellCode_HunterMongooseBite, SpellCode_HunterWingClip:
		default:
			if !spell.Flags.Matches(SpellFlagTrap) {
				return
			}
		}
		if spell.Cost != nil {
			spell.Cost.Multiplier -= 30 * hunter.Talents.Resourcefulness
		}
	})

	procAura := hunter.RegisterAura(core.Aura{
		Label:    "Resourcefulness",
		ActionID: core.ActionID{SpellID: 1242688},
		Duration: time.Second * 30,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			hunter.PseudoStats.SpiritRegenRateCasting += 0.5
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			hunter.PseudoStats.SpiritRegenRateCasting -= 0.5
		},
	})

	core.MakeProcTriggerAura(&hunter.Unit, core.ProcTrigger{
		Name:       "Resourcefulness Trigger",
		Callback:   core.CallbackOnSpellHitDealt,
		ProcMask:   core.ProcMaskMeleeOrRanged,
		Outcome:    core.OutcomeCrit,
		ProcChance: 0.5 * float64(hunter.Talents.Resourcefulness),
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			procAura.Activate(sim)
		},
	})
}

// Expose Prey: a landed attack on a target with Hunter's Mark has a 5% chance a rank to open
// Mongoose Bite for 10 sec (1310726, 5 sec before the 2026-10-08 build). The client's proc mask
// covers melee and ranged attacks, white and special.
func (hunter *Hunter) applyExposePrey() {
	if hunter.Talents.ExposePrey == 0 {
		return
	}

	core.MakeProcTriggerAura(&hunter.Unit, core.ProcTrigger{
		Name:       "Expose Prey",
		Callback:   core.CallbackOnSpellHitDealt,
		ProcMask:   core.ProcMaskMeleeOrRanged,
		Outcome:    core.OutcomeLanded,
		ProcChance: 0.05 * float64(hunter.Talents.ExposePrey),
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Target.HasActiveAuraWithTag(core.HuntersMarkAuraTag) {
				hunter.DefensiveState.Duration = time.Second * 10
				hunter.DefensiveState.Activate(sim)
			}
		},
	})
}
