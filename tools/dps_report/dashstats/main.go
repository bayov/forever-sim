// dashstats prints a RaidSimRequest's first player's stat sheet as JSON (local dashboard helper).
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	sim.RegisterAll()
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	req := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, req); err != nil {
		panic(err)
	}
	res := core.ComputeStats(&proto.ComputeStatsRequest{Raid: req.Raid, Encounter: req.Encounter})
	if res.ErrorResult != "" {
		fmt.Fprintln(os.Stderr, res.ErrorResult)
		os.Exit(1)
	}
	p := res.RaidStats.Parties[0].Players[0]
	out := map[string]any{"sets": p.Sets}
	layers := map[string]*proto.UnitStats{"base": p.BaseStats, "gear": p.GearStats, "talents": p.TalentsStats, "buffs": p.BuffsStats, "consumes": p.ConsumesStats, "final": p.FinalStats}
	for name, us := range layers {
		m := map[string]float64{}
		if us != nil {
			for i, v := range us.Stats {
				if v != 0 {
					m[stats.Stat(i).StatName()] = v
				}
			}
			for i, v := range us.PseudoStats {
				if v != 0 {
					m[fmt.Sprintf("pseudo%d", i)] = v
				}
			}
		}
		out[name] = m
	}
	b, _ := json.Marshal(out)
	fmt.Println(string(b))
}
