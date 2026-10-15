package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// normalMobRequest is a level 30 orc shaman with a Whirlwind Axe, white swings only,
// soloing the given mobs from the front until they're all dead.
func normalMobRequest(mobs []*proto.Target, inFront bool, iterations int32) *proto.RaidSimRequest {
	player := newOrcShaman(30, "", &proto.EnhancementShaman_Options{})
	for range proto.ItemSlot_ItemSlotMainHand {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
	}
	player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: whirlwind})
	player.InFrontOfTarget = inFront
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	return &proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Duration: 300, UseHealth: true, Targets: mobs},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1, Iterations: iterations},
	}
}

// TestNormalMobStats checks the average normal mob table against a few mobs it's made from.
// A level 30 mob's health sits between Young Stranglethorn Tiger's (1050 at 30, health
// multiplier 1.1) and the plain level 30 warrior class mob's 955 in the 1.12 database, and
// it hits for about what the 30 to 31 mobs there do (Young Stranglethorn Tiger: 29 to 36
// plus 15 attack power on a 1.3 sec swing).
func TestNormalMobStats(t *testing.T) {
	for _, c := range []struct {
		level                int32
		health, armor        [2]float64
		minDamage, maxDamage [2]float64
	}{
		{1, [2]float64{40, 45}, [2]float64{10, 20}, [2]float64{0.5, 1.5}, [2]float64{1, 2}},
		{30, [2]float64{955, 1100}, [2]float64{1150, 1350}, [2]float64{40, 55}, [2]float64{50, 65}},
		{60, [2]float64{3600, 4300}, [2]float64{3400, 3900}, [2]float64{105, 135}, [2]float64{130, 165}},
	} {
		mob := core.NormalMobStatsAt(c.level)
		for _, check := range []struct {
			name string
			got  float64
			want [2]float64
		}{
			{"health", mob.Health, c.health}, {"armor", mob.Armor, c.armor},
			{"min damage", mob.MinDamage, c.minDamage}, {"max damage", mob.MaxDamage, c.maxDamage},
		} {
			if check.got < check.want[0] || check.got > check.want[1] {
				t.Errorf("level %d %s: %.1f, want %.0f to %.0f", c.level, check.name, check.got, check.want[0], check.want[1])
			}
		}
		share := 0.0
		for _, s := range mob.MobTypes {
			share += s
		}
		if math.Abs(share-1) > 0.005 {
			t.Errorf("level %d creature types add up to %.3f", c.level, share)
		}
	}
}

// TestNormalMobSwingOutcomes rolls a normal mob's white swing against us many times. The
// table is Classic's: 5% miss and 5% crit at our level, 0.2% each a level of difference
// (0.04% for each point of our defense over its weapon skill, both being 5 a level), and our
// dodge from the character sheet, 0.2% less a level up. We have no shield and no Spirit
// Weapons, so no block or parry. Crushing blows take a mob 3 levels above us.
//
// While we cast we can't dodge, parry or block, so a hit lands more often.
func TestNormalMobSwingOutcomes(t *testing.T) {
	for _, c := range []struct {
		name     string
		mobLevel int32
		casting  bool
		crush    float64
	}{
		{"a level 29 mob", 29, false, 0},
		{"a level 30 mob", 30, false, 0},
		{"a level 31 mob", 31, false, 0},
		{"a level 33 mob", 33, false, 0.15},
		{"a level 30 mob while we cast", 30, true, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim := core.NewSim(normalMobRequest([]*proto.Target{core.NormalMobTarget(c.mobLevel, proto.MobType_MobTypeBeast, 1)}, true, 1), simsignals.CreateSignals())
			sim.Reset()
			enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
			mob := sim.Encounter.TargetUnits[0]
			if c.casting {
				enh.Hardcast.Expires = core.NeverExpires
			}

			delta := 0.002 * float64(30-c.mobLevel)
			dodge := enh.GetStat(stats.Dodge)/100 + delta
			if c.casting {
				dodge = 0
			}
			want := map[string]float64{"miss": 0.05 + delta, "dodge": dodge, "parry": 0, "block": 0, "crit": 0.05 - delta, "crush": c.crush}

			const rolls = 200000
			swing := mob.AutoAttacks.MHAuto()
			got := map[string]int{}
			for range rolls {
				result := swing.CalcDamage(sim, &enh.Unit, 1, swing.OutcomeEnemyMeleeWhite)
				switch {
				case result.Outcome.Matches(core.OutcomeMiss):
					got["miss"]++
				case result.Outcome.Matches(core.OutcomeDodge):
					got["dodge"]++
				case result.Outcome.Matches(core.OutcomeParry):
					got["parry"]++
				case result.Outcome.Matches(core.OutcomeBlock):
					got["block"]++
				case result.Outcome.Matches(core.OutcomeCrit):
					got["crit"]++
				case result.Outcome.Matches(core.OutcomeCrush):
					got["crush"]++
				}
			}
			for outcome, w := range want {
				rate := float64(got[outcome]) / rolls
				if se := math.Sqrt(max(w*(1-w), 1e-6) / rolls); math.Abs(rate-w) > 4*se {
					t.Errorf("%s: %.4f, want %.4f +- %.4f", outcome, rate, w, 4*se)
				}
			}
		})
	}
}

