package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	googleProto "google.golang.org/protobuf/proto"
)

// TestOrcShamanMissChance checks the chance our white hits and Stormstrikes miss, by the
// target's level, our weapon skill and our hit.
//
// Miss is 5% plus 0.1% for each point the target's defense (its level times 5) is above
// our weapon skill, or 0.2% a point when that gap is more than 10. Our hit comes off it.
// When the gap is more than 10, the first (gap - 10) * 0.2% of our hit does nothing, so a
// level 60 with 300 skill needs 9% hit against a level 63 boss (magey's 2019 Classic attack
// table tests, and the 1% Blizzard described).
//
// Forever's Orc Axe Specialization is crit, not skill, so our axes add no skill here.
// Huge Thorium Battleaxe gives 2 two-handed axe skill under Forever (10 in Classic), and
// Tidal Focus 1% hit a point.
func TestOrcShamanMissChance(t *testing.T) {
	cases := []struct {
		name               string
		level, targetLevel int32
		weapon             int32
		skill              float64
		talents            string
		hit                float64
		want               float64
	}{
		{"level 60 against a level 63 boss", 60, 63, darkEdge, 0, stormstrike, 0, 0.08},
		{"the first 1% of hit does nothing against the boss", 60, 63, darkEdge, 0, stormstrike, 1, 0.08},
		{"5% hit against the boss", 60, 63, darkEdge, 0, stormstrike, 5, 0.04},
		{"the 9% hit cap", 60, 63, darkEdge, 0, stormstrike, 9, 0},
		{"hit over the cap", 60, 63, darkEdge, 0, stormstrike, 12, 0},
		{"Tidal Focus 5/5 against the boss", 60, 63, darkEdge, 0, stormstrike + "-00005", 0, 0.04},
		{"level 30 against Vishas (level 32)", 30, 32, whirlwind, 0, stormstrike, 0, 0.06},
		{"1% hit against Vishas", 30, 32, whirlwind, 0, stormstrike, 1, 0.05},
		{"level 30 against a level 33", 30, 33, whirlwind, 0, stormstrike, 0, 0.08},
		{"152 skill against a level 33", 30, 33, whirlwind, 2, stormstrike, 0, 0.076},
		{"152 skill and 1% hit against a level 33", 30, 33, whirlwind, 2, stormstrike, 1, 0.072},
		{"Huge Thorium Battleaxe's +2 skill against the boss", 60, 63, thorium, 0, stormstrike, 0, 0.076},
		{"the same level", 30, 30, whirlwind, 0, stormstrike, 0, 0.05},
		{"5 levels below us", 30, 25, whirlwind, 0, stormstrike, 0, 0.025},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim, enh := newTargetLevelSim(c.level, c.targetLevel, c.weapon, c.skill, c.talents)
			enh.AddStatDynamic(sim, stats.MeleeHit, c.hit*core.MeleeHitRatingPerHitChance)

			if enh.Talents.TidalFocus > 0 {
				if got := enh.GetStat(stats.MeleeHit); got != 5*core.MeleeHitRatingPerHitChance {
					t.Fatalf("Tidal Focus %d/5 gives %.2f melee hit, want 5", enh.Talents.TidalFocus, got)
				}
			}
			if enh.Stormstrike == nil {
				t.Fatalf("no Stormstrike")
			}
			white := enh.AutoAttacks.MHAuto()
			attackTable := enh.AttackTables[enh.CurrentTarget.UnitIndex][white.CastType]
			// The same sum the white and yellow attack tables roll against.
			for _, spell := range []*core.Spell{white, enh.Stormstrike} {
				got := max(0, attackTable.BaseMissChance-spell.PhysicalHitChance(attackTable))
				if math.Abs(got-c.want) > 1e-9 {
					t.Errorf("%s misses %.4f, want %.4f", spell.ActionID, got, c.want)
				}
			}
		})
	}
}

// The two-handed axes and the talents the attack table tests use.
const (
	darkEdge    = 21134 // Dark Edge of Insanity, a level 60 two-handed axe.
	whirlwind   = 6975  // Whirlwind Axe, a level 30 two-handed axe.
	treeChop    = 2907  // Dwarven Tree Chopper, a two-handed axe with 0.6% expertise and no skill.
	thorium     = 12775 // Huge Thorium Battleaxe, a two-handed axe with +2 skill under Forever.
	stormstrike = "-0000000000001"
)

