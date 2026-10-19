package enhancement

import (
	"regexp"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

var (
	flameShockCastLine = regexp.MustCompile(`\] \[ \(#1\)\] Casting \{SpellID: 8052\}`)
	flameShockHitLine  = regexp.MustCompile(`\] \[ \(#1\)\] \[(Target \d+)\] \{SpellID: 8052\} (?:Hit|Crit) for`)
	flameShockDotLine  = regexp.MustCompile(`\] \[(Target \d+)\] Aura (gained|faded): \{SpellID: 8052\}`)
)

// TestLevelingDotFollowsTarget: Dot Is Active looks at the DoT on the mob we fight now. It used
// to look at the mob that was our target when the rotation was built, so in a leveling session
// "Flame Shock when its DoT is down" recast Flame Shock on every cooldown and clipped its DoT.
func TestLevelingDotFollowsTarget(t *testing.T) {
	req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{MobLevelOffset: -3}, 1), rotationOf(healLine,
		`{"action":{"condition":{"not":{"val":{"dotIsActive":{"spellId":{"spellId":8052}}}}},"castSpell":{"spellId":{"spellId":8052}}}}`))
	req.SimOptions.Debug = true
	result := runLeveling(t, req)

	dotted := map[string]bool{}
	casting, casts, clipped := false, 0, 0
	logLines(t, result.Logs, func(at float64, line string) {
		if flameShockCastLine.MatchString(line) {
			casting = true
			casts++
		} else if m := flameShockHitLine.FindStringSubmatch(line); m != nil && casting {
			casting = false
			if dotted[m[1]] {
				clipped++
			}
		} else if m := flameShockDotLine.FindStringSubmatch(line); m != nil {
			dotted[m[1]] = m[2] == "gained"
		}
	})
	if clipped > 0 {
		t.Errorf("%d of %d Flame Shocks hit a mob whose DoT was still up", clipped, casts)
	}
	if casts < 10 {
		t.Errorf("%d Flame Shock casts, want one a mob or so", casts)
	}
	if result.EncounterMetrics.Leveling.MobsKilled < 5 {
		t.Errorf("%.0f mobs killed, want a few", result.EncounterMetrics.Leveling.MobsKilled)
	}
}

var earthShockCastLine = regexp.MustCompile(`\] \[ \(#1\)\] Casting \{SpellID: 8042\}`)

// TestLevelingPullTargetHealth: Current Health % reads the Pull Target. It used to give up on
// it when the rotation was built, before any mob was found, and the line's condition was lost.
func TestLevelingPullTargetHealth(t *testing.T) {
	pulls := func(op string) int {
		rotation := rotationOf(healLine,
			`{"action":{"condition":{"cmp":{"op":"`+op+`","lhs":{"currentHealthPercent":{"sourceUnit":{"type":"PullTarget"}}},"rhs":{"const":{"val":"50%"}}}},"castSpell":{"spellId":{"spellId":8042},"target":{"type":"PullTarget"}}}}`,
			`{"action":{"levelingWalkToPullTarget":{"within":{"const":{"val":"20"}}}}}`)
		req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{}, 1), rotation)
		req.SimOptions.Debug = true
		result := runLeveling(t, req)
		n := 0
		logLines(t, result.Logs, func(_ float64, line string) {
			if earthShockCastLine.MatchString(line) {
				n++
			}
		})
		return n
	}
	// A mob we haven't pulled is at full health.
	if n := pulls("OpGt"); n < 5 {
		t.Errorf("%d Earth Shock pulls above 50%% health, want one a mob", n)
	}
	if n := pulls("OpLt"); n != 0 {
		t.Errorf("%d Earth Shock pulls below 50%% health, want none", n)
	}
}
