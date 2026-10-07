package main

import (
	"fmt"
	"os"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
)

// Forever's own items have ids from 247000 up (Novice's Practice Wand 247789, Northshire
// Hammer 248005). Below that is Classic, and from 200000 to here Season of Discovery,
// whose items the client carries but Forever does not use. From 990000 up are our made up
// Phase 1 items (tools/database/forever_synthetic_items.go).
const (
	foreverFirstItemID   = 247000
	sodFirstItemID       = 200000
	syntheticFirstItemID = 990000
)

func isSodItem(id int32) bool       { return id >= sodFirstItemID && id < foreverFirstItemID }
func isSyntheticItem(id int32) bool { return id >= syntheticFirstItemID }

// scaleSpellPower multiplies the spell power on Forever's new items above level 30.
//
// The devs said the level 40 to 60 gear in the client is months old and not tuned like
// the beta's level 1 to 30 gear yet. We measured what a point of spell power costs in an
// item's stat budget (next to 1 for a point of Strength) on 2026-10-05, against an Era
// item of the same item level, because Era itself prices spell power lower the higher the
// item level (0.95 at item level 30, 0.70 at 58, 0.54 at 75).
//
// On the tuned level 1 to 30 gear, Forever's new items pay about 0.8 of what Era pays at
// the same item level (0.85 at item level 17, 0.76 at 30). On the untuned high level gear
// the answer depends on how we fit it. Per item level band, the item level 55 to 62 items
// pay 1.14 of Era's price and the 63 to 70 items 0.76. One fit over item level 55 to 70
// says 1.2. So a tuned item at Phase 1 item levels would carry 1.0 to 1.5 times the spell
// power it has today, and this takes that factor as a knob.
//
// Only spell power changes. Every other stat, the weapon damage and the procs stay as they
// are. Classic items keep theirs: at level 1 to 30 Forever left most of them alone, and
// where it traded Spirit for spell power it did so one for one.
func scaleSpellPower(scale float64, dbPath string) error {
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return err
	}
	db := &proto.UIDatabase{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, db); err != nil {
		return err
	}

	spellPowers := []stats.Stat{stats.SpellPower, stats.ArcanePower, stats.FirePower, stats.FrostPower,
		stats.HolyPower, stats.NaturePower, stats.ShadowPower, stats.HealingPower}
	changed := 0
	for _, uiItem := range db.Items {
		// The synthetic items are priced at the tuned rate already.
		if uiItem.Id < foreverFirstItemID || isSyntheticItem(uiItem.Id) {
			continue
		}
		if uiItem.RequiredLevel <= 30 && !(uiItem.RequiredLevel == 0 && uiItem.Ilvl > 35) {
			continue
		}
		item, ok := core.ItemsByID[uiItem.Id]
		if !ok {
			continue
		}
		any := false
		for _, s := range spellPowers {
			if item.Stats[s] != 0 {
				item.Stats[s] *= scale
				any = true
			}
		}
		if any {
			core.ItemsByID[uiItem.Id] = item
			changed++
		}
	}
	fmt.Printf("spell power times %.2f on %d of Forever's new items above level 30\n", scale, changed)
	return nil
}
