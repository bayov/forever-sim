package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/signal"
	"slices"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	goproto "google.golang.org/protobuf/proto"
)

// The stat weights of a plan's gear (-weights), which the gear plan judges items by.
//
// At every 5th level we play the plan with its gear, talents and rotation there, and again with a
// little more of each stat, on the same random rolls. A stat is worth the mobs it adds in a
// session next to what a point of attack power adds.
//
// The weights move with the gear, because Stamina and the rest decide how often we die and rest,
// and the gear the plan picks by one set of weights changes the next. So we take rounds: the gear
// of the first is by the weights the plan came with, and of each later one by the mean of the
// rounds before it. The result is the mean of all the rounds.

// A probe is an amount of one stat on top of the gear.
type probe struct {
	// The stat's name in the weights file.
	name   string
	stat   proto.Stat
	pseudo bool
	weapon proto.PseudoStat
	amount float64
}

func statProbe(name string, amount float64) probe {
	return probe{name: name, stat: proto.Stat(proto.Stat_value["Stat"+name]), amount: amount}
}

// addTo gives the player the probe's stat.
func (p probe) addTo(player *proto.Player) {
	bonus := &proto.UnitStats{Stats: make([]float64, len(proto.Stat_name)), PseudoStats: make([]float64, len(proto.PseudoStat_name))}
	if player.BonusStats != nil {
		copy(bonus.Stats, player.BonusStats.Stats)
		copy(bonus.PseudoStats, player.BonusStats.PseudoStats)
	}
	if p.pseudo {
		bonus.PseudoStats[p.weapon] += p.amount
	} else {
		bonus.Stats[p.stat] += p.amount
	}
	player.BonusStats = bonus
}

// The probes of attack power and hit have their own rules (weightsRow.fill).
const (
	probeAttackPower = "AttackPower"
	probeHit         = "MeleeHit"
)

// probesFor are the stats we probe for a class and what it holds at a level.
func probesFor(class proto.Class, gear map[proto.ItemSlot]gearPiece) []probe {
	out := []probe{
		statProbe(probeAttackPower, 20),
		statProbe("Strength", 10),
		statProbe("Agility", 10),
		statProbe("Stamina", 10),
		statProbe("Spirit", 10),
		statProbe(probeHit, 1),
		statProbe(probeHit, -1),
		statProbe("MeleeCrit", 2),
		statProbe("Armor", 200),
		statProbe("Defense", 10),
		statProbe("Dodge", 5),
		statProbe("Parry", 5),
		{name: "MainHandDps", pseudo: true, weapon: proto.PseudoStat_PseudoStatMainHandDps, amount: 2},
	}
	if class != proto.Class_ClassWarrior && class != proto.Class_ClassRogue {
		out = append(out, statProbe("Intellect", 10), statProbe("MP5", 5), statProbe("SpellPower", 20), statProbe("FirePower", 20), statProbe("NaturePower", 20))
	}
	switch offHand, ok := gear[proto.ItemSlot_ItemSlotOffHand]; {
	case !ok || offHand.item.WeaponType == proto.WeaponType_WeaponTypeOffHand:
	case offHand.item.WeaponType == proto.WeaponType_WeaponTypeShield:
		out = append(out, statProbe("Block", 5), statProbe("BlockValue", 20))
	default:
		out = append(out, probe{name: "OffHandDps", pseudo: true, weapon: proto.PseudoStat_PseudoStatOffHandDps, amount: 2})
	}
	return out
}

// The weights a plan starts from when it has none. They only pick the gear of the first round.
var startWeights = map[string]float64{
	"AttackPower": 1, "Strength": 2, "Agility": 0.5, "Stamina": 0.5, "MeleeHit": 10, "MeleeCrit": 10,
	"Armor": 0.02, "MainHandDps": 12, "OffHandDps": 6, "Block": 1, "BlockValue": 0.2,
}

type weightsOptions struct {
	hoursOptions
	rounds int
	// The style to measure from level 1, or the plan's own styles.
	style *proto.LevelingWeaponStyle
}

// weightLevels are the levels we measure at: every 5th, and the plan's last.
func weightLevels(plan *proto.LevelingPlanSettings) []int32 {
	var out []int32
	levels := planLevels(plan)
	for i, level := range levels {
		if level%5 == 0 || i == len(levels)-1 {
			out = append(out, level)
		}
	}
	return out
}

