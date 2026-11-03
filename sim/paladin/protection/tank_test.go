package protection

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

// The Protection paladin's tank abilities and talents under Forever (client 1.60.1.70753, the
// October 30 hotfix's rows; docs/tank_paladin.md has the audit). Each test checks that the
// mechanic moves exactly what the client says, against the same paladin without it.

const mirahsSong = 15806         // a one-handed sword
const draconianDeflector = 12602 // a shield, 7 Defense
const arcaniteReaper = 12784     // a two-hander

// protSim is a level 60 Human Protection paladin under Forever with a sword and a shield (or the
// items given), Righteous Fury on, no rotation, against a level 63 boss that attacks the paladin.
func protSim(t *testing.T, talents string, options *proto.PaladinOptions, items ...int32) (*core.Simulation, *ProtectionPaladin) {
	t.Helper()
	if options == nil {
		options = &proto.PaladinOptions{PrimarySeal: proto.PaladinSeal_Fury, RighteousFury: true}
	}
	if len(items) == 0 {
		items = []int32{mirahsSong, draconianDeflector}
	}
	equipment := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range equipment {
		equipment[i] = &proto.ItemSpec{}
	}
	equipment[proto.ItemSlot_ItemSlotMainHand].Id = items[0]
	if len(items) > 1 {
		equipment[proto.ItemSlot_ItemSlotOffHand].Id = items[1]
	}
	player := &proto.Player{
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassPaladin,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{Items: equipment},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec:          &proto.Player_ProtectionPaladin{ProtectionPaladin: &proto.ProtectionPaladin{Options: options}},
	}
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
	target.Level = 63
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{target}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*ProtectionPaladin)
}

func near(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Errorf("%s = %.4f, want %.4f", name, got, want)
	}
}

func spell(t *testing.T, pal *ProtectionPaladin, id int32) *core.Spell {
	t.Helper()
	s := pal.GetSpell(core.ActionID{SpellID: id})
	if s == nil {
		t.Fatalf("spell %d not registered", id)
	}
	return s
}

// bossHit is one boss melee swing of raw damage on the paladin that always lands as a plain hit,
// and what the paladin took from it.
func bossHit(sim *core.Simulation, pal *ProtectionPaladin, raw float64) float64 {
	boss := sim.Encounter.TargetUnits[0]
	swing := boss.AutoAttacks.MHAuto()
	return swing.CalcAndDealDamage(sim, &pal.Unit, raw, swing.OutcomeAlwaysHit).Damage
}

