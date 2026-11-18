package core

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Report switches are the settings the DPS Report (tools/dps_report) gives the sim through
// environment variables. Each one does nothing unless its variable is set, so a sim run
// without them is the plain sim. This file is the only place that reads them. They are read
// once, when the package loads, and call sites read the parsed values in reportSwitches.

type reportSwitchSet struct {
	// DASH_NAMES=<file>: for the DPS Report only. When set, each sim's finalize appends every
	// unit's auras to the file, one "spellId itemId otherId tag label" line each. The report
	// uses it to label auras and procs. Unset, nothing is written.
	names string

	// DASH_SPLITS=<n>: for the DPS Report only. When set, a sim's iterations run in n
	// concurrent parts instead of one per CPU, so the report's raw outputs are the same on
	// every machine. Unset (or not a positive number), the sim splits one part per CPU.
	splits int32

	// DASH_RACIALS_OFF=all or =<part>[,<part>]: for the DPS Report's race breakdown only. When
	// set, it leaves those Forever racial parts out (see DashRacialOff for the part names).
	// Unset, every racial applies.
	racialsOff map[string]bool

	// DASH_RACIALS_OFF_CLASS=<Class>: for the DPS Report's healing and tanking race benches
	// only. When set, DASH_RACIALS_OFF applies to the players of that class alone (Priest,
	// Paladin, ...), so the bench raid's other players keep their racials. Unset,
	// DASH_RACIALS_OFF applies to every class.
	racialsOffClass string

	// DASH_ASSUME="key=value,..." : for the DPS Report's assumption scenarios only
	// (tools/dps_report/assumptions_def.py lists the keys). When set, each key overrides the
	// constant it names. Unset, every knob is its default, which is master's constant.
	assume map[string]float64

	// The knobs read for every hit, tick or cast, parsed once here.
	bleedTickCrit         float64 // bleed_tick_crit, when bleedTickCritSet
	bleedTickCritSet      bool    //
	rageScaling           float64 // warr_rage_s, default ForeverNormalizedRageScaling
	damageTakenRageFactor float64 // warr_dt_rage_k, default ForeverDamageTakenRageFactor
}

var reportSwitches = parseReportSwitches(os.Getenv)

// ReloadReportSwitches reads the report switches from the environment again. Tests only: a
// sim reads them once, when the package loads.
func ReloadReportSwitches() {
	reportSwitches = parseReportSwitches(os.Getenv)
}

func parseReportSwitches(getenv func(string) string) reportSwitchSet {
	s := reportSwitchSet{
		names:           getenv("DASH_NAMES"),
		racialsOff:      map[string]bool{},
		racialsOffClass: strings.ToLower(strings.TrimSpace(getenv("DASH_RACIALS_OFF_CLASS"))),
		assume:          map[string]float64{},
	}
	if n, err := strconv.Atoi(getenv("DASH_SPLITS")); err == nil && n > 0 {
		s.splits = int32(n)
	}
	for _, part := range strings.Split(getenv("DASH_RACIALS_OFF"), ",") {
		if part = strings.TrimSpace(strings.ToLower(part)); part != "" {
			s.racialsOff[part] = true
		}
	}
	for _, kv := range strings.Split(getenv("DASH_ASSUME"), ",") {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			s.assume[strings.TrimSpace(k)] = f
		}
	}
	s.bleedTickCrit, s.bleedTickCritSet = s.assume["bleed_tick_crit"]
	s.rageScaling = s.assumeOr("warr_rage_s", ForeverNormalizedRageScaling)
	s.damageTakenRageFactor = s.assumeOr("warr_dt_rage_k", ForeverDamageTakenRageFactor)
	return s
}

func (s *reportSwitchSet) assumeOr(key string, def float64) float64 {
	if v, ok := s.assume[key]; ok {
		return v
	}
	return def
}

// concurrentSplits is how many parts runSimConcurrent splits a sim's iterations into:
// DASH_SPLITS when set, else one per CPU.
func concurrentSplits() int32 {
	if reportSwitches.splits > 0 {
		return reportSwitches.splits
	}
	return int32(runtime.NumCPU())
}

// dumpAuraNames appends every unit's auras to the DASH_NAMES file, when it is set.
func dumpAuraNames(env *Environment) {
	if reportSwitches.names == "" {
		return
	}
	f, err := os.OpenFile(reportSwitches.names, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	for _, u := range env.AllUnits {
		for _, a := range u.auras {
			fmt.Fprintf(f, "%d\t%d\t%d\t%d\t%s\n", a.ActionID.SpellID, a.ActionID.ItemID, a.ActionID.OtherID, a.ActionID.Tag, a.Label)
		}
	}
}

// DashRacialOff reports whether DASH_RACIALS_OFF leaves the named racial part out. The parts
// are spirit, ruthless, precision, beast, dodge, elune, shadowmeld, mind, eureka, command,
// bloodfury, touch, endurance, berserking, windblessed, insight, leyline, and the priest
// racial spells darksacrifice, starshards, chastise and shadowguard. "all" leaves every one out.
func DashRacialOff(name string) bool {
	return reportSwitches.racialsOff["all"] || reportSwitches.racialsOff[name]
}

// DashRacialOffFor reports whether DASH_RACIALS_OFF leaves the named racial part out for a
// character of this class, which DASH_RACIALS_OFF_CLASS can limit to one class.
func DashRacialOffFor(class proto.Class, name string) bool {
	if reportSwitches.racialsOffClass != "" && strings.ToLower(strings.TrimPrefix(class.String(), "Class")) != reportSwitches.racialsOffClass {
		return false
	}
	return DashRacialOff(name)
}

// AssumeOk returns the DASH_ASSUME override for key, and whether there is one. Read it when a
// spell or aura is set up, not for every hit or tick.
func AssumeOk(key string) (float64, bool) {
	v, ok := reportSwitches.assume[key]
	return v, ok
}

// Assume returns the DASH_ASSUME override for key, or def when there is none. Read it when a
// spell or aura is set up, not for every hit or tick.
func Assume(key string, def float64) float64 {
	return reportSwitches.assumeOr(key, def)
}

// AssumeForTest reloads the report switches with DASH_ASSUME set to env. Tests only.
func AssumeForTest(env string) {
	os.Setenv("DASH_ASSUME", env)
	ReloadReportSwitches()
}

// AssumeBloodFuryWeapon is the report assumption orc_bf_weapon: 0 leaves out Blood Fury's 10%
// weapon damage, keeping the attack and spell power the tooltip gives. The logs show the
// weapon part. The user takes it for a game bug.
func AssumeBloodFuryWeapon() bool {
	return Assume("orc_bf_weapon", 1) != 0
}

// tickCritMultiplier is a periodic tick's crit multiplier. The report assumption
// bleed_tick_crit scales a physical tick crit's bonus: 0.5 makes a 200% bleed crit deal 150%.
func (dot *Dot) tickCritMultiplier(attackTable *AttackTable) float64 {
	critMultiplier := dot.Spell.CritMultiplier(attackTable)
	if reportSwitches.bleedTickCritSet && dot.Spell.SchoolIndex == stats.SchoolIndexPhysical {
		critMultiplier = 1 + (critMultiplier-1)*reportSwitches.bleedTickCrit
	}
	return critMultiplier
}
