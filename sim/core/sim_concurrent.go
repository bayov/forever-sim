package core

import (
	"fmt"
	"log"
	"math"
	"reflect"
	"runtime"
	"runtime/debug"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	googleProto "google.golang.org/protobuf/proto"
)

// Will split into min(splitCount, iterations) requests.
func SplitSimRequestForConcurrency(request *proto.RaidSimRequest, splitCount int32) *proto.RaidSimRequestSplitResult {
	res := &proto.RaidSimRequestSplitResult{}

	if splitCount <= 0 {
		res.ErrorResult = "Split count can't be 0 or negative!"
		return res
	}

	if request.SimOptions.Iterations <= 0 {
		res.ErrorResult = "Iterations can't be 0 or negative!"
		return res
	}

	splitCount = min(splitCount, request.SimOptions.Iterations)

	split := make([]*proto.RaidSimRequest, splitCount)
	iterPerSplit := request.SimOptions.Iterations / splitCount

	split[0] = googleProto.Clone(request).(*proto.RaidSimRequest)
	split[0].SimOptions.Iterations = iterPerSplit + request.SimOptions.Iterations%splitCount
	// No logs from the splits. Each one picks its typical, lucky and unlucky iterations,
	// and we replay the overall ones once the results are combined (see
	// ReplayIterationsRequest).
	split[0].SimOptions.DebugFirstIteration = false

	// Sims increment their seed each iteration. Offset starting seed of each split to emulate that.
	nextStartSeed := split[0].SimOptions.RandomSeed + int64(split[0].SimOptions.Iterations)

	for i := 1; i < int(splitCount); i++ {
		split[i] = googleProto.Clone(request).(*proto.RaidSimRequest)
		split[i].SimOptions.Iterations = iterPerSplit
		split[i].SimOptions.DebugFirstIteration = false
		split[i].SimOptions.RandomSeed = nextStartSeed
		nextStartSeed += int64(split[i].SimOptions.Iterations)
	}

	res.SplitsDone = splitCount
	res.Requests = split
	return res
}

type raidSimResultCombiner struct {
	Debug    bool
	Combined *proto.RaidSimResult
}

func (rsrc *raidSimResultCombiner) newDistMetrics() *proto.DistributionMetrics {
	return &proto.DistributionMetrics{
		Min:            math.MaxFloat64,
		MinSeed:        math.MaxInt64,
		Hist:           make(map[int32]int32),
		AllValues:      make([]float64, 0),
		AggregatorData: &proto.AggregatorData{},
	}
}

func (rsrc *raidSimResultCombiner) newUnitMetrics(baseUnit *proto.UnitMetrics) *proto.UnitMetrics {
	newUm := &proto.UnitMetrics{
		Name:      baseUnit.Name,
		UnitIndex: baseUnit.UnitIndex,
		Dps:       rsrc.newDistMetrics(),
		Dpasp:     rsrc.newDistMetrics(),
		Threat:    rsrc.newDistMetrics(),
		Dtps:      rsrc.newDistMetrics(),
		Tmi:       rsrc.newDistMetrics(),
		Hps:       rsrc.newDistMetrics(),
		Tto:       rsrc.newDistMetrics(),
		ManaEnd:   rsrc.newDistMetrics(),
		Actions:   make([]*proto.ActionMetrics, 0, len(baseUnit.Actions)),
		Auras:     make([]*proto.AuraMetrics, len(baseUnit.Auras)),
		Resources: make([]*proto.ResourceMetrics, 0, len(baseUnit.Resources)),
		Pets:      make([]*proto.UnitMetrics, len(baseUnit.Pets)),
	}

	for i, aura := range baseUnit.Auras {
		newUm.Auras[i] = &proto.AuraMetrics{
			Id:             aura.Id,
			AggregatorData: &proto.AggregatorData{},
		}
	}

	for i, pet := range baseUnit.Pets {
		newUm.Pets[i] = rsrc.newUnitMetrics(pet)
	}

	return newUm
}

func (rsrc *raidSimResultCombiner) newPartyMetrics(baseParty *proto.PartyMetrics) *proto.PartyMetrics {
	newPm := &proto.PartyMetrics{
		Dps:     rsrc.newDistMetrics(),
		Hps:     rsrc.newDistMetrics(),
		Players: make([]*proto.UnitMetrics, len(baseParty.Players)),
	}

	for i, player := range baseParty.Players {
		newPm.Players[i] = rsrc.newUnitMetrics(player)
	}

	return newPm
}

