package core

// Forever's downranking penalty.
//
// Blizzard's beta development notes (2026-10-01, us.forums.blizzard.com/en/wow/t/2360696/4) say
// ranks "vastly below your current level receive less benefit from your Spell Damage and Spell
// Healing". They give no numbers. https://foreverchanges.pro/downranking measured the size from
// beta heals at level 30 and below: a rank keeps 1 - 5% for each level the caster is more than
// 17 levels past the rank's learn level, clamped to [0, 1]. Classic keeps its own rules.
//
// Which spells take it is open. The 2026-10-10 beta damage log (TestOrcShamanSpellPowerBeta)
// shows the shaman shocks at their full coefficient where the formula would cut them, so the
// penalty may be heals only. Until that is ruled, the damage spells that took it before it moved
// here (Lightning Bolt, Chain Lightning, Smite, Holy Fire, Mind Blast, Shadow Word: Pain and
// Devouring Plague) keep the assumed penalty. See notes.md, "Forever's downranking penalty".

// DownrankKind is which side of the penalty's scope a spell sits on. A spell opts in by
// setting SpellConfig.Downrank; the zero value keeps its coefficient.
type DownrankKind uint8

const (
	DownrankNone   DownrankKind = 0
	DownrankHeal   DownrankKind = 1 << 0 // A heal, the kind the measurement was made on.
	DownrankDamage DownrankKind = 1 << 1 // A damage spell, the penalty assumed from the notes.
)

// ForeverDownrankScope is the kinds the penalty hits under Forever. "Heals only" is
// DownrankHeal.
var ForeverDownrankScope = DownrankHeal | DownrankDamage

// DownrankFactor is what a rank learned at learnLevel keeps of its spell power coefficient
// when cast at casterLevel.
func DownrankFactor(casterLevel int32, learnLevel int) float64 {
	return max(0, min(1, 1-0.05*float64(int(casterLevel)-17-learnLevel)))
}

// applyDownrank scales a ranked spell's coefficients (direct, DoT and HoT) by DownrankFactor,
// read from its RequiredLevel, when it is in ForeverDownrankScope under Forever.
func (unit *Unit) applyDownrank(config *SpellConfig) {
	if config.Downrank&ForeverDownrankScope == 0 || config.Rank <= 0 || unit.Env == nil || !unit.Env.IsForever() {
		return
	}
	factor := DownrankFactor(unit.Level, config.RequiredLevel)
	config.BonusCoefficient *= factor
	config.Dot.BonusCoefficient *= factor
	config.Hot.BonusCoefficient *= factor
}
