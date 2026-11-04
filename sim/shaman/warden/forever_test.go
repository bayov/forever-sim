package warden

import (
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
)

// The Warden's tank tests (docs/tank_shaman.md): every ability and talent a shaman tank uses,
// at level 60 against the level 63 test boss, which attacks the Warden.

const (
	deathbringer       = 17068 // a 2.9 sec one-handed axe
	draconianDeflector = 12602 // a shield, 40 block value and 10 Defense
)

// wardenTalents is the Warden's Level 60 build, the base the talent tests change one talent of.
const wardenTalents = "35050013-0502333135510102"

// The talents' places in their trees (proto/shaman.proto ShamanTalents; the November 2 tree, with
// Improved Ghost Wolf moved to row 4, after Stormstrike).
const (
	treeElemental, treeEnhancement, treeRestoration = 0, 1, 2

	convection, concussion, reverberation, elementalWarding                                          = 0, 1, 3, 7
	thunderingStrikes, guardianTotems, mentalDexterity, improvedLightningShield, elementalWeapons    = 1, 3, 4, 5, 6
	shamanisticFocus, anticipation, toughness, spiritWeapons, mentalQuickness, improvedStormstrikeIx = 7, 8, 9, 13, 14, 15
	tidalFocus, naturalGrace, improvedReincarnation, ancestralHealing                                = 1, 3, 5, 6
)

// withTalent is talents with one talent set to points.
func withTalent(talents string, tree, index, points int) string {
	trees := strings.Split(talents, "-")
	for len(trees) < 3 {
		trees = append(trees, "")
	}
	digits := []byte(trees[tree])
	for len(digits) <= index {
		digits = append(digits, '0')
	}
	digits[index] = byte('0' + points)
	trees[tree] = string(digits)
	return strings.Join(trees, "-")
}

// newWarden builds a level 60 Orc Warden with Deathbringer and Draconian Deflector, Rockbiter
// Weapon from its options (or imbue), no rotation, tanking the level 63 test boss.
func newWarden(t *testing.T, talents string, imbue proto.WeaponImbue) (*core.Simulation, *WardenShaman) {
	t.Helper()
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: deathbringer}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: draconianDeflector}
	player := &proto.Player{
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassShaman,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{Items: items},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec: &proto.Player_WardenShaman{WardenShaman: &proto.WardenShaman{
			Options: &proto.WardenShaman_Options{ShamanImbue: imbue},
		}},
	}
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Duration: 600, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1, ExactDamage: true},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*WardenShaman)
}

func near(t *testing.T, what string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Errorf("%s is %.5f, want %.5f", what, got, want)
	}
}

// castNow casts spell at the current time, the GCD freed first.
func castNow(t *testing.T, sim *core.Simulation, w *WardenShaman, spell *core.Spell) {
	t.Helper()
	w.SetGCDTimer(sim, sim.CurrentTime)
	w.AddMana(sim, w.MaxMana(), w.GetManaNotCastingMetrics())
	if !spell.Cast(sim, w.CurrentTarget) {
		t.Fatalf("at %s: %s didn't cast", sim.CurrentTime, spell.ActionID)
	}
}

// advance runs the sim to at, with nothing else scheduled but what's already pending.
func advance(sim *core.Simulation, at time.Duration) {
	sim.AddPendingAction(&core.PendingAction{NextActionAt: at, OnAction: func(*core.Simulation) {}})
	for sim.CurrentTime < at && !sim.Step() {
	}
}

