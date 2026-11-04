package shaman

import (
	"fmt"
	"math"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// The shaman's heals. A heal cast at a player lands on that player (a healer's rotation aims
// them with Most Hurt, Lowest Health or a tank, docs/healing_sim.md). One cast at a mob, the
// rotation's default target, lands on the shaman, as the leveling sim's self heals between and
// during pulls do (docs/leveling_sim.md).
//
// Chain Heal heals 3 targets in all, each jump 50% as much as the one before (the tooltip,
// Wowhead's Classic and Forever pages agree), and jumps to the most hurt player each time. Cast
// at a party member, it jumps only within that party. Everyone is in reach for now. A
// shaman soloing mobs has no one to jump to.
//
// Every value is the Forever client's (1.60.1.70338, SpellEffect, SpellLevels, SpellMisc and
// SpellPower on wago.tools). A rank heals its base at the level it's learned, plus its growth
// per level for 5 levels after, and the variance spreads it evenly around that:
// base * (1 - variance / 2) to base * (1 + variance / 2). Lightning Bolt's ranks read the same
// way give the sim's Lightning Bolt table to the point. The spell power coefficients follow the
// cast time (1.5 / 3.5 and so on), and Forever's downranking takes them down for a rank far
// below our level (sim/core/downrank.go).
type healRank struct {
	spellID              int32
	base, perLevel       float64
	variance             float64
	learnLevel, maxLevel int32
	coefficient          float64
	castTime             time.Duration
	manaCost             float64
}

// At is the rank's lowest and highest heal at the character's level.
func (r healRank) At(level int32) (float64, float64) {
	value := r.base + r.perLevel*float64(min(level, r.maxLevel)-r.learnLevel)
	return value * (1 - r.variance/2), value * (1 + r.variance/2)
}

var HealingWaveRanks = []healRank{
	{},
	{331, 39, 0.7, 0.25641027093, 1, 5, 0.429, 1500 * time.Millisecond, 25},
	{332, 71, 1.0, 0.19718310237, 6, 11, 0.571, 2000 * time.Millisecond, 45},
	{547, 131, 1.5, 0.18309858441, 12, 17, 0.714, 2500 * time.Millisecond, 80},
	{913, 256, 2.3, 0.16438356042, 18, 23, 0.857, 3000 * time.Millisecond, 155},
	{939, 342, 2.7, 0.15686275065, 24, 29, 0.857, 3000 * time.Millisecond, 200},
	{959, 484, 3.3, 0.14853195846, 32, 37, 0.857, 3000 * time.Millisecond, 265},
	{8005, 685, 3.9, 0.14303638041, 40, 45, 0.857, 3000 * time.Millisecond, 340},
	{10395, 972, 4.7, 0.13736264408, 48, 53, 0.857, 3000 * time.Millisecond, 440},
	{10396, 1344, 5.5, 0.13251365721, 56, 61, 0.857, 3000 * time.Millisecond, 560},
	{25357, 1615, 5.5, 0.1325648427, 60, 65, 0.857, 3000 * time.Millisecond, 620},
}

var LesserHealingWaveRanks = []healRank{
	{},
	{8004, 153, 1.7, 0.13793103397, 20, 25, 0.429, 1500 * time.Millisecond, 105},
	{8008, 224, 2.1, 0.12878787518, 28, 33, 0.429, 1500 * time.Millisecond, 145},
	{8010, 304, 2.5, 0.12256267667, 36, 41, 0.429, 1500 * time.Millisecond, 185},
	{10466, 426, 3.0, 0.11522634327, 44, 49, 0.429, 1500 * time.Millisecond, 235},
	{10467, 608, 3.6, 0.11077844352, 52, 57, 0.429, 1500 * time.Millisecond, 305},
	{10468, 820, 4.2, 0.10909090936, 60, 65, 0.429, 1500 * time.Millisecond, 380},
}

var ChainHealRanks = []healRank{
	{},
	{1064, 251, 2.5, 0.13953489065, 40, 45, 0.714, 2500 * time.Millisecond, 260},
	{10622, 335, 2.8, 0.13793103397, 46, 51, 0.714, 2500 * time.Millisecond, 315},
	{10623, 490, 3.3, 0.13220338523, 54, 59, 0.714, 2500 * time.Millisecond, 405},
}

// Riptide (the 31 point Restoration talent) has three ranks in Forever: the talent teaches rank
// 1 (408521) at 40, the trainer rank 2 (1239242) at 50 and rank 3 (1239243) at 60 (client
// 1.60.1.70753, SkillLineAbility on the Restoration line). In the client each is an instant heal,
// a heal every 3 sec for 15 sec and 25% more on a Chain Heal cast at the target while it's up,
// with a 6 sec cooldown (category 1228, every rank):
//
// | Rank | Learned | Heal | Grows a level, to | Tick | Mana |
// |---|---|---|---|---|---|
// | 1 | 40 | 480 | 3.7, to 48 | 89 | 245 |
// | 2 | 50 | 612 | 4.2, to 58 | 115 | 300 |
// | 3 | 60 | 841 | 4.9, to 68 | 161, 3.7 a level past 60 | 385 |
//
// Blizzard's development notes (post 9, 2026-11-02, devnotes_post9_2026-11-02.txt) take the
// instant heal out, raise the cooldown to 12 sec and cut the mana by "about 40%". Client 70753
// (and 70755) is unchanged and no server hotfix has them, so under Forever the sim models the
// notes (standards 37): ticks only, 12 sec, and 60% of the client's mana rounded to whole mana
// (147 / 180 / 231). The ticks and the Chain Heal bonus are the client's. The ticks'
// coefficient is 0.1 (0.214 was the instant heal's); sim/core takes it down for a rank far below
// our level (downrank.go).
type riptideRank struct {
	healRank
	tick, tickPerLevel float64
}

// TickAt is a tick's heal at the character's level, before healing power.
func (r riptideRank) TickAt(level int32) float64 {
	return r.tick + r.tickPerLevel*float64(min(level, r.maxLevel)-r.learnLevel)
}

var RiptideRanks = []riptideRank{
	{},
	{healRank{408521, 480, 3.7, 0.10000000149, 40, 48, 0.214, 0, 245}, 89, 0},
	{healRank{1239242, 612, 4.2, 0.10000000149, 50, 58, 0.214, 0, 300}, 115, 0},
	{healRank{1239243, 841, 4.9, 0.10000000149, 60, 68, 0.214, 0, 385}, 161, 3.7},
}

const (
	RiptideTickCoefficient = 0.1
	RiptideChainHealBonus  = 0.25
	RiptideTicks           = 5

	// The development notes' Riptide (post 9, 2026-11-02; client 1.60.1.70753 unchanged).
	RiptideNotesManaScale = 0.6
	RiptideNotesCooldown  = 12 * time.Second
	riptideClientCooldown = 6 * time.Second
)

// riptideNotes is whether Riptide takes the development notes' values: no instant heal, a 12 sec
// cooldown and 40% less mana (Forever only, standards 37).
func (shaman *Shaman) riptideNotes() bool {
	return shaman.Env != nil && shaman.Env.IsForever()
}

// RiptideManaCost is a rank's mana: the notes' 60% of the client's under Forever, rounded.
func (shaman *Shaman) RiptideManaCost(rank int) float64 {
	cost := RiptideRanks[rank].manaCost
	if shaman.riptideNotes() {
		cost = math.Round(cost * RiptideNotesManaScale)
	}
	return cost
}

func riptideRankIDs() []int32 {
	ids := make([]int32, 0, len(RiptideRanks)-1)
	for _, r := range RiptideRanks[1:] {
		ids = append(ids, r.spellID)
	}
	return ids
}

func (shaman *Shaman) registerHealingSpells() {
	shaman.HealingWave = shaman.registerHealRanks(HealingWaveRanks, SpellCode_ShamanHealingWave)
	shaman.LesserHealingWave = shaman.registerHealRanks(LesserHealingWaveRanks, SpellCode_ShamanLesserHealingWave)
	shaman.ChainHeal = shaman.registerHealRanks(ChainHealRanks, SpellCode_ShamanChainHeal)
	shaman.registerRiptideSpell()
}

func (shaman *Shaman) registerHealRanks(ranks []healRank, spellCode int32) []*core.Spell {
	spells := make([]*core.Spell, len(ranks))
	for rank := 1; rank < len(ranks); rank++ {
		if shaman.KnowsRank(ranks[rank].spellID, ranks[rank].learnLevel) {
			spells[rank] = shaman.RegisterSpell(shaman.newHealSpellConfig(ranks[rank], rank, spellCode))
		}
	}
	return spells
}

func (shaman *Shaman) newHealSpellConfig(r healRank, rank int, spellCode int32) core.SpellConfig {
	low, high := r.At(shaman.Level)

	castTime := r.castTime
	multiplier := shaman.purificationMultiplier()
	if spellCode == SpellCode_ShamanHealingWave {
		castTime -= shaman.improvedHealingWaveReduction()
		multiplier *= shaman.healingWayMultiplier()
	}

	var spell *core.Spell // this rank, once registered, for ExpectedHealing
	config := core.SpellConfig{
		SpellCode:   spellCode,
		ActionID:    core.ActionID{SpellID: r.spellID},
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       SpellFlagShaman | core.SpellFlagHelpful | core.SpellFlagAPL,

		RequiredLevel: int(r.learnLevel),
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: r.manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: castTime,
			},
		},

		DamageMultiplier: multiplier,
		ThreatMultiplier: 1,
		BonusCoefficient: r.coefficient,
		// Healing Focus: a chance for a hit not to push the heal back.
		PushbackReduction: shaman.healingFocusChance(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			target = shaman.healTarget(target)
			// A heal at a player who died as it landed does nothing, and jumps nowhere.
			if target.IsDead() {
				return
			}
			bonus := shaman.riptideChainHealMultiplier(spell, target)
			spell.DamageMultiplier *= bonus
			spell.CalcAndDealHealing(sim, target, sim.Roll(low, high), spell.OutcomeHealingCrit)
			if spellCode == SpellCode_ShamanChainHeal {
				healed := []*core.Unit{target}
				candidates := shaman.chainHealCandidates(target)
				jump := 1.0
				for len(healed) < ChainHealTargets {
					next := core.MostHurtAmong(candidates, healed...)
					if next == nil {
						break
					}
					jump *= ChainHealJumpMultiplier
					spell.DamageMultiplier *= ChainHealJumpMultiplier
					spell.CalcAndDealHealing(sim, next, sim.Roll(low, high), spell.OutcomeHealingCrit)
					healed = append(healed, next)
				}
				spell.DamageMultiplier /= jump
			}
			spell.DamageMultiplier /= bonus
		},
		// Other healers see the heal in flight at its average (heal_comm.go). Chain Heal's jumps
		// aren't counted, as HealComm counted the target only.
		ExpectedHealing: func(target *core.Unit) float64 {
			if spell == nil {
				spell = shaman.GetSpell(core.ActionID{SpellID: r.spellID})
			}
			target = shaman.healTarget(target)
			return spell.ExpectedHealing(target, (low+high)/2) * shaman.riptideChainHealMultiplier(spell, target)
		},
	}
	return config
}

