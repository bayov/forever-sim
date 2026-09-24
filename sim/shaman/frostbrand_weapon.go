package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const FrostbrandWeaponRanks = 5

var FrostbrandWeaponSpellId = [FrostbrandWeaponRanks + 1]int32{0, 8033, 8038, 10456, 16355, 16356}
var FrostbrandWeaponEnchantId = [FrostbrandWeaponRanks + 1]int32{0, 2, 12, 524, 1667, 1668}
var FrostbrandWeaponLevel = [FrostbrandWeaponRanks + 1]int32{0, 20, 28, 38, 48, 58}

// Damage per proc at each rank's cap, and its growth per level. The damage is the Forever
// tooltip (wowhead). The growth per level is 1.12's, scaled down with the damage.
var FrostbrandWeaponBaseDamage = [FrostbrandWeaponRanks + 1]float64{0, 45, 72, 117, 159, 203}
var FrostbrandWeaponScaling = [FrostbrandWeaponRanks + 1]core.RankScaling{{}, {26, 2}, {36, 2.8}, {46, 4.7}, {56, 3.8}, {60, 6.1}}

func (shaman *Shaman) frostbrandRank() int {
	return core.HighestRankAt(shaman.Level, FrostbrandWeaponLevel[:])
}

func (shaman *Shaman) FrostbrandDebuffAura(target *core.Unit) *core.Aura {
	rank := shaman.frostbrandRank()
	spellId := FrostbrandWeaponSpellId[rank]

	return target.GetOrRegisterAura(core.Aura{
		Label:    "Frostbrand Attack-" + shaman.Label,
		ActionID: core.ActionID{SpellID: spellId},
		Duration: time.Second * 8,
	})
}

func (shaman *Shaman) newFrostbrandImbueSpell() *core.Spell {
	rank := shaman.frostbrandRank()
	spellId := FrostbrandWeaponSpellId[rank]
	baseDamage := FrostbrandWeaponScaling[rank].At(FrostbrandWeaponBaseDamage[rank], shaman.Level)

	return shaman.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellId},
		SpellSchool: core.SpellSchoolFrost,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamageProc,

		DamageMultiplier: []float64{1, 1.05, 1.1, 1.15}[shaman.Talents.ElementalWeapons],
		ThreatMultiplier: 1,
		BonusCoefficient: 0.1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})
}

func (shaman *Shaman) RegisterFrostbrandImbue(procMask core.ProcMask) {
	// Frostbrand Weapon is learned at 20.
	rank := shaman.frostbrandRank()
	if procMask == core.ProcMaskUnknown || rank == 0 {
		return
	}
	enchantId := FrostbrandWeaponEnchantId[rank]

	if procMask.Matches(core.ProcMaskMeleeMH) {
		shaman.MainHand().TempEnchant = enchantId
	}
	if procMask.Matches(core.ProcMaskMeleeOH) {
		shaman.OffHand().TempEnchant = enchantId
	}

	ppmm := shaman.AutoAttacks.NewPPMManager(9.0, procMask)

	mhSpell := shaman.newFrostbrandImbueSpell()
	ohSpell := shaman.newFrostbrandImbueSpell()

	fbDebuffAuras := shaman.NewEnemyAuraArray(shaman.FrostbrandDebuffAura)

	aura := shaman.RegisterAura(core.Aura{
		Label:    "Frostbrand Imbue",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() {
				return
			}

			if ppmm.Proc(sim, spell.ProcMask, "Frostbrand Weapon") {
				if spell.IsMH() {
					mhSpell.Cast(sim, result.Target)
				} else {
					ohSpell.Cast(sim, result.Target)
				}
				fbDebuffAuras.Get(result.Target).Activate(sim)
			}
		},
	})

	shaman.ItemSwap.RegisterOnSwapItemForEffectWithPPMManager(3784, 9.0, &ppmm, aura)
}

func (shaman *Shaman) ApplyFrostbrandImbue(procMask core.ProcMask) {
	if procMask.Matches(core.ProcMaskMeleeMH) && shaman.HasMHWeapon() {
		shaman.ApplyFrostbrandImbueToItem(shaman.MainHand())
	}

	if procMask.Matches(core.ProcMaskMeleeOH) && shaman.HasOHWeapon() {
		shaman.ApplyFrostbrandImbueToItem(shaman.OffHand())
	}
}

func (shaman *Shaman) ApplyFrostbrandImbueToItem(item *core.Item) {
	if item == nil {
		return
	}

	rank := shaman.frostbrandRank()
	if rank == 0 {
		return
	}
	item.TempEnchant = FrostbrandWeaponEnchantId[rank]
}
