package database

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The wowhead Forever item listing, scraped by tools/database/forever_wowhead/scrape.py.
//
// Forever reworked a lot of Classic gear (Spirit became Spell Power on low level cloth,
// dungeon weapons hit harder, white items turned green) and added its own, and the
// Classic inputs know none of that. The listing rows carry the item's slot and quality,
// the jsonequip block carries its stats, so together they are enough to build a UIItem.
// Hit and critical strike rating go on the item as they are, and the sim turns them into
// percent at the character's level. Other ratings (expertise, defense, dodge) are dropped
// because the sim has no use for them.
type ForeverWowheadDB struct {
	MaxReqLevel int32                `json:"maxReqLevel"`
	Items       []ForeverWowheadItem `json:"items"`
}

type ForeverWowheadItem struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	Quality  int32  `json:"quality"`
	Level    int32  `json:"level"`
	ReqLevel int32  `json:"reqlevel"`
	// wowhead inventory slot, item class (2 weapon, 4 armor) and subclass.
	Slot     int32 `json:"slot"`
	Class    int32 `json:"classs"`
	Subclass int32 `json:"subclass"`
	ReqClass int32 `json:"reqclass"`
	// 1 Alliance, 2 Horde, 3 both.
	Side int32 `json:"side"`
	// wowhead source kinds (1 crafted, 2 drop, 4 quest, 5 vendor). An item with none is in
	// the client but nowhere in the world yet, as far as wowhead knows.
	Source     []int32                `json:"source"`
	SourceMore []ForeverWowheadSource `json:"sourcemore"`
	// Binds when picked up. Only fetched for crafted items, where it means the crafter
	// is the only one who can wear it.
	BindOnPickup bool `json:"bop"`
	// The classes that can take the quest the item comes from, as a wowhead class mask,
	// and the side that can take it. Only set when a quest is the item's only source.
	// See mark_quest_restrictions.
	QuestClassMask int32 `json:"qclassmask"`
	QuestSide      int32 `json:"qside"`
	// The jsonequip block. Stats are numbers, a few keys (appearances) are objects.
	Eq map[string]json.RawMessage `json:"eq"`
}

// One source of the item. t is the kind: 1 npc drop, 2 object, 3 contained in an item,
// 5 quest reward, 6 crafted (ti is the spell, s the skill line).
type ForeverWowheadSource struct {
	Type     int32  `json:"t"`
	Name     string `json:"n"`
	TargetID int32  `json:"ti"`
	Zone     int32  `json:"z"`
	Skill    int32  `json:"s"`
}

func ParseForeverWowheadDB(contents string) ForeverWowheadDB {
	var db ForeverWowheadDB
	if err := json.Unmarshal([]byte(contents), &db); err != nil {
		log.Fatalf("failed to parse forever wowhead items: %s", err)
	}
	fmt.Printf("\n--\nForever wowhead items loaded: %d (required level up to %d)\n--\n", len(db.Items), db.MaxReqLevel)
	return db
}

// jsonequip stat keys and the sim stat they feed. Both spell power flavours land on
// SpellPower, healing only gear keeps its own stat.
var foreverWowheadStats = map[string]proto.Stat{
	"str": proto.Stat_StatStrength, "agi": proto.Stat_StatAgility, "sta": proto.Stat_StatStamina,
	"int": proto.Stat_StatIntellect, "spi": proto.Stat_StatSpirit,
	"splpwr": proto.Stat_StatSpellPower, "spldmg": proto.Stat_StatSpellPower, "splheal": proto.Stat_StatHealingPower,
	"arcsplpwr": proto.Stat_StatArcanePower, "firsplpwr": proto.Stat_StatFirePower, "frosplpwr": proto.Stat_StatFrostPower,
	"holsplpwr": proto.Stat_StatHolyPower, "natsplpwr": proto.Stat_StatNaturePower, "shasplpwr": proto.Stat_StatShadowPower,
	"manargn": proto.Stat_StatMP5, "splhitpct": proto.Stat_StatSpellHit, "splcritstrkpct": proto.Stat_StatSpellCrit,
	"splhastepct": proto.Stat_StatSpellHaste, "splpen": proto.Stat_StatSpellPenetration,
	"atkpwr": proto.Stat_StatAttackPower, "mleatkpwr": proto.Stat_StatAttackPower, "rgdatkpwr": proto.Stat_StatRangedAttackPower,
	"mlehitpct": proto.Stat_StatMeleeHit, "mlecritstrkpct": proto.Stat_StatMeleeCrit, "mlehastepct": proto.Stat_StatMeleeHaste,
	"armor": proto.Stat_StatArmor, "armorbonus": proto.Stat_StatBonusArmor,
	"def": proto.Stat_StatDefense, "defrtng": proto.Stat_StatDefense,
	"blockpct": proto.Stat_StatBlock, "blockamount": proto.Stat_StatBlockValue,
	"dodgepct": proto.Stat_StatDodge, "parrypct": proto.Stat_StatParry,
	"health": proto.Stat_StatHealth, "mana": proto.Stat_StatMana,
	"arcres": proto.Stat_StatArcaneResistance, "firres": proto.Stat_StatFireResistance, "frores": proto.Stat_StatFrostResistance,
	"natres": proto.Stat_StatNatureResistance, "shares": proto.Stat_StatShadowResistance,
}

