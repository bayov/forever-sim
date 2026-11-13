package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/classic/sim/core/proto"
	goproto "google.golang.org/protobuf/proto"
)

// Enchants of a BiS list (docs/bis_lists.md). Once the gear search has settled on a set, we
// try every enchant the database has for each slot of it, the way the item ranking tries
// items. The best one becomes the slot's enchant, and the rest are listed best first with
// what they lose against it. Then the items are ranked on the set with those enchants, so
// the list's DPS is the set's with its best enchants.
//
// The enchants come from the whole database, not from enchants.go's level 20 list. A slot
// takes what fits its item: the enchant's item type, a two hand enchant on a two hander
// only, a kit's item and character level, the class and profession an enchant may need.
// -gear-phase also holds for enchants, so a Zul'Gurub or Naxxramas one never comes up.
// What's left out on top of that is in the -gear-rank-enchants file, a hand check of what
// the game has open, keyed by the enchant's spell ID: "out" leaves an enchant out, and
// "avail" and "note" go on its entry the way phase0_content.json's do on items.
//
// Each slot holds one enchant, so things that share it can't stack: Iron Counterweight
// (a blacksmith's attachment) and Crusader are two options of the same weapon slot.
//
// Enchants are a few DPS apart, and haste or crit pair badly from one run to the next (two
// runs with 1% haste between them drift apart after the first swing), so the ranking runs
// at -gear-rank-enchant-iters, more than the items' -iters. A first try on 2026-11-14 at
// 1000 iterations had Arcanum of Rapidity 4 DPS behind on the head and 6 ahead on the legs,
// where single runs at 3000 put it level on both.
//
// An enchant that scores exactly what no enchant scores does nothing in the sim (Stamina,
// resistances, Minor Speed), so it isn't listed. The runs share their seed, so any stat
// the fight uses moves the DPS a little.

var (
	gearRankEnchants     = flag.String("gear-rank-enchants", "", "gear search: with -gear-rank-out, also rank every enchant of each slot on the set, and use the best. The value is the hand check of what's open, keyed by spell ID (tools/rotopt/phase0_enchants.json)")
	gearRankEnchantIters = flag.Int("gear-rank-enchant-iters", 0, "gear search: the iterations of the -gear-rank-enchants ranking, 0 for -iters. Enchants are a few DPS apart, so they want more than the items")
	gearRankEnchantKeep  = flag.Int("gear-rank-enchant-keep", 10, "gear search: how many enchants each slot of the -gear-rank-out list keeps, the pick included")
	clientEnchantsPath   = flag.String("client-enchants", "assets/db_inputs/forever_enchants.json", "the Forever client's enchants by spell ID, for the game's enchant IDs in the BiS list")
)

// Every enchant in the database and the highest phase the gear search takes, set by
// loadGearPool.
var (
	allEnchants     []*proto.UIEnchant
	enchantMaxPhase int32
)

// One line of a slot's enchant list, as the list file has it.
type bisEnchant struct {
	// ID is the enchant's effect ID in the sim, what an ItemSpec's enchant holds.
	ID int32 `json:"id"`
	// Game is the enchant ID the game has on an item link, when it isn't ID. Forever's own
	// enchants have their spell ID as effect ID in the sim (tools/database/forever_enchants.go).
	Game int32 `json:"game,omitempty"`
	// Spell is the enchant's spell, the one its formula teaches, and Item the item that
	// applies it (an arcanum, a kit, Iron Counterweight).
	Spell  int32   `json:"spell,omitempty"`
	Item   int32   `json:"item,omitempty"`
	Name   string  `json:"name"`
	Effect string  `json:"effect,omitempty"`
	DPS    float64 `json:"dps"`
	Loss   float64 `json:"loss"`
	BiS    bool    `json:"bis,omitempty"`
	Avail  string  `json:"avail,omitempty"`
	Note   string  `json:"note,omitempty"`
}

type bisEnchantSlot struct {
	name    string
	entries []bisEnchant
}

// The hand check of an enchant (-gear-rank-enchants).
type enchantCheck struct {
	Name  string `json:"name"`
	Avail string `json:"avail"`
	Note  string `json:"note"`
}

func loadEnchantChecks(path string) (map[int32]enchantCheck, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw map[string]enchantCheck
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	out := map[int32]enchantCheck{}
	for k, v := range raw {
		id, err := strconv.Atoi(k)
		if err != nil {
			return nil, fmt.Errorf("%s: key %q is not a spell ID", path, k)
		}
		out[int32(id)] = v
	}
	return out, nil
}

