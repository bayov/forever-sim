// Package compsearch is the raid sim's comp search (docs/raid_sim.md, Comp search): which DPS builds, how
// many of each, and which tank and healer presets give a raid the most DPS. It judges each build on its
// own under a set of buffs and debuffs, with single-player sims it caches, and confirms the best sets
// with whole-raid sims.
package compsearch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

// Build is one of the DPS Report's builds (tools/dps_report/req): a base build, a duty variant (the base
// build holding an assignment) or an encounter variant (the base build against one creature type).
type Build struct {
	ID        string
	Class     proto.Class
	ClassName string
	Spec      string
	Name      string
	Duty      proto.Duty
	BaseID    string        // for a variant
	MobType   proto.MobType // an encounter variant's creature type
	Published float64       // the report's published DPS
	Race      map[proto.Faction]proto.Race
	Req       *proto.RaidSimRequest

	// Filled in by the pool.
	Base   *Build                   // self for a base build
	Duties map[proto.Duty]*Build    // a base build's duty variants
	Gives  Gift                     // what it gives the raid without a duty
	Pkg    proto.RaidPackage        // the party package a shaman gives
	Casts  Gift                     // what its own rotation casts on itself (left off its own request)
	Vs     map[proto.MobType]*Build // a base build's encounter variants
	Pets   []*Build                 // a base build's pet variants (a Tallstrider, a wolf), builds of their own
}

func (b *Build) IsBase() bool { return b.BaseID == "" }

func (b *Build) String() string { return b.ID }

// Pool is every build the search can use.
type Pool struct {
	Root   string
	Builds map[string]*Build
	Bases  []*Build // the base builds, sorted by id
}

type buildMeta struct {
	ID        string          `json:"id"`
	Class     string          `json:"class"`
	Spec      string          `json:"spec"`
	Build     string          `json:"build"`
	Duty      string          `json:"duty"`
	BaseBuild string          `json:"baseBuild"`
	Encounter string          `json:"encounter"`
	Published json.RawMessage `json:"published"`
}

type reportBuild struct {
	ID           string `json:"id"`
	HordeRace    string `json:"hordeRace"`
	AllianceRace string `json:"allianceRace"`
}

var dutyNames = map[string]proto.Duty{
	"Battle Shout":              proto.Duty_DutyBattleShout,
	"Faerie Fire":               proto.Duty_DutyFaerieFire,
	"Curse of the Elements":     proto.Duty_DutyCurseOfTheElements,
	"Curse of Recklessness":     proto.Duty_DutyCurseOfRecklessness,
	"Judgement of the Crusader": proto.Duty_DutyJudgementOfTheCrusader,
	"Judgement of Wisdom":       proto.Duty_DutyJudgementOfWisdom,
	"Sunder Armor":              proto.Duty_DutySunderUpkeep,
}

var mobTypes = map[string]proto.MobType{
	"Humanoid":  proto.MobType_MobTypeHumanoid,
	"Undead":    proto.MobType_MobTypeUndead,
	"Dragonkin": proto.MobType_MobTypeDragonkin,
	"Demon":     proto.MobType_MobTypeDemon,
	"Beast":     proto.MobType_MobTypeBeast,
}

// The report's race names (builds.json) as proto races.
var raceNames = map[string]proto.Race{
	"Human":      proto.Race_RaceHuman,
	"Dwarf":      proto.Race_RaceDwarf,
	"Night Elf":  proto.Race_RaceNightElf,
	"Gnome":      proto.Race_RaceGnome,
	"High Order": proto.Race_RaceSkyborneHighOrder,
	"Orc":        proto.Race_RaceOrc,
	"Undead":     proto.Race_RaceUndead,
	"Tauren":     proto.Race_RaceTauren,
	"Troll":      proto.Race_RaceTroll,
	"Windshaper": proto.Race_RaceSkyborneWindshaper,
}

