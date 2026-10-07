package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

const RockbiterWeaponRanks = 7

var RockbiterWeaponEnchantId = [RockbiterWeaponRanks + 1]int32{0, 29, 6, 1, 503, 1663, 683, 1664}
var RockbiterWeaponBonusTPS = [RockbiterWeaponRanks + 1]float64{0, 6, 10, 16, 27, 41, 55, 72}
var RockbiterWeaponLevel = [RockbiterWeaponRanks + 1]int32{0, 1, 8, 16, 24, 34, 44, 54}

// The attack power each rank gives once it has grown to its cap, and how it grows. In
// 1.12 the imbue is an attack power aura (spell 16313 for rank 7) that gains a few points
// per level from the rank's learn level up to the cap, so a level 20 shaman with rank 3
// has 108 of the 118 it will have at 22.
// Forever keeps Classic's values. Rank 7 keeps growing up to level 62 (686 there), so
// at 60 it is 653, the same as Classic and ForeverChanges' spellbook. wowhead's Forever
// tooltip shows the level 62 value.
var RockbiterWeaponBonusAP = [RockbiterWeaponRanks + 1]float64{0, 49.5, 79, 118, 193.8, 355, 521.8, 653}
var RockbiterWeaponScaling = [RockbiterWeaponRanks + 1]core.RankScaling{{}, {6, 4.1}, {14, 3.5}, {22, 5}, {32, 8.1}, {42, 18}, {52, 16.1}, {60, 16.5}}

func (shaman *Shaman) rockbiterRank() int {
	return core.HighestRankAt(shaman.Level, RockbiterWeaponLevel[:])
}

func (shaman *Shaman) RegisterRockbiterImbue(procMask core.ProcMask) {
	if procMask == core.ProcMaskUnknown {
		return
	}

	rank := shaman.rockbiterRank()
	enchantId := RockbiterWeaponEnchantId[rank]
	bonusThreat := RockbiterWeaponBonusTPS[rank]

	duration := time.Minute * 5

	hasMHImbue := procMask.Matches(core.ProcMaskMeleeMH)
	hasOHImbue := procMask.Matches(core.ProcMaskMeleeOH)

	if hasMHImbue {
		shaman.MainHand().TempEnchant = enchantId
		shaman.AutoAttacks.MHConfig().FlatThreatBonus += bonusThreat * shaman.AutoAttacks.MH().SwingSpeed
	}
	if hasOHImbue {
		shaman.OffHand().TempEnchant = enchantId
		shaman.AutoAttacks.MHConfig().FlatThreatBonus += bonusThreat * shaman.AutoAttacks.OH().SwingSpeed
	}

	shaman.OnSpellRegistered(func(spell *core.Spell) {
		if spell.ProcMask.Matches(procMask) {
			spell.FlatThreatBonus += bonusThreat
		}
	})

	aura := shaman.RegisterAura(core.Aura{
		Label:    "Rockbiter Imbue",
		Duration: duration,
	})

	shaman.RegisterOnItemSwapWithImbue(enchantId, &procMask, aura)
}

func (shaman *Shaman) ApplyRockbiterImbue(procMask core.ProcMask) {
	if procMask.Matches(core.ProcMaskMeleeMH) && shaman.HasMHWeapon() {
		shaman.ApplyRockbiterImbueToItem(shaman.MainHand())
	}

	if procMask.Matches(core.ProcMaskMeleeOH) && shaman.HasOHWeapon() {
		shaman.ApplyRockbiterImbueToItem(shaman.OffHand())
	}
}

func (shaman *Shaman) ApplyRockbiterImbueToItem(item *core.Item) {
	if item == nil {
		return
	}

	rank := shaman.rockbiterRank()
	enchantId := RockbiterWeaponEnchantId[rank]

	bonusAP := RockbiterWeaponScaling[rank].At(RockbiterWeaponBonusAP[rank], shaman.Level) * []float64{1, 1.07, 1.13, 1.2}[shaman.Talents.ElementalWeapons]

	newStats := stats.Stats{stats.AttackPower: bonusAP}

	item.Stats = item.Stats.Add(newStats)
	item.TempEnchant = enchantId
}
