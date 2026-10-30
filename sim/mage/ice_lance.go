package mage

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const IceLanceRanks = 6

var IceLanceSpellId = [IceLanceRanks + 1]int32{0, 1312002, 400640, 1240044, 1240045, 1240046, 1240047}

// Ice Lance is a talent under Forever (Frost, row 3), with trainer ranks behind it. Damage at each
// rank's cap (or at 60), its growth per level and the mana are the Forever client's (1.60.1.70338, unchanged in 70425).
// Rank 6 is 136 to 161 for 160 mana, instant.
var IceLanceBaseDamage = [IceLanceRanks + 1][]float64{{0}, {27.59, 32.61}, {34.47, 40.73}, {44, 52}, {76.27, 90.13}, {94.97, 112.23}, {136.22, 160.98}}
var IceLanceScaling = [IceLanceRanks + 1]core.RankScaling{{}, {26, 0.35}, {32, 0.4}, {40, 0.5}, {48, 0.7}, {56, 0.7}, {60, 0.9}}
var IceLanceManaCost = [IceLanceRanks + 1]float64{0, 45, 55, 70, 105, 120, 160}
var IceLanceLevel = [IceLanceRanks + 1]int{0, 20, 28, 34, 42, 48, 56}

// The client has no spell power coefficient on Ice Lance, on any of the six ranks, where every
// other new Forever damage spell carries its own (Frostfire Bolt 0.814, Arcane Blast 0.714).
// SoD's client had 0 too (1.15.9 spell 400640 on wago.tools), and SoD's server gave it 42.9%.
// Frozen targets take 200% more (the client's dummy effect and the tooltip, 300% before
// 1.60.1.70425), so 3 times the damage. The other two dummies (20 and 5) look like SoD's Glaciate (20% a stack, 5 stacks),
// which no Forever tooltip mentions, so they are left out.
//
// Report assumption mage_ice_lance_coef, user ruling 2026-10-15: 0.429, SoD's server value,
// since SoD's client showed the same 0 and the server added the coefficient.
const IceLanceSpellCoeff = 0.429
const IceLanceFrozenMultiplier = 3

func (mage *Mage) registerIceLanceSpell() {
	if !mage.Talents.IceLance {
		return
	}

	cdTimer := mage.NewTimer()
	mage.IceLance = make([]*core.Spell, IceLanceRanks+1)
	for rank := 1; rank <= IceLanceRanks; rank++ {
		if IceLanceLevel[rank] <= int(mage.Level) {
			mage.IceLance[rank] = mage.RegisterSpell(mage.newIceLanceSpellConfig(rank, cdTimer))
		}
	}
}

// 1.60.1.70707 gives Ice Lance an 8 sec cooldown (category 19), the 10-27 to 10-29 hotfixes put
// into the client, but only on rank 1 (SpellCooldowns 101119, 1312002): ranks 2 to 6 have no
// cooldown row. Report assumption mage_ice_lance_cd, user ruling 2026-10-30: every rank shares
// it, as the tooltip and foreverchanges.pro have it.
const IceLanceCooldown = time.Second * 8

func (mage *Mage) newIceLanceSpellConfig(rank int, cdTimer *core.Timer) core.SpellConfig {
	baseDamageLow := IceLanceScaling[rank].At(IceLanceBaseDamage[rank][0], mage.Level)
	baseDamageHigh := IceLanceScaling[rank].At(IceLanceBaseDamage[rank][1], mage.Level)

	// Under Forever it is not binary: the 1240047 row has no slow or other non-damage effect, and
	// wowsims/forever's Ice Lance rolls partial resists.
	flags := SpellFlagMage | core.SpellFlagAPL
	if !mage.Env.IsForever() {
		flags |= core.SpellFlagBinary
	}

	return core.SpellConfig{
		ActionID:     core.ActionID{SpellID: IceLanceSpellId[rank]},
		SpellCode:    SpellCode_MageIceLance,
		SpellSchool:  core.SpellSchoolFrost,
		DefenseType:  core.DefenseTypeMagic,
		ProcMask:     core.ProcMaskSpellDamage,
		Flags:        flags,
		MissileSpeed: 38,

		RequiredLevel: IceLanceLevel[rank],
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: IceLanceManaCost[rank],
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: IceLanceCooldown,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: IceLanceSpellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)

			frozen := mage.isTargetFrozen(target)
			if frozen {
				spell.DamageMultiplier *= IceLanceFrozenMultiplier
			}
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			if frozen {
				spell.DamageMultiplier /= IceLanceFrozenMultiplier
			}
			mage.procWintersChillOnIceLance(sim, result, frozen)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	}
}
