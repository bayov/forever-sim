package enhancement

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanMaelstromWeaponRanks checks what each Maelstrom Weapon stack does at every rank,
// and how long the stacks last (shaman_audit.md 8.33).
//
// Since Blizzard's class tuning hotfixes of 2026-11-10 each stack takes 20% off the next Lightning
// Bolt's cast time and cost at every rank ("20% reduced cast time and Mana cost per application at
// all ranks (was 4/8/12/16/20%)"), so 5 stacks make it free and instant. The stacks last 30 sec
// (408505). We set 1 to 5 stacks at level 60 for each rank, then stop our swings, give one stack at
// 1 sec and want it gone between 30.9 and 31.1.
//
// A miss gives no stack. At 32 sec we miss 20 times and want none. How often a landed hit gives
// one is TestOrcShamanMaelstromWeaponProcRate's.
func TestOrcShamanMaelstromWeaponRanks(t *testing.T) {
	for points := 1; points <= 5; points++ {
		// Maelstrom Weapon is the seventeenth Enhancement talent.
		sim, enh := newShamanSim(newOrcShaman(60, fmt.Sprintf("-0000000000000000%d", points), &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 40)
		if enh.Talents.MaelstromWeapon != int32(points) {
			t.Fatalf("got Maelstrom Weapon %d, want %d", enh.Talents.MaelstromWeapon, points)
		}
		lb := topRank(t, "Lightning Bolt", enh.LightningBolt)
		cost := shaman.LightningBoltManaCost[lb.Rank]
		mw := enh.MaelstromWeaponAura

		at(sim, 0.5, func(sim *core.Simulation) {
			for stacks := 1; stacks <= 5; stacks++ {
				mw.Activate(sim)
				mw.SetStacks(sim, int32(stacks))
				share := max(0, 1-.2*float64(stacks))
				if got := lb.Cost.GetCurrentCost(); math.Abs(got-cost*share) > 1e-9 {
					t.Errorf("Maelstrom Weapon %d with %d stacks: Lightning Bolt costs %.2f, want %.2f", points, stacks, got, cost*share)
				}
				want := time.Duration(float64(lb.DefaultCast.CastTime) * share)
				if got := lb.CastTime(); (got - want).Abs() > time.Microsecond {
					t.Errorf("Maelstrom Weapon %d with %d stacks: Lightning Bolt casts in %s, want %s", points, stacks, got, want)
				}
			}
			mw.Deactivate(sim)
		})
		// We stop our swings so they don't give more stacks.
		at(sim, 1, func(sim *core.Simulation) {
			enh.AutoAttacks.CancelAutoSwing(sim)
			mw.Activate(sim)
			mw.AddStack(sim)
		})
		at(sim, 30.9, func(sim *core.Simulation) {
			if !mw.IsActive() {
				t.Errorf("Maelstrom Weapon %d: the stack ran out before 30 sec", points)
			}
		})
		at(sim, 31.1, func(sim *core.Simulation) {
			if mw.IsActive() {
				t.Errorf("Maelstrom Weapon %d: the stack lasted past 30 sec", points)
			}
		})
		at(sim, 32, func(sim *core.Simulation) {
			for _, c := range []struct {
				name    string
				spell   *core.Spell
				outcome core.HitOutcome
				want    bool
			}{
				{"white miss", enh.AutoAttacks.MHAuto(), core.OutcomeMiss, false},
			} {
				for range 20 {
					mw.Deactivate(sim)
					c.spell.CalcAndDealDamage(sim, enh.CurrentTarget, 100, forcedOutcome(c.outcome))
					if got := mw.IsActive() && mw.GetStacks() == 1; got != c.want {
						t.Errorf("Maelstrom Weapon %d: a %s gave a stack: %t, want %t", points, c.name, got, c.want)
						break
					}
				}
			}
		})
		runSim(sim)
	}
}

// TestOrcShamanMaelstromWeaponProcRate checks how often a landed white swing gives a Maelstrom
// Weapon stack under Forever.
//
// Blizzard's class tuning hotfixes of 2026-11-10: it "activates 6/7/8/9/10 times per minute" by
// rank, and the tooltip shows the chance a swing by main hand speed: "5 points into Maelstrom
// Weapon and ... a 3.6 speed weapon ... 60% chance to activate on swing, compared to ... a 2.0
// speed weapon ... 33%". That is (5 + rank) PPM times the speed over 60: 60% at 5/5 and 36% at 1/5
// with a 3.6 speed staff, 33.3% at 5/5 with a 2.0 speed mace. Before it every landed melee hit
// gave a stack (our logs of 2026-11-08 and 09).
//
// We stop our swings and land 20,000 white hits for each case, and want the share that gave a
// stack within 4 standard errors.
func TestOrcShamanMaelstromWeaponProcRate(t *testing.T) {
	const (
		swings             = 20000
		thaumaturgistStaff = 15275 // a 3.6 speed staff
		rubidiumHammer     = 11805 // a 2.0 speed one-handed mace
	)
	for _, c := range []struct {
		name   string
		weapon int32
		speed  float64
		points int
		want   float64
	}{
		{"5/5, 3.6 speed", thaumaturgistStaff, 3.6, 5, .6},
		{"1/5, 3.6 speed", thaumaturgistStaff, 3.6, 1, .36},
		{"5/5, 2.0 speed", rubidiumHammer, 2.0, 5, 1.0 / 3},
	} {
		t.Run(c.name, func(t *testing.T) {
			// Maelstrom Weapon is the seventeenth Enhancement talent.
			sim, enh := weaponSim{level: 60, talents: fmt.Sprintf("-0000000000000000%d", c.points), weapon: &proto.ItemSpec{Id: c.weapon}, seconds: 5}.start()
			if got := enh.MainHand().SwingSpeed; got != c.speed {
				t.Fatalf("the main hand swings every %.2f sec, want %.2f", got, c.speed)
			}
			mw := enh.MaelstromWeaponAura
			stacks := 0
			at(sim, 1, func(sim *core.Simulation) {
				enh.AutoAttacks.CancelAutoSwing(sim)
				white := enh.AutoAttacks.MHAuto()
				for range swings {
					mw.Deactivate(sim)
					white.CalcAndDealDamage(sim, enh.CurrentTarget, 100, forcedOutcome(core.OutcomeHit))
					if mw.IsActive() {
						stacks++
					}
				}
			})
			runSim(sim)
			rate := float64(stacks) / swings
			if se := math.Sqrt(c.want * (1 - c.want) / swings); math.Abs(rate-c.want) > 4*se {
				t.Errorf("%.4f of landed white swings gave a stack, want %.4f", rate, c.want)
			}
		})
	}
}

// TestOrcShamanMaelstromWeaponSkipsWindfuryTotem checks that Windfury Totem's extra swing gives no
// Maelstrom Weapon stack under Forever, while another proc's extra swing (Hand of Justice) still
// does at the white swing's rate.
//
// Combat logs of 2026-11-13 and 14 (5/5, after push 114060): 0 of 140 landed totem swings at
// totem ranks 1 and 2 and 0 of 121 at rank 3 gave a stack, against 0.59 a landed white swing
// (wowsims/forever EV-0009, 434723c516).
//
// We give 400 extra swings from each proc one at a time with a 3.6 speed staff at 5/5 (60% a
// landed swing), and count the stacks from the ones that landed, leaving out any that a normal
// swing came close to.
func TestOrcShamanMaelstromWeaponSkipsWindfuryTotem(t *testing.T) {
	const thaumaturgistStaff = 15275
	sim, enh := weaponSim{level: 60, talents: "-00000000000000005", weapon: &proto.ItemSpec{Id: thaumaturgistStaff}, seconds: 3600}.start()
	sim.PrePull()
	mw := enh.MaelstromWeaponAura
	for sim.CurrentTime < time.Second && !sim.Step() {
	}
	mh := enh.AutoAttacks.MHAuto()

	for _, c := range []struct {
		name   string
		source int32
		want   float64
	}{
		{"Windfury Totem rank 3", core.WindfuryBuffSpellId[3], 0},
		{"Windfury Totem rank 1", core.WindfuryBuffSpellId[1], 0},
		{"Hand of Justice", 15601, .6},
	} {
		landed, stacks := 0, 0
		// A normal swing lands a batch after its cast, so it can give a stack in the next round too.
		nearNormal := false
		for range 400 {
			mw.Deactivate(sim)
			normal := mh.SpellMetrics()[enh.CurrentTarget.UnitIndex].Casts
			enh.AutoAttacks.ExtraMHAttack(sim, 1, core.ActionID{SpellID: c.source}, core.ActionID{})
			// The extra swings' metrics (split 1), which the call above switched to.
			metrics := &mh.SpellMetrics()[enh.CurrentTarget.UnitIndex]
			before := *metrics
			// The swing comes at once and lands a batch later.
			for end := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
			}
			if metrics.Casts != before.Casts+1 {
				t.Fatalf("%s: %d extra swings, want 1", c.name, metrics.Casts-before.Casts)
			}
			mh.SetMetricsSplit(0)
			wasNear := nearNormal
			nearNormal = mh.SpellMetrics()[enh.CurrentTarget.UnitIndex].Casts != normal
			if wasNear || nearNormal {
				continue
			}
			if metrics.Misses != before.Misses || metrics.Dodges != before.Dodges || metrics.Parries != before.Parries {
				continue
			}
			landed++
			if mw.IsActive() {
				stacks++
			}
		}
		if landed < 200 {
			t.Fatalf("%s: %d of 400 extra swings landed clear of a normal swing, want at least 200", c.name, landed)
		}
		rate := float64(stacks) / float64(landed)
		t.Logf("%s: %d of %d landed extra swings gave a stack", c.name, stacks, landed)
		if se := math.Sqrt(.6 * .4 / float64(landed)); c.want == 0 && stacks != 0 || math.Abs(rate-c.want) > 4*se {
			t.Errorf("%s: %d of %d landed extra swings gave a stack (%.3f), want %.2f", c.name, stacks, landed, rate, c.want)
		}
	}
}
