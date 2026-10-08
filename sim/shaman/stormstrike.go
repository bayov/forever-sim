package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Stormstrike under Forever (beta client): 125 Mana, 8 sec cooldown, and the strike marks the
// target so the shaman's next Lightning Bolt, Chain Lightning or Earth Shock on it deals 20%
// more. The mark lasts 12 sec and is the shaman's own, so another shaman's Stormstrike gives
// nothing (the external debuff is skipped under Forever in core/debuffs.go).
//
// Under Classic it keeps the 21% base mana cost, the 20 sec cooldown and the shared debuff
// with two Nature charges.
const StormstrikeForeverBonus = 1.2

func (shaman *Shaman) registerStormstrikeSpell() {
	if !shaman.Talents.Stormstrike {
		return
	}

	forever := shaman.Env.IsForever()
	actionID := core.ActionID{SpellID: 17364}

	var stormStrikeAuras core.AuraArray
	if forever {
		stormStrikeAuras = shaman.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			return target.RegisterAura(core.Aura{
				Label:    "Stormstrike",
				ActionID: actionID,
				Duration: time.Second * 12,
			})
		})
		shaman.stormstrikeMarks = stormStrikeAuras
	} else {
		stormStrikeAuras = shaman.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			return core.StormstrikeAura(target)
		})
	}

	manaCost := core.ManaCostOptions{BaseCost: .21}
	cooldown := time.Second * 20
	if forever {
		manaCost = core.ManaCostOptions{FlatCost: 125}
		cooldown = time.Second * 8
	}

	shaman.Stormstrike = shaman.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_ShamanStormstrike,
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       SpellFlagShaman | core.SpellFlagAPL | core.SpellFlagMeleeMetrics,

		ManaCost: manaCost,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: cooldown,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		// Flat "+N damage" effects like Zandalarian Hero Medallion add to it, as to a white hit.
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := shaman.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if result.Landed() {
				aura := stormStrikeAuras.Get(target)
				aura.Activate(sim)
				if aura.MaxStacks > 0 {
					aura.SetStacks(sim, aura.MaxStacks)
				}
			}
		},
	})
}

// spendStormstrike returns the damage multiplier the shaman's Lightning Bolt, Chain
// Lightning or Earth Shock gets on this target, and uses the mark up. Under Classic the
// debuff on the target already raises the Nature damage it takes, so this is 1.
func (shaman *Shaman) spendStormstrike(sim *core.Simulation, target *core.Unit) float64 {
	if shaman.stormstrikeMarks == nil {
		return 1
	}

	aura := shaman.stormstrikeMarks.Get(target)
	if !aura.IsActive() {
		return 1
	}
	aura.Deactivate(sim)
	return StormstrikeForeverBonus
}
