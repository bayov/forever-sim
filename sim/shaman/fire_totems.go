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
			FlatCost:   manaCost,
			Multiplier: shaman.totemManaMultiplier(),
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: totemGCD,
			},
			IgnoreHaste: true,
		},

		Dot: core.DotConfig{
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
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				attackSpell.Cast(sim, target)
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			if shaman.ActiveTotems[FireTotem] != nil {
				shaman.ActiveTotems[FireTotem].Dot(sim.GetTargetUnit(0)).Cancel(sim)
			}
			spell.Dot(sim.GetTargetUnit(0)).Apply(sim)
			// +1 needed because of rounding issues with totem tick time.
			shaman.TotemExpirations[FireTotem] = sim.CurrentTime + duration + 1
			shaman.ActiveTotems[FireTotem] = spell
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

	duration := time.Second * 20
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
			for _, aoeTarget := range sim.Encounter.TargetUnits {
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
			FlatCost:   manaCost,
			Multiplier: shaman.totemManaMultiplier(),
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: totemGCD,
			},
			IgnoreHaste: true,
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: fmt.Sprintf("Magma Totem (Rank %d)", rank),
			},
			NumberOfTicks: int32(duration / attackInterval),
			TickLength:    attackInterval,

			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				aoeSpell.Cast(sim, target)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if shaman.ActiveTotems[FireTotem] != nil {
				shaman.ActiveTotems[FireTotem].Dot(sim.GetTargetUnit(0)).Cancel(sim)
			}
			spell.Dot(sim.GetTargetUnit(0)).Apply(sim)
			// +1 needed because of rounding issues with totem tick time.
			shaman.TotemExpirations[FireTotem] = sim.CurrentTime + duration + 1
			shaman.ActiveTotems[FireTotem] = spell
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
// Every rank has a 21.4% coefficient, where the old totem had 10% and 14.3%, and a few
// points less base damage. A rank grows for five levels after it's learned, and the ranges
// below are at that fifth level (client 70338).
const FireNovaRanks = 5

var FireNovaSpellId = [FireNovaRanks + 1]int32{0, 408341, 408342, 408343, 408344, 408345}
var FireNovaBaseDamage = [FireNovaRanks + 1][]float64{{0, 0}, {51.23, 59.77}, {102.94, 117.06}, {182.12, 205.88}, {280.06, 315.94}, {396.95, 443.05}}
var FireNovaScaling = [FireNovaRanks + 1]core.RankScaling{{}, {17, 1.1}, {27, 1.6}, {37, 2.2}, {47, 2.8}, {57, 3.4}}
var FireNovaSpellCoeff = [FireNovaRanks + 1]float64{0, .214, .214, .214, .214, .214}
var FireNovaManaCost = [FireNovaRanks + 1]float64{0, 95, 170, 280, 395, 520}
var FireNovaLevel = [FireNovaRanks + 1]int{0, 12, 22, 32, 42, 52}

func (shaman *Shaman) registerFireNovaSpell() {
	shaman.FireNova = make([]*core.Spell, FireNovaRanks+1)

	for rank := 1; rank <= FireNovaRanks; rank++ {
		if FireNovaLevel[rank] <= int(shaman.Level) {
			shaman.FireNova[rank] = shaman.RegisterSpell(shaman.newFireNovaSpellConfig(rank))
		}
	}
}

func (shaman *Shaman) newFireNovaSpellConfig(rank int) core.SpellConfig {
	baseDamageLow := FireNovaScaling[rank].At(FireNovaBaseDamage[rank][0], shaman.Level)
	baseDamageHigh := FireNovaScaling[rank].At(FireNovaBaseDamage[rank][1], shaman.Level)

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
			return shaman.ActiveTotems[FireTotem] != nil && shaman.TotemExpirations[FireTotem] > sim.CurrentTime
		},

		// Call of Flame and Improved Fire Nova are both percent damage modifiers on this
		// spell. In Classic those add together (15% + 20% = 35%) instead of multiplying, and
		// the SoD sim does the same.
		DamageMultiplier: shaman.callOfFlameMultiplier() + shaman.improvedFireNovaBonus(),
		ThreatMultiplier: 1,
		BonusCoefficient: FireNovaSpellCoeff[rank],

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			for _, aoeTarget := range sim.Encounter.TargetUnits {
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
		if shaman.ActiveTotems[FireTotem] != nil {
			shaman.ActiveTotems[FireTotem].Dot(sim.GetTargetUnit(0)).Cancel(sim)
		}
		spell.Dot(sim.GetTargetUnit(0)).Apply(sim)
		shaman.TotemExpirations[FireTotem] = sim.CurrentTime + duration
		shaman.ActiveTotems[FireTotem] = spell
	}
	return spell
}
