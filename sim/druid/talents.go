package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The Forever talents (client 1.60.1.70338, wago.tools TraitDefinitionEffectPoints for the
// value of every rank). Each spell modifier names the spells it touches by class mask, and
// those are the spells listed with it here.
func (druid *Druid) ApplyTalents() {
	// Balance
	druid.registerMoonkinFormSpell()
	druid.applyOmenOfClarity()
	druid.applyImprovedWrath()
	druid.applyGenesis()
	druid.applyMoonglow()
	druid.applyImprovedMoonfire()
	druid.applyVengeance()
	druid.applyImprovedStarfire()
	druid.applyNaturesGrace()
	druid.applyEclipse()

	// Nature's Majesty: 2% crit a point with spells and attacks.
	druid.AddStats(stats.Stats{
		stats.MeleeCrit: 2 * float64(druid.Talents.NaturesMajesty) * core.CritRatingPerCritChance,
		stats.SpellCrit: 2 * float64(druid.Talents.NaturesMajesty) * core.SpellCritRatingPerCritChance,
	})
	// Nature's Reach: 2% hit a point with spells and attacks.
	druid.AddStats(stats.Stats{
		stats.MeleeHit: 2 * float64(druid.Talents.NaturesReach) * core.MeleeHitRatingPerHitChance,
		stats.SpellHit: 2 * float64(druid.Talents.NaturesReach) * core.SpellHitRatingPerHitChance,
	})
	// Moonfury: 2% a point more Arcane and Nature damage, a school damage aura, so it
	// multiplies with the spell modifiers.
	if druid.Talents.Moonfury > 0 {
		multiplier := 1 + 0.02*float64(druid.Talents.Moonfury)
		druid.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexArcane] *= multiplier
		druid.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexNature] *= multiplier
	}

	// Feral
	druid.applySavageFury()
	druid.applyBloodFrenzy()
	druid.applyNaturalInstinct()

	// Heart of the Wild: 2% Intellect a point (its Cat Form Strength is in forms.go).
	if druid.Talents.HeartOfTheWild > 0 {
		druid.MultiplyStat(stats.Intellect, 1+0.02*float64(druid.Talents.HeartOfTheWild))
	}
	// Natural Reaction: 1% dodge a point. Its Rage on a dodge is a bear's.
	druid.AddStat(stats.Dodge, float64(druid.Talents.NaturalReaction)*core.DodgeRatingPerDodgeChance)
	// Thick Hide: 1 base armor a level a point in the forms. The defense skill part is a
	// tank's, and the form armor multipliers are left out.
	if druid.Talents.ThickHide > 0 && druid.StartingForm.Matches(Cat|Bear|Moonkin) {
		druid.AddStat(stats.Armor, float64(druid.Talents.ThickHide)*float64(druid.Level))
	}

	// Restoration
	druid.applyFuror()

	// Naturalist: 1% a point more damage of every school.
	druid.PseudoStats.DamageDealtMultiplier *= 1 + 0.01*float64(druid.Talents.Naturalist)
	// Reflection: 17, 33 or 50% of mana regeneration goes on while casting.
	druid.PseudoStats.SpiritRegenRateCasting += []float64{0, 0.17, 0.33, 0.5}[druid.Talents.Reflection]
	// Living Spirit: 5% Spirit a point.
	if druid.Talents.LivingSpirit > 0 {
		druid.MultiplyStat(stats.Spirit, 1+0.05*float64(druid.Talents.LivingSpirit))
	}
}

// Spells whose rank list is kept per rank. Moonfire's list holds only the ranks the level
// knows, the others hold nil for the rest.
func (druid *Druid) balanceSpells(lists ...[]*DruidSpell) []*DruidSpell {
	return core.FilterSlice(core.Flatten(lists), func(spell *DruidSpell) bool { return spell != nil })
}

// Wrath, Moonfire, Starfire, Insect Swarm and Hurricane, the damage spells most Balance
// modifiers name.
func (druid *Druid) damageSpells() []*DruidSpell {
	return druid.balanceSpells(druid.Wrath, druid.Moonfire, druid.Starfire, druid.InsectSwarm, druid.Hurricane)
}

func (druid *Druid) onInit(label string, apply func()) {
	druid.RegisterAura(core.Aura{
		Label: label,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			apply()
		},
	})
}

