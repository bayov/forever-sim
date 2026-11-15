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

// druidFeralCastsNearSwing runs the Feral report request once on a fixed 180 sec fight with the
// druid_feral template's shipped knobs plus wfGate and counts the main hand abilities (Shred,
// Rake, Rip, Ferocious Bite, Ravage) cast less than 0.1 sec after a main hand white swing, and all
// of them.
func druidFeralCastsNearSwing(t *testing.T, wfGate float64) (near, all int) {
	t.Helper()
	data, err := os.ReadFile("../dps_report/req/druid__feral-cat.json")
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	k := Knobs{}
	for _, kn := range (druidFeral{}).Knobs() {
		k[kn.Name] = kn.Default
	}
	k["rake"], k["wfGate"] = 1, wfGate
	rsr.Raid.Parties[0].Players[0].Rotation = druidFeral{}.Build(k)
	rsr.Encounter.Duration = 180
	rsr.Encounter.DurationVariation = 0
	rsr.SimOptions.Iterations = 1
	rsr.SimOptions.Debug = true
	res := core.RunRaidSim(rsr)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	re := regexp.MustCompile(`(?m)^\[([0-9.]+)\] \[ \(#1\)\] Casting \{(OtherID: 7, Tag: 1|SpellID: (\d+))\}`)
	mh := map[string]bool{"9830": true, "9904": true, "9896": true, "22829": true, "9867": true}
	lastSwing := -1.0
	for _, m := range re.FindAllStringSubmatch(res.Logs, -1) {
		at, _ := strconv.ParseFloat(m[1], 64)
		if m[3] == "" {
			lastSwing = at
			continue
		}
		if !mh[m[3]] {
			continue
		}
		all++
		// The log rounds to 0.01 sec.
		if lastSwing >= 0 && at-lastSwing < 0.095 {
			near++
		}
	}
	return near, all
}

// wfGate=1 keeps the cat's main hand abilities out of the 0.1 sec after a white swing, where
// Windfury Totem's 100 ms internal cooldown would stop them rolling it; without the gate some land
// there.
func TestDruidFeralWFGate(t *testing.T) {
	sim.RegisterAll()
	near0, all0 := druidFeralCastsNearSwing(t, 0)
	near1, all1 := druidFeralCastsNearSwing(t, 1)
	t.Logf("near a swing: %d of %d without the gate, %d of %d with it", near0, all0, near1, all1)
	if near, all := near0, all0; near == 0 || all < 50 {
		t.Errorf("wfGate=0: %d of %d abilities within 0.1 sec of a swing, want some", near, all)
	}
	if near, all := near1, all1; near != 0 || all < 50 {
		t.Errorf("wfGate=1: %d of %d abilities within 0.1 sec of a swing, want none", near, all)
	}
}
