package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// ForeverItems are items that exist only in WoW Forever, entered by hand.
//
// The Classic inputs (wowhead dump, atlasloot, wago) know nothing about them,
// so we type the stats from the wowhead Forever tooltips
// (https://nether.wowhead.com/forever/tooltip/item/ID?dataEnv=17&locale=0) and
// merge them in after the Classic items. Quest names, quest levels and
// factions come from https://wowforevertalents.com/dungeons/ and
// /reputations/, which read the beta client. Chance on hit effects
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
}
