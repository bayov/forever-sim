package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanStormstrikeOutcomes checks what a Stormstrike costs and leaves behind on each
// outcome.
//
// It always costs 125 mana and starts its 8 sec cooldown, and only a Stormstrike that lands
// puts the mark up. A blocked one lands. The beta log showed all of it (shaman_audit.md 5.8): a
// parried Stormstrike took 125 mana, and the 8 misses, 3 dodges and 19 parries put no mark up
// where the one block did. We cast 400 from in front of a level 63 boss, so every outcome
// comes up.
func TestOrcShamanStormstrikeOutcomes(t *testing.T) {
	player := newOrcShaman(60, stormstrike, &proto.EnhancementShaman_Options{})
	player.InFrontOfTarget = true
	for range proto.ItemSlot_ItemSlotMainHand {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
	}
	player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: darkEdge})
	sim, enh := newShamanSim(player, &proto.Debuffs{}, 10)
	mark := enh.CurrentTarget.GetAuraByID(enh.Stormstrike.ActionID)
	metrics := &enh.Stormstrike.SpellMetrics[enh.CurrentTarget.UnitIndex]

	outcome := func(before core.SpellMetrics) string {
		switch {
		case metrics.Misses > before.Misses:
			return "miss"
		case metrics.Dodges > before.Dodges:
			return "dodge"
		case metrics.Parries > before.Parries:
			return "parry"
		case metrics.Blocks > before.Blocks:
			return "block"
		case metrics.Crits > before.Crits:
			return "crit"
		case metrics.Hits > before.Hits:
			return "hit"
		}
		return "nothing"
	}
	landed := map[string]bool{"miss": false, "dodge": false, "parry": false, "block": true, "crit": true, "hit": true}

	seen := map[string]int{}
	at(sim, 1, func(sim *core.Simulation) {
		for range 400 {
			mark.Deactivate(sim)
			enh.Stormstrike.CD.Reset()
			enh.AddMana(sim, enh.MaxMana()-enh.CurrentMana(), enh.GetManaNotCastingMetrics())

			before, mana := *metrics, enh.CurrentMana()
			castNow(t, sim, enh, enh.Stormstrike)
			o := outcome(before)
			seen[o]++

			if spent := mana - enh.CurrentMana(); spent != 125 {
				t.Fatalf("a %s cost %.1f mana, want 125", o, spent)
			}
			if readyAt := enh.Stormstrike.CD.ReadyAt(); readyAt != sim.CurrentTime+8*time.Second {
				t.Fatalf("after a %s the cooldown ends at %s, want %s", o, readyAt, sim.CurrentTime+8*time.Second)
			}
			if mark.IsActive() != landed[o] {
				t.Fatalf("a %s leaves the mark up: %t, want %t", o, mark.IsActive(), landed[o])
			}
		}
	})
	runSim(sim)

	for o := range landed {
		if seen[o] == 0 {
			t.Errorf("no %s in 400 Stormstrikes (%v)", o, seen)
		}
	}
}
