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

// Bonus healing on Forever gear carries a damage component with it, so that healing
// gear is not dead weight outside a raid. Hide of the Wild reads 42 healing and 14
// damage, which is the only published pair, so a third is the rate used here. It feeds
// SpellDamage rather than SpellPower because the damage half does not heal.
const ForeverHealingToSpellDamage = 1.0 / 3.0

func (character *Character) addHealingSpellDamage(equipStats stats.Stats) stats.Stats {
	equipStats[stats.SpellDamage] += equipStats[stats.HealingPower] * ForeverHealingToSpellDamage
	return equipStats
}

// RatingPerPercent is how much rating makes 1% at a level, given what it takes at 60.
//
// wowhead's Forever item pages print 4 hit rating as "(0.40% @ L60)", so 10 hit rating
// and 14 crit rating are 1% at level 60. Those are the TBC client's level 60 values, and
// nothing we have shows the Forever client's values below 60. So we assume the TBC
// curve, which takes (level - 8) / 52 of the level 60 value from level 10 on, and 2 / 52
// below that. At level 20 that makes 14 crit rating on Fletcher's Gloves worth 4.3%.
func RatingPerPercent(level int32, at60 float64) float64 {
	return at60 * float64(max(min(level, 60), 10)-8) / 52
}

const (
	HitRatingPerPercentAt60  = 10.0
	CritRatingPerPercentAt60 = 14.0
)

func (character *Character) addEquipRatings(equipStats stats.Stats) stats.Stats {
	for _, item := range character.Equipment {
		equipStats[stats.MeleeHit] += item.HitRating / RatingPerPercent(character.Level, HitRatingPerPercentAt60)
		equipStats[stats.MeleeCrit] += item.CritRating / RatingPerPercent(character.Level, CritRatingPerPercentAt60)
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
