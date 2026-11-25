package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/tools/healscore"
	goproto "google.golang.org/protobuf/proto"
)

// The healer objective (-heal-roles, docs/healing_sim.md, The rotopt objective). A set is scored by
// the damage it carries in tools/healscore's bench, the score the Tanking and Healing report
// publishes, on one slot per role and boss. The score is 100 times the mean carry over the slots,
// in percent of the boss's damage.
//
// A full carry is about 10 sims, too slow for a gear search. So the first set scored (the starting
// set) gets the full carry k on each slot and the failure's slope s there (healscore.Carry.Slope),
// and every set after is one sim at k: carry ≈ k + (1 - f)/s, with f the deaths over their line.
// When f is outside healTrust or no fight failed, the line is too far for the slope, and the run
// brackets it from k instead (healFallback).

var (
	healRoles        = flag.String("heal-roles", "", "score the player as a healer by the damage carried (docs/healing_sim.md): tank, raid or tank,raid, one slot per role and boss, the score 100 x their mean carry")
	healEncounter    = flag.String("heal-encounter", "tools/encounters/magmadar.json,tools/encounters/golemagg.json,tools/encounters/onyxia.json,tools/encounters/broodlord.json,tools/encounters/vaelastrasz.json,tools/encounters/chromaggus.json,tools/encounters/ragnaros.json", "with -heal-roles: the boss files, comma separated")
	healTank         = flag.String("heal-tank", healscore.DefaultTank, "with -heal-roles: the tank's request (healbench -tank)")
	healFiller       = flag.String("heal-filler", healscore.DefaultFiller, "with -heal-roles: the request whose first player fills the raid (healbench -filler)")
	healRaidSettings = flag.String("heal-raid-settings", "", "with -heal-roles raid: the raid preset's request (tools/healbench/req/*_raid.json), whose first player the raid slots sim with the search's gear on it")
	healSeed         = flag.Int64("heal-seed", healscore.DefaultSeed, "with -heal-roles: the healer sims' random seed, the same for every set (healbench -seed)")
	healRefIters     = flag.Int("heal-ref-iters", 1000, "with -heal-roles: the iterations of the starting set's full carry on each slot, which every later set is estimated around")
)

// The band of failures (deaths over their line) the linear estimate is trusted in.
const healTrustLo, healTrustHi = 0.3, 3.0

// healSlot is one role on one boss.
type healSlot struct {
	role, boss string
	bench      healscore.Bench
	// raid is the -heal-raid-settings request, for a raid slot: its player and buffs replace the
	// setup's, with the setup's gear.
	raid *proto.RaidSimRequest

	// The starting set's carry and the failure's slope there, set by healReference.
	ref   healscore.Carry
	slope float64
	// healerAt is the party and place of the first copy of the healer in the slot's raid.
	healerAt [2]int
}

// healSlots are the healer objective's slots, nil without -heal-roles. Their index takes the place
// of a fight length in the search.
var healSlots []*healSlot

var healRefOnce sync.Once

// setupHeal reads the -heal-* flags into healSlots, and sets the fight lengths to the slots' indexes.
func setupHeal(durs *string, combined *bool) error {
	if *useDur > 0 {
		return fmt.Errorf("-heal-roles scores the uses inside the fights: run it with -use-dur 0")
	}
	var roles []string
	for _, r := range strings.Split(*healRoles, ",") {
		if r = strings.TrimSpace(r); r != "tank" && r != "raid" {
			return fmt.Errorf("-heal-roles is tank, raid or tank,raid, got %q", *healRoles)
		}
		roles = append(roles, r)
	}
	tankReq, err := healscore.ReadRequest(*healTank)
	if err != nil {
		return err
	}
	fillerReq, err := healscore.ReadRequest(*healFiller)
	if err != nil {
		return err
	}
	var raid *proto.RaidSimRequest
	for _, r := range roles {
		if r == "raid" && raid == nil {
			if *healRaidSettings == "" {
				return fmt.Errorf("-heal-roles raid needs -heal-raid-settings, the raid preset's request")
			}
			if raid, err = healscore.ReadRequest(*healRaidSettings); err != nil {
				return err
			}
		}
	}
	var labels []string
	for _, role := range roles {
		for _, path := range strings.Split(*healEncounter, ",") {
			path = strings.TrimSpace(path)
			fight, err := healscore.LoadFight(path)
			if err != nil {
				return err
			}
			o := healscore.DefaultOptions(role)
			o.Seed = *healSeed
			o.Duration = fight.Duration(role)
			slot := &healSlot{role: role, boss: strings.TrimSuffix(filepath.Base(path), ".json"),
				bench: healscore.Bench{O: o, Fight: fight, Tank: healscore.FirstPlayer(tankReq), Filler: healscore.FirstPlayer(fillerReq)}}
			if role == "raid" {
				slot.raid = raid
			}
			labels = append(labels, strconv.Itoa(len(healSlots)))
			healSlots = append(healSlots, slot)
		}
	}
	*durs = strings.Join(labels, ",")
	*combined = true
	fmt.Printf("healer objective: 100 x the mean damage carried over %d slots (their index takes the fight length's place), roles %s\n", len(healSlots), *healRoles)
	return nil
}

