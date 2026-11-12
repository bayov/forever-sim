package core

import "testing"

// Which landed hits roll Windfury Totem under Forever, as our live logs of Nov 6-11 show them.
func TestWindfuryTotemRollsOnForever(t *testing.T) {
	cases := []struct {
		name     string
		spellID  int32
		procMask ProcMask
		want     bool
	}{
		{"white main hand swing", 0, ProcMaskMeleeMHAuto, true},
		{"white off-hand swing", 0, ProcMaskMeleeOHAuto, false},
		{"Heroic Strike", 11567, ProcMaskMeleeMHSpecial | ProcMaskMeleeMHAuto, true},
		{"Maul", 9881, ProcMaskMeleeMHSpecial, true},
		{"Raptor Strike", 14266, ProcMaskMeleeMHSpecial, true},
		{"Sunder Armor", 11597, ProcMaskMeleeMHSpecial, true},
		{"Execute", 20662, ProcMaskMeleeMHSpecial, true},
		{"Sinister Strike", 11294, ProcMaskMeleeMHSpecial, true},
		{"Eviscerate", 11300, ProcMaskMeleeMHSpecial, true},
		{"Stormstrike", 17364, ProcMaskMeleeMHSpecial, false},
		{"Holy Strike", 10333, ProcMaskMeleeMHSpecial, false},
		{"Bloodthirst", 23894, ProcMaskMeleeMHSpecial, false},
		{"Whirlwind", 1680, ProcMaskMeleeMHSpecial, false},
		{"Claw", 9850, ProcMaskMeleeMHSpecial, false},
		{"Seal of Command proc", 20424, ProcMaskMeleeMHSpecial, false},
		{"Mutilate off-hand hit", 1241584, ProcMaskMeleeOHSpecial, false},
	}
	for _, c := range cases {
		spell := &Spell{ActionID: ActionID{SpellID: c.spellID}, ProcMask: c.procMask}
		if got := windfuryTotemRollsOnForever(spell); got != c.want {
			t.Errorf("%s: rolls Windfury Totem = %v, want %v", c.name, got, c.want)
		}
	}
}
