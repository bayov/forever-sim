package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// The mana caster Faerie Fire (rank 4, 9907) costs.
const faerieFireManaCost = 115

func (druid *Druid) registerFaerieFireSpell() {
	if druid.Env.IsForever() {
		druid.registerForeverFaerieFireSpell()
		return
	}

	spellCode := SpellCode_DruidFaerieFire
	actionID := core.ActionID{SpellID: 9907}
	manaCostOptions := core.ManaCostOptions{
		FlatCost: faerieFireManaCost,
	}
	gcd := core.GCDDefault
	ignoreHaste := false
	cd := core.Cooldown{}
	flatThreatBonus := 2. * 54
	flags := core.SpellFlagNone
	formMask := Humanoid | Moonkin

	druid.FaerieFireAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.FaerieFireAura(target)
	})

	if druid.InForm(Cat|Bear) && druid.Level >= 25 {
		spellCode = SpellCode_DruidFaerieFireFeral
		actionID = core.ActionID{SpellID: 17392}
		manaCostOptions = core.ManaCostOptions{}
		gcd = time.Second
		ignoreHaste = true
		formMask = Cat | Bear
		cd = core.Cooldown{
			Timer:    druid.NewTimer(),
			Duration: time.Second * 6,
		}
		druid.FaerieFireAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			return core.FaerieFireFeralAura(target)
		})
	}
	flags |= core.SpellFlagAPL | core.SpellFlagResetAttackSwing

	druid.FaerieFire = druid.RegisterSpell(formMask, core.SpellConfig{
		SpellCode:   spellCode,
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolNature,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       flags,

		ManaCost: manaCostOptions,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: gcd,
			},
			IgnoreHaste: ignoreHaste,
			CD:          cd,
		},

		ThreatMultiplier: 1,
		FlatThreatBonus:  flatThreatBonus,
		DamageMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				druid.FaerieFireAuras.Get(target).Activate(sim)
			}

			if druid.InForm(Humanoid | Moonkin) {
				druid.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime, false)
			}
		},

		RelatedAuras: []core.AuraArray{druid.FaerieFireAuras},
	})
}

// Under Forever Faerie Fire (Feral) is gone: 16857 and 17390 to 17392 have no rows in the client
// (1.60.1.70338 to 70685), and foreverchanges.pro's spellbook says "Not in Forever". Instead the
// caster Faerie Fire can be cast in Cat, Bear and Dire Bear Form too, where it "costs no resources"
// (the development notes, orchestration/audit/ref/client_update/sources/devnotes.txt). It is the
// same spell in every form: 505 armor for 40 sec (the same debuff as any other druid's, one per
// target), a 1.5 sec global cooldown and no cooldown, 115 mana in caster and Moonkin Form.
//
// Forever's Faerie Fire no longer resets the swing timer (development notes, 2026-10-01, build
// 1.60.1.70170): the client dropped InterruptFlags 8, the auto attack reset, from every caster
// rank (770, 778, 9749, 9907).
func (druid *Druid) registerForeverFaerieFireSpell() {
	druid.FaerieFireAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.FaerieFireAura(target)
	})

	// The form the cost was last set for: free in Cat and Bear Form. A cast and the rotation's
	// can-cast check both run ExtraCastCondition before they look at the cost.
	free := false

	druid.FaerieFire = druid.RegisterSpell(Humanoid|Moonkin|Cat|Bear, core.SpellConfig{
		SpellCode:   SpellCode_DruidFaerieFire,
		ActionID:    core.ActionID{SpellID: 9907},
		SpellSchool: core.SpellSchoolNature,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: faerieFireManaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
			if inForm := druid.InForm(Cat | Bear); inForm != free {
				free = inForm
				druid.FaerieFire.Cost.FlatModifier += core.TernaryInt32(free, -faerieFireManaCost, faerieFireManaCost)
			}
			return true
		},

		ThreatMultiplier: 1,
		FlatThreatBonus:  2. * 54,
		DamageMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				druid.FaerieFireAuras.Get(target).Activate(sim)
			}
		},

		RelatedAuras: []core.AuraArray{druid.FaerieFireAuras},
	})
}

func (druid *Druid) ShouldFaerieFire(sim *core.Simulation, target *core.Unit) bool {
	if druid.FaerieFire == nil {
		return false
	}

	if !druid.FaerieFire.IsReady(sim) {
		return false
	}

	debuff := druid.FaerieFireAuras.Get(target)
	return !debuff.IsActive() || debuff.RemainingDuration(sim) < time.Second*4
}
