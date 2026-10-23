package dpsrogue

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// The Mutilate preset's talents (tools/dps_report/req/rogue__mutilate.json).
const mutilatePresetTalents = "0053031033150103-302303202015-002"

func newMutilateRogueSim(t *testing.T, seconds float64) (*core.Simulation, *DpsRogue) {
	return newMutilateRogueSimWith(t, seconds, 18816) // Perdition's Blade, a dagger
}

func newMutilateRogueSimWith(t *testing.T, seconds float64, offHandID int32) (*core.Simulation, *DpsRogue) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotOffHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 18805} // Core Hound Tooth, a dagger
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: offHandID}
	player := &proto.Player{
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassRogue,
		Level:         60,
		TalentsString: mutilatePresetTalents,
		Equipment:     &proto.EquipmentSpec{Items: items},
		Rotation:      &proto.APLRotation{},
		Spec:          DefaultRogue,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: seconds, Targets: []*proto.Target{{Level: 63, MobType: proto.MobType_MobTypeHumanoid}}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	rogue := sim.Raid.Parties[0].Players[0].(*DpsRogue)
	if rogue.Mutilate == nil || rogue.Talents.SealFate == 0 {
		t.Fatal("the preset's talents give no Mutilate or no Seal Fate")
	}
	return sim, rogue
}

// "Mutilate will trigger this exactly once if either or both hands generates a critical
// strike" (World of Warcraft: Forever class deep dive, Rogue). With both halves always
// critting, Seal Fate's chance per Mutilate is its own chance, not 1-(1-p)².
func TestForeverSealFateOncePerMutilate(t *testing.T) {
	const casts = 3000
	sim, rogue := newMutilateRogueSim(t, casts+10)
	rogue.AddStatDynamic(sim, stats.MeleeCrit, 200*core.CritRatingPerCritChance)
	rogue.AddStatDynamic(sim, stats.MeleeHit, 20*core.MeleeHitRatingPerHitChance)
	target := rogue.CurrentTarget

	extra, landed := 0, 0
	for i := 0; i < casts; i++ {
		sim.AddPendingAction(&core.PendingAction{
			NextActionAt: core.DurationFromSeconds(float64(i) + 1),
			OnAction: func(sim *core.Simulation) {
				rogue.SpendComboPoints(sim, rogue.Mutilate)
				rogue.AddEnergy(sim, 100, rogue.EnergyRefundMetrics)
				rogue.SetGCDTimer(sim, sim.CurrentTime)
				before := rogue.ComboPoints()
				if !rogue.Mutilate.Cast(sim, target) {
					t.Fatalf("Mutilate didn't cast at %s", sim.CurrentTime)
				}
				if gained := rogue.ComboPoints() - before; gained >= 2 {
					landed++
					extra += int(gained) - 2
				}
			},
		})
	}
	for !sim.Step() {
	}

	p := 0.2 * float64(rogue.Talents.SealFate)
	rate := float64(extra) / float64(landed)
	// 3000 Bernoulli trials: the standard error is under 0.01.
	if math.Abs(rate-p) > 0.03 {
		t.Errorf("Seal Fate gave %.3f extra combo points a Mutilate, want %.2f (twice-rolled would be %.3f)", rate, p, 1-(1-p)*(1-p))
	}
}

// Cold Blood covers both halves of a Mutilate and is spent once by it.
func TestForeverColdBloodBothMutilateHalves(t *testing.T) {
	sim, rogue := newMutilateRogueSim(t, 60)
	if rogue.ColdBlood == nil {
		t.Skip("the preset's talents have no Cold Blood")
	}
	target := rogue.CurrentTarget
	rogue.AddStatDynamic(sim, stats.MeleeHit, 20*core.MeleeHitRatingPerHitChance)
	if !rogue.ColdBlood.Cast(sim, target) {
		t.Fatal("Cold Blood didn't cast")
	}
	coldBlood := rogue.GetAura("Cold Blood")
	oh := rogue.GetSpell(rogue.Mutilate.ActionID.WithTag(2))
	ohCrits := oh.SpellMetrics[target.UnitIndex].Crits
	rogue.SetGCDTimer(sim, sim.CurrentTime)
	if !rogue.Mutilate.Cast(sim, target) {
		t.Fatal("Mutilate didn't cast")
	}
	if coldBlood.IsActive() {
		t.Error("Cold Blood still up after Mutilate")
	}
	if got := oh.SpellMetrics[target.UnitIndex].Crits - ohCrits; got != 1 {
		t.Errorf("Mutilate's off hand half under Cold Blood: %d crits, want 1", got)
	}
}

// Mutilate needs a dagger in both hands: the client asks for one on the cast and both
// halves (SpellEquippedItems 1310705 to 1310707), and the class deep dive says "both must
// be daggers".
func TestForeverMutilateNeedsOffHandDagger(t *testing.T) {
	sim, rogue := newMutilateRogueSimWith(t, 60, 18832) // Brutality Blade, a sword
	rogue.SetGCDTimer(sim, sim.CurrentTime)
	if rogue.Mutilate.CanCast(sim, rogue.CurrentTarget) {
		t.Error("Mutilate can be cast with a sword in the off hand")
	}
}
