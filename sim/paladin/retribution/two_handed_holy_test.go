package retribution

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	googleProto "google.golang.org/protobuf/proto"
)

// Phase45RetTalents with Two-Handed Weapon Specialization at 0 instead of 3.
const phase45RetNo2hSpec = "050050002-503-05233101201001351"

// TestTwoHandedSpecializationHolyReach checks Blizzard's November 19 hotfix (t/2403040 post 13):
// "Two-Handed Weapon Specialization now correctly increases all Holy damage you deal while a
// two-handed weapon is equipped." Under Forever every Holy spell takes 6% at 3/3 with a two-hander,
// once, through the Holy school: the seal hits, Holy Strike, Judgement of Command and Consecration
// no longer carry it on their own multipliers too. Physical keeps it school-wide.
func TestTwoHandedSpecializationHolyReach(t *testing.T) {
	items := map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: arcaniteReaper}
	holy := stats.SchoolIndexHoly
	for _, c := range []struct {
		name             string
		seal             proto.PaladinSeal
		id               int32
		talents, without string
	}{
		{"Seal of Righteousness's hit", proto.PaladinSeal_Righteousness, 25742, Phase45RetTalents, phase45RetNo2hSpec},
		{"Seal of Command's hit", proto.PaladinSeal_Command, 20424, Phase45RetTalents, phase45RetNo2hSpec},
		{"Judgement of Command", proto.PaladinSeal_Command, 20966, Phase45RetTalents, phase45RetNo2hSpec},
		{"Holy Strike", proto.PaladinSeal_Righteousness, 10333, Phase45RetTalents, phase45RetNo2hSpec},
		{"Consecration", proto.PaladinSeal_Righteousness, 20924, Phase45RetTalents, phase45RetNo2hSpec},
		{"Judgement of Righteousness", proto.PaladinSeal_Righteousness, 20286, Phase45RetTalents, phase45RetNo2hSpec},
		{"Exorcism", proto.PaladinSeal_Righteousness, 10314, Phase45RetTalents, phase45RetNo2hSpec},
		{"Hammer of Wrath", proto.PaladinSeal_Righteousness, 24239, Phase45RetTalents, phase45RetNo2hSpec},
		{"Holy Wrath", proto.PaladinSeal_Righteousness, 10318, Phase45RetTalents, phase45RetNo2hSpec},
		{"Holy Shock", proto.PaladinSeal_Righteousness, 20930, "05025211220021051--000000000003", "05025211220021051"},
	} {
		t.Run(c.name, func(t *testing.T) {
			options := &proto.PaladinOptions{PrimarySeal: c.seal}
			_, with := retSimOptions(t, c.talents, items, 1, proto.MobType_MobTypeUndead, options)
			_, without := retSimOptions(t, c.without, items, 1, proto.MobType_MobTypeUndead, options)
			s, b := with.GetSpell(core.ActionID{SpellID: c.id}), without.GetSpell(core.ActionID{SpellID: c.id})
			if s == nil || b == nil {
				t.Fatalf("spell %d is not registered", c.id)
			}
			if s.DamageMultiplier != b.DamageMultiplier {
				t.Errorf("%s (%d) own multiplier %.4f with 3/3, want %.4f (the Holy school carries it)", c.name, c.id, s.DamageMultiplier, b.DamageMultiplier)
			}
			got := s.DamageMultiplier * with.PseudoStats.SchoolDamageDealtMultiplier[holy] /
				(b.DamageMultiplier * without.PseudoStats.SchoolDamageDealtMultiplier[holy])
			if got < 1.0599 || got > 1.0601 {
				t.Errorf("Two-Handed Weapon Specialization 3/3 on %s (%d) = %.4f, want 1.06", c.name, c.id, got)
			}
		})
	}

	_, with := retSimOptions(t, Phase45RetTalents, items, 1, proto.MobType_MobTypeUnknown, &proto.PaladinOptions{})
	_, without := retSimOptions(t, phase45RetNo2hSpec, items, 1, proto.MobType_MobTypeUnknown, &proto.PaladinOptions{})
	physical := stats.SchoolIndexPhysical
	if got := with.PseudoStats.SchoolDamageDealtMultiplier[physical] / without.PseudoStats.SchoolDamageDealtMultiplier[physical]; got < 1.0599 || got > 1.0601 {
		t.Errorf("Physical damage with 3/3 and a two-hander = %.4f, want 1.06", got)
	}
}

