package priest

import (
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The talents follow the Forever client (1.60.1.70338): the per-rank values are the
// TraitDefinitionEffectPoints curves and the spells each talent touches are its spell
// class mask (SpellEffect), read on wago.tools. A few masks differ from the tooltips:
// Twin Disciplines and Inner Focus's crit leave out Shadow Word: Death, Mental Agility
// leaves out Mind Blast, Mind Flay and Shadow Word: Death, and Shadow Focus leaves out
// Starshards.
func (priest *Priest) ApplyTalents() {
	// Discipline
	priest.applyTwinDisciplines()
	priest.applyMentalAgility()
	priest.registerInnerFocus()
	priest.applyHolyPrecision()

	if priest.Talents.SilentResolve > 0 {
		threat := 1 - 0.1*float64(priest.Talents.SilentResolve)
		priest.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagPriest) && spell.SpellSchool.Matches(core.SpellSchoolHoly) {
				spell.ThreatMultiplier *= threat
			}
		})
	}

	priest.PseudoStats.SpiritRegenRateCasting = []float64{0.0, 0.17, 0.33, 0.5}[priest.Talents.Meditation]

	if priest.Talents.MentalStrength > 0 {
		priest.MultiplyStat(stats.Intellect, 1.0+0.03*float64(priest.Talents.MentalStrength))
	}

	// Holy
	priest.applyHolySpecialization()
	priest.applySearingLight()

	priest.PseudoStats.SchoolDamageTakenMultiplier.MultiplyMagicSchools(1 - 0.02*float64(priest.Talents.SpellWarding))

	if priest.Talents.SpiritualGuidance > 0 {
		// 1, 3, 5, 6 and 8% of Spirit as spell damage (the healing part is 5% a rank).
		priest.AddStatDependency(stats.Spirit, stats.SpellDamage, []float64{0, 0.01, 0.03, 0.05, 0.06, 0.08}[priest.Talents.SpiritualGuidance])
	}

	// Shadow
	priest.registerVampiricEmbraceSpell()
	priest.registerShadowform()
	priest.applyShadowAffinity()
	priest.applyShadowFocus()
	priest.applyImprovedMindFlay()
	priest.applyShadowWeaving()
	priest.applyDevouringContagion()
	priest.applyDarkness()
}

// spellIn reports whether the spell is one of the priest's spells with one of the codes.
func spellIn(spell *core.Spell, codes ...int32) bool {
	return spell.Flags.Matches(SpellFlagPriest) && slices.Contains(codes, spell.SpellCode)
}

// Twin Disciplines: 1% a rank more damage on the instant priest spells in its class mask.
// Of the damage spells the sim has that is Shadow Word: Pain and Devouring Plague (the
// periodic part) and Chastise.
func (priest *Priest) applyTwinDisciplines() {
	if priest.Talents.TwinDisciplines == 0 {
		return
	}

	bonus := 0.01 * float64(priest.Talents.TwinDisciplines)
	priest.OnSpellRegistered(func(spell *core.Spell) {
		if spellIn(spell, SpellCode_PriestShadowWordPain, SpellCode_PriestDevouringPlague, SpellCode_PriestChastise) {
			spell.DamageMultiplierAdditive += bonus
		}
	})
}

// Mental Agility: 3, 7 and 10% off the mana cost of Smite, Holy Fire and the instant
// spells in its mask (Shadow Word: Pain, Devouring Plague, Vampiric Embrace, Shadowform,
// Power Infusion, Chastise and others). Mind Blast, Mind Flay and Shadow Word: Death
// are not in it.
func (priest *Priest) applyMentalAgility() {
	if priest.Talents.MentalAgility == 0 {
		return
	}

	discount := []int32{0, 3, 7, 10}[priest.Talents.MentalAgility]
	priest.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Cost == nil || !spell.Flags.Matches(SpellFlagPriest) {
			return
		}
		if spellIn(spell, SpellCode_PriestShadowWordPain, SpellCode_PriestDevouringPlague, SpellCode_PriestSmite, SpellCode_PriestHolyFire, SpellCode_PriestChastise) ||
			slices.Contains([]int32{15286, 15473, 10060}, spell.SpellID) {
			spell.Cost.Multiplier -= discount
		}
	})
}

// Holy Precision: 6% a rank more hit for Smite, Holy Fire and Penance.
func (priest *Priest) applyHolyPrecision() {
	if priest.Talents.HolyPrecision == 0 {
		return
	}

	bonusHit := 6 * float64(priest.Talents.HolyPrecision) * core.SpellHitRatingPerHitChance
	priest.OnSpellRegistered(func(spell *core.Spell) {
		if spellIn(spell, SpellCode_PriestSmite, SpellCode_PriestHolyFire, SpellCode_PriestPenance) {
			spell.BonusHitRating += bonusHit
		}
	})
}

