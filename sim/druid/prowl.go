package druid

import (
	"github.com/wowsims/classic/sim/core"
)

// Prowl under Forever (client rows 5215, 6783, 9913): Cat Form stealth, off the global cooldown
// with a 10 sec category cooldown, 30% slower movement, and the cancel auto attack bit
// (Attributes[0] 0x100000), so melee waits until a cast ends it. Its not-in-combat bit
// (Attributes[0] 0x10000000) keeps it out of the fight, and the fight is in combat from the
// pull, so the sim never casts it: no class starts the fight in Prowl (orchestration/standards.md
// item 13). A cat opens from stealth only through a Night Elf's Shadowmeld in combat.
var ProwlSpellId = []int32{0, 5215, 6783, 9913}
var ProwlLevel = []int32{0, 20, 40, 60}

func (druid *Druid) registerProwlSpell() {
	rank := core.HighestRankAt(druid.Level, ProwlLevel)
	if rank == 0 {
		return
	}
	actionID := core.ActionID{SpellID: ProwlSpellId[rank]}

	druid.ProwlAura = druid.RegisterAura(core.Aura{
		Label:    "Prowl",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			druid.AutoAttacks.CancelAutoSwing(sim)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.AutoAttacks.EnableAutoSwing(sim)
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.ActionID != actionID && spell.Flags.Matches(core.SpellFlagAPL|core.SpellFlagMCD) {
				aura.Deactivate(sim)
			}
		},
	})

	druid.Prowl = druid.RegisterSpell(Cat, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL | core.SpellFlagNoOnCastComplete,
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			// Out of combat only, and the sim's fights are in combat from start to end.
			return false
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.ProwlAura.Activate(sim)
		},
	})
}

// Stealthed is whether the abilities that need stealth (Ravage) can go: Prowl or Shadowmeld is up.
func (druid *Druid) Stealthed() bool {
	return (druid.ProwlAura != nil && druid.ProwlAura.IsActive()) || druid.IsShadowmelded()
}

// BreakStealth ends Prowl and Shadowmeld, as a stealth ability does when it is cast.
func (druid *Druid) BreakStealth(sim *core.Simulation) {
	if druid.ProwlAura != nil && druid.ProwlAura.IsActive() {
		druid.ProwlAura.Deactivate(sim)
	}
	druid.BreakShadowmeld(sim)
}
