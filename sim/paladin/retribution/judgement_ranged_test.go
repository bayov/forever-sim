package retribution

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestJudgementsAreRanged checks Blizzard's hotfix of November 11 (t/2403040 post 8): "Judgement
// changed to now be considered a Ranged Attack. This means it no longer counts as a Melee attack
// for the purposes of item and effect procs and uses the Paladin's Ranged Hit chance." Judgement
// of Righteousness is a ranged special, not melee, so melee-keyed procs miss it, and it rolls hit
// and crit on the ranged attack table, not the main hand's. The November 13 hotfixes (t/2403040,
// "November 13, 2026") made Command melee again: "Judgement of Command now activates on-hit
// effects and counts as a melee attack. All other judgements will continue to be considered
// Ranged attacks." So Command is a main hand melee special on the main hand's table (also client
// SpellCategories Melee, and a log after the hotfix shows it taking the melee-only Two-Handed Weapon
// Specialization bonus; user ruling 2026-11-15).
func TestJudgementsAreRanged(t *testing.T) {
	for _, c := range []struct {
		name   string
		seal   proto.PaladinSeal
		id     int32
		ranged bool
	}{
		{"Judgement of Righteousness", proto.PaladinSeal_Righteousness, 20286, true},
		{"Judgement of Command", proto.PaladinSeal_Command, 20966, false},
	} {
		sim, ret := retSimOptions(t, Phase45RetTalents, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: arcaniteReaper},
			1, proto.MobType_MobTypeUnknown, &proto.PaladinOptions{PrimarySeal: c.seal})
		judge := ret.GetSpell(core.ActionID{SpellID: c.id})
		if judge == nil {
			t.Fatalf("%s (%d) is not registered", c.name, c.id)
		}
		if c.ranged {
			if judge.ProcMask != core.ProcMaskRangedSpecial || judge.ProcMask.Matches(core.ProcMaskMelee) {
				t.Errorf("%s proc mask %v, want ranged special and no melee", c.name, judge.ProcMask)
			}
			if judge.CastType != proto.CastType_CastTypeRanged || judge.DefenseType != core.DefenseTypeRanged {
				t.Errorf("%s cast type %v, defense %v, want ranged", c.name, judge.CastType, judge.DefenseType)
			}
		} else {
			if judge.ProcMask != core.ProcMaskMeleeMHSpecial {
				t.Errorf("%s proc mask %v, want main hand melee special", c.name, judge.ProcMask)
			}
			// EV-0155, EV-0156: its dummy and its damage each roll the weapon's procs.
			if !judge.Flags.Matches(core.SpellFlagRollsWeaponProcsTwice) {
				t.Errorf("%s doesn't roll weapon procs twice", c.name)
			}
			if judge.CastType != proto.CastType_CastTypeMainHand || judge.DefenseType != core.DefenseTypeMelee {
				t.Errorf("%s cast type %v, defense %v, want main hand melee", c.name, judge.CastType, judge.DefenseType)
			}
		}

		// The roll is on its own table: a sure miss there misses every time, and a sure miss
		// on the other table doesn't reach it.
		tables := ret.AttackTables[ret.CurrentTarget.UnitIndex]
		ranged, mh := tables[proto.CastType_CastTypeRanged], tables[proto.CastType_CastTypeMainHand]
		metrics := &judge.SpellMetrics()[ret.CurrentTarget.UnitIndex]
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
		own, other := cast(2, -1), cast(-1, 2)
		if !c.ranged {
			own, other = other, own
		}
		if own != 50 {
			t.Errorf("%s missed %d of 50 with its own table always missing, want 50", c.name, own)
		}
		if other != 0 {
			t.Errorf("%s missed %d of 50 with only the other table always missing, want 0", c.name, other)
		}
	}
}
