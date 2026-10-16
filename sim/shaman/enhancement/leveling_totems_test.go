package enhancement

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

const healLine = `{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"40%"}}}},"castSpell":{"spellId":{"spellId":8008}}}}`

// searchingOnly is a condition true only while we search out of combat: moving, no mob on us
// and none found.
const searchingOnly = `{"isMoving":{}},{"cmp":{"op":"OpEq","lhs":{"levelingMobsOnUs":{}},"rhs":{"const":{"val":"0"}}}},{"cmp":{"op":"OpEq","lhs":{"levelingDistance":{"targetUnit":{"type":"PullTarget"}}},"rhs":{"const":{"val":"0"}}}}`

func rotationOf(lines ...string) string {
	return `{"type":"TypeAPL","priorityList":[` + strings.Join(lines, ",") + `]}`
}

var (
	soeCastLine   = regexp.MustCompile(`\] \[ \(#1\)\] Casting \{SpellID: 8160\}`)
	soeGainedLine = regexp.MustCompile(`\] \[ \(#1\)\] Aura gained: \{SpellID: 8160\}`)
	soeFadedLine  = regexp.MustCompile(`\] \[ \(#1\)\] Aura faded: \{SpellID: 8160\}`)
)

func logLines(t *testing.T, logs string, each func(at float64, line string)) {
	for _, line := range strings.Split(logs, "\n") {
		if m := logTime.FindStringSubmatch(line); m != nil {
			at, _ := strconv.ParseFloat(m[1], 64)
			each(at, line)
		}
	}
}

// TestLevelingTotemReach: Strength of Earth Totem dropped as we start a search reaches us
// for its 30 yards, so we lose its buff about 30 / 7 sec into the run. Walking back to it with a
// mob on us brings the buff back.
func TestLevelingTotemReach(t *testing.T) {
	rotation := rotationOf(healLine,
		`{"action":{"condition":{"and":{"vals":[`+searchingOnly+`,{"not":{"val":{"totemRemainingTime":{"totemType":"Earth"}}}}]}},"castSpell":{"spellId":{"spellId":8160}}}}`,
		`{"action":{"condition":{"and":{"vals":[{"cmp":{"op":"OpGe","lhs":{"levelingMobsOnUs":{}},"rhs":{"const":{"val":"1"}}}},{"cmp":{"op":"OpGt","lhs":{"totemDistance":{"totemType":"Earth"}},"rhs":{"const":{"val":"30"}}}}]}},"shamanWalkToTotem":{"totemType":"Earth","within":{"const":{"val":"5"}}}}}`)
	req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{}, 1), rotation)
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	searchAt, castAt, fades, regains := -1.0, -1.0, 0, 0
	inReach := false
	logLines(t, result.Logs, func(at float64, line string) {
		switch {
		case searchingLine.MatchString(line):
			searchAt = at
		case soeCastLine.MatchString(line):
			castAt = at
		case soeGainedLine.MatchString(line):
			if at != castAt {
				regains++
			}
			inReach = true
		case soeFadedLine.MatchString(line) && inReach:
			fades++
			inReach = false
			// Dropped as a search started, it's out of reach 30 yards from the totem, which lands
			// 1.5 to 3.5 yards from us in any direction.
			if castAt >= 0 && castAt >= searchAt && castAt-searchAt < 0.1 {
				if d := at - castAt; d < 26.5/7-0.1 || d > 33.5/7+1.1 {
					t.Errorf("at %.2f the buff faded %.2f sec after we dropped the totem and ran on, want %.2f to %.2f", at, d, 26.5/7, 33.5/7+1)
				}
			}
			castAt = -1
		}
	})
	t.Logf("%d fades out of reach, %d times back in reach", fades, regains)
	if fades < 4 || regains < 3 {
		t.Errorf("%d fades out of reach and %d times back in reach, want a few of each", fades, regains)
	}
}

// totemSession runs 20 sessions and gives the player's metrics.
func totemSession(t *testing.T, opts *proto.LevelingOptions, lines ...string) *proto.UnitMetrics {
	result := runLeveling(t, withRotation(t, levelingRequest(t, opts, 20), rotationOf(lines...)))
	return result.RaidMetrics.Parties[0].Players[0]
}

func spellDamage(player *proto.UnitMetrics, spellID int32) (damage float64, casts int32) {
	for _, action := range player.Actions {
		if action.Id.GetSpellId() == spellID {
			for _, target := range action.Targets {
				damage += target.Damage
				casts += target.Casts
			}
		}
	}
	return
}

// TestTotemicRecall gives back a quarter of the totem's 65 mana (Strength of Earth Totem
// rank 2), and takes the totem away, so the rotation drops it again.
func TestTotemicRecall(t *testing.T) {
	player := totemSession(t, &proto.LevelingOptions{}, healLine,
		`{"action":{"condition":{"not":{"val":{"totemRemainingTime":{"totemType":"Earth"}}}},"castSpell":{"spellId":{"spellId":8160}}}}`,
		`{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"totemRemainingTime":{"totemType":"Earth"}},"rhs":{"const":{"val":"295s"}}}},"castSpell":{"spellId":{"spellId":36936}}}}`)
	_, drops := spellDamage(player, 8160)
	_, recalls := spellDamage(player, 36936)
	refund := 0.0
	for _, r := range player.Resources {
		if r.Id.GetSpellId() == 36936 {
			refund += r.Gain
		}
	}
	t.Logf("%d drops, %d recalls, %.0f mana back", drops, recalls, refund)
	if recalls < 50 || drops < recalls {
		t.Fatalf("%d drops and %d recalls, want a recall 5 sec after each drop", drops, recalls)
	}
	if got := refund / float64(recalls); math.Abs(got-65*0.25) > 1e-6 {
		t.Errorf("a recall gave back %.3f mana, want %.3f", got, 65*0.25)
	}
}

