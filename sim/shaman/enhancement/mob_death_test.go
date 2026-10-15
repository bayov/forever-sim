package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/shaman"
)

// packSim is a naked orc shaman of the given level with a Whirlwind Axe and white swings
// only, in front of n normal mobs of its level, in a health fight: each mob dies when its
// health runs out, and the fight ends when all have.
//
// watch, if set, is called on our Chance of Death aura before the sim starts, to hook what
// we deal and take. The aura is always up in a health fight.
func packSim(t *testing.T, level int32, n int, watch func(watcher *core.Aura)) (*core.Simulation, *EnhancementShaman) {
	mobs := make([]*proto.Target, n)
	for i := range mobs {
		mobs[i] = core.NormalMobTarget(level, proto.MobType_MobTypeBeast, 1)
	}
	req := normalMobRequest(mobs, true, 1)
	req.Raid.Parties[0].Players[0].Level = level
	sim := core.NewSim(req, simsignals.CreateSignals())
	enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
	if watch != nil {
		watcher := enh.GetAura(core.ChanceOfDeathAuraLabel)
		if watcher == nil {
			t.Fatal("no Chance of Death aura in a health fight")
		}
		watch(watcher)
	}
	sim.Reset()
	sim.PrePull()
	return sim, enh
}

// killNow kills the target with one blow from a spell of ours that can't miss.
func killNow(sim *core.Simulation, enh *EnhancementShaman, target *core.Unit) {
	blow := enh.GetOrRegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{OtherID: proto.OtherAction_OtherActionAttack, Tag: 99},
		SpellSchool:      core.SpellSchoolPhysical,
		ProcMask:         core.ProcMaskEmpty,
		DamageMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, 1e6, spell.OutcomeAlwaysHit)
		},
	})
	blow.Cast(sim, target)
}

// TestPackMobsDieOneByOne fights 3 level 30 mobs with white swings. We hit the first one
// until it dies, then the next, and each mob stops swinging at us when it dies. Nothing of
// ours lands on a dead mob, and the fight ends with the last kill.
func TestPackMobsDieOneByOne(t *testing.T) {
	died := map[*core.Unit]time.Duration{}
	sim, _ := packSim(t, 30, 3, func(watcher *core.Aura) {
		taken := watcher.OnSpellHitTaken
		watcher.OnSpellHitTaken = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if _, dead := died[spell.Unit]; dead {
				t.Errorf("at %s %s swung at us, dead since %s", sim.CurrentTime, spell.Unit.Label, died[spell.Unit])
			}
			taken(aura, sim, spell, result)
		}
		watcher.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if _, dead := died[result.Target]; dead {
				t.Errorf("at %s we hit %s, dead since %s", sim.CurrentTime, result.Target.Label, died[result.Target])
			}
		}
	})
	mobs := sim.Encounter.TargetUnits

	var order []*core.Unit
	for !sim.Step() {
		for _, mob := range mobs {
			if _, ok := died[mob]; !ok && mob.IsDead() {
				died[mob] = sim.CurrentTime
				order = append(order, mob)
			}
		}
	}

	if len(order) != 3 || order[0] != mobs[0] || order[1] != mobs[1] || order[2] != mobs[2] {
		t.Fatalf("mobs died in the order %v, want all 3 in order", order)
	}
	if len(sim.Encounter.ActiveTargetUnits) != 0 {
		t.Errorf("%d mobs still alive at the end", len(sim.Encounter.ActiveTargetUnits))
	}
	if last := died[mobs[2]]; sim.CurrentTime != last {
		t.Errorf("the fight ended at %s, the last mob died at %s", sim.CurrentTime, last)
	}
	// Overkill doesn't count: the damage taken is the 3 mobs' health.
	if want := 3 * core.NormalMobStatsAt(30).Health; sim.Encounter.DamageTaken != want {
		t.Errorf("the mobs took %.0f damage, want their %.0f health", sim.Encounter.DamageTaken, want)
	}
}

