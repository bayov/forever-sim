//go:build practice

package practice

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Sheet is your character sheet as the sim has it when the session starts, with the buffs and
// consumables of the setup up: the numbers the room's character frame shows.
type Sheet struct {
	Level int32 `json:"level"`
	// Stats are by proto.Stat index. Base holds the race and class part of each, so the frame can
	// show the gear and buff part in green like the game.
	Stats []float64 `json:"stats"`
	Base  []float64 `json:"base"`
	// Pseudo are by proto.PseudoStat index: the weapon skill bonuses and the speed multipliers.
	Pseudo  []float64     `json:"pseudo"`
	Weapons []SheetWeapon `json:"weapons"`
	Items   []SheetItem   `json:"items"`
	// DamageMultiplier is the physical damage multiplier the game folds into the damage line.
	DamageMultiplier float64 `json:"damageMultiplier"`
}

// SheetWeapon is a weapon's damage line: base damage plus attack power, and the hasted speed.
type SheetWeapon struct {
	Slot  string  `json:"slot"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Speed float64 `json:"speed"`
	Type  string  `json:"type,omitempty"`
}

// SheetItem is an equipped item, by proto.ItemSlot index.
type SheetItem struct {
	Slot    int32  `json:"slot"`
	ID      int32  `json:"id"`
	Quality int32  `json:"quality"`
	Name    string `json:"name"`
	// Enchant is the enchant's effect id, 0 when there's none.
	Enchant int32 `json:"enchant,omitempty"`
}

// Sheet reads the character sheet off the session's character.
func (s *Session) Sheet() Sheet {
	char := s.run.char
	st := char.GetStats()
	out := Sheet{
		Level:            char.Level,
		Stats:            st.ToFloatArray(),
		Base:             char.GetBaseStats().ToFloatArray(),
		Pseudo:           char.GetPseudoStatsProto(),
		DamageMultiplier: char.PseudoStats.DamageDealtMultiplier * char.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical],
	}
	weapon := func(slot string, w *core.Weapon, ap float64, haste float64, item *core.Item) {
		if w == nil || w.SwingSpeed <= 0 || item == nil || item.ID == 0 {
			return
		}
		bonus := w.SwingSpeed * ap / w.AttackPowerPerDPS
		sw := SheetWeapon{Slot: slot, Min: w.BaseDamageMin + bonus, Max: w.BaseDamageMax + bonus, Speed: w.SwingSpeed / haste}
		if item.WeaponType != proto.WeaponType_WeaponTypeUnknown {
			sw.Type = item.WeaponType.String()
		} else if item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeUnknown {
			sw.Type = item.RangedWeaponType.String()
		}
		out.Weapons = append(out.Weapons, sw)
	}
	aa := char.AutoAttacks
	weapon("mainHand", aa.MH(), st[stats.AttackPower], char.SwingSpeed(), char.MainHand())
	if char.HasOHWeapon() {
		weapon("offHand", aa.OH(), st[stats.AttackPower], char.SwingSpeed(), char.OffHand())
	}
	if char.Ranged() != nil && char.Ranged().RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeUnknown {
		weapon("ranged", aa.Ranged(), st[stats.RangedAttackPower], char.RangedSwingSpeed(), char.Ranged())
	}
	for i := range char.Equipment {
		item := &char.Equipment[i]
		if item.ID == 0 {
			continue
		}
		si := SheetItem{Slot: int32(i), ID: item.ID, Quality: int32(item.Quality), Name: item.Name}
		si.Enchant = item.Enchant.EffectID
		out.Items = append(out.Items, si)
	}
	return out
}