const (
	ChainHealTargets        = 3
	ChainHealJumpMultiplier = 0.5
)

// healTarget is where a heal cast at target lands: the player it was cast at, or ourselves for
// a cast at a mob or at nothing.
func (shaman *Shaman) healTarget(target *core.Unit) *core.Unit {
	if target == nil || target.Type == core.EnemyUnit {
		return &shaman.Unit
	}
	return target
}

// chainHealCandidates are who a Chain Heal at target can jump to: the target's party when it's
// in ours, the raid otherwise.
func (shaman *Shaman) chainHealCandidates(target *core.Unit) []*core.Unit {
	for _, agent := range shaman.Party.Players {
		if &agent.GetCharacter().Unit == target {
			units := make([]*core.Unit, 0, len(shaman.Party.Players))
			for _, a := range shaman.Party.Players {
				units = append(units, &a.GetCharacter().Unit)
			}
			return units
		}
	}
	return shaman.Env.Raid.AllPlayerUnits
}

func (shaman *Shaman) registerRiptideSpell() {
	shaman.RiptideRanks = make([]*core.Spell, len(RiptideRanks))
	if !shaman.Talents.Riptide {
		return
	}
	cooldown := shaman.NewTimer()
	for rank := 1; rank < len(RiptideRanks); rank++ {
		if r := RiptideRanks[rank]; shaman.KnowsRank(r.spellID, r.learnLevel) {
			shaman.RiptideRanks[rank] = shaman.RegisterSpell(shaman.newRiptideSpellConfig(rank, cooldown))
			shaman.Riptide = shaman.RiptideRanks[rank]
		}
	}
}

