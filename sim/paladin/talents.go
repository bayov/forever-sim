package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The Forever trees of client 1.60.1.70707 with the launch realms' October 30 hotfix (wago.tools
// hotfixes, region 91, build 70707: the cached pseudo-version 1.60.1.70707+hf, see
// tools/sources/hotfixes.py). Talents the sim does not read: Spiritual Focus, Unyielding Faith,
// Infusion of Light, Illumination, Light's Vigil (heals only since the hotfix), Oath of Light,
// Boundless Light, Holy Power (crit healing only since the hotfix), Guardian's Favor, Improved
// Hammer of Justice, Templar's Bulwark, Stoicism, Pursuit of Justice, Eye for an Eye, Turn Evil
// and Repentance. Healing Light only touches heals.
func (paladin *Paladin) ApplyTalents() {
	paladin.AddStat(stats.MeleeHit, float64(paladin.Talents.Precision)*core.MeleeHitRatingPerHitChance)
	// Forever's Precision gives spell hit too (client 1.60.1.70338, spell 20189: auras 54
	// and 55, "Improves your chance to hit by 1%").
	if paladin.Env.IsForever() {
		paladin.AddStat(stats.SpellHit, float64(paladin.Talents.Precision)*core.SpellHitRatingPerHitChance)
	}
	// Conviction: 2% crit a rank with all spells and attacks since the October 30 hotfix (spell
	// 20117 in 1.60.1.70707+hf: aura 290, A_MOD_CRIT_PCT, 3 ranks; it was 1% melee crit a rank,
	// aura 52, 5 ranks), so Hammer of Wrath, Holy Shock and the consumables' spells get it too.
	paladin.AddStat(stats.MeleeCrit, 2*float64(paladin.Talents.Conviction)*core.CritRatingPerCritChance)
	paladin.AddStat(stats.SpellCrit, 2*float64(paladin.Talents.Conviction)*core.SpellCritRatingPerCritChance)

	if paladin.Talents.Toughness > 0 {
		paladin.ApplyEquipScaling(stats.Armor, 1.0+0.02*float64(paladin.Talents.Toughness))
	}

	// These are no-op if untalented. Sacred Duty no longer gives Stamina since the hotfix (spell
	// 1224697 in 1.60.1.70707+hf: effect 0 is a Forbearance duration cut, see below).
	paladin.MultiplyStat(stats.Strength, 1.0+0.02*float64(paladin.Talents.DivineStrength))
	paladin.MultiplyStat(stats.Intellect, 1.0+0.02*float64(paladin.Talents.DivineIntellect))
	paladin.AddStat(stats.Defense, 4*float64(paladin.Talents.Anticipation))
	paladin.AddStat(stats.Parry, 1*float64(paladin.Talents.Deflection))
	// Sacred Duty takes 5 sec a rank off Forbearance (1224697, effect 0: -10000 ms at 2 ranks).
	paladin.forbearanceReduction += time.Second * 5 * time.Duration(paladin.Talents.SacredDuty)

	// Shield Specialization bonus is additive, 15% a rank of 2 since the hotfix (spell 20150's
	// trait ranks: 15 / 30%; it was 10% a rank of 3). NOTE: Total SBV will be inflated until
	// https://github.com/wowsims/sod/issues/1025 gets resolved.
	paladin.PseudoStats.BlockValueMultiplier += 0.15 * float64(paladin.Talents.ShieldSpecialization)

	// Reverence lets part of the mana regeneration through while the paladin is inside the five
	// second rule, which a paladin that seals and judges always is: 15% a rank of 2 since the hotfix
	// (spell 1310899, aura 134 at 30 for the second rank; it was 10% a rank of 3).
	paladin.PseudoStats.SpiritRegenRateCasting += 0.15 * float64(paladin.Talents.Reverence)

	// Champion of the Light turns Intellect into spell damage and healing: 12% of Intellect a rank
	// of 5 since the hotfix (spell 1311084, aura 174 on Intellect, the trait ranks 12 to 60%; it was
	// 20% a rank of 3). Its Hammer of Wrath cast cut is in hammer_of_wrath.go.
	if paladin.Talents.ChampionOfTheLight > 0 {
		paladin.AddStatDependency(stats.Intellect, stats.SpellPower, 0.12*float64(paladin.Talents.ChampionOfTheLight))
	}

	paladin.applyWeaponSpecialization()
	paladin.applyVengeance()
	paladin.applyRedoubt()
	paladin.applyReckoning()
	paladin.applyShieldSpecializationMana()
	paladin.applyConsecratedGround()
}

