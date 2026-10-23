package core

import (
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// ReplaceMHSwing is called right before a main hand auto attack fires.
// It must never return nil, but either a replacement spell or the passed in regular mhSwingSpell.
// This allows for abilities that convert a white attack into a yellow attack.
type ReplaceMHSwing func(sim *Simulation, mhSwingSpell *Spell) *Spell

// Represents a generic weapon. Pets / unarmed / various other cases don't use
// actual weapon items so this is an abstraction of a Weapon.
type Weapon struct {
	BaseDamageMin        float64
	BaseDamageMax        float64
	AttackPowerPerDPS    float64
	SwingSpeed           float64
	NormalizedSwingSpeed float64
	SpellSchool          SpellSchool
}

func (weapon *Weapon) DPS() float64 {
	if weapon.SwingSpeed == 0 {
		return 0
	}
	return (weapon.BaseDamageMin + weapon.BaseDamageMax) / 2.0 / weapon.SwingSpeed
}

func newWeaponFromUnarmed() Weapon {
	// These numbers are probably wrong but nobody cares.
	return Weapon{
		BaseDamageMin:        0,
		BaseDamageMax:        0,
		SwingSpeed:           1,
		NormalizedSwingSpeed: 1,
		AttackPowerPerDPS:    DefaultAttackPowerPerDPS,
	}
}

func newWeaponFromItem(item *Item, bonusDps float64) Weapon {
	normalizedWeaponSpeed := 2.4
	if item.WeaponType == proto.WeaponType_WeaponTypeDagger {
		normalizedWeaponSpeed = 1.7
	} else if item.HandType == proto.HandType_HandTypeTwoHand {
		normalizedWeaponSpeed = 3.3
	} else if item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeUnknown {
		normalizedWeaponSpeed = 2.8
	}

	return Weapon{
		BaseDamageMin:        item.WeaponDamageMin + bonusDps*item.SwingSpeed,
		BaseDamageMax:        item.WeaponDamageMax + bonusDps*item.SwingSpeed,
		SwingSpeed:           item.SwingSpeed,
		NormalizedSwingSpeed: normalizedWeaponSpeed,
		AttackPowerPerDPS:    DefaultAttackPowerPerDPS,
	}
}

// Returns weapon stats using the main hand equipped weapon.
func (character *Character) WeaponFromMainHand() Weapon {
	if weapon := character.GetMHWeapon(); weapon != nil {
		return newWeaponFromItem(weapon, character.PseudoStats.BonusMHDps)
	} else {
		return newWeaponFromUnarmed()
	}
}

// Returns weapon stats using the off-hand equipped weapon.
func (character *Character) WeaponFromOffHand() Weapon {
	if weapon := character.GetOHWeapon(); weapon != nil {
		return newWeaponFromItem(weapon, character.PseudoStats.BonusOHDps)
	} else {
		return Weapon{}
	}
}

// Returns weapon stats using the ranged equipped weapon.
func (character *Character) WeaponFromRanged() Weapon {
	if weapon := character.GetRangedWeapon(); weapon != nil {
		return newWeaponFromItem(weapon, character.PseudoStats.BonusRangedDps)
	} else {
		return Weapon{}
	}
}

func (weapon *Weapon) GetSpellSchool() SpellSchool {
	if weapon.SpellSchool == SpellSchoolNone {
		return SpellSchoolPhysical
	} else {
		return weapon.SpellSchool
	}
}

func (weapon *Weapon) EnemyWeaponDamage(sim *Simulation, attackPower float64, damageSpread float64) float64 {
	// Maximum damage range is 133% of minimum damage; AP contribution is % of minimum damage roll.
	// Patchwerk follows special damage range rules.
	// TODO: Scrape more logs to determine these values more accurately. AP defined in constants.go

	rand := 1 + damageSpread*sim.RandomFloat("Enemy Weapon Damage")

	return weapon.BaseDamageMin * (rand + attackPower*EnemyAutoAttackAPCoefficient)
}

func (weapon *Weapon) BaseDamage(sim *Simulation) float64 {
	return weapon.BaseDamageMin + (weapon.BaseDamageMax-weapon.BaseDamageMin)*sim.RandomFloat("Weapon Base Damage")
}

func (weapon *Weapon) AverageDamage() float64 {
	return (weapon.BaseDamageMin + weapon.BaseDamageMax) / 2
}

func (weapon *Weapon) CalculateWeaponDamage(sim *Simulation, attackPower float64) float64 {
	return weapon.BaseDamage(sim) + (weapon.SwingSpeed*attackPower)/weapon.AttackPowerPerDPS
}

func (weapon *Weapon) CalculateAverageWeaponDamage(attackPower float64) float64 {
	return weapon.AverageDamage() + (weapon.SwingSpeed*attackPower)/weapon.AttackPowerPerDPS
}

func (weapon *Weapon) CalculateNormalizedWeaponDamage(sim *Simulation, attackPower float64) float64 {
	return weapon.BaseDamage(sim) + (weapon.NormalizedSwingSpeed*attackPower)/weapon.AttackPowerPerDPS
}

func (unit *Unit) MHWeaponDamage(sim *Simulation, attackPower float64) float64 {
	return unit.AutoAttacks.mh.CalculateWeaponDamage(sim, attackPower)
}
func (unit *Unit) MHNormalizedWeaponDamage(sim *Simulation, attackPower float64) float64 {
	return unit.AutoAttacks.mh.CalculateNormalizedWeaponDamage(sim, attackPower)
}

func (unit *Unit) OHWeaponDamage(sim *Simulation, attackPower float64) float64 {
	return 0.5 * unit.AutoAttacks.oh.CalculateWeaponDamage(sim, attackPower)
}
func (unit *Unit) OHNormalizedWeaponDamage(sim *Simulation, attackPower float64) float64 {
	return 0.5 * unit.AutoAttacks.oh.CalculateNormalizedWeaponDamage(sim, attackPower)
}

func (unit *Unit) RangedWeaponDamage(sim *Simulation, attackPower float64) float64 {
	return unit.AutoAttacks.ranged.CalculateWeaponDamage(sim, attackPower)
}

// Returns whether this hit effect is associated with the main-hand weapon.
func (spell *Spell) IsMH() bool {
	return spell.ProcMask.Matches(ProcMaskMeleeMH)
}

// Returns whether this hit effect is associated with the off-hand weapon.
func (spell *Spell) IsOH() bool {
	return spell.ProcMask.Matches(ProcMaskMeleeOH)
}

// Returns whether this hit effect is associated with either melee weapon.
func (spell *Spell) IsMelee() bool {
	return spell.ProcMask.Matches(ProcMaskMelee)
}

func (aa *AutoAttacks) MH() *Weapon {
	return aa.mh.getWeapon()
}

func (aa *AutoAttacks) SetMH(weapon Weapon) {
	aa.mh.setWeapon(weapon)

	if aa.mh.extraAttacksAura == nil {
		aa.mh.extraAttacksAura = aa.mh.unit.GetAuraByID(ActionID{SpellID: 21919})
	}
}

func (aa *AutoAttacks) OH() *Weapon {
	return aa.oh.getWeapon()
}

func (aa *AutoAttacks) SetOH(weapon Weapon) {
	aa.oh.setWeapon(weapon)
}

func (aa *AutoAttacks) Ranged() *Weapon {
	return aa.ranged.getWeapon()
}

func (aa *AutoAttacks) SetRanged(weapon Weapon) {
	aa.ranged.setWeapon(weapon)
}

func (aa *AutoAttacks) MHAuto() *Spell {
	return aa.mh.spell
}

func (aa *AutoAttacks) OHAuto() *Spell {
	return aa.oh.spell
}

func (aa *AutoAttacks) RangedAuto() *Spell {
	return aa.ranged.spell
}

func (aa *AutoAttacks) MainhandSwingAt() time.Duration {
	return aa.mh.swingAt
}

func (aa *AutoAttacks) OffhandSwingAt() time.Duration {
	return aa.oh.swingAt
}

func (aa *AutoAttacks) SetOffhandSwingAt(offhandSwingAt time.Duration) {
	aa.oh.swingAt = offhandSwingAt
}

func (aa *AutoAttacks) SetReplaceMHSwing(replaceSwing ReplaceMHSwing) {
	aa.mh.replaceSwing = replaceSwing
}

func (aa *AutoAttacks) MHConfig() *SpellConfig {
	return &aa.mh.config
}

func (aa *AutoAttacks) OHConfig() *SpellConfig {
	return &aa.oh.config
}

func (aa *AutoAttacks) RangedConfig() *SpellConfig {
	return &aa.ranged.config
}

const (
	tagMainhand    = 1
	tagOffhand     = 2
	tagExtraAttack = 3
)

type WeaponAttack struct {
	Weapon

	agent Agent
	unit  *Unit

	config SpellConfig
	spell  *Spell

	replaceSwing ReplaceMHSwing

	swingAt             time.Duration
	lastSwingAt         time.Duration
	extraAttacks        int32 // extra attacks that happen right away
	extraAttacksStored  int32 // extra attack that happen on next auto (e.g reckoning)
	extraAttacksPending int32 // extraAttacks prior to previous ones resolving for spell metrics
	extraAttacksAura    *Aura

	curSwingSpeed    float64
	curSwingDuration time.Duration
	heldByCast       bool // The swing came due during a cast that resets it and waits for the cast to end.
}

func (wa *WeaponAttack) getWeapon() *Weapon {
	return &wa.Weapon
}

func (wa *WeaponAttack) setWeapon(weapon Weapon) {
	wa.Weapon = weapon
	wa.updateSwingDuration(wa.curSwingSpeed)
}

// inlineable stub for swing
func (wa *WeaponAttack) trySwing(sim *Simulation) time.Duration {
	if sim.CurrentTime < wa.swingAt {
		return wa.swingAt
	}
	return wa.swing(sim)
}

func (wa *WeaponAttack) castExtraAttacksStored(sim *Simulation) {
	if wa.extraAttacksAura != nil {
		wa.extraAttacksStored = wa.extraAttacksAura.GetStacks()
		wa.extraAttacksAura.SetStacks(sim, 0)
	}

	wa.castExtraAttacks(sim, wa.extraAttacksStored, 0)
	wa.extraAttacksStored = 0

}

func (wa *WeaponAttack) castExtraAttacksTriggered(sim *Simulation, moreAttacks bool) {
	if moreAttacks {
		wa.castExtraAttacks(sim, wa.extraAttacks, 0)
	} else {
		wa.castExtraAttacks(sim, wa.extraAttacks, 1)
	}
	wa.extraAttacks = 0
}

func (wa *WeaponAttack) castExtraAttacks(sim *Simulation, numExtraAttacks int32, startIndex int32) bool {
	if numExtraAttacks > 0 {
		// if startIndex==1, Ignore the first extra attack, that was used to speed up next attack
		wa.spell.SetMetricsSplit(1)

		for i := int32(startIndex); i < numExtraAttacks; i++ {
			// use original attacks for subsequent extra Attacks
			wa.spell.Cast(sim, wa.unit.CurrentTarget)
		}

		return true
	}
	return false

}

func (wa *WeaponAttack) swing(sim *Simulation) time.Duration {
	// A cast ending now lands first, and one that resets the swing moves this one.
	wa.unit.completeDueHardcast(sim)
	if sim.CurrentTime < wa.swingAt {
		return wa.swingAt
	}

	// Under Forever a melee swing waits out a cast that resets it. The completed cast sets the
	// real swing time; this one only has to land past the cast's end, wherever pushback moves it.
	// An extra attack doesn't wait: it swings right after the hit that granted it, in the same
	// instant, as its own cast (wowsims/forever extra_attack.go), so a resetting cast started on
	// that instant can't swallow it.
	if hc := &wa.unit.Hardcast; hc.ResetsSwing && hc.Expires > sim.CurrentTime && wa != &wa.unit.AutoAttacks.ranged && wa.extraAttacksPending == 0 {
		wa.swingAt = hc.Expires + wa.curSwingDuration
		wa.heldByCast = true
		return wa.swingAt
	}

	isExtraAttack := wa.extraAttacksPending > 0

	if isExtraAttack {
		// Needs to happen before any attacks gets cast
		wa.extraAttacks = wa.extraAttacksPending
		wa.extraAttacksPending = 0
		// Any further procs will be added to extraAttacksPending to be processed next batch
	}

	wa.castExtraAttacksStored(sim)

	attackSpell := wa.spell

	// Under Forever only a cast that resets the swing holds it (above). A swing lands through a
	// cast that keeps the swing timer and through a channel, and so does a queued Heroic Strike
	// or Raptor Strike (wowsims/forever 4fb2849945).
	throughCast := sim.IsForever() && wa != &wa.unit.AutoAttacks.ranged

	if wa.replaceSwing != nil {
		// Need to check APL here to allow last-moment HS queue casts.
		wa.unit.Rotation.DoNextAction(sim)

		// Allow MH swing to be overridden for abilities like Heroic Strike.
		wa.unit.swingingThroughCast = throughCast
		attackSpell = wa.replaceSwing(sim, attackSpell)
	}

	wa.unit.swingingThroughCast = throughCast
	canSwing := attackSpell.CanCast(sim, wa.unit.CurrentTarget)
	wa.unit.swingingThroughCast = false

	if canSwing {
		// Update swing timer BEFORE the cast, so that APL checks for TimeToNextAuto behave correctly
		// if the attack causes APL evaluations (e.g. from rage gain).

		wa.swingAt = sim.CurrentTime + wa.curSwingDuration
		wa.lastSwingAt = sim.CurrentTime
		wa.heldByCast = false

		// don't update isExtraAttack here

		var originalCastTime time.Duration
		if isExtraAttack {
			originalCastTime = wa.spell.DefaultCast.CastTime
			wa.spell.DefaultCast.CastTime = 0
			wa.spell.SetMetricsSplit(1)
		} else {
			wa.spell.SetMetricsSplit(0)
		}

		if throughCast {
			wa.unit.swingSpellThroughCast = attackSpell
		}
		if sim.IsForever() && wa == &wa.unit.AutoAttacks.ranged {
			wa.unit.AutoAttacks.meleeAttackOff = true
		}
		attackSpell.Cast(sim, wa.unit.CurrentTarget)
		wa.unit.swingSpellThroughCast = nil

		moreAttacks := !isExtraAttack && wa.extraAttacksPending > 0 // True if above cast is a normal Auto attack that triggered an Extra Attack
		wa.castExtraAttacksTriggered(sim, moreAttacks)              // more attacks means we don't count the above cast

		if isExtraAttack {
			// For ranged extra attacks, we have to wait for the spell to hit before resettings the cast time and metrics split
			if originalCastTime > 0 {
				wa.spell.WaitTravelTime(sim, func(sim *Simulation) {
					wa.spell.DefaultCast.CastTime = originalCastTime
					wa.spell.SetMetricsSplit(0)
				})
			} else {
				wa.spell.SetMetricsSplit(0)
			}
		}

		if wa.extraAttacksPending > 0 {
			wa.spell.SetMetricsSplit(1)
			wa.swingAt = sim.CurrentTime + SpellBatchWindow
			wa.lastSwingAt = sim.CurrentTime
			sim.rescheduleWeaponAttack(wa.swingAt) // Required to fix extra attack procs triggered during swing
		}

		if !sim.Options.Interactive && wa.unit.Rotation != nil {
			wa.unit.Rotation.DoNextAction(sim)
		}
	} else {
		// Delay till cast finishes if casting or 100 ms if not
		wa.swingAt = max(wa.unit.Hardcast.Expires, sim.CurrentTime+time.Millisecond*100)
	}

	return wa.swingAt
}

func (wa *WeaponAttack) updateSwingDuration(curSwingSpeed float64) {
	wa.curSwingSpeed = curSwingSpeed
	wa.curSwingDuration = DurationFromSeconds(wa.SwingSpeed / wa.curSwingSpeed)
}

func (wa *WeaponAttack) addWeaponAttack(sim *Simulation, swingSpeed float64) {
	wa.updateSwingDuration(swingSpeed)
	sim.addWeaponAttack(wa)
	sim.rescheduleWeaponAttack(wa.swingAt)
}

type AutoAttacks struct {
	AutoSwingMelee  bool
	AutoSwingRanged bool

	IsDualWielding bool

	mh     WeaponAttack
	oh     WeaponAttack
	ranged WeaponAttack

	enabled bool

	// Under Forever starting a ranged auto turns the melee attack off, so the next walk into melee
	// starts it again (see EnableAutoSwing).
	meleeAttackOff bool
}

// Options for initializing auto attacks.
type AutoAttackOptions struct {
	MainHand        Weapon
	OffHand         Weapon
	Ranged          Weapon
	AutoSwingMelee  bool // If true, core engine will handle calling SwingMelee() for you.
	AutoSwingRanged bool // If true, core engine will handle calling SwingRanged() for you.
	ReplaceMHSwing  ReplaceMHSwing
}

func (unit *Unit) EnableAutoAttacks(agent Agent, options AutoAttackOptions) {
	if options.MainHand.AttackPowerPerDPS == 0 {
		options.MainHand.AttackPowerPerDPS = DefaultAttackPowerPerDPS
	}
	if options.OffHand.AttackPowerPerDPS == 0 {
		options.OffHand.AttackPowerPerDPS = DefaultAttackPowerPerDPS
	}

	unit.AutoAttacks = AutoAttacks{
		AutoSwingMelee:  options.AutoSwingMelee,
		AutoSwingRanged: options.AutoSwingRanged,

		IsDualWielding: options.OffHand.SwingSpeed != 0,

		mh: WeaponAttack{
			agent:        agent,
			unit:         unit,
			Weapon:       options.MainHand,
			replaceSwing: options.ReplaceMHSwing,
		},
		oh: WeaponAttack{
			agent:  agent,
			unit:   unit,
			Weapon: options.OffHand,
		},
		ranged: WeaponAttack{
			agent:  agent,
			unit:   unit,
			Weapon: options.Ranged,
		},
	}

	unit.AutoAttacks.mh.config = SpellConfig{
		ActionID:    ActionID{OtherID: proto.OtherAction_OtherActionAttack, Tag: tagMainhand},
		SpellSchool: options.MainHand.GetSpellSchool(),
		DefenseType: DefenseTypeMelee,
		ProcMask:    ProcMaskMeleeMHAuto,
		Flags:       SpellFlagMeleeMetrics | SpellFlagNoOnCastComplete,
		CastType:    proto.CastType_CastTypeMainHand,

		MetricSplits: 2,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: TernaryFloat64(options.MainHand.GetSpellSchool() == SpellSchoolPhysical, 1, 0),

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {

			autoInProgress := *spell

			baseDamage := autoInProgress.Unit.MHWeaponDamage(sim, autoInProgress.MeleeAttackPower(target))
			result := autoInProgress.CalcDamage(sim, target, baseDamage, autoInProgress.OutcomeMeleeWhite)

			StartDelayedAction(sim, DelayedActionOptions{
				DoAt: sim.CurrentTime + SpellBatchWindow,
				OnAction: func(s *Simulation) {
					autoInProgress.DealDamage(sim, result)
				},
			})
		},
	}

	unit.AutoAttacks.oh.config = SpellConfig{
		ActionID:    ActionID{OtherID: proto.OtherAction_OtherActionAttack, Tag: tagOffhand},
		SpellSchool: options.OffHand.GetSpellSchool(),
		DefenseType: DefenseTypeMelee,
		ProcMask:    ProcMaskMeleeOHAuto,
		Flags:       SpellFlagMeleeMetrics | SpellFlagNoOnCastComplete,
		CastType:    proto.CastType_CastTypeOffHand,

		MetricSplits: 2,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: TernaryFloat64(options.OffHand.GetSpellSchool() == SpellSchoolPhysical, 1, 0),

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {

			autoInProgress := *spell
			baseDamage := autoInProgress.Unit.OHWeaponDamage(sim, autoInProgress.MeleeAttackPower(target))
			result := autoInProgress.CalcDamage(sim, target, baseDamage, autoInProgress.OutcomeMeleeWhite)

			StartDelayedAction(sim, DelayedActionOptions{
				DoAt: sim.CurrentTime + SpellBatchWindow,
				OnAction: func(s *Simulation) {
					autoInProgress.DealDamage(sim, result)
				},
			})
		},
	}

	unit.AutoAttacks.ranged.config = SpellConfig{
		ActionID:     ActionID{OtherID: proto.OtherAction_OtherActionShoot},
		SpellSchool:  options.Ranged.GetSpellSchool(),
		DefenseType:  DefenseTypeRanged,
		ProcMask:     ProcMaskRangedAuto,
		Flags:        SpellFlagMeleeMetrics,
		CastType:     proto.CastType_CastTypeRanged,
		MissileSpeed: 24,

		MetricSplits: 2,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: TernaryFloat64(options.Ranged.GetSpellSchool() == SpellSchoolPhysical, 1, 0),

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			baseDamage := spell.Unit.RangedWeaponDamage(sim, spell.RangedAttackPower(target, false))
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	}

	if unit.Type == EnemyUnit {
		unit.AutoAttacks.mh.config.ApplyEffects = func(sim *Simulation, target *Unit, spell *Spell) {
			ap := max(0, spell.Unit.stats[stats.AttackPower])
			baseDamage := spell.Unit.AutoAttacks.mh.EnemyWeaponDamage(sim, ap, spell.Unit.PseudoStats.DamageSpread)

			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeEnemyMeleeWhite)
		}
		unit.AutoAttacks.oh.config.ApplyEffects = func(sim *Simulation, target *Unit, spell *Spell) {
			ap := max(0, spell.Unit.stats[stats.AttackPower])
			baseDamage := spell.Unit.AutoAttacks.mh.EnemyWeaponDamage(sim, ap, spell.Unit.PseudoStats.DamageSpread) * 0.5

			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeEnemyMeleeWhite)
		}
	}
}

