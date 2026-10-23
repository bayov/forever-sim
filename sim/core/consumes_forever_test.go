package core

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

func foreverTestCharacter() *Character {
	return &Character{Unit: Unit{Env: &Environment{Ruleset: proto.Ruleset_RulesetForever}, Level: 60}}
}

func expectStats(t *testing.T, name string, got stats.Stats, want stats.Stats) {
	t.Helper()
	for s := range got {
		if got[s] != want[s] {
			t.Errorf("%s: %s is %.2f, want %.2f", name, stats.Stat(s).StatName(), got[s], want[s])
		}
	}
}

// The values from client 1.60.1.70338 (SpellEffect of each item's spell).
func TestForeverConsumeStats(t *testing.T) {
	crit := stats.Stats{stats.MeleeCrit: 2 * CritRatingPerCritChance, stats.SpellCrit: 2 * SpellCritRatingPerCritChance}
	cases := []struct {
		name     string
		consumes *proto.Consumes
		want     stats.Stats
	}{
		{"Elixir of the Owl", &proto.Consumes{IntellectElixir: proto.IntellectElixir_ElixirOfTheOwl}, crit.Add(stats.Stats{stats.Intellect: 25})},
		{"Elixir of Greater Intellect", &proto.Consumes{IntellectElixir: proto.IntellectElixir_ElixirOfGreaterIntellect}, stats.Stats{stats.Intellect: 25}},
		{"Elixir of Intellect", &proto.Consumes{IntellectElixir: proto.IntellectElixir_ElixirOfIntellect}, stats.Stats{stats.Intellect: 10}},
		{"Elixir of Sages", &proto.Consumes{SpiritElixir: proto.SpiritElixir_ElixirOfSages}, crit.Add(stats.Stats{stats.Spirit: 25})},
		// Only one elixir's crit counts (wowsims/forever 2c0fd57e68).
		{"Owl and Sages", &proto.Consumes{IntellectElixir: proto.IntellectElixir_ElixirOfTheOwl, SpiritElixir: proto.SpiritElixir_ElixirOfSages}, crit.Add(stats.Stats{stats.Intellect: 25, stats.Spirit: 25})},
		{"Mongoose and Grizzly", &proto.Consumes{AgilityElixir: proto.AgilityElixir_ElixirOfTheMongoose, StrengthBuff: proto.StrengthBuff_ElixirOfTheGrizzly}, crit.Add(stats.Stats{stats.Agility: 25, stats.Strength: 25})},
		{"Mongoose, Grizzly, Owl and Whale", &proto.Consumes{AgilityElixir: proto.AgilityElixir_ElixirOfTheMongoose, StrengthBuff: proto.StrengthBuff_ElixirOfTheGrizzly,
			IntellectElixir: proto.IntellectElixir_ElixirOfTheOwl, SpiritElixir: proto.SpiritElixir_ElixirOfTheWhale},
			crit.Add(stats.Stats{stats.Agility: 25, stats.Strength: 25, stats.Intellect: 25, stats.Spirit: 25})},
		{"Elixir of the Grizzly", &proto.Consumes{StrengthBuff: proto.StrengthBuff_ElixirOfTheGrizzly}, crit.Add(stats.Stats{stats.Strength: 25})},
		{"Elixir of Ferocity", &proto.Consumes{StrengthBuff: proto.StrengthBuff_ElixirOfFerocity}, stats.Stats{stats.Strength: 18, stats.Agility: 18}},
		{"Elixir of Cunning", &proto.Consumes{AgilityElixir: proto.AgilityElixir_ElixirOfCunning}, stats.Stats{stats.Agility: 25, stats.Intellect: 25}},
		{"Elixir of Nature Power", &proto.Consumes{NaturePowerBuff: proto.NaturePowerBuff_ElixirOfNaturePower}, stats.Stats{stats.NaturePower: 40}},
		{"Greater Mageblood Elixir", &proto.Consumes{ManaRegenElixir: proto.ManaRegenElixir_GreaterMagebloodElixir}, stats.Stats{stats.MP5: 20}},
		{"Greater Cleric's Elixir", &proto.Consumes{HealingPowerBuff: proto.HealingPowerBuff_GreaterClericsElixir}, stats.Stats{stats.HealingPower: 40}},
		{"Lobster Stew", &proto.Consumes{Food: proto.Food_FoodLobsterStew}, stats.Stats{stats.SpellDamage: 28}},
		{"Mightfish Steak", &proto.Consumes{Food: proto.Food_FoodMightfishSteak}, stats.Stats{stats.AttackPower: 40, stats.RangedAttackPower: 40}},
		{"Flank au Poivre", &proto.Consumes{Food: proto.Food_FoodFlankAuPoivre}, stats.Stats{stats.Agility: 20}},
		{"Savory Stag Sliders", &proto.Consumes{Food: proto.Food_FoodSavoryStagSliders}, stats.Stats{stats.Strength: 15}},
		{"Prehistoric Pulled Raptor", &proto.Consumes{Food: proto.Food_FoodPrehistoricPulledRaptor}, stats.Stats{stats.Intellect: 20}},
		{"Wicked Smoothie", &proto.Consumes{Food: proto.Food_FoodWickedSmoothie}, stats.Stats{stats.Spirit: 20}},
	}
	for _, c := range cases {
		character := foreverTestCharacter()
		applyFoodConsumes(character, c.consumes)
		applyPhysicalBuffConsumes(character, c.consumes)
		applySpellBuffConsumes(character, c.consumes, &proto.RaidBuffs{})
		applyForeverElixirCrit(character, c.consumes)
		expectStats(t, c.name, character.GetStats(), c.want)
	}
}

