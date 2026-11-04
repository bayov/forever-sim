package priest

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// The priest's heals (docs/healing_priest.md). A heal cast at a player lands on that player (a
// healer's rotation aims them with Most Hurt, Lowest Health or a tank, docs/healing_sim.md); one
// cast at a mob, the rotation's default target, lands on the priest.
//
// Every value is the Forever client's (1.60.1.70753, the November 2 build with the October 30
// hotfix: SpellEffect, SpellLevels, SpellMisc, SpellPower and SpellCastTimes on wago.tools), read
// the way the shaman's heals are (sim/shaman/healing.go): a rank heals its base at the level it's
// learned plus its growth per level up to its max level, spread evenly by the variance around
// that. The coefficients are the client's per rank, and Forever's downranking (sim/core/downrank.go,
// from each spell's Rank and RequiredLevel, or DownrankLevel on a rank-less sub-spell) takes them
// down for a rank far below our level.
type healRank struct {
	spellID              int32
	base, perLevel       float64
	variance             float64
	learnLevel, maxLevel int32
	coefficient          float64
	castTime             time.Duration
	manaCost             float64
}

// SpellID is the rank's spell ID.
func (r healRank) SpellID() int32 {
	return r.spellID
}

// Coefficient is the rank's spell power coefficient before downranking.
func (r healRank) Coefficient() float64 {
	return r.coefficient
}

// LearnLevel is the level the rank is learned at.
func (r healRank) LearnLevel() int32 {
	return r.learnLevel
}

// At is the rank's lowest and highest heal at the character's level.
func (r healRank) At(level int32) (float64, float64) {
	value := r.base + r.perLevel*float64(max(0, min(level, r.maxLevel)-r.learnLevel))
	return value * (1 - r.variance/2), value * (1 + r.variance/2)
}

var LesserHealRanks = []healRank{
	{},
	{2050, 51, 0.3000000119, 0.1960784346, 1, 3, 0.4289999902, 1500 * time.Millisecond, 30},
	{2052, 94, 0.6999999881, 0.1794871837, 4, 9, 0.57099998, 2000 * time.Millisecond, 45},
	{2053, 163, 1.200000048, 0.1506849378, 10, 15, 0.7139999866, 2500 * time.Millisecond, 75},
}

var HealRanks = []healRank{
	{},
	{2054, 270, 2, 0.1446540952, 16, 21, 0.8569999933, 3000 * time.Millisecond, 155},
	{2055, 375, 2.900000095, 0.1347826123, 22, 27, 0.8569999933, 3000 * time.Millisecond, 205},
	{6063, 523, 4, 0.1258278191, 28, 33, 0.8569999933, 3000 * time.Millisecond, 255},
	{6064, 651, 4.5, 0.1213720292, 34, 39, 0.8569999933, 3000 * time.Millisecond, 305},
}

var GreaterHealRanks = []healRank{
	{},
	{2060, 844, 5.099999905, 0.1192468628, 40, 45, 0.8569999933, 3000 * time.Millisecond, 370},
	{10963, 1099, 5.800000191, 0.1148482263, 46, 51, 0.8569999933, 3000 * time.Millisecond, 455},
	{10964, 1403, 6.599999905, 0.1129349992, 52, 57, 0.8569999933, 3000 * time.Millisecond, 545},
	{10965, 1782, 7.5, 0.1093585715, 58, 63, 0.8569999933, 3000 * time.Millisecond, 655},
	{25314, 1960, 8.100000381, 0.109615393, 60, 65, 0.8569999933, 3000 * time.Millisecond, 710},
}

var FlashHealRanks = []healRank{
	{},
	{2061, 194, 1.899999976, 0.204651162, 20, 25, 0.4289999902, 1500 * time.Millisecond, 125},
	{9472, 249, 2.200000048, 0.1958041936, 26, 31, 0.4289999902, 1500 * time.Millisecond, 155},
	{9473, 312, 2.5, 0.1833333373, 32, 37, 0.4289999902, 1500 * time.Millisecond, 185},
	{9474, 383, 2.799999952, 0.1776765436, 38, 43, 0.4289999902, 1500 * time.Millisecond, 215},
	{10915, 507, 3.299999952, 0.1728395075, 44, 49, 0.4289999902, 1500 * time.Millisecond, 265},
	{10916, 644, 3.700000048, 0.1704545468, 50, 55, 0.4289999902, 1500 * time.Millisecond, 315},
	{10917, 825, 4.199999809, 0.1649717539, 56, 61, 0.4289999902, 1500 * time.Millisecond, 380},
}

