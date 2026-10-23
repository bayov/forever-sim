package core

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

const MaxRage = 100.0
const ThreatPerRageGained = 5

// OnRageChange is called any time rage is increased.
type OnRageChange func(aura *Aura, sim *Simulation, metrics *ResourceMetrics)

type rageBar struct {
	unit *Unit

	damageDealtMultiplier  float64 // Multiplier for rage generation from damage dealt
	damageTakenMultiplier  float64 // Multiplier for rage generation from damage taken
	offHandDealtMultiplier float64 // Extra multiplier for rage generation from off-hand damage dealt
	critDealtMultiplier    float64 // Extra multiplier for rage generation from critical auto attacks
	normalizedDealt        bool    // Forever: rage from auto attacks by weapon speed, not damage

	flatDamageDealtBonusRage float64
	flatDamageTakenBonusRage float64

	startingRage float64
	currentRage  float64
	maxRage      float64

	RageRefundMetrics *ResourceMetrics
}

type RageBarOptions struct {
	StartingRage          float64
	DamageDealtMultiplier float64
	DamageTakenMultiplier float64
}

// Forever normalizes Rage from dealing damage: "Rage gained is now based on the speed of the
// weapon that dealt the damage" (Blizzard's Priest & Warrior deep dive). Blizzard gives no
// formula. A landed white hit gives k × the weapon's swing speed, with k measured by beta testers
// at level 30 (forum thread 2355684): about 3.46 a second for a one-hander and 4.5 for a
// two-hander. The off hand gets half: a beta log with advanced logging (Arlewow's
// gist a622100ad149f4fc31ddf95f9b3aab2c, linked from magey/forever-warrior issue 3, 20 September)
// shows a 1.8 sec main hand dagger giving 6.2 to 6.3 Rage a swing (3.46 × 1.8) and a 1.6 sec off
// hand dagger 2.7 to 2.8 (3.46 × 1.6 × 0.5), glancing blows included, on about 50 and 60 clean
// swings (assumptions audit, 2026-10-16). Talent multipliers (Dual Wield Specialization's 10% off
// hand Rage a point, client 1.60.1.70425 curve 111755) stay on top, and a crit adds its bonus at
// the one-hand rate (foreverWhiteHitRage). Misses, dodges and parries give nothing. The rest is
// an assumption until the beta test in orchestration/followups.md ("Rage needs a beta test"). k is a level 30 number, and flat at 60 it gives about half of Classic's Rage a swing, so
// how it grows past 30 is an assumption too (standards 21): k(L) = k30 × (1 + s × (L − 30) / 30),
// with s = 0 flat, 0.5 the midpoint (×1.5 at 60, the default) and 1 Classic-like (×2 at 60). The
// variables are for the dashboard's assumption runs (spec_assumptions.md, warr_rage_k).
var (
	ForeverNormalizedRageOneHand  = 3.46
	ForeverNormalizedRageTwoHand  = 4.5
	ForeverNormalizedRageK        = 1.0 // Multiplies both constants
	ForeverNormalizedRageOffHand  = 0.5 // Off hand factor (measured)
	ForeverNormalizedRageScaling  = 0.5 // s, how k grows from level 30 to 60
	foreverTwoHandNormalizedSpeed = 3.3
)

// ForeverNormalizedRageLevelFactor is k(level) / k(30): 1 up to level 30, then
// 1 + s × (level − 30) / 30.
func ForeverNormalizedRageLevelFactor(level int32) float64 {
	if level <= 30 {
		return 1
	}
	return 1 + ForeverNormalizedRageScaling*float64(level-30)/30
}

// ForeverNormalizedRage is the Rage a landed white hit with this weapon gives a character of this
// level under Forever, before the off hand, crit and talent multipliers.
func ForeverNormalizedRage(weapon *Weapon, level int32) float64 {
	return foreverOneHandRateRage(weapon, level) * foreverTwoHandRageFactor(weapon)
}

// foreverOneHandRateRage is the Rage a landed white hit with this weapon would give at the
// one-hand rate.
func foreverOneHandRateRage(weapon *Weapon, level int32) float64 {
	return ForeverNormalizedRageOneHand * ForeverNormalizedRageK * ForeverNormalizedRageLevelFactor(level) * weapon.SwingSpeed
}

