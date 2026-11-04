package tankscore

import (
	"math"
	"os"
	"testing"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

var tankRequests = []string{
	"../tankbench/req/shaman_warden_balanced.json",
	"../tankbench/req/druid_bear_balanced.json",
	"../tankbench/req/paladin_protection_balanced.json",
	"../tankbench/req/warrior_protection_balanced.json",
}

func init() {
	sim.RegisterAll()
}

func scoredRequest(t *testing.T, path string, hps float64) *proto.RaidSimRequest {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	rsr.SimOptions.Iterations = 20
	rsr.SimOptions.RandomSeed = 7
	MakeTank(rsr, &proto.HealingModel{Hps: hps, CadenceSeconds: 2, CadenceVariation: 1}, &proto.TankScoring{})
	return rsr
}

func runScored(t *testing.T, rsr *proto.RaidSimRequest) *proto.UnitMetrics {
	t.Helper()
	res := core.RunRaidSim(rsr)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	return res.RaidMetrics.Parties[0].Players[0]
}

// A scored tank faces the boss, whatever its request says: the boss parries and blocks its attacks
// (tank_mechanics-20261104-1).
func TestMakeTankInFront(t *testing.T) {
	rsr := scoredRequest(t, "../tankbench/req/warrior_protection_balanced.json", 600)
	rsr.Raid.Parties[0].Players[0].InFrontOfTarget = false
	MakeTank(rsr, rsr.Raid.Parties[0].Players[0].HealingModel, &proto.TankScoring{})
	if !rsr.Raid.Parties[0].Players[0].InFrontOfTarget {
		t.Fatal("MakeTank left the tank behind the boss")
	}
	var parries, blocks int32
	for _, action := range runScored(t, rsr).Actions {
		for _, target := range action.Targets {
			parries += target.Parries
			blocks += target.Blocks
		}
	}
	if parries == 0 {
		t.Errorf("the boss parried none of the scored tank's attacks")
	}
	t.Logf("boss parried %d and blocked %d of the tank's attacks", parries, blocks)
}

// Nothing in a scoring run reads the tank's health, so its threat and damage taken don't depend on
// the stand-in healer's HPS, even with a survival cooldown (Greater Stoneshield Potion) set to fire
// under 90% health (tank_mechanics-20261104-8). The warrior pays health for Bloodrage and casts Shield
// Wall on a timer, neither of which may read its health either.
func TestScoringRunReadsNoHealth(t *testing.T) {
	for _, path := range tankRequests {
		var dtps, safe []float64
		for _, hps := range []float64{200, 5000} {
			rsr := scoredRequest(t, path, hps)
			tank := rsr.Raid.Parties[0].Players[0]
			tank.Cooldowns = &proto.Cooldowns{HpPercentForDefensives: 0.9}
			tank.Consumes.DefaultPotion = proto.Potions_GreaterStoneshieldPotion
			// The warrior's APL drinks its potion at the pull and has no Autocast Other Cooldowns, so
			// no major cooldown of its fires on its own: drop the potion line and add the autocast,
			// and the potion is due at the health share as it is for the other tanks.
			autocast := false
			list := tank.Rotation.PriorityList[:0]
			for _, item := range tank.Rotation.PriorityList {
				autocast = autocast || item.GetAction().GetAutocastOtherCooldowns() != nil
				if item.GetAction().GetCastSpell().GetSpellId().GetOtherId() != proto.OtherAction_OtherActionPotion {
					list = append(list, item)
				}
			}
			if !autocast {
				list = append([]*proto.APLListItem{{Action: &proto.APLAction{
					Action: &proto.APLAction_AutocastOtherCooldowns{AutocastOtherCooldowns: &proto.APLActionAutocastOtherCooldowns{}},
				}}}, list...)
			}
			tank.Rotation.PriorityList = list
			m := runScored(t, rsr)
			dtps = append(dtps, m.Dtps.Avg)
			safe = append(safe, m.TankScore.SafeDps.Avg)
		}
		if dtps[0] != dtps[1] || safe[0] != safe[1] {
			t.Errorf("%s: DTPS %v and SafeDPS %v move with the HPS", path, dtps, safe)
		}
	}
}

// An APL condition on the tank's health reads it as at full health in a scoring run, as a survival
// cooldown's health share does: a rotation that stops for a second whenever the tank is hurt keeps
// the same threat and damage taken at any stand-in HPS (tank_mechanics-20261104-14).
func TestScoringRunAPLReadsNoHealth(t *testing.T) {
	conditions := map[string]string{
		"Current Health %": `{"cmp": {"op": "OpLt", "lhs": {"currentHealthPercent": {}}, "rhs": {"const": {"val": "90%"}}}}`,
		"Current Health":   `{"cmp": {"op": "OpLt", "lhs": {"currentHealth": {}}, "rhs": {"maxHealth": {}}}}`,
		"Health Missing":   `{"cmp": {"op": "OpGt", "lhs": {"healthMissing": {}}, "rhs": {"const": {"val": "0"}}}}`,
	}
	for name, condition := range conditions {
		item := &proto.APLListItem{}
		if err := protojson.Unmarshal([]byte(`{"action": {"condition": `+condition+
			`, "wait": {"duration": {"const": {"val": "1s"}}}}}`), item); err != nil {
			t.Fatal(err)
		}
		for _, path := range tankRequests {
			var dtps, safe []float64
			for _, hps := range []float64{200, 5000} {
				rsr := scoredRequest(t, path, hps)
				rotation := rsr.Raid.Parties[0].Players[0].Rotation
				rotation.PriorityList = append([]*proto.APLListItem{item}, rotation.PriorityList...)
				m := runScored(t, rsr)
				dtps = append(dtps, m.Dtps.Avg)
				safe = append(safe, m.TankScore.SafeDps.Avg)
			}
			if dtps[0] != dtps[1] || safe[0] != safe[1] {
				t.Errorf("%s, %s: DTPS %v and SafeDPS %v move with the HPS", name, path, dtps, safe)
			}
		}
	}
}

// A scored tank's attacks the boss parries haste the boss's next swing when the encounter says so:
// with parryHaste on the boss swings more often and the tank takes more damage than with it off,
// the same seed otherwise (tank_mechanics-20261104-11).
func TestMakeTankParryHaste(t *testing.T) {
	for _, path := range []string{
		"../tankbench/req/warrior_protection_balanced.json",
		"../tankbench/req/druid_bear_balanced.json",
	} {
		var dtps []float64
		var swings []int32
		for _, haste := range []bool{false, true} {
			rsr := scoredRequest(t, path, 600)
			for _, target := range rsr.Encounter.Targets {
				target.ParryHaste = haste
			}
			res := core.RunRaidSim(rsr)
			if res.Error != nil {
				t.Fatal(res.Error.Message)
			}
			var n int32
			for _, action := range res.EncounterMetrics.Targets[0].Actions {
				if action.Id.GetOtherId() == proto.OtherAction_OtherActionAttack {
					for _, target := range action.Targets {
						n += target.Casts
					}
				}
			}
			dtps = append(dtps, res.RaidMetrics.Parties[0].Players[0].Dtps.Avg)
			swings = append(swings, n)
		}
		if swings[1] <= swings[0] || dtps[1] <= dtps[0] {
			t.Errorf("%s: parry haste off/on: boss swings %v, DTPS %v; on should be more", path, swings, dtps)
		}
		t.Logf("%s: parry haste off/on: boss swings %v, DTPS %v", path, swings, dtps)
	}
}

// OwnDebuffs takes off the raid's debuffs a warrior or bear tank keeps up itself, and only those.
func TestOwnDebuffs(t *testing.T) {
	full := func() *proto.Debuffs {
		return &proto.Debuffs{
			ThunderClap:       proto.TristateEffect_TristateEffectImproved,
			DemoralizingShout: proto.TristateEffect_TristateEffectImproved,
			DemoralizingRoar:  proto.TristateEffect_TristateEffectImproved,
			SunderRamp:        &proto.SunderRamp{PullSunders: 2, Tank: true},
			FaerieFire:        true,
		}
	}
	request := func(spec any) *proto.RaidSimRequest {
		player := &proto.Player{}
		switch s := spec.(type) {
		case *proto.Player_TankWarrior:
			player.Spec = s
		case *proto.Player_FeralTankDruid:
			player.Spec = s
		case *proto.Player_ProtectionPaladin:
			player.Spec = s
		}
		return &proto.RaidSimRequest{Raid: &proto.Raid{
			Parties: []*proto.Party{{Players: []*proto.Player{player}}},
			Debuffs: full(),
		}}
	}
	missing := proto.TristateEffect_TristateEffectMissing

	rsr := request(&proto.Player_TankWarrior{})
	OwnDebuffs(rsr)
	d := rsr.Raid.Debuffs
	if d.ThunderClap != missing || d.DemoralizingShout != missing || d.DemoralizingRoar != missing {
		t.Errorf("warrior: Thunder Clap %v, Demoralizing Shout %v, Roar %v, want all missing", d.ThunderClap, d.DemoralizingShout, d.DemoralizingRoar)
	}
	if d.SunderRamp.Tank || d.SunderRamp.PullSunders != 2 {
		t.Errorf("warrior: Sunder ramp %v, want the tank's ramp off and the 2 pull Sunders kept", d.SunderRamp)
	}
	if !d.FaerieFire {
		t.Error("warrior: Faerie Fire went, want it kept")
	}

	rsr = request(&proto.Player_FeralTankDruid{})
	OwnDebuffs(rsr)
	d = rsr.Raid.Debuffs
	if d.DemoralizingShout != missing || d.DemoralizingRoar != missing {
		t.Errorf("bear: Demoralizing Shout %v, Roar %v, want both missing", d.DemoralizingShout, d.DemoralizingRoar)
	}
	if d.ThunderClap == missing || !d.SunderRamp.Tank || !d.FaerieFire {
		t.Errorf("bear: took off Thunder Clap, the Sunder ramp or Faerie Fire: %v", d)
	}

	rsr = request(&proto.Player_ProtectionPaladin{})
	OwnDebuffs(rsr)
	if !googleProto.Equal(rsr.Raid.Debuffs, full()) {
		t.Errorf("paladin: debuffs changed to %v, want them as they were", rsr.Raid.Debuffs)
	}

	rsr = request(&proto.Player_TankWarrior{})
	rsr.Raid.Debuffs = nil
	OwnDebuffs(rsr) // no debuffs: nothing to take off, and no panic
}

// MakeTank makes the first player the tank of every boss, healed by a copy of the stand-in
// stream (an HPS of 0 kept just above 0, so the engine doesn't put in its presim healer), scored,
// and keeping its own debuffs.
func TestMakeTankSetup(t *testing.T) {
	rsr := &proto.RaidSimRequest{
		Raid: &proto.Raid{
			Parties: []*proto.Party{{Players: []*proto.Player{{Spec: &proto.Player_TankWarrior{}}, {}}}},
			Debuffs: &proto.Debuffs{ThunderClap: proto.TristateEffect_TristateEffectRegular},
		},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{TankIndex: 3}, {TankIndex: 2}}},
	}
	healing := &proto.HealingModel{Hps: 0, CadenceSeconds: 2}
	scoring := &proto.TankScoring{DpsStart: 4}
	MakeTank(rsr, healing, scoring)

	if len(rsr.Raid.Tanks) != 1 || rsr.Raid.Tanks[0].Type != proto.UnitReference_Player || rsr.Raid.Tanks[0].Index != 0 {
		t.Errorf("tanks %v, want the first player alone", rsr.Raid.Tanks)
	}
	for i, target := range rsr.Encounter.Targets {
		if target.TankIndex != 0 {
			t.Errorf("boss %d attacks tank %d, want 0", i, target.TankIndex)
		}
	}
	tank := rsr.Raid.Parties[0].Players[0]
	if !tank.InFrontOfTarget || tank.TankScoring != scoring {
		t.Errorf("tank in front %v, scoring %v: want in front and scored", tank.InFrontOfTarget, tank.TankScoring)
	}
	if tank.HealingModel == healing || tank.HealingModel.CadenceSeconds != 2 || tank.HealingModel.Hps <= 0 || tank.HealingModel.Hps > 1e-6 {
		t.Errorf("healing model %v, want a copy with an HPS just above 0", tank.HealingModel)
	}
	if healing.Hps != 0 {
		t.Error("MakeTank changed the caller's healing model")
	}
	if rsr.Raid.Debuffs.ThunderClap != proto.TristateEffect_TristateEffectMissing {
		t.Error("MakeTank left the warrior's Thunder Clap to the raid")
	}
	if rsr.Raid.Parties[0].Players[1].HealingModel != nil || rsr.Raid.Parties[0].Players[1].InFrontOfTarget {
		t.Error("MakeTank touched the second player")
	}
}