// A stat elixir adds only what it gives above the Scroll of the same stat, but its crit counts in
// full, once for the two elixirs.
func TestForeverElixirOverScroll(t *testing.T) {
	character := foreverTestCharacter()
	consumes := &proto.Consumes{IntellectElixir: proto.IntellectElixir_ElixirOfTheOwl, SpiritElixir: proto.SpiritElixir_ElixirOfSages}
	applySpellBuffConsumes(character, consumes, &proto.RaidBuffs{ScrollOfIntellect: true, ScrollOfSpirit: true})
	applyForeverElixirCrit(character, consumes)
	want := stats.Stats{
		stats.Intellect: 25 - 16,
		stats.Spirit:    25 - 15,
		stats.MeleeCrit: 2 * CritRatingPerCritChance,
		stats.SpellCrit: 2 * SpellCritRatingPerCritChance,
	}
	expectStats(t, "Owl and Sages over scrolls", character.GetStats(), want)
}

func TestForeverStatPotions(t *testing.T) {
	cases := []struct {
		potion proto.Potions
		itemID int32
		stat   stats.Stat
		amount float64
	}{
		{proto.Potions_MajorSpellblastingPotion, 250937, stats.SpellDamage, 47},
		{proto.Potions_MajorFrenzyPotion, 250943, stats.AttackPower, 80},
		{proto.Potions_MajorFrenzyPotion, 250943, stats.RangedAttackPower, 80},
	}
	for _, c := range cases {
		sim := NewSim(&proto.RaidSimRequest{
			SimOptions: &proto.SimOptions{RandomSeed: 100, Ruleset: proto.Ruleset_RulesetForever},
			Raid: &proto.Raid{Parties: []*proto.Party{{
				Players: []*proto.Player{{
					Name:      "Caster",
					Class:     proto.Class_ClassShaman,
					Level:     60,
					Consumes:  &proto.Consumes{DefaultPotion: c.potion},
					Buffs:     &proto.IndividualBuffs{},
					Spec:      &proto.Player_ElementalShaman{},
					Equipment: &proto.EquipmentSpec{},
				}},
				Buffs: &proto.PartyBuffs{},
			}}},
			Encounter: &proto.Encounter{
				Targets:  []*proto.Target{{Name: "target", Level: 63}},
				Duration: 180,
			},
		}, simsignals.CreateSignals())
		sim.Reset()

		character := sim.Raid.Parties[0].Players[0].GetCharacter()
		mcd := character.GetMajorCooldown(ActionID{ItemID: c.itemID})
		if mcd == nil {
			t.Fatalf("%s: no potion cooldown", c.potion)
		}
		before := character.GetStat(c.stat)
		mcd.Spell.Cast(sim, nil)
		if got := character.GetStat(c.stat) - before; got != c.amount {
			t.Errorf("%s: %s went up %.2f, want %.2f", c.potion, c.stat.StatName(), got, c.amount)
		}
	}
}

