//go:build with_db

package common_test

import (
	"testing"

	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/common/itemhelpers"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestForeverWeaponStrikeChain checks which hits let a weapon proc's melee strike roll weapon procs
// again (R-0040, EV-0145): under Forever only a strike a main-hand white swing cast does; one cast by
// an ability, an off-hand swing or another strike carries Suppress Weapon Procs. Classic lets each
// one roll.
func TestForeverWeaponStrikeChain(t *testing.T) {
	for _, ruleset := range []proto.Ruleset{proto.Ruleset_RulesetForever, proto.Ruleset_RulesetClassic} {
		sim, m := newMageVs(t, ruleset, proto.MobType_MobTypeHumanoid, proto.ItemSlot_ItemSlotMainHand, common.HanzoSword)
		if m.GetSpell(core.ActionID{SpellID: 16405}) == nil {
			t.Fatalf("%s: Hanzo Sword's strike 16405 is not registered", ruleset)
		}
		var suppressed []bool
		strike := m.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 16405, Tag: 9},
			DefenseType: core.DefenseTypeMelee,
			ProcMask:    core.ProcMaskMeleeSpecial,
			ApplyEffects: func(_ *core.Simulation, _ *core.Unit, spell *core.Spell) {
				suppressed = append(suppressed, spell.Flags.Matches(core.SpellFlagSuppressWeaponProcs))
			},
		})
		ability := m.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 1},
			DefenseType: core.DefenseTypeMelee,
			ProcMask:    core.ProcMaskMeleeMHSpecial,
		})
		mainHand := m.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 4},
			DefenseType: core.DefenseTypeMelee,
			ProcMask:    core.ProcMaskMeleeMHAuto,
		})
		offHand := m.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 2},
			DefenseType: core.DefenseTypeMelee,
			ProcMask:    core.ProcMaskMeleeOHAuto,
		})

		forever := ruleset == proto.Ruleset_RulesetForever
		for _, c := range []struct {
			name           string
			trigger        *core.Spell
			wantSuppressed bool
		}{
			{"a main-hand white swing", mainHand, false},
			{"a main-hand ability", ability, forever},
			{"an off-hand swing", offHand, forever},
			{"another strike", strike, forever},
		} {
			suppressed = suppressed[:0]
			itemhelpers.CastWeaponProcSpell(sim, &m.Character, strike, c.trigger, m.CurrentTarget)
			if len(suppressed) != 1 {
				t.Fatalf("%s, %s: the strike was cast %d times, want 1", ruleset, c.name, len(suppressed))
			}
			if suppressed[0] != c.wantSuppressed {
				t.Errorf("%s: the strike off %s suppresses weapon procs: %v, want %v", ruleset, c.name, suppressed[0], c.wantSuppressed)
			}
			if strike.Flags.Matches(core.SpellFlagSuppressWeaponProcs) {
				t.Errorf("%s, %s: the strike keeps Suppress Weapon Procs after its cast", ruleset, c.name)
			}
		}
	}
}

// TestWeaponProcRollsTwice checks that a weapon proc trigger (one that skips Suppress Weapon Procs
// hits) rolls twice on a hit flagged SpellFlagRollsWeaponProcsTwice, Forever's Judgement of Command
// (EV-0155, EV-0156), and that any other trigger hears it once.
func TestWeaponProcRollsTwice(t *testing.T) {
	sim, m := newMageVs(t, proto.Ruleset_RulesetForever, proto.MobType_MobTypeHumanoid, proto.ItemSlot_ItemSlotMainHand, common.HanzoSword)
	weaponProcs, equipProcs := 0, 0
	weaponProbe, equipProbe := &core.Aura{}, &core.Aura{}
	core.ApplyProcTriggerCallback(&m.Unit, weaponProbe, core.ProcTrigger{
		Name:              "Weapon Probe",
		Callback:          core.CallbackOnSpellHitDealt,
		ProcMask:          core.ProcMaskMelee,
		SpellFlagsExclude: core.SpellFlagSuppressWeaponProcs,
		Outcome:           core.OutcomeLanded,
		Handler:           func(*core.Simulation, *core.Spell, *core.SpellResult) { weaponProcs++ },
	})
	core.ApplyProcTriggerCallback(&m.Unit, equipProbe, core.ProcTrigger{
		Name:     "Equip Probe",
		Callback: core.CallbackOnSpellHitDealt,
		ProcMask: core.ProcMaskMelee,
		Outcome:  core.OutcomeLanded,
		Handler:  func(*core.Simulation, *core.Spell, *core.SpellResult) { equipProcs++ },
	})

	for _, c := range []struct {
		flags core.SpellFlag
		want  int
	}{{core.SpellFlagNone, 1}, {core.SpellFlagRollsWeaponProcsTwice, 2}} {
		spell := m.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 3, Tag: int32(c.want)},
			DefenseType: core.DefenseTypeMelee,
			ProcMask:    core.ProcMaskMeleeMHSpecial,
			Flags:       c.flags,
		})
		result := &core.SpellResult{Target: m.CurrentTarget, Outcome: core.OutcomeHit, Damage: 100}
		weaponProcs, equipProcs = 0, 0
		weaponProbe.OnSpellHitDealt(weaponProbe, sim, spell, result)
		equipProbe.OnSpellHitDealt(equipProbe, sim, spell, result)
		if weaponProcs != c.want || equipProcs != 1 {
			t.Errorf("flags %v: the weapon proc rolled %d times and the equip proc %d, want %d and 1", c.flags, weaponProcs, equipProcs, c.want)
		}
	}
}

// TestForeverWarbladeSuppressesCasterProcs checks that Warblade of Caer Darrow's on-hit damage
// reaches none of the wearer's procs (EV-0165).
func TestForeverWarbladeSuppressesCasterProcs(t *testing.T) {
	_, m := newMageVs(t, proto.Ruleset_RulesetForever, proto.MobType_MobTypeHumanoid, proto.ItemSlot_ItemSlotMainHand, common.WarbladeOfCaerDarrow)
	proc := m.GetSpell(core.ActionID{SpellID: 1298499})
	if proc == nil || !proc.Flags.Matches(core.SpellFlagNoProcs) {
		t.Fatal("Warblade of Caer Darrow's 1298499 is missing or reaches the wearer's procs")
	}
}
