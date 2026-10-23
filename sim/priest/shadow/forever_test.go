package shadow

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/priest"
)

func newShadowPriestSim(t *testing.T, talents string) (*core.Simulation, *ShadowPriest) {
	return newShadowPriestSimRace(t, talents, proto.Race_RaceUndead)
}

func newShadowPriestSimRace(t *testing.T, talents string, race proto.Race) (*core.Simulation, *ShadowPriest) {
	player := &proto.Player{
		Race:          race,
		Class:         proto.Class_ClassPriest,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec:          PlayerOptionsBasic,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 10, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*ShadowPriest)
}

func topRank(t *testing.T, name string, ranks []*core.Spell) *core.Spell {
	for i := len(ranks) - 1; i >= 0; i-- {
		if ranks[i] != nil {
			return ranks[i]
		}
	}
	t.Fatalf("no %s rank", name)
	return nil
}

// TestShadowformCosts checks that Shadowform's half cost and the percent cost talents all
// multiply (Devouring Contagion 2/2 and Mental Agility 3/3 on Devouring Plague: 50% of 90%,
// then half of that; wowsims/forever fef2bcaebd), and which spells each one reaches.
// Mental Agility leaves out Mind Blast, Mind Flay and Shadow Word: Death.
func TestShadowformCosts(t *testing.T) {
	sim, sp := newShadowPriestSim(t, "00503523103--50052101231302051")
	sp.ShadowformAura.Activate(sim)

	checks := []struct {
		name  string
		spell *core.Spell
		want  float64
	}{
		{"Devouring Plague", topRank(t, "Devouring Plague", sp.DevouringPlague), 985 * 0.5 * 0.9 * 0.5},
		{"Shadow Word: Pain", topRank(t, "Shadow Word: Pain", sp.ShadowWordPain), 470 * 0.9 * 0.5},
		{"Mind Blast", topRank(t, "Mind Blast", sp.MindBlast), 350 * 0.5},
		{"Mind Flay", topRank(t, "Mind Flay", sp.MindFlay), 205 * 0.5},
		{"Shadow Word: Death", topRank(t, "Shadow Word: Death", sp.ShadowWordDeath), 340 * 0.5},
	}
	for _, c := range checks {
		if got := c.spell.Cost.GetCurrentCost(); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%s costs %.2f in Shadowform, want %.2f", c.name, got, c.want)
		}
	}
}

// TestShadowformCritBonus checks Shadowform's +100% crit bonus: Mind Blast and, since 70425,
// Shadow Word: Death crit for 200%, and Mind Blast goes back after Shadowform.
func TestShadowformCritBonus(t *testing.T) {
	sim, sp := newShadowPriestSim(t, "00503523103--50052101231302051")
	target := sp.CurrentTarget
	at := sp.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand]
	mindBlast := topRank(t, "Mind Blast", sp.MindBlast)
	death := topRank(t, "Shadow Word: Death", sp.ShadowWordDeath)

	check := func(label string, spell *core.Spell, want float64) {
		if got := spell.CritMultiplier(at); math.Abs(got-want) > 1e-9 {
			t.Errorf("%s: %s crits for %.3f, want %.3f", label, spell.ActionID, got, want)
		}
	}
	check("caster form", mindBlast, 1.5)
	sp.ShadowformAura.Activate(sim)
	check("Shadowform", mindBlast, 2.0)
	check("Shadowform", death, 2.0)
	sp.ShadowformAura.Deactivate(sim)
	check("caster form again", mindBlast, 1.5)
}

// TestDevouringPlagueCooldown checks the 24 sec cooldown every race has under Forever (70425).
func TestDevouringPlagueCooldown(t *testing.T) {
	_, sp := newShadowPriestSim(t, "00503523103--50052101231302051")
	if got := topRank(t, "Devouring Plague", sp.DevouringPlague).CD.Duration; got != priest.DevouringPlagueCooldown {
		t.Errorf("Devouring Plague cooldown %s, want %s", got, priest.DevouringPlagueCooldown)
	}
}

