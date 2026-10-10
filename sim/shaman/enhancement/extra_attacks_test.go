package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanExtraAttacks checks how extra attacks (Hand of Justice and Windfury Totem) line
// up with the swing timer, how often Hand of Justice procs, and how long Windfury Totem's
// attack power lasts.
func TestOrcShamanExtraAttacks(t *testing.T) {
	t.Run("swing timer", testExtraAttackSwingTimer)
	t.Run("Hand of Justice", testHandOfJustice)

	// The Forever client (70291) gives it 1 sec, where Era has 1.5 sec.
	t.Run("Windfury Totem attack power", func(t *testing.T) {
		_, enh := newGraceOfAirSim(proto.TotemWeaponBuff_TotemWeaponBuffWindfury, &proto.EnhancementShaman_Options{ShamanImbue: proto.WeaponImbue_FlametongueWeapon})
		buff := enh.GetAura("Windfury Totem Raid Buff")
		if buff == nil {
			t.Fatalf("no Windfury Totem buff")
		}
		if buff.Duration != time.Second {
			t.Errorf("Windfury Totem's attack power lasts %s, want 1s", buff.Duration)
		}
	})
}

// An extra attack is a white swing that restarts the swing timer, as in 2019 Classic
// (magey's tests). We give one when the first white swing lands, which comes 10 ms later and
// pushes the next swing back by those 10 ms. Then we give one from a Stormstrike at 5 sec,
// halfway through a swing. That one comes right away, and the next swing comes a full 3.6
// sec after it. Windfury Weapon's two attacks at 10 sec are yellow attacks, not extra
// attacks, so the next swing stays where it was.
func testExtraAttackSwingTimer(t *testing.T) {
	sim, enh := newWeaponSim(30, stormstrike, &proto.ItemSpec{Id: whirlwind}, proto.WeaponImbue_WindfuryWeapon, proto.TotemWeaponBuff_TotemWeaponBuffNone, 0, 17)
	sim.PrePull()

	// We take over the imbue's hit callback, so Windfury Weapon procs only when we cast it.
	var swings []time.Duration
	fromWhite, fromStormstrike := false, false
	enh.GetAura("Windfury Imbue").OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		switch {
		case spell.ProcMask == core.ProcMaskMeleeMHAuto:
			swings = append(swings, sim.CurrentTime)
			if !fromWhite {
				fromWhite = true
				enh.AutoAttacks.ExtraMHAttackProc(sim, 1, core.ActionID{SpellID: 15600}, spell)
			}
		case spell == enh.Stormstrike && !fromStormstrike:
			fromStormstrike = true
			enh.AutoAttacks.ExtraMHAttackProc(sim, 1, core.ActionID{SpellID: 15600}, spell)
		}
	}
	at(sim, 5, func(sim *core.Simulation) { castNow(t, sim, enh, enh.Stormstrike) })
	at(sim, 10, func(sim *core.Simulation) {
		enh.WindfuryWeaponMH.Cast(sim, enh.CurrentTarget)
		enh.WindfuryWeaponMH.Cast(sim, enh.CurrentTarget)
	})
	runSim(sim)

	// Each white hit lands 10 ms after its swing (the server batch window).
	want := []float64{0, 0.01, 3.61, 5, 8.6, 12.2, 15.8}
	if len(swings) != len(want) {
		t.Fatalf("white hits landed at %v, want swings at %v sec", swings, want)
	}
	for i, swing := range swings {
		if math.Abs((swing-core.SpellBatchWindow).Seconds()-want[i]) > 1e-6 {
			t.Fatalf("white hits landed at %v, want swings at %v sec", swings, want)
		}
	}
}

// Hand of Justice procs on 1% of landed melee hits under Forever (wowhead Forever's tooltip,
// the user's go on 2026-10-10), with a 2 sec cooldown. We count the white swings on the timer
// over a long fight. They come every 3.6 sec, so the cooldown never stops one. The extra
// swing comes 10 ms after the proc, during the cooldown, so we leave it out.
func testHandOfJustice(t *testing.T) {
	const chance = 0.01
	sim, enh := newWeaponSim(30, stormstrike, &proto.ItemSpec{Id: whirlwind}, proto.WeaponImbue_WeaponImbueUnknown, proto.TotemWeaponBuff_TotemWeaponBuffNone, common.HandOfJustice, 36000)
	sim.PrePull()
	hoj := enh.GetAura("Hand of Justice")
	if hoj == nil {
		t.Fatalf("no Hand of Justice")
	}

	roll := hoj.OnSpellHitDealt
	landed, procs := 0, 0
	lastProc := -time.Hour
	hoj.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		roll(aura, sim, spell, result)
		if spell.ProcMask != core.ProcMaskMeleeMHAuto || !result.Landed() || sim.CurrentTime-lastProc <= core.SpellBatchWindow {
			return
		}
		landed++
		// A proc moves the next swing to right now.
		if enh.AutoAttacks.MainhandSwingAt() == sim.CurrentTime {
			procs++
			lastProc = sim.CurrentTime
		}
	}
	runSim(sim)

	if landed < 8000 {
		t.Fatalf("only %d landed white hits", landed)
	}
	rate := float64(procs) / float64(landed)
	if se := math.Sqrt(chance * (1 - chance) / float64(landed)); math.Abs(rate-chance) > 4*se {
		t.Errorf("Hand of Justice procced on %.4f of %d landed white hits, want %.4f +- %.4f", rate, landed, chance, 4*se)
	}
}
