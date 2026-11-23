package main

import (
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

func repSource(f proto.Faction) *proto.UIItemSource {
	return &proto.UIItemSource{Source: &proto.UIItemSource_Rep{Rep: &proto.RepSource{RepFactionId: 1, PlayerFaction: f}}}
}

func dropSource(npc int32) *proto.UIItemSource {
	return &proto.UIItemSource{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{NpcId: npc}}}
}

func questSource(id int32) *proto.UIItemSource {
	return &proto.UIItemSource{Source: &proto.UIItemSource_Quest{Quest: &proto.QuestSource{Id: id}}}
}

func TestFactionCheck(t *testing.T) {
	tags := map[int32]proto.Faction{1: proto.Faction_Horde}
	quests := map[int32]proto.Faction{500: proto.Faction_Alliance}
	items := []struct {
		item *proto.UIItem
		// What the check says on a Night Elf and on an Orc.
		nightElf, orc string
	}{
		// A content tag wins over everything else.
		{&proto.UIItem{Id: 1, Sources: []*proto.UIItemSource{dropSource(10)}}, "content", ""},
		{&proto.UIItem{Id: 2, FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY}, "", "item"},
		// Every source one faction's: a reputation, a quest only one side can take, a
		// faction vendor.
		{&proto.UIItem{Id: 3, Sources: []*proto.UIItemSource{repSource(proto.Faction_Horde)}}, "sources", ""},
		{&proto.UIItem{Id: 4, Sources: []*proto.UIItemSource{questSource(500)}}, "", "sources"},
		{&proto.UIItem{Id: 5, Sources: []*proto.UIItemSource{dropSource(263569)}}, "", "sources"},
		{&proto.UIItem{Id: 6, Sources: []*proto.UIItemSource{dropSource(263570), repSource(proto.Faction_Horde)}}, "sources", ""},
		// One source open to both factions keeps it, and so do two sides' reputations.
		{&proto.UIItem{Id: 7, Sources: []*proto.UIItemSource{repSource(proto.Faction_Horde), dropSource(10)}}, "", ""},
		{&proto.UIItem{Id: 8, Sources: []*proto.UIItemSource{repSource(proto.Faction_Alliance), repSource(proto.Faction_Horde)}}, "", ""},
		{&proto.UIItem{Id: 9, Sources: []*proto.UIItemSource{questSource(501)}}, "", ""},
		{&proto.UIItem{Id: 10}, "", ""},
	}
	for _, race := range []proto.Race{proto.Race_RaceNightElf, proto.Race_RaceOrc} {
		c := &factionCheck{faction: map[proto.Race]proto.Faction{proto.Race_RaceNightElf: proto.Faction_Alliance, proto.Race_RaceOrc: proto.Faction_Horde}[race], tags: tags, quests: quests}
		for _, tc := range items {
			want := tc.nightElf
			if race == proto.Race_RaceOrc {
				want = tc.orc
			}
			if got := c.otherFaction(tc.item); got != want {
				t.Errorf("%v, item %d: got %q, want %q", race, tc.item.Id, got, want)
			}
		}
	}
	var none *factionCheck
	if got := none.otherFaction(items[0].item); got != "" {
		t.Errorf("no check: got %q", got)
	}
}

// The real files: the content check's tags and wowhead's quest sides load, the build's faction
// follows its race (a Skyborne Windshaper is Horde), and the Horde quest rewards the Night Elf
// druids picked are left out for them.
func TestFactionCheckFiles(t *testing.T) {
	for race, want := range map[proto.Race]proto.Faction{
		proto.Race_RaceNightElf:           proto.Faction_Alliance,
		proto.Race_RaceSkyborneHighOrder:  proto.Faction_Alliance,
		proto.Race_RaceSkyborneWindshaper: proto.Faction_Horde,
		proto.Race_RaceUndead:             proto.Faction_Horde,
	} {
		c, err := newFactionCheck(race, "phase0_content.json", "../../assets/db_inputs/forever_wowhead_new_quests.json")
		if err != nil {
			t.Fatal(err)
		}
		if c.faction != want {
			t.Errorf("%v: faction %v, want %v", race, c.faction, want)
		}
		if len(c.quests) == 0 {
			t.Errorf("no quest sides read")
		}
		eye := &proto.UIItem{Id: 13968, Name: "Eye of the Beast"}
		theramore := &proto.UIItem{Id: 271955, Name: "Theramore Pauldrons"}
		if got := c.otherFaction(eye) != ""; got != (want == proto.Faction_Alliance) {
			t.Errorf("%v: Eye of the Beast left out %v", race, got)
		}
		if got := c.otherFaction(theramore) != ""; got != (want == proto.Faction_Horde) {
			t.Errorf("%v: Theramore Pauldrons left out %v", race, got)
		}
	}
}

func TestEnchantFaction(t *testing.T) {
	checks, err := loadEnchantChecks("phase0_enchants.json")
	if err != nil {
		t.Fatal(err)
	}
	fiery := checks[6296]
	if !otherFactionEnchant(fiery, proto.Race_RaceOrc) || otherFactionEnchant(fiery, proto.Race_RaceHuman) {
		t.Errorf("Fiery Blaze Enchantment, an Alliance quest's, should be out for an Orc only")
	}
	if _, err := parseFaction("Scourge"); err == nil {
		t.Errorf("an unknown faction should be an error")
	}
}