// TestInnerFocusCrit checks that a waiting Inner Focus adds 25% crit to the direct spells
// it can be spent on but not to Shadow Word: Death or to spells that only tick (Shadow Word:
// Pain, Devouring Plague, Mind Flay), whose ticks roll crit live under Forever.
func TestInnerFocusCrit(t *testing.T) {
	sim, sp := newShadowPriestSim(t, "00503521103--50052101231302051")
	checks := []struct {
		name  string
		spell *core.Spell
		bonus float64
	}{
		{"Mind Blast", topRank(t, "Mind Blast", sp.MindBlast), 25},
		{"Shadow Word: Death", topRank(t, "Shadow Word: Death", sp.ShadowWordDeath), 0},
		{"Shadow Word: Pain", topRank(t, "Shadow Word: Pain", sp.ShadowWordPain), 0},
		{"Devouring Plague", topRank(t, "Devouring Plague", sp.DevouringPlague), 0},
		{"Mind Flay", topRank(t, "Mind Flay", sp.MindFlay), 0},
	}
	before := make([]float64, len(checks))
	for i, c := range checks {
		before[i] = c.spell.BonusCritRating
	}
	sp.InnerFocusAura.Activate(sim)
	for i, c := range checks {
		want := before[i] + c.bonus*core.SpellCritRatingPerCritChance
		if got := c.spell.BonusCritRating; math.Abs(got-want) > 1e-9 {
			t.Errorf("%s has %.2f bonus crit rating under Inner Focus, want %.2f", c.name, got, want)
		}
	}
}

// TestInnerFocusCost checks that Inner Focus makes Mind Blast and Devouring Plague free.
// Devouring Plague is in the free cast's class mask through the fourth mask word.
func TestInnerFocusCost(t *testing.T) {
	sim, sp := newShadowPriestSim(t, "00503521103--50052101231302051")
	mindBlast := topRank(t, "Mind Blast", sp.MindBlast)
	plague := topRank(t, "Devouring Plague", sp.DevouringPlague)
	sp.InnerFocusAura.Activate(sim)
	if got := mindBlast.Cost.GetCurrentCost(); got != 0 {
		t.Errorf("Mind Blast costs %.2f under Inner Focus, want 0", got)
	}
	if got := plague.Cost.GetCurrentCost(); got != 0 {
		t.Errorf("Devouring Plague costs %.2f under Inner Focus, want 0", got)
	}
}

// TestDownrankCoefficient checks Forever's downranking penalty at level 60: a rank keeps its
// full spell power coefficient up to 17 levels below the caster and loses 5% of it for each
// level past that (Smite rank 7 at 46 keeps it, rank 6 at 38 keeps 75%, rank 4 at 22 nothing).
func TestDownrankCoefficient(t *testing.T) {
	_, sp := newShadowPriestSim(t, "00503523103--50052101231302051")
	checks := []struct {
		rank int
		want float64
	}{
		{8, priest.SmiteSpellCoef[8]},
		{7, priest.SmiteSpellCoef[7]},
		{6, priest.SmiteSpellCoef[6] * 0.75},
		{5, priest.SmiteSpellCoef[5] * 0.35},
		{4, 0},
	}
	for _, c := range checks {
		got := sp.Smite[c.rank].BonusCoefficient
		if math.Abs(got-c.want) > 1e-9 {
			t.Errorf("Smite rank %d coefficient %.4f, want %.4f", c.rank, got, c.want)
		}
	}
}

// TestSmiteRankTuning checks the report assumption the user ruled on 2026-10-15: Smite rank 8
// deals 9% more than the client's 166.62 to 186.38, and rank 7 is unchanged.
func TestSmiteRankTuning(t *testing.T) {
	for _, c := range []struct {
		rank      int
		low, high float64
	}{
		{8, 166.62 * 1.09, 186.38 * 1.09},
		{7, 123.74, 139.26},
	} {
		low, high := priest.SmiteBaseDamageAt(c.rank, 60)
		if math.Abs(low-c.low) > 1e-9 || math.Abs(high-c.high) > 1e-9 {
			t.Errorf("Smite rank %d at 60: %.2f to %.2f, want %.2f to %.2f", c.rank, low, high, c.low, c.high)
		}
	}
}

