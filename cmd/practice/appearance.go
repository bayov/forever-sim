//go:build practice

package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Appearance recipes: what the page needs to build a dressed character or a creature from the
// client's files.
//
// The server reads the game tables and answers with FileDataIDs (the client's file numbers): the
// body model, the geosets (body parts) to show, the layers to paint onto each composited
// texture, and the models that hang on attachment points. The page fetches those files from
// /api/assets and does the drawing.
//
// The logic is a port of wow.export's character tab (MIT, github.com/Kruithne/wow.export):
// ui/character-appearance.js, modules/tab_characters.js and the db/caches readers it uses. Where
// we differ, a comment says why. Forever is a Classic Era client, so a few retail tables are
// missing (ChrCustomizationCondModel) and capes are a body texture, not a model.

// dummyDisplayID is the creature display we use for the room's training dummy. It's the wooden
// target dummy (CreatureModelData 270, creature/object/woodendummy.m2, FileDataID 125259) at
// scale 1. Display 16074 is the same model at scale 1.5.
const dummyDisplayID = 3019

// Inventory slots, as WoW numbers them.
const (
	slotHead     = 1
	slotShoulder = 3
	slotShirt    = 4
	slotChest    = 5
	slotWaist    = 6
	slotLegs     = 7
	slotFeet     = 8
	slotWrist    = 9
	slotHands    = 10
	slotBack     = 15
	slotMainHand = 16
	slotOffHand  = 17
	slotRanged   = 18
	slotTabard   = 19
)

// appearanceSlots are the slots that show on the model, in the order we apply them.
var appearanceSlots = []int{slotHead, slotShoulder, slotShirt, slotChest, slotWaist, slotLegs, slotFeet,
	slotWrist, slotHands, slotBack, slotMainHand, slotOffHand, slotRanged, slotTabard}

// M2 attachment points (wowmodelviewer's POSITION_SLOTS, as wow.export's EquipmentSlots.js lists them).
const (
	attachShield        = 0
	attachHandRight     = 1
	attachHandLeft      = 2
	attachShoulderRight = 5
	attachShoulderLeft  = 6
	attachHelmet        = 11
	attachBack          = 12
	attachSheathMain    = 26
	attachSheathOff     = 27
	attachSheathShield  = 28
	attachHipLeft       = 32
	attachHipRight      = 33
)

// slotAttachments are the attachment points for each slot's models, in ItemDisplayInfo's
// ModelResourcesID order. A slot's models past this list are collection models, drawn with the
// character's own skeleton.
//
// wow.export splits the shoulder into two slots (3 and 30). We keep the one WoW slot and give it
// both points, left first, because the shoulder's two models come back left then right.
var slotAttachments = map[int][]int{
	slotHead:     {attachHelmet},
	slotShoulder: {attachShoulderLeft, attachShoulderRight},
	slotBack:     {attachBack},
	slotMainHand: {attachHandRight},
	slotOffHand:  {attachHandLeft, attachShield},
	slotRanged:   {attachHandRight},
}

// slotLayer orders item textures on the body: a higher layer paints on top. The final order key
// is layer*100 + the body section.
var slotLayer = map[int]int{
	slotShirt: 10, slotLegs: 10, slotHead: 11, slotFeet: 11, slotShoulder: 13, slotChest: 13,
	slotTabard: 17, slotWaist: 18, slotWrist: 19, slotHands: 20, slotMainHand: 21, slotOffHand: 22,
	slotBack: 23,
}

// Character geoset groups. A group's geosets are numbered group*100 + value.
const (
	cgGloves         = 4
	cgBoots          = 5
	cgSleeves        = 8
	cgKneepads       = 9
	cgChest          = 10
	cgPants          = 11
	cgTabard         = 12
	cgTrousers       = 13
	cgCloak          = 15
	cgBelt           = 18
	cgFeet           = 20
	cgSkull          = 21
	cgTorso          = 22
	cgHandAttachment = 23
	cgShoulders      = 26
	cgHelm           = 27
	cgArmUpper       = 28
)

type slotGeoset struct {
	index int // into ItemDisplayInfo's GeosetGroup
	group int
	feet  bool // the boots' feet group counts differently
}

// slotGeosets says which body geoset groups each slot's item switches (DBItemGeosets.js, from
// wowmodelviewer's WoWItem.cpp).
var slotGeosets = map[int][]slotGeoset{
	slotHead:     {{0, cgHelm, false}, {1, cgSkull, false}},
	slotShoulder: {{0, cgShoulders, false}},
	slotShirt:    {{0, cgSleeves, false}, {1, cgChest, false}},
	slotChest:    {{0, cgSleeves, false}, {1, cgChest, false}, {2, cgTrousers, false}, {3, cgTorso, false}, {4, cgArmUpper, false}},
	slotWaist:    {{0, cgBelt, false}},
	slotLegs:     {{0, cgPants, false}, {1, cgKneepads, false}, {2, cgTrousers, false}},
	slotFeet:     {{0, cgBoots, false}, {1, cgFeet, true}},
	slotHands:    {{0, cgGloves, false}, {1, cgHandAttachment, false}},
	slotBack:     {{0, cgCloak, false}},
	slotTabard:   {{0, cgTabard, false}},
}

// geosetPriority picks the winning slot when two slots switch the same group: gloves over chest
// over shirt for sleeves, and so on.
var geosetPriority = map[int][]int{
	cgSleeves: {slotHands, slotChest, slotShirt}, cgChest: {slotChest, slotShirt},
	cgTrousers: {slotChest, slotLegs}, cgTabard: {slotTabard}, cgCloak: {slotBack}, cgBelt: {slotWaist},
	cgFeet: {slotFeet}, cgTorso: {slotChest}, cgHandAttachment: {slotHands}, cgHelm: {slotHead},
	cgArmUpper: {slotChest}, cgSkull: {slotHead}, cgShoulders: {slotShoulder}, cgBoots: {slotFeet},
	cgGloves: {slotHands}, cgPants: {slotLegs}, cgKneepads: {slotLegs},
}

// collectionGeosets are the groups a slot's collection model shows, from ItemDisplayInfo's
// AttachmentGeosetGroup (tab_characters.js).
var collectionGeosets = map[int][]slotGeoset{
	slotHead:  {{0, cgHelm, false}, {1, cgSkull, false}},
	slotChest: {{0, cgSleeves, false}, {1, cgChest, false}, {2, cgTrousers, false}, {3, cgTorso, false}, {4, cgArmUpper, false}},
	slotWaist: {{0, cgBelt, false}},
	slotLegs:  {{0, cgPants, false}, {1, cgKneepads, false}, {2, cgTrousers, false}},
	slotFeet:  {{0, cgBoots, false}, {1, cgFeet, false}},
	slotHands: {{0, cgGloves, false}, {1, cgHandAttachment, false}},
	slotBack:  {{0, cgCloak, false}},
}

