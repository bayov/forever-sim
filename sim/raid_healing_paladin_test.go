package sim

// The Holy paladin's presets in a healing fight (docs/healing_paladin.md): the rotations cast what they
// should at whom they should, and the talents' mana and procs come through. The heals and talents one by
// one are sim/paladin/holy's tests.

import (
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// paladinHealingRaid is healingRaid with tools/healbench's Holy paladin request (the tank healer, or
// with raid the raid healer) in the healers' places.
func paladinHealingRaid(t *testing.T, raid bool, tanks, healers, mages int, events ...*proto.RaidDamageEvent) *proto.RaidSimRequest {
	rsr := healingRaid(t, tanks, 0, mages, events...)
	path := "../tools/healbench/req/paladin_holy.json"
	if raid {
		path = "../tools/healbench/req/paladin_holy_raid.json"
	}
	healer := benchPlayer(t, path)
	var players []*proto.Player
	for _, party := range rsr.Raid.Parties {
		players = append(players, party.Players...)
	}
	for i := 0; i < healers; i++ {
		p := googleProto.Clone(healer).(*proto.Player)
		players = append(players[:tanks+i], append([]*proto.Player{p}, players[tanks+i:]...)...)
	}
	out := raidOf(t, "mage__frost", players, 20)
	out.Raid.Tanks = rsr.Raid.Tanks
	out.Encounter = rsr.Encounter
	return out
}

func castsBySpell(m *proto.UnitMetrics) map[int32]int32 {
	casts := map[int32]int32{}
	for _, a := range m.Actions {
		for _, tgt := range a.Targets {
			casts[a.Id.GetSpellId()] += tgt.Casts
		}
	}
	return casts
}

// The tank healer in a 5 minute fight: Oath of Light on the tank, Divine Favor on its cooldown, Holy
// Light at the tank, Illumination's mana back, and the tank alive for most of it (one healer against
// 1,500 damage a second runs dry by the end).
func TestHealingFightHolyPaladinTankHealer(t *testing.T) {
	pulse := &proto.RaidDamageEvent{School: proto.SpellSchool_SpellSchoolFire, MinDamage: 600, MaxDamage: 700, Who: proto.RaidDamageEvent_NotTanks, EverySeconds: 6, FirstAt: 3}
	rsr := paladinHealingRaid(t, false, 1, 1, 3, pulse)
	rsr.Encounter.Duration = 300
	res := runRaid(t, rsr)
	healer := playerMetrics(res, 1)
	casts := castsBySpell(healer)
	if casts[1328548] == 0 {
		t.Errorf("no Oath of Light")
	}
	// 2 min cooldown, used when a Holy Light is due: about twice in 5 minutes.
	if casts[20216] < 30 {
		t.Errorf("Divine Favor cast %d times in 20 fights", casts[20216])
	}
	tank := playerMetrics(res, 0).UnitIndex
	onTank, _ := healingOn(healer, tank)
	if casts[25292]+casts[10329] == 0 || onTank <= 0 {
		t.Errorf("Holy Light 9 / 8 cast %d / %d times, %.0f healing on the tank", casts[25292], casts[10329], onTank)
	}
	illumination := 0.0
	for _, r := range healer.Resources {
		if r.Id.GetSpellId() == 20272 {
			illumination += r.Gain
		}
	}
	if illumination <= 0 {
		t.Errorf("no mana from Illumination")
	}
	if hf := res.RaidMetrics.HealingFight; hf == nil || hf.FightsAnyDeath > 0 && hf.FirstDeathSecondsSum/float64(hf.FightsAnyDeath) < 150 {
		t.Errorf("the tank died: %v", hf)
	}
}

// The raid healer heals the mages the raid damage hits.
func TestHealingFightHolyPaladinRaidHealer(t *testing.T) {
	pulse := &proto.RaidDamageEvent{School: proto.SpellSchool_SpellSchoolFire, MinDamage: 800, MaxDamage: 900, Who: proto.RaidDamageEvent_NotTanks, EverySeconds: 5, FirstAt: 3}
	rsr := paladinHealingRaid(t, true, 0, 1, 4, pulse)
	rsr.Encounter.Targets[0].TankIndex = -1
	res := runRaid(t, rsr)
	healer := playerMetrics(res, 0)
	for i := 1; i < 5; i++ {
		if h, _ := healingOn(healer, playerMetrics(res, i).UnitIndex); h <= 0 {
			t.Errorf("mage %d got no healing", i)
		}
	}
	if healer.Ehps.Avg <= 0 {
		t.Errorf("EHPS %.1f", healer.Ehps.Avg)
	}
}
