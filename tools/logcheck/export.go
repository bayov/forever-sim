package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// export is the character JSON from the user's /bayov export (BayovCore addon).
//
// It holds the gear, talents, skills and the sheet, but no buffs on purpose. The sheet is read
// with whatever buffs were up, so we only use it as a check.
type export struct {
	When struct {
		Build   int    `json:"build"`
		Date    string `json:"date"`
		Version string `json:"version"`
	} `json:"when"`
	Who struct {
		Name      string `json:"name"`
		Realm     string `json:"realm"`
		ClassFile string `json:"classFile"`
		RaceFile  string `json:"raceFile"`
		Level     int32  `json:"level"`
	} `json:"who"`
	// Gear is a list in slot order. The first exports (before 2026-10-11) had an object keyed
	// by slot number instead, which loadExport turns into the list.
	RawGear json.RawMessage `json:"gear"`
	Gear    []exportItem    `json:"-"`
	Imbues  map[string]struct {
		Enchant int32 `json:"enchant"`
	} `json:"imbues"`
	Skills []struct {
		Name string `json:"name"`
		Rank int    `json:"rank"`
		Max  int    `json:"max"`
	} `json:"skills"`
	Talents struct {
		Talents []struct {
			Name    string `json:"name"`
			SpellID int32  `json:"spellID"`
			Rank    int    `json:"rank"`
		} `json:"talents"`
	} `json:"talents"`
	Stats map[string]json.RawMessage `json:"stats"`

	path string
	time time.Time
	// history is true for an export rebuilt from a BayovCore history record, whose time is
	// within a few seconds of the change it records.
	history bool
}

type exportItem struct {
	ID      int32  `json:"id"`
	Enchant int32  `json:"enchant"`
	Suffix  int32  `json:"suffix"`
	Name    string `json:"name"`
}

func loadExport(path string) (*export, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseExport(data, path)
}

// parseExport reads an export's JSON. `path` says where it came from, for the report.
func parseExport(data []byte, path string) (*export, error) {
	ex := &export{path: path}
	var err error
	if err := json.Unmarshal(data, ex); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if err := json.Unmarshal(ex.RawGear, &ex.Gear); err != nil {
		var bySlot map[string]exportItem
		if err := json.Unmarshal(ex.RawGear, &bySlot); err != nil {
			return nil, fmt.Errorf("%s: gear: %w", path, err)
		}
		slots := make([]string, 0, len(bySlot))
		for slot := range bySlot {
			slots = append(slots, slot)
		}
		slices.SortFunc(slots, func(a, b string) int { return len(a)*100 + strings.Compare(a, b) - len(b)*100 })
		for _, slot := range slots {
			ex.Gear = append(ex.Gear, bySlot[slot])
		}
	}
	ex.time, err = time.Parse(stampFormat, ex.When.Date)
	if err != nil {
		return nil, fmt.Errorf("%s: when.date: %w", path, err)
	}
	return ex, nil
}

// loadExports reads every export we have: the pasted ones in <archive>/exports, and one for each
// record of BayovCore's history in <archive>/history.
func loadExports(archive string) ([]*export, error) {
	paths, _ := filepath.Glob(filepath.Join(archive, "exports", "export-*.json"))
	var all []*export
	for _, path := range paths {
		ex, err := loadExport(path)
		if err != nil {
			return nil, err
		}
		all = append(all, ex)
	}
	history, err := loadHistory(filepath.Join(archive, "history"))
	if err != nil {
		return nil, err
	}
	return append(all, history...), nil
}

// pickExport is the newest export taken before the window ends, or else the first one after
// it.
func pickExport(all []*export, end time.Time) (*export, error) {
	var before, after *export
	for _, ex := range all {
		if !ex.time.After(end) {
			if before == nil || !ex.time.Before(before.time) {
				before = ex
			}
		} else if after == nil || ex.time.Before(after.time) {
			after = ex
		}
	}
	if before != nil {
		return before, nil
	}
	if after != nil {
		return after, nil
	}
	return nil, fmt.Errorf("no exports and no history")
}

