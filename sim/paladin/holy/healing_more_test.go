package holy

// The Holy paladin mechanics audit finding healer_mechanics-20261104-10 listed as untested
// (orchestration/audit/reports/healer_mechanics.md, "Holy paladin"): the lower heal ranks, Holy Shock's
// and Light's Vigil's lower ranks, Benediction's reach, Divine Intellect, Holy Power, Illumination,
// Infusion of Light and Divine Favor on the paths the first tests skipped, the Oath on a downranked
// rank, Lay on Hands' ranks, Justice Armor, Libram of Infusion and Spiritual Focus as pushback. The
// expected numbers are the client's, as docs/healing_paladin.md records them.

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/paladin"
)

// holyOpts is a fight for newHolySim: n players (2 by default) of the level (60) and race (Human),
// the first with the talents, the items by slot and the aura.
type holyOpts struct {
	talents string
	level   int32
	race    proto.Race
	aura    proto.PaladinAura
	items   map[proto.ItemSlot]int32
	n       int
	// hitter gives the boss a 1 damage Fire spell (a raid damage event that never fires on its
	// own), for a test to hit the paladin with: bossHit.
	hitter bool
}

// newHolySim is holySim with more knobs and an hour-long fight, for tests that run many casts.
func newHolySim(t *testing.T, o holyOpts) (*core.Simulation, *HolyPaladin, []*core.Unit) {
	t.Helper()
	if o.level == 0 {
		o.level = 60
	}
	if o.race == proto.Race_RaceUnknown {
		o.race = proto.Race_RaceHuman
	}
	if o.n == 0 {
		o.n = 2
	}
	raid := &proto.Raid{Buffs: &proto.RaidBuffs{}, Debuffs: &proto.Debuffs{}}
	for i := 0; i < o.n; i++ {
		if i%5 == 0 {
			raid.Parties = append(raid.Parties, &proto.Party{Buffs: &proto.PartyBuffs{}})
		}
		options := &proto.PaladinOptions{}
		player := &proto.Player{
			Name:      "Paladin " + string(rune('A'+i)),
			Race:      o.race,
			Class:     proto.Class_ClassPaladin,
			Level:     o.level,
			Equipment: &proto.EquipmentSpec{},
			Rotation:  &proto.APLRotation{},
			Spec:      &proto.Player_HolyPaladin{HolyPaladin: &proto.HolyPaladin{Options: options}},
		}
		if i == 0 {
			player.TalentsString = o.talents
			options.Aura = o.aura
		}
		for slot := range proto.ItemSlot_ItemSlotRanged + 1 {
			id := int32(0)
			if i == 0 {
				id = o.items[slot]
			}
			player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: id})
		}
		raid.Parties[len(raid.Parties)-1].Players = append(raid.Parties[len(raid.Parties)-1].Players, player)
	}
	raid.NumActiveParties = int32(len(raid.Parties))
	encounter := &proto.Encounter{Duration: 3600, TrackHealth: true, Targets: []*proto.Target{core.NewDefaultTarget()}}
	if o.hitter {
		encounter.RaidDamage = []*proto.RaidDamageEvent{{School: proto.SpellSchool_SpellSchoolFire, MinDamage: 1, MaxDamage: 1, Who: proto.RaidDamageEvent_NotTanks, EverySeconds: 99999, FirstAt: 99999}}
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1, ExactDamage: true},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*HolyPaladin), sim.Raid.AllPlayerUnits
}

// noHealingPower takes the paladin's healing power and spell power to 0, then adds hp.
func noHealingPower(sim *core.Simulation, pal *HolyPaladin, hp float64) {
	pal.AddStatDynamic(sim, stats.HealingPower, hp-pal.GetStat(stats.HealingPower)-pal.GetStat(stats.SpellPower))
}

// advance runs the sim d forward.
func advance(sim *core.Simulation, d time.Duration) {
	end := sim.CurrentTime + d
	sim.AddPendingAction(&core.PendingAction{NextActionAt: end, OnAction: func(*core.Simulation) {}})
	for sim.CurrentTime < end {
		if sim.Step() {
			break
		}
	}
}

// healBooked applies s at target once and returns what booked (the spell that books the heal) healed
// there and whether it crit.
func healBooked(sim *core.Simulation, s, booked *core.Spell, target *core.Unit) (float64, bool) {
	m := &booked.SpellMetrics[target.UnitIndex]
	before, crits := m.TotalHealing, m.Crits
	s.ApplyEffects(sim, target, s)
	return m.TotalHealing - before, m.Crits > crits
}

