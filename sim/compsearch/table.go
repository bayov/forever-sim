package compsearch

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// The raid sim's table of who gives what (docs/raid_sim.md, The provider table).
//
// The search, the resolver and the page read everything they know about classes from here: which effects
// a raid can have and how each one goes into a request, what each spec gives and needs, the blessings,
// totems, judgements and duties, and the tank and healer presets (presets.go). The page gets a copy as
// ui/raid_sim/data/table.json, which tools/raid_sim_data writes.
//
// When Forever changes a class, we edit these rows. A row can move an effect to another class, add a
// blessing, or take a totem away without touching the search. What stays in code is the three ways a
// raid shares its effects: a paladin's blessings go to each class, a shaman's totems go to its party, and
// warriors Sunder the boss at the pull. A new kind of sharing needs code.

// Effect is one raid-wide effect a member can give, in the order the page's strip shows them.
type Effect struct {
	Gift Gift
	// Name is the effect's name in buff records, required effects and the builds' "gives".
	Name string
	// Label is the name the page shows, when it differs.
	Label  string
	Icon   string
	Debuff bool
	// Alias marks a second name for another row's gift, like Salvation, which comes with every paladin.
	Alias bool
	// write puts the effect into a receiver's request.
	write func(w *buffWriter)
}

// buffWriter is the request a seat is being written into, and the spec receiving it.
type buffWriter struct {
	rb   *proto.RaidBuffs
	db   *proto.Debuffs
	spec *SpecRow
}

var regular = proto.TristateEffect_TristateEffectRegular

var Effects = []*Effect{
	{Gift: GiftFort, Name: "Fortitude", Label: "Prayer of Fortitude", Icon: "spell_holy_prayeroffortitude",
		write: func(w *buffWriter) { w.rb.PowerWordFortitude = regular }},
	{Gift: GiftGotW, Name: "Gift of the Wild", Icon: "spell_nature_regeneration",
		write: func(w *buffWriter) { w.rb.GiftOfTheWild = regular }},
	{Gift: GiftAB, Name: "Arcane Brilliance", Icon: "spell_holy_arcaneintellect",
		write: func(w *buffWriter) { w.rb.ArcaneBrilliance = true }},
	{Gift: GiftPoS, Name: "Prayer of Spirit", Icon: "spell_holy_prayerofspirit",
		write: func(w *buffWriter) { w.rb.DivineSpirit = true }},
	// Each class's blessings are the seat's own (Seat.Blessings), so this row writes nothing.
	{Gift: GiftBlessing, Name: "Blessings", Label: "Greater Blessings", Icon: "spell_magic_greaterblessingofkings"},
	// Forever's Salvation (spell 1038) is no longer a Blessing: any paladin gives it, and it lowers threat
	// only, so requiring it is requiring a paladin.
	{Gift: GiftBlessing, Name: "Salvation", Icon: "spell_holy_sealofsalvation", Alias: true},
	{Gift: GiftBattleShout, Name: "Battle Shout", Icon: "ability_warrior_battleshout",
		write: func(w *buffWriter) { w.rb.BattleShout = regular }},
	{Gift: GiftFuriousHowl, Name: "Furious Howl", Icon: "ability_hunter_pet_wolf",
		write: func(w *buffWriter) { w.rb.FuriousHowl = true }},
	{Gift: GiftTrueshot, Name: "Trueshot Aura", Icon: "ability_trueshot",
		write: func(w *buffWriter) { w.rb.TrueshotAura = true }},
	{Gift: GiftLotP, Name: "Leader of the Pack", Icon: "spell_nature_unyeildingstamina",
		write: func(w *buffWriter) { w.rb.LeaderOfThePack = true }},
	{Gift: GiftMoonkin, Name: "Moonkin Aura", Icon: "spell_nature_moonglow",
		write: func(w *buffWriter) { w.rb.MoonkinAura = true }},
	{Gift: GiftSoE, Name: "Strength of Earth", Label: "Strength of Earth Totem", Icon: "spell_nature_earthbindtotem",
		write: func(w *buffWriter) { w.rb.StrengthOfEarthTotem = regular }},
	{Gift: GiftGoA, Name: "Grace of Air", Label: "Grace of Air Totem", Icon: "spell_nature_invisibilitytotem",
		write: func(w *buffWriter) { w.rb.GraceOfAirTotem = regular }},
	// A receiver takes one weapon totem: Flametongue when its spec wants it, else Windfury.
	{Gift: GiftWF, Name: "Windfury Totem", Icon: "spell_nature_windfury",
		write: func(w *buffWriter) {
			if !w.spec.FlametongueTotem {
				w.rb.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffWindfury
			}
		}},
	{Gift: GiftFT, Name: "Flametongue Totem", Icon: "spell_nature_guardianward",
		write: func(w *buffWriter) {
			if w.spec.FlametongueTotem {
				w.rb.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffFlametongue
			}
		}},
	// The tank's Sunder ramp also counts the other warriors' pull Sunders, so Seat.Write writes it.
	{Gift: GiftSunderTank, Name: "Sunder upkeep", Label: "Sunder Armor", Icon: "ability_warrior_sunder", Debuff: true},
	{Gift: GiftMinorArmor, Name: "Minor armor slot", Label: "Faerie Fire or Curse of Recklessness", Icon: "spell_nature_faeriefire", Debuff: true,
		write: func(w *buffWriter) { w.db.FaerieFire = true }},
	{Gift: GiftCoE, Name: "Curse of the Elements", Icon: "spell_shadow_chilltouch", Debuff: true,
		write: func(w *buffWriter) { w.db.CurseOfElements = true }},
	{Gift: GiftCrusader, Name: "Judgement of the Crusader", Icon: "spell_holy_holysmite", Debuff: true,
		write: func(w *buffWriter) { w.db.JudgementOfTheCrusader = regular }},
	{Gift: GiftJoW, Name: "Judgement of Wisdom", Icon: "spell_holy_righteousnessaura", Debuff: true,
		write: func(w *buffWriter) { w.db.JudgementOfWisdom = true }},
}

