package mage

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const IgniteTicks = 2

// Ignite: Fire crits burn the target for 8% of the crit's damage per point over 4 sec (2 ticks).
//
// The model is the SoD sim's, which Forever's client code comes from: a new crit adds its share
// to what is left of the current Ignite and starts the 4 sec over, with no cap on how much it
// holds. Classic Era's 1.12 servers stopped adding after 5 crits, and the fork did too. Which one
// Forever uses is open (notes.md).
//
// Since the 2026-09-24 build Ignite "no longer counts damage bonuses twice". The crit it comes
// from already has every damage bonus, on the mage and on the target, so the ticks take none of
// them again. The SoD sim still multiplied the mage's bonuses in, and the fork both sides.
func (mage *Mage) applyIgnite() {
	if mage.Talents.Ignite == 0 {
		return
	}

	igniteMultiplier := 0.08 * float64(mage.Talents.Ignite)

	mage.Ignite = mage.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 12654},
		SpellCode:   SpellCode_MageIgnite,
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellProc,
		Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagIgnoreModifiers | SpellFlagMage,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Cast: core.CastConfig{
			IgnoreHaste: true,
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Ignite",
			},
			NumberOfTicks: IgniteTicks,
			TickLength:    time.Second * 2,
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.Dot(target).ApplyOrReset(sim)
		},
	})

	core.MakePermanent(mage.RegisterAura(core.Aura{
		Label: "Ignite Talent",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.ProcMask.Matches(core.ProcMaskSpellDamage) || !spell.SpellSchool.Matches(core.SpellSchoolFire) || !result.DidCrit() {
				return
			}

			dot := mage.Ignite.Dot(result.Target)
			newDamage := result.Damage * igniteMultiplier
			outstanding := 0.0
			if dot.IsActive() {
				outstanding = dot.SnapshotBaseDamage * float64(dot.NumberOfTicks-dot.TickCount)
			}
			dot.Snapshot(result.Target, (outstanding+newDamage)/IgniteTicks, false)
			mage.Ignite.Cast(sim, result.Target)
		},
	}))
}