func (aa *AutoAttacks) finalize() {
	if aa.AutoSwingMelee {
		aa.mh.spell = aa.mh.unit.GetOrRegisterSpell(aa.mh.config)
		aa.mh.spell.TagSplitMetric(1, tagExtraAttack)
		// Will keep the OH spell registered for Item swapping
		aa.oh.spell = aa.oh.unit.GetOrRegisterSpell(aa.oh.config)
		aa.oh.spell.TagSplitMetric(1, tagExtraAttack)
	}
	if aa.AutoSwingRanged {
		aa.ranged.spell = aa.ranged.unit.GetOrRegisterSpell(aa.ranged.config)
		aa.ranged.spell.TagSplitMetric(1, tagExtraAttack)
	}
}

func (aa *AutoAttacks) reset(sim *Simulation) {
	if !aa.AutoSwingMelee && !aa.AutoSwingRanged {
		return
	}

	aa.enabled = false
	aa.meleeAttackOff = false

	aa.mh.swingAt = NeverExpires
	aa.oh.swingAt = NeverExpires
	aa.mh.heldByCast = false
	aa.oh.heldByCast = false

	// Make sure extra attacks are reset
	aa.mh.extraAttacks = 0
	aa.mh.extraAttacksPending = 0
	aa.mh.spell.SetMetricsSplit(0)

	if aa.ranged.spell != nil {
		aa.ranged.extraAttacks = 0
		aa.ranged.extraAttacksPending = 0
		aa.ranged.spell.SetMetricsSplit(0)
	}

	if aa.AutoSwingMelee {
		aa.mh.updateSwingDuration(aa.mh.unit.SwingSpeed())
		aa.mh.swingAt = 0

		if aa.IsDualWielding {
			aa.oh.extraAttacks = 0
			aa.oh.extraAttacksPending = 0
			aa.oh.spell.SetMetricsSplit(0)
			aa.oh.updateSwingDuration(aa.mh.curSwingSpeed)
			aa.oh.swingAt = 0

			// Apply random delay of 0 - 50% swing time, to one of the weapons if dual wielding
			if aa.oh.unit.Type == EnemyUnit {
				aa.oh.swingAt = DurationFromSeconds(aa.mh.SwingSpeed / 2)
			} else {
				if sim.RandomFloat("SwingResetWeapon") < 0.5 {
					aa.mh.swingAt = DurationFromSeconds(sim.RandomFloat("SwingResetDelay") * aa.mh.SwingSpeed / 2)
				} else {
					aa.oh.swingAt = DurationFromSeconds(sim.RandomFloat("SwingResetDelay") * aa.mh.SwingSpeed / 2)
				}
			}
		}

	}

	aa.ranged.swingAt = NeverExpires

	if aa.AutoSwingRanged {
		aa.ranged.updateSwingDuration(aa.ranged.unit.RangedSwingSpeed())
		aa.ranged.swingAt = 0
	}
}

