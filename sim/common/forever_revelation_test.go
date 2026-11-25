//go:build with_db

package common_test

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/mage"
)

// A level 60 Forever mage with Revelation on a Hanzo Sword, against targets enemies.
func newRevelationMage(t *testing.T, targets int) (*core.Simulation, *mage.Mage) {
	t.Helper()
	specs := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range specs {
		specs[i] = &proto.ItemSpec{}
	}
	specs[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: common.HanzoSword, Enchant: common.RevelationEffectID}
	player := &proto.Player{
		Race:      proto.Race_RaceGnome,
		Class:     proto.Class_ClassMage,
		Level:     60,
		Equipment: &proto.EquipmentSpec{Items: specs},
		Spec:      &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{}}},
		Rotation:  &proto.APLRotation{},
	}
	encounter := &proto.Encounter{Duration: 600}
	for range targets {
		encounter.Targets = append(encounter.Targets, core.NewDefaultTarget())
	}
	s := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	s.Reset()
	s.PrePull()
	return s, s.Raid.Parties[0].Players[0].(*mage.Mage)
}

func alwaysHit(_ *core.Simulation, result *core.SpellResult, _ *core.AttackTable) {
	result.Outcome = core.OutcomeHit
}

func alwaysCrit(_ *core.Simulation, result *core.SpellResult, _ *core.AttackTable) {
	result.Outcome = core.OutcomeCrit
}

// TestForeverRevelationChance pins Revelation's chance by class from EV-0149's counts, and that a
// warrior or rogue has none.
func TestForeverRevelationChance(t *testing.T) {
	for class, want := range map[proto.Class]float64{
		proto.Class_ClassMage:    0.0894,
		proto.Class_ClassWarlock: 0.0884,
		proto.Class_ClassPriest:  0.0865,
		proto.Class_ClassDruid:   0.0790,
		proto.Class_ClassShaman:  0.0730,
		proto.Class_ClassPaladin: 0.0702,
		proto.Class_ClassHunter:  0.08,
		proto.Class_ClassWarrior: 0,
		proto.Class_ClassRogue:   0,
	} {
		if got := common.RevelationClassChance[class]; got != want {
			t.Errorf("%s: Revelation chance %.4f, want %.4f", class, got, want)
		}
	}
}

// TestForeverRevelationRollsAndSpends checks which hits roll Revelation and that a cast spends it
// as it completes, so every target of an area spell crits (EV-0028, EV-0149, R-0051).
func TestForeverRevelationRollsAndSpends(t *testing.T) {
	defer func(old float64) { common.RevelationProcChance = old }(common.RevelationProcChance)
	common.RevelationProcChance = 1

	sim, m := newRevelationMage(t, 2)
	buff := m.GetAura("Revelation")
	if buff == nil {
		t.Fatal("no Revelation aura")
	}
	full := float64(100 * core.SpellCritRatingPerCritChance)

	var critRatings []float64
	newSpell := func(id int32, procMask core.ProcMask, flags core.SpellFlag, outcome core.OutcomeApplier, allTargets bool) *core.Spell {
		return m.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 9100 + id},
			SpellSchool: core.SpellSchoolArcane,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    procMask,
			Flags:       flags,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				targets := []*core.Unit{target}
				if allTargets {
					targets = sim.Encounter.TargetUnits
				}
				for _, t := range targets {
					critRatings = append(critRatings, spell.BonusCritRating)
					spell.CalcAndDealDamage(sim, t, 100, outcome)
				}
			},
		})
	}
	bolt := newSpell(1, core.ProcMaskSpellDamage, core.SpellFlagAPL, alwaysHit, false)
	area := newSpell(2, core.ProcMaskSpellDamage, core.SpellFlagAPL, alwaysCrit, true)

	// A hit that lands without a crit gives the charge; its own cast doesn't spend it.
	bolt.Cast(sim, m.CurrentTarget)
	if !buff.IsActive() || bolt.BonusCritRating != full {
		t.Fatalf("after a landed hit: Revelation %v and %.0f crit rating, want it up and %.0f", buff.IsActive(), bolt.BonusCritRating, full)
	}

	// The next cast crits on every target, and spends the charge as it completes.
	sim.CurrentTime += time.Second
	critRatings = critRatings[:0]
	area.Cast(sim, m.CurrentTarget)
	if len(critRatings) != 2 || critRatings[0] != full || critRatings[1] != full {
		t.Errorf("area spell crit rating per target %v, want %.0f on both", critRatings, full)
	}
	if buff.IsActive() || area.BonusCritRating != 0 {
		t.Errorf("after the area cast: Revelation %v and %.0f crit rating, want it spent", buff.IsActive(), area.BonusCritRating)
	}

	for _, c := range []struct {
		name     string
		procMask core.ProcMask
		flags    core.SpellFlag
		want     bool
	}{
		{"an item's magic proc (R-0051)", core.ProcMaskEmpty, 0, true},
		{"Thunderfury's proc", core.ProcMaskSpellProc | core.ProcMaskSpellDamageProc, 0, true},
		{"Flametongue Weapon", core.ProcMaskSpellDamageProc, 0, false},
		{"a totem's spell", core.ProcMaskSpellDamage, core.SpellFlagCastFromTotem, false},
		{"a damage shield", core.ProcMaskEmpty, core.SpellFlagNoPartialResist, false},
		{"a judgement", core.ProcMaskMeleeMHSpecial, core.SpellFlagAPL, false},
	} {
		buff.Deactivate(sim)
		sim.CurrentTime += time.Second
		newSpell(int32(10+len(c.name)), c.procMask, c.flags, alwaysHit, false).Cast(sim, m.CurrentTarget)
		if buff.IsActive() != c.want {
			t.Errorf("%s gives Revelation: %v, want %v", c.name, buff.IsActive(), c.want)
		}
	}
}

