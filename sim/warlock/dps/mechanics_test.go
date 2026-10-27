package dps

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/warlock"
)

func newWarlockSim(t *testing.T, talents string, options *proto.WarlockOptions) (*core.Simulation, *DpsWarlock) {
	player := &proto.Player{
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarlock,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec:          &proto.Player_Warlock{Warlock: &proto.Warlock{Options: options}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 120, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*DpsWarlock)
}

func castNow(t *testing.T, sim *core.Simulation, lock *DpsWarlock, spell *core.Spell) {
	lock.SetGCDTimer(sim, sim.CurrentTime)
	lock.Hardcast.Expires = sim.CurrentTime
	if !spell.Cast(sim, lock.CurrentTarget) {
		t.Fatalf("at %s: %s didn't cast", sim.CurrentTime, spell.ActionID)
	}
}

func at(sim *core.Simulation, seconds float64, fn func(sim *core.Simulation)) {
	sim.AddPendingAction(&core.PendingAction{NextActionAt: core.DurationFromSeconds(seconds), OnAction: fn})
}

func runSim(sim *core.Simulation) {
	for !sim.Step() {
	}
}

// TestWarlockBanesAndCurses checks that a bane and a curse share the target under Forever, and
// that a second bane takes the first one's place. Spell hit is maxed so nothing misses.
func TestWarlockBanesAndCurses(t *testing.T) {
	sim, lock := newWarlockSim(t, "", &proto.WarlockOptions{})
	target := lock.CurrentTarget
	agony := lock.BaneOfAgony[len(lock.BaneOfAgony)-1]

	at(sim, 0.5, func(sim *core.Simulation) {
		lock.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
		castNow(t, sim, lock, lock.CurseOfElements)
	})
	at(sim, 3, func(sim *core.Simulation) { castNow(t, sim, lock, agony) })
	at(sim, 6, func(sim *core.Simulation) {
		if !lock.CurseOfElementsAuras.Get(target).IsActive() || !agony.Dot(target).IsActive() {
			t.Errorf("Curse of the Elements up %v, Bane of Agony up %v, want both",
				lock.CurseOfElementsAuras.Get(target).IsActive(), agony.Dot(target).IsActive())
		}
		castNow(t, sim, lock, lock.BaneOfDoom)
	})
	at(sim, 9, func(sim *core.Simulation) {
		if agony.Dot(target).IsActive() || !lock.BaneOfDoom.Dot(target).IsActive() {
			t.Errorf("Bane of Agony up %v, Bane of Doom up %v, want only Doom",
				agony.Dot(target).IsActive(), lock.BaneOfDoom.Dot(target).IsActive())
		}
		if !lock.CurseOfElementsAuras.Get(target).IsActive() {
			t.Errorf("Bane of Doom took Curse of the Elements off")
		}
	})
	runSim(sim)
}

// TestWarlockDemonicPact checks the Imp sacrificed before the pull: its 15% Shadow damage stays
// with the Succubus out only with Demonic Pact, and goes with any demon out without it. When
// it stays it is up from the reset, before any pre-pull cast, not from the pull.
func TestWarlockDemonicPact(t *testing.T) {
	const withPact = "-0305003131221001351"
	const withoutPact = "-030500313122100135"
	for _, c := range []struct {
		talents string
		summon  proto.WarlockOptions_Summon
		want    bool
	}{
		{withPact, proto.WarlockOptions_Succubus, true},
		{withoutPact, proto.WarlockOptions_Succubus, false},
		{withoutPact, proto.WarlockOptions_NoSummon, true},
	} {
		sim, lock := newWarlockSim(t, c.talents, &proto.WarlockOptions{Summon: c.summon, SacrificedPet: proto.WarlockOptions_Imp})
		if upAtReset := lock.DemonicSacrificeAuras[lock.Imp].IsActive(); upAtReset != c.want {
			t.Errorf("%s with demon %s: Burning Shadow up before the pull %v, want %v", c.talents, c.summon, upAtReset, c.want)
		}
		var up bool
		at(sim, 1, func(sim *core.Simulation) {
			up = lock.DemonicSacrificeAuras[lock.Imp].IsActive()
		})
		runSim(sim)
		if up != c.want {
			t.Errorf("%s with demon %s: Burning Shadow up %v, want %v", c.talents, c.summon, up, c.want)
		}
	}
}

// TestWarlockSiphonLifeCrits checks that Siphon Life's ticks can crit under Forever, like the
// warlock's other damage over time spells (forever-bugs #303, fixed for the client after 70338).
func TestWarlockSiphonLifeCrits(t *testing.T) {
	sim, lock := newWarlockSim(t, "00000000000001", &proto.WarlockOptions{})
	siphonLife := lock.SiphonLife[len(lock.SiphonLife)-1]

	at(sim, 0.5, func(sim *core.Simulation) {
		lock.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
		lock.AddStatDynamic(sim, stats.SpellCrit, 100*core.SpellCritRatingPerCritChance)
		castNow(t, sim, lock, siphonLife)
	})
	runSim(sim)

	metrics := siphonLife.SpellMetrics[lock.CurrentTarget.UnitIndex]
	if metrics.CritTicks == 0 || metrics.Ticks != 0 {
		t.Errorf("Siphon Life at 100%% crit: %d crit ticks and %d plain ticks, want only crits", metrics.CritTicks, metrics.Ticks)
	}
}

// TestWarlockCurseOfRecklessnessLands checks that the warlock's own Curse of Recklessness goes on
// the target and stays for its 2 min (the Curse of Recklessness duty presets keep it up).
func TestWarlockCurseOfRecklessnessLands(t *testing.T) {
	sim, lock := newWarlockSim(t, "", &proto.WarlockOptions{})
	target := lock.CurrentTarget

	at(sim, 0.5, func(sim *core.Simulation) {
		lock.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
		castNow(t, sim, lock, lock.CurseOfRecklessness)
	})
	at(sim, 60, func(sim *core.Simulation) {
		if !lock.CurseOfRecklessnessAuras.Get(target).IsActive() {
			t.Errorf("Curse of Recklessness not up 59.5 sec after the cast")
		}
	})
	runSim(sim)
}

// TestWarlockShadowMastery checks Shadow Mastery at 5/5: 10% on Shadow Bolt's damage and on
// Corruption's ticks (2% a point since client 1.60.1.70425, 1% before).
func TestWarlockShadowMastery(t *testing.T) {
	_, without := newWarlockSim(t, "", &proto.WarlockOptions{})
	_, with := newWarlockSim(t, "0000000000000005", &proto.WarlockOptions{})
	last := func(spells []*core.Spell) *core.Spell { return spells[len(spells)-1] }

	if got := last(with.ShadowBolt).ImpactDamageMultiplierAdditive / last(without.ShadowBolt).ImpactDamageMultiplierAdditive; got < 1.0999 || got > 1.1001 {
		t.Errorf("Shadow Mastery 5/5 makes Shadow Bolt x%.3f, want x1.10", got)
	}
	if got := last(with.Corruption).PeriodicDamageMultiplierAdditive / last(without.Corruption).PeriodicDamageMultiplierAdditive; got < 1.0999 || got > 1.1001 {
		t.Errorf("Shadow Mastery 5/5 makes Corruption's ticks x%.3f, want x1.10", got)
	}
}

// TestWarlockPushbackProtection checks Intensity and Fel Concentration at 3/3: a 70% chance to
// ignore a hit's pushback on the spells in their client masks, Incinerate not among Intensity's.
func TestWarlockPushbackProtection(t *testing.T) {
	_, intensity := newWarlockSim(t, "--0000000030000001", &proto.WarlockOptions{})
	_, felConcentration := newWarlockSim(t, "00000003", &proto.WarlockOptions{})
	last := func(spells []*core.Spell) *core.Spell { return spells[len(spells)-1] }

	for _, c := range []struct {
		name  string
		spell *core.Spell
		want  float64
	}{
		{"Intensity: Shadow Bolt", last(intensity.ShadowBolt), 0.70},
		{"Intensity: Immolate", last(intensity.Immolate), 0.70},
		{"Intensity: Incinerate", last(intensity.Incinerate), 0},
		{"Fel Concentration: Drain Life", last(felConcentration.DrainLife), 0.70},
		{"Fel Concentration: Shadow Bolt", last(felConcentration.ShadowBolt), 0},
	} {
		if got := c.spell.PushbackReduction; got < c.want-1e-9 || got > c.want+1e-9 {
			t.Errorf("%s: pushback reduction %.2f, want %.2f", c.name, got, c.want)
		}
	}
}

// TestWarlockDemonCrit checks that a demon's melee and spell crit are its master's spell crit and
// nothing of its own (forever-bugs #331, Blizzard's intent), also after the warlock gains crit.
func TestWarlockDemonCrit(t *testing.T) {
	for _, summon := range []proto.WarlockOptions_Summon{proto.WarlockOptions_Succubus, proto.WarlockOptions_Imp, proto.WarlockOptions_Felhunter} {
		sim, lock := newWarlockSim(t, "", &proto.WarlockOptions{Summon: summon})
		check := func(when string) {
			want := lock.GetStat(stats.SpellCrit)
			for _, stat := range []stats.Stat{stats.MeleeCrit, stats.SpellCrit} {
				if got := lock.ActivePet.GetStat(stat); got != want {
					t.Errorf("%s %s: stat %d is %.3f, want the warlock's spell crit %.3f", summon, when, stat, got, want)
				}
			}
		}
		if lock.GetStat(stats.SpellCrit) <= 0 {
			t.Fatalf("the warlock has no spell crit to pass on")
		}
		check("at the pull")
		at(sim, 1, func(sim *core.Simulation) {
			lock.AddStatDynamic(sim, stats.SpellCrit, 5*core.SpellCritRatingPerCritChance)
		})
		at(sim, 2, func(sim *core.Simulation) { check("after +5% spell crit") })
		runSim(sim)
	}
}

// TestWarlockDemonStatInheritance checks Warlock Pet Scaling's shares (wowsims/forever's): 0.17
// attack power and 0.10 spell damage a point of the warlock's spell damage, its spell hit whole and
// 8/17 of it as melee hit, 7.5 health a point of its Stamina, 35% of its armor and MP5 from its
// Intellect (0.30 for the Imp, 0.10 for the others), also after the warlock gains stats. The Imp's
// mana is 576 + 5 a point of its Intellect, and no demon gets regen from its Spirit.
func TestWarlockDemonStatInheritance(t *testing.T) {
	near := func(a, b float64) bool { return a-b < 1e-6 && b-a < 1e-6 }
	for _, c := range []struct {
		summon       proto.WarlockOptions_Summon
		mp5PerOwnInt float64
	}{{proto.WarlockOptions_Succubus, 0.10}, {proto.WarlockOptions_Imp, 0.30}} {
		sim, lock := newWarlockSim(t, "", &proto.WarlockOptions{Summon: c.summon})
		pet := lock.ActivePet
		before := pet.GetStats()
		owner := lock.GetStats()

		if got, want := pet.GetStat(stats.MP5), owner[stats.Intellect]*c.mp5PerOwnInt; !near(got, want) {
			t.Errorf("%s: MP5 %.2f, want %.2f", c.summon, got, want)
		}
		if got, want := pet.GetStat(stats.MeleeHit)/core.MeleeHitRatingPerHitChance, owner[stats.SpellHit]/core.SpellHitRatingPerHitChance*8/17; !near(got, want) {
			t.Errorf("%s: melee hit %.3f%%, want %.3f%%", c.summon, got, want)
		}
		if got := pet.ManaRegenPerSecondWhileNotCasting(); !near(got, pet.GetStat(stats.MP5)/5) {
			t.Errorf("%s: regen %.3f a second out of casting, want only its MP5's %.3f", c.summon, got, pet.GetStat(stats.MP5)/5)
		}
		if c.summon == proto.WarlockOptions_Imp {
			if got, want := pet.MaxMana(), 576+5*pet.GetStat(stats.Intellect); !near(got, want) {
				t.Errorf("Imp: mana %.1f, want %.1f", got, want)
			}
		}

		at(sim, 1, func(sim *core.Simulation) {
			lock.AddStatsDynamic(sim, stats.Stats{
				stats.ShadowPower: 100,
				stats.SpellHit:    2 * core.SpellHitRatingPerHitChance,
				stats.Stamina:     10,
				stats.Armor:       200,
			})
		})
		at(sim, 2, func(sim *core.Simulation) {
			after := pet.GetStats()
			for _, s := range []struct {
				name string
				stat stats.Stat
				want float64
			}{
				{"attack power", stats.AttackPower, 17},
				{"spell damage", stats.SpellDamage, 10},
				{"spell hit", stats.SpellHit, 2 * core.SpellHitRatingPerHitChance},
				{"melee hit", stats.MeleeHit, 2 * 8.0 / 17 * core.MeleeHitRatingPerHitChance},
				{"armor", stats.Armor, 70},
			} {
				if got := after[s.stat] - before[s.stat]; !near(got, s.want) {
					t.Errorf("%s: +100 Shadow power, +2%% hit, +200 armor: %s +%.3f, want +%.3f", c.summon, s.name, got, s.want)
				}
			}
			// The warlock's Stamina is worth 10 health to it, the demon's share 7.5 a point.
			if got, want := after[stats.Health]-before[stats.Health], 7.5*(lock.GetStat(stats.Stamina)-owner[stats.Stamina]); !near(got, want) {
				t.Errorf("%s: +10 Stamina: health +%.2f, want +%.2f", c.summon, got, want)
			}
		})
		runSim(sim)
	}
}

// TestWarlockDemonSchoolPowerInheritance checks that a demon's share of its master's spell damage is
// worked out from the master's totals, as wowsims/forever does (af11e5e9ff): it takes the larger of
// Fire and Shadow power, so +100 Shadow power with a Firestone's 21 Fire power up is worth 79, and a
// Shadow power buff that falls off leaves the demon as it was.
func TestWarlockDemonSchoolPowerInheritance(t *testing.T) {
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-6 }
	for _, c := range []struct {
		imbue proto.WarlockOptions_WeaponImbue
		gain  float64
	}{{proto.WarlockOptions_NoWeaponImbue, 100}, {proto.WarlockOptions_Firestone, 79}} {
		sim, lock := newWarlockSim(t, "", &proto.WarlockOptions{Summon: proto.WarlockOptions_Succubus, WeaponImbue: c.imbue})
		pet := lock.ActivePet
		before := pet.GetStats()

		at(sim, 1, func(sim *core.Simulation) { lock.AddStatDynamic(sim, stats.ShadowPower, 100) })
		at(sim, 2, func(sim *core.Simulation) {
			after := pet.GetStats()
			if got, want := after[stats.AttackPower]-before[stats.AttackPower], 0.17*c.gain; !near(got, want) {
				t.Errorf("%s: +100 Shadow power: attack power +%.3f, want +%.3f", c.imbue, got, want)
			}
			if got, want := after[stats.SpellDamage]-before[stats.SpellDamage], 0.10*c.gain; !near(got, want) {
				t.Errorf("%s: +100 Shadow power: spell damage +%.3f, want +%.3f", c.imbue, got, want)
			}
			lock.AddStatDynamic(sim, stats.ShadowPower, -100)
		})
		at(sim, 3, func(sim *core.Simulation) {
			after := pet.GetStats()
			for _, stat := range []stats.Stat{stats.AttackPower, stats.SpellDamage} {
				if !near(after[stat], before[stat]) {
					t.Errorf("%s: after the Shadow power falls off, stat %d is %.3f, want %.3f", c.imbue, stat, after[stat], before[stat])
				}
			}
		})
		runSim(sim)
	}
}

