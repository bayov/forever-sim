package core

import "testing"

// Which landed hits roll Windfury Totem under Forever since Blizzard's hotfixes of November 13,
// 2026: every main hand melee hit, white or special, and no off-hand one or seal proc.
func TestWindfuryTotemRollsOnForever(t *testing.T) {
	cases := []struct {
		name     string
		spellID  int32
		procMask ProcMask
		flags    SpellFlag
		want     bool
	}{
		{"white main hand swing", 0, ProcMaskMeleeMHAuto, 0, true},
		{"white off-hand swing", 0, ProcMaskMeleeOHAuto, 0, false},
		{"Heroic Strike", 11567, ProcMaskMeleeMHSpecial | ProcMaskMeleeMHAuto, 0, true},
		{"Maul", 9881, ProcMaskMeleeMHSpecial, 0, true},
		{"Sinister Strike", 11294, ProcMaskMeleeMHSpecial, 0, true},
		{"Stormstrike", 17364, ProcMaskMeleeMHSpecial, 0, true},
		{"Holy Strike", 10333, ProcMaskMeleeMHSpecial, 0, true},
		{"Bloodthirst", 23894, ProcMaskMeleeMHSpecial, 0, true},
		{"Whirlwind main hand", 1680, ProcMaskMeleeMHSpecial, 0, true},
		{"Whirlwind off-hand", 1680, ProcMaskMeleeOHSpecial, 0, false},
		{"Judgement of Command", 20966, ProcMaskMeleeMHSpecial, 0, true},
		{"Windfury Weapon main hand attack", 16362, ProcMaskMeleeMHSpecial, 0, true},
		{"Seal of Righteousness proc", 20293, ProcMaskMeleeMHSpecial, SpellFlagNoWindfuryTotem, false},
		{"Seal of Command proc", 20424, ProcMaskMeleeMHSpecial | ProcMaskMeleeProc, SpellFlagNoWindfuryTotem, false},
		{"Mutilate off-hand hit", 1241584, ProcMaskMeleeOHSpecial, 0, false},
		{"Judgement of Righteousness (ranged)", 20286, ProcMaskRangedSpecial, 0, false},
	}
	for _, c := range cases {
		spell := &Spell{ActionID: ActionID{SpellID: c.spellID}, ProcMask: c.procMask, Flags: c.flags}
		if got := windfuryTotemRollsOnForever(spell); got != c.want {
			t.Errorf("%s: rolls Windfury Totem = %v, want %v", c.name, got, c.want)
		}
	}
}
