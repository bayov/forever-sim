package core

import (
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's races have Classic's base attributes since client 1.60.1.70753 ("All Races now have
// the correct amount of base stats. For example, Tauren Warriors now start with the correct +5
// strength relative to Human Warriors (was incorrectly +2)", development notes 2026-11-02). Below
// 60 that is 1.12's rows (player_levelstats), and the rows the sim builds for pairs 1.12 doesn't
// have (Undead paladin, Skyborne) and the Orc shaman's level 30 sheet follow from them.
func TestForeverRaceBaseAttributes(t *testing.T) {
	cases := []struct {
		race  proto.Race
		class proto.Class
		level int32
		attrs [5]float64 // Strength, Agility, Stamina, Intellect, Spirit
	}{
		// 1.12's level 1 rows.
		{proto.Race_RaceHuman, proto.Class_ClassWarrior, 1, [5]float64{23, 20, 22, 20, 21}},
		{proto.Race_RaceTauren, proto.Class_ClassWarrior, 1, [5]float64{28, 15, 24, 15, 22}},
		{proto.Race_RaceUndead, proto.Class_ClassWarrior, 1, [5]float64{22, 18, 23, 18, 25}},
		{proto.Race_RaceOrc, proto.Class_ClassWarlock, 1, [5]float64{23, 17, 23, 19, 25}},
		{proto.Race_RaceTroll, proto.Class_ClassPriest, 1, [5]float64{21, 22, 21, 18, 24}},
		// The Human paladin's row and the Undead's warrior gap to the Human.
		{proto.Race_RaceUndead, proto.Class_ClassPaladin, 1, [5]float64{21, 18, 23, 18, 26}},
		// The Orc shaman's level 30 sheet without the old Orc shift.
		{proto.Race_RaceOrc, proto.Class_ClassShaman, 30, [5]float64{51, 30, 52, 44, 56}},
		// The Skyborne keep their sheets' offset: the Human's row less 1 Strength and Stamina and 1 more
		// Agility and Intellect, with no Human Spirit.
		{proto.Race_RaceSkyborneHighOrder, proto.Class_ClassWarrior, 1, [5]float64{22, 21, 21, 21, 20}},
		{proto.Race_RaceSkyborneWindshaper, proto.Class_ClassHunter, 1, [5]float64{19, 24, 20, 21, 21}}, // the 2026-09-22 sheet
		// Level 60: the class baseline plus Classic's offset. These are 1.12's level 60 warrior rows but for
		// the Human's Spirit (47 there, The Human Spirit baked in) and the Gnome's Intellect (35, Expansive
		// Mind baked in).
		{proto.Race_RaceHuman, proto.Class_ClassWarrior, 60, [5]float64{120, 80, 110, 30, 45}},
		{proto.Race_RaceOrc, proto.Class_ClassWarrior, 60, [5]float64{123, 77, 112, 27, 48}},
		{proto.Race_RaceDwarf, proto.Class_ClassWarrior, 60, [5]float64{122, 76, 113, 29, 44}},
		{proto.Race_RaceNightElf, proto.Class_ClassWarrior, 60, [5]float64{117, 85, 109, 30, 45}},
		{proto.Race_RaceUndead, proto.Class_ClassWarrior, 60, [5]float64{119, 78, 111, 28, 50}},
		{proto.Race_RaceTauren, proto.Class_ClassWarrior, 60, [5]float64{125, 75, 112, 25, 47}},
		{proto.Race_RaceGnome, proto.Class_ClassWarrior, 60, [5]float64{115, 83, 109, 33, 45}},
		{proto.Race_RaceTroll, proto.Class_ClassWarrior, 60, [5]float64{121, 82, 111, 26, 46}},
		{proto.Race_RaceSkyborneHighOrder, proto.Class_ClassWarrior, 60, [5]float64{119, 81, 109, 31, 45}},
		{proto.Race_RaceTauren, proto.Class_ClassDruid, 60, [5]float64{70, 55, 72, 95, 112}},
		{proto.Race_RaceOrc, proto.Class_ClassShaman, 60, [5]float64{88, 52, 97, 87, 103}},
	}
	for _, c := range cases {
		base := getBaseStatsCombo(c.race, c.class, c.level)
		got := [5]float64{base[stats.Strength], base[stats.Agility], base[stats.Stamina], base[stats.Intellect], base[stats.Spirit]}
		if got != c.attrs {
			t.Errorf("level %d %s %s: got %v, want %v", c.level, c.race, c.class, got, c.attrs)
		}
	}
}

// A level 1 Tauren druid on the beta had 2 x Strength - 20 attack power, with no gain per level.
func TestDruidBaseAttackPowerBelow60(t *testing.T) {
	if ap := getBaseStatsCombo(proto.Race_RaceTauren, proto.Class_ClassDruid, 1)[stats.AttackPower]; ap != -20 {
		t.Errorf("level 1 druid base attack power: got %v, want -20", ap)
	}
}
