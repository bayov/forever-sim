package tankwarrior

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	googleProto "google.golang.org/protobuf/proto"
)

// The engine's rolls on a tank against its character sheet (tank_mechanics-20261104-11).

// outcomeRun runs a tanked warrior with no rotation, a sword and the Elementium Reinforced Bulwark
// against the default level 63 boss under Forever's rules, and returns the boss's outcomes with the
// spell (0 for its white swing) and the warrior's tank sheet.
func outcomeRun(t *testing.T, target *proto.Target, raidDamage []*proto.RaidDamageEvent, spellId int32, iterations int32) (*proto.TargetedActionMetrics, *proto.TankSheet) {
	t.Helper()
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
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	res := core.RunRaidSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Duration: 300, Targets: []*proto.Target{target}, RaidDamage: raidDamage},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 3, Iterations: iterations},
	})
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	got := &proto.TargetedActionMetrics{}
	for _, a := range res.EncounterMetrics.Targets[0].Actions {
		if spellId == 0 && a.Id.GetOtherId() != proto.OtherAction_OtherActionAttack || spellId != 0 && a.Id.GetSpellId() != spellId {
			continue
		}
		for _, tm := range a.Targets {
			got.Casts += tm.Casts
			got.Misses += tm.Misses
			got.Dodges += tm.Dodges
			got.Parries += tm.Parries
			got.Blocks += tm.Blocks
			got.Hits += tm.Hits
			got.Crushes += tm.Crushes
			got.Crits += tm.Crits
			got.BlockedCrits += tm.BlockedCrits
		}
	}
	sheet := res.RaidMetrics.Parties[0].Players[0].TankSheet
	if sheet == nil {
		t.Fatal("no tank sheet")
	}
	return got, sheet
}

// near checks a rolled count against n tries at chance p, within 4 binomial standard errors.
func nearShare(t *testing.T, name string, count int32, n int32, p float64) {
	t.Helper()
	want := float64(n) * p
	if sd := math.Sqrt(float64(n) * p * (1 - p)); math.Abs(float64(count)-want) > 4*sd+1e-9 {
		t.Errorf("%s: %d of %d, the sheet's %.4f gives %.0f ± %.0f", name, count, n, p, want, sd)
	}
}

// The boss's white swings land on the warrior's sheet: each outcome's share of the swings is the
// sheet's, crushing blows included (15% from a boss three levels up), from the same attack table
// (core/tank_sheet.go against outcomeEnemyMeleeWhite).
func TestForeverRolledOutcomesMatchSheet(t *testing.T) {
	got, sheet := outcomeRun(t, core.NewDefaultTarget(), nil, 0, 50)
	if sheet.Crush != 0.15 {
		t.Errorf("the sheet's crush is %.4f, want 15%% from a level 63 boss", sheet.Crush)
	}
	n := got.Misses + got.Dodges + got.Parries + got.Blocks + got.Crits + got.Crushes + got.Hits
	if n < 5000 {
		t.Fatalf("only %d boss swings", n)
	}
	nearShare(t, "Crushes", got.Crushes, n, sheet.Crush)
	nearShare(t, "Misses", got.Misses, n, sheet.Miss)
	nearShare(t, "Dodges", got.Dodges, n, sheet.Dodge)
	nearShare(t, "Parries", got.Parries, n, sheet.Parry)
	nearShare(t, "Blocks", got.Blocks, n, sheet.Block)
	nearShare(t, "Crits", got.Crits, n, sheet.Crit)
	nearShare(t, "Hits", got.Hits, n, sheet.Hit)
}

// Block value is the shield's (and gear's) block value times BlockValueMultiplier, plus Strength
// over BlockValuePerStrength less 1 and never under 0, outside the multiplier.
func TestBlockValueFormula(t *testing.T) {
	sim, war := newForeverTank(t, "")
	shield := war.GetStat(stats.BlockValue)
	if shield <= 0 {
		t.Fatalf("the Bulwark gives %.0f block value", shield)
	}
	perStrength := war.PseudoStats.BlockValuePerStrength
	if perStrength != 1.0/20 {
		t.Errorf("block value per Strength is %.4f, want 1/20", perStrength)
	}
	want := func() float64 {
		return max(0, war.GetStat(stats.Strength)*perStrength-1) + war.PseudoStats.BlockValueMultiplier*war.GetStat(stats.BlockValue)
	}
	if got := war.BlockValue(); math.Abs(got-want()) > 1e-9 {
		t.Errorf("block value %.3f, want %.3f", got, want())
	}
	// 200 more Strength: 10 more block value, untouched by a 30% multiplier.
	before := war.BlockValue()
	war.AddStatDynamic(sim, stats.Strength, 200)
	if got := war.BlockValue() - before; math.Abs(got-10) > 1e-9 {
		t.Errorf("200 Strength gives %.3f block value, want 10", got)
	}
	before = war.BlockValue()
	war.PseudoStats.BlockValueMultiplier *= 1.3
	if got := war.BlockValue() - before; math.Abs(got-0.3*shield) > 1e-9 {
		t.Errorf("a 30%% multiplier gives %.3f block value, want 30%% of the shield's %.0f", got, shield)
	}
	// Under 20 Strength the Strength part is 0, not negative.
	war.AddStatDynamic(sim, stats.Strength, -war.GetStat(stats.Strength)+10)
	if got := war.BlockValue(); math.Abs(got-war.PseudoStats.BlockValueMultiplier*war.GetStat(stats.BlockValue)) > 1e-9 {
		t.Errorf("at 10 Strength block value is %.3f, want the shield's alone", got)
	}
}

// A boss's melee ability (OutcomeEnemyMeleeSpecial) under Forever's crit: a level 63 boss crits a
// level 60 tank 0.6% less 0.04% a point of bonus Defense, rolled on what isn't missed, dodged or
// parried, as the sheet says; Classic's 5.6% would be ten times as many. It never crushes.
func TestForeverBossMeleeAbilityForeverCrit(t *testing.T) {
	target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
	target.SwingSpeed = 100
	got, sheet := outcomeRun(t, target, []*proto.RaidDamageEvent{{
		SpellId: 24573, School: proto.SpellSchool_SpellSchoolPhysical, MinDamage: 2000, MaxDamage: 2000,
		Who: proto.RaidDamageEvent_BossTarget, EverySeconds: 1, Melee: true,
	}}, 24573, 100)
	// The Bulwark's 7 Defense.
	if want := 0.006 - 7*0.0004; math.Abs(sheet.Crit-want) > 1e-9 {
		t.Errorf("the sheet's crit is %.4f, want Forever's %.4f", sheet.Crit, want)
	}
	rolled := got.Blocks + got.Hits + got.Crits + got.BlockedCrits
	if rolled < 20000 {
		t.Fatalf("only %d of the ability's uses rolled for a crit", rolled)
	}
	nearShare(t, "Crits and blocked crits", got.Crits+got.BlockedCrits, rolled, sheet.Crit)
	if got.Crushes != 0 {
		t.Errorf("the ability crushed %d times, want none", got.Crushes)
	}
}