// span heals n times and returns the lowest, highest and mean heal.
func span(sim *core.Simulation, s, booked *core.Spell, target *core.Unit, n int) (lo, hi, mean float64) {
	lo, hi = math.Inf(1), math.Inf(-1)
	for range n {
		h, _ := healBooked(sim, s, booked, target)
		lo, hi, mean = min(lo, h), max(hi, h), mean+h
	}
	return lo, hi, mean / float64(n)
}

// checkSpan checks 400 heals fall within [low, high] give or take tol and span it.
func checkSpan(t *testing.T, name string, sim *core.Simulation, s, booked *core.Spell, target *core.Unit, low, high, tol float64) {
	t.Helper()
	lo, hi, _ := span(sim, s, booked, target, 400)
	if lo < low-tol || hi > high+tol || lo > low+0.05*(high-low)+tol || hi < high-0.05*(high-low)-tol {
		t.Errorf("%s healed %.2f to %.2f, want %.2f to %.2f", name, lo, hi, low, high)
	}
}

// docAt60 is a rank's heal at 60 from the doc's columns: base at the learn level, growth a level to
// its max level, the variance to three places.
func docAt60(base, perLevel float64, learn, maxLevel int, variance float64) (float64, float64) {
	v := base + perLevel*float64(min(60, maxLevel)-learn)
	return v * (1 - variance/2), v * (1 + variance/2)
}

// TestLowerHealRanks: Holy Light 1 to 5, 7 and 8 and Flash of Light 1, 2, 4 and 5 heal the doc's
// amounts at 60 with no healing power and cost the client's mana; Holy Light 8 (10329, which the
// presets cast) also takes its full 0.714 of healing power.
func TestLowerHealRanks(t *testing.T) {
	for _, c := range []struct {
		name      string
		id        int32
		ranks     []float64 // the doc's base, per level, learn level, max level, variance
		low, high float64   // the doc's "At 60" where it gives one
		mana      float64
	}{
		{"Holy Light 1", 635, []float64{43, 0.8, 1, 5, 0.186}, 41.9, 50.5, 35},
		{"Holy Light 2", 639, []float64{83, 1.1, 6, 11, 0.169}, 0, 0, 60},
		{"Holy Light 3", 647, []float64{155, 1.7, 14, 19, 0.162}, 0, 0, 110},
		{"Holy Light 4", 1026, []float64{287, 2.4, 22, 27, 0.138}, 0, 0, 190},
		{"Holy Light 5", 1042, []float64{452, 3.1, 30, 35, 0.119}, 439.7, 495.3, 275},
		{"Holy Light 7", 10328, []float64{899, 4.6, 46, 51, 0.108}, 872.2, 971.8, 465},
		{"Holy Light 8", 10329, []float64{1217, 5.2, 54, 59, 0.108}, 1176.0, 1310.0, 580},
		{"Flash of Light 1", 19750, []float64{46, 1.0, 20, 25, 0.149}, 0, 0, 35},
		{"Flash of Light 2", 19939, []float64{66, 1.3, 26, 31, 0.136}, 0, 0, 50},
		{"Flash of Light 4", 19941, []float64{151, 1.9, 42, 47, 0.115}, 151.3, 169.7, 90},
		{"Flash of Light 5", 19942, []float64{223, 2.2, 50, 55, 0.113}, 220.8, 247.2, 115},
	} {
		low, high := docAt60(c.ranks[0], c.ranks[1], int(c.ranks[2]), int(c.ranks[3]), c.ranks[4])
		// The doc's variance is to three places: half a thousandth of the heal.
		tol := 0.0005*high/2 + 0.01
		if c.low != 0 {
			// Its "At 60" is to one place.
			if math.Abs(low-c.low) > 0.06+tol || math.Abs(high-c.high) > 0.06+tol {
				t.Fatalf("%s: the doc's columns give %.2f to %.2f, its At 60 %.1f to %.1f", c.name, low, high, c.low, c.high)
			}
			low, high, tol = c.low, c.high, 0.06+tol
		}
		sim, pal, units := newHolySim(t, holyOpts{talents: talentHolyShock})
		noCrit(sim, pal, 0)
		noHealingPower(sim, pal, 0)
		s := spell(t, pal, c.id)
		checkSpan(t, c.name, sim, s, s, units[1], low, high, tol)
		if mana, _ := castAt(t, sim, pal, s, units[1]); math.Abs(mana-c.mana) > 0.5 {
			t.Errorf("%s cost %.1f mana, want %.0f", c.name, mana, c.mana)
		}
		if c.id == 10329 {
			// Learned at 54, within 17 levels of 60: the whole 0.714.
			noHealingPower(sim, pal, 500)
			checkSpan(t, c.name+" at 500 healing", sim, s, s, units[1], c.low+0.714*500, c.high+0.714*500, tol)
		}
	}
}

