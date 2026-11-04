package tankwarrior

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/warrior"
)

// The Fury talents a tank warrior reaches, and the Heroic Strike and Shield Block values the audit
// (tank_mechanics-20261104-11) found untested: each one checked against its value, so a change to
// the mechanic fails here.

// testHealth is a health source for setting the warrior's health in a test.
func testHealth(war *warrior.Warrior) *core.ResourceMetrics {
	return war.NewHealthMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionHealingModel})
}

// landedSince reports whether a spell's last ApplyEffects landed (a hit, crit, block or glance).
func landedSince(before, after core.SpellMetrics) bool {
	return after.Hits+after.Crits+after.Blocks+after.Glances > before.Hits+before.Crits+before.Blocks+before.Glances
}

// TestForeverTankUnbridledWrath checks Unbridled Wrath (Fury, 12322 / energize 12964): 12% a point for
// 1 Rage on a landed white swing (client 70338 curve 12 to 60), and nothing on Heroic Strike or
// Cleave, which take a swing's place but are melee specials (12322's ProcTypeMask 4, auto attacks
// only; tank_mechanics-20261104-6).
func TestForeverTankUnbridledWrath(t *testing.T) {
	uwID := &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 12964}}

	// White swings over 20 fights: 60% of the landed ones at 5/5, 1 Rage each.
	m, _ := tankedRun(t, "-0005")
	var landed int32
	for _, a := range m.Actions {
		if a.Id.GetOtherId() != proto.OtherAction_OtherActionAttack {
			continue
		}
		for _, tm := range a.Targets {
			landed += tm.Hits + tm.Crits + tm.Glances + tm.Blocks + tm.BlockedCrits + tm.Crushes
		}
	}
	events, gain := rageEvents(m, uwID)
	if landed < 1000 {
		t.Fatalf("only %d white swings landed", landed)
	}
	if gain != float64(events) {
		t.Errorf("Unbridled Wrath gave %.1f Rage over %d procs, want 1 a proc", gain, events)
	}
	p := float64(events) / float64(landed)
	if se := math.Sqrt(0.6 * 0.4 / float64(landed)); math.Abs(p-0.6) > 4*se {
		t.Errorf("Unbridled Wrath 5/5 procs on %.3f of %d landed white swings, want 0.60", p, landed)
	}
	if m, _ := tankedRun(t, "-0001"); m != nil {
		events, _ := rageEvents(m, uwID)
		if p := float64(events) / float64(landed); math.Abs(p-0.12) > 0.03 {
			t.Errorf("Unbridled Wrath 1/5 procs on about %.3f of the landed white swings, want 0.12", p)
		}
	}

	// Heroic Strike and Cleave: a landed one gives the same Rage (none) with and without the talent.
	for _, talents := range []string{"-0005", ""} {
		sim, war := newTank(t, talents, bulwark)
		war.CurrentTarget.AutoAttacks.CancelAutoSwing(sim)
		target := war.CurrentTarget
		for _, spell := range []*core.Spell{war.HeroicStrike.Spell, war.Cleave.Spell} {
			metrics := &spell.SpellMetrics[target.UnitIndex]
			hits := 0
			for i := 0; i < 400; i++ {
				war.SpendRage(sim, war.CurrentRage(), war.RageRefundMetrics)
				war.AddRage(sim, 50, war.RageRefundMetrics)
				before := *metrics
				spell.ApplyEffects(sim, target, spell)
				if !landedSince(before, *metrics) {
					continue
				}
				hits++
				if got := war.CurrentRage(); got != 50 {
					t.Fatalf("a landed %v with talents %q left %.1f Rage of 50: Unbridled Wrath procs on it", spell.ActionID, talents, got)
				}
			}
			if hits < 200 {
				t.Errorf("%v landed %d times of 400", spell.ActionID, hits)
			}
		}
	}
}

