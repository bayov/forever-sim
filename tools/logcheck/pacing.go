package main

import (
	"fmt"
	"maps"
	"math"
	"slices"
	"sort"
	"strings"
	"time"
)

// Where a unit's position sits in its advanced fields, counted from infoStart. The game gives
// it in yards.
const (
	infoX = 14
	infoY = 15
)

// The pacing measurements leave out what isn't focused grinding.
const (
	// A gap between two fights longer than this, after taking out the rests, is AFK or
	// something other than grinding.
	pacingMaxGap = 60 * time.Second
	// A mob we leave alone for this long is no longer in the fight.
	pacingQuiet = 15 * time.Second
	// Mobs more than this many levels from ours aren't the ones we grind.
	pacingLevels = 3
)

// The auras on us that mean we sat down to rest.
var restAuras = map[string]bool{"Food": true, "Drink": true, "Refreshment": true, "First Aid": true}

// runPacing measures how our characters paced their grinding in the archived logs (or just
// logPath), for the leveling sim's session defaults (docs/leveling_sim.md, "From the logs").
//
// It looks at the mobs within 3 levels of ours that we fought and killed. For each kill it
// takes the time to the next pull, less the time we spent eating, drinking, bandaging or
// casting a heal, and drops gaps over 60 sec. For each pull it takes how we pulled, from how
// far, and how long the mob took to reach us. For each fight it counts the mobs and how they
// joined. from and to limit it to a stretch of play, like -from and -to do for one window.
func runPacing(archive, flavor, logPath, fromFlag, toFlag string) error {
	logs := []string{logPath}
	if logPath == "" {
		logs = flavorLogs(archive, flavor)
	}
	if len(logs) == 0 {
		return fmt.Errorf("no combat logs in %s", archive)
	}
	exports, err := loadExports(archive)
	if err != nil {
		return err
	}
	byChar := map[string][]*export{}
	for _, ex := range exports {
		if ex.guid != "" {
			byChar[ex.guid] = append(byChar[ex.guid], ex)
		}
	}
	for _, list := range byChar {
		slices.SortFunc(list, func(a, b *export) int { return a.time.Compare(b.time) })
	}

	p := &pacing{hours: map[string]*pacingHour{}}
	for _, path := range logs {
		events, err := loadLog(path)
		if err != nil {
			return err
		}
		if len(events) == 0 {
			continue
		}
		from, err := parseWindow(fromFlag, events[0].t, events[0].t)
		if err != nil {
			return err
		}
		to, err := parseWindow(toFlag, events[0].t, events[len(events)-1].t)
		if err != nil {
			return err
		}
		p.read(events, byChar, from, to)
	}
	p.print()
	return nil
}

// engagement is our time on one mob, from the first attack between us to its death, or to the
// last attack when we left it alive.
type engagement struct {
	guid  string
	level int

	first time.Time
	// How it started: "spell" (one of ours landed first, spell says which), "melee" (our
	// white swing), "mob" (it attacked us first).
	pulledBy, spell string

	// Where we and it stood when our pull landed, when the log has both.
	ourPos, mobPos [2]float64
	hasPos         bool
	landed         time.Time

	firstSwing time.Time // its first attack on us
	last       time.Time
	died       time.Time
	shared     bool    // another player hit it
	damage     float64 // what we and our totems did to it
}

func (g *engagement) end() time.Time {
	if !g.died.IsZero() {
		return g.died
	}
	return g.last
}

// pacingFight is the mobs we fought together, one after another while some were still alive.
type pacingFight struct {
	char       string
	mobs       []*engagement
	start, end time.Time
}

type pacingHour struct {
	kills         int
	fightTime     float64
	damage        float64
	fightLengths  []float64
	nonGrindKills int
}

type pacing struct {
	hours map[string]*pacingHour

	pulls        map[string]int
	pullDistance map[string][]float64 // by pull spell
	runIn        map[string][]float64 // by pull spell, seconds from the pull landing to its first swing
	mobSpeed     []float64            // yd/s, from spell pulls over 8 yd

	gaps       []float64            // kill to next pull, less rests, up to pacingMaxGap
	gapsBy     map[string][]float64 // by how the next mob was pulled
	rawGaps    []float64            // the same gaps before taking the rests out
	longGaps   int
	deathGaps  int
	restGaps   int
	restTime   []float64
	travel     []float64 // yd between our spot at the kill and at the next pull
	travelRate []float64 // yd/s over the gap
	wolfGaps   int

	mobsPerFight map[int]int
	mostAtOnce   map[int]int
	joined       map[string][]float64 // later mobs in a fight: seconds after the mob before joined, by who started it
	fights       int
	sharedFights int
}

