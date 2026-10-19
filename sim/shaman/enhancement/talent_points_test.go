package enhancement

import (
	"regexp"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

var boltCastLine = regexp.MustCompile(`\] \[ \(#1\)\] Casting \{SpellID: 403\}`)

// TestTalentPoints: Talent Points reads our points in a talent, and a line it gates doesn't run
// without them. A talent the class doesn't have reads 0 too, rather than losing the condition.
func TestTalentPoints(t *testing.T) {
	bolts := func(talents, talent string) int {
		req := withRotation(t, levelingRequest(t, &proto.LevelingOptions{MobLevelOffset: -3}, 1), rotationOf(healLine,
			`{"action":{"condition":{"cmp":{"op":"OpGe","lhs":{"talentPoints":{"talent":"`+talent+`"}},"rhs":{"const":{"val":"3"}}}},"castSpell":{"spellId":{"spellId":403}}}}`))
		req.Raid.Parties[0].Players[0].TalentsString = talents
		req.SimOptions.Debug = true
		result := runLeveling(t, req)
		n := 0
		logLines(t, result.Logs, func(_ float64, line string) {
			if boltCastLine.MatchString(line) {
				n++
			}
		})
		return n
	}
	if n := bolts("-005", "ancestralKnowledge"); n < 5 {
		t.Errorf("%d Lightning Bolts with Ancestral Knowledge 5/5, want many", n)
	}
	if n := bolts("-002", "ancestralKnowledge"); n != 0 {
		t.Errorf("%d Lightning Bolts with Ancestral Knowledge 2/5, want none", n)
	}
	if n := bolts("-005", "ancestral_knowledge"); n < 5 {
		t.Errorf("%d Lightning Bolts by the proto field name, want many", n)
	}
	if n := bolts("-005", "noSuchTalent"); n != 0 {
		t.Errorf("%d Lightning Bolts on a talent the class doesn't have, want none", n)
	}
}
