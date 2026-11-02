package tankwarrior

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/warrior"
)

// newForeverTank builds a level 60 Orc warrior in Defensive Stance with a sword and a shield and
// no rotation.
func newForeverTank(t *testing.T, talents string) (*core.Simulation, *warrior.Warrior) {
	return newForeverTankOf(t, talents, false)
}

// newForeverTankOf is newForeverTank, with the boss attacking the warrior when tanked.
func newForeverTankOf(t *testing.T, talents string, tanked bool) (*core.Simulation, *warrior.Warrior) {
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 991014}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: 19349} // Elementium Reinforced Bulwark
	player := &proto.Player{
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{Items: items},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec:          PlayerOptionsBasic,
	}
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	if tanked {
		raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	war := sim.Raid.Parties[0].Players[0].(*TankWarrior).GetWarrior()
	war.Stance = warrior.DefensiveStance
	return sim, war
}

// TestForeverShieldWallAndBlock checks the Forever rows (client 70425): Shield Wall takes 60% less
// damage for 12 sec on a 15 min cooldown and the global cooldown (871), less 5.5 min a point of
// Improved Shield Wall (12312), and Shield Block blocks 2 attacks over 7 sec (2565).
func TestForeverShieldWallAndBlock(t *testing.T) {
	sim, war := newForeverTank(t, P1Talents) // Improved Shield Wall 2/2

	wall := war.GetSpell(core.ActionID{SpellID: 871})
	if got, want := wall.CD.Duration, 4*time.Minute; got != want {
		t.Errorf("Shield Wall with Improved Shield Wall 2/2 has a %s cooldown, want %s", got, want)
	}
	if got := wall.DefaultCast.GCD; got != core.GCDDefault {
		t.Errorf("Shield Wall has a %s global cooldown, want %s", got, core.GCDDefault)
	}
	taken := war.PseudoStats.DamageTakenMultiplier
	if !wall.Cast(sim, war.CurrentTarget) {
		t.Fatal("Shield Wall didn't cast in Defensive Stance with a shield")
	}
	aura := war.GetAura("Shield Wall")
	if aura.Duration != 12*time.Second {
		t.Errorf("Shield Wall lasts %s, want 12s", aura.Duration)
	}
	if got := war.PseudoStats.DamageTakenMultiplier / taken; math.Abs(got-0.4) > 1e-9 {
		t.Errorf("Shield Wall leaves %.2f of the damage taken, want 0.4", got)
	}

	if got := war.ShieldBlockAura.Duration; got != 7*time.Second {
		t.Errorf("Shield Block lasts %s, want 7s", got)
	}
	if got := war.ShieldBlockAura.MaxStacks; got != 2 {
		t.Errorf("Shield Block blocks %d attacks, want 2", got)
	}
	if got := war.ShieldBlock.CD.Duration; got != 5*time.Second {
		t.Errorf("Shield Block has a %s cooldown, want 5s", got)
	}
}

// TestForeverShieldBlockRecastRefills checks that casting Shield Block while its buff is still up
// (7 sec on a 5 sec cooldown) refreshes it with both charges, as wowsims/forever's shield_block.go
// does.
func TestForeverShieldBlockRecastRefills(t *testing.T) {
	sim, war := newForeverTank(t, P1Talents)
	aura := war.ShieldBlockAura
	cast := func() {
		t.Helper()
		war.AddRage(sim, 100, war.RageRefundMetrics)
		if !war.ShieldBlock.Cast(sim, war.CurrentTarget) {
			t.Fatalf("Shield Block didn't cast at %s", sim.CurrentTime)
		}
	}

	cast()
	aura.RemoveStack(sim) // a block spends a charge
	if got := aura.GetStacks(); got != 1 {
		t.Fatalf("Shield Block has %d charges after a block, want 1", got)
	}
	sim.AddPendingAction(&core.PendingAction{NextActionAt: 5 * time.Second, OnAction: func(*core.Simulation) {}})
	for !sim.Step() && sim.CurrentTime < 5*time.Second {
	}
	if !aura.IsActive() {
		t.Fatal("Shield Block ended before its cooldown came back")
	}
	cast()
	if got := aura.GetStacks(); got != 2 {
		t.Errorf("Shield Block recast while up has %d charges, want 2", got)
	}
	if got, want := aura.ExpiresAt(), sim.CurrentTime+7*time.Second; got != want {
		t.Errorf("Shield Block recast at %s ends at %s, want %s", sim.CurrentTime, got, want)
	}
}

