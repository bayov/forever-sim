package feral

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/druid"
)

// Forever's druid form rules, from Blizzard's World of Warcraft: Forever class deep dive (Druid,
// "Shapeshift"), checked here against the 1.12 behaviour the Classic ruleset keeps.

const earthshaker = 17073 // two-hand mace, 175-263 damage, 3.5 speed

func newFormSim(t *testing.T, ruleset proto.Ruleset, consumes *proto.Consumes) (*core.Simulation, *FeralDruid) {
	return newFormSimBuffs(t, ruleset, consumes, &proto.RaidBuffs{})
}

func newFormSimBuffs(t *testing.T, ruleset proto.Ruleset, consumes *proto.Consumes, raidBuffs *proto.RaidBuffs) (*core.Simulation, *FeralDruid) {
	t.Helper()
	player := &proto.Player{
		Race:      proto.Race_RaceTauren,
		Class:     proto.Class_ClassDruid,
		Level:     60,
		Equipment: &proto.EquipmentSpec{Items: []*proto.ItemSpec{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {Id: earthshaker}, {}, {}}},
		Consumes:  consumes,
		Rotation:  &proto.APLRotation{},
		Spec: &proto.Player_FeralDruid{FeralDruid: &proto.FeralDruid{Options: &proto.FeralDruid_Options{
			InnervateTarget: &proto.UnitReference{},
		}}},
		Profession1: proto.Profession_Alchemy,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, raidBuffs, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: ruleset, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*FeralDruid)
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

// "The Druid's melee auto attack DPS ... in Bear Form, Cat Form, or Dire Bear Form is now the
// same as the DPS of the Druid's equipped weapon. But the speed of attack is changed to 1.0 or
// 2.5 seconds ... Abilities that deal weapon damage likewise use these damage values." Cat
// Form swings Earthshaker's 62.57 DPS at 1.0 sec, 20% either side (50.06-75.09 a hit); Classic
// keeps the paw.
func TestForeverCatSwingsItsWeapon(t *testing.T) {
	_, cat := newFormSim(t, proto.Ruleset_RulesetForever, &proto.Consumes{})
	if !cat.InForm(druid.Cat) {
		t.Fatal("not in Cat Form")
	}
	mh := cat.AutoAttacks.MH()
	dps := (175 + 263) / 2 / 3.5
	if !near(mh.SwingSpeed, 1) || !near(mh.BaseDamageMin, dps*0.8) || !near(mh.BaseDamageMax, dps*1.2) {
		t.Errorf("Forever cat weapon %.2f-%.2f at %.1f, want %.2f-%.2f at 1.0", mh.BaseDamageMin, mh.BaseDamageMax, mh.SwingSpeed, dps*0.8, dps*1.2)
	}

	_, era := newFormSim(t, proto.Ruleset_RulesetClassic, &proto.Consumes{})
	if mh := era.AutoAttacks.MH(); !near(mh.BaseDamageMin, 43.84) || !near(mh.BaseDamageMax, 65.76) {
		t.Errorf("Classic cat weapon %.2f-%.2f, want the 43.84-65.76 paw", mh.BaseDamageMin, mh.BaseDamageMax)
	}
}

// "All potions and almost all similar consumables are now useable while shapeshifted and do not
// cancel the Druid's shapeshift." "Rage potions and Thistletea are also now usable by Druids."
// The Mighty Rage Potion's 60 Strength goes up without leaving Cat Form; Classic won't drink it.
func TestForeverCatDrinksInForm(t *testing.T) {
	consumes := &proto.Consumes{DefaultPotion: proto.Potions_MightyRagePotion}
	sim, cat := newFormSim(t, proto.Ruleset_RulesetForever, consumes)
	mcd := cat.GetMajorCooldown(core.ActionID{ItemID: 13442})
	if mcd == nil || !mcd.ShouldActivate(sim, cat.GetCharacter()) {
		t.Fatal("Mighty Rage Potion isn't usable in Cat Form")
	}
	strength := cat.GetStat(stats.Strength)
	if !mcd.Spell.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Mighty Rage Potion didn't cast")
	}
	if !cat.InForm(druid.Cat) {
		t.Error("drinking left Cat Form")
	}
	// 60 Strength, times Heart of the Wild's 0 points here and no Kings: 60.
	if got := cat.GetStat(stats.Strength) - strength; !near(got, 60) {
		t.Errorf("Mighty Rage Potion gave %.1f Strength, want 60", got)
	}

	sim, era := newFormSim(t, proto.Ruleset_RulesetClassic, consumes)
	if mcd := era.GetMajorCooldown(core.ActionID{ItemID: 13442}); mcd != nil && mcd.ShouldActivate(sim, era.GetCharacter()) {
		t.Error("Classic drinks a potion in Cat Form")
	}
}

