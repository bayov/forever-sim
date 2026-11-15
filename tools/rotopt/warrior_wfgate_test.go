package main

import (
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// warriorStrikesAfterSwing runs the dual wield Fury report request once on a fixed 180 sec fight
// with the warrior_fury template built from k, and counts the Bloodthirst, Whirlwind and Hamstring
// casts that come within 0.1 sec after a main hand swing (white or Heroic Strike) and before it
// (the next swing within 0.1 sec), out of all of them.
func warriorStrikesAfterSwing(t *testing.T, k Knobs) (after, before, all int) {
	t.Helper()
	data, err := os.ReadFile("../dps_report/req/warrior__fury.json")
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	rsr.Raid.Parties[0].Players[0].Rotation = warriorFury{}.Build(k)
	rsr.Encounter.Duration = 180
	rsr.Encounter.DurationVariation = 0
	rsr.SimOptions.Iterations = 1
	rsr.SimOptions.Debug = true
	res := core.RunRaidSim(rsr)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	// In log order, so a Windfury Totem attack a strike procs (it can take a queued Heroic Strike)
	// does not count as a swing ahead of it. The log rounds to 10 ms.
	re := regexp.MustCompile(`\[([0-9.]+)\].*Casting \{(OtherID: 7, Tag: 1|SpellID: 11567, Tag: 1|SpellID: 23894|SpellID: 1680|SpellID: 7373)\}`)
	last := -10.0
	var strikes []float64
	for _, m := range re.FindAllStringSubmatch(res.Logs, -1) {
		at, _ := strconv.ParseFloat(m[1], 64)
		if m[2] == "OtherID: 7, Tag: 1" || m[2] == "SpellID: 11567, Tag: 1" {
			last = at
			for _, st := range strikes {
				if at > st+0.005 && at-st < 0.095 {
					before++
				}
			}
			continue
		}
		all++
		strikes = append(strikes, at)
		if at-last < 0.095 {
			after++
		}
	}
	return after, before, all
}

// wfGate keeps the main hand strikes out of Windfury Totem's 100 ms internal cooldown after a
// main hand swing (1), and out of the 100 ms before the next one too (2).
func TestWarriorWfGate(t *testing.T) {
	sim.RegisterAll()
	k := defaultKnobs(warriorFury{})
	k["hamstringRage"] = 10
	if after, _, all := warriorStrikesAfterSwing(t, k); after == 0 || all < 50 {
		t.Fatalf("wfGate=0: %d of %d strikes right after a swing, want some (the check would prove nothing)", after, all)
	}
	k["wfGate"] = 1
	if after, _, all := warriorStrikesAfterSwing(t, k); after != 0 || all < 50 {
		t.Errorf("wfGate=1: %d of %d strikes within 0.1 sec after a main hand swing, want none", after, all)
	}
	k["wfGate"] = 2
	if after, before, all := warriorStrikesAfterSwing(t, k); after != 0 || before != 0 || all < 50 {
		t.Errorf("wfGate=2: %d / %d of %d strikes within 0.1 sec after / before a main hand swing, want none", after, before, all)
	}
}