func (aa *AutoAttacks) startPull(sim *Simulation) {

	if !aa.AutoSwingMelee && !aa.AutoSwingRanged {
		return
	}

	if aa.enabled {
		return
	}

	aa.enabled = true

	if aa.AutoSwingMelee && aa.mh.unit.DistanceFromTarget <= MaxMeleeAttackDistance {
		aa.mh.addWeaponAttack(sim, aa.mh.unit.SwingSpeed())
		if aa.IsDualWielding {
			aa.oh.addWeaponAttack(sim, aa.mh.curSwingSpeed)
		}
	}

	if aa.AutoSwingRanged && aa.mh.unit.DistanceFromTarget >= minRangedAttackDistance(sim) {
		if sim.IsForever() {
			// The first shot winds up too.
			aa.ranged.swingAt = max(aa.ranged.swingAt, RangedSwingWindup)
		}
		aa.ranged.addWeaponAttack(sim, aa.ranged.unit.RangedSwingSpeed())
	}
}

// Stops the auto swing action for the rest of the iteration. Used for pets
// after being disabled.
func (aa *AutoAttacks) CancelAutoSwing(sim *Simulation) {
	if !aa.AutoSwingMelee && !aa.AutoSwingRanged {
		return
	}

	if !aa.enabled {
		return
	}

	aa.enabled = false

	if aa.AutoSwingMelee {
		sim.removeWeaponAttack(&aa.mh)
		if aa.IsDualWielding {
			sim.removeWeaponAttack(&aa.oh)
		}
	}

	if aa.AutoSwingRanged {
		sim.removeWeaponAttack(&aa.ranged)
	}
}

