package enhancement

import (
	"fmt"
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanElementalFury checks which crits Elemental Fury makes bigger (shaman_audit.md
// 8.15).
//
// Each point adds 20% to the crit damage bonus, so at 5/5 a spell crit deals 2.0 times a hit
// instead of 1.5. The client's mask holds our Fire, Frost and Nature damage spells, Searing and
// Magma Totem and the Flametongue and Frostbrand Weapon procs. Flametongue Totem's procs and
// Healing Stream aren't in it, and Stormstrike is a melee attack.
func TestOrcShamanElementalFury(t *testing.T) {
	for _, points := range []int{0, 5} {
		spellCrit := 1.5 + .1*float64(points)
		// Elemental Fury points, Lava Burst and Stormstrike.
		talents := fmt.Sprintf("00000000000010%d-0000000000001", points)
		check := func(enh *EnhancementShaman, name string, spell *core.Spell, want float64) {
			if spell == nil {
				t.Fatalf("Elemental Fury %d: no %s", points, name)
			}
			table := enh.AttackTables[enh.CurrentTarget.Index][spell.CastType]
			if got := spell.CritMultiplier(table); math.Abs(got-want) > 1e-9 {
				t.Errorf("Elemental Fury %d: %s crits for %.2f times a hit, want %.2f", points, name, got, want)
			}
		}

		_, enh := weaponSim{level: 60, talents: talents, weapon: &proto.ItemSpec{Id: darkEdge}, imbue: proto.WeaponImbue_FlametongueWeapon, seconds: 10}.start()
		for _, s := range []struct {
			name  string
			spell *core.Spell
		}{
			{"Lightning Bolt", topRank(t, "Lightning Bolt", enh.LightningBolt)},
			{"Chain Lightning", topRank(t, "Chain Lightning", enh.ChainLightning)},
			{"Earth Shock", topRank(t, "Earth Shock", enh.EarthShock)},
			{"Flame Shock", topRank(t, "Flame Shock", enh.FlameShock)},
			{"Frost Shock", topRank(t, "Frost Shock", enh.FrostShock)},
			{"Lava Burst", enh.LavaBurst},
			{"Fire Nova", topRank(t, "Fire Nova", enh.FireNovaHits)},
			{"Searing Totem", lastSpell(enh, shaman.SearingTotemAttackSpellId[:], 0)},
			{"Magma Totem", enh.GetSpell(core.ActionID{SpellID: shaman.MagmaTotemAoeSpellId[4]})},
			{"Flametongue Weapon", lastSpell(enh, shaman.FlametongueWeaponSpellId[:], 0)},
		} {
			check(enh, s.name, s.spell, spellCrit)
		}
		check(enh, "Healing Stream", enh.GetSpell(core.ActionID{SpellID: 10461}), 1.5)
		meleeCrit := enh.Stormstrike.CritMultiplier(enh.AttackTables[enh.CurrentTarget.Index][enh.Stormstrike.CastType])
		if meleeCrit != 2 {
			t.Errorf("Elemental Fury %d: Stormstrike crits for %.2f times a hit, want 2", points, meleeCrit)
		}

		// Flametongue Weapon would keep Flametongue Totem off the weapon.
		_, enh = weaponSim{level: 60, talents: talents, weapon: &proto.ItemSpec{Id: darkEdge}, imbue: proto.WeaponImbue_FrostbrandWeapon, totem: proto.TotemWeaponBuff_TotemWeaponBuffFlametongue, seconds: 10}.start()
		check(enh, "Frostbrand Weapon", lastSpell(enh, shaman.FrostbrandWeaponSpellId[:], 0), spellCrit)
		check(enh, "Flametongue Totem", lastSpell(enh, core.FlametongueTotemProcSpellId[:], 0), 1.5)
	}
}
