// Package healscore sims healers healing a raid that takes damage, and scores them the way
// docs/healing_sim.md settles it: the damage they carry, in a role. tools/healbench is its command
// line, and rotopt's healer objective (-heal-roles) scores gear with it.
//
// A Bench builds a raid around a healer: the tanks (DefaultTank, the tank bench's warrior), copies of
// the healer, and fillers (DefaultFiller, the DPS report's Frost mage). It turns on the healing fight
// (Encounter.track_health), so every player loses health and can die, and nobody has the stand-in
// healing model. The role picks what the healers face and what fails:
//
//   - tank: the boss's melee and its abilities on the tanks, nothing on the raid; the healers are
//     assigned the tanks; a fight fails when a tank dies. The line is 2% of fights.
//   - raid: only the damage on the raid (no boss melee, no tanks); a fight fails when anyone dies.
//     The line is 5%.
//   - both: everything; a tank death over 2% or another death over 5% fails.
//
// Carry finds the damage carried: the scale of all the damage at which the deaths reach their line,
// by bisection on paired seeds and interpolation between the two closest scales.
package healscore

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/tools/tankscore"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

// The death lines (the user's ruling, 2026-11-02).
const (
	TankLine  = 0.02
	OtherLine = 0.05
)

// The raid around the healer, healbench's -tank and -filler defaults.
const (
	DefaultTank   = "tools/tankbench/req/warrior_protection.json"
	DefaultFiller = "tools/dps_report/req/mage__frost.json"
)

// DefaultSeed is healbench's -seed, the report's HEAL_SEED (tools/dps_report/tankheal.py).
const DefaultSeed = 101

// Options are the bench's settings, healbench's flags.
type Options struct {
	StopCast   float64
	Reaction   float64
	Role       string
	Scale      float64
	Healers    int
	Fillers    int
	Iterations int
	Duration   float64
	Seed       int64
	Labeled    bool
	AllValues  bool
}

// DefaultOptions are healbench's defaults for a role, with the healers its default for the role and
// the duration 0 (SetDuration fills it in).
func DefaultOptions(role string) Options {
	return Options{StopCast: 0.5, Reaction: 0.3, Role: role, Scale: 1, Healers: DefaultHealers(role), Fillers: 7,
		Iterations: 1000, Seed: DefaultSeed}
}

// DefaultHealers is the healers a role has by default: 1 for tank, 2 otherwise.
func DefaultHealers(role string) int {
	if role == "tank" {
		return 1
	}
	return 2
}

// Fight is the damage: the boss with its melee, and the events on the raid.
type Fight struct {
	Name, Description string
	Tanks             int
	Encounter         *proto.Encounter
	// The median kill in the logs (the file's raidHealing.fightSeconds), 0 when it has none.
	KillSeconds float64
}

// LoadFight is a boss file of tools/encounters as a fight.
func LoadFight(path string) (Fight, error) {
	enc, err := tankscore.LoadEncounter(path)
	if err != nil {
		return Fight{}, err
	}
	f := EncounterFight(enc)
	f.KillSeconds = KillSeconds(path)
	return f, nil
}

// Duration is the fight length a role is judged at, the boss's own (the user's ruling): for role
// raid and both the logs' median kill, for tank the file's duration, the main tank's time on the
// boss (Onyxia's first phase, one tank's stint on Vaelastrasz).
func (f Fight) Duration(role string) float64 {
	if role != "tank" && f.KillSeconds > 0 {
		return f.KillSeconds
	}
	return f.Encounter.Duration
}

