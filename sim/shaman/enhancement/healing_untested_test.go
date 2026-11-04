package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/shaman"
)

// The Restoration shaman's mechanics the other healing tests left out (healer_mechanics-20261104-10).
// Every value is the Forever client's (1.60.1.70753 on wago.tools), not the sim's own tables.

// TestOrcShamanHealRankTable checks every rank of Healing Wave, Lesser Healing Wave and Chain
// Heal against the client: its base heal at the level it's learned, its growth a level up to
// MaxLevel, its variance and its mana (SpellEffect, SpellLevels, SpellPower), and its cast time
// and coefficient (SpellMisc to SpellCastTimes, SpellEffect BonusCoefficient). Each rank is read
// at its learn level, where Forever's downranking keeps all of its coefficient.
func TestOrcShamanHealRankTable(t *testing.T) {
	type row struct {
		spellID                   int32
		learn, max                int32
		base, growth, variance    float64
		mana, castMs, coefficient float64
	}
	heals := []struct {
		name  string
		ranks []row
		at    func(rank int, level int32) (float64, float64)
		spell func(enh *EnhancementShaman, rank int) *core.Spell
	}{
		{name: "Healing Wave", ranks: []row{
			{331, 1, 5, 39, 0.7, 0.25641027093, 25, 1500, 0.429},
			{332, 6, 11, 71, 1, 0.19718310237, 45, 2000, 0.571},
			{547, 12, 17, 131, 1.5, 0.18309858441, 80, 2500, 0.714},
			{913, 18, 23, 256, 2.3, 0.16438356042, 155, 3000, 0.857},
			{939, 24, 29, 342, 2.7, 0.15686275065, 200, 3000, 0.857},
			{959, 32, 37, 484, 3.3, 0.14853195846, 265, 3000, 0.857},
			{8005, 40, 45, 685, 3.9, 0.14303638041, 340, 3000, 0.857},
			{10395, 48, 53, 972, 4.7, 0.13736264408, 440, 3000, 0.857},
			{10396, 56, 61, 1344, 5.5, 0.13251365721, 560, 3000, 0.857},
			{25357, 60, 65, 1615, 5.5, 0.1325648427, 620, 3000, 0.857},
		},
			at:    func(rank int, level int32) (float64, float64) { return shaman.HealingWaveRanks[rank].At(level) },
			spell: func(enh *EnhancementShaman, rank int) *core.Spell { return enh.HealingWave[rank] },
		},
		{name: "Lesser Healing Wave", ranks: []row{
			{8004, 20, 25, 153, 1.7, 0.13793103397, 105, 1500, 0.429},
			{8008, 28, 33, 224, 2.1, 0.12878787518, 145, 1500, 0.429},
			{8010, 36, 41, 304, 2.5, 0.12256267667, 185, 1500, 0.429},
			{10466, 44, 49, 426, 3, 0.11522634327, 235, 1500, 0.429},
			{10467, 52, 57, 608, 3.6, 0.11077844352, 305, 1500, 0.429},
			{10468, 60, 65, 820, 4.2, 0.10909090936, 380, 1500, 0.429},
		},
			at:    func(rank int, level int32) (float64, float64) { return shaman.LesserHealingWaveRanks[rank].At(level) },
			spell: func(enh *EnhancementShaman, rank int) *core.Spell { return enh.LesserHealingWave[rank] },
		},
		{name: "Chain Heal", ranks: []row{
			{1064, 40, 45, 251, 2.5, 0.13953489065, 260, 2500, 0.714},
			{10622, 46, 51, 335, 2.8, 0.13793103397, 315, 2500, 0.714},
			{10623, 54, 59, 490, 3.3, 0.13220338523, 405, 2500, 0.714},
		},
			at:    func(rank int, level int32) (float64, float64) { return shaman.ChainHealRanks[rank].At(level) },
			spell: func(enh *EnhancementShaman, rank int) *core.Spell { return enh.ChainHeal[rank] },
		},
	}
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-4 }
	for _, h := range heals {
		for i, r := range h.ranks {
			rank := i + 1
			// The heal at the learn level is the base, spread by the variance.
			low, high := h.at(rank, r.learn)
			if mid := (low + high) / 2; !near(mid, r.base) || !near((high-low)/mid, r.variance) {
				t.Errorf("%s %d (%d) at %d heals %.2f to %.2f, want %.0f with %.4f variance", h.name, rank, r.spellID, r.learn, low, high, r.base, r.variance)
			}
			// It grows to MaxLevel and stops there.
			low, high = h.at(rank, r.max)
			want := r.base + r.growth*float64(r.max-r.learn)
			if mid := (low + high) / 2; !near(mid, want) {
				t.Errorf("%s %d (%d) at %d heals %.2f on average, want %.2f (%.1f a level)", h.name, rank, r.spellID, r.max, mid, want, r.growth)
			}
			if l, hi := h.at(rank, r.max+10); l != low || hi != high {
				t.Errorf("%s %d (%d) grows past level %d", h.name, rank, r.spellID, r.max)
			}

			_, enh := healSim(t, r.learn, "", 0, 0)
			spell := h.spell(enh, rank)
			if spell == nil {
				t.Errorf("a level %d doesn't know %s %d (%d)", r.learn, h.name, rank, r.spellID)
				continue
			}
			if spell.SpellID != r.spellID {
				t.Errorf("%s %d is %d, want %d", h.name, rank, spell.SpellID, r.spellID)
			}
			if got := spell.Cost.GetCurrentCost(); got != r.mana {
				t.Errorf("%s %d (%d) costs %.1f, want %.0f", h.name, rank, r.spellID, got, r.mana)
			}
			if got, want := spell.DefaultCast.CastTime, time.Duration(r.castMs)*time.Millisecond; got != want {
				t.Errorf("%s %d (%d) casts in %s, want %s", h.name, rank, r.spellID, got, want)
			}
			if got := spell.BonusCoefficient; !near(got, r.coefficient) {
				t.Errorf("%s %d (%d) has a %.4f coefficient at %d, want %.3f", h.name, rank, r.spellID, got, r.learn, r.coefficient)
			}
		}
	}
}

