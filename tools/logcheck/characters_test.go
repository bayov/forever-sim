package main

import (
	"testing"
	"time"
)

// TestMeSpans: the log's own player is the one with the "mine" flags, whatever its name. A
// party member (0x512) and a stranger (0x518) are not us, and we change character halfway.
func TestMeSpans(t *testing.T) {
	stamp := func(sec int) string {
		return time.Date(2026, 10, 16, 18, 0, sec, 0, time.UTC).Format("1/2/2006 15:04:05.0000")
	}
	text := stamp(0) + `  SPELL_CAST_SUCCESS,Player-1-A,"Bayov-ClassicBetaPvE2-",0x511,0x0,0000000000000000,nil,0x80000000,0x80000000,403,"Lightning Bolt",0x8` + "\n" +
		stamp(1) + `  SPELL_CAST_SUCCESS,Player-1-B,"Friend-ClassicBetaPvE2-",0x512,0x0,0000000000000000,nil,0x80000000,0x80000000,403,"Lightning Bolt",0x8` + "\n" +
		stamp(2) + `  SWING_DAMAGE,Creature-0-1-2-3-4-5,"Mob",0xa48,0x0,Player-1-A,"Bayov-Oom-ClassicBetaPvE2-",0x511,0x0` + "\n" +
		stamp(3) + `  SPELL_CAST_SUCCESS,Player-1-C,"Stranger-ClassicBetaPvE2-",0x518,0x0,0000000000000000,nil,0x80000000,0x80000000,403,"Lightning Bolt",0x8` + "\n" +
		stamp(9) + `  SPELL_CAST_SUCCESS,Player-1-D,"Alt-ClassicBetaPvE2-",0x511,0x0,0000000000000000,nil,0x80000000,0x80000000,403,"Lightning Bolt",0x8` + "\n"
	events, err := parseLines(text)
	if err != nil {
		t.Fatal(err)
	}
	spans := meSpans(events)
	if len(spans) != 2 || spans[0].GUID != "Player-1-A" || spans[0].To != events[2].t || spans[1].GUID != "Player-1-D" {
		t.Fatalf("spans %+v, want Player-1-A for 0 to 2 sec, then Player-1-D", spans)
	}

	// An export without a GUID takes the character the log had at its time, and a whole
	// history session takes what most of its records were placed on. One with a GUID keeps it.
	at := func(sec int) time.Time { return time.Date(2026, 10, 16, 18, 0, sec, 0, time.UTC) }
	pasted := &export{time: at(1)}
	login := &export{time: at(1), history: true, session: "h#1"}
	quiet := &export{time: at(0).Add(-2 * time.Hour), history: true, session: "h#1"}
	tagged := &export{time: at(1)}
	tagged.Who.GUID = "Player-1-Z"
	placeExports([]*export{pasted, login, quiet, tagged}, spans)
	if pasted.guid != "Player-1-A" || login.guid != "Player-1-A" || quiet.guid != "Player-1-A" || tagged.guid != "Player-1-Z" {
		t.Errorf("placed %q %q %q %q, want Player-1-A three times and Player-1-Z", pasted.guid, login.guid, quiet.guid, tagged.guid)
	}
}
