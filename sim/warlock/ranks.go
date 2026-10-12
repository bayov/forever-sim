package warlock

import (
	"time"
)

// spellRank is one rank of a warlock spell as the Forever client (1.60.1.70338, wago.tools
// SpellEffect, SpellLevels, SpellPower, SpellMisc) has it. Forever cut most damage to about
// half of 1.12 and gave the damage over time spells bigger coefficients.
//
// Damage is Base + PerLevel for each level past Level, up to MaxLevel (the client's
// RealPointsPerLevel and SpellLevels), and a direct hit rolls Variance around that
// (Base 268 and Variance 0.1098 is Shadow Bolt rank 10's 253 to 283 at 60). Over time
// effects have no variance and Base is the damage of one tick.
type spellRank struct {
	SpellID  int32
	Level    int
	MaxLevel int
	Base     float64
	PerLevel float64
	Variance float64
	Coef     float64
	Cost     float64
	CastMs   int
}

// damageAt is the rank's average damage at the character's level.
func (r spellRank) damageAt(level int32) float64 {
	lvl := int(level)
	if r.MaxLevel > 0 && lvl > r.MaxLevel {
		lvl = r.MaxLevel
	}
	if lvl < r.Level {
		lvl = r.Level
	}
	return r.Base + r.PerLevel*float64(lvl-r.Level)
}

// rangeAt is the low and high roll of a direct damage rank at the character's level.
func (r spellRank) rangeAt(level int32) (float64, float64) {
	avg := r.damageAt(level)
	return avg * (1 - r.Variance/2), avg * (1 + r.Variance/2)
}

func (r spellRank) castTime() time.Duration {
	return time.Millisecond * time.Duration(r.CastMs)
}
