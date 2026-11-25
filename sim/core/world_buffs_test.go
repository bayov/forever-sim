package core

import (
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

var worldBuffLabels = []string{
	"Rallying Cry of the Dragonslayer",
	"Warchief's Blessing",
	"Songflower Serenade",
	"Fengus' Ferocity",
	"Moldar's Moxie",
	"Slip'kik's Savvy",
	"Sayge's Dark Fortune of Damage",
}

func worldBuffsSim(ruleset proto.Ruleset, encounter *proto.Encounter) *Character {
	encounter.Targets = []*proto.Target{{Name: "target", Level: 63, MobType: proto.MobType_MobTypeDemon}}
	encounter.Duration = 180
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100, Ruleset: ruleset},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name:     "Caster",
				Class:    proto.Class_ClassShaman,
				Consumes: &proto.Consumes{},
				Buffs: &proto.IndividualBuffs{
					RallyingCryOfTheDragonslayer: true,
					WarchiefsBlessing:            true,
					SongflowerSerenade:           true,
					FengusFerocity:               true,
					MoldarsMoxie:                 true,
					SlipkiksSavvy:                true,
					SaygesFortune:                proto.SaygesFortune_SaygesDamage,
					SpiritOfZandalar:             true,
				},
				Spec:      &proto.Player_ElementalShaman{},
				Equipment: &proto.EquipmentSpec{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: encounter,
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Under Forever the world buffs whose auras read "No effect while in a raid instance" do nothing
// in a raid fight, and work outside one. Spirit of Zandalar has no such text and works in both.
// Classic keeps them all in a raid.
func TestWorldBuffsInRaidInstance(t *testing.T) {
	for _, c := range []struct {
		name      string
		ruleset   proto.Ruleset
		encounter *proto.Encounter
		want      bool
	}{
		{"Forever raid", proto.Ruleset_RulesetForever, &proto.Encounter{}, false},
		{"Forever outside a raid", proto.Ruleset_RulesetForever, &proto.Encounter{OutsideRaid: true}, true},
		{"Forever outdoors", proto.Ruleset_RulesetForever, &proto.Encounter{OutdoorTerrain: proto.OutdoorTerrain_OutdoorTerrainForest}, true},
		{"Forever PvP", proto.Ruleset_RulesetForever, &proto.Encounter{Pvp: true}, true},
		{"Classic raid", proto.Ruleset_RulesetClassic, &proto.Encounter{}, true},
	} {
		character := worldBuffsSim(c.ruleset, c.encounter)
		for _, label := range worldBuffLabels {
			if got := character.GetAura(label) != nil; got != c.want {
				t.Errorf("%s: %s registered %v, want %v", c.name, label, got, c.want)
			}
		}
		if character.GetAura("Spirit of Zandalar") == nil {
			t.Errorf("%s: Spirit of Zandalar missing", c.name)
		}
	}
}
