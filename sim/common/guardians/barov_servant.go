package guardians

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// https://www.wowhead.com/forever/item=14022/barov-peasant-caller
// https://www.wowhead.com/classic/npc=11637/servant-of-alexi-barov
//
// Use: Death by Peasant (18307) summons three Servants of Alexi Barov that melee our target.
//
// The numbers come from three uses in the user's Forever combat logs on 2026-11-20 and 21
// (WoWCombatLog-112026_111427 lines 297072 and 313526, WoWCombatLog-112126_201216 line 16005).
// All three servants appear in the same millisecond as the cast and swing in sync every 2.0 sec,
// with the first swing on the summon. Their raw (pre-armor) hits were 41 to 57 (about 48 on
// average), with 1 crit and 2 dodges in 61 swings. The log shows 263 AP on each servant, but we
// fold that into the weapon damage because the hits we measured already include it. The last
// swings landed at 14.0 to 14.3 sec, which fits a 15 sec duration, but the log has no despawn line.
// Mobs never attacked them.
var (
	BarovPeasantCallerHorde    = int32(14022)
	BarovPeasantCallerAlliance = int32(14023)
)

const BarovServantCount = 3

type BarovServant struct {
	core.Pet
}

func NewBarovServant(character *core.Character, index int) *BarovServant {
	servant := &BarovServant{
		Pet: core.NewPet("Servant of Alexi Barov "+string(rune('1'+index)), character, stats.Stats{
			stats.Health: 926,
		}, func(stats.Stats) stats.Stats { return stats.Stats{} }, false, true),
	}
	// The log shows the servants at the summoner's level.
	servant.Level = character.Level

	servant.EnableAutoAttacks(servant, core.AutoAttackOptions{
		MainHand: core.Weapon{
			BaseDamageMin: 41,
			BaseDamageMax: 57,
			SwingSpeed:    2.0,
			SpellSchool:   core.SpellSchoolPhysical,
		},
		AutoSwingMelee: true,
	})

	return servant
}

func (servant *BarovServant) Initialize() {
}

func (servant *BarovServant) ExecuteCustomRotation(sim *core.Simulation) {
	// The guard the other guardians use (core_hound.go): a swing due now, or a wait into the past,
	// panicked with "cannot wait negative time" (Subtlety's trinket use values, 2026-11-26).
	if servant.AutoAttacks.NextAttackAt() != sim.CurrentTime+servant.AutoAttacks.MainhandSwingSpeed() && servant.AutoAttacks.NextAnyAttackAt()-1 > sim.CurrentTime {
		servant.WaitUntil(sim, servant.AutoAttacks.NextAttackAt()-1)
	}
}

func (servant *BarovServant) Reset(sim *core.Simulation) {
	servant.Disable(sim)
}

func (servant *BarovServant) OnPetDisable(sim *core.Simulation) {
}

func (servant *BarovServant) GetPet() *core.Pet {
	return &servant.Pet
}

func constructBarovServants(character *core.Character) {
	if !character.HasTrinketEquipped(BarovPeasantCallerHorde) && !character.HasTrinketEquipped(BarovPeasantCallerAlliance) {
		return
	}
	for i := 0; i < BarovServantCount; i++ {
		character.AddPet(NewBarovServant(character, i))
	}
}