// TestHolyShieldRank4 checks Holy Shield's rank 4 at level 60 (1239570, the October 30 hotfix's new
// top rank): 240 mana, 30% block for 10 sec or 4 blocks, 221 Holy damage a block plus 8% of spell
// power at 1.2 times the threat (on top of Righteous Fury's), on the 10 sec cooldown the ranks share.
func TestHolyShieldRank4(t *testing.T) {
	sim, pal := protSim(t, "-00000000000001", nil)
	hs := spell(t, pal, 1239570)
	if got := hs.Cost.GetCurrentCost(); got != 240 {
		t.Errorf("Holy Shield rank 4 costs %.0f, want 240", got)
	}
	if hs.RequiredLevel != 60 || hs.CD.Duration != 10*time.Second {
		t.Errorf("Holy Shield rank 4: level %d, cooldown %s, want 60 and 10s", hs.RequiredLevel, hs.CD.Duration)
	}
	block := pal.GetStat(stats.Block)
	hs.Cast(sim, pal.CurrentTarget)
	aura := pal.GetAuraByID(core.ActionID{SpellID: 1239570})
	if aura == nil || !aura.IsActive() || aura.Duration != 10*time.Second || aura.GetStacks() != 4 {
		t.Fatalf("Holy Shield rank 4 aura: %v, want 10 sec and 4 charges", aura)
	}
	near(t, "Holy Shield block", pal.GetStat(stats.Block)-block, 30*core.BlockRatingPerBlockChance, 1e-6)
	if spell(t, pal, 20928).CD.IsReady(sim) {
		t.Errorf("rank 3 is ready right after rank 4, want the shared 10 sec cooldown")
	}

	proc := pal.GetSpell(core.ActionID{SpellID: 1239570, Tag: 1})
	if proc == nil {
		t.Fatalf("no rank 4 block damage spell")
	}
	// A crit-free, resist-free reading of the damage: base 221 plus 8% of spell power.
	pal.AddStatDynamic(sim, stats.SpellPower, 100)
	result := proc.CalcDamage(sim, pal.CurrentTarget, 221, proc.OutcomeAlwaysHit)
	want := (221 + 0.08*pal.GetStat(stats.SpellPower)) * result.ResistanceMultiplier
	near(t, "Holy Shield block damage", result.Damage, want, 1e-6)
	// Threat: 1.2 for the shield, 1.9 for Righteous Fury's Holy threat, 1.33 for its all threat.
	near(t, "Holy Shield threat a point", result.Threat/result.Damage, 1.2*1.9*1.33, 1e-9)

	// A block spends a charge and deals the damage.
	boss := sim.Encounter.TargetUnits[0]
	swing := boss.AutoAttacks.MHAuto()
	blocked := swing.NewResult(&pal.Unit)
	blocked.Outcome = core.OutcomeBlock
	casts := proc.SpellMetrics[pal.CurrentTarget.UnitIndex].Casts
	aura.OnSpellHitTaken(aura, sim, swing, blocked)
	if aura.GetStacks() != 3 || proc.SpellMetrics[pal.CurrentTarget.UnitIndex].Casts != casts+1 {
		t.Errorf("a block left %d charges and %d damage casts, want 3 and 1", aura.GetStacks(), proc.SpellMetrics[pal.CurrentTarget.UnitIndex].Casts-casts)
	}
}

