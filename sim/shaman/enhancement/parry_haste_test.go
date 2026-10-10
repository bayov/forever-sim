package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	googleProto "google.golang.org/protobuf/proto"
)

const spiritWeapons = "-00000000000001"

// TestOrcShamanParryHaste checks how much sooner a parry brings the next swing, for the enemy
// and for us.
//
// With more than 60% of the swing left the next swing comes 40% of the swing sooner. With 20
// to 60% left it comes 20% of the swing after the parry, and with 20% or less left nothing
// changes. The beta log follows this rule for mobs and for us (shaman_audit.md 6.2). We give
// each parry at a set point of the swing, so no roll decides it. The enemy swings every 2.0
// sec and parries our white hit. We hold Whirlwind Axe and parry the enemy's swing with Spirit
// Weapons.
func TestOrcShamanParryHaste(t *testing.T) {
	cases := []struct {
		name string
		into float64 // When the parry comes, as a share of the swing after the last swing.
		want float64 // When the next swing comes, as a share of the swing after the last swing.
	}{
		{"80% left", 0.2, 0.6},
		{"50% left", 0.5, 0.7},
		{"15% left", 0.85, 1},
	}
	for _, c := range cases {
		for _, enemyParries := range []bool{true, false} {
			sim, enh := newTankSim(newOrcShaman(30, spiritWeapons, &proto.EnhancementShaman_Options{}), 30)
			parrier, attack := enh.CurrentTarget, enh.AutoAttacks.MHAuto()
			who := "the enemy"
			if !enemyParries {
				parrier, attack = &enh.Unit, enh.CurrentTarget.AutoAttacks.MHAuto()
				who = "we"
			}

			checked := false
			at(sim, 5, func(sim *core.Simulation) {
				lastSwing := parrier.AutoAttacks.MainhandSwingAt()
				speed := parrier.AutoAttacks.MainhandSwingSpeed()
				share := func(f float64) time.Duration { return lastSwing + time.Duration(f*float64(speed)) }
				sim.AddPendingAction(&core.PendingAction{NextActionAt: share(c.into), OnAction: func(sim *core.Simulation) {
					attack.CalcAndDealDamage(sim, parrier, 0, forcedOutcome(core.OutcomeParry))
					checked = true
					if got, want := parrier.AutoAttacks.MainhandSwingAt(), share(c.want); (got - want).Abs() > time.Millisecond {
						t.Errorf("%s, %s parried: next swing at %s, want %s", c.name, who, got-lastSwing, want-lastSwing)
					}
				}})
			})
			runSim(sim)
			if !checked {
				t.Errorf("%s, %s parried: the fight ended before the parry", c.name, who)
			}
		}
	}
}

// newTankSim starts the player with Whirlwind Axe, tanking an enemy of targetLevel who swings
// every 2.0 sec.
func newTankSim(player *proto.Player, targetLevel int32) (*core.Simulation, *EnhancementShaman) {
	for range proto.ItemSlot_ItemSlotMainHand {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
	}
	player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: whirlwind})
	target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
	target.Level = targetLevel
	target.MinBaseDamage = 52
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Duration: 20, Targets: []*proto.Target{target}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
}
