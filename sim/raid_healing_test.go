package sim

// The healing fight's groundwork (docs/healing_sim.md, H1): heals aimed at hurt players,
// overhealing, every player's health tracked, death, the boss's raid damage and tank swaps. The
// healer is tools/healbench's Restoration shaman. Needs the item database: go test -tags with_db.

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

func benchPlayer(t testing.TB, path string) *proto.Player {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	return rsr.Raid.Parties[0].Players[0]
}

// healingRaid is tanks, healers and Frost mages in that order, in a healing fight against the
// report's boss with the events given.
func healingRaid(t *testing.T, tanks, healers, mages int, events ...*proto.RaidDamageEvent) *proto.RaidSimRequest {
	tank := benchPlayer(t, "../tools/tankbench/req/warrior_protection.json")
	healer := benchPlayer(t, "../tools/healbench/req/shaman_restoration.json")
	mage := raidPlayer(t, "mage__frost")
	var players []*proto.Player
	for i := 0; i < tanks; i++ {
		players = append(players, googleProto.Clone(tank).(*proto.Player))
	}
	for i := 0; i < healers; i++ {
		players = append(players, googleProto.Clone(healer).(*proto.Player))
	}
	for i := 0; i < mages; i++ {
		players = append(players, googleProto.Clone(mage).(*proto.Player))
	}
	rsr := raidOf(t, "mage__frost", players, 20)
	for i := 0; i < tanks; i++ {
		rsr.Raid.Tanks = append(rsr.Raid.Tanks, &proto.UnitReference{Type: proto.UnitReference_Player, Index: int32(i)})
	}
	rsr.Encounter.TrackHealth = true
	rsr.Encounter.RaidDamageHitsPerMinute = 0
	rsr.Encounter.DurationVariation = 0
	rsr.Encounter.RaidDamage = events
	rsr.Encounter.Targets[0].TankIndex = 0
	rsr.Encounter.Targets[0].SwingSpeed = 2
	rsr.Encounter.Targets[0].MinBaseDamage = 3000
	return rsr
}

func playerMetrics(res *proto.RaidSimResult, index int) *proto.UnitMetrics {
	return res.RaidMetrics.Parties[index/5].Players[index%5]
}

// healingOn is a player's healing on the unit, and the part of it that was overhealing.
func healingOn(m *proto.UnitMetrics, unitIndex int32) (healing, overhealing float64) {
	for _, a := range m.Actions {
		for _, tgt := range a.Targets {
			if tgt.UnitIndex == unitIndex {
				healing += tgt.Healing
				overhealing += tgt.Overhealing
			}
		}
	}
	return
}

// The shaman heals the tank the boss hits and the mages the raid damage hits, not itself only,
// and some of it is overhealing but most isn't.
func TestHealingFightHealsOthers(t *testing.T) {
	pulse := &proto.RaidDamageEvent{School: proto.SpellSchool_SpellSchoolFire, MinDamage: 400, MaxDamage: 500, EverySeconds: 10, FirstAt: 5}
	res := runRaid(t, healingRaid(t, 1, 1, 2, pulse))
	healer := playerMetrics(res, 1)

	for i := 0; i < 4; i++ {
		if i == 1 {
			continue
		}
		healing, overhealing := healingOn(healer, playerMetrics(res, i).UnitIndex)
		if healing <= 0 {
			t.Errorf("player %d got no healing", i)
		}
		if overhealing < 0 || overhealing > healing {
			t.Errorf("player %d: %.0f overhealing of %.0f healing", i, overhealing, healing)
		}
	}
	if healer.Ehps.Avg <= 0 || healer.Ehps.Avg > healer.Hps.Avg {
		t.Errorf("EHPS %.1f against HPS %.1f", healer.Ehps.Avg, healer.Hps.Avg)
	}
	for i := 2; i < 4; i++ {
		if m := playerMetrics(res, i); m.Dtps.Avg <= 0 {
			t.Errorf("mage %d took no raid damage", i)
		}
	}
}

