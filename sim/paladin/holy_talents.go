package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// The Holy tree's healing talents of the launch realms' October 30 hotfix (client 1.60.1.70753,
// tools/forever_talents/client/paladin.json). Healing Light, Holy Power and Spiritual Focus are on
// the heals themselves (healing.go), Divine Favor in divine_favor.go and Light's Vigil in
// lights_vigil.go.
func (paladin *Paladin) applyHealingTalents() {
	paladin.applyIllumination()
	paladin.applyInfusionOfLight()
	paladin.registerOathOfLight()
}

// Illumination (spell 20210): "After getting a critical effect from your Flash of Light, Holy
// Light, Light's Vigil, or Holy Shock heal spell you have a 20% chance to gain Mana equal to 50% of
// the base cost of the spell." 20% a rank of 5 (Classic's was 100% of the cost); effect 2 holds the
// 50%. The base cost is the rank's own, before Benediction and the like.
func (paladin *Paladin) applyIllumination() {
	if paladin.Talents.Illumination == 0 {
		return
	}
	actionID := core.ActionID{SpellID: 20272} // Illumination's mana, as Classic's log names it
	metrics := paladin.NewManaMetrics(actionID)
	chance := 0.2 * float64(paladin.Talents.Illumination)
	paladin.illuminate = func(sim *core.Simulation, baseCost float64) {
		if sim.Proc(chance, "Illumination") {
			paladin.AddMana(sim, 0.5*baseCost, metrics)
		}
	}

	core.MakePermanent(paladin.RegisterAura(core.Aura{
		Label: "Illumination",
		OnHealDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.DidCrit() || !isPaladinHeal(spell) {
				return
			}
			if cost, ok := paladin.illuminationBaseCost[spell]; ok {
				paladin.illuminate(sim, cost)
			} else if spell.Cost != nil {
				paladin.illuminate(sim, spell.Cost.BaseCost)
			}
		},
	}))
}

// Infusion of Light (spell 426065, the buff 437063): "Your Holy Shock and Flash of Light critical
// hits reduce the cast time of your next Holy Light cast within 15 sec by 0.5 sec" a rank of 2. The
// proc flags take both helpful and harmful spells, so Holy Shock's damage crits count too. The buff
// is one charge, spent by the next Holy Light that completes.
func (paladin *Paladin) applyInfusionOfLight() {
	if paladin.Talents.InfusionOfLight == 0 {
		return
	}
	reduction := time.Millisecond * 500 * time.Duration(paladin.Talents.InfusionOfLight)
	buff := paladin.RegisterAura(core.Aura{
		Label:    "Infusion of Light",
		ActionID: core.ActionID{SpellID: 437063},
		Duration: time.Second * 15,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			paladin.infusionReduction = reduction
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			paladin.infusionReduction = 0
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.SpellCode == SpellCode_PaladinHolyLight {
				aura.Deactivate(sim)
			}
		},
	})
	trigger := func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if result.DidCrit() && (spell.SpellCode == SpellCode_PaladinFlashOfLight || spell.SpellCode == SpellCode_PaladinHolyShock || spell.SpellCode == SpellCode_PaladinHolyShockHeal) {
			buff.Activate(sim)
		}
	}
	core.MakePermanent(paladin.RegisterAura(core.Aura{
		Label: "Infusion of Light Trigger",
		OnHealDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			trigger(sim, spell, result)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			trigger(sim, spell, result)
		},
	}))
}

// infusionOfLightReduction is what Infusion of Light takes off the next Holy Light's cast now.
func (paladin *Paladin) infusionOfLightReduction() time.Duration {
	return paladin.infusionReduction
}

// Oath of Light (spell 1328548, a talent of the hotfix): "Swear an Oath to protect target ally for
// 5 min, increasing the healing of your Holy Light on them by up to 400, your Flash of Light by up
// to 115, and your Holy Shock by up to 230. Only one Oath of Light may be sworn by the Paladin at
// any one time." 8% of base mana, instant. The three values grow with the paladin's level (6.67,
// 1.92 and 3.83 a level, so 400, 115 and 230 at 60). Like Classic's Blessing of Light, a rank of
// each heal has its own bonus heal spell (Holy Light 19982 to 1313352, Flash of Light 19993 to
// 1313346); the sim adds the bonus to the heal itself, so it crits with it and Healing Light raises
// it, gets no healing power, and is "up to": a rank Forever's downranking takes down keeps the
// same share of it (Calls for the user, docs/healing_paladin.md).
func (paladin *Paladin) registerOathOfLight() {
	if !paladin.Talents.OathOfLight {
		return
	}
	paladin.oathAuras = paladin.NewRaidPlayerAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(core.Aura{
			Label:    "Oath of Light-" + paladin.Label,
			ActionID: core.ActionID{SpellID: 1328548},
			Duration: time.Minute * 5,
		})
	})

	var current *core.Aura
	paladin.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_PaladinOathOfLight,
		ActionID:    core.ActionID{SpellID: 1328548},
		SpellSchool: core.SpellSchoolHoly,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagHelpful | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			// No Benediction: 1328548's SpellClassOptions mask is empty, so Benediction's class
			// mask (20101 e0 aura 108 misc 14) can't reach it.
			BaseCost: 0.08,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			target = paladin.healTarget(target)
			if target.IsDead() {
				return
			}
			if current != nil {
				current.Deactivate(sim)
			}
			current = paladin.oathAuras.Get(target)
			current.Activate(sim)
		},
	})
}

// oathOfLightBonus is what our Oath of Light adds to a rank of a heal learned at learnLevel, 0
// without the talent.
func (paladin *Paladin) oathOfLightBonus(spellCode int32, learnLevel int32) float64 {
	if !paladin.Talents.OathOfLight {
		return 0
	}
	perLevel := 0.0
	switch spellCode {
	case SpellCode_PaladinHolyLight:
		perLevel = 6.66666984558
	case SpellCode_PaladinFlashOfLight:
		perLevel = 1.91666996479
	case SpellCode_PaladinHolyShockHeal:
		perLevel = 3.833329916
	}
	return perLevel * float64(paladin.Level) * paladin.downrankFactor(learnLevel)
}

// HasOathOfLight is whether our Oath of Light is on the unit.
func (paladin *Paladin) HasOathOfLight(unit *core.Unit) bool {
	return paladin.oathAuras != nil && unit.Type != core.EnemyUnit && paladin.oathAuras.Get(unit).IsActive()
}

// oathOn is a heal's Oath of Light bonus when our Oath is on the target.
func (paladin *Paladin) oathOn(target *core.Unit, bonus float64) float64 {
	if bonus == 0 || !paladin.HasOathOfLight(target) {
		return 0
	}
	return bonus
}

// boundlessLightCrit is Boundless Light's crit for a Holy Shock heal at target: 15% a rank of 2 at
// an ally our Oath of Light isn't on (spell 1328550).
func (paladin *Paladin) boundlessLightCrit(target *core.Unit) float64 {
	if paladin.Talents.BoundlessLight == 0 || paladin.HasOathOfLight(target) {
		return 0
	}
	return 15 * float64(paladin.Talents.BoundlessLight) * core.SpellCritRatingPerCritChance
}
