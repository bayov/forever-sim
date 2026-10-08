package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanPrepull checks what a shield and totems cast before the pull leave us with at
// level 60.
//
// The sim starts at the first prepull cast with a full mana bar, and the imbue is already on.
// Each cast pays its mana then and starts the five second rule. Shield and totem time counts
// from the cast. At the pull our first swing lands right away, the GCD of a totem cast at -1
// sec is over, and nothing has hit the target before it.
func TestOrcShamanPrepull(t *testing.T) {
	prepull := []struct {
		spellID int32
		at      string
	}{
		{shaman.LightningShieldSpellId[7], "-6s"},
		{shaman.StrengthOfEarthTotemSpellId[5], "-4.5s"},
		{shaman.ManaSpringTotemSpellId[4], "-3.5s"},
		{shaman.SearingTotemSpellId[6], "-1s"},
	}
	player := newOrcShaman(60, "", &proto.EnhancementShaman_Options{ShamanImbue: proto.WeaponImbue_WindfuryWeapon})
	// Dark Edge of Insanity.
	player.Equipment = &proto.EquipmentSpec{}
	for range proto.ItemSlot_ItemSlotMainHand {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
	}
	player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: 21134})
	for _, p := range prepull {
		player.Rotation.PrepullActions = append(player.Rotation.PrepullActions, &proto.APLPrepullAction{
			Action: &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
				SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: p.spellID}},
			}}},
			DoAtValue: &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: p.at}}},
		})
	}
	sim, enh := newShamanSim(player, &proto.Debuffs{}, 10)
	shield := enh.LightningShieldAuras[7]
	shieldSpell := enh.LightningShield[7]
	searing := enh.GetSpell(core.ActionID{SpellID: shaman.SearingTotemAttackSpellId[6]})
	mainHand := enh.AutoAttacks.MHAuto()
	if shield == nil || searing == nil {
		t.Fatalf("missing Lightning Shield rank 7 or Searing Totem rank 6")
	}
	if sim.CurrentTime != -6*time.Second {
		t.Fatalf("the sim starts at %s, want -6 sec", sim.CurrentTime)
	}
	casts := func(spell *core.Spell) int32 { return spell.SpellMetrics[enh.CurrentTarget.UnitIndex].Casts }

	// Our prepull casts go first, so the shield is up by the time we look.
	at(sim, -6, func(sim *core.Simulation) {
		if aura := enh.GetAura("Windfury Imbue"); aura == nil || !aura.IsActive() {
			t.Errorf("Windfury Weapon isn't on at the first prepull cast")
		}
		if got, want := enh.CurrentMana(), enh.MaxMana()-shieldSpell.Cost.GetCurrentCost(); got != want {
			t.Errorf("after Lightning Shield: %.2f mana, want %.2f", got, want)
		}
	})

	at(sim, -0.5, func(sim *core.Simulation) {
		if got := enh.GCD.ReadyAt(); got != 0 {
			t.Errorf("the GCD of the totem at -1 sec ends at %s, want 0", got)
		}
		if got := casts(mainHand); got != 0 {
			t.Errorf("we swung %d times before the pull", got)
		}
	})

	at(sim, 0.001, func(sim *core.Simulation) {
		if got := enh.PseudoStats.FiveSecondRuleRefreshTime; got != 4*time.Second {
			t.Errorf("the five second rule ends at %s, want 4 sec after the totem at -1 sec", got)
		}
		if !shield.IsActive() || shield.ExpiresAt() != 10*time.Minute-6*time.Second {
			t.Errorf("Lightning Shield runs out at %s, want 594 sec", shield.ExpiresAt())
		}
		// The totem and its buff run out together. Searing Totem has no buff.
		expirations := []struct {
			name string
			slot int
			buff int32
			want time.Duration
		}{
			{"Strength of Earth Totem", shaman.EarthTotem, shaman.StrengthOfEarthTotemSpellId[5], 5*time.Minute - 4500*time.Millisecond},
			{"Mana Spring Totem", shaman.WaterTotem, shaman.ManaSpringTotemSpellId[4], 5*time.Minute - 3500*time.Millisecond},
			// The fire totem gets 1 ns on top, so its last attack lands before it goes.
			{"Searing Totem", shaman.FireTotem, 0, 55*time.Second - time.Second + 1},
		}
		for _, e := range expirations {
			if got := enh.TotemExpirations[e.slot]; got != e.want {
				t.Errorf("%s runs out at %s, want %s", e.name, got, e.want)
			}
			if e.buff == 0 {
				continue
			}
			if buff := enh.GetAuraByID(core.ActionID{SpellID: e.buff}); buff == nil || !buff.IsActive() || buff.ExpiresAt() != e.want {
				t.Errorf("%s: the buff isn't up until %s", e.name, e.want)
			}
		}
		if got := casts(mainHand); got != 1 {
			t.Errorf("we swung %d times at the pull, want 1", got)
		}
		if got := casts(searing); got != 0 {
			t.Errorf("Searing Totem attacked %d times before the pull", got)
		}
	})

	// Searing Totem attacks every 2.5 sec from the cast, so first at 1.5 sec.
	at(sim, 1.49, func(sim *core.Simulation) {
		if got := casts(searing); got != 0 {
			t.Errorf("Searing Totem attacked %d times by 1.49 sec, want 0", got)
		}
	})
	at(sim, 1.51, func(sim *core.Simulation) {
		if got := casts(searing); got != 1 {
			t.Errorf("Searing Totem attacked %d times by 1.51 sec, want 1", got)
		}
	})
	runSim(sim)
}