// Raid damage nobody can heal kills every mage: they stop casting, so they deal less damage than
// mages who live, and the dead stay dead.
func TestHealingFightDeath(t *testing.T) {
	deadly := &proto.RaidDamageEvent{School: proto.SpellSchool_SpellSchoolShadow, MinDamage: 2500, MaxDamage: 2500, Who: proto.RaidDamageEvent_NotTanks, EverySeconds: 5, FirstAt: 20}
	// No tank, so the boss hits nobody.
	dead := runRaid(t, healingRaid(t, 0, 0, 2, deadly))
	alive := runRaid(t, healingRaid(t, 0, 0, 2))
	if hf := dead.RaidMetrics.HealingFight; hf == nil || hf.FightsAnyDeath != hf.Fights || hf.FightsOtherDeath != hf.Fights || hf.FightsTankDeath != 0 || hf.FirstDeathSecondsSum/float64(hf.Fights) < 20 {
		t.Errorf("the raid's deaths: %v", hf)
	}
	if hf := alive.RaidMetrics.HealingFight; hf == nil || hf.Fights != 20 || hf.FightsAnyDeath != 0 {
		t.Errorf("the raid's deaths with no raid damage: %v", hf)
	}
	for i := 0; i < 2; i++ {
		d, a := playerMetrics(dead, i), playerMetrics(alive, i)
		if d.ChanceOfDeath != 1 {
			t.Errorf("mage %d died in %.0f%% of fights, want all", i, 100*d.ChanceOfDeath)
		}
		if a.ChanceOfDeath != 0 {
			t.Errorf("mage %d died with no raid damage", i)
		}
		if d.Dps.Avg >= 0.5*a.Dps.Avg {
			t.Errorf("mage %d dealt %.0f DPS dying at 30 s, %.0f alive", i, d.Dps.Avg, a.Dps.Avg)
		}
	}
}

// With a tank swap every 20 s both tanks take the boss's melee, and without one only the first.
func TestHealingFightTankSwap(t *testing.T) {
	rsr := healingRaid(t, 2, 2, 0)
	res := runRaid(t, rsr)
	if second := playerMetrics(res, 1).Dtps.Avg; second != 0 {
		t.Errorf("the second tank took %.0f DTPS without a swap", second)
	}

	rsr = healingRaid(t, 2, 2, 0)
	rsr.Encounter.Targets[0].TankSwapSeconds = 20
	res = runRaid(t, rsr)
	first, second := playerMetrics(res, 0).Dtps.Avg, playerMetrics(res, 1).Dtps.Avg
	if first <= 0 || second <= 0 || second < 0.5*first || first < 0.5*second {
		t.Errorf("tanks took %.0f and %.0f DTPS with a swap every 20 s", first, second)
	}
}

// healerRotation gives the healer (player index i) a rotation of the given items.
func healerRotation(rsr *proto.RaidSimRequest, i int, items ...string) {
	player := rsr.Raid.Parties[i/5].Players[i%5]
	rot := &proto.APLRotation{}
	data := `{"type":"TypeAPL","priorityList":[` + strings.Join(items, ",") + `]}`
	if err := protojson.Unmarshal([]byte(data), rot); err != nil {
		panic(err)
	}
	player.Rotation = rot
}

// Chain Heal jumps from its target to the two most hurt, each jump healing half the one before.
func TestHealingFightChainHealJumps(t *testing.T) {
	pulse := &proto.RaidDamageEvent{School: proto.SpellSchool_SpellSchoolFire, MinDamage: 800, MaxDamage: 800, EverySeconds: 6, FirstAt: 2}
	rsr := healingRaid(t, 1, 1, 3, pulse)
	healerRotation(rsr, 1, `{"action":{"condition":{"cmp":{"op":"OpGt","lhs":{"healthMissing":{"sourceUnit":{"type":"Tank"}}},"rhs":{"const":{"val":"500"}}}},"castSpell":{"spellId":{"spellId":10623},"target":{"type":"Tank"}}}}`)
	res := runRaid(t, rsr)
	healer := playerMetrics(res, 1)
	tankIndex := playerMetrics(res, 0).UnitIndex
	var casts, hitsOnTank, jumps int32
	var onTank, onJumps float64
	for _, a := range healer.Actions {
		if a.Id.GetSpellId() != 10623 {
			continue
		}
		for _, tgt := range a.Targets {
			casts += tgt.Casts
			if tgt.UnitIndex == tankIndex {
				hitsOnTank += tgt.Hits + tgt.Crits
				onTank += tgt.Healing
			} else {
				jumps += tgt.Hits + tgt.Crits
				onJumps += tgt.Healing
			}
		}
	}
	if casts == 0 || hitsOnTank != casts {
		t.Fatalf("%d casts, %d landed on the tank", casts, hitsOnTank)
	}
	if jumps < casts || jumps > 2*casts {
		t.Errorf("%d jumps from %d casts, want 1 to 2 each with the raid hurt", jumps, casts)
	}
	// Each jump heals at most half the primary's average (the second a quarter).
	if perJump, perPrimary := onJumps/float64(jumps), onTank/float64(casts); perJump > 0.55*perPrimary || perJump < 0.2*perPrimary {
		t.Errorf("a jump heals %.0f against %.0f on the target", perJump, perPrimary)
	}
}

