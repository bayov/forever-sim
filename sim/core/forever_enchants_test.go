//go:build with_db

package core

import (
	"testing"

	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's Enchanting 300 enchants and Rugged armor kits, from client 1.60.1.70338
// (SpellItemEnchantment), as the sim reads them from the database.
func TestForeverHighLevelEnchants(t *testing.T) {
	cases := []struct {
		effectID int32
		name     string
		want     stats.Stats
	}{
		{1219580, "2H Spellblasting", stats.Stats{stats.SpellPower: 65}},
		{1248607, "2H Mighty Spell Power", stats.Stats{stats.SpellPower: 55}},
		{1220624, "Bracer Greater Spellpower", stats.Stats{stats.SpellPower: 16}},
		{1248661, "Bracer Superior Intellect", stats.Stats{stats.Intellect: 9}},
		{1213622, "Gloves Holy Power", stats.Stats{stats.HolyPower: 20}},
		{1213626, "Gloves Arcane Power", stats.Stats{stats.ArcanePower: 20}},
		{1248645, "Gloves Natural Power", stats.Stats{stats.NaturePower: 20}},
		{1248640, "Gloves Superior Strength", stats.Stats{stats.Strength: 15}},
		{1219587, "Cloak Agility", stats.Stats{stats.Agility: 5}},
		{1219579, "Off-Hand Wisdom", stats.Stats{stats.Intellect: 6, stats.Spirit: 5}},
		{1254966, "Mystic Rugged Armor Kit", stats.Stats{stats.SpellPower: 6, stats.BonusArmor: 40}},
		{1254965, "Forceful Rugged Armor Kit", stats.Stats{stats.AttackPower: 10, stats.RangedAttackPower: 10, stats.BonusArmor: 40}},
		{1843, "Rugged Armor Kit", stats.Stats{stats.Stamina: 5, stats.BonusArmor: 40}},
	}
	for _, c := range cases {
		enchant, ok := EnchantsByEffectID[c.effectID]
		if !ok {
			t.Errorf("%s: no enchant with effect ID %d", c.name, c.effectID)
			continue
		}
		expectStats(t, c.name, enchant.Stats, c.want)
	}
}