// TestWarlockPercentModsMultiply checks that the percent damage rows on a spell are each their own
// factor (wowsims/forever fef2bcaebd, #377): Aftermath 5/5 with Agonizing Flames 3/3 makes Immolate's
// hit 1.5 x 1.1 = 1.65 and its ticks 1.1, Improved Corruption 5/5 with Shadow Mastery 5/5 makes
// Corruption's ticks 1.1 x 1.1 = 1.21, and Cataclysm 3/3 is a 0.9 cost factor on Shadow Bolt.
func TestWarlockPercentModsMultiply(t *testing.T) {
	_, without := newWarlockSim(t, "", &proto.WarlockOptions{})
	_, destro := newWarlockSim(t, "--0000350003", &proto.WarlockOptions{})
	_, affli := newWarlockSim(t, "0050000000000005", &proto.WarlockOptions{})
	last := func(spells []*core.Spell) *core.Spell { return spells[len(spells)-1] }
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

	for _, c := range []struct {
		name      string
		got, want float64
	}{
		{"Immolate's hit", last(destro.Immolate).ImpactDamageMultiplierAdditive / last(without.Immolate).ImpactDamageMultiplierAdditive, 1.65},
		{"Immolate's ticks", last(destro.Immolate).PeriodicDamageMultiplierAdditive / last(without.Immolate).PeriodicDamageMultiplierAdditive, 1.1},
		{"Corruption's ticks", last(affli.Corruption).PeriodicDamageMultiplierAdditive / last(without.Corruption).PeriodicDamageMultiplierAdditive, 1.21},
		{"Shadow Bolt's cost", last(destro.ShadowBolt).Cost.Factor / last(without.ShadowBolt).Cost.Factor, 0.9},
		{"Rain of Fire's hit", last(destro.RainOfFire).ImpactDamageMultiplierAdditive / last(without.RainOfFire).ImpactDamageMultiplierAdditive, 1.1},
	} {
		if !near(c.got, c.want) {
			t.Errorf("%s: x%.4f, want x%.4f", c.name, c.got, c.want)
		}
	}
}

