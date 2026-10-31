package main

// The training search (docs/leveling_plan.md, Training): what each spell rank the plan casts is
// worth, and when we train each one under a gold budget.
//
// A rank's worth at a level is the hours the level takes when we haven't trained it, less the
// hours with it, when it's our highest rank of the spell there. Both sims play the same
// sessions, so most of their noise cancels (Simulation.reseedWithin).
//
// The budget is the gold earned by each level, the trainer's prices and riding at its level.
// Where the gold runs short, we push the ranks back that cost the fewest hours for the gold they
// free. Pushing ranks back changes what the others are worth (two missing ranks cost more than
// the two alone), so we measure the worth again under the new schedule and schedule again, until
// the schedule stays the same.

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// The trainer prices and each rank's level and source, from wowhead's class spell listings.
const classSpellsPath = "assets/db_inputs/forever_wowhead_class_spells.json"

const (
	// Honored with our capital takes 10% off the trainers (Forever's rule seen on the beta, see
	// BayovCore's RecipeSources.DISCOUNT). Our route gets there in the 20s, before most of the
	// gold is spent.
	trainerDiscount = 0.10
	// The gold earned by a level grows with the level to this power. Both the kills a level
	// takes and the coin each kill gives grow with the level.
	goldGrowth = 3.5
	// How many times we measure the worth again under a new schedule, at most.
	trainingRounds = 4
	// Never: a buy level past the last level.
	neverLevel = maxPlanLevel + 1
	// wowhead's source of a rank a class trainer teaches.
	trainerSource = 6
)

// The pocket money we keep after training, in copper, from BayovCore's Training.lua RESERVE.
var pocketMoney = []struct {
	level  int32
	copper float64
}{{55, 15000}, {50, 12500}, {45, 10000}, {40, 7500}, {36, 5000}, {32, 3500}, {28, 2500}, {24, 1800},
	{20, 1200}, {16, 800}, {14, 500}, {11, 300}, {1, 0}}

func reserveAt(level int32) float64 {
	for _, band := range pocketMoney {
		if level >= band.level {
			return band.copper
		}
	}
	return 0
}

type spellRank struct {
	id     int32
	spell  string
	rank   int // 1 for the spell's first rank
	level  int32
	price  float64 // copper, after the discount
	bought bool    // a trainer sells it, so we choose when to buy it
}

type rankTable struct {
	bySpell map[string][]*spellRank // by rank, lowest first
	byID    map[int32]*spellRank
}

