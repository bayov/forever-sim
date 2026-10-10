package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanJudgementOfWisdom checks Judgement of Wisdom on the boss: 59 mana on a 50%
// roll from each white swing, and a swing that misses or is dodged rolls too (Classic).
//
// The user keeps it at rank 3 (2026-10-08), because we only sim a raid at level 60.
//
// A naked level 60 shaman swings unarmed for 15 min. About 15% of the swings miss or are
// dodged by the level 63 boss. If those didn't roll, Judgement of Wisdom would give mana on
// about 43% of the swings and not 50%.
func TestOrcShamanJudgementOfWisdom(t *testing.T) {
	sim, enh := newShamanSim(newOrcShaman(60, "", &proto.EnhancementShaman_Options{}), &proto.Debuffs{JudgementOfWisdom: true}, 900)
	// We start low, so no proc is lost to a full mana bar.
	at(sim, 0.5, func(sim *core.Simulation) { enh.SpendMana(sim, enh.MaxMana(), enh.GetManaNotCastingMetrics()) })
	runSim(sim)

	swings := enh.AutoAttacks.MHAuto().SpellMetrics[enh.CurrentTarget.UnitIndex]
	missed := swings.Misses + swings.Dodges + swings.Parries
	if swings.Casts < 400 || missed*100 < swings.Casts*10 {
		t.Fatalf("got %d swings and %d misses, want 400 or more with 10%% or more missed", swings.Casts, missed)
	}
	jow := enh.JowManaMetrics
	if jow == nil || jow.Events == 0 {
		t.Fatalf("Judgement of Wisdom never gave mana")
	}
	if jow.Gain != 59*float64(jow.Events) {
		t.Errorf("got %.1f mana from %d procs, want 59 each", jow.Gain, jow.Events)
	}
	if rate := float64(jow.Events) / float64(swings.Casts); rate < 0.465 || rate > 0.535 {
		t.Errorf("Judgement of Wisdom gave mana on %.1f%% of %d swings, want 50%%", rate*100, swings.Casts)
	}
}

// TestOrcShamanWaterShield checks Water Shield's globes against raid damage.
//
// Each globe gives 2% of max mana (the beta client text) and there are 3. The client says
// only one globe activates "every few seconds" and doesn't say how long, so we use 3.5 sec.
// Raid damage comes once every 5 to 15 sec (the user, 2026-10-08), so that wait only shows
// when hits come faster than that. The enemy lands a melee hit on us at set times.
func TestOrcShamanWaterShield(t *testing.T) {
	cases := []struct {
		every float64 // Seconds between the enemy's hits, from the first one.
		// The globes left just after each check time, and how many we spent by then.
		checks []struct {
			at     float64
			stacks int32
		}
	}{
		// A hit every 5 sec (5, 10, 15): every hit spends a globe.
		{5, []struct {
			at     float64
			stacks int32
		}{{4, 3}, {6, 2}, {11, 1}, {16, 0}}},
		// A hit every 3 sec (3, 6, 9, 12, 15): the hits at 6 and 12 come too soon after a
		// globe and do nothing.
		{3, []struct {
			at     float64
			stacks int32
		}{{2, 3}, {4, 2}, {7, 2}, {10, 1}, {13, 1}, {16, 0}}},
	}
	for _, c := range cases {
		// Water Shield.
		sim, enh := newShamanSim(newOrcShaman(60, "--000000001", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 30)
		full := enh.ManaRegenPerSecondWhileNotCasting()
		globe := 0.02 * enh.MaxMana()
		melee := enh.CurrentTarget.RegisterSpell(core.SpellConfig{
			ActionID:         core.ActionID{SpellID: 1},
			SpellSchool:      core.SpellSchoolPhysical,
			DefenseType:      core.DefenseTypeMelee,
			ProcMask:         core.ProcMaskMeleeMHAuto,
			DamageMultiplier: 1,
		})
		for hitAt := c.every; hitAt < 16; hitAt += c.every {
			at(sim, hitAt, func(sim *core.Simulation) { melee.CalcAndDealDamage(sim, &enh.Unit, 1, forcedOutcome(core.OutcomeHit)) })
		}

		// Water Shield is free, so the five second rule never starts and we regen in full.
		var start float64
		at(sim, 0.5, func(sim *core.Simulation) {
			enh.SpendMana(sim, enh.MaxMana()-200, enh.GetManaNotCastingMetrics())
			castNow(t, sim, enh, enh.WaterShield)
			start = enh.CurrentMana()
		})
		for _, check := range c.checks {
			at(sim, check.at, func(sim *core.Simulation) {
				stacks := int32(0)
				if enh.WaterShieldAura.IsActive() {
					stacks = enh.WaterShieldAura.GetStacks()
				}
				if stacks != check.stacks {
					t.Errorf("a hit every %.0f sec, at %s: got %d globes, want %d", c.every, sim.CurrentTime, stacks, check.stacks)
				}
				want := start + (check.at-0.5)*full + float64(3-check.stacks)*globe
				if have := enh.CurrentMana(); math.Abs(have-want) > 1e-6 {
					t.Errorf("a hit every %.0f sec, at %s: got %.3f mana, want %.3f", c.every, sim.CurrentTime, have, want)
				}
			})
		}
		runSim(sim)
	}
}

// TestOrcShamanManaTide checks Mana Tide Totem at level 60: 290 mana every 3 sec, 4 times,
// starting 3 sec after we put it down. It takes the water slot, so Mana Spring Totem stops.
func TestOrcShamanManaTide(t *testing.T) {
	// Mana Tide Totem.
	sim, enh := newShamanSim(newOrcShaman(60, "--000000000001", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 30)
	full := enh.ManaRegenPerSecondWhileNotCasting()
	manaSpring := enh.ManaSpringTotem[4]
	manaSpringAura := enh.GetAura("Mana Spring Totem (Rank 4)")
	manaTide := enh.ManaTideTotem[3]

	checkMana := func(sim *core.Simulation, want float64) {
		if have := enh.CurrentMana(); math.Abs(have-want) > 1e-6 {
			t.Errorf("at %s: got %.3f mana, want %.3f", sim.CurrentTime, have, want)
		}
	}

	var mana float64
	at(sim, 0.5, func(sim *core.Simulation) {
		enh.SpendMana(sim, enh.MaxMana()-200, enh.GetManaNotCastingMetrics())
		mana = enh.CurrentMana()
	})
	at(sim, 1, func(sim *core.Simulation) {
		castNow(t, sim, enh, manaSpring)
		mana += 0.5*full - manaSpring.CurCast.Cost
		checkMana(sim, mana)
	})
	// No Spirit regen during the five second rule, which the totems start (1 and 2 sec).
	at(sim, 2, func(sim *core.Simulation) {
		castNow(t, sim, enh, manaTide)
		mana -= manaTide.CurCast.Cost
		checkMana(sim, mana)
		if manaSpringAura.IsActive() {
			t.Errorf("Mana Spring Totem is still up after Mana Tide Totem")
		}
	})
	// Mana Spring would have ticked at 3 sec.
	at(sim, 4.5, func(sim *core.Simulation) { checkMana(sim, mana) })
	at(sim, 5.5, func(sim *core.Simulation) { checkMana(sim, mana+290) })
	at(sim, 14.5, func(sim *core.Simulation) { checkMana(sim, mana+4*290+7.5*full) })
	at(sim, 18, func(sim *core.Simulation) { checkMana(sim, mana+4*290+11*full) })
	runSim(sim)
}
