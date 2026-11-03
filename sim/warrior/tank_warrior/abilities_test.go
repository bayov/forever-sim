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
	googleProto "google.golang.org/protobuf/proto"
)

// The tank warrior's abilities and talents under Forever (docs/tank_warrior.md, Rules audit): each
// ability's cost, cooldown, duration and threat, and each talent on against off. The damage rolls of
// Revenge, Shield Slam and Thunder Clap are in dps_warrior's TestForeverWarriorProtectionValues.

const bulwark = 19349 // Elementium Reinforced Bulwark

// newTank is a level 60 Orc warrior in Defensive Stance with a sword and the given off hand (0 for
// none), no rotation, against the default level 63 boss.
func newTank(t *testing.T, talents string, offHand int32) (*core.Simulation, *warrior.Warrior) {
	t.Helper()
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 991014}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: offHand}
	player := &proto.Player{
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{Items: items},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec:          defensiveOptions,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 300, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	war := sim.Raid.Parties[0].Players[0].(*TankWarrior).GetWarrior()
	war.AutoAttacks.CancelAutoSwing(sim)
	return sim, war
}

// ready clears a spell's cooldown and the GCD and tops the Rage up.
func ready(sim *core.Simulation, war *warrior.Warrior, spell *core.Spell) {
	if spell.CD.Timer != nil {
		spell.CD.Reset()
	}
	war.AddRage(sim, 100, war.RageRefundMetrics)
	war.GCD.Reset() // after the Rage, which wakes the rotation up
}

// advance runs the sim's clock forward by d.
func advance(sim *core.Simulation, d time.Duration) {
	end := sim.CurrentTime + d
	sim.AddPendingAction(&core.PendingAction{NextActionAt: end, OnAction: func(*core.Simulation) {}})
	for sim.CurrentTime < end && !sim.Step() {
	}
}

func near(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s is %.4f, want %.4f", name, got, want)
	}
}

// The P1 build: Defensive Stance with Defiance 3/3 and a shield is 1.3 x 1.15.
const p1Stance = 1.3 * 1.15

// TestForeverTankAbilityCosts checks the Rage costs and cooldowns of the tank's buttons against
// client 70753, with the P1 talents (Improved Thunder Clap 3/3, Improved Sunder Armor 3/3, Focused
// Rage 3/3, Improved Heroic Strike 3/3) and with none.
func TestForeverTankAbilityCosts(t *testing.T) {
	_, bare := newTank(t, "", bulwark)
	_, p1 := newTank(t, P1Talents+"", bulwark)
	cases := []struct {
		name      string
		spell     func(*warrior.Warrior) *core.Spell
		bare, p1  float64
		cooldown  time.Duration
		gcd       bool
		needsTree bool
	}{
		{"Shield Block (2565)", func(w *warrior.Warrior) *core.Spell { return w.ShieldBlock.Spell }, 10, 10, 5 * time.Second, false, false},
		{"Shield Slam (23925)", func(w *warrior.Warrior) *core.Spell { return w.ShieldSlam.Spell }, 20, 17, 6 * time.Second, true, true},
		{"Revenge (11601)", func(w *warrior.Warrior) *core.Spell { return w.Revenge.Spell }, 5, 2, 5 * time.Second, true, false},
		{"Sunder Armor (11597)", func(w *warrior.Warrior) *core.Spell { return w.SunderArmor.Spell }, 15, 9, 0, true, false},
		{"Thunder Clap (11581)", func(w *warrior.Warrior) *core.Spell { return w.ThunderClap.Spell }, 20, 11, 6 * time.Second, true, false},
		{"Demoralizing Shout (11556)", func(w *warrior.Warrior) *core.Spell { return w.DemoralizingShout.Spell }, 10, 7, 0, true, false},
		{"Heroic Strike (11567)", func(w *warrior.Warrior) *core.Spell { return w.HeroicStrike.Spell }, 15, 9, 0, false, false},
		{"Cleave (20569)", func(w *warrior.Warrior) *core.Spell { return w.Cleave.Spell }, 20, 17, 0, false, false},
	}
	for _, c := range cases {
		for _, w := range []struct {
			war  *warrior.Warrior
			want float64
		}{{bare, c.bare}, {p1, c.p1}} {
			if c.needsTree && w.war == bare {
				continue
			}
			spell := c.spell(w.war)
			if got := spell.Cost.GetCurrentCost(); math.Abs(got-w.want) > 1e-9 {
				t.Errorf("%s costs %.1f Rage with talents %q, want %.1f", c.name, got, w.war.Talents.String(), w.want)
			}
			if spell.CD.Duration != c.cooldown {
				t.Errorf("%s has a %s cooldown, want %s", c.name, spell.CD.Duration, c.cooldown)
			}
			if got := spell.DefaultCast.GCD > 0; got != c.gcd {
				t.Errorf("%s on the GCD: %v, want %v", c.name, got, c.gcd)
			}
		}
	}
	if bare.ShieldSlam != nil {
		t.Error("Shield Slam is there without the talent")
	}
}

