package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanWindfuryWeapon checks how often Windfury Weapon procs and what its attacks deal
// (shaman_audit.md 7.2).
//
// It procs on 20% of our landed main hand melee hits, and then not again for 1.5 sec. Each proc
// makes two attacks right away. On the beta, 145 procs came in 772 hits off cooldown (18.8%),
// every proc had both attacks unless the first one killed the mob, and the shortest gap between
// two procs was 1.55 sec.
//
// The attacks add the rank's extra attack power at the level it's learned: 119 for rank 2 at
// 40, 249 for rank 3 at 50 and 333 for rank 4 at 60. Elemental Weapons tests rank 1 at 30.
func TestOrcShamanWindfuryWeapon(t *testing.T) {
	t.Run("procs", testWindfuryWeaponProcs)
	t.Run("damage", testWindfuryWeaponDamage)
}

// We fight a level 63 target for 2 hours with Stormstrike on cooldown, so about 2400 white hits
// and Stormstrikes land off cooldown. Their proc rate is then within 4 standard errors (3.3%)
// of 20%, unless the sim is off.
func testWindfuryWeaponProcs(t *testing.T) {
	player := newOrcShaman(60, stormstrike, &proto.EnhancementShaman_Options{ShamanImbue: proto.WeaponImbue_WindfuryWeapon})
	for range proto.ItemSlot_ItemSlotMainHand {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
	}
	player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: darkEdge})
	player.Rotation.PriorityList = []*proto.APLListItem{{Action: &proto.APLAction{
		Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
			SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 17364}},
		}},
	}}}
	sim, enh := newShamanSim(player, &proto.Debuffs{}, 7200)
	metrics := enh.GetManaNotCastingMetrics()

	const cooldown = 1500 * time.Millisecond
	lastProc := -time.Hour
	hits, procs, attacks := 0, 0, 0
	imbue := enh.GetAura("Windfury Imbue")
	onHit := imbue.OnSpellHitDealt
	imbue.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if spell == enh.WindfuryWeaponMH {
			attacks++
			onHit(aura, sim, spell, result)
			return
		}
		if spell == enh.Stormstrike {
			// We never run out of mana for the next Stormstrike.
			enh.AddMana(sim, enh.MaxMana(), metrics)
		}
		before := attacks
		onHit(aura, sim, spell, result)
		made := attacks - before
		ready := sim.CurrentTime-lastProc >= cooldown
		if made == 0 {
			if ready && result.Landed() {
				hits++
			}
			return
		}
		if made != 2 {
			t.Errorf("at %s: a proc made %d attacks, want 2", sim.CurrentTime, made)
		}
		if !ready {
			t.Errorf("at %s: Windfury procced %s after the last proc", sim.CurrentTime, sim.CurrentTime-lastProc)
		}
		hits++
		procs++
		lastProc = sim.CurrentTime
	}
	runSim(sim)

	rate := float64(procs) / float64(hits)
	se := math.Sqrt(0.2 * 0.8 / float64(hits))
	if math.Abs(rate-0.2) > 4*se {
		t.Errorf("Windfury procced on %d of %d hits (%.3f), want 0.2 +- %.3f", procs, hits, rate, 4*se)
	}
	if hits < 2000 {
		t.Errorf("only %d hits off cooldown", hits)
	}
}

// We pin the weapon's damage roll to its minimum and take the target's armor away, so every
// normal Windfury hit deals the weapon's minimum plus (AP + bonus) / 14 times its speed.
func testWindfuryWeaponDamage(t *testing.T) {
	cases := []struct {
		level   int32
		rank    int
		extraAP float64
	}{{40, 2, 119}, {50, 3, 249}, {60, 4, 333}}
	for _, c := range cases {
		sim, enh := weaponSim{level: c.level, weapon: &proto.ItemSpec{Id: darkEdge}, imbue: proto.WeaponImbue_WindfuryWeapon, seconds: 10}.start()
		sim.PrePull()
		if enh.WindfuryWeaponMH.ActionID.SpellID != shaman.WindfuryWeaponSpellId[c.rank] {
			t.Fatalf("at %d: Windfury Weapon is %s, want rank %d", c.level, enh.WindfuryWeaponMH.ActionID, c.rank)
		}
		mh := enh.AutoAttacks.MH()
		mh.BaseDamageMax = mh.BaseDamageMin
		want := mh.BaseDamageMin + mh.SwingSpeed*(enh.GetStat(stats.AttackPower)+c.extraAP)/core.DefaultAttackPowerPerDPS

		// We take over the imbue's hit callback, so the attacks we cast don't proc more.
		var hits []float64
		enh.GetAura("Windfury Imbue").OnSpellHitDealt = func(_ *core.Aura, _ *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell == enh.WindfuryWeaponMH && result.Outcome == core.OutcomeHit {
				hits = append(hits, result.Damage)
			}
		}
		for range 20 {
			enh.WindfuryWeaponMH.Cast(sim, enh.CurrentTarget)
		}
		if len(hits) == 0 {
			t.Errorf("at %d: no normal hits in 20 attacks", c.level)
		}
		for _, damage := range hits {
			if math.Abs(damage-want) > 1e-9 {
				t.Errorf("at %d: Windfury hit for %.2f, want %.2f", c.level, damage, want)
				break
			}
		}
	}
}
