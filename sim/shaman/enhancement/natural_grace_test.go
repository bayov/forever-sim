package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanNaturalGrace checks which threat Natural Grace cuts (shaman_audit.md 8.38).
//
// 3/3 cuts the threat of everything we cast in a magic school by 15%, as the client has it
// (29187, every school but physical). Stormstrike and white hits are physical and keep their
// threat. We compare each spell's threat multiplier at level 60 with and without the talent.
func TestOrcShamanNaturalGrace(t *testing.T) {
	start := func(talents string) *EnhancementShaman {
		_, enh := weaponSim{level: 60, talents: talents, weapon: &proto.ItemSpec{Id: darkEdge}, imbue: proto.WeaponImbue_FlametongueWeapon, seconds: 10}.start()
		return enh
	}
	// Stormstrike, and Natural Grace 3/3.
	plain, grace := start("-0000000000001"), start("-0000000000001-0003")
	spells := func(enh *EnhancementShaman) []*core.Spell {
		return []*core.Spell{
			topRank(t, "Lightning Bolt", enh.LightningBolt),
			topRank(t, "Flame Shock", enh.FlameShock),
			lastSpell(enh, shaman.FlametongueWeaponSpellId[:], 0),
			enh.GetSpell(core.ActionID{SpellID: 10461}),
			enh.Stormstrike,
			enh.AutoAttacks.MHAuto(),
		}
	}
	names := []string{"Lightning Bolt", "Flame Shock", "Flametongue Weapon", "Healing Stream Totem", "Stormstrike", "white hits"}
	wants := []float64{.85, .85, .85, .85, 1, 1}
	before, after := spells(plain), spells(grace)
	for i, spell := range after {
		if spell == nil || before[i] == nil {
			t.Fatalf("no %s", names[i])
		}
		if got := spell.ThreatMultiplier / before[i].ThreatMultiplier; math.Abs(got-wants[i]) > 1e-9 {
			t.Errorf("Natural Grace 3/3: %s has %.2f times the threat, want %.2f", names[i], got, wants[i])
		}
	}
}
