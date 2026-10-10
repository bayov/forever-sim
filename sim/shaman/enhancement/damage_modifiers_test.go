package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/shaman"
	googleProto "google.golang.org/protobuf/proto"
)

// TestOrcShamanDamageModifiers checks how flat and percent bonuses change our physical damage,
// with Rage of the Storm as the one percent bonus we have.
//
// Rage of the Storm (280604) gives Stormstrike 10% more damage and leaves white hits and
// Windfury alone. A flat "+N damage" effect adds to the hit before the percent bonuses, so
// the 10% applies to it too. We roll a 100 damage hit with +10 flat damage against a target
// with no armor, so a normal hit deals 110, or 121 for a Stormstrike with the mace.
func TestOrcShamanDamageModifiers(t *testing.T) {
	const flatBonus = 10.0
	cases := []struct {
		name                         string
		weapon                       int32
		white, windfury, stormstrike float64
	}{
		{"Whirlwind Axe", whirlwind, 110, 110, 110},
		{"Rage of the Storm", shaman.RageOfTheStorm, 110, 110, 121},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim, enh := newWeaponSim(30, stormstrike, &proto.ItemSpec{Id: c.weapon}, proto.WeaponImbue_WindfuryWeapon, proto.TotemWeaponBuff_TotemWeaponBuffNone, 0, 60)
			enh.PseudoStats.BonusPhysicalDamage += flatBonus
			if enh.Stormstrike == nil || enh.WindfuryWeaponMH == nil {
				t.Fatalf("missing Stormstrike or Windfury Weapon")
			}

			white := enh.AutoAttacks.MHAuto()
			for _, attack := range []struct {
				spell   *core.Spell
				outcome core.OutcomeApplier
				want    float64
			}{
				{white, white.OutcomeMeleeWhite, c.white},
				{enh.WindfuryWeaponMH, enh.WindfuryWeaponMH.OutcomeMeleeWeaponSpecialHitAndCrit, c.windfury},
				{enh.Stormstrike, enh.Stormstrike.OutcomeMeleeWeaponSpecialHitAndCrit, c.stormstrike},
			} {
				hit := 0.0
				for range 100 {
					result := attack.spell.CalcDamage(sim, enh.CurrentTarget, 100, attack.outcome)
					if result.Outcome == core.OutcomeHit {
						hit = result.Damage
						break
					}
				}
				if math.Abs(hit-attack.want) > 1e-9 {
					t.Errorf("%s hit for %.4f, want %.4f", attack.spell.ActionID, hit, attack.want)
				}
			}
		})
	}
}

// newWeaponSim starts a Forever sim of seconds of an Orc shaman of level with talents,
// holding weapon with imbue on it, against a target 2 levels up with no armor. Another
// shaman gives us totem (Windfury or Flametongue Totem), and a trinket that isn't 0 goes in
// the first trinket slot.
func newWeaponSim(level int32, talents string, weapon *proto.ItemSpec, imbue proto.WeaponImbue, totem proto.TotemWeaponBuff, trinket int32, seconds float64) (*core.Simulation, *EnhancementShaman) {
	player := newOrcShaman(level, talents, &proto.EnhancementShaman_Options{ShamanImbue: imbue})
	for range proto.ItemSlot_ItemSlotMainHand {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
	}
	player.Equipment.Items = append(player.Equipment.Items, weapon)
	player.Equipment.Items[proto.ItemSlot_ItemSlotTrinket1].Id = trinket
	target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
	target.Level = level + 2
	target.Stats[proto.Stat_StatArmor] = 0
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{TotemWeaponBuff: totem}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: seconds, Targets: []*proto.Target{target}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
}
