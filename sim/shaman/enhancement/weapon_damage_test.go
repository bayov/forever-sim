package enhancement

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanWeaponDamage checks the damage of our white hits, Windfury attacks and
// Stormstrikes at level 60.
//
// Each one rolls between the weapon's min and max damage and adds AP / 14 times the weapon's
// speed. Superior Impact adds 9 to the weapon's damage, and a flat "+N damage" effect like
// Zandalarian Hero Medallion adds to every hit. Windfury adds its 333 AP at level 60.
//
// The target has no armor, so a normal hit (not a crit or a glancing blow) deals just that.
// We check that every normal hit is in its range, and that the rolls reach both ends of it.
func TestOrcShamanWeaponDamage(t *testing.T) {
	const (
		weaponMin, weaponMax, speed = 242.0, 364.0, 3.5 // Dark Edge of Insanity
		superiorImpact              = 9.0
		flatBonus                   = 20.0
		windfuryAP                  = 333.0
	)
	// Stormstrike.
	player := newOrcShaman(60, "-0000000000001", &proto.EnhancementShaman_Options{ShamanImbue: proto.WeaponImbue_WindfuryWeapon})
	player.Equipment = &proto.EquipmentSpec{}
	for range proto.ItemSlot_ItemSlotMainHand {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
	}
	player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: 21134, Enchant: 1896})
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
		extraAP  float64
		min, max float64 // the lowest and highest roll we saw, from 0 to 1
		hits     int
	}{
		{name: "white hit", spell: enh.AutoAttacks.MHAuto()},
		{name: "Windfury", spell: enh.WindfuryWeaponMH, extraAP: windfuryAP},
		{name: "Stormstrike", spell: enh.Stormstrike},
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
			bonus := superiorImpact + flatBonus + speed*(enh.GetStat(stats.AttackPower)+a.extraAP)/core.DefaultAttackPowerPerDPS
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
