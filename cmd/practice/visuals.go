//go:build practice

package main

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"
)

// Spell visuals, weapon glows, melee sounds and unit sounds, read from the client's game tables
// (docs/practice_room.md, "Spell visuals and imbue glows").
//
// A spell names its look through SpellXSpellVisual. The SpellVisual then lists events (precast,
// cast, impact, aura and so on), and each event plays a visual kit on a unit. A kit is a list of
// effects: an animation for the unit, a sound, or an effect model hung on one of the unit's
// attachment points. We resolve all of it down to file ids (FileDataIDs), so the page only has to
// load models and sounds and play them at the right time.
//
// What the event and target numbers mean isn't documented anywhere we could find, so they come
// from reading the tables:
//
//   - StartEvent and EndEvent: 1 precast start, 2 precast end, 3 cast, 4 travel start, 5 travel
//     end, 6 impact, 7 aura start, 8 aura end, 9 area start, 10 area end, 11 channel start,
//     12 channel end, 13 one shot (the kit plays once and needs no end).
//   - TargetType: 1 the caster, 2 the target, 3 the ground, 4 the unit the spell hit, 5 the
//     missile. Classic's data mostly lists an impact or aura kit twice, once for the caster and
//     once for the target, so that it also shows when a spell lands on its own caster. We keep
//     the target one and drop the caster copy when it adds nothing (see dedupeSelf).
//   - SpellVisualKitEffect.EffectType: 1 procedural effect, 2 model attach, 3 camera effect,
//     5 sound kit, 6 animation, 10 a sound of the unit's own (by sound type).

type visualSound struct {
	FDID   int     `json:"fdid"`
	Volume float64 `json:"volume"`
	// Kit groups the entries of one sound kit. The client plays one entry per kit, picked at
	// random with Weight.
	Kit    int     `json:"kit"`
	Weight float64 `json:"weight,omitempty"`
	// Loop is the kit's looping flag (0x200), set on the precast and missile sounds.
	Loop bool `json:"loop,omitempty"`
}

type visualModel struct {
	FDID int `json:"fdid,omitempty"`
	// CreatureDisplay is set when the effect model is a creature display rather than a file.
	CreatureDisplay int        `json:"creatureDisplay,omitempty"`
	Attachment      int        `json:"attachment"`
	Scale           float64    `json:"scale"`
	Offset          [3]float64 `json:"offset"`
	Yaw             float64    `json:"yaw"`
	Pitch           float64    `json:"pitch,omitempty"`
	Roll            float64    `json:"roll,omitempty"`
	Alpha           float64    `json:"alpha,omitempty"`
	DelayMs         int        `json:"delayMs,omitempty"`
	// Anim is the effect model's own animation (AnimationData ids), not the unit's.
	Anim *visualAnim `json:"anim,omitempty"`
}

// visualAnim is an animation by AnimationData id: Start plays once, then Loop until the event
// ends. A null field means none.
type visualAnim struct {
	Start   *int `json:"start"`
	Loop    *int `json:"loop"`
	End     *int `json:"end"`
	AnimKit int  `json:"animKit,omitempty"`
}

type visualEvent struct {
	Event  string `json:"event"`
	Until  string `json:"until,omitempty"`
	Target string `json:"target"`
	// TargetType is the table's raw value, for when Target's reading turns out wrong.
	TargetType int `json:"targetType"`
	// SelfOnly marks a caster copy of an impact or aura kit that differs from the target one.
	// It plays only when the spell lands on its own caster.
	SelfOnly   bool          `json:"selfOnly,omitempty"`
	DelayMs    int           `json:"delayMs"`
	DelayMaxMs int           `json:"delayMaxMs,omitempty"`
	Kit        int           `json:"kit"`
	Sounds     []visualSound `json:"sounds"`
	Models     []visualModel `json:"models"`
	Anim       *visualAnim   `json:"anim,omitempty"`
	// UnitSounds are sound types of the unit's own sound set (EffectType 10).
	UnitSounds []int `json:"unitSounds,omitempty"`
	// Unsupported lists effects we don't resolve yet, like "procedural 374235" or "camera 1096".
	Unsupported []string `json:"unsupported,omitempty"`
}