// TestSealOfFury checks Seal of Fury rank 7 (20423): 200 mana, 30 sec; each landed white swing adds
// 35 Holy damage plus 10% of spell power (20418), and with a shield on, an absorb of half of it;
// Judgement of Fury (20414) does 153 + 2 × 3.69 Holy, an 8.76% spread, plus 42.9% of spell power.
// All of it is Holy, so Righteous Fury's 1.9 × 1.33 threat.
func TestSealOfFury(t *testing.T) {
	sim, pal := protSim(t, "", nil)
	seal := spell(t, pal, 20423)
	if got := seal.Cost.GetCurrentCost(); got != 200 {
		t.Errorf("Seal of Fury costs %.0f, want 200", got)
	}
	seal.Cast(sim, pal.CurrentTarget)
	aura := pal.GetAuraByID(core.ActionID{SpellID: 20423})
	if aura == nil || !aura.IsActive() || aura.Duration != 30*time.Second {
		t.Fatalf("Seal of Fury aura %v, want up for 30 sec", aura)
	}

	proc := spell(t, pal, 20418)
	pal.AddStatDynamic(sim, stats.SpellPower, 200)
	sp := pal.GetStat(stats.SpellPower)
	mh := pal.AutoAttacks.MHAuto()
	white := mh.NewResult(pal.CurrentTarget)
	white.Outcome = core.OutcomeHit
	shield := pal.GetAuraByID(core.ActionID{SpellID: 20423, Tag: 1})
	for i := 0; i < 20; i++ {
		before := proc.SpellMetrics[pal.CurrentTarget.UnitIndex]
		aura.OnSpellHitDealt(aura, sim, mh, white)
		after := proc.SpellMetrics[pal.CurrentTarget.UnitIndex]
		if after.Casts != before.Casts+1 {
			t.Fatalf("a white hit cast the seal's hit %d times, want once", after.Casts-before.Casts)
		}
		if after.Crits != before.Crits {
			continue
		}
		dealt := after.TotalDamage - before.TotalDamage
		near(t, "Seal of Fury hit", dealt, (35+0.1*sp)*pal.CurrentTarget.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexHoly], 2)
		near(t, "Seal of Fury threat a point", (after.TotalThreat-before.TotalThreat)/dealt, 1.9*1.33, 1e-6)
		if !shield.IsActive() {
			t.Fatalf("no Seal of Fury absorb with a shield on")
		}
		// The absorb is half the hit: a boss swing bigger than it loses exactly that much.
		shield.Deactivate(sim)
		raw := bossHit(sim, pal, 1000)
		aura.OnSpellHitDealt(aura, sim, mh, white)
		near(t, "Seal of Fury absorb", raw-bossHit(sim, pal, 1000), dealt/2, 2)
		if shield.IsActive() {
			t.Errorf("the absorb is still up after a hit bigger than it")
		}
		break
	}
	missed := mh.NewResult(pal.CurrentTarget)
	missed.Outcome = core.OutcomeMiss
	casts := proc.SpellMetrics[pal.CurrentTarget.UnitIndex].Casts
	aura.OnSpellHitDealt(aura, sim, mh, missed)
	if proc.SpellMetrics[pal.CurrentTarget.UnitIndex].Casts != casts {
		t.Errorf("a missed swing cast the seal's hit")
	}

	judgement := spell(t, pal, 20414)
	low, high := math.Inf(1), 0.0
	// Roll the judgement itself and read its normal hits.
	judgement.BonusHitRating = 10000
	for i := 0; i < 400; i++ {
		before := judgement.SpellMetrics[pal.CurrentTarget.UnitIndex]
		judgement.ApplyEffects(sim, pal.CurrentTarget, judgement)
		after := judgement.SpellMetrics[pal.CurrentTarget.UnitIndex]
		if after.Crits != before.Crits {
			continue
		}
		d := (after.TotalDamage - before.TotalDamage) / pal.CurrentTarget.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexHoly]
		low, high = min(low, d), max(high, d)
	}
	base := 153 + 2*3.69
	near(t, "Judgement of Fury lowest", low, base*(1-0.08759/2)+0.429*sp, 3)
	near(t, "Judgement of Fury highest", high, base*(1+0.08759/2)+0.429*sp, 3)

	// Judgement judges the active seal.
	jud := spell(t, pal, 20271)
	before := judgement.SpellMetrics[pal.CurrentTarget.UnitIndex].Casts
	jud.Cast(sim, pal.CurrentTarget)
	if judgement.SpellMetrics[pal.CurrentTarget.UnitIndex].Casts != before+1 {
		t.Errorf("Judgement under Seal of Fury didn't cast Judgement of Fury")
	}
}

// TestSealOfFuryNeedsShield checks that without a shield the seal's hit gives no absorb.
func TestSealOfFuryNeedsShield(t *testing.T) {
	sim, pal := protSim(t, "", nil, arcaniteReaper)
	spell(t, pal, 20423).Cast(sim, pal.CurrentTarget)
	aura := pal.GetAuraByID(core.ActionID{SpellID: 20423})
	mh := pal.AutoAttacks.MHAuto()
	white := mh.NewResult(pal.CurrentTarget)
	white.Outcome = core.OutcomeHit
	aura.OnSpellHitDealt(aura, sim, mh, white)
	if pal.GetAuraByID(core.ActionID{SpellID: 20423, Tag: 1}).IsActive() {
		t.Errorf("a Seal of Fury absorb with a two-hander, want none")
	}
}

