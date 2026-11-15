package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// enhDebugLog runs the Enhancement Level 60 report request once on a fixed fight of the given length with
// the shaman_enh template built from the defaults and the given knobs, and returns the log.
func enhDebugLog(t *testing.T, knobs map[string]float64, seconds float64) string {
	t.Helper()
	data, err := os.ReadFile("../dps_report/req/shaman__enhancement-level60.json")
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	// Other tests leave the template's level and talent points behind; the request is level 60.
	defer func(level int32, points int) { playerLevel, maxTalentPoints = level, points }(playerLevel, maxTalentPoints)
	playerLevel, maxTalentPoints = 60, 51
	k := Knobs{}
	for _, knob := range (shamanEnh{}).Knobs() {
		k[knob.Name] = knob.Default
	}
	for name, v := range map[string]float64{"shock": 4, "airTotem": 2, "fireTotem": 2} {
		k[name] = v
	}
	for name, v := range knobs {
		k[name] = v
	}
	rsr.Raid.Parties[0].Players[0].Rotation = shamanEnh{}.Build(k)
	rsr.Encounter.Duration = seconds
	rsr.Encounter.DurationVariation = 0
	rsr.SimOptions.Iterations = 1
	rsr.SimOptions.Debug = true
	res := core.RunRaidSim(rsr)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	return res.Logs
}

var enhLogTime = regexp.MustCompile(`^\[(-?[0-9.]+)\]`)

// enhEvents returns, in log order, the time of each line that contains one of the markers, and
// which marker it was.
func enhEvents(logs string, markers ...string) (at []float64, which []int) {
	for _, line := range strings.Split(logs, "\n") {
		for i, m := range markers {
			if strings.Contains(line, m) {
				if tm := enhLogTime.FindStringSubmatch(line); tm != nil {
					v, _ := strconv.ParseFloat(tm[1], 64)
					at, which = append(at, v), append(which, i)
				}
				break
			}
		}
	}
	return at, which
}

// stormstrikesAfterWindfury counts the Stormstrikes cast less than 1.5 sec after a Windfury
// Weapon proc, while the weapon's cooldown is still running, and all Stormstrikes.
func stormstrikesAfterWindfury(logs string) (early, all int) {
	at, which := enhEvents(logs, "Casting {SpellID: 17364}", "Casting {SpellID: 16362, Tag: 1}")
	lastProc := -100.0
	for i, w := range which {
		if w == 1 {
			lastProc = at[i]
			continue
		}
		all++
		if at[i]-lastProc < 1.5-1e-6 {
			early++
		}
	}
	return early, all
}

// stormstrikeWfHold keeps Stormstrike for Windfury Weapon's 1.5 sec cooldown: with the full hold
// no Stormstrike goes while the cooldown runs, and without it some do.
func TestShamanEnhStormstrikeWindfuryHold(t *testing.T) {
	sim.RegisterAll()
	early, all := stormstrikesAfterWindfury(enhDebugLog(t, nil, 180))
	t.Logf("no hold: %d of %d Stormstrikes during the cooldown", early, all)
	if early == 0 || all < 15 {
		t.Fatalf("without the hold, %d of %d Stormstrikes went during the cooldown, want some", early, all)
	}
	for _, idle := range []float64{0, 1} {
		early, all := stormstrikesAfterWindfury(enhDebugLog(t, map[string]float64{"stormstrikeWfHold": 1.5, "stormstrikeWfIdle": idle}, 180))
		t.Logf("stormstrikeWfIdle=%v: %d of %d", idle, early, all)
		if early != 0 || all < 15 {
			t.Errorf("stormstrikeWfIdle=%v: %d of %d Stormstrikes went during Windfury Weapon's cooldown, want none", idle, early, all)
		}
	}
}

// stormstrikesAfter counts the Stormstrikes cast less than 100 ms after a log line with the
// marker, and all Stormstrikes.
func stormstrikesAfter(logs, marker string) (early, all int) {
	at, which := enhEvents(logs, "Casting {SpellID: 17364}", marker)
	lastProc := -100.0
	for i, w := range which {
		if w == 1 {
			lastProc = at[i]
			continue
		}
		all++
		if at[i]-lastProc < 0.1-1e-6 {
			early++
		}
	}
	return early, all
}

// stormstrikeTotemGate keeps Stormstrike out of Windfury Totem's 100 ms cooldown: 2 and 3 off the
// 100 ms after every main hand swing, 1 and 4 off the 100 ms after a totem proc. The sim seldom
// puts a Stormstrike in a proc's 100 ms anyway (none in this fight), so 1 and 4 have no
// counter-example.
func TestShamanEnhStormstrikeTotemGate(t *testing.T) {
	sim.RegisterAll()
	const swing, totemProc = "Casting {OtherID: 7, Tag: 1}", "extra main-hand attack from {SpellID: 10610}"
	early, all := stormstrikesAfter(enhDebugLog(t, nil, 1200), swing)
	t.Logf("no gate: %d of %d Stormstrikes within 100 ms of a swing", early, all)
	if early == 0 || all < 100 {
		t.Fatalf("without the gate, %d of %d Stormstrikes went within 100 ms of a swing, want some", early, all)
	}
	for _, c := range []struct {
		gate   float64
		marker string
	}{{1, totemProc}, {2, swing}, {3, swing}, {4, totemProc}} {
		early, all := stormstrikesAfter(enhDebugLog(t, map[string]float64{"stormstrikeTotemGate": c.gate}, 1200), c.marker)
		t.Logf("stormstrikeTotemGate=%v: %d of %d", c.gate, early, all)
		if early != 0 || all < 100 {
			t.Errorf("stormstrikeTotemGate=%v: %d of %d Stormstrikes went within 100 ms of %q, want none", c.gate, early, all, c.marker)
		}
	}
}
