package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	googleProto "google.golang.org/protobuf/proto"
)

// newPvPSim starts a level 30 Orc with Whirlwind Axe and Stormstrike, in front of a level 30
// enemy player in PvP mode.
//
// dodge, parry and block are the enemy's own chances in percent, where all zero keeps the
// level based ones. downtime is the share of the fight we spend out of melee range.
func newPvPSim(dodge, parry, block, downtime, seconds float64) (*core.Simulation, *EnhancementShaman) {
	player := newOrcShaman(30, stormstrike, &proto.EnhancementShaman_Options{})
	player.InFrontOfTarget = true
	for range proto.ItemSlot_ItemSlotMainHand {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
	}
	player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: whirlwind})
	target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
	target.Level = 30
	target.Stats[proto.Stat_StatDodge] = dodge
	target.Stats[proto.Stat_StatParry] = parry
	target.Stats[proto.Stat_StatBlock] = block
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter: &proto.Encounter{
			Duration:         seconds,
			Pvp:              true,
			PvpMeleeDowntime: downtime,
			Targets:          []*proto.Target{target},
		},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
}

// TestOrcShamanPvPTable checks our white hits against an enemy player, from in front.
//
// White hits never glance in PvP mode, because the enemy isn't a mob. When the enemy has its
// own dodge, parry or block (a rogue's 14% dodge, a shield's 12% block), those replace the
// level based 5% each. The miss chance is the same as against a mob of our level, 5% with a
// two hander (shaman_audit.md 6.7).
func TestOrcShamanPvPTable(t *testing.T) {
	cases := []struct {
		name                     string
		dodge, parry, block      float64
		wantDodge, wantParry, wB float64
	}{
		{"the enemy's own chances", 14, 5, 12, 0.14, 0.05, 0.12},
		{"none set", 0, 0, 0, 0.05, 0.05, 0.05},
	}
	for _, c := range cases {
		sim, enh := newPvPSim(c.dodge, c.parry, c.block, 0, 10)
		white := enh.AutoAttacks.MHAuto()
		const rolls = 20000
		counts := map[core.HitOutcome]int{}
		for range rolls {
			result := white.CalcDamage(sim, enh.CurrentTarget, 100, white.OutcomeMeleeWhite)
			for _, outcome := range []core.HitOutcome{core.OutcomeMiss, core.OutcomeDodge, core.OutcomeParry, core.OutcomeBlock, core.OutcomeGlance} {
				if result.Outcome.Matches(outcome) {
					counts[outcome]++
				}
			}
		}
		for _, want := range []struct {
			name    string
			outcome core.HitOutcome
			chance  float64
		}{
			{"miss", core.OutcomeMiss, 0.05},
			{"dodge", core.OutcomeDodge, c.wantDodge},
			{"parry", core.OutcomeParry, c.wantParry},
			{"block", core.OutcomeBlock, c.wB},
			{"glance", core.OutcomeGlance, 0},
		} {
			rate := float64(counts[want.outcome]) / rolls
			if se := math.Sqrt(want.chance * (1 - want.chance) / rolls); math.Abs(rate-want.chance) > max(4*se, 1e-9) {
				t.Errorf("%s: %.4f %s, want %.4f", c.name, rate, want.name, want.chance)
			}
		}
	}
}

// TestOrcShamanMeleeDowntime checks the time we spend out of melee range in PvP.
//
// With half the fight out of range, we look every 0.1 sec for an hour. We should be out of
// range half the time. While out of range our white swings stop, and so do Stormstrike and
// Fire Nova (the enemy has left our fire totem), but Earth Shock still reaches. Back in range
// all three can be cast.
func TestOrcShamanMeleeDowntime(t *testing.T) {
	const seconds = 3600
	sim, enh := newPvPSim(0, 0, 0, 0.5, seconds)
	white := enh.AutoAttacks.MHAuto()
	searing := topRank(t, "Searing Totem", enh.SearingTotem)
	spells := []struct {
		name  string
		spell *core.Spell
		melee bool
	}{
		{"Stormstrike", enh.Stormstrike, true},
		{"Fire Nova", topRank(t, "Fire Nova", enh.FireNova), true},
		{"Earth Shock", topRank(t, "Earth Shock", enh.EarthShock), false},
	}
	// checkCasts puts down Searing Totem for Fire Nova, then checks which spells we could
	// cast now.
	checkCasts := func(sim *core.Simulation, inRange bool) {
		castNow(t, sim, enh, searing)
		enh.SetGCDTimer(sim, sim.CurrentTime)
		for _, s := range spells {
			if want := inRange || !s.melee; s.spell.CanCast(sim, enh.CurrentTarget) != want {
				t.Errorf("at %s, in range %v: %s can be cast %v, want %v", sim.CurrentTime, inRange, s.name, !want, want)
			}
		}
	}

	samples, outOfRange := 0, 0
	wasOut, lastSwings := false, int32(0)
	checkedOut, checkedIn := false, false
	core.StartPeriodicAction(sim, core.PeriodicActionOptions{
		Period: 100 * time.Millisecond,
		OnAction: func(sim *core.Simulation) {
			out := !enh.IsInMeleeRange()
			swings := white.SpellMetrics[enh.CurrentTarget.UnitIndex].Casts
			samples++
			if out {
				outOfRange++
				if wasOut && swings != lastSwings {
					t.Errorf("at %s: %d white swings while out of melee range", sim.CurrentTime, swings-lastSwings)
				}
			}
			if sim.CurrentTime > 5*time.Second {
				if out && !checkedOut {
					checkedOut = true
					checkCasts(sim, false)
				} else if !out && checkedOut && !checkedIn {
					checkedIn = true
					checkCasts(sim, true)
				}
			}
			wasOut, lastSwings = out, swings
		},
	})
	runSim(sim)

	if !checkedOut || !checkedIn {
		t.Errorf("never checked the casts out of range (%v) and back in range (%v)", checkedOut, checkedIn)
	}
	if swings := white.SpellMetrics[enh.CurrentTarget.UnitIndex].Casts; swings < seconds/2/4 {
		t.Errorf("%d white swings in range, want at least %d", swings, seconds/2/4)
	}
	if share := float64(outOfRange) / float64(samples); math.Abs(share-0.5) > 0.05 {
		t.Errorf("out of melee range %.3f of the fight, want 0.5", share)
	}
}
