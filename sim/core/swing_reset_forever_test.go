package core

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

func init() {
	RegisterAgentFactory(
		proto.Player_RetributionPaladin{},
		proto.Spec_SpecRetributionPaladin,
		NewFakeSwingCaster,
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_RetributionPaladin)
			if !ok {
				panic("Invalid spec value for Retribution Paladin!")
			}
			player.Spec = playerSpec
		},
	)
}

const (
	swingResetCastTime = time.Second * 2
	swingResetMHSpeed  = 2.6
	swingResetOHSpeed  = 1.8
)

type FakeSwingCaster struct {
	Character

	Resetting *Spell // A hardcast that restarts the melee swing under Forever.
	Keeping   *Spell // The same hardcast, flagged to leave the swing timer running.
	Instant   *Spell // An instant cast.
	Channel   *Spell // A 3 sec channel.
}

func (fc *FakeSwingCaster) GetCharacter() *Character { return &fc.Character }

func (fc *FakeSwingCaster) ApplyTalents()            {}
func (fc *FakeSwingCaster) Reset(_ *Simulation)      {}
func (fc *FakeSwingCaster) OnGCDReady(_ *Simulation) {}

func (fc *FakeSwingCaster) Initialize() {
	register := func(id int32, flags SpellFlag, castTime time.Duration) *Spell {
		return fc.RegisterSpell(SpellConfig{
			ActionID:     ActionID{SpellID: id},
			SpellSchool:  SpellSchoolHoly,
			ProcMask:     ProcMaskEmpty,
			Flags:        flags | SpellFlagNoMetrics,
			Cast:         CastConfig{DefaultCast: Cast{GCD: GCDDefault, CastTime: castTime}},
			ApplyEffects: func(_ *Simulation, _ *Unit, _ *Spell) {},
		})
	}
	fc.Resetting = register(1, 0, swingResetCastTime)
	fc.Keeping = register(2, SpellFlagKeepsSwingTimer, swingResetCastTime)
	fc.Instant = register(3, 0, 0)
	fc.Channel = fc.RegisterSpell(SpellConfig{
		ActionID:    ActionID{SpellID: 4},
		SpellSchool: SpellSchoolHoly,
		ProcMask:    ProcMaskEmpty,
		Flags:       SpellFlagChanneled | SpellFlagNoMetrics,
		Cast:        CastConfig{DefaultCast: Cast{GCD: GCDDefault}},
		Dot: DotConfig{
			Aura:          Aura{Label: "fake channel"},
			NumberOfTicks: 3,
			TickLength:    time.Second,
			OnTick:        func(_ *Simulation, _ *Unit, _ *Dot) {},
		},
		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			spell.Dot(target).Apply(sim)
		},
	})
}

func NewFakeSwingCaster(char *Character, _ *proto.Player) Agent {
	fc := &FakeSwingCaster{
		Character: *char,
	}

	fc.EnableAutoAttacks(fc, AutoAttackOptions{
		MainHand:       Weapon{SwingSpeed: swingResetMHSpeed, BaseDamageMin: 100, BaseDamageMax: 100},
		OffHand:        Weapon{SwingSpeed: swingResetOHSpeed, BaseDamageMin: 50, BaseDamageMax: 50},
		AutoSwingMelee: true,
	})

	return fc
}

// A dual-wielding caster one second into the fight: the MH swung at 0 and is next due at 2.6s,
// the OH at 1.8s.
func setupSwingResetSim(t *testing.T, ruleset proto.Ruleset) (*Simulation, *FakeSwingCaster) {
	t.Helper()
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100, Ruleset: ruleset},
		Raid: &proto.Raid{
			Parties: []*proto.Party{{
				Players: []*proto.Player{{
					Name:      "Paladin",
					Class:     proto.Class_ClassPaladin,
					Level:     60,
					Consumes:  &proto.Consumes{},
					Buffs:     &proto.IndividualBuffs{},
					Spec:      &proto.Player_RetributionPaladin{},
					Equipment: &proto.EquipmentSpec{},
					Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
				}},
				Buffs: &proto.PartyBuffs{},
			}},
		},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()

	fc := sim.Raid.Parties[0].Players[0].(*FakeSwingCaster)
	stepSwingSim(t, sim, time.Second)
	return sim, fc
}