// npcSlots maps NPCModelItemSlotDisplayInfo.ItemSlot to inventory slots.
var npcSlots = map[int]int{0: slotHead, 1: slotShoulder, 2: slotShirt, 3: slotChest, 4: slotWaist, 5: slotLegs,
	6: slotFeet, 7: slotWrist, 8: slotHands, 9: slotTabard, 10: slotBack}

// The reply.

type appearance struct {
	Model             int                      `json:"model"`
	ChrModelID        int                      `json:"chrModelId,omitempty"`
	CreatureDisplayID int                      `json:"creatureDisplayId,omitempty"`
	Scale             float64                  `json:"scale"`
	Geosets           []int                    `json:"geosets"`
	GeosetsExact      bool                     `json:"geosetsExact"`
	Textures          map[int]*composedTexture `json:"textures,omitempty"`
	Replaceable       map[int]int              `json:"replaceable,omitempty"`
	Attachments       []attachedModel          `json:"attachments,omitempty"`
	Collections       []collectionModel        `json:"collections,omitempty"`
	Options           []customizationOption    `json:"options,omitempty"`
	Warnings          []string                 `json:"warnings,omitempty"`
}

// composedTexture is one texture type the page paints from layers, like the skin (type 1).
type composedTexture struct {
	Width  int            `json:"width"`
	Height int            `json:"height"`
	Layers []textureLayer `json:"layers"`
}

// textureLayer is one image painted into a rectangle of a composed texture. Layers come sorted in
// draw order. Order is the sort key: the ChrModelTextureTargetID for customization layers (13 and
// 14 are the default underwear), and slot layer*100 + section for item layers.
type textureLayer struct {
	Fdid      int  `json:"fdid"`
	X         int  `json:"x"`
	Y         int  `json:"y"`
	W         int  `json:"w"`
	H         int  `json:"h"`
	BlendMode int  `json:"blendMode"`
	Order     int  `json:"order"`
	Alpha     bool `json:"alpha"`
	Section   int  `json:"section"`        // CharComponentTextureSections.SectionType, -1 for the whole texture
	Slot      int  `json:"slot,omitempty"` // the item's slot, for item layers
}

// attachedModel hangs on one of the character's attachment points.
type attachedModel struct {
	Slot       int `json:"slot"`
	Attachment int `json:"attachment"`
	// SheathedAttachment is where a weapon goes while sheathed, -1 when it isn't shown sheathed.
	// Only weapon slots have it.
	SheathedAttachment *int        `json:"sheathedAttachment,omitempty"`
	Sheathe            *int        `json:"sheathe,omitempty"` // Item.SheatheType
	Model              int         `json:"model"`
	Textures           []int       `json:"textures"`              // by texture type from 2 up, like wow.export's list
	Replaceable        map[int]int `json:"replaceable,omitempty"` // the same, by texture type
	Item               int         `json:"item,omitempty"`
	Display            int         `json:"display"`
}

// collectionModel is drawn with the character's skeleton (bones matched by id). Geosets are the
// only ones to show, or all of them when it's empty.
type collectionModel struct {
	Slot        int         `json:"slot"`
	Model       int         `json:"model"`
	Textures    []int       `json:"textures"`
	Replaceable map[int]int `json:"replaceable,omitempty"`
	Geosets     []int       `json:"geosets,omitempty"`
	Item        int         `json:"item,omitempty"`
	Display     int         `json:"display,omitempty"`
}

type customizationOption struct {
	ID      int                   `json:"id"`
	Name    string                `json:"name"`
	Choices []customizationChoice `json:"choices"`
	Chosen  int                   `json:"chosen"`
}

type customizationChoice struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Swatch []int  `json:"swatch,omitempty"` // two colors as signed ARGB ints, for color pickers
}

// appearanceResolver reads the tables. Skin gives a model's geoset ids from its first .skin file,
// or nil when the file can't be read. Without it we work on every possible id, and the list
// comes back with ids the model may not have (geosetsExact is false).
type appearanceResolver struct {
	db *db2Tables
	// skin and textureTypes read the model file itself. Either may be nil.
	skin         func(model int) []int
	textureTypes func(model int) []int
}

// characterRequest is a character to dress.
type characterRequest struct {
	Race, Sex, Class int
	Items            map[int]int // slot -> item id
	Choices          map[int]int // option id -> choice id
}

// equipped is one item on the character, found by item id or (for NPCs) by display id.
type equipped struct {
	slot     int
	item     int
	display  db2Row
	sheathe  int
	invType  int
	class    int
	subclass int
}

type activeChoice struct{ option, choice int }

func (r *appearanceResolver) t(name string) *db2Table { return r.db.MustTable(name) }

func (r *appearanceResolver) character(req characterRequest) (*appearance, error) {
	var chrModel int
	for _, row := range r.t("ChrRaceXChrModel").Where("ChrRacesID", req.Race) {
		if row.Int("Sex") == req.Sex {
			chrModel = row.Int("ChrModelID")
		}
	}
	if chrModel == 0 {
		return nil, fmt.Errorf("no character model for race %d sex %d", req.Race, req.Sex)
	}
	var warnings []string
	var items []equipped
	for _, slot := range appearanceSlots {
		id, ok := req.Items[slot]
		if !ok || id == 0 {
			continue
		}
		e, err := r.itemByID(slot, id)
		if err != nil {
			warnings = append(warnings, err.Error())
			continue
		}
		items = append(items, e)
	}
	shown := map[int]bool{}
	for _, slot := range appearanceSlots {
		shown[slot] = true
	}
	for slot := range req.Items {
		// The neck, rings and trinkets (2, 11 to 14) don't show, so they're no surprise.
		if !shown[slot] && slot != 2 && (slot < 11 || slot > 14) {
			warnings = append(warnings, fmt.Sprintf("slot %d doesn't show on the model", slot))
		}
	}
	a, err := r.dress(chrModel, req.Race, req.Sex, req.Class, req.Choices, items, 0)
	if err != nil {
		return nil, err
	}
	a.Warnings = append(warnings, a.Warnings...)
	return a, nil
}

