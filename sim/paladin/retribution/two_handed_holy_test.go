package retribution

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Phase45RetTalents with Two-Handed Weapon Specialization at 0 instead of 3.
const phase45RetNo2hSpec = "050050002-503-05233101201001351"

// TestTwoHandedSpecializationHolyReach checks the user ruling of 2026-11-15, after wowsims/forever's
// logs (EV-0064, EV-0022; their fca3d825d9): under Forever Two-Handed Weapon Specialization's Holy
// part reaches the Holy hits that roll on melee defense (the seal hits, Holy Strike) and Judgement
// of Command, and Consecration, 6% at 3/3 with a two-hander; the ranged judgements (Righteousness,
// Fury), Exorcism, Hammer of Wrath, Holy Wrath and Holy Shock take none. Physical keeps it
// school-wide.
func TestTwoHandedSpecializationHolyReach(t *testing.T) {
	items := map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: arcaniteReaper}
	holy := stats.SchoolIndexHoly
	for _, c := range []struct {
		name string
		seal proto.PaladinSeal
		id   int32
		want float64
	}{
		{"Seal of Righteousness's hit", proto.PaladinSeal_Righteousness, 25742, 1.06},
		{"Seal of Command's hit", proto.PaladinSeal_Command, 20424, 1.06},
		{"Judgement of Command", proto.PaladinSeal_Command, 20966, 1.06},
		{"Holy Strike", proto.PaladinSeal_Righteousness, 10333, 1.06},
		{"Consecration", proto.PaladinSeal_Righteousness, 20924, 1.06},
		{"Judgement of Righteousness", proto.PaladinSeal_Righteousness, 20286, 1},
		{"Exorcism", proto.PaladinSeal_Righteousness, 10314, 1},
		{"Hammer of Wrath", proto.PaladinSeal_Righteousness, 24239, 1},
		{"Holy Wrath", proto.PaladinSeal_Righteousness, 10318, 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			options := &proto.PaladinOptions{PrimarySeal: c.seal}
			_, with := retSimOptions(t, Phase45RetTalents, items, 1, proto.MobType_MobTypeUndead, options)
			_, without := retSimOptions(t, phase45RetNo2hSpec, items, 1, proto.MobType_MobTypeUndead, options)
			s, b := with.GetSpell(core.ActionID{SpellID: c.id}), without.GetSpell(core.ActionID{SpellID: c.id})
			if s == nil || b == nil {
				t.Fatalf("spell %d is not registered", c.id)
			}
			got := s.DamageMultiplier * with.PseudoStats.SchoolDamageDealtMultiplier[holy] /
				(b.DamageMultiplier * without.PseudoStats.SchoolDamageDealtMultiplier[holy])
			if got < c.want-0.0001 || got > c.want+0.0001 {
				t.Errorf("Two-Handed Weapon Specialization 3/3 on %s (%d) = %.4f, want %.2f", c.name, c.id, got, c.want)
			}
		})
	}

	_, with := retSimOptions(t, Phase45RetTalents, items, 1, proto.MobType_MobTypeUnknown, &proto.PaladinOptions{})
	_, without := retSimOptions(t, phase45RetNo2hSpec, items, 1, proto.MobType_MobTypeUnknown, &proto.PaladinOptions{})
	physical := stats.SchoolIndexPhysical
	if got := with.PseudoStats.SchoolDamageDealtMultiplier[physical] / without.PseudoStats.SchoolDamageDealtMultiplier[physical]; got < 1.0599 || got > 1.0601 {
		t.Errorf("Physical damage with 3/3 and a two-hander = %.4f, want 1.06", got)
	}
}
