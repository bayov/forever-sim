package warlock

import (
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The Forever warlock talents, from the beta client tree (tools/forever_talents/client,
// build 1.60.1.69876, checked against 1.60.1.70338). Per rank values are the client tooltips.
// Which spells a talent touches is the client's class mask on the talent's effect
// (wago.tools SpellEffect, 70338), and a percent on a spell's damage is a direct damage
// modifier (impact) or a periodic one depending on the effect, which add up with each other
// like 1.12's spell modifiers.

var (
	// Malediction, Pandemic and the periodic half of Shadow Mastery.
	afflictionPeriodicSpells = []int32{SpellCode_WarlockBaneOfAgony, SpellCode_WarlockBaneOfDoom, SpellCode_WarlockCorruption,
		SpellCode_WarlockDrainLife, SpellCode_WarlockDrainSoul, SpellCode_WarlockSiphonLife, SpellCode_WarlockWrack}
	// Malevolence.
	malevolenceSpells = append(slices.Clone(afflictionPeriodicSpells), SpellCode_WarlockDeathCoil, SpellCode_WarlockShadowBolt,
		SpellCode_WarlockShadowburn)
	// Shadow Mastery's direct half.
	shadowMasteryDirectSpells = []int32{SpellCode_WarlockShadowBolt, SpellCode_WarlockShadowburn, SpellCode_WarlockDeathCoil}
	// Ruin, Cataclysm and Agonizing Flames: the Destruction spells by the client's mask.
	destructionSpells = []int32{SpellCode_WarlockConflagrate, SpellCode_WarlockImmolate, SpellCode_WarlockIncinerate,
		SpellCode_WarlockSearingPain, SpellCode_WarlockShadowBolt, SpellCode_WarlockShadowburn, SpellCode_WarlockSoulFire}
	// Improved Drains and Soul Siphon.
	drainSpells = []int32{SpellCode_WarlockDrainLife, SpellCode_WarlockDrainSoul, SpellCode_WarlockWrack}
)

func (warlock *Warlock) onSpellCodes(codes []int32, apply func(spell *core.Spell)) {
	warlock.OnSpellRegistered(func(spell *core.Spell) {
		if spell.SpellCode != 0 && slices.Contains(codes, spell.SpellCode) {
			apply(spell)
		}
	})
}

func (warlock *Warlock) ApplyTalents() {
	warlock.applyWeaponImbue()

	// Affliction
	warlock.applySuppression()
	warlock.applyImprovedCorruption()
	warlock.applyMalediction()
	warlock.applyImprovedDrains()
	warlock.applyImprovedBaneOfAgony()
	warlock.applyPandemic()
	warlock.applyMalevolence()
	warlock.applyNightfall()
	warlock.applyShadowMastery()
	warlock.applyWrack()

	// Demonology
	warlock.applyDemonicEmbrace()
	warlock.applyUnholyPower()
	warlock.applyFelVitality()
	warlock.registerFelDominationCD()
	warlock.applyMasterSummoner()
	warlock.applyMasterDemonologist()
	warlock.applyDemonicSacrifice()
	warlock.applySoulLink()
	warlock.applyDemonicKnowledge()
	warlock.applyDecimation()
	warlock.applyDemonicBrand()

	// Destruction
	warlock.applyImprovedShadowBolt()
	warlock.applyBane()
	warlock.applyCataclysm()
	warlock.applyAftermath()
	warlock.applyRuin()
	warlock.applyAgonizingFlames()
	warlock.applyFireAndBrimstone()
	warlock.applyShadowAndFlame()
}

// Firestone and Spellstone are weapon imbues under Forever (the same enchant effect as the
// shaman imbues and rogue poisons, so they stack with an oil). Major Firestone (56): Fire
// damage up to 21 and 2% spell crit. Major Spellstone (60): Shadow damage up to 21 and 2%
// spell haste (client 23480 to 23483 and 1237162 to 1237165).
var firestoneRanks = []struct {
	level      int32
	damage     float64
	critOrHast float64
}{{28, 10, 1}, {36, 14, 1}, {46, 17, 2}, {56, 21, 2}}

var spellstoneRanks = []struct {
	level      int32
	damage     float64
	critOrHast float64
}{{36, 14, 1}, {48, 17, 2}, {60, 21, 2}}

func (warlock *Warlock) applyWeaponImbue() {
	switch warlock.Options.WeaponImbue {
	case proto.WarlockOptions_Firestone:
		for i := len(firestoneRanks) - 1; i >= 0; i-- {
			if r := firestoneRanks[i]; warlock.Level >= r.level {
				warlock.AddStat(stats.FirePower, r.damage)
				warlock.AddStat(stats.SpellCrit, r.critOrHast*core.SpellCritRatingPerCritChance)
				return
			}
		}
	case proto.WarlockOptions_Spellstone:
		for i := len(spellstoneRanks) - 1; i >= 0; i-- {
			if r := spellstoneRanks[i]; warlock.Level >= r.level {
				warlock.AddStat(stats.ShadowPower, r.damage)
				warlock.MultiplyCastSpeed(1 + r.critOrHast/100)
				return
			}
		}
	}
}

///////////////////////////////////////////////////////////////////////////
//                            Affliction
///////////////////////////////////////////////////////////////////////////

// Suppression is 1% hit a point with every spell and attack (and 4% less threat).
func (warlock *Warlock) applySuppression() {
	if warlock.Talents.Suppression == 0 {
		return
	}

	points := float64(warlock.Talents.Suppression)
	warlock.AddStat(stats.SpellHit, points*core.SpellHitRatingPerHitChance)
	warlock.AddStat(stats.MeleeHit, points*core.MeleeHitRatingPerHitChance)
	warlock.PseudoStats.ThreatMultiplier *= 1 - 0.04*points
}

// Improved Corruption: 0.4 sec off the cast (in getCorruptionConfig) and 2% periodic damage a point.
func (warlock *Warlock) applyImprovedCorruption() {
	if warlock.Talents.ImprovedCorruption == 0 {
		return
	}

	bonus := 0.02 * float64(warlock.Talents.ImprovedCorruption)
	warlock.onSpellCodes([]int32{SpellCode_WarlockCorruption}, func(spell *core.Spell) {
		spell.PeriodicDamageMultiplierAdditive += bonus
	})
}

// Malediction: 1% periodic damage a point on every warlock damage over time spell.
func (warlock *Warlock) applyMalediction() {
	if warlock.Talents.Malediction == 0 {
		return
	}

	bonus := 0.01 * float64(warlock.Talents.Malediction)
	warlock.onSpellCodes(append(slices.Clone(afflictionPeriodicSpells), SpellCode_WarlockImmolate), func(spell *core.Spell) {
		spell.PeriodicDamageMultiplierAdditive += bonus
	})
}

// Improved Drains: 7 / 13 / 20% on Drain Life, Drain Soul and Wrack.
func (warlock *Warlock) applyImprovedDrains() {
	if warlock.Talents.ImprovedDrains == 0 {
		return
	}

	bonus := []float64{0, 0.07, 0.13, 0.20}[warlock.Talents.ImprovedDrains]
	warlock.onSpellCodes(drainSpells, func(spell *core.Spell) {
		spell.PeriodicDamageMultiplierAdditive += bonus
	})
}

// Improved Bane of Agony: 5% a point.
func (warlock *Warlock) applyImprovedBaneOfAgony() {
	if warlock.Talents.ImprovedBaneOfAgony == 0 {
		return
	}

	bonus := 0.05 * float64(warlock.Talents.ImprovedBaneOfAgony)
	warlock.onSpellCodes([]int32{SpellCode_WarlockBaneOfAgony}, func(spell *core.Spell) {
		spell.PeriodicDamageMultiplierAdditive += bonus
	})
}

// Pandemic: 33 / 67 / 100% more critical strike bonus on the Affliction damage over time
// spells, so a crit at 3/3 doubles the tick like Ruin does for Destruction.
func (warlock *Warlock) applyPandemic() {
	if warlock.Talents.Pandemic == 0 {
		return
	}

	bonus := float64(warlock.Talents.Pandemic) / 3
	warlock.onSpellCodes(afflictionPeriodicSpells, func(spell *core.Spell) {
		spell.CritDamageBonus += bonus
	})
}

// Malevolence: 1% crit a point on the Shadow spells (the client's mask, Lash of Pain too, in
// registerSuccubusLashOfPainSpell).
func (warlock *Warlock) applyMalevolence() {
	if warlock.Talents.Malevolence == 0 {
		return
	}

	bonus := float64(warlock.Talents.Malevolence) * core.SpellCritRatingPerCritChance
	warlock.onSpellCodes(malevolenceSpells, func(spell *core.Spell) {
		spell.BonusCritRating += bonus
	})
}

// Nightfall: 2% a point on each tick of Corruption, Drain Soul, Drain Life and Wrack to make
// the next Shadow Bolt instant.
func (warlock *Warlock) applyNightfall() {
	if warlock.Talents.Nightfall <= 0 {
		return
	}

	shadowTranceAura := warlock.RegisterAura(core.Aura{
		Label:    "Nightfall Shadow Trance",
		ActionID: core.ActionID{SpellID: 17941},
		Duration: time.Second * 10,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.ShadowBolt {
				spell.CastTimeMultiplier -= 1
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.ShadowBolt {
				spell.CastTimeMultiplier += 1
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			// Check if the shadowbolt was instant cast and not a normal one
			if spell.SpellCode == SpellCode_WarlockShadowBolt && spell.CurCast.CastTime == 0 {
				aura.Deactivate(sim)
			}
		},
	})

	procChance := 0.02 * float64(warlock.Talents.Nightfall)
	procSpells := []int32{SpellCode_WarlockCorruption, SpellCode_WarlockDrainLife, SpellCode_WarlockDrainSoul, SpellCode_WarlockWrack}

	core.MakePermanent(warlock.RegisterAura(core.Aura{
		Label: "Nightfall Hidden Aura",
		OnPeriodicDamageDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if slices.Contains(procSpells, spell.SpellCode) && sim.Proc(procChance, "Nightfall") {
				shadowTranceAura.Activate(sim)
			}
		},
	}))
}

// Shadow Mastery: 1% a point, as a direct damage modifier on Shadow Bolt, Shadowburn and
// Death Coil and a periodic one on the Affliction damage over time spells.
func (warlock *Warlock) applyShadowMastery() {
	if warlock.Talents.ShadowMastery == 0 {
		return
	}

	bonus := 0.01 * float64(warlock.Talents.ShadowMastery)
	warlock.onSpellCodes(shadowMasteryDirectSpells, func(spell *core.Spell) {
		spell.ImpactDamageMultiplierAdditive += bonus
	})
	warlock.onSpellCodes(afflictionPeriodicSpells, func(spell *core.Spell) {
		spell.PeriodicDamageMultiplierAdditive += bonus
	})
}

// Wrack makes Corruption and Bane of Agony deal 10% more while it is channelled on the target.
func (warlock *Warlock) applyWrack() {
	if !warlock.Talents.Wrack {
		return
	}

	warlock.targetDamageModifiers = append(warlock.targetDamageModifiers, func(spell *core.Spell, target *core.Unit) float64 {
		if (spell.SpellCode == SpellCode_WarlockCorruption || spell.SpellCode == SpellCode_WarlockBaneOfAgony) && warlock.wrackActive(target) {
			return 1.1
		}
		return 1
	})
}

// soulSiphonMultiplier is Soul Siphon on the drains: 4% a point for each of the warlock's
// other Affliction effects on the target, up to 3 of them. We count Corruption, the bane,
// Siphon Life and the curse. The client leaves the count to the server, so which effects
// count is our reading of the tooltip, and we multiply it with the drains' other bonuses.
func (warlock *Warlock) soulSiphonMultiplier(target *core.Unit) float64 {
	if warlock.Talents.SoulSiphon == 0 {
		return 1
	}

	effects := 0
	for _, spell := range warlock.Corruption {
		if spell.Dot(target).IsActive() {
			effects++
			break
		}
	}
	if aura := warlock.ActiveBaneAura.Get(target); aura != nil && aura.IsActive() {
		effects++
	}
	for _, spell := range warlock.SiphonLife {
		if spell.Dot(target).IsActive() {
			effects++
			break
		}
	}
	if aura := warlock.ActiveCurseAura.Get(target); aura != nil && aura.IsActive() {
		effects++
	}

	return 1 + 0.04*float64(warlock.Talents.SoulSiphon)*float64(min(effects, 3))
}

///////////////////////////////////////////////////////////////////////////
//                            Demonology Talents
///////////////////////////////////////////////////////////////////////////

// Demonic Embrace: 3% Stamina a point. Forever dropped 1.12's Spirit cost from the tooltip.
// The client still has a Spirit effect on the talent with the server filling in its value,
// see notes.md.
func (warlock *Warlock) applyDemonicEmbrace() {
	if warlock.Talents.DemonicEmbrace == 0 {
		return
	}

	warlock.MultiplyStat(stats.Stamina, 1+.03*float64(warlock.Talents.DemonicEmbrace))
}

// Unholy Power: 2% a point on all damage the demons deal, spells too (1.12: 4% on melee).
// Improved Imp and Improved Sayaad are on the pet spells.
func (warlock *Warlock) applyUnholyPower() {
	if warlock.Talents.UnholyPower == 0 {
		return
	}

	multiplier := 1 + 0.02*float64(warlock.Talents.UnholyPower)
	for _, pet := range warlock.BasePets {
		pet.PseudoStats.DamageDealtMultiplier *= multiplier
	}
}

// Fel Vitality: 5% a point to the warlock's maximum mana and the demons' health and mana.
func (warlock *Warlock) applyFelVitality() {
	if warlock.Talents.FelVitality == 0 {
		return
	}

	multiplier := 1 + 0.05*float64(warlock.Talents.FelVitality)
	warlock.MultiplyStat(stats.Mana, multiplier)
	for _, pet := range warlock.BasePets {
		pet.MultiplyStat(stats.Mana, multiplier)
		pet.MultiplyStat(stats.Health, multiplier)
	}
}

func (warlock *Warlock) applyMasterSummoner() {
	if warlock.Talents.MasterSummoner == 0 {
		return
	}

	castTimeReduction := time.Second * 2 * time.Duration(warlock.Talents.MasterSummoner)
	costReduction := 20 * warlock.Talents.MasterSummoner

	// Use an aura because the summon spells aren't registered by this point
	warlock.RegisterAura(core.Aura{
		Label:    "Master Summoner Hidden Aura",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.SummonDemonSpells {
				spell.DefaultCast.CastTime -= castTimeReduction
				spell.Cost.Multiplier -= costReduction
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.SummonDemonSpells {
				spell.DefaultCast.CastTime += castTimeReduction
				spell.Cost.Multiplier += costReduction
			}
		},
	})
}

// Master Demonologist under Forever: the Imp gives 2% Fire damage a point and the Succubus 2%
// Shadow damage a point, to the warlock and the demon. The Voidwalker takes 2% a point off
// Physical damage taken and the Felhunter off Magic damage taken.
func (warlock *Warlock) applyMasterDemonologist() {
	if warlock.Talents.MasterDemonologist == 0 {
		return
	}

	points := float64(warlock.Talents.MasterDemonologist)
	damageDealtMultiplier := 1 + 0.02*points
	damageTakenMultiplier := 1 - 0.02*points
	magicSchools := []stats.SchoolIndex{stats.SchoolIndexArcane, stats.SchoolIndexFire, stats.SchoolIndexFrost, stats.SchoolIndexHoly,
		stats.SchoolIndexNature, stats.SchoolIndexShadow}

	schoolDamageConfig := func(tag int32, label string, school stats.SchoolIndex) core.Aura {
		return core.Aura{
			Label:    "Master Demonologist (" + label + ")",
			ActionID: core.ActionID{SpellID: 23825, Tag: tag},
			Duration: core.NeverExpires,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				aura.Unit.PseudoStats.SchoolDamageDealtMultiplier[school] *= damageDealtMultiplier
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				aura.Unit.PseudoStats.SchoolDamageDealtMultiplier[school] /= damageDealtMultiplier
			},
		}
	}

	impConfig := schoolDamageConfig(1, "Imp", stats.SchoolIndexFire)
	succubusConfig := schoolDamageConfig(3, "Succubus", stats.SchoolIndexShadow)

	voidwalkerConfig := core.Aura{
		Label:    "Master Demonologist (Voidwalker)",
		ActionID: core.ActionID{SpellID: 23825, Tag: 2},
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical] *= damageTakenMultiplier
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical] /= damageTakenMultiplier
		},
	}

	felhunterConfig := core.Aura{
		Label:    "Master Demonologist (Felhunter)",
		ActionID: core.ActionID{SpellID: 23825, Tag: 4},
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, school := range magicSchools {
				aura.Unit.PseudoStats.SchoolDamageTakenMultiplier[school] *= damageTakenMultiplier
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, school := range magicSchools {
				aura.Unit.PseudoStats.SchoolDamageTakenMultiplier[school] /= damageTakenMultiplier
			}
		},
	}

	for _, pair := range []struct {
		pet    *WarlockPet
		config core.Aura
	}{{warlock.Imp, impConfig}, {warlock.Voidwalker, voidwalkerConfig}, {warlock.Succubus, succubusConfig}, {warlock.Felhunter, felhunterConfig}} {
		warlockAura := warlock.RegisterAura(pair.config)
		petAura := pair.pet.RegisterAura(pair.config)
		pair.pet.ApplyOnPetEnable(func(sim *core.Simulation) {
			if warlock.MasterDemonologistAura != nil {
				warlock.MasterDemonologistAura.Deactivate(sim)
			}
			warlock.MasterDemonologistAura = warlockAura
			warlockAura.Activate(sim)
			petAura.Activate(sim)
		})
		pair.pet.ApplyOnPetDisable(func(sim *core.Simulation, isSacrifice bool) {
			petAura.Deactivate(sim)
			if warlock.MasterDemonologistAura == warlockAura {
				warlockAura.Deactivate(sim)
				warlock.MasterDemonologistAura = nil
			}
		})
	}
}

