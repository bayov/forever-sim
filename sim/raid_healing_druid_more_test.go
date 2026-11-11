package sim

// The Restoration druid's paths the healer mechanics audit listed as untested
// (orchestration/audit/reports/healer_mechanics.md, healer_mechanics-20261104-10): Swiftmend on a
// Regrowth, which HoT it reads, its talents and the Tier 1 set; Gift of Nature and Genesis on
// Tranquility and Wild Growth; Nature's Splendor on Regrowth; HoT tick crits; Nature's Swiftness's
// heals; Moonglow on heals; Bloom on Regrowth. raid_healing_druid_test.go has the helpers. Needs the
// item database: go test -tags with_db.

import (
	"strconv"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

func both(conds ...string) string {
	s := `{"and":{"vals":[`
	for i, c := range conds {
		if i > 0 {
			s += ","
		}
		s += c
	}
	return s + `]}}`
}

func hotOn(spellID int32, target string) string {
	return `{"dotIsActive":{"spellId":{"spellId":` + strconv.Itoa(int(spellID)) + `},"targetUnit":{"type":"` + target + `"}}}`
}

func after(seconds string) string {
	return `{"cmp":{"op":"OpGt","lhs":{"currentTime":{}},"rhs":{"const":{"val":"` + seconds + `"}}}}`
}

// Swiftmend on a Regrowth heals all of its HoT, Nature's Splendor's 2 more ticks too (9 in all), with
// Gift of Nature's and Genesis's 10% each on it; Regrowth's HoT ticks 9 times.
func TestDruidSwiftmendOnRegrowth(t *testing.T) {
	talents := map[string]int{"swiftmend": 1, "naturesSplendor": 1, "giftOfNature": 5, "genesis": 5}
	l := runDruidLog(t, druidRaid(t, talents, 0, castAt(smID, "Tank", ""), castAt(rgTop, "Tank", hotOff(rgTop, "Tank"))))
	sm := l.directs(smID)
	if len(sm) < 5 {
		t.Fatalf("%d Swiftmends", len(sm))
	}
	for _, h := range sm {
		if !druidNear(h.base, 9*(142+0.071*h.healingPower), 2) {
			t.Errorf("a Swiftmend of %.1f, want 9 Regrowth ticks at %.0f healing power", h.base, h.healingPower)
		}
	}
	if m := l.multiplier(t, sm); !druidNear(m, 1.1*1.1, 0.002) {
		t.Errorf("Swiftmend multiplier %.4f with Gift of Nature 5/5 and Genesis 5/5, want 1.21", m)
	}
	ticks, casts := len(l.ticks(rgTop)), len(l.casts[rgTop])
	if ticks < 9*(casts-1) || ticks > 9*casts {
		t.Errorf("%d Regrowth ticks from %d casts with Nature's Splendor, want 9 a cast", ticks, casts)
	}
}

// With both HoTs up, Swiftmend reads the one with the least time left.
func TestDruidSwiftmendPicksLeastTimeLeft(t *testing.T) {
	sm := castAt(smID, "Tank", both(hotOn(rjTop, "Tank"), hotOn(rgTop, "Tank")))
	// Rejuvenation at 0 (ends at 12), Regrowth after it (ends at 24.5): Swiftmend reads Rejuvenation.
	l := runDruidLog(t, druidRaid(t, map[string]int{"swiftmend": 1}, 0, sm,
		castAt(rjTop, "Tank", hotOff(rjTop, "Tank")), castAt(rgTop, "Tank", hotOff(rgTop, "Tank"))))
	if h := l.directs(smID); len(h) == 0 || !druidNear(h[0].base, 4*(194+0.2*h[0].healingPower), 2) {
		t.Errorf("the first Swiftmend %v, want Rejuvenation's 4 ticks", h[:min(1, len(h))])
	}
	// Regrowth at 0 (ends at 23), Rejuvenation at 12 (ends at 24): Swiftmend reads Regrowth.
	l = runDruidLog(t, druidRaid(t, map[string]int{"swiftmend": 1}, 0, sm,
		castAt(rgTop, "Tank", hotOff(rgTop, "Tank")), castAt(rjTop, "Tank", both(hotOff(rjTop, "Tank"), after("12s")))))
	if h := l.directs(smID); len(h) == 0 || !druidNear(h[0].base, 7*(142+0.071*h[0].healingPower), 2) {
		t.Errorf("the first Swiftmend %v, want Regrowth's 7 ticks", h[:min(1, len(h))])
	}
}

// druidSetPieces makes the healer's Tier 1 Restoration set n pieces: the bench gear carries two, so the
// test lends the set's name to other equipped items with no set of their own (and takes it from the
// set's pieces past n). It returns a func that puts the item database back.
func druidSetPieces(t *testing.T, rsr *proto.RaidSimRequest, n int) func() {
	t.Helper()
	const setName = "Restoration Druid Synthetic (Phase 1)"
	saved := map[int32]core.Item{}
	count := 0
	for _, spec := range rsr.Raid.Parties[0].Players[1].Equipment.Items {
		if spec == nil || spec.Id == 0 {
			continue
		}
		item, ok := core.ItemsByID[spec.Id]
		if !ok || (item.SetName != "" && item.SetName != setName) || (item.SetID != 0 && item.SetName != setName) {
			continue
		}
		if _, done := saved[spec.Id]; done {
			continue
		}
		saved[spec.Id] = item
		if count < n {
			item.SetName, item.SetID = setName, 0
			count++
		} else {
			item.SetName, item.SetID = "", 0
		}
		core.ItemsByID[spec.Id] = item
	}
	if count != n {
		t.Fatalf("only %d items to make set pieces of", count)
	}
	return func() {
		for id, item := range saved {
			core.ItemsByID[id] = item
		}
	}
}

// Forever's Tier 1 Restoration set: 2 pieces +10 Spirit, 4 pieces +26 healing, 5 pieces 3 sec off
// Swiftmend's cooldown (with the bench's Idol of Swiftness, 15 - 3 - 3 = 9).
func TestDruidTier1Restoration(t *testing.T) {
	type got struct{ bloom, healingPower, swiftmendCD float64 }
	run := func(n int) got {
		rsr := druidRaid(t, map[string]int{"swiftmend": 1, "bloom": 1}, 0,
			castAt(smID, "Tank", ""), castAt(rjTop, "Tank", hotOff(rjTop, "Tank")), castAt(htTop, "Tank", ""))
		restore := druidSetPieces(t, rsr, n)
		defer restore()
		l := runDruidLog(t, rsr)
		if len(l.directs(bloom)) == 0 || len(l.directs(htTop)) == 0 || len(l.cooldowns[smID]) == 0 {
			t.Fatalf("%d pieces: no Bloom, Healing Touch or Swiftmend", n)
		}
		return got{l.directs(bloom)[0].base, l.directs(htTop)[0].healingPower, l.cooldowns[smID][0]}
	}
	none, two, four, five := run(0), run(2), run(4), run(5)
	// Bloom heals for our Spirit; the raid's Spirit buffs may scale the 10 up a little.
	if d := two.bloom - none.bloom; d < 9.9 || d > 12.5 {
		t.Errorf("Bloom %.1f with 2 pieces, %.1f with none: want 10 Spirit more", two.bloom, none.bloom)
	}
	if d := four.healingPower - two.healingPower; !druidNear(d, 26, 0.5) {
		t.Errorf("healing power %.1f with 4 pieces, %.1f with 2: want 26 more", four.healingPower, two.healingPower)
	}
	if none.swiftmendCD != 12 || four.swiftmendCD != 12 || five.swiftmendCD != 9 {
		t.Errorf("Swiftmend cooldowns with 0, 4 and 5 pieces %.1f, %.1f, %.1f; want 12, 12 and 9", none.swiftmendCD, four.swiftmendCD, five.swiftmendCD)
	}
}

// Gift of Nature 5/5 adds 10% to Tranquility and Wild Growth; Genesis 5/5 10% to Wild Growth and
// nothing to Tranquility.
func TestDruidTalentsOnTranquilityAndWildGrowth(t *testing.T) {
	mults := func(talents map[string]int) (float64, float64) {
		talents["wildGrowth"] = 1
		l := runDruidLog(t, druidRaid(t, talents, 3, castAt(trTop, "", ""), castAt(wgTop, "Tank", "")))
		tr, wg := l.ticks(trTop), l.ticks(wgTop)
		if len(tr) == 0 || len(wg) == 0 {
			t.Fatalf("%v: %d Tranquility and %d Wild Growth ticks", talents, len(tr), len(wg))
		}
		// The first tick of a Wild Growth takes its 6% on top.
		return l.multiplier(t, tr), wg[0].afterMods / wg[0].base
	}
	tr0, wg0 := mults(map[string]int{})
	for _, c := range []struct {
		talents  map[string]int
		tr, wg   float64
		describe string
	}{
		{map[string]int{"giftOfNature": 5}, 1.1, 1.1, "Gift of Nature 5/5"},
		{map[string]int{"genesis": 5}, 1, 1.1, "Genesis 5/5"},
	} {
		tr, wg := mults(c.talents)
		if !druidNear(tr/tr0, c.tr, 0.002) || !druidNear(wg/wg0, c.wg, 0.002) {
			t.Errorf("%s: Tranquility x%.4f, Wild Growth x%.4f; want x%.2f and x%.2f", c.describe, tr/tr0, wg/wg0, c.tr, c.wg)
		}
	}
}

// HoT ticks crit under Forever (Rejuvenation's and Tranquility's), and Improved Regrowth 5/5's 50% crit
// reaches Regrowth's HoT ticks as well as its direct heal.
func TestDruidHotTickCrits(t *testing.T) {
	share := func(events []healEvent) float64 {
		crits := 0
		for _, h := range events {
			if h.crit {
				crits++
			}
		}
		return float64(crits) / float64(len(events))
	}
	l := runDruidLog(t, druidRaid(t, nil, 3, castAt(trTop, "", ""), castAt(rjTop, "Tank", hotOff(rjTop, "Tank"))))
	if share(l.ticks(rjTop)) == 0 || share(l.ticks(trTop)) == 0 {
		t.Errorf("Rejuvenation ticks crit %.2f, Tranquility ticks %.2f: want some crits", share(l.ticks(rjTop)), share(l.ticks(trTop)))
	}
	base := runDruidLog(t, druidRaid(t, nil, 0, castAt(rgTop, "Tank", hotOff(rgTop, "Tank"))))
	imp := runDruidLog(t, druidRaid(t, map[string]int{"improvedRegrowth": 5}, 0, castAt(rgTop, "Tank", hotOff(rgTop, "Tank"))))
	if b, i := share(base.ticks(rgTop)), share(imp.ticks(rgTop)); len(imp.ticks(rgTop)) < 40 || i-b < 0.3 {
		t.Errorf("Regrowth ticks crit %.2f with Improved Regrowth 5/5 against %.2f (%d ticks)", i, b, len(imp.ticks(rgTop)))
	}
}

// Nature's Swiftness makes the next Regrowth instant too, and Wild Growth since hotfix push 113885:
// a Wild Growth cast first takes it, and the Regrowths after it are 2 sec.
func TestDruidNaturesSwiftnessOnRegrowth(t *testing.T) {
	l := runDruidLog(t, druidRaid(t, map[string]int{"naturesSwiftness": 1, "wildGrowth": 1}, 0,
		castAt(nsID, "", ""), castAt(wgTop, "Tank", `{"not":{"val":`+after("1s")+`}}`), castAt(rgTop, "Tank", "")))
	wg, rg := l.casts[wgTop], l.casts[rgTop]
	if len(wg) != 1 || wg[0].castTime != 0 {
		t.Errorf("Wild Growth casts %v, want one instant", wg)
	}
	if len(rg) < 2 || rg[0].castTime != 2 || rg[1].castTime != 2 {
		t.Errorf("Regrowth casts %v, want 2 sec ones after Wild Growth took Nature's Swiftness", rg[:min(2, len(rg))])
	}

	l = runDruidLog(t, druidRaid(t, map[string]int{"naturesSwiftness": 1}, 0, castAt(nsID, "", ""), castAt(rgTop, "Tank", "")))
	if rg := l.casts[rgTop]; len(rg) < 2 || rg[0].castTime != 0 || rg[1].castTime != 2 {
		t.Errorf("Regrowth casts %v, want the first instant and the next 2 sec", rg[:min(2, len(rg))])
	}
}

// Moonglow takes nothing off the heals.
func TestDruidMoonglowLeavesHeals(t *testing.T) {
	l := runDruidLog(t, druidRaid(t, map[string]int{"moonglow": 3}, 0,
		castAt(rjTop, "Tank", hotOff(rjTop, "Tank")), castAt(rgTop, "Tank", hotOff(rgTop, "Tank")), castAt(htTop, "Tank", "")))
	if l.casts[htTop][0].cost != 840 || l.casts[rgTop][0].cost != 525 || l.casts[rjTop][0].cost != 360 {
		t.Errorf("with Moonglow 3/3 Healing Touch, Regrowth and Rejuvenation cost %.1f, %.1f and %.1f", l.casts[htTop][0].cost, l.casts[rgTop][0].cost, l.casts[rjTop][0].cost)
	}
}

// Bloom also heals when a Regrowth runs out.
func TestDruidBloomOnRegrowth(t *testing.T) {
	l := runDruidLog(t, druidRaid(t, map[string]int{"bloom": 1}, 0, castAt(rgTop, "Tank", hotOff(rgTop, "Tank"))))
	if blooms, casts := len(l.directs(bloom)), len(l.casts[rgTop]); blooms < casts-1 || blooms > casts || casts < 5 {
		t.Errorf("%d Blooms from %d Regrowths left to run out", blooms, casts)
	}
}