// TestTwoHandedSpecializationItemSwap checks that the talent follows the main hand ("while a
// two-handed weapon is equipped"): swapping the two-hander for a one-hander drops the 6% on Physical
// and Holy and the 20% threat cut, swapping back brings them back, and the sim's reset, which swaps
// the items back, leaves them as they were at the start.
func TestTwoHandedSpecializationItemSwap(t *testing.T) {
	player := &proto.Player{
		Race:           proto.Race_RaceHuman,
		Class:          proto.Class_ClassPaladin,
		Level:          60,
		Equipment:      &proto.EquipmentSpec{},
		TalentsString:  Phase45RetTalents,
		Rotation:       &proto.APLRotation{},
		EnableItemSwap: true,
		ItemSwap:       &proto.ItemSwap{MhItem: &proto.ItemSpec{Id: thunderfury}},
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
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{target}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	ret := sim.Raid.Parties[0].Players[0].(*RetributionPaladin)
	if !ret.ItemSwap.IsEnabled() {
		t.Fatalf("item swap is not enabled")
	}

	type state struct{ physical, holy, threat float64 }
	read := func() state {
		return state{
			ret.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical],
			ret.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexHoly],
			ret.PseudoStats.ThreatMultiplier,
		}
	}
	near := func(what string, got, want float64) {
		t.Helper()
		if got < want-0.0001 || got > want+0.0001 {
			t.Errorf("%s = %.4f, want %.4f", what, got, want)
		}
	}
	start := read()
	mainHand := []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}

	ret.ItemSwap.SwapItems(sim, mainHand)
	if ret.MainHand().ID != thunderfury {
		t.Fatalf("main hand %d after the swap, want %d", ret.MainHand().ID, thunderfury)
	}
	oneHand := read()
	near("Physical with a one-hander, as a share of the two-hander's", oneHand.physical/start.physical, 1/1.06)
	near("Holy with a one-hander, as a share of the two-hander's", oneHand.holy/start.holy, 1/1.06)
	near("threat with a one-hander, as a share of the two-hander's", oneHand.threat/start.threat, 1/0.8)

	ret.ItemSwap.SwapItems(sim, mainHand)
	back := read()
	near("Physical after swapping back", back.physical, start.physical)
	near("Holy after swapping back", back.holy, start.holy)
	near("threat after swapping back", back.threat, start.threat)

	ret.ItemSwap.SwapItems(sim, mainHand)
	sim.Cleanup()
	sim.Reset()
	sim.PrePull()
	if ret.MainHand().ID != arcaniteReaper {
		t.Fatalf("main hand %d after the reset, want %d", ret.MainHand().ID, arcaniteReaper)
	}
	reset := read()
	near("Physical after the reset", reset.physical, start.physical)
	near("Holy after the reset", reset.holy, start.holy)
	near("threat after the reset", reset.threat, start.threat)
}

