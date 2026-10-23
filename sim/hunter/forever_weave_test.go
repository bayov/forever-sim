package hunter

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

// A Forever hunter with the given gear and talents at the given range, past the pull, with a
// rotation that does nothing on its own, and a way to run the sim up to a time.
func newWeaveSim(t *testing.T, gear string, talents string, distance float64) (*core.Simulation, *Hunter, func(time.Duration)) {
	t.Helper()
	player := &proto.Player{
		Class:              proto.Class_ClassHunter,
		Race:               proto.Race_RaceOrc,
		Level:              60,
		Equipment:          core.GetGearSet("../../ui/hunter/gear_sets", gear).GearSet,
		TalentsString:      talents,
		Spec:               PlayerOptionsNoPet,
		Rotation:           &proto.APLRotation{},
		DistanceFromTarget: distance,
		ReactionTimeMs:     150,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	hunter := sim.Raid.Parties[0].Players[0].(HunterAgent).GetHunter()
	stepTo := func(at time.Duration) {
		t.Helper()
		sim.AddPendingAction(&core.PendingAction{NextActionAt: at, OnAction: func(*core.Simulation) {}})
		for i := 0; sim.CurrentTime < at; i++ {
			if i > 100000 {
				t.Fatalf("the sim did not reach %v", at)
			}
			sim.Step()
		}
	}
	stepTo(time.Second)
	return sim, hunter, stepTo
}

// Walks to the given range and returns when the walk ended.
func walkTo(t *testing.T, sim *core.Simulation, hunter *Hunter, stepTo func(time.Duration), to float64) time.Duration {
	t.Helper()
	end := sim.CurrentTime + hunter.TimeToMoveTo(to)
	hunter.MoveTo(to, sim)
	stepTo(end + time.Microsecond)
	if hunter.IsMoving() || hunter.DistanceFromTarget != to {
		t.Fatalf("at %v the walk to %v yd is at %v yd, still moving %v", sim.CurrentTime, to, hunter.DistanceFromTarget, hunter.IsMoving())
	}
	return end
}

// A walk ends on its range, a fraction of a yard included, at the time Move Time gives: 7.01 yd
// takes 7 whole yards at 142 ms and a hundredth of one more.
func TestForeverWalkEndsOnAFractionOfAYard(t *testing.T) {
	sim, hunter, stepTo := newWeaveSim(t, "level60_phase1_synthetic_ranged", MarksmanshipTalents, 12)
	if got, want := hunter.TimeToMoveTo(4.99), 7*142*time.Millisecond+1420*time.Microsecond; (got - want).Abs() > time.Microsecond {
		t.Errorf("12 to 4.99 yd takes %v, want %v", got, want)
	}
	start := sim.CurrentTime
	if end := walkTo(t, sim, hunter, stepTo, 4.99); end-start != hunter.TimeToMoveTo(12) {
		t.Errorf("the walk in took %v, and the walk back out would take %v", end-start, hunter.TimeToMoveTo(12))
	}
}

// Under Forever an Auto Shot that came due while the hunter was in melee or on the move winds up
// for half a second a reaction time after the walk out ends, and one whose windup a walk overlaps
// does too.
func TestForeverAutoShotWindsUpAReactionAfterTheWalk(t *testing.T) {
	sim, hunter, stepTo := newWeaveSim(t, "level60_phase1_synthetic_ranged", MarksmanshipTalents, 12)
	aa := &hunter.AutoAttacks

	due := aa.NextRangedAttackAt()
	walkTo(t, sim, hunter, stepTo, 4.99)
	end := walkTo(t, sim, hunter, stepTo, 12)
	if end <= due {
		t.Fatalf("the weave ended at %v, before the shot came due at %v", end, due)
	}
	if got, want := aa.NextRangedAttackAt(), end+hunter.ReactionTime+core.RangedSwingWindup; got != want {
		t.Errorf("a shot due at %v during a weave that ended at %v fires at %v, want %v", due, end, got, want)
	}

	// A step that ends 0.3 sec before the next shot cuts into its windup.
	stepTo(aa.NextRangedAttackAt() + time.Millisecond)
	next := aa.NextRangedAttackAt()
	stepTo(next - 300*time.Millisecond - hunter.TimeToMoveTo(13))
	end = walkTo(t, sim, hunter, stepTo, 13)
	if got, want := aa.NextRangedAttackAt(), end+hunter.ReactionTime+core.RangedSwingWindup; got != want {
		t.Errorf("a step into the windup of the shot due at %v ended at %v and fires it at %v, want %v", next, end, got, want)
	}

	// One that ends before the windup leaves the shot alone.
	stepTo(aa.NextRangedAttackAt() + time.Millisecond)
	next = aa.NextRangedAttackAt()
	walkTo(t, sim, hunter, stepTo, 12)
	if got := aa.NextRangedAttackAt(); got != next {
		t.Errorf("a step well before the shot due at %v moved it to %v", next, got)
	}
}

// Forever's deadzone ends 8 - 4/3 yd past melee range (wowsims/forever 8a192337d8).
func TestForeverDeadzone(t *testing.T) {
	_, hunter, _ := newWeaveSim(t, "level60_phase1_synthetic_ranged", MarksmanshipTalents, 12)
	for _, c := range []struct {
		distance float64
		inRange  bool
	}{{11.6, false}, {11.67, true}, {12, true}} {
		hunter.DistanceFromTarget = c.distance
		if got := hunter.inRangedRange(); got != c.inRange {
			t.Errorf("at %v yd the hunter's shots are in range %v, want %v", c.distance, got, c.inRange)
		}
	}
}

// Strider Kick runs the hunter 30% faster for 3 sec.
func TestForeverStriderKickRunSpeed(t *testing.T) {
	sim, hunter, stepTo := newWeaveSim(t, "level60_phase1_synthetic_survival", SurvivalTalents, 5)
	if math.Abs(hunter.MovementHandler.MoveSpeed-7) > 1e-9 {
		t.Fatalf("the hunter runs %v yd a sec, want 7", hunter.MovementHandler.MoveSpeed)
	}
	if !hunter.StriderKick.Cast(sim, hunter.CurrentTarget) {
		t.Fatal("Strider Kick could not be cast")
	}
	if math.Abs(hunter.MovementHandler.MoveSpeed-7*1.3) > 1e-9 {
		t.Errorf("after Strider Kick the hunter runs %v yd a sec, want %v", hunter.MovementHandler.MoveSpeed, 7*1.3)
	}
	stepTo(sim.CurrentTime + 3*time.Second + time.Millisecond)
	if math.Abs(hunter.MovementHandler.MoveSpeed-7) > 1e-9 {
		t.Errorf("3 sec after Strider Kick the hunter runs %v yd a sec, want 7", hunter.MovementHandler.MoveSpeed)
	}
}

// Auto Shot turns the melee attack off, so the next walk in starts it again and the off hand's first
// swing comes half a swing after the arrival (wowsims/forever start_attack_test.go).
func TestForeverAutoShotTurnsTheAttackOff(t *testing.T) {
	sim, hunter, stepTo := newWeaveSim(t, "level60_phase1_synthetic_survival", SurvivalTalents, 12)
	aa := &hunter.AutoAttacks
	if !aa.IsDualWielding {
		t.Fatal("the Survival hunter does not dual wield")
	}
	stepTo(aa.NextRangedAttackAt() + time.Millisecond) // a shot goes off
	arrival := walkTo(t, sim, hunter, stepTo, 4.99)
	if got, want := aa.OffhandSwingAt(), arrival+aa.OffhandSwingSpeed()/2+50*time.Millisecond; got < want {
		t.Errorf("back in at %v after a shot, the off hand swings at %v, want no sooner than %v", arrival, got, want)
	}
}
