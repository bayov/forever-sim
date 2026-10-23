package core

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// foreverResistUnits returns a level 60 Forever caster and a level 63 target with the given Shadow
// resistance.
func foreverResistUnits(shadowResistance float64) (*Unit, *Unit) {
	env := &Environment{Ruleset: proto.Ruleset_RulesetForever}
	attacker := &Unit{Type: PlayerUnit, Level: 60, Env: env}
	defender := &Unit{Type: EnemyUnit, Level: 63, Env: env}
	defender.stats[stats.ShadowResistance] = shadowResistance
	return attacker, defender
}

func shadowSpell(attacker *Unit, flags SpellFlag) *Spell {
	return &Spell{Unit: attacker, SpellSchool: SpellSchoolShadow, SchoolIndex: stats.SchoolIndexShadow, Flags: flags}
}

// TestForeverAverageResist: R / (|R| + 9L) for the caster's level, a tenth of it on a DoT's tick
// but not on a channel's or Insect Swarm's, and nothing for the target's level.
func TestForeverAverageResist(t *testing.T) {
	for _, c := range []struct {
		name       string
		resistance float64
		tick       bool
		flags      SpellFlag
		want       float64
	}{
		{"no resistance", 0, false, 0, 0},
		{"no resistance, tick", 0, true, 0, 0},
		{"100 resistance", 100, false, 0, 100.0 / 640},
		{"100 resistance, binary", 100, false, SpellFlagBinary, 100.0 / 640},
		{"100 resistance, tick", 100, true, 0, 100.0 / 640 / 10},
		{"100 resistance, pure DoT tick", 100, true, SpellFlagPureDot, 100.0 / 640 / 10},
		{"100 resistance, channel tick", 100, true, SpellFlagChanneled, 100.0 / 640},
		{"100 resistance, full resistance tick", 100, true, SpellFlagFullResistanceTicks, 100.0 / 640},
	} {
		attacker, defender := foreverResistUnits(c.resistance)
		if got := defender.foreverAverageResist(shadowSpell(attacker, c.flags), attacker, c.tick); math.Abs(got-c.want) > 1e-12 {
			t.Errorf("%s: average resist %v, want %v", c.name, got, c.want)
		}
	}

	attacker, defender := foreverResistUnits(100)
	if got, want := defender.binaryHitChance(shadowSpell(attacker, SpellFlagBinary), attacker), 1-100.0/640; math.Abs(got-want) > 1e-12 {
		t.Errorf("binary hit chance against 100 resistance = %v, want %v", got, want)
	}

	holy := &Spell{Unit: attacker, SpellSchool: SpellSchoolHoly, SchoolIndex: stats.SchoolIndexHoly}
	attacker.stats[stats.SpellPenetration] = 50
	if got := defender.foreverAverageResist(holy, attacker, false); got != 0 {
		t.Errorf("a Holy spell with 50 Spell Penetration: average resist %v, want 0", got)
	}
}