// TestTemplarsBulwark checks Templar's Bulwark (1311015): an absorb of the paladin's whole health for
// 8 sec, 110 mana, off the global cooldown, a 5 min cooldown less 30 sec a rank of Sacred Duty, and
// Forbearance, under which it can't be cast.
func TestTemplarsBulwark(t *testing.T) {
	sim, pal := protSim(t, "-000001", nil)
	tb := spell(t, pal, 1311015)
	if tb.CD.Duration != 5*time.Minute || tb.Cost.GetCurrentCost() != 110 || tb.DefaultCast.GCD != 0 {
		t.Errorf("Templar's Bulwark: cooldown %s, cost %.0f, GCD %s; want 5m, 110 and none", tb.CD.Duration, tb.Cost.GetCurrentCost(), tb.DefaultCast.GCD)
	}
	raw := bossHit(sim, pal, 1000)
	if !tb.Cast(sim, pal.CurrentTarget) {
		t.Fatalf("Templar's Bulwark didn't cast")
	}
	aura := pal.GetAuraByID(core.ActionID{SpellID: 1311015})
	if !aura.IsActive() || aura.Duration != 8*time.Second {
		t.Fatalf("Templar's Bulwark aura %v, want 8 sec", aura)
	}
	if !pal.GetAura("Forbearance").IsActive() {
		t.Errorf("no Forbearance after Templar's Bulwark")
	}
	// The absorb holds the paladin's max health in raw-sized hits.
	health := pal.MaxHealth()
	hits := 0
	taken := 0.0
	for aura.IsActive() && hits < 1000 {
		taken += bossHit(sim, pal, 1000)
		hits++
	}
	// Within the rounding the engine does to a hit after its damage taken modifiers.
	near(t, "damage the Bulwark absorbed", float64(hits)*raw-taken, health, float64(hits))

	tb.CD.Reset()
	if tb.CanCast(sim, pal.CurrentTarget) {
		t.Errorf("Templar's Bulwark can be cast under Forbearance")
	}

	_, sd := protSim(t, "-000021", nil)
	if got := spell(t, sd, 1311015).CD.Duration; got != 4*time.Minute {
		t.Errorf("Templar's Bulwark's cooldown with Sacred Duty 2/2 is %s, want 4m", got)
	}
	if got := sd.GetAura("Forbearance").Duration; got != 50*time.Second {
		t.Errorf("Forbearance with Sacred Duty 2/2 lasts %s, want 50s", got)
	}
	_, none := protSim(t, "", nil)
	if none.GetSpell(core.ActionID{SpellID: 1311015}) != nil {
		t.Errorf("Templar's Bulwark without the talent")
	}
}

