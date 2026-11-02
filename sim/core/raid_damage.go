package core

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

// The boss's abilities on the raid in a healing fight (Encounter.raid_damage,
// docs/healing_sim.md): who each one hits, when, how hard, and whether it ticks on.
//
// Each event is a spell of its boss that always lands: physical hits go through the player's
// armor, the others through its resistances, and both through its damage taken modifiers. Only
// living players are picked, and a dead boss casts nothing more. Unlike raid_hits.go's hits, which
// only feed procs in the DPS Report, these are meant to be healed, so they belong in a fight with
// Encounter.track_health.
func (env *Environment) registerRaidDamage() {
	for i, event := range env.Encounter.RaidDamage {
		if event.MaxDamage <= 0 && event.MinDamage <= 0 {
			continue
		}
		if event.TargetIndex < 0 || int(event.TargetIndex) >= len(env.Encounter.TargetUnits) {
			continue
		}
		env.registerRaidDamageEvent(i, event, env.Encounter.TargetUnits[event.TargetIndex])
	}
}

func (env *Environment) registerRaidDamageEvent(index int, event *proto.RaidDamageEvent, boss *Unit) {
	actionID := ActionID{SpellID: event.SpellId}
	if event.SpellId == 0 {
		actionID = ActionID{OtherID: proto.OtherAction_OtherActionDamageTaken, Tag: int32(index + 1)}
	}
	school := SpellSchoolFromProto(event.School)
	defenseType := DefenseTypeMagic
	if school == SpellSchoolPhysical {
		defenseType = DefenseTypeMelee
	}
	minDamage, maxDamage := event.MinDamage, max(event.MaxDamage, event.MinDamage)
	roll := func(sim *Simulation) float64 {
		return sim.RollWithLabel(minDamage, maxDamage, "Raid Damage")
	}

	// A melee ability rolls the player's avoidance and block, and is a melee hit for its procs
	// (Holy Shield, Shield Block's charges).
	// A ticking event is never a melee hit: it lands and ticks, and its debuff comes with it.
	procMask := ProcMaskSpellDamage
	if event.Melee && event.Ticks == 0 {
		procMask = ProcMaskMeleeMHSpecial
		defenseType = DefenseTypeMelee
	}

	// What a landed hit leaves on its target, one aura a player.
	var debuffs map[*Unit]*Aura
	if m := event.HealingTakenMultiplier; m > 0 && m != 1 && event.DebuffSeconds > 0 {
		debuffs = map[*Unit]*Aura{}
		for _, player := range env.Raid.AllPlayerUnits {
			debuffs[player] = player.RegisterAura(Aura{
				Label:    "Raid Damage Debuff " + strconv.Itoa(index+1),
				ActionID: actionID,
				Duration: DurationFromSeconds(event.DebuffSeconds),
				OnGain: func(aura *Aura, sim *Simulation) {
					aura.Unit.PseudoStats.HealingTakenMultiplier *= m
				},
				OnExpire: func(aura *Aura, sim *Simulation) {
					aura.Unit.PseudoStats.HealingTakenMultiplier /= m
				},
			})
		}
	}
	deal := func(sim *Simulation, target *Unit, spell *Spell) {
		outcome := spell.OutcomeAlwaysHit
		if event.Melee {
			outcome = spell.OutcomeEnemyMeleeSpecial
		}
		result := spell.CalcAndDealDamage(sim, target, roll(sim), outcome)
		if aura := debuffs[target]; aura != nil && result.Landed() {
			aura.Activate(sim)
		}
	}

	config := SpellConfig{
		ActionID:    actionID,
		SpellSchool: school,
		DefenseType: defenseType,
		ProcMask:    procMask,
		Flags:       SpellFlagNoOnCastComplete | SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
	}
	if event.Ticks > 0 && event.TickSeconds > 0 {
		config.Dot = DotConfig{
			Aura:          Aura{Label: "Raid Damage " + strconv.Itoa(index+1)},
			NumberOfTicks: event.Ticks,
			TickLength:    DurationFromSeconds(event.TickSeconds),
			OnSnapshot: func(sim *Simulation, target *Unit, dot *Dot, isRollover bool) {
				dot.Snapshot(target, roll(sim), isRollover)
			},
			OnTick: func(sim *Simulation, target *Unit, dot *Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		}
		config.ApplyEffects = func(sim *Simulation, target *Unit, spell *Spell) {
			spell.Dot(target).Apply(sim)
			if aura := debuffs[target]; aura != nil {
				aura.Activate(sim)
			}
		}
	} else {
		config.ApplyEffects = deal
	}
	spell := boss.RegisterSpell(config)

	every := DurationFromSeconds(max(event.EverySeconds, 0))
	spread := DurationFromSeconds(max(event.SpreadSeconds, 0))
	wait := func(sim *Simulation) time.Duration {
		if event.RandomTiming {
			return time.Duration(sim.RandomExpFloat("Raid Damage Wait") * float64(every))
		}
		if spread == 0 {
			return every
		}
		return every - spread + time.Duration(sim.RandomFloat("Raid Damage Wait")*float64(2*spread))
	}

	boss.RegisterResetEffect(func(sim *Simulation) {
		pa := &PendingAction{NextActionAt: DurationFromSeconds(max(event.FirstAt, 0))}
		pa.OnAction = func(sim *Simulation) {
			if boss.dead {
				return
			}
			for _, target := range env.raidDamageTargets(sim, event, boss) {
				spell.Cast(sim, target)
			}
			if every <= 0 {
				return
			}
			pa.NextActionAt = sim.CurrentTime + max(wait(sim), time.Millisecond)
			sim.AddPendingAction(pa)
		}
		sim.AddPendingAction(pa)
	})
}

// raidDamageTargets picks the living players an event hits this time.
func (env *Environment) raidDamageTargets(sim *Simulation, event *proto.RaidDamageEvent, boss *Unit) []*Unit {
	isTank := func(unit *Unit) bool {
		for _, tank := range env.RaidTanks {
			if tank == unit {
				return true
			}
		}
		return false
	}

	var group []*Unit
	switch event.Who {
	case proto.RaidDamageEvent_BossTarget:
		if t := boss.CurrentTarget; t != nil && t.Type != EnemyUnit && !t.dead {
			group = append(group, t)
		}
	default:
		for _, player := range env.Raid.AllPlayerUnits {
			if player.dead {
				continue
			}
			switch event.Who {
			case proto.RaidDamageEvent_Tanks:
				if !isTank(player) {
					continue
				}
			case proto.RaidDamageEvent_NotTanks:
				if isTank(player) {
					continue
				}
			}
			group = append(group, player)
		}
	}

	count := int(event.Count)
	if event.Share > 0 {
		n := event.Share * float64(len(group))
		count = int(n)
		if sim.RandomFloat("Raid Damage Share") < n-float64(count) {
			count++
		}
		count = max(count, 1)
	}
	if count <= 0 || count >= len(group) {
		return group
	}
	// count random ones, each as likely: a partial Fisher-Yates shuffle of a copy.
	picked := append([]*Unit(nil), group...)
	for i := 0; i < count; i++ {
		j := i + int(sim.RandomFloat("Raid Damage Target")*float64(len(picked)-i))
		picked[i], picked[j] = picked[j], picked[i]
	}
	return picked[:count]
}
