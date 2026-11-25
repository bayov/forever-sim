// Command healbench sims healers healing a raid that takes damage, and scores them the way
// docs/healing_sim.md settles it: the damage they carry, in a role.
//
//	go run -tags=with_db ./tools/healbench -role tank -carry tools/healbench/req/shaman_restoration.json
//	go run -tags=with_db ./tools/healbench -encounter tools/encounters/onyxia.json -role tank -carry ...
//
// Each request is a RaidSimRequest with the healer as the first player. The bench (tools/healscore)
// builds a raid around it: the tanks from -tank (the tank bench's requests), -healers copies of the
// healer, and -fillers copies of the first player of -filler. It turns on the healing fight
// (Encounter.track_health), so every player loses health and can die, and nobody has the stand-in
// healing model.
//
// The damage is a made-up -profile or a boss from tools/encounters (-encounter), judged at the
// boss's own fight length (the user's ruling): for -role raid and both the logs' median kill
// (raidHealing.fightSeconds), for tank the file's duration, the main tank's time on the boss. -role picks what the healers face and what fails:
//
//   - tank: the boss's melee and its abilities on the tanks, nothing on the raid; the healers are
//     assigned the tanks; a fight fails when a tank dies. The line is 2% of fights.
//   - raid: only the damage on the raid (no boss melee, no tanks); a fight fails when anyone dies.
//     The line is 5%.
//   - both: everything; a tank death over 2% or another death over 5% fails.
//
// With -carry the bench finds the damage carried: the scale of all the damage at which the deaths
// reach their line, by bisection on paired seeds and interpolation between the two closest scales.
// It reports the death floor next to it: the deaths at that scale when every player is topped off
// every 1.5 s or so, which no healer can stop (a one-shot, a burst between two heals).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/tools/healscore"
)

