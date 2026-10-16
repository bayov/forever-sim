package enhancement

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// pullRotation heals below 40% health, keeps Strength of Earth Totem down (cast "at" the
// mob we found, which mustn't pull it), and pulls the mob we found with rank 1 Earth Shock
// within its 20 yards, or with rank 1 Lightning Bolt beyond that while no mob is on us.
// Further out than the bolt's 30 yards it does nothing, so we walk up to the mob and shock
// it once it's in range.
const pullRotation = `{"type":"TypeAPL","priorityList":[
	{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"40%"}}}},"castSpell":{"spellId":{"spellId":8008}}}},
	{"action":{"condition":{"not":{"val":{"totemRemainingTime":{"totemType":"Earth"}}}},"castSpell":{"spellId":{"spellId":8075}}}},
	{"action":{"condition":{"cmp":{"op":"OpLe","lhs":{"levelingDistance":{"targetUnit":{"type":"PullTarget"}}},"rhs":{"const":{"val":"20"}}}},"castSpell":{"spellId":{"spellId":8042},"target":{"type":"PullTarget"}}}},
	{"action":{"condition":{"cmp":{"op":"OpEq","lhs":{"levelingMobsOnUs":{}},"rhs":{"const":{"val":"0"}}}},"castSpell":{"spellId":{"spellId":403},"target":{"type":"PullTarget"}}}}
]}`

func withRotation(t *testing.T, req *proto.RaidSimRequest, rotation string) *proto.RaidSimRequest {
	apl := &proto.APLRotation{}
	if err := protojson.Unmarshal([]byte(rotation), apl); err != nil {
		t.Fatal(err)
	}
	req.Raid.Parties[0].Players[0].Rotation = apl
	return req
}

var (
	searchingLine = regexp.MustCompile(`Leveling: searching for a mob`)
	castLine      = regexp.MustCompile(`\] \[ \(#1\)\] Casting \{SpellID: (\d+)\}`)
	completedLine = regexp.MustCompile(`\] \[ \(#1\)\] Completed cast \{SpellID: (\d+)\}`)
)

// TestLevelingPulls pulls with the rotation's spells: a mob found within its aggro radius
// notices us at once, one 20 to 30 yards away takes a bolt, and we walk up to one further
// away until the shock reaches it or it notices us. The bolt's mob only notices us once the bolt hits
// it, and then runs to us from where it stood. While we search we run, so the rotation
// starts no hard cast.
func TestLevelingPulls(t *testing.T) {
	req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{}, 1), pullRotation)
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	var found map[string]int
	searching := false
	kinds := map[string]int{}
	var lastBolt float64 = -1
	for _, line := range strings.Split(result.Logs, "\n") {
		m := logTime.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		at, _ := strconv.ParseFloat(m[1], 64)
		switch {
		case searchingLine.MatchString(line):
			searching = true
			found = map[string]int{}
		case foundLine.MatchString(line):
			f := foundLine.FindStringSubmatch(line)
			searching = false
			found[f[1]], _ = strconv.Atoi(f[2])
		case castLine.MatchString(line):
			spell := castLine.FindStringSubmatch(line)[1]
			if searching && (spell == "403" || spell == "8008") {
				t.Errorf("at %.2f started a hard cast of %s while searching", at, spell)
			}
		case completedLine.MatchString(line):
			if completedLine.FindStringSubmatch(line)[1] == "403" {
				lastBolt = at
			}
		case pulledLine.MatchString(line):
			p := pulledLine.FindStringSubmatch(line)
			mob, runIn := p[1], parseRunIn(p[4], p[5])
			from, _ := strconv.Atoi(p[3])
			stood, radius := found[mob], aggroRadius(p[2])
			checkRunIn(t, at, mob, from, runIn)
			switch {
			case stood <= radius:
				kinds["noticed"]++
				if from != stood {
					t.Errorf("at %.2f %s found %d yards away, in its %d yard aggro radius, was pulled from %d, want at once", at, mob, stood, radius, from)
				}
			case stood <= 20:
				kinds["shock"]++
				if from != stood {
					t.Errorf("at %.2f %s found %d yards away was pulled from %d, want a shock from where we stood", at, mob, stood, from)
				}
			case stood <= 30:
				kinds["bolt"]++
				// The bolt flies 20 yards a sec.
				if from != stood || lastBolt < 0 || at-lastBolt < float64(stood)/20-0.05 || at-lastBolt > float64(stood)/20+0.05 {
					t.Errorf("at %.2f %s found %d yards away was pulled from %d, %.2f sec after the last bolt, want a bolt from there", at, mob, stood, from, at-lastBolt)
				}
			default:
				kinds["walk"]++
				if from > max(20, radius) {
					t.Errorf("at %.2f %s found %d yards away was pulled from %d, want walking up to shock range or its %d yard aggro radius", at, mob, stood, from, radius)
				}
			}
		}
	}
	t.Logf("pulls: %v", kinds)
	for _, kind := range []string{"noticed", "bolt", "walk"} {
		if kinds[kind] < 2 {
			t.Errorf("%d %s pulls in the session, want a few of each: %v", kinds[kind], kind, kinds)
		}
	}
}

