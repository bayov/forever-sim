package retribution

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	googleProto "google.golang.org/protobuf/proto"
)

const arcaniteReaper = 12784 // 153 - 256, 3.8 speed, two-hand axe
const thunderfury = 19019    // 44 - 115, 1.9 speed, one-hand sword

// retSim starts a Forever sim of a level 60 Human paladin with talents, holding weapon,
// against a level 60 target with no armor, and returns it before the first action.
func retSim(t *testing.T, talents string, weapon int32) (*core.Simulation, *RetributionPaladin) {
	t.Helper()
	return retSimTargets(t, talents, weapon, 1)
}

// retSimTargets is retSim against n such targets.
func retSimTargets(t *testing.T, talents string, weapon int32, n int) (*core.Simulation, *RetributionPaladin) {
	t.Helper()
	return retSimGear(t, talents, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: weapon}, n, proto.MobType_MobTypeUnknown)
}

// retSimGear is retSim with items in their slots, against n targets of mobType.
func retSimGear(t *testing.T, talents string, items map[proto.ItemSlot]int32, n int, mobType proto.MobType) (*core.Simulation, *RetributionPaladin) {
	t.Helper()
	return retSimOptions(t, talents, items, n, mobType, &proto.PaladinOptions{PrimarySeal: proto.PaladinSeal_Righteousness})
}

// retSimOptions is retSimGear with the paladin's options (Righteous Fury, say).
func retSimOptions(t *testing.T, talents string, items map[proto.ItemSlot]int32, n int, mobType proto.MobType, options *proto.PaladinOptions) (*core.Simulation, *RetributionPaladin) {
	t.Helper()
	return retSimRaidBuffs(t, talents, items, n, mobType, options, &proto.RaidBuffs{})
}

// retSimRaidBuffs is retSimOptions with the raid's buffs.
func retSimRaidBuffs(t *testing.T, talents string, items map[proto.ItemSlot]int32, n int, mobType proto.MobType, options *proto.PaladinOptions, raidBuffs *proto.RaidBuffs) (*core.Simulation, *RetributionPaladin) {
	t.Helper()
	player := &proto.Player{
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassPaladin,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
			Options: options,
		}},
	}
	for slot := range proto.ItemSlot_ItemSlotRanged + 1 {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: items[slot]})
	}
	var targets []*proto.Target
	for range n {
		target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
		target.Level = 60
		target.Stats[proto.Stat_StatArmor] = 0
		target.MobType = mobType
		targets = append(targets, target)
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, raidBuffs, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: targets},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*RetributionPaladin)
}

// landedHit returns the damage of one cast of spell when it was a normal hit, and false
// for anything else.
func landedHit(sim *core.Simulation, ret *RetributionPaladin, spell *core.Spell) (float64, bool) {
	metrics := &spell.SpellMetrics[ret.CurrentTarget.UnitIndex]
	before := *metrics
	spell.ApplyEffects(sim, ret.CurrentTarget, spell)
	if metrics.Hits == before.Hits+1 && metrics.Crits == before.Crits {
		return metrics.TotalDamage - before.TotalDamage, true
	}
	return 0, false
}

// TestHolyStrikeDamage checks Holy Strike rank 8 against the client's spell (10333): 50% of
// (normalized weapon damage at Forever's 3.6 for a two-hander, plus 93 Holy with a 25%
// spread). Arcanite Reaper's 153 - 256 and the paladin's attack power give the range a normal
// hit must fall in. The old model (50% of the weapon at its own 3.8 speed, plus the whole 93)
// lands at least 38 higher.
func TestHolyStrikeDamage(t *testing.T) {
	sim, ret := retSim(t, "", arcaniteReaper)
	holyStrike := ret.GetSpell(core.ActionID{SpellID: 10333})
	if holyStrike == nil {
		t.Fatalf("no Holy Strike rank 8")
	}
	ap := holyStrike.MeleeAttackPower(ret.CurrentTarget)
	lo := 0.5 * (153 + 3.6*ap/14 + 93*0.875)
	hi := 0.5 * (256 + 3.6*ap/14 + 93*1.125)
	minHit, maxHit, hits := hi+1000, 0.0, 0
	for range 2000 {
		if hit, ok := landedHit(sim, ret, holyStrike); ok {
			minHit, maxHit, hits = min(minHit, hit), max(maxHit, hit), hits+1
		}
	}
	if hits < 1000 {
		t.Fatalf("only %d hits", hits)
	}
	// Each hit is rounded to whole points, up or down at random.
	if minHit < math.Floor(lo)-1e-6 || maxHit > math.Ceil(hi)+1e-6 {
		t.Errorf("hits %.1f to %.1f, want within %.1f to %.1f", minHit, maxHit, lo, hi)
	}
	if minHit > lo+5 || maxHit < hi-5 {
		t.Errorf("hits %.1f to %.1f do not span %.1f to %.1f", minHit, maxHit, lo, hi)
	}
}

func castByID(sim *core.Simulation, ret *RetributionPaladin, spellID int32) {
	spell := ret.GetSpell(core.ActionID{SpellID: spellID})
	spell.ApplyEffects(sim, ret.CurrentTarget, spell)
}

// TestLightsVigilHealsOnly checks Light's Vigil since the October 30 hotfix (spell 1310911 in
// 1.60.1.70707+hf: a friendly target only, heals only): with the talent there is no Vigil or
// detonation on an enemy (the Vigil itself is a heal now, sim/paladin/holy's tests), and Holy Shock
// at an enemy goes on its 10 sec cooldown.
func TestLightsVigilHealsOnly(t *testing.T) {
	sim, ret := retSim(t, "05025211220021051-5-05233", arcaniteReaper)
	if !ret.Talents.LightsVigil || !ret.Talents.HolyShock {
		t.Fatalf("no Light's Vigil or Holy Shock in the talents")
	}
	for _, id := range []int32{1311598, 1311599} {
		if spell := ret.GetSpell(core.ActionID{SpellID: id}); spell != nil {
			t.Errorf("spell %d registered, want no Light's Vigil on an enemy", id)
		}
	}
	holyShock := ret.GetSpell(core.ActionID{SpellID: 20930})
	holyShock.ApplyEffects(sim, ret.CurrentTarget, holyShock)
	holyShock.CD.Use(sim)
	if holyShock.CD.IsReady(sim) {
		t.Errorf("Holy Shock off cooldown")
	}
}