func main() {
	tankPath := flag.String("tank", healscore.DefaultTank, "the tank's request")
	fillerPath := flag.String("filler", healscore.DefaultFiller, "the request whose first player fills the raid")
	profileName := flag.String("profile", "raid", "a made-up damage profile: "+strings.Join(profileNames(), ", "))
	encounterPath := flag.String("encounter", "", "a boss from tools/encounters, in place of -profile")
	roleName := flag.String("role", "both", "tank, raid or both")
	scale := flag.Float64("scale", 1, "scales all of the damage, the boss's melee too")
	healers := flag.Int("healers", 0, "copies of the healer (default 1 for -role tank, 2 otherwise)")
	fillers := flag.Int("fillers", 7, "copies of the filler")
	iterations := flag.Int("iters", 1000, "iterations a sim")
	duration := flag.Float64("dur", 0, "fight length in seconds (default the encounter's: the logs' median kill for -role raid or both, the main tank's time for tank; 180 for a profile)")
	carry := flag.Bool("carry", false, "also find the damage carried, and the death floor there")
	floorOnly := flag.Bool("floor", false, "print the death floor at -scale instead")
	logOne := flag.Bool("log", false, "print one iteration's log of the first build instead")
	stopCast := flag.Float64("stopcast", 0.5, "healers cancel a heal that would overheal by more than this share (0 never)")
	reaction := flag.Float64("reaction", 0.3, "the healers' average reaction time in seconds, spread ±25% between them (0 sees at once)")
	seed := flag.Int64("seed", healscore.DefaultSeed, "the first fight's random seed (fight i is the fight a one-fight run with seed + i sims)")
	labeled := flag.Bool("labeled", false, "give each kind of random roll its own stream (SimOptions.useLabeledRands), so two healers' fights stay alike past where they differ")
	scales := flag.String("scales", "", "with -vectors: the damage scales to sim each healer at, comma separated, in place of the table")
	vectors := flag.String("vectors", "", "append each healer's per-fight values at each of -scales to this file, one JSON line a scale: the deaths (1 a tank, 2 another, 3 both), HPS, useful HPS and mana left")
	flag.Parse()
	sim.RegisterAll()

	o := healscore.Options{StopCast: *stopCast, Reaction: *reaction, Role: *roleName, Scale: *scale, Healers: *healers, Fillers: *fillers, Iterations: *iterations, Duration: *duration,
		Seed: *seed, Labeled: *labeled, AllValues: *vectors != ""}
	if o.Role != "tank" && o.Role != "raid" && o.Role != "both" {
		fail(fmt.Errorf("-role is tank, raid or both"))
	}
	if o.Healers == 0 {
		o.Healers = healscore.DefaultHealers(o.Role)
	}

	var fight healscore.Fight
	if *encounterPath != "" {
		var err error
		if fight, err = healscore.LoadFight(*encounterPath); err != nil {
			fail(err)
		}
	} else {
		p, ok := profiles[*profileName]
		if !ok {
			fail(fmt.Errorf("no profile %q (%s)", *profileName, strings.Join(profileNames(), ", ")))
		}
		fight = p.fight()
	}
	if o.Duration == 0 {
		o.Duration = fight.Duration(o.Role)
	}
	tank := healscore.FirstPlayer(read(*tankPath))
	filler := healscore.FirstPlayer(read(*fillerPath))

	fmt.Printf("%s, role %s ×%.2f, %.0f s, %d iterations a sim. Healers react in %.2f s ± 25%% and cancel heals over %.0f%% overheal.\n%s\n\n",
		fight.Name, o.Role, o.Scale, o.Duration, o.Iterations, o.Reaction, 100*o.StopCast, fight.Description)
	header := []string{"Healer", "HPS", "EHPS", "Overheal", "EH per mana", "Mana left", "OOM at", "Tank DTPS", "Raid DTPS", "Tank died", "Other died", "First death", "Tank scares", "Raid scares"}
	if *carry {
		header = append(header, "Carry", "Death floor")
	}
	fmt.Println("| " + strings.Join(header, " | ") + " |")
	fmt.Println("|" + strings.Repeat("---|", len(header)))

	for _, path := range flag.Args() {
		rsr := read(path)
		b := &healscore.Bench{O: o, Fight: fight, Tank: tank, Healer: healscore.FirstPlayer(rsr), Filler: filler, Buffs: rsr.Raid.Buffs, Debuffs: rsr.Raid.Debuffs}
		if *vectors != "" {
			for _, part := range strings.Split(*scales, ",") {
				k, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
				if err != nil {
					fail(fmt.Errorf("-scales: %w", err))
				}
				writeVectors(b, *vectors, path, k)
			}
			continue
		}
		if *logOne {
			req, err := b.Request(o.Scale, false)
			if err != nil {
				fail(err)
			}
			req.SimOptions.Iterations = 1
			req.SimOptions.Debug = true
			res := core.RunRaidSim(req)
			if res.Error != nil {
				fail(fmt.Errorf("%s: %s", path, res.Error.Message))
			}
			fmt.Println(res.Logs)
			return
		}
		if *floorOnly {
			f := run(b, o.Scale, true)
			fmt.Printf("%s: death floor at ×%.2f: %s tank, %s other\n", path, o.Scale, pct(f.TankDied), pct(f.OtherDied))
			continue
		}
		r := run(b, o.Scale, false)
		row := append([]string{strings.TrimSuffix(filepath.Base(path), ".json")}, cells(r)...)
		if *carry {
			c, err := b.Carry()
			if err != nil {
				fail(err)
			}
			floor := run(b, c.K, true)
			row = append(row, fmt.Sprintf("%.3f", c.K), fmt.Sprintf("%s tank, %s other", pct(floor.TankDied), pct(floor.OtherDied)))
		}
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}

// writeVectors appends a healer's per-fight values at a scale of the damage, in seed order, as one
// JSON line: what tools/dps_report/tankheal.py pairs fight by fight between races and racials.
func writeVectors(b *healscore.Bench, file, path string, scale float64) {
	req, err := b.Request(scale, false)
	if err != nil {
		fail(err)
	}
	res := core.RunRaidSimConcurrent(req)
	if res.Error != nil {
		fail(fmt.Errorf("%s", res.Error.Message))
	}
	hf := res.RaidMetrics.HealingFight
	if hf == nil || len(hf.FightDeaths) != b.O.Iterations {
		fail(fmt.Errorf("%s: no per-fight deaths", path))
	}
	hps := make([]float64, b.O.Iterations)
	ehps := make([]float64, b.O.Iterations)
	manaEnd := make([]float64, b.O.Iterations)
	for pi, party := range res.RaidMetrics.Parties {
		for si, m := range party.Players {
			if b.Slots[pi][si] != healscore.RoleHealer {
				continue
			}
			for i := range hps {
				hps[i] += m.Hps.AllValues[i]
				ehps[i] += m.Ehps.AllValues[i]
				manaEnd[i] += m.ManaEnd.AllValues[i] / float64(b.O.Healers)
			}
		}
	}
	line, err := json.Marshal(map[string]any{"build": strings.TrimSuffix(filepath.Base(path), ".json"), "scale": scale,
		"deaths": hf.FightDeaths, "hps": hps, "ehps": ehps, "manaEnd": manaEnd})
	if err != nil {
		fail(err)
	}
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fail(err)
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		fail(err)
	}
}