type visualMissile struct {
	Model visualModel `json:"model"`
	// Models has every missile of the set when there is more than one.
	Models                []visualModel `json:"models,omitempty"`
	Speed                 float64       `json:"speed"`
	Attachment            int           `json:"attachment"`
	DestinationAttachment int           `json:"destinationAttachment"`
	Sound                 []visualSound `json:"sound"`
	CastOffset            [3]float64    `json:"castOffset"`
	ImpactOffset          [3]float64    `json:"impactOffset"`
	Motion                int           `json:"motion,omitempty"`
}

type visualTotem struct {
	Creature int    `json:"creature"`
	Element  string `json:"element"`
	Display  int    `json:"display"`
	Model    int    `json:"model"`
	// Source says where Display comes from. The client tables don't link a totem creature to
	// its display, so it is the element's display that Classic's totems use.
	Source string `json:"source"`
}

type spellVisual struct {
	Spell  int `json:"spell"`
	Visual int `json:"visual"`
	// Self is set when every effect of the spell lands on the caster (Lightning Shield, the
	// imbues, Ghost Wolf). Then the "target" events play on the caster.
	Self           bool            `json:"self"`
	Anims          map[string]*int `json:"anims"`
	Events         []visualEvent   `json:"events"`
	Missile        *visualMissile  `json:"missile"`
	AnimEventSound []visualSound   `json:"animEventSound,omitempty"`
	Totem          *visualTotem    `json:"totem,omitempty"`
}

var visualEventNames = map[int]string{
	1: "precast", 2: "precastEnd", 3: "cast", 4: "travel", 5: "travelEnd", 6: "impact", 7: "aura",
	8: "auraEnd", 9: "area", 10: "areaEnd", 11: "channel", 12: "channelEnd", 13: "oneshot",
}

func visualEventName(id int) string {
	if n, ok := visualEventNames[id]; ok {
		return n
	}
	return "event" + strconv.Itoa(id)
}

var visualTargetNames = map[int]string{1: "caster", 2: "target", 3: "area", 4: "target", 5: "missile"}

// visuals resolves and caches the routes' answers.
type visuals struct {
	db *db2Tables

	mu     sync.Mutex
	spells map[int]*spellVisual
}

func newVisuals(db *db2Tables) *visuals {
	return &visuals{db: db, spells: map[int]*spellVisual{}}
}

var (
	visualsMu    sync.Mutex
	visualsByDB2 = map[*db2Tables]*visuals{}
)

// visualsFor keeps one resolver per table set, so the routes need no field on server.
func visualsFor(db *db2Tables) *visuals {
	visualsMu.Lock()
	defer visualsMu.Unlock()
	v, ok := visualsByDB2[db]
	if !ok {
		v = newVisuals(db)
		visualsByDB2[db] = v
	}
	return v
}

func optAnim(id int) *int {
	if id < 0 {
		return nil
	}
	return &id
}

// soundKit gives a sound kit's entries.
func (v *visuals) soundKit(id int) []visualSound {
	if id <= 0 {
		return nil
	}
	kit := v.db.MustTable("SoundKit").Get(id)
	volume := 1.0
	loop := false
	if kit.Ok() {
		volume = kit.Float("VolumeFloat")
		loop = kit.Int("Flags")&0x200 != 0
	}
	var out []visualSound
	for _, e := range v.db.MustTable("SoundKitEntry").Where("SoundKitID", id) {
		if e.Int("FileDataID") == 0 {
			continue
		}
		vol := e.Float("Volume")
		if vol == 0 {
			vol = 1
		}
		out = append(out, visualSound{FDID: e.Int("FileDataID"), Volume: round3(volume * vol), Kit: id, Weight: e.Float("Frequency"), Loop: loop})
	}
	return out
}

func round3(f float64) float64 {
	return float64(int(f*1000+0.5)) / 1000
}

