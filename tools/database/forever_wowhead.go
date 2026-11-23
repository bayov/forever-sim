package database

import (
	"encoding/json"
	"fmt"
	"log"
	"math"

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
// percent at the character's level. The sim keeps the other ratings in percent, so those
// are turned into percent here (foreverWowheadPercentRatings).
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

// jsonequip stat keys and the sim stat they feed. The spell power keys aren't here,
// spellStats reads them.
var foreverWowheadStats = map[string]proto.Stat{
	"str": proto.Stat_StatStrength, "agi": proto.Stat_StatAgility, "sta": proto.Stat_StatStamina,
	"int": proto.Stat_StatIntellect, "spi": proto.Stat_StatSpirit,
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

// spellStats turns Wowhead's three spell power keys into the sim's stats.
//
// The game has three spell stats: "damage and healing" (Mug of Muddled Memories), damage
// only (Braided Seer Stones) and healing only. An item with "healing done by up to 94 and
// damage done by up to 31" (Staff of the Deeproar) has the last two, and it heals with 94,
// not 94 + 31. Wowhead's numbers add up to the tooltip as splpwr + spldmg for damage and
// splpwr + splheal for healing. Staff of the Deeproar is spldmg 31 and splheal 94, Thrash's
// Trash is splpwr 5 and spldmg 9 for "healing by up to 5 and damage by up to 14", and Naga
// Battle Gloves are splpwr 5 and splheal 10 for "healing by up to 15 and damage by up to 5".
// The tooltips are the ones BayovCore's item sweep recorded on the beta.
//
// The sim deals damage with SpellPower plus SpellDamage and heals with SpellPower plus
// HealingPower (sim/core/spell_result.go). We put what damage and healing share on
// SpellPower and only the rest on the other two, the way the Classic items already are
// (Naga Battle Gloves). That matters for healers, who have no weight for SpellDamage or
// HealingPower: Barkskin Scepter's "damage and healing by up to 50" as SpellDamage 50 plus
// HealingPower 50 would heal right in the sim but score nothing in a healer's gear list.
func (wi ForeverWowheadItem) spellStats() (spellPower, spellDamage, healingPower float64) {
	damage := wi.num("splpwr") + wi.num("spldmg")
	healing := wi.num("splpwr") + wi.num("splheal")
	spellPower = min(damage, healing)
	return spellPower, damage - spellPower, healing - spellPower
}

// Ratings the sim keeps in percent, with how much rating is 1%.
//
// Every tooltip BayovCore recorded in game agrees on these (2026-11-23, 102 items with
// dodge, 26 with block, 17 with haste), and so does Wowhead's rating on the items both
// have. Haste is "attack speed and casting speed", so it goes into both.
var foreverWowheadPercentRatings = map[string]struct {
	stats []proto.Stat
	per   float64
}{
	"exprtng":   {[]proto.Stat{proto.Stat_StatExpertise}, 10},
	"parryrtng": {[]proto.Stat{proto.Stat_StatParry}, 15},
	"dodgertng": {[]proto.Stat{proto.Stat_StatDodge}, 12},
	"blockrtng": {[]proto.Stat{proto.Stat_StatBlock}, 5},
	"hastertng": {[]proto.Stat{proto.Stat_StatMeleeHaste, proto.Stat_StatSpellHaste}, 10},
}

// Rating stats the sim cannot use. Counted so the run reports how much it dropped.
var foreverWowheadRatings = []string{"armorpenrtng"}

// Weapon skill on the Forever items that had it in Classic, from wowhead's Forever
// tooltips (2026-10-09).
//
// Forever made weapon skill a plain item stat with much smaller amounts (Huge Thorium
// Battleaxe went from +10 to +2), and moved some of it to the expertise-like stat. The
// client's ItemSparse (1.60.1.70291) confirms which items have it: stat type 90 is
// two-handed axes, 91 two-handed maces, 96 daggers, and Dwarven Tree Chopper only has
// expertise (type 37). But jsonequip carries none of it, so without this table an item
// keeps its Classic skill. An entry with no skills takes the skill off.
//
// Items that aren't in Wowhead's listing but have a line in forever_seen_in_game.json
// (hidden items revealed later) take their entry there too (ReadForeverSeenInGame).
var foreverWeaponSkills = map[int32]map[proto.WeaponSkill]float64{
	2907:  {},                                               // Dwarven Tree Chopper, Classic +2 two-handed axes.
	4548:  {proto.WeaponSkill_WeaponSkillTwoHandedMaces: 1}, // Servomechanic Sledgehammer, Classic +7.
	12062: {proto.WeaponSkill_WeaponSkillDaggers: 1},        // Skilled Fighting Blade, Classic +4.
	12775: {proto.WeaponSkill_WeaponSkillTwoHandedAxes: 2},  // Huge Thorium Battleaxe, Classic +10.
	16007: {},                                               // Flawless Arcanite Rifle, Classic +4 guns.
	21126: {proto.WeaponSkill_WeaponSkillDaggers: 3},        // Death's Sting, the same as Classic.
	23577: {proto.WeaponSkill_WeaponSkillSwords: 6},         // The Hungering Cold, the same as Classic.
	// Edgemaster's Handguards, Classic +7 each. Revealed Nov 7-8 (foreverchanges.pro
	// hidden-items tooltip, Wowhead Forever tooltip dataEnv 16 agrees).
	14551:  {proto.WeaponSkill_WeaponSkillAxes: 1, proto.WeaponSkill_WeaponSkillDaggers: 1, proto.WeaponSkill_WeaponSkillSwords: 1},
	286619: {proto.WeaponSkill_WeaponSkillUnarmed: 1}, // Big Basher's Belt, +1 fist weapons. Revealed Nov 7-8.
	280068: {proto.WeaponSkill_WeaponSkillUnarmed: 1}, // Bloody Claws of Attack, +1 fist weapons. Revealed Nov 8-9 (Wowhead Forever tooltip).
	// From BayovCore's item sweep on client 1.60.1.70938 (2026-11-13), which reads the
	// skill line ("Equip: Increased Daggers +1."). Where the client's hotfix cache has the
	// item, its ItemSparse skill stat agrees (forever_client_cache.json): 91 two-handed
	// maces, 93 axes, 94 bows, 96 daggers, 101 polearms.
	9375:   {proto.WeaponSkill_WeaponSkillAxes: 1},            // Expert Goldminer's Helmet, Classic +7.
	18392:  {proto.WeaponSkill_WeaponSkillDaggers: 1},         // Distracting Dagger, Classic +6.
	18505:  {proto.WeaponSkill_WeaponSkillDaggers: 1},         // Mugger's Belt, Classic +5.
	15860:  {proto.WeaponSkill_WeaponSkillDaggers: 1},         // Blinkstrike Armguards, none in Classic.
	2915:   {proto.WeaponSkill_WeaponSkillTwoHandedMaces: 1},  // Taran Icebreaker, none in Classic.
	13056:  {proto.WeaponSkill_WeaponSkillPolearms: 1},        // Frenzied Striker, none in Classic.
	280026: {proto.WeaponSkill_WeaponSkillBows: 1},            // Albinofur Grips.
	274929: {proto.WeaponSkill_WeaponSkillTwoHandedAxes: 1},   // Hurley's Felling Axe.
	275669: {proto.WeaponSkill_WeaponSkillUnarmed: 1},         // Punchy's Punchers, +1 fist weapons.
	276102: {proto.WeaponSkill_WeaponSkillTwoHandedSwords: 1}, // Bladewind.
	280016: {proto.WeaponSkill_WeaponSkillFeralCombat: 1},     // Spiderchitin Wrap.
	286616: {proto.WeaponSkill_WeaponSkillFeralCombat: 1},     // Matriarch's Prowlers.
	// Classic's ranged skill is gone. The sweep reads each one's other Equip line (ranged
	// hit 1%), and Sash of the Grand Hunt's full tooltip from a loot (1.60.1.70888) agrees.
	18296: {}, // Marksman Bands, Classic +2 bows, crossbows and guns.
	22207: {}, // Sash of the Grand Hunt, Classic +2 bows, crossbows and guns.
	18388: {}, // Stoneshatter, Classic +4 crossbows.
}

// Equip stats that Wowhead's jsonequip and forever_seen_in_game.json don't carry: "+2 Weapon
// Damage" (the item's BonusPhysicalDamage) and "+7 to All Resistances" (each of the five
// resistances). From BayovCore's full item sweep on client 1.60.1.71037 (2026-11-21), which
// reads them from GetItemStats as ITEM_MOD_PHYSICAL_DAMAGE_DONE_SHORT and
// ITEM_MOD_SPELL_RESISTANCE_ALL_SCHOOLS_SHORT. Where the client's hotfix cache has the item, its
// ItemSparse lists the same stats as types 83 and 124 (forever_client_cache.json), and
// foreverchanges.pro's tooltip of Sharpened Cutlery has its "+1 Weapon Damage".
//
// ToProto and ReadForeverSeenInGame both apply it, since either one's stats can be the last.
// A seen entry with its own bonusPhysicalDamage or allResistances wins over this table.
type foreverEquipStat struct {
	weaponDamage   float64
	allResistances float64
}

var foreverEquipStats = map[int32]foreverEquipStat{
	13255:  {weaponDamage: 1},    // Trueaim Gauntlets.
	14551:  {weaponDamage: 1},    // Edgemaster's Handguards.
	21189:  {weaponDamage: 4},    // Might of Cenarius.
	226830: {weaponDamage: 1},    // Darkmantle Bracers.
	249470: {weaponDamage: 2},    // Molten Heart of the Mountain.
	274092: {weaponDamage: 1},    // Sharpened Cutlery.
	274614: {weaponDamage: 2},    // Knucklebone Ring.
	280012: {weaponDamage: 2},    // Talon of Tor'Rel.
	12065:  {allResistances: 7},  // Ward of the Elements.
	12252:  {allResistances: 6},  // Staff of Protection.
	12609:  {allResistances: 20}, // Polychromatic Visionwrap.
	20504:  {allResistances: 5},  // Lightforged Blade.
	22195:  {allResistances: 5},  // Light Obsidian Belt.
	22197:  {allResistances: 5},  // Heavy Obsidian Belt.
	22198:  {allResistances: 5},  // Jagged Obsidian Shield.
	277559: {allResistances: 7},  // Shen'dralar Relic.
	280009: {allResistances: 4},  // Titanosaur Scale Pendant.
}

// The sim's resistance stats, which "to All Resistances" adds to.
var allResistanceStats = []proto.Stat{
	proto.Stat_StatArcaneResistance, proto.Stat_StatFireResistance, proto.Stat_StatFrostResistance,
	proto.Stat_StatNatureResistance, proto.Stat_StatShadowResistance,
}

// apply puts the equip stats on an item whose stats array doesn't have them yet.
func (e foreverEquipStat) apply(item *proto.UIItem) {
	if e.weaponDamage != 0 {
		item.BonusPhysicalDamage = e.weaponDamage
	}
	for _, stat := range allResistanceStats {
		item.Stats[stat] += e.allResistances
	}
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
	statsArr[proto.Stat_StatSpellPower], statsArr[proto.Stat_StatSpellDamage], statsArr[proto.Stat_StatHealingPower] = wi.spellStats()
	// "+12 Attack Power" is ranged attack power too, as the Classic tooltips read it.
	statsArr[proto.Stat_StatRangedAttackPower] += wi.num("atkpwr")
	for key, rating := range foreverWowheadPercentRatings {
		for _, stat := range rating.stats {
			// Four digits, as in forever_seen_in_game.json.
			statsArr[stat] += math.Round(wi.num(key)/rating.per*1e4) / 1e4
		}
	}
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
	foreverEquipStats[wi.ID].apply(item)
	if itemType == proto.ItemType_ItemTypeWeapon || itemType == proto.ItemType_ItemTypeRanged {
		item.WeaponDamageMin = wi.num("dmgmin1")
		item.WeaponDamageMax = wi.num("dmgmax1")
		item.WeaponSpeed = wi.num("speed")
	}
	// A wand's school, for the wands neither ItemSparse has (AttachWandSchools). Wowhead's
	// dmgtype1 matches ItemSparse's DamageType on every wand both have.
	if item.RangedWeaponType == proto.RangedWeaponType_RangedWeaponTypeWand {
		item.SpellSchool = DamageTypeSchool(int(wi.num("dmgtype1")))
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
