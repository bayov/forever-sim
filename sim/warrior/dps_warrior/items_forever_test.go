package dpswarrior

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

func newWarriorWithMainHand(t *testing.T, ruleset proto.Ruleset, mainHand int32) (*core.Simulation, *DpsWarrior) {
	return newWarriorWithGear(t, ruleset, mainHand, 0, &proto.Consumes{}, &proto.Debuffs{})
}

func newWarriorWithGear(t *testing.T, ruleset proto.Ruleset, mainHand int32, offHand int32, consumes *proto.Consumes, debuffs *proto.Debuffs) (*core.Simulation, *DpsWarrior) {
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: mainHand}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: offHand}
	player := &proto.Player{
		Race:      proto.Race_RaceTauren,
		Class:     proto.Class_ClassWarrior,
		Level:     60,
		Equipment: &proto.EquipmentSpec{Items: items},
		Consumes:  consumes,
		Rotation:  &proto.APLRotation{},
		Spec:      &proto.Player_Warrior{Warrior: &proto.Warrior{Options: &proto.Warrior_Options{}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, debuffs),
		Encounter:  &proto.Encounter{Duration: 600, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: ruleset, RandomSeed: 1, ExactDamage: true},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*DpsWarrior)
}

// procRange casts the spell n times and returns the lowest and highest damage of its landed,
// uncritted casts, with resists off.
func procRange(sim *core.Simulation, war *DpsWarrior, spell *core.Spell, n int) (float64, float64) {
	spell.Flags |= core.SpellFlagIgnoreResists
	metrics := &spell.SpellMetrics[war.CurrentTarget.UnitIndex]
	lo, hi := 1e9, 0.0
	for i := 0; i < n; i++ {
		damage, hits := metrics.TotalDamage, metrics.Hits
		spell.Cast(sim, war.CurrentTarget)
		if metrics.Hits == hits+1 {
			d := metrics.TotalDamage - damage
			lo, hi = min(lo, d), max(hi, d)
		}
	}
	return lo, hi
}

// Alcor's Sunrazor's Firebolt (18833) is 105 with variance 0.2, 94.5 to 115.5, since Forever
// client 1.60.1.70425 (75 to 105 before, and in Classic).
func TestForeverAlcorsSunrazorFirebolt(t *testing.T) {
	for _, c := range []struct {
		ruleset  proto.Ruleset
		min, max float64
	}{
		{proto.Ruleset_RulesetForever, 94.5, 115.5},
		{proto.Ruleset_RulesetClassic, 75, 105},
	} {
		sim, war := newWarriorWithMainHand(t, c.ruleset, 14555)
		firebolt := war.GetSpell(core.ActionID{SpellID: 18833})
		if firebolt == nil {
			t.Fatalf("%s: no Firebolt with Alcor's Sunrazor", c.ruleset)
		}
		lo, hi := procRange(sim, war, firebolt, 2000)
		// 2000 rolls come within 1% of each end.
		if lo < c.min-1e-6 || hi > c.max+1e-6 || lo > c.min+0.3 || hi < c.max-0.3 {
			t.Errorf("%s: Firebolt %.1f to %.1f, want %.1f to %.1f", c.ruleset, lo, hi, c.min, c.max)
		}
	}
}

// Shadowstrike's Drain Life (21170) gets no spell power under Forever since client 70009.
func TestForeverShadowstrikeNoSpellPower(t *testing.T) {
	_, war := newWarriorWithMainHand(t, proto.Ruleset_RulesetForever, 17074)
	if got := war.GetSpell(core.ActionID{SpellID: 461683}).BonusCoefficient; got != 0 {
		t.Errorf("Forever Shadowstrike proc coefficient %.2f, want 0", got)
	}
	_, era := newWarriorWithMainHand(t, proto.Ruleset_RulesetClassic, 17074)
	if got := era.GetSpell(core.ActionID{SpellID: 461683}).BonusCoefficient; got != 1 {
		t.Errorf("Classic Shadowstrike proc coefficient %.2f, want 1", got)
	}
}

