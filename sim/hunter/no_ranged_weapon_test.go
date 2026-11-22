package hunter

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

// A hunter with the ranged slot empty (a gear ranking sims the set without its ranged pick) has no
// Auto Shot. Its ranged swing duration was 0, so the swing came due again in the same instant and
// the sim never left it. With the shot at NeverExpires, a rotation that waits on it (the MM weave
// waits for a moment after the next Auto Shot) must wait out the fight, not overflow into the past
// and end its wait in the same instant forever.
func TestNoRangedWeaponFinishes(t *testing.T) {
	waitOnShot := &proto.APLRotation{}
	if err := protojson.Unmarshal([]byte(`{"priorityList": [{"action": {"wait": {"duration": {"math": {
		"lhs": {"const": {"val": "0.2s"}}, "op": "OpSub", "rhs": {"math": {
			"lhs": {"autoSwingTime": {"autoType": "Ranged"}}, "op": "OpSub",
			"rhs": {"autoTimeToNext": {"autoType": "Ranged"}}}}}}}}}]}`), waitOnShot); err != nil {
		t.Fatal(err)
	}
	for _, ruleset := range []proto.Ruleset{proto.Ruleset_RulesetForever, proto.Ruleset_RulesetClassic} {
		for _, rotation := range []*proto.APLRotation{{}, waitOnShot} {
			t.Run(ruleset.String(), func(t *testing.T) {
				gear := googleProto.Clone(core.GetGearSet("../../ui/hunter/gear_sets", "level60_phase1_synthetic_ranged").GearSet).(*proto.EquipmentSpec)
				gear.Items[proto.ItemSlot_ItemSlotRanged] = &proto.ItemSpec{}
				player := &proto.Player{
					Class:              proto.Class_ClassHunter,
					Race:               proto.Race_RaceDwarf,
					Level:              60,
					Equipment:          gear,
					Spec:               PlayerOptionsNoPet,
					Rotation:           rotation,
					DistanceFromTarget: 20,
				}
				done := make(chan *proto.RaidSimResult, 1)
				go func() {
					done <- core.RunRaidSim(&proto.RaidSimRequest{
						Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
						Encounter:  &proto.Encounter{Duration: 30, Targets: []*proto.Target{core.NewDefaultTarget()}},
						SimOptions: &proto.SimOptions{Ruleset: ruleset, RandomSeed: 1, Iterations: 2},
					})
				}()
				select {
				case res := <-done:
					if res.Error != nil {
						t.Fatalf("sim failed: %v", res.Error.Message)
					}
				case <-time.After(30 * time.Second):
					t.Fatalf("sim with an empty ranged slot did not finish")
				}
			})
		}
	}
}