// foreverTwoHandRageFactor is 4.5 / 3.46 for a two-hander and 1 for a one-hander.
func foreverTwoHandRageFactor(weapon *Weapon) float64 {
	if weapon.NormalizedSwingSpeed >= foreverTwoHandNormalizedSpeed {
		return ForeverNormalizedRageTwoHand / ForeverNormalizedRageOneHand
	}
	return 1
}

// foreverWhiteHitRage is the Rage a white swing gives under Forever, before the flat bonus. Only a
// swing that lands makes Rage: misses, dodges and parries make none (wowsims/forever f9f9f21883,
// six level 8 and 9 combat logs). The off hand, Dual Wield Specialization and the other talent
// multipliers scale the swing's Rage at the one-hand rate; the two-hand factor then multiplies it,
// and a crit adds the crit bonus (100%) times the one-hand-rate Rage on top, so a two-hander crit
// gives (4.5 + 3.46) × speed, not 9 × speed. The game files' Rule of Rage (DND) row 1322574
// states the 100; warrior and bear combat logs show the bonus (wowsims/forever 5ac4ffbcb6,
// 913fb51b9f).
func (unit *Unit) foreverWhiteHitRage(spell *Spell, result *SpellResult) float64 {
	if !result.Landed() {
		return 0
	}
	weapon := unit.AutoAttacks.MH()
	swingRage := unit.rageBar.damageDealtMultiplier
	if spell.ProcMask == ProcMaskMeleeOHAuto {
		weapon = unit.AutoAttacks.OH()
		swingRage *= ForeverNormalizedRageOffHand * unit.rageBar.offHandDealtMultiplier
	}
	swingRage *= foreverOneHandRateRage(weapon, unit.Level)

	generatedRage := swingRage * foreverTwoHandRageFactor(weapon)
	if result.DidCrit() {
		generatedRage += (unit.rageBar.critDealtMultiplier - 1) * swingRage
	}
	return generatedRage
}

// ForeverDamageTakenRage is the Rage a hit or a periodic tick taken gives under Forever:
// 2.5 × (1 − A) × D / c, with c the Rage conversion at the attacker's level (230.6 at 60) and A
// the armor the game assumes in place of ours, by the attacker's level (see
// foreverDamageTakenRageArmor). Our level, health and armor play no part. wowsims/forever
// ee2d69b16a, 73b317c7cf and de550b7df9 fit it on combat logs from after the 8 October build:
// warriors at 18 and 20 and bears at 30, attackers from level 16 to 60, geared and naked alike,
// periodic ticks included (5ac4ffbcb6). At 60 a 1500 hit from a level 63 boss gives 9.2 Rage;
// Classic gave 15.4.
//
// D is the attacker's hit before anything of ours touches it, times the crit or crushing blow
// multiplier. Blizzard's 8 October notes (us.forums.blizzard.com/en/wow/t/2360696, post 5) count
// "increased damage taken from combat multipliers, in particular from taking critical strikes or
// crushing blows" and ignore absorbs "as if the Warrior/Druid had no absorb active". The Warrior
// deep dive (news.blizzard.com/en-us/article/24301514) says "wearing Armor will no longer reduce
// the Rage you gain". Neither names stances, Shield Wall, talents or blocks; by the user's reading
// (2026-10-16: being more defensive doesn't penalize Rage) those don't count either (report
// assumption warr_dt_rage_mods).
func ForeverDamageTakenRage(spell *Spell, result *SpellResult) float64 {
	if !result.Landed() {
		return 0
	}
	damage := foreverRageDamage(spell, result)
	if damage <= 0 {
		return 0
	}
	attackerLevel := spell.Unit.Level
	return ForeverDamageTakenRageFactor * (1 - foreverDamageTakenRageArmor(attackerLevel)) * damage / foreverRageConversion(attackerLevel)
}

// ForeverDamageTakenRageFactor is the 2.5 of ForeverDamageTakenRage (report assumption
// warr_dt_rage_k).
const ForeverDamageTakenRageFactor = 2.5