// TestAoEOnlyHitsLivingMobs kills 2 of 3 mobs and casts Chain Lightning (up to 3 targets)
// and Fire Nova (every target, with a Searing Totem down for it) at level 40. Each hits the
// one mob left, once.
func TestAoEOnlyHitsLivingMobs(t *testing.T) {
	hits := map[core.ActionID]map[*core.Unit]int{}
	sim, enh := packSim(t, 40, 3, func(watcher *core.Aura) {
		watcher.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if hits[spell.ActionID] == nil {
				hits[spell.ActionID] = map[*core.Unit]int{}
			}
			hits[spell.ActionID][result.Target]++
		}
	})
	mobs := sim.Encounter.TargetUnits

	chain := topRank(t, "Chain Lightning", enh.ChainLightning)
	nova := topRank(t, "Fire Nova", enh.FireNova)
	at(sim, 0.5, func(sim *core.Simulation) {
		killNow(sim, enh, mobs[0])
		killNow(sim, enh, mobs[1])
		if enh.CurrentTarget != mobs[2] {
			t.Errorf("our target is %s, want the one alive", enh.CurrentTarget.Label)
		}
	})
	cast := func(spell *core.Spell) func(sim *core.Simulation) {
		return func(sim *core.Simulation) {
			enh.AddMana(sim, enh.MaxMana(), enh.GetManaNotCastingMetrics())
			castNow(t, sim, enh, spell)
		}
	}
	at(sim, 1, cast(chain))
	// Fire Nova needs a fire totem down under Forever.
	at(sim, 7, cast(topRank(t, "Searing Totem", enh.SearingTotem)))
	at(sim, 9, cast(nova))
	checked := false
	at(sim, 10, func(sim *core.Simulation) {
		checked = true
		for _, spell := range []*core.Spell{chain, nova} {
			got := hits[spell.ActionID]
			if len(got) != 1 || got[mobs[2]] != 1 {
				t.Errorf("%s hit %v, want the living mob once", spell.ActionID, got)
			}
		}
	})
	for !sim.Step() && sim.CurrentTime < 11*time.Second {
	}
	if !checked {
		t.Fatalf("the fight ended at %s, before the check", sim.CurrentTime)
	}
}

// TestSearingTotemMovesOn drops a Searing Totem on 2 mobs and kills the first. The totem
// keeps going and shoots the second one.
func TestSearingTotemMovesOn(t *testing.T) {
	var shots []*core.Unit
	sim, enh := packSim(t, 30, 2, func(watcher *core.Aura) {
		watcher.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.SpellCode == shaman.SpellCode_ShamanSearingTotem && !spell.Flags.Matches(shaman.SpellFlagTotem) {
				shots = append(shots, result.Target)
			}
		}
	})
	mobs := sim.Encounter.TargetUnits
	totem := topRank(t, "Searing Totem", enh.SearingTotem)

	at(sim, 0.1, func(sim *core.Simulation) { castNow(t, sim, enh, totem) })
	at(sim, 4, func(sim *core.Simulation) { killNow(sim, enh, mobs[0]) })
	for !sim.Step() && sim.CurrentTime < 12*time.Second {
	}
	var before, after int
	for _, mob := range shots {
		if mob == mobs[0] {
			before++
		} else {
			after++
		}
	}
	if before == 0 || after < 2 {
		t.Errorf("the totem shot the first mob %d times and the second %d times, want both", before, after)
	}
	if !totem.AOEDot().IsActive() {
		t.Error("the totem is down after its target died")
	}
}

// TestCastStopsWhenTargetDies starts a Chain Lightning (2.5 sec) at the first of 2 mobs and
// kills that mob 1 sec in. The cast stops there: no mana spent, no damage and its cooldown
// given back, so we can cast it again at the mob that's left.
func TestCastStopsWhenTargetDies(t *testing.T) {
	hit := 0
	sim, enh := packSim(t, 40, 2, func(watcher *core.Aura) {
		watcher.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.SpellCode == shaman.SpellCode_ShamanChainLightning {
				hit++
			}
		}
	})
	mobs := sim.Encounter.TargetUnits
	chain := topRank(t, "Chain Lightning", enh.ChainLightning)

	var mana float64
	checked := false
	at(sim, 1, func(sim *core.Simulation) {
		enh.AddMana(sim, enh.MaxMana(), enh.GetManaNotCastingMetrics())
		mana = enh.CurrentMana()
		castNow(t, sim, enh, chain)
	})
	at(sim, 2, func(sim *core.Simulation) {
		killNow(sim, enh, mobs[0])
		if enh.IsCasting(sim) {
			t.Errorf("still casting after the target died, until %s", enh.Hardcast.Expires)
		}
		if enh.CurrentMana() < mana {
			t.Errorf("the stopped cast cost %.0f mana", mana-enh.CurrentMana())
		}
		// We can act again when the GCD from the cast's start runs out.
		if got := enh.GCD.ReadyAt(); got != 2500*time.Millisecond {
			t.Errorf("the GCD is ready at %s, want 2.5s", got)
		}
	})
	at(sim, 2.6, func(sim *core.Simulation) {
		checked = true
		if hit != 0 {
			t.Errorf("the stopped cast hit %d times", hit)
		}
		if !chain.CD.IsReady(sim) {
			t.Errorf("Chain Lightning is on cooldown for %s after the stopped cast", chain.CD.TimeToReady(sim))
		}
		enh.AddMana(sim, enh.MaxMana(), enh.GetManaNotCastingMetrics())
		castNow(t, sim, enh, chain)
		if enh.Hardcast.Target != mobs[1] {
			t.Errorf("the new cast is at %v, want the mob that's left", enh.Hardcast.Target)
		}
	})
	for !sim.Step() && sim.CurrentTime < 3*time.Second {
	}
	if !checked {
		t.Fatalf("the fight ended at %s, before the check", sim.CurrentTime)
	}
}