// effectModel resolves a SpellVisualEffectName to a model.
func (v *visuals) effectModel(id int) (visualModel, bool) {
	row := v.db.MustTable("SpellVisualEffectName").Get(id)
	if !row.Ok() {
		return visualModel{}, false
	}
	m := visualModel{FDID: row.Int("ModelFileDataID"), Scale: row.Float("Scale"), Alpha: row.Float("Alpha")}
	if m.Scale == 0 {
		m.Scale = 1
	}
	if m.Alpha == 1 {
		m.Alpha = 0
	}
	// Type 2 names a creature display in GenericID. Type 1 (an item) has no model file here.
	switch row.Int("Type") {
	case 2:
		m.CreatureDisplay = row.Int("GenericID")
		if m.FDID == 0 {
			m.FDID = v.displayModel(m.CreatureDisplay)
		}
	}
	return m, m.FDID != 0 || m.CreatureDisplay != 0
}

// displayModel gives a creature display's model file.
func (v *visuals) displayModel(display int) int {
	d := v.db.MustTable("CreatureDisplayInfo").Get(display)
	if !d.Ok() {
		return 0
	}
	return v.db.MustTable("CreatureModelData").Get(d.Int("ModelID")).Int("FileDataID")
}

// kitContent fills an event with one visual kit's effects.
func (v *visuals) kitContent(ev *visualEvent, kitID int) {
	kit := v.db.MustTable("SpellVisualKit").Get(kitID)
	if kit.Ok() {
		ev.DelayMs += kit.Int("DelayMin")
		if max := kit.Int("DelayMax"); max > kit.Int("DelayMin") {
			ev.DelayMaxMs = ev.DelayMs + max - kit.Int("DelayMin")
		}
	}
	effects := v.db.MustTable("SpellVisualKitEffect").Where("ParentSpellVisualKitID", kitID)
	sort.Slice(effects, func(i, j int) bool { return effects[i].Int("ID") < effects[j].Int("ID") })
	for _, e := range effects {
		id := e.Int("Effect")
		switch e.Int("EffectType") {
		case 2:
			a := v.db.MustTable("SpellVisualKitModelAttach").Get(id)
			if !a.Ok() {
				continue
			}
			m, ok := v.effectModel(a.Int("SpellVisualEffectNameID"))
			if !ok {
				continue
			}
			if s := a.Float("Scale"); s != 0 {
				m.Scale = round3(m.Scale * s)
			}
			m.Attachment = a.Int("AttachmentID")
			m.Offset = [3]float64{a.Float("Offset_0"), a.Float("Offset_1"), a.Float("Offset_2")}
			m.Yaw, m.Pitch, m.Roll = a.Float("Yaw"), a.Float("Pitch"), a.Float("Roll")
			m.DelayMs = int(a.Float("StartDelay") * 1000)
			if a.Int("StartAnimID") >= 0 || a.Int("AnimID") >= 0 || a.Int("EndAnimID") >= 0 || a.Int("AnimKitID") > 0 {
				m.Anim = &visualAnim{Start: optAnim(a.Int("StartAnimID")), Loop: optAnim(a.Int("AnimID")), End: optAnim(a.Int("EndAnimID")), AnimKit: a.Int("AnimKitID")}
			}
			ev.Models = append(ev.Models, m)
		case 5:
			ev.Sounds = append(ev.Sounds, v.soundKit(id)...)
		case 6:
			a := v.db.MustTable("SpellVisualAnim").Get(id)
			if a.Ok() {
				ev.Anim = &visualAnim{Start: optAnim(a.Int("InitialAnimID")), Loop: optAnim(a.Int("LoopAnimID")), AnimKit: a.Int("AnimKitID")}
			}
		case 10:
			ev.UnitSounds = append(ev.UnitSounds, id)
		case 1:
			ev.Unsupported = append(ev.Unsupported, fmt.Sprintf("procedural %d", id))
		case 3:
			ev.Unsupported = append(ev.Unsupported, fmt.Sprintf("camera %d", id))
		default:
			ev.Unsupported = append(ev.Unsupported, fmt.Sprintf("type%d %d", e.Int("EffectType"), id))
		}
	}
	if ev.Sounds == nil {
		ev.Sounds = []visualSound{}
	}
	if ev.Models == nil {
		ev.Models = []visualModel{}
	}
}