// The armor ForeverDamageTakenRage assumes, by the attacker's level: 20% up to 42, 30% at 44, 36%
// at 46 and 38% at 50 from the logs (wowsims/forever 73b317c7cf: k 1.993 on 187 melee hits from
// attackers 16 to 42, 1.752 at 44, 1.611 at 46, 1.552 from 50 to 60), and 40% at 63, the top of
// the notes' "20% to 40% range, varying by level" (de550b7df9). Levels in between interpolate.
var foreverDamageTakenRageArmorRows = []struct {
	level     int32
	reduction float64
}{
	{42, 0.20},
	{44, 0.30},
	{46, 0.36},
	{50, 0.38},
	{63, 0.40},
}

func foreverDamageTakenRageArmor(attackerLevel int32) float64 {
	rows := foreverDamageTakenRageArmorRows
	if attackerLevel <= rows[0].level {
		return rows[0].reduction
	}
	for i := 1; i < len(rows); i++ {
		if attackerLevel <= rows[i].level {
			lo, hi := rows[i-1], rows[i]
			return lo.reduction + (hi.reduction-lo.reduction)*float64(attackerLevel-lo.level)/float64(hi.level-lo.level)
		}
	}
	return rows[len(rows)-1].reduction
}

// foreverRageConversion is the Rage conversion ForeverDamageTakenRage was fitted with: Classic's
// level 45 and up formula at every level.
func foreverRageConversion(level int32) float64 {
	l := float64(level)
	return 0.0091107836*l*l + 3.225598133*l + 4.2652911
}

// foreverRageDamage is the damage ForeverDamageTakenRage counts for a hit.
func foreverRageDamage(spell *Spell, result *SpellResult) float64 {
	if result.AttackerDamage <= 0 {
		// A result built by hand rather than by CalcDamage has no attacker damage: count what
		// landed before our armor and resists, which already carries its crit.
		if result.ResistanceMultiplier <= 0 {
			return result.Damage
		}
		return result.Damage / result.ResistanceMultiplier
	}
	damage := result.AttackerDamage
	attackTable := spell.Unit.AttackTables[result.Target.UnitIndex][spell.CastType]
	if result.DidCrit() && spell.DefenseType != DefenseTypeNone {
		damage *= spell.CritMultiplier(attackTable)
	} else if result.DidCrush() {
		damage *= 1.5
	}
	switch ForeverDamageTakenRageMods {
	case ForeverDamageTakenRageModsAll:
		damage *= spell.TargetDamageMultiplier(attackTable, false)
	case ForeverDamageTakenRageModsIncreases:
		damage *= max(1, spell.TargetDamageMultiplier(attackTable, false))
	}
	return damage
}

// Which of our damage taken modifiers (stances, Shield Wall, Recklessness, talents) change the
// Rage a hit gives (report assumption warr_dt_rage_mods).
const (
	// None: a defensive choice costs no Rage, and Recklessness or Berserker Stance add none.
	ForeverDamageTakenRageModsNone = 0
	// All: the damage after our modifiers, before armor, as master had before 2026-10-16.
	ForeverDamageTakenRageModsAll = 1
	// Increases only: Recklessness and the like add Rage, reductions take none away.
	ForeverDamageTakenRageModsIncreases = 2
)

var ForeverDamageTakenRageMods = ForeverDamageTakenRageModsNone

func GetRageConversion(attacker_level int32) float64 {
	if attacker_level == 25 {
		return 82.25 // Tested
	} else if attacker_level == 40 {
		return 140.5 // Tested
	} else if attacker_level < 45 {
		// Poor fit, but better then current formula below 45
		return 0.0215*float64(attacker_level*attacker_level) + 2.66*float64(attacker_level) + 0.89
	} else {
		// Rage conversion is adjusted according to target stats (https://web.archive.org/web/20201118213002/https://blue.mmo-champion.com/topic/18325-the-new-rage-formula-by-kalgan/)\
		// So this is probably only the base value formula and will be slightly wrong for most target
		return 0.0091107836*float64(attacker_level*attacker_level) + 3.225598133*float64(attacker_level) + 4.2652911
	}
}

