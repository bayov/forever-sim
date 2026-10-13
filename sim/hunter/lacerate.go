package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Lacerate is a new Forever trainer spell from level 30 (beta client 24118, 24119, 24120 and
// 1299332, build 1.60.1.70338): a bleed for 21 sec, a tick every 3 sec, 95 mana at rank 4. Each tick
// is 58 at rank 4's level 60 plus 0.8 a level above it, so 406 over the duration at 60 (wowhead's
// 451 is at level 68). The client gives it no attack power share, and no beta log has shown one,
// so it deals the flat amount. It rolls the melee table to land, and its ticks crit under Forever
// like every other dot.
func (hunter *Hunter) registerLacerateSpell() {
	rank := 0
	for i := 4; i >= 1; i-- {
		if []int{0, 30, 40, 50, 60}[i] <= int(hunter.Level) {
			rank = i
			break
		}
	}
	if rank == 0 {
		return
	}

	spellID := [5]int32{0, 24118, 24119, 24120, 1299332}[rank]
	baseLevel := [5]int32{0, 30, 40, 50, 60}[rank]
	maxLevel := [5]int32{0, 38, 48, 58, 68}[rank]
	tickDamage := [5]float64{0, 17, 28, 40, 58}[rank] +
		[5]float64{0, 0.5, 0.6, 0.7, 0.8}[rank]*float64(min(hunter.Level, maxLevel)-baseLevel)
	manaCost := [5]float64{0, 40, 65, 80, 95}[rank]

	hunter.Lacerate = hunter.RegisterSpell(core.SpellConfig{
		SpellCode:     SpellCode_HunterLacerate,
		ActionID:      core.ActionID{SpellID: spellID},
		SpellSchool:   core.SpellSchoolPhysical,
		DefenseType:   core.DefenseTypeMelee,
		ProcMask:      core.ProcMaskMeleeMHSpecial,
		Flags:         core.SpellFlagMeleeMetrics | core.SpellFlagAPL | core.SpellFlagPureDot | SpellFlagStrike,
		Rank:          rank,
		RequiredLevel: 30,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.inMeleeRange()
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Lacerate" + hunter.Label,
			},
			NumberOfTicks: 7,
			TickLength:    time.Second * 3,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, tickDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
			spell.DealOutcome(sim, result)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
		},
	})
}
