//go:build practice

package practice

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/shaman"
	goproto "google.golang.org/protobuf/proto"
)

const (
	// The prepull window: a fresh sim starts this long before the pull.
	prepullWindow = 10 * time.Minute
	// A Free mode fight ends here if you never click Stop.
	freeModeCap = 60 * time.Minute
	// When the prepull window runs down to this, we swap to a new one.
	rewindowAt = -30 * time.Second
	// No input is replayed earlier than this, because before minus 60 minutes every cooldown and
	// the GCD look busy (startingCDTime in the core).
	earliestInput = -50 * time.Minute

	defaultSpellQueueWindow = 400 * time.Millisecond

	// The session queue's one action in the sim runs after everything the core does at the same
	// time (ActionPriorityLow is -1).
	sessionPriority core.ActionPriority = -100
	// The check at a cast's end runs before the cast lands.
	castEndCheckPriority core.ActionPriority = 100

	// The party cooldowns of other players are checked this often.
	externalsPeriod = 500 * time.Millisecond
)

// Phases of a fight.
const (
	phasePrepull   = "prepull"
	phaseCountdown = "countdown"
	phaseFight     = "fight"
	phaseEnded     = "ended"
)

var registerOnce sync.Once

// RegisterAll registers every class, once.
func RegisterAll() {
	registerOnce.Do(sim.RegisterAll)
}

// recordedInput is an input as the session queue took it: the time it applies, in the current
// sim's timeline, and its place in arrival order.
type recordedInput struct {
	At    time.Duration
	Seq   int
	Input Input
}

// Session is one attempt: a sim run in real time, swapped for a fresh one at the pull.
type Session struct {
	request  *proto.RaidSimRequest
	settings Settings
	seed     int64

	run *run

	// Every input so far, in the current sim's timeline.
	inputs  []recordedInput
	nextSeq int

	// encounter is whether Start Encounter was clicked: the current sim's 0 is the pull.
	encounter bool
	pulled    bool
	earlyPull bool
	ended     bool
	stopped   bool
	endedAt   time.Duration
	summary   *Summary

	lastShift time.Duration

	// spare is the next swap's sim, built ahead of time (PrepareSpare).
	spare *run

	// The HUD's last state, from just before the fight's end.
	finalPlayer *PlayerState
	finalTarget *TargetState

	// The total shift of every swap so far. A time in the first sim's timeline plus it is the
	// same moment in the current one.
	totalShift time.Duration

	// keepLog keeps every log line of the current sim, for replay checks.
	keepLog bool

	// Failed is the sim's panic, when one ended the attempt.
	Failed string

	answered map[int]bool
	batch    Batch
	logLines []string
}

// NewSession builds the sim from a request the server already normalized (Normalize), and
// starts it at the start of the prepull window.
func NewSession(request *proto.RaidSimRequest, settings Settings, seed int64) (*Session, error) {
	if seed == 0 {
		return nil, errors.New("the seed must not be 0")
	}
	if settings.Mode == "" {
		settings.Mode = "free"
	}
	if settings.Mode != "free" && settings.Mode != "encounter" {
		return nil, fmt.Errorf("unknown mode %q", settings.Mode)
	}
	if settings.LengthMs <= 0 {
		settings.LengthMs = request.GetEncounter().GetDuration() * 1000
	}
	if settings.LengthMs <= 0 {
		settings.LengthMs = 180000
	}
	if settings.SpellQueueWindowMs <= 0 {
		settings.SpellQueueWindowMs = float64(defaultSpellQueueWindow / time.Millisecond)
	}
	RegisterAll()

	s := &Session{
		request:  request,
		settings: settings,
		seed:     seed,
		answered: map[int]bool{},
	}
	r, err := s.newRun(-prepullWindow)
	if err != nil {
		return nil, err
	}
	s.run = r
	return s, nil
}

// Settings returns the session's fight settings.
func (s *Session) Settings() Settings { return s.settings }

// Seed is the seed every sim of the session uses.
func (s *Session) Seed() int64 { return s.seed }

// Now is the current sim time: the time of the last event the sim ran.
func (s *Session) Now() time.Duration { return s.run.sim.CurrentTime }

// Start is where the current sim's clock started.
func (s *Session) Start() time.Duration { return s.run.start }

// Ended is whether the fight is over.
func (s *Session) Ended() bool { return s.ended }

// NextEventTime is the time of the next thing the session or the sim will do, so the worker
// can sleep until then.
func (s *Session) NextEventTime() time.Duration {
	if s.ended {
		return core.NeverExpires
	}
	next := s.run.sim.NextEventTime()
	if !s.pulled && !s.encounter {
		next = min(next, rewindowAt)
	}
	if s.encounter && !s.pulled {
		next = min(next, 0)
	}
	if s.pulled {
		next = min(next, s.run.sim.Duration)
	}
	return next
}