// Score is threat members × SafeDPS less w × DPS slot / one healer's HPS × healing needed, with the
// two errors added in quadrature; a tank no healing saves gets Rejected, worse the more it dies.
func TestScore(t *testing.T) {
	o := Objective{Weight: 1, ThreatMembers: 10, DpsSlot: 1000, HealerEach: 500, Death: 0.05}
	hps := make([]float64, 100)
	for i := range hps {
		hps[i] = 300
	}
	m := &proto.UnitMetrics{TankScore: &proto.TankScore{SurviveHps: hps, SafeDps: &proto.DistributionMetrics{Avg: 800, Stdev: 50}}}
	s := o.Score(m, 100)
	if want := 10*800 - 1000.0/500*300; math.Abs(s.Value-want) > 1e-9 {
		t.Errorf("score %.3f, want %.3f", s.Value, want)
	}
	if math.Abs(s.SafeDpsErr-5) > 1e-9 || math.Abs(s.StdErr-50) > 1e-9 {
		t.Errorf("SafeDPS error %.3f and score error %.3f, want 5 and 50 (no spread in healing needed)", s.SafeDpsErr, s.StdErr)
	}
	o.Weight = 3
	if got := o.Score(m, 100).Value; math.Abs(got-(8000-3*600)) > 1e-9 {
		t.Errorf("at w = 3 the score is %.3f, want %.3f", got, 8000.0-3*600)
	}

	// 10% of the iterations die whatever the healing: rejected, and lower at 20%.
	for i := range 10 {
		hps[i] = math.Inf(1)
	}
	at10 := o.Score(m, 100)
	for i := range 20 {
		hps[i] = math.Inf(1)
	}
	at20 := o.Score(m, 100)
	if at10.Value >= Rejected || at20.Value >= at10.Value || at10.Needed >= 0 {
		t.Errorf("rejected scores %.0f at a 10%% death floor, %.0f at 20%%: want under Rejected and falling", at10.Value, at20.Value)
	}

	if s := o.Score(&proto.UnitMetrics{}, 100); !math.IsNaN(s.Value) {
		t.Errorf("no tank score gives %v, want NaN", s.Value)
	}
}

