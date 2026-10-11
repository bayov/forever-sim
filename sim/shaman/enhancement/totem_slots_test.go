package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/shaman"
)

// slotTotem is one of our totems, with how to see that it's doing its thing.
type slotTotem struct {
	name  string
	spell *core.Spell
	// up is nil for a totem that does nothing in the sim.
	up    func() bool
	lasts time.Duration
}

// newTotemSlotSim starts a level 60 Orc shaman with Mana Tide Totem and returns every totem
// it knows, by element.
func newTotemSlotSim(t *testing.T, seconds float64) (*core.Simulation, *EnhancementShaman, map[string][]slotTotem) {
	// Mana Tide Totem.
	sim, enh := newShamanSim(newOrcShaman(60, "--000000000001", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, seconds)
	aura := func(label string) func() bool {
		a := enh.GetAura(label)
		if a == nil {
			t.Fatalf("no %s aura", label)
		}
		return a.IsActive
	}
	dot := func(spell *core.Spell) func() bool {
		return func() bool { return spell.Dot(enh.CurrentTarget).IsActive() }
	}
	healingStream := topRank(t, "Healing Stream Totem", enh.HealingStreamTotem)
	searing := topRank(t, "Searing Totem", enh.SearingTotem)
	magma := topRank(t, "Magma Totem", enh.MagmaTotem)
	flametongue := topRank(t, "Flametongue Totem", enh.FlametongueTotem)
	windfury := enh.GetAuraByID(core.ActionID{SpellID: shaman.WindfuryBuffAuraId[3]})
	return sim, enh, map[string][]slotTotem{
		"earth": {
			{"Strength of Earth Totem", topRank(t, "Strength of Earth Totem", enh.StrengthOfEarthTotem), aura("Strength of Earth Totem"), 5 * time.Minute},
			{"Stoneskin Totem", topRank(t, "Stoneskin Totem", enh.StoneskinTotem), aura("Stoneskin"), 5 * time.Minute},
			{"Tremor Totem", enh.TremorTotem, nil, 5 * time.Minute},
		},
		"water": {
			{"Healing Stream Totem", healingStream, func() bool { return healingStream.Hot(&enh.Unit).IsActive() }, 5 * time.Minute},
			{"Mana Spring Totem", topRank(t, "Mana Spring Totem", enh.ManaSpringTotem), aura("Mana Spring Totem (Rank 4)"), 5 * time.Minute},
			{"Mana Tide Totem", topRank(t, "Mana Tide Totem", enh.ManaTideTotem), aura("Mana Tide Totem (Rank 3)"), 13 * time.Second},
		},
		"fire": {
			{"Searing Totem", searing, dot(searing), 55 * time.Second},
			{"Magma Totem", magma, dot(magma), 20 * time.Second},
			{"Flametongue Totem", flametongue, dot(flametongue), 5 * time.Minute},
		},
		"air": {
			{"Windfury Totem", topRank(t, "Windfury Totem", enh.WindfuryTotem), windfury.IsActive, 5 * time.Minute},
			{"Grace of Air Totem", topRank(t, "Grace of Air Totem", enh.GraceOfAirTotem), aura("Grace of Air Totem"), 5 * time.Minute},
			{"Windwall Totem", topRank(t, "Windwall Totem", enh.WindwallTotem), nil, 5 * time.Minute},
		},
	}
}

// TestOrcShamanTotemSlots checks that we have one totem of each element down at a time
// (shaman_audit.md 7.17).
//
// The client summons each totem into its element's slot, and a new one takes the place of the
// one there. So putting down a totem ends what the totem before it in the same element was
// doing, and leaves the other elements alone. On the beta every Healing Stream heal on us came
// while Healing Stream was the last water totem we put down, and every Mana Spring tick while
// Mana Spring was.
//
// We put down each element's totems in turn, one a second, and then the first one again. A
// Healing Stream Totem replaces Mana Tide Totem before its first tick, and then Mana Tide
// gives no mana. One tick is 290 mana, far more than we regen in that time.
func TestOrcShamanTotemSlots(t *testing.T) {
	sim, enh, elements := newTotemSlotSim(t, 30)
	manaTide := elements["water"][2]

	when := 1.0
	for _, element := range []string{"earth", "water", "fire", "air"} {
		totems := elements[element]
		for i := range len(totems) + 1 {
			placed := totems[i%len(totems)]
			at(sim, when, func(sim *core.Simulation) {
				enh.AddMana(sim, enh.MaxMana()-enh.CurrentMana(), enh.GetManaNotCastingMetrics())
				castNow(t, sim, enh, placed.spell)
				for _, totem := range totems {
					if totem.up != nil && totem.up() != (totem.spell == placed.spell) {
						t.Errorf("after %s: %s up: %v", placed.name, totem.name, totem.up())
					}
				}
				// The elements we put down before keep their last totem.
				for _, other := range []string{"earth", "water", "fire", "air"} {
					if other == element {
						break
					}
					last := elements[other][0]
					if last.up != nil && !last.up() {
						t.Errorf("after %s: %s went down", placed.name, last.name)
					}
				}
			})
			when++
		}
	}

	// Mana Tide Totem, then Healing Stream Totem a second later.
	var mana float64
	at(sim, when, func(sim *core.Simulation) {
		enh.SpendMana(sim, enh.CurrentMana()-200, enh.GetManaNotCastingMetrics())
		manaTide.spell.CD.Reset()
		castNow(t, sim, enh, manaTide.spell)
	})
	at(sim, when+1, func(sim *core.Simulation) {
		castNow(t, sim, enh, elements["water"][0].spell)
		mana = enh.CurrentMana()
	})
	at(sim, when+12.5, func(sim *core.Simulation) {
		if gained := enh.CurrentMana() - mana; gained >= 290 {
			t.Errorf("we gained %.0f mana after Healing Stream Totem took Mana Tide Totem's place", gained)
		}
	})
	runSim(sim)
}

// TestOrcShamanTotemDurations checks that each totem's effect ends when its time is up.
//
// Most totems last 5 min under Forever, Searing Totem rank 6 55 sec, Magma Totem 20 sec and
// Mana Tide Totem 13 sec, as in the client. Our own Stoneskin Totem's buff used to stay up
// for the rest of the fight. We put down each element's totems one after another, and each one
// when the one before has run out.
func TestOrcShamanTotemDurations(t *testing.T) {
	sim, enh, elements := newTotemSlotSim(t, 700)
	for _, element := range []string{"earth", "water", "fire", "air"} {
		when := 1.0
		for _, totem := range elements[element] {
			if totem.up == nil {
				continue
			}
			at(sim, when, func(sim *core.Simulation) {
				enh.AddMana(sim, enh.MaxMana()-enh.CurrentMana(), enh.GetManaNotCastingMetrics())
				castNow(t, sim, enh, totem.spell)
			})
			lasts := totem.lasts.Seconds()
			at(sim, when+lasts-0.1, func(sim *core.Simulation) {
				if !totem.up() {
					t.Errorf("%s is gone %.1f sec after we put it down", totem.name, lasts-0.1)
				}
			})
			at(sim, when+lasts+0.1, func(sim *core.Simulation) {
				if totem.up() {
					t.Errorf("%s is still up %.1f sec after we put it down", totem.name, lasts+0.1)
				}
			})
			when += lasts + 1
		}
	}
	runSim(sim)
}

// TestOrcShamanTotemSlotsKeepRaidBuffs checks that our totems leave another shaman's totem
// buff alone.
//
// With another shaman's Strength of Earth Totem in the raid buffs, our own Strength of Earth
// Totem shares its buff. When our Stoneskin Totem takes the place of our Strength of Earth, the
// other shaman's totem still stands, so the Strength stays.
func TestOrcShamanTotemSlotsKeepRaidBuffs(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(newOrcShaman(60, "", &proto.EnhancementShaman_Options{}), &proto.PartyBuffs{},
			&proto.RaidBuffs{StrengthOfEarthTotem: proto.TristateEffect_TristateEffectRegular}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 10, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
	strength := enh.GetAura("Strength of Earth Totem")
	at(sim, 1, func(sim *core.Simulation) {
		castNow(t, sim, enh, topRank(t, "Strength of Earth Totem", enh.StrengthOfEarthTotem))
		castNow(t, sim, enh, topRank(t, "Stoneskin Totem", enh.StoneskinTotem))
		if !strength.IsActive() {
			t.Errorf("our Stoneskin Totem took down the other shaman's Strength of Earth")
		}
	})
	runSim(sim)
}
