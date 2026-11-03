package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// The paladin's heals (docs/healing_paladin.md). A heal cast at a player lands on that player (a
// healer's rotation aims them with Most Hurt, Lowest Health or a tank, docs/healing_sim.md); one
// cast at a mob, the rotation's default target, lands on the paladin.
//
// Every value is the Forever client's (1.60.1.70753, the build that carries the launch realms'
// October 30 hotfix: SpellEffect, SpellLevels, SpellMisc, SpellPower and SpellCastTimes on
// wago.tools), read the way the shaman's heals are (sim/shaman/healing.go): a rank heals its base
// at the level it's learned plus its growth per level up to its max level, spread evenly by the
// variance around that, base * (1 - variance / 2) to base * (1 + variance / 2). The level 38
// footage of the BlizzCon build agrees (Holy Light rank 6 610 to 682, Flash of Light rank 3 101 to
// 114, forever-wiki/classes/paladin.md). The spell power coefficients are the client's, and
// Forever's downranking takes them down for a rank learned far below our level
// (downrankCoefficient).
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

// Holy Light: 2.5 sec, 71.4% of healing power on every rank.
var HolyLightRanks = []healRank{
	{},
	{635, 43, 0.8, 0.18604651093, 1, 5, 0.714, 2500 * time.Millisecond, 35},
	{639, 83, 1.1, 0.16867469251, 6, 11, 0.714, 2500 * time.Millisecond, 60},
	{647, 155, 1.7, 0.16184970737, 14, 19, 0.714, 2500 * time.Millisecond, 110},
	{1026, 287, 2.4, 0.13813814521, 22, 27, 0.714, 2500 * time.Millisecond, 190},
	{1042, 452, 3.1, 0.11877393723, 30, 35, 0.714, 2500 * time.Millisecond, 275},
	{3472, 646, 3.8, 0.11096075922, 38, 43, 0.714, 2500 * time.Millisecond, 365},
	{10328, 899, 4.6, 0.10810811073, 46, 51, 0.714, 2500 * time.Millisecond, 465},
	{10329, 1217, 5.2, 0.10782080144, 54, 59, 0.714, 2500 * time.Millisecond, 580},
	{25292, 1580, 5.8, 0.10714285821, 60, 65, 0.714, 2500 * time.Millisecond, 660},
}

// Flash of Light: 1.5 sec, 42.9% of healing power.
var FlashOfLightRanks = []healRank{
	{},
	{19750, 46, 1.0, 0.14925372601, 20, 25, 0.429, 1500 * time.Millisecond, 35},
	{19939, 66, 1.3, 0.13592232764, 26, 31, 0.429, 1500 * time.Millisecond, 50},
	{19940, 101, 1.6, 0.11688310653, 34, 39, 0.429, 1500 * time.Millisecond, 70},
	{19941, 151, 1.9, 0.11483254284, 42, 47, 0.429, 1500 * time.Millisecond, 90},
	{19942, 223, 2.2, 0.11307421327, 50, 55, 0.429, 1500 * time.Millisecond, 115},
	{19943, 303, 2.6, 0.11019284278, 58, 63, 0.429, 1500 * time.Millisecond, 140},
}

// Holy Shock's heal (the talent's rank 1 at 30 and the trainer ranks): the heal spell each cast
// spell triggers on an ally (1311605, 25914, 25913, 25903), 42.9% of healing power and no growth
// per level. Its mana and cooldown are the cast spell's (holy_shock.go).
var HolyShockHealRanks = []healRank{
	{},
	{1311605, 114, 0, 0.07547169924, 30, 37, 0.429, 0, 160},
	{25914, 156, 0, 0.07547169924, 40, 47, 0.429, 0, 225},
	{25913, 230, 0, 0.07586207241, 48, 55, 0.429, 0, 275},
	{25903, 320, 0, 0.07894736528, 56, 60, 0.429, 0, 325},
}

