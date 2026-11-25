package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	goproto "google.golang.org/protobuf/proto"
)

// The EP pre-trim of the gear search's pool (-gear-ep-keep). A tank set's score is six boss
// sims and a healer's fourteen, and a slot's pool holds 500 to 2500 options (random suffix
// rolls), so even the screen takes hours. Before the search we measure linear stat weights of
// the search's own objective by finite differences on a reference set, and each slot keeps
// only its best options by those weights, plus what a linear weight can't judge (keepByEP).

var (
	gearEpKeep     = flag.Int("gear-ep-keep", 0, "gear search: before the search, keep only each slot's best N options by stat weights measured on the search's own objective (tank, healer or DPS), plus every item with a use or proc, every set piece, the starting items, -gear-include and the best N of each weapon class by EP and by weapon DPS; 0 keeps the whole pool")
	gearEpSettings = flag.String("gear-ep-settings", "", "with -gear-ep-keep: a RaidSimRequest or settings file whose first player's gear is the set the weights are measured on (the -settings start set by default)")
	gearEpIters    = flag.Int("gear-ep-iters", 1000, "with -gear-ep-keep: the iterations of each weight's runs (paired seeds with the reference run)")
)

// epDeltas is what each weight's run adds of a stat, about an item's worth, so the gain stands
// out of the noise and stays in the linear range. Stats not listed get epDefaultDelta.
var epDeltas = map[proto.Stat]float64{
	proto.Stat_StatStrength: 25, proto.Stat_StatAgility: 25, proto.Stat_StatStamina: 25,
	proto.Stat_StatIntellect: 25, proto.Stat_StatSpirit: 25,
	proto.Stat_StatSpellPower: 40, proto.Stat_StatArcanePower: 40, proto.Stat_StatFirePower: 40,
	proto.Stat_StatFrostPower: 40, proto.Stat_StatHolyPower: 40, proto.Stat_StatNaturePower: 40,
	proto.Stat_StatShadowPower: 40, proto.Stat_StatSpellDamage: 40, proto.Stat_StatHealingPower: 60,
	proto.Stat_StatMP5: 10, proto.Stat_StatMana: 300, proto.Stat_StatHealth: 300,
	proto.Stat_StatSpellHit: 2, proto.Stat_StatSpellCrit: 2, proto.Stat_StatSpellHaste: 2,
	proto.Stat_StatMeleeHit: 2, proto.Stat_StatMeleeCrit: 2, proto.Stat_StatMeleeHaste: 2,
	proto.Stat_StatSpellPenetration: 20, proto.Stat_StatArmorPenetration: 100,
	// Expertise is in percent of dodge and parry the target loses (an item's weapon skill), and
	// a boss dodges about 5.6%.
	proto.Stat_StatExpertise:   1,
	proto.Stat_StatAttackPower: 50, proto.Stat_StatRangedAttackPower: 50, proto.Stat_StatFeralAttackPower: 50,
	proto.Stat_StatArmor: 500, proto.Stat_StatBonusArmor: 500,
	proto.Stat_StatDefense: 15, proto.Stat_StatDodge: 2, proto.Stat_StatParry: 2, proto.Stat_StatBlock: 2,
	proto.Stat_StatBlockValue:       40,
	proto.Stat_StatArcaneResistance: 30, proto.Stat_StatFireResistance: 30, proto.Stat_StatFrostResistance: 30,
	proto.Stat_StatNatureResistance: 30, proto.Stat_StatShadowResistance: 30,
}

const epDefaultDelta = 10

// epHealScale multiplies the deltas of the healer objective.
const epHealScale = 4

// epName is a weight's name. Forever pays an item's hit and crit to melee and spells alike
// (core.unifyEquipHitAndCrit), so the melee index stands for both.
func epName(stat int) string {
	switch proto.Stat(stat) {
	case proto.Stat_StatMeleeHit:
		return "Hit"
	case proto.Stat_StatMeleeCrit:
		return "Crit"
	}
	return proto.Stat(stat).String()[len("Stat"):]
}

// epStats is what an option gives in the weights' terms: the item's stats and its suffix's,
// with hit and crit (spell, melee and Forever's ratings) folded into the melee index.
func epStats(o gearOption) []float64 {
	v := make([]float64, len(proto.Stat_name))
	item := itemsByID[o.spec.Id]
	for _, src := range [][]float64{item.GetStats(), o.suffixStats} {
		for i, x := range src {
			if i < len(v) {
				v[i] += x
			}
		}
	}
	hit, crit, spellHit, spellCrit := proto.Stat_StatMeleeHit, proto.Stat_StatMeleeCrit, proto.Stat_StatSpellHit, proto.Stat_StatSpellCrit
	v[hit] += v[spellHit] + item.GetHitRating()/core.HitRatingPerPercent
	v[crit] += v[spellCrit] + item.GetCritRating()/core.CritRatingPerPercent
	v[spellHit], v[spellCrit] = 0, 0
	return v
}

