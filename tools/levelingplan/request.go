package main

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	goproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var inf = math.Inf(1)

// What one sim of a level came to.
type levelHours struct {
	hours float64
	// The share of the hours that is noise.
	relativeError float64
}

// One sim to run: a level with a talents string.
type levelJob struct {
	level   int32
	talents string
}

// planner holds the plan and its gear plan, and makes and runs each level's request the way
// the Leveling Plan tab does (makeRequest in leveling_plan_tab.tsx).
type planner struct {
	req   *proto.LevelingPlanSearchRequest
	plan  *proto.LevelingPlanSettings
	shape *proto.LevelingTalentShape
	gear  *gearPlan
	seed  int64
	sims  atomic.Int32
}

func newPlanner(req *proto.LevelingPlanSearchRequest, db *proto.UIDatabase) (*planner, error) {
	if req.Base == nil || req.Plan == nil || req.TalentShape == nil {
		return nil, fmt.Errorf("the request needs base, plan and talent_shape")
	}
	p := &planner{req: req, plan: req.Plan, shape: req.TalentShape, seed: req.Base.GetSimOptions().GetRandomSeed()}
	p.gear = planLevelingGear(gearInputs(req, db))
	if err := p.checkUIGear(); err != nil {
		return nil, err
	}
	// The sims read the item database without a lock, so we add the plain items now, before
	// any sim runs. A request's own database would add them while other sims read it.
	database := &proto.SimDatabase{}
	for _, item := range p.gear.madeUp {
		sim, err := simItem(item)
		if err != nil {
			return nil, err
		}
		database.Items = append(database.Items, sim)
	}
	core.AddToDatabase(database)
	return p, nil
}

func gearInputs(req *proto.LevelingPlanSearchRequest, db *proto.UIDatabase) *gearPlanInputs {
	in := &gearPlanInputs{
		items:     db.Items,
		suffixes:  map[int32]*proto.ItemRandomSuffix{},
		itemsByID: map[int32]*proto.UIItem{},
		rules:     req.ClassRules,
		faction:   req.Faction,
		seed:      req.GetGearPlan().GetSeed(),
		pins:      req.GetGearPlan().GetPins(),
	}
	if player := req.Base.GetRaid().GetParties()[0].GetPlayers()[0]; player != nil {
		in.playerClass = player.Class
	}
	for _, s := range db.RandomSuffixes {
		in.suffixes[s.Id] = s
	}
	for _, item := range db.Items {
		in.itemsByID[item.Id] = item
	}
	presets := req.GearWeights
	fallback := weightsFromProto(req.GearDefaultWeights)
	// Between two presets we blend them by level, and outside them we take the nearest
	// (weightsAt in leveling_plan_gear.tsx).
	in.weightsAt = func(level int32) weights {
		if len(presets) == 0 {
			return fallback
		}
		above := -1
		for i, preset := range presets {
			if preset.Level >= level {
				above = i
				break
			}
		}
		if above == -1 {
			return weightsFromProto(presets[len(presets)-1].Weights)
		}
		if above == 0 || presets[above].Level == level {
			return weightsFromProto(presets[above].Weights)
		}
		lo, hi := weightsFromProto(presets[above-1].Weights), weightsFromProto(presets[above].Weights)
		t := float64(level-presets[above-1].Level) / float64(presets[above].Level-presets[above-1].Level)
		blend := func(a, b []float64) []float64 {
			out := make([]float64, max(len(a), len(b)))
			for i := range out {
				out[i] = at(a, i)*(1-t) + at(b, i)*t
			}
			return out
		}
		return weights{stats: blend(lo.stats, hi.stats), pseudo: blend(lo.pseudo, hi.pseudo)}
	}
	return in
}

// checkUIGear compares our gear plan with the UI's at each level the UI sent. They differ
// when leveling_gear.ts or gear.go changed without the other, or when the UI's item database
// isn't the one we have (a rebuilt db.bin the dev server hasn't served yet, say).
func (p *planner) checkUIGear() error {
	for i, spec := range p.req.UiGear {
		level := p.plan.FromLevel + int32(i)
		if level < 1 || int(level) > len(p.gear.levels) {
			continue
		}
		gear := p.gear.levels[level-1].gear
		for s, slot := range gearSlots {
			var want *proto.ItemSpec
			if s < len(spec.Items) {
				want = spec.Items[s]
			}
			piece, ok := gear[slot]
			var id, suffix int32
			if ok {
				id = piece.item.Id
				if piece.suffix != nil {
					suffix = piece.suffix.Id
				}
			}
			if want.GetId() != id || want.GetRandomSuffix() != suffix {
				return fmt.Errorf("the gear plan here differs from the UI's at level %d, %s: item %d (suffix %d) here, %d (suffix %d) in the UI",
					level, slot, id, suffix, want.GetId(), want.GetRandomSuffix())
			}
		}
	}
	return nil
}

