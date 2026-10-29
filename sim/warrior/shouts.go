package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const ShoutExpirationThreshold = time.Second * 3

// newShoutSpellConfig is a shout that puts allyAuras on its receivers: a raid aura array,
// which has nil entries for the pets and enemies.
//
// Each receiver adds the shout's threat. Extra hits on the target stand in for the party
// members a smaller party lacks, so a lone warrior's shout still counts as buffing a party of 5.
func (warrior *Warrior) newShoutSpellConfig(actionID core.ActionID, rank int32, allyAuras core.AuraArray) *WarriorSpell {
	receivers := core.FilterSlice(allyAuras, func(aura *core.Aura) bool { return aura != nil })
	partyReceivers := 0
	for _, aura := range receivers {
		if warrior.Env.Raid.GetPlayerParty(aura.Unit) == warrior.Party {
			partyReceivers++
		}
	}
	extraHits := max(0, 5-partyReceivers)

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
			for _, aura := range receivers {
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
	// Ranks from the Ahn'Qiraj books are out under Forever (no source in client 70338).
	aq := warrior.Env.IncludeAQRanks()
	rank := core.TernaryInt32(aq, 7, 6)
	actionId := core.BattleShoutSpellId[rank]
	has3pcWrath := warrior.HasSetBonus(ItemSetBattleGearOfWrath, 3)

	// It reaches every player in the raid (client 70685: 25289 targets the caster's raid,
	// players only), and the strongest of several warriors' shouts counts.
	warrior.BattleShout = warrior.newShoutSpellConfig(core.ActionID{SpellID: actionId}, rank, warrior.NewRaidPlayerAuraArray(func(unit *core.Unit) *core.Aura {
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
