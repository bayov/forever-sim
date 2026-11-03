package tank

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/druid"
	googleProto "google.golang.org/protobuf/proto"
)

// The Forever bear's rules, ability by ability and talent by talent, against client 1.60.1.70753
// (docs/tank_druid.md has the audit). Each test builds a level 60 bear with nothing but what it
// checks, casts by hand and reads the results.

// feralOrder is the Feral tree in the talent string's order (proto/druid.proto).
var feralOrder = []string{"ferocity", "furor", "feralSwiftness", "feralInstinct", "brutalImpact", "thickHide",
	"shreddingAttacks", "savageFury", "feralCharge", "sharpenedClaws", "predatoryStrikes", "lacerate",
	"predatoryDiscipline", "bloodFrenzy", "leaderOfThePack", "naturalInstinct", "rendAndTear", "naturalReaction",
	"primalBite", "shiftingPower"}

// talents is a talent string with the named Feral talents, and resto's Restoration digits.
func talents(feral map[string]int, resto string) string {
	var b strings.Builder
	for _, name := range feralOrder {
		b.WriteByte(byte('0' + feral[name]))
	}
	return "-" + strings.TrimRight(b.String(), "0") + "-" + resto
}

// bearOpts are the knobs of a test bear.
type bearOpts struct {
	talents string
	level   int32
	idol    int32
	trinket int32
	targets int
	armor   float64 // the target's armor; -1 keeps the default boss's
	bonus   stats.Stats
	tanked  bool
	rage    float64 // starting Rage
}

// newBear builds a Tauren bear on the Feral cat's synthetic two-hander, in Dire Bear Form after the
// reset, against level 63 targets with no armor unless asked.
func newBear(t *testing.T, o bearOpts) (*core.Simulation, *druid.Druid) {
	t.Helper()
	if o.level == 0 {
		o.level = 60
	}
	if o.targets == 0 {
		o.targets = 1
	}
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 998113}
	items[proto.ItemSlot_ItemSlotChest] = &proto.ItemSpec{Id: 998105}
	if o.idol != 0 {
		items[proto.ItemSlot_ItemSlotRanged] = &proto.ItemSpec{Id: o.idol}
	}
	if o.trinket != 0 {
		items[proto.ItemSlot_ItemSlotTrinket2] = &proto.ItemSpec{Id: o.trinket}
	}
	player := &proto.Player{
		Race:          proto.Race_RaceTauren,
		Class:         proto.Class_ClassDruid,
		Level:         o.level,
		Equipment:     &proto.EquipmentSpec{Items: items},
		TalentsString: o.talents,
		Rotation:      &proto.APLRotation{},
		Spec: &proto.Player_FeralTankDruid{FeralTankDruid: &proto.FeralTankDruid{
			Options: &proto.FeralTankDruid_Options{StartingRage: o.rage}}},
	}
	if o.bonus != (stats.Stats{}) {
		player.BonusStats = &proto.UnitStats{Stats: o.bonus[:]}
	}
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	if o.tanked {
		raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	}
	target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
	target.Level = 63
	if o.armor >= 0 {
		target.Stats[stats.Armor] = o.armor
	}
	// A boss that hits softly, so a tanked test bear on two items lives through the test.
	target.MinBaseDamage = 100
	var targets []*proto.Target
	for i := 0; i < o.targets; i++ {
		targets = append(targets, target)
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Duration: 300, Targets: targets},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 3, ExactDamage: true},
	}, simsignals.CreateSignals())
	sim.Reset()
	bear := sim.Raid.Parties[0].Players[0].(*FeralTankDruid).GetDruid()
	if !bear.InForm(druid.Bear) {
		t.Fatal("the tank isn't in Bear Form after the reset")
	}
	return sim, bear
}

// advance runs the sim to time at.
func advance(sim *core.Simulation, at time.Duration) {
	sim.AddPendingAction(&core.PendingAction{NextActionAt: at, OnAction: func(*core.Simulation) {}})
	for sim.CurrentTime < at && !sim.Step() {
	}
}

// ready clears the global cooldown and the spell's own.
func ready(bear *druid.Druid, spell *druid.DruidSpell) {
	bear.GCD.Reset()
	if spell.CD.Timer != nil {
		spell.CD.Reset()
	}
}

// hit is one landed result of a spell.
type hit struct {
	outcome        core.HitOutcome
	damage, threat float64
	tick           bool
}

// recorderItem is the test bear's chest (the Feral cat's synthetic one): its effect in these tests
// keeps every result the bear deals, since an aura can't be added once the sim is built.
const recorderItem = 998105

type recordedHit struct {
	id core.ActionID
	hit
}

var recorded = map[*core.Unit]*[]recordedHit{}

// dodged counts the attacks the bear dodged.
var dodged = map[*core.Unit]*int{}

func init() {
	core.NewItemEffect(recorderItem, func(agent core.Agent) {
		unit := &agent.GetCharacter().Unit
		all := &[]recordedHit{}
		recorded[unit] = all
		dodges := new(int)
		dodged[unit] = dodges
		core.MakePermanent(unit.RegisterAura(core.Aura{
			Label: "Test Recorder",
			OnSpellHitTaken: func(_ *core.Aura, _ *core.Simulation, _ *core.Spell, result *core.SpellResult) {
				if result.Outcome.Matches(core.OutcomeDodge) {
					*dodges++
				}
			},
			OnSpellHitDealt: func(_ *core.Aura, _ *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				*all = append(*all, recordedHit{spell.ActionID, hit{result.Outcome, result.Damage, result.Threat, false}})
			},
			OnPeriodicDamageDealt: func(_ *core.Aura, _ *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				*all = append(*all, recordedHit{spell.ActionID, hit{result.Outcome, result.Damage, result.Threat, true}})
			},
		}))
	})
}

