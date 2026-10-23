package priest

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (priest *Priest) registerVampiricEmbraceSpell() {
	if !priest.Talents.VampiricEmbrace {
		return
	}

	actionID := core.ActionID{SpellID: 15286}
	manaCost := 40.0
	// Forever: 30 sec (Classic 1 min) on a 30 sec cooldown (client 1.60.1.70425; 1 min in
	// 70338), and Improved Vampiric Embrace is gone, so it heals the party for 20% of the
	// Shadow damage.
	duration := time.Second * 30

	partyPlayers := priest.Env.Raid.GetPlayerParty(&priest.Unit).Players
	healthMetrics := priest.NewHealthMetrics(actionID)
	healthReturnedMultuplier := 0.2

	priest.VampiricEmbraceAuras = priest.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			ActionID: actionID,
			Label:    "Vampiric Embrace (Health) - " + target.Label,
			Duration: duration,
			OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if result.Landed() && spell.SpellSchool.Matches(core.SpellSchoolShadow) {
					healthGained := result.Damage * healthReturnedMultuplier
					for _, player := range partyPlayers {
						player.GetCharacter().GainHealth(sim, healthGained, healthMetrics)
					}
				}
			},
		})
	})

	// Under Forever it is a harmful spell cast like any other, so it rolls on-cast procs such as
	// Vestments of the Devout's 5 pieces (wowsims/forever casts it with the spell damage proc
	// mask, vampiric_embrace.go).
	procMask := core.ProcMaskEmpty
	if priest.Env.IsForever() {
		procMask = core.ProcMaskSpellDamage
	}

	priest.VampiricEmbrace = priest.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolShadow,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    procMask,
		Flags:       SpellFlagPriest | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    priest.NewTimer(),
				Duration: time.Second * 30,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Forever 70425 made it unresistable (SpellMisc Attributes_3 +0x40000, the bit Shadow
			// Weaving got in 70170).
			outcome := spell.OutcomeMagicHit
			if priest.Env.IsForever() {
				outcome = spell.OutcomeAlwaysHit
			}
			result := spell.CalcOutcome(sim, target, outcome)
			if result.Landed() {
				priest.VampiricEmbraceAuras.Get(target).Activate(sim)
			}
			spell.DealOutcome(sim, result)
		},
	})
}