// effectOf is each gift's own row (not an alias).
var effectOf = func() map[Gift]*Effect {
	m := map[Gift]*Effect{}
	for _, e := range Effects {
		if !e.Alias {
			m[e.Gift] = e
		}
	}
	for i := 0; i < numGifts; i++ {
		if m[Gift(1)<<i] == nil {
			panic("compsearch: no Effects row for gift bit " + string(rune('0'+i)))
		}
	}
	return m
}()

// SpecRow is what a DPS spec gives the raid and how it takes what the raid gives.
//
// Gives is what every build of the spec gives. What only some builds give (a talent's aura, a hunter's
// pet) is read from each build (PlayerGifts), and May lists it here for the search's buff sets.
type SpecRow struct {
	Spec  proto.Spec
	Class proto.Class
	Gives Gift
	May   Gift
	// Casts is what its own rotation casts, which its own request then leaves out.
	Casts Gift
	// Package is whether it drops a Mana Spring for its party.
	Package bool
	// PullSunder is whether it casts one Sunder Armor at the pull.
	PullSunder bool
	Melee      bool
	Caster     bool // Power Infusion's magic damage is for it
	Mana       bool
	// FlametongueTotem is whether it takes Flametongue Totem rather than Windfury (Windfury Weapon makes
	// Windfury Totem useless to it).
	FlametongueTotem bool
}