func loadRankTable(class proto.Class) (*rankTable, error) {
	data, err := os.ReadFile(classSpellsPath)
	if err != nil {
		return nil, err
	}
	var file struct {
		Classes map[string]map[string]struct {
			Cost   float64
			Level  int32
			Name   string
			Rank   string
			Source []int
		}
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	name := strings.ToUpper(strings.TrimPrefix(class.String(), "Class"))
	spells, ok := file.Classes[name]
	if !ok {
		return nil, fmt.Errorf("%s has no class %s", classSpellsPath, name)
	}
	t := &rankTable{bySpell: map[string][]*spellRank{}, byID: map[int32]*spellRank{}}
	for idText, s := range spells {
		// Talents (level 0) and spells wowhead has no source for (Season of Discovery's) aren't
		// a trainer's.
		if s.Level <= 0 || len(s.Source) == 0 {
			continue
		}
		id, err := strconv.Atoi(idText)
		if err != nil {
			return nil, err
		}
		// Every character of the class starts with the level 1 ranks (or gets them at once, like
		// the shaman's Rockbiter Weapon), so there's no choosing when to buy them.
		r := &spellRank{id: int32(id), spell: s.Name, level: s.Level, bought: s.Level > 1 && slices.Contains(s.Source, trainerSource)}
		if r.bought {
			r.price = math.Round(s.Cost * (1 - trainerDiscount))
		}
		t.bySpell[s.Name] = append(t.bySpell[s.Name], r)
		t.byID[r.id] = r
	}
	for _, ranks := range t.bySpell {
		sort.Slice(ranks, func(i, j int) bool { return ranks[i].level < ranks[j].level })
		for i, r := range ranks {
			// wowhead's own number, since we leave out a first rank it has no source for, like
			// the Healing Wave every shaman starts with.
			r.rank = i + 1
			if n, err := strconv.Atoi(strings.TrimPrefix(spells[strconv.Itoa(int(r.id))].Rank, "Rank ")); err == nil {
				r.rank = n
			}
		}
	}
	return t, nil
}

// A schedule is the level we buy each rank at, for the ranks we don't buy at their own level.
type schedule map[int32]int32

func (s schedule) buyLevel(r *spellRank) int32 {
	if level, ok := s[r.id]; ok {
		return level
	}
	return r.level
}

func (s schedule) clone() schedule {
	out := schedule{}
	for id, level := range s {
		out[id] = level
	}
	return out
}

func (s schedule) untrained(level int32) []int32 {
	var ids []int32
	for id, buy := range s {
		if buy > level {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	return ids
}

// Our highest rank of a spell at a level, nil when we know none.
func (t *rankTable) top(spell string, level int32, s schedule) *spellRank {
	var top *spellRank
	for _, r := range t.bySpell[spell] {
		if r.level <= level && s.buyLevel(r) <= level {
			top = r
		}
	}
	return top
}

// The spells of each imbue, which a range's options cast and the rotation doesn't.
var imbueSpells = map[proto.WeaponImbue]string{
	proto.WeaponImbue_RockbiterWeapon:   "Rockbiter Weapon",
	proto.WeaponImbue_WindfuryWeapon:    "Windfury Weapon",
	proto.WeaponImbue_FlametongueWeapon: "Flametongue Weapon",
	proto.WeaponImbue_FrostbrandWeapon:  "Frostbrand Weapon",
}

type budget struct {
	goldAt40    float64
	ridingLevel int32
	ridingGold  float64
}

func budgetOf(t *proto.LevelingTrainingSettings) budget {
	return budget{goldAt40: t.GetGoldAt_40(), ridingLevel: t.GetRidingLevel(), ridingGold: t.GetRidingGold()}
}

// The copper earned by the time we reach a level.
func (b budget) earned(level int32) float64 {
	return b.goldAt40 * 10000 * math.Pow(float64(level)/40, goldGrowth)
}

// worth is what we measured: the hours each level takes under a schedule, and the hours we'd
// lose at each level without our highest rank of each spell the level casts.
type worth struct {
	hours map[int32]float64
	loss  map[int32]map[string]float64 // by level, then spell
	top   map[int32]map[string]int32   // the rank each loss is for
}

type trainingSearch struct {
	p        *planner
	ranks    *rankTable
	budget   budget
	levels   []int32
	sessions int32
	// The ranks of the spells the plan doesn't cast that we train anyway, at their level.
	always []*spellRank
	// The ranks we chose not to train before a level (LevelingTrainingSettings.holds). Every
	// schedule we sim has them, and the search doesn't move them.
	holds schedule
	// What we measured for each rank at each level, from every round, the latest kept. A rank
	// pushed back is our highest at levels where it wasn't before, and we take its loss there
	// from the nearest level we measured it at, until the next round measures it there.
	rankLoss map[int32]map[int32]float64
}

type trainingJob struct {
	level     int32
	spell     string // "" for the level under the schedule itself
	untrained []int32
}

type trainingJobResult struct {
	hours float64
	// The standard error of the hours, from the spread of the mobs killed.
	err   float64
	casts map[int32]int32
	ok    bool
}

func (ts *trainingSearch) run(ctx context.Context, jobs []trainingJob) []trainingJobResult {
	results := make([]trainingJobResult, len(jobs))
	var next atomic.Int64
	var wg sync.WaitGroup
	for range runtime.NumCPU() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				i := int(next.Add(1) - 1)
				if i >= len(jobs) {
					return
				}
				job := jobs[i]
				lj := levelJob{level: job.level, talents: planTalentsAt(ts.p.plan, job.level, ts.p.shape), untrained: job.untrained, ownUntrained: true}
				res := core.RunRaidSim(ts.p.request(lj, ts.sessions, levelSeed(ts.p.seed, job.level)))
				m := res.EncounterMetrics.GetLeveling()
				if res.Error != nil || m == nil {
					continue
				}
				r := trainingJobResult{hours: hoursAtLevel(job.level, m.XpPerHour), ok: true}
				if m.MobsKilled > 0 {
					r.err = r.hours * m.MobsKilledStdev / math.Sqrt(float64(ts.sessions)) / m.MobsKilled
				}
				if job.spell == "" {
					r.casts = map[int32]int32{}
					for _, action := range res.RaidMetrics.Parties[0].Players[0].Actions {
						for _, target := range action.Targets {
							r.casts[action.Id.GetSpellId()] += target.Casts
						}
					}
				}
				results[i] = r
				ts.p.sims.Add(1)
			}
		}()
	}
	wg.Wait()
	return results
}