// weightsRound measures the weights on the gear the request's weights pick.
func weightsRound(ctx context.Context, db *proto.UIDatabase, req *proto.LevelingPlanSearchRequest, opts weightsOptions, levels []int32) (*weightsFile, error) {
	var planners []*planner
	for i := range opts.gearSeeds {
		p, err := newPlanner(withGearSeed(req, opts.firstGearSeed+i), db)
		if err != nil {
			return nil, err
		}
		planners = append(planners, p)
	}
	type job struct {
		level, seed int
		// -1 for the plan itself.
		probe int
		add   probe
		kills []float64
		mobs  float64
		dead  float64
		err   error
	}
	var jobs []*job
	class := req.Base.Raid.Parties[0].Players[0].Class
	for l, level := range levels {
		for g, p := range planners {
			jobs = append(jobs, &job{level: l, seed: g, probe: -1})
			for i, add := range probesFor(class, p.gear.levels[level-1].gear) {
				jobs = append(jobs, &job{level: l, seed: g, probe: i, add: add})
			}
		}
	}
	parallel(ctx, len(jobs), func(i int) {
		j := jobs[i]
		p, level := planners[j.seed], levels[j.level]
		request := p.request(levelJob{level: level, talents: planTalentsAt(p.plan, level, p.shape)}, opts.sessions, levelSeed(opts.seed+int64(j.seed)*104_729, level))
		request.SimOptions.SaveAllValues = true
		// Rounded hits only make the gap between two runs noisier.
		request.SimOptions.ExactDamage = true
		if j.probe >= 0 {
			j.add.addTo(request.Raid.Parties[0].Players[0])
		}
		res := core.RunRaidSim(request)
		m := res.EncounterMetrics.GetLeveling()
		switch {
		case res.Error != nil:
			j.err = fmt.Errorf("level %d: %s", level, firstLine(res.Error.Message))
		case m == nil:
			j.err = fmt.Errorf("the request has no leveling session")
		default:
			j.kills, j.mobs, j.dead = m.AllKills, m.MobsKilled, m.Deaths
		}
	})
	if ctx.Err() != nil {
		return nil, errStopped
	}
	file := &weightsFile{}
	for l, level := range levels {
		// The gains of a probe by gear seed. A seed without the probe (no shield yet, say) counts
		// as none, since it gains nothing from the stat there.
		gains := map[string][]float64{}
		variance := map[string]float64{}
		row := weightsRow{Level: level}
		n := float64(len(planners))
		for g := range planners {
			var base *job
			for _, j := range jobs {
				if j.level == l && j.seed == g && j.probe == -1 {
					base = j
				}
			}
			if base.err != nil {
				return nil, base.err
			}
			row.Mobs += base.mobs / n
			row.Deaths += base.dead / n
			for _, j := range jobs {
				if j.level != l || j.seed != g || j.probe == -1 {
					continue
				}
				if j.err != nil {
					return nil, j.err
				}
				gain, err := pairedGain(base.kills, j.kills)
				gain, err = gain/j.add.amount, err/math.Abs(j.add.amount)
				key := j.add.name
				if j.add.amount < 0 {
					key += "-"
				}
				gains[key] = append(gains[key], gain)
				variance[key] += err * err
			}
		}
		mean := map[string]float64{}
		errs := map[string]float64{}
		for key, values := range gains {
			sum := 0.0
			for _, v := range values {
				sum += v
			}
			mean[key] = sum / n
			errs[key] = math.Sqrt(variance[key]) / n
		}
		if err := row.fill(mean, errs); err != nil {
			return nil, err
		}
		file.Levels = append(file.Levels, row)
	}
	return file, nil
}

// fill sets the row's weights from the mobs a point of each stat adds.
//
// Hit is the larger of 1% more and 1% less. At the hit cap more adds nothing, but an item that
// takes 1% away still costs us. A weight below 0 is noise, and we write 0.
func (row *weightsRow) fill(gains, errs map[string]float64) error {
	ap := gains[probeAttackPower]
	if ap <= 2*errs[probeAttackPower] {
		return fmt.Errorf("level %d: attack power adds %.4f mobs a point (±%.4f), too close to nothing to compare with. Play more sessions, or the rotation doesn't fight", row.Level, ap, errs[probeAttackPower])
	}
	if less, ok := gains[probeHit+"-"]; ok {
		if less > gains[probeHit] {
			gains[probeHit], errs[probeHit] = less, errs[probeHit+"-"]
		}
		delete(gains, probeHit+"-")
		delete(errs, probeHit+"-")
	}
	row.MobsPerAP = round(ap, 4)
	row.Mobs, row.Deaths = round(row.Mobs, 2), round(row.Deaths, 2)
	row.Weights, row.Errors = map[string]float64{}, map[string]float64{}
	for name, gain := range gains {
		row.Weights[name] = round(max(0, gain/ap), 3)
		row.Errors[name] = round(errs[name]/ap, 3)
	}
	return nil
}

// pairedGain is what the second run's sessions kill more than the first's on average, and its
// standard error. Both runs play the same sessions, so we compare them session by session.
func pairedGain(base, with []float64) (float64, float64) {
	n := min(len(base), len(with))
	if n == 0 {
		return 0, 0
	}
	diffs := make([]float64, n)
	for i := range diffs {
		diffs[i] = with[i] - base[i]
	}
	m := meanErrorOf(diffs)
	return m.Mean, m.Error
}