// exportsAround returns the character's exports from the last one taken at or before the
// window starts to the first one taken at or after it ends, in time order. Another character
// with the same name (a deleted and remade one) has another class or race, and is left out.
func exportsAround(all []*export, ex *export, from, to time.Time) []*export {
	var mine []*export
	for _, other := range all {
		if other.Who.Name == ex.Who.Name && other.Who.Realm == ex.Who.Realm && other.Who.ClassFile == ex.Who.ClassFile && other.Who.RaceFile == ex.Who.RaceFile {
			mine = append(mine, other)
		}
	}
	slices.SortStableFunc(mine, func(a, b *export) int { return a.time.Compare(b.time) })
	first, last := 0, len(mine)-1
	for i, other := range mine {
		if !other.time.After(from) {
			first = i
		}
	}
	for i := len(mine) - 1; i >= 0; i-- {
		if !mine[i].time.Before(to) {
			last = i
		}
	}
	if last < first {
		return nil
	}
	return mine[first : last+1]
}

// skillChange is a weapon skill going up between two exports.
type skillChange struct {
	from, to, max int
}

// exportChanges lists what changed from one export to the next that moves the sim: the
// character's level, the gear and the talents. The skill ups of a weapon either one holds come
// apart, by skill name, because history has a record for each point.
func exportChanges(a, b *export) ([]string, map[string]skillChange) {
	var changes []string
	skillUps := map[string]skillChange{}
	if a.Who.Level != b.Who.Level {
		changes = append(changes, fmt.Sprintf("level %d to %d", a.Who.Level, b.Who.Level))
	}

	weaponSkills := map[string]bool{}
	for _, ex := range []*export{a, b} {
		for _, item := range ex.Gear {
			if known, ok := core.ItemsByID[item.ID]; ok && weaponSkillName(known) != "" {
				weaponSkills[weaponSkillName(known)] = true
			}
		}
	}
	ranks := map[string]int{}
	for _, skill := range a.Skills {
		ranks[skill.Name] = skill.Rank
	}
	for _, skill := range b.Skills {
		if weaponSkills[skill.Name] && ranks[skill.Name] != skill.Rank {
			skillUps[skill.Name] = skillChange{ranks[skill.Name], skill.Rank, skill.Max}
		}
	}

	itemName := func(item exportItem) string {
		switch {
		case item.ID == 0:
			return "nothing"
		case item.Name != "":
			return item.Name
		}
		return fmt.Sprint(item.ID)
	}
	for i := range max(len(a.Gear), len(b.Gear)) {
		var before, after exportItem
		if i < len(a.Gear) {
			before = a.Gear[i]
		}
		if i < len(b.Gear) {
			after = b.Gear[i]
		}
		switch {
		case before.ID != after.ID:
			changes = append(changes, fmt.Sprintf("%s to %s", itemName(before), itemName(after)))
		case before.Enchant != after.Enchant:
			changes = append(changes, fmt.Sprintf("enchant on %s %d to %d", itemName(after), before.Enchant, after.Enchant))
		}
	}

	talents := func(ex *export) map[string]int {
		ranks := map[string]int{}
		for _, talent := range ex.Talents.Talents {
			ranks[talent.Name] = talent.Rank
		}
		return ranks
	}
	before, after := talents(a), talents(b)
	var talentChanges []string
	for name := range after {
		if before[name] != after[name] {
			talentChanges = append(talentChanges, fmt.Sprintf("%s %d to %d", name, before[name], after[name]))
		}
	}
	for name := range before {
		if _, ok := after[name]; !ok && before[name] > 0 {
			talentChanges = append(talentChanges, fmt.Sprintf("%s %d to 0", name, before[name]))
		}
	}
	slices.Sort(talentChanges)
	if len(talentChanges) > 0 {
		changes = append(changes, "talents: "+strings.Join(talentChanges, ", "))
	}
	return changes, skillUps
}

// character is what we build from an export, with notes on anything we had to leave out.
type character struct {
	player *proto.Player
	notes  []string
}

// buildPlayer turns the export into a sim player with no buffs, no consumes and no rotation.
func buildPlayer(ex *export, tables *clientTables) (*character, error) {
	c := &character{}
	player := &proto.Player{
		Name:      ex.Who.Name,
		Level:     ex.Who.Level,
		Consumes:  &proto.Consumes{},
		Rotation:  &proto.APLRotation{},
		Equipment: &proto.EquipmentSpec{Items: make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)},
	}
	for i := range player.Equipment.Items {
		player.Equipment.Items[i] = &proto.ItemSpec{}
	}
	c.player = player

	race := ex.Who.RaceFile
	if race == "Scourge" {
		race = "Undead"
	}
	raceValue, ok := proto.Race_value["Race"+race]
	if !ok {
		return nil, fmt.Errorf("race %q has no sim race", ex.Who.RaceFile)
	}
	player.Race = proto.Race(raceValue)

	if err := setSpec(player, ex.Who.ClassFile); err != nil {
		return nil, err
	}
	c.equip(ex, tables)
	if err := c.setTalents(ex); err != nil {
		return nil, err
	}
	c.setProfessions(ex)
	return c, nil
}