// Improved Wrath: 0.1 sec a point off Wrath's cast and 10% a point off its cost.
func (druid *Druid) applyImprovedWrath() {
	points := druid.Talents.ImprovedWrath
	if points == 0 {
		return
	}
	druid.onInit("Improved Wrath", func() {
		for _, spell := range druid.balanceSpells(druid.Wrath) {
			spell.DefaultCast.CastTime -= time.Millisecond * 100 * time.Duration(points)
			spell.Cost.Factor *= 1 - 0.1*float64(points)
		}
	})
}

// Genesis: 1% a point more periodic damage from Moonfire, Insect Swarm, Rake and Rip (and
// healing over time, which the sim does not use).
func (druid *Druid) applyGenesis() {
	points := druid.Talents.Genesis
	if points == 0 {
		return
	}
	druid.onInit("Genesis", func() {
		spells := druid.balanceSpells(druid.Moonfire, druid.InsectSwarm, []*DruidSpell{druid.Rake, druid.Rip})
		for _, spell := range spells {
			spell.PeriodicDamageMultiplierAdditive += 0.01 * float64(points)
		}
	})
}

// Moonglow: 8, 17 or 25% off the cost of the damage spells. The client's cost modifiers
// multiply each other (shaman_audit.md 8.1), so it multiplies with Improved Wrath.
func (druid *Druid) applyMoonglow() {
	points := druid.Talents.Moonglow
	if points == 0 {
		return
	}
	factor := 1 - []float64{0, 0.08, 0.17, 0.25}[points]
	druid.onInit("Moonglow", func() {
		for _, spell := range druid.damageSpells() {
			spell.Cost.Factor *= factor
		}
	})
}

// Improved Moonfire: 5% a point more crit chance, hit damage and periodic damage on
// Moonfire.
func (druid *Druid) applyImprovedMoonfire() {
	points := float64(druid.Talents.ImprovedMoonfire)
	if points == 0 {
		return
	}
	druid.onInit("Improved Moonfire", func() {
		for _, spell := range druid.balanceSpells(druid.Moonfire) {
			spell.BonusCritRating += 5 * points * core.SpellCritRatingPerCritChance
			spell.ImpactDamageMultiplierAdditive += 0.05 * points
			spell.PeriodicDamageMultiplierAdditive += 0.05 * points
		}
	})
}

// Vengeance: 20% a point more crit damage bonus on the Arcane and Nature damage spells.
func (druid *Druid) applyVengeance() {
	points := float64(druid.Talents.Vengeance)
	if points == 0 {
		return
	}
	druid.onInit("Vengeance", func() {
		for _, spell := range druid.damageSpells() {
			spell.CritDamageBonus += 0.2 * points
		}
	})
}

// Improved Starfire: 0.1 sec a point off Starfire's cast. Its stun does nothing to a boss.
func (druid *Druid) applyImprovedStarfire() {
	points := druid.Talents.ImprovedStarfire
	if points == 0 {
		return
	}
	druid.onInit("Improved Starfire", func() {
		for _, spell := range druid.balanceSpells(druid.Starfire) {
			spell.DefaultCast.CastTime -= time.Millisecond * 100 * time.Duration(points)
		}
	})
}

// Nature's Grace under Forever (16880, 16886): a spell crit that is not periodic gives 10%
// cast speed and 10% off the global cooldown of the Balance spells for 3 sec. 1.12 took 0.5
// sec off the next cast instead.
func (druid *Druid) applyNaturesGrace() {
	if !druid.Talents.NaturesGrace {
		return
	}

	var affectedSpells []*DruidSpell
	gcdCut := core.GCDDefault / 10
	druid.NaturesGraceProcAura = druid.RegisterAura(core.Aura{
		Label:    "Natures Grace Proc",
		ActionID: core.ActionID{SpellID: 16886},
		Duration: time.Second * 3,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells = druid.damageSpells()
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			druid.MultiplyCastSpeed(1.1)
			for _, spell := range affectedSpells {
				spell.DefaultCast.GCD -= gcdCut
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.MultiplyCastSpeed(1 / 1.1)
			for _, spell := range affectedSpells {
				spell.DefaultCast.GCD += gcdCut
			}
		},
	})

	core.MakePermanent(druid.RegisterAura(core.Aura{
		Label: "Natures Grace",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			// Spells with travel times have their own implementation because the proc occurs as the cast finishes
			if spell.MissileSpeed == 0 && spell.ProcMask.Matches(core.ProcMaskSpellDamage) && result.DidCrit() {
				druid.NaturesGraceProcAura.Activate(sim)
			}
		},
	}))
}