// pickVisual chooses the spell's SpellXSpellVisual row: the normal difficulty with no viewer or
// caster conditions, then the highest priority.
func (v *visuals) pickVisual(spell int) int {
	rows := v.db.MustTable("SpellXSpellVisual").Where("SpellID", spell)
	best, bestScore := 0, -1<<30
	for _, r := range rows {
		score := r.Int("Priority")
		if r.Int("DifficultyID") != 0 {
			score -= 1000
		}
		for _, c := range []string{"ViewerUnitConditionID", "ViewerPlayerConditionID", "CasterUnitConditionID", "CasterPlayerConditionID"} {
			if r.Int(c) != 0 {
				score -= 100
			}
		}
		if score > bestScore {
			best, bestScore = r.Int("SpellVisualID"), score
		}
	}
	return best
}

// spellIsSelf says whether every effect of the spell targets the caster (implicit target 1).
func (v *visuals) spellIsSelf(spell int) bool {
	effects := v.db.MustTable("SpellEffect").Where("SpellID", spell)
	if len(effects) == 0 {
		return false
	}
	for _, e := range effects {
		if t := e.Int("ImplicitTarget_0"); t != 1 && t != 0 {
			return false
		}
	}
	return true
}

func (v *visuals) Spell(spell int) (*spellVisual, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if sv, ok := v.spells[spell]; ok {
		return sv, nil
	}
	visualID := v.pickVisual(spell)
	if visualID == 0 {
		return nil, fmt.Errorf("spell %d has no visual", spell)
	}
	vis := v.db.MustTable("SpellVisual").Get(visualID)
	sv := &spellVisual{Spell: spell, Visual: visualID, Self: v.spellIsSelf(spell), Anims: map[string]*int{"precast": nil, "cast": nil, "channel": nil, "ready": nil}, Events: []visualEvent{}}

	rows := v.db.MustTable("SpellVisualEvent").Where("SpellVisualID", visualID)
	sort.Slice(rows, func(i, j int) bool { return rows[i].Int("ID") < rows[j].Int("ID") })
	for _, r := range rows {
		start, end, tt := r.Int("StartEvent"), r.Int("EndEvent"), r.Int("TargetType")
		ev := visualEvent{Event: visualEventName(start), Target: visualTargetNames[tt], TargetType: tt, DelayMs: r.Int("StartMinOffsetMs"), Kit: r.Int("SpellVisualKitID")}
		if ev.Target == "" {
			ev.Target = "target"
		}
		if max := r.Int("StartMaxOffsetMs"); max > ev.DelayMs {
			ev.DelayMaxMs = max
		}
		if end != 13 && end != start && end != 0 {
			ev.Until = visualEventName(end)
		}
		v.kitContent(&ev, ev.Kit)
		sv.Events = append(sv.Events, ev)
	}
	sv.Events = dedupeSelf(sv.Events)

	// The caster's animations, from its precast, cast and channel kits.
	for _, ev := range sv.Events {
		if ev.Target != "caster" || ev.Anim == nil {
			continue
		}
		key := map[string]string{"precast": "precast", "cast": "cast", "channel": "channel"}[ev.Event]
		if key == "" || sv.Anims[key] != nil {
			continue
		}
		if ev.Anim.Loop != nil {
			sv.Anims[key] = ev.Anim.Loop
		} else {
			sv.Anims[key] = ev.Anim.Start
		}
	}

	sv.AnimEventSound = v.soundKit(vis.Int("AnimEventSoundID"))
	sv.Missile = v.missile(spell, vis)
	sv.Totem = v.totem(spell)
	v.spells[spell] = sv
	return sv, nil
}

// dedupeSelf drops the caster copy of an impact or aura kit when the target's kit for the same
// event already has everything in it.
//
// Classic's data lists Lightning Bolt's impact twice: kit 283 on the target (the spark model and
// its sound) and kit 261808 on the caster (the same spark model, no sound). Read literally the
// shaman would spark on every hit. We read the caster copy as "when the spell lands on its
// caster", which the target kit covers too, so the copy would only double the model. A caster
// copy with something new stays, marked SelfOnly.
func dedupeSelf(events []visualEvent) []visualEvent {
	hitEvent := map[string]bool{"impact": true, "aura": true, "auraEnd": true, "travelEnd": true}
	var out []visualEvent
	for i, ev := range events {
		if ev.TargetType != 1 || !hitEvent[ev.Event] {
			out = append(out, ev)
			continue
		}
		var twins []visualEvent
		for j, o := range events {
			if j != i && o.Event == ev.Event && (o.TargetType == 2 || o.TargetType == 4) {
				twins = append(twins, o)
			}
		}
		if len(twins) == 0 {
			out = append(out, ev)
			continue
		}
		if !coveredBy(ev, twins) {
			ev.SelfOnly = true
			out = append(out, ev)
		}
	}
	return out
}