// measure sims each level under the schedule, then again without our highest rank of each
// spell the level casts.
func (ts *trainingSearch) measure(ctx context.Context, s schedule) (*worth, error) {
	var jobs []trainingJob
	for _, level := range ts.levels {
		jobs = append(jobs, trainingJob{level: level, untrained: s.untrained(level)})
	}
	base := ts.run(ctx, jobs)
	w := &worth{hours: map[int32]float64{}, loss: map[int32]map[string]float64{}, top: map[int32]map[string]int32{}}
	var capped []trainingJob
	var tops []*spellRank
	for i, level := range ts.levels {
		if !base[i].ok {
			return nil, fmt.Errorf("level %d didn't sim", level)
		}
		w.hours[level] = base[i].hours
		for _, spell := range ts.castSpells(level, base[i].casts) {
			top := ts.ranks.top(spell, level, s)
			if top == nil || !top.bought {
				continue
			}
			capped = append(capped, trainingJob{level: level, spell: spell, untrained: append(s.untrained(level), top.id)})
			tops = append(tops, top)
		}
	}
	results := ts.run(ctx, capped)
	for i, job := range capped {
		if !results[i].ok {
			return nil, fmt.Errorf("level %d without %s didn't sim", job.level, job.spell)
		}
		loss := results[i].hours - w.hours[job.level]
		if w.loss[job.level] == nil {
			w.loss[job.level], w.top[job.level] = map[string]float64{}, map[string]int32{}
		}
		w.loss[job.level][job.spell] = loss
		w.top[job.level][job.spell] = tops[i].id
		if ts.rankLoss[tops[i].id] == nil {
			ts.rankLoss[tops[i].id] = map[int32]float64{}
		}
		ts.rankLoss[tops[i].id][job.level] = loss
	}
	return w, ctx.Err()
}

// The spells a level casts that we train, and its imbue.
func (ts *trainingSearch) castSpells(level int32, casts map[int32]int32) []string {
	seen := map[string]bool{}
	for id, n := range casts {
		if r := ts.ranks.byID[id]; r != nil && n > 0 {
			seen[r.spell] = true
		}
	}
	if r := rangeAt(ts.p.plan, level); r != nil {
		if spell, ok := imbueSpells[r.Imbue]; ok {
			seen[spell] = true
		}
	}
	var spells []string
	for spell := range seen {
		spells = append(spells, spell)
	}
	sort.Strings(spells)
	return spells
}

// The hours we'd lose at a level without a rank when it's our highest there.
func (ts *trainingSearch) lossAt(r *spellRank, level int32) float64 {
	byLevel := ts.rankLoss[r.id]
	if loss, ok := byLevel[level]; ok {
		return max(loss, 0)
	}
	nearest, best := 0.0, int32(math.MaxInt32)
	for l, loss := range byLevel {
		if d := max(l-level, level-l); d < best {
			nearest, best = loss, d
		}
	}
	return max(nearest, 0)
}

// The ranks the plan casts at some level, the ones the budget pays for.
func (ts *trainingSearch) planRanks(w *worth) []*spellRank {
	spells := map[string]bool{}
	for _, bySpell := range w.loss {
		for spell := range bySpell {
			spells[spell] = true
		}
	}
	var ranks []*spellRank
	for spell := range spells {
		for _, r := range ts.ranks.bySpell[spell] {
			if r.bought && r.level <= maxPlanLevel {
				ranks = append(ranks, r)
			}
		}
	}
	sort.Slice(ranks, func(i, j int) bool { return ranks[i].id < ranks[j].id })
	return ranks
}