func (p *pacing) init() {
	if p.pulls != nil {
		return
	}
	p.pulls = map[string]int{}
	p.pullDistance = map[string][]float64{}
	p.runIn = map[string][]float64{}
	p.gapsBy = map[string][]float64{}
	p.mobsPerFight = map[int]int{}
	p.mostAtOnce = map[int]int{}
	p.joined = map[string][]float64{}
}

// pacingChar is what we follow of one of our characters in a log.
type pacingChar struct {
	char     string
	track    []trackPoint // our positions, from lines where we're the info unit
	rests    []restSpan   // eating, drinking, bandaging and casting heals
	deaths   []time.Time
	wolf     []time.Time // Ghost Wolf casts
	open     map[string]*engagement
	all      []*engagement
	casting  map[string]time.Time // heals we started casting, by spell name
	ourUnits map[string]bool      // our GUID and our totems'
}

type trackPoint struct {
	t    time.Time
	x, y float64
}

type restSpan struct{ from, to time.Time }

func (p *pacing) read(events []*event, byChar map[string][]*export, from, to time.Time) {
	p.init()
	mobLevels := map[string]int{}
	for _, e := range events {
		if guid := e.infoGUID(); isCreature(guid) {
			if level := e.infoInt(infoLevel); level > 0 {
				mobLevels[guid] = level
			}
		}
	}
	// Our characters by GUID, the ones we have exports of, and their totems.
	chars := map[string]*pacingChar{}
	owner := map[string]*pacingChar{}
	for _, e := range events {
		for _, guid := range []string{e.srcGUID, e.dstGUID} {
			if byChar[guid] == nil || chars[guid] != nil {
				continue
			}
			c := &pacingChar{char: guid, open: map[string]*engagement{}, casting: map[string]time.Time{}, ourUnits: map[string]bool{guid: true}}
			chars[guid] = c
			owner[guid] = c
		}
	}
	if len(chars) == 0 {
		return
	}

	for _, e := range events {
		if e.t.Before(from) || e.t.After(to) {
			continue
		}
		if c := chars[e.srcGUID]; c != nil && e.kind == "SPELL_SUMMON" {
			c.ourUnits[e.dstGUID] = true
			owner[e.dstGUID] = c
			continue
		}
		if c := chars[e.infoGUID()]; c != nil && e.infoStart() >= 0 {
			if x, y, ok := e.position(); ok {
				c.track = append(c.track, trackPoint{e.t, x, y})
			}
		}
		if c := chars[e.dstGUID]; c != nil {
			switch {
			case e.kind == "UNIT_DIED":
				c.deaths = append(c.deaths, e.t)
			case e.kind == "SPELL_AURA_APPLIED" && restAuras[e.spell]:
				c.rests = append(c.rests, restSpan{e.t, e.t})
			case e.kind == "SPELL_AURA_REMOVED" && restAuras[e.spell]:
				for i := len(c.rests) - 1; i >= 0; i-- {
					if c.rests[i].from == c.rests[i].to {
						c.rests[i].to = e.t
						break
					}
				}
			}
		}
		if c := chars[e.srcGUID]; c != nil && strings.Contains(e.spell, "Heal") {
			switch e.kind {
			case "SPELL_CAST_START":
				c.casting[e.spell] = e.t
			case "SPELL_CAST_SUCCESS", "SPELL_CAST_FAILED":
				if start, ok := c.casting[e.spell]; ok {
					c.rests = append(c.rests, restSpan{start, e.t})
					delete(c.casting, e.spell)
				}
			}
		}
		if c := chars[e.srcGUID]; c != nil && e.kind == "SPELL_CAST_SUCCESS" && e.spell == "Ghost Wolf" {
			c.wolf = append(c.wolf, e.t)
		}

		if e.kind == "UNIT_DIED" && isCreature(e.dstGUID) {
			for _, c := range chars {
				if g := c.open[e.dstGUID]; g != nil && e.t.Sub(g.last) <= pacingQuiet {
					g.died = e.t
					delete(c.open, e.dstGUID)
				}
			}
			continue
		}
		if !isAttack(e.kind) {
			continue
		}
		var c *pacingChar
		var mob string
		ourAttack := false
		switch {
		case owner[e.srcGUID] != nil && isCreature(e.dstGUID) && owner[e.dstGUID] == nil:
			c, mob, ourAttack = owner[e.srcGUID], e.dstGUID, true
		case chars[e.dstGUID] != nil && isCreature(e.srcGUID) && owner[e.srcGUID] == nil:
			c, mob = chars[e.dstGUID], e.srcGUID
		case strings.HasPrefix(e.srcGUID, "Player-") && chars[e.srcGUID] == nil && isCreature(e.dstGUID):
			for _, c := range chars {
				if g := c.open[e.dstGUID]; g != nil {
					g.shared = true
				}
			}
			continue
		default:
			continue
		}
		g := c.open[mob]
		if g != nil && e.t.Sub(g.last) > pacingQuiet {
			delete(c.open, mob)
			g = nil
		}
		if g == nil {
			g = &engagement{guid: mob, level: mobLevels[mob], first: e.t}
			switch {
			case !ourAttack:
				g.pulledBy = "mob"
			case strings.HasPrefix(e.kind, "SWING_"):
				g.pulledBy = "melee"
			default:
				g.pulledBy, g.spell = "spell", e.spell
			}
			c.open[mob] = g
			c.all = append(c.all, g)
		}
		g.last = e.t
		if ourAttack {
			if d, ok := e.damage(); ok {
				g.damage += float64(d.amount)
			}
			// Our pull spell landed: its line has the mob's position, and our last line
			// before it has ours.
			if g.pulledBy == "spell" && !g.hasPos && e.t.Sub(g.first) <= 2*time.Second && e.infoGUID() == mob {
				if mx, my, ok := e.position(); ok {
					if pt, ok := c.positionAt(e.t, time.Second); ok {
						g.ourPos, g.mobPos, g.hasPos, g.landed = [2]float64{pt.x, pt.y}, [2]float64{mx, my}, true, e.t
					}
				}
			}
		} else if g.firstSwing.IsZero() {
			g.firstSwing = e.t
		}
	}

	for _, c := range chars {
		p.readChar(c, byChar[c.char])
	}
}