// A dead player's mana stops coming back, and a health fight ends when the whole raid is dead.
func TestHealingFightDeathEndsRegenAndFight(t *testing.T) {
	deadly := &proto.RaidDamageEvent{School: proto.SpellSchool_SpellSchoolShadow, MinDamage: 2500, MaxDamage: 2500, Who: proto.RaidDamageEvent_NotTanks, EverySeconds: 5, FirstAt: 20}
	regen := func(m *proto.UnitMetrics) (gain float64) {
		for _, r := range m.Resources {
			if r.Id.GetOtherId() == proto.OtherAction_OtherActionManaRegen {
				gain += r.ActualGain
			}
		}
		return
	}
	dead := runRaid(t, healingRaid(t, 0, 0, 2, deadly))
	alive := runRaid(t, healingRaid(t, 0, 0, 2))
	if d, a := regen(playerMetrics(dead, 0)), regen(playerMetrics(alive, 0)); d > 0.5*a {
		t.Errorf("a mage dead at 30 s regenerated %.0f mana, alive %.0f", d, a)
	}

	rsr := healingRaid(t, 0, 0, 2, deadly)
	rsr.Encounter.UseHealth = true
	rsr.Encounter.Targets[0].Stats[stats.Health] = 1e7
	if res := runRaid(t, rsr); res.AvgIterationDuration > 60 {
		t.Errorf("the wiped health fight lasted %.0f s", res.AvgIterationDuration)
	}
}

// Change Target and Channel Spell at a heal target wait while nobody is hurt.
func TestHealingFightEmptyHealTarget(t *testing.T) {
	rsr := healingRaid(t, 1, 1, 1)
	healerRotation(rsr, 1, `{"action":{"changeTarget":{"newTarget":{"type":"MostHurt"}}}}`, `{"action":{"castSpell":{"spellId":{"spellId":10468},"target":{"type":"LowestHealth"}}}}`)
	rsr.SimOptions.Debug = true
	runRaid(t, rsr)
}

// healerOverheal is the healers' overhealing share over all their healing.
func healerOverheal(res *proto.RaidSimResult, healers ...int) float64 {
	var healing, overheal float64
	for _, i := range healers {
		for _, a := range playerMetrics(res, i).Actions {
			for _, tgt := range a.Targets {
				healing += tgt.Healing
				overheal += tgt.Overhealing
			}
		}
	}
	return overheal / healing
}

// Two healers on one tank see each other's heals in flight, so they don't both start a big heal on
// the same hit; and cancelling a heal that would mostly overheal cuts the overhealing a reaction
// time brings.
func TestHealingFightHealComm(t *testing.T) {
	bigHeal := `{"action":{"condition":{"cmp":{"op":"OpGt","lhs":{"healthMissing":{"sourceUnit":{"type":"MostHurt"}}},"rhs":{"const":{"val":"1500"}}}},"castSpell":{"spellId":{"spellId":10396},"target":{"type":"MostHurt"}}}}`
	request := func(reactionMs int32, stopCast float64) *proto.RaidSimRequest {
		rsr := healingRaid(t, 1, 2, 0)
		rsr.SimOptions.Iterations = 50
		for _, i := range []int{1, 2} {
			healerRotation(rsr, i, bigHeal)
			p := rsr.Raid.Parties[0].Players[i]
			p.ReactionTimeMs = reactionMs
			if reactionMs > 0 {
				p.ReactionTimeMs += 50 * int32(i)
			}
			p.StopCastOverheal = stopCast
		}
		return rsr
	}

	// With no reaction time, a healer starting Healing Wave at the tank in the same instant as the
	// other one only does when the tank misses enough for both.
	rsr := request(0, 0)
	rsr.SimOptions.Iterations = 1
	rsr.SimOptions.Debug = true
	starts := map[string]int{}
	casts := 0
	for _, line := range strings.Split(runRaid(t, rsr).Logs, "\n") {
		if strings.Contains(line, "Casting {SpellID: 10396}") {
			starts[line[:strings.Index(line, "]")]]++
			casts++
		}
	}
	same := 0
	for _, n := range starts {
		if n > 1 {
			same += n
		}
	}
	if casts < 20 || same > casts/10 {
		t.Errorf("%d of %d Healing Waves started in the same instant as the other healer's", same, casts)
	}

	late, cancelling := healerOverheal(runRaid(t, request(300, 0)), 1, 2), healerOverheal(runRaid(t, request(300, 0.5)), 1, 2)
	if cancelling >= late {
		t.Errorf("cancelling heals overheals %.0f%%, not cancelling %.0f%%", 100*cancelling, 100*late)
	}
}

