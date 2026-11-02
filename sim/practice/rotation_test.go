//go:build practice

package practice

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	goproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

var castLine = regexp.MustCompile(`^\[.*?\] Casting \{(SpellID|ItemID): (\d+)(?:, Tag: (-?\d+))?\}`)

type aplCast struct {
	at      time.Duration
	spellID int32
	itemID  int32
}

// runAPL runs the request's own rotation for one seed, and returns the player's DPS and every cast
// it made with its sim time.
func runAPL(t *testing.T, request *proto.RaidSimRequest, seed int64) (float64, []aplCast) {
	request = goproto.Clone(request).(*proto.RaidSimRequest)
	request.SimOptions = &proto.SimOptions{Iterations: 1, RandomSeed: seed, UseLabeledRands: true}
	RegisterAll()
	sim := core.NewSim(request, simsignals.Signals{})
	char := sim.Raid.Parties[0].Players[0].GetCharacter()
	var casts []aplCast
	prefix := "[" + char.Label + "] Casting "
	sim.Log = func(message string, vals ...interface{}) {
		if !strings.HasPrefix(message, "[") {
			return
		}
		line := fmt.Sprintf(message, vals...)
		if !strings.HasPrefix(line, prefix) {
			return
		}
		m := castLine.FindStringSubmatch(line)
		if m == nil {
			return
		}
		if m[3] != "" && m[3] != "0" {
			return // tagged spells: the other players' cooldowns and procs
		}
		id, _ := strconv.Atoi(m[2])
		c := aplCast{at: sim.CurrentTime}
		if m[1] == "SpellID" {
			c.spellID = int32(id)
		} else {
			c.itemID = int32(id)
		}
		casts = append(casts, c)
	}
	sim.Reset()
	sim.PrePull()
	for !sim.Step() {
	}
	sim.Cleanup()
	return char.GetMetricsProto().Dps.Avg, casts
}

// The rotation check (docs/practice_room.md, Milestone 1 step 1): the enh APL's own casts, with
// their sim times, fed back as inputs in Encounter mode give the same DPS as the APL run. An APL
// casts inside the event that triggered it and an input runs after it, so single seeds drift
// apart. So we compare the means over 100 seeds with labeled random streams.
func TestRotationCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("slow")
	}
	data, err := os.ReadFile("testdata/enh60.json")
	if err != nil {
		t.Fatal(err)
	}
	original, err := ParseRequest(data)
	if err != nil {
		t.Fatal(err)
	}
	original.Encounter.Duration = 60
	original.Encounter.DurationVariation = 0
	normalized, err := Normalize(original)
	if err != nil {
		t.Fatal(err)
	}
	normalized.SimOptions.UseLabeledRands = true
	requestJSON, err := protojson.Marshal(normalized)
	if err != nil {
		t.Fatal(err)
	}

	const seeds = 100
	var apl, practice []float64
	for seed := int64(1); seed <= seeds; seed++ {
		dps, casts := runAPL(t, original, seed)
		apl = append(apl, dps)

		// You stand in melee behind the dummy, as the sim's fights assume, and turn auto attack
		// on at the pull. Only the casts you can press are inputs: the log's other "Casting"
		// lines are totem pulses and procs.
		inputs := []AttemptInput{{AppliedAtNs: int64(-prepullWindow), Seq: 0, Input: Input{Kind: "move", Pos: [2]float64{0, 3}, Facing: -math.Pi / 2}}}
		inputs = append(inputs, AttemptInput{AppliedAtNs: 0, Seq: 1, Input: Input{Kind: "startAttack"}})
		toolkit := toolkitIDs(t, normalized)
		for i, c := range casts {
			if !toolkit[c.spellID] && !toolkit[-c.itemID] {
				continue
			}
			in := Input{ID: i + 10, Kind: "cast", SpellID: c.spellID, Target: &UnitRef{Kind: "enemy"}}
			if c.itemID != 0 {
				in = Input{ID: i + 10, Kind: "useItem", ItemID: c.itemID}
			}
			inputs = append(inputs, AttemptInput{AppliedAtNs: int64(c.at), Seq: i + 2, Input: in})
		}
		// Inputs go in order of time, then arrival.
		attempt := &Attempt{
			Version: 1, Request: requestJSON, Seed: seed, Mode: "encounter", SetLengthMs: 60000,
			StartNs: int64(-prepullWindow), EndedAtMs: 60000, Inputs: sortInputs(inputs),
		}
		s, err := Replay(attempt)
		if err != nil {
			t.Fatal(err)
		}
		practice = append(practice, s.Summary().DPS)
	}

	meanA, seA := meanSE(apl)
	meanP, seP := meanSE(practice)
	se := math.Sqrt(seA*seA + seP*seP)
	t.Logf("APL %.1f ± %.1f DPS, inputs %.1f ± %.1f DPS over %d seeds", meanA, seA, meanP, seP, seeds)
	if math.Abs(meanA-meanP) > 3*se+0.01*meanA {
		t.Errorf("the inputs give %.1f DPS, the APL %.1f: more than the noise apart", meanP, meanA)
	}
}

func sortInputs(in []AttemptInput) []AttemptInput {
	out := append([]AttemptInput(nil), in...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && (out[j].AppliedAtNs < out[j-1].AppliedAtNs); j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func meanSE(xs []float64) (float64, float64) {
	var sum, sq float64
	for _, x := range xs {
		sum += x
	}
	mean := sum / float64(len(xs))
	for _, x := range xs {
		sq += (x - mean) * (x - mean)
	}
	sd := math.Sqrt(sq / float64(len(xs)-1))
	return mean, sd / math.Sqrt(float64(len(xs)))
}

// toolkitIDs are the spell ids, and the item ids negated, that a session's toolkit holds.
func toolkitIDs(t *testing.T, request *proto.RaidSimRequest) map[int32]bool {
	s, err := NewSession(request, Settings{}, 1)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[int32]bool{}
	for _, spell := range s.Toolkit() {
		if spell.ItemID != 0 {
			ids[-spell.ItemID] = true
		} else {
			ids[spell.SpellID] = true
		}
	}
	return ids
}
