package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanFrostbrandWeapon checks how often Frostbrand Weapon procs and what its hit deals
// (shaman_audit.md 7.5).
//
// It procs 9 times a minute, so each landed main hand hit has a 9 * speed / 60 chance. The hit
// deals the rank's flat damage, which doesn't depend on weapon speed, plus 10% of spell power.
// Elemental Weapons 3/3 multiplies the whole hit by 1.15. The beta hasn't shown us either one
// yet, so notes.md Need to Verify has them.
func TestOrcShamanFrostbrandWeapon(t *testing.T) {
	t.Run("procs", testFrostbrandWeaponProcs)
	t.Run("damage", testFrostbrandWeaponDamage)
}

// We swing for 2 hours at a level 32 target with Rage of the Storm (3.3 speed, 49.5% a hit) and
// Bloody Brass Knuckles (1.6 speed, 24% a hit). Each one's rate is then within 4 standard errors
// (4.7% and 2.8%) of its chance, unless the sim is off. A flat chance would fail one of them.
func testFrostbrandWeaponProcs(t *testing.T) {
	weapons := []struct {
		name  string
		id    int32
		speed float64
	}{{"Rage of the Storm", 280604, 3.3}, {"Bloody Brass Knuckles", 7683, 1.6}}
	for _, w := range weapons {
		sim, enh := weaponSim{level: 30, weapon: &proto.ItemSpec{Id: w.id}, imbue: proto.WeaponImbue_FrostbrandWeapon, seconds: 7200}.start()
		sim.PrePull()
		frostbrand := core.ActionID{SpellID: shaman.FrostbrandWeaponSpellId[2]}

		hits, procs := 0, 0
		imbue := enh.GetAura("Frostbrand Imbue")
		onHit := imbue.OnSpellHitDealt
		imbue.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ActionID == frostbrand {
				procs++
				onHit(aura, sim, spell, result)
				return
			}
			before := procs
			onHit(aura, sim, spell, result)
			if result.Landed() && spell.ProcMask.Matches(core.ProcMaskMeleeMH) {
				hits++
			} else if procs != before {
				t.Errorf("%s: a %s that didn't land procced Frostbrand", w.name, spell.ActionID)
			}
		}
		runSim(sim)

		chance := 9 * w.speed / 60
		rate := float64(procs) / float64(hits)
		se := math.Sqrt(chance * (1 - chance) / float64(hits))
		if math.Abs(rate-chance) > 4*se {
			t.Errorf("%s: Frostbrand procced on %d of %d hits (%.3f), want %.3f +- %.3f", w.name, procs, hits, rate, chance, 4*se)
		}
		if hits < 1500 {
			t.Errorf("%s: only %d landed hits", w.name, hits)
		}
	}
}

// Each rank at its top level, and rank 2 at 30 for the beta, with 100 spell power and Elemental
// Weapons 3/3.
func testFrostbrandWeaponDamage(t *testing.T) {
	const spellPower = 100.0
	cases := []struct {
		level  int32
		rank   int
		damage float64
	}{{30, 2, 54}, {36, 2, 72}, {46, 3, 117}, {56, 4, 159}, {60, 5, 169.2}}
	for _, c := range cases {
		sim, enh := weaponSim{level: c.level, talents: "-00000003", weapon: &proto.ItemSpec{Id: darkEdge}, imbue: proto.WeaponImbue_FrostbrandWeapon, seconds: 10}.start()
		sim.PrePull()
		spell := enh.GetSpell(core.ActionID{SpellID: shaman.FrostbrandWeaponSpellId[c.rank]})
		if spell == nil {
			t.Fatalf("at %d: no Frostbrand Weapon rank %d", c.level, c.rank)
		}
		enh.AddStatDynamic(sim, stats.SpellPower, spellPower)

		// We take over the imbue's hit callback to collect the hits.
		var hits []float64
		enh.GetAura("Frostbrand Imbue").OnSpellHitDealt = func(_ *core.Aura, _ *core.Simulation, s *core.Spell, result *core.SpellResult) {
			if s == spell && result.Outcome == core.OutcomeHit {
				hits = append(hits, result.Damage)
			}
		}
		for range 20 {
			spell.Cast(sim, enh.CurrentTarget)
		}

		want := (c.damage + spellPower/10) * 1.15
		if len(hits) == 0 {
			t.Errorf("at %d: no normal hits in 20 casts", c.level)
		}
		for _, damage := range hits {
			if math.Abs(damage-want) > 1e-9 {
				t.Errorf("at %d: Frostbrand hit for %.4f, want %.4f", c.level, damage, want)
				break
			}
		}
	}
}
