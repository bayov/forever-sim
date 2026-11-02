//go:build practice

package practice

import (
	"math"
	"os"
	"sort"
	"testing"
	"time"
)

// Spike 4's worst case, natively or in WASM (GOOS=js GOARCH=wasm go test -exec go_js_wasm_exec):
// a full prepull window of moving, turning and totems, then the pull, which replays all of it
// into a fresh sim. It also times each step of a fight. Run with PRACTICE_TIMING=1.
func TestSwapTiming(t *testing.T) {
	if os.Getenv("PRACTICE_TIMING") == "" {
		t.Skip("set PRACTICE_TIMING=1")
	}
	start := time.Now()
	s := newEnhSession(t, Settings{})
	t.Logf("session start (NewSim, Reset, PrePull): %s", time.Since(start))
	d := newDriver(t, s)

	// 9 minutes of prepull: strafing and turning all the time (a move input every 100ms and a
	// world tick every 50ms), and a totem or shield every 10 seconds.
	spells := []int32{strengthOfEarth, windfuryTotem, graceOfAir, lightningShield, stoneskin}
	for i := 0; d.now < -60*time.Second; i++ {
		angle := float64(i) * 0.05
		d.press(Input{Kind: "move", Pos: [2]float64{10 * math.Cos(angle), 10 * math.Sin(angle)}, Velocity: [2]float64{-7 * math.Sin(angle), 7 * math.Cos(angle)}, Facing: angle, TurnRate: 0.5})
		if i%100 == 0 {
			d.press(Input{Kind: "cast", SpellID: spells[(i/100)%len(spells)]})
		}
		d.wait(100 * time.Millisecond)
	}
	t.Logf("%d inputs before the pull", len(s.inputs))
	d.standAt(0, -3)
	var swapTime time.Duration
	testOnSwap = func(old, new *run, shift time.Duration) {}
	defer func() { testOnSwap = nil }()
	t0 := time.Now()
	d.press(Input{Kind: "cast", SpellID: earthShock, Target: &UnitRef{Kind: "enemy"}})
	d.wait(time.Millisecond)
	swapTime = time.Since(t0)
	if !s.pulled {
		t.Fatal("no pull")
	}
	t.Logf("the pull after a 9 minute prepull, swap included: %s", swapTime)

	// A 60 second fight with a press every 300ms: each step's time.
	d.press(Input{Kind: "startAttack"})
	var steps []time.Duration
	keys := []int32{stormstrike, earthShock, lightningBolt, flameShock}
	for i := 0; i < 200; i++ {
		t0 := time.Now()
		d.press(Input{Kind: "cast", SpellID: keys[i%len(keys)], Target: &UnitRef{Kind: "enemy"}})
		d.wait(300 * time.Millisecond)
		steps = append(steps, time.Since(t0))
	}
	sort.Slice(steps, func(i, j int) bool { return steps[i] < steps[j] })
	t.Logf("a press and 300ms of fight: p50 %s, p99 %s, max %s", steps[len(steps)/2], steps[len(steps)*99/100], steps[len(steps)-1])
}