// TestCooldowns70425 checks the cooldowns client 1.60.1.70425 changed: Devouring Plague 24
// sec, Shadow Word: Death 18 sec, Vampiric Embrace 30 sec.
func TestCooldowns70425(t *testing.T) {
	_, sp := newShadowPriestSim(t, "00503523103--50052101231302051")
	for _, c := range []struct {
		name  string
		spell *core.Spell
		want  time.Duration
	}{
		{"Devouring Plague", topRank(t, "Devouring Plague", sp.DevouringPlague), 24 * time.Second},
		{"Shadow Word: Death", topRank(t, "Shadow Word: Death", sp.ShadowWordDeath), 18 * time.Second},
		{"Vampiric Embrace", sp.VampiricEmbrace, 30 * time.Second},
	} {
		if got := c.spell.CD.Duration; got != c.want {
			t.Errorf("%s cooldown %s, want %s", c.name, got, c.want)
		}
	}
}

// TestImprovedMindFlay checks 70425's Improved Mind Flay: each tick deals 10% a rank more
// than the one before it, compounding (1, 1.2, 1.44 at 2 points).
func TestImprovedMindFlay(t *testing.T) {
	_, sp := newShadowPriestSim(t, "00503523103--50052101231302051")
	for tick, want := range []float64{1, 1, 1.2, 1.44} {
		if tick == 0 {
			continue
		}
		if got := sp.ImprovedMindFlayTickMultiplier(int32(tick)); math.Abs(got-want) > 1e-9 {
			t.Errorf("Improved Mind Flay tick %d multiplier %.4f, want %.4f", tick, got, want)
		}
	}
}

// TestDarkInspiration checks the 70425 talent's Spirit buff: 3% a rank for 10 sec.
func TestDarkInspiration(t *testing.T) {
	sim, sp := newShadowPriestSim(t, "00503523103--50052101231302051")
	spirit := sp.GetStat(stats.Spirit)
	sp.DarkInspirationAura.Activate(sim)
	// Primary stats are floored, so within a point.
	if got, want := sp.GetStat(stats.Spirit), spirit*1.09; math.Abs(got-want) > 1 {
		t.Errorf("Spirit %.2f under Dark Inspiration 3/5, want %.2f", got, want)
	}
	if got := sp.DarkInspirationAura.Duration; got != 10*time.Second {
		t.Errorf("Dark Inspiration lasts %s, want 10s", got)
	}
}

// TestTwilightFocus checks the pushback protection: 70% at 3/3 on the cast and channelled
// spells in Twilight Focus's mask, none on the instants.
func TestTwilightFocus(t *testing.T) {
	_, sp := newShadowPriestSim(t, "-3025011323-50052101231302051")
	for _, c := range []struct {
		name  string
		spell *core.Spell
		want  float64
	}{
		{"Mind Blast", topRank(t, "Mind Blast", sp.MindBlast), 0.70},
		{"Mind Flay", topRank(t, "Mind Flay", sp.MindFlay), 0.70},
		{"Shadow Word: Pain", topRank(t, "Shadow Word: Pain", sp.ShadowWordPain), 0},
	} {
		if got := c.spell.PushbackReduction; math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%s ignores pushback %.2f of the time, want %.2f", c.name, got, c.want)
		}
	}
}

