package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// Forever's Arcane Shot (beta client, build 1.60.1.70338) hits harder at most ranks (rank 8: 217,
// Classic 183) and has no spell power coefficient (Classic 42.9%). The launch realms' server
// hotfix of 2026-11-03 (foreverchanges.pro, read from the realms' hotfix cache) raises the ranks to
// 25 / 35 / 50 / 85 / 120 / 175 / 220 / 280, Blizzard's 11-02 notes' "~30%"; client 1.60.1.70753
// still has 20 to 217. The client gives it no attack power share, but the beta's combat logs show
// one (about 0.11 in ElliotWood/Forever's fit, 10% on the forever-hunter wiki, the sim's 10% by
// report assumption hunter_shot_ap_shares, user ruling 2026-10-15). Blizzard's 11-02 notes and
// foreverchanges.pro's reading of them give it as "raised from 0.1 to 0.15", so it is 15% now.
// The 6 sec cooldown is shared with Summon Hawk (client category 1173).
const arcaneShotRAPCoefficient = 0.15

var ArcaneShotBaseDamage = [9]float64{0, 25, 35, 50, 85, 120, 175, 220, 280}

func (hunter *Hunter) getArcaneShotConfig(rank int, timer *core.Timer) core.SpellConfig {
	spellId := [9]int32{0, 3044, 14281, 14282, 14283, 14284, 14285, 14286, 14287}[rank]
	baseDamage := ArcaneShotBaseDamage[rank]
	manaCost := [9]float64{0, 25, 35, 50, 80, 105, 135, 160, 190}[rank]
	level := [9]int{0, 6, 12, 20, 28, 36, 44, 52, 60}[rank]

	return core.SpellConfig{
		SpellCode:     SpellCode_HunterArcaneShot,
		ActionID:      core.ActionID{SpellID: spellId},
		SpellSchool:   core.SpellSchoolArcane,
		DefenseType:   core.DefenseTypeRanged,
		ProcMask:      core.ProcMaskRangedSpecial,
		Flags:         core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagShot,
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
			IgnoreHaste: true,
			// Improved Arcane Shot: 0.3 sec a rank (Classic 0.2), on Arcane Shot alone and not on
			// Summon Hawk, which shares the cooldown.
			CD: core.Cooldown{
				Timer:    timer,
				Duration: time.Second*6 - time.Millisecond*300*time.Duration(hunter.Talents.ImprovedArcaneShot),
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.inRangedRange()
		},
		// We only get in or out of range with an event, like the PvP downtime.
		ExtraCastConditionNeedsEvent: true,

		CritDamageBonus: hunter.mortalShots(),

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := baseDamage + arcaneShotRAPCoefficient*spell.RangedAttackPower(target, false)
			result := spell.CalcDamage(sim, target, damage, spell.OutcomeRangedHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	}
}

func (hunter *Hunter) registerArcaneShotSpell(timer *core.Timer) {
	for i := 8; i >= 1; i-- {
		config := hunter.getArcaneShotConfig(i, timer)

		if config.RequiredLevel <= int(hunter.Level) {
			hunter.ArcaneShot = hunter.GetOrRegisterSpell(config)
			return
		}
	}
}