// Input puts an input in the session queue. pressedAt is the press, in the current sim's
// timeline. The input applies half a ping later, as the server hears it in game. An input the
// sim has already passed applies now.
func (s *Session) Input(in Input, pressedAt time.Duration) {
	s.InputApplied(in, pressedAt+fromMs(s.settings.PingMs/2))
}

// InputApplied puts an input in the session queue at the time it applies, for replays.
func (s *Session) InputApplied(in Input, at time.Duration) {
	if s.ended || (s.pulled && at >= s.run.sim.Duration) {
		s.answer(in.ID, false, "")
		return
	}
	at = max(at, s.run.sim.CurrentTime)
	rec := recordedInput{At: at, Seq: s.nextSeq, Input: in}
	s.nextSeq++
	s.inputs = append(s.inputs, rec)
	s.run.enqueueInput(rec)
}

// Advance runs every event before until, and stops. It never moves the clock to until itself,
// because settling regen and auras at frame times would make results depend on the frame rate.
//
// A panic in the sim ends the attempt with the error, because the sim's state can't be trusted
// after it. It comes from a spell the sim doesn't model for this spec (see run.safeCast).
func (s *Session) Advance(until time.Duration) {
	defer func() {
		if p := recover(); p != nil {
			s.Failed = fmt.Sprint(p)
			s.ended = true
			s.batch.Errors = append(s.batch.Errors, "The sim failed: "+s.Failed)
			s.batch.Ended = &Ended{StoppedEarly: true}
		}
	}()
	for !s.ended {
		r := s.run

		// Encounter mode: the pull is at the countdown's end.
		if s.encounter && !s.pulled && until > 0 && r.sim.NextEventTime() >= 0 {
			s.markPulled(false)
		}

		// The fight's end: Stop, the set length, or the Free mode cap.
		if s.pulled && until > r.sim.Duration && r.sim.NextEventTime() > r.sim.Duration {
			s.end(r.sim.Duration, false)
			return
		}

		// Before the pull in Free mode, sim time 0 must never come. We swap to a new window first.
		if !s.pulled && !s.encounter && until > rewindowAt && r.sim.NextEventTime() >= rewindowAt {
			s.rewindow(rewindowAt)
			until += s.lastShift
			continue
		}

		next := r.sim.NextEventTime()
		if next >= until {
			return
		}
		if s.pulled && next > r.sim.Duration {
			continue
		}
		r.sim.Step()

		if r.pullAt != nil {
			at := *r.pullAt
			r.pullAt = nil
			if at < 0 {
				early := s.encounter
				s.swap(-at, r.sim.CurrentTime)
				until += s.lastShift
				s.earlyPull = early
				s.markPulled(early)
			}
		}
	}
}

// StartEncounter starts the countdown: the pull comes when it runs out. The session swaps at
// once, so the pull is at the new sim's 0.
//
// now is the session's current time (the clock, not the last event), and the pull comes a
// countdown after it. It returns the swap's shift, for the caller's clock.
func (s *Session) StartEncounter(now time.Duration) (time.Duration, error) {
	if s.ended || s.pulled {
		return 0, errors.New("the fight already started")
	}
	if s.encounter {
		return 0, errors.New("the countdown already runs")
	}
	now = max(now, s.run.sim.CurrentTime)
	s.Advance(now)
	if s.pulled {
		return 0, errors.New("the fight already started")
	}
	s.encounter = true
	countdown := fromMs(s.settings.CountdownMs)
	shift := -(now + countdown)
	s.swap(shift, now)
	return shift, nil
}

// KeepLog keeps every log line of the sim from now on, for FullLog.
func (s *Session) KeepLog() { s.keepLog = true }

// FullLog is every log line of the current sim, when KeepLog is on.
func (s *Session) FullLog() []string { return s.run.fullLog }

// Stop ends the fight at the given time. Before the pull it ends the attempt with no summary.
func (s *Session) Stop(at time.Duration) {
	if s.ended {
		return
	}
	s.stopped = true
	if !s.pulled {
		s.ended = true
		s.batch.Ended = &Ended{StoppedEarly: true}
		return
	}
	at = max(at, s.run.sim.CurrentTime, 0)
	at = min(at, s.run.sim.Duration)
	s.Advance(at)
	if !s.ended {
		s.end(at, s.encounter)
	}
}

func (s *Session) markPulled(early bool) {
	if s.pulled {
		return
	}
	s.pulled = true
	s.batch.Pulled = &Pulled{Early: early}
	s.run.onPulled()
}