// TestHolyShockRanks: Holy Shock 1 to 3 cost 160, 225 and 275 mana; ranks 2 and 3 heal 150.1 to 161.9
// and 221.3 to 238.7 at 60, with 0.85 and all of the 0.429 coefficient (rank 2 learned at 40).
func TestHolyShockRanks(t *testing.T) {
	for _, c := range []struct {
		cast, heal int32
		mana       float64
		low, high  float64
		coef       float64
	}{
		{1311606, 1311605, 160, 109.7, 118.3, 0.429 * 0.35},
		{20473, 25914, 225, 150.1, 161.9, 0.429 * 0.85},
		{20929, 25913, 275, 221.3, 238.7, 0.429},
	} {
		sim, pal, units := newHolySim(t, holyOpts{talents: talentHolyShock})
		noCrit(sim, pal, 0)
		noHealingPower(sim, pal, 0)
		hs, heal := spell(t, pal, c.cast), spell(t, pal, c.heal)
		checkSpan(t, heal.ActionID.String(), sim, hs, heal, units[1], c.low, c.high, 0.06)
		noHealingPower(sim, pal, 500)
		checkSpan(t, heal.ActionID.String()+" at 500 healing", sim, hs, heal, units[1], c.low+c.coef*500, c.high+c.coef*500, 0.06)
		hs.CD.Reset()
		if mana, _ := castAt(t, sim, pal, hs, units[1]); math.Abs(mana-c.mana) > 0.01 {
			t.Errorf("Holy Shock %d cost %.2f, want %.0f", c.cast, mana, c.mana)
		}
	}
}

// TestLightsVigilRanks: the Vigil's ranks 1 and 2 cost 730 and 1000 mana, and rank 2's party heal is
// 487 + 1.5 a level to 59, variance 0.056: 486.4 to 514.6 at 60 with no healing power.
func TestLightsVigilRanks(t *testing.T) {
	for _, c := range []struct {
		id   int32
		mana float64
	}{{1310911, 730}, {1311590, 1000}} {
		sim, pal, units := newHolySim(t, holyOpts{talents: talentVigil})
		if mana, ct := castAt(t, sim, pal, spell(t, pal, c.id), units[1]); math.Abs(mana-c.mana) > 0.5 || ct != 1500*time.Millisecond {
			t.Errorf("Light's Vigil %d cost %.1f in %s, want %.0f in 1.5 s", c.id, mana, ct, c.mana)
		}
	}

	sim, pal, units := newHolySim(t, holyOpts{talents: talentVigil})
	noCrit(sim, pal, 0)
	noHealingPower(sim, pal, 0)
	low, high := docAt60(487, 1.5, 50, 59, 0.0562)
	vigil, hs, heal := spell(t, pal, 1311590), spell(t, pal, 20930), spell(t, pal, 1311591)
	lo, hi := math.Inf(1), math.Inf(-1)
	for range 200 {
		vigil.CD.Reset()
		castAt(t, sim, pal, vigil, units[1])
		before := heal.SpellMetrics[units[1].UnitIndex].TotalHealing
		castAt(t, sim, pal, hs, units[1])
		h := heal.SpellMetrics[units[1].UnitIndex].TotalHealing - before
		lo, hi = min(lo, h), max(hi, h)
	}
	if lo < low-0.05 || hi > high+0.05 || lo > low+1.5 || hi < high-1.5 {
		t.Errorf("Light's Vigil 2 healed %.2f to %.2f, want %.2f to %.2f", lo, hi, low, high)
	}
}

// Benediction is the Retribution tree's second talent.
var (
	talentBenediction      = holyTalents(map[int]int{7: 1, 12: 1}) + "-0-05"
	talentNoBenediction    = holyTalents(map[int]int{7: 1, 12: 1})
	talentDivineIntellect5 = holyTalents(nil) + "-5"
)

