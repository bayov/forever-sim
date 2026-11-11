package enhancement

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanWeaponTotemSlot checks which totem buff our main hand holds (shaman_audit.md
// 7.18).
//
// The weapon holds one Windfury or Flametongue Totem buff, next to its permanent enchant, our
// imbue and an oil or stone. The last weapon totem placed holds the slot (the user saw the two
// totems replace each other on the beta, 2026-11-08). Since Blizzard's class tuning hotfixes of
// 2026-11-10 ("Weapon imbues and Weapon totems can now be used at the same time") Windfury Totem
// procs next to Windfury Weapon on the main hand too, as often as next to Rockbiter: 20% of 100
// landed white hits 2 sec apart, so about 20 (we want 5 to 40). Before it Windfury Weapon turned
// the totem off. When the totem in the slot goes, our other weapon totem takes the slot back if
// it still stands, and else another shaman's totem from the raid buffs.
func TestOrcShamanWeaponTotemSlot(t *testing.T) {
	const (
		ownWindfury    = "Windfury Totem (Rank 3)"
		raidWindfury   = "Windfury Totem Raid"
		ownFlametongue = "Flametongue Totem (Rank 4)"
	)
	start := func(imbue proto.WeaponImbue, raid proto.TotemWeaponBuff) (*core.Simulation, *EnhancementShaman) {
		sim, enh := weaponSim{level: 60, weapon: &proto.ItemSpec{Id: darkEdge}, imbue: imbue, totem: raid, seconds: 300}.start()
		sim.PrePull()
		at(sim, 0.5, enh.AutoAttacks.CancelAutoSwing)
		return sim, enh
	}
	place := func(sim *core.Simulation, enh *EnhancementShaman, spell *core.Spell) {
		enh.AddMana(sim, enh.MaxMana()-enh.CurrentMana(), enh.GetManaNotCastingMetrics())
		castNow(t, sim, enh, spell)
	}
	// holds checks which of the totem buffs are up.
	holds := func(sim *core.Simulation, enh *EnhancementShaman, after string, want map[string]bool) {
		for label, up := range want {
			aura := enh.GetAura(label)
			if aura == nil {
				t.Fatalf("no %s aura", label)
			}
			if aura.IsActive() != up {
				t.Errorf("after %s: %s up: %v, want %v", after, label, aura.IsActive(), up)
			}
		}
	}

	// We put down our Windfury Totem next to another shaman's, and land 100 white hits 2 sec
	// apart. A Windfury Totem proc moves our next swing to right now.
	windfuryTotemProcs := func(imbue proto.WeaponImbue) int {
		sim, enh := start(imbue, proto.TotemWeaponBuff_TotemWeaponBuffWindfury)
		procs := 0
		at(sim, 1, func(sim *core.Simulation) {
			place(sim, enh, topRank(t, "Windfury Totem", enh.WindfuryTotem))
			holds(sim, enh, "our Windfury Totem", map[string]bool{ownWindfury: true, raidWindfury: false})
		})
		for i := range 100 {
			at(sim, 2+2*float64(i), func(sim *core.Simulation) {
				enh.AutoAttacks.MHAuto().CalcAndDealDamage(sim, enh.CurrentTarget, 100, forcedOutcome(core.OutcomeHit))
				if enh.AutoAttacks.MainhandSwingAt() == sim.CurrentTime {
					procs++
				}
			})
		}
		runSim(sim)
		return procs
	}
	for _, imbue := range []proto.WeaponImbue{proto.WeaponImbue_WindfuryWeapon, proto.WeaponImbue_RockbiterWeapon} {
		if procs := windfuryTotemProcs(imbue); procs < 5 || procs > 40 {
			t.Errorf("Windfury Totem procced %d times in 100 hits with %s, want about 20", procs, imbue)
		}
	}

	// Our Flametongue Totem takes the slot from another shaman's Windfury Totem, and gives it
	// back when a Searing Totem takes its place.
	sim, enh := start(proto.WeaponImbue_RockbiterWeapon, proto.TotemWeaponBuff_TotemWeaponBuffWindfury)
	at(sim, 1, func(sim *core.Simulation) {
		place(sim, enh, topRank(t, "Flametongue Totem", enh.FlametongueTotem))
		holds(sim, enh, "our Flametongue Totem", map[string]bool{ownFlametongue: true, raidWindfury: false})
		place(sim, enh, topRank(t, "Searing Totem", enh.SearingTotem))
		holds(sim, enh, "our Searing Totem", map[string]bool{ownFlametongue: false, raidWindfury: true})
	})
	runSim(sim)

	// Our Windfury Totem, then our Flametongue Totem. When a Searing Totem takes Flametongue's
	// place, Windfury Totem still stands and takes the slot back. A Grace of Air Totem in its
	// place leaves the slot empty.
	sim, enh = start(proto.WeaponImbue_RockbiterWeapon, proto.TotemWeaponBuff_TotemWeaponBuffNone)
	at(sim, 1, func(sim *core.Simulation) {
		place(sim, enh, topRank(t, "Windfury Totem", enh.WindfuryTotem))
		holds(sim, enh, "our Windfury Totem", map[string]bool{ownWindfury: true})
		place(sim, enh, topRank(t, "Flametongue Totem", enh.FlametongueTotem))
		holds(sim, enh, "our Flametongue Totem", map[string]bool{ownWindfury: false, ownFlametongue: true})
		place(sim, enh, topRank(t, "Searing Totem", enh.SearingTotem))
		holds(sim, enh, "our Searing Totem", map[string]bool{ownWindfury: true, ownFlametongue: false})
		place(sim, enh, topRank(t, "Grace of Air Totem", enh.GraceOfAirTotem))
		holds(sim, enh, "our Grace of Air Totem", map[string]bool{ownWindfury: false, ownFlametongue: false})
	})
	runSim(sim)

	// With Windfury Weapon on, our Windfury Totem takes the slot from our Flametongue Totem, and
	// Flametongue Totem put down again takes it back.
	sim, enh = start(proto.WeaponImbue_WindfuryWeapon, proto.TotemWeaponBuff_TotemWeaponBuffNone)
	at(sim, 1, func(sim *core.Simulation) {
		place(sim, enh, topRank(t, "Flametongue Totem", enh.FlametongueTotem))
		holds(sim, enh, "our Flametongue Totem", map[string]bool{ownFlametongue: true})
		place(sim, enh, topRank(t, "Windfury Totem", enh.WindfuryTotem))
		holds(sim, enh, "our Windfury Totem", map[string]bool{ownWindfury: true, ownFlametongue: false})
		place(sim, enh, topRank(t, "Flametongue Totem", enh.FlametongueTotem))
		holds(sim, enh, "our Flametongue Totem again", map[string]bool{ownWindfury: false, ownFlametongue: true})
	})
	runSim(sim)
}
