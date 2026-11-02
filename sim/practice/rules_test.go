//go:build practice

package practice

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// activeAura finds your active aura whose label contains the text.
func activeAura(r *run, text string) *core.Aura {
	for _, aura := range r.char.GetAuras() {
		if aura.IsActive() && strings.Contains(aura.Label, text) && aura.ExpiresAt() > r.sim.CurrentTime {
			return aura
		}
	}
	return nil
}

func ms(n int) time.Duration { return time.Duration(n) * time.Millisecond }

// The queue: the last press in the window wins, a press outside it fails.
func TestSpellQueue(t *testing.T) {
	s := newEnhSession(t, Settings{})
	d := newDriver(t, s)
	d.mustCast(lightningShield) // 1.5s GCD from here
	start := s.run.gcdStart

	// 500ms before the GCD's end is outside the 400ms window.
	d.wait(ms(999))
	if a := d.cast(strengthOfEarth); a.OK || a.Error != errNotReady {
		t.Fatalf("a press 500ms early: %+v, want %q", a, errNotReady)
	}

	// Two presses in the window: the last one wins, at the GCD's end exactly.
	d.wait(ms(200))
	first := d.press(Input{Kind: "cast", SpellID: strengthOfEarth})
	d.wait(ms(100))
	second := d.press(Input{Kind: "cast", SpellID: stoneskin})
	batches := d.wait(ms(500))
	if a := answerOf(batches, first); a == nil || a.OK {
		t.Errorf("the replaced press: %+v, want a silent failure", a)
	}
	if a := answerOf(batches, second); a == nil || !a.OK {
		t.Errorf("the last press: %+v, want it cast", a)
	}
	if got := s.run.gcdStart; got != start+1500*time.Millisecond {
		t.Errorf("the queued cast went off at %s, want the GCD's end %s", got, start+1500*time.Millisecond)
	}
}

// Ping moves when an input applies by half the ping, but a queued press still fires exactly at
// the GCD's end.
func TestPingShiftsApplyTimes(t *testing.T) {
	s := newEnhSession(t, Settings{PingMs: 100})
	d := newDriver(t, s)
	d.wait(time.Second)
	pressed := d.now
	d.mustCastAfter(lightningShield, ms(60))
	if got, want := s.run.gcdStart, pressed+50*time.Millisecond; got != want {
		t.Errorf("a press at %s cast at %s, want %s", pressed, got, want)
	}
	// Queued 300ms before the GCD's end, it fires at the end, not 50ms after.
	gcdEnd := s.run.char.GCD.ReadyAt()
	d.wait(gcdEnd - d.now - ms(300))
	d.mustCastAfter(strengthOfEarth, ms(400))
	if got := s.run.gcdStart; got != gcdEnd {
		t.Errorf("the queued press cast at %s, want the GCD's end %s", got, gcdEnd)
	}
}

func (d *driver) mustCastAfter(spellID int32, wait time.Duration) {
	d.t.Helper()
	id := d.press(Input{Kind: "cast", SpellID: spellID, Target: &UnitRef{Kind: "target"}})
	if a := answerOf(d.wait(wait), id); a == nil || !a.OK {
		d.t.Fatalf("cast %d: %+v", spellID, a)
	}
}

