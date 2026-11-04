// Package healing is the Holy and Discipline priest of the healing sim (docs/healing_priest.md,
// H3). Its heals, shields and healing talents are the priest's (sim/priest/healing.go,
// healing_talents.go); this agent is a caster that doesn't melee and heals what its rotation aims
// at.
package healing

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/priest"
)

func RegisterHealingPriest() {
	core.RegisterAgentFactory(
		proto.Player_HealingPriest{},
		proto.Spec_SpecHealingPriest,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewHealingPriest(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_HealingPriest)
			if !ok {
				panic("Invalid spec value for Healing Priest!")
			}
			player.Spec = playerSpec
		},
	)
}

type HealingPriest struct {
	*priest.Priest

	Options *proto.HealingPriest_Options
}

func NewHealingPriest(character *core.Character, options *proto.Player) *HealingPriest {
	healingOptions := options.GetHealingPriest()

	// A healer stands back and casts: no auto attacks.
	return &HealingPriest{
		Priest:  priest.New(character, options.TalentsString),
		Options: healingOptions.GetOptions(),
	}
}

func (hpriest *HealingPriest) GetPriest() *priest.Priest {
	return hpriest.Priest
}

func (hpriest *HealingPriest) Initialize() {
	hpriest.Priest.Initialize()
	hpriest.Priest.RegisterHealingSpells()
}

func (hpriest *HealingPriest) Reset(_ *core.Simulation) {
	hpriest.Priest.ResetHealing()
}
