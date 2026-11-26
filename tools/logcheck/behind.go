package main

import (
	"math"
	"sort"
	"strconv"
	"time"
)

// Where a unit's facing sits in its advanced fields, counted from infoStart. It's in radians,
// on the same axes as the position: 0 faces along +x, and it turns toward +y.
//
// We checked that against the swings mobs land on us, since a mob faces the one it swings at.
// On Nov 17, the angle between a mob's logged facing and the direction from it to us was under
// 45 degrees on 650 of 873 swings, with a median of 12 degrees. Reading the angle any other way
// (x and y swapped, or y flipped) puts the median near 90 degrees.
const infoFacing = 17

// We look for a unit's position this far from the moment we need it. We move a few yards a
// second, so an older one could put us on the wrong side of the mob.
const positionSlack = time.Second

// A mob faces the one it swings at. We take its swing as the sign of who that is for this long,
// which is longer than any mob's swing timer we've seen.
const victimHold = 3 * time.Second

// unitPoint is where a unit stood and which way it faced at one moment.
type unitPoint struct {
	t         time.Time
	x, y      float64
	facing    float64
	hasFacing bool
}

// meleeSide counts our melee attempts on mobs of one level by which side of the mob we stood on.
type meleeSide struct {
	attempts int
	// known are the attempts where the log had both our position and the mob's, and behind are
	// those of them where we stood behind the mob.
	known, behind int
}

// share is the part of our melee attempts we made from behind the mob. With no positions we
// say 0, so we stay in front of the mob as when questing.
func (s meleeSide) share() float64 {
	if s.known == 0 {
		return 0
	}
	return float64(s.behind) / float64(s.known)
}

// isBehind says whether a unit at (x, y) stands behind a mob at (mobX, mobY) that faces
// facing.
//
// Behind is outside the mob's front 180 degrees, where it can't parry or block our attacks.
func isBehind(mobX, mobY, facing, x, y float64) bool {
	toUs := math.Atan2(y-mobY, x-mobX)
	return math.Abs(math.Remainder(toUs-facing, 2*math.Pi)) > math.Pi/2
}

// meleeSides finds, for each of our melee attempts in the window, whether we stood behind the
// mob, and counts them by mob level.
//
// We used to sim us in front of the mob all the time, where it can parry and block. But in
// dungeons a tank holds the mob and we hit it from behind. Summed over Nov 14 to 25 at level 60 mobs
// (2059 melee attempts), the log had 26 parries where the sim expected 101, and 30 blocks where
// it expected 98, while dodge, crit and glance matched.
//
// Each line with advanced fields says where one unit stood and which way it faced. Which unit
// that is depends on the line. On our SWING_DAMAGE it's us, and on the SWING_DAMAGE_LANDED
// that follows it's the mob. On our SPELL_DAMAGE (Stormstrike, a Windfury attack) it's the mob.
// A mob's swing at us gives the mob. SWING_MISSED and SPELL_MISSED have no advanced fields. So
// we keep each unit's track from every line where the info unit is the line's source or
// target, and look up both of us near the moment of the attack.
//
// The logged facing of a mob lags behind. In the Nov 22 dungeon log, a mob's logged facing
// pointed within 90 degrees of us on only 639 of the 871 swings it landed on us (73%). When we
// took the side from the logged facing, 27 of the log's 96 parries and 22 of its 63 blocks
// came from attempts we put behind the mob. But a mob does face the one it swings at. So when
// the mob swung at someone in the last 3 seconds and we know where they stood, we take the
// direction to them as its facing, and we use the logged facing only for the rest. That leaves
// 8 of the 96 parries and 6 of the 63 blocks behind the mob, out of 3998 attempts there.
func (cl *combatLog) meleeSides() map[int]*meleeSide {
	sf := newSideFinder(cl)
	sides := map[int]*meleeSide{}
	for _, e := range cl.events {
		if e.srcGUID != cl.me || !isCreature(e.dstGUID) || !isMelee(e) {
			continue
		}
		level := cl.levels[e.dstGUID]
		if level == 0 {
			continue
		}
		s := sides[level]
		if s == nil {
			s = &meleeSide{}
			sides[level] = s
		}
		s.attempts++
		if behind, known := sf.behind(e); known {
			s.known++
			if behind {
				s.behind++
			}
		}
	}
	return sides
}

