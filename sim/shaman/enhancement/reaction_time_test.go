package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanReactionTime checks that the rotation sees a proc only after the player's
// reaction time, and our own casts at once.
//
// Our swings give Maelstrom Weapon stacks, and the rotation casts Lightning Bolt at 5. So the
// bolt goes 200 ms after the 5th stack, and not before. Blood Fury goes the moment Rage of the
// Farseer is up, because we cast Rage of the Farseer ourselves and queue Blood Fury after it.
func TestOrcShamanReactionTime(t *testing.T) {
	const reactionTime = 200 * time.Millisecond
	spellID := func(id int32) *proto.ActionID { return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}} }
	cast := func(id int32, condition *proto.APLValue) *proto.APLListItem {
		return &proto.APLListItem{Action: &proto.APLAction{
			Condition: condition,
			Action:    &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: spellID(id)}},
		}}
	}
	farseerUp := &proto.APLValue{Value: &proto.APLValue_AuraIsActive{AuraIsActive: &proto.APLValueAuraIsActive{AuraId: spellID(425336)}}}
	fiveStacks := &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{
		Op:  proto.APLValueCompare_OpGe,
		Lhs: &proto.APLValue{Value: &proto.APLValue_AuraNumStacks{AuraNumStacks: &proto.APLValueAuraNumStacks{AuraId: spellID(408505)}}},
		Rhs: &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: "5"}}},
	}}}

	// Maelstrom Weapon 5/5 and Rage of the Farseer.
	player := newOrcShaman(60, "-000000000000000051", &proto.EnhancementShaman_Options{})
	player.ReactionTimeMs = int32(reactionTime / time.Millisecond)
	player.Rotation.PriorityList = []*proto.APLListItem{
		cast(20572, farseerUp),  // Blood Fury
		cast(425336, nil),       // Rage of the Farseer
		cast(15208, fiveStacks), // Lightning Bolt rank 10
	}
	// Unarmed swings fill Maelstrom Weapon slowly, so we fight for 10 min.
	sim, enh := newShamanSim(player, &proto.Debuffs{}, 600)
	metrics := enh.GetManaNotCastingMetrics()
	mw := enh.MaelstromWeaponAura

	// We note when the 5th stack comes and how long the bolt that uses it up takes.
	var fullAt time.Duration
	var waits []time.Duration
	onStacksChange := mw.OnStacksChange
	mw.OnStacksChange = func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
		if onStacksChange != nil {
			onStacksChange(aura, sim, oldStacks, newStacks)
		}
		switch {
		case newStacks == 5:
			fullAt = sim.CurrentTime
		case oldStacks == 5 && newStacks == 0:
			waits = append(waits, sim.CurrentTime-fullAt)
			// We never run out of mana for the next bolt.
			enh.AddMana(sim, enh.MaxMana(), metrics)
		}
	}

	at(sim, 1, func(sim *core.Simulation) {
		farseer := enh.GetAuraByID(core.ActionID{SpellID: 425336})
		bloodFury := enh.GetAuraByID(core.ActionID{SpellID: 20572})
		if !farseer.IsActive() || !bloodFury.IsActive() {
			t.Fatalf("Rage of the Farseer or Blood Fury didn't go at the pull")
		}
		if farseer.StartedAt() != bloodFury.StartedAt() {
			t.Errorf("Blood Fury went at %s, %s after Rage of the Farseer", bloodFury.StartedAt(), bloodFury.StartedAt()-farseer.StartedAt())
		}
	})
	runSim(sim)

	if len(waits) < 10 {
		t.Fatalf("only %d bolts used up 5 Maelstrom Weapon stacks", len(waits))
	}
	// We look again every 50 ms while the GCD is free and nothing can go.
	for _, wait := range waits {
		if wait < reactionTime || wait > reactionTime+50*time.Millisecond {
			t.Errorf("a bolt went %s after the 5th stack, want 200 to 250 ms", wait)
		}
	}
}