// TestForeverSpellPenetrationAddsDamage: Spell Penetration beyond the target's resistance makes the
// average negative, and the roll adds 10% or 20% damage, never a resisted outcome. A target
// resistance below 0 counts as 0.
func TestForeverSpellPenetrationAddsDamage(t *testing.T) {
	attacker, defender := foreverResistUnits(0)
	attacker.stats[stats.SpellPenetration] = 15
	spell := shadowSpell(attacker, 0)

	a := 15.0 / (15 + 540)
	if got := defender.foreverAverageResist(spell, attacker, false); math.Abs(got+a) > 1e-12 {
		t.Errorf("15 Spell Penetration against no resistance: average resist %v, want %v", got, -a)
	}
	_, debuffed := foreverResistUnits(-75)
	if got := debuffed.foreverAverageResist(spell, attacker, false); math.Abs(got+a) > 1e-12 {
		t.Errorf("15 Spell Penetration against -75 resistance: average resist %v, want %v", got, -a)
	}

	sim := &Simulation{rand: NewSplitMix(1)}
	attackTable := NewAttackTable(attacker, defender, nil)
	const n = 100_000
	counts := map[float64]int{}
	for range n {
		multiplier, outcome := spell.ResistanceMultiplier(sim, false, attackTable)
		if outcome != OutcomeEmpty {
			t.Fatalf("a Spell Penetration bonus came with outcome %v", outcome)
		}
		counts[math.Round(multiplier*10)/10]++
	}
	for multiplier, want := range map[float64]float64{1: 1 - 7.5*a, 1.1: 5 * a, 1.2: 2.5 * a} {
		if got := float64(counts[multiplier]) / n; math.Abs(got-want) > 0.005 {
			t.Errorf("share of hits at x%v = %.4f, want %.4f", multiplier, got, want)
		}
	}
	if len(counts) != 3 {
		t.Errorf("multipliers rolled: %v, want only 1, 1.1 and 1.2", counts)
	}
}

// TestForeverResistRollSteps: the roll spreads around the average resist in 10% steps, and its
// mean is the average.
func TestForeverResistRollSteps(t *testing.T) {
	for _, c := range []struct {
		a    float64
		want map[int]float64
	}{
		{0.06, map[int]float64{0: 0.55, 1: 0.30, 2: 0.15}},
		{0.25, map[int]float64{1: 0.125, 2: 0.375, 3: 0.375, 4: 0.125}},
		{0.30, map[int]float64{2: 0.25, 3: 0.5, 4: 0.25}},
		{0.95, map[int]float64{10: 0.625, 9: 0.25, 8: 0.125}},
	} {
		sim := &Simulation{rand: NewSplitMix(7)}
		const n = 100_000
		counts := map[int]int{}
		total := 0
		for range n {
			tenths := rollResistTenths(sim, c.a)
			counts[tenths]++
			total += tenths
		}
		for tenths, want := range c.want {
			if got := float64(counts[tenths]) / n; math.Abs(got-want) > 0.005 {
				t.Errorf("a = %v: share of %d0%% resists = %.4f, want %.4f", c.a, tenths, got, want)
			}
		}
		for tenths := range counts {
			if _, ok := c.want[tenths]; !ok {
				t.Errorf("a = %v: rolled %d0%%, outside the steps around the average", c.a, tenths)
			}
		}
		if mean := float64(total) / n / 10; math.Abs(mean-c.a) > 0.005 {
			t.Errorf("a = %v: mean resist %.4f", c.a, mean)
		}
	}
}

// TestForeverBinaryPartialResists: against 150 resistance a binary spell that lands rolls partial
// resists on its hits and ticks as any other spell does, and one flagged SpellFlagNoPartialResist
// never does. Classic keeps a binary spell's hits whole.
func TestForeverBinaryPartialResists(t *testing.T) {
	partials := func(attacker, defender *Unit, flags SpellFlag, isPeriodic bool) int {
		sim := &Simulation{rand: NewSplitMix(1)}
		spell := shadowSpell(attacker, flags)
		attackTable := NewAttackTable(attacker, defender, nil)
		n := 0
		for range 1000 {
			if multiplier, outcome := spell.ResistanceMultiplier(sim, isPeriodic, attackTable); multiplier != 1 || outcome != OutcomeEmpty {
				n++
			}
		}
		return n
	}

	attacker, defender := foreverResistUnits(150)
	if n := partials(attacker, defender, 0, false); n == 0 {
		t.Errorf("a plain shadow spell was never partially resisted in 1000 hits")
	}
	if n := partials(attacker, defender, SpellFlagBinary, false); n == 0 {
		t.Errorf("a binary spell's hit was never partially resisted in 1000")
	}
	if n := partials(attacker, defender, SpellFlagBinary, true); n == 0 {
		t.Errorf("a binary spell's tick was never partially resisted in 1000")
	}
	if n := partials(attacker, defender, SpellFlagBinary|SpellFlagNoPartialResist, false); n != 0 {
		t.Errorf("a binary spell flagged SpellFlagNoPartialResist was partially resisted %d times in 1000", n)
	}

	classicAttacker := &Unit{Type: PlayerUnit, Level: 60}
	classicDefender := &Unit{Type: EnemyUnit, Level: 63}
	classicDefender.stats[stats.ShadowResistance] = 150
	if n := partials(classicAttacker, classicDefender, SpellFlagBinary, false); n != 0 {
		t.Errorf("under Classic a binary spell was partially resisted %d times in 1000", n)
	}
}

