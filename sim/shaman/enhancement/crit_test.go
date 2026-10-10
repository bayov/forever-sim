package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanMeleeCrit checks how a mob's level changes our melee crit chance, and that a
// melee crit deals twice a normal hit.
//
// We lose 0.2% crit for each point the mob's defense (its level times 5) is above our level
// times 5, and a flat 1.8% more against a mob 3 or more levels up. Skill from gear doesn't
// help. Against a mob below our level we gain 0.04% a point (Blizzard's 2019 Classic posts
// and magey's tests). The crit cap of white hits is in TestOrcShamanAttackTable.
func TestOrcShamanMeleeCrit(t *testing.T) {
	cases := []struct {
		name               string
		level, targetLevel int32
		weapon             int32
		skill, lost        float64
	}{
		{"level 60 against a level 63 boss", 60, 63, darkEdge, 0, 0.048},
		{"310 skill against the boss", 60, 63, darkEdge, 10, 0.048},
		{"level 30 against a level 34", 30, 34, whirlwind, 0, 0.058},
		{"level 30 against a level 33", 30, 33, whirlwind, 0, 0.048},
		{"level 30 against Vishas (level 32)", 30, 32, whirlwind, 0, 0.02},
		{"a level 31", 30, 31, whirlwind, 0, 0.01},
		{"the same level", 30, 30, whirlwind, 0, 0},
		{"5 levels below us", 30, 25, whirlwind, 0, -0.01},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim, enh := newTargetLevelSim(c.level, c.targetLevel, c.weapon, c.skill, stormstrike)
			white := enh.AutoAttacks.MHAuto()
			attackTable := enh.AttackTables[enh.CurrentTarget.UnitIndex][white.CastType]
			ours := enh.GetStat(stats.MeleeCrit) / (core.CritRatingPerCritChance * 100)
			for _, spell := range []*core.Spell{white, enh.Stormstrike} {
				if got := ours - spell.PhysicalCritChance(attackTable); math.Abs(got-c.lost) > 1e-9 {
					t.Errorf("%s loses %.4f crit, want %.4f", spell.ActionID, got, c.lost)
				}
			}

			// Every crit, white or yellow, deals twice a normal hit of the same attack. We add
			// crit so that there are crits to see against the higher mobs.
			enh.AddStatDynamic(sim, stats.MeleeCrit, 20*core.CritRatingPerCritChance)
			for _, attack := range []struct {
				spell   *core.Spell
				outcome core.OutcomeApplier
			}{
				{white, white.OutcomeMeleeWhite},
				{enh.Stormstrike, enh.Stormstrike.OutcomeMeleeWeaponSpecialHitAndCrit},
			} {
				var hit, crit float64
				for range 2000 {
					result := attack.spell.CalcDamage(sim, enh.CurrentTarget, 100, attack.outcome)
					switch {
					case result.Outcome.Matches(core.OutcomeCrit):
						crit = result.Damage
					case result.Outcome.Matches(core.OutcomeHit):
						hit = result.Damage
					}
				}
				if hit == 0 || math.Abs(crit/hit-2) > 1e-9 {
					t.Errorf("%s crit for %.2f and hit for %.2f, want twice", attack.spell.ActionID, crit, hit)
				}
			}
		})
	}
}

// TestOrcShamanSpellCrit checks that a mob's level doesn't change our spell crit chance, and
// that a spell crit deals 1.5 times a normal hit.
//
// Melee crit is cut against higher mobs (TestOrcShamanMeleeCrit), but spell crit isn't. On
// the beta our spells crit on 13% of hits against level 33 mobs with 12.3% on the sheet, but
// on only 4% against level 34 mobs. So a cut at 4 levels up is still open (notes.md). We
// check Lightning Bolt, Earth Shock (our one binary spell), Flame Shock's direct part and a
// Searing Totem attack.
func TestOrcShamanSpellCrit(t *testing.T) {
	cases := []struct {
		name               string
		level, targetLevel int32
		weapon             int32
	}{
		{"level 60 against a level 63 boss", 60, 63, darkEdge},
		{"level 30 against a level 34", 30, 34, whirlwind},
		{"level 30 against a level 33", 30, 33, whirlwind},
		{"level 30 against Vishas (level 32)", 30, 32, whirlwind},
		{"the same level", 30, 30, whirlwind},
		{"5 levels below us", 30, 25, whirlwind},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim, enh := newTargetLevelSim(c.level, c.targetLevel, c.weapon, 0, "")
			spells := []*core.Spell{
				topRank(t, "Lightning Bolt", enh.LightningBolt),
				topRank(t, "Earth Shock", enh.EarthShock),
				topRank(t, "Flame Shock", enh.FlameShock),
				lastSpell(enh, shaman.SearingTotemAttackSpellId[:], 0),
			}
			ours := enh.GetStat(stats.SpellCrit) / (core.SpellCritRatingPerCritChance * 100)
			for _, spell := range spells {
				if got := spell.SpellCritChance(enh.CurrentTarget); math.Abs(got-ours) > 1e-9 {
					t.Errorf("%s crits %.4f, want our %.4f", spell.ActionID, got, ours)
				}
			}
			if c.targetLevel != c.level {
				return
			}

			// We add crit so that there are crits to see. A mob of our level never partly resists,
			// so every hit of the same spell deals the same.
			enh.AddStatDynamic(sim, stats.SpellCrit, 20*core.SpellCritRatingPerCritChance)
			for _, spell := range spells {
				var hit, crit float64
				for range 2000 {
					result := spell.CalcDamage(sim, enh.CurrentTarget, 100, spell.OutcomeMagicHitAndCrit)
					switch {
					case result.Outcome.Matches(core.OutcomeCrit):
						crit = result.Damage
					case result.Outcome.Matches(core.OutcomeHit):
						hit = result.Damage
					}
				}
				if hit == 0 || math.Abs(crit/hit-1.5) > 1e-9 {
					t.Errorf("%s crit for %.2f and hit for %.2f, want 1.5 times", spell.ActionID, crit, hit)
				}
			}
		})
	}
}