// itemByID follows an item to its display: ItemModifiedAppearance (modifier 0, or else the
// lowest), ItemAppearance, ItemDisplayInfo.
func (r *appearanceResolver) itemByID(slot, item int) (equipped, error) {
	ia := r.t("ItemAppearance")
	best, bestMod := 0, -1
	for _, row := range r.t("ItemModifiedAppearance").Where("ItemID", item) {
		display := ia.Get(row.Int("ItemAppearanceID")).Int("ItemDisplayInfoID")
		if display == 0 {
			continue
		}
		mod := row.Int("ItemAppearanceModifierID")
		// A later row for the same modifier wins, as in wow.export's map.
		if bestMod == -1 || mod < bestMod || mod == bestMod {
			best, bestMod = display, mod
		}
	}
	if best == 0 {
		return equipped{}, fmt.Errorf("item %d (slot %d) has no appearance", item, slot)
	}
	disp := r.t("ItemDisplayInfo").Get(best)
	if !disp.Ok() {
		return equipped{}, fmt.Errorf("item %d (slot %d): no ItemDisplayInfo %d", item, slot, best)
	}
	e := equipped{slot: slot, item: item, display: disp}
	if row := r.t("Item").Get(item); row.Ok() {
		e.sheathe, e.invType = row.Int("SheatheType"), row.Int("InventoryType")
		e.class, e.subclass = row.Int("ClassID"), row.Int("SubclassID")
	}
	return e, nil
}

// dress builds a character model's recipe. Bake is a baked NPC skin (FileDataID) or 0.
func (r *appearanceResolver) dress(chrModel, race, sex, class int, choices map[int]int, items []equipped, bake int) (*appearance, error) {
	cm := r.t("ChrModel").Get(chrModel)
	if !cm.Ok() {
		return nil, fmt.Errorf("no ChrModel %d", chrModel)
	}
	a := &appearance{ChrModelID: chrModel, Textures: map[int]*composedTexture{}, Replaceable: map[int]int{}}
	model, scale := r.displayModel(cm.Int("DisplayID"))
	if model == 0 {
		return nil, fmt.Errorf("ChrModel %d: no model for display %d", chrModel, cm.Int("DisplayID"))
	}
	a.Model, a.Scale = model, scale
	layout := cm.Int("CharComponentTextureLayoutID")

	active, options, warnings := r.customization(chrModel, class, choices)
	a.Options, a.Warnings = options, warnings

	a.Geosets, a.GeosetsExact = r.characterGeosets(model, active, items, race, sex)
	r.characterTextures(a, layout, active, items, race, sex, bake)
	r.itemModels(a, items, race, sex)
	r.skinnedModels(a, active)
	if len(a.Replaceable) == 0 {
		a.Replaceable = nil
	}
	return a, nil
}

// displayModel gives a creature display's model file and its scale.
func (r *appearanceResolver) displayModel(display int) (int, float64) {
	cdi := r.t("CreatureDisplayInfo").Get(display)
	if !cdi.Ok() {
		return 0, 0
	}
	cmd := r.t("CreatureModelData").Get(cdi.Int("ModelID"))
	scale := cdi.Float("CreatureModelScale")
	if scale == 0 {
		scale = 1
	}
	if s := cmd.Float("ModelScale"); s > 0 {
		scale *= s
	}
	return cmd.Int("FileDataID"), scale
}

// customization picks a choice for each option: the one asked for, else the first one the
// class may use. Options with flag 0x20 get no default, as in wow.export (Grime, Eye Style).
func (r *appearanceResolver) customization(chrModel, class int, asked map[int]int) ([]activeChoice, []customizationOption, []string) {
	var warnings []string
	opts := r.t("ChrCustomizationOption").Where("ChrModelID", chrModel)
	sort.SliceStable(opts, func(i, j int) bool { return opts[i].Int("OrderIndex") < opts[j].Int("OrderIndex") })
	var active []activeChoice
	var out []customizationOption
	seen := map[int]bool{}
	for _, opt := range opts {
		id := opt.Int("ID")
		if !r.classAllowed(opt.Int("Requirement"), class) {
			continue
		}
		seen[id] = true
		o := customizationOption{ID: id, Name: opt.Str("Name_lang")}
		if o.Name == "" {
			o.Name = fmt.Sprintf("Option %d", opt.Int("OrderIndex"))
		}
		choiceRows := r.t("ChrCustomizationChoice").Where("ChrCustomizationOptionID", id)
		sort.SliceStable(choiceRows, func(i, j int) bool {
			return choiceRows[i].Int("OrderIndex") < choiceRows[j].Int("OrderIndex")
		})
		valid := map[int]bool{}
		for _, c := range choiceRows {
			if !r.classAllowed(c.Int("ChrCustomizationReqID"), class) {
				continue
			}
			ch := customizationChoice{ID: c.Int("ID"), Name: c.Str("Name_lang")}
			if ch.Name == "" {
				ch.Name = fmt.Sprintf("Choice %d", c.Int("OrderIndex"))
			}
			if s0, s1 := c.Int("SwatchColor_0"), c.Int("SwatchColor_1"); s0 != 0 || s1 != 0 {
				ch.Swatch = []int{s0, s1}
			}
			valid[ch.ID] = true
			o.Choices = append(o.Choices, ch)
		}
		if want, ok := asked[id]; ok {
			if valid[want] {
				o.Chosen = want
			} else {
				warnings = append(warnings, fmt.Sprintf("choice %d isn't one of option %d's", want, id))
			}
		}
		if o.Chosen == 0 && opt.Int("Flags")&0x20 == 0 && len(o.Choices) > 0 {
			o.Chosen = o.Choices[0].ID
		}
		if o.Chosen != 0 {
			active = append(active, activeChoice{id, o.Chosen})
		}
		out = append(out, o)
	}
	for id := range asked {
		if !seen[id] {
			warnings = append(warnings, fmt.Sprintf("option %d isn't one of this model's", id))
		}
	}
	return active, out, warnings
}

// classAllowed reads a ChrCustomizationReq's class mask. Zero means any class.
func (r *appearanceResolver) classAllowed(req, class int) bool {
	if req == 0 || class <= 0 {
		return true
	}
	mask := r.t("ChrCustomizationReq").Get(req).Int("ClassMask")
	return mask == 0 || mask&(1<<(class-1)) != 0
}

func isActive(active []activeChoice, choice int) bool {
	for _, c := range active {
		if c.choice == choice {
			return true
		}
	}
	return false
}

// customizationGeoset turns a ChrCustomizationGeoset row into a geoset id. wow.export pads both
// numbers to two digits and joins them, which is type*100 + id for ids under 100.
func (r *appearanceResolver) customizationGeoset(id int) int {
	row := r.t("ChrCustomizationGeoset").Get(id)
	if !row.Ok() {
		return -1
	}
	v, _ := strconv.Atoi(fmt.Sprintf("%02d%02d", row.Int("GeosetType"), row.Int("GeosetID")))
	return v
}