func (warlock *Warlock) applySoulLink() {
	if !warlock.Talents.SoulLink {
		return
	}

	actionID := core.ActionID{SpellID: 19028}
	soulLinkConfig := core.Aura{
		Label:    "Soul Link Aura",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.DamageTakenMultiplier /= 1.3
			aura.Unit.PseudoStats.DamageDealtMultiplier *= 1.03
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.DamageDealtMultiplier /= 1.03
			aura.Unit.PseudoStats.DamageTakenMultiplier *= 1.3
		},
	}

	warlock.SoulLinkAura = warlock.RegisterAura(soulLinkConfig)
	for _, pet := range warlock.BasePets {
		pet.SoulLinkAura = pet.RegisterAura(soulLinkConfig)

		oldOnPetDisable := pet.OnPetDisable
		pet.OnPetDisable = func(sim *core.Simulation, isSacrifice bool) {
			oldOnPetDisable(sim, isSacrifice)
			warlock.SoulLinkAura.Deactivate(sim)
			pet.SoulLinkAura.Deactivate(sim)
		}
	}

	warlock.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return warlock.ActivePet != nil
		},

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.2,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			warlock.SoulLinkAura.Activate(sim)
			warlock.ActivePet.SoulLinkAura.Activate(sim)
		},
	})
}

