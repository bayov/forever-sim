package shaman

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// setActiveAirTotem puts the totem in the air slot and swaps in its buff.
//
// Windwall Totem has no buff in the sim, so it comes with a nil aura.
func (shaman *Shaman) setActiveAirTotem(sim *core.Simulation, spell *core.Spell, aura *core.Aura) {
	duration := time.Minute * 5 // Forever totems last 5 min, Classic 2
	if aura != nil {
		duration = aura.Duration
	}
	shaman.TotemExpirations[AirTotem] = sim.CurrentTime + duration
	shaman.ActiveTotems[AirTotem] = spell

	if shaman.ActiveTotemBuffs[AirTotem] != nil {
		shaman.ActiveTotemBuffs[AirTotem].Deactivate(sim)
	}

	shaman.ActiveTotemBuffs[AirTotem] = aura
	if aura != nil {
		aura.Activate(sim)
	}
}

const WindfuryTotemRanks = 3

var WindfuryTotemSpellId = [WindfuryTotemRanks + 1]int32{0, 8512, 10613, 10614}

// The rotations ask for this aura to see whether the totem stands.
//
// It's the party aura each rank gives in the Forever client (70338), which procs the attack
// power buff. We used Classic's 8514, 10607 and 10611 before, which the Forever client doesn't
// have (shaman_audit.md 7.1).
var WindfuryBuffAuraId = [WindfuryTotemRanks + 1]int32{0, 8515, 10609, 10612}

// Rotations saved before the rename still ask for 10611.
//
// We keep the old IDs in the same rank family as the new ones, so a custom preset that asks
// for 10611 still finds whichever Windfury Totem aura we have.
var oldWindfuryBuffAuraId = [WindfuryTotemRanks + 1]int32{0, 8514, 10607, 10611}
var WindfuryTotemManaCost = [WindfuryTotemRanks + 1]float64{0, 115, 175, 250}
var WindfuryTotemLevel = core.WindfuryTotemLevel

func (shaman *Shaman) registerWindfuryTotemSpell() {
	shaman.WindfuryTotem = make([]*core.Spell, WindfuryTotemRanks+1)

	for rank := 1; rank <= WindfuryTotemRanks; rank++ {
		// Only the ranks the level has learned, so nothing below (a totem buff aura) is
		// built for a rank the shaman cannot cast.
		if WindfuryTotemLevel[rank] <= int(shaman.Level) {
			config := shaman.newWindfuryTotemSpellConfig(rank)
			shaman.WindfuryTotem[rank] = shaman.RegisterSpell(config)
		}
	}

	shaman.AirTotems = append(
		shaman.AirTotems,
		core.FilterSlice(shaman.WindfuryTotem, func(spell *core.Spell) bool { return spell != nil })...,
	)
}

func (shaman *Shaman) newWindfuryTotemSpellConfig(rank int) core.SpellConfig {
	spellId := WindfuryTotemSpellId[rank]
	manaCost := WindfuryTotemManaCost[rank]
	level := WindfuryTotemLevel[rank]

	// The totem's buff takes the totem slot of the shaman's own weapon, so it works next to
	// Rockbiter, Flametongue or Frostbrand Weapon and an oil or stone. It does not work next
	// to Flametongue Totem's buff. Windfury Weapon on the main hand turns it off, and then it
	// leaves the slot to Flametongue Totem.
	var buffAura *core.Aura
	if shaman.HasMHWeapon() {
		buffAura = core.WindfuryTotemBuffAura(&shaman.Character, int32(rank), fmt.Sprintf("Windfury Totem (Rank %d)", rank))
	}

	periodicTriggerAura := shaman.RegisterAura(core.Aura{
		Label:    fmt.Sprintf("Windfury Trigger Dummy (Rank %d)", rank),
		ActionID: core.ActionID{SpellID: WindfuryBuffAuraId[rank]},
		Duration: time.Minute * 5, // Forever totems last 5 min, Classic 2
		OnGain: func(_ *core.Aura, sim *core.Simulation) {
			if buffAura != nil {
				shaman.setTotemWeaponBuff(sim, buffAura, proto.WeaponImbue_WindfuryWeapon)
			}
		},
		OnExpire: func(_ *core.Aura, sim *core.Simulation) {
			if buffAura != nil {
				shaman.clearTotemWeaponBuff(sim, buffAura)
			}
		},
	})

	spell := shaman.newTotemSpellConfig(manaCost, spellId)
	spell.RequiredLevel = level
	spell.Rank = rank
	spell.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		shaman.setActiveAirTotem(sim, spell, periodicTriggerAura)
	}
	return spell
}