// TestForeverTankShieldBlockSpends checks that a Shield Block cast takes its 10 Rage and gives 75%
// block, which the tank sheet shows.
func TestForeverTankShieldBlockSpends(t *testing.T) {
	sim, war := newTank(t, P1Talents, bulwark)
	before := war.GetStat(stats.Block)
	ready(sim, war, war.ShieldBlock.Spell)
	if !war.ShieldBlock.Cast(sim, war.CurrentTarget) {
		t.Fatal("Shield Block didn't cast")
	}
	near(t, "Rage after Shield Block", war.CurrentRage(), 90, 1e-9)
	near(t, "Block rating from Shield Block", war.GetStat(stats.Block)-before, 75*core.BlockRatingPerBlockChance, 1e-9)
}

// TestForeverTankThreat checks each threat button's threat per landed hit in Defensive Stance with
// Defiance 3/3 (x1.3 x1.15): Sunder Armor 206 + 5% of attack power (client E_THREAT, wowsims/forever's
// capture), Heroic Strike its damage + 2.54 a spell level (142.24), Cleave + 1.53 a level (91.8),
// Revenge 2.25 x its damage, Thunder Clap 2.5 x (Classic's), Shield Slam its damage + 508 (the sim's,
// unconfirmed), Demoralizing Shout 47 (client E_THREAT at 60).
func TestForeverTankThreat(t *testing.T) {
	sim, war := newTank(t, P1Talents, bulwark)
	war.AddStatDynamic(sim, stats.AttackPower, 1000)
	near(t, "Defensive Stance's threat multiplier with Defiance 3/3", war.PseudoStats.ThreatMultiplier, p1Stance, 1e-9)
	ap := war.ShieldSlam.MeleeAttackPower(war.CurrentTarget)

	cases := []struct {
		name       string
		spell      *core.Spell
		multiplier float64
		flat       float64
	}{
		{"Sunder Armor", war.SunderArmor.Spell, 0, 206 + 0.05*ap},
		{"Heroic Strike", war.HeroicStrike.Spell, 1, 2.54 * 56},
		{"Cleave", war.Cleave.Spell, 1, 1.53 * 60},
		{"Revenge", war.Revenge.Spell, 2.25, 0},
		{"Thunder Clap", war.ThunderClap.Spell, 2.5, 0},
		{"Shield Slam", war.ShieldSlam.Spell, 1, 508},
		{"Demoralizing Shout", war.DemoralizingShout.Spell, 0, 47},
	}
	target := war.CurrentTarget
	for _, c := range cases {
		metrics := &c.spell.SpellMetrics[target.UnitIndex]
		landed := 0
		for i := 0; i < 200; i++ {
			before := *metrics
			c.spell.ApplyEffects(sim, target, c.spell)
			threat, damage := metrics.TotalThreat-before.TotalThreat, metrics.TotalDamage-before.TotalDamage
			if threat == 0 {
				continue // missed, dodged or parried
			}
			landed++
			want := (c.multiplier*damage + c.flat) * p1Stance
			// Damage lands in whole points, and the threat reads it before the rounding.
			if math.Abs(threat-want) > c.multiplier*p1Stance+1e-6*want {
				t.Errorf("%s: %.0f damage made %.2f threat, want %.2f", c.name, damage, threat, want)
				break
			}
		}
		if landed < 50 {
			t.Errorf("%s landed %d times of 200", c.name, landed)
		}
	}
}

