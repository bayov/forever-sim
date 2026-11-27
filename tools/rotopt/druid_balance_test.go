package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	goproto "google.golang.org/protobuf/proto"
)

// balanceTrinketHold names what makes a rotation hold for a trinket or use one by itself: a
// synthetic item (withRealTrinket turns it into the worn on-use trinket) or holdFor's
// time-to-ready and uses-lost values.
func balanceTrinketHold(t *testing.T, rot *proto.APLRotation) string {
	t.Helper()
	if hasSyntheticItem(rot.ProtoReflect()) {
		return "a synthetic item"
	}
	data, err := protojson.Marshal(rot)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"spellTimeToReady", "spellUsesLostByDelay"} {
		if strings.Contains(string(data), `"`+key+`"`) {
			return key
		}
	}
	return ""
}

// The druid_balance template writes no trinket hold and no pre-pull trinket, at any knob's ends
// and with the knobs druid_presets.sh ships, and those knobs give the report's Balance rotations.
func TestDruidBalanceNoTrinketHold(t *testing.T) {
	tmpl := druidBalance{}
	for _, kn := range tmpl.Knobs() {
		for _, v := range []float64{kn.Min, kn.Max} {
			k := defaultKnobs(tmpl)
			k["cdStack"], k["elune"] = 1, 1
			k[kn.Name] = v
			if what := balanceTrinketHold(t, tmpl.Build(k)); what != "" {
				t.Errorf("%s=%v: the rotation has %s", kn.Name, v, what)
			}
		}
	}

	script, err := os.ReadFile("druid_presets.sh")
	if err != nil {
		t.Fatal(err)
	}
	gen := regexp.MustCompile(`(?m)^gen druid_balance "\$balance" (\S+) (\S+) "([^"]*)"$`)
	matches := gen.FindAllStringSubmatch(string(script), -1)
	if len(matches) != 4 {
		t.Fatalf("druid_presets.sh: %d druid_balance gen lines, want 4", len(matches))
	}
	for _, m := range matches {
		k := defaultKnobs(tmpl)
		for _, kv := range strings.Split(m[3], ",") {
			name, val, _ := strings.Cut(kv, "=")
			if _, ok := k[name]; !ok {
				t.Fatalf("%s: unknown knob %q", m[2], name)
			}
			k[name], _ = strconv.ParseFloat(val, 64)
		}
		rot := tmpl.Build(k)
		if what := balanceTrinketHold(t, rot); what != "" {
			t.Errorf("%s: the rotation has %s", m[2], what)
		}
		data, err := os.ReadFile("../../" + m[1] + "/" + m[2] + ".apl.json")
		if err != nil {
			t.Fatal(err)
		}
		shipped := &proto.APLRotation{}
		if err := protojson.Unmarshal(data, shipped); err != nil {
			t.Fatal(err)
		}
		if what := balanceTrinketHold(t, shipped); what != "" {
			t.Errorf("%s/%s.apl.json has %s", m[1], m[2], what)
		}
		if !goproto.Equal(rot, shipped) {
			got, _ := protojson.Marshal(rot)
			t.Errorf("%s: the template's rotation is not the shipped one\ngot  %s\nwant %s", m[2], got, data)
		}
	}
}