// Demonic Sacrifice under Forever swaps the schools and lasts 2 hours: the Imp gives 15%
// Shadow damage (Burning Shadow, 18789) and the Succubus 15% Fire damage (Touch of Fire,
// 18791), the Voidwalker 2% of maximum mana every 4 sec (Fel Energy, 18792) and the
// Felhunter 3% of maximum health (Fel Stamina, 18790). Summoning a demon cancels it, and with
// Demonic Pact only summoning the demon that was sacrificed does.
//
// WarlockOptions.sacrificed_pet sacrifices a demon before the pull, so a warlock with Demonic
// Pact can start with its buff and the demon of the Summon option out.
func (warlock *Warlock) applyDemonicSacrifice() {
	if !warlock.Talents.DemonicSacrifice {
		return
	}

	schoolAura := func(spellID int32, label string, school stats.SchoolIndex) *core.Aura {
		return warlock.GetOrRegisterAura(core.Aura{
			Label:    label,
			ActionID: core.ActionID{SpellID: spellID},
			Duration: 2 * time.Hour,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				warlock.PseudoStats.SchoolDamageDealtMultiplier[school] *= 1.15
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				warlock.PseudoStats.SchoolDamageDealtMultiplier[school] /= 1.15
			},
		})
	}

	impAura := schoolAura(18789, "Burning Shadow", stats.SchoolIndexShadow)
	succubusAura := schoolAura(18791, "Touch of Fire", stats.SchoolIndexFire)

	var vwPa *core.PendingAction
	manaMetric := warlock.NewManaMetrics(core.ActionID{SpellID: 18792})
	voidwalkerAura := warlock.GetOrRegisterAura(core.Aura{
		Label:    "Fel Energy",
		ActionID: core.ActionID{SpellID: 18792},
		Duration: 2 * time.Hour,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			vwPa = core.NewPeriodicAction(sim, core.PeriodicActionOptions{
				Period: time.Second * 4,
				OnAction: func(s *core.Simulation) {
					warlock.AddMana(sim, warlock.MaxMana()*0.02, manaMetric)
				},
			})
			sim.AddPendingAction(vwPa)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			vwPa.Cancel(sim)
		},
	})

	var fhPa *core.PendingAction
	healthMetric := warlock.NewHealthMetrics(core.ActionID{SpellID: 18790})
	felhunterAura := warlock.GetOrRegisterAura(core.Aura{
		Label:    "Fel Stamina",
		ActionID: core.ActionID{SpellID: 18790},
		Duration: 2 * time.Hour,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			fhPa = core.NewPeriodicAction(sim, core.PeriodicActionOptions{
				Period: time.Second * 4,
				OnAction: func(s *core.Simulation) {
					warlock.GainHealth(sim, warlock.MaxHealth()*0.03, healthMetric)
				},
			})
			sim.AddPendingAction(fhPa)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			fhPa.Cancel(sim)
		},
	})

	warlock.DemonicSacrificeAuras = map[*WarlockPet]*core.Aura{
		warlock.Imp:        impAura,
		warlock.Succubus:   succubusAura,
		warlock.Voidwalker: voidwalkerAura,
		warlock.Felhunter:  felhunterAura,
	}

	for _, pet := range warlock.BasePets {
		pet.ApplyOnPetEnable(func(sim *core.Simulation) {
			for sacrificed, dsAura := range warlock.DemonicSacrificeAuras {
				if warlock.Talents.DemonicPact && sacrificed != pet {
					continue
				}
				dsAura.Deactivate(sim)
			}
		})
	}

	sacrificedPet := map[proto.WarlockOptions_Summon]*WarlockPet{
		proto.WarlockOptions_Imp:        warlock.Imp,
		proto.WarlockOptions_Voidwalker: warlock.Voidwalker,
		proto.WarlockOptions_Succubus:   warlock.Succubus,
		proto.WarlockOptions_Felhunter:  warlock.Felhunter,
	}[warlock.Options.SacrificedPet]
	if sacrificedPet != nil {
		core.MakePermanent(warlock.RegisterAura(core.Aura{
			Label: "Demonic Sacrifice Before Pull",
			OnReset: func(aura *core.Aura, sim *core.Simulation) {
				core.StartDelayedAction(sim, core.DelayedActionOptions{
					DoAt:     sim.CurrentTime,
					Priority: core.ActionPriorityAuto,
					OnAction: func(sim *core.Simulation) {
						// Without Demonic Pact the demon that is out would have cancelled it.
						if warlock.ActivePet == sacrificedPet || (warlock.ActivePet != nil && !warlock.Talents.DemonicPact) {
							return
						}
						warlock.DemonicSacrificeAuras[sacrificedPet].Activate(sim)
					},
				})
			},
		}))
	}

	warlock.GetOrRegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_WarlockDemonicSacrifice,
		ActionID:    core.ActionID{SpellID: 18788},
		SpellSchool: core.SpellSchoolShadow,
		Flags:       core.SpellFlagAPL,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return warlock.ActivePet != nil
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for _, dsAura := range warlock.DemonicSacrificeAuras {
				dsAura.Deactivate(sim)
			}
			warlock.DemonicSacrificeAuras[warlock.ActivePet].Activate(sim)
			warlock.changeActivePet(sim, nil, true)
		},
	})
}

