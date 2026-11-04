// Command healbench sims healers healing a raid that takes damage, and scores them the way
// docs/healing_sim.md settles it: the damage they carry, in a role.
//
//	go run -tags=with_db ./tools/healbench -role tank -carry tools/healbench/req/shaman_restoration.json
//	go run -tags=with_db ./tools/healbench -encounter tools/encounters/onyxia.json -role tank -carry ...
//
// Each request is a RaidSimRequest with the healer as the first player. The bench builds a raid
// around it: the tanks from -tank (the tank bench's requests), -healers copies of the healer, and
// -fillers copies of the first player of -filler. It turns on the healing fight
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
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/tools/tankscore"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

// The death lines (the user's ruling, 2026-11-02).
const (
	tankLine  = 0.02
	otherLine = 0.05
)

type options struct {
	stopCast   float64
	reaction   float64
	role       string
	scale      float64
	healers    int
	fillers    int
	iterations int
	duration   float64
}

func main() {
	tankPath := flag.String("tank", "tools/tankbench/req/warrior_protection.json", "the tank's request")
	fillerPath := flag.String("filler", "tools/dps_report/req/mage__frost.json", "the request whose first player fills the raid")
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
	flag.Parse()
	sim.RegisterAll()

	o := options{stopCast: *stopCast, reaction: *reaction, role: *roleName, scale: *scale, healers: *healers, fillers: *fillers, iterations: *iterations, duration: *duration}
	if o.role != "tank" && o.role != "raid" && o.role != "both" {
		fail(fmt.Errorf("-role is tank, raid or both"))
	}
	if o.healers == 0 {
		o.healers = 2
		if o.role == "tank" {
			o.healers = 1
		}
	}

	var fight fight
	if *encounterPath != "" {
		enc, err := tankscore.LoadEncounter(*encounterPath)
		if err != nil {
			fail(err)
		}
		fight = encounterFight(enc)
		fight.killSeconds = killSeconds(*encounterPath)
	} else {
		p, ok := profiles[*profileName]
		if !ok {
			fail(fmt.Errorf("no profile %q (%s)", *profileName, strings.Join(profileNames(), ", ")))
		}
		fight = p.fight()
	}
	if o.duration == 0 {
		o.duration = fight.encounter.Duration
		// The raid's damage is the whole kill's (the logs' median kill), where the file's own length is
		// the main tank's time on the boss (Onyxia's first phase, one tank's stint on Vaelastrasz).
		if o.role != "tank" && fight.killSeconds > 0 {
			o.duration = fight.killSeconds
		}
	}
	tank := firstPlayer(read(*tankPath))
	filler := firstPlayer(read(*fillerPath))

	fmt.Printf("%s, role %s ×%.2f, %.0f s, %d iterations a sim. Healers react in %.2f s ± 25%% and cancel heals over %.0f%% overheal.\n%s\n\n",
		fight.name, o.role, o.scale, o.duration, o.iterations, o.reaction, 100*o.stopCast, fight.description)
	header := []string{"Healer", "HPS", "EHPS", "Overheal", "EH per mana", "Mana left", "OOM at", "Tank DTPS", "Raid DTPS", "Tank died", "Other died", "First death", "Tank scares", "Raid scares"}
	if *carry {
		header = append(header, "Carry", "Death floor")
	}
	fmt.Println("| " + strings.Join(header, " | ") + " |")
	fmt.Println("|" + strings.Repeat("---|", len(header)))

	for _, path := range flag.Args() {
		rsr := read(path)
		b := bench{o: o, fight: fight, tank: tank, healer: firstPlayer(rsr), filler: filler, buffs: rsr.Raid.Buffs, debuffs: rsr.Raid.Debuffs}
		if *logOne {
			req := b.request(o.scale, false)
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
			f := b.run(o.scale, true)
			fmt.Printf("%s: death floor at ×%.2f: %s tank, %s other\n", path, o.scale, pct(f.tankDied), pct(f.otherDied))
			continue
		}
		r := b.run(o.scale, false)
		row := append([]string{strings.TrimSuffix(filepath.Base(path), ".json")}, r.cells(o)...)
		if *carry {
			k := b.carry()
			floor := b.run(k, true)
			row = append(row, fmt.Sprintf("%.3f", k), fmt.Sprintf("%s tank, %s other", pct(floor.tankDied), pct(floor.otherDied)))
		}
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}

// fight is the damage: the boss with its melee, and the events on the raid.
type fight struct {
	name, description string
	tanks             int
	encounter         *proto.Encounter
	// The median kill in the logs (the file's raidHealing.fightSeconds), 0 when it has none.
	killSeconds float64
}

// killSeconds is an encounter file's median kill length from the logs (raidHealing.fightSeconds,
// the 10th, 50th and 90th percentile), or 0.
func killSeconds(path string) float64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var f struct {
		RaidHealing struct {
			FightSeconds []float64 `json:"fightSeconds"`
		} `json:"raidHealing"`
	}
	if json.Unmarshal(data, &f) != nil || len(f.RaidHealing.FightSeconds) < 2 {
		return 0
	}
	return f.RaidHealing.FightSeconds[1]
}

