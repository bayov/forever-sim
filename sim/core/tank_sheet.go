package core

import (
	"math"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The tank sheet is a tank's defenses against the boss that attacks it, worked out from the
// same attack table the boss's swings roll on (outcomeEnemyMeleeWhite), with the thresholds
// tanks gear for: the Defense that stops crits, the coverage that pushes crushing blows off the
// table, armor and effective health, and our own miss, dodge and parry against the boss, which
// set threat. See docs/tank_sim.md.
//
// It's taken once, as the first iteration's prepull starts, so permanent buffs and consumables
// are on and no cooldown or proc is. What the fight did with them (Shield Block, Holy Shield) shows
// in the boss's swing outcomes instead.

// takeTankSheet takes the sheet at the start of the first iteration's prepull, against the first
// target attacking us: after the fight's permanent auras are on, before anything we cast before
// the pull (a prepull Holy Shield) or during the fight.
func (character *Character) takeTankSheet() {
	if !character.Metrics.isTanking || character.Metrics.tankSheet != nil {
		return
	}
	for _, target := range character.Env.Encounter.TargetUnits {
		if target.CurrentTarget == &character.Unit {
			character.Metrics.tankSheet = TankSheet(&character.Unit, target)
			return
		}
	}
}

// TankSheet is defender's sheet against attacker's white swings from the front, from both
// units' stats now.
func TankSheet(defender *Unit, attacker *Unit) *proto.TankSheet {
	table := NewAttackTable(attacker, defender, nil)
	bonusDefense := defender.stats[stats.Defense]
	defense := float64(defender.Level*5) + bonusDefense/DefenseRatingPerDefense

	sheet := &proto.TankSheet{
		BossLevel: attacker.Level,
		Defense:   defense,
	}

	// The same chances as the applyEnemyAttackTable* functions, in their order. Each one is
	// what's left of the roll after those before it.
	miss := max(0, table.BaseMissChance+attacker.PseudoStats.IncreasedMissChance+bonusDefense*DefenseRatingToChanceReduction)
	if attacker.AutoAttacks.IsDualWielding && !attacker.PseudoStats.DisableDWMissPenalty {
		miss += 0.19
	}
	dodge := max(0, table.BaseDodgeChance+defender.GetStat(stats.Dodge)/100)
	parry := 0.0
	if defender.PseudoStats.CanParry {
		parry = max(0, table.BaseParryChance+defender.GetStat(stats.Parry)/100)
	}
	block := 0.0
	if defender.PseudoStats.CanBlock {
		block = max(0, table.BaseBlockChance+defender.stats[stats.Block]/BlockRatingPerBlockChance/100)
	}
	critUncapped := table.BaseCritChance - bonusDefense*DefenseRatingToChanceReduction - defender.PseudoStats.ReducedCritTakenChance
	crit := max(0, critUncapped)
	crush := table.BaseCrushChance

	left := 1.0
	take := func(chance float64) float64 {
		got := min(chance, left)
		left -= got
		return got
	}
	sheet.Miss = take(miss)
	sheet.Dodge = take(dodge)
	sheet.Parry = take(parry)
	sheet.Block = take(block)
	sheet.Crit = take(crit)
	sheet.Crush = take(crush)
	sheet.Hit = left

	// Crits stop where the crit chance reaches 0. Each point of Defense takes 0.04% off it. Past
	// that, it's the Defense we could drop to.
	sheet.DefenseCritImmune = defense + math.Ceil(critUncapped/(MissDodgeParryBlockCritChancePerDefense/100)-1e-9)
	sheet.Coverage = miss + dodge + parry + block
	if crush > 0 {
		sheet.CrushGap = max(0, 1-crit-sheet.Coverage)
	}

	sheet.Armor = defender.Armor()
	sheet.ArmorReduction = 1 - table.GetArmorDamageModifier()
	sheet.BlockValue = defender.BlockValue()
	sheet.Health = defender.MaxHealth()
	taken := (1 - sheet.ArmorReduction) * defender.PseudoStats.DamageTakenMultiplier *
		defender.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical]
	if taken > 0 {
		sheet.EffectiveHealth = sheet.Health / taken
	}

	// Our side, from the attack tables our attacks roll on.
	if tables := defender.AttackTables; int(attacker.UnitIndex) < len(tables) {
		if ours := tables[attacker.UnitIndex][proto.CastType_CastTypeMainHand]; ours != nil {
			hit := defender.stats[stats.MeleeHit]/(MeleeHitRatingPerHitChance*100) - ours.HitSuppression
			sheet.MeleeMiss = max(0, ours.BaseMissChance-max(hit, 0))
			expertise := defender.stats[stats.Expertise] / 100
			sheet.BossDodge = max(0, ours.BaseDodgeChance-attacker.PseudoStats.DodgeReduction-expertise)
			sheet.BossParry = max(0, ours.BaseParryChance-expertise)
			spellHit := defender.stats[stats.SpellHit] / (SpellHitRatingPerHitChance * 100)
			sheet.SpellMiss = max(0.01, ours.BaseSpellMissChance-spellHit)
		}
	}
	return sheet
}
