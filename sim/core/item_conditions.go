package core

import (
	"slices"

	"github.com/wowsims/classic/sim/core/proto"
	goproto "google.golang.org/protobuf/proto"
)

// ItemCondition is a fight condition that makes an item's effect bigger: the target's creature
// type, its creature family, or the outdoor terrain of the fight.
//
// An item effect whose value depends on one declares it with RegisterItemConditions, next to
// the effect. The BiS lists (tools/rotopt, docs/bis_lists.md) read them to rank each such item a
// second time with its condition met. So a Seal of the Dawn, which only gives attack power
// against Undead, shows where it would go on the trinket list in Scholomance.
//
// A condition sets one field of the encounter (Apply). The zero value of a field means the
// condition doesn't touch it.
type ItemCondition struct {
	// Label is what the player reads next to the item, like "vs Undead" or "in snow".
	Label string

	MobType   proto.MobType
	MobFamily proto.MobFamily
	Terrain   proto.OutdoorTerrain
}

// The conditions the items share. An item that counts two families or terrains alike (Blackhand's
// Breadth doubles on Orcs and Ogres) uses one of them, since the effect is the same on both.
var (
	ConditionVsOrcsAndOgres      = ItemCondition{Label: "vs Orcs and Ogres", MobFamily: proto.MobFamily_MobFamilyOrc}
	ConditionVsTrolls            = ItemCondition{Label: "vs Trolls", MobFamily: proto.MobFamily_MobFamilyTroll}
	ConditionVsDinosaurs         = ItemCondition{Label: "vs Dinosaurs", MobFamily: proto.MobFamily_MobFamilyDinosaur}
	ConditionInForestOrGrassland = ItemCondition{Label: "in forest or grassland", Terrain: proto.OutdoorTerrain_OutdoorTerrainForest}
	ConditionInSnow              = ItemCondition{Label: "in snow", Terrain: proto.OutdoorTerrain_OutdoorTerrainSnow}
	ConditionVsPlants            = ItemCondition{Label: "vs Plants", MobFamily: proto.MobFamily_MobFamilyPlant}
	ConditionVsDwarves           = ItemCondition{Label: "vs Dwarves", MobFamily: proto.MobFamily_MobFamilyDwarf}
	ConditionInMarshOrSwamp      = ItemCondition{Label: "in marsh or swamp", Terrain: proto.OutdoorTerrain_OutdoorTerrainSwamp}
	ConditionInMountains         = ItemCondition{Label: "in mountains", Terrain: proto.OutdoorTerrain_OutdoorTerrainMountains}
)

var mobTypeLabels = map[proto.MobType]string{
	proto.MobType_MobTypeBeast:      "vs Beasts",
	proto.MobType_MobTypeDemon:      "vs Demons",
	proto.MobType_MobTypeDragonkin:  "vs Dragonkin",
	proto.MobType_MobTypeElemental:  "vs Elementals",
	proto.MobType_MobTypeGiant:      "vs Giants",
	proto.MobType_MobTypeHumanoid:   "vs Humanoids",
	proto.MobType_MobTypeMechanical: "vs Mechanicals",
	proto.MobType_MobTypeUndead:     "vs Undead",
}

// MobTypeCondition is the condition of fighting a creature type, like "vs Undead".
func MobTypeCondition(mobType proto.MobType) ItemCondition {
	return ItemCondition{Label: mobTypeLabels[mobType], MobType: mobType}
}

var itemConditions = map[int32][]ItemCondition{}

// RegisterItemConditions declares that the item's effect grows under each of the conditions. Call
// it from an init function, next to the item's effect. A condition the item already has is
// ignored.
func RegisterItemConditions(itemID int32, conditions ...ItemCondition) {
	for _, c := range conditions {
		if !slices.Contains(itemConditions[itemID], c) {
			itemConditions[itemID] = append(itemConditions[itemID], c)
		}
	}
}

// ItemConditions is the conditions the item's effect grows under, or nil.
func ItemConditions(itemID int32) []ItemCondition {
	return itemConditions[itemID]
}

// Apply returns a copy of the encounter with the condition met. A creature condition goes on
// every target, so a fight with several targets has them all of that type or family.
func (c ItemCondition) Apply(encounter *proto.Encounter) *proto.Encounter {
	out := goproto.Clone(encounter).(*proto.Encounter)
	for _, target := range out.Targets {
		if c.MobType != proto.MobType_MobTypeUnknown {
			target.MobType = c.MobType
		}
		if c.MobFamily != proto.MobFamily_MobFamilyUnknown {
			target.MobFamily = c.MobFamily
		}
	}
	if c.Terrain != proto.OutdoorTerrain_OutdoorTerrainNone {
		out.OutdoorTerrain = c.Terrain
	}
	return out
}

// MetBy says whether the encounter already meets the condition, on every target. A ranking on
// such an encounter already counts the item's bigger effect.
func (c ItemCondition) MetBy(encounter *proto.Encounter) bool {
	if c.Terrain != proto.OutdoorTerrain_OutdoorTerrainNone && encounter.GetOutdoorTerrain() != c.Terrain {
		return false
	}
	if c.MobType == proto.MobType_MobTypeUnknown && c.MobFamily == proto.MobFamily_MobFamilyUnknown {
		return true
	}
	for _, target := range encounter.GetTargets() {
		if c.MobType != proto.MobType_MobTypeUnknown && target.MobType != c.MobType {
			return false
		}
		if c.MobFamily != proto.MobFamily_MobFamilyUnknown && target.MobFamily != c.MobFamily {
			return false
		}
	}
	return len(encounter.GetTargets()) > 0
}
