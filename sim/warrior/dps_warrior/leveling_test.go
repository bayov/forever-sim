package dpswarrior

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/warrior"
	"google.golang.org/protobuf/encoding/protojson"
)

// The warrior alone in a leveling session (docs/leveling_warrior.md).

type levelingWarrior struct {
	level    int32
	race     proto.Race
	talents  string
	stance   proto.WarriorStance
	mobs     proto.MobType
	packs    int32
	offset   int32  // the mobs' level less ours
	rotation string // the priority list's actions, as JSON
}

// request is an Undead warrior (unless it names another race) with the Whirlwind Axe and nothing
// else.
func (w levelingWarrior) request(t *testing.T) *proto.RaidSimRequest {
	t.Helper()
	if w.race == proto.Race_RaceUnknown {
		w.race = proto.Race_RaceUndead
	}
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 6975}
	apl := &proto.APLRotation{}
	if err := protojson.Unmarshal([]byte(`{"type":"TypeAPL","priorityList":[`+w.rotation+`]}`), apl); err != nil {
		t.Fatal(err)
	}
	player := &proto.Player{
		Race:            w.race,
		Class:           proto.Class_ClassWarrior,
		Level:           w.level,
		Equipment:       &proto.EquipmentSpec{Items: items},
		TalentsString:   w.talents,
		Rotation:        apl,
		InFrontOfTarget: true,
		Spec:            &proto.Player_Warrior{Warrior: &proto.Warrior{Options: &proto.Warrior_Options{Stance: w.stance}}},
	}
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	return &proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Leveling: &proto.LevelingOptions{MobType: w.mobs, MaxPackSize: w.packs, MobLevelOffset: w.offset}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1, Iterations: 1, Debug: true},
	}
}

// start is the session at its first moment, with nothing pulled yet.
func (w levelingWarrior) start(t *testing.T) (*core.Simulation, *warrior.Warrior) {
	t.Helper()
	req := w.request(t)
	req.SimOptions.Debug = false
	sim := core.NewSim(req, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*DpsWarrior).GetWarrior()
}

// run plays the whole session and returns its log.
func (w levelingWarrior) run(t *testing.T) ([]logEvent, *proto.RaidSimResult) {
	t.Helper()
	result := core.RunRaidSim(w.request(t))
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
	return events, result
}

type logEvent struct {
	at   float64
	text string
}

var logLine = regexp.MustCompile(`^\[(-?[0-9.]+)\] (.*)`)

const cast = `{"action":{"castSpell":{"spellId":{"spellId":%d}}}}`

const (
	castTag     = `{"action":{"castSpell":{"spellId":{"spellId":%d,"tag":1}}}}` // queue Heroic Strike or Cleave
	castAtPull  = `{"action":{"castSpell":{"spellId":{"spellId":%d},"target":{"type":"PullTarget"}}}}`
	walk        = `{"action":{"levelingWalkToPullTarget":{}}}`
	rest        = `{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"60%"}}}},"levelingRest":{"eat":true}}}`
	autocast    = `{"action":{"autocastOtherCooldowns":{}}}`
	lookForMore = `{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"levelingMobsThisPull":{}},"rhs":{"const":{"val":"%d"}}}},"levelingLookForMob":{"stand":true}}}`
)

func casts(spellIDs ...int) string {
	var actions []string
	for _, id := range spellIDs {
		actions = append(actions, fmt.Sprintf(cast, id))
	}
	return strings.Join(actions, ",")
}

// count is how many log lines match the pattern.
func count(events []logEvent, pattern string) int {
	re := regexp.MustCompile(pattern)
	n := 0
	for _, e := range events {
		if re.MatchString(e.text) {
			n++
		}
	}
	return n
}

// advance runs the sim's clock forward by d.
func advance(sim *core.Simulation, d time.Duration) {
	end := sim.CurrentTime + d
	sim.AddPendingAction(&core.PendingAction{NextActionAt: end, OnAction: func(*core.Simulation) {}})
	for sim.CurrentTime < end && !sim.Step() {
	}
}

// toCombat runs the session until we're fighting a mob that has reached us.
func toCombat(t *testing.T, sim *core.Simulation, war *warrior.Warrior) {
	t.Helper()
	for sim.CurrentTime < 5*time.Minute && !sim.Step() {
		if sim.Encounter.LevelingInCombat(sim) && sim.Encounter.LevelingDistance(sim, war.CurrentTarget) <= 5 {
			return
		}
	}
	t.Fatal("no fight")
}

