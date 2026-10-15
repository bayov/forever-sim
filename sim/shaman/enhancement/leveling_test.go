package enhancement

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// levelingRequest is a naked level 30 orc shaman with a Whirlwind Axe and white swings, who
// casts Lesser Healing Wave below 40% health, grinding a session with the given options.
func levelingRequest(t *testing.T, opts *proto.LevelingOptions, iterations int32) *proto.RaidSimRequest {
	req := normalMobRequest(nil, true, iterations)
	req.Encounter = &proto.Encounter{Leveling: opts}
	apl := &proto.APLRotation{}
	if err := protojson.Unmarshal([]byte(`{"type":"TypeAPL","priorityList":[{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"40%"}}}},"castSpell":{"spellId":{"spellId":8008}}}}]}`), apl); err != nil {
		t.Fatal(err)
	}
	req.Raid.Parties[0].Players[0].Rotation = apl
	return req
}

// aggroRadius is the aggro radius of a mob of this level at our level 30.
func aggroRadius(level string) int {
	l, _ := strconv.Atoi(level)
	return 20 + l - 30
}

// checkRunIn: a mob runs to melee range at 7 to 8.5 yards a sec.
func checkRunIn(t *testing.T, at float64, mob string, from int, runIn float64) {
	t.Helper()
	if lo, hi := float64(from-5)/8.5, float64(from-5)/6.5; runIn < lo-0.01 || runIn > hi+0.01 {
		t.Errorf("at %.2f %s runs in %.2f sec from %d yards, want %.2f to %.2f", at, mob, runIn, from, lo, hi)
	}
}

func parseRunIn(value, unit string) float64 {
	runIn, _ := strconv.ParseFloat(value, 64)
	switch unit {
	case "ms":
		runIn /= 1e3
	case "µs":
		runIn /= 1e6
	case "ns":
		runIn /= 1e9
	}
	return runIn
}

func runLeveling(t *testing.T, req *proto.RaidSimRequest) *proto.RaidSimResult {
	result := core.RunRaidSim(req)
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	if result.EncounterMetrics.Leveling == nil {
		t.Fatal("no leveling metrics")
	}
	return result
}

var (
	logTime      = regexp.MustCompile(`^\[(-?[0-9.]+)\] `)
	foundLine    = regexp.MustCompile(`Leveling: found (Target \d+) \(level \d+\) (\d+) yards away`)
	pulledLine   = regexp.MustCompile(`Leveling: pulled (Target \d+) \(level (\d+)\) from (\d+) yards, here in ([0-9.]+)(ms|s|µs|ns)`)
	mobSwingLine = regexp.MustCompile(`\] \[(Target \d+)\] \[ \(#1\)\] \{OtherID: 7, Tag: 1\} (Hit|Crit|Miss|Dodge|Parry|Block|Crush)`)
	ourSwingLine = regexp.MustCompile(`\] \[ \(#1\)\] \[(Target \d+)\] \{OtherID: 7, Tag: 1\} (Hit|Crit|Miss|Dodge|Parry|Block|Glance)`)
	diedLine     = regexp.MustCompile(`\[(Target \d+)\] Died`)
)

// TestLevelingSession runs one session with its log and checks the loop: we find a mob 10 to
// 40 yards away, and with a rotation that doesn't pull, walk up to it until it notices us
// (its aggro radius, 20 yards at our level, a yard more for each level over us), or it
// notices us at once when we found it that close. A pulled mob runs to us and only then swings, and we only reach it in melee then. A
// dead mob swings no more until it's pulled again. The time in each state adds up to the
// session, and we kill a steady stream of mobs.
func TestLevelingSession(t *testing.T) {
	req := levelingRequest(t, &proto.LevelingOptions{}, 1)
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	arrival := map[string]float64{}
	dead := map[string]bool{}
	found := map[string]int{}
	pulls := 0
	for _, line := range strings.Split(result.Logs, "\n") {
		m := logTime.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		at, _ := strconv.ParseFloat(m[1], 64)
		if f := foundLine.FindStringSubmatch(line); f != nil {
			d, _ := strconv.Atoi(f[2])
			if d < 10 || d > 40 {
				t.Errorf("at %.2f found %s %d yards away, want 10 to 40", at, f[1], d)
			}
			found[f[1]] = d
			continue
		}
		if p := pulledLine.FindStringSubmatch(line); p != nil {
			runIn := parseRunIn(p[4], p[5])
			from, _ := strconv.Atoi(p[3])
			if want := min(found[p[1]], aggroRadius(p[2])); from != want {
				t.Errorf("at %.2f %s found %d yards away was pulled from %d, want %d: where it noticed us as we walked up", at, p[1], found[p[1]], from, want)
			}
			checkRunIn(t, at, p[1], from, runIn)
			arrival[p[1]] = at + runIn
			dead[p[1]] = false
			pulls++
			continue
		}
		if d := diedLine.FindStringSubmatch(line); d != nil {
			dead[d[1]] = true
			continue
		}
		for _, s := range []struct {
			re   *regexp.Regexp
			what string
		}{{mobSwingLine, "swung at us"}, {ourSwingLine, "took our swing"}} {
			if w := s.re.FindStringSubmatch(line); w != nil {
				mob := w[1]
				if dead[mob] {
					t.Errorf("at %.2f dead %s %s", at, mob, s.what)
				} else if at < arrival[mob]-1e-6 {
					t.Errorf("at %.2f %s %s before it arrived at %.2f", at, mob, s.what, arrival[mob])
				}
			}
		}
	}

	m := result.EncounterMetrics.Leveling
	if pulls < 15 || m.MobsKilled < 15 {
		t.Errorf("%d pulls and %.1f mobs killed in 20 min, want a steady stream", pulls, m.MobsKilled)
	}
	if total := m.SecondsInCombat + m.SecondsResting + m.SecondsSearching + m.SecondsDead; math.Abs(total-1200) > 0.01 {
		t.Errorf("the states add up to %.2f sec, want the 1200 sec session", total)
	}
	if m.MobsPerPull != 1 {
		t.Errorf("%.2f mobs a pull, want 1", m.MobsPerPull)
	}
	if want := m.MobsKilled; m.Xp < 190*(want-1) || m.Xp > 205*want {
		t.Errorf("%.0f XP for %.1f level 29 to 31 mobs at level 30, want about 195 each", m.Xp, m.MobsKilled)
	}
}

