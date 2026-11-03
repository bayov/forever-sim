package priest

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The healing talents (docs/healing_priest.md, the audit). Values are the talents' client curves
// (1.60.1.70753, TraitDefinitionEffectPoints) and the spells each touches are its class masks
// (SpellEffect), as for the damage talents in talents.go. Forever's percent spell modifiers
// multiply each other (as the cost ones do, reduceCost).

// healCodes are the priest's heals that aren't a shield.
var healCodes = []int32{SpellCode_PriestLesserHeal, SpellCode_PriestHeal, SpellCode_PriestGreaterHeal, SpellCode_PriestFlashHeal,
	SpellCode_PriestBindingHeal, SpellCode_PriestPrayerOfHealing, SpellCode_PriestHolyNova, SpellCode_PriestRenew,
	SpellCode_PriestPrayerOfMending, SpellCode_PriestPenance}

// isHeal is whether the spell is one of the priest's heals (Penance's healing bolts, not its
// damage ones).
func isHeal(spell *core.Spell, codes ...int32) bool {
	return spellIn(spell, codes...) && spell.ProcMask.Matches(core.ProcMaskSpellHealing)
}

func (priest *Priest) applyHealingTalents() {
	// Spiritual Healing: 3, 7 and 10% more healing on every heal but Power Word: Shield (14898:
	// effect 0 the direct heals and Penance, effect 1 Renew's ticks).
	if priest.Talents.SpiritualHealing > 0 {
		multiplier := 1 + []float64{0, 0.03, 0.07, 0.10}[priest.Talents.SpiritualHealing]
		priest.OnSpellRegistered(func(spell *core.Spell) {
			if isHeal(spell, healCodes...) {
				spell.DamageMultiplier *= multiplier
			}
		})
	}

	// Improved Renew: 5, 10 and 15% more from Renew.
	if priest.Talents.ImprovedRenew > 0 {
		multiplier := 1 + 0.05*float64(priest.Talents.ImprovedRenew)
		priest.OnSpellRegistered(func(spell *core.Spell) {
			if spellIn(spell, SpellCode_PriestRenew) {
				spell.DamageMultiplier *= multiplier
			}
		})
	}

	// Improved Healing: 5, 10 and 15% off Lesser Heal, Heal, Greater Heal, Penance and Prayer of
	// Mending (14912).
	if priest.Talents.ImprovedHealing > 0 {
		percent := 5 * priest.Talents.ImprovedHealing
		priest.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Cost != nil && spellIn(spell, SpellCode_PriestLesserHeal, SpellCode_PriestHeal, SpellCode_PriestGreaterHeal,
				SpellCode_PriestPrayerOfMending) || spell.Cost != nil && spell.Flags.Matches(SpellFlagPriest) && isPenanceCast(spell) {
				priest.reduceCost(spell, percent)
			}
		})
	}

	// Soul Warding's cost part: 15% off Power Word: Shield (its cooldown part is in
	// registerPowerWordShieldSpell).
	if priest.Talents.SoulWarding {
		priest.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Cost != nil && spellIn(spell, SpellCode_PriestPowerWordShield) {
				priest.reduceCost(spell, 15)
			}
		})
	}

	// Spiritual Guidance's healing part: 5% of Spirit a rank as healing (14901 effect 1, aura 175;
	// the damage part is in ApplyTalents).
	if priest.Talents.SpiritualGuidance > 0 {
		priest.AddStatDependency(stats.Spirit, stats.HealingPower, 0.05*float64(priest.Talents.SpiritualGuidance))
	}

	priest.applyBlessedRecovery()
}

// isPenanceCast is whether the spell is one of Penance's channels (its cost is the channel's).
func isPenanceCast(spell *core.Spell) bool {
	for _, id := range PenanceSpellId[1:] {
		if spell.ActionID.SpellID == id {
			return true
		}
	}
	return false
}

