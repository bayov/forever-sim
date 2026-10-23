package hunter

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Summon Hawk, the 16 point Beast Mastery talent (beta client 1293241 and 1293525 to 1293527, build
// 1.60.1.70338). The dive bomb deals 32/47/85/108 by rank plus 5% of ranged attack power, costs
// 80/105/135/190 mana and shares Arcane Shot's 6 sec cooldown. It leaves a hawk that attacks the
// target for 18 sec (1293248), at most 2 at once (the spell's third effect), so a third replaces the
// one with the least time left. The dive bomb carries the client's always-hit attribute. Unleashed
// Fury and Ferocity name Summon Hawk in their class masks, so the dive bomb deals 3% more a rank of
// the first and crits 2% more a rank of the second.
const hawkDuration = time.Second * 18

// The hawk is a guardian the client does not describe; the row that scales it, Hunter Guardian Hawk
// Pet Scaling 01 (1293586), is not in the client. wowsims/forever fitted it to level 30 combat logs
// (81651da664: 2,862 swings at 407 to 1,023 owner ranged attack power; b89878e6f9: 4,340 swings with
// Ferocity 5/5), all with Unleashed Fury 5/5. A swing before armor is 3.3 to 6.7 plus 0.0179 a point
// of the owner's ranged attack power, the same line in every log: the hawk has 80% of the owner's
// ranged attack power and none of its own at 112 attack power a point of dps over its 2.5 sec swing,
// and the fixed part is 0.0815 of the creature auto attack dps for its level over the 2.5 sec, a
// third either side. Both are stated here before the logs' Unleashed Fury. The sim takes the level
// 60 creature dps at every level, as the tamed pet does (pet.go): about 12.3 a swing after the
// talent at 60, where the level 30 logs read 5.0.
const hawkLogUnleashedFury = 1.15

var hawkSwingAverage = 0.0815 * creatureAutoAttackDps60 * 2.5 / hawkLogUnleashedFury

// A hawk swings on arrival and every 2.5 sec after at its owner's ranged speed, the quiver, Quick
// Shots and Rapid Fire included, following each change while it is up; attack speed (Berserking)
// does not reach it (0fc13fcb30, e759dc1802: 2.277 sec with a 10% quiver, 1.764 sec with Quick Shots
// too, 2.27 and 2.22 sec under Berserking alone). Its crit is half the hunter's (the tamed pet's,
// pet.go) with none from Agility, plus Ferocity whole: 4,340 swings with Ferocity 5/5 read 10 plus
// half the owner's crit against targets 2 levels below to 3 above, and 0-Ferocity logs half the
// tamed pet's (b89878e6f9). Its melee hit and expertise are the hunter's, as the tamed pet's are: 1%
// hit took a hawk from 1 miss in 32 swings to none in 515 against a target 8 levels below it
// (81651da664). 30% of the hunter's armor, as the tamed pet; its health (5% of the creature's) is
// left out, as nothing hits it. The hunter's ranged attack power, crit and hit reach it live; Hunter's
// Mark, a bonus to ranged attackers of the target, does not, as the hawk attacks in melee.
func hawkStatInheritance(ownerStats stats.Stats) stats.Stats {
	return stats.Stats{
		stats.AttackPower: 0.8 * ownerStats[stats.RangedAttackPower],
		stats.MeleeCrit:   0.5 * ownerStats[stats.MeleeCrit],
		stats.MeleeHit:    ownerStats[stats.MeleeHit],
		stats.Expertise:   ownerStats[stats.Expertise],
		stats.Armor:       0.3 * ownerStats[stats.Armor],
	}
}

type HawkPet struct {
	core.Pet
}

func (hawk *HawkPet) GetPet() *core.Pet             { return &hawk.Pet }
func (hawk *HawkPet) Initialize()                   {}
func (hawk *HawkPet) Reset(_ *core.Simulation)      {}
func (hawk *HawkPet) OnGCDReady(_ *core.Simulation) {}
func (hawk *HawkPet) ExecuteCustomRotation(sim *core.Simulation) {
	// A hawk only swings.
	hawk.WaitUntil(sim, sim.CurrentTime+hawkDuration)
}

func summonHawkRank(level int32) int {
	for i := 4; i >= 1; i-- {
		if []int32{0, 25, 36, 48, 60}[i] <= level {
			return i
		}
	}
	return 0
}

// The hawks, guardians with no player buffs. A hawk deals 3% more a rank of Unleashed Fury, and takes
// the damage the hunter deals as it is summoned (Focused Fire with the pet out, Lone Wolf without
// it, the world buffs) on top, as wowsims/forever's hawk does (81651da664). The hunter's damage done
// against a creature type (Improved Tracking) is not in that and doesn't reach it.
func (hunter *Hunter) newHawks() []*HawkPet {
	if !hunter.Talents.SummonHawk || summonHawkRank(hunter.Level) == 0 {
		return nil
	}

	hawks := make([]*HawkPet, 2)
	for i := range hawks {
		hawk := &HawkPet{
			Pet: core.NewPet("Hawk "+strconv.Itoa(i+1), &hunter.Character, stats.Stats{}, hawkStatInheritance, false, true),
		}
		hawk.FollowsOwnerRangedSpeed = true
		hawk.EnableAutoAttacks(hawk, core.AutoAttackOptions{
			MainHand: core.Weapon{
				BaseDamageMin:     hawkSwingAverage * (1 - 1.0/3),
				BaseDamageMax:     hawkSwingAverage * (1 + 1.0/3),
				SwingSpeed:        2.5,
				AttackPowerPerDPS: 112 * hawkLogUnleashedFury,
			},
			AutoSwingMelee: true,
		})

		hawk.PseudoStats.DamageDealtMultiplier *= 1 + 0.03*float64(hunter.Talents.UnleashedFury)
		hawk.AddStat(stats.MeleeCrit, 2*float64(hunter.Talents.Ferocity)*core.CritRatingPerCritChance)

		inherited := 1.0
		hawk.OnPetEnable = func(sim *core.Simulation) {
			inherited = hunter.PseudoStats.DamageDealtMultiplier
			hawk.PseudoStats.DamageDealtMultiplier *= inherited
			hawk.EnableDynamicStats(hawkStatInheritance)
			hawk.AutoAttacks.ResetMeleeSwingAt()
		}
		hawk.OnPetDisable = func(sim *core.Simulation) {
			hawk.PseudoStats.DamageDealtMultiplier /= inherited
		}

		hunter.AddPet(hawk)
		hawks[i] = hawk
	}
	return hawks
}

func (hunter *Hunter) registerSummonHawkSpell(timer *core.Timer) {
	if len(hunter.hawks) == 0 {
		return
	}

	rank := summonHawkRank(hunter.Level)
	spellID := [5]int32{0, 1293241, 1293525, 1293526, 1293527}[rank]
	baseDamage := [5]float64{0, 32, 47, 85, 108}[rank]
	manaCost := [5]float64{0, 80, 105, 135, 190}[rank]
	unleashedFury := 1 + 0.03*float64(hunter.Talents.UnleashedFury)

	// Each hawk's time out is a dot on the hunter, which the rotations read (Summon Hawk while
	// fewer than 2 hawks are out).
	lifetimes := make([]*core.Spell, len(hunter.hawks))
	for i, hawk := range hunter.hawks {
		lifetimes[i] = hunter.RegisterSpell(core.SpellConfig{
			SpellCode:   SpellCode_HunterSummonHawk,
			ActionID:    core.ActionID{SpellID: spellID, Tag: int32(i + 1)},
			SpellSchool: core.SpellSchoolPhysical,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagNoMetrics | core.SpellFlagNoOnCastComplete,

			Dot: core.DotConfig{
				IsAOE: true,
				Aura: core.Aura{
					Label: "Summon Hawk " + strconv.Itoa(i+1),
					OnGain: func(aura *core.Aura, sim *core.Simulation) {
						hawk.Enable(sim, hawk)
					},
					OnExpire: func(aura *core.Aura, sim *core.Simulation) {
						hawk.Disable(sim)
					},
				},
				NumberOfTicks: 1,
				TickLength:    hawkDuration,
				OnTick:        func(_ *core.Simulation, _ *core.Unit, _ *core.Dot) {},
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
			slot := 0
			for i, l := range lifetimes[1:] {
				cur, dot := lifetimes[slot].AOEDot(), l.AOEDot()
				if cur.IsActive() && (!dot.IsActive() || dot.RemainingDuration(sim) < cur.RemainingDuration(sim)) {
					slot = i + 1
				}
			}

			hunter.hawks[slot].CurrentTarget = target
			lifetimes[slot].AOEDot().Apply(sim)
		},
	})
}