// defaultGeoset is the M2 renderer's first guess for a character geoset: 0, every group's first
// variant (x01) and the 32xx group, but no eye glows (17xx) or piercings (35xx).
func defaultGeoset(id int) bool {
	s := strconv.Itoa(id)
	if strings.HasPrefix(s, "17") || strings.HasPrefix(s, "35") {
		return false
	}
	return id == 0 || strings.HasSuffix(s, "01") || strings.HasPrefix(s, "32")
}

// candidateGeosets gives the model's geoset ids, or every id up to 9999 when we can't read it.
// Every rule below decides each id on its own, so the result is the same either way once the
// page drops the ids its model doesn't have.
func (r *appearanceResolver) candidateGeosets(model int) ([]int, bool) {
	if r.skin != nil {
		if ids := r.skin(model); ids != nil {
			return ids, true
		}
	}
	ids := make([]int, 10000)
	for i := range ids {
		ids[i] = i
	}
	return ids, false
}

func (r *appearanceResolver) characterGeosets(model int, active []activeChoice, items []equipped, race, sex int) ([]int, bool) {
	ids, exact := r.candidateGeosets(model)
	shown := map[int]bool{}
	for _, id := range ids {
		shown[id] = defaultGeoset(id)
	}
	set := func(id int, v bool) {
		if _, ok := shown[id]; ok && id != 0 {
			shown[id] = v
		}
	}
	hideGroup := func(group int) {
		for id := group*100 + 1; id <= group*100+99; id++ {
			set(id, false)
		}
	}

	// Customization. wow.export keeps one geoset per choice, but a choice like Facial Hair sets
	// several groups (1xx and 2xx), so we take them all. We hide every other choice's geosets
	// first and then show the chosen one's, so two choices that share a geoset can't hide it.
	elements := r.t("ChrCustomizationElement")
	for _, ac := range active {
		for _, c := range r.t("ChrCustomizationChoice").Where("ChrCustomizationOptionID", ac.option) {
			if c.Int("ID") == ac.choice {
				continue
			}
			for _, e := range elements.Where("ChrCustomizationChoiceID", c.Int("ID")) {
				if g := e.Int("ChrCustomizationGeosetID"); g != 0 {
					set(r.customizationGeoset(g), false)
				}
			}
		}
	}
	for _, ac := range active {
		for _, e := range elements.Where("ChrCustomizationChoiceID", ac.choice) {
			g := e.Int("ChrCustomizationGeosetID")
			if g == 0 {
				continue
			}
			if rel := e.Int("RelatedChrCustomizationChoiceID"); rel != 0 && !isActive(active, rel) {
				continue
			}
			set(r.customizationGeoset(g), true)
		}
	}

	// A customization model (ChrCustomizationSkinnedModel) replaces the body's own geosets in
	// its group.
	for _, ac := range active {
		for _, e := range elements.Where("ChrCustomizationChoiceID", ac.choice) {
			if sm := e.Int("ChrCustomizationSkinnedModelID"); sm != 0 {
				if row := r.t("ChrCustomizationSkinnedModel").Get(sm); row.Ok() {
					hideGroup(row.Int("GeosetType"))
				}
			}
		}
	}

	// Equipment: each slot's GeosetGroup picks one variant in its groups, with the priority
	// list deciding between slots.
	type slotValue struct{ slot, value int }
	values := map[int][]slotValue{}
	var groups []int
	for _, e := range items {
		for _, sg := range slotGeosets[e.slot] {
			g := e.display.Int(fmt.Sprintf("GeosetGroup_%d", sg.index))
			v := 1 + g
			if sg.feet {
				// Boots' feet group: 0 means variant 2, anything else is the variant itself.
				v = g
				if g == 0 {
					v = 2
				}
			}
			if _, ok := values[sg.group]; !ok {
				groups = append(groups, sg.group)
			}
			values[sg.group] = append(values[sg.group], slotValue{e.slot, v})
		}
	}
	for _, group := range groups {
		vs := values[group]
		value := vs[0].value
		if order, ok := geosetPriority[group]; ok {
		pick:
			for _, slot := range order {
				for _, sv := range vs {
					if sv.slot == slot {
						value = sv.value
						break pick
					}
				}
			}
		}
		hideGroup(group)
		set(group*100+value, true)
	}

	// Helmets hide hair, ears and so on, by race (HelmetGeosetData).
	for _, e := range items {
		if e.slot != slotHead {
			continue
		}
		vis := e.display.Int(fmt.Sprintf("HelmetGeosetVis_%d", sex))
		if vis == 0 {
			continue
		}
		for _, row := range r.t("HelmetGeosetData").Where("HelmetGeosetVisDataID", vis) {
			if row.Int("RaceID") == race {
				hideGroup(row.Int("HideGeosetGroup"))
			}
		}
	}

	var out []int
	for _, id := range ids {
		if shown[id] {
			out = append(out, id)
		}
	}
	sort.Ints(out)
	return out, exact
}

// layoutData is a texture layout's tables: materials by texture type, layers by target and the
// body sections.
type layoutData struct {
	materials map[int]db2Row // texture type -> ChrModelMaterial
	layers    map[int]db2Row // ChrModelTextureTargetID -> ChrModelTextureLayer
	order     []int          // targets in table order (first appearance)
	sections  []db2Row
}

func (r *appearanceResolver) layout(id int) layoutData {
	l := layoutData{materials: map[int]db2Row{}, layers: map[int]db2Row{}}
	for _, row := range r.t("ChrModelMaterial").Where("CharComponentTextureLayoutsID", id) {
		l.materials[row.Int("TextureType")] = row
	}
	for _, row := range r.t("ChrModelTextureLayer").Where("CharComponentTextureLayoutsID", id) {
		target := row.Int("ChrModelTextureTargetID_0")
		if _, ok := l.layers[target]; !ok {
			l.order = append(l.order, target)
		}
		l.layers[target] = row
	}
	l.sections = r.t("CharComponentTextureSections").Where("CharComponentTextureLayoutID", id)
	return l
}

// textureFdids gives a material resource's texture files (TextureFileData, usage 0).
func (r *appearanceResolver) textureFdids(matRes int) []int {
	if matRes == 0 {
		return nil
	}
	var out []int
	for _, row := range r.t("TextureFileData").Where("MaterialResourcesID", matRes) {
		if row.Int("UsageType") == 0 {
			out = append(out, row.Int("FileDataID"))
		}
	}
	return out
}

