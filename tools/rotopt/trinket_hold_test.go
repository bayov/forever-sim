package main

import (
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
	"google.golang.org/protobuf/encoding/protojson"
)

// wardenTrinketUses runs the Warden's Balanced preset once on a 180 sec fight with the template
// built from k and returns when the synthetic armor trinket (999216) was used, in seconds.
func wardenTrinketUses(t *testing.T, k Knobs) []float64 {
	t.Helper()
	data, err := os.ReadFile("../tankbench/req/shaman_warden_balanced.json")
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	rsr.Raid.Parties[0].Players[0].Rotation = shamanTank{}.Build(k)
	rsr.Encounter.Duration = 180
	rsr.Encounter.DurationVariation = 0
	rsr.SimOptions.Iterations = 1
	rsr.SimOptions.Debug = true
	res := core.RunRaidSim(rsr)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	var uses []float64
	id := strconv.Itoa(shaman.WardenSyntheticOnUseTrinket)
	re := regexp.MustCompile(`\[(-?[0-9.]+)\].*Aura gained: \{ItemID: ` + id + `\}`)
	for _, m := range re.FindAllStringSubmatch(res.Logs, -1) {
		at, _ := strconv.ParseFloat(m[1], 64)
		uses = append(uses, at)
	}
	return uses
}

// The Warden's armor trinket (1,665 armor for 20 sec, 2 min cooldown) is a survival cooldown, so
// armorTrinket=0 never uses it (it was typed DPS and autocast on cooldown), and trinketAt holds the
// second use for the last trinketAt seconds at every value of its range, also under the buff's
// 20 sec (those acted like 0). Tank_mechanics-20261104-16.
func TestWardenArmorTrinketHold(t *testing.T) {
	sim.RegisterAll()
	cases := []struct {
		armorTrinket, trinketAt float64
		second                  float64 // when the second use starts; 0 for none
	}{
		{0, 0, -1},
		{1, 0, 120},
		{1, 5, 175},
		{1, 10, 170},
		{1, 15, 165},
		{1, 25, 155},
		{1, 55, 125},
	}
	for _, c := range cases {
		k := defaultKnobs(shamanTank{})
		k["armorTrinket"], k["trinketAt"] = c.armorTrinket, c.trinketAt
		uses := wardenTrinketUses(t, k)
		if c.second < 0 {
			if len(uses) != 0 {
				t.Errorf("armorTrinket=0: used at %v, want never", uses)
			}
			continue
		}
		if len(uses) != 2 || uses[0] > 1.5 || uses[1] < c.second-1.5 || uses[1] > c.second+1.5 {
			t.Errorf("trinketAt=%v: used at %v, want at the pull and about %v", c.trinketAt, uses, c.second)
		}
	}
}
