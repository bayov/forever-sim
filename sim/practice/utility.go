//go:build practice

package practice

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/shaman"
)

// Spells the sim doesn't have (docs/practice_room.md, "Your toolkit", kind 3): Ghost Wolf, Purge,
// the utility totems and the like. The session registers them with their cost, cast time, GCD,
// cooldown and range. Effects that change the fight are coded here: Ghost Wolf's speed, and the
// totems taking their element's slot. The rest only cost their GCD and mana.
//
// The values are the top ranks' from wowhead's Classic tooltips. Reading them from the client's
// spell tables, every rank included, is the next step.

type utilitySpell struct {
	id       int32
	level    int32
	flatCost float64
	// baseCost is a share of base mana, for spells priced that way.
	baseCost float64
	cast     time.Duration
	cd       time.Duration
	harmful  bool
	maxRange float64
	// totem is the element slot a totem goes in, -1 for a spell that isn't a totem.
	totem         int
	totemDuration time.Duration
	ghostWolf     bool
}

const ghostWolfID = 2645

var shamanUtility = []utilitySpell{
	{id: ghostWolfID, level: 20, flatCost: 100, cast: 3 * time.Second, totem: -1, ghostWolf: true},
	{id: 8012, level: 32, baseCost: 0.10, harmful: true, maxRange: 30, totem: -1}, // Purge rank 2 is 8012
	{id: 526, level: 16, baseCost: 0.09, maxRange: 30, totem: -1},                 // Cure Poison
	{id: 2870, level: 22, baseCost: 0.09, maxRange: 30, totem: -1},                // Cure Disease
	{id: 546, level: 28, flatCost: 95, maxRange: 30, totem: -1},                   // Water Walking
	{id: 131, level: 22, flatCost: 50, maxRange: 30, totem: -1},                   // Water Breathing
	{id: 6196, level: 26, flatCost: 80, cast: 2 * time.Second, totem: -1},         // Far Sight
	{id: 556, level: 30, flatCost: 150, cast: 10 * time.Second, cd: 15 * time.Minute, totem: -1}, // Astral Recall
	{id: 2484, level: 6, baseCost: 0.06, cd: 15 * time.Second, totem: shaman.EarthTotem, totemDuration: 45 * time.Second},     // Earthbind
	{id: 10428, level: 58, flatCost: 140, cd: 30 * time.Second, totem: shaman.EarthTotem, totemDuration: 15 * time.Second},    // Stoneclaw rank 6
	{id: 10479, level: 54, flatCost: 180, totem: shaman.FireTotem, totemDuration: 2 * time.Minute},                         // Frost Resistance
	{id: 10538, level: 58, flatCost: 180, totem: shaman.WaterTotem, totemDuration: 2 * time.Minute},                        // Fire Resistance
	{id: 10601, level: 60, flatCost: 180, totem: shaman.AirTotem, totemDuration: 2 * time.Minute},                          // Nature Resistance
	{id: 8177, level: 30, baseCost: 0.06, cd: 15 * time.Second, totem: shaman.AirTotem, totemDuration: 45 * time.Second},     // Grounding
	{id: 8166, level: 22, baseCost: 0.10, totem: shaman.WaterTotem, totemDuration: 2 * time.Minute},                        // Poison Cleansing
	{id: 8170, level: 38, baseCost: 0.10, totem: shaman.WaterTotem, totemDuration: 2 * time.Minute},                        // Disease Cleansing
	{id: 25908, level: 50, baseCost: 0.08, totem: shaman.AirTotem, totemDuration: 2 * time.Minute},                         // Tranquil Air
	{id: 6495, level: 34, flatCost: 65, totem: shaman.AirTotem, totemDuration: 5 * time.Minute},                            // Sentry
}

// registerUtility registers the spells the sim doesn't have, for a shaman, while the sim is built.
func (r *run) registerUtility(sh *shaman.Shaman) {
	for _, u := range shamanUtility {
		u := u
		if sh.Level < u.level || sh.GetSpell(core.ActionID{SpellID: u.id}) != nil {
			continue
		}
		gcd := core.GCDDefault
		factor := 1.0
		if u.totem >= 0 {
			gcd = time.Second
			factor = sh.TotemicFocusFactor()
		}
		cast := u.cast
		if u.ghostWolf {
			// Improved Ghost Wolf takes 0.5 sec off per point.
			cast -= time.Duration(sh.Talents.ImprovedGhostWolf) * 500 * time.Millisecond
		}
		flags := core.SpellFlagAPL | shaman.SpellFlagShaman
		procMask := core.ProcMaskEmpty
		if u.harmful {
			procMask = core.ProcMaskSpellDamage
		} else {
			flags |= core.SpellFlagHelpful
		}
		config := core.SpellConfig{
			ActionID:    core.ActionID{SpellID: u.id},
			SpellSchool: core.SpellSchoolNature,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    procMask,
			Flags:       flags,
			MaxRange:    u.maxRange,
			ManaCost:    core.ManaCostOptions{FlatCost: u.flatCost, BaseCost: u.baseCost, Factor: factor},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{GCD: gcd, CastTime: cast},
			},
			DamageMultiplier: 1,
			ThreatMultiplier: 1,
		}
		if u.cd > 0 {
			config.Cast.CD = core.Cooldown{Timer: sh.NewTimer(), Duration: u.cd}
		}
		switch {
		case u.ghostWolf:
			r.ghostWolf = sh.RegisterAura(core.Aura{
				Label:    "Ghost Wolf",
				ActionID: core.ActionID{SpellID: ghostWolfID},
				Duration: core.NeverExpires,
			})
			config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				r.ghostWolf.Activate(sim)
			}
		case u.totem >= 0:
			slot := u.totem
			config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				sh.PlaceTotem(sim, slot, spell, u.totemDuration)
			}
		case u.harmful:
			config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
			}
		default:
			config.ApplyEffects = func(*core.Simulation, *core.Unit, *core.Spell) {}
		}
		sh.RegisterSpell(config)
	}
}

// inGhostWolf is whether you're in Ghost Wolf form now.
func (r *run) inGhostWolf() bool {
	return r.ghostWolf != nil && r.ghostWolf.IsActive()
}

func (r *run) isUtilityTotem(id int32) bool {
	for _, u := range shamanUtility {
		if u.id == id {
			return u.totem >= 0
		}
	}
	return false
}
