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
	// rogue_combat keeps the Goblin Land Mine, as Seal Fate Daggers does. Combat Swords and Seal
	// Fate Swords leave it out, since its 1 sec cast resets the swing and it loses for them
	// (ui/rogue/presets.ts P1SapperOnlyConsumes, standards 11: Combat Swords 1.3 to 2.9 and Seal Fate
	// Swords -0.1 to 1.3 behind at 120 / 180 / 300 sec).
	specs["rogue_combat_swords"] = withConsumes(specs["rogue_combat"], func(consumes *proto.Consumes) {
		consumes.FillerExplosive = proto.Explosive_ExplosiveUnknown
	})
	// Combat Daggers drinks the Flask of Natural Accuracy (ui/rogue/presets.ts
	// P1CombatDaggersConsumes): on the 8% sets its 5% hit is worth more to the white hits
	// than Precision's expertise, +1.2 to +2.1 at 120 / 180 / 300 sec on both seed sets. It
	// leaves the Land Mine out too (0.3 to 1.8 behind).
	specs["rogue_combat_daggers"] = withConsumes(specs["rogue_combat"], func(consumes *proto.Consumes) {
		consumes.Flask = proto.Flask_FlaskOfNaturalAccuracy
		consumes.FillerExplosive = proto.Explosive_ExplosiveUnknown
	})
}
