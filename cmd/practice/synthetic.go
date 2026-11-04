//go:build practice

package main

import (
	"fmt"
	"html"
	"strings"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/tools/database"
)

// The made up Synthetic items (tools/database/forever_synthetic_items*.go, ids from 990001) are
// unknown to Wowhead. Their icon and a short tooltip come from their own definition.

var syntheticItems = func() map[int]*proto.UIItem {
	out := map[int]*proto.UIItem{}
	for _, it := range database.ForeverSyntheticItems {
		out[int(it.Id)] = it
	}
	return out
}()

var itemTypeNames = map[proto.ItemType]string{
	proto.ItemType_ItemTypeHead: "Head", proto.ItemType_ItemTypeNeck: "Neck", proto.ItemType_ItemTypeShoulder: "Shoulder",
	proto.ItemType_ItemTypeBack: "Back", proto.ItemType_ItemTypeChest: "Chest", proto.ItemType_ItemTypeWrist: "Wrist",
	proto.ItemType_ItemTypeHands: "Hands", proto.ItemType_ItemTypeWaist: "Waist", proto.ItemType_ItemTypeLegs: "Legs",
	proto.ItemType_ItemTypeFeet: "Feet", proto.ItemType_ItemTypeFinger: "Finger", proto.ItemType_ItemTypeTrinket: "Trinket",
	proto.ItemType_ItemTypeWeapon: "Weapon", proto.ItemType_ItemTypeRanged: "Ranged",
}

// syntheticInfo gives a Synthetic item's icon, name and tooltip, and false for any other item.
func syntheticInfo(kind string, id int) (iconInfo, bool) {
	it, ok := syntheticItems[id]
	if kind != "item" || !ok {
		return iconInfo{}, false
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<table><tr><td><b class="q%d">%s</b><br>`, it.Quality, html.EscapeString(it.Name))
	slot := itemTypeNames[it.Type]
	if it.ArmorType != proto.ArmorType_ArmorTypeUnknown {
		slot += " · " + strings.TrimPrefix(it.ArmorType.String(), "ArmorType")
	}
	if slot != "" {
		b.WriteString(slot + "<br>")
	}
	st := stats.FromFloatArray(it.Stats)
	if st[stats.Armor] > 0 {
		fmt.Fprintf(&b, "%.0f Armor<br>", st[stats.Armor])
	}
	for _, s := range []struct {
		stat stats.Stat
		name string
	}{{stats.Strength, "Strength"}, {stats.Agility, "Agility"}, {stats.Stamina, "Stamina"}, {stats.Intellect, "Intellect"}, {stats.Spirit, "Spirit"}} {
		if v := st[s.stat]; v > 0 {
			fmt.Fprintf(&b, "+%.0f %s<br>", v, s.name)
		}
	}
	if it.RequiredLevel > 0 {
		fmt.Fprintf(&b, "Requires Level %d<br>", it.RequiredLevel)
	}
	equip := func(format string, v float64) {
		if v > 0 {
			fmt.Fprintf(&b, `<span class="q2">Equip: `+format+`</span><br>`, v)
		}
	}
	equip("Increases damage and healing done by magical spells and effects by up to %.0f.", st[stats.SpellPower])
	equip("+%.0f Attack Power.", st[stats.AttackPower])
	equip("Improves your chance to get a critical strike by %.0f%%.", it.CritRating)
	equip("Improves your chance to hit by %.0f%%.", it.HitRating)
	equip("Restores %.0f mana per 5 sec.", st[stats.MP5])
	b.WriteString(`<span class="q0">Synthetic item for the sim</span></td></tr></table>`)
	return iconInfo{Icon: it.Icon, Name: it.Name, Tooltip: b.String(), Fetched: true}, true
}
