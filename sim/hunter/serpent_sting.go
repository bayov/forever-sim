package hunter

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// Forever's Serpent Sting (beta client, build 1.60.1.70338) keeps rank 9's 555 over 15 sec and
// lowers the ranks below it (rank 8: 415, Classic 490). It has no spell power coefficient (Classic
// 100% over the duration). The client gives it no attack power share either, but the beta's combat
// logs put each tick at the base plus a share of ranged attack power, read live on every tick:
// about 0.035 in ElliotWood/Forever's fit (sim/hunter/serpent_sting.go), 3% a tick on the
// forever-hunter wiki. Report assumption hunter_shot_ap_shares, user ruling 2026-10-15: the wiki's
// 3%, as for Arcane Shot. The ticks crit with the hunter's ranged crit under Forever, and Mortal
// Shots is in its class mask.
const serpentStingRAPPerTick = 0.03

func (hunter *Hunter) getSerpentStingConfig(rank int) core.SpellConfig {
	spellId := [10]int32{0, 1978, 13549, 13550, 13551, 13552, 13553, 13554, 13555, 25295}[rank]
	baseDamage := [10]float64{0, 10, 30, 60, 110, 170, 240, 320, 415, 555}[rank] / 5
	manaCost := [10]float64{0, 15, 30, 50, 80, 115, 150, 190, 230, 250}[rank]
	level := [10]int{0, 4, 10, 18, 26, 34, 42, 50, 58, 60}[rank]

	// Improved Stings: 6/13/20% (client curve) on the periodic damage.
	impStings := []float64{1, 1.06, 1.13, 1.20}[hunter.Talents.ImprovedStings]

	return core.SpellConfig{
		SpellCode:     SpellCode_HunterSerpentSting,
		ActionID:      core.ActionID{SpellID: spellId},
		SpellSchool:   core.SpellSchoolNature,
		DefenseType:   core.DefenseTypeRanged,
		ProcMask:      core.ProcMaskRangedSpecial,
		Flags:         core.SpellFlagAPL | core.SpellFlagPureDot | core.SpellFlagPoison | SpellFlagSting,
		CastType:      proto.CastType_CastTypeRanged,
		Rank:          rank,
		RequiredLevel: level,
		MissileSpeed:  40,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true, // Hunter GCD is locked at 1.5s
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.inRangedRange()
		},

		CritDamageBonus:  hunter.mortalShots(),
		DamageMultiplier: impStings,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "SerpentSting" + hunter.Label + strconv.Itoa(rank),
				Tag:   "SerpentSting",
			},
			NumberOfTicks: 5,
			TickLength:    time.Second * 3,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.SnapshotBaseDamage = (baseDamage + serpentStingRAPPerTick*dot.Spell.RangedAttackPower(target, false)) * dot.Spell.BaseDamageMultiplierAdditive
				if sim.IsForever() {
					dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTickPhysicalCrit)
				} else {
					dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
				}
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeRangedHitNoHitCounter)

			spell.WaitTravelTime(sim, func(s *core.Simulation) {
				spell.DealOutcome(sim, result)

				if result.Landed() {
					spell.Dot(target).Apply(sim)
				}
			})
		},
	}
}

// Rank 9 comes from an Ahn'Qiraj book, out under Forever (rank 8's 415 instead of 555).
func (hunter *Hunter) registerSerpentStingSpell() {
	maxRank := 9
	if !hunter.Env.IncludeAQRanks() {
		maxRank = 8
	}
	for rank := maxRank; rank >= 1; rank-- {
		config := hunter.getSerpentStingConfig(rank)

		if config.RequiredLevel <= int(hunter.Level) {
			hunter.SerpentSting = hunter.GetOrRegisterSpell(config)
			return
		}
	}
}