// Prayer of Healing heals its target and the target's party (the deep dive of 2026-09-30: "heals
// all the members of the target's party", ImplicitTarget 132/31 within 40 yards). Forever's ranks
// heal each member far less than Classic's (rank 5: 645 to 681 against 1041 to 1099) for the
// same mana.
var PrayerOfHealingRanks = []healRank{
	{},
	{596, 178, 0.8000000119, 0.06430868059, 30, 39, 0.2860000134, 3000 * time.Millisecond, 410},
	{996, 265, 1, 0.06113537028, 40, 49, 0.2860000134, 3000 * time.Millisecond, 560},
	{10960, 401, 1.299999952, 0.05621301755, 50, 59, 0.2860000134, 3000 * time.Millisecond, 770},
	{10961, 583, 1.5, 0.05388601124, 60, 69, 0.2860000134, 3000 * time.Millisecond, 1030},
	{25316, 649, 1.600000024, 0.05420560762, 60, 69, 0.2860000134, 3000 * time.Millisecond, 1070},
}

// Holy Nova (the Holy talent): instant, heals the priest's party (ImplicitTarget 20, 10 yards;
// everyone is in reach for now). The heal is its triggered spell's (23455 and so on); the
// damage part on enemies within 10 yards is left out, as a healer stands back.
var HolyNovaRanks = []healRank{
	{},
	{15237, 51, 0.400000006, 0.1428571492, 20, 26, 0.1070000008, 0, 185},
	{15430, 82, 0.5, 0.1304347813, 28, 34, 0.1070000008, 0, 290},
	{15431, 116, 0.6000000238, 0.1384615451, 36, 42, 0.1070000008, 0, 400},
	{27799, 159, 0.6999999881, 0.1494252831, 44, 50, 0.1070000008, 0, 520},
	{27800, 238, 0.8000000119, 0.1422924846, 52, 58, 0.1070000008, 0, 635},
	{27801, 311, 0.8999999762, 0.1472392678, 60, 66, 0.1070000008, 0, 750},
}

// Renew heals its base every 3 sec for 15 sec, with a coefficient of 0.2 a tick. Forever's
// rank 10 heals 166 a tick (Classic 194). Its ticks can crit since client 70009 (Attributes_8
// 0x200), and several priests' Renews can share a target (the deep dive).
var RenewRanks = []healRank{
	{},
	{139, 9, 0, 0, 8, 13, 0.200000003, 0, 30},
	{6074, 15, 0, 0, 14, 19, 0.200000003, 0, 65},
	{6075, 25, 0, 0, 20, 25, 0.200000003, 0, 105},
	{6076, 32, 0, 0, 26, 31, 0.200000003, 0, 140},
	{6077, 41, 0, 0, 32, 37, 0.200000003, 0, 170},
	{6078, 54, 0, 0, 38, 43, 0.200000003, 0, 205},
	{10927, 74, 0, 0, 44, 49, 0.200000003, 0, 250},
	{10928, 102, 0, 0, 50, 55, 0.200000003, 0, 305},
	{10929, 134, 0, 0, 56, 61, 0.200000003, 0, 365},
	{25315, 166, 0, 0, 60, 65, 0.200000003, 0, 410},
}

const (
	RenewTicks      = 5
	RenewTickLength = 3 * time.Second
)

