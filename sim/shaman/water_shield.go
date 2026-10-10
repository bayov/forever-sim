package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (shaman *Shaman) registerWaterShieldSpell() {
	if !shaman.Talents.WaterShield {
		return
	}

	// The beta client's spell (build 1.60.1.69876). It costs no mana. Its tooltip said it
	// had a 15 sec cooldown, but the 2026-10-08 beta build notes say it never had one, so
	// we recast it as soon as its three globes are spent.
	actionID := core.ActionID{SpellID: 408510}
	manaMetrics := shaman.NewManaMetrics(actionID)
	globes := int32(3)

	// We use a globe when an enemy lands a direct hit on us or we crit with a heal, at most one
	// globe every 3.5 sec.
	//
	// Like Lightning Shield, melee, ranged attacks and harmful spells all count, AoE too, but
	// damage over time ticks don't. The tooltip only says "Only one globe will activate every
	// few seconds", but the client (70291) gives the spell a 3500 ms proc cooldown, like
	// Lightning Shield's, and the same proc flags plus our heals (0x262a8).
	icd := core.Cooldown{
		Timer:    shaman.NewTimer(),
		Duration: time.Millisecond * 3500,
	}

	consumeGlobe := func(sim *core.Simulation, _ *core.Unit) {
		if !icd.IsReady(sim) {
			return
		}

		icd.Use(sim)
		shaman.AddMana(sim, shaman.MaxMana()*0.02, manaMetrics)
		shaman.WaterShieldAura.RemoveStack(sim)
	}

	shaman.WaterShieldAura = shaman.RegisterAura(core.Aura{
		Label:     "Water Shield",
		ActionID:  actionID,
		Duration:  time.Minute * 10,
		MaxStacks: globes,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.SetStacks(sim, globes)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			if shaman.ActiveShieldAura.ActionID == aura.ActionID {
				shaman.ActiveShieldAura = nil
				shaman.ActiveShield = nil
			}
		},
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Landed() && spell.Unit.IsOpponent(&shaman.Unit) {
				consumeGlobe(sim, spell.Unit)
			}
		},
		OnHealDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.DidCrit() {
				consumeGlobe(sim, nil)
			}
		},
	})
	shaman.shieldHitTaken[shaman.WaterShieldAura] = consumeGlobe

	shaman.WaterShield = shaman.RegisterSpell(core.SpellConfig{
		SpellCode: SpellCode_ShamanWaterShield,
		ActionID:  actionID,
		ProcMask:  core.ProcMaskEmpty,
		Flags:     core.SpellFlagAPL | SpellFlagShaman,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			if shaman.ActiveShieldAura != nil {
				shaman.ActiveShieldAura.Deactivate(sim)
			}
			shaman.ActiveShield = spell
			shaman.ActiveShieldAura = shaman.WaterShieldAura
			shaman.ActiveShieldAura.Activate(sim)
		},
	})
}
