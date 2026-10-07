package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The Forever trees from the beta client (build 1.60.1.69876). Talents the sim does not
// read: Spiritual Focus, Unyielding Faith, Voice of Truth, Infusion of Light,
// Illumination, Light's Vigil, Guardian's Favor, Improved Seal of Fury, Swift Judgement,
// Improved Hammer of Justice, Templar's Bulwark, Pursuit of Justice, Eye for an Eye,
// Repentance, and the Echo half of Twist of Light. Healing Light only touches heals.
func (paladin *Paladin) ApplyTalents() {
	paladin.AddStat(stats.MeleeHit, float64(paladin.Talents.Precision)*core.MeleeHitRatingPerHitChance)
	paladin.AddStat(stats.SpellHit, float64(paladin.Talents.DivinePrecision)*6*core.SpellHitRatingPerHitChance)
	paladin.AddStat(stats.MeleeCrit, float64(paladin.Talents.Conviction)*core.CritRatingPerCritChance)
	paladin.AddStat(stats.SpellCrit, float64(paladin.Talents.HolyPower)*core.SpellCritRatingPerCritChance)

	if paladin.Talents.Toughness > 0 {
		paladin.ApplyEquipScaling(stats.Armor, 1.0+0.02*float64(paladin.Talents.Toughness))
	}

	// These are no-op if untalented.
	paladin.MultiplyStat(stats.Strength, 1.0+0.02*float64(paladin.Talents.DivineStrength))
	paladin.MultiplyStat(stats.Intellect, 1.0+0.02*float64(paladin.Talents.DivineIntellect))
	paladin.MultiplyStat(stats.Stamina, 1.0+0.02*float64(paladin.Talents.SacredDuty))
	paladin.AddStat(stats.Defense, 4*float64(paladin.Talents.Anticipation))
	paladin.AddStat(stats.Parry, 1*float64(paladin.Talents.Deflection))

	// Shield Specialization bonus is additive. NOTE: Total SBV will be inflated until
	// https://github.com/wowsims/sod/issues/1025 gets resolved.
	paladin.PseudoStats.BlockValueMultiplier += 0.1 * float64(paladin.Talents.ShieldSpecialization)

	// Reverence lets part of the mana regeneration through while the paladin is inside
	// the five second rule, which a paladin that seals and judges always is.
	paladin.PseudoStats.SpiritRegenRateCasting += 0.1 * float64(paladin.Talents.Reverence)

	// Champion of the Light turns Intellect into spell damage and healing.
	//
	// The 2026-10-01 development notes cut it to 20 / 40 / 60% (it was 33 / 66 / 100%).
	if paladin.Talents.ChampionOfTheLight > 0 {
		paladin.AddStatDependency(stats.Intellect, stats.SpellPower, []float64{0, .2, .4, .6}[paladin.Talents.ChampionOfTheLight])
	}

	paladin.applyWeaponSpecialization()
	paladin.applyVengeance()
	paladin.applyVindication()
	paladin.applyRedoubt()
	paladin.applyReckoning()
	paladin.applyShieldSpecializationMana()
	paladin.applyConsecratedGround()
}

// Improved Seals raises the damage of every seal proc and judgement.
func (paladin *Paladin) improvedSeals() float64 {
	return []float64{1, 1.05, 1.10, 1.15}[paladin.Talents.ImprovedSeals]
}

// Benediction takes 2% per point off instant spells and abilities.
func (paladin *Paladin) benediction() int32 {
	return []int32{100, 98, 96, 94, 92, 90}[paladin.Talents.Benediction]
}

// The seals take Benediction's discount, and since the 2026-09-24 beta build Twist of
// Light takes another 20% off them. We add the two, the way percent cost reductions
// usually combine.
func (paladin *Paladin) sealCostMultiplier() int32 {
	if paladin.Talents.TwistOfLight {
		return paladin.benediction() - 20
	}
	return paladin.benediction()
}

// Holy Conduit takes 20% per point off Consecration, Holy Wrath, Exorcism and Hammer of Wrath.
func (paladin *Paladin) holyConduit() int32 {
	return []int32{100, 80, 60}[paladin.Talents.HolyConduit]
}