// setSpec picks the sim spec for the class. We only know the class from the export, so each
// class gets the spec we sim it as.
func setSpec(player *proto.Player, classFile string) error {
	switch classFile {
	case "SHAMAN":
		player.Class = proto.Class_ClassShaman
		player.Spec = &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
			Options: &proto.EnhancementShaman_Options{SyncType: proto.ShamanSyncType_Auto},
		}}
	case "ROGUE":
		player.Class = proto.Class_ClassRogue
		player.Spec = &proto.Player_Rogue{Rogue: &proto.Rogue{Options: &proto.RogueOptions{}}}
	case "PALADIN":
		player.Class = proto.Class_ClassPaladin
		player.Spec = &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
			Options: &proto.PaladinOptions{},
		}}
	default:
		return fmt.Errorf("logcheck doesn't know which spec to sim a %s as yet", classFile)
	}
	return nil
}

// equip puts the export's gear in the sim's slots.
//
// The export lists gear in inventory slot order but leaves out empty slots, so we place each
// item by its type. The first ring and trinket go in slot 1, the second in slot 2. An off-hand
// item goes in the off hand, and any other weapon in the main hand first.
func (c *character) equip(ex *export, tables *clientTables) {
	items := c.player.Equipment.Items
	var left []string
	for _, gear := range ex.Gear {
		// An export rebuilt from history has an empty entry for each empty slot.
		if gear.ID == 0 {
			continue
		}
		item, ok := core.ItemsByID[gear.ID]
		if !ok {
			left = append(left, fmt.Sprintf("%s (%d)", gear.Name, gear.ID))
			continue
		}
		slot, ok := itemSlot(item, items)
		if !ok {
			left = append(left, fmt.Sprintf("%s (%d, no free slot)", gear.Name, gear.ID))
			continue
		}
		spec := &proto.ItemSpec{Id: gear.ID, RandomSuffix: gear.Suffix}
		if gear.Enchant != 0 {
			if enchant, ok := findEnchant(gear.Enchant, item, tables); ok {
				spec.Enchant = enchant
			} else {
				c.addClientEnchant(gear.Enchant, gear.Name, tables)
			}
		}
		items[slot] = spec
	}
	if len(left) > 0 {
		c.notes = append(c.notes, "not in the sim's database, left off: "+strings.Join(left, ", "))
	}
}

// The client's stat and resistance numbers for an enchant's effects, as sim stats. The stats
// are the client's ItemModType, and the resistances its school order, where 0 is armor.
var (
	enchantStats = map[int32]stats.Stat{
		0: stats.Mana, 1: stats.Health, 3: stats.Agility, 4: stats.Strength, 5: stats.Intellect,
		6: stats.Spirit, 7: stats.Stamina, 38: stats.AttackPower, 41: stats.HealingPower,
		42: stats.SpellDamage, 43: stats.MP5, 45: stats.SpellPower,
	}
	enchantResistances = map[int32]stats.Stat{
		0: stats.BonusArmor, 2: stats.FireResistance, 3: stats.NatureResistance,
		4: stats.FrostResistance, 5: stats.ShadowResistance, 6: stats.ArcaneResistance,
	}
)

// addClientEnchant adds an enchant the sim's database lacks as bonus stats, from the client's
// SpellItemEnchantment row. Forever's new leg and boot enchants (8482, 8485) are like that. An
// enchant that does more than add stats (a proc, weapon damage) is left off with a note.
func (c *character) addClientEnchant(id int32, itemName string, tables *clientTables) {
	enchant, ok := tables.enchants[id]
	if !ok {
		c.notes = append(c.notes, fmt.Sprintf("enchant %d on %s is in neither the sim's database nor the client's, left off", id, itemName))
		return
	}
	var added stats.Stats
	var parts []string
	for _, effect := range enchant.effects {
		var stat stats.Stat
		switch effect.kind {
		case 4:
			stat, ok = enchantResistances[effect.arg]
		case 5:
			stat, ok = enchantStats[effect.arg]
		default:
			ok = false
		}
		if !ok {
			c.notes = append(c.notes, fmt.Sprintf("enchant %d on %s (%s) does more than add stats, left off", id, itemName, enchant.name))
			return
		}
		added[stat] += float64(effect.amount)
		parts = append(parts, fmt.Sprintf("+%d %s", effect.amount, stat.StatName()))
	}
	if c.player.BonusStats == nil {
		c.player.BonusStats = &proto.UnitStats{Stats: make([]float64, stats.Len)}
	}
	for i, v := range added {
		c.player.BonusStats.Stats[i] += v
	}
	c.notes = append(c.notes, fmt.Sprintf("enchant %d on %s isn't in the sim's database, added from the client data as %s", id, itemName, strings.Join(parts, ", ")))
}