// epValue is an option's worth by the weights.
func epValue(o gearOption, weights []float64) float64 {
	var sum float64
	for i, x := range epStats(o) {
		if i < len(weights) {
			sum += x * weights[i]
		}
	}
	return sum
}

// weaponDPS is a weapon's damage per second, 0 for a shield or an item held in the off hand.
func weaponDPS(o gearOption) float64 {
	item := itemsByID[o.spec.Id]
	if item.GetWeaponSpeed() <= 0 {
		return 0
	}
	return (item.GetWeaponDamageMin() + item.GetWeaponDamageMax()) / 2 / item.GetWeaponSpeed()
}

// weaponClass is the setup class of a weapon option that applyWeaponRules tells apart, "" for
// a slot that isn't a hand.
func weaponClass(slot int, o gearOption) string {
	switch proto.ItemSlot(slot) {
	case proto.ItemSlot_ItemSlotMainHand:
		if o.twoHand {
			return "two-hand"
		}
		return "one-hand"
	case proto.ItemSlot_ItemSlotOffHand:
		switch itemsByID[o.spec.Id].GetWeaponType() {
		case proto.WeaponType_WeaponTypeShield:
			return "shield"
		case proto.WeaponType_WeaponTypeOffHand:
			return "held"
		}
		return "one-hand"
	case proto.ItemSlot_ItemSlotRanged:
		return "ranged"
	}
	return ""
}

// keepByEP trims each slot to its best keep options by the weights, and always keeps:
//   - every option with a use or proc effect (effect), modelled or not, since a weight can't
//     price it and the hand check has to see the unmodelled ones,
//   - every piece of an item set, whose bonus the weights leave out,
//   - an item with weapon skill, which the weights leave out too,
//   - the starting items (start, any slot's, so a ring or a one hander counts in both slots),
//   - the -gear-include items,
//   - in a weapon slot, the best keep of each weapon class by EP and by weapon DPS, so no
//     weapon setup loses all its options.
func keepByEP(pool [][]gearOption, weights []float64, keep int, start []*proto.ItemSpec, include []int32, effect func(int32) bool) [][]gearOption {
	out := make([][]gearOption, len(pool))
	for slot, options := range pool {
		kept := make([]bool, len(options))
		for i, o := range options {
			item := itemsByID[o.spec.Id]
			kept[i] = effect(o.spec.Id) || item.GetSetName() != "" || item.GetSetId() != 0 ||
				slices.ContainsFunc(item.GetWeaponSkills(), func(v float64) bool { return v > 0 }) ||
				slices.Contains(include, o.spec.Id) ||
				slices.ContainsFunc(start, func(s *proto.ItemSpec) bool {
					return s != nil && s.Id == o.spec.Id && s.RandomSuffix == o.spec.RandomSuffix
				})
		}
		top := func(of []int, by func(gearOption) float64) {
			sort.SliceStable(of, func(a, b int) bool { return by(options[of[a]]) > by(options[of[b]]) })
			for _, i := range of[:min(keep, len(of))] {
				kept[i] = true
			}
		}
		ep := func(o gearOption) float64 { return epValue(o, weights) }
		all := make([]int, len(options))
		classes := map[string][]int{}
		for i, o := range options {
			all[i] = i
			if c := weaponClass(slot, o); c != "" {
				classes[c] = append(classes[c], i)
			}
		}
		top(all, ep)
		for _, of := range classes {
			top(slices.Clone(of), ep)
			armed := slices.DeleteFunc(slices.Clone(of), func(i int) bool { return weaponDPS(options[i]) <= 0 })
			top(armed, weaponDPS)
		}
		for i, o := range options {
			if kept[i] {
				out[slot] = append(out[slot], o)
			}
		}
	}
	return out
}

// equipProc catches an equip line the tooltip check misses: "Equip: Chance on ..." or "When
// struck ...".
var equipProc = regexp.MustCompile(`Equip: [^<]*([Cc]hance|When struck|Sometimes)`)

