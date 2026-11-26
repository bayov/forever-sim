package core

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

// TestDazeChance checks the chance against cmangos' Unit::CalculateEffectiveDazeChance.
func TestDazeChance(t *testing.T) {
	for _, c := range []struct {
		level, mobLevel int32
		defense, want   float64
	}{
		{60, 60, 0, 0.20},
		{60, 62, 0, 0.216},  // 10 more weapon skill than our Defense
		{60, 60, 50, 0.12},  // 50 Defense over our level's
		{10, 10, 0, 0.07},   // 0.65 a level and 0.5 under 30
		{29, 29, 0, 0.1935}, // 19.35
		{1, 1, 100, 0},      // never under 0
		{30, 63, 0, 0.40},   // never over 40
	} {
		if got := dazeChance(c.level, c.mobLevel, c.defense); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("level %d hit by a level %d with %.0f bonus Defense: %.4f, want %.4f", c.level, c.mobLevel, c.defense, got, c.want)
		}
	}
}

// sessionWithMobs is a session of a level 30 at the origin, with mobs in these states.
func sessionWithMobs(opts *proto.LevelingOptions, mobs ...levelingMob) *levelingSession {
	env := &Environment{State: Finalized}
	env.sim = &Simulation{Environment: env}
	s := &levelingSession{env: env, opts: resolveLevelingOptions(opts, 30), level: 30, mobs: mobs}
	env.Encounter.leveling = s
	for i, m := range mobs {
		target := &Target{Unit: Unit{Type: EnemyUnit, Index: int32(i), Level: 30, Env: env}}
		env.Encounter.Targets = append(env.Encounter.Targets, target)
		env.Encounter.TargetUnits = append(env.Encounter.TargetUnits, &target.Unit)
		if m.state == mobRunning || m.state == mobArrived {
			env.Encounter.ActiveTargetUnits = append(env.Encounter.ActiveTargetUnits, &target.Unit)
		}
	}
	return s
}

// TestBehindChance: up to three mobs are in front of us while we stand, and each one past them
// puts more of the swings behind us. The ones that still run in don't count. While we run looking
// ahead the mobs on us are behind, and while we strafe they're as if we stood.
func TestBehindChance(t *testing.T) {
	zero := 0.0
	for _, c := range []struct {
		name          string
		opts          *proto.LevelingOptions
		onUs, running int
		moving        bool
		strafing      bool
		want          float64
	}{
		{"one", &proto.LevelingOptions{}, 1, 0, false, false, 0},
		{"three and two coming", &proto.LevelingOptions{}, 3, 2, false, false, 0},
		{"five", &proto.LevelingOptions{}, 5, 0, false, false, 0.08},
		{"ten", &proto.LevelingOptions{}, 10, 0, false, false, 0.25},
		{"ten against a wall", &proto.LevelingOptions{Surround: &zero}, 10, 0, false, false, 0},
		{"one while we run", &proto.LevelingOptions{Surround: &zero}, 1, 0, true, false, 0.95},
		{"four while we run", &proto.LevelingOptions{}, 4, 0, true, false, 0.75},
		{"five while we strafe", &proto.LevelingOptions{}, 5, 0, true, true, 0.08},
		{"four while we run, facing off", &proto.LevelingOptions{IgnoreFacing: true}, 4, 0, true, false, 0},
	} {
		var mobs []levelingMob
		for range c.onUs {
			mobs = append(mobs, levelingMob{state: mobArrived})
		}
		for range c.running {
			mobs = append(mobs, levelingMob{state: mobRunning})
		}
		s := sessionWithMobs(c.opts, mobs...)
		if c.moving {
			s.vel, s.runForward = Vec2{7, 0}, !c.strafing && !s.opts.ignoreFacing
		}
		if got := s.behindChance(); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%s: %.3f of the swings from behind, want %.3f", c.name, got, c.want)
		}
	}
}

// TestLevelingAreaTargets: an area ability reaches the mobs on us and those nearly here. It
// reaches neither the mob we found and haven't pulled nor one that's still 20 yards off. At a mob
// 20 yards off it reaches that one alone, and the mob it opens on counts too.
func TestLevelingAreaTargets(t *testing.T) {
	s := sessionWithMobs(&proto.LevelingOptions{},
		levelingMob{state: mobArrived},
		levelingMob{state: mobFound, x: Vec2{8, 0}},
		levelingMob{state: mobRunning, x: Vec2{20, 0}},
		levelingMob{state: mobRunning, x: Vec2{0, 9}},
		levelingMob{state: mobArrived},
		levelingMob{},
	)
	s.env.Encounter.Targets[5].dead = true
	enc := &s.env.Encounter
	walk := func(from int) []int32 {
		var hit []int32
		s.areaTarget = enc.TargetUnits[from]
		target := enc.TargetUnits[from]
		for range enc.AoETargetCount(0) {
			hit = append(hit, target.Index)
			target = s.env.NextTargetUnit(target)
		}
		return hit
	}
	for _, c := range []struct {
		from int
		want []int32
	}{{0, []int32{0, 3, 4}}, {4, []int32{4, 0, 3}}, {2, []int32{2}}, {1, []int32{1, 4, 0}}} {
		got := walk(c.from)
		seen := map[int32]bool{}
		for _, i := range got {
			seen[i] = true
		}
		ok := len(got) == len(c.want) && len(seen) == len(got)
		for _, i := range c.want {
			ok = ok && seen[i]
		}
		if !ok {
			t.Errorf("an area ability at mob %d hits %v, want %v", c.from, got, c.want)
		}
	}
	if n := enc.AoETargetCount(2); n != 2 {
		t.Errorf("one that hits two reaches %d", n)
	}
}

// TestLevelingBigPullDefaults: a session without the options has no camp and packs of five, as
// ever, and one with a camp looks around faster.
func TestLevelingBigPullDefaults(t *testing.T) {
	o := resolveLevelingOptions(&proto.LevelingOptions{}, 30)
	if o.campMobs != 0 || o.maxPack != 5 || o.surround != 1 || o.runForward != 0.5 || o.ignoreFacing || o.searchMax.Seconds() != 10 || o.chainSearchMax.Seconds() != 3 || o.standingSearchMax.Seconds() != 5 {
		t.Errorf("defaults %+v", o)
	}
	o = resolveLevelingOptions(&proto.LevelingOptions{CampMobs: 12, ChainSearchMax: 2}, 30)
	if o.campOutside != 7 || o.respawnTime.Seconds() != 300 || o.lootTime.Seconds() != 1 || o.searchMax.Seconds() != 1 || o.chainSearchMax.Seconds() != 2 || o.standingSearchMax.Seconds() != 1 {
		t.Errorf("camp defaults %+v", o)
	}
}
