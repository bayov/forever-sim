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
// imbue and an oil or stone. Windfury Weapon on the main hand turns Windfury Totem off for us,
// from our totem or another shaman's, as its tooltip says. Then our own Windfury Totem doesn't
// take the slot. Without it, the last weapon totem we put down holds the slot. When it goes,
// our other weapon totem takes the slot back if it still stands, and else another shaman's
// totem from the raid buffs.
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
			blocked := imbue == proto.WeaponImbue_WindfuryWeapon
			holds(sim, enh, "our Windfury Totem", map[string]bool{ownWindfury: !blocked, raidWindfury: blocked})
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
	if procs := windfuryTotemProcs(proto.WeaponImbue_WindfuryWeapon); procs != 0 {
		t.Errorf("Windfury Totem procced %d times with Windfury Weapon on the main hand", procs)
	}
	if procs := windfuryTotemProcs(proto.WeaponImbue_RockbiterWeapon); procs < 5 {
		t.Errorf("Windfury Totem procced %d times in 100 hits with Rockbiter, want about 20", procs)
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
}
