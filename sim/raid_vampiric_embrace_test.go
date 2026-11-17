package sim

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/priest"
)

// Vampiric Embrace since Blizzard's hotfixes of 2026-11-16 (t/2403040/11, live with the weekly
// maintenance of November 17/18): the party heals for 10% of the casting priest's Shadow damage and
// the priest for 20% ("receives 100% increased healing from this effect"). Another player's Shadow
// damage on the target heals no one.
func TestRaidForeverVampiricEmbrace(t *testing.T) {
	shadowReq := raidPlayer(t, "priest__shadow")
	rsr := raidOfParties(t, "mage__fire", [][]*proto.Player{
		{idle(shadowReq), idle(raidPlayer(t, "mage__fire")), idle(raidPlayer(t, "warlock__affliction"))},
	}, 1)
	rsr.Raid.Buffs = &proto.RaidBuffs{}
	sim := stepRaid(rsr)
	shadow := sim.Raid.Parties[0].Players[0].(priest.PriestAgent).GetPriest()
	mage := member(sim, 0, 1)
	lock := member(sim, 0, 2)
	if shadow.VampiricEmbrace == nil {
		t.Fatalf("the priest has no Vampiric Embrace")
	}
	raidAt(sim, 1, func(sim *core.Simulation) {
		castFree(sim, shadow.VampiricEmbrace)
		for _, c := range []*core.Character{&shadow.Character, mage} {
			c.RemoveHealth(sim, c.CurrentHealth()/2)
		}
		priestBefore, mageBefore := shadow.CurrentHealth(), mage.CurrentHealth()
		mb := topRank(t, shadow.MindBlast)
		mb.BonusHitRating = 1e6
		metrics := &mb.SpellMetrics[shadow.CurrentTarget.UnitIndex]
		damageBefore := metrics.TotalDamage
		castFree(sim, mb)
		damage := metrics.TotalDamage - damageBefore
		if damage <= 0 {
			t.Fatalf("Mind Blast dealt no damage")
		}
		near := func(name string, got, want float64) {
			if got < want-1e-6 || got > want+1e-6 {
				t.Errorf("%s healed %.3f from %.1f Shadow damage, want %.3f", name, got, damage, want)
			}
		}
		near("the priest", shadow.CurrentHealth()-priestBefore, 0.2*damage)
		near("the party's mage", mage.CurrentHealth()-mageBefore, 0.1*damage)

		// The warlock's Shadow damage on the embraced target heals no one.
		mageBefore = mage.CurrentHealth()
		for _, spell := range lock.Spellbook {
			if spell.SpellSchool == core.SpellSchoolShadow && spell.ProcMask.Matches(core.ProcMaskSpellDamage) && spell.Flags.Matches(core.SpellFlagAPL) {
				spell.CalcAndDealDamage(sim, lock.CurrentTarget, 500, spell.OutcomeAlwaysHit)
				break
			}
		}
		if got := mage.CurrentHealth() - mageBefore; got != 0 {
			t.Errorf("the warlock's Shadow damage healed the mage %.1f", got)
		}
	})
	runSteps(sim)
}
