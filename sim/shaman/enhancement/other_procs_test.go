package enhancement

import (
	"slices"
	"testing"

	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

const (
	judgementOfWisdom = "Judgement of Wisdom"
	crusaderProc      = "Crusader"
	handOfJustice     = "Hand of Justice"
)

// TestOrcShamanOtherProcs checks which of our attacks trigger Judgement of Wisdom on the
// target and the melee item procs (Crusader and Hand of Justice).
//
// Judgement of Wisdom gives us mana on half of our white hits, even the misses, and on half
// of our landed Stormstrikes, Windfury Weapon attacks and spells. The melee item procs
// trigger on landed white hits, Stormstrikes and Windfury Weapon attacks. None of them come
// from proc damage, totem attacks or Lightning Shield orbs.
//
// We force each outcome 2000 times, 2 sec apart, because Hand of Justice procs on 1% with a
// 2 sec cooldown. It misses all 2000 one time in 500 million.
func TestOrcShamanOtherProcs(t *testing.T) {
	const (
		rolls = 2000
		every = 2.0
	)
	type attack struct {
		name    string
		spell   *core.Spell
		outcome core.HitOutcome
		procs   []string
	}
	all := []string{judgementOfWisdom, crusaderProc, handOfJustice}
	attacks := func(enh *EnhancementShaman) []attack {
		white := enh.AutoAttacks.MHAuto()
		list := []attack{
			{"white hit", white, core.OutcomeHit, all},
			{"white miss", white, core.OutcomeMiss, []string{judgementOfWisdom}},
			{"Stormstrike", enh.Stormstrike, core.OutcomeHit, all},
			{"missed Stormstrike", enh.Stormstrike, core.OutcomeMiss, nil},
			{"Windfury Weapon attack", enh.WindfuryWeaponMH, core.OutcomeHit, all},
			{"missed Windfury Weapon attack", enh.WindfuryWeaponMH, core.OutcomeMiss, nil},
			{"Flametongue Totem damage", lastSpell(enh, core.FlametongueTotemProcSpellId[:], 1), core.OutcomeHit, nil},
			{"Searing Totem attack", lastSpell(enh, shaman.SearingTotemAttackSpellId[:], 0), core.OutcomeHit, nil},
			{"Lightning Shield orb", topRank(t, "Lightning Shield", enh.LightningShieldProcs), core.OutcomeHit, nil},
			{"missed Earth Shock", topRank(t, "Earth Shock", enh.EarthShock), core.OutcomeMiss, nil},
		}
		for _, spell := range []struct {
			name  string
			ranks []*core.Spell
		}{
			{"Earth Shock", enh.EarthShock},
			{"Flame Shock", enh.FlameShock},
			{"Frost Shock", enh.FrostShock},
			{"Lightning Bolt", enh.LightningBolt},
			{"Chain Lightning", enh.ChainLightning},
			{"Fire Nova", enh.FireNova},
		} {
			list = append(list, attack{spell.name, topRank(t, spell.name, spell.ranks), core.OutcomeHit, []string{judgementOfWisdom}})
		}
		return list
	}

	setup := weaponSim{
		level:   60,
		weapon:  &proto.ItemSpec{Id: darkEdge, Enchant: 1900},
		imbue:   proto.WeaponImbue_WindfuryWeapon,
		totem:   proto.TotemWeaponBuff_TotemWeaponBuffFlametongue,
		trinket: common.HandOfJustice,
		debuffs: &proto.Debuffs{JudgementOfWisdom: true},
	}
	_, enh := setup.start()
	setup.seconds = every * float64(rolls*len(attacks(enh))+1)
	sim, enh := setup.start()
	sim.PrePull()
	// Combat start turns the swings on, so we turn them off after it.
	at(sim, 0.5, enh.AutoAttacks.CancelAutoSwing)
	// We roll each attack on its own, without the Windfury and Flametongue Totem procs it
	// would bring.
	enh.GetAura("Windfury Imbue").Deactivate(sim)
	enh.GetAura("Flametongue Totem Raid (Rank 4)").Deactivate(sim)
	// A proc from anything but a main hand attack would show up on the off hand buff.
	crusader := []*core.Aura{enh.GetAura("Crusader Enchant MH"), enh.GetAura("Crusader Enchant OH")}
	metrics := enh.GetManaNotCastingMetrics()

	// We queue one roll at a time, because tens of thousands of queued actions slow the sim
	// down a lot.
	list := attacks(enh)
	seen := map[string]bool{}
	done := 0
	var roll func(n int) func(sim *core.Simulation)
	roll = func(n int) func(sim *core.Simulation) {
		return func(sim *core.Simulation) {
			a := list[n/rolls]
			done++
			// Judgement of Wisdom's mana must have room to show.
			if enh.CurrentMana() > enh.MaxMana()-100 {
				enh.SpendMana(sim, 1000, metrics)
			}
			mana := enh.CurrentMana()
			a.spell.CalcAndDealDamage(sim, enh.CurrentTarget, 100, forcedOutcome(a.outcome))
			if enh.CurrentMana() > mana {
				seen[judgementOfWisdom] = true
			}
			for _, aura := range crusader {
				if aura.IsActive() {
					seen[crusaderProc] = true
					aura.Deactivate(sim)
				}
			}
			// A Hand of Justice proc moves the next swing to right now.
			if enh.AutoAttacks.MainhandSwingAt() == sim.CurrentTime {
				seen[handOfJustice] = true
			}

			if n%rolls == rolls-1 {
				for _, name := range all {
					if want := slices.Contains(a.procs, name); seen[name] != want {
						t.Errorf("%s triggered %s: %t, want %t", a.name, name, seen[name], want)
					}
				}
				seen = map[string]bool{}
			}
			if n+1 < len(list)*rolls {
				at(sim, every*float64(n+2), roll(n+1))
			}
		}
	}
	at(sim, every, roll(0))
	runSim(sim)
	if done != len(list)*rolls {
		t.Fatalf("rolled %d times, want %d", done, len(list)*rolls)
	}
}