// The chance-on-hit damage the Forever client's rows state (wowsims/forever's store), against
// Classic's.
func TestForeverWeaponProcDamage(t *testing.T) {
	for _, c := range []struct {
		name                   string
		item                   int32
		foreverSpell, eraSpell int32
		foreverMin, foreverMax float64
		eraMin, eraMax         float64
	}{
		{"Fiery War Axe", 870, 18796, 18796, 231, 293, 155, 197},
		{"Volcanic Hammer", 12792, 18082, 18082, 200, 256, 100, 128},
		{"Hammer of the Northern Wind", 810, 13439, 13439, 40, 60, 20, 30},
		{"Linken's Sword of Mastery", 11902, 18089, 18089, 74, 122, 45, 75},
		{"Baron Charr's Sceptre", 18671, 13442, 13442, 90, 90, 35, 35},
		{"Nightblade", 1982, 18211, 18211, 161, 301, 125, 275},
		{"Ebon Hand", 19170, 18211, 18211, 161, 301, 125, 275},
		{"Teebu's Blazing Longsword", 1728, 1300753, 18086, 140, 140, 150, 150},
		{"The Lobotomizer", 19324, 1290950, 24388, 200, 300, 200, 300},
		{"Gryphon Rider's Stormhammer", 9651, 18081, 18081, 77, 105, 91, 125},
	} {
		for _, r := range []struct {
			ruleset  proto.Ruleset
			spellID  int32
			min, max float64
		}{
			{proto.Ruleset_RulesetForever, c.foreverSpell, c.foreverMin, c.foreverMax},
			{proto.Ruleset_RulesetClassic, c.eraSpell, c.eraMin, c.eraMax},
		} {
			sim, war := newWarriorWithMainHand(t, r.ruleset, c.item)
			spell := war.GetSpell(core.ActionID{SpellID: r.spellID})
			if spell == nil {
				t.Errorf("%s %s: no proc spell %d", r.ruleset, c.name, r.spellID)
				continue
			}
			lo, hi := procRange(sim, war, spell, 2000)
			if lo < r.min-1e-6 || hi > r.max+1e-6 || lo > r.min+0.02*(r.max-r.min)+1e-6 || hi < r.max-0.02*(r.max-r.min)-1e-6 {
				t.Errorf("%s %s: %.1f to %.1f, want %.0f to %.0f", r.ruleset, c.name, lo, hi, r.min, r.max)
			}
		}
	}
}

// Thunderstrike's blast goes to up to 3 different targets, once each: one target takes one blast,
// three take one each. Forever's is Chain Lightning 21179, Classic's 461686, 150 to 250 first.
func TestThunderstrikeHitsEachTargetOnce(t *testing.T) {
	for _, c := range []struct {
		ruleset proto.Ruleset
		spellID int32
	}{
		{proto.Ruleset_RulesetForever, 21179},
		{proto.Ruleset_RulesetClassic, 461686},
	} {
		for _, targets := range []int{1, 2, 3, 4} {
			items := make([]*proto.ItemSpec, 17)
			for i := range items {
				items[i] = &proto.ItemSpec{}
			}
			items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 17223}
			player := &proto.Player{
				Race: proto.Race_RaceTauren, Class: proto.Class_ClassWarrior, Level: 60,
				Equipment: &proto.EquipmentSpec{Items: items}, Consumes: &proto.Consumes{}, Rotation: &proto.APLRotation{},
				Spec: &proto.Player_Warrior{Warrior: &proto.Warrior{Options: &proto.Warrior_Options{}}},
			}
			encounter := &proto.Encounter{Duration: 600}
			for range targets {
				encounter.Targets = append(encounter.Targets, core.NewDefaultTarget())
			}
			sim := core.NewSim(&proto.RaidSimRequest{
				Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
				Encounter:  encounter,
				SimOptions: &proto.SimOptions{Ruleset: c.ruleset, RandomSeed: 1, ExactDamage: true},
			}, simsignals.CreateSignals())
			sim.Reset()
			sim.PrePull()
			war := sim.Raid.Parties[0].Players[0].(*DpsWarrior)
			spell := war.GetSpell(core.ActionID{SpellID: c.spellID})
			if spell == nil {
				t.Fatalf("%s: no Thunderstrike proc %d", c.ruleset, c.spellID)
			}
			spell.Flags |= core.SpellFlagIgnoreResists
			war.AddStatsDynamic(sim, stats.Stats{stats.SpellHit: 100 * core.SpellHitRatingPerHitChance})
			const casts = 200
			for range casts {
				spell.Cast(sim, war.CurrentTarget)
			}
			for i := range targets {
				m := spell.SpellMetrics[sim.GetTargetUnit(int32(i)).UnitIndex]
				blasts := m.Hits + m.Crits + m.Misses
				lo, hi := int32(casts), int32(casts)
				switch {
				case i >= 3:
					lo, hi = 0, 0
				case c.ruleset == proto.Ruleset_RulesetClassic && i == 0:
					// Classic's first roll is an outcome check of its own, counted as a hit or a miss.
					lo, hi = casts, 2*casts
				case c.ruleset == proto.Ruleset_RulesetClassic:
					// Classic's jumps go out only after the first roll lands.
					lo = casts * 9 / 10
				}
				if blasts < lo || blasts > hi {
					t.Errorf("%s, %d targets: target %d took %d blasts in %d procs, want %d to %d", c.ruleset, targets, i, blasts, casts, lo, hi)
				}
			}
		}
	}
}