// A plain one-hand mace for the proc tests, so the core tests need no item database.
const testProcWeaponID = 999_001

func setupProcSim(t *testing.T, ruleset proto.Ruleset, consumes *proto.Consumes) (*Simulation, *Character) {
	t.Helper()
	if _, ok := ItemsByID[testProcWeaponID]; !ok {
		ItemsByID[testProcWeaponID] = Item{
			ID: testProcWeaponID, Name: "Test Mace", Type: proto.ItemType_ItemTypeWeapon,
			WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeOneHand,
			WeaponDamageMin: 50, WeaponDamageMax: 100, SwingSpeed: 2,
		}
		t.Cleanup(func() { delete(ItemsByID, testProcWeaponID) })
	}

	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100, Ruleset: ruleset},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name:      "Caster",
				Class:     proto.Class_ClassShaman,
				Level:     60,
				Consumes:  consumes,
				Buffs:     &proto.IndividualBuffs{},
				Spec:      &proto.Player_ElementalShaman{},
				Equipment: &proto.EquipmentSpec{Items: []*proto.ItemSpec{{Id: testProcWeaponID}}},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 60}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Casts the proc spell many times and returns the lowest and highest damage of its
// normal hits (no crit, no partial resist). The caster has no spell power.
func procDamageRange(t *testing.T, sim *Simulation, spell *Spell) (float64, float64) {
	t.Helper()
	target := sim.GetTargetUnit(0)
	metrics := &spell.SpellMetrics[target.UnitIndex]
	lo, hi := 1e9, 0.0
	hits := 0
	for i := 0; i < 4000; i++ {
		before := *metrics
		spell.Cast(sim, target)
		if metrics.Hits-before.Hits != 1 || metrics.Crits != before.Crits || metrics.ResistedHits != before.ResistedHits {
			continue
		}
		damage := metrics.TotalDamage - before.TotalDamage
		lo, hi = min(lo, damage), max(hi, damage)
		hits++
	}
	if hits < 1000 {
		t.Fatalf("%s: only %d normal hits", spell.ActionID, hits)
	}
	return lo, hi
}

