package main

import (
	"slices"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core/stats"
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
		buffStep{at: 0, ids: []int32{wellFed}})

	// A full record before the window says Well Fed has 90 seconds left a minute before it, so
	// it runs out 30 seconds in. The export's guess no longer counts.
	timeline := &buffTimeline{records: []buffRecord{
		{t: at(-60), full: true, gained: []buffGain{{id: wellFed, dur: seconds(900), left: seconds(90)}}},
	}}
	want("ran out", steps(&export{Buffs: guess, buffLog: timeline}),
		buffStep{at: 0, ids: []int32{wellFed}}, buffStep{at: seconds(30), ids: nil})

	// We ate again in combat. The record after it came at 100 seconds, after 40 seconds of the
	// game hiding our buffs, and the buff had 880 of its 900 seconds left. So we ate at 80
	// seconds, and the death at 110 took it off.
	timeline.records = append(timeline.records, buffRecord{
		t: at(100), hidden: seconds(40), gained: []buffGain{{id: wellFed, dur: seconds(900), left: seconds(880)}},
	})
	want("eaten in combat", steps(&export{buffLog: timeline}),
		buffStep{at: 0, ids: []int32{wellFed}}, buffStep{at: seconds(30), ids: nil},
		buffStep{at: seconds(80), ids: []int32{wellFed}}, buffStep{at: seconds(110), ids: nil})

	// A full record in the middle of the window: the export's guess counts until then, and a
	// buff that isn't in the full record isn't up.
	timeline = &buffTimeline{records: []buffRecord{{t: at(50), full: true}}}
	want("full record mid window", steps(&export{Buffs: guess, buffLog: timeline}),
		buffStep{at: 0, ids: []int32{wellFed}}, buffStep{at: seconds(50), ids: nil})
}

// The mobs' Demoralizing Shout grows with the caster's level, so a level 60 Stonehearth Marshal's
// is -520 attack power, which is what our logged attack power lost under it on Nov 25. It's on
// us only while the log says, not for the whole window.
func TestWindowBuffsMobDebuffLevel(t *testing.T) {
	const shout = 23511
	base := time.Date(2026, 11, 25, 20, 30, 35, 0, time.UTC)
	at := func(sec float64) time.Time { return base.Add(seconds(sec)) }
	const me, marshal = "Player-1-0001", "Creature-0-1-30-1-14765-0000000001"
	aura := func(sec float64, kind string) *event {
		return &event{t: at(sec), kind: kind, srcGUID: marshal, dstGUID: me, spellID: shout}
	}
	all := []*event{
		{t: at(-60), kind: "SWING_DAMAGE", srcGUID: me},
		aura(104, "SPELL_AURA_APPLIED"),
		aura(119, "SPELL_AURA_REMOVED"),
	}
	cl := &combatLog{me: me, from: at(0), to: at(144), levels: map[string]int{marshal: 60}}
	tables := &clientTables{
		spellNames:  map[int32]string{shout: "Demoralizing Shout"},
		buffEffects: map[int32][]buffEffect{shout: {{aura: 99, amount: -40, perLevel: -8}}},
		buffLevels:  map[int32]spellLevels{shout: {base: 0}},
	}
	steps := windowBuffs(all, cl, &export{}, tables, map[int32]bool{})
	if len(steps) != 3 || steps[1].at != seconds(104) || steps[2].at != seconds(119) || len(steps[2].ids) != 0 {
		t.Fatalf("got steps %v, want the shout from 104s to 119s", steps)
	}
	if got := tables.buffTotal(steps[1].ids, steps[1].levels, stats.Stats{}, campItems{})[stats.AttackPower]; got != -520 {
		t.Errorf("shout attack power %v, want -520", got)
	}
}
