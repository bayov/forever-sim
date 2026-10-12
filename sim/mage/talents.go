package mage

import (
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The talents follow the Forever beta client's tree (tools/forever_talents/client/mage.json,
// build 1.60.1.69876 with the Fire changes up to 1.60.1.70338). The ones that change nothing in a
// single target fight against a raid boss are not simulated, and the talent picker marks them.
// A raid boss can't be frozen, so the target only counts as Frozen through Fingers of Frost.
func (mage *Mage) ApplyTalents() {
	mage.applyArcaneTalents()
	mage.applyFireTalents()
	mage.applyFrostTalents()
}

// Adds to the spells of the given schools that the mage casts.
func (mage *Mage) onMageSpellOfSchool(school core.SpellSchool, f func(spell *core.Spell)) {
	mage.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Flags.Matches(SpellFlagMage) && spell.SpellSchool.Matches(school) {
			f(spell)
		}
	})
}

func (mage *Mage) applyArcaneTalents() {
	mage.applyArcaneConcentration()
	mage.applyMissileBarrage()
	mage.registerPresenceOfMindCD()
	mage.registerArcanePowerCD()

	// Arcane Subtlety: 8 / 15 less resistance to all the mage's spells, and 15 / 30% less
	// threat from Arcane spells.
	if mage.Talents.ArcaneSubtlety > 0 {
		mage.AddStat(stats.SpellPenetration, []float64{0, 8, 15}[mage.Talents.ArcaneSubtlety])
		threatMultiplier := 1 - .15*float64(mage.Talents.ArcaneSubtlety)
		mage.onMageSpellOfSchool(core.SpellSchoolArcane, func(spell *core.Spell) {
			spell.ThreatMultiplier *= threatMultiplier
		})
	}

	// Arcane Focus: 1% hit with Arcane spells per point (Classic took 2% off the resist chance).
	if mage.Talents.ArcaneFocus > 0 {
		bonusHit := float64(mage.Talents.ArcaneFocus) * core.SpellHitRatingPerHitChance
		mage.onMageSpellOfSchool(core.SpellSchoolArcane, func(spell *core.Spell) {
			spell.BonusHitRating += bonusHit
		})
	}

	// Magic Absorption: 5 resistance to every school per point. The mana back from fully
	// resisted spells needs an enemy that casts.
	if mage.Talents.MagicAbsorption > 0 {
		mage.AddResistances(5 * float64(mage.Talents.MagicAbsorption))
	}

	// Arcane Resilience: armor equal to 25% of Intellect per point.
	if mage.Talents.ArcaneResilience > 0 {
		mage.AddStatDependency(stats.Intellect, stats.BonusArmor, 0.25*float64(mage.Talents.ArcaneResilience))
	}

	// 2% crit with Arcane spells per point (Classic's was Arcane Explosion only).
	if mage.Talents.ArcaneImpact > 0 {
		bonusCrit := 2 * float64(mage.Talents.ArcaneImpact) * core.SpellCritRatingPerCritChance
		mage.onMageSpellOfSchool(core.SpellSchoolArcane, func(spell *core.Spell) {
			spell.BonusCritRating += bonusCrit
		})
	}

	// Arcane Meditation: 17 / 33 / 50% of mana regeneration goes on while casting (Classic 5 / 10 / 15%).
	mage.PseudoStats.SpiritRegenRateCasting += []float64{0, 0.17, 0.33, 0.50}[mage.Talents.ArcaneMeditation]

	// Arcane Mind: 2% Intellect per point (Classic gave max mana) and 20% more critical strike
	// damage bonus on Arcane spells per point, the same kind of bonus as Ice Shards.
	if mage.Talents.ArcaneMind > 0 {
		mage.MultiplyStat(stats.Intellect, 1.0+0.02*float64(mage.Talents.ArcaneMind))
		critBonus := 0.20 * float64(mage.Talents.ArcaneMind)
		mage.onMageSpellOfSchool(core.SpellSchoolArcane, func(spell *core.Spell) {
			spell.CritDamageBonus += critBonus
		})
	}

	// Arcane Instability: 1% damage and 1% crit with every spell per point.
	if mage.Talents.ArcaneInstability > 0 {
		bonusDamageMultiplierAdditive := .01 * float64(mage.Talents.ArcaneInstability)
		bonusCritRating := 1 * float64(mage.Talents.ArcaneInstability) * core.SpellCritRatingPerCritChance

		mage.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagMage) {
				spell.DamageMultiplierAdditive += bonusDamageMultiplierAdditive
				spell.BonusCritRating += bonusCritRating
			}
		})
	}
}