// Forever's oil and Dragonbreath Chili procs at level 60, from client 1.60.1.70338
// (SpellEffect, SpellLevels) and wowhead's Forever tooltips: Shadow Bolt (350026) 63-71,
// Frostbolt (350025) 39-43, Dragonbreath Chili (15851) 61-69. Classic keeps 52-61, 33-38
// and a flat 60.
func TestForeverConsumeProcDamage(t *testing.T) {
	cases := []struct {
		name     string
		consumes *proto.Consumes
		spellID  int32
		ruleset  proto.Ruleset
		min, max float64
	}{
		{"Shadow Oil", &proto.Consumes{MainHandImbue: proto.WeaponImbue_ShadowOil}, 1382, proto.Ruleset_RulesetForever, 63, 71},
		{"Frost Oil", &proto.Consumes{MainHandImbue: proto.WeaponImbue_FrostOil}, 1191, proto.Ruleset_RulesetForever, 39, 43},
		{"Dragonbreath Chili", &proto.Consumes{DragonBreathChili: true}, 15851, proto.Ruleset_RulesetForever, 61, 69},
		{"Classic Shadow Oil", &proto.Consumes{MainHandImbue: proto.WeaponImbue_ShadowOil}, 1382, proto.Ruleset_RulesetClassic, 52, 61},
		{"Classic Frost Oil", &proto.Consumes{MainHandImbue: proto.WeaponImbue_FrostOil}, 1191, proto.Ruleset_RulesetClassic, 33, 38},
		{"Classic Dragonbreath Chili", &proto.Consumes{DragonBreathChili: true}, 15851, proto.Ruleset_RulesetClassic, 60, 60},
	}
	for _, c := range cases {
		sim, character := setupProcSim(t, c.ruleset, c.consumes)
		if character.Level != 60 {
			t.Fatalf("%s: caster is level %d", c.name, character.Level)
		}
		spell := character.GetSpell(ActionID{SpellID: c.spellID})
		if spell == nil {
			t.Fatalf("%s: no proc spell %d", c.name, c.spellID)
		}
		lo, hi := procDamageRange(t, sim, spell)
		// The rolls are continuous, so after thousands of hits they come within 0.5 of
		// each end.
		if lo < c.min-1e-6 || hi > c.max+1e-6 || lo > c.min+0.5 || hi < c.max-0.5 {
			t.Errorf("%s: hits did %.2f to %.2f, want %.0f to %.0f", c.name, lo, hi, c.min, c.max)
		}
	}
}

// A level 60 shaman with the consumes, against targets copies of a level 63 boss, under Forever.
func newForeverConsumeSim(t *testing.T, consumes *proto.Consumes, encounter *proto.Encounter, targets int, profession proto.Profession) (*Simulation, *Character) {
	t.Helper()
	for range targets {
		encounter.Targets = append(encounter.Targets, &proto.Target{Name: "target", Level: 63})
	}
	if encounter.Duration == 0 {
		encounter.Duration = 180
	}
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100, Ruleset: proto.Ruleset_RulesetForever},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name:        "Caster",
				Class:       proto.Class_ClassShaman,
				Level:       60,
				Consumes:    consumes,
				Buffs:       &proto.IndividualBuffs{},
				Spec:        &proto.Player_ElementalShaman{},
				Equipment:   &proto.EquipmentSpec{},
				Profession1: profession,
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: encounter,
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// The natural flasks give 60 Stamina anywhere, and their bonus only in a natural flask zone
// (spells 1293740 to 1293743).
func TestForeverNaturalFlasks(t *testing.T) {
	_, plain := newForeverConsumeSim(t, &proto.Consumes{}, &proto.Encounter{NaturalFlaskZone: true}, 1, proto.Profession_ProfessionUnknown)
	base := plain.GetStats()
	cases := []struct {
		flask proto.Flask
		want  stats.Stats
		haste float64
	}{
		{proto.Flask_FlaskOfNaturalAccuracy, stats.Stats{stats.MeleeHit: 5 * MeleeHitRatingPerHitChance, stats.SpellHit: 5 * SpellHitRatingPerHitChance}, 1},
		{proto.Flask_FlaskOfNaturalAggression, stats.Stats{stats.MeleeCrit: 4 * CritRatingPerCritChance, stats.SpellCrit: 4 * SpellCritRatingPerCritChance}, 1},
		{proto.Flask_FlaskOfNaturalPrecision, stats.Stats{stats.Expertise: 5 * ExpertiseRatingPerExpertiseChance}, 1},
		{proto.Flask_FlaskOfNaturalSwiftness, stats.Stats{}, 1.05},
	}
	for _, c := range cases {
		for _, zone := range []bool{true, false} {
			_, character := newForeverConsumeSim(t, &proto.Consumes{Flask: c.flask}, &proto.Encounter{NaturalFlaskZone: zone}, 1, proto.Profession_ProfessionUnknown)
			gained := character.GetStats().Subtract(base)
			if gained[stats.Stamina] < 60 {
				t.Errorf("%s (zone %v): %.1f Stamina, want at least 60", c.flask, zone, gained[stats.Stamina])
			}
			want, haste := c.want, c.haste
			if !zone {
				want, haste = stats.Stats{}, 1
			}
			for _, s := range []stats.Stat{stats.MeleeHit, stats.SpellHit, stats.MeleeCrit, stats.SpellCrit, stats.Expertise} {
				if gained[s] != want[s] {
					t.Errorf("%s (zone %v): %s went up %.2f, want %.2f", c.flask, zone, s.StatName(), gained[s], want[s])
				}
			}
			ps := character.PseudoStats
			if ps.MeleeSpeedMultiplier != haste || ps.RangedSpeedMultiplier != haste || ps.CastSpeedMultiplier != haste {
				t.Errorf("%s (zone %v): melee, ranged and cast speed x%.3f, x%.3f, x%.3f, want x%.2f", c.flask, zone,
					ps.MeleeSpeedMultiplier, ps.RangedSpeedMultiplier, ps.CastSpeedMultiplier, haste)
			}
		}
	}
}

