//go:build practice

package practice

import (
	"testing"
	"time"
)

func TestToolkitHasTheEnhSpells(t *testing.T) {
	s := newEnhSession(t, Settings{})
	ids := map[int32]bool{}
	for _, spell := range s.Toolkit() {
		ids[spell.SpellID] = true
	}
	for _, id := range []int32{lightningBolt, earthShock, stormstrike, flameShock, magmaTotem, searingTotem, windfuryTotem, strengthOfEarth, lightningShield} {
		if !ids[id] {
			t.Errorf("spell %d isn't in the toolkit", id)
		}
	}
	if t.Failed() {
		t.Logf("toolkit: %+v", s.Toolkit())
	}
}

func TestFreeModePullAndStop(t *testing.T) {
	s := newEnhSession(t, Settings{Mode: "free"})
	d := newDriver(t, s)
	d.wait(2 * time.Second)

	// Buff up before the pull, then walk into melee.
	d.mustCast(lightningShield)
	d.wait(1500 * time.Millisecond)
	d.mustCast(strengthOfEarth)
	d.wait(time.Second)
	d.standAt(0, -3)
	if s.pulled {
		t.Fatalf("pulled before attacking")
	}

	// Earth Shock is the pull.
	d.mustCast(earthShock)
	if !s.pulled {
		t.Fatalf("Earth Shock didn't pull")
	}
	if s.Now() < 0 || s.Now() > 10*time.Millisecond {
		t.Fatalf("after the pull the sim is at %s, want about 0", s.Now())
	}
	if s.run.damage <= 0 {
		t.Fatalf("the pull's Earth Shock did no damage")
	}

	// Auto attack and Stormstrike for 10 seconds.
	d.press(Input{Kind: "startAttack"})
	d.wait(1500 * time.Millisecond)
	d.mustCast(stormstrike)
	d.wait(10 * time.Second)

	s.Stop(d.now)
	if !s.Ended() {
		t.Fatalf("Stop didn't end the fight")
	}
	sum := s.Summary()
	if sum == nil || sum.Damage <= 0 || sum.DPS <= 0 {
		t.Fatalf("summary: %+v", sum)
	}
	if want := toMs(d.now); sum.DurationMs < want-1 || sum.DurationMs > want+1 {
		t.Errorf("the summary covers %.0fms, want %.0fms", sum.DurationMs, want)
	}
	t.Logf("summary: %.0f damage in %.1fs, %.1f DPS, %d casts, idle GCD %.1fs", sum.Damage, sum.DurationMs/1000, sum.DPS, sum.Casts, sum.IdleGCDMs/1000)
	for _, sp := range sum.Spells {
		t.Logf("  %+v", sp)
	}
}
