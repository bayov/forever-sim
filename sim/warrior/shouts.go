package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const ShoutExpirationThreshold = time.Second * 3

func (warrior *Warrior) newShoutSpellConfig(actionID core.ActionID, rank int32, allyAuras core.AuraArray) *WarriorSpell {
	// Use extra hits to simulate buffing your party for threat
	extraHits := 5 - len(allyAuras)

	return warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL | core.SpellFlagHelpful,

		RageCost: core.RageCostOptions{
			Cost: warrior.shoutRageCost(),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		FlatThreatBonus: float64(core.BattleShoutLevel[rank]),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for _, aura := range allyAuras {
				spell.CalcAndDealOutcome(sim, aura.Unit, spell.OutcomeAlwaysHit)
				aura.Activate(sim)
			}

			for i := 0; i < extraHits; i++ {
				spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
			}
		},

		RelatedAuras: []core.AuraArray{allyAuras},
	})
}

func (warrior *Warrior) registerBattleShout() {
	rank := core.TernaryInt32(core.IncludeAQ, 7, 6)
	actionId := core.BattleShoutSpellId[rank]
	has3pcWrath := warrior.HasSetBonus(ItemSetBattleGearOfWrath, 3)

	warrior.BattleShout = warrior.newShoutSpellConfig(core.ActionID{SpellID: actionId}, rank, warrior.NewPartyAuraArray(func(unit *core.Unit) *core.Aura {
		// Improved Battle Shout is gone from the Forever tree. Under Forever BattleShoutAura uses
		// the client's 139 at 60 and ignores the talent points passed here.
		return core.BattleShoutAura(unit, 5, 0, has3pcWrath)
	}))
}

// Forever's Booming Voice takes 5% a point off the shouts' Rage cost (client 70338: curve -5 to
// -25 on spell 12321, on the shout spell mask).
func (warrior *Warrior) shoutRageCost() float64 {
	return 10 * (1 - 0.05*float64(warrior.Talents.BoomingVoice))
}

func (warrior *Warrior) registerShouts() {
	warrior.registerBattleShout()
}
