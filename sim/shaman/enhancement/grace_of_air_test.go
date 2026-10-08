package enhancement

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// newGraceOfAirSim starts a level 60 Forever sim of an Orc shaman with a two-hander and
// another shaman's Grace of Air.
func newGraceOfAirSim(totemWeaponBuff proto.TotemWeaponBuff, options *proto.EnhancementShaman_Options) (*core.Simulation, *EnhancementShaman) {
	player := newOrcShaman(60, "", options)
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for slot := range items {
		items[slot] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 12784} // Arcanite Reaper
	player.Equipment = &proto.EquipmentSpec{Items: items}
	raidBuffs := &proto.RaidBuffs{GraceOfAirTotem: proto.TristateEffect_TristateEffectRegular, TotemWeaponBuff: totemWeaponBuff}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, raidBuffs, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 30, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
}

// TestOrcShamanGraceOfAirWithWindfury checks that Grace of Air gives no Agility while a
// Windfury Totem buff is up.
//
// Forever doesn't stack the two for now (the user, 2026-10-08). The Grace of Air here is
// another shaman's. When the Windfury Totem is another shaman's too, Grace of Air stays off
// the whole fight. When it's our own starting totem, put down 290 sec before the pull, it runs
// out at 10 sec and Grace of Air's 89 Agility comes back.
func TestOrcShamanGraceOfAirWithWindfury(t *testing.T) {
	_, alone := newGraceOfAirSim(proto.TotemWeaponBuff_TotemWeaponBuffNone, &proto.EnhancementShaman_Options{})
	agility := alone.GetStat(stats.Agility)

	_, raidWindfury := newGraceOfAirSim(proto.TotemWeaponBuff_TotemWeaponBuffWindfury, &proto.EnhancementShaman_Options{})
	if got, want := raidWindfury.GetStat(stats.Agility), agility-89; got != want {
		t.Errorf("with another shaman's Windfury Totem: %.0f Agility, want %.0f", got, want)
	}

	sim, enh := newGraceOfAirSim(proto.TotemWeaponBuff_TotemWeaponBuffNone, &proto.EnhancementShaman_Options{StartingTotems: &proto.StartingTotems{
		Air:                  proto.AirTotem_WindfuryTotem,
		AirSecondsBeforePull: 290,
	}})
	at(sim, 5, func(sim *core.Simulation) {
		if got, want := enh.GetStat(stats.Agility), agility-89; got != want {
			t.Errorf("at 5 sec with our Windfury Totem: %.0f Agility, want %.0f", got, want)
		}
	})
	at(sim, 10.001, func(sim *core.Simulation) {
		if got := enh.GetStat(stats.Agility); got != agility {
			t.Errorf("after our Windfury Totem ran out: %.0f Agility, want %.0f", got, agility)
		}
	})
	runSim(sim)
}
