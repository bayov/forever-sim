package dpswarrior

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
)

// TestDashRacialsOffBloodFury: DASH_RACIALS_OFF=bloodfury (or all) leaves the Orc's Blood Fury
// out, DASH_RACIALS_OFF_CLASS of another class keeps it, and with neither set the Orc has it.
func TestDashRacialsOffBloodFury(t *testing.T) {
	for _, c := range []struct {
		off, class string
		want       bool
	}{
		{"", "", true},
		{"bloodfury", "", false},
		{"all", "", false},
		{"berserking", "", true},
		{"all", "Warrior", false},
		{"all", "Priest", true},
	} {
		t.Run(c.off+"/"+c.class, func(t *testing.T) {
			t.Cleanup(core.ReloadReportSwitches)
			t.Setenv("DASH_RACIALS_OFF", c.off)
			t.Setenv("DASH_RACIALS_OFF_CLASS", c.class)
			_, dw := assumeWarrior(t, "") // reloads the switches
			if got := dw.GetWarrior().GetAura("Blood Fury") != nil; got != c.want {
				t.Errorf("Blood Fury registered %v, want %v", got, c.want)
			}
		})
	}
}