// TestBenediction: 5/5 takes 10% off Holy Shock (325 to 292.5) and nothing off Divine Favor (4% of base
// mana, 60.48) or Oath of Light (8%, 120.96): the client's class masks of 20216 and 1328548 are empty, so
// Benediction's mask can't reach them. The Divine Favor and Oath assertions fail on the code that put
// paladin.benediction() on both (54.43 and 108.86).
func TestBenediction(t *testing.T) {
	for _, c := range []struct {
		talents         string
		hs, df, oath    float64
		withBenediction bool
	}{
		{talentNoBenediction, 325, 60.48, 120.96, false},
		{talentBenediction, 292.5, 60.48, 120.96, true},
	} {
		sim, pal, units := newHolySim(t, holyOpts{talents: c.talents})
		if pal.Talents.Benediction != map[bool]int32{true: 5, false: 0}[c.withBenediction] {
			t.Fatalf("Benediction %d", pal.Talents.Benediction)
		}
		if math.Abs(pal.BaseMana-1512) > 0.01 {
			t.Fatalf("base mana %.0f, want 1512", pal.BaseMana)
		}
		for _, s := range []struct {
			id     int32
			target *core.Unit
			want   float64
		}{{20930, units[1], c.hs}, {20216, &pal.Unit, c.df}, {1328548, units[1], c.oath}} {
			if mana, _ := castAt(t, sim, pal, spell(t, pal, s.id), s.target); math.Abs(mana-s.want) > 0.01 {
				t.Errorf("Benediction %d: %d cost %.2f, want %.2f", pal.Talents.Benediction, s.id, mana, s.want)
			}
		}
	}
}

// TestDivineIntellect: 5/5 (the Protection tree's first talent) is 10% more Intellect, and the mana
// that comes with it.
func TestDivineIntellect(t *testing.T) {
	_, off, _ := newHolySim(t, holyOpts{talents: talentHolyShock})
	_, on, _ := newHolySim(t, holyOpts{talents: talentDivineIntellect5})
	if on.Talents.DivineIntellect != 5 {
		t.Fatalf("Divine Intellect %d", on.Talents.DivineIntellect)
	}
	i0, i1 := off.GetStat(stats.Intellect), on.GetStat(stats.Intellect)
	if math.Abs(i1/i0-1.10) > 1e-9 {
		t.Errorf("Divine Intellect 5 gives %.2f Intellect against %.2f, want 10%% more", i1, i0)
	}
	if d := on.MaxMana() - off.MaxMana(); math.Abs(d-15*(i1-i0)) > 0.01 {
		t.Errorf("Divine Intellect 5 gives %.2f more mana for %.2f Intellect, want 15 a point", d, i1-i0)
	}
}

// critRatio is a crit's mean heal over a normal one's, at 0 and 100 crit in turn.
func critRatio(sim *core.Simulation, pal *HolyPaladin, s, booked *core.Spell, target *core.Unit) float64 {
	var hit, crit, nh, nc float64
	for i := range 4000 {
		noCrit(sim, pal, float64(100*(i%2)))
		h, c := healBooked(sim, s, booked, target)
		if c {
			crit, nc = crit+h, nc+1
		} else {
			hit, nh = hit+h, nh+1
		}
	}
	return (crit / nc) / (hit / nh)
}