// TestForeverShieldSlamIsOffHand checks that under Forever Shield Slam is an off-hand special,
// struck with the shield (wowsims/forever b6ce30c972): main hand procs such as Windfury Totem and
// weapon enchants never see it, Dual Wield Specialization and Furious Precision leave it alone,
// and a procs-per-minute rate on it goes by the main hand's speed.
func TestForeverShieldSlamIsOffHand(t *testing.T) {
	_, bare := newForeverTank(t, "--251533120331221351")
	_, specced := newForeverTank(t, "-00003005-251533120331221351") // Furious Precision 3/3, Dual Wield Specialization 5/5
	slam := specced.ShieldSlam.Spell

	if slam.ProcMask != core.ProcMaskMeleeOHSpecial {
		t.Errorf("Shield Slam's proc mask is %v, want the off-hand special mask", slam.ProcMask)
	}
	if slam.CastType != proto.CastType_CastTypeMainHand {
		t.Errorf("Shield Slam uses the %v attack table, want the main hand's", slam.CastType)
	}
	if bareSlam := bare.ShieldSlam.Spell; slam.DamageMultiplier != bareSlam.DamageMultiplier || slam.BonusHitRating != bareSlam.BonusHitRating {
		t.Errorf("Shield Slam with the off-hand talents: damage %v, hit %v; want the untalented %v and %v",
			slam.DamageMultiplier, slam.BonusHitRating, bareSlam.DamageMultiplier, bareSlam.BonusHitRating)
	}

	mainHandOnly := specced.AutoAttacks.NewPPMManager(6, core.ProcMaskMeleeMH)
	if got := mainHandOnly.Chance(slam.ProcMask); got != 0 {
		t.Errorf("a main hand weapon proc rolls on Shield Slam at %v, want never", got)
	}
	anyMelee := specced.AutoAttacks.NewPPMManager(6, core.ProcMaskMelee)
	if got, want := anyMelee.Chance(slam.ProcMask), specced.AutoAttacks.MH().SwingSpeed*6/60; math.Abs(got-want) > 1e-9 {
		t.Errorf("a 6 PPM melee proc rolls on Shield Slam at %v, want the main hand's %v", got, want)
	}
}

