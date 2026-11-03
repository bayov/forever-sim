//go:build practice

package practice

import (
	"testing"
	"time"
)

// Totem hits say their totem, and procs carry the client's spell for their look.
func TestHitVisuals(t *testing.T) {
	s := newEnhSession(t, Settings{})
	d := newDriver(t, s)
	d.standAt(0, -3)
	d.mustCast(searingTotem)
	d.press(Input{Kind: "startAttack"})
	var hits []HitEvent
	for i := 0; i < 20; i++ {
		for _, b := range d.wait(time.Second) {
			hits = append(hits, b.Hits...)
		}
	}
	var bolts, procs int
	for _, h := range hits {
		switch h.SpellID {
		case 10436:
			bolts++
			if h.Source != "totem" || h.Totem != "fire" {
				t.Errorf("Searing Totem bolt %+v: want source totem, totem fire", h)
			}
			if h.NextAt != 0 && h.NextAt-h.At < 2000 {
				t.Errorf("Searing Totem bolt at %.0f says the next hits at %.0f", h.At, h.NextAt)
			}
		case 16362:
			procs++
			if h.VisualSpellID != windfuryAttack || h.Source != "player" {
				t.Errorf("Windfury proc %+v: want visual %d from the player", h, windfuryAttack)
			}
		}
	}
	// Fire Nova goes off at the fire totem, with its damage spell's look.
	d.press(Input{Kind: "cast", SpellID: 408345, Target: &UnitRef{Kind: "target"}})
	var nova []HitEvent
	var novaCasts []CastEvent
	for _, b := range d.wait(time.Second) {
		nova = append(nova, b.Hits...)
		novaCasts = append(novaCasts, b.Casts...)
	}
	novaHit, novaCast := false, false
	for _, h := range nova {
		if h.SpellID == 408345 && h.Tag == 1 {
			novaHit = h.Totem == "fire" && h.Source == "totem" && h.VisualSpellID == 408428
		}
	}
	for _, c := range novaCasts {
		if c.SpellID == 408345 {
			novaCast = c.Totem == "fire" && c.VisualSpellID == 408428
		}
	}
	if !novaHit || !novaCast {
		t.Errorf("Fire Nova: hit ok %v, cast ok %v (%+v, %+v)", novaHit, novaCast, nova, novaCasts)
	}
	if bolts < 5 {
		t.Errorf("%d Searing Totem bolts in 20s, want about 8", bolts)
	}
	ids := map[[3]int32]int{}
	for _, h := range hits {
		ids[[3]int32{h.SpellID, h.OtherID, h.VisualSpellID}]++
	}
	t.Logf("%d bolts, %d Windfury procs, %v", bolts, procs, ids)
}