// position is where the line's info unit stood.
func (e *event) position() (float64, float64, bool) {
	start := e.infoStart()
	if start < 0 {
		return 0, 0, false
	}
	var x, y float64
	if _, err := fmt.Sscan(e.f[start+infoX], &x); err != nil {
		return 0, 0, false
	}
	if _, err := fmt.Sscan(e.f[start+infoY], &y); err != nil {
		return 0, 0, false
	}
	return x, y, true
}

// positionAt is where we stood at t: the nearest point of our track within slack of it.
func (c *pacingChar) positionAt(t time.Time, slack time.Duration) (trackPoint, bool) {
	i := sort.Search(len(c.track), func(i int) bool { return c.track[i].t.After(t) })
	best, found := trackPoint{}, false
	bestGap := slack + 1
	for _, j := range []int{i - 1, i} {
		if j < 0 || j >= len(c.track) {
			continue
		}
		gap := c.track[j].t.Sub(t)
		if gap < 0 {
			gap = -gap
		}
		if gap <= slack && gap < bestGap {
			best, bestGap, found = c.track[j], gap, true
		}
	}
	return best, found
}

func (c *pacingChar) restBetween(from, to time.Time) time.Duration {
	var total time.Duration
	for _, r := range c.rests {
		a, b := r.from, r.to
		if a.Before(from) {
			a = from
		}
		if b.After(to) {
			b = to
		}
		if b.After(a) {
			total += b.Sub(a)
		}
	}
	return total
}

func (c *pacingChar) anyBetween(ts []time.Time, from, to time.Time) bool {
	for _, t := range ts {
		if !t.Before(from) && !t.After(to) {
			return true
		}
	}
	return false
}

