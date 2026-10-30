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
	// Party 1's Enhancement shaman casts its own Mana Spring there, so nobody in it gets one given; party 2
	// has no shaman.
	if war.RaidBuffs.ManaSpringTotem != 0 || mage.RaidBuffs.ManaSpringTotem != 0 {
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
	// With no warrior tank, a DPS warrior holds the biggest job it has a build for: Sunder, before Battle Shout.
	comp2 := &proto.RaidComp{Size: 5, Slots: []*proto.CompSlot{slotOf(1, 0, 0, p, "warrior__fury"), slotOf(2, 0, 1, p, "mage__fire")}}
	res2 := Resolve(comp2)
	if res2.Error != nil {
		t.Fatal(res2.Error.Message)
	}
	sunders := false
	for _, a := range res2.Assignments {
		if d, ok := a.Job.(*proto.Assignment_Duty); ok && d.Duty == proto.Duty_DutySunderUpkeep && a.Holder == 1 {
			sunders = true
		}
	}
	if !sunders || res2.Request.Raid.Debuffs.SunderRamp != nil {
		t.Errorf("the warrior should keep Sunder up itself, with no given ramp: %v", res2.Assignments)
	}
}

func dutyOf(res *proto.RaidCompResolution, holder int32) proto.Duty {
	for _, a := range res.Assignments {
		if d, ok := a.Job.(*proto.Assignment_Duty); ok && a.Holder == holder {
			return d.Duty
		}
	}
	return proto.Duty_DutyNone
}