var Specs = []*SpecRow{
	{Spec: proto.Spec_SpecWarrior, Class: proto.Class_ClassWarrior, PullSunder: true, Melee: true},
	{Spec: proto.Spec_SpecRetributionPaladin, Class: proto.Class_ClassPaladin, Gives: GiftBlessing, Melee: true, Mana: true},
	{Spec: proto.Spec_SpecHunter, Class: proto.Class_ClassHunter, May: GiftTrueshot | GiftFuriousHowl | GiftMinorArmor, Mana: true},
	{Spec: proto.Spec_SpecRogue, Class: proto.Class_ClassRogue, Melee: true},
	{Spec: proto.Spec_SpecShadowPriest, Class: proto.Class_ClassPriest, Gives: GiftPoS | GiftFort, Caster: true, Mana: true},
	{Spec: proto.Spec_SpecEnhancementShaman, Class: proto.Class_ClassShaman, Gives: GiftSoE | GiftWF, Casts: GiftSoE | GiftWF,
		Package: true, Melee: true, Mana: true, FlametongueTotem: true},
	{Spec: proto.Spec_SpecElementalShaman, Class: proto.Class_ClassShaman, Package: true, Caster: true, Mana: true},
	{Spec: proto.Spec_SpecMage, Class: proto.Class_ClassMage, Gives: GiftAB, Caster: true, Mana: true},
	{Spec: proto.Spec_SpecWarlock, Class: proto.Class_ClassWarlock, Caster: true, Mana: true},
	{Spec: proto.Spec_SpecBalanceDruid, Class: proto.Class_ClassDruid, Gives: GiftGotW, May: GiftMoonkin, Caster: true, Mana: true},
	{Spec: proto.Spec_SpecFeralDruid, Class: proto.Class_ClassDruid, Gives: GiftGotW | GiftLotP, Melee: true},
}

var specRows = func() map[proto.Spec]*SpecRow {
	m := map[proto.Spec]*SpecRow{}
	for _, s := range Specs {
		m[s.Spec] = s
	}
	return m
}()

// noSpec is the row of a spec the table doesn't have: it gives nothing and takes Windfury.
var noSpec = &SpecRow{Mana: true}

// specOfPlayer is a player's row.
func specOfPlayer(p *proto.Player) *SpecRow {
	if s, ok := specRows[core.PlayerProtoToSpec(p)]; ok && s.Class == p.Class {
		return s
	}
	return noSpec
}

// ClassRow is a class as the page shows it, in the order it lists them (the blessings matrix).
type ClassRow struct {
	Class proto.Class
	Name  string
	Icon  string
}

var Classes = []*ClassRow{
	{proto.Class_ClassWarrior, "Warrior", "class_warrior"},
	{proto.Class_ClassPaladin, "Paladin", "class_paladin"},
	{proto.Class_ClassHunter, "Hunter", "class_hunter"},
	{proto.Class_ClassRogue, "Rogue", "class_rogue"},
	{proto.Class_ClassPriest, "Priest", "class_priest"},
	{proto.Class_ClassShaman, "Shaman", "class_shaman"},
	{proto.Class_ClassMage, "Mage", "class_mage"},
	{proto.Class_ClassWarlock, "Warlock", "class_warlock"},
	{proto.Class_ClassDruid, "Druid", "class_druid"},
}

// BlessingRow is one Greater Blessing. A class gets one blessing from each paladin, so the raid needs as
// many paladins as there are rows to give every class all of them.
type BlessingRow struct {
	Blessing proto.Blessing
	Name     string
	Icon     string
	write    func(ib *proto.IndividualBuffs)
}

var Blessings = []*BlessingRow{
	{proto.Blessing_BlessingKings, "Kings", "spell_magic_greaterblessingofkings", func(ib *proto.IndividualBuffs) { ib.BlessingOfKings = true }},
	{proto.Blessing_BlessingMight, "Might", "spell_holy_greaterblessingofkings", func(ib *proto.IndividualBuffs) { ib.BlessingOfMight = regular }},
	{proto.Blessing_BlessingWisdom, "Wisdom", "spell_holy_greaterblessingofwisdom", func(ib *proto.IndividualBuffs) { ib.BlessingOfWisdom = regular }},
}