// TestHolyPowerShockAndVigil: Holy Power 5/5 makes Holy Shock's heal and Light's Vigil's heal crit for
// 225% (150% without), and leaves Holy Shock's damage crits at 150%: since the hotfix it is crit
// healing only.
func TestHolyPowerShockAndVigil(t *testing.T) {
	for _, c := range []struct {
		talents string
		want    float64
	}{{talentVigil, 1.5}, {holyTalents(map[int]int{16: 5, 17: 1}), 2.25}} {
		sim, pal, units := newHolySim(t, holyOpts{talents: c.talents})
		hs, vigilHeal := spell(t, pal, 20930), spell(t, pal, 1311596)
		if r := critRatio(sim, pal, hs, spell(t, pal, 25903), units[1]); math.Abs(r-c.want) > 0.015 {
			t.Errorf("Holy Shock's heal crits for %.3f times a hit, want %.2f", r, c.want)
		}
		if r := critRatio(sim, pal, vigilHeal, vigilHeal, units[1]); math.Abs(r-c.want) > 0.015 {
			t.Errorf("Light's Vigil's heal crits for %.3f times a hit, want %.2f", r, c.want)
		}

		// Holy Shock at the boss: every landing a crit at 100 crit, none at 0.
		pal.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
		boss := pal.CurrentTarget
		m := &hs.SpellMetrics[boss.UnitIndex]
		perLanding := func(crit float64) float64 {
			noCrit(sim, pal, crit)
			dmg := m.TotalDamage
			landed := m.Hits + m.ResistedHits + m.Crits + m.ResistedCrits
			crits := m.Crits + m.ResistedCrits
			for range 2000 {
				hs.ApplyEffects(sim, boss, hs)
			}
			nl := m.Hits + m.ResistedHits + m.Crits + m.ResistedCrits - landed
			nc := m.Crits + m.ResistedCrits - crits
			if (crit == 0 && nc != 0) || (crit == 100 && nc != nl) {
				t.Errorf("%d crits of %d landings at %.0f crit", nc, nl, crit)
			}
			return (m.TotalDamage - dmg) / float64(nl)
		}
		if r := perLanding(100) / perLanding(0); math.Abs(r-1.5) > 0.02 {
			t.Errorf("Holy Shock's damage crits for %.3f times a hit with %s, want 1.5", r, c.talents)
		}
	}
}

// TestIlluminationVigilAndLibram: Illumination 5/5 gives back half the Vigil's base cost (670 for rank
// 3) once a detonation that crits, not once a player healed, and nothing on one that doesn't; and half
// Holy Light 9's base cost (330), not of its cost under Libram of Economy (627, which would be 313.5).
func TestIlluminationVigilAndLibram(t *testing.T) {
	sim, pal, units := newHolySim(t, holyOpts{talents: holyTalents(map[int]int{4: 2, 8: 5, 17: 1}), n: 5})
	vigilHeal := spell(t, pal, 1311596)
	pal.SpendMana(sim, 3000, pal.GetManaNotCastingMetrics())
	for _, c := range []struct {
		crit float64
		want float64
	}{{0, 0}, {100, 670}} {
		noCrit(sim, pal, c.crit)
		crits := 0
		before := pal.CurrentMana()
		for _, u := range units {
			crits -= int(vigilHeal.SpellMetrics[u.UnitIndex].Crits)
		}
		vigilHeal.ApplyEffects(sim, units[1], vigilHeal)
		for _, u := range units {
			crits += int(vigilHeal.SpellMetrics[u.UnitIndex].Crits)
		}
		if got := pal.CurrentMana() - before; math.Abs(got-c.want) > 0.01 || (c.crit == 100 && crits != 5) {
			t.Errorf("a Vigil at %.0f crit (%d crits) gave back %.2f, want %.0f", c.crit, crits, got, c.want)
		}
	}

	sim, pal, units = newHolySim(t, holyOpts{talents: talentIllumination, items: map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotRanged: libramOfEconomy}})
	hl := spell(t, pal, 25292)
	if mana, _ := castAt(t, sim, pal, hl, units[1]); math.Abs(mana-627) > 0.5 {
		t.Fatalf("Holy Light 9 under Libram of Economy cost %.1f, want 627", mana)
	}
	pal.SpendMana(sim, 2000, pal.GetManaNotCastingMetrics())
	noCrit(sim, pal, 100)
	before := pal.CurrentMana()
	hl.ApplyEffects(sim, units[1], hl)
	if got := pal.CurrentMana() - before; math.Abs(got-330) > 0.01 {
		t.Errorf("a Holy Light 9 crit under Libram of Economy gave back %.2f, want 330", got)
	}
}