// Holy Specialization: 1% crit a rank for Smite, Holy Fire, Penance and Holy Nova.
func (priest *Priest) applyHolySpecialization() {
	if priest.Talents.HolySpecialization == 0 {
		return
	}

	priest.OnSpellRegistered(func(spell *core.Spell) {
		if spellIn(spell, SpellCode_PriestSmite, SpellCode_PriestHolyFire, SpellCode_PriestPenance) {
			spell.BonusCritRating += 1 * float64(priest.Talents.HolySpecialization) * core.CritRatingPerCritChance
		}
	})
}

// Searing Light: 2 and 5% more Holy damage (a school damage aura, so it multiplies). The
// free Holy Nova part is not simulated.
func (priest *Priest) applySearingLight() {
	if priest.Talents.SearingLight == 0 {
		return
	}

	priest.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexHoly] *= []float64{1, 1.02, 1.05}[priest.Talents.SearingLight]
}

// Shadow Affinity: 10% a rank less threat from the Shadow spells.
func (priest *Priest) applyShadowAffinity() {
	if priest.Talents.ShadowAffinity == 0 {
		return
	}

	priest.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Flags.Matches(SpellFlagPriest) && spell.SpellSchool.Matches(core.SpellSchoolShadow) {
			spell.ThreatMultiplier *= 1 - 0.1*float64(priest.Talents.ShadowAffinity)
		}
	})
}

// Shadow Focus: 1% hit a rank (Classic: 2% a rank) for the priest's Shadow spells.
func (priest *Priest) applyShadowFocus() {
	if priest.Talents.ShadowFocus == 0 {
		return
	}

	bonusHit := 1 * float64(priest.Talents.ShadowFocus) * core.SpellHitRatingPerHitChance
	priest.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Flags.Matches(SpellFlagPriest) && spell.SpellSchool.Matches(core.SpellSchoolShadow) {
			spell.BonusHitRating += bonusHit
		}
	})
}

// Improved Mind Flay: 10% a rank more Mind Flay damage. The range and the weaker slow
// change nothing in the sim.
func (priest *Priest) applyImprovedMindFlay() {
	if priest.Talents.ImprovedMindFlay == 0 {
		return
	}

	priest.OnSpellRegistered(func(spell *core.Spell) {
		if spellIn(spell, SpellCode_PriestMindFlay) {
			spell.DamageMultiplierAdditive += 0.1 * float64(priest.Talents.ImprovedMindFlay)
		}
	})
}

// Devouring Contagion: 25% a rank off Devouring Plague's mana cost. The jump on a kill
// does nothing against one boss.
func (priest *Priest) applyDevouringContagion() {
	if priest.Talents.DevouringContagion == 0 {
		return
	}

	priest.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Cost != nil && spellIn(spell, SpellCode_PriestDevouringPlague) {
			spell.Cost.Multiplier -= 25 * priest.Talents.DevouringContagion
		}
	})
}

// Shadow Weaving under Forever: each landed Shadow damage spell has a 33, 67 or 100%
// chance (the client's curve) to put a stack of 15258 on the target, up to 5. Each stack
// is 2% more Shadow damage from the priest who put it there (aura 270 on the target), so
// the sim gives it to the priest's own Shadow damage. That is the same thing against one
// target. Under Classic it is the 3% a stack debuff for everyone.
func (priest *Priest) applyShadowWeaving() {
	if priest.Talents.ShadowWeaving == 0 {
		return
	}

	if !priest.Env.IsForever() {
		rank := min(int(priest.Talents.ShadowWeaving), 5)
		priest.ShadowWeavingAuras = priest.NewEnemyAuraArray(func(unit *core.Unit) *core.Aura {
			return core.ShadowWeavingAura(unit, rank)
		})
	} else {
		priest.ShadowWeavingAuras = priest.NewEnemyAuraArray(func(unit *core.Unit) *core.Aura {
			return unit.GetOrRegisterAura(core.Aura{
				Label:     "Shadow Weaving (Forever) - " + priest.Label,
				ActionID:  core.ActionID{SpellID: 15258},
				Duration:  time.Second * 15,
				MaxStacks: 5,
				OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks int32, newStacks int32) {
					priest.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexShadow] *= (1.0 + 0.02*float64(newStacks)) / (1.0 + 0.02*float64(oldStacks))
				},
			})
		})
	}

	procChance := []float64{0, 0.33, 0.67, 1}[priest.Talents.ShadowWeaving]

	priest.ShadowWeavingProc = priest.GetOrRegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 15257},
		Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagNoMetrics,
		SpellSchool: core.SpellSchoolShadow,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if procChance == 1.0 || sim.RollWithLabel(0, 1, "ShadowWeaving") < procChance {
				aura := priest.ShadowWeavingAuras.Get(target)
				aura.Activate(sim)
				aura.AddStack(sim)
			}
		},
	})
}