func (unit *Unit) EnableRageBar(options RageBarOptions) {
	rageFromDamageTakenMetrics := unit.NewRageMetrics(ActionID{OtherID: proto.OtherAction_OtherActionDamageTaken})
	rageConversion := GetRageConversion(unit.Level)

	unit.SetCurrentPowerBar(RageBar)
	unit.RegisterAura(Aura{
		Label:    "RageBar",
		Duration: NeverExpires,
		OnInit: func(aura *Aura, sim *Simulation) {
			// Initialize resource metrics for rage gain for auto attacks here to make sure tag is correct.
			// Extra attacks change the tag from 1 to 3 for mh hits.
			mhSpell := unit.AutoAttacks.MHAuto()
			if mhSpell != nil {
				mhSpell.ResourceMetrics = unit.NewRageMetrics(mhSpell.ActionID)
			}
			ohSpell := unit.AutoAttacks.OHAuto()
			if ohSpell != nil {
				ohSpell.ResourceMetrics = unit.NewRageMetrics(ohSpell.ActionID)
			}
		},
		OnReset: func(aura *Aura, sim *Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if unit.GetCurrentPowerBar() != RageBar {
				return
			}
			if result.Outcome.Matches(OutcomeMiss) {
				return
			}
			if spell.ProcMask != ProcMaskMeleeMHAuto && spell.ProcMask != ProcMaskMeleeOHAuto {
				return
			}

			var generatedRage float64
			if unit.rageBar.normalizedDealt {
				generatedRage = unit.foreverWhiteHitRage(spell, result)
				if generatedRage <= 0 {
					return
				}
			} else {
				damage := result.Damage
				if result.Outcome.Matches(OutcomeDodge | OutcomeParry) {
					// Rage is still generated for dodges/parries, based on the damage it WOULD have done.
					damage = result.PreOutcomeDamage
				}

				generatedRage = damage * 7.5 / rageConversion
				generatedRage *= unit.rageBar.damageDealtMultiplier
				if spell.ProcMask == ProcMaskMeleeOHAuto {
					generatedRage *= unit.rageBar.offHandDealtMultiplier
				}
				if result.Outcome.Matches(OutcomeCrit) {
					generatedRage *= unit.rageBar.critDealtMultiplier
				}
			}
			generatedRage += unit.rageBar.flatDamageDealtBonusRage

			var metrics *ResourceMetrics
			if spell.Cost != nil {
				metrics = spell.Cost.SpellCostFunctions.(*RageCost).ResourceMetrics
			} else {
				// Seems like only auto attacks are using this. See OnInit handler of this aura.
				if spell.ResourceMetrics == nil {
					panic(fmt.Sprintf("Spell ResourceMetrics are nil for spell %v", spell.ActionID))
				}
				metrics = spell.ResourceMetrics
			}
			unit.AddRage(sim, generatedRage, metrics)
		},
		OnSpellHitTaken: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if unit.GetCurrentPowerBar() != RageBar {
				return
			}
			var generatedRage float64
			if sim.IsForever() {
				generatedRage = ForeverDamageTakenRage(spell, result)
			} else {
				generatedRage = result.Damage * 2.5 / GetRageConversion(spell.Unit.Level)
			}
			generatedRage *= unit.rageBar.damageTakenMultiplier
			generatedRage += unit.rageBar.flatDamageTakenBonusRage
			unit.AddRage(sim, generatedRage, rageFromDamageTakenMetrics)
		},
		// Under Forever a periodic tick taken gives Rage like a direct hit (wowsims/forever
		// 5ac4ffbcb6); Classic's ticks give none.
		OnPeriodicDamageTaken: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if unit.GetCurrentPowerBar() != RageBar || !sim.IsForever() {
				return
			}
			generatedRage := ForeverDamageTakenRage(spell, result) * unit.rageBar.damageTakenMultiplier
			if generatedRage > 0 {
				unit.AddRage(sim, generatedRage, rageFromDamageTakenMetrics)
			}
		},
	})

	// Not a real spell, just holds metrics from rage gain threat.
	unit.RegisterSpell(SpellConfig{
		ActionID: ActionID{OtherID: proto.OtherAction_OtherActionRageGain},
	})

	unit.rageBar = rageBar{
		unit:                   unit,
		damageDealtMultiplier:  options.DamageDealtMultiplier,
		damageTakenMultiplier:  options.DamageTakenMultiplier,
		offHandDealtMultiplier: 1,
		critDealtMultiplier:    1,
		startingRage:           max(0, min(options.StartingRage, MaxRage)),
		maxRage:                MaxRage,
		RageRefundMetrics:      unit.NewRageMetrics(ActionID{OtherID: proto.OtherAction_OtherActionRefund}),
	}
}