// Casting checks: moving, facing, turning away during a cast, Esc.
func TestCastingChecks(t *testing.T) {
	s := newEnhSession(t, Settings{})
	d := newDriver(t, s)
	d.standAt(0, -20)

	// A cast while moving fails.
	d.press(Input{Kind: "move", Pos: [2]float64{0, -20}, Velocity: [2]float64{0, 7}, Facing: math.Pi / 2})
	if a := d.cast(lightningBolt); a.OK || a.Error != errMoving {
		t.Fatalf("Lightning Bolt on the move: %+v", a)
	}
	// An instant works on the move, but the dummy isn't in Earth Shock's range yet... we're at
	// about 20 yards, in range.
	d.standAt(0, -20)

	// Facing away fails.
	d.faceAway()
	if a := d.cast(earthShock); a.OK || a.Error != errFacing {
		t.Fatalf("Earth Shock facing away: %+v", a)
	}
	d.standAt(0, -20)

	// Out of range.
	d.standAt(0, -40)
	if a := d.cast(earthShock); a.OK || a.Error != errOutOfRange {
		t.Fatalf("Earth Shock at 40 yards: %+v", a)
	}
	d.standAt(0, -25)

	// Esc during a cast cancels it and keeps the target.
	d.mustCast(lightningBolt)
	if !s.run.char.IsCasting(s.run.sim) {
		t.Fatal("Lightning Bolt isn't casting")
	}
	d.press(Input{Kind: "cancelCast"})
	d.wait(ms(10))
	if s.run.char.IsCasting(s.run.sim) || !s.run.hasTarget {
		t.Fatalf("after Esc: casting %v, target %v", s.run.char.IsCasting(s.run.sim), s.run.hasTarget)
	}
	if s.pulled {
		t.Fatal("a cancelled bolt pulled")
	}

	// Turning away during the cast fails it at the end.
	d.wait(2 * time.Second)
	d.mustCast(lightningBolt)
	d.wait(time.Second)
	d.faceAway()
	d.wait(3 * time.Second)
	if s.pulled {
		t.Fatal("a bolt that failed at its end pulled")
	}
	if s.run.char.IsCasting(s.run.sim) {
		t.Fatal("still casting")
	}
}

// Moving starts interrupt a cast.
func TestMovingInterruptsCast(t *testing.T) {
	s := newEnhSession(t, Settings{})
	d := newDriver(t, s)
	d.standAt(0, -20)
	d.mustCast(lightningBolt)
	d.wait(ms(500))
	d.press(Input{Kind: "move", Pos: [2]float64{0, -20}, Velocity: [2]float64{7, 0}, Facing: math.Pi / 2})
	d.wait(ms(5))
	if s.run.char.IsCasting(s.run.sim) {
		t.Fatal("moving didn't stop the cast")
	}
	d.wait(3 * time.Second)
	if s.pulled {
		t.Fatal("an interrupted bolt pulled")
	}
}

// Auto attack: no swing before you turn it on, leaving melee range stops swings, and coming back
// resumes them without a timer reset.
func TestAutoAttack(t *testing.T) {
	s := newEnhSession(t, Settings{})
	d := newDriver(t, s)
	d.standAt(0, -3)
	d.wait(10 * time.Second)
	if s.pulled || s.run.damage > 0 {
		t.Fatal("swung without auto attack on")
	}

	// /startattack in range pulls, and the first swing comes at once.
	d.press(Input{Kind: "startAttack"})
	d.wait(ms(50))
	if !s.pulled {
		t.Fatal("auto attack on in range didn't pull")
	}
	r := s.run
	first := r.hits[hitKey{otherID: int32(1), tag: 1}]
	_ = first
	if r.damage == 0 && r.hits == nil {
		t.Fatal("no swing at the pull")
	}
	swings := func() int32 {
		var n int32
		for k, h := range r.hits {
			if k.otherID != 0 {
				n += h.Hits + h.Misses
			}
		}
		return n
	}
	if swings() != 1 {
		t.Fatalf("%d swings at the pull, want 1", swings())
	}

	// Walk out of range for 10 seconds: no swings.
	speed := r.char.AutoAttacks.MainhandSwingSpeed()
	d.wait(speed / 2)
	d.standAt(0, -15)
	before := swings()
	d.wait(10 * time.Second)
	if got := swings(); got != before {
		t.Fatalf("%d swings out of range", got-before)
	}
	// Come back: the overdue swing lands at once, not a full swing later.
	d.standAt(0, -3)
	d.wait(ms(50))
	if got := swings(); got != before+1 {
		t.Fatalf("coming back gave %d swings, want 1 at once", got-before)
	}
}

