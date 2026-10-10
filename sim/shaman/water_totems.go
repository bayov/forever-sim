package shaman

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const HealingStreamTotemRanks = 5

var HealingStreamTotemSpellId = [HealingStreamTotemRanks + 1]int32{0, 5394, 6375, 6377, 10462, 10463}
var HealingStreamTotemHealId = [HealingStreamTotemRanks + 1]int32{0, 5672, 6371, 6372, 10460, 10461}

// Forever's ranks heal less than Classic's 6, 8, 10, 12 and 14.
var HealingStreamTotemBaseHealing = [HealingStreamTotemRanks + 1]float64{0, 5, 6, 7, 9, 11}
var HealingStreamTotemSpellCoeff = [HealingStreamTotemRanks + 1]float64{0, .022, .022, .022, .022, .022}
var HealingStreamTotemManaCost = [HealingStreamTotemRanks + 1]float64{0, 40, 50, 60, 70, 80}
var HealingStreamTotemLevel = [HealingStreamTotemRanks + 1]int{0, 20, 30, 40, 50, 60}

func (shaman *Shaman) registerHealingStreamTotemSpell() {
	shaman.HealingStreamTotem = make([]*core.Spell, HealingStreamTotemRanks+1)

	for rank := 1; rank <= HealingStreamTotemRanks; rank++ {
		// Only the ranks the level has learned, so nothing below (a totem buff aura) is
		// built for a rank the shaman cannot cast.
		if HealingStreamTotemLevel[rank] <= int(shaman.Level) {
			config := shaman.newHealingStreamTotemSpellConfig(rank)
			shaman.HealingStreamTotem[rank] = shaman.RegisterSpell(config)
		}
	}

	shaman.WaterTotems = append(
		shaman.WaterTotems,
		core.FilterSlice(shaman.HealingStreamTotem, func(spell *core.Spell) bool { return spell != nil })...,
	)
}

func (shaman *Shaman) newHealingStreamTotemSpellConfig(rank int) core.SpellConfig {
	spellId := HealingStreamTotemSpellId[rank]
	healId := HealingStreamTotemHealId[rank]
	// Restorative Totems adds 10% a point, but Purification adds nothing.
	//
	// Purification's tooltip says "your healing spells", but in the Forever client (70291) and
	// the Classic Era client (1.15.9) it only lists Healing Wave, Lesser Healing Wave and Chain
	// Heal, and we have none of those. Healing Stream's heal isn't on the list in either client.
	baseHealing := HealingStreamTotemBaseHealing[rank] * (1 + shaman.restorativeTotemsModifier())
	spellCoeff := HealingStreamTotemSpellCoeff[rank]
	manaCost := HealingStreamTotemManaCost[rank]
	level := HealingStreamTotemLevel[rank]

	duration := time.Minute * 5 // Forever totems last 5 min, Classic 1
	healInterval := time.Second * 2

	config := shaman.newTotemSpellConfig(manaCost, spellId)
	config.RequiredLevel = level
	config.Rank = rank

	healSpell := shaman.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: healId},
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       core.SpellFlagHelpful | core.SpellFlagNoOnCastComplete | core.SpellFlagNoLogs | core.SpellFlagNoMetrics,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: spellCoeff,

		// Under Forever each heal can crit for 1.5 times, at our spell crit at the time of the heal.
		//
		// The Forever client (70338) gives the heal the flag that lets periodic effects crit,
		// like Flame Shock's ticks (shaman_audit.md 5.7). Classic Era's client doesn't. Tidal
		// Mastery and Elemental Fury don't list Healing Stream in the client, so they don't
		// touch its crits. The heal is a periodic aura in the client, and Water Shield's proc
		// flags leave periodic heals out, so a crit here never spends a Water Shield globe.
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			outcome := spell.OutcomeHealing
			if sim.IsForever() {
				outcome = spell.OutcomeHealingCrit
			}
			spell.DealPeriodicHealing(sim, spell.CalcHealing(sim, target, baseHealing, outcome))
		},
	})

	config.Hot = core.DotConfig{
		Aura: core.Aura{
			Label: fmt.Sprintf("Healing Stream HoT (Rank %d)", rank),
		},
		NumberOfTicks: int32(duration / healInterval),
		TickLength:    healInterval,
		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			healSpell.Cast(sim, target)
		},
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		shaman.TotemExpirations[WaterTotem] = sim.CurrentTime + duration
		shaman.ActiveTotems[WaterTotem] = spell

		for _, agent := range shaman.Party.Players {
			spell.Hot(&agent.GetCharacter().Unit).Activate(sim)
		}
	}

	return config
}