func (unit *Unit) HasRageBar() bool {
	return unit.rageBar.unit != nil
}

func (unit *Unit) AddDamageDealtRageMultiplier(multi float64) {
	unit.rageBar.damageDealtMultiplier *= multi
}

func (unit *Unit) AddOffHandDealtRageMultiplier(multi float64) {
	unit.rageBar.offHandDealtMultiplier *= multi
}

// Forever gives Rage users more Rage from a critical auto attack (Kaivax's 2 October post,
// https://us.forums.blizzard.com/en/wow/t/2369360; an assumption until the beta test in
// orchestration/followups.md, "Rage needs a beta test") (the 1 October development
// notes: 100% more for warriors).
func (unit *Unit) AddCritDealtRageMultiplier(multi float64) {
	unit.rageBar.critDealtMultiplier *= multi
}

// EnableForeverNormalizedRage switches Rage from dealing damage to Forever's normalized Rage (see
// ForeverNormalizedRage). Rage from taking damage follows Forever's rule for any Rage bar under
// the Forever ruleset (ForeverDamageTakenRage).
func (unit *Unit) EnableForeverNormalizedRage() {
	if unit.Env.IsForever() {
		unit.rageBar.normalizedDealt = true
	}
}

func (unit *Unit) AddMaxRage(amount float64) {
	unit.rageBar.maxRage += amount
}

func (unit *Unit) AddDamageTakenRageMultiplier(multi float64) {
	unit.rageBar.damageTakenMultiplier *= multi
}

func (unit *Unit) AddDamageDealtRageBonus(bonus float64) {
	unit.rageBar.flatDamageDealtBonusRage += bonus
}

func (unit *Unit) AddDamageTakenRageBonus(bonus float64) {
	unit.rageBar.flatDamageTakenBonusRage += bonus
}

func (rb *rageBar) CurrentRage() float64 {
	return rb.currentRage
}

func (rb *rageBar) AddRage(sim *Simulation, amount float64, metrics *ResourceMetrics) {
	if amount < 0 {
		panic("Trying to add negative rage!")
	}

	newRage := min(rb.currentRage+amount, rb.maxRage)
	metrics.AddEvent(amount, newRage-rb.currentRage)

	if sim.Log != nil {
		rb.unit.Log(sim, "Gained %0.3f rage from %s (%0.3f --> %0.3f).", amount, metrics.ActionID, rb.currentRage, newRage)
	}

	rb.currentRage = newRage
	if !sim.Options.Interactive {
		rb.unit.Rotation.DoNextAction(sim)
	}
	StartDelayedAction(sim, DelayedActionOptions{
		DoAt: sim.CurrentTime + time.Millisecond*1,
		OnAction: func(sim *Simulation) {
			rb.unit.OnRageChange(sim, metrics)
		},
	})

}

func (rb *rageBar) SpendRage(sim *Simulation, amount float64, metrics *ResourceMetrics) {
	if amount < 0 {
		panic("Trying to spend negative rage!")
	}

	newRage := rb.currentRage - amount
	metrics.AddEvent(-amount, -amount)

	if sim.Log != nil {
		rb.unit.Log(sim, "Spent %0.3f rage from %s (%0.3f --> %0.3f).", amount, metrics.ActionID, rb.currentRage, newRage)
	}

	rb.currentRage = newRage

	rb.unit.OnRageChange(sim, metrics)
}

func (rb *rageBar) reset(_ *Simulation) {
	if rb.unit == nil {
		return
	}

	rb.currentRage = rb.startingRage
}

