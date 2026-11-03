package holy

// Every Holy paladin heal and healing talent (docs/healing_paladin.md): each heal heals its target by
// the client's numbers, costs its mana, takes its cast time and cooldown, and each talent on against off
// moves exactly what it should.

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

func init() {
	RegisterHolyPaladin()
}

// holyTalents is a Holy talent string with the ranks given by position in the tree (1 is the first):
// 1 Spiritual Focus, 2 Divine Strength, 3 Healing Light, 4 Reverence, 5 Improved Seals, 6 Purifying
// Power, 7 Divine Favor, 8 Illumination, 9 Holy Conduit, 10 Divine Precision, 11 Infusion of Light, 12
// Oath of Light, 13 Consecrated Ground, 14 Holy Shock, 15 Boundless Light, 16 Holy Power, 17 Light's
// Vigil. Holy Shock (14) is always in, for its heal.
func holyTalents(ranks map[int]int) string {
	digits := []byte("00000000000000000")
	digits[13] = '1'
	for pos, rank := range ranks {
		digits[pos-1] = byte('0' + rank)
	}
	return string(digits)
}

var (
	talentHolyShock      = holyTalents(nil)
	talentSpiritualFocus = holyTalents(map[int]int{1: 5, 17: 1})
	talentHealingLight   = holyTalents(map[int]int{3: 3})
	talentDivineFavor    = holyTalents(map[int]int{7: 1})
	talentIllumination   = holyTalents(map[int]int{4: 2, 8: 5})
	talentIllumination1  = holyTalents(map[int]int{4: 2, 8: 1})
	talentInfusion       = holyTalents(map[int]int{11: 2})
	talentOath           = holyTalents(map[int]int{12: 1})
	talentOathBoundless  = holyTalents(map[int]int{12: 1, 15: 2})
	talentHolyPower      = holyTalents(map[int]int{16: 5})
	talentVigil          = holyTalents(map[int]int{17: 1})
)

const libramOfEconomy = 272436

// holySim starts a Forever healing fight of n level 60 Human Holy paladins in parties of five, naked, the
// first with talents and the ranged item, against a level 63 boss nobody tanks. It returns the first
// paladin and every player's unit, before the first action.
func holySim(t *testing.T, talents string, ranged int32, n int) (*core.Simulation, *HolyPaladin, []*core.Unit) {
	t.Helper()
	raid := &proto.Raid{Buffs: &proto.RaidBuffs{}, Debuffs: &proto.Debuffs{}}
	for i := 0; i < n; i++ {
		if i%5 == 0 {
			raid.Parties = append(raid.Parties, &proto.Party{Buffs: &proto.PartyBuffs{}})
		}
		player := &proto.Player{
			Name:      "Paladin " + string(rune('A'+i)),
			Race:      proto.Race_RaceHuman,
			Class:     proto.Class_ClassPaladin,
			Level:     60,
			Equipment: &proto.EquipmentSpec{},
			Rotation:  &proto.APLRotation{},
			Spec:      &proto.Player_HolyPaladin{HolyPaladin: &proto.HolyPaladin{Options: &proto.PaladinOptions{}}},
		}
		if i == 0 {
			player.TalentsString = talents
		}
		for slot := range proto.ItemSlot_ItemSlotRanged + 1 {
			id := int32(0)
			if i == 0 && slot == proto.ItemSlot_ItemSlotRanged {
				id = ranged
			}
			player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: id})
		}
		raid.Parties[len(raid.Parties)-1].Players = append(raid.Parties[len(raid.Parties)-1].Players, player)
	}
	raid.NumActiveParties = int32(len(raid.Parties))
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Duration: 60, TrackHealth: true, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1, ExactDamage: true},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	pal := sim.Raid.Parties[0].Players[0].(*HolyPaladin)
	return sim, pal, sim.Raid.AllPlayerUnits
}

// noCrit takes the paladin's spell crit to 0 (bonus adds some back).
func noCrit(sim *core.Simulation, pal *HolyPaladin, bonus float64) {
	pal.AddStatDynamic(sim, stats.SpellCrit, bonus-pal.GetStat(stats.SpellCrit))
}