// record gives the results the bear deals with the spell's ID from now on.
func record(bear *druid.Druid, id core.ActionID) func() []hit {
	all := recorded[&bear.Unit]
	from := len(*all)
	return func() []hit {
		var out []hit
		for _, h := range (*all)[from:] {
			if h.id.SameActionIgnoreTag(id) {
				out = append(out, h.hit)
			}
		}
		return out
	}
}

// cast casts the spell n times at the target, clearing its cooldowns and topping Rage up to 100
// before each, and returns the Rage each cast spent.
func castN(t *testing.T, sim *core.Simulation, bear *druid.Druid, spell *druid.DruidSpell, n int) []float64 {
	t.Helper()
	spent := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		// A moment between casts, so a Clearcasting that the last one gave counts as older than this one.
		advance(sim, sim.CurrentTime+10*time.Millisecond)
		// No Clearcasting from the last one: TestBearClearcasting has it.
		bear.ClearcastingAura.Deactivate(sim)
		topUp(sim, bear)
		ready(bear, spell)
		before := bear.CurrentRage()
		if !spell.Cast(sim, bear.CurrentTarget) {
			t.Fatalf("%s didn't cast", spell.ActionID)
		}
		spent = append(spent, before-bear.CurrentRage())
	}
	return spent
}

var testRage *core.ResourceMetrics

func topUp(sim *core.Simulation, bear *druid.Druid) {
	if testRage == nil || testRage.ActionID.SpellID != 1 {
		testRage = bear.NewRageMetrics(core.ActionID{SpellID: 1})
	}
	bear.AddRage(sim, 100-bear.CurrentRage(), testRage)
}