// KillSeconds is an encounter file's median kill length from the logs (raidHealing.fightSeconds,
// the 10th, 50th and 90th percentile), or 0.
func KillSeconds(path string) float64 {
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

// EncounterFight is a loaded encounter as a fight, with no kill length.
func EncounterFight(enc *tankscore.Encounter) Fight {
	f := Fight{Name: enc.Name, Description: enc.Notes, Encounter: enc.Encounter, Tanks: 1}
	for _, t := range enc.Encounter.Targets {
		if t.TankSwapSeconds > 0 {
			f.Tanks = 2
		}
	}
	return f
}

// onTanks is whether an event only ever hits tanks.
func onTanks(e *proto.RaidDamageEvent) bool {
	return e.Who == proto.RaidDamageEvent_Tanks || e.Who == proto.RaidDamageEvent_BossTarget
}

// Role is who a raid slot holds.
type Role int

const (
	RoleTank Role = iota
	RoleHealer
	RoleFiller
)

// Bench is a healer in a fight, with the raid around it.
type Bench struct {
	O                    Options
	Fight                Fight
	Tank, Healer, Filler *proto.Player
	Buffs                *proto.RaidBuffs
	Debuffs              *proto.Debuffs
	// Slots is who each raid slot holds, by party, and TankIndexes the tanks' raid indexes, as the
	// last Request built them.
	Slots       [][]Role
	TankIndexes []int32
}

// Request builds the raid and the fight at a scale of the damage. With floor, every player is
// topped off by the stand-in heal stream instead of the healers healing.
func (b *Bench) Request(scale float64, floor bool) (*proto.RaidSimRequest, error) {
	tanks := b.Fight.Tanks
	fillers := b.O.Fillers
	if b.O.Role == "raid" {
		tanks = 0
	}
	// The tank role keeps the fillers, idle (below): a raid around the tank, so a heal that jumps
	// (Prayer of Mending) has the raid to land on, not only the tank and the healer
	// (healer_fairness-20261104-1). They take no damage there, since the role drops the raid's.

	// The tanks go in the first party, the healers one to a party in turn, and the fillers in the
	// slots left, so Chain Heal and the party heals see a raid's spread.
	total := tanks + b.O.Healers + fillers
	if total > 40 {
		return nil, fmt.Errorf("%d players, 40 at most", total)
	}
	numParties := (total + 4) / 5
	b.Slots = make([][]Role, numParties)
	place := func(r Role, party int) {
		for len(b.Slots[party]) == 5 {
			party = (party + 1) % numParties
		}
		b.Slots[party] = append(b.Slots[party], r)
	}
	for i := 0; i < tanks; i++ {
		place(RoleTank, 0)
	}
	for i := 0; i < b.O.Healers; i++ {
		place(RoleHealer, i%numParties)
	}
	for i := 0; i < fillers; i++ {
		place(RoleFiller, 0)
	}

	raid := &proto.Raid{Buffs: b.Buffs, Debuffs: b.Debuffs}
	b.TankIndexes = nil
	for pi, party := range b.Slots {
		raid.Parties = append(raid.Parties, &proto.Party{})
		for si, r := range party {
			if r == RoleTank {
				idx := int32(pi*5 + si)
				b.TankIndexes = append(b.TankIndexes, idx)
				raid.Tanks = append(raid.Tanks, &proto.UnitReference{Type: proto.UnitReference_Player, Index: idx})
			}
		}
	}
	healerNumber := 0
	for pi, party := range b.Slots {
		for si, r := range party {
			player := clonePlayer(map[Role]*proto.Player{RoleTank: b.Tank, RoleHealer: b.Healer, RoleFiller: b.Filler}[r])
			if r == RoleTank {
				// A tank faces the boss: the boss parries and blocks it, and its parries haste the swing.
				player.InFrontOfTarget = true
			}
			if r == RoleHealer {
				// Each healer reacts a little differently, so identical healers don't act in lockstep.
				spread := 0.0
				if b.O.Healers > 1 {
					spread = -0.25 + 0.5*float64(healerNumber)/float64(b.O.Healers-1)
				}
				player.ReactionTimeMs = int32(1000 * b.O.Reaction * (1 + spread))
				player.StopCastOverheal = b.O.StopCast
				healerNumber++
			}
			player.HealingModel = nil
			if r == RoleFiller && b.O.Role == "tank" {
				player.Rotation = &proto.APLRotation{Type: proto.APLRotation_TypeAPL}
			}
			player.Name = fmt.Sprintf("Player %d", pi*5+si+1)
			if r == RoleHealer && b.O.Role == "tank" {
				for _, idx := range b.TankIndexes {
					player.HealAssignment = append(player.HealAssignment, &proto.UnitReference{Type: proto.UnitReference_Player, Index: idx})
				}
			}
			if floor {
				// A full heal every 1.5 ± 0.5 s: only a one-shot or a burst between heals kills.
				player.HealingModel = &proto.HealingModel{Hps: 50000, CadenceSeconds: 1.5, CadenceVariation: 0.5}
				if r == RoleHealer {
					player.Rotation = &proto.APLRotation{Type: proto.APLRotation_TypeAPL}
				}
			}
			raid.Parties[pi].Players = append(raid.Parties[pi].Players, player)
		}
	}
	raid.NumActiveParties = int32(len(raid.Parties))

	encounter := googleProto.Clone(b.Fight.Encounter).(*proto.Encounter)
	encounter.Duration = b.O.Duration
	encounter.DurationVariation = 0
	encounter.TrackHealth = true
	encounter.RaidDamageHitsPerMinute = 0
	encounter.RaidDamage = nil
	for _, e := range b.Fight.Encounter.RaidDamage {
		if (b.O.Role == "tank" && !onTanks(e)) || (b.O.Role == "raid" && onTanks(e)) {
			continue
		}
		encounter.RaidDamage = append(encounter.RaidDamage, googleProto.Clone(e).(*proto.RaidDamageEvent))
	}
	for _, target := range encounter.Targets {
		if b.O.Role == "raid" {
			target.TankIndex = -1
			target.TankSwapSeconds = 0
		}
	}
	ScaleDamage(encounter, scale)

	return &proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{Iterations: int32(b.O.Iterations), RandomSeed: b.O.Seed, UseLabeledRands: b.O.Labeled, SaveAllValues: b.O.AllValues},
	}, nil
}