// TestForeverWardenTankSheet checks what the boss's swing at the Warden rolls (core.TankSheet): a
// shaman with a shield blocks (5% less the level gap's 0.6%, plus Defense), and parries only with
// Spirit Weapons (the server grants it, client 16268 has no parry effect). Its block value is the
// shield's plus Strength / 20 less 1, as every Classic shield user's.
func TestForeverWardenTankSheet(t *testing.T) {
	sim, w := newWarden(t, wardenTalents, proto.WeaponImbue_RockbiterWeapon)
	boss := sim.Encounter.TargetUnits[0]
	sheet := core.TankSheet(&w.Unit, boss)
	// The shield's 10 Defense: 0.4% more of each of miss, parry and block.
	near(t, "Defense", sheet.Defense, 310, 1e-9)
	near(t, "Block", sheet.Block, 0.05-0.006+0.004, 1e-9)
	near(t, "Parry with Spirit Weapons", sheet.Parry, 0.05-0.006+0.004, 1e-9)
	near(t, "Block value", sheet.BlockValue, 40+w.GetStat(stats.Strength)/20-1, 1e-9)

	_, noShield := newWarden(t, wardenTalents, proto.WeaponImbue_RockbiterWeapon)
	noShield.PseudoStats.CanBlock = false
	if got := core.TankSheet(&noShield.Unit, boss).Block; got != 0 {
		t.Errorf("block %.4f without a shield, want 0", got)
	}

	_, none := newWarden(t, withTalent(wardenTalents, treeEnhancement, spiritWeapons, 0), proto.WeaponImbue_RockbiterWeapon)
	if got := core.TankSheet(&none.Unit, boss).Parry; got != 0 {
		t.Errorf("parry %.4f without Spirit Weapons, want 0", got)
	}
}

