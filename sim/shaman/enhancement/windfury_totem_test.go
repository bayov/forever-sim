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

// TestOrcShamanWindfuryTotem checks Windfury Totem's attack power by rank, how often it procs
// and its wait between procs (shaman_audit.md 7.19).
//
// The extra swing gets 95, 179 or 246 attack power at levels 32, 42 and 52, as in the Forever
// client. Each landed white hit has a 20% chance to proc it, so 500 hits 2 sec apart should
// proc it about 100 times. The sim waits 100 ms between procs under Forever, the client's
// internal cooldown on the raid aura 10612 (Era's sim had 1.5 sec), so hits 0.5 sec apart
// proc it as often as hits 2 sec apart; notes.md asks to check the 100 ms in game. The totem
// is another shaman's, and we have Rockbiter on, so Windfury Weapon's own procs don't get in
// the way of the count (TestOrcShamanImbueWithWeaponTotem has both). A proc moves our next swing
// to right now.
func TestOrcShamanWindfuryTotem(t *testing.T) {
	start := func(level int32, seconds float64) (*core.Simulation, *EnhancementShaman) {
		sim, enh := weaponSim{level: level, weapon: &proto.ItemSpec{Id: darkEdge}, imbue: proto.WeaponImbue_RockbiterWeapon,
			totem: proto.TotemWeaponBuff_TotemWeaponBuffWindfury, seconds: seconds}.start()
		sim.PrePull()
		at(sim, 0.5, enh.AutoAttacks.CancelAutoSwing)
		return sim, enh
	}

	for _, c := range []struct {
		level int32
		ap    float64
	}{{32, 95}, {42, 179}, {52, 246}} {
		sim, enh := start(c.level, 2)
		at(sim, 1, func(sim *core.Simulation) {
			before := enh.GetStat(stats.AttackPower)
			enh.GetAura("Windfury Attack").Activate(sim)
			if got := enh.GetStat(stats.AttackPower) - before; got != c.ap {
				t.Errorf("level %d: Windfury Totem gave %.0f attack power, want %.0f", c.level, got, c.ap)
			}
		})
		runSim(sim)
	}

	// procTimes lands n white hits every apart, and returns when they procced it.
	procTimes := func(n int, every float64) []time.Duration {
		sim, enh := start(60, 1+every*float64(n))
		var procs []time.Duration
		for i := range n {
			at(sim, 1+every*float64(i), func(sim *core.Simulation) {
				enh.AutoAttacks.MHAuto().CalcAndDealDamage(sim, enh.CurrentTarget, 100, forcedOutcome(core.OutcomeHit))
				if enh.AutoAttacks.MainhandSwingAt() == sim.CurrentTime {
					procs = append(procs, sim.CurrentTime)
				}
			})
		}
		runSim(sim)
		return procs
	}

	// 100 procs give or take 9, and we allow 4 times that.
	if procs := len(procTimes(500, 2)); math.Abs(float64(procs)-100) > 36 {
		t.Errorf("500 white hits 2 sec apart procced Windfury Totem %d times, want about 100", procs)
	}

	// 400 hits 0.5 sec apart: 80 procs give or take 8, where a 1.5 sec cooldown would allow about 50.
	if procs := len(procTimes(400, 0.5)); math.Abs(float64(procs)-80) > 32 || procs <= 60 {
		t.Errorf("400 white hits 0.5 sec apart procced Windfury Totem %d times, want about 80", procs)
	}
}

// TestOrcShamanImbueWithWeaponTotem checks that a weapon imbue and the same weapon totem both
// proc on one shaman.
//
// Blizzard's class tuning hotfixes of 2026-11-10: "Weapon imbues and Weapon totems can now be
// used at the same time ... the Shaman can benefit from Windfury Weapon and Windfury Totem at the
// same time, granting their attacks a chance to activate both." Before it Windfury Weapon turned
// Windfury Totem off for us, and Flametongue Weapon Flametongue Totem. Each Windfury proc keeps
// its own 20% and cooldown (1.5 sec for the weapon, 100 ms for the totem; report assumption
// wf_weapon_totem_icd).
//
// We swing with Dark Edge of Insanity (3.5 speed) for 10 min at a level 63 boss, with another
// shaman's totem, and count each proc's metrics. About 150 white swings land, so Windfury Totem
// should give about 30 Windfury Attack buffs and Windfury Weapon about as many procs of 2 attacks
// each; Flametongue Totem hits once a landed white swing and Flametongue Weapon once a landed hit.
func TestOrcShamanImbueWithWeaponTotem(t *testing.T) {
	run := func(imbue proto.WeaponImbue, totem proto.TotemWeaponBuff) *proto.UnitMetrics {
		player := newOrcShaman(60, "", &proto.EnhancementShaman_Options{ShamanImbue: imbue})
		for range proto.ItemSlot_ItemSlotMainHand {
			player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
		}
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: darkEdge})
		result := core.RunRaidSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{TotemWeaponBuff: totem}, &proto.Debuffs{}),
			Encounter:  &proto.Encounter{Duration: 600, Targets: []*proto.Target{core.NewDefaultTarget()}},
			SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, Iterations: 1, RandomSeed: 1},
		})
		if result.Error != nil {
			t.Fatal(result.Error.Message)
		}
		return result.RaidMetrics.Parties[0].Players[0]
	}
	// hits is how many times the action with id landed (casts for a proc that always lands).
	hits := func(player *proto.UnitMetrics, id core.ActionID) int32 {
		var n int32
		for _, action := range player.Actions {
			if core.ProtoToActionID(action.Id) != id {
				continue
			}
			for _, target := range action.Targets {
				n += target.Hits + target.Crits + target.Glances + target.Blocks
			}
		}
		return n
	}
	procs := func(player *proto.UnitMetrics, id core.ActionID) float64 {
		for _, aura := range player.Auras {
			if core.ProtoToActionID(aura.Id) == id {
				return aura.ProcsAvg
			}
		}
		return 0
	}

	t.Run("Windfury Weapon and Windfury Totem", func(t *testing.T) {
		player := run(proto.WeaponImbue_WindfuryWeapon, proto.TotemWeaponBuff_TotemWeaponBuffWindfury)
		weapon := hits(player, core.ActionID{SpellID: shaman.WindfuryWeaponSpellId[4], Tag: 1})
		totem := procs(player, core.ActionID{SpellID: core.WindfuryBuffSpellId[3]})
		// Each Windfury Weapon proc brings 2 attacks.
		if weapon < 30 || totem < 15 {
			t.Errorf("in 10 min Windfury Weapon landed %d attacks (%d procs) and Windfury Totem procced %.0f times, want about 30 procs each", weapon, weapon/2, totem)
		}
	})
	t.Run("Flametongue Weapon and Flametongue Totem", func(t *testing.T) {
		player := run(proto.WeaponImbue_FlametongueWeapon, proto.TotemWeaponBuff_TotemWeaponBuffFlametongue)
		weapon := hits(player, core.ActionID{SpellID: shaman.FlametongueWeaponSpellId[len(shaman.FlametongueWeaponSpellId)-1]})
		totem := hits(player, core.ActionID{SpellID: core.FlametongueTotemProcSpellId[len(core.FlametongueTotemProcSpellId)-1], Tag: 1})
		if weapon < 100 || totem < 100 {
			t.Errorf("in 10 min Flametongue Weapon hit %d times and Flametongue Totem %d times, want both about 150", weapon, totem)
		}
	})
}