// TestOrcShamanChainHealAmount casts Chain Heal rank 3 (10623) at ourselves, with no one to
// jump to. At 60 it heals 490 + 3.3 a level to 59, 506.5, +- 6.6% (SpellEffect, SpellLevels),
// and 0.714 of healing power on top (BonusCoefficient). With the same seed, 1000 healing power
// adds 714, where 0.857 or 0.571 would add 857 or 571.
func TestOrcShamanChainHealAmount(t *testing.T) {
	heal := func(power float64) float64 {
		sim, enh := healSim(t, 60, "", power, 0)
		got := healOnce(t, sim, enh, enh.ChainHeal[3], 1, 2.5)
		runSim(sim)
		return *got
	}
	without, with := heal(0), heal(1000)
	if low, high := 506.5*(1-0.13220338523/2), 506.5*(1+0.13220338523/2); without < low-0.5 || without > high+0.5 {
		t.Errorf("Chain Heal healed %.1f with no healing power, want %.1f to %.1f", without, low, high)
	}
	if got := with - without; math.Abs(got-714) > 0.5 {
		t.Errorf("1000 healing power adds %.1f to Chain Heal, want 714", got)
	}
}

// TestOrcShamanRiptideTickCrits: Riptide's ticks crit under Forever (1239243's SpellMisc
// Attributes_8 has 0x200, the periodic crit flag), for 1.5 times, and Tidal Mastery's crit
// (16194: aura 107 on mask word 2 0x10, Riptide's) reaches them.
func TestOrcShamanRiptideTickCrits(t *testing.T) {
	// Every tick crits: 5 of 161 + 10% of healing power, Purification's 10%, 1.5 times.
	sim, enh := healSim(t, 60, restoRiptidePurif, 100, 200)
	healOnce(t, sim, enh, enh.Riptide, 1, 0)
	var start float64
	at(sim, 1.15, func(sim *core.Simulation) { start = enh.CurrentHealth() })
	at(sim, 16.5, func(sim *core.Simulation) {
		if got, want := enh.CurrentHealth()-start, 5*(161+0.1*100)*1.1*1.5; math.Abs(got-want) > 1 {
			t.Errorf("Riptide's ticks healed %.1f with every tick a crit, want %.1f", got, want)
		}
	})
	runSim(sim)
	if got := enh.Riptide.SpellMetrics[enh.UnitIndex].CritTicks; got != 5 {
		t.Errorf("%d of Riptide's 5 ticks crit at 100%% crit, want 5", got)
	}

	// Tidal Mastery 5 and no other crit: 5% of the ticks crit. We tick the heal over time by
	// hand, so the count doesn't need a long fight.
	const ticks = 4000
	crits := func(talents string) int32 {
		sim, enh := healSim(t, 60, talents, 0, 0)
		var n int32
		at(sim, 1, func(sim *core.Simulation) {
			castNow(t, sim, enh, enh.Riptide)
			hot := enh.Riptide.Hot(&enh.Unit)
			for range ticks {
				hot.TickOnce(sim)
			}
			n = enh.Riptide.SpellMetrics[enh.UnitIndex].CritTicks
		})
		runSim(sim)
		return n
	}
	if got := crits(restoRiptidePurif); got != 0 {
		t.Errorf("%d of %d Riptide ticks crit with no crit, want none", got, ticks)
	}
	got := crits(resto(restoIdxTidalMastery, 5, "000051"))
	want, margin := ticks*0.05, 4*math.Sqrt(ticks*0.05*0.95)
	if math.Abs(float64(got)-want) > margin {
		t.Errorf("Tidal Mastery 5: %d of %d Riptide ticks crit, want %.0f +- %.0f", got, ticks, want, margin)
	}
}