// No swing lands during a cast.
func TestNoSwingDuringCast(t *testing.T) {
	s := newEnhSession(t, Settings{})
	d := newDriver(t, s)
	d.standAt(0, -3)
	d.press(Input{Kind: "startAttack"})
	d.wait(ms(100))
	r := s.run
	countSwings := func() int32 {
		var n int32
		for k, h := range r.hits {
			if k.otherID != 0 {
				n += h.Hits + h.Misses
			}
		}
		return n
	}
	d.mustCast(lightningBolt)
	castEnd := r.char.Hardcast.Expires
	before := countSwings()
	d.wait(castEnd - d.now - ms(1))
	if got := countSwings(); got != before {
		t.Fatalf("%d swings during the cast", got-before)
	}
}

// The pull: Lightning Bolt counts from its landing, Magma Totem's first pulse pulls, walking into
// range with auto attack on pulls, a failed cast doesn't.
func TestPullTriggers(t *testing.T) {
	t.Run("bolt", func(t *testing.T) {
		s := newEnhSession(t, Settings{})
		d := newDriver(t, s)
		d.standAt(0, -25)
		d.mustCast(lightningBolt)
		castEnd := s.run.char.Hardcast.Expires
		d.wait(castEnd - d.now - ms(1))
		if s.pulled {
			t.Fatal("pulled before the bolt landed")
		}
		d.wait(s.run.toolkit[lightningBolt].TravelTime() + ms(100))
		if !s.pulled || s.run.damage <= 0 {
			t.Fatalf("the bolt didn't pull: pulled %v, damage %.0f", s.pulled, s.run.damage)
		}
		if h := s.run.hits[hitKey{spellID: lightningBolt}]; h == nil || h.Hits+h.Misses != 1 {
			t.Fatalf("the bolt isn't counted once: %+v", h)
		}
	})
	t.Run("magma", func(t *testing.T) {
		s := newEnhSession(t, Settings{})
		d := newDriver(t, s)
		d.standAt(0, -4)
		d.mustCast(magmaTotem)
		d.wait(5 * time.Second)
		if !s.pulled {
			t.Fatal("Magma Totem's pulse didn't pull")
		}
	})
	t.Run("magma out of reach", func(t *testing.T) {
		s := newEnhSession(t, Settings{})
		d := newDriver(t, s)
		d.standAt(0, -15)
		d.mustCast(magmaTotem)
		d.wait(25 * time.Second)
		if s.pulled {
			t.Fatal("a Magma Totem 13 yards away pulled")
		}
	})
	t.Run("walk in", func(t *testing.T) {
		s := newEnhSession(t, Settings{})
		d := newDriver(t, s)
		d.standAt(0, -15)
		d.press(Input{Kind: "startAttack"})
		d.wait(time.Second)
		if s.pulled {
			t.Fatal("auto attack out of range pulled")
		}
		d.press(Input{Kind: "move", Pos: [2]float64{0, -15}, Velocity: [2]float64{0, 7}, Facing: math.Pi / 2})
		d.wait(3 * time.Second)
		if !s.pulled {
			t.Fatal("walking into range didn't pull")
		}
	})
	t.Run("failed cast", func(t *testing.T) {
		s := newEnhSession(t, Settings{})
		d := newDriver(t, s)
		d.standAt(0, -40)
		if a := d.cast(earthShock); a.OK {
			t.Fatal("Earth Shock at 40 yards went off")
		}
		d.wait(time.Second)
		if s.pulled {
			t.Fatal("a failed cast pulled")
		}
	})
}