// healOf applies spell at target once and returns what it healed there (before overhealing) and
// whether it crit. Holy Shock's heal is its heal spell's (Holy Shock rank 4 heals with 25903).
func healOf(sim *core.Simulation, spell *core.Spell, target *core.Unit) (float64, bool) {
	booked := spell
	if spell.ActionID.SpellID == 20930 {
		booked = spell.Unit.GetSpell(core.ActionID{SpellID: 25903})
	}
	m := &booked.SpellMetrics[target.UnitIndex]
	before, crits := m.TotalHealing, m.Crits
	spell.ApplyEffects(sim, target, spell)
	return m.TotalHealing - before, m.Crits > crits
}

func spell(t *testing.T, pal *HolyPaladin, id int32) *core.Spell {
	t.Helper()
	s := pal.GetSpell(core.ActionID{SpellID: id})
	if s == nil {
		t.Fatalf("no spell %d", id)
	}
	return s
}

// within checks every one of n heals of spell at target falls within [low, high] and that they span it.
func within(t *testing.T, sim *core.Simulation, s *core.Spell, target *core.Unit, low, high float64) {
	t.Helper()
	lo, hi := math.Inf(1), math.Inf(-1)
	for range 400 {
		h, _ := healOf(sim, s, target)
		lo, hi = min(lo, h), max(hi, h)
	}
	if lo < low-0.01 || hi > high+0.01 {
		t.Errorf("%s healed %.2f to %.2f, want %.2f to %.2f", s.ActionID, lo, hi, low, high)
	}
	if lo > low+0.05*(high-low) || hi < high-0.05*(high-low) {
		t.Errorf("%s healed %.2f to %.2f, which doesn't span %.2f to %.2f", s.ActionID, lo, hi, low, high)
	}
}

// TestHealRanks checks the rank tables against the Forever client (1.60.1.70753) and which ranks a level
// 60 paladin has.
func TestHealRanks(t *testing.T) {
	for _, c := range []struct {
		name      string
		low, high float64
		at        func() (float64, float64)
	}{
		// 1580 at 60, 10.71% variance.
		{"Holy Light 9", 1495.36, 1664.64, func() (float64, float64) { return paladin.HolyLightRanks[9].At(60) }},
		// 646 at 38 plus 3.8 a level to 43: the level 38 footage's 610 to 682.
		{"Holy Light 6 at 38", 610.16, 681.84, func() (float64, float64) { return paladin.HolyLightRanks[6].At(38) }},
		{"Holy Light 6 at 60", 1.0 * (646 + 3.8*5) * (1 - 0.11096075922/2), (646 + 3.8*5) * (1 + 0.11096075922/2), func() (float64, float64) { return paladin.HolyLightRanks[6].At(60) }},
		// 303 at 58 plus 2.6 a level.
		{"Flash of Light 6", 308.2 * (1 - 0.11019284278/2), 308.2 * (1 + 0.11019284278/2), func() (float64, float64) { return paladin.FlashOfLightRanks[6].At(60) }},
		// The level 38 footage's 101 to 114.
		{"Flash of Light 3 at 38", 101.12, 113.68, func() (float64, float64) { return paladin.FlashOfLightRanks[3].At(38) }},
		// The talent tooltip's 110 to 118.
		{"Holy Shock 1", 109.70, 118.30, func() (float64, float64) { return paladin.HolyShockHealRanks[1].At(60) }},
		{"Holy Shock 4", 307.37, 332.63, func() (float64, float64) { return paladin.HolyShockHealRanks[4].At(60) }},
		// The talent tooltip's rank 1 at 60: 326 to 344.
		{"Light's Vigil 1", 325.39, 344.21, func() (float64, float64) { return paladin.LightsVigilHealRanks[1].At(60) }},
		{"Light's Vigil 3", 684.21, 723.79, func() (float64, float64) { return paladin.LightsVigilHealRanks[3].At(60) }},
	} {
		low, high := c.at()
		if math.Abs(low-c.low) > 0.01 || math.Abs(high-c.high) > 0.01 {
			t.Errorf("%s heals %.2f to %.2f, want %.2f to %.2f", c.name, low, high, c.low, c.high)
		}
	}

	_, pal, _ := holySim(t, "", 0, 1)
	count := func(spells []*core.Spell) (n int) {
		for _, s := range spells {
			if s != nil {
				n++
			}
		}
		return
	}
	if hl, fol := count(pal.HolyLight), count(pal.FlashOfLight); hl != 9 || fol != 6 {
		t.Errorf("%d Holy Light and %d Flash of Light ranks at 60, want 9 and 6", hl, fol)
	}
}

