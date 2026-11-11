package main

import (
	"slices"
	"testing"
	"time"
)

// Well Fed (3 Strength) never shows in the log, so only the export or BayovCore's buff records
// know about it.
const wellFed = 1248422

func TestWindowBuffsTimeline(t *testing.T) {
	base := time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC)
	at := func(sec float64) time.Time { return base.Add(seconds(sec)) }
	const me = "Player-1-0001"
	all := []*event{
		{t: at(-120), kind: "SWING_DAMAGE", srcGUID: me},
		{t: at(110), kind: "UNIT_DIED", dstGUID: me},
	}
	cl := &combatLog{me: me, from: at(0), to: at(120)}
	tables := &clientTables{spellNames: map[int32]string{wellFed: "Well Fed"}}
	steps := func(ex *export) []buffStep {
		return windowBuffs(all, cl, ex, tables, map[int32]bool{})
	}
	want := func(name string, got []buffStep, want ...buffStep) {
		t.Helper()
		if !slices.EqualFunc(got, want, func(a, b buffStep) bool { return a.at == b.at && slices.Equal(a.ids, b.ids) }) {
			t.Errorf("%s: got %v, want %v", name, got, want)
		}
	}
	guess := []byte(`[1248422]`)

	// Without buff records, the export's Well Fed is up for the whole window, until the death.
	want("export only", steps(&export{Buffs: guess}),
		buffStep{0, []int32{wellFed}})

	// A full record before the window says Well Fed has 90 seconds left a minute before it, so
	// it runs out 30 seconds in. The export's guess no longer counts.
	timeline := &buffTimeline{records: []buffRecord{
		{t: at(-60), full: true, gained: []buffGain{{id: wellFed, dur: seconds(900), left: seconds(90)}}},
	}}
	want("ran out", steps(&export{Buffs: guess, buffLog: timeline}),
		buffStep{0, []int32{wellFed}}, buffStep{seconds(30), nil})

	// We ate again in combat. The record after it came at 100 seconds, after 40 seconds of the
	// game hiding our buffs, and the buff had 880 of its 900 seconds left. So we ate at 80
	// seconds, and the death at 110 took it off.
	timeline.records = append(timeline.records, buffRecord{
		t: at(100), hidden: seconds(40), gained: []buffGain{{id: wellFed, dur: seconds(900), left: seconds(880)}},
	})
	want("eaten in combat", steps(&export{buffLog: timeline}),
		buffStep{0, []int32{wellFed}}, buffStep{seconds(30), nil},
		buffStep{seconds(80), []int32{wellFed}}, buffStep{seconds(110), nil})

	// A full record in the middle of the window: the export's guess counts until then, and a
	// buff that isn't in the full record isn't up.
	timeline = &buffTimeline{records: []buffRecord{{t: at(50), full: true}}}
	want("full record mid window", steps(&export{Buffs: guess, buffLog: timeline}),
		buffStep{0, []int32{wellFed}}, buffStep{seconds(50), nil})
}
