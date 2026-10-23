package mage

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

const ArcaneBlastRanks = 5

var ArcaneBlastSpellId = [ArcaneBlastRanks + 1]int32{0, 400574, 1239696, 1239697, 1239699, 1239700}

// Arcane Blast is a talent under Forever (Arcane, row 3), with trainer ranks behind it. Damage at
// each rank's cap (or at 60), its growth per level up to there and the 42.9% coefficient are the
// Forever client's (1.60.1.70425): base points plus 15.1% variance, so rank 5 is 218 to 254 at
// 60. Every rank costs 15% of base mana and casts in 1.5 sec (2.5 sec and 71.4% before 70425).
var ArcaneBlastBaseDamage = [ArcaneBlastRanks + 1][]float64{{0}, {33.28, 38.72}, {78.03, 90.77}, {100.77, 117.23}, {164.93, 191.87}, {218.18, 253.82}}
var ArcaneBlastScaling = [ArcaneBlastRanks + 1]core.RankScaling{{}, {28, 0.5}, {38, 0.8}, {48, 1.0}, {58, 1.3}, {}}
var ArcaneBlastLevel = [ArcaneBlastRanks + 1]int{0, 20, 30, 40, 50, 60}

// Each Arcane Blast puts a stack of 400573 on the mage, up to 4: 15% of Intellect as spell
// damage (all schools) and 175% more mana on Arcane Blast per stack. Since 70425 the buff lasts
// 15 sec from its first stack, more stacks don't refresh it, and other spells no longer end it
// (the client gives it no charges and no proc flags). The 10% damage bonus on other spells is gone.
const ArcaneBlastIntellectPerStack = 0.15
const ArcaneBlastMaxStacks = 4

func (mage *Mage) registerArcaneBlastSpell() {
	if !mage.Talents.ArcaneBlast {
		return
	}

	mage.ArcaneBlast = make([]*core.Spell, ArcaneBlastRanks+1)

	spellPowerDeps := make([]*stats.StatDependency, ArcaneBlastMaxStacks+1)
	for stacks := 1; stacks <= ArcaneBlastMaxStacks; stacks++ {
		spellPowerDeps[stacks] = mage.NewDynamicStatDependency(stats.Intellect, stats.SpellPower, ArcaneBlastIntellectPerStack*float64(stacks))
	}

	mage.ArcaneBlastAura = mage.RegisterAura(core.Aura{
		Label:     "Arcane Blast",
		ActionID:  core.ActionID{SpellID: 400573},
		Duration:  time.Second * 15,
		MaxStacks: ArcaneBlastMaxStacks,
		// The mage's own casts give the stacks, so the rotation knows them at once.
		NoReaction: true,
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks int32, newStacks int32) {
			if oldStacks > 0 {
				mage.DisableDynamicStatDep(sim, spellPowerDeps[oldStacks])
			}
			if newStacks > 0 {
				mage.EnableDynamicStatDep(sim, spellPowerDeps[newStacks])
			}
			// One cost row of 175% a stack, which multiplies Arcane Power's under Forever.
			for _, spell := range mage.ArcaneBlast {
				if spell != nil {
					mage.changeCostPercent(spell, 175*oldStacks, false)
					mage.changeCostPercent(spell, 175*newStacks, true)
				}
			}
		},
	})

	for rank := 1; rank <= ArcaneBlastRanks; rank++ {
		if ArcaneBlastLevel[rank] <= int(mage.Level) {
			mage.ArcaneBlast[rank] = mage.RegisterSpell(mage.newArcaneBlastSpellConfig(rank))
		}
	}
}

func (mage *Mage) newArcaneBlastSpellConfig(rank int) core.SpellConfig {
	baseDamageLow := ArcaneBlastScaling[rank].At(ArcaneBlastBaseDamage[rank][0], mage.Level)
	baseDamageHigh := ArcaneBlastScaling[rank].At(ArcaneBlastBaseDamage[rank][1], mage.Level)

	return core.SpellConfig{
		ActionID:    core.ActionID{SpellID: ArcaneBlastSpellId[rank]},
		SpellCode:   SpellCode_MageArcaneBlast,
		SpellSchool: core.SpellSchoolArcane,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagMage | core.SpellFlagAPL,

		RequiredLevel: ArcaneBlastLevel[rank],
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.15,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 1500,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 0.429,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			spell.DealDamage(sim, spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit))
			// The stack comes after the hit, so this cast doesn't use its own spell damage.
			// More stacks don't move the 15 sec on from the first one.
			if !mage.ArcaneBlastAura.IsActive() {
				mage.ArcaneBlastAura.Activate(sim)
			}
			mage.ArcaneBlastAura.AddStack(sim)
		},
	}
}
