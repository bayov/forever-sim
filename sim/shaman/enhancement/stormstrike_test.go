package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanStormstrikeOutcomes checks what a Stormstrike costs and leaves behind on each
// outcome.
//
// It always costs 125 mana and starts its 8 sec cooldown, and only a Stormstrike that lands
// puts the mark up. A blocked one lands. The beta log showed all of it (shaman_audit.md 5.8): a
// parried Stormstrike took 125 mana, and the 8 misses, 3 dodges and 19 parries put no mark up
// where the one block did. We cast 400 from in front of a level 63 boss, so every outcome
// comes up.
func TestOrcShamanStormstrikeOutcomes(t *testing.T) {
	player := newOrcShaman(60, stormstrike, &proto.EnhancementShaman_Options{})
	player.InFrontOfTarget = true
	for range proto.ItemSlot_ItemSlotMainHand {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
	}
	player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: darkEdge})
	sim, enh := newShamanSim(player, &proto.Debuffs{}, 10)
	mark := enh.CurrentTarget.GetAuraByID(enh.Stormstrike.ActionID)
	metrics := &enh.Stormstrike.SpellMetrics[enh.CurrentTarget.UnitIndex]

	outcome := func(before core.SpellMetrics) string {
		switch {
		case metrics.Misses > before.Misses:
			return "miss"
		case metrics.Dodges > before.Dodges:
			return "dodge"
		case metrics.Parries > before.Parries:
			return "parry"
		case metrics.Blocks > before.Blocks:
			return "block"
		case metrics.Crits > before.Crits:
			return "crit"
		case metrics.Hits > before.Hits:
			return "hit"
		}
		return "nothing"
	}
	landed := map[string]bool{"miss": false, "dodge": false, "parry": false, "block": true, "crit": true, "hit": true}

	seen := map[string]int{}
	at(sim, 1, func(sim *core.Simulation) {
		for range 400 {
			mark.Deactivate(sim)
			enh.Stormstrike.CD.Reset()
			enh.AddMana(sim, enh.MaxMana()-enh.CurrentMana(), enh.GetManaNotCastingMetrics())

			before, mana := *metrics, enh.CurrentMana()
			castNow(t, sim, enh, enh.Stormstrike)
			o := outcome(before)
			seen[o]++

			if spent := mana - enh.CurrentMana(); spent != 125 {
				t.Fatalf("a %s cost %.1f mana, want 125", o, spent)
			}
			if readyAt := enh.Stormstrike.CD.ReadyAt(); readyAt != sim.CurrentTime+8*time.Second {
				t.Fatalf("after a %s the cooldown ends at %s, want %s", o, readyAt, sim.CurrentTime+8*time.Second)
			}
			if mark.IsActive() != landed[o] {
				t.Fatalf("a %s leaves the mark up: %t, want %t", o, mark.IsActive(), landed[o])
			}
		}
	})
	runSim(sim)

	for o := range landed {
		if seen[o] == 0 {
			t.Errorf("no %s in 400 Stormstrikes (%v)", o, seen)
		}
	}
}