func near(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

// rageFrom is the Rage the spell's metrics booked this iteration (spent Rage is negative).
func rageFrom(bear *druid.Druid, spellID int32) float64 {
	total := 0.0
	for _, m := range bear.Metrics.ToProto().Resources {
		if m.Id.GetSpellId() == spellID && m.Type == proto.ResourceType_ResourceTypeRage {
			total += m.Gain
		}
	}
	return total
}

// itemArmor is the armor on the bear's items (the chest only here).
func itemArmor(bear *druid.Druid) float64 {
	return bear.Equipment[proto.ItemSlot_ItemSlotChest].Stats[stats.Armor]
}

// weaponRange is the bear's weapon damage range at its attack power, before multipliers.
func weaponRange(bear *druid.Druid) (lo, hi float64) {
	w := bear.AutoAttacks.MH()
	ap := bear.GetStat(stats.AttackPower) * w.SwingSpeed / core.DefaultAttackPowerPerDPS
	return w.BaseDamageMin + ap, w.BaseDamageMax + ap
}

// plainHits are the results that landed as plain hits (no crit, glance, block or tick).
func plainHits(hits []hit) []hit {
	var out []hit
	for _, h := range hits {
		if h.outcome == core.OutcomeHit && !h.tick {
			out = append(out, h)
		}
	}
	return out
}

// TestBearForm: Dire Bear Form (9634, its passive 9635) at 60 gives 1,240 health, 180 attack power,
// 4.6 times the armor from items, 1.5 times the threat (21178), the Rage bar, and a 2.5 sec swing.
func TestBearForm(t *testing.T) {
	sim, bear := newBear(t, bearOpts{armor: 0})
	health, ap, armor := bear.MaxHealth(), bear.GetStat(stats.AttackPower), bear.GetStat(stats.Armor)
	threat := bear.PseudoStats.ThreatMultiplier
	if got := bear.AutoAttacks.MH().SwingSpeed; got != 2.5 {
		t.Errorf("the bear swings every %.1f sec, want 2.5", got)
	}
	bear.CancelShapeshift(sim)
	if bear.InForm(druid.Bear) {
		t.Fatal("CancelShapeshift left the bear in its form")
	}
	// The Tauren's Endurance (5% health) raises it.
	if got := health - bear.MaxHealth(); !near(got, 1240*1.05, 0.5) {
		t.Errorf("Dire Bear Form gives %.0f health, want 1,240 (1,302 for a Tauren)", got)
	}
	if got := ap - bear.GetStat(stats.AttackPower); !near(got, 180, 0.5) {
		t.Errorf("Dire Bear Form gives %.0f attack power, want 180", got)
	}
	if got := (armor - bear.GetStat(stats.Armor)) / itemArmor(bear); !near(got, 3.6, 0.01) {
		t.Errorf("Dire Bear Form adds %.3f times the armor from items, want 3.6 (+360%%)", got)
	}
	if got := threat / bear.PseudoStats.ThreatMultiplier; !near(got, 1.5, 1e-9) {
		t.Errorf("Bear Form's threat multiplier is %.3f, want 1.5", got)
	}
}

// TestBearFormLevels: Bear Form below 40 (1178: 180%, 20 + 18 a level of health, 30 + 3 a level of
// attack power from 10).
func TestBearFormLevels(t *testing.T) {
	sim, bear := newBear(t, bearOpts{level: 30})
	health, armor := bear.MaxHealth(), bear.GetStat(stats.Armor)
	bear.CancelShapeshift(sim)
	if got := health - bear.MaxHealth(); !near(got, (20+18*20)*1.05, 0.5) {
		t.Errorf("Bear Form at 30 gives %.0f health, want 380 (399 for a Tauren)", got)
	}
	if got := (armor - bear.GetStat(stats.Armor)) / itemArmor(bear); !near(got, 1.8, 0.01) {
		t.Errorf("Bear Form adds %.3f times the armor from items, want 1.8 (+180%%)", got)
	}
}

// TestMaul: rank 7 (9881) is the weapon's damage plus 128 on the next swing, 15 Rage (Ferocity 1 a
// point less, Idol of Brutality 2 less), threat its damage times the form's 1.5.
func TestMaul(t *testing.T) {
	sim, bear := newBear(t, bearOpts{armor: 0, rage: 100})
	hits := record(bear, bear.Maul.ActionID)
	lo, hi := weaponRange(bear)
	if !bear.MaulQueueSpell.Cast(sim, bear.CurrentTarget) {
		t.Fatal("Maul didn't queue")
	}
	advance(sim, 30*time.Second)
	if len(hits()) == 0 {
		t.Fatal("no Maul landed on the swing after it was queued")
	}
	if len(hits()) != 1 {
		t.Errorf("one queued Maul struck %d times", len(hits()))
	}
	// Then Maul on every swing.
	for i := 0; i < 200; i++ {
		topUp(sim, bear)
		bear.MaulQueueSpell.Cast(sim, bear.CurrentTarget)
		advance(sim, sim.CurrentTime+2500*time.Millisecond)
	}
	plain := plainHits(hits())
	if len(plain) < 50 {
		t.Fatalf("only %d plain Maul hits in 200 swings", len(plain))
	}
	for _, h := range plain {
		if h.damage < 128+lo-1 || h.damage > 128+hi+1 {
			t.Fatalf("a Maul hit for %.0f, want 128 plus the weapon's %.0f to %.0f", h.damage, lo, hi)
		}
		if !near(h.threat, 1.5*h.damage, 0.01) {
			t.Fatalf("a Maul hit for %.0f made %.0f threat, want 1.5 times its damage", h.damage, h.threat)
		}
	}
	if got := bear.Maul.DefaultCast.Cost; got != 15 {
		t.Errorf("Maul costs %.0f Rage, want 15", got)
	}
	_, ferocious := newBear(t, bearOpts{talents: talents(map[string]int{"ferocity": 5}, ""), idol: druid.IdolOfBrutality})
	if got := ferocious.Maul.DefaultCast.Cost; got != 8 {
		t.Errorf("Maul costs %.0f Rage with Ferocity 5/5 and Idol of Brutality, want 8", got)
	}
}

// TestSwipe: rank 5 (9908) is 83 plus 3% of attack power (wowsims/forever's fit) to up to 3 enemies,
// 20 Rage; Feral Instinct 10% a point more damage, Savage Fury 5% a point (additive), Idol of
// Brutality 2 Rage less.
func TestSwipe(t *testing.T) {
	sim, bear := newBear(t, bearOpts{armor: 0, targets: 4})
	hits := record(bear, bear.SwipeBear.ActionID)
	spent := castN(t, sim, bear, bear.SwipeBear, 1)
	if spent[0] != 20 {
		t.Errorf("Swipe spent %.0f Rage, want 20", spent[0])
	}
	if len(hits()) != 3 {
		t.Errorf("Swipe hit %d of 4 enemies, want 3", len(hits()))
	}
	castN(t, sim, bear, bear.SwipeBear, 100)
	want := 83 + 0.03*bear.GetStat(stats.AttackPower)
	for _, h := range plainHits(hits()) {
		if !near(h.damage, want, 0.5) {
			t.Fatalf("a Swipe hit for %.1f, want 83 + 3%% of attack power, %.1f", h.damage, want)
		}
	}

	sim, bear = newBear(t, bearOpts{armor: 0, talents: talents(map[string]int{"feralInstinct": 3, "savageFury": 2}, ""), idol: druid.IdolOfBrutality})
	hits = record(bear, bear.SwipeBear.ActionID)
	if spent := castN(t, sim, bear, bear.SwipeBear, 100); spent[0] != 18 {
		t.Errorf("Swipe with Idol of Brutality spent %.0f Rage, want 18", spent[0])
	}
	want = (83 + 0.03*bear.GetStat(stats.AttackPower)) * (1.3 * 1.1)
	for _, h := range plainHits(hits()) {
		if !near(h.damage, want, 0.5) {
			t.Fatalf("a Swipe hit with Feral Instinct 3/3 and Savage Fury 2/2 for %.1f, want %.1f", h.damage, want)
		}
	}
}

// TestDemoralizingRoar: rank 5 (9898) cuts every enemy's attack power by 204 at 60 for 30 sec, 10
// Rage, and makes 42 + 0.8 a level from 52 threat a target (48.4 at 60), times the form's 1.5.
func TestDemoralizingRoar(t *testing.T) {
	sim, bear := newBear(t, bearOpts{targets: 2})
	hits := record(bear, bear.DemoralizingRoar.ActionID)
	boss := sim.Encounter.ActiveTargetUnits[0]
	ap := boss.GetStat(stats.AttackPower)
	if spent := castN(t, sim, bear, bear.DemoralizingRoar, 1); spent[0] != 10 {
		t.Errorf("Demoralizing Roar spent %.0f Rage, want 10", spent[0])
	}
	if got := ap - boss.GetStat(stats.AttackPower); !near(got, 204, 1) {
		t.Errorf("Demoralizing Roar cut the boss's attack power by %.0f, want 204", got)
	}
	aura := bear.DemoralizingRoarAuras.Get(boss)
	if aura.Duration != 30*time.Second {
		t.Errorf("Demoralizing Roar lasts %s, want 30s", aura.Duration)
	}
	if !bear.DemoralizingRoarAuras.Get(sim.Encounter.ActiveTargetUnits[1]).IsActive() {
		t.Error("Demoralizing Roar missed the second enemy")
	}
	for _, h := range hits() {
		if h.outcome.Matches(core.OutcomeLanded) && !near(h.threat, 48.4*1.5, 0.01) {
			t.Errorf("Demoralizing Roar made %.2f threat on a target, want 48.4 x 1.5", h.threat)
		}
	}
}

// TestLacerate: rank 5 (1235827 at 58): 15 Rage (Shredding Attacks 1 a point less), a bleed of 15 a
// stack every 3 sec for 15 sec, 5 stacks, a hit of 10% of weapon damage a stack already on; Rend
// and Tear on its ticks.
func TestLacerate(t *testing.T) {
	sim, bear := newBear(t, bearOpts{armor: 0, talents: talents(map[string]int{"lacerate": 1}, "")})
	if bear.Lacerate == nil || bear.Lacerate.ActionID.SpellID != 1235827 {
		t.Fatalf("Lacerate at 60 is %v, want rank 5 (1235827)", bear.Lacerate)
	}
	hits := record(bear, bear.Lacerate.ActionID)
	lo, hi := weaponRange(bear)
	dot := bear.Lacerate.Dot(bear.CurrentTarget)
	for stacks := 1; stacks <= 6; {
		before := len(hits())
		spent := castN(t, sim, bear, bear.Lacerate, 1)
		h := hits()[before]
		if !h.outcome.Matches(core.OutcomeLanded) {
			continue
		}
		if spent[0] != 15 {
			t.Errorf("Lacerate spent %.0f Rage, want 15", spent[0])
		}
		existing := float64(min(stacks, 6) - 1)
		if h.outcome == core.OutcomeHit && (h.damage < 0.1*existing*lo-0.5 || h.damage > 0.1*existing*hi+0.5) {
			t.Errorf("Lacerate with %0.f stacks on hit for %.0f, want 10%% of the weapon's %.0f to %.0f a stack", existing, h.damage, lo, hi)
		}
		if got, want := dot.GetStacks(), int32(min(stacks, 5)); got != want {
			t.Errorf("after %d Lacerates the bleed has %d stacks, want %d", stacks, got, want)
		}
		stacks++
	}
	if dot.TickLength != 3*time.Second || dot.NumberOfTicks != 5 {
		t.Errorf("Lacerate ticks %d times every %s, want 5 every 3s", dot.NumberOfTicks, dot.TickLength)
	}
	ticksFrom := len(hits())
	advance(sim, sim.CurrentTime+16*time.Second)
	var ticks []hit
	for _, h := range hits()[ticksFrom:] {
		if h.tick {
			ticks = append(ticks, h)
		}
	}
	if len(ticks) != 5 {
		t.Errorf("Lacerate ticked %d times after the last one, want 5", len(ticks))
	}
	for _, h := range ticks {
		if !near(h.damage, 75, 0.01) {
			t.Errorf("a 5 stack Lacerate tick did %.2f, want 75", h.damage)
		}
	}

	sim, bear = newBear(t, bearOpts{armor: 0, talents: talents(map[string]int{"lacerate": 1, "shreddingAttacks": 3, "rendAndTear": 5}, "")})
	hits = record(bear, bear.Lacerate.ActionID)
	for bear.Lacerate.Dot(bear.CurrentTarget).GetStacks() < 5 {
		if spent := castN(t, sim, bear, bear.Lacerate, 1); !near(spent[0], 12, 1e-9) && !near(spent[0], 12*0.2, 1e-9) {
			t.Errorf("Lacerate with Shredding Attacks 3/3 spent %.1f Rage, want 12", spent[0])
		}
	}
	castN(t, sim, bear, bear.Lacerate, 1)
	from := len(hits())
	advance(sim, sim.CurrentTime+4*time.Second)
	for _, h := range hits()[from:] {
		if h.tick && !near(h.damage, 75*1.1, 0.01) {
			t.Errorf("a 5 stack Lacerate tick with Rend and Tear 5/5 did %.2f, want 82.5", h.damage)
		}
	}
}

// TestLacerateRanks: 1328617 at 25 (4 a tick), 1328616 at 34 (6), 414644 at 42 (10), 1235826 at 50
// (12), 1235827 at 58 (15).
func TestLacerateRanks(t *testing.T) {
	for _, c := range []struct {
		level int32
		id    int32
	}{{25, 1328617}, {34, 1328616}, {42, 414644}, {50, 1235826}, {58, 1235827}} {
		_, bear := newBear(t, bearOpts{level: c.level, talents: talents(map[string]int{"lacerate": 1}, "")})
		if bear.Lacerate == nil || bear.Lacerate.ActionID.SpellID != c.id {
			t.Errorf("Lacerate at %d is %v, want %d", c.level, bear.Lacerate, c.id)
		}
	}
}

// TestPrimalBite: rank 3 (1238070) at 60: the weapon's damage plus 77, 20 Rage (Ferocity 1 a point
// less), 6 sec cooldown; 407995 (42) at 40 and 1238069 (62) at 50; 80% of its Rage back when it
// fails.
func TestPrimalBite(t *testing.T) {
	for _, c := range []struct {
		level int32
		id    int32
	}{{40, 407995}, {50, 1238069}, {60, 1238070}} {
		_, bear := newBear(t, bearOpts{level: c.level, talents: talents(map[string]int{"primalBite": 1}, "")})
		if bear.PrimalBite == nil || bear.PrimalBite.ActionID.SpellID != c.id {
			t.Errorf("Primal Bite at %d is %v, want %d", c.level, bear.PrimalBite, c.id)
		}
	}
	sim, bear := newBear(t, bearOpts{armor: 0, talents: talents(map[string]int{"primalBite": 1}, "")})
	if got := bear.PrimalBite.CD.Duration; got != 6*time.Second {
		t.Errorf("Primal Bite's cooldown is %s, want 6s", got)
	}
	hits := record(bear, bear.PrimalBite.ActionID)
	lo, hi := weaponRange(bear)
	spent := castN(t, sim, bear, bear.PrimalBite, 400)
	refunded := false
	for i, h := range hits() {
		switch {
		case h.outcome == core.OutcomeHit:
			if spent[i] != 20 {
				t.Fatalf("a Primal Bite that hit spent %.1f Rage, want 20", spent[i])
			}
			if h.damage < 77+lo-1 || h.damage > 77+hi+1 {
				t.Fatalf("a Primal Bite hit for %.0f, want 77 plus the weapon's %.0f to %.0f", h.damage, lo, hi)
			}
			if !near(h.threat, 1.5*h.damage, 0.01) {
				t.Fatalf("a Primal Bite made %.0f threat for %.0f damage, want 1.5 times", h.threat, h.damage)
			}
		case h.outcome.Matches(core.OutcomeMiss | core.OutcomeDodge | core.OutcomeParry):
			refunded = true
			if !near(spent[i], 4, 1e-9) {
				t.Fatalf("a Primal Bite that failed spent %.1f Rage, want 20 less 80%% back", spent[i])
			}
		}
	}
	if !refunded {
		t.Error("no Primal Bite failed in 400 casts against a level 63 boss")
	}
	_, ferocious := newBear(t, bearOpts{talents: talents(map[string]int{"primalBite": 1, "ferocity": 5}, "")})
	if got := ferocious.PrimalBite.DefaultCast.Cost; got != 15 {
		t.Errorf("Primal Bite costs %.0f Rage with Ferocity 5/5, want 15", got)
	}
}

// TestEnrage: 5229 gives 10 Rage at once and 2 a second for 10 sec (30 in all), off the global
// cooldown on a 1 min cooldown, with 16% less armor from items in Dire Bear Form while it lasts;
// Enraged Idol makes the ticks 3.
func TestEnrage(t *testing.T) {
	for _, c := range []struct {
		idol int32
		want float64
	}{{0, 30}, {druid.EnragedIdol, 40}} {
		sim, bear := newBear(t, bearOpts{idol: c.idol})
		armor := bear.GetStat(stats.Armor)
		if !bear.Enrage.Cast(sim, bear.CurrentTarget) {
			t.Fatal("Enrage didn't cast")
		}
		if bear.Enrage.DefaultCast.GCD != 0 {
			t.Error("Enrage is on the global cooldown")
		}
		if got := bear.Enrage.CD.Duration; got != time.Minute {
			t.Errorf("Enrage's cooldown is %s, want 1m", got)
		}
		if got := bear.CurrentRage(); got != 10 {
			t.Errorf("Enrage gave %.0f Rage at once, want 10", got)
		}
		if got := (armor - bear.GetStat(stats.Armor)) / (4.6 * itemArmor(bear)); !near(got, 0.16, 0.001) {
			t.Errorf("Enrage takes %.3f of Dire Bear Form's armor from items, want 0.16", got)
		}
		advance(sim, 10*time.Second+time.Millisecond)
		if got := rageFrom(bear, 5229); got != c.want {
			t.Errorf("Enrage with idol %d gave %.0f Rage in all, want %.0f", c.idol, got, c.want)
		}
		if bear.EnrageAura.IsActive() || !near(bear.GetStat(stats.Armor), armor, 0.01) {
			t.Error("Enrage's armor cut outlasted its 10 sec")
		}
	}
}

// TestFrenziedRegeneration: 22842 turns up to 10 Rage a second into 1% of max health a point for 10
// sec, on a 3 min cooldown and the global cooldown.
func TestFrenziedRegeneration(t *testing.T) {
	sim, bear := newBear(t, bearOpts{rage: 100})
	bear.RemoveHealth(sim, bear.MaxHealth()*0.9)
	health := bear.CurrentHealth()
	if !bear.FrenziedRegeneration.Cast(sim, bear.CurrentTarget) {
		t.Fatal("Frenzied Regeneration didn't cast")
	}
	if bear.FrenziedRegeneration.CD.Duration != 3*time.Minute || bear.FrenziedRegeneration.DefaultCast.GCD != core.GCDDefault {
		t.Errorf("Frenzied Regeneration: cooldown %s and GCD %s, want 3m and 1.5s",
			bear.FrenziedRegeneration.CD.Duration, bear.FrenziedRegeneration.DefaultCast.GCD)
	}
	advance(sim, 3*time.Second+time.Millisecond)
	if got := rageFrom(bear, 22842); got != -30 {
		t.Errorf("3 sec of Frenzied Regeneration took %.0f Rage, want 30", -got)
	}
	if got := (bear.CurrentHealth() - health) / bear.MaxHealth(); !near(got, 0.30, 0.001) {
		t.Errorf("3 sec of Frenzied Regeneration healed %.3f of max health, want 0.30", got)
	}
}

// TestBarkskin: 22812, cast in Bear Form: 20% less Physical damage taken for 15 sec, other schools
// unchanged, 1 min cooldown, the global cooldown, no cost.
func TestBarkskin(t *testing.T) {
	sim, bear := newBear(t, bearOpts{})
	physical := bear.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical]
	fire := bear.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire]
	if !bear.Barkskin.Cast(sim, bear.CurrentTarget) {
		t.Fatal("Barkskin didn't cast in Bear Form")
	}
	if !bear.InForm(druid.Bear) {
		t.Error("Barkskin took the bear out of its form")
	}
	if got := bear.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical] / physical; !near(got, 0.8, 1e-9) {
		t.Errorf("Barkskin leaves %.2f of the Physical damage, want 0.8", got)
	}
	if bear.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire] != fire || bear.PseudoStats.DamageTakenMultiplier != 1 {
		t.Error("Barkskin cut damage of another school")
	}
	if bear.Barkskin.CD.Duration != time.Minute || bear.Barkskin.DefaultCast.GCD != core.GCDDefault || bear.BarkskinAura.Duration != 15*time.Second {
		t.Errorf("Barkskin: cooldown %s, GCD %s, %s, want 1m, 1.5s and 15s", bear.Barkskin.CD.Duration, bear.Barkskin.DefaultCast.GCD, bear.BarkskinAura.Duration)
	}
	advance(sim, 15*time.Second+time.Millisecond)
	if bear.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical] != physical {
		t.Error("Barkskin's cut outlasted it")
	}
}

