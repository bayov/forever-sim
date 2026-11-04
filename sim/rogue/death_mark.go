package rogue

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Death Mark, the 31-point Subtlety talent of Blizzard's November 2 development notes: "Finishing move
// that marks the target, storing a portion of the damage dealt by your attacks per Combo Point.
// Garrote and Rupture damage is stored at twice the rate. After 6 sec, the stored damage is dealt
// up to a maximum increased by Attack Power and Critical Strike chance." 25 Energy, 10 sec cooldown.
//
// The spell is in the client since 70425 (1316709 at 40, 1325756 at 48, 1325757 at 56, 1325758 at
// 60; client 1.60.1.70753): 25 Energy and the combo points, a 10 sec cooldown, a 1 sec GCD, a melee
// defense, physical. Its aura is a 6 sec periodic damage effect (aura 3, period 6000) plus a dummy
// that procs on any damage the target takes (ProcTypeMask 0x800AAAA8). SpellDescriptionVariables
// 1021: $ratio 12 (12% stored a combo point) and $maxAP 0.04; rank 4's cap is 156 + 245 a combo
// point + 4% of attack power a combo point, rank 1's 63 + 102 a point, "before critical strike".
//
// Modelled: the rogue's damage to the marked target (direct, periodic, auto attacks and poisons,
// Rupture and Garrote twice) is stored at 12% a combo point; after 6 sec the store, up to the cap
// (attack power read at the cast), is dealt as one physical hit that can crit and that no armor or
// modifier touches again, since the stored damage already went through them. Coup de Grace raises
// the storage rate and the damage by 7/13/20%. Assumed (notes.md, Rogue): the cap applies before
// the crit roll, ranks 2 and 3 use rank 1's cap (their tooltips give none), and the mark's own hit
// isn't stored.
func (rogue *Rogue) registerDeathMark() {
	if !rogue.Talents.DeathMark || rogue.Level < 40 {
		return
	}

	spellID := rankSpellID(rogue.Level, map[int32]int32{40: 1316709, 48: 1325756, 56: 1325757, 60: 1325758})
	capBase, capPerPoint := 63.0, 102.0
	if rogue.Level >= 60 {
		capBase, capPerPoint = 156, 245
	}
	const storedPerPoint = 0.12
	const capAPPerPoint = 0.04
	coupDeGrace := rogue.coupDeGraceMultiplier()

	var storeRate, stored, damageCap float64

	detonate := rogue.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellID, Tag: 1},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete | core.SpellFlagIgnoreResists | core.SpellFlagIgnoreModifiers | core.SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := min(stored, damageCap) * coupDeGrace
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeSpecialCritOnly)
		},
	})

	rogue.DeathMarkAuras = rogue.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			Label:    "Death Mark-" + strconv.Itoa(int(rogue.Index)),
			ActionID: core.ActionID{SpellID: spellID},
			Duration: time.Second * 6,
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				// Only when the 6 sec run out, not when the fight ends first.
				if aura.ExpiresAt() <= sim.CurrentTime {
					detonate.Cast(sim, aura.Unit)
				}
				stored = 0
			},
		})
	})

	store := func(spell *core.Spell, result *core.SpellResult) {
		if result.Damage <= 0 || spell == detonate || spell == rogue.DeathMark {
			return
		}
		aura := rogue.DeathMarkAuras.Get(result.Target)
		if !aura.IsActive() {
			return
		}
		rate := storeRate
		if spell.SpellCode == SpellCode_RogueRupture || spell.SpellCode == SpellCode_RogueGarrote {
			rate *= 2
		}
		stored += result.Damage * rate
	}

	core.MakePermanent(rogue.RegisterAura(core.Aura{
		Label: "Death Mark Storage",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			store(spell, result)
		},
		OnPeriodicDamageDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			store(spell, result)
		},
	}))

	rogue.DeathMark = rogue.RegisterSpell(core.SpellConfig{
		SpellCode:    SpellCode_RogueDeathMark,
		ActionID:     core.ActionID{SpellID: spellID},
		SpellSchool:  core.SpellSchoolPhysical,
		DefenseType:  core.DefenseTypeMelee,
		ProcMask:     core.ProcMaskMeleeMHSpecial,
		Flags:        rogue.finisherFlags(),
		MetricSplits: 6,

		EnergyCost: core.EnergyCostOptions{
			Cost:   25,
			Refund: rogue.finisherMissRefund(),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: time.Second * 10,
			},
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				spell.SetMetricsSplit(spell.Unit.ComboPoints())
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return rogue.ComboPoints() > 0
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
			if result.Landed() {
				points := float64(rogue.ComboPoints())
				aura := rogue.DeathMarkAuras.Get(target) // the 10 sec cooldown outlasts the 6 sec mark
				storeRate = storedPerPoint * points * coupDeGrace
				damageCap = capBase + capPerPoint*points + capAPPerPoint*points*spell.MeleeAttackPower(target)
				stored = 0
				aura.Activate(sim)
				rogue.SpendComboPoints(sim, spell)
			} else {
				spell.IssueRefund(sim)
			}
			spell.DealOutcome(sim, result)
		},
	})
	rogue.Finishers = append(rogue.Finishers, rogue.DeathMark)
}
