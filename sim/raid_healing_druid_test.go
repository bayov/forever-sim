package sim

// The Restoration druid's heals and healing talents (docs/healing_druid.md): each heal heals its
// target, costs its mana, takes its cast time and cooldown and ticks as often as it should, and each
// talent moves exactly what it should. Most checks read one iteration's debug log, which prints every
// heal's base (with its healing power bonus), the caster's multipliers on it and each cast's cost and
// cast time. Needs the item database: go test -tags with_db.

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/druid"
)

// The druid's talent trees, in the order of their talent strings (proto/druid.proto).
var druidTrees = [3][]string{
	{"improvedWrath", "genesis", "moonglow", "improvedMoonfire", "naturesMajesty", "naturesReach", "improvedEntanglingRoots",
		"naturesSplendor", "insectSwarm", "vengeance", "improvedStarfire", "overgrowth", "naturesGrace", "eclipse", "moonfury", "moonkinForm"},
	{"ferocity", "furor", "feralSwiftness", "feralInstinct", "brutalImpact", "thickHide", "shreddingAttacks", "savageFury",
		"feralCharge", "sharpenedClaws", "predatoryStrikes", "lacerate", "predatoryDiscipline", "bloodFrenzy", "leaderOfThePack",
		"naturalInstinct", "rendAndTear", "naturalReaction", "primalBite", "shiftingPower"},
	{"naturesFocus", "heartOfTheWild", "naturalist", "reflection", "naturalShapeshifter", "subtlety", "giftOfNature", "bloom",
		"tranquilSpirit", "improvedRejuvenation", "swiftmend", "naturesSwiftness", "livingSpirit", "improvedTranquility",
		"improvedRegrowth", "wildGrowth"},
}

// druidTalents is a talent string with the points given and none elsewhere.
func druidTalents(points map[string]int) string {
	var trees []string
	for _, tree := range druidTrees {
		s := ""
		for _, name := range tree {
			s += strconv.Itoa(points[name])
		}
		trees = append(trees, strings.TrimRight(s, "0"))
	}
	return strings.Join(trees, "-")
}

// druidRaid is a tank, the bench's Restoration druid (player 1) with these talents and rotation
// lines, and Frost mages, the boss hitting the tank for about 600 every 2 sec.
func druidRaid(t *testing.T, talents map[string]int, mages int, items ...string) *proto.RaidSimRequest {
	rsr := healingRaid(t, 1, 0, mages)
	healer := benchPlayer(t, "../tools/healbench/req/druid_restoration.json")
	healer.TalentsString = druidTalents(talents)
	// A race without haste, so cast times read as the spells'.
	healer.Race = proto.Race_RaceNightElf
	rsr.Raid.Parties[0].Players = append([]*proto.Player{rsr.Raid.Parties[0].Players[0], healer}, rsr.Raid.Parties[0].Players[1:]...)
	for i, p := range rsr.Raid.Parties[0].Players {
		p.Name = strconv.Itoa(i)
	}
	rsr.Encounter.Targets[0].MinBaseDamage = 600
	healerRotation(rsr, 1, items...)
	return rsr
}

func castAt(spellID int32, target string, cond string) string {
	c := ""
	if cond != "" {
		c = `"condition":` + cond + `,`
	}
	tgt := ""
	if target != "" {
		tgt = `,"target":{"type":"` + target + `"}`
	}
	return fmt.Sprintf(`{"action":{%s"castSpell":{"spellId":{"spellId":%d}%s}}}`, c, spellID, tgt)
}

func hotOff(spellID int32, target string) string {
	return fmt.Sprintf(`{"not":{"val":{"dotIsActive":{"spellId":{"spellId":%d},"targetUnit":{"type":"%s"}}}}}`, spellID, target)
}

// healEvent is one heal of the healer in the debug log.
type healEvent struct {
	at                            float64
	target                        string
	healingPower, base, afterMods float64
	tick, crit                    bool
}

// castEvent is one cast of the healer: its cost and cast time.
type castEvent struct {
	at, cost, castTime, effective float64
}

type druidLog struct {
	heals     map[int32][]healEvent
	casts     map[int32][]castEvent
	cooldowns map[int32][]float64 // "on cooldown until", less the time it was cast
	lines     []string
}