// TestForeverWardenTalents checks every talent a Warden can reach that moves its sheet, its threat
// or its abilities, each one on against off from the Level 60 build (client 70753's talent curves,
// forever-wiki/classes/shaman.md). DPS-only talents (Flurry, Maelstrom Weapon, Elemental
// Devastation, ...) are tested in sim/shaman/enhancement.
func TestForeverWardenTalents(t *testing.T) {
	type measure func(t *testing.T, sim *core.Simulation, w *WardenShaman) float64
	sheet := func(f func(*proto.TankSheet) float64) measure {
		return func(t *testing.T, sim *core.Simulation, w *WardenShaman) float64 {
			return f(core.TankSheet(&w.Unit, sim.Encounter.TargetUnits[0]))
		}
	}
	stat := func(s stats.Stat) measure {
		return func(_ *testing.T, _ *core.Simulation, w *WardenShaman) float64 { return w.GetStat(s) }
	}
	earthShock := func(w *WardenShaman) *core.Spell { return w.EarthShock[7] }
	cases := []struct {
		name                string
		tree, index, points int // the talent and its points (the other side is the base build's)
		measure             measure
		ratio               bool    // compare with / without as a ratio, not a difference
		want                float64 // the talent's side against the base build's
		tolerance           float64
	}{
		{"Anticipation 3/3: 6% dodge", treeEnhancement, anticipation, 0,
			sheet(func(s *proto.TankSheet) float64 { return s.Dodge }), false, -0.06, 1e-9},
		{"Toughness 5/5: 10% Stamina", treeEnhancement, toughness, 0,
			stat(stats.Stamina), true, 1 / 1.10, 0.015}, // the server drops the fraction
		{"Improved Reincarnation 2/2: 4% health", treeRestoration, improvedReincarnation, 2,
			stat(stats.Health), true, 1.04, 0.001},
		{"Thundering Strikes 5/5: 5% crit", treeEnhancement, thunderingStrikes, 0,
			stat(stats.MeleeCrit), false, -5 * core.CritRatingPerCritChance, 1e-9},
		{"Tidal Focus 5/5: 5% hit", treeRestoration, tidalFocus, 5,
			stat(stats.MeleeHit), false, 5 * core.MeleeHitRatingPerHitChance, 1e-9},
		{"Elemental Warding 3/3: 10% less Fire, Frost and Nature damage", treeElemental, elementalWarding, 0,
			func(_ *testing.T, _ *core.Simulation, w *WardenShaman) float64 {
				return w.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire] *
					w.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFrost] *
					w.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexNature]
			}, true, 1 / (0.9 * 0.9 * 0.9), 1e-9},
		{"Elemental Weapons 3/3: Rockbiter's 653 attack power 20% more", treeEnhancement, elementalWeapons, 0,
			stat(stats.AttackPower), false, -653 * 0.2, 0.01},
		{"Mental Dexterity 3/3: attack power from 150% of Intellect (the November 2 notes)", treeEnhancement, mentalDexterity, 0,
			func(_ *testing.T, _ *core.Simulation, w *WardenShaman) float64 {
				return w.GetStat(stats.AttackPower) - []float64{0, .5, 1, 1.5}[w.Talents.MentalDexterity]*w.GetStat(stats.Intellect)
			}, false, 0, 0.01},
		{"Mental Quickness 2/2: spell power from 30% of Intellect", treeEnhancement, mentalQuickness, 2,
			func(_ *testing.T, _ *core.Simulation, w *WardenShaman) float64 {
				return w.GetStat(stats.SpellPower) - 0.15*float64(w.Talents.MentalQuickness)*w.GetStat(stats.Intellect)
			}, false, 0, 0.01},
		{"Spirit Weapons with Rockbiter: 0.7 x 1.86 threat", treeEnhancement, spiritWeapons, 0,
			func(_ *testing.T, _ *core.Simulation, w *WardenShaman) float64 { return w.PseudoStats.ThreatMultiplier },
			true, 1 / 1.302, 1e-9},
		{"Natural Grace 3/3: 15% less threat from Earth Shock", treeRestoration, naturalGrace, 3,
			func(_ *testing.T, _ *core.Simulation, w *WardenShaman) float64 { return earthShock(w).ThreatMultiplier },
			true, 0.85, 1e-9},
		{"Concussion 5/5: Earth Shock 5% more damage", treeElemental, concussion, 0,
			func(_ *testing.T, _ *core.Simulation, w *WardenShaman) float64 {
				return earthShock(w).DamageMultiplierAdditive
			},
			false, -0.05, 1e-9},
		{"Reverberation 5/5: shocks 1 sec sooner", treeElemental, reverberation, 0,
			func(_ *testing.T, _ *core.Simulation, w *WardenShaman) float64 {
				return earthShock(w).CD.Duration.Seconds()
			},
			false, 1, 1e-9},
		{"Convection 3/5 with Shamanistic Focus: Earth Shock costs 450 x 0.94 x 0.55", treeElemental, convection, 0,
			func(_ *testing.T, _ *core.Simulation, w *WardenShaman) float64 {
				return earthShock(w).Cost.GetCurrentCost()
			},
			false, 450 * 0.06 * 0.55, 1e-9},
		{"Shamanistic Focus: Lightning Shield costs 45% less", treeEnhancement, shamanisticFocus, 0,
			func(_ *testing.T, _ *core.Simulation, w *WardenShaman) float64 {
				return w.LightningShield[7].Cost.GetCurrentCost()
			},
			false, 370 * 0.45, 1e-9},
		{"Improved Lightning Shield 3/3: orbs 15% more damage", treeEnhancement, improvedLightningShield, 0,
			func(_ *testing.T, _ *core.Simulation, w *WardenShaman) float64 {
				return w.LightningShieldProcs[7].DamageMultiplier
			},
			true, 1 / 1.15, 1e-9},
		{"Guardian Totems 2/2: Stoneskin takes 20% more off", treeEnhancement, guardianTotems, 0,
			func(t *testing.T, sim *core.Simulation, w *WardenShaman) float64 {
				castNow(t, sim, w, w.StoneskinTotem[6])
				return w.PseudoStats.BonusPhysicalDamageTakenBeforeArmor
			}, false, 36 - 30, 1e-9},
		// 25% more armor from items only (16177: "Increases your armor value from items"), so the
		// measure is the gain over the armor from items, plus 1 (1 for the base build, which gains none).
		{"Ancestral Healing 3/3: a heal crit gives 25% of all armor (aura 101, healer_mechanics-20261104-3)", treeRestoration, ancestralHealing, 3,
			func(t *testing.T, sim *core.Simulation, w *WardenShaman) float64 {
				before := w.Armor()
				heal := w.LesserHealingWave[len(w.LesserHealingWave)-1]
				w.SetGCDTimer(sim, sim.CurrentTime)
				heal.CalcAndDealHealing(sim, &w.Unit, 100, func(_ *core.Simulation, r *core.SpellResult, _ *core.AttackTable) {
					r.Outcome = core.OutcomeCrit
				})
				return w.Armor() / before
			}, true, 1.25, 1e-6},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim, base := newWarden(t, wardenTalents, proto.WeaponImbue_RockbiterWeapon)
			other := withTalent(wardenTalents, c.tree, c.index, c.points)
			if other == wardenTalents {
				t.Fatalf("%s doesn't change the build", other)
			}
			sim2, w := newWarden(t, other, proto.WeaponImbue_RockbiterWeapon)
			a, b := c.measure(t, sim2, w), c.measure(t, sim, base)
			got := a - b
			if c.ratio {
				got = a / b
			}
			near(t, "the talent's side against the base build's", got, c.want, c.tolerance)
		})
	}
}