// pickTexture picks the file for a race, sex and class from ComponentTextureFileData
// (DBComponentTextureFileData.js): drop the other sex and other classes, then prefer our own
// sex over "any", our class over generic, our race over others. Gender 3 is "any".
func (r *appearanceResolver) pickTexture(fdids []int, race, sex, class int) int {
	if len(fdids) == 0 {
		return 0
	}
	if len(fdids) == 1 {
		return fdids[0]
	}
	ctfd := r.t("ComponentTextureFileData")
	type cand struct {
		fdid int
		info db2Row
	}
	var tagged []cand
	untagged, anyTagged := 0, false
	for _, f := range fdids {
		info := ctfd.Get(f)
		if !info.Ok() {
			if untagged == 0 {
				untagged = f
			}
			continue
		}
		anyTagged = true
		g, c := info.Int("GenderIndex"), info.Int("ClassID")
		if (g == sex || g == 3) && (c == 0 || c == class) {
			tagged = append(tagged, cand{f, info})
		}
	}
	if !anyTagged {
		return fdids[0]
	}
	if len(tagged) == 0 {
		return untagged
	}
	rank := func(c cand) [3]int {
		var k [3]int
		if c.info.Int("GenderIndex") != sex {
			k[0] = 1
		}
		// Class 0 here means no class wanted, so the subject-class rank never wins (as in
		// wow.export, which passes class 12 only for demon hunters).
		k[1] = 1
		if c.info.Int("RaceID") != race {
			k[2] = 1
		}
		return k
	}
	sort.SliceStable(tagged, func(i, j int) bool {
		a, b := rank(tagged[i]), rank(tagged[j])
		for k := range a {
			if a[k] != b[k] {
				return a[k] < b[k]
			}
		}
		return false
	})
	return tagged[0].fdid
}

func (r *appearanceResolver) characterTextures(a *appearance, layoutID int, active []activeChoice, items []equipped, race, sex, bake int) {
	l := r.layout(layoutID)
	add := func(textureType int, layer textureLayer) {
		mat, ok := l.materials[textureType]
		if !ok {
			return
		}
		t := a.Textures[textureType]
		if t == nil {
			t = &composedTexture{Width: mat.Int("Width"), Height: mat.Int("Height")}
			a.Textures[textureType] = t
		}
		layer.Alpha = true
		t.Layers = append(t.Layers, layer)
	}

	// A baked NPC skin covers the layout's lowest texture type, and customization doesn't
	// paint that type then.
	bakedType := -1
	if bake != 0 {
		for tt := range l.materials {
			if bakedType == -1 || tt < bakedType {
				bakedType = tt
			}
		}
		if bakedType != -1 {
			m := l.materials[bakedType]
			add(bakedType, textureLayer{Fdid: bake, W: m.Int("Width"), H: m.Int("Height"), BlendMode: 0, Order: 0, Section: -1})
		}
	}

	elements := r.t("ChrCustomizationElement")
	mats := r.t("ChrCustomizationMaterial")
	for _, ac := range active {
		for _, e := range elements.Where("ChrCustomizationChoiceID", ac.choice) {
			matID := e.Int("ChrCustomizationMaterialID")
			if matID == 0 {
				continue
			}
			if rel := e.Int("RelatedChrCustomizationChoiceID"); rel != 0 && !isActive(active, rel) {
				continue
			}
			mat := mats.Get(matID)
			files := r.textureFdids(mat.Int("MaterialResourcesID"))
			if len(files) == 0 {
				continue
			}
			// wow.export's map keeps the last file of a material resource.
			fdid := files[len(files)-1]
			target := mat.Int("ChrModelTextureTargetID")
			layer, ok := l.layers[target]
			if !ok {
				continue
			}
			tt := layer.Int("TextureType")
			if _, ok := l.materials[tt]; !ok {
				// Not a composed type (a blindfold, say): the model takes the file as is.
				a.Replaceable[tt] = fdid
				continue
			}
			if tt == bakedType {
				continue
			}
			x, y, w, h, section, ok := l.sectionRect(layer, tt)
			if !ok {
				continue
			}
			add(tt, textureLayer{Fdid: fdid, X: x, Y: y, W: w, H: h, BlendMode: layer.Int("BlendMode"), Order: target, Section: section})
		}
	}

	// Item textures, one per body section. Each section paints with the layout's layer for it,
	// or the base skin layer when the layout has none.
	sectionRows := map[int]db2Row{}
	for _, s := range l.sections {
		if _, ok := sectionRows[s.Int("SectionType")]; !ok {
			sectionRows[s.Int("SectionType")] = s
		}
	}
	var base db2Row
	bySection := map[int]db2Row{}
	for _, target := range l.order {
		layer := l.layers[target]
		mask := layer.Int("TextureSectionTypeBitMask")
		if mask == -1 {
			if !base.Ok() && layer.Int("TextureType") == 1 {
				base = layer
			}
			continue
		}
		for st := 0; st < 9; st++ {
			if mask&(1<<st) != 0 {
				if _, ok := bySection[st]; !ok {
					bySection[st] = layer
				}
			}
		}
	}
	if base.Ok() {
		for st := 0; st < 9; st++ {
			if _, ok := bySection[st]; !ok {
				bySection[st] = base
			}
		}
	}
	for _, e := range items {
		for _, row := range r.t("ItemDisplayInfoMaterialRes").Where("ItemDisplayInfoID", e.display.Int("ID")) {
			st := row.Int("ComponentSection")
			fdid := r.pickTexture(r.textureFdids(row.Int("MaterialResourcesID")), race, sex, 0)
			if fdid == 0 {
				continue
			}
			sec, ok := sectionRows[st]
			if !ok {
				continue
			}
			layer, ok := bySection[st]
			if !ok {
				continue
			}
			blend := layer.Int("BlendMode")
			// Item textures go over the skin. A plain copy (modes 0 and 1) would erase the body
			// where the texture is transparent, so these blend by alpha (mode 15).
			if blend == 0 || blend == 1 {
				blend = 15
			}
			layerOf, ok := slotLayer[e.slot]
			if !ok {
				layerOf = 10
			}
			add(layer.Int("TextureType"), textureLayer{
				Fdid: fdid, X: sec.Int("X"), Y: sec.Int("Y"), W: sec.Int("Width"), H: sec.Int("Height"),
				BlendMode: blend, Order: layerOf*100 + st, Section: st, Slot: e.slot,
			})
		}
	}

	for _, t := range a.Textures {
		sort.SliceStable(t.Layers, func(i, j int) bool { return t.Layers[i].Order < t.Layers[j].Order })
	}
	if len(a.Textures) == 0 {
		a.Textures = nil
	}
}