func newLevelWarrior(t *testing.T, level int32, talents string) *warrior.Warrior {
	t.Helper()
	_, war := levelingWarrior{level: level, talents: talents}.start(t)
	return war
}

// TestWarriorRanksByLevel checks which rank of each ability the warrior knows just before and at each
// rank's level (client 71091, SpellLevels), and that it knows none before the first.
func TestWarriorRanksByLevel(t *testing.T) {
	spell := func(ws *warrior.WarriorSpell) int32 {
		if ws == nil {
			return 0
		}
		return ws.SpellID
	}
	for _, c := range []struct {
		name   string
		levels []int32
		ids    []int32
		get    func(war *warrior.Warrior) *warrior.WarriorSpell
	}{
		{"Battle Shout", []int32{1, 12, 22, 32, 42, 52}, []int32{6673, 5242, 6192, 11549, 11550, 11551}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.BattleShout }},
		{"Charge", []int32{4, 26, 46}, []int32{100, 6178, 11578}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.Charge }},
		{"Rend", []int32{4, 10, 20, 30, 40, 50, 60}, []int32{772, 6546, 6547, 6548, 11572, 11573, 11574}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.Rend }},
		{"Thunder Clap", []int32{6, 18, 28, 38, 48, 58}, []int32{6343, 8198, 8204, 8205, 11580, 11581}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.ThunderClap }},
		{"Hamstring", []int32{8, 32, 54}, []int32{1715, 7372, 7373}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.Hamstring }},
		{"Bloodrage", []int32{10}, []int32{2687}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.Bloodrage }},
		{"Sunder Armor", []int32{10, 22, 34, 46, 58}, []int32{7386, 7405, 8380, 11596, 11597}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.SunderArmor }},
		{"Defensive Stance", []int32{10}, []int32{71}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.DefensiveStance }},
		{"Overpower", []int32{12, 28, 44, 60}, []int32{7384, 7887, 11584, 11585}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.Overpower }},
		{"Demoralizing Shout", []int32{14, 24, 34, 44, 54}, []int32{1160, 6190, 11554, 11555, 11556}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.DemoralizingShout }},
		{"Revenge", []int32{14, 24, 34, 44, 54}, []int32{6572, 6574, 7379, 11600, 11601}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.Revenge }},
		{"Cleave", []int32{20, 30, 40, 50, 60}, []int32{845, 7369, 11608, 11609, 20569}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.Cleave }},
		{"Retaliation", []int32{20}, []int32{20230}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.Retaliation }},
		{"Victory Rush", []int32{20}, []int32{402927}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.VictoryRush }},
		{"Execute", []int32{24, 32, 40, 48, 56}, []int32{5308, 20658, 20660, 20661, 20662}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.Execute }},
		{"Berserker Rage", []int32{30}, []int32{18499}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.BerserkerRage }},
		{"Berserker Stance", []int32{30}, []int32{2458}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.BerserkerStance }},
		{"Intercept", []int32{30, 42, 52}, []int32{20252, 20616, 20617}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.Intercept }},
		{"Whirlwind", []int32{36}, []int32{1680}, func(w *warrior.Warrior) *warrior.WarriorSpell { return w.Whirlwind }},
	} {
		for i, level := range c.levels {
			if level > 1 {
				before := int32(0)
				if i > 0 {
					before = c.ids[i-1]
				}
				if got := spell(c.get(newLevelWarrior(t, level-1, ""))); got != before {
					t.Errorf("%s at level %d is %d, want %d", c.name, level-1, got, before)
				}
			}
			if got := spell(c.get(newLevelWarrior(t, level, ""))); got != c.ids[i] {
				t.Errorf("%s at level %d is %d, want %d", c.name, level, got, c.ids[i])
			}
		}
	}
	// Heroic Strike's ranks go by the queue, which every level has.
	if hs := newLevelWarrior(t, 1, "").HeroicStrike; hs.SpellID != 78 {
		t.Errorf("Heroic Strike at level 1 is %d, want 78", hs.SpellID)
	}
}