// How much gold we're short at a level, in copper: what we've paid for ranks and riding by
// then and the pocket money we keep, less what we've earned.
func (ts *trainingSearch) short(ranks []*spellRank, s schedule, level int32) float64 {
	// Early on we may have earned less than the pocket money, and keep what we have.
	spent := min(reserveAt(level), ts.budget.earned(level))
	for _, r := range ranks {
		if s.buyLevel(r) <= level {
			spent += r.price
		}
	}
	for _, r := range ts.always {
		if s.buyLevel(r) <= level {
			spent += r.price
		}
	}
	if ts.budget.ridingLevel > 0 && level >= ts.budget.ridingLevel {
		spent += ts.budget.ridingGold * 10000
	}
	return spent - ts.budget.earned(level)
}

// skipTails never buys the last ranks of a spell when none of them saves time, with every rank
// trained at its level (minutes). Their gold is saved for nothing, or they even cost time, like
// a totem rank whose bigger buff is worth less than its mana. A rank that saves time keeps the
// ranks before it, even when they cost a little more than it saves, because we measured each
// rank only against the one below it: without Earth Shock 5 and 6, Earth Shock 4 is two ranks
// short at 48, which costs far more than 6 against 5.
func skipTails(ranks []*spellRank, minutes map[int32]float64) schedule {
	s := schedule{}
	bySpell := map[string][]*spellRank{}
	for _, r := range ranks {
		bySpell[r.spell] = append(bySpell[r.spell], r)
	}
	for _, spellRanks := range bySpell {
		sort.Slice(spellRanks, func(i, j int) bool { return spellRanks[i].level < spellRanks[j].level })
		keep := 0
		for i, r := range spellRanks {
			if minutes[r.id] > 0 {
				keep = i + 1
			}
		}
		for _, r := range spellRanks[keep:] {
			s[r.id] = neverLevel
		}
	}
	return s
}

// deferRuns trains ranks later that save nothing until then, when we need them only to buy a
// later rank. Their gold then comes when we have more of it, for the same time or less.
//
// A rank the plan never casts while it's our highest waits for the next rank it does cast, and
// we buy them together: Lightning Bolt 2 to 7 save nothing until Maelstrom Weapon casts Lightning
// Bolt 8 at 44. A rank the plan casts waits until the level where what it saved so far at each
// level (loss, by rank and then level, in hours) adds up to the least, at most until the next
// rank. Earth Shock 3 next to the Earth Shock 2 we downrank to waits for Earth Shock 4 at 24, and
// Earth Shock 5 costs a little at 36 and 37 but saves more from 38, so it waits until 38. We
// measured each rank against the one below it, so we know what one rank less costs, but not two:
// Flame Shock 1 in place of Flame Shock 3 at 30 costs far more than 3 against 2 and 2 against 1
// at their own levels. So a cast rank never waits past the next one.
func deferRuns(ranks []*spellRank, loss map[int32]map[int32]float64, cast map[int32]bool, skipped schedule) schedule {
	s := skipped.clone()
	bySpell := map[string][]*spellRank{}
	for _, r := range ranks {
		if _, never := skipped[r.id]; !never {
			bySpell[r.spell] = append(bySpell[r.spell], r)
		}
	}
	for _, spellRanks := range bySpell {
		sort.Slice(spellRanks, func(i, j int) bool { return spellRanks[i].level < spellRanks[j].level })
		for i := 0; i+1 < len(spellRanks); i++ {
			r, next := spellRanks[i], spellRanks[i+1]
			if !cast[r.id] {
				j := i + 1
				for j < len(spellRanks) && !cast[spellRanks[j].id] {
					j++
				}
				if j == len(spellRanks) {
					break
				}
				for _, run := range spellRanks[i:j] {
					s[run.id] = spellRanks[j].level
				}
				i = j - 1
				continue
			}
			best, at, sum := 0.0, r.level, 0.0
			for l := r.level; l < next.level; l++ {
				sum += loss[r.id][l]
				if sum < best {
					best, at = sum, l+1
				}
			}
			if at > r.level {
				s[r.id] = at
			}
		}
	}
	return s
}