func itemSlot(item core.Item, items []*proto.ItemSpec) (proto.ItemSlot, bool) {
	free := func(slots ...proto.ItemSlot) (proto.ItemSlot, bool) {
		for _, slot := range slots {
			if items[slot].Id == 0 {
				return slot, true
			}
		}
		return 0, false
	}
	switch item.Type {
	case proto.ItemType_ItemTypeFinger:
		return free(proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2)
	case proto.ItemType_ItemTypeTrinket:
		return free(proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2)
	case proto.ItemType_ItemTypeWeapon:
		if item.HandType == proto.HandType_HandTypeOffHand {
			return free(proto.ItemSlot_ItemSlotOffHand)
		}
		return free(proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand)
	case proto.ItemType_ItemTypeRanged:
		return free(proto.ItemSlot_ItemSlotRanged)
	}
	slots := map[proto.ItemType]proto.ItemSlot{
		proto.ItemType_ItemTypeHead:     proto.ItemSlot_ItemSlotHead,
		proto.ItemType_ItemTypeNeck:     proto.ItemSlot_ItemSlotNeck,
		proto.ItemType_ItemTypeShoulder: proto.ItemSlot_ItemSlotShoulder,
		proto.ItemType_ItemTypeBack:     proto.ItemSlot_ItemSlotBack,
		proto.ItemType_ItemTypeChest:    proto.ItemSlot_ItemSlotChest,
		proto.ItemType_ItemTypeWrist:    proto.ItemSlot_ItemSlotWrist,
		proto.ItemType_ItemTypeHands:    proto.ItemSlot_ItemSlotHands,
		proto.ItemType_ItemTypeWaist:    proto.ItemSlot_ItemSlotWaist,
		proto.ItemType_ItemTypeLegs:     proto.ItemSlot_ItemSlotLegs,
		proto.ItemType_ItemTypeFeet:     proto.ItemSlot_ItemSlotFeet,
	}
	if slot, ok := slots[item.Type]; ok {
		return free(slot)
	}
	return 0, false
}

// findEnchant finds the sim's enchant for an enchant ID from an item link.
//
// The sim's database keys most enchants by that ID, but the newer entries by the spell that
// applies the enchant (Enchant Cloak - Minor Agility is 13419 there, and 247 in the link). So
// we try the ID first, then each spell the client says applies it, and keep one that fits the
// item.
func findEnchant(id int32, item core.Item, tables *clientTables) (int32, bool) {
	fits := func(enchant dbEnchant) bool {
		return enchant.Type == item.Type || slices.Contains(enchant.ExtraTypes, item.Type)
	}
	for _, enchant := range dbEnchants() {
		if enchant.EffectID == id && fits(enchant) {
			return id, true
		}
	}
	for _, spell := range tables.enchantSpells[id] {
		for _, enchant := range dbEnchants() {
			if enchant.SpellID == spell && fits(enchant) {
				return enchant.EffectID, true
			}
		}
	}
	return 0, false
}

// dbEnchant is an enchant in assets/database/db.json. The sim keeps only the effect ID and
// stats, so we read the item types and the spell from the file.
type dbEnchant struct {
	EffectID   int32            `json:"effectId"`
	SpellID    int32            `json:"spellId"`
	Type       proto.ItemType   `json:"type"`
	ExtraTypes []proto.ItemType `json:"extraTypes"`
}

var enchantsCache []dbEnchant

func dbEnchants() []dbEnchant {
	if enchantsCache != nil {
		return enchantsCache
	}
	var db struct {
		Enchants []dbEnchant `json:"enchants"`
	}
	data, err := os.ReadFile(filepath.Join(repoRoot, "assets/database/db.json"))
	if err == nil {
		err = json.Unmarshal(data, &db)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "reading the enchants:", err)
	}
	enchantsCache = db.Enchants
	return enchantsCache
}

