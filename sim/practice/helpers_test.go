//go:build practice

package practice

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

// The level 60 enh shaman from the DPS Report (Windfury Weapon, Flametongue Totem in the raid
// buffs, the synthetic gear).
func enhRequest(t *testing.T) *proto.RaidSimRequest {
	t.Helper()
	data, err := os.ReadFile("testdata/enh60.json")
	if err != nil {
		t.Fatal(err)
	}
	request, err := ParseRequest(data)
	if err != nil {
		t.Fatal(err)
	}
	request, err = Normalize(request)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func newEnhSession(t *testing.T, settings Settings) *Session {
	t.Helper()
	s, err := NewSession(enhRequest(t), settings, 12345)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// Spell ids of the enh toolkit at 60.
const (
	lightningBolt   = 15208
	earthShock      = 10414
	stormstrike     = 17364
	flameShock      = 29228
	magmaTotem      = 10587
	searingTotem    = 10438
	windfuryTotem   = 10614
	strengthOfEarth = 10442
	stoneskin       = 10408
	lightningShield = 10432
	graceOfAir      = 10627
)

// driver plays a session the way the worker does: inputs at their times, and the sim run up to
// each one.
type driver struct {
	t      *testing.T
	s      *Session
	nextID int
	// now is the driver's clock, in the session's current timeline.
	now time.Duration
}

func newDriver(t *testing.T, s *Session) *driver {
	return &driver{t: t, s: s, now: s.Now(), nextID: 1}
}

// wait moves the clock and runs the sim up to it, following each swap's shift.
func (d *driver) wait(dt time.Duration) []*Batch {
	var batches []*Batch
	target := d.now + dt
	for {
		d.s.Advance(target + 1)
		b := d.s.TakeBatch()
		batches = append(batches, b)
		if b.Rebase != nil {
			target += fromMs(b.Rebase.ShiftMs)
			continue
		}
		break
	}
	d.now = target
	return batches
}

// press sends an input now and runs the sim through it. It returns the input's id.
func (d *driver) press(in Input) int {
	in.ID = d.nextID
	d.nextID++
	d.s.Input(in, d.now)
	return in.ID
}

// answerOf finds an input's answer in batches.
func answerOf(batches []*Batch, id int) *Answer {
	for _, b := range batches {
		for i := range b.Answers {
			if b.Answers[i].InputID == id {
				return &b.Answers[i]
			}
		}
	}
	return nil
}

// cast presses a spell, runs the sim a moment, and returns the answer.
func (d *driver) cast(spellID int32) *Answer {
	d.t.Helper()
	id := d.press(Input{Kind: "cast", SpellID: spellID, Target: &UnitRef{Kind: "target"}})
	batches := d.wait(time.Millisecond + fromMs(d.s.settings.PingMs/2))
	a := answerOf(batches, id)
	if a == nil {
		return &Answer{InputID: id, OK: false, Error: "(no answer yet)"}
	}
	return a
}

func (d *driver) mustCast(spellID int32) {
	d.t.Helper()
	if a := d.cast(spellID); !a.OK {
		d.t.Fatalf("cast %d at %s failed: %q", spellID, d.now, a.Error)
	}
}

// walkTo puts you at a spot facing the dummy, standing still.
func (d *driver) standAt(x, y float64) {
	facing := math.Atan2(-y, -x)
	d.press(Input{Kind: "move", Pos: [2]float64{x, y}, Facing: facing})
	d.wait(time.Millisecond + fromMs(d.s.settings.PingMs/2))
}

func (d *driver) faceAway() {
	r := d.s.run
	p := r.me.posAt(d.s.Now())
	d.press(Input{Kind: "move", Pos: [2]float64(p), Facing: math.Atan2(p[1], p[0])})
	d.wait(time.Millisecond)
}

func (d *driver) startEncounter() {
	d.t.Helper()
	shift, err := d.s.StartEncounter(d.now)
	if err != nil {
		d.t.Fatal(err)
	}
	d.now += shift
	d.s.TakeBatch()
}
