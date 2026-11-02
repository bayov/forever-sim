package core

// killPlayer is a player dying in a healing fight (Encounter.track_health, docs/healing_sim.md).
// It stops acting: its cast in progress, its rotation and its swings end, and its auras go. The
// mobs attacking it turn to the next living tank, or to another living player when no tank is
// left. Heals don't reach it (health.go), and nothing it would cast goes (Spell.Cast and
// CanCast). The fight goes on without it. Its pets and the dots it put up carry on.
func (env *Environment) killPlayer(sim *Simulation, player *Unit) {
	if player.dead {
		return
	}
	player.dead = true

	if player.IsCasting(sim) {
		player.interruptHardcast(sim)
	}
	if player.gcdAction != nil {
		player.CancelGCDTimer(sim)
	}
	player.AutoAttacks.CancelAutoSwing(sim)
	player.auraTracker.expireAll(sim)

	for _, target := range env.Encounter.TargetUnits {
		if target.CurrentTarget == player {
			if next := env.nextLivingTank(player); next != nil {
				target.CurrentTarget = next
			}
		}
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
		if every <= 0 || len(env.RaidTanks) < 2 {
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