// Thistle Tea: 100 Energy for the cat under Forever, nothing under Classic.
func TestForeverCatThistleTea(t *testing.T) {
	consumes := &proto.Consumes{DefaultConjured: proto.Conjured_ConjuredRogueThistleTea}
	sim, cat := newFormSim(t, proto.Ruleset_RulesetForever, consumes)
	mcd := cat.GetMajorCooldown(core.ActionID{ItemID: 7676})
	if mcd == nil {
		t.Fatal("no Thistle Tea for the cat")
	}
	cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.Shred.EnergyMetrics())
	if !mcd.Spell.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Thistle Tea didn't cast")
	}
	if got := cat.CurrentEnergy(); !near(got, 100) {
		t.Errorf("Thistle Tea gave %.0f Energy, want 100", got)
	}
	if _, era := newFormSim(t, proto.Ruleset_RulesetClassic, consumes); era.GetMajorCooldown(core.ActionID{ItemID: 7676}) != nil {
		t.Error("Classic gives the cat Thistle Tea")
	}
}

// "Weapon enchantments and procs now all work while shapeshifted, including effects like
// weightstones, Crusader enchant, and Windfury Totem." An Elemental Sharpening Stone's 2% crit
// counts for the cat; a Dense Weightstone's 8 damage a swing goes in as the weapon's DPS, so the
// 1.0 sec cat swing gets 8 / 3.5 of it.
func TestForeverCatWeaponImbues(t *testing.T) {
	_, plain := newFormSim(t, proto.Ruleset_RulesetForever, &proto.Consumes{})
	_, stone := newFormSim(t, proto.Ruleset_RulesetForever, &proto.Consumes{MainHandImbue: proto.WeaponImbue_ElementalSharpeningStone})
	if got := stone.GetStat(stats.MeleeCrit) - plain.GetStat(stats.MeleeCrit); !near(got, 2*core.CritRatingPerCritChance) {
		t.Errorf("Elemental Sharpening Stone gave the cat %.2f crit, want 2", got)
	}
	_, era := newFormSim(t, proto.Ruleset_RulesetClassic, &proto.Consumes{MainHandImbue: proto.WeaponImbue_ElementalSharpeningStone})
	_, eraPlain := newFormSim(t, proto.Ruleset_RulesetClassic, &proto.Consumes{})
	if got := era.GetStat(stats.MeleeCrit) - eraPlain.GetStat(stats.MeleeCrit); got != 0 {
		t.Errorf("Classic Elemental Sharpening Stone gave the cat %.2f crit, want 0", got)
	}

	_, weight := newFormSim(t, proto.Ruleset_RulesetForever, &proto.Consumes{MainHandImbue: proto.WeaponImbue_DenseWeightstone})
	if got := weight.AutoAttacks.MH().DPS() - plain.AutoAttacks.MH().DPS(); !near(got, 8/3.5) {
		t.Errorf("Dense Weightstone added %.3f DPS to the cat's swing, want %.3f", got, 8/3.5)
	}
}

