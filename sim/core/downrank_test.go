package core

import (
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

// TestDownrank checks Forever's downranking penalty as RegisterSpell applies it: the formula at
// its edges, a sub-spell's DownrankLevel, the scope and the kind read from the ProcMask, Classic, DoTs and HoTs, and the
// spells it leaves alone.
func TestDownrank(t *testing.T) {
	forever := &Environment{Ruleset: proto.Ruleset_RulesetForever}
	classic := &Environment{Ruleset: proto.Ruleset_RulesetClassic}

	cases := []struct {
		name       string
		env        *Environment
		level      int32
		learn      int
		rank       int
		kind       DownrankKind
		scope      DownrankKind
		wantFactor float64
	}{
		{"17 levels past the learn level keeps it all", forever, 60, 43, 1, DownrankDamage, ForeverDownrankScope, 1},
		{"18 levels past keeps 0.95", forever, 60, 42, 1, DownrankDamage, ForeverDownrankScope, .95},
		{"Lightning Bolt rank 7 at 60 keeps 0.75", forever, 60, 38, 7, DownrankDamage, ForeverDownrankScope, .75},
		{"37 levels past is clamped at 0", forever, 60, 23, 1, DownrankHeal, ForeverDownrankScope, 0},
		{"far past stays at 0", forever, 60, 1, 1, DownrankHeal, ForeverDownrankScope, 0},
		{"a rank above the cap keeps it all", forever, 30, 30, 1, DownrankHeal, ForeverDownrankScope, 1},
		{"Classic keeps its coefficient", classic, 60, 1, 1, DownrankDamage, ForeverDownrankScope, 1},
		{"an exempt spell keeps it", forever, 60, 1, 1, DownrankExempt, ForeverDownrankScope, 1},
		{"a spell without a RequiredLevel keeps it", forever, 60, 0, 1, DownrankAuto, ForeverDownrankScope, 1},
		{"a rank-less sub-spell takes its DownrankLevel", forever, 60, -38, 0, DownrankAuto, ForeverDownrankScope, .75},
		{"auto reads a heal from its ProcMask", forever, 60, 38, 1, DownrankAuto, DownrankHeal, .75},
		{"auto reads anything else as damage", forever, 60, 38, 1, DownrankAuto, DownrankDamage, .75},
		{"auto damage is out of a heals-only scope", forever, 60, 38, 2, DownrankAuto, DownrankHeal, 1},
		{"a rank-less spell keeps it", forever, 60, 1, 0, DownrankDamage, ForeverDownrankScope, 1},
		{"heals only: damage keeps it", forever, 60, 1, 1, DownrankDamage, DownrankHeal, 1},
		{"heals only: heals take it", forever, 60, 38, 1, DownrankHeal, DownrankHeal, .75},
	}

	saved := ForeverDownrankScope
	defer func() { ForeverDownrankScope = saved }()

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ForeverDownrankScope = c.scope
			unit := &Unit{Level: c.level, Env: c.env}
			config := SpellConfig{
				Rank:             c.rank,
				RequiredLevel:    max(c.learn, 0),
				DownrankLevel:    max(-c.learn, 0),
				Downrank:         c.kind,
				ProcMask:         Ternary(c.rank == 1 && c.kind == DownrankAuto && c.scope == DownrankHeal, ProcMaskSpellHealing, ProcMaskSpellDamage),
				BonusCoefficient: .8,
				Dot:              DotConfig{BonusCoefficient: .2},
				Hot:              DotConfig{BonusCoefficient: .4},
			}
			config.SpellSchool = SpellSchoolArcane
			unit.applyDownrank(&config)
			for _, got := range []struct {
				what      string
				got, base float64
			}{
				{"direct", config.BonusCoefficient, .8},
				{"DoT", config.Dot.BonusCoefficient, .2},
				{"HoT", config.Hot.BonusCoefficient, .4},
			} {
				if want := got.base * c.wantFactor; !WithinToleranceFloat64(want, got.got, 1e-9) {
					t.Errorf("%s coefficient = %v, want %v", got.what, got.got, want)
				}
			}
		})
	}

	// A physical ability's coefficient is weapon bonus damage, not spell power.
	physical := SpellConfig{Rank: 1, RequiredLevel: 1, SpellSchool: SpellSchoolPhysical, BonusCoefficient: 1}
	(&Unit{Level: 60, Env: forever}).applyDownrank(&physical)
	if physical.BonusCoefficient != 1 {
		t.Errorf("physical coefficient = %v, want 1", physical.BonusCoefficient)
	}

	// Heals and damage alike (user ruling, 2026-11-04).
	if saved != DownrankHeal|DownrankDamage {
		t.Errorf("ForeverDownrankScope = %v, want heals and damage", saved)
	}
}