// TestLevelingTotemRange: a Magma Totem dropped while we search is left behind as we run
// on, so its 8 yard pulses rarely reach the mobs we pull, where one dropped with a mob on us
// hits it every pulse.
func TestLevelingTotemRange(t *testing.T) {
	magma := func(condition string) float64 {
		rotation := `{"type":"TypeAPL","priorityList":[
			{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"40%"}}}},"castSpell":{"spellId":{"spellId":8008}}}},
			{"action":{"condition":{"and":{"vals":[` + condition + `,{"not":{"val":{"totemRemainingTime":{"totemType":"Fire"}}}}]}},"castSpell":{"spellId":{"spellId":8190}}}}
		]}`
		result := runLeveling(t, withRotation(t, levelingRequest(t, &proto.LevelingOptions{}, 20), rotation))
		damage, casts := 0.0, int32(0)
		for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
			for _, target := range action.Targets {
				switch action.Id.GetSpellId() {
				case 8187:
					damage += target.Damage
				case 8190:
					casts += target.Casts
				}
			}
		}
		if casts == 0 {
			t.Fatalf("the rotation never dropped Magma Totem (%s)", condition)
		}
		return damage
	}
	searching := magma(`{"isMoving":{}},{"cmp":{"op":"OpEq","lhs":{"levelingMobsOnUs":{}},"rhs":{"const":{"val":"0"}}}},{"cmp":{"op":"OpEq","lhs":{"levelingDistance":{"targetUnit":{"type":"PullTarget"}}},"rhs":{"const":{"val":"0"}}}}`)
	fighting := magma(`{"cmp":{"op":"OpGe","lhs":{"levelingMobsOnUs":{}},"rhs":{"const":{"val":"1"}}}}`)
	t.Logf("Magma Totem damage a session: %.0f dropped while searching, %.0f with a mob on us", searching, fighting)
	if fighting <= 0 || searching > fighting/10 {
		t.Errorf("Magma Totem dropped while searching dealt %.0f, with a mob on us %.0f, want much less", searching, fighting)
	}
}

// chainRotation is pullRotation, but it bolts a mob 20 to 30 yards away with mobs on us too.
const chainRotation = `{"type":"TypeAPL","priorityList":[
	{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"40%"}}}},"castSpell":{"spellId":{"spellId":8008}}}},
	{"action":{"condition":{"cmp":{"op":"OpLe","lhs":{"levelingDistance":{"targetUnit":{"type":"PullTarget"}}},"rhs":{"const":{"val":"20"}}}},"castSpell":{"spellId":{"spellId":8042},"target":{"type":"PullTarget"}}}},
	{"action":{"castSpell":{"spellId":{"spellId":403},"target":{"type":"PullTarget"}}}}
]}`

var lookingLine = regexp.MustCompile(`Leveling: looking for another mob, running: (true|false)`)