// TestOrcShamanRiptideTickCritsProcNothing: a Riptide tick that crits doesn't give Ancestral
// Fortitude or spend a Water Shield globe. Their proc flags (SpellAuraOptions ProcTypeMask:
// 16176 0x15550, 408510 0x362a8) leave out periodic heals (0x40000). A Lesser Healing Wave
// crit after does both.
func TestOrcShamanRiptideTickCritsProcNothing(t *testing.T) {
	// Ancestral Healing 3, Water Shield, Purification 5 and Riptide.
	sim, enh := healSim(t, 60, "--0000003010000051", 0, 200)
	fortitude := enh.GetAura("Ancestral Fortitude")
	if fortitude == nil || enh.WaterShield == nil {
		t.Fatal("no Ancestral Fortitude or Water Shield")
	}
	at(sim, 0.5, func(sim *core.Simulation) { castNow(t, sim, enh, enh.WaterShield) })
	healOnce(t, sim, enh, enh.Riptide, 1, 0)
	// After each tick, at 4, 7, 10, 13 and 16 sec.
	for _, s := range []float64{4.1, 7.1, 10.1, 13.1, 16.1} {
		at(sim, s, func(sim *core.Simulation) {
			if fortitude.IsActive() {
				t.Errorf("at %s: Ancestral Fortitude from a Riptide tick", sim.CurrentTime)
			}
			if got := enh.WaterShieldAura.GetStacks(); got != 3 {
				t.Errorf("at %s: Water Shield has %d globes after Riptide ticks, want 3", sim.CurrentTime, got)
			}
		})
	}
	healOnce(t, sim, enh, enh.LesserHealingWave[6], 17, 1.5)
	at(sim, 19, func(sim *core.Simulation) {
		if crits := enh.Riptide.SpellMetrics[enh.UnitIndex].CritTicks; crits != 5 {
			t.Errorf("%d of Riptide's 5 ticks crit, want 5", crits)
		}
		if !fortitude.IsActive() || enh.WaterShieldAura.GetStacks() != 2 {
			t.Errorf("after a Lesser Healing Wave crit: Ancestral Fortitude up %v, %d globes; want up and 2",
				fortitude.IsActive(), enh.WaterShieldAura.GetStacks())
		}
	})
	runSim(sim)
}

// TestOrcShamanNaturesSwiftnessKept: Riptide and a totem don't use up Nature's Swiftness.
// 16188's mask (0x9c3, word 2 0x100000, word 3 0x10000) has neither Riptide (word 2 0x10) nor
// Healing Stream Totem (0x80000), and neither has a cast time. The Healing Wave after uses it.
func TestOrcShamanNaturesSwiftnessKept(t *testing.T) {
	// Nature's Swiftness and Riptide.
	sim, enh := healSim(t, 60, "--0000000000000101", 0, 0)
	ns := enh.GetAura("Natures Swiftness")
	if ns == nil || enh.Riptide == nil {
		t.Fatal("no Nature's Swiftness or Riptide")
	}
	at(sim, 1, func(sim *core.Simulation) {
		ns.Activate(sim)
		castNow(t, sim, enh, enh.Riptide)
		if !ns.IsActive() {
			t.Error("Riptide used up Nature's Swiftness")
		}
	})
	at(sim, 3, func(sim *core.Simulation) {
		castNow(t, sim, enh, topRank(t, "Healing Stream Totem", enh.HealingStreamTotem))
		if !ns.IsActive() {
			t.Error("Healing Stream Totem used up Nature's Swiftness")
		}
	})
	at(sim, 5, func(sim *core.Simulation) {
		castNow(t, sim, enh, enh.HealingWave[10])
		if ns.IsActive() || enh.IsCasting(sim) {
			t.Errorf("after Healing Wave: Nature's Swiftness up %v, casting %v; want neither", ns.IsActive(), enh.IsCasting(sim))
		}
	})
	runSim(sim)
}

