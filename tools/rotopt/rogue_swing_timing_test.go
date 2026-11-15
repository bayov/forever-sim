package main

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
)

// swingTimedLines counts the cast lines whose condition reads the main hand swing timer.
func swingTimedLines(tmpl Template, k Knobs) int {
	n := 0
	for _, item := range tmpl.Build(k).PriorityList {
		if item.Action.GetCastSpell() != nil && strings.Contains(protojson.Format(item.Action.GetCondition()), "autoTimeToNext") {
			n++
		}
	}
	return n
}

// The Windfury Totem swing timing knobs (mhGap, preSwing, evisPreSwing) are off by default, and the
// rotations then read the swing timer only on the Land Mine line, as before the knobs. Set, mhGap
// times the builder and Eviscerate, preSwing the builder and evisPreSwing Eviscerate.
func TestRogueSwingTimingKnobs(t *testing.T) {
	for _, tmpl := range []Template{rogueSS{}, rogueBS{}} {
		off := defaultKnobs(tmpl)
		off["ssAfterSwing"], off["bsAfterSwing"] = 0, 0
		if got := swingTimedLines(tmpl, off); got != 1 {
			t.Errorf("%T with the knobs off: %d lines read the swing timer, want 1 (the Land Mine)", tmpl, got)
		}
		for name, want := range map[string]int{"mhGap": 3, "preSwing": 2, "evisPreSwing": 2} {
			on := defaultKnobs(tmpl)
			on["ssAfterSwing"], on["bsAfterSwing"] = 0, 0
			on[name] = 0.1
			if got := swingTimedLines(tmpl, on); got != want {
				t.Errorf("%T %s=0.1: %d lines read the swing timer, want %d", tmpl, name, got, want)
			}
		}
	}
}
