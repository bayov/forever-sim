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

// TestForeverAverageResist: R / (|R| + 8L + 19) for the caster's level (EV-0091), a tenth of it on
// any tick, a channel's included, but not on a spell flagged SpellFlagFullResistanceTicks (EV-0142),
// and nothing for the target's level.
func TestForeverAverageResist(t *testing.T) {
	const k = 8*60 + 19
	for _, c := range []struct {
		name       string
		resistance float64
		tick       bool
		flags      SpellFlag
		want       float64
	}{
		{"no resistance", 0, false, 0, 0},
		{"no resistance, tick", 0, true, 0, 0},
		{"100 resistance", 100, false, 0, 100.0 / (100 + k)},
		{"100 resistance, binary", 100, false, SpellFlagBinary, 100.0 / (100 + k)},
		{"100 resistance, tick", 100, true, 0, 100.0 / (100 + k) / 10},
		{"100 resistance, pure DoT tick", 100, true, SpellFlagPureDot, 100.0 / (100 + k) / 10},
		{"100 resistance, channel tick", 100, true, SpellFlagChanneled, 100.0 / (100 + k) / 10},
		{"100 resistance, full resistance tick", 100, true, SpellFlagFullResistanceTicks, 100.0 / (100 + k)},
		{"100 resistance, full resistance channel tick", 100, true, SpellFlagChanneled | SpellFlagFullResistanceTicks, 100.0 / (100 + k)},
	} {
		attacker, defender := foreverResistUnits(c.resistance)
		if got := defender.foreverAverageResist(shadowSpell(attacker, c.flags), attacker, c.tick); math.Abs(got-c.want) > 1e-12 {
			t.Errorf("%s: average resist %v, want %v", c.name, got, c.want)
		}
	}

	attacker, defender := foreverResistUnits(100)
	if got, want := defender.binaryHitChance(shadowSpell(attacker, SpellFlagBinary), attacker), 1-100.0/(100+k); math.Abs(got-want) > 1e-12 {
		t.Errorf("binary hit chance against 100 resistance = %v, want %v", got, want)
	}

	holy := &Spell{Unit: attacker, SpellSchool: SpellSchoolHoly, SchoolIndex: stats.SchoolIndexHoly}
	attacker.stats[stats.SpellPenetration] = 50
	if got := defender.foreverAverageResist(holy, attacker, false); got != 0 {
		t.Errorf("a Holy spell with 50 Spell Penetration: average resist %v, want 0", got)
	}
	// Searing and Magma Totem's attacks take none of the caster's Spell Piercing (EV-0142).
	if got, want := defender.foreverAverageResist(shadowSpell(attacker, SpellFlagNoSpellPiercing), attacker, false), 100.0/(100+k); math.Abs(got-want) > 1e-12 {
		t.Errorf("a spell flagged SpellFlagNoSpellPiercing with 50 Spell Penetration: average resist %v, want %v", got, want)
	}
}

