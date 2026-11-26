package tankwarrior

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/warrior"
	"google.golang.org/protobuf/encoding/protojson"
)

// What heals us in a fight and what hurts who hits us, in a leveling session (sim/core/leveling_heals.go
// and sim/common/leveling_enchants.go).

// healWarrior is a level 41 Undead Protection warrior with a sword, the Bulwark, these enchants on them
// and these other items.
type healWarrior struct {
	level             int32
	mainHand, offHand int32
	items             []int32
	options           *proto.LevelingOptions
}

func (w healWarrior) request(t *testing.T) *proto.RaidSimRequest {
	t.Helper()
	if w.level == 0 {
		w.level = 41
	}
	if w.options == nil {
		w.options = &proto.LevelingOptions{}
	}
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 991014, Enchant: w.mainHand}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: bulwark, Enchant: w.offHand}
	for i, id := range w.items {
		items[i] = &proto.ItemSpec{Id: id}
	}
	apl := &proto.APLRotation{}
	if err := protojson.Unmarshal([]byte(`{"type":"TypeAPL","priorityList":[
		{"action":{"castSpell":{"spellId":{"spellId":6572}}}},
		{"action":{"castSpell":{"spellId":{"spellId":7386}}}},
		{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"60%"}}}},"levelingRest":{"eat":true}}}
	]}`), apl); err != nil {
		t.Fatal(err)
	}
	player := &proto.Player{
		Race:            proto.Race_RaceUndead,
		Class:           proto.Class_ClassWarrior,
		Level:           w.level,
		Equipment:       &proto.EquipmentSpec{Items: items},
		Rotation:        apl,
		InFrontOfTarget: true,
		Spec:            defensiveOptions,
	}
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	return &proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Leveling: w.options},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1, Iterations: 1},
	}
}

func (w healWarrior) start(t *testing.T) (*core.Simulation, *warrior.Warrior) {
	t.Helper()
	sim := core.NewSim(w.request(t), simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*TankWarrior).GetWarrior()
}

// log plays a whole session and gives what it logged.
func (w healWarrior) log(t *testing.T) []logEvent {
	t.Helper()
	req := w.request(t)
	req.SimOptions.Debug = true
	result := core.RunRaidSim(req)
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	var events []logEvent
	for _, line := range strings.Split(result.Logs, "\n") {
		if m := logLine.FindStringSubmatch(line); m != nil {
			at, _ := strconv.ParseFloat(m[1], 64)
			events = append(events, logEvent{at, m[2]})
		}
	}
	return events
}

type logEvent struct {
	at   float64
	text string
}

var logLine = regexp.MustCompile(`^\[(-?[0-9.]+)\] (.*)`)

// fight is the session with a mob on us.
func (w healWarrior) fight(t *testing.T) (*core.Simulation, *warrior.Warrior) {
	t.Helper()
	sim, war := w.start(t)
	for sim.CurrentTime < 5*time.Minute && !sim.Step() {
		if sim.Encounter.LevelingInCombat(sim) && sim.Encounter.LevelingDistance(sim, war.CurrentTarget) <= 5 {
			return sim, war
		}
	}
	t.Fatal("no fight")
	return nil, nil
}

// recoveryEnchant is Enchant Weapon - Recovery.
const recoveryEnchant = 8721

// deal has the spell land on the target with this outcome, for 1 damage if it hits.
func deal(sim *core.Simulation, spell *core.Spell, target *core.Unit, outcome core.HitOutcome) {
	result := spell.NewResult(target)
	result.Outcome = outcome
	if outcome.Matches(core.OutcomeLanded) {
		result.Damage = 1
	}
	spell.DealDamage(sim, result)
}

