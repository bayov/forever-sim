package compsearch

import (
	"fmt"
	"sort"
	"strings"

	"github.com/wowsims/classic/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// The sensitivity matrix (docs/raid_sim.md, Sensitivity): the search again under each scenario that needs
// a re-run, and whether the winner holds. The scenarios the engine can't express yet (pet uptime, the
// unverified stacking rules, one Journeyman camp) aren't in it.

// MatrixScenario is one re-run.
type MatrixScenario struct {
	Name  string
	Why   string
	Camps bool // only for a camps-on result
	Apply func(req *proto.CompSearchRequest)
}

// Matrix is the scenarios for a request.
func Matrix(req *proto.CompSearchRequest) []MatrixScenario {
	out := []MatrixScenario{
		{Name: "120 s", Why: "Fury, rogues and the hunters fall faster than casters with length", Apply: func(r *proto.CompSearchRequest) {
			r.Encounter.Duration, r.Encounter.UseHealth = 120, false
		}},
		{Name: "300 s", Why: "hunters run out of mana by 300", Apply: func(r *proto.CompSearchRequest) {
			r.Encounter.Duration, r.Encounter.UseHealth = 300, false
		}},
		{Name: "raid damage 0 a minute", Why: "Demonology loses 63 DPS at 10 a minute; Fury 2H gains 56", Apply: func(r *proto.CompSearchRequest) {
			r.Encounter.SetRaidDamage, r.Encounter.RaidDamageHitsPerMinute = true, 0
		}},
		{Name: "raid damage 10 a minute", Why: "rage and procs from damage taken", Apply: func(r *proto.CompSearchRequest) {
			r.Encounter.SetRaidDamage, r.Encounter.RaidDamageHitsPerMinute = true, 10
		}},
		{Name: "no flask", Why: "the flask is worth 112 to a fire mage and 38 to a Beast Mastery hunter", Apply: func(r *proto.CompSearchRequest) {
			r.Encounter.NoFlask = true
		}},
		{Name: "Camp Chair keeps its physical crit under Moonkin Aura", Why: "how the chair's exclusion works is unverified", Camps: true,
			Apply: func(r *proto.CompSearchRequest) {
				if r.Assumptions == nil {
					r.Assumptions = &proto.RaidAssumptions{}
				}
				r.Assumptions.CampChairSpellCritOnly = true
			}},
		{Name: "camp buffs stack with the group versions", Why: "the tooltips name only the single-target spells", Camps: true,
			Apply: func(r *proto.CompSearchRequest) {
				if r.Assumptions == nil {
					r.Assumptions = &proto.RaidAssumptions{}
				}
				r.Assumptions.CampBuffsStackWithGroupBuffs = true
			}},
	}
	if req.Encounter.MobType == proto.MobType_MobTypeDragonkin {
		out = append(out, MatrixScenario{Name: "the generic boss in the flask zone", Why: "each raid's boss", Apply: func(r *proto.CompSearchRequest) {
			r.Encounter.MobType, r.Encounter.NaturalFlaskZone, r.Encounter.Name = proto.MobType_MobTypeUnknown, true, "Generic"
		}})
	} else {
		out = append(out, MatrixScenario{Name: "Onyxia (Dragonkin, outside the flask zone)", Why: "off the zone the flasks give only Stamina, unevenly",
			Apply: func(r *proto.CompSearchRequest) {
				r.Encounter.MobType, r.Encounter.NaturalFlaskZone, r.Encounter.Name = proto.MobType_MobTypeDragonkin, false, "Onyxia"
			}})
	}
	return out
}

// compIdentity is a result comp's members and placeholders, in a stable order.
func compIdentity(c *proto.CompResult) string {
	var ids []string
	for _, m := range c.Members {
		id := m.BuildId
		if len(m.Duties) > 0 {
			id += "+" + m.Duties[0].String()
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return strings.Join(ids, ",")
}

// RunMatrix runs the matrix's scenarios and adds, to each of the base result's scenarios, whether its
// winner holds under each.
func RunMatrix(pool *Pool, sim *Simmer, req *proto.CompSearchRequest, base *proto.CompSearchResult, progress Progress) error {
	for _, ms := range Matrix(req) {
		r := googleProto.Clone(req).(*proto.CompSearchRequest)
		r.SkipSensitivity, r.Confirm = true, false
		ms.Apply(r)
		if ms.Camps {
			r.CampBuffs = []bool{true}
		}
		hasCamps := false
		for _, c := range req.CampBuffs {
			hasCamps = hasCamps || c
		}
		if ms.Camps && !hasCamps {
			continue
		}
		res, err := Search(pool, sim, r, progress)
		if err != nil {
			return fmt.Errorf("%s: %w", ms.Name, err)
		}
		for _, bs := range base.Scenarios {
			if ms.Camps && !bs.CampBuffs || len(bs.Comps) == 0 {
				continue
			}
			for _, rs := range res.Scenarios {
				if rs.Faction != bs.Faction || rs.CampBuffs != bs.CampBuffs || len(rs.Comps) == 0 {
					continue
				}
				want := compIdentity(bs.Comps[0])
				holds := false
				for i := 0; i < int(max(1, rs.TieSet)) && i < len(rs.Comps); i++ {
					if compIdentity(rs.Comps[i]) == want {
						holds = true
					}
				}
				s := &proto.CompSensitivity{Scenario: "re-run: " + ms.Name, WinnerHolds: holds, Note: ms.Why}
				if !holds {
					w := googleProto.Clone(rs.Comps[0]).(*proto.CompResult)
					w.BuffSet = compText(w)
					s.Winners = append(s.Winners, w)
				}
				bs.Sensitivity = append(bs.Sensitivity, s)
			}
		}
		base.SimsRun += res.SimsRun
		base.CacheHits += res.CacheHits
	}
	return nil
}