// TestBash: 8983 at 46: a 4 sec stun, 10 Rage, 1 min cooldown; Brutal Impact 0.5 sec more stun and
// 15 sec less cooldown a point.
func TestBash(t *testing.T) {
	for _, c := range []struct {
		impact   int
		cooldown time.Duration
		stun     time.Duration
	}{{0, time.Minute, 4 * time.Second}, {2, 30 * time.Second, 5 * time.Second}} {
		sim, bear := newBear(t, bearOpts{talents: talents(map[string]int{"brutalImpact": c.impact}, "")})
		if bear.Bash.ActionID.SpellID != 8983 {
			t.Errorf("Bash at 60 is %d, want 8983", bear.Bash.ActionID.SpellID)
		}
		if got := bear.Bash.CD.Duration; got != c.cooldown {
			t.Errorf("Bash with Brutal Impact %d has a %s cooldown, want %s", c.impact, got, c.cooldown)
		}
		landed := false
		for i := 0; i < 20 && !landed; i++ {
			spent := castN(t, sim, bear, bear.Bash, 1)
			if aura := sim.Encounter.ActiveTargetUnits[0].GetAuraByID(bear.Bash.ActionID); aura != nil && aura.IsActive() {
				landed = true
				if spent[0] != 10 {
					t.Errorf("Bash spent %.0f Rage, want 10", spent[0])
				}
				if aura.Duration != c.stun {
					t.Errorf("Bash with Brutal Impact %d stuns for %s, want %s", c.impact, aura.Duration, c.stun)
				}
			}
		}
		if !landed {
			t.Error("no Bash landed in 20 casts")
		}
	}
}