// sectionRect is where a customization layer paints: the whole texture, or the first body
// section in its mask.
func (l layoutData) sectionRect(layer db2Row, textureType int) (x, y, w, h, section int, ok bool) {
	mat := l.materials[textureType]
	mask := layer.Int("TextureSectionTypeBitMask")
	if mask == -1 {
		return 0, 0, mat.Int("Width"), mat.Int("Height"), -1, true
	}
	for _, s := range l.sections {
		if (1<<s.Int("SectionType"))&mask != 0 {
			return s.Int("X"), s.Int("Y"), s.Int("Width"), s.Int("Height"), s.Int("SectionType"), true
		}
	}
	return 0, 0, 0, 0, 0, false
}

// modelFdids gives a model resource's files (ModelFileData).
func (r *appearanceResolver) modelFdids(res int) []int {
	if res == 0 {
		return nil
	}
	var out []int
	for _, row := range r.t("ModelFileData").Where("ModelResourcesID", res) {
		out = append(out, row.Int("FileDataID"))
	}
	return out
}

// pickModel picks a model file for a race and sex (ComponentModelFileData): exact match, then
// the race with any sex (2), then any race, then the first.
func (r *appearanceResolver) pickModel(fdids []int, race, sex int) int {
	if len(fdids) == 0 {
		return 0
	}
	if len(fdids) == 1 {
		return fdids[0]
	}
	cmfd := r.t("ComponentModelFileData")
	for _, test := range []func(db2Row) bool{
		func(i db2Row) bool { return i.Int("RaceID") == race && i.Int("GenderIndex") == sex },
		func(i db2Row) bool { return i.Int("RaceID") == race && i.Int("GenderIndex") == 2 },
		func(i db2Row) bool { return i.Int("RaceID") == 0 },
	} {
		for _, f := range fdids {
			if info := cmfd.Get(f); info.Ok() && test(info) {
				return f
			}
		}
	}
	return fdids[0]
}

// pickShoulders splits a shoulder's model pool by PositionIndex: 0 is the left one, 1 the right.
func (r *appearanceResolver) pickShoulders(fdids []int, race, sex int) (left, right int) {
	cmfd := r.t("ComponentModelFileData")
	byPos := map[int][]int{}
	for _, f := range fdids {
		if info := cmfd.Get(f); info.Ok() {
			if p := info.Int("PositionIndex"); p == 0 || p == 1 {
				byPos[p] = append(byPos[p], f)
			}
		}
	}
	best := func(c []int) int {
		for _, test := range []func(db2Row) bool{
			func(i db2Row) bool { return i.Int("RaceID") == race && i.Int("GenderIndex") == sex },
			func(i db2Row) bool { return i.Int("RaceID") == race && i.Int("GenderIndex") == 2 },
			func(i db2Row) bool { return i.Int("RaceID") == 0 },
		} {
			for _, f := range c {
				if test(cmfd.Get(f)) {
					return f
				}
			}
		}
		if len(c) > 0 {
			return c[0]
		}
		return 0
	}
	return best(byPos[0]), best(byPos[1])
}

// modelTextures gives the textures for one of a display's models, by texture type. We read
// ItemDisplayInfoModelMatRes for the model's index, and fall back to the display's own
// ModelMaterialResourcesID as the item texture (type 2).
func (r *appearanceResolver) modelTextures(display db2Row, index, race, sex int) map[int]int {
	out := map[int]int{}
	for _, row := range r.t("ItemDisplayInfoModelMatRes").Where("ItemDisplayInfoID", display.Int("ID")) {
		if row.Int("ModelIndex") != index {
			continue
		}
		if f := r.pickTexture(r.textureFdids(row.Int("MaterialResourcesID")), race, sex, 0); f != 0 {
			out[row.Int("TextureType")] = f
		}
	}
	if _, ok := out[2]; !ok {
		if f := r.pickTexture(r.textureFdids(display.Int(fmt.Sprintf("ModelMaterialResourcesID_%d", index))), race, sex, 0); f != 0 {
			out[2] = f
		}
	}
	return out
}

// textureList lays a by-type map out as wow.export's list: index 0 is type 2.
func textureList(m map[int]int) []int {
	var out []int
	for tt := 2; tt <= 4; tt++ {
		if f, ok := m[tt]; ok {
			for len(out) < tt-2 {
				out = append(out, 0)
			}
			out = append(out, f)
		}
	}
	if out == nil {
		out = []int{}
	}
	return out
}

// sheathedAt says where a weapon hangs while sheathed, from its Item.SheatheType: two-handers
// and staves on the back, one-handers on the hip, shields on the back. Main hand weapons go on
// the left side (the right hand draws across the body), off hand ones on the right.
func sheathedAt(slot, sheathe int) int {
	switch sheathe {
	case 1, 2:
		if slot == slotOffHand {
			return attachSheathOff
		}
		return attachSheathMain
	case 3, 5, 7:
		if slot == slotOffHand {
			return attachHipRight
		}
		return attachHipLeft
	case 6:
		return attachHipRight
	case 4:
		return attachSheathShield
	}
	return -1
}

