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

// warriorAuraGains runs the dual wield Fury report request once on a fixed 180 sec fight with the
// warrior_fury template built from k and returns when the aura matching id (as the debug log
// writes it, e.g. "SpellID: 12328") was gained, in seconds.
func warriorAuraGains(t *testing.T, k Knobs, id string) []float64 {
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
	var at []float64
	re := regexp.MustCompile(`\[(-?[0-9.]+)\].*Aura gained: \{` + regexp.QuoteMeta(id) + `\}`)
	for _, m := range re.FindAllStringSubmatch(res.Logs, -1) {
		v, _ := strconv.ParseFloat(m[1], 64)
		at = append(at, v)
	}
	return at
}

// The warrior DPS holds (dwEnd for Death Wish's 30 sec, potionAt and trinketAt for the potion's and
// the trinket's 15 sec) keep the held use for the last N seconds also when N is under the buff: with
// the buff as the time a delayed use needed left, a hold under it never counted the held use, so the
// use went at once (Death Wish at the pull, the potion's and trinket's second use on cooldown at 120).
func TestWarriorDpsHoldUnderBuff(t *testing.T) {
	sim.RegisterAll()
	cases := []struct {
		knob  string
		value float64
		aura  string
		want  []float64 // when each use starts
	}{
		{"dwEnd", 15, "SpellID: 12328", []float64{165}},
		{"dwEnd", 21, "SpellID: 12328", []float64{159}},
		{"dwEnd", 36, "SpellID: 12328", []float64{144}},
		{"potionAt", 10, "ItemID: 13442", []float64{0, 170}},
		{"potionAt", 20, "ItemID: 13442", []float64{0, 160}},
		{"trinketAt", 5, "ItemID: 991018", []float64{0, 175}},
		{"trinketAt", 10, "ItemID: 991018", []float64{0, 170}},
		{"trinketAt", 30, "ItemID: 991018", []float64{0, 150}},
	}
	for _, c := range cases {
		k := defaultKnobs(warriorFury{})
		k["deathWish"], k["potionHoldForDw"], k["potionPull"] = 2, 2, 0
		k[c.knob] = c.value
		uses := warriorAuraGains(t, k, c.aura)
		ok := len(uses) == len(c.want)
		for i := 0; ok && i < len(uses); i++ {
			// A use may wait a little for Rage (Death Wish's 10): white swings alone feed Windfury
			// Totem under Forever, so dwEnd=15 lands about 2 sec late. Early is the bug this guards.
			ok = uses[i] >= c.want[i]-1.5 && uses[i] <= c.want[i]+3
		}
		if !ok {
			t.Errorf("%s=%v: %s gained at %v, want about %v", c.knob, c.value, c.aura, uses, c.want)
		}
	}
}