// Eclipse (408248, 408255): each Wrath that lands gives 2 charges, at most 4, for 15 sec.
// Each charge takes 0.17, 0.33 or 0.5 sec off one Starfire's cast.
func (druid *Druid) applyEclipse() {
	points := druid.Talents.Eclipse
	if points == 0 {
		return
	}
	cut := time.Millisecond * time.Duration([]int{0, 170, 330, 500}[points])

	var starfires []*DruidSpell
	druid.EclipseAura = druid.RegisterAura(core.Aura{
		Label:     "Eclipse",
		ActionID:  core.ActionID{SpellID: 408255},
		Duration:  time.Second * 15,
		MaxStacks: 4,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			starfires = druid.balanceSpells(druid.Starfire)
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range starfires {
				spell.DefaultCast.CastTime -= cut
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range starfires {
				spell.DefaultCast.CastTime += cut
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			// A charge goes with the Starfire it made shorter. One that started before the
			// charges came (a Wrath landing mid-cast) keeps them for the next.
			if spell.SpellCode == SpellCode_DruidStarfire && sim.CurrentTime-spell.CurCast.CastTime >= aura.StartedAt() {
				aura.RemoveStack(sim)
			}
		},
	})

	core.MakePermanent(druid.RegisterAura(core.Aura{
		Label: "Eclipse Trigger",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.SpellCode == SpellCode_DruidWrath && result.Landed() {
				druid.EclipseAura.Activate(sim)
				druid.EclipseAura.SetStacks(sim, min(druid.EclipseAura.MaxStacks, druid.EclipseAura.GetStacks()+2))
			}
		},
	}))
}

// Savage Fury: 5% a point more damage from Claw, Rake and Shred, and from Rake's bleed (two
// modifiers in the client, one for the hits and one for the periodic part).
func (druid *Druid) applySavageFury() {
	points := float64(druid.Talents.SavageFury)
	if points == 0 {
		return
	}
	druid.onInit("Savage Fury", func() {
		for _, spell := range []*DruidSpell{druid.Claw, druid.Rake, druid.Shred} {
			if spell != nil {
				spell.ImpactDamageMultiplierAdditive += 0.05 * points
			}
		}
		if druid.Rake != nil {
			druid.Rake.PeriodicDamageMultiplierAdditive += 0.05 * points
		}
	})
}

// Natural Instinct: 10% a point more crit damage bonus on the melee abilities (its healing
// part does nothing here).
func (druid *Druid) applyNaturalInstinct() {
	points := float64(druid.Talents.NaturalInstinct)
	if points == 0 {
		return
	}
	druid.onInit("Natural Instinct", func() {
		for _, spell := range []*DruidSpell{druid.Claw, druid.Rake, druid.Shred, druid.FerociousBite} {
			if spell != nil {
				spell.CritDamageBonus += 0.1 * points
			}
		}
	})
}

// Rend and Tear: 2% a point more damage from the melee abilities on a bleeding target. The
// effect is a dummy the server runs. The sim takes it as a multiplier on the hits of Claw,
// Shred, Rake and Ferocious Bite, not on bleeds, while our Rake or Rip is up or the spec
// assumes another bleed on the target.
func (druid *Druid) rendAndTearMultiplier(target *core.Unit) float64 {
	if druid.Talents.RendAndTear == 0 {
		return 1
	}
	bleeding := druid.AssumeBleedActive ||
		(druid.Rake != nil && druid.Rake.Dot(target).IsActive()) ||
		(druid.Rip != nil && druid.Rip.Dot(target).IsActive())
	return core.TernaryFloat64(bleeding, 1+0.02*float64(druid.Talents.RendAndTear), 1)
}

