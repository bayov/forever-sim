//go:build practice

package practice

import (
	"math"
	"time"
)

// The world: where you and the dummy stand, which way you face, and whether you're in range.
//
// Your movement comes in as move inputs, each with where you are, how you move from then on and
// how fast you turn. Between inputs we move you in a straight line. While you move or turn, a
// world tick every 50ms checks whether you reached or left range or turned away, so a swing
// stops or starts within 50ms of the moment, about as often as the game hears from the client.
// The ticks are session actions, so a replay of the same inputs makes the same ticks.

const (
	worldTick = 50 * time.Millisecond

	// Every player race's combat reach, and a training dummy's.
	playerReach = 1.5
	dummyReach  = 1.5
	playerHeight = 2.0
	dummyHeight  = 2.3

	// The game's nominal melee range, and the core's.
	nominalMeleeRange = 5.0

	// A jump keeps you in the air this long.
	jumpTime = 830 * time.Millisecond
)

type vec2 [2]float64

func (v vec2) add(o vec2) vec2         { return vec2{v[0] + o[0], v[1] + o[1]} }
func (v vec2) sub(o vec2) vec2         { return vec2{v[0] - o[0], v[1] - o[1]} }
func (v vec2) scale(f float64) vec2    { return vec2{v[0] * f, v[1] * f} }
func (v vec2) length() float64         { return math.Hypot(v[0], v[1]) }
func (v vec2) isZero() bool            { return v[0] == 0 && v[1] == 0 }
func fromAngle(angle float64) vec2     { return vec2{math.Cos(angle), math.Sin(angle)} }
func (v vec2) dot(o vec2) float64      { return v[0]*o[0] + v[1]*o[1] }
func (v vec2) dist(o vec2) float64     { return v.sub(o).length() }
func durationSeconds(d time.Duration) float64 { return d.Seconds() }

// body is one unit in the world. Its position and facing are known at time `at`, and it moves
// with vel and turns with turnRate from then on.
type body struct {
	pos      vec2
	vel      vec2
	facing   float64
	turnRate float64
	at       time.Duration

	airborneUntil time.Duration
	reach         float64
	height        float64
}

func (b *body) posAt(t time.Duration) vec2 {
	return b.pos.add(b.vel.scale(durationSeconds(t - b.at)))
}

func (b *body) facingAt(t time.Duration) float64 {
	return b.facing + b.turnRate*durationSeconds(t-b.at)
}

// settle moves the body's known point to t.
func (b *body) settle(t time.Duration) {
	b.pos = b.posAt(t)
	b.facing = normalizeAngle(b.facingAt(t))
	b.at = t
}

func (b *body) moving() bool {
	return !b.vel.isZero()
}

func (b *body) airborne(t time.Duration) bool {
	return t < b.airborneUntil
}

func normalizeAngle(a float64) float64 {
	a = math.Mod(a, 2*math.Pi)
	if a < 0 {
		a += 2 * math.Pi
	}
	return a
}

// inArc is whether `to` is within the front arc of a unit at `from` facing `facing`. arc is the
// whole arc, so 180° is the front half.
func inArc(from vec2, facing float64, to vec2, arc float64) bool {
	d := to.sub(from)
	if d.isZero() {
		return true
	}
	angle := math.Atan2(d[1], d[0]) - facing
	angle = math.Mod(angle+3*math.Pi, 2*math.Pi) - math.Pi
	return math.Abs(angle) <= arc/2+1e-9
}

// meleeRange is the larger of 5 yards and both units' combat reach plus 4/3 yards. The game
// adds 8/3 yards when both units move, which never happens with a dummy.
func meleeRange(a, b *body) float64 {
	return math.Max(nominalMeleeRange, a.reach+b.reach+4.0/3.0)
}

// spellRange is a spell's max range plus both units' combat reach, as the engine measures it.
func spellRange(maxRange float64, a, b *body) float64 {
	return maxRange + a.reach + b.reach
}

const (
	spellArc = math.Pi         // spells and abilities need the target in your front 180°
	swingArc = 2 * math.Pi / 3 // auto attack needs it in your front 120°
)