// TestEveryLevelRuns plays a session at each level from 1 to 60 with every ability in the rotation.
func TestEveryLevelRuns(t *testing.T) {
	rotation := casts(100, 2457, 402927, 20230, 20252, 12809, 5308, 7384, 6572, 1680, 845, 6343, 772, 1160, 6673, 7386, 1715, 20577) + "," + fmt.Sprintf(castTag, 845) + "," + fmt.Sprintf(castTag, 78) + "," + autocast + "," + rest
	for level := int32(1); level <= 60; level++ {
		_, result := levelingWarrior{level: level, rotation: rotation, mobs: proto.MobType_MobTypeHumanoid, offset: -4}.run(t)
		if kills := result.EncounterMetrics.Leveling.MobsKilled; kills < 1 {
			t.Errorf("level %d killed %.1f mobs", level, kills)
		}
	}
}

// TestLevelingStances: the stance we ask to start in counts from the level we learn it, and Battle
// Stance stands in before.
func TestLevelingStances(t *testing.T) {
	for _, c := range []struct {
		level  int32
		stance proto.WarriorStance
		want   warrior.Stance
	}{
		{9, proto.WarriorStance_WarriorStanceDefensive, warrior.BattleStance},
		{10, proto.WarriorStance_WarriorStanceDefensive, warrior.DefensiveStance},
		{29, proto.WarriorStance_WarriorStanceBerserker, warrior.BattleStance},
		{30, proto.WarriorStance_WarriorStanceBerserker, warrior.BerserkerStance},
	} {
		_, war := levelingWarrior{level: c.level, stance: c.stance}.start(t)
		if war.Stance != c.want {
			t.Errorf("at level %d asking for stance %v we start in %d, want %d", c.level, c.stance, war.Stance, c.want)
		}
	}
}

var (
	foundRange = regexp.MustCompile(`Leveling: found (Target \d+) .* (\d+) yards away`)
	mobSwung   = regexp.MustCompile(`^\[(Target \d+)\] \[ \(#1\)\] \{OtherID: 7, Tag: 1\} (Hit|Crit|Miss|Dodge|Parry|Block|Crush)`)
	mobPulled  = regexp.MustCompile(`Leveling: pulled (Target \d+) \(level \d+\) from (\d+) yards`)
)

// TestLevelingCharge: Charge opens pulls from 8 to 25 yards, out of combat only, gives its Rage and
// has us swinging within a second. The charged mob doesn't swing or parry while it's stunned, 1 sec.
func TestLevelingCharge(t *testing.T) {
	events, _ := levelingWarrior{level: 20, talents: "0002", rotation: casts(100) + "," + rest, offset: -4}.run(t)
	charges, pulls := 0, 0
	fighting := false
	chargedAt, target, awaitingSwing := -10.0, "", false
	for _, e := range events {
		switch {
		case strings.Contains(e.text, "Leveling: now "):
			fighting = strings.Contains(e.text, "now fighting")
		case strings.Contains(e.text, "Casting {SpellID: 100}"):
			if fighting {
				t.Errorf("at %.2f charged in combat", e.at)
			}
			charges++
			chargedAt, awaitingSwing = e.at, true
		case strings.Contains(e.text, "rage from {SpellID: 100}") && !strings.Contains(e.text, "Gained 15.000"):
			t.Errorf("at %.2f Charge with 2 points of Improved Charge: %s, want 15 Rage", e.at, e.text)
		case strings.Contains(e.text, "Completed cast {OtherID: 7, Tag: 1}") && strings.HasPrefix(e.text, "[ (#1)]") && awaitingSwing:
			awaitingSwing = false
			if e.at > chargedAt+1 {
				t.Errorf("charged at %.2f and only swung at %.2f", chargedAt, e.at)
			}
		case strings.Contains(e.text, "Parry") && strings.HasPrefix(e.text, "[ (#1)]") && e.at < chargedAt+0.99:
			t.Errorf("at %.2f a mob stunned at %.2f parried", e.at, chargedAt)
		}
		if m := mobPulled.FindStringSubmatch(e.text); m != nil {
			pulls++
			target = m[1]
			if from, _ := strconv.Atoi(m[2]); e.at == chargedAt && (from < 8 || from > 25) {
				t.Errorf("at %.2f charged from %d yards", e.at, from)
			}
		}
		if m := mobSwung.FindStringSubmatch(e.text); m != nil && m[1] == target && e.at < chargedAt+0.99 {
			t.Errorf("at %.2f %s swung, stunned at %.2f", e.at, target, chargedAt)
		}
	}
	if charges < 10 || charges < pulls*3/4 {
		t.Errorf("%d charges in %d pulls, want nearly every pull to open with one", charges, pulls)
	}
}

