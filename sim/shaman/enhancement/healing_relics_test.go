package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// withRelic puts the relic in the shaman's ranged slot.
func withRelic(player *proto.Player, relic int32) *proto.Player {
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotRanged].Id = relic
	player.Equipment = &proto.EquipmentSpec{Items: items}
	return player
}

// TestOrcShamanHealingRelics: Tidal Totem takes 5% off Healing Wave's cost and nothing else's, and
// Totem of Urgency gives Lesser Healing Wave 4% more crit and nothing else (client 70753, 1291076
// and 1306448).
func TestOrcShamanHealingRelics(t *testing.T) {
	relicSim := func(relic int32) *EnhancementShaman {
		player := newOrcShaman(60, restoTidalFocus, &proto.EnhancementShaman_Options{})
		if relic != 0 {
			player = withRelic(player, relic)
		}
		_, enh := newShamanSim(player, &proto.Debuffs{}, 10)
		return enh
	}
	plain, tidal, urgency := relicSim(0), relicSim(shaman.TidalTotem), relicSim(shaman.TotemOfUrgency)

	for _, c := range []struct {
		name         string
		spell        func(*EnhancementShaman) *core.Spell
		tidalPercent float64
	}{
		{"Healing Wave 10", func(e *EnhancementShaman) *core.Spell { return e.HealingWave[10] }, 0.95},
		{"Healing Wave 6", func(e *EnhancementShaman) *core.Spell { return e.HealingWave[6] }, 0.95},
		{"Lesser Healing Wave 6", func(e *EnhancementShaman) *core.Spell { return e.LesserHealingWave[6] }, 1},
		{"Chain Heal 3", func(e *EnhancementShaman) *core.Spell { return e.ChainHeal[3] }, 1},
	} {
		base := c.spell(plain).Cost.GetCurrentCost()
		if got, want := c.spell(tidal).Cost.GetCurrentCost(), base*c.tidalPercent; math.Abs(got-want) > 0.01 {
			t.Errorf("Tidal Totem: %s costs %.2f, want %.2f", c.name, got, want)
		}
		wantCrit := 0.0
		if c.name == "Lesser Healing Wave 6" {
			wantCrit = 0.04
		}
		if got := c.spell(urgency).HealingCritChance() - c.spell(plain).HealingCritChance(); math.Abs(got-wantCrit) > 1e-9 {
			t.Errorf("Totem of Urgency: %s gets %.3f more crit, want %.3f", c.name, got, wantCrit)
		}
	}
}