// Forever's EZ-Thro Dark Bomb, EZ-Thro Thorium Grenade and Holy Hand Grenade need no
// Engineering, and roll 225 to 675 Fire, 300 to 500 Fire and 300 to 500 Holy on every enemy. The
// Engineering explosives still need the profession.
func TestForeverOpenExplosives(t *testing.T) {
	cases := []struct {
		explosive proto.Explosive
		actionID  ActionID
		school    SpellSchool
		min, max  float64
	}{
		{proto.Explosive_ExplosiveEZThroDarkBomb, EZThroDarkBombActionID, SpellSchoolFire, 225, 675},
		{proto.Explosive_ExplosiveEZThroThoriumGrenade, EZThroThoriumGrenadeActionID, SpellSchoolFire, 300, 500},
		{proto.Explosive_ExplosiveHolyHandGrenade, HolyHandGrenadeActionID, SpellSchoolHoly, 300, 500},
	}
	for _, c := range cases {
		sim, character := newForeverConsumeSim(t, &proto.Consumes{FillerExplosive: c.explosive}, &proto.Encounter{}, 2, proto.Profession_ProfessionUnknown)
		mcd := character.GetMajorCooldown(c.actionID)
		if mcd == nil {
			t.Fatalf("%s: no cooldown without Engineering", c.explosive)
		}
		spell := mcd.Spell
		if spell.SpellSchool != c.school || spell.DefaultCast.CastTime != time.Second || spell.SharedCD.Duration != time.Minute {
			t.Errorf("%s: school %v, cast %v, shared cooldown %v; want %v, 1s, 1m", c.explosive, spell.SpellSchool, spell.DefaultCast.CastTime, spell.SharedCD.Duration, c.school)
		}
		spell.Flags |= SpellFlagIgnoreResists
		lo, hi := 1e9, 0.0
		for range 2000 {
			before := spell.SpellMetrics[0]
			spell.ApplyEffects(sim, sim.GetTargetUnit(0), spell)
			after := spell.SpellMetrics[0]
			if after.Hits == before.Hits+1 && after.Crits == before.Crits {
				d := after.TotalDamage - before.TotalDamage
				lo, hi = min(lo, d), max(hi, d)
			}
		}
		for i := range 2 {
			if got := spell.SpellMetrics[i].Hits; got < 1900 {
				t.Errorf("%s: target %d was hit %d times in 2000", c.explosive, i, got)
			}
		}
		if lo < c.min-1e-6 || hi > c.max+1e-6 || lo > c.min+(c.max-c.min)*0.01 || hi < c.max-(c.max-c.min)*0.01 {
			t.Errorf("%s: %.1f to %.1f, want %.0f to %.0f", c.explosive, lo, hi, c.min, c.max)
		}
	}

	_, engineerless := newForeverConsumeSim(t, &proto.Consumes{FillerExplosive: proto.Explosive_ExplosiveGoblinLandMine}, &proto.Encounter{}, 1, proto.Profession_ProfessionUnknown)
	if engineerless.GetMajorCooldown(GoblinLandMineActionID) != nil {
		t.Errorf("Goblin Land Mine without Engineering")
	}
}