// TestLevelingChainPulls: pulling two at a time (mobs a few levels below us, so the naked
// shaman lives through them), we look for a second mob while fighting the first, running
// with it following us, when no hard cast starts. The rotation bolts the second mob
// from where it stands, mobs on us or not, and pulls stay two mobs.
func TestLevelingChainPulls(t *testing.T) {
	rotation := strings.TrimSuffix(chainRotation, "\n]}") + ",\n" + strings.Join(packLines(2), ",\n") + "\n]}"
	req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{MobLevelOffset: -4}, 1), rotation)
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	running, looks, runs, chainBolts := false, 0, 0, 0
	inCombat := map[string]bool{}
	for _, line := range strings.Split(result.Logs, "\n") {
		m := logTime.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		at, _ := strconv.ParseFloat(m[1], 64)
		switch {
		case lookingLine.MatchString(line):
			looks++
			running = lookingLine.FindStringSubmatch(line)[1] == "true"
			if running {
				runs++
			}
		case foundLine.MatchString(line), searchingLine.MatchString(line):
			running = false
		case castLine.MatchString(line):
			if spell := castLine.FindStringSubmatch(line)[1]; running && (spell == "403" || spell == "8008") {
				t.Errorf("at %.2f started a hard cast of %s while running to look for a mob", at, spell)
			}
		case pulledLine.MatchString(line):
			p := pulledLine.FindStringSubmatch(line)
			if from, _ := strconv.Atoi(p[3]); from > 21 && len(inCombat) > 0 {
				chainBolts++
			}
			inCombat[p[1]] = true
		case diedLine.MatchString(line):
			delete(inCombat, diedLine.FindStringSubmatch(line)[1])
		case strings.Contains(line, "Leveling: died"):
			inCombat = map[string]bool{}
		}
	}
	t.Logf("%d looks for another mob, %d running; %d chain pulls by bolt; %.2f mobs a pull", looks, runs, chainBolts, result.EncounterMetrics.Leveling.MobsPerPull)
	if looks < 5 || runs != looks {
		t.Errorf("%d looks for another mob, %d of them running, want all", looks, runs)
	}
	if chainBolts < 2 {
		t.Errorf("%d chain pulls by bolt from beyond shock range", chainBolts)
	}
	if mpp := result.EncounterMetrics.Leveling.MobsPerPull; mpp < 1.7 || mpp > 2 {
		t.Errorf("%.2f mobs a pull pulling two at a time", mpp)
	}
}

// TestLevelingChainSearchStanding: Look For Mob with Stand never runs, and its searches take
// the standing time (4 to 16 sec by default) where running ones take 2 to 8.
func TestLevelingChainSearchStanding(t *testing.T) {
	for _, c := range []struct {
		stand   bool
		lo, hi  float64
		running string
	}{{false, 2, 8, "true"}, {true, 4, 16, "false"}} {
		lines := packLines(2)
		if c.stand {
			for i := range lines {
				lines[i] = strings.ReplaceAll(lines[i], `"levelingLookForMob":{}`, `"levelingLookForMob":{"stand":true}`)
			}
		}
		rotation := strings.TrimSuffix(chainRotation, "\n]}") + ",\n" + strings.Join(lines, ",\n") + "\n]}"
		req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{MobLevelOffset: -4}, 1), rotation)
		req.SimOptions.Debug = true
		result := runLeveling(t, req)

		lookedAt := -1.0
		var searches []float64
		for _, line := range strings.Split(result.Logs, "\n") {
			m := logTime.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			at, _ := strconv.ParseFloat(m[1], 64)
			switch {
			case lookingLine.MatchString(line):
				if got := lookingLine.FindStringSubmatch(line)[1]; got != c.running {
					t.Errorf("stand %v: at %.2f looking for another mob with running: %s", c.stand, at, got)
				}
				lookedAt = at
			case searchingLine.MatchString(line), strings.Contains(line, "Leveling: died"):
				lookedAt = -1
			case foundLine.MatchString(line) && lookedAt >= 0:
				searches = append(searches, at-lookedAt)
				lookedAt = -1
			}
		}
		if len(searches) < 5 {
			t.Fatalf("stand %v: %d chain searches found a mob", c.stand, len(searches))
		}
		sum := 0.0
		for _, d := range searches {
			sum += d
			if d < c.lo-0.01 || d > c.hi+0.01 {
				t.Errorf("stand %v: a chain search took %.2f sec, want %.0f to %.0f", c.stand, d, c.lo, c.hi)
			}
		}
		t.Logf("stand %v: %d chain searches, %.1f sec on average", c.stand, len(searches), sum/float64(len(searches)))
	}
}

