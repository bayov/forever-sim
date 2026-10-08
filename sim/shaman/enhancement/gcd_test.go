package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanGCD checks the global cooldown of each shaman spell at level 60.
//
// Spells and Stormstrike take 1.5 sec and totems 1 sec. Rage of the Farseer and Nature's
// Swiftness take none (wowhead Forever spell pages, 2026-10-08). Haste never shortens it, as
// in 1.12. A cast longer than 1.5 sec holds it until the cast ends, and a shorter one still
// holds it for 1.5 sec.
func TestOrcShamanGCD(t *testing.T) {
	// Stormstrike, Maelstrom Weapon 5/5, Rage of the Farseer, Water Shield and Nature's
	// Swiftness.
	sim, enh := newShamanSim(newOrcShaman(60, "-000000000000100051-00000000100001", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 120)
	metrics := enh.GetManaNotCastingMetrics()
	lb := topRank(t, "Lightning Bolt", enh.LightningBolt)

	// We cast with a full mana bar and every cooldown reset.
	cast := func(sim *core.Simulation, spell *core.Spell) {
		enh.AddMana(sim, enh.MaxMana(), metrics)
		if spell.CD.Timer != nil {
			spell.CD.Reset()
		}
		if spell.SharedCD.Timer != nil {
			spell.SharedCD.Reset()
		}
		castNow(t, sim, enh, spell)
	}
	checkGCD := func(sim *core.Simulation, label string, want time.Duration) {
		if got := enh.GCD.TimeToReady(sim); got != want {
			t.Errorf("%s: the GCD is up for %s, want %s", label, got, want)
		}
	}

	spells := []struct {
		name  string
		spell *core.Spell
		gcd   time.Duration
	}{
		{"Earth Shock", topRank(t, "Earth Shock", enh.EarthShock), 1500 * time.Millisecond},
		{"Flame Shock", topRank(t, "Flame Shock", enh.FlameShock), 1500 * time.Millisecond},
		{"Frost Shock", topRank(t, "Frost Shock", enh.FrostShock), 1500 * time.Millisecond},
		{"Stormstrike", enh.Stormstrike, 1500 * time.Millisecond},
		{"Lightning Shield", topRank(t, "Lightning Shield", enh.LightningShield), 1500 * time.Millisecond},
		{"Water Shield", enh.WaterShield, 1500 * time.Millisecond},
		{"Strength of Earth Totem", topRank(t, "Strength of Earth Totem", enh.StrengthOfEarthTotem), time.Second},
		{"Stoneskin Totem", topRank(t, "Stoneskin Totem", enh.StoneskinTotem), time.Second},
		{"Tremor Totem", enh.TremorTotem, time.Second},
		{"Magma Totem", topRank(t, "Magma Totem", enh.MagmaTotem), time.Second},
		{"Flametongue Totem", topRank(t, "Flametongue Totem", enh.FlametongueTotem), time.Second},
		{"Searing Totem", topRank(t, "Searing Totem", enh.SearingTotem), time.Second},
		// Fire Nova needs the fire totem we just put down.
		{"Fire Nova", topRank(t, "Fire Nova", enh.FireNova), 1500 * time.Millisecond},
		{"Mana Spring Totem", topRank(t, "Mana Spring Totem", enh.ManaSpringTotem), time.Second},
		{"Healing Stream Totem", topRank(t, "Healing Stream Totem", enh.HealingStreamTotem), time.Second},
		{"Windfury Totem", topRank(t, "Windfury Totem", enh.WindfuryTotem), time.Second},
		{"Grace of Air Totem", topRank(t, "Grace of Air Totem", enh.GraceOfAirTotem), time.Second},
		{"Windwall Totem", topRank(t, "Windwall Totem", enh.WindwallTotem), time.Second},
	}
	for i, s := range spells {
		at(sim, 1+2*float64(i), func(sim *core.Simulation) {
			cast(sim, s.spell)
			checkGCD(sim, s.name, s.gcd)
		})
	}

	// Rage of the Farseer and Nature's Swiftness go while the GCD is up, and leave it alone.
	at(sim, 40, func(sim *core.Simulation) {
		cast(sim, topRank(t, "Earth Shock", enh.EarthShock))
		for _, id := range []int32{425336, 16188} {
			spell := enh.GetSpell(core.ActionID{SpellID: id})
			if spell == nil || !spell.Cast(sim, enh.CurrentTarget) {
				t.Fatalf("spell %d didn't cast during the GCD", id)
			}
			checkGCD(sim, spell.ActionID.String(), 1500*time.Millisecond)
		}
	})

	// Lightning Bolt with no haste, with 25% and 100% cast speed, and with 3 stacks of
	// Maelstrom Weapon.
	bolts := []struct {
		label    string
		at       float64
		speed    float64
		stacks   int32
		castTime time.Duration
		gcd      time.Duration
	}{
		{"Lightning Bolt", 50, 1, 0, 2500 * time.Millisecond, 2500 * time.Millisecond},
		{"Lightning Bolt at 25% haste", 55, 1.25, 0, 2000 * time.Millisecond, 2000 * time.Millisecond},
		{"Lightning Bolt at 100% haste", 60, 2, 0, 1250 * time.Millisecond, 1500 * time.Millisecond},
		{"Lightning Bolt with 3 Maelstrom Weapon stacks", 65, 1, 3, time.Second, 1500 * time.Millisecond},
	}
	for _, b := range bolts {
		at(sim, b.at, func(sim *core.Simulation) {
			// Nature's Swiftness from above would make the bolt instant, and each bolt that
			// hits can give a Maelstrom Weapon stack to the next one.
			enh.GetAuraByID(core.ActionID{SpellID: 16188}).Deactivate(sim)
			enh.MaelstromWeaponAura.Deactivate(sim)
			enh.MultiplyCastSpeed(b.speed)
			if b.stacks > 0 {
				enh.MaelstromWeaponAura.Activate(sim)
				enh.MaelstromWeaponAura.SetStacks(sim, b.stacks)
			}
			cast(sim, lb)
			if got := enh.Hardcast.Expires - sim.CurrentTime; (got - b.castTime).Abs() > time.Microsecond {
				t.Errorf("%s: casts in %s, want %s", b.label, got, b.castTime)
			}
			if got := enh.GCD.TimeToReady(sim); (got - b.gcd).Abs() > time.Microsecond {
				t.Errorf("%s: the GCD is up for %s, want %s", b.label, got, b.gcd)
			}
		})
		at(sim, b.at+3, func(sim *core.Simulation) {
			if b.speed != 1 {
				enh.MultiplyCastSpeed(1 / b.speed)
			}
		})
	}

	// Haste doesn't shorten the GCD of an instant or a totem.
	at(sim, 80, func(sim *core.Simulation) {
		enh.MultiplyCastSpeed(2)
		cast(sim, topRank(t, "Earth Shock", enh.EarthShock))
		checkGCD(sim, "Earth Shock at 100% haste", 1500*time.Millisecond)
	})
	at(sim, 82, func(sim *core.Simulation) {
		cast(sim, topRank(t, "Searing Totem", enh.SearingTotem))
		checkGCD(sim, "Searing Totem at 100% haste", time.Second)
		enh.MultiplyCastSpeed(0.5)
	})
	runSim(sim)
}
