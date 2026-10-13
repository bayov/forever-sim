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
}
