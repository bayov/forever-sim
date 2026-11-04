package sim

// The healing engine's fixes of 2026-11-04 (healer_mechanics-20261104-7 and -8, sim/core): a direct
// heal takes the caster's school crit bonus, and current health follows max health down. On the
// priest's small healing fight (raid_healing_priest_test.go). Needs the item database: go test -tags
// with_db.

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// testSetItemTypes are the slots wearTestSet fills, in order.
var testSetItemTypes = []proto.ItemType{proto.ItemType_ItemTypeHead, proto.ItemType_ItemTypeShoulder, proto.ItemType_ItemTypeChest,
	proto.ItemType_ItemTypeHands, proto.ItemType_ItemTypeLegs, proto.ItemType_ItemTypeFeet, proto.ItemType_ItemTypeWaist,
	proto.ItemType_ItemTypeWrist}

// wearTestSet puts pieces of the named set on the player: test items with no stats of their own
// (IDs from firstID), so only the set's bonuses move the player.
func wearTestSet(t *testing.T, p *proto.Player, setName string, firstID int32, pieces int) {
	t.Helper()
	p.Equipment = &proto.EquipmentSpec{}
	for i := 0; i < pieces; i++ {
		id := firstID + int32(i)
		item := core.Item{ID: id, Name: setName + " test piece", Type: testSetItemTypes[i], SetName: setName}
		if old, ok := core.ItemsByID[id]; ok && old.Name != item.Name {
			t.Fatalf("item %d is %s", id, old.Name)
		} else if !ok {
			core.ItemsByID[id] = item
			t.Cleanup(func() { delete(core.ItemsByID, id) })
		}
		p.Equipment.Items = append(p.Equipment.Items, &proto.ItemSpec{Id: id})
	}
}

// at runs the sim to its end, doing each step at its time.
func at(sim *core.Simulation, steps map[time.Duration]func(*core.Simulation)) {
	for when, do := range steps {
		sim.AddPendingAction(&core.PendingAction{NextActionAt: when, OnAction: do})
	}
	for !sim.Step() {
	}
}

// A direct heal crits with the caster's school crit bonus (healer_mechanics-20261104-7): Vestments
// of Prophecy's 5 piece (+2% Holy crit) raises Greater Heal's crit chance by 2 points, and a 50 point
// Holy bonus makes about half again as many of its heals crit.
func TestHealingCoreDirectHealSchoolCrit(t *testing.T) {
	chance := func(pieces int) float64 {
		rsr := priestRaid(t, "", 0)
		wearTestSet(t, rsr.Raid.Parties[0].Players[1], "Vestments of Prophecy", 999_300, pieces)
		_, hp := newPriestSim(t, rsr)
		return hp.GreaterHeal[5].HealingCritChance()
	}
	if four, five := chance(4), chance(5); !near(five-four, 0.02, 1e-9) {
		t.Errorf("Greater Heal crits %.4f with 5 pieces of Vestments of Prophecy, %.4f with 4: want 0.02 more", five, four)
	}

	rsr := priestRaid(t, "", 0)
	rsr.Encounter.Targets[0].MinBaseDamage = 0
	sim, hp := newPriestSim(t, rsr)
	tank := &sim.Raid.Parties[0].Players[0].GetCharacter().Unit
	gh := hp.GreaterHeal[5]
	critRate := func(sim *core.Simulation) float64 {
		const n = 1000
		before := gh.SpellMetrics[tank.UnitIndex].Crits
		for i := 0; i < n; i++ {
			gh.ApplyEffects(sim, tank, gh)
		}
		return float64(gh.SpellMetrics[tank.UnitIndex].Crits-before) / n
	}
	at(sim, map[time.Duration]func(*core.Simulation){time.Second: func(sim *core.Simulation) {
		without := critRate(sim)
		hp.PseudoStats.SchoolBonusCritChance[stats.SchoolIndexHoly] += 50 * core.SpellCritRatingPerCritChance
		with := critRate(sim)
		hp.PseudoStats.SchoolBonusCritChance[stats.SchoolIndexHoly] -= 50 * core.SpellCritRatingPerCritChance
		if with-without < 0.45 || with-without > 0.55 {
			t.Errorf("Greater Heal crit %.1f%% of the time with 50%% more Holy crit, %.1f%% without", 100*with, 100*without)
		}
	}})
}

// Current health comes down with max health (healer_mechanics-20261104-8): a full health tank whose
// Last Stand ends, or who loses Stamina, has its new max health, not more.
func TestHealingCoreHealthFollowsMaxHealthDown(t *testing.T) {
	rsr := priestRaid(t, "", 0)
	rsr.Encounter.Targets[0].MinBaseDamage = 0
	rsr.Raid.Parties[0].Players[0].Rotation = &proto.APLRotation{}
	sim, _ := newPriestSim(t, rsr)
	tank := sim.Raid.Parties[0].Players[0].GetCharacter()
	lastStand := tank.GetAura("Last Stand")
	if lastStand == nil {
		t.Fatal("the tank has no Last Stand")
	}
	full := 0.0
	at(sim, map[time.Duration]func(*core.Simulation){
		time.Second: func(sim *core.Simulation) {
			full = tank.MaxHealth()
			if tank.CurrentHealth() != full {
				t.Fatalf("the tank starts at %.0f of %.0f health", tank.CurrentHealth(), full)
			}
			lastStand.Activate(sim)
			if !near(tank.MaxHealth(), 1.3*full, 1e-9) || !near(tank.CurrentHealth(), 1.3*full, 1e-9) {
				t.Errorf("Last Stand: %.0f of %.0f health, want %.0f of %.0f", tank.CurrentHealth(), tank.MaxHealth(), 1.3*full, 1.3*full)
			}
		},
		2 * time.Second: func(sim *core.Simulation) {
			lastStand.Deactivate(sim)
			if !near(tank.MaxHealth(), full, 1e-9) || tank.CurrentHealth() != tank.MaxHealth() {
				t.Errorf("after Last Stand: %.0f of %.0f health, want %.0f of %.0f", tank.CurrentHealth(), tank.MaxHealth(), full, full)
			}
		},
		3 * time.Second: func(sim *core.Simulation) {
			tank.AddStatsDynamic(sim, stats.Stats{stats.Stamina: -100})
			if tank.MaxHealth() >= full || tank.CurrentHealth() != tank.MaxHealth() {
				t.Errorf("100 Stamina less: %.0f of %.0f health (from %.0f), want full at the new max", tank.CurrentHealth(), tank.MaxHealth(), full)
			}
		},
	})
}
