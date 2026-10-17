package enhancement

import (
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/shaman"
	"google.golang.org/protobuf/encoding/protojson"
)

// Restoration talent strings (the tree's order: Improved Healing Wave, Totemic Focus,
// Mindfulness, Natural Grace, Tidal Focus, Improved Reincarnation, Ancestral Healing, Healing
// Focus, Water Shield, Tidal Mastery, Restorative Totems, Mana Tide Totem, Healing Way,
// Nature's Swiftness, Purification, Riptide).
const (
	restoImprovedHealingWave = "--5"
	restoTidalFocus          = "--05"
	restoAncestralHealing    = "--0000003"
	restoHealingFocus        = "--00000003"
	restoTidalMastery        = "--0000000005"
	restoNaturesSwiftness    = "--00000000000001"
	restoHealingWayPurif     = "--000000000000305"
	restoRiptidePurif        = "--0000000000000051"
)

// healSim is a naked level orc shaman with talents and healingPower, who never crits unless
// critBonus says so, at 1 health so every heal counts in full.
func healSim(t *testing.T, level int32, talents string, healingPower, critBonus float64) (*core.Simulation, *EnhancementShaman) {
	sim, enh := newShamanSim(newOrcShaman(level, talents, &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 60)
	// Room for the biggest heal, so none is cut short by our max health.
	enh.AddStatDynamic(sim, stats.Health, 5000)
	enh.AddStatDynamic(sim, stats.HealingPower, healingPower)
	enh.AddStatDynamic(sim, stats.SpellCrit, critBonus-enh.GetStat(stats.SpellCrit))
	return sim, enh
}

// healOnce casts the heal at the given time at the mob we're fighting, and returns how much
// health we got from it once it lands.
func healOnce(t *testing.T, sim *core.Simulation, enh *EnhancementShaman, spell *core.Spell, at_ float64, castTime float64) *float64 {
	var before float64
	gained := new(float64)
	at(sim, at_, func(sim *core.Simulation) {
		enh.RemoveHealth(sim, enh.CurrentHealth()-1)
		enh.AddMana(sim, enh.MaxMana(), enh.GetManaNotCastingMetrics())
		before = enh.CurrentHealth()
		castNow(t, sim, enh, spell)
	})
	// An empty rotation finishes a cast on its next 50 ms step.
	at(sim, at_+castTime+0.1, func(sim *core.Simulation) {
		*gained = enh.CurrentHealth() - before
	})
	return gained
}

// TestOrcShamanHealRanks checks the heal tables, read from the Forever client (healing.go):
// each rank's heal at a level and which ranks a level has.
func TestOrcShamanHealRanks(t *testing.T) {
	for _, c := range []struct {
		name      string
		rank      func() (float64, float64)
		low, high float64
	}{
		// 1615 at 60, 13.26% variance.
		{"Healing Wave 10 at 60", func() (float64, float64) { return shaman.HealingWaveRanks[10].At(60) }, 1507.95, 1722.05},
		// 39 at 1 plus 0.7 a level to 5.
		{"Healing Wave 1 at 30", func() (float64, float64) { return shaman.HealingWaveRanks[1].At(30) }, 36.44, 47.16},
		// 224 at 28 plus 2.1 a level.
		{"Lesser Healing Wave 2 at 30", func() (float64, float64) { return shaman.LesserHealingWaveRanks[2].At(30) }, 213.51, 242.89},
		// 480 at 40 plus 3.7 a level to 48: the talent tooltip's 486 to 534.
		{"Riptide at 60", func() (float64, float64) { return shaman.RiptideRank.At(60) }, 484.12, 535.08},
	} {
		low, high := c.rank()
		if math.Abs(low-c.low) > 0.01 || math.Abs(high-c.high) > 0.01 {
			t.Errorf("%s heals %.2f to %.2f, want %.2f to %.2f", c.name, low, high, c.low, c.high)
		}
	}

	for _, c := range []struct {
		level         int32
		hw, lhw, heal int
	}{{30, 5, 2, 0}, {60, 10, 6, 3}} {
		_, enh := healSim(t, c.level, "", 0, 0)
		count := func(spells []*core.Spell) int {
			n := 0
			for _, s := range spells {
				if s != nil {
					n++
				}
			}
			return n
		}
		if hw, lhw, heal := count(enh.HealingWave), count(enh.LesserHealingWave), count(enh.ChainHeal); hw != c.hw || lhw != c.lhw || heal != c.heal {
			t.Errorf("level %d has %d / %d / %d ranks of Healing Wave / Lesser Healing Wave / Chain Heal, want %d / %d / %d",
				c.level, hw, lhw, heal, c.hw, c.lhw, c.heal)
		}
	}
}

// TestOrcShamanHealAmounts casts each heal at 1 health with 200 healing power and no crits,
// and checks what we get. The coefficients follow the cast time, and Forever's downranking
// takes a rank learned more than 17 levels below us down 5% a level. Purification adds 10%
// to every heal and Healing Way 25% to Healing Wave, and the two multiply.
func TestOrcShamanHealAmounts(t *testing.T) {
	type heal struct {
		name     string
		spell    func(enh *EnhancementShaman) *core.Spell
		castTime float64
		low      float64
		high     float64
	}
	hw10 := func(enh *EnhancementShaman) *core.Spell { return enh.HealingWave[10] }
	hw1 := func(enh *EnhancementShaman) *core.Spell { return enh.HealingWave[1] }
	lhw6 := func(enh *EnhancementShaman) *core.Spell { return enh.LesserHealingWave[6] }
	for _, c := range []struct {
		name    string
		talents string
		heals   []heal
	}{
		{"no talents", "", []heal{
			{"Healing Wave 10", hw10, 3, 1507.95 + 0.857*200, 1722.05 + 0.857*200},
			// Learned at 1: 0.429 * (1 - 0.05 * 42) is nothing at 60.
			{"Healing Wave 1", hw1, 1.5, 36.44 + 0.7*0, 47.16 + 0.7*0},
			{"Lesser Healing Wave 6", lhw6, 1.5, 820*(1-0.109090909/2) + 0.429*200, 820*(1+0.109090909/2) + 0.429*200},
		}},
		{"Healing Way 3 and Purification 5", restoHealingWayPurif, []heal{
			{"Healing Wave 10", hw10, 3, (1507.95 + 0.857*200) * 1.1 * 1.25, (1722.05 + 0.857*200) * 1.1 * 1.25},
			{"Lesser Healing Wave 6", lhw6, 1.5, (820*(1-0.109090909/2) + 0.429*200) * 1.1, (820*(1+0.109090909/2) + 0.429*200) * 1.1},
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			// Healing Wave 1 at 60 grows to its level 5 value: 39 + 0.7 * 4.
			if c.name == "no talents" {
				c.heals[1].low, c.heals[1].high = 41.8*(1-0.25641027/2), 41.8*(1+0.25641027/2)
			}
			for _, h := range c.heals {
				for range 20 {
					sim, enh := healSim(t, 60, c.talents, 200, 0)
					got := healOnce(t, sim, enh, h.spell(enh), 1, h.castTime)
					runSim(sim)
					if *got < h.low-0.5 || *got > h.high+0.5 {
						t.Fatalf("%s healed %.1f, want %.1f to %.1f", h.name, *got, h.low, h.high)
					}
				}
			}
		})
	}
}

// TestOrcShamanHealTalents checks the talents that change a heal's cast, cost, crit and
// pushback.
func TestOrcShamanHealTalents(t *testing.T) {
	castTime := func(talents string, spell func(*EnhancementShaman) *core.Spell) time.Duration {
		_, enh := healSim(t, 60, talents, 0, 0)
		return spell(enh).DefaultCast.CastTime
	}
	hw := func(enh *EnhancementShaman) *core.Spell { return enh.HealingWave[10] }
	lhw := func(enh *EnhancementShaman) *core.Spell { return enh.LesserHealingWave[6] }
	if got := castTime("", hw); got != 3*time.Second {
		t.Errorf("Healing Wave casts in %s, want 3s", got)
	}
	if got := castTime(restoImprovedHealingWave, hw); got != 2500*time.Millisecond {
		t.Errorf("Improved Healing Wave 5: Healing Wave casts in %s, want 2.5s", got)
	}
	if got := castTime(restoImprovedHealingWave, lhw); got != 1500*time.Millisecond {
		t.Errorf("Improved Healing Wave 5: Lesser Healing Wave casts in %s, want 1.5s", got)
	}

	// Tidal Focus: 5% off every heal's cost.
	for _, c := range []struct {
		talents string
		hw, lhw float64
	}{{"", 620, 380}, {restoTidalFocus, 589, 361}} {
		_, enh := healSim(t, 60, c.talents, 0, 0)
		for _, s := range []struct {
			spell *core.Spell
			want  float64
		}{{enh.HealingWave[10], c.hw}, {enh.LesserHealingWave[6], c.lhw}} {
			if got := s.spell.Cost.GetCurrentCost(); math.Abs(got-s.want) > 0.01 {
				t.Errorf("%q: %s costs %.1f, want %.0f", c.talents, s.spell.ActionID, got, s.want)
			}
		}
	}

	// Tidal Mastery: 5% more crit on heals, none on Lightning Bolt.
	_, plain := healSim(t, 60, "", 0, 0)
	_, mastery := healSim(t, 60, restoTidalMastery, 0, 0)
	if got := mastery.HealingWave[10].HealingCritChance() - plain.HealingWave[10].HealingCritChance(); math.Abs(got-0.05) > 1e-9 {
		t.Errorf("Tidal Mastery 5 adds %.3f crit to Healing Wave, want 0.05", got)
	}
	if mastery.LightningBolt[10].BonusCritRating != plain.LightningBolt[10].BonusCritRating {
		t.Error("Tidal Mastery changes Lightning Bolt's crit")
	}

	// Healing Focus: 70% to ignore a hit's pushback on each heal with a cast time.
	_, focus := healSim(t, 60, restoHealingFocus, 0, 0)
	for _, spell := range []*core.Spell{focus.HealingWave[10], focus.LesserHealingWave[6], focus.ChainHeal[3]} {
		if spell.PushbackReduction != 0.70 {
			t.Errorf("Healing Focus 3: %s ignores pushback %.0f%% of the time, want 70%%", spell.ActionID, 100*spell.PushbackReduction)
		}
	}
}

// TestOrcShamanNaturesSwiftnessHeal: Nature's Swiftness makes the next Healing Wave instant.
func TestOrcShamanNaturesSwiftnessHeal(t *testing.T) {
	sim, enh := healSim(t, 60, restoNaturesSwiftness, 0, 0)
	ns := enh.GetAura("Natures Swiftness")
	if ns == nil {
		t.Fatal("no Nature's Swiftness aura")
	}
	at(sim, 1, func(sim *core.Simulation) {
		ns.Activate(sim)
		enh.RemoveHealth(sim, enh.CurrentHealth()-1)
		castNow(t, sim, enh, enh.HealingWave[10])
		if enh.IsCasting(sim) {
			t.Errorf("Healing Wave is casting until %s with Nature's Swiftness up", enh.Hardcast.Expires)
		}
		if enh.CurrentHealth() < 1000 {
			t.Errorf("the instant Healing Wave left us at %.0f health", enh.CurrentHealth())
		}
		if ns.IsActive() {
			t.Error("Nature's Swiftness is still up after the heal")
		}
	})
	runSim(sim)
}

// TestOrcShamanAncestralHealing: a heal crit gives us 25% more armor for 15 sec with 3
// points.
func TestOrcShamanAncestralHealing(t *testing.T) {
	sim, enh := healSim(t, 60, restoAncestralHealing, 0, 200)
	buff := enh.GetAura("Ancestral Fortitude")
	if buff == nil {
		t.Fatal("no Ancestral Fortitude aura")
	}
	var armor float64
	at(sim, 0.5, func(sim *core.Simulation) { armor = enh.Armor() })
	healOnce(t, sim, enh, enh.LesserHealingWave[6], 1, 1.5)
	at(sim, 3, func(sim *core.Simulation) {
		if !buff.IsActive() {
			t.Fatal("no Ancestral Fortitude after a crit heal")
		}
		if got := enh.Armor(); math.Abs(got-armor*1.25) > 0.01 {
			t.Errorf("armor %.1f with the buff, want %.1f", got, armor*1.25)
		}
	})
	at(sim, 18, func(sim *core.Simulation) {
		if buff.IsActive() || math.Abs(enh.Armor()-armor) > 0.01 {
			t.Errorf("15 sec on, the buff is up %v and armor %.1f, want %.1f", buff.IsActive(), enh.Armor(), armor)
		}
	})
	runSim(sim)
}

// TestOrcShamanRiptide: an instant heal, then 89 + 10% of healing power every 3 sec for 15
// sec, both with Purification's 10%. Chain Heal at us while it's up heals 25% more.
func TestOrcShamanRiptide(t *testing.T) {
	sim, enh := healSim(t, 60, restoRiptidePurif, 100, 0)
	if enh.Riptide == nil {
		t.Fatal("no Riptide with the talent at 60")
	}
	direct := healOnce(t, sim, enh, enh.Riptide, 1, 0)
	var afterDirect float64
	at(sim, 1.15, func(sim *core.Simulation) { afterDirect = enh.CurrentHealth() })
	at(sim, 16.5, func(sim *core.Simulation) {
		low, high := shaman.RiptideRank.At(60)
		if *direct < (low+0.214*100)*1.1-0.5 || *direct > (high+0.214*100)*1.1+0.5 {
			t.Errorf("Riptide healed %.1f at once, want %.1f to %.1f", *direct, (low+0.214*100)*1.1, (high+0.214*100)*1.1)
		}
		ticks := enh.CurrentHealth() - afterDirect
		if want := 5 * (89 + 0.1*100) * 1.1; math.Abs(ticks-want) > 1 {
			t.Errorf("Riptide's ticks healed %.1f, want %.1f", ticks, want)
		}
	})
	runSim(sim)
	if enh.ChainHeal[3] == nil {
		t.Fatal("no Chain Heal at 60")
	}

	// Chain Heal with Riptide on us and without.
	chain := func(withRiptide bool) float64 {
		sim, enh := healSim(t, 60, restoRiptidePurif, 0, 0)
		if withRiptide {
			at(sim, 0.5, func(sim *core.Simulation) {
				enh.AddMana(sim, enh.MaxMana(), enh.GetManaNotCastingMetrics())
				castNow(t, sim, enh, enh.Riptide)
			})
		}
		// Riptide ticks at 3.5 and 6.5, so none lands while this one is measured.
		got := healOnce(t, sim, enh, enh.ChainHeal[3], 3.6, 2.5)
		runSim(sim)
		return *got
	}
	low, high := shaman.ChainHealRanks[3].At(60)
	if got := chain(false); got < low*1.1-0.5 || got > high*1.1+0.5 {
		t.Errorf("Chain Heal healed %.1f, want %.1f to %.1f", got, low*1.1, high*1.1)
	}
	if got := chain(true); got < low*1.1*1.25-0.5 || got > high*1.1*1.25+0.5 {
		t.Errorf("Chain Heal with Riptide up healed %.1f, want %.1f to %.1f", got, low*1.1*1.25, high*1.1*1.25)
	}
}

// TestHealAtDyingMob: a heal cast at the mob we fight (a rotation's default target) heals us,
// and the mob dying during the cast doesn't stop it.
func TestHealAtDyingMob(t *testing.T) {
	sim, enh := packSim(t, 30, 2, nil)
	mobs := sim.Encounter.TargetUnits
	heal := topRank(t, "Lesser Healing Wave", enh.LesserHealingWave)
	at(sim, 1, func(sim *core.Simulation) {
		enh.RemoveHealth(sim, enh.CurrentHealth()/2)
		castNow(t, sim, enh, heal)
		if enh.Hardcast.Target != mobs[0] {
			t.Fatalf("the heal is cast at %v, want the mob", enh.Hardcast.Target)
		}
	})
	at(sim, 1.5, func(sim *core.Simulation) { killNow(sim, enh, mobs[0]) })
	// The other mob's swings can push the heal back, by up to its 1.5 sec cast.
	at(sim, 4.2, func(sim *core.Simulation) {
		if got := heal.SpellMetrics[enh.UnitIndex].TotalHealing; got < 200 {
			t.Errorf("the heal healed us for %.0f", got)
		}
	})
	for !sim.Step() && sim.CurrentTime < 4500*time.Millisecond {
	}
}

// TestHealingKeepsUsAlive: a naked level 30 with white swings dies to one level 30 mob some
// of the time (TestSoloNormalMob). A rotation that casts Lesser Healing Wave below 40% health
// keeps it alive.
func TestHealingKeepsUsAlive(t *testing.T) {
	deaths := func(rotation string) float64 {
		req := normalMobRequest([]*proto.Target{core.NormalMobTarget(30, proto.MobType_MobTypeBeast, 1)}, true, 300)
		if rotation != "" {
			apl := &proto.APLRotation{}
			if err := protojson.Unmarshal([]byte(rotation), apl); err != nil {
				t.Fatal(err)
			}
			req.Raid.Parties[0].Players[0].Rotation = apl
		}
		result := core.RunRaidSim(req)
		if result.Error != nil {
			t.Fatal(result.Error.Message)
		}
		return result.RaidMetrics.Parties[0].Players[0].ChanceOfDeath
	}
	without := deaths("")
	with := deaths(`{"type":"TypeAPL","priorityList":[{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"40%"}}}},"castSpell":{"spellId":{"spellId":8008}}}}]}`)
	t.Logf("we died %.1f%% of the time without heals and %.1f%% with them", 100*without, 100*with)
	if without < 0.1 || with > 0.01 {
		t.Errorf("we died %.1f%% of the time without heals and %.1f%% with them, want some and none", 100*without, 100*with)
	}
}

// TestHealRankFamily: a rotation that names a heal rank we haven't learned, like one written
// at 60, casts the highest rank of that heal our level knows. A level 30 shaman casts Healing
// Wave rank 5 for rank 10 and Lesser Healing Wave rank 2 for rank 6.
func TestHealRankFamily(t *testing.T) {
	for _, c := range []struct {
		named, cast int32
	}{
		{25357, 939},
		{10468, 8008},
	} {
		req := normalMobRequest([]*proto.Target{core.NormalMobTarget(30, proto.MobType_MobTypeBeast, 1)}, true, 1)
		req = withRotation(t, req, rotationOf(fmt.Sprintf(
			`{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"60%%"}}}},"castSpell":{"spellId":{"spellId":%d}}}}`, c.named)))
		req.SimOptions.Debug = true
		result := core.RunRaidSim(req)
		if result.Error != nil {
			t.Fatal(result.Error.Message)
		}
		casts := 0
		logLines(t, result.Logs, func(_ float64, line string) {
			if m := completedLine.FindStringSubmatch(line); m != nil && m[1] == strconv.Itoa(int(c.cast)) {
				casts++
			}
		})
		if casts == 0 {
			t.Errorf("a rotation naming %d cast no %d", c.named, c.cast)
		}
	}
}
