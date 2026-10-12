package dps

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

func newWarlockSim(t *testing.T, talents string, options *proto.WarlockOptions) (*core.Simulation, *DpsWarlock) {
	player := &proto.Player{
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarlock,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec:          &proto.Player_Warlock{Warlock: &proto.Warlock{Options: options}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 120, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*DpsWarlock)
}

func castNow(t *testing.T, sim *core.Simulation, lock *DpsWarlock, spell *core.Spell) {
	lock.SetGCDTimer(sim, sim.CurrentTime)
	lock.Hardcast.Expires = sim.CurrentTime
	if !spell.Cast(sim, lock.CurrentTarget) {
		t.Fatalf("at %s: %s didn't cast", sim.CurrentTime, spell.ActionID)
	}
}

func at(sim *core.Simulation, seconds float64, fn func(sim *core.Simulation)) {
	sim.AddPendingAction(&core.PendingAction{NextActionAt: core.DurationFromSeconds(seconds), OnAction: fn})
}

func runSim(sim *core.Simulation) {
	for !sim.Step() {
	}
}

// TestWarlockBanesAndCurses checks that a bane and a curse share the target under Forever, and
// that a second bane takes the first one's place. Spell hit is maxed so nothing misses.
func TestWarlockBanesAndCurses(t *testing.T) {
	sim, lock := newWarlockSim(t, "", &proto.WarlockOptions{})
	target := lock.CurrentTarget
	agony := lock.BaneOfAgony[len(lock.BaneOfAgony)-1]

	at(sim, 0.5, func(sim *core.Simulation) {
		lock.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
		castNow(t, sim, lock, lock.CurseOfElements)
	})
	at(sim, 3, func(sim *core.Simulation) { castNow(t, sim, lock, agony) })
	at(sim, 6, func(sim *core.Simulation) {
		if !lock.CurseOfElementsAuras.Get(target).IsActive() || !agony.Dot(target).IsActive() {
			t.Errorf("Curse of the Elements up %v, Bane of Agony up %v, want both",
				lock.CurseOfElementsAuras.Get(target).IsActive(), agony.Dot(target).IsActive())
		}
		castNow(t, sim, lock, lock.BaneOfDoom)
	})
	at(sim, 9, func(sim *core.Simulation) {
		if agony.Dot(target).IsActive() || !lock.BaneOfDoom.Dot(target).IsActive() {
			t.Errorf("Bane of Agony up %v, Bane of Doom up %v, want only Doom",
				agony.Dot(target).IsActive(), lock.BaneOfDoom.Dot(target).IsActive())
		}
		if !lock.CurseOfElementsAuras.Get(target).IsActive() {
			t.Errorf("Bane of Doom took Curse of the Elements off")
		}
	})
	runSim(sim)
}

// TestWarlockDemonicPact checks the Imp sacrificed before the pull: its 15% Shadow damage stays
// with the Succubus out only with Demonic Pact, and goes with any demon out without it.
func TestWarlockDemonicPact(t *testing.T) {
	const withPact = "-0305003131221001351"
	const withoutPact = "-030500313122100135"
	for _, c := range []struct {
		talents string
		summon  proto.WarlockOptions_Summon
		want    bool
	}{
		{withPact, proto.WarlockOptions_Succubus, true},
		{withoutPact, proto.WarlockOptions_Succubus, false},
		{withoutPact, proto.WarlockOptions_NoSummon, true},
	} {
		sim, lock := newWarlockSim(t, c.talents, &proto.WarlockOptions{Summon: c.summon, SacrificedPet: proto.WarlockOptions_Imp})
		var up bool
		at(sim, 1, func(sim *core.Simulation) {
			up = lock.DemonicSacrificeAuras[lock.Imp].IsActive()
		})
		runSim(sim)
		if up != c.want {
			t.Errorf("%s with demon %s: Burning Shadow up %v, want %v", c.talents, c.summon, up, c.want)
		}
	}
}