// Demonic Knowledge: spell damage of 33 / 67 / 100% of the warlock's level (60 at 3/3 and 60)
// for the warlock and the demon while a demon is out.
func (warlock *Warlock) applyDemonicKnowledge() {
	if warlock.Talents.DemonicKnowledge == 0 {
		return
	}

	bonus := []float64{0, 0.33, 0.67, 1}[warlock.Talents.DemonicKnowledge] * float64(warlock.Level)
	config := core.Aura{
		Label:    "Demonic Knowledge",
		ActionID: core.ActionID{SpellID: 412732},
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatDynamic(sim, stats.SpellDamage, bonus)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatDynamic(sim, stats.SpellDamage, -bonus)
		},
	}

	warlockAura := warlock.RegisterAura(config)
	for _, pet := range warlock.BasePets {
		petAura := pet.RegisterAura(config)
		pet.ApplyOnPetEnable(func(sim *core.Simulation) {
			warlockAura.Activate(sim)
			petAura.Activate(sim)
		})
		pet.ApplyOnPetDisable(func(sim *core.Simulation, isSacrifice bool) {
			warlockAura.Deactivate(sim)
			petAura.Deactivate(sim)
		})
	}
}

// Decimation: Shadow Bolt and Searing Pain deal 3% more a point to a target under 35% health,
// and casting one at it takes 20% a point off the next 10 sec of Soul Fire casts (and their
// shard, which we don't track). The cooldown part is in getSoulFireBaseConfig. The damage part
// is the tooltip's ("they deal 3% increased damage"), the client leaves it to the server.
func (warlock *Warlock) applyDecimation() {
	if warlock.Talents.Decimation == 0 {
		return
	}

	points := float64(warlock.Talents.Decimation)
	castMultiplier := 0.2 * points
	warlock.DecimationAura = warlock.RegisterAura(core.Aura{
		Label:    "Decimation",
		ActionID: core.ActionID{SpellID: 440873},
		Duration: time.Second * 10,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.SoulFire {
				spell.CastTimeMultiplier -= castMultiplier
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.SoulFire {
				spell.CastTimeMultiplier += castMultiplier
			}
		},
	})

	core.MakePermanent(warlock.RegisterAura(core.Aura{
		Label: "Decimation Trigger",
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if (spell.SpellCode == SpellCode_WarlockShadowBolt || spell.SpellCode == SpellCode_WarlockSearingPain) && sim.IsExecutePhase35() {
				warlock.DecimationAura.Activate(sim)
			}
		},
	}))

	warlock.targetDamageModifiers = append(warlock.targetDamageModifiers, func(spell *core.Spell, target *core.Unit) float64 {
		if (spell.SpellCode == SpellCode_WarlockShadowBolt || spell.SpellCode == SpellCode_WarlockSearingPain) && warlock.sim.IsExecutePhase35() {
			return 1 + 0.03*points
		}
		return 1
	})
}