// Has the next EnableAutoSwing swing the main hand at once, not when the swing left from before
// CancelAutoSwing comes due: for a guardian that arrives anew each time it is summoned.
func (aa *AutoAttacks) ResetMeleeSwingAt() {
	aa.mh.swingAt = 0
}

// Re-enables the auto swing action for the iteration
func (aa *AutoAttacks) EnableAutoSwing(sim *Simulation) {
	if !aa.AutoSwingMelee && !aa.AutoSwingRanged {
		return
	}

	if aa.enabled {
		return
	}

	aa.enabled = true

	if aa.AutoSwingMelee && aa.mh.unit.DistanceFromTarget <= MaxMeleeAttackDistance {
		aa.mh.swingAt = max(aa.mh.swingAt, sim.CurrentTime, 0)
		aa.mh.addWeaponAttack(sim, aa.mh.unit.SwingSpeed())
		if aa.IsDualWielding {
			aa.oh.swingAt = max(aa.oh.swingAt, sim.CurrentTime, 0)
			aa.oh.addWeaponAttack(sim, aa.mh.unit.SwingSpeed())
			// The attack Auto Shot turned off starts again, and the off hand's first swing comes
			// half a swing after it (wowsims/forever start_attack_test.go).
			if aa.meleeAttackOff {
				aa.oh.swingAt = max(aa.oh.swingAt, sim.CurrentTime+aa.oh.curSwingDuration/2)
				sim.rescheduleWeaponAttack(aa.oh.swingAt)
			}
		}
		aa.meleeAttackOff = false
	}

	if aa.AutoSwingRanged && aa.mh.unit.DistanceFromTarget >= minRangedAttackDistance(sim) {
		aa.ranged.swingAt = max(aa.ranged.swingAt, sim.CurrentTime, 0)
		aa.ranged.addWeaponAttack(sim, aa.ranged.unit.RangedSwingSpeed())
	}
}