// newTargetLevelSim starts a Forever sim of an Orc shaman at level, holding weapon, against
// one target of targetLevel. skill is extra two-handed axe skill on top of the weapon's.
//
// Most tests give the skill this way rather than through an item, so that they don't
// change when Forever changes an item's skill.
func newTargetLevelSim(level, targetLevel, weapon int32, skill float64, talents string) (*core.Simulation, *EnhancementShaman) {
	player := newOrcShaman(level, talents, &proto.EnhancementShaman_Options{})
	pseudoStats := make([]float64, len(proto.PseudoStat_name))
	pseudoStats[proto.PseudoStat_PseudoStatTwoHandedAxesSkill] = skill
	player.BonusStats = &proto.UnitStats{PseudoStats: pseudoStats}
	for range proto.ItemSlot_ItemSlotMainHand {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
	}
	player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: weapon})
	target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
	target.Level = targetLevel
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{target}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
}

// TestOrcShamanSpellMissChance checks the chance our spells miss, by the target's level and
// our spell hit.
//
// It's 4% against our level, 1% less for every level below us and never under 1%. Above us
// it's 5% and 6% at 1 and 2 levels, 17% at 3 levels and 11% more for every level past that.
// Our spell hit comes off it, down to the 1% floor. Under Forever the hit on our gear counts
// for spells too, so Tarnished Elven Ring's 1% melee hit helps Lightning Bolt. Tidal Focus
// gives 1% spell hit a point.
func TestOrcShamanSpellMissChance(t *testing.T) {
	const tarnishedElvenRing = 18500
	const tidalFocus = "--00005"
	cases := []struct {
		name               string
		level, targetLevel int32
		talents            string
		ring               int32
		hit                float64
		want               float64
	}{
		{"5 levels below us", 30, 25, "", 0, 0, 0.01},
		{"3 levels below us", 30, 27, "", 0, 0, 0.01},
		{"2 levels below us", 30, 28, "", 0, 0, 0.02},
		{"1 level below us", 30, 29, "", 0, 0, 0.03},
		{"the same level", 30, 30, "", 0, 0, 0.04},
		{"1 level above us", 30, 31, "", 0, 0, 0.05},
		{"Vishas (2 levels above us)", 30, 32, "", 0, 0, 0.06},
		{"3 levels above us", 30, 33, "", 0, 0, 0.17},
		{"4 levels above us", 30, 34, "", 0, 0, 0.28},
		{"5 levels above us", 30, 35, "", 0, 0, 0.39},
		{"level 60 against a level 63 boss", 60, 63, "", 0, 0, 0.17},
		{"Tidal Focus 5/5 against the boss", 60, 63, tidalFocus, 0, 0, 0.12},
		{"a ring's 1% melee hit against the boss", 60, 63, "", tarnishedElvenRing, 0, 0.16},
		{"Tidal Focus 5/5 at the same level", 30, 30, tidalFocus, 0, 0, 0.01},
		{"20% hit against the boss", 60, 63, "", 0, 20, 0.01},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			player := newOrcShaman(c.level, c.talents, &proto.EnhancementShaman_Options{})
			for range proto.ItemSlot_ItemSlotFinger1 {
				player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
			}
			player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: c.ring})
			target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
			target.Level = c.targetLevel
			sim := core.NewSim(&proto.RaidSimRequest{
				Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
				Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{target}},
				SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
			}, simsignals.CreateSignals())
			sim.Reset()
			enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
			enh.AddStatDynamic(sim, stats.SpellHit, c.hit*core.SpellHitRatingPerHitChance)

			bolt := topRank(t, "Lightning Bolt", enh.LightningBolt)
			attackTable := enh.AttackTables[enh.CurrentTarget.UnitIndex][bolt.CastType]
			if got := bolt.SpellChanceToMiss(attackTable); math.Abs(got-c.want) > 1e-9 {
				t.Errorf("Lightning Bolt misses %.4f, want %.4f", got, c.want)
			}
		})
	}
}

// TestOrcShamanSpellMissRolls checks that Lightning Bolt really misses a level 63 boss 17% of
// the time at level 60, over 40000 rolls (within 4 standard errors, 0.75%).
func TestOrcShamanSpellMissRolls(t *testing.T) {
	sim, enh := newShamanSim(newOrcShaman(60, "", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 10)
	if level := enh.CurrentTarget.Level; level != 63 {
		t.Fatalf("the target is level %d, want 63", level)
	}
	bolt := topRank(t, "Lightning Bolt", enh.LightningBolt)
	const rolls = 40000
	misses := 0
	at(sim, 1, func(sim *core.Simulation) {
		for range rolls {
			if result := bolt.CalcDamage(sim, enh.CurrentTarget, 1, bolt.OutcomeMagicHitAndCrit); !result.Landed() {
				misses++
			}
		}
	})
	runSim(sim)

	got := float64(misses) / rolls
	if limit := 4 * math.Sqrt(0.17*0.83/rolls); math.Abs(got-0.17) > limit {
		t.Errorf("Lightning Bolt missed %.4f of %d rolls, want 0.17 within %.4f", got, rolls, limit)
	}
}