// TestHealAmounts heals another player with each heal at 0 crit and checks the amount: base plus the
// client's coefficient times healing power, Forever's downranking taking a rank learned more than 17
// levels below us down 5% a level.
func TestHealAmounts(t *testing.T) {
	sim, pal, units := holySim(t, talentHolyShock, 0, 2)
	noCrit(sim, pal, 0)
	pal.AddStatDynamic(sim, stats.HealingPower, 500-pal.GetStat(stats.HealingPower)-pal.GetStat(stats.SpellPower))
	hp := 500.0
	target := units[1]
	for _, c := range []struct {
		id        int32
		low, high float64
	}{
		{25292, 1495.36 + 0.714*hp, 1664.64 + 0.714*hp},
		// Learned at 38: 75% of its coefficient at 60.
		{3472, 665 * (1 - 0.11096075922/2), 665 * (1 + 0.11096075922/2)},
		{19943, 308.2*(1-0.11019284278/2) + 0.429*hp, 308.2*(1+0.11019284278/2) + 0.429*hp},
		// Holy Shock rank 4's heal, at an ally.
		{20930, 307.37 + 0.429*hp, 332.63 + 0.429*hp},
	} {
		s := spell(t, pal, c.id)
		if c.id == 3472 {
			c.low += 0.714 * 0.75 * hp
			c.high += 0.714 * 0.75 * hp
		}
		within(t, sim, s, target, c.low, c.high)
	}

	// A heal at a mob lands on the paladin.
	hl := spell(t, pal, 25292)
	before := hl.SpellMetrics[pal.UnitIndex].TotalHealing
	hl.ApplyEffects(sim, pal.CurrentTarget, hl)
	if hl.SpellMetrics[pal.UnitIndex].TotalHealing <= before {
		t.Errorf("Holy Light at the boss didn't heal the paladin")
	}
	// Holy Shock at a mob deals damage, and heals nobody.
	hs := spell(t, pal, 20930)
	dmg := hs.SpellMetrics[pal.CurrentTarget.UnitIndex].TotalDamage
	healed := hs.SpellMetrics[pal.UnitIndex].TotalHealing
	for range 20 {
		hs.ApplyEffects(sim, pal.CurrentTarget, hs)
	}
	if hs.SpellMetrics[pal.CurrentTarget.UnitIndex].TotalDamage <= dmg || hs.SpellMetrics[pal.UnitIndex].TotalHealing != healed {
		t.Errorf("Holy Shock at the boss: %.0f damage, healing moved", hs.SpellMetrics[pal.CurrentTarget.UnitIndex].TotalDamage-dmg)
	}
}

// castAt casts spell at target now with the global cooldown free, and returns the mana it took once the
// cast ends (it runs the sim to then).
func castAt(t *testing.T, sim *core.Simulation, pal *HolyPaladin, s *core.Spell, target *core.Unit) (mana float64, castTime time.Duration) {
	t.Helper()
	pal.SetGCDTimer(sim, sim.CurrentTime)
	pal.AddMana(sim, pal.MaxMana(), pal.GetManaNotCastingMetrics())
	before := pal.CurrentMana()
	start := sim.CurrentTime
	if !s.Cast(sim, target) {
		t.Fatalf("%s didn't cast", s.ActionID)
	}
	castTime = s.CurCast.CastTime
	end := start + castTime
	for sim.CurrentTime < end || pal.IsCasting(sim) {
		if sim.Step() {
			break
		}
	}
	return before - pal.CurrentMana(), castTime
}