// TestNormalMobHitDamage checks a normal hit from a level 30 mob: its raw damage, spread
// evenly between the table's minimum and maximum, less our armor at the mob's level:
// armor / (armor + 400 + 85 * 30).
func TestNormalMobHitDamage(t *testing.T) {
	sim := core.NewSim(normalMobRequest([]*proto.Target{core.NormalMobTarget(30, proto.MobType_MobTypeBeast, 1)}, true, 1), simsignals.CreateSignals())
	sim.Reset()
	enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
	mob := sim.Encounter.TargetUnits[0]
	stats30 := core.NormalMobStatsAt(30)

	armor := enh.Armor()
	if armor <= 0 {
		t.Fatalf("our armor is %.0f", armor)
	}
	reduction := armor / (armor + 400 + 85*30)
	lo, hi := stats30.MinDamage*(1-reduction), stats30.MaxDamage*(1-reduction)

	const rolls = 50000
	swing := mob.AutoAttacks.MHAuto()
	sum, n := 0.0, 0
	for range rolls {
		result := swing.CalcDamage(sim, &enh.Unit, swing.Unit.AutoAttacks.MH().EnemyWeaponDamage(sim, 0, swing.Unit.PseudoStats.DamageSpread), swing.OutcomeEnemyMeleeWhite)
		if result.Outcome != core.OutcomeHit {
			continue
		}
		if result.Damage < lo-1e-9 || result.Damage > hi+1e-9 {
			t.Fatalf("a hit did %.2f, want %.2f to %.2f", result.Damage, lo, hi)
		}
		sum += result.Damage
		n++
	}
	if mean := sum / float64(n); math.Abs(mean-(lo+hi)/2) > 0.01*(lo+hi)/2 {
		t.Errorf("hits did %.2f on average, want %.2f", mean, (lo+hi)/2)
	}
}

// TestSoloNormalMob fights one level 30 mob to the death with white swings and checks the
// damage we take against the mob's swing table: its average hit after armor, times the
// share of swings that land (crits counting double), divided by its swing time.
//
// From behind the mob can't parry us, so it never gets parry haste. From the front it can
// (5% at our level), and each parry brings its next swing sooner, so we take a little more.
func TestSoloNormalMob(t *testing.T) {
	mob := core.NormalMobStatsAt(30)
	behind := core.RunRaidSim(normalMobRequest([]*proto.Target{core.NormalMobTarget(30, proto.MobType_MobTypeBeast, 1)}, false, 2000))
	front := core.RunRaidSim(normalMobRequest([]*proto.Target{core.NormalMobTarget(30, proto.MobType_MobTypeBeast, 1)}, true, 2000))
	for _, r := range []*proto.RaidSimResult{behind, front} {
		if r.Error != nil {
			t.Fatal(r.Error.Message)
		}
	}

	// About 8 white swings of 120 on a 3.6 sec axe.
	if d := behind.AvgIterationDuration; d < 25 || d > 40 {
		t.Errorf("killing a level 30 mob took %.1f sec, want 25 to 40", d)
	}

	// Our armor and dodge, from a single sim of the same character.
	sim := core.NewSim(normalMobRequest([]*proto.Target{core.NormalMobTarget(30, proto.MobType_MobTypeBeast, 1)}, false, 1), simsignals.CreateSignals())
	sim.Reset()
	enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
	armor := enh.Armor()
	avgHit := (mob.MinDamage + mob.MaxDamage) / 2 * (1 - armor/(armor+400+85*30))
	landed := 1 - 0.05 - enh.GetStat(stats.Dodge)/100
	wantDTPS := avgHit * (landed + 0.05) / mob.SwingSpeed

	player := behind.RaidMetrics.Parties[0].Players[0]
	if got := player.Dtps.Avg; math.Abs(got-wantDTPS) > 0.03*wantDTPS {
		t.Errorf("from behind we took %.2f damage a second, want %.2f", got, wantDTPS)
	}
	// Naked, with 815 health and only white swings, we take about 790 in the 32 sec it
	// takes, so we die some of the time but not most of it.
	if cod := player.ChanceOfDeath; cod < 0.1 || cod > 0.6 {
		t.Errorf("a naked level 30 died to one level 30 mob %.1f%% of the time, want 10 to 60%%", 100*cod)
	}
	if got, back := front.RaidMetrics.Parties[0].Players[0].Dtps.Avg, player.Dtps.Avg; got <= back {
		t.Errorf("from the front we took %.2f damage a second, from behind %.2f: parry haste should add some", got, back)
	}
}

// TestNormalMobPackKillsUs pulls three level 32 mobs onto a naked level 30 with only white
// swings. It can't kill them all before they kill it, and the sim now counts that death
// without a healing model.
func TestNormalMobPackKillsUs(t *testing.T) {
	mobs := []*proto.Target{
		core.NormalMobTarget(32, proto.MobType_MobTypeBeast, 1),
		core.NormalMobTarget(32, proto.MobType_MobTypeBeast, 1),
		core.NormalMobTarget(32, proto.MobType_MobTypeBeast, 1),
	}
	result := core.RunRaidSim(normalMobRequest(mobs, true, 200))
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	if cod := result.RaidMetrics.Parties[0].Players[0].ChanceOfDeath; cod < 0.95 {
		t.Errorf("we died %.1f%% of the time to three level 32 mobs, want nearly always", 100*cod)
	}
}
