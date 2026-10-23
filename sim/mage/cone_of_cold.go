package mage

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const ConeOfColdRanks = 5

var ConeOfColdSpellId = [ConeOfColdRanks + 1]int32{0, 120, 8492, 10159, 10160, 10161}

// Damage at each rank's cap (or at 60) and its growth per level up to there, from the Forever
// client (wowsims/forever's store at bce49ac6a2, from 70425 with the 2026-10-20 hotfix): rank 5 is
// 340 at 58 with 1.5 a level and an 8.6% spread, 328 to 358 at 60, for 555 mana. Every rank has a
// 12.9% coefficient, a 10 sec cooldown (category 50, which the ranks share) and a 10 yard radius.
var ConeOfColdBaseDamage = [ConeOfColdRanks + 1][]float64{{0}, {96, 106}, {142, 156}, {200, 220}, {260, 287}, {328, 358}}
var ConeOfColdScaling = [ConeOfColdRanks + 1]core.RankScaling{{}, {31, 0.8}, {39, 1}, {47, 1.2}, {55, 1.3}, {60, 1.5}}
var ConeOfColdManaCost = [ConeOfColdRanks + 1]float64{0, 210, 290, 380, 465, 555}
var ConeOfColdLevel = [ConeOfColdRanks + 1]int{0, 26, 34, 42, 50, 58}

const ConeOfColdSpellCoeff = 0.129
const ConeOfColdRadius = 10

// Winter's Chill since the 2026-10-20 hotfix to 70450 (wago.tools hotfix push 112694, SpellEffect
// 688253, 1364854 and 1364855; wowsims/forever's store has the same since cf82c75efb): each stack
// on the target is 5% more damage taken from the caster's Cone of Cold and 7% of spell power more
// on its coefficient, up to 5 stacks.
const WintersChillDamagePerStack = 0.05
const WintersChillCoeffPerStack = 0.07

// Cone of Cold under Forever: an instant cone at the targets in front, which a mage can only cast
// with the target within its radius (Arctic Reach adds 10% a point). The slow does nothing to a
// boss. It is a Chill (the Permafrost mask, 0x100000), so each hit rolls Fingers of Frost, and
// binary like every spell with a slow. Its mask (0x200) takes Improved Cone of Cold, Ice Shards,
// Elemental Precision, Arcane Power, Arcane Instability, Clearcasting and Ice Barrier's and Frigid
// Power's Frost damage, and Frost Channeling's cost (0x80000); Combustion's mask leaves it out, and
// Cold Snap resets it. Every rank is a trainer rank, so there is no book rank to leave out.
func (mage *Mage) registerConeOfColdSpell() {
	if !mage.Env.IsForever() {
		return
	}

	mage.ConeOfCold = make([]*core.Spell, ConeOfColdRanks+1)
	cdTimer := mage.NewTimer()
	for rank := 1; rank <= ConeOfColdRanks; rank++ {
		if ConeOfColdLevel[rank] <= int(mage.Level) {
			mage.ConeOfCold[rank] = mage.RegisterSpell(mage.newConeOfColdSpellConfig(rank, cdTimer))
		}
	}
}

