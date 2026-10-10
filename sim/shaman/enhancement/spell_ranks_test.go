package enhancement

import (
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanSpellRanks checks that a rotation naming a spell's top rank casts the highest
// rank our level knows.
//
// Our rotations are written at level 60, so they name rank 7 of Earth Shock, and a level 30
// shaman has to cast rank 4. The ranks we want are from the learn levels in the Forever client
// (70338, wago.tools SpellLevels), so this also checks the sim's learn levels. Lava Burst comes
// from a talent and has its own test (TestOrcShamanLavaBurstRanks).
func TestOrcShamanSpellRanks(t *testing.T) {
	cases := []struct {
		name       string
		ids        []int32
		at20, at30 int
	}{
		{"Lightning Bolt", shaman.LightningBoltSpellId[:], 4, 5},
		{"Chain Lightning", shaman.ChainLightningSpellId[:], 0, 0},
		{"Earth Shock", shaman.EarthShockSpellId[:], 3, 4},
		{"Flame Shock", shaman.FlameShockSpellId[:], 2, 3},
		{"Frost Shock", shaman.FrostShockSpellId[:], 1, 1},
		{"Lightning Shield", shaman.LightningShieldSpellId[:], 2, 3},
		{"Searing Totem", shaman.SearingTotemSpellId[:], 2, 3},
		{"Magma Totem", shaman.MagmaTotemSpellId[:], 0, 1},
		{"Fire Nova", shaman.FireNovaSpellId[:], 1, 2},
		{"Flametongue Totem", shaman.FlametongueTotemSpellId[:], 0, 1},
		{"Strength of Earth Totem", shaman.StrengthOfEarthTotemSpellId[:], 1, 2},
		{"Stoneskin Totem", shaman.StoneskinTotemSpellId[:], 2, 3},
		{"Windfury Totem", shaman.WindfuryTotemSpellId[:], 0, 0},
		{"Grace of Air Totem", shaman.GraceOfAirTotemSpellId[:], 0, 0},
		{"Windwall Totem", shaman.WindwallTotemSpellId[:], 0, 0},
		{"Healing Stream Totem", shaman.HealingStreamTotemSpellId[:], 1, 2},
		{"Mana Spring Totem", shaman.ManaSpringTotemSpellId[:], 0, 1},
	}
	for _, level := range []int32{20, 30, 60} {
		_, enh := newTargetLevelSim(level, level, whirlwind, 0, "")
		for _, c := range cases {
			want := len(c.ids) - 1
			switch level {
			case 20:
				want = c.at20
			case 30:
				want = c.at30
			}
			top := &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: c.ids[len(c.ids)-1]}}
			spell := enh.Rotation.GetAPLSpell(top)
			switch {
			case want == 0 && spell != nil:
				t.Errorf("level %d: %s cast rank %d, want none", level, c.name, spell.Rank)
			case want == 0:
			case spell == nil:
				t.Errorf("level %d: %s cast nothing, want rank %d", level, c.name, want)
			case spell.Rank != want || spell.ActionID.SpellID != c.ids[want]:
				t.Errorf("level %d: %s cast %d rank %d, want %d rank %d", level, c.name, spell.ActionID.SpellID, spell.Rank, c.ids[want], want)
			}
		}
	}
}
