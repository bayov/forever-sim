package shaman

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

const SearingTotemRanks = 6

var SearingTotemSpellId = [SearingTotemRanks + 1]int32{0, 3599, 6363, 6364, 6365, 10437, 10438}
var SearingTotemAttackSpellId = [SearingTotemRanks + 1]int32{0, 3606, 6350, 6351, 6352, 10435, 10436}
var SearingTotemBaseDamage = [SearingTotemRanks + 1][]float64{{0}, {9, 11}, {13, 17}, {19, 25}, {26, 34}, {33, 45}, {40, 54}}

// Forever cut the Searing Totem bolt's coefficient to 1.7% on every rank (wowhead Forever
// spell pages, Classic had 5.2% and 8.3%). The damage is Classic's.
var SearingTotemSpellCoef = [SearingTotemRanks + 1]float64{0, .017, .017, .017, .017, .017, .017}
var SearingTotemManaCost = [SearingTotemRanks + 1]float64{0, 25, 45, 75, 110, 145, 170}
var SearingTotemDuration = [SearingTotemRanks + 1]int{0, 30, 35, 40, 45, 50, 55}
var SearingTotemLevel = [SearingTotemRanks + 1]int{0, 10, 20, 30, 40, 50, 60}

func (shaman *Shaman) registerSearingTotemSpell() {
	shaman.SearingTotem = make([]*core.Spell, SearingTotemRanks+1)

	for rank := 1; rank <= SearingTotemRanks; rank++ {
		// Only the ranks the level has learned, so nothing below (a totem buff aura) is
		// built for a rank the shaman cannot cast.
		if SearingTotemLevel[rank] <= int(shaman.Level) {
			config := shaman.newSearingTotemSpellConfig(rank)
			shaman.SearingTotem[rank] = shaman.RegisterSpell(config)
		}
	}

	shaman.FireTotems = append(
		shaman.FireTotems,
		core.FilterSlice(shaman.SearingTotem, func(spell *core.Spell) bool { return spell != nil })...,
	)
}

func (shaman *Shaman) newSearingTotemSpellConfig(rank int) core.SpellConfig {
	totemSpellId := SearingTotemSpellId[rank]
	baseDamageLow := SearingTotemBaseDamage[rank][0]
	baseDamageHigh := SearingTotemBaseDamage[rank][1]
	spellCoeff := SearingTotemSpellCoef[rank]
	manaCost := SearingTotemManaCost[rank]
	duration := time.Second * time.Duration(SearingTotemDuration[rank])
	level := SearingTotemLevel[rank]

	attackInterval := time.Millisecond * 2500

	attackSpell := shaman.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_ShamanSearingTotem,
		ActionID:    core.ActionID{SpellID: SearingTotemAttackSpellId[rank]},
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskEmpty,

		DamageMultiplier: shaman.callOfFlameMultiplier(),
		BonusCoefficient: spellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})

	spell := core.SpellConfig{
		SpellCode:   SpellCode_ShamanSearingTotem,
		ActionID:    core.ActionID{SpellID: totemSpellId},
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       SpellFlagTotem | core.SpellFlagAPL,

		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
			Factor:   shaman.totemicFocusFactor(),
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: totemGCD,
			},
			IgnoreHaste: true,
		},

		// The totem runs as a dot on the shaman, so it keeps going when its target dies. It
		// shoots our current target, which moves on to the next mob when one dies, as long as
		// it's in range (searingTotemTarget).
		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				Label: fmt.Sprintf("Searing Totem (Rank %d)", rank),
			},
			// The totem casts its 2.2 sec bolt again and again. On the Forever beta (2026-10-10) the next
			// cast started about 0.2 sec after the last one finished, or about 0.6 sec in a quarter of
			// the gaps. It didn't wait for the bolt to land, because the gaps were the same with the
			// totem 2 or 19 yards from the mob. The gap averaged 2.53 sec over 72 gaps, and a 40 sec
			// totem fired 16 bolts, which 2.5 sec ticks give too.
			NumberOfTicks: int32(duration / attackInterval),
			TickLength:    attackInterval,
			OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
				if target := shaman.searingTotemTarget(sim); target != nil {
					attackSpell.Cast(sim, target)
				}
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			// +1 needed because of rounding issues with totem tick time.
			shaman.placeTotem(sim, FireTotem, spell, duration+1, shaman.endFireTotem(spell))
			spell.AOEDot().Apply(sim)
		},
	}

	return spell
}