// TestLevelingIntercept: Intercept takes a mob 8 to 25 yards away in Berserker Stance, in combat too,
// costs 10 Rage, and the mob doesn't swing for 3 sec.
func TestLevelingIntercept(t *testing.T) {
	events, _ := levelingWarrior{level: 42, stance: proto.WarriorStance_WarriorStanceBerserker, packs: 3, offset: -4,
		rotation: strings.Join([]string{fmt.Sprintf(castAtPull, 20252), walk, fmt.Sprintf(lookForMore, 3), rest}, ",")}.run(t)
	landed := regexp.MustCompile(`\[(Target \d+)\] \{SpellID: 20616\} (Hit|Crit) for`)
	inCombat, fighting, stunned := 0, false, map[string]float64{}
	for _, e := range events {
		if strings.Contains(e.text, "Leveling: now ") {
			fighting = strings.Contains(e.text, "now fighting")
		}
		if strings.Contains(e.text, "Casting {SpellID: 20616}") {
			if !strings.Contains(e.text, "Cost = 10.0") {
				t.Errorf("Intercept: %s, want 10 Rage", e.text)
			}
			if fighting {
				inCombat++
			}
		}
		if m := landed.FindStringSubmatch(e.text); m != nil {
			stunned[m[1]] = e.at
		}
		if m := mobSwung.FindStringSubmatch(e.text); m != nil {
			if at, ok := stunned[m[1]]; ok && e.at > at && e.at < at+2.99 {
				t.Errorf("at %.2f %s swung, stunned at %.2f", e.at, m[1], at)
			}
		}
	}
	if len(stunned) == 0 || inCombat == 0 {
		t.Errorf("%d mobs intercepted, %d of the casts in combat", len(stunned), inCombat)
	}
	_, war := levelingWarrior{level: 42, talents: "-00000000000000" + "2"}.start(t)
	if got := war.Intercept.CD.Duration; got != 20*time.Second {
		t.Errorf("Intercept with 2 points of Improved Intercept is on a %v cooldown, want 20s", got)
	}
}

// TestLevelingVictoryRush: usable once after a kill for 20 sec, 30 sec cooldown, no Rage, 1 plus 15%
// of attack power in damage, heals 10% of max health.
func TestLevelingVictoryRush(t *testing.T) {
	events, _ := levelingWarrior{level: 30, rotation: casts(402927) + "," + rest, offset: -4}.run(t)
	died, last, uses := -100.0, -100.0, 0
	for _, e := range events {
		if strings.HasSuffix(e.text, "Died") {
			died = e.at
		}
		if strings.Contains(e.text, "Casting {SpellID: 402927}") {
			uses++
			if e.at > died+20 {
				t.Errorf("Victory Rush at %.2f, the last kill was at %.2f", e.at, died)
			}
			if e.at < last+30 {
				t.Errorf("Victory Rush at %.2f and %.2f", last, e.at)
			}
			last = e.at
		}
	}
	if uses < 5 {
		t.Errorf("%d Victory Rushes in 20 min", uses)
	}

	sim, war := levelingWarrior{level: 30}.start(t)
	toCombat(t, sim, war)
	if war.VictoryRush.CanCast(sim, war.CurrentTarget) {
		t.Error("Victory Rush is usable without a kill")
	}
	war.GetAura("Victorious").Activate(sim)
	war.RemoveHealth(sim, 0.5*war.MaxHealth())
	health := war.CurrentHealth()
	if !war.VictoryRush.Cast(sim, war.CurrentTarget) {
		t.Fatal("Victory Rush isn't usable after a kill")
	}
	if got, want := war.CurrentHealth()-health, 0.1*war.MaxHealth(); math.Abs(got-want) > 0.01 {
		t.Errorf("Victory Rush healed %.1f, want %.1f", got, want)
	}
	if war.GetAura("Victorious").IsActive() {
		t.Error("Victory Rush left Victorious up")
	}
	// Level 41 mobs would be gray to a level 60: none of their kills count.
	if sim, war := (levelingWarrior{level: 30}).start(t); war.VictoryRush.CanCast(sim, war.CurrentTarget) {
		t.Error("Victory Rush is usable before the first kill")
	}
}