// TestDivineShieldAndProtection checks the two immunities: Divine Shield rank 2 (1020) takes every
// hit to 0 for 12 sec and halves the paladin's damage, Divine Protection rank 2 (5573) takes every
// hit to 0 for 8 sec and stops the swings; they share a 5 min cooldown (Sacred Duty 2/2: 4 min) and
// Forbearance.
func TestDivineShieldAndProtection(t *testing.T) {
	sim, pal := protSim(t, "-00002", nil)
	ds := spell(t, pal, 1020)
	dp := spell(t, pal, 5573)
	if ds.CD.Duration != 4*time.Minute || ds.Cost.GetCurrentCost() != 110 || dp.Cost.GetCurrentCost() != 35 {
		t.Errorf("Divine Shield cooldown %s and cost %.0f, Divine Protection cost %.0f; want 4m, 110 and 35", ds.CD.Duration, ds.Cost.GetCurrentCost(), dp.Cost.GetCurrentCost())
	}
	dealt := pal.PseudoStats.DamageDealtMultiplier
	if !ds.Cast(sim, pal.CurrentTarget) {
		t.Fatalf("Divine Shield didn't cast")
	}
	if got := pal.GetAuraByID(core.ActionID{SpellID: 1020}).Duration; got != 12*time.Second {
		t.Errorf("Divine Shield lasts %s, want 12s", got)
	}
	if got := bossHit(sim, pal, 1000); got != 0 {
		t.Errorf("a hit under Divine Shield did %.0f, want 0", got)
	}
	near(t, "damage dealt under Divine Shield", pal.PseudoStats.DamageDealtMultiplier/dealt, 0.5, 1e-9)
	if dp.CD.IsReady(sim) {
		t.Errorf("Divine Protection ready after Divine Shield, want the shared cooldown")
	}
	pal.GetAuraByID(core.ActionID{SpellID: 1020}).Deactivate(sim)
	near(t, "damage dealt after Divine Shield", pal.PseudoStats.DamageDealtMultiplier, dealt, 1e-9)

	sim, pal = protSim(t, "", nil)
	dp = spell(t, pal, 5573)
	swings := func(until time.Duration) int32 {
		mh := pal.AutoAttacks.MHAuto()
		before := mh.SpellMetrics[pal.CurrentTarget.UnitIndex].Casts
		sim.AddPendingAction(&core.PendingAction{NextActionAt: until, OnAction: func(*core.Simulation) {}})
		for !sim.Step() && sim.CurrentTime < until {
		}
		return mh.SpellMetrics[pal.CurrentTarget.UnitIndex].Casts - before
	}
	if got := swings(time.Second); got == 0 { // the pull starts the swings
		t.Fatalf("no swings in the first second")
	}
	if !dp.Cast(sim, pal.CurrentTarget) {
		t.Fatalf("Divine Protection didn't cast")
	}
	if got := bossHit(sim, pal, 1000); got != 0 {
		t.Errorf("a hit under Divine Protection did %.0f, want 0", got)
	}
	if got := swings(sim.CurrentTime + 7*time.Second); got != 0 {
		t.Errorf("the paladin swung %d times under Divine Protection, want 0", got)
	}
	if got := swings(sim.CurrentTime + 8*time.Second); got == 0 {
		t.Errorf("the paladin doesn't swing again after Divine Protection")
	}
}

// TestRighteousFuryThreat checks Righteous Fury (25780): 33% more of all threat (a white swing) and
// 90% more again on Holy (Holy Strike), and Improved Righteous Fury's 3% less damage taken a rank.
func TestRighteousFuryThreat(t *testing.T) {
	_, off := protSim(t, "", &proto.PaladinOptions{PrimarySeal: proto.PaladinSeal_Fury})
	_, on := protSim(t, "", nil)
	near(t, "all threat", on.PseudoStats.ThreatMultiplier/off.PseudoStats.ThreatMultiplier, 1.33, 1e-9)
	near(t, "Holy Strike threat", spell(t, on, 10333).ThreatMultiplier/spell(t, off, 10333).ThreatMultiplier, 1.9, 1e-9)
	near(t, "white swing threat", on.AutoAttacks.MHAuto().ThreatMultiplier/off.AutoAttacks.MHAuto().ThreatMultiplier, 1, 1e-9)
	near(t, "damage taken with Righteous Fury alone", on.PseudoStats.DamageTakenMultiplier/off.PseudoStats.DamageTakenMultiplier, 1, 1e-9)
	_, irf := protSim(t, "-00000000003", nil)
	near(t, "Improved Righteous Fury 3/3", irf.PseudoStats.DamageTakenMultiplier/on.PseudoStats.DamageTakenMultiplier, 0.91, 1e-9)
	_, irfOff := protSim(t, "-00000000003", &proto.PaladinOptions{PrimarySeal: proto.PaladinSeal_Fury})
	near(t, "Improved Righteous Fury without Righteous Fury", irfOff.PseudoStats.DamageTakenMultiplier/off.PseudoStats.DamageTakenMultiplier, 1, 1e-9)
}