// TestTotemicProjection: a Magma Totem dropped while we search is left behind (a mob that
// runs past it on its way to us may take a pulse), but projecting it to us once a mob is on
// us brings its pulses back.
//
// We search for 2 to 10 sec, so we walk well away from where we dropped it.
func TestTotemicProjection(t *testing.T) {
	opts := &proto.LevelingOptions{SearchMin: 2, SearchMax: 10}
	magma := `{"action":{"condition":{"and":{"vals":[` + searchingOnly + `,{"not":{"val":{"totemRemainingTime":{"totemType":"Fire"}}}}]}},"castSpell":{"spellId":{"spellId":8190}}}}`
	project := `{"action":{"condition":{"and":{"vals":[{"cmp":{"op":"OpGe","lhs":{"levelingMobsOnUs":{}},"rhs":{"const":{"val":"1"}}}},{"cmp":{"op":"OpGt","lhs":{"totemDistance":{"totemType":"Fire"}},"rhs":{"const":{"val":"8"}}}}]}},"castSpell":{"spellId":{"spellId":437009}}}}`
	left, _ := spellDamage(totemSession(t, opts, healLine, magma), 8187)
	player := totemSession(t, opts, healLine, magma, project)
	projected, _ := spellDamage(player, 8187)
	_, projections := spellDamage(player, 437009)
	t.Logf("Magma Totem damage: %.0f left behind, %.0f projected with %d projections", left, projected, projections)
	if left > projected/20 || projected <= 0 || projections == 0 {
		t.Errorf("Magma Totem dealt %.0f left behind and %.0f projected to us (%d projections), want next to none and some", left, projected, projections)
	}
}

// TestLevelingTotemsVanishOnDeath: when we die our totems go, so back from the corpse run
// the rotation drops Strength of Earth Totem again at once, long before its 5 min are up.
func TestLevelingTotemsVanishOnDeath(t *testing.T) {
	req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{MobLevelOffset: 2}, 1), rotationOf(append([]string{
		`{"action":{"condition":{"not":{"val":{"totemRemainingTime":{"totemType":"Earth"}}}},"castSpell":{"spellId":{"spellId":8160}}}}`}, packLines(4)...)...))
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	var revives, casts []float64
	logLines(t, result.Logs, func(at float64, line string) {
		switch {
		case strings.Contains(line, "Leveling: died"):
			if at+180 < 1200 {
				revives = append(revives, at+180)
			}
		case soeCastLine.MatchString(line):
			casts = append(casts, at)
		}
	})
	if len(revives) == 0 {
		t.Fatal("no deaths")
	}
	for _, revive := range revives {
		recast := false
		for _, cast := range casts {
			recast = recast || math.Abs(cast-revive) < 0.06
		}
		if !recast {
			t.Errorf("back from the corpse run at %.2f without dropping the totem again, want it gone", revive)
		}
	}
}

// TestLevelingChase: walking back to a totem left behind while a mob runs at us, the mob
// turns to follow and still reaches us, a little later than if we'd stood still.
func TestLevelingChase(t *testing.T) {
	rotation := rotationOf(healLine,
		`{"action":{"condition":{"and":{"vals":[`+searchingOnly+`,{"not":{"val":{"totemRemainingTime":{"totemType":"Earth"}}}}]}},"castSpell":{"spellId":{"spellId":8160}}}}`,
		`{"action":{"condition":{"and":{"vals":[{"cmp":{"op":"OpGe","lhs":{"levelingMobsRunningIn":{}},"rhs":{"const":{"val":"1"}}}},{"cmp":{"op":"OpGt","lhs":{"totemDistance":{"totemType":"Earth"}},"rhs":{"const":{"val":"30"}}}}]}},"shamanWalkToTotem":{"totemType":"Earth","within":{"const":{"val":"0"}}}}}`)
	req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{}, 1), rotation)
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	chasing := regexp.MustCompile(`Leveling: walking \d+ yards toward a point`)
	walks := 0
	logLines(t, result.Logs, func(at float64, line string) {
		if chasing.MatchString(line) {
			walks++
		}
	})
	m := result.EncounterMetrics.Leveling
	t.Logf("%d walks back with a mob running at us, %.1f kills", walks, m.MobsKilled)
	if walks < 3 {
		t.Errorf("%d walks back to the totem with a mob running in, want a few", walks)
	}
	if m.MobsKilled < 15 {
		t.Errorf("%.1f mobs killed, want a steady stream: a mob chasing us must still reach us", m.MobsKilled)
	}
}

// TestOwnTotemRank: our own Strength of Earth Totem gives the Strength of the rank we cast,
// rank 1's 7 rather than the 14 of rank 2 a level 30 knows.
func TestOwnTotemRank(t *testing.T) {
	for _, c := range []struct {
		spell    int32
		strength string
	}{{8075, `"Strength": 7.000`}, {8160, `"Strength": 14.000`}} {
		req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{}, 1), rotationOf(
			`{"action":{"condition":{"not":{"val":{"totemRemainingTime":{"totemType":"Earth"}}}},"castSpell":{"spellId":{"spellId":`+strconv.Itoa(int(c.spell))+`}}}}`))
		req.SimOptions.Debug = true
		result := runLeveling(t, req)
		if !strings.Contains(result.Logs, "Dynamic stat change: {"+c.strength) {
			t.Errorf("Strength of Earth Totem %d never gave %s", c.spell, c.strength)
		}
	}
}
