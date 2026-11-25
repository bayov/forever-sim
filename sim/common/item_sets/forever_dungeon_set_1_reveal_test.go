//go:build with_db

package item_sets_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// The bonuses of the Dungeon Set 1 sets Forever revealed in November 2026 (client ItemSetSpell,
// 1.60.1.71037).

func init() {
	sim.RegisterAll()
}

// The slots of a set's first six pieces in the tests: wrist, waist, shoulder, legs, head, hands.
var revealSlots = []proto.ItemSlot{
	proto.ItemSlot_ItemSlotWrist, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotShoulder,
	proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotHands,
}

type revealSet struct {
	id     int32
	name   string
	pieces []int32 // in the order of revealSlots
	player func() *proto.Player
}

func (set revealSet) item(t *testing.T, i int) core.Item {
	t.Helper()
	item, ok := core.ItemsByID[set.pieces[i]]
	if !ok {
		t.Fatalf("no item %d in the database", set.pieces[i])
	}
	if item.SetID != set.id {
		t.Fatalf("item %d has set ID %d, want %d", item.ID, item.SetID, set.id)
	}
	return item
}

// statAgent is a bare Forever character, to read a stat bonus without the stat dependencies a
// class adds.
type statAgent struct {
	core.Character
}

func (a *statAgent) GetCharacter() *core.Character { return &a.Character }
func (a *statAgent) Initialize()                   {}
func (a *statAgent) ApplyTalents()                 {}
func (a *statAgent) Reset(_ *core.Simulation)      {}

// checkStatBonus checks that n pieces count as the set's n-piece bonus and that the bonus adds
// exactly want.
func checkStatBonus(t *testing.T, set revealSet, n int32, want stats.Stats) {
	t.Helper()
	agent := &statAgent{Character: core.Character{Unit: core.Unit{Env: &core.Environment{Ruleset: proto.Ruleset_RulesetForever}}}}
	c := agent.GetCharacter()
	for i := range n {
		c.Equipment[revealSlots[i]] = set.item(t, int(i))
	}
	if names := c.GetActiveSetBonusNames(); !slices.Contains(names, fmt.Sprintf("%s (%dpc)", set.name, n)) {
		t.Fatalf("%d pieces give %v", n, names)
	}
	itemSet := core.ItemSetOf(c.Equipment[revealSlots[0]])
	if itemSet == nil || itemSet.ID != set.id {
		t.Fatalf("item %d is in set %v, want %d", set.pieces[0], itemSet, set.id)
	}
	itemSet.Bonuses[n](agent)
	for s := range stats.Len {
		if c.GetStat(s) != want[s] {
			t.Errorf("%s %d-piece gives %.0f %s, want %.0f", set.name, n, c.GetStat(s), s.StatName(), want[s])
		}
	}
}

// revealSim builds a Forever sim of the set's player wearing its first n pieces.
func revealSim(t *testing.T, set revealSet, n int) *core.Character {
	t.Helper()
	return revealSimWith(t, set, n, nil)
}

// revealSimWith is revealSim with other items in the slots the set's pieces leave free.
func revealSimWith(t *testing.T, set revealSet, n int, others map[proto.ItemSlot]int32) *core.Character {
	t.Helper()
	specs := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range specs {
		specs[i] = &proto.ItemSpec{}
	}
	for i := range n {
		specs[revealSlots[i]].Id = set.item(t, i).ID
	}
	for slot, id := range others {
		specs[slot].Id = id
	}
	player := set.player()
	player.Level = 60
	player.Equipment = &proto.EquipmentSpec{Items: specs}
	player.Rotation = &proto.APLRotation{}
	s := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	s.Reset()
	return s.Raid.Parties[0].Players[0].GetCharacter()
}

// checkProc checks that the 4-piece bonus registers the aura and 3 pieces don't.
func checkProc(t *testing.T, set revealSet, aura string) {
	t.Helper()
	if revealSim(t, set, 3).GetAura(aura) != nil {
		t.Errorf("3 pieces of %s have %s", set.name, aura)
	}
	if revealSim(t, set, 4).GetAura(aura) == nil {
		t.Errorf("4 pieces of %s lack %s", set.name, aura)
	}
}