var (
	reDebug = regexp.MustCompile(`^\[([\d.]+)\] \[1 \(#2\)\] \[([^\]]+)\] \{SpellID: (\d+)\} \[DEBUG\] HealingPower: ([\d.]+), BaseHealing:([\d.]+), AfterCasterMods:([\d.]+)`)
	reDealt = regexp.MustCompile(`^\[[\d.]+\] \[1 \(#2\)\] \[[^\]]+\] \{SpellID: (\d+)\} (tick )?(Hit|Crit) for [\d.]+ healing`)
	reCast  = regexp.MustCompile(`^\[([\d.]+)\] \[1 \(#2\)\] Casting \{SpellID: (\d+)\} \(Cost = ([\d.]+), Cast Time = ([\d.]+)(m?s), Effective Time = ([\d.]+)(m?s)`)
	reCD    = regexp.MustCompile(`^\[([\d.]+)\] \[1 \(#2\)\] \{SpellID: (\d+)\} on cooldown until ([\d.]+)`)
)

// runDruidLog runs one iteration with the debug log and reads the healer's (named 1 in the log)
// heals and casts from it.
func runDruidLog(t *testing.T, rsr *proto.RaidSimRequest) druidLog {
	t.Helper()
	rsr.SimOptions.Iterations = 1
	rsr.SimOptions.Debug = true
	res := runRaid(t, rsr)
	l := druidLog{heals: map[int32][]healEvent{}, casts: map[int32][]castEvent{}, cooldowns: map[int32][]float64{}}
	l.lines = strings.Split(res.Logs, "\n")
	f := func(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }
	dur := func(s, unit string) float64 { return f(s) / core.TernaryFloat64(unit == "ms", 1000, 1) }
	var pending []healEvent
	var pendingID []int32
	for _, line := range l.lines {
		if m := reDebug.FindStringSubmatch(line); m != nil {
			id, _ := strconv.Atoi(m[3])
			pending = append(pending, healEvent{at: f(m[1]), target: m[2], healingPower: f(m[4]), base: f(m[5]), afterMods: f(m[6])})
			pendingID = append(pendingID, int32(id))
		} else if m := reDealt.FindStringSubmatch(line); m != nil {
			id, _ := strconv.Atoi(m[1])
			for i := range pending {
				if pendingID[i] == int32(id) {
					h := pending[i]
					h.tick, h.crit = m[2] != "", m[3] == "Crit"
					l.heals[int32(id)] = append(l.heals[int32(id)], h)
					pending = append(pending[:i], pending[i+1:]...)
					pendingID = append(pendingID[:i], pendingID[i+1:]...)
					break
				}
			}
		} else if m := reCast.FindStringSubmatch(line); m != nil {
			id, _ := strconv.Atoi(m[2])
			ct := f(m[4])
			if m[5] == "ms" {
				ct /= 1000
			}
			l.casts[int32(id)] = append(l.casts[int32(id)], castEvent{at: f(m[1]), cost: f(m[3]), castTime: ct, effective: dur(m[6], m[7])})
		} else if m := reCD.FindStringSubmatch(line); m != nil {
			id, _ := strconv.Atoi(m[2])
			l.cooldowns[int32(id)] = append(l.cooldowns[int32(id)], f(m[3])-f(m[1]))
		}
	}
	return l
}

func (l druidLog) directs(id int32) (out []healEvent) {
	for _, h := range l.heals[id] {
		if !h.tick {
			out = append(out, h)
		}
	}
	return
}

func (l druidLog) ticks(id int32) (out []healEvent) {
	for _, h := range l.heals[id] {
		if h.tick {
			out = append(out, h)
		}
	}
	return
}

// multiplier is a heal's caster multiplier, the same on every heal of it.
func (l druidLog) multiplier(t *testing.T, events []healEvent) float64 {
	t.Helper()
	if len(events) == 0 {
		t.Fatal("no heals")
	}
	m := events[0].afterMods / events[0].base
	for _, h := range events {
		if math.Abs(h.afterMods/h.base-m) > 0.002 {
			t.Errorf("multipliers %.4f and %.4f on one heal", m, h.afterMods/h.base)
			break
		}
	}
	return m
}

