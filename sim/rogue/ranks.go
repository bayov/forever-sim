package rogue

import "github.com/wowsims/classic/sim/core"

// The value for the highest rank a rogue of this level has learned, from a table keyed by
// the level the trainer teaches each rank at. The zero value when no rank is known yet.
func rankAt[V any](level int32, ranks map[int32]V) V {
	var best int32 = -1
	var value V
	for learnedAt, v := range ranks {
		if learnedAt <= level && learnedAt > best {
			best, value = learnedAt, v
		}
	}
	return value
}

// The spell ID of the highest rank learned, with every rank registered as one family so a
// rotation naming another rank still finds it.
func rankSpellID(level int32, ranks map[int32]int32) int32 {
	ids := make([]int32, 0, len(ranks))
	for _, id := range ranks {
		ids = append(ids, id)
	}
	core.RegisterSpellRanks(ids...)
	return rankAt(level, ranks)
}
