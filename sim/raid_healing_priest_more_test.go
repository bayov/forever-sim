package sim

// More of the healing priest (raid_healing_priest_test.go): every heal rank against the client's
// tables, the heals' spread, Inner Focus, Power Infusion, Prayer of Mending's charge, Renew's ranks,
// Blessed Recovery, Twilight Focus in a cast and the priest's set bonuses (the healer mechanics
// audit's Untested list, orchestration/audit/reports/healer_mechanics.md). Needs the item database:
// go test -tags with_db.

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/priest"
	"github.com/wowsims/classic/sim/priest/healing"
)

// clientRank is a heal rank's values in the Forever client 1.60.1.70753 (SpellEffect, SpellLevels,
// SpellMisc, SpellCastTimes, SpellPower on wago.tools), as orchestration/tools/healers/priest_audit.py
// reads them: the heal effect's base points, growth a level and variance, its coefficient, the learn
// and max levels, the cast time and the mana cost.
type clientRank struct {
	rank                 int
	spellID              int32
	base, perLevel, vari float64
	learn, max           int32
	coef                 float64
	castMs               int
	cost                 float64
}

// at is the rank's lowest and highest heal at the level, the client's way.
func (c clientRank) at(level int32) (float64, float64) {
	v := c.base + c.perLevel*float64(max(0, min(level, c.max)-c.learn))
	return v * (1 - c.vari/2), v * (1 + c.vari/2)
}

