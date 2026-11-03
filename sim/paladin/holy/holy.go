// Package holy is the Holy paladin, a healer of the healing sim (docs/healing_sim.md, H3, and
// docs/healing_paladin.md). Its heals and healing talents are the paladin's (sim/paladin); this
// agent is a caster that doesn't melee and heals what its rotation aims at.
package holy

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/paladin"
)

func RegisterHolyPaladin() {
	core.RegisterAgentFactory(
		proto.Player_HolyPaladin{},
		proto.Spec_SpecHolyPaladin,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewHolyPaladin(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_HolyPaladin)
			if !ok {
				panic("Invalid spec value for Holy Paladin!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewHolyPaladin(character *core.Character, options *proto.Player) *HolyPaladin {
	holyOptions := options.GetHolyPaladin().GetOptions()
	if holyOptions == nil {
		holyOptions = &proto.PaladinOptions{}
	}

	// A healer stands back and casts: no auto attacks.
	return &HolyPaladin{
		Paladin: paladin.NewPaladin(character, options, holyOptions),
	}
}

type HolyPaladin struct {
	*paladin.Paladin
}

func (holy *HolyPaladin) GetPaladin() *paladin.Paladin {
	return holy.Paladin
}

func (holy *HolyPaladin) Initialize() {
	holy.Paladin.Initialize()
}

func (holy *HolyPaladin) Reset(sim *core.Simulation) {
	holy.Paladin.Reset(sim)
}
