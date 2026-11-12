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

// TestForeverWindfuryTotemMainHandAbilities checks standards 17 (user ruling 2026-11-12, the
// log audit in tools/logcheck/findings.md): under Forever the warrior abilities whose client spell
// requires a main hand weapon or is on next swing (Heroic Strike, Cleave, Sunder Armor, Revenge,
// Execute, Rend) roll Windfury Totem at 20%, and Whirlwind, whose spell isn't flagged, never does.
// Classic keeps every main hand hit. White swings still proc it
// (sim/shaman/enhancement/proc_matrix_test.go); sim/core/windfury_totem_procs_forever_test.go
// pins the log list.
func TestForeverWindfuryTotemMainHandAbilities(t *testing.T) {
	const n = 400
	sim, war := newWindfuryTank(t, proto.Ruleset_RulesetForever)
	if war.RaidTotemWeaponBuff == nil || war.GetAura("Windfury Attack") == nil {
		t.Fatal("no Windfury Totem on the Forever warrior")
	}
	for _, s := range []*warrior.WarriorSpell{war.HeroicStrike, war.Cleave, war.SunderArmor, war.Revenge, war.Execute, war.Rend} {
		if s == nil {
			t.Fatal("a warrior ability isn't registered")
		}
		if got := windfuryProcs(sim, war, s.Spell, n); got < n/10 || got > n/3 {
			t.Errorf("Forever %s procs Windfury Totem %d times in %d, want about a fifth", s.ActionID, got, n)
		}
	}
	if got := windfuryProcs(sim, war, war.Whirlwind.Spell, n); got != 0 {
		t.Errorf("Forever Whirlwind procs Windfury Totem %d times in %d, want never", got, n)
	}

	era, eraWar := newWindfuryTank(t, proto.Ruleset_RulesetClassic)
	for _, s := range []*warrior.WarriorSpell{eraWar.HeroicStrike, eraWar.Whirlwind} {
		if got := windfuryProcs(era, eraWar, s.Spell, n); got < n/10 {
			t.Errorf("Classic %s procs Windfury Totem %d times in %d, want about a fifth", s.ActionID, got, n)
		}
	}
}
