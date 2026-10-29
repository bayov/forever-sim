package sim

// Raids of the DPS Report's builds (tools/dps_report/req), for the raid sim's bug fixes
// (docs/raid_sim.md, milestone B). Needs the item database: go test -tags with_db.

import (
	"crypto/sha256"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

const raidReqDir = "../tools/dps_report/req"

func raidReq(t testing.TB, name string) *proto.RaidSimRequest {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(raidReqDir, name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	return rsr
}

func raidPlayer(t testing.TB, name string) *proto.Player {
	return raidReq(t, name).Raid.Parties[0].Players[0]
}

// raidReqNames is every request, or with base only the ones with no duty and no encounter
// variant (meta.json has neither "duty" nor "baseBuild").
func raidReqNames(t testing.TB, base bool) []string {
	files, _ := filepath.Glob(filepath.Join(raidReqDir, "*.json"))
	var names []string
	for _, f := range files {
		if strings.HasSuffix(f, ".meta.json") {
			continue
		}
		name := strings.TrimSuffix(filepath.Base(f), ".json")
		if base {
			meta, _ := os.ReadFile(filepath.Join(raidReqDir, name+".meta.json"))
			if strings.Contains(string(meta), `"duty"`) || strings.Contains(string(meta), `"baseBuild"`) {
				continue
			}
		}
		names = append(names, name)
	}
	slices.Sort(names)
	if len(names) == 0 {
		t.Fatal("no requests in " + raidReqDir)
	}
	return names
}

// raidOf puts players in parties of 5, in order, under the first request's buffs, debuffs
// and encounter.
func raidOf(t testing.TB, base string, players []*proto.Player, iterations int32) *proto.RaidSimRequest {
	rsr := raidReq(t, base)
	raid := &proto.Raid{Buffs: rsr.Raid.Buffs, Debuffs: rsr.Raid.Debuffs}
	for i, p := range players {
		if i%5 == 0 {
			raid.Parties = append(raid.Parties, &proto.Party{})
		}
		p.Name = fmt.Sprintf("%d", i)
		raid.Parties[i/5].Players = append(raid.Parties[i/5].Players, p)
	}
	raid.NumActiveParties = int32(len(raid.Parties))
	rsr.Raid = raid
	rsr.SimOptions = &proto.SimOptions{Iterations: iterations, RandomSeed: 1}
	return rsr
}

func runRaid(t *testing.T, rsr *proto.RaidSimRequest) *proto.RaidSimResult {
	t.Helper()
	res := core.RunRaidSim(rsr)
	if res.Error != nil {
		msg := res.Error.Message
		t.Fatalf("raid sim failed: %s", msg[:min(len(msg), 400)])
	}
	return res
}

func memberDps(res *proto.RaidSimResult, party, slot int) *proto.DistributionMetrics {
	return res.RaidMetrics.Parties[party].Players[slot].GetDps()
}

// Bug: more than 200 auras on the boss panicked (aura.go registerAura). A fire mage puts 33
// there, so 6 failed.
func TestRaidForeverAuraCap(t *testing.T) {
	var players []*proto.Player
	for i := 0; i < 12; i++ {
		players = append(players, raidPlayer(t, "mage__fire"))
	}
	runRaid(t, raidOf(t, "mage__fire", players, 2))
}

// Check 3 and B's done-when: raids of 20 and 40 report builds run, a 40-player health
// fight included. Bugs: the aura cap, and presims indexed a 25-slot list by raid index.
func TestRaidForeverSize(t *testing.T) {
	base := raidReqNames(t, true)
	for _, size := range []int{20, 40} {
		for _, health := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d health=%v", size, health), func(t *testing.T) {
				var players []*proto.Player
				for i := 0; i < size; i++ {
					players = append(players, raidPlayer(t, base[i%len(base)]))
				}
				rsr := raidOf(t, "mage__fire", players, 2)
				if health {
					rsr.Encounter.UseHealth = true
					rsr.Encounter.Targets[0].Stats[stats.Health] = 6e6
				}
				res := runRaid(t, rsr)
				for p, party := range res.RaidMetrics.Parties {
					for s := range party.Players {
						if memberDps(res, p, s).Avg <= 0 {
							t.Errorf("member %d did no damage", p*5+s)
						}
					}
				}
			})
		}
	}
}

