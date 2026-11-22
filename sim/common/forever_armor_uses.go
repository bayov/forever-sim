package common

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// The "Use:" effects of cloth and leather armor that the Phase 0 gear search meets (2026-11-22).
// Values are from client 1.60.1.71037 (wago ItemXItemEffect, ItemEffect, SpellEffect, SpellMisc,
// SpellDuration), and the Classic ones from the Classic tooltips.
//
// Two uses in the same pools aren't modelled because they do nothing for our damage:
//   - Bloodsail Admiral's Hat (12185): Summon Blood Parrot (17567), a pet bird with no combat use.
//   - Robe of the Void (14153): Heal Pet (18386) heals the warlock's demon for 300 to 900, 10 min
//     cooldown. The sim's demons aren't attacked.
const (
	HelmOfFire        = 8348
	CircleOfFlame     = 11808
	RobeOfTheArchmage = 14152
)

func init() {
	// https://www.wowhead.com/forever/item=8348/helm-of-fire
	// Use: Fireball (10578) hits for 331 Fire with variance 0.272, 286 to 376, and burns for 33
	// every 2 sec for 8 sec (Classic: 10 a tick, 40 in all). Forever has a 5 min cooldown
	// (Classic 15 min) and 10 sec on category 1141 ("Item - Burst Trinket"), which the other
	// Forever burst items share (client 1.60.1.71037 ItemEffect 98606). No spell power coefficient.
	core.NewItemEffect(HelmOfFire, func(agent core.Agent) {
		character := agent.GetCharacter()
		forever := character.Env.IsForever()
		tickDamage := core.TernaryFloat64(forever, 33, 10)
		cast := core.CastConfig{
			CD: core.Cooldown{
				Timer:    character.NewTimer(),
				Duration: core.TernaryDuration(forever, time.Minute*5, time.Minute*15),
			},
		}
		if forever {
			cast.SharedCD = core.Cooldown{
				Timer:    character.GetOffensiveTrinketCD(),
				Duration: time.Second * 10,
			}
		}

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{ItemID: HelmOfFire},
			SpellSchool: core.SpellSchoolFire,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagOffensiveEquipment,
			Cast:        cast,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			Dot: core.DotConfig{
				Aura: core.Aura{
					Label: "Helm of Fire Fireball",
				},
				TickLength:    2 * time.Second,
				NumberOfTicks: 4,
				OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
					dot.Snapshot(target, tickDamage, isRollover)
				},
				OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
					dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
				},
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				result := spell.CalcAndDealDamage(sim, target, sim.Roll(286, 376), spell.OutcomeMagicHitAndCrit)
				if result.Landed() {
					spell.Dot(target).Apply(sim)
				}
			},
		})

		character.AddMajorCooldown(core.MajorCooldown{
			Type:  core.CooldownTypeDPS,
			Spell: spell,
		})
	})

	// https://www.wowhead.com/forever/item=11808/circle-of-flame
	// Use: Circle of Flame (17447) is a 10 sec channel that gives 79 mana with variance 0.5, 59 to
	// 99, every 1 sec and takes as much health (Classic: 75 a tick). 5 min cooldown (client
	// 1.60.1.71037 ItemEffect 99200, SpellMisc DurationIndex 1 is 10000 ms). We don't model the
	// health. The channel holds our casts for 10 sec, so we press it only when we're 1000 mana or
	// more short, and the gear search sees it cost damage in a short fight.
	core.NewItemEffect(CircleOfFlame, func(agent core.Agent) {
		character := agent.GetCharacter()
		if !character.HasManaBar() {
			return
		}
		perTickMin, perTickMax := 59.25, 98.75
		if !character.Env.IsForever() {
			perTickMin, perTickMax = 56.25, 93.75
		}
		actionID := core.ActionID{SpellID: 17447}
		manaMetrics := character.NewManaMetrics(actionID)
		spell := character.RegisterSpell(core.SpellConfig{
			ActionID: actionID,
			ProcMask: core.ProcMaskEmpty,
			Flags:    core.SpellFlagHelpful | core.SpellFlagChanneled | core.SpellFlagNoOnCastComplete,
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 5,
				},
			},
			Hot: core.DotConfig{
				SelfOnly: true,
				Aura: core.Aura{
					Label: "Circle of Flame",
				},
				NumberOfTicks: 10,
				TickLength:    time.Second,
				OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
					character.AddMana(sim, sim.Roll(perTickMin, perTickMax), manaMetrics)
				},
			},
			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				spell.SelfHot().Apply(sim)
			},
		})
		character.AddMajorCooldown(core.MajorCooldown{
			Type:  core.CooldownTypeMana,
			Spell: spell,
			ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
				return character.MaxMana()-character.CurrentMana() >= 1000
			},
			ShouldActivateNeedsEvent: true,
		})
	})

	// https://www.wowhead.com/forever/item=14152/robe-of-the-archmage
	// Use: Replenish Mana (18385) restores 500 mana with variance 0.5, 375 to 625. 5 min cooldown,
	// the same as Classic, and no category cooldown, so it doesn't share one with mana gems or
	// runes (client 1.60.1.71037 ItemEffect 99905). It's off the GCD, so we press it once we're
	// 625 mana short.
	core.NewItemEffect(RobeOfTheArchmage, func(agent core.Agent) {
		character := agent.GetCharacter()
		if !character.HasManaBar() {
			return
		}
		actionID := core.ActionID{SpellID: 18385}
		manaMetrics := character.NewManaMetrics(actionID)
		spell := character.RegisterSpell(core.SpellConfig{
			ActionID: actionID,
			ProcMask: core.ProcMaskEmpty,
			Flags:    core.SpellFlagNoOnCastComplete,
			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 5,
				},
			},
			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				character.AddMana(sim, sim.Roll(375, 625), manaMetrics)
			},
		})
		character.AddMajorCooldown(core.MajorCooldown{
			Type:  core.CooldownTypeMana,
			Spell: spell,
			ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
				return character.MaxMana()-character.CurrentMana() >= 625
			},
			ShouldActivateNeedsEvent: true,
		})
	})
}
