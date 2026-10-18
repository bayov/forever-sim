package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// ForeverItems are items that exist only in WoW Forever, entered by hand.
//
// Since 2026-09-22 the wowhead Forever listing (forever_wowhead.go, required level
// 25 and under) covers these too and its stats win, this file still gives the quest
// names and levels wowhead lacks and the items above level 25.
//
// The Classic inputs (wowhead dump, atlasloot, wago) know nothing about them,
// so we type the stats from the wowhead Forever tooltips
// (https://nether.wowhead.com/forever/tooltip/item/ID?dataEnv=16&locale=0) and
// merge them in after the Classic items. The quest names and levels match
// QuestieDB's Forever data and Wowhead's Forever listing (checked 2026-10-17:
// every name, and every level but Leaders of the Fang's, which QuestieDB has
// from level 11). Two quests, Abominable Creatures and The Windshaper's
// Wrath, aren't on Wowhead yet. Chance on hit effects
// (Plaguefang) are not implemented in the sim, the weapon is listed with its
// plain stats.
//
// Quest rewards carry no required level on the beta, so we leave
// RequiredLevel at 0 and the level 20 gear search takes them on item level
// (up to level + 6). The quest comments give the level a character needs to
// pick the quest up. Three Blackfathom Deeps rewards (Silvered Gauntlets
// 270025, Dark Ritual Leggings 270031, Cultist's Armguards 270032) have no
// wowhead tooltip yet and are left out.

// foreverQuest is a quest source for the reward item.
func foreverQuest(id int32, name string) *proto.UIItemSource {
	return &proto.UIItemSource{Source: &proto.UIItemSource_Quest{Quest: &proto.QuestSource{Id: id, Name: name}}}
}

// foreverRep is a reputation vendor source for the item. Ratchet and the
// other goblin towns sell gear to both factions in Forever.
func foreverRep(factionId int32, level proto.RepLevel) *proto.UIItemSource {
	return &proto.UIItemSource{Source: &proto.UIItemSource_Rep{Rep: &proto.RepSource{RepFactionId: factionId, RepLevel: level}}}
}

// foreverRare is a drop off a rare in the zone. We have no NPC ID for the Forever rares,
// so the rare goes in by name.
func foreverRare(zoneId int32, name string) *proto.UIItemSource {
	return &proto.UIItemSource{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{ZoneId: zoneId, OtherName: name + " (rare)"}}}
}

// ForeverFactions are reputations the Classic inputs do not carry.
var ForeverFactions = []*proto.UIFaction{
	{Id: 470, Name: "Ratchet", Expansion: proto.Expansion_ExpansionVanilla},
}