// Check 2: every base build twice in one raid, the second within noise of the build alone.
// Bugs: APL target-aura references found the first caster's aura with the spell id (Fire
// Vulnerability, Winter's Chill, Shadow Weaving, Light's Vigil, Hemorrhage), and two
// Enhancement shamans panicked on Stormstrike's fixed label.
func TestRaidForeverDoubles(t *testing.T) {
	if testing.Short() {
		t.Skip("about 45 s")
	}
	const iterations = 1000
	for _, name := range raidReqNames(t, true) {
		t.Run(name, func(t *testing.T) {
			alone := raidReq(t, name)
			alone.SimOptions.Iterations = iterations
			a := memberDps(runRaid(t, alone), 0, 0)
			d := memberDps(runRaid(t, raidOf(t, name, []*proto.Player{raidPlayer(t, name), raidPlayer(t, name)}, iterations)), 0, 1)
			limit := 4 * math.Sqrt(2) * a.Stdev / math.Sqrt(iterations)
			if diff := d.Avg - a.Avg; math.Abs(diff) > limit {
				t.Errorf("second %s: %+.1f DPS against alone, limit %.1f", name, diff, limit)
			}
		})
	}
}

// Bug: applyCharacterEffects and the rotation setup looked players up by their place in the
// party with the gaps taken out, and parties by their place with missing parties taken out.
// A nil player or party panicked.
func TestRaidForeverGaps(t *testing.T) {
	mage, fury := raidPlayer(t, "mage__fire"), raidPlayer(t, "warrior__fury")
	compact := raidOf(t, "warrior__fury", []*proto.Player{mage, fury}, 50)
	want := runRaid(t, googleProto.Clone(compact).(*proto.RaidSimRequest))

	check := func(t *testing.T, rsr *proto.RaidSimRequest, slots [][2]int) {
		res := runRaid(t, rsr)
		for i, s := range slots {
			if got, w := memberDps(res, s[0], s[1]).Avg, memberDps(want, 0, i).Avg; got != w {
				t.Errorf("member %d: %.3f DPS, %.3f without gaps", i, got, w)
			}
		}
	}
	t.Run("empty slots", func(t *testing.T) {
		rsr := googleProto.Clone(compact).(*proto.RaidSimRequest)
		p := rsr.Raid.Parties[0].Players
		rsr.Raid.Parties[0].Players = []*proto.Player{{}, p[0], nil, p[1]}
		check(t, rsr, [][2]int{{0, 1}, {0, 3}})
	})
	t.Run("missing party", func(t *testing.T) {
		rsr := googleProto.Clone(compact).(*proto.RaidSimRequest)
		rsr.Raid.Parties = []*proto.Party{nil, rsr.Raid.Parties[0]}
		rsr.Raid.NumActiveParties = 2
		res := runRaid(t, rsr)
		// The result keeps the parties where the request has them.
		if len(res.RaidMetrics.Parties) != 2 {
			t.Fatalf("%d parties in the result, want 2", len(res.RaidMetrics.Parties))
		}
		check(t, rsr, [][2]int{{1, 0}, {1, 1}})
	})
	t.Run("player reference past a missing party", func(t *testing.T) {
		rsr := googleProto.Clone(compact).(*proto.RaidSimRequest)
		rsr.Raid.Parties = []*proto.Party{nil, rsr.Raid.Parties[0]}
		rsr.Raid.NumActiveParties = 2
		env, _, _ := core.NewEnvironment(rsr.Raid, rsr.Encounter, rsr.SimOptions.Ruleset, false)
		unit := env.GetUnit(&proto.UnitReference{Type: proto.UnitReference_Player, Index: 6}, nil)
		if unit == nil || unit != &env.Raid.Parties[0].Players[1].GetCharacter().Unit {
			t.Errorf("raid index 6 found %v, want the warrior", unit)
		}
	})
}