func sixPieces(first int32) []int32 {
	return []int32{first, first + 1, first + 2, first + 3, first + 4, first + 5}
}

var (
	resist8 = stats.Stats{stats.ArcaneResistance: 8, stats.FireResistance: 8, stats.FrostResistance: 8,
		stats.NatureResistance: 8, stats.ShadowResistance: 8}
	mp5     = stats.Stats{stats.MP5: 8}
	ap40    = stats.Stats{stats.AttackPower: 40, stats.RangedAttackPower: 40}
	sp23    = stats.Stats{stats.SpellPower: 23}
	noStats = stats.Stats{}
)

func TestForeverFeralheartRaiment(t *testing.T) {
	set := revealSet{1667, "Feralheart Raiment", sixPieces(226815), func() *proto.Player {
		return &proto.Player{Race: proto.Race_RaceTauren, Class: proto.Class_ClassDruid,
			Spec: &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{Options: &proto.BalanceDruid_Options{}}}}
	}}
	checkStatBonus(t, set, 2, resist8)
	checkStatBonus(t, set, 3, mp5)
	checkStatBonus(t, set, 6, ap40.Add(sp23))
	checkProc(t, set, "Feralheart Raiment - 4pc (Mana)")
}

func TestForeverBeastmasterArmor(t *testing.T) {
	set := revealSet{1669, "Beastmaster Armor", sixPieces(226897), func() *proto.Player {
		return &proto.Player{Race: proto.Race_RaceOrc, Class: proto.Class_ClassHunter,
			Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{}}}}
	}}
	checkStatBonus(t, set, 2, resist8)
	checkStatBonus(t, set, 3, mp5)
	checkStatBonus(t, set, 6, ap40)
	checkProc(t, set, "Beastmaster Armor - 4pc")
}

func TestForeverSorcerersRegalia(t *testing.T) {
	set := revealSet{1671, "Sorcerer's Regalia", sixPieces(226937), func() *proto.Player {
		return &proto.Player{Race: proto.Race_RaceGnome, Class: proto.Class_ClassMage,
			Spec: &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{}}}}
	}}
	checkStatBonus(t, set, 2, resist8)
	checkStatBonus(t, set, 3, mp5)
	checkStatBonus(t, set, 6, sp23)
	checkProc(t, set, "Sorcerer's Regalia - 4pc")
}

// A tank piece (Soulforge Wristguards 226985) counts with the Retribution pieces.
func TestForeverSoulforgeArmor(t *testing.T) {
	set := revealSet{1673, "Soulforge Armor", append([]int32{226985}, sixPieces(226993)[1:]...), func() *proto.Player {
		return &proto.Player{Race: proto.Race_RaceHuman, Class: proto.Class_ClassPaladin,
			Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
				Options: &proto.PaladinOptions{PrimarySeal: proto.PaladinSeal_Righteousness}}}}
	}}
	checkStatBonus(t, set, 2, resist8)
	checkStatBonus(t, set, 3, mp5)
	checkStatBonus(t, set, 6, sp23)
	checkProc(t, set, "Crusader's Wrath (Melee)")
	checkProc(t, set, "Crusader's Wrath (Spell)")
	checkProc(t, set, "Crusader's Wrath")
}

func TestForeverVestmentsOfTheVirtuous(t *testing.T) {
	set := revealSet{1675, "Vestments of the Virtuous", sixPieces(226961), func() *proto.Player {
		return &proto.Player{Race: proto.Race_RaceHuman, Class: proto.Class_ClassPriest,
			Spec: &proto.Player_ShadowPriest{ShadowPriest: &proto.ShadowPriest{Options: &proto.ShadowPriest_Options{}}}}
	}}
	checkStatBonus(t, set, 2, resist8)
	checkStatBonus(t, set, 3, mp5)
	checkStatBonus(t, set, 6, sp23)
	checkProc(t, set, "Vestments of the Virtuous - 4pc")
}