// Rating stats the sim cannot use. Counted so the run reports how much it dropped.
var foreverWowheadRatings = []string{"hastertng", "dodgertng", "parryrtng", "blockrtng", "armorpenrtng"}

// Weapon skill on the Forever items that had it in Classic, from wowhead's Forever
// tooltips (2026-10-09).
//
// Forever made weapon skill a plain item stat with much smaller amounts (Huge Thorium
// Battleaxe went from +10 to +2), and moved some of it to the expertise-like stat. The
// client's ItemSparse (1.60.1.70291) confirms which items have it: stat type 90 is
// two-handed axes, 91 two-handed maces, 96 daggers, and Dwarven Tree Chopper only has
// expertise (type 37). But jsonequip carries none of it, so without this table an item
// keeps its Classic skill. An entry with no skills takes the skill off.
var foreverWeaponSkills = map[int32]map[proto.WeaponSkill]float64{
	2907:  {},                                               // Dwarven Tree Chopper, Classic +2 two-handed axes.
	4548:  {proto.WeaponSkill_WeaponSkillTwoHandedMaces: 1}, // Servomechanic Sledgehammer, Classic +7.
	12062: {proto.WeaponSkill_WeaponSkillDaggers: 1},        // Skilled Fighting Blade, Classic +4.
	12775: {proto.WeaponSkill_WeaponSkillTwoHandedAxes: 2},  // Huge Thorium Battleaxe, Classic +10.
	16007: {},                                               // Flawless Arcanite Rifle, Classic +4 guns.
	21126: {proto.WeaponSkill_WeaponSkillDaggers: 3},        // Death's Sting, the same as Classic.
	23577: {proto.WeaponSkill_WeaponSkillSwords: 6},         // The Hungering Cold, the same as Classic.
}

// foreverWeaponSkillSlice turns an entry of foreverWeaponSkills into the item's list of
// skill by weapon type.
func foreverWeaponSkillSlice(skills map[proto.WeaponSkill]float64) []float64 {
	slice := make([]float64, stats.WeaponSkillLen)
	for skill, amount := range skills {
		slice[skill] = amount
	}
	return slice
}

func (wi ForeverWowheadItem) num(key string) float64 {
	v, ok := wi.Eq[key]
	if !ok {
		return 0
	}
	var f float64
	json.Unmarshal(v, &f)
	return f
}

// Whether the item is gear the sim can wear: shirts, tabards, bags and the like are not.
func (wi ForeverWowheadItem) itemType() (proto.ItemType, proto.HandType, bool) {
	switch wi.Slot {
	case 1:
		return proto.ItemType_ItemTypeHead, 0, true
	case 2:
		return proto.ItemType_ItemTypeNeck, 0, true
	case 3:
		return proto.ItemType_ItemTypeShoulder, 0, true
	case 16:
		return proto.ItemType_ItemTypeBack, 0, true
	case 5, 20:
		return proto.ItemType_ItemTypeChest, 0, true
	case 9:
		return proto.ItemType_ItemTypeWrist, 0, true
	case 10:
		return proto.ItemType_ItemTypeHands, 0, true
	case 6:
		return proto.ItemType_ItemTypeWaist, 0, true
	case 7:
		return proto.ItemType_ItemTypeLegs, 0, true
	case 8:
		return proto.ItemType_ItemTypeFeet, 0, true
	case 11:
		return proto.ItemType_ItemTypeFinger, 0, true
	case 12:
		return proto.ItemType_ItemTypeTrinket, 0, true
	case 13:
		return proto.ItemType_ItemTypeWeapon, proto.HandType_HandTypeOneHand, true
	case 17:
		return proto.ItemType_ItemTypeWeapon, proto.HandType_HandTypeTwoHand, true
	case 21:
		return proto.ItemType_ItemTypeWeapon, proto.HandType_HandTypeMainHand, true
	case 14, 22, 23:
		return proto.ItemType_ItemTypeWeapon, proto.HandType_HandTypeOffHand, true
	case 15, 25, 26, 28:
		return proto.ItemType_ItemTypeRanged, 0, true
	}
	return 0, 0, false
}

