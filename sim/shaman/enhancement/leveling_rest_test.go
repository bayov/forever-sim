package enhancement

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

var (
	sitLine   = regexp.MustCompile(`Leveling: sitting down \(eat: (true|false), drink: (true|false)\) at (\d+)% health and (\d+)% mana`)
	standLine = regexp.MustCompile(`Leveling: standing up at (\d+)% health and (\d+)% mana`)
	regenLine = regexp.MustCompile(`\] \[ \(#1\)\] Gained ([0-9.]+) health from \{SpellID: 20577, Tag: 1\}`)
)

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// TestLevelingRotationRests: the rotation sits down to eat and drink below 60% health or 40%
// mana and stands up at 90% of both, and the fixed policy stays out of it.
func TestLevelingRotationRests(t *testing.T) {
	req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{}, 1), rotationOf(healLine,
		`{"action":{"condition":{"or":{"vals":[{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"60%"}}}},{"cmp":{"op":"OpLt","lhs":{"currentManaPercent":{}},"rhs":{"const":{"val":"40%"}}}}]}},"levelingRest":{"eat":true,"drink":true,"healthUntil":{"const":{"val":"90%"}},"manaUntil":{"const":{"val":"90%"}}}}}`))
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	sits := 0
	logLines(t, result.Logs, func(at float64, line string) {
		if m := sitLine.FindStringSubmatch(line); m != nil {
			sits++
			if m[1] != "true" || m[2] != "true" || (atoi(m[3]) >= 60 && atoi(m[4]) >= 40) {
				t.Errorf("at %.2f %s, want eating and drinking below 60%% health or 40%% mana", at, m[0])
			}
		} else if m := standLine.FindStringSubmatch(line); m != nil {
			if atoi(m[1]) < 90 || atoi(m[2]) < 90 || atoi(m[1]) > 99 && atoi(m[2]) > 99 {
				t.Errorf("at %.2f %s, want 90%% of both, not full", at, m[0])
			}
		}
	})
	m := result.EncounterMetrics.Leveling
	t.Logf("%d rests, %.0f food, %.0f drinks, %.0f bandages, %.0f s resting", sits, m.Food, m.Drinks, m.Bandages, m.SecondsResting)
	if sits < 3 || m.Food < 1 || m.Drinks < 1 {
		t.Errorf("%d rests with %.0f food and %.0f drinks, want a few", sits, m.Food, m.Drinks)
	}
	if m.Bandages != 0 {
		t.Errorf("%.0f bandages: the fixed policy bandaged though the rotation rests", m.Bandages)
	}
}

// TestLevelingSitting: sitting without food or water to rest to full health regens a third
// more a second from Spirit than standing (as we do while we search).
func TestLevelingSitting(t *testing.T) {
	req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{}, 1), rotationOf(
		`{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"50%"}}}},"levelingRest":{}}}`))
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	sitting := false
	var sat, stood []float64
	logLines(t, result.Logs, func(at float64, line string) {
		switch {
		case sitLine.MatchString(line):
			sitting = true
		case standLine.MatchString(line):
			sitting = false
		case strings.Contains(line, "Leveling: pulled"):
			sitting = false
		}
		if m := regenLine.FindStringSubmatch(line); m != nil {
			gain, _ := strconv.ParseFloat(m[1], 64)
			if sitting {
				sat = append(sat, gain)
			} else {
				stood = append(stood, gain)
			}
		}
	})
	if len(sat) == 0 || len(stood) == 0 {
		t.Fatalf("%d ticks sitting and %d standing", len(sat), len(stood))
	}
	// The last tick of a rest can be cut short by full health, so take the most.
	maxOf := func(xs []float64) float64 {
		m := 0.0
		for _, x := range xs {
			m = math.Max(m, x)
		}
		return m
	}
	if ratio := maxOf(sat) / maxOf(stood); math.Abs(ratio-4.0/3) > 2e-3 {
		t.Errorf("a second's regen sitting is %.4f times standing, want 4/3", ratio)
	}
}

// TestLevelingRotationBandages: the rotation bandages below 70% health, never twice within
// the minute of Recently Bandaged.
func TestLevelingRotationBandages(t *testing.T) {
	req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{}, 1), rotationOf(healLine,
		`{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"70%"}}}},"levelingBandage":{}}}`))
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	last := -100.0
	n := 0
	logLines(t, result.Logs, func(at float64, line string) {
		if strings.Contains(line, "Leveling: bandaging") {
			n++
			if at-last < 60-0.01 {
				t.Errorf("at %.2f bandaged %.2f sec after the last one, want a minute", at, at-last)
			}
			last = at
		}
	})
	if n < 3 || float64(n) != result.EncounterMetrics.Leveling.Bandages {
		t.Errorf("%d bandages in the log, %.0f in the metrics, want a few and the same", n, result.EncounterMetrics.Leveling.Bandages)
	}
}

