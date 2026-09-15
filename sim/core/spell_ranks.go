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
func (unit *Unit) getSpellByRankFamily(actionID ActionID) *Spell {
	if actionID.SpellID == 0 {
		return nil
	}
	for _, spell := range unit.Spellbook {
		if spell.ActionID.Tag == actionID.Tag && SameRankFamily(spell.ActionID.SpellID, actionID.SpellID) {
			return spell
		}
	}
	return nil
}

func (at *auraTracker) getAuraByRankFamily(actionID ActionID) *Aura {
	if actionID.SpellID == 0 {
		return nil
	}
	for _, aura := range at.auras {
		if aura.ActionID.Tag == actionID.Tag && SameRankFamily(aura.ActionID.SpellID, actionID.SpellID) {
			return aura
		}
	}
	return nil
}
