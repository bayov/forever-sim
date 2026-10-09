package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanElementalWeapons checks what Elemental Weapons 3/3 does to Windfury Weapon and
// Rockbiter Weapon for a level 30 Orc shaman.
//
// It adds 40% to Windfury's extra attack power once, so rank 1's 46 becomes 64.4. On the beta
// (2026-10-09) the tooltip said 64, and 28 normal Windfury hits averaged 1.196 times 91 normal
// white hits (1.201 for once, 1.281 for twice like the SoD sim). It adds 20% to Rockbiter's
// attack power, so rank 4's 177.6 becomes 213.1, and the beta sheet went up 213.
//
// It adds 15% to the whole Flametongue Weapon hit. On the beta with 6 spell power and a 3.6
// speed axe, rank 1 hit 19 every time, where (15.84 + 0.6) * 1.15 is 18.9 (16 or 17 with no
// points). Rank 3 hit 31 or 32, where its 27.5 with no points times 1.15 is 31.6. So the per-rank
// cut we see on the beta (shaman_audit.md 7.3) comes off before the 15%.
//
// We pin Dark Edge of Insanity's damage roll to its minimum and take the level 30 target's
// armor away, so every normal Windfury hit deals exactly 242 + 3.5 * (AP + 64.4) / 14. Every
// normal Flametongue hit (rank 3, no spell power) deals 35.28 / 4 * 3.5 * 1.15.
func TestOrcShamanElementalWeapons(t *testing.T) {
	newSim := func(talents string, imbue proto.WeaponImbue) (*core.Simulation, *EnhancementShaman) {
		items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 21134} // Dark Edge of Insanity
		player := newOrcShaman(30, talents, &proto.EnhancementShaman_Options{ShamanImbue: imbue})
		player.Equipment = &proto.EquipmentSpec{Items: items}
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  &proto.Encounter{Duration: 600, Targets: []*proto.Target{{Level: 30, MobType: proto.MobType_MobTypeBeast}}},
			SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()
		sim.PrePull()
		return sim, sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
	}

	t.Run("Rockbiter", func(t *testing.T) {
		_, none := newSim("", proto.WeaponImbue_WindfuryWeapon)
		_, rockbiter := newSim("", proto.WeaponImbue_RockbiterWeapon)
		_, talented := newSim("-00000003", proto.WeaponImbue_RockbiterWeapon)
		base := none.GetStat(stats.AttackPower)
		if got := rockbiter.GetStat(stats.AttackPower) - base; math.Abs(got-177.6) > 1e-9 {
			t.Errorf("Rockbiter gives %.2f attack power with 0 points, want 177.6", got)
		}
		if got := talented.GetStat(stats.AttackPower) - base; math.Abs(got-177.6*1.2) > 1e-9 {
			t.Errorf("Rockbiter gives %.2f attack power with 3 points, want %.2f", got, 177.6*1.2)
		}
	})

	for _, c := range []struct {
		talents string
		extraAP float64
	}{{"", 46}, {"-00000003", 46 * 1.4}} {
		t.Run("Windfury "+c.talents, func(t *testing.T) {
			sim, enh := newSim(c.talents, proto.WeaponImbue_WindfuryWeapon)
			enh.CurrentTarget.PseudoStats.ArmorMultiplier = 0
			mh := enh.AutoAttacks.MH()
			mh.BaseDamageMax = mh.BaseDamageMin

			want := 242 + 3.5*(enh.GetStat(stats.AttackPower)+c.extraAP)/core.DefaultAttackPowerPerDPS
			hits := 0
			imbue := enh.GetAura("Windfury Imbue")
			onHit := imbue.OnSpellHitDealt
			imbue.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				onHit(aura, sim, spell, result)
				if spell != enh.WindfuryWeaponMH || result.Outcome != core.OutcomeHit {
					return
				}
				if math.Abs(result.Damage-want) > 1e-9 {
					t.Errorf("at %s: Windfury hit for %.2f, want %.2f", sim.CurrentTime, result.Damage, want)
				}
				hits++
			}
			runSim(sim)

			if hits < 30 {
				t.Errorf("only %d normal Windfury hits", hits)
			}
		})
	}

	for _, c := range []struct {
		talents    string
		multiplier float64
	}{{"", 1}, {"-00000003", 1.15}} {
		t.Run("Flametongue "+c.talents, func(t *testing.T) {
			sim, enh := newSim(c.talents, proto.WeaponImbue_FlametongueWeapon)
			if sp := enh.GetStat(stats.SpellPower) + enh.GetStat(stats.FirePower); sp != 0 {
				t.Fatalf("fire spell power is %.0f, want 0", sp)
			}

			want := 35.28 / 4 * 3.5 * c.multiplier
			hits := 0
			imbue := enh.GetAura("Flametongue Imbue")
			onHit := imbue.OnSpellHitDealt
			imbue.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				onHit(aura, sim, spell, result)
				if spell.ActionID.SpellID != shaman.FlametongueWeaponSpellId[3] || result.Outcome != core.OutcomeHit {
					return
				}
				if math.Abs(result.Damage-want) > 1e-9 {
					t.Errorf("at %s: Flametongue hit for %.2f, want %.2f", sim.CurrentTime, result.Damage, want)
				}
				hits++
			}
			runSim(sim)

			if hits < 30 {
				t.Errorf("only %d normal Flametongue hits", hits)
			}
		})
	}
}
