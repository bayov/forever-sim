package main

import (
	"slices"

	"github.com/wowsims/classic/sim/core/proto"
)

// The weapon styles of a plan (LevelingPlanSettings.weapon_styles), the same as WeaponBag in
// ui/core/leveling_gear.ts.
//
// A plan with styles keeps every weapon, shield and held off-hand it finds in a bag, and wears the
// best set of its level's style from the bag. So a warrior who levels with a two-hander and takes a
// shield at 25 has the shields it came across before. The items we come across don't depend on the
// styles, so two plans with other styles and the same seed find the same items.

// What a find for a style we switch to later counts, next to one that helps now. It's small, so a
// quest reward for a later style only wins over rewards that don't help us now.
const laterStyleShare = 0.001

// The level we can buy white weapons from.
const buyWhiteLevel = 5

// sortedStyles are the ranges by level, lowest first.
func sortedStyles(ranges []*proto.LevelingWeaponStyleRange) []*proto.LevelingWeaponStyleRange {
	out := slices.Clone(ranges)
	slices.SortStableFunc(out, func(a, b *proto.LevelingWeaponStyleRange) int { return int(a.Level - b.Level) })
	return out
}

// weaponStyleAt is the style the sorted ranges have at a level. The lowest range counts below its
// level too, as with the rotations.
func weaponStyleAt(sorted []*proto.LevelingWeaponStyleRange, level int32) proto.LevelingWeaponStyle {
	if len(sorted) == 0 {
		return proto.LevelingWeaponStyle_LevelingWeaponStyleFree
	}
	style := sorted[0].Style
	for _, r := range sorted {
		if r.Level <= level {
			style = r.Style
		}
	}
	return style
}

// What we hold: the main hand and the off hand, either of them nothing.
type handSet struct {
	mainHand, offHand *gearPiece
	score             float64
	// Whether it's a set of the style we asked for.
	full bool
}

type weaponBag struct {
	in     *gearPlanInputs
	pieces []*gearPiece
}

func (b *weaponBag) add(p gearPiece) { b.pieces = append(b.pieces, &p) }

func (b *weaponBag) canDualWield(level int32) bool {
	return b.in.rules.CanDualWield && level >= b.in.rules.DualWieldLevel
}

// best is the set we wear of a style at a level, by the style's weights. Without a full set of the
// style (no two-hander found yet, or no Dual Wield yet) we wear the best set of any style.
func (b *weaponBag) best(level int32, style proto.LevelingWeaponStyle) handSet {
	return b.bestOf(b.pieces, level, style)
}

func (b *weaponBag) bestOf(pieces []*gearPiece, level int32, style proto.LevelingWeaponStyle) handSet {
	w := b.in.weightsAt(level, style)
	twoHand, mainHand, shield := handSet{score: -inf}, handSet{score: -inf}, handSet{score: -inf}
	var offHands, mainHands []*gearPiece
	better := func(a *handSet, p *gearPiece, slot proto.ItemSlot) {
		score := scorePiece(*p, slot, &w)
		if score <= a.score+pickEpsilon {
			return
		}
		if slot == proto.ItemSlot_ItemSlotMainHand {
			*a = handSet{mainHand: p, score: score}
		} else {
			*a = handSet{offHand: p, score: score}
		}
	}
	canDualWield := b.canDualWield(level)
	for _, p := range pieces {
		item := p.item
		switch {
		case item.WeaponType == proto.WeaponType_WeaponTypeShield:
			better(&shield, p, proto.ItemSlot_ItemSlotOffHand)
			offHands = append(offHands, p)
		case item.WeaponType == proto.WeaponType_WeaponTypeOffHand:
			offHands = append(offHands, p)
		case item.HandType == proto.HandType_HandTypeTwoHand:
			better(&twoHand, p, proto.ItemSlot_ItemSlotMainHand)
		default:
			if item.HandType != proto.HandType_HandTypeOffHand {
				better(&mainHand, p, proto.ItemSlot_ItemSlotMainHand)
				mainHands = append(mainHands, p)
			}
			if item.HandType != proto.HandType_HandTypeMainHand && canDualWield {
				offHands = append(offHands, p)
			}
		}
	}
	// The best main hand with an off hand from the list. One weapon can't be in both hands, so the
	// best main hand may have to go in the off hand.
	pair := func(allowed func(p *gearPiece) bool) handSet {
		best := handSet{score: -inf}
		for _, oh := range offHands {
			if !allowed(oh) {
				continue
			}
			ohScore := scorePiece(*oh, proto.ItemSlot_ItemSlotOffHand, &w)
			for _, mh := range mainHands {
				if mh == oh {
					continue
				}
				if score := scorePiece(*mh, proto.ItemSlot_ItemSlotMainHand, &w) + ohScore; score > best.score+pickEpsilon {
					best = handSet{mainHand: mh, offHand: oh, score: score}
				}
			}
		}
		return best
	}
	isWeapon := func(p *gearPiece) bool {
		return p.item.WeaponType != proto.WeaponType_WeaponTypeShield && p.item.WeaponType != proto.WeaponType_WeaponTypeOffHand
	}
	set := handSet{score: -inf}
	switch style {
	case proto.LevelingWeaponStyle_LevelingWeaponStyleTwoHand:
		set = twoHand
	case proto.LevelingWeaponStyle_LevelingWeaponStyleDualWield:
		set = pair(isWeapon)
	case proto.LevelingWeaponStyle_LevelingWeaponStyleOneHandShield:
		if mainHand.mainHand != nil && shield.offHand != nil {
			set = handSet{mainHand: mainHand.mainHand, offHand: shield.offHand, score: mainHand.score + shield.score}
		}
	}
	if set.mainHand != nil {
		set.full = true
		return set
	}
	for _, other := range []handSet{twoHand, mainHand, pair(func(*gearPiece) bool { return true })} {
		if other.mainHand != nil && other.score > set.score+pickEpsilon {
			set = other
		}
	}
	if set.mainHand == nil {
		set.score = 0
	}
	return set
}