// TestForeverTankSheet checks the tank sheet (core/tank_sheet.go) of an Orc warrior with only a
// sword and the Bulwark against a level 63 boss under Forever: the level gap takes 0.6% off each of miss, parry
// and block (0.2% a level), the boss crits only by the level gap (0.6%, so 15 Defense stops it
// instead of Classic's 140), and crushes 15%.
func TestForeverTankSheet(t *testing.T) {
	sim, war := newForeverTankOf(t, "", true)
	boss := sim.Encounter.TargetUnits[0]
	sheet := core.TankSheet(&war.Unit, boss)

	near := func(name string, got, want float64) {
		t.Helper()
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("%s is %.5f, want %.5f", name, got, want)
		}
	}
	// The Bulwark's 7 Defense: 0.28% more miss, parry and block, 0.28% less crit.
	near("Defense", sheet.Defense, 307)
	near("Miss", sheet.Miss, 0.044+0.0028)
	near("Parry", sheet.Parry, 0.044+0.0028)
	near("Block", sheet.Block, 0.044+0.0028)
	near("Crit", sheet.Crit, 0.006-0.0028)
	near("Crush", sheet.Crush, 0.15)
	near("Defense to stop crits", sheet.DefenseCritImmune, 315)
	near("Coverage", sheet.Coverage, sheet.Miss+sheet.Dodge+sheet.Parry+sheet.Block)
	near("Crush gap", sheet.CrushGap, 1-sheet.Crit-sheet.Coverage)
	near("Table total", sheet.Miss+sheet.Dodge+sheet.Parry+sheet.Block+sheet.Crit+sheet.Crush+sheet.Hit, 1)
	if sheet.Dodge <= 0 || sheet.Dodge > 0.1 {
		t.Errorf("Dodge is %.4f, want a naked warrior's few percent", sheet.Dodge)
	}
	// 1104 boss armor would be the target's; ours is the warrior's own armor from Agility.
	near("Armor reduction", sheet.ArmorReduction, sheet.Armor/(sheet.Armor+400+85*63))
	near("Effective health", sheet.EffectiveHealth, sheet.Health/(1-sheet.ArmorReduction)/war.PseudoStats.DamageTakenMultiplier)

	// 20 more Defense: crits gone, and 0.8% more of each of miss, dodge, parry and block.
	war.AddStatDynamic(sim, stats.Defense, 20)
	more := core.TankSheet(&war.Unit, boss)
	near("Crit at 327 Defense", more.Crit, 0)
	near("Defense to stop crits at 327", more.DefenseCritImmune, 315)
	near("Miss at 327 Defense", more.Miss, 0.044+0.0108)
	near("Coverage at 320 Defense", more.Coverage-sheet.Coverage, 4*0.008)

	// The test boss swings 4192.05 to 5589.3 before its 320 attack power (4192.05 * 320 / 2478 on
	// top), and at 327 Defense it can no longer crit, so its biggest swing is a crushing blow.
	top := 4192.05 * (1.3333 + 320.0/(14*177))
	near("Biggest hit", more.BiggestHit, top*1.5*(1-more.ArmorReduction)*war.PseudoStats.DamageTakenMultiplier)
	near("Biggest hit while it can crit", sheet.BiggestHit, top*2*(1-sheet.ArmorReduction)*war.PseudoStats.DamageTakenMultiplier)

	// With Classic's crit (Target.classic_crit) the boss crits 5.6% less our Defense past 300, and
	// 440 Defense stops it.
	boss.PseudoStats.ClassicCrit = true
	classic := core.TankSheet(&war.Unit, boss)
	near("Classic crit at 327 Defense", classic.Crit, 0.056-27*0.0004)
	near("Classic Defense to stop crits", classic.DefenseCritImmune, 440)
	boss.PseudoStats.ClassicCrit = false

	// Our side: 8% special miss against a level 63 boss with no hit (Forever has no hit
	// suppression), and the boss dodges and parries by its Defense over our weapon skill.
	near("Melee miss", sheet.MeleeMiss, war.AttackTables[boss.UnitIndex][proto.CastType_CastTypeMainHand].BaseMissChance)
	if sheet.BossParry < sheet.BossDodge {
		t.Errorf("the boss parries %.4f, under its dodge %.4f", sheet.BossParry, sheet.BossDodge)
	}
}

// TestArmorCap checks that armor takes off at most 75% of a boss's hit, at 17,265 armor against a
// level 63 boss (400 + 85 * 63 = 5,755, and 17,265 / (17,265 + 5,755) = 0.75).
func TestArmorCap(t *testing.T) {
	sim, war := newForeverTank(t, "")
	boss := sim.Encounter.TargetUnits[0]
	for _, c := range []struct{ armor, want float64 }{{5755, 0.5}, {17265, 0.75}, {30000, 0.75}} {
		war.AddStatDynamic(sim, stats.Armor, c.armor-war.Armor())
		if got := core.TankSheet(&war.Unit, boss).ArmorReduction; math.Abs(got-c.want) > 1e-6 {
			t.Errorf("%.0f armor takes off %.4f, want %.4f", c.armor, got, c.want)
		}
	}
}
