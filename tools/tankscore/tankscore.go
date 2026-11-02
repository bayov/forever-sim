// Package tankscore is the tank optimizer's objective (docs/tank_sim.md, The optimizer's
// objective), shared by tools/tankbench and rotopt's -tank: the encounter files
// (tools/encounters), setting a request up to score its first player as the tank, and the score.
//
// The score is in raid DPS: threat members × SafeDPS − w × DPS slot × healing needed / one healer's
// HPS. A healer on the tank costs one DPS slot, and each point of SafeDPS frees the DPS of the raid
// members close to the tank's threat. The presets weigh survival by w: Max threat 1/3, Balanced 1,
// Max survival 3 (the user's ruling, 2026-11-02).
package tankscore

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

// Presets are the survival weights of the three tank presets.
var Presets = map[string]float64{"max-threat": 1.0 / 3, "balanced": 1, "max-survival": 3}

// Encounter is a tank encounter file (tools/encounters/README.md).
type Encounter struct {
	Name    string          `json:"name"`
	Source  string          `json:"source"`
	Notes   string          `json:"notes"`
	Raw     json.RawMessage `json:"encounter"`
	Healing *Healing        `json:"healing"`

	Encounter *proto.Encounter `json:"-"`
}

// Healing is the healers a raid puts on its tank: how many, each one's healing per second on it,
// and their heals' cast time.
type Healing struct {
	Healers     float64 `json:"healers"`
	HpsEach     float64 `json:"hpsEach"`
	CastSeconds float64 `json:"castSeconds"`
	Notes       string  `json:"notes"`
}

// Model is the stand-in stream for the healers: their HPS added up, a heal every cast time over
// the healers, each about one healer's heal, the timing spread by half the cadence.
func (h Healing) Model() *proto.HealingModel {
	cadence := h.CastSeconds / h.Healers
	return &proto.HealingModel{Hps: h.Healers * h.HpsEach, CadenceSeconds: cadence, CadenceVariation: cadence / 2, BurstWindow: 6}
}

func LoadEncounter(path string) (*Encounter, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	enc := &Encounter{}
	if err := json.Unmarshal(data, enc); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	enc.Encounter = &proto.Encounter{}
	if err := protojson.Unmarshal(enc.Raw, enc.Encounter); err != nil {
		return nil, fmt.Errorf("%s: encounter: %w", path, err)
	}
	if len(enc.Encounter.Targets) == 0 {
		return nil, fmt.Errorf("%s: the encounter has no boss", path)
	}
	return enc, nil
}

// UseEncounter puts the encounter's fight in the request, keeping what the file leaves unset of the
// request's own fight setup (the execute phases, the natural flask zone).
func UseEncounter(request *proto.RaidSimRequest, enc *Encounter) {
	fight := googleProto.Clone(enc.Encounter).(*proto.Encounter)
	if old := request.Encounter; old != nil {
		if fight.ExecuteProportion_20 == 0 && fight.ExecuteProportion_25 == 0 && fight.ExecuteProportion_35 == 0 {
			fight.ExecuteProportion_20 = old.ExecuteProportion_20
			fight.ExecuteProportion_25 = old.ExecuteProportion_25
			fight.ExecuteProportion_35 = old.ExecuteProportion_35
		}
		fight.NaturalFlaskZone = fight.NaturalFlaskZone || old.NaturalFlaskZone
	}
	request.Encounter = fight
}

// MakeTank makes the request's first player the tank every boss attacks, healed by the stand-in
// stream and scored.
func MakeTank(request *proto.RaidSimRequest, healing *proto.HealingModel, scoring *proto.TankScoring) {
	request.Raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	for _, target := range request.Encounter.Targets {
		target.TankIndex = 0
	}
	tank := request.Raid.Parties[0].Players[0]
	tank.HealingModel = googleProto.Clone(healing).(*proto.HealingModel)
	// An HPS of 0 asks the engine for its presim healer.
	tank.HealingModel.Hps = max(tank.HealingModel.Hps, 1e-9)
	tank.TankScoring = scoring
}

// Objective is how the score is weighed.
type Objective struct {
	Weight        float64 // w, the survival weight
	ThreatMembers float64 // raid members close to the tank's threat
	DpsSlot       float64 // a raid DPS member's DPS
	HealerEach    float64 // one healer's HPS on the tank
	Death         float64 // the chance of death healing needed aims at (5% for tanks)
}

// Score is a tank's score from its metrics.
type Score struct {
	Value, StdErr          float64
	SafeDps, SafeDpsErr    float64
	Needed, NeededErr      float64 // healing needed, -1 when no healing saves it
	DeathFloor, Iterations float64
}

// Rejected is the score of a tank no healing saves at the death share: far below any real score,
// and lower the more often it dies, so a search still climbs out of it.
const Rejected = -1e6

func (o Objective) Score(m *proto.UnitMetrics, iterations int32) Score {
	ts := m.GetTankScore()
	s := Score{Iterations: float64(iterations)}
	if ts == nil {
		return Score{Value: math.NaN()}
	}
	s.Needed, s.NeededErr, s.DeathFloor = core.TankScoreSummary(ts, o.Death)
	s.SafeDps = ts.SafeDps.GetAvg()
	s.SafeDpsErr = ts.SafeDps.GetStdev() / math.Sqrt(float64(iterations))
	if s.Needed < 0 {
		s.Value = Rejected * (1 + s.DeathFloor)
		return s
	}
	perHps := o.Weight * o.DpsSlot / o.HealerEach
	s.Value = o.ThreatMembers*s.SafeDps - perHps*s.Needed
	s.StdErr = math.Hypot(o.ThreatMembers*s.SafeDpsErr, perHps*s.NeededErr)
	return s
}