// AddShadowWeavingStack rolls Shadow Weaving after a Shadow damage spell landed.
func (priest *Priest) AddShadowWeavingStack(sim *core.Simulation, target *core.Unit) {
	if priest.ShadowWeavingProc == nil {
		return
	}

	priest.ShadowWeavingProc.Cast(sim, target)
}

// Darkness: 2% a rank more Shadow damage. Under Forever it is a school damage aura (79),
// so it covers every Shadow damage the priest deals and multiplies with Shadowform.
func (priest *Priest) applyDarkness() {
	if priest.Talents.Darkness == 0 {
		return
	}

	priest.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexShadow] *= 1 + 0.02*float64(priest.Talents.Darkness)
}

// Inner Focus: the next spell is free and, if it is not periodic, gets 25% more crit.
// Its crit mask leaves out Shadow Word: Death.
func (priest *Priest) registerInnerFocus() {
	if !priest.Talents.InnerFocus {
		return
	}

	actionID := core.ActionID{SpellID: 14751}

	// Ticks roll crit live under Forever, so the bonus stays off the periodic spells and
	// channels, or their ticks would crit more while Inner Focus waits to be used. Holy Fire's
	// own ticks take the bonus back off (holy_fire.go).
	critSpell := func(spell *core.Spell) bool {
		return spell.Flags.Matches(SpellFlagPriest) && spell.SpellCode != SpellCode_PriestShadowWordDeath &&
			!spell.Flags.Matches(core.SpellFlagPureDot|core.SpellFlagChanneled)
	}

	priest.InnerFocusAura = priest.RegisterAura(core.Aura{
		Label:    "Inner Focus",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range priest.Spellbook {
				if spell.Flags.Matches(SpellFlagPriest) && spell.Cost != nil {
					spell.Cost.Multiplier -= 100
				}
				if critSpell(spell) {
					spell.BonusCritRating += 25 * core.SpellCritRatingPerCritChance
				}
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range priest.Spellbook {
				if spell.Flags.Matches(SpellFlagPriest) && spell.Cost != nil {
					spell.Cost.Multiplier += 100
				}
				if critSpell(spell) {
					spell.BonusCritRating -= 25 * core.SpellCritRatingPerCritChance
				}
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagPriest) && spell.Cost != nil {
				// Remove the buff and put skill on CD
				aura.Deactivate(sim)
				priest.InnerFocus.CD.Use(sim)
				priest.UpdateMajorCooldowns()
			}
		},
	})

	priest.InnerFocus = priest.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    priest.NewTimer(),
				Duration: time.Minute * 3,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			priest.InnerFocusAura.Activate(sim)
		},
	})
}

// Shadowform under Forever: 10% more Shadow damage (Classic 15%), half the mana cost of
// Shadow spells, double the crit bonus of the Shadow damage spells in its mask (Shadow
// Word: Pain, Mind Blast, Mind Flay, Devouring Plague, so crits deal 200%, but not Shadow
// Word: Death) and 15% less Physical damage taken. It costs 40% of base mana and blocks
// healing spells, not Smite or Holy Fire. The cost cut is a school power cost aura (72),
// which multiplies with the talents' percent cost mods (Devouring Contagion, Mental
// Agility) where those add up among themselves.
func (priest *Priest) registerShadowform() {
	if !priest.Talents.Shadowform {
		return
	}

	actionID := core.ActionID{SpellID: 15473}
	critSpell := func(spell *core.Spell) bool {
		return spellIn(spell, SpellCode_PriestShadowWordPain, SpellCode_PriestMindBlast, SpellCode_PriestMindFlay, SpellCode_PriestDevouringPlague)
	}
	costSpell := func(spell *core.Spell) bool {
		return spell.Cost != nil && spell.Flags.Matches(SpellFlagPriest) && spell.SpellSchool.Matches(core.SpellSchoolShadow)
	}

	priest.ShadowformAura = priest.RegisterAura(core.Aura{
		Label:    "Shadowform",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexShadow] *= 1.10
			aura.Unit.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical] *= 0.85
			for _, spell := range priest.Spellbook {
				if critSpell(spell) {
					spell.CritDamageBonus += 1
				}
				if costSpell(spell) {
					spell.Cost.Factor *= 0.5
				}
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexShadow] /= 1.10
			aura.Unit.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical] /= 0.85
			for _, spell := range priest.Spellbook {
				if critSpell(spell) {
					spell.CritDamageBonus -= 1
				}
				if costSpell(spell) {
					spell.Cost.Factor /= 0.5
				}
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.Flags.Matches(core.SpellFlagHelpful) && spell.ProcMask.Matches(core.ProcMaskSpellHealing) {
				aura.Deactivate(sim)
			}
		},
	})

	priest.Shadowform = priest.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    SpellFlagPriest | core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.4,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			priest.ShadowformAura.Activate(sim)
		},
	})
}