// TestForeverTankDefensiveStance checks Defensive Stance (71, passive 7376): 10% less damage taken
// and dealt, 30% more threat, and Defiance's 5% a point only with a shield (12792 names the shield),
// multiplied in as the tooltip's formula does.
func TestForeverTankDefensiveStance(t *testing.T) {
	_, none := newTank(t, "", bulwark)
	_, shield := newTank(t, "--000000000300000000", bulwark)
	_, noShield := newTank(t, "--000000000300000000", 0)
	near(t, "threat with no Defiance", none.PseudoStats.ThreatMultiplier, 1.3, 1e-9)
	near(t, "threat with Defiance 3/3 and a shield", shield.PseudoStats.ThreatMultiplier, 1.3*1.15, 1e-9)
	near(t, "threat with Defiance 3/3 and no shield", noShield.PseudoStats.ThreatMultiplier, 1.3, 1e-9)
	near(t, "damage taken", none.PseudoStats.DamageTakenMultiplier, 0.9, 1e-9)
	near(t, "damage dealt", none.PseudoStats.DamageDealtMultiplier, 0.9, 1e-9)
}

// TestForeverTankDebuffs checks the tank's own debuffs: Thunder Clap slows the boss's attacks by 20%
// for 30 sec, Demoralizing Shout takes 204 attack power for 45 sec, and Sunder Armor takes 450 armor
// a stack up to 5 for 30 sec (client 70753: 11581, 11556, 11597).
func TestForeverTankDebuffs(t *testing.T) {
	sim, war := newTank(t, P1Talents, bulwark)
	boss := war.CurrentTarget
	war.AddStatDynamic(sim, stats.MeleeHit, 100*core.MeleeHitRatingPerHitChance)
	war.AddStatDynamic(sim, stats.SpellHit, 100*core.SpellHitRatingPerHitChance)

	speed := boss.PseudoStats.MeleeSpeedMultiplier
	war.ThunderClap.ApplyEffects(sim, boss, war.ThunderClap.Spell)
	near(t, "the boss's attack speed under Thunder Clap", boss.PseudoStats.MeleeSpeedMultiplier/speed, 1/1.2, 1e-9)
	if got := war.ThunderClapAuras.Get(boss).Duration; got != 30*time.Second {
		t.Errorf("Thunder Clap's slow lasts %s, want 30s", got)
	}

	ap := boss.GetStat(stats.AttackPower)
	war.DemoralizingShout.ApplyEffects(sim, boss, war.DemoralizingShout.Spell)
	if !war.DemoralizingShoutAuras.Get(boss).IsActive() {
		t.Fatal("Demoralizing Shout didn't land")
	}
	near(t, "the boss's attack power cut", ap-boss.GetStat(stats.AttackPower), 204, 1)

	armor := boss.GetStat(stats.Armor)
	for i := 0; i < 20; i++ {
		ready(sim, war, war.SunderArmor.Spell)
		war.SunderArmor.Cast(sim, boss)
	}
	sunder := war.SunderArmorAuras.Get(boss)
	if sunder.GetStacks() != 5 || sunder.Duration != 30*time.Second {
		t.Errorf("Sunder Armor has %d stacks for %s, want 5 for 30s", sunder.GetStacks(), sunder.Duration)
	}
	near(t, "armor Sunder Armor takes at 5", armor-boss.GetStat(stats.Armor), 5*450, 1e-6)
}