const GraceOfAirTotemRanks = 3

var GraceOfAirTotemSpellId = [GraceOfAirTotemRanks + 1]int32{0, 8835, 10627, 25359}
var GraceOfAirTotemManaCost = [GraceOfAirTotemRanks + 1]float64{0, 155, 250, 310}
var GraceOfAirTotemLevel = [GraceOfAirTotemRanks + 1]int{0, 42, 56, 60}

func (shaman *Shaman) registerGraceOfAirTotemSpell() {
	shaman.GraceOfAirTotem = make([]*core.Spell, GraceOfAirTotemRanks+1)

	for rank := 1; rank <= GraceOfAirTotemRanks; rank++ {
		// Only the ranks the level has learned, so nothing below (a totem buff aura) is
		// built for a rank the shaman cannot cast.
		if GraceOfAirTotemLevel[rank] <= int(shaman.Level) {
			config := shaman.newGraceOfAirTotemSpellConfig(rank)
			shaman.GraceOfAirTotem[rank] = shaman.RegisterSpell(config)
		}
	}

	shaman.AirTotems = append(
		shaman.AirTotems,
		core.FilterSlice(shaman.GraceOfAirTotem, func(spell *core.Spell) bool { return spell != nil })...,
	)
}

func (shaman *Shaman) newGraceOfAirTotemSpellConfig(rank int) core.SpellConfig {
	spellId := GraceOfAirTotemSpellId[rank]
	manaCost := GraceOfAirTotemManaCost[rank]
	level := GraceOfAirTotemLevel[rank]

	buffAura := ownTotemAura(core.GraceOfAirTotemAura(&shaman.Unit, 1))

	spell := shaman.newTotemSpellConfig(manaCost, spellId)
	spell.RequiredLevel = level
	spell.Rank = rank
	spell.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		shaman.setActiveAirTotem(sim, spell, buffAura)
	}
	return spell
}

const WindwallTotemRanks = 3

var WindwallTotemSpellId = [WindwallTotemRanks + 1]int32{0, 15107, 15111, 15112}
var WindwallTotemManaCost = [WindwallTotemRanks + 1]float64{0, 115, 170, 225}
var WindwallTotemLevel = [WindwallTotemRanks + 1]int{0, 36, 46, 56}

func (shaman *Shaman) registerWindwallTotemSpell() {
	shaman.WindwallTotem = make([]*core.Spell, WindwallTotemRanks+1)

	for rank := 1; rank <= WindwallTotemRanks; rank++ {
		// Only the ranks the level has learned, so nothing below (a totem buff aura) is
		// built for a rank the shaman cannot cast.
		if WindwallTotemLevel[rank] <= int(shaman.Level) {
			config := shaman.newWindwallTotemSpellConfig(rank)
			shaman.WindwallTotem[rank] = shaman.RegisterSpell(config)
		}
	}

	shaman.AirTotems = append(
		shaman.AirTotems,
		core.FilterSlice(shaman.WindwallTotem, func(spell *core.Spell) bool { return spell != nil })...,
	)
}

func (shaman *Shaman) newWindwallTotemSpellConfig(rank int) core.SpellConfig {
	spellId := WindwallTotemSpellId[rank]
	manaCost := WindwallTotemManaCost[rank]
	level := WindwallTotemLevel[rank]

	spell := shaman.newTotemSpellConfig(manaCost, spellId)
	spell.RequiredLevel = level
	spell.Rank = rank
	spell.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		shaman.setActiveAirTotem(sim, spell, nil)
	}
	return spell
}