// Power Word: Shield absorbs every school (misc 127) for 30 sec, at a coefficient of 0.1 on
// every rank. The ranks share a 4 sec cooldown (category recovery 4000), which Soul Warding
// takes off. It leaves Weakened Soul on its target for 15 sec, and a shield can't go on a target
// with Weakened Soul. Since client 70009 a new shield may replace one still up (Attributes_4
// 0x100000) when Weakened Soul is gone.
var PowerWordShieldRanks = []healRank{
	{},
	{17, 44, 0.8000000119, 0, 6, 11, 0.1000000015, 0, 45},
	{592, 86, 1.200000048, 0, 12, 17, 0.1000000015, 0, 80},
	{600, 154, 1.600000024, 0, 18, 23, 0.1000000015, 0, 130},
	{3747, 226, 2, 0, 24, 29, 0.1000000015, 0, 175},
	{6065, 291, 2.299999952, 0, 30, 35, 0.1000000015, 0, 210},
	{6066, 368, 2.599999905, 0, 36, 41, 0.1000000015, 0, 250},
	{10898, 470, 3, 0, 42, 47, 0.1000000015, 0, 300},
	{10899, 591, 3.400000095, 0, 48, 53, 0.1000000015, 0, 355},
	{10900, 749, 3.900000095, 0, 54, 59, 0.1000000015, 0, 425},
	{10901, 928, 4.300000191, 0, 60, 65, 0.1000000015, 0, 500},
}

const (
	PowerWordShieldDuration  = 30 * time.Second
	PowerWordShieldCooldown  = 4 * time.Second
	WeakenedSoulDuration     = 15 * time.Second
	WeakenedSoulSpellID      = 6788
	PowerWordShieldAuraLabel = "Power Word: Shield"
)

// Prayer of Mending (the 31 point Holy talent, 401859 and its ranks 1240826 and 1240827): it sits
// on its target for 30 sec and heals it the next time it takes damage or a heal that isn't a
// tick, then jumps to another player in the raid (within 20 yards: anyone, for now) for 30 sec
// more, healing 5 times in all (the deep dive: "until it heals a total of 5 times"). One priest
// has one up at a time, and a player carries only one. 10 sec cooldown. The heal is the rank's
// dummy base at a 0.429 coefficient, and it can crit.
var PrayerOfMendingRanks = []healRank{
	{},
	{401859, 172, 0, 0, 40, 40, 0.4289999902, 0, 210},
	{1240826, 298, 0, 0, 50, 50, 0.4289999902, 0, 305},
	{1240827, 413, 0, 0, 60, 60, 0.4289999902, 0, 390},
}

const (
	PrayerOfMendingCharges  = 5
	PrayerOfMendingDuration = 30 * time.Second
	PrayerOfMendingCooldown = 10 * time.Second
	PrayerOfMendingHealID   = 401863
)

func init() {
	for _, ranks := range [][]healRank{LesserHealRanks, HealRanks, GreaterHealRanks, FlashHealRanks,
		PrayerOfHealingRanks, HolyNovaRanks, RenewRanks, PowerWordShieldRanks, PrayerOfMendingRanks} {
		core.RegisterSpellRanks(healRankIDs(ranks)...)
	}
}

// healRankIDs are the spell IDs of every rank of a heal, for its rank family.
func healRankIDs(ranks []healRank) []int32 {
	ids := make([]int32, 0, len(ranks)-1)
	for _, r := range ranks[1:] {
		ids = append(ids, r.spellID)
	}
	return ids
}

// RegisterHealingSpells registers the heals, for the healing priest only, so the Shadow priest's
// spellbook and numbers stay as they are.
func (priest *Priest) RegisterHealingSpells() {
	priest.registerHealingTalentAuras()
	if priest.Talents.LitanyOfLight > 0 {
		priest.litanyMetrics = priest.NewManaMetrics(core.ActionID{SpellID: 1317006})
	}

	priest.LesserHeal = priest.registerHealRanks(LesserHealRanks, SpellCode_PriestLesserHeal)
	priest.Heal = priest.registerHealRanks(HealRanks, SpellCode_PriestHeal)
	priest.GreaterHeal = priest.registerHealRanks(GreaterHealRanks, SpellCode_PriestGreaterHeal)
	priest.FlashHeal = priest.registerHealRanks(FlashHealRanks, SpellCode_PriestFlashHeal)
	priest.PrayerOfHealing = priest.registerHealRanks(PrayerOfHealingRanks, SpellCode_PriestPrayerOfHealing)
	if priest.Talents.HolyNova {
		priest.HolyNova = priest.registerHealRanks(HolyNovaRanks, SpellCode_PriestHolyNova)
	}
	priest.registerRenewSpell()
	priest.registerPowerWordShieldSpell()
	priest.registerPrayerOfMendingSpell()
	priest.registerHolyBinding()
}