func (shaman *Shaman) newRiptideSpellConfig(rank int, timer *core.Timer) core.SpellConfig {
	r := RiptideRanks[rank]
	low, high := r.At(shaman.Level)
	tick := r.TickAt(shaman.Level)
	notes := shaman.riptideNotes()
	coefficient := r.coefficient
	cooldown := riptideClientCooldown
	if notes {
		coefficient, cooldown = 0, RiptideNotesCooldown
	}

	var spell *core.Spell // this rank, once registered, for ExpectedHealing
	return core.SpellConfig{
		SpellCode:   SpellCode_ShamanRiptide,
		ActionID:    core.ActionID{SpellID: r.spellID},
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       SpellFlagShaman | core.SpellFlagHelpful | core.SpellFlagAPL,

		RequiredLevel: int(r.learnLevel),
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: shaman.RiptideManaCost(rank),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    timer,
				Duration: cooldown,
			},
		},

		// Purification takes the heal and its ticks (the client's two effects on Riptide, the
		// ticks' through the snapshot).
		DamageMultiplier: shaman.purificationMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: coefficient,

		Hot: core.DotConfig{
			Aura: core.Aura{
				// Each rank its own aura: ranks sharing a label would share one on the target.
				Label: fmt.Sprintf("Riptide (Rank %d)", rank),
			},
			NumberOfTicks:    RiptideTicks,
			TickLength:       time.Second * 3,
			BonusCoefficient: RiptideTickCoefficient,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.SnapshotHeal(target, tick, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotHealing(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			target = shaman.healTarget(target)
			if target.IsDead() {
				return
			}
			if !notes {
				spell.CalcAndDealHealing(sim, target, sim.Roll(low, high), spell.OutcomeHealingCrit)
			}
			// One Riptide of ours on a target: a rank replaces another's heal over time.
			if hot := shaman.riptideOn(target); hot != nil && hot.Spell != spell {
				hot.Deactivate(sim)
			}
			if hot := spell.Hot(target); hot != nil {
				hot.Apply(sim)
			}
		},
		// For a rotation that casts it when it fits what's missing (APLValueSpellExpectedHealing):
		// the notes' Riptide is all ticks, so the whole heal over time; the client's is the instant
		// heal's average, without the ticks.
		ExpectedHealing: func(target *core.Unit) float64 {
			if spell == nil {
				spell = shaman.GetSpell(core.ActionID{SpellID: r.spellID})
			}
			target = shaman.healTarget(target)
			if !notes {
				return spell.ExpectedHealing(target, (low+high)/2)
			}
			hot := spell.Hot(target)
			return RiptideTicks * spell.ExpectedHealing(target, tick+hot.BonusCoefficient*spell.HealingPower(target)) * hot.DamageMultiplier
		},
	}
}