// Improved Seals raises the damage of every seal proc and judgement: 3% a rank of 5 since the
// October 30 hotfix (spell 20224, aura 108 misc 0 on the seals' and judgements' class mask, the
// trait ranks 3 to 15%; it was 5% a rank of 3). It moved from Retribution to Holy.
func (paladin *Paladin) improvedSeals() float64 {
	return 1 + 0.03*float64(paladin.Talents.ImprovedSeals)
}

// Holy Power gives no damage crit since the October 30 hotfix: spell 5923 in 1.60.1.70707+hf is
// aura 50 (critical healing amount) with the Holy school, "Increases the critical strike healing
// bonus of your Holy spells by 10 to 50%". The sim does not model the paladin's heals. Before the
// hotfix it was a crit modifier on a list of the paladin's spells (3% a point for Holy Shock and
// Holy Strike, 1% for the rest); holyPowerCrit stays as the hook those spells read, at 0.
func (paladin *Paladin) holyPowerCrit() float64 {
	return 0
}

// Benediction takes 2% per point off instant spells and abilities.
func (paladin *Paladin) benediction() int32 {
	return []int32{100, 98, 96, 94, 92, 90}[paladin.Talents.Benediction]
}

// divinePrecisionHitRating is Divine Precision's 4% hit a rank of 2 since the October 30 hotfix
// (spell 1310904 in 1.60.1.70707+hf: aura 107 misc 16 at 8 for the second rank; it was 6% a rank
// of 3). Its class mask covers Consecration, Exorcism, Holy Shock, Holy Strike, Holy Wrath and
// Light's Vigil (plus utility spells), but not the seals, the judgements or Hammer of Wrath, so it
// goes on those spells rather than on the whole Holy school.
func (paladin *Paladin) divinePrecisionHitRating(ratingPerHitChance float64) float64 {
	return 4 * float64(paladin.Talents.DivinePrecision) * ratingPerHitChance
}

// The seals take Benediction's discount. Twist of Light took another 20% off them; its successor
// Unsealed Wrath (spell 1310735 in 1.60.1.70707+hf) keeps the cost effect on the seals' class mask
// at 0 (aura 108, misc 14, base points -20 in the client's 70707 row), so the seals cost their
// Benediction price again.
func (paladin *Paladin) sealCostMultiplier() int32 {
	return paladin.benediction()
}

// Libram of Invocation's 5% off the seals (item 249442, as wowsims/forever has it in
// item_librams.go) is another percent cost modifier, so it multiplies with the others. It also
// counts in Sanctified Judgement's refund (sanctifiedJudgementBase).
func (paladin *Paladin) sealCostFactor() float64 {
	if paladin.Ranged().ID == LibramOfInvocation {
		return 0.95
	}
	return 1
}

// Holy Conduit takes 20% per point off Consecration, Holy Wrath, Exorcism and Hammer of Wrath.
func (paladin *Paladin) holyConduit() int32 {
	return []int32{100, 80, 60}[paladin.Talents.HolyConduit]
}

// holyConduitInstant is Holy Conduit's discount plus Benediction's, for the spells Holy
// Conduit covers when they are instant (Consecration and Exorcism). Benediction's text says "all
// instant cast spells and abilities". Percent cost modifiers multiply under Forever (0.6 * 0.9 =
// 54% at Holy Conduit 2 and Benediction 5). Hammer of Wrath is not on Benediction's list, even
// when Champion of the Light 5/5 makes it instant.
func (paladin *Paladin) holyConduitInstant() int32 {
	return paladin.holyConduit() * paladin.benediction() / 100
}