// Bug: GetRaidBuffs wrote the class hooks' buffs (a mage's Arcane Brilliance) into the
// request.
func TestRaidForeverRequestUnchanged(t *testing.T) {
	rsr := raidOf(t, "mage__fire", []*proto.Player{raidPlayer(t, "mage__fire"), raidPlayer(t, "priest__shadow")}, 2)
	rsr.Raid.Buffs = &proto.RaidBuffs{}
	before := googleProto.Clone(rsr)
	runRaid(t, rsr)
	if !googleProto.Equal(before, rsr) {
		t.Errorf("the request changed: raid buffs %v", rsr.Raid.Buffs)
	}
}

// Bug: AddPartyBuffs ran twice for a class without its own (druid, shaman, warlock), so a
// druid's Atiesh counted double.
func TestRaidForeverAtiesh(t *testing.T) {
	druid := raidPlayer(t, "druid__balance")
	druid.Equipment.Items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 22632}
	druid.Equipment.Items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{}
	rsr := raidOf(t, "druid__balance", []*proto.Player{druid, raidPlayer(t, "mage__fire")}, 1)
	env, _, _ := core.NewEnvironment(rsr.Raid, rsr.Encounter, rsr.SimOptions.Ruleset, false)
	mage := env.Raid.Parties[0].Players[1].GetCharacter()
	if !mage.HasAura("Atiesh Greatstaff of the Guardian (MP5) 1") || mage.HasAura("Atiesh Greatstaff of the Guardian (MP5) 2") {
		t.Errorf("want one Atiesh on the druid's party")
	}
}

// Bug: the Imp's Blood Pact went raid-wide through the warlock's AddRaidBuffs. It's a party
// aura in Forever.
func TestRaidForeverBloodPact(t *testing.T) {
	stamina := func(raid *proto.Raid, party, slot int) float64 {
		res := core.ComputeStats(&proto.ComputeStatsRequest{Raid: raid, Encounter: raidReq(t, "rogue__subtlety").Encounter})
		return res.RaidStats.Parties[party].Players[slot].FinalStats.Stats[stats.Stamina]
	}
	lock := raidPlayer(t, "warlock__destruction")
	lock.GetWarlock().Options.Summon = proto.WarlockOptions_Imp
	players := []*proto.Player{lock, raidPlayer(t, "rogue__subtlety")}
	for i := 0; i < 3; i++ {
		players = append(players, raidPlayer(t, "mage__fire"))
	}
	players = append(players, raidPlayer(t, "rogue__subtlety")) // party 2
	rsr := raidOf(t, "rogue__subtlety", players, 1)
	rsr.Raid.Buffs.BloodPact = proto.TristateEffect_TristateEffectMissing
	alone := raidOf(t, "rogue__subtlety", []*proto.Player{raidPlayer(t, "rogue__subtlety")}, 1)
	alone.Raid.Buffs.BloodPact = proto.TristateEffect_TristateEffectMissing
	base := stamina(alone.Raid, 0, 0)
	if got := stamina(rsr.Raid, 0, 1); got <= base {
		t.Errorf("the warlock's party: %.0f Stamina, want more than %.0f", got, base)
	}
	if got := stamina(rsr.Raid, 1, 0); got != base {
		t.Errorf("another party: %.0f Stamina, want %.0f", got, base)
	}
}

