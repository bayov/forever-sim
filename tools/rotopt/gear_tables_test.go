package main

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Every class with a DPS spec in the report can hold a weapon and something in the ranged slot.
func TestGearTablesCoverEveryClass(t *testing.T) {
	classes := []proto.Class{proto.Class_ClassWarrior, proto.Class_ClassPaladin, proto.Class_ClassHunter, proto.Class_ClassRogue,
		proto.Class_ClassPriest, proto.Class_ClassShaman, proto.Class_ClassMage, proto.Class_ClassWarlock, proto.Class_ClassDruid}
	for _, class := range classes {
		if len(weaponTypesFor[class]) == 0 {
			t.Errorf("%v has no weapon types", class)
		}
		if len(rangedTypesFor[class]) == 0 {
			t.Errorf("%v has no ranged types", class)
		}
	}
	ranged := map[proto.Class]proto.RangedWeaponType{
		proto.Class_ClassHunter:  proto.RangedWeaponType_RangedWeaponTypeBow,
		proto.Class_ClassDruid:   proto.RangedWeaponType_RangedWeaponTypeIdol,
		proto.Class_ClassMage:    proto.RangedWeaponType_RangedWeaponTypeWand,
		proto.Class_ClassPriest:  proto.RangedWeaponType_RangedWeaponTypeWand,
		proto.Class_ClassWarlock: proto.RangedWeaponType_RangedWeaponTypeWand,
		proto.Class_ClassPaladin: proto.RangedWeaponType_RangedWeaponTypeLibram,
		proto.Class_ClassShaman:  proto.RangedWeaponType_RangedWeaponTypeTotem,
	}
	for class, want := range ranged {
		if !slices.Contains(rangedTypesFor[class], want) {
			t.Errorf("%v can't hold %v", class, want)
		}
	}
	for _, c := range []struct {
		class proto.Class
		can   []proto.WeaponType
		cant  []proto.WeaponType
	}{
		{proto.Class_ClassHunter, []proto.WeaponType{proto.WeaponType_WeaponTypePolearm, proto.WeaponType_WeaponTypeAxe}, []proto.WeaponType{proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeShield}},
		{proto.Class_ClassDruid, []proto.WeaponType{proto.WeaponType_WeaponTypeStaff, proto.WeaponType_WeaponTypeMace}, []proto.WeaponType{proto.WeaponType_WeaponTypeSword, proto.WeaponType_WeaponTypeAxe}},
		{proto.Class_ClassMage, []proto.WeaponType{proto.WeaponType_WeaponTypeStaff, proto.WeaponType_WeaponTypeSword, proto.WeaponType_WeaponTypeOffHand}, []proto.WeaponType{proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeShield}},
		{proto.Class_ClassPriest, []proto.WeaponType{proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeOffHand}, []proto.WeaponType{proto.WeaponType_WeaponTypeSword}},
		{proto.Class_ClassWarlock, []proto.WeaponType{proto.WeaponType_WeaponTypeSword, proto.WeaponType_WeaponTypeDagger}, []proto.WeaponType{proto.WeaponType_WeaponTypeMace}},
	} {
		for _, w := range c.can {
			if !slices.Contains(weaponTypesFor[c.class], w) {
				t.Errorf("%v can't hold %v", c.class, w)
			}
		}
		for _, w := range c.cant {
			if slices.Contains(weaponTypesFor[c.class], w) {
				t.Errorf("%v holds %v", c.class, w)
			}
		}
	}
}

func suffixOf(stat stats.Stat) *proto.ItemRandomSuffix {
	st := make([]float64, stats.Len)
	st[stat] = 10
	return &proto.ItemRandomSuffix{Stats: st}
}

func TestUsefulSuffix(t *testing.T) {
	for _, c := range []struct {
		class  proto.Class
		stat   stats.Stat
		useful bool
	}{
		{proto.Class_ClassMage, stats.FirePower, true},
		{proto.Class_ClassMage, stats.FrostPower, true},
		{proto.Class_ClassMage, stats.ShadowPower, false},
		{proto.Class_ClassMage, stats.Strength, false},
		{proto.Class_ClassMage, stats.Intellect, true},
		{proto.Class_ClassWarlock, stats.ShadowPower, true},
		{proto.Class_ClassWarlock, stats.FrostPower, false},
		{proto.Class_ClassPriest, stats.ShadowPower, true},
		{proto.Class_ClassPriest, stats.SpellDamage, true},
		{proto.Class_ClassPriest, stats.Agility, false},
		{proto.Class_ClassDruid, stats.Agility, true},
		{proto.Class_ClassDruid, stats.ArcanePower, true},
		{proto.Class_ClassDruid, stats.FirePower, false},
		{proto.Class_ClassHunter, stats.RangedAttackPower, true},
		{proto.Class_ClassHunter, stats.Intellect, false},
		{proto.Class_ClassRogue, stats.Agility, true},
		{proto.Class_ClassRogue, stats.SpellPower, false},
		// The shaman and paladin pools are as before the casters joined: Enhancement's list
		// must not move.
		{proto.Class_ClassShaman, stats.NaturePower, true},
		{proto.Class_ClassShaman, stats.Strength, true},
		{proto.Class_ClassShaman, stats.ShadowPower, false},
		{proto.Class_ClassShaman, stats.SpellDamage, false},
		{proto.Class_ClassPaladin, stats.FrostPower, false},
	} {
		if got := usefulSuffix(c.class, suffixOf(c.stat)); got != c.useful {
			t.Errorf("usefulSuffix(%v, %v) = %v, want %v", c.class, c.stat, got, c.useful)
		}
	}
}