// Bows, guns and crossbows can't fire inside this range. Forever keeps the 8 yd minimum of Auto
// Shot's row, counted from both units' combat reach, where melee range is both reaches plus 4/3 yd:
// the deadzone past melee range is 8 - 4/3 yd whatever the target's size (wowsims/forever
// 8a192337d8, from a positioned weaving log on the Gordok Spirit).
func minRangedAttackDistance(sim *Simulation) float64 {
	if sim.IsForever() {
		return ForeverMinRangedAttackDistance
	}
	return MinRangedAttackDistance
}

// Under Forever a ranged auto winds up for RangedSwingWindup before it fires, inside its swing, and
// moving cancels the windup, so a shot can't go off on the move. A shot whose windup the move
// overlapped, or that came due on the way or in melee, winds up a reaction time after the move
// ends, as the player starts Auto Shot again by hand (wowsims/forever 11f285a52a and 966362d038:
// in the weaving log the windup started 0.20 to 0.35 sec after each walk out). A walk into position
// before the pull starts it at the pull, as Auto Shot would pull.
func (aa *AutoAttacks) resumeRangedAfterMove(sim *Simulation) {
	if !sim.IsForever() || !aa.AutoSwingRanged || !aa.enabled || aa.ranged.unit.DistanceFromTarget < ForeverMinRangedAttackDistance {
		return
	}
	if aa.ranged.swingAt < max(sim.CurrentTime, 0)+RangedSwingWindup {
		aa.ranged.swingAt = max(sim.CurrentTime+aa.ranged.unit.ReactionTime, 0) + RangedSwingWindup
		sim.rescheduleWeaponAttack(aa.ranged.swingAt)
	}
}