// The game's enchant ID of each enchant spell, from the client's data. A spell it doesn't
// have keeps the sim's effect ID.
func loadClientEnchantIDs(path string) map[int32]int32 {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("  no client enchants (%v), the list's game IDs are the sim's\n", err)
		return nil
	}
	var file struct {
		Enchants map[string]struct {
			Enchant int32 `json:"enchant"`
		} `json:"enchants"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		fmt.Printf("  %s: %v, the list's game IDs are the sim's\n", path, err)
		return nil
	}
	out := map[int32]int32{}
	for k, v := range file.Enchants {
		if id, err := strconv.Atoi(k); err == nil && v.Enchant != 0 {
			out[int32(id)] = v.Enchant
		}
	}
	return out
}

// enchantFitsItem is whether the enchant goes on the item for this character.
func enchantFitsItem(e *proto.UIEnchant, item *proto.UIItem, player *proto.Player, level int32) bool {
	if len(e.ClassAllowlist) > 0 && !slices.Contains(e.ClassAllowlist, player.Class) {
		return false
	}
	if e.RequiredProfession != proto.Profession_ProfessionUnknown && e.RequiredProfession != player.Profession1 && e.RequiredProfession != player.Profession2 {
		return false
	}
	if enchantMaxPhase > 0 && e.Phase > enchantMaxPhase {
		return false
	}
	if e.Type != item.Type && !slices.Contains(e.ExtraTypes, item.Type) {
		return false
	}
	switch item.Type {
	case proto.ItemType_ItemTypeWeapon:
		switch e.EnchantType {
		case proto.EnchantType_EnchantTypeShield:
			return item.WeaponType == proto.WeaponType_WeaponTypeShield
		case proto.EnchantType_EnchantTypeOffHand:
			return item.WeaponType == proto.WeaponType_WeaponTypeOffHand
		case proto.EnchantType_EnchantTypeTwoHand:
			return item.HandType == proto.HandType_HandTypeTwoHand
		case proto.EnchantType_EnchantTypeStaff:
			return item.WeaponType == proto.WeaponType_WeaponTypeStaff
		default:
			return item.WeaponType != proto.WeaponType_WeaponTypeShield && item.WeaponType != proto.WeaponType_WeaponTypeOffHand
		}
	case proto.ItemType_ItemTypeRanged:
		// Scopes, on a bow, crossbow or gun. A relic takes none.
		return slices.Contains([]proto.RangedWeaponType{proto.RangedWeaponType_RangedWeaponTypeBow, proto.RangedWeaponType_RangedWeaponTypeCrossbow, proto.RangedWeaponType_RangedWeaponTypeGun}, item.RangedWeaponType)
	}
	return e.EnchantType != proto.EnchantType_EnchantTypeKit || kitFits(item, e.EffectId, level)
}

// The enchants the item in the slot can take, one per effect ID, without the ones checks
// marks out.
func (s *searcher) bisEnchantOptions(item *proto.UIItem, checks map[int32]enchantCheck) []*proto.UIEnchant {
	var out []*proto.UIEnchant
	for _, e := range allEnchants {
		if checks[e.SpellId].Avail == "out" || !enchantFitsItem(e, item, s.setup.player(), s.level) {
			continue
		}
		if slices.ContainsFunc(out, func(o *proto.UIEnchant) bool { return o.EffectId == e.EffectId }) {
			continue
		}
		out = append(out, e)
	}
	return out
}

// What an enchant gives, short: "+8 Strength", "+4 all stats", "+10 AP, +40 armor". Empty
// for an enchant with a proc or a weapon effect, which its name says.
func enchantEffect(e *proto.UIEnchant) string {
	short := map[proto.Stat]string{
		proto.Stat_StatAttackPower: "AP", proto.Stat_StatRangedAttackPower: "", proto.Stat_StatSpellPower: "spell power",
		proto.Stat_StatSpellDamage: "spell damage", proto.Stat_StatHealingPower: "healing", proto.Stat_StatBonusArmor: "armor",
		proto.Stat_StatMP5: "mp5", proto.Stat_StatNaturePower: "Nature damage", proto.Stat_StatMeleeHaste: "% attack speed",
		proto.Stat_StatMana: "mana", proto.Stat_StatHealth: "health",
	}
	five := []proto.Stat{proto.Stat_StatStrength, proto.Stat_StatAgility, proto.Stat_StatStamina, proto.Stat_StatIntellect, proto.Stat_StatSpirit}
	all := len(e.Stats) > int(proto.Stat_StatSpirit) && e.Stats[proto.Stat_StatStrength] != 0
	for _, st := range five {
		all = all && e.Stats[st] == e.Stats[proto.Stat_StatStrength]
	}
	var parts []string
	if all {
		parts = append(parts, fmt.Sprintf("+%g all stats", e.Stats[proto.Stat_StatStrength]))
	}
	for i, v := range e.Stats {
		st := proto.Stat(i)
		if v == 0 || (all && slices.Contains(five, st)) {
			continue
		}
		name, ok := short[st]
		if !ok {
			name = " " + strings.TrimPrefix(st.String(), "Stat")
		} else if name == "" {
			continue
		} else if !strings.HasPrefix(name, "%") {
			name = " " + name
		}
		parts = append(parts, fmt.Sprintf("+%g%s", v, name))
	}
	return strings.Join(parts, ", ")
}

// rankEnchants puts the best enchant on each slot of best and returns each slot's list,
// best first. Each slot tries every enchant against the rest of the set, and the best one
// stays when it beats the slot's enchant by more than the noise. Then the next slot goes,
// on the set as it now is. We go round until a round changes nothing, so that round's
// lists are all scored on the final set.
func (s *searcher) rankEnchants(best []*proto.ItemSpec, knobs Knobs) ([]bisEnchantSlot, error) {
	checks, err := loadEnchantChecks(*gearRankEnchants)
	if err != nil {
		return nil, err
	}
	gameIDs := loadClientEnchantIDs(*clientEnchantsPath)
	keep := *gearRankEnchantKeep
	if *gearRankEnchantIters > 0 {
		full := s.iterations
		s.iterations = int32(*gearRankEnchantIters)
		defer func() { s.iterations = full }()
	}
	type slotRank struct {
		slot   int
		ranked []rankedGear
	}
	byOption := map[string]*proto.UIEnchant{}
	var last []slotRank
	for round := 1; round <= 5; round++ {
		last = nil
		changed := false
		for slot, spec := range best {
			if spec == nil || itemsByID[spec.Id] == nil {
				continue
			}
			enchants := s.bisEnchantOptions(itemsByID[spec.Id], checks)
			if len(enchants) == 0 {
				continue
			}
			var options []gearOption
			for _, e := range enchants {
				o := goproto.Clone(spec).(*proto.ItemSpec)
				o.Enchant = e.EffectId
				name := fmt.Sprintf("%s (%d)", e.Name, e.EffectId)
				byOption[name] = e
				options = append(options, gearOption{spec: o, name: name})
			}
			with := func(o gearOption) [][]*proto.ItemSpec {
				candidate := slices.Clone(best)
				candidate[slot] = o.spec
				return [][]*proto.ItemSpec{candidate}
			}
			bare := goproto.Clone(spec).(*proto.ItemSpec)
			bare.Enchant = 0
			ranked := s.rankGear(gearSlotNames[slot]+" enchants", options, func(o gearOption) [][]*proto.ItemSpec { return with(o) },
				with(gearOption{spec: bare}), keep+keep/2, knobs)
			last = append(last, slotRank{slot, ranked})
			if len(ranked) == 0 {
				if spec.Enchant != 0 {
					fmt.Printf("  round %d: %s enchant %s does nothing, now none\n", round, gearSlotNames[slot], enchantName(spec.Enchant))
					best[slot] = bare
					changed = true
				}
				continue
			}
			top := ranked[0]
			if top.option.spec.Enchant == spec.Enchant {
				continue
			}
			cur := slices.IndexFunc(ranked, func(r rankedGear) bool { return r.option.spec.Enchant == spec.Enchant })
			if cur >= 0 {
				if gain := top.result.dps - ranked[cur].result.dps; gain <= s.confidence*gainNoise(ranked[cur].result, top.result) {
					continue
				}
			}
			fmt.Printf("  round %d: %s enchant = %s, was %s\n", round, gearSlotNames[slot], byOption[top.option.name].Name, enchantName(spec.Enchant))
			best[slot] = top.option.spec
			changed = true
		}
		if !changed {
			break
		}
		if round == 5 {
			fmt.Println("  the enchants still changed after 5 rounds, the lists are from the last one")
		}
	}

	s.setGear(best)
	base := s.score(knobs)
	round2 := func(x float64) float64 {
		if r := math.Round(x*100) / 100; r != 0 {
			return r
		}
		return 0
	}
	var out []bisEnchantSlot
	for _, sr := range last {
		var entries []bisEnchant
		for _, r := range sr.ranked {
			e := byOption[r.option.name]
			entry := bisEnchant{ID: e.EffectId, Spell: e.SpellId, Item: e.ItemId, Name: e.Name, Effect: enchantEffect(e),
				DPS: round2(r.result.dps), Loss: round2(base.dps - r.result.dps), BiS: e.EffectId == best[sr.slot].Enchant}
			if game, ok := gameIDs[e.SpellId]; ok && game != e.EffectId {
				entry.Game = game
			}
			if c, ok := checks[e.SpellId]; ok {
				if c.Avail != "now" {
					entry.Avail = c.Avail
				}
				entry.Note = c.Note
			} else {
				fmt.Printf("  %s (spell %d) has no hand check in %s\n", e.Name, e.SpellId, *gearRankEnchants)
			}
			entries = append(entries, entry)
		}
		// The pick first, then best first.
		slices.SortStableFunc(entries, func(a, b bisEnchant) int {
			switch {
			case a.BiS && !b.BiS:
				return -1
			case b.BiS && !a.BiS:
				return 1
			}
			return 0
		})
		if len(entries) > keep {
			entries = entries[:keep]
		}
		if len(entries) == 0 {
			continue
		}
		out = append(out, bisEnchantSlot{gearSlotNames[sr.slot], entries})
	}
	fmt.Printf("  enchants: %s\n", base)
	for _, slot := range out {
		fmt.Printf("  %s enchants:\n", slot.name)
		for _, e := range slot.entries {
			fmt.Printf("    %6.1f %6.1f  %s %s  %s %s\n", e.DPS, 0-e.Loss, e.Name, e.Effect, e.Avail, e.Note)
		}
	}
	return out, nil
}