func (rb *rageBar) doneIteration() {
	if rb.unit == nil {
		return
	}

	rageGainSpell := rb.unit.GetSpell(ActionID{OtherID: proto.OtherAction_OtherActionRageGain})

	for _, resourceMetrics := range rb.unit.Metrics.resources {
		if resourceMetrics.Type != proto.ResourceType_ResourceTypeRage {
			continue
		}
		if resourceMetrics.ActionID.SameActionIgnoreTag(ActionID{OtherID: proto.OtherAction_OtherActionDamageTaken}) {
			continue
		}
		if resourceMetrics.ActionID.SameActionIgnoreTag(ActionID{OtherID: proto.OtherAction_OtherActionRefund}) {
			continue
		}
		if resourceMetrics.ActualGainForCurrentIteration() <= 0 {
			continue
		}

		// Need to exclude rage gained from white hits. Rather than have a manual list of all IDs that would
		// apply here (autos, WF attack, sword spec procs, etc), just check if the effect caused any damage.
		sourceSpell := rb.unit.GetSpell(resourceMetrics.ActionID)
		if sourceSpell != nil && sourceSpell.SpellMetrics[0].TotalDamage > 0 {
			continue
		}

		rageGainSpell.SpellMetrics[0].Casts += resourceMetrics.EventsForCurrentIteration()
		rageGainSpell.ApplyAOEThreatIgnoreMultipliers(resourceMetrics.ActualGainForCurrentIteration() * ThreatPerRageGained)
	}
}

type RageCostOptions struct {
	Cost float64

	Refund        float64
	RefundMetrics *ResourceMetrics // Optional, will default to unit.RageRefundMetrics if not supplied.
}
type RageCost struct {
	Refund          float64 // The share of the cost a miss refunds.
	refundBase      float64 // The base cost Classic refunds that share of.
	RefundMetrics   *ResourceMetrics
	ResourceMetrics *ResourceMetrics
}

func newRageCost(spell *Spell, options RageCostOptions) *SpellCost {
	if options.Refund > 0 && options.RefundMetrics == nil {
		options.RefundMetrics = spell.Unit.RageRefundMetrics
	}

	return &SpellCost{
		spell:      spell,
		BaseCost:   options.Cost,
		Multiplier: 100,
		Factor:     1,
		SpellCostFunctions: &RageCost{
			Refund:          options.Refund,
			refundBase:      options.Cost,
			RefundMetrics:   options.RefundMetrics,
			ResourceMetrics: spell.Unit.NewRageMetrics(spell.ActionID),
		},
	}
}

func (rc *RageCost) CostType() CostType {
	return CostTypeRage
}

func (rc *RageCost) MeetsRequirement(_ *Simulation, spell *Spell) bool {
	spell.CurCast.Cost = spell.Cost.GetCurrentCost()
	return spell.Unit.CurrentRage() >= spell.CurCast.Cost
}
func (rc *RageCost) CostFailureReason(sim *Simulation, spell *Spell) string {
	return fmt.Sprintf("not enough rage (Current Rage = %0.03f, Rage Cost = %0.03f)", spell.Unit.CurrentRage(), spell.CurCast.Cost)
}
func (rc *RageCost) SpendCost(sim *Simulation, spell *Spell) {
	if spell.CurCast.Cost > 0 {
		spell.Unit.SpendRage(sim, spell.CurCast.Cost, rc.ResourceMetrics)
	}
}

// Under Forever a miss refunds its share of the rage the cast paid, after talents and other cost
// modifiers, as energy and focus do (wowsims/forever: Refund * CurCast.Cost, only when above 0).
func (rc *RageCost) IssueRefund(sim *Simulation, spell *Spell) {
	refund := rc.Refund * rc.refundBase
	if sim.IsForever() {
		refund = rc.Refund * spell.CurCast.Cost
	}
	if refund > 0 {
		spell.Unit.AddRage(sim, refund, rc.RefundMetrics)
	}
}

func (unit *Unit) MultiplyMaxRage(multiplier float64) {
	unit.rageBar.maxRage *= multiplier
}