// holyConduitInstant is Holy Conduit's discount plus Benediction's, for the spells Holy
// Conduit covers when they are instant (Consecration, Exorcism, and Hammer of Wrath with
// both points of Instrument of Law). Benediction's text says "all instant cast spells and
// abilities". We add the two like sealCostMultiplier does.
func (paladin *Paladin) holyConduitInstant() int32 {
	return paladin.holyConduit() + paladin.benediction() - 100
}

// Purifying Power shortens the Exorcism and Holy Wrath cooldowns by 17% per point.
func (paladin *Paladin) purifyingPower(cd time.Duration) time.Duration {
	return time.Duration(float64(cd) * []float64{1, 0.83, 0.67}[paladin.Talents.PurifyingPower])
}

// Redoubt now procs from any damaging melee attack taken (10% chance), Classic needed a
// crit. The block bonus is 6% per point for 10 sec or 5 blocks.
func (paladin *Paladin) applyRedoubt() {
	if paladin.Talents.Redoubt == 0 {
		return
	}

	// Build 70245 gives 4% a rank (Classic and the 69876 beta client 6%).
	blockBonus := 4.0 * float64(paladin.Talents.Redoubt) * core.BlockRatingPerBlockChance

	paladin.redoubtAura = paladin.RegisterAura(core.Aura{
		Label:     "Redoubt",
		ActionID:  core.ActionID{SpellID: 20134},
		Duration:  time.Second * 10,
		MaxStacks: 5,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			paladin.AddStatDynamic(sim, stats.Block, blockBonus)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			paladin.AddStatDynamic(sim, stats.Block, -blockBonus)
		},
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.DidBlock() {
				aura.RemoveStack(sim)
			}
		},
	})

	core.MakeProcTriggerAura(&paladin.Unit, core.ProcTrigger{
		Name:       "Redoubt Trigger",
		Callback:   core.CallbackOnSpellHitTaken,
		Outcome:    core.OutcomeLanded,
		ProcMask:   core.ProcMaskMelee,
		ProcChance: 0.1,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			paladin.redoubtAura.Activate(sim)
			paladin.redoubtAura.SetStacks(sim, 5)
		},
	})
}

// Reckoning gives an extra attack on 20% per point of the crits taken, and on 8% per
// point of the blocks.
func (paladin *Paladin) applyReckoning() {
	if paladin.Talents.Reckoning == 0 {
		return
	}

	procID := core.ActionID{SpellID: 20178} // Reckoning Proc ID
	critChance := 0.2 * float64(paladin.Talents.Reckoning)
	blockChance := 0.08 * float64(paladin.Talents.Reckoning)

	paladin.RegisterAura(core.Aura{
		Label:    "Reckoning Trigger",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.ProcMask.Matches(core.ProcMaskMeleeOrRanged) {
				return
			}
			if result.DidCrit() && sim.Proc(critChance, "Reckoning") ||
				result.DidBlock() && sim.Proc(blockChance, "Reckoning") {
				paladin.AutoAttacks.ExtraMHAttack(sim, 1, procID, spell.ActionID)
			}
		},
	})
}

// Shield Specialization's second half: a block has a 33% chance per point to restore
// 6% of the paladin's maximum mana, once every 3 sec.
func (paladin *Paladin) applyShieldSpecializationMana() {
	if paladin.Talents.ShieldSpecialization == 0 {
		return
	}

	actionID := core.ActionID{SpellID: 20150}
	manaMetrics := paladin.NewManaMetrics(actionID)
	procChance := []float64{0, 0.33, 0.66, 1}[paladin.Talents.ShieldSpecialization]
	icd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * 3,
	}

	paladin.RegisterAura(core.Aura{
		Label:    "Shield Specialization Mana",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.DidBlock() && icd.IsReady(sim) && sim.Proc(procChance, "Shield Specialization") {
				icd.Use(sim)
				paladin.AddMana(sim, 0.06*paladin.MaxMana(), manaMetrics)
			}
		},
	})
}