// TestForeverWardenRockbiterThreat checks Rockbiter Weapon's threat under Spirit Weapons (client
// 70753): every landed melee hit with the weapon, white or Stormstrike, adds Rockbiter Weapon
// Proc's 72 threat for each second of the weapon's speed (20871, cast off the passive 16313 on
// melee auto and special hits), and the whole hit's threat is 0.7 x 1.86 times. With Windfury
// Weapon the weapon adds nothing and the threat is 0.7 times; without Spirit Weapons it's 1.
func TestForeverWardenRockbiterThreat(t *testing.T) {
	_, w := newWarden(t, wardenTalents, proto.WeaponImbue_RockbiterWeapon)
	speed := w.AutoAttacks.MH().SwingSpeed
	near(t, "Deathbringer's speed", speed, 2.9, 1e-9)
	white := w.AutoAttacks.MHAuto()
	near(t, "a white hit's threat for 100 damage", white.ThreatFromDamage(core.OutcomeHit, 100), (100+72*speed)*1.302, 1e-6)
	near(t, "Stormstrike's threat for 100 damage", w.Stormstrike.ThreatFromDamage(core.OutcomeHit, 100), (100+72*speed)*1.302, 1e-6)
	near(t, "a dodged white swing's threat", white.ThreatFromDamage(core.OutcomeDodge, 0), 0, 1e-9)
	// Earth Shock makes twice its damage in threat (Classic's, the client's "high amount of threat").
	near(t, "Earth Shock's threat for 100 damage", w.EarthShock[7].ThreatFromDamage(core.OutcomeHit, 100), 200*1.302, 1e-6)
	near(t, "an orb's threat for 100 damage", w.LightningShieldProcs[7].ThreatFromDamage(core.OutcomeHit, 100), 100*1.302, 1e-6)

	_, wf := newWarden(t, wardenTalents, proto.WeaponImbue_WindfuryWeapon)
	near(t, "a white hit's threat with Windfury Weapon", wf.AutoAttacks.MHAuto().ThreatFromDamage(core.OutcomeHit, 100), 70, 1e-6)
	near(t, "Stormstrike's threat with Windfury Weapon", wf.Stormstrike.ThreatFromDamage(core.OutcomeHit, 100), 70, 1e-6)

	_, plain := newWarden(t, withTalent(wardenTalents, treeEnhancement, spiritWeapons, 0), proto.WeaponImbue_RockbiterWeapon)
	near(t, "a white hit's threat with Rockbiter and no Spirit Weapons", plain.AutoAttacks.MHAuto().ThreatFromDamage(core.OutcomeHit, 100), 100+72*speed, 1e-6)
}

