package core

import (
	"testing"

	"github.com/wowsims/classic/sim/core/stats"
)

// The replay follows max health down as the live run does (unit.go cuts current health to a lower
// max health): a tank at 100 of 100 whose max health drops to 50 (Last Stand ending) and then takes
// 60 dies, where a replay that kept the old 100 would live on 40 (tank_mechanics-20261104-14).
func TestTankScoreReplayCapsAtMaxHealthDrop(t *testing.T) {
	unit := &Unit{}
	unit.healthBar.unit = unit
	unit.stats[stats.Health] = 100
	ts := &tankScore{replay: true, hps: 1}
	ts.startHealth = 100

	unit.stats[stats.Health] = 50
	ts.damageTaken(unit, 60)

	if ts.lives(0) {
		t.Errorf("the replay kept health over the dropped max health: lived through 60 damage at 50 max health")
	}
	if ts.lives(1e9) {
		t.Errorf("no stand-in heal came before the damage, yet the replay lived")
	}

	// A heal after the drop still caps at the lower max health.
	ts.reset()
	ts.startHealth = 50
	ts.healTaken(unit, 1000, &ResourceMetrics{})
	ts.damageTaken(unit, 49)
	if !ts.lives(0) {
		t.Errorf("the tank at 50 of 50 died to 49 damage")
	}
}