func (rsrc *raidSimResultCombiner) combineDistMetrics(base *proto.DistributionMetrics, add *proto.DistributionMetrics, isLast bool, weight float64) {
	base.Avg += add.Avg * weight

	if add.Max > base.Max {
		base.Max = add.Max
		base.MaxSeed = add.MaxSeed
	}

	if add.Min == 0 || add.Min < base.Min {
		base.Min = add.Min
		base.MinSeed = add.MinSeed
	} else if add.Min == base.Min {
		base.MinSeed = add.MinSeed
	}

	for idx, val := range add.Hist {
		base.Hist[idx] += val
	}

	base.AllValues = append(base.AllValues, add.AllValues...)

	base.AggregatorData.N += add.AggregatorData.N
	base.AggregatorData.SumSq += add.AggregatorData.SumSq
	if isLast {
		base.Stdev = math.Sqrt(base.AggregatorData.SumSq/float64(base.AggregatorData.N) - base.Avg*base.Avg)
	}
}

func (rsrc *raidSimResultCombiner) addActionMetrics(unit *proto.UnitMetrics, add *proto.ActionMetrics) {
	var am *proto.ActionMetrics

	addKey := add.Id.String()
	for _, baseAction := range unit.Actions {
		if baseAction.Id.String() == addKey {
			am = baseAction
			break
		}
	}

	if am == nil {
		am = &proto.ActionMetrics{
			Id:          add.Id,
			IsMelee:     add.IsMelee,
			IsPassive:   add.IsPassive,
			Targets:     make([]*proto.TargetedActionMetrics, len(add.Targets)),
			SpellSchool: add.SpellSchool,
		}
		for i, addTgt := range add.Targets {
			am.Targets[i] = &proto.TargetedActionMetrics{
				UnitIndex: addTgt.UnitIndex,
			}
		}
		unit.Actions = append(unit.Actions, am)
	}

	for i, baseTgt := range am.Targets {
		addTgt := add.Targets[i]
		if baseTgt.UnitIndex != addTgt.UnitIndex {
			panic("Unitidx doesn't match?!")
		}
		baseTgt.Casts += addTgt.Casts
		baseTgt.Hits += addTgt.Hits
		baseTgt.ResistedHits += addTgt.ResistedHits
		baseTgt.Crits += addTgt.Crits
		baseTgt.ResistedCrits += addTgt.ResistedCrits
		baseTgt.Ticks += addTgt.Ticks
		baseTgt.ResistedTicks += addTgt.ResistedTicks
		baseTgt.CritTicks += addTgt.CritTicks
		baseTgt.ResistedCritTicks += addTgt.ResistedCritTicks
		baseTgt.Misses += addTgt.Misses
		baseTgt.Dodges += addTgt.Dodges
		baseTgt.Parries += addTgt.Parries
		baseTgt.Blocks += addTgt.Blocks
		baseTgt.BlockedCrits += addTgt.BlockedCrits
		baseTgt.Glances += addTgt.Glances
		baseTgt.Crushes += addTgt.Crushes
		baseTgt.Damage += addTgt.Damage
		baseTgt.ResistedDamage += addTgt.ResistedDamage
		baseTgt.CritDamage += addTgt.CritDamage
		baseTgt.ResistedCritDamage += addTgt.ResistedCritDamage
		baseTgt.TickDamage += addTgt.TickDamage
		baseTgt.ResistedTickDamage += addTgt.ResistedTickDamage
		baseTgt.CritTickDamage += addTgt.CritTickDamage
		baseTgt.ResistedCritTickDamage += addTgt.ResistedCritTickDamage
		baseTgt.GlanceDamage += addTgt.GlanceDamage
		baseTgt.CrushDamage += addTgt.CrushDamage
		baseTgt.BlockDamage += addTgt.BlockDamage
		baseTgt.BlockedCritDamage += addTgt.BlockedCritDamage
		baseTgt.Threat += addTgt.Threat
		baseTgt.HitThreat += addTgt.HitThreat
		baseTgt.CritThreat += addTgt.CritThreat
		baseTgt.TickThreat += addTgt.TickThreat
		baseTgt.CritTickThreat += addTgt.CritTickThreat
		baseTgt.GlanceThreat += addTgt.GlanceThreat
		baseTgt.BlockThreat += addTgt.BlockThreat
		baseTgt.BlockedCritThreat += addTgt.BlockedCritThreat
		baseTgt.CrushThreat += addTgt.CrushThreat
		baseTgt.Healing += addTgt.Healing
		baseTgt.CritHealing += addTgt.CritHealing
		baseTgt.Shielding += addTgt.Shielding
		baseTgt.CastTimeMs += addTgt.CastTimeMs
	}
}