// The least tolerance of check: what a spell's ranks bought later or never may cost in all, in
// hours, for the gold they save.
const checkTolerance = 0.25 / 60

// check sims each spell's ranks that the schedule buys later or never against buying them at
// their levels, on the same sessions, and keeps the change only when it costs no time. The worth
// we schedule by is each rank against the one below it, with every other rank trained, and a
// change can leave us more than one rank short, or cost more next to the other spells' changes
// (they share our mana), which it doesn't see.
//
// "No time" is within the noise of the levels the change touches (one standard error), or
// checkTolerance when that's less: a change that saves gold may cost what we can't tell from
// nothing. When a spell's
// change costs more, we try keeping only its ranks the plan doesn't cast while they're our
// highest (`cast`), like Healing Wave 5 to 9 next to Lesser Healing Wave, before we buy them all
// at their levels.
func (ts *trainingSearch) check(ctx context.Context, s schedule, cast map[int32]bool) (schedule, error) {
	out := s.clone()
	bySpell := map[string][]int32{}
	for id := range s {
		if _, held := ts.holds[id]; held {
			continue
		}
		if r := ts.ranks.byID[id]; r != nil {
			bySpell[r.spell] = append(bySpell[r.spell], id)
		}
	}
	candidates := map[string][]int32{}
	for spell, ids := range bySpell {
		candidates[spell] = ids
	}
	for round := 0; round < 2 && len(candidates) > 0; round++ {
		costs, noise, err := ts.checkCosts(ctx, out, candidates)
		if err != nil {
			return nil, err
		}
		next := map[string][]int32{}
		var spells []string
		for spell := range candidates {
			spells = append(spells, spell)
		}
		sort.Strings(spells)
		for _, spell := range spells {
			tolerance := max(checkTolerance, noise[spell])
			if costs[spell] <= tolerance {
				log.Printf("check: %s's changes cost %+.1f min (±%.1f), kept", spell, 60*costs[spell], 60*noise[spell])
				continue
			}
			for _, id := range candidates[spell] {
				delete(out, id)
			}
			var uncast []int32
			for _, id := range candidates[spell] {
				if !cast[id] {
					uncast = append(uncast, id)
				}
			}
			if round == 0 && len(uncast) > 0 && len(uncast) < len(candidates[spell]) {
				for _, id := range uncast {
					out[id] = s[id]
				}
				next[spell] = uncast
				log.Printf("check: %s's changes cost %+.1f min (±%.1f), trying only its ranks the plan doesn't cast", spell, 60*costs[spell], 60*noise[spell])
			} else {
				log.Printf("check: %s's changes cost %+.1f min (±%.1f), bought at their levels instead", spell, 60*costs[spell], 60*noise[spell])
			}
		}
		candidates = next
	}
	return out, nil
}

// checkCosts sims the schedule at each level a spell's changes touch, and again with them gone
// (the spell's ranks bought at their levels), and says what each spell's changes cost in hours,
// and the noise of that.
func (ts *trainingSearch) checkCosts(ctx context.Context, s schedule, changes map[string][]int32) (map[string]float64, map[string]float64, error) {
	var jobs []trainingJob
	type touch struct {
		spell string
		level int32
		job   int
	}
	var touches []touch
	baseJob := map[int32]int{}
	for spell, ids := range changes {
		for _, level := range ts.levels {
			touched := false
			for _, id := range ids {
				if r := ts.ranks.byID[id]; r.level <= level && s.buyLevel(r) > level {
					touched = true
				}
			}
			if !touched {
				continue
			}
			if _, ok := baseJob[level]; !ok {
				baseJob[level] = len(jobs)
				jobs = append(jobs, trainingJob{level: level, untrained: s.untrained(level)})
			}
			var others []int32
			for _, id := range s.untrained(level) {
				if !slices.Contains(ids, id) {
					others = append(others, id)
				}
			}
			touches = append(touches, touch{spell, level, len(jobs)})
			jobs = append(jobs, trainingJob{level: level, spell: spell, untrained: others})
		}
	}
	results := ts.run(ctx, jobs)
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	costs, variance := map[string]float64{}, map[string]float64{}
	for _, t := range touches {
		base, alt := results[baseJob[t.level]], results[t.job]
		if !base.ok || !alt.ok {
			return nil, nil, fmt.Errorf("level %d didn't sim in the check of %s", t.level, t.spell)
		}
		costs[t.spell] += base.hours - alt.hours
		// Two sims of the same sessions: about half the spread of two unrelated ones
		// (Simulation.reseedWithin).
		variance[t.spell] += 0.25 * (base.err*base.err + alt.err*alt.err)
	}
	noise := map[string]float64{}
	for spell, v := range variance {
		noise[spell] = math.Sqrt(v)
	}
	return costs, noise, nil
}

