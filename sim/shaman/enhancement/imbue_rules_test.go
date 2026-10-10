package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanImbueRules checks how the shaman imbue goes with the consumables' main hand
// slot for a level 30 Orc shaman (shaman_audit.md 7.6).
//
// Under Forever the shaman imbue and an oil or stone sit in two different places on the weapon,
// so Rockbiter's 177.6 attack power and Brilliant Wizard Oil's 36 spell power and 1% spell crit
// both apply. Old saved settings can still hold Rockbiter in the consumables. When the class
// settings hold it too, its attack power only counts once.
func TestOrcShamanImbueRules(t *testing.T) {
	newShaman := func(shamanImbue, consumesImbue proto.WeaponImbue) *EnhancementShaman {
		player := newOrcShaman(30, "", &proto.EnhancementShaman_Options{ShamanImbue: shamanImbue})
		for range proto.ItemSlot_ItemSlotMainHand {
			player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
		}
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: darkEdge})
		player.Consumes = &proto.Consumes{MainHandImbue: consumesImbue}
		_, enh := newShamanSim(player, &proto.Debuffs{}, 10)
		return enh
	}
	none := newShaman(proto.WeaponImbue_WeaponImbueUnknown, proto.WeaponImbue_WeaponImbueUnknown)

	rockbiter, oil := proto.WeaponImbue_RockbiterWeapon, proto.WeaponImbue_BrilliantWizardOil
	for _, c := range []struct {
		name                  string
		shamanImbue, consumes proto.WeaponImbue
		want                  stats.Stats
	}{
		{"Rockbiter and Brilliant Wizard Oil", rockbiter, oil, stats.Stats{stats.AttackPower: 177.6, stats.SpellPower: 36, stats.SpellCrit: core.SpellCritRatingPerCritChance}},
		{"Rockbiter in both", rockbiter, rockbiter, stats.Stats{stats.AttackPower: 177.6}},
		{"Rockbiter in the consumables", proto.WeaponImbue_WeaponImbueUnknown, rockbiter, stats.Stats{stats.AttackPower: 177.6}},
	} {
		enh := newShaman(c.shamanImbue, c.consumes)
		for _, stat := range []stats.Stat{stats.AttackPower, stats.SpellPower, stats.SpellCrit} {
			if got := enh.GetStat(stat) - none.GetStat(stat); math.Abs(got-c.want[stat]) > 1e-9 {
				t.Errorf("%s: %s went up %.2f, want %.2f", c.name, stat.StatName(), got, c.want[stat])
			}
		}
	}
}