// TestLevelingRecovery: Enchant Weapon - Recovery heals 5% of max health when our own attack is dodged
// or parried, swing or ability, at most once in 10 sec, and nothing else does it.
func TestLevelingRecovery(t *testing.T) {
	if _, war := (healWarrior{}).start(t); war.HasAura("Recovery") {
		t.Error("Recovery is up without the enchant")
	}
	sim, war := healWarrior{mainHand: 8721}.fight(t)
	mob := war.CurrentTarget
	heal := war.GetSpell(core.ActionID{SpellID: 1248759})
	casts := func() int32 { return heal.SpellMetrics()[war.UnitIndex].Casts }
	ours, revenge, theirs := war.AutoAttacks.MHAuto(), war.Revenge.Spell, mob.AutoAttacks.MHAuto()

	deal(sim, ours, mob, core.OutcomeHit)
	deal(sim, ours, mob, core.OutcomeMiss)
	for _, outcome := range []core.HitOutcome{core.OutcomeDodge, core.OutcomeParry, core.OutcomeBlock, core.OutcomeHit} {
		deal(sim, theirs, &war.Unit, outcome)
	}
	if casts() != 0 {
		t.Fatalf("%d heals without a dodge or parry of ours", casts())
	}
	war.RemoveHealth(sim, 0.5*war.MaxHealth())
	health := war.CurrentHealth()
	deal(sim, ours, mob, core.OutcomeDodge)
	if got, want := war.CurrentHealth()-health, 0.05*war.MaxHealth(); math.Abs(got-want) > 0.01 && math.Abs(got-1.5*want) > 0.01 {
		t.Errorf("Recovery healed %.2f, want %.2f or a crit", got, want)
	}
	deal(sim, revenge, mob, core.OutcomeParry)
	advance(sim, 9900*time.Millisecond)
	before := casts()
	deal(sim, ours, mob, core.OutcomeParry)
	if before != 1 || casts() != 1 {
		t.Errorf("%d heals within 10 sec, want 1", casts())
	}
	advance(sim, 200*time.Millisecond)
	deal(sim, revenge, mob, core.OutcomeParry)
	if casts() != 2 {
		t.Errorf("%d heals, the parry after 10 sec gave none", casts())
	}
}

// TestLevelingRecoveryFromLevel: the recoveryFromLevel option puts Recovery on the main hand from
// that level on, and never over an enchant the weapon already has.
func TestLevelingRecoveryFromLevel(t *testing.T) {
	const crusader = 1900
	from := func(level int32) *proto.LevelingOptions {
		return &proto.LevelingOptions{RecoveryFromLevel: level}
	}
	for _, c := range []struct {
		name    string
		w       healWarrior
		enchant int32
	}{
		{"never", healWarrior{}, 0},
		{"from 41 at 40", healWarrior{level: 40, options: from(41)}, 0},
		{"from 41 at 41", healWarrior{options: from(41)}, recoveryEnchant},
		{"from 20 at 41", healWarrior{options: from(20)}, recoveryEnchant},
		{"from 41 on Crusader", healWarrior{mainHand: crusader, options: from(41)}, crusader},
	} {
		_, war := c.w.start(t)
		if got := war.GetMHWeapon().Enchant.EffectID; got != c.enchant {
			t.Errorf("%s: main hand enchant %d, want %d", c.name, got, c.enchant)
		}
		if on, want := war.HasAura("Recovery"), c.enchant == recoveryEnchant; on != want {
			t.Errorf("%s: Recovery is on: %v, want %v", c.name, on, want)
		}
	}
}