func coveredBy(ev visualEvent, twins []visualEvent) bool {
	type modelKey struct{ fdid, attach int }
	models := map[modelKey]bool{}
	sounds := map[int]bool{}
	anims := map[string]bool{}
	animKey := func(a *visualAnim) string {
		if a == nil {
			return ""
		}
		k := ""
		for _, p := range []*int{a.Start, a.Loop, a.End} {
			if p == nil {
				k += "-,"
			} else {
				k += strconv.Itoa(*p) + ","
			}
		}
		return k + strconv.Itoa(a.AnimKit)
	}
	for _, t := range twins {
		for _, m := range t.Models {
			models[modelKey{m.FDID, m.Attachment}] = true
		}
		for _, s := range t.Sounds {
			sounds[s.FDID] = true
		}
		anims[animKey(t.Anim)] = true
	}
	for _, m := range ev.Models {
		if !models[modelKey{m.FDID, m.Attachment}] {
			return false
		}
	}
	for _, s := range ev.Sounds {
		if !sounds[s.FDID] {
			return false
		}
	}
	if ev.Anim != nil && !anims[animKey(ev.Anim)] {
		return false
	}
	return len(ev.UnitSounds) == 0
}

// missile resolves the visual's missile set. The speed is the spell's (SpellMisc.Speed, yards
// per second), or the effect model's base speed when the spell has none.
func (v *visuals) missile(spell int, vis db2Row) *visualMissile {
	set := vis.Int("SpellVisualMissileSetID")
	if set == 0 {
		return nil
	}
	rows := v.db.MustTable("SpellVisualMissile").Where("SpellVisualMissileSetID", set)
	if len(rows) == 0 {
		return nil
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Int("ID") < rows[j].Int("ID") })
	var out *visualMissile
	for _, r := range rows {
		m, ok := v.effectModel(r.Int("SpellVisualEffectNameID"))
		if !ok {
			continue
		}
		attach, dest := r.Int("Attachment"), r.Int("DestinationAttachment")
		if attach < 0 {
			attach = vis.Int("MissileAttachment")
		}
		if dest < 0 {
			dest = vis.Int("MissileDestinationAttachment")
		}
		m.Attachment = attach
		if out == nil {
			speed := v.db.MustTable("SpellMisc").First("SpellID", spell).Float("Speed")
			if speed == 0 {
				speed = v.db.MustTable("SpellVisualEffectName").Get(r.Int("SpellVisualEffectNameID")).Float("BaseMissileSpeed")
			}
			out = &visualMissile{
				Model:                 m,
				Speed:                 speed,
				Attachment:            attach,
				DestinationAttachment: dest,
				Sound:                 v.soundKit(r.Int("SoundEntriesID")),
				CastOffset:            [3]float64{r.Float("CastOffset_0"), r.Float("CastOffset_1"), r.Float("CastOffset_2")},
				ImpactOffset:          [3]float64{r.Float("ImpactOffset_0"), r.Float("ImpactOffset_1"), r.Float("ImpactOffset_2")},
				Motion:                r.Int("SpellMissileMotionID"),
			}
			if out.Sound == nil {
				out.Sound = []visualSound{}
			}
		}
		out.Models = append(out.Models, m)
	}
	if out != nil && len(out.Models) < 2 {
		out.Models = nil
	}
	return out
}

// Classic's player totems use one display per element. The client tables have no row for the
// totem creatures (they live on the server), so this is the link we can't read from them.
var totemElements = map[int]struct {
	name    string
	display int
}{
	1: {"fire", 4589},
	2: {"earth", 4588},
	3: {"water", 4587},
	4: {"air", 4590},
}