func (s *Session) end(at time.Duration, stoppedEarly bool) {
	r := s.run
	// The HUD keeps the state at the end: Cleanup takes auras off, which would leave more mana
	// than the max.
	s.finalPlayer, s.finalTarget = r.playerState(), r.targetState()
	r.sim.Duration = at
	r.sim.Cleanup()
	s.ended = true
	s.endedAt = at
	s.summary = r.summary(at)
	s.batch.Ended = &Ended{Summary: s.summary, StoppedEarly: stoppedEarly}
}

// lastShift is the shift of the last swap, for Advance to move its until.
func (s *Session) shiftInputs(shift time.Duration) {
	for i := range s.inputs {
		s.inputs[i].At += shift
	}
}

// swap builds a fresh sim whose timeline is the current one moved by shift, feeds it the inputs
// so far, and runs it to the moment the current one is at (now, in the old timeline). The same
// seed and the same inputs in the same order replay exactly.
func (s *Session) swap(shift time.Duration, now time.Duration) {
	s.lastShift = shift
	s.totalShift += shift
	s.shiftInputs(shift)

	old := s.run
	r, err := s.newRun(old.start + shift)
	if err != nil {
		panic(err) // The same request built before.
	}
	r.quiet = true
	for _, rec := range s.inputs {
		r.enqueueInput(rec)
	}
	s.run = r
	r.runUntil(now + shift)
	r.quiet = false
	if testOnSwap != nil {
		testOnSwap(old, r, shift)
	}

	// What the old sim reported in this batch is gone with it.
	s.batch.Hits = nil
	s.batch.Casts = nil
	s.logLines = nil
	prev := 0.0
	if s.batch.Rebase != nil {
		prev = s.batch.Rebase.ShiftMs
	}
	s.batch.Rebase = &Rebase{ShiftMs: prev + toMs(shift)}
}

// rewindow moves the prepull window back by 10 minutes, dropping inputs older than 40 minutes.
// What they left behind (where you stand, your target, auto attack) is put back at the start.
func (s *Session) rewindow(now time.Duration) {
	shift := -prepullWindow - now
	newStart := s.run.start + shift
	cut := max(earliestInput, newStart) - shift // in the current timeline

	var kept []recordedInput
	var lastMove, lastTarget, lastAttack, lastImbue *recordedInput
	for i := range s.inputs {
		rec := s.inputs[i]
		if rec.At >= cut {
			kept = append(kept, rec)
			continue
		}
		switch rec.Input.Kind {
		case "move":
			lastMove = &s.inputs[i]
		case "target", "clearTarget":
			lastTarget = &s.inputs[i]
		case "startAttack", "stopAttack":
			lastAttack = &s.inputs[i]
		case "cast", "setImbue":
			if s.run.isImbueCast(rec.Input.SpellID) {
				lastImbue = &s.inputs[i]
			}
		}
	}
	var lasting []recordedInput
	for _, rec := range []*recordedInput{lastMove, lastTarget, lastAttack, lastImbue} {
		if rec == nil {
			continue
		}
		in := rec.Input
		if rec == lastImbue {
			// The imbue goes back on with no cost and no GCD.
			in = Input{Kind: "setImbue", SpellID: in.SpellID}
		}
		if in.Kind == "move" {
			b := body{pos: in.Pos, vel: in.Velocity, facing: in.Facing, turnRate: in.TurnRate, at: rec.At}
			in.Pos = b.posAt(cut)
			in.Facing = normalizeAngle(b.facingAt(cut))
		}
		in.ID = -1
		lasting = append(lasting, recordedInput{At: cut, Seq: rec.Seq, Input: in})
	}
	sort.SliceStable(lasting, func(i, j int) bool { return lasting[i].Seq < lasting[j].Seq })
	s.inputs = append(lasting, kept...)
	s.swap(shift, now)
}

func (s *Session) answer(id int, ok bool, err string) {
	if id < 0 || s.answered[id] {
		return
	}
	s.answered[id] = true
	s.batch.Answers = append(s.batch.Answers, Answer{InputID: id, OK: ok, Error: err})
	if !ok && err != "" {
		s.batch.Errors = append(s.batch.Errors, err)
	}
}

// TakeBatch returns what happened since the last batch, with the state the HUD draws.
func (s *Session) TakeBatch() *Batch {
	b := s.batch
	s.batch = Batch{}
	r := s.run
	b.SimTime = toMs(r.sim.CurrentTime)
	b.LogLines = s.logLines
	s.logLines = nil
	b.Units = r.unitStates()
	switch {
	case !s.ended:
		b.Player = r.playerState()
		b.Target = r.targetState()
	case s.finalPlayer != nil:
		b.Player, b.Target = s.finalPlayer, s.finalTarget
	}
	b.Fight = s.fightInfo()
	return &b
}

