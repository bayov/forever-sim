package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanFlametongueTotem checks the fire damage that Flametongue Totem adds to our
// main hand hits, for a level 30 Orc shaman with Rockbiter Weapon and 55 spell power.
//
// Rank 1 adds 548 / 100 per second of weapon speed, so 19.18 with Dark Edge of Insanity (3.5
// speed), and spell power adds nothing. On the beta the same shaman with 55 spell power hit 20
// on a 3.6 speed axe (19.73) (2026-10-08). The target is level 30, so a normal hit isn't
// partly resisted.
func TestOrcShamanFlametongueTotem(t *testing.T) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 21134} // Dark Edge of Insanity

	player := newOrcShaman(30, "", &proto.EnhancementShaman_Options{ShamanImbue: proto.WeaponImbue_RockbiterWeapon})
	player.Equipment = &proto.EquipmentSpec{Items: items}
	player.BonusStats = &proto.UnitStats{Stats: stats.Stats{stats.SpellPower: 55}.ToFloatArray()}

	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 300, Targets: []*proto.Target{{Level: 30, MobType: proto.MobType_MobTypeBeast}}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)

	if sp := enh.GetStat(stats.SpellPower); sp != 55 {
		t.Fatalf("spell power is %.0f, want 55", sp)
	}
	totem := enh.FlametongueTotem[1]
	if totem == nil || enh.FlametongueTotem[2] != nil {
		t.Fatalf("want only rank 1 of Flametongue Totem at level 30")
	}
	at(sim, 0, func(sim *core.Simulation) { castNow(t, sim, enh, totem) })

	want := 548.0 / 100 * 3.5
	hits := 0
	buff := enh.GetAura("Flametongue Totem (Rank 1)")
	onHit := buff.OnSpellHitDealt
	buff.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		onHit(aura, sim, spell, result)
		if spell.ActionID.SpellID != core.FlametongueTotemProcSpellId[1] || result.Outcome != core.OutcomeHit {
			return
		}
		if math.Abs(result.Damage-want) > 1e-9 {
			t.Errorf("at %s: Flametongue Totem hit for %.2f, want %.2f", sim.CurrentTime, result.Damage, want)
		}
		hits++
	}
	runSim(sim)

	if hits < 30 {
		t.Errorf("only %d normal Flametongue Totem hits", hits)
	}
}

// TestOrcShamanFlametongueTotemCrits checks that Flametongue Totem's proc crits at our spell
// crit under Forever, as the beta showed.
//
// We proc it 1000 times with 100% spell crit and 1000 times with none. The target is level
// 30, so 4% of the procs miss, and we expect about 960 crits and then none.
func TestOrcShamanFlametongueTotemCrits(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(newOrcShaman(30, "", &proto.EnhancementShaman_Options{}), &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 10, Targets: []*proto.Target{{Level: 30, MobType: proto.MobType_MobTypeBeast}}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)

	proc := enh.GetSpell(core.ActionID{SpellID: core.FlametongueTotemProcSpellId[1]})
	if proc == nil {
		t.Fatalf("no Flametongue Totem rank 1 proc")
	}
	metrics := &proc.SpellMetrics[enh.CurrentTarget.UnitIndex]

	procs := func(sim *core.Simulation) int32 {
		crits := metrics.Crits
		for range 1000 {
			proc.Cast(sim, enh.CurrentTarget)
		}
		return metrics.Crits - crits
	}
	allCrit := 100.0 * core.SpellCritRatingPerCritChance
	var withCrit, withoutCrit int32
	at(sim, 1, func(sim *core.Simulation) {
		enh.AddStatDynamic(sim, stats.SpellCrit, allCrit)
		withCrit = procs(sim)
		enh.AddStatDynamic(sim, stats.SpellCrit, -2*allCrit)
		withoutCrit = procs(sim)
	})
	runSim(sim)

	// 960 crits give or take 6.
	if withCrit < 930 || withoutCrit != 0 {
		t.Errorf("got %d crits in 1000 procs with 100%% spell crit and %d with none, want about 960 and 0", withCrit, withoutCrit)
	}
}