// The amount of time between two MH swings.
func (aa *AutoAttacks) MainhandSwingSpeed() time.Duration {
	return aa.mh.curSwingDuration
}

// The amount of time between two OH swings.
func (aa *AutoAttacks) OffhandSwingSpeed() time.Duration {
	return aa.oh.curSwingDuration
}

// The amount of time between two Ranged swings.
func (aa *AutoAttacks) RangedSwingSpeed() time.Duration {
	return aa.ranged.curSwingDuration - aa.RangedAuto().CastTime()
}

func (aa *AutoAttacks) UpdateSwingTimers(sim *Simulation) {
	if !aa.enabled {
		return
	}

	if aa.AutoSwingRanged {
		aa.ranged.updateSwingDuration(aa.ranged.unit.RangedSwingSpeed())
		// ranged attack speed changes aren't applied mid-"swing"
	}

	if aa.AutoSwingMelee {
		oldSwingSpeed := aa.mh.curSwingSpeed
		aa.mh.updateSwingDuration(aa.mh.unit.SwingSpeed())
		f := oldSwingSpeed / aa.mh.curSwingSpeed

		if remainingSwingTime := aa.mh.swingAt - sim.CurrentTime; remainingSwingTime > 0 {
			aa.mh.swingAt = sim.CurrentTime + time.Duration(float64(remainingSwingTime)*f)
		}

		sim.rescheduleWeaponAttack(aa.mh.swingAt)

		if aa.IsDualWielding {
			aa.oh.updateSwingDuration(aa.mh.curSwingSpeed)

			if remainingSwingTime := aa.oh.swingAt - sim.CurrentTime; remainingSwingTime > 0 {
				aa.oh.swingAt = sim.CurrentTime + time.Duration(float64(remainingSwingTime)*f)
			}

			sim.rescheduleWeaponAttack(aa.oh.swingAt)
		}
	}
}

func (aa *AutoAttacks) ExtraMHAttackProc(sim *Simulation, attacks int32, actionID ActionID, spell *Spell) {
	if spell.Flags.Matches(SpellFlagBatchStopAttackMacro) {
		aa.StoreExtraMHAttack(sim, attacks, actionID, spell.ActionID)
	} else {
		aa.ExtraMHAttack(sim, attacks, actionID, spell.ActionID)
	}
}

// ExtraMHAttack should be used for all "extra attack" procs in Classic Era versions, including Wild Strikes and Hand of Justice. In vanilla, these procs don't actually grant a full extra attack, but instead just advance the MH swing timer.
func (aa *AutoAttacks) ExtraMHAttack(sim *Simulation, attacks int32, actionID ActionID, triggerAction ActionID) {
	if attacks == 0 {
		return
	}
	if sim.Log != nil {
		attacksText := Ternary(attacks == 1, "attack", "attacks")
		aa.mh.unit.Log(sim, "gained %d extra main-hand %s from %s triggered by %s", attacks, attacksText, actionID, triggerAction)
	}

	aa.mh.swingAt = sim.CurrentTime
	aa.mh.spell.SetMetricsSplit(1)
	sim.rescheduleWeaponAttack(aa.mh.swingAt)
	aa.mh.extraAttacksPending += attacks
}

func (aa *AutoAttacks) StoreExtraMHAttack(sim *Simulation, attacks int32, actionID ActionID, triggerAction ActionID) {
	if attacks == 0 {
		return
	}

	if aa.mh.extraAttacksAura == nil {
		aa.mh.extraAttacksAura = aa.mh.unit.GetAuraByID(ActionID{SpellID: 21919})
	}

	if aa.mh.extraAttacksAura == nil {
		aa.mh.extraAttacksAura = MakePermanent(aa.mh.unit.GetOrRegisterAura(Aura{
			Label:     "Extra Attacks (Main Hand)", // Tracks Stored Extra Attacks from all sources
			ActionID:  ActionID{SpellID: 21919},    // Thrash ID
			Duration:  NeverExpires,
			MaxStacks: 4, // Max is 4 extra attacks stored - more can proc after
		}))
	}

	if !aa.mh.extraAttacksAura.IsActive() {
		aa.mh.extraAttacksAura.Activate(sim)
	}

	aa.mh.extraAttacksAura.AddStacks(sim, attacks)

	if sim.Log != nil {
		aa.mh.unit.Log(sim, "stored %d extra main-hand attacks from %s triggered by %s, total is %d", attacks, actionID, triggerAction, aa.mh.extraAttacksStored)
	}
}

// ExtraMHAttack should be used for all "extra attack" procs in Classic Era versions, including Wild Strikes and Hand of Justice. In vanilla, these procs don't actually grant a full extra attack, but instead just advance the MH swing timer.
func (aa *AutoAttacks) ExtraOHAttack(sim *Simulation, attacks int32, actionID ActionID, triggerAction ActionID) {
	if attacks == 0 {
		return
	}
	if sim.Log != nil {
		attacksText := Ternary(attacks == 1, "attack", "attacks")
		aa.oh.unit.Log(sim, "gained %d extra off-hand %s from %s triggered by %s", attacks, attacksText, actionID, triggerAction)
	}
	aa.oh.swingAt = sim.CurrentTime + SpellBatchWindow
	aa.oh.spell.SetMetricsSplit(1)
	sim.rescheduleWeaponAttack(aa.oh.swingAt)
	aa.oh.extraAttacksPending += attacks
}