// TestInfusionOfLightMore: 1/2 takes 0.5 sec off the next Holy Light; a Holy Shock damage crit at a mob
// gives it (the proc flags take harmful spells); the buff lasts 15 sec.
func TestInfusionOfLightMore(t *testing.T) {
	sim, pal, units := newHolySim(t, holyOpts{talents: holyTalents(map[int]int{11: 1})})
	hl, fol := spell(t, pal, 25292), spell(t, pal, 19943)
	noCrit(sim, pal, 100)
	castAt(t, sim, pal, fol, units[1])
	noCrit(sim, pal, 0)
	if _, ct := castAt(t, sim, pal, hl, units[1]); ct != 2000*time.Millisecond {
		t.Errorf("Holy Light after a Flash of Light crit at Infusion 1 took %s, want 2 s", ct)
	}

	sim, pal, units = newHolySim(t, holyOpts{talents: talentInfusion})
	hl, hs := spell(t, pal, 25292), spell(t, pal, 20930)
	pal.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
	noCrit(sim, pal, 100)
	boss := pal.CurrentTarget
	crits := hs.SpellMetrics[boss.UnitIndex].Crits
	castAt(t, sim, pal, hs, boss)
	if hs.SpellMetrics[boss.UnitIndex].Crits != crits+1 {
		t.Fatalf("Holy Shock at the boss didn't crit")
	}
	noCrit(sim, pal, 0)
	if _, ct := castAt(t, sim, pal, hl, units[1]); ct != 1500*time.Millisecond {
		t.Errorf("Holy Light after a Holy Shock damage crit took %s, want 1.5 s", ct)
	}

	for _, c := range []struct {
		wait time.Duration
		want time.Duration
	}{{14900 * time.Millisecond, 1500 * time.Millisecond}, {15100 * time.Millisecond, 2500 * time.Millisecond}} {
		sim, pal, units := newHolySim(t, holyOpts{talents: talentInfusion})
		hl, fol := spell(t, pal, 25292), spell(t, pal, 19943)
		noCrit(sim, pal, 100)
		castAt(t, sim, pal, fol, units[1])
		noCrit(sim, pal, 0)
		advance(sim, c.wait)
		if _, ct := castAt(t, sim, pal, hl, units[1]); ct != c.want {
			t.Errorf("Holy Light %s after the Flash of Light crit took %s, want %s", c.wait, ct, c.want)
		}
	}
}

// TestDivineFavorShockDamageAndVigil: Divine Favor makes Holy Shock's damage crit and goes with it; the
// Vigil's heal neither crits from it nor spends it.
func TestDivineFavorShockDamageAndVigil(t *testing.T) {
	sim, pal, units := newHolySim(t, holyOpts{talents: holyTalents(map[int]int{7: 1, 17: 1}), n: 5})
	noCrit(sim, pal, 0)
	pal.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
	aura := pal.GetAura("Divine Favor")
	hs := spell(t, pal, 20930)
	boss := pal.CurrentTarget
	m := &hs.SpellMetrics[boss.UnitIndex]
	for range 5 {
		aura.Activate(sim)
		crits := m.Crits + m.ResistedCrits
		hs.ApplyEffects(sim, boss, hs)
		if m.Crits+m.ResistedCrits != crits+1 || aura.IsActive() {
			t.Errorf("Holy Shock at the boss under Divine Favor: crit %v, buff still up %v", m.Crits+m.ResistedCrits > crits, aura.IsActive())
		}
	}

	vigilHeal := spell(t, pal, 1311596)
	aura.Activate(sim)
	for range 20 {
		for _, u := range units {
			if _, crit := healBooked(sim, vigilHeal, vigilHeal, u); crit {
				t.Fatalf("the Vigil's heal crit under Divine Favor at 0 crit")
			}
		}
	}
	if !aura.IsActive() {
		t.Errorf("the Vigil's heal spent Divine Favor")
	}
}

// TestOathDownranked: the Oath's bonus on a rank Forever's downranking cuts keeps the rank's share of
// it: Holy Light 6 (learned 38) 75% of 400, Flash of Light 3 (learned 34) 55% of 115; Holy Light 8
// (learned 54) all 400.
func TestOathDownranked(t *testing.T) {
	sim, pal, units := newHolySim(t, holyOpts{talents: talentOath, n: 3})
	noCrit(sim, pal, 0)
	castAt(t, sim, pal, spell(t, pal, 1328548), units[1])
	for _, c := range []struct {
		id    int32
		bonus float64
	}{{3472, 300}, {19940, 63.25}, {10329, 400}} {
		s := spell(t, pal, c.id)
		_, _, on := span(sim, s, s, units[1], 3000)
		_, _, off := span(sim, s, s, units[2], 3000)
		if d := on - off; math.Abs(d-c.bonus) > 0.02*c.bonus {
			t.Errorf("%d heals %.2f more on the Oath's target, want %.2f", c.id, d, c.bonus)
		}
	}
}

