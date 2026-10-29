package compsearch

import (
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

func slotOf(id, party, index int32, p *Pool, buildID string) *proto.CompSlot {
	b := p.Builds[buildID]
	mb := &proto.CompMemberBuild{Base: player0(cloneReq(b.Req))}
	for d, v := range b.Duties {
		mb.Duties = append(mb.Duties, &proto.CompDutyPlayer{Duty: d, Player: player0(cloneReq(v.Req))})
	}
	return &proto.CompSlot{Id: id, Party: party, Index: index, Member: &proto.CompSlot_Build{Build: mb}, ReportId: buildID}
}

func phSlot(id, party, index int32, preset proto.RaidPreset) *proto.CompSlot {
	return &proto.CompSlot{Id: id, Party: party, Index: index, Member: &proto.CompSlot_Placeholder{Placeholder: preset}}
}

// Check 7 (docs/raid_sim.md, Checking it): the resolver, with no sims.
func TestResolve(t *testing.T) {
	p := loadTestPool(t)
	comp := &proto.RaidComp{Size: 10, CampBuffs: true, Slots: []*proto.CompSlot{
		phSlot(1, 0, 0, proto.RaidPreset_RaidPresetProtectionWarrior),
		phSlot(2, 0, 1, proto.RaidPreset_RaidPresetRestorationShaman),
		slotOf(3, 0, 3, p, "warrior__fury"), // a gap at index 2
		phSlot(4, 1, 0, proto.RaidPreset_RaidPresetHolyPaladin),
		phSlot(5, 1, 1, proto.RaidPreset_RaidPresetProtectionPaladin),
		slotOf(6, 1, 2, p, "mage__fire"),
		slotOf(7, 1, 4, p, "hunter__beast-mastery"),
		slotOf(8, 0, 4, p, "shaman__enhancement-level60"),
	}}
	res := Resolve(comp)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	if !slices.Equal(res.SlotOfPlayer, []int32{3, 8, 6, 7}) {
		t.Errorf("slot of each player: %v", res.SlotOfPlayer)
	}
	players := map[int32]*proto.Player{}
	i := 0
	for _, party := range res.Request.Raid.Parties {
		for _, pl := range party.Players {
			players[res.SlotOfPlayer[i]] = pl
			i++
		}
	}
	war, mage, hunter, enh := players[3], players[6], players[7], players[8]
	// Two paladins: each class gets the 2 blessings it wants first, one from each paladin.
	if !war.Buffs.BlessingOfKings || war.Buffs.BlessingOfMight == 0 || war.Buffs.BlessingOfWisdom != 0 {
		t.Errorf("the warrior's blessings: %v", war.Buffs)
	}
	if !mage.Buffs.BlessingOfKings || mage.Buffs.BlessingOfWisdom == 0 || mage.Buffs.BlessingOfMight != 0 {
		t.Errorf("the mage's blessings: %v", mage.Buffs)
	}
	// The warrior tank gives Battle Shout and keeps Sunder up; nobody gives the minor armor slot.
	if war.RaidBuffs.BattleShout == 0 || res.Request.Raid.Debuffs.SunderRamp == nil || !res.Request.Raid.Debuffs.SunderRamp.Tank {
		t.Errorf("the warrior tank's Battle Shout and Sunder: %v, %v", war.RaidBuffs.BattleShout, res.Request.Raid.Debuffs.SunderRamp)
	}
	if res.Request.Raid.Debuffs.FaerieFire || !slices.ContainsFunc(res.Misses, func(s string) bool { return strings.Contains(s, "Minor armor") }) {
		t.Errorf("the minor armor slot should be missing: %v", res.Misses)
	}
	// The Enhancement shaman drops Strength of Earth and Windfury itself, so nobody gets them given; the
	// shaman placeholder's Mana Spring goes to its own party only.
	for _, pl := range []*proto.Player{war, mage, hunter, enh} {
		if pl.RaidBuffs.StrengthOfEarthTotem != 0 || pl.RaidBuffs.TotemWeaponBuff == proto.TotemWeaponBuff_TotemWeaponBuffWindfury {
			t.Errorf("%s got a cast totem given: %v", pl.Name, pl.RaidBuffs)
		}
	}
	if war.RaidBuffs.ManaSpringTotem != proto.TristateEffect_TristateEffectImproved || mage.RaidBuffs.ManaSpringTotem != 0 {
		t.Errorf("Mana Spring: party 1 %v, party 2 %v", war.RaidBuffs.ManaSpringTotem, mage.RaidBuffs.ManaSpringTotem)
	}
	if res.Request.Raid.Parties[0].Buffs.ManaTideTotems != 1 || res.Request.Raid.Parties[1].Buffs.ManaTideTotems != 0 {
		t.Errorf("Mana Tide should reach only the shaman's party")
	}
	// Flametongue goes to Enhancement, whose Windfury Weapon makes Windfury Totem useless to it.
	if enh.RaidBuffs.TotemWeaponBuff != proto.TotemWeaponBuff_TotemWeaponBuffFlametongue {
		t.Errorf("Enhancement's weapon totem: %v", enh.RaidBuffs.TotemWeaponBuff)
	}
	// The mage's Arcane Brilliance is given to everyone, the camp buffs to everyone.
	if !war.RaidBuffs.ArcaneBrilliance || war.RaidBuffs.CampBuffs == nil || !war.RaidBuffs.CampBuffs.FishBowl {
		t.Errorf("Arcane Brilliance and camps: %v", war.RaidBuffs)
	}
	// One judgement per paladin: Wisdom and the Crusader.
	if !res.Request.Raid.Debuffs.JudgementOfWisdom || res.Request.Raid.Debuffs.JudgementOfTheCrusader == 0 {
		t.Errorf("judgements: %v", res.Request.Raid.Debuffs)
	}

	// Set by hand: a duty the member has no build for is refused, and two slots in one place too.
	bad := &proto.RaidComp{Size: 10, Slots: []*proto.CompSlot{slotOf(1, 0, 0, p, "mage__fire")},
		Assignments: []*proto.Assignment{{Holder: 1, SetByHand: true, Job: &proto.Assignment_Duty{Duty: proto.Duty_DutyBattleShout}}}}
	if Resolve(bad).Error == nil {
		t.Errorf("a mage can't hold Battle Shout")
	}
	bad = &proto.RaidComp{Size: 10, Slots: []*proto.CompSlot{slotOf(1, 0, 0, p, "mage__fire"), slotOf(2, 0, 0, p, "mage__fire")}}
	if Resolve(bad).Error == nil {
		t.Errorf("two slots in one place")
	}
	// With no warrior tank, a DPS warrior with a Battle Shout build holds it.
	comp2 := &proto.RaidComp{Size: 5, Slots: []*proto.CompSlot{slotOf(1, 0, 0, p, "warrior__fury"), slotOf(2, 0, 1, p, "mage__fire")}}
	res2 := Resolve(comp2)
	if res2.Error != nil {
		t.Fatal(res2.Error.Message)
	}
	shouts := false
	for _, a := range res2.Assignments {
		if d, ok := a.Job.(*proto.Assignment_Duty); ok && d.Duty == proto.Duty_DutyBattleShout && a.Holder == 1 {
			shouts = true
		}
	}
	if !shouts || res2.Request.Raid.Parties[0].Players[1].RaidBuffs.BattleShout != 0 {
		t.Errorf("the warrior should cast Battle Shout, and nobody get it given")
	}
}

// M1b's done-when: a comp round-trips through JSON.
func TestCompRoundTrip(t *testing.T) {
	p := loadTestPool(t)
	comp := &proto.RaidComp{Size: 10, Faction: proto.Faction_Alliance, CampBuffs: true,
		Assumptions: &proto.RaidAssumptions{CampChairSpellCritOnly: true},
		Slots:       []*proto.CompSlot{phSlot(1, 0, 0, proto.RaidPreset_RaidPresetFeralTank), slotOf(2, 0, 1, p, "warlock__demonology")},
		Assignments: []*proto.Assignment{{Holder: 2, SetByHand: true, Job: &proto.Assignment_Duty{Duty: proto.Duty_DutyCurseOfTheElements}},
			{Holder: 1, Job: &proto.Assignment_Cooldown{Cooldown: &proto.CooldownTarget{Target: 2}}}}}
	data, err := protojson.Marshal(comp)
	if err != nil {
		t.Fatal(err)
	}
	back := &proto.RaidComp{}
	if err := protojson.Unmarshal(data, back); err != nil {
		t.Fatal(err)
	}
	if !googleProto.Equal(comp, back) {
		t.Errorf("the comp changed through JSON")
	}
}

// A hand-built raid's member values: every distinct member gets one, with its slots, and a duty holder
// its duty's cost.
func TestCompValues(t *testing.T) {
	if testing.Short() {
		t.Skip("about 30 s")
	}
	p := loadTestPool(t)
	sim, _ := NewSimmer(p, "")
	comp := &proto.RaidComp{Size: 5, SimOptions: &proto.SimOptions{Iterations: 60},
		Encounter: &proto.Encounter{Duration: 90, NaturalFlaskZone: true},
		Slots: []*proto.CompSlot{phSlot(1, 0, 0, proto.RaidPreset_RaidPresetProtectionWarrior),
			slotOf(2, 0, 1, p, "mage__fire"), slotOf(3, 0, 2, p, "mage__fire"), slotOf(4, 0, 3, p, "warlock__demonology"),
			slotOf(5, 0, 4, p, "hunter__beast-mastery")},
		Assignments: []*proto.Assignment{{Holder: 4, SetByHand: true, Job: &proto.Assignment_Duty{Duty: proto.Duty_DutyCurseOfTheElements}}}}
	out, err := compValues(p, sim, comp)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.MemberValues) != 3 || out.Score <= 0 {
		t.Fatalf("want 3 distinct members, got %d (score %.1f)", len(out.MemberValues), out.Score)
	}
	for _, mv := range out.MemberValues {
		switch mv.BuildId {
		case "mage__fire":
			if len(mv.Slots) != 2 {
				t.Errorf("the two mages share one value: %v", mv.Slots)
			}
		case "warlock__demonology-duty-curse":
			if mv.DutyCost <= 0 || mv.Contribution <= mv.OwnDps {
				t.Errorf("the curse warlock: duty cost %.1f, contribution %.1f, own %.1f", mv.DutyCost, mv.Contribution, mv.OwnDps)
			}
		}
	}
}