var clientHealRanks = map[string][]clientRank{
	"Lesser Heal": {
		{1, 2050, 51, 0.3000000119, 0.1960784346, 1, 3, 0.4289999902, 1500, 30},
		{2, 2052, 94, 0.6999999881, 0.1794871837, 4, 9, 0.57099998, 2000, 45},
		{3, 2053, 163, 1.200000048, 0.1506849378, 10, 15, 0.7139999866, 2500, 75},
	},
	"Heal": {
		{1, 2054, 270, 2, 0.1446540952, 16, 21, 0.8569999933, 3000, 155},
		{2, 2055, 375, 2.900000095, 0.1347826123, 22, 27, 0.8569999933, 3000, 205},
		{3, 6063, 523, 4, 0.1258278191, 28, 33, 0.8569999933, 3000, 255},
		{4, 6064, 651, 4.5, 0.1213720292, 34, 39, 0.8569999933, 3000, 305},
	},
	"Greater Heal": {
		{1, 2060, 844, 5.099999905, 0.1192468628, 40, 45, 0.8569999933, 3000, 370},
		{2, 10963, 1099, 5.800000191, 0.1148482263, 46, 51, 0.8569999933, 3000, 455},
		{3, 10964, 1403, 6.599999905, 0.1129349992, 52, 57, 0.8569999933, 3000, 545},
		{4, 10965, 1782, 7.5, 0.1093585715, 58, 63, 0.8569999933, 3000, 655},
		{5, 25314, 1960, 8.100000381, 0.109615393, 60, 65, 0.8569999933, 3000, 710},
	},
	"Flash Heal": {
		{1, 2061, 194, 1.899999976, 0.204651162, 20, 25, 0.4289999902, 1500, 125},
		{2, 9472, 249, 2.200000048, 0.1958041936, 26, 31, 0.4289999902, 1500, 155},
		{3, 9473, 312, 2.5, 0.1833333373, 32, 37, 0.4289999902, 1500, 185},
		{4, 9474, 383, 2.799999952, 0.1776765436, 38, 43, 0.4289999902, 1500, 215},
		{5, 10915, 507, 3.299999952, 0.1728395075, 44, 49, 0.4289999902, 1500, 265},
		{6, 10916, 644, 3.700000048, 0.1704545468, 50, 55, 0.4289999902, 1500, 315},
		{7, 10917, 825, 4.199999809, 0.1649717539, 56, 61, 0.4289999902, 1500, 380},
	},
	"Prayer of Healing": {
		{1, 596, 178, 0.8000000119, 0.06430868059, 30, 39, 0.2860000134, 3000, 410},
		{2, 996, 265, 1, 0.06113537028, 40, 49, 0.2860000134, 3000, 560},
		{3, 10960, 401, 1.299999952, 0.05621301755, 50, 59, 0.2860000134, 3000, 770},
		{4, 10961, 583, 1.5, 0.05388601124, 60, 69, 0.2860000134, 3000, 1030},
		{5, 25316, 649, 1.600000024, 0.05420560762, 60, 69, 0.2860000134, 3000, 1070},
	},
	// The heal is the triggered spell's (23455 and so on); the cast is instant.
	"Holy Nova": {
		{1, 15237, 51, 0.400000006, 0.1428571492, 20, 26, 0.1070000008, 0, 185},
		{2, 15430, 82, 0.5, 0.1304347813, 28, 34, 0.1070000008, 0, 290},
		{3, 15431, 116, 0.6000000238, 0.1384615451, 36, 42, 0.1070000008, 0, 400},
		{4, 27799, 159, 0.6999999881, 0.1494252831, 44, 50, 0.1070000008, 0, 520},
		{5, 27800, 238, 0.8000000119, 0.1422924846, 52, 58, 0.1070000008, 0, 635},
		{6, 27801, 311, 0.8999999762, 0.1472392678, 60, 66, 0.1070000008, 0, 750},
	},
	// A tick's heal.
	"Renew": {
		{1, 139, 9, 0, 0, 8, 13, 0.200000003, 0, 30},
		{2, 6074, 15, 0, 0, 14, 19, 0.200000003, 0, 65},
		{3, 6075, 25, 0, 0, 20, 25, 0.200000003, 0, 105},
		{4, 6076, 32, 0, 0, 26, 31, 0.200000003, 0, 140},
		{5, 6077, 41, 0, 0, 32, 37, 0.200000003, 0, 170},
		{6, 6078, 54, 0, 0, 38, 43, 0.200000003, 0, 205},
		{7, 10927, 74, 0, 0, 44, 49, 0.200000003, 0, 250},
		{8, 10928, 102, 0, 0, 50, 55, 0.200000003, 0, 305},
		{9, 10929, 134, 0, 0, 56, 61, 0.200000003, 0, 365},
		{10, 25315, 166, 0, 0, 60, 65, 0.200000003, 0, 410},
	},
	// The shield's absorb.
	"Power Word: Shield": {
		{1, 17, 44, 0.8000000119, 0, 6, 11, 0.1000000015, 0, 45},
		{2, 592, 86, 1.200000048, 0, 12, 17, 0.1000000015, 0, 80},
		{3, 600, 154, 1.600000024, 0, 18, 23, 0.1000000015, 0, 130},
		{4, 3747, 226, 2, 0, 24, 29, 0.1000000015, 0, 175},
		{5, 6065, 291, 2.299999952, 0, 30, 35, 0.1000000015, 0, 210},
		{6, 6066, 368, 2.599999905, 0, 36, 41, 0.1000000015, 0, 250},
		{7, 10898, 470, 3, 0, 42, 47, 0.1000000015, 0, 300},
		{8, 10899, 591, 3.400000095, 0, 48, 53, 0.1000000015, 0, 355},
		{9, 10900, 749, 3.900000095, 0, 54, 59, 0.1000000015, 0, 425},
		{10, 10901, 928, 4.300000191, 0, 60, 65, 0.1000000015, 0, 500},
	},
}

// rankTable is a rank of the sim's heal tables (sim/priest/healing.go).
type rankTable interface {
	SpellID() int32
	Coefficient() float64
	LearnLevel() int32
	At(level int32) (float64, float64)
}

type simHeal struct {
	ranks  []rankTable
	spells []*core.Spell
}

// simHealRanks is each heal's rank table and registered spells, by the heal's name.
func simHealRanks(hp *healing.HealingPriest) map[string]simHeal {
	tables := map[string]simHeal{}
	add := func(name string, spells []*core.Spell, ranks ...rankTable) {
		tables[name] = simHeal{ranks, spells}
	}
	add("Lesser Heal", hp.LesserHeal, toTables(priest.LesserHealRanks)...)
	add("Heal", hp.Heal, toTables(priest.HealRanks)...)
	add("Greater Heal", hp.GreaterHeal, toTables(priest.GreaterHealRanks)...)
	add("Flash Heal", hp.FlashHeal, toTables(priest.FlashHealRanks)...)
	add("Prayer of Healing", hp.PrayerOfHealing, toTables(priest.PrayerOfHealingRanks)...)
	add("Holy Nova", hp.HolyNova, toTables(priest.HolyNovaRanks)...)
	add("Renew", hp.Renew, toTables(priest.RenewRanks)...)
	add("Power Word: Shield", hp.PowerWordShield, toTables(priest.PowerWordShieldRanks)...)
	return tables
}

