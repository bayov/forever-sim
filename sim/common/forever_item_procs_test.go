//go:build with_db

package common_test

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/mage"
)

// A level 60 mage on ruleset wearing item in slot and nothing else, against a target of mobType.
func newMageVs(t *testing.T, ruleset proto.Ruleset, mobType proto.MobType, slot proto.ItemSlot, item int32) (*core.Simulation, *mage.Mage) {
	t.Helper()
	specs := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range specs {
		specs[i] = &proto.ItemSpec{}
	}
	specs[slot].Id = item
	player := &proto.Player{
		Race:      proto.Race_RaceGnome,
		Class:     proto.Class_ClassMage,
		Level:     60,
		Equipment: &proto.EquipmentSpec{Items: specs},
		Spec:      &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{}}},
		Rotation:  &proto.APLRotation{},
	}
	target := core.NewDefaultTarget()
	target.MobType = mobType
	s := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 600, Targets: []*proto.Target{target}},
		SimOptions: &proto.SimOptions{Ruleset: ruleset, RandomSeed: 1},
	}, simsignals.CreateSignals())
	s.Reset()
	s.PrePull()
	return s, s.Raid.Parties[0].Players[0].(*mage.Mage)
}

// The power against a creature type of the items whose Forever value is its own stat in the
// client's ItemSparse (client 1.60.1.71037), against Classic's.
func TestForeverMobTypePower(t *testing.T) {
	for _, c := range []struct {
		name             string
		slot             proto.ItemSlot
		item             int32
		mobType          proto.MobType
		spell            bool
		forever, classic float64
	}{
		{"Gauntlets of Undead Slaying", proto.ItemSlot_ItemSlotHands, common.GauntletsOfUndeadSlaying, proto.MobType_MobTypeUndead, false, 30, 60},
		{"Handwraps of Undead Slaying", proto.ItemSlot_ItemSlotHands, common.HandwrapsOfUndeadSlaying, proto.MobType_MobTypeUndead, false, 30, 60},
		{"Handguards of Undead Slaying", proto.ItemSlot_ItemSlotHands, common.HandguardsOfUndeadSlaying, proto.MobType_MobTypeUndead, false, 30, 60},
		{"Gloves of Undead Cleansing", proto.ItemSlot_ItemSlotHands, common.GlovesOfUndeadCleansing, proto.MobType_MobTypeUndead, true, 18, 35},
		{"Enchanted Azsharite Felbane Sword", proto.ItemSlot_ItemSlotMainHand, common.EnchantedAzshariteSword, proto.MobType_MobTypeDemon, false, 15, 33},
		{"Pip's Skinner", proto.ItemSlot_ItemSlotMainHand, common.PipsSkinner, proto.MobType_MobTypeBeast, false, 30, 45},
		{"Fiendish Machete", proto.ItemSlot_ItemSlotMainHand, common.FiendishMachete, proto.MobType_MobTypeElemental, false, 18, 36},
		{"Argent Avenger", proto.ItemSlot_ItemSlotMainHand, common.ArgentAvenger, proto.MobType_MobTypeUndead, false, 12, 0},
	} {
		for _, r := range []struct {
			ruleset proto.Ruleset
			want    float64
		}{
			{proto.Ruleset_RulesetForever, c.forever},
			{proto.Ruleset_RulesetClassic, c.classic},
		} {
			_, m := newMageVs(t, r.ruleset, c.mobType, c.slot, c.item)
			at := m.AttackTables[m.CurrentTarget.UnitIndex][proto.CastType_CastTypeMainHand]
			got := core.Ternary(c.spell, at.BonusSpellDamageTaken, at.BonusAttackPowerTaken)
			if got != r.want {
				t.Errorf("%s %s: %.0f, want %.0f", r.ruleset, c.name, got, r.want)
			}
		}
	}
}

// Serenity Field and Flamebreak Talisman are Forever uses. Flamebreak's Agility doubles with an
// elemental in the fight.
func TestForeverItemUses(t *testing.T) {
	for _, c := range []struct {
		name               string
		item               int32
		mobType            proto.MobType
		stat               stats.Stat
		bonus              float64
		cooldown, category time.Duration
	}{
		{"Serenity Field", common.SerenityField, proto.MobType_MobTypeUnknown, stats.Spirit, 72, 6 * time.Minute, 15 * time.Second},
		{"Flamebreak Talisman", common.FlamebreakTalisman, proto.MobType_MobTypeUnknown, stats.Agility, 33, 5 * time.Minute, 30 * time.Second},
		{"Flamebreak Talisman vs Elementals", common.FlamebreakTalisman, proto.MobType_MobTypeElemental, stats.Agility, 66, 5 * time.Minute, 30 * time.Second},
	} {
		_, era := newMageVs(t, proto.Ruleset_RulesetClassic, c.mobType, proto.ItemSlot_ItemSlotTrinket1, c.item)
		if era.GetSpell(core.ActionID{ItemID: c.item}) != nil {
			t.Errorf("Classic %s: has a use", c.name)
		}
		s, m := newMageVs(t, proto.Ruleset_RulesetForever, c.mobType, proto.ItemSlot_ItemSlotTrinket1, c.item)
		spell := m.GetSpell(core.ActionID{ItemID: c.item})
		if spell == nil {
			t.Errorf("Forever %s: no use", c.name)
			continue
		}
		if spell.CD.Duration != c.cooldown || spell.SharedCD.Duration != c.category {
			t.Errorf("%s cooldowns %v and %v, want %v and %v", c.name, spell.CD.Duration, spell.SharedCD.Duration, c.cooldown, c.category)
		}
		before := m.GetStat(c.stat)
		spell.Cast(s, m.CurrentTarget)
		if got := m.GetStat(c.stat) - before; got < c.bonus || got > c.bonus*1.2 {
			t.Errorf("%s gives %.1f, want %.0f before multipliers", c.name, got, c.bonus)
		}
	}
}
