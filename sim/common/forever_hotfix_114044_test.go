//go:build with_db

package common_test

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/mage"
)

// The Darkmoon card and Onyxia Blood Talisman changes of wago hotfix push 114044 (2026-11-16).

func init() {
	sim.RegisterAll()
}

func newForeverMage(t *testing.T, trinket int32) (*core.Simulation, *mage.Mage) {
	t.Helper()
	specs := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range specs {
		specs[i] = &proto.ItemSpec{}
	}
	specs[proto.ItemSlot_ItemSlotTrinket1].Id = trinket
	player := &proto.Player{
		Race:      proto.Race_RaceGnome,
		Class:     proto.Class_ClassMage,
		Level:     60,
		Equipment: &proto.EquipmentSpec{Items: specs},
		Spec:      &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{}}},
		Rotation:  &proto.APLRotation{},
	}
	s := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 100000, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	s.Reset()
	s.PrePull()
	return s, s.Raid.Parties[0].Players[0].(*mage.Mage)
}

// Darkmoon Card: Blue Dragon procs Aura of the Blue Dragon (+150 Spirit) on 3% of spell casts.
// 20000 Fire Blasts should give about 600 procs (sd about 24); the old 2% would give 400.
func TestForeverDarkmoonCardBlueDragon(t *testing.T) {
	s, m := newForeverMage(t, common.DarkmoonCardBlueDragon)
	aura := m.GetAura("Aura of the Blue Dragon")
	if aura == nil {
		t.Fatal("no Aura of the Blue Dragon with the card")
	}
	var fireBlast *core.Spell
	for _, rank := range m.FireBlast {
		if rank != nil {
			fireBlast = rank
		}
	}
	refill := &core.ResourceMetrics{Type: proto.ResourceType_ResourceTypeMana}
	spirit := m.GetStat(stats.Spirit)
	procs := 0
	for range 20000 {
		s.CurrentTime += 10 * time.Second
		m.AddMana(s, m.MaxMana(), refill)
		if !fireBlast.Cast(s, m.CurrentTarget) {
			t.Fatal("Fire Blast didn't cast")
		}
		if aura.IsActive() {
			procs++
			if got := m.GetStat(stats.Spirit) - spirit; got != 150 {
				t.Fatalf("Aura of the Blue Dragon gives %.0f Spirit, want 150", got)
			}
			aura.Deactivate(s)
		}
	}
	if procs < 500 || procs > 700 {
		t.Errorf("%d Blue Dragon procs from 20000 casts, want about 600 (3%%)", procs)
	}
}

// Darkmoon Card: Heroism heals for 198 with a variance of 0.4: 158 to 238.
func TestForeverDarkmoonCardHeroism(t *testing.T) {
	s, m := newForeverMage(t, common.DarkmoonCardHeroism)
	heal := m.GetSpell(core.ActionID{SpellID: 23689})
	if heal == nil {
		t.Fatal("no Heroism heal with the card")
	}
	lo, hi, sum := 1e9, 0.0, 0.0
	const n = 2000
	for range n {
		m.RemoveHealth(s, 500)
		before := m.CurrentHealth()
		heal.Cast(s, &m.Unit)
		got := m.CurrentHealth() - before
		lo, hi, sum = min(lo, got), max(hi, got), sum+got
	}
	if lo < 158 || hi > 238 {
		t.Errorf("Heroism heals %.1f to %.1f, want within 158 to 238", lo, hi)
	}
	if avg := sum / n; avg < 194 || avg > 202 {
		t.Errorf("Heroism heals %.1f on average, want about 198", avg)
	}
}

// Onyxia Blood Talisman's stats in the database: the Forever ones (Wowhead Forever tooltip
// 2026-11-17, foreverchanges' notes), not the 20 Fire Resistance it had before.
func TestForeverOnyxiaBloodTalisman(t *testing.T) {
	talisman, ok := core.ItemsByID[18406]
	if !ok {
		t.Fatal("no Onyxia Blood Talisman in the database")
	}
	want := stats.Stats{stats.FireResistance: 25, stats.Defense: 8, stats.Parry: 1}
	if talisman.Stats != want {
		t.Errorf("Onyxia Blood Talisman has %v, want %v", talisman.Stats.FlatString(), want.FlatString())
	}
}