func (rsrc *raidSimResultCombiner) combineAuraMetrics(base *proto.AuraMetrics, add *proto.AuraMetrics, weight float64, isLast bool) {
	base.UptimeSecondsAvg += add.UptimeSecondsAvg * weight
	base.ProcsAvg += add.ProcsAvg * weight

	base.AggregatorData.N += add.AggregatorData.N
	base.AggregatorData.SumSq += add.AggregatorData.SumSq
	if isLast {
		base.UptimeSecondsStdev = math.Sqrt(base.AggregatorData.SumSq/float64(base.AggregatorData.N) - base.UptimeSecondsAvg*base.UptimeSecondsAvg)
	}
}

func (rsrc *raidSimResultCombiner) addResourceMetrics(unit *proto.UnitMetrics, add *proto.ResourceMetrics) {
	var rm *proto.ResourceMetrics

	rkey := func(r *proto.ResourceMetrics) string {
		return fmt.Sprintf("%s-%d", r.Id.String(), r.Type)
	}

	for _, baseResource := range unit.Resources {
		if rkey(baseResource) == rkey(add) {
			rm = baseResource
			break
		}
	}

	if rm == nil {
		rm = &proto.ResourceMetrics{
			Id:   add.Id,
			Type: add.Type,
		}
		unit.Resources = append(unit.Resources, rm)
	}

	rm.Events += add.Events
	rm.Gain += add.Gain
	rm.ActualGain += add.ActualGain
}

func (rsrc *raidSimResultCombiner) combineUnitMetrics(base *proto.UnitMetrics, add *proto.UnitMetrics, isLast bool, weight float64) {
	// An empty slot in a party has no metrics to add.
	if add.Dps == nil || base.Dps == nil {
		return
	}
	rsrc.combineDistMetrics(base.Dps, add.Dps, isLast, weight)
	rsrc.combineDistMetrics(base.Dpasp, add.Dpasp, isLast, weight)
	rsrc.combineDistMetrics(base.Threat, add.Threat, isLast, weight)
	rsrc.combineDistMetrics(base.Dtps, add.Dtps, isLast, weight)
	rsrc.combineDistMetrics(base.Tmi, add.Tmi, isLast, weight)
	rsrc.combineDistMetrics(base.Hps, add.Hps, isLast, weight)
	rsrc.combineDistMetrics(base.Tto, add.Tto, isLast, weight)
	rsrc.combineDistMetrics(base.ManaEnd, add.ManaEnd, isLast, weight)

	base.SecondsOomAvg += add.SecondsOomAvg * weight
	base.ChanceOfDeath += add.ChanceOfDeath * weight

	for _, addAction := range add.Actions {
		rsrc.addActionMetrics(base, addAction)
	}

	for i, addAura := range add.Auras {
		rsrc.combineAuraMetrics(base.Auras[i], addAura, weight, isLast)
	}

	for _, addResource := range add.Resources {
		rsrc.addResourceMetrics(base, addResource)
	}

	for i, addPet := range add.Pets {
		rsrc.combineUnitMetrics(base.Pets[i], addPet, isLast, weight)
	}
}