// Nightfall's Spell Vulnerability (23605) is 15% more damage taken from the six magic schools,
// a multiplier, in both rulesets.
func TestNightfallSpellVulnerability(t *testing.T) {
	for _, ruleset := range []proto.Ruleset{proto.Ruleset_RulesetForever, proto.Ruleset_RulesetClassic} {
		sim, war := newWarriorWithMainHand(t, ruleset, 19169)
		target := war.CurrentTarget
		aura := target.GetAura("Spell Vulnerability")
		if aura == nil {
			t.Fatalf("%s: no Spell Vulnerability on the target", ruleset)
		}
		aura.Activate(sim)
		taken := target.PseudoStats.SchoolDamageTakenMultiplier
		for school := stats.SchoolIndexArcane; school < stats.SchoolLen; school++ {
			if !nearly(taken[school], 1.15) {
				t.Errorf("%s: school %d takes %.3f, want 1.15", ruleset, school, taken[school])
			}
		}
		if taken[stats.SchoolIndexPhysical] != 1 {
			t.Errorf("%s: Physical takes %.3f, want 1", ruleset, taken[stats.SchoolIndexPhysical])
		}
		aura.Deactivate(sim)
		if !nearly(target.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire], 1) {
			t.Errorf("%s: Fire takes %.3f after it fades, want 1", ruleset, target.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire])
		}
	}
}

// Forever's Annihilator takes 165 armor a stack and stacks with Faerie Fire; Classic's takes 200
// a stack.
func TestForeverAnnihilatorArmorShatter(t *testing.T) {
	for _, c := range []struct {
		ruleset proto.Ruleset
		debuffs *proto.Debuffs
		want    float64
	}{
		{proto.Ruleset_RulesetForever, &proto.Debuffs{FaerieFire: true}, 3 * 165},
		{proto.Ruleset_RulesetClassic, &proto.Debuffs{}, 3 * 200},
	} {
		sim, war := newWarriorWithGear(t, c.ruleset, 12798, 0, &proto.Consumes{}, c.debuffs)
		target := war.CurrentTarget
		before := target.GetStat(stats.Armor)
		shatter := target.GetAura("Armor Shatter")
		shatter.Activate(sim)
		if !shatter.IsActive() {
			t.Fatalf("%s: Armor Shatter did not land", c.ruleset)
		}
		shatter.SetStacks(sim, 3)
		if got := before - target.GetStat(stats.Armor); !nearly(got, c.want) {
			t.Errorf("%s: 3 stacks take %.0f armor, want %.0f", c.ruleset, got, c.want)
		}
	}
}

// Forever's Elemental Sharpening Stone crits only the attacks of the weapon it is on.
func TestForeverElementalSharpeningStoneHand(t *testing.T) {
	stone := proto.WeaponImbue_ElementalSharpeningStone
	for _, c := range []struct {
		name     string
		consumes *proto.Consumes
		mh, oh   float64
	}{
		{"main hand", &proto.Consumes{MainHandImbue: stone}, 2, 0},
		{"off hand", &proto.Consumes{OffHandImbue: stone}, 0, 2},
		{"both", &proto.Consumes{MainHandImbue: stone, OffHandImbue: stone}, 2, 2},
	} {
		_, plain := newWarriorWithGear(t, proto.Ruleset_RulesetForever, 11902, 2164, &proto.Consumes{}, &proto.Debuffs{})
		_, war := newWarriorWithGear(t, proto.Ruleset_RulesetForever, 11902, 2164, c.consumes, &proto.Debuffs{})
		crit := func(w *DpsWarrior, spell *core.Spell) float64 {
			return w.GetStat(stats.MeleeCrit) + spell.BonusCritRating
		}
		if got := crit(war, war.AutoAttacks.MHAuto()) - crit(plain, plain.AutoAttacks.MHAuto()); !nearly(got, c.mh) {
			t.Errorf("%s: main hand +%.1f crit, want +%.1f", c.name, got, c.mh)
		}
		if got := crit(war, war.AutoAttacks.OHAuto()) - crit(plain, plain.AutoAttacks.OHAuto()); !nearly(got, c.oh) {
			t.Errorf("%s: off hand +%.1f crit, want +%.1f", c.name, got, c.oh)
		}
	}
}

func nearly(a, b float64) bool { return a-b < 1e-6 && b-a < 1e-6 }
