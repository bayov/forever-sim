package main

import "github.com/wowsims/classic/sim/core/proto"

// Racial actives the autocast line doesn't use, for the templates of the races that have them.
// A cast of a spell the character doesn't know is dropped with a warning, so these lines do
// nothing for the other races.

// readLeyline is the Skyborne High Order's Read Leyline (1259705, sim/core/racials_forever.go):
// a 2 sec cast on a 2 min cooldown, 50% more Spirit-based mana regeneration for 15 sec.
var readLeyline = spellID(1259705)

// readLeylineKnob: Read Leyline below this much mana in percent, 0 never (the default, which adds
// no line).
func readLeylineKnob() Knob {
	return Knob{Name: "leylineMana", Default: 0, Min: 0, Max: 100, Step: 20}
}

// readLeylineLines cast Read Leyline under the leylineMana knob, with more than its 15 sec left.
// extra is an added condition, nil for none.
func readLeylineLines(k Knobs, extra value) []*proto.APLListItem {
	if k["leylineMana"] <= 0 {
		return nil
	}
	cond := and(lt(currentManaPercent(), num(k["leylineMana"]/100)), gt(remainingTime(), seconds(17)), extra)
	if k["leylineMana"] >= 100 {
		cond = and(gt(remainingTime(), seconds(17)), extra)
	}
	return []*proto.APLListItem{cast(readLeyline, cond, "Read Leyline (Skyborne High Order) for the mana regeneration. Other races ignore the warning on this line.")}
}
