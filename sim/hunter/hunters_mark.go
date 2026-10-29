package hunter

import (
	"github.com/wowsims/classic/sim/core"
)

// Hunter's Mark as the hunter's own cast. Every level 60 preset keeps up its own (standards 23,
// user ruling 2026-10-25: Mark duty is assumed): it switches the raid's Hunter's Mark debuff off and
// keeps the mark up itself, a global cooldown and the mana each time. Under Forever the mark is the same one the debuff gives (core.ForeverHuntersMarkAura),
// 2 min long; rank 4 is 60 mana (client 14325, wowhead's Forever data).
//
// Only one hunter in a raid needs to mark the target. With the skip_hunters_mark option another
// hunter keeps it up, so this one never casts it: the rotation's Hunter's Mark lines do nothing, and
// the settings' Hunter's Mark (or the other hunter's cast one) stands in.
func (hunter *Hunter) registerHuntersMarkSpell() {
	levels := [5]int32{0, 6, 22, 40, 58}
	rank := 0
	for r := 4; r >= 1; r-- {
		if levels[r] <= hunter.Level {
			rank = r
			break
		}
	}
	if rank == 0 {
		return
	}
	spellID := [5]int32{0, 1130, 14323, 14324, 14325}[rank]
	manaCost := [5]float64{0, 15, 30, 45, 60}[rank]

	auras := hunter.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		if hunter.Env.IsForever() {
			return core.ForeverHuntersMarkAura(target, hunter.Level)
		}
		return core.HuntersMarkAura(target, 0)
	})

	config := core.SpellConfig{
		ActionID:      core.ActionID{SpellID: spellID},
		ProcMask:      core.ProcMaskEmpty,
		Flags:         core.SpellFlagAPL,
		Rank:          rank,
		RequiredLevel: int(levels[rank]),

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true, // Hunter GCD is locked at 1.5s
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			auras.Get(target).Activate(sim)
		},
	}
	if hunter.Options.GetSkipHuntersMark() {
		config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool { return false }
		// It never changes, so the rotation needn't wake up for it.
		config.ExtraCastConditionNeedsEvent = true
	}
	hunter.HuntersMark = hunter.RegisterSpell(config)
}
