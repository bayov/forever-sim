package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/wowsims/classic/sim/compsearch"
	"github.com/wowsims/classic/sim/core/proto"
)

// tableFile is the page's copy of the search's table of who gives what (sim/compsearch/table.go),
// relative to the checkout.
const tableFile = "ui/raid_sim/data/table.json"

type tableEffect struct {
	Name   string `json:"name"`
	Label  string `json:"label,omitempty"`
	Icon   string `json:"icon"`
	Debuff bool   `json:"debuff,omitempty"`
	// From is who can give it, for the page's Who gives what table: the presets, the pool's DPS builds and
	// the duties.
	From string `json:"from"`
}

type tablePreset struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Role              string   `json:"role"`
	Class             string   `json:"class"`
	Talents           string   `json:"talents"`
	Icon              string   `json:"icon"`
	Gives             []string `json:"gives"`
	Summary           string   `json:"summary"`
	Paladin           bool     `json:"paladin,omitempty"`
	Shaman            bool     `json:"shaman,omitempty"`
	Package           string   `json:"package,omitempty"`
	PowerInfusion     bool     `json:"powerInfusion,omitempty"`
	Innervate         bool     `json:"innervate,omitempty"`
	InnervateOptional bool     `json:"innervateOptional,omitempty"`
}

type named struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Label  string `json:"label,omitempty"`
	Icon   string `json:"icon,omitempty"`
	Effect string `json:"effect,omitempty"`
	// A totem's element, or the class that holds a duty.
	Of string `json:"of,omitempty"`
}

// table is the page's table file's contents.
func table(pool *compsearch.Pool) []byte {
	out := map[string]any{}

	var effects []tableEffect
	for _, e := range compsearch.Effects {
		effects = append(effects, tableEffect{Name: e.Name, Label: e.Label, Icon: e.Icon, Debuff: e.Debuff, From: giversOf(pool, e)})
	}
	out["effects"] = effects

	var presets []tablePreset
	for _, p := range compsearch.Presets {
		tp := tablePreset{ID: p.ID.String(), Name: p.Name, Role: p.Role.String(), Class: p.Class.String(), Talents: p.Talents,
			Icon: p.Icon, Gives: presetGives(p), Summary: presetSummary(p), Paladin: p.Paladin(), Shaman: p.Shaman(), PowerInfusion: p.PI,
			Innervate: p.Innervate, InnervateOptional: p.InnervateOptional}
		if p.Package != proto.RaidPackage_RaidPackageNone {
			tp.Package = p.Package.String()
		}
		presets = append(presets, tp)
	}
	out["presets"] = presets

	var classes, blessings, judgements, totems, elements, packages, duties []named
	for _, c := range compsearch.Classes {
		classes = append(classes, named{ID: c.Class.String(), Name: c.Name, Icon: c.Icon})
	}
	for _, b := range compsearch.Blessings {
		blessings = append(blessings, named{ID: b.Blessing.String(), Name: b.Name, Icon: b.Icon})
	}
	for _, j := range compsearch.Judgements {
		judgements = append(judgements, named{ID: j.Judgement.String(), Name: j.Name, Icon: j.Icon, Effect: compsearch.GiftName(j.Gift)})
	}
	for _, t := range compsearch.Totems {
		n := named{ID: t.Totem.String(), Name: t.Name, Icon: t.Icon, Of: t.Element.String()}
		if t.Gift != 0 {
			n.Effect = compsearch.GiftName(t.Gift)
		}
		totems = append(totems, n)
	}
	for _, e := range compsearch.Elements {
		elements = append(elements, named{ID: e.Element.String(), Name: e.Name})
	}
	for _, p := range compsearch.Packages {
		packages = append(packages, named{ID: p.Package.String(), Name: p.Name})
	}
	for _, d := range compsearch.Duties {
		duties = append(duties, named{ID: d.Duty.String(), Name: d.Name, Label: d.Label, Icon: d.Icon,
			Effect: compsearch.GiftName(d.Gift), Of: d.Class.String()})
	}
	out["classes"] = classes
	out["blessings"] = blessings
	out["judgements"] = judgements
	out["totems"] = totems
	out["elements"] = elements
	out["packages"] = packages
	out["duties"] = duties
	out["specIcons"] = compsearch.SpecIcons
	out["icons"] = map[string]string{
		"powerInfusion": compsearch.PowerInfusionIcon,
		"innervate":     compsearch.InnervateIcon,
		"manaTide":      compsearch.ManaTideIcon,
	}

	data, err := json.MarshalIndent(out, "", "\t")
	if err != nil {
		fail(err)
	}
	return append(data, '\n')
}

// presetGives is every effect a preset can give: its own, and each choice it can make.
func presetGives(p *compsearch.Preset) []string {
	g := p.Gives
	if p.Paladin() {
		for _, j := range compsearch.Judgements {
			g |= j.Gift
		}
	}
	if p.Shaman() {
		for _, t := range compsearch.Totems {
			g |= t.Gift
		}
	}
	var out []string
	for _, e := range compsearch.Effects {
		if g.Has(e.Gift) {
			out = append(out, e.Name)
		}
	}
	return out
}