// TestLevelingLookForMob: the rotation looks for another mob only while we're above 70%
// health with fewer than 3 mobs pulled, and shocks it or walks up to it. Pulls take as many
// mobs as that allows, not a fixed number.
func TestLevelingLookForMob(t *testing.T) {
	rotation := `{"type":"TypeAPL","priorityList":[
		{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"40%"}}}},"castSpell":{"spellId":{"spellId":8008}}}},
		{"action":{"condition":{"cmp":{"op":"OpLe","lhs":{"levelingDistance":{"targetUnit":{"type":"PullTarget"}}},"rhs":{"const":{"val":"20"}}}},"castSpell":{"spellId":{"spellId":8042},"target":{"type":"PullTarget"}}}},
		{"action":{"levelingWalkToPullTarget":{}}},
		{"action":{"condition":{"and":{"vals":[{"cmp":{"op":"OpGt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"70%"}}}},{"cmp":{"op":"OpLt","lhs":{"math":{"op":"OpAdd","lhs":{"levelingMobsOnUs":{}},"rhs":{"levelingMobsRunningIn":{}}}},"rhs":{"const":{"val":"3"}}}}]}},"levelingLookForMob":{}}}
	]}`
	req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{MobLevelOffset: -4}, 1), rotation)
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	pullSizes := map[int]int{}
	inPull := 0
	logLines(t, result.Logs, func(at float64, line string) {
		switch {
		case pulledLine.MatchString(line):
			inPull++
		case searchingLine.MatchString(line), strings.Contains(line, "Leveling: died"):
			if inPull > 0 {
				pullSizes[inPull]++
			}
			inPull = 0
		}
	})
	t.Logf("pulls by mobs taken: %v, %.2f mobs a pull", pullSizes, result.EncounterMetrics.Leveling.MobsPerPull)
	if pullSizes[1] == 0 || pullSizes[2]+pullSizes[3] == 0 {
		t.Errorf("pulls by mobs taken %v, want some single and some chain pulls", pullSizes)
	}
	for size := range pullSizes {
		if size > 3 {
			t.Errorf("a pull took %d mobs, but the rotation looks only while fewer than 3 are pulled", size)
		}
	}
}

// TestLevelingWalkIntoBoltRange: the rotation walks up to a mob only until it's 30 yards
// away, then bolts it, so every pull of a mob found further away comes from 30 yards.
func TestLevelingWalkIntoBoltRange(t *testing.T) {
	rotation := `{"type":"TypeAPL","priorityList":[
		{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"40%"}}}},"castSpell":{"spellId":{"spellId":8008}}}},
		{"action":{"castSpell":{"spellId":{"spellId":403},"target":{"type":"PullTarget"}}}},
		{"action":{"levelingWalkToPullTarget":{"within":{"const":{"val":"30"}}}}}
	]}`
	req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{}, 1), rotation)
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	found := map[string]int{}
	far := 0
	logLines(t, result.Logs, func(at float64, line string) {
		if f := foundLine.FindStringSubmatch(line); f != nil {
			found[f[1]], _ = strconv.Atoi(f[2])
		} else if p := pulledLine.FindStringSubmatch(line); p != nil {
			from, _ := strconv.Atoi(p[3])
			if found[p[1]] > 30 {
				far++
				if from != 30 {
					t.Errorf("at %.2f %s found %d yards away was pulled from %d, want 30", at, p[1], found[p[1]], from)
				}
			}
		}
	})
	if far < 3 {
		t.Errorf("%d mobs found beyond 30 yards, want a few", far)
	}
}

var bearingLine = regexp.MustCompile(`Leveling: found Target \d+ \(level \d+\) \d+ yards away, bearing (\d+)°`)

// TestLevelingFoundBearing: the mobs we find stand in any direction from us, not on a line.
func TestLevelingFoundBearing(t *testing.T) {
	req := levelingRequest(t, &proto.LevelingOptions{}, 5)
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	quadrants := [4]int{}
	n := 0
	logLines(t, result.Logs, func(at float64, line string) {
		if b := bearingLine.FindStringSubmatch(line); b != nil {
			bearing, _ := strconv.Atoi(b[1])
			quadrants[(bearing/90)%4]++
			n++
		}
	})
	t.Logf("%d mobs found, by quadrant: %v", n, quadrants)
	for q, count := range quadrants {
		if n < 40 || float64(count) < 0.1*float64(n) {
			t.Errorf("%d of %d mobs found in quadrant %d, want them spread all around", count, n, q)
		}
	}
}
