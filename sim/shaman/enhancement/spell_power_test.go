package enhancement

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanSpellPower checks spell damage and healing from gear against a level 30 Orc
// shaman on the Forever beta (2026-10-08).
//
// The beta's GetSpellBonusDamage gave 49 for both nature and fire, and GetSpellBonusHealing
// gave 59. Most of the gear says "damage and healing by up to X", which counts for both.
// Naga Battle Gloves say "healing by up to 15 and damage by up to 5", so healing ends up 10
// above damage. The Mystic armor kits on the legs and feet add 4 and 2 to both.
//
// We also check armor, because every item there has its armor listed once on the beta and
// the kits add theirs on top.
func TestOrcShamanSpellPower(t *testing.T) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotShoulder] = &proto.ItemSpec{Id: 4197}                 // Berylline Pads
	items[proto.ItemSlot_ItemSlotChest] = &proto.ItemSpec{Id: 7065, Enchant: 866}      // Green Silk Armor, Lesser Stats
	items[proto.ItemSlot_ItemSlotHands] = &proto.ItemSpec{Id: 888, Enchant: 927}       // Naga Battle Gloves, Greater Strength
	items[proto.ItemSlot_ItemSlotWaist] = &proto.ItemSpec{Id: 252461}                  // Skirmisher's Leather Belt
	items[proto.ItemSlot_ItemSlotLegs] = &proto.ItemSpec{Id: 252458, Enchant: 1255096} // Totemic Leather Leggings, Mystic Heavy Armor Kit
	items[proto.ItemSlot_ItemSlotFeet] = &proto.ItemSpec{Id: 4320, Enchant: 1255124}   // Spidersilk Boots, Mystic Medium Armor Kit

	player := &proto.Player{
		Race:      proto.Race_RaceOrc,
		Class:     proto.Class_ClassShaman,
		Level:     30,
		Equipment: &proto.EquipmentSpec{Items: items},
		Spec: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
			Options: &proto.EnhancementShaman_Options{},
		}},
	}
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	env, _, _ := core.NewEnvironment(raid, &proto.Encounter{}, proto.Ruleset_RulesetForever, false)
	enh := env.Raid.Parties[0].Players[0].(*EnhancementShaman)
	self := &enh.Unit

	topRank := func(name string, ranks []*core.Spell) *core.Spell {
		for i := len(ranks) - 1; i >= 0; i-- {
			if ranks[i] != nil {
				return ranks[i]
			}
		}
		t.Fatalf("no %s rank at level 30", name)
		return nil
	}

	check := func(name string, have, want float64) {
		if have != want {
			t.Errorf("%s: got %.1f, want %.1f", name, have, want)
		}
	}
	check("nature damage", topRank("Lightning Bolt", enh.LightningBolt).GetSchoolDamage(self), 49)
	check("fire damage", topRank("Flame Shock", enh.FlameShock).GetSchoolDamage(self), 49)
	// The sim has no shaman heals besides Healing Stream Totem.
	check("healing", topRank("Healing Stream Totem", enh.HealingStreamTotem).HealingPower(self), 59)
	// 367 from the items and kits, and 2 Agility per point on 30 + 2 Agility.
	check("armor", enh.GetStats()[stats.Armor], 367+64)
}