// meanWeights is the mean of the rounds. The error is the larger of the sessions' noise and the
// rounds' spread, since a weight can move a lot from one round's gear to the next.
func meanWeights(rounds []*weightsFile) *weightsFile {
	out := &weightsFile{}
	n := float64(len(rounds))
	for l := range rounds[0].Levels {
		row := weightsRow{Level: rounds[0].Levels[l].Level, Weights: map[string]float64{}, Errors: map[string]float64{}}
		names := map[string]bool{}
		for _, r := range rounds {
			row.Mobs += r.Levels[l].Mobs / n
			row.Deaths += r.Levels[l].Deaths / n
			row.MobsPerAP += r.Levels[l].MobsPerAP / n
			for name := range r.Levels[l].Weights {
				names[name] = true
			}
		}
		row.Mobs, row.Deaths, row.MobsPerAP = round(row.Mobs, 2), round(row.Deaths, 2), round(row.MobsPerAP, 4)
		for name := range names {
			var values []float64
			squares := 0.0
			for _, r := range rounds {
				values = append(values, r.Levels[l].Weights[name])
				squares += r.Levels[l].Errors[name] * r.Levels[l].Errors[name]
			}
			m := meanErrorOf(values)
			row.Weights[name] = round(m.Mean, 3)
			row.Errors[name] = round(max(m.Error, math.Sqrt(squares)/n), 3)
		}
		out.Levels = append(out.Levels, row)
	}
	return out
}

// withWeights is the request judging the style's items by these weights.
func withWeights(req *proto.LevelingPlanSearchRequest, style proto.LevelingWeaponStyle, file *weightsFile) (*proto.LevelingPlanSearchRequest, error) {
	presets, err := file.presets(style)
	if err != nil {
		return nil, err
	}
	out := goproto.Clone(req).(*proto.LevelingPlanSearchRequest)
	out.GearWeights = slices.DeleteFunc(out.GearWeights, func(w *proto.LevelingGearWeights) bool { return w.Style == style })
	out.GearWeights = append(out.GearWeights, presets...)
	return out, nil
}

// measureWeights plays the rounds.
func measureWeights(ctx context.Context, db *proto.UIDatabase, req *proto.LevelingPlanSearchRequest, opts weightsOptions, progress func(round int, file *weightsFile)) (*weightsFile, error) {
	req = goproto.Clone(req).(*proto.LevelingPlanSearchRequest)
	if err := routeProblem(req.Plan, req.TalentShape); err != nil {
		return nil, err
	}
	// The style whose weights we measure. A plan with its own styles measures one set for all of
	// them, which is right only for a plan of one style.
	style := proto.LevelingWeaponStyle_LevelingWeaponStyleFree
	if opts.style != nil {
		style = *opts.style
		req.Plan.OneHandAndShield = false
		req.Plan.WeaponStyles = nil
		if style != proto.LevelingWeaponStyle_LevelingWeaponStyleFree {
			req.Plan.WeaponStyles = []*proto.LevelingWeaponStyleRange{{Level: 1, Style: style}}
		}
	}
	levels := weightLevels(req.Plan)
	if len(levels) == 0 {
		return nil, fmt.Errorf("the plan has no levels")
	}
	if len(req.GearWeights) == 0 {
		start, err := withWeights(req, style, &weightsFile{Levels: []weightsRow{{Level: levels[0], Weights: startWeights}}})
		if err != nil {
			return nil, err
		}
		req = start
	}
	var rounds []*weightsFile
	for i := range opts.rounds {
		if i > 0 {
			next, err := withWeights(req, style, meanWeights(rounds))
			if err != nil {
				return nil, err
			}
			req = next
		}
		file, err := weightsRound(ctx, db, req, opts, levels)
		if err != nil {
			return nil, err
		}
		rounds = append(rounds, file)
		progress(i+1, file)
	}
	return meanWeights(rounds), nil
}

// weightsOfFile measures a plan file's weights and writes them.
func weightsOfFile(db *proto.UIDatabase, in, out string, opts weightsOptions) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	req, err := readPlanRequest(in)
	if err != nil {
		return err
	}
	file, err := measureWeights(ctx, db, req, opts, func(round int, file *weightsFile) {
		mobs, deaths := 0.0, 0.0
		for _, row := range file.Levels {
			mobs += row.Mobs
			deaths += row.Deaths
		}
		fmt.Fprintf(os.Stderr, "round %d: %.1f mobs and %.2f deaths a session on average\n", round, mobs/float64(len(file.Levels)), deaths/float64(len(file.Levels)))
	})
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(file, "", "\t")
	if err != nil {
		return err
	}
	if out == "" {
		_, err = os.Stdout.Write(append(data, '\n'))
		return err
	}
	return os.WriteFile(out, append(data, '\n'), 0o644)
}

func round(x float64, digits int) float64 {
	scale := math.Pow(10, float64(digits))
	return math.Round(x*scale) / scale
}
