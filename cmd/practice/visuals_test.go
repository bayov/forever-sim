//go:build practice

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The tests read the tables the server cached for the client's build. They skip on a machine
// without them, since fetching them would make the tests depend on wago.tools.
const visualsTestBuild = "1.60.1.70707"

func testVisuals(t *testing.T) *visuals {
	t.Helper()
	home, _ := os.UserHomeDir()
	root := filepath.Join(home, "forever-practice")
	dir := filepath.Join(root, "db2", visualsTestBuild)
	for _, name := range []string{"SpellXSpellVisual", "SpellVisual", "SpellVisualEvent", "SpellVisualKit", "SpellVisualKitEffect",
		"SpellVisualKitModelAttach", "SpellVisualEffectName", "SpellVisualMissile", "SpellVisualAnim", "SoundKit", "SoundKitEntry",
		"SpellEffect", "SpellMisc", "SummonProperties", "CreatureDisplayInfo", "CreatureModelData", "SpellItemEnchantment", "ItemVisuals",
		"ItemSubClass", "WeaponSwingSounds2", "WeaponImpactSounds", "CreatureSoundData", "FootstepTerrainLookup", "TerrainType", "Creature"} {
		if _, err := os.Stat(filepath.Join(dir, name+".csv")); err != nil {
			t.Skipf("no cached %s table for build %s", name, visualsTestBuild)
		}
	}
	return newVisuals(newDB2Tables(root, visualsTestBuild))
}

func eventsOf(sv *spellVisual, event, target string) []visualEvent {
	var out []visualEvent
	for _, e := range sv.Events {
		if e.Event == event && e.Target == target {
			out = append(out, e)
		}
	}
	return out
}

func TestLightningBoltVisual(t *testing.T) {
	v := testVisuals(t)
	sv, err := v.Spell(15208)
	if err != nil {
		t.Fatal(err)
	}
	if sv.Visual != 173 || sv.Self {
		t.Errorf("visual %d self %v, want 173 and not self", sv.Visual, sv.Self)
	}
	if sv.Anims["precast"] == nil || *sv.Anims["precast"] != 51 || sv.Anims["cast"] == nil || *sv.Anims["cast"] != 53 {
		t.Errorf("caster anims %v, want precast 51 (ReadySpellDirected) and cast 53 (SpellCastDirected)", sv.Anims)
	}
	cast := eventsOf(sv, "cast", "caster")
	if len(cast) != 1 || len(cast[0].Sounds) == 0 || len(cast[0].Models) != 2 {
		t.Fatalf("cast events %+v, want one with a sound and a glow on each hand", cast)
	}
	if hands := []int{cast[0].Models[0].Attachment, cast[0].Models[1].Attachment}; hands[0] != 21 || hands[1] != 22 {
		t.Errorf("cast glows on attachments %v, want the spell hands 21 and 22", hands)
	}
	precast := eventsOf(sv, "precast", "caster")
	if len(precast) != 1 || precast[0].Until != "precastEnd" || len(precast[0].Sounds) == 0 || !precast[0].Sounds[0].Loop {
		t.Errorf("precast %+v, want a looping sound until precastEnd", precast)
	}
	// The caster copy of the impact kit (261808) adds nothing to the target's (283), so it's gone.
	if n := len(eventsOf(sv, "impact", "caster")); n != 0 {
		t.Errorf("%d impact events on the caster, want 0", n)
	}
	impact := eventsOf(sv, "impact", "target")
	if len(impact) != 1 || impact[0].Kit != 283 || len(impact[0].Sounds) == 0 || len(impact[0].Models) != 1 || impact[0].Models[0].Attachment != 34 {
		t.Errorf("impact %+v, want kit 283 with a sound and a model on the chest (34)", impact)
	}
	m := sv.Missile
	if m == nil || m.Model.FDID != 166497 || m.Speed != 20 || m.Attachment != 38 || m.DestinationAttachment != 34 || len(m.Sound) == 0 {
		t.Errorf("missile %+v, want model 166497 at 20 yd/s from attachment 38 to 34 with a sound", m)
	}
}

func TestStormstrikeVisual(t *testing.T) {
	v := testVisuals(t)
	sv, err := v.Spell(17364)
	if err != nil {
		t.Fatal(err)
	}
	if sv.Anims["cast"] == nil || *sv.Anims["cast"] != 57 {
		t.Errorf("cast anim %v, want 57 (the special one-hand attack)", sv.Anims["cast"])
	}
	if sv.Missile != nil {
		t.Errorf("Stormstrike has a missile: %+v", sv.Missile)
	}
	cast := eventsOf(sv, "cast", "caster")
	if len(cast) != 1 || len(cast[0].Sounds) == 0 {
		t.Errorf("cast %+v, want one with the swing sound", cast)
	}
	impact := eventsOf(sv, "impact", "target")
	if len(impact) != 2 || impact[0].Models[0].FDID != 166844 || impact[1].DelayMs != 400 || len(impact[1].Sounds) == 0 {
		t.Errorf("impact %+v, want the model 166844 and a sound 400ms later", impact)
	}
}