// plan makes the schedule: every rank at its level but the tails we skip, then at each level
// where the gold runs short we push back the highest ranks we've bought by then that cost the
// fewest hours for the gold, one level at a time. It also says how many hours the pushes cost
// by our measures.
func (ts *trainingSearch) plan(ranks []*spellRank, skipped schedule) (schedule, float64) {
	s := skipped.clone()
	if ts.budget.goldAt40 <= 0 {
		return s, 0
	}
	spells := map[string]bool{}
	for _, r := range ranks {
		spells[r.spell] = true
	}
	cost := 0.0
	for level := ts.levels[0]; level <= maxPlanLevel; level++ {
		for {
			short := ts.short(ranks, s, level)
			if short <= 0 {
				break
			}
			// Each spell's highest rank bought by now, and what pushing it to the next level
			// costs: its loss at each level from the one we bought it at.
			var items []knapsackItem
			for spell := range spells {
				top := ts.ranks.top(spell, level, s)
				if top == nil || !top.bought || top.price <= 0 {
					continue
				}
				c := 0.0
				for l := s.buyLevel(top); l <= level; l++ {
					if l >= ts.levels[0] && l < maxPlanLevel {
						c += ts.lossAt(top, l)
					}
				}
				items = append(items, knapsackItem{rank: top, weight: top.price, cost: c})
			}
			if len(items) == 0 {
				log.Printf("training: %.0fg short at level %d with nothing left to push back", short/10000, level)
				break
			}
			for _, item := range cheapestCover(items, short) {
				s[item.rank.id] = level + 1
				cost += item.cost
			}
		}
	}
	// A rank pushed past the last level is one we never buy.
	for id, level := range s {
		if level > maxPlanLevel {
			s[id] = neverLevel
		}
	}
	return s, cost
}

type knapsackItem struct {
	rank         *spellRank
	weight, cost float64
}

// cheapestCover picks the items whose weights add up to at least need for the least cost, or
// all of them when even that's not enough. Weights are copper, and we count them in silver.
func cheapestCover(items []knapsackItem, need float64) []knapsackItem {
	total := 0.0
	for _, item := range items {
		total += item.weight
	}
	if total <= need {
		return items
	}
	n := int(math.Ceil(need / 100))
	// best[w] is the least cost of a set weighing at least w silver, and pick says how we got it.
	best := make([]float64, n+1)
	for w := 1; w <= n; w++ {
		best[w] = math.Inf(1)
	}
	choice := make([][]bool, len(items))
	for i, item := range items {
		choice[i] = make([]bool, n+1)
		w := int(math.Floor(item.weight / 100))
		for cap := n; cap >= 1; cap-- {
			from := max(cap-w, 0)
			if c := best[from] + item.cost; c < best[cap] {
				best[cap] = c
				choice[i][cap] = true
			}
		}
	}
	var picked []knapsackItem
	cap := n
	for i := len(items) - 1; i >= 0 && cap > 0; i-- {
		if choice[i][cap] {
			picked = append(picked, items[i])
			cap = max(cap-int(math.Floor(items[i].weight/100)), 0)
		}
	}
	return picked
}

func sameSchedule(a, b schedule) bool {
	if len(a) != len(b) {
		return false
	}
	for id, level := range a {
		if b[id] != level {
			return false
		}
	}
	return true
}

func totalHours(w *worth) float64 {
	total := 0.0
	for _, h := range w.hours {
		total += h
	}
	return total
}

