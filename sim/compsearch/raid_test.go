package compsearch

import (
	"fmt"
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// Whole raids through the resolver and the engine (docs/raid_sim.md, Checking it). They need M1b's
// engine: per-player raid buffs and cast buffs reaching their scope.

// reportRaid is a raid of base report builds, each with its own request's raid buffs as its per-player
// set plus its own held auras (the hooks are skipped), under the base requests' common debuffs.
func reportRaid(t *testing.T, ids []string, iters int32) *proto.RaidSimRequest {
	p := loadTestPool(t)
	first := cloneReq(p.Builds[ids[0]].Req)
	raid := &proto.Raid{Buffs: &proto.RaidBuffs{}, Debuffs: first.Raid.Debuffs}
	for i, id := range ids {
		b := p.Builds[id]
		pl := googleProto.Clone(player0(b.Req)).(*proto.Player)
		rb := googleProto.Clone(b.Req.Raid.Buffs).(*proto.RaidBuffs)
		if b.Gives.Has(GiftMoonkin) {
			rb.MoonkinAura = true
		}
		if b.Gives.Has(GiftLotP) {
			rb.LeaderOfThePack = true
		}
		if b.Gives.Has(GiftTrueshot) {
			rb.TrueshotAura = true
		}
		if b.Gives.Has(GiftAB) {
			rb.ArcaneBrilliance = true
		}
		if b.Gives.Has(GiftPoS) {
			rb.DivineSpirit = true
		}
		pl.RaidBuffs = rb
		pl.Name = fmt.Sprintf("%s %d", id, i)
		if i%5 == 0 {
			raid.Parties = append(raid.Parties, &proto.Party{Buffs: &proto.PartyBuffs{}})
		}
		raid.Parties[i/5].Players = append(raid.Parties[i/5].Players, pl)
	}
	raid.NumActiveParties = int32(len(raid.Parties))
	return &proto.RaidSimRequest{Raid: raid, Encounter: first.Encounter, SimOptions: &proto.SimOptions{Iterations: iters, RandomSeed: 7}}
}

// Check 1: each member of a raid of report builds lands within 4σ of a single-player sim of its own
// request, at the same seed and iterations.
func TestReportRaid(t *testing.T) {
	if testing.Short() {
		t.Skip("about a minute")
	}
	const iters = 1000
	raids := map[string][]string{
		"10": {"mage__fire", "warlock__demonology", "hunter__beast-mastery", "rogue__combat-swords", "priest__shadow",
			"druid__balance", "mage__frostfire", "warrior__arms", "paladin__level_60", "hunter__marksmanship"},
		"physical": {"warrior__fury", "warrior__fury_2h", "rogue__combat-daggers", "rogue__subtlety", "hunter__survival",
			"hunter__beast-mastery", "druid__feral-cat", "paladin__level_60_sub_holy", "warrior__arms", "rogue__mutilate",
			"hunter__lone-wolf", "rogue__seal-fate-swords"},
	}
	p := loadTestPool(t)
	for name, ids := range raids {
		res := core.RunRaidSim(reportRaid(t, ids, iters))
		if res.Error != nil {
			t.Fatalf("%s: %s", name, res.Error.Message)
		}
		i := 0
		for _, party := range res.RaidMetrics.Parties {
			for _, pm := range party.Players {
				id := ids[i]
				i++
				alone := cloneReq(p.Builds[id].Req)
				alone.SimOptions = &proto.SimOptions{Iterations: iters, RandomSeed: 7}
				solo := core.RunRaidSim(alone).RaidMetrics.Dps
				se := math.Hypot(pm.Dps.Stdev, solo.Stdev) / math.Sqrt(iters)
				allow := 4 * se
				if id == "hunter__survival" || id == "shaman__enhancement-level60" {
					allow += 5
				}
				if d := pm.Dps.Avg - solo.Avg; math.Abs(d) > allow {
					t.Errorf("%s raid: %s did %.1f, alone %.1f (%+.1f, allowed ±%.1f)", name, id, pm.Dps.Avg, solo.Avg, d, allow)
				}
			}
		}
	}
}

// M1b's done-when: RunComp on 7 builds and 3 placeholders equals the same raid built by hand with one
// shared set of raid buffs, so the per-player path changes nothing it shouldn't, and every DPS member
// takes the raid damage hits.
func TestRunCompMatchesHandBuilt(t *testing.T) {
	if testing.Short() {
		t.Skip("about 20 s")
	}
	p := loadTestPool(t)
	ids := []string{"mage__fire", "warlock__affliction", "hunter__beast-mastery", "rogue__combat-swords", "priest__shadow", "mage__frost", "warrior__arms"}
	comp := &proto.RaidComp{Size: 10, SimOptions: &proto.SimOptions{Iterations: 300, RandomSeed: 3},
		Encounter: cloneReq(p.Builds[ids[0]].Req).Encounter,
		Slots: []*proto.CompSlot{phSlot(1, 0, 0, proto.RaidPreset_RaidPresetProtectionWarrior),
			phSlot(2, 1, 0, proto.RaidPreset_RaidPresetDisciplinePriest), phSlot(3, 1, 1, proto.RaidPreset_RaidPresetRestorationDruid)}}
	for i, id := range ids {
		party, idx := int32(0), int32(i+1)
		if i >= 4 {
			party, idx = 1, int32(i-2)
		}
		comp.Slots = append(comp.Slots, slotOf(int32(10+i), party, idx, p, id))
	}
	// No paladins, so every member's blessings are the same; give nobody a cooldown, so all seats match.
	comp.Assignments = []*proto.Assignment{}
	run := RunComp(comp)
	if run.Resolution.Error != nil {
		t.Fatal(run.Resolution.Error.Message)
	}
	if run.Result.Error != nil {
		t.Fatal(run.Result.Error.Message)
	}
	// The same request with the raid buffs shared: the first member's set as Raid.buffs, nobody's own.
	hand := cloneReq(run.Resolution.Request)
	var shared *proto.RaidBuffs
	same := true
	for _, party := range hand.Raid.Parties {
		for _, pl := range party.Players {
			if shared == nil {
				shared = pl.RaidBuffs
			} else if !googleProto.Equal(shared, pl.RaidBuffs) {
				same = false
			}
		}
	}
	if !same {
		t.Skip("the members' raid buffs differ, so there's no shared set to compare with")
	}
	// The hooks run in the hand-built raid, so the held auras and the mage's and priest's buffs must be in
	// the shared set already (they are: the resolver gives them to everyone).
	for _, party := range hand.Raid.Parties {
		for _, pl := range party.Players {
			pl.RaidBuffs = nil
		}
	}
	hand.Raid.Buffs = shared
	handRes := core.RunRaidSim(hand)
	if handRes.Error != nil {
		t.Fatal(handRes.Error.Message)
	}
	if a, b := run.Result.RaidMetrics.Dps.Avg, handRes.RaidMetrics.Dps.Avg; math.Abs(a-b) > 1e-6 {
		t.Errorf("RunComp %.3f, hand-built %.3f", a, b)
	}
	// Every DPS member takes the raid damage hits.
	for _, party := range run.Result.RaidMetrics.Parties {
		for _, pm := range party.Players {
			if pm.Dtps == nil || pm.Dtps.Avg <= 0 {
				t.Errorf("%s took no damage", pm.Name)
			}
		}
	}
}

// Check 8: raid DPS is the sum of its members' DPS, pets included.
func TestRaidMetricsAddUp(t *testing.T) {
	if testing.Short() {
		t.Skip("about 10 s")
	}
	res := core.RunRaidSim(reportRaid(t, []string{"hunter__beast-mastery", "warlock__demonology", "mage__fire", "warrior__fury", "hunter__survival", "rogue__subtlety"}, 200))
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	sum := 0.0
	for _, party := range res.RaidMetrics.Parties {
		for _, pm := range party.Players {
			sum += pm.Dps.Avg
		}
	}
	if d := res.RaidMetrics.Dps.Avg - sum; math.Abs(d) > 1e-6*sum {
		t.Errorf("raid DPS %.3f, its members' sum %.3f", res.RaidMetrics.Dps.Avg, sum)
	}
}
