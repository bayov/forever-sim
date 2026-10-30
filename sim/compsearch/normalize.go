package compsearch

import (
	_ "embed"
	"encoding/json"
	"sort"

	"github.com/wowsims/classic/sim/core/proto"
)

// The cache drops from a build's request the effects it can't read (docs/raid_sim.md, The cache), so
// Moonkin Aura on or off is one sim for every warrior. Which effects a build can't read is a static
// table, drops.json, written by `compsearch -drops` from bit-identical sims with and without each
// effect, under the full set and the empty set, camps on and off, and checked again by
// TestDropsAreUnread.

// Field is one setting a seat writes.
type Field struct {
	Name  string
	Clear func(req *proto.RaidSimRequest)
	Set   func(req *proto.RaidSimRequest) // the effect on, for the drop table's sims
}

func rb(req *proto.RaidSimRequest) *proto.RaidBuffs { return req.Raid.Buffs }
func db(req *proto.RaidSimRequest) *proto.Debuffs   { return req.Raid.Debuffs }
func ib(req *proto.RaidSimRequest) *proto.IndividualBuffs {
	return req.Raid.Parties[0].Players[0].Buffs
}
func camp(req *proto.RaidSimRequest) *proto.CampBuffs {
	if req.Raid.Buffs.CampBuffs == nil {
		req.Raid.Buffs.CampBuffs = &proto.CampBuffs{}
	}
	return req.Raid.Buffs.CampBuffs
}

const reg = proto.TristateEffect_TristateEffectRegular
const missing = proto.TristateEffect_TristateEffectMissing