// TestOrcShamanStormstrikeMark checks what uses up the Stormstrike mark and how long it lasts
// (shaman_audit.md 7.7).
//
// Our next Earth Shock or Lightning Bolt on the marked target uses it up and deals 20% more,
// even when it misses. So the Earth Shock after it gets nothing. Flame Shock, Frost Shock, a
// Flametongue Weapon hit and a Lightning Shield orb leave it up and get nothing. On the beta
// only an Earth Shock, a death or the 12 sec running out took the mark off, and all 12 missed
// Earth Shocks took it.
//
// We cast the same spells in two sims with the same seed, and only one of them puts the mark
// up. The random rolls match, so each cast that uses the mark deals exactly 1.2 times what it
// deals in the other sim, and every other hit deals the same. The level 63 target makes 17%
// of our spells miss.
func TestOrcShamanStormstrikeMark(t *testing.T) {
	type hit struct {
		name   string
		damage float64
	}
	run := func(marked bool) []hit {
		player := newOrcShaman(60, stormstrike, &proto.EnhancementShaman_Options{ShamanImbue: proto.WeaponImbue_FlametongueWeapon})
		for range proto.ItemSlot_ItemSlotMainHand {
			player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
		}
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: darkEdge})
		sim, enh := newShamanSim(player, &proto.Debuffs{}, 25)
		mark := enh.CurrentTarget.GetAuraByID(enh.Stormstrike.ActionID)
		earthShock := topRank(t, "Earth Shock", enh.EarthShock)
		lightningBolt := topRank(t, "Lightning Bolt", enh.LightningBolt)
		shield := topRank(t, "Lightning Shield", enh.LightningShield)
		others := []struct {
			name  string
			spell *core.Spell
		}{
			{"Flame Shock", topRank(t, "Flame Shock", enh.FlameShock)},
			{"Frost Shock", topRank(t, "Frost Shock", enh.FrostShock)},
			{"Flametongue Weapon", lastSpell(enh, shaman.FlametongueWeaponSpellId[:], 0)},
			{"Lightning Shield orb", topRank(t, "Lightning Shield orb", enh.LightningShieldProcs)},
		}
		names := map[*core.Spell]string{lightningBolt: "Lightning Bolt"}
		for _, o := range others {
			names[o.spell] = o.name
		}

		// We take every hit, and the white swings' Flametongue procs too.
		var hits []hit
		imbue := enh.GetAura("Flametongue Imbue")
		onHit := imbue.OnSpellHitDealt
		imbue.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			onHit(aura, sim, spell, result)
			if name, ok := names[spell]; ok {
				hits = append(hits, hit{name, result.Damage})
			}
		}

		misses := 0
		at(sim, 1, func(sim *core.Simulation) {
			for range 100 {
				if marked {
					mark.Activate(sim)
				}
				shield.ApplyEffects(sim, enh.CurrentTarget, shield)
				for _, o := range others {
					o.spell.ApplyEffects(sim, enh.CurrentTarget, o.spell)
					if marked && !mark.IsActive() {
						t.Fatalf("%s used up the mark", o.name)
					}
				}

				names[earthShock] = "Earth Shock on the mark"
				before := earthShock.SpellMetrics[enh.CurrentTarget.UnitIndex].Misses
				earthShock.ApplyEffects(sim, enh.CurrentTarget, earthShock)
				if earthShock.SpellMetrics[enh.CurrentTarget.UnitIndex].Misses > before {
					misses++
				}
				if mark.IsActive() {
					t.Fatalf("Earth Shock left the mark up")
				}
				names[earthShock] = "Earth Shock after it"
				earthShock.ApplyEffects(sim, enh.CurrentTarget, earthShock)

				if marked {
					mark.Activate(sim)
				}
				lightningBolt.ApplyEffects(sim, enh.CurrentTarget, lightningBolt)
				if mark.IsActive() {
					t.Fatalf("Lightning Bolt left the mark up")
				}
			}
		})
		if marked {
			// The mark lasts 12 sec from the last time it went up. Flame Shock ticks on it
			// meanwhile.
			at(sim, 2, func(sim *core.Simulation) { mark.Activate(sim) })
			at(sim, 8, func(sim *core.Simulation) { mark.Activate(sim) })
			at(sim, 19.99, func(sim *core.Simulation) {
				if !mark.IsActive() {
					t.Errorf("the mark is gone 11.99 sec after it went up")
				}
			})
			at(sim, 20.01, func(sim *core.Simulation) {
				if mark.IsActive() {
					t.Errorf("the mark is still up 12.01 sec after it went up")
				}
			})
		}
		runSim(sim)

		if misses == 0 {
			t.Errorf("none of 100 marked Earth Shocks missed")
		}
		return hits
	}

	marked, plain := run(true), run(false)
	if len(marked) != len(plain) {
		t.Fatalf("%d hits with the mark and %d without", len(marked), len(plain))
	}
	bonus := map[string]float64{"Earth Shock on the mark": 1.2, "Lightning Bolt": 1.2}
	for i := range marked {
		m, p := marked[i], plain[i]
		want := p.damage * max(bonus[m.name], 1)
		if m.name != p.name || math.Abs(m.damage-want) > 1e-6 {
			t.Fatalf("hit %d: %s for %.2f with the mark, want %s for %.2f", i, m.name, m.damage, p.name, want)
		}
	}
}
