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
	// 70338), and Improved Vampiric Embrace is gone. What it heals is vampiricEmbraceHealing.
	duration := time.Second * 30

	partyPlayers := priest.Env.Raid.GetPlayerParty(&priest.Unit).Players
	healthMetrics := priest.NewHealthMetrics(actionID)
	priest.VampiricEmbraceAuras = priest.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			ActionID: actionID,
			Label:    "Vampiric Embrace (Health) - " + target.Label,
			Duration: duration,
			// Only the priest's own Shadow damage heals ("based on 10% of the Shadow damage done by the
			// casting Priest").
			OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if spell.Unit == &priest.Unit && result.Landed() && spell.SpellSchool.Matches(core.SpellSchoolShadow) {
					for _, player := range partyPlayers {
						char := player.GetCharacter()
						char.GainHealth(sim, vampiricEmbraceHealing(result.Damage, char == &priest.Character), healthMetrics)
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

// vampiricEmbraceHealing is what a party member gains from the casting priest's Shadow damage
// under Vampiric Embrace: 10% of it, and 20% for the priest, who "receives 100% increased healing
// from this effect" (Blizzard's hotfixes of 2026-11-16, t/2403040/11, live with the weekly
// maintenance of November 17/18; before, everyone got 20%).
func vampiricEmbraceHealing(damage float64, isCaster bool) float64 {
	healing := 0.1 * damage
	if isCaster {
		healing *= 2
	}
	return healing
}
