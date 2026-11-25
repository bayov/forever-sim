package item_sets

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// EV-0144: the dungeon set procs whose rows state 100% and no rate fire at the procs per minute
// wowsims/forever measured, a cast priced on its cast time and at least 1.5 sec.
func TestForeverDungeonSetProcRates(t *testing.T) {
	for _, c := range []struct {
		name      string
		got, want float64
	}{
		{"Crusader's Wrath", CrusadersWrathPPM, 0.9},
		{"The Furious Storm", FuriousStormPPM, 1},
		{"Nature's Bounty", naturesBountyPPM, 1},
		{"Rogue Armor Energize", rogueArmorEnergizePPM, 0.8},
		{"Warrior's Resolve", warriorsResolvePPM, 1},
		{"instant at 1 PPM", foreverCastTimePPMChance(1, 0), 0.025},
		{"1 sec cast at 1 PPM", foreverCastTimePPMChance(1, time.Second), 0.025},
		{"3 sec cast at 0.9 PPM", foreverCastTimePPMChance(0.9, 3*time.Second), 0.045},
	} {
		if math.Abs(c.got-c.want) > 1e-9 {
			t.Errorf("%s: %v, want %v", c.name, c.got, c.want)
		}
	}
}

// EV-0144: a damaging or healing spell rolls, and so does any other cast that costs a resource (a
// seal, a totem); a melee or ranged ability never does, and a free cast never does.
func TestForeverCastRolls(t *testing.T) {
	cost := &core.SpellCost{BaseCost: 50}
	for _, c := range []struct {
		name  string
		spell *core.Spell
		want  bool
	}{
		{"damaging spell", &core.Spell{ProcMask: core.ProcMaskSpellDamage}, true},
		{"healing spell", &core.Spell{ProcMask: core.ProcMaskSpellHealing}, true},
		{"seal", &core.Spell{ProcMask: core.ProcMaskEmpty, Cost: cost}, true},
		{"free cast", &core.Spell{ProcMask: core.ProcMaskEmpty, Cost: &core.SpellCost{}}, false},
		{"no cost", &core.Spell{ProcMask: core.ProcMaskEmpty}, false},
		{"Holy Strike", &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial, Cost: cost}, false},
		{"ranged ability", &core.Spell{ProcMask: core.ProcMaskRangedSpecial, Cost: cost}, false},
	} {
		if got := foreverCastRolls(c.spell); got != c.want {
			t.Errorf("%s rolls: %v, want %v", c.name, got, c.want)
		}
	}
}

// EV-0167 and EV-0168: the Forever Scholomance 3-piece procs fire at the chance per landed hit
// wowsims/forever read off Warcraft Logs; Windfury Weapon's hits roll Bloodmail's bleed.
func TestForeverScholomanceSetChances(t *testing.T) {
	if BloodmailBleedChance != 0.107 || NecropileDrainChance != 0.139 {
		t.Errorf("Bloodmail %v, Necropile %v, want 0.107 and 0.139", BloodmailBleedChance, NecropileDrainChance)
	}
	if len(windfuryWeaponHits) != 4 || windfuryWeaponHits[3] != 16362 {
		t.Errorf("Windfury Weapon hits: %v", windfuryWeaponHits)
	}
}