// TestForeverTankRevengeWindow checks that Revenge needs a block, dodge or parry (Defensive State,
// 5302: 5 sec, one use) and Defensive Stance.
func TestForeverTankRevengeWindow(t *testing.T) {
	sim, war := newTank(t, P1Talents, bulwark)
	ready(sim, war, war.Revenge.Spell)
	if war.Revenge.Cast(sim, war.CurrentTarget) {
		t.Fatal("Revenge cast with no block, dodge or parry")
	}
	window := war.GetAura("Revenge")
	window.Activate(sim)
	if window.Duration != 5*time.Second {
		t.Errorf("Revenge's window lasts %s, want 5s", window.Duration)
	}
	if !war.Revenge.Cast(sim, war.CurrentTarget) {
		t.Fatal("Revenge didn't cast in its window")
	}
	if window.IsActive() {
		t.Error("Revenge's window is still open after a Revenge")
	}
	war.Stance = warrior.BattleStance
	window.Activate(sim)
	ready(sim, war, war.Revenge.Spell)
	if war.Revenge.Cast(sim, war.CurrentTarget) {
		t.Error("Revenge cast outside Defensive Stance")
	}
}

// TestForeverTankShieldNeeded checks that Shield Block, Shield Slam and Shield Wall need a shield.
func TestForeverTankShieldNeeded(t *testing.T) {
	sim, war := newTank(t, P1Talents, 0)
	wall := war.GetSpell(core.ActionID{SpellID: 871})
	for _, spell := range []*core.Spell{war.ShieldBlock.Spell, war.ShieldSlam.Spell, wall} {
		ready(sim, war, spell)
		if spell.Cast(sim, war.CurrentTarget) {
			t.Errorf("%v cast with no shield", spell.ActionID)
		}
	}
}

// TestForeverTankBloodrage checks Bloodrage (2687, 29131): 10 Rage at once and 1 a second for 10
// sec, all of it 25% a point more with Improved Bloodrage (12301), for 20% of base health, on a 1 min
// cooldown with no GCD.
func TestForeverTankBloodrage(t *testing.T) {
	for _, c := range []struct {
		talents string
		rage    float64
	}{{"", 20}, {"--1", 25}, {"--2", 30}} {
		sim, war := newTank(t, c.talents, bulwark)
		war.SpendRage(sim, war.CurrentRage(), war.RageRefundMetrics)
		health := war.CurrentHealth()
		if !war.Bloodrage.Cast(sim, war.CurrentTarget) {
			t.Fatal("Bloodrage didn't cast")
		}
		near(t, "Bloodrage's health cost", health-war.CurrentHealth(), 0.2*war.GetBaseStats()[stats.Health], 1e-6)
		advance(sim, 11*time.Second)
		near(t, "Bloodrage's Rage with "+c.talents, war.CurrentRage(), c.rage, 1e-6)
		if war.Bloodrage.CD.Duration != time.Minute || war.Bloodrage.DefaultCast.GCD != 0 {
			t.Errorf("Bloodrage: %s cooldown, %s GCD; want 1m and none", war.Bloodrage.CD.Duration, war.Bloodrage.DefaultCast.GCD)
		}
	}
}

// TestForeverTankLastStand checks Last Stand (12975, 12976): 30% more max health for 20 sec, the
// health given, then taken away, on a 3 min cooldown.
func TestForeverTankLastStand(t *testing.T) {
	sim, war := newTank(t, P1Talents, bulwark)
	spell := war.GetSpell(core.ActionID{SpellID: 12975})
	if spell == nil || spell.CD.Duration != 3*time.Minute {
		t.Fatal("Last Stand is missing or not on a 3 min cooldown")
	}
	maxHealth := war.MaxHealth()
	if !spell.Cast(sim, war.CurrentTarget) {
		t.Fatal("Last Stand didn't cast")
	}
	near(t, "max health under Last Stand", war.MaxHealth()/maxHealth, 1.3, 1e-9)
	near(t, "health under Last Stand", war.CurrentHealth(), war.MaxHealth(), 1e-6)
	advance(sim, 21*time.Second)
	near(t, "max health after Last Stand", war.MaxHealth(), maxHealth, 1e-6)
	if _, none := newTank(t, "", bulwark); none.GetSpell(core.ActionID{SpellID: 12975}) != nil {
		t.Error("Last Stand is there without the talent")
	}
}

