package retribution

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	googleProto "google.golang.org/protobuf/proto"
)

var (
	sealOfWisdom3      = core.ActionID{SpellID: 20357}
	judgementOfWisdom3 = core.ActionID{SpellID: 20355}
	judgementSpell     = core.ActionID{SpellID: 20271}
	sealOfTheCrusader6 = core.ActionID{SpellID: 20308}
	judgementOfTheCru6 = core.ActionID{SpellID: 20303}
)

// sowSim is a Forever sim of a level 60 Human paladin with no talents and an Arcanite Reaper,
// against a level 60 target with no armor, with rotation and the raid's debuffs.
func sowSim(t *testing.T, rotation *proto.APLRotation, debuffs *proto.Debuffs) (*core.Simulation, *RetributionPaladin) {
	t.Helper()
	player := &proto.Player{
		Race:      proto.Race_RaceHuman,
		Class:     proto.Class_ClassPaladin,
		Level:     60,
		Equipment: &proto.EquipmentSpec{},
		Rotation:  rotation,
		Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
			Options: &proto.PaladinOptions{PrimarySeal: proto.PaladinSeal_Righteousness},
		}},
	}
	for slot := range proto.ItemSlot_ItemSlotRanged + 1 {
		id := int32(0)
		if slot == proto.ItemSlot_ItemSlotMainHand {
			id = arcaniteReaper
		}
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: id})
	}
	target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
	target.Level = 60
	target.Stats[proto.Stat_StatArmor] = 0
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, debuffs),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{target}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*RetributionPaladin)
}

// A rotation that names Seal of Wisdom (and Seal of the Crusader), so the paladin has it.
func sowRotation() *proto.APLRotation {
	cast := func(id int32) *proto.APLListItem {
		return &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_CastSpell{
			CastSpell: &proto.APLActionCastSpell{SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}}},
		}}}
	}
	return &proto.APLRotation{PriorityList: []*proto.APLListItem{cast(20357), cast(20308), cast(20271)}}
}

// castNow casts spell at the target with its cooldown and the global cooldown ready.
func castNow(t *testing.T, sim *core.Simulation, ret *RetributionPaladin, spell *core.Spell) {
	t.Helper()
	ret.SetGCDTimer(sim, sim.CurrentTime)
	if spell.CD.Timer != nil {
		spell.CD.Reset()
	}
	if !spell.Cast(sim, ret.CurrentTarget) {
		t.Fatalf("%s didn't cast", spell.ActionID)
	}
}

// swings lands n main-hand hits on the target (the hit listeners only, no damage) and returns how
// many gave exactly mana back. Mana is spent first so a gain always has room.
func swings(sim *core.Simulation, ret *RetributionPaladin, n int, mana float64) (gains int) {
	mh := ret.AutoAttacks.MHAuto()
	spent := ret.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionAttack})
	for range n {
		if room := ret.MaxMana() - ret.CurrentMana(); room < 2*mana {
			ret.SpendMana(sim, 2*mana-room, spent)
		}
		before := ret.CurrentMana()
		hit := mh.NewResult(ret.CurrentTarget)
		hit.Outcome = core.OutcomeHit
		ret.OnSpellHitDealt(sim, mh, hit)
		ret.CurrentTarget.OnSpellHitTaken(sim, mh, hit)
		if mana > 0 && math.Abs(ret.CurrentMana()-before-mana) < 1e-6 {
			gains++
		}
	}
	return
}