// sideFinder knows where each unit stood in the window and who each mob swung at.
type sideFinder struct {
	me     string
	tracks map[string][]unitPoint
	// victims are the units each mob swung at, in time order.
	victims map[string][]mobSwing
}

type mobSwing struct {
	t      time.Time
	target string
}

func newSideFinder(cl *combatLog) *sideFinder {
	sf := &sideFinder{me: cl.me, tracks: map[string][]unitPoint{}, victims: map[string][]mobSwing{}}
	for _, e := range cl.events {
		if (e.kind == "SWING_DAMAGE" || e.kind == "SWING_MISSED") && isCreature(e.srcGUID) {
			sf.victims[e.srcGUID] = append(sf.victims[e.srcGUID], mobSwing{e.t, e.dstGUID})
		}
		// A line of another layout (SPELL_ABSORBED, SPELL_CAST_FAILED) has no unit where the
		// advanced fields would start, so we check the unit is one the line is about.
		guid := e.infoGUID()
		if guid == "" || (guid != e.srcGUID && guid != e.dstGUID) {
			continue
		}
		x, y, ok := e.position()
		if !ok || (x == 0 && y == 0) {
			continue
		}
		p := unitPoint{t: e.t, x: x, y: y}
		if facing, err := strconv.ParseFloat(e.f[e.infoStart()+infoFacing], 64); err == nil {
			p.facing, p.hasFacing = facing, true
		}
		sf.tracks[guid] = append(sf.tracks[guid], p)
	}
	return sf
}

// behind says whether we stood behind the mob at our attack e, and known whether the log had
// what it takes to tell.
func (sf *sideFinder) behind(e *event) (behind, known bool) {
	mob, okMob := nearestPoint(sf.tracks[e.dstGUID], e.t, positionSlack)
	us, okUs := nearestPoint(sf.tracks[sf.me], e.t, positionSlack)
	if !okMob || !okUs {
		return false, false
	}
	facing, ok := sf.facing(e.dstGUID, mob, e.t)
	if !ok {
		return false, false
	}
	return isBehind(mob.x, mob.y, facing, us.x, us.y), true
}

// facing is which way the mob at mobAt faced at t: toward the unit it last swung at, or else
// its logged facing.
func (sf *sideFinder) facing(mob string, mobAt unitPoint, t time.Time) (float64, bool) {
	swings := sf.victims[mob]
	i := sort.Search(len(swings), func(i int) bool { return swings[i].t.After(t) })
	if i > 0 && t.Sub(swings[i-1].t) <= victimHold {
		if v, ok := nearestPoint(sf.tracks[swings[i-1].target], t, positionSlack); ok && (v.x != mobAt.x || v.y != mobAt.y) {
			return math.Atan2(v.y-mobAt.y, v.x-mobAt.x), true
		}
	}
	return mobAt.facing, mobAt.hasFacing
}

// isMelee says whether the line is one of our melee attempts that a mob in front of us could
// parry or block: a white swing, or a Physical spell like Stormstrike or a Windfury attack.
// An evade is left out, as in the report's rows.
func isMelee(e *event) bool {
	switch e.kind {
	case "SWING_DAMAGE":
		return true
	case "SWING_MISSED":
		return e.missType() != "EVADE"
	case "SPELL_DAMAGE":
		return len(e.f) > 11 && e.f[11] == "0x1"
	case "SPELL_MISSED":
		return len(e.f) > 11 && e.f[11] == "0x1" && e.missType() != "EVADE"
	}
	return false
}

// nearestPoint is the point of a track nearest to t, within slack of it.
func nearestPoint(track []unitPoint, t time.Time, slack time.Duration) (unitPoint, bool) {
	i := sort.Search(len(track), func(i int) bool { return track[i].t.After(t) })
	best, found := unitPoint{}, false
	bestGap := slack + 1
	for _, j := range []int{i - 1, i} {
		if j < 0 || j >= len(track) {
			continue
		}
		gap := track[j].t.Sub(t)
		if gap < 0 {
			gap = -gap
		}
		if gap <= slack && gap < bestGap {
			best, bestGap, found = track[j], gap, true
		}
	}
	return best, found
}