func druidNear(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

const (
	htTop  = 25297 // Healing Touch rank 11
	rgTop  = 9858  // Regrowth rank 9
	rjTop  = 25299 // Rejuvenation rank 11
	trTop  = 9863  // Tranquility rank 4
	wgTop  = 1238215
	smID   = 18562
	nsID   = 17116
	bloom  = 1328622
	innerv = 29166
)

// Healing Touch rank 11: 2332 ± 8.3% plus all of our healing power, 840 mana, a 3.5 sec cast.
func TestDruidHealingTouch(t *testing.T) {
	l := runDruidLog(t, druidRaid(t, nil, 0, castAt(htTop, "Tank", "")))
	heals := l.directs(htTop)
	if len(heals) < 10 {
		t.Fatalf("%d Healing Touches", len(heals))
	}
	low, high := druid.HealingTouchRanks[11].At(60)
	for _, h := range heals {
		if roll := h.base - h.healingPower; roll < low-0.2 || roll > high+0.2 || h.target != "0 (#1)" {
			t.Fatalf("a Healing Touch of %.1f with %.0f healing power at %s, want %.0f to %.0f plus it at the tank", h.base, h.healingPower, h.target, low, high)
		}
	}
	if m := l.multiplier(t, heals); !druidNear(m, 1, 0.001) {
		t.Errorf("multiplier %.4f with no talents", m)
	}
	c := l.casts[htTop][0]
	if c.cost != 840 || c.castTime != 3.5 {
		t.Errorf("cost %.1f, cast %.2f s; want 840 and 3.5", c.cost, c.castTime)
	}
}

// Forever's downranking (sim/core/downrank.go) takes a rank learned more than 17 levels below 60 down 5%
// a level, the direct heal and the HoT alike: Healing Touch rank 6 (learned at 32) keeps 45% of its
// coefficient, Rejuvenation rank 6 (34) 55% of its 20% a tick, Regrowth rank 4 (30) 35%, and rank 8 of
// Healing Touch (44) all of it.
func TestDruidDownrankedHeals(t *testing.T) {
	for _, c := range []struct {
		name    string
		spellID int32
		tick    bool
		base    func(hp float64) (float64, float64)
	}{
		{"Healing Touch rank 6", 6778, false, func(hp float64) (float64, float64) {
			low, high := druid.HealingTouchRanks[6].At(60)
			return low + 0.45*hp, high + 0.45*hp
		}},
		{"Healing Touch rank 8", 9758, false, func(hp float64) (float64, float64) {
			low, high := druid.HealingTouchRanks[8].At(60)
			return low + hp, high + hp
		}},
		{"Rejuvenation rank 6", 3627, true, func(hp float64) (float64, float64) {
			tick := druid.RejuvenationRanks[6].TickAt(60) + 0.55*0.2*hp
			return tick, tick
		}},
		{"Regrowth rank 4", 8940, true, func(hp float64) (float64, float64) {
			tick := druid.RegrowthRanks[4].TickAt(60) + 0.35*0.071*hp
			return tick, tick
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			cond, heals := "", []healEvent(nil)
			if c.tick {
				cond = hotOff(c.spellID, "Tank")
			}
			l := runDruidLog(t, druidRaid(t, nil, 0, castAt(c.spellID, "Tank", cond)))
			if heals = l.directs(c.spellID); c.tick {
				heals = l.ticks(c.spellID)
			}
			if len(heals) == 0 {
				t.Fatalf("no heals")
			}
			for _, h := range heals {
				if low, high := c.base(h.healingPower); h.base < low-0.5 || h.base > high+0.5 {
					t.Fatalf("%.1f with %.0f healing power, want %.1f to %.1f", h.base, h.healingPower, low, high)
				}
			}
		})
	}
}

// Regrowth rank 9: 1021 ± 5.5% plus 28.6% of healing power, then 7 ticks of 142 plus 7.1%, 525 mana, 2
// sec. Recast while its HoT is up, it starts over.
func TestDruidRegrowth(t *testing.T) {
	l := runDruidLog(t, druidRaid(t, nil, 0, castAt(rgTop, "Tank", hotOff(rgTop, "Tank"))))
	directs, ticks := l.directs(rgTop), l.ticks(rgTop)
	low, high := druid.RegrowthRanks[9].At(60)
	for _, h := range directs {
		if roll := h.base - 0.286*h.healingPower; roll < low-0.5 || roll > high+0.5 {
			t.Fatalf("a Regrowth of %.1f with %.0f healing power, want %.0f to %.0f plus 28.6%%", h.base, h.healingPower, low, high)
		}
	}
	for _, h := range ticks {
		if !druidNear(h.base, 142+0.071*h.healingPower, 0.2) {
			t.Fatalf("a Regrowth tick of %.1f with %.0f healing power", h.base, h.healingPower)
		}
	}
	casts := len(l.casts[rgTop])
	// The last cast's HoT may not finish.
	if len(ticks) < 7*(casts-1) || len(ticks) > 7*casts {
		t.Errorf("%d ticks from %d Regrowths, want 7 a cast", len(ticks), casts)
	}
	if c := l.casts[rgTop][0]; c.cost != 525 || c.castTime != 2 {
		t.Errorf("cost %.1f, cast %.2f s; want 525 and 2", c.cost, c.castTime)
	}
}

// Rejuvenation rank 11: 4 ticks of 194 plus 20% of healing power, 360 mana, instant; Nature's Splendor
// makes it 5.
func TestDruidRejuvenation(t *testing.T) {
	for _, splendor := range []int{0, 1} {
		l := runDruidLog(t, druidRaid(t, map[string]int{"naturesSplendor": splendor}, 0, castAt(rjTop, "Tank", hotOff(rjTop, "Tank"))))
		ticks := l.ticks(rjTop)
		for _, h := range ticks {
			if !druidNear(h.base, 194+0.2*h.healingPower, 0.2) {
				t.Fatalf("a Rejuvenation tick of %.1f with %.0f healing power", h.base, h.healingPower)
			}
		}
		casts, want := len(l.casts[rjTop]), 4+splendor
		if len(ticks) < want*(casts-1) || len(ticks) > want*casts || casts < 10 {
			t.Errorf("Nature's Splendor %d: %d ticks from %d Rejuvenations, want %d a cast", splendor, len(ticks), casts, want)
		}
		if c := l.casts[rjTop][0]; c.cost != 360 || c.castTime != 0 {
			t.Errorf("cost %.1f, cast %.2f s; want 360 and instant", c.cost, c.castTime)
		}
		if len(l.directs(rjTop)) != 0 {
			t.Error("Rejuvenation healed at once")
		}
	}
}

// Tranquility rank 4: a 10 sec channel healing every member of our party 285 plus 6.7% of healing
// power every 2 sec, 925 mana, 5 min cooldown; Improved Tranquility 2/2 takes 60% off.
func TestDruidTranquility(t *testing.T) {
	for _, imp := range []int{0, 2} {
		l := runDruidLog(t, druidRaid(t, map[string]int{"improvedTranquility": imp}, 3, castAt(trTop, "", "")))
		casts := l.casts[trTop]
		if len(casts) == 0 || casts[0].cost != 925 {
			t.Fatalf("Tranquility casts %v, want 925 mana", casts)
		}
		if cd, want := l.cooldowns[trTop][0], 300*(1-0.3*float64(imp)); !druidNear(cd, want, 0.01) {
			t.Errorf("Improved Tranquility %d: cooldown %.1f s, want %.0f", imp, cd, want)
		}
		ticks := l.ticks(trTop)
		perTarget := map[string]int{}
		for _, h := range ticks {
			if !druidNear(h.base, 285+0.067*h.healingPower, 0.2) {
				t.Fatalf("a Tranquility tick of %.1f with %.0f healing power", h.base, h.healingPower)
			}
			if h.at < 10.01 {
				perTarget[h.target]++
			}
		}
		if len(perTarget) != 5 {
			t.Errorf("the first channel healed %v, want the party of 5", perTarget)
		}
		for who, n := range perTarget {
			if n != 5 {
				t.Errorf("%s got %d ticks of the first channel, want 5", who, n)
			}
		}
	}
}

// Swiftmend can only be cast at a target with our Rejuvenation or Regrowth, consumes it, and heals its
// whole HoT: every tick it would heal (Nature's Splendor's too) at its snapshot, before the HoT's own
// modifiers. 20% of base mana, 15 sec cooldown.
func TestDruidSwiftmend(t *testing.T) {
	talents := map[string]int{"swiftmend": 1, "naturesSplendor": 1}
	l := runDruidLog(t, druidRaid(t, talents, 0, castAt(smID, "Tank", ""), castAt(rjTop, "Tank", hotOff(rjTop, "Tank"))))
	sm := l.directs(smID)
	if len(sm) < 8 {
		t.Fatalf("%d Swiftmends", len(sm))
	}
	for _, h := range sm {
		if !druidNear(h.base, 5*(194+0.2*h.healingPower), 2) {
			t.Errorf("a Swiftmend of %.1f, want 5 Rejuvenation ticks at %.0f healing power", h.base, h.healingPower)
		}
	}
	// The bench gear's Idol of Swiftness takes 3 sec off; without it, 15.
	if cd := l.cooldowns[smID][0]; cd != 12 {
		t.Errorf("Swiftmend cooldown %.1f with Idol of Swiftness", cd)
	}
	noIdol := druidRaid(t, talents, 0, castAt(smID, "Tank", ""), castAt(rjTop, "Tank", hotOff(rjTop, "Tank")))
	noIdol.Raid.Parties[0].Players[1].Equipment.Items[16] = &proto.ItemSpec{}
	if cd := runDruidLog(t, noIdol).cooldowns[smID][0]; cd != 15 {
		t.Errorf("Swiftmend cooldown %.1f without the idol", cd)
	}
	if c := l.casts[smID][0]; c.cost < 200 || c.cost > 300 {
		t.Errorf("Swiftmend cost %.1f, want 20%% of base mana", c.cost)
	}
	// Each Swiftmend took its Rejuvenation: Rejuvenation, cast whenever it's off, goes out again on the
	// next global cooldown.
	for _, s := range l.casts[smID] {
		recast := false
		for _, r := range l.casts[rjTop] {
			recast = recast || (r.at > s.at && r.at <= s.at+1.6)
		}
		if !recast {
			t.Errorf("no Rejuvenation after the Swiftmend at %.1f: it left the HoT on", s.at)
		}
	}

	// With no HoT on the tank, no Swiftmend.
	l = runDruidLog(t, druidRaid(t, talents, 0, castAt(smID, "Tank", ""), castAt(htTop, "Tank", "")))
	if n := len(l.casts[smID]); n != 0 {
		t.Errorf("%d Swiftmends with no HoT up", n)
	}
}

// Wild Growth rank 3: a 3 sec cast (2.5 with Naturalist 5/5), 6 sec cooldown, 1050 mana, a HoT on the
// target's party of 7 ticks of 97 plus 3.3% of healing power, the first 6% over and the last 6% under.
func TestDruidWildGrowth(t *testing.T) {
	for _, naturalist := range []int{0, 5} {
		l := runDruidLog(t, druidRaid(t, map[string]int{"wildGrowth": 1, "naturalist": naturalist}, 3, castAt(wgTop, "Tank", "")))
		c := l.casts[wgTop]
		if len(c) == 0 || c[0].cost != 1050 || !druidNear(c[0].castTime, 3-0.1*float64(naturalist), 0.001) {
			t.Fatalf("Naturalist %d: Wild Growth casts %v", naturalist, c[:min(1, len(c))])
		}
		// The cooldown starts as the cast ends.
		if cd := l.cooldowns[wgTop][0]; !druidNear(cd, c[0].castTime+6, 0.001) {
			t.Errorf("cooldown %.1f from the cast's start, want the cast and 6 sec", cd)
		}
		first := c[0].at + c[0].castTime
		perTarget := map[string][]healEvent{}
		for _, h := range l.ticks(wgTop) {
			if !druidNear(h.base, 97+0.033*h.healingPower, 0.2) {
				t.Fatalf("a Wild Growth tick of %.1f with %.0f healing power", h.base, h.healingPower)
			}
			if h.at <= first+7.01 {
				perTarget[h.target] = append(perTarget[h.target], h)
			}
		}
		if len(perTarget) != 5 {
			t.Errorf("the first Wild Growth reached %d players, want the party of 5", len(perTarget))
		}
		for who, ticks := range perTarget {
			if len(ticks) != 7 {
				t.Errorf("%s got %d ticks, want 7", who, len(ticks))
				continue
			}
			if r := (ticks[0].afterMods / ticks[0].base) / (ticks[6].afterMods / ticks[6].base); !druidNear(r, 1.06/0.94, 0.003) {
				t.Errorf("the first tick heals %.3f times the last, want %.3f", r, 1.06/0.94)
			}
		}
	}
}

// Bloom heals for our Spirit when a Rejuvenation or Regrowth runs out, and not when it's recast before.
func TestDruidBloom(t *testing.T) {
	expiring := runDruidLog(t, druidRaid(t, map[string]int{"bloom": 1}, 0, castAt(rjTop, "Tank", hotOff(rjTop, "Tank"))))
	blooms := expiring.directs(bloom)
	if casts := len(expiring.casts[rjTop]); len(blooms) < casts-1 || len(blooms) > casts {
		t.Errorf("%d Blooms from %d Rejuvenations left to run out", len(blooms), casts)
	}
	if len(blooms) > 0 && (blooms[0].base < 150 || blooms[0].base > 1000) {
		t.Errorf("Bloom heals %.0f", blooms[0].base)
	}
	// Recast every global cooldown, no Rejuvenation ever runs out.
	recast := runDruidLog(t, druidRaid(t, map[string]int{"bloom": 1}, 0, castAt(rjTop, "Tank", "")))
	if n := len(recast.directs(bloom)); n > 1 {
		t.Errorf("%d Blooms with Rejuvenation recast before it runs out", n)
	}
	none := runDruidLog(t, druidRaid(t, nil, 0, castAt(rjTop, "Tank", hotOff(rjTop, "Tank"))))
	if n := len(none.directs(bloom)); n != 0 {
		t.Errorf("%d Blooms without the talent", n)
	}
	// Living Spirit 3/3: 15% more Spirit, so Bloom heals 15% more.
	spirited := runDruidLog(t, druidRaid(t, map[string]int{"bloom": 1, "livingSpirit": 3}, 0, castAt(rjTop, "Tank", hotOff(rjTop, "Tank"))))
	if a, b := blooms[0].base, spirited.directs(bloom)[0].base; !druidNear(b/a, 1.15, 0.01) {
		t.Errorf("Bloom heals %.0f with Living Spirit 3/3 against %.0f", b, a)
	}
}

// Nature's Swiftness makes the next Healing Touch instant and goes on a 3 min cooldown.
func TestDruidNaturesSwiftness(t *testing.T) {
	l := runDruidLog(t, druidRaid(t, map[string]int{"naturesSwiftness": 1}, 0, castAt(nsID, "", ""), castAt(htTop, "Tank", "")))
	ns, ht := l.casts[nsID], l.casts[htTop]
	if len(ns) == 0 || len(ns) > 2 || (len(ns) == 2 && ns[1].at-ns[0].at < 180) {
		t.Fatalf("Nature's Swiftness casts %v, want one each 3 min", ns)
	}
	ht = ht[:min(len(ht), 10)]
	instants := 0
	for _, c := range ht {
		if c.castTime == 0 {
			instants++
		}
	}
	if instants != 1 || ht[0].castTime != 0 || ht[1].castTime != 3.5 {
		t.Errorf("%d instant Healing Touches, the first two %.1f and %.1f s", instants, ht[0].castTime, ht[1].castTime)
	}
}

// The talents' multipliers on each heal (Gift of Nature, Genesis, Improved Rejuvenation), the cost and
// cast time talents, and what they leave alone.
func TestDruidHealingTalentMultipliers(t *testing.T) {
	rotation := []string{castAt(rjTop, "Tank", hotOff(rjTop, "Tank")), castAt(rgTop, "Tank", hotOff(rgTop, "Tank")), castAt(htTop, "Tank", "")}
	type want struct{ ht, rgDirect, rgTick, rj float64 }
	for _, tc := range []struct {
		talents map[string]int
		want    want
	}{
		{nil, want{1, 1, 1, 1}},
		{map[string]int{"giftOfNature": 5}, want{1.1, 1.1, 1.1, 1.1}},
		{map[string]int{"genesis": 5}, want{1, 1, 1.1, 1.1}},
		{map[string]int{"improvedRejuvenation": 3}, want{1, 1, 1, 1.15}},
		{map[string]int{"giftOfNature": 5, "genesis": 5, "improvedRejuvenation": 3}, want{1.1, 1.1, 1.21, 1.1 * 1.1 * 1.15}},
		// Naturalist's 10% more damage is damage only.
		{map[string]int{"naturalist": 5}, want{1, 1, 1, 1}},
	} {
		l := runDruidLog(t, druidRaid(t, tc.talents, 0, rotation...))
		got := want{l.multiplier(t, l.directs(htTop)), l.multiplier(t, l.directs(rgTop)), l.multiplier(t, l.ticks(rgTop)), l.multiplier(t, l.ticks(rjTop))}
		if !druidNear(got.ht, tc.want.ht, 0.002) || !druidNear(got.rgDirect, tc.want.rgDirect, 0.002) || !druidNear(got.rgTick, tc.want.rgTick, 0.002) || !druidNear(got.rj, tc.want.rj, 0.002) {
			t.Errorf("%v: Healing Touch, Regrowth, its HoT and Rejuvenation take %+v, want %+v", tc.talents, got, tc.want)
		}
	}

	// Tranquil Spirit 5/5: 10% off Healing Touch and Tranquility, not Regrowth. Naturalist 5/5: 0.5 sec off
	// Healing Touch.
	l := runDruidLog(t, druidRaid(t, map[string]int{"tranquilSpirit": 5, "naturalist": 5}, 0, castAt(trTop, "", ""), rotation[1], rotation[2]))
	if c := l.casts[htTop][0]; !druidNear(c.cost, 756, 0.01) || c.castTime != 3 {
		t.Errorf("Healing Touch costs %.1f in %.2f s, want 756 in 3", c.cost, c.castTime)
	}
	if c := l.casts[trTop][0]; !druidNear(c.cost, 832.5, 0.01) {
		t.Errorf("Tranquility costs %.1f, want 832.5", c.cost)
	}
	if c := l.casts[rgTop][0]; c.cost != 525 || c.castTime != 2 {
		t.Errorf("Regrowth costs %.1f in %.2f s", c.cost, c.castTime)
	}
}

// critRate is the crit share of a heal's direct hits over many iterations.
func critRate(t *testing.T, rsr *proto.RaidSimRequest, spellID int32) float64 {
	rsr.SimOptions.Iterations = 40
	var hits, crits int32
	for _, a := range playerMetrics(runRaid(t, rsr), 1).Actions {
		if a.Id.GetSpellId() == spellID {
			for _, tgt := range a.Targets {
				hits += tgt.Hits + tgt.Crits
				crits += tgt.Crits
			}
		}
	}
	if hits < 200 {
		t.Fatalf("%d hits of %d", hits, spellID)
	}
	return float64(crits) / float64(hits)
}

// Improved Regrowth 5/5: 50% more crit on Regrowth, nothing on Healing Touch. Nature's Majesty 2/2: 4% on
// both.
func TestDruidHealingCritTalents(t *testing.T) {
	rot := []string{castAt(rgTop, "Tank", hotOff(rgTop, "Tank")), castAt(htTop, "Tank", "")}
	base := critRate(t, druidRaid(t, nil, 0, rot...), rgTop)
	baseHT := critRate(t, druidRaid(t, nil, 0, rot...), htTop)
	if imp := critRate(t, druidRaid(t, map[string]int{"improvedRegrowth": 5}, 0, rot...), rgTop); !druidNear(imp-base, 0.5, 0.06) {
		t.Errorf("Regrowth crits %.3f with Improved Regrowth 5/5 against %.3f", imp, base)
	}
	if ht := critRate(t, druidRaid(t, map[string]int{"improvedRegrowth": 5}, 0, rot...), htTop); !druidNear(ht, baseHT, 0.04) {
		t.Errorf("Healing Touch crits %.3f with Improved Regrowth against %.3f", ht, baseHT)
	}
	if nm := critRate(t, druidRaid(t, map[string]int{"naturesMajesty": 2}, 0, rot...), htTop); !druidNear(nm-baseHT, 0.04, 0.025) {
		t.Errorf("Healing Touch crits %.3f with Nature's Majesty 2/2 against %.3f", nm, baseHT)
	}
}

// Natural Instinct 2/2 adds 25% of our Intellect to healing power, Heart of the Wild 5/5 10% Intellect.
func TestDruidIntellectTalents(t *testing.T) {
	hp := func(talents map[string]int) float64 {
		return runDruidLog(t, druidRaid(t, talents, 0, castAt(htTop, "Tank", ""))).directs(htTop)[0].healingPower
	}
	base, ni, both := hp(nil), hp(map[string]int{"naturalInstinct": 2}), hp(map[string]int{"naturalInstinct": 2, "heartOfTheWild": 5})
	// The healer's Intellect is about 300 with buffs.
	if gain := ni - base; gain < 50 || gain > 120 {
		t.Errorf("Natural Instinct 2/2 adds %.0f healing power", gain)
	}
	if r := (both - base) / (ni - base); !druidNear(r, 1.1, 0.02) {
		t.Errorf("Heart of the Wild 5/5 moves Natural Instinct's healing power by %.3f, want 1.1", r)
	}
}

// Omen of Clarity procs on heals and makes the next heal free.
func TestDruidOmenOfClarityOnHeals(t *testing.T) {
	l := runDruidLog(t, druidRaid(t, nil, 0, castAt(htTop, "Tank", "")))
	free := 0
	for _, c := range l.casts[htTop] {
		if c.cost == 0 {
			free++
		}
	}
	if free == 0 {
		t.Error("no free Healing Touch from Clearcasting in 180 sec")
	}
}

// Nature's Grace's 10% cast speed and global cooldown cut come from a heal's crit: a Rejuvenation cast
// while it's up takes 1.35 sec. Without the talent none does. (Under Forever the buff comes a spell
// batch after the crit, so the cast that starts with it isn't hasted, and the rotation idles until the
// tank misses a Healing Touch.)
func TestDruidNaturesGraceOnHeals(t *testing.T) {
	hasted := func(talents map[string]int) (n int) {
		graced := `{"auraIsActive":{"auraId":{"spellId":16886}}}`
		hurt := `{"cmp":{"op":"OpGe","lhs":{"healthMissing":{"sourceUnit":{"type":"Tank"}}},"rhs":{"const":{"val":"2000"}}}}`
		l := runDruidLog(t, druidRaid(t, talents, 0, castAt(rjTop, "Tank", graced), castAt(htTop, "Tank", hurt)))
		for _, c := range l.casts[rjTop] {
			if druidNear(c.effective, 1.35, 0.001) {
				n++
			}
		}
		return n
	}
	if n := hasted(map[string]int{"naturesGrace": 1, "naturesMajesty": 2}); n == 0 {
		t.Error("no Rejuvenation's global cooldown cut by Nature's Grace")
	}
	if n := hasted(map[string]int{"naturesMajesty": 2}); n != 0 {
		t.Errorf("%d Rejuvenations hasted without Nature's Grace", n)
	}
}

// Innervate on ourselves: 400% more Spirit regen and all of it while casting, for 20 sec.
func TestDruidInnervateOnSelf(t *testing.T) {
	cond := `{"cmp":{"op":"OpLe","lhs":{"currentManaPercent":{}},"rhs":{"const":{"val":"0.5"}}}}`
	l := runDruidLog(t, druidRaid(t, nil, 0, castAt(innerv, "", cond), castAt(htTop, "Tank", "")))
	if n := len(l.casts[innerv]); n != 1 {
		t.Fatalf("%d Innervates", n)
	}
	gained := 0.0
	re := regexp.MustCompile(`1 \(#2\)\] Gained ([\d.]+) mana from \{OtherID: 2[,}]`)
	at := l.casts[innerv][0].at
	for _, line := range l.lines {
		if m := re.FindStringSubmatch(line); m != nil {
			s, _ := strconv.ParseFloat(line[1:strings.Index(line, "]")], 64)
			if s > at && s <= at+20 {
				v, _ := strconv.ParseFloat(m[1], 64)
				gained += v
			}
		}
	}
	if gained < 1500 {
		t.Errorf("%.0f mana regenerated in Innervate's 20 sec", gained)
	}
}
