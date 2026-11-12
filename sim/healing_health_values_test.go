package sim

// Needs the item database: go test -tags with_db.

import (
	"strconv"
	"strings"
	"testing"
)

// Current Health % of Lowest Health reads the health the healer has seen, like the pick: a line
// under it must not fire within the reaction time of the hit.
func TestHealingFightHealthValuesSeen(t *testing.T) {
	const reactionMs = 1500
	rsr := healingRaid(t, 1, 1, 0)
	healerRotation(rsr, 1, `{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{"sourceUnit":{"type":"LowestHealth"}}},"rhs":{"const":{"val":"0.99"}}}},"castSpell":{"spellId":{"spellId":10396},"target":{"type":"Player","index":0}}}}`)
	rsr.Raid.Parties[0].Players[1].ReactionTimeMs = reactionMs
	rsr.SimOptions.Iterations = 1
	rsr.SimOptions.Debug = true
	hit, cast := -1.0, -1.0
	for _, line := range strings.Split(runRaid(t, rsr).Logs, "\n") {
		if !strings.HasPrefix(line, "[") {
			continue
		}
		at, err := strconv.ParseFloat(line[1:strings.Index(line, "]")], 64)
		if err != nil {
			continue
		}
		if hit < 0 && strings.Contains(line, "] [0 (#1)] Spent ") && strings.Contains(line, " health from ") {
			hit = at
		}
		if cast < 0 && strings.Contains(line, "Casting {SpellID: 10396}") {
			cast = at
		}
	}
	t.Logf("first health change %.3f, first cast %.3f", hit, cast)
	if cast < 0 || hit < 0 || cast < hit+reactionMs/1000.0-0.001 {
		t.Errorf("heal cast at %.3f for a hit at %.3f, within the %d ms reaction time", cast, hit, reactionMs)
	}
}