// TestHealCostsAndCasts checks each heal's mana, cast time and cooldown: Holy Light 9 660 mana in 2.5
// sec (627 with Libram of Economy), Flash of Light 6 140 in 1.5, Holy Shock 4 325 (Benediction is a
// Retribution talent, none here) instant on a 10 sec cooldown that every rank shares.
func TestHealCostsAndCasts(t *testing.T) {
	for _, c := range []struct {
		name     string
		ranged   int32
		id       int32
		mana     float64
		castTime time.Duration
	}{
		{"Holy Light 9", 0, 25292, 660, 2500 * time.Millisecond},
		{"Holy Light 9, Libram of Economy", libramOfEconomy, 25292, 627, 2500 * time.Millisecond},
		{"Holy Light 6", 0, 3472, 365, 2500 * time.Millisecond},
		{"Flash of Light 6", 0, 19943, 140, 1500 * time.Millisecond},
		{"Flash of Light 6, Libram of Economy", libramOfEconomy, 19943, 140, 1500 * time.Millisecond},
		{"Holy Shock 4", 0, 20930, 325, 0},
	} {
		sim, pal, units := holySim(t, talentHolyShock, c.ranged, 2)
		mana, castTime := castAt(t, sim, pal, spell(t, pal, c.id), units[1])
		// Mana regen ticks may land during the cast.
		if math.Abs(mana-c.mana) > 0.5 || castTime != c.castTime {
			t.Errorf("%s cost %.1f mana in %s, want %.0f in %s", c.name, mana, castTime, c.mana, c.castTime)
		}
	}

	sim, pal, units := holySim(t, talentHolyShock, 0, 2)
	castAt(t, sim, pal, spell(t, pal, 20930), units[1])
	for _, id := range []int32{20930, 20929, 1311606} {
		if s := spell(t, pal, id); s.CD.TimeToReady(sim) != 10*time.Second {
			t.Errorf("Holy Shock %d ready in %s after rank 4, want 10 s", id, s.CD.TimeToReady(sim))
		}
	}
}

// TestHealingLight: 3/3 heals 12% more with Holy Light, Flash of Light and Holy Shock, not Light's Vigil.
func TestHealingLight(t *testing.T) {
	avg := func(talents string, id int32) float64 {
		sim, pal, units := holySim(t, talents, 0, 2)
		noCrit(sim, pal, 0)
		s := spell(t, pal, id)
		total := 0.0
		for range 2000 {
			h, _ := healOf(sim, s, units[1])
			total += h
		}
		return total / 2000
	}
	for _, id := range []int32{25292, 19943, 20930} {
		off, on := avg(talentHolyShock, id), avg(talentHealingLight, id)
		if r := on / off; math.Abs(r-1.12) > 0.004 {
			t.Errorf("Healing Light 3 heals %.4f times as much with %d, want 1.12", r, id)
		}
	}
}

// TestHolyPower: 5/5 makes a heal's crit 225% of a normal one (150% without), aura 50 on the whole crit.
func TestHolyPower(t *testing.T) {
	ratio := func(talents string) float64 {
		sim, pal, units := holySim(t, talents, 0, 2)
		s := spell(t, pal, 19942) // Flash of Light 5: no growth past 55, no Oath
		var hit, crit, nh, nc float64
		for i := range 4000 {
			noCrit(sim, pal, float64(100*(i%2)))
			h, c := healOf(sim, s, units[1])
			if c {
				crit, nc = crit+h, nc+1
			} else {
				hit, nh = hit+h, nh+1
			}
		}
		return (crit / nc) / (hit / nh)
	}
	if r := ratio(talentHolyShock); math.Abs(r-1.5) > 0.01 {
		t.Errorf("a crit heals %.3f times a hit without Holy Power, want 1.5", r)
	}
	if r := ratio(talentHolyPower); math.Abs(r-2.25) > 0.015 {
		t.Errorf("a crit heals %.3f times a hit with Holy Power 5, want 2.25", r)
	}
}