// A UI item as the sim's database takes it, the way the UI's Gear.toDatabase does.
func simItem(item *proto.UIItem) (*proto.SimItem, error) {
	data, err := protojson.Marshal(item)
	if err != nil {
		return nil, err
	}
	out := &proto.SimItem{}
	err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(data, out)
	return out, err
}

// The request of one level, with these talents, sessions and seed.
func (p *planner) request(level int32, talents string, iterations int32, seed int64) *proto.RaidSimRequest {
	request := goproto.Clone(p.req.Base).(*proto.RaidSimRequest)
	request.SimOptions.DebugFirstIteration = false
	request.SimOptions.RandomSeed = seed
	request.SimOptions.Iterations = iterations
	player := request.Raid.Parties[0].Players[0]
	player.Level = level
	player.BonusTalentPoints = 0
	if level >= bonusTalentPointsLevel {
		player.BonusTalentPoints = p.plan.BonusTalentPoints
	}
	player.TalentsString = talents

	equipment := &proto.EquipmentSpec{Items: make([]*proto.ItemSpec, len(gearSlots))}
	gear := p.gear.levels[min(int(level), len(p.gear.levels))-1].gear
	for i, slot := range gearSlots {
		equipment.Items[i] = &proto.ItemSpec{}
		piece, ok := gear[slot]
		if !ok {
			continue
		}
		equipment.Items[i].Id = piece.item.Id
		if piece.suffix != nil {
			equipment.Items[i].RandomSuffix = piece.suffix.Id
		}
	}
	player.Equipment = equipment

	if !p.plan.UseSettingsBuffs {
		player.Consumes = &proto.Consumes{}
		player.Buffs = &proto.IndividualBuffs{}
		request.Raid.Buffs = &proto.RaidBuffs{}
		request.Raid.Debuffs = &proto.Debuffs{}
		for _, party := range request.Raid.Parties {
			party.Buffs = &proto.PartyBuffs{}
		}
	}
	if rotation := rotationAt(p.plan, level); rotation != nil {
		player.Rotation = rotation
	}
	if r := rangeAt(p.plan, level); r != nil && r.Imbue != proto.WeaponImbue_WeaponImbueUnknown {
		setShamanImbue(player, r.Imbue)
	}
	if leveling := request.Encounter.GetLeveling(); leveling != nil {
		leveling.Plan = false
	}
	return request
}

// Sets the imbue in the spec's options, for any spec whose options have a shaman_imbue.
func setShamanImbue(player *proto.Player, imbue proto.WeaponImbue) {
	m := player.ProtoReflect()
	oneof := m.Descriptor().Oneofs().ByName("spec")
	if oneof == nil {
		return
	}
	field := m.WhichOneof(oneof)
	if field == nil {
		return
	}
	spec := m.Mutable(field).Message()
	optionsField := spec.Descriptor().Fields().ByName("options")
	if optionsField == nil {
		return
	}
	options := spec.Mutable(optionsField).Message()
	if f := options.Descriptor().Fields().ByName("shaman_imbue"); f != nil {
		options.Set(f, protoreflect.ValueOfEnum(protoreflect.EnumNumber(imbue)))
	}
}

// A level's seed comes from its number, so two talents strings at a level play the same mobs
// and the same rolls.
func levelSeed(seed int64, level int32) int64 {
	return seed + int64(level)*1_000_003
}

// simLevels sims each job on all the cores, one sim a core, and gives back each one's hours.
// A job we didn't get to before ctx ended gives nothing (ok false).
func (p *planner) simLevels(ctx context.Context, jobs []levelJob, iterations int32, seed int64) ([]levelHours, []bool) {
	results := make([]levelHours, len(jobs))
	done := make([]bool, len(jobs))
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
				res := core.RunRaidSim(p.request(job.level, job.talents, iterations, levelSeed(seed, job.level)))
				if res.Error != nil {
					continue
				}
				m := res.EncounterMetrics.GetLeveling()
				if m == nil {
					continue
				}
				r := levelHours{hours: hoursAtLevel(job.level, m.XpPerHour)}
				if m.MobsKilled > 0 {
					r.relativeError = m.MobsKilledStdev / math.Sqrt(float64(max(iterations, 1))) / m.MobsKilled
				}
				results[i], done[i] = r, true
				p.sims.Add(1)
			}
		}()
	}
	wg.Wait()
	return results, done
}
