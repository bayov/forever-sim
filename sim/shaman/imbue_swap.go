package shaman

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	googleProto "google.golang.org/protobuf/proto"
)

// Imbue swaps for the practice room (docs/practice_room.md, "Your toolkit"), where the player can
// cast any weapon imbue at any time. A normal sim never calls EnableImbueSwaps, and builds only
// the imbue the settings pick.
//
// Each imbue does its work through its aura (the Windfury, Flametongue and Frostbrand procs), so
// a swap turns exactly one of them on. Rockbiter's attack power is added to the weapon's stats
// when the character is built, so we move it with a dynamic stat instead. The main hand's
// TempEnchant is what the totem weapon buffs read on every hit, and our own standing weapon
// totem picks its buff again, as it does when it goes down.

// ImbueCastTag tags the imbue casts, apart from the imbues' proc spells, which share their ids.
const ImbueCastTag = 99

// The mana cost of each imbue's top rank (wowhead's Classic tooltips).
var imbueCastCost = map[proto.WeaponImbue]float64{
	proto.WeaponImbue_RockbiterWeapon:   150,
	proto.WeaponImbue_FlametongueWeapon: 155,
	proto.WeaponImbue_FrostbrandWeapon:  160,
	proto.WeaponImbue_WindfuryWeapon:    165,
}

type imbueSwaps struct {
	auras    map[proto.WeaponImbue]*core.Aura
	enchants map[proto.WeaponImbue]int32
	spells   map[proto.WeaponImbue]int32

	// The attack power Rockbiter gives, whether the build put it in the weapon's stats, and
	// how much we have added or taken away since.
	rockbiterAP    float64
	rockbiterBaked bool
	rockbiterDelta float64

	active  proto.WeaponImbue
	initial proto.WeaponImbue

	// Our own weapon totems standing, with the imbue that turns their buff off for us.
	weaponTotems map[*core.Aura]proto.WeaponImbue
}

// EnableImbueSwaps registers every imbue and a cast for each, so any of them can go on the main
// hand. Call it while the sim is built (core.EnvHooks), before it's finalized.
func (shaman *Shaman) EnableImbueSwaps() {
	if !shaman.HasMHWeapon() || shaman.imbueSwaps != nil {
		return
	}
	swaps := &imbueSwaps{
		auras:        map[proto.WeaponImbue]*core.Aura{},
		enchants:     map[proto.WeaponImbue]int32{},
		spells:       map[proto.WeaponImbue]int32{},
		weaponTotems: map[*core.Aura]proto.WeaponImbue{},
	}
	shaman.imbueSwaps = swaps

	initialEnchant := shaman.MainHand().TempEnchant
	switch {
	case shaman.ShamanImbue != proto.WeaponImbue_WeaponImbueUnknown:
		swaps.initial = shaman.ShamanImbue
	case isShamanImbue(shaman.Consumes.MainHandImbue):
		swaps.initial = shaman.Consumes.MainHandImbue
	}

	// The imbues the build didn't register yet, on the main hand. Each sets the main hand's
	// enchant, which we put back.
	register := map[proto.WeaponImbue]struct {
		label string
		fn    func(core.ProcMask)
	}{
		proto.WeaponImbue_WindfuryWeapon:    {"Windfury Imbue", shaman.RegisterWindfuryImbue},
		proto.WeaponImbue_FlametongueWeapon: {"Flametongue Imbue", shaman.RegisterFlametongueImbue},
		proto.WeaponImbue_FrostbrandWeapon:  {"Frostbrand Imbue", shaman.RegisterFrostbrandImbue},
	}
	for imbue, r := range register {
		if shaman.GetAura(r.label) == nil {
			r.fn(core.ProcMaskMeleeMH)
		}
		if aura := shaman.GetAura(r.label); aura != nil {
			swaps.auras[imbue] = aura
		}
	}
	shaman.MainHand().TempEnchant = initialEnchant

	if rank := shaman.windfuryRank(); rank > 0 {
		swaps.enchants[proto.WeaponImbue_WindfuryWeapon] = WindfuryWeaponEnchantId[rank]
		swaps.spells[proto.WeaponImbue_WindfuryWeapon] = WindfuryWeaponSpellId[rank]
	}
	if rank := shaman.flametongueRank(); rank > 0 {
		swaps.enchants[proto.WeaponImbue_FlametongueWeapon] = FlametongueWeaponEnchantId[rank]
		swaps.spells[proto.WeaponImbue_FlametongueWeapon] = FlametongueWeaponSpellId[rank]
	}
	if rank := shaman.frostbrandRank(); rank > 0 {
		swaps.enchants[proto.WeaponImbue_FrostbrandWeapon] = FrostbrandWeaponEnchantId[rank]
		swaps.spells[proto.WeaponImbue_FrostbrandWeapon] = FrostbrandWeaponSpellId[rank]
	}
	if rank := shaman.rockbiterRank(); rank > 0 {
		swaps.enchants[proto.WeaponImbue_RockbiterWeapon] = RockbiterWeaponEnchantId[rank]
		swaps.spells[proto.WeaponImbue_RockbiterWeapon] = RockbiterWeaponSpellId[rank]
		swaps.rockbiterAP = RockbiterWeaponScaling[rank].At(RockbiterWeaponBonusAP[rank], shaman.Level) * []float64{1, 1.07, 1.13, 1.2}[shaman.Talents.ElementalWeapons]
		swaps.rockbiterBaked = swaps.initial == proto.WeaponImbue_RockbiterWeapon
	}

	// A cast for each imbue: instant, with the GCD and its mana cost.
	for imbue, spellID := range swaps.spells {
		imbue := imbue
		shaman.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: spellID, Tag: ImbueCastTag},
			SpellSchool: core.SpellSchoolNature,
			Flags:       core.SpellFlagAPL | core.SpellFlagHelpful | SpellFlagShaman,
			ManaCost:    core.ManaCostOptions{FlatCost: imbueCastCost[imbue]},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{GCD: core.GCDDefault},
			},
			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				shaman.SwapMainHandImbue(sim, imbue)
			},
		})
	}
	swaps.active = swaps.initial
}

