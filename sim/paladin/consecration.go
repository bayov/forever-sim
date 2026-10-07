package paladin

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Consecration is baseline under Forever (Classic's 11 point Holy talent) and reworked:
// everyone in the area takes a quarter of Classic's damage over 8 sec, and the first four
// enemies to enter take an extra amount on top. wowhead's Forever tooltips give both parts
// for every rank, e.g. rank 1 does 16 plus 32 (Classic 64). The damage runs through a
// server script, so no page shows its spell power coefficient. We keep the 9.5% per tick
// an older Forever tooltip showed. Every sim target counts as one of the first four.
func (paladin *Paladin) registerConsecration() {
	ranks := []struct {
		level    int32
		spellID  int32
		manaCost float64
		damage   float64
		// Forever's damage to everyone in the area, and the extra to the first four.
		foreverAll   float64
		foreverExtra float64
	}{
		{level: 20, spellID: 26573, manaCost: 135, damage: 64, foreverAll: 16, foreverExtra: 32},
		{level: 30, spellID: 20116, manaCost: 235, damage: 120, foreverAll: 24, foreverExtra: 56},
		{level: 40, spellID: 20922, manaCost: 320, damage: 192, foreverAll: 48, foreverExtra: 88},
		{level: 50, spellID: 20923, manaCost: 435, damage: 280, foreverAll: 64, foreverExtra: 160},
		{level: 60, spellID: 20924, manaCost: 565, damage: 384, foreverAll: 96, foreverExtra: 216},
	}

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * 8,
	}

	for i, rank := range ranks {
		rank := rank
		if paladin.Level < rank.level {
			break
		}
		tickDamage := rank.damage / 8
		tickCoefficient := 0.042
		if paladin.Env.IsForever() {
			tickDamage = (rank.foreverAll + rank.foreverExtra) / 8
			tickCoefficient = 0.095
		}

		paladin.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellDamage,
			Flags:       core.SpellFlagPureDot | core.SpellFlagAPL,

			RequiredLevel: int(rank.level),
			Rank:          i + 1,

			SpellCode: SpellCode_PaladinConsecration,
			ManaCost: core.ManaCostOptions{
				FlatCost:   rank.manaCost,
				Multiplier: paladin.holyConduitInstant(),
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
				CD: cd,
			},
			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			BonusCoefficient: tickCoefficient,
			Dot: core.DotConfig{
				IsAOE: true,
				Aura: core.Aura{
					Label: "Consecration" + paladin.Label + strconv.Itoa(i+1),
				},
				NumberOfTicks: 8,
				TickLength:    time.Second * 1,

				BonusCoefficient: tickCoefficient,

				OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
					dot.Snapshot(target, tickDamage, isRollover)
				},
				OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
					// Consecration can miss, showing up as either a resist in logs or a
					// silent failure (missing damage tick).
					for _, aoeTarget := range sim.Encounter.TargetUnits {
						dot.CalcAndDealPeriodicSnapshotDamage(sim, aoeTarget, dot.OutcomeMagicHitAndTick)
					}
				},
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.AOEDot().Apply(sim)
				if paladin.consecratedGroundAura != nil {
					paladin.consecratedGroundAura.Activate(sim)
				}
			},
		})
	}
}