func (mage *Mage) applyFireTalents() {
	mage.applyIgnite()
	mage.applyImprovedScorch()
	mage.applyMasterOfElements()
	mage.applyHeatingUp()

	mage.registerCombustionCD()

	// Burning Soul: 10% less threat from Fire spells per point. Its pushback protection needs an
	// enemy that hits the mage.
	if mage.Talents.BurningSoul > 0 {
		threatMultiplier := 1 - .10*float64(mage.Talents.BurningSoul)
		mage.onMageSpellOfSchool(core.SpellSchoolFire, func(spell *core.Spell) {
			spell.ThreatMultiplier *= threatMultiplier
		})
	}

	// Incineration: 2% crit per point on Fire Blast, Scorch, Ice Lance and Arcane Blast (Forever
	// added the last two and a third rank).
	if mage.Talents.Incineration > 0 {
		bonusCrit := 2 * float64(mage.Talents.Incineration) * core.SpellCritRatingPerCritChance
		mage.OnSpellRegistered(func(spell *core.Spell) {
			switch spell.SpellCode {
			case SpellCode_MageFireBlast, SpellCode_MageScorch, SpellCode_MageIceLance, SpellCode_MageArcaneBlast:
				spell.BonusCritRating += bonusCrit
			}
		})
	}

	// Critical Mass
	if mage.Talents.CriticalMass > 0 {
		bonusCrit := 2 * float64(mage.Talents.CriticalMass) * core.SpellCritRatingPerCritChance
		mage.onMageSpellOfSchool(core.SpellSchoolFire, func(spell *core.Spell) {
			spell.BonusCritRating += bonusCrit
		})
	}

	// Fire Power
	if mage.Talents.FirePower > 0 {
		bonusDamageMultiplierAdditive := 0.02 * float64(mage.Talents.FirePower)
		mage.onMageSpellOfSchool(core.SpellSchoolFire, func(spell *core.Spell) {
			// Fire Power buffs pretty much all mage fire spells EXCEPT ignite
			if spell.SpellCode != SpellCode_MageIgnite {
				spell.DamageMultiplierAdditive += bonusDamageMultiplierAdditive
			}
		})
	}
}

func (mage *Mage) applyFrostTalents() {
	mage.registerColdSnapCD()
	mage.registerIceBarrierSpell()
	mage.applyWintersChill()
	mage.applyFingersOfFrost()

	// Elemental Precision: 1% hit with Fire and Frost spells per point (Classic took 2% off the
	// resist chance, 3 points).
	if mage.Talents.ElementalPrecision > 0 {
		bonusHit := float64(mage.Talents.ElementalPrecision) * core.SpellHitRatingPerHitChance
		mage.onMageSpellOfSchool(core.SpellSchoolFire|core.SpellSchoolFrost, func(spell *core.Spell) {
			spell.BonusHitRating += bonusHit
		})
	}

	// Ice Shards
	if mage.Talents.IceShards > 0 {
		critBonus := .20 * float64(mage.Talents.IceShards)
		mage.onMageSpellOfSchool(core.SpellSchoolFrost, func(spell *core.Spell) {
			spell.CritDamageBonus += critBonus
		})
	}

	// Piercing Ice
	if mage.Talents.PiercingIce > 0 {
		bonusDamageMultiplierAdditive := 0.02 * float64(mage.Talents.PiercingIce)
		mage.onMageSpellOfSchool(core.SpellSchoolFrost, func(spell *core.Spell) {
			spell.DamageMultiplierAdditive += bonusDamageMultiplierAdditive
		})
	}

	// Frost Channeling: 5% cheaper Frost spells and 10% less threat per point. The client's
	// spell mask leaves Frostfire Bolt out.
	if mage.Talents.FrostChanneling > 0 {
		manaCostMultiplier := 5 * mage.Talents.FrostChanneling
		threatMultiplier := 1 - .10*float64(mage.Talents.FrostChanneling)
		mage.onMageSpellOfSchool(core.SpellSchoolFrost, func(spell *core.Spell) {
			if spell.SpellCode == SpellCode_MageFrostfireBolt {
				return
			}
			if spell.Cost != nil {
				spell.Cost.Multiplier -= manaCostMultiplier
			}
			spell.ThreatMultiplier *= threatMultiplier
		})
	}
}