func (rsrc *raidSimResultCombiner) AddResult(result *proto.RaidSimResult, isLast bool, weight float64) {
	rsrc.combineDistMetrics(rsrc.Combined.RaidMetrics.Dps, result.RaidMetrics.Dps, isLast, weight)
	rsrc.combineDistMetrics(rsrc.Combined.RaidMetrics.Hps, result.RaidMetrics.Hps, isLast, weight)

	for partyIdx, party := range result.RaidMetrics.Parties {
		baseParty := rsrc.Combined.RaidMetrics.Parties[partyIdx]
		rsrc.combineDistMetrics(baseParty.Dps, party.Dps, isLast, weight)
		rsrc.combineDistMetrics(baseParty.Hps, party.Hps, isLast, weight)
		for playerIdx, player := range party.Players {
			rsrc.combineUnitMetrics(baseParty.Players[playerIdx], player, isLast, weight)
		}
	}

	for i, tar := range result.EncounterMetrics.Targets {
		rsrc.combineUnitMetrics(rsrc.Combined.EncounterMetrics.Targets[i], tar, isLast, weight)
	}
	if add := result.EncounterMetrics.Leveling; add != nil {
		combineLevelingMetrics(rsrc.Combined.EncounterMetrics.Leveling, add, isLast, weight)
	}

	rsrc.Combined.AvgIterationDuration += result.AvgIterationDuration * weight
	rsrc.Combined.IterationsDone += result.IterationsDone

	if rsrc.Debug {
		rsrc.Combined.Logs += "-SIMSTART-\n" + result.Logs
	}
}

func (rsrc *raidSimResultCombiner) SetBaseResult(baseRsr *proto.RaidSimResult) {
	newRsr := &proto.RaidSimResult{
		RaidMetrics: &proto.RaidMetrics{
			Dps:     rsrc.newDistMetrics(),
			Hps:     rsrc.newDistMetrics(),
			Parties: make([]*proto.PartyMetrics, len(baseRsr.RaidMetrics.Parties)),
		},
		EncounterMetrics: &proto.EncounterMetrics{
			Targets: make([]*proto.UnitMetrics, len(baseRsr.EncounterMetrics.Targets)),
		},
		FirstIterationDuration: baseRsr.FirstIterationDuration,
	}

	if !rsrc.Debug {
		newRsr.Logs = baseRsr.Logs
	}

	for i, party := range baseRsr.RaidMetrics.Parties {
		newRsr.RaidMetrics.Parties[i] = rsrc.newPartyMetrics(party)
	}

	for i, tar := range baseRsr.EncounterMetrics.Targets {
		newRsr.EncounterMetrics.Targets[i] = rsrc.newUnitMetrics(tar)
	}
	if baseRsr.EncounterMetrics.Leveling != nil {
		newRsr.EncounterMetrics.Leveling = &proto.LevelingMetrics{}
	}

	rsrc.Combined = newRsr
}

func CombineConcurrentSimResults(results []*proto.RaidSimResult, isDebug bool) *proto.RaidSimResult {
	numResults := len(results)

	if numResults == 0 {
		panic("Result set is empty!")
	}

	if numResults == 1 {
		return results[0]
	}

	var totalIterations int32 = 0
	for _, req := range results {
		totalIterations += req.IterationsDone
	}

	rsrc := raidSimResultCombiner{Debug: isDebug}
	rsrc.SetBaseResult(results[0])
	for i, result := range results {
		resultWeight := float64(results[i].IterationsDone) / float64(totalIterations)
		rsrc.AddResult(result, i == numResults-1, resultWeight)
	}

	combinePickedIterations(rsrc.Combined, results)

	return rsrc.Combined
}

// combinePickedIterations picks the overall typical, lucky and unlucky iterations from the
// ones each split sim picked, and in a leveling run the overall died one.
//
// Lucky and unlucky are the highest and lowest of the splits. The typical one is the
// split's typical iteration closest to the combined average, which is close to the
// iteration closest to that average over all iterations.
func combinePickedIterations(combined *proto.RaidSimResult, results []*proto.RaidSimResult) {
	leveling := combined.EncounterMetrics.GetLeveling() != nil
	score := func(it *proto.IterationLog) float64 { return iterationScore(it, leveling) }
	mean := combined.RaidMetrics.Dps.Avg
	if leveling {
		mean = combined.EncounterMetrics.Leveling.MobsKilled
	}
	for _, result := range results {
		if it := result.TypicalIteration; it != nil {
			if combined.TypicalIteration == nil || math.Abs(score(it)-mean) < math.Abs(score(combined.TypicalIteration)-mean) {
				combined.TypicalIteration = it
			}
		}
		if it := result.LuckyIteration; it != nil {
			if combined.LuckyIteration == nil || score(it) > score(combined.LuckyIteration) {
				combined.LuckyIteration = it
			}
		}
		if it := result.UnluckyIteration; it != nil {
			if combined.UnluckyIteration == nil || score(it) < score(combined.UnluckyIteration) {
				combined.UnluckyIteration = it
			}
		}
		combined.DiedIteration = pickDied(combined.DiedIteration, result.DiedIteration, mean)
	}
}

