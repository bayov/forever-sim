package main

import (
	"fmt"

	"github.com/wowsims/classic/sim/core/proto"
)

// What each class can wear and wield (proto_utils/utils.ts), for the requests we make ourselves
// (config.go). The UI's requests have their own.
type classGear struct {
	// The heaviest armor we can wear. A class with mail or plate gets it at 40.
	armor              proto.ArmorType
	weapons, twoHanded []proto.WeaponType
	ranged             []proto.RangedWeaponType
	canDualWield       bool
	// The level we learn Dual Wield at.
	//
	// A warrior's is a guess (notes.md, Need to Verify). The client doesn't give the trainer's
	// level, but a beta combat log has a warrior dual wielding at 13, so it's not Classic's 20.
	dualWieldLevel int32
}

var classGears = map[proto.Class]classGear{
	proto.Class_ClassShaman: {
		armor: proto.ArmorType_ArmorTypeMail,
		weapons: []proto.WeaponType{
			proto.WeaponType_WeaponTypeAxe, proto.WeaponType_WeaponTypeDagger, proto.WeaponType_WeaponTypeFist, proto.WeaponType_WeaponTypeMace,
			proto.WeaponType_WeaponTypeOffHand, proto.WeaponType_WeaponTypeShield, proto.WeaponType_WeaponTypeStaff,
		},
		twoHanded: []proto.WeaponType{proto.WeaponType_WeaponTypeAxe, proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeStaff},
		ranged:    []proto.RangedWeaponType{proto.RangedWeaponType_RangedWeaponTypeTotem},
	},
	proto.Class_ClassWarrior: {
		armor: proto.ArmorType_ArmorTypePlate,
		// The weapon master teaches the ones we don't start with, and we count them as trained.
		weapons: []proto.WeaponType{
			proto.WeaponType_WeaponTypeAxe, proto.WeaponType_WeaponTypeDagger, proto.WeaponType_WeaponTypeFist, proto.WeaponType_WeaponTypeMace,
			proto.WeaponType_WeaponTypeOffHand, proto.WeaponType_WeaponTypePolearm, proto.WeaponType_WeaponTypeShield,
			proto.WeaponType_WeaponTypeStaff, proto.WeaponType_WeaponTypeSword,
		},
		twoHanded: []proto.WeaponType{
			proto.WeaponType_WeaponTypeAxe, proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypePolearm,
			proto.WeaponType_WeaponTypeStaff, proto.WeaponType_WeaponTypeSword,
		},
		ranged: []proto.RangedWeaponType{
			proto.RangedWeaponType_RangedWeaponTypeBow, proto.RangedWeaponType_RangedWeaponTypeCrossbow,
			proto.RangedWeaponType_RangedWeaponTypeGun, proto.RangedWeaponType_RangedWeaponTypeThrown,
		},
		canDualWield:   true,
		dualWieldLevel: 13,
	},
}

// classRules are the class's rules as a request has them.
func classRules(class proto.Class) (*proto.LevelingGearClassRules, error) {
	gear, ok := classGears[class]
	if !ok {
		return nil, fmt.Errorf("no gear rules for %s yet, add them to classGears (class_rules.go)", class)
	}
	rules := &proto.LevelingGearClassRules{
		WeaponTypes:        gear.weapons,
		TwoHandWeaponTypes: gear.twoHanded,
		RangedWeaponTypes:  gear.ranged,
		CanDualWield:       gear.canDualWield,
		DualWieldLevel:     gear.dualWieldLevel,
	}
	for level := 1; level <= maxPlanLevel; level++ {
		armor := gear.armor
		if level < 40 && armor >= proto.ArmorType_ArmorTypeMail {
			armor--
		}
		rules.MaxArmorType = append(rules.MaxArmorType, armor)
	}
	return rules, nil
}
