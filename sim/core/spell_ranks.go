package core

import "sync"

// Spell rank families. A rotation written at level 60 names the top rank of each ability,
// and a lower level character knows a lower rank under another spell ID. A class registers
// every rank of an ability as one family, and an APL lookup that misses by exact ID then
// takes whichever rank of that family the unit knows.

var spellRankFamilies = map[int32]int32{}

// Every character registers its ranks as it is built, and sims build characters in
// parallel, so the registry is locked. Lookups happen when a rotation is built.
var spellRankMutex sync.RWMutex

// RegisterSpellRanks marks the spell IDs as ranks of one ability.
func RegisterSpellRanks(spellIDs ...int32) {
	if len(spellIDs) == 0 {
		return
	}
	spellRankMutex.Lock()
	defer spellRankMutex.Unlock()
	head := spellIDs[0]
	for _, id := range spellIDs {
		if existing, ok := spellRankFamilies[id]; ok {
			head = existing
			break
		}
	}
	for _, id := range spellIDs {
		spellRankFamilies[id] = head
	}
}

// SameRankFamily is true when both IDs are ranks of the same ability.
func SameRankFamily(a, b int32) bool {
	if a == b {
		return true
	}
	spellRankMutex.RLock()
	defer spellRankMutex.RUnlock()
	fa, ok := spellRankFamilies[a]
	fb, ok2 := spellRankFamilies[b]
	return ok && ok2 && fa == fb
}

// The unit's spell from the same rank family as the action, when the exact ID is unknown.
// A class that registers every rank it has learned (shaman) gets the highest one.
func (unit *Unit) getSpellByRankFamily(actionID ActionID) *Spell {
	if actionID.SpellID == 0 {
		return nil
	}
	var best *Spell
	for _, spell := range unit.Spellbook {
		if spell.ActionID.Tag == actionID.Tag && SameRankFamily(spell.ActionID.SpellID, actionID.SpellID) {
			if best == nil || spell.Rank > best.Rank {
				best = spell
			}
		}
	}
	return best
}

func (at *auraTracker) getAuraByRankFamily(actionID ActionID) *Aura {
	if actionID.SpellID == 0 {
		return nil
	}
	var best *Aura
	for _, aura := range at.auras {
		if aura.ActionID.Tag == actionID.Tag && SameRankFamily(aura.ActionID.SpellID, actionID.SpellID) {
			if best == nil || aura.ActionID.SpellID > best.ActionID.SpellID {
				best = aura
			}
		}
	}
	return best
}

// RankScaling is how a vanilla spell rank grows with the character's level. A rank is
// PerLevel weaker for every level the character is below MaxLevel, down to the level the
// trainer teaches it at. A zero MaxLevel means the rank does not scale.
type RankScaling struct {
	MaxLevel int32
	PerLevel float64
}

// At is the rank's value at the character's level, from the value the rank has at its
// MaxLevel (which is what its tooltip shows at 60, and what the class tables hold).
func (s RankScaling) At(maxValue float64, level int32) float64 {
	if s.MaxLevel == 0 || level >= s.MaxLevel {
		return maxValue
	}
	return maxValue - s.PerLevel*float64(s.MaxLevel-level)
}

// HighestRankAt is the highest rank whose learn level the character has reached, from a
// table indexed by rank with a zero entry first. Zero when no rank is known yet.
func HighestRankAt[L int | int32](level int32, learnLevels []L) int {
	rank := 0
	for r := 1; r < len(learnLevels); r++ {
		if int32(learnLevels[r]) <= level {
			rank = r
		}
	}
	return rank
}