// TestProtectionStatTalents checks the talents that move the tank sheet: Anticipation 4 Defense a
// rank, Toughness 2% armor from items a rank, Redoubt's 6% block a rank for 10 sec or 5 blocks,
// Shield Specialization's 15% block value a rank, Precision's 1% hit a rank, Divine Intellect's 2%
// Intellect a rank.
func TestProtectionStatTalents(t *testing.T) {
	_, base := protSim(t, "", nil)
	sheet := func(pal *ProtectionPaladin) *proto.TankSheet {
		return core.TankSheet(&pal.Unit, pal.Env.Encounter.TargetUnits[0])
	}
	_, anticipation := protSim(t, "-000000000005", nil)
	near(t, "Anticipation 5/5 Defense", anticipation.GetStat(stats.Defense)-base.GetStat(stats.Defense), 20, 1e-9)
	near(t, "Anticipation 5/5 dodge on the sheet", sheet(anticipation).Dodge-sheet(base).Dodge, 20*0.0004, 1e-9)

	_, toughness := protSim(t, "-000000005", nil)
	near(t, "Toughness 5/5 armor", toughness.GetStat(stats.Armor)-base.GetStat(stats.Armor), 0.1*base.Equipment.Stats()[stats.Armor], 1e-6)

	sim, redoubt := protSim(t, "-05", nil)
	block := redoubt.GetStat(stats.Block)
	aura := redoubt.GetAuraByID(core.ActionID{SpellID: 20128})
	aura.Activate(sim)
	aura.SetStacks(sim, 5)
	near(t, "Redoubt 5/5 block", redoubt.GetStat(stats.Block)-block, 30*core.BlockRatingPerBlockChance, 1e-6)
	if aura.Duration != 10*time.Second {
		t.Errorf("Redoubt lasts %s, want 10s", aura.Duration)
	}
	boss := sim.Encounter.TargetUnits[0]
	blocked := boss.AutoAttacks.MHAuto().NewResult(&redoubt.Unit)
	blocked.Outcome = core.OutcomeBlock
	for range 5 {
		aura.OnSpellHitTaken(aura, sim, boss.AutoAttacks.MHAuto(), blocked)
	}
	if aura.IsActive() {
		t.Errorf("Redoubt still up after 5 blocks")
	}
	// It procs on 10% of the melee hits taken.
	procs := 0
	const swings = 20000
	for range swings {
		aura.Deactivate(sim)
		bossHit(sim, redoubt, 100)
		if aura.IsActive() {
			procs++
		}
	}
	near(t, "Redoubt proc rate", float64(procs)/swings, 0.1, 0.01)

	_, shieldSpec := protSim(t, "-05000002", nil)
	_, redoubtOnly := protSim(t, "-05", nil)
	// The shield's block value takes the 30%; the part from Strength doesn't (core Unit.BlockValue).
	near(t, "Shield Specialization 2/2 block value", shieldSpec.BlockValue()-redoubtOnly.BlockValue(), 0.3*redoubtOnly.GetStat(stats.BlockValue), 1e-6)

	_, precision := protSim(t, "-003", nil)
	near(t, "Precision 3/3 melee hit", precision.GetStat(stats.MeleeHit)-base.GetStat(stats.MeleeHit), 3*core.MeleeHitRatingPerHitChance, 1e-9)
	near(t, "Precision 3/3 spell hit", precision.GetStat(stats.SpellHit)-base.GetStat(stats.SpellHit), 3*core.SpellHitRatingPerHitChance, 1e-9)

	_, di := protSim(t, "-5", nil)
	near(t, "Divine Intellect 5/5", di.GetStat(stats.Intellect)/base.GetStat(stats.Intellect), 1.1, 1e-2)
}