const ManaSpringTotemRanks = 4

var ManaSpringTotemSpellId = [ManaSpringTotemRanks + 1]int32{0, 5675, 10495, 10496, 10497}
var ManaSpringTotemManaRestore = [ManaSpringTotemRanks + 1]int32{0, 4, 6, 8, 10}
var ManaSpringTotemManaCost = [ManaSpringTotemRanks + 1]float64{0, 40, 60, 80, 100}
var ManaSpringTotemLevel = [ManaSpringTotemRanks + 1]int{0, 26, 36, 46, 56}

func (shaman *Shaman) registerManaSpringTotemSpell() {
	shaman.ManaSpringTotem = make([]*core.Spell, ManaSpringTotemRanks+1)

	for rank := 1; rank <= ManaSpringTotemRanks; rank++ {
		// Only the ranks the level has learned, so nothing below (a totem buff aura) is
		// built for a rank the shaman cannot cast.
		if ManaSpringTotemLevel[rank] <= int(shaman.Level) {
			config := shaman.newManaSpringTotemSpellConfig(rank)
			shaman.ManaSpringTotem[rank] = shaman.RegisterSpell(config)
		}
	}

	shaman.WaterTotems = append(
		shaman.WaterTotems,
		core.FilterSlice(shaman.ManaSpringTotem, func(spell *core.Spell) bool { return spell != nil })...,
	)
}

func (shaman *Shaman) newManaSpringTotemSpellConfig(rank int) core.SpellConfig {
	spellId := ManaSpringTotemSpellId[rank]
	manaCost := ManaSpringTotemManaCost[rank]
	level := ManaSpringTotemLevel[rank]

	// The totem gives its mana to the shaman every 2 sec. Restorative Totems adds 5% per
	// point under Forever. The raid buff in buffs.go stays for other players' totems.
	manaRestore := float64(ManaSpringTotemManaRestore[rank]) * (1 + 0.05*float64(shaman.Talents.RestorativeTotems))

	// The ticks get their own tag, because the totem's cast cost is already reported under
	// the plain spell ID. With both under one ID the results of a multi threaded sim and a
	// single threaded one did not match.
	manaMetrics := shaman.NewManaMetrics(core.ActionID{SpellID: spellId, Tag: 1})

	var tick *core.PendingAction
	aura := shaman.RegisterAura(core.Aura{
		Label:    fmt.Sprintf("Mana Spring Totem (Rank %d)", rank),
		ActionID: core.ActionID{SpellID: spellId},
		Duration: time.Minute * 5, // Forever totems last 5 min, Classic 1
		OnGain: func(_ *core.Aura, sim *core.Simulation) {
			tick = core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				Period: time.Second * 2,
				OnAction: func(sim *core.Simulation) {
					shaman.AddMana(sim, manaRestore, manaMetrics)
				},
			})
		},
		OnExpire: func(_ *core.Aura, sim *core.Simulation) {
			tick.Cancel(sim)
		},
	})

	shaman.manaSpringAuras = append(shaman.manaSpringAuras, aura)

	spell := shaman.newTotemSpellConfig(manaCost, spellId)
	spell.RequiredLevel = level
	spell.Rank = rank
	spell.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		shaman.TotemExpirations[WaterTotem] = sim.CurrentTime + aura.Duration
		shaman.ActiveTotems[WaterTotem] = spell
		// A new water totem replaces the old one, so we restart the ticks from the drop.
		aura.Deactivate(sim)
		aura.Activate(sim)
	}
	return spell
}