const MagmaTotemRanks = 4

var MagmaTotemSpellId = [MagmaTotemRanks + 1]int32{0, 8190, 10585, 10586, 10587}
var MagmaTotemAoeSpellId = [MagmaTotemRanks + 1]int32{0, 8187, 10579, 10580, 10581}

// Forever damage per pulse (wowhead Forever tooltips), 2 less than 1.12 at every rank.
var MagmaTotemBaseDamage = [MagmaTotemRanks + 1]float64{0, 20, 35, 52, 73}
var MagmaTotemSpellCoeff = [MagmaTotemRanks + 1]float64{0, .033, .033, .033, .033}
var MagmaTotemManaCost = [MagmaTotemRanks + 1]float64{0, 230, 360, 500, 650}
var MagmaTotemLevel = [MagmaTotemRanks + 1]int{0, 26, 36, 46, 56}

func (shaman *Shaman) registerMagmaTotemSpell() {
	shaman.MagmaTotem = make([]*core.Spell, MagmaTotemRanks+1)

	for rank := 1; rank <= MagmaTotemRanks; rank++ {
		// Only the ranks the level has learned, so nothing below (a totem buff aura) is
		// built for a rank the shaman cannot cast.
		if MagmaTotemLevel[rank] <= int(shaman.Level) {
			config := shaman.newMagmaTotemSpellConfig(rank)
			shaman.MagmaTotem[rank] = shaman.RegisterSpell(config)
		}
	}

	shaman.FireTotems = append(
		shaman.FireTotems,
		core.FilterSlice(shaman.MagmaTotem, func(spell *core.Spell) bool { return spell != nil })...,
	)
}

func (shaman *Shaman) newMagmaTotemSpellConfig(rank int) core.SpellConfig {
	spellId := MagmaTotemSpellId[rank]
	baseDamage := MagmaTotemBaseDamage[rank]
	spellCoeff := MagmaTotemSpellCoeff[rank]
	manaCost := MagmaTotemManaCost[rank]
	level := MagmaTotemLevel[rank]

	// Improved Fire Nova (the client's Trial by Fire) adds 10 sec a point since client
	// 1.60.1.70425.
	duration := time.Second*20 + shaman.improvedFireNovaMagmaDuration()
	attackInterval := time.Second * 2

	aoeSpell := shaman.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_ShamanMagmaTotem,
		ActionID:    core.ActionID{SpellID: MagmaTotemAoeSpellId[rank]},
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskEmpty,

		DamageMultiplier: shaman.callOfFlameMultiplier(),
		BonusCoefficient: spellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for _, aoeTarget := range shaman.fireTotemMobs(sim, magmaTotemRadius) {
				spell.CalcAndDealDamage(sim, aoeTarget, baseDamage, spell.OutcomeMagicHitAndCrit)
			}
		},
	})

	spell := core.SpellConfig{
		SpellCode:   SpellCode_ShamanMagmaTotem,
		ActionID:    core.ActionID{SpellID: spellId},
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       SpellFlagTotem | core.SpellFlagAPL,

		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
			Factor:   shaman.totemicFocusFactor(),
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: totemGCD,
			},
			IgnoreHaste: true,
		},

		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				Label: fmt.Sprintf("Magma Totem (Rank %d)", rank),
			},
			NumberOfTicks: int32(duration / attackInterval),
			TickLength:    attackInterval,

			OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
				aoeSpell.Cast(sim, shaman.CurrentTarget)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// +1 needed because of rounding issues with totem tick time.
			shaman.placeTotem(sim, FireTotem, spell, duration+1, shaman.endFireTotem(spell))
			spell.AOEDot().Apply(sim)
		},
	}

	return spell
}

// Fire Nova is a spell under Forever, not a totem. It goes off at once around the fire
// totem the shaman already has down, so without one it cannot be cast. It has a 10 sec
// cooldown of its own (Classic's totem had 15) and Totemic Focus no longer discounts it.
//
// The damage comes from SoD's Fire Nova damage spells (408423 to 408428), not the old
// totem's (8349, 8502, 8503, 11306 and 11307). The client's tooltip still reads the old
// totem's numbers, but the beta's combat log names 408424 on every rank 2 hit (2026-10-10).
// Since client 1.60.1.70425 every rank has a 14.3% coefficient (21.4% before, the old totem
// had 10% and 14.3%), and less base damage and growth: 33 / 55 / 130 / 179 / 239 at the learn
// level, plus 0.4 / 0.6 / 0.9 / 1.1 / 1.3 a level (50 / 102 / 183 / 284 / 403 plus 1.1 / 1.6 /
// 2.2 / 2.8 / 3.4 before). A rank grows for five levels after it's learned, and the ranges
// below are at that fifth level, with each rank's spread as before.
const FireNovaRanks = 5

