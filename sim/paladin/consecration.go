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
// server script; its damage spells (1280345 to 1280349) carry the two parts as two effects,
// and only the extra one has a spell power coefficient: 9.5% a tick in client 1.60.1.70338
// and 4.2% (Era's) in 1.60.1.70425, which the sim uses. The rank's periodic dummy names 4
// targets. So the first four enemies take the extra damage and its spell power, and the
// rest take the flat base tick alone (wowsims/forever consecration.go).
//
// The ground marks those four enemies (client 1280353), and Consecrated Ground's Holy
// damage bonus goes on the mark, see applyConsecratedGround.
const consecrationBonusTargets = 4

// The mark outlives the ground's 8 sec by its own 0.6 sec (client 1280353), so it is still up
// on the last tick.
const consecrationMarkDuration = time.Second*8 + time.Millisecond*600

func (paladin *Paladin) registerConsecrationMarks() {
	paladin.consecrationMarks = paladin.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			Label:    "Consecration Mark" + paladin.Label,
			ActionID: core.ActionID{SpellID: 1280353},
			Duration: consecrationMarkDuration,
		})
	})
}

// markConsecrationTargets lays the mark on the first four enemies.
func (paladin *Paladin) markConsecrationTargets(sim *core.Simulation) {
	for i, target := range sim.Encounter.ActiveTargetUnits {
		if i >= consecrationBonusTargets {
			break
		}
		paladin.consecrationMarks.Get(target).Activate(sim)
	}
}

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

	paladin.registerConsecrationMarks()

	for i, rank := range ranks {
		rank := rank
		if paladin.Level < rank.level {
			break
		}
		tickDamage := rank.damage / 8
		tickCoefficient := 0.042
		// Forever's base and extra parts of a tick, snapshotted at the cast.
		var baseTick, extraTick, baseSnapshot, extraSnapshot float64
		if paladin.Env.IsForever() {
			baseTick = rank.foreverAll / 8
			extraTick = rank.foreverExtra / 8
			tickCoefficient = 0.042 // client 70425 (spells 1280345-1280349; 0.095 in 70338)
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
			BonusHitRating:   paladin.divinePrecisionHitRating(core.SpellHitRatingPerHitChance),
			// The ticks crit under Forever (core's periodic crit; wowsims/forever 2f953756bd and
			// fd449cd4eb, a Forever group log with 4,871 ticks), and Holy Power's list (client 5923)
			// takes in Consecration and its damage spells.
			BonusCritRating:  paladin.holyPowerCrit(),
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
					if !sim.IsForever() {
						dot.Snapshot(target, tickDamage, isRollover)
						return
					}
					dot.Snapshot(target, extraTick, isRollover)
					extraSnapshot = dot.SnapshotBaseDamage
					baseSnapshot = baseTick * dot.Spell.BaseDamageMultiplierAdditive
				},
				OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
					// Consecration can miss, showing up as either a resist in logs or a
					// silent failure (missing damage tick).
					for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
						if !sim.IsForever() {
							dot.CalcAndDealPeriodicSnapshotDamage(sim, aoeTarget, dot.OutcomeMagicHitAndTick)
							continue
						}
						if paladin.consecrationMarks.Get(aoeTarget).IsActive() {
							dot.SnapshotBaseDamage = baseSnapshot + extraSnapshot
							dot.CalcAndDealPeriodicSnapshotDamage(sim, aoeTarget, dot.OutcomeMagicHitAndTick)
							continue
						}
						// The base tick has no spell power coefficient, so the Holy damage the
						// target takes extra doesn't reach it either.
						dot.SnapshotBaseDamage = baseSnapshot
						dot.BonusCoefficient = 0
						dot.CalcAndDealPeriodicSnapshotDamage(sim, aoeTarget, dot.OutcomeMagicHitAndTick)
						dot.BonusCoefficient = tickCoefficient
					}
				},
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				paladin.markConsecrationTargets(sim)
				spell.AOEDot().Apply(sim)
			},
		})
	}
}