// The one and two handed specializations. Forever's ranks are 3/7/10% for one handers
// and 2/4/6% for two handers (3/6/9% before the 2026-09-24 beta build), and they scale
// everything physical the weapon deals,
// plus the seal procs and judgements that roll as melee.
func (paladin *Paladin) getWeaponSpecializationModifier() float64 {
	handType := paladin.MainHand().HandType
	if handType == proto.HandType_HandTypeMainHand || handType == proto.HandType_HandTypeOneHand {
		return []float64{1, 1.03, 1.07, 1.10}[paladin.Talents.OneHandedWeaponSpecialization]
	} else if handType == proto.HandType_HandTypeTwoHand {
		return []float64{1, 1.02, 1.04, 1.06}[paladin.Talents.TwoHandedWeaponSpecialization]
	} else {
		return 1.
	}
}

// Affects all physical damage or spells that can be rolled as physical.
func (paladin *Paladin) applyWeaponSpecialization() {
	paladin.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= paladin.getWeaponSpecializationModifier()
}

// Vengeance stacks 1% per point of Physical and Holy damage for every crit, up to 3
// stacks, for 30 sec. Classic was a flat 3% per point for 8 sec.
//
// The 2026-09-24 beta build cut it from 5 stacks to 3 and made only non-periodic crits
// count. The hit callback below never sees a periodic tick (those go to the periodic
// callback), so the second half needs nothing here.
func (paladin *Paladin) applyVengeance() {
	if paladin.Talents.Vengeance == 0 {
		return
	}

	perStack := 0.01 * float64(paladin.Talents.Vengeance)
	multiplier := func(stacks int32) float64 {
		return 1 + perStack*float64(stacks)
	}

	procAura := paladin.RegisterAura(core.Aura{
		Label:     "Vengeance Proc",
		ActionID:  core.ActionID{SpellID: 20059},
		Duration:  time.Second * 30,
		MaxStacks: 3,
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks int32, newStacks int32) {
			aura.Unit.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexHoly] *= multiplier(newStacks) / multiplier(oldStacks)
			aura.Unit.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= multiplier(newStacks) / multiplier(oldStacks)
		},
	})

	paladin.RegisterAura(core.Aura{
		Label:    "Vengeance",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.DidCrit() {
				procAura.Activate(sim)
				procAura.AddStack(sim)
			}
		},
	})
}

// Vindication gives the paladin 1% per point of attack power for 30 sec on a melee hit
// (and takes attack power from the target, which the sim does not score).
func (paladin *Paladin) applyVindication() {
	if paladin.Talents.Vindication == 0 {
		return
	}
	vindicationMultiplier := paladin.NewDynamicMultiplyStat(stats.AttackPower, 1+0.01*float64(paladin.Talents.Vindication))

	vindicationAura := paladin.RegisterAura(core.Aura{
		Label:    "Vindication Proc",
		ActionID: core.ActionID{SpellID: 26021},
		Duration: time.Second * 30,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			paladin.EnableDynamicStatDep(sim, vindicationMultiplier)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			paladin.DisableDynamicStatDep(sim, vindicationMultiplier)
		},
	})
	paladin.RegisterAura(core.Aura{
		Label:    "Vindication Talent",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			// TODO: Replace with actual proc mask / proc chance
			if result.Landed() && spell.ProcMask.Matches(core.ProcMaskMelee) {
				vindicationAura.Activate(sim)
			}
		},
	})
}

// Consecrated Ground: Holy spells deal 5% per point more to the first four enemies
// inside the paladin's Consecration. The sim's targets all stand in it, so the bonus is
// a paladin buff that Consecration puts up for its 8 sec.
func (paladin *Paladin) applyConsecratedGround() {
	if paladin.Talents.ConsecratedGround == 0 {
		return
	}
	multiplier := 1 + 0.05*float64(paladin.Talents.ConsecratedGround)
	paladin.consecratedGroundAura = paladin.RegisterAura(core.Aura{
		Label:    "Consecrated Ground",
		ActionID: core.ActionID{SpellID: 1310905},
		Duration: time.Second * 8,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			paladin.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexHoly] *= multiplier
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			paladin.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexHoly] /= multiplier
		},
	})
}
