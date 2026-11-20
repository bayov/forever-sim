package main

import (
	"slices"
	"testing"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core/proto"
)

// A use gets its "vs 4 mobs" line only when it is worth more than aoeMinGain more on the mobs.
func TestAoeGains(t *testing.T) {
	for _, c := range []struct {
		single, aoe float64
		want        bool
	}{
		{2.7, 9.5, true},
		{2.7, 3.7, false}, // exactly aoeMinGain is within the noise
		{2.7, 3.8, true},
		{78.3, 78.1, false},
	} {
		if got := aoeGains(c.single, c.aoe); got != c.want {
			t.Errorf("aoeGains(%g, %g) = %t, want %t", c.single, c.aoe, got, c.want)
		}
	}
}

// With useTargets, the use runs fight 4 copies of the boss and the run at the fight length stays
// on the boss. Smokey's Lighter's fire cone hits all 4, so its use gains.
//
// Manual Crowd Pummeler's haste gains a little too (+67.1 to +72.1 at 3000 iterations on
// 2026-11-20), because more swings give more Dragonbreath Chili procs, and its fire hits all 4
// mobs. That is real on a pull, so we don't test it for no gain.
func TestUseValuesAgainstMobs(t *testing.T) {
	sim.RegisterAll()
	defer func(dur float64, iters, targets int) { *useDur, *useIters, *useAoeTargets = dur, iters, targets }(*useDur, *useIters, *useAoeTargets)
	*useDur, *useIters, *useAoeTargets = 90, 3000, 4

	setup, err := loadSetup("../dps_report/req/shaman__enhancement-level60.json")
	if err != nil {
		t.Fatal(err)
	}
	items := setup.player().Equipment.Items
	items[proto.ItemSlot_ItemSlotTrinket1] = &proto.ItemSpec{Id: common.SmokeysLighter}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: common.ManualCrowdPummeler}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{}
	s := &searcher{setup: setup, template: fixedRotation{setup.player().Rotation}, durations: []float64{180},
		iterations: 300, useScores: true, cache: map[string]result{}}

	score := func(targets int) result {
		s.useTargets = targets
		s.cache = map[string]result{}
		return s.scoreAll([]*pendingScore{s.prepare(nil)})[0]
	}
	boss, mobs := score(0), score(4)
	if len(setup.request.Encounter.Targets) != 1 {
		t.Fatalf("the setup's encounter has %d targets after the runs, want 1", len(setup.request.Encounter.Targets))
	}
	// The part at the fight length is the same run, to the last digit.
	if a, b := boss.dps-boss.useSum, mobs.dps-mobs.useSum; a != b {
		t.Errorf("the 180s part against the boss %.4f and with the uses on 4 mobs %.4f, want the same", a, b)
	}
	lighter, pummeler := int32(common.SmokeysLighter), int32(common.ManualCrowdPummeler)
	t.Logf("Smokey's Lighter's use %+.1f on the boss, %+.1f on 4 mobs", boss.uses[lighter], mobs.uses[lighter])
	t.Logf("Manual Crowd Pummeler's use %+.1f on the boss, %+.1f on 4 mobs", boss.uses[pummeler], mobs.uses[pummeler])
	if !aoeGains(boss.uses[lighter], mobs.uses[lighter]) {
		t.Errorf("Smokey's Lighter's use %+.1f on the boss and %+.1f on 4 mobs, want a gain over %g", boss.uses[lighter], mobs.uses[lighter], aoeMinGain)
	}

	// Both have a use now that they ran, so both are ranked "vs 4 mobs", and Smokey's Lighter vs
	// Plants too.
	labels := func(id int32) []string {
		var out []string
		for _, c := range itemBisConditions(id, setup.request.Encounter) {
			out = append(out, c.label())
		}
		return out
	}
	if got := labels(lighter); !slices.Equal(got, []string{"vs Plants", "vs 4 mobs"}) {
		t.Errorf("Smokey's Lighter's conditions %q, want vs Plants and vs 4 mobs", got)
	}
	if got := labels(pummeler); !slices.Equal(got, []string{"vs 4 mobs"}) {
		t.Errorf("Manual Crowd Pummeler's conditions %q, want vs 4 mobs", got)
	}
}
