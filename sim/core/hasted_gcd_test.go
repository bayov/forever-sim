package core

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

func hasteGCDSpell(haste float64, school SpellSchool, defense DefenseType) *Spell {
	return &Spell{
		Unit:        &Unit{CastSpeed: 1 / haste},
		SpellSchool: school,
		DefenseType: defense,
	}
}

// A spell's GCD under Forever is 1.5 / (1 + spell haste), down to 1 sec (cast.go hastedGCD).
func TestHastedGCD(t *testing.T) {
	ms := time.Millisecond
	cases := []struct {
		name    string
		haste   float64
		school  SpellSchool
		defense DefenseType
		gcd     time.Duration
		want    time.Duration
	}{
		{"no haste", 1, SpellSchoolNature, DefenseTypeMagic, GCDDefault, GCDDefault},
		{"Warchief's Blessing", 1.15, SpellSchoolNature, DefenseTypeMagic, GCDDefault, 1304 * ms},
		{"seal (no defense type, Holy)", 1.15, SpellSchoolHoly, DefenseTypeNone, GCDDefault, 1304 * ms},
		{"floor", 2, SpellSchoolFire, DefenseTypeMagic, GCDDefault, GCDMin},
		{"totem's 1s", 1.15, SpellSchoolFire, DefenseTypeNone, time.Second, time.Second},
		{"under 1s floors at 1s", 1.15, SpellSchoolFire, DefenseTypeMagic, 500 * ms, GCDMin},
		{"Nature's Grace with its own cut", 1.15 * 1.1, SpellSchoolNature, DefenseTypeMagic, GCDDefault - GCDDefault/10, 1067 * ms},
		{"melee ability", 1.15, SpellSchoolNature, DefenseTypeMelee, GCDDefault, GCDDefault},
		{"physical school", 1.15, SpellSchoolPhysical, DefenseTypeMelee, GCDDefault, GCDDefault},
		{"ranged shot, magic school", 1.15, SpellSchoolArcane, DefenseTypeRanged, GCDDefault, GCDDefault},
		{"rogue and cat 1s", 1.15, SpellSchoolPhysical, DefenseTypeNone, time.Second, time.Second},
		{"shout (no school)", 1.15, SpellSchoolNone, DefenseTypeNone, GCDDefault, GCDDefault},
		{"no GCD", 1.15, SpellSchoolNature, DefenseTypeMagic, 0, 0},
	}
	for _, c := range cases {
		got := hasteGCDSpell(c.haste, c.school, c.defense).hastedGCD(c.gcd)
		if got != c.want {
			t.Errorf("%s: GCD %v, want %v", c.name, got, c.want)
		}
	}
}

func warchiefsSim(ruleset proto.Ruleset) *Simulation {
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100, Ruleset: ruleset},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name:      "Caster",
				Class:     proto.Class_ClassShaman,
				Consumes:  &proto.Consumes{},
				Buffs:     &proto.IndividualBuffs{WarchiefsBlessing: true},
				Spec:      &proto.Player_ElementalShaman{},
				Equipment: &proto.EquipmentSpec{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		// Outside a raid instance, where Forever's world buffs work (ForeverRaidInstance).
		Encounter: &proto.Encounter{
			Targets:     []*proto.Target{{Name: "target", Level: 63, MobType: proto.MobType_MobTypeDemon}},
			Duration:    180,
			OutsideRaid: true,
		},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim
}

// Warchief's Blessing hastes casts 15% under Forever and melee only under Classic.
func TestWarchiefsBlessingCastSpeed(t *testing.T) {
	for _, c := range []struct {
		ruleset proto.Ruleset
		want    float64
	}{{proto.Ruleset_RulesetForever, 1 / 1.15}, {proto.Ruleset_RulesetClassic, 1}} {
		sim := warchiefsSim(c.ruleset)
		unit := &sim.Raid.Parties[0].Players[0].GetCharacter().Unit
		if math.Abs(unit.CastSpeed-c.want) > 1e-9 {
			t.Errorf("%v: cast speed %v, want %v", c.ruleset, unit.CastSpeed, c.want)
		}
		if c.ruleset == proto.Ruleset_RulesetForever {
			spell := hasteGCDSpell(1, SpellSchoolNature, DefenseTypeMagic)
			spell.Unit = unit
			if got := spell.hastedGCD(GCDDefault); got != 1304*time.Millisecond {
				t.Errorf("GCD with Warchief's Blessing %v, want 1.304s", got)
			}
		}
	}
}
