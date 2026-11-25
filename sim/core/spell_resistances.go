package core

import (
	"math"

	"github.com/wowsims/classic/sim/core/stats"
)

func (result *SpellResult) applyResistances(sim *Simulation, spell *Spell, isPeriodic bool, attackTable *AttackTable) {
	// Forever's Stoneskin Totem takes a flat amount off each physical hit before armor.
	//
	// On the beta our rank 3 totem's 11 fit the hits we took only when it comes off before armor
	// (shaman_audit.md 7.29). We leave periodic damage alone, because no enemy in the sim deals
	// physical damage over time.
	//
	// A hit it cuts still deals at least 1 damage after armor. On the beta our rank 1 totem's 4
	// left 1 damage from raw 3 and 4 swings, and no hit in any log dealt 0.
	flatReduced := false
	if bonus := attackTable.Defender.PseudoStats.BonusPhysicalDamageTakenBeforeArmor; bonus != 0 && !isPeriodic &&
		spell.SpellSchool.Matches(SpellSchoolPhysical) && !spell.Flags.Matches(SpellFlagIgnoreTargetModifiers) {
		flatReduced = result.Damage > 0
		result.Damage = max(0, result.Damage+bonus)
	}

	resistanceMultiplier, outcome := spell.ResistanceMultiplier(sim, isPeriodic, attackTable)

	result.Damage *= resistanceMultiplier
	if flatReduced {
		result.Damage = max(1, result.Damage)
	}
	result.Outcome |= outcome

	result.ResistanceMultiplier = resistanceMultiplier
	result.PreOutcomeDamage = result.Damage
}

// Modifies damage based on Armor or Magic resistances, depending on the damage type.
func (spell *Spell) ResistanceMultiplier(sim *Simulation, isPeriodic bool, attackTable *AttackTable) (float64, HitOutcome) {
	if spell.Flags.Matches(SpellFlagIgnoreResists) {
		return 1, OutcomeEmpty
	}

	if spell.SpellSchool.Matches(SpellSchoolPhysical) {
		if spell.SchoolIndex == stats.SchoolIndexPhysical || MultiSchoolShouldUseArmor(spell, attackTable.Defender) {
			// All physical dots (Bleeds) ignore armor.
			if isPeriodic {
				return 1, OutcomeEmpty
			}

			// Physical resistance (armor).
			return attackTable.GetArmorDamageModifier(), OutcomeEmpty
		}
	}

	// Magical resistance.
	if attackTable.Attacker.foreverResists() {
		return spell.foreverResistanceMultiplier(sim, isPeriodic, attackTable)
	}

	if spell.Flags.Matches(SpellFlagBinary) {
		return 1, OutcomeEmpty
	}

	resistanceRoll := sim.RandomFloat("Partial Resist")

	threshold00, threshold25, threshold50 := attackTable.GetPartialResistThresholds(spell, spell.Flags.Matches(SpellFlagPureDot))
	//if sim.Log != nil {
	//	sim.Log("Resist thresholds: %0.04f, %0.04f, %0.04f", threshold00, threshold25, threshold50)
	//}

	if resistanceRoll > threshold00 {
		// No partial resist.
		return 1, OutcomeEmpty
	} else if resistanceRoll > threshold25 {
		return 0.75, OutcomePartial1_4
	} else if resistanceRoll > threshold50 {
		return 0.5, OutcomePartial2_4
	} else {
		return 0.25, OutcomePartial3_4
	}
}

// Decide whether to use armor for physical multi school spells.
//
// TODO: This is most likely not accurate for the case: armor near resistance but not 0
//
// A short test showed that the game uses armor if it's far enough below resistance,
// but not simply if it's lower.
// 49 (and above) armor vs 57 res => used resistance
// 7 (and below) armor vs 57 res => used armor/no partials anymore
// If level based resist is used in this decission process is also not known as it was tested PvP.
//
// For most purposes this should work fine for now, but should be properly tested and fixed if
// spells using it become important and boss armor can actually go below (level based) resistance values.
func MultiSchoolShouldUseArmor(spell *Spell, target *Unit) bool {
	resistance := 100000.0
	lowestIsArmor := true
	for _, baseSchoolIndex := range spell.SchoolBaseIndices {
		resiVal := target.GetResistanceForSchool(baseSchoolIndex)
		if resiVal < resistance {
			resistance = resiVal
			lowestIsArmor = baseSchoolIndex == stats.SchoolIndexPhysical
		}
	}
	return lowestIsArmor
}

