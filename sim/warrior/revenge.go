package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const RevengeRanks = 6

var RevengeSpellId = [RevengeRanks + 1]int32{0, 6572, 6574, 7379, 11600, 11601, 25288}
var RevengeBaseDamage = [RevengeRanks + 1][]float64{{0, 0}, {12, 14}, {18, 22}, {25, 31}, {43, 53}, {64, 78}, {81, 99}}
var RevengeLevel = [RevengeRanks + 1]int{0, 14, 24, 34, 44, 54, 60}

func (warrior *Warrior) registerRevengeSpell(cdTimer *core.Timer) {
	// Ranks from the Ahn'Qiraj books are out under Forever (no source in client 70338).
	aq := warrior.Env.IncludeAQRanks()
	actionID := core.ActionID{SpellID: core.TernaryInt32(aq, 25288, 11601)}
	has2pcDreadnaught := warrior.HasSetBonus(ItemSetDreadnaughtsBattlegear, 2)
	damageLow := core.TernaryFloat64(aq, 81, 64)
	damageHigh := core.TernaryFloat64(aq, 99, 78)
	revengeLevel := core.TernaryFloat64(aq, 60.0, 54.0)
	apCoef := 0.0
	flatThreat := 2.25 * 2 * revengeLevel
	// Forever's rank 5 deals 121 ± 19.7% (97 to 145) (client 70425, spell 11601) plus 28% of
	// attack power, which wowsims/forever fits from logs (3a6c7df46e: 0.277 ± 0.009 over 254 rank 1
	// hits). Its threat is 2.25 times its damage with nothing on top (913a5bcf69, measured in game).
	if warrior.Env.IsForever() && !aq {
		damageLow, damageHigh = 97, 145
		apCoef = 0.28
		flatThreat = 0
	}
	basedamageLow := damageLow + core.TernaryFloat64(has2pcDreadnaught, 75, 0)
	basedamageHigh := damageHigh + core.TernaryFloat64(has2pcDreadnaught, 75, 0)

	warrior.revengeProcAura = warrior.RegisterAura(core.Aura{
		Label:    "Revenge",
		Duration: 5 * time.Second,
		ActionID: actionID,
	})

	warrior.RegisterAura(core.Aura{
		Label:    "Revenge Trigger",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Outcome.Matches(core.OutcomeBlock | core.OutcomeDodge | core.OutcomeParry) {
				warrior.revengeProcAura.Activate(sim)
			}
		},
	})

	warrior.Revenge = warrior.RegisterSpell(DefensiveStance, core.SpellConfig{
		SpellCode:   SpellCode_WarriorRevenge,
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagOffensive,

		RageCost: core.RageCostOptions{
			Cost:   5,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: time.Second * 5,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return warrior.revengeProcAura.IsActive()
		},

		CritDamageBonus: warrior.impale(),

		DamageMultiplier: 1 + 0.2*float64(warrior.Talents.ImprovedRevenge),
		ThreatMultiplier: 2.25,
		FlatThreatBonus:  flatThreat,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(basedamageLow, basedamageHigh) + apCoef*spell.MeleeAttackPower(target)
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			if !result.Landed() {
				spell.IssueRefund(sim)
			}

			warrior.revengeProcAura.Deactivate(sim)
		},
	})
}