// TestBearShiftFuror: a shift into Bear Form (from Bear Form too) starts the Rage at 0, and Furor
// gives 10 on top, 20% of the time a point (17056, 17057).
func TestBearShiftFuror(t *testing.T) {
	for _, c := range []struct {
		furor int
		want  float64
	}{{0, 0}, {5, 10}} {
		sim, bear := newBear(t, bearOpts{rage: 50, talents: talents(map[string]int{"furor": c.furor}, "")})
		if !bear.BearForm.Cast(sim, bear.CurrentTarget) {
			t.Fatal("Bear Form didn't cast from Bear Form")
		}
		if got := bear.CurrentRage(); got != c.want {
			t.Errorf("a shift with Furor %d/5 left %.0f Rage of 50, want %.0f", c.furor, got, c.want)
		}
		if !bear.InForm(druid.Bear) {
			t.Error("the shift left the bear out of its form")
		}
	}
	// 2/5 is a 40% chance.
	sim, bear := newBear(t, bearOpts{talents: talents(map[string]int{"furor": 2}, "")})
	procs := 0
	for i := 0; i < 2000; i++ {
		ready(bear, bear.BearForm)
		bear.AddMana(sim, bear.MaxMana(), bear.NewManaMetrics(core.ActionID{SpellID: 2}))
		bear.BearForm.Cast(sim, bear.CurrentTarget)
		if bear.CurrentRage() == 10 {
			procs++
		}
	}
	if got := float64(procs) / 2000; !near(got, 0.4, 0.04) {
		t.Errorf("Furor 2/5 gave Rage on %.1f%% of shifts, want 40%%", 100*got)
	}
}