func (at *AttackTable) GetArmorDamageModifier() float64 {
	armorPenRating := at.Attacker.stats[stats.ArmorPenetration]
	defenderArmor := max(at.Defender.Armor()*(1-at.Attacker.PseudoStats.ArmorIgnorePercent)-armorPenRating, 0.0)
	// Armor takes off at most 75% (17,265 armor against a level 63 boss), the game's cap. Only a
	// player tank's armor gets near it.
	return max(1-defenderArmor/(defenderArmor+400+85*float64(at.Attacker.Level)), 0.25)
}

func (at *AttackTable) GetPartialResistThresholds(spell *Spell, pureDot bool) (float64, float64, float64) {
	return at.Defender.partialResistRollThresholds(spell, at.Attacker, pureDot)
}

func (at *AttackTable) GetBinaryHitChance(spell *Spell) float64 {
	return at.Defender.binaryHitChance(spell, at.Attacker)
}

// Only for base schools!
func (unit *Unit) GetResistanceForSchool(schoolIndex stats.SchoolIndex) float64 {
	switch schoolIndex {
	case stats.SchoolIndexNone:
		return 0
	case stats.SchoolIndexPhysical:
		return unit.GetStat(stats.Armor)
	case stats.SchoolIndexArcane:
		return unit.GetStat(stats.ArcaneResistance)
	case stats.SchoolIndexFire:
		return unit.GetStat(stats.FireResistance)
	case stats.SchoolIndexFrost:
		return unit.GetStat(stats.FrostResistance)
	case stats.SchoolIndexHoly:
		return 0 // Holy resistance doesn't exist.
	case stats.SchoolIndexNature:
		return unit.GetStat(stats.NatureResistance)
	case stats.SchoolIndexShadow:
		return unit.GetStat(stats.ShadowResistance)
	default:
		return 0
	}
}

// All of the following calculations are based on this guide:
// https://royalgiraffe.github.io/resist-guide

// partialResistPerLevel is the average partial resist a non-binary spell meets for each level its
// target is above the caster: Classic's 2%, 6% against a level 63 boss. Forever has none: about 520
// beta hits on normal mobs 3 to 5 levels up were never partly resisted while mobs with school
// resistance were (user ruling 2026-10-16, report assumption level_partial_resist), and
// wowsims/forever's logs agree (ed4b4e93e5: none in thousands of hits on mobs 1 to 9 levels up).
func (unit *Unit) partialResistPerLevel() float64 {
	if unit.foreverResists() {
		return 0
	}
	return AverageMagicPartialResistPerLevelMultiplier
}

// foreverResists is whether the unit's spells meet Forever's resistance model.
func (unit *Unit) foreverResists() bool {
	return unit.Env != nil && unit.Env.IsForever()
}

// schoolResistance is the target's resistance to the spell's school, the lowest of a multi school
// spell's schools.
func (unit *Unit) schoolResistance(spell *Spell) float64 {
	if !spell.SchoolIndex.IsMultiSchool() {
		return unit.GetResistanceForSchool(spell.SchoolIndex)
	}
	resistance := 1000.0
	for _, baseSchoolIndex := range spell.SchoolBaseIndices {
		resistance = min(resistance, unit.GetResistanceForSchool(baseSchoolIndex))
	}
	return resistance
}