// Demonic Brand: Searing Pain brands the target for 10 sec, and the demon's next 2 / 4 / 6
// hits on it deal 65 to 68 more, Fire for the Imp and Shadow for the others. The damage is
// the 69876 client's tooltip, 70338 shows a placeholder.
func (warlock *Warlock) applyDemonicBrand() {
	if warlock.Talents.DemonicBrand == 0 {
		return
	}

	charges := 2 * warlock.Talents.DemonicBrand
	warlock.DemonicBrandAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(core.Aura{
			Label:     "Demonic Brand-" + warlock.Label,
			ActionID:  core.ActionID{SpellID: 1293696},
			Duration:  time.Second * 10,
			MaxStacks: charges,
		})
	})

	for _, pet := range warlock.BasePets {
		school := core.SpellSchoolShadow
		if pet == warlock.Imp {
			school = core.SpellSchoolFire
		}
		brandSpell := pet.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 1293695},
			SpellSchool: school,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagPassiveSpell,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, sim.Roll(65, 68), spell.OutcomeMagicCrit)
			},
		})

		core.MakePermanent(pet.RegisterAura(core.Aura{
			Label: "Demonic Brand Trigger",
			OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if spell == brandSpell || !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMeleeMH|core.ProcMaskSpellDamage) {
					return
				}
				brand := warlock.DemonicBrandAuras.Get(result.Target)
				if !brand.IsActive() {
					return
				}
				brand.RemoveStack(sim)
				brandSpell.Cast(sim, result.Target)
			},
		}))
	}
}