func (wi ForeverWowheadItem) weaponType() proto.WeaponType {
	if wi.Slot == 14 {
		return proto.WeaponType_WeaponTypeShield
	}
	if wi.Slot == 23 {
		return proto.WeaponType_WeaponTypeOffHand
	}
	if wi.Class != 2 {
		return 0
	}
	switch wi.Subclass {
	case 0, 1:
		return proto.WeaponType_WeaponTypeAxe
	case 4, 5:
		return proto.WeaponType_WeaponTypeMace
	case 6:
		return proto.WeaponType_WeaponTypePolearm
	case 7, 8:
		return proto.WeaponType_WeaponTypeSword
	case 10:
		return proto.WeaponType_WeaponTypeStaff
	case 13:
		return proto.WeaponType_WeaponTypeFist
	case 15:
		return proto.WeaponType_WeaponTypeDagger
	}
	return 0
}

func (wi ForeverWowheadItem) rangedWeaponType() proto.RangedWeaponType {
	if wi.Class == 2 {
		switch wi.Subclass {
		case 2:
			return proto.RangedWeaponType_RangedWeaponTypeBow
		case 3:
			return proto.RangedWeaponType_RangedWeaponTypeGun
		case 16:
			return proto.RangedWeaponType_RangedWeaponTypeThrown
		case 18:
			return proto.RangedWeaponType_RangedWeaponTypeCrossbow
		case 19:
			return proto.RangedWeaponType_RangedWeaponTypeWand
		}
	}
	if wi.Class == 4 {
		switch wi.Subclass {
		case 7:
			return proto.RangedWeaponType_RangedWeaponTypeLibram
		case 8:
			return proto.RangedWeaponType_RangedWeaponTypeIdol
		case 9:
			return proto.RangedWeaponType_RangedWeaponTypeTotem
		}
	}
	return 0
}

func (wi ForeverWowheadItem) armorType() proto.ArmorType {
	if wi.Class != 4 {
		return 0
	}
	switch wi.Subclass {
	case 1:
		return proto.ArmorType_ArmorTypeCloth
	case 2:
		return proto.ArmorType_ArmorTypeLeather
	case 3:
		return proto.ArmorType_ArmorTypeMail
	case 4:
		return proto.ArmorType_ArmorTypePlate
	}
	return 0
}

// Crafting skill lines wowhead names in a crafted source.
var foreverWowheadSkills = map[int32]proto.Profession{
	164: proto.Profession_Blacksmithing, 165: proto.Profession_Leatherworking, 171: proto.Profession_Alchemy,
	197: proto.Profession_Tailoring, 202: proto.Profession_Engineering, 333: proto.Profession_Enchanting,
}

func (wi ForeverWowheadItem) sources() []*proto.UIItemSource {
	var out []*proto.UIItemSource
	for _, s := range wi.SourceMore {
		switch s.Type {
		case 1:
			out = append(out, &proto.UIItemSource{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{NpcId: s.TargetID, ZoneId: s.Zone}}})
		case 5:
			out = append(out, &proto.UIItemSource{Source: &proto.UIItemSource_Quest{Quest: &proto.QuestSource{Id: s.TargetID, Name: s.Name}}})
		case 6:
			out = append(out, &proto.UIItemSource{Source: &proto.UIItemSource_Crafted{Crafted: &proto.CraftedSource{Profession: foreverWowheadSkills[s.Skill], SpellId: s.TargetID}}})
		}
	}
	return out
}

// The character level a profession skill requirement stands for: Apprentice at 5,
// Journeyman at 10, Expert at 20, Artisan at 35, as in Classic. Engineering goggles
// carry "Requires Engineering (270)" and Forever dropped their level requirement to
// 10, but nobody reaches 270 before level 35.
//
// An item with no level requirement at all keeps it that way (the Servomechanic
// Sledgehammer quest reward wants Engineering 100 and nothing else), because the gear
// search treats those by item level instead.
func (wi ForeverWowheadItem) skillLevel() int32 {
	rank := int32(wi.num("reqskillrank"))
	switch {
	case rank == 0:
		return 0
	case rank <= 75:
		return 5
	case rank <= 150:
		return 10
	case rank <= 225:
		return 20
	default:
		return 35
	}
}

