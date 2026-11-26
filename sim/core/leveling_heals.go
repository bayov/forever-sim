package core

import (
	"time"
)

// What heals us while we fight in a leveling session and isn't one of our own abilities: the Legacy
// perk Thrill of Adventure, Krol'dok Battlegear's bonus, and health per 5 sec (docs/leveling_sim.md).
// Each is off unless the session's options or our gear turn it on.

// Thrill of Adventure (client 71091, 1225472 and its buff 1261483): a kill that gives experience
// heals us 1% of max health a point over 10 sec, in 5 ticks, and gives as much mana. It doesn't
// work in dungeons, raids and battlegrounds.
//
// Our logs (41,229 ticks of other players and ours, 2026-10-10 to 2026-11-24) have 0.2% of max
// health a tick and point, whole health picked up or down at random. Another kill starts the 10 sec
// and the tick's 2 sec over, with the same heal a tick, so whatever is left of the first one is lost.
const (
	thrillOfAdventureMaxPoints = 5
	thrillOfAdventureTicks     = 5
	thrillOfAdventureTick      = 2 * time.Second
	thrillOfAdventureShare     = 0.01
)

// applyLevelingHeals turns on what the session's options ask for.
func (character *Character) applyLevelingHeals() {
	s := character.Env.Encounter.leveling
	if s == nil || s.character != character {
		return
	}
	character.registerThrillOfAdventure(min(s.opts.thrillOfAdventure, thrillOfAdventureMaxPoints))
	from := s.opts.krolDokFromLevel
	if character.krolDokPieces() >= krolDokPiecesNeeded || (from > 0 && s.level >= from) {
		character.registerKrolDokResolve()
	}
	if s.opts.healthRegenPer5 > 0 {
		s.gearRegenMetrics = character.NewHealthMetrics(ActionID{SpellID: healthRegenPer5ID})
	}
}

func (character *Character) registerThrillOfAdventure(points int32) {
	if points <= 0 {
		return
	}
	s := character.Env.Encounter.leveling
	actionID := ActionID{SpellID: 1261483}
	healthMetrics := character.NewHealthMetrics(actionID)
	var manaMetrics *ResourceMetrics
	if character.HasManaBar() {
		manaMetrics = character.NewManaMetrics(actionID)
	}
	share := thrillOfAdventureShare * float64(points) / thrillOfAdventureTicks

	var ticks *PendingAction
	aura := character.RegisterAura(Aura{
		Label:    "Thrill of Adventure",
		ActionID: actionID,
		Duration: thrillOfAdventureTicks * thrillOfAdventureTick,
	})
	character.RegisterResetEffect(func(*Simulation) {
		ticks = nil
	})
	stop := func(sim *Simulation) {
		if ticks != nil {
			ticks.Cancel(sim)
			ticks = nil
		}
	}
	character.RegisterKillEffect(func(sim *Simulation, _ *Unit, gaveXP bool) {
		if !gaveXP || s.state == levelingDead {
			return
		}
		stop(sim)
		aura.Activate(sim)
		ticks = StartPeriodicAction(sim, PeriodicActionOptions{
			Period:   thrillOfAdventureTick,
			NumTicks: thrillOfAdventureTicks,
			Priority: ActionPriorityDOT,
			OnAction: func(sim *Simulation) {
				character.GainHealth(sim, share*character.MaxHealth()*character.PseudoStats.HealingTakenMultiplier, healthMetrics)
				if manaMetrics != nil {
					character.AddMana(sim, share*character.MaxMana(), manaMetrics)
				}
			},
		})
	})
	character.RegisterLevelingDeathEffect(func(sim *Simulation) {
		stop(sim)
		aura.Deactivate(sim)
	})
}

// Krol'dok Battlegear's 4 piece bonus, Krol'dok Resolve (client 71091, set 2135, 1324298 and its heal
// 1324296): every attack we block heals 18 (13.5 to 22.5), at most once a second since the server
// hotfix of 2026-11-16 (SpellAuraOptions 248630, push 114115). "Healing is doubled in Strongholds and
// Cities" isn't modelled, because we don't grind there (it's a flat 44 in our logs).
//
// Our logs back it: 375 of 377 blocks with a second or more since the last heal healed, in full or
// in part blocked, and no two heals are under a second apart. It crits for half more now and then.
const (
	krolDokSetID        = 2135
	krolDokPiecesNeeded = 4

	krolDokHealMin  = 13.5
	krolDokHealMax  = 22.5
	krolDokCooldown = time.Second
)

// krolDokPieces is how many pieces of the set we wear. The set isn't in sim/common's item sets, whose
// bonuses every raid and item test builds, because its one bonus only matters in a session.
func (character *Character) krolDokPieces() int {
	pieces := 0
	for _, item := range character.Equipment {
		if item.SetID == krolDokSetID {
			pieces++
		}
	}
	return pieces
}

func (character *Character) registerKrolDokResolve() {
	heal := character.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 1324296},
		SpellSchool:      SpellSchoolPhysical,
		DefenseType:      DefenseTypeMagic,
		ProcMask:         ProcMaskEmpty,
		Flags:            SpellFlagHelpful | SpellFlagNoOnCastComplete | SpellFlagPassiveSpell,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			spell.CalcAndDealHealing(sim, target, sim.Roll(krolDokHealMin, krolDokHealMax), spell.OutcomeHealingCrit)
		},
	})
	icd := Cooldown{Timer: character.NewTimer(), Duration: krolDokCooldown}
	MakePermanent(character.RegisterAura(Aura{
		Label:    "Krol'dok Resolve",
		ActionID: ActionID{SpellID: 1324298},
		OnSpellHitTaken: func(_ *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if result.DidBlock() && spell.ProcMask.Matches(ProcMaskMeleeOrRanged) && icd.IsReady(sim) {
				icd.Use(sim)
				heal.Cast(sim, &character.Unit)
			}
		},
	}))
}

// healthRegenPer5ID is the "Restores 4 health per 5 sec" spell of items, for the metrics.
const healthRegenPer5ID = 21589

// gearRegen is a second of "Restores X health per 5 sec" from the health_regen_per5 option. Those items
// carry aura 161 (client 71091, 21587 to 21617), which heals in combat and out of it, on top of
// Spirit's regen and without the sitting bonus.
func (s *levelingSession) gearRegen(sim *Simulation, seconds float64) bool {
	if s.opts.healthRegenPer5 <= 0 || s.state == levelingDead {
		return false
	}
	return s.heal(sim, s.opts.healthRegenPer5/5*seconds, s.gearRegenMetrics)
}
