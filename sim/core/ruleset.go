package core

import (
	"github.com/wowsims/classic/sim/core/proto"
)

func (sim *Simulation) Ruleset() proto.Ruleset {
	return sim.Options.Ruleset
}

func (sim *Simulation) IsForever() bool {
	return sim.Ruleset() == proto.Ruleset_RulesetForever
}

// Periodic damage rolls for crits under the Forever ruleset. Spells that
// should keep ticking for flat damage opt out with SpellFlagNoPeriodicCrit.
func (dot *Dot) canCrit(sim *Simulation) bool {
	return sim.IsForever() &&
		dot.Spell.DefenseType != DefenseTypeNone &&
		!dot.Spell.Flags.Matches(SpellFlagNoPeriodicCrit)
}