func TestApplyWeaponRules(t *testing.T) {
	items := map[int32]proto.WeaponType{
		1: proto.WeaponType_WeaponTypeSword, 2: proto.WeaponType_WeaponTypeDagger, 3: proto.WeaponType_WeaponTypeStaff,
		4: proto.WeaponType_WeaponTypeShield, 5: proto.WeaponType_WeaponTypeOffHand,
	}
	for id, w := range items {
		itemsByID[id] = &proto.UIItem{Id: id, WeaponType: w}
	}
	defer func() {
		for id := range items {
			delete(itemsByID, id)
		}
	}()
	option := func(id int32) gearOption {
		return gearOption{spec: &proto.ItemSpec{Id: id}, twoHand: id == 3}
	}
	ids := func(options []gearOption) []int32 {
		var out []int32
		for _, o := range options {
			out = append(out, o.spec.Id)
		}
		return out
	}
	for _, c := range []struct {
		rule           string
		mhTypes        []proto.WeaponType
		ohTypes        []proto.WeaponType
		wantMH, wantOH []int32
	}{
		{"any", nil, nil, []int32{1, 2, 3}, []int32{1, 2, 4, 5}},
		{"two-hand", nil, nil, []int32{3}, nil},
		{"dual-wield", nil, nil, []int32{1, 2}, []int32{1, 2}},
		{"one-hand", nil, nil, []int32{1, 2}, []int32{4, 5}},
		{"dual-wield", []proto.WeaponType{proto.WeaponType_WeaponTypeDagger}, []proto.WeaponType{proto.WeaponType_WeaponTypeDagger}, []int32{2}, []int32{2}},
	} {
		pool := make([][]gearOption, len(gearSlotNames))
		pool[proto.ItemSlot_ItemSlotMainHand] = []gearOption{option(1), option(2), option(3)}
		pool[proto.ItemSlot_ItemSlotOffHand] = []gearOption{option(1), option(2), option(4), option(5)}
		if err := applyWeaponRules(pool, c.rule, c.mhTypes, c.ohTypes); err != nil {
			t.Fatal(err)
		}
		if got := ids(pool[proto.ItemSlot_ItemSlotMainHand]); !slices.Equal(got, c.wantMH) {
			t.Errorf("%s %v: main hand %v, want %v", c.rule, c.mhTypes, got, c.wantMH)
		}
		if got := ids(pool[proto.ItemSlot_ItemSlotOffHand]); !slices.Equal(got, c.wantOH) {
			t.Errorf("%s %v: off hand %v, want %v", c.rule, c.ohTypes, got, c.wantOH)
		}
	}
	if err := applyWeaponRules(make([][]gearOption, len(gearSlotNames)), "sword-and-board", nil, nil); err == nil {
		t.Error("an unknown rule went through")
	}
}

// The spec table names a report build (a tank or healer spec its bench preset), a weapon setup and weapon types that exist for every
// spec, and two specs never write the same list.
func TestPhase0SpecTable(t *testing.T) {
	data, err := os.ReadFile("phase0_specs.json")
	if err != nil {
		t.Fatal(err)
	}
	var table struct {
		Specs map[string]struct {
			Req, ReqFile, RaidReqFile, Bench, Weapons, MhTypes, OhTypes, Dir, Out string
			Start                                    []any
		}
	}
	if err := json.Unmarshal(data, &table); err != nil {
		t.Fatal(err)
	}
	outs := map[string]string{}
	for name, s := range table.Specs {
		req := "../dps_report/req/" + s.Req + ".json"
		if s.Bench != "" {
			req = "../../" + s.ReqFile
			if s.Bench != "tank" && s.Bench != "heal" {
				t.Errorf("%s: bench %q", name, s.Bench)
			}
		}
		if _, err := os.Stat(req); err != nil {
			t.Errorf("%s: %v", name, err)
		}
		if s.RaidReqFile != "" {
			if _, err := os.Stat("../../" + s.RaidReqFile); err != nil {
				t.Errorf("%s: %v", name, err)
			}
		}
		if !slices.Contains(weaponRules, s.Weapons) {
			t.Errorf("%s: weapons %q", name, s.Weapons)
		}
		if _, err := parseWeaponTypes(s.MhTypes + "," + s.OhTypes); err != nil {
			t.Errorf("%s: %v", name, err)
		}
		if s.Start != nil && len(s.Start) != len(gearSlotNames) {
			t.Errorf("%s: %d starting items, want %d", name, len(s.Start), len(gearSlotNames))
		}
		out := s.Out
		if out == "" {
			out = "ui/" + s.Dir + "/bis/level60_phase0_" + name + ".bis.json"
		}
		if other, ok := outs[out]; ok {
			t.Errorf("%s and %s both write %s", name, other, out)
		}
		outs[out] = name
	}
}
