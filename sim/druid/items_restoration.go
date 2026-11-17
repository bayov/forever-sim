package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// IdolOfSwiftness (279250, the Forever Restoration guide's idol): Swiftmend's cooldown 3 sec shorter
// (1306488, aura 107 on Swiftmend's cooldown).
const IdolOfSwiftness = 279250

// IdolOfRejuvenation (22398): Forever's is "Increases the initial healing of Regrowth by 5% when cast
// on a target affected by your Rejuvenation", and no longer Unique (client 1.60.1.70938 tooltip,
// https://foreverchanges.pro/item/22398, revealed 2026-11-15). Classic's +50 Rejuvenation healing is
// gone.
const IdolOfRejuvenation = 22398

func init() {
	core.NewItemEffect(IdolOfSwiftness, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		druid.swiftmendCooldownCut += 3 * time.Second
	})
	core.NewItemEffect(IdolOfRejuvenation, func(agent core.Agent) {
		agent.(DruidAgent).GetDruid().regrowthOnRejuvBonus += 0.05
	})
}

// The bonuses of Forever's Tier 1 Restoration druid set (Molten Core; the client has the bonus spells
// 1300943, 1301006, 1301077 and 1301250), which the synthetic Phase 1 Restoration druid set carries on
// two pieces (tools/database/forever_synthetic_items_druid_restoration.go).
var ItemSetRestorationDruidSyntheticPhase1 = core.NewItemSet(core.ItemSet{
	Name: "Restoration Druid Synthetic (Phase 1)",
	Bonuses: map[int32]core.ApplyEffect{
		// +10 Spirit.
		2: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.Spirit, 10)
		},
		// Reduces the cooldown of your Barkskin spell by 10 sec.
		3: func(agent core.Agent) {
			// Nothing to do
		},
		// Increases healing done by up to 26 and damage done by up to 9 for all magical spells and effects.
		4: func(agent core.Agent) {
			agent.GetCharacter().AddStats(stats.Stats{stats.HealingPower: 17, stats.SpellPower: 9})
		},
		// Reduces the cooldown of your Swiftmend spell by 3 sec.
		5: func(agent core.Agent) {
			agent.(DruidAgent).GetDruid().swiftmendCooldownCut += 3 * time.Second
		},
	},
})
