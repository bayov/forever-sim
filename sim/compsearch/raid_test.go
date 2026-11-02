package compsearch

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/priest"
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
	// What a member casts reaches its scope in the raid, so nobody gets it given too (docs/raid_sim.md,
	// Checking it, check 1): an Enhancement shaman's Strength of Earth and Windfury Totem for everyone,
	// and every DPS shaman's Mana Spring for its own party, Elemental's too (the resolver's dropsTotems).
	// The given Mana Spring is a plain MP5 and the cast one ticks of its own, so left on both they stack.
	for pi, party := range raid.Parties {
		for i, pl := range party.Players {
			b := p.Builds[ids[pi*5+i]]
			if b.Casts.Has(GiftSoE) {
				for _, other := range raid.Parties {
					for _, o := range other.Players {
						o.RaidBuffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectMissing
						if o.RaidBuffs.TotemWeaponBuff == proto.TotemWeaponBuff_TotemWeaponBuffWindfury {
							o.RaidBuffs.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffNone
						}
					}
				}
			}
			if dropsTotems(b) {
				for _, o := range party.Players {
					if o != pl {
						o.RaidBuffs.ManaSpringTotem = proto.TristateEffect_TristateEffectMissing
					}
				}
			}
		}
	}
	return &proto.RaidSimRequest{Raid: raid, Encounter: first.Encounter, SimOptions: &proto.SimOptions{Iterations: iters, RandomSeed: 7}}
}

// soloRequest is member i's own request with what the raid's other members cast for it given, as the
// resolver gives it (docs/raid_sim.md, Checking it, check 1), at reportRaid's seed:
//   - an Enhancement shaman's Strength of Earth and Windfury Totem (a hunter's request has neither);
//   - one Sunder Armor at the pull from each other warrior (Seat.PullSunders), on top of the raid's given
//     ramp: with a warrior or two in the raid the boss's armor is down sooner, worth about 1.5 DPS a
//     warrior to a Beast Mastery hunter;
//   - a Shadow priest's Dark Inspiration for its party (+3.7 to Demonology, +4.4 to Fire), which the
//     search doesn't give yet but the raid's priest casts.
func soloRequest(p *Pool, ids []string, i int, iters int32) *proto.RaidSimRequest {
	b := p.Builds[ids[i]]
	alone := cloneReq(b.Req)
	alone.SimOptions = &proto.SimOptions{Iterations: iters, RandomSeed: 7}
	if slices.ContainsFunc(ids, func(x string) bool { return p.Builds[x].Casts.Has(GiftSoE) }) && !b.Casts.Has(GiftSoE) {
		alone.Raid.Buffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectRegular
		alone.Raid.Buffs.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffWindfury
	}
	// The raid's debuffs are the first member's (reportRaid), its given ramp included.
	ramp := &proto.SunderRamp{}
	if r := p.Builds[ids[0]].Req.Raid.Debuffs.SunderRamp; r != nil {
		ramp = googleProto.Clone(r).(*proto.SunderRamp)
	}
	for j, x := range ids {
		if j != i && pullSunderer(p.Builds[x]) {
			ramp.PullSunders++
		}
	}
	alone.Raid.Debuffs.SunderRamp = ramp
	for j := i / 5 * 5; j < min(len(ids), i/5*5+5); j++ {
		if j != i && darkInspirer(p.Builds[ids[j]]) {
			alone.Raid.Parties[0].Buffs.DarkInspiration = true
		}
	}
	return alone
}

// darkInspirer is whether a build is a priest with Dark Inspiration, which reaches its party.
func darkInspirer(b *Build) bool {
	pl := player0(b.Req)
	if pl.Class != proto.Class_ClassPriest {
		return false
	}
	raid := &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{googleProto.Clone(pl).(*proto.Player)}}}}
	env, _, _ := core.NewEnvironment(raid, &proto.Encounter{}, proto.Ruleset_RulesetForever, false)
	pa, ok := env.Raid.Parties[0].Players[0].(priest.PriestAgent)
	return ok && pa.GetPriest().Talents.DarkInspiration > 0
}

