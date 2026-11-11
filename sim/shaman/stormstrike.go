package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Stormstrike under Forever (beta client): 8% of base mana (125 Mana before client 1.60.1.70425), 8 sec cooldown, and the strike marks the
// target so the shaman's next Lightning Bolt, Chain Lightning or Earth Shock on it deals 20%
// more. The mark lasts 12 sec and is the shaman's own, so another shaman's Stormstrike gives
// nothing (the external debuff is skipped under Forever in core/debuffs.go).
//
// Under Classic it keeps the 21% base mana cost, the 20 sec cooldown and the shared debuff
// with two Nature charges.
//
// The November 2 development notes (post 9, for launch) make it 150% normalized weapon damage (was
// 100%) and cut the cooldown to 6 sec (was 8). Client 1.60.1.70753 had neither; hotfix push 113380
// (wago.tools/hotfixes, region 91) puts both in client data: SpellEffect 1368308 and
// SpellCooldowns 55241 (tier 1).
//
// Blizzard's class tuning of 2026-11-10 (hotfix article and Kaivax's "Class Tuning Incoming -
// November 10") takes it back to 100% weapon damage and an 8 sec cooldown. No server row carries
// that yet (SpellEffect 1368308 still 150, SpellCooldowns 55241 still 6000 ms at push 113380;
// client 1.60.1.70888 unchanged), so these are notes-only values (standards 37).
const StormstrikeForeverBonus = 1.2

// StormstrikeWeaponDamagePercent is the share of normalized weapon damage the strike deals.
const StormstrikeWeaponDamagePercent = 1.0

// StormstrikeSpeed is the speed Forever's Stormstrike multiplies AP / 14 by.
//
// Forever's spell data says "Normalized Weapon Damage" (Classic's says "Deal Weapon Damage"),
// so the bonus is AP / 14 times a fixed speed for the weapon type instead of the weapon's own
// speed. Beta tests on a level 30 Orc hit harder than the usual 3.3 for two-handers and 2.4 for
// one-handers, and fit 3.6 and 2.7 (see shaman_audit.md 4.1). That's 0.3 more than the usual
// normalized speed.
//
// Daggers don't follow that rule. 27 Stormstrikes with Poniard fit 1.79 +- 0.02, where 1.7 +
// 0.3 would be 2.0 (beta_results.txt, 2026-10-11). So daggers get 1.8.
func StormstrikeSpeed(weapon *core.Weapon) float64 {
	if weapon.NormalizedSwingSpeed == 1.7 {
		return 1.8
	}
	return weapon.NormalizedSwingSpeed + 0.3
}

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
				Label:    "Stormstrike - " + shaman.Label,
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
		// 8% of base mana since client 1.60.1.70425 (125 before). 8 sec since the 2026-11-10
		// class tuning (6 sec from push 113380 until then).
		manaCost = core.ManaCostOptions{BaseCost: .08}
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
			var baseDamage float64
			if forever {
				baseDamage = shaman.stormstrikeWeaponDamage(sim, spell.MeleeAttackPower(target))
			} else {
				baseDamage = shaman.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
			}
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

// stormstrikeWeaponDamage rolls Stormstrike's weapon damage under Forever, see
// StormstrikeSpeed.
func (shaman *Shaman) stormstrikeWeaponDamage(sim *core.Simulation, attackPower float64) float64 {
	weapon := shaman.AutoAttacks.MH()
	return StormstrikeWeaponDamagePercent * (weapon.BaseDamage(sim) + StormstrikeSpeed(weapon)*attackPower/weapon.AttackPowerPerDPS)
}

// stormstrikeBonus returns the damage multiplier the shaman's Lightning Bolt, Chain Lightning
// or Earth Shock gets on this target from our Stormstrike mark. Under Classic the debuff on the
// target already raises the Nature damage it takes, so this is 1.
func (shaman *Shaman) stormstrikeBonus(target *core.Unit) float64 {
	if shaman.stormstrikeMarks == nil || !shaman.stormstrikeMarks.Get(target).IsActive() {
		return 1
	}
	return StormstrikeForeverBonus
}

// applyStormstrike raises a result's damage by our Stormstrike mark on the target.
//
// We add the mark after the damage calc, not through the spell's DamageMultiplier, because the
// game leaves it out of the log's raw damage field. An Earth Shock right after a Stormstrike
// logs the same raw as one without, and only its amount is 20% higher (Nov 5-8 logs: amount
// over raw is 1.004 with no Stormstrike before it and 1.202 right after one). So the mark works
// like a modifier on the target, and the result's PreOutcomeDamage stays the raw the log shows.
// The damage dealt and the threat are the same either way, so we redo the threat here.
func (shaman *Shaman) applyStormstrike(spell *core.Spell, target *core.Unit, result *core.SpellResult) {
	bonus := shaman.stormstrikeBonus(target)
	if bonus == 1 {
		return
	}
	result.Damage *= bonus
	result.Threat = spell.ThreatFromDamage(result.Outcome, result.Damage)
}

// spendStormstrike uses the mark on the target up once the spell's result is in. Since client
// 1.60.1.70425 a spell that misses leaves it (the 15 October development notes: "Stormstrike's
// debuff is no longer consumed if the spell does not land").
func (shaman *Shaman) spendStormstrike(sim *core.Simulation, target *core.Unit, result *core.SpellResult) {
	if shaman.stormstrikeMarks == nil || !result.Landed() {
		return
	}
	shaman.stormstrikeMarks.Get(target).Deactivate(sim)
}
