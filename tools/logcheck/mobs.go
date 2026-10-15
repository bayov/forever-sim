package main

import (
	"fmt"
	"maps"
	"math"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// The advanced fields of the unit a line describes, counted from infoStart. On a mob's swing
// at us that unit is the mob.
const (
	infoMaxHealth = 3
	infoArmor     = 6
)

// runMobs measures the mobs that swung at our characters in the archived logs (or just
// logPath), for the leveling sim's normal mob table (core.NormalMobStatsAt,
// docs/leveling_sim.md). By mob level it prints the mobs' health, armor, swing timer and raw
// swing damage next to the table's. By the mob's level less ours it prints how their white
// swings landed on us.
//
// Elites, rares and bosses are in the log too, and the log doesn't say which they are, so a
// level's averages carry some of them. The medians are the safer numbers.
func runMobs(archive, flavor, logPath string) error {
	logs := []string{logPath}
	if logPath == "" {
		logs = archivedLogs(filepath.Join(archive, flavor))
	}
	if len(logs) == 0 {
		return fmt.Errorf("no combat logs in %s", filepath.Join(archive, flavor))
	}
	exports, err := loadExports(archive)
	if err != nil {
		return err
	}
	byChar := map[string][]*export{}
	for _, ex := range exports {
		byChar[charKey(ex)] = append(byChar[charKey(ex)], ex)
	}
	for _, list := range byChar {
		slices.SortFunc(list, func(a, b *export) int { return a.time.Compare(b.time) })
	}

	m := &mobTally{levels: map[int]*mobLevel{}, diffs: map[int]*swingOutcomes{}, seen: map[string]bool{}}
	for _, path := range logs {
		events, err := loadLog(path)
		if err != nil {
			return err
		}
		m.read(events, byChar)
	}
	m.print()
	return nil
}

// ourLevel is the character's level at t: the newest export or history record at or before
// t, or the first one after it.
func ourLevel(list []*export, t time.Time) int {
	level := 0
	for _, ex := range list {
		if ex.time.After(t) {
			if level == 0 {
				level = int(ex.Who.Level)
			}
			break
		}
		level = int(ex.Who.Level)
	}
	return level
}

type swingOutcomes struct {
	swings                                 int
	miss, dodge, parry, block, crit, crush int
}

type mobLevel struct {
	health, armor []float64 // one entry per mob
	gaps          []float64 // seconds between a mob's swings at us, under 6 sec
	raw           []float64 // raw damage of its landed swings, before our armor and the crit
}

type mobTally struct {
	levels map[int]*mobLevel
	diffs  map[int]*swingOutcomes
	seen   map[string]bool // mob GUIDs whose health and armor we have
	ours   int             // swings at our characters
	noLvl  int             // of those, with our level or the mob's unknown
}

func (m *mobTally) level(l int) *mobLevel {
	if m.levels[l] == nil {
		m.levels[l] = &mobLevel{}
	}
	return m.levels[l]
}

func (m *mobTally) read(events []*event, byChar map[string][]*export) {
	// Mob levels come from every line the mob is the info unit of, as in newCombatLog.
	mobLevels := map[string]int{}
	for _, e := range events {
		if guid := e.infoGUID(); isCreature(guid) {
			if level := e.infoInt(infoLevel); level > 0 {
				mobLevels[guid] = level
			}
		}
	}
	// Our characters' GUIDs, from the names of the ones we have exports of.
	chars := map[string]string{}
	for _, e := range events {
		for _, unit := range [][2]string{{e.srcGUID, e.src}, {e.dstGUID, e.dst}} {
			if !strings.HasPrefix(unit[0], "Player-") || chars[unit[0]] != "" {
				continue
			}
			for char := range byChar {
				if isChar(unit[1], char) {
					chars[unit[0]] = char
				}
			}
		}
	}

	lastSwing := map[string]time.Time{}
	for _, e := range events {
		if (e.kind != "SWING_DAMAGE" && e.kind != "SWING_MISSED") || !isCreature(e.srcGUID) {
			continue
		}
		char := chars[e.dstGUID]
		if char == "" {
			continue
		}
		m.ours++
		mobLvl, ourLvl := mobLevels[e.srcGUID], ourLevel(byChar[char], e.t)
		if mobLvl == 0 || ourLvl == 0 {
			m.noLvl++
			continue
		}
		ml := m.level(mobLvl)
		if prev, ok := lastSwing[e.srcGUID]; ok {
			if gap := e.t.Sub(prev).Seconds(); gap > 0 && gap < 6 {
				ml.gaps = append(ml.gaps, gap)
			}
		}
		lastSwing[e.srcGUID] = e.t

		d := m.diffs[mobLvl-ourLvl]
		if d == nil {
			d = &swingOutcomes{}
			m.diffs[mobLvl-ourLvl] = d
		}
		if e.kind == "SWING_MISSED" {
			switch e.missType() {
			case "MISS":
				d.miss++
			case "DODGE":
				d.dodge++
			case "PARRY":
				d.parry++
			case "BLOCK":
				d.block++
			default: // ABSORB, IMMUNE, EVADE: nothing to count
				continue
			}
			d.swings++
			continue
		}

		dmg, ok := e.damage()
		if !ok {
			continue
		}
		d.swings++
		switch {
		case dmg.crit:
			d.crit++
		case dmg.crush:
			d.crush++
		case dmg.blocked > 0:
			d.block++
		}
		if dmg.raw > 0 {
			ml.raw = append(ml.raw, float64(dmg.raw))
		}
		if !m.seen[e.srcGUID] {
			if hp, err := strconv.ParseFloat(e.f[e.infoStart()+infoMaxHealth], 64); err == nil && hp > 0 {
				armor, _ := strconv.ParseFloat(e.f[e.infoStart()+infoArmor], 64)
				ml.health = append(ml.health, hp)
				ml.armor = append(ml.armor, armor)
				m.seen[e.srcGUID] = true
			}
		}
	}
}

func median(xs []float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	s := slices.Clone(xs)
	sort.Float64s(s)
	if len(s)%2 == 1 {
		return s[len(s)/2]
	}
	return (s[len(s)/2-1] + s[len(s)/2]) / 2
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}

func (m *mobTally) print() {
	fmt.Printf("%d mob swings at our characters, %d left out with our level or the mob's unknown\n\n", m.ours, m.noLvl)

	fmt.Println("By mob level, the log (median, mean) next to the sim's normal mob table:")
	fmt.Println("level  mobs  health log         sim    armor log          sim    swing sec log   sim    swings  raw log         sim range")
	for _, l := range slices.Sorted(maps.Keys(m.levels)) {
		ml := m.levels[l]
		sim := "      -"
		simArmor, simSwing, simRaw := "     -", "    -", "        -"
		if l >= 1 && l <= 60 {
			s := core.NormalMobStatsAt(int32(l))
			sim = fmt.Sprintf("%7.0f", s.Health)
			simArmor = fmt.Sprintf("%6.0f", s.Armor)
			simSwing = fmt.Sprintf("%5.2f", s.SwingSpeed)
			simRaw = fmt.Sprintf("%4.0f-%-4.0f", s.MinDamage, s.MaxDamage)
		}
		fmt.Printf("%5d  %4d  %6.0f %6.0f  %s  %6.0f %6.0f  %s  %5.2f  %s  %6d  %5.1f %5.1f  %s\n",
			l, len(ml.health), median(ml.health), mean(ml.health), sim, median(ml.armor), mean(ml.armor), simArmor,
			median(ml.gaps), simSwing, len(ml.raw), median(ml.raw), mean(ml.raw), simRaw)
	}

	fmt.Println()
	fmt.Println("How their white swings landed on us, by the mob's level less ours. The sim's Classic table")
	fmt.Println("gives 5% miss and 5% crit at our level, 0.2% each a level, and crushing blows from +3:")
	fmt.Println(" diff  swings   miss          dodge         parry         block         crit          crush")
	for _, diff := range slices.Sorted(maps.Keys(m.diffs)) {
		d := m.diffs[diff]
		rate := func(n int) string {
			if d.swings == 0 {
				return "-"
			}
			p := float64(n) / float64(d.swings)
			return fmt.Sprintf("%5.1f%% +-%4.1f", 100*p, 196*math.Sqrt(p*(1-p)/float64(d.swings)))
		}
		fmt.Printf("%+5d  %6d   %s  %s  %s  %s  %s  %s\n", diff, d.swings,
			rate(d.miss), rate(d.dodge), rate(d.parry), rate(d.block), rate(d.crit), rate(d.crush))
	}
}