// TestForeverWardenAbilities checks the tank abilities' client rows (70753): Earth Shock rank 7 (10414,
// 450 mana, the shocks' 6 sec shared cooldown less Reverberation, the 1.5 sec GCD), Stormstrike (17364,
// 8% of base mana; 6 sec by the November 2 notes, the client's 8 sec before), Lightning Shield rank 7 (10432, 370 mana, 3 orbs, 10 min, an orb at most
// every 3.5 sec at whoever hits us), and Stoneskin Totem rank 6 (10408, 210 mana, 5 min, 30 off each
// physical hit before armor, 36 with Guardian Totems 2/2).
func TestForeverWardenAbilities(t *testing.T) {
	sim, w := newWarden(t, wardenTalents, proto.WeaponImbue_RockbiterWeapon)
	boss := sim.Encounter.TargetUnits[0]

	es := w.EarthShock[7]
	if got := es.DefaultCast.GCD; got != core.GCDDefault {
		t.Errorf("Earth Shock's GCD is %s", got)
	}
	castNow(t, sim, w, es)
	if got := es.CD.TimeToReady(sim); got != 5*time.Second {
		t.Errorf("Earth Shock is ready in %s with Reverberation 5/5, want 5s", got)
	}
	if fs := w.FlameShock[len(w.FlameShock)-1]; fs.CanCast(sim, w.CurrentTarget) {
		t.Error("Flame Shock casts right after Earth Shock, want the shared shock cooldown")
	}

	castNow(t, sim, w, w.Stormstrike)
	if got := w.Stormstrike.CD.TimeToReady(sim); got != 6*time.Second {
		t.Errorf("Stormstrike is ready in %s, want 6s (the November 2 notes)", got)
	}
	near(t, "Stormstrike's cost", w.Stormstrike.Cost.GetCurrentCost(), 0.08*w.BaseMana, 1e-9)

	// Lightning Shield: the boss lands a hit every second. Its hits fire an orb at most every 3.5
	// sec, each one spends a charge, and the third ends the shield.
	castNow(t, sim, w, w.LightningShield[7])
	shield := w.LightningShieldAuras[7]
	if shield.Duration != 10*time.Minute || shield.GetStacks() != 3 {
		t.Errorf("Lightning Shield lasts %s with %d orbs, want 10m0s and 3", shield.Duration, shield.GetStacks())
	}
	var orbTimes []time.Duration
	for orbs := int32(2); orbs >= 0; orbs-- {
		for shield.IsActive() && shield.GetStacks() > orbs && sim.CurrentTime < time.Minute {
			advance(sim, sim.CurrentTime+time.Second)
			boss.AutoAttacks.MHAuto().CalcAndDealDamage(sim, &w.Unit, 100, func(_ *core.Simulation, r *core.SpellResult, _ *core.AttackTable) {
				r.Outcome = core.OutcomeHit
			})
		}
		orbTimes = append(orbTimes, sim.CurrentTime)
	}
	for i := 1; i < len(orbTimes); i++ {
		if gap := orbTimes[i] - orbTimes[i-1]; gap < 3500*time.Millisecond {
			t.Errorf("orbs %s apart, want at least 3.5s", gap)
		}
	}
	if shield.IsActive() {
		t.Errorf("Lightning Shield still up with %d orbs after 3 fired", shield.GetStacks())
	}

	// Stoneskin: 36 off a 1,000 damage hit before armor, so (1000 - 36) x (1 - armor's share).
	castNow(t, sim, w, w.StoneskinTotem[6])
	near(t, "Stoneskin Totem's cost", w.StoneskinTotem[6].Cost.GetCurrentCost(), 210, 1e-9)
	swing := boss.AutoAttacks.MHAuto()
	hit := swing.CalcDamage(sim, &w.Unit, 1000, swing.OutcomeAlwaysHit)
	armor := w.Armor()
	reduction := armor / (armor + 400 + 85*63)
	near(t, "a 1,000 damage hit under Stoneskin", hit.Damage, (1000-36)*(1-reduction), 1e-6)
	if got := w.GetAura("Stoneskin").Duration; got != 5*time.Minute {
		t.Errorf("Stoneskin lasts %s, want 5m0s", got)
	}
}

// TestForeverWardenImprovedStormstrikeReset checks the Warden's half of Improved Stormstrike: at 2/2
// every dodge or parry of the boss's swing resets Stormstrike's cooldown (a blocked hit doesn't).
func TestForeverWardenImprovedStormstrikeReset(t *testing.T) {
	for _, c := range []struct {
		outcome core.HitOutcome
		resets  bool
	}{{core.OutcomeDodge, true}, {core.OutcomeParry, true}, {core.OutcomeBlock, false}, {core.OutcomeHit, false}} {
		sim, w := newWarden(t, wardenTalents, proto.WeaponImbue_RockbiterWeapon)
		boss := sim.Encounter.TargetUnits[0]
		w.Stormstrike.CD.Use(sim)
		boss.AutoAttacks.MHAuto().CalcAndDealDamage(sim, &w.Unit, 100, func(_ *core.Simulation, r *core.SpellResult, _ *core.AttackTable) {
			r.Outcome = c.outcome
		})
		if got := w.Stormstrike.CD.IsReady(sim); got != c.resets {
			t.Errorf("after a %s Stormstrike's ready is %v, want %v", c.outcome, got, c.resets)
		}
	}
}