// The state at the pull is the same with and without the swap.
func TestSwapKeepsState(t *testing.T) {
	s := newEnhSession(t, Settings{})
	d := newDriver(t, s)
	d.mustCast(lightningShield)
	d.wait(2 * time.Second)
	d.mustCast(strengthOfEarth)
	d.wait(2 * time.Second)
	d.mustCast(windfuryTotem)
	// Mana fills up again, so regen settled at different times doesn't matter.
	d.wait(2 * time.Minute)
	d.standAt(0, -3)
	d.wait(3 * time.Second)

	type snapshot struct {
		mana     float64
		gcd      time.Duration
		auras    map[string]time.Duration
		cooldown map[int32]time.Duration
	}
	take := func(r *run) snapshot {
		now := r.sim.CurrentTime
		snap := snapshot{mana: r.char.CurrentMana(), gcd: r.char.GCD.ReadyAt() - now, auras: map[string]time.Duration{}, cooldown: map[int32]time.Duration{}}
		for _, aura := range r.char.GetAuras() {
			if aura.IsActive() && aura.ExpiresAt() != core.NeverExpires && aura.ExpiresAt() > now {
				snap.auras[aura.Label] = aura.ExpiresAt() - now
			}
		}
		for id, spell := range r.toolkit {
			if !spell.IsReady(r.sim) {
				snap.cooldown[id] = spell.ReadyAt() - now
			}
		}
		return snap
	}
	var before, after snapshot
	var swapped bool
	testOnSwap = func(old, new *run, shift time.Duration) {
		before, swapped = take(old), true
		// The new sim stops just before 0. Its auras and cooldowns are counted from 0.
		after = take(new)
		for k, v := range after.auras {
			after.auras[k] = v + new.sim.CurrentTime
		}
		for k, v := range after.cooldown {
			after.cooldown[k] = v + new.sim.CurrentTime
		}
		after.gcd += new.sim.CurrentTime
	}
	defer func() { testOnSwap = nil }()

	d.press(Input{Kind: "startAttack"})
	d.wait(ms(1))
	if !swapped {
		t.Fatal("no swap")
	}
	if before.mana != after.mana {
		t.Errorf("mana %.3f before the swap, %.3f after", before.mana, after.mana)
	}
	if max(before.gcd, 0) != max(after.gcd, 0) {
		t.Errorf("GCD %s before, %s after", before.gcd, after.gcd)
	}
	if len(before.auras) == 0 {
		t.Error("no auras with time left")
	}
	for label, left := range before.auras {
		if after.auras[label] != left {
			t.Errorf("%s: %s left before the swap, %s after", label, left, after.auras[label])
		}
	}
	for id, left := range before.cooldown {
		if after.cooldown[id] != left {
			t.Errorf("spell %d: %s cooldown before, %s after", id, left, after.cooldown[id])
		}
	}
}

// The prepull window: a totem dropped before the pull is up at the pull, and a totem replaced
// before the pull loses its buff at once.
func TestPrepullTotems(t *testing.T) {
	s := newEnhSession(t, Settings{})
	d := newDriver(t, s)
	d.mustCast(strengthOfEarth)
	d.wait(30 * time.Second)
	if activeAura(s.run, "Strength of Earth") == nil {
		t.Fatal("no Strength of Earth after the totem")
	}
	d.standAt(0, -3)
	d.mustCast(earthShock)
	if !s.pulled {
		t.Fatal("no pull")
	}
	if activeAura(s.run, "Strength of Earth") == nil {
		t.Fatal("Strength of Earth is gone at the pull")
	}
}

func TestReplacedTotemLosesBuffBeforePull(t *testing.T) {
	s := newEnhSession(t, Settings{})
	d := newDriver(t, s)
	d.mustCast(strengthOfEarth)
	d.wait(2 * time.Second)
	d.mustCast(stoneskin)
	d.wait(ms(10))
	if activeAura(s.run, "Strength of Earth") != nil {
		t.Fatal("Strength of Earth stayed after Stoneskin Totem replaced it")
	}
}