func (mage *Mage) applyArcaneConcentration() {
	if mage.Talents.ArcaneConcentration == 0 {
		return
	}

	procChance := 0.02 * float64(mage.Talents.ArcaneConcentration)

	mage.ClearcastingAura = mage.RegisterAura(core.Aura{
		Label:    "Clearcasting",
		ActionID: core.ActionID{SpellID: 12536},
		Duration: time.Second * 15,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.SchoolCostMultiplier.AddToMagicSchools(-100)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.SchoolCostMultiplier.AddToMagicSchools(100)
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			// OnCastComplete is called after OnSpellHitDealt / etc, so don't deactivate if it was just activated.
			if aura.RemainingDuration(sim) == aura.Duration {
				return
			}
			// The next damage spell with a cost uses it up. Clearcasting makes every cost 0,
			// so the base cost tells which spells have one (the fork's check on the current
			// cost never ended the aura early).
			if !spell.Flags.Matches(SpellFlagMage) || !spell.ProcMask.Matches(core.ProcMaskSpellDamage) {
				return
			}
			if spell.Cost == nil || spell.Cost.BaseCost == 0 {
				return
			}
			aura.Deactivate(sim)
		},
	})

	// The Forever client gives the talent a 1 sec internal cooldown, so the five missiles of one
	// Arcane Missiles can't each roll it.
	icd := core.Cooldown{
		Timer:    mage.NewTimer(),
		Duration: time.Second,
	}

	mage.RegisterAura(core.Aura{
		Label:    "Arcane Concentration",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.Flags.Matches(SpellFlagMage) || spell.SpellCode == SpellCode_MageArcaneMissiles {
				return
			}
			if !spell.ProcMask.Matches(core.ProcMaskSpellDamage) || !icd.IsReady(sim) {
				return
			}

			if sim.Proc(procChance, "Arcane Concentration") {
				icd.Use(sim)
				mage.ClearcastingAura.Activate(sim)
			}
		},
	})
}

// Missile Barrage: Arcane Blast has a 40% chance, Fireball, Frostbolt and Frostfire Bolt a 20%
// chance, to make the next Arcane Missiles within 15 sec free, with its missiles every 0.5 sec
// (the channel takes half as long). The roll is on the finished cast, as in the SoD sim.
func (mage *Mage) applyMissileBarrage() {
	if !mage.Talents.MissileBarrage {
		return
	}

	arcaneMissiles := []*core.Spell{}
	mage.MissileBarrageAura = mage.RegisterAura(core.Aura{
		Label:    "Missile Barrage",
		ActionID: core.ActionID{SpellID: 400589},
		Duration: time.Second * 15,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			arcaneMissiles = core.FilterSlice(mage.ArcaneMissiles, func(spell *core.Spell) bool { return spell != nil })
		},
		// The client takes 100% off the cost. Percent cost modifiers add up, so Arcane Power's
		// 30% is still paid.
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range arcaneMissiles {
				spell.Cost.Multiplier -= 100
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range arcaneMissiles {
				spell.Cost.Multiplier += 100
			}
		},
	})

	core.MakePermanent(mage.RegisterAura(core.Aura{
		Label: "Missile Barrage Talent",
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			var procChance float64
			switch spell.SpellCode {
			case SpellCode_MageArcaneBlast:
				procChance = 0.40
			case SpellCode_MageFireball, SpellCode_MageFrostbolt:
				procChance = 0.20
			case SpellCode_MageFrostfireBolt:
				procChance = 0.20 + mage.frostfireMissileBarrageBonus
			default:
				return
			}
			if sim.Proc(procChance, "Missile Barrage") {
				mage.MissileBarrageAura.Activate(sim)
			}
		},
	}))
}

func (mage *Mage) registerPresenceOfMindCD() {
	if !mage.Talents.PresenceOfMind {
		return
	}

	actionID := core.ActionID{SpellID: 12043}
	cooldown := time.Second * 180

	affectedSpells := []*core.Spell{}
	pomAura := mage.RegisterAura(core.Aura{
		Label:    "Presence of Mind",
		ActionID: actionID,
		Duration: time.Second * 15,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			for spellIdx := range mage.Spellbook {
				if spell := mage.Spellbook[spellIdx]; spell.DefaultCast.CastTime > 0 {
					affectedSpells = append(affectedSpells, spell)
				}
			}
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) {
				spell.CastTimeMultiplier -= 1
			})
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) {
				spell.CastTimeMultiplier += 1
			})
			mage.PresenceOfMind.CD.Use(sim)
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if !slices.Contains(affectedSpells, spell) {
				return
			}

			aura.Deactivate(sim)
		},
	})

	mage.PresenceOfMind = mage.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    mage.NewTimer(),
				Duration: cooldown,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return mage.GCD.IsReady(sim)
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			pomAura.Activate(sim)
		},
	})

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: mage.PresenceOfMind,
		Type:  core.CooldownTypeDPS,
	})
}