// foreverResistanceMultiplier is a magic hit's resist under Forever: a roll in 10% steps around the
// target's average resist (wowsims/forever 679ca6f875, chi-squared fits of their combat logs; our
// Elder Cloud Serpents took 20% or 30%, never 25 or 50%). Spell Penetration beyond the resistance
// makes the average negative, and the same roll adds 10% or 20% damage (679ca6f875: a mage with 15
// Spell Penetration gained 10 or 20% on 1,671 of 2,960 hits on mobs with no resistance).
//
// EV-0142 (wowsims/forever, 992 Warcraft Logs reports and 213 combat logs of the live realms,
// 2026-09-27 to 2026-11-24): a binary spell's hit roll takes its whole resist, so one that lands
// takes no partial resist (0 of 43,465 binary direct hits, 0 of 12,082 binary DoT ticks), while
// Spell Piercing still rolls its bonus on it (Frostbolt 1,863 of 6,276 against 1,802 expected). No
// partial resist takes more than 75% of a hit: every partial on a landed hit lands on a 10% step up
// to 70% or on exactly 75% (69 direct hits and 2 ticks at 75%, none above), so the roll's 80, 90 and
// 100% steps take 75%, and a spell that is not binary never resists whole (Arcane Explosion RESIST 0
// of 16,264).
func (spell *Spell) foreverResistanceMultiplier(sim *Simulation, isPeriodic bool, attackTable *AttackTable) (float64, HitOutcome) {
	if spell.Flags.Matches(SpellFlagNoPartialResist) {
		return 1, OutcomeEmpty
	}

	averageResist := attackTable.Defender.foreverAverageResist(spell, attackTable.Attacker, isPeriodic)
	if averageResist == 0 || averageResist > 0 && spell.Flags.Matches(SpellFlagBinary) && !spell.Flags.Matches(SpellFlagBinaryPartialResist) {
		return 1, OutcomeEmpty
	}
	chances := resistTenthsChances(math.Abs(averageResist))

	// Expected damage takes the roll's mean, with the 75% step.
	if spell.expecting {
		share := 0.0
		for tenths, chance := range chances {
			share += chance * resistedShare(tenths)
		}
		return 1 - math.Copysign(share, averageResist), OutcomeEmpty
	}

	tenths := rollResistTenths(sim, chances)
	if tenths == 0 {
		return 1, OutcomeEmpty
	}
	share := resistedShare(tenths)
	if averageResist < 0 {
		return 1 + share, OutcomeEmpty
	}
	if share < 0.1*float64(tenths) {
		return 1 - share, OutcomePartial3_4
	}
	return 1 - share, partialResistOutcome(tenths)
}

// resistedShare is the share of the hit a roll of the given tenths takes: at most 75% (EV-0142).
func resistedShare(tenths int) float64 {
	return min(0.1*float64(tenths), 0.75)
}

// foreverAverageResist is the average share of a spell's damage the target resists under Forever,
// R / (|R| + 8L + 19) for the caster's level L, with the caster's Spell Penetration taken off R
// unless the spell is flagged SpellFlagNoSpellPiercing. The target's own resistance stops at 0, so a
// debuff that takes it below 0 adds nothing, and Holy has none. A target above the caster's level
// resists no more for its level.
//
// EV-0091 (wowsims/forever, Spell Piercing logs at levels 20 to 50 and mobs near 38 casting at
// players of known resistance): K = 8L + 19 fits every level (chi-squared 1.1), where the 9L we took
// from their earlier fit of levels 20 to 30 (8249a6d0e3) gives 18.9 and sits 3.7 sigma off at level
// 50 (a mage with 15 piercing: a bonus on 1,500 of 5,756 hits, K = 417, 9L 450). 499 at level 60.
//
// EV-0142: Spell Piercing beyond the resistance gives no bonus while the share it makes is under 1%
// (5 piercing at level 60, s 0.99%: a bonus on 0.64% of hits, the floor without piercing; 6 and up
// follow 7.5s). A DoT's or channel's tick meets a tenth of a positive share and takes no piercing
// bonus (Corruption, Siphon Life, Wrack 0.06 to 0.08 of the expected bonuses; Drain Life ticks a
// tenth, 0.087; Hurricane 19 of 7,577). Spells flagged SpellFlagFullResistanceTicks meet the full
// share and take the bonus in full (Devouring Plague 0.95, Pyroblast's DoT 0.98, Starshards 1.19;
// EV-0092: Pyroblast's DoT 175 of 632 binned against the casters' 27.3% direct rate).
func (unit *Unit) foreverAverageResist(spell *Spell, attacker *Unit, isPeriodic bool) float64 {
	if spell.SchoolIndex <= stats.SchoolIndexPhysical || spell.SpellSchool.Matches(SpellSchoolHoly) {
		return 0
	}

	resistance := max(0, unit.schoolResistance(spell))
	if !spell.Flags.Matches(SpellFlagNoSpellPiercing) {
		resistance -= attacker.stats[stats.SpellPenetration]
	}
	averageResist := resistance / (math.Abs(resistance) + 8*float64(attacker.Level) + 19)
	if averageResist < 0 && averageResist > -0.01 {
		averageResist = 0
	}

	if isPeriodic && !spell.Flags.Matches(SpellFlagFullResistanceTicks) {
		averageResist = max(0, averageResist) / 10
	}
	return averageResist
}