// Light's Vigil's heal on the target and its party (1310912, 1311591, 1311596), 14.3% of healing
// power. The mana and cast time are the Vigil's (lights_vigil.go).
var LightsVigilHealRanks = []healRank{
	{},
	{1310912, 324, 1.2, 0.05621301755, 40, 49, 0.143, 1500 * time.Millisecond, 730},
	{1311591, 487, 1.5, 0.05621301755, 50, 59, 0.143, 1500 * time.Millisecond, 1000},
	{1311596, 704, 1.8, 0.05621301755, 60, 69, 0.143, 1500 * time.Millisecond, 1340},
}

func init() {
	core.RegisterSpellRanks(healRankIDs(HolyLightRanks)...)
	core.RegisterSpellRanks(healRankIDs(FlashOfLightRanks)...)
}

// healRankIDs are the spell IDs of every rank of a heal, for its rank family.
func healRankIDs(ranks []healRank) []int32 {
	ids := make([]int32, 0, len(ranks)-1)
	for _, r := range ranks[1:] {
		ids = append(ids, r.spellID)
	}
	return ids
}

// HealRankLevels are the levels a paladin learns each rank of a heal at, with 0 for rank 0 like
// the other spells' level tables.
func HealRankLevels(ranks []healRank) []int32 {
	levels := make([]int32, len(ranks))
	for i, r := range ranks {
		levels[i] = r.learnLevel
	}
	return levels
}

// downrankCoefficient is what a rank keeps of its spell power coefficient under Forever: 5% less
// for each level the caster is past the rank's learn level + 17, down to nothing (Blizzard's
// 2026-10-01 development notes, https://foreverchanges.pro/downranking; the shaman's and the
// priest's rule). Holy Light rank 6 (learned at 38) keeps 75% at 60.
func (paladin *Paladin) downrankCoefficient(coefficient float64, learnLevel int32) float64 {
	return coefficient * paladin.downrankFactor(learnLevel)
}

func (paladin *Paladin) downrankFactor(learnLevel int32) float64 {
	if paladin.Env == nil || !paladin.Env.IsForever() {
		return 1
	}
	return max(0, min(1, 1-0.05*float64(paladin.Level-17-learnLevel)))
}

// healTarget is where a heal cast at target lands: the player it was cast at, or ourselves for a
// cast at a mob or at nothing.
func (paladin *Paladin) healTarget(target *core.Unit) *core.Unit {
	if target == nil || target.Type == core.EnemyUnit {
		return &paladin.Unit
	}
	return target
}

// The paladin's heals that Healing Light, Divine Favor and the Oath of Light raise.
func isPaladinHeal(spell *core.Spell) bool {
	switch spell.SpellCode {
	case SpellCode_PaladinHolyLight, SpellCode_PaladinFlashOfLight, SpellCode_PaladinHolyShockHeal:
		return true
	}
	return false
}

// healingLightMultiplier is Healing Light's 4% a rank on Holy Light, Flash of Light and Holy
// Shock's heal (spell 20237, aura 108 on their class mask, the trait ranks 4 / 8 / 12%).
func (paladin *Paladin) healingLightMultiplier() float64 {
	return 1 + 0.04*float64(paladin.Talents.HealingLight)
}

// holyPowerCritBonus is Holy Power since the October 30 hotfix (spell 5923: aura 50, critical
// healing amount, Holy school, the trait ranks 10 to 50%): the CritDamageBonus that makes a heal's
// whole crit 10% bigger a rank, 150% to 225% at 5/5. That is how aura 50 works on the server's
// crit heal (it multiplies the critical heal, as the user's ruling on Tauren Endurance's 2% read
// the same aura, notes.md), so the bonus part (50%) grows by 0.15 a rank: CritDamageBonus 1 + 0.3
// a rank. Calls for the user (docs/healing_paladin.md) has the other readings.
func (paladin *Paladin) holyPowerCritBonus() float64 {
	return 0.3 * float64(paladin.Talents.HolyPower)
}

