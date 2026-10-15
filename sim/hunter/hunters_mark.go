package hunter

import (
	"github.com/wowsims/classic/sim/core"
)

// Hunter's Mark as the hunter's own cast, for the presets on Hunter's Mark duty (standards 23):
// they switch the raid's Hunter's Mark debuff off and keep the mark up themselves, a global
// cooldown and the mana each time. The base presets take the mark from the raid's debuff and
// never cast it. Under Forever the mark is the same one the debuff gives (core.ForeverHuntersMarkAura),
// 2 min long; rank 4 is 60 mana (client 14325, wowhead's Forever data).
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

	hunter.HuntersMark = hunter.RegisterSpell(core.SpellConfig{
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
	})
}