// healRequest is a slot's sim of the player in the setup request (its gear, talents and rotation
// already on it), at damage scale 1. A raid slot sims the raid preset's player in that gear.
func (slot *healSlot) healRequest(setupReq *proto.RaidSimRequest, iterations int32) *proto.RaidSimRequest {
	if slot.raid != nil {
		player := goproto.Clone(healscore.FirstPlayer(slot.raid)).(*proto.Player)
		player.Equipment = setupReq.Raid.Parties[0].Players[0].Equipment
		setupReq.Raid.Parties[0].Players[0] = player
		setupReq.Raid.Buffs, setupReq.Raid.Debuffs = slot.raid.Raid.Buffs, slot.raid.Raid.Debuffs
		// The raid preset's own rotation, with a synthetic trinket's use moved to the one worn.
		setRotation(setupReq, player.Rotation)
	}
	b := slot.bench
	b.Healer = setupReq.Raid.Parties[0].Players[0]
	b.Buffs, b.Debuffs = setupReq.Raid.Buffs, setupReq.Raid.Debuffs
	b.O.Iterations = int(iterations)
	req, err := b.Request(1, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	// The first healer's party and place, the same in every request of the slot.
find:
	for pi, party := range b.Slots {
		for si, who := range party {
			if who == healscore.RoleHealer {
				slot.healerAt = [2]int{pi, si}
				break find
			}
		}
	}
	return req
}

// healReference measures each slot's full carry on the set in setupReq, the starting set, and the
// failure's slope there. Slots run side by side, like scoreAll's runs.
func healReference(setupReq *proto.RaidSimRequest) {
	var wg sync.WaitGroup
	errs := make([]error, len(healSlots))
	sem := make(chan struct{}, scoreParallel)
	for i, slot := range healSlots {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			req := slot.healRequest(goproto.Clone(setupReq).(*proto.RaidSimRequest), int32(*healRefIters))
			slot.ref, errs[i] = slot.fullCarry(req)
			slot.slope = slot.ref.Slope()
		}()
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			fmt.Fprintf(os.Stderr, "heal reference %s %s: %v\n", healSlots[i].role, healSlots[i].boss, err)
			os.Exit(1)
		}
	}
	for i, slot := range healSlots {
		fmt.Printf("heal slot %d, %s %s: reference carry %.3f, failure slope %.2f (failures %.2f at %.3f, %.2f at %.3f)\n",
			i, slot.role, slot.boss, slot.ref.K, slot.slope, slot.ref.FailLo, slot.ref.Lo, slot.ref.FailHi, slot.ref.Hi)
	}
}

// fullCarry is healscore's carry of a slot request built at scale 1, as healbench finds it.
func (slot *healSlot) fullCarry(req *proto.RaidSimRequest) (healscore.Carry, error) {
	return healscore.CarrySearch(func(k float64) (float64, error) {
		run, err := slot.runAt(req, k)
		return run.fail, err
	})
}

// healRun is one sim of a slot at a scale: the failure, and which fights failed in seed order.
type healRun struct {
	fail    float64
	failed  []bool
	healer  *proto.UnitMetrics
	fights  int
	failCnt int
}

// runAt sims a slot request (built at scale 1) at a scale of its damage.
func (slot *healSlot) runAt(req *proto.RaidSimRequest, scale float64) (healRun, error) {
	req = goproto.Clone(req).(*proto.RaidSimRequest)
	healscore.ScaleDamage(req.Encounter, scale)
	res := core.RunRaidSimConcurrent(req)
	if res.Error != nil {
		return healRun{}, fmt.Errorf("%s", res.Error.Message)
	}
	hf := res.RaidMetrics.HealingFight
	if hf == nil || hf.Fights == 0 {
		return healRun{}, fmt.Errorf("no healing fight metrics")
	}
	run := healRun{fights: int(hf.Fights), failed: make([]bool, len(hf.FightDeaths))}
	for i, d := range hf.FightDeaths {
		if healscore.Failed(slot.role, d) {
			run.failed[i] = true
			run.failCnt++
		}
	}
	run.fail = float64(run.failCnt) / float64(run.fights) / healscore.Line(slot.role)
	run.healer = res.RaidMetrics.Parties[slot.healerAt[0]].Players[slot.healerAt[1]]
	return run, nil
}

// healPart is a slot's share of a set's score, for gainNoise: the fights that failed at the
// reference scale when the score is the linear estimate there.
type healPart struct {
	slot   *healSlot
	failed []bool // nil when the score came from the fallback bracket
	stderr float64
}

// estimate is the carry from one run at the reference scale: k + (1 - f)/s, when f is in the band
// the slope is trusted in. ok is false outside it.
func estimate(k, slope, f float64) (float64, bool) {
	if f <= 0 || f < healTrustLo || f > healTrustHi || !(slope > 0) || math.IsInf(slope, 0) {
		return 0, false
	}
	return k + (1-f)/slope, true
}

