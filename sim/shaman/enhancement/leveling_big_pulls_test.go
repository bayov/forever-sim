package enhancement

import (
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

var (
	movingLine  = regexp.MustCompile(`Leveling: we're at \((-?[0-9.]+), (-?[0-9.]+)\), moving \((-?[0-9.]+), (-?[0-9.]+)\)`)
	mobAtLine   = regexp.MustCompile(`Leveling: (Target \d+) at \((-?[0-9.]+), (-?[0-9.]+)\), moving \((-?[0-9.]+), (-?[0-9.]+)\)`)
	reachedLine = regexp.MustCompile(`Leveling: (Target \d+) reached us`)
	noticedLine = regexp.MustCompile(`Leveling: (Target \d+) noticed us`)
	// A swing we start, which lands a moment later.
	ourSwingStartLine = regexp.MustCompile(`\] \[ \(#1\)\] Casting \{OtherID: 7, Tag: 1\}`)
	dazedLine         = regexp.MustCompile(`Aura (gained|faded): \{SpellID: 1604\}`)
)

func parse(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// bigPull is a session of a level 30 against mobs far under her, which she lives through, pulling
// this many at a time.
func bigPull(t *testing.T, opts *proto.LevelingOptions, pack int, iterations int32) *proto.RaidSimRequest {
	spread := int32(0)
	opts.MobLevelSpread = &spread
	if opts.MobLevelOffset == 0 {
		opts.MobLevelOffset = -12
	}
	opts.MaxPackSize = 12
	req := levelingRequest(t, opts, iterations)
	req.Raid.Parties[0].Players[0].BonusStats = &proto.UnitStats{Stats: stats.Stats{stats.Armor: 4000, stats.Stamina: 300}.ToFloatArray()}
	return withRotation(t, req, rotationOf(append([]string{healLine}, packLines(pack)...)...))
}

// TestLevelingTenPulls: a rotation that pulls ten at a time gets ten, with room for twelve.
func TestLevelingTenPulls(t *testing.T) {
	m := runLeveling(t, bigPull(t, &proto.LevelingOptions{}, 10, 20)).EncounterMetrics.Leveling
	t.Logf("%.1f mobs, %.2f a pull, %.2f deaths", m.MobsKilled, m.MobsPerPull, m.Deaths)
	if m.MobsPerPull < 9 || m.MobsPerPull > 10 {
		t.Errorf("%.2f mobs a pull pulling ten at a time", m.MobsPerPull)
	}
}

// TestLevelingFacing: running ahead of the mobs on us we don't swing at them, and their hits daze us
// now and then. A daze halves the speed of the run we're on for four seconds. Strafing, or with
// facing ignored, we swing as we run.
func TestLevelingFacing(t *testing.T) {
	ahead, none := 1.0, 0.0
	for _, opts := range []*proto.LevelingOptions{{RunForward: &ahead}, {IgnoreFacing: true}, {RunForward: &none, Surround: &none}} {
		ignore := opts.RunForward != &ahead
		req := bigPull(t, opts, 6, 1)
		req.SimOptions.Debug = true
		speed, onUs := 0.0, map[string]bool{}
		swings, mobSwings, avoided, dazes, slowRuns := 0, 0, 0, 0, 0
		dazed, dazedAt := false, 0.0
		logLines(t, runLeveling(t, req).Logs, func(at float64, line string) {
			if m := movingLine.FindStringSubmatch(line); m != nil {
				speed = math.Hypot(parse(m[3]), parse(m[4]))
				if dazed && speed > 0 {
					slowRuns++
				}
				if want := map[bool]float64{false: 7, true: 3.5}[dazed]; speed > 0 && math.Abs(speed-want) > 0.02 {
					t.Errorf("at %.2f we run at %.2f yards a sec, want %.2f (dazed: %v)", at, speed, want, dazed)
				}
			}
			if m := reachedLine.FindStringSubmatch(line); m != nil {
				onUs[m[1]] = true
			}
			if m := diedLine.FindStringSubmatch(line); m != nil {
				delete(onUs, m[1])
			}
			if strings.Contains(line, "Leveling: died") {
				onUs = map[string]bool{}
			}
			if m := dazedLine.FindStringSubmatch(line); m != nil {
				dazed = m[1] == "gained"
				if dazed {
					dazes, dazedAt = dazes+1, at
				} else if at < dazedAt+3.99 {
					t.Errorf("at %.2f the daze of %.2f ended early", at, dazedAt)
				}
			}
			if speed > 0 && ourSwingStartLine.MatchString(line) {
				swings++
			}
			if m := mobSwingLine.FindStringSubmatch(line); m != nil && speed > 0 {
				mobSwings++
				if m[2] == "Dodge" || m[2] == "Parry" || m[2] == "Block" {
					avoided++
				}
			}
		})
		t.Logf("always in front %v: %d of our swings while running, %d of the mobs', %d of them avoided, %d dazes, %d runs dazed", ignore, swings, mobSwings, avoided, dazes, slowRuns)
		if ignore && (swings == 0 || dazes > 0 || avoided == 0) {
			t.Errorf("always in front: %d swings while running, %d dazes and %d avoided, want some, none and some", swings, dazes, avoided)
		}
		if !ignore && (swings > 0 || dazes == 0 || slowRuns == 0 || mobSwings == 0) {
			t.Errorf("%d swings while running, %d dazes, %d runs dazed and %d mob swings meanwhile, want none, then some of each", swings, dazes, slowRuns, mobSwings)
		}
	}
}

// TestLevelingSwingsAtMobsOnUs: while mobs stand on us we swing at those, so when our target dies
// the next swing isn't held for one that's still coming.
func TestLevelingSwingsAtMobsOnUs(t *testing.T) {
	req := bigPull(t, &proto.LevelingOptions{}, 8, 1)
	req.SimOptions.Debug = true
	onUs := map[string]bool{}
	swings := 0
	logLines(t, runLeveling(t, req).Logs, func(at float64, line string) {
		if m := reachedLine.FindStringSubmatch(line); m != nil {
			onUs[m[1]] = true
		}
		if m := diedLine.FindStringSubmatch(line); m != nil {
			delete(onUs, m[1])
		}
		if strings.Contains(line, "Leveling: died") {
			onUs = map[string]bool{}
		}
		if m := ourSwingLine.FindStringSubmatch(line); m != nil {
			swings++
			if len(onUs) > 0 && !onUs[m[1]] {
				t.Errorf("at %.2f swung at %s, which isn't on us", at, m[1])
			}
		}
	})
	if swings < 50 {
		t.Errorf("%d swings", swings)
	}
}

// TestLevelingCamp: with a camp of eight and nothing around it, the mobs stand at their places, the
// next one we find is the nearest, and none is there again before it respawned. So pulling more
// at once can't kill more than the spot has.
func TestLevelingCamp(t *testing.T) {
	none := 0.0
	const respawn = 240
	opts := &proto.LevelingOptions{CampMobs: 8, CampOutsideMobs: &none, Respawn: respawn, SessionLength: 1200}
	req := bigPull(t, opts, 10, 1)
	req.SimOptions.Debug = true
	type place struct{ x, y float64 }
	at, diedAt := map[string]place{}, map[place]float64{}
	kills := 0
	result := runLeveling(t, req)
	logLines(t, result.Logs, func(now float64, line string) {
		if m := mobAtLine.FindStringSubmatch(line); m != nil && parse(m[4]) == 0 && parse(m[5]) == 0 {
			p := place{parse(m[2]), parse(m[3])}
			at[m[1]] = p
			if math.Hypot(p.x, p.y) > 100 {
				t.Errorf("a mob at (%.1f, %.1f), want all within 100 yards", p.x, p.y)
			}
			if died, ok := diedAt[p]; ok && now < died+respawn-0.01 {
				t.Errorf("at %.2f a mob stands at (%.1f, %.1f) again, where one died at %.2f", now, p.x, p.y, died)
			}
		}
		if m := diedLine.FindStringSubmatch(line); m != nil {
			diedAt[at[m[1]]] = now
			kills++
		}
	})
	t.Logf("%d kills at %d places", kills, len(diedAt))
	if len(diedAt) != 8 || kills < 16 || kills > 8*(1200/respawn) {
		t.Errorf("%d kills at %d places, want 16 to %d at 8", kills, len(diedAt), 8*(1200/respawn))
	}
}

// TestLevelingCampNotices: in a dense camp the mobs we pass notice us without us looking for them,
// and one that's pulled brings the ones within 10 yards along at the same moment.
func TestLevelingCampNotices(t *testing.T) {
	req := bigPull(t, &proto.LevelingOptions{CampMobs: 60, MobLevelOffset: -20}, 1, 1)
	req.SimOptions.Debug = true
	type place struct{ x, y float64 }
	at := map[string]place{}
	noticed, linked := 0, 0
	var pulledAt []float64
	var pulledFrom []place
	logLines(t, runLeveling(t, req).Logs, func(now float64, line string) {
		if m := mobAtLine.FindStringSubmatch(line); m != nil && parse(m[4]) == 0 && parse(m[5]) == 0 {
			at[m[1]] = place{parse(m[2]), parse(m[3])}
		}
		if noticedLine.MatchString(line) {
			noticed++
		}
		if i := strings.Index(line, "Leveling: pulled Target "); i >= 0 {
			name := strings.Join(strings.Fields(line[i:])[2:4], " ")
			p := at[name]
			if slices.ContainsFunc(pulledFrom, func(q place) bool { return math.Hypot(p.x-q.x, p.y-q.y) <= 10 }) && pulledAt[len(pulledAt)-1] == now {
				linked++
			}
			if len(pulledAt) > 0 && pulledAt[len(pulledAt)-1] != now {
				pulledFrom = nil
			}
			pulledAt, pulledFrom = append(pulledAt, now), append(pulledFrom, p)
		}
	})
	t.Logf("%d pulled, %d noticed us, %d came along", len(pulledAt), noticed, linked)
	if noticed < 5 || linked < 5 {
		t.Errorf("%d mobs noticed us and %d came along with another, want some of each", noticed, linked)
	}
}

// TestLevelingCampDeath: dying in a camp costs the corpse run. We come back next to it, and the
// mobs stand at their places again with full health.
func TestLevelingCampDeath(t *testing.T) {
	req := bigPull(t, &proto.LevelingOptions{CampMobs: 30, MobLevelOffset: 3}, 8, 1)
	req.SimOptions.Debug = true
	died, lastX, lastY := 0, 0.0, 0.0
	result := runLeveling(t, req)
	logLines(t, result.Logs, func(now float64, line string) {
		if strings.Contains(line, "Leveling: died") {
			died++
		}
		if m := movingLine.FindStringSubmatch(line); m != nil {
			x, y := parse(m[1]), parse(m[2])
			if math.Hypot(x-lastX, y-lastY) > 100 && died > 0 {
				t.Errorf("at %.2f we jumped from (%.1f, %.1f) to (%.1f, %.1f)", now, lastX, lastY, x, y)
			}
			lastX, lastY = x, y
		}
	})
	if died < 2 || result.EncounterMetrics.Leveling.SecondsDead < 300 {
		t.Errorf("pulling eight mobs three levels over us died %d times and lay dead for %.0f sec", died, result.EncounterMetrics.Leveling.SecondsDead)
	}
}