// Quests wowhead marks for one class that every class can take in the beta.
//
// wowhead has the Library quests (Friend of the Library for the necklace at 10 books,
// Greater Friend of the Library for the ring at 20) as mage only, so Erudite's Amulet
// and Philanthropist's Ring came out mage gear and the rogue only Field Researcher's
// Loop came out as gear nobody can wear. A level 1 warrior has taken the first quest
// in the beta (https://foreverchanges.pro/library-books), and the mages only get a
// quest of their own on top.
var foreverOpenQuests = map[int32]bool{
	78150: true, // Friend of the Library
	79536: true, // Greater Friend of the Library
}

// Whether every quest that hands the item over is open to every class.
func (wi ForeverWowheadItem) fromOpenQuest() bool {
	quests := 0
	for _, s := range wi.SourceMore {
		if s.Type != 5 {
			continue
		}
		if !foreverOpenQuests[s.TargetID] {
			return false
		}
		quests++
	}
	return quests > 0
}

// ToProto builds the item, or nil when it is not gear. The second value is how many
// rating stats were dropped.
func (wi ForeverWowheadItem) ToProto() (*proto.UIItem, int) {
	itemType, handType, ok := wi.itemType()
	if !ok {
		return nil, 0
	}
	// The item's own class restriction and the one on the quest that hands it over both
	// have to let a class through. A quest a class can take and nobody wears the reward
	// of means the item is not gear at all.
	classMask := wi.ReqClass
	if wi.QuestClassMask != 0 && !wi.fromOpenQuest() {
		if classMask == 0 {
			classMask = wi.QuestClassMask
		} else if classMask &= wi.QuestClassMask; classMask == 0 {
			return nil, 0
		}
	}
	reqLevel := wi.ReqLevel
	if reqLevel > 0 {
		reqLevel = max(reqLevel, wi.skillLevel())
	}
	statsArr := make([]float64, proto.Stat_StatFeralAttackPower+1)
	for key, stat := range foreverWowheadStats {
		statsArr[stat] += wi.num(key)
	}
	// "+12 Attack Power" is ranged attack power too, as the Classic tooltips read it.
	statsArr[proto.Stat_StatRangedAttackPower] += wi.num("atkpwr")
	// Expertise is a flat percent like hit and crit: the tooltip shows 10 rating as "1.0%"
	// less chance to be dodged or parried. The sim's Expertise stat is in percent.
	statsArr[proto.Stat_StatExpertise] += wi.num("exprtng") / 10
	dropped := 0
	for _, key := range foreverWowheadRatings {
		if wi.num(key) != 0 {
			dropped++
		}
	}
	item := &proto.UIItem{
		Id: wi.ID, Name: wi.Name, Icon: wi.Icon,
		Type: itemType, ArmorType: wi.armorType(), WeaponType: wi.weaponType(), HandType: handType, RangedWeaponType: wi.rangedWeaponType(),
		Stats:         statsArr,
		Ilvl:          wi.Level,
		RequiredLevel: reqLevel,
		Phase:         1,
		Quality:       proto.ItemQuality(wi.Quality),
		Expansion:     proto.Expansion_ExpansionVanilla,
		Sources:       wi.sources(),
		HitRating:     wi.num("hitrtng"),
		CritRating:    wi.num("critstrkrtng"),
	}
	if skills, ok := foreverWeaponSkills[wi.ID]; ok {
		item.WeaponSkills = foreverWeaponSkillSlice(skills)
	}
	if itemType == proto.ItemType_ItemTypeWeapon || itemType == proto.ItemType_ItemTypeRanged {
		item.WeaponDamageMin = wi.num("dmgmin1")
		item.WeaponDamageMax = wi.num("dmgmax1")
		item.WeaponSpeed = wi.num("speed")
	}
	// Only a profession the item needs to be worn counts, like "Requires Engineering (215)"
	// on Spellpower Goggles Xtreme (its reqskill). Crafting a bind on pickup item does not
	// count. A player can level the profession, craft the item and then drop the
	// profession, and still wear it (Polished Driftwood Icon, made with Enchanting).
	item.RequiredProfession = foreverWowheadSkills[int32(wi.num("reqskill"))]
	item.BindOnPickup = wi.BindOnPickup
	if classMask != 0 {
		item.ClassAllowlist = WowheadItem{ClassMask: uint16(classMask)}.getClassRestriction()
	}
	side := wi.Side
	if side == 0 {
		side = wi.QuestSide
	}
	switch side {
	case 1:
		item.FactionRestriction = proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY
	case 2:
		item.FactionRestriction = proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY
	}
	return item, dropped
}