// maxPaladins is how many paladins give a class every blessing. More add nothing but their judgements.
var maxPaladins = len(Blessings)

var allBlessings = func() []proto.Blessing {
	var out []proto.Blessing
	for _, b := range Blessings {
		out = append(out, b.Blessing)
	}
	return out
}()

// JudgementRow is a judgement a paladin keeps on the boss. A paladin placeholder judges one of them.
type JudgementRow struct {
	Judgement proto.PaladinJudgement
	Name      string // after "Judgement of"
	Gift      Gift
	Icon      string
}

var Judgements = []*JudgementRow{
	{proto.PaladinJudgement_PaladinJudgementCrusader, "the Crusader", GiftCrusader, "spell_holy_holysmite"},
	{proto.PaladinJudgement_PaladinJudgementWisdom, "Wisdom", GiftJoW, "spell_holy_righteousnessaura"},
}

// TotemRow is a totem a shaman placeholder can drop: one of each element.
type TotemRow struct {
	Totem   proto.RaidTotem
	Element proto.TotemElement
	Name    string
	Gift    Gift // none for Mana Spring, which is its party's package
	Icon    string
}

var Totems = []*TotemRow{
	{proto.RaidTotem_RaidTotemStrengthOfEarth, proto.TotemElement_TotemElementEarth, "Strength of Earth", GiftSoE, "spell_nature_earthbindtotem"},
	{proto.RaidTotem_RaidTotemGraceOfAir, proto.TotemElement_TotemElementAir, "Grace of Air", GiftGoA, "spell_nature_invisibilitytotem"},
	{proto.RaidTotem_RaidTotemWindfury, proto.TotemElement_TotemElementAir, "Windfury", GiftWF, "spell_nature_windfury"},
	{proto.RaidTotem_RaidTotemManaSpring, proto.TotemElement_TotemElementWater, "Mana Spring", 0, "spell_nature_manaregentotem"},
	{proto.RaidTotem_RaidTotemFlametongue, proto.TotemElement_TotemElementFire, "Flametongue", GiftFT, "spell_nature_guardianward"},
}

// ElementRow is a totem element, in the order the page lists them.
type ElementRow struct {
	Element proto.TotemElement
	Name    string
}

var Elements = []*ElementRow{
	{proto.TotemElement_TotemElementEarth, "Earth"},
	{proto.TotemElement_TotemElementAir, "Air"},
	{proto.TotemElement_TotemElementWater, "Water"},
	{proto.TotemElement_TotemElementFire, "Fire"},
}

func totemRow(t proto.RaidTotem) *TotemRow {
	for _, r := range Totems {
		if r.Totem == t {
			return r
		}
	}
	return nil
}

// elementTotems is an element's totems.
func elementTotems(e proto.TotemElement) []proto.RaidTotem {
	var out []proto.RaidTotem
	for _, r := range Totems {
		if r.Element == e {
			out = append(out, r.Totem)
		}
	}
	return out
}

// airTotems is the choice a shaman placeholder makes: its air totem. Its other elements have one totem each.
var airTotems = elementTotems(proto.TotemElement_TotemElementAir)

// fixedTotems is a shaman placeholder's totems of the elements with no choice.
var fixedTotems = func() []proto.RaidTotem {
	var out []proto.RaidTotem
	for _, e := range Elements {
		if ts := elementTotems(e.Element); len(ts) == 1 {
			out = append(out, ts[0])
		}
	}
	return out
}()

// fixedTotemGifts is what a shaman placeholder's fixed totems give the raid.
var fixedTotemGifts = func() Gift {
	var g Gift
	for _, t := range fixedTotems {
		g |= totemRow(t).Gift
	}
	return g
}()

// PackageRow is what a party gets from its shamans' water totems.
type PackageRow struct {
	Package proto.RaidPackage
	Name    string
	write   func(rb *proto.RaidBuffs, pb *proto.PartyBuffs)
}