// TestWarlockConflagrateMiss checks that a Conflagrate that misses leaves Immolate burning and one
// that lands consumes it (wowsims/forever conflagrate.go).
func TestWarlockConflagrateMiss(t *testing.T) {
	for _, lands := range []bool{false, true} {
		sim, lock := newWarlockSim(t, "--00000000001", &proto.WarlockOptions{})
		immolate := last(lock.Immolate)
		at(sim, 0.5, func(sim *core.Simulation) {
			lock.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
			castNow(t, sim, lock, immolate)
		})
		at(sim, 3, func(sim *core.Simulation) {
			if !immolate.Dot(lock.CurrentTarget).IsActive() {
				t.Fatalf("Immolate not up before Conflagrate (mana %.0f)", lock.CurrentMana())
			}
			if !lands {
				lock.AddStatDynamic(sim, stats.SpellHit, -300*core.SpellHitRatingPerHitChance)
			}
			// Not castNow: Immolate's cast is over, and moving its end to now would complete it again.
			lock.SetGCDTimer(sim, sim.CurrentTime)
			if !last(lock.Conflagrate).Cast(sim, lock.CurrentTarget) {
				t.Fatalf("Conflagrate didn't cast")
			}
		})
		at(sim, 4, func(sim *core.Simulation) {
			if up := immolate.Dot(lock.CurrentTarget).IsActive(); up == lands {
				t.Errorf("Conflagrate landing %v: Immolate up %v, want %v", lands, up, !lands)
			}
		})
		runSim(sim)
	}
}