// searchTraining measures each rank's worth with every rank trained at its level, then makes
// the schedule under the plan's budget.
func searchTraining(ctx context.Context, p *planner, sessions int32) (*proto.LevelingTrainingSettings, error) {
	ranks, err := loadRankTable(p.req.Base.Raid.Parties[0].Players[0].Class)
	if err != nil {
		return nil, err
	}
	settings := p.plan.GetTraining()
	if settings == nil {
		return nil, fmt.Errorf("the plan has no training settings")
	}
	ts := &trainingSearch{p: p, ranks: ranks, budget: budgetOf(settings), levels: planLevels(p.plan), sessions: sessions, rankLoss: map[int32]map[int32]float64{}, holds: schedule{}}
	for _, h := range settings.Holds {
		r := ranks.byID[h.SpellId]
		if r == nil {
			return nil, fmt.Errorf("the class has no rank %d to hold", h.SpellId)
		}
		if h.Level > r.level {
			ts.holds[h.SpellId] = h.Level
		}
	}

	full, err := ts.measure(ctx, ts.holds)
	if err != nil {
		return nil, err
	}
	planRanks := ts.planRanks(full)
	log.Printf("every rank at its level: %.2fh, %d ranks the plan casts, %d held", totalHours(full), len(planRanks), len(ts.holds))

	// A rank's worth: the hours we'd lose over the levels it's our highest rank, when every rank
	// is trained at its level.
	minutes := map[int32]float64{}
	for level, bySpell := range full.top {
		for spell, id := range bySpell {
			minutes[id] += 60 * full.loss[level][spell]
		}
	}
	var worthList []*proto.LevelingRankWorth
	planSpells := map[string]bool{}
	for _, r := range planRanks {
		planSpells[r.spell] = true
		worthList = append(worthList, &proto.LevelingRankWorth{SpellId: r.id, Level: r.level, Price: int32(r.price), Minutes: math.Round(minutes[r.id]*10) / 10, Rank: int32(r.rank), Spell: r.spell})
	}

	// The spells the plan doesn't cast, and the ones of them we train anyway.
	always := map[string]bool{}
	for _, spell := range settings.AlwaysSpells {
		if _, ok := ranks.bySpell[spell]; !ok {
			return nil, fmt.Errorf("the class has no spell %q to always train", spell)
		}
		if planSpells[spell] {
			return nil, fmt.Errorf("the plan casts %s, so it's scheduled like the plan's other spells", spell)
		}
		always[spell] = true
	}
	var others []*proto.LevelingRankWorth
	for spell, spellRanks := range ranks.bySpell {
		if planSpells[spell] {
			continue
		}
		for _, r := range spellRanks {
			if !r.bought || r.level > maxPlanLevel {
				continue
			}
			others = append(others, &proto.LevelingRankWorth{SpellId: r.id, Level: r.level, Price: int32(r.price), Rank: int32(r.rank), Spell: spell})
			if always[spell] {
				ts.always = append(ts.always, r)
			}
		}
	}
	slices.SortFunc(others, func(a, b *proto.LevelingRankWorth) int { return int(a.SpellId - b.SpellId) })
	alwaysCost := 0.0
	for _, r := range ts.always {
		if r.level <= 40 {
			alwaysCost += r.price
		}
	}
	log.Printf("%d spells we always train, %.1fg by 40", len(always), alwaysCost/10000)

	cast := map[int32]bool{}
	for _, bySpell := range full.top {
		for _, id := range bySpell {
			cast[id] = true
		}
	}
	// Ranks learned before the plan's first level we train at their level. The sims don't play
	// those levels, so they can't say what Earth Shock 1 is worth at 4 to 7, or Healing Wave 2 at
	// 6 to 9.
	// The ranks we hold are bought at their hold level, and the search doesn't move them.
	var choosable []*spellRank
	for _, r := range planRanks {
		if _, held := ts.holds[r.id]; !held && r.level >= ts.levels[0] {
			choosable = append(choosable, r)
		}
	}
	skipped := deferRuns(choosable, ts.rankLoss, cast, skipTails(choosable, minutes))
	for id, level := range ts.holds {
		skipped[id] = level
	}
	if skipped, err = ts.check(ctx, skipped, cast); err != nil {
		return nil, err
	}
	s, cost := ts.plan(planRanks, skipped)
	last := full
	for round := 1; round <= trainingRounds && len(s) > 0; round++ {
		log.Printf("round %d: %d ranks never bought or bought with a later one, %d pushed back, %.1f min by the measures so far", round, len(skipped), len(s)-len(skipped), 60*cost)
		w, err := ts.measure(ctx, s)
		if err != nil {
			return nil, err
		}
		last = w
		next, nextCost := ts.plan(planRanks, skipped)
		if sameSchedule(next, s) {
			break
		}
		s, cost = next, nextCost
		last = nil
	}
	if last == nil {
		if last, err = ts.measure(ctx, s); err != nil {
			return nil, err
		}
	}
	log.Printf("under the budget: %.2fh (%+.1f min)", totalHours(last), 60*(totalHours(last)-totalHours(full)))

	out := &proto.LevelingTrainingSettings{GoldAt_40: settings.GoldAt_40, RidingLevel: settings.RidingLevel, RidingGold: settings.RidingGold, Worth: worthList,
		AlwaysSpells: settings.AlwaysSpells, Others: others, Holds: settings.Holds}
	for id, level := range s {
		out.Buys = append(out.Buys, &proto.LevelingTrainingBuy{SpellId: id, Level: level})
	}
	slices.SortFunc(out.Buys, func(a, b *proto.LevelingTrainingBuy) int { return int(a.SpellId - b.SpellId) })
	ts.report(planRanks, minutes, s)
	return out, nil
}