var Packages = []*PackageRow{
	{proto.RaidPackage_RaidPackageManaSpring, "Mana Spring", func(rb *proto.RaidBuffs, pb *proto.PartyBuffs) {
		rb.ManaSpringTotem = regular
	}},
	{proto.RaidPackage_RaidPackageManaSpringTide, "Mana Spring and Mana Tide", func(rb *proto.RaidBuffs, pb *proto.PartyBuffs) {
		rb.ManaSpringTotem = regular
		pb.ManaTideTotems = 1
	}},
	{proto.RaidPackage_RaidPackageRestoration, "Mana Spring ×1.25 and Mana Tide", func(rb *proto.RaidBuffs, pb *proto.PartyBuffs) {
		rb.ManaSpringTotem = proto.TristateEffect_TristateEffectImproved
		pb.ManaTideTotems = 1
	}},
}

func packageRow(p proto.RaidPackage) *PackageRow {
	for _, r := range Packages {
		if r.Package == p {
			return r
		}
	}
	return nil
}

// DutyRow is an assignment a DPS member can hold: its duty variant casts it and keeps it up.
type DutyRow struct {
	Duty proto.Duty
	// Name is the report's name for it (a duty variant's meta.json), Label the page's.
	Name  string
	Label string
	Class proto.Class
	Gift  Gift
	Icon  string
}

var Duties = []*DutyRow{
	{proto.Duty_DutyBattleShout, "Battle Shout", "", proto.Class_ClassWarrior, GiftBattleShout, "ability_warrior_battleshout"},
	{proto.Duty_DutySunderUpkeep, "Sunder Armor", "Sunder Armor upkeep", proto.Class_ClassWarrior, GiftSunderTank, "ability_warrior_sunder"},
	{proto.Duty_DutyFaerieFire, "Faerie Fire", "", proto.Class_ClassDruid, GiftMinorArmor, "spell_nature_faeriefire"},
	{proto.Duty_DutyCurseOfTheElements, "Curse of the Elements", "", proto.Class_ClassWarlock, GiftCoE, "spell_shadow_chilltouch"},
	{proto.Duty_DutyCurseOfRecklessness, "Curse of Recklessness", "", proto.Class_ClassWarlock, GiftMinorArmor, "spell_shadow_unholystrength"},
	{proto.Duty_DutyJudgementOfTheCrusader, "Judgement of the Crusader", "", proto.Class_ClassPaladin, GiftCrusader, "spell_holy_holysmite"},
	{proto.Duty_DutyJudgementOfWisdom, "Judgement of Wisdom", "", proto.Class_ClassPaladin, GiftJoW, "spell_holy_righteousnessaura"},
}

func dutyRow(d proto.Duty) *DutyRow {
	for _, r := range Duties {
		if r.Duty == d {
			return r
		}
	}
	return nil
}

// SpecIcons is each DPS spec's icon on the page, by the report's class and spec names. A spec without one
// shows its class's icon.
var SpecIcons = map[string]string{
	"Druid:Balance":        "spell_nature_starfall",
	"Druid:Feral":          "ability_druid_catform",
	"Hunter:Beast Mastery": "ability_hunter_beasttaming",
	"Hunter:Marksmanship":  "ability_marksmanship",
	"Hunter:Survival":      "ability_hunter_swiftstrike",
	"Mage:Arcane":          "spell_holy_magicalsentry",
	"Mage:Fire":            "spell_fire_firebolt02",
	"Mage:Frost":           "spell_frost_frostbolt02",
	"Mage:Frostfire":       "ability_mage_frostfirebolt",
	"Paladin:Retribution":  "spell_holy_auraoflight",
	"Priest:Shadow":        "spell_shadow_shadowwordpain",
	"Priest:Smite":         "spell_holy_holysmite",
	"Rogue:Assassination":  "ability_rogue_eviscerate",
	"Rogue:Combat":         "ability_backstab",
	"Rogue:Subtlety":       "ability_stealth",
	"Shaman:Elemental":     "spell_nature_lightning",
	"Shaman:Enhancement":   "ability_shaman_stormstrike",
	"Warlock:Affliction":   "spell_shadow_deathcoil",
	"Warlock:Demonology":   "spell_shadow_metamorphosis",
	"Warlock:Destruction":  "spell_shadow_rainoffire",
	"Warrior:Arms":         "ability_warrior_savageblow",
	"Warrior:Fury":         "ability_warrior_innerrage",
}