// TestWarlockDecimation checks Decimation 2/2 with the fight under 35% from the start: Shadow Bolt
// and Searing Pain hit for 6% more as a factor on the hit, and only a Searing Pain that lands gives
// the Soul Fire buff (wowsims/forever 13b1bd31c5).
func TestWarlockDecimation(t *testing.T) {
	for _, lands := range []bool{false, true} {
		player := &proto.Player{
			Race:          proto.Race_RaceOrc,
			Class:         proto.Class_ClassWarlock,
			Level:         60,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "-000000000002",
			Rotation:      &proto.APLRotation{},
			Spec:          &proto.Player_Warlock{Warlock: &proto.Warlock{Options: &proto.WarlockOptions{}}},
		}
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  &proto.Encounter{Duration: 60, ExecuteProportion_35: 1, Targets: []*proto.Target{core.NewDefaultTarget()}},
			SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()
		sim.PrePull()
		lock := sim.Raid.Parties[0].Players[0].(*DpsWarlock)

		at(sim, 0.5, func(sim *core.Simulation) {
			if got := last(lock.ShadowBolt).ImpactDamageMultiplierAdditive; math.Abs(got-1.06) > 1e-9 {
				t.Errorf("Shadow Bolt's hit under 35%%: x%.4f, want x1.06", got)
			}
			lock.AddStatDynamic(sim, stats.SpellHit, core.TernaryFloat64(lands, 20, -300)*core.SpellHitRatingPerHitChance)
			castNow(t, sim, lock, last(lock.SearingPain))
		})
		at(sim, 3, func(sim *core.Simulation) {
			if up := lock.DecimationAura.IsActive(); up != lands {
				t.Errorf("Searing Pain landing %v: Decimation up %v, want %v", lands, up, lands)
			}
		})
		runSim(sim)
	}
}

