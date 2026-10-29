package hunter

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// The raid sim's hunter pieces (docs/raid_sim.md, Hunter pets and Assignments): the given Dust
// Cloud, and a hunter that leaves Hunter's Mark to another.

func hunterRaidRequest(options *proto.Hunter_Options, rotation *proto.APLRotation, debuffs *proto.Debuffs, ruleset proto.Ruleset) *proto.RaidSimRequest {
	player := &proto.Player{
		Class:     proto.Class_ClassHunter,
		Race:      proto.Race_RaceHuman,
		Level:     60,
		Equipment: &proto.EquipmentSpec{},
		Rotation:  rotation,
		Spec:      &proto.Player_Hunter{Hunter: &proto.Hunter{Options: options}},
	}
	return &proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, debuffs),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: ruleset, RandomSeed: 1, Iterations: 20},
	}
}

func newHunterRaidSim(t *testing.T, rsr *proto.RaidSimRequest) (*core.Simulation, *Hunter) {
	t.Helper()
	sim := core.NewSim(rsr, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(HunterAgent).GetHunter()
}

// casts is how many times a unit's metrics say it cast id, a run's average.
func casts(unit *proto.UnitMetrics, id int32) float64 {
	n := 0
	for _, action := range unit.Actions {
		if action.Id.GetSpellId() == id {
			for _, target := range action.Targets {
				n += int(target.Casts)
			}
		}
	}
	return float64(n)
}

func runHunterRaid(t *testing.T, rsr *proto.RaidSimRequest) *proto.UnitMetrics {
	t.Helper()
	res := core.RunRaidSim(rsr)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	return res.RaidMetrics.Parties[0].Players[0]
}

// Debuffs.dust_cloud is a Tallstrider's Dust Cloud for the whole fight: 505 armor at 60 (1265904), in
// the minor armor slot with Faerie Fire and Curse of Recklessness, so only the strongest counts.
// Classic's Tallstriders have no Dust Cloud.
func TestForeverGivenDustCloud(t *testing.T) {
	armor := func(debuffs *proto.Debuffs, ruleset proto.Ruleset) (float64, *core.Unit) {
		sim, _ := newHunterRaidSim(t, hunterRaidRequest(&proto.Hunter_Options{}, &proto.APLRotation{}, debuffs, ruleset))
		sim.PrePull()
		target := sim.Encounter.TargetUnits[0]
		return target.GetStat(stats.Armor), target
	}
	none, _ := armor(&proto.Debuffs{}, proto.Ruleset_RulesetForever)
	cloud, target := armor(&proto.Debuffs{DustCloud: true}, proto.Ruleset_RulesetForever)
	if got := none - cloud; got != 505 {
		t.Errorf("the given Dust Cloud takes %.0f armor, want 505", got)
	}
	if aura := target.GetAuraByID(core.ActionID{SpellID: 1265904}); aura == nil || !aura.IsActive() || aura.Duration != core.NeverExpires {
		t.Error("the given Dust Cloud isn't up for the whole fight")
	}
	both, _ := armor(&proto.Debuffs{DustCloud: true, FaerieFire: true, CurseOfRecklessness: true}, proto.Ruleset_RulesetForever)
	if got := none - both; got != 505 {
		t.Errorf("Dust Cloud, Faerie Fire and Curse of Recklessness take %.0f armor together, want 505 (one slot)", got)
	}
	classicNone, _ := armor(&proto.Debuffs{}, proto.Ruleset_RulesetClassic)
	classicCloud, _ := armor(&proto.Debuffs{DustCloud: true}, proto.Ruleset_RulesetClassic)
	if classicNone != classicCloud {
		t.Errorf("the Classic Dust Cloud takes %.0f armor, want none", classicNone-classicCloud)
	}
}

// A Tallstrider keeps its own Dust Cloud up when the minor armor slot is free, and leaves it to the
// raid's (the given Dust Cloud or Faerie Fire) when it isn't.
func TestForeverTallstriderDustCloud(t *testing.T) {
	tallstrider := &proto.Hunter_Options{PetType: proto.Hunter_Options_Tallstrider, PetUptime: 1}
	for _, c := range []struct {
		name    string
		debuffs *proto.Debuffs
		cast    bool
	}{
		{"no debuffs", &proto.Debuffs{}, true},
		{"the given Dust Cloud", &proto.Debuffs{DustCloud: true}, false},
		{"Faerie Fire", &proto.Debuffs{FaerieFire: true}, false},
	} {
		player := runHunterRaid(t, hunterRaidRequest(tallstrider, &proto.APLRotation{}, c.debuffs, proto.Ruleset_RulesetForever))
		if len(player.Pets) != 1 {
			t.Fatalf("%s: %d pets, want the Tallstrider", c.name, len(player.Pets))
		}
		if got := casts(player.Pets[0], 1265904); (got > 0) != c.cast {
			t.Errorf("%s: the Tallstrider cast Dust Cloud %.0f times in 20 fights, want any: %v", c.name, got, c.cast)
		}
	}
}

// With skip_hunters_mark another hunter keeps up the mark, so this one never casts it, whatever its
// rotation says. Without it the rotation's mark goes out as before.
func TestForeverSkipHuntersMark(t *testing.T) {
	mark := &proto.APLRotation{PriorityList: []*proto.APLListItem{{Action: &proto.APLAction{Action: &proto.APLAction_CastSpell{
		CastSpell: &proto.APLActionCastSpell{SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 14325}}},
	}}}}}
	marks := casts(runHunterRaid(t, hunterRaidRequest(&proto.Hunter_Options{}, mark, &proto.Debuffs{}, proto.Ruleset_RulesetForever)), 14325)
	if marks == 0 {
		t.Fatal("the hunter never cast Hunter's Mark")
	}
	skip := &proto.Hunter_Options{SkipHuntersMark: true}
	if got := casts(runHunterRaid(t, hunterRaidRequest(skip, mark, &proto.Debuffs{}, proto.Ruleset_RulesetForever)), 14325); got != 0 {
		t.Errorf("with skip_hunters_mark the hunter cast Hunter's Mark %.0f times, want none", got)
	}

	// The other hunter's mark is the given one, and it still counts for this hunter.
	sim, hunter := newHunterRaidSim(t, hunterRaidRequest(skip, mark, &proto.Debuffs{HuntersMark: proto.TristateEffect_TristateEffectRegular}, proto.Ruleset_RulesetForever))
	sim.PrePull()
	if hunter.HuntersMark.CanCast(sim, hunter.CurrentTarget) {
		t.Error("with skip_hunters_mark Hunter's Mark can be cast")
	}
	if aura := hunter.CurrentTarget.GetAuraByID(core.ActionID{SpellID: 14325}); aura == nil || !aura.IsActive() {
		t.Error("the given Hunter's Mark isn't on the target")
	}
}