func TestShamanEnchantsAndTotems(t *testing.T) {
	v := testVisuals(t)
	ev, err := v.Enchant(1669)
	if err != nil {
		t.Fatal(err)
	}
	if len(ev.Models) != 1 || ev.Models[0].FDID != 166022 || ev.Models[0].Attachment != 3 {
		t.Errorf("Windfury 4 glow %+v, want shaman_wind.m2 (166022) on attachment 3", ev.Models)
	}
	if ids := v.imbueEnchants()["Windfury Weapon"]; !containsInt(ids, 1669) {
		t.Errorf("Windfury Weapon enchants %v miss rank 4 (1669)", ids)
	}
	sv, err := v.Spell(10614)
	if err != nil {
		t.Fatal(err)
	}
	if sv.Totem == nil || sv.Totem.Element != "air" || sv.Totem.Model != 125990 {
		t.Errorf("Windfury Totem %+v, want the air totem model 125990", sv.Totem)
	}
	sv, err = v.Spell(10432)
	if err != nil {
		t.Fatal(err)
	}
	if !sv.Self || len(eventsOf(sv, "aura", "target")) != 1 || len(eventsOf(sv, "aura", "caster")) != 0 {
		t.Errorf("Lightning Shield %+v, want a self spell with one aura event", sv)
	}
}

func TestMeleeAndUnitSounds(t *testing.T) {
	v := testVisuals(t)
	melee := v.Melee(1, false, 0) // two-hand axe
	for _, key := range []string{"swing", "swingCrit", "impact", "impactCrit"} {
		if len(melee[key]) == 0 {
			t.Errorf("no %s sounds for a two-hand axe", key)
		}
	}
	if melee["swing"][0].Kit != 237 {
		t.Errorf("two-hand axe swings with kit %d, want the large woosh 237", melee["swing"][0].Kit)
	}
	unit, err := v.Unit(51894)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"attack", "wound", "woundCrit", "death"} {
		if len(unit.Sounds[key]) == 0 {
			t.Errorf("Orc male has no %s sound", key)
		}
	}
	if len(unit.Footsteps["Grass"]) == 0 {
		t.Errorf("Orc male has no grass footsteps")
	}
}

func containsInt(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// TestVisualFilesExtract checks that a sample of the resolved file ids are real files of the
// right kind in the client. It needs TACTTool and the client, so it skips without them.
func TestVisualFilesExtract(t *testing.T) {
	v := testVisuals(t)
	home, _ := os.UserHomeDir()
	toolDir := filepath.Join(home, "forever-practice", "tools", "tact-linux")
	wow := filepath.Join(home, "forever-practice", "wow")
	if _, err := os.Stat(filepath.Join(toolDir, "TACTTool")); err != nil {
		t.Skip("no TACTTool")
	}
	if _, err := os.Stat(wow); err != nil {
		t.Skip("no WoW client")
	}
	want := map[int]string{}
	lb, _ := v.Spell(15208)
	want[lb.Missile.Model.FDID] = "MD21"
	want[lb.Missile.Sound[0].FDID] = "OggS"
	for _, e := range lb.Events {
		for _, m := range e.Models {
			want[m.FDID] = "MD21"
		}
		for _, s := range e.Sounds {
			want[s.FDID] = "OggS"
		}
	}
	ev, _ := v.Enchant(1669)
	want[ev.Models[0].FDID] = "MD21"
	unit, _ := v.Unit(51894)
	want[unit.Sounds["attack"][0].FDID] = "OggS"
	want[v.Melee(1, false, 0)["impact"][0].FDID] = "OggS"
	totem, _ := v.Spell(10614)
	want[totem.Totem.Model] = "MD21"

	dir := t.TempDir()
	var list strings.Builder
	for fdid := range want {
		fmt.Fprintf(&list, "%d;%d.bin\n", fdid, fdid)
	}
	listPath := filepath.Join(dir, "list.txt")
	os.WriteFile(listPath, []byte(list.String()), 0o644)
	// TACTTool reads its decryption keys from its working folder.
	cmd := exec.Command("./TACTTool", "-p", "wowf", "-r", "eu", "-d", wow, "-m", "list", "-i", listPath, "-o", filepath.Join(dir, "out"))
	cmd.Dir = toolDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("TACTTool: %v\n%s", err, out)
	}
	for fdid, magic := range want {
		data, err := os.ReadFile(filepath.Join(dir, "out", fmt.Sprintf("%d.bin", fdid)))
		if err != nil {
			t.Errorf("file %d: %v", fdid, err)
			continue
		}
		if !bytes.HasPrefix(data, []byte(magic)) {
			t.Errorf("file %d starts with %q, want %s", fdid, data[:min(4, len(data))], magic)
		}
	}
}
