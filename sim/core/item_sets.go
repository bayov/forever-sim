package core

import (
	"fmt"
	"slices"

	"github.com/wowsims/classic/sim/core/stats"
)

type ItemSet struct {
	ID              int32
	Name            string
	AlternativeName string

	// Maps set piece requirement to an ApplyEffect function that will be called
	// before the Sim starts.
	//
	// The function should apply any benefits provided by the set bonus.
	Bonuses map[int32]ApplyEffect
}

func (set *ItemSet) Items() []Item {
	var items []Item
	for _, item := range ItemsByID {
		if findItemSet(item) == set {
			items = append(items, item)
		}
	}
	// Sort so the order of IDs is always consistent, for tests.
	slices.SortFunc(items, func(a, b Item) int {
		return int(a.ID - b.ID)
	})
	return items
}

var sets []*ItemSet

// Registers a new ItemSet with item IDs populated.
func NewItemSet(set ItemSet) *ItemSet {
	foundID := set.ID == 0
	foundName := false
	foundAlternativeName := set.AlternativeName == ""
	for _, item := range ItemsByID {
		if item.SetName == "" && item.SetID == 0 {
			continue
		}
		foundID = foundID || (item.SetID > 0 && item.SetID == set.ID)
		foundName = foundName || (item.SetName != "" && item.SetName == set.Name)
		foundAlternativeName = foundAlternativeName || (item.SetName != "" && item.SetName == set.AlternativeName)
		if foundID && foundName && foundAlternativeName {
			break
		}
	}
	// A set Forever added has only pieces the client puts in it by ID, and no piece carries its
	// name (item names come from Wowhead's Classic tooltips).
	foundName = foundName || (set.ID > 0 && foundID)

	if WITH_DB {
		if !foundID {
			panic(fmt.Sprintf("No items found for set id %d", set.ID))
		}
		if !foundName {
			panic("No items found for set " + set.Name)
		}
		if len(set.AlternativeName) > 0 && !foundAlternativeName {
			panic("No items found for set alternative " + set.AlternativeName)
		}
	}

	sets = append(sets, &set)
	return &set
}

func (character *Character) HasSetBonus(set *ItemSet, numItems int32) bool {
	if character.Env != nil && character.Env.IsFinalized() {
		panic("HasSetBonus is very slow and should never be called after finalization. Try caching the value during construction instead!")
	}

	if _, ok := set.Bonuses[numItems]; !ok {
		panic(fmt.Sprintf("Item set %s does not have a bonus with %d pieces.", set.Name, numItems))
	}

	var count int32
	for _, item := range character.Equipment {
		if item.SetName == "" && item.SetID == 0 {
			continue
		}
		if (item.SetName != "" && (item.SetName == set.Name || item.SetName == set.AlternativeName)) || (item.SetID > 0 && item.SetID == set.ID) {
			count++
			if count >= numItems {
				return true
			}
		}
	}

	return false
}

type ActiveSetBonus struct {
	// Name of the set.
	Name string

	// Number of pieces required for this bonus.
	NumPieces int32

	// Function for applying the effects of this set bonus.
	BonusEffect ApplyEffect
}

// The set an item counts toward: the set registered with the client's set ID for it when there is
// one, so sets with different names but a shared ID count together, and otherwise the set of its
// name. Forever's new pieces, like the Imperial Plate Gauntlets and the Premier PvP pieces, carry
// only the client's set ID, since an item's set name comes from Wowhead's Classic tooltips.
func findItemSet(item Item) *ItemSet {
	if item.SetID > 0 {
		for _, set := range sets {
			if set.ID == item.SetID {
				return set
			}
		}
	}
	if item.SetName != "" {
		for _, set := range sets {
			if set.Name == item.SetName || set.AlternativeName == item.SetName {
				return set
			}
		}
	}
	return nil
}

// Returns a list describing all active set bonuses.
func (character *Character) GetActiveSetBonuses() []ActiveSetBonus {
	var activeBonuses []ActiveSetBonus

	setItemCount := make(map[*ItemSet]int32)
	for _, item := range character.Equipment {
		foundSet := findItemSet(item)
		if foundSet != nil {
			setItemCount[foundSet]++
			if bonusEffect, ok := foundSet.Bonuses[setItemCount[foundSet]]; ok {
				activeBonuses = append(activeBonuses, ActiveSetBonus{
					Name:        foundSet.Name,
					NumPieces:   setItemCount[foundSet],
					BonusEffect: bonusEffect,
				})
			}
		}
	}

	return activeBonuses
}

// Apply effects from item set bonuses.
func (character *Character) applyItemSetBonusEffects(agent Agent) {
	activeSetBonuses := character.GetActiveSetBonuses()

	for _, activeSetBonus := range activeSetBonuses {
		activeSetBonus.BonusEffect(agent)
	}
}

// Returns the names of all active set bonuses.
func (character *Character) GetActiveSetBonusNames() []string {
	activeSetBonuses := character.GetActiveSetBonuses()

	names := make([]string, len(activeSetBonuses))
	for i, activeSetBonus := range activeSetBonuses {
		names[i] = fmt.Sprintf("%s (%dpc)", activeSetBonus.Name, activeSetBonus.NumPieces)
	}
	return names
}

// The bonuses of a set Forever reworked: the forever ones under the Forever ruleset and the
// classic ones under any other. The two may grant their bonuses at different piece counts.
func RulesetSetBonuses(classic, forever map[int32]ApplyEffect) map[int32]ApplyEffect {
	bonuses := make(map[int32]ApplyEffect, len(classic)+len(forever))
	for _, layout := range []map[int32]ApplyEffect{classic, forever} {
		for numPieces := range layout {
			bonuses[numPieces] = func(agent Agent) {
				bonus := classic[numPieces]
				if agent.GetCharacter().Env.IsForever() {
					bonus = forever[numPieces]
				}
				if bonus != nil {
					bonus(agent)
				}
			}
		}
	}
	return bonuses
}

// A set bonus that only adds stats.
func SetBonusStats(bonus stats.Stats) ApplyEffect {
	return func(agent Agent) {
		agent.GetCharacter().AddStats(bonus)
	}
}
