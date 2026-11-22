package dpswarrior

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// hitSpell registers a spell on caster that deals 10 Fire to its target.
func hitSpell(caster *core.Unit, id int32) *core.Spell {
	return caster.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: id},
		SpellSchool:      core.SpellSchoolFire,
		DefenseType:      core.DefenseTypeMagic,
		ProcMask:         core.ProcMaskEmpty,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, 10, spell.OutcomeAlwaysHit)
		},
	})
}

// Flame Wrath's fire shield hits back whoever hits the wearer, but not the wearer itself: damage
// the wearer dealt to itself set the shield off on the wearer again without end (a stack overflow
// in the two-hand warrior gear search, 2026-11-22).
func TestFlameWrathShieldIgnoresSelfDamage(t *testing.T) {
	sim, war := newWarriorWithMainHand(t, proto.Ruleset_RulesetForever, common.FlameWrath)
	shield := war.GetAura("Flame Wrath")
	if shield == nil {
		t.Fatal("no Flame Wrath shield aura")
	}
	shield.Activate(sim)
	reflect := war.GetSpell(core.ActionID{SpellID: 461152})
	hitSpell(&war.Unit, 1).Cast(sim, &war.Unit)
	if n := reflect.SpellMetrics()[war.Unit.UnitIndex].Casts; n != 0 {
		t.Fatalf("the shield hit back %d times on self damage, want 0", n)
	}
	target := war.CurrentTarget
	hitSpell(target, 2).Cast(sim, &war.Unit)
	casts := int32(0)
	for _, m := range reflect.SpellMetrics() {
		casts += m.Casts
	}
	if casts != 1 {
		t.Fatalf("the shield hit back %d times on the target's hit, want 1", casts)
	}
}

// Headmaster's Charge's Use gives +20 Intellect for 15 min on a 10 min cooldown (spell 18264,
// client 1.60.1.71037).
func TestHeadmastersChargeUse(t *testing.T) {
	sim, war := newWarriorWithMainHand(t, proto.Ruleset_RulesetForever, common.HeadmastersCharge)
	use := war.GetSpell(core.ActionID{ItemID: common.HeadmastersCharge})
	if use == nil {
		t.Fatal("no Headmaster's Charge use spell")
	}
	before := war.GetStat(stats.Intellect)
	use.Cast(sim, &war.Unit)
	if got := war.GetStat(stats.Intellect) - before; got != 20 {
		t.Fatalf("Intellect rose by %v, want 20", got)
	}
	if got := use.CD.Duration; got != 10*time.Minute {
		t.Fatalf("cooldown %v, want 10m", got)
	}
	aura := war.GetAuraByID(core.ActionID{ItemID: common.HeadmastersCharge})
	if aura == nil || aura.Duration != 15*time.Minute {
		t.Fatalf("aura %v, want a 15m aura", aura)
	}
}

// Axe of the Deep Woods' Wrath (18104) deals 105 to 147 Nature under Forever (126, variance
// 0.333, client 1.60.1.71037), and isn't simulated under Classic.
func TestAxeOfTheDeepWoodsProc(t *testing.T) {
	sim, war := newWarriorWithMainHand(t, proto.Ruleset_RulesetForever, common.AxeOfTheDeepWoods)
	wrath := war.GetSpell(core.ActionID{SpellID: 18104})
	if wrath == nil {
		t.Fatal("no Wrath proc with Axe of the Deep Woods")
	}
	lo, hi := procRange(sim, war, wrath, 2000)
	if lo < 105-1e-6 || hi > 147+1e-6 || lo > 105.5 || hi < 146.5 {
		t.Errorf("Wrath %.1f to %.1f, want 105 to 147", lo, hi)
	}
	_, era := newWarriorWithMainHand(t, proto.Ruleset_RulesetClassic, common.AxeOfTheDeepWoods)
	if era.GetSpell(core.ActionID{SpellID: 18104}) != nil {
		t.Error("Classic has a Wrath proc, want none")
	}
}