// TestLevelingRetaliation: for 15 sec each melee hit we take is answered with a weapon hit on who
// dealt it, for 30 of them at most.
func TestLevelingRetaliation(t *testing.T) {
	sim, war := levelingWarrior{level: 30}.start(t)
	toCombat(t, sim, war)
	mob := war.CurrentTarget
	if !war.Retaliation.Cast(sim, mob) {
		t.Fatal("Retaliation isn't usable in Battle Stance")
	}
	aura := war.GetAura("Retaliation")
	strike := war.GetSpell(core.ActionID{SpellID: 20240})
	swing := mob.AutoAttacks.MHAuto()
	for i := range 40 {
		result := swing.NewResult(&war.Unit)
		result.Outcome = core.OutcomeHit
		result.Damage = 1
		swing.DealDamage(sim, result)
		if want := int32(min(i+1, 30)); strike.SpellMetrics()[mob.UnitIndex].Casts != want {
			t.Fatalf("%d hits taken, %d counterattacks, want %d", i+1, strike.SpellMetrics()[mob.UnitIndex].Casts, want)
		}
	}
	if aura.IsActive() {
		t.Error("Retaliation is still up after 30 hits")
	}
	miss := swing.NewResult(&war.Unit)
	miss.Outcome = core.OutcomeDodge
	aura.Activate(sim)
	swing.DealDamage(sim, miss)
	if aura.GetStacks() != 30 {
		t.Error("Retaliation answered a dodge")
	}
}

// TestLevelingConcussionBlow: 10 Rage, off the global cooldown, and the mob doesn't swing for 5 sec.
func TestLevelingConcussionBlow(t *testing.T) {
	sim, war := levelingWarrior{level: 30, talents: "--000000000000001", packs: 1}.start(t)
	toCombat(t, sim, war)
	war.AutoAttacks.CancelAutoSwing(sim)
	mob := war.CurrentTarget
	for tries := 0; !mob.PseudoStats.Stunned; tries++ {
		if tries == 20 {
			t.Fatal("20 Concussion Blows missed")
		}
		war.ConcussionBlow.CD.Reset()
		war.AddRage(sim, 100, war.RageRefundMetrics)
		rage := war.CurrentRage()
		if !war.ConcussionBlow.Cast(sim, mob) {
			t.Fatal("Concussion Blow isn't castable")
		}
		if spent := rage - war.CurrentRage(); mob.PseudoStats.Stunned && spent != 10 {
			t.Errorf("Concussion Blow cost %.1f Rage, want 10", spent)
		}
	}
	if war.ConcussionBlow.DefaultCast.GCD != 0 {
		t.Error("Concussion Blow is on the global cooldown")
	}
	health := war.CurrentHealth()
	advance(sim, 4900*time.Millisecond)
	if war.CurrentHealth() != health || !mob.PseudoStats.Stunned {
		t.Error("the mob swung while it was stunned")
	}
	advance(sim, 5*time.Second)
	if war.CurrentHealth() == health || mob.PseudoStats.Stunned {
		t.Error("the stun lasts more than 5 sec")
	}
}

// TestLevelingGoreDrinker: after Bloodrage the next 3 melee hits heal 1% of max health each.
func TestLevelingGoreDrinker(t *testing.T) {
	sim, war := levelingWarrior{level: 40, talents: "-00000000000002"}.start(t)
	toCombat(t, sim, war)
	war.AutoAttacks.CancelAutoSwing(sim)
	war.RemoveHealth(sim, 0.5*war.MaxHealth())
	war.Bloodrage.Cast(sim, war.CurrentTarget)
	health := war.CurrentHealth()
	auto := war.AutoAttacks.MHAuto()
	for range 5 {
		result := auto.NewResult(war.CurrentTarget)
		result.Outcome = core.OutcomeHit
		result.Damage = 1
		auto.DealDamage(sim, result)
	}
	if got, want := war.CurrentHealth()-health, 0.03*war.MaxHealth(); math.Abs(got-want) > 0.01 {
		t.Errorf("5 hits after Bloodrage healed %.1f, want %.1f", got, want)
	}
	war.GoreDrinkerAura.Activate(sim)
	advance(sim, 9*time.Second)
	if war.GoreDrinkerAura.IsActive() {
		t.Error("Gore Drinker lasts more than 8 sec")
	}
}

