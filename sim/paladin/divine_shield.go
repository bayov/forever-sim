package paladin

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Divine Shield and Divine Protection under Forever (client 1.60.1.70753), the paladin's two
// immunities. They share one 5 min category cooldown (37), which Sacred Duty cuts by 30 sec a rank,
// apply Forbearance and can't be cast under it, and take the global cooldown.
//
//   - Divine Shield (642 at 34: 10 sec, 75 mana; 1020 at 50: 12 sec, 110 mana): "Protects the
//     paladin from all damage and spells for 12 sec, but reduces all damage you deal by 50%."
//     (aura 39 on every school, aura 79 at -50 on every school).
//   - Divine Protection (498 at 6: 6 sec, 15 mana; 5573 at 18: 8 sec, 35 mana): "You are protected
//     from all physical attacks and spells for 8 sec, but during that time you cannot attack or use
//     physical abilities yourself." (aura 39 on every school, and aura 25, pacify.)
//
// The sim takes all damage to 0 while one is up (the hits still land, so blocks and procs still
// roll), and Divine Protection stops the paladin's swings (pacify) for its duration; the spells a
// paladin tank casts are Holy, so it leaves them alone. Like Templar's Bulwark these are emergency
// buttons, and whether a boss keeps swinging at an immune tank or turns to the next on its threat
// list is not modelled. A bubble loses the paladin its aggro (user ruling 2026-11-04, standards 40),
// so no paladin tank rotation casts these (no_bubble_test.go in protection).
var DivineShieldSpellId = [...]int32{0, 642, 1020}
var DivineProtectionSpellId = [...]int32{0, 498, 5573}

func (paladin *Paladin) registerDivineShieldAndProtection() {
	if !paladin.Env.IsForever() {
		return
	}
	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Minute*5 - paladin.sacredDutyCooldownCut(),
	}

	var immune []*core.Aura
	paladin.AddDynamicDamageTakenModifier(func(_ *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		for _, aura := range immune {
			if aura.IsActive() {
				result.Damage = 0
				return
			}
		}
	})

	register := func(level int32, spellID int32, rank int, manaCost float64, duration time.Duration, label string, onGain, onExpire func(*core.Aura, *core.Simulation)) {
		if paladin.Level < level {
			return
		}
		aura := paladin.RegisterAura(core.Aura{
			Label:    label + paladin.Label + strconv.Itoa(rank),
			ActionID: core.ActionID{SpellID: spellID},
			Duration: duration,
			OnGain:   onGain,
			OnExpire: onExpire,
		})
		immune = append(immune, aura)
		paladin.RegisterSpell(core.SpellConfig{
			ActionID:      aura.ActionID,
			SpellSchool:   core.SpellSchoolHoly,
			ProcMask:      core.ProcMaskEmpty,
			Flags:         core.SpellFlagAPL | SpellFlag_Forbearance,
			RequiredLevel: int(level),
			Rank:          rank,
			ManaCost: core.ManaCostOptions{
				FlatCost:   manaCost,
				Multiplier: paladin.benediction(),
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{GCD: core.GCDDefault},
				CD:          cd,
			},
			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				aura.Activate(sim)
			},
		})
	}

	halveDamage := func(aura *core.Aura, _ *core.Simulation) { aura.Unit.PseudoStats.DamageDealtMultiplier *= 0.5 }
	restoreDamage := func(aura *core.Aura, _ *core.Simulation) { aura.Unit.PseudoStats.DamageDealtMultiplier /= 0.5 }
	register(34, 642, 1, 75, time.Second*10, "Divine Shield", halveDamage, restoreDamage)
	register(50, 1020, 2, 110, time.Second*12, "Divine Shield", halveDamage, restoreDamage)

	pacify := func(_ *core.Aura, sim *core.Simulation) { paladin.AutoAttacks.CancelAutoSwing(sim) }
	unpacify := func(_ *core.Aura, sim *core.Simulation) { paladin.AutoAttacks.EnableAutoSwing(sim) }
	register(6, 498, 1, 15, time.Second*6, "Divine Protection", pacify, unpacify)
	register(18, 5573, 2, 35, time.Second*8, "Divine Protection", pacify, unpacify)
}