// packLines pull n mobs at a time: they look for another mob while the pull has taken fewer
// than n (Mobs This Pull counts the dead ones too), and walk up to it while fighting.
func packLines(n int) []string {
	return []string{
		`{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"levelingMobsThisPull":{}},"rhs":{"const":{"val":"` + strconv.Itoa(n) + `"}}}},"levelingLookForMob":{}}}`,
		`{"action":{"condition":{"cmp":{"op":"OpGe","lhs":{"levelingMobsThisPull":{}},"rhs":{"const":{"val":"1"}}}},"levelingWalkToPullTarget":{}}}`,
	}
}

// TestLevelingPacks: a rotation that pulls three mobs at a time takes three a pull, not
// more as the first ones die, and three are too much for a naked level 30, who dies and
// loses the corpse runs' time.
func TestLevelingPacks(t *testing.T) {
	one := runLeveling(t, levelingRequest(t, &proto.LevelingOptions{}, 50)).EncounterMetrics.Leveling
	three := runLeveling(t, withRotation(t, levelingRequest(t, &proto.LevelingOptions{}, 50), rotationOf(append([]string{healLine}, packLines(3)...)...))).EncounterMetrics.Leveling
	t.Logf("single pulls: %.1f mobs, %.2f deaths; three at a time: %.1f mobs, %.2f a pull, %.2f deaths", one.MobsKilled, one.Deaths, three.MobsKilled, three.MobsPerPull, three.Deaths)
	if three.MobsPerPull < 2.7 || three.MobsPerPull > 3 {
		t.Errorf("%.2f mobs a pull pulling three at a time, want close to 3 and never more", three.MobsPerPull)
	}
	if three.Deaths < 2 || three.SecondsDead < 300 {
		t.Errorf("pulling three died %.1f times for %.0f sec, want a few corpse runs", three.Deaths, three.SecondsDead)
	}
	if one.Deaths > 0.1 {
		t.Errorf("single pulls died %.2f times a session", one.Deaths)
	}
	if three.MobsKilled >= one.MobsKilled {
		t.Errorf("pulling three killed %.1f, single pulls %.1f: dying should cost more than it gains", three.MobsKilled, one.MobsKilled)
	}
}

// TestLevelingRests: a rotation that rests below 99% sits after every pull, and one with its
// only Rest line switched off never sits. One with no Rest or Bandage line gets the default
// ones: it bandages below 50% health, and eats and drinks below 50% health or 40% mana.
// Resting takes time but keeps us alive.
func TestLevelingRests(t *testing.T) {
	restAlways := `{"action":{"condition":{"or":{"vals":[{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"99%"}}}},{"cmp":{"op":"OpLt","lhs":{"currentManaPercent":{}},"rhs":{"const":{"val":"99%"}}}}]}},"levelingRest":{"eat":true,"drink":true}}}`
	restOff := `{"hide":true,"action":{"levelingRest":{"eat":true,"drink":true}}}`
	always := runLeveling(t, withRotation(t, levelingRequest(t, &proto.LevelingOptions{}, 20), rotationOf(healLine, restAlways))).EncounterMetrics.Leveling
	never := runLeveling(t, withRotation(t, levelingRequest(t, &proto.LevelingOptions{}, 20), rotationOf(healLine, restOff))).EncounterMetrics.Leveling
	usual := runLeveling(t, levelingRequest(t, &proto.LevelingOptions{}, 20)).EncounterMetrics.Leveling
	t.Logf("resting %.0f / %.0f / %.0f sec, deaths %.2f / %.2f / %.2f (always, default lines, never)", always.SecondsResting, usual.SecondsResting, never.SecondsResting, always.Deaths, usual.Deaths, never.Deaths)
	if always.SecondsResting <= usual.SecondsResting {
		t.Errorf("resting below 99%% sat for %.0f sec a session, the default lines %.0f, want more", always.SecondsResting, usual.SecondsResting)
	}
	if usual.Food < 1 || usual.Drinks < 1 || usual.Bandages < 1 {
		t.Errorf("the default lines ate %.1f, drank %.1f and bandaged %.1f times a session, want some of each", usual.Food, usual.Drinks, usual.Bandages)
	}
	if never.SecondsResting != 0 || never.Food != 0 || never.Bandages != 0 {
		t.Errorf("with the Rest line switched off we sat for %.0f sec, ate %.1f and bandaged %.1f times, want none", never.SecondsResting, never.Food, never.Bandages)
	}
	if never.Deaths <= always.Deaths {
		t.Errorf("never resting died %.2f times, always resting %.2f", never.Deaths, always.Deaths)
	}
}