// TestLevelingRageDecay: out of combat the Rage bar loses 1.25 a second, 30% less with Anger
// Management, and only after 10 sec with Lingering Rage 5/5. In combat it loses nothing.
func TestLevelingRageDecay(t *testing.T) {
	for _, c := range []struct {
		name    string
		talents string
		want    float64 // lost over 8 sec of search
	}{
		{"no talents", "", 10},
		{"Anger Management", "0000001", 7},
		{"Lingering Rage 5/5", "-005", 0},
		{"Lingering Rage 2/5", "-002", 5},
	} {
		sim, war := levelingWarrior{level: 30, talents: c.talents}.start(t)
		war.AddRage(sim, 50, war.RageRefundMetrics)
		advance(sim, 8*time.Second+100*time.Millisecond)
		if sim.Encounter.LevelingInCombat(sim) {
			t.Fatal("in combat too early")
		}
		got := 50 - war.CurrentRage()
		if math.Abs(got-c.want) > 0.01 {
			t.Errorf("%s: lost %.2f Rage in 8 sec, want %.2f", c.name, got, c.want)
		}
	}

	sim, war := levelingWarrior{level: 30, talents: "0000001"}.start(t)
	toCombat(t, sim, war)
	war.AutoAttacks.CancelAutoSwing(sim)
	rage := war.CurrentRage()
	advance(sim, 4*time.Second)
	if war.CurrentRage() < rage+1 {
		t.Errorf("Rage went from %.2f to %.2f in combat, want Anger Management's", rage, war.CurrentRage())
	}
}

// TestLevelingHealthCosts: Bloodrage costs 20% of base health and Reckless Swings 25%.
func TestLevelingHealthCosts(t *testing.T) {
	sim, war := levelingWarrior{level: 40, talents: "-0000000001"}.start(t)
	base := war.GetBaseStats()[stats.Health]
	for _, c := range []struct {
		name  string
		spell *warrior.WarriorSpell
		share float64
	}{{"Bloodrage", war.Bloodrage, 0.2}, {"Reckless Swings", war.RecklessSwings, 0.25}} {
		war.GCD.Reset()
		before := war.CurrentHealth()
		if !c.spell.Cast(sim, war.CurrentTarget) {
			t.Fatalf("%s isn't castable", c.name)
		}
		if got := before - war.CurrentHealth(); math.Abs(got-c.share*base) > 0.01 {
			t.Errorf("%s cost %.1f health, want %.1f", c.name, got, c.share*base)
		}
	}
}

// TestLevelingExecute: Execute needs this mob at 20% health.
func TestLevelingExecute(t *testing.T) {
	sim, war := levelingWarrior{level: 30}.start(t)
	toCombat(t, sim, war)
	mob := war.CurrentTarget
	war.AddRage(sim, 100, war.RageRefundMetrics)
	war.GCD.Reset()
	mob.RemoveHealth(sim, mob.CurrentHealth()-0.21*mob.MaxHealth())
	if war.Execute.CanCast(sim, mob) {
		t.Error("Execute is usable at 21% health")
	}
	mob.RemoveHealth(sim, 0.02*mob.MaxHealth())
	if !war.Execute.CanCast(sim, mob) {
		t.Error("Execute isn't usable at 19% health")
	}
}