// toTables is a rank table, its rank 0 placeholder included, as rankTables.
func toTables[R rankTable](ranks []R) []rankTable {
	out := make([]rankTable, len(ranks))
	for i, r := range ranks {
		out[i] = r
	}
	return out
}

// Every rank of every heal (and Renew's tick and the shield's absorb) at 60 is the client's: the
// rank's spell, its heal's base, growth, max level and spread, its coefficient and learn level, and
// the registered spell's mana and cast time.
func TestHealingPriestRankTables(t *testing.T) {
	_, hp := newPriestSim(t, priestRaid(t, priestTalents(map[string]int{"holy_nova": 1, "prayer_of_mending": 1, "penance": 1}), 0))
	tables := simHealRanks(hp)
	for name, client := range clientHealRanks {
		table, ok := tables[name]
		if !ok || len(table.ranks) != len(client)+1 {
			t.Errorf("%s: the sim has %d ranks, the client %d", name, len(table.ranks)-1, len(client))
			continue
		}
		for _, c := range client {
			r := table.ranks[c.rank]
			what := name + " " + strconv.Itoa(c.rank)
			if r.SpellID() != c.spellID {
				t.Errorf("%s is spell %d, want %d", what, r.SpellID(), c.spellID)
			}
			low, high := r.At(60)
			wantLow, wantHigh := c.at(60)
			if !near(low, wantLow, 1e-6) || !near(high, wantHigh, 1e-6) {
				t.Errorf("%s heals %.2f to %.2f at 60, want %.2f to %.2f", what, low, high, wantLow, wantHigh)
			}
			if math.Abs(r.Coefficient()-c.coef) > 1e-9 || r.LearnLevel() != c.learn {
				t.Errorf("%s: coefficient %.4f learned at %d, want %.4f at %d", what, r.Coefficient(), r.LearnLevel(), c.coef, c.learn)
			}
			spell := table.spells[c.rank]
			if spell == nil {
				t.Errorf("%s isn't registered", what)
				continue
			}
			if got := spell.Cost.GetCurrentCost(); got != c.cost {
				t.Errorf("%s costs %.0f, want %.0f", what, got, c.cost)
			}
			if got := spell.DefaultCast.CastTime; got != time.Duration(c.castMs)*time.Millisecond {
				t.Errorf("%s casts in %s, want %d ms", what, got, c.castMs)
			}
		}
	}
}

// A heal's hits spread evenly over the client's range at 60 (its variance): on many heals of each
// heal with a spread, none falls outside it and the lowest and highest come near its ends.
func TestHealingPriestHealVariance(t *testing.T) {
	rsr := priestRaid(t, priestTalents(map[string]int{"holy_nova": 1}), 0)
	rsr.Encounter.Targets[0].MinBaseDamage = 0
	sim, hp := newPriestSim(t, rsr)
	tank := &sim.Raid.Parties[0].Players[0].GetCharacter().Unit
	at(sim, map[time.Duration]func(*core.Simulation){time.Second: func(sim *core.Simulation) {
		for _, c := range []struct {
			name  string
			spell *core.Spell
			rank  int
		}{
			{"Lesser Heal", hp.LesserHeal[3], 3},
			{"Heal", hp.Heal[4], 4},
			{"Greater Heal", hp.GreaterHeal[5], 5},
			{"Flash Heal", hp.FlashHeal[7], 7},
			{"Prayer of Healing", hp.PrayerOfHealing[5], 5},
			{"Holy Nova", hp.HolyNova[6], 6},
		} {
			lowBase, highBase := clientHealRanks[c.name][c.rank-1].at(60)
			// The game rolls the base heal in whole points (core.Simulation.Roll).
			lowBase, highBase = math.Round(lowBase), math.Round(highBase)
			low, high := c.spell.ExpectedHealing(tank, lowBase), c.spell.ExpectedHealing(tank, highBase)
			m := &c.spell.SpellMetrics[tank.UnitIndex]
			least, most := math.Inf(1), math.Inf(-1)
			for i := 0; i < 400; i++ {
				healing, crit := m.TotalHealing, m.TotalCritHealing
				c.spell.ApplyEffects(sim, tank, c.spell)
				if m.TotalCritHealing != crit {
					continue
				}
				got := m.TotalHealing - healing
				least, most = min(least, got), max(most, got)
			}
			spread := high - low
			if least < low-1e-6 || most > high+1e-6 || least > low+0.05*spread || most < high-0.05*spread {
				t.Errorf("%s %d healed %.1f to %.1f, want %.1f to %.1f", c.name, c.rank, least, most, low, high)
			}
		}
	}})
}