// ScaleDamage multiplies all of an encounter's damage, the boss's melee too, by scale. A request
// Request built at scale 1 and scaled here is the one it builds at that scale.
func ScaleDamage(encounter *proto.Encounter, scale float64) {
	for _, e := range encounter.RaidDamage {
		e.MinDamage *= scale
		e.MaxDamage *= scale
	}
	for _, target := range encounter.Targets {
		target.MinBaseDamage *= scale
	}
}

// Result is what a sim tells about the healers and the raid.
type Result struct {
	Hps, Ehps, Overheal, Healing, ManaSpent, ManaEnd, Tto float64
	TankDtps, RaidDtps, TankScares, RaidScares            float64
	TankDied, OtherDied, AnyDied, FirstDeath              float64
	Duration                                              float64
	// Iterations is the fights the sim ran.
	Iterations int
}

// Run sims the bench at a scale of the damage.
func (b *Bench) Run(scale float64, floor bool) (Result, error) {
	req, err := b.Request(scale, floor)
	if err != nil {
		return Result{}, err
	}
	res := core.RunRaidSimConcurrent(req)
	if res.Error != nil {
		return Result{}, fmt.Errorf("%s", res.Error.Message)
	}
	return b.Read(res), nil
}

// Read is the result of a sim of a request the bench built.
func (b *Bench) Read(res *proto.RaidSimResult) Result {
	r := Result{Duration: res.AvgIterationDuration}
	healers := float64(b.O.Healers)
	for pi, party := range res.RaidMetrics.Parties {
		for si, m := range party.Players {
			who := b.Slots[pi][si]
			r.RaidDtps += m.Dtps.Avg
			r.RaidScares += m.ScaresAvg
			switch who {
			case RoleTank:
				r.TankDtps += m.Dtps.Avg
				r.TankScares += m.ScaresAvg
			case RoleHealer:
				r.Hps += m.Hps.Avg
				r.Ehps += m.Ehps.Avg
				r.ManaSpent += m.ManaSpentAvg
				r.ManaEnd += m.ManaEnd.Avg / healers
				r.Tto += m.Tto.Avg / healers
				for _, a := range m.Actions {
					for _, t := range a.Targets {
						r.Healing += t.Healing
						r.Overheal += t.Overhealing
					}
				}
			}
		}
	}
	if hf := res.RaidMetrics.HealingFight; hf != nil && hf.Fights > 0 {
		n := float64(hf.Fights)
		r.Iterations = int(hf.Fights)
		r.TankDied = float64(hf.FightsTankDeath) / n
		r.OtherDied = float64(hf.FightsOtherDeath) / n
		r.AnyDied = float64(hf.FightsAnyDeath) / n
		if hf.FightsAnyDeath > 0 {
			r.FirstDeath = hf.FirstDeathSecondsSum / float64(hf.FightsAnyDeath)
		}
	}
	return r
}