// The first audit's resolver findings, each as a case.
func TestResolveAssignments(t *testing.T) {
	p := loadTestPool(t)
	hand := func(holder int32, a *proto.Assignment) *proto.Assignment {
		a.Holder, a.SetByHand = holder, true
		return a
	}
	// A shaman placeholder set by hand to Grace of Air only doesn't take the other's Strength of Earth.
	res := Resolve(&proto.RaidComp{Size: 10, Slots: []*proto.CompSlot{
		phSlot(1, 0, 0, proto.RaidPreset_RaidPresetRestorationShaman), phSlot(2, 1, 0, proto.RaidPreset_RaidPresetWardenShaman),
		slotOf(3, 0, 1, p, "warrior__fury")},
		Assignments: []*proto.Assignment{hand(2, &proto.Assignment{Job: &proto.Assignment_Totem{Totem: &proto.TotemAssignment{Totem: proto.RaidTotem_RaidTotemGraceOfAir}}})}})
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	if res.Request.Raid.Parties[0].Players[0].RaidBuffs.StrengthOfEarthTotem == 0 {
		t.Errorf("the Restoration shaman's Strength of Earth went missing")
	}
	// A paladin placeholder set by hand to Wisdom: the automatic one takes the Crusader.
	res = Resolve(&proto.RaidComp{Size: 10, Slots: []*proto.CompSlot{
		phSlot(1, 0, 0, proto.RaidPreset_RaidPresetHolyPaladin), phSlot(2, 0, 1, proto.RaidPreset_RaidPresetProtectionPaladin),
		slotOf(3, 0, 2, p, "mage__fire")},
		Assignments: []*proto.Assignment{hand(1, &proto.Assignment{Job: &proto.Assignment_Judgement{Judgement: proto.PaladinJudgement_PaladinJudgementWisdom}})}})
	if d := res.Request.Raid.Debuffs; !d.JudgementOfWisdom || d.JudgementOfTheCrusader == 0 {
		t.Errorf("want Wisdom and the Crusader: %v", d)
	}
	// A Ret member set by hand to the Crusader runs its Crusader duty build.
	res = Resolve(&proto.RaidComp{Size: 5, Slots: []*proto.CompSlot{slotOf(1, 0, 0, p, "paladin__level_60_sub_holy")},
		Assignments: []*proto.Assignment{hand(1, &proto.Assignment{Job: &proto.Assignment_Judgement{Judgement: proto.PaladinJudgement_PaladinJudgementCrusader}})}})
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	cast := false
	for _, b := range res.Buffs {
		if b.Effect == "Judgement of the Crusader" && !b.Given {
			cast = true
		}
	}
	if !cast || res.Request.Raid.Debuffs.JudgementOfTheCrusader != 0 {
		t.Errorf("the Ret should cast the Crusader through its duty build: %v", res.Buffs)
	}
	// Power Infusion goes to a caster, not a hunter; its holder is the priest.
	res = Resolve(&proto.RaidComp{Size: 10, Slots: []*proto.CompSlot{
		phSlot(1, 0, 0, proto.RaidPreset_RaidPresetDisciplinePriest), slotOf(2, 0, 1, p, "hunter__marksmanship"), slotOf(3, 1, 0, p, "mage__fire")}})
	for _, a := range res.Assignments {
		if c, ok := a.Job.(*proto.Assignment_Cooldown); ok && c.Cooldown.PowerInfusion && (c.Cooldown.Target != 3 || a.Holder != 1) {
			t.Errorf("Power Infusion: holder %d, target %d", a.Holder, c.Cooldown.Target)
		}
	}
	// A Tallstrider hunter already fills the minor armor slot, so the Balance druid holds no Faerie Fire duty.
	res = Resolve(&proto.RaidComp{Size: 5, Slots: []*proto.CompSlot{slotOf(1, 0, 0, p, "hunter__beast-mastery-tallstrider"), slotOf(2, 0, 1, p, "druid__balance")}})
	if dutyOf(res, 2) != proto.Duty_DutyNone {
		t.Errorf("the druid shouldn't hold Faerie Fire next to a Tallstrider")
	}
	// One element set by hand (fire: none) leaves the other elements to the resolver.
	res = Resolve(&proto.RaidComp{Size: 10, Slots: []*proto.CompSlot{phSlot(1, 0, 0, proto.RaidPreset_RaidPresetRestorationShaman),
		slotOf(2, 0, 1, p, "warrior__fury"), slotOf(3, 0, 2, p, "rogue__combat-swords"), slotOf(4, 0, 3, p, "warrior__arms")},
		Assignments: []*proto.Assignment{hand(1, &proto.Assignment{Job: &proto.Assignment_Totem{Totem: &proto.TotemAssignment{Element: proto.TotemElement_TotemElementFire}}})}})
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	if rb := res.Request.Raid.Parties[0].Players[0].RaidBuffs; rb.StrengthOfEarthTotem == 0 || rb.TotemWeaponBuff != proto.TotemWeaponBuff_TotemWeaponBuffWindfury {
		t.Errorf("earth and air should stay automatic (Windfury for a melee raid): %v", rb)
	}
	// A member set by hand to no duty holds none, though it has a build for one.
	res = Resolve(&proto.RaidComp{Size: 5, Slots: []*proto.CompSlot{slotOf(1, 0, 0, p, "druid__balance"), slotOf(2, 0, 1, p, "warrior__fury")},
		Assignments: []*proto.Assignment{hand(1, &proto.Assignment{Job: &proto.Assignment_Duty{Duty: proto.Duty_DutyNone}})}})
	if res.Error != nil || dutyOf(res, 1) != proto.Duty_DutyNone || res.Request.Raid.Debuffs.FaerieFire {
		t.Errorf("the druid set to no duty: %v %v", res.Error, res.Assignments)
	}
	// A DPS shaman in a shaman placeholder's party casts its own Mana Spring: the party gets no given one
	// too, but keeps the placeholder's Mana Tide.
	res = Resolve(&proto.RaidComp{Size: 5, Slots: []*proto.CompSlot{phSlot(1, 0, 0, proto.RaidPreset_RaidPresetRestorationShaman),
		slotOf(2, 0, 1, p, "shaman__elemental-level60"), slotOf(3, 0, 2, p, "mage__fire")}})
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	if mage := res.Request.Raid.Parties[0].Players[1]; mage.RaidBuffs.ManaSpringTotem != 0 || res.Request.Raid.Parties[0].Buffs.ManaTideTotems != 1 {
		t.Errorf("Mana Spring given beside a cast one (%v), or Mana Tide lost", mage.RaidBuffs.ManaSpringTotem)
	}
	// One paladin placeholder judges Wisdom, so a Ret member with a Crusader build takes the Crusader.
	res = Resolve(&proto.RaidComp{Size: 5, Slots: []*proto.CompSlot{phSlot(1, 0, 0, proto.RaidPreset_RaidPresetHolyPaladin),
		slotOf(2, 0, 1, p, "paladin__level_60"), slotOf(3, 0, 2, p, "mage__fire")}})
	if d := res.Request.Raid.Debuffs; !d.JudgementOfWisdom || dutyOf(res, 2) != proto.Duty_DutyJudgementOfTheCrusader {
		t.Errorf("want Wisdom from the placeholder and the Crusader from the Ret: %v %v", d, res.Assignments)
	}
	// Two Discipline priests, one Power Infusion set by hand: the other priest's still goes out.
	res = Resolve(&proto.RaidComp{Size: 10, Slots: []*proto.CompSlot{phSlot(1, 0, 0, proto.RaidPreset_RaidPresetDisciplinePriest),
		phSlot(2, 0, 1, proto.RaidPreset_RaidPresetDisciplinePriest), slotOf(3, 0, 2, p, "mage__fire"), slotOf(4, 1, 0, p, "warlock__demonology")},
		Assignments: []*proto.Assignment{hand(1, &proto.Assignment{Job: &proto.Assignment_Cooldown{Cooldown: &proto.CooldownTarget{PowerInfusion: true, Target: 3}}})}})
	pis := 0
	for _, party := range res.Request.Raid.Parties {
		for _, pl := range party.Players {
			pis += int(pl.Buffs.PowerInfusions)
		}
	}
	if pis != 2 {
		t.Errorf("two Power Infusions, one by hand: got %d", pis)
	}
	// The automatic one is the other priest's, not the hand-set one's again.
	for _, a := range res.Assignments {
		if cd, ok := a.Job.(*proto.Assignment_Cooldown); ok && !a.SetByHand && cd.Cooldown.PowerInfusion && a.Holder != 2 {
			t.Errorf("the automatic Power Infusion is held by slot %d, want 2", a.Holder)
		}
	}
	// An assignment not set by hand is the resolver's to redo, and a cooldown from someone without it is refused.
	res = Resolve(&proto.RaidComp{Size: 5, Slots: []*proto.CompSlot{slotOf(1, 0, 0, p, "mage__fire")},
		Assignments: []*proto.Assignment{hand(1, &proto.Assignment{Job: &proto.Assignment_Cooldown{Cooldown: &proto.CooldownTarget{PowerInfusion: true, Target: 1}}})}})
	if res.Error == nil {
		t.Errorf("a mage can't give Power Infusion")
	}
}

