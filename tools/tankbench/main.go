// Command tankbench sims tank builds against a boss that attacks them, with a stand-in healer,
// and prints what tank builds trade: threat against damage taken, spikes and the chance of
// death, next to the tank sheet's thresholds (sim/core/tank_sheet.go). See docs/tank_sim.md.
//
//	go run -tags=with_db ./tools/tankbench tools/tankbench/req/*.json
//
// Each request is a RaidSimRequest with the tank as the first player, the way the DPS Report's
// requests are. The boss is made to attack it, and the healer replaces whatever healing model
// the request has.
//
// With -encounter, the fight is a tank encounter (tools/encounters/*.json, shared with the healing sim): a boss's
// melee and abilities and the healing a raid puts on its tank, as healers on the tank, each
// one's healing per second and their heals' cast time. Until the healing sim's healers heal
// (docs/healing_sim.md), that healing is the engine's stand-in stream: the healers' HPS added up,
// a heal every cast time over the healers, each about one healer's heal.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/tools/tankscore"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
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
	classicCrit := flag.Bool("classic-crit", false, "the boss crits at Classic's 5.6% (less our Defense past 300) instead of Forever's 0.6% (docs/tank_sim.md, Forever's boss crit)")
	bisect := flag.Bool("bisect", false, "also find healing needed and the death floor by bisection over whole sims (about 10 more sims a build), to check the replay's")
	deathTarget := flag.Float64("death", 0.05, "the chance of death healing needed aims at (the user's ruling for tanks: 5%)")
	dpsStart := flag.Float64("dps-start", 2, "SafeDPS: seconds after the pull the reference DPS member starts")
	dpsThreat := flag.Float64("dps-threat", 1, "SafeDPS: the reference DPS member's threat a point of damage (1: Forever removed Salvation, which was 0.7)")
	threatMembers := flag.Float64("threat-members", 2, "the objective: raid members close to the tank's threat, whose DPS a point of SafeDPS frees")
	dpsSlot := flag.Float64("dps-slot", 900, "the objective: a raid DPS member's DPS, what a healer slot costs")
	healerEach := flag.Float64("healer-each", 0, "the objective: one healer's HPS on the tank, 0 for the encounter's (or 400 without one)")
	encounterPath := flag.String("encounter", "", "a tank encounter file (tools/encounters/*.json, shared with the healing sim) in place of the request's encounter and the healer flags' defaults")
	healers := flag.Float64("healers", 0, "healers on the tank, in place of the encounter's (with -healer-hps and -heal-cast; -hps and -cadence win over them)")
	healerHps := flag.Float64("healer-hps", 0, "each healer's healing per second on the tank, in place of the encounter's")
	healCast := flag.Float64("heal-cast", 0, "seconds a healer's heal takes to cast, in place of the encounter's")
	bossStats := flag.Bool("boss-stats", false, "print, as JSON, the boss's attack power, flat weapon damage and attack speed under each request's debuffs, and sim nothing (tools/encounters/from_logs.py reads it to turn a log's swings into the file's)")
	flag.Parse()
	sim.RegisterAll()

	set := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { set[f.Name] = true })
	var enc *tankEncounter
	if *encounterPath != "" {
		enc = loadEncounter(*encounterPath)
	}
	// The healing: the encounter's healers, any of them changed by the healer flags, and the
	// stream's own flags over all of that.
	team := tankHealing{}
	if enc != nil && enc.Healing != nil {
		team = *enc.Healing
	}
	if set["healers"] {
		team.Healers = *healers
	}
	if set["healer-hps"] {
		team.HpsEach = *healerHps
	}
	if set["heal-cast"] {
		team.CastSeconds = *healCast
	}
	partial := team.Healers > 0 || team.HpsEach > 0 || team.CastSeconds > 0
	if partial && !(team.Healers > 0 && team.HpsEach > 0 && team.CastSeconds > 0) {
		fail(fmt.Errorf("the healing needs all of -healers, -healer-hps and -heal-cast (or an encounter's healing): got %g healers at %g HPS, %g s heals",
			team.Healers, team.HpsEach, team.CastSeconds))
	}
	if team.Healers > 0 && team.HpsEach > 0 && team.CastSeconds > 0 && !(set["hps"] && set["cadence"]) {
		if !set["hps"] {
			*hps = team.Healers * team.HpsEach
		}
		if !set["cadence"] {
			*cadence = team.CastSeconds / team.Healers
		}
		if !set["cadence-variation"] {
			*variation = *cadence / 2
		}
		fmt.Printf("Healing: %g healers on the tank at %.0f HPS each, %.1f s heals (%.0f a heal).\n",
			team.Healers, team.HpsEach, team.CastSeconds, team.HpsEach*team.CastSeconds)
	}
	if enc != nil {
		fmt.Printf("Encounter: %s. %s\n", enc.Name, enc.Notes)
	}
	fmt.Printf("Stand-in healer: %.0f HPS every %.2f ± %.2f s. TMI over %d s windows. %d iterations.\n\n",
		*hps, *cadence, *variation, *window, *iterations)
	if *classicCrit {
		fmt.Println("The boss crits at Classic's rate.")
	}
	each := *healerEach
	if each == 0 {
		each = 400
		if team.HpsEach > 0 {
			each = team.HpsEach
		}
	}
	fmt.Printf("Objective: %g x SafeDPS - w x %g x healing needed at %g%% deaths / %g, in raid DPS (Max threat w = 1/3, Balanced 1, Max survival 3).\n\n",
		*threatMembers, *dpsSlot, 100**deathTarget, each)
	header := []string{"Build", "Max threat", "Balanced", "Max survival", "SafeDPS", "Healing needed", "Death floor", "TPS", "DPS", "DTPS", "TMI", "Death", "Health", "EH", "Biggest hit", "Armor", "Def", "No crit at", "Miss", "Dodge", "Parry", "Block", "Crit", "Crush", "Crush gap", "Took: avoid", "block", "crit", "crush"}
	if *bisect {
		header = append(header, "Bisection: needed", "floor")
	}
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
		if enc != nil {
			// The fight is the file's. What it leaves unset of the request's own fight setup
			// (the execute phases, the natural flask zone) carries over.
			fight := googleProto.Clone(enc.encounter).(*proto.Encounter)
			if fight.ExecuteProportion_20 == 0 && fight.ExecuteProportion_25 == 0 && fight.ExecuteProportion_35 == 0 {
				fight.ExecuteProportion_20 = rsr.Encounter.ExecuteProportion_20
				fight.ExecuteProportion_25 = rsr.Encounter.ExecuteProportion_25
				fight.ExecuteProportion_35 = rsr.Encounter.ExecuteProportion_35
			}
			fight.NaturalFlaskZone = fight.NaturalFlaskZone || rsr.Encounter.NaturalFlaskZone
			rsr.Encounter = fight
		}
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
			if *classicCrit {
				target.ClassicCrit = true
			}
		}
		tankscore.OwnDebuffs(rsr)
		tank := rsr.Raid.Parties[0].Players[0]
		tank.TankScoring = &proto.TankScoring{DpsStart: *dpsStart, ThreatPerDamage: *dpsThreat}
		tank.HealingModel = &proto.HealingModel{
			Hps:              noPresim(*hps),
			CadenceSeconds:   *cadence,
			CadenceVariation: *variation,
			BurstWindow:      int32(*window),
		}

		if *bossStats {
			printBossStats(path, rsr)
			continue
		}
		res := run(path, rsr)
		m := res.RaidMetrics.Parties[0].Players[0]
		pct := func(x float64) string { return fmt.Sprintf("%.1f%%", 100*x) }
		s := m.TankSheet
		if s == nil {
			fail(fmt.Errorf("%s: the boss didn't attack the tank", path))
		}
		took := tookShares(res.EncounterMetrics.Targets[0], m.UnitIndex)
		objective := tankscore.Objective{ThreatMembers: *threatMembers, DpsSlot: *dpsSlot, HealerEach: each, Death: *deathTarget}
		score := func(preset string) string {
			objective.Weight = tankscore.Presets[preset]
			sc := objective.Score(m, int32(*iterations))
			if sc.Needed < 0 {
				return "none"
			}
			return fmt.Sprintf("%.0f ± %.0f", sc.Value, sc.StdErr)
		}
		sc := objective.Score(m, int32(*iterations))
		need := "none works"
		if sc.Needed >= 0 {
			need = fmt.Sprintf("%.0f ± %.0f (%.2fx DTPS)", sc.Needed, sc.NeededErr, sc.Needed/max(m.Dtps.Avg, 1))
		}
		row := []string{
			strings.TrimSuffix(filepath.Base(path), ".json"),
			score("max-threat"), score("balanced"), score("max-survival"),
			fmt.Sprintf("%.0f ± %.0f", sc.SafeDps, sc.SafeDpsErr),
			need, pct(sc.DeathFloor),
			fmt.Sprintf("%.0f", m.Threat.Avg),
			fmt.Sprintf("%.0f", m.Dps.Avg),
			fmt.Sprintf("%.0f", m.Dtps.Avg),
			fmt.Sprintf("%.0f", m.Tmi.Avg),
			pct(m.ChanceOfDeath),
			fmt.Sprintf("%.0f", s.Health),
			fmt.Sprintf("%.0f", s.EffectiveHealth),
			fmt.Sprintf("%.0f (%.1f to kill)", s.BiggestHit, s.Health/max(s.BiggestHit, 1)),
			fmt.Sprintf("%.0f (%s)", s.Armor, pct(s.ArmorReduction)),
			fmt.Sprintf("%.0f", s.Defense),
			fmt.Sprintf("%.0f", s.DefenseCritImmune),
			pct(s.Miss), pct(s.Dodge), pct(s.Parry), pct(s.Block), pct(s.Crit), pct(s.Crush), pct(s.CrushGap),
			pct(took.avoided), pct(took.blocked), pct(took.crit), pct(took.crush),
		}
		if *bisect {
			hps, floor := healingNeeded(path, rsr, m.Dtps.Avg, *deathTarget)
			need := "none works"
			if hps == 0 {
				need = "0"
			} else if hps > 0 {
				need = fmt.Sprintf("%.0f", hps)
			}
			row = append(row, need, pct(floor))
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

// noPresim keeps an HPS of 0 from asking the engine for its presim, which heals at 1.5 times the
// DTPS (GetPresimOptions, sim/core/health.go).
func noPresim(hps float64) float64 {
	return max(hps, 1e-9)
}

func run(path string, rsr *proto.RaidSimRequest) *proto.RaidSimResult {
	res := core.RunRaidSimConcurrent(rsr)
	if res.Error != nil {
		fail(fmt.Errorf("%s: %s", path, res.Error.Message))
	}
	return res
}

func deathAt(path string, rsr *proto.RaidSimRequest, hps float64) float64 {
	rsr = googleProto.Clone(rsr).(*proto.RaidSimRequest)
	rsr.Raid.Parties[0].Players[0].HealingModel.Hps = noPresim(hps)
	return run(path, rsr).RaidMetrics.Parties[0].Players[0].ChanceOfDeath
}

// healingNeeded is the stand-in healer's HPS at which the tank dies in at most target of the
// fights, to 1% of DTPS, and the death floor: the chance of death when every heal tops the tank
// off (1000 times its DTPS, since a heal is the HPS times the gap since the last one, and gaps go
// down to half the cadence), so only damage between two heals kills it. When the floor is over the
// target, no healing gets there (a one-shot or a burst faster than the heals), and the HPS is -1.
// No damage taken needs no healing (0). Every run uses the request's seed, so the chance of death
// only moves with the HPS, as long as nothing in the request acts on the tank's health (a health
// potion, a defensive at a health share, an APL condition on health).
func healingNeeded(path string, rsr *proto.RaidSimRequest, dtps, target float64) (float64, float64) {
	if dtps <= 0 {
		return 0, 0
	}
	floor := deathAt(path, rsr, 1000*dtps)
	if floor > target {
		return -1, floor
	}
	ceiling := 1000 * dtps
	lo, hi := 0.0, dtps
	for hi < ceiling && deathAt(path, rsr, hi) > target {
		lo, hi = hi, min(2*hi, ceiling)
	}
	for hi-lo > 0.01*dtps {
		mid := (lo + hi) / 2
		if deathAt(path, rsr, mid) > target {
			lo = mid
		} else {
			hi = mid
		}
	}
	return hi, floor
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

// tankEncounter is a tank encounter file: the fight, and the healing a raid puts on its tank.
type tankEncounter struct {
	Name string `json:"name"`
	// Where the numbers come from.
	Source string `json:"source"`
	Notes  string `json:"notes"`
	// A proto Encounter in protojson: the boss's melee (Target), its abilities on the tank and
	// the raid (raid_damage), tank swaps.
	Encounter json.RawMessage `json:"encounter"`
	Healing   *tankHealing    `json:"healing"`

	encounter *proto.Encounter
}

// tankHealing is the healers a raid puts on its tank: how many, each one's healing per second
// on it, and their heals' cast time, which sets how big and how far apart the heals come.
type tankHealing struct {
	Healers     float64 `json:"healers"`
	HpsEach     float64 `json:"hpsEach"`
	CastSeconds float64 `json:"castSeconds"`
	// What the numbers stand for, e.g. "2 priests on Greater Heal rank 4 and a paladin".
	Notes string `json:"notes"`
}

// printBossStats prints the first boss's attack power, flat weapon damage, damage done and attack
// speed 30 s into the first iteration, with the raid's debuffs on it (Demoralizing Shout, Curse of
// Weakness, Thunder Clap). A log's boss swings had a raid's debuffs on them, so turning them into
// an encounter file's weapon damage and swing speed takes these out.
func printBossStats(path string, rsr *proto.RaidSimRequest) {
	rsr = googleProto.Clone(rsr).(*proto.RaidSimRequest)
	rsr.SimOptions.Iterations = 1
	s := core.NewSim(rsr, simsignals.CreateSignals())
	s.Reset()
	s.PrePull()
	// Step returns true when the fight is over.
	for s.CurrentTime < 30*time.Second && !s.Step() {
	}
	t := s.Encounter.TargetUnits[0]
	out, _ := json.Marshal(map[string]any{
		"request":         path,
		"baseAttackPower": rsr.Encounter.Targets[0].Stats[stats.AttackPower],
		"attackPower":     t.GetStat(stats.AttackPower),
		"apCoefficient":   core.EnemyAutoAttackAPCoefficient,
		"flatDamage":      t.PseudoStats.BonusPhysicalDamage,
		"damageDone":      t.PseudoStats.DamageDealtMultiplier * t.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical],
		"attackSpeed":     t.PseudoStats.MeleeSpeedMultiplier,
	})
	fmt.Println(string(out))
}

func loadEncounter(path string) *tankEncounter {
	data, err := os.ReadFile(path)
	if err != nil {
		fail(err)
	}
	enc := &tankEncounter{}
	if err := json.Unmarshal(data, enc); err != nil {
		fail(fmt.Errorf("%s: %w", path, err))
	}
	enc.encounter = &proto.Encounter{}
	if err := protojson.Unmarshal(enc.Encounter, enc.encounter); err != nil {
		fail(fmt.Errorf("%s: encounter: %w", path, err))
	}
	if len(enc.encounter.Targets) == 0 {
		fail(fmt.Errorf("%s: the encounter has no boss", path))
	}
	return enc
}

// tookShares is how the boss's melee swings at the tank landed over the whole fight, with
// Shield Block, Holy Shield and the like in.
func tookShares(boss *proto.UnitMetrics, tankIndex int32) shares {
	var n, avoided, blocked, crit, crush float64
	for _, action := range boss.Actions {
		// The white swings only, not the boss's abilities.
		if action.Id.GetOtherId() != proto.OtherAction_OtherActionAttack {
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
