package sim

// The healing fight's groundwork (docs/healing_sim.md, H1): heals aimed at hurt players,
// overhealing, every player's health tracked, death, the boss's raid damage and tank swaps. The
// healer is tools/healbench's Restoration shaman. Needs the item database: go test -tags with_db.

import (
	"os"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

func benchPlayer(t testing.TB, path string) *proto.Player {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	return rsr.Raid.Parties[0].Players[0]
}

// healingRaid is tanks, healers and Frost mages in that order, in a healing fight against the
// report's boss with the events given.
func healingRaid(t *testing.T, tanks, healers, mages int, events ...*proto.RaidDamageEvent) *proto.RaidSimRequest {
	tank := benchPlayer(t, "../tools/tankbench/req/warrior_protection.json")
	healer := benchPlayer(t, "../tools/healbench/req/shaman_restoration.json")
	mage := raidPlayer(t, "mage__frost")
	var players []*proto.Player
	for i := 0; i < tanks; i++ {
		players = append(players, googleProto.Clone(tank).(*proto.Player))
	}
	for i := 0; i < healers; i++ {
		players = append(players, googleProto.Clone(healer).(*proto.Player))
	}
	for i := 0; i < mages; i++ {
		players = append(players, googleProto.Clone(mage).(*proto.Player))
	}
	rsr := raidOf(t, "mage__frost", players, 20)
	for i := 0; i < tanks; i++ {
		rsr.Raid.Tanks = append(rsr.Raid.Tanks, &proto.UnitReference{Type: proto.UnitReference_Player, Index: int32(i)})
	}
	rsr.Encounter.TrackHealth = true
	rsr.Encounter.RaidDamageHitsPerMinute = 0
	rsr.Encounter.DurationVariation = 0
	rsr.Encounter.RaidDamage = events
	rsr.Encounter.Targets[0].TankIndex = 0
	rsr.Encounter.Targets[0].SwingSpeed = 2
	rsr.Encounter.Targets[0].MinBaseDamage = 3000
	return rsr
}

func playerMetrics(res *proto.RaidSimResult, index int) *proto.UnitMetrics {
	return res.RaidMetrics.Parties[index/5].Players[index%5]
}

// healingOn is a player's healing on the unit, and the part of it that was overhealing.
func healingOn(m *proto.UnitMetrics, unitIndex int32) (healing, overhealing float64) {
	for _, a := range m.Actions {
		for _, tgt := range a.Targets {
			if tgt.UnitIndex == unitIndex {
				healing += tgt.Healing
				overhealing += tgt.Overhealing
			}
		}
	}
	return
}

// The shaman heals the tank the boss hits and the mages the raid damage hits, not itself only,
// and some of it is overhealing but most isn't.
func TestHealingFightHealsOthers(t *testing.T) {
	pulse := &proto.RaidDamageEvent{School: proto.SpellSchool_SpellSchoolFire, MinDamage: 400, MaxDamage: 500, EverySeconds: 10, FirstAt: 5}
	res := runRaid(t, healingRaid(t, 1, 1, 2, pulse))
	healer := playerMetrics(res, 1)

	for i := 0; i < 4; i++ {
		if i == 1 {
			continue
		}
		healing, overhealing := healingOn(healer, playerMetrics(res, i).UnitIndex)
		if healing <= 0 {
			t.Errorf("player %d got no healing", i)
		}
		if overhealing < 0 || overhealing > healing {
			t.Errorf("player %d: %.0f overhealing of %.0f healing", i, overhealing, healing)
		}
	}
	if healer.Ehps.Avg <= 0 || healer.Ehps.Avg > healer.Hps.Avg {
		t.Errorf("EHPS %.1f against HPS %.1f", healer.Ehps.Avg, healer.Hps.Avg)
	}
	for i := 2; i < 4; i++ {
		if m := playerMetrics(res, i); m.Dtps.Avg <= 0 {
			t.Errorf("mage %d took no raid damage", i)
		}
	}
}

// Raid damage nobody can heal kills every mage: they stop casting, so they deal less damage than
// mages who live, and the dead stay dead.
func TestHealingFightDeath(t *testing.T) {
	deadly := &proto.RaidDamageEvent{School: proto.SpellSchool_SpellSchoolShadow, MinDamage: 2500, MaxDamage: 2500, Who: proto.RaidDamageEvent_NotTanks, EverySeconds: 5, FirstAt: 20}
	// No tank, so the boss hits nobody.
	dead := runRaid(t, healingRaid(t, 0, 0, 2, deadly))
	alive := runRaid(t, healingRaid(t, 0, 0, 2))
	for i := 0; i < 2; i++ {
		d, a := playerMetrics(dead, i), playerMetrics(alive, i)
		if d.ChanceOfDeath != 1 {
			t.Errorf("mage %d died in %.0f%% of fights, want all", i, 100*d.ChanceOfDeath)
		}
		if a.ChanceOfDeath != 0 {
			t.Errorf("mage %d died with no raid damage", i)
		}
		if d.Dps.Avg >= 0.5*a.Dps.Avg {
			t.Errorf("mage %d dealt %.0f DPS dying at 30 s, %.0f alive", i, d.Dps.Avg, a.Dps.Avg)
		}
	}
}

// With a tank swap every 20 s both tanks take the boss's melee, and without one only the first.
func TestHealingFightTankSwap(t *testing.T) {
	rsr := healingRaid(t, 2, 2, 0)
	res := runRaid(t, rsr)
	if second := playerMetrics(res, 1).Dtps.Avg; second != 0 {
		t.Errorf("the second tank took %.0f DTPS without a swap", second)
	}

	rsr = healingRaid(t, 2, 2, 0)
	rsr.Encounter.Targets[0].TankSwapSeconds = 20
	res = runRaid(t, rsr)
	first, second := playerMetrics(res, 0).Dtps.Avg, playerMetrics(res, 1).Dtps.Avg
	if first <= 0 || second <= 0 || second < 0.5*first || first < 0.5*second {
		t.Errorf("tanks took %.0f and %.0f DTPS with a swap every 20 s", first, second)
	}
}
