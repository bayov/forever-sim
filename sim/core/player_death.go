package core

// killPlayer is a player dying in a healing fight (Encounter.track_health, docs/healing_sim.md).
// It stops acting: its cast or channel in progress, its rotation and its swings end, its auras
// go and its mana stops coming back (mana.go). The mobs attacking it turn to the next living
// tank, or to another living player when no tank is left. Heals don't reach it (health.go): a
// heal being cast at it stops unfinished and unpaid, as the game's cast fails, and nothing it
// would cast goes (Spell.Cast and CanCast). The fight goes on without it, and a health fight
// ends when the whole raid is dead. Its pets and the dots it put up carry on, its totems don't.
func (env *Environment) killPlayer(sim *Simulation, player *Unit) {
	if player.dead {
		return
	}
	player.dead = true

	// A channel isn't a hardcast: its dot lives on its target.
	if dot := player.ChanneledDot; dot != nil {
		dot.Cancel(sim)
	}
	if player.IsCasting(sim) {
		player.interruptHardcast(sim)
	}
	if player.gcdAction != nil {
		player.CancelGCDTimer(sim)
	}
	player.auraTracker.expireAll(sim)
	// After the auras, since some turn swings back on as they end (a druid's form).
	player.AutoAttacks.CancelAutoSwing(sim)

	if agent := env.Raid.GetPlayerFromUnit(player); agent != nil {
		for _, effect := range agent.GetCharacter().levelingDeathEffects {
			effect(sim)
		}
	}

	for _, unit := range env.Raid.AllUnits {
		if unit != player && unit.Hardcast.Target == player && unit.IsCasting(sim) {
			unit.interruptHardcast(sim)
		}
	}

	for _, target := range env.Encounter.TargetUnits {
		if target.CurrentTarget == player {
			if next := env.nextLivingTank(player); next != nil {
				target.CurrentTarget = next
			}
		}
	}

	// A health fight ends when nobody is left to kill the boss.
	if env.Encounter.EndFightAtHealth > 0 && env.nextLivingTank(nil) == nil {
		sim.endOfCombatDamage = -1
	}
}

// nextLivingTank is the living tank after from in Raid.tanks' order, wrapping around, or with no
// tank alive the first living player. Nil when everyone is dead.
func (env *Environment) nextLivingTank(from *Unit) *Unit {
	start := -1
	for i, tank := range env.RaidTanks {
		if tank == from {
			start = i
		}
	}
	for k := 1; k <= len(env.RaidTanks); k++ {
		tank := env.RaidTanks[(start+k+len(env.RaidTanks))%len(env.RaidTanks)]
		if tank != nil && !tank.dead {
			return tank
		}
	}
	for _, player := range env.Raid.AllPlayerUnits {
		if !player.dead {
			return player
		}
	}
	return nil
}

// registerTankSwaps makes each target with Target.tank_swap_seconds turn to the next living tank
// on that timer (docs/healing_sim.md).
func (env *Environment) registerTankSwaps() {
	for _, target := range env.Encounter.Targets {
		every := target.tankSwapEvery
		// A swap trades the boss between tanks, so it needs one to start on (Target.tank_index).
		if every <= 0 || len(env.RaidTanks) < 2 || target.CurrentTarget == nil {
			continue
		}
		unit := &target.Unit
		unit.RegisterResetEffect(func(sim *Simulation) {
			swap := &PendingAction{NextActionAt: every}
			swap.OnAction = func(sim *Simulation) {
				if next := env.nextLivingTank(unit.CurrentTarget); next != nil && next != unit.CurrentTarget {
					if sim.Log != nil {
						unit.Log(sim, "Turns to %s", next.Label)
					}
					unit.CurrentTarget = next
				}
				swap.NextActionAt = sim.CurrentTime + every
				sim.AddPendingAction(swap)
			}
			sim.AddPendingAction(swap)
		})
	}
}