// poolEffects is keepByEP's effect check: an item the sim gives an effect, or one whose
// Classic tooltip has a use or proc line (unmodelledEffect, equipProc).
func poolEffects() (func(int32) bool, error) {
	tooltips, err := readTooltips(*itemTooltips)
	if err != nil {
		return nil, err
	}
	return func(id int32) bool {
		tip := tooltips[id]
		return core.HasItemEffect(id) || core.HasWeaponEffect(id) || core.ItemHasUse(id) ||
			unmodelledEffect(id, tip) != "" || equipProc.MatchString(tip)
	}, nil
}

// poolStatsInUse is every stat an option of the pool has, in epStats' terms.
func poolStatsInUse(pool [][]gearOption) []int {
	var out []int
	seen := map[int]bool{}
	for _, options := range pool {
		for _, o := range options {
			for i, x := range epStats(o) {
				if x != 0 && !seen[i] {
					seen[i] = true
					out = append(out, i)
				}
			}
		}
	}
	slices.Sort(out)
	return out
}

// epWeights measures the objective's weight of each stat the pool has: on the reference set,
// the score with a delta of the stat in the player's bonus stats less the score without, over
// the delta. It scores through prepare and scoreAll like the search, so a tank's or healer's
// weights are in its own score over every encounter, and the runs share their seeds.
//
// The healer objective measures each slot's reference carry on the first set it scores
// (heal.go), which must stay the starting set. So that is measured first, on the start.
func (s *searcher) epWeights(pool [][]gearOption, knobs Knobs) []float64 {
	player := s.setup.player()
	startGear := goproto.Clone(player.Equipment).(*proto.EquipmentSpec)
	startBonus := player.BonusStats
	iterations := s.iterations
	defer func() {
		player.Equipment, player.BonusStats = startGear, startBonus
		s.iterations = iterations
		s.cache = map[string]result{}
	}()
	if healSlots != nil {
		// The start as searchGear scores it: an enchant it can't hold comes off first.
		gear := equipmentSlots(startGear)
		for slot, item := range gear {
			if item != nil && item.Enchant != 0 {
				item.Enchant = s.carriedEnchant(slot, item.Id, item.Enchant)
			}
		}
		s.setGear(gear)
		s.setup.runRequest(s.template.Build(knobs), 0, s.iterations)
	}
	ref := "the start set"
	refGear := startGear
	if *gearEpSettings != "" {
		r, err := loadSetup(*gearEpSettings)
		if err != nil {
			fmt.Fprintf(os.Stderr, "-gear-ep-settings: %v\n", err)
			os.Exit(2)
		}
		ref, refGear = *gearEpSettings, r.player().Equipment
	}
	player.Equipment = goproto.Clone(refGear).(*proto.EquipmentSpec)
	s.iterations = int32(*gearEpIters)

	with := func(stat int, delta float64) *pendingScore {
		bonus := &proto.UnitStats{}
		if startBonus != nil {
			bonus = goproto.Clone(startBonus).(*proto.UnitStats)
		}
		for len(bonus.Stats) < len(proto.Stat_name) {
			bonus.Stats = append(bonus.Stats, 0)
		}
		if stat >= 0 {
			bonus.Stats[stat] += delta
			switch proto.Stat(stat) {
			case proto.Stat_StatMeleeHit:
				bonus.Stats[proto.Stat_StatSpellHit] += delta
			case proto.Stat_StatMeleeCrit:
				bonus.Stats[proto.Stat_StatSpellCrit] += delta
			}
		}
		player.BonusStats = bonus
		// The cache key leaves out gear and bonus stats, so every run gets its own.
		s.cache = map[string]result{}
		return s.prepare(knobs)
	}
	statList := poolStatsInUse(pool)
	deltas := make([]float64, len(statList))
	pending := []*pendingScore{with(-1, 0)}
	for i, stat := range statList {
		deltas[i] = epDefaultDelta
		if d, ok := epDeltas[proto.Stat(stat)]; ok {
			deltas[i] = d
		}
		// A healer's score moves by whole failed fights (heal.go), so an item's worth of a stat
		// is lost in its noise. Four items' worth stands out.
		if healSlots != nil {
			deltas[i] *= epHealScale
		}
		pending = append(pending, with(stat, deltas[i]))
	}
	results := s.scoreAll(pending)
	base := results[0]
	weights := make([]float64, len(proto.Stat_name))
	fmt.Printf("EP weights on %s at %d iterations (score %s), per point:\n", ref, *gearEpIters, base)
	for i, stat := range statList {
		r := results[i+1]
		weights[stat] = (s.objective(r) - s.objective(base)) / deltas[i]
		fmt.Printf("  %-18s %+8.4f (±%.4f, +%g)\n", epName(stat), weights[stat], gainNoise(base, r)/deltas[i], deltas[i])
	}
	return weights
}