func (mage *Mage) registerArcanePowerCD() {
	if !mage.Talents.ArcanePower {
		return
	}

	actionID := core.ActionID{SpellID: 12042}

	affectedSpells := []*core.Spell{}

	mage.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Flags.Matches(SpellFlagMage) {
			affectedSpells = append(affectedSpells, spell)
		}
	})

	mage.ArcanePowerAura = mage.RegisterAura(core.Aura{
		Label:    "Arcane Power",
		ActionID: actionID,
		Duration: time.Second * 15,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedSpells {
				spell.DamageMultiplierAdditive += 0.3
				if spell.Cost != nil {
					spell.Cost.Multiplier += 30
				}
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedSpells {
				spell.DamageMultiplierAdditive -= 0.3
				if spell.Cost != nil {
					spell.Cost.Multiplier -= 30
				}
			}
		},
	})
	core.RegisterPercentDamageModifierEffect(mage.ArcanePowerAura, 1.3)

	spell := mage.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    mage.NewTimer(),
				Duration: time.Second * 180,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			mage.ArcanePowerAura.Activate(sim)
		},
	})

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeDPS,
	})
}

// Improved Scorch: Scorch has a 33 / 67 / 100% chance to apply Fire Vulnerability (22959), 3%
// more Fire damage per stack, up to 5, for 30 sec. Under Forever it only counts for the mage who
// applied it ("all Fire damage you deal to your target"), so it lives on the mage's own attack
// tables rather than on the target.
func (mage *Mage) applyImprovedScorch() {
	if mage.Talents.ImprovedScorch == 0 {
		return
	}

	mage.ImprovedScorchAuras = mage.NewEnemyAuraArray(func(unit *core.Unit) *core.Aura {
		return unit.RegisterAura(core.Aura{
			Label:     "Fire Vulnerability - " + mage.Label,
			ActionID:  core.ActionID{SpellID: 22959},
			Duration:  time.Second * 30,
			MaxStacks: 5,
			// The mage's own Scorch puts it up, so the rotation knows the stacks at once.
			NoReaction: true,
		})
	})

	// The attack tables are made after the talents.
	mage.Env.RegisterPostFinalizeEffect(func() {
		for _, target := range mage.Env.Encounter.TargetUnits {
			aura := mage.ImprovedScorchAuras.Get(target)
			for _, at := range mage.AttackTables[target.UnitIndex] {
				at.DamageDoneByCasterMultiplier = func(spell *core.Spell, _ *core.AttackTable) float64 {
					if !aura.IsActive() || !spell.SpellSchool.Matches(core.SpellSchoolFire) {
						return 1
					}
					return 1 + 0.03*float64(aura.GetStacks())
				}
			}
		}
	})
}

func (mage *Mage) applyMasterOfElements() {
	if mage.Talents.MasterOfElements == 0 {
		return
	}

	refundCoeff := 0.1 * float64(mage.Talents.MasterOfElements)
	manaMetrics := mage.NewManaMetrics(core.ActionID{SpellID: 29076})

	mage.RegisterAura(core.Aura{
		Label:    "Master of Elements",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.ProcMask.Matches(core.ProcMaskSpellDamage) || !spell.SpellSchool.Matches(core.SpellSchoolFire|core.SpellSchoolFrost) {
				return
			}
			if spell.Cost == nil || spell.CurCast.Cost == 0 {
				return
			}
			if result.DidCrit() {
				mage.AddMana(sim, spell.Cost.BaseCost*refundCoeff, manaMetrics)
			}
		},
	})
}