// Waiting 70 minutes before the pull and then casting works.
func TestLongPrepull(t *testing.T) {
	s := newEnhSession(t, Settings{})
	d := newDriver(t, s)
	d.standAt(0, -3)
	d.press(Input{Kind: "target", Target: &UnitRef{Kind: "enemy"}})
	for i := 0; i < 70; i++ {
		d.wait(time.Minute)
		if s.pulled {
			t.Fatalf("pulled after %d minutes", i)
		}
	}
	if s.Now() > 0 {
		t.Fatalf("the sim is at %s before the pull", s.Now())
	}
	d.mustCast(lightningShield)
	d.wait(2 * time.Second)
	d.mustCast(earthShock)
	if !s.pulled || s.run.damage <= 0 {
		t.Fatalf("Earth Shock after 70 minutes: pulled %v", s.pulled)
	}
}

// Spells cast 9 minutes before the pull have the right aura times.
func TestAuraTimesAtMinusNine(t *testing.T) {
	s := newEnhSession(t, Settings{})
	d := newDriver(t, s)
	d.wait(time.Minute)
	for _, id := range []int32{lightningShield, strengthOfEarth, windfuryTotem, magmaTotem} {
		d.wait(2 * time.Second)
		d.mustCast(id)
	}
	for _, aura := range s.run.char.GetAuras() {
		if !aura.IsActive() || aura.ExpiresAt() == core.NeverExpires {
			continue
		}
		if got := aura.ExpiresAt() - aura.StartedAt(); got != aura.Duration && aura.Duration != core.NeverExpires {
			t.Errorf("%s: up for %s from %s, want %s", aura.Label, got, aura.StartedAt(), aura.Duration)
		}
		if aura.StartedAt() > s.Now() || aura.StartedAt() < -10*time.Minute {
			t.Errorf("%s started at %s", aura.Label, aura.StartedAt())
		}
	}
}

// Free mode: the 60 minute cap ends the fight, and the dummy is never in execute range unless
// the setting says so.
func TestFreeModeCapAndExecute(t *testing.T) {
	s := newEnhSession(t, Settings{})
	d := newDriver(t, s)
	d.standAt(0, -3)
	d.mustCast(earthShock)
	d.wait(30 * time.Minute)
	if s.run.sim.IsExecutePhase20() {
		t.Error("the dummy is in execute range at 30 minutes")
	}
	d.wait(31 * time.Minute)
	if !s.Ended() {
		t.Fatal("the 60 minute cap didn't end the fight")
	}
	if got := s.Summary().DurationMs; got != 3600000 {
		t.Errorf("the capped fight lasted %.0fms", got)
	}

	s = newEnhSession(t, Settings{ExecuteFromStart: true})
	d = newDriver(t, s)
	d.standAt(0, -3)
	d.mustCast(earthShock)
	d.wait(time.Second)
	if !s.run.sim.IsExecutePhase20() {
		t.Error("the Execute range setting didn't put the dummy in execute range")
	}
}