// bindRank is a line casting one rank of a heal at the tank once it misses what the rank heals.
func bindRank(spellID int32) string {
	id := strconv.Itoa(int(spellID))
	return `{"action":{"condition":{"cmp":{"op":"OpGe","lhs":{"healthMissing":{"sourceUnit":{"type":"Tank"}}},"rhs":{"spellExpectedHealing":{"spellId":{"spellId":` + id + `},"targetUnit":{"type":"Tank"}}}}},"castSpell":{"spellId":{"spellId":` + id + `},"target":{"type":"Tank"}}}}`
}

// A healer casts the biggest of its bound ranks that what's missing covers, and binds no more than
// 4 ranks of a heal: a 5th rank's line is off.
func TestHealingFightBoundRanks(t *testing.T) {
	ranks := []int32{25357, 10395, 959, 913, 332} // Healing Wave 10, 8, 6, 4 and 2
	rsr := healingRaid(t, 1, 1, 0)
	rsr.SimOptions.Iterations = 20
	var lines []string
	for _, id := range ranks {
		lines = append(lines, bindRank(id))
	}
	healerRotation(rsr, 1, lines...)
	casts := map[int32]int32{}
	for _, a := range playerMetrics(runRaid(t, rsr), 1).Actions {
		for _, tgt := range a.Targets {
			casts[a.Id.GetSpellId()] += tgt.Casts
		}
	}
	if casts[25357] == 0 || casts[10395]+casts[959]+casts[913] == 0 {
		t.Errorf("casts by rank %v: want the top rank and lower ones", casts)
	}
	if casts[332] != 0 {
		t.Errorf("the 5th bound rank was cast %d times", casts[332])
	}
}

// A share hits that part of the group a cast, the fraction a chance of one more: 0.375 of four
// mages is one or two a cast, 1.5 on average, where hitting them all is 4 a cast.
func TestHealingFightRaidDamageShare(t *testing.T) {
	every := &proto.RaidDamageEvent{School: proto.SpellSchool_SpellSchoolArcane, MinDamage: 100, MaxDamage: 100, Who: proto.RaidDamageEvent_NotTanks, EverySeconds: 2, FirstAt: 1}
	part := googleProto.Clone(every).(*proto.RaidDamageEvent)
	part.Share = 0.375
	// The event's hits on the players (it always lands).
	hits := func(res *proto.RaidSimResult) (n float64) {
		for _, a := range res.EncounterMetrics.Targets[0].Actions {
			if a.Id.GetOtherId() != proto.OtherAction_OtherActionDamageTaken {
				continue
			}
			for _, tgt := range a.Targets {
				n += float64(tgt.Hits)
			}
		}
		return n
	}
	all, some := hits(runRaid(t, healingRaid(t, 0, 0, 4, every))), hits(runRaid(t, healingRaid(t, 0, 0, 4, part)))
	if all == 0 {
		t.Fatal("no casts of the event")
	}
	if r := some / all; r < 0.35 || r > 0.40 {
		t.Errorf("a 0.375 share hit %.3f as often as hitting everyone (%.0f of %.0f)", r, some, all)
	}
}

// The Restoration shaman's tank healer preset (tools/healbench's request) in a 5 minute fight: it
// keeps Water Shield and Mana Spring Totem up and puts Mana Tide down when its mana runs low, and
// the totems' mana reaches the mages in its party.
func TestHealingFightRestorationManaTools(t *testing.T) {
	pulse := &proto.RaidDamageEvent{School: proto.SpellSchool_SpellSchoolFire, MinDamage: 600, MaxDamage: 700, Who: proto.RaidDamageEvent_NotTanks, EverySeconds: 6, FirstAt: 3}
	rsr := healingRaid(t, 1, 1, 3, pulse)
	rsr.Encounter.Duration = 300
	res := runRaid(t, rsr)
	healer := playerMetrics(res, 1)
	casts := map[int32]int32{}
	for _, a := range healer.Actions {
		for _, tgt := range a.Targets {
			casts[a.Id.GetSpellId()] += tgt.Casts
		}
	}
	for id, name := range map[int32]string{408510: "Water Shield", 10497: "Mana Spring Totem", 1325256: "Mana Tide Totem"} {
		if casts[id] == 0 {
			t.Errorf("no %s cast", name)
		}
	}
	for i := 2; i < 5; i++ {
		gained := map[int32]float64{}
		for _, r := range playerMetrics(res, i).Resources {
			gained[r.Id.GetSpellId()] += r.Gain
		}
		if gained[10497] <= 0 || gained[1325256] <= 0 {
			t.Errorf("mage %d got %.0f mana from Mana Spring and %.0f from Mana Tide", i, gained[10497], gained[1325256])
		}
	}
	if healer.Ehps.Avg <= 0 {
		t.Errorf("EHPS %.1f", healer.Ehps.Avg)
	}
}