// TestBearClearcasting: Clearcasting (16870) makes the bear's attacks free and goes with the first:
// Lacerate, Primal Bite, Swipe, Demoralizing Roar and Maul are in its class mask.
func TestBearClearcasting(t *testing.T) {
	sim, bear := newBear(t, bearOpts{talents: talents(map[string]int{"lacerate": 1, "primalBite": 1}, "")})
	for _, spell := range []*druid.DruidSpell{bear.Lacerate, bear.PrimalBite, bear.SwipeBear, bear.DemoralizingRoar, bear.Bash} {
		bear.ClearcastingAura.Activate(sim)
		advance(sim, sim.CurrentTime+10*time.Millisecond)
		topUp(sim, bear)
		ready(bear, spell)
		before := bear.CurrentRage()
		if !spell.Cast(sim, bear.CurrentTarget) {
			t.Fatalf("%s didn't cast", spell.ActionID)
		}
		if got := before - bear.CurrentRage(); got != 0 {
			t.Errorf("%s spent %.0f Rage under Clearcasting, want 0", spell.ActionID, got)
		}
		// A hit can proc a new one (the test didn't start Omen's cooldown), which starts it afresh.
		if bear.ClearcastingAura.IsActive() && bear.ClearcastingAura.RemainingDuration(sim) < bear.ClearcastingAura.Duration {
			t.Errorf("%s left Clearcasting up", spell.ActionID)
		}
		bear.ClearcastingAura.Deactivate(sim)
	}
	topUp(sim, bear)
	bear.ClearcastingAura.Activate(sim)
	bear.MaulQueueSpell.Cast(sim, bear.CurrentTarget)
	before := bear.CurrentRage()
	hits := record(bear, bear.Maul.ActionID)
	advance(sim, sim.CurrentTime+3*time.Second)
	if len(hits()) != 1 {
		t.Fatalf("%d Mauls landed on the next swing, want 1", len(hits()))
	}
	if got := before - bear.CurrentRage(); got > 0 {
		t.Errorf("Maul spent %.0f Rage under Clearcasting, want 0", got)
	}
}