// TestLayOnHandsRanks: a paladin knows only its top rank: 633 from 10 (no mana back), 2800 from 30 (250
// back), 10310 from 50 (550 back), none at 9; it takes no five second rule.
func TestLayOnHandsRanks(t *testing.T) {
	for _, c := range []struct {
		level int32
		id    int32
		back  float64
	}{{9, 0, 0}, {10, 633, 0}, {29, 633, 0}, {30, 2800, 250}, {49, 2800, 250}, {50, 10310, 550}, {60, 10310, 550}} {
		sim, pal, units := newHolySim(t, holyOpts{level: c.level})
		for _, id := range []int32{633, 2800, 10310} {
			if has := pal.GetSpell(core.ActionID{SpellID: id}) != nil; has != (id == c.id) {
				t.Errorf("level %d knows Lay on Hands %d: %v", c.level, id, has)
			}
		}
		if c.id == 0 {
			continue
		}
		other := sim.Raid.Parties[0].Players[1].GetCharacter()
		other.SpendMana(sim, other.CurrentMana(), other.GetManaNotCastingMetrics())
		fsr := pal.PseudoStats.FiveSecondRuleRefreshTime
		castAt(t, sim, pal, spell(t, pal, c.id), units[1])
		if pal.CurrentMana() != 0 || math.Abs(other.CurrentMana()-c.back) > 0.01 {
			t.Errorf("level %d Lay on Hands left %.0f mana, gave %.0f, want 0 and %.0f", c.level, pal.CurrentMana(), other.CurrentMana(), c.back)
		}
		if pal.PseudoStats.FiveSecondRuleRefreshTime != fsr {
			t.Errorf("level %d Lay on Hands started the five second rule (to %s)", c.level, pal.PseudoStats.FiveSecondRuleRefreshTime)
		}
	}
	// A heal with a mana cost does start it.
	sim, pal, units := newHolySim(t, holyOpts{})
	castAt(t, sim, pal, spell(t, pal, 25292), units[1])
	if pal.PseudoStats.FiveSecondRuleRefreshTime <= sim.CurrentTime {
		t.Errorf("Holy Light didn't start the five second rule")
	}
}

// TestVigilNoHealingLight: Healing Light 3/3 doesn't raise the Vigil's heal (its aura's mask is Holy
// Light, Flash of Light and Holy Shock's heal).
func TestVigilNoHealingLight(t *testing.T) {
	mean := func(talents string) float64 {
		sim, pal, units := newHolySim(t, holyOpts{talents: talents})
		noCrit(sim, pal, 0)
		noHealingPower(sim, pal, 0)
		h := spell(t, pal, 1311596)
		_, _, m := span(sim, h, h, units[1], 2000)
		return m
	}
	if off, on := mean(talentVigil), mean(holyTalents(map[int]int{3: 3, 17: 1})); math.Abs(on/off-1) > 0.003 {
		t.Errorf("the Vigil heals %.4f times as much with Healing Light 3, want 1", on/off)
	}
}

// Justice Armor, Forever's Tier 1 Holy set (2107): gloves, boots, legs, chest, shoulders and head.
var justiceArmor = []struct {
	slot proto.ItemSlot
	id   int32
}{
	{proto.ItemSlot_ItemSlotHands, 280931},
	{proto.ItemSlot_ItemSlotFeet, 280932},
	{proto.ItemSlot_ItemSlotLegs, 280933},
	{proto.ItemSlot_ItemSlotChest, 280934},
	{proto.ItemSlot_ItemSlotShoulder, 280935},
	{proto.ItemSlot_ItemSlotHead, 280936},
}

// TestJusticeArmor: 2 pieces +10 Spirit, 4 +26 healing, 5 a 9 sec Holy Shock cooldown on every rank.
// A Dwarf, so the Spirit isn't scaled by The Human Spirit.
func TestJusticeArmor(t *testing.T) {
	type got struct{ spirit, healing float64 }
	var base got
	for _, pieces := range []int{1, 2, 3, 4, 5} {
		items := map[proto.ItemSlot]int32{}
		for _, p := range justiceArmor[:pieces] {
			items[p.slot] = p.id
		}
		sim, pal, units := newHolySim(t, holyOpts{talents: talentHolyShock, race: proto.Race_RaceDwarf, items: items})
		g := got{pal.GetStat(stats.Spirit), pal.GetStat(stats.HealingPower)}
		if pieces == 1 {
			base = g
		}
		want := got{base.spirit, base.healing}
		if pieces >= 2 {
			want.spirit += 10
		}
		if pieces >= 4 {
			want.healing += 26
		}
		if math.Abs(g.spirit-want.spirit) > 0.01 || math.Abs(g.healing-want.healing) > 0.01 {
			t.Errorf("%d pieces: %.1f Spirit and %.0f healing, want %.1f and %.0f", pieces, g.spirit, g.healing, want.spirit, want.healing)
		}
		cd := 10 * time.Second
		if pieces >= 5 {
			cd = 9 * time.Second
		}
		castAt(t, sim, pal, spell(t, pal, 20930), units[1])
		for _, id := range []int32{20930, 20929, 20473, 1311606} {
			if s := spell(t, pal, id); s.CD.TimeToReady(sim) != cd {
				t.Errorf("%d pieces: Holy Shock %d ready in %s, want %s", pieces, id, s.CD.TimeToReady(sim), cd)
			}
		}
	}
}

