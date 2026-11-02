//go:build wasm && practice

// practice.wasm: the practice session for the room's worker (docs/practice_room.md). Every entry
// point is a synchronous call that takes and returns JSON strings.
package main

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"syscall/js"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/practice"
	"google.golang.org/protobuf/encoding/protojson"
)

var session *practice.Session

func main() {
	core.SetRunningInWasm()
	practice.RegisterAll()

	js.Global().Set("practiceStart", js.FuncOf(guard(start)))
	js.Global().Set("practiceInput", js.FuncOf(guard(input)))
	js.Global().Set("practiceAdvance", js.FuncOf(guard(advance)))
	js.Global().Set("practiceStartEncounter", js.FuncOf(guard(startEncounter)))
	js.Global().Set("practiceStop", js.FuncOf(guard(stop)))
	js.Global().Set("practiceAttempt", js.FuncOf(guard(attempt)))
	js.Global().Set("practicePrepare", js.FuncOf(guard(prepare)))
	js.Global().Call("wasmready")
	select {}
}

// guard turns a panic into an error reply, so one bad input doesn't kill the worker.
func guard(fn func(args []js.Value) (any, error)) func(js.Value, []js.Value) any {
	return func(_ js.Value, args []js.Value) (reply any) {
		defer func() {
			if p := recover(); p != nil {
				reply = errorReply(fmt.Errorf("%v\n%s", p, debug.Stack()))
			}
		}()
		out, err := fn(args)
		if err != nil {
			return errorReply(err)
		}
		data, err := json.Marshal(out)
		if err != nil {
			return errorReply(err)
		}
		return string(data)
	}
}

func errorReply(err error) string {
	data, _ := json.Marshal(map[string]string{"error": err.Error()})
	return string(data)
}

type startArgs struct {
	Request  json.RawMessage   `json:"request"`
	Settings practice.Settings `json:"settings"`
	Seed     int64             `json:"seed"`
}

type startReply struct {
	Toolkit []practice.Spell `json:"toolkit"`
	StartMs float64          `json:"startMs"`
	Batch   *practice.Batch  `json:"batch"`
}

func start(args []js.Value) (any, error) {
	var a startArgs
	if err := json.Unmarshal([]byte(args[0].String()), &a); err != nil {
		return nil, err
	}
	request := &proto.RaidSimRequest{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(a.Request, request); err != nil {
		return nil, err
	}
	s, err := practice.NewSession(request, a.Settings, a.Seed)
	if err != nil {
		return nil, err
	}
	session = s
	return startReply{Toolkit: s.Toolkit(), StartMs: ms(s.Now()), Batch: s.TakeBatch()}, nil
}

// input takes an input and the sim time of its press in ms, runs the sim to nowMs, and returns
// the batch.
func input(args []js.Value) (any, error) {
	if session == nil {
		return nil, fmt.Errorf("no session")
	}
	var in practice.Input
	if err := json.Unmarshal([]byte(args[0].String()), &in); err != nil {
		return nil, err
	}
	session.Input(in, fromMs(args[1].Float()))
	return advanceTo(args[2].Float()), nil
}

type advanceReply struct {
	Batch       *practice.Batch `json:"batch"`
	NextEventMs float64         `json:"nextEventMs"`
}

func advanceTo(nowMs float64) advanceReply {
	session.Advance(fromMs(nowMs))
	next := session.NextEventTime()
	reply := advanceReply{Batch: session.TakeBatch(), NextEventMs: -1}
	if next != core.NeverExpires {
		reply.NextEventMs = ms(next)
	}
	return reply
}

func advance(args []js.Value) (any, error) {
	if session == nil {
		return nil, fmt.Errorf("no session")
	}
	return advanceTo(args[0].Float()), nil
}

func startEncounter(args []js.Value) (any, error) {
	if session == nil {
		return nil, fmt.Errorf("no session")
	}
	shift, err := session.StartEncounter(fromMs(args[0].Float()))
	if err != nil {
		return nil, err
	}
	return advanceTo(args[0].Float() + ms(shift)), nil
}

func stop(args []js.Value) (any, error) {
	if session == nil {
		return nil, fmt.Errorf("no session")
	}
	session.Stop(fromMs(args[0].Float()))
	return advanceReply{Batch: session.TakeBatch(), NextEventMs: -1}, nil
}

// prepare builds the next swap's sim, in the worker's idle time.
func prepare(args []js.Value) (any, error) {
	if session != nil {
		session.PrepareSpare()
	}
	return map[string]bool{"ok": true}, nil
}

func attempt(args []js.Value) (any, error) {
	if session == nil {
		return nil, fmt.Errorf("no session")
	}
	return session.Attempt(), nil
}

func ms(d time.Duration) float64 { return float64(d) / float64(time.Millisecond) }

func fromMs(v float64) time.Duration { return time.Duration(v * float64(time.Millisecond)) }