// TestWarlockSpellstone checks that Major Spellstone gives its 21 damage to Shadow spells only
// (1237165 effect 0: Misc 32 since client 1.60.1.70630, was 36, Fire | Shadow).
func TestWarlockSpellstone(t *testing.T) {
	_, without := newWarlockSim(t, "", &proto.WarlockOptions{})
	_, with := newWarlockSim(t, "", &proto.WarlockOptions{WeaponImbue: proto.WarlockOptions_Spellstone})
	for stat, want := range map[stats.Stat]float64{stats.FirePower: 0, stats.ShadowPower: 21} {
		if got := with.GetStat(stat) - without.GetStat(stat); got != want {
			t.Errorf("Major Spellstone: stat %d +%.0f, want +%.0f", stat, got, want)
		}
	}
}

func last(spells []*core.Spell) *core.Spell { return spells[len(spells)-1] }

// TestWarlockSummonSwap checks a summon cast mid-fight: the demon out stays, stunned for Summoning
// Disorientation's 6 sec from the start of the cast, until the 10 sec cast ends and the new demon
// replaces it (wowsims/forever summon.go, d7968fbd13).
func TestWarlockSummonSwap(t *testing.T) {
	sim, lock := newWarlockSim(t, "", &proto.WarlockOptions{Summon: proto.WarlockOptions_Succubus})
	summonImp := lock.GetSpell(core.ActionID{SpellID: 688})
	at(sim, 1, func(sim *core.Simulation) {
		castNow(t, sim, lock, summonImp)
	})
	check := func(when float64, wantPet *warlock.WarlockPet, wantStun bool) {
		at(sim, when, func(sim *core.Simulation) {
			if lock.ActivePet != wantPet || !wantPet.IsEnabled() {
				t.Errorf("at %.1f s: %s out (enabled %v), want %s", when, lock.ActivePet.Name, lock.ActivePet.IsEnabled(), wantPet.Name)
			}
			if stunned := lock.Succubus.SummonStunAura.IsActive(); stunned != wantStun {
				t.Errorf("at %.1f s: Succubus stunned %v, want %v", when, stunned, wantStun)
			}
		})
	}
	check(3, lock.Succubus, true)
	check(8, lock.Succubus, false)
	check(12, lock.Imp, false)
	at(sim, 12, func(sim *core.Simulation) {
		if lock.Succubus.IsEnabled() {
			t.Errorf("Succubus still out after the Imp's summon")
		}
	})
	runSim(sim)
}

