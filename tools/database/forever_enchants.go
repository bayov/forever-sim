package database

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/tools"
)

// Forever's enchants and armor kits up to Enchanting 225, the most a level 20 enchanter
// can learn, from the list https://foreverchanges.pro builds its level 20 BiS enchants
// from.
//
// Past that, the Enchanting 300 enchants and Rugged armor kits that Forever added or
// changed. We list only those with a known source outside raids: the formulas wowhead's
// Forever listing has sold by vendors, and the leatherworking kits. The Grand Crusader /
// Sorcerer / Arcanist / Inquisitor, Off-Hand Superior Intellect and Excellent Spirit, and
// Lotus Claw formulas have no known source, so they're left out. Bracer Superior Agility
// and 2H Strength have no known source either, but they're in the database for the melee
// specs to compare. Their launch availability is unconfirmed, so no preset uses them.
//
// This list only says which enchants we offer, and their slots and level rules. What
// each one gives comes from the client: tools/sources/enchants.py reads it out of
// SpellItemEnchantment into assets/db_inputs/forever_enchants.json, and gen_db fills in
// the stats with ApplyClientEnchantStats. So a client build that changes an enchant
// changes the sim at the next gen_db run.
//
// Forever buffed a lot of the low Classic enchants. Cloak and bracer Minor Agility went
// from +1 to +3, Gloves Agility from +5 to +7, and 2H Minor Impact from +2 to +4 damage.
// It also added necklace enchants, +15 Strength and Agility for two-handers, and
// Forceful (Attack Power) and Mystic (Spell Power) armor kits. Classic shares one effect
// ID between enchants with equal stats, e.g. 247 is +1 Agility on a cloak, bracer and
// boots. Now those enchants differ, so every enchant here that is new or changed gets its
// own effect ID, which is its spell ID. ReplaceForeverEnchants drops the Classic entry of
// the same spell, so the enchant picker shows each enchant once.
//
// The armor kits keep the Classic kits' IDs and only gain the Stamina that Forever gave
// them. Heavy kits need an item of level 15 or above and Thick kits level 25 or above.
// The kits also need a character level (hyjal.cc recipe pages): the plain Medium,
// Heavy and Thick kits 5, 20 and 30, the Forceful and Mystic ones 15, 30 and 40. So a
// level 20 gets the Forceful or Mystic Medium kit or the plain Heavy kit at best. The
// sim does not check either rule, but the UI enchant picker and the rotopt enchant
// search do.
var ForeverEnchants = []*proto.UIEnchant{
	// Armor kits.
	foreverKit(15, 2304, 2831, 1, "Light Armor Kit"),
	foreverKit(16, 2313, 2832, 5, "Medium Armor Kit"),
	foreverKit(17, 4265, 2833, 20, "Heavy Armor Kit"),
	foreverKit(18, 8173, 10344, 30, "Thick Armor Kit"),
	foreverKit(1255123, 0, 1255123, 15, "Forceful Medium Armor Kit"),
	foreverKit(1255124, 0, 1255124, 15, "Mystic Medium Armor Kit"),
	foreverKit(1255095, 0, 1255095, 30, "Forceful Heavy Armor Kit"),
	foreverKit(1255096, 0, 1255096, 30, "Mystic Heavy Armor Kit"),
	foreverKit(1255073, 0, 1255073, 40, "Forceful Thick Armor Kit"),
	foreverKit(1255074, 0, 1255074, 40, "Mystic Thick Armor Kit"),
	// Rugged kits need level 50 for Forceful and Mystic, 40 for the plain one, and an item of
	// level 35 or above.
	foreverKit(1843, 15564, 19057, 40, "Rugged Armor Kit"),
	foreverKit(1254965, 0, 1254965, 50, "Forceful Rugged Armor Kit"),
	foreverKit(1254966, 0, 1254966, 50, "Mystic Rugged Armor Kit"),

	// Necklace, new in Forever (Enchanting 210, formulas for Merchant's Favor).
	foreverEnchant(1249059, "Enchant Necklace - Agility", proto.ItemType_ItemTypeNeck),
	foreverEnchant(1249019, "Enchant Necklace - Strength", proto.ItemType_ItemTypeNeck),
	foreverEnchant(1249057, "Enchant Necklace - Spell Power", proto.ItemType_ItemTypeNeck),
	foreverEnchant(1249058, "Enchant Necklace - Healing Power", proto.ItemType_ItemTypeNeck),

	// Cloak.
	foreverEnchant(13419, "Enchant Cloak - Minor Agility", proto.ItemType_ItemTypeBack),
	foreverEnchant(1219587, "Enchant Cloak - Agility", proto.ItemType_ItemTypeBack),

	// Chest. Superior Stamina (Enchanting 220) was Superior Health (+50 health) in Classic.
	foreverEnchant(13858, "Enchant Chest - Superior Stamina", proto.ItemType_ItemTypeChest),

	// Bracer.
	foreverEnchant(7779, "Enchant Bracer - Minor Agility", proto.ItemType_ItemTypeWrist),
	foreverEnchant(7782, "Enchant Bracer - Minor Strength", proto.ItemType_ItemTypeWrist),
	foreverEnchant(1248458, "Enchant Bracer - Minor Intellect", proto.ItemType_ItemTypeWrist),
	foreverEnchant(1248459, "Enchant Bracer - Minor Healing Power", proto.ItemType_ItemTypeWrist),
	foreverEnchant(1248460, "Enchant Bracer - Lesser Agility", proto.ItemType_ItemTypeWrist),
	foreverEnchant(13536, "Enchant Bracer - Lesser Strength", proto.ItemType_ItemTypeWrist),
	// Client 1.60.1.70425 points 13622 at enchant 8734, +3 (enchant 723, +5, through 70338).
	foreverEnchant(13622, "Enchant Bracer - Lesser Intellect", proto.ItemType_ItemTypeWrist),
	foreverEnchant(13661, "Enchant Bracer - Strength", proto.ItemType_ItemTypeWrist),
	foreverEnchant(1248497, "Enchant Bracer - Agility", proto.ItemType_ItemTypeWrist),
	foreverEnchant(1248498, "Enchant Bracer - Lesser Healing Power", proto.ItemType_ItemTypeWrist),
	foreverEnchant(1248661, "Enchant Bracer - Superior Intellect", proto.ItemType_ItemTypeWrist),
	foreverEnchant(1220624, "Enchant Bracer - Greater Spellpower", proto.ItemType_ItemTypeWrist),
	// Client enchant 7656. Launch availability unconfirmed, so no preset uses it.
	foreverEnchant(1248599, "Enchant Bracer - Superior Agility", proto.ItemType_ItemTypeWrist),

	// Gloves.
	foreverEnchant(13815, "Enchant Gloves - Agility", proto.ItemType_ItemTypeHands),
	foreverEnchant(13887, "Enchant Gloves - Strength", proto.ItemType_ItemTypeHands),
	foreverEnchant(1248640, "Enchant Gloves - Superior Strength", proto.ItemType_ItemTypeHands),
	foreverEnchant(1213622, "Enchant Gloves - Holy Power", proto.ItemType_ItemTypeHands),
	foreverEnchant(1213626, "Enchant Gloves - Arcane Power", proto.ItemType_ItemTypeHands),
	foreverEnchant(1248645, "Enchant Gloves - Natural Power", proto.ItemType_ItemTypeHands),

	// Boots.
	foreverEnchant(7867, "Enchant Boots - Minor Agility", proto.ItemType_ItemTypeFeet),
	foreverEnchant(13637, "Enchant Boots - Lesser Agility", proto.ItemType_ItemTypeFeet),

	// Weapon. Winter's Might went from +7 to +15 Frost spell damage.
	foreverEnchant(21931, "Enchant Weapon - Winter's Might", proto.ItemType_ItemTypeWeapon),

	// Enchanting 140. The proc is a weapon effect in sim/common.
	foreverEnchant(1248805, "Enchant Weapon - Revelation", proto.ItemType_ItemTypeWeapon),

	// Off-hand, for held in off hand items only (not shields).
	foreverOffHand(1219579, "Enchant Off-Hand - Wisdom"),

	// Two-hand weapon. The Impact enchants are weapon effects in sim/common.
	foreverTwoHand(1219580, "Enchant 2H Weapon - Spellblasting"),
	foreverTwoHand(1248607, "Enchant 2H Weapon - Mighty Spell Power"),
	// Client enchant 8215. Launch availability unconfirmed, so no preset uses it.
	foreverTwoHand(1248668, "Enchant 2H Weapon - Strength"),
	foreverTwoHand(7793, "Enchant 2H Weapon - Lesser Intellect"),
	foreverTwoHand(7745, "Enchant 2H Weapon - Minor Impact"),
	foreverTwoHand(13529, "Enchant 2H Weapon - Lesser Impact"),
	foreverTwoHand(13695, "Enchant 2H Weapon - Impact"),
	// These two are in the client data but not in the beta (user, 2026-10-05), so no preset or
	// search uses them. Lesser Agility's enchant is named "Agility +15" in client 1.60.1.70450,
	// but the spell it casts while worn is still Classic's +9, so the sim gives +9.
	foreverTwoHand(1248510, "Enchant 2H Weapon - Lesser Agility"),
	foreverTwoHand(1248511, "Enchant 2H Weapon - Lesser Strength"),
}

