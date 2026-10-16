package hunter

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	googleProto "google.golang.org/protobuf/proto"
)

// newHunterSim builds a level 60 Forever hunter on the Survival set with the given talents.
func newHunterSim(t *testing.T, talents string, debuffs *proto.Debuffs) (*core.Simulation, *Hunter) {
	return newHunterSimAgainst(t, talents, debuffs, core.NewDefaultTarget())
}

func newHunterSimAgainst(t *testing.T, talents string, debuffs *proto.Debuffs, target *proto.Target) (*core.Simulation, *Hunter) {
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
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{target}},
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

// TestResourcefulnessClassMask checks that Resourcefulness 1 (440529) takes 30% off the costs its
// mask names, Strider Kick included since client 1.60.1.70425 (mask word 2 bit 0x200000), and not
// Lacerate's (word 1 bit 4096).
func TestResourcefulnessClassMask(t *testing.T) {
	// The Survival preset's 34 points, with Resourcefulness 1 and Strider Kick.
	_, hunter := newHunterSim(t, "--500230230250122151", &proto.Debuffs{})
	if hunter.Talents.Resourcefulness != 1 || !hunter.Talents.StriderKick {
		t.Fatalf("Resourcefulness %d, Strider Kick %v, want 1 and true", hunter.Talents.Resourcefulness, hunter.Talents.StriderKick)
	}
	for _, c := range []struct {
		name string
		code int32
		want int32
	}{
		{"Raptor Strike", SpellCode_HunterRaptorStrike, 70},
		{"Mongoose Bite", SpellCode_HunterMongooseBite, 70},
		{"Wing Clip", SpellCode_HunterWingClip, 70},
		{"Strider Kick", SpellCode_HunterStriderKick, 70},
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

// TestImprovedTrackingGenericBoss checks that Improved Tracking's 5% counts against the generic
// Level 60 boss (type Unknown, standards 19), whose real type a hunter can always track, and
// against a Humanoid, but not against a Mechanical, which no tracking covers.
func TestImprovedTrackingGenericBoss(t *testing.T) {
	// Improved Tracking 5.
	const talents = "--5"
	for _, c := range []struct {
		mobType proto.MobType
		want    float64
	}{
		{proto.MobType_MobTypeUnknown, 1.05},
		{proto.MobType_MobTypeHumanoid, 1.05},
		{proto.MobType_MobTypeMechanical, 1},
	} {
		target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
		target.MobType = c.mobType
		_, hunter := newHunterSimAgainst(t, talents, &proto.Debuffs{}, target)
		if hunter.Talents.ImprovedTracking != 5 {
			t.Fatalf("Improved Tracking %d, want 5", hunter.Talents.ImprovedTracking)
		}
		for castType, at := range hunter.AttackTables[hunter.CurrentTarget.UnitIndex] {
			if at != nil && at.DamageDealtMultiplier != c.want {
				t.Errorf("%v, cast type %v: damage multiplier %.3f, want %.3f", c.mobType, castType, at.DamageDealtMultiplier, c.want)
			}
		}
	}
}

// TestPetAbilityDamage checks Claw's and Bite's damage at each level the pet has them against the
// client (build 1.60.1.70338): base points times 1 -/+ half the variance, rounded. Bite at 60 is
// 81 to 99 (17261: 90, variance 0.2), as in Classic.
func TestPetAbilityDamage(t *testing.T) {
	type effect struct{ basePoints, variance float64 }
	for _, c := range []struct {
		name   string
		table  map[int32][2]float64
		client map[int32]effect
	}{
		{"Claw", petClawDamage, map[int32]effect{25: {19, 0.31578946114}, 40: {31, 0.32258063555}, 50: {42, 0.33333334327}, 60: {51, 0.3137255013}}},
		{"Bite", petBiteDamage, map[int32]effect{25: {34, 0.17647059262}, 40: {54, 0.18518517911}, 50: {73, 0.19178082049}, 60: {90, 0.20000000298}}},
	} {
		for level, e := range c.client {
			want := [2]float64{math.Round(e.basePoints * (1 - e.variance/2)), math.Round(e.basePoints * (1 + e.variance/2))}
			if got := c.table[level]; got != want {
				t.Errorf("%s at level %d: %v, want %v", c.name, level, got, want)
			}
		}
	}
}

// TestHuntersMarkDuty checks the hunter's own Hunter's Mark, cast by the presets on Hunter's Mark
// duty (standards 23) with the raid's debuff off: it puts the same mark on the target as the
// debuff (140 ranged attack power at level 60, core.ForeverHuntersMarkRank4), for 2 min, at 60 mana.
func TestHuntersMarkDuty(t *testing.T) {
	sim, hunter := newHunterSim(t, "", &proto.Debuffs{})
	target := hunter.CurrentTarget
	table := hunter.AttackTables[target.UnitIndex][proto.CastType_CastTypeRanged]
	if table.BonusAttackPowerTaken != 0 {
		t.Fatalf("ranged attack power against the target before the mark: %v, want 0", table.BonusAttackPowerTaken)
	}
	if hunter.HuntersMark == nil {
		t.Fatal("no Hunter's Mark spell")
	}
	if got := hunter.HuntersMark.ActionID.SpellID; got != 14325 {
		t.Errorf("Hunter's Mark spell %d, want rank 4 (14325)", got)
	}
	mana := hunter.CurrentMana()
	if !hunter.HuntersMark.Cast(sim, target) {
		t.Fatal("Hunter's Mark didn't cast")
	}
	if spent := mana - hunter.CurrentMana(); spent != 60 {
		t.Errorf("Hunter's Mark cost %v mana, want 60", spent)
	}
	if got := table.BonusAttackPowerTaken; got != core.ForeverHuntersMarkRank4 {
		t.Errorf("ranged attack power against the marked target: %v, want %v", got, core.ForeverHuntersMarkRank4)
	}
	if aura := target.GetActiveAuraWithTag(core.HuntersMarkAuraTag); aura == nil || aura.RemainingDuration(sim) != 2*time.Minute {
		t.Errorf("mark aura %v, want 2 min left", aura)
	}
}

// TestScorpidStingDuty checks the hunter's own Scorpid Sting (standards 23 duty): 9% of base mana,
// 2% off the target's chance to hit for 20 sec, 65 with Improved Stings 3, and it takes the place of
// the hunter's Serpent Sting.
func TestScorpidStingDuty(t *testing.T) {
	for _, c := range []struct {
		talents  string
		duration time.Duration
	}{
		{"", 20 * time.Second},
		// Lethal Attacks 5 and Improved Stings 3.
		{"-0053", 65 * time.Second},
	} {
		sim, hunter := newHunterSim(t, c.talents, &proto.Debuffs{})
		target := hunter.CurrentTarget
		hunter.DistanceFromTarget = 25
		if hunter.ScorpidSting == nil {
			t.Fatal("no Scorpid Sting spell")
		}
		hunter.SerpentSting.Dot(target).Apply(sim)
		miss := target.PseudoStats.IncreasedMissChance
		mana := hunter.CurrentMana()
		if !hunter.ScorpidSting.Cast(sim, target) {
			t.Fatal("Scorpid Sting didn't cast")
		}
		if spent, want := mana-hunter.CurrentMana(), 0.09*hunter.BaseMana; math.Abs(spent-want) > 1e-6 {
			t.Errorf("Scorpid Sting cost %v mana, want %v (9%% of base mana)", spent, want)
		}
		aura := hunter.scorpidStingAuras.Get(target)
		for !aura.IsActive() && sim.CurrentTime < 2*time.Second {
			if sim.Step() {
				break
			}
		}
		if !aura.IsActive() {
			t.Fatal("Scorpid Sting didn't land")
		}
		if got := aura.Duration; got != c.duration {
			t.Errorf("talents %q: Scorpid Sting lasts %v, want %v", c.talents, got, c.duration)
		}
		if got := target.PseudoStats.IncreasedMissChance - miss; math.Abs(got-0.02) > 1e-9 {
			t.Errorf("Scorpid Sting raised the target's miss chance by %v, want 0.02", got)
		}
		if hunter.SerpentSting.Dot(target).IsActive() {
			t.Error("Serpent Sting still up after Scorpid Sting")
		}
	}
}

// TestPushbackFlags checks the client's interrupt flags (1.60.1.70425): Multi-Shot's lack the
// pushback bit, Aimed Shot's carry it.
func TestPushbackFlags(t *testing.T) {
	_, hunter := newHunterSim(t, "", &proto.Debuffs{})
	if got := hunter.MultiShot.PushbackReduction; got != 1 {
		t.Errorf("Multi-Shot pushback reduction %v, want 1", got)
	}
	if got := hunter.AimedShot.PushbackReduction; got != 0 {
		t.Errorf("Aimed Shot pushback reduction %v, want 0", got)
	}
}
