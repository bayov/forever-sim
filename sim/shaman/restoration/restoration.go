// Package restoration is the Restoration shaman, the healing sim's first healer
// (docs/healing_sim.md, H2). Its heals, Water Shield and water totems are the shaman's
// (sim/shaman); this agent is a caster that doesn't melee and heals what its rotation aims at.
package restoration

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

func RegisterRestorationShaman() {
	core.RegisterAgentFactory(
		proto.Player_RestorationShaman{},
		proto.Spec_SpecRestorationShaman,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewRestorationShaman(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_RestorationShaman)
			if !ok {
				panic("Invalid spec value for Restoration Shaman!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewRestorationShaman(character *core.Character, options *proto.Player) *RestorationShaman {
	restoOptions := options.GetRestorationShaman()

	shaman.OneMainHandImbue(character, restoOptions.GetOptions().GetShamanImbue())
	resto := &RestorationShaman{
		Shaman: shaman.NewShaman(character, options.TalentsString),
	}
	resto.AssignedTotems = options.AssignedTotems

	// A healer stands back and casts: no auto attacks.

	resto.ShamanImbue = restoOptions.GetOptions().GetShamanImbue()
	resto.ApplyShamanImbue()

	return resto
}

type RestorationShaman struct {
	*shaman.Shaman
}

func (resto *RestorationShaman) GetShaman() *shaman.Shaman {
	return resto.Shaman
}

func (resto *RestorationShaman) Reset(sim *core.Simulation) {
	resto.Shaman.Reset(sim)
}