// spiritualFocusChance is Spiritual Focus's chance for a hit not to push back Holy Light, Flash of
// Light or Light's Vigil: 14% a rank of 5 (spell 20205, aura 108 misc 9).
func (paladin *Paladin) spiritualFocusChance() float64 {
	return 0.14 * float64(paladin.Talents.SpiritualFocus)
}

func (paladin *Paladin) registerHealingSpells() {
	paladin.HolyLight = paladin.registerHealRanks(HolyLightRanks, SpellCode_PaladinHolyLight)
	paladin.FlashOfLight = paladin.registerHealRanks(FlashOfLightRanks, SpellCode_PaladinFlashOfLight)
}

func (paladin *Paladin) registerHealRanks(ranks []healRank, spellCode int32) []*core.Spell {
	spells := make([]*core.Spell, len(ranks))
	for rank := 1; rank < len(ranks); rank++ {
		if paladin.KnowsRank(ranks[rank].spellID, ranks[rank].learnLevel) {
			spells[rank] = paladin.RegisterSpell(paladin.newHealSpellConfig(ranks[rank], rank, spellCode))
		}
	}
	return spells
}

func (paladin *Paladin) newHealSpellConfig(r healRank, rank int, spellCode int32) core.SpellConfig {
	low, high := r.At(paladin.Level)
	oath := paladin.oathOfLightBonus(spellCode, r.learnLevel)

	var spell *core.Spell // this rank, once registered, for ExpectedHealing
	config := core.SpellConfig{
		SpellCode:   spellCode,
		ActionID:    core.ActionID{SpellID: r.spellID},
		SpellSchool: core.SpellSchoolHoly,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       core.SpellFlagHelpful | core.SpellFlagAPL,

		RequiredLevel: int(r.learnLevel),
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: r.manaCost,
			Factor:   paladin.healCostFactor(spellCode),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: r.castTime,
			},
		},

		DamageMultiplier: paladin.healingLightMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: paladin.downrankCoefficient(r.coefficient, r.learnLevel),
		CritDamageBonus:  paladin.holyPowerCritBonus(),
		// Spiritual Focus: a chance for a hit not to push the heal back.
		PushbackReduction: paladin.spiritualFocusChance(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			target = paladin.healTarget(target)
			// A heal at a player who died as it landed does nothing.
			if target.IsDead() {
				return
			}
			spell.CalcAndDealHealing(sim, target, sim.Roll(low, high)+paladin.oathOn(target, oath), spell.OutcomeHealingCrit)
		},
		// Other healers see the heal in flight at its average (heal_comm.go).
		ExpectedHealing: func(target *core.Unit) float64 {
			if spell == nil {
				spell = paladin.GetSpell(core.ActionID{SpellID: r.spellID})
			}
			target = paladin.healTarget(target)
			return spell.ExpectedHealing(target, (low+high)/2+paladin.oathOn(target, oath))
		},
	}
	if spellCode == SpellCode_PaladinHolyLight {
		// Infusion of Light takes its 0.5 sec a rank off the next Holy Light (talents in
		// holy_talents.go).
		config.Cast.CastTime = func(spell *core.Spell) time.Duration {
			return paladin.ApplyCastSpeedForSpell(spell.DefaultCast.CastTime-paladin.infusionOfLightReduction(), spell)
		}
	}
	return config
}

// LibramOfEconomy (272436, spell 1291089, a Stranglethorn drop like Libram of Law) takes 5% off Holy
// Light's mana cost (aura 108 misc 14 on its class mask).
const LibramOfEconomy = 272436

func (paladin *Paladin) healCostFactor(spellCode int32) float64 {
	if spellCode == SpellCode_PaladinHolyLight && paladin.Ranged().ID == LibramOfEconomy {
		return 0.95
	}
	return 1
}

// TopHealRankID is the spell ID of a heal's top rank once a paladin of this level knows any rank
// of it, 0 before. A rotation that names it casts the highest rank the paladin knows.
func TopHealRankID(ranks []healRank, level int32) int32 {
	if ranks[1].learnLevel > level {
		return 0
	}
	return ranks[len(ranks)-1].spellID
}
