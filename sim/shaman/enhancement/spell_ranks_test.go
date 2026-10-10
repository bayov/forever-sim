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

// TestOrcShamanSpellRanks checks that a rotation naming a spell's top rank casts the highest
// rank our level knows.
//
// Our rotations are written at level 60, so they name rank 7 of Earth Shock, and a level 30
// shaman has to cast rank 4. The ranks we want are from the learn levels in the Forever client
// (70338, wago.tools SpellLevels), so this also checks the sim's learn levels. Lava Burst comes
// from a talent and has its own test (TestOrcShamanLavaBurstRanks).
func TestOrcShamanSpellRanks(t *testing.T) {
	cases := []struct {
		name       string
		ids        []int32
		at20, at30 int
	}{
		{"Lightning Bolt", shaman.LightningBoltSpellId[:], 4, 5},
		{"Chain Lightning", shaman.ChainLightningSpellId[:], 0, 0},
		{"Earth Shock", shaman.EarthShockSpellId[:], 3, 4},
		{"Flame Shock", shaman.FlameShockSpellId[:], 2, 3},
		{"Frost Shock", shaman.FrostShockSpellId[:], 1, 1},
		{"Lightning Shield", shaman.LightningShieldSpellId[:], 2, 3},
		{"Searing Totem", shaman.SearingTotemSpellId[:], 2, 3},
		{"Magma Totem", shaman.MagmaTotemSpellId[:], 0, 1},
		{"Fire Nova", shaman.FireNovaSpellId[:], 1, 2},
		{"Flametongue Totem", shaman.FlametongueTotemSpellId[:], 0, 1},
		{"Strength of Earth Totem", shaman.StrengthOfEarthTotemSpellId[:], 1, 2},
		{"Stoneskin Totem", shaman.StoneskinTotemSpellId[:], 2, 3},
		{"Windfury Totem", shaman.WindfuryTotemSpellId[:], 0, 0},
		{"Grace of Air Totem", shaman.GraceOfAirTotemSpellId[:], 0, 0},
		{"Windwall Totem", shaman.WindwallTotemSpellId[:], 0, 0},
		{"Healing Stream Totem", shaman.HealingStreamTotemSpellId[:], 1, 2},
		{"Mana Spring Totem", shaman.ManaSpringTotemSpellId[:], 0, 1},
	}
	for _, level := range []int32{20, 30, 60} {
		_, enh := newTargetLevelSim(level, level, whirlwind, 0, "")
		for _, c := range cases {
			want := len(c.ids) - 1
			switch level {
			case 20:
				want = c.at20
			case 30:
				want = c.at30
			}
			top := &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: c.ids[len(c.ids)-1]}}
			spell := enh.Rotation.GetAPLSpell(top)
			switch {
			case want == 0 && spell != nil:
				t.Errorf("level %d: %s cast rank %d, want none", level, c.name, spell.Rank)
			case want == 0:
			case spell == nil:
				t.Errorf("level %d: %s cast nothing, want rank %d", level, c.name, want)
			case spell.Rank != want || spell.ActionID.SpellID != c.ids[want]:
				t.Errorf("level %d: %s cast %d rank %d, want %d rank %d", level, c.name, spell.ActionID.SpellID, spell.Rank, c.ids[want], want)
			}
		}
	}
}