// quietPriestSim is the priest raid with nothing hurting the tank and the tank doing nothing.
func quietPriestSim(t *testing.T, talents map[string]int, mages int) (*core.Simulation, *healing.HealingPriest, *core.Unit) {
	rsr := priestRaid(t, priestTalents(talents), mages)
	rsr.Encounter.Targets[0].MinBaseDamage = 0
	rsr.Raid.Parties[0].Players[0].Rotation = &proto.APLRotation{}
	sim, hp := newPriestSim(t, rsr)
	return sim, hp, &sim.Raid.Parties[0].Players[0].GetCharacter().Unit
}

// Inner Focus makes Renew and Power Word: Shield free, gives Renew's ticks no crit, goes with the
// spell it made free and is then on its 3 min cooldown.
func TestHealingPriestInnerFocusRenewAndShield(t *testing.T) {
	sim, hp, tank := quietPriestSim(t, map[string]int{"inner_focus": 1}, 0)
	renew, shield := top(hp.Renew), top(hp.PowerWordShield)
	var critWithout float64
	at(sim, map[time.Duration]func(*core.Simulation){
		time.Second: func(sim *core.Simulation) {
			renew.Cast(sim, &hp.Unit)
			critWithout = renew.Hot(&hp.Unit).SnapshotCritChance
			if !hp.InnerFocus.Cast(sim, nil) || !hp.InnerFocusAura.IsActive() {
				t.Fatal("Inner Focus didn't go up")
			}
			if r, s := renew.Cost.GetCurrentCost(), shield.Cost.GetCurrentCost(); r != 0 || s != 0 {
				t.Errorf("under Inner Focus Renew costs %.0f and Power Word: Shield %.0f, want both free", r, s)
			}
		},
		3 * time.Second: func(sim *core.Simulation) {
			mana := hp.CurrentMana()
			renew.Cast(sim, tank)
			if got := hp.CurrentMana(); got != mana {
				t.Errorf("Renew under Inner Focus took %.0f mana", mana-got)
			}
			if got := renew.Hot(tank).SnapshotCritChance; !near(got, critWithout, 1e-9) {
				t.Errorf("Renew under Inner Focus ticks with %.3f crit, %.3f without", got, critWithout)
			}
			if hp.InnerFocusAura.IsActive() {
				t.Error("Renew didn't use up Inner Focus")
			}
			if got := hp.InnerFocus.TimeToReady(sim); got != 3*time.Minute {
				t.Errorf("Inner Focus is ready in %s, want 3m", got)
			}
			if renew.Cost.GetCurrentCost() == 0 {
				t.Error("Renew is still free after Inner Focus")
			}
		},
		3*time.Second + 2*time.Minute: func(sim *core.Simulation) {
			if hp.InnerFocus.CanCast(sim, nil) {
				t.Error("Inner Focus is ready 2 min after its use")
			}
		},
	})
}