// TestDivineFavor: the next Holy Light, Flash of Light or Holy Shock crits, the buff goes with it, and it
// costs 4% of base mana (60.48) on a 2 min cooldown off the global cooldown.
func TestDivineFavor(t *testing.T) {
	sim, pal, units := holySim(t, talentDivineFavor, 0, 2)
	noCrit(sim, pal, 0)
	df := spell(t, pal, 20216)
	mana, _ := castAt(t, sim, pal, df, &pal.Unit)
	if math.Abs(mana-0.04*pal.BaseMana) > 0.01 || math.Abs(mana-60.48) > 0.01 {
		t.Errorf("Divine Favor cost %.2f, want 4%% of base mana %.0f", mana, pal.BaseMana)
	}
	if df.CD.TimeToReady(sim) != 2*time.Minute || df.DefaultCast.GCD != 0 {
		t.Errorf("Divine Favor ready in %s, GCD %s", df.CD.TimeToReady(sim), df.DefaultCast.GCD)
	}
	aura := pal.GetAura("Divine Favor")
	for _, id := range []int32{19943, 25292, 20930} {
		aura.Activate(sim)
		if _, crit := healOf(sim, spell(t, pal, id), units[1]); !crit || aura.IsActive() {
			t.Errorf("%d under Divine Favor: crit %v, buff still up %v", id, crit, aura.IsActive())
		}
	}
	// Without it, none of 200 heals crit at 0 crit.
	for range 200 {
		if _, crit := healOf(sim, spell(t, pal, 25292), units[1]); crit {
			t.Fatalf("a crit at 0 crit without Divine Favor")
		}
	}
}

// TestIllumination: 5/5 gives back 50% of a heal's base cost on every crit (330 for Holy Light 9, 162.5
// for Holy Shock 4), and nothing on a normal heal.
func TestIllumination(t *testing.T) {
	for _, c := range []struct {
		id     int32
		refund float64
	}{{25292, 330}, {19943, 70}, {20930, 162.5}} {
		sim, pal, units := holySim(t, talentIllumination, 0, 2)
		s := spell(t, pal, c.id)
		pal.SpendMana(sim, 2000, pal.GetManaNotCastingMetrics())
		for _, crit := range []bool{false, true} {
			noCrit(sim, pal, map[bool]float64{true: 100, false: 0}[crit])
			before := pal.CurrentMana()
			s.ApplyEffects(sim, units[1], s)
			want := map[bool]float64{true: c.refund, false: 0}[crit]
			if got := pal.CurrentMana() - before; math.Abs(got-want) > 0.01 {
				t.Errorf("%d crit %v gave back %.2f mana, want %.2f", c.id, crit, got, want)
			}
		}
	}
	// 1/5 gives it back on 20% of the crits.
	sim, pal, units := holySim(t, talentIllumination1, 0, 2)
	noCrit(sim, pal, 100)
	s := spell(t, pal, 25292)
	n := 0
	for range 2000 {
		pal.SpendMana(sim, pal.CurrentMana(), pal.GetManaNotCastingMetrics())
		s.ApplyEffects(sim, units[1], s)
		if pal.CurrentMana() > 0 {
			n++
		}
	}
	if r := float64(n) / 2000; math.Abs(r-0.2) > 0.03 {
		t.Errorf("Illumination 1 gave mana on %.3f of the crits, want 0.2", r)
	}
}

