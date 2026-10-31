package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["rogue_assassination"] = &specSetup{
		uiDir: "ui/rogue",
		raid:  rogueLevel60Raid,
		// Mutilate leaves the Goblin Land Mine out: a tie (-0.3 to +0.3 at 120 / 180 / 300 sec),
		// and explosives go in only when they gain (ui/rogue/presets.ts P1SapperOnlyConsumes,
		// standards 11).
		consumes: func() *proto.Consumes {
			consumes := rogueLevel60Consumes()
			consumes.FillerExplosive = proto.Explosive_ExplosiveUnknown
			return consumes
		},
		professions: [2]proto.Profession{proto.Profession_Engineering, proto.Profession_Leatherworking},
		withSpec:    rogueWithSpec,
	}
}