const ManaTideTotemRanks = 3

// Forever's ranks (ForeverChanges spellbook, build 70009). Rank 1 comes at level 25 and gives
// 88 mana a tick where Classic's gives 170, rank 2 gives 197 where Classic's gives 230, and
// rank 3 is Classic's 290.
var ManaTideTotemSpellId = [ManaTideTotemRanks + 1]int32{0, 16190, 17354, 17359}
var ManaTideTotemManaRestore = [ManaTideTotemRanks + 1]float64{0, 88, 197, 290}
var ManaTideTotemManaCost = [ManaTideTotemRanks + 1]float64{0, 10, 30, 60}
var ManaTideTotemLevel = [ManaTideTotemRanks + 1]int{0, 25, 48, 58}

// Mana Tide Totem is a water totem that stands for 12 sec and gives the group its mana every
// 3 sec, 4 times. It takes the water slot, so Mana Spring Totem goes away and has to be put
// down again once the tide is over. All ranks share the 5 min cooldown.
func (shaman *Shaman) registerManaTideTotemSpell() {
	shaman.ManaTideTotem = make([]*core.Spell, ManaTideTotemRanks+1)
	if !shaman.Talents.ManaTideTotem {
		return
	}

	cooldown := shaman.NewTimer()
	for rank := 1; rank <= ManaTideTotemRanks; rank++ {
		if ManaTideTotemLevel[rank] <= int(shaman.Level) {
			shaman.ManaTideTotem[rank] = shaman.RegisterSpell(shaman.newManaTideTotemSpellConfig(rank, cooldown))
		}
	}

	shaman.WaterTotems = append(
		shaman.WaterTotems,
		core.FilterSlice(shaman.ManaTideTotem, func(spell *core.Spell) bool { return spell != nil })...,
	)
}

func (shaman *Shaman) newManaTideTotemSpellConfig(rank int, cooldown *core.Timer) core.SpellConfig {
	spellId := ManaTideTotemSpellId[rank]
	manaRestore := ManaTideTotemManaRestore[rank]
	duration := time.Second * 12

	var players []*core.Character
	var metrics []*core.ResourceMetrics
	for _, agent := range shaman.Party.Players {
		char := agent.GetCharacter()
		if char.HasManaBar() {
			players = append(players, char)
			metrics = append(metrics, char.NewManaMetrics(core.ActionID{SpellID: spellId, Tag: 1}))
		}
	}

	aura := shaman.RegisterAura(core.Aura{
		Label:    fmt.Sprintf("Mana Tide Totem (Rank %d)", rank),
		ActionID: core.ActionID{SpellID: spellId},
		Duration: duration,
		OnGain: func(_ *core.Aura, sim *core.Simulation) {
			core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				Period:   time.Second * 3,
				NumTicks: 4,
				OnAction: func(sim *core.Simulation) {
					for i, char := range players {
						char.AddMana(sim, manaRestore, metrics[i])
					}
				},
			})
		},
	})

	config := shaman.newTotemSpellConfig(ManaTideTotemManaCost[rank], spellId)
	config.RequiredLevel = ManaTideTotemLevel[rank]
	config.Rank = rank
	config.Cast.CD = core.Cooldown{
		Timer:    cooldown,
		Duration: time.Minute * 5,
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		shaman.TotemExpirations[WaterTotem] = sim.CurrentTime + duration
		shaman.ActiveTotems[WaterTotem] = spell
		for _, manaSpring := range shaman.manaSpringAuras {
			manaSpring.Deactivate(sim)
		}
		for _, healingStream := range shaman.HealingStreamTotem {
			if healingStream == nil {
				continue
			}
			for _, agent := range shaman.Party.Players {
				healingStream.Hot(&agent.GetCharacter().Unit).Deactivate(sim)
			}
		}
		aura.Activate(sim)
	}
	return config
}