// TestLevelingShieldSpikes: each melee attack we block hurts who hit us, with no cooldown, through
// armor, and nothing else does.
func TestLevelingShieldSpikes(t *testing.T) {
	for _, c := range []struct {
		enchant, spellID int32
		low, high        float64
	}{{43, 9784, 8, 12}, {463, 9782, 16, 20}, {1704, 16624, 20, 30}} {
		sim, war := healWarrior{offHand: c.enchant}.fight(t)
		mob := war.CurrentTarget
		spike := war.GetSpell(core.ActionID{SpellID: c.spellID})
		metrics := &spike.SpellMetrics()[mob.UnitIndex]
		swing := mob.AutoAttacks.MHAuto()
		for _, outcome := range []core.HitOutcome{core.OutcomeDodge, core.OutcomeParry, core.OutcomeMiss, core.OutcomeHit, core.OutcomeCrit} {
			deal(sim, swing, &war.Unit, outcome)
		}
		deal(sim, war.AutoAttacks.MHAuto(), mob, core.OutcomeBlock)
		if metrics.Casts != 0 {
			t.Errorf("enchant %d: %d hits without a block", c.enchant, metrics.Casts)
		}
		for range 50 {
			deal(sim, swing, &war.Unit, core.OutcomeBlock)
		}
		// Defensive Stance's 10% less damage.
		low, high := 0.9*c.low, 0.9*c.high
		if mean := metrics.TotalDamage / 50; metrics.Casts != 50 || metrics.Crits != 0 || mean < low || mean > high {
			t.Errorf("enchant %d: %d hits (%d crits) for %.1f on average after 50 blocks, want %.1f to %.1f",
				c.enchant, metrics.Casts, metrics.Crits, mean, low, high)
		}
	}
	if _, war := (healWarrior{}).start(t); war.GetSpell(core.ActionID{SpellID: 9784}) != nil {
		t.Error("a shield without a spike has one")
	}
}

// TestLevelingKrolDok: a block heals 13.5 to 22.5, at most once a second, with 4 pieces of Krol'dok
// Battlegear or with the option from its level on.
func TestLevelingKrolDok(t *testing.T) {
	const helm, shoulders, chest, legs = 275006, 287014, 275005, 287013
	pieces := make([]int32, 9)
	pieces[proto.ItemSlot_ItemSlotHead], pieces[proto.ItemSlot_ItemSlotShoulder] = helm, shoulders
	pieces[proto.ItemSlot_ItemSlotChest], pieces[proto.ItemSlot_ItemSlotLegs] = chest, legs
	three := append([]int32{}, pieces...)
	three[proto.ItemSlot_ItemSlotLegs] = 0
	from := func(level int32) *proto.LevelingOptions {
		return &proto.LevelingOptions{KrolDokSetFromLevel: level}
	}
	for _, c := range []struct {
		name string
		w    healWarrior
		on   bool
	}{
		{"nothing", healWarrior{}, false},
		{"3 pieces", healWarrior{items: three}, false},
		{"4 pieces", healWarrior{items: pieces}, true},
		{"from 41 at 40", healWarrior{level: 40, options: from(41)}, false},
		{"from 41 at 41", healWarrior{options: from(41)}, true},
		{"4 pieces and from 41", healWarrior{items: pieces, options: from(41)}, true},
	} {
		sim, war := c.w.start(t)
		if on := war.HasAura("Krol'dok Resolve"); on != c.on {
			t.Errorf("%s: the bonus is on: %v", c.name, on)
		}
		if !c.on {
			continue
		}
		mob := war.CurrentTarget
		heal := war.GetSpell(core.ActionID{SpellID: 1324296})
		metrics := &heal.SpellMetrics()[war.UnitIndex]
		swing := mob.AutoAttacks.MHAuto()
		for _, outcome := range []core.HitOutcome{core.OutcomeDodge, core.OutcomeParry, core.OutcomeHit} {
			deal(sim, swing, &war.Unit, outcome)
		}
		if metrics.Casts != 0 {
			t.Errorf("%s: %d heals without a block", c.name, metrics.Casts)
		}
		war.RemoveHealth(sim, 0.5*war.MaxHealth())
		health := war.CurrentHealth()
		deal(sim, swing, &war.Unit, core.OutcomeBlock)
		if got := war.CurrentHealth() - health + 1; got < 12.5 || got > 1.5*22.5 {
			t.Errorf("%s: a block healed %.1f", c.name, got)
		}
		deal(sim, swing, &war.Unit, core.OutcomeBlock)
		if metrics.Casts != 1 {
			t.Errorf("%s: %d heals of 2 blocks at once", c.name, metrics.Casts)
		}
		advance(sim, time.Second)
		deal(sim, swing, &war.Unit, core.OutcomeBlock)
		if metrics.Casts != 2 {
			t.Errorf("%s: %d heals, the block a second later gave none", c.name, metrics.Casts)
		}
	}
}

