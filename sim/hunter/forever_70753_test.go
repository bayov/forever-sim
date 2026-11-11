package hunter

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestForeverHotfix20261110Costs checks Arcane Shot and Summon Hawk at 11% of base mana (server
// hotfix 113892 of 2026-11-10, SpellPower rows for every rank of both): 189.2 at 60 (1720 base mana).
func TestForeverHotfix20261110Costs(t *testing.T) {
	_, hunter := newHunterSim(t, BeastMasteryHawkTalents, &proto.Debuffs{})
	for _, spell := range []*core.Spell{hunter.ArcaneShot, hunter.SummonHawk} {
		if spell == nil || spell.Cost == nil {
			t.Fatalf("missing spell or cost")
		}
		if got, want := spell.Cost.BaseCost, 0.11*hunter.BaseMana; math.Abs(got-want) > 1e-9 || math.Abs(want-189.2) > 1e-9 {
			t.Errorf("%v base cost %v, want %v (11%% of %v base mana)", spell.ActionID, got, want, hunter.BaseMana)
		}
	}
}

// TestForeverHotfix20261103Damage checks the shots' damage after the launch realms' server hotfix
// of 2026-11-03 (push 113534 on client 1.60.1.70707, wago.tools' hotfix log; foreverchanges.pro;
// Blizzard's 11-02 notes): Aimed Shot about +110%, Arcane Shot about +30%, Serpent Sting about
// +20% and Sniper Shot +100%. Serpent Sting's hotfix rows give a tick (rank 8: 100, 500 over the
// 15 sec).
func TestForeverHotfix20261103Damage(t *testing.T) {
	for _, c := range []struct {
		name string
		got  []float64
		want []float64
	}{
		{"Aimed Shot", AimedShotBonus[1:], []float64{40, 70, 115, 185, 260, 350}},
		{"Arcane Shot", ArcaneShotBaseDamage[1:], []float64{25, 35, 50, 85, 120, 175, 220, 280}},
		{"Serpent Sting", SerpentStingDamage[1:], []float64{15, 35, 70, 130, 200, 290, 375, 500, 675}},
		{"Sniper Shot", SniperShotBonus[1:], []float64{320, 450, 590}},
	} {
		for i := range c.want {
			if c.got[i] != c.want[i] {
				t.Errorf("%s rank %d: %v, want %v", c.name, i+1, c.got[i], c.want[i])
			}
		}
	}
}

// TestForeverHotfix20261103Talents checks Barrage (5/10/15%, curve 82803) and Lightning Reflexes
// (3% Agility a rank, curve 109548) after the same hotfix.
func TestForeverHotfix20261103Talents(t *testing.T) {
	for rank, want := range []float64{1, 1.05, 1.10, 1.15} {
		hunter := &Hunter{}
		hunter.Talents = &proto.HunterTalents{Barrage: int32(rank)}
		if got := hunter.barrageMultiplier(); math.Abs(got-want) > 1e-9 {
			t.Errorf("Barrage %d: x%v, want x%v", rank, got, want)
		}
	}

	// Survival with and without Lightning Reflexes 5 (Lacerating Strikes left out).
	_, without := newHunterSim(t, "--500230230250122100", &proto.Debuffs{})
	_, with := newHunterSim(t, "--500230230250122150", &proto.Debuffs{})
	if without.Talents.LightningReflexes != 0 || with.Talents.LightningReflexes != 5 {
		t.Fatalf("Lightning Reflexes %d and %d, want 0 and 5", without.Talents.LightningReflexes, with.Talents.LightningReflexes)
	}
	ratio := with.GetStat(stats.Agility) / without.GetStat(stats.Agility)
	if math.Abs(ratio-1.15) > 0.002 {
		t.Errorf("Lightning Reflexes 5: Agility x%.4f, want x1.15", ratio)
	}
}

// TestForeverHotfix20261103Aspects checks Aspect of the Hawk rank 6 (110 ranged attack power) after
// the hotfix. Trueshot Aura's ranks are in sim/core's TestForeverBestLearnedRank.
func TestForeverHotfix20261103Aspects(t *testing.T) {
	sim, hunter := newHunterSim(t, "--500230230250122151", &proto.Debuffs{})
	before := hunter.GetStat(stats.RangedAttackPower)
	hunter.AspectOfTheHawkAura.Activate(sim)
	if got := hunter.GetStat(stats.RangedAttackPower) - before; got != 110 {
		t.Errorf("Aspect of the Hawk gives %v ranged attack power, want 110", got)
	}
}

// TestForeverTrapsShareCooldown checks that every trap shares one 30 sec cooldown, as client
// 1.60.1.70753 puts Freezing and Frost Trap into the Fire traps' category 411 ("Hunter Traps").
func TestForeverTrapsShareCooldown(t *testing.T) {
	_, hunter := newHunterSim(t, "--500230230250122151", &proto.Debuffs{})
	if hunter.ImmolationTrap == nil || hunter.ExplosiveTrap == nil || hunter.FreezingTrap == nil {
		t.Fatalf("missing a trap")
	}
	timer := hunter.ImmolationTrap.CD.Timer
	if hunter.ExplosiveTrap.CD.Timer != timer || hunter.FreezingTrap.CD.Timer != timer {
		t.Errorf("the traps don't share one cooldown")
	}
}