// binomialErr is the standard error of a carry from the share of fights failed at a failure slope:
// the binomial error of the share, over the line and the slope. A share of 0 counts as one fight.
func binomialErr(failCnt, fights int, line, slope float64) float64 {
	p := max(float64(failCnt), 1) / float64(fights)
	return math.Sqrt(p*(1-p)/float64(fights)) / line / slope
}

// simulateHeal scores one slot: the linear estimate at the reference carry, or the fallback
// bracket around it. The score is in percent of the boss's damage.
func simulateHeal(req *proto.RaidSimRequest, slot *healSlot) (result, error) {
	k := slot.ref.K
	run, err := slot.runAt(req, k)
	if err != nil {
		return result{}, err
	}
	line := healscore.Line(slot.role)
	carry, ok := estimate(k, slot.slope, run.fail)
	var part healPart
	if ok {
		part = healPart{slot: slot, failed: run.failed, stderr: 100 * binomialErr(run.failCnt, run.fights, line, slot.slope)}
	} else {
		var c healscore.Carry
		var nearCnt int
		if c, nearCnt, err = slot.fallback(req, k, run); err != nil {
			return result{}, err
		}
		carry = c.K
		slope := c.Slope()
		if !(slope > 0) || math.IsInf(slope, 0) {
			slope = slot.slope
		}
		part = healPart{slot: slot, stderr: 100 * binomialErr(nearCnt, run.fights, line, slope)}
	}
	r := result{dps: 100 * carry, stderr: part.stderr, healParts: []healPart{part}}
	if h := run.healer; h != nil {
		n := math.Sqrt(float64(run.fights))
		r.mana, r.manaStderr, r.tto = h.ManaEnd.GetAvg(), h.ManaEnd.GetStdev()/n, h.Tto.GetAvg()
	}
	return r, nil
}

// fallback brackets the line from the reference scale k (healBracket), simming the slot.
func (slot *healSlot) fallback(req *proto.RaidSimRequest, k float64, first healRun) (healscore.Carry, int, error) {
	return healBracket(k, first.fail, first.failCnt, func(x float64) (float64, int, error) {
		run, err := slot.runAt(req, x)
		return run.fail, run.failCnt, err
	})
}

// healBracket finds the carry from a scale k whose failure f0 (from cnt0 failed fights) is too far
// from the line for the linear estimate: up or down by 10% a step until the failure crosses 1, two
// bisections, and the interpolation between the two closest scales. It also gives the failed fights
// at the closer of the two, for the error. failure sims a scale.
func healBracket(k, f0 float64, cnt0 int, failure func(float64) (float64, int, error)) (healscore.Carry, int, error) {
	var err error
	cnt := map[float64]int{k: cnt0}
	f := func(x float64) float64 {
		if err != nil {
			return 0
		}
		var fx float64
		fx, cnt[x], err = failure(x)
		return fx
	}
	lo, hi, flo, fhi := k, k, f0, f0
	if f0 < 1 {
		for i := 0; i < 50 && fhi < 1; i++ {
			lo, flo = hi, fhi
			hi *= 1.1
			fhi = f(hi)
		}
	} else {
		for i := 0; i < 50 && flo >= 1; i++ {
			hi, fhi = lo, flo
			lo /= 1.1
			flo = f(lo)
		}
	}
	for i := 0; i < 2; i++ {
		mid := (lo + hi) / 2
		if fm := f(mid); fm < 1 {
			lo, flo = mid, fm
		} else {
			hi, fhi = mid, fm
		}
	}
	near := lo
	if math.Abs(fhi-1) < math.Abs(1-flo) {
		near = hi
	}
	return healscore.Interpolate(lo, hi, flo, fhi), cnt[near], err
}

// healGainNoise is the paired error of b's heal score less a's: where both slots scored by the
// estimate at the same scale, the error of the mean difference of their failed fights, over the
// line and the slope; elsewhere the two errors added up. ok is false when the results have no
// heal parts to pair.
func healGainNoise(a, b result) (float64, bool) {
	n := len(a.healParts)
	if n == 0 || n != len(b.healParts) {
		return 0, false
	}
	variance := 0.0
	for i := range a.healParts {
		pa, pb := a.healParts[i], b.healParts[i]
		if pa.failed == nil || len(pa.failed) != len(pb.failed) || pa.slot != pb.slot {
			variance += pa.stderr*pa.stderr + pb.stderr*pb.stderr
			continue
		}
		var sum, sumSq float64
		for j := range pa.failed {
			d := 0.0
			if pb.failed[j] {
				d++
			}
			if pa.failed[j] {
				d--
			}
			sum += d
			sumSq += d * d
		}
		m := float64(len(pa.failed))
		mean := sum / m
		se := math.Sqrt(max(sumSq/m-mean*mean, 0)/m) / healscore.Line(pa.slot.role) / pa.slot.slope
		variance += 100 * 100 * se * se
	}
	return math.Sqrt(variance) / float64(n), true
}