// TestInfusionOfLight: 2/2 takes 1 sec off the next Holy Light after a Flash of Light or Holy Shock
// crit, once.
func TestInfusionOfLight(t *testing.T) {
	sim, pal, units := holySim(t, talentInfusion, 0, 2)
	hl := spell(t, pal, 25292)
	if _, ct := castAt(t, sim, pal, hl, units[1]); ct != 2500*time.Millisecond {
		t.Errorf("Holy Light took %s with no Infusion", ct)
	}
	noCrit(sim, pal, 100)
	castAt(t, sim, pal, spell(t, pal, 19943), units[1])
	noCrit(sim, pal, 0)
	if _, ct := castAt(t, sim, pal, hl, units[1]); ct != 1500*time.Millisecond {
		t.Errorf("Holy Light after a Flash of Light crit took %s, want 1.5 s", ct)
	}
	if _, ct := castAt(t, sim, pal, hl, units[1]); ct != 2500*time.Millisecond {
		t.Errorf("the second Holy Light took %s, want 2.5 s", ct)
	}
	// A normal Flash of Light gives nothing; a Holy Shock crit does.
	castAt(t, sim, pal, spell(t, pal, 19943), units[1])
	if _, ct := castAt(t, sim, pal, hl, units[1]); ct != 2500*time.Millisecond {
		t.Errorf("Holy Light after a normal Flash of Light took %s", ct)
	}
	noCrit(sim, pal, 100)
	castAt(t, sim, pal, spell(t, pal, 20930), units[1])
	noCrit(sim, pal, 0)
	if _, ct := castAt(t, sim, pal, hl, units[1]); ct != 1500*time.Millisecond {
		t.Errorf("Holy Light after a Holy Shock crit took %s, want 1.5 s", ct)
	}
}

// TestOathOfLight: the Oath on a player adds 400 to Holy Light, 115 to Flash of Light and 230 to Holy
// Shock there (times Healing Light), nothing elsewhere, and one Oath at a time. It costs 8% of base mana.
func TestOathOfLight(t *testing.T) {
	sim, pal, units := holySim(t, talentOath, 0, 3)
	noCrit(sim, pal, 0)
	oath := spell(t, pal, 1328548)
	mana, _ := castAt(t, sim, pal, oath, units[1])
	if math.Abs(mana-0.08*pal.BaseMana) > 0.01 {
		t.Errorf("Oath of Light cost %.2f", mana)
	}
	if !pal.HasOathOfLight(units[1]) || pal.HasOathOfLight(units[2]) {
		t.Fatalf("Oath of Light on the wrong players")
	}
	mean := func(s *core.Spell, target *core.Unit) float64 {
		total := 0.0
		for range 2000 {
			h, _ := healOf(sim, s, target)
			total += h
		}
		return total / 2000
	}
	for _, c := range []struct {
		id    int32
		bonus float64
	}{{25292, 400}, {19943, 115}, {20930, 230}} {
		s := spell(t, pal, c.id)
		if d := mean(s, units[1]) - mean(s, units[2]); math.Abs(d-c.bonus) > 0.03*c.bonus {
			t.Errorf("%d heals %.1f more on the Oath's target, want %.0f", c.id, d, c.bonus)
		}
	}
	castAt(t, sim, pal, oath, units[2])
	if pal.HasOathOfLight(units[1]) || !pal.HasOathOfLight(units[2]) {
		t.Errorf("a second Oath left the first up")
	}
}

// TestBoundlessLight: 2/2 adds 30% crit to Holy Shock's heal at a player our Oath of Light isn't on, and
// none at the Oath's target.
func TestBoundlessLight(t *testing.T) {
	sim, pal, units := holySim(t, talentOathBoundless, 0, 3)
	noCrit(sim, pal, 0)
	castAt(t, sim, pal, spell(t, pal, 1328548), units[1])
	hs := spell(t, pal, 20930)
	rate := func(target *core.Unit) float64 {
		n := 0
		for range 4000 {
			if _, crit := healOf(sim, hs, target); crit {
				n++
			}
		}
		return float64(n) / 4000
	}
	if r := rate(units[2]); math.Abs(r-0.3) > 0.025 {
		t.Errorf("Holy Shock crit %.3f off the Oath's target, want 0.30", r)
	}
	if r := rate(units[1]); r != 0 {
		t.Errorf("Holy Shock crit %.3f on the Oath's target, want 0", r)
	}
}