// TestLevelingThrillOfAdventure: each kill heals 0.2% of max health a point every 2 sec for 10 sec, and
// the next kill starts that over. Off by default, and a mob that gives no experience gives none.
func TestLevelingThrillOfAdventure(t *testing.T) {
	options := func(points, offset int32) healWarrior {
		return healWarrior{options: &proto.LevelingOptions{ThrillOfAdventure: points, MobLevelOffset: offset}}
	}
	tick := regexp.MustCompile(`Gained ([0-9.]+) health from {SpellID: 1261483}`)
	for _, c := range []struct {
		points, offset int32
		share          float64
	}{{0, 0, 0}, {1, -4, 0.002}, {5, -4, 0.01}, {9, -4, 0.01}, {5, -20, 0}} {
		w := options(c.points, c.offset)
		_, war := w.start(t)
		kill, ticks, kills, last := -100.0, 0, 0, -100.0
		for _, e := range w.log(t) {
			if strings.HasSuffix(e.text, "Died") {
				kill = e.at
				kills++
			}
			m := tick.FindStringSubmatch(e.text)
			if m == nil {
				continue
			}
			ticks++
			amount, _ := strconv.ParseFloat(m[1], 64)
			if want := c.share * war.MaxHealth(); math.Abs(amount-want) > 0.01 {
				t.Fatalf("%d points: a tick of %.3f, want %.3f", c.points, amount, want)
			}
			since := e.at - kill
			if since > 10.01 || since < 1.99 || math.Abs(since-2*math.Round(since/2)) > 0.011 || e.at < last+1.99 {
				t.Fatalf("%d points: a tick %.2f sec after the last kill and %.2f after the last tick", c.points, since, e.at-last)
			}
			last = e.at
		}
		if kills < 10 || (ticks > 0) != (c.share > 0) || (c.share > 0 && ticks < 3*kills) {
			t.Errorf("%d points against mobs %d levels off: %d ticks of %d kills", c.points, c.offset, ticks, kills)
		}
	}
}

// TestLevelingHealthRegenPer5: the option heals its amount every 5 sec, in combat too, and nothing while
// we're dead. Spirit's regen stays what it was.
func TestLevelingHealthRegenPer5(t *testing.T) {
	gained := regexp.MustCompile(`Gained ([0-9.]+) health from {SpellID: 21589}`)
	for _, c := range []struct{ perFive, tick float64 }{{0, 0}, {10, 2}} {
		w := healWarrior{options: &proto.LevelingOptions{HealthRegenPer5: c.perFive}}
		fighting, dead := false, false
		inCombat, outside := 0, 0
		for _, e := range w.log(t) {
			switch {
			case strings.Contains(e.text, "Leveling: now "):
				fighting, dead = strings.HasSuffix(e.text, "fighting"), strings.HasSuffix(e.text, "dead")
			case gained.MatchString(e.text):
				amount, _ := strconv.ParseFloat(gained.FindStringSubmatch(e.text)[1], 64)
				if amount > c.tick+0.001 || dead {
					t.Fatalf("a second healed %.3f (dead: %v), want %.3f", amount, dead, c.tick)
				}
				if fighting {
					inCombat++
				} else {
					outside++
				}
			}
		}
		if on := c.perFive > 0; (inCombat > 0) != on || (outside > 0) != on {
			t.Errorf("%.0f per 5 sec: %d heals in combat and %d out of it", c.perFive, inCombat, outside)
		}
	}
}
