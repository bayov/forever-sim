//go:build practice

package practice

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// Attempt is one saved fight: the request, the seed and every input with the time it applied.
// A replay runs the same session without a clock and gets the same fight.
type Attempt struct {
	Version          int             `json:"version"`
	SimCommit        string          `json:"simCommit,omitempty"`
	Request          json.RawMessage `json:"request"`
	Seed             int64           `json:"seed"`
	PingMs           float64         `json:"pingMs"`
	Mode             string          `json:"mode"`
	SetLengthMs      float64         `json:"setLengthMs"`
	CountdownMs      float64         `json:"countdownMs,omitempty"`
	ExecuteFromStart bool            `json:"executeFromStart"`
	SpellQueueMs     float64         `json:"spellQueueWindowMs"`
	StartNs          int64           `json:"startNs"`
	EndedAtMs        float64         `json:"endedAtMs"`
	StoppedEarly     bool            `json:"stoppedEarly"`
	EarlyPull        bool            `json:"earlyPull"`
	Inputs           []AttemptInput  `json:"inputs"`
	Result           *Summary        `json:"result,omitempty"`
}

type AttemptInput struct {
	AppliedAtNs int64 `json:"appliedAtNs"`
	Seq         int   `json:"seq"`
	Input       Input `json:"input"`
}

// SimCommit is the git commit of the sim build, set by the build.
var SimCommit = ""

// Attempt returns the finished fight as an attempt, nil while it isn't over or never pulled.
func (s *Session) Attempt() *Attempt {
	if !s.ended || s.summary == nil {
		return nil
	}
	request, err := protojson.Marshal(s.request)
	if err != nil {
		return nil
	}
	a := &Attempt{
		Version:          1,
		SimCommit:        SimCommit,
		Request:          request,
		Seed:             s.seed,
		PingMs:           s.settings.PingMs,
		Mode:             "free",
		SetLengthMs:      toMs(freeModeCap),
		ExecuteFromStart: s.settings.ExecuteFromStart,
		SpellQueueMs:     s.settings.SpellQueueWindowMs,
		StartNs:          int64(s.run.start),
		EndedAtMs:        toMs(s.endedAt),
		StoppedEarly:     s.stopped && s.encounter,
		EarlyPull:        s.earlyPull,
		Result:           s.summary,
	}
	if s.encounter {
		a.Mode = "encounter"
		a.SetLengthMs = s.settings.LengthMs
		a.CountdownMs = s.settings.CountdownMs
	}
	for _, rec := range s.inputs {
		a.Inputs = append(a.Inputs, AttemptInput{AppliedAtNs: int64(rec.At), Seq: rec.Seq, Input: rec.Input})
	}
	return a
}

// Replay runs an attempt again from its stored inputs and returns the finished session.
func Replay(a *Attempt) (*Session, error) {
	if a.Version != 1 {
		return nil, errors.New("unknown attempt version")
	}
	request := &proto.RaidSimRequest{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(a.Request, request); err != nil {
		return nil, err
	}
	settings := Settings{
		Mode:               a.Mode,
		LengthMs:           a.SetLengthMs,
		CountdownMs:        a.CountdownMs,
		PingMs:             a.PingMs,
		ExecuteFromStart:   a.ExecuteFromStart,
		SpellQueueWindowMs: a.SpellQueueMs,
	}
	RegisterAll()
	s := &Session{
		request:   request,
		settings:  settings,
		seed:      a.Seed,
		answered:  map[int]bool{},
		encounter: a.Mode == "encounter",
		earlyPull: a.EarlyPull,
		keepLog:   true,
	}
	r, err := s.newRun(time.Duration(a.StartNs))
	if err != nil {
		return nil, err
	}
	s.run = r
	for _, in := range a.Inputs {
		rec := recordedInput{At: time.Duration(in.AppliedAtNs), Seq: in.Seq, Input: in.Input}
		s.inputs = append(s.inputs, rec)
		r.enqueueInput(rec)
		s.nextSeq = max(s.nextSeq, in.Seq+1)
	}
	// The live session marked the pull once every event before 0 had run.
	r.runUntil(0)
	s.markPulled(a.EarlyPull)

	endedAt := fromMs(a.EndedAtMs)
	if a.StoppedEarly || (a.Mode == "free" && endedAt < freeModeCap) {
		s.Advance(endedAt)
		if !s.ended {
			s.stopped = true
			s.end(endedAt, a.StoppedEarly)
		}
	} else {
		s.Advance(endedAt + 1)
	}
	if !s.ended {
		return nil, errors.New("the replay didn't reach the fight's end")
	}
	return s, nil
}
