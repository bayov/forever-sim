package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanBloodFury checks the Orc racial against the beta (2026-10-11, beta_results.txt).
//
// Blood Fury goes while the GCD is up and leaves it alone. On the beta, Lightning Shield went
// out 0.245 sec after it. For 15 sec it adds 10% to attack power, spell power and bonus
// healing. The character sheet went from 59 to 65 healing with 49 spell power and the +10
// healing gloves, so the healing-only stat gets its 10% too. We give the character the same
// spell power and gloves.
//
// The 10% follows the total while it's up. On the beta (2026-10-12) a Scroll of Strength II
// (+9 Strength, 18 attack power) used during Blood Fury took attack power from 421 to 441, and
// it was 401 after Blood Fury ended. We add the same 18 attack power during it.
//
// The weapon's damage range also goes up by 10% of its average, with the same spread. With
// Bloody Brass Knuckles (24 to 46) that's 27.5 to 49.5, as on the character sheet, and the
// average hit with attack power comes out 1.1 times bigger.
func TestOrcShamanBloodFury(t *testing.T) {
	player := newOrcShaman(30, "", &proto.EnhancementShaman_Options{})
	player.BonusStats = &proto.UnitStats{Stats: stats.Stats{stats.SpellPower: 49, stats.HealingPower: 10}.ToFloatArray()}
	for range proto.ItemSlot_ItemSlotMainHand {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
	}
	player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: 7683})
	sim, enh := newShamanSim(player, &proto.Debuffs{}, 30)
	bloodFury := enh.GetSpell(core.ActionID{SpellID: 20572})
	aura := enh.GetAuraByID(core.ActionID{SpellID: 20572})
	mh := enh.AutoAttacks.MH()

	healing := func() float64 { return enh.GetStat(stats.SpellPower) + enh.GetStat(stats.HealingPower) }
	ap, sp, heal := enh.GetStat(stats.AttackPower), enh.GetStat(stats.SpellPower), healing()
	hit := mh.CalculateAverageWeaponDamage(ap)
	check := func(label string, got, want float64) {
		if math.Abs(got-want) > 1e-6 {
			t.Errorf("%s: %s is %.2f, want %.2f", sim.CurrentTime, label, got, want)
		}
	}

	at(sim, 1, func(sim *core.Simulation) {
		if aura.IsActive() {
			t.Fatalf("Blood Fury is up before we cast it")
		}
		castNow(t, sim, enh, topRank(t, "Earth Shock", enh.EarthShock))
		if !bloodFury.Cast(sim, enh.CurrentTarget) {
			t.Fatalf("Blood Fury didn't cast during the GCD")
		}
		if got := enh.GCD.TimeToReady(sim); got != 1500*time.Millisecond {
			t.Errorf("the GCD is up for %s after Blood Fury, want the Earth Shock's 1.5s", got)
		}
		check("attack power", enh.GetStat(stats.AttackPower), ap*1.1)
		check("spell power", enh.GetStat(stats.SpellPower), sp*1.1)
		check("healing", healing(), heal*1.1)
		check("weapon minimum", mh.BaseDamageMin, 27.5)
		check("weapon maximum", mh.BaseDamageMax, 49.5)
		check("average hit", mh.CalculateAverageWeaponDamage(enh.GetStat(stats.AttackPower)), hit*1.1)

		enh.AddStatDynamic(sim, stats.AttackPower, 18)
		check("attack power with the scroll", enh.GetStat(stats.AttackPower), (ap+18)*1.1)
	})
	if heal != 59 {
		t.Fatalf("healing is %.2f before Blood Fury, want the beta's 59", heal)
	}

	at(sim, 16.5, func(sim *core.Simulation) {
		if aura.IsActive() {
			t.Errorf("Blood Fury is still up after 15 sec")
		}
		check("attack power", enh.GetStat(stats.AttackPower), ap+18)
		check("spell power", enh.GetStat(stats.SpellPower), sp)
		check("healing", healing(), heal)
		check("weapon minimum", mh.BaseDamageMin, 24)
		check("weapon maximum", mh.BaseDamageMax, 46)
	})
	runSim(sim)
}
