package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const WindfuryWeaponRanks = 4

var WindfuryWeaponSpellId = [WindfuryWeaponRanks + 1]int32{0, 8232, 8235, 10486, 16362}
var WindfuryWeaponEnchantId = [WindfuryWeaponRanks + 1]int32{0, 283, 284, 525, 1669}
var WindfuryWeaponLevel = [WindfuryWeaponRanks + 1]int32{0, 30, 40, 50, 60}

// The extra attack power at each rank's cap and its growth per level, like Rockbiter.
// Forever keeps Classic's values: 333 for rank 4 at level 60.
//
// wowhead's Forever tooltip shows 433 for rank 4, but that is the value at level 68. The
// rank grows 12.5 a level from 60 to 68 (its <!--ppl60:68:333:1250--> marker), and
// wowhead's Forever pages evaluate it at the top of that range while the Classic pages use
// level 60. ForeverChanges' spellbook says 333.
var WindfuryWeaponBonusAP = [WindfuryWeaponRanks + 1]float64{0, 103.6, 221.4, 315.4, 333}
var WindfuryWeaponScaling = [WindfuryWeaponRanks + 1]core.RankScaling{{}, {38, 7.2}, {48, 12.8}, {58, 8.3}, {60, 12.5}}

func (shaman *Shaman) windfuryRank() int {
	return core.HighestRankAt(shaman.Level, WindfuryWeaponLevel[:])
}

func (shaman *Shaman) newWindfuryImbueSpell(isMH bool) *core.Spell {
	rank := shaman.windfuryRank()

	ewMultiplier := []float64{1, 1.13, 1.27, 1.4}[shaman.Talents.ElementalWeapons]
	bonusAP := WindfuryWeaponScaling[rank].At(WindfuryWeaponBonusAP[rank], shaman.Level)

	actionID := core.ActionID{SpellID: WindfuryWeaponSpellId[rank]}.WithTag(core.TernaryInt32(isMH, 1, 2))
	procMask := core.ProcMaskMeleeMHSpecial
	damageMultiplier := 1.0
	weaponDamageFunc := shaman.MHWeaponDamage
	if !isMH {
		procMask = core.ProcMaskMeleeOHSpecial
		damageMultiplier = shaman.AutoAttacks.OHConfig().DamageMultiplier
		weaponDamageFunc = shaman.OHWeaponDamage
	}

	spellConfig := core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    procMask,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		DamageMultiplier: damageMultiplier,
		ThreatMultiplier: 1,
		// Flat "+N damage" effects like Zandalarian Hero Medallion add to it, as to a white hit.
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Elemental Weapons applies twice to the extra attack power (1.4 * 1.4 at 3
			// points). The SoD sim does this on purpose (its commit bc9aa4d3, "update Windfury
			// Weapon to double dip from Elemental Weapons"), and Forever's mechanics follow
			// SoD's.
			// TODO: Beta can confirm it, by comparing Windfury hits with 0 and 3 points.
			mAP := spell.MeleeAttackPower(target) + bonusAP*ewMultiplier*ewMultiplier
			baseDamage := weaponDamageFunc(sim, mAP)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		},
	}

	return shaman.RegisterSpell(spellConfig)
}

func (shaman *Shaman) RegisterWindfuryImbue(procMask core.ProcMask) {
	// Windfury Weapon is learned at 30.
	rank := shaman.windfuryRank()
	if procMask == core.ProcMaskUnknown || rank == 0 {
		return
	}
	enchantId := WindfuryWeaponEnchantId[rank]

	icdDuration := time.Millisecond * 1500

	if procMask.Matches(core.ProcMaskMeleeMH) {
		shaman.MainHand().TempEnchant = enchantId
	}
	if procMask.Matches(core.ProcMaskMeleeOH) {
		shaman.OffHand().TempEnchant = enchantId
	}

	var proc = 0.2
	if procMask == core.ProcMaskMelee {
		proc = 0.36
	}

	icd := core.Cooldown{
		Timer:    shaman.NewTimer(),
		Duration: icdDuration,
	}

	shaman.WindfuryWeaponMH = shaman.newWindfuryImbueSpell(true)

	aura := shaman.RegisterAura(core.Aura{
		Label:    "Windfury Imbue",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(procMask) || spell.Flags.Matches(core.SpellFlagSuppressEquipProcs) {
				return
			}

			if !icd.IsReady(sim) {
				return
			}

			if sim.RandomFloat("Windfury Imbue") < proc {
				icd.Use(sim)

				// TODO: Vanilla uses two extra attacks but SoD replaced this with yellow hits
				// This needs to be refactored
				shaman.WindfuryWeaponMH.Cast(sim, result.Target)
				shaman.WindfuryWeaponMH.Cast(sim, result.Target)
			}
		},
	})

	shaman.RegisterOnItemSwapWithImbue(enchantId, &procMask, aura)
}

func (shaman *Shaman) ApplyWindfuryImbue(procMask core.ProcMask) {
	if procMask.Matches(core.ProcMaskMeleeMH) && shaman.HasMHWeapon() {
		shaman.ApplyWindfuryImbueToItem(shaman.MainHand())
	}

	if procMask.Matches(core.ProcMaskMeleeOH) && shaman.HasOHWeapon() {
		shaman.ApplyWindfuryImbueToItem(shaman.OffHand())
	}
}

func (shaman *Shaman) ApplyWindfuryImbueToItem(item *core.Item) {
	if item == nil {
		return
	}

	rank := shaman.windfuryRank()
	if rank == 0 {
		return
	}
	item.TempEnchant = WindfuryWeaponEnchantId[rank]
}