// Windfury Totem from the raid now works on a cat too (same quote as above).
func TestForeverCatWindfuryTotem(t *testing.T) {
	buffs := &proto.RaidBuffs{TotemWeaponBuff: proto.TotemWeaponBuff_TotemWeaponBuffWindfury}
	if _, cat := newFormSimBuffs(t, proto.Ruleset_RulesetForever, &proto.Consumes{}, buffs); cat.RaidTotemWeaponBuff == nil {
		t.Error("no Windfury Totem on the Forever cat")
	}
	if _, era := newFormSimBuffs(t, proto.Ruleset_RulesetClassic, &proto.Consumes{}, buffs); era.RaidTotemWeaponBuff != nil {
		t.Error("Windfury Totem on the Classic cat")
	}
}

// "Tiger's Fury: (Removed.)" under Forever (sim/druid/tigers_fury.go); Classic keeps the beta's
// one rank, 15% more Physical damage.
func TestForeverTigersFuryRemoved(t *testing.T) {
	if _, cat := newFormSim(t, proto.Ruleset_RulesetForever, &proto.Consumes{}); cat.TigersFury != nil {
		t.Error("the Forever cat has Tiger's Fury")
	}
	sim, era := newFormSim(t, proto.Ruleset_RulesetClassic, &proto.Consumes{})
	if era.TigersFury == nil {
		t.Fatal("the Classic cat has no Tiger's Fury")
	}
	physical := era.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical]
	era.TigersFuryAura.Activate(sim)
	if got := era.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical]; !near(got, physical*1.15) {
		t.Errorf("Physical damage with Tiger's Fury %.4f, want %.4f", got, physical*1.15)
	}
}

// Faerie Fire duty (standards 23): the cat casts Faerie Fire (Feral) from Cat Form, free, on a
// 1 sec global cooldown and a 6 sec cooldown, and it takes 505 armor for 40 sec.
func TestForeverCatFaerieFireFeral(t *testing.T) {
	sim, cat := newFormSim(t, proto.Ruleset_RulesetForever, &proto.Consumes{})
	if cat.FaerieFire == nil || cat.FaerieFire.SpellID != 17392 {
		t.Fatal("the Forever cat has no Faerie Fire (Feral)")
	}
	target := cat.CurrentTarget
	target.AddStatDynamic(sim, stats.Armor, -target.GetStat(stats.Armor))
	target.AddStatDynamic(sim, stats.Armor, 3731)
	mana := cat.CurrentMana()
	for i := 0; i < 20 && !cat.FaerieFireAuras.Get(target).IsActive(); i++ {
		cat.SetGCDTimer(sim, sim.CurrentTime)
		cat.FaerieFire.CD.Reset()
		cat.FaerieFire.Cast(sim, target)
	}
	aura := cat.FaerieFireAuras.Get(target)
	if !aura.IsActive() {
		t.Fatal("Faerie Fire (Feral) never landed")
	}
	if got := target.GetStat(stats.Armor); !near(got, 3731-505) {
		t.Errorf("armor with Faerie Fire (Feral) %.0f, want %d", got, 3731-505)
	}
	if aura.Duration.Seconds() != 40 {
		t.Errorf("Faerie Fire (Feral) lasts %s, want 40s", aura.Duration)
	}
	if got := cat.CurrentMana(); !near(got, mana) {
		t.Errorf("Faerie Fire (Feral) cost %.0f mana, want 0", mana-got)
	}
	if got := cat.FaerieFire.CD.Duration.Seconds(); got != 6 {
		t.Errorf("Faerie Fire (Feral) cooldown %.0f sec, want 6", got)
	}
	if got := cat.FaerieFire.DefaultCast.GCD.Seconds(); got != 1 {
		t.Errorf("Faerie Fire (Feral) global cooldown %.1f sec, want 1", got)
	}
}