func (priest *Priest) registerHealRanks(ranks []healRank, spellCode int32) []*core.Spell {
	spells := make([]*core.Spell, len(ranks))
	for rank := 1; rank < len(ranks); rank++ {
		if priest.KnowsRank(ranks[rank].spellID, ranks[rank].learnLevel) {
			spells[rank] = priest.registerHealRank(ranks[rank], rank, spellCode)
		}
	}
	return spells
}

func (priest *Priest) registerHealRank(r healRank, rank int, spellCode int32) *core.Spell {
	low, high := r.At(priest.Level)

	castTime := r.castTime
	if spellCode == SpellCode_PriestHeal || spellCode == SpellCode_PriestGreaterHeal {
		// Divine Fury: 0.1 sec a rank off Heal and Greater Heal (and Smite and Holy Fire).
		castTime -= 100 * time.Millisecond * time.Duration(priest.Talents.DivineFury)
	}
	gcd := core.GCDDefault

	var spell *core.Spell
	config := core.SpellConfig{
		SpellCode:   spellCode,
		ActionID:    core.ActionID{SpellID: r.spellID},
		SpellSchool: core.SpellSchoolHoly,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       SpellFlagPriest | core.SpellFlagHelpful | core.SpellFlagAPL,

		RequiredLevel: int(r.learnLevel),
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: r.manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      gcd,
				CastTime: castTime,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: r.coefficient,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			target = priest.healTarget(target)
			// A heal at a player who died as it landed does nothing.
			if target.IsDead() && spellCode != SpellCode_PriestHolyNova {
				return
			}
			priest.litanyOfLight(sim, spell, r.manaCost)
			heal := func(unit *core.Unit) {
				if unit.IsDead() {
					return
				}
				priest.directHeal(sim, spell, unit, sim.Roll(low, high))
			}
			switch spellCode {
			case SpellCode_PriestPrayerOfHealing:
				for _, unit := range priest.partyOf(target) {
					heal(unit)
				}
			case SpellCode_PriestHolyNova:
				for _, unit := range priest.partyOf(&priest.Unit) {
					heal(unit)
				}
			default:
				heal(target)
			}
		},
		// Other healers see the heal in flight at its average on its target (heal_comm.go):
		// Prayer of Healing's other targets aren't counted, as Chain Heal's jumps aren't.
		ExpectedHealing: func(target *core.Unit) float64 {
			return spell.ExpectedHealing(priest.healTarget(target), (low+high)/2)
		},
	}
	spell = priest.RegisterSpell(config)
	return spell
}

// directHeal is one heal that isn't a tick on unit, with what follows it: Renewed Hope's crit and
// Weakened Soul cut, Divine Aegis and Inspiration on a crit, and Holy Binding.
func (priest *Priest) directHeal(sim *core.Simulation, spell *core.Spell, unit *core.Unit, base float64) *core.SpellResult {
	critBonus := priest.renewedHopeCrit(spell, unit)
	spell.BonusCritRating += critBonus
	result := spell.CalcAndDealHealing(sim, unit, base, spell.OutcomeHealingCrit)
	spell.BonusCritRating -= critBonus
	priest.afterDirectHeal(sim, spell, result)
	return result
}

// healTarget is where a heal cast at target lands: the player it was cast at, or the priest for a
// cast at a mob or at nothing.
func (priest *Priest) healTarget(target *core.Unit) *core.Unit {
	if target == nil || target.Type == core.EnemyUnit {
		return &priest.Unit
	}
	return target
}

