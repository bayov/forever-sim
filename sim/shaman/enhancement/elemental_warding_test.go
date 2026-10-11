package enhancement

import (
	"fmt"
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanElementalWarding checks how much Elemental Warding takes off the damage we take
// (shaman_audit.md 8.3).
//
// The client gives 3, 7 and 10% less damage taken from Fire, Frost and Nature for 1, 2 and 3
// points, and nothing off the other schools. Our presets never meet an enemy that casts at us,
// so this only matters when one does. Each enemy spell hits us for 100 and ignores our
// resistances and armor, so what we take is 100 times the talent's multiplier.
func TestOrcShamanElementalWarding(t *testing.T) {
	schools := []struct {
		name   string
		school core.SpellSchool
		warded bool
	}{
		{"Fire", core.SpellSchoolFire, true},
		{"Frost", core.SpellSchoolFrost, true},
		{"Nature", core.SpellSchoolNature, true},
		{"Arcane", core.SpellSchoolArcane, false},
		{"Shadow", core.SpellSchoolShadow, false},
		{"Physical", core.SpellSchoolPhysical, false},
	}
	for points, taken := range []float64{1, .97, .93, .90} {
		// Convection 0, Concussion 0, Elemental Warding points.
		sim, enh := newShamanSim(newOrcShaman(30, fmt.Sprintf("00%d", points), &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 10)
		if enh.Talents.ElementalWarding != int32(points) {
			t.Fatalf("got Elemental Warding %d, want %d", enh.Talents.ElementalWarding, points)
		}
		for i, s := range schools {
			spell := enh.CurrentTarget.RegisterSpell(core.SpellConfig{
				ActionID:         core.ActionID{SpellID: int32(i + 1)},
				SpellSchool:      s.school,
				DefenseType:      core.DefenseTypeMagic,
				ProcMask:         core.ProcMaskSpellDamage,
				Flags:            core.SpellFlagIgnoreResists,
				DamageMultiplier: 1,
			})
			want := 100.0
			if s.warded {
				want *= taken
			}
			if got := spell.CalcDamage(sim, &enh.Unit, 100, forcedOutcome(core.OutcomeHit)).Damage; math.Abs(got-want) > 1e-9 {
				t.Errorf("Elemental Warding %d: a 100 damage %s hit dealt %.4f, want %.4f", points, s.name, got, want)
			}
		}
	}
}