// gain is what a find adds to the set of the level's style, or a small share of what it adds to a
// style we switch to later. The first full set of a style counts from nothing, so a plan with a
// shield takes its first shield even when a two-hander would score more.
func (b *weaponBag) gain(p gearPiece, level int32) float64 {
	with := append(slices.Clone(b.pieces), &p)
	gainFor := func(style proto.LevelingWeaponStyle) float64 {
		old, found := b.bestOf(b.pieces, level, style), b.bestOf(with, level, style)
		if found.full && !old.full {
			return found.score
		}
		return found.score - old.score
	}
	if now := gainFor(weaponStyleAt(b.in.weaponStyles, level)); now > pickEpsilon {
		return now
	}
	later := 0.0
	for _, r := range b.in.weaponStyles {
		if r.Level > level {
			later = max(later, gainFor(r.Style))
		}
	}
	return later * laterStyleShare
}

// buy gets the white pieces the level's style misses, like a first two-hander.
//
// We'd buy a white weapon from a vendor rather than play another style until we come across one.
// Below buyWhiteLevel we don't have the money, and we wear what we started with.
func (b *weaponBag) buy(pool *itemPool, level int32) {
	style := weaponStyleAt(b.in.weaponStyles, level)
	if level < buyWhiteLevel || b.best(level, style).full {
		return
	}
	add := func(item *proto.UIItem) {
		if item != nil {
			b.add(gearPiece{item: item, source: "vendor", since: level})
		}
	}
	switch style {
	case proto.LevelingWeaponStyle_LevelingWeaponStyleTwoHand:
		add(pool.plainWeapon(true, level))
	case proto.LevelingWeaponStyle_LevelingWeaponStyleDualWield:
		for i := 0; i < 2 && b.canDualWield(level) && !b.best(level, style).full; i++ {
			add(pool.plainWeapon(false, level))
		}
	case proto.LevelingWeaponStyle_LevelingWeaponStyleOneHandShield:
		if pool.canWield(proto.WeaponType_WeaponTypeShield, false) {
			add(pool.plainShield(level))
		}
		if !b.best(level, style).full {
			add(pool.plainWeapon(false, level))
		}
	}
}

// wear puts the level's set in the gear, and says which slots changed.
func (b *weaponBag) wear(gear map[proto.ItemSlot]gearPiece, level int32) []proto.ItemSlot {
	set := b.best(level, weaponStyleAt(b.in.weaponStyles, level))
	var changed []proto.ItemSlot
	for _, hand := range []struct {
		slot  proto.ItemSlot
		piece *gearPiece
	}{{proto.ItemSlot_ItemSlotMainHand, set.mainHand}, {proto.ItemSlot_ItemSlotOffHand, set.offHand}} {
		worn, ok := gear[hand.slot]
		if hand.piece == nil {
			delete(gear, hand.slot)
			continue
		}
		if ok && worn.item.Id == hand.piece.item.Id && worn.suffix.GetId() == hand.piece.suffix.GetId() {
			continue
		}
		p := *hand.piece
		p.since = level
		gear[hand.slot] = p
		changed = append(changed, hand.slot)
	}
	return changed
}
