package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

type Stance uint8

const (
	BattleStance Stance = 1 << iota
	DefensiveStance
	BerserkerStance

	AnyStance = BattleStance | DefensiveStance | BerserkerStance
)

func (stance Stance) Matches(other Stance) bool {
	return (stance & other) != 0
}

var StanceCodes = []int32{SpellCode_WarriorStanceBattle, SpellCode_WarriorStanceDefensive, SpellCode_WarriorStanceBerserker}

const stanceEffectCategory = "Stance"

func (warrior *Warrior) StanceMatches(other Stance) bool {
	return warrior.Stance.Matches(other)
}

func (warrior *Warrior) makeStanceSpell(stance Stance, aura *core.Aura, stanceCD *core.Timer) *WarriorSpell {
	spellCode := map[Stance]int32{
		BattleStance:    SpellCode_WarriorStanceBattle,
		DefensiveStance: SpellCode_WarriorStanceDefensive,
		BerserkerStance: SpellCode_WarriorStanceBerserker,
	}[stance]
	actionID := aura.ActionID
	// Forever trains Tactical Mastery at 14, keeping 10 Rage (client 70338, spell 1310185), and
	// Improved Tactical Mastery adds 5 a point over 3 points (development notes post 9,
	// 2026-11-02: "5/10/15 additional Rage (was a 5-point talent granting 3/6/9/12/15)"; client
	// 70753 still has 3 a point on spell 12295; hotfix push 113539, curve 83313, carries 5/10/15).
	maxRetainedRage := 3 * float64(warrior.Talents.ImprovedTacticalMastery)
	if warrior.Env.IsForever() {
		maxRetainedRage = 5 * float64(min(warrior.Talents.ImprovedTacticalMastery, 3))
		if warrior.Level >= 14 {
			maxRetainedRage += 10
		}
	}
	rageMetrics := warrior.NewRageMetrics(actionID)

	stanceSpell := warrior.RegisterSpell(AnyStance, core.SpellConfig{
		SpellCode: spellCode,
		ActionID:  actionID,
		Flags:     core.SpellFlagAPL,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    stanceCD,
				Duration: time.Second,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !warrior.StanceMatches(stance)
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			if warrior.CurrentRage() > maxRetainedRage {
				warrior.SpendRage(sim, warrior.CurrentRage()-maxRetainedRage, rageMetrics)
			}

			if warrior.WarriorInputs.StanceSnapshot {
				// Delayed, so same-GCD casts are affected by the current aura.
				//  Alternatively, those casts could just (artificially) happen before the stance change.
				core.StartDelayedAction(sim, core.DelayedActionOptions{
					DoAt:     sim.CurrentTime + 10*time.Millisecond,
					OnAction: aura.Activate,
				})
			} else {
				aura.Activate(sim)
			}

			warrior.PreviousStance = warrior.Stance
			warrior.Stance = stance
		},
	})

	warrior.Stances = append(warrior.Stances, stanceSpell)

	return stanceSpell
}

func (warrior *Warrior) registerBattleStanceAura() {
	warrior.BattleStanceAura = warrior.RegisterAura(core.Aura{
		Label:    "Battle Stance",
		ActionID: core.ActionID{SpellID: 2457},
		Duration: core.NeverExpires,
	})
	warrior.BattleStanceAura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.PseudoStats.ThreatMultiplier *= 0.8
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.PseudoStats.ThreatMultiplier /= 0.8
		},
	})
}

func (warrior *Warrior) registerDefensiveStanceAura() {
	// Defiance only pays out with a shield equipped now. Defensive Stance Passive 7376 (30%) and
	// Defiance 12792 (15% at rank 3) are two A_MOD_THREAT auras, so they multiply.
	// Under Forever the November 2 development notes raise Defensive Stance to 80% (was 30%) and
	// Defiance to 7/13/20% (was 5/10/15%)
	// (orchestration/audit/ref/client_update/sources/devnotes_post9_2026-11-02.txt). Clients 70753
	// and 70755 unchanged; standards 37 models the notes' values.
	stanceThreat, defiance := 1.3, 0.05*float64(warrior.Talents.Defiance)
	if warrior.Env.IsForever() {
		stanceThreat, defiance = 1.8, []float64{0, 0.07, 0.13, 0.20}[min(int(warrior.Talents.Defiance), 3)]
	}
	if !warrior.PseudoStats.CanBlock {
		defiance = 0
	}
	warrior.defensiveStanceThreatMultiplier = stanceThreat * (1 + defiance)

	warrior.DefensiveStanceAura = warrior.RegisterAura(core.Aura{
		Label:    "Defensive Stance",
		ActionID: core.ActionID{SpellID: 71},
		Duration: core.NeverExpires,
	})
	warrior.DefensiveStanceAura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.PseudoStats.ThreatMultiplier *= warrior.defensiveStanceThreatMultiplier
			ee.Aura.Unit.PseudoStats.DamageDealtMultiplier *= 0.9
			ee.Aura.Unit.PseudoStats.DamageTakenMultiplier *= 0.9
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.PseudoStats.ThreatMultiplier /= warrior.defensiveStanceThreatMultiplier
			ee.Aura.Unit.PseudoStats.DamageDealtMultiplier /= 0.9
			ee.Aura.Unit.PseudoStats.DamageTakenMultiplier /= 0.9
		},
	})
}

func (warrior *Warrior) registerBerserkerStanceAura() {
	warrior.BerserkerStanceAura = warrior.RegisterAura(core.Aura{
		Label:    "Berserker Stance",
		ActionID: core.ActionID{SpellID: 2458},
		Duration: core.NeverExpires,
	})
	warrior.BerserkerStanceAura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.PseudoStats.ThreatMultiplier *= 0.8
			ee.Aura.Unit.PseudoStats.DamageTakenMultiplier *= 1.1
			ee.Aura.Unit.AddStatDynamic(sim, stats.MeleeCrit, core.CritRatingPerCritChance*3)
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.PseudoStats.ThreatMultiplier /= 0.8
			ee.Aura.Unit.PseudoStats.DamageTakenMultiplier /= 1.1
			ee.Aura.Unit.AddStatDynamic(sim, stats.MeleeCrit, -core.CritRatingPerCritChance*3)
		},
	})
}

func (warrior *Warrior) registerStances() {
	warrior.Stances = make([]*WarriorSpell, 0)
	stanceCD := warrior.NewTimer()
	warrior.registerBattleStanceAura()
	warrior.registerDefensiveStanceAura()
	warrior.registerBerserkerStanceAura()
	warrior.BattleStance = warrior.makeStanceSpell(BattleStance, warrior.BattleStanceAura, stanceCD)
	warrior.DefensiveStance = warrior.makeStanceSpell(DefensiveStance, warrior.DefensiveStanceAura, stanceCD)
	warrior.BerserkerStance = warrior.makeStanceSpell(BerserkerStance, warrior.BerserkerStanceAura, stanceCD)
}