// TestPowerInLight checks that Power in Light (5/5, 10%) counts the priest's Holy Fire or
// Shadow Word: Pain on the target (15 October notes; client 70425 tooltip of 1309969).
func TestPowerInLight(t *testing.T) {
	for _, c := range []struct {
		name  string
		ranks func(sp *ShadowPriest) []*core.Spell
	}{
		{"Shadow Word: Pain", func(sp *ShadowPriest) []*core.Spell { return sp.ShadowWordPain }},
		{"Holy Fire", func(sp *ShadowPriest) []*core.Spell { return sp.HolyFire }},
	} {
		sim, sp := newShadowPriestSim(t, "525035231030001001-30505110002-3")
		target := sim.GetTargetUnit(0)
		if got := sp.PowerInLightMultiplier(target); got != 1 {
			t.Fatalf("Power in Light %.3f with nothing on the target, want 1", got)
		}
		topRank(t, c.name, c.ranks(sp)).Dot(target).Apply(sim)
		if got, want := sp.PowerInLightMultiplier(target), 1.1; math.Abs(got-want) > 1e-9 {
			t.Errorf("Power in Light %.3f with %s on the target, want %.3f", got, c.name, want)
		}
	}
}

// TestShadowWeavingLive checks that Forever's Shadow Weaving stacks sit on the target as
// Shadow damage taken from the priest, so a Shadow Word: Pain already ticking gains the
// stacks put on after it, and that they do not touch the priest's Holy damage.
func TestShadowWeavingLive(t *testing.T) {
	sim, sp := newShadowPriestSim(t, "00503523103--50052101231302051")
	if sp.Talents.ShadowWeaving == 0 {
		t.Fatal("the build has no Shadow Weaving")
	}
	target := sim.GetTargetUnit(0)
	at := sp.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand]
	pain := topRank(t, "Shadow Word: Pain", sp.ShadowWordPain)
	smite := topRank(t, "Smite", sp.Smite)

	dot := pain.Dot(target)
	dot.Apply(sim)
	snapshot := dot.SnapshotAttackerMultiplier
	before := pain.TargetDamageMultiplier(at, true)
	smiteBefore := smite.TargetDamageMultiplier(at, false)

	aura := sp.ShadowWeavingAuras.Get(target)
	aura.Activate(sim)
	aura.SetStacks(sim, 5)

	if got, want := pain.TargetDamageMultiplier(at, true), before*1.1; math.Abs(got-want) > 1e-9 {
		t.Errorf("Shadow Word: Pain ticks take %.4f with 5 stacks, want %.4f", got, want)
	}
	if got := dot.SnapshotAttackerMultiplier; got != snapshot {
		t.Errorf("the running dot's snapshot moved from %.4f to %.4f", snapshot, got)
	}
	if got := smite.TargetDamageMultiplier(at, false); got != smiteBefore {
		t.Errorf("Smite takes %.4f with 5 stacks, want %.4f", got, smiteBefore)
	}
}

// TestDarkInspirationNeedsHit checks that only a Mind Blast that lands gives Dark Inspiration.
func TestDarkInspirationNeedsHit(t *testing.T) {
	sim, sp := newShadowPriestSim(t, "00503523103--50052101231302051")
	target := sim.GetTargetUnit(0)
	mindBlast := topRank(t, "Mind Blast", sp.MindBlast)
	mindBlast.BonusHitRating -= 20 * core.SpellHitRatingPerHitChance

	metrics := &mindBlast.SpellMetrics[target.UnitIndex]
	missed, landed := false, false
	for i := 0; i < 200 && !(missed && landed); i++ {
		sp.DarkInspirationAura.Deactivate(sim)
		misses := metrics.Misses
		mindBlast.ApplyEffects(sim, target, mindBlast)
		miss := metrics.Misses > misses
		if active := sp.DarkInspirationAura.IsActive(); active == miss {
			t.Fatalf("Dark Inspiration active %v after a Mind Blast that missed %v", active, miss)
		}
		missed, landed = missed || miss, landed || !miss
	}
	if !missed || !landed {
		t.Fatalf("no miss and hit in 200 Mind Blasts (missed %v, landed %v)", missed, landed)
	}
}

