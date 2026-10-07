package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// newFiveSecondRuleSim builds an Orc shaman with Polished Driftwood Icon and 30 bonus Spirit.
func newFiveSecondRuleSim(level int32, talents string) (*core.Simulation, *EnhancementShaman) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotRanged] = &proto.ItemSpec{Id: 249398} // Polished Driftwood Icon

	player := newOrcShaman(level, talents, &proto.EnhancementShaman_Options{})
	player.Equipment = &proto.EquipmentSpec{Items: items}
	player.BonusStats = &proto.UnitStats{Stats: stats.Stats{stats.Spirit: 30}.ToFloatArray()}
	return newShamanSim(player, &proto.Debuffs{}, 60)
}

// newOrcShaman is a naked Orc shaman with no rotation, so only the casts a test schedules
// happen.
func newOrcShaman(level int32, talents string, options *proto.EnhancementShaman_Options) *proto.Player {
	return &proto.Player{
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassShaman,
		Level:         level,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
			Options: options,
		}},
	}
}

// newShamanSim starts a Forever sim of the player against one target, ready for at() and
// runSim.
func newShamanSim(player *proto.Player, debuffs *proto.Debuffs, seconds float64) (*core.Simulation, *EnhancementShaman) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, debuffs),
		Encounter:  &proto.Encounter{Duration: seconds, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
}

func topRank(t *testing.T, name string, ranks []*core.Spell) *core.Spell {
	for i := len(ranks) - 1; i >= 0; i-- {
		if ranks[i] != nil {
			return ranks[i]
		}
	}
	t.Fatalf("no %s rank", name)
	return nil
}

// at runs fn at the given second of the fight. Mana is settled up to then before fn runs.
func at(sim *core.Simulation, seconds float64, fn func(sim *core.Simulation)) {
	sim.AddPendingAction(&core.PendingAction{NextActionAt: core.DurationFromSeconds(seconds), OnAction: fn})
}

// castNow casts spell at the current time.
//
// An empty rotation keeps the GCD busy in 50 ms steps, so we free it first.
func castNow(t *testing.T, sim *core.Simulation, enh *EnhancementShaman, spell *core.Spell) {
	enh.SetGCDTimer(sim, sim.CurrentTime)
	if !spell.Cast(sim, enh.CurrentTarget) {
		t.Fatalf("at %s: %s didn't cast", sim.CurrentTime, spell.ActionID)
	}
}

func runSim(sim *core.Simulation) {
	for !sim.Step() {
	}
}

