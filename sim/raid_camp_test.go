package sim

// The raid sim's camp buffs and Sunder ramp (docs/raid_sim.md, M1's engine pieces). Needs the item
// database: go test -tags with_db.

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

func finalStats(t *testing.T, rsr *proto.RaidSimRequest) stats.Stats {
	t.Helper()
	res := core.ComputeStats(&proto.ComputeStatsRequest{Raid: rsr.Raid, Encounter: rsr.Encounter})
	if res.ErrorResult != "" {
		t.Fatal(res.ErrorResult)
	}
	return stats.FromFloatArray(res.RaidStats.Parties[0].Players[0].FinalStats.Stats)
}

// Each camp buff adds its stat to a player without its class buff, and nothing to one with it.
func TestRaidForeverCampBuffs(t *testing.T) {
	cases := []struct {
		name  string
		camp  func(*proto.CampBuffs)
		class func(*proto.RaidSimRequest)
		stat  stats.Stat
	}{
		{"Fish Bowl", func(c *proto.CampBuffs) { c.FishBowl = true },
			func(r *proto.RaidSimRequest) { r.Raid.Parties[0].Players[0].Buffs.BlessingOfKings = true }, stats.Agility},
		{"Lodestone", func(c *proto.CampBuffs) { c.Lodestone = true },
			func(r *proto.RaidSimRequest) {
				r.Raid.Parties[0].Players[0].Buffs.BlessingOfMight = proto.TristateEffect_TristateEffectRegular
			}, stats.AttackPower},
		{"Sharpening Wheel", func(c *proto.CampBuffs) { c.SharpeningWheel = true },
			func(r *proto.RaidSimRequest) {
				r.Raid.Buffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectRegular
			}, stats.Strength},
		{"Camp Chair", func(c *proto.CampBuffs) { c.CampChair = true },
			func(r *proto.RaidSimRequest) { r.Raid.Buffs.MoonkinAura = true }, stats.MeleeCrit},
		{"Incense Candle", func(c *proto.CampBuffs) { c.IncenseCandle = true },
			func(r *proto.RaidSimRequest) { r.Raid.Buffs.ArcaneBrilliance = true }, stats.Intellect},
		{"Faction Banner", func(c *proto.CampBuffs) { c.FactionBanner = true },
			func(r *proto.RaidSimRequest) { r.Raid.Buffs.DivineSpirit = true }, stats.Spirit},
		{"Mana Well", func(c *proto.CampBuffs) { c.ManaWell = true },
			func(r *proto.RaidSimRequest) {
				r.Raid.Parties[0].Players[0].Buffs.BlessingOfWisdom = proto.TristateEffect_TristateEffectRegular
			}, stats.MP5},
		{"Enchanted Lute", func(c *proto.CampBuffs) { c.EnchantedLute = true },
			func(r *proto.RaidSimRequest) { r.Raid.Buffs.GiftOfTheWild = proto.TristateEffect_TristateEffectRegular }, stats.Stamina},
		{"First Aid Kit", func(c *proto.CampBuffs) { c.FirstAidKit = true },
			func(r *proto.RaidSimRequest) {
				r.Raid.Buffs.PowerWordFortitude = proto.TristateEffect_TristateEffectRegular
			}, stats.Stamina},
	}
	for _, c := range cases {
		build := func(camp, class bool) stats.Stats {
			// A Ret paladin uses mana, melee attack power and every attribute.
			rsr := raidReq(t, "paladin__level_60")
			rsr.Raid.Buffs = &proto.RaidBuffs{}
			rsr.Raid.Parties[0].Players[0].Buffs = &proto.IndividualBuffs{}
			if camp {
				rsr.Raid.Buffs.CampBuffs = &proto.CampBuffs{}
				c.camp(rsr.Raid.Buffs.CampBuffs)
			}
			if class {
				c.class(rsr)
			}
			return finalStats(t, rsr)
		}
		none, campOnly, classOnly, both := build(false, false), build(true, false), build(false, true), build(true, true)
		if campOnly[c.stat] <= none[c.stat] {
			t.Errorf("%s alone: %s %.2f, want more than %.2f", c.name, c.stat.StatName(), campOnly[c.stat], none[c.stat])
		}
		if both != classOnly {
			t.Errorf("%s with its class buff: the stats changed (%s %.2f, want %.2f)", c.name, c.stat.StatName(), both[c.stat], classOnly[c.stat])
		}
	}
}

