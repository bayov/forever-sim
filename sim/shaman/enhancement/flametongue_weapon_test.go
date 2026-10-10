package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanFlametongueWeapon checks a rank 3 Flametongue Weapon hit with spell power for a
// level 30 Orc shaman (shaman_audit.md 7.3).
//
// A normal hit deals 8.84 per second of weapon speed, minus the 5 rank 3 loses on the beta,
// plus 10% of spell power that doesn't grow with speed. A crit deals 1.5 times that. On the
// beta with 91 spell power, Rage of the Storm (3.3 speed) hit 33.3 on average in 230 hits, where
// this gives 33.27. We also check Bloody Brass Knuckles (1.6), so a coefficient that grows with
// speed would fail.
func TestOrcShamanFlametongueWeapon(t *testing.T) {
	const spellPower = 91.0
	weapons := []struct {
		name  string
		id    int32
		speed float64
	}{{"Rage of the Storm", 280604, 3.3}, {"Bloody Brass Knuckles", 7683, 1.6}}
	for _, w := range weapons {
		t.Run(w.name, func(t *testing.T) {
			sim, enh := weaponSim{level: 30, weapon: &proto.ItemSpec{Id: w.id}, imbue: proto.WeaponImbue_FlametongueWeapon, seconds: 10}.start()
			sim.PrePull()
			if speed := enh.MainHand().SwingSpeed; speed != w.speed {
				t.Fatalf("%s has speed %.1f, want %.1f", w.name, speed, w.speed)
			}
			if sp := enh.GetStat(stats.SpellPower) + enh.GetStat(stats.FirePower); sp != 0 {
				t.Fatalf("fire spell power is %.0f, want 0", sp)
			}
			enh.AddStatDynamic(sim, stats.SpellPower, spellPower)
			spell := enh.GetSpell(core.ActionID{SpellID: shaman.FlametongueWeaponSpellId[3]})
			if spell == nil {
				t.Fatalf("no Flametongue Weapon rank 3")
			}

			// We take over the imbue's hit callback to collect the hits.
			var hits, crits []float64
			enh.GetAura("Flametongue Imbue").OnSpellHitDealt = func(_ *core.Aura, _ *core.Simulation, s *core.Spell, result *core.SpellResult) {
				if s != spell {
					return
				}
				switch result.Outcome {
				case core.OutcomeHit:
					hits = append(hits, result.Damage)
				case core.OutcomeCrit:
					crits = append(crits, result.Damage)
				}
			}
			for range 20 {
				spell.Cast(sim, enh.CurrentTarget)
			}
			enh.AddStatDynamic(sim, stats.SpellCrit, 100*core.SpellCritRatingPerCritChance)
			for range 20 {
				spell.Cast(sim, enh.CurrentTarget)
			}

			want := 884.0/100*w.speed - 5 + spellPower/10
			for _, c := range []struct {
				name   string
				damage []float64
				want   float64
			}{{"hit", hits, want}, {"crit", crits, want * 1.5}} {
				if len(c.damage) < 10 {
					t.Errorf("only %d %ss in 20 casts", len(c.damage), c.name)
				}
				for _, damage := range c.damage {
					if math.Abs(damage-c.want) > 1e-9 {
						t.Errorf("a %s for %.4f, want %.4f", c.name, damage, c.want)
						break
					}
				}
			}
		})
	}
}