// TestOrcShamanFiveSecondRule checks what starts the five second rule and what we regen
// during it under Forever.
//
// It works like the other Classic versions (the user, 2026-10-08, and cmangos
// Spell::TakePower). Spending mana starts it. A spell with a cast time pays at the end of
// the cast, so the five seconds count from there and we regen in full during the cast. A
// cast that costs nothing (Clearcasting here) starts nothing.
//
// The character regens 16.625 mana a second and keeps 8% of it (1.33) while casting
// through Polished Driftwood Icon (TestOrcShamanDriftwoodIcon).
func TestOrcShamanFiveSecondRule(t *testing.T) {
	// Elemental Focus, for a real free cast.
	sim, enh := newFiveSecondRuleSim(30, "0000001")
	const full, casting = 16.625, 1.33
	if got := enh.ManaRegenPerSecondWhileNotCasting(); math.Abs(got-full) > 1e-9 {
		t.Fatalf("got %.4f mana a second, want %.4f", got, full)
	}
	if got := enh.ManaRegenPerSecondWhileCasting(); math.Abs(got-casting) > 1e-9 {
		t.Fatalf("got %.4f mana a second while casting, want %.4f", got, casting)
	}

	flameShock := topRank(t, "Flame Shock", enh.FlameShock)
	lightningBolt := topRank(t, "Lightning Bolt", enh.LightningBolt)
	castTime := lightningBolt.DefaultCast.CastTime

	checkMana := func(sim *core.Simulation, want float64) {
		if have := enh.CurrentMana(); math.Abs(have-want) > 1e-6 {
			t.Errorf("at %s: got %.3f mana, want %.3f", sim.CurrentTime, have, want)
		}
	}
	checkFSR := func(sim *core.Simulation, want time.Duration) {
		if have := enh.PseudoStats.FiveSecondRuleRefreshTime; have != want {
			t.Errorf("at %s: five second rule runs until %s, want %s", sim.CurrentTime, have, want)
		}
	}
	// Elemental Focus can proc Clearcasting off any of our casts. We drop it, so that only
	// the cast at 30 sec is free.
	cast := func(sim *core.Simulation, spell *core.Spell) {
		castNow(t, sim, enh, spell)
		enh.ClearcastingAura.Deactivate(sim)
	}

	// Spending mana without a spell (here to get below full) doesn't start the rule.
	var mana float64
	at(sim, 0.5, func(sim *core.Simulation) {
		enh.SpendMana(sim, 800, enh.GetManaNotCastingMetrics())
		mana = enh.CurrentMana()
		checkFSR(sim, 0)
	})

	// An instant starts it when we cast it.
	at(sim, 1, func(sim *core.Simulation) {
		mana += 0.5 * full
		checkMana(sim, mana)
		cast(sim, flameShock)
		if flameShock.CurCast.Cost <= 0 {
			t.Fatalf("Flame Shock costs %.1f mana", flameShock.CurCast.Cost)
		}
		mana -= flameShock.CurCast.Cost
		checkMana(sim, mana)
		checkFSR(sim, 6*time.Second)
	})
	at(sim, 4, func(sim *core.Simulation) { checkMana(sim, mana+3*casting) })
	at(sim, 9, func(sim *core.Simulation) { checkMana(sim, mana+5*casting+3*full) })

	// A cast with a cast time starts it when the cast ends, and we regen in full until then.
	castEnd := 10*time.Second + castTime
	at(sim, 10, func(sim *core.Simulation) {
		mana += 5*casting + 4*full
		checkMana(sim, mana)
		cast(sim, lightningBolt)
		if enh.Hardcast.Expires != castEnd {
			t.Fatalf("Lightning Bolt cast ends at %s, want %s", enh.Hardcast.Expires, castEnd)
		}
		checkMana(sim, mana)
		checkFSR(sim, 6*time.Second)
	})
	at(sim, (castEnd - 250*time.Millisecond).Seconds(), func(sim *core.Simulation) {
		checkMana(sim, mana+(castTime-250*time.Millisecond).Seconds()*full)
	})
	at(sim, (castEnd + 2*time.Second).Seconds(), func(sim *core.Simulation) {
		if lightningBolt.CurCast.Cost <= 0 {
			t.Fatalf("Lightning Bolt costs %.1f mana", lightningBolt.CurCast.Cost)
		}
		mana += castTime.Seconds()*full - lightningBolt.CurCast.Cost
		checkMana(sim, mana+2*casting)
		checkFSR(sim, castEnd+5*time.Second)
	})
	at(sim, (castEnd + 7*time.Second).Seconds(), func(sim *core.Simulation) {
		checkMana(sim, mana+5*casting+2*full)
	})

	// A free cast doesn't start it.
	var manaAt30 float64
	at(sim, 30, func(sim *core.Simulation) {
		enh.ClearcastingAura.Activate(sim)
		manaAt30 = enh.CurrentMana()
		castNow(t, sim, enh, flameShock)
		checkMana(sim, manaAt30)
		checkFSR(sim, castEnd+5*time.Second)
	})
	at(sim, 32, func(sim *core.Simulation) { checkMana(sim, manaAt30+2*full) })

	runSim(sim)
}

// TestOrcShamanCastingRegenCap checks that Improved Stormstrike changes the casting regen
// right away, and that the share of Spirit regen we keep while casting stops at 100%.
//
// A level 60 shaman with Mindfulness 3/3 (50%) and Polished Driftwood Icon (8%) keeps 58%.
// Improved Stormstrike adds 50% for 15 sec, which would make 108%. cmangos caps it at 100%,
// so we regen the same as when not casting.
func TestOrcShamanCastingRegenCap(t *testing.T) {
	// Stormstrike, Improved Stormstrike 2/2 and Mindfulness 3/3.
	sim, enh := newFiveSecondRuleSim(60, "-000000000000100200-003")
	full := enh.ManaRegenPerSecondWhileNotCasting()
	casting := 0.58 * full
	if got := enh.ManaRegenPerSecondWhileCasting(); math.Abs(got-casting) > 1e-9 {
		t.Fatalf("got %.4f mana a second while casting, want %.4f", got, casting)
	}

	impStormstrike := enh.GetAura("Improved Stormstrike")
	checkMana := func(sim *core.Simulation, want float64) {
		if have := enh.CurrentMana(); math.Abs(have-want) > 1e-6 {
			t.Errorf("at %s: got %.3f mana, want %.3f", sim.CurrentTime, have, want)
		}
	}

	var mana float64
	at(sim, 0.5, func(sim *core.Simulation) {
		enh.SpendMana(sim, 1000, enh.GetManaNotCastingMetrics())
	})
	at(sim, 1, func(sim *core.Simulation) {
		castNow(t, sim, enh, topRank(t, "Flame Shock", enh.FlameShock))
		mana = enh.CurrentMana()
	})
	at(sim, 2, func(sim *core.Simulation) {
		mana += casting
		checkMana(sim, mana)
		impStormstrike.Activate(sim)
		if got := enh.ManaRegenPerSecondWhileCasting(); math.Abs(got-full) > 1e-9 {
			t.Errorf("got %.4f mana a second while casting with Improved Stormstrike, want %.4f", got, full)
		}
	})
	at(sim, 4, func(sim *core.Simulation) {
		mana += 2 * full
		checkMana(sim, mana)
		impStormstrike.Deactivate(sim)
	})
	at(sim, 5, func(sim *core.Simulation) { checkMana(sim, mana+casting) })

	runSim(sim)
}