// The Camp Chair is one aura, so Moonkin Aura takes its physical crit too, while Leader of the Pack
// stacks with it.
func TestRaidForeverCampChair(t *testing.T) {
	crit := func(moonkin, lotp bool) float64 {
		rsr := raidReq(t, "rogue__subtlety")
		rsr.Raid.Buffs = &proto.RaidBuffs{CampBuffs: &proto.CampBuffs{CampChair: true}, MoonkinAura: moonkin, LeaderOfThePack: lotp}
		return finalStats(t, rsr)[stats.MeleeCrit] / core.CritRatingPerCritChance
	}
	if d := crit(false, false) - crit(true, false); math.Abs(d-2) > 1e-9 {
		t.Errorf("Moonkin Aura should take the chair's 2%% melee crit, took %.3f", d)
	}
	if d := crit(false, true) - crit(false, false); math.Abs(d-3) > 1e-9 {
		t.Errorf("Leader of the Pack should add its 3%% on top of the chair, added %.3f", d)
	}
}

// The Sharpening Wheel adds nothing while the player's own Strength of Earth Totem stands.
func TestRaidForeverSharpeningWheel(t *testing.T) {
	dps := func(name string, wheel bool) float64 {
		rsr := raidReq(t, name)
		rsr.Raid.Buffs.StrengthOfEarthTotem = proto.TristateEffect_TristateEffectMissing
		rsr.Raid.Buffs.CampBuffs = &proto.CampBuffs{SharpeningWheel: wheel}
		rsr.SimOptions = &proto.SimOptions{Iterations: 200, RandomSeed: 7}
		return runRaid(t, rsr).RaidMetrics.Dps.Avg
	}
	if on, off := dps("warrior__fury", true), dps("warrior__fury", false); on <= off+1 {
		t.Errorf("a warrior with the wheel: %.1f DPS, want more than %.1f", on, off)
	}
	// Enhancement drops its own totem before the pull, so the wheel never counts in the fight.
	if on, off := dps("shaman__enhancement-level60", true), dps("shaman__enhancement-level60", false); math.Abs(on-off) > 0.5 {
		t.Errorf("Enhancement with its own totem: %.2f DPS with the wheel, %.2f without", on, off)
	}
}

// The ramp builds the given Sunder from the pull instead of starting full.
func TestRaidForeverSunderRamp(t *testing.T) {
	run := func(full bool, ramp *proto.SunderRamp) (float64, float64) {
		rsr := raidReq(t, "rogue__combat-swords")
		rsr.Raid.Debuffs.SunderArmor = full
		rsr.Raid.Debuffs.SunderRamp = ramp
		rsr.Encounter.DurationVariation = 0
		rsr.SimOptions = &proto.SimOptions{Iterations: 300, RandomSeed: 3}
		res := runRaid(t, rsr)
		uptime := 0.0
		for _, a := range res.EncounterMetrics.Targets[0].Auras {
			if a.Id.GetSpellId() == 11597 {
				uptime = a.UptimeSecondsAvg / rsr.Encounter.Duration
			}
		}
		return res.RaidMetrics.Dps.Avg, uptime
	}
	full, fullUp := run(true, nil)
	tank, tankUp := run(false, &proto.SunderRamp{Tank: true, PullSunders: 2})
	pull, pullUp := run(false, &proto.SunderRamp{PullSunders: 2})
	none, noneUp := run(false, nil)
	if !(full > tank && tank > pull && pull > none) {
		t.Errorf("DPS should fall from full (%.1f) to a tank's ramp (%.1f), pull Sunders only (%.1f) and none (%.1f)", full, tank, pull, none)
	}
	if fullUp < 0.999 || tankUp < 0.99 || noneUp != 0 {
		t.Errorf("uptimes: full %.3f, tank %.3f, none %.3f", fullUp, tankUp, noneUp)
	}
	// Two pull Sunders land 1.5 s in and fall off 30 s later.
	if want := 30.0 / 180; math.Abs(pullUp-want) > 0.01 {
		t.Errorf("pull Sunders only: uptime %.3f, want %.3f", pullUp, want)
	}
}