// totem gives the totem a spell summons: SpellEffect 28 (summon) names the creature and the
// SummonProperties row, whose Slot is the totem's element.
func (v *visuals) totem(spell int) *visualTotem {
	for _, e := range v.db.MustTable("SpellEffect").Where("SpellID", spell) {
		if e.Int("Effect") != 28 {
			continue
		}
		props := v.db.MustTable("SummonProperties").Get(e.Int("EffectMiscValue_1"))
		el, ok := totemElements[props.Int("Slot")]
		if !props.Ok() || !ok {
			continue
		}
		t := &visualTotem{Creature: e.Int("EffectMiscValue_0"), Element: el.name, Display: el.display, Model: v.displayModel(el.display), Source: "element"}
		// A Creature row, when the build has one, names the real display.
		if c := v.db.MustTable("Creature").Get(t.Creature); c.Ok() && c.Int("DisplayID_0") != 0 {
			t.Display, t.Source = c.Int("DisplayID_0"), "Creature"
			t.Model = v.displayModel(t.Display)
		}
		return t
	}
	return nil
}

type enchantVisual struct {
	Enchant    int           `json:"enchant"`
	Name       string        `json:"name"`
	ItemVisual int           `json:"itemVisual"`
	Models     []visualModel `json:"models"`
	// Kits are the ItemVisuals' visual kits, when it has any (none of the shaman imbues do).
	Kits []visualEvent `json:"kits,omitempty"`
}

// Enchant gives a weapon enchant's glow. ItemVisuals has five model slots, and slot i hangs on
// the weapon model's attachment i (the item visual points 0 to 4). The shaman imbues use slot 3
// only, most other glows fill all five.
func (v *visuals) Enchant(id int) (*enchantVisual, error) {
	e := v.db.MustTable("SpellItemEnchantment").Get(id)
	if !e.Ok() {
		return nil, fmt.Errorf("no enchant %d", id)
	}
	out := &enchantVisual{Enchant: id, Name: e.Str("Name_lang"), ItemVisual: e.Int("ItemVisual"), Models: []visualModel{}}
	iv := v.db.MustTable("ItemVisuals").Get(out.ItemVisual)
	if !iv.Ok() {
		return out, nil
	}
	for i := 0; i < 5; i++ {
		if f := iv.Int(fmt.Sprintf("ModelFileID_%d", i)); f != 0 {
			out.Models = append(out.Models, visualModel{FDID: f, Attachment: i, Scale: 1})
		}
	}
	for _, col := range []string{"SpellVisualKitId1", "SpellVisualKitId2"} {
		if k := iv.Int(col); k != 0 {
			ev := visualEvent{Event: "state", Target: "weapon", Kit: k}
			v.kitContent(&ev, k)
			out.Kits = append(out.Kits, ev)
		}
	}
	return out, nil
}

// Melee gives the swing and hit sounds for a weapon type (ItemSubClass of class 2). The swing
// size comes from ItemSubClass.WeaponSwingSize (0 dagger and fist, 1 one-hand, 2 two-hand). The
// hit sounds are from WeaponImpactSounds, whose arrays are by the struck material: 0 flesh,
// 1 chain, 2 plate, 3 metal shield, 4 wooden shield, 5 and 6 parries, 7 wood, 8 stone,
// 9 ethereal (read from the sound file names).
func (v *visuals) Melee(subclass int, twoHanded bool, material int) map[string][]visualSound {
	size := -1
	for _, r := range v.db.MustTable("ItemSubClass").Where("ClassID", 2) {
		if r.Int("SubClassID") == subclass {
			size = r.Int("WeaponSwingSize")
		}
	}
	if size < 0 {
		size = 1
		if twoHanded {
			size = 2
		}
	}
	out := map[string][]visualSound{"swing": {}, "swingCrit": {}, "impact": {}, "impactCrit": {}}
	for _, r := range v.db.MustTable("WeaponSwingSounds2").Where("SwingType", size) {
		key := "swing"
		if r.Int("Crit") != 0 {
			key = "swingCrit"
		}
		out[key] = append(out[key], v.soundKit(r.Int("SoundID"))...)
	}
	// Each weapon type has a row for wooden and metal weapons (ParrySoundType 0 and 1). Their
	// hit sounds on flesh are the same, so we take the metal one when there is one.
	var row db2Row
	for _, r := range v.db.MustTable("WeaponImpactSounds").Where("WeaponSubClassID", subclass) {
		if !row.Ok() || r.Int("ParrySoundType") == 1 {
			row = r
		}
	}
	if row.Ok() {
		out["impact"] = append(out["impact"], v.soundKit(row.Int(fmt.Sprintf("ImpactSoundID_%d", material)))...)
		out["impactCrit"] = append(out["impactCrit"], v.soundKit(row.Int(fmt.Sprintf("CritImpactSoundID_%d", material)))...)
	}
	return out
}

