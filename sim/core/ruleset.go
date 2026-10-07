package core

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Simulation embeds Environment, so sim.IsForever() resolves here too.
func (env *Environment) IsForever() bool {
	return env.Ruleset == proto.Ruleset_RulesetForever
}

// Periodic damage rolls for crits under the Forever ruleset. Spells that
// should keep ticking for flat damage opt out with SpellFlagNoPeriodicCrit.
func (dot *Dot) canCrit(sim *Simulation) bool {
	return sim.IsForever() && !dot.Spell.Flags.Matches(SpellFlagNoPeriodicCrit)
}

// Ticks roll against the caster's crit chance at the time of the tick instead of the
// chance snapshotted when the dot went up. Dots that are applied by hand, like Deep
// Wounds, never snapshot one at all, so rolling live is also the only way for them to
// crit at the right rate.
func (dot *Dot) critCheck(sim *Simulation, target *Unit, attackTable *AttackTable) bool {
	if dot.Spell.SchoolIndex == stats.SchoolIndexPhysical {
		return dot.Spell.PhysicalCritCheck(sim, attackTable)
	}
	return dot.Spell.MagicCritCheck(sim, target)
}

// Forever has no hit or crit rating. Its items give a flat percent, at every level.
//
// wowhead's Forever item pages still print the stat as TBC style rating, like 4 hit
// rating "(0.40% @ L60)" on Pyrewood Signet Ring or 14 crit rating on Fletcher's Gloves,
// which the game shows as "Improves your chance to get a critical strike by 1.0%". So we
// turn the rating back into percent with TBC's level 60 values.
const (
	HitRatingPerPercent  = 10.0
	CritRatingPerPercent = 14.0
)

func (character *Character) addEquipRatings(equipStats stats.Stats) stats.Stats {
	for _, item := range character.Equipment {
		equipStats[stats.MeleeHit] += item.HitRating / HitRatingPerPercent
		equipStats[stats.MeleeCrit] += item.CritRating / CritRatingPerPercent
	}
	return equipStats
}

// Forever pays out hit and critical strike from gear against every kind of attack
// rather than splitting them into a melee and a spell pool. Attribute conversions are
// untouched: only the hit and crit an item spells out become universal.
func (character *Character) unifyEquipHitAndCrit(equipStats stats.Stats) stats.Stats {
	hit := equipStats[stats.MeleeHit] + equipStats[stats.SpellHit]
	crit := equipStats[stats.MeleeCrit] + equipStats[stats.SpellCrit]

	equipStats[stats.MeleeHit] = hit
	equipStats[stats.SpellHit] = hit
	equipStats[stats.MeleeCrit] = crit
	equipStats[stats.SpellCrit] = crit

	return equipStats
}