// registerHealingTalentAuras makes the buffs and shields the healing talents put on others, once
// the raid exists.
func (priest *Priest) registerHealingTalentAuras() {
	// Inspiration: a heal's crit (not a tick's) gives its target 8, 17 or 25% more armor for 15
	// sec (14892, 14893). One buff a player, whichever priest's crit gave it.
	if priest.Talents.Inspiration > 0 {
		armor := 1 + []float64{0, 0.08, 0.17, 0.25}[priest.Talents.Inspiration]
		priest.InspirationAuras = priest.NewRaidAuraArray(func(unit *core.Unit) *core.Aura {
			return unit.GetOrRegisterAura(core.Aura{
				Label:    "Inspiration",
				ActionID: core.ActionID{SpellID: 14893},
				Duration: time.Second * 15,
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					aura.Unit.PseudoStats.ArmorMultiplier *= armor
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					aura.Unit.PseudoStats.ArmorMultiplier /= armor
				},
			})
		})
	}

	// Divine Aegis: a heal's crit (not a tick's) shields its target for 5, 10 or 15% of the heal
	// for 12 sec, up to 20 a level (431622 effect 1, "Shields up to ${$s2*$pl}"; 431624). A new
	// crit adds to a shield still up, to that limit (the deep dive said "does not stack" at 70170,
	// before 70425 gave it the limit; docs/healing_priest.md, Calls for the user).
	if priest.Talents.DivineAegis > 0 {
		priest.DivineAegisAbsorbs = priest.NewAbsorbArray(core.Aura{
			Label:    "Divine Aegis",
			ActionID: core.ActionID{SpellID: 431624},
			Duration: time.Second * 12,
		})
		priest.divineAegisSpell = priest.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 431624},
			SpellSchool: core.SpellSchoolHoly,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       SpellFlagPriest | core.SpellFlagHelpful | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,
		})
	}
}

// DivineAegisShare and DivineAegisPerLevel are Divine Aegis's share of a crit heal a rank and its
// limit a level.
const (
	DivineAegisShare    = 0.05
	DivineAegisPerLevel = 20
)

// afterDirectHeal is what follows one of our heals that isn't a tick.
func (priest *Priest) afterDirectHeal(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	target := result.Target
	if target.IsDead() {
		return
	}
	if result.DidCrit() {
		if priest.InspirationAuras != nil {
			if aura := priest.InspirationAuras.Get(target); aura != nil {
				aura.Activate(sim)
			}
		}
		if priest.DivineAegisAbsorbs != nil {
			amount := DivineAegisShare * float64(priest.Talents.DivineAegis) * result.Damage
			priest.DivineAegisAbsorbs.Get(target).Add(sim, priest.divineAegisSpell, amount, DivineAegisPerLevel*float64(priest.Level))
		}
	}
	// Renewed Hope: each of its heals takes 1 sec a rank off the target's Weakened Soul.
	if priest.Talents.RenewedHope > 0 && renewedHopeSpell(spell) && priest.WeakenedSouls != nil {
		if ws := priest.WeakenedSouls.Get(target); ws.IsActive() {
			expires := ws.ExpiresAt() - time.Second*time.Duration(priest.Talents.RenewedHope)
			if expires <= sim.CurrentTime {
				ws.Deactivate(sim)
			} else {
				ws.UpdateExpires(sim, expires)
			}
		}
	}
	// Holy Binding: a Flash Heal on another heals us for half as much.
	if priest.holyBinding != nil && spell.SpellCode == SpellCode_PriestFlashHeal && target != &priest.Unit {
		priest.holyBinding.CalcAndDealHealing(sim, &priest.Unit, HolyBindingShare*result.Damage, priest.holyBinding.OutcomeHealing)
	}
}