// applyDemonicBrand on a landed Searing Pain.
func (warlock *Warlock) applyDemonicBrandTo(sim *core.Simulation, target *core.Unit) {
	if warlock.DemonicBrandAuras == nil {
		return
	}
	brand := warlock.DemonicBrandAuras.Get(target)
	brand.Activate(sim)
	brand.SetStacks(sim, brand.MaxStacks)
}

///////////////////////////////////////////////////////////////////////////
//                            Destruction Talents
///////////////////////////////////////////////////////////////////////////

// Improved Shadow Bolt under Forever: a Shadow Bolt crit makes the target take 4% more Shadow
// damage a point from the warlock for 12 sec (Shadow Vulnerability, 17794, aura 270 "damage
// taken from caster"), with no charges and nothing for the raid.
func (warlock *Warlock) applyImprovedShadowBolt() {
	if warlock.Talents.ImprovedShadowBolt == 0 {
		return
	}

	multiplier := 1 + 0.04*float64(warlock.Talents.ImprovedShadowBolt)
	warlock.ImprovedShadowBoltAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(core.Aura{
			Label:    "Shadow Vulnerability-" + warlock.Label,
			ActionID: core.ActionID{SpellID: 17794},
			Duration: time.Second * 12,
		})
	})

	core.MakePermanent(warlock.RegisterAura(core.Aura{
		Label: "ISB Trigger",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.SpellCode == SpellCode_WarlockShadowBolt && result.DidCrit() {
				warlock.ImprovedShadowBoltAuras.Get(result.Target).Activate(sim)
			}
		},
	}))

	warlock.targetDamageModifiers = append(warlock.targetDamageModifiers, func(spell *core.Spell, target *core.Unit) float64 {
		if spell.SpellSchool.Matches(core.SpellSchoolShadow) && warlock.ImprovedShadowBoltAuras.Get(target).IsActive() {
			return multiplier
		}
		return 1
	})
}