func (p *pacing) readChar(c *pacingChar, exports []*export) {
	slices.SortFunc(c.all, func(a, b *engagement) int { return a.first.Compare(b.first) })
	// Only the kills of mobs near our level make fights. A fight where we left a mob alive, or
	// that has a mob far from our level, still splits the gaps around it, but we don't
	// measure it.
	var fights []*pacingFight
	for _, g := range c.all {
		if n := len(fights); n > 0 && !g.first.After(fights[n-1].end) {
			f := fights[n-1]
			f.mobs = append(f.mobs, g)
			if g.end().After(f.end) {
				f.end = g.end()
			}
			continue
		}
		fights = append(fights, &pacingFight{char: c.char, mobs: []*engagement{g}, start: g.first, end: g.end()})
	}
	grind := func(f *pacingFight) bool {
		ourLvl := ourLevel(exports, f.start)
		for _, g := range f.mobs {
			if g.died.IsZero() || g.level == 0 || ourLvl == 0 || abs(g.level-ourLvl) > pacingLevels {
				return false
			}
		}
		return true
	}

	for i, f := range fights {
		hour := p.hour(f.start)
		if !grind(f) {
			hour.nonGrindKills += len(f.mobs)
			continue
		}
		shared := false
		for _, g := range f.mobs {
			shared = shared || g.shared
		}
		p.fights++
		if shared {
			p.sharedFights++
			continue
		}
		length := f.end.Sub(f.start).Seconds()
		hour.kills += len(f.mobs)
		hour.fightTime += length
		hour.fightLengths = append(hour.fightLengths, length)
		for _, g := range f.mobs {
			hour.damage += g.damage
		}

		// How the fight's first mob was pulled.
		first := f.mobs[0]
		pull := pullName(first)
		p.pulls[pull]++
		if first.hasPos {
			d := math.Hypot(first.ourPos[0]-first.mobPos[0], first.ourPos[1]-first.mobPos[1])
			p.pullDistance[pull] = append(p.pullDistance[pull], d)
			if !first.firstSwing.IsZero() {
				runIn := first.firstSwing.Sub(first.landed).Seconds()
				p.runIn[pull] = append(p.runIn[pull], runIn)
				if d > 8 && runIn > 0.5 {
					p.mobSpeed = append(p.mobSpeed, (d-5)/runIn)
				}
			}
		}

		// The mobs that joined it.
		p.mobsPerFight[len(f.mobs)]++
		p.mostAtOnce[mostAtOnce(f.mobs)]++
		for j := 1; j < len(f.mobs); j++ {
			who := "we pulled it"
			if f.mobs[j].pulledBy == "mob" {
				who = "it came to us"
			}
			p.joined[who] = append(p.joined[who], f.mobs[j].first.Sub(f.mobs[j-1].first).Seconds())
		}

		// The gap to the next fight, when that one is a grinding fight too.
		if i+1 >= len(fights) || !grind(fights[i+1]) {
			continue
		}
		next := fights[i+1]
		if c.anyBetween(c.deaths, f.start, next.start) {
			p.deathGaps++
			continue
		}
		raw := next.start.Sub(f.end)
		rest := c.restBetween(f.end, next.start)
		gap := raw - rest
		if gap > pacingMaxGap {
			p.longGaps++
			continue
		}
		if rest > 0 {
			p.restGaps++
			p.restTime = append(p.restTime, rest.Seconds())
		}
		p.gaps = append(p.gaps, gap.Seconds())
		p.rawGaps = append(p.rawGaps, raw.Seconds())
		kind := next.mobs[0].pulledBy
		p.gapsBy[kind] = append(p.gapsBy[kind], gap.Seconds())
		a, okA := c.positionAt(f.end, 5*time.Second)
		b, okB := c.positionAt(next.start, 5*time.Second)
		if okA && okB {
			d := math.Hypot(a.x-b.x, a.y-b.y)
			p.travel = append(p.travel, d)
			if gap > 2*time.Second {
				p.travelRate = append(p.travelRate, d/gap.Seconds())
			}
		}
		if c.anyBetween(c.wolf, f.end, next.start) {
			p.wolfGaps++
		}
	}
}

func pullName(g *engagement) string {
	switch g.pulledBy {
	case "spell":
		return g.spell
	case "melee":
		return "our white swing"
	default:
		return "the mob attacked first"
	}
}

// mostAtOnce is the most of the mobs that were on us at the same time.
func mostAtOnce(mobs []*engagement) int {
	type edge struct {
		t     time.Time
		delta int
	}
	var edges []edge
	for _, g := range mobs {
		edges = append(edges, edge{g.first, 1}, edge{g.end(), -1})
	}
	slices.SortFunc(edges, func(a, b edge) int {
		if c := a.t.Compare(b.t); c != 0 {
			return c
		}
		return a.delta - b.delta
	})
	most, now := 0, 0
	for _, e := range edges {
		now += e.delta
		most = max(most, now)
	}
	return most
}