// TestForeverSpellPenetrationAddsDamage: Spell Penetration beyond the target's resistance makes the
// average negative, and the roll adds 10% or 20% damage, never a resisted outcome. A target
// resistance below 0 counts as 0. A share under 1% gives nothing, and neither does a tick at a
// tenth, while a tick at the full share takes it whole (EV-0142).
func TestForeverSpellPenetrationAddsDamage(t *testing.T) {
	attacker, defender := foreverResistUnits(0)
	attacker.stats[stats.SpellPenetration] = 15
	spell := shadowSpell(attacker, 0)

	a := 15.0 / (15 + 8*60 + 19)
	if got := defender.foreverAverageResist(spell, attacker, false); math.Abs(got+a) > 1e-12 {
		t.Errorf("15 Spell Penetration against no resistance: average resist %v, want %v", got, -a)
	}
	_, debuffed := foreverResistUnits(-75)
	if got := debuffed.foreverAverageResist(spell, attacker, false); math.Abs(got+a) > 1e-12 {
		t.Errorf("15 Spell Penetration against -75 resistance: average resist %v, want %v", got, -a)
	}
	if got := defender.foreverAverageResist(spell, attacker, true); got != 0 {
		t.Errorf("15 Spell Penetration on a tick: average resist %v, want 0", got)
	}
	if got := defender.foreverAverageResist(shadowSpell(attacker, SpellFlagFullResistanceTicks), attacker, true); math.Abs(got+a) > 1e-12 {
		t.Errorf("15 Spell Penetration on a full share tick: average resist %v, want %v", got, -a)
	}
	// 5 piercing at level 60 is a share of 0.99%, under the cut; 6 is 1.19%.
	attacker.stats[stats.SpellPenetration] = 5
	if got := defender.foreverAverageResist(spell, attacker, false); got != 0 {
		t.Errorf("5 Spell Penetration at level 60: average resist %v, want 0", got)
	}
	attacker.stats[stats.SpellPenetration] = 6
	if got, want := defender.foreverAverageResist(spell, attacker, false), -6.0/(6+8*60+19); math.Abs(got-want) > 1e-12 {
		t.Errorf("6 Spell Penetration at level 60: average resist %v, want %v", got, want)
	}
	attacker.stats[stats.SpellPenetration] = 15

	attackTable := NewAttackTable(attacker, defender, nil)
	const n = 100_000
	for _, flags := range []SpellFlag{0, SpellFlagBinary} {
		spell := shadowSpell(attacker, flags)
		sim := &Simulation{rand: NewSplitMix(1)}
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
				t.Errorf("flags %b: share of hits at x%v = %.4f, want %.4f", flags, multiplier, got, want)
			}
		}
		if len(counts) != 3 {
			t.Errorf("flags %b: multipliers rolled: %v, want only 1, 1.1 and 1.2", flags, counts)
		}
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
		chances := resistTenthsChances(c.a)
		for range n {
			tenths := rollResistTenths(sim, chances)
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

// TestForeverResistStopsAt75: no partial resist takes more than 75% of a hit; the roll's 80, 90 and
// 100% steps take 75% and log as a 75% resist (EV-0142).
func TestForeverResistStopsAt75(t *testing.T) {
	attacker, defender := foreverResistUnits(4000) // a = 0.89
	spell := shadowSpell(attacker, 0)
	attackTable := NewAttackTable(attacker, defender, nil)
	sim := &Simulation{rand: NewSplitMix(3)}
	capped := 0
	for range 10_000 {
		multiplier, outcome := spell.ResistanceMultiplier(sim, false, attackTable)
		if multiplier < 0.25-1e-12 {
			t.Fatalf("a hit kept %v of its damage, under 25%%", multiplier)
		}
		if multiplier == 0.25 {
			capped++
			if outcome != OutcomePartial3_4 || (OutcomeHit|outcome).String() != "Hit (75% Resist)" {
				t.Fatalf("a 75%% resist has outcome %v", outcome)
			}
		}
	}
	if capped == 0 {
		t.Errorf("no hit at a share of 0.89 lost 75%%")
	}

	spell.expecting = true
	a := defender.foreverAverageResist(spell, attacker, false)
	chances := resistTenthsChances(a)
	want := 1 - (0.6*chances[6] + 0.7*chances[7] + 0.75*(chances[8]+chances[9]+chances[10]))
	if got, _ := spell.ResistanceMultiplier(sim, false, attackTable); math.Abs(got-want) > 1e-12 {
		t.Errorf("expected multiplier at a share of %.3f = %v, want %v", a, got, want)
	}
}

// TestForeverBinaryPartialResists: against 150 resistance a binary spell that lands takes no partial
// resist on its hits or ticks (EV-0142), a plain spell does, and one flagged SpellFlagNoPartialResist
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
	if n := partials(attacker, defender, 0, true); n == 0 {
		t.Errorf("a plain shadow spell's tick was never partially resisted in 1000")
	}
	if n := partials(attacker, defender, SpellFlagBinary, false); n != 0 {
		t.Errorf("a binary spell's hit was partially resisted %d times in 1000", n)
	}
	if n := partials(attacker, defender, SpellFlagBinary, true); n != 0 {
		t.Errorf("a binary spell's tick was partially resisted %d times in 1000", n)
	}
	if n := partials(attacker, defender, SpellFlagNoPartialResist, false); n != 0 {
		t.Errorf("a spell flagged SpellFlagNoPartialResist was partially resisted %d times in 1000", n)
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

// TestForeverNoFullResistOnPlainSpells: a spell that is not binary never resists whole: against
// any resistance a hit lands and loses at most 75% (EV-0142: Arcane Explosion RESIST 0 of 16,264). A
// binary spell's hit roll resists it whole at the average share, which is a miss that lands nothing.
// Expected damage takes the average resist instead of rolling it.
func TestForeverNoFullResistOnPlainSpells(t *testing.T) {
	sim, character := setupProcSim(t, proto.Ruleset_RulesetForever, nil)
	target := sim.GetTargetUnit(0)
	newSpell := func(id int32, flags SpellFlag) *Spell {
		return character.RegisterSpell(SpellConfig{
			ActionID:         ActionID{SpellID: id},
			SpellSchool:      SpellSchoolShadow,
			DefenseType:      DefenseTypeMagic,
			ProcMask:         ProcMaskSpellDamage,
			Flags:            flags,
			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			ExpectedInitialDamage: func(sim *Simulation, target *Unit, spell *Spell, _ bool) *SpellResult {
				return spell.CalcDamage(sim, target, 100, spell.OutcomeExpectedMagicAlwaysHit)
			},
		})
	}
	spell := newSpell(999_002, 0)
	binary := newSpell(999_003, SpellFlagBinary)

	target.stats[stats.ShadowResistance] = 1e12
	for range 200 {
		result := spell.CalcDamage(sim, target, 100, spell.OutcomeMagicHitAndCrit)
		if result.Landed() && result.ResistanceMultiplier != 0.25 {
			t.Fatalf("a hit against 1e12 resistance kept %v of its damage, want 0.25", result.ResistanceMultiplier)
		}
		if result := binary.CalcDamage(sim, target, 100, binary.OutcomeMagicHitAndCrit); result.Landed() || result.Damage != 0 {
			t.Fatalf("a binary hit against 1e12 resistance landed with %v damage", result.Damage)
		}
	}
	if metrics := spell.SpellMetrics()[target.UnitIndex]; metrics.Hits+metrics.Crits < 190 {
		t.Errorf("a plain spell landed %d of 200 hits against 1e12 resistance", metrics.Hits+metrics.Crits)
	}

	target.stats[stats.ShadowResistance] = 100
	want := 100 * spell.AttackerDamageMultiplier(spell.Unit.AttackTables[target.UnitIndex][spell.CastType], false) *
		(1 - 100.0/(100+8*float64(character.Level)+19))
	if got := spell.ExpectedInitialDamage(sim, target); math.Abs(got-want) > 1e-6 {
		t.Errorf("expected damage against 100 resistance = %v, want %v", got, want)
	}
}
