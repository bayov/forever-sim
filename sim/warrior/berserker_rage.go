package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (warrior *Warrior) registerBerserkerRageSpell() {
	// Forever trains it at 30 (Classic 32).
	if warrior.Level < core.TernaryInt32(warrior.Env.IsForever(), 30, 32) {
		return
	}

	actionID := core.ActionID{SpellID: 18499}
	rageMetrics := warrior.NewRageMetrics(actionID)
	instantRage := 5 * float64(warrior.Talents.ImprovedBerserkerRage) // client 70338: 5 and 10

	// Under Forever every hit and tick taken gives double Rage while it is up (wowsims/forever
	// f9f9f21883; the game files state no amount, so the double is an assumption theirs keeps a
	// TODO on). Classic adds about 2 Rage a melee hit.
	if warrior.Env.IsForever() {
		warrior.BerserkerRageAura = warrior.RegisterAura(core.Aura{
			Label:    "Berserker Rage",
			ActionID: actionID,
			Duration: time.Second * 10,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				warrior.AddDamageTakenRageMultiplier(2)
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				warrior.AddDamageTakenRageMultiplier(0.5)
			},
		})
	} else {
		warrior.BerserkerRageAura = warrior.registerClassicBerserkerRageAura(actionID, rageMetrics)
	}

	warrior.BerserkerRage = warrior.RegisterSpell(BerserkerStance, core.SpellConfig{
		ActionID: actionID,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Second * 30,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			if instantRage > 0 {
				warrior.AddRage(sim, instantRage, rageMetrics)
			}
			warrior.BerserkerRageAura.Activate(sim)
		},
	})

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: warrior.BerserkerRage.Spell,
		Type:  core.CooldownTypeSurvival,
	})
}

func (warrior *Warrior) registerClassicBerserkerRageAura(actionID core.ActionID, rageMetrics *core.ResourceMetrics) *core.Aura {
	return warrior.RegisterAura(core.Aura{
		Label:    "Berserker Rage",
		ActionID: actionID,
		Duration: time.Second * 10,

		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || result.Damage <= 0 {
				return
			}

			if spell.ProcMask.Matches(core.ProcMaskMeleeOrRanged) {
				// For melee attacks we find that it gives around 2.0 extra rage regardless of attacker level
				// This may give less rage for fast attacks, but we default to 2.0 for now
				warrior.AddRage(sim, 2.0, rageMetrics)
			} else if spell.ProcMask.Matches(core.ProcMaskSpellDamage) {
				// Spell attacks generally give 1 - 2 times unmodified damage as additional rage.
				rageConversionDamageTaken := core.GetRageConversion(spell.Unit.Level)
				generatedRage := result.RawDamage() * 2.5 / rageConversionDamageTaken
				generatedRage *= 1.0 // Using 1.0 because we don't know why it gives more sometimes
				warrior.AddRage(sim, generatedRage, rageMetrics)
			}
		},
	})
}
