package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Report assumption warr_rend_ap, user ruling 2026-10-15: under Forever each Rend tick adds 2% of
// attack power, about 14% over a full Rend at 60. The client holds no coefficient. wowsims/forever
// 1172f1c745 measured 7 + 0.02 x attack power a tick on rank 2 (combat logs at levels 13 and 14);
// the higher ranks are assumed to keep the share.
const ForeverRendAPPerTick = 0.02

func (warrior *Warrior) registerRendSpell() {
	spellIDs := []int32{0, 772, 6546, 6547, 6548, 11572, 11573, 11574}
	rank := warrior.knownRank([]int32{0, 4, 10, 20, 30, 40, 50, 60}, spellIDs)
	if rank == 0 {
		return
	}
	// Client 71091: a tick every 3 sec, for 9 to 21 sec.
	baseDamage := []float64{0, 5, 7, 9, 11, 14, 18, 21}[rank]
	ticks := []int32{0, 3, 4, 5, 6, 7, 7, 7}[rank]
	rendAPPerTick := core.Assume("warr_rend_ap", ForeverRendAPPerTick) // DPS Report assumption knob

	damageMultiplier := []float64{1, 1.12, 1.23, 1.35}[warrior.Talents.ImprovedRend] // client 70338: 12, 23 and 35%

	warrior.Rend = warrior.RegisterSpell(BattleStance|DefensiveStance, core.SpellConfig{
		SpellCode:   SpellCode_WarriorRend,
		ActionID:    core.ActionID{SpellID: spellIDs[rank]},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagAPL | core.SpellFlagNoOnCastComplete | SpellFlagOffensive,

		RageCost: core.RageCostOptions{
			Cost:   10,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		// Under Forever Rend's ticks crit (its rows carry Periodic Can Crit, Attributes_8 0x200;
		// wowsims/forever 872e6def91), and Impale's class mask covers Rend.
		CritDamageBonus: core.TernaryFloat64(warrior.Env.IsForever(), warrior.impale(), 0),

		DamageMultiplier: damageMultiplier,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Rend",
				Tag:   "Rend",
			},
			NumberOfTicks: ticks,
			TickLength:    time.Second * 3,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				// Under Forever each tick reads attack power and the damage modifiers when it lands:
				// cancelling Battle Shout mid-Rend lowers the next ticks (wowsims/forever 1172f1c745
				// and 36cfc58328, level 13 and 14 combat logs).
				if sim.IsForever() {
					damage := baseDamage + rendAPPerTick*dot.Spell.MeleeAttackPower(target)
					dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, dot.OutcomeTick)
					return
				}
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			} else {
				spell.IssueRefund(sim)
			}

			spell.DealOutcome(sim, result)
		},
	})

}
