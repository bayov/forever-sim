package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Item IDs
const (
	WolfsheadHelm       = 8345
	IdolOfFerocity      = 22397
	IdolOfTheMoon       = 23197
	IdolOfBrutality     = 23198
	RuneOfMetamorphosis = 19340

	// Forever's Phase 1 idols (1.60.1.70338 client).
	HowlingIdol  = 272427
	SwarmingIdol = 272430

	IdolOfTheDream = 220606

	EnragedIdol          = 272428
	IdolOfTheUrsineTwins = 279251
	TalonsOfWrath        = 249441
	MysticMushroom       = 249396
)

func init() {
	core.AddEffectsToTest = false

	// https://www.wowhead.com/classic/item=22397/idol-of-ferocity
	// Equip: Reduces the energy cost of Claw and Rake by 2 under Forever (3 in Classic). Its equip
	// spell 27851 is a flat cost modifier of -2 in client 1.60.1.70888 (SpellEffect 704563), and
	// hotfix push 113860 keeps -2: it only drops an unused word-1 bit from the class mask, and
	// Claw and Rake stay on word 0's 0x1000.
	core.NewItemEffect(IdolOfFerocity, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		reduction := core.TernaryInt32(druid.Env.IsForever(), 2, 3)

		druid.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellCode == SpellCode_DruidRake || spell.SpellCode == SpellCode_DruidClaw {
				spell.Cost.FlatModifier -= reduction
			}
		})
	})

	// https://www.wowhead.com/classic/item=23197/idol-of-the-moon
	// Equip: Increases the damage of your Moonfire spell by up to 33.
	core.NewItemEffect(IdolOfTheMoon, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		druid.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellCode == SpellCode_DruidMoonfire {
				spell.BonusDamage += 33
			}
		})
	})

	// Howling Idol, Forever. Equip: Reduces the cooldown of your Shifting Power ability by 0.5 sec
	// (spell 1291059, a flat cooldown modifier of -500 ms since client 1.60.1.70425; -1000 ms
	// in 70338).
	core.NewItemEffect(HowlingIdol, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		druid.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellCode == SpellCode_DruidShiftingPower {
				spell.CD.Duration -= time.Millisecond * 500
			}
		})
	})

	// Swarming Idol, Forever. Equip: Increases the duration of your Insect Swarm spell by 2 sec
	// (spell 1291061, a flat duration modifier of +2000 ms).
	core.NewItemEffect(SwarmingIdol, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		druid.InsectSwarmBonusDuration += time.Second * 2
	})

	// Idol of the Dream, Forever. Equip: Increases the duration of your Rip by 2 sec (spell 446212,
	// a flat duration modifier of +2000 ms on Rip's class mask), a seventh tick of the same damage.
	core.NewItemEffect(IdolOfTheDream, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		druid.RipBonusTicks += 1
	})

	// Talons of Wrath, Forever. A Wrath cast has a 50% chance to return 35 mana (1248996, ProcFlags
	// 0x10000, a harmful spell cast, which triggers 1302521, a 35 mana energize). The proc row names
	// no spell; the tooltip names Wrath.
	core.NewItemEffect(TalonsOfWrath, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		manaMetrics := druid.NewManaMetrics(core.ActionID{SpellID: 1302521})
		core.MakePermanent(druid.RegisterAura(core.Aura{
			Label: "Talons of Wrath",
			OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
				if spell.SpellCode == SpellCode_DruidWrath && sim.Proc(0.5, "Talons of Wrath") {
					druid.AddMana(sim, 35, manaMetrics)
				}
			},
		}))
	})

	// Mystic Mushroom, Forever. Equip: Increases your Spirit by 5% (spell 1248751).
	core.NewItemEffect(MysticMushroom, func(agent core.Agent) {
		character := agent.GetCharacter()
		character.MultiplyStat(stats.Spirit, 1.05)
	})

	// Idol of Brutality. Equip: Reduces the Rage cost of Maul and Swipe by 2 (28855 under Forever, a
	// flat cost modifier of -20 tenths on their shared class mask 0x800; 1.12's tooltip said 3).
	// Read in bear_attacks.go and swipe.go (brutalityRage).
	core.NewItemEffect(IdolOfBrutality, func(agent core.Agent) {})

	// Enraged Idol, Forever. Equip: Your Enrage ability generates an additional 10 Rage over its
	// duration (1291060: +10 tenths on Enrage's effect 0, its 1 sec tick, so 3 Rage a tick).
	core.NewItemEffect(EnragedIdol, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		druid.enrageTickBonus += 1
	})

	// Idol of the Ursine Twins, Forever. Equip: Your Lacerate hits have a 10% chance to reset the
	// cooldown on Primal Bite (1306483, ProcChance 10, on a landed melee ability). In bear_attacks.go.
	core.NewItemEffect(IdolOfTheUrsineTwins, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		druid.ursineTwins = true
	})

	// https://www.wowhead.com/classic/item=19340/rune-of-metamorphosis
	// Use: Decreases the mana cost of all Druid shapeshifting forms by 100% for 20 sec. (5 Min Cooldown)
	core.NewItemEffect(RuneOfMetamorphosis, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()

		actionID := core.ActionID{SpellID: 23724}
		duration := time.Second * 20
		cooldown := time.Minute * 5

		buffAura := druid.GetOrRegisterAura(core.Aura{
			ActionID: actionID,
			Label:    "Metamorphosis Rune",
			Duration: duration,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				druid.CatForm.Cost.Multiplier -= 100
				//druid.BearForm.Cost.Multiplier -= 100
				//druid.MoonkinForm.Cost.Multiplier -= 100
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				druid.CatForm.Cost.Multiplier += 100
				//druid.BearForm.Cost.Multiplier += 100
				//druid.MoonkinForm.Cost.Multiplier += 100
			},
		})

		spell := druid.GetOrRegisterSpell(core.SpellConfig{
			ActionID: actionID,
			Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagOffensiveEquipment,
			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    druid.NewTimer(),
					Duration: cooldown,
				},
			},
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				buffAura.Activate(sim)
			},
		})

		druid.AddMajorCooldown(core.MajorCooldown{
			Spell: spell,
			Type:  core.CooldownTypeDPS,
		})
	})

	core.AddEffectsToTest = true
}