func encounterFight(enc *tankscore.Encounter) fight {
	f := fight{name: enc.Name, description: enc.Notes, encounter: enc.Encounter, tanks: 1}
	for _, t := range enc.Encounter.Targets {
		if t.TankSwapSeconds > 0 {
			f.tanks = 2
		}
	}
	return f
}

// onTanks is whether an event only ever hits tanks.
func onTanks(e *proto.RaidDamageEvent) bool {
	return e.Who == proto.RaidDamageEvent_Tanks || e.Who == proto.RaidDamageEvent_BossTarget
}

type role int

const (
	roleTank role = iota
	roleHealer
	roleFiller
)

type bench struct {
	o                      options
	fight                  fight
	tank, healer, filler   *proto.Player
	buffs                  *proto.RaidBuffs
	debuffs                *proto.Debuffs
	slots                  [][]role
	tankIndexes, healIndex []int32
}

// request builds the raid and the fight at a scale of the damage. With floor, every player is
// topped off by the stand-in heal stream instead of the healers healing.
func (b *bench) request(scale float64, floor bool) *proto.RaidSimRequest {
	tanks := b.fight.tanks
	fillers := b.o.fillers
	if b.o.role == "raid" {
		tanks = 0
	}
	// The tank role keeps the fillers, idle (below): a raid around the tank, so a heal that jumps
	// (Prayer of Mending) has the raid to land on, not only the tank and the healer
	// (healer_fairness-20261104-1). They take no damage there, since the role drops the raid's.

	// The tanks go in the first party, the healers one to a party in turn, and the fillers in the
	// slots left, so Chain Heal and the party heals see a raid's spread.
	total := tanks + b.o.healers + fillers
	if total > 40 {
		fail(fmt.Errorf("%d players, 40 at most", total))
	}
	numParties := (total + 4) / 5
	b.slots = make([][]role, numParties)
	place := func(r role, party int) {
		for len(b.slots[party]) == 5 {
			party = (party + 1) % numParties
		}
		b.slots[party] = append(b.slots[party], r)
	}
	for i := 0; i < tanks; i++ {
		place(roleTank, 0)
	}
	for i := 0; i < b.o.healers; i++ {
		place(roleHealer, i%numParties)
	}
	for i := 0; i < fillers; i++ {
		place(roleFiller, 0)
	}

	raid := &proto.Raid{Buffs: b.buffs, Debuffs: b.debuffs}
	b.tankIndexes = nil
	for pi, party := range b.slots {
		raid.Parties = append(raid.Parties, &proto.Party{})
		for si, r := range party {
			if r == roleTank {
				idx := int32(pi*5 + si)
				b.tankIndexes = append(b.tankIndexes, idx)
				raid.Tanks = append(raid.Tanks, &proto.UnitReference{Type: proto.UnitReference_Player, Index: idx})
			}
		}
	}
	healerNumber := 0
	for pi, party := range b.slots {
		for si, r := range party {
			player := clone(map[role]*proto.Player{roleTank: b.tank, roleHealer: b.healer, roleFiller: b.filler}[r])
			if r == roleTank {
				// A tank faces the boss: the boss parries and blocks it, and its parries haste the swing.
				player.InFrontOfTarget = true
			}
			if r == roleHealer {
				// Each healer reacts a little differently, so identical healers don't act in lockstep.
				spread := 0.0
				if b.o.healers > 1 {
					spread = -0.25 + 0.5*float64(healerNumber)/float64(b.o.healers-1)
				}
				player.ReactionTimeMs = int32(1000 * b.o.reaction * (1 + spread))
				player.StopCastOverheal = b.o.stopCast
				healerNumber++
			}
			player.HealingModel = nil
			if r == roleFiller && b.o.role == "tank" {
				player.Rotation = &proto.APLRotation{Type: proto.APLRotation_TypeAPL}
			}
			player.Name = fmt.Sprintf("Player %d", pi*5+si+1)
			if r == roleHealer && b.o.role == "tank" {
				for _, idx := range b.tankIndexes {
					player.HealAssignment = append(player.HealAssignment, &proto.UnitReference{Type: proto.UnitReference_Player, Index: idx})
				}
			}
			if floor {
				// A full heal every 1.5 ± 0.5 s: only a one-shot or a burst between heals kills.
				player.HealingModel = &proto.HealingModel{Hps: 50000, CadenceSeconds: 1.5, CadenceVariation: 0.5}
				if r == roleHealer {
					player.Rotation = &proto.APLRotation{Type: proto.APLRotation_TypeAPL}
				}
			}
			raid.Parties[pi].Players = append(raid.Parties[pi].Players, player)
		}
	}
	raid.NumActiveParties = int32(len(raid.Parties))

	encounter := googleProto.Clone(b.fight.encounter).(*proto.Encounter)
	encounter.Duration = b.o.duration
	encounter.DurationVariation = 0
	encounter.TrackHealth = true
	encounter.RaidDamageHitsPerMinute = 0
	encounter.RaidDamage = nil
	for _, e := range b.fight.encounter.RaidDamage {
		if (b.o.role == "tank" && !onTanks(e)) || (b.o.role == "raid" && onTanks(e)) {
			continue
		}
		e = googleProto.Clone(e).(*proto.RaidDamageEvent)
		e.MinDamage *= scale
		e.MaxDamage *= scale
		encounter.RaidDamage = append(encounter.RaidDamage, e)
	}
	for _, target := range encounter.Targets {
		target.MinBaseDamage *= scale
		if b.o.role == "raid" {
			target.TankIndex = -1
			target.TankSwapSeconds = 0
		}
	}

	return &proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{Iterations: int32(b.o.iterations), RandomSeed: 101},
	}
}

