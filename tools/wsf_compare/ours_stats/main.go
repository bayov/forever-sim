// Prints the player's and the target's stats at the pull for a RaidSimRequest, the way
// wowsims/forever's cmpstats does for theirs (tools/wsf_compare/theirs/cmpstats), so align.py and
// layers.py can put the two sheets side by side.
//
//	go build -tags=with_db -o ours_stats ./tools/wsf_compare/ours_stats
//	ours_stats <request.json>
package main

import (
	"fmt"
	"os"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	sim.RegisterAll()
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	req := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(b, req); err != nil {
		panic(err)
	}
	s := core.NewSim(req, simsignals.CreateSignals())
	s.Reset()
	s.PrePull()
	c := s.Raid.Parties[0].Players[0].GetCharacter()
	fmt.Println("PLAYER", c.GetStats().FlatString())
	fmt.Printf("PSEUDO %+v\n", c.PseudoStats)
	fmt.Println("TARGET", s.Encounter.TargetUnits[0].GetStats().FlatString())
	fmt.Printf("TPSEUDO %+v\n", s.Encounter.TargetUnits[0].PseudoStats)
}
