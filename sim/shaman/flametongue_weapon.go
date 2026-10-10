package shaman

import (
	"github.com/wowsims/classic/sim/core"
)

const FlametongueWeaponRanks = 6

var FlametongueWeaponSpellId = [FlametongueWeaponRanks + 1]int32{0, 8024, 8027, 8030, 16339, 16341, 16342}
var FlametongueWeaponEnchantId = [FlametongueWeaponRanks + 1]int32{0, 5, 4, 3, 523, 1665, 1666}
var FlametongueWeaponLevel = [FlametongueWeaponRanks + 1]int32{0, 10, 18, 26, 36, 46, 56}

// Damage per 4 seconds of weapon speed at each rank's cap, and its growth per level. The
// Forever tooltip gives it as N / 25 (440, 653, 1052, 1728, 2372 and 3122), the same as
// 1.12.
//
// We keep the exact N / 25, so rank 2 is 26.12 and not 26.1. The Forever client (70338) has
// the same points (shaman_audit.md 7.1).
//
// Rank 6 keeps growing up to level 64 and wowhead shows it there (3122 / 25). At level 60
// it is 2810 / 25 = 112.4, the same as Classic and ForeverChanges' spellbook.
var FlametongueWeaponMaxDamage = [FlametongueWeaponRanks + 1]float64{0, 440.0 / 25, 653.0 / 25, 1052.0 / 25, 1728.0 / 25, 2372.0 / 25, 2810.0 / 25}
var FlametongueWeaponScaling = [FlametongueWeaponRanks + 1]core.RankScaling{{}, {16, .76}, {24, 1.16}, {34, 1.68}, {44, 2.92}, {54, 2.48}, {60, 3.12}}

func (shaman *Shaman) flametongueRank() int {
	return core.HighestRankAt(shaman.Level, FlametongueWeaponLevel[:])
}

func (shaman *Shaman) newFlametongueImbueSpell(weapon *core.Item) *core.Spell {
	rank := shaman.flametongueRank()
	spellID := FlametongueWeaponSpellId[rank]
	maxDamage := FlametongueWeaponScaling[rank].At(FlametongueWeaponMaxDamage[rank], shaman.Level)

	baseDamage := maxDamage / 4
	spellCoeff := .1

	return shaman.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamageProc,
		// A shaman spell, so Elemental Fury raises its crits. Forever's Elemental Fury lists
		// the Flametongue Weapon procs among the spells it affects.
		Flags: SpellFlagShaman | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		DamageMultiplier: []float64{1, 1.05, 1.1, 1.15}[shaman.Talents.ElementalWeapons],
		ThreatMultiplier: 1,
		BonusCoefficient: spellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if weapon.SwingSpeed != 0 {
				damage := (baseDamage * weapon.SwingSpeed)
				spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHitAndCrit)
			}
		},
	})
}

func (shaman *Shaman) ApplyFlametongueImbueToItem(item *core.Item) {
	if item == nil {
		return
	}

	rank := shaman.flametongueRank()
	if rank == 0 {
		return
	}
	item.TempEnchant = FlametongueWeaponEnchantId[rank]
}

func (shaman *Shaman) ApplyFlametongueImbue(procMask core.ProcMask) {
	if procMask.Matches(core.ProcMaskMeleeMH) && shaman.HasMHWeapon() {
		shaman.ApplyFlametongueImbueToItem(shaman.MainHand())
	}

	if procMask.Matches(core.ProcMaskMeleeOH) && shaman.HasOHWeapon() {
		shaman.ApplyFlametongueImbueToItem(shaman.OffHand())
	}
}

func (shaman *Shaman) RegisterFlametongueImbue(procMask core.ProcMask) {
	rank := shaman.flametongueRank()
	if procMask == core.ProcMaskUnknown && !shaman.ItemSwap.IsEnabled() || rank == 0 {
		return
	}
	enchantId := FlametongueWeaponEnchantId[rank]

	mhSpell := shaman.newFlametongueImbueSpell(shaman.MainHand())
	ohSpell := shaman.newFlametongueImbueSpell(shaman.OffHand())

	aura := shaman.RegisterAura(core.Aura{
		Label:    "Flametongue Imbue",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(procMask) {
				return
			}

			if spell.IsMH() {
				mhSpell.Cast(sim, result.Target)
			} else {
				ohSpell.Cast(sim, result.Target)
			}
		},
	})

	shaman.RegisterOnItemSwapWithImbue(enchantId, &procMask, aura)
}