// TestChastiseShadowform checks that Chastise cannot be cast in Shadowform (StanceExclude
// 0x8000000 on 1277335).
func TestChastiseShadowform(t *testing.T) {
	sim, sp := newShadowPriestSimRace(t, "00503523103--50052101231302051", proto.Race_RaceDwarf)
	target := sim.GetTargetUnit(0)
	target.MobType = proto.MobType_MobTypeHumanoid
	chastise := topRank(t, "Chastise", sp.Chastise)
	if !chastise.CanCast(sim, target) {
		t.Fatal("Chastise cannot be cast in caster form")
	}
	sp.ShadowformAura.Activate(sim)
	if chastise.CanCast(sim, target) {
		t.Error("Chastise can be cast in Shadowform")
	}
	sp.ShadowformAura.Deactivate(sim)
	if !chastise.CanCast(sim, target) {
		t.Error("Chastise cannot be cast after Shadowform")
	}
}

// TestPenanceCastProcs checks that a Penance rolls on-cast procs (Vestments of the Devout's 5
// pieces listen for spell damage casts) through its channel, and its three bolts don't, so a
// Penance rolls once. The bolts fly at speed 40 (wowsims/forever fe7af8b9b9).
func TestPenanceCastProcs(t *testing.T) {
	_, sp := newShadowPriestSim(t, "525035231030001001-30505110002-3")
	channel := topRank(t, "Penance", sp.Penance)
	if !channel.ProcMask.Matches(core.ProcMaskSpellDamage) || channel.Flags.Matches(core.SpellFlagNoOnCastComplete) {
		t.Errorf("the Penance channel has proc mask %v and flags %v, want spell damage and on-cast procs", channel.ProcMask, channel.Flags)
	}
	bolt := sp.GetSpell(core.ActionID{SpellID: priest.PenanceBoltSpellId[priest.PenanceRanks]})
	if bolt == nil {
		t.Fatal("no Penance bolt")
	}
	if !bolt.Flags.Matches(core.SpellFlagNoOnCastComplete) {
		t.Error("a Penance bolt rolls on-cast procs")
	}
	if bolt.MissileSpeed != 40 {
		t.Errorf("a Penance bolt flies at %.0f, want 40", bolt.MissileSpeed)
	}
}

// TestVampiricEmbraceCastProcs checks that Vampiric Embrace is a spell damage cast under
// Forever, so it rolls on-cast procs.
func TestVampiricEmbraceCastProcs(t *testing.T) {
	_, sp := newShadowPriestSim(t, "00500501103-3-50052101231302251")
	if sp.VampiricEmbrace == nil {
		t.Fatal("no Vampiric Embrace")
	}
	if !sp.VampiricEmbrace.ProcMask.Matches(core.ProcMaskSpellDamage) || sp.VampiricEmbrace.Flags.Matches(core.SpellFlagNoOnCastComplete) {
		t.Errorf("Vampiric Embrace has proc mask %v and flags %v, want spell damage and on-cast procs", sp.VampiricEmbrace.ProcMask, sp.VampiricEmbrace.Flags)
	}
}