// TestBearOmenOfClarity: Omen of Clarity procs from the bear's melee, 2 a minute at its 2.5 sec swing
// (8.3% a hit), with its 10 sec cooldown.
func TestBearOmenOfClarity(t *testing.T) {
	sim, bear := newBear(t, bearOpts{armor: 0})
	procs := 0
	bear.ClearcastingAura.ApplyOnGain(func(*core.Aura, *core.Simulation) { procs++ })
	advance(sim, 300*time.Second)
	if procs < 3 {
		t.Errorf("Clearcasting procced %d times in 300 sec of bear swings, want a few", procs)
	}
}

// statDelta is a stat of a bear with the talents less one without, both in their form.
func statDelta(t *testing.T, stat stats.Stat, with, without bearOpts) float64 {
	_, a := newBear(t, with)
	_, b := newBear(t, without)
	return a.GetStat(stat) - b.GetStat(stat)
}

// TestBearTalentStats: Feral Swiftness 2% dodge a point, Natural Reaction 1% a point, Sharpened Claws
// 3% crit a point, Predatory Strikes 50% of the level as attack power a point, Leader of the Pack 3%
// crit (its raid aura), Heart of the Wild 4% Stamina a point in the bear forms, Thick Hide 1 base
// armor a level and 0.67 a point of Defense a point, both times the form's 4.6.
func TestBearTalentStats(t *testing.T) {
	none := bearOpts{armor: 0}
	with := func(name string, points int) bearOpts {
		return bearOpts{armor: 0, talents: talents(map[string]int{name: points}, "")}
	}
	if got := statDelta(t, stats.Dodge, with("feralSwiftness", 2), none) / core.DodgeRatingPerDodgeChance; !near(got, 4, 1e-6) {
		t.Errorf("Feral Swiftness 2/2 adds %.2f%% dodge, want 4", got)
	}
	if got := statDelta(t, stats.Dodge, with("naturalReaction", 5), none) / core.DodgeRatingPerDodgeChance; !near(got, 5, 1e-6) {
		t.Errorf("Natural Reaction 5/5 adds %.2f%% dodge, want 5", got)
	}
	if got := statDelta(t, stats.MeleeCrit, with("sharpenedClaws", 2), none) / core.CritRatingPerCritChance; !near(got, 6, 1e-6) {
		t.Errorf("Sharpened Claws 2/2 adds %.2f%% crit, want 6", got)
	}
	if got := statDelta(t, stats.AttackPower, with("predatoryStrikes", 3), none); !near(got, 90, 0.5) {
		t.Errorf("Predatory Strikes 3/3 adds %.1f attack power at 60, want 90", got)
	}
	if got := statDelta(t, stats.MeleeCrit, with("leaderOfThePack", 1), none) / core.CritRatingPerCritChance; !near(got, 3, 1e-6) {
		t.Errorf("Leader of the Pack adds %.2f%% crit, want 3", got)
	}
	_, plain := newBear(t, none)
	if got := statDelta(t, stats.Stamina, bearOpts{armor: 0, talents: talents(nil, "05")}, none); !near(got, 0.2*plain.GetStat(stats.Stamina), 1) {
		t.Errorf("Heart of the Wild 5/5 adds %.0f Stamina to %.0f in Bear Form, want 20%%", got, plain.GetStat(stats.Stamina))
	}
	defense := 20.0
	thick := bearOpts{armor: 0, talents: talents(map[string]int{"thickHide": 3}, ""), bonus: stats.Stats{stats.Defense: defense}}
	if got, want := statDelta(t, stats.Armor, thick, bearOpts{armor: 0, bonus: stats.Stats{stats.Defense: defense}}), (3*60+2.0*defense)*4.6; !near(got, want, 1) {
		t.Errorf("Thick Hide 3/3 with %.0f Defense adds %.0f armor in Dire Bear Form, want %.0f", defense, got, want)
	}
}

// TestBearTalentDamage: Savage Fury 5% a point on Maul, Naturalist 2% a point on everything, Rend and
// Tear 2% a point on a bleeding target, Natural Instinct 10% a point more crit damage bonus on the
// bear's attacks.
func TestBearTalentDamage(t *testing.T) {
	mean := func(o bearOpts, bleed bool) float64 {
		sim, bear := newBear(t, o)
		hits := record(bear, bear.Maul.ActionID)
		lacerate := bear.Lacerate
		for i := 0; i < 300; i++ {
			topUp(sim, bear)
			if bleed && !lacerate.Dot(bear.CurrentTarget).IsActive() {
				ready(bear, lacerate)
				lacerate.Cast(sim, bear.CurrentTarget)
			}
			bear.MaulQueueSpell.Cast(sim, bear.CurrentTarget)
			advance(sim, sim.CurrentTime+2500*time.Millisecond)
		}
		sum, n := 0.0, 0.0
		for _, h := range plainHits(hits()) {
			sum += h.damage
			n++
		}
		return sum / n
	}
	base := mean(bearOpts{armor: 0}, false)
	if got := mean(bearOpts{armor: 0, talents: talents(map[string]int{"savageFury": 2}, "")}, false) / base; !near(got, 1.1, 0.02) {
		t.Errorf("Savage Fury 2/2 moves Maul by %.3f, want 1.1", got)
	}
	if got := mean(bearOpts{armor: 0, talents: talents(nil, "005")}, false) / base; !near(got, 1.1, 0.02) {
		t.Errorf("Naturalist 5/5 moves Maul by %.3f, want 1.1", got)
	}
	lac := talents(map[string]int{"lacerate": 1}, "")
	bleeding := mean(bearOpts{armor: 0, talents: lac}, true)
	if got := mean(bearOpts{armor: 0, talents: talents(map[string]int{"lacerate": 1, "rendAndTear": 5}, "")}, true) / bleeding; !near(got, 1.1, 0.02) {
		t.Errorf("Rend and Tear 5/5 moves Maul on a bleeding target by %.3f, want 1.1", got)
	}
	_, instinct := newBear(t, bearOpts{talents: talents(map[string]int{"naturalInstinct": 2, "lacerate": 1, "primalBite": 1}, "")})
	for _, spell := range []*druid.DruidSpell{instinct.Maul, instinct.SwipeBear, instinct.PrimalBite, instinct.Lacerate} {
		if !near(spell.CritDamageBonus-1, 0.2, 1e-9) {
			t.Errorf("Natural Instinct 2/2 gives %s %.2f more crit damage bonus, want 0.2", spell.ActionID, spell.CritDamageBonus)
		}
	}
}

