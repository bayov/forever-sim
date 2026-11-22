package dpswarrior

import (
	"testing"

	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
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