// Heating Up (Hot Streak until the 2026-10-08 build): non-periodic crits with Fireball,
// Frostfire Bolt, Fire Blast and Scorch take 25% off the next Pyroblast's cast time, stacking to
// 3 (75%), for 20 sec. The next Pyroblast uses up every stack (the client buff 400625 has one
// charge and a -25% cast time modifier per stack).
func (mage *Mage) applyHeatingUp() {
	if !mage.Talents.HeatingUp {
		return
	}

	var pyroblasts []*core.Spell
	mage.HeatingUpAura = mage.RegisterAura(core.Aura{
		Label:     "Heating Up",
		ActionID:  core.ActionID{SpellID: 400625},
		Duration:  time.Second * 20,
		MaxStacks: 3,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			pyroblasts = core.FilterSlice(mage.Pyroblast, func(spell *core.Spell) bool { return spell != nil })
		},
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks int32, newStacks int32) {
			for _, spell := range pyroblasts {
				spell.CastTimeMultiplier -= 0.25 * float64(newStacks-oldStacks)
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.SpellCode == SpellCode_MagePyroblast {
				aura.Deactivate(sim)
			}
		},
	})

	core.MakePermanent(mage.RegisterAura(core.Aura{
		Label: "Heating Up Talent",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.DidCrit() {
				return
			}
			switch spell.SpellCode {
			case SpellCode_MageFireball, SpellCode_MageFrostfireBolt, SpellCode_MageFireBlast, SpellCode_MageScorch:
				mage.HeatingUpAura.Activate(sim)
				mage.HeatingUpAura.AddStack(sim)
			}
		},
	}))
}

// Combustion: every Fire spell hit adds 10% crit to the mage's Fire spells, until 3 non-periodic
// Fire crits (the 2026-10-08 build put it back to 3 from the beta's 4).
func (mage *Mage) registerCombustionCD() {
	if !mage.Talents.Combustion {
		return
	}

	actionID := core.ActionID{SpellID: 11129}
	cd := core.Cooldown{
		Timer:    mage.NewTimer(),
		Duration: time.Minute * 3,
	}

	var fireSpells []*core.Spell
	mage.onMageSpellOfSchool(core.SpellSchoolFire, func(spell *core.Spell) {
		fireSpells = append(fireSpells, spell)
	})

	const maxCrits = 3
	numCrits := 0
	critPerStack := 10.0 * core.SpellCritRatingPerCritChance

	mage.CombustionAura = mage.RegisterAura(core.Aura{
		Label:     "Combustion",
		ActionID:  actionID,
		Duration:  core.NeverExpires,
		MaxStacks: 20,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			numCrits = 0
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			cd.Use(sim)
			mage.UpdateMajorCooldowns()
		},
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks int32, newStacks int32) {
			bonusCrit := critPerStack * float64(newStacks-oldStacks)
			for _, spell := range fireSpells {
				spell.BonusCritRating += bonusCrit
			}
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || numCrits >= maxCrits || !spell.SpellSchool.Matches(core.SpellSchoolFire) || !spell.Flags.Matches(SpellFlagMage) {
				return
			}
			if spell.SpellCode == SpellCode_MageIgnite {
				return
			}

			// TODO: This wont work properly with flamestrike
			aura.AddStack(sim)

			if result.DidCrit() {
				numCrits++
				if numCrits == maxCrits {
					aura.Deactivate(sim)
				}
			}
		},
	})

	spell := mage.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: cd,
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !mage.CombustionAura.IsActive()
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			mage.CombustionAura.Activate(sim)
			mage.CombustionAura.AddStack(sim)
		},
	})

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeDPS,
	})
}

func (mage *Mage) registerColdSnapCD() {
	if !mage.Talents.ColdSnap {
		return
	}

	// Grab all frost spells with a CD > 0
	var affectedSpells = []*core.Spell{}
	mage.OnSpellRegistered(func(spell *core.Spell) {
		if spell.SpellSchool.Matches(core.SpellSchoolFrost) && spell.CD.Duration > 0 {
			affectedSpells = append(affectedSpells, spell)
		}
	})

	spell := mage.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{SpellID: 12472},
		Flags:    core.SpellFlagNoOnCastComplete,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    mage.NewTimer(),
				Duration: time.Duration(time.Minute * 10),
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			for _, spell := range affectedSpells {
				spell.CD.Reset()
			}
		},
	})

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeDPS,
	})
}