// TestShieldSpecializationMakesNoThreat checks that Shield Specialization's mana (1310925, the no
// threat attribute) books no threat, where Forever's other mana gains make half a threat a mana.
func TestShieldSpecializationMakesNoThreat(t *testing.T) {
	sim, pal := protSim(t, "-05000002", nil)
	aura := pal.GetAura("Shield Specialization Mana")
	pal.SpendMana(sim, 0.5*pal.MaxMana(), pal.NewManaMetrics(core.ActionID{OtherID: 1}))
	boss := sim.Encounter.TargetUnits[0]
	blocked := boss.AutoAttacks.MHAuto().NewResult(&pal.Unit)
	blocked.Outcome = core.OutcomeBlock
	before := pal.CurrentMana()
	manaGain := pal.GetSpell(core.ActionID{OtherID: proto.OtherAction_OtherActionManaGain})
	threat := manaGain.SpellMetrics[pal.CurrentTarget.UnitIndex].TotalThreat
	aura.OnSpellHitTaken(aura, sim, boss.AutoAttacks.MHAuto(), blocked)
	near(t, "Shield Specialization mana", pal.CurrentMana()-before, 0.04*pal.MaxMana(), 1e-6)
	if got := manaGain.SpellMetrics[pal.CurrentTarget.UnitIndex].TotalThreat - threat; got != 0 {
		t.Errorf("Shield Specialization's mana made %.1f threat, want none", got)
	}
	// Another mana gain does make threat, half a threat a mana times Righteous Fury's 1.33.
	pal.AddMana(sim, 100, pal.NewManaMetrics(core.ActionID{SpellID: 1}))
	near(t, "threat of 100 other mana", manaGain.SpellMetrics[pal.CurrentTarget.UnitIndex].TotalThreat-threat, 100*0.5*1.33, 1e-6)
}

// TestIronCreedAndShieldsworn checks Iron Creed 3/3 (1311034: 30% Holy Strike threat, and with
// Righteous Fury 6% less damage taken for 6 sec after any Holy Strike, a missed one too) and
// Shieldsworn (1328544: Holy Strike 4 sec faster; a landed one deals Holy Shield's damage without
// a charge).
func TestIronCreedAndShieldsworn(t *testing.T) {
	_, base := protSim(t, "", nil)
	sim, pal := protSim(t, "-0000000000000103", nil)
	hs := spell(t, pal, 10333)
	near(t, "Iron Creed 3/3 threat", hs.ThreatMultiplier/spell(t, base, 10333).ThreatMultiplier, 1.3, 1e-9)
	hs.BonusHitRating = -100000 // every Holy Strike misses
	taken := pal.PseudoStats.DamageTakenMultiplier
	hs.ApplyEffects(sim, pal.CurrentTarget, hs)
	ic := pal.GetAura("Iron Creed")
	if ic == nil || !ic.IsActive() || ic.Duration != 6*time.Second {
		t.Fatalf("no 6 sec Iron Creed after a missed Holy Strike")
	}
	near(t, "Iron Creed 3/3 damage taken", pal.PseudoStats.DamageTakenMultiplier/taken, 0.94, 1e-9)

	_, noRF := protSim(t, "-0000000000000103", &proto.PaladinOptions{PrimarySeal: proto.PaladinSeal_Fury})
	if noRF.GetAura("Iron Creed") != nil {
		t.Errorf("Iron Creed's damage cut without Righteous Fury")
	}

	sim, sw := protSim(t, "-00000000000001031", nil)
	strike := spell(t, sw, 10333)
	if strike.CD.Duration != 6*time.Second {
		t.Errorf("Holy Strike with Shieldsworn: %s cooldown, want 6s", strike.CD.Duration)
	}
	spell(t, sw, 1239570).Cast(sim, sw.CurrentTarget)
	proc := sw.GetSpell(core.ActionID{SpellID: 1239570, Tag: 1})
	strike.BonusHitRating = 100000
	casts := proc.SpellMetrics[sw.CurrentTarget.UnitIndex].Casts
	strike.ApplyEffects(sim, sw.CurrentTarget, strike)
	if got := proc.SpellMetrics[sw.CurrentTarget.UnitIndex].Casts - casts; got != 1 {
		t.Errorf("Shieldsworn Holy Strike cast Holy Shield's damage %d times, want 1", got)
	}
	if got := sw.GetAuraByID(core.ActionID{SpellID: 1239570}).GetStacks(); got != 4 {
		t.Errorf("Holy Shield has %d charges after a Shieldsworn Holy Strike, want 4", got)
	}
}

