//go:build practice

package practice

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"
)

// playFight plays a short fight with prepull casts, movement, auto attack and a queued cast.
func playFight(t *testing.T, s *Session, encounter bool) {
	d := newDriver(t, s)
	d.mustCast(lightningShield)
	d.wait(1700 * time.Millisecond)
	d.mustCast(strengthOfEarth)
	d.wait(ms(1100))
	d.mustCast(windfuryTotem)
	d.wait(3 * time.Second)
	d.press(Input{Kind: "move", Pos: [2]float64{0, -15}, Velocity: [2]float64{0, 7}, Facing: math.Pi / 2})
	d.wait(ms(1700))
	d.standAt(0, -3)
	if encounter {
		d.startEncounter()
		d.wait(2 * time.Second)
		d.mustCast(magmaTotem)
		d.wait(3 * time.Second)
	} else {
		d.mustCast(magmaTotem)
		d.wait(3 * time.Second)
	}
	d.press(Input{Kind: "startAttack"})
	for i := 0; i < 8; i++ {
		d.wait(ms(1300))
		d.press(Input{Kind: "cast", SpellID: stormstrike})
		d.wait(ms(1300))
		d.press(Input{Kind: "cast", SpellID: earthShock})
	}
	d.press(Input{Kind: "move", Pos: [2]float64{0, -3}, Velocity: [2]float64{0, -7}, Facing: math.Pi / 2})
	d.wait(2 * time.Second)
	d.standAt(0, -3)
	d.wait(5 * time.Second)
	if !encounter {
		s.Stop(d.now)
	} else {
		d.wait(60 * time.Second)
	}
	if !s.Ended() {
		t.Fatal("the fight didn't end")
	}
}

func TestReplayMatchesLiveFight(t *testing.T) {
	for _, mode := range []string{"free", "encounter"} {
		t.Run(mode, func(t *testing.T) {
			s := newEnhSession(t, Settings{Mode: mode, LengthMs: 30000, CountdownMs: 5000, PingMs: 60})
			s.KeepLog()
			playFight(t, s, mode == "encounter")
			live := s.FullLog()
			attempt := s.Attempt()
			if attempt == nil {
				t.Fatal("no attempt")
			}

			// Through JSON, as the server stores it.
			data, err := json.Marshal(attempt)
			if err != nil {
				t.Fatal(err)
			}
			var loaded Attempt
			if err := json.Unmarshal(data, &loaded); err != nil {
				t.Fatal(err)
			}
			replay, err := Replay(&loaded)
			if err != nil {
				t.Fatal(err)
			}
			got := replay.FullLog()
			if len(live) < 100 {
				t.Fatalf("the live log has only %d lines", len(live))
			}
			if strings.Join(got, "\n") != strings.Join(live, "\n") {
				for i := 0; i < min(len(got), len(live)); i++ {
					if got[i] != live[i] {
						t.Fatalf("line %d differs:\nlive   %s\nreplay %s", i, live[i], got[i])
					}
				}
				t.Fatalf("the replay has %d lines, the live fight %d", len(got), len(live))
			}
			if replay.Summary().Damage != s.Summary().Damage {
				t.Errorf("replay damage %.0f, live %.0f", replay.Summary().Damage, s.Summary().Damage)
			}
			t.Logf("%s: %d log lines, %.1f DPS over %.1fs", mode, len(live), s.Summary().DPS, s.Summary().DurationMs/1000)
		})
	}
}

// A long prepull with window swaps replays the same.
func TestReplayAfterRewindow(t *testing.T) {
	s := newEnhSession(t, Settings{})
	s.KeepLog()
	d := newDriver(t, s)
	d.standAt(0, -3)
	d.mustCast(lightningShield)
	for i := 0; i < 25; i++ {
		d.wait(time.Minute)
	}
	d.mustCast(earthShock)
	d.press(Input{Kind: "startAttack"})
	d.wait(10 * time.Second)
	s.Stop(d.now)
	replay, err := Replay(s.Attempt())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(replay.FullLog(), "\n") != strings.Join(s.FullLog(), "\n") {
		t.Fatalf("the replay differs: %d lines against %d", len(replay.FullLog()), len(s.FullLog()))
	}
	_ = time.Second
}
