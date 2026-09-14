package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (shaman *Shaman) registerWaterShieldSpell() {
	if !shaman.Talents.WaterShield {
		return
	}

	actionID := core.ActionID{SpellID: 52127}
	manaMetrics := shaman.NewManaMetrics(actionID)
	globes := int32(3)

	// TODO: "Only one globe will activate every few seconds", the tooltip never said how long.
	// Lightning Shield's ICD is used until beta shows otherwise.
	icd := core.Cooldown{
		Timer:    shaman.NewTimer(),
		Duration: time.Millisecond * 3500,
	}

	consumeGlobe := func(sim *core.Simulation) {
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
			if result.Landed() {
				consumeGlobe(sim)
			}
		},
		OnHealDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.DidCrit() {
				consumeGlobe(sim)
			}
		},
	})

	shaman.WaterShield = shaman.RegisterSpell(core.SpellConfig{
		SpellCode: SpellCode_ShamanWaterShield,
		ActionID:  actionID,
		ProcMask:  core.ProcMaskEmpty,
		Flags:     core.SpellFlagAPL | SpellFlagShaman,

		ManaCost: core.ManaCostOptions{
			BaseCost: .06,
		},
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
