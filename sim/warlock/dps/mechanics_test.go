package dps

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
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

	if got := last(with.ShadowBolt).ImpactDamageMultiplierAdditive - last(without.ShadowBolt).ImpactDamageMultiplierAdditive; got < 0.0999 || got > 0.1001 {
		t.Errorf("Shadow Mastery 5/5 adds %.3f to Shadow Bolt, want 0.10", got)
	}
	if got := last(with.Corruption).PeriodicDamageMultiplierAdditive - last(without.Corruption).PeriodicDamageMultiplierAdditive; got < 0.0999 || got > 0.1001 {
		t.Errorf("Shadow Mastery 5/5 adds %.3f to Corruption's ticks, want 0.10", got)
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
