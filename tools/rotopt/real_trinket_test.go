package main

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

func furyRequest(t *testing.T) *proto.RaidSimRequest {
	t.Helper()
	data, err := os.ReadFile("../dps_report/req/warrior__fury.json")
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	return rsr
}

// A run and a written rotation cast the gear's real on-use trinket in place of the template's
// synthetic one, drop the cast when the gear has none, and keep it when the gear wears it.
func TestRealTrinket(t *testing.T) {
	sim.RegisterAll()
	k := defaultKnobs(warriorFury{})
	k["trinketAt"] = 10
	rot := warriorFury{}.Build(k)
	synthetic := rotationJSON(rot)
	if !strings.Contains(synthetic, `"itemId":991018`) {
		t.Fatalf("the template no longer casts the synthetic trinket: %s", synthetic)
	}
	trinket := func(rsr *proto.RaidSimRequest, slot proto.ItemSlot, id int32) {
		rsr.Raid.Parties[0].Players[0].Equipment.Items[slot] = &proto.ItemSpec{Id: id}
	}
	cases := []struct {
		name       string
		t1, t2     int32
		want, gone string
	}{
		// Hand of Justice has no use, Blackhand's Breadth has one.
		{"the one use", 11815, 13965, `"itemId":13965`, `"itemId":991018`},
		{"the first use", 10455, 13965, `"itemId":10455`, `"itemId":13965`},
		{"none", 11815, 19289, "", `"itemId":991018`},
		{"synthetic worn", 991018, 13965, `"itemId":991018`, `"itemId":13965`},
	}
	for _, c := range cases {
		rsr := furyRequest(t)
		trinket(rsr, proto.ItemSlot_ItemSlotTrinket1, c.t1)
		trinket(rsr, proto.ItemSlot_ItemSlotTrinket2, c.t2)
		got := rotationJSON(withRealTrinket(rot, rsr))
		if c.want != "" && !strings.Contains(got, c.want) || strings.Contains(got, c.gone) {
			t.Errorf("%s: want %s and no %s in %s", c.name, c.want, c.gone, got)
		}
	}
	if rotationJSON(rot) != synthetic {
		t.Errorf("withRealTrinket changed the template's rotation")
	}
}

// A template with two synthetic trinkets (the warrior tank's bulwark and battle trinkets) casts
// each of the gear's two on-use trinkets, the first synthetic one the first trinket slot's, not
// one of them twice. With one use worn, both cast it.
func TestRealTrinketPerSlot(t *testing.T) {
	sim.RegisterAll()
	data, err := os.ReadFile("../tankbench/req/warrior_protection_balanced.json")
	if err != nil {
		t.Fatal(err)
	}
	rot := warriorTank{}.Build(defaultKnobs(warriorTank{}))
	casts := func(r *proto.APLRotation) []int32 {
		var ids []int32
		for _, l := range r.PriorityList {
			if id := l.GetAction().GetCastSpell().GetSpellId().GetItemId(); id != 0 && id != 10646 {
				ids = append(ids, id)
			}
		}
		return ids
	}
	if got := casts(rot); !slices.Equal(got, []int32{990617, 991018}) {
		t.Fatalf("the template's trinket lines cast %v, want the bulwark then the battle trinket", got)
	}
	cases := []struct {
		name   string
		t1, t2 int32
		want   []int32
	}{
		// Chained Essence of Eranikus and Adaptive Combat Assistant have uses, Hand of Justice none.
		{"two uses", 10455, 272437, []int32{10455, 272437}},
		{"two uses, swapped", 272437, 10455, []int32{272437, 10455}},
		{"one use", 11815, 272437, []int32{272437, 272437}},
	}
	for _, c := range cases {
		rsr := &proto.RaidSimRequest{}
		if err := protojson.Unmarshal(data, rsr); err != nil {
			t.Fatal(err)
		}
		items := rsr.Raid.Parties[0].Players[0].Equipment.Items
		items[proto.ItemSlot_ItemSlotTrinket1] = &proto.ItemSpec{Id: c.t1}
		items[proto.ItemSlot_ItemSlotTrinket2] = &proto.ItemSpec{Id: c.t2}
		if got := casts(withRealTrinket(rot, rsr)); !slices.Equal(got, c.want) {
			t.Errorf("%s: casts %v, want %v", c.name, got, c.want)
		}
	}
}