// TestLightsVigil: the Vigil marks a player for 30 sec; our next Holy Shock there has no cooldown and
// heals the target and the four others of its party for rank 3's 684 to 724, none in another party. One
// Vigil in each party, 1340 mana, 1.5 sec, a 6 sec cooldown.
func TestLightsVigil(t *testing.T) {
	sim, pal, units := holySim(t, talentVigil, 0, 7)
	noCrit(sim, pal, 0)
	pal.AddStatDynamic(sim, stats.HealingPower, -pal.GetStat(stats.HealingPower)-pal.GetStat(stats.SpellPower))
	vigil := spell(t, pal, 1311595)
	mana, ct := castAt(t, sim, pal, vigil, units[2])
	if math.Abs(mana-1340) > 0.5 || ct != 1500*time.Millisecond || vigil.CD.TimeToReady(sim) != 6*time.Second {
		t.Errorf("Light's Vigil cost %.1f in %s, ready in %s", mana, ct, vigil.CD.TimeToReady(sim))
	}
	if !pal.HasLightsVigil(units[2]) {
		t.Fatalf("no Vigil on its target")
	}
	// A second one on the same party moves it.
	vigil.CD.Reset()
	castAt(t, sim, pal, vigil, units[3])
	if pal.HasLightsVigil(units[2]) || !pal.HasLightsVigil(units[3]) {
		t.Errorf("two Vigils in one party")
	}

	heal := spell(t, pal, 1311596)
	hs := spell(t, pal, 20930)
	before := map[int32]float64{}
	for _, u := range units {
		before[u.UnitIndex] = heal.SpellMetrics[u.UnitIndex].TotalHealing
	}
	castAt(t, sim, pal, hs, units[3])
	if !hs.CD.IsReady(sim) {
		t.Errorf("Holy Shock at the Vigil went on cooldown")
	}
	if pal.HasLightsVigil(units[3]) {
		t.Errorf("the Vigil is still up")
	}
	for i, u := range units {
		got := heal.SpellMetrics[u.UnitIndex].TotalHealing - before[u.UnitIndex]
		if i < 5 && (got < 684.21-0.01 || got > 723.79+0.01) {
			t.Errorf("player %d got %.1f from the Vigil, want 684 to 724", i, got)
		}
		if i >= 5 && got != 0 {
			t.Errorf("player %d in the other party got %.1f from the Vigil", i, got)
		}
	}
	// The next Holy Shock, with no Vigil, goes on its cooldown.
	castAt(t, sim, pal, hs, units[3])
	if hs.CD.IsReady(sim) {
		t.Errorf("Holy Shock with no Vigil had no cooldown")
	}
}

// TestSpiritualFocus: 5/5 keeps 70% of the hits from pushing back Holy Light, Flash of Light and Light's
// Vigil.
func TestSpiritualFocus(t *testing.T) {
	_, pal, _ := holySim(t, talentSpiritualFocus, 0, 1)
	for _, id := range []int32{25292, 19943, 1311595} {
		if got := spell(t, pal, id).PushbackReduction; math.Abs(got-0.7) > 1e-9 {
			t.Errorf("%d resists %.2f of the pushback, want 0.70", id, got)
		}
	}
	_, pal, _ = holySim(t, "", 0, 1)
	if got := spell(t, pal, 25292).PushbackReduction; got != 0 {
		t.Errorf("Holy Light resists %.2f of the pushback with no talent", got)
	}
}

// TestLayOnHands: it heals the target for the paladin's max health, takes all the paladin's mana and
// gives the target 550.
func TestLayOnHands(t *testing.T) {
	sim, pal, units := holySim(t, "", 0, 2)
	other := sim.Raid.Parties[0].Players[1].GetCharacter()
	other.SpendMana(sim, 1000, other.GetManaNotCastingMetrics())
	otherMana := other.CurrentMana()
	other.RemoveHealth(sim, other.CurrentHealth()-1)
	loh := spell(t, pal, 10310)
	heal, _ := healOf(sim, loh, units[1])
	if math.Abs(heal-pal.MaxHealth()) > 0.01 || pal.CurrentMana() != 0 || math.Abs(other.CurrentMana()-otherMana-550) > 0.01 {
		t.Errorf("Lay on Hands healed %.0f (max health %.0f), left %.0f mana, gave %.0f", heal, pal.MaxHealth(), pal.CurrentMana(), other.CurrentMana()-otherMana)
	}
	if loh.CD.Duration != 20*time.Minute {
		t.Errorf("Lay on Hands cooldown %s", loh.CD.Duration)
	}
}
