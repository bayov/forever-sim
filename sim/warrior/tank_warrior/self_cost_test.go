package tankwarrior

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	googleProto "google.golang.org/protobuf/proto"
)

// selfCostRun sims the tanked warrior with the stand-in healer and the tank score, casting only the
// given spells, so the boss's damage is the same whatever they are (none of them rolls anything).
func selfCostRun(t *testing.T, talents string, spells ...int32) *proto.UnitMetrics {
	t.Helper()
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 991014}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: bulwark}
	rotation := &proto.APLRotation{}
	for _, id := range spells {
		rotation.PriorityList = append(rotation.PriorityList, &proto.APLListItem{Action: &proto.APLAction{
			Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
				SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}},
			}},
		}})
	}
	player := &proto.Player{
		Race: proto.Race_RaceOrc, Class: proto.Class_ClassWarrior, Level: 60,
		Equipment: &proto.EquipmentSpec{Items: items}, TalentsString: talents,
		Rotation: rotation, Spec: defensiveOptions,
		HealingModel: &proto.HealingModel{Hps: 800, CadenceSeconds: 2, BurstWindow: 6},
		TankScoring:  &proto.TankScoring{},
	}
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
	target.MinBaseDamage = 300
	res := core.RunRaidSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Duration: 120, Targets: []*proto.Target{target}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1, Iterations: 50},
	})
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	return res.RaidMetrics.Parties[0].Players[0]
}

// TestForeverTankSelfCosts checks that the health Bloodrage and Reckless Swings cost a tank is
// healing the tank needs (the tank score's replay) but neither DTPS nor TMI, which measure the
// boss's damage: casting them leaves both as they are and raises the healing needed. Their base
// health is our class base health at 60 (1689 for an Orc warrior, the emulator convention; to
// verify against a level 60 Bloodrage in a combat log, docs/tank_warrior.md).
func TestForeverTankSelfCosts(t *testing.T) {
	_, war := newTank(t, "", bulwark)
	near(t, "an Orc warrior's base health at 60", war.GetBaseStats()[stats.Health], 1689, 1e-9)

	talents := "-0000000001" // Reckless Swings
	base := selfCostRun(t, talents)
	paid := selfCostRun(t, talents, 2687, 1329555)
	if base.Dtps.Avg <= 0 || base.Tmi.Avg == 0 {
		t.Fatalf("the base run has DTPS %.1f and TMI %.1f, want both", base.Dtps.Avg, base.Tmi.Avg)
	}
	if paid.Dtps.Avg != base.Dtps.Avg {
		t.Errorf("DTPS is %.3f with the self costs, %.3f without: it must not count them", paid.Dtps.Avg, base.Dtps.Avg)
	}
	if paid.Tmi.Avg != base.Tmi.Avg {
		t.Errorf("TMI is %.3f with the self costs, %.3f without: it must not count them", paid.Tmi.Avg, base.Tmi.Avg)
	}
	needBase, _, _ := core.TankScoreSummary(base.TankScore, 0.05)
	needPaid, _, _ := core.TankScoreSummary(paid.TankScore, 0.05)
	// Bloodrage every minute and Reckless Swings every 15 s: (0.2 × 2 + 0.25 × 8) × 1689 over 120 s
	// is about 34 HPS more to put back.
	if needPaid <= needBase || math.IsNaN(needPaid) {
		t.Errorf("healing needed is %.1f with the self costs, %.1f without: it must count them", needPaid, needBase)
	}
}
