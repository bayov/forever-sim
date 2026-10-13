package mage

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const FrostfireBoltRanks = 3

var FrostfireBoltSpellId = [FrostfireBoltRanks + 1]int32{0, 401502, 1237312, 1237313}

// Frostfire Bolt is a Forever trainer spell (SoD's rune), learned at 40, 50 and 60. It is Fire and
// Frost at once, so the talents and buffs of both schools count, and it chills (40% slower for 9
// sec). Damage at each rank's cap (or at 60), its growth per level, the 81.4% coefficient, the 3
// sec cast and the DoT (3 ticks, 3 sec apart, no coefficient in the client) are the Forever
// client's (1.60.1.70338). Rank 3 is 270 to 314 plus 57 over 9 sec for 370 mana.
var FrostfireBoltBaseDamage = [FrostfireBoltRanks + 1][]float64{{0}, {101.93, 118.87}, {180.6, 210.6}, {269.6, 314.4}}
var FrostfireBoltScaling = [FrostfireBoltRanks + 1]core.RankScaling{{}, {48, 1.3}, {58, 1.7}, {}}
var FrostfireBoltDotTick = [FrostfireBoltRanks + 1]float64{0, 9, 13, 19}
var FrostfireBoltManaCost = [FrostfireBoltRanks + 1]float64{0, 205, 285, 370}
var FrostfireBoltLevel = [FrostfireBoltRanks + 1]int{0, 40, 50, 60}

func (mage *Mage) registerFrostfireBoltSpell() {
	mage.FrostfireBolt = make([]*core.Spell, FrostfireBoltRanks+1)
	for rank := 1; rank <= FrostfireBoltRanks; rank++ {
		if FrostfireBoltLevel[rank] <= int(mage.Level) {
			mage.FrostfireBolt[rank] = mage.RegisterSpell(mage.newFrostfireBoltSpellConfig(rank))
		}
	}
}

func (mage *Mage) newFrostfireBoltSpellConfig(rank int) core.SpellConfig {
	actionID := core.ActionID{SpellID: FrostfireBoltSpellId[rank]}
	baseDamageLow := FrostfireBoltScaling[rank].At(FrostfireBoltBaseDamage[rank][0], mage.Level)
	baseDamageHigh := FrostfireBoltScaling[rank].At(FrostfireBoltBaseDamage[rank][1], mage.Level)
	dotTick := FrostfireBoltDotTick[rank]

	return core.SpellConfig{
		ActionID:     actionID,
		SpellCode:    SpellCode_MageFrostfireBolt,
		SpellSchool:  core.SpellSchoolFire | core.SpellSchoolFrost,
		DefenseType:  core.DefenseTypeMagic,
		ProcMask:     core.ProcMaskSpellDamage,
		Flags:        SpellFlagMage | SpellFlagChillSpell | core.SpellFlagAPL,
		MissileSpeed: 24,

		RequiredLevel: FrostfireBoltLevel[rank],
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: FrostfireBoltManaCost[rank],
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
				// Improved Fireball covers Frostfire Bolt under Forever.
				CastTime: time.Millisecond*3000 - time.Millisecond*100*time.Duration(mage.Talents.ImprovedFireball),
			},
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:    fmt.Sprintf("Frostfire Bolt (Rank %d)", rank),
				ActionID: actionID.WithTag(1),
			},
			NumberOfTicks: 3,
			TickLength:    time.Second * 3,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, dotTick, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 0.814,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			// Manaflare Regalia 5: 10% more crit while Combustion is up.
			bonusCrit := 0.0
			if mage.frostfireCombustionCrit > 0 && mage.CombustionAura != nil && mage.CombustionAura.IsActive() {
				bonusCrit = mage.frostfireCombustionCrit
			}
			spell.BonusCritRating += bonusCrit
			result := mage.calcMageDamage(sim, target, baseDamage, spell)
			spell.BonusCritRating -= bonusCrit

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
				if result.Landed() {
					spell.Dot(target).Apply(sim)
				}
			})
		},
	}
}