// TestLibramOfInfusionHeal: Libram of Infusion's 6% Holy Shock crit reaches the heal.
func TestLibramOfInfusionHeal(t *testing.T) {
	rate := func(ranged int32) float64 {
		sim, pal, units := newHolySim(t, holyOpts{talents: talentHolyShock, items: map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotRanged: ranged}})
		noCrit(sim, pal, 0)
		hs, heal := spell(t, pal, 20930), spell(t, pal, 25903)
		n := 0
		for range 6000 {
			if _, crit := healBooked(sim, hs, heal, units[1]); crit {
				n++
			}
		}
		return float64(n) / 6000
	}
	if r := rate(0); r != 0 {
		t.Errorf("Holy Shock's heal crit %.3f at 0 crit with no libram", r)
	}
	if r := rate(paladin.LibramOfInfusion); math.Abs(r-0.06) > 0.012 {
		t.Errorf("Holy Shock's heal crit %.3f with Libram of Infusion, want 0.06", r)
	}
}

// TestSpiritualFocusPushback: a hit 1 sec into a Holy Light pushes it back 1 sec. Spiritual Focus 5/5
// stops 70% of them, Concentration Aura 35%, and the two together (70 + 35, capped at 100%) all.
func TestSpiritualFocusPushback(t *testing.T) {
	for _, c := range []struct {
		name    string
		talents string
		aura    proto.PaladinAura
		want    float64 // the share of casts pushed back
	}{
		{"no talent", "", 0, 1},
		{"Spiritual Focus 5", talentSpiritualFocus, 0, 0.3},
		{"Concentration Aura", "", proto.PaladinAura_ConcentrationAura, 0.65},
		{"both", talentSpiritualFocus, proto.PaladinAura_ConcentrationAura, 0},
	} {
		sim, pal, units := newHolySim(t, holyOpts{talents: c.talents, aura: c.aura, hitter: true})
		swing := pal.CurrentTarget.GetSpell(core.ActionID{OtherID: proto.OtherAction_OtherActionDamageTaken, Tag: 1})
		if swing == nil {
			t.Fatalf("the boss has no spell to hit with")
		}
		hl := spell(t, pal, 25292)
		const n = 300
		pushed := 0
		for range n {
			pal.SetGCDTimer(sim, sim.CurrentTime)
			pal.AddMana(sim, pal.MaxMana(), pal.GetManaNotCastingMetrics())
			start := sim.CurrentTime
			if !hl.Cast(sim, units[1]) {
				t.Fatalf("Holy Light didn't cast")
			}
			var end time.Duration
			sim.AddPendingAction(&core.PendingAction{NextActionAt: start + time.Second, OnAction: func(sim *core.Simulation) {
				swing.CalcAndDealDamage(sim, &pal.Unit, 1, swing.OutcomeAlwaysHit)
				if pal.CurrentHealth() < 100 {
					t.Fatalf("the paladin is dying")
				}
				end = pal.Hardcast.Expires
			}})
			for sim.CurrentTime < start+time.Second || pal.IsCasting(sim) {
				if sim.Step() {
					t.Fatalf("the fight ended")
				}
			}
			switch end - start {
			case 3500 * time.Millisecond:
				pushed++
			case 2500 * time.Millisecond:
			default:
				t.Fatalf("%s: Holy Light ends %s after its start", c.name, end-start)
			}
		}
		if r := float64(pushed) / n; math.Abs(r-c.want) > 0.07 || (c.want == 0) != (pushed == 0) || (c.want == 1) != (pushed == n) {
			t.Errorf("%s: %.3f of the Holy Lights pushed back, want %.2f", c.name, r, c.want)
		}
	}
}