// TestLevelingCannibalize: an Undead heals 35% of max health over 10 sec at a Humanoid's corpse, every
// 2 min at most. Against beasts, or as an Orc, there is none.
func TestLevelingCannibalize(t *testing.T) {
	undead := levelingWarrior{level: 20, rotation: casts(20577), mobs: proto.MobType_MobTypeHumanoid}
	events, _ := undead.run(t)
	uses, last, ticks := 0, -200.0, 0
	for _, e := range events {
		if strings.Contains(e.text, "Casting {SpellID: 20577}") {
			uses++
			if e.at < last+120 {
				t.Errorf("Cannibalize at %.2f and %.2f", last, e.at)
			}
			last = e.at
		}
		if strings.Contains(e.text, "health from {SpellID: 20577}") {
			ticks++
			if e.at > last+10.01 {
				t.Errorf("at %.2f a tick of the Cannibalize from %.2f", e.at, last)
			}
		}
	}
	if uses < 3 || ticks < 2*uses || ticks > 5*uses {
		t.Errorf("%d Cannibalizes with %d ticks", uses, ticks)
	}
	beasts, orc := undead, undead
	beasts.mobs, orc.race = proto.MobType_MobTypeBeast, proto.Race_RaceOrc
	for name, w := range map[string]levelingWarrior{"an Undead against beasts": beasts, "an Orc": orc} {
		if events, _ := w.run(t); count(events, `Casting \{SpellID: 20577\}`) > 0 {
			t.Errorf("%s cannibalized", name)
		}
	}
}

// TestLevelingQueuedSwings: a rotation that names Heroic Strike without its queue tag casts nothing,
// and one with it puts a Heroic Strike in place of a swing.
func TestLevelingQueuedSwings(t *testing.T) {
	w := levelingWarrior{level: 20, rotation: casts(78, 845), offset: -4}
	if events, _ := w.run(t); count(events, `Casting \{SpellID: (285|845)\}`) > 0 {
		t.Error("Heroic Strike or Cleave was cast outright")
	}
	w.rotation = fmt.Sprintf(castTag, 78)
	events, _ := w.run(t)
	last, strikes := -10.0, 0
	for _, e := range events {
		strike := strings.Contains(e.text, "Casting {SpellID: 285}") // rank 3, level 16
		if strike || strings.HasPrefix(e.text, "[ (#1)] Casting {OtherID: 7, Tag: 1}") {
			if strike {
				strikes++
			}
			if e.at < last+1.5 {
				t.Errorf("swings at %.2f and %.2f", last, e.at)
			}
			last = e.at
		}
	}
	if strikes < 10 {
		t.Errorf("%d Heroic Strikes", strikes)
	}
}

// TestLevelingAreaReach: Cleave only hits mobs that have reached us (Whirlwind and Thunder Clap reach 8
// yards, so a mob on its last steps too), and the caps hold.
func TestLevelingAreaReach(t *testing.T) {
	for _, c := range []struct {
		name   string
		spell  int
		id     int
		stance proto.WarriorStance
		most   int
	}{
		{"Whirlwind", 1680, 1680, proto.WarriorStance_WarriorStanceBerserker, 4},
		{"Thunder Clap", 6343, 8205, proto.WarriorStance_WarriorStanceBattle, 4},
		{"Cleave", 845, 11608, proto.WarriorStance_WarriorStanceBattle, 2},
	} {
		spell := casts(c.spell)
		if c.spell == 845 {
			spell = fmt.Sprintf(castTag, c.spell)
		}
		events, _ := levelingWarrior{level: 40, stance: c.stance, packs: 6, offset: -8,
			rotation: strings.Join([]string{spell, walk, fmt.Sprintf(lookForMore, 6), rest}, ",")}.run(t)
		hit := regexp.MustCompile(fmt.Sprintf(`^\[ \(#1\)\] \[(Target \d+)\] \{SpellID: %d(, Tag: \d+)?\} [A-Z]`, c.id))
		reached := regexp.MustCompile(`Leveling: (Target \d+) reached us`)
		running := map[string]bool{}
		at, n, most := -1.0, 0, 0
		for _, e := range events {
			if m := mobPulled.FindStringSubmatch(e.text); m != nil {
				running[m[1]] = true
			}
			if m := reached.FindStringSubmatch(e.text); m != nil {
				running[m[1]] = false
			}
			if m := hit.FindStringSubmatch(e.text); m != nil && !strings.Contains(e.text, "DEBUG") {
				if running[m[1]] && c.spell == 845 {
					t.Errorf("at %.2f %s hit %s still running at us", e.at, c.name, m[1])
				}
				if e.at != at {
					at, n = e.at, 0
				}
				n++
				most = max(most, n)
			}
		}
		if most < 2 || most > c.most {
			t.Errorf("%s hit up to %d mobs, want 2 to %d", c.name, most, c.most)
		}
	}
}
