package dpswarrior

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

// TestForeverWindfuryTotemFurySpecials checks Blizzard's hotfixes of November 13, 2026
// (t/2403040) on a dual wielding Fury warrior: "Windfury Totem is now correctly activated by all
// melee abilities ... Strikes from abilities that involve your offhand weapon like Whirlwind or
// Mutilate do not have a chance to activate Windfury. The main-hand strike from those abilities
// still does." Bloodthirst and Whirlwind's main hand strike roll it at 20%, Whirlwind's off-hand
// strike and the off-hand white swing never do (standards 17).
func TestForeverWindfuryTotemFurySpecials(t *testing.T) {
	const n = 400
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 991014}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: 991015}
	player := &proto.Player{
		Race:          proto.Race_RaceTauren,
		Class:         proto.Class_ClassWarrior,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{Items: items},
		TalentsString: P1Talents, // Bloodthirst
		Rotation:      &proto.APLRotation{},
		Spec:          &proto.Player_Warrior{Warrior: &proto.Warrior{Options: &proto.Warrior_Options{}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(player, &proto.PartyBuffs{},
			&proto.RaidBuffs{TotemWeaponBuff: proto.TotemWeaponBuff_TotemWeaponBuffWindfury}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 600, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	war := sim.Raid.Parties[0].Players[0].(*DpsWarrior).GetWarrior()
	totem, buff := war.GetAura("Windfury Totem Raid"), war.GetAura("Windfury Attack")
	if totem == nil || buff == nil || war.Bloodthirst == nil || war.Whirlwind == nil {
		t.Fatal("no Windfury Totem, Bloodthirst or Whirlwind on the warrior")
	}

	for _, c := range []struct {
		name  string
		spell *core.Spell
		rolls bool
	}{
		{"Bloodthirst", war.Bloodthirst.Spell, true},
		{"Whirlwind's main hand strike", war.Whirlwind.Spell, true},
		{"Whirlwind's off-hand strike", war.GetSpell(core.ActionID{SpellID: 1680}.WithTag(2)), false},
		{"off-hand white swing", war.AutoAttacks.OHAuto(), false},
	} {
		if c.spell == nil {
			t.Fatalf("%s isn't registered", c.name)
		}
		procs := 0
		for range n {
			sim.CurrentTime += 2 * time.Second
			buff.Deactivate(sim)
			result := c.spell.NewResult(war.CurrentTarget)
			result.Outcome = core.OutcomeHit
			totem.OnSpellHitDealt(totem, sim, c.spell, result)
			if buff.IsActive() {
				procs++
			}
		}
		if c.rolls && (procs < n/10 || procs > n/3) {
			t.Errorf("%s procced Windfury Totem %d times in %d landed hits, want about a fifth", c.name, procs, n)
		}
		if !c.rolls && procs != 0 {
			t.Errorf("%s procced Windfury Totem %d times in %d landed hits, want none", c.name, procs, n)
		}
	}
}