// TestForeverRevelationChannel checks that a channel keeps the charge to its end, every missile
// critting meanwhile, and that a charge a missile gives is kept to the channel's end (R-0048).
func TestForeverRevelationChannel(t *testing.T) {
	defer func(old float64) { common.RevelationProcChance = old }(common.RevelationProcChance)
	common.RevelationProcChance = 1

	for _, buffedAtStart := range []bool{true, false} {
		sim, m := newRevelationMage(t, 1)
		buff := m.GetAura("Revelation")
		var missiles *core.Spell
		for rank := len(m.ArcaneMissiles) - 1; rank > 0 && missiles == nil; rank-- {
			missiles = m.ArcaneMissiles[rank]
		}
		full := float64(100 * core.SpellCritRatingPerCritChance)

		sim.CurrentTime = time.Second
		if buffedAtStart {
			buff.Activate(sim)
		}
		sim.CurrentTime = 2 * time.Second
		if !missiles.Cast(sim, m.CurrentTarget) {
			t.Fatal("Arcane Missiles did not cast")
		}
		dot := m.ChanneledDot
		if dot == nil {
			t.Fatal("Arcane Missiles is not channeling")
		}
		end := dot.ExpiresAt()
		var tick *core.Spell
		for _, s := range m.Spellbook {
			if s.ActionID == missiles.ActionID.WithTag(1) {
				tick = s
			}
		}
		sawBuff := false
		for !sim.Step() && sim.CurrentTime < end {
			if buffedAtStart && !buff.IsActive() {
				t.Fatalf("Revelation went at %s, before the channel's end at %s", sim.CurrentTime, end)
			}
			if buff.IsActive() {
				sawBuff = true
				if tick != nil && tick.BonusCritRating != full {
					t.Errorf("a missile at %s has %.0f crit rating, want %.0f", sim.CurrentTime, tick.BonusCritRating, full)
				}
			}
		}
		for m.ChanneledDot == dot && !sim.Step() {
		}
		if !sawBuff {
			t.Errorf("buffed at start %v: Revelation never came up during the channel", buffedAtStart)
		}
		if buff.IsActive() {
			t.Errorf("buffed at start %v: Revelation still up after the channel ended at %s (now %s)", buffedAtStart, end, sim.CurrentTime)
		}
	}
}