// Winter's Chill: Frost damage spells have a 20% chance per point to put a stack of Winter's
// Chill (12579) on the target, at most one stack per point. Each stack gives the mage's own
// Frostbolt and Ice Lance 2% crit on that target for 15 sec. Classic's gave every Frost spell of
// everyone the crit.
func (mage *Mage) applyWintersChill() {
	if mage.Talents.WintersChill == 0 {
		return
	}

	procChance := float64(mage.Talents.WintersChill) * 0.2
	maxStacks := mage.Talents.WintersChill

	mage.WintersChillAuras = mage.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(core.Aura{
			Label:     "Winter's Chill - " + mage.Label,
			ActionID:  core.ActionID{SpellID: 12579},
			Duration:  time.Second * 15,
			MaxStacks: maxStacks,
		})
	})

	mage.RegisterAura(core.Aura{
		Label:    "Winters Chill Talent",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.SpellSchool.Matches(core.SpellSchoolFrost) || !spell.Flags.Matches(SpellFlagMage) {
				return
			}
			if !spell.ProcMask.Matches(core.ProcMaskSpellDamage) {
				return
			}

			if sim.Proc(procChance, "Winters Chill") {
				aura := mage.WintersChillAuras.Get(result.Target)
				aura.Activate(sim)
				aura.AddStack(sim)
			}
		},
	})
}

// Fingers of Frost: Chill effects have a 15% chance to make the next spell (two spells at rank 2)
// within 15 sec act as if the target were Frozen, which turns on Shatter and Ice Lance's bonus.
// A charge goes when a damage spell finishes casting, as in the SoD sim.
func (mage *Mage) applyFingersOfFrost() {
	if mage.Talents.FingersOfFrost == 0 {
		return
	}

	mage.fingersOfFrostChance = 0.15

	mage.FingersOfFrostAura = mage.RegisterAura(core.Aura{
		Label:     "Fingers of Frost",
		ActionID:  core.ActionID{SpellID: 400670},
		Duration:  time.Second * 15,
		MaxStacks: mage.Talents.FingersOfFrost,
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			// OnCastComplete is called after OnSpellHitDealt / etc, so don't take a charge for the proc itself.
			if aura.RemainingDuration(sim) == aura.Duration {
				return
			}
			if !spell.Flags.Matches(SpellFlagMage) || !spell.ProcMask.Matches(core.ProcMaskSpellDamage) {
				return
			}
			aura.RemoveStack(sim)
		},
	})

	mage.RegisterAura(core.Aura{
		Label:    "Fingers of Frost Talent",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Landed() && spell.Flags.Matches(SpellFlagChillSpell) {
				mage.procFingersOfFrost(sim, spell)
			}
		},
	})
}

// procFingersOfFrost rolls Fingers of Frost for a chill effect from spell.
func (mage *Mage) procFingersOfFrost(sim *core.Simulation, spell *core.Spell) {
	if mage.FingersOfFrostAura == nil {
		return
	}
	chance := mage.fingersOfFrostChance
	if spell.SpellCode == SpellCode_MageFrostfireBolt {
		chance += mage.frostfireFingersOfFrostBonus
	}
	if sim.Proc(chance, "Fingers of Frost") {
		mage.FingersOfFrostAura.Activate(sim)
		mage.FingersOfFrostAura.SetStacks(sim, mage.FingersOfFrostAura.MaxStacks)
	}
}

// isTargetFrozen is whether the mage's spells treat the target as Frozen. Raid bosses can't be
// frozen, so only Fingers of Frost does it.
func (mage *Mage) isTargetFrozen(_ *core.Unit) bool {
	return mage.FingersOfFrostAura != nil && mage.FingersOfFrostAura.IsActive()
}

// personalCritRating is the crit the mage's own effects on the target add to spell: Winter's
// Chill on Frostbolt and Ice Lance, and Shatter (17 / 33 / 50%) while the target counts as Frozen.
func (mage *Mage) personalCritRating(spell *core.Spell, target *core.Unit) float64 {
	bonus := 0.0
	if mage.WintersChillAuras != nil && (spell.SpellCode == SpellCode_MageFrostbolt || spell.SpellCode == SpellCode_MageIceLance) {
		if aura := mage.WintersChillAuras.Get(target); aura.IsActive() {
			bonus += 2 * float64(aura.GetStacks()) * core.SpellCritRatingPerCritChance
		}
	}
	if mage.Talents.Shatter > 0 && mage.isTargetFrozen(target) {
		bonus += []float64{0, 17, 33, 50}[mage.Talents.Shatter] * core.SpellCritRatingPerCritChance
	}
	return bonus
}

// calcMageDamage is spell.CalcDamage with the mage's personal crit on the target.
func (mage *Mage) calcMageDamage(sim *core.Simulation, target *core.Unit, baseDamage float64, spell *core.Spell) *core.SpellResult {
	bonus := mage.personalCritRating(spell, target)
	spell.BonusCritRating += bonus
	result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	spell.BonusCritRating -= bonus
	return result
}