// TestShadowguard checks the Troll's Shadowguard: castable only in Shadowform, 3 charges,
// and a raid hit (sim/core/raid_hits.go, a spell from the boss) strikes the boss for 96 plus
// 26.7% of spell power, at most once every 3.5 sec. With 10 hits a minute and the charges
// kept at 3, a 9 min fight fires about one strike every 3.5 + 6 sec.
func TestShadowguard(t *testing.T) {
	const minutes = 9
	const hitsPerMinute = 10.0
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(&proto.Player{
			Race:          proto.Race_RaceTroll,
			Class:         proto.Class_ClassPriest,
			Level:         60,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "00500501103-3-50052101231302251",
			Rotation:      &proto.APLRotation{},
			Spec:          PlayerOptionsBasic,
		}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter: &proto.Encounter{Duration: minutes * 60, Targets: []*proto.Target{core.NewDefaultTarget()},
			RaidDamageHitsPerMinute: hitsPerMinute, RaidDamagePerHit: 1500},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1, ExactDamage: true},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	sp := sim.Raid.Parties[0].Players[0].(*ShadowPriest)
	target := sim.GetTargetUnit(0)
	guard := topRank(t, "Shadowguard", sp.Shadowguard)
	aura := sp.ShadowguardAuras[priest.ShadowguardRanks]
	strike := sp.GetSpell(core.ActionID{SpellID: priest.ShadowguardStrikeSpellId[priest.ShadowguardRanks]})

	var perHit float64
	sim.AddPendingAction(&core.PendingAction{NextActionAt: core.DurationFromSeconds(0.5), OnAction: func(sim *core.Simulation) {
		sp.SetGCDTimer(sim, sim.CurrentTime)
		if guard.CanCast(sim, target) {
			t.Error("Shadowguard can be cast outside Shadowform")
		}
		sp.ShadowformAura.Activate(sim)
		if !guard.Cast(sim, target) {
			t.Fatal("Shadowguard didn't cast in Shadowform")
		}
		if aura.GetStacks() != 3 {
			t.Errorf("Shadowguard has %d charges, want 3", aura.GetStacks())
		}
		sp.AddStatDynamic(sim, stats.SpellPower, 1000)
		at := sp.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand]
		perHit = (96 + 0.267*1000) * strike.AttackerDamageMultiplier(at, false) * strike.TargetDamageMultiplier(at, false)
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period:   time.Second,
			OnAction: func(sim *core.Simulation) { aura.SetStacks(sim, 3) },
		})
	}})
	for !sim.Step() {
	}

	metrics := strike.SpellMetrics[target.UnitIndex]
	fired := float64(metrics.Hits + metrics.Misses)
	if want := minutes * 60 / (3.5 + 60/hitsPerMinute); math.Abs(fired-want) > 10 {
		t.Errorf("%.0f strikes fired at the boss, want about %.0f", fired, want)
	}
	if metrics.Hits == 0 {
		t.Fatal("no strike landed")
	}
	if got := metrics.TotalDamage / float64(metrics.Hits); math.Abs(got-perHit) > 0.01*perHit {
		t.Errorf("a strike deals %.1f, want %.1f", got, perHit)
	}
	if metrics.Crits > 0 {
		t.Errorf("%d strikes crit", metrics.Crits)
	}
}

// newWandPriestSim is a level 60 priest holding the given wand with an empty rotation, for a
// fight of the given length under Forever.
func newWandPriestSim(t *testing.T, talents string, wandID int32, seconds int32) (*core.Simulation, *ShadowPriest) {
	return newWandPriestSimAPL(t, talents, wandID, seconds, &proto.APLRotation{})
}

func newWandPriestSimAPL(t *testing.T, talents string, wandID int32, seconds int32, rotation *proto.APLRotation) (*core.Simulation, *ShadowPriest) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotRanged] = &proto.ItemSpec{Id: wandID}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(&proto.Player{
			Race:          proto.Race_RaceHuman,
			Class:         proto.Class_ClassPriest,
			Level:         60,
			Equipment:     &proto.EquipmentSpec{Items: items},
			TalentsString: talents,
			Rotation:      rotation,
			Spec:          PlayerOptionsBasic,
		}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: float64(seconds), Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*ShadowPriest)
}

func wandShoot(t *testing.T, sp *ShadowPriest) *core.Spell {
	shoot := sp.GetSpell(core.ActionID{SpellID: core.WandShootSpellID})
	if shoot == nil {
		t.Fatal("no Shoot")
	}
	return shoot
}

