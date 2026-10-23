package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["rogue_combat"] = &specSetup{
		uiDir: "ui/rogue",
		raid:  rogueLevel60Raid,
		consumes: func() *proto.Consumes {
			return rogueLevel60Consumes()
		},
		professions: [2]proto.Profession{proto.Profession_Engineering, proto.Profession_Leatherworking},
		withSpec:    rogueWithSpec,
	}
	// Combat Daggers drinks the Flask of Natural Accuracy (ui/rogue/presets.ts
	// P1CombatDaggersConsumes): on the 8% sets its 5% hit is worth more to the white hits
	// than Precision's expertise, +1.2 to +2.1 at 120 / 180 / 300 sec on both seed sets.
	specs["rogue_combat_daggers"] = &specSetup{
		uiDir: "ui/rogue",
		raid:  rogueLevel60Raid,
		consumes: func() *proto.Consumes {
			consumes := rogueLevel60Consumes()
			consumes.Flask = proto.Flask_FlaskOfNaturalAccuracy
			return consumes
		},
		professions: [2]proto.Profession{proto.Profession_Engineering, proto.Profession_Leatherworking},
		withSpec:    rogueWithSpec,
	}
}
