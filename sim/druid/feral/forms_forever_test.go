package feral

import (
	"math"
	"testing"
	"time"

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
// counts for the cat (the weightstone's damage is in TestForeverCatFlatWeaponDamage).
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

// Faerie Fire (Feral) is gone from Forever (16857, 17390 to 17392 have no client rows): the cat casts
// the caster Faerie Fire (9907), free in Cat Form, on a 1.5 sec global cooldown with no cooldown. It
// takes 505 armor for 40 sec and is the same debuff as the settings' and any other druid's. In
// caster form it costs its 115 mana.
func TestForeverCatFaerieFire(t *testing.T) {
	sim, cat := newFormSim(t, proto.Ruleset_RulesetForever, &proto.Consumes{})
	if cat.FaerieFire == nil || cat.FaerieFire.SpellID != 9907 {
		t.Fatal("the Forever cat has no caster Faerie Fire")
	}
	if cat.GetSpell(core.ActionID{SpellID: 17392}) != nil {
		t.Error("the Forever cat still has Faerie Fire (Feral)")
	}
	target := cat.CurrentTarget
	target.AddStatDynamic(sim, stats.Armor, -target.GetStat(stats.Armor))
	target.AddStatDynamic(sim, stats.Armor, 3731)
	aura := cat.FaerieFireAuras.Get(target)
	if aura != core.FaerieFireAura(target) {
		t.Error("the cat's Faerie Fire isn't the shared Faerie Fire debuff")
	}

	mana := cat.CurrentMana()
	casts := 0
	for i := 0; i < 20 && !aura.IsActive(); i++ {
		if !cat.FaerieFire.CanCast(sim, target) {
			t.Fatalf("Faerie Fire can't be cast in Cat Form (try %d)", i)
		}
		cat.FaerieFire.Cast(sim, target)
		casts++
		// No cooldown: the next try waits only for the global cooldown.
		if got := cat.GCD.ReadyAt() - sim.CurrentTime; got != 1500*time.Millisecond {
			t.Errorf("Faerie Fire's global cooldown in Cat Form %s, want 1.5s", got)
		}
		cat.SetGCDTimer(sim, sim.CurrentTime)
	}
	if !aura.IsActive() {
		t.Fatal("Faerie Fire never landed")
	}
	if !cat.InForm(druid.Cat) {
		t.Error("Faerie Fire took the druid out of Cat Form")
	}
	if got := target.GetStat(stats.Armor); !near(got, 3731-505) {
		t.Errorf("armor with Faerie Fire %.0f, want %d", got, 3731-505)
	}
	if aura.Duration.Seconds() != 40 {
		t.Errorf("Faerie Fire lasts %s, want 40s", aura.Duration)
	}
	if got := cat.CurrentMana(); !near(got, mana) {
		t.Errorf("%d Faerie Fires in Cat Form cost %.0f mana, want 0", casts, mana-got)
	}
	if cat.FaerieFire.CD.Timer != nil || cat.FaerieFire.CD.Duration != 0 {
		t.Errorf("Faerie Fire has a %s cooldown, want none", cat.FaerieFire.CD.Duration)
	}
	if !cat.FaerieFire.CanCast(sim, target) {
		t.Error("Faerie Fire can't be cast again right after its global cooldown")
	}

	// Out of Cat Form it costs its mana again.
	cat.CancelShapeshift(sim)
	if cat.InForm(druid.Cat) {
		t.Fatal("still in Cat Form")
	}
	mana = cat.CurrentMana()
	cat.FaerieFire.Cast(sim, target)
	if got := mana - cat.CurrentMana(); !near(got, 115) {
		t.Errorf("Faerie Fire in caster form cost %.1f mana, want 115", got)
	}
}

// Client 1.60.1.70425: Moonkin Aura is 3% spell crit and Leader of the Pack 3% physical crit,
// each its own raid aura, so the cat gets melee crit only from Leader of the Pack.
func TestForeverCritAuras(t *testing.T) {
	_, none := newFormSimBuffs(t, proto.Ruleset_RulesetForever, &proto.Consumes{}, &proto.RaidBuffs{})
	_, moonkin := newFormSimBuffs(t, proto.Ruleset_RulesetForever, &proto.Consumes{}, &proto.RaidBuffs{MoonkinAura: true})
	_, lotp := newFormSimBuffs(t, proto.Ruleset_RulesetForever, &proto.Consumes{}, &proto.RaidBuffs{LeaderOfThePack: true})
	_, both := newFormSimBuffs(t, proto.Ruleset_RulesetForever, &proto.Consumes{}, &proto.RaidBuffs{MoonkinAura: true, LeaderOfThePack: true})
	melee := func(c *FeralDruid) float64 { return c.GetStat(stats.MeleeCrit) - none.GetStat(stats.MeleeCrit) }
	spell := func(c *FeralDruid) float64 { return c.GetStat(stats.SpellCrit) - none.GetStat(stats.SpellCrit) }
	if !near(melee(moonkin), 0) || !near(spell(moonkin), 3*core.SpellCritRatingPerCritChance) {
		t.Errorf("Moonkin Aura gave %.2f melee and %.2f spell crit rating, want 0 and 3%%", melee(moonkin), spell(moonkin))
	}
	if !near(melee(lotp), 3*core.CritRatingPerCritChance) || !near(spell(lotp), 0) {
		t.Errorf("Leader of the Pack gave %.2f melee and %.2f spell crit rating, want 3%% and 0", melee(lotp), spell(lotp))
	}
	if !near(melee(both), 3*core.CritRatingPerCritChance) || !near(spell(both), 3*core.SpellCritRatingPerCritChance) {
		t.Errorf("both auras gave %.2f melee and %.2f spell crit rating, want 3%% each", melee(both), spell(both))
	}
}

// A permanent "+N weapon damage" enchant goes on the cat's swing whole, as does a stone's flat
// damage (wowsims/forever b5d2e0b992, a560405a65): 2H Impact's +6 and a Dense Weightstone's +8
// raise the 1.0 sec swing by 6 and 8. Out of Cat Form the weapon's own swing keeps them.
func TestForeverCatFlatWeaponDamage(t *testing.T) {
	newSim := func(enchant int32, consumes *proto.Consumes) (*core.Simulation, *FeralDruid) {
		t.Helper()
		player := &proto.Player{
			Race:      proto.Race_RaceTauren,
			Class:     proto.Class_ClassDruid,
			Level:     60,
			Equipment: &proto.EquipmentSpec{Items: []*proto.ItemSpec{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {Id: earthshaker, Enchant: enchant}, {}, {}}},
			Consumes:  consumes,
			Rotation:  &proto.APLRotation{},
			Spec: &proto.Player_FeralDruid{FeralDruid: &proto.FeralDruid{Options: &proto.FeralDruid_Options{
				InnervateTarget: &proto.UnitReference{},
			}}},
		}
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
			SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()
		sim.PrePull()
		cat := sim.Raid.Parties[0].Players[0].(*FeralDruid)
		if !cat.InForm(druid.Cat) {
			t.Fatal("not in Cat Form")
		}
		return sim, cat
	}
	humanoid := func(sim *core.Simulation, cat *FeralDruid) core.Weapon {
		cat.CancelShapeshift(sim)
		if cat.InForm(druid.Cat) {
			t.Fatal("still in Cat Form")
		}
		return *cat.AutoAttacks.MH()
	}
	plainSim, plainCat := newSim(0, &proto.Consumes{})
	plain := *plainCat.AutoAttacks.MH()
	plainHumanoid := humanoid(plainSim, plainCat)
	for _, c := range []struct {
		name     string
		enchant  int32
		consumes *proto.Consumes
		want     float64
	}{
		{"Impact", 13695, &proto.Consumes{}, 6},
		{"Dense Weightstone", 0, &proto.Consumes{MainHandImbue: proto.WeaponImbue_DenseWeightstone}, 8},
		{"both", 13695, &proto.Consumes{MainHandImbue: proto.WeaponImbue_DenseWeightstone}, 14},
	} {
		sim, cat := newSim(c.enchant, c.consumes)
		mh := cat.AutoAttacks.MH()
		if !near(mh.BaseDamageMin-plain.BaseDamageMin, c.want) || !near(mh.BaseDamageMax-plain.BaseDamageMax, c.want) {
			t.Errorf("%s added %.2f-%.2f to the cat's swing, want %.0f", c.name, mh.BaseDamageMin-plain.BaseDamageMin, mh.BaseDamageMax-plain.BaseDamageMax, c.want)
		}
		weapon := humanoid(sim, cat)
		if !near(weapon.BaseDamageMin-plainHumanoid.BaseDamageMin, c.want) || !near(weapon.BaseDamageMax-plainHumanoid.BaseDamageMax, c.want) {
			t.Errorf("%s added %.2f-%.2f to the swing out of Cat Form, want %.0f", c.name, weapon.BaseDamageMin-plainHumanoid.BaseDamageMin, weapon.BaseDamageMax-plainHumanoid.BaseDamageMax, c.want)
		}
	}
}

// Omen of Clarity rolls 2 procs a minute on a cat's special at the 1.0 sec swing speed under
// Forever, 3.3% a Shred (wowsims/forever ca7b3ab462: the equipped weapon's speed "is a game bug
// fixed in the next beta build"), and at Earthshaker's 3.5 sec, 11.7%, under Classic.
func TestForeverOmenOfClaritySwingSpeed(t *testing.T) {
	for _, c := range []struct {
		ruleset proto.Ruleset
		want    float64
	}{
		{proto.Ruleset_RulesetForever, 2 * 1.0 / 60},
		{proto.Ruleset_RulesetClassic, 2 * 3.5 / 60},
	} {
		sim, cat := newFormSim(t, c.ruleset, &proto.Consumes{})
		omen := cat.GetAura("Omen of Clarity")
		const trials = 6000
		procs := 0
		for range trials {
			// Past the 10 sec cooldown each time.
			sim.CurrentTime += 11 * time.Second
			cat.ClearcastingAura.Deactivate(sim)
			result := cat.Shred.NewResult(cat.CurrentTarget)
			result.Outcome = core.OutcomeHit
			omen.OnSpellHitDealt(omen, sim, cat.Shred.Spell, result)
			if cat.ClearcastingAura.IsActive() {
				procs++
			}
		}
		if got := float64(procs) / trials; math.Abs(got-c.want) > 0.01 {
			t.Errorf("%s: Omen of Clarity on %.1f%% of Shreds, want %.1f%%", c.ruleset, 100*got, 100*c.want)
		}
	}
}