func foreverEnchant(spellID int32, name string, itemType proto.ItemType) *proto.UIEnchant {
	return &proto.UIEnchant{EffectId: spellID, SpellId: spellID, Name: name, Quality: proto.ItemQuality_ItemQualityCommon, Type: itemType}
}

func foreverTwoHand(spellID int32, name string) *proto.UIEnchant {
	enchant := foreverEnchant(spellID, name, proto.ItemType_ItemTypeWeapon)
	enchant.EnchantType = proto.EnchantType_EnchantTypeTwoHand
	return enchant
}

func foreverOffHand(spellID int32, name string) *proto.UIEnchant {
	enchant := foreverEnchant(spellID, name, proto.ItemType_ItemTypeWeapon)
	enchant.EnchantType = proto.EnchantType_EnchantTypeOffHand
	return enchant
}

func foreverKit(effectID int32, itemID int32, spellID int32, level int32, name string) *proto.UIEnchant {
	return &proto.UIEnchant{
		EffectId: effectID, ItemId: itemID, SpellId: spellID, Name: name, Quality: proto.ItemQuality_ItemQualityCommon,
		RequiredLevel: level,
		Type:          proto.ItemType_ItemTypeChest,
		ExtraTypes:    []proto.ItemType{proto.ItemType_ItemTypeLegs, proto.ItemType_ItemTypeHands, proto.ItemType_ItemTypeFeet},
		EnchantType:   proto.EnchantType_EnchantTypeKit,
	}
}