// Inner Focus on Penance: the channel is free and uses Inner Focus up, so the bolts after the first
// get no crit from it (the audit's open question allows the first bolt at most).
func TestHealingPriestInnerFocusPenance(t *testing.T) {
	sim, hp, tank := quietPriestSim(t, map[string]int{"inner_focus": 1, "penance": 1}, 0)
	penance := top(hp.Penance)
	bolt := hp.GetSpell(core.ActionID{SpellID: priest.PenanceHealSpellId[4]})
	critBefore := bolt.BonusCritRating
	at(sim, map[time.Duration]func(*core.Simulation){
		time.Second: func(sim *core.Simulation) {
			hp.InnerFocus.Cast(sim, nil)
			if got := penance.Cost.GetCurrentCost(); got != 0 {
				t.Errorf("Penance costs %.0f under Inner Focus", got)
			}
			mana := hp.CurrentMana()
			penance.Cast(sim, tank)
			if got := hp.CurrentMana(); got != mana {
				t.Errorf("Penance under Inner Focus took %.0f mana", mana-got)
			}
		},
		1500 * time.Millisecond: func(sim *core.Simulation) {
			if hp.InnerFocusAura.IsActive() || hp.InnerFocus.TimeToReady(sim) <= 0 {
				t.Error("Penance didn't use up Inner Focus")
			}
			if got := bolt.BonusCritRating; got != critBefore {
				t.Errorf("Penance's second and third bolts have %.0f crit, %.0f without Inner Focus", got, critBefore)
			}
		},
	})
	if got := bolt.SpellMetrics[tank.UnitIndex].Hits + bolt.SpellMetrics[tank.UnitIndex].Crits; got != 3 {
		t.Errorf("Penance under Inner Focus fired %d heal bolts, want 3", got)
	}
}

// Power Infusion costs 20% of base mana and is on a 3 min cooldown; a Renew cast under it keeps its
// 20% more for every tick, those after it ran out too.
func TestHealingPriestPowerInfusion(t *testing.T) {
	sim, hp, tank := quietPriestSim(t, map[string]int{"power_infusion": 1, "penance": 1}, 0)
	pi, renew := hp.PowerInfusion, top(hp.Renew)
	if got, want := pi.Cost.GetCurrentCost(), 0.2*hp.BaseMana; got != want || want <= 0 {
		t.Errorf("Power Infusion costs %.0f, want %.0f (20%% of base mana)", got, want)
	}
	m := &renew.SpellMetrics[tank.UnitIndex]
	type ticks struct{ healing, n float64 }
	var under ticks
	at(sim, map[time.Duration]func(*core.Simulation){
		time.Second: func(sim *core.Simulation) {
			mana := hp.CurrentMana()
			pi.Cast(sim, &hp.Unit)
			if got := mana - hp.CurrentMana(); !near(got, 0.2*hp.BaseMana, 1e-9) {
				t.Errorf("Power Infusion took %.0f mana, want %.0f", got, 0.2*hp.BaseMana)
			}
			if got := pi.TimeToReady(sim); got != 3*time.Minute {
				t.Errorf("Power Infusion is ready in %s, want 3m", got)
			}
		},
		// Ticks at 13, 16, 19, 22 and 25 s; Power Infusion runs out at 16.
		10 * time.Second: func(sim *core.Simulation) { renew.Cast(sim, tank) },
		26 * time.Second: func(sim *core.Simulation) {
			under = ticks{m.TotalHealing - m.TotalCritHealing, float64(m.Ticks)}
			renew.Cast(sim, tank)
		},
	})
	without := ticks{m.TotalHealing - m.TotalCritHealing - under.healing, float64(m.Ticks) - under.n}
	if under.n == 0 || without.n == 0 {
		t.Fatalf("%.0f ticks under Power Infusion, %.0f without", under.n, without.n)
	}
	if got := (under.healing / under.n) / (without.healing / without.n); !near(got, 1.2, 1e-9) {
		t.Errorf("a Renew cast under Power Infusion ticks for %.3f times one cast without, want 1.2", got)
	}
}

// pomHolders is the's Prayer of Mending charges up in the raid, by unit.
func pomHolders(sim *core.Simulation) map[*core.Unit]int {
	holders := map[*core.Unit]int{}
	for _, unit := range sim.Raid.AllPlayerUnits {
		for _, aura := range unit.GetAurasWithTag("PrayerOfMending") {
			if aura.IsActive() {
				holders[unit]++
			}
		}
	}
	return holders
}

