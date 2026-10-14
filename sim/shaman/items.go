package shaman

import (
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const (
	// Keep these ordered by ID
	NaturalAlignmentCrystal  = 19344
	WushoolaysCharmOfSpirits = 19956
	TotemOfRage              = 22395
	TotemOfTheStorm          = 23199
	TotemOfThunder           = 228176
	PolishedDriftwoodIcon    = 249398
	TotemOfChargedFlames     = 263412
	FirestormTotem           = 263436
	ForeverTotemOfTheStorm   = 272432
	BurningTotem             = 272433
	RageOfTheStorm           = 280604
)

func init() {
	core.AddEffectsToTest = false

	// Keep these ordered by name

	// https://www.wowhead.com/classic/item=19344/natural-alignment-crystal
	// Use: Aligns the Shaman with nature, increasing the damage done by spells by 20%, improving heal effects by 20%, and increasing mana cost of spells by 20% for 20 sec.
	// (5 Min Cooldown)
	core.NewItemEffect(NaturalAlignmentCrystal, func(agent core.Agent) {
		shaman := agent.(ShamanAgent).GetShaman()

		duration := time.Second * 20

		aura := shaman.RegisterAura(core.Aura{
			ActionID: core.ActionID{ItemID: NaturalAlignmentCrystal},
			Label:    "Nature Aligned",
			Duration: duration,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				shaman.PseudoStats.SchoolDamageDealtMultiplier.MultiplyMagicSchools(1.20)
				shaman.PseudoStats.HealingDealtMultiplier *= 1.20
				shaman.PseudoStats.SchoolCostMultiplier.AddToAllSchools(20)
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				shaman.PseudoStats.SchoolDamageDealtMultiplier.MultiplyMagicSchools(1 / 1.20)
				shaman.PseudoStats.HealingDealtMultiplier /= 1.20
				shaman.PseudoStats.SchoolCostMultiplier.AddToAllSchools(-20)
			},
		})

		spell := shaman.RegisterSpell(core.SpellConfig{
			ActionID: core.ActionID{ItemID: NaturalAlignmentCrystal},
			ProcMask: core.ProcMaskEmpty,
			Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagOffensiveEquipment,
			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    shaman.NewTimer(),
					Duration: time.Minute * 5,
				},
				SharedCD: core.Cooldown{
					Timer:    shaman.GetOffensiveTrinketCD(),
					Duration: duration,
				},
			},
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				aura.Activate(sim)
			},
		})

		shaman.AddMajorCooldown(core.MajorCooldown{
			Spell:    spell,
			Priority: core.CooldownPriorityBloodlust,
			Type:     core.CooldownTypeDPS,
		})
	})

	// https://www.wowhead.com/forever/item=272433/burning-totem
	// Equip: Increases the duration of your Flame Shock ability by 3 sec.
	//
	// Flame Shock ticks every 3 sec, so this is one more tick.
	core.NewItemEffect(BurningTotem, func(agent core.Agent) {
		shaman := agent.(ShamanAgent).GetShaman()
		core.MakePermanent(shaman.RegisterAura(core.Aura{
			Label: "Burning Totem",
			OnInit: func(aura *core.Aura, sim *core.Simulation) {
				for _, spell := range shaman.FlameShock {
					if spell == nil {
						continue
					}
					for _, dot := range spell.Dots() {
						if dot != nil {
							dot.NumberOfTicks += 1
							dot.RecomputeAuraDuration()
						}
					}
				}
			},
		}))
	})

	// https://www.wowhead.com/forever/item=263436/firestorm-totem
	// Equip: Reduces the mana cost of your totems by 5%.
	//
	// The client's mask (1.60.1.70338, spell 1270492) holds every totem and not the Fire Nova
	// spell. We multiply it with Totemic Focus, the way The Earthshatterer's 2 piece does.
	core.NewItemEffect(FirestormTotem, func(agent core.Agent) {
		shaman := agent.(ShamanAgent).GetShaman()
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagTotem) && spell.Cost != nil {
				spell.Cost.Factor *= .95
			}
		})
	})

	// https://www.wowhead.com/forever/item=263412/totem-of-charged-flames
	// Equip: Reduces the mana cost of Flame Shock and Lightning Bolt by 5 Mana.
	//
	// The client (spell 1270478) takes 5 off before the percent cost cuts.
	core.NewItemEffect(TotemOfChargedFlames, func(agent core.Agent) {
		shaman := agent.(ShamanAgent).GetShaman()
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			if (spell.SpellCode == SpellCode_ShamanFlameShock || spell.SpellCode == SpellCode_ShamanLightningBolt) && spell.Cost != nil {
				spell.Cost.FlatModifier -= 5
			}
		})
	})

	// https://www.wowhead.com/forever/item=228176/totem-of-thunder
	// Equip: Increases the critical strike chance of Lightning Bolt by 1%.
	//
	// Lightning Overload's copies share the bolt's spell code, so they get it too.
	core.NewItemEffect(TotemOfThunder, func(agent core.Agent) {
		shaman := agent.(ShamanAgent).GetShaman()
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellCode == SpellCode_ShamanLightningBolt {
				spell.BonusCritRating += 1 * core.SpellCritRatingPerCritChance
			}
		})
	})

	// https://www.wowhead.com/forever/item=272432/totem-of-the-storm
	// Equip: Your Lightning Bolt ability can now also trigger the Maelstrom Weapon talent, but with a 50% reduced chance.
	//
	// Forever reused the name of the Classic Totem of the Storm (23199) for a new item.
	core.NewItemEffect(ForeverTotemOfTheStorm, func(agent core.Agent) {
		agent.(ShamanAgent).GetShaman().lightningBoltTriggersMaelstrom = true
	})

	// https://www.wowhead.com/forever/item=249398/polished-driftwood-icon
	// Equip: Allows 8% of your Mana regeneration to continue while casting.
	//
	// Enchanters make it (Enchanting, Forever).
	core.NewItemEffect(PolishedDriftwoodIcon, func(agent core.Agent) {
		agent.(ShamanAgent).GetShaman().PseudoStats.SpiritRegenRateCasting += .08
	})

	// https://www.wowhead.com/forever/item=280604/rage-of-the-storm
	// Equip: Increases the damage dealt by your Stormstrike ability by 10%.
	core.NewItemEffect(RageOfTheStorm, func(agent core.Agent) {
		shaman := agent.(ShamanAgent).GetShaman()
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellCode == SpellCode_ShamanStormstrike {
				spell.DamageMultiplierAdditive += 0.10
			}
		})
	})

	// https://www.wowhead.com/classic/item=23199/totem-of-the-storm
	// Equip: Increases damage done by Chain Lightning and Lightning Bolt by up to 33.
	core.NewItemEffect(TotemOfTheStorm, func(agent core.Agent) {
		shaman := agent.(ShamanAgent).GetShaman()
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellCode == SpellCode_ShamanLightningBolt || spell.SpellCode == SpellCode_ShamanChainLightning {
				spell.BonusDamage += 33
			}
		})
	})

	// Totem of Rage
	// Equip: Increases damage done by Earth Shock, Flame Shock, and Frost Shock by up to 30.
	// Acts as extra 30 spellpower for shocks.
	core.NewItemEffect(TotemOfRage, func(agent core.Agent) {
		shaman := agent.(ShamanAgent).GetShaman()
		affectedSpellCodes := []int32{SpellCode_ShamanEarthShock, SpellCode_ShamanFlameShock, SpellCode_ShamanFrostShock}
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			if slices.Contains(affectedSpellCodes, spell.SpellCode) {
				spell.BonusDamage += 30
			}
		})
	})

	// https://www.wowhead.com/classic/item=19956/wushoolays-charm-of-spirits
	// Use: Increases the damage dealt by your Lightning Shield spell by 100% for 20 sec. (3 Min Cooldown)
	core.NewItemEffect(WushoolaysCharmOfSpirits, func(agent core.Agent) {
		shaman := agent.(ShamanAgent).GetShaman()

		duration := time.Second * 20
		actionID := core.ActionID{ItemID: WushoolaysCharmOfSpirits}

		var affectedSpells []*core.Spell

		aura := shaman.RegisterAura(core.Aura{
			ActionID: actionID,
			Label:    "Wushoolay's Charm of Spirits",
			Duration: time.Second * 20,
			OnInit: func(aura *core.Aura, sim *core.Simulation) {
				affectedSpells = core.FilterSlice(
					shaman.LightningShieldProcs,
					func(spell *core.Spell) bool { return spell != nil },
				)
			},
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				for _, spell := range affectedSpells {
					spell.DamageMultiplier *= 2
				}
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				for _, spell := range affectedSpells {
					spell.DamageMultiplier /= 2
				}
			},
		})

		spell := shaman.RegisterSpell(core.SpellConfig{
			ActionID:    actionID,
			SpellSchool: core.SpellSchoolNature,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagOffensiveEquipment,
			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    shaman.NewTimer(),
					Duration: time.Minute * 3,
				},
				SharedCD: core.Cooldown{
					Timer:    shaman.GetOffensiveTrinketCD(),
					Duration: duration,
				},
			},
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				aura.Activate(sim)
			},
		})

		shaman.AddMajorCooldown(core.MajorCooldown{
			Spell:    spell,
			Priority: core.CooldownPriorityDefault,
			Type:     core.CooldownTypeDPS,
		})
	})

	core.AddEffectsToTest = true
}