// TestWarlockSummonCostFactors checks that Master Summoner's 20% a point and Fel Domination's 50%
// multiply on a summon's cost (wowsims/forever fef2bcaebd): 0.6 x 0.5 at 2/2.
func TestWarlockSummonCostFactors(t *testing.T) {
	sim, lock := newWarlockSim(t, "-0000000000201", &proto.WarlockOptions{Summon: proto.WarlockOptions_Imp})
	summon := lock.GetSpell(core.ActionID{SpellID: 712})
	if got := summon.Cost.Factor; math.Abs(got-0.6) > 1e-9 {
		t.Errorf("Master Summoner 2/2: cost x%.4f, want x0.6", got)
	}
	lock.GetAura("Fel Domination").Activate(sim)
	if got := summon.Cost.Factor; math.Abs(got-0.3) > 1e-9 {
		t.Errorf("Master Summoner 2/2 and Fel Domination: cost x%.4f, want x0.3", got)
	}
}

// TestWarlockDemonicBrandAlwaysHits checks that Demonic Brand's damage never misses: both rows
// (1293697, 1293698) carry ATTR_EX_3_ALWAYS_HIT.
func TestWarlockDemonicBrandAlwaysHits(t *testing.T) {
	sim, lock := newWarlockSim(t, "-00000000000003", &proto.WarlockOptions{Summon: proto.WarlockOptions_Imp})
	brand := lock.Imp.GetSpell(core.ActionID{SpellID: 1293695})
	target := lock.CurrentTarget
	at(sim, 1, func(sim *core.Simulation) {
		lock.Imp.AddStatDynamic(sim, stats.SpellHit, -300*core.SpellHitRatingPerHitChance)
		for range 100 {
			brand.Cast(sim, target)
		}
		if m := brand.SpellMetrics[target.UnitIndex]; m.Misses != 0 || m.Hits+m.Crits != 100 {
			t.Errorf("Demonic Brand at -300%% hit: %d misses, %d hits and %d crits in 100", m.Misses, m.Hits, m.Crits)
		}
	})
	runSim(sim)
}

// TestWarlockImpFireboltTravel checks that the Imp's Firebolt flies at its rows' Speed 16 from the
// warlock's distance, up to its 30 yd range, and has no cooldown (wowsims/forever a21497c37b).
func TestWarlockImpFireboltTravel(t *testing.T) {
	_, lock := newWarlockSim(t, "", &proto.WarlockOptions{Summon: proto.WarlockOptions_Imp})
	firebolt := lock.Imp.GetSpell(core.ActionID{SpellID: 11763})
	if firebolt.MissileSpeed != 16 || firebolt.CD.Timer != nil {
		t.Errorf("Firebolt: speed %.0f, cooldown %v, want 16 and none", firebolt.MissileSpeed, firebolt.CD.Timer != nil)
	}
	if got, want := lock.Imp.StartDistanceFromTarget, min(lock.StartDistanceFromTarget, 30); got != want {
		t.Errorf("Imp's distance %.0f, want %.0f", got, want)
	}
}
