package mage

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const ArcaneBlastRanks = 5

var ArcaneBlastSpellId = [ArcaneBlastRanks + 1]int32{0, 400574, 1239696, 1239697, 1239699, 1239700}

// Arcane Blast is a talent under Forever (Arcane, row 3), with trainer ranks behind it. Damage at
// each rank's cap (or at 60), its growth per level up to there and the 71.4% coefficient are the
// Forever client's (1.60.1.70338). Every rank costs 15% of base mana and casts in 2.5 sec.
var ArcaneBlastBaseDamage = [ArcaneBlastRanks + 1][]float64{{0}, {56.58, 65.82}, {130.54, 151.86}, {168.07, 195.53}, {268.84, 312.76}, {364.25, 423.75}}
var ArcaneBlastScaling = [ArcaneBlastRanks + 1]core.RankScaling{{}, {28, 0.9}, {38, 1.4}, {48, 1.6}, {58, 2.1}, {}}
var ArcaneBlastLevel = [ArcaneBlastRanks + 1]int{0, 20, 30, 40, 50, 60}

// Each Arcane Blast puts a stack of 400573 on the mage, up to 4, for 8 sec: 10% more damage on
// the mage's other spells and 175% more mana on Arcane Blast per stack. Any other damage spell
// ends it. The tooltip says "all your other spells", and so does the sim, Arcane Missiles too.
// The client's 400573 leaves Arcane Missiles out of the damage part, and a second buff, 400586
// (15% a stack on Arcane Missiles, from SoD), is still in the client. Which of the two Forever
// gives Arcane Missiles is open (notes.md).
const ArcaneBlastStackBonus = 0.10

func (mage *Mage) registerArcaneBlastSpell() {
	if !mage.Talents.ArcaneBlast {
		return
	}

	mage.ArcaneBlast = make([]*core.Spell, ArcaneBlastRanks+1)

	// The stack bonus goes on every other mage damage spell. Arcane Missiles takes it at the
	// start of the channel (see arcane_missiles.go), so its missiles are left out here.
	var boostedSpells []*core.Spell
	mage.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Flags.Matches(SpellFlagMage) && spell.ProcMask.Matches(core.ProcMaskSpellDamage) &&
			spell.SpellCode != SpellCode_MageArcaneBlast && spell.SpellCode != SpellCode_MageArcaneMissilesTick {
			boostedSpells = append(boostedSpells, spell)
		}
	})

	mage.ArcaneBlastAura = mage.RegisterAura(core.Aura{
		Label:     "Arcane Blast",
		ActionID:  core.ActionID{SpellID: 400573},
		Duration:  time.Second * 8,
		MaxStacks: 4,
		// The mage's own casts give the stacks, so the rotation knows them at once.
		NoReaction: true,
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks int32, newStacks int32) {
			for _, spell := range boostedSpells {
				spell.DamageMultiplierAdditive += ArcaneBlastStackBonus * float64(newStacks-oldStacks)
			}
			for _, spell := range mage.ArcaneBlast {
				if spell != nil {
					spell.Cost.Multiplier += 175 * (newStacks - oldStacks)
				}
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.SpellCode == SpellCode_MageArcaneBlast || !spell.Flags.Matches(SpellFlagMage) || !spell.ProcMask.Matches(core.ProcMaskSpellDamage) {
				return
			}
			aura.Deactivate(sim)
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
				CastTime: time.Millisecond * 2500,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 0.714,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			spell.DealDamage(sim, mage.calcMageDamage(sim, target, baseDamage, spell))
			mage.ArcaneBlastAura.Activate(sim)
			mage.ArcaneBlastAura.AddStack(sim)
		},
	}
}

// arcaneBlastBonus is the damage bonus the Arcane Blast stacks give the mage's other spells now.
func (mage *Mage) arcaneBlastBonus() float64 {
	if mage.ArcaneBlastAura == nil || !mage.ArcaneBlastAura.IsActive() {
		return 0
	}
	return ArcaneBlastStackBonus * float64(mage.ArcaneBlastAura.GetStacks())
}