// Fields is every effect the drop table may drop.
var Fields = []Field{
	{"moonkin_aura", func(r *proto.RaidSimRequest) { rb(r).MoonkinAura = false }, func(r *proto.RaidSimRequest) { rb(r).MoonkinAura = true }},
	{"leader_of_the_pack", func(r *proto.RaidSimRequest) { rb(r).LeaderOfThePack = false }, func(r *proto.RaidSimRequest) { rb(r).LeaderOfThePack = true }},
	{"trueshot_aura", func(r *proto.RaidSimRequest) { rb(r).TrueshotAura = false }, func(r *proto.RaidSimRequest) { rb(r).TrueshotAura = true }},
	{"arcane_brilliance", func(r *proto.RaidSimRequest) { rb(r).ArcaneBrilliance = false }, func(r *proto.RaidSimRequest) { rb(r).ArcaneBrilliance = true }},
	{"divine_spirit", func(r *proto.RaidSimRequest) { rb(r).DivineSpirit = false }, func(r *proto.RaidSimRequest) { rb(r).DivineSpirit = true }},
	{"furious_howl", func(r *proto.RaidSimRequest) { rb(r).FuriousHowl = false }, func(r *proto.RaidSimRequest) { rb(r).FuriousHowl = true }},
	{"battle_shout", func(r *proto.RaidSimRequest) { rb(r).BattleShout = missing }, func(r *proto.RaidSimRequest) { rb(r).BattleShout = reg }},
	{"strength_of_earth_totem", func(r *proto.RaidSimRequest) { rb(r).StrengthOfEarthTotem = missing }, func(r *proto.RaidSimRequest) { rb(r).StrengthOfEarthTotem = reg }},
	{"grace_of_air_totem", func(r *proto.RaidSimRequest) { rb(r).GraceOfAirTotem = missing }, func(r *proto.RaidSimRequest) { rb(r).GraceOfAirTotem = reg }},
	// The weapon totem is one field with two values: each is dropped on its own, so a build that can't
	// read Windfury Totem (Enhancement, with Windfury Weapon) keeps its Flametongue.
	{"totem_weapon_windfury", func(r *proto.RaidSimRequest) {
		if rb(r).TotemWeaponBuff == proto.TotemWeaponBuff_TotemWeaponBuffWindfury {
			rb(r).TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffNone
		}
	}, func(r *proto.RaidSimRequest) { rb(r).TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffWindfury }},
	{"totem_weapon_flametongue", func(r *proto.RaidSimRequest) {
		if rb(r).TotemWeaponBuff == proto.TotemWeaponBuff_TotemWeaponBuffFlametongue {
			rb(r).TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffNone
		}
	}, func(r *proto.RaidSimRequest) { rb(r).TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffFlametongue }},
	{"gift_of_the_wild", func(r *proto.RaidSimRequest) { rb(r).GiftOfTheWild = missing }, func(r *proto.RaidSimRequest) { rb(r).GiftOfTheWild = reg }},
	{"power_word_fortitude", func(r *proto.RaidSimRequest) { rb(r).PowerWordFortitude = missing }, func(r *proto.RaidSimRequest) { rb(r).PowerWordFortitude = reg }},
	{"mana_spring_totem", func(r *proto.RaidSimRequest) { rb(r).ManaSpringTotem = missing }, func(r *proto.RaidSimRequest) { rb(r).ManaSpringTotem = reg }},
	{"mana_tide_totems", func(r *proto.RaidSimRequest) { r.Raid.Parties[0].Buffs.ManaTideTotems = 0 }, func(r *proto.RaidSimRequest) { r.Raid.Parties[0].Buffs.ManaTideTotems = 1 }},
	{"blessing_of_kings", func(r *proto.RaidSimRequest) { ib(r).BlessingOfKings = false }, func(r *proto.RaidSimRequest) { ib(r).BlessingOfKings = true }},
	{"blessing_of_might", func(r *proto.RaidSimRequest) { ib(r).BlessingOfMight = missing }, func(r *proto.RaidSimRequest) { ib(r).BlessingOfMight = reg }},
	{"blessing_of_wisdom", func(r *proto.RaidSimRequest) { ib(r).BlessingOfWisdom = missing }, func(r *proto.RaidSimRequest) { ib(r).BlessingOfWisdom = reg }},
	{"innervates", func(r *proto.RaidSimRequest) { ib(r).Innervates = 0 }, func(r *proto.RaidSimRequest) { ib(r).Innervates = 1 }},
	{"power_infusions", func(r *proto.RaidSimRequest) { ib(r).PowerInfusions = 0 }, func(r *proto.RaidSimRequest) { ib(r).PowerInfusions = 1 }},
	{"curse_of_elements", func(r *proto.RaidSimRequest) { db(r).CurseOfElements = false }, func(r *proto.RaidSimRequest) { db(r).CurseOfElements = true }},
	{"judgement_of_the_crusader", func(r *proto.RaidSimRequest) { db(r).JudgementOfTheCrusader = missing }, func(r *proto.RaidSimRequest) { db(r).JudgementOfTheCrusader = reg }},
	{"judgement_of_wisdom", func(r *proto.RaidSimRequest) { db(r).JudgementOfWisdom = false }, func(r *proto.RaidSimRequest) { db(r).JudgementOfWisdom = true }},
	{"faerie_fire", func(r *proto.RaidSimRequest) { db(r).FaerieFire = false }, func(r *proto.RaidSimRequest) { db(r).FaerieFire = true }},
	{"camp.fish_bowl", func(r *proto.RaidSimRequest) { camp(r).FishBowl = false }, func(r *proto.RaidSimRequest) { camp(r).FishBowl = true }},
	{"camp.lodestone", func(r *proto.RaidSimRequest) { camp(r).Lodestone = false }, func(r *proto.RaidSimRequest) { camp(r).Lodestone = true }},
	{"camp.sharpening_wheel", func(r *proto.RaidSimRequest) { camp(r).SharpeningWheel = false }, func(r *proto.RaidSimRequest) { camp(r).SharpeningWheel = true }},
	{"camp.camp_chair", func(r *proto.RaidSimRequest) { camp(r).CampChair = false }, func(r *proto.RaidSimRequest) { camp(r).CampChair = true }},
	{"camp.incense_candle", func(r *proto.RaidSimRequest) { camp(r).IncenseCandle = false }, func(r *proto.RaidSimRequest) { camp(r).IncenseCandle = true }},
	{"camp.faction_banner", func(r *proto.RaidSimRequest) { camp(r).FactionBanner = false }, func(r *proto.RaidSimRequest) { camp(r).FactionBanner = true }},
	{"camp.mana_well", func(r *proto.RaidSimRequest) { camp(r).ManaWell = false }, func(r *proto.RaidSimRequest) { camp(r).ManaWell = true }},
	{"camp.enchanted_lute", func(r *proto.RaidSimRequest) { camp(r).EnchantedLute = false }, func(r *proto.RaidSimRequest) { camp(r).EnchantedLute = true }},
	{"camp.first_aid_kit", func(r *proto.RaidSimRequest) { camp(r).FirstAidKit = false }, func(r *proto.RaidSimRequest) { camp(r).FirstAidKit = true }},
}

var fieldByName = func() map[string]Field {
	m := map[string]Field{}
	for _, f := range Fields {
		m[f.Name] = f
	}
	return m
}()

//go:embed drops.json
var dropsJSON []byte

// Drops is the static table: build id -> the fields it can't read.
var Drops = func() map[string][]string {
	m := map[string][]string{}
	_ = json.Unmarshal(dropsJSON, &m)
	return m
}()

// normalize clears the effects the build can't read, and an empty camp list.
func normalize(req *proto.RaidSimRequest, b *Build) {
	for _, name := range Drops[DropKey(b.ID, req.Raid.Buffs.CampBuffs != nil)] {
		if f, ok := fieldByName[name]; ok {
			f.Clear(req)
		}
	}
	if c := req.Raid.Buffs.CampBuffs; c != nil && !c.FishBowl && !c.Lodestone && !c.SharpeningWheel && !c.CampChair &&
		!c.IncenseCandle && !c.FactionBanner && !c.ManaWell && !c.EnchantedLute && !c.FirstAidKit {
		req.Raid.Buffs.CampBuffs = nil
	}
}

// WriteDrops writes a drop table as drops.json's JSON.
func WriteDrops(drops map[string][]string) ([]byte, error) {
	for _, v := range drops {
		sort.Strings(v)
	}
	return json.MarshalIndent(drops, "", "\t")
}