// Prayer of Mending heals on a heal taken that isn't a tick (Renew's ticks don't set it off), then
// jumps; casting it again moves the charge: one up per priest.
func TestHealingPriestPrayerOfMendingTriggers(t *testing.T) {
	sim, hp, tank := quietPriestSim(t, map[string]int{"prayer_of_mending": 1}, 2)
	pom := top(hp.PrayerOfMending)
	heal := hp.GetSpell(core.ActionID{SpellID: priest.PrayerOfMendingHealID, Tag: 3})
	healed := func() int32 {
		n := int32(0)
		for _, m := range heal.SpellMetrics {
			n += m.Hits + m.Crits
		}
		return n
	}
	at(sim, map[time.Duration]func(*core.Simulation){
		time.Second: func(sim *core.Simulation) { pom.Cast(sim, tank) },
		3 * time.Second: func(sim *core.Simulation) {
			top(hp.Renew).Cast(sim, tank)
		},
		// Renew ticked at 6 and 9 s.
		9500 * time.Millisecond: func(sim *core.Simulation) {
			if n, holders := healed(), pomHolders(sim); n != 0 || holders[tank] != 1 || len(holders) != 1 {
				t.Errorf("after two Renew ticks: %d Prayer of Mending heals, holders %v, want none and the tank", n, holders)
			}
			hp.FlashHeal[7].Cast(sim, tank)
		},
		12 * time.Second: func(sim *core.Simulation) {
			holders := pomHolders(sim)
			if n := healed(); n != 1 || heal.SpellMetrics[tank.UnitIndex].Hits+heal.SpellMetrics[tank.UnitIndex].Crits != 1 {
				t.Errorf("a Flash Heal on the tank set off %d Prayer of Mending heals, want 1 on the tank", n)
			}
			if holders[tank] != 0 || len(holders) != 1 {
				t.Errorf("after its heal Prayer of Mending is on %v, want one other player", holders)
			}
			if !pom.Cast(sim, &hp.Unit) {
				t.Fatal("Prayer of Mending couldn't be cast again")
			}
			if holders := pomHolders(sim); holders[&hp.Unit] != 1 || len(holders) != 1 {
				t.Errorf("cast again at the priest, Prayer of Mending is on %v, want the priest alone", holders)
			}
		},
	})
}

// twoPriestRaid is the tank (0), a priest (1) and a second priest (2) with these talents, with
// nothing hurting the tank.
func twoPriestRaid(t *testing.T, talents string, seed int64) (*core.Simulation, *healing.HealingPriest, *healing.HealingPriest, *core.Unit) {
	rsr := priestRaid(t, talents, 0)
	other := testPriest(talents)
	other.Name = "2"
	rsr.Raid.Parties[0].Players = append(rsr.Raid.Parties[0].Players, other)
	rsr.Encounter.Targets[0].MinBaseDamage = 0
	rsr.Raid.Parties[0].Players[0].Rotation = &proto.APLRotation{}
	rsr.SimOptions.RandomSeed = seed
	sim, hp := newPriestSim(t, rsr)
	return sim, hp, sim.Raid.Parties[0].Players[2].(*healing.HealingPriest), &sim.Raid.Parties[0].Players[0].GetCharacter().Unit
}

// Prayer of Mending jumps to a player that carries none: with the second priest's on the first
// priest, the first priest's jumps from the tank to the second priest every time.
func TestHealingPriestPrayerOfMendingSkipsHolders(t *testing.T) {
	for seed := int64(1); seed <= 12; seed++ {
		sim, a, b, tank := twoPriestRaid(t, priestTalents(map[string]int{"prayer_of_mending": 1}), seed)
		at(sim, map[time.Duration]func(*core.Simulation){
			time.Second: func(sim *core.Simulation) {
				top(a.PrayerOfMending).Cast(sim, tank)
				top(b.PrayerOfMending).Cast(sim, &a.Unit)
			},
			// Lands at 4 s.
			2500 * time.Millisecond: func(sim *core.Simulation) { a.FlashHeal[7].Cast(sim, tank) },
			5 * time.Second: func(sim *core.Simulation) {
				holders := pomHolders(sim)
				if holders[tank] != 0 || holders[&a.Unit] != 1 || holders[&b.Unit] != 1 {
					t.Errorf("seed %d: Prayer of Mending jumped from the tank to leave the tank %d, the first priest %d and the second %d, want 0, 1 and 1",
						seed, holders[tank], holders[&a.Unit], holders[&b.Unit])
				}
			},
		})
	}
}

