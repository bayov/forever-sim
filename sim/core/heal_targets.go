package core

import (
	"github.com/wowsims/classic/sim/core/proto"
)

// Heal targets for a healer's rotation (docs/healing_sim.md): Most Hurt and Lowest Health pick
// among the living players in the healer's assignment (Player.heal_assignment), or the whole raid
// without one. Nobody hurt gives none, so a heal on it waits.

// setHealAssignment resolves Player.heal_assignment once the raid is built. Player and Tank
// references only; others are skipped.
func (character *Character) setHealAssignment(refs []*proto.UnitReference) {
	character.healAssignment = nil
	for _, ref := range refs {
		if ref == nil || (ref.Type != proto.UnitReference_Player && ref.Type != proto.UnitReference_Tank) {
			continue
		}
		if unit := character.Env.GetUnit(ref, &character.Unit); unit != nil {
			character.healAssignment = append(character.healAssignment, unit)
		}
	}
}

// HealCandidates is who our heal targets pick from: our assignment, or the whole raid.
func (unit *Unit) HealCandidates() []*Unit {
	if len(unit.healAssignment) > 0 {
		return unit.healAssignment
	}
	return unit.Env.Raid.AllPlayerUnits
}

// pickHealTarget is the living, hurt player among our candidates missing the most health (Most
// Hurt) or at the lowest health percent (Lowest Health). Ties go to the first in order. Nil when
// nobody is hurt.
func (unit *Unit) pickHealTarget(pick proto.UnitReference_Type) *Unit {
	var best *Unit
	bestScore := 0.0
	for _, player := range unit.HealCandidates() {
		if player.dead || !player.HasHealthBar() {
			continue
		}
		missing := player.MaxHealth() - player.CurrentHealth()
		if missing <= 0 {
			continue
		}
		score := missing
		if pick == proto.UnitReference_LowestHealth {
			score = 1 - player.CurrentHealthPercent()
		}
		if best == nil || score > bestScore {
			best, bestScore = player, score
		}
	}
	return best
}

// MostHurtAmong is the living player in units missing the most health, leaving out skip, for a
// heal that jumps (Chain Heal). Nil when nobody else is hurt.
func MostHurtAmong(units []*Unit, skip ...*Unit) *Unit {
	var best *Unit
	bestMissing := 0.0
outer:
	for _, player := range units {
		if player.dead || !player.HasHealthBar() {
			continue
		}
		for _, s := range skip {
			if s == player {
				continue outer
			}
		}
		if missing := player.MaxHealth() - player.CurrentHealth(); missing > bestMissing {
			best, bestMissing = player, missing
		}
	}
	return best
}