// TestBearRageTalents: Blood Frenzy 2/2 gives 5 Rage on every crit of a bear's attack (16959), and
// Natural Reaction 5/5 5 Rage on every dodge (417053); Ferocity and Shredding Attacks are in the cost
// tests above.
func TestBearRageTalents(t *testing.T) {
	crit := stats.Stats{stats.MeleeCrit: 100 * core.CritRatingPerCritChance}
	sim, bear := newBear(t, bearOpts{armor: 0, bonus: crit, talents: talents(map[string]int{"primalBite": 1, "bloodFrenzy": 2}, "")})
	hits := record(bear, bear.PrimalBite.ActionID)
	spent := castN(t, sim, bear, bear.PrimalBite, 50)
	crits := 0
	for i, h := range hits() {
		if h.outcome == core.OutcomeCrit {
			crits++
			if !near(spent[i], 15, 1e-9) {
				t.Fatalf("a Primal Bite crit with Blood Frenzy 2/2 cost %.0f Rage net, want 20 less 5", spent[i])
			}
		}
	}
	if crits == 0 {
		t.Fatal("no Primal Bite crit at 100% crit")
	}

}

// TestNaturalReaction: 5 Rage on 20% of dodges a point (417051, 417053), in a whole fight on the
// bench's bear (Natural Reaction 5/5): the Rage it books is 5 a dodge of the boss's swings.
func TestNaturalReaction(t *testing.T) {
	res, iterations := runBench(t, func(rsr *proto.RaidSimRequest) {})
	m := res.RaidMetrics.Parties[0].Players[0]
	gained := 0.0
	for _, r := range m.Resources {
		if r.Id.GetSpellId() == 417053 {
			gained += r.Gain
		}
	}
	dodges := 0.0
	for _, a := range res.EncounterMetrics.Targets[0].Actions {
		for _, tgt := range a.Targets {
			dodges += float64(tgt.Dodges)
		}
	}
	if dodges == 0 || !near(gained, 5*dodges, 1e-6*dodges) {
		t.Errorf("Natural Reaction 5/5 gave %.1f Rage a fight for %.1f dodges, want 5 each", gained/iterations, dodges/iterations)
	}
}

// TestUrsineTwins: Idol of the Ursine Twins resets Primal Bite's cooldown on 10% of the Lacerates
// that land (1306483).
func TestUrsineTwins(t *testing.T) {
	sim, bear := newBear(t, bearOpts{idol: druid.IdolOfTheUrsineTwins, talents: talents(map[string]int{"lacerate": 1, "primalBite": 1}, "")})
	hits := record(bear, bear.Lacerate.ActionID)
	resets := 0
	for i := 0; i < 2000; i++ {
		bear.PrimalBite.CD.Use(sim)
		n := len(hits())
		castN(t, sim, bear, bear.Lacerate, 1)
		if hits()[n].outcome.Matches(core.OutcomeLanded) && bear.PrimalBite.CD.IsReady(sim) {
			resets++
		}
	}
	landed := 0
	for _, h := range hits() {
		if !h.tick && h.outcome.Matches(core.OutcomeLanded) {
			landed++
		}
	}
	if got := float64(resets) / float64(landed); !near(got, 0.1, 0.02) {
		t.Errorf("Idol of the Ursine Twins reset Primal Bite after %.1f%% of the landed Lacerates, want 10%%", 100*got)
	}
}

// TestArmorTrinket: the synthetic armor trinket (990815) gives 1,665 armor for 20 sec on a 2 min
// cooldown.
func TestArmorTrinket(t *testing.T) {
	sim, bear := newBear(t, bearOpts{trinket: druid.FeralTankSyntheticArmorTrinket})
	spell := bear.GetSpell(core.ActionID{ItemID: druid.FeralTankSyntheticArmorTrinket})
	armor := bear.GetStat(stats.Armor)
	if !spell.Cast(sim, bear.CurrentTarget) {
		t.Fatal("the armor trinket didn't cast")
	}
	if got := bear.GetStat(stats.Armor) - armor; got != 1665 {
		t.Errorf("the armor trinket gives %.0f armor, want 1,665", got)
	}
	if spell.CD.Duration != 2*time.Minute {
		t.Errorf("the armor trinket's cooldown is %s, want 2m", spell.CD.Duration)
	}
	advance(sim, 20*time.Second+time.Millisecond)
	if got := bear.GetStat(stats.Armor); got != armor {
		t.Errorf("the armor trinket's armor outlasted its 20 sec (%.0f against %.0f)", got, armor)
	}
}
