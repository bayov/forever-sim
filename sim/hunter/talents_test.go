package hunter

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

// newHunterSim builds a level 60 Forever hunter on the Survival set with the given talents.
func newHunterSim(t *testing.T, talents string, debuffs *proto.Debuffs) (*core.Simulation, *Hunter) {
	player := &proto.Player{
		Class:         proto.Class_ClassHunter,
		Race:          proto.Race_RaceHuman,
		Level:         60,
		Equipment:     core.GetGearSet("../../ui/hunter/gear_sets", "level60_phase1_synthetic_survival").GearSet,
		TalentsString: talents,
		Spec:          PlayerOptionsCat,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, debuffs),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(HunterAgent).GetHunter()
}

func costMultiplier(t *testing.T, hunter *Hunter, code int32) int32 {
	for _, spell := range hunter.Spellbook {
		if spell.SpellCode == code && spell.Cost != nil {
			return spell.Cost.Multiplier
		}
	}
	t.Fatalf("no spell with code %d and a cost", code)
	return 0
}

// TestEfficiencyClassMask checks that Efficiency (19416) cuts the costs its class mask names
// (Raptor Strike and Mongoose Bite share bit 2, Wing Clip 64) and not Lacerate's (mask word 1
// bit 4096, outside Efficiency's 0x80080000).
func TestEfficiencyClassMask(t *testing.T) {
	// Lethal Attacks 5 and Efficiency 5.
	_, hunter := newHunterSim(t, "-00505", &proto.Debuffs{})
	if hunter.Talents.Efficiency != 5 {
		t.Fatalf("Efficiency %d, want 5", hunter.Talents.Efficiency)
	}
	for _, c := range []struct {
		name string
		code int32
		want int32
	}{
		{"Raptor Strike", SpellCode_HunterRaptorStrike, 85},
		{"Mongoose Bite", SpellCode_HunterMongooseBite, 85},
		{"Wing Clip", SpellCode_HunterWingClip, 85},
		{"Lacerate", SpellCode_HunterLacerate, 100},
	} {
		if got := costMultiplier(t, hunter, c.code); got != c.want {
			t.Errorf("%s cost multiplier %d, want %d", c.name, got, c.want)
		}
	}
}

// TestExposePreyWindow checks that Expose Prey (5% a rank) opens Mongoose Bite for 10 sec, as
// it has since build 70291, and only on a target with Hunter's Mark.
func TestExposePreyWindow(t *testing.T) {
	// The Survival preset's 34 points, with Expose Prey 2.
	const talents = "--500230230250122151"
	for _, mark := range []bool{true, false} {
		debuffs := &proto.Debuffs{}
		if mark {
			debuffs.HuntersMark = proto.TristateEffect_TristateEffectRegular
		}
		sim, hunter := newHunterSim(t, talents, debuffs)
		if hunter.Talents.ExposePrey != 2 {
			t.Fatalf("Expose Prey %d, want 2", hunter.Talents.ExposePrey)
		}
		target := hunter.CurrentTarget
		// 10% a hit, so 200 hits leave about 1 in 10^9 for no proc.
		for range 200 {
			hunter.WingClip.CalcAndDealDamage(sim, target, 1, hunter.WingClip.OutcomeAlwaysHit)
			if hunter.DefensiveState.IsActive() {
				break
			}
		}
		switch {
		case !mark && hunter.DefensiveState.IsActive():
			t.Errorf("Expose Prey opened Mongoose Bite without Hunter's Mark")
		case mark && !hunter.DefensiveState.IsActive():
			t.Errorf("Expose Prey never opened Mongoose Bite in 200 hits")
		case mark:
			if got := hunter.DefensiveState.RemainingDuration(sim); got != 10*time.Second {
				t.Errorf("Expose Prey opened Mongoose Bite for %v, want 10s", got)
			}
		}
	}
}