// riptideOn is our Riptide's heal over time on the target, of whichever rank, or nil.
func (shaman *Shaman) riptideOn(target *core.Unit) *core.Dot {
	for _, spell := range shaman.RiptideRanks {
		if spell == nil {
			continue
		}
		if hot := spell.Hot(target); hot != nil && hot.IsActive() {
			return hot
		}
	}
	return nil
}

// riptideChainHealMultiplier is Riptide's 25% more on a Chain Heal cast at a target that has our
// Riptide, on the whole chain (docs/healing_shaman.md, Calls for the user: the client's effect is
// a server-side dummy, and the tooltip reads "the effectiveness of your Chain Heal casts directly
// on that target").
func (shaman *Shaman) riptideChainHealMultiplier(spell *core.Spell, target *core.Unit) float64 {
	if spell.SpellCode != SpellCode_ShamanChainHeal || shaman.riptideOn(target) == nil {
		return 1
	}
	return 1 + RiptideChainHealBonus
}

// healRankIDs are the spell IDs of every rank of a heal, for its rank family.
func healRankIDs(ranks []healRank) []int32 {
	ids := make([]int32, 0, len(ranks)-1)
	for _, r := range ranks[1:] {
		ids = append(ids, r.spellID)
	}
	return ids
}

// HealRankLevels are the levels a shaman learns each rank of a heal at (HealingWaveRanks,
// LesserHealingWaveRanks), with 0 for rank 0 like the other spells' level tables.
func HealRankLevels(ranks []healRank) []int32 {
	levels := make([]int32, len(ranks))
	for i, r := range ranks {
		levels[i] = r.learnLevel
	}
	return levels
}

// TopHealRankID is the spell ID of a heal's top rank (HealingWaveRanks,
// LesserHealingWaveRanks) once a shaman of this level knows any rank of it, 0 before. A
// rotation that names it casts the highest rank the shaman knows (core.TopRankID).
func TopHealRankID(ranks []healRank, level int32) int32 {
	if ranks[1].learnLevel > level {
		return 0
	}
	return ranks[len(ranks)-1].spellID
}
