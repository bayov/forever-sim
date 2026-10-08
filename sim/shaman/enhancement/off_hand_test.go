package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanOffHandWeapon checks that a one-hander in our off hand does nothing.
//
// Forever's shamans can't dual wield (the user, 2026-10-08), and the UI doesn't let us pick an
// off-hand weapon. A gear set from elsewhere can still have one, like the one-handers rotopt's
// gear search put there before. So we run Deathbringer (2.9 speed) with Crul'shorukh (13
// Stamina, 36 attack power) in the off hand and without it. The off hand should add no stats,
// never swing and give no dual wield miss penalty, so with the same seed both fights play out
// hit for hit. The main hand swings every 2.9 sec.
func TestOrcShamanOffHandWeapon(t *testing.T) {
	type fight struct {
		enh    *EnhancementShaman
		sim    *core.Simulation
		damage float64
		swings int32
	}
	run := func(offHand int32) fight {
		items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotOffHand+1)
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 17068} // Deathbringer
		items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: offHand}

		player := newOrcShaman(30, "", &proto.EnhancementShaman_Options{ShamanImbue: proto.WeaponImbue_WindfuryWeapon})
		player.Equipment = &proto.EquipmentSpec{Items: items}
		sim, enh := newShamanSim(player, &proto.Debuffs{}, 120)
		f := fight{enh: enh, sim: sim}

		if got, want := enh.AutoAttacks.MainhandSwingSpeed(), 2900*time.Millisecond; got != want {
			t.Errorf("off hand %d: the main hand swings every %s, want %s", offHand, got, want)
		}
		runSim(sim)
		mh := enh.AutoAttacks.MHAuto().SpellMetrics[0]
		f.damage, f.swings = mh.TotalDamage, mh.Casts
		return f
	}

	without := run(0)
	with := run(19363) // Crul'shorukh, Edge of Chaos

	if with.enh.HasOHWeapon() || with.enh.OffHand().ID != 0 {
		t.Errorf("the off hand still holds item %d", with.enh.OffHand().ID)
	}
	if with.enh.AutoAttacks.IsDualWielding {
		t.Errorf("we dual wield")
	}
	if oh := with.enh.AutoAttacks.OHAuto(); oh != nil && oh.SpellMetrics[0].Casts > 0 {
		t.Errorf("the off hand swung %d times", oh.SpellMetrics[0].Casts)
	}
	if got, want := with.enh.GetStats(), without.enh.GetStats(); got != want {
		t.Errorf("stats with the off hand differ:\n%s\nwant\n%s", got, want)
	}
	if with.swings != without.swings || with.damage != without.damage {
		t.Errorf("main hand: %d swings for %.0f with the off hand, %d for %.0f without", with.swings, with.damage, without.swings, without.damage)
	}
	if without.swings < 39 {
		t.Errorf("only %d main hand swings in 120 sec", without.swings)
	}
}
