package balance

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/druid"
)

func newMoonkinSim(t *testing.T, talents string) (*core.Simulation, *BalanceDruid) {
	return newMoonkinSimFor(t, talents, 60)
}

func newMoonkinSimFor(t *testing.T, talents string, duration float64) (*core.Simulation, *BalanceDruid) {
	player := &proto.Player{
		Race:          proto.Race_RaceTauren,
		Class:         proto.Class_ClassDruid,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec:          &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{Options: &proto.BalanceDruid_Options{}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: duration, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*BalanceDruid)
}

func topRank(spells []*druid.DruidSpell) *druid.DruidSpell {
	for i := len(spells) - 1; i >= 0; i-- {
		if spells[i] != nil {
			return spells[i]
		}
	}
	return nil
}

// TestBalanceRankValues checks the top ranks against the Forever client (1.60.1.70338):
// Wrath rank 8 costs 120 and casts in 2 sec, Starfire rank 6 costs 315 in 3.5 sec (rank 7 is an
// Ahn'Qiraj book rank, left out under Forever), Moonfire rank 10 costs 375 and ticks 4 times,
// Insect Swarm rank 5 costs 160.
func TestBalanceRankValues(t *testing.T) {
	_, moonkin := newMoonkinSim(t, "000000001")
	checks := []struct {
		name     string
		spell    *druid.DruidSpell
		cost     float64
		castTime time.Duration
	}{
		{"Wrath", topRank(moonkin.Wrath), 120, 2 * time.Second},
		{"Starfire", topRank(moonkin.Starfire), 315, 3500 * time.Millisecond},
		{"Moonfire", topRank(moonkin.Moonfire), 375, 0},
		{"Insect Swarm", topRank(moonkin.InsectSwarm), 160, 0},
	}
	for _, c := range checks {
		if c.spell == nil {
			t.Fatalf("no %s", c.name)
		}
		if got := c.spell.Cost.GetCurrentCost(); math.Abs(got-c.cost) > 1e-9 {
			t.Errorf("%s costs %.2f, want %.2f", c.name, got, c.cost)
		}
		if got := c.spell.DefaultCast.CastTime; got != c.castTime {
			t.Errorf("%s casts in %s, want %s", c.name, got, c.castTime)
		}
	}
	if ticks := topRank(moonkin.Moonfire).Dot(moonkin.CurrentTarget).NumberOfTicks; ticks != 4 {
		t.Errorf("Moonfire ticks %d times, want 4", ticks)
	}
	if moonkin.Starfire[druid.StarfireRanks] != nil {
		t.Errorf("Starfire rank %d (an Ahn'Qiraj book) is registered under Forever", druid.StarfireRanks)
	}
}

// TestBalanceWrathRankTuning checks the report assumption the user ruled on 2026-10-15: Wrath
// rank 8 deals 8.5% more than the client's 91.57 to 102.43, and rank 7 is unchanged.
func TestBalanceWrathRankTuning(t *testing.T) {
	for _, c := range []struct {
		rank      int
		low, high float64
	}{
		{8, 91.57 * 1.085, 102.43 * 1.085},
		{7, 69.34, 77.46},
	} {
		low, high := druid.WrathBaseDamageAt(c.rank, 60)
		if math.Abs(low-c.low) > 1e-9 || math.Abs(high-c.high) > 1e-9 {
			t.Errorf("Wrath rank %d at 60: %.2f to %.2f, want %.2f to %.2f", c.rank, low, high, c.low, c.high)
		}
	}
}

// TestBalanceCostAndCastTalents checks Improved Wrath (0.1 sec and 10% a point), Moonglow
// (8, 17, 25%, multiplying with Improved Wrath like every client cost modifier), Improved
// Starfire (0.1 sec a point) and Nature's Splendor (one more tick on both dots).
func TestBalanceCostAndCastTalents(t *testing.T) {
	// Improved Wrath 5, Moonglow 3, Nature's Majesty 2, Nature's Splendor, Insect Swarm,
	// Improved Starfire 5.
	_, moonkin := newMoonkinSim(t, "50302001105")
	wrath := topRank(moonkin.Wrath)
	if got, want := wrath.Cost.GetCurrentCost(), 120*0.5*0.75; math.Abs(got-want) > 1e-9 {
		t.Errorf("Wrath costs %.2f, want %.2f", got, want)
	}
	if got := wrath.DefaultCast.CastTime; got != 1500*time.Millisecond {
		t.Errorf("Wrath casts in %s, want 1.5s", got)
	}
	starfire := topRank(moonkin.Starfire)
	if got, want := starfire.Cost.GetCurrentCost(), 315*0.75; math.Abs(got-want) > 1e-9 {
		t.Errorf("Starfire costs %.2f, want %.2f", got, want)
	}
	if got := starfire.DefaultCast.CastTime; got != 3*time.Second {
		t.Errorf("Starfire casts in %s, want 3s", got)
	}
	if ticks := topRank(moonkin.Moonfire).Dot(moonkin.CurrentTarget).NumberOfTicks; ticks != 5 {
		t.Errorf("Moonfire ticks %d times with Nature's Splendor, want 5", ticks)
	}
	if ticks := topRank(moonkin.InsectSwarm).Dot(moonkin.CurrentTarget).NumberOfTicks; ticks != 7 {
		t.Errorf("Insect Swarm ticks %d times with Nature's Splendor, want 7", ticks)
	}
}

// TestBalanceEclipse checks that a Wrath that lands gives 2 Eclipse charges, at most 4, and
// each one takes 0.5 sec (Eclipse 3/3) off the next Starfire.
func TestBalanceEclipse(t *testing.T) {
	// Eclipse sits in row 5, so the build needs 20 points above it.
	sim, moonkin := newMoonkinSim(t, "55302200105003")
	starfire := topRank(moonkin.Starfire)
	base := starfire.DefaultCast.CastTime
	moonkin.EclipseAura.Activate(sim)
	moonkin.EclipseAura.SetStacks(sim, 2)
	if got := starfire.DefaultCast.CastTime; got != base-500*time.Millisecond {
		t.Errorf("Starfire with Eclipse casts in %s, want %s", got, base-500*time.Millisecond)
	}
	moonkin.EclipseAura.RemoveStack(sim)
	moonkin.EclipseAura.RemoveStack(sim)
	if got := starfire.DefaultCast.CastTime; got != base {
		t.Errorf("Starfire after Eclipse casts in %s, want %s", got, base)
	}
}

// TestClearcastingMidCast checks that a Starfire that started before Clearcasting procs pays
// its full mana cost and leaves Clearcasting up for the next spell.
func TestClearcastingMidCast(t *testing.T) {
	sim, moonkin := newMoonkinSim(t, "55302200105003")
	starfire := topRank(moonkin.Starfire)
	fullCost := starfire.Cost.GetCurrentCost()
	manaBefore := moonkin.CurrentMana()
	target := sim.GetTargetUnit(0)
	start := sim.CurrentTime
	starfire.Cast(sim, target)
	core.StartDelayedAction(sim, core.DelayedActionOptions{
		DoAt: start + time.Second,
		OnAction: func(sim *core.Simulation) {
			moonkin.ClearcastingAura.Activate(sim)
		},
	})
	for sim.CurrentTime <= start+starfire.DefaultCast.CastTime+100*time.Millisecond {
		if sim.Step() {
			break
		}
	}
	if !moonkin.ClearcastingAura.IsActive() {
		t.Errorf("Clearcasting was used up by a Starfire that started before it")
	}
	if spent := manaBefore - moonkin.CurrentMana(); math.Abs(spent-fullCost) > 10 { // less the mana regenerated during the cast
		t.Errorf("the Starfire spent %.1f mana, want its full cost %.1f", spent, fullCost)
	}
}

// TestNaturesFocus checks Nature's Focus 5 (spell 17063): a 70% chance for Wrath, Starfire and
// Hurricane to keep their cast time on a hit taken, and none without the talent or for Moonfire.
func TestNaturesFocus(t *testing.T) {
	_, moonkin := newMoonkinSim(t, "5232220115501351--5053")
	for name, spell := range map[string]*druid.DruidSpell{
		"Wrath":     topRank(moonkin.Wrath),
		"Starfire":  topRank(moonkin.Starfire),
		"Hurricane": topRank(moonkin.Hurricane),
	} {
		if spell == nil {
			t.Fatalf("no %s", name)
		}
		if got := spell.PushbackReduction; math.Abs(got-0.7) > 1e-9 {
			t.Errorf("%s pushback reduction %.2f, want 0.70", name, got)
		}
	}
	if got := topRank(moonkin.Moonfire).PushbackReduction; got != 0 {
		t.Errorf("Moonfire pushback reduction %.2f, want 0", got)
	}
	_, none := newMoonkinSim(t, "")
	if got := topRank(none.Wrath).PushbackReduction; got != 0 {
		t.Errorf("Wrath pushback reduction without the talent %.2f, want 0", got)
	}
}

// TestNaturesGraceTiming checks that under Forever Nature's Grace comes a spell batch after the
// crit's damage: a Wrath's when the bolt lands, a Moonfire's 10 ms after the cast, so the cast
// queued at the crit's instant is not hasted (wowsims/forever e003c2dca9).
func TestNaturesGraceTiming(t *testing.T) {
	sim, moonkin := newMoonkinSim(t, P1Talents)
	moonkin.DistanceFromTarget = 20
	target := sim.GetTargetUnit(0)
	naturesGrace := moonkin.NaturesGraceProcAura
	advance := func(until time.Duration) {
		sim.AddPendingAction(&core.PendingAction{NextActionAt: until, OnAction: func(*core.Simulation) {}})
		for sim.CurrentTime < until && !sim.Step() {
		}
	}
	for _, spells := range [][]*druid.DruidSpell{moonkin.Wrath, moonkin.Moonfire} {
		spell := topRank(spells)
		spell.BonusHitRating += 100 * core.SpellHitRatingPerHitChance
		spell.BonusCritRating += 200 * core.SpellCritRatingPerCritChance
	}

	wrath := topRank(moonkin.Wrath)
	start := sim.CurrentTime
	wrath.Cast(sim, target)
	castEnd := start + wrath.CurCast.CastTime
	impact := castEnd + wrath.TravelTime()
	if wrath.TravelTime() < time.Second/2 {
		t.Fatalf("Wrath travels %s, want a travel time to test", wrath.TravelTime())
	}
	advance(castEnd + time.Millisecond)
	if naturesGrace.IsActive() {
		t.Errorf("Nature's Grace up when the Wrath cast finished, want it after the bolt lands")
	}
	advance(impact + core.SpellBatchWindow - time.Millisecond)
	if naturesGrace.IsActive() {
		t.Errorf("Nature's Grace up within a spell batch of the Wrath landing")
	}
	advance(impact + core.SpellBatchWindow)
	if !naturesGrace.IsActive() {
		t.Errorf("Nature's Grace not up a spell batch after the Wrath crit landed")
	}

	naturesGrace.Deactivate(sim)
	advance(sim.CurrentTime + 2*time.Second)
	moonfire := topRank(moonkin.Moonfire)
	moonkin.SetGCDTimer(sim, sim.CurrentTime)
	if !moonfire.Cast(sim, target) {
		t.Fatalf("Moonfire not cast at %s", sim.CurrentTime)
	}
	if naturesGrace.IsActive() {
		t.Errorf("Nature's Grace up at the Moonfire crit's instant")
	}
	advance(sim.CurrentTime + core.SpellBatchWindow)
	if !naturesGrace.IsActive() {
		t.Errorf("Nature's Grace not up a spell batch after the Moonfire crit")
	}
}

// TestOmenFromFaerieFire checks that under Forever caster Faerie Fire rolls Omen of Clarity as an
// instant, 2 * 1.5 / 60 a cast that lands, doubled in Moonkin Form to 10% (wowsims/forever
// 1a2f22fe3f).
func TestOmenFromFaerieFire(t *testing.T) {
	const casts = 2000
	sim, moonkin := newMoonkinSimFor(t, P1Talents, 11*casts)
	target := sim.GetTargetUnit(0)
	faerieFire := moonkin.FaerieFire
	faerieFire.BonusHitRating += 100 * core.SpellHitRatingPerHitChance
	faerieFire.Cost.Multiplier = 0
	if !moonkin.MoonkinFormAura.IsActive() {
		t.Fatalf("not in Moonkin Form")
	}
	procs := 0
	for i := 0; i < casts; i++ {
		moonkin.SetGCDTimer(sim, sim.CurrentTime)
		moonkin.ClearcastingAura.Deactivate(sim)
		if !faerieFire.Cast(sim, target) {
			t.Fatalf("Faerie Fire not cast at %s", sim.CurrentTime)
		}
		// No auto attacks, which roll Omen of Clarity too.
		moonkin.AutoAttacks.CancelAutoSwing(sim)
		if moonkin.ClearcastingAura.IsActive() {
			procs++
		}
		until := sim.CurrentTime + 11*time.Second
		sim.AddPendingAction(&core.PendingAction{NextActionAt: until, OnAction: func(*core.Simulation) {}})
		for sim.CurrentTime < until && !sim.Step() {
		}
	}
	if rate := float64(procs) / casts; rate < 0.08 || rate > 0.125 {
		t.Errorf("Faerie Fire gave Clearcasting on %.1f%% of casts, want 10%%", 100*rate)
	}
}

// TestForeverMoonkinSwingAndHurricane checks that under Forever caster Faerie Fire leaves the
// swing timer alone (770 and 9907 have no InterruptFlags 8) and that Hurricane's ticks can't
// crit (17402 and its tick row 1278759 lack SPELL_ATTR8_PERIODIC_CAN_CRIT) but roll hit each.
func TestForeverMoonkinSwingAndHurricane(t *testing.T) {
	sim, moonkin := newMoonkinSim(t, P1Talents)
	target := sim.GetTargetUnit(0)
	swingAt := moonkin.AutoAttacks.MainhandSwingAt()
	moonkin.SetGCDTimer(sim, sim.CurrentTime)
	if !moonkin.FaerieFire.Cast(sim, target) {
		t.Fatalf("Faerie Fire not cast")
	}
	if got := moonkin.AutoAttacks.MainhandSwingAt(); got != swingAt {
		t.Errorf("Faerie Fire in Moonkin Form moved the swing from %s to %s", swingAt, got)
	}

	hurricane := topRank(moonkin.Hurricane)
	if !hurricane.Flags.Matches(core.SpellFlagNoPeriodicCrit) {
		t.Error("Hurricane ticks should not crit under Forever")
	}
	hurricane.Cost.Multiplier = 0
	moonkin.SetGCDTimer(sim, sim.CurrentTime)
	if !hurricane.Cast(sim, target) {
		t.Fatalf("Hurricane not cast")
	}
	until := sim.CurrentTime + 11*time.Second
	sim.AddPendingAction(&core.PendingAction{NextActionAt: until, OnAction: func(*core.Simulation) {}})
	for sim.CurrentTime < until && !sim.Step() {
	}
	metrics := hurricane.SpellMetrics[target.UnitIndex]
	if metrics.CritTicks != 0 {
		t.Errorf("Hurricane critted %d ticks", metrics.CritTicks)
	}
	if metrics.Ticks+metrics.Misses != 10 {
		t.Errorf("Hurricane rolled %d ticks and %d misses, want 10 rolls", metrics.Ticks, metrics.Misses)
	}
}

func newMoonkinSimWearing(t *testing.T, items map[proto.ItemSlot]int32) (*core.Simulation, *BalanceDruid) {
	specs := make([]*proto.ItemSpec, 17)
	for i := range specs {
		specs[i] = &proto.ItemSpec{Id: items[proto.ItemSlot(i)]}
	}
	player := &proto.Player{
		Race:      proto.Race_RaceTauren,
		Class:     proto.Class_ClassDruid,
		Level:     60,
		Equipment: &proto.EquipmentSpec{Items: specs},
		Rotation:  &proto.APLRotation{},
		Spec:      &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{Options: &proto.BalanceDruid_Options{}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*BalanceDruid)
}

// TestBalanceRelicsAndHaruspex checks Mystic Mushroom (249396, spell 1248751: 5% more Spirit),
// Haruspex's Garb at 5 pieces (24480: Starfire 3% more crit, 21618: 4 mana every 5 sec) and
// Talons of Wrath (249441, 1248996: a Wrath cast gives back 35 mana half the time).
func TestBalanceRelicsAndHaruspex(t *testing.T) {
	_, plain := newMoonkinSimWearing(t, nil)
	_, mushroom := newMoonkinSimWearing(t, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotRanged: druid.MysticMushroom})
	if got, want := mushroom.GetStat(stats.Spirit), math.Floor(plain.GetStat(stats.Spirit)*1.05); math.Abs(got-want) > 1e-6 {
		t.Errorf("Mystic Mushroom gives %.2f Spirit, want %.2f", got, want)
	}

	_, haruspex := newMoonkinSimWearing(t, map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotNeck:     19613,
		proto.ItemSlot_ItemSlotChest:    19838,
		proto.ItemSlot_ItemSlotWaist:    19839,
		proto.ItemSlot_ItemSlotWrist:    19840,
		proto.ItemSlot_ItemSlotTrinket1: 19955,
	})
	starfire := topRank(haruspex.Starfire)
	if got, want := starfire.BonusCritRating-topRank(plain.Starfire).BonusCritRating, 3*float64(core.SpellCritRatingPerCritChance); math.Abs(got-want) > 1e-6 {
		t.Errorf("Haruspex's Garb gives Starfire %.2f crit rating, want %.2f", got, want)
	}

	sim, talons := newMoonkinSimWearing(t, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotRanged: druid.TalonsOfWrath})
	wrath := topRank(talons.Wrath)
	metrics := talons.NewManaMetrics(core.ActionID{SpellID: 1})
	// An instant Wrath with no global cooldown, so the casts can follow one another.
	wrath.DefaultCast.CastTime = 0
	casts, procs := 400, 0
	for i := 0; i < casts; i++ {
		talons.AddMana(sim, talons.MaxMana(), metrics)
		talons.SpendMana(sim, 1000, metrics)
		talons.SetGCDTimer(sim, sim.CurrentTime)
		before, cost := talons.CurrentMana(), wrath.Cost.GetCurrentCost()
		if !wrath.Cast(sim, talons.CurrentTarget) {
			t.Fatalf("Wrath didn't cast at %v", sim.CurrentTime)
		}
		if talons.CurrentMana()-(before-cost) >= 30 {
			procs++
		}
	}
	if got := float64(procs) / float64(casts); math.Abs(got-0.5) > 0.08 {
		t.Errorf("Talons of Wrath gave back 35 mana on %.0f%% of Wraths, want 50%%", 100*got)
	}
}