// presetSummary is what a preset gives in a line, its choices included, like "Greater Blessings,
// Salvation, Judgement of the Crusader or Wisdom".
func presetSummary(p *compsearch.Preset) string {
	label := func(e *compsearch.Effect) string {
		if e.Label != "" {
			return e.Label
		}
		return e.Name
	}
	fixed := p.Gives
	if p.Shaman() {
		fixed = 0 // its totems are listed by element below
	}
	var parts []string
	for _, e := range compsearch.Effects {
		if !fixed.Has(e.Gift) || e.Alias && e.Gift != compsearch.GiftBlessing {
			continue
		}
		// What its class casts for it, like a druid's Faerie Fire for the minor armor slot.
		name := label(e)
		for _, d := range compsearch.Duties {
			if d.Gift == e.Gift && d.Class == p.Class {
				name = d.Name
			}
		}
		parts = append(parts, name)
	}
	if p.Paladin() {
		var names []string
		for _, j := range compsearch.Judgements {
			names = append(names, j.Name)
		}
		parts = append(parts, "Judgement of "+orList(names))
	}
	if p.Shaman() {
		for _, el := range compsearch.Elements {
			var names []string
			for _, t := range compsearch.Totems {
				if t.Element == el.Element && t.Gift != 0 {
					names = append(names, t.Name)
				}
			}
			if len(names) > 0 {
				parts = append(parts, orList(names)+" Totem")
			}
		}
		parts = append(parts, compsearch.Packages[0].Name+" for its party")
		for _, pk := range compsearch.Packages {
			if pk.Package == p.Package {
				parts[len(parts)-1] = pk.Name + " for its party"
			}
		}
	}
	if p.PI {
		parts = append(parts, "Power Infusion")
	}
	if p.Innervate {
		parts = append(parts, "Innervate")
	}
	if p.InnervateOptional {
		parts = append(parts, "Innervate when switched on")
	}
	return strings.Join(parts, ", ")
}

// giversOf is who can give an effect, like "Discipline priest, Holy priest, any DPS priest".
func giversOf(pool *compsearch.Pool, e *compsearch.Effect) string {
	var who []string
	for _, p := range compsearch.Presets {
		for _, name := range presetGives(p) {
			if name == e.Name {
				who = append(who, p.Name)
			}
		}
	}

	// The DPS builds, by class and pet: "Balance or Feral druid", "any DPS mage", "Beast Mastery or
	// Survival hunter (Wolf)".
	type group struct {
		class proto.Class
		pet   string
	}
	specs := map[group][]string{}
	var order []group
	all := map[proto.Class]int{} // the class's base builds' specs
	classSpecs := map[proto.Class]map[string]bool{}
	for _, b := range pool.Bases {
		if classSpecs[b.Class] == nil {
			classSpecs[b.Class] = map[string]bool{}
		}
		if !classSpecs[b.Class][b.Spec] {
			classSpecs[b.Class][b.Spec] = true
			all[b.Class]++
		}
	}
	add := func(b *compsearch.Build, pet string) {
		if !b.Gives.Has(e.Gift) {
			return
		}
		g := group{b.Class, pet}
		if _, ok := specs[g]; !ok {
			order = append(order, g)
		}
		if !contains(specs[g], b.Spec) {
			specs[g] = append(specs[g], b.Spec)
		}
	}
	for _, b := range pool.Bases {
		add(b, "")
		// A pet variant counts only for what its pet gives.
		if !b.Gives.Has(e.Gift) {
			for _, pv := range b.Pets {
				add(pv, petName(pv))
			}
		}
	}
	sort.SliceStable(order, func(i, j int) bool { return classIndex(order[i].class) < classIndex(order[j].class) })
	for _, g := range order {
		name := strings.ToLower(className(g.class))
		var s string
		if g.pet == "" && len(specs[g]) == all[g.class] {
			s = "any DPS " + name
		} else {
			s = orList(specs[g]) + " " + name
		}
		if g.pet != "" {
			s += " (" + g.pet + ")"
		}
		who = append(who, s)
	}

	for _, d := range compsearch.Duties {
		if d.Gift != e.Gift || !poolHasDuty(pool, d.Duty) {
			continue
		}
		s := "a " + strings.ToLower(className(d.Class)) + " on its duty"
		if !contains(who, s) {
			who = append(who, s)
		}
	}
	if len(who) == 0 {
		return "Nobody in the pool"
	}
	text := strings.Join(who, ", ")
	return strings.ToUpper(text[:1]) + text[1:]
}

func poolHasDuty(pool *compsearch.Pool, d proto.Duty) bool {
	for _, b := range pool.Bases {
		if _, ok := b.Duties[d]; ok {
			return true
		}
	}
	return false
}

// petName is a pet variant's pet, from its name: "Beast Mastery (Wolf)" is a Wolf.
func petName(b *compsearch.Build) string {
	name := b.Name
	if i := strings.LastIndex(name, "("); i >= 0 && strings.HasSuffix(name, ")") {
		parts := strings.Split(name[i+1:len(name)-1], ", ")
		return parts[len(parts)-1]
	}
	return name
}

func className(c proto.Class) string {
	for _, r := range compsearch.Classes {
		if r.Class == c {
			return r.Name
		}
	}
	return strings.TrimPrefix(c.String(), "Class")
}

func classIndex(c proto.Class) int {
	for i, r := range compsearch.Classes {
		if r.Class == c {
			return i
		}
	}
	return len(compsearch.Classes)
}

func orList(s []string) string {
	switch len(s) {
	case 0:
		return ""
	case 1:
		return s[0]
	}
	return fmt.Sprintf("%s or %s", strings.Join(s[:len(s)-1], ", "), s[len(s)-1])
}

func contains(s []string, x string) bool {
	for _, y := range s {
		if y == x {
			return true
		}
	}
	return false
}
