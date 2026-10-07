package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanManaConsumables checks the mana each potion, Demonic Rune and Minor
// Recombobulator restores, and their cooldowns, against wowhead Forever (2026-10-08).
//
// The potions and Demonic Rune have the Classic values. Minor Recombobulator gives 96 to 160
// mana under Forever, where Classic gives 150 to 250, and it has a 5 min cooldown. We use
// each one 300 times and check that every roll lands in the range and that the rolls reach
// both ends of it.
//
// The rotation uses one once the missing mana is at least its top roll plus 2 sec of regen
// while casting, so no roll goes to waste. A naked shaman has no regen while casting.
func TestOrcShamanManaConsumables(t *testing.T) {
	cases := []struct {
		name     string
		itemID   int32
		consumes *proto.Consumes
		min, max float64
		cooldown time.Duration
	}{
		{"Lesser Mana Potion", 3385, &proto.Consumes{DefaultPotion: proto.Potions_LesserManaPotion}, 280, 360, 2 * time.Minute},
		{"Mana Potion", 3827, &proto.Consumes{DefaultPotion: proto.Potions_ManaPotion}, 455, 585, 2 * time.Minute},
		{"Greater Mana Potion", 6149, &proto.Consumes{DefaultPotion: proto.Potions_GreaterManaPotion}, 700, 900, 2 * time.Minute},
		{"Superior Mana Potion", 13443, &proto.Consumes{DefaultPotion: proto.Potions_SuperiorManaPotion}, 900, 1500, 2 * time.Minute},
		{"Major Mana Potion", 13444, &proto.Consumes{DefaultPotion: proto.Potions_MajorManaPotion}, 1350, 2250, 2 * time.Minute},
		{"Demonic Rune", 12662, &proto.Consumes{DefaultConjured: proto.Conjured_ConjuredDemonicRune}, 900, 1500, 2 * time.Minute},
		{"Minor Recombobulator", 4381, &proto.Consumes{DefaultConjured: proto.Conjured_ConjuredMinorRecombobulator}, 96, 160, 5 * time.Minute},
	}
	for _, c := range cases {
		player := newOrcShaman(60, "", &proto.EnhancementShaman_Options{})
		player.Consumes = c.consumes
		// Enough mana that a Major Mana Potion's top roll fits.
		player.BonusStats = &proto.UnitStats{Stats: stats.Stats{stats.Intellect: 200}.ToFloatArray()}
		sim, enh := newShamanSim(player, &proto.Debuffs{}, 30)

		mcd := enh.GetMajorCooldown(core.ActionID{ItemID: c.itemID})
		if mcd == nil {
			t.Fatalf("%s: no cooldown registered", c.name)
		}
		if got := mcd.Spell.CD.Duration; got != c.cooldown {
			t.Errorf("%s: got a %s cooldown, want %s", c.name, got, c.cooldown)
		}

		at(sim, 1, func(sim *core.Simulation) {
			enh.SpendMana(sim, c.max-1, enh.GetManaNotCastingMetrics())
			if mcd.ShouldActivate(sim, enh.GetCharacter()) {
				t.Errorf("%s: used with %.0f mana missing", c.name, c.max-1)
			}
			enh.SpendMana(sim, 1, enh.GetManaNotCastingMetrics())
			if !mcd.ShouldActivate(sim, enh.GetCharacter()) {
				t.Errorf("%s: not used with %.0f mana missing", c.name, c.max)
			}

			lowest, highest := c.max, c.min
			for range 300 {
				enh.SpendMana(sim, enh.CurrentMana(), enh.GetManaNotCastingMetrics())
				mcd.Spell.CD.Reset()
				if !mcd.Spell.Cast(sim, &enh.Unit) {
					t.Fatalf("%s didn't cast", c.name)
				}
				gain := enh.CurrentMana()
				lowest, highest = min(lowest, gain), max(highest, gain)
			}
			if lowest < c.min || highest > c.max {
				t.Errorf("%s: rolled %.0f to %.0f, want %.0f to %.0f", c.name, lowest, highest, c.min, c.max)
			}
			if margin := (c.max - c.min) * 0.05; lowest > c.min+margin || highest < c.max-margin {
				t.Errorf("%s: rolled only %.0f to %.0f of %.0f to %.0f", c.name, lowest, highest, c.min, c.max)
			}
		})
		runSim(sim)
	}
}