func stepSwingSim(t *testing.T, sim *Simulation, until time.Duration) {
	t.Helper()
	sim.AddPendingAction(&PendingAction{NextActionAt: until, OnAction: func(*Simulation) {}})
	for steps := 0; sim.CurrentTime < until; steps++ {
		if steps > 10000 {
			t.Fatalf("the sim did not reach %v", until)
		}
		sim.Step()
	}
}

// Under Forever completing a cast starts a full swing on each hand, at the attack speed the cast
// completes with, and no swing lands while it is underway (wowsims/forever 4fb2849945,
// 7146db7c30).
func TestForeverCompletedCastResetsSwingTimer(t *testing.T) {
	t.Run("the next swings land one full swing after the cast", func(t *testing.T) {
		sim, fc := setupSwingResetSim(t, proto.Ruleset_RulesetForever)
		aa := &fc.AutoAttacks
		mhPrevious, ohPrevious := aa.mh.lastSwingAt, aa.oh.lastSwingAt

		fc.Resetting.Cast(sim, fc.CurrentTarget)
		castEnd := sim.CurrentTime + swingResetCastTime
		stepSwingSim(t, sim, castEnd)

		if aa.mh.lastSwingAt != mhPrevious || aa.oh.lastSwingAt != ohPrevious {
			t.Fatalf("expected no swing during the cast, got MH at %v and OH at %v", aa.mh.lastSwingAt, aa.oh.lastSwingAt)
		}
		if expected := castEnd + aa.MainhandSwingSpeed(); aa.MainhandSwingAt() != expected {
			t.Fatalf("MH swing: expected %v, got %v", expected, aa.MainhandSwingAt())
		}
		if expected := castEnd + aa.OffhandSwingSpeed(); aa.OffhandSwingAt() != expected {
			t.Fatalf("OH swing: expected %v, got %v", expected, aa.OffhandSwingAt())
		}
	})

	t.Run("haste gained during the cast sets the swing that follows it", func(t *testing.T) {
		sim, fc := setupSwingResetSim(t, proto.Ruleset_RulesetForever)
		aa := &fc.AutoAttacks

		fc.Resetting.Cast(sim, fc.CurrentTarget)
		castEnd := sim.CurrentTime + swingResetCastTime
		stepSwingSim(t, sim, sim.CurrentTime+swingResetCastTime/2)
		fc.MultiplyMeleeSpeed(sim, 1.3)
		stepSwingSim(t, sim, castEnd)

		if expected := castEnd + DurationFromSeconds(swingResetMHSpeed/1.3); aa.MainhandSwingAt() != expected {
			t.Fatalf("MH swing: expected one hasted swing after the cast (%v), got %v", expected, aa.MainhandSwingAt())
		}
	})

	t.Run("an interrupted cast lets the held swings land at once", func(t *testing.T) {
		sim, fc := setupSwingResetSim(t, proto.Ruleset_RulesetForever)
		aa := &fc.AutoAttacks

		fc.Resetting.Cast(sim, fc.CurrentTarget)
		// Past the MH's 2.6s and the OH's 1.8s, so both are held.
		stepSwingSim(t, sim, sim.CurrentTime+swingResetCastTime*19/20)
		if !aa.mh.heldByCast || !aa.oh.heldByCast {
			t.Fatalf("expected both swings held by the cast at %v", sim.CurrentTime)
		}
		fc.interruptHardcast(sim)

		if aa.MainhandSwingAt() != sim.CurrentTime || aa.OffhandSwingAt() != sim.CurrentTime {
			t.Fatalf("expected both swings at %v, got MH %v and OH %v", sim.CurrentTime, aa.MainhandSwingAt(), aa.OffhandSwingAt())
		}
	})

	// A cast that keeps the swing timer, and a channel, let a swing land when it comes due.
	for _, c := range []struct {
		label string
		spell func(fc *FakeSwingCaster) *Spell
	}{
		{"a cast that keeps the swing timer", func(fc *FakeSwingCaster) *Spell { return fc.Keeping }},
		{"a channel", func(fc *FakeSwingCaster) *Spell { return fc.Channel }},
	} {
		t.Run(c.label+" lets the swings land during it", func(t *testing.T) {
			sim, fc := setupSwingResetSim(t, proto.Ruleset_RulesetForever)
			aa := &fc.AutoAttacks
			mhDue := aa.MainhandSwingAt()

			if !c.spell(fc).Cast(sim, fc.CurrentTarget) {
				t.Fatalf("the spell didn't cast")
			}
			stepSwingSim(t, sim, sim.CurrentTime+swingResetCastTime*19/20)
			if !fc.IsCasting(sim) && !fc.IsChanneling(sim) {
				t.Fatalf("expected the cast or channel underway at %v", sim.CurrentTime)
			}

			if aa.mh.lastSwingAt != mhDue {
				t.Fatalf("MH swing: expected it to land during the cast at %v, got %v", mhDue, aa.mh.lastSwingAt)
			}
			if expected := mhDue + aa.MainhandSwingSpeed(); aa.MainhandSwingAt() != expected {
				t.Fatalf("MH swing: expected the next one at %v, got %v", expected, aa.MainhandSwingAt())
			}
		})
	}

	t.Run("a cast ending as a swing comes due completes before the swing", func(t *testing.T) {
		sim, fc := setupSwingResetSim(t, proto.Ruleset_RulesetClassic)

		// A cast that ends as the MH swing after next comes due. The weapon attacks at an
		// instant go before its pending actions, so stepping to it stops after the swing.
		stepSwingSim(t, sim, fc.AutoAttacks.MainhandSwingAt())
		mhDue := fc.AutoAttacks.MainhandSwingAt()
		stepSwingSim(t, sim, mhDue-swingResetCastTime)
		// The empty rotation pauses the GCD 50ms at a time.
		fc.GCD.Reset()
		if !fc.Keeping.Cast(sim, fc.CurrentTarget) {
			t.Fatalf("the cast didn't start")
		}
		stepSwingSim(t, sim, mhDue)

		if fc.Hardcast.Expires != startingCDTime {
			t.Fatalf("expected the cast to have completed at %v, before the swing ran the rotation", sim.CurrentTime)
		}
		if fc.AutoAttacks.mh.lastSwingAt != mhDue {
			t.Fatalf("MH swing: expected it at the cast's end %v, got %v", mhDue, fc.AutoAttacks.mh.lastSwingAt)
		}
	})

	t.Run("an instant cast leaves the swing timer alone", func(t *testing.T) {
		sim, fc := setupSwingResetSim(t, proto.Ruleset_RulesetForever)
		aa := &fc.AutoAttacks
		mhAt, ohAt := aa.MainhandSwingAt(), aa.OffhandSwingAt()

		fc.Instant.Cast(sim, fc.CurrentTarget)

		if aa.MainhandSwingAt() != mhAt || aa.OffhandSwingAt() != ohAt {
			t.Fatalf("expected the swings at %v and %v, got %v and %v", mhAt, ohAt, aa.MainhandSwingAt(), aa.OffhandSwingAt())
		}
	})

	t.Run("Classic leaves the swing timer to the spells that reset it", func(t *testing.T) {
		sim, fc := setupSwingResetSim(t, proto.Ruleset_RulesetClassic)
		aa := &fc.AutoAttacks

		fc.Resetting.Cast(sim, fc.CurrentTarget)
		castEnd := sim.CurrentTime + swingResetCastTime
		stepSwingSim(t, sim, castEnd)

		if aa.mh.lastSwingAt != castEnd {
			t.Fatalf("MH swing: expected it at the cast's end %v, got %v", castEnd, aa.mh.lastSwingAt)
		}
	})
}
