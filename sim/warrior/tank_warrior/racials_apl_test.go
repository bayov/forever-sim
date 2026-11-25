package tankwarrior

import (
	"os"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// racialCasts sims a tank bench request as the given race on a fixed-length fight and returns the
// average casts per fight of each of the given spells, summed over their tags.
func racialCasts(t *testing.T, req string, race proto.Race, seconds float64, spells ...int32) map[int32]float64 {
	t.Helper()
	data, err := os.ReadFile("../../../tools/tankbench/req/" + req + ".json")
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	const iterations = 20
	rsr.Raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	rsr.Encounter.Duration, rsr.Encounter.DurationVariation = seconds, 0
	rsr.SimOptions.Iterations, rsr.SimOptions.RandomSeed = iterations, 1
	rsr.Raid.Parties[0].Players[0].Race = race
	res := core.RunRaidSim(rsr)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	casts := map[int32]float64{}
	for _, id := range spells {
		casts[id] = 0
	}
	for _, a := range res.RaidMetrics.Parties[0].Players[0].Actions {
		if _, ok := casts[a.Id.GetSpellId()]; !ok {
			continue
		}
		for _, tm := range a.Targets {
			casts[a.Id.GetSpellId()] += float64(tm.Casts) / iterations
		}
	}
	return casts
}

// TestForeverTankRacialLines pins the level 60 Protection APLs' racial lines on a 170 s fight (a
// 180 s one can end on the frame a 3 min cooldown comes back): an Orc casts Blood Fury (20572, 2 min
// cooldown) at the pull and at 2 min, a Gnome Eureka! (1259812, 2 min) the same, a Troll Berserking
// (26297, 3 min) and a Night Elf Elune's Light (1259799, 3 min) once at the pull, a Dwarf Stoneform
// (20594) once after 20 s (Balanced and Max survival; Max threat has no Stoneform line), and a
// Human none of them. The lines are the same in every race's APL, so a missing or broken line shows
// here.
func TestForeverTankRacialLines(t *testing.T) {
	const bloodFury, berserking, eureka, elunesLight, stoneform = 20572, 26297, 1259812, 1259799, 20594
	racials := []int32{bloodFury, berserking, eureka, elunesLight, stoneform}
	for _, req := range []string{"warrior_protection_balanced", "warrior_protection_max_threat", "warrior_protection_max_survival"} {
		t.Run(req, func(t *testing.T) {
			for _, c := range []struct {
				race  proto.Race
				spell int32
				casts float64
			}{
				{proto.Race_RaceOrc, bloodFury, 2},
				{proto.Race_RaceGnome, eureka, 2},
				{proto.Race_RaceTroll, berserking, 1},
				{proto.Race_RaceNightElf, elunesLight, 1},
				{proto.Race_RaceDwarf, stoneform, 1},
				{proto.Race_RaceHuman, 0, 0},
			} {
				got := racialCasts(t, req, c.race, 170, racials...)
				for _, id := range racials {
					want := 0.0
					if id == c.spell && !(id == stoneform && req == "warrior_protection_max_threat") {
						want = c.casts
					}
					if got[id] != want {
						t.Errorf("%s casts %d %.2f times a 170 s fight, want %.0f", c.race, id, got[id], want)
					}
				}
			}
		})
	}
}