func isShamanImbue(imbue proto.WeaponImbue) bool {
	switch imbue {
	case proto.WeaponImbue_RockbiterWeapon, proto.WeaponImbue_FlametongueWeapon, proto.WeaponImbue_FrostbrandWeapon, proto.WeaponImbue_WindfuryWeapon:
		return true
	}
	return false
}

// ImbueCastSpellID is the spell id of the imbue's cast, 0 when the shaman doesn't know it.
func (shaman *Shaman) ImbueCastSpellID(imbue proto.WeaponImbue) int32 {
	if shaman.imbueSwaps == nil {
		return 0
	}
	return shaman.imbueSwaps.spells[imbue]
}

// ActiveImbue is the imbue on the main hand now.
func (shaman *Shaman) ActiveImbue() proto.WeaponImbue {
	if shaman.imbueSwaps == nil {
		return shaman.ShamanImbue
	}
	return shaman.imbueSwaps.active
}

// ResetImbueSwaps puts the settings' imbue back after the sim's Reset, which turns every imbue
// aura on.
func (shaman *Shaman) ResetImbueSwaps(sim *core.Simulation) {
	swaps := shaman.imbueSwaps
	if swaps == nil {
		return
	}
	swaps.rockbiterDelta = 0
	clear(swaps.weaponTotems)
	swaps.active = proto.WeaponImbue_WeaponImbueUnknown
	shaman.SwapMainHandImbue(sim, swaps.initial)
}

// SwapMainHandImbue puts the imbue on the main hand in place of the one there.
func (shaman *Shaman) SwapMainHandImbue(sim *core.Simulation, imbue proto.WeaponImbue) {
	swaps := shaman.imbueSwaps
	if swaps == nil {
		return
	}
	for kind, aura := range swaps.auras {
		if kind == imbue {
			aura.Activate(sim)
		} else {
			aura.Deactivate(sim)
		}
	}
	shaman.MainHand().TempEnchant = swaps.enchants[imbue]
	shaman.ShamanImbue = imbue
	if isShamanImbue(shaman.Consumes.MainHandImbue) {
		consumes := googleProto.Clone(shaman.Consumes).(*proto.Consumes)
		consumes.MainHandImbue = proto.WeaponImbue_WeaponImbueUnknown
		shaman.Consumes = consumes
	}

	// Rockbiter's attack power: on with Rockbiter, off with any other.
	want := 0.0
	if imbue == proto.WeaponImbue_RockbiterWeapon && !swaps.rockbiterBaked {
		want = swaps.rockbiterAP
	} else if imbue != proto.WeaponImbue_RockbiterWeapon && swaps.rockbiterBaked {
		want = -swaps.rockbiterAP
	}
	if delta := want - swaps.rockbiterDelta; delta != 0 {
		shaman.AddStatDynamic(sim, stats.AttackPower, delta)
		swaps.rockbiterDelta = want
	}
	swaps.active = imbue

	// Our own weapon totems pick their buff again: a same kind imbue turns the totem's buff off
	// for us, and any other lets it back on.
	for aura, sameKind := range swaps.weaponTotems {
		if shaman.getImbueProcMask(sameKind) != core.ProcMaskUnknown {
			shaman.ReleaseTotemWeaponBuff(sim, aura, &shaman.Unit)
		} else {
			shaman.HoldTotemWeaponBuff(sim, aura, &shaman.Unit, sameKind == proto.WeaponImbue_WindfuryWeapon)
		}
	}
}

// noteWeaponTotem keeps track of our standing weapon totems, for imbue swaps.
func (shaman *Shaman) noteWeaponTotem(aura *core.Aura, sameKindImbue proto.WeaponImbue, standing bool) {
	if shaman.imbueSwaps == nil {
		return
	}
	if standing {
		shaman.imbueSwaps.weaponTotems[aura] = sameKindImbue
	} else {
		delete(shaman.imbueSwaps.weaponTotems, aura)
	}
}
