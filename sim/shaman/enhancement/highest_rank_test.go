package enhancement

import (
	"os"
	"strconv"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"google.golang.org/protobuf/encoding/protojson"
)

// TestHighestRankPick: a rotation that names an ability's top rank casts the highest rank we
// know at any level, and one that names a rank we know casts that rank. The Leveling Plan's
// rotation for levels 10 to 14 names top ranks, so at 14 it casts Earth Shock rank 3, learned
// at 14, where it cast rank 2 before. Healing Wave named by its top rank casts rank 2 at 10
// and rank 3 (learned at 12) at 14.
func TestHighestRankPick(t *testing.T) {
	completed := func(req *proto.RaidSimRequest) map[string]int {
		req.SimOptions.Debug = true
		result := core.RunRaidSim(req)
		if result.Error != nil {
			t.Fatal(result.Error.Message)
		}
		casts := map[string]int{}
		logLines(t, result.Logs, func(_ float64, line string) {
			if m := completedLine.FindStringSubmatch(line); m != nil {
				casts[m[1]]++
			}
		})
		return casts
	}
	at := func(level int32, rotation *proto.APLRotation) *proto.RaidSimRequest {
		req := normalMobRequest([]*proto.Target{core.NormalMobTarget(level, proto.MobType_MobTypeBeast, 1)}, true, 3)
		req.Raid.Parties[0].Players[0].Level = level
		req.Raid.Parties[0].Players[0].Rotation = rotation
		return req
	}

	data, err := os.ReadFile("../../../ui/enhancement_shaman/apls/leveling_plan_10.apl.json")
	if err != nil {
		t.Fatal(err)
	}
	plan := &proto.APLRotation{}
	if err := protojson.Unmarshal(data, plan); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		level      int32
		earthShock int32
	}{{10, 8044}, {14, 8045}} {
		if casts := completed(at(c.level, plan)); casts[strconv.Itoa(int(c.earthShock))] == 0 {
			t.Errorf("level %d: the plan's rotation cast no Earth Shock %d (casts %v)", c.level, c.earthShock, casts)
		}
	}

	// A heal as soon as a hit lands, naming Healing Wave's top rank or rank 2 on purpose.
	heal := func(spellID int32) *proto.APLRotation {
		rotation := &proto.APLRotation{}
		if err := protojson.Unmarshal([]byte(rotationOf(
			`{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"99%"}}}},"castSpell":{"spellId":{"spellId":`+strconv.Itoa(int(spellID))+`}}}}`)), rotation); err != nil {
			t.Fatal(err)
		}
		return rotation
	}
	for _, c := range []struct {
		level       int32
		named, cast int32
		notCast     int32
	}{{10, 25357, 332, 547}, {14, 25357, 547, 332}, {14, 332, 332, 547}} {
		casts := completed(at(c.level, heal(c.named)))
		if casts[strconv.Itoa(int(c.cast))] == 0 || casts[strconv.Itoa(int(c.notCast))] > 0 {
			t.Errorf("level %d: a rotation naming Healing Wave %d cast %d %d times and %d %d times, want only %d",
				c.level, c.named, c.cast, casts[strconv.Itoa(int(c.cast))], c.notCast, casts[strconv.Itoa(int(c.notCast))], c.cast)
		}
	}

	// The editor's "highest rank" pick: the highest rank we know carries the top rank's ID,
	// the lower ranks and a top rank we know carry none.
	for _, c := range []struct {
		level   int32
		spellID int32
		want    int32
	}{{14, 547, 25357}, {14, 332, 0}, {14, 8045, 10414}, {60, 25357, 0}} {
		sim := core.NewSim(at(c.level, &proto.APLRotation{}), simsignals.CreateSignals())
		enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
		var got int32 = -1
		for _, spell := range enh.GetMetadata().Spells {
			if spell.Id.GetSpellId() == c.spellID {
				got = spell.HighestRankId
			}
		}
		if got != c.want {
			t.Errorf("level %d: spell %d has highest rank ID %d, want %d", c.level, c.spellID, got, c.want)
		}
	}
}