func (warlock *Warlock) applyBane() {
	if warlock.Talents.Bane == 0 {
		return
	}

	points := time.Duration(warlock.Talents.Bane)
	warlock.OnSpellRegistered(func(spell *core.Spell) {
		switch spell.SpellCode {
		case SpellCode_WarlockShadowBolt, SpellCode_WarlockImmolate, SpellCode_WarlockIncinerate:
			spell.DefaultCast.CastTime -= time.Millisecond * 100 * points
		case SpellCode_WarlockSoulFire:
			spell.DefaultCast.CastTime -= time.Millisecond * 400 * points
		}
	})
}

// Cataclysm: 3 / 6 / 10% off the cost of the Destruction spells.
func (warlock *Warlock) applyCataclysm() {
	if warlock.Talents.Cataclysm == 0 {
		return
	}

	reduction := []int32{0, 3, 6, 10}[warlock.Talents.Cataclysm]
	warlock.onSpellCodes(destructionSpells, func(spell *core.Spell) {
		if spell.Cost != nil {
			spell.Cost.Multiplier -= reduction
		}
	})
}

// Aftermath: 10% a point on Immolate's direct damage (its Daze does nothing for damage).
func (warlock *Warlock) applyAftermath() {
	if warlock.Talents.Aftermath == 0 {
		return
	}

	bonus := 0.1 * float64(warlock.Talents.Aftermath)
	warlock.onSpellCodes([]int32{SpellCode_WarlockImmolate}, func(spell *core.Spell) {
		spell.ImpactDamageMultiplierAdditive += bonus
	})
}