// clientEnchant is one entry of assets/db_inputs/forever_enchants.json.
type clientEnchant struct {
	Enchant int32              `json:"enchant"`
	Name    string             `json:"name"`
	Stats   map[string]float64 `json:"stats"`
}

// ApplyClientEnchantStats gives each enchant the stats the client's tables say it has,
// from forever_enchants.json (see tools/sources/enchants.py).
//
// It stops gen_db when an enchant in the list isn't in the file, or gives a stat the sim
// has no name for. Either one means the list and the client disagree, and a silent zero
// would put a worthless enchant in the picker.
func ApplyClientEnchantStats(arr []*proto.UIEnchant, path string) {
	var file struct {
		Enchants map[string]clientEnchant `json:"enchants"`
	}
	if err := json.Unmarshal([]byte(tools.ReadFile(path)), &file); err != nil {
		log.Fatalf("Cannot parse %s: %v", path, err)
	}
	byName := map[string]stats.Stat{}
	for s := stats.Stat(0); s < stats.Len; s++ {
		byName[s.StatName()] = s
	}
	for _, enchant := range arr {
		entry, ok := file.Enchants[strconv.Itoa(int(enchant.SpellId))]
		if !ok {
			log.Fatalf("%s has no enchant for spell %d (%s). Run tools/sources/enchants.py for the current client build.", path, enchant.SpellId, enchant.Name)
		}
		var s stats.Stats
		for name, value := range entry.Stats {
			stat, ok := byName[name]
			if !ok {
				log.Fatalf("%s: spell %d (%s) gives %s, which the sim has no stat for", path, enchant.SpellId, enchant.Name, name)
			}
			s[stat] = value
		}
		enchant.Stats = s.ToFloatArray()
	}
}

// ReplaceForeverEnchants puts the Forever enchants in, in place of the Classic entries
// for the same spell.
func (db *WowDatabase) ReplaceForeverEnchants(arr []*proto.UIEnchant) {
	for _, enchant := range arr {
		for key, old := range db.Enchants {
			if old.SpellId == enchant.SpellId && key != EnchantToDBKey(enchant) {
				delete(db.Enchants, key)
			}
		}
		db.MergeEnchant(enchant)
	}
}
