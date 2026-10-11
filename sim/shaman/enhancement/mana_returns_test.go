package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
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

// TestOrcShamanWaterShieldRecast checks what our heals do to Water Shield, what a recast does,
// and how long it lasts (shaman_audit.md 7.16).
//
// A heal of ours that crits spends a globe for 2% of max mana, and one that doesn't crit
// leaves it alone. Water Shield has no cooldown, so we cast it again right away and it's back
// to 3 globes. The 3.5 sec wait carries over to the new shield, as Lightning Shield's does on
// the beta (7.15), so an enemy hit 1 sec later spends nothing and one 3.5 sec after the crit
// spends a globe. Then a shield that takes no hits lasts 10 min.
func TestOrcShamanWaterShieldRecast(t *testing.T) {
	// Water Shield.
	sim, enh := newShamanSim(newOrcShaman(60, "--000000001", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 620)
	aura := enh.WaterShieldAura
	heal := enh.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 1},
		SpellSchool:      core.SpellSchoolNature,
		DefenseType:      core.DefenseTypeMagic,
		ProcMask:         core.ProcMaskSpellHealing,
		DamageMultiplier: 1,
	})
	melee := enh.CurrentTarget.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 2},
		SpellSchool:      core.SpellSchoolPhysical,
		DefenseType:      core.DefenseTypeMelee,
		ProcMask:         core.ProcMaskMeleeMHAuto,
		DamageMultiplier: 1,
	})
	globes := func(sim *core.Simulation, label string, want int32) {
		got := int32(0)
		if aura.IsActive() {
			got = aura.GetStacks()
		}
		if got != want {
			t.Errorf("at %s, after %s: %d globes, want %d", sim.CurrentTime, label, got, want)
		}
	}
	hit := func(label string, want int32) func(sim *core.Simulation) {
		return func(sim *core.Simulation) {
			melee.CalcAndDealDamage(sim, &enh.Unit, 1, forcedOutcome(core.OutcomeHit))
			globes(sim, label, want)
		}
	}

	at(sim, 1, func(sim *core.Simulation) {
		castNow(t, sim, enh, enh.WaterShield)
		enh.SpendMana(sim, enh.MaxMana()/2, enh.GetManaNotCastingMetrics())
		heal.CalcAndDealHealing(sim, &enh.Unit, 10, forcedOutcome(core.OutcomeHit))
		globes(sim, "a heal", 3)
		mana := enh.CurrentMana()
		heal.CalcAndDealHealing(sim, &enh.Unit, 10, forcedOutcome(core.OutcomeCrit))
		globes(sim, "a heal crit", 2)
		if got, want := enh.CurrentMana()-mana, 0.02*enh.MaxMana(); math.Abs(got-want) > 1e-6 {
			t.Errorf("the heal crit's globe gave %.3f mana, want %.3f", got, want)
		}
		castNow(t, sim, enh, enh.WaterShield)
		globes(sim, "the recast", 3)
	})
	at(sim, 2, hit("an enemy hit 1 sec after the globe", 3))
	at(sim, 4.5, hit("an enemy hit 3.5 sec after the globe", 2))

	at(sim, 5, func(sim *core.Simulation) { castNow(t, sim, enh, enh.WaterShield) })
	at(sim, 604.9, func(sim *core.Simulation) { globes(sim, "599.9 sec", 3) })
	at(sim, 605.1, func(sim *core.Simulation) { globes(sim, "600.1 sec", 0) })
	runSim(sim)
}

// TestOrcShamanManaSpring checks the mana our own Mana Spring Totem gives at level 30
// (shaman_audit.md 7.26).
//
// Rank 1 gives 4 mana every 2 sec from 2 sec after the drop. On the beta all 26 ticks on us
// were 4 mana, 1.91 to 2.08 sec apart. Restorative Totems adds 5% a point, so 5/5 makes it 5.
// We look at the mana a moment before and after each time a tick could come, so the regen in
// between is under 0.1 mana.
func TestOrcShamanManaSpring(t *testing.T) {
	for _, c := range []struct {
		talents string
		tick    float64
	}{{"", 4}, {"--00000000005", 5}} {
		sim, enh := newShamanSim(newOrcShaman(30, c.talents, &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 10)
		if c.talents != "" && enh.Talents.RestorativeTotems != 5 {
			t.Fatalf("got Restorative Totems %d, want 5", enh.Talents.RestorativeTotems)
		}
		manaSpring := topRank(t, "Mana Spring Totem", enh.ManaSpringTotem)
		if manaSpring.Rank != 1 {
			t.Fatalf("level 30 casts Mana Spring Totem rank %d, want 1", manaSpring.Rank)
		}
		at(sim, 1, func(sim *core.Simulation) {
			enh.SpendMana(sim, enh.CurrentMana()-200, enh.GetManaNotCastingMetrics())
			castNow(t, sim, enh, manaSpring)
		})
		var before float64
		jump := func(seconds, want float64) {
			at(sim, seconds-0.001, func(sim *core.Simulation) { before = enh.CurrentMana() })
			at(sim, seconds+0.001, func(sim *core.Simulation) {
				if got := enh.CurrentMana() - before; math.Abs(got-want) > 0.1 {
					t.Errorf("talents %q, at %.0f sec: mana went up %.2f, want %.0f", c.talents, seconds, got, want)
				}
			})
		}
		jump(2, 0)
		jump(3, c.tick)
		jump(4, 0)
		jump(5, c.tick)
		jump(7, c.tick)
		runSim(sim)
	}
}

// TestOrcShamanManaTide checks Mana Tide Totem at level 60: 290 mana every 3 sec, 4 times,
// starting 3 sec after we put it down. It takes the water slot, so Mana Spring Totem stops.
//
// The totem stands for 13 sec, as in the Forever client (shaman_audit.md 7.1), so the water
// slot stays taken a second after the last tick, with no fifth tick.
func TestOrcShamanManaTide(t *testing.T) {
	// Mana Tide Totem.
	sim, enh := newShamanSim(newOrcShaman(60, "--000000000001", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 30)
	full := enh.ManaRegenPerSecondWhileNotCasting()
	manaSpring := enh.ManaSpringTotem[4]
	manaSpringAura := enh.GetAura("Mana Spring Totem (Rank 4)")
	manaTide := enh.ManaTideTotem[3]
	manaTideAura := enh.GetAura("Mana Tide Totem (Rank 3)")

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
	at(sim, 14.5, func(sim *core.Simulation) {
		checkMana(sim, mana+4*290+7.5*full)
		if !manaTideAura.IsActive() || enh.TotemExpirations[shaman.WaterTotem] != 15*time.Second {
			t.Errorf("at 14.5 sec: Mana Tide Totem is up %v until %s, want up until 15s", manaTideAura.IsActive(), enh.TotemExpirations[shaman.WaterTotem])
		}
	})
	at(sim, 15.5, func(sim *core.Simulation) {
		if manaTideAura.IsActive() {
			t.Errorf("at 15.5 sec: Mana Tide Totem is still up")
		}
	})
	at(sim, 18, func(sim *core.Simulation) { checkMana(sim, mana+4*290+11*full) })
	runSim(sim)
}