// ReplayIterationsRequest returns the request that runs the typical, lucky and unlucky
// iterations of a combined concurrent result again with logs on, or nil when the result
// has none.
func ReplayIterationsRequest(request *proto.RaidSimRequest, result *proto.RaidSimResult) *proto.RaidSimRequest {
	if result.TypicalIteration == nil || result.LuckyIteration == nil || result.UnluckyIteration == nil {
		return nil
	}
	replay := googleProto.Clone(request).(*proto.RaidSimRequest)
	replay.SimOptions.Iterations = 1
	replay.SimOptions.ReplaySeeds = []int64{result.TypicalIteration.Seed, result.LuckyIteration.Seed, result.UnluckyIteration.Seed}
	if result.DiedIteration != nil {
		replay.SimOptions.ReplaySeeds = append(replay.SimOptions.ReplaySeeds, result.DiedIteration.Seed)
	}
	return replay
}

// ApplyIterationReplays copies the logs from the replay of ReplayIterationsRequest into
// the combined result.
func ApplyIterationReplays(result *proto.RaidSimResult, replay *proto.RaidSimResult) {
	if replay == nil || replay.Error != nil {
		return
	}
	result.Logs = replay.Logs
	result.FirstIterationDuration = replay.FirstIterationDuration
	result.TypicalIteration = replay.TypicalIteration
	result.LuckyIteration = replay.LuckyIteration
	result.UnluckyIteration = replay.UnluckyIteration
	result.DiedIteration = replay.DiedIteration
}

type concurrentSimData struct {
	Concurrency     int32
	IterationsTotal int32
	IterationsDone  []int32

	DpsValues []float64
	HpsValues []float64

	FinalResults []*proto.RaidSimResult
}

func (csd *concurrentSimData) GetIterationsDone() int32 {
	var total int32 = 0
	for _, done := range csd.IterationsDone {
		total += done
	}
	return total
}

func (csd *concurrentSimData) GetDpsAvg() float64 {
	total := 0.0
	for _, done := range csd.DpsValues {
		total += done
	}
	return total / float64(csd.Concurrency)
}

func (csd *concurrentSimData) GetHpsAvg() float64 {
	total := 0.0
	for _, done := range csd.HpsValues {
		total += done
	}
	return total / float64(csd.Concurrency)
}

func (csd *concurrentSimData) UpdateProgress(idx int, msg *proto.ProgressMetrics) bool {
	csd.IterationsDone[idx] = msg.CompletedIterations
	csd.DpsValues[idx] = msg.Dps
	csd.HpsValues[idx] = msg.Hps

	if msg.FinalRaidResult != nil {
		csd.FinalResults[idx] = msg.FinalRaidResult
		return true
	}

	return false
}

func (csd *concurrentSimData) MakeProgressMetrics() *proto.ProgressMetrics {
	return &proto.ProgressMetrics{
		TotalIterations:     csd.IterationsTotal,
		CompletedIterations: csd.GetIterationsDone(),
		Dps:                 csd.GetDpsAvg(),
		Hps:                 csd.GetHpsAvg(),
	}
}

