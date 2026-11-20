package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// itemSwitchRun runs the DPS Report's level 60 Enhancement fight with Smokey's Lighter (a use)
// and Hand of Justice (a proc) in the trinket slots, with the item switches set by edit.
func itemSwitchRun(t *testing.T, iterations int32, edit func(*proto.Player)) *proto.RaidSimResult {
	t.Helper()
	data, err := os.ReadFile("../dps_report/req/shaman__enhancement-level60.json")
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	player := rsr.Raid.Parties[0].Players[0]
	player.Equipment.Items[12] = &proto.ItemSpec{Id: common.SmokeysLighter}
	player.Equipment.Items[13] = &proto.ItemSpec{Id: common.HandOfJustice}
	edit(player)
	// One iteration is a debug run, long enough that Hand of Justice surely procs.
	rsr.Encounter.Duration = 60
	if iterations == 1 {
		rsr.Encounter.Duration = 600
	}
	rsr.SimOptions.Iterations = iterations
	rsr.SimOptions.Debug = iterations == 1
	res := core.RunRaidSim(rsr)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	return res
}

// Player.item_uses_off keeps an item's use from being pressed, and Player.item_effects_off
// leaves out all of an item's effects. Neither changes anything for an item we don't wear.
func TestItemSwitches(t *testing.T) {
	sim.RegisterAll()
	none := func(*proto.Player) {}
	lighterUsed := func(res *proto.RaidSimResult) bool {
		return strings.Contains(res.Logs, "{ItemID: 13171}")
	}
	handProcced := func(res *proto.RaidSimResult) bool {
		return strings.Contains(res.Logs, "attack from {SpellID: 15600}")
	}

	base := itemSwitchRun(t, 1, none)
	if !lighterUsed(base) || !handProcced(base) {
		t.Fatalf("the plain run should use Smokey's Lighter (%t) and proc Hand of Justice (%t)", lighterUsed(base), handProcced(base))
	}
	if !core.ItemHasUse(common.SmokeysLighter) || core.ItemHasUse(common.HandOfJustice) {
		t.Errorf("ItemHasUse: Smokey's Lighter %t, Hand of Justice %t, want true and false",
			core.ItemHasUse(common.SmokeysLighter), core.ItemHasUse(common.HandOfJustice))
	}

	usesOff := itemSwitchRun(t, 1, func(p *proto.Player) { p.ItemUsesOff = []int32{common.SmokeysLighter, common.HandOfJustice} })
	if lighterUsed(usesOff) || !handProcced(usesOff) {
		t.Errorf("uses off: Smokey's Lighter used %t, Hand of Justice procced %t, want false and true", lighterUsed(usesOff), handProcced(usesOff))
	}
	effectsOff := itemSwitchRun(t, 1, func(p *proto.Player) { p.ItemEffectsOff = []int32{common.SmokeysLighter, common.HandOfJustice} })
	if lighterUsed(effectsOff) || handProcced(effectsOff) {
		t.Errorf("effects off: Smokey's Lighter used %t, Hand of Justice procced %t, want both false", lighterUsed(effectsOff), handProcced(effectsOff))
	}
	// The trinkets keep their stats with their effects off: Hand of Justice's 20 attack power.
	if a, b := base.RaidMetrics.Parties[0].Players[0].Dps.Avg, effectsOff.RaidMetrics.Parties[0].Players[0].Dps.Avg; a == b {
		t.Errorf("effects off scores the same as the plain run (%.2f)", a)
	}

	// An item we don't wear changes nothing, to the last digit.
	dps := func(res *proto.RaidSimResult) float64 { return res.RaidMetrics.Parties[0].Players[0].Dps.Avg }
	plain := dps(itemSwitchRun(t, 200, none))
	other := dps(itemSwitchRun(t, 200, func(p *proto.Player) {
		p.ItemUsesOff, p.ItemEffectsOff = []int32{common.BlackhandsBreadth}, []int32{common.BlackhandsBreadth}
	}))
	if plain != other {
		t.Errorf("switches on an item we don't wear: %.4f, want %.4f", other, plain)
	}
}
