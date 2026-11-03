// Package tank is the Feral tank: a druid in Bear Form, or Dire Bear Form from level 40, under
// Forever (sim/druid/bear.go, bear_attacks.go; docs/tank_sim.md, Rules audit: Feral tank).
package tank

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/druid"
)

func RegisterFeralTankDruid() {
	core.RegisterAgentFactory(
		proto.Player_FeralTankDruid{},
		proto.Spec_SpecFeralTankDruid,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewFeralTankDruid(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_FeralTankDruid)
			if !ok {
				panic("Invalid spec value for Feral Tank Druid!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewFeralTankDruid(character *core.Character, options *proto.Player) *FeralTankDruid {
	tankOptions := options.GetFeralTankDruid()
	bear := &FeralTankDruid{
		Druid:   druid.New(character, druid.Bear, druid.SelfBuffs{}, options.TalentsString),
		Options: tankOptions.GetOptions(),
	}
	if bear.Options == nil {
		bear.Options = &proto.FeralTankDruid_Options{}
	}
	bear.SelfBuffs.InnervateTarget = &proto.UnitReference{}
	if bear.Options.InnervateTarget != nil {
		bear.SelfBuffs.InnervateTarget = bear.Options.InnervateTarget
	}

	bear.EnableRageBar(core.RageBarOptions{
		StartingRage:          bear.Options.StartingRage,
		DamageDealtMultiplier: 1,
		DamageTakenMultiplier: 1,
	})
	// The cat's placeholder paw, so the flat damage of weapon enchants and stones can be taken off
	// it in Initialize (druid.KeepFlatWeaponDamage), when the Env is there for the bear's weapon.
	bear.EnableAutoAttacks(bear, core.AutoAttackOptions{
		MainHand:       bear.GetCatWeapon(),
		AutoSwingMelee: true,
		ReplaceMHSwing: bear.TryMaul,
	})
	bear.ReplaceBearMHFunc = bear.TryMaul
	bear.PseudoStats.FeralCombatEnabled = true
	return bear
}

type FeralTankDruid struct {
	*druid.Druid

	Options *proto.FeralTankDruid_Options
}

func (bear *FeralTankDruid) GetDruid() *druid.Druid {
	return bear.Druid
}

func (bear *FeralTankDruid) Initialize() {
	bear.Druid.Initialize()
	bear.KeepFlatWeaponDamage()
	bear.AutoAttacks.SetMH(bear.GetBearWeapon())
	bear.RegisterFeralTankSpells()
}

func (bear *FeralTankDruid) Reset(sim *core.Simulation) {
	bear.Druid.Reset(sim)
	bear.Druid.CancelShapeshift(sim)
	bear.BearFormAura.Activate(sim)
}