// TestWandShoot checks Forever's wand Shoot (wowsims/forever 7ddd02e9e2, dcc2aaf371 and
// 924140e5b2): it shoots the wand's school, the first shot comes 0.5 sec after Shoot starts
// and no sooner than a wand speed after the last cast, then one shot every wand speed; a shot
// misses on the ranged weapon table (8% against a level 63 boss, which physical hit lowers and
// spell hit doesn't) and crits on spell crit for 1.5 times.
func TestWandShoot(t *testing.T) {
	const smiteTalents = "525035231030001001-30505110002-3"
	// Bonecreeper Stylus: 83 - 155 Arcane, speed 1.9.
	const bonecreeper = 13938

	t.Run("school and Wand Specialization", func(t *testing.T) {
		_, sp := newWandPriestSim(t, smiteTalents, bonecreeper, 10)
		shoot := wandShoot(t, sp)
		if shoot.SpellSchool != core.SpellSchoolArcane {
			t.Errorf("Shoot with Bonecreeper Stylus deals school %v, want Arcane", shoot.SpellSchool)
		}
		if math.Abs(shoot.DamageMultiplier-1.25) > 1e-9 {
			t.Errorf("Wand Specialization 2/2 gives Shoot %.3f, want 1.25", shoot.DamageMultiplier)
		}
		// Wicked Lich's Finger is a physical wand, which the talent's school mask leaves out.
		_, sp = newWandPriestSim(t, smiteTalents, 281263, 10)
		if shoot := wandShoot(t, sp); shoot.SpellSchool != core.SpellSchoolPhysical || shoot.DamageMultiplier != 1 {
			t.Errorf("Shoot with a physical wand deals school %v at %.3f, want physical at 1", shoot.SpellSchool, shoot.DamageMultiplier)
		}
	})

	// The times Shoot lands, from a cast at castAt, after Smite cast at 0 when smiteFirst.
	shotTimes := func(t *testing.T, castAt float64, smiteFirst bool) []float64 {
		sim, sp := newWandPriestSim(t, smiteTalents, bonecreeper, 12)
		shoot := wandShoot(t, sp)
		target := sim.GetTargetUnit(0)
		var times []float64
		if smiteFirst {
			sim.AddPendingAction(&core.PendingAction{NextActionAt: 0, OnAction: func(sim *core.Simulation) {
				if !topRank(t, "Smite", sp.Smite).Cast(sim, target) {
					t.Fatal("Smite didn't cast")
				}
			}})
		}
		sim.AddPendingAction(&core.PendingAction{NextActionAt: core.DurationFromSeconds(castAt), OnAction: func(sim *core.Simulation) {
			if !shoot.Cast(sim, target) {
				t.Fatal("Shoot didn't cast")
			}
		}})
		// A shot counts its outcome when it fires.
		metrics := &shoot.SpellMetrics[target.UnitIndex]
		shots := int32(0)
		for !sim.Step() {
			if n := metrics.Hits + metrics.Crits + metrics.Misses; n > shots {
				shots = n
				times = append(times, sim.CurrentTime.Seconds())
			}
		}
		return times
	}
	check := func(t *testing.T, got, want []float64) {
		t.Helper()
		if len(got) < len(want) {
			t.Fatalf("Shoot landed at %v, want %v first", got, want)
		}
		for i := range want {
			if math.Abs(got[i]-want[i]) > 1e-6 {
				t.Fatalf("Shoot landed at %v, want %v first", got, want)
			}
		}
	}

	t.Run("the first shot after the wind-up, then one each wand speed", func(t *testing.T) {
		check(t, shotTimes(t, 1, false), []float64{1.5, 3.4, 5.3, 7.2})
	})

	t.Run("the first shot a wand speed after the last cast", func(t *testing.T) {
		// Smite's 2 sec cast (Divine Fury 5/5) ends at 2; the wand timer restarts there.
		check(t, shotTimes(t, 2.1, true), []float64{3.9, 5.8})
	})

	t.Run("the ranged table: physical hit lowers the miss, spell hit doesn't; spell crit for 1.5 times", func(t *testing.T) {
		shoot300 := func(bonus stats.Stats) core.SpellMetrics {
			sim, sp := newWandPriestSim(t, smiteTalents, bonecreeper, 600)
			shoot := wandShoot(t, sp)
			target := sim.GetTargetUnit(0)
			sim.AddPendingAction(&core.PendingAction{NextActionAt: 0, OnAction: func(sim *core.Simulation) {
				sp.AddStatsDynamic(sim, bonus)
				shoot.Cast(sim, target)
			}})
			for !sim.Step() {
			}
			return shoot.SpellMetrics[target.UnitIndex]
		}

		at := shotTablesFor(t)
		if math.Abs(at-0.08) > 1e-9 {
			t.Errorf("the wand's miss against a level 63 boss is %.4f, want 0.08", at)
		}
		spellHit := shoot300(stats.Stats{stats.SpellHit: 20})
		if shots := spellHit.Hits + spellHit.Crits + spellHit.Misses; shots < 300 || spellHit.Misses < 12 || spellHit.Misses > 40 {
			t.Errorf("with 20%% spell hit Shoot missed %d of %d shots, want about 8%%", spellHit.Misses, shots)
		}
		if physicalHit := shoot300(stats.Stats{stats.MeleeHit: 8}); physicalHit.Misses != 0 {
			t.Errorf("with 8%% physical hit Shoot missed %d shots, want none", physicalHit.Misses)
		}
		crit := shoot300(stats.Stats{stats.MeleeHit: 8, stats.SpellCrit: 100})
		if crit.Hits != 0 || crit.Crits == 0 {
			t.Errorf("with 100%% spell crit Shoot landed %d hits and %d crits, want only crits", crit.Hits, crit.Crits)
		}
		noCrit := shoot300(stats.Stats{stats.MeleeHit: 8, stats.MeleeCrit: 100, stats.SpellCrit: -100})
		if noCrit.Crits != 0 {
			t.Errorf("with 100%% physical crit and no spell crit Shoot crit %d times, want none", noCrit.Crits)
		}
		if ratio := (crit.TotalDamage / float64(crit.Crits)) / (noCrit.TotalDamage / float64(noCrit.Hits)); math.Abs(ratio-1.5) > 0.05 {
			t.Errorf("a crit deals %.3f times a hit, want 1.5", ratio)
		}
	})
}