// TestOrcShamanImbueDamage checks a Flametongue Weapon and a Frostbrand hit at some levels
// against the Forever client (70338, shaman_audit.md 7.1).
//
// Flametongue Weapon deals N / 25 per 4 sec of weapon speed at a rank's top level, and 1.68 a
// level less for rank 3 below it. Frostbrand rank 1 grows 2.1 a level from 32 at 20 to 44.6 at
// 26. We have no spell power and the target is our level, so every hit that doesn't crit deals
// exactly that.
func TestOrcShamanImbueDamage(t *testing.T) {
	flametongue, frostbrand := proto.WeaponImbue_FlametongueWeapon, proto.WeaponImbue_FrostbrandWeapon
	cases := []struct {
		name     string
		imbue    proto.WeaponImbue
		level    int32
		spellID  int32
		damage   float64
		perSpeed bool
	}{
		{"Flametongue rank 1 at 16", flametongue, 16, shaman.FlametongueWeaponSpellId[1], 440.0 / 25, true},
		{"Flametongue rank 2 at 24", flametongue, 24, shaman.FlametongueWeaponSpellId[2], 653.0 / 25, true},
		{"Flametongue rank 3 at 30", flametongue, 30, shaman.FlametongueWeaponSpellId[3], 1052.0/25 - 4*1.68, true},
		{"Flametongue rank 3 at 34", flametongue, 34, shaman.FlametongueWeaponSpellId[3], 1052.0 / 25, true},
		{"Flametongue rank 4 at 44", flametongue, 44, shaman.FlametongueWeaponSpellId[4], 1728.0 / 25, true},
		{"Flametongue rank 5 at 54", flametongue, 54, shaman.FlametongueWeaponSpellId[5], 2372.0 / 25, true},
		{"Flametongue rank 6 at 60", flametongue, 60, shaman.FlametongueWeaponSpellId[6], 2810.0 / 25, true},
		{"Frostbrand rank 1 at 20", frostbrand, 20, shaman.FrostbrandWeaponSpellId[1], 32, false},
		{"Frostbrand rank 1 at 26", frostbrand, 26, shaman.FrostbrandWeaponSpellId[1], 44.6, false},
		{"Frostbrand rank 1 at 27", frostbrand, 27, shaman.FrostbrandWeaponSpellId[1], 44.6, false},
	}
	for _, c := range cases {
		player := newOrcShaman(c.level, "", &proto.EnhancementShaman_Options{ShamanImbue: c.imbue})
		for range proto.ItemSlot_ItemSlotMainHand {
			player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
		}
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: whirlwind})
		target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
		target.Level = c.level
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  &proto.Encounter{Duration: 10, Targets: []*proto.Target{target}},
			SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()
		sim.PrePull()
		enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)

		want := c.damage
		if c.perSpeed {
			want *= enh.MainHand().SwingSpeed / 4
		}
		spell := enh.GetSpell(core.ActionID{SpellID: c.spellID})
		if spell == nil {
			t.Fatalf("%s: no spell %d", c.name, c.spellID)
		}
		var hits []float64
		imbue := enh.GetAura(map[proto.WeaponImbue]string{flametongue: "Flametongue Imbue", frostbrand: "Frostbrand Imbue"}[c.imbue])
		imbue.OnSpellHitDealt = func(_ *core.Aura, _ *core.Simulation, s *core.Spell, result *core.SpellResult) {
			if s == spell && result.Outcome == core.OutcomeHit {
				hits = append(hits, result.Damage)
			}
		}
		for range 20 {
			spell.Cast(sim, enh.CurrentTarget)
		}
		if len(hits) == 0 {
			t.Errorf("%s: no normal hits in 20 casts", c.name)
		}
		for _, damage := range hits {
			if math.Abs(damage-want) > 1e-9 {
				t.Errorf("%s: hit for %.4f, want %.4f", c.name, damage, want)
				break
			}
		}
	}
}

// TestOrcShamanWindfuryTotemAura checks that a rotation finds the Windfury Totem aura of the
// rank our level knows.
//
// Our rotations ask for rank 3's aura, 10612, the party aura in the Forever client (70338).
// Rotations saved before shaman_audit.md 7.1 ask for 10611, which isn't in the client. Both
// should find rank 1's 8515 at 40, rank 2's 10609 at 50 and rank 3's 10612 at 60. When the
// rotation finds the aura, it puts the totem down once in 30 sec, as the totem lasts 5 min.
func TestOrcShamanWindfuryTotemAura(t *testing.T) {
	cases := []struct {
		level int32
		rank  int
	}{{40, 1}, {50, 2}, {60, 3}}
	for _, c := range cases {
		for _, auraID := range []int32{10612, 10611} {
			player := newOrcShaman(c.level, "", &proto.EnhancementShaman_Options{})
			isActive := &proto.APLValue{Value: &proto.APLValue_AuraIsActive{AuraIsActive: &proto.APLValueAuraIsActive{
				AuraId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: auraID}},
			}}}
			player.Rotation = &proto.APLRotation{PriorityList: []*proto.APLListItem{{Action: &proto.APLAction{
				Condition: &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: isActive}}},
				Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
					SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: shaman.WindfuryTotemSpellId[3]}},
				}},
			}}}}
			sim, enh := newShamanSim(player, &proto.Debuffs{}, 30)
			runSim(sim)

			aura := enh.GetAuraByID(core.ActionID{SpellID: shaman.WindfuryBuffAuraId[c.rank]})
			if aura == nil || !aura.IsActive() {
				t.Errorf("level %d, asking for %d: aura %d isn't up", c.level, auraID, shaman.WindfuryBuffAuraId[c.rank])
			}
			totem := enh.WindfuryTotem[c.rank]
			if casts := totem.SpellMetrics[enh.CurrentTarget.UnitIndex].Casts; casts != 1 {
				t.Errorf("level %d, asking for %d: Windfury Totem rank %d cast %d times, want 1", c.level, auraID, c.rank, casts)
			}
		}
	}
}