// Dragonfire Potion (1252554) lands 1 sec after use on every enemy: 280 Fire, then 35 every 2 sec
// for 8 sec, and never misses. It reaches 12 yd. Caustic Smog Potion's cloud (1252543) lands
// after the missile's flight, at least 1.5 sec, and ticks 31 to 35 a second for 12 sec on every
// enemy. Both share the potion cooldown.
func TestForeverOffensivePotions(t *testing.T) {
	cases := []struct {
		potion       proto.Potions
		itemID       int32
		lands        time.Duration
		reach        float64
		hit          float64
		ticks        int32
		tickMin, max float64
	}{
		{proto.Potions_DragonfirePotion, 250953, time.Second, 12, 280, 4, 35, 35},
		{proto.Potions_CausticSmogPotion, 250951, time.Millisecond * 1500, 20, 0, 12, 31, 35},
	}
	for _, c := range cases {
		sim, character := newForeverConsumeSim(t, &proto.Consumes{DefaultPotion: c.potion}, &proto.Encounter{Duration: 600}, 3, proto.Profession_ProfessionUnknown)
		mcd := character.GetMajorCooldown(ActionID{ItemID: c.itemID})
		if mcd == nil {
			t.Fatalf("%s: no potion cooldown", c.potion)
		}
		spell := mcd.Spell
		spell.Flags |= SpellFlagIgnoreResists
		character.AddStatsDynamic(sim, stats.Stats{stats.SpellHit: -100 * SpellHitRatingPerHitChance, stats.SpellCrit: -100 * SpellCritRatingPerCritChance})

		character.DistanceFromTarget = c.reach + 1
		if spell.CanCast(sim, sim.GetTargetUnit(0)) {
			t.Errorf("%s: usable from %.0f yd", c.potion, character.DistanceFromTarget)
		}
		character.DistanceFromTarget = 5
		if !spell.Cast(sim, sim.GetTargetUnit(0)) {
			t.Fatalf("%s: could not be used", c.potion)
		}
		if spell.CanCast(sim, sim.GetTargetUnit(0)) {
			t.Errorf("%s: usable again at once", c.potion)
		}
		start := sim.CurrentTime
		runUntil := func(at time.Duration) {
			for len(sim.pendingActions) > 0 && sim.pendingActions[len(sim.pendingActions)-1].NextActionAt <= at {
				if sim.Step() {
					break
				}
			}
		}
		runUntil(start + c.lands - time.Millisecond)
		for i := range 3 {
			if spell.SpellMetrics[i].TotalDamage != 0 || spell.Dot(sim.GetTargetUnit(int32(i))).IsActive() {
				t.Errorf("%s: target %d took damage before %v", c.potion, i, c.lands)
			}
		}
		runUntil(start + c.lands + time.Second*13)
		for i := range 3 {
			m := spell.SpellMetrics[i]
			if m.Misses != 0 || m.Ticks != c.ticks {
				t.Errorf("%s: target %d took %d ticks and %d misses, want %d ticks", c.potion, i, m.Ticks, m.Misses, c.ticks)
			}
			if got := m.TotalDamage - m.TotalTickDamage; got != c.hit {
				t.Errorf("%s: target %d took %.1f from the hit, want %.0f", c.potion, i, got, c.hit)
			}
			if m.TotalTickDamage < float64(c.ticks)*c.tickMin || m.TotalTickDamage > float64(c.ticks)*c.max {
				t.Errorf("%s: target %d took %.1f over time, want %d ticks of %.0f to %.0f", c.potion, i, m.TotalTickDamage, c.ticks, c.tickMin, c.max)
			}
		}
	}
}
