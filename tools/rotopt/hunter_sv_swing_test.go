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

// svLog runs the Survival request once on a 120 sec fight with the hunter_sv rotation built from k
// and returns the debug log.
func svLog(t *testing.T, k Knobs) string {
	t.Helper()
	data, err := os.ReadFile("../dps_report/req/hunter__survival.json")
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	rsr.Raid.Parties[0].Players[0].Rotation = hunterSV{}.Build(k)
	rsr.Encounter.Duration = 120
	rsr.Encounter.DurationVariation = 0
	rsr.SimOptions.Iterations = 1
	rsr.SimOptions.Debug = true
	res := core.RunRaidSim(rsr)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	return res.Logs
}

// The gaps between each Wing Clip and the main hand swing before it, in seconds: the hunter's
// white main hand swings and Raptor Strike (which takes a swing's place) start the gap.
func svWingClipGaps(t *testing.T, k Knobs) []float64 {
	t.Helper()
	logs := svLog(t, k)
	re := regexp.MustCompile(`(?m)^\[([0-9.]+)\] \[ \(#1\)\] Casting \{(OtherID: 7, Tag: 1|SpellID: 14266, Tag: 1|SpellID: 14268)\}`)
	var gaps []float64
	last := -1.0
	for _, m := range re.FindAllStringSubmatch(logs, -1) {
		at, _ := strconv.ParseFloat(m[1], 64)
		if m[2] != "SpellID: 14268" {
			last = at
		} else if last >= 0 {
			gaps = append(gaps, at-last)
		}
	}
	return gaps
}

// swingGate=1 with swingGap holds Wing Clip (which rolls Windfury Totem) until swingGap after the
// main hand swing; without the gate some Wing Clips go out inside that window.
func TestSVSwingGap(t *testing.T) {
	sim.RegisterAll()
	k := Knobs{}
	for _, kn := range (hunterSV{}).Knobs() {
		k[kn.Name] = kn.Default
	}
	k["wingClip"] = 1
	k["swingGap"] = 0.2
	early := func(gaps []float64) int {
		n := 0
		for _, g := range gaps {
			if g < 0.2-1e-6 {
				n++
			}
		}
		return n
	}
	k["swingGate"] = 0
	free := svWingClipGaps(t, k)
	if len(free) == 0 || early(free) == 0 {
		t.Fatalf("ungated: %d Wing Clips, %d inside 0.2 sec of a swing; want some of each", len(free), early(free))
	}
	t.Logf("ungated: %d Wing Clips, %d inside 0.2 sec", len(free), early(free))
	k["swingGate"] = 1
	gated := svWingClipGaps(t, k)
	if len(gated) == 0 || early(gated) != 0 {
		t.Fatalf("gated: %d Wing Clips, %d inside 0.2 sec of a swing; want some and none", len(gated), early(gated))
	}
}