// TestForeverWardenTank runs the tank bench's Balanced Warden (tools/tankbench/req/shaman_warden_balanced.json)
// against its boss with a stand-in healer and checks that it tanks: the boss attacks it, it blocks with its
// shield and parries with Spirit Weapons, its rotation's abilities go out, Lightning Shield and Stoneskin stay
// up, Rockbiter's threat is on its swings, and the tank score replays every iteration.
func TestForeverWardenTank(t *testing.T) {
	data, err := os.ReadFile("../../../tools/tankbench/req/shaman_warden_balanced.json")
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	const iterations, seconds = 200, 120
	rsr.Raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	rsr.Encounter.Duration, rsr.Encounter.DurationVariation = seconds, 0
	rsr.SimOptions.Iterations = iterations
	rsr.SimOptions.RandomSeed = 7
	player := rsr.Raid.Parties[0].Players[0]
	player.HealingModel = &proto.HealingModel{Hps: 900, CadenceSeconds: 2, CadenceVariation: 1}
	player.TankScoring = &proto.TankScoring{}
	res := core.RunRaidSimConcurrent(rsr)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	m := res.RaidMetrics.Parties[0].Players[0]

	sheet := m.TankSheet
	if sheet == nil {
		t.Fatal("no tank sheet: the boss didn't attack the Warden")
	}
	if sheet.Block < 0.05 || sheet.Parry < 0.05 || sheet.BlockValue < 50 {
		t.Errorf("block %.3f, parry %.3f and block value %.0f, want a shield's block and Spirit Weapons' parry", sheet.Block, sheet.Parry, sheet.BlockValue)
	}

	casts := map[int32]float64{}
	threat := map[int32]float64{}
	damage := map[int32]float64{}
	for _, a := range m.Actions {
		for _, tgt := range a.Targets {
			casts[a.Id.GetSpellId()] += float64(tgt.Casts) / iterations
			if a.Id.GetOtherId() == proto.OtherAction_OtherActionAttack && a.Id.GetTag() == 1 {
				threat[0] += tgt.Threat
				damage[0] += tgt.Damage
			}
		}
	}
	for id, name := range map[int32]string{17364: "Stormstrike", 10414: "Earth Shock"} {
		if casts[id] < 5 {
			t.Errorf("%s cast %.1f times a fight, want it in the rotation", name, casts[id])
		}
	}
	// Each landed white hit makes its damage and 72 x 2.9 threat, times 1.302.
	if damage[0] == 0 || threat[0] < damage[0]*1.302*1.1 {
		t.Errorf("the white swings made %.0f threat from %.0f damage, want Rockbiter's threat on top of 1.302 x the damage", threat[0], damage[0])
	}

	uptime := map[int32]float64{}
	for _, a := range m.Auras {
		uptime[a.Id.GetSpellId()] = a.UptimeSecondsAvg / seconds
	}
	for id, name := range map[int32]string{10405: "Stoneskin", 10432: "Lightning Shield"} {
		if uptime[id] < 0.8 {
			t.Errorf("%s up %.0f%% of the fight, want it kept up", name, 100*uptime[id])
		}
	}
	if m.Threat.Avg <= 0 || math.IsNaN(m.Dtps.Avg) || m.Dtps.Avg <= 0 {
		t.Errorf("TPS %.0f and DTPS %.0f, want both above 0", m.Threat.Avg, m.Dtps.Avg)
	}
	if got := len(m.TankScore.GetSurviveHps()); got != iterations {
		t.Errorf("the tank score replayed %d iterations, want %d", got, iterations)
	}
}
