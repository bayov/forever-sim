package tankwarrior

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/warrior"
)

// newLevelingTank is an Undead Protection warrior with a sword and the given off hand at the start of
// a leveling session.
func newLevelingTank(t *testing.T, level int32, talents string, offHand int32) (*core.Simulation, *warrior.Warrior) {
	t.Helper()
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 991014}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: offHand}
	player := &proto.Player{
		Race:            proto.Race_RaceUndead,
		Class:           proto.Class_ClassWarrior,
		Level:           level,
		Equipment:       &proto.EquipmentSpec{Items: items},
		TalentsString:   talents,
		Rotation:        &proto.APLRotation{},
		InFrontOfTarget: true,
		Spec:            defensiveOptions,
	}
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Leveling: &proto.LevelingOptions{MaxPackSize: 10}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*TankWarrior).GetWarrior()
}

// TestLevelingTank checks what a Protection warrior alone has by level: Defensive Stance at 10 (armor
// and all in Battle Stance before), Shield Block at 16, Shield Wall at 28, and its talents.
func TestLevelingTank(t *testing.T) {
	for _, c := range []struct {
		level                   int32
		stance                  warrior.Stance
		shieldBlock, shieldWall bool
	}{
		{1, warrior.BattleStance, false, false},
		{9, warrior.BattleStance, false, false},
		{10, warrior.DefensiveStance, false, false},
		{15, warrior.DefensiveStance, false, false},
		{16, warrior.DefensiveStance, true, false},
		{27, warrior.DefensiveStance, true, false},
		{28, warrior.DefensiveStance, true, true},
	} {
		sim, war := newLevelingTank(t, c.level, "", bulwark)
		wall := war.GetSpell(core.ActionID{SpellID: 871})
		if war.Stance != c.stance || (war.ShieldBlock != nil) != c.shieldBlock || (wall != nil) != c.shieldWall {
			t.Errorf("level %d: stance %d, Shield Block %v, Shield Wall %v", c.level, war.Stance, war.ShieldBlock != nil, wall != nil)
		}
		want := 1.0
		if c.stance == warrior.DefensiveStance {
			want = 0.9
		}
		near(t, "damage taken", war.PseudoStats.DamageTakenMultiplier, want, 1e-9)
		if c.level >= 4 {
			advance(sim, 5*time.Second)
		}
	}

	// Anticipation 1/3, Deflection 1/5 and Shield Specialization 1/5 at level 12. Each point of
	// Anticipation's Defense is another 0.04% parry and block.
	_, bare := newLevelingTank(t, 12, "", bulwark)
	_, talented := newLevelingTank(t, 12, "01-0-0101", bulwark)
	near(t, "Anticipation's Defense", talented.GetStat(stats.Defense)-bare.GetStat(stats.Defense), 4, 1e-9)
	near(t, "Deflection's parry", talented.GetStat(stats.Parry)-bare.GetStat(stats.Parry), 1.16, 1e-9)
	near(t, "Shield Specialization's block", talented.GetStat(stats.Block)-bare.GetStat(stats.Block), 1.16, 1e-9)
}

// TestLevelingShieldBlockManyMobs: Shield Block's 2 blocks go to whichever mobs hit first, and Last
// Stand adds 30% of max health. The block that spends the last charge leaves it up to the end of its
// spell batch, so a third mob hitting at the same moment is blocked too.
func TestLevelingShieldBlockManyMobs(t *testing.T) {
	sim, war := newLevelingTank(t, 40, "--0000001", bulwark)
	war.AddRage(sim, 100, war.RageRefundMetrics)
	if !war.ShieldBlock.Cast(sim, war.CurrentTarget) {
		t.Fatal("Shield Block isn't castable")
	}
	for i, mob := range sim.Encounter.TargetUnits[:3] {
		result := mob.AutoAttacks.MHAuto().NewResult(&war.Unit)
		result.Outcome = core.OutcomeBlock
		result.Damage = 1
		mob.AutoAttacks.MHAuto().DealDamage(sim, result)
		if !war.ShieldBlockAura.IsActive() {
			t.Errorf("after %d blocks at once Shield Block is down", i+1)
		}
	}
	advance(sim, core.SpellBatchWindow+time.Millisecond)
	if war.ShieldBlockAura.IsActive() {
		t.Error("Shield Block is up after its spell batch, with both blocks spent")
	}
	health := war.MaxHealth()
	last := war.GetSpell(core.ActionID{SpellID: 12975})
	if !last.Cast(sim, war.CurrentTarget) {
		t.Fatal("Last Stand isn't castable")
	}
	near(t, "Last Stand's health", war.MaxHealth(), 1.3*health, 0.01)
}

// TestLevelingUnbridledWrathExceptions: Victory Rush and Retaliation's strikes give no Unbridled Wrath
// Rage (wowsims/forever EV-0148).
func TestLevelingUnbridledWrathExceptions(t *testing.T) {
	sim, war := newLevelingTank(t, 40, "-0005", bulwark)
	mob := war.CurrentTarget
	for _, spellID := range []int32{402927, 20240} {
		spell := war.GetSpell(core.ActionID{SpellID: spellID})
		war.SpendRage(sim, war.CurrentRage(), war.RageRefundMetrics)
		for range 50 {
			result := spell.NewResult(mob)
			result.Outcome = core.OutcomeHit
			result.Damage = 1
			spell.DealDamage(sim, result)
		}
		if war.CurrentRage() > 0 {
			t.Errorf("50 hits of %d gave %.0f Rage", spellID, war.CurrentRage())
		}
	}
}