// Encounter mode.
func TestEncounterMode(t *testing.T) {
	t.Run("countdown", func(t *testing.T) {
		s := newEnhSession(t, Settings{Mode: "encounter", LengthMs: 30000, CountdownMs: 5000})
		d := newDriver(t, s)
		d.standAt(0, -3)
		d.startEncounter()
		if d.now != -5*time.Second || s.Now() > d.now {
			t.Fatalf("after Start Encounter the clock is at %s and the sim at %s, want -5s", d.now, s.Now())
		}
		d.wait(4 * time.Second)
		if s.pulled {
			t.Fatal("pulled during the countdown")
		}
		d.wait(time.Second + ms(1))
		if !s.pulled || s.earlyPull {
			t.Fatalf("at the countdown's end: pulled %v early %v", s.pulled, s.earlyPull)
		}
		d.press(Input{Kind: "startAttack"})
		d.wait(31 * time.Second)
		if !s.Ended() {
			t.Fatal("the set length didn't end the fight")
		}
		if got := s.Summary().DurationMs; got != 30000 {
			t.Errorf("the fight lasted %.0fms, want 30000", got)
		}
		if s.batchEnded() {
			t.Error("stopped early on its own")
		}
	})
	t.Run("early pull", func(t *testing.T) {
		s := newEnhSession(t, Settings{Mode: "encounter", LengthMs: 30000, CountdownMs: 5000})
		d := newDriver(t, s)
		d.standAt(0, -3)
		d.startEncounter()
		d.wait(2 * time.Second)
		d.mustCast(earthShock)
		if !s.pulled || !s.earlyPull {
			t.Fatalf("Earth Shock in the countdown: pulled %v early %v", s.pulled, s.earlyPull)
		}
		if s.Now() > ms(5) {
			t.Fatalf("the early pull is at %s, want 0", s.Now())
		}
	})
	t.Run("bolt at the countdown's end", func(t *testing.T) {
		// A first session finds how long a bolt takes from the press to its landing.
		probe := newEnhSession(t, Settings{})
		pd := newDriver(t, probe)
		pd.standAt(0, -25)
		pressed := pd.now
		pd.mustCast(lightningBolt)
		castTime := probe.run.char.Hardcast.Expires - pressed
		travel := probe.run.toolkit[lightningBolt].TravelTime()

		s := newEnhSession(t, Settings{Mode: "encounter", LengthMs: 30000, CountdownMs: 5000})
		d := newDriver(t, s)
		d.standAt(0, -25)
		d.startEncounter()
		d.wait(-d.now - castTime - travel)
		d.mustCast(lightningBolt)
		d.wait(castTime + travel + ms(10))
		if !s.pulled || s.earlyPull {
			t.Fatalf("a bolt landing at 0: pulled %v early %v", s.pulled, s.earlyPull)
		}
		if h := s.run.hits[hitKey{spellID: lightningBolt}]; h == nil {
			t.Fatal("the bolt landing at 0 didn't count")
		}
	})
	t.Run("hit before Start Encounter", func(t *testing.T) {
		s := newEnhSession(t, Settings{Mode: "encounter", LengthMs: 30000, CountdownMs: 5000})
		d := newDriver(t, s)
		d.standAt(0, -3)
		d.mustCast(earthShock)
		if !s.pulled || s.encounter {
			t.Fatalf("a hit before Start Encounter: pulled %v encounter %v", s.pulled, s.encounter)
		}
		if s.fightInfo().Mode != "free" {
			t.Fatal("not a Free mode fight")
		}
	})
	t.Run("stop", func(t *testing.T) {
		s := newEnhSession(t, Settings{Mode: "encounter", LengthMs: 30000, CountdownMs: 5000})
		d := newDriver(t, s)
		d.standAt(0, -3)
		d.startEncounter()
		d.wait(10 * time.Second)
		s.Stop(d.now)
		b := s.TakeBatch()
		if b.Ended == nil || !b.Ended.StoppedEarly {
			t.Fatalf("Stop in Encounter mode: %+v", b.Ended)
		}
		if got := s.Summary().DurationMs; got < 4990 || got > 5010 {
			t.Errorf("stopped at %.0fms, want 5000", got)
		}
	})
}

func (s *Session) batchEnded() bool {
	return s.batch.Ended != nil && s.batch.Ended.StoppedEarly
}

// An input at the fight's end is dropped.
func TestInputAtEndDropped(t *testing.T) {
	s := newEnhSession(t, Settings{Mode: "encounter", LengthMs: 10000, CountdownMs: 0})
	d := newDriver(t, s)
	d.standAt(0, -3)
	d.startEncounter()
	d.wait(10 * time.Second)
	id := d.press(Input{Kind: "cast", SpellID: earthShock})
	b := d.wait(ms(10))
	if a := answerOf(b, id); a == nil || a.OK {
		t.Fatalf("a press at the end: %+v", a)
	}
}
