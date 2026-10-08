package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (shaman *Shaman) registerWaterShieldSpell() {
	if !shaman.Talents.WaterShield {
		return
	}

	// The beta client's spell (build 1.60.1.69876). It costs no mana and has a 15 sec
	// cooldown, so once its three globes are spent the shaman waits for the recast.
	actionID := core.ActionID{SpellID: 408510}
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

	// Raid damage as a steady stream of spell hits. Spells cannot be dodged or parried,
	// so unlike a boss meleeing the shaman this feeds Water Shield without also firing
	// Improved Stormstrike's reset or parry haste.
	//
	// With a variation, each iteration picks its own rate between the rate minus the
	// variation and the rate plus it, the way Duration +/- picks each fight's length.
	if shaman.RaidDamageHitsPerMinute > 0 {
		shaman.RegisterResetEffect(func(sim *core.Simulation) {
			hitsPerMinute := shaman.RaidDamageHitsPerMinute
			if variation := shaman.RaidDamageHitsPerMinuteVariation; variation > 0 {
				hitsPerMinute += (sim.RandomFloat("Raid Damage Hits")*2 - 1) * variation
			}
			if hitsPerMinute <= 0 {
				return
			}
			period := time.Duration(float64(time.Minute) / hitsPerMinute)
			core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				Period: period,
				OnAction: func(sim *core.Simulation) {
					if shaman.WaterShieldAura.IsActive() {
						consumeGlobe(sim)
					}
				},
			})
		})
	}

	shaman.WaterShield = shaman.RegisterSpell(core.SpellConfig{
		SpellCode: SpellCode_ShamanWaterShield,
		ActionID:  actionID,
		ProcMask:  core.ProcMaskEmpty,
		Flags:     core.SpellFlagAPL | SpellFlagShaman,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Second * 15,
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