// TestForeverTankVanguard checks that Charge works in Defensive Stance only with Vanguard (1310317),
// gives 15 Rage, and none of it makes threat.
func TestForeverTankVanguard(t *testing.T) {
	_, without := newTank(t, "--00000000000", bulwark)
	if without.Charge.StanceMask.Matches(warrior.DefensiveStance) {
		t.Error("Charge works in Defensive Stance without Vanguard")
	}
	sim, war := newTank(t, "--000000000001", bulwark)
	war.SpendRage(sim, war.CurrentRage(), war.RageRefundMetrics)
	if !war.Charge.Cast(sim, war.CurrentTarget) {
		t.Fatal("Charge didn't cast in Defensive Stance with Vanguard")
	}
	near(t, "Charge's Rage", war.CurrentRage(), 15, 1e-9)
}

// TestForeverTankTalentStats checks the talents a tank sheet shows: Shield Specialization 1% block a
// point, Anticipation 4 Defense a point, Deflection 1% parry a point, Cruelty 1% crit a point, and
// Bastion's 2% Physical damage a point with a shield and none without.
func TestForeverTankTalentStats(t *testing.T) {
	sim, base := newTank(t, "", bulwark)
	boss := sim.Encounter.TargetUnits[0]
	sheet := core.TankSheet(&base.Unit, boss)
	sheetOf := func(talents string, offHand int32) (*warrior.Warrior, *proto.TankSheet) {
		sim, war := newTank(t, talents, offHand)
		return war, core.TankSheet(&war.Unit, sim.Encounter.TargetUnits[0])
	}
	_, ss := sheetOf("--05", bulwark)
	near(t, "Shield Specialization 5/5's block", ss.Block-sheet.Block, 0.05, 1e-9)
	_, ant := sheetOf("--0005", bulwark)
	near(t, "Anticipation 5/5's Defense", ant.Defense-sheet.Defense, 20, 1e-9)
	_, defl := sheetOf("05", bulwark)
	near(t, "Deflection 5/5's parry", defl.Parry-sheet.Parry, 0.05, 1e-9)
	cruel, _ := sheetOf("-05", bulwark)
	near(t, "Cruelty 5/5's crit", cruel.GetStat(stats.MeleeCrit)-base.GetStat(stats.MeleeCrit), 5*core.CritRatingPerCritChance, 1e-9)
	bastion, _ := sheetOf("--00000000000000005", bulwark)
	near(t, "Bastion 5/5 with a shield", bastion.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical], 1.1, 1e-9)
	noShield, _ := sheetOf("--00000000000000005", 0)
	near(t, "Bastion 5/5 without a shield", noShield.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical], 1, 1e-9)
}

// TestForeverTankTalentModifiers checks the talents that change one ability: Improved Revenge 20%
// damage a point (12797), Improved Shield Wall 5.5 min off a point (12312), Booming Voice 5% off the
// shouts a point (12321), and Focused Rage 1 Rage off the offensive abilities a point but not Shield
// Block (29787's class mask).
func TestForeverTankTalentModifiers(t *testing.T) {
	_, base := newTank(t, "", bulwark)
	_, impRev := newTank(t, "--00003", bulwark)
	near(t, "Improved Revenge 3/3", impRev.Revenge.DamageMultiplier/base.Revenge.DamageMultiplier, 1.6, 1e-9)
	_, isw := newTank(t, "--00000000000001", bulwark)
	if got := isw.GetSpell(core.ActionID{SpellID: 871}).CD.Duration; got != 15*time.Minute-330*time.Second {
		t.Errorf("Shield Wall with Improved Shield Wall 1/2 has a %s cooldown, want 9m30s", got)
	}
	_, bv := newTank(t, "-5", bulwark)
	near(t, "Demoralizing Shout with Booming Voice 5/5", bv.DemoralizingShout.Cost.GetCurrentCost(), 7.5, 1e-9)
	_, fr := newTank(t, "--0000000000000003", bulwark)
	near(t, "Revenge with Focused Rage 3/3", fr.Revenge.Cost.GetCurrentCost(), 2, 1e-9)
	near(t, "Thunder Clap with Focused Rage 3/3", fr.ThunderClap.Cost.GetCurrentCost(), 17, 1e-9)
	near(t, "Shield Block with Focused Rage 3/3", fr.ShieldBlock.Cost.GetCurrentCost(), 10, 1e-9)
}