// TestForeverTankEnrage checks Enrage's buff (12880): 2% more Physical damage a point for 12 sec,
// nothing to other schools (its 30% chance on any damaging hit taken is in TestForeverTankProcTalents).
func TestForeverTankEnrage(t *testing.T) {
	for points := 1; points <= 5; points++ {
		talents := "-0000000" + string(rune('0'+points))
		sim, war := newTank(t, talents, bulwark)
		physical := war.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical]
		fire := war.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexFire]
		war.EnrageAura.Activate(sim)
		near(t, "Physical damage under Enrage with "+talents,
			war.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical]/physical, 1+0.02*float64(points), 1e-9)
		near(t, "Fire damage under Enrage", war.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexFire], fire, 1e-9)
		if war.EnrageAura.Duration != 12*time.Second {
			t.Errorf("Enrage lasts %s, want 12s", war.EnrageAura.Duration)
		}
		advance(sim, 13*time.Second)
		near(t, "Physical damage after Enrage", war.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical], physical, 1e-9)
	}
}

// TestForeverTankBloodCraze checks Blood Craze (16487, heal 16488/16490/16491): 1% of max health a
// point over 6 sec in 3 ticks, after a crit taken or a single hit of more than 20% of max health
// (client 70707), and not after a smaller hit; its healing makes no threat (the November 2 notes).
func TestForeverTankBloodCraze(t *testing.T) {
	for points := 1; points <= 3; points++ {
		sim, war := newForeverTankOf(t, "-000000"+string(rune('0'+points)), true)
		boss := war.CurrentTarget
		boss.AutoAttacks.CancelAutoSwing(sim)
		war.AutoAttacks.CancelAutoSwing(sim)
		swing := boss.AutoAttacks.MHAuto()
		maxHealth := war.MaxHealth()

		// A small hit: no heal.
		war.RemoveHealth(sim, 0.5*maxHealth)
		swing.CalcAndDealDamage(sim, &war.Unit, 0.02*maxHealth, swing.OutcomeAlwaysHit)
		health := war.CurrentHealth()
		advance(sim, 7*time.Second)
		near(t, "health healed after a small hit", war.CurrentHealth()-health, 0, 1e-6)

		// A hit of more than 20% of max health: 3 ticks of a third each.
		result := swing.CalcAndDealDamage(sim, &war.Unit, 0.6*maxHealth, swing.OutcomeAlwaysHit)
		if result.Damage <= 0.2*maxHealth {
			t.Fatalf("the big hit dealt %.0f of %.0f health", result.Damage, maxHealth)
		}
		war.RemoveHealth(sim, 0.5*maxHealth) // room for the heal (a fight with no healing model keeps the health full)
		health = war.CurrentHealth()
		advance(sim, 3*time.Second) // one tick at 2 sec
		near(t, "Blood Craze's first tick", war.CurrentHealth()-health, 0.01*float64(points)*maxHealth/3, 1e-6)
		advance(sim, 4*time.Second)
		near(t, "Blood Craze's heal", war.CurrentHealth()-health, 0.01*float64(points)*maxHealth, 1e-6)

		// A crit, however small.
		war.GainHealth(sim, maxHealth, testHealth(war))
		war.RemoveHealth(sim, 0.5*maxHealth)
		for i := 0; i < 2000; i++ {
			if swing.CalcAndDealDamage(sim, &war.Unit, 1, swing.OutcomeEnemyMeleeWhite).DidCrit() {
				break
			}
		}
		health = war.CurrentHealth()
		advance(sim, 7*time.Second)
		near(t, "Blood Craze's heal after a crit", war.CurrentHealth()-health, 0.01*float64(points)*maxHealth, 1e-6)

		hot := war.GetSpell(core.ActionID{SpellID: []int32{0, 16488, 16490, 16491}[points]})
		if hot == nil || hot.ThreatMultiplier != 0 {
			t.Errorf("Blood Craze %d/3's heal is missing or makes threat", points)
		}
	}
}

