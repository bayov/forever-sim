package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanCastSwingTimer checks what our casts do to the swing timer at level 60.
//
// We don't swing during a cast, and the swing timer starts over when it ends. So the next
// swing comes one full swing after the cast. An instant Lightning Bolt (5 Maelstrom Weapon
// stacks or Nature's Swiftness) leaves the swing timer alone, like a shock (the user,
// 2026-10-08).
func TestOrcShamanCastSwingTimer(t *testing.T) {
	// Maelstrom Weapon 5/5 and Nature's Swiftness.
	sim, enh := newShamanSim(newOrcShaman(60, "-00000000000000005-00000000000001", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 80)
	metrics := enh.GetManaNotCastingMetrics()
	lb := topRank(t, "Lightning Bolt", enh.LightningBolt)
	cl := topRank(t, "Chain Lightning", enh.ChainLightning)
	es := topRank(t, "Earth Shock", enh.EarthShock)
	mw := enh.MaelstromWeaponAura
	naturesSwiftness := enh.GetSpell(core.ActionID{SpellID: 16188})

	cases := []struct {
		label  string
		spell  *core.Spell
		stacks int32
		ns     bool
		// The cast time, or -1 when the swing timer should stay as it was.
		castTime time.Duration
	}{
		{"Lightning Bolt", lb, 0, false, 2500 * time.Millisecond},
		{"Chain Lightning", cl, 0, false, 2000 * time.Millisecond},
		{"Lightning Bolt with 3 Maelstrom Weapon stacks", lb, 3, false, time.Second},
		{"Lightning Bolt with 5 Maelstrom Weapon stacks", lb, 5, false, -1},
		{"Lightning Bolt with Nature's Swiftness", lb, 0, true, -1},
		{"Earth Shock", es, 0, false, -1},
	}
	for i, c := range cases {
		at(sim, 5+10*float64(i), func(sim *core.Simulation) {
			// We cast 0.4 sec before the next swing, so the swing timer starting over shows.
			castAt := enh.AutoAttacks.MainhandSwingAt() - 400*time.Millisecond
			for castAt <= sim.CurrentTime {
				castAt += enh.AutoAttacks.MainhandSwingSpeed()
			}
			sim.AddPendingAction(&core.PendingAction{NextActionAt: castAt, OnAction: func(sim *core.Simulation) {
				enh.AddMana(sim, enh.MaxMana(), metrics)
				// Our swings give Maelstrom Weapon stacks, so we set them by hand.
				mw.Deactivate(sim)
				if c.stacks > 0 {
					mw.Activate(sim)
					mw.SetStacks(sim, c.stacks)
				}
				if c.ns && !naturesSwiftness.Cast(sim, enh.CurrentTarget) {
					t.Fatalf("Nature's Swiftness didn't cast")
				}
				swingAt := enh.AutoAttacks.MainhandSwingAt()
				if swingAt-sim.CurrentTime != 400*time.Millisecond {
					t.Fatalf("%s: the next swing is %s away, want 0.4 sec", c.label, swingAt-sim.CurrentTime)
				}
				castNow(t, sim, enh, c.spell)

				want := swingAt
				if c.castTime >= 0 {
					want = sim.CurrentTime + c.castTime + enh.AutoAttacks.MainhandSwingSpeed()
				}
				if got := enh.AutoAttacks.MainhandSwingAt(); (got - want).Abs() > time.Microsecond {
					t.Errorf("%s: the next swing comes at %s, want %s (it was %s)", c.label, got, want, swingAt)
				}
			}})
		})
	}
	runSim(sim)
}
