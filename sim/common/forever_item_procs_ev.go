package common

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Chance-on-hit weapons the sim had no effect for, whose Forever rate wowsims/forever measured or
// observed on Warcraft Logs (the EV id beside each). The proc spell and its numbers are the client
// rows wowsims/forever's spell data carries; an effect a damage sim leaves out (a slow, a stun) is
// left out here too. Classic's versions stay without effects.
func init() {
	// EV-0003: Drakefang Butcher (12463): Rend (14118), Physical 23 every 1 s for 7 s, 1.5 procs a minute.
	newForeverWeaponProc(12463, "Drakefang Butcher", foreverProc{PPM: 1.5, SpellID: 14118, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, Tick: 23, Ticks: 7, TickLength: 1000 * time.Millisecond})

	// EV-0003: Gatorbite Axe (17730): Rend (21949), Physical 30 every 1 s for 7 s, 1.2 procs a minute.
	newForeverWeaponProc(17730, "Gatorbite Axe", foreverProc{PPM: 1.2, SpellID: 21949, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, Tick: 30, Ticks: 7, TickLength: 1000 * time.Millisecond})

	// EV-0003: Verdant Keeper's Aim (17753): Keeper's Sting (29655), Nature 23.33 to 32.67, 3.5 procs a minute.
	newForeverWeaponProc(17753, "Verdant Keeper's Aim", foreverProc{PPM: 3.5, SpellID: 29655, School: core.SpellSchoolNature, DefenseType: core.DefenseTypeRanged, DmgMin: 23.3333, DmgRange: 9.33333})

	// EV-0003: Cremation Longbow (275441): Cremate (1297921), Fire 56, 2.3 procs a minute.
	newForeverWeaponProc(275441, "Cremation Longbow", foreverProc{PPM: 2.3, SpellID: 1297921, School: core.SpellSchoolFire, DefenseType: core.DefenseTypeRanged, DmgMin: 56})

	// EV-0022: Ripsaw (9478): Wound (1292573), Physical 70, 2 procs a minute.
	newForeverWeaponProc(9478, "Ripsaw", foreverProc{PPM: 2, SpellID: 1292573, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, DmgMin: 70})

	// EV-0024: Sword of Dismemberment (287005): Dismember (1324687), Physical 335 to 1005, 0.6 procs a minute.
	newForeverWeaponProc(287005, "Sword of Dismemberment", foreverProc{PPM: 0.6, SpellID: 1324687, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, DmgMin: 335, DmgRange: 670})

	// EV-0122: Rot-Covered Harpoon (274963): Rot (1295744), Nature 24 every 1 s for 7 s, 1.45 procs a minute.
	newForeverWeaponProc(274963, "Rot-Covered Harpoon", foreverProc{PPM: 1.45, SpellID: 1295744, School: core.SpellSchoolNature, DefenseType: core.DefenseTypeMagic, Tick: 24, Ticks: 7, TickLength: 1000 * time.Millisecond})

	// EV-0146: Blunderflame Bone Bow (275439): Blunderflame (1296730), Fire 1.4 to 110.6, 1.95 procs a minute.
	newForeverWeaponProc(275439, "Blunderflame Bone Bow", foreverProc{PPM: 1.95, SpellID: 1296730, School: core.SpellSchoolFire, DefenseType: core.DefenseTypeRanged, DmgMin: 1.4, DmgRange: 109.2})

	// EV-0153: Plaguefang (279876): Poison (1309315), Nature 2 every 1 s for 10 s, 2 procs a minute.
	newForeverWeaponProc(279876, "Plaguefang", foreverProc{PPM: 2, SpellID: 1309315, School: core.SpellSchoolNature, DefenseType: core.DefenseTypeMagic, Tick: 2, Ticks: 10, TickLength: 1000 * time.Millisecond})

	// EV-0153: Bleeding Crescent (6738): Rend (16403), Physical 17 every 6 s for 30 s, once a minute.
	newForeverWeaponProc(6738, "Bleeding Crescent", foreverProc{PPM: 1, SpellID: 16403, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, Tick: 17, Ticks: 5, TickLength: 6000 * time.Millisecond})

	// EV-0153: Wyvern Tailspike (5752): Poison (16400), Nature 6 every 3 s for 15 s, 2.1 procs a minute.
	newForeverWeaponProc(5752, "Wyvern Tailspike", foreverProc{PPM: 2.1, SpellID: 16400, School: core.SpellSchoolNature, DefenseType: core.DefenseTypeMagic, Tick: 6, Ticks: 5, TickLength: 3000 * time.Millisecond})

	// EV-0177: Siegebreaker's Blaster (282561): Burning Shell (1314040), Fire 7 every 3 s for 9 s, 5.25 procs a minute.
	newForeverWeaponProc(282561, "Siegebreaker's Blaster", foreverProc{PPM: 5.25, SpellID: 1314040, School: core.SpellSchoolFire, DefenseType: core.DefenseTypeRanged, Tick: 7, Ticks: 3, TickLength: 3000 * time.Millisecond})

	// EV-0183: Bloodpike (13057): Rend (18202), Physical 14 every 2 s for 14 s, 0.9 procs a minute.
	newForeverWeaponProc(13057, "Bloodpike", foreverProc{PPM: 0.9, SpellID: 18202, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, Tick: 14, Ticks: 7, TickLength: 2000 * time.Millisecond})

	// EV-0183: Darkmist Dirk (285102): Poison (1317583), Nature 8 every 3 s for 9 s, 3.65 procs a minute.
	newForeverWeaponProc(285102, "Darkmist Dirk", foreverProc{PPM: 3.65, SpellID: 1317583, School: core.SpellSchoolNature, DefenseType: core.DefenseTypeMagic, Tick: 8, Ticks: 3, TickLength: 3000 * time.Millisecond})

	// EV-0183: Flesh Piercer (3336): Rend (18078), Physical 6 every 6 s for 30 s, 0.95 procs a minute.
	newForeverWeaponProc(3336, "Flesh Piercer", foreverProc{PPM: 0.95, SpellID: 18078, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, Tick: 6, Ticks: 5, TickLength: 6000 * time.Millisecond})

	// EV-0183: Naraxis' Fang (4449): Poison (16400), Nature 6 every 3 s for 15 s, 2.1 procs a minute.
	newForeverWeaponProc(4449, "Naraxis' Fang", foreverProc{PPM: 2.1, SpellID: 16400, School: core.SpellSchoolNature, DefenseType: core.DefenseTypeMagic, Tick: 6, Ticks: 5, TickLength: 3000 * time.Millisecond})

	// EV-0183: Needletooth's Needletooth (282703): Puncture (1314305), Physical 21, 4 procs a minute.
	newForeverWeaponProc(282703, "Needletooth's Needletooth", foreverProc{PPM: 4, SpellID: 1314305, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, DmgMin: 21})

	// EV-0183: Venom Web Fang (899): Poison (18077), Nature 3 every 3 s for 15 s, 1.1 procs a minute.
	newForeverWeaponProc(899, "Venom Web Fang", foreverProc{PPM: 1.1, SpellID: 18077, School: core.SpellSchoolNature, DefenseType: core.DefenseTypeMagic, Tick: 3, Ticks: 5, TickLength: 3000 * time.Millisecond})

	// EV-0186: Whipfang's Skinsearer (284386): Fire Blast (1316039), Fire 67.2 to 100.8, 0.75 procs a minute.
	newForeverWeaponProc(284386, "Whipfang's Skinsearer", foreverProc{PPM: 0.75, SpellID: 1316039, School: core.SpellSchoolFire, DefenseType: core.DefenseTypeRanged, DmgMin: 67.2, DmgRange: 33.6})

	// EV-0186: Winter's Bite (10623): Frostbolt (13439), Frost 40 to 60, 2.95 procs a minute.
	newForeverWeaponProc(10623, "Winter's Bite", foreverProc{PPM: 2.95, SpellID: 13439, School: core.SpellSchoolFrost, DefenseType: core.DefenseTypeMagic, DmgMin: 40, DmgRange: 20})

	// EV-0136: Axe of Inevitable Ruin (274484): Wound (1294226), Physical 140 to 196, 2.05 procs a minute.
	newForeverWeaponProc(274484, "Axe of Inevitable Ruin", foreverProc{PPM: 2.05, SpellID: 1294226, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, DmgMin: 140, DmgRange: 56})

	// EV-0136: Barbaric Crossbow (272999): Wound (1291551), Physical 14, 3.05 procs a minute.
	newForeverWeaponProc(272999, "Barbaric Crossbow", foreverProc{PPM: 3.05, SpellID: 1291551, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeRanged, DmgMin: 14})

	// EV-0136: Black Malice (3194): Shadow Bolt (1292155), Shadow 55 to 85, 1.85 procs a minute.
	newForeverWeaponProc(3194, "Black Malice", foreverProc{PPM: 1.85, SpellID: 1292155, School: core.SpellSchoolShadow, DefenseType: core.DefenseTypeMagic, DmgMin: 55, DmgRange: 30})

	// EV-0136: Blade of the Wretched (10803): Corruption (18088), Shadow 30 every 1 s for 3 s, 2 procs a minute.
	newForeverWeaponProc(10803, "Blade of the Wretched", foreverProc{PPM: 2, SpellID: 18088, School: core.SpellSchoolShadow, DefenseType: core.DefenseTypeMagic, Tick: 30, Ticks: 3, TickLength: 1000 * time.Millisecond})

	// EV-0136: Bloodspiller (7753): Rend (18200), Physical 25 every 1 s for 10 s, 0.9 procs a minute.
	newForeverWeaponProc(7753, "Bloodspiller", foreverProc{PPM: 0.9, SpellID: 18200, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, Tick: 25, Ticks: 10, TickLength: 1000 * time.Millisecond})

	// EV-0136: Boneslicer (273819): Wound (1292573), Physical 70, 2.95 procs a minute.
	newForeverWeaponProc(273819, "Boneslicer", foreverProc{PPM: 2.95, SpellID: 1292573, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, DmgMin: 70})

	// EV-0136: Claw of Celebras (17738): Poison (21952), Nature 13 every 1 s for 7 s, 2.05 procs a minute.
	newForeverWeaponProc(17738, "Claw of Celebras", foreverProc{PPM: 2.05, SpellID: 21952, School: core.SpellSchoolNature, DefenseType: core.DefenseTypeMagic, Tick: 13, Ticks: 7, TickLength: 1000 * time.Millisecond})

	// EV-0136: Cobalt Crusher (7730): Frost Blast (18204), Frost 127.2 to 138.8, 2 procs a minute.
	newForeverWeaponProc(7730, "Cobalt Crusher", foreverProc{PPM: 2, SpellID: 18204, School: core.SpellSchoolFrost, DefenseType: core.DefenseTypeMagic, DmgMin: 127.217, DmgRange: 11.5652})

	// EV-0136: Darkwater Talwar (11121): Shadow Bolt (16408), Shadow 28, 2.95 procs a minute.
	newForeverWeaponProc(11121, "Darkwater Talwar", foreverProc{PPM: 2.95, SpellID: 16408, School: core.SpellSchoolShadow, DefenseType: core.DefenseTypeMagic, DmgMin: 28})

	// EV-0136: Galgann's Fireblaster (9412): Fire Blast (29644), Fire 22.4 to 33.6, 4 procs a minute.
	newForeverWeaponProc(9412, "Galgann's Fireblaster", foreverProc{PPM: 4, SpellID: 29644, School: core.SpellSchoolFire, DefenseType: core.DefenseTypeRanged, DmgMin: 22.4, DmgRange: 11.2})

	// EV-0136: Glutton's Cleaver (10772): Rend (18075), Physical 10 every 1 s for 7 s, 2 procs a minute.
	newForeverWeaponProc(10772, "Glutton's Cleaver", foreverProc{PPM: 2, SpellID: 18075, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, Tick: 10, Ticks: 7, TickLength: 1000 * time.Millisecond})

	// EV-0136: Grimclaw (1481): Shadow Bolt (1292158), Shadow 42, 2.05 procs a minute.
	newForeverWeaponProc(1481, "Grimclaw", foreverProc{PPM: 2.05, SpellID: 1292158, School: core.SpellSchoolShadow, DefenseType: core.DefenseTypeMagic, DmgMin: 42})

	// EV-0136: Ironforge Greathammer (279895): Firebolt (1322295), Fire 42, 2.9 procs a minute.
	newForeverWeaponProc(279895, "Ironforge Greathammer", foreverProc{PPM: 2.9, SpellID: 1322295, School: core.SpellSchoolFire, DefenseType: core.DefenseTypeMagic, DmgMin: 42})

	// EV-0136: Jom-jom's Gabbar (274952): Bolder'ok Boulder (1295736), Physical 98, 1.85 procs a minute.
	newForeverWeaponProc(274952, "Jom-jom's Gabbar", foreverProc{PPM: 1.85, SpellID: 1295736, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeRanged, DmgMin: 98})

	// EV-0136: Killmaim (13016): Rend (13318), Physical 10 every 3 s for 30 s, once a minute.
	newForeverWeaponProc(13016, "Killmaim", foreverProc{PPM: 1, SpellID: 13318, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, Tick: 10, Ticks: 10, TickLength: 3000 * time.Millisecond})

	// EV-0136: Loksey's Training Stick (7710): Tough Love (1293482), Physical 91, 1.95 procs a minute.
	newForeverWeaponProc(7710, "Loksey's Training Stick", foreverProc{PPM: 1.95, SpellID: 1293482, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, DmgMin: 91})

	// EV-0136: Naga Heartpiercer (3078): Pierce Heart (1292647), Physical 42, 2.1 procs a minute.
	newForeverWeaponProc(3078, "Naga Heartpiercer", foreverProc{PPM: 2.1, SpellID: 1292647, School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, DmgMin: 42})

	// EV-0136: Night Reaver (1318): Shadow Bolt (13480), Shadow 67.2 to 100.8, 1.7 procs a minute.
	newForeverWeaponProc(1318, "Night Reaver", foreverProc{PPM: 1.7, SpellID: 13480, School: core.SpellSchoolShadow, DefenseType: core.DefenseTypeMagic, DmgMin: 67.2, DmgRange: 33.6})

	// EV-0136: Pyromancer's Baton (286650): Ablaze (1320925), Fire 30 every 3 s for 9 s, 1.95 procs a minute.
	newForeverWeaponProc(286650, "Pyromancer's Baton", foreverProc{PPM: 1.95, SpellID: 1320925, School: core.SpellSchoolFire, DefenseType: core.DefenseTypeMagic, Tick: 30, Ticks: 3, TickLength: 3000 * time.Millisecond})

	// EV-0136: Quillshooter (10567): Quill Shot (29646), Nature 26.56 to 39.44, 2.85 procs a minute.
	newForeverWeaponProc(10567, "Quillshooter", foreverProc{PPM: 2.85, SpellID: 29646, School: core.SpellSchoolNature, DefenseType: core.DefenseTypeRanged, DmgMin: 26.561, DmgRange: 12.878})

	// EV-0136: Scorpion Sting (1265): Poison (18208), Nature 13 every 5 s for 25 s, 0.95 procs a minute.
	newForeverWeaponProc(1265, "Scorpion Sting", foreverProc{PPM: 0.95, SpellID: 18208, School: core.SpellSchoolNature, DefenseType: core.DefenseTypeMagic, Tick: 13, Ticks: 5, TickLength: 5000 * time.Millisecond})

	// EV-0136: Searing Blade (12992): Fireball (1299850), Fire 80, 6 every 2 s for 6 s, 1.9 procs a minute.
	newForeverWeaponProc(12992, "Searing Blade", foreverProc{PPM: 1.9, SpellID: 1299850, School: core.SpellSchoolFire, DefenseType: core.DefenseTypeMagic, DmgMin: 80, Tick: 6, Ticks: 3, TickLength: 2000 * time.Millisecond})

	// EV-0136: Skeletal Club (2256): Shadow Bolt (13440), Shadow 30, 0.95 procs a minute.
	newForeverWeaponProc(2256, "Skeletal Club", foreverProc{PPM: 0.95, SpellID: 13440, School: core.SpellSchoolShadow, DefenseType: core.DefenseTypeMagic, DmgMin: 30})

	// EV-0136: Sliverblade (5756): Frost Blast (1293653), Frost 63, 2 procs a minute.
	newForeverWeaponProc(5756, "Sliverblade", foreverProc{PPM: 2, SpellID: 1293653, School: core.SpellSchoolFrost, DefenseType: core.DefenseTypeMagic, DmgMin: 63})

	// EV-0136: Sword of Corruption (13032): Corruption (17510), Shadow 10 every 1 s for 3 s, once a minute.
	newForeverWeaponProc(13032, "Sword of Corruption", foreverProc{PPM: 1, SpellID: 17510, School: core.SpellSchoolShadow, DefenseType: core.DefenseTypeMagic, Tick: 10, Ticks: 3, TickLength: 1000 * time.Millisecond})

	// EV-0136: Tainted Pierce (8225): Corruption (13530), Shadow 21 every 1 s for 3 s, 2.05 procs a minute.
	newForeverWeaponProc(8225, "Tainted Pierce", foreverProc{PPM: 2.05, SpellID: 13530, School: core.SpellSchoolShadow, DefenseType: core.DefenseTypeMagic, Tick: 21, Ticks: 3, TickLength: 1000 * time.Millisecond})

	// EV-0136: Twilight Forgehammer (281665): Forge Blast (1312176), Fire 51, 2.95 procs a minute.
	newForeverWeaponProc(281665, "Twilight Forgehammer", foreverProc{PPM: 2.95, SpellID: 1312176, School: core.SpellSchoolFire, DefenseType: core.DefenseTypeMagic, DmgMin: 51})

	// EV-0136: Twilight Maul (273841): Lightning Bolt (1292679), Nature 70, 1.8 procs a minute.
	newForeverWeaponProc(273841, "Twilight Maul", foreverProc{PPM: 1.8, SpellID: 1292679, School: core.SpellSchoolNature, DefenseType: core.DefenseTypeMagic, DmgMin: 70})

	// EV-0136: Venomstrike (6469): Venom Shot (29653), Nature 22.84 to 33.16, 1.85 procs a minute.
	newForeverWeaponProc(6469, "Venomstrike", foreverProc{PPM: 1.85, SpellID: 29653, School: core.SpellSchoolNature, DefenseType: core.DefenseTypeRanged, DmgMin: 22.8421, DmgRange: 10.3158})
}