// The cooldowns a placeholder gives one member, and the party package's Mana Tide, for the page's icons.
const (
	PowerInfusionIcon = "spell_holy_powerinfusion"
	InnervateIcon     = "spell_nature_lightning"
	ManaTideIcon      = "spell_frost_summonwaterelemental"
)

// toggles are the roles DPS members can bring to a buff set, each with what comes with it: a spec's
// effects that every build of it gives come together (a priest's Fortitude and Prayer of Spirit), and so
// do a spec's always effects with one it may give (a Balance druid's Moonkin Aura brings Gift of the
// Wild). Blessings are counted by paladins instead. They're in the table's order.
var toggles = func() []Gift {
	var sources []Gift
	for _, s := range Specs {
		if g := s.Gives &^ GiftBlessing; g != 0 {
			sources = append(sources, g)
		}
		for i := 0; i < numGifts; i++ {
			if bit := Gift(1) << i; s.May.Has(bit) {
				sources = append(sources, s.Gives&^GiftBlessing|bit)
			}
		}
	}
	for _, d := range Duties {
		for _, s := range Specs {
			if s.Class == d.Class {
				sources = append(sources, s.Gives&^GiftBlessing|d.Gift)
			}
		}
	}
	var out []Gift
	seen := Gift(0)
	for _, e := range Effects {
		if e.Alias || e.Gift == GiftBlessing || seen.Has(e.Gift) {
			continue
		}
		with := ^Gift(0)
		found := false
		for _, src := range sources {
			if src.Has(e.Gift) {
				with &= src
				found = true
			}
		}
		if !found {
			continue
		}
		out = append(out, with)
		seen |= with
	}
	return out
}()

// blesses is whether a build gives Greater Blessings (a paladin).
func blesses(b *Build) bool { return b.Gives.Has(GiftBlessing) }

// dropsTotems is whether a build drops a Mana Spring for its party, which makes it a package giver (a
// shaman).
func dropsTotems(b *Build) bool { return b.Pkg != proto.RaidPackage_RaidPackageNone }

// judgementGift is what a judgement keeps on the boss.
func judgementGift(j proto.PaladinJudgement) Gift {
	for _, r := range Judgements {
		if r.Judgement == j {
			return r.Gift
		}
	}
	return 0
}

// totemGift is what a totem gives the raid (none for Mana Spring, its party's package).
func totemGift(t proto.RaidTotem) Gift {
	if r := totemRow(t); r != nil {
		return r.Gift
	}
	return 0
}

func totemElement(t proto.RaidTotem) proto.TotemElement {
	if r := totemRow(t); r != nil {
		return r.Element
	}
	return proto.TotemElement_TotemElementUnknown
}

// otherAir is the air totem a second shaman placeholder drops.
func otherAir(t proto.RaidTotem) proto.RaidTotem {
	for _, a := range airTotems {
		if a != t {
			return a
		}
	}
	return t
}

// totemName is a totem's effect name, like "Windfury Totem".
func totemName(t proto.RaidTotem) string {
	if g := totemGift(t); g != 0 {
		return GiftName(g)
	}
	return t.String()
}

func packageName(p proto.RaidPackage) string {
	if r := packageRow(p); r != nil {
		return r.Name
	}
	return "none"
}