// Ruin: 20% more critical strike bonus a point on the Destruction spells (1.12: 100% in one).
func (warlock *Warlock) applyRuin() {
	if warlock.Talents.Ruin == 0 {
		return
	}

	bonus := 0.2 * float64(warlock.Talents.Ruin)
	warlock.onSpellCodes(destructionSpells, func(spell *core.Spell) {
		spell.CritDamageBonus += bonus
	})
}

// Agonizing Flames: 3 / 7 / 10% damage on the Destruction spells (Immolate's ticks too) and
// the same in crit on Searing Pain.
func (warlock *Warlock) applyAgonizingFlames() {
	if warlock.Talents.AgonizingFlames == 0 {
		return
	}

	bonus := []float64{0, 0.03, 0.07, 0.10}[warlock.Talents.AgonizingFlames]
	warlock.onSpellCodes(destructionSpells, func(spell *core.Spell) {
		spell.ImpactDamageMultiplierAdditive += bonus
		if spell.SpellCode == SpellCode_WarlockImmolate {
			spell.PeriodicDamageMultiplierAdditive += bonus
		}
		if spell.SpellCode == SpellCode_WarlockSearingPain {
			spell.BonusCritRating += 100 * bonus * core.SpellCritRatingPerCritChance
		}
	})
}

// Fire and Brimstone: 8 / 17 / 25% crit on Conflagrate.
func (warlock *Warlock) applyFireAndBrimstone() {
	if warlock.Talents.FireAndBrimstone == 0 {
		return
	}

	bonus := []float64{0, 8, 17, 25}[warlock.Talents.FireAndBrimstone] * core.SpellCritRatingPerCritChance
	warlock.onSpellCodes([]int32{SpellCode_WarlockConflagrate}, func(spell *core.Spell) {
		spell.BonusCritRating += bonus
	})
}

// Shadow and Flame: a Conflagrate hit gives 2% Shadow damage a point for 20 sec, a Shadowburn
// hit 2% Fire damage a point. Its 20% a point to keep Immolate is in getConflagrateConfig, the
// Soul Shard refund does nothing in the sim.
func (warlock *Warlock) applyShadowAndFlame() {
	if warlock.Talents.ShadowAndFlame == 0 {
		return
	}

	multiplier := 1 + 0.02*float64(warlock.Talents.ShadowAndFlame)
	schoolAura := func(spellID int32, label string, school stats.SchoolIndex) *core.Aura {
		return warlock.RegisterAura(core.Aura{
			Label:    label,
			ActionID: core.ActionID{SpellID: spellID},
			Duration: time.Second * 20,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				warlock.PseudoStats.SchoolDamageDealtMultiplier[school] *= multiplier
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				warlock.PseudoStats.SchoolDamageDealtMultiplier[school] /= multiplier
			},
		})
	}
	shadowAura := schoolAura(1293816, "Shadow and Flame (Shadow)", stats.SchoolIndexShadow)
	fireAura := schoolAura(426311, "Shadow and Flame (Fire)", stats.SchoolIndexFire)

	core.MakePermanent(warlock.RegisterAura(core.Aura{
		Label: "Shadow and Flame Trigger",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() {
				return
			}
			switch spell.SpellCode {
			case SpellCode_WarlockConflagrate:
				shadowAura.Activate(sim)
			case SpellCode_WarlockShadowburn:
				fireAura.Activate(sim)
			}
		},
	}))
}

// applyTargetDamageModifiers hooks the warlock's own target side bonuses into the attack
// tables, so they count on every tick while the target has them and not on the snapshot.
func (warlock *Warlock) applyTargetDamageModifiers() {
	if len(warlock.targetDamageModifiers) == 0 {
		return
	}

	modifier := func(spell *core.Spell, attackTable *core.AttackTable) float64 {
		multiplier := 1.0
		for _, mod := range warlock.targetDamageModifiers {
			multiplier *= mod(spell, attackTable.Defender)
		}
		return multiplier
	}
	for _, tables := range warlock.AttackTables {
		for _, table := range tables {
			if table != nil {
				table.DamageDoneByCasterMultiplier = modifier
			}
		}
	}
}
