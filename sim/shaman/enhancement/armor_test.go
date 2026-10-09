package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	googleProto "google.golang.org/protobuf/proto"
)

// TestOrcShamanArmor checks the target's armor under the Forever armor debuffs and how much
// of a white hit it takes.
//
// A hit loses armor / (armor + 400 + 85 * our level) (the 2019 Classic formula). The
// debuffs take the rank the raid's level allows (client 1.60.1.70291): Sunder Armor 90 to
// 450 a stack for 5 stacks, Expose Armor the same per combo point for 5 points, Faerie Fire
// and Curse of Recklessness 175 to 505. Sunder Armor and Expose Armor don't stack, and under
// Forever neither do Faerie Fire and Curse of Recklessness.
func TestOrcShamanArmor(t *testing.T) {
	cases := []struct {
		name               string
		level, targetLevel int32
		armor              float64
		debuffs            *proto.Debuffs
		want               float64
	}{
		{"level 30 against Vishas", 30, 32, 1063, &proto.Debuffs{}, 1063},
		{"level 60 against the boss", 60, 63, 3731, &proto.Debuffs{}, 3731},
		{"Sunder Armor at 60", 60, 63, 3731, &proto.Debuffs{SunderArmor: true}, 3731 - 5*450},
		{"Expose Armor at 60", 60, 63, 3731, &proto.Debuffs{ExposeArmor: proto.TristateEffect_TristateEffectRegular}, 3731 - 5*450},
		{"Faerie Fire at 60", 60, 63, 3731, &proto.Debuffs{FaerieFire: true}, 3731 - 505},
		{"Curse of Recklessness at 60", 60, 63, 3731, &proto.Debuffs{CurseOfRecklessness: true}, 3731 - 505},
		{"every debuff at 60", 60, 63, 3731, &proto.Debuffs{
			SunderArmor: true, ExposeArmor: proto.TristateEffect_TristateEffectImproved, FaerieFire: true, CurseOfRecklessness: true,
		}, 3731 - 5*450 - 505},
		// Level 30 learns rank 2 of each: Sunder Armor at 22, Expose Armor at 26, Curse of
		// Recklessness at 28 and Faerie Fire at 30.
		{"every debuff at 30", 30, 32, 1500, &proto.Debuffs{
			SunderArmor: true, ExposeArmor: proto.TristateEffect_TristateEffectRegular, FaerieFire: true, CurseOfRecklessness: true,
		}, 1500 - 5*180 - 285},
		{"Expose Armor at 30", 30, 32, 1500, &proto.Debuffs{ExposeArmor: proto.TristateEffect_TristateEffectRegular}, 1500 - 5*180},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			player := newOrcShaman(c.level, stormstrike, &proto.EnhancementShaman_Options{})
			for range proto.ItemSlot_ItemSlotMainHand {
				player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
			}
			player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: whirlwind})
			target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
			target.Level = c.targetLevel
			target.Stats[proto.Stat_StatArmor] = c.armor
			sim := core.NewSim(&proto.RaidSimRequest{
				Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, c.debuffs),
				Encounter:  &proto.Encounter{Duration: 6, Targets: []*proto.Target{target}},
				SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
			}, simsignals.CreateSignals())
			sim.Reset()
			sim.PrePull()
			enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)

			// Sunder Armor stacks up and Expose Armor goes up in the first 3 sec.
			checked := false
			at(sim, 5, func(sim *core.Simulation) {
				checked = true
				if got := enh.CurrentTarget.Armor(); got != c.want {
					t.Errorf("the target has %.0f armor, want %.0f", got, c.want)
				}
				armor := max(c.want, 0)
				want := 100 * (1 - armor/(armor+400+85*float64(c.level)))
				white := enh.AutoAttacks.MHAuto()
				for range 100 {
					result := white.CalcDamage(sim, enh.CurrentTarget, 100, white.OutcomeMeleeWhite)
					if result.Outcome.Matches(core.OutcomeHit) {
						if math.Abs(result.Damage-want) > 1e-6 {
							t.Errorf("a 100 damage white hit dealt %.4f, want %.4f", result.Damage, want)
						}
						return
					}
				}
				t.Errorf("no normal hit in 100 rolls")
			})
			runSim(sim)
			if !checked {
				t.Fatalf("the check at 5 sec never ran")
			}
		})
	}
}