// tankedRun sims the warrior with no rotation under a boss for its proc talents.
func tankedRun(t *testing.T, talents string) (*proto.UnitMetrics, *proto.ActionMetrics) {
	t.Helper()
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 991014}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: bulwark}
	player := &proto.Player{
		Race: proto.Race_RaceOrc, Class: proto.Class_ClassWarrior, Level: 60,
		Equipment: &proto.EquipmentSpec{Items: items}, TalentsString: talents,
		Rotation: &proto.APLRotation{}, Spec: defensiveOptions,
		HealingModel: &proto.HealingModel{Hps: 5000, CadenceSeconds: 1},
	}
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
	target.MinBaseDamage = 300
	res := core.RunRaidSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Duration: 300, Targets: []*proto.Target{target}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1, Iterations: 20},
	})
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	var melee *proto.ActionMetrics
	for _, a := range res.EncounterMetrics.Targets[0].Actions {
		if a.Id.GetOtherId() == proto.OtherAction_OtherActionAttack {
			melee = a
		}
	}
	return res.RaidMetrics.Parties[0].Players[0], melee
}

func outcomes(a *proto.ActionMetrics) (blocks, avoids, damaging int32) {
	for _, tm := range a.Targets {
		blocks += tm.Blocks + tm.BlockedCrits
		avoids += tm.Dodges + tm.Parries
		damaging += tm.Hits + tm.Crits + tm.Crushes + tm.Blocks + tm.BlockedCrits
	}
	return
}

func rageEvents(m *proto.UnitMetrics, id *proto.ActionID) (events int32, gain float64) {
	for _, r := range m.Resources {
		if googleProto.Equal(r.Id, id) {
			events += r.Events
			gain += r.Gain
		}
	}
	return
}

// TestForeverTankProcTalents checks the talents that proc on the boss's swings, over 20 fights:
// Shield Specialization gives 5 Rage on 20% of blocks a point (12298, energize 1310318), Master of
// Defense 5 Rage on 50% of dodges and parries a point (1310316, energize 23602), and Enrage (Fury,
// 12317) a 30% chance on any damaging hit taken.
func TestForeverTankProcTalents(t *testing.T) {
	ssID := &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 12727}}
	modID := &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 12727}, Tag: 1}

	m, melee := tankedRun(t, "--05000002")
	blocks, avoids, _ := outcomes(melee)
	if events, gain := rageEvents(m, ssID); events != blocks || gain != 5*float64(blocks) {
		t.Errorf("Shield Specialization 5/5: %d Rage events for %.0f Rage on %d blocks, want one of 5 a block", events, gain, blocks)
	}
	if events, gain := rageEvents(m, modID); events != avoids || gain != 5*float64(avoids) {
		t.Errorf("Master of Defense 2/2: %d Rage events for %.0f Rage on %d dodges and parries, want one of 5 each", events, gain, avoids)
	}

	m, melee = tankedRun(t, "--01000001")
	blocks, avoids, _ = outcomes(melee)
	ss, _ := rageEvents(m, ssID)
	mod, _ := rageEvents(m, modID)
	rate := func(name string, n, of int32, want float64) {
		t.Helper()
		p := float64(n) / float64(of)
		if se := math.Sqrt(want * (1 - want) / float64(of)); math.Abs(p-want) > 4*se {
			t.Errorf("%s procs on %.3f of %d, want %.2f", name, p, of, want)
		}
	}
	rate("Shield Specialization 1/5", ss, blocks, 0.2)
	rate("Master of Defense 1/2", mod, avoids, 0.5)

	m, melee = tankedRun(t, "-0000000005")
	_, _, damaging := outcomes(melee)
	var procs float64
	for _, a := range m.Auras {
		if a.Id.GetSpellId() == 12880 {
			procs = a.ProcsAvg * 20
		}
	}
	rate("Enrage", int32(math.Round(procs)), damaging, 0.3)
}

// defensiveOptions start the warrior in Defensive Stance whatever its talents.
var defensiveOptions = &proto.Player_TankWarrior{TankWarrior: &proto.TankWarrior{Options: &proto.TankWarrior_Options{
	Shout: proto.WarriorShout_WarriorShoutNone, Stance: proto.WarriorStance_WarriorStanceDefensive,
}}}
