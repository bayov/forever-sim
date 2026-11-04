package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Shank is a Forever ability from Blizzard's November 2 development notes, which server hotfix push
// 113483 (launch realms, build 70707, 2026-11-03; orchestration/round_70753/hotfix_113483.md) carries:
// 1329283 at level 24 and 1329549 at 44. 5 Energy, 10 sec cooldown, 1 sec GCD, an off-hand strike for
// 50% weapon damage (effect 0) that gives 15 Energy, 25 at rank 2 (effect 1), and no combo point.
// "Shank cannot Miss": its attributes are the client's Always Hit (Attributes[3] 0x40000, so no miss,
// dodge or parry; it can still crit) and Requires Off-Hand Weapon (0x1000000), and it suppresses the
// caster's procs (Attributes[3] 0x10000, Attributes[4] 0x800000 Suppress Weapon Procs): no poison,
// weapon enchant or Hack and Slash proc from Shank. Venom and Boot Knife, which its own tooltip
// names, still go off. Assumed (notes.md, Rogue): the 50% is of the off-hand's normalized weapon damage
// after the usual off-hand halving, like the other off-hand strikes.
var ShankActionID = core.ActionID{SpellID: 1329549}

func (rogue *Rogue) registerShank() {
	// Forever only: Classic has no Shank (standards 37).
	if !rogue.Env.IsForever() || rogue.Level < 24 || !rogue.HasOHWeapon() {
		return
	}

	// Flash of Steel (1329284): 3 more Energy and 0.4 sec off the cooldown a rank. Boot Knife
	// (1329286): a 20% chance a rank to add 2 combo points (1329285). Both from hotfix push 113483.
	energyGain := rankAt(rogue.Level, map[int32]float64{24: 15, 44: 25}) + 3*float64(rogue.Talents.FlashOfSteel)
	cooldown := time.Second*10 - 400*time.Millisecond*time.Duration(rogue.Talents.FlashOfSteel)
	bootKnifeChance := 0.2 * float64(rogue.Talents.BootKnife)
	actionID := core.ActionID{SpellID: rankSpellID(rogue.Level, map[int32]int32{24: 1329283, 44: 1329549})}
	energyMetrics := rogue.NewEnergyMetrics(actionID)
	cpMetrics := rogue.NewComboPointMetrics(actionID.WithTag(1))

	rogue.Shank = rogue.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_RogueShank,
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeOHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL | core.SpellFlagSuppressWeaponProcs | core.SpellFlagSuppressEquipProcs,

		EnergyCost: core.EnergyCostOptions{
			Cost: 5,
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

		DamageMultiplier: 0.5 * rogue.AutoAttacks.OHConfig().DamageMultiplier,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			baseDamage := rogue.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialCritOnly)

			if result.Landed() {
				rogue.AddEnergy(sim, energyGain, energyMetrics)
				if bootKnifeChance > 0 && sim.Proc(bootKnifeChance, "Boot Knife") {
					rogue.AddComboPoints(sim, 2, target, cpMetrics)
				}
				if rogue.VenomAura != nil {
					rogue.VenomAura.Activate(sim)
				}
			}
		},
	})
}