// The confirmation raid sims the variant it scored: a Tallstrider hunter keeps its Tallstrider.
func TestCompOfKeepsVariants(t *testing.T) {
	p := loadTestPool(t)
	sim, _ := NewSimmer(p, "")
	sc := &Scenario{Pool: p, Sim: sim, Faction: proto.Faction_Horde, Fight: Fight{Duration: 60, FlaskZone: true}}
	b := p.Builds["hunter__beast-mastery-tallstrider"]
	r := NewRaid(5, []*Preset{PresetByID(proto.RaidPreset_RaidPresetProtectionWarrior)}, nil)
	c := &Comp{Raid: r, PHParty: []int{0}, Plan: Plan{Up: r.phUp(Plan{}) | b.Gives}, Members: []*Member{{B: b}}}
	rc, err := sc.compOf(c, 10, 1)
	if err != nil {
		t.Fatal(err)
	}
	res := Resolve(rc)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	pl := res.Request.Raid.Parties[0].Players[0]
	if pl.GetHunter().GetOptions().GetPetType() != proto.Hunter_Options_Tallstrider || rc.Slots[1].ReportId != b.ID {
		t.Errorf("the confirmation runs %v as %s", pl.GetHunter().GetOptions().GetPetType(), rc.Slots[1].ReportId)
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
	out, err := compValues(p, sim, comp, nil)
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

// Check 7's "nobody provides it": a lone mage gets no buff or debuff anyone else would give.
func TestResolveNobodyProvides(t *testing.T) {
	p := loadTestPool(t)
	res := Resolve(&proto.RaidComp{Size: 5, Slots: []*proto.CompSlot{slotOf(1, 0, 0, p, "mage__fire")}})
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	rb, db := res.Request.Raid.Parties[0].Players[0].RaidBuffs, res.Request.Raid.Debuffs
	if rb.BattleShout != 0 || rb.FuriousHowl || rb.MoonkinAura || rb.LeaderOfThePack || rb.TrueshotAura || rb.DivineSpirit ||
		rb.GiftOfTheWild != 0 || rb.StrengthOfEarthTotem != 0 || rb.GraceOfAirTotem != 0 || rb.TotemWeaponBuff != 0 || rb.ManaSpringTotem != 0 {
		t.Errorf("buffs nobody gives: %v", rb)
	}
	if db.JudgementOfWisdom || db.JudgementOfTheCrusader != 0 || db.CurseOfElements || db.FaerieFire || db.SunderRamp != nil || db.SunderArmor {
		t.Errorf("debuffs nobody gives: %v", db)
	}
	if b := res.Request.Raid.Parties[0].Players[0].Buffs; b.BlessingOfKings || b.BlessingOfMight != 0 || b.BlessingOfWisdom != 0 || b.PowerInfusions != 0 {
		t.Errorf("blessings or cooldowns nobody gives: %v", b)
	}
	// Its own Arcane Brilliance is given back to it, since the hooks are skipped.
	if !rb.ArcaneBrilliance {
		t.Errorf("the mage's own Arcane Brilliance")
	}
	for _, want := range []string{"Judgement of Wisdom", "Curse of the Elements", "Battle Shout", "Sunder upkeep"} {
		if !slices.ContainsFunc(res.Misses, func(s string) bool { return strings.Contains(s, want) }) {
			t.Errorf("the misses should name %s: %v", want, res.Misses)
		}
	}
}

// Member values see a hand-built raid's parties as the resolver gives them: beside a DPS shaman, a
// Restoration placeholder's party gets the DPS shaman's cast Mana Spring and the placeholder's Mana Tide.
func TestCompValuesSeatsMatchTheResolver(t *testing.T) {
	p := loadTestPool(t)
	comp := &proto.RaidComp{Size: 5, Slots: []*proto.CompSlot{phSlot(1, 0, 0, proto.RaidPreset_RaidPresetRestorationShaman),
		slotOf(2, 0, 1, p, "shaman__elemental-level60"), slotOf(3, 0, 2, p, "mage__fire")}}
	sc := &Scenario{Pool: p, Faction: proto.Faction_Horde, Fight: Fight{Duration: 180}}
	c, ids, err := compFromRaidComp(p, sc, comp)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range c.Members {
		if ids[m] == 3 && m.Seat.Package != proto.RaidPackage_RaidPackageManaSpringTide {
			t.Errorf("the mage's package is %v, want Mana Spring and Mana Tide", m.Seat.Package)
		}
	}
}
