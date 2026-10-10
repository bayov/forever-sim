package core

import (
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Naked characters with no talents on the Forever beta (2026-10-10, /bayov export). Forever
// reworked the Horde races, so these rows check each race's shift against 1.12, the rows the
// sim builds for pairs 1.12 doesn't have (Undead paladin, Windshaper hunter), and the Orc
// shaman's level 30 row.
func TestForeverRaceBaseAttributes(t *testing.T) {
	cases := []struct {
		race  proto.Race
		class proto.Class
		level int32
		attrs [5]float64 // Strength, Agility, Stamina, Intellect, Spirit
	}{
		{proto.Race_RaceUndead, proto.Class_ClassPaladin, 1, [5]float64{24, 19, 23, 18, 21}},
		{proto.Race_RaceUndead, proto.Class_ClassPaladin, 20, [5]float64{44, 30, 42, 30, 34}},
		{proto.Race_RaceUndead, proto.Class_ClassRogue, 1, [5]float64{23, 22, 22, 18, 20}},
		{proto.Race_RaceUndead, proto.Class_ClassPriest, 1, [5]float64{22, 19, 21, 20, 23}},
		{proto.Race_RaceUndead, proto.Class_ClassMage, 1, [5]float64{22, 19, 21, 21, 22}},
		{proto.Race_RaceUndead, proto.Class_ClassWarrior, 1, [5]float64{25, 19, 23, 18, 20}},
		{proto.Race_RaceUndead, proto.Class_ClassWarlock, 1, [5]float64{22, 19, 22, 20, 22}},
		{proto.Race_RaceOrc, proto.Class_ClassShaman, 1, [5]float64{24, 17, 22, 20, 22}},
		{proto.Race_RaceOrc, proto.Class_ClassShaman, 30, [5]float64{51, 30, 51, 46, 53}},
		{proto.Race_RaceOrc, proto.Class_ClassWarlock, 1, [5]float64{23, 17, 22, 21, 22}},
		{proto.Race_RaceTauren, proto.Class_ClassHunter, 1, [5]float64{22, 21, 23, 18, 21}},
		{proto.Race_RaceTauren, proto.Class_ClassDruid, 1, [5]float64{23, 18, 22, 20, 22}},
		{proto.Race_RaceTroll, proto.Class_ClassPriest, 1, [5]float64{21, 22, 20, 19, 23}},
		{proto.Race_RaceSkyborneWindshaper, proto.Class_ClassHunter, 1, [5]float64{19, 24, 20, 21, 21}},
	}
	for _, c := range cases {
		base := getBaseStatsCombo(c.race, c.class, c.level)
		got := [5]float64{base[stats.Strength], base[stats.Agility], base[stats.Stamina], base[stats.Intellect], base[stats.Spirit]}
		if got != c.attrs {
			t.Errorf("level %d %s %s: got %v, the beta has %v", c.level, c.race, c.class, got, c.attrs)
		}
	}
}

// A level 1 Tauren druid on the beta had 2 x Strength - 20 attack power, with no gain per level.
func TestDruidBaseAttackPowerBelow60(t *testing.T) {
	if ap := getBaseStatsCombo(proto.Race_RaceTauren, proto.Class_ClassDruid, 1)[stats.AttackPower]; ap != -20 {
		t.Errorf("level 1 druid base attack power: got %v, want -20", ap)
	}
}