var unitSoundColumns = []struct{ key, col string }{
	{"attack", "SoundExertionID"},
	{"attackCrit", "SoundExertionCriticalID"},
	{"wound", "SoundInjuryID"},
	{"woundCrit", "SoundInjuryCriticalID"},
	{"woundCrushing", "SoundInjuryCrushingBlowID"},
	{"death", "SoundDeathID"},
	{"stun", "SoundStunID"},
	{"stand", "SoundStandID"},
	{"aggro", "SoundAggroID"},
	{"alert", "SoundAlertID"},
	{"jumpStart", "SoundJumpStartID"},
	{"jumpEnd", "SoundJumpEndID"},
	{"battleShout", "BattleShoutSoundID"},
	{"battleShoutCrit", "BattleShoutCriticalSoundID"},
	{"spellCast", "SpellCastDirectedSoundID"},
	{"loop", "LoopSoundID"},
}

type unitSounds struct {
	Display   int                      `json:"display"`
	SoundData int                      `json:"soundData"`
	Sounds    map[string][]visualSound `json:"sounds"`
	// Footsteps are by terrain (TerrainType names), from FootstepTerrainLookup.
	Footsteps map[string][]visualSound `json:"footsteps,omitempty"`
}

// Unit gives a creature display's own sounds. The display's SoundID wins over its model's.
func (v *visuals) Unit(display int) (*unitSounds, error) {
	d := v.db.MustTable("CreatureDisplayInfo").Get(display)
	if !d.Ok() {
		return nil, fmt.Errorf("no creature display %d", display)
	}
	id := d.Int("SoundID")
	if id == 0 {
		id = v.db.MustTable("CreatureModelData").Get(d.Int("ModelID")).Int("SoundID")
	}
	out := &unitSounds{Display: display, SoundData: id, Sounds: map[string][]visualSound{}}
	row := v.db.MustTable("CreatureSoundData").Get(id)
	if !row.Ok() {
		return out, nil
	}
	for _, c := range unitSoundColumns {
		if s := v.soundKit(row.Int(c.col)); len(s) > 0 {
			out.Sounds[c.key] = s
		}
	}
	for i := 0; i < 5; i++ {
		out.Sounds["fidget"] = append(out.Sounds["fidget"], v.soundKit(row.Int(fmt.Sprintf("SoundFidget_%d", i)))...)
	}
	for i := 0; i < 4; i++ {
		out.Sounds["customAttack"] = append(out.Sounds["customAttack"], v.soundKit(row.Int(fmt.Sprintf("CustomAttack_%d", i)))...)
	}
	for k, s := range out.Sounds {
		if len(s) == 0 {
			delete(out.Sounds, k)
		}
	}
	if foot := row.Int("SoundFootstepID"); foot != 0 {
		// FootstepTerrainLookup's TerrainSoundID is TerrainType's SoundID, which some terrains
		// share (Grass and Dusty Grass), so each lookup row can fill several names.
		names := map[int][]string{}
		terrain := v.db.MustTable("TerrainType")
		for i := 0; i < terrain.Len(); i++ {
			t := terrain.Row(i)
			names[t.Int("SoundID")] = append(names[t.Int("SoundID")], t.Str("TerrainDesc"))
		}
		out.Footsteps = map[string][]visualSound{}
		for _, r := range v.db.MustTable("FootstepTerrainLookup").Where("CreatureFootstepID", foot) {
			sounds := v.soundKit(r.Int("SoundID"))
			for _, name := range names[r.Int("TerrainSoundID")] {
				out.Footsteps[name] = append(out.Footsteps[name], sounds...)
			}
		}
	}
	return out, nil
}

