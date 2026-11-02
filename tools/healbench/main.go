// Command healbench sims healers healing a raid that takes damage, and prints what a healer
// build is scored on (docs/healing_sim.md): effective healing, overhealing, mana, deaths and
// scares.
//
//	go run -tags=with_db ./tools/healbench -profile raid tools/healbench/req/shaman_restoration.json
//
// Each request is a RaidSimRequest with the healer as the first player. The bench builds a raid
// around it: the tanks from -tank (the tank bench's requests), -healers copies of the healer, and
// -fillers copies of the first player of -filler, who take the raid damage. It turns on the
// healing fight (Encounter.track_health), so every player loses health and can die, and nobody
// has the stand-in healing model.
//
// Time to OOM isn't shown: a healer with no heal it can pay for finds nothing to cast instead of
// waiting for mana, so the engine never marks it OOM, and the estimate from the mana left at the
// end reads wrong once the raid dies. Mana left is the mana measure until that's fixed.
//
// The profiles are made up to test the engine (H1). Sourced ones from Classic and Season of
// Discovery bosses come with H4.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

func main() {
	tankPath := flag.String("tank", "tools/tankbench/req/warrior_protection.json", "the tank's request")
	fillerPath := flag.String("filler", "tools/dps_report/req/mage__frost.json", "the request whose first player fills the raid")
	profile := flag.String("profile", "raid", "the damage: "+strings.Join(profileNames(), ", "))
	scale := flag.Float64("scale", 1, "scales all of the profile's damage, the boss's melee too")
	healers := flag.Int("healers", 2, "copies of the healer")
	fillers := flag.Int("fillers", 7, "copies of the filler")
	iterations := flag.Int("iters", 1000, "iterations per build")
	duration := flag.Float64("dur", 180, "fight length in seconds")
	logOne := flag.Bool("log", false, "print one iteration's log of the first build instead")
	flag.Parse()
	sim.RegisterAll()

	p, ok := profiles[*profile]
	if !ok {
		fail(fmt.Errorf("no profile %q (%s)", *profile, strings.Join(profileNames(), ", ")))
	}
	tank := firstPlayer(read(*tankPath))
	filler := firstPlayer(read(*fillerPath))

	fmt.Printf("Profile %s ×%.2f: %s\n%d tank(s), %d healers, %d fillers, %.0f s, %d iterations.\n\n",
		*profile, *scale, p.description, p.tanks, *healers, *fillers, *duration, *iterations)
	header := []string{"Healer", "HPS", "EHPS", "Overheal", "Mana left", "Tank DTPS", "Raid DTPS", "Tank deaths", "Most deaths (one player)", "Tank scares", "Raid scares"}
	fmt.Println("| " + strings.Join(header, " | ") + " |")
	fmt.Println("|" + strings.Repeat("---|", len(header)))

	for _, path := range flag.Args() {
		rsr := read(path)
		healer := firstPlayer(rsr)

		// The tanks go in the first party, the healers one to a party in turn, and the fillers in the
		// slots left, so Chain Heal and the party heals see a raid's spread.
		total := p.tanks + *healers + *fillers
		if total > 40 {
			fail(fmt.Errorf("%d players, 40 at most", total))
		}
		numParties := (total + 4) / 5
		slots := make([][]role, numParties)
		place := func(r role, party int) {
			for len(slots[party]) == 5 {
				party = (party + 1) % numParties
			}
			slots[party] = append(slots[party], r)
		}
		for i := 0; i < p.tanks; i++ {
			place(roleTank, 0)
		}
		for i := 0; i < *healers; i++ {
			place(roleHealer, i%numParties)
		}
		for i := 0; i < *fillers; i++ {
			place(roleFiller, 0)
		}
		raid := &proto.Raid{Buffs: rsr.Raid.Buffs, Debuffs: rsr.Raid.Debuffs}
		for pi, party := range slots {
			raid.Parties = append(raid.Parties, &proto.Party{})
			for si, r := range party {
				player := clone(map[role]*proto.Player{roleTank: tank, roleHealer: healer, roleFiller: filler}[r])
				player.HealingModel = nil
				player.Name = fmt.Sprintf("Player %d", pi*5+si+1)
				raid.Parties[pi].Players = append(raid.Parties[pi].Players, player)
				if r == roleTank {
					raid.Tanks = append(raid.Tanks, &proto.UnitReference{Type: proto.UnitReference_Player, Index: int32(pi*5 + si)})
				}
			}
		}
		raid.NumActiveParties = int32(len(raid.Parties))

		encounter := googleProto.Clone(rsr.Encounter).(*proto.Encounter)
		encounter.Duration = *duration
		encounter.DurationVariation = 0
		encounter.TrackHealth = true
		encounter.RaidDamageHitsPerMinute = 0
		encounter.RaidDamage = nil
		for _, e := range p.events {
			e = googleProto.Clone(e).(*proto.RaidDamageEvent)
			e.MinDamage *= *scale
			e.MaxDamage *= *scale
			encounter.RaidDamage = append(encounter.RaidDamage, e)
		}
		for _, target := range encounter.Targets {
			target.TankIndex = 0
			target.MinBaseDamage = p.bossSwing * *scale
			target.SwingSpeed = 2
			target.TankSwapSeconds = p.tankSwap
		}

		req := &proto.RaidSimRequest{
			Raid:       raid,
			Encounter:  encounter,
			SimOptions: &proto.SimOptions{Iterations: int32(*iterations), RandomSeed: 101},
		}
		if *logOne {
			req.SimOptions.Iterations = 1
			req.SimOptions.Debug = true
			res := core.RunRaidSim(req)
			if res.Error != nil {
				fail(fmt.Errorf("%s: %s", path, res.Error.Message))
			}
			fmt.Println(res.Logs)
			return
		}
		res := core.RunRaidSimConcurrent(req)
		if res.Error != nil {
			fail(fmt.Errorf("%s: %s", path, res.Error.Message))
		}

		var hps, ehps, overheal, healing, manaEnd, tankDtps, raidDtps, tankDeaths, mostDeaths, tankScares, raidScares float64
		for pi, party := range res.RaidMetrics.Parties {
			for si, m := range party.Players {
				r := slots[pi][si]
				raidDtps += m.Dtps.Avg
				raidScares += m.ScaresAvg
				mostDeaths = max(mostDeaths, m.ChanceOfDeath)
				if r == roleTank {
					tankDtps += m.Dtps.Avg
					tankDeaths = max(tankDeaths, m.ChanceOfDeath)
					tankScares += m.ScaresAvg
				}
				if r == roleHealer {
					hps += m.Hps.Avg
					ehps += m.Ehps.Avg
					manaEnd += m.ManaEnd.Avg / float64(*healers)
					for _, a := range m.Actions {
						for _, t := range a.Targets {
							healing += t.Healing
							overheal += t.Overhealing
						}
					}
				}
			}
		}
		pct := func(x float64) string { return fmt.Sprintf("%.1f%%", 100*x) }
		overhealShare := 0.0
		if healing > 0 {
			overhealShare = overheal / healing
		}
		row := []string{
			strings.TrimSuffix(filepath.Base(path), ".json"),
			fmt.Sprintf("%.0f", hps),
			fmt.Sprintf("%.0f", ehps),
			pct(overhealShare),
			fmt.Sprintf("%.0f", manaEnd),
			fmt.Sprintf("%.0f", tankDtps),
			fmt.Sprintf("%.0f", raidDtps),
			pct(tankDeaths),
			pct(mostDeaths),
			fmt.Sprintf("%.2f", tankScares),
			fmt.Sprintf("%.2f", raidScares),
		}
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}

type role int

const (
	roleTank role = iota
	roleHealer
	roleFiller
)

type profile struct {
	description string
	tanks       int
	bossSwing   float64 // the boss's smallest white hit before armor, every 2 s
	tankSwap    float64 // seconds between tank swaps, 0 for none
	events      []*proto.RaidDamageEvent
}

var profiles = map[string]profile{
	"tank": {
		description: "the boss's melee on one tank, nothing on the raid",
		tanks:       1,
		bossSwing:   3000,
	},
	"raid": {
		description: "the boss's melee on one tank; a Fire pulse on everyone every 8 ± 2 s; a Shadow bolt on 3 random players every 6 s on average",
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
