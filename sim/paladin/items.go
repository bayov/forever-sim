package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Libram IDs
const (
	SanctifiedOrb  = 20512
	LibramOfHope   = 22401
	LibramOfFervor = 23203

	// Forever's level 60 librams (client 1.60.1.70338). Libram of Law (spell 1291086) is a 4%
	// damage modifier whose spell mask holds Judgement of Righteousness (and Judgement of Fury)
	// but not Judgement of Command, though its tooltip says "your Judgement ability". Libram of
	// Infusion (spell 1306429) is 6% crit on Holy Shock.
	LibramOfLaw      = 272435
	LibramOfInfusion = 279248

	// Tenets of the Silver Hand (spell 1302545) is 1% more damage against Undead, and Libram of
	// Invocation 5% off the seals' mana cost (sealCostFactor).
	TenetsOfTheSilverHand = 249397
	LibramOfInvocation    = 249442
)

func init() {
	core.NewSimpleStatOffensiveTrinketEffect(SanctifiedOrb, stats.Stats{stats.MeleeCrit: 3 * core.CritRatingPerCritChance, stats.SpellCrit: 3 * core.CritRatingPerCritChance}, time.Second*25, time.Minute*3)

	// Tenets of the Silver Hand's equip spell 1302545 is A_MOD_DAMAGE_DONE_VERSUS with base points 1
	// on mask 32: 1% more damage of every kind against Undead. It raises damage alone, not crit
	// damage, as wowsims/forever has it (item_librams.go).
	core.NewItemEffect(TenetsOfTheSilverHand, func(agent core.Agent) {
		c := agent.GetCharacter()
		undead := core.FilterSlice(c.Env.Encounter.TargetUnits, func(unit *core.Unit) bool {
			return unit.MobType == proto.MobType_MobTypeUndead
		})
		core.MakePermanent(c.RegisterAura(core.Aura{
			Label:    "Tenets of the Silver Hand",
			ActionID: core.ActionID{SpellID: 1302545},
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				for _, target := range undead {
					for _, at := range c.AttackTables[target.UnitIndex] {
						at.DamageDealtMultiplier *= 1.01
					}
				}
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				for _, target := range undead {
					for _, at := range c.AttackTables[target.UnitIndex] {
						at.DamageDealtMultiplier /= 1.01
					}
				}
			},
		}))
	})

	// The Judgement of the Crusader PvP gloves (Knight-Lieutenant's Lamellar Gauntlets 16410 and
	// 23274, Marshal's Lamellar Gloves 16471) carry spell 23300, a flat 20 on the judgement's Holy
	// damage bonus (A_ADD_FLAT_MODIFIER, all effects, Judgement of the Crusader's class mask).
	for _, itemID := range []int32{16410, 16471, 23274} {
		core.NewItemEffect(itemID, func(agent core.Agent) {
			paladin := agent.(PaladinAgent).GetPaladin()
			if paladin.Env.IsForever() {
				paladin.judgementOfTheCrusaderBonus += 20
			}
		})
	}
}
