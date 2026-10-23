package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (warrior *Warrior) RegisterShieldWallCD() {
	duration := time.Second * 10
	//This is the inverse of the tooltip since it is a damage TAKEN coefficient
	damageTaken := 0.25
	cooldownBase := time.Minute * 30
	gcd := time.Duration(0)
	// Forever's Shield Wall takes 60% less damage for 12 sec on a 15 min cooldown and the global
	// cooldown (client 70425, spell 871).
	if warrior.Env.IsForever() {
		duration = time.Second * 12
		damageTaken = 0.4
		cooldownBase = time.Minute * 15
		gcd = core.GCDDefault
	}

	actionID := core.ActionID{SpellID: 871}
	swAura := warrior.RegisterAura(core.Aura{
		Label:    "Shield Wall",
		ActionID: actionID,
		Duration: duration,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			warrior.PseudoStats.DamageTakenMultiplier *= damageTaken
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			warrior.PseudoStats.DamageTakenMultiplier /= damageTaken
		},
	})

	// Improved Shield Wall takes 5.5 / 11 min off the cooldown instead of extending the duration
	// (client 70338, curve of spell 12312).
	cooldownDur := cooldownBase - time.Second*330*time.Duration(warrior.Talents.ImprovedShieldWall)

	swSpell := warrior.RegisterSpell(DefensiveStance, core.SpellConfig{
		ActionID: actionID,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: gcd,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: cooldownDur,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return warrior.PseudoStats.CanBlock
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			swAura.Activate(sim)
		},
	})

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: swSpell.Spell,
		Type:  core.CooldownTypeSurvival,
	})
}