// LoadPool reads the report's requests and races from a forever-sim checkout.
func LoadPool(root string) (*Pool, error) {
	sim.RegisterAll()
	reqDir := filepath.Join(root, "tools", "dps_report", "req")
	metas, err := filepath.Glob(filepath.Join(reqDir, "*.meta.json"))
	if err != nil || len(metas) == 0 {
		return nil, fmt.Errorf("no requests in %s", reqDir)
	}
	races := map[string]reportBuild{}
	if data, err := os.ReadFile(filepath.Join(root, "ui", "dps_report", "data", "builds.json")); err == nil {
		var list []reportBuild
		if err := json.Unmarshal(data, &list); err != nil {
			return nil, fmt.Errorf("builds.json: %w", err)
		}
		for _, r := range list {
			races[r.ID] = r
		}
	}

	pool := &Pool{Root: root, Builds: map[string]*Build{}}
	for _, mf := range metas {
		data, err := os.ReadFile(mf)
		if err != nil {
			return nil, err
		}
		var m buildMeta
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("%s: %w", mf, err)
		}
		reqData, err := os.ReadFile(strings.TrimSuffix(mf, ".meta.json") + ".json")
		if err != nil {
			return nil, err
		}
		req := &proto.RaidSimRequest{}
		if err := protojson.Unmarshal(reqData, req); err != nil {
			return nil, fmt.Errorf("%s: %w", m.ID, err)
		}
		var pub struct {
			DPS float64 `json:"dps"`
		}
		_ = json.Unmarshal(m.Published, &pub)
		player := req.Raid.Parties[0].Players[0]
		b := &Build{
			ID:        m.ID,
			Class:     player.Class,
			ClassName: m.Class,
			Spec:      m.Spec,
			Name:      m.Build,
			BaseID:    m.BaseBuild,
			Published: pub.DPS,
			Req:       req,
			Race:      map[proto.Faction]proto.Race{},
			Duties:    map[proto.Duty]*Build{},
			Vs:        map[proto.MobType]*Build{},
		}
		if m.Duty != "" {
			d, ok := dutyNames[m.Duty]
			if !ok {
				return nil, fmt.Errorf("%s: unknown duty %q", m.ID, m.Duty)
			}
			b.Duty = d
		}
		if m.Encounter != "" {
			t, ok := mobTypes[m.Encounter]
			if !ok {
				return nil, fmt.Errorf("%s: unknown encounter %q", m.ID, m.Encounter)
			}
			b.MobType = t
		}
		if r, ok := races[m.ID]; ok {
			b.Race[proto.Faction_Horde] = raceNames[r.HordeRace]
			b.Race[proto.Faction_Alliance] = raceNames[r.AllianceRace]
		}
		for f, r := range b.Race {
			if r == proto.Race_RaceUnknown {
				delete(b.Race, f)
			}
		}
		pool.Builds[b.ID] = b
	}

	for _, b := range pool.Builds {
		if b.IsBase() {
			b.Base = b
			pool.Bases = append(pool.Bases, b)
			continue
		}
		base, ok := pool.Builds[b.BaseID]
		if !ok {
			return nil, fmt.Errorf("%s: base build %s missing", b.ID, b.BaseID)
		}
		b.Base = base
		if b.Duty != proto.Duty_DutyNone {
			base.Duties[b.Duty] = b
		} else if b.MobType != proto.MobType_MobTypeUnknown {
			base.Vs[b.MobType] = b
		} else {
			// A variant that's neither a duty nor an encounter is a build of its own in the pool, like a
			// hunter with another pet: it fills and provides on its own terms.
			base.Pets = append(base.Pets, b)
		}
	}
	slices.SortFunc(pool.Bases, func(a, b *Build) int { return strings.Compare(a.ID, b.ID) })
	for _, b := range pool.Bases {
		slices.SortFunc(b.Pets, func(a, b *Build) int { return strings.Compare(a.ID, b.ID) })
	}
	for _, b := range pool.Builds {
		b.Gives, b.Pkg, b.Casts = buildGifts(b)
		// A variant without its own race takes its base build's.
		for _, f := range []proto.Faction{proto.Faction_Horde, proto.Faction_Alliance} {
			if _, ok := b.Race[f]; !ok {
				if r, ok := b.Base.Race[f]; ok {
					b.Race[f] = r
				} else {
					b.Race[f] = player0(b.Req).Race
				}
			}
		}
	}
	return pool, nil
}

func player0(req *proto.RaidSimRequest) *proto.Player {
	return req.Raid.Parties[0].Players[0]
}

