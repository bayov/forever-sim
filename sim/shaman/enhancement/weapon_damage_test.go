package enhancement

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanWeaponDamage checks the damage of our white hits, Windfury attacks and
// Stormstrikes at level 60, with a two-hander and a one-hander.
//
// Each one rolls between the weapon's min and max damage and adds AP / 14 times the weapon's
// speed. Stormstrike uses the weapon type's normalized speed plus 0.3 there (3.6 for a
// two-hander, 2.7 for a one-hander), not the weapon's own. Superior Impact adds 9 to the
// two-hander's damage, and a flat "+N damage" effect like Zandalarian Hero Medallion adds to
// every hit. Windfury adds its 333 AP at level 60.
//
// The target has no armor, so a normal hit (not a crit or a glancing blow) deals just that.
// We check that every normal hit is in its range, and that the rolls reach both ends of it.
func TestOrcShamanWeaponDamage(t *testing.T) {
	weapons := []struct {
		name                   string
		item                   *proto.ItemSpec
		weaponMin, weaponMax   float64
		speed, normalizedSpeed float64
		impact                 float64
	}{
		// Dark Edge of Insanity with Superior Impact.
		{"two-hander", &proto.ItemSpec{Id: 21134, Enchant: 1896}, 242, 364, 3.5, 3.3, 9},
		// Crul'shorukh, Edge of Chaos.
		{"one-hander", &proto.ItemSpec{Id: 19363}, 101, 188, 2.3, 2.4, 0},
	}
	for _, w := range weapons {
		t.Run(w.name, func(t *testing.T) {
			testWeaponDamage(t, w.item, w.weaponMin, w.weaponMax, w.speed, w.normalizedSpeed+shaman.StormstrikeExtraSpeed, w.impact)
		})
	}
}

func testWeaponDamage(t *testing.T, item *proto.ItemSpec, weaponMin, weaponMax, speed, stormstrikeSpeed, impact float64) {
	const (
		flatBonus  = 20.0
		windfuryAP = 333.0
	)
	// Stormstrike.
	player := newOrcShaman(60, "-0000000000001", &proto.EnhancementShaman_Options{ShamanImbue: proto.WeaponImbue_WindfuryWeapon})
	player.Equipment = &proto.EquipmentSpec{}
	for range proto.ItemSlot_ItemSlotMainHand {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
	}
	player.Equipment.Items = append(player.Equipment.Items, item)
	player.Rotation.PriorityList = []*proto.APLListItem{{Action: &proto.APLAction{
		Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
			SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 17364}},
		}},
	}}}
	// We fight for 30 min, so each attack lands enough normal hits to cover its range.
	sim, enh := newShamanSim(player, &proto.Debuffs{}, 1800)
	enh.CurrentTarget.PseudoStats.ArmorMultiplier = 0
	enh.PseudoStats.BonusPhysicalDamage += flatBonus
	metrics := enh.GetManaNotCastingMetrics()

	if enh.Stormstrike == nil || enh.WindfuryWeaponMH == nil {
		t.Fatalf("missing Stormstrike or Windfury Weapon")
	}
	attacks := []struct {
		name     string
		spell    *core.Spell
		speed    float64
		extraAP  float64
		min, max float64 // the lowest and highest roll we saw, from 0 to 1
		hits     int
	}{
		{name: "white hit", spell: enh.AutoAttacks.MHAuto(), speed: speed},
		{name: "Windfury", spell: enh.WindfuryWeaponMH, speed: speed, extraAP: windfuryAP},
		{name: "Stormstrike", spell: enh.Stormstrike, speed: stormstrikeSpeed},
	}
	for i := range attacks {
		attacks[i].min, attacks[i].max = 1, 0
	}

	imbue := enh.GetAura("Windfury Imbue")
	onHit := imbue.OnSpellHitDealt
	imbue.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		onHit(aura, sim, spell, result)
		if spell == enh.Stormstrike {
			// We never run out of mana for the next Stormstrike.
			enh.AddMana(sim, enh.MaxMana(), metrics)
		}
		for i := range attacks {
			a := &attacks[i]
			// A white hit is a copy of the auto attack spell, so we match on the action.
			if spell.ActionID != a.spell.ActionID || result.Outcome != core.OutcomeHit {
				continue
			}
			bonus := impact + flatBonus + a.speed*(enh.GetStat(stats.AttackPower)+a.extraAP)/core.DefaultAttackPowerPerDPS
			roll := (result.Damage - weaponMin - bonus) / (weaponMax - weaponMin)
			if roll < -1e-9 || roll > 1+1e-9 {
				t.Errorf("at %s: a %s for %.2f, want %.2f to %.2f", sim.CurrentTime, a.name, result.Damage, weaponMin+bonus, weaponMax+bonus)
			}
			a.min, a.max = min(a.min, roll), max(a.max, roll)
			a.hits++
		}
	}
	runSim(sim)

	for _, a := range attacks {
		if a.hits < 100 {
			t.Errorf("only %d normal %s hits", a.hits, a.name)
		}
		if a.min > 0.05 || a.max < 0.95 {
			t.Errorf("%s rolls went from %.3f to %.3f of the weapon's range, want 0 to 1", a.name, a.min, a.max)
		}
	}
}