// TestWandShootGivesWay checks that Shoot from the rotation ends when the rotation can cast
// something else: with Smite above it, the priest wands while out of mana and starts Smite
// within the rotation's 50 ms look once the mana is back, without waiting for the next shot.
func TestWandShootGivesWay(t *testing.T) {
	rotation := core.APLRotationFromJsonString(`{"type":"TypeAPL","priorityList":[
		{"action":{"castSpell":{"spellId":{"spellId":10933}}}},
		{"action":{"castSpell":{"spellId":{"spellId":5019}}}}
	]}`)
	sim, sp := newWandPriestSimAPL(t, "525035231030001001-30505110002-3", 13938, 20, rotation)
	shoot := wandShoot(t, sp)
	smite := sp.GetSpell(core.ActionID{SpellID: 10933})
	manaMetrics := sp.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen})
	sim.AddPendingAction(&core.PendingAction{NextActionAt: -time.Second, OnAction: func(sim *core.Simulation) {
		sp.SpendMana(sim, sp.CurrentMana(), manaMetrics)
	}})
	sim.AddPendingAction(&core.PendingAction{NextActionAt: core.DurationFromSeconds(4), OnAction: func(sim *core.Simulation) {
		if !sp.IsChanneling(sim) || sp.ChanneledDot.Spell != shoot {
			t.Error("the priest isn't wanding while out of mana")
		}
		sp.AddMana(sim, 1000, manaMetrics)
	}})
	smiteStart := time.Duration(-1)
	for !sim.Step() {
		if smiteStart < 0 && sp.IsCasting(sim) && sp.Hardcast.ActionID == smite.ActionID {
			smiteStart = sim.CurrentTime
		}
	}
	if smiteStart < core.DurationFromSeconds(4) || smiteStart > core.DurationFromSeconds(4.05) {
		t.Errorf("Smite started at %v once the mana came back at 4s, want by 4.05s", smiteStart)
	}
}

// The miss chance of the wand's ranged attack table against the default level 63 target.
func shotTablesFor(t *testing.T) float64 {
	sim, sp := newWandPriestSim(t, "", 13938, 10)
	table := sp.AttackTables[sim.GetTargetUnit(0).UnitIndex][proto.CastType_CastTypeRanged]
	if table == nil {
		t.Fatal("no ranged attack table for the wand")
	}
	return table.BaseMissChance
}
