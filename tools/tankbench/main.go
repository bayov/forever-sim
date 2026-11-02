// Command tankbench sims tank builds against a boss that attacks them, with a stand-in healer,
// and prints what tank builds trade: threat against damage taken, spikes and the chance of
// death, next to the tank sheet's thresholds (sim/core/tank_sheet.go). See docs/tank_sim.md.
//
//	go run -tags=with_db ./tools/tankbench tools/tankbench/req/*.json
//
// Each request is a RaidSimRequest with the tank as the first player, the way the DPS Report's
// requests are. The boss is made to attack it, and the healer replaces whatever healing model
// the request has.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	hps := flag.Float64("hps", 600, "the stand-in healer's healing per second on the tank")
	cadence := flag.Float64("cadence", 2, "seconds between the stand-in healer's heals on average")
	variation := flag.Float64("cadence-variation", 1, "how far a heal's timing varies from the cadence, in seconds")
	window := flag.Int("tmi-window", 6, "seconds of the TMI's moving window")
	iterations := flag.Int("iters", 3000, "iterations per build")
	duration := flag.Float64("dur", 0, "fight length in seconds, 0 keeps the request's")
	bossLevel := flag.Int("boss-level", 0, "the boss's level, 0 keeps the request's")
	swing := flag.Float64("boss-damage", 0, "the boss's smallest white hit before armor, 0 keeps the request's")
	actions := flag.Bool("actions", false, "also print each build's threat and damage by ability")
	flag.Parse()
	sim.RegisterAll()

	fmt.Printf("Stand-in healer: %.0f HPS every %.1f ± %.1f s. TMI over %d s windows. %d iterations.\n\n",
		*hps, *cadence, *variation, *window, *iterations)
	header := []string{"Build", "TPS", "DPS", "DTPS", "TMI", "Death", "Health", "EH", "Armor", "Def", "No crit at", "Miss", "Dodge", "Parry", "Block", "Crit", "Crush", "Crush gap", "Took: avoid", "block", "crit", "crush"}
	fmt.Println("| " + strings.Join(header, " | ") + " |")
	fmt.Println("|" + strings.Repeat("---|", len(header)))

	var breakdown []string
	for _, path := range flag.Args() {
		data, err := os.ReadFile(path)
		if err != nil {
			fail(err)
		}
		rsr := &proto.RaidSimRequest{}
		if err := protojson.Unmarshal(data, rsr); err != nil {
			fail(fmt.Errorf("%s: %w", path, err))
		}
		rsr.SimOptions.Iterations = int32(*iterations)
		rsr.Raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
		if *duration > 0 {
			rsr.Encounter.Duration = *duration
		}
		for _, target := range rsr.Encounter.Targets {
			target.TankIndex = 0
			if *bossLevel > 0 {
				target.Level = int32(*bossLevel)
			}
			if *swing > 0 {
				target.MinBaseDamage = *swing
			}
		}
		tank := rsr.Raid.Parties[0].Players[0]
		tank.HealingModel = &proto.HealingModel{
			Hps:              *hps,
			CadenceSeconds:   *cadence,
			CadenceVariation: *variation,
			BurstWindow:      int32(*window),
		}

		res := core.RunRaidSimConcurrent(rsr)
		if res.Error != nil {
			fail(fmt.Errorf("%s: %s", path, res.Error.Message))
		}
		m := res.RaidMetrics.Parties[0].Players[0]
		s := m.TankSheet
		if s == nil {
			fail(fmt.Errorf("%s: the boss didn't attack the tank", path))
		}
		took := tookShares(res.EncounterMetrics.Targets[0], m.UnitIndex)
		pct := func(x float64) string { return fmt.Sprintf("%.1f%%", 100*x) }
		row := []string{
			strings.TrimSuffix(filepath.Base(path), ".json"),
			fmt.Sprintf("%.0f", m.Threat.Avg),
			fmt.Sprintf("%.0f", m.Dps.Avg),
			fmt.Sprintf("%.0f", m.Dtps.Avg),
			fmt.Sprintf("%.0f", m.Tmi.Avg),
			pct(m.ChanceOfDeath),
			fmt.Sprintf("%.0f", s.Health),
			fmt.Sprintf("%.0f", s.EffectiveHealth),
			fmt.Sprintf("%.0f (%s)", s.Armor, pct(s.ArmorReduction)),
			fmt.Sprintf("%.0f", s.Defense),
			fmt.Sprintf("%.0f", s.DefenseCritImmune),
			pct(s.Miss), pct(s.Dodge), pct(s.Parry), pct(s.Block), pct(s.Crit), pct(s.Crush), pct(s.CrushGap),
			pct(took.avoided), pct(took.blocked), pct(took.crit), pct(took.crush),
		}
		fmt.Println("| " + strings.Join(row, " | ") + " |")
		if *actions {
			breakdown = append(breakdown, actionLines(row[0], m, float64(*iterations), res.AvgIterationDuration)...)
		}
	}
	for _, line := range breakdown {
		fmt.Println(line)
	}
}

// actionLines is a build's threat and damage per second by ability, most threat first.
func actionLines(name string, m *proto.UnitMetrics, n, seconds float64) []string {
	type line struct {
		id             string
		casts          int32
		threat, damage float64
	}
	var lines []line
	for _, a := range m.Actions {
		l := line{id: a.Id.String()}
		for _, t := range a.Targets {
			l.casts += t.Casts
			l.threat += t.Threat
			l.damage += t.Damage
		}
		lines = append(lines, l)
	}
	sort.Slice(lines, func(i, j int) bool { return lines[i].threat > lines[j].threat })
	out := []string{"", name + ": threat and damage per second by ability (casts per fight)"}
	for _, l := range lines {
		out = append(out, fmt.Sprintf("  %-60s %7.1f TPS %7.1f DPS %6.1f casts", l.id, l.threat/n/seconds, l.damage/n/seconds, float64(l.casts)/n))
	}
	return out
}

type shares struct{ avoided, blocked, crit, crush float64 }

// tookShares is how the boss's melee swings at the tank landed over the whole fight, with
// Shield Block, Holy Shield and the like in.
func tookShares(boss *proto.UnitMetrics, tankIndex int32) shares {
	var n, avoided, blocked, crit, crush float64
	for _, action := range boss.Actions {
		if !action.IsMelee {
			continue
		}
		for _, t := range action.Targets {
			if t.UnitIndex != tankIndex {
				continue
			}
			avoided += float64(t.Misses + t.Dodges + t.Parries)
			blocked += float64(t.Blocks + t.BlockedCrits)
			crit += float64(t.Crits)
			crush += float64(t.Crushes)
			n += float64(t.Misses+t.Dodges+t.Parries+t.Blocks+t.BlockedCrits+t.Crits+t.Crushes+t.Hits) + float64(t.Glances)
		}
	}
	if n == 0 {
		return shares{}
	}
	return shares{avoided / n, blocked / n, crit / n, crush / n}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