// Purifying Power shortens the Exorcism and Holy Wrath cooldowns by 17% per point.
func (paladin *Paladin) purifyingPower(cd time.Duration) time.Duration {
	return time.Duration(float64(cd) * []float64{1, 0.83, 0.67}[paladin.Talents.PurifyingPower])
}

// Redoubt now procs from any damaging melee attack taken (10% chance), Classic needed a
// crit. The block bonus lasts 10 sec or 5 blocks.
func (paladin *Paladin) applyRedoubt() {
	if paladin.Talents.Redoubt == 0 {
		return
	}

	// 6% a rank since the October 30 hotfix (spell 20127's trait ranks in 1.60.1.70707+hf, 6 to
	// 30%); builds 70245 to 70707 gave 4% a rank, Classic and the 69876 beta client 6%.
	blockBonus := 6.0 * float64(paladin.Talents.Redoubt) * core.BlockRatingPerBlockChance

	paladin.redoubtAura = paladin.RegisterAura(core.Aura{
		Label:     "Redoubt",
		ActionID:  core.ActionID{SpellID: 20128}, // the block aura, 10 sec or 5 blocks (20134 is no client spell)
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

// Reckoning gives a charge on 20% per point of the crits taken, and on 8% per point of the
// blocks, at most once every 1.5 sec (the 2026-09-12 beta notes, and the client's proc cooldown
// on spell 20177). Since client 1.60.1.70425 the charges stack up to 4 (spell 1326084:
// CumulativeAura 4, one charge, on a done melee auto attack) and the paladin's next auto attack
// spends them all as extra attacks: the core's stored extra attacks, which hold up to 4 and fire
// on the next main-hand swing. 70338 gave the extra attack at once (trigger 20178).
func (paladin *Paladin) applyReckoning() {
	if paladin.Talents.Reckoning == 0 {
		return
	}

	procID := core.ActionID{SpellID: 20178} // Reckoning Proc ID
	critChance := 0.2 * float64(paladin.Talents.Reckoning)
	blockChance := 0.08 * float64(paladin.Talents.Reckoning)
	// "Cannot occur from Blocks more than once every 1.5 sec" (20177, ProcCategoryRecovery 1500):
	// the cooldown is the blocks'. A crit taken gives its charge whenever it comes, as
	// wowsims/forever's talents_protection.go has it.
	icd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Millisecond * 1500,
	}

	// The stored extra attacks, registered up front: StoreExtraMHAttack would add the aura mid-fight.
	core.MakePermanent(paladin.GetOrRegisterAura(core.Aura{
		Label:     "Extra Attacks  (Main Hand)", // the same aura as Windfury Totem's, see core/buffs.go
		ActionID:  core.ActionID{SpellID: 21919},
		Duration:  core.NeverExpires,
		MaxStacks: 4,
	}))

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
			if result.DidCrit() && sim.Proc(critChance, "Reckoning") {
				paladin.AutoAttacks.StoreExtraMHAttack(sim, 1, procID, spell.ActionID)
			} else if result.DidBlock() && icd.IsReady(sim) && sim.Proc(blockChance, "Reckoning") {
				icd.Use(sim)
				paladin.AutoAttacks.StoreExtraMHAttack(sim, 1, procID, spell.ActionID)
			}
		},
	})
}

