package sim

// Raids of the DPS Report's builds (tools/dps_report/req), for the raid sim's bug fixes and its
// buffs reaching their scope (docs/raid_sim.md, milestones B and M1b). Needs the item database:
// go test -tags with_db.

import (
	"crypto/sha256"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/priest"
	"github.com/wowsims/classic/sim/shaman"
	"github.com/wowsims/classic/sim/warrior"
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
	var parties [][]*proto.Player
	for i, p := range players {
		if i%5 == 0 {
			parties = append(parties, nil)
		}
		parties[i/5] = append(parties[i/5], p)
	}
	return raidOfParties(t, base, parties, iterations)
}

// raidOfParties is raidOf with the parties given.
func raidOfParties(t testing.TB, base string, parties [][]*proto.Player, iterations int32) *proto.RaidSimRequest {
	rsr := raidReq(t, base)
	raid := &proto.Raid{Buffs: rsr.Raid.Buffs, Debuffs: rsr.Raid.Debuffs}
	for i, players := range parties {
		raid.Parties = append(raid.Parties, &proto.Party{})
		for j, p := range players {
			p.Name = fmt.Sprintf("%d", i*5+j)
			raid.Parties[i].Players = append(raid.Parties[i].Players, p)
		}
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

// idle takes the player's rotation away, so only what a test casts happens.
func idle(p *proto.Player) *proto.Player {
	p.Rotation = &proto.APLRotation{}
	return p
}

// stepRaid starts the request's sim for a test to step through with raidAt.
func stepRaid(rsr *proto.RaidSimRequest) *core.Simulation {
	sim := core.NewSim(rsr, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim
}

func raidAt(sim *core.Simulation, seconds float64, fn func(sim *core.Simulation)) {
	sim.AddPendingAction(&core.PendingAction{NextActionAt: core.DurationFromSeconds(seconds), OnAction: fn})
}

func runSteps(sim *core.Simulation) {
	for !sim.Step() {
	}
}

func member(sim *core.Simulation, party, slot int) *core.Character {
	return sim.Raid.Parties[party].Players[slot].GetCharacter()
}

// castFree puts out a spell's effects at once, without its cost, cast time or GCD.
func castFree(sim *core.Simulation, spell *core.Spell) {
	spell.ApplyEffects(sim, spell.Unit.CurrentTarget, spell)
}

func topRank(t *testing.T, ranks []*core.Spell) *core.Spell {
	t.Helper()
	for i := len(ranks) - 1; i >= 0; i-- {
		if ranks[i] != nil {
			return ranks[i]
		}
	}
	t.Fatal("no rank known")
	return nil
}

func auraUp(char *core.Unit, label string) bool {
	return char.GetAura(label).IsActive()
}

// Check 4: a party buff reaches only its own party, a raid buff every player and no pet.
// Strength of Earth Totem and Battle Shout reach the raid, and Mana Spring Totem and Dark
// Inspiration the caster's party.
func TestRaidForeverBuffScope(t *testing.T) {
	rsr := raidOfParties(t, "mage__fire", [][]*proto.Player{
		{idle(raidPlayer(t, "shaman__enhancement-level60")), raidPlayer(t, "priest__shadow"), idle(raidPlayer(t, "mage__fire"))},
		{idle(raidPlayer(t, "warrior__fury")), idle(raidPlayer(t, "mage__fire")), idle(raidPlayer(t, "warlock__destruction"))},
	}, 1)
	rsr.Raid.Buffs = &proto.RaidBuffs{}
	sim := stepRaid(rsr)
	sham := sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
	shadow := sim.Raid.Parties[0].Players[1].(priest.PriestAgent).GetPriest()
	fury := sim.Raid.Parties[1].Players[0].(warrior.WarriorAgent).GetWarrior()
	everyone := []*core.Character{member(sim, 0, 0), member(sim, 0, 1), member(sim, 0, 2), member(sim, 1, 0), member(sim, 1, 1), member(sim, 1, 2)}
	imp := member(sim, 1, 2).Pets[0]

	raidAt(sim, 1, func(sim *core.Simulation) {
		castFree(sim, topRank(t, sham.StrengthOfEarthTotem))
		castFree(sim, topRank(t, sham.ManaSpringTotem))
		castFree(sim, fury.BattleShout.Spell)
	})
	raidAt(sim, 2, func(sim *core.Simulation) {
		for i, char := range everyone {
			for _, label := range []string{"Strength of Earth Totem", "Battle Shout"} {
				if !auraUp(&char.Unit, label) {
					t.Errorf("member %d has no %s", i, label)
				}
			}
		}
		for _, label := range []string{"Strength of Earth Totem", "Battle Shout"} {
			if auraUp(&imp.Unit, label) {
				t.Errorf("the warlock's pet has %s", label)
			}
		}
		// Mana Spring reaches the shaman's party's players with mana.
		for i, want := range []bool{true, true, true, false, false, false} {
			if got := auraUp(&everyone[i].Unit, "Mana Spring Totem (Rank 4)"); got != want {
				t.Errorf("member %d has Mana Spring: %v, want %v", i, got, want)
			}
		}
	})
	// The priest's Mind Blasts and Mind Flays put Dark Inspiration on the party's mage while
	// it's on the priest, and never on the other party's.
	up := 0
	for s := 3; s < 60; s++ {
		raidAt(sim, float64(s), func(sim *core.Simulation) {
			if on := shadow.DarkInspirationAura.IsActive(); on != auraUp(&member(sim, 0, 2).Unit, "Dark Inspiration") {
				t.Errorf("at %ds: Dark Inspiration up on the priest %v, on the party's mage %v", s, on, !on)
			} else if on {
				up++
			}
			if member(sim, 1, 1).GetAura("Dark Inspiration") != nil {
				t.Errorf("the other party's mage has Dark Inspiration")
			}
		})
	}
	runSteps(sim)
	if up == 0 {
		t.Errorf("Dark Inspiration never came up")
	}
}

// Check 4: a cast Battle Shout reaches the raid, and of two warriors' shouts the strongest
// counts, whichever came first. With 3 pieces of Battle Gear of Wrath a shout is 30 stronger.
// The rogue is in neither warrior's party.
func TestRaidForeverCastBattleShout(t *testing.T) {
	wrath := raidPlayer(t, "warrior__fury")
	for slot, id := range map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotHead: 16963, proto.ItemSlot_ItemSlotShoulder: 16961, proto.ItemSlot_ItemSlotChest: 16966} {
		wrath.Equipment.Items[slot] = &proto.ItemSpec{Id: id}
	}
	gains := func(strongFirst bool) (first, second float64) {
		rsr := raidOfParties(t, "rogue__subtlety", [][]*proto.Player{
			{idle(raidPlayer(t, "warrior__fury"))}, {idle(googleProto.Clone(wrath).(*proto.Player))}, {idle(raidPlayer(t, "rogue__subtlety"))},
		}, 1)
		rsr.Raid.Buffs.BattleShout = proto.TristateEffect_TristateEffectMissing
		sim := stepRaid(rsr)
		shouts := []*core.Spell{
			sim.Raid.Parties[0].Players[0].(warrior.WarriorAgent).GetWarrior().BattleShout.Spell,
			sim.Raid.Parties[1].Players[0].(warrior.WarriorAgent).GetWarrior().BattleShout.Spell,
		}
		if strongFirst {
			shouts[0], shouts[1] = shouts[1], shouts[0]
		}
		rogue := member(sim, 2, 0)
		var base float64
		raidAt(sim, 1, func(sim *core.Simulation) {
			base = rogue.GetStat(stats.AttackPower)
			castFree(sim, shouts[0])
			first = rogue.GetStat(stats.AttackPower) - base
		})
		raidAt(sim, 2, func(sim *core.Simulation) {
			castFree(sim, shouts[1])
			second = rogue.GetStat(stats.AttackPower) - base
		})
		runSteps(sim)
		return first, second
	}
	weak, both := gains(false)
	strong, bothReversed := gains(true)
	if weak <= 0 || strong <= weak {
		t.Errorf("a shout gives another party %.1f attack power, a Wrath shout %.1f", weak, strong)
	}
	if both != strong || bothReversed != strong {
		t.Errorf("both shouts give %.1f attack power (%.1f the other way round), the stronger %.1f", both, bothReversed, strong)
	}
}

// Check 4: a dropped totem's buff stays while a second shaman's identical totem still stands,
// on the receivers and on the first shaman. It ends with the last of them.
func TestRaidForeverTotemHolds(t *testing.T) {
	for _, firstEndsFirst := range []bool{true, false} {
		rsr := raidOfParties(t, "mage__fire", [][]*proto.Player{
			{idle(raidPlayer(t, "shaman__enhancement-level60"))}, {idle(raidPlayer(t, "shaman__enhancement-level60"))}, {idle(raidPlayer(t, "mage__fire"))},
		}, 1)
		rsr.Raid.Buffs = &proto.RaidBuffs{}
		sim := stepRaid(rsr)
		shamans := []*shaman.Shaman{
			sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman(),
			sim.Raid.Parties[1].Players[0].(shaman.ShamanAgent).GetShaman(),
		}
		mage := member(sim, 2, 0)
		ends := []int{0, 1}
		if !firstEndsFirst {
			ends = []int{1, 0}
		}
		const soe, goa = "Strength of Earth Totem", "Grace of Air Totem"
		check := func(when string, units []*core.Unit, label string, want bool) {
			for _, unit := range units {
				if got := auraUp(unit, label); got != want {
					t.Errorf("%s: %s up on %s: %v, want %v", when, label, unit.Label, got, want)
				}
			}
		}
		raidAt(sim, 1, func(sim *core.Simulation) {
			castFree(sim, topRank(t, shamans[0].StrengthOfEarthTotem))
			castFree(sim, topRank(t, shamans[0].GraceOfAirTotem))
		})
		raidAt(sim, 2, func(sim *core.Simulation) {
			castFree(sim, topRank(t, shamans[1].StrengthOfEarthTotem))
			castFree(sim, topRank(t, shamans[1].GraceOfAirTotem))
		})
		raidAt(sim, 3, func(sim *core.Simulation) {
			// Stoneskin and Windwall take the slots of the first shaman to end theirs.
			castFree(sim, topRank(t, shamans[ends[0]].StoneskinTotem))
			castFree(sim, topRank(t, shamans[ends[0]].WindwallTotem))
			everyone := []*core.Unit{&shamans[0].Unit, &shamans[1].Unit, &mage.Unit}
			check("one totem left", everyone, soe, true)
			check("one totem left", everyone, goa, true)
			// The buff lasts as long as the totem left standing.
			want := core.DurationFromSeconds(float64(2-ends[0])) + 5*time.Minute
			if got := mage.GetAura(soe).ExpiresAt(); got != want {
				t.Errorf("the buff runs out at %s, want %s", got, want)
			}
		})
		raidAt(sim, 4, func(sim *core.Simulation) {
			castFree(sim, topRank(t, shamans[ends[1]].StoneskinTotem))
			castFree(sim, topRank(t, shamans[ends[1]].WindwallTotem))
			everyone := []*core.Unit{&shamans[0].Unit, &shamans[1].Unit, &mage.Unit}
			check("both totems gone", everyone, soe, false)
			check("both totems gone", everyone, goa, false)
		})
		runSteps(sim)
	}
}

// A wolf's cast Furious Howl reaches the raid's players in Battle Shout's slot, but not
// another hunter's pet.
func TestRaidForeverCastFuriousHowl(t *testing.T) {
	wolf := idle(raidPlayer(t, "hunter__beast-mastery"))
	wolf.GetHunter().Options.PetType = proto.Hunter_Options_Wolf
	rsr := raidOfParties(t, "rogue__subtlety", [][]*proto.Player{
		{wolf}, {idle(raidPlayer(t, "rogue__subtlety")), idle(raidPlayer(t, "hunter__beast-mastery"))},
	}, 1)
	rsr.Raid.Buffs.BattleShout = proto.TristateEffect_TristateEffectMissing
	sim := stepRaid(rsr)
	var howl *core.Spell
	for _, pet := range member(sim, 0, 0).Pets {
		if spell := pet.GetSpell(core.ActionID{SpellID: 24597}); spell != nil {
			howl = spell
		}
	}
	if howl == nil {
		t.Fatal("the wolf has no Furious Howl")
	}
	rogue := member(sim, 1, 0)
	raidAt(sim, 1, func(sim *core.Simulation) {
		ap := rogue.GetStat(stats.AttackPower)
		castFree(sim, howl)
		if got := rogue.GetStat(stats.AttackPower) - ap; got != 83 {
			t.Errorf("another party's rogue gains %.1f attack power from Furious Howl, want 83", got)
		}
		for _, pet := range member(sim, 1, 1).Pets {
			if auraUp(&pet.Unit, "Furious Howl") {
				t.Errorf("another hunter's pet has Furious Howl")
			}
		}
	})
	runSteps(sim)
}

// The weapon totem per receiver: a receiver takes one of Windfury and Flametongue Totem.
// Windfury Totem does nothing for the Enhancement shaman's Windfury Weapon, so that one takes
// Flametongue. Everyone else (a rogue, and a mage it does nothing for) takes Windfury,
// whichever came down last, and an Elemental shaman its own. When Windfury Totem goes, the
// rogue takes Flametongue.
func TestRaidForeverWeaponTotems(t *testing.T) {
	const windfury, flametongue = "Windfury Totem (Rank 3)", "Flametongue Totem (Rank 4)"
	for _, windfuryFirst := range []bool{true, false} {
		rsr := raidOfParties(t, "rogue__combat-swords", [][]*proto.Player{
			{idle(raidPlayer(t, "shaman__enhancement-level60"))},
			{idle(raidPlayer(t, "shaman__elemental-level60"))},
			{idle(raidPlayer(t, "rogue__combat-swords")), idle(raidPlayer(t, "mage__fire"))},
		}, 1)
		rsr.Raid.Buffs = &proto.RaidBuffs{}
		sim := stepRaid(rsr)
		enh := sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
		ele := sim.Raid.Parties[1].Players[0].(shaman.ShamanAgent).GetShaman()
		rogue := member(sim, 2, 0)
		totems := []*core.Spell{topRank(t, ele.WindfuryTotem), topRank(t, enh.FlametongueTotem)}
		if !windfuryFirst {
			totems[0], totems[1] = totems[1], totems[0]
		}
		holds := func(when string, unit *core.Unit, want string) {
			for _, label := range []string{windfury, flametongue} {
				if got := auraUp(unit, label); got != (label == want) {
					t.Errorf("windfury first %v, %s: %s up on %s: %v", windfuryFirst, when, label, unit.Label, got)
				}
			}
		}
		raidAt(sim, 1, func(sim *core.Simulation) { castFree(sim, totems[0]) })
		raidAt(sim, 2, func(sim *core.Simulation) {
			castFree(sim, totems[1])
			holds("both totems", &rogue.Unit, windfury)
			holds("both totems", &member(sim, 2, 1).Unit, windfury)
			holds("both totems", &enh.Unit, flametongue)
			holds("both totems", &ele.Unit, windfury)
		})
		raidAt(sim, 3, func(sim *core.Simulation) {
			castFree(sim, topRank(t, ele.GraceOfAirTotem))
			holds("Flametongue left", &rogue.Unit, flametongue)
			holds("Flametongue left", &ele.Unit, flametongue)
		})
		runSteps(sim)
	}
}

// Player.raid_buffs replaces the raid's buffs whole, and the player's class hook doesn't
// run: a mage with its own empty set has no Arcane Brilliance, and a Balance druid with its own
// set gives the raid no Moonkin Aura.
func TestRaidForeverOwnRaidBuffs(t *testing.T) {
	encounter := raidReq(t, "mage__fire").Encounter
	final := func(raid *proto.Raid, party, slot int) stats.Stats {
		res := core.ComputeStats(&proto.ComputeStatsRequest{Raid: raid, Encounter: encounter})
		return stats.FromFloatArray(res.RaidStats.Parties[party].Players[slot].FinalStats.Stats)
	}
	raidWith := func(raidBuffs *proto.RaidBuffs, players ...*proto.Player) *proto.Raid {
		rsr := raidOf(t, "mage__fire", players, 1)
		rsr.Raid.Buffs = raidBuffs
		return rsr.Raid
	}
	mageWith := func(own *proto.RaidBuffs) *proto.Player {
		p := raidPlayer(t, "mage__fire")
		p.RaidBuffs = own
		return p
	}
	full := raidReq(t, "mage__fire").Raid.Buffs

	// Replaced whole: the raid's set doesn't matter.
	own := &proto.RaidBuffs{ArcaneBrilliance: true, GiftOfTheWild: proto.TristateEffect_TristateEffectRegular}
	if got, want := final(raidWith(full, mageWith(own), raidPlayer(t, "rogue__subtlety")), 0, 0), final(raidWith(own, raidPlayer(t, "mage__fire")), 0, 0); got != want {
		t.Errorf("a mage with its own buffs in a buffed raid: %v, alone with them as the raid's: %v", got, want)
	}
	// No hook: the mage's own Arcane Brilliance isn't there.
	if got, want := final(raidWith(full, mageWith(&proto.RaidBuffs{})), 0, 0)[stats.Intellect], final(raidWith(&proto.RaidBuffs{}, mageWith(&proto.RaidBuffs{ArcaneBrilliance: true})), 0, 0)[stats.Intellect]; got >= want {
		t.Errorf("a mage with its own empty buffs has %.0f Intellect, with Arcane Brilliance %.0f", got, want)
	}
	// The druid's hook gives Moonkin Aura to the raid, but not with its own set.
	crit := func(druidOwn *proto.RaidBuffs) float64 {
		druid := raidPlayer(t, "druid__balance")
		druid.RaidBuffs = druidOwn
		return final(raidWith(&proto.RaidBuffs{}, druid, raidPlayer(t, "mage__fire")), 0, 1)[stats.SpellCrit]
	}
	alone := final(raidWith(&proto.RaidBuffs{}, raidPlayer(t, "mage__fire")), 0, 0)[stats.SpellCrit]
	if hooked, owned := crit(nil), crit(&proto.RaidBuffs{}); hooked <= alone || owned != alone {
		t.Errorf("the mage's spell crit next to a Balance druid: %.2f, next to one with its own buffs %.2f, alone %.2f", hooked, owned, alone)
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