// Blood Frenzy under Forever (16958): a crit from a combo point builder has a 50% chance a
// point to give a second combo point (1.12's Blood Frenzy was a haste buff, this is 1.12's
// Primal Fury cat half).
func (druid *Druid) applyBloodFrenzy() {
	if druid.Talents.BloodFrenzy == 0 {
		return
	}

	procChance := 0.5 * float64(druid.Talents.BloodFrenzy)
	actionID := core.ActionID{SpellID: 16958}
	cpMetrics := druid.NewComboPointMetrics(actionID)

	core.MakePermanent(druid.RegisterAura(core.Aura{
		Label: "Blood Frenzy",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if druid.InForm(Cat) &&
				result.Target == aura.Unit.CurrentTarget &&
				spell.Flags.Matches(SpellFlagBuilder) &&
				result.Outcome.Matches(core.OutcomeCrit) &&
				sim.Proc(procChance, "Blood Frenzy") {
				druid.AddComboPoints(sim, 1, result.Target, cpMetrics)
			}
		},
	}))
}

// We're using an aura so that the APL can know if the Druid has furor
func (druid *Druid) applyFuror() {
	if druid.Talents.Furor == 0 {
		return
	}

	spellID := []int32{0, 17056, 17058, 17059, 17060, 17061}[druid.Talents.Furor]

	druid.FurorAura = druid.RegisterAura(core.Aura{
		Label:    "Furor",
		ActionID: core.ActionID{SpellID: spellID},
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
	})
}

// Omen of Clarity is a baseline passive under Forever, learned at 20 (16864): spells and
// attacks give Clearcasting (16870), so the next damage or healing spell or ability other
// than Wrath costs nothing, with a 10 sec cooldown. Moonkin Form doubles the chance and
// halves the cooldown (its spell modifiers on Omen of Clarity, +100% chance and -50%).
//
// The client leaves the chance to the server. The sim keeps 1.12's 2 procs a minute on
// melee hits (weapon speed based), and gives a damage spell that lands the same rate over
// its cast time, 2 * max(cast time, 1.5 sec) / 60 a cast, which is our own guess.
func (druid *Druid) applyOmenOfClarity() {
	if druid.Level < 20 {
		return
	}

	var affectedSpells []*core.Spell
	druid.ClearcastingAura = druid.RegisterAura(core.Aura{
		Label:    "Clearcasting",
		ActionID: core.ActionID{SpellID: 16870},
		Duration: time.Second * 15,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells = core.FilterSlice(druid.Spellbook, func(spell *core.Spell) bool {
				return spell.Flags.Matches(SpellFlagOmen) && spell.SpellCode != SpellCode_DruidWrath
			})
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedSpells {
				spell.Cost.Multiplier -= 100
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedSpells {
				spell.Cost.Multiplier += 100
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			// OnCastComplete is called after OnSpellHitDealt / etc, so don't deactivate if it was just activated.
			if aura.RemainingDuration(sim) == aura.Duration {
				return
			}

			if spell.Flags.Matches(SpellFlagOmen) && spell.SpellCode != SpellCode_DruidWrath && spell.DefaultCast.Cost > 0 {
				aura.Deactivate(sim)
			}
		},
	})

	ppmm := druid.AutoAttacks.NewPPMManager(2.0, core.ProcMaskMelee)
	icd := core.Cooldown{
		Timer:    druid.NewTimer(),
		Duration: time.Second * 10,
	}
	inMoonkin := func() bool { return druid.MoonkinFormAura != nil && druid.MoonkinFormAura.IsActive() }

	druid.RegisterAura(core.Aura{
		Label:    "Omen of Clarity",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !icd.IsReady(sim) {
				return
			}
			procced := false
			if spell.ProcMask.Matches(core.ProcMaskMelee) {
				procced = ppmm.ProcWithWeaponSpecials(sim, spell.ProcMask, "Omen of Clarity")
			} else if spell.ProcMask.Matches(core.ProcMaskSpellDamage) && spell.Flags.Matches(SpellFlagOmen) {
				chance := 2 * max(spell.DefaultCast.CastTime, core.GCDDefault).Seconds() / 60
				if inMoonkin() {
					chance *= 2
				}
				procced = sim.Proc(chance, "Omen of Clarity")
			}
			if procced {
				icd.Use(sim)
				if inMoonkin() {
					icd.Timer.Set(sim.CurrentTime + icd.Duration/2)
				}
				druid.ClearcastingAura.Activate(sim)
			}
		},
	})
}