// TestUnsealedWrathMainHandSwapDropsEchoes checks that swapping the main hand removes every Echo of
// Unsealed Wrath, which needs a weapon (EV-0065).
func TestUnsealedWrathMainHandSwapDropsEchoes(t *testing.T) {
	player := &proto.Player{
		Race:           proto.Race_RaceHuman,
		Class:          proto.Class_ClassPaladin,
		Level:          60,
		Equipment:      &proto.EquipmentSpec{},
		TalentsString:  unsealedWrath,
		Rotation:       &proto.APLRotation{},
		EnableItemSwap: true,
		ItemSwap:       &proto.ItemSwap{MhItem: &proto.ItemSpec{Id: thunderfury}},
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
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	ret := sim.Raid.Parties[0].Players[0].(*RetributionPaladin)
	uw, righteousness := ret.GetAura("Unsealed Wrath"), ret.GetAura("Echo of Righteousness")
	castByID(sim, ret, sealOfRighteousness9)
	whiteHit(sim, ret, uw, core.OutcomeHit)
	if !righteousness.IsActive() {
		t.Fatal("a landed swing stored no Echo")
	}
	ret.ItemSwap.SwapItems(sim, []proto.ItemSlot{proto.ItemSlot_ItemSlotRanged})
	if !righteousness.IsActive() {
		t.Error("a swap that leaves the main hand removed the Echo")
	}
	ret.ItemSwap.SwapItems(sim, []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand})
	if righteousness.IsActive() {
		t.Error("a main-hand swap left the Echo up")
	}
}

// TestRetributionAuraTakesNoCrusaderShare checks that Retribution Aura's hit takes no share of the
// target's flat Holy bonus, Judgement of the Crusader's (EV-0066), where a Holy spell with a
// coefficient does.
func TestRetributionAuraTakesNoCrusaderShare(t *testing.T) {
	sim, ret := retSimOptions(t, "", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: arcaniteReaper}, 1,
		proto.MobType_MobTypeUnknown, &proto.PaladinOptions{PrimarySeal: proto.PaladinSeal_Righteousness, Aura: proto.PaladinAura_RetributionAura})
	aura := ret.GetSpell(core.ActionID{SpellID: 10301})
	if aura == nil || aura.BonusCoefficient <= 0 {
		t.Fatal("no Retribution Aura rank 5 with a spell power coefficient")
	}
	holy := ret.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 9066},
		SpellSchool:      core.SpellSchoolHoly,
		DefenseType:      core.DefenseTypeMagic,
		ProcMask:         core.ProcMaskSpellDamage,
		DamageMultiplier: 1,
		BonusCoefficient: aura.BonusCoefficient,
	})
	target := ret.CurrentTarget
	damage := func(spell *core.Spell) float64 {
		return spell.CalcDamage(sim, target, 20, spell.OutcomeAlwaysHit).Damage
	}
	auraBefore, holyBefore := damage(aura), damage(holy)
	target.PseudoStats.SchoolBonusDamageTaken[stats.SchoolIndexHoly] += 140
	if got := damage(aura); got != auraBefore {
		t.Errorf("Retribution Aura deals %.2f under a 140 Holy bonus, want %.2f as without", got, auraBefore)
	}
	if got, want := damage(holy), holyBefore+140*aura.BonusCoefficient; got < want-1e-6 || got > want+1e-6 {
		t.Errorf("a Holy spell with the same coefficient deals %.2f under a 140 Holy bonus, want %.2f", got, want)
	}
}

// TestVengeanceSkipsWarblade checks that a crit that reaches none of the wearer's procs, Warblade
// of Caer Darrow's on-hit crit (sim/common's TestForeverWarbladeSuppressesCasterProcs), gives no
// Vengeance, where another crit does (EV-0165).
func TestVengeanceSkipsWarblade(t *testing.T) {
	sim, ret := retSim(t, "--000000000000003", arcaniteReaper)
	proc := ret.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 1298499},
		SpellSchool: core.SpellSchoolFrost,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeSpecial,
		Flags:       core.SpellFlagPassiveSpell | core.SpellFlagNoProcs,
	})
	vengeance, stacks := ret.GetAura("Vengeance"), ret.GetAura("Vengeance Proc")
	crit := func(spell *core.Spell) {
		result := spell.NewResult(ret.CurrentTarget)
		result.Outcome = core.OutcomeCrit
		vengeance.OnSpellHitDealt(vengeance, sim, spell, result)
	}
	crit(proc)
	if stacks.IsActive() {
		t.Error("a crit with Suppress Caster Procs gave Vengeance")
	}
	crit(ret.AutoAttacks.MHAuto())
	if !stacks.IsActive() {
		t.Error("a white crit gave no Vengeance")
	}
}