// Failure is how far over its lines a sim's deaths are: 1 is on the line.
func Failure(role string, r Result) float64 {
	switch role {
	case "tank":
		return r.TankDied / TankLine
	case "raid":
		return r.AnyDied / OtherLine
	default:
		return max(r.TankDied/TankLine, r.OtherDied/OtherLine)
	}
}

// Failure is how far over its lines a sim's deaths are: 1 is on the line.
func (b *Bench) Failure(r Result) float64 { return Failure(b.O.Role, r) }

// Line is a role's death line: the share of fights that fail at a failure of 1. Role both has two,
// and gives the tanks'.
func Line(role string) float64 {
	if role == "raid" {
		return OtherLine
	}
	return TankLine
}

// Failed is whether a fight failed in a role, from its deaths (HealingFightMetrics.fight_deaths:
// 1 a tank, 2 another, 3 both). For role both it is any death.
func Failed(role string, deaths int32) bool {
	if role == "tank" {
		return deaths&1 != 0
	}
	return deaths != 0
}

// Carry is the damage carried and the bracket it was found in: the two closest scales the search
// simmed, on either side of the line, and their failures.
type Carry struct {
	K              float64
	Lo, Hi         float64
	FailLo, FailHi float64
}

// Slope is the failure's change per unit of scale between the bracket's two scales.
func (c Carry) Slope() float64 { return (c.FailHi - c.FailLo) / (c.Hi - c.Lo) }

// Carry is the damage scale at which the deaths reach their line: bracketed by doubling or halving,
// narrowed by bisection, then interpolated between the two closest scales.
func (b *Bench) Carry() (Carry, error) {
	return CarrySearch(func(k float64) (float64, error) {
		r, err := b.Run(k, false)
		return b.Failure(r), err
	})
}

// CarrySearch is Carry's search on a failure function of the scale. It stops at the first error.
func CarrySearch(failure func(float64) (float64, error)) (Carry, error) {
	var err error
	f := func(k float64) float64 {
		if err != nil {
			return 0
		}
		var fk float64
		fk, err = failure(k)
		return fk
	}
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
	if err != nil {
		return Carry{}, err
	}
	return Interpolate(lo, hi, flo, fhi), nil
}

// Interpolate is the scale where the failure reaches 1 on the line through two scales on either
// side of it, or lo when the failure doesn't rise between them.
func Interpolate(lo, hi, flo, fhi float64) Carry {
	c := Carry{K: lo, Lo: lo, Hi: hi, FailLo: flo, FailHi: fhi}
	if fhi > flo && !math.IsInf(fhi, 0) {
		c.K = lo + (hi-lo)*(1-flo)/(fhi-flo)
	}
	return c
}

// ReadRequest reads a RaidSimRequest JSON.
func ReadRequest(path string) (*proto.RaidSimRequest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return rsr, nil
}

// FirstPlayer is a request's first player, the one a bench request is for.
func FirstPlayer(rsr *proto.RaidSimRequest) *proto.Player {
	return rsr.Raid.Parties[0].Players[0]
}

func clonePlayer(p *proto.Player) *proto.Player {
	return googleProto.Clone(p).(*proto.Player)
}