// ExtraRangedAttack should be used for all "extra ranged attack" procs in Classic Era versions, including Hand of Injustice. In vanilla, these procs don't actually grant a full extra attack, but instead just advance the Ranged swing timer.
// Note that Hand of Injustice doesn't seem to reset the swing timer however.
func (aa *AutoAttacks) ExtraRangedAttack(sim *Simulation, attacks int32, actionID ActionID, triggerAction ActionID) {
	if sim.Log != nil {
		attacksText := Ternary(attacks == 1, "attack", "attacks")
		aa.mh.unit.Log(sim, "gained %d extra ranged %s from %s triggered by %s", attacks, attacksText, actionID, triggerAction)
	}
	aa.ranged.swingAt = sim.CurrentTime + SpellBatchWindow
	aa.ranged.spell.SetMetricsSplit(1)
	sim.rescheduleWeaponAttack(aa.ranged.swingAt)
	aa.ranged.extraAttacks += attacks
}

// StopMeleeUntil should be used whenever a non-melee spell is cast. It stops melee, then restarts it
// at end of cast, but with a reset swing timer (as if swings had just landed).
func (aa *AutoAttacks) StopMeleeUntil(sim *Simulation, readyAt time.Duration, desyncOH bool) {
	if !aa.AutoSwingMelee { // if not auto swinging, don't auto restart.
		return
	}

	// Under Forever a swing reset that takes effect now leaves a pending extra attack to swing at
	// once; the extra attack then resets the swing itself.
	keepExtraAttack := func(wa *WeaponAttack) bool {
		return wa.extraAttacksPending > 0 && readyAt <= sim.CurrentTime && sim.IsForever()
	}

	if !keepExtraAttack(&aa.mh) {
		aa.mh.swingAt = readyAt + aa.mh.curSwingDuration
		aa.mh.heldByCast = false
		sim.rescheduleWeaponAttack(aa.mh.swingAt)
	}

	if aa.IsDualWielding && !keepExtraAttack(&aa.oh) {
		aa.oh.swingAt = readyAt + aa.oh.curSwingDuration
		aa.oh.heldByCast = false
		if desyncOH {
			// Used by warrior to desync offhand after unglyphed Shattering Throw.
			aa.oh.swingAt += aa.oh.curSwingDuration / 2
		}
		sim.rescheduleWeaponAttack(aa.oh.swingAt)
	}
}

// A cast that ends without completing lets the swings it held land at once: their timers ran out
// while it was underway (wowsims/forever 7146db7c30).
func (aa *AutoAttacks) releaseMeleeHeldByCast(sim *Simulation) {
	for _, wa := range []*WeaponAttack{&aa.mh, &aa.oh} {
		if wa.heldByCast {
			wa.heldByCast = false
			wa.swingAt = sim.CurrentTime
			sim.rescheduleWeaponAttack(wa.swingAt)
		}
	}
}

func (aa *AutoAttacks) StopRangedUntil(sim *Simulation, readyAt time.Duration) {
	if !aa.AutoSwingRanged { // if not auto swinging, don't auto restart.
		return
	}
	aa.ranged.swingAt = readyAt + aa.ranged.curSwingDuration
	sim.rescheduleWeaponAttack(aa.ranged.swingAt)
}

// Delays all swing timers for the specified amount. Only used by Slam.
func (aa *AutoAttacks) DelayMeleeBy(sim *Simulation, delay time.Duration) {
	if delay <= 0 {
		return
	}

	aa.mh.swingAt += delay
	sim.rescheduleWeaponAttack(aa.mh.swingAt)

	if aa.IsDualWielding {
		aa.oh.swingAt += delay
		sim.rescheduleWeaponAttack(aa.oh.swingAt)
	}
}

func (aa *AutoAttacks) DelayRangedUntil(sim *Simulation, readyAt time.Duration) {
	if readyAt <= aa.ranged.swingAt {
		return
	}

	aa.ranged.swingAt = readyAt
	sim.rescheduleWeaponAttack(aa.ranged.swingAt)
}

// Returns the time at which the next melee attack will occur.
func (aa *AutoAttacks) NextAttackAt() time.Duration {
	return min(aa.mh.swingAt, aa.oh.swingAt)
}

// Returns the time at which the next attack will occur.
func (aa *AutoAttacks) NextAnyAttackAt() time.Duration {
	return min(min(aa.mh.swingAt, aa.oh.swingAt), aa.ranged.swingAt)
}

// Returns the time at which the next ranged attack will occur.
func (aa *AutoAttacks) NextRangedAttackAt() time.Duration {
	return aa.ranged.swingAt
}

type PPMManager struct {
	ppm         float64
	procMasks   []ProcMask
	procChances []float64

	// For feral druids, certain PPM effects use their equipped weapon speed
	// instead of their paw attack speed.
	mhSpecialProcChance float64
	ohSpecialProcChance float64
}

// Returns whether the effect procced.
func (ppmm *PPMManager) Proc(sim *Simulation, procMask ProcMask, label string) bool {
	for i, m := range ppmm.procMasks {
		if m.Matches(procMask) {
			return sim.RandomFloat(label) < ppmm.procChances[i]
		}
	}
	return false
}

