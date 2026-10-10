package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/shaman"
	googleProto "google.golang.org/protobuf/proto"
)

// TestOrcShamanDamageModifiers checks how flat and percent bonuses change our physical damage,
// with Rage of the Storm as the one percent bonus we have.
//
// Rage of the Storm (280604) gives Stormstrike 10% more damage and leaves white hits and
// Windfury alone. A flat "+N damage" effect adds to the hit before the percent bonuses, so
// the 10% applies to it too. We roll a 100 damage hit with +10 flat damage against a target
// with no armor, so a normal hit deals 110, or 121 for a Stormstrike with the mace.
func TestOrcShamanDamageModifiers(t *testing.T) {
	const flatBonus = 10.0
	cases := []struct {
		name                         string
		weapon                       int32
		white, windfury, stormstrike float64
	}{
		{"Whirlwind Axe", whirlwind, 110, 110, 110},
		{"Rage of the Storm", shaman.RageOfTheStorm, 110, 110, 121},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim, enh := weaponSim{level: 30, weapon: &proto.ItemSpec{Id: c.weapon}, imbue: proto.WeaponImbue_WindfuryWeapon, seconds: 60}.start()
			enh.PseudoStats.BonusPhysicalDamage += flatBonus
			if enh.Stormstrike == nil || enh.WindfuryWeaponMH == nil {
				t.Fatalf("missing Stormstrike or Windfury Weapon")
			}

			white := enh.AutoAttacks.MHAuto()
			for _, attack := range []struct {
				spell   *core.Spell
				outcome core.OutcomeApplier
				want    float64
			}{
				{white, white.OutcomeMeleeWhite, c.white},
				{enh.WindfuryWeaponMH, enh.WindfuryWeaponMH.OutcomeMeleeWeaponSpecialHitAndCrit, c.windfury},
				{enh.Stormstrike, enh.Stormstrike.OutcomeMeleeWeaponSpecialHitAndCrit, c.stormstrike},
			} {
				hit := 0.0
				for range 100 {
					result := attack.spell.CalcDamage(sim, enh.CurrentTarget, 100, attack.outcome)
					if result.Outcome == core.OutcomeHit {
						hit = result.Damage
						break
					}
				}
				if math.Abs(hit-attack.want) > 1e-9 {
					t.Errorf("%s hit for %.4f, want %.4f", attack.spell.ActionID, hit, attack.want)
				}
			}
		})
	}
}

// weaponSim is a Forever sim of seconds of an Orc shaman of level, holding weapon with imbue
// on it, against a target 2 levels up with no armor. The talents are Stormstrike when not
// set. Another shaman can give us totem (Windfury or Flametongue Totem), and a trinket that
// isn't 0 goes in the first trinket slot.
type weaponSim struct {
	level   int32
	talents string
	weapon  *proto.ItemSpec
	imbue   proto.WeaponImbue
	totem   proto.TotemWeaponBuff
	trinket int32
	debuffs *proto.Debuffs
	seconds float64
}

func (ws weaponSim) start() (*core.Simulation, *EnhancementShaman) {
	talents := ws.talents
	if talents == "" {
		talents = stormstrike
	}
	debuffs := ws.debuffs
	if debuffs == nil {
		debuffs = &proto.Debuffs{}
	}
	player := newOrcShaman(ws.level, talents, &proto.EnhancementShaman_Options{ShamanImbue: ws.imbue})
	for range proto.ItemSlot_ItemSlotMainHand {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
	}
	player.Equipment.Items = append(player.Equipment.Items, ws.weapon)
	player.Equipment.Items[proto.ItemSlot_ItemSlotTrinket1].Id = ws.trinket
	target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
	target.Level = ws.level + 2
	target.Stats[proto.Stat_StatArmor] = 0
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{TotemWeaponBuff: ws.totem}, debuffs),
		Encounter:  &proto.Encounter{Duration: ws.seconds, Targets: []*proto.Target{target}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
}

