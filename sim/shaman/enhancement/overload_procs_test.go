package enhancement

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanOverloadProcs checks that Lightning Overload's copies are procs: a copy is a
// triggered spell without Not a Proc, so listeners without Can Proc From Procs don't hear it.
// Judgement of Wisdom (20355) and Elemental Devastation are such listeners, so a copy's crit gives
// neither mana nor the melee crit, where the bolt's own crit gives both (the wowsims/forever
// comparison, 2026-11-02).
//
// Elemental Devastation 3/3 and Lightning Overload 3/3, with Judgement of Wisdom on the boss. Each
// spell crits 60 times; the bolt's 50% Judgement of Wisdom roll misses all 60 one time in 10^18.
func TestOrcShamanOverloadProcs(t *testing.T) {
	const rolls = 60
	sim, enh := newShamanSim(newOrcShaman(60, "00300000000003", &proto.EnhancementShaman_Options{}), &proto.Debuffs{JudgementOfWisdom: true}, 10)
	devastation := enh.GetAura("Elemental Devastation Proc")
	if devastation == nil {
		t.Fatalf("no Elemental Devastation Proc")
	}
	lb := topRank(t, "Lightning Bolt", enh.LightningBolt)
	cl := topRank(t, "Chain Lightning", enh.ChainLightning)
	cases := []struct {
		name  string
		spell *core.Spell
		procs bool
	}{
		{"Lightning Bolt", lb, true},
		{"Lightning Bolt's overload", enh.LightningBoltOverload[lb.Rank], false},
		{"Chain Lightning", cl, true},
		{"Chain Lightning's overload", enh.ChainLightningOverload[cl.Rank], false},
	}
	at(sim, 0.5, func(sim *core.Simulation) {
		for _, c := range cases {
			if c.spell == nil {
				t.Fatalf("no %s", c.name)
			}
			enh.SpendMana(sim, enh.CurrentMana(), enh.GetManaNotCastingMetrics())
			devastated := false
			for range rolls {
				c.spell.CalcAndDealDamage(sim, enh.CurrentTarget, 100, forcedOutcome(core.OutcomeCrit))
				if devastation.IsActive() {
					devastated = true
					devastation.Deactivate(sim)
				}
			}
			if devastated != c.procs {
				t.Errorf("%s crits gave Elemental Devastation: %t, want %t", c.name, devastated, c.procs)
			}
			if gotMana := enh.CurrentMana() > 0; gotMana != c.procs {
				t.Errorf("%s crits gave Judgement of Wisdom mana: %t, want %t", c.name, gotMana, c.procs)
			}
		}
	})
	runSim(sim)
}