var FireNovaSpellId = [FireNovaRanks + 1]int32{0, 408341, 408342, 408343, 408344, 408345}
var FireNovaBaseDamage = [FireNovaRanks + 1][]float64{{0, 0}, {32.31, 37.69}, {54.28, 61.72}, {126.26, 142.74}, {173.39, 195.61}, {232.03, 258.97}}
var FireNovaScaling = [FireNovaRanks + 1]core.RankScaling{{}, {17, 0.4}, {27, 0.6}, {37, 0.9}, {47, 1.1}, {57, 1.3}}
var FireNovaSpellCoeff = [FireNovaRanks + 1]float64{0, .143, .143, .143, .143, .143}

// Client 1.60.1.70425 cut the costs by 25 to 32% (95 / 170 / 280 / 395 / 520 before).
var FireNovaManaCost = [FireNovaRanks + 1]float64{0, 65, 120, 210, 290, 375}
var FireNovaLevel = [FireNovaRanks + 1]int{0, 12, 22, 32, 42, 52}

func (shaman *Shaman) registerFireNovaSpell() {
	shaman.FireNova = make([]*core.Spell, FireNovaRanks+1)
	shaman.FireNovaHits = make([]*core.Spell, FireNovaRanks+1)

	for rank := 1; rank <= FireNovaRanks; rank++ {
		if FireNovaLevel[rank] <= int(shaman.Level) {
			shaman.FireNovaHits[rank] = shaman.RegisterSpell(shaman.newFireNovaHitSpellConfig(rank))
			shaman.FireNova[rank] = shaman.RegisterSpell(shaman.newFireNovaSpellConfig(rank, shaman.FireNovaHits[rank]))
		}
	}
}

// The cast deals nothing itself: it casts the damage spell of its rank, which hits. So the
// cast spends Clearcasting and rolls Elemental Focus, and the hit carries the damage talents.
func (shaman *Shaman) newFireNovaSpellConfig(rank int, hits *core.Spell) core.SpellConfig {
	return core.SpellConfig{
		SpellCode:   SpellCode_ShamanFireNova,
		ActionID:    core.ActionID{SpellID: FireNovaSpellId[rank]},
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagShaman | core.SpellFlagAPL | core.SpellFlagCastFromTotem,

		RequiredLevel: FireNovaLevel[rank],
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: FireNovaManaCost[rank],
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Second*10 - shaman.improvedFireNovaCooldownReduction(),
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, _ *core.Unit) bool {
			return shaman.ActiveTotems[FireTotem] != nil && shaman.TotemExpirations[FireTotem] > sim.CurrentTime &&
				sim.Encounter.LevelingPosition(sim).Dist(shaman.totemPositions[FireTotem]) <= fireNovaCastRange
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
			hits.Cast(sim, target)
		},
	}
}

// Fire Nova's hits are a proc, so item procs and talents that listen to spell hits don't hear
// them: the damage row is a triggered spell that lacks Not a Proc. In wowsims/forever's combat
// logs Cursed Murloc Eye, a 100% proc on melee and spell hits, fired on 14 of 533 Fire Nova hits,
// all coincidences, against 80 to 100% of direct spells (3c39b8afe6). On the beta 4 Fire Nova
// crits gave no Elemental Devastation (shaman_audit.md 4.13). Judgement of Wisdom and Water
// Shield, whose rows can't proc from procs either, give no mana for them. The hits report as the
// cast's spell tagged 1.
func (shaman *Shaman) newFireNovaHitSpellConfig(rank int) core.SpellConfig {
	baseDamageLow := FireNovaScaling[rank].At(FireNovaBaseDamage[rank][0], shaman.Level)
	baseDamageHigh := FireNovaScaling[rank].At(FireNovaBaseDamage[rank][1], shaman.Level)

	return core.SpellConfig{
		ActionID:    core.ActionID{SpellID: FireNovaSpellId[rank], Tag: 1},
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamageProc,
		// A shaman spell, so Elemental Fury raises its crits.
		Flags: SpellFlagShaman | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		// Call of Flame and Improved Fire Nova multiply each other under Forever.
		//
		// Classic adds percent modifiers of the same kind, and the SoD sim adds these two
		// (15% + 20% = 35% at the top ranks). But on the beta 4 of 25 Fire Nova hits with
		// both talents were over the most that adding allows, and all 25 fit 1.15 * 1.2 =
		// 1.38 (shaman_audit.md 5.6). We use that until more beta data says otherwise.
		DamageMultiplier: shaman.callOfFlameMultiplier() * (1 + shaman.improvedFireNovaBonus()),
		ThreatMultiplier: 1,
		BonusCoefficient: FireNovaSpellCoeff[rank],

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			for _, aoeTarget := range shaman.fireTotemMobs(sim, fireNovaRadius) {
				spell.CalcAndDealDamage(sim, aoeTarget, sim.Roll(baseDamageLow, baseDamageHigh), spell.OutcomeMagicHitAndCrit)
			}
		},
	}
}

