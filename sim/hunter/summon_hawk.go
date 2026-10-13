package hunter

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// Summon Hawk, the 16 point Beast Mastery talent (beta client 1293241 and 1293525 to 1293527, build
// 1.60.1.70338). The dive bomb deals 32/47/85/108 by rank plus 5% of ranged attack power, costs
// 80/105/135/190 mana and shares Arcane Shot's 6 sec cooldown. It leaves a hawk that attacks the
// target for 18 sec (1293248), at most 2 at once (the spell's third effect), so a third replaces the
// one with the least time left. The dive bomb carries the client's always-hit attribute.
//
// The hawk is a guardian the client does not describe. Its swings come from the beta logs
// (ElliotWood/Forever, sim/hunter/summon_hawk.go: 534 landed swings of rank 1 at levels 25 to 30):
// one on arrival, then one every 2.5 sec sped up by the hunter's ranged haste, each about 0.35 of the
// rank's dive bomb base and hardly ever a crit. Unleashed Fury and Ferocity name Summon Hawk in their
// class masks, so the hawks deal 3% more a rank of the first and the dive bomb crits 2% more a rank
// of the second.
const hawkSwingShare = 0.35
const hawkSwingInterval = time.Millisecond * 2500
const hawkDuration = time.Second * 18

func (hunter *Hunter) registerSummonHawkSpell(timer *core.Timer) {
	if !hunter.Talents.SummonHawk {
		return
	}

	rank := 0
	for i := 4; i >= 1; i-- {
		if []int{0, 25, 36, 48, 60}[i] <= int(hunter.Level) {
			rank = i
			break
		}
	}
	if rank == 0 {
		return
	}

	spellID := [5]int32{0, 1293241, 1293525, 1293526, 1293527}[rank]
	baseDamage := [5]float64{0, 32, 47, 85, 108}[rank]
	manaCost := [5]float64{0, 80, 105, 135, 190}[rank]
	swingDamage := baseDamage * hawkSwingShare
	unleashedFury := 1 + 0.03*float64(hunter.Talents.UnleashedFury)

	hawks := make([]*core.Spell, 2)
	for i := range hawks {
		hawks[i] = hunter.RegisterSpell(core.SpellConfig{
			SpellCode:   SpellCode_HunterSummonHawk,
			ActionID:    core.ActionID{SpellID: spellID, Tag: int32(i + 1)},
			SpellSchool: core.SpellSchoolPhysical,
			DefenseType: core.DefenseTypeMelee,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete,

			DamageMultiplier: unleashedFury,
			ThreatMultiplier: 1,

			Dot: core.DotConfig{
				Aura: core.Aura{
					Label: "Summon Hawk " + strconv.Itoa(i+1) + hunter.Label,
				},
				NumberOfTicks: int32(hawkDuration / hawkSwingInterval),
				TickLength:    hawkSwingInterval,

				OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
					dot.Snapshot(target, swingDamage, isRollover)
				},
				OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
					result := dot.Spell.CalcDamage(sim, target, dot.SnapshotBaseDamage, dot.Spell.OutcomeMeleeSpecialHit)
					dot.Spell.DealDamage(sim, result)
				},
			},
		})
	}

	hunter.SummonHawk = hunter.RegisterSpell(core.SpellConfig{
		SpellCode:     SpellCode_HunterSummonHawk,
		ActionID:      core.ActionID{SpellID: spellID},
		SpellSchool:   core.SpellSchoolPhysical,
		DefenseType:   core.DefenseTypeRanged,
		ProcMask:      core.ProcMaskRangedSpecial,
		Flags:         core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		CastType:      proto.CastType_CastTypeRanged,
		Rank:          rank,
		RequiredLevel: 25,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    timer,
				Duration: time.Second * 6,
			},
		},

		BonusCritRating:  2 * float64(hunter.Talents.Ferocity) * core.CritRatingPerCritChance,
		DamageMultiplier: unleashedFury,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := baseDamage + 0.05*spell.RangedAttackPower(target, false)
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeRangedCritOnly)

			// A free slot, or else the hawk with the least time left.
			hawk := hawks[0].Dot(target)
			for _, h := range hawks[1:] {
				if dot := h.Dot(target); hawk.IsActive() && (!dot.IsActive() || dot.RemainingDuration(sim) < hawk.RemainingDuration(sim)) {
					hawk = dot
				}
			}

			tick := time.Duration(float64(hawkSwingInterval) / hunter.RangedSwingSpeed())
			hawk.TickLength = tick
			hawk.NumberOfTicks = int32(hawkDuration / tick)
			hawk.Apply(sim)
			hawk.TickOnce(sim)
		},
	})
}