// Seal of Wisdom (client 1.60.1.70685): rank 3 at 58 for 200 mana, 30 sec, a melee hit gives the
// paladin 90 mana on our assumed 50%. Its judgement puts Judgement of Wisdom (20355) on the target
// for 40 sec, 59 mana to an attacker on half its hits, and the paladin's melee hits refresh it.
func TestSealOfWisdom(t *testing.T) {
	sim, ret := sowSim(t, sowRotation(), &proto.Debuffs{})
	for _, id := range []int32{20166, 20356, 20357, 20186, 20354, 20355} {
		if ret.GetSpell(core.ActionID{SpellID: id}) == nil {
			t.Errorf("no spell %d", id)
		}
	}
	seal := ret.GetSpell(sealOfWisdom3)
	if seal == nil {
		t.Fatal("no Seal of Wisdom")
	}

	mana := ret.CurrentMana()
	castNow(t, sim, ret, seal)
	if got := mana - ret.CurrentMana(); math.Abs(got-200) > 1e-6 {
		t.Errorf("Seal of Wisdom cost %.1f mana, want 200", got)
	}
	aura := ret.GetAuraByID(sealOfWisdom3)
	if aura == nil || !aura.IsActive() || aura.Duration != 30*time.Second {
		t.Fatalf("Seal of Wisdom isn't up for 30 sec: %v", aura)
	}

	// The seal's own mana, with nothing on the target.
	if rate := float64(swings(sim, ret, 4000, 90)) / 4000; math.Abs(rate-0.5) > 0.03 {
		t.Errorf("Seal of Wisdom gave 90 mana on %.1f%% of hits, want 50%%", rate*100)
	}

	// The judgement.
	castNow(t, sim, ret, ret.GetSpell(judgementSpell))
	jow := ret.CurrentTarget.GetAuraByID(judgementOfWisdom3)
	if jow == nil || !jow.IsActive() {
		t.Fatal("Judgement of Wisdom isn't on the target")
	}
	if jow.Duration != 40*time.Second {
		t.Errorf("Judgement of Wisdom lasts %s, want 40s", jow.Duration)
	}
	// Under Forever the judgement leaves the seal up.
	if !aura.IsActive() {
		t.Error("the judgement took Seal of Wisdom off")
	}

	// The paladin's melee hits refresh it.
	sim.CurrentTime += 20 * time.Second
	swings(sim, ret, 1, 0)
	if got := jow.RemainingDuration(sim); got != 40*time.Second {
		t.Errorf("Judgement of Wisdom has %s left after a melee hit, want 40s", got)
	}

	// Judgement of Wisdom's mana: 59 on half the hits, with no seal up.
	aura.Deactivate(sim)
	if rate := float64(swings(sim, ret, 4000, 59)) / 4000; math.Abs(rate-0.5) > 0.03 {
		t.Errorf("Judgement of Wisdom gave 59 mana on %.1f%% of swings, want 50%%", rate*100)
	}

	// One judgement per paladin: judging the Crusader takes our Wisdom off, and back.
	castNow(t, sim, ret, ret.GetSpell(sealOfTheCrusader6))
	castNow(t, sim, ret, ret.GetSpell(judgementSpell))
	jotc := ret.CurrentTarget.GetAuraByID(judgementOfTheCru6)
	if !jotc.IsActive() || jow.IsActive() {
		t.Errorf("after judging the Crusader: Crusader up %v (want true), Wisdom up %v (want false)", jotc.IsActive(), jow.IsActive())
	}
	castNow(t, sim, ret, seal)
	castNow(t, sim, ret, ret.GetSpell(judgementSpell))
	if jotc.IsActive() || !jow.IsActive() {
		t.Errorf("after judging Wisdom: Crusader up %v (want false), Wisdom up %v (want true)", jotc.IsActive(), jow.IsActive())
	}
}

// With the raid's Judgement of Wisdom given, the paladin's judgement only refreshes it: it stays the
// settings' permanent one, and judging the Crusader leaves it up.
func TestSealOfWisdomWithGivenJudgement(t *testing.T) {
	sim, ret := sowSim(t, sowRotation(), &proto.Debuffs{JudgementOfWisdom: true})
	jow := ret.CurrentTarget.GetAuraByID(judgementOfWisdom3)
	if jow == nil || !jow.IsActive() || jow.Duration != core.NeverExpires {
		t.Fatal("the given Judgement of Wisdom isn't up for good")
	}
	castNow(t, sim, ret, ret.GetSpell(sealOfWisdom3))
	castNow(t, sim, ret, ret.GetSpell(judgementSpell))
	castNow(t, sim, ret, ret.GetSpell(sealOfTheCrusader6))
	castNow(t, sim, ret, ret.GetSpell(judgementSpell))
	if !jow.IsActive() || jow.Duration != core.NeverExpires {
		t.Error("the paladin's judgements took the given Judgement of Wisdom off or made it expire")
	}
}

// A rotation that doesn't name Seal of Wisdom has none, so its auras don't join the results of
// every paladin build (the report's results stay as they were).
func TestNoSealOfWisdomWithoutTheRotation(t *testing.T) {
	_, ret := sowSim(t, &proto.APLRotation{}, &proto.Debuffs{})
	if ret.GetSpell(sealOfWisdom3) != nil || ret.GetAuraByID(sealOfWisdom3) != nil {
		t.Error("a rotation without Seal of Wisdom has it")
	}
	if ret.CurrentTarget.GetAuraByID(judgementOfWisdom3) != nil {
		t.Error("a rotation without Seal of Wisdom put a Judgement of Wisdom aura on the target")
	}
}
