package dpsrogue

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

// newTotemRogueSim starts a Forever sim of a naked Orc rogue with Brutality Blade (2.5 speed)
// in the main hand and Perdition's Blade (1.8 speed) in the off hand, under another shaman's
// totem. The rotation is empty, so the rogue only swings.
func newTotemRogueSim(level int32, totem proto.TotemWeaponBuff, seconds float64) (*core.Simulation, *DpsRogue) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotOffHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 18832} // Brutality Blade
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: 18816}  // Perdition's Blade

	player := &proto.Player{
		Race:      proto.Race_RaceOrc,
		Class:     proto.Class_ClassRogue,
		Level:     level,
		Equipment: &proto.EquipmentSpec{Items: items},
		Rotation:  &proto.APLRotation{},
		Spec:      DefaultRogue,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{TotemWeaponBuff: totem}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: seconds, Targets: []*proto.Target{{Level: level, MobType: proto.MobType_MobTypeBeast}}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*DpsRogue)
}

// TestOrcRogueFlametongueTotemOffHand checks that our off-hand white swings proc Flametongue
// Totem under Forever, by the off hand's speed.
//
// Rank 1 adds 548 / 100 per second of weapon speed, so 13.70 for the main hand and 9.86 for
// the off hand. The target is level 30, so a normal hit isn't partly resisted.
func TestOrcRogueFlametongueTotemOffHand(t *testing.T) {
	sim, rogue := newTotemRogueSim(30, proto.TotemWeaponBuff_TotemWeaponBuffFlametongue, 120)

	buff := rogue.GetAura("Flametongue Totem Raid (Rank 1)")
	if buff == nil {
		t.Fatalf("no Flametongue Totem rank 1")
	}
	speeds := map[core.ProcMask]float64{core.ProcMaskMeleeMHAuto: 2.5, core.ProcMaskMeleeOHAuto: 1.8}
	hits := map[core.ProcMask]int{}
	var hand core.ProcMask
	onHit := buff.OnSpellHitDealt
	buff.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if spell.ActionID.SpellID != core.FlametongueTotemProcSpellId[1] {
			hand = spell.ProcMask
			onHit(aura, sim, spell, result)
			return
		}
		if result.Outcome != core.OutcomeHit {
			return
		}
		if want := 548.0 / 100 * speeds[hand]; math.Abs(result.Damage-want) > 1e-9 {
			t.Errorf("at %s: Flametongue Totem from %v hit for %.2f, want %.2f", sim.CurrentTime, hand, result.Damage, want)
		}
		hits[hand]++
	}
	for !sim.Step() {
	}

	for mask, name := range map[core.ProcMask]string{core.ProcMaskMeleeMHAuto: "main hand", core.ProcMaskMeleeOHAuto: "off hand"} {
		if hits[mask] < 20 {
			t.Errorf("only %d normal Flametongue Totem hits from the %s", hits[mask], name)
		}
	}
}

// TestOrcRogueWindfuryTotemOffHand checks that our off-hand white swings proc Windfury Totem
// under Forever.
//
// We see a proc when the totem's cooldown starts during a hit. The off hand swings about 160
// times in 5 minutes, so we expect around 20 procs from it.
func TestOrcRogueWindfuryTotemOffHand(t *testing.T) {
	sim, rogue := newTotemRogueSim(60, proto.TotemWeaponBuff_TotemWeaponBuffWindfury, 300)

	totem := rogue.GetAura("Windfury Totem Raid")
	buff := rogue.GetAura("Windfury Totem Raid Buff")
	if totem == nil || buff == nil || buff.Icd == nil {
		t.Fatalf("no Windfury Totem")
	}
	procs := map[core.ProcMask]int{}
	onHit := totem.OnSpellHitDealt
	totem.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		ready := buff.Icd.IsReady(sim)
		onHit(aura, sim, spell, result)
		if ready && !buff.Icd.IsReady(sim) {
			procs[spell.ProcMask]++
		}
	}
	for !sim.Step() {
	}

	for mask, name := range map[core.ProcMask]string{core.ProcMaskMeleeMHAuto: "main hand", core.ProcMaskMeleeOHAuto: "off hand"} {
		if procs[mask] < 8 {
			t.Errorf("only %d Windfury Totem procs from the %s", procs[mask], name)
		}
	}
}
