package tankwarrior

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/warrior"
)

// newWindfuryTank builds newForeverTank's warrior under the given ruleset, with the raid's
// Windfury Totem.
func newWindfuryTank(t *testing.T, ruleset proto.Ruleset) (*core.Simulation, *warrior.Warrior) {
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 991014}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: 19349}
	player := &proto.Player{
		Race:      proto.Race_RaceOrc,
		Class:     proto.Class_ClassWarrior,
		Level:     60,
		Equipment: &proto.EquipmentSpec{Items: items},
		Rotation:  &proto.APLRotation{},
		Spec:      PlayerOptionsBasic,
	}
	buffs := &proto.RaidBuffs{TotemWeaponBuff: proto.TotemWeaponBuff_TotemWeaponBuffWindfury}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, buffs, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 600, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: ruleset, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*TankWarrior).GetWarrior()
}

// windfuryProcs lands the spell n times, 2 sec apart (past either ruleset's internal cooldown),
// and counts the totem's procs (its attack power buff going up).
func windfuryProcs(sim *core.Simulation, war *warrior.Warrior, spell *core.Spell, n int) int {
	buff := war.GetAura("Windfury Attack")
	procs := 0
	buff.ApplyOnGain(func(*core.Aura, *core.Simulation) { procs++ })
	for i := 0; i < n; i++ {
		sim.CurrentTime += 2 * time.Second
		buff.Deactivate(sim)
		spell.ApplyEffects(sim, war.CurrentTarget, spell)
	}
	return procs
}

// TestForeverWindfuryTotemMainHandAbilities checks Blizzard's hotfixes of November 13, 2026
// (t/2403040): "Windfury Totem is now correctly activated by all melee abilities ... Strikes from
// abilities that involve your offhand weapon like Whirlwind or Mutilate do not have a chance to
// activate Windfury. The main-hand strike from those abilities still does." Under Forever every
// main hand warrior ability (Heroic Strike, Cleave, Sunder Armor, Revenge, Execute, Rend and
// Whirlwind's main hand strike; this tank holds a shield) rolls Windfury Totem at 20%; off-hand
// strikes never do (sim/core/windfury_totem_rolls_forever_test.go). Execute's cast rolls it once
// more before its damage (wowsims/forever EV-0155, after push 114313; sim/warrior/execute.go), so
// about a third of its casts proc it. Classic keeps every main hand hit. White swings still proc
// it (sim/shaman/enhancement/proc_matrix_test.go).
func TestForeverWindfuryTotemMainHandAbilities(t *testing.T) {
	const n = 400
	sim, war := newWindfuryTank(t, proto.Ruleset_RulesetForever)
	if war.RaidTotemWeaponBuff == nil || war.GetAura("Windfury Attack") == nil {
		t.Fatal("no Windfury Totem on the Forever warrior")
	}
	for _, s := range []*warrior.WarriorSpell{war.HeroicStrike, war.Cleave, war.SunderArmor, war.Revenge, war.Rend} {
		if s == nil {
			t.Fatal("a warrior ability isn't registered")
		}
		if got := windfuryProcs(sim, war, s.Spell, n); got < n/10 || got > n/3 {
			t.Errorf("Forever %s procs Windfury Totem %d times in %d, want about a fifth", s.ActionID, got, n)
		}
	}
	if got := windfuryProcs(sim, war, war.Execute.Spell, n); got < n/4 || got > 2*n/5 {
		t.Errorf("Forever Execute procs Windfury Totem %d times in %d, want about a third (two rolls)", got, n)
	}
	if got := windfuryProcs(sim, war, war.Whirlwind.Spell, n); got < n/10 || got > n/3 {
		t.Errorf("Forever Whirlwind's main hand strike procs Windfury Totem %d times in %d, want about a fifth", got, n)
	}

	era, eraWar := newWindfuryTank(t, proto.Ruleset_RulesetClassic)
	for _, s := range []*warrior.WarriorSpell{eraWar.HeroicStrike, eraWar.Whirlwind} {
		if got := windfuryProcs(era, eraWar, s.Spell, n); got < n/10 {
			t.Errorf("Classic %s procs Windfury Totem %d times in %d, want about a fifth", s.ActionID, got, n)
		}
	}
}