// TestForeverPartialResistOutcome: the outcome carries the resisted tenths and the log reads them.
func TestForeverPartialResistOutcome(t *testing.T) {
	for tenths := 1; tenths <= 10; tenths++ {
		outcome := OutcomeHit | partialResistOutcome(tenths)
		if got := outcome.ResistedTenths(); got != tenths {
			t.Errorf("ResistedTenths of a %d0%% resist = %d", tenths, got)
		}
		if !outcome.Matches(OutcomePartial) || !outcome.Matches(OutcomeHit) {
			t.Errorf("a %d0%% resist outcome lost its bits: %b", tenths, outcome)
		}
	}
	if got := (OutcomeCrit | partialResistOutcome(3)).String(); got != "Crit (30% Resist)" {
		t.Errorf("a 30%% resisted crit reads %q", got)
	}
	if got := (OutcomeHit | OutcomePartial2_4).String(); got != "Hit (50% Resist)" {
		t.Errorf("a Classic 50%% resisted hit reads %q", got)
	}
}

// TestForeverFullResistIsAMiss: a hit the target resists in full is a miss. It lands nothing and
// counts one miss, and expected damage takes the average resist instead of rolling it.
func TestForeverFullResistIsAMiss(t *testing.T) {
	sim, character := setupProcSim(t, proto.Ruleset_RulesetForever, nil)
	target := sim.GetTargetUnit(0)
	spell := character.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 999_002},
		SpellSchool:      SpellSchoolShadow,
		DefenseType:      DefenseTypeMagic,
		ProcMask:         ProcMaskSpellDamage,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ExpectedInitialDamage: func(sim *Simulation, target *Unit, spell *Spell, _ bool) *SpellResult {
			return spell.CalcDamage(sim, target, 100, spell.OutcomeExpectedMagicAlwaysHit)
		},
	})

	target.stats[stats.ShadowResistance] = 1e12
	for range 200 {
		result := spell.CalcDamage(sim, target, 100, spell.OutcomeMagicHitAndCrit)
		if result.Landed() || result.Outcome != OutcomeMiss || result.Damage != 0 {
			t.Fatalf("a fully resisted hit landed %v with outcome %v and %v damage, want a miss with none",
				result.Landed(), result.Outcome, result.Damage)
		}
	}
	metrics := spell.SpellMetrics[target.UnitIndex]
	if metrics.Misses != 200 || metrics.Hits != 0 || metrics.Crits != 0 || metrics.ResistedHits != 0 {
		t.Errorf("200 fully resisted hits counted %d misses, %d hits, %d crits and %d resisted hits, want 200 misses only",
			metrics.Misses, metrics.Hits, metrics.Crits, metrics.ResistedHits)
	}

	target.stats[stats.ShadowResistance] = 100
	want := 100 * spell.AttackerDamageMultiplier(spell.Unit.AttackTables[target.UnitIndex][spell.CastType], false) *
		(1 - 100.0/(100+9*float64(character.Level)))
	if got := spell.ExpectedInitialDamage(sim, target); math.Abs(got-want) > 1e-6 {
		t.Errorf("expected damage against 100 resistance = %v, want %v", got, want)
	}
}
