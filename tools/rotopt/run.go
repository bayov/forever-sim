package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	goproto "google.golang.org/protobuf/proto"
)

// The sim setup everything is measured against: the player, buffs and encounter from
// either a RaidSimRequest or the IndividualSimSettings JSON the UI exports. The rotation
// and fight length are replaced per run.
type setup struct {
	request *proto.RaidSimRequest
}

func loadSetup(path string) (*setup, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, err
	}

	request := &proto.RaidSimRequest{}
	if _, isRequest := probe["raid"]; isRequest {
		if err := protojson.Unmarshal(data, request); err != nil {
			return nil, err
		}
		return &setup{request: request}, nil
	}

	settings := &proto.IndividualSimSettings{}
	if err := protojson.Unmarshal(data, settings); err != nil {
		return nil, err
	}
	request.Raid = core.SinglePlayerRaidProto(settings.Player, settings.PartyBuffs, settings.RaidBuffs, settings.Debuffs)
	request.Encounter = settings.Encounter
	request.SimOptions = &proto.SimOptions{}
	if settings.Settings != nil {
		request.SimOptions.Ruleset = settings.Settings.Ruleset
	}
	return &setup{request: request}, nil
}

func (s *setup) player() *proto.Player {
	return s.request.Raid.Parties[0].Players[0]
}

type result struct {
	dps    float64
	stderr float64
}

func (r result) String() string {
	return fmt.Sprintf("%.1f(%.1f)", r.dps, r.stderr)
}

// One sim run. The random seed is fixed so two rotations see the same crit rolls and
// the difference between them is far less noisy than either number alone.
func (s *setup) run(rot *proto.APLRotation, duration float64, iterations int32) (result, error) {
	request := goproto.Clone(s.request).(*proto.RaidSimRequest)
	request.Raid.Parties[0].Players[0].Rotation = rot
	request.Encounter.Duration = duration
	request.SimOptions.Iterations = iterations
	request.SimOptions.RandomSeed = 1

	res := core.RunRaidSimConcurrent(request)
	if res.Error != nil {
		return result{}, fmt.Errorf("%s", res.Error.Message)
	}
	dps := res.RaidMetrics.Parties[0].Players[0].Dps
	return result{dps: dps.Avg, stderr: dps.Stdev / math.Sqrt(float64(iterations))}, nil
}

// One logged iteration, reduced to the casts that are not the filler: cooldowns, items
// and finishers. For eyeballing whether a rotation does what its knobs say.
func (s *setup) timeline(rot *proto.APLRotation, duration float64) string {
	request := goproto.Clone(s.request).(*proto.RaidSimRequest)
	request.Raid.Parties[0].Players[0].Rotation = rot
	request.Encounter.Duration = duration
	request.Encounter.DurationVariation = 0
	request.SimOptions.Iterations = 1
	request.SimOptions.RandomSeed = 1
	request.SimOptions.Debug = true

	res := core.RunRaidSim(request)
	if res.Error != nil {
		return "sim failed: " + res.Error.Message
	}
	filler := regexp.MustCompile(`OtherID|SpellID: (11294|11340|25347|15851|13218)[,}]`)
	var b strings.Builder
	for _, line := range strings.Split(res.Logs, "\n") {
		if strings.Contains(line, "] Casting {") && !filler.MatchString(line) {
			b.WriteString(line[:strings.Index(line, " (Cost")] + "\n")
		}
	}
	return b.String()
}

// Rotation JSON in the layout of the presets under ui/*/apls: one priority list line per
// row, so a diff of the file reads as a diff of the rotation.
func rotationJSON(rot *proto.APLRotation) string {
	var b strings.Builder
	b.WriteString("{\n    \"type\": \"TypeAPL\",\n")
	if len(rot.PrepullActions) > 0 {
		b.WriteString("    \"prepullActions\": [\n")
		for i, item := range rot.PrepullActions {
			b.WriteString("        " + string(protojson.MarshalOptions{}.Format(item)))
			if i < len(rot.PrepullActions)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString("    ],\n")
	}
	b.WriteString("    \"priorityList\": [\n")
	for i, item := range rot.PriorityList {
		b.WriteString("        " + compactJSON(item))
		if i < len(rot.PriorityList)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("    ]\n}\n")
	return b.String()
}

// protojson pads its output with random spaces on purpose, so we round trip through
// encoding/json for a stable compact form.
func compactJSON(m goproto.Message) string {
	data, _ := protojson.Marshal(m)
	var v interface{}
	json.Unmarshal(data, &v)
	out, _ := json.Marshal(v)
	return string(out)
}