// partyOf is the players in unit's party, unit first, or unit alone when it isn't a raid player.
func (priest *Priest) partyOf(unit *core.Unit) []*core.Unit {
	agent := priest.Env.Raid.GetPlayerFromUnit(unit)
	if agent == nil {
		return []*core.Unit{unit}
	}
	units := []*core.Unit{unit}
	for _, member := range agent.GetCharacter().Party.Players {
		if _, ok := member.(*core.TargetDummy); ok {
			continue
		}
		if u := &member.GetCharacter().Unit; u != unit {
			units = append(units, u)
		}
	}
	return units
}

func (priest *Priest) registerRenewSpell() {
	priest.Renew = make([]*core.Spell, len(RenewRanks))
	for rank := 1; rank < len(RenewRanks); rank++ {
		r := RenewRanks[rank]
		if !priest.KnowsRank(r.spellID, r.learnLevel) {
			continue
		}
		rank := rank
		priest.Renew[rank] = priest.RegisterSpell(core.SpellConfig{
			SpellCode:   SpellCode_PriestRenew,
			ActionID:    core.ActionID{SpellID: r.spellID},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       SpellFlagPriest | core.SpellFlagHelpful | core.SpellFlagPureDot | core.SpellFlagAPL,

			RequiredLevel: int(r.learnLevel),
			Rank:          rank,

			ManaCost: core.ManaCostOptions{
				FlatCost: r.manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
			},

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			Hot: core.DotConfig{
				Aura: core.Aura{
					Label: fmt.Sprintf("Renew (Rank %d)", rank),
				},
				NumberOfTicks:    RenewTicks,
				TickLength:       RenewTickLength,
				BonusCoefficient: r.coefficient,
				OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
					dot.SnapshotHeal(target, r.base, isRollover)
				},
				OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
					dot.CalcAndDealPeriodicSnapshotHealing(sim, target, dot.OutcomeTick)
				},
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				target = priest.healTarget(target)
				if target.IsDead() {
					return
				}
				priest.litanyOfLight(sim, spell, r.manaCost)
				// Our Renew of another rank on the target goes: one priest has one Renew on a target.
				for other, s := range priest.Renew {
					if other != rank && s != nil && s.Hot(target).IsActive() {
						s.Hot(target).Deactivate(sim)
					}
				}
				spell.SpellMetrics[target.UnitIndex].Hits++
				spell.Hot(target).Apply(sim)
			},
			// What the HoT heals in all at its target, before crits, for a rotation's conditions.
			ExpectedHealing: func(target *core.Unit) float64 {
				spell := priest.Renew[rank]
				target = priest.healTarget(target)
				hot := spell.Hot(target)
				tick := r.base + hot.BonusCoefficient*spell.HealingPower(target)
				return RenewTicks * spell.ExpectedHealing(target, tick) * hot.DamageMultiplier
			},
		})
	}
}

// RenewOn is our Renew of any rank on the target, nil when there's none.
func (priest *Priest) RenewOn(target *core.Unit) *core.Dot {
	for _, spell := range priest.Renew {
		if spell != nil && spell.Hot(target).IsActive() {
			return spell.Hot(target)
		}
	}
	return nil
}