// A Renew of ours replaces our Renew of another rank on its target, and leaves another priest's.
func TestHealingPriestRenewReplacesOurRank(t *testing.T) {
	sim, a, b, tank := twoPriestRaid(t, "", 1)
	at(sim, map[time.Duration]func(*core.Simulation){
		time.Second: func(sim *core.Simulation) {
			a.Renew[10].Cast(sim, tank)
			b.Renew[10].Cast(sim, tank)
		},
		3 * time.Second: func(sim *core.Simulation) {
			a.Renew[9].Cast(sim, tank)
			if a.Renew[10].Hot(tank).IsActive() || !a.Renew[9].Hot(tank).IsActive() {
				t.Errorf("after our Renew 9, Renew 10 up %v and Renew 9 up %v, want only 9", a.Renew[10].Hot(tank).IsActive(), a.Renew[9].Hot(tank).IsActive())
			}
			if !b.Renew[10].Hot(tank).IsActive() {
				t.Error("our Renew 9 took off the other priest's Renew 10")
			}
		},
	})
}

// Blessed Recovery: a big melee hit heals 25% of it over 6 sec, and a second one during it carries
// what the first had left over, so the two heal 25% of both in all; a small hit does nothing.
func TestHealingPriestBlessedRecoveryCarryOver(t *testing.T) {
	sim, hp, _ := quietPriestSim(t, map[string]int{"blessed_recovery": 3}, 0)
	talent := hp.GetAura("Blessed Recovery Talent")
	hot := hp.GetSpell(core.ActionID{SpellID: 27813})
	melee := sim.Encounter.TargetUnits[0].AutoAttacks.MHAuto()
	hit := func(sim *core.Simulation, damage float64) {
		talent.OnSpellHitTaken(talent, sim, melee, &core.SpellResult{Target: &hp.Unit, Damage: damage, Outcome: core.OutcomeHit})
	}
	big := 0.5 * hp.MaxHealth()
	at(sim, map[time.Duration]func(*core.Simulation){
		time.Second: func(sim *core.Simulation) { hit(sim, 0.2*hp.MaxHealth()) },
		// Its ticks at 3, 5 and 7 s; the second hit after the first tick.
		1100 * time.Millisecond: func(sim *core.Simulation) {
			if hot.SelfHot().IsActive() {
				t.Error("a hit for 20% of our health set off Blessed Recovery")
			}
			hit(sim, big)
		},
		4 * time.Second: func(sim *core.Simulation) { hit(sim, big) },
	})
	if got, want := hot.SpellMetrics[hp.UnitIndex].TotalHealing, 0.25*2*big; !near(got, want, 1e-9) {
		t.Errorf("two hits of %.0f gave Blessed Recovery %.1f healing, want %.1f", big, got, want)
	}
}

// Twilight Focus gives a heal a 70% chance (3 ranks) to ignore each hit's pushback: under raid
// hits, Greater Heal's casts run about 30% as far over their 3 sec as without it.
func TestHealingPriestTwilightFocusPushback(t *testing.T) {
	hits := &proto.RaidDamageEvent{School: proto.SpellSchool_SpellSchoolFire, MinDamage: 50, MaxDamage: 50, Who: proto.RaidDamageEvent_Everyone, EverySeconds: 1.3, FirstAt: 0.5}
	over := func(talents map[string]int) float64 {
		_, m := fightTalent(t, talents, 20, []*proto.RaidDamageEvent{hits}, 0, greaterHealAtTank)
		var casts int32
		var ms float64
		for _, a := range m.Actions {
			if a.Id.GetSpellId() == priest.GreaterHealRanks[5].SpellID() {
				for _, tgt := range a.Targets {
					casts += tgt.Casts
					ms += tgt.CastTimeMs
				}
			}
		}
		return ms/float64(casts) - 3000
	}
	with, without := over(map[string]int{"twilight_focus": 3}), over(map[string]int{})
	if without < 500 || with/without < 0.15 || with/without > 0.5 {
		t.Errorf("Greater Heal casts ran %.0f ms over with Twilight Focus, %.0f ms without: want about 30%% as far", with, without)
	}
}