// TestOrcShamanSpellDamageModifiers checks how the percent bonuses on our spells combine, for a
// level 60 Orc shaman.
//
// Talent bonuses on the same spell add up, so Call of Flame 3 and Improved Fire Nova 2 give
// Fire Nova 1 + 15% + 20%. The Stormstrike mark and Curse of the Elements are other kinds, and
// they multiply with the talents and with each other. Forever's Curse of the Elements adds 10%
// to every magic school at level 60. So Earth Shock on a marked target deals 1.05 * 1.2 * 1.1
// with Concussion 5.
//
// We run two sims with the same seed, one with none of the bonuses and one with all of them.
// No bonus changes a roll, so both sims roll the same hits, crits and damage in the same order,
// and the ratio of each spell's total damage is exactly its multiplier.
func TestOrcShamanSpellDamageModifiers(t *testing.T) {
	// Concussion 5, Call of Flame 3, Improved Fire Nova 2, Improved Lightning Shield 3 and
	// Stormstrike.
	const bonusTalents = "050030002-0000003000001"
	newSim := func(talents string, debuffs *proto.Debuffs) (*core.Simulation, *EnhancementShaman) {
		target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
		target.Level = 60
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(newOrcShaman(60, talents, &proto.EnhancementShaman_Options{}), &proto.PartyBuffs{}, &proto.RaidBuffs{}, debuffs),
			Encounter:  &proto.Encounter{Duration: 30, Targets: []*proto.Target{target}},
			SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()
		sim.PrePull()
		return sim, sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
	}

	run := func(talents string, debuffs *proto.Debuffs, marked bool) map[string]core.SpellMetrics {
		sim, enh := newSim(talents, debuffs)
		shield := topRank(t, "Lightning Shield", enh.LightningShield)
		mark := enh.CurrentTarget.GetAuraByID(enh.Stormstrike.ActionID)
		spells := []struct {
			name  string
			spell *core.Spell
		}{
			{"Earth Shock", topRank(t, "Earth Shock", enh.EarthShock)},
			{"Flame Shock", topRank(t, "Flame Shock", enh.FlameShock)},
			{"Fire Nova", topRank(t, "Fire Nova", enh.FireNova)},
			{"Searing Totem", lastSpell(enh, shaman.SearingTotemAttackSpellId[:], 0)},
			{"Lightning Shield orb", topRank(t, "Lightning Shield orb", enh.LightningShieldProcs)},
		}
		at(sim, 1, func(sim *core.Simulation) {
			for range 500 {
				for _, s := range spells {
					// Earth Shock takes the mark, and an orb takes a charge of the shield.
					if s.name == "Earth Shock" && marked {
						mark.Activate(sim)
					}
					if s.name == "Lightning Shield orb" {
						shield.ApplyEffects(sim, enh.CurrentTarget, shield)
					}
					s.spell.ApplyEffects(sim, enh.CurrentTarget, s.spell)
				}
			}
			if marked && mark.IsActive() {
				t.Errorf("Earth Shock didn't take the Stormstrike mark")
			}
		})
		// The last Flame Shock ticks over the next 12 sec.
		runSim(sim)
		metrics := map[string]core.SpellMetrics{}
		for _, s := range spells {
			metrics[s.name] = s.spell.SpellMetrics[enh.CurrentTarget.UnitIndex]
		}
		return metrics
	}

	plain := run(stormstrike, &proto.Debuffs{}, false)
	buffed := run(bonusTalents, &proto.Debuffs{CurseOfElements: true}, true)
	cases := []struct {
		name string
		want float64
		tick float64
	}{
		{"Earth Shock", 1.05 * 1.2 * 1.1, 0},
		{"Flame Shock", 1.15 * 1.1, 1.15 * 1.1},
		{"Fire Nova", (1 + .15 + .2) * 1.1, 0},
		{"Searing Totem", 1.15 * 1.1, 0},
		{"Lightning Shield orb", 1.15 * 1.1, 0},
	}
	for _, c := range cases {
		a, b := plain[c.name], buffed[c.name]
		if a.Hits != b.Hits || a.Crits != b.Crits || a.Hits < 400 {
			t.Fatalf("%s: the sims rolled apart (%d hits and %d crits, %d and %d)", c.name, a.Hits, a.Crits, b.Hits, b.Crits)
		}
		if got := b.TotalDamage / a.TotalDamage; math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%s dealt %.4f times as much, want %.4f", c.name, got, c.want)
		}
		if c.tick == 0 {
			continue
		}
		if a.Ticks == 0 || a.Ticks != b.Ticks {
			t.Fatalf("%s ticked %d and %d times", c.name, a.Ticks, b.Ticks)
		}
		if got := b.TotalTickDamage / a.TotalTickDamage; math.Abs(got-c.tick) > 1e-9 {
			t.Errorf("%s ticked for %.4f times as much, want %.4f", c.name, got, c.tick)
		}
	}
}