func (mage *Mage) newConeOfColdSpellConfig(rank int, cdTimer *core.Timer) core.SpellConfig {
	baseDamageLow := ConeOfColdScaling[rank].At(ConeOfColdBaseDamage[rank][0], mage.Level)
	baseDamageHigh := ConeOfColdScaling[rank].At(ConeOfColdBaseDamage[rank][1], mage.Level)
	radius := ConeOfColdRadius * (1 + 0.1*float64(mage.Talents.ArcticReach))

	// Improved Cone of Cold: 12 / 23 / 35% (curve 11190).
	improvedBonus := []float64{0, 0.12, 0.23, 0.35}[mage.Talents.ImprovedConeOfCold]

	return core.SpellConfig{
		ActionID:    core.ActionID{SpellID: ConeOfColdSpellId[rank]},
		SpellCode:   SpellCode_MageConeOfCold,
		SpellSchool: core.SpellSchoolFrost,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagMage | SpellFlagChillSpell | core.SpellFlagBinary | core.SpellFlagAPL,
		MaxRange:    radius,

		RequiredLevel: ConeOfColdLevel[rank],
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: ConeOfColdManaCost[rank],
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: time.Second * 10,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return mage.DistanceFromTarget <= radius
		},

		DamageMultiplier: 1 + improvedBonus,
		ThreatMultiplier: 1,
		BonusCoefficient: ConeOfColdSpellCoeff,

		// Winter's Chill raises the damage and coefficient against each target by that target's
		// own stacks, which the hit then spends.
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
				if sim.Encounter.LevelingDistance(sim, aoeTarget) > radius {
					continue
				}
				damageMultiplier, bonusCoefficient := mage.wintersChillBonus(aoeTarget)
				spell.DamageMultiplier *= damageMultiplier
				spell.BonusCoefficient += bonusCoefficient
				baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
				result := spell.CalcAndDealDamage(sim, aoeTarget, baseDamage, spell.OutcomeMagicHitAndCrit)
				spell.DamageMultiplier /= damageMultiplier
				spell.BonusCoefficient -= bonusCoefficient

				if result.Landed() && mage.wintersChillStacks(aoeTarget) > 0 {
					mage.WintersChillAuras.Get(aoeTarget).Deactivate(sim)
				}
			}
		},
	}
}

// Winter's Chill (5 points): Frostbolt and Frost Blast hits that land (the sim has no Frost Blast),
// and an Ice Lance cast at a Frozen target, have a 20% chance a point to put a stack on the target,
// up to 5, for 15 sec (wowsims/forever talents_frost.go). Each mage has their own stacks, which only
// their own Cone of Cold reads, and spends. Under Forever only: in Classic it was a crit debuff for
// everyone's Frost spells, which the sim never modelled.
func (mage *Mage) applyWintersChill() {
	if mage.Talents.WintersChill == 0 || !mage.Env.IsForever() {
		return
	}

	procChance := 0.2 * float64(mage.Talents.WintersChill)

	mage.WintersChillAuras = mage.NewEnemyAuraArray(func(unit *core.Unit) *core.Aura {
		return unit.RegisterAura(core.Aura{
			Label:     "Winter's Chill - " + mage.Label,
			ActionID:  core.ActionID{SpellID: 12579},
			Duration:  time.Second * 15,
			MaxStacks: 5,
		})
	})

	core.MakePermanent(mage.RegisterAura(core.Aura{
		Label: "Winter's Chill Talent",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.SpellCode == SpellCode_MageFrostbolt && result.Landed() {
				mage.procWintersChill(sim, result.Target, procChance)
			}
		},
	}))
	mage.wintersChillChance = procChance
}

// procWintersChill rolls Winter's Chill for a stack on target.
func (mage *Mage) procWintersChill(sim *core.Simulation, target *core.Unit, chance float64) {
	if sim.Proc(chance, "Winter's Chill") {
		aura := mage.WintersChillAuras.Get(target)
		aura.Activate(sim)
		aura.AddStack(sim)
	}
}

// procWintersChillOnIceLance rolls Winter's Chill for an Ice Lance on its cast result, where its
// damage read the target as Frozen: Fingers of Frost may have no charge left by the time it lands.
func (mage *Mage) procWintersChillOnIceLance(sim *core.Simulation, result *core.SpellResult, frozen bool) {
	if mage.WintersChillAuras != nil && frozen && result.Landed() {
		mage.procWintersChill(sim, result.Target, mage.wintersChillChance)
	}
}

// wintersChillBonus is what the mage's Winter's Chill stacks on target add to their Cone of Cold:
// a damage factor and spell power coefficient.
func (mage *Mage) wintersChillBonus(target *core.Unit) (float64, float64) {
	stacks := float64(mage.wintersChillStacks(target))
	return 1 + WintersChillDamagePerStack*stacks, WintersChillCoeffPerStack * stacks
}

// wintersChillStacks is the mage's own Winter's Chill stacks on target.
func (mage *Mage) wintersChillStacks(target *core.Unit) int32 {
	if mage.WintersChillAuras == nil {
		return 0
	}
	return mage.WintersChillAuras.Get(target).GetStacks()
}