// setTalents builds the talent string from the export's talent names.
//
// The Forever client has no talent tabs to read, so we match each learned talent by name (or
// spell ID) against the sim's tree in ui/core/talents/trees, whose order is the talent string's
// order.
func (c *character) setTalents(ex *export) error {
	class := strings.ToLower(ex.Who.ClassFile)
	data, err := os.ReadFile(filepath.Join(repoRoot, "ui/core/talents/trees", class+".json"))
	if err != nil {
		return err
	}
	var trees []struct {
		Talents []struct {
			Name     string  `json:"name"`
			SpellIDs []int32 `json:"spellIds"`
			Max      int     `json:"maxPoints"`
		} `json:"talents"`
	}
	if err := json.Unmarshal(data, &trees); err != nil {
		return err
	}

	digits := make([][]byte, len(trees))
	index := map[string][2]int{}
	bySpell := map[int32][2]int{}
	for t, tree := range trees {
		digits[t] = []byte(strings.Repeat("0", len(tree.Talents)))
		for i, talent := range tree.Talents {
			index[strings.ToLower(talent.Name)] = [2]int{t, i}
			for _, id := range talent.SpellIDs {
				bySpell[id] = [2]int{t, i}
			}
		}
	}
	points := 0
	var unknown []string
	for _, talent := range ex.Talents.Talents {
		if talent.Rank == 0 {
			continue
		}
		at, ok := index[strings.ToLower(talent.Name)]
		if !ok {
			at, ok = bySpell[talent.SpellID]
		}
		if !ok {
			unknown = append(unknown, fmt.Sprintf("%s %d", talent.Name, talent.Rank))
			continue
		}
		digits[at[0]][at[1]] = byte('0' + talent.Rank)
		points += talent.Rank
	}
	if len(unknown) > 0 {
		c.notes = append(c.notes, "talents the sim's tree doesn't have, left out: "+strings.Join(unknown, ", "))
	}
	parts := make([]string, len(digits))
	for i, d := range digits {
		parts[i] = strings.TrimRight(string(d), "0")
	}
	c.player.TalentsString = strings.Join(parts, "-")
	if granted := int(c.player.Level) - 9; points > granted {
		c.player.BonusTalentPoints = int32(points - granted)
	}
	return nil
}

// setProfessions takes the first two professions the sim knows. Forever lists each profession
// twice, so we skip the second line.
func (c *character) setProfessions(ex *export) {
	var professions []proto.Profession
	for _, skill := range ex.Skills {
		value, ok := proto.Profession_value[skill.Name]
		if !ok || slices.Contains(professions, proto.Profession(value)) {
			continue
		}
		professions = append(professions, proto.Profession(value))
	}
	if len(professions) > 0 {
		c.player.Profession1 = professions[0]
	}
	if len(professions) > 1 {
		c.player.Profession2 = professions[1]
	}
}

// setSkillGaps sets how far the main hand and off hand weapons' skill is below the most for
// our level, from the export's skills.
func setSkillGaps(player *proto.Player, ex *export) {
	gap := func(slot proto.ItemSlot) int32 {
		item, ok := core.ItemsByID[player.Equipment.Items[slot].Id]
		if !ok {
			return 0
		}
		name := weaponSkillName(item)
		for _, skill := range ex.Skills {
			if skill.Name == name && skill.Rank < skill.Max {
				return int32(skill.Max - skill.Rank)
			}
		}
		return 0
	}
	player.MainHandSkillGap = gap(proto.ItemSlot_ItemSlotMainHand)
	player.OffHandSkillGap = gap(proto.ItemSlot_ItemSlotOffHand)
}

func weaponSkillName(item core.Item) string {
	twoHand := item.HandType == proto.HandType_HandTypeTwoHand
	switch item.WeaponType {
	case proto.WeaponType_WeaponTypeAxe:
		if twoHand {
			return "Two-Handed Axes"
		}
		return "Axes"
	case proto.WeaponType_WeaponTypeMace:
		if twoHand {
			return "Two-Handed Maces"
		}
		return "Maces"
	case proto.WeaponType_WeaponTypeSword:
		if twoHand {
			return "Two-Handed Swords"
		}
		return "Swords"
	case proto.WeaponType_WeaponTypeDagger:
		return "Daggers"
	case proto.WeaponType_WeaponTypeFist:
		return "Unarmed"
	case proto.WeaponType_WeaponTypeStaff:
		return "Staves"
	case proto.WeaponType_WeaponTypePolearm:
		return "Polearms"
	}
	return ""
}
