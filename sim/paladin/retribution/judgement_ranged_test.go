package retribution

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestJudgementsAreRanged checks Blizzard's hotfix of November 11 (t/2403040 post 8): "Judgement
// changed to now be considered a Ranged Attack. This means it no longer counts as a Melee attack
// for the purposes of item and effect procs and uses the Paladin's Ranged Hit chance." Judgement
// of Righteousness and of Command are ranged specials, not melee, so melee-keyed procs miss them,
// and they roll hit and crit on the ranged attack table, not the main hand's.
func TestJudgementsAreRanged(t *testing.T) {
	for _, c := range []struct {
		name string
		seal proto.PaladinSeal
		id   int32
	}{
		{"Judgement of Righteousness", proto.PaladinSeal_Righteousness, 20286},
		{"Judgement of Command", proto.PaladinSeal_Command, 20966},
	} {
		sim, ret := retSimOptions(t, Phase45RetTalents, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: arcaniteReaper},
			1, proto.MobType_MobTypeUnknown, &proto.PaladinOptions{PrimarySeal: c.seal})
		judge := ret.GetSpell(core.ActionID{SpellID: c.id})
		if judge == nil {
			t.Fatalf("%s (%d) is not registered", c.name, c.id)
		}
		if judge.ProcMask != core.ProcMaskRangedSpecial || judge.ProcMask.Matches(core.ProcMaskMelee) {
			t.Errorf("%s proc mask %v, want ranged special and no melee", c.name, judge.ProcMask)
		}
		if judge.CastType != proto.CastType_CastTypeRanged || judge.DefenseType != core.DefenseTypeRanged {
			t.Errorf("%s cast type %v, defense %v, want ranged", c.name, judge.CastType, judge.DefenseType)
		}

		// The roll is on the ranged table: a sure miss there misses every time, and a sure
		// miss on the main hand's table doesn't reach it.
		tables := ret.AttackTables[ret.CurrentTarget.UnitIndex]
		ranged, mh := tables[proto.CastType_CastTypeRanged], tables[proto.CastType_CastTypeMainHand]
		metrics := &judge.SpellMetrics[ret.CurrentTarget.UnitIndex]
		cast := func(rangedMiss, mhMiss float64) int32 {
			savedRanged, savedMH := ranged.BaseMissChance, mh.BaseMissChance
			defer func() { ranged.BaseMissChance, mh.BaseMissChance = savedRanged, savedMH }()
			ranged.BaseMissChance, mh.BaseMissChance = rangedMiss, mhMiss
			before := metrics.Misses
			for range 50 {
				judge.ApplyEffects(sim, ret.CurrentTarget, judge)
			}
			return metrics.Misses - before
		}
		if got := cast(2, 0); got != 50 {
			t.Errorf("%s missed %d of 50 with the ranged table always missing, want 50", c.name, got)
		}
		if got := cast(-1, 2); got != 0 {
			t.Errorf("%s missed %d of 50 with only the main hand table always missing, want 0", c.name, got)
		}
	}
}