// Check 1: each member of a raid of report builds lands within 4σ of a single-player sim of its own
// request, at the same seed and iterations, given what the raid's other members cast for it (soloRequest).
func TestReportRaid(t *testing.T) {
	if testing.Short() {
		t.Skip("about a minute")
	}
	const iters = 1000
	raids := map[string][]string{
		"10": {"mage__fire", "warlock__demonology", "hunter__beast-mastery", "rogue__combat-daggers", "priest__shadow",
			"druid__balance", "mage__frostfire", "warrior__arms", "paladin__level_60_sub_holy", "hunter__marksmanship"},
		"20": {"shaman__enhancement-imp-gw", "warrior__fury", "rogue__combat-daggers", "hunter__marksmanship", "mage__fire",
			"warlock__demonology", "priest__shadow", "druid__balance", "paladin__level_60_sub_holy", "hunter__beast-mastery",
			"mage__frostfire", "priest__smite", "warrior__arms", "rogue__seal-fate-daggers", "druid__feral-cat",
			"hunter__survival", "mage__arcane", "warlock__destruction", "warlock__affliction", "shaman__elemental-level60"},
		"physical": {"warrior__fury", "rogue__combat-daggers", "rogue__subtlety", "hunter__survival",
			"hunter__beast-mastery", "druid__feral-cat", "paladin__level_60_sub_holy", "warrior__arms", "rogue__seal-fate-daggers",
			"hunter__marksmanship"},
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
				alone := soloRequest(p, ids, i-1, iters)
				solo := core.RunRaidSim(alone).RaidMetrics.Dps
				se := math.Hypot(pm.Dps.Stdev, solo.Stdev) / math.Sqrt(iters)
				allow := 4 * se
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
	ids := []string{"mage__fire", "warlock__affliction", "hunter__beast-mastery", "rogue__combat-daggers", "priest__shadow", "mage__frost", "warrior__arms"}
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
		t.Fatal("the members' raid buffs differ, so there's no shared set to compare with")
	}
	// What each placeholder gives, checked on its own, so a buff the resolver forgot can't hide on both
	// sides of the comparison below.
	if shared.BattleShout == 0 || shared.GiftOfTheWild == 0 || shared.PowerWordFortitude == 0 || !shared.DivineSpirit || !shared.ArcaneBrilliance {
		t.Errorf("the placeholders' and the mage's buffs: %v", shared)
	}
	if d := hand.Raid.Debuffs; !d.FaerieFire || d.SunderRamp == nil || !d.SunderRamp.Tank {
		t.Errorf("the druid's Faerie Fire and the warrior tank's Sunder: %v", d)
	}
	pi, inn := 0, 0
	for _, party := range hand.Raid.Parties {
		for _, pl := range party.Players {
			pi += int(pl.Buffs.PowerInfusions)
			inn += int(pl.Buffs.Innervates)
		}
	}
	if pi != 1 || inn != 1 {
		t.Errorf("one Power Infusion and one Innervate: %d and %d", pi, inn)
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
	// The members' totals are their own damage and their pets': the hunter's and the warlock's pets deal
	// damage, and the boss takes all of it.
	for _, party := range res.RaidMetrics.Parties {
		for _, pm := range party.Players {
			if strings.HasPrefix(pm.Name, "hunter__beast-mastery") || strings.HasPrefix(pm.Name, "warlock__demonology") {
				petDps := 0.0
				for _, pet := range pm.Pets {
					petDps += pet.Dps.Avg
				}
				if petDps <= 0 {
					t.Errorf("%s's pet dealt no damage", pm.Name)
				}
			}
		}
	}
	if taken := res.EncounterMetrics.Targets[0].Dtps.Avg; math.Abs(taken-sum) > 0.01*sum {
		t.Errorf("the boss took %.1f DPS, the raid dealt %.1f", taken, sum)
	}
}