// Renewed Hope's heals: Flash Heal, Lesser Heal, Heal, Greater Heal and Penance (425280's
// tooltip in 70753; the deep dive also named Binding Heal, which the newer tooltip leaves out).
func renewedHopeSpell(spell *core.Spell) bool {
	return isHeal(spell, SpellCode_PriestFlashHeal, SpellCode_PriestLesserHeal, SpellCode_PriestHeal, SpellCode_PriestGreaterHeal,
		SpellCode_PriestPenance)
}

// renewedHopeCrit is Renewed Hope's 2% crit a rank for its heals on a target with Weakened Soul.
func (priest *Priest) renewedHopeCrit(spell *core.Spell, target *core.Unit) float64 {
	if priest.Talents.RenewedHope == 0 || priest.WeakenedSouls == nil || !renewedHopeSpell(spell) || !priest.WeakenedSouls.Get(target).IsActive() {
		return 0
	}
	return 2 * float64(priest.Talents.RenewedHope) * core.CritRatingPerCritChance
}

// HolyBindingShare is Holy Binding's share of a Flash Heal on another that heals us (1328569).
const HolyBindingShare = 0.5

// registerHolyBinding is Holy Binding's self heal (1328570), which makes no threat and takes no
// modifiers: it is half of what the Flash Heal healed.
func (priest *Priest) registerHolyBinding() {
	if !priest.Talents.HolyBinding {
		return
	}
	priest.holyBinding = priest.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 1328570},
		SpellSchool: core.SpellSchoolHoly,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       core.SpellFlagHelpful | core.SpellFlagIgnoreModifiers | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		ThreatMultiplier: 0,
	})
}

// litanyOfLight is Litany of Light: casting a heal whose name isn't the last heal's gives back 5
// or 10% of its base cost (1317006; the deep dive: "if the name of the spell is different than the
// previous healing spell you cast", so ranks of one heal are the same spell). The fight's first
// heal counts as different.
func (priest *Priest) litanyOfLight(sim *core.Simulation, spell *core.Spell, baseCost float64) {
	if priest.Talents.LitanyOfLight == 0 {
		return
	}
	if spell.SpellCode != priest.lastHealCode {
		priest.AddMana(sim, 0.05*float64(priest.Talents.LitanyOfLight)*baseCost, priest.litanyMetrics)
	}
	priest.lastHealCode = spell.SpellCode
}

// Blessed Recovery: a melee or ranged crit on us, or a melee or ranged hit for more than 30% of
// our health, heals 8, 17 or 25% of it over 6 sec, 3 ticks (27811, 27813; its proc flags are melee
// and ranged hits taken). A new one carries over what the old one had left.
func (priest *Priest) applyBlessedRecovery() {
	if priest.Talents.BlessedRecovery == 0 {
		return
	}
	share := []float64{0, 0.08, 0.17, 0.25}[priest.Talents.BlessedRecovery]
	const ticks = 3
	var perTick float64

	hot := priest.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 27813},
		SpellSchool: core.SpellSchoolHoly,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       core.SpellFlagHelpful | core.SpellFlagIgnoreModifiers | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Hot: core.DotConfig{
			SelfOnly: true,
			Aura: core.Aura{
				Label: "Blessed Recovery",
			},
			NumberOfTicks: ticks,
			TickLength:    time.Second * 2,
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				if target.IsDead() {
					return
				}
				dot.Spell.CalcAndDealPeriodicHealing(sim, target, perTick, dot.Spell.OutcomeHealing)
			},
		},
	})

	core.MakePermanent(priest.RegisterAura(core.Aura{
		Label: "Blessed Recovery Talent",
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Damage <= 0 || priest.IsDead() || !spell.ProcMask.Matches(core.ProcMaskMeleeOrRanged) {
				return
			}
			if !result.DidCrit() && result.Damage <= 0.3*priest.MaxHealth() {
				return
			}
			dot := hot.SelfHot()
			left := 0.0
			if dot.IsActive() {
				left = perTick * float64(dot.MaxTicksRemaining())
			}
			perTick = (left + share*result.Damage) / ticks
			dot.Apply(sim)
		},
	}))
}
