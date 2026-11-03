package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Shank is a Forever ability from Blizzard's November 2 development notes: "Learned at level 24. 5
// Energy cost. 10 second cooldown. Melee range. Instantly strike with your off-hand weapon, dealing
// 50% weapon damage and generating 15 Energy. Shank cannot Miss. Rank 2 learned at level 44 and
// generates 25 Energy." Client 1.60.1.70753 has no Shank yet, so the notes are the only source and
// the spell ID is ours (ShankActionID). Assumed (notes.md, Rogue): the 50% is of the off-hand's
// normalized weapon damage after the usual off-hand halving, like the other off-hand strikes; it
// can be dodged, parried and can crit, only the miss is gone; it adds no combo point.
var ShankActionID = core.ActionID{SpellID: 1399248}

func (rogue *Rogue) registerShank() {
	if rogue.Level < 24 || !rogue.HasOHWeapon() {
		return
	}

	// Flash of Steel: 3 more Energy and 0.4 sec off the cooldown a rank. Boot Knife: a 20% chance a
	// rank to add 2 combo points (both November 2 development notes, no client spell).
	energyGain := rankAt(rogue.Level, map[int32]float64{24: 15, 44: 25}) + 3*float64(rogue.Talents.FlashOfSteel)
	cooldown := time.Second*10 - 400*time.Millisecond*time.Duration(rogue.Talents.FlashOfSteel)
	bootKnifeChance := 0.2 * float64(rogue.Talents.BootKnife)
	energyMetrics := rogue.NewEnergyMetrics(ShankActionID)
	cpMetrics := rogue.NewComboPointMetrics(ShankActionID.WithTag(1))

	rogue.Shank = rogue.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_RogueShank,
		ActionID:    ShankActionID,
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeOHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		EnergyCost: core.EnergyCostOptions{
			Cost:   5,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: cooldown,
			},
		},

		// "Shank cannot Miss": enough hit to take the miss off the table.
		BonusHitRating: 100 * core.MeleeHitRatingPerHitChance,

		DamageMultiplier: 0.5 * rogue.AutoAttacks.OHConfig().DamageMultiplier,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			baseDamage := rogue.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if result.Landed() {
				rogue.AddEnergy(sim, energyGain, energyMetrics)
				if bootKnifeChance > 0 && sim.Proc(bootKnifeChance, "Boot Knife") {
					rogue.AddComboPoints(sim, 2, target, cpMetrics)
				}
				if rogue.VenomAura != nil {
					rogue.VenomAura.Activate(sim)
				}
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}
