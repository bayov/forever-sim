package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanMeleeProcs checks how often a PPM proc triggers from our melee hits, with
// Crusader on a level 30 shaman's Whirlwind Axe.
//
// Each landed hit procs with a chance of PPM * weapon speed / 60, so 1 * 3.6 / 60 = 6% here.
// Stormstrike and Windfury attacks get the same chance as a white hit, and haste doesn't
// change it (2019 Classic, magey's tests). Misses and dodges never proc. Crusader gives 100
// Strength at every level up to 60 (client 70291's Holy Strength).
func TestOrcShamanMeleeProcs(t *testing.T) {
	const (
		crusader = 1900
		rolls    = 40000
		chance   = 3.6 / 60
	)
	sim, enh := newWeaponSim(30, &proto.ItemSpec{Id: whirlwind, Enchant: crusader}, proto.WeaponImbue_WindfuryWeapon, proto.TotemWeaponBuff_TotemWeaponBuffNone, 0, 60)
	holyStrength := enh.GetAura("Crusader Enchant MH")
	if holyStrength == nil || enh.Stormstrike == nil || enh.WindfuryWeaponMH == nil {
		t.Fatalf("missing Crusader, Stormstrike or Windfury Weapon")
	}

	strength := enh.GetStat(stats.Strength)
	holyStrength.Activate(sim)
	if got := enh.GetStat(stats.Strength) - strength; got != 100 {
		t.Errorf("Crusader gave %.0f Strength, want 100", got)
	}
	holyStrength.Deactivate(sim)

	// We roll each attack on its own, so Windfury Weapon mustn't add its attacks to the
	// white hits and Stormstrikes.
	enh.GetAura("Windfury Imbue").Deactivate(sim)
	enh.MultiplyAttackSpeed(sim, 1.5)

	white := enh.AutoAttacks.MHAuto()
	for _, attack := range []struct {
		spell   *core.Spell
		outcome core.OutcomeApplier
	}{
		{white, white.OutcomeMeleeWhite},
		{enh.Stormstrike, enh.Stormstrike.OutcomeMeleeWeaponSpecialHitAndCrit},
		{enh.WindfuryWeaponMH, enh.WindfuryWeaponMH.OutcomeMeleeWeaponSpecialHitAndCrit},
	} {
		landed, procs := 0, 0
		for range rolls {
			result := attack.spell.CalcAndDealDamage(sim, enh.CurrentTarget, 100, attack.outcome)
			procced := holyStrength.IsActive()
			holyStrength.Deactivate(sim)
			switch {
			case result.Landed():
				landed++
				if procced {
					procs++
				}
			case procced:
				t.Fatalf("%s procced Crusader on a %s", attack.spell.ActionID, result.Outcome)
			}
		}
		rate := float64(procs) / float64(landed)
		if se := math.Sqrt(chance * (1 - chance) / float64(landed)); math.Abs(rate-chance) > 4*se {
			t.Errorf("%s procced Crusader on %.4f of %d landed hits, want %.4f +- %.4f", attack.spell.ActionID, rate, landed, chance, 4*se)
		}
	}
}