// report prints the ranks by level, with their price, worth and when we buy them.
func (ts *trainingSearch) report(ranks []*spellRank, minutes map[int32]float64, s schedule) {
	sorted := slices.Clone(ranks)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].level != sorted[j].level {
			return sorted[i].level < sorted[j].level
		}
		return sorted[i].spell < sorted[j].spell
	})
	fmt.Fprintf(os.Stderr, "%-4s %-26s %8s %8s %8s  %s\n", "lvl", "rank", "price", "minutes", "min/g", "buy")
	for _, r := range sorted {
		buy := ""
		if level, ok := s[r.id]; ok {
			buy = "at " + strconv.Itoa(int(level))
			if level == neverLevel {
				buy = "never"
			}
		}
		perGold := 0.0
		if r.price > 0 {
			perGold = minutes[r.id] / (r.price / 10000)
		}
		fmt.Fprintf(os.Stderr, "%-4d %-26s %7.2fg %8.1f %8.1f  %s\n", r.level, fmt.Sprintf("%s %d", r.spell, r.rank), r.price/10000, minutes[r.id], perGold, buy)
	}
	fmt.Fprintf(os.Stderr, "\nlevel  earned  spent  short\n")
	for level := int32(10); level <= maxPlanLevel; level += 5 {
		spent := ts.short(ranks, s, level) + ts.budget.earned(level)
		fmt.Fprintf(os.Stderr, "%5d %6.0fg %5.0fg %5.0fg\n", level, ts.budget.earned(level)/10000, spent/10000, max(0, spent-ts.budget.earned(level))/10000)
	}
}

func trainingFile(db *proto.UIDatabase, in, out string, sessions int32, goldAt40 float64, ridingLevel int32, ridingGold float64, alwaysSpells []string) error {
	data, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	req, err := readRequest(data)
	if err != nil {
		return err
	}
	// The search starts from every rank at its level, so the plan's own schedule doesn't count.
	// The spells we always train are the plan's, unless we name others, and we keep its holds.
	if len(alwaysSpells) == 0 {
		alwaysSpells = req.Plan.GetTraining().GetAlwaysSpells()
	}
	req.Plan.Training = &proto.LevelingTrainingSettings{GoldAt_40: goldAt40, RidingLevel: ridingLevel, RidingGold: ridingGold, AlwaysSpells: alwaysSpells,
		Holds: req.Plan.GetTraining().GetHolds()}
	p, err := newPlanner(req, db)
	if err != nil {
		return err
	}
	result, err := searchTraining(context.Background(), p, sessions)
	if err != nil {
		return err
	}
	log.Printf("%d sims", p.sims.Load())
	data, err = protojson.MarshalOptions{Multiline: true}.Marshal(result)
	if err != nil {
		return err
	}
	if out == "" {
		_, err = os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(out, data, 0o644)
}