// The 3-piece (Vitality, 15 health per 5 sec) has no sim stat, so it adds nothing.
func TestForeverDarkmantleArmor(t *testing.T) {
	set := revealSet{1676, "Darkmantle Armor", sixPieces(226845), func() *proto.Player {
		return &proto.Player{Race: proto.Race_RaceHuman, Class: proto.Class_ClassRogue,
			Spec: &proto.Player_Rogue{Rogue: &proto.Rogue{Options: &proto.RogueOptions{}}}}
	}}
	checkStatBonus(t, set, 2, resist8)
	checkStatBonus(t, set, 3, noStats)
	checkStatBonus(t, set, 6, ap40)
	checkProc(t, set, "Darkmantle Armor - 4pc")
}

// The 4-piece (Dark Reward, a 300 heal on 5% of spell casts) changes no damage and isn't modelled.
func TestForeverDeathmistRaiment(t *testing.T) {
	set := revealSet{1681, "Deathmist Raiment", sixPieces(226921), func() *proto.Player {
		return &proto.Player{Race: proto.Race_RaceHuman, Class: proto.Class_ClassWarlock,
			Spec: &proto.Player_Warlock{Warlock: &proto.Warlock{Options: &proto.WarlockOptions{}}}}
	}}
	checkStatBonus(t, set, 2, resist8)
	checkStatBonus(t, set, 3, mp5)
	checkStatBonus(t, set, 6, sp23)
	if names := revealSim(t, set, 4).GetActiveSetBonusNames(); slices.Contains(names, "Deathmist Raiment (4pc)") {
		t.Errorf("Deathmist Raiment has a 4-piece bonus: %v", names)
	}
}

// The 3-piece (Vitality) has no sim stat, so it adds nothing.
func TestForeverBattlegearOfHeroism(t *testing.T) {
	set := revealSet{1778, "Battlegear of Heroism", sixPieces(226873), func() *proto.Player {
		return &proto.Player{Race: proto.Race_RaceOrc, Class: proto.Class_ClassWarrior,
			Spec: &proto.Player_Warrior{Warrior: &proto.Warrior{Options: &proto.Warrior_Options{}}}}
	}}
	checkStatBonus(t, set, 2, resist8)
	checkStatBonus(t, set, 3, noStats)
	checkStatBonus(t, set, 6, stats.Stats{stats.Strength: 20})
	checkProc(t, set, "Battlegear of Heroism - 4pc")
}

// Four pieces of a new set and four of the Classic set of its name fit together, and both
// 4-piece bonuses apply without registering the same aura twice.
func TestForeverRevealSetsWithClassicNamesakes(t *testing.T) {
	soulforge := revealSet{1673, "Soulforge Armor", sixPieces(226993), func() *proto.Player {
		return &proto.Player{Race: proto.Race_RaceHuman, Class: proto.Class_ClassPaladin,
			Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
				Options: &proto.PaladinOptions{PrimarySeal: proto.PaladinSeal_Righteousness}}}}
	}}
	darkmantle := revealSet{1676, "Darkmantle Armor", sixPieces(226845), func() *proto.Player {
		return &proto.Player{Race: proto.Race_RaceHuman, Class: proto.Class_ClassRogue,
			Spec: &proto.Player_Rogue{Rogue: &proto.Rogue{Options: &proto.RogueOptions{}}}}
	}}
	for _, tc := range []struct {
		set     revealSet
		classic []int32 // head, hands, chest, feet
		auras   []string
	}{
		{soulforge, []int32{22091, 22090, 22089, 22087}, []string{"Crusader's Wrath (Melee)", "Item - Crusader's Wrath Proc - Soulforge Armor"}},
		{darkmantle, []int32{22005, 22006, 22009, 22003}, []string{"Darkmantle Armor - 4pc", "Rogue Armor Energize"}},
	} {
		set := tc.set
		c := revealSimWith(t, set, 4, map[proto.ItemSlot]int32{
			proto.ItemSlot_ItemSlotHead: tc.classic[0], proto.ItemSlot_ItemSlotHands: tc.classic[1],
			proto.ItemSlot_ItemSlotChest: tc.classic[2], proto.ItemSlot_ItemSlotFeet: tc.classic[3],
		})
		for _, aura := range tc.auras {
			if c.GetAura(aura) == nil {
				t.Errorf("%s 4 + 4 lacks %s", set.name, aura)
			}
		}
	}
}
