package druid

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Forever's Hurricane (1.60.1.70338, wago.tools): every tick is its own spell (1278965,
// 1278968, 1278759) with 68 + 0.2, 98 + 0.2 and 132 + 0.3 a level (1.12 has 70, 100 and 134),
// the same 3% coefficient a tick, and no cooldown (1.12 has 1 min).
func (druid *Druid) registerHurricaneSpell() {
	ranks := []struct {
		level      int32
		spellID    int32
		manaCost   float64
		scaleLevel int32
		damage     float64
		scale      float64
	}{
		{level: 40, spellID: 16914, manaCost: 880, scaleLevel: 46, damage: 68, scale: 0.2},
		{level: 50, spellID: 17401, manaCost: 1180, scaleLevel: 56, damage: 98, scale: 0.2},
		{level: 60, spellID: 17402, manaCost: 1495, scaleLevel: 66, damage: 132, scale: 0.3},
	}

	// Under Forever each tick is a hit of its own that can miss but not crit: the tick rows
	// carry no SPELL_ATTR8_PERIODIC_CAN_CRIT (0x200; 17402 and 1278759 have Attr8 0x1000), and
	// wowsims/forever casts every tick with OutcomeMagicHit.
	flags := SpellFlagOmen | core.SpellFlagChanneled | core.SpellFlagBinary | core.SpellFlagAPL
	tickRollsHit := druid.Env.IsForever()
	if tickRollsHit {
		flags |= core.SpellFlagNoPeriodicCrit
	}

	for i, rank := range ranks {
		if druid.Level < rank.level {
			break
		}

		damage := rank.damage + float64(min(druid.Level, rank.scaleLevel)-rank.level)*rank.scale
		spell := druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellSchool: core.SpellSchoolNature,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellDamage,
			Flags:       flags,

			PushbackReduction: druid.naturesFocusChance(),

			RequiredLevel: int(rank.level),
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{
				FlatCost: rank.manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
			},

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			Dot: core.DotConfig{
				IsAOE: true,
				Aura: core.Aura{
					Label: "Hurricane" + druid.Label + strconv.Itoa(i+1),
				},
				NumberOfTicks: 10,
				TickLength:    time.Second * 1,

				BonusCoefficient: 0.03,

				OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
					dot.Snapshot(target, damage, isRollover)
				},
				OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
					outcome := core.Ternary(tickRollsHit, dot.OutcomeMagicHitAndTick, dot.OutcomeTick)
					for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
						dot.CalcAndDealPeriodicSnapshotDamage(sim, aoeTarget, outcome)
					}
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				druid.AutoAttacks.CancelAutoSwing(sim)
				spell.AOEDot().Apply(sim)
			},
		})

		druid.Hurricane = append(druid.Hurricane, spell)
	}
}