// Flametongue Totem gives every main hand hit of the party extra Fire damage, by weapon
// speed like Flametongue Weapon: the value below per 4 sec of swing time (548 / 25 on
// wowhead's Forever tooltip for rank 1, which does not grow with level). Forever made it
// last 5 min.
//
// Its buff takes the totem slot of the weapon, so it works next to Windfury, Rockbiter or
// Frostbrand Weapon and an oil or stone. Flametongue Weapon on the main hand turns it off.
// It does not work next to the shaman's own Windfury Totem or another shaman's totem buff,
// see core/totem_weapon_buffs.go.
const FlametongueTotemRanks = core.FlametongueTotemRanks

var FlametongueTotemSpellId = core.FlametongueTotemSpellId
var FlametongueTotemManaCost = [FlametongueTotemRanks + 1]float64{0, 90, 140, 200, 275}
var FlametongueTotemLevel = core.FlametongueTotemLevel

func (shaman *Shaman) registerFlametongueTotemSpell() {
	shaman.FlametongueTotem = make([]*core.Spell, FlametongueTotemRanks+1)

	for rank := 1; rank <= FlametongueTotemRanks; rank++ {
		if FlametongueTotemLevel[rank] <= int(shaman.Level) {
			shaman.FlametongueTotem[rank] = shaman.RegisterSpell(shaman.newFlametongueTotemSpellConfig(rank))
		}
	}

	shaman.FireTotems = append(
		shaman.FireTotems,
		core.FilterSlice(shaman.FlametongueTotem, func(spell *core.Spell) bool { return spell != nil })...,
	)
}

func (shaman *Shaman) newFlametongueTotemSpellConfig(rank int) core.SpellConfig {
	duration := time.Minute * 5
	buffAura := core.FlametongueTotemBuffAura(&shaman.Character, rank, "Flametongue Totem", 0)

	spell := shaman.newTotemSpellConfig(FlametongueTotemManaCost[rank], FlametongueTotemSpellId[rank])
	spell.RequiredLevel = FlametongueTotemLevel[rank]
	spell.Rank = rank
	// The totem is a one tick dot like the other fire totems, so that placing another fire
	// totem takes this one down the same way.
	spell.Dot = core.DotConfig{
		IsAOE: true,
		Aura: core.Aura{
			Label: fmt.Sprintf("Flametongue Totem (Rank %d) Totem", rank),
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				shaman.setTotemWeaponBuff(sim, buffAura, proto.WeaponImbue_FlametongueWeapon)
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				shaman.clearTotemWeaponBuff(sim, buffAura)
			},
		},
		NumberOfTicks: 1,
		TickLength:    duration,
		OnTick:        func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {},
	}
	spell.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		shaman.placeTotem(sim, FireTotem, spell, duration, shaman.endFireTotem(spell))
		spell.AOEDot().Apply(sim)
		shaman.setTotemReach(FireTotem,
			func(sim *core.Simulation) {
				shaman.setTotemWeaponBuff(sim, buffAura, proto.WeaponImbue_FlametongueWeapon)
			},
			func(sim *core.Simulation) {
				shaman.clearTotemWeaponBuff(sim, buffAura)
			})
	}
	return spell
}

// endFireTotem ends a fire totem, which the sim runs as a dot on the shaman.
func (shaman *Shaman) endFireTotem(spell *core.Spell) func(sim *core.Simulation) {
	return func(sim *core.Simulation) {
		spell.AOEDot().Cancel(sim)
	}
}