// result is what a sim tells about the healers and the raid.
type result struct {
	hps, ehps, overheal, healing, manaSpent, manaEnd, tto float64
	tankDtps, raidDtps, tankScares, raidScares            float64
	tankDied, otherDied, anyDied, firstDeath              float64
	duration                                              float64
}

func (b *bench) run(scale float64, floor bool) result {
	req := b.request(scale, floor)
	res := core.RunRaidSimConcurrent(req)
	if res.Error != nil {
		fail(fmt.Errorf("%s", res.Error.Message))
	}
	r := result{duration: res.AvgIterationDuration}
	healers := float64(b.o.healers)
	for pi, party := range res.RaidMetrics.Parties {
		for si, m := range party.Players {
			who := b.slots[pi][si]
			r.raidDtps += m.Dtps.Avg
			r.raidScares += m.ScaresAvg
			switch who {
			case roleTank:
				r.tankDtps += m.Dtps.Avg
				r.tankScares += m.ScaresAvg
			case roleHealer:
				r.hps += m.Hps.Avg
				r.ehps += m.Ehps.Avg
				r.manaSpent += m.ManaSpentAvg
				r.manaEnd += m.ManaEnd.Avg / healers
				r.tto += m.Tto.Avg / healers
				for _, a := range m.Actions {
					for _, t := range a.Targets {
						r.healing += t.Healing
						r.overheal += t.Overhealing
					}
				}
			}
		}
	}
	if hf := res.RaidMetrics.HealingFight; hf != nil && hf.Fights > 0 {
		n := float64(hf.Fights)
		r.tankDied = float64(hf.FightsTankDeath) / n
		r.otherDied = float64(hf.FightsOtherDeath) / n
		r.anyDied = float64(hf.FightsAnyDeath) / n
		if hf.FightsAnyDeath > 0 {
			r.firstDeath = hf.FirstDeathSecondsSum / float64(hf.FightsAnyDeath)
		}
	}
	return r
}