func (priest *Priest) registerPowerWordShieldSpell() {
	priest.WeakenedSouls = priest.NewRaidAuraArray(func(unit *core.Unit) *core.Aura {
		// One debuff whichever priest's shield put it there.
		return unit.GetOrRegisterAura(core.Aura{
			Label:    "Weakened Soul",
			ActionID: core.ActionID{SpellID: WeakenedSoulSpellID},
			Duration: WeakenedSoulDuration,
		})
	})
	priest.PowerWordShieldAbsorbs = priest.NewAbsorbArray(core.Aura{
		Label:    PowerWordShieldAuraLabel,
		ActionID: core.ActionID{SpellID: PowerWordShieldRanks[len(PowerWordShieldRanks)-1].spellID},
		Duration: PowerWordShieldDuration,
	})

	// Improved Power Word: Shield: 7, 14 and 20% more absorbed (client 14748, its curve).
	absorbMultiplier := 1 + []float64{0, 0.07, 0.14, 0.20}[priest.Talents.ImprovedPowerWordShield]
	cooldown := PowerWordShieldCooldown
	if priest.Talents.SoulWarding {
		// Soul Warding: 4 sec off the cooldown (all of it) and 15% off the cost (client 402000).
		cooldown -= 4 * time.Second
	}
	var cd core.Cooldown
	if cooldown > 0 {
		cd = core.Cooldown{Timer: priest.NewTimer(), Duration: cooldown}
	}

	priest.PowerWordShield = make([]*core.Spell, len(PowerWordShieldRanks))
	priest.powerWordShieldAbsorbs = make([]func(*core.Unit) float64, len(PowerWordShieldRanks))
	for rank := 1; rank < len(PowerWordShieldRanks); rank++ {
		r := PowerWordShieldRanks[rank]
		if !priest.KnowsRank(r.spellID, r.learnLevel) {
			continue
		}
		low, _ := r.At(priest.Level)
		var spell *core.Spell
		// spell.BonusCoefficient is the rank's after Forever's downranking (sim/core/downrank.go).
		amount := func(target *core.Unit) float64 {
			return (low + spell.BonusCoefficient*spell.HealingPower(target)) * absorbMultiplier * priest.PseudoStats.ShieldDealtMultiplier
		}
		priest.powerWordShieldAbsorbs[rank] = amount
		spell = priest.RegisterSpell(core.SpellConfig{
			SpellCode:   SpellCode_PriestPowerWordShield,
			ActionID:    core.ActionID{SpellID: r.spellID},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       SpellFlagPriest | core.SpellFlagHelpful | core.SpellFlagAPL,

			RequiredLevel: int(r.learnLevel),
			Rank:          rank,

			ManaCost: core.ManaCostOptions{
				FlatCost: r.manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
				CD: cd,
			},
			// Weakened Soul blocks it, and a shield of ours still up can be replaced only by an equal
			// or higher rank (the November 2 notes: "can now correctly be replaced by a Power Word:
			// Shield of equal or greater rank"; client 70753 has no field for it).
			ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
				target = priest.healTarget(target)
				if target.IsDead() || priest.WeakenedSouls.Get(target).IsActive() {
					return false
				}
				shield := priest.PowerWordShieldAbsorbs.Get(target)
				return !shield.Aura.IsActive() || shield.Spell.Rank <= rank
			},

			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			BonusCoefficient: r.coefficient,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				target = priest.healTarget(target)
				if target.IsDead() {
					return
				}
				spell.SpellMetrics[target.UnitIndex].Hits++
				priest.PowerWordShieldAbsorbs.Get(target).Apply(sim, spell, amount(target))
				priest.WeakenedSouls.Get(target).Activate(sim)
			},
			// What the shield absorbs, for a rotation's conditions.
			ExpectedHealing: func(target *core.Unit) float64 {
				return amount(priest.healTarget(target))
			},
		})
		priest.PowerWordShield[rank] = spell
	}
}

// PowerWordShieldAbsorb is what the rank's shield absorbs on the target.
func (priest *Priest) PowerWordShieldAbsorb(rank int, target *core.Unit) float64 {
	return priest.powerWordShieldAbsorbs[rank](target)
}