// TestReckoning checks Reckoning 5/5 (20177): a crit taken always stores an extra attack, a block
// 40% of the time and at most once every 1.5 sec.
func TestReckoning(t *testing.T) {
	sim, pal := protSim(t, "-000000000000005", nil)
	trigger := pal.GetAura("Reckoning Trigger")
	boss := sim.Encounter.TargetUnits[0]
	swing := boss.AutoAttacks.MHAuto()
	stored := pal.GetAuraByID(core.ActionID{SpellID: 21919})
	result := swing.NewResult(&pal.Unit)
	result.Outcome = core.OutcomeBlock
	n, charges := 0, 0
	for range 4000 {
		stored.SetStacks(sim, 0)
		sim.CurrentTime += 2 * time.Second // past the block cooldown
		trigger.OnSpellHitTaken(trigger, sim, swing, result)
		if stored.GetStacks() > 0 {
			charges++
		}
		n++
	}
	near(t, "Reckoning 5/5 block chance", float64(charges)/float64(n), 0.4, 0.03)
}

// TestBlessingOfSanctuaryGone checks that Forever has no Blessing of Sanctuary (no client row).
func TestBlessingOfSanctuaryGone(t *testing.T) {
	_, pal := protSim(t, "", &proto.PaladinOptions{PrimarySeal: proto.PaladinSeal_Fury, RighteousFury: true, PersonalBlessing: proto.Blessings_BlessingOfSanctuary})
	if pal.GetAura("Blessing of Sanctuary Trigger") != nil {
		t.Errorf("Blessing of Sanctuary under Forever")
	}
}

// TestForeverProtTankScore runs the tank bench's Balanced request (the synthetic set, Seal of Fury,
// the paladin_tank rotation) as a scored tank: the replay agrees with the run, and the rotation keeps
// Holy Shield up, casts Seal of Fury's hits and Judgement of Fury, and Templar's Bulwark on its timer.
func TestForeverProtTankScore(t *testing.T) {
	data, err := os.ReadFile("../../../tools/tankbench/req/paladin_protection_balanced.json")
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	const iterations, seconds, hps = 300, 120, 500
	rsr.Raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	rsr.Encounter.Duration, rsr.Encounter.DurationVariation = seconds, 0
	rsr.SimOptions.Iterations = iterations
	player := rsr.Raid.Parties[0].Players[0]
	player.HealingModel = &proto.HealingModel{Hps: hps, CadenceSeconds: 2, CadenceVariation: 1}
	player.TankScoring = &proto.TankScoring{}
	res := core.RunRaidSimConcurrent(rsr)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	m := res.RaidMetrics.Parties[0].Players[0]
	over := 0
	for _, h := range m.TankScore.GetSurviveHps() {
		if h > hps {
			over++
		}
	}
	near(t, "the run's deaths against the replay's", m.ChanceOfDeath, float64(over)/iterations, 1e-9)

	casts := map[int32]float64{}
	for _, a := range m.Actions {
		if id := a.Id.GetSpellId(); id != 0 {
			for _, tgt := range a.Targets {
				casts[id] += float64(tgt.Casts) / iterations
			}
		}
	}
	// Holy Shield every 10 sec, Holy Strike every 6, Judgement every 8, Bulwark once in 2 min.
	for _, c := range []struct {
		name     string
		id       int32
		min, max float64
	}{
		{"Holy Shield rank 4", 1239570, 10, 13},
		{"Holy Strike", 10333, 16, 21},
		{"Judgement of Fury", 20414, 12, 17},
		{"Templar's Bulwark", 1311015, 1, 1},
		{"Seal of Fury's hits", 20418, 40, 200},
	} {
		if got := casts[c.id]; got < c.min || got > c.max {
			t.Errorf("%s: %.1f casts a fight, want %.0f to %.0f", c.name, got, c.min, c.max)
		}
	}
}