func (p *pacing) hour(t time.Time) *pacingHour {
	key := t.Format("2006-01-02 15h")
	if p.hours[key] == nil {
		p.hours[key] = &pacingHour{}
	}
	return p.hours[key]
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// quantile is the q quantile of xs, by linear interpolation between the closest ranks.
func quantile(xs []float64, q float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	s := slices.Clone(xs)
	sort.Float64s(s)
	pos := q * float64(len(s)-1)
	lo := int(math.Floor(pos))
	hi := min(lo+1, len(s)-1)
	return s[lo] + (pos-float64(lo))*(s[hi]-s[lo])
}

// sample prints a sample as its count, fast quartile, median, mean and slow quartile.
func sample(xs []float64) string {
	if len(xs) == 0 {
		return "none"
	}
	return fmt.Sprintf("%4d  %6.1f  %6.1f  %6.1f  %6.1f", len(xs), quantile(xs, 0.25), median(xs), mean(xs), quantile(xs, 0.75))
}

func (p *pacing) print() {
	p.init()
	fmt.Println("Our kills by hour, of mobs within 3 levels of ours (a test stretch shows as low damage a second):")
	fmt.Println("hour            kills  fight sec (median)  damage/sec  other kills")
	for _, key := range slices.Sorted(maps.Keys(p.hours)) {
		h := p.hours[key]
		dps := math.NaN()
		if h.fightTime > 0 {
			dps = h.damage / h.fightTime
		}
		fmt.Printf("%s  %5d  %18.1f  %10.0f  %11d\n", key, h.kills, median(h.fightLengths), dps, h.nonGrindKills)
	}
	fmt.Printf("\n%d fights, %d left out because another player hit one of the mobs\n", p.fights, p.sharedFights)

	fmt.Println()
	fmt.Println("                                          n    p25  median    mean     p75")
	fmt.Printf("Kill to next pull, less rests (sec)   %s\n", sample(p.gaps))
	for _, kind := range slices.Sorted(maps.Keys(p.gapsBy)) {
		fmt.Printf("  when the next pull was %-12s  %s\n", map[string]string{"spell": "a spell", "melee": "a swing", "mob": "the mob's"}[kind], sample(p.gapsBy[kind]))
	}
	fmt.Printf("Kill to next pull, with rests (sec)   %s\n", sample(p.rawGaps))
	fmt.Printf("Rest in a gap (sec)                   %s\n", sample(p.restTime))
	fmt.Printf("Our travel in the gap (yd)            %s\n", sample(p.travel))
	fmt.Printf("Our travel speed in the gap (yd/s)    %s\n", sample(p.travelRate))
	fmt.Printf("%d gaps over %.0f sec and %d with a death of ours left out. %d of %d gaps had a rest, %d had a Ghost Wolf cast.\n",
		p.longGaps, pacingMaxGap.Seconds(), p.deathGaps, p.restGaps, len(p.gaps), p.wolfGaps)

	fmt.Println()
	fmt.Println("How we pulled the first mob of a fight:")
	total := 0
	for _, n := range p.pulls {
		total += n
	}
	keys := slices.Sorted(maps.Keys(p.pulls))
	slices.SortStableFunc(keys, func(a, b string) int { return p.pulls[b] - p.pulls[a] })
	for _, k := range keys {
		fmt.Printf("  %-24s %4d  %4.0f%%\n", k, p.pulls[k], 100*float64(p.pulls[k])/float64(max(total, 1)))
	}
	fmt.Println("                                          n    p25  median    mean     p75")
	for _, k := range keys {
		if len(p.pullDistance[k]) > 0 {
			fmt.Printf("  %-18s from (yd)          %s\n", k, sample(p.pullDistance[k]))
			fmt.Printf("  %-18s run-in (sec)       %s\n", k, sample(p.runIn[k]))
		}
	}
	fmt.Printf("  Mob run speed, (yd - 5) / run-in     %s\n", sample(p.mobSpeed))

	fmt.Println()
	fmt.Println("Mobs a fight, and the most on us at once:")
	for _, n := range slices.Sorted(maps.Keys(p.mobsPerFight)) {
		fmt.Printf("  %d mobs: %4d fights\n", n, p.mobsPerFight[n])
	}
	for _, n := range slices.Sorted(maps.Keys(p.mostAtOnce)) {
		fmt.Printf("  %d at once: %4d fights\n", n, p.mostAtOnce[n])
	}
	fmt.Println("A later mob in a fight joined this many sec after the one before it:")
	for _, who := range slices.Sorted(maps.Keys(p.joined)) {
		fmt.Printf("  %-14s                        %s\n", who, sample(p.joined[who]))
	}
}