// failure is how far over its lines a sim's deaths are: 1 is on the line.
func (b *bench) failure(r result) float64 {
	switch b.o.role {
	case "tank":
		return r.tankDied / tankLine
	case "raid":
		return r.anyDied / otherLine
	default:
		return max(r.tankDied/tankLine, r.otherDied/otherLine)
	}
}

// carry is the damage scale at which the deaths reach their line: bracketed by doubling or halving,
// narrowed by bisection, then interpolated between the two closest scales.
func (b *bench) carry() float64 {
	f := func(k float64) float64 { return b.failure(b.run(k, false)) }
	lo, hi := 1.0, 1.0
	flo, fhi := f(1), 0.0
	if flo >= 1 {
		hi, fhi = lo, flo
		for flo >= 1 && lo > 1.0/64 {
			lo /= 2
			flo = f(lo)
		}
	} else {
		hi = 2
		fhi = f(hi)
		for fhi < 1 && hi < 64 {
			lo, flo = hi, fhi
			hi *= 2
			fhi = f(hi)
		}
	}
	for i := 0; i < 5; i++ {
		mid := (lo + hi) / 2
		if fm := f(mid); fm < 1 {
			lo, flo = mid, fm
		} else {
			hi, fhi = mid, fm
		}
	}
	if fhi <= flo || math.IsInf(fhi, 0) {
		return lo
	}
	return lo + (hi-lo)*(1-flo)/(fhi-flo)
}

func pct(x float64) string { return fmt.Sprintf("%.1f%%", 100*x) }

func (r result) cells(o options) []string {
	overheal, perMana := 0.0, 0.0
	if r.healing > 0 {
		overheal = r.overheal / r.healing
	}
	if r.manaSpent > 0 {
		perMana = r.ehps * r.duration / r.manaSpent
	}
	oom := fmt.Sprintf("%.0f s", r.tto)
	if r.tto > r.duration {
		oom = fmt.Sprintf("(%.0f s)", r.tto)
	}
	first := "-"
	if r.anyDied > 0 {
		first = fmt.Sprintf("%.0f s", r.firstDeath)
	}
	return []string{
		fmt.Sprintf("%.0f", r.hps),
		fmt.Sprintf("%.0f", r.ehps),
		pct(overheal),
		fmt.Sprintf("%.2f", perMana),
		fmt.Sprintf("%.0f", r.manaEnd),
		oom,
		fmt.Sprintf("%.0f", r.tankDtps),
		fmt.Sprintf("%.0f", r.raidDtps),
		pct(r.tankDied),
		pct(r.otherDied),
		first,
		fmt.Sprintf("%.2f", r.tankScares),
		fmt.Sprintf("%.2f", r.raidScares),
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

func (p profile) fight() fight {
	return fight{
		name:        "A made-up profile",
		description: p.description,
		tanks:       p.tanks,
		encounter: &proto.Encounter{
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
	data, err := os.ReadFile(path)
	if err != nil {
		fail(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		fail(fmt.Errorf("%s: %w", path, err))
	}
	return rsr
}

func firstPlayer(rsr *proto.RaidSimRequest) *proto.Player {
	return rsr.Raid.Parties[0].Players[0]
}

func clone(p *proto.Player) *proto.Player {
	return googleProto.Clone(p).(*proto.Player)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
