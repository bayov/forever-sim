package core

// Forever's downranking penalty.
//
// Blizzard's beta development notes (2026-10-01, us.forums.blizzard.com/en/wow/t/2360696/4) say
// ranks "vastly below your current level receive less benefit from your Spell Damage and Spell
// Healing". They give no numbers. https://foreverchanges.pro/downranking measured the size from
// beta heals at level 30 and below: a rank keeps 1 - 5% for each level the caster is more than
// 17 levels past the rank's learn level, clamped to [0, 1]. Classic keeps its own rules.
//
// It applies to every ranked spell, heals and damage alike (user ruling, 2026-11-04). The
// 2026-10-10 beta log (TestOrcShamanSpellPowerBeta) has Earth Shock and Flame Shock rank 1 at
// full coefficient at level 30, but that log is taken to predate a later beta fix. See notes.md,
// "Forever's downranking penalty".

// DownrankKind is how a spell sits in the penalty's scope. The zero value, DownrankAuto, reads
// the kind from the spell: a heal if its ProcMask has ProcMaskSpellHealing, else damage.
type DownrankKind uint8

const (
	DownrankAuto   DownrankKind = 0
	DownrankHeal   DownrankKind = 1 << 0
	DownrankDamage DownrankKind = 1 << 1
	// DownrankExempt keeps the spell's coefficient: for a ranked spell whose RequiredLevel is
	// not the learn level of a rank the player casts (an imbue's proc, say).
	DownrankExempt DownrankKind = 1 << 7
)

// ForeverDownrankScope is the kinds the penalty hits under Forever: heals and damage.
var ForeverDownrankScope = DownrankHeal | DownrankDamage

// DownrankFactor is what a rank learned at learnLevel keeps of its spell power coefficient
// when cast at casterLevel.
func DownrankFactor(casterLevel int32, learnLevel int) float64 {
	return max(0, min(1, 1-0.05*float64(int(casterLevel)-17-learnLevel)))
}

// applyDownrank scales a spell's coefficients (direct, DoT and HoT) by DownrankFactor when its
// kind is in ForeverDownrankScope under Forever. The learn level is DownrankLevel when set (a
// rank-less sub-spell such as a channel's tick, a totem's attack or an imbue's proc), else a
// ranked spell's RequiredLevel. A spell with neither keeps its coefficients, and so does a
// physical one, whose coefficient scales weapon bonus damage rather than spell power.
func (unit *Unit) applyDownrank(config *SpellConfig) {
	learnLevel := config.DownrankLevel
	if learnLevel == 0 && config.Rank > 0 {
		learnLevel = config.RequiredLevel
	}
	if learnLevel <= 0 || config.SpellSchool == SpellSchoolPhysical || unit.Env == nil || !unit.Env.IsForever() {
		return
	}
	kind := config.Downrank
	if kind == DownrankAuto {
		kind = Ternary(config.ProcMask.Matches(ProcMaskSpellHealing), DownrankHeal, DownrankDamage)
	}
	if kind&ForeverDownrankScope == 0 {
		return
	}
	factor := DownrankFactor(unit.Level, learnLevel)
	config.BonusCoefficient *= factor
	config.Dot.BonusCoefficient *= factor
	config.Hot.BonusCoefficient *= factor
}