func (r *appearanceResolver) itemModels(a *appearance, items []equipped, race, sex int) {
	for _, e := range items {
		d := e.display
		res0, res1 := d.Int("ModelResourcesID_0"), d.Int("ModelResourcesID_1")
		opts0, opts1 := r.modelFdids(res0), r.modelFdids(res1)

		// Classic capes have no model. Their texture paints the body's cape geosets (texture
		// type 2, the object skin).
		if e.slot == slotBack && len(opts0) == 0 && len(opts1) == 0 {
			if f := r.modelTextures(d, 0, race, sex)[2]; f != 0 {
				a.Replaceable[2] = f
			}
			continue
		}

		type modelRef struct{ fdid, index int }
		var models []modelRef
		if len(opts0) > 0 && len(opts1) > 0 && equalInts(opts0, opts1) {
			// Shoulders: one model pool, split into left and right.
			left, right := r.pickShoulders(opts0, race, sex)
			if left != 0 {
				models = append(models, modelRef{left, 0})
			}
			if right != 0 {
				models = append(models, modelRef{right, 1})
			}
		} else {
			for i, opts := range [][]int{opts0, opts1} {
				if f := r.pickModel(opts, race, sex); f != 0 {
					models = append(models, modelRef{f, i})
				}
			}
		}
		if len(models) == 0 {
			continue
		}

		points := slotAttachments[e.slot]
		switch {
		case e.slot == slotOffHand && e.invType == 14:
			// Shields hang on the shield point, not in the hand.
			points = []int{attachShield}
		case (e.slot == slotMainHand || e.slot == slotRanged) && e.class == 2 && e.subclass == 2:
			// Bows are held in the left hand.
			points = []int{attachHandLeft}
		}
		for i, m := range models {
			tex := r.modelTextures(d, m.index, race, sex)
			if i < len(points) {
				am := attachedModel{Slot: e.slot, Attachment: points[i], Model: m.fdid, Textures: textureList(tex),
					Replaceable: tex, Item: e.item, Display: d.Int("ID")}
				if e.slot == slotMainHand || e.slot == slotOffHand || e.slot == slotRanged {
					sh := e.sheathe
					at := sheathedAt(e.slot, sh)
					if e.slot == slotRanged && at == -1 {
						// Bows and guns have no sheathe type but show on the back. This is a
						// guess at the point: the tables don't say.
						at = attachSheathMain
					}
					am.Sheathe, am.SheathedAttachment = &sh, &at
				}
				a.Attachments = append(a.Attachments, am)
				continue
			}
			cm := collectionModel{Slot: e.slot, Model: m.fdid, Textures: textureList(tex), Replaceable: tex,
				Item: e.item, Display: d.Int("ID")}
			for _, sg := range collectionGeosets[e.slot] {
				v := d.Int(fmt.Sprintf("AttachmentGeosetGroup_%d", sg.index))
				cm.Geosets = append(cm.Geosets, sg.group*100+1+v)
			}
			a.Collections = append(a.Collections, cm)
		}
	}
}