// resistTenthsChances is the chance of each step of Forever's partial resist roll, in tenths of the
// damage, around the average share a (wowsims/forever 679ca6f875; their Thunderhawks fit it at
// chi-squared 4.5 and 4.9 on 2 degrees of freedom). Up to a = 10% it gives 0, 1 or 2 tenths at
// 1 - 7.5a, 5a and 2.5a, from 90% the same mirrored from 10 tenths, and in between each 10% step x
// within 20% of a at 0.5 - 2.5|x - a|. Its mean is a, before the 75% step.
func resistTenthsChances(a float64) (chances [11]float64) {
	switch {
	case a <= 0.1:
		chances[0], chances[1], chances[2] = 1-7.5*a, 5*a, 2.5*a
	case a >= 0.9:
		chances[10], chances[9], chances[8] = 1-7.5*(1-a), 5*(1-a), 2.5*(1-a)
	default:
		for tenths := range chances {
			chances[tenths] = max(0, 0.5-2.5*math.Abs(float64(tenths)/10-a))
		}
	}
	return chances
}

// rollResistTenths draws one step of the roll, taking the steps from 0 up.
func rollResistTenths(sim *Simulation, chances [11]float64) int {
	roll := sim.RandomFloat("Partial Resist")
	cumulative, last := 0.0, 0
	for tenths, chance := range chances {
		if chance <= 0 {
			continue
		}
		cumulative += chance
		last = tenths
		if roll < cumulative {
			return tenths
		}
	}
	return last
}

func (unit *Unit) resistCoeff(spell *Spell, attacker *Unit, binary bool, pureDot bool) float64 {
	if spell.SchoolIndex <= stats.SchoolIndexPhysical {
		return 0
	}

	resistance := max(0, unit.schoolResistance(spell)-attacker.stats[stats.SpellPenetration])

	resistanceCap := float64(attacker.Level * 5)
	resistanceCoef := resistance / resistanceCap

	// Pre-TBC all dots that don't have an initial damage component
	// use a 1/10 of the resistance score
	if pureDot {
		resistanceCoef /= 10
	}

	if perLevel := attacker.partialResistPerLevel(); perLevel > 0 && !binary && unit.Type == EnemyUnit && unit.Level > attacker.Level {
		avgMitigationAdded := perLevel * float64(unit.Level-attacker.Level)
		// coef is scaled 0 to 1, not 0 to 0.75
		resistanceCoef += avgMitigationAdded * 1 / 0.75
	}

	return min(1, resistanceCoef)
}

// binaryHitChance is the share of a binary spell's hit rolls its target doesn't resist in full. Under
// Forever that takes the average resist as the chance (wowsims/forever 679ca6f875: Insect Swarm 3 of 10
// fully resisted by Grimtotems that took 34.9% off Wrath).
func (unit *Unit) binaryHitChance(spell *Spell, attacker *Unit) float64 {
	if attacker.foreverResists() {
		return 1 - max(0, unit.foreverAverageResist(spell, attacker, false))
	}
	resistCoeff := unit.resistCoeff(spell, attacker, true, false)
	return 1 - 0.75*resistCoeff
}

// Roll threshold for each type of partial resist.
func (unit *Unit) partialResistRollThresholds(spell *Spell, attacker *Unit, pureDot bool) (float64, float64, float64) {
	resistCoeff := unit.resistCoeff(spell, attacker, false, pureDot)

	// Based on the piecewise linear regression estimates at https://royalgiraffe.github.io/partial-resist-table.
	if val := resistCoeff * 3; val <= 1 {
		return 0.76 * val, 0.21 * val, 0.03 * val
	} else if val <= 2 {
		val -= 1
		return 0.76 + 0.24*val, 0.21 + 0.57*val, 0.03 + 0.19*val
	} else {
		val -= 2
		return 1, 0.78 + 0.18*val, 0.22 + 0.58*val
	}
}

// BinaryUnderForever is SpellFlagBinary under Forever and nothing outside it, for a spell whose
// row makes it binary on the live realms (EV-0142: a non-damage aura on the target, an interrupt,
// E_DUMMY, E_SCRIPT_EFFECT, E_ENERGIZE or E_CREATE_AREATRIGGER beside the damage) where Classic
// keeps it partial.
func BinaryUnderForever(env *Environment) SpellFlag {
	if env != nil && env.IsForever() {
		return SpellFlagBinary
	}
	return SpellFlagNone
}