// Shield Specialization's second half: a block has a 50% chance a rank of 2 to restore 4% of the
// paladin's maximum mana (spell 20150's trait ranks in 1.60.1.70707+hf; it was 33% a rank of 3,
// client 1.60.1.70425, spell 1310925; 6% of the mana in 70338), once every 3 sec.
func (paladin *Paladin) applyShieldSpecializationMana() {
	if paladin.Talents.ShieldSpecialization == 0 {
		return
	}

	actionID := core.ActionID{SpellID: 20150}
	manaMetrics := paladin.NewManaMetrics(actionID)
	// The mana spell (1310925) has Attributes_1 0x400, no threat, so its mana makes none (Forever's
	// mana gains make half a threat a mana otherwise, core/mana.go).
	manaMetrics.NoThreat = true
	procChance := 0.5 * float64(paladin.Talents.ShieldSpecialization)
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
				paladin.AddMana(sim, 0.04*paladin.MaxMana(), manaMetrics)
			}
		},
	})

	// Shieldsworn's Holy Strike rolls the mana as a block would, without the 3 sec cooldown.
	paladin.shieldSpecializationMana = func(sim *core.Simulation) {
		if sim.Proc(procChance, "Shield Specialization") {
			paladin.AddMana(sim, 0.04*paladin.MaxMana(), manaMetrics)
		}
	}
}

// Two-Handed Weapon Specialization: 2 / 4 / 6% (since the 2026-09-24 beta build, 3 / 6 / 9%
// before). One-Handed Weapon Specialization left the tree with the October 30 hotfix.
//
// The hotfix rewrote the tooltip to "Increases the physical damage you deal by 6%, and reduces
// all threat you generate by 20% while a two-handed melee weapon is equipped" (spell 20111 in
// 1.60.1.70707+hf), but its damage effect is the row it was (effect 0, aura 79 on the Physical
// school), so it keeps reaching what it reached in game before: everything physical the weapon
// deals, plus the seal procs and judgements that roll as melee and Holy Strike (wowsims/forever
// d35afca4da measured Seal and Judgement of Righteousness gaining from it). Report assumption
// pal_2h_spec_holy.
func (paladin *Paladin) getWeaponSpecializationModifier() float64 {
	if paladin.MainHand().HandType == proto.HandType_HandTypeTwoHand {
		return 1 + 0.02*float64(paladin.Talents.TwoHandedWeaponSpecialization)
	}
	return 1
}

// Affects all physical damage or spells that can be rolled as physical. The threat cut is the
// hotfix's new effect 1 (aura 10, all schools, -20 at 3 ranks; the trait ranks 7 / 14 / 20%), a
// percent row of its own, so it multiplies with Righteous Fury's.
func (paladin *Paladin) applyWeaponSpecialization() {
	paladin.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= paladin.getWeaponSpecializationModifier()
	if paladin.has2hEquipped() && paladin.Talents.TwoHandedWeaponSpecialization > 0 {
		paladin.PseudoStats.ThreatMultiplier *= 1 - []float64{0, 0.07, 0.14, 0.20}[paladin.Talents.TwoHandedWeaponSpecialization]
	}
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

// Consecrated Ground: Holy spells deal 5% per point more to the first four enemies inside
// the paladin's Consecration. It is a Holy damage from caster modifier on Consecration's mark
// (client 1280353, aura 271 with the Holy school), which goes on those four, so a fifth enemy
// gets no bonus (wowsims/forever 7b094b56f0).
func (paladin *Paladin) applyConsecratedGround() {
	if paladin.Talents.ConsecratedGround == 0 {
		return
	}
	multiplier := 1 + 0.05*float64(paladin.Talents.ConsecratedGround)

	// The attack tables are made after the talents, and the marks with Consecration.
	paladin.Env.RegisterPostFinalizeEffect(func() {
		for _, target := range paladin.Env.Encounter.TargetUnits {
			mark := paladin.consecrationMarks.Get(target)
			for _, at := range paladin.AttackTables[target.UnitIndex] {
				if at == nil {
					continue
				}
				previous := at.DamageDoneByCasterMultiplier
				at.DamageDoneByCasterMultiplier = func(spell *core.Spell, at *core.AttackTable) float64 {
					m := 1.0
					if previous != nil {
						m = previous(spell, at)
					}
					if mark.IsActive() && spell.SpellSchool.Matches(core.SpellSchoolHoly) {
						m *= multiplier
					}
					return m
				}
			}
		}
	})
}