// run sims the bench, or fails.
func run(b *healscore.Bench, scale float64, floor bool) healscore.Result {
	r, err := b.Run(scale, floor)
	if err != nil {
		fail(err)
	}
	return r
}

func pct(x float64) string { return fmt.Sprintf("%.1f%%", 100*x) }

func cells(r healscore.Result) []string {
	overheal, perMana := 0.0, 0.0
	if r.Healing > 0 {
		overheal = r.Overheal / r.Healing
	}
	if r.ManaSpent > 0 {
		perMana = r.Ehps * r.Duration / r.ManaSpent
	}
	oom := fmt.Sprintf("%.0f s", r.Tto)
	if r.Tto > r.Duration {
		oom = fmt.Sprintf("(%.0f s)", r.Tto)
	}
	first := "-"
	if r.AnyDied > 0 {
		first = fmt.Sprintf("%.0f s", r.FirstDeath)
	}
	return []string{
		fmt.Sprintf("%.0f", r.Hps),
		fmt.Sprintf("%.0f", r.Ehps),
		pct(overheal),
		fmt.Sprintf("%.2f", perMana),
		fmt.Sprintf("%.0f", r.ManaEnd),
		oom,
		fmt.Sprintf("%.0f", r.TankDtps),
		fmt.Sprintf("%.0f", r.RaidDtps),
		pct(r.TankDied),
		pct(r.OtherDied),
		first,
		fmt.Sprintf("%.2f", r.TankScares),
		fmt.Sprintf("%.2f", r.RaidScares),
	}
}

// profile is a made-up fight to test the engine with, before the bosses have raid damage.
type profile struct {
	description string
	tanks       int
	bossSwing   float64 // the boss's smallest white hit before armor, every 2 s
	tankSwap    float64 // seconds between tank swaps, 0 for none
	events      []*proto.RaidDamageEvent
}

func (p profile) fight() healscore.Fight {
	return healscore.Fight{
		Name:        "A made-up profile",
		Description: p.description,
		Tanks:       p.tanks,
		Encounter: &proto.Encounter{
			Duration:         180,
			NaturalFlaskZone: true,
			Targets: []*proto.Target{{
				Name: "Level 60", Level: 63, MinBaseDamage: p.bossSwing, SwingSpeed: 2, DamageSpread: 0.3333,
				ParryHaste: true, TankSwapSeconds: p.tankSwap,
				Stats: func() []float64 { s := make([]float64, 44); s[17], s[26] = 805, 3731; return s }(),
			}},
			RaidDamage: p.events,
		},
	}
}

var profiles = map[string]profile{
	"tank": {
		description: "the boss's melee on one tank, nothing on the raid",
		tanks:       1,
		bossSwing:   3000,
	},
	"raid": {
		description: "the boss's melee on one tank; a Fire pulse on everyone every 8 ± 2 s; a Shadow bolt on 3 random players other than the tanks every 6 s on average",
		tanks:       1,
		bossSwing:   3000,
		events: []*proto.RaidDamageEvent{
			{School: proto.SpellSchool_SpellSchoolFire, MinDamage: 500, MaxDamage: 700, EverySeconds: 8, SpreadSeconds: 2, FirstAt: 5},
			{School: proto.SpellSchool_SpellSchoolShadow, MinDamage: 1200, MaxDamage: 1600, Who: proto.RaidDamageEvent_NotTanks, Count: 3, EverySeconds: 6, RandomTiming: true, FirstAt: 3},
		},
	},
	"swap": {
		description: "two tanks trading the boss every 20 s; a Physical bleed on the boss's target ticking 400 every 2 s for 10 s, every 10 s (refreshed, not stacked)",
		tanks:       2,
		bossSwing:   3000,
		tankSwap:    20,
		events: []*proto.RaidDamageEvent{
			{School: proto.SpellSchool_SpellSchoolPhysical, MinDamage: 400, MaxDamage: 400, Who: proto.RaidDamageEvent_BossTarget, EverySeconds: 10, FirstAt: 2, Ticks: 5, TickSeconds: 2},
		},
	},
}

func profileNames() []string {
	return []string{"tank", "raid", "swap"}
}

func read(path string) *proto.RaidSimRequest {
	rsr, err := healscore.ReadRequest(path)
	if err != nil {
		fail(err)
	}
	return rsr
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
