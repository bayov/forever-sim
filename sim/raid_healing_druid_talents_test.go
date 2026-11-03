package sim

// More of the Restoration druid's talents (raid_healing_druid_test.go has the helpers): the mana,
// pushback and threat ones, and a HoT check at Most Hurt.

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

// logTime is a log line's time.
func logTime(line string) float64 {
	v, _ := strconv.ParseFloat(line[1:max(1, strings.Index(line, "]"))], 64)
	return v
}

// regenGained is the healer's mana from regeneration in the log before until.
func (l druidLog) regenGained(until float64) (gained float64) {
	re := regexp.MustCompile(`1 \(#2\)\] Gained ([\d.]+) mana from \{OtherID: 2[,}]`)
	for _, line := range l.lines {
		if m := re.FindStringSubmatch(line); m != nil && logTime(line) < until {
			v, _ := strconv.ParseFloat(m[1], 64)
			gained += v
		}
	}
	return gained
}

// Reflection 3/3: half our Spirit regen goes on while we cast, so a druid casting from the pull gets back
// much more mana.
func TestDruidReflection(t *testing.T) {
	rot := castAt(htTop, "Tank", "")
	without := runDruidLog(t, druidRaid(t, nil, 0, rot)).regenGained(40)
	with := runDruidLog(t, druidRaid(t, map[string]int{"reflection": 3}, 0, rot)).regenGained(40)
	if with < 1.3*without {
		t.Errorf("%.0f mana regenerated with Reflection 3/3, %.0f without", with, without)
	}
}

// Nature's Focus 5/5: a 70% chance a hit taken doesn't push a Healing Touch back.
func TestDruidNaturesFocus(t *testing.T) {
	pushbacks := func(points int) (n int) {
		hits := &proto.RaidDamageEvent{School: proto.SpellSchool_SpellSchoolFire, MinDamage: 50, MaxDamage: 50, Who: proto.RaidDamageEvent_NotTanks, EverySeconds: 1, FirstAt: 1}
		rsr := druidRaid(t, map[string]int{"naturesFocus": points}, 0, castAt(htTop, "Tank", ""))
		rsr.Encounter.RaidDamage = []*proto.RaidDamageEvent{hits}
		for _, line := range runDruidLog(t, rsr).lines {
			if strings.Contains(line, "(#2)] Cast {SpellID: 25297} pushed back") {
				n++
			}
		}
		return n
	}
	without, with := pushbacks(0), pushbacks(5)
	if without < 20 || float64(with) > 0.45*float64(without) || float64(with) < 0.15*float64(without) {
		t.Errorf("%d pushbacks with Nature's Focus 5/5, %d without; want about 30%%", with, without)
	}
}

// Subtlety 3/3: 30% less threat from our heals.
func TestDruidSubtlety(t *testing.T) {
	threat := func(points int) (threat, healing float64) {
		rsr := druidRaid(t, map[string]int{"subtlety": points}, 0, castAt(htTop, "Tank", ""))
		rsr.SimOptions.Iterations = 10
		for _, a := range playerMetrics(runRaid(t, rsr), 1).Actions {
			if a.Id.GetSpellId() == htTop {
				for _, tgt := range a.Targets {
					threat += tgt.Threat
					healing += tgt.Healing - tgt.Overhealing
				}
			}
		}
		return
	}
	t0, h0 := threat(0)
	t3, h3 := threat(3)
	if r := (t3 / h3) / (t0 / h0); !near(r, 0.7, 0.01) {
		t.Errorf("threat per effective healing %.3f times without Subtlety, want 0.7", r)
	}
}

// A HoT check at Most Hurt keeps its condition though nobody is hurt when the rotation is built: the
// druid puts Rejuvenation on the most hurt only while it has none, so it isn't recast every global
// cooldown.
func TestDruidHotAtMostHurt(t *testing.T) {
	l := runDruidLog(t, druidRaid(t, nil, 0, castAt(rjTop, "MostHurt", hotOff(rjTop, "MostHurt"))))
	casts, ticks := len(l.casts[rjTop]), len(l.ticks(rjTop))
	if casts == 0 || ticks < 3*casts {
		t.Errorf("%d Rejuvenations at the most hurt ticked %d times: recast over its own HoT", casts, ticks)
	}
}
