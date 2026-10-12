package retribution

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	googleProto "google.golang.org/protobuf/proto"
)

const arcaniteReaper = 12784 // 153 - 256, 3.8 speed, two-hand axe

// retSim starts a Forever sim of a level 60 Human paladin with talents, holding weapon,
// against a level 60 target with no armor, and returns it before the first action.
func retSim(t *testing.T, talents string, weapon int32) (*core.Simulation, *RetributionPaladin) {
	t.Helper()
	player := &proto.Player{
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassPaladin,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
			Options: &proto.PaladinOptions{PrimarySeal: proto.PaladinSeal_Righteousness},
		}},
	}
	for range proto.ItemSlot_ItemSlotMainHand {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
	}
	player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: weapon})
	target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
	target.Level = 60
	target.Stats[proto.Stat_StatArmor] = 0
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{target}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*RetributionPaladin)
}

// landedHit returns the damage of one cast of spell when it was a normal hit, and false
// for anything else.
func landedHit(sim *core.Simulation, ret *RetributionPaladin, spell *core.Spell) (float64, bool) {
	metrics := &spell.SpellMetrics[ret.CurrentTarget.UnitIndex]
	before := *metrics
	spell.ApplyEffects(sim, ret.CurrentTarget, spell)
	if metrics.Hits == before.Hits+1 && metrics.Crits == before.Crits {
		return metrics.TotalDamage - before.TotalDamage, true
	}
	return 0, false
}

// TestHolyStrikeDamage checks Holy Strike rank 8 against the client's spell (10333): 50% of
// (normalized weapon damage at Forever's 3.6 for a two-hander, plus 93 Holy with a 25%
// spread). Arcanite Reaper's 153 - 256 and the paladin's attack power give the range a normal
// hit must fall in. The old model (50% of the weapon at its own 3.8 speed, plus the whole 93)
// lands at least 38 higher.
func TestHolyStrikeDamage(t *testing.T) {
	sim, ret := retSim(t, "", arcaniteReaper)
	holyStrike := ret.GetSpell(core.ActionID{SpellID: 10333})
	if holyStrike == nil {
		t.Fatalf("no Holy Strike rank 8")
	}
	ap := holyStrike.MeleeAttackPower(ret.CurrentTarget)
	lo := 0.5 * (153 + 3.6*ap/14 + 93*0.875)
	hi := 0.5 * (256 + 3.6*ap/14 + 93*1.125)
	minHit, maxHit, hits := hi+1000, 0.0, 0
	for range 2000 {
		if hit, ok := landedHit(sim, ret, holyStrike); ok {
			minHit, maxHit, hits = min(minHit, hit), max(maxHit, hit), hits+1
		}
	}
	if hits < 1000 {
		t.Fatalf("only %d hits", hits)
	}
	if minHit < lo-1e-6 || maxHit > hi+1e-6 {
		t.Errorf("hits %.1f to %.1f, want within %.1f to %.1f", minHit, maxHit, lo, hi)
	}
	if minHit > lo+5 || maxHit < hi-5 {
		t.Errorf("hits %.1f to %.1f do not span %.1f to %.1f", minHit, maxHit, lo, hi)
	}
}

// TestTwistOfLightEcho checks Twist of Light's Echo: Seal of Righteousness replaced by Seal
// of Command leaves an Echo of Righteousness, and the next landed white hit fires a Seal of
// Righteousness proc (25713) under Seal of Command and uses the Echo up, so a minute of
// swings under Seal of Command gives exactly one. Without the talent it gives none.
func TestTwistOfLightEcho(t *testing.T) {
	for _, c := range []struct {
		name    string
		talents string
		procs   int
	}{
		{"Twist of Light", "--25225331001330301", 1},
		{"no Twist of Light", "--25225331001330300", 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim, ret := retSim(t, c.talents, arcaniteReaper)
			target := ret.CurrentTarget
			procMetrics := &ret.GetSpell(core.ActionID{SpellID: 25713}).SpellMetrics[target.UnitIndex]
			castByID(sim, ret, 20293)
			castByID(sim, ret, 20920)
			echo := ret.GetAura("Echo of Righteousness")
			if (echo != nil && echo.IsActive()) != (c.procs > 0) {
				t.Fatalf("Echo of Righteousness up: %v", echo != nil && echo.IsActive())
			}
			for !sim.Step() {
			}
			if procs := int(procMetrics.Hits + procMetrics.Crits); procs != c.procs {
				t.Errorf("%d Seal of Righteousness procs, want %d", procs, c.procs)
			}
			if c.procs > 0 && echo.IsActive() {
				t.Errorf("Echo of Righteousness still up after the fight")
			}
		})
	}
}

func castByID(sim *core.Simulation, ret *RetributionPaladin, spellID int32) {
	spell := ret.GetSpell(core.ActionID{SpellID: spellID})
	spell.ApplyEffects(sim, ret.CurrentTarget, spell)
}