func (s *Session) fightInfo() FightInfo {
	info := FightInfo{Mode: "free", LengthMs: toMs(freeModeCap), EarlyPull: s.earlyPull}
	if s.encounter {
		info.Mode = "encounter"
		info.LengthMs = s.settings.LengthMs
	}
	switch {
	case s.ended:
		info.Phase = phaseEnded
		info.EndedAtMs = toMs(s.endedAt)
	case s.pulled:
		info.Phase = phaseFight
	case s.encounter:
		info.Phase = phaseCountdown
	default:
		info.Phase = phasePrepull
	}
	return info
}

// Summary returns the fight summary once the fight ended.
func (s *Session) Summary() *Summary { return s.summary }

// Inputs returns every input so far, in the current sim's timeline.
func (s *Session) Inputs() []recordedInput { return s.inputs }

func toMs(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

func fromMs(ms float64) time.Duration {
	return time.Duration(math.Round(ms * float64(time.Millisecond)))
}

// testOnSwap lets a test look at both sims of a swap.
var testOnSwap func(old, new *run, shift time.Duration)

// The run being built, for the environment hook. Building a sim is quick, and sims are built
// one at a time.
var (
	buildMu  sync.Mutex
	building *run
)

func init() {
	core.EnvHooks = append(core.EnvHooks, func(env *core.Environment) {
		if building != nil {
			building.onEnvironment(env)
		}
	})
}

// PrepareSpare builds the next swap's sim ahead of time. Building one takes a few ms in WASM, so
// the worker calls this in idle time, and the pull's swap only replays the inputs.
func (s *Session) PrepareSpare() {
	if s.ended || (s.spare != nil && s.spare.encounter == s.encounter) {
		return
	}
	if r, err := s.buildRun(); err == nil {
		s.spare = r
	}
}

// newRun gives a fresh sim whose clock starts at start: the spare when it fits the mode.
func (s *Session) newRun(start time.Duration) (*run, error) {
	r := s.spare
	s.spare = nil
	if r == nil || r.encounter != s.encounter {
		var err error
		if r, err = s.buildRun(); err != nil {
			return nil, err
		}
	}
	r.start = start
	r.setup()
	return r, nil
}

// buildRun builds a sim for the current mode, not yet set up.
func (s *Session) buildRun() (*run, error) {
	request := goproto.Clone(s.request).(*proto.RaidSimRequest)
	request.SimOptions.RandomSeed = s.seed
	enc := request.Encounter
	if s.encounter {
		enc.Duration = s.settings.LengthMs / 1000
	} else {
		// Free mode: the dummy never enters execute range, unless the setting puts it there
		// from the start.
		enc.Duration = freeModeCap.Seconds()
		p := 0.0
		if s.settings.ExecuteFromStart {
			p = 1
		}
		enc.ExecuteProportion_20, enc.ExecuteProportion_25, enc.ExecuteProportion_35 = p, p, p
	}

	r := &run{s: s, encounter: s.encounter, items: map[int32]*core.Spell{}, toolkit: map[int32]*core.Spell{}}

	var simulation *core.Simulation
	err := func() (err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("the sim couldn't be built: %v", p)
			}
		}()
		buildMu.Lock()
		building = r
		defer func() {
			building = nil
			buildMu.Unlock()
		}()
		simulation = core.NewSim(request, simsignals.Signals{})
		return nil
	}()
	if err != nil {
		return nil, err
	}
	r.sim = simulation
	return r, nil
}

// Errors the game shows, from GlobalStrings.
const (
	errMoving        = "Can't do that while moving"
	errFacing        = "Target needs to be in front of you."
	errOutOfRange    = "Out of range."
	errNotReady      = "Spell is not ready yet."
	errAbilityReady  = "Ability is not ready yet."
	errNoTarget      = "You have no target."
	errNothingAttack = "There is nothing to attack."
	errItemNotReady  = "Item is not ready yet."
	errUnknownSpell  = "Unknown spell"
	errCantDoThat    = "Can't do that yet"
	errInvalidTarget = "Invalid target"
)

// manaError turns the core's cost failure into the game's text.
func costError(reason string) string {
	switch {
	case strings.Contains(strings.ToLower(reason), "energy"):
		return "Not enough energy"
	case strings.Contains(strings.ToLower(reason), "rage"):
		return "Not enough rage"
	default:
		return "Not enough mana"
	}
}

var _ = shaman.TalentTreeSizes
