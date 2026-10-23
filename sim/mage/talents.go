package mage

import (
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The talents follow the Forever beta client's tree (tools/forever_talents/client/mage.json,
// build 1.60.1.69876 with the changes up to 1.60.1.70425). The ones that change nothing in a
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

// changeCostPercent puts on (gain) or takes off a percent cost modifier on spell, +30 for 30% more.
// Under Forever each SPELLMOD_COST row multiplies the others (wowsims/forever fef2bcaebd: Improved
// Wrath with Moonglow leaves a 120-mana Wrath at 45, not 30), and a row of -100% or lower makes the
// spell free while it is up, whatever the other rows are. The free rows go on Cost.Multiplier, the
// rest on Cost.Factor. Classic adds them all up.
func (mage *Mage) changeCostPercent(spell *core.Spell, percent int32, gain bool) {
	if spell.Cost == nil {
		return
	}
	if !mage.Env.IsForever() || percent <= -100 {
		spell.Cost.Multiplier += core.TernaryInt32(gain, percent, -percent)
		return
	}
	if gain {
		spell.Cost.Factor *= 1 + float64(percent)/100
	} else {
		spell.Cost.Factor /= 1 + float64(percent)/100
	}
}

// changeDamagePercent puts on (gain) or takes off a percent damage modifier on spell, 0.3 for 30%
// more. Under Forever each SPELLMOD_DAMAGE and SPELLMOD_DOT row is its own factor (wowsims/forever
// fef2bcaebd and 810101380a: Arcane Power with Arcane Instability 3 is 1.3 x 1.03), so it goes on
// DamageMultiplier. Classic adds them up.
func (mage *Mage) changeDamagePercent(spell *core.Spell, bonus float64, gain bool) {
	if !mage.Env.IsForever() {
		spell.DamageMultiplierAdditive += core.TernaryFloat64(gain, bonus, -bonus)
		return
	}
	if gain {
		spell.DamageMultiplier *= 1 + bonus
	} else {
		spell.DamageMultiplier /= 1 + bonus
	}
}

func (mage *Mage) applyArcaneTalents() {
	mage.applyArcaneConcentration()
	mage.applyMissileBarrage()
	mage.registerPresenceOfMindCD()
	mage.registerArcanePowerCD()

	// Improved Channeling: per point, a 20% chance for Arcane Missiles and 14% for Arcane Blast and
	// Arcane Wrath not to lose time to a hit (11237: effect 0 masks Arcane Missiles, effect 1
	// Arcane Blast and Arcane Wrath).
	if mage.Talents.ImprovedChanneling > 0 {
		points := float64(mage.Talents.ImprovedChanneling)
		mage.OnSpellRegistered(func(spell *core.Spell) {
			switch spell.SpellCode {
			case SpellCode_MageArcaneMissiles:
				spell.PushbackReduction += 0.20 * points
			case SpellCode_MageArcaneBlast, SpellCode_MageArcaneWrath:
				spell.PushbackReduction += 0.14 * points
			}
		})
	}

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
		bonusDamage := .01 * float64(mage.Talents.ArcaneInstability)
		bonusCritRating := 1 * float64(mage.Talents.ArcaneInstability) * core.SpellCritRatingPerCritChance

		mage.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagMage) {
				mage.changeDamagePercent(spell, bonusDamage, true)
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

	// Burning Soul: 10% less threat from Fire spells per point, and a 23 / 47 / 70% chance not to
	// lose casting time to a hit (11083 effect 0, whose mask is Fireball, Scorch, Pyroblast,
	// Flamestrike and Frostfire Bolt).
	if mage.Talents.BurningSoul > 0 {
		threatMultiplier := 1 - .10*float64(mage.Talents.BurningSoul)
		mage.onMageSpellOfSchool(core.SpellSchoolFire, func(spell *core.Spell) {
			spell.ThreatMultiplier *= threatMultiplier
		})
		pushback := []float64{0, 0.23, 0.47, 0.70}[mage.Talents.BurningSoul]
		mage.OnSpellRegistered(func(spell *core.Spell) {
			switch spell.SpellCode {
			case SpellCode_MageFireball, SpellCode_MageScorch, SpellCode_MagePyroblast, SpellCode_MageFlamestrike, SpellCode_MageFrostfireBolt:
				spell.PushbackReduction += pushback
			}
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

	// Critical Mass: 2% crit with Fire spells per point, 5 points since 1.60.1.70425 (3 before).
	if mage.Talents.CriticalMass > 0 {
		bonusCrit := 2 * float64(mage.Talents.CriticalMass) * core.SpellCritRatingPerCritChance
		mage.onMageSpellOfSchool(core.SpellSchoolFire, func(spell *core.Spell) {
			spell.BonusCritRating += bonusCrit
		})
	}

	mage.applyFireTargetMultipliers()
}

func (mage *Mage) applyFrostTalents() {
	mage.registerColdSnapCD()
	mage.registerIceBarrierSpell()
	mage.applyFingersOfFrost()
	mage.applyShatter()
	mage.applyWintersChill()

	// Elemental Precision: 2% hit with Fire and Frost spells per point, 3 points since the
	// 2026-10-20 hotfix to 70450 (5 points of 1% before; Classic took 2% off the resist chance).
	if mage.Talents.ElementalPrecision > 0 {
		bonusHit := 2 * float64(mage.Talents.ElementalPrecision) * core.SpellHitRatingPerHitChance
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

	// Frost Channeling: 5% cheaper Frost spells and 10% less threat per point. The cost row's
	// mask (11160 effect 0: 0xA0000, word 1 0x1000) takes Frostfire Bolt by word 1 bit 12, the
	// bit Ice Shards and Critical Mass reach it by, and Ice Barrier too.
	if mage.Talents.FrostChanneling > 0 {
		costPercent := -5 * mage.Talents.FrostChanneling
		threatMultiplier := 1 - .10*float64(mage.Talents.FrostChanneling)
		mage.onMageSpellOfSchool(core.SpellSchoolFrost, func(spell *core.Spell) {
			mage.changeCostPercent(spell, costPercent, true)
			spell.ThreatMultiplier *= threatMultiplier
		})
		mage.OnSpellRegistered(func(spell *core.Spell) {
			if slices.Contains(IceBarrierSpellId[1:], spell.SpellID) {
				mage.changeCostPercent(spell, costPercent, true)
			}
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

// Missile Barrage (2 points since 1.60.1.70425, Arcane row 1): per point, Arcane Blast has a 20%
// chance, every Arcane Missiles missile that hits 5%, and Arcane Wrath, Fireball, Frostbolt and
// Frostfire Bolt 10%, to make the next Arcane Missiles within 15 sec free, with its missiles every
// 0.5 sec (the channel takes half as long). The roll is on the finished cast, as in the SoD sim,
// and on the hit for the missiles.
func (mage *Mage) applyMissileBarrage() {
	if mage.Talents.MissileBarrage == 0 {
		return
	}

	points := float64(mage.Talents.MissileBarrage)

	arcaneMissiles := []*core.Spell{}
	mage.MissileBarrageAura = mage.RegisterAura(core.Aura{
		Label:    "Missile Barrage",
		ActionID: core.ActionID{SpellID: 400589},
		Duration: time.Second * 15,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			arcaneMissiles = core.FilterSlice(mage.ArcaneMissiles, func(spell *core.Spell) bool { return spell != nil })
		},
		// The client takes 100% off the cost (400589 effect 2). Under Forever that makes the
		// missiles free whatever else is up, Arcane Power included.
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range arcaneMissiles {
				mage.changeCostPercent(spell, -100, true)
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range arcaneMissiles {
				mage.changeCostPercent(spell, -100, false)
			}
		},
	})

	core.MakePermanent(mage.RegisterAura(core.Aura{
		Label: "Missile Barrage Talent",
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			var procChance float64
			switch spell.SpellCode {
			case SpellCode_MageArcaneBlast:
				procChance = 0.20 * points
			case SpellCode_MageArcaneWrath, SpellCode_MageFireball, SpellCode_MageFrostbolt:
				procChance = 0.10 * points
			case SpellCode_MageFrostfireBolt:
				procChance = 0.10*points + mage.frostfireMissileBarrageBonus
			default:
				return
			}
			if sim.Proc(procChance, "Missile Barrage") {
				mage.MissileBarrageAura.Activate(sim)
			}
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.SpellCode != SpellCode_MageArcaneMissilesTick || !result.Landed() {
				return
			}
			if sim.Proc(0.05*points, "Missile Barrage") {
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

			mage.presenceOfMindSpentAt = sim.CurrentTime
			aura.Deactivate(sim)
		},
	})

	mage.PresenceOfMindAura = pomAura

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
				mage.changeDamagePercent(spell, 0.3, true)
				mage.changeCostPercent(spell, 30, true)
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedSpells {
				mage.changeDamagePercent(spell, 0.3, false)
				mage.changeCostPercent(spell, 30, false)
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
}

// Fire Power (3 points since 1.60.1.70425, 5 before): the mage's Fire spells do 1% more damage
// per point for each of the mage's own Fire effects on the target, at most 4 of them (3 points:
// 3% each, up to 12%). Effects of more than one school, like Frostfire Bolt's DoT, don't count.
// The effects are the DoTs of the mage's Fire spells (Fireball, Pyroblast, Flamestrike, Ignite)
// and Fire Vulnerability. Ignite takes a share of the crit, which already has the bonus, so it
// gets none itself. Frostfire Bolt is a Fire spell too and takes it (the tooltip doesn't say).
const FirePowerMaxEffects = 4

func (mage *Mage) firePowerEffects(target *core.Unit, dots []*core.Spell) int {
	count := 0
	if mage.ImprovedScorchAuras != nil && mage.ImprovedScorchAuras.Get(target).IsActive() {
		count++
	}
	for _, spell := range dots {
		if dot := spell.AOEDot(); dot != nil {
			if dot.IsActive() {
				count++
			}
		} else if spell.Dot(target).IsActive() {
			count++
		}
		if count >= FirePowerMaxEffects {
			return FirePowerMaxEffects
		}
	}
	return count
}

// applyFireTargetMultipliers puts Fire Vulnerability and Fire Power on the mage's attack tables.
// Both only count for the mage's own spells.
func (mage *Mage) applyFireTargetMultipliers() {
	if mage.Talents.ImprovedScorch == 0 && mage.Talents.FirePower == 0 {
		return
	}

	firePowerPerEffect := 0.01 * float64(mage.Talents.FirePower)
	var fireDots []*core.Spell
	if mage.Talents.FirePower > 0 {
		mage.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagMage) && spell.SpellSchool == core.SpellSchoolFire && (len(spell.Dots()) > 0 || spell.AOEDot() != nil) {
				fireDots = append(fireDots, spell)
			}
		})
	}

	// The attack tables are made after the talents.
	mage.Env.RegisterPostFinalizeEffect(func() {
		for _, target := range mage.Env.Encounter.TargetUnits {
			var fireVulnerability *core.Aura
			if mage.ImprovedScorchAuras != nil {
				fireVulnerability = mage.ImprovedScorchAuras.Get(target)
			}
			for _, at := range mage.AttackTables[target.UnitIndex] {
				at.DamageDoneByCasterMultiplier = func(spell *core.Spell, at *core.AttackTable) float64 {
					if !spell.SpellSchool.Matches(core.SpellSchoolFire) {
						return 1
					}
					multiplier := 1.0
					if fireVulnerability != nil && fireVulnerability.IsActive() {
						multiplier *= 1 + 0.03*float64(fireVulnerability.GetStacks())
					}
					if firePowerPerEffect > 0 && spell.Flags.Matches(SpellFlagMage) && spell.SpellCode != SpellCode_MageIgnite {
						multiplier *= 1 + firePowerPerEffect*float64(mage.firePowerEffects(at.Defender, fireDots))
					}
					return multiplier
				}
			}
		}
	})
}

// Master of Elements: Fire and Frost crits give back 10% of the spell's base cost per point.
// Under Forever a Clearcasting cast's crit does too (wowsims/forever c999a66b3f).
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
			if spell.Cost == nil || spell.Cost.BaseCost == 0 || (!mage.Env.IsForever() && spell.CurCast.Cost == 0) {
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
// 3 (75%), for 20 sec. Since 1.60.1.70425 every Fire Blast that deals damage adds a stack too,
// so a Fire Blast crit adds two (the 400624 tooltip: "Fire Blast grants an additional stack
// every time it deals damage"). The next Pyroblast uses up every stack (the client buff 400625
// has a -25% cast time modifier per stack), unless Presence of Mind made it instant (70425 notes:
// Heating Up is no longer consumed if Presence of Mind was active).
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
			if spell.SpellCode != SpellCode_MagePyroblast {
				return
			}
			// Presence of Mind may have ended just before this, on the same cast.
			if mage.PresenceOfMindAura != nil && (mage.PresenceOfMindAura.IsActive() || mage.presenceOfMindSpentAt == sim.CurrentTime) {
				return
			}
			aura.Deactivate(sim)
		},
	})

	core.MakePermanent(mage.RegisterAura(core.Aura{
		Label: "Heating Up Talent",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.SpellCode == SpellCode_MageFireBlast && result.Landed() {
				mage.HeatingUpAura.Activate(sim)
				mage.HeatingUpAura.AddStack(sim)
			}
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
		Label:    "Combustion",
		ActionID: actionID,
		Duration: core.NeverExpires,
		// The client's 28682 stacks to 10.
		MaxStacks: 10,
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

// Fingers of Frost: Chill effects have a 15% chance to make the next spell (two spells at rank 2)
// within 15 sec act as if the target were Frozen, which turns on Shatter and Ice Lance's bonus.
// As in wowsims/forever (talents_frost.go, 1373a0b96f), from their logs: every damage spell the
// rotation casts spends a charge a spell batch (10 ms) after it finishes, so a spell cast in the
// same batch still reads the target as Frozen, and charges granted inside that batch stay full.
// A channel spends none at its cast (Blizzard casts with charges up spent none); each Arcane
// Missiles missile that lands spends one as it rolls its damage. A cast that granted the charges
// from inside its own effects, the hit of a rotation spell without travel time such as Cone of
// Cold, spends none of them; a missile's impact exempts nothing (wowsims/forever e875ae5bd3: by
// timestamp, a Frostbolt finishing its cast as another landed and procced skipped its spend).
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
			if spell == mage.fingersOfFrostGrantedBy {
				mage.fingersOfFrostGrantedBy = nil
				return
			}
			if !spell.Flags.Matches(SpellFlagMage) || !spell.Flags.Matches(core.SpellFlagAPL) || spell.Flags.Matches(core.SpellFlagChanneled) {
				return
			}
			if !spell.ProcMask.Matches(core.ProcMaskSpellDamage) {
				return
			}
			grantsAtCast := mage.fingersOfFrostGrants
			core.StartDelayedAction(sim, core.DelayedActionOptions{
				DoAt:     sim.CurrentTime + core.SpellBatchWindow,
				Priority: core.ActionPriorityDOT,
				OnAction: func(sim *core.Simulation) {
					if aura.IsActive() && mage.fingersOfFrostGrants == grantsAtCast {
						aura.RemoveStack(sim)
					}
				},
			})
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

// spendFingersOfFrostOnMissile takes the charge an Arcane Missiles missile that lands spends.
func (mage *Mage) spendFingersOfFrostOnMissile(sim *core.Simulation, result *core.SpellResult) {
	if mage.FingersOfFrostAura != nil && mage.FingersOfFrostAura.IsActive() && result.Landed() {
		mage.FingersOfFrostAura.RemoveStack(sim)
	}
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
		if spell.MissileSpeed == 0 && spell.Flags.Matches(core.SpellFlagAPL) {
			mage.fingersOfFrostGrantedBy = spell
		}
		mage.FingersOfFrostAura.Activate(sim)
		mage.FingersOfFrostAura.SetStacks(sim, mage.FingersOfFrostAura.MaxStacks)
		mage.fingersOfFrostGrants++
	}
}

// isTargetFrozen is whether the mage's spells treat the target as Frozen. Raid bosses can't be
// frozen, so only Fingers of Frost does it.
func (mage *Mage) isTargetFrozen(_ *core.Unit) bool {
	return mage.FingersOfFrostAura != nil && mage.FingersOfFrostAura.IsActive()
}

// personalCritRating is the crit the mage's own effects on the target add to spell: Shatter
// (17 / 33 / 50%) while the target counts as Frozen. Winter's Chill gave Frostbolt and Ice Lance
// crit until 1.60.1.70425, which made it a Cone of Cold debuff instead (cone_of_cold.go).
func (mage *Mage) personalCritRating(spell *core.Spell, target *core.Unit) float64 {
	bonus := 0.0
	if mage.Talents.Shatter > 0 && mage.isTargetFrozen(target) {
		bonus += []float64{0, 17, 33, 50}[mage.Talents.Shatter] * core.SpellCritRatingPerCritChance
	}
	return bonus
}

// applyShatter gives every mage damage spell Shatter's crit while it rolls, as theirs does
// (wowsims/forever talents_frost.go: the bonus is on each matching spell for its own effects).
// Blizzard's ticks take it too: theirs are a spell of their own.
func (mage *Mage) applyShatter() {
	if mage.Talents.Shatter == 0 {
		return
	}
	mage.OnSpellRegistered(func(spell *core.Spell) {
		if !spell.Flags.Matches(SpellFlagMage) || !spell.ProcMask.Matches(core.ProcMaskSpellDamage) || spell.ApplyEffects == nil {
			return
		}
		applyEffects := spell.ApplyEffects
		spell.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			mage.withPersonalCrit(spell, target, func() {
				applyEffects(sim, target, spell)
			})
		}
	})
}

// withPersonalCrit runs f with the mage's personal crit on target added to spell.
func (mage *Mage) withPersonalCrit(spell *core.Spell, target *core.Unit, f func()) {
	bonus := mage.personalCritRating(spell, target)
	spell.BonusCritRating += bonus
	f()
	spell.BonusCritRating -= bonus
}