// TestLevelingHealBetweenPulls: a rotation that heals out of combat finishes its cast before
// we search, since running would stop it.
func TestLevelingHealBetweenPulls(t *testing.T) {
	req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{}, 1), rotationOf(
		`{"action":{"condition":{"and":{"vals":[{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"80%"}}}},{"cmp":{"op":"OpEq","lhs":{"levelingMobsOnUs":{}},"rhs":{"const":{"val":"0"}}}},{"cmp":{"op":"OpEq","lhs":{"levelingMobsRunningIn":{}},"rhs":{"const":{"val":"0"}}}}]}},"castSpell":{"spellId":{"spellId":8008}}}}`,
		`{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"30%"}}}},"levelingRest":{"eat":true,"drink":true}}}`))
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	heals, interrupted := 0, 0
	logLines(t, result.Logs, func(at float64, line string) {
		switch {
		case strings.Contains(line, "Completed cast {SpellID: 8008}"):
			heals++
		case strings.Contains(line, "Cast of {SpellID: 8008} interrupted"):
			interrupted++
		}
	})
	if heals < 5 || interrupted > 0 {
		t.Errorf("%d heals out of combat, %d of them cut short, want a few and none", heals, interrupted)
	}
}

// TestLevelingTrollRegen: a Troll keeps 10% of its health regen in combat (Regeneration).
func TestLevelingTrollRegen(t *testing.T) {
	req := levelingRequest(t, &proto.LevelingOptions{}, 1)
	req.Raid.Parties[0].Players[0].Race = proto.Race_RaceTroll
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	inCombat := false
	var combat, outside []float64
	logLines(t, result.Logs, func(at float64, line string) {
		switch {
		case strings.Contains(line, "Leveling: pulled"):
			inCombat = true
		case strings.Contains(line, "Leveling: searching"), strings.Contains(line, "Leveling: bandaging"), strings.Contains(line, "Leveling: sitting"):
			inCombat = false
		}
		if m := regenLine.FindStringSubmatch(line); m != nil {
			gain, _ := strconv.ParseFloat(m[1], 64)
			if inCombat {
				combat = append(combat, gain)
			} else {
				outside = append(outside, gain)
			}
		}
	})
	if len(combat) == 0 || len(outside) == 0 {
		t.Fatalf("%d regen ticks in combat and %d out", len(combat), len(outside))
	}
	if ratio := combat[0] / outside[0]; math.Abs(ratio-0.1) > 2e-3 {
		t.Errorf("in combat a Troll regens %.4f of its out of combat regen, want 0.1", ratio)
	}
}

// TestLevelingRestWithMobWaiting: we look for the next mob while fighting, and rest with it
// waiting once the fight is over. Nothing pulls it while we sit: not a swing of ours from
// where it stands (our target moves to it when the last mob dies), and not the rotation,
// since sitting stops everything else.
func TestLevelingRestWithMobWaiting(t *testing.T) {
	req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{MobLevelOffset: -3}, 1), rotationOf(
		`{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"90%"}}}},"levelingRest":{"eat":true,"drink":true}}}`,
		`{"action":{"condition":{"cmp":{"op":"OpEq","lhs":{"math":{"op":"OpAdd","lhs":{"levelingMobsOnUs":{}},"rhs":{"levelingMobsRunningIn":{}}}},"rhs":{"const":{"val":"0"}}}},"levelingWalkToPullTarget":{}}}`,
		`{"action":{"levelingLookForMob":{}}}`))
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	sitting, rests, waiting := false, 0, 0
	found := ""
	logLines(t, result.Logs, func(at float64, line string) {
		switch {
		case foundLine.MatchString(line):
			found = foundLine.FindStringSubmatch(line)[1]
		case sitLine.MatchString(line):
			sitting = true
			rests++
			if found != "" {
				waiting++
			}
		case standLine.MatchString(line):
			sitting = false
		case pulledLine.MatchString(line):
			if sitting {
				t.Errorf("at %.2f a mob was pulled while we sat: %s", at, line)
			}
			// Level 26 to 28 mobs notice a level 30 from 16 to 18 yards, and nothing else
			// pulls them here.
			if from := atoi(pulledLine.FindStringSubmatch(line)[3]); from > 18 {
				t.Errorf("at %.2f a mob was pulled from %d yards, beyond its aggro radius: %s", at, from, line)
			}
			found = ""
		}
	})
	if waiting < 3 {
		t.Errorf("%d of %d rests had a mob found and waiting, want a few", waiting, rests)
	}
}