// TestForeverTankHeroicStrikeDamage checks Heroic Strike rank 8 (11567): the main-hand weapon's
// damage plus 138 (the sim's value; the client's rank 8 adds 138 too).
func TestForeverTankHeroicStrikeDamage(t *testing.T) {
	sim, war := newTank(t, "", bulwark)
	target := war.CurrentTarget
	target.AutoAttacks.CancelAutoSwing(sim)
	target.AddStatDynamic(sim, stats.Armor, -target.GetStat(stats.Armor))
	war.AddStatDynamic(sim, stats.AttackPower, 1000)
	spell := war.HeroicStrike.Spell
	spell.BonusCritRating -= 200 * core.CritRatingPerCritChance
	at := war.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand]
	m := spell.AttackerDamageMultiplier(at, false)
	mh := war.AutoAttacks.MH()
	ap := (mh.SwingSpeed * spell.MeleeAttackPower(target)) / mh.AttackPowerPerDPS
	wantLow, wantHigh := (138+mh.BaseDamageMin+ap)*m, (138+mh.BaseDamageMax+ap)*m

	low, high := math.Inf(1), math.Inf(-1)
	metrics := &spell.SpellMetrics[target.UnitIndex]
	for i := 0; i < 800; i++ {
		war.AddRage(sim, 100, war.RageRefundMetrics)
		before := *metrics
		spell.ApplyEffects(sim, target, spell)
		if metrics.Hits == before.Hits || metrics.Crits != before.Crits || metrics.Blocks != before.Blocks {
			continue
		}
		damage := metrics.TotalDamage - before.TotalDamage
		low, high = min(low, damage), max(high, damage)
	}
	// Damage lands in whole points, and 800 rolls come within 2% of the ends.
	slack := 1 + 0.02*(wantHigh-wantLow)
	if low < wantLow-1 || low > wantLow+slack || high > wantHigh+1 || high < wantHigh-slack {
		t.Errorf("Heroic Strike at %.0f attack power hits for %.1f to %.1f, want %.1f to %.1f (weapon + 138)",
			spell.MeleeAttackPower(target), low, high, wantLow, wantHigh)
	}
}

// TestForeverTankShieldBlockCharges checks that Shield Block's 2 charges (client 70425, 2565) go
// one a block the boss's swings actually make, and that the 75% block goes with the last one.
func TestForeverTankShieldBlockCharges(t *testing.T) {
	sim, war := newForeverTankOf(t, P1Talents, true)
	boss := war.CurrentTarget
	boss.AutoAttacks.CancelAutoSwing(sim)
	war.AutoAttacks.CancelAutoSwing(sim)
	block := war.GetStat(stats.Block)
	ready(sim, war, war.ShieldBlock.Spell)
	if !war.ShieldBlock.Cast(sim, boss) {
		t.Fatal("Shield Block didn't cast")
	}
	aura := war.ShieldBlockAura
	if aura.GetStacks() != 2 || aura.Duration != 7*time.Second {
		t.Fatalf("Shield Block has %d charges for %s, want 2 for 7s", aura.GetStacks(), aura.Duration)
	}
	swing := boss.AutoAttacks.MHAuto()
	blocks, unblocked := 0, 0
	for i := 0; i < 200 && aura.IsActive(); i++ {
		war.GainHealth(sim, war.MaxHealth(), testHealth(war))
		stacks := aura.GetStacks()
		result := swing.CalcAndDealDamage(sim, &war.Unit, 100, swing.OutcomeEnemyMeleeWhite)
		switch {
		case result.DidBlock():
			blocks++
			if aura.IsActive() && aura.GetStacks() != stacks-1 {
				t.Errorf("a block left %d charges of %d", aura.GetStacks(), stacks)
			}
		case aura.GetStacks() != stacks:
			t.Errorf("a %v took a charge", result.Outcome)
		default:
			unblocked++
		}
	}
	if blocks != 2 || aura.IsActive() {
		t.Errorf("Shield Block went after %d blocks (active: %v), want 2", blocks, aura.IsActive())
	}
	near(t, "block rating after the last charge", war.GetStat(stats.Block), block, 1e-9)
}
