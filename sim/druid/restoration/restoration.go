// Package restoration is the Restoration druid, a healer of the healing sim (docs/healing_druid.md,
// docs/healing_sim.md H3), in place of the old wowsims package (_restoration, never built). Its heals
// are the druid's (sim/druid/healing.go); this agent is a caster in caster form that doesn't melee and
// heals what its rotation aims at.
package restoration

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/druid"
)

func RegisterRestorationDruid() {
	core.RegisterAgentFactory(
		proto.Player_RestorationDruid{},
		proto.Spec_SpecRestorationDruid,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewRestorationDruid(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_RestorationDruid)
			if !ok {
				panic("Invalid spec value for Restoration Druid!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewRestorationDruid(character *core.Character, options *proto.Player) *RestorationDruid {
	restoOptions := options.GetRestorationDruid().GetOptions()

	resto := &RestorationDruid{
		Druid: druid.New(character, druid.Humanoid, druid.SelfBuffs{}, options.TalentsString),
	}

	// Innervate goes on ourselves unless the options name someone else.
	resto.SelfBuffs.InnervateTarget = &proto.UnitReference{Type: proto.UnitReference_Self}
	if target := restoOptions.GetInnervateTarget(); target != nil && target.Type != proto.UnitReference_Unknown {
		resto.SelfBuffs.InnervateTarget = target
	}

	// A healer stands back and casts: no auto attacks.

	return resto
}

type RestorationDruid struct {
	*druid.Druid
}

func (resto *RestorationDruid) GetDruid() *druid.Druid {
	return resto.Druid
}

func (resto *RestorationDruid) ApplyTalents() {
	resto.Druid.ApplyTalents()
	// Natural Instinct's healing part (1223242, its second effect, aura 175): 12 or 25% of our
	// Intellect as spell healing. Its crit damage part is the cat's (sim/druid/talents.go).
	if points := resto.Talents.NaturalInstinct; points > 0 {
		resto.AddStatDependency(stats.Intellect, stats.HealingPower, []float64{0, 0.12, 0.25}[points])
	}
}

func (resto *RestorationDruid) Initialize() {
	resto.Druid.Initialize()
	resto.RegisterHealingSpells()
}

func (resto *RestorationDruid) Reset(sim *core.Simulation) {
	resto.Druid.Reset(sim)
}