var ForeverItems = []*proto.UIItem{
	// Ruins of Lordaeron (5 player dungeon, levels 15 to 20) quest rewards.
	// The quest rewards have no required level on the beta, so the level 20
	// gear search takes them on item level (24 is within level + 6).
	{
		Id: 251533, Name: "Forsaken Greataxe", Icon: "inv_axe_18",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeTwoHand,
		Stats:           stats.Stats{stats.Strength: 8, stats.Stamina: 4}.ToFloatArray(),
		WeaponDamageMin: 49, WeaponDamageMax: 74, WeaponSpeed: 3.0,
		Ilvl: 24, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(92422, "The Wrath of Rath'mael")},
	},
	{
		Id: 251534, Name: "Gnarled Necromancer's Staff", Icon: "inv_staff_16",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:           stats.Stats{stats.Intellect: 10}.ToFloatArray(),
		WeaponDamageMin: 45, WeaponDamageMax: 69, WeaponSpeed: 2.8,
		Ilvl: 24, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(92422, "The Wrath of Rath'mael")},
	},
	{
		Id: 251485, Name: "Edwards' Knife", Icon: "inv_weapon_shortblade_14",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeDagger, HandType: proto.HandType_HandTypeOneHand,
		Stats:           stats.Stats{stats.Agility: 4, stats.Stamina: 4}.ToFloatArray(),
		WeaponDamageMin: 17, WeaponDamageMax: 33, WeaponSpeed: 1.6,
		Ilvl: 24, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(92401, "A Frightened Request")},
	},
	{
		Id: 251486, Name: "Tabitha's Cuffs", Icon: "inv_bracer_10",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 3, stats.Intellect: 6, stats.Armor: 18}.ToFloatArray(),
		Ilvl:  24, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(92401, "A Frightened Request")},
	},
	{
		Id: 279874, Name: "The Stitcher", Icon: "inv_mace_08",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeMainHand,
		Stats:           stats.Stats{stats.Stamina: 6}.ToFloatArray(),
		WeaponDamageMin: 26, WeaponDamageMax: 49, WeaponSpeed: 2.4,
		Ilvl: 24, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(92421, "Light's Justice")},
	},
	{
		Id: 279875, Name: "Spare Part Bindings", Icon: "inv_bracer_07",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Agility: 3, stats.Stamina: 5, stats.Spirit: 3, stats.Armor: 40}.ToFloatArray(),
		Ilvl:  24, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(92421, "Light's Justice")},
	},
	{
		Id: 279876, Name: "Plaguefang", Icon: "inv_weapon_shortblade_54",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
		Stats:           stats.Stats{}.ToFloatArray(),
		WeaponDamageMin: 23, WeaponDamageMax: 43, WeaponSpeed: 2.1,
		Ilvl: 24, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla, Unique: true,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(95216, "The New Plague")},
	},
	{
		Id: 279877, Name: "Blight Gloves", Icon: "inv_gauntlets_21",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Intellect: 7, stats.Spirit: 5, stats.NatureResistance: 3, stats.Armor: 26}.ToFloatArray(),
		Ilvl:  24, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(95216, "The New Plague")},
	},
	{
		Id: 279864, Name: "Monstrous Cleaver", Icon: "inv_sword_23",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeTwoHand,
		Stats:           stats.Stats{stats.Stamina: 9, stats.AttackPower: 10}.ToFloatArray(),
		WeaponDamageMin: 52, WeaponDamageMax: 78, WeaponSpeed: 3.3,
		Ilvl: 23, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(95250, "Abominable Creatures")},
	},
	{
		Id: 279865, Name: "Grave Shroud", Icon: "inv_misc_cape_10",
		Type:  proto.ItemType_ItemTypeBack,
		Stats: stats.Stats{stats.Strength: 3, stats.Agility: 2, stats.Stamina: 5, stats.Armor: 20}.ToFloatArray(),
		Ilvl:  23, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(95250, "Abominable Creatures")},
	},
	{
		Id: 279867, Name: "Slain Baron's Signet", Icon: "inv_jewelry_ring_19",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 5, stats.Defense: 2}.ToFloatArray(),
		Ilvl:  23, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(95250, "Abominable Creatures")},
	},
	{
		Id: 279868, Name: "Duty Bound Leggings", Icon: "inv_pants_06",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Agility: 6, stats.Stamina: 9, stats.AttackPower: 12, stats.Armor: 81}.ToFloatArray(),
		Ilvl:  24, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(95195, "Bloodied Insignia")},
	},
	{
		Id: 279869, Name: "Remembrance Armor", Icon: "inv_chest_chain",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Strength: 4, stats.Stamina: 10, stats.Spirit: 4, stats.Armor: 198}.ToFloatArray(),
		Ilvl:  24, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(95195, "Bloodied Insignia")},
	},

	// Ragefire Chasm quest rewards.
	// Returning the Lost Satchel (level 16, from level 9, Horde).
	{
		Id: 270003, Name: "Garrison Cuffs", Icon: "inv_bracer_13",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Spirit: 5, stats.Armor: 78}.ToFloatArray(),
		Ilvl:  18, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(5724, "Returning the Lost Satchel")},
	},

	// The Hall of Thanes (new Forever dungeon, levels 13 to 18) quest rewards.
	// An Ancient Grudge (level 15, from level 10).
	{
		Id: 279899, Name: "Catacomb Cloak", Icon: "inv_misc_cape_11",
		Type:  proto.ItemType_ItemTypeBack,
		Stats: stats.Stats{stats.Stamina: 3, stats.AttackPower: 6, stats.Armor: 17}.ToFloatArray(),
		Ilvl:  17, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		Sources: []*proto.UIItemSource{foreverQuest(96395, "An Ancient Grudge")},
	},
	{
		Id: 279900, Name: "Deepgrave Trousers", Icon: "inv_pants_01",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Strength: 7, stats.Agility: 3, stats.Spirit: 3, stats.Armor: 70}.ToFloatArray(),
		Ilvl:  17, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		Sources: []*proto.UIItemSource{foreverQuest(96395, "An Ancient Grudge")},
	},
	// Important Heirlooms (level 15, from level 10).
	{
		Id: 279898, Name: "Dwarven Tome", Icon: "inv_misc_book_07",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats: stats.Stats{stats.Spirit: 2, stats.SpellPower: 5}.ToFloatArray(),
		Ilvl:  17, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		Sources: []*proto.UIItemSource{foreverQuest(96403, "Important Heirlooms")},
	},
	{
		Id: 280096, Name: "Tomb Robber's Gloves", Icon: "inv_gauntlets_18",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 3, stats.Intellect: 6, stats.Armor: 21}.ToFloatArray(),
		Ilvl:  17, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		Sources: []*proto.UIItemSource{foreverQuest(96403, "Important Heirlooms")},
	},
	// The Restless Dead (level 15, from level 10, Alliance).
	{
		Id: 279897, Name: "Dusty Belt", Icon: "inv_belt_03",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Agility: 5, stats.Spirit: 4, stats.Armor: 45}.ToFloatArray(),
		Ilvl:  17, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(96394, "The Restless Dead")},
	},
	{
		Id: 280095, Name: "Cryptwalker Bracers", Icon: "inv_bracer_03",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Strength: 4, stats.Stamina: 3, stats.Armor: 76}.ToFloatArray(),
		Ilvl:  17, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(96394, "The Restless Dead")},
	},
	// Old Ironforge Incursion (level 16, from level 9).
	{
		Id: 279894, Name: "Calibrated Blunderbuss", Icon: "inv_weapon_rifle_01",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeGun,
		Stats:           stats.Stats{stats.Stamina: 3}.ToFloatArray(),
		WeaponDamageMin: 25, WeaponDamageMax: 48, WeaponSpeed: 2.30,
		Ilvl: 18, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		Sources: []*proto.UIItemSource{foreverQuest(96393, "Old Ironforge Incursion")},
	},
	{
		Id: 279895, Name: "Ironforge Greathammer", Icon: "inv_hammer_21",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeTwoHand,
		Stats:           stats.Stats{stats.Strength: 4, stats.Stamina: 7}.ToFloatArray(),
		WeaponDamageMin: 39, WeaponDamageMax: 60, WeaponSpeed: 3.10,
		Ilvl: 18, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		Sources: []*proto.UIItemSource{foreverQuest(96393, "Old Ironforge Incursion")},
	},
	{
		Id: 279896, Name: "Deepblaze", Icon: "inv_wand_11",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats:           stats.Stats{stats.Intellect: 3}.ToFloatArray(),
		WeaponDamageMin: 14, WeaponDamageMax: 27, WeaponSpeed: 1.70,
		Ilvl: 18, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		Sources: []*proto.UIItemSource{foreverQuest(96393, "Old Ironforge Incursion")},
	},

	// Deadmines quest rewards.
	// Red Silk Bandanas (level 17, from level 14, Alliance).
	{
		Id: 270005, Name: "Monastic Hammer", Icon: "inv_hammer_11",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeMainHand,
		Stats:           stats.Stats{stats.Intellect: 2}.ToFloatArray(),
		WeaponDamageMin: 14, WeaponDamageMax: 26, WeaponSpeed: 2.10,
		Ilvl: 19, Phase: 1, Quality: proto.ItemQuality_ItemQualityUncommon, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(214, "Red Silk Bandanas")},
	},
	// Collecting Memories (level 18, from level 14, Alliance).
	{
		Id: 270007, Name: "Worn Miner's Waistcord", Icon: "inv_belt_22",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 2, stats.Intellect: 3, stats.Armor: 18}.ToFloatArray(),
		Ilvl:  18, Phase: 1, Quality: proto.ItemQuality_ItemQualityUncommon, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(168, "Collecting Memories")},
	},
	// Oh Brother... (level 20, from level 15, Alliance).
	{
		Id: 270012, Name: "Miner's Workgloves", Icon: "inv_gauntlets_16",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 5, stats.Armor: 21}.ToFloatArray(),
		Ilvl:  20, Phase: 1, Quality: proto.ItemQuality_ItemQualityUncommon, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(167, "Oh Brother...")},
	},
	{
		Id: 270013, Name: "Miner's Workboots", Icon: "inv_boots_05",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Spirit: 5, stats.Armor: 53}.ToFloatArray(),
		Ilvl:  20, Phase: 1, Quality: proto.ItemQuality_ItemQualityUncommon, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(167, "Oh Brother...")},
	},
	// Underground Assault (level 20, from level 15, Alliance).
	{
		Id: 270015, Name: "Bravo's Armbands", Icon: "inv_bracer_07",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Strength: 2, stats.Agility: 4, stats.Spirit: 4, stats.Armor: 39}.ToFloatArray(),
		Ilvl:  22, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(2040, "Underground Assault")},
	},
	{
		Id: 270016, Name: "Dreamer's Leggings", Icon: "inv_pants_06",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Intellect: 11, stats.SpellPower: 4, stats.Armor: 78}.ToFloatArray(),
		Ilvl:  22, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(2040, "Underground Assault")},
	},

	// Wailing Caverns quest rewards.
	// Serpentbloom (level 18, from level 14, Horde).
	{
		Id: 270008, Name: "Heat Resistant Mitts", Icon: "inv_gauntlets_15",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Agility: 4, stats.Stamina: 1, stats.Armor: 48}.ToFloatArray(),
		Ilvl:  20, Phase: 1, Quality: proto.ItemQuality_ItemQualityUncommon, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(962, "Serpentbloom")},
	},
	{
		Id: 270009, Name: "Safety Boots", Icon: "inv_boots_01",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Strength: 4, stats.Stamina: 2, stats.Armor: 115}.ToFloatArray(),
		Ilvl:  20, Phase: 1, Quality: proto.ItemQuality_ItemQualityUncommon, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(962, "Serpentbloom")},
	},
	// Leaders of the Fang (level 22, from level 10, Horde).
	{
		Id: 270018, Name: "Hammerbone", Icon: "inv_hammer_15",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeTwoHand,
		Stats:           stats.Stats{stats.Strength: 10, stats.Stamina: 7}.ToFloatArray(),
		WeaponDamageMin: 55, WeaponDamageMax: 84, WeaponSpeed: 3.40,
		Ilvl: 24, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(914, "Leaders of the Fang")},
	},

	// Shadowfang Keep quest rewards.
	// Deathstalkers in Shadowfang (level 25, from level 18, Horde).
	{
		Id: 270023, Name: "Tanned Shoulderpads", Icon: "inv_shoulder_27",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Stamina: 9, stats.SpellDamage: 3, stats.HealingPower: 9, stats.Armor: 73}.ToFloatArray(),
		Ilvl:  27, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(1098, "Deathstalkers in Shadowfang")},
	},
	{
		Id: 270024, Name: "Bronzed Shoulderguards", Icon: "inv_shoulder_27",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Stamina: 6, stats.Spirit: 8, stats.Armor: 156}.ToFloatArray(),
		Ilvl:  27, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(1098, "Deathstalkers in Shadowfang")},
	},
	// The Book of Ur (level 26, from level 16, Horde).
	{
		Id: 270030, Name: "Tattered Mittens", Icon: "inv_gauntlets_27",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 8, stats.Spirit: 6, stats.SpellDamage: 4, stats.HealingPower: 11, stats.Armor: 29}.ToFloatArray(),
		Ilvl:  28, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(1013, "The Book of Ur")},
	},

	// Blackfathom Deeps quest rewards.
	// Researching the Corruption (level 24, from level 18, Alliance).
	{
		Id: 270021, Name: "Staghide Armguards", Icon: "inv_bracer_07",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Agility: 5, stats.Stamina: 5, stats.Armor: 42}.ToFloatArray(),
		Ilvl:  26, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(1275, "Researching the Corruption")},
	},

	// Gnomeregan quest rewards.
	// Gyrodrillmatic Excavationators (level 30, from level 20, Alliance).
	{
		Id: 270045, Name: "Operator's Gloves", Icon: "inv_gauntlets_22",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Strength: 7, stats.Stamina: 7, stats.Spirit: 6, stats.Armor: 67}.ToFloatArray(),
		Ilvl:  32, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(2928, "Gyrodrillmatic Excavationators")},
	},

	// Ratchet reputation rewards, all require level 20.
	{
		Id: 274740, Name: "Quilted Cloak", Icon: "inv_misc_cape_05",
		Type:  proto.ItemType_ItemTypeBack,
		Stats: stats.Stats{stats.Agility: 3, stats.Stamina: 4, stats.Armor: 19}.ToFloatArray(),
		Ilvl:  25, RequiredLevel: 20, Phase: 1, Quality: proto.ItemQuality_ItemQualityUncommon, Expansion: proto.Expansion_ExpansionVanilla,
		Sources: []*proto.UIItemSource{foreverRep(470, proto.RepLevel_RepLevelFriendly)},
	},
	{
		Id: 274741, Name: "Rumpled Kilt", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 7, stats.SpellPower: 5, stats.Armor: 34}.ToFloatArray(),
		Ilvl:  25, RequiredLevel: 20, Phase: 1, Quality: proto.ItemQuality_ItemQualityUncommon, Expansion: proto.Expansion_ExpansionVanilla,
		Sources: []*proto.UIItemSource{foreverRep(470, proto.RepLevel_RepLevelFriendly)},
	},
	{
		Id: 274742, Name: "Ratchet Wristwraps", Icon: "inv_bracer_22a",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Agility: 3, stats.Intellect: 4, stats.Armor: 37}.ToFloatArray(),
		Ilvl:  25, RequiredLevel: 20, Phase: 1, Quality: proto.ItemQuality_ItemQualityUncommon, Expansion: proto.Expansion_ExpansionVanilla,
		Sources: []*proto.UIItemSource{foreverRep(470, proto.RepLevel_RepLevelFriendly)},
	},
	{
		Id: 274743, Name: "Defective Samophlange", Icon: "inv_gizmo_06",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats: stats.Stats{stats.Intellect: 5, stats.Spirit: 5}.ToFloatArray(),
		Ilvl:  25, RequiredLevel: 20, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		Sources: []*proto.UIItemSource{foreverRep(470, proto.RepLevel_RepLevelHonored)},
	},
	{
		Id: 274744, Name: "Barrens Basher", Icon: "inv_hammer_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeMainHand,
		Stats:           stats.Stats{stats.Strength: 2, stats.Stamina: 5}.ToFloatArray(),
		WeaponDamageMin: 23, WeaponDamageMax: 44, WeaponSpeed: 2.10,
		Ilvl: 25, RequiredLevel: 20, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		Sources: []*proto.UIItemSource{foreverRep(470, proto.RepLevel_RepLevelHonored)},
	},
	// The Blessing of Kalimdor set, one piece off each of three Forever rares in the
	// Barrens. wowhead has the items but no source, ForeverChanges has the rares from
	// player reports it has not checked itself. The set's 2 piece bonus is movement speed
	// in the Barrens and Stonetalon.
	{
		Id: 285331, Name: "Mark of the Pack Leader",
		Sources: []*proto.UIItemSource{foreverRare(17, "Humar the Pridelord")},
	},
	{
		Id: 285329, Name: "Raptor Hide Cloak",
		Sources: []*proto.UIItemSource{foreverRare(17, "Takk the Leaper")},
	},
	{
		Id: 285330, Name: "Signet of the Zhevra",
		Sources: []*proto.UIItemSource{foreverRare(17, "Swiftmane")},
	},
	// Items wowhead's Forever listing has with no source, placed by ForeverChanges
	// (https://foreverchanges.pro/item/ID, checked 2026-10-06). The rares come from player
	// reports the site has not checked itself.
	{
		Id: 285190, Name: "Wyvern Heart Band",
		Sources: []*proto.UIItemSource{foreverRare(400, "Heartrazor")},
	},
	{
		Id: 284399, Name: "Seared Grove Shoulderpads",
		Sources: []*proto.UIItemSource{foreverRare(406, "Sister Riven")},
	},
	// Brother Ravenoak or Sentinel Amarassan in the north of Stonetalon. The report does not
	// say which.
	{
		Id: 284403, Name: "Shapeshifting Sentinel's Strides",
		Sources: []*proto.UIItemSource{foreverRare(406, "Brother Ravenoak or Sentinel Amarassan")},
	},
	{
		Id: 285344, Name: "Guard Captain's Barrier",
		Sources: []*proto.UIItemSource{foreverRare(17, "Captain Gerogg Hammertoe")},
	},
	{
		Id: 274068, Name: "Thermaplugg Medal of Honor",
		Sources: []*proto.UIItemSource{{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{NpcId: 6229, ZoneId: 721}}}},
	},
	// Khan Jehn is a new Gelkis quest in Desolace. It needs Honored with the Gelkis and opens
	// at level 30. The rewards carry no required level, so without one here the gear search
	// would judge them on item level 43 and leave them out.
	{
		Id: 271801, Name: "Abandoned Ferocity", RequiredLevel: 30,
		Sources: []*proto.UIItemSource{foreverQuest(93128, "Khan Jehn")},
	},
	{
		Id: 271802, Name: "Bludgeon of Betrayed Virtues", RequiredLevel: 30,
		Sources: []*proto.UIItemSource{foreverQuest(93128, "Khan Jehn")},
	},
	// A Forever rare in the Wetlands. It drops Rotheap Innards, which Rethiel the
	// Greenwarden (Alliance) trades for the ring, once per character.
	{
		Id: 282283, Name: "Malignant Root",
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
		Sources:            []*proto.UIItemSource{foreverRare(11, "Nightveiled Rotheap")},
	},

	// Undead paladin class quest (level 22, from level 18, Horde).
	//
	// wowhead has the quest chain and it has the sword, but no link between them, so the
	// sword arrives with no required level and the level 20 gear search would take it on
	// item level 31 and leave it out. The chain opens at the Undead paladin stronghold
	// with Diplomatic Incident (91858) and ends with The Windshaper's Wrath, which is
	// where the wind and Holy proc on the sword comes from.
	//
	// The proc, 49 Holystorm damage on hit, is not in the sim, so it swings as a plain
	// weapon and the real thing hits harder.
	{
		Id: 267369, Name: "Wolfsbane", Icon: "inv_sword_25",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeTwoHand,
		Stats:           stats.Stats{stats.Stamina: 9, stats.Intellect: 10}.ToFloatArray(),
		WeaponDamageMin: 69, WeaponDamageMax: 105, WeaponSpeed: 3.40,
		Ilvl: 31, RequiredLevel: 18, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		ClassAllowlist:     []proto.Class{proto.Class_ClassPaladin},
		Sources:            []*proto.UIItemSource{foreverQuest(96204, "The Windshaper's Wrath")},
	},

	// New in beta build 1.60.1.70009 (https://foreverchanges.pro/beta). These are low level
	// quest and drop greens from the Horde and Night Elf starting zones, and we know no
	// source or faction for them yet. Rotmender's Robes, Gloves and Sash from Ruins of
	// Lordaeron are left out, because nobody has seen their stats yet (the client only
	// holds their armor).
	{
		Id: 286742, Name: "Riptear's Cleaver",
		Icon: "inv_throwingaxe_02", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeTwoHand,
		Stats:           stats.Stats{stats.Strength: 3}.ToFloatArray(),
		WeaponDamageMin: 23, WeaponDamageMax: 36, WeaponSpeed: 3.30,
		Ilvl: 13, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286730, Name: "Pristine Orcish Dagger",
		Icon: "inv_weapon_shortblade_14", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeDagger, HandType: proto.HandType_HandTypeOneHand,
		Stats:           stats.Stats{stats.Agility: 1}.ToFloatArray(),
		WeaponDamageMin: 10, WeaponDamageMax: 20, WeaponSpeed: 2.40,
		Ilvl: 12, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286755, Name: "Helm Splitter",
		Icon: "inv_sword_04", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
		Stats:           stats.Stats{stats.Strength: 1}.ToFloatArray(),
		WeaponDamageMin: 10, WeaponDamageMax: 19, WeaponSpeed: 2.40,
		Ilvl: 11, RequiredLevel: 6, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286744, Name: "Thrice-Stitched Flesh",
		Icon: "inv_misc_cape_07", Type: proto.ItemType_ItemTypeBack,
		Stats: stats.Stats{stats.Armor: 12, stats.Agility: 1}.ToFloatArray(),
		Ilvl:  13, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286752, Name: "Igleggings",
		Icon: "inv_pants_02", Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Armor: 54, stats.Intellect: 2}.ToFloatArray(),
		Ilvl:  12, RequiredLevel: 7, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286731, Name: "Aggor's Refitted Belt",
		Icon: "inv_belt_09", Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Armor: 69, stats.Stamina: 2}.ToFloatArray(),
		Ilvl:  12, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286754, Name: "Ukta's Conduit",
		Icon: "inv_staff_07", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:           stats.Stats{stats.Stamina: 2}.ToFloatArray(),
		WeaponDamageMin: 19, WeaponDamageMax: 36, WeaponSpeed: 3.30,
		Ilvl: 12, RequiredLevel: 7, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286753, Name: "Snarlsnout Shooter",
		Icon: "inv_weapon_rifle_05", Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeGun,
		Stats:           stats.Stats{stats.Strength: 1}.ToFloatArray(),
		WeaponDamageMin: 8, WeaponDamageMax: 16, WeaponSpeed: 2.80,
		Ilvl: 10, RequiredLevel: 5, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286748, Name: "Bristlebark Bow",
		Icon: "inv_weapon_bow_03", Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow,
		Stats:           stats.Stats{stats.Agility: 1}.ToFloatArray(),
		WeaponDamageMin: 8, WeaponDamageMax: 16, WeaponSpeed: 2.80,
		Ilvl: 10, RequiredLevel: 5, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286750, Name: "Wisesight Wand",
		Icon: "inv_wand_01", Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats:           stats.Stats{stats.Intellect: 1}.ToFloatArray(),
		WeaponDamageMin: 8, WeaponDamageMax: 16, WeaponSpeed: 1.70,
		Ilvl: 10, RequiredLevel: 5, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286743, Name: "Riptear's Spare Arm",
		Icon: "spell_shadow_fingerofdeath", Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats:           stats.Stats{}.ToFloatArray(),
		WeaponDamageMin: 11, WeaponDamageMax: 21, WeaponSpeed: 1.70,
		Ilvl: 13, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286751, Name: "Ghostfang's Steps",
		Icon: "inv_boots_05", Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Armor: 13, stats.Spirit: 1}.ToFloatArray(),
		Ilvl:  10, RequiredLevel: 5, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286745, Name: "Furlfeather Ring",
		Icon: "inv_jewelry_ring_12", Type: proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.HealingPower: 1}.ToFloatArray(),
		Ilvl:  12, RequiredLevel: 7, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286746, Name: "Shal'ma's Shawl",
		Icon: "inv_helmet_48", Type: proto.ItemType_ItemTypeBack,
		Stats: stats.Stats{stats.Armor: 8}.ToFloatArray(),
		Ilvl:  9, RequiredLevel: 4, Quality: proto.ItemQuality_ItemQualityCommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286749, Name: "Wrathroot",
		Icon: "inv_staff_02", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:           stats.Stats{}.ToFloatArray(),
		WeaponDamageMin: 10, WeaponDamageMax: 19, WeaponSpeed: 3.30,
		Ilvl: 8, RequiredLevel: 3, Quality: proto.ItemQuality_ItemQualityCommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286741, Name: "Centaur Skull Basher",
		Icon: "inv_hammer_09", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeTwoHand,
		Stats:           stats.Stats{}.ToFloatArray(),
		WeaponDamageMin: 13, WeaponDamageMax: 20, WeaponSpeed: 3.30,
		Ilvl: 9, Quality: proto.ItemQuality_ItemQualityCommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286733, Name: "Fightin' Fish",
		Icon: "inv_misc_fish_02", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeOneHand,
		Stats:           stats.Stats{}.ToFloatArray(),
		WeaponDamageMin: 4, WeaponDamageMax: 9, WeaponSpeed: 2.40,
		Ilvl: 7, Quality: proto.ItemQuality_ItemQualityCommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286728, Name: "Kolkar Hammer",
		Icon: "inv_hammer_16", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeOneHand,
		Stats:           stats.Stats{}.ToFloatArray(),
		WeaponDamageMin: 5, WeaponDamageMax: 11, WeaponSpeed: 2.40,
		Ilvl: 8, Quality: proto.ItemQuality_ItemQualityCommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286729, Name: "Kolkar Bow",
		Icon: "inv_weapon_bow_12", Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow,
		Stats:           stats.Stats{}.ToFloatArray(),
		WeaponDamageMin: 5, WeaponDamageMax: 10, WeaponSpeed: 2.80,
		Ilvl: 8, Quality: proto.ItemQuality_ItemQualityCommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286735, Name: "Sentinel's Slasher",
		Icon: "inv_sword_23", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
		Stats:           stats.Stats{}.ToFloatArray(),
		WeaponDamageMin: 6, WeaponDamageMax: 12, WeaponSpeed: 2.40,
		Ilvl: 9, Quality: proto.ItemQuality_ItemQualityCommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286740, Name: "Proud Brave's Guard",
		Icon: "inv_shield_09", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeShield, HandType: proto.HandType_HandTypeOffHand,
		Stats: stats.Stats{stats.Armor: 135, stats.BlockValue: 3}.ToFloatArray(),
		Ilvl:  9, Quality: proto.ItemQuality_ItemQualityCommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286737, Name: "Avala's Binding",
		Icon: "inv_jewelry_ring_07", Type: proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.FrostResistance: 1}.ToFloatArray(),
		Ilvl:  8, Quality: proto.ItemQuality_ItemQualityCommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	// Quest rewards that wowhead's Forever listing has with no source. The item pages show the
	// quest (checked 2026-10-04), so we give them the source here and the gear search keeps
	// them. The Alliance rewards found the same way are left out.
	//
	// Dreamer's Chestguard is not in the listing at all, so it goes in with its stats from the
	// tooltip. Baron Aquanis is the Horde quest in Blackfathom Deeps (quest level 30).
	{
		Id: 270043, Name: "Dreamer's Chestguard", Icon: "inv_chest_leather_01",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Strength: 10, stats.Stamina: 5, stats.Spirit: 11, stats.Armor: 106}.ToFloatArray(),
		Ilvl:  32, Phase: 1, Quality: proto.ItemQuality_ItemQualityRare, Expansion: proto.Expansion_ExpansionVanilla,
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(6922, "Baron Aquanis")},
	},
	{
		Id: 271719, Name: "Furs of the Earthen Ring",
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(98823, "Earthen Echo")},
	},
	{
		Id: 271732, Name: "Dirt-Heavy Bracers",
		Sources: []*proto.UIItemSource{foreverQuest(95682, "Open the Maw")},
	},
	{
		Id: 271740, Name: "Knife-Polishing Rag",
		Sources: []*proto.UIItemSource{foreverQuest(95682, "Open the Maw")},
	},
	{
		Id: 271766, Name: "Heavehammer",
		Sources: []*proto.UIItemSource{foreverQuest(98823, "Earthen Echo"), foreverQuest(98824, "Prehistoric Prism")},
	},
	{
		Id: 271767, Name: "Healer's Staff",
		Sources: []*proto.UIItemSource{foreverQuest(98823, "Earthen Echo"), foreverQuest(98824, "Prehistoric Prism")},
	},
	{
		Id: 276727, Name: "Glistening Stompers",
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(515, "Elixir of Agony")},
	},
	{
		Id: 276895, Name: "Transformative Cocoon",
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(96261, "The Broodmother")},
	},
	{
		Id: 277206, Name: "Brilliant Cloak",
		Sources: []*proto.UIItemSource{foreverQuest(95534, "Strahnbrad Mystery")},
	},
	{
		Id: 277214, Name: "Crusty Cuffs",
		Sources: []*proto.UIItemSource{foreverQuest(97048, "Returning to Port")},
	},
	{
		Id: 277224, Name: "Kurzen Headshrinker's Cinch",
		Sources: []*proto.UIItemSource{foreverQuest(97048, "Returning to Port")},
	},
	{
		Id: 277232, Name: "Jailer's Discarded Chain",
		Sources: []*proto.UIItemSource{foreverQuest(97048, "Returning to Port")},
	},
	{
		Id: 277254, Name: "Truthseeker's Bow",
		Sources: []*proto.UIItemSource{foreverQuest(82208, "Greater Friend of the Library")},
	},
	{
		Id: 277514, Name: "Aka'rai's Tattered Vestments",
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(97265, "Hate the Hatefury")},
	},
	{
		Id: 277515, Name: "Dagger of Deals",
		FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
		Sources:            []*proto.UIItemSource{foreverQuest(97265, "Hate the Hatefury")},
	},
	// Hidden items the beta server revealed from October 10 to 13
	// (https://foreverchanges.pro/hidden-items), with the stats from the item pages and the
	// wowhead Forever tooltips, which agree. No source is known yet, so the gear search
	// leaves them out until one is.
	{
		Id: 284064, Name: "Cliffbreaker's Pauldrons", Icon: "inv_shoulder_24",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats: stats.Stats{stats.Armor: 410, stats.Strength: 12, stats.Stamina: 12, stats.SpellPower: 13}.ToFloatArray(),
		Ilvl:  52, RequiredLevel: 47, Quality: proto.ItemQuality_ItemQualityRare, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 284285, Name: "Feardred's Gauntlets", Icon: "inv_gauntlets_29",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Armor: 183, stats.Strength: 11, stats.SpellPower: 13}.ToFloatArray(),
		Ilvl:  54, RequiredLevel: 49, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 281661, Name: "Well-Worn Pitchfork", Icon: "inv_misc_2h_farmpitchfork_a_01",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypePolearm, HandType: proto.HandType_HandTypeTwoHand,
		Stats:           stats.Stats{stats.Strength: 14, stats.Stamina: 3}.ToFloatArray(),
		WeaponDamageMin: 77, WeaponDamageMax: 116, WeaponSpeed: 3.60,
		Ilvl: 38, RequiredLevel: 33, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 283545, Name: "Moonchaser's Headdress", Icon: "inv_helmet_17",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Armor: 101, stats.Intellect: 13, stats.SpellPower: 15}.ToFloatArray(),
		Ilvl:  46, RequiredLevel: 41, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 281896, Name: "Surveyor's Gloves", Icon: "inv_gauntlets_27",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Armor: 75, stats.Stamina: 8, stats.SpellPower: 12}.ToFloatArray(),
		Ilvl:  44, RequiredLevel: 39, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286569, Name: "Firebird Leggings", Icon: "inv_pants_cloth_14",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Armor: 55, stats.Intellect: 14, stats.FirePower: 17}.ToFloatArray(),
		Ilvl:  48, RequiredLevel: 43, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286640, Name: "High Priestess's Kilt", Icon: "inv_pants_cloth_06",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Armor: 73, stats.Stamina: 23, stats.Intellect: 17, stats.SpellPenetration: 5}.ToFloatArray(),
		Ilvl:  59, RequiredLevel: 54, Quality: proto.ItemQuality_ItemQualityRare, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	// Revealed October 12 to 13. Bind on equip, no source known yet.
	{
		Id: 284065, Name: "Crown of Maldazzar", Icon: "inv_helmet_12",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Armor: 61, stats.Stamina: 17, stats.ShadowPower: 24}.ToFloatArray(),
		Ilvl:  58, RequiredLevel: 53, Quality: proto.ItemQuality_ItemQualityUncommon, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	// Revealed October 15 to 16, a batch of bind on equip rares for levels 26 to 40 (Kirin Tor
	// and city guard names, so probably a new Forever dungeon, but no source is known yet).
	// Stats from the ForeverChanges item pages; the wowhead Forever tooltips agree where they
	// exist (none yet for Ancient Cleaver and Arcane Remnant Leggings). The thrown weapons'
	// damage is from wowhead, ForeverChanges shows it as "set by hand". Hit and crit are
	// typed as rating (0.5% = 5 hit or 7 crit rating, as wowhead prints them).
	{
		Id: 286983, Name: "Ancient Cleaver", Icon: "inv_weapon_halbard_01",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeTwoHand,
		Stats:           stats.Stats{stats.Strength: 12, stats.SpellPower: 14}.ToFloatArray(),
		WeaponDamageMin: 80, WeaponDamageMax: 120, WeaponSpeed: 3.60,
		Ilvl: 33, RequiredLevel: 28, Quality: proto.ItemQuality_ItemQualityRare, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286996, Name: "City Guard's Tarnished Blade", Icon: "inv_sword_23",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
		Stats:           stats.Stats{stats.Stamina: 6, stats.SpellPower: 5}.ToFloatArray(),
		WeaponDamageMin: 39, WeaponDamageMax: 73, WeaponSpeed: 2.50,
		Ilvl: 34, RequiredLevel: 29, Quality: proto.ItemQuality_ItemQualityRare, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 287068, Name: "Excavated Dagger", Icon: "inv_weapon_shortblade_19",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeThrown,
		Stats:           stats.Stats{stats.Strength: 4, stats.Agility: 4}.ToFloatArray(),
		WeaponDamageMin: 34, WeaponDamageMax: 35, WeaponSpeed: 1.80,
		Ilvl: 33, RequiredLevel: 28, Quality: proto.ItemQuality_ItemQualityRare, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 287070, Name: "Skypierce Javelin", Icon: "inv_spear_05",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeThrown,
		Stats:           stats.Stats{stats.Strength: 3}.ToFloatArray(),
		CritRating:      7,
		WeaponDamageMin: 79, WeaponDamageMax: 80, WeaponSpeed: 2.90,
		Ilvl: 45, RequiredLevel: 40, Quality: proto.ItemQuality_ItemQualityRare, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286984, Name: "Mark of the Marsh", Icon: "inv_jewelry_necklace_09",
		Type:  proto.ItemType_ItemTypeNeck,
		Stats: stats.Stats{stats.Strength: 5, stats.Agility: 8}.ToFloatArray(),
		Ilvl:  33, RequiredLevel: 28, Quality: proto.ItemQuality_ItemQualityRare, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286995, Name: "Guardian's Emblem", Icon: "spell_arcane_arcaneresilience",
		Type:  proto.ItemType_ItemTypeNeck,
		Stats: stats.Stats{stats.Stamina: 8, stats.SpellPower: 6}.ToFloatArray(),
		Ilvl:  33, RequiredLevel: 28, Quality: proto.ItemQuality_ItemQualityRare, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286994, Name: "Tattered Kirin Tor Cloak", Icon: "inv_misc_cape_05",
		Type:      proto.ItemType_ItemTypeBack,
		Stats:     stats.Stats{stats.Armor: 25, stats.Intellect: 8}.ToFloatArray(),
		HitRating: 5,
		Ilvl:      33, RequiredLevel: 28, Quality: proto.ItemQuality_ItemQualityRare, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286992, Name: "Construct's Grips", Icon: "inv_gauntlets_04",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Armor: 142, stats.Strength: 12, stats.Stamina: 4}.ToFloatArray(),
		Ilvl:  33, RequiredLevel: 28, Quality: proto.ItemQuality_ItemQualityRare, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286990, Name: "Arcane Remnant Leggings", Icon: "inv_pants_leather_05",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Armor: 95, stats.Intellect: 12, stats.ArcanePower: 17}.ToFloatArray(),
		Ilvl:  33, RequiredLevel: 28, Quality: proto.ItemQuality_ItemQualityRare, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286993, Name: "Suffused Bindings", Icon: "inv_belt_15",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Armor: 126, stats.Stamina: 9, stats.Intellect: 4, stats.SpellPower: 6}.ToFloatArray(),
		Ilvl:  32, RequiredLevel: 27, Quality: proto.ItemQuality_ItemQualityRare, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286985, Name: "Thicket Bindings", Icon: "inv_bracer_12",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Armor: 46, stats.Stamina: 8}.ToFloatArray(),
		HitRating: 5,
		Ilvl:      31, RequiredLevel: 26, Quality: proto.ItemQuality_ItemQualityRare, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
	{
		Id: 286986, Name: "Abandoned Leggings", Icon: "inv_pants_09",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Armor: 42, stats.Stamina: 8, stats.Spirit: 6, stats.SpellPenetration: 4}.ToFloatArray(),
		Ilvl:  31, RequiredLevel: 26, Quality: proto.ItemQuality_ItemQualityRare, Phase: 1, Expansion: proto.Expansion_ExpansionVanilla,
	},
}