// setPriest builds a priest wearing pieces of a set (test pieces with no stats), or with the bonus
// stats given instead.
func setPriest(t *testing.T, talents string, setName string, firstID int32, pieces int, bonus stats.Stats) *healing.HealingPriest {
	rsr := priestRaid(t, talents, 0)
	p := rsr.Raid.Parties[0].Players[1]
	wearTestSet(t, p, setName, firstID, pieces)
	for s, v := range bonus {
		p.BonusStats.Stats[s] += v
	}
	_, hp := newPriestSim(t, rsr)
	return hp
}

// checkSetStats is whether each piece count's priest has the stats of a priest with no set and
// the bonus stats given.
func checkSetStats(t *testing.T, setName string, firstID int32, pieces int, bonus stats.Stats) {
	t.Helper()
	worn := setPriest(t, "", setName, firstID, pieces, stats.Stats{})
	given := setPriest(t, "", setName, firstID, 0, bonus)
	bare := setPriest(t, "", setName, firstID, 0, stats.Stats{})
	for s, v := range bonus {
		if v == 0 {
			continue
		}
		stat := stats.Stat(s)
		if got, want := worn.GetStat(stat), given.GetStat(stat); math.Abs(got-want) > 1e-9 || got == bare.GetStat(stat) {
			t.Errorf("%s %d pieces: %s %.2f, want %.2f (%.2f without)", setName, pieces, stat.StatName(), got, want, bare.GetStat(stat))
		}
	}
}

// Vestments of Conviction (the Smite set's bonuses): 2 pieces 10 Spirit, 4 pieces 26 healing and
// 9 damage, 5 pieces 1 sec off Penance's and Prayer of Mending's cooldowns (healer_mechanics-
// 20261104-5).
func TestHealingPriestVestmentsOfConviction(t *testing.T) {
	const set = "Smite Synthetic (Phase 1)"
	checkSetStats(t, set, 999_400, 2, stats.Stats{stats.Spirit: 10})
	checkSetStats(t, set, 999_400, 4, stats.Stats{stats.Spirit: 10, stats.HealingPower: 26, stats.SpellDamage: 9})
	talents := priestTalents(map[string]int{"prayer_of_mending": 1, "penance": 1})
	for _, c := range []struct {
		pieces       int
		pom, penance time.Duration
	}{{4, 10 * time.Second, 12 * time.Second}, {5, 9 * time.Second, 11 * time.Second}} {
		hp := setPriest(t, talents, set, 999_400, c.pieces, stats.Stats{})
		for rank, spell := range hp.PrayerOfMending[1:] {
			if spell.CD.Duration != c.pom {
				t.Errorf("%d pieces: Prayer of Mending %d's cooldown is %s, want %s", c.pieces, rank+1, spell.CD.Duration, c.pom)
			}
		}
		for rank, spell := range hp.Penance[1:] {
			if spell.CD.Duration != c.penance {
				t.Errorf("%d pieces: Penance %d's cooldown is %s, want %s", c.pieces, rank+1, spell.CD.Duration, c.penance)
			}
		}
	}
}

// The Era priest sets' healing bonuses: Confessor's Raiment 2 pieces 22 healing, Vestments of the
// Virtuous 6 pieces 23 spell power (Forever).
func TestHealingPriestEraSetBonuses(t *testing.T) {
	checkSetStats(t, "Confessor's Raiment", 999_500, 2, stats.Stats{stats.HealingPower: 22})
	checkSetStats(t, "Vestments of the Virtuous", 999_510, 6, stats.Stats{stats.SpellPower: 23})
}

// Vestments of Transcendence's 3 pieces: 15% of mana regen goes on while casting, on top of
// Meditation's.
func TestHealingPriestTranscendenceRegen(t *testing.T) {
	talents := priestTalents(map[string]int{"meditation": 3})
	two := setPriest(t, talents, "Vestments of Transcendence", 999_520, 2, stats.Stats{})
	three := setPriest(t, talents, "Vestments of Transcendence", 999_520, 3, stats.Stats{})
	if got := three.PseudoStats.SpiritRegenRateCasting - two.PseudoStats.SpiritRegenRateCasting; !near(got, 0.15, 1e-9) {
		t.Errorf("Vestments of Transcendence 3 pieces: %.2f more regen while casting, want 0.15", got)
	}
}