// Run sim on multiple threads concurrently by splitting interations over multiple sims, transparently combining results into the progress channel.
func runSimConcurrent(request *proto.RaidSimRequest, progress chan *proto.ProgressMetrics, signals simsignals.Signals) (result *proto.RaidSimResult) {
	defer func() {
		if !request.SimOptions.IsTest {
			if err := recover(); err != nil {
				errStr := ""
				switch errt := err.(type) {
				case string:
					errStr = errt
				case error:
					errStr = errt.Error()
				}

				errStr += "\nStack Trace:\n" + string(debug.Stack())
				result = &proto.RaidSimResult{Error: &proto.ErrorOutcome{Message: errStr}}

				if progress != nil {
					progress <- &proto.ProgressMetrics{FinalRaidResult: result}
				}

				signals.Abort.Trigger()
			}
		}

		if progress != nil {
			close(progress)
		}
	}()

	splitRes := SplitSimRequestForConcurrency(request, TernaryInt32(request.SimOptions.IsTest, 3, int32(runtime.NumCPU())))

	if splitRes.ErrorResult != "" {
		panic(splitRes.ErrorResult)
	}

	threads := splitRes.SplitsDone
	substituteChannels := make([]chan *proto.ProgressMetrics, threads)
	substituteCases := make([]reflect.SelectCase, threads)
	running := threads

	csd := concurrentSimData{
		Concurrency:     threads,
		IterationsTotal: request.SimOptions.Iterations,
		IterationsDone:  make([]int32, threads),
		DpsValues:       make([]float64, threads),
		HpsValues:       make([]float64, threads),
		FinalResults:    make([]*proto.RaidSimResult, threads),
	}

	for i := 0; i < int(threads); i++ {
		substituteChannels[i] = make(chan *proto.ProgressMetrics, 20)
		substituteCases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(substituteChannels[i])}
	}

	if !request.SimOptions.IsTest {
		log.Printf("Running %d iterations on %d concurrent sims.", csd.IterationsTotal, csd.Concurrency)
	}

	for i, req := range splitRes.Requests {
		go RunSim(req, substituteChannels[i], signals)
	}

	progressCounter := 0

	for running > 0 {
		i, val, ok := reflect.Select(substituteCases)

		if signals.Abort.IsTriggered() {
			quitResult := &proto.RaidSimResult{Error: &proto.ErrorOutcome{Type: proto.ErrorOutcomeType_ErrorOutcomeAborted}}
			if progress != nil {
				progress <- &proto.ProgressMetrics{FinalRaidResult: quitResult}
			}
			return quitResult
		}

		if !ok {
			substituteCases[i].Chan = reflect.ValueOf(nil)
			running -= 1
			continue
		}

		msg := val.Interface().(*proto.ProgressMetrics)
		if csd.UpdateProgress(i, msg) {
			if msg.FinalRaidResult != nil && msg.FinalRaidResult.Error != nil {
				if progress != nil {
					progress <- msg
				}
				log.Printf("Thread %d had an error. Cancelling all sims!", i)
				signals.Abort.Trigger()
				return msg.FinalRaidResult
			}
			substituteCases[i].Chan = reflect.ValueOf(nil)
			running -= 1
			continue
		}

		if progress != nil {
			progressCounter++ // Don't spam progress
			if progressCounter%int(threads) == 0 {
				progress <- csd.MakeProgressMetrics()
			}
		}
	}

	for _, res := range csd.FinalResults {
		if res == nil {
			panic("Missing one or more final sim result(s)!")
		}
	}

	if !request.SimOptions.IsTest {
		log.Printf("All %d sims finished successfully.", csd.Concurrency)
	}

	result = CombineConcurrentSimResults(csd.FinalResults, request.SimOptions.Debug)

	if request.SimOptions.DebugFirstIteration && !request.SimOptions.Debug {
		if replayRequest := ReplayIterationsRequest(request, result); replayRequest != nil {
			ApplyIterationReplays(result, RunSim(replayRequest, nil, signals))
		}
	}

	if progress != nil {
		pm := csd.MakeProgressMetrics()
		pm.FinalRaidResult = result
		progress <- pm
	}

	return result
}

// combineLevelingMetrics adds one thread's leveling averages, weighted by its share of the
// iterations. The spread of mobs killed adds up as mean squares (stdev^2 + mean^2) and comes
// back to a standard deviation on the last one.
func combineLevelingMetrics(base, add *proto.LevelingMetrics, isLast bool, weight float64) {
	base.MobsKilled += add.MobsKilled * weight
	base.MobsKilledStdev += (add.MobsKilledStdev*add.MobsKilledStdev + add.MobsKilled*add.MobsKilled) * weight
	base.Xp += add.Xp * weight
	base.XpPerHour += add.XpPerHour * weight
	base.Deaths += add.Deaths * weight
	base.SecondsInCombat += add.SecondsInCombat * weight
	base.SecondsResting += add.SecondsResting * weight
	base.SecondsSearching += add.SecondsSearching * weight
	base.SecondsDead += add.SecondsDead * weight
	base.MobsPerPull += add.MobsPerPull * weight
	base.Food += add.Food * weight
	base.Drinks += add.Drinks * weight
	base.Bandages += add.Bandages * weight
	base.AllKills = append(base.AllKills, add.AllKills...)
	if isLast {
		base.MobsKilledStdev = math.Sqrt(max(base.MobsKilledStdev-base.MobsKilled*base.MobsKilled, 0))
	}
}