// Bug: two warriors' Battle Shouts were one aura, so each receiver kept the first one's
// value. With 3 pieces of Battle Gear of Wrath a shout is 30 stronger.
func TestRaidForeverBattleShout(t *testing.T) {
	rsr := raidOf(t, "rogue__subtlety", []*proto.Player{raidPlayer(t, "rogue__subtlety")}, 1)
	rsr.Raid.Buffs.BattleShout = proto.TristateEffect_TristateEffectMissing
	ap := func(shouts ...bool) float64 {
		core.EnvHooks = []func(env *core.Environment){func(env *core.Environment) {
			unit := &env.Raid.Parties[0].Players[0].GetCharacter().Unit
			for _, wrath := range shouts {
				core.MakePermanent(core.BattleShoutAura(unit, 5, 0, wrath))
			}
		}}
		defer func() { core.EnvHooks = nil }()
		res := runRaid(t, googleProto.Clone(rsr).(*proto.RaidSimRequest))
		return res.RaidMetrics.Parties[0].Players[0].Dps.Avg
	}
	if strong, both := ap(true), ap(false, true); both != strong {
		t.Errorf("a shout and a stronger one: %.3f DPS, the stronger alone %.3f", both, strong)
	}
}

// Bug: a given Mana Tide (PartyBuffs.mana_tide_totems) paid every party member once for each
// member, five times in a party of 5.
func TestRaidForeverManaTide(t *testing.T) {
	var players []*proto.Player
	for i := 0; i < 5; i++ {
		players = append(players, raidPlayer(t, "mage__arcane"))
	}
	rsr := raidOf(t, "mage__arcane", players, 20)
	rsr.Raid.Parties[0].Buffs = &proto.PartyBuffs{ManaTideTotems: 1}
	res := runRaid(t, rsr)
	for i, pl := range res.RaidMetrics.Parties[0].Players {
		gain := 0.0
		for _, r := range pl.Resources {
			if r.Id.GetSpellId() == core.ManaTideTotemActionID.SpellID {
				gain += r.Gain
			}
		}
		// One totem a fight (first at 60 s, 5 min cooldown), 4 ticks of 290.
		if got := gain / 20; got != 4*290 {
			t.Errorf("member %d gained %.0f mana from Mana Tide, want %d", i, got, 4*290)
		}
	}
}

// Bug: RaidBuffs.furious_howl was never read.
func TestRaidForeverFuriousHowl(t *testing.T) {
	dps := func(shout bool, howl bool) float64 {
		rsr := raidReq(t, "warrior__fury")
		rsr.SimOptions.Iterations = 50
		if !shout {
			rsr.Raid.Buffs.BattleShout = proto.TristateEffect_TristateEffectMissing
		}
		rsr.Raid.Buffs.FuriousHowl = howl
		return runRaid(t, rsr).RaidMetrics.Dps.Avg
	}
	if none, howl := dps(false, false), dps(false, true); howl <= none {
		t.Errorf("Furious Howl: %.1f DPS, none %.1f", howl, none)
	}
	// It shares Battle Shout's slot, and the shout is stronger.
	if shout, both := dps(true, false), dps(true, true); both != shout {
		t.Errorf("Battle Shout and Furious Howl: %.3f DPS, Battle Shout alone %.3f", both, shout)
	}
}

// Bug: in a health fight the duration variation still moved the end of fight the rotations
// see (GetRemainingDuration), and drew a random number each iteration.
func TestRaidForeverHealthFightVariation(t *testing.T) {
	run := func(variation float64) []byte {
		rsr := raidOf(t, "warlock__destruction", []*proto.Player{raidPlayer(t, "warlock__destruction"), raidPlayer(t, "rogue__subtlety")}, 30)
		rsr.Encounter.UseHealth = true
		rsr.Encounter.Targets[0].Stats[stats.Health] = 3e5
		rsr.Encounter.DurationVariation = variation
		b, _ := googleProto.MarshalOptions{Deterministic: true}.Marshal(runRaid(t, rsr).RaidMetrics.Dps)
		return b
	}
	if sha256.Sum256(run(20)) != sha256.Sum256(run(0)) {
		t.Errorf("a health fight's result depends on the duration variation")
	}
}