// ForBoss is the build to use against a creature type: its encounter variant when it has one.
func (b *Build) ForBoss(mob proto.MobType) *Build {
	if v, ok := b.Base.Vs[mob]; ok && b.IsBase() {
		return v
	}
	return b
}

// WithDuty is the build holding a duty: its duty variant, built from its encounter variant's base.
func (b *Build) WithDuty(d proto.Duty) *Build {
	if d == proto.Duty_DutyNone {
		return b
	}
	return b.Base.Duties[d]
}

// Player is the build's player at its best race in a faction, with that race's own rotation and gear
// tuning from req/races (as tools/dps_report/race_override.py applies them).
func (p *Pool) Player(b *Build, faction proto.Faction) (*proto.Player, error) {
	player := googleProto.Clone(player0(b.Req)).(*proto.Player)
	race, ok := b.Race[faction]
	if !ok || race == player.Race {
		return player, nil
	}
	player.Race = race
	ov, err := p.raceOverride(b, race)
	if err != nil || ov == nil {
		return player, err
	}
	if ov.Rotation != nil {
		player.Rotation = ov.Rotation
	}
	if len(ov.BonusStats) > 0 {
		if player.BonusStats == nil {
			player.BonusStats = &proto.UnitStats{}
		}
		for name, v := range ov.BonusStats {
			if i, ok := proto.Stat_value["Stat"+name]; ok {
				for len(player.BonusStats.Stats) <= int(i) {
					player.BonusStats.Stats = append(player.BonusStats.Stats, 0)
				}
				player.BonusStats.Stats[i] += v
				continue
			}
			i, ok := proto.PseudoStat_value["PseudoStat"+name]
			if !ok {
				i, ok = proto.PseudoStat_value[name]
			}
			if !ok {
				return nil, fmt.Errorf("%s: unknown stat %q in a race override", b.ID, name)
			}
			for len(player.BonusStats.PseudoStats) <= int(i) {
				player.BonusStats.PseudoStats = append(player.BonusStats.PseudoStats, 0)
			}
			player.BonusStats.PseudoStats[i] += v
		}
	}
	return player, nil
}

type raceOverride struct {
	Rotation   *proto.APLRotation
	BonusStats map[string]float64
}

func raceFileName(race proto.Race) string {
	return strings.TrimPrefix(race.String(), "Race")
}

// raceOverride reads req/races/<id>.<Race>.json, or for a variant its base build's file, keeping only
// the parts that fit the variant: the rotation when the variant's rotation is the base's, the bonus
// stats when its gear is the base's.
func (p *Pool) raceOverride(b *Build, race proto.Race) (*raceOverride, error) {
	read := func(id string) (*raceOverride, error) {
		data, err := os.ReadFile(filepath.Join(p.Root, "tools", "dps_report", "req", "races", id+"."+raceFileName(race)+".json"))
		if err != nil {
			return nil, nil
		}
		var raw struct {
			Rotation   json.RawMessage    `json:"rotation"`
			BonusStats map[string]float64 `json:"bonusStats"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("%s %s: %w", id, race, err)
		}
		ov := &raceOverride{BonusStats: raw.BonusStats}
		if len(raw.Rotation) > 0 && string(raw.Rotation) != "null" {
			ov.Rotation = &proto.APLRotation{}
			if err := protojson.Unmarshal(raw.Rotation, ov.Rotation); err != nil {
				return nil, fmt.Errorf("%s %s rotation: %w", id, race, err)
			}
		}
		return ov, nil
	}
	ov, err := read(b.ID)
	if ov != nil || err != nil || b.IsBase() {
		return ov, err
	}
	ov, err = read(b.BaseID)
	if ov == nil || err != nil {
		return ov, err
	}
	mine, theirs := player0(b.Req), player0(b.Base.Req)
	if !googleProto.Equal(mine.Rotation, theirs.Rotation) {
		ov.Rotation = nil
	}
	if !googleProto.Equal(mine.Equipment, theirs.Equipment) {
		ov.BonusStats = nil
	}
	if ov.Rotation == nil && len(ov.BonusStats) == 0 {
		return nil, nil
	}
	return ov, nil
}
