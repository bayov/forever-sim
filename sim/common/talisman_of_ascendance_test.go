//go:build with_db

package common_test

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Talisman of Ascendance: every spell we cast while the use is up, a damage spell or not, adds a
// stack of 40 spell damage after it lands, up to 5, and the 6th cast ends it (28200's 6 charges and
// ProcTypeMask 87376, item_effects.go).
func TestTalismanOfAscendanceCountsEveryCast(t *testing.T) {
	s, m := newForeverMageWearing(t, proto.ItemSlot_ItemSlotTrinket1, common.TalismanOfAscendance)
	use := m.GetSpell(core.ActionID{ItemID: common.TalismanOfAscendance})
	if use == nil {
		t.Fatal("no Talisman of Ascendance use")
	}
	if got := use.CD.Duration; got != time.Minute {
		t.Errorf("Talisman of Ascendance cooldown %v, want 1m", got)
	}
	aura := m.GetAura("Ascendance")

	// A spell of ours with no damage or healing proc mask (an armor, Arcane Intellect, ...), one
	// with spell damage, and one that isn't a cast of ours at all.
	var noDamage, damage, notOurs *core.Spell
	for _, spell := range m.Spellbook {
		switch {
		case spell.Flags.Matches(core.SpellFlagNoOnCastComplete):
		case !spell.Flags.Matches(core.SpellFlagAPL):
			notOurs = spell
		case !spell.ProcMask.Matches(core.ProcMaskSpellDamage|core.ProcMaskSpellHealing) && noDamage == nil:
			noDamage = spell
		case spell.ProcMask.Matches(core.ProcMaskSpellDamage) && damage == nil:
			damage = spell
		}
	}
	if noDamage == nil || damage == nil {
		t.Fatalf("the mage has no spell without a damage proc mask (%v) or none with one (%v)", noDamage, damage)
	}

	base := m.GetStat(stats.SpellDamage)
	if !use.Cast(s, m.CurrentTarget) || !aura.IsActive() {
		t.Fatal("the use put no Ascendance up")
	}
	if notOurs != nil {
		m.OnCastComplete(s, notOurs)
		if aura.GetStacks() != 0 {
			t.Errorf("%v, not a cast of ours, gave a stack", notOurs.ActionID)
		}
	}
	casts := []*core.Spell{noDamage, damage, noDamage, damage, noDamage}
	for i, spell := range casts {
		m.OnCastComplete(s, spell)
		if got, want := aura.GetStacks(), int32(i+1); got != want {
			t.Fatalf("after cast %d (%v) %d stacks, want %d", i+1, spell.ActionID, got, want)
		}
		if got, want := m.GetStat(stats.SpellDamage)-base, 40*float64(i+1); got != want {
			t.Errorf("after cast %d spell damage +%.0f, want +%.0f", i+1, got, want)
		}
	}
	m.OnCastComplete(s, noDamage)
	if aura.IsActive() {
		t.Error("Ascendance still up after the 6th cast")
	}
	if got := m.GetStat(stats.SpellDamage); got != base {
		t.Errorf("spell damage %.0f once Ascendance went, want %.0f", got, base)
	}
}
