package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/wowsims/classic/sim/core/proto"
)

// A stat weights file, like ui/enhancement_shaman/leveling_weights.json: the weights at some
// levels by stat name, which -weights writes and a config (config.go) reads.
type weightsFile struct {
	Levels []weightsRow `json:"levels"`
}

type weightsRow struct {
	Level  int32   `json:"level"`
	Mobs   float64 `json:"mobs,omitempty"`
	Deaths float64 `json:"deaths,omitempty"`
	// The mobs a session one attack power adds, which every weight is next to.
	MobsPerAP float64            `json:"mobsPerAP,omitempty"`
	Weights   map[string]float64 `json:"weights"`
	// How far off each weight can be (one standard error), from the sessions of the probes.
	Errors map[string]float64 `json:"errors,omitempty"`
}

// Under Forever some weights count for more than the stat they name, as the UI's presets have it
// (EPLevelingGearPlan in ui/enhancement_shaman/presets.ts).
var weightNames = map[string][]proto.Stat{
	"MeleeHit":   {proto.Stat_StatMeleeHit, proto.Stat_StatSpellHit},
	"MeleeCrit":  {proto.Stat_StatMeleeCrit, proto.Stat_StatSpellCrit},
	"SpellPower": {proto.Stat_StatSpellPower, proto.Stat_StatSpellDamage},
	"Armor":      {proto.Stat_StatArmor, proto.Stat_StatBonusArmor},
}

var weightPseudoNames = map[string]proto.PseudoStat{
	"MainHandDps": proto.PseudoStat_PseudoStatMainHandDps,
	"OffHandDps":  proto.PseudoStat_PseudoStatOffHandDps,
	"RangedDps":   proto.PseudoStat_PseudoStatRangedDps,
}

// weightsToProto is a row's weights as the gear plan takes them.
func weightsToProto(weights map[string]float64) (*proto.UnitStats, error) {
	out := &proto.UnitStats{Stats: make([]float64, len(proto.Stat_name)), PseudoStats: make([]float64, 3)}
	for name, value := range weights {
		if pseudo, ok := weightPseudoNames[name]; ok {
			out.PseudoStats[pseudo] = value
			continue
		}
		stats, ok := weightNames[name]
		if !ok {
			s, known := proto.Stat_value["Stat"+name]
			if !known {
				return nil, fmt.Errorf("no stat %q", name)
			}
			stats = []proto.Stat{proto.Stat(s)}
		}
		for _, s := range stats {
			out.Stats[s] = value
		}
	}
	return out, nil
}

func readWeightsFile(path string) (*weightsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	file := &weightsFile{}
	if err := json.Unmarshal(data, file); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return file, nil
}

// presets are the file's weights as a request has them, for a weapon style.
func (f *weightsFile) presets(style proto.LevelingWeaponStyle) ([]*proto.LevelingGearWeights, error) {
	var out []*proto.LevelingGearWeights
	for _, row := range f.Levels {
		weights, err := weightsToProto(row.Weights)
		if err != nil {
			return nil, fmt.Errorf("level %d: %w", row.Level, err)
		}
		out = append(out, &proto.LevelingGearWeights{Level: row.Level, Weights: weights, Style: style})
	}
	slices.SortStableFunc(out, func(a, b *proto.LevelingGearWeights) int { return int(a.Level - b.Level) })
	return out, nil
}