// Returns whether the effect procced.
// This is different from Proc() in that yellow melee hits use a proc chance based on the equipped
// weapon speed rather than the base attack speed. This distinction matters for feral druids.
func (ppmm *PPMManager) ProcWithWeaponSpecials(sim *Simulation, procMask ProcMask, label string) bool {
	if procMask.Matches(ProcMaskMeleeMHSpecial) {
		return sim.RandomFloat(label) < ppmm.mhSpecialProcChance
	} else if procMask.Matches(ProcMaskMeleeOHSpecial) {
		return sim.RandomFloat(label) < ppmm.ohSpecialProcChance
	} else {
		return ppmm.Proc(sim, procMask, label)
	}
}

func (ppmm *PPMManager) Chance(procMask ProcMask) float64 {
	for i, m := range ppmm.procMasks {
		if m.Matches(procMask) {
			return ppmm.procChances[i]
		}
	}
	return 0
}

func (ppmm *PPMManager) GetPPM() float64 {
	return ppmm.ppm
}

func (aa *AutoAttacks) NewPPMManager(ppm float64, procMask ProcMask) PPMManager {
	if !aa.AutoSwingMelee && !aa.AutoSwingRanged {
		return PPMManager{}
	}

	ppmm := PPMManager{ppm: ppm, procMasks: make([]ProcMask, 0, 2), procChances: make([]float64, 0, 2)}

	mergeOrAppend := func(speed float64, mask ProcMask) {
		if speed == 0 || mask == 0 {
			return
		}

		if i := slices.Index(ppmm.procChances, speed); i != -1 {
			ppmm.procMasks[i] |= mask
			return
		}

		ppmm.procMasks = append(ppmm.procMasks, mask)
		ppmm.procChances = append(ppmm.procChances, speed)
	}

	mhMask := procMask &^ ProcMaskRanged &^ ProcMaskMeleeOH // "everything else", even if not explicitly flagged MH
	ohMask := procMask & ProcMaskMeleeOH
	if aa.oh.SwingSpeed == 0 {
		// With no off-hand weapon an off-hand hit is a shield's (Forever's Shield Slam), and its
		// procs per minute go by the main hand's speed (wowsims/forever b6ce30c972).
		mhMask |= ohMask
		ohMask = 0
	}
	mergeOrAppend(aa.mh.SwingSpeed, mhMask)
	mergeOrAppend(aa.oh.SwingSpeed, ohMask)
	mergeOrAppend(aa.ranged.SwingSpeed, procMask&ProcMaskRanged)

	for i := range ppmm.procChances {
		ppmm.procChances[i] *= ppm / 60
	}

	character := aa.mh.agent.GetCharacter()
	if procMask.Matches(ProcMaskMeleeMH) {
		if character != nil {
			if mhWeapon := character.GetMHWeapon(); mhWeapon != nil {
				ppmm.mhSpecialProcChance = mhWeapon.SwingSpeed * ppm / 60
			}
		}
	}
	if procMask.Matches(ProcMaskMeleeOH) {
		if character != nil {
			if ohWeapon := character.GetOHWeapon(); ohWeapon != nil {
				ppmm.ohSpecialProcChance = ohWeapon.SwingSpeed * ppm / 60
			} else if mhWeapon := character.GetMHWeapon(); mhWeapon != nil {
				ppmm.ohSpecialProcChance = mhWeapon.SwingSpeed * ppm / 60
			}
		}
	}
	return ppmm
}

// Returns whether a PPM-based effect procced.
// Using NewPPMManager() is preferred; this function should only be used when
// the attacker is not known at initialization time.
func (aa *AutoAttacks) PPMProc(sim *Simulation, ppm float64, procMask ProcMask, label string, spell *Spell) bool {
	if !aa.AutoSwingMelee && !aa.AutoSwingRanged {
		return false
	}

	switch {
	case spell.ProcMask.Matches(procMask &^ ProcMaskMeleeOH &^ ProcMaskRanged):
		return sim.RandomFloat(label) < ppm*aa.mh.SwingSpeed/60.0
	case spell.ProcMask.Matches(procMask & ProcMaskMeleeOH):
		// A shield's hit goes by the main hand's speed, as in NewPPMManager.
		return sim.RandomFloat(label) < ppm*TernaryFloat64(aa.oh.SwingSpeed == 0, aa.mh.SwingSpeed, aa.oh.SwingSpeed)/60.0
	case spell.ProcMask.Matches(procMask & ProcMaskRanged):
		return sim.RandomFloat(label) < ppm*aa.ranged.SwingSpeed/60.0
	}
	return false
}

func (unit *Unit) applyParryHaste() {
	if !unit.PseudoStats.ParryHaste || !unit.AutoAttacks.AutoSwingMelee {
		return
	}

	unit.RegisterAura(Aura{
		Label:    "Parry Haste",
		Duration: NeverExpires,
		OnReset: func(aura *Aura, sim *Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitTaken: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if !result.DidParry() {
				return
			}

			// A parry brings our next swing 40% of the swing sooner, but never closer than 20% of
			// the swing after the parry.
			//
			// So with more than 60% of the swing left it comes 40% sooner, with 20 to 60% left
			// it comes 20% of the swing after the parry, and with 20% or less left nothing
			// changes. That's the Classic rule, and the beta log follows it for mobs and for us
			// (shaman_audit.md 6.2). A mob with a 2.0 sec swing that parried us 1.44 sec after
			// its swing swung again at 1.85 sec, where 40% sooner would have been 1.2 sec, before
			// the parry.
			mh := &aura.Unit.AutoAttacks.mh
			minRemainingTime := time.Duration(float64(mh.curSwingDuration) * 0.2)
			if mh.swingAt-sim.CurrentTime <= minRemainingTime {
				return
			}

			newReadyAt := max(mh.swingAt-2*minRemainingTime, sim.CurrentTime+minRemainingTime)
			if sim.Log != nil {
				aura.Unit.Log(sim, "MH Swing reduced by %s due to parry haste, will now occur at %s", mh.swingAt-newReadyAt, newReadyAt)
			}

			mh.swingAt = newReadyAt
			sim.rescheduleWeaponAttack(newReadyAt)
		},
	})
}
