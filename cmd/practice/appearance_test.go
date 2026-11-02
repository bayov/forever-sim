//go:build practice

package main

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
)

// testAppearance reads the cached wago.tools tables for the client's build, and the model files
// already extracted into ~/forever-practice/assets. It skips when the tables aren't there, since
// the test must not fetch them.
func testAppearance(t *testing.T) *appearanceResolver {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip(err)
	}
	const build = "1.60.1.70707"
	dir := filepath.Join(home, "forever-practice", "db2", build)
	for _, name := range []string{
		"ChrRaceXChrModel", "ChrModel", "CreatureDisplayInfo", "CreatureModelData", "ChrCustomizationOption",
		"ChrCustomizationChoice", "ChrCustomizationElement", "ChrCustomizationGeoset", "ChrCustomizationMaterial",
		"ChrCustomizationReq", "ChrCustomizationSkinnedModel", "ChrModelMaterial", "ChrModelTextureLayer",
		"CharComponentTextureSections", "TextureFileData", "ModelFileData", "ComponentTextureFileData",
		"ComponentModelFileData", "ItemModifiedAppearance", "ItemAppearance", "ItemDisplayInfo",
		"ItemDisplayInfoMaterialRes", "ItemDisplayInfoModelMatRes", "HelmetGeosetData", "Item",
	} {
		if _, err := os.Stat(filepath.Join(dir, name+".csv")); err != nil {
			t.Skipf("no cached %s table in %s", name, dir)
		}
	}
	assets := filepath.Join(home, "forever-practice", "assets")
	return newAppearanceResolver(newDB2Tables(filepath.Join(home, "forever-practice"), build), func(fdid int) ([]byte, error) {
		return os.ReadFile(filepath.Join(assets, strconv.Itoa(fdid)))
	})
}

// TestOrcShaman dresses an Orc male shaman in pieces of the Enhancement phase 1 set.
func TestOrcShaman(t *testing.T) {
	r := testAppearance(t)
	a, err := r.character(characterRequest{Race: 2, Sex: 0, Class: 7, Items: map[int]int{
		slotHead:     18817, // Crown of Destruction
		slotShoulder: 12927, // Truestrike Shoulders
		slotBack:     13340, // Cape of the Black Baron
		slotChest:    11726, // Savage Gladiator Chain
		slotHands:    15063, // Devilsaur Gauntlets
		slotLegs:     15062, // Devilsaur Leggings
		slotFeet:     13210, // Pads of the Dread Wolf
		slotMainHand: 17182, // Sulfuras, Hand of Ragnaros
	}})
	if err != nil {
		t.Fatal(err)
	}
	if a.Model != 917116 || a.ChrModelID != 3 {
		t.Errorf("model %d chrModel %d, want orcmale_hd.m2 (917116) and ChrModel 3", a.Model, a.ChrModelID)
	}
	// 0 is the body. The gloves, boots and cape pick 401, 504 (GeosetGroup 3) and 1502.
	want := []int{0, 401, 504, 1502}
	if !a.GeosetsExact {
		// Without the extracted skin the list has every candidate id, so only check these.
		want = []int{0, 401, 504, 1502, 2002}
	}
	for _, id := range want {
		if !slices.Contains(a.Geosets, id) {
			t.Errorf("geosets %v miss %d", a.Geosets, id)
		}
	}
	for _, id := range []int{501, 1501} {
		if slices.Contains(a.Geosets, id) {
			t.Errorf("geosets %v still show %d under the boots and cape", a.Geosets, id)
		}
	}
	skin := a.Textures[1]
	if skin == nil || skin.Width != 2048 || skin.Height != 1024 || len(skin.Layers) < 5 {
		t.Fatalf("skin texture %+v, want a 2048x1024 texture with base, face and item layers", skin)
	}
	if skin.Layers[0].Section != -1 {
		t.Errorf("first skin layer %+v, want the whole-texture base skin", skin.Layers[0])
	}
	for i := 1; i < len(skin.Layers); i++ {
		if skin.Layers[i].Order < skin.Layers[i-1].Order {
			t.Errorf("skin layers out of order at %d: %+v", i, skin.Layers)
		}
	}
	hasLayer := func(fdid int) bool {
		for _, l := range skin.Layers {
			if l.Fdid == fdid {
				return true
			}
		}
		return false
	}
	// The chest's upper torso texture for males, and the gloves' hand texture.
	for _, f := range []int{163766, 154007} {
		if !hasLayer(f) {
			t.Errorf("no skin layer with %d", f)
		}
	}
	if a.Replaceable[2] != 137376 {
		t.Errorf("cape texture %d, want 137376 on texture type 2", a.Replaceable[2])
	}
	var weapon, helm *attachedModel
	shoulders := 0
	for i, at := range a.Attachments {
		switch at.Slot {
		case slotMainHand:
			weapon = &a.Attachments[i]
		case slotHead:
			helm = &a.Attachments[i]
		case slotShoulder:
			shoulders++
		}
	}
	if weapon == nil || weapon.Attachment != attachHandRight || weapon.Model != 145809 ||
		weapon.SheathedAttachment == nil || *weapon.SheathedAttachment != attachSheathMain || len(weapon.Textures) == 0 {
		t.Errorf("weapon %+v, want Sulfuras (145809) in the right hand, sheathed on the back", weapon)
	}
	if helm == nil || helm.Model != 139154 {
		t.Errorf("helm %+v, want the Orc male model 139154", helm)
	}
	if shoulders != 2 {
		t.Errorf("%d shoulder models, want left and right", shoulders)
	}
	var options []int
	for _, o := range a.Options {
		options = append(options, o.ID)
		if o.ID == 19 && (o.Chosen == 0 || len(o.Choices) == 0) {
			t.Errorf("skin color %+v has no default", o)
		}
	}
	if !slices.Contains(options, 21) {
		t.Errorf("options %v miss Hair Style (21)", options)
	}
}

// TestHairStyleChoice checks that a hair style choice shows its geoset and its scalp texture.
func TestHairStyleChoice(t *testing.T) {
	r := testAppearance(t)
	// Topknot (403) is hair geoset 2. Its scalp texture needs the hair color, here the
	// default 415.
	a, err := r.character(characterRequest{Race: 2, Sex: 0, Class: 7, Choices: map[int]int{21: 403}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(a.Geosets, 2) {
		t.Errorf("geosets %v miss the Topknot (2)", a.Geosets)
	}
	for _, id := range []int{3, 4, 5} {
		if slices.Contains(a.Geosets, id) {
			t.Errorf("geosets %v show another hair style (%d)", a.Geosets, id)
		}
	}
	found := false
	for _, l := range a.Textures[1].Layers {
		if l.Order == 11 {
			found = true
		}
	}
	if !found {
		t.Errorf("no scalp layer (target 11) in %+v", a.Textures[1].Layers)
	}
}

func TestDummy(t *testing.T) {
	r := testAppearance(t)
	a, err := r.creature(dummyDisplayID)
	if err != nil {
		t.Fatal(err)
	}
	if a.Model != 125259 || a.Replaceable[11] != 125258 {
		t.Errorf("dummy %+v, want woodendummy.m2 (125259) with skin 125258", a)
	}
}
