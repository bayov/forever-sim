package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core/proto"

	"github.com/wowsims/classic/sim/core"
)

func (paladin *Paladin) registerHammerOfWrath() {
	// Forever's damage at level 60, about 5% below Classic. Ranks 1 and 2 are wowhead's
	// Forever tooltip, which shows them at their cap (level 49 and 57). Rank 3 keeps growing
	// 3.1 a level up to 65 and wowhead shows it there (489 to 538), so at 60 it is 15.5
	// lower. ForeverChanges' spellbook agrees (474 to 522).
	ranks := []struct {
		level     int32
		spellID   int32
		minDamage float64
		maxDamage float64
		manaCost  float64
	}{
		{level: 44, spellID: 24275, manaCost: 295, minDamage: 286, maxDamage: 314},
		{level: 52, spellID: 24274, manaCost: 360, minDamage: 382, maxDamage: 421},
		{level: 60, spellID: 24239, manaCost: 425, minDamage: 473.5, maxDamage: 522.5},
	}

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * 6,
	}

	for i, rank := range ranks {
		rank := rank
		if paladin.Level < rank.level {
			break
		}

		paladin.GetOrRegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeRanged,
			ProcMask:    core.ProcMaskRangedSpecial, // TODO to be tested
			Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
			CastType:    proto.CastType_CastTypeRanged,

			Rank:          i + 1,
			RequiredLevel: int(rank.level),
			SpellCode:     SpellCode_PaladinHammerOfWrath,

			ManaCost: core.ManaCostOptions{
				FlatCost: rank.manaCost,
				// Instrument of Law 2 makes it instant, and then Benediction counts too.
				Multiplier: core.TernaryInt32(paladin.Talents.InstrumentOfLaw == 2, paladin.holyConduitInstant(), paladin.holyConduit()),
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: time.Second,
					// Instrument of Law takes half a second per point off the cast.
					CastTime: time.Second - time.Millisecond*500*time.Duration(paladin.Talents.InstrumentOfLaw),
				},
				IgnoreHaste: true,
				CD:          cd,
			},

			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			BonusCoefficient: 0.429,
			BonusHitRating:   -float64(paladin.Talents.Precision) * core.MeleeHitRatingPerHitChance,

			ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
				return sim.IsExecutePhase20()
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				damage := sim.Roll(rank.minDamage, rank.maxDamage)
				spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeRangedHitAndCrit)
			},
		})
	}
}