// TestOrcShamanHealingStreamNoTidalTalents: Tidal Focus (16179, mask 0x1c0 and word 2 0x10) and
// Tidal Mastery (16194, mask 0x5c0 and word 2 0x10) leave out Healing Stream Totem (5394-10463,
// 0x80000) and its heal (5672-10461, 0x2000): the totem costs its 80 mana (SpellPower of
// 10463), and the heal's crit and multiplier stay as without them.
func TestOrcShamanHealingStreamNoTidalTalents(t *testing.T) {
	_, plain := healSim(t, 60, "", 0, 0)
	// Tidal Focus 5 and Tidal Mastery 5.
	_, enh := healSim(t, 60, "--0500000005", 0, 0)
	if got := enh.HealingWave[10].Cost.GetCurrentCost(); got != 620*0.95 {
		t.Fatalf("Healing Wave costs %.1f with Tidal Focus 5, want %.1f", got, 620*0.95)
	}
	totem := topRank(t, "Healing Stream Totem", enh.HealingStreamTotem)
	if totem.SpellID != 10463 {
		t.Fatalf("the top Healing Stream Totem is %d, want 10463", totem.SpellID)
	}
	if got := totem.Cost.GetCurrentCost(); got != 80 {
		t.Errorf("Healing Stream Totem costs %.1f with Tidal Focus 5, want 80", got)
	}
	heal, base := enh.GetSpell(core.ActionID{SpellID: 10461}), plain.GetSpell(core.ActionID{SpellID: 10461})
	if heal == nil || base == nil {
		t.Fatal("no Healing Stream heal (10461)")
	}
	if heal.BonusCritRating != base.BonusCritRating || heal.HealingCritChance() != base.HealingCritChance() {
		t.Errorf("Tidal Mastery 5 gives Healing Stream's heal %.3f crit, want %.3f as without it",
			heal.HealingCritChance(), base.HealingCritChance())
	}
	if heal.DamageMultiplier != base.DamageMultiplier {
		t.Errorf("Healing Stream's heal is x%.3f with the talents, want x%.3f", heal.DamageMultiplier, base.DamageMultiplier)
	}
}

// TestOrcShamanAncestralFortitudeTwoShamans: two shamans with Ancestral Healing 3 crit on the
// same player. Ancestral Fortitude (16177) is one buff spell with no stacks (no
// SpellAuraOptions row), so the second crit refreshes the first's 15 sec and armor stays 25%
// more, not 56%.
func TestOrcShamanAncestralFortitudeTwoShamans(t *testing.T) {
	player := func() *proto.Player {
		return withItemArmor(newOrcShaman(60, restoAncestralHealing, &proto.EnhancementShaman_Options{}), 2000)
	}
	raid := core.SinglePlayerRaidProto(player(), &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Parties[0].Players = append(raid.Parties[0].Players, player())
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Duration: 30, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1, ExactDamage: true},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	a := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
	b := sim.Raid.Parties[0].Players[1].(*EnhancementShaman)
	for _, s := range []*EnhancementShaman{a, b} {
		s.AddStatDynamic(sim, stats.Health, 5000)
		s.AddStatDynamic(sim, stats.SpellCrit, 200-s.GetStat(stats.SpellCrit))
	}
	buff := b.GetAura("Ancestral Fortitude")
	if buff == nil {
		t.Fatal("no Ancestral Fortitude")
	}
	heal := func(caster *EnhancementShaman, target *core.Unit) func(sim *core.Simulation) {
		return func(sim *core.Simulation) {
			b.RemoveHealth(sim, b.CurrentHealth()-1)
			caster.SetGCDTimer(sim, sim.CurrentTime)
			if !caster.LesserHealingWave[6].Cast(sim, target) {
				t.Fatalf("at %s: Lesser Healing Wave didn't cast", sim.CurrentTime)
			}
		}
	}
	var armor float64
	at(sim, 0.5, func(sim *core.Simulation) { armor = b.Armor() })
	// a heals b, landing at 2.5 sec; b heals itself, landing at 9.5.
	at(sim, 1, heal(a, &b.Unit))
	at(sim, 8, heal(b, sim.Encounter.TargetUnits[0]))
	check := func(s float64, up bool) {
		at(sim, s, func(sim *core.Simulation) {
			want := armor
			if up {
				want *= 1.25
			}
			if buff.IsActive() != up || math.Abs(b.Armor()-want) > 0.01 {
				t.Errorf("at %s: the buff is up %v and armor %.1f, want up %v and %.1f", sim.CurrentTime, buff.IsActive(), b.Armor(), up, want)
			}
		})
	}
	check(3, true)
	check(11, true)
	// a's buff alone would have ended at 17.5.
	check(19, true)
	check(25, false)
	runSim(sim)
}