// PairedDiff is the score of b less a's; a shift every iteration shares has no error, and runs that
// don't pair fall back.
func TestPairedDiff(t *testing.T) {
	o := Objective{Weight: 1, ThreatMembers: 10, DpsSlot: 1000, HealerEach: 500, Death: 0.05}
	n := 200
	a := &proto.TankScore{SafeDpsEach: make([]float64, n), SurviveHps: make([]float64, n)}
	b := &proto.TankScore{SafeDpsEach: make([]float64, n), SurviveHps: make([]float64, n)}
	for i := range n {
		a.SafeDpsEach[i] = 500 + float64(i%17)*20
		b.SafeDpsEach[i] = a.SafeDpsEach[i] + 30
		a.SurviveHps[i] = 300
		b.SurviveHps[i] = 300
	}
	diff, stdErr, ok := o.PairedDiff(a, b)
	if !ok || math.Abs(diff-300) > 1e-9 || stdErr > 1e-9 {
		t.Errorf("a 30 SafeDPS shift: diff %.3f ± %.3f (ok %v), want 300 ± 0", diff, stdErr, ok)
	}
	// Against the unpaired errors, which see the iterations' spread.
	sa := Objective.Score(o, &proto.UnitMetrics{TankScore: &proto.TankScore{SurviveHps: a.SurviveHps, SafeDps: &proto.DistributionMetrics{Stdev: 100}}}, int32(n))
	if stdErr >= sa.StdErr {
		t.Errorf("paired error %.3f not under one run's %.3f", stdErr, sa.StdErr)
	}

	// b's healing needed spreads over its iterations: its 95th percentile moves from one resample
	// to the next, so the difference has an error now.
	for i := range n {
		b.SurviveHps[i] = 300 + float64(i)
	}
	diff, stdErr, ok = o.PairedDiff(a, b)
	if !ok || stdErr <= 0 {
		t.Errorf("diff %.3f ± %.3f (ok %v), want an error above 0", diff, stdErr, ok)
	}

	if _, _, ok := o.PairedDiff(a, &proto.TankScore{SafeDpsEach: b.SafeDpsEach[:n-1], SurviveHps: b.SurviveHps[:n-1]}); ok {
		t.Error("runs of other iteration counts paired")
	}
	dead := &proto.TankScore{SafeDpsEach: b.SafeDpsEach, SurviveHps: make([]float64, n)}
	for i := range n {
		dead.SurviveHps[i] = math.Inf(1)
	}
	if _, _, ok := o.PairedDiff(a, dead); ok {
		t.Error("a rejected run paired")
	}
}