// skinnedModels adds customization models (ChrCustomizationSkinnedModel), which draw with the
// character's skeleton and show one geoset each. They take the composed skin by texture type.
// Slot 0 marks them as not from an item.
func (r *appearanceResolver) skinnedModels(a *appearance, active []activeChoice) {
	byModel := map[int]*collectionModel{}
	for _, ac := range active {
		for _, e := range r.t("ChrCustomizationElement").Where("ChrCustomizationChoiceID", ac.choice) {
			sm := e.Int("ChrCustomizationSkinnedModelID")
			if sm == 0 {
				continue
			}
			row := r.t("ChrCustomizationSkinnedModel").Get(sm)
			if !row.Ok() {
				continue
			}
			f := row.Int("CollectionsFileDataID")
			cm := byModel[f]
			if cm == nil {
				a.Collections = append(a.Collections, collectionModel{Model: f, Textures: []int{}})
				cm = &a.Collections[len(a.Collections)-1]
				byModel[f] = cm
			}
			cm.Geosets = append(cm.Geosets, row.Int("GeosetType")*100+row.Int("GeosetID"))
		}
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// creature builds a creature display's recipe. Displays with character models (an
// ExtendedDisplayInfoID) come back as a dressed character with their baked skin, choices and
// armor.
func (r *appearanceResolver) creature(display int) (*appearance, error) {
	cdi := r.t("CreatureDisplayInfo").Get(display)
	if !cdi.Ok() {
		return nil, fmt.Errorf("no creature display %d", display)
	}
	if extID := cdi.Int("ExtendedDisplayInfoID"); extID > 0 {
		if a, err := r.npcCharacter(cdi, extID); err == nil {
			return a, nil
		}
	}
	model, scale := r.displayModel(display)
	if model == 0 {
		return nil, fmt.Errorf("creature display %d has no model", display)
	}
	a := &appearance{Model: model, CreatureDisplayID: display, Scale: scale, Replaceable: map[int]int{}}
	var tex []int
	for i := 0; i < 3; i++ {
		f := cdi.Int(fmt.Sprintf("TextureVariationFileDataID_%d", i))
		tex = append(tex, f)
		if f > 0 {
			a.Replaceable[11+i] = f
		}
	}
	// wow.export also hands the same list to types 2 to 4 when the model uses them.
	if r.textureTypes != nil {
		for _, tt := range r.textureTypes(model) {
			if tt >= 2 && tt <= 4 && tex[tt-2] > 0 {
				a.Replaceable[tt] = tex[tt-2]
			}
		}
	}
	if len(a.Replaceable) == 0 {
		a.Replaceable = nil
	}

	ids, exact := r.candidateGeosets(model)
	a.GeosetsExact = exact
	cmd := r.t("CreatureModelData").Get(cdi.Int("ModelID"))
	if cmd.Int("CreatureGeosetDataID") > 0 {
		// The display picks its geosets: hide 1 to 899 and show the listed ones, each
		// (GeosetIndex+1)*100 + GeosetValue.
		extra := map[int]bool{}
		for _, row := range r.t("CreatureDisplayInfoGeosetData").Where("CreatureDisplayInfoID", display) {
			extra[(row.Int("GeosetIndex")+1)*100+row.Int("GeosetValue")] = true
		}
		for _, id := range ids {
			if extra[id] || ((id == 0 || id >= 900) && defaultGeoset(id)) {
				a.Geosets = append(a.Geosets, id)
			}
		}
	} else {
		for _, id := range ids {
			s := strconv.Itoa(id)
			if strings.HasSuffix(s, "0") || strings.HasSuffix(s, "01") {
				a.Geosets = append(a.Geosets, id)
			}
		}
	}
	return a, nil
}

func (r *appearanceResolver) npcCharacter(cdi db2Row, extID int) (*appearance, error) {
	ext := r.t("CreatureDisplayInfoExtra").Get(extID)
	if !ext.Ok() {
		return nil, fmt.Errorf("no CreatureDisplayInfoExtra %d", extID)
	}
	race, sex, class := ext.Int("DisplayRaceID"), ext.Int("DisplaySexID"), ext.Int("DisplayClassID")
	var chrModel int
	for _, row := range r.t("ChrRaceXChrModel").Where("ChrRacesID", race) {
		if row.Int("Sex") == sex {
			chrModel = row.Int("ChrModelID")
		}
	}
	if chrModel == 0 {
		return nil, fmt.Errorf("no character model for race %d sex %d", race, sex)
	}
	choices := map[int]int{}
	for _, row := range r.t("CreatureDisplayInfoOption").Where("CreatureDisplayInfoExtraID", extID) {
		choices[row.Int("ChrCustomizationOptionID")] = row.Int("ChrCustomizationChoiceID")
	}
	bakeRes := ext.Int("HDBakeMaterialResourcesID")
	if bakeRes == 0 {
		bakeRes = ext.Int("BakeMaterialResourcesID")
	}
	bake := 0
	if files := r.textureFdids(bakeRes); len(files) > 0 {
		bake = files[len(files)-1]
	}
	bySlot := map[int]int{}
	for _, row := range r.t("NPCModelItemSlotDisplayInfo").Where("NpcModelID", extID) {
		if slot, ok := npcSlots[row.Int("ItemSlot")]; ok && row.Int("ItemDisplayInfoID") != 0 {
			bySlot[slot] = row.Int("ItemDisplayInfoID")
		}
	}
	var items []equipped
	for _, slot := range appearanceSlots {
		if id, ok := bySlot[slot]; ok {
			if d := r.t("ItemDisplayInfo").Get(id); d.Ok() {
				items = append(items, equipped{slot: slot, display: d})
			}
		}
	}
	a, err := r.dress(chrModel, race, sex, class, choices, items, bake)
	if err != nil {
		return nil, err
	}
	a.CreatureDisplayID = cdi.Int("ID")
	if s := cdi.Float("CreatureModelScale"); s > 0 {
		a.Scale *= s
	}
	// An NPC's look is fixed, so the picker list is noise here.
	a.Options = nil
	return a, nil
}

// Reading the model files.

// m2Chunks reads a model's MD20 header and its first skin's FileDataID.
func m2Chunks(data []byte) (md20 []byte, skin int, ok bool) {
	if len(data) < 8 || string(data[:4]) != "MD21" {
		return nil, 0, false
	}
	for pos := 0; pos+8 <= len(data); {
		id := string(data[pos : pos+4])
		size := int(binary.LittleEndian.Uint32(data[pos+4:]))
		body := data[pos+8 : min(len(data), pos+8+size)]
		switch id {
		case "MD21":
			md20 = body
		case "SFID":
			if len(body) >= 4 {
				skin = int(binary.LittleEndian.Uint32(body))
			}
		}
		pos += 8 + size
	}
	return md20, skin, md20 != nil
}

// m2TextureTypes reads the texture types from an MD20 header (the textures array at 0x50, 16
// bytes each with the type first).
func m2TextureTypes(md20 []byte) []int {
	if len(md20) < 0x58 {
		return nil
	}
	n := int(binary.LittleEndian.Uint32(md20[0x50:]))
	off := int(binary.LittleEndian.Uint32(md20[0x54:]))
	var out []int
	for i := 0; i < n && off+i*16+4 <= len(md20); i++ {
		out = append(out, int(binary.LittleEndian.Uint32(md20[off+i*16:])))
	}
	return out
}

// skinGeosetIDs reads the submesh ids (geosets) of a .skin file. Its header is the magic and
// four arrays (vertices, indices, bones, submeshes) of count and offset. Each submesh is 48
// bytes with the id first.
func skinGeosetIDs(data []byte) []int {
	if len(data) < 36 || string(data[:4]) != "SKIN" {
		return nil
	}
	n := int(binary.LittleEndian.Uint32(data[28:]))
	off := int(binary.LittleEndian.Uint32(data[32:]))
	seen := map[int]bool{}
	var out []int
	for i := 0; i < n && off+i*48+2 <= len(data); i++ {
		id := int(binary.LittleEndian.Uint16(data[off+i*48:]))
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	sort.Ints(out)
	return out
}

// fileReader reads a client file by FileDataID.
type fileReader func(fdid int) ([]byte, error)

func newAppearanceResolver(db *db2Tables, read fileReader) *appearanceResolver {
	r := &appearanceResolver{db: db}
	if read == nil {
		return r
	}
	r.skin = func(model int) []int {
		data, err := read(model)
		if err != nil {
			return nil
		}
		_, skin, ok := m2Chunks(data)
		if !ok || skin == 0 {
			return nil
		}
		sd, err := read(skin)
		if err != nil {
			return nil
		}
		return skinGeosetIDs(sd)
	}
	r.textureTypes = func(model int) []int {
		data, err := read(model)
		if err != nil {
			return nil
		}
		md20, _, ok := m2Chunks(data)
		if !ok {
			return nil
		}
		return m2TextureTypes(md20)
	}
	return r
}

// appearance gives the server's resolver, which reads model files through the asset store
// (extracting them on a miss) when there is one.
func (s *server) appearance() (*appearanceResolver, error) {
	if s.db2 == nil {
		return nil, errors.New("no WoW client, so no game tables")
	}
	var read fileReader
	if s.assets != nil {
		read = func(fdid int) ([]byte, error) {
			if err := s.assets.get(fdid); err != nil {
				return nil, err
			}
			return os.ReadFile(s.assets.path(fdid))
		}
	}
	return newAppearanceResolver(s.db2, read), nil
}

// parsePairs reads "a:b,c:d" into a map.
func parsePairs(s string) (map[int]int, error) {
	out := map[int]int{}
	if s == "" {
		return out, nil
	}
	for _, part := range strings.Split(s, ",") {
		k, v, ok := strings.Cut(strings.TrimSpace(part), ":")
		if !ok {
			return nil, fmt.Errorf("%q isn't key:value", part)
		}
		ki, err1 := strconv.Atoi(k)
		vi, err2 := strconv.Atoi(v)
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("%q isn't two numbers", part)
		}
		out[ki] = vi
	}
	return out, nil
}

// handleAppearanceCharacter answers
// /api/appearance/character?race=2&sex=0&class=7&items=1:12345,3:...&choices=option:choice,...
func (s *server) handleAppearanceCharacter(w http.ResponseWriter, r *http.Request) {
	res, err := s.appearance()
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	q := r.URL.Query()
	req := characterRequest{}
	for name, dst := range map[string]*int{"race": &req.Race, "sex": &req.Sex, "class": &req.Class} {
		if v := q.Get(name); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil {
				writeError(w, http.StatusBadRequest, fmt.Errorf("bad %s", name))
				return
			}
			*dst = n
		}
	}
	if req.Race == 0 {
		writeError(w, http.StatusBadRequest, errors.New("race is required"))
		return
	}
	if req.Items, err = parsePairs(q.Get("items")); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("items: %w", err))
		return
	}
	if req.Choices, err = parsePairs(q.Get("choices")); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("choices: %w", err))
		return
	}
	a, err := res.character(req)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, a)
}

// handleAppearanceCreature answers /api/appearance/creature/{displayId}. "dummy" stands for the
// room's training dummy.
func (s *server) handleAppearanceCreature(w http.ResponseWriter, r *http.Request) {
	res, err := s.appearance()
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	id := dummyDisplayID
	if v := r.PathValue("displayId"); v != "dummy" {
		if id, err = strconv.Atoi(v); err != nil {
			writeError(w, http.StatusBadRequest, errors.New("bad display id"))
			return
		}
	}
	a, err := res.creature(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, a)
}
