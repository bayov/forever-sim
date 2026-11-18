package core

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

// setReportSwitches sets the given DASH_* variables for one test and reloads the switches. Its
// cleanup reloads them again once the variables are restored.
func setReportSwitches(t *testing.T, kv ...string) {
	t.Helper()
	t.Cleanup(ReloadReportSwitches) // registered first, so it runs after t.Setenv restores the variables
	for _, k := range []string{"DASH_NAMES", "DASH_SPLITS", "DASH_RACIALS_OFF", "DASH_RACIALS_OFF_CLASS", "DASH_ASSUME"} {
		t.Setenv(k, "")
	}
	for i := 0; i+1 < len(kv); i += 2 {
		t.Setenv(kv[i], kv[i+1])
	}
	ReloadReportSwitches()
}

func TestReportSwitchesUnset(t *testing.T) {
	setReportSwitches(t)
	s := reportSwitches
	if s.names != "" || s.splits != 0 || len(s.racialsOff) != 0 || s.racialsOffClass != "" || len(s.assume) != 0 {
		t.Fatalf("unset switches parsed as %+v", s)
	}
	if s.bleedTickCritSet {
		t.Errorf("unset: bleed_tick_crit set %v", s.bleedTickCritSet)
	}
	if s.rageScaling != ForeverNormalizedRageScaling || s.damageTakenRageFactor != ForeverDamageTakenRageFactor {
		t.Errorf("unset: rage scaling %v and damage taken factor %v, want master's %v and %v",
			s.rageScaling, s.damageTakenRageFactor, ForeverNormalizedRageScaling, ForeverDamageTakenRageFactor)
	}
	if got := concurrentSplits(); got != int32(runtime.NumCPU()) {
		t.Errorf("unset: %d splits, want one per CPU (%d)", got, runtime.NumCPU())
	}
	if DashRacialOff("bloodfury") || DashRacialOffFor(proto.Class_ClassWarrior, "bloodfury") {
		t.Error("unset: a racial part is off")
	}
	if v, ok := AssumeOk("warr_rend_ap"); ok || Assume("warr_rend_ap", 0.05) != 0.05 {
		t.Errorf("unset: warr_rend_ap reads %v (%v), want the default", v, ok)
	}
	if !AssumeBloodFuryWeapon() {
		t.Error("unset: Blood Fury lost its weapon damage")
	}
}

func TestDashSplits(t *testing.T) {
	for _, c := range []struct {
		env  string
		want int32
	}{{"4", 4}, {"1", 1}, {"0", int32(runtime.NumCPU())}, {"-2", int32(runtime.NumCPU())}, {"four", int32(runtime.NumCPU())}} {
		setReportSwitches(t, "DASH_SPLITS", c.env)
		if got := concurrentSplits(); got != c.want {
			t.Errorf("DASH_SPLITS=%q: %d splits, want %d", c.env, got, c.want)
		}
	}
}

func TestDashNames(t *testing.T) {
	unit := &Unit{}
	unit.auras = []*Aura{
		{Label: "Blood Fury", ActionID: ActionID{SpellID: 20572}},
		{Label: "Tag", ActionID: ActionID{ItemID: 7, Tag: 2}},
	}
	env := &Environment{AllUnits: []*Unit{unit}}
	path := filepath.Join(t.TempDir(), "names.tsv")

	setReportSwitches(t)
	dumpAuraNames(env)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("unset: %s exists (%v)", path, err)
	}

	setReportSwitches(t, "DASH_NAMES", path)
	dumpAuraNames(env)
	dumpAuraNames(env) // appends, one copy per sim
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Repeat("20572\t0\t0\t0\tBlood Fury\n0\t7\t0\t2\tTag\n", 2)
	if string(b) != want {
		t.Errorf("DASH_NAMES file is %q, want %q", b, want)
	}
}

func TestDashRacialsOff(t *testing.T) {
	setReportSwitches(t, "DASH_RACIALS_OFF", " BloodFury, beast ")
	for part, want := range map[string]bool{"bloodfury": true, "beast": true, "berserking": false} {
		if DashRacialOff(part) != want || DashRacialOffFor(proto.Class_ClassHunter, part) != want {
			t.Errorf("DASH_RACIALS_OFF=BloodFury,beast: %s off is %v, want %v", part, DashRacialOff(part), want)
		}
	}
	setReportSwitches(t, "DASH_RACIALS_OFF", "all")
	if !DashRacialOff("leyline") || !DashRacialOff("shadowguard") {
		t.Error("DASH_RACIALS_OFF=all keeps a part")
	}
}

func TestDashRacialsOffClass(t *testing.T) {
	setReportSwitches(t, "DASH_RACIALS_OFF", "all", "DASH_RACIALS_OFF_CLASS", "Priest")
	if !DashRacialOffFor(proto.Class_ClassPriest, "chastise") {
		t.Error("DASH_RACIALS_OFF_CLASS=Priest keeps the priest's racial")
	}
	if DashRacialOffFor(proto.Class_ClassWarrior, "bloodfury") {
		t.Error("DASH_RACIALS_OFF_CLASS=Priest takes the warrior's racial")
	}
	// Alone, it does nothing.
	setReportSwitches(t, "DASH_RACIALS_OFF_CLASS", "Priest")
	if DashRacialOffFor(proto.Class_ClassPriest, "chastise") {
		t.Error("DASH_RACIALS_OFF_CLASS alone takes a racial")
	}
}

func TestDashAssume(t *testing.T) {
	setReportSwitches(t, "DASH_ASSUME", "bleed_tick_crit=0.5,warr_rage_s=0.2,warr_dt_rage_k=3,orc_bf_weapon=0,junk,bad=x")
	s := reportSwitches
	if !s.bleedTickCritSet || s.bleedTickCrit != 0.5 || s.rageScaling != 0.2 || s.damageTakenRageFactor != 3 {
		t.Errorf("DASH_ASSUME parsed as %+v", s)
	}
	if AssumeBloodFuryWeapon() {
		t.Error("orc_bf_weapon=0 keeps Blood Fury's weapon damage")
	}
	if _, ok := AssumeOk("bad"); ok {
		t.Error("a knob with no number was kept")
	}
	if got := ForeverNormalizedRageLevelFactor(60); got != 1.2 {
		t.Errorf("warr_rage_s=0.2: level 60 rage factor %v, want 1.2", got)
	}
}
