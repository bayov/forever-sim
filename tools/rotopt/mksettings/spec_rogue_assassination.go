package main

import "github.com/wowsims/classic/sim/core/proto"

func init() {
	specs["rogue_assassination"] = &specSetup{
		uiDir: "ui/rogue",
		raid:  rogueLevel60Raid,
		consumes: func() *proto.Consumes {
			return rogueLevel60Consumes(proto.WeaponImbue_DeadlyPoison, proto.WeaponImbue_InstantPoison)
		},
		professions: [2]proto.Profession{proto.Profession_Engineering, proto.Profession_Leatherworking},
		withSpec:    rogueWithSpec,
	}
}
