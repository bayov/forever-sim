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
	// talents replaces the player's talent string when set. The talent search moves it.
	talents string
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

// A copy of the request with the talent override applied, for one run to fill in.
func (s *setup) clone() *proto.RaidSimRequest {
	request := goproto.Clone(s.request).(*proto.RaidSimRequest)
	if s.talents != "" {
		request.Raid.Parties[0].Players[0].TalentsString = s.talents
	}
	return request
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
	request := s.clone()
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
	request := s.clone()
	request.Raid.Parties[0].Players[0].Rotation = rot
	request.Encounter.Duration = duration
	request.Encounter.DurationVariation = 0
	request.SimOptions.Iterations = 1
	request.SimOptions.RandomSeed = 1
	request.SimOptions.Debug = true

	var b strings.Builder
	// The rotation's own warnings first: a line for a spell the build does not have is
	// dropped, and it should be clear which ones were.
	stats := core.ComputeStats(&proto.ComputeStatsRequest{Raid: request.Raid, Encounter: request.Encounter, Ruleset: request.SimOptions.Ruleset})
	if stats.RaidStats != nil {
		for i, item := range stats.RaidStats.Parties[0].Players[0].RotationStats.PriorityList {
			for _, w := range item.Warnings {
				fmt.Fprintf(&b, "line %d: %s\n", i+1, w)
			}
		}
	}

	res := core.RunRaidSim(request)
	if res.Error != nil {
		return "sim failed: " + res.Error.Message
	}
	if os.Getenv("ROTOPT_FULL_LOG") != "" {
		return b.String() + res.Logs
	}
	// Every rank of the builders and poisons, so the filter holds at any level.
	filler := regexp.MustCompile(`OtherID|SpellID: (1752|1757|1758|1759|1760|8621|11293|11294|53|2589|2590|2591|8721|11279|11280|11281|25300|8679|8686|8688|11338|11339|11340|2823|2824|11355|11356|25347|15851|13218|13222|13223|13224)[,}]`)
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

// The player's final stats as the sim sees them, one per line.
func (s *setup) finalStats(rot *proto.APLRotation) string {
	request := s.clone()
	request.Raid.Parties[0].Players[0].Rotation = rot
	stats := core.ComputeStats(&proto.ComputeStatsRequest{Raid: request.Raid, Encounter: request.Encounter, Ruleset: request.SimOptions.Ruleset})
	var sb strings.Builder
	for i, v := range stats.RaidStats.Parties[0].Players[0].FinalStats.Stats {
		if v != 0 {
			fmt.Fprintf(&sb, "%s %g\n", strings.TrimPrefix(proto.Stat(i).String(), "Stat"), v)
		}
	}
	return sb.String()
}

// The setup written back out as a RaidSimRequest JSON, so a later run can start from the
// gear or talents this one found.
func (s *setup) save(path string) error {
	data, err := protojson.MarshalOptions{Multiline: true, Indent: " "}.Marshal(s.clone())
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Stat weights on the current gear: DPS per point of each stat, and the same as
// attack power equivalents. A stat mod of 1 point is tiny at low level, so this asks for
// many iterations.
func (s *setup) statWeights(rot *proto.APLRotation, duration float64, iterations int32) string {
	request := s.clone()
	player := request.Raid.Parties[0].Players[0]
	player.Rotation = rot
	request.Encounter.Duration = duration
	res := core.StatWeights(&proto.StatWeightsRequest{
		Player:     player,
		RaidBuffs:  request.Raid.Buffs,
		PartyBuffs: request.Raid.Parties[0].Buffs,
		Debuffs:    request.Raid.Debuffs,
		Encounter:  request.Encounter,
		SimOptions: &proto.SimOptions{Iterations: iterations, RandomSeed: 1, Ruleset: request.SimOptions.Ruleset},
		StatsToWeigh: []proto.Stat{
			proto.Stat_StatAgility, proto.Stat_StatStrength, proto.Stat_StatAttackPower,
			proto.Stat_StatMeleeCrit, proto.Stat_StatMeleeHit, proto.Stat_StatStamina,
		},
		PseudoStatsToWeigh: []proto.PseudoStat{proto.PseudoStat_PseudoStatMainHandDps, proto.PseudoStat_PseudoStatOffHandDps},
		EpReferenceStat:    proto.Stat_StatAttackPower,
	})
	if res.Error != nil {
		return "stat weights failed: " + res.Error.Message
	}
	var sb strings.Builder
	names := []string{"Agility", "Strength", "AttackPower", "MeleeCrit", "MeleeHit", "Stamina"}
	for _, name := range names {
		i := proto.Stat_value["Stat"+name]
		fmt.Fprintf(&sb, "%-12s %6.3f dps  %6.2f ap (±%.2f)\n", name, res.Dps.Weights.Stats[i], res.Dps.EpValues.Stats[i], res.Dps.EpValuesStdev.Stats[i])
	}
	for _, name := range []string{"MainHandDps", "OffHandDps"} {
		i := proto.PseudoStat_value["PseudoStat"+name]
		fmt.Fprintf(&sb, "%-12s %6.3f dps  %6.2f ap (±%.2f)\n", name, res.Dps.Weights.PseudoStats[i], res.Dps.EpValues.PseudoStats[i], res.Dps.EpValuesStdev.PseudoStats[i])
	}
	return sb.String()
}