// The shaman imbues' enchant ids, by rank, for the page to ask /api/visuals/enchant about.
func (v *visuals) imbueEnchants() map[string][]int {
	out := map[string][]int{}
	names := v.db.MustTable("SpellName")
	effects := v.db.MustTable("SpellEffect")
	for i := 0; i < effects.Len(); i++ {
		e := effects.Row(i)
		if e.Int("Effect") != 54 && e.Int("Effect") != 360 {
			continue
		}
		n := names.Get(e.Int("SpellID")).Str("Name_lang")
		switch n {
		case "Windfury Weapon", "Flametongue Weapon", "Rockbiter Weapon", "Frostbrand Weapon":
			out[n] = append(out[n], e.Int("EffectMiscValue_0"))
		}
	}
	for _, ids := range out {
		sort.Ints(ids)
	}
	return out
}

func (s *server) visualsOrError(w http.ResponseWriter) *visuals {
	if s.db2 == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("no game tables (no WoW client)"))
		return nil
	}
	return visualsFor(s.db2)
}

func pathInt(w http.ResponseWriter, r *http.Request, name string) (int, bool) {
	id, err := strconv.Atoi(r.PathValue(name))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("bad %s", name))
		return 0, false
	}
	return id, true
}

func (s *server) handleSpellVisual(w http.ResponseWriter, r *http.Request) {
	v := s.visualsOrError(w)
	id, ok := pathInt(w, r, "spellId")
	if v == nil || !ok {
		return
	}
	sv, err := v.Spell(id)
	if err != nil {
		// Many auras and procs have no visual. That's a normal answer, not a failed request
		// (which the browser's console would show as an error).
		writeJSON(w, nil)
		return
	}
	writeJSON(w, sv)
}

func (s *server) handleEnchantVisual(w http.ResponseWriter, r *http.Request) {
	v := s.visualsOrError(w)
	id, ok := pathInt(w, r, "enchantId")
	if v == nil || !ok {
		return
	}
	ev, err := v.Enchant(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, ev)
}

// handleImbueEnchants lists the shaman imbues' enchant ids by name.
func (s *server) handleImbueEnchants(w http.ResponseWriter, r *http.Request) {
	if v := s.visualsOrError(w); v != nil {
		writeJSON(w, v.imbueEnchants())
	}
}

var impactMaterials = map[string]int{"flesh": 0, "chain": 1, "plate": 2, "wood": 7, "stone": 8, "ethereal": 9}

func (s *server) handleMeleeSounds(w http.ResponseWriter, r *http.Request) {
	v := s.visualsOrError(w)
	if v == nil {
		return
	}
	q := r.URL.Query()
	// No weapon: fists swing light and hit like fist weapons (subclass 13).
	subclass := 13
	if q.Has("weaponSubclass") {
		n, err := strconv.Atoi(q.Get("weaponSubclass"))
		if err != nil {
			writeError(w, http.StatusBadRequest, errors.New("bad weaponSubclass"))
			return
		}
		subclass = n
	}
	material, ok := impactMaterials[q.Get("material")]
	if q.Get("material") != "" && !ok {
		writeError(w, http.StatusBadRequest, errors.New("material is flesh, chain, plate, wood, stone or ethereal"))
		return
	}
	writeJSON(w, v.Melee(subclass, q.Get("twoHanded") == "true", material))
}

func (s *server) handleUnitSounds(w http.ResponseWriter, r *http.Request) {
	v := s.visualsOrError(w)
	id, ok := pathInt(w, r, "displayId")
	if v == nil || !ok {
		return
	}
	us, err := v.Unit(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	// The sound types go at the top level ("attack", "wound", ...) next to the ids.
	out := map[string]any{"display": us.Display, "soundData": us.SoundData}
	for k, s := range us.Sounds {
		out[k] = s
	}
	if us.Footsteps != nil {
		out["footsteps"] = us.Footsteps
	}
	writeJSON(w, out)
}
