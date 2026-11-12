package sim

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"google.golang.org/protobuf/encoding/protojson"
)

// windfuryLogList is the user's log audit of Nov 6-11 (tools/logcheck/findings.md, "Windfury
// Totem procs from main hand abilities"): the abilities that proc Windfury Totem and those that
// never do, by the spell IDs the sim's builds register them under.
var windfuryLogList = []struct {
	name  string
	id    int32
	procs bool
}{
	{"Heroic Strike", 11567, true},   // 21/109
	{"Cleave", 20569, true},          // 64/308
	{"Maul", 9881, true},             // 32/174
	{"Raptor Strike", 14266, true},   // 16/127
	{"Sunder Armor", 11597, true},    // 57/428
	{"Revenge", 11601, true},         // 33/257
	{"Execute", 20662, true},         // 20/186
	{"Rend", 11574, true},            // 4/39
	{"Sinister Strike", 11294, true}, // 178/1610
	{"Eviscerate", 11300, true},      // 18/206
	{"Wing Clip", 14268, true},       // 6/37
	{"Mongoose Bite", 14271, true},   // 3/32
	{"Stormstrike", 17364, false},    // 0/4467
	{"Holy Strike", 10333, false},    // 0/1175
	{"Bloodthirst", 23894, false},    // 0/258
	{"Whirlwind", 1680, false},       // 0/292
	{"Claw", 9850, false},            // 0/165
	{"Swipe", 9908, false},           // 0/1634
	{"Rake", 9904, false},            // 0/91
	{"Seal of Command", 20424, false},
	{"Seal of Righteousness", 25742, false},
	{"Judgement of Command", 20947, false},
	{"Windfury Weapon", 16362, false},
}

// windfuryLogClientOnly are logged abilities the report's builds don't register (or the sim
// doesn't have), checked on the client's flag alone.
var windfuryLogClientOnly = []struct {
	name  string
	id    int32
	procs bool
}{
	{"Ghostly Strike", 14278, true}, // 4/15
	{"Kidney Shot", 8643, true},     // 2/19
	{"Shield Slam", 23925, false},   // 0/125
	{"Cheap Shot", 1833, false},     // 0/45
}

// TestForeverWindfuryTotemLogList pins standards 17 (user ruling 2026-11-12): under Forever
// Windfury Totem rolls on white main hand swings and on the melee abilities whose client spell
// requires a main hand weapon or is on next swing, which lands every ability of the log audit
// on its logged side. Off-hand swings, Auto Shot and extra swings never roll.
func TestForeverWindfuryTotemLogList(t *testing.T) {
	RegisterAll()
	for _, a := range windfuryLogClientOnly {
		if got := core.ClientMainHandSpell(a.id); got != a.procs {
			t.Errorf("%s (%d): client main hand flag %v, logs say procs %v", a.name, a.id, got, a.procs)
		}
	}

	var reqs []*proto.RaidSimRequest
	for _, name := range raidReqNames(t, false) {
		reqs = append(reqs, raidReq(t, name))
	}
	tanks, _ := filepath.Glob("../tools/tankbench/req/*.json")
	for _, f := range tanks {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		rsr := &proto.RaidSimRequest{}
		if err := protojson.Unmarshal(data, rsr); err != nil {
			t.Fatal(err)
		}
		reqs = append(reqs, rsr)
	}

	seen := map[int32]bool{}
	white := 0
	for _, rsr := range reqs {
		if rsr.SimOptions == nil {
			rsr.SimOptions = &proto.SimOptions{}
		}
		rsr.SimOptions.Ruleset = proto.Ruleset_RulesetForever
		sim := core.NewSim(rsr, simsignals.CreateSignals())
		for _, party := range sim.Raid.Parties {
			for _, p := range party.Players {
				ch := p.GetCharacter()
				for _, s := range ch.Spellbook {
					switch s.ProcMask {
					case core.ProcMaskMeleeMHAuto:
						white++
						if !core.RollsForeverWindfuryTotem(s) {
							t.Errorf("%v %s: a white main hand swing doesn't roll", ch.Class, s.ActionID)
						}
					case core.ProcMaskMeleeOHAuto, core.ProcMaskRangedAuto:
						if core.RollsForeverWindfuryTotem(s) {
							t.Errorf("%v %s (mask %d) rolls", ch.Class, s.ActionID, s.ProcMask)
						}
					}
					for _, a := range windfuryLogList {
						// The on-next-swing queue spells hit nothing (no proc mask).
						if s.ActionID.SpellID != a.id || s.ProcMask == core.ProcMaskUnknown || s.ProcMask == core.ProcMaskEmpty {
							continue
						}
						seen[a.id] = true
						if got := core.RollsForeverWindfuryTotem(s); got != a.procs {
							t.Errorf("%v %s %s: rolls %v, logs say %v (mask %d, client %v)", ch.Class, a.name, s.ActionID, got, a.procs, s.ProcMask, s.ClientMainHand)
						}
					}
				}
			}
		}
	}
	if white == 0 {
		t.Error("no white main hand swing found")
	}
	for _, a := range windfuryLogList {
		if !seen[a.id] {
			t.Errorf("%s (%d): no build registers it", a.name, a.id)
		}
	}
}
