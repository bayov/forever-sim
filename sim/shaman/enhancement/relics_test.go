package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanRelics checks what each shaman relic with a damage or mana effect does
// (shaman_audit.md 9.2).
//
// The values come from the Forever client (1.60.1.70338). We build a level 60 shaman with Dark
// Edge of Insanity and each relic, and compare it with one that has no relic. Polished
// Driftwood Icon has its own beta reading (TestOrcShamanDriftwoodIcon).
func TestOrcShamanRelics(t *testing.T) {
	// Lightning Overload 3/3, Stormstrike and Maelstrom Weapon 5/5.
	const talents = "000000000000300-00000000000010005"
	relicSim := func(relic int32) (*core.Simulation, *EnhancementShaman) {
		return weaponSim{level: 60, talents: talents, weapon: &proto.ItemSpec{Id: darkEdge}, relic: relic, seconds: 10}.start()
	}
	_, plain := relicSim(0)
	target := plain.CurrentTarget

	t.Run("Burning Totem", func(t *testing.T) {
		// Flame Shock lasts 3 sec longer, which is one more 3 sec tick.
		_, enh := relicSim(shaman.BurningTotem)
		dot := topRank(t, "Flame Shock", enh.FlameShock).Dot(enh.CurrentTarget)
		if dot.NumberOfTicks != 5 || dot.Duration != 15*time.Second {
			t.Errorf("Flame Shock has %d ticks over %s, want 5 over 15s", dot.NumberOfTicks, dot.Duration)
		}
		if plainDot := topRank(t, "Flame Shock", plain.FlameShock).Dot(target); plainDot.NumberOfTicks != 4 {
			t.Errorf("without the relic Flame Shock has %d ticks, want 4", plainDot.NumberOfTicks)
		}
	})

	t.Run("Totem of the Storm (Classic)", func(t *testing.T) {
		// 33 spell power on Lightning Bolt and Chain Lightning, and none on the shocks.
		_, enh := relicSim(shaman.TotemOfTheStorm)
		for _, s := range []struct {
			name        string
			with, plain []*core.Spell
			want        float64
		}{
			{"Lightning Bolt", enh.LightningBolt, plain.LightningBolt, 33},
			{"Chain Lightning", enh.ChainLightning, plain.ChainLightning, 33},
			{"Earth Shock", enh.EarthShock, plain.EarthShock, 0},
		} {
			got := topRank(t, s.name, s.with).GetBonusDamage(enh.CurrentTarget) - topRank(t, s.name, s.plain).GetBonusDamage(target)
			if math.Abs(got-s.want) > 1e-9 {
				t.Errorf("%s gets %.1f spell power, want %.0f", s.name, got, s.want)
			}
		}
	})

	t.Run("Totem of Thunder and Lightning's Grasp", func(t *testing.T) {
		// Each gives 1% crit on Lightning Bolt and its overloads, and none on Chain Lightning.
		// Lightning's Grasp is a one-hand weapon we hold in place of Dark Edge. Its Intellect
		// moves every spell's crit, so we measure the bolt's crit over Chain Lightning's.
		_, totem := relicSim(shaman.TotemOfThunder)
		_, grasp := weaponSim{level: 60, talents: talents, weapon: &proto.ItemSpec{Id: shaman.LightningsGrasp}, seconds: 10}.start()
		lead := func(e *EnhancementShaman, spell *core.Spell) float64 {
			return spell.SpellCritChance(e.CurrentTarget) - topRank(t, "Chain Lightning", e.ChainLightning).SpellCritChance(e.CurrentTarget)
		}
		for _, e := range []struct {
			name string
			enh  *EnhancementShaman
		}{{"Totem of Thunder", totem}, {"Lightning's Grasp", grasp}} {
			lb := topRank(t, "Lightning Bolt", e.enh.LightningBolt)
			plainLB := topRank(t, "Lightning Bolt", plain.LightningBolt)
			for _, s := range []struct {
				name        string
				with, plain *core.Spell
			}{
				{"Lightning Bolt", lb, plainLB},
				{"its overload", e.enh.LightningBoltOverload[lb.Rank], plain.LightningBoltOverload[plainLB.Rank]},
			} {
				if got := lead(e.enh, s.with) - lead(plain, s.plain); math.Abs(got-.01) > 1e-9 {
					t.Errorf("%s: %s gets %.2f%% more crit than Chain Lightning, want 1%%", e.name, s.name, 100*got)
				}
			}
		}
	})

	t.Run("Totem of Charged Flames", func(t *testing.T) {
		// 5 mana off Flame Shock and Lightning Bolt, and none off Earth Shock.
		_, enh := relicSim(shaman.TotemOfChargedFlames)
		for _, s := range []struct {
			name        string
			with, plain []*core.Spell
			want        float64
		}{
			{"Flame Shock", enh.FlameShock, plain.FlameShock, 5},
			{"Lightning Bolt", enh.LightningBolt, plain.LightningBolt, 5},
			{"Earth Shock", enh.EarthShock, plain.EarthShock, 0},
		} {
			if got := topRank(t, s.name, s.plain).Cost.GetCurrentCost() - topRank(t, s.name, s.with).Cost.GetCurrentCost(); math.Abs(got-s.want) > 1e-9 {
				t.Errorf("%s costs %.2f less, want %.0f", s.name, got, s.want)
			}
		}
	})

	t.Run("Firestorm Totem", func(t *testing.T) {
		// 5% off every totem, and nothing off the Fire Nova spell.
		_, enh := relicSim(shaman.FirestormTotem)
		for _, s := range []struct {
			name        string
			with, plain []*core.Spell
			want        float64
		}{
			{"Searing Totem", enh.SearingTotem, plain.SearingTotem, .95},
			{"Windfury Totem", enh.WindfuryTotem, plain.WindfuryTotem, .95},
			{"Mana Spring Totem", enh.ManaSpringTotem, plain.ManaSpringTotem, .95},
			{"Fire Nova", enh.FireNova, plain.FireNova, 1},
		} {
			if got := topRank(t, s.name, s.with).Cost.GetCurrentCost() / topRank(t, s.name, s.plain).Cost.GetCurrentCost(); math.Abs(got-s.want) > 1e-9 {
				t.Errorf("%s costs %.3f of its cost, want %.2f", s.name, got, s.want)
			}
		}
	})

	t.Run("Totem of the Storm (Forever)", func(t *testing.T) {
		// A Lightning Bolt that lands rolls half the chance of a main hand swing to give a
		// Maelstrom Weapon stack. Without the relic a bolt never gives one. We land 20,000 bolts
		// with each and count the stacks.
		const bolts = 20000
		for _, c := range []struct {
			name  string
			relic int32
			half  bool
		}{{"without the relic", 0, false}, {"with the relic", shaman.ForeverTotemOfTheStorm, true}} {
			sim, enh := relicSim(c.relic)
			lb := topRank(t, "Lightning Bolt", enh.LightningBolt)
			mw := enh.MaelstromWeaponAura
			stacks := 0
			at(sim, 1, func(sim *core.Simulation) {
				enh.AutoAttacks.CancelAutoSwing(sim)
				for range bolts {
					mw.Deactivate(sim)
					lb.DealDamage(sim, lb.CalcDamage(sim, enh.CurrentTarget, 1, forcedOutcome(core.OutcomeHit)))
					if mw.IsActive() {
						stacks++
					}
				}
			})
			runSim(sim)
			want := 0.0
			if c.half {
				want = .5 * 10 * enh.AutoAttacks.MH().SwingSpeed / 60
			}
			rate := float64(stacks) / bolts
			if se := math.Sqrt(want * (1 - want) / bolts); math.Abs(rate-want) > 4*se+1e-9 {
				t.Errorf("%s: %.4f of bolts gave a stack, want %.4f", c.name, rate, want)
			}
		}
	})
}