// Bug: the last paladin to register Judgement of the Crusader took over its refresh hook,
// and the first one's bonus applied whoever judged. Two paladins on Crusader duty, one
// without Libram of Fervor (161 instead of 194): the bonus on the boss is the last judge's.
func TestRaidForeverJudgementOfTheCrusader(t *testing.T) {
	strong := raidPlayer(t, "paladin__level_60_shock-duty-crusader")
	weak := raidPlayer(t, "paladin__level_60_shock-duty-crusader")
	weak.Equipment.Items[proto.ItemSlot_ItemSlotRanged] = &proto.ItemSpec{}
	for _, order := range [][2]float64{{194, 161}, {161, 194}} {
		players := []*proto.Player{googleProto.Clone(strong).(*proto.Player), googleProto.Clone(weak).(*proto.Player)}
		if order[0] == 161 {
			players[0], players[1] = players[1], players[0]
		}
		rsr := raidOf(t, "paladin__level_60_shock-duty-crusader", players, 3)
		rsr.Raid.Debuffs.JudgementOfTheCrusader = proto.TristateEffect_TristateEffectMissing
		judged, bad := map[string]int{}, 0
		core.EnvHooks = []func(env *core.Environment){func(env *core.Environment) {
			boss := env.Encounter.TargetUnits[0]
			jotc := boss.GetAura("Judgement of the Crusader")
			bonus := map[*core.Unit]float64{}
			for i, party := range env.Raid.Parties[0].Players {
				bonus[&party.GetCharacter().Unit] = order[i]
			}
			var last *core.Unit
			core.MakePermanent(boss.RegisterAura(core.Aura{
				Label: "JotC check",
				OnSpellHitTaken: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					// The judgement's own hit comes before it puts the debuff up.
					if last != nil && jotc.IsActive() && boss.PseudoStats.SchoolBonusDamageTaken[stats.SchoolIndexHoly] != bonus[last] {
						bad++
					}
					if spell.ActionID.SpellID == 20303 {
						last = spell.Unit
						judged[spell.Unit.Label]++
					}
				},
			}))
		}}
		runRaid(t, rsr)
		core.EnvHooks = nil
		if len(judged) != 2 || bad > 0 {
			t.Errorf("judgements %v, %d hits with another paladin's bonus", judged, bad)
		}
	}
}

// B's rule: none of the fixes moves a single-player number. With RAID_FOREVER_HASHES set,
// this writes a hash of each report request's whole result (metrics sorted, first iteration's
// logs included) at its own seed and iterations, to diff against the same run on the commit
// before.
func TestRaidForeverReportHashes(t *testing.T) {
	out := os.Getenv("RAID_FOREVER_HASHES")
	if out == "" {
		t.Skip("set RAID_FOREVER_HASHES to a file to write")
	}
	var lines []string
	for _, name := range raidReqNames(t, false) {
		rsr := raidReq(t, name)
		rsr.SimOptions.DebugFirstIteration = true
		lines = append(lines, fmt.Sprintf("%s %x", name, sha256.Sum256(canonicalResult(runRaid(t, rsr)))))
	}
	if err := os.WriteFile(out, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
}

// canonicalResult marshals a result with its metric lists sorted, since actions, auras and
// resources come out in map order.
func canonicalResult(res *proto.RaidSimResult) []byte {
	res = googleProto.Clone(res).(*proto.RaidSimResult)
	mo := googleProto.MarshalOptions{Deterministic: true}
	key := func(m googleProto.Message) string { b, _ := mo.Marshal(m); return string(b) }
	var fix func(u *proto.UnitMetrics)
	fix = func(u *proto.UnitMetrics) {
		slices.SortFunc(u.Actions, func(a, b *proto.ActionMetrics) int { return strings.Compare(key(a), key(b)) })
		slices.SortFunc(u.Auras, func(a, b *proto.AuraMetrics) int { return strings.Compare(key(a), key(b)) })
		slices.SortFunc(u.Resources, func(a, b *proto.ResourceMetrics) int { return strings.Compare(key(a), key(b)) })
		for _, p := range u.Pets {
			fix(p)
		}
	}
	for _, party := range res.RaidMetrics.Parties {
		for _, player := range party.Players {
			fix(player)
		}
	}
	for _, target := range res.EncounterMetrics.GetTargets() {
		fix(target)
	}
	b, _ := mo.Marshal(res)
	return b
}