func (priest *Priest) registerPrayerOfMendingSpell() {
	if !priest.Talents.PrayerOfMending {
		return
	}
	timer := priest.NewTimer()

	// The rank and heal spell the charge on a target carries, and the charges left.
	var current *core.Spell
	charges := 0
	var onTarget *core.Aura

	heals := make([]*core.Spell, len(PrayerOfMendingRanks))
	var auras core.AuraArray
	var jump func(sim *core.Simulation, from *core.Unit)

	trigger := func(sim *core.Simulation, aura *core.Aura) {
		if current == nil || aura != onTarget {
			return
		}
		unit := aura.Unit
		heal := current
		aura.Deactivate(sim)
		heal.Cast(sim, unit)
		charges--
		if charges > 0 {
			jump(sim, unit)
		} else {
			current = nil
		}
	}

	auras = priest.NewRaidAuraArray(func(unit *core.Unit) *core.Aura {
		return unit.RegisterAura(core.Aura{
			Label:    "Prayer of Mending-" + priest.Label,
			ActionID: core.ActionID{SpellID: PrayerOfMendingRanks[len(PrayerOfMendingRanks)-1].spellID},
			Tag:      "PrayerOfMending",
			Duration: PrayerOfMendingDuration,
			OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if result.Damage > 0 && spell.Unit.IsOpponent(aura.Unit) {
					trigger(sim, aura)
				}
			},
			OnPeriodicDamageTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if result.Damage > 0 && spell.Unit.IsOpponent(aura.Unit) {
					trigger(sim, aura)
				}
			},
			OnHealTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if spell.ActionID.SpellID != PrayerOfMendingHealID {
					trigger(sim, aura)
				}
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				if aura == onTarget {
					onTarget = nil
				}
			},
		})
	})

	place := func(sim *core.Simulation, unit *core.Unit) {
		if onTarget != nil {
			old := onTarget
			onTarget = nil
			old.Deactivate(sim)
		}
		onTarget = auras.Get(unit)
		onTarget.Activate(sim)
	}

	// The next player: one alive with no Prayer of Mending on it, at random.
	jump = func(sim *core.Simulation, from *core.Unit) {
		var candidates []*core.Unit
		for _, unit := range priest.Env.Raid.AllPlayerUnits {
			if unit == from || unit.IsDead() {
				continue
			}
			if _, ok := priest.Env.Raid.GetPlayerFromUnit(unit).(*core.TargetDummy); ok {
				continue
			}
			if slices.ContainsFunc(unit.GetAurasWithTag("PrayerOfMending"), func(a *core.Aura) bool { return a.IsActive() }) {
				continue
			}
			candidates = append(candidates, unit)
		}
		if len(candidates) == 0 {
			current = nil
			return
		}
		place(sim, candidates[int(sim.RandomFloat("Prayer of Mending")*float64(len(candidates)))%len(candidates)])
	}

	priest.PrayerOfMending = make([]*core.Spell, len(PrayerOfMendingRanks))
	for rank := 1; rank < len(PrayerOfMendingRanks); rank++ {
		r := PrayerOfMendingRanks[rank]
		if !priest.KnowsRank(r.spellID, r.learnLevel) {
			continue
		}
		heal := priest.RegisterSpell(core.SpellConfig{
			SpellCode:   SpellCode_PriestPrayerOfMending,
			ActionID:    core.ActionID{SpellID: PrayerOfMendingHealID, Tag: int32(rank)},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       SpellFlagPriest | core.SpellFlagHelpful | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

			// The jumps' heal has no rank of its own: it downranks as the rank cast.
			DownrankLevel: int(r.learnLevel),

			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			BonusCoefficient: r.coefficient,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				if !target.IsDead() {
					priest.directHeal(sim, spell, target, r.base)
				}
			},
		})
		heals[rank] = heal

		priest.PrayerOfMending[rank] = priest.RegisterSpell(core.SpellConfig{
			SpellCode:   SpellCode_PriestPrayerOfMending,
			ActionID:    core.ActionID{SpellID: r.spellID},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       SpellFlagPriest | core.SpellFlagHelpful | core.SpellFlagAPL,

			RequiredLevel: int(r.learnLevel),
			Rank:          rank,

			ManaCost: core.ManaCostOptions{
				FlatCost: r.manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
				CD: core.Cooldown{
					Timer:    timer,
					Duration: PrayerOfMendingCooldown,
				},
			},
			ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
				return !priest.healTarget(target).IsDead()
			},

			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				target = priest.healTarget(target)
				priest.litanyOfLight(sim, spell, r.manaCost)
				spell.SpellMetrics[target.UnitIndex].Hits++
				current = heal
				charges = PrayerOfMendingCharges
				place(sim, target)
			},
			// One heal at its target, before crits.
			ExpectedHealing: func(target *core.Unit) float64 {
				return heal.ExpectedHealing(priest.healTarget(target), r.base)
			},
		})
	}
	priest.prayerOfMendingReset = func() {
		current = nil
		charges = 0
		onTarget = nil
	}
}
